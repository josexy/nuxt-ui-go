package nui_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func awaitCommandResults(t *testing.T, tt *ui.Tester, ready func() bool) {
	t.Helper()
	until := time.Now().Add(2 * time.Second)
	for !ready() && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
		tt.Frame()
	}
	if !ready() {
		t.Fatal("worker did not publish")
	}
}
func TestCommandSearchAndCancellation(t *testing.T) {
	actions := make([]nui.CommandAction, 100000)
	for i := range actions {
		actions[i] = nui.CommandAction{ID: fmt.Sprint(i), Label: fmt.Sprintf("Build %06d", i)}
	}
	actions = append(actions, nui.CommandAction{ID: "unicode", Label: "打开工作区"})
	index := nui.NewCommandIndex(actions)
	matches := index.Search(context.Background(), "打工", "", 20)
	if len(matches) != 1 || fmt.Sprint(matches[0].Highlights) != "[0 2]" {
		t.Fatalf("rune highlights: %+v", matches)
	}
	if n := len(index.Search(context.Background(), "build", "", 100)); n != 100 {
		t.Fatal(n)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if len(index.Search(ctx, "build", "", 100)) != 0 {
		t.Fatal("canceled search")
	}
	var s nui.CommandPaletteState
	s.SetIndex(index)
	selected := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{OnSelect: func(id string) { selected = id }})
	}, 700, 600)
	tt.Key(ui.Ctrl, ui.KeyK)
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	click(t, tt, "Search Command palette")
	tt.Type("打工")
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if selected != "unicode" || s.Open {
		t.Fatalf("activation %q open %v", selected, s.Open)
	}
	s.Open = true
	tt.Frame()
	oldCtx, ticket := s.Results.Start(context.Background(), "old")
	s.Close()
	if oldCtx.Err() == nil || s.Results.Complete(ticket, matches, nil) {
		t.Fatal("close accepted old results")
	}
	tt.Frame()
}
func TestNestedCommandsAndShortcut(t *testing.T) {
	var s nui.CommandPaletteState
	called := 0
	s.SetIndex(nui.NewCommandIndex([]nui.CommandAction{{ID: "appearance", Label: "Appearance", Children: []nui.CommandAction{{ID: "dark", Label: "Dark", OnSelect: func() { called++ }}}}, {ID: "quick", Label: "Quick", Key: ui.KeyJ, Mods: ui.Ctrl, OnSelect: func() { called += 10 }}}))
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{})
	}, 700, 600)
	tt.Key(ui.Ctrl, ui.KeyJ)
	if called != 10 {
		t.Fatal("action shortcut")
	}
	tt.Key(ui.Ctrl, ui.KeyK)
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	click(t, tt, "Appearance")
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	click(t, tt, "Dark")
	if called != 11 || s.Open {
		t.Fatal("nested activation")
	}
}

func TestCommandPendingDoesNotActivateOldResult(t *testing.T) {
	var s nui.CommandPaletteState
	var ticket nui.QueryTicket
	var queryCtx context.Context
	called := 0
	s.Open = true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{Search: func(ctx context.Context, next nui.QueryTicket, query, parent string) { queryCtx, ticket = ctx, next }})
	}, 700, 600)
	if !s.Results.Complete(ticket, []nui.CommandMatch{{Action: nui.CommandAction{ID: "old", Label: "Old result", OnSelect: func() { called++ }}}}, nil) {
		t.Fatal("initial result")
	}
	tt.Frame()
	click(t, tt, "Search Command palette")
	tt.Type("new query")
	tt.Key(0, ui.KeyEnter)
	if called != 0 || !s.Open || !s.Results.Loading {
		t.Fatal("stale action activated")
	}
	old := ticket
	tt.Key(0, ui.KeyEscape)
	if queryCtx.Err() == nil || s.Results.Complete(old, nil, nil) {
		t.Fatal("external search survived close")
	}
}

func TestHorizontalNavigationKeyboard(t *testing.T) {
	active := "overview"
	var state nui.NavigationState
	items := []nui.NavigationItem{{ID: "overview", Label: "Overview"}, {ID: "tools", Label: "Tools", Children: []nui.NavigationItem{{ID: "export", Label: "Export report"}}}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.NavigationMenu(c, th, &active, &state, items, nui.NavigationOptions{})
	}, 500, 400)
	click(t, tt, "Overview")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Tools") {
		t.Fatal("horizontal focus group")
	}
	tt.Key(0, ui.KeyEnter)
	click(t, tt, "Export report")
	if active != "export" {
		t.Fatal("nested navigation callback")
	}
	tt.Frame()
	items = items[:1]
	tt.Frame()
}
func TestUploadsValidationDropAndTasks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.txt")
	if err := os.WriteFile(path, []byte("review"), 0600); err != nil {
		t.Fatal(err)
	}
	var s nui.FileUploadState
	rules := nui.UploadRules{Accept: []string{".txt"}, MaxBytes: 10, MaxFiles: 2}
	chooseCalls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FileUpload(c, th, &s, nui.FileUploadOptions{Label: "Attachments", Rules: rules, Choose: func(ctx context.Context) ([]nui.UploadFile, error) {
			return nui.InspectUploadFiles(ctx, []string{path})
		}, OnChange: func([]nui.UploadFile) { chooseCalls++ }}).Width(450)
	}, 500, 600)
	// The interactive area is the picker's button (D7).
	click(t, tt, "Attachments")
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	if len(s.Files) != 1 || s.Files[0].Status != nui.UploadReady || chooseCalls != 1 {
		t.Fatalf("picker %+v", s.Files)
	}
	file := s.Files[0]
	ctx, ticket := s.Begin(file.ID)
	s.Report(ticket, .5, false, nil)
	tt.Frame()
	click(t, tt, "Cancel review.txt")
	if ctx.Err() == nil || s.Report(ticket, 1, true, nil) {
		t.Fatal("canceled upload accepted completion")
	}
	_, retry := s.Begin(file.ID)
	if !s.Report(retry, .4, false, errors.New("Connection lost")) || s.Files[0].Status != nui.UploadFailed {
		t.Fatal("failure")
	}
	tt.Frame() // Publish the asynchronous status and changed button geometry.
	click(t, tt, "Remove review.txt")
	if len(s.Files) != 0 {
		t.Fatal("remove")
	}
	r, ok := tt.Find("Attachments")
	if !ok {
		t.Fatal("drop zone")
	}
	if !tt.DropFiles(r.X+r.W/2, r.Y+10, path) {
		t.Fatal("drop refused")
	}
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	if len(s.Files) != 1 {
		t.Fatal("dropped file")
	}
	s.Add([]nui.UploadFile{{Path: "big.txt", Name: "big.txt", Size: 11}, {Path: "image.png", Name: "image.png", Size: 2}, {Path: "missing.txt", Name: "missing.txt", Error: "Not found"}}, rules)
	for _, f := range s.Files[1:] {
		if f.Status != nui.UploadRejected || f.Error == "" {
			t.Fatal("validation", f)
		}
	}
	s.Add([]nui.UploadFile{file}, rules)
	if len(s.Files) != 4 {
		t.Fatal("duplicate paths")
	}
	s.Dispose()
}
func TestUploadDisposeIgnoresLatePicker(t *testing.T) {
	var s nui.FileUploadState
	started := make(chan struct{})
	release := make(chan struct{})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FileUpload(c, th, &s, nui.FileUploadOptions{Label: "Files", Choose: func(ctx context.Context) ([]nui.UploadFile, error) {
			close(started)
			<-release
			return []nui.UploadFile{{Name: "late.txt", Size: 1}}, nil
		}})
	}, 500, 400)
	click(t, tt, "Files")
	<-started
	s.Dispose()
	close(release)
	tt.Frame()
	if s.Pending() || len(s.Files) != 0 {
		t.Fatal("late picker retained")
	}
}
func TestChartScaleGapsAndKeyboard(t *testing.T) {
	var s nui.ChartState
	if err := s.SetData([]float64{0, 1, 2, 3, 4}, []nui.ChartSeries{{ID: "a", Label: "Builds", Values: []float64{1, 2, math.NaN(), -1, 3}}, {ID: "b", Label: "Reviews", Values: []float64{3, 2, 1, 2, 1}}}); err != nil {
		t.Fatal(err)
	}
	kind := nui.ChartLine
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Chart(c, th, &s, nui.ChartOptions{Label: "Activity", Kind: kind}).Width(600)
	}, 640, 400)
	click(t, tt, "Activity")
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyRight)
	if s.Selected != 1 {
		t.Fatalf("chart selected %d", s.Selected)
	}
	// Both series must paint: a second Draw callback must not replace the first.
	im := tt.Image()
	green, blue := 0, 0
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			r, g, b, _ := im.At(x, y).RGBA()
			if g > r*2 && g > b*13/10 {
				green++
			}
			if b > r*2 && b > g*11/10 {
				blue++
			}
		}
	}
	if green < 50 || blue < 50 {
		t.Fatalf("series colors green=%d blue=%d", green, blue)
	}
	for _, k := range []nui.ChartKind{nui.ChartBar, nui.ChartArea} {
		kind = k
		tt.Frame()
		_ = tt.Image()
	}
	if err := s.SetData([]float64{1, 0}, nil); err == nil {
		t.Fatal("unsorted X")
	}
	if err := s.SetData(nil, nil); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("No chart data"); !ok {
		t.Fatal("empty")
	}
}
func TestChartWorkload(t *testing.T) {
	for _, n := range []int{100, 10000, 100000} {
		var s nui.ChartState
		x, y := make([]float64, n), make([]float64, n)
		for i := range x {
			x[i] = float64(i)
			y[i] = math.Sin(float64(i) / 25)
		}
		if err := s.SetData(x, []nui.ChartSeries{{ID: "a", Values: y}}); err != nil {
			t.Fatal(err)
		}
		s.Prepare(640)
		if s.SampleCount() > 2564 {
			t.Fatalf("unbounded samples %d", s.SampleCount())
		}
		t.Logf("points=%d sampled=%d", n, s.SampleCount())
		for i := range y {
			if i%2 == 0 {
				y[i] = math.NaN()
			}
		}
		_ = s.SetData(x, []nui.ChartSeries{{ID: "a", Values: y}})
		s.Prepare(640)
		if s.SampleCount() > 2564 {
			t.Fatal("unbounded sparse samples")
		}
	}
}
func TestDrawerDragAndKeyboard(t *testing.T) {
	s := nui.DrawerState{Open: true}
	closed := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		nui.Drawer(c, th, &s, nui.DrawerOptions{Title: "Inspector", OnClose: func() { closed++ }, SnapPoints: []float32{.4, .8}}, func(t nui.Theme) {
			for range 20 {
				ui.Text(c, "Scrollable content")
			}
		})
	}, 800, 600)
	click(t, tt, "Resize Inspector")
	tt.Key(0, ui.KeyEnd)
	if s.Snap != 1 {
		t.Fatal("keyboard snap")
	}
	r, _ := tt.Find("Resize Inspector")
	tt.Press(r.X+20, r.Y+8)
	tt.Move(r.X+20, r.Y+160)
	tt.SetFocused(false)
	tt.SetFocused(true)
	tt.Frame()
	if !s.Open || s.Snap != 1 {
		t.Fatalf("canceled drag changed state: %+v", s)
	}
	click(t, tt, "Resize Inspector")
	tt.Key(0, ui.KeyHome)
	r, _ = tt.Find("Resize Inspector")
	tt.Press(r.X+20, r.Y+8)
	tt.Move(r.X+20, r.Y+180)
	tt.Advance(time.Second)
	tt.Release(r.X+20, r.Y+180)
	tt.Advance(time.Second)
	if s.Open || closed != 1 {
		t.Fatalf("drag close open=%v count=%d", s.Open, closed)
	}
	s.Open = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if s.Open {
		t.Fatal("escape")
	}
}
func TestCarouselMotionAndRemoval(t *testing.T) {
	var s nui.CarouselState
	active, show := true, true
	items := []nui.CarouselItem{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}, {ID: "three", Label: "Three"}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		if show {
			nui.Carousel(c, th, &s, items, nui.CarouselOptions{Label: "Guide", Autoplay: time.Second, Active: nui.Set(active), Wrap: true}).Width(400)
		}
	}, 500, 400)
	tt.Move(490, 390)
	tt.Frame()
	tt.Advance(1100 * time.Millisecond)
	if s.Index != 1 {
		t.Fatalf("autoplay %d", s.Index)
	}
	click(t, tt, "Pause Guide")
	index := s.Index
	tt.Move(490, 390)
	tt.Advance(2 * time.Second)
	if s.Index != index {
		t.Fatal("pause")
	}
	click(t, tt, "Guide slides")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyRight)
	if s.Index != 0 {
		t.Fatal("wrapped keyboard")
	}
	r, _ := tt.Find("Guide slides")
	tt.Press(r.X+300, r.Y+30)
	tt.Move(r.X+50, r.Y+30)
	tt.Advance(time.Second)
	tt.Release(r.X+50, r.Y+30)
	if s.Index != 1 {
		t.Fatalf("drag %d", s.Index)
	}
	s.Paused = false
	active = false
	tt.Frame()
	index = s.Index
	tt.Advance(3 * time.Second)
	if s.Index != index || tt.FramePending() {
		t.Fatal("inactive animation")
	}
	active = true
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Advance(3 * time.Second)
	if s.Index != index {
		t.Fatal("reduced motion")
	}
	items = items[:1]
	tt.Frame()
	if s.Index != 0 {
		t.Fatal("content removal")
	}
	show = false
	tt.Frame()
	tt.Advance(3 * time.Second)
	if tt.FramePending() {
		t.Fatal("unmounted carousel running")
	}
}
func TestMarqueeStops(t *testing.T) {
	var s nui.MarqueeState
	active, show := true, true
	key := "a"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		if show {
			nui.Marquee(c, th, &s, nui.MarqueeOptions{Label: "Updates", Active: nui.Set(active), ContentKey: key}, func() { ui.Text(c, strings.Repeat("Build passed   ", 30)).NoWrap() }).Width(300)
		}
	}, 400, 200)
	tt.Move(390, 190)
	tt.Frame()
	tt.Advance(time.Second)
	if s.Offset == 0 {
		t.Fatal("marquee did not advance")
	}
	s.Paused = true
	tt.Frame()
	before := s.Offset
	tt.Advance(time.Second)
	if s.Offset != before {
		t.Fatal("pause")
	}
	s.Paused = false
	active = false
	tt.Frame()
	tt.Advance(time.Second)
	if s.Offset != before || tt.FramePending() {
		t.Fatal("inactive")
	}
	key = "b"
	tt.Frame()
	if s.Offset != 0 {
		t.Fatal("content reset")
	}
	active = true
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Advance(time.Second)
	if s.Offset != 0 || tt.FramePending() {
		t.Fatal("reduced")
	}
	show = false
	tt.Frame()
	tt.Advance(time.Second)
	if tt.FramePending() {
		t.Fatal("unmounted")
	}
}
func TestNavigationSplitterAndLayouts(t *testing.T) {
	active := "one"
	var nav nui.NavigationState
	collapsed := false
	value := float32(50)
	count := 0
	items := []nui.NavigationItem{{ID: "one", Label: "Overview"}, {ID: "group", Label: "Manage", Children: []nui.NavigationItem{{ID: "two", Label: "Members"}, {ID: "disabled", Label: "Unavailable", Disabled: true}}}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Children(func() {
			nui.HeaderLayout(c, th, nui.HeaderOptions{Title: "Workspace with a long title"})
			ui.Row(c).Grow(1).MinHeight(0).Children(func() {
				nui.Sidebar(c, th, &collapsed, nui.SidebarOptions{Title: "Workspace"}, func(c *ui.Context, th nui.Theme, compact bool) {
					nui.NavigationMenu(c, th, &active, &nav, items, nui.NavigationOptions{Vertical: true, OnNavigate: func(string) { count++ }})
				})
				nui.Splitter(c, th, &value, nui.SplitterOptions{Percent: true, Min: nui.Set(float32(30)), Max: nui.Set(float32(70))}, func() { ui.Text(c, "First") }, func() { ui.Text(c, "Second") }).Grow(1)
			})
			nui.FooterLayout(c, th, nui.FooterOptions{Text: "Ready"})
		})
	}, 800, 500)
	click(t, tt, "Manage")
	click(t, tt, "Members")
	if active != "two" || count != 1 {
		t.Fatal("navigation")
	}
	click(t, tt, "Unavailable")
	if active != "two" {
		t.Fatal("disabled")
	}
	click(t, tt, "Toggle sidebar")
	if !collapsed {
		t.Fatal("sidebar")
	}
	click(t, tt, "Divider")
	for range 20 {
		tt.Key(0, ui.KeyRight)
	}
	if value > 70 {
		t.Fatalf("split limit %g", value)
	}
	for range 30 {
		tt.Key(0, ui.KeyLeft)
	}
	if value < 30 {
		t.Fatal("split min")
	}
	tt.SetSize(400, 600)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5})
	_ = tt.Image()
}
func BenchmarkCommandSearch100k(b *testing.B) {
	actions := make([]nui.CommandAction, 100000)
	for i := range actions {
		actions[i] = nui.CommandAction{ID: fmt.Sprint(i), Label: fmt.Sprintf("Workspace %06d", i)}
	}
	index := nui.NewCommandIndex(actions)
	b.ResetTimer()
	for b.Loop() {
		index.Search(context.Background(), "ws123", "", 200)
	}
}
func BenchmarkChartPrepare100k(b *testing.B) {
	x, y := make([]float64, 100000), make([]float64, 100000)
	for i := range x {
		x[i] = float64(i)
		y[i] = math.Sin(float64(i) / 25)
	}
	var s nui.ChartState
	_ = s.SetData(x, []nui.ChartSeries{{ID: "a", Values: y}})
	b.ResetTimer()
	for b.Loop() {
		s.Prepare(640)
		s.Prepare(641)
	}
}

func BenchmarkChartFrame(b *testing.B) {
	for _, count := range []int{100, 10000, 100000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			x, y := make([]float64, count), make([]float64, count)
			for i := range x {
				x[i], y[i] = float64(i), math.Sin(float64(i)/25)
			}
			var s nui.ChartState
			if err := s.SetData(x, []nui.ChartSeries{{ID: "a", Label: "Signal", Values: y}}); err != nil {
				b.Fatal(err)
			}
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Chart(c, th, &s, nui.ChartOptions{Label: "Signal", Height: 240}).Width(640)
			}, 700, 350)
			tt.Move(690, 340)
			tt.Frame()
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				tt.Frame()
			}
		})
	}
}
