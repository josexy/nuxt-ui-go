package nui_test

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

// Components that wait, or that are hidden, must not keep the window building:
// every build runs the application's whole view. Steady frames must not copy
// data that did not change.

// buildsDuring returns how many times the view was built while the Tester's
// clock moved on by d.
func buildsDuring(tt *ui.Tester, builds *int, d time.Duration) int {
	before := *builds
	tt.Advance(d)
	return *builds - before
}

func keptTabs(c *ui.Context, th nui.Theme, active *string, hidden func(*ui.Context, nui.Theme)) {
	nui.Tabs(c, th, active, []nui.TabItem{
		{ID: "shown", Label: "Shown", Content: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Shown panel") }},
		{ID: "kept", Label: "Kept", Content: hidden},
	}, nui.TabsOptions{Label: "Panels", KeepMounted: true})
}

func TestHiddenChartWaitsAndMeasuresWhenShown(t *testing.T) {
	var chart nui.ChartState
	if err := chart.SetData([]float64{1, 2, 3}, []nui.ChartSeries{{ID: "s", Label: "Series", Values: []float64{1, 3, 2}}}); err != nil {
		t.Fatal(err)
	}
	active, builds := "shown", 0
	tt := ui.NewTester(func(c *ui.Context) {
		builds++
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Width(400).Children(func() {
			keptTabs(c, th, &active, func(c *ui.Context, th nui.Theme) {
				nui.Chart(c, th, &chart, nui.ChartOptions{Label: "Revenue", AspectRatio: 2, XAxis: nui.Set(false), YAxis: nui.Set(false), Legend: nui.Set(false), Animate: nui.Set(false)})
			})
		})
	}, 600, 500)
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 2 {
		t.Fatalf("a chart in a hidden panel built the view %d times in 2s", n)
	}
	click(t, tt, "Kept")
	tt.Advance(100 * time.Millisecond)
	plot := bounds(t, tt, "Revenue")
	if plot.W < 390 || plot.H < plot.W/2-1 || plot.H > plot.W/2+1 {
		t.Fatalf("shown chart kept its guessed size: %v", plot)
	}
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 2 {
		t.Fatalf("a shown chart built the view %d times in 2s", n)
	}
}

func TestHiddenHorizontalListboxWaitsAndShowsRows(t *testing.T) {
	items := make([]nui.ListboxItem, 50)
	for i := range items {
		items[i] = nui.ListboxItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Item %02d", i)}
	}
	var state nui.ListboxState
	var selected []string
	active, builds := "shown", 0
	tt := ui.NewTester(func(c *ui.Context) {
		builds++
		th := nui.App(c, nui.AppConfig{})
		keptTabs(c, th, &active, func(c *ui.Context, th nui.Theme) {
			nui.Listbox(c, th, &selected, &state, items, nui.ListboxOptions{Label: "Choices", Orientation: nui.Set(nui.Horizontal), ItemWidth: 120}).Width(360)
		})
	}, 1200, 300)
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 2 {
		t.Fatalf("a horizontal listbox in a hidden panel built the view %d times in 2s", n)
	}
	click(t, tt, "Kept")
	tt.Advance(100 * time.Millisecond)
	// The 360 DIP viewport shows three 120 DIP rows; the window's 1200 DIPs,
	// guessed while it was hidden, would build ten and more.
	if !tt.HasText("Item 00") || !tt.HasText("Item 02") || tt.HasText("Item 08") {
		t.Fatalf("shown listbox did not measure its viewport: %v", tt.Texts())
	}
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 2 {
		t.Fatalf("a shown horizontal listbox built the view %d times in 2s", n)
	}
}

// bytesPerFrame returns how many bytes a steady frame allocates.
func bytesPerFrame(tt *ui.Tester) uint64 {
	const frames = 10
	tt.Frame()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for range frames {
		tt.Frame()
	}
	runtime.ReadMemStats(&after)
	return (after.TotalAlloc - before.TotalAlloc) / frames
}

func TestTableSelectionNoticeKeepsOutsideEditsSilentWithoutCopying(t *testing.T) {
	const n = 20000
	rows := make([]string, n)
	for i := range rows {
		rows[i] = fmt.Sprintf("Row %05d", i)
	}
	var state nui.TableState[string]
	state.SetRows(rows, func(v string) string { return v })
	cols := []nui.TableColumn[string]{{ID: "name", Title: "Name", Width: 300, Value: func(v string) string { return v }}}
	var reported [][]string
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Table(c, nui.App(c, nui.AppConfig{}), &state, cols, nui.TableOptions{Label: "Rows", OnSelectionChange: func(ids []string) { reported = append(reported, ids) }}).Width(400)
	}, 500, 400)
	for _, id := range rows[n/2:] {
		state.Selection.Add(id)
	}
	tt.Frame()
	// A frame that copied the choice would allocate its 10000 IDs again.
	if b := bytesPerFrame(tt); b > 64<<10 {
		t.Fatalf("a steady frame with %d rows chosen allocated %d bytes", state.Selection.Len(), b)
	}
	// An outside edit that keeps the count is not the user's either.
	state.Selection.Clear()
	for _, id := range rows[:n/2] {
		state.Selection.Add(id)
	}
	tt.Frame()
	if len(reported) != 0 {
		t.Fatalf("outside edits were reported as the user's: %d notices", len(reported))
	}
	click(t, tt, "Row 00003")
	if len(reported) != 1 || len(reported[0]) != 1 || reported[0][0] != "Row 00003" {
		t.Fatalf("user choice after outside edits: %v", reported)
	}
}

// awaitAdvancing waits for a worker without forcing frames: Advance builds
// only when something asked for a frame, as a window does.
func awaitAdvancing(t *testing.T, tt *ui.Tester, ready func() bool) {
	t.Helper()
	until := time.Now().Add(2 * time.Second)
	for !ready() && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
		tt.Advance(16 * time.Millisecond)
	}
	if !ready() {
		t.Fatal("the worker's result did not reach a build")
	}
}

func TestFileUploadWaitsForItsChooserWithoutPolling(t *testing.T) {
	release := make(chan struct{})
	var state nui.FileUploadState
	builds := 0
	tt := ui.NewTester(func(c *ui.Context) {
		builds++
		nui.FileUpload(c, nui.App(c, nui.AppConfig{}), &state, nui.FileUploadOptions{Label: "Files", Choose: func(ctx context.Context) ([]nui.UploadFile, error) {
			select {
			case <-release:
				return []nui.UploadFile{{Path: "/tmp/report.txt", Name: "report.txt", Size: 12}}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}})
	}, 500, 400)
	click(t, tt, "Files")
	if !state.Pending() {
		t.Fatal("the chooser did not start")
	}
	// A native file dialog stays open for as long as the user takes.
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 2 {
		t.Fatalf("a pending chooser built the view %d times in 2s", n)
	}
	close(release)
	awaitAdvancing(t, tt, func() bool { return !state.Pending() && len(state.Files) == 1 })
}

func TestCommandPaletteResultsArriveWithoutPolling(t *testing.T) {
	actions := make([]nui.CommandAction, 300)
	for i := range actions {
		actions[i] = nui.CommandAction{ID: fmt.Sprint("action-", i), Label: fmt.Sprintf("Action %03d", i)}
	}
	var state nui.CommandPaletteState
	state.SetIndex(nui.NewCommandIndex(actions))
	state.Open = true
	builds := 0
	// Without motion or a caret, nothing but the search asks for frames.
	tt := ui.NewTesterWith(func(c *ui.Context) {
		builds++
		nui.CommandPalette(c, nui.App(c, nui.AppConfig{}), &state, nui.CommandPaletteOptions{SearchInput: nui.Set(false)})
	}, 900, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	awaitAdvancing(t, tt, func() bool { return !state.Searching() && len(state.Results.Items) > 0 })
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 2 {
		t.Fatalf("an idle palette built the view %d times in 2s", n)
	}
	state.Query = "Action 29"
	state.Refresh()
	tt.Frame() // starts the search on a worker
	awaitAdvancing(t, tt, func() bool {
		return !state.Searching() && len(state.Results.Items) > 0 && state.Results.Items[0].Action.Label == "Action 290"
	})
	if !tt.HasText("Action 290") {
		t.Fatalf("results were not built: %v", tt.Texts())
	}
}

func TestCommandPaletteKeepsItsRowsBetweenFrames(t *testing.T) {
	actions := make([]nui.CommandAction, 1000)
	for i := range actions {
		actions[i] = nui.CommandAction{ID: fmt.Sprint("action-", i), Label: fmt.Sprintf("Action %03d", i), Group: fmt.Sprint("Group ", i%5)}
	}
	actions[5].Active = true // the second row of the first group
	var state nui.CommandPaletteState
	state.SetIndex(nui.NewCommandIndex(actions))
	state.Open = true
	tt := ui.NewTesterWith(func(c *ui.Context) {
		nui.CommandPalette(c, nui.App(c, nui.AppConfig{}), &state, nui.CommandPaletteOptions{Limit: 1000, SearchInput: nui.Set(false)})
	}, 900, 700, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	awaitAdvancing(t, tt, func() bool { return !state.Searching() && len(state.Results.Items) == 1000 })
	// Grouping and copying 1000 results again allocated megabytes a frame.
	if b := bytesPerFrame(tt); b > 64<<10 {
		t.Fatalf("a steady frame with 1000 results allocated %d bytes", b)
	}
	// Kept rows still follow the theme: the active row takes its Elevated.
	for _, dark := range []bool{false, true} {
		tt.SetDark(dark)
		tt.Frame()
		want := nui.Light().Elevated
		if dark {
			want = nui.Dark().Elevated
		}
		r := bounds(t, tt, "Action 005")
		if got := tt.Image().RGBAAt(int(r.X)+3, int(r.Y+r.H/2)); got.R != want.R || got.G != want.G || got.B != want.B {
			t.Fatalf("dark=%v: active row %v, want the theme's Elevated %v", dark, got, want)
		}
	}
}

func TestToastCountdownPaintsWithoutBuilding(t *testing.T) {
	builds := 0
	tt := ui.NewTester(func(c *ui.Context) {
		builds++
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{Title: "Countdown", Color: nui.Error})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 600, 400)
	click(t, tt, "Notify")
	tt.Move(5, 5)
	tt.Advance(300 * time.Millisecond) // the enter transition
	panel := bounds(t, tt, "Countdown")
	bar := func(at float32) color.RGBA { return tt.Image().RGBAAt(int(panel.X+panel.W*at), int(panel.Y+panel.H-2)) }
	full := bar(.6)
	// The bar empties from its end: past half of the default 5s, the
	// point at 60% shows the track instead.
	n := buildsDuring(tt, &builds, 2500*time.Millisecond)
	if bar(.6) == full {
		t.Fatal("the countdown bar did not move")
	}
	if n > 4 {
		t.Fatalf("a toast counting down built the view %d times in 2.5s", n)
	}
	// The pointer on the toasts stops their time: nothing moves, nothing builds.
	tt.Move(panel.X+panel.W/2, panel.Y+panel.H/2)
	tt.Advance(300 * time.Millisecond) // the hover's color transitions
	// About 2.8s of 5s have gone: the bar ends near 44%, and stays there.
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 1 || bar(.4) != full || !tt.HasText("Countdown") {
		t.Fatalf("a held toast built %d times, bar at 40%% %v, fill %v", n, bar(.4), full)
	}
	tt.Move(5, 5)
	tt.Advance(4 * time.Second)
	if tt.HasText("Countdown") {
		t.Fatal("the toast did not close once its time ran out")
	}
}

// redEdges returns where red content starts along the middle of r, across
// it, or down it when vertical.
func redEdges(im *image.RGBA, r ui.Rect, vertical bool) []int {
	var at []int
	inside := false
	n := int(r.W)
	if vertical {
		n = int(r.H)
	}
	for i := range n {
		x, y := int(r.X)+i, int(r.Y+r.H/2)
		if vertical {
			x, y = int(r.X+r.W/2), int(r.Y)+i
		}
		p := im.RGBAAt(x, y)
		red := p.R > 150 && p.G < 120
		if red && !inside {
			at = append(at, i)
		}
		inside = red
	}
	return at
}

func TestMarqueeMovesInFramesThatOnlyPaint(t *testing.T) {
	for _, o := range []nui.MarqueeOptions{
		{Label: "Logos", Overlay: nui.Set(false)},
		{Label: "Logos", Overlay: nui.Set(false), Reverse: true},
		{Label: "Logos", Overlay: nui.Set(false), Orientation: nui.Vertical, Height: 200},
	} {
		t.Run(fmt.Sprintf("reverse=%v/vertical=%v", o.Reverse, o.Orientation == nui.Vertical), func(t *testing.T) {
			var state nui.MarqueeState
			builds := 0
			tt := ui.NewTester(func(c *ui.Context) {
				builds++
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				nui.Marquee(c, th, &state, o, func() {
					for i := range 3 {
						ui.Box(c.Key(i)).Size(40, 24).Background(ui.RGB(200, 30, 30))
					}
				}).Width(300)
			}, 400, 300)
			tt.Advance(100 * time.Millisecond) // measured, then moving
			region := bounds(t, tt, "Logos")
			vertical := o.Orientation == nui.Vertical
			before := redEdges(tt.Image(), region, vertical)
			n := buildsDuring(tt, &builds, 2500*time.Millisecond)
			moved := redEdges(tt.Image(), region, vertical)
			if len(before) == 0 || fmt.Sprint(before) == fmt.Sprint(moved) {
				t.Fatalf("the content did not move: %v -> %v", before, moved)
			}
			if n > 4 {
				t.Fatalf("a marquee built the view %d times in 2.5s", n)
			}
			// A build in the middle of the motion shows the content where the
			// frames that only paint had moved it.
			tt.Frame()
			built := redEdges(tt.Image(), region, vertical)
			if len(built) != len(moved) {
				t.Fatalf("a build moved the content: %v -> %v", moved, built)
			}
			for i := range built {
				if d := built[i] - moved[i]; d < -1 || d > 1 {
					t.Fatalf("a build moved the content: %v -> %v", moved, built)
				}
			}
			if state.Offset == 0 {
				t.Fatal("a build did not bring Offset up to date")
			}
		})
	}
}

func TestMarqueeOutOfItsScrollAreaWaits(t *testing.T) {
	var state nui.MarqueeState
	builds := 0
	tt := ui.NewTester(func(c *ui.Context) {
		builds++
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c, th, nui.ScrollAreaOptions{}, func() {
			ui.Box(c.Key("spacer")).Size(10, 300)
			nui.Marquee(c.Key("marquee"), th, &state, nui.MarqueeOptions{Label: "Logos"}, func() { ui.Text(c, "Partner") }).Width(300)
		}).Height(100).Width(400)
	}, 500, 600)
	tt.Frame() // built again with its committed geometry
	tt.Advance(100 * time.Millisecond)
	// The marquee lies below the area's 100 DIPs, though inside the window.
	if n := buildsDuring(tt, &builds, 2*time.Second); n > 2 || tt.FramePending() {
		t.Fatalf("a marquee scrolled out of view built %d times in 2s (pending=%v)", n, tt.FramePending())
	}
}
