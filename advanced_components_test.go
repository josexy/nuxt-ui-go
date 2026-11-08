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
	"github.com/josexy/nuxt-ui-go/icons/lucide"
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
	metadata := []nui.CommandAction{{ID: "description", Label: "Create", Description: "空白发票", Kbds: []string{"Ctrl", "N"}}, {ID: "suffix", Label: "Workspace", Suffix: "Local"}}
	metadataIndex := nui.NewCommandIndex(metadata)
	metadata[0].Description = "changed"
	metadata[0].Kbds[0] = "changed"
	if hits := metadataIndex.Search(context.Background(), "发票", "", 10); len(hits) != 1 || fmt.Sprint(hits[0].DescriptionHighlights) != "[2 3]" || hits[0].Action.Kbds[0] != "Ctrl" {
		t.Fatalf("description search or index ownership: %+v", hits)
	}
	if hits := metadataIndex.Search(context.Background(), "local", "", 10); len(hits) != 1 || len(hits[0].SuffixHighlights) != 5 {
		t.Fatalf("suffix search: %+v", hits)
	}
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
	limit := 10
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{Limit: limit, OnSelect: func(id string) { selected = id }})
	}, 700, 600)
	tt.Key(ui.Ctrl, ui.KeyK)
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	limit = 25
	tt.Frame()
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	if len(s.Results.Items) != 25 {
		t.Fatalf("updated command limit: %d", len(s.Results.Items))
	}
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
	var returnFocus ui.Handle
	called := 0
	keepOpen, prevent := false, false
	events := 0
	s.SetIndex(nui.NewCommandIndex([]nui.CommandAction{{ID: "appearance", Label: "Appearance", Children: []nui.CommandAction{{ID: "dark", Label: "Dark", OnSelect: func() { called++ }}}}, {ID: "quick", Label: "Quick", Key: ui.KeyJ, Mods: ui.Ctrl, OnSelect: func() { called += 10 }}}))
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c.Key("target"), th, "Focus destination", nui.ButtonOptions{}).Bind(&returnFocus)
		nui.Button(c.Key("opener"), th, "Open nested commands", nui.ButtonOptions{}).OnClick(func() { s.Open = true })
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{CloseOnSelect: nui.Set(!keepOpen), OnSelectEvent: func(e *nui.CommandEvent) {
			events++
			if prevent {
				e.PreventDefault()
			}
		}, Modal: nui.ModalOptions{ReturnFocus: &returnFocus}})
	}, 700, 600)
	tt.Key(ui.Ctrl, ui.KeyJ)
	if called != 10 {
		t.Fatal("action shortcut")
	}
	click(t, tt, "Open nested commands")
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	click(t, tt, "Appearance")
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	click(t, tt, "Dark")
	if called != 11 || s.Open {
		t.Fatal("nested activation")
	}
	tt.Frame()
	if !tt.Focused("Focus destination") {
		t.Fatal("palette skipped explicit return focus")
	}
	s.Open = true
	tt.Frame()
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	click(t, tt, "Search Command palette")
	tt.Key(0, ui.KeyBackspace)
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	if !tt.HasText("Appearance") {
		t.Fatal("empty Backspace did not return to parent commands")
	}
	keepOpen, prevent = true, true
	tt.Frame()
	before := called
	click(t, tt, "Quick")
	if called != before || !s.Open {
		t.Fatal("canceled command executed or closed")
	}
	prevent = false
	tt.Frame()
	click(t, tt, "Quick")
	click(t, tt, "Quick")
	if called != before+20 || !s.Open || events != 5 {
		t.Fatalf("repeat keep-open activation: calls=%d events=%d open=%v", called, events, s.Open)
	}
}

func TestCommandPendingDoesNotActivateOldResult(t *testing.T) {
	var s nui.CommandPaletteState
	var ticket nui.QueryTicket
	var queryCtx context.Context
	called := 0
	delay := time.Duration(0)
	disabled := false
	s.Open = true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{SearchDelay: delay, Disabled: disabled, Search: func(ctx context.Context, next nui.QueryTicket, query, parent string) { queryCtx, ticket = ctx, next }})
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
	s.Open = true
	tt.Frame()
	old = ticket
	s.Close()
	s.Open = true
	tt.Frame()
	if ticket == old || !s.Results.Loading {
		t.Fatal("same-frame close/reopen did not start a fresh search")
	}
	if !s.Results.Complete(ticket, []nui.CommandMatch{{Action: nui.CommandAction{ID: "parent", Label: "Parent command", Children: []nui.CommandAction{{ID: "leaf", Label: "Leaf command"}}}}}, nil) {
		t.Fatal("parent result rejected")
	}
	tt.Frame()
	click(t, tt, "Parent command")
	old = ticket
	s.Back()
	if queryCtx.Err() == nil || s.Results.Complete(old, nil, nil) {
		t.Fatal("back did not immediately cancel its child request")
	}
	tt.Frame()
	old = ticket
	oldCtx := queryCtx
	delay = 100 * time.Millisecond
	tt.Frame()
	click(t, tt, "Search Command palette")
	tt.Type("x")
	if oldCtx.Err() == nil || ticket != old || !s.Searching() {
		t.Fatal("debounce did not cancel immediately and defer dispatch")
	}
	tt.Advance(50 * time.Millisecond)
	tt.Type("y")
	tt.Advance(99 * time.Millisecond)
	if ticket != old {
		t.Fatal("debounce did not restart after another keystroke")
	}
	tt.Advance(time.Millisecond)
	tt.Frame()
	if ticket == old || ticket.Query != "xy" {
		t.Fatalf("trailing search was lost: ticket=%+v query=%q", ticket, s.Query)
	}
	disabled = true
	tt.Frame()
	if queryCtx.Err() == nil || s.Results.Complete(ticket, nil, nil) {
		t.Fatal("disabled palette accepted its pending result")
	}
	tt.Key(ui.Ctrl, ui.KeyK)
	if !s.Open {
		t.Fatal("disabled palette accepted its global shortcut")
	}
}

func TestHorizontalNavigationKeyboard(t *testing.T) {
	active := "overview"
	prevent := true
	navigations, events := 0, 0
	var state nui.NavigationState
	items := []nui.NavigationItem{{ID: "overview", Label: "Overview", Icon: lucide.House()}, {ID: "tools", Label: "Tools", Children: []nui.NavigationItem{{ID: "export", Label: "Export report"}}}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.NavigationMenu(c, th, &active, &state, items, nui.NavigationOptions{Compact: true, OnSelect: func(e *nui.NavigationEvent) {
			events++
			if prevent {
				e.PreventDefault()
			}
		}, OnNavigate: func(string) { navigations++ }})
	}, 500, 400)
	if bounds(t, tt, "Overview").W < 70 {
		t.Fatal("horizontal navigation incorrectly hid its label in Compact mode")
	}
	click(t, tt, "Overview")
	events = 0
	tt.Key(0, ui.KeyLeft)
	if !tt.Focused("Overview") {
		t.Fatal("horizontal navigation wrapped at its start")
	}
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Tools") {
		t.Fatal("horizontal focus group")
	}
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Tools") {
		t.Fatal("horizontal navigation wrapped at its end")
	}
	tt.Key(0, ui.KeyEnter)
	click(t, tt, "Export report")
	if active != "overview" || !tt.HasText("Export report") {
		t.Fatal("canceled navigation changed destination or closed its menu")
	}
	prevent = false
	tt.Frame()
	click(t, tt, "Export report")
	if active != "export" || navigations != 1 || events != 2 {
		t.Fatalf("nested navigation callback: active=%s navigations=%d events=%d", active, navigations, events)
	}
	tt.Frame()
	items = items[:1]
	tt.Frame()
	if !tt.Focused("Overview") {
		t.Fatal("removing the focused navigation item lost focus")
	}
}
func TestUploadsValidationDropAndTasks(t *testing.T) {
	for _, accept := range [][]string{{"*"}, {"*/*"}, {" , "}, {".tar.gz"}, {".png, .TAR.GZ"}, {"application/*"}} {
		if err := (nui.UploadRules{Accept: accept}).Validate(nui.UploadFile{Name: "ARCHIVE.TAR.GZ", MIME: "application/gzip; charset=binary", Size: 1}); err != nil {
			t.Fatalf("accepted type %v: %v", accept, err)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "review.txt")
	if err := os.WriteFile(path, []byte("review"), 0600); err != nil {
		t.Fatal(err)
	}
	var s nui.FileUploadState
	rules := nui.UploadRules{Accept: []string{".txt"}, MaxBytes: 10, MaxFiles: 2}
	chooseCalls, rejected, chooses := 0, 0, 0
	dropzone, empty := true, false
	inside := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		position := nui.UploadOutside
		if inside {
			position = nui.UploadInside
		}
		nui.FileUpload(c, th, &s, nui.FileUploadOptions{Label: "Attachments", Rules: rules, Layout: nui.UploadList, Position: position, Dropzone: nui.Set(dropzone), Choose: func(ctx context.Context) ([]nui.UploadFile, error) {
			chooses++
			if empty {
				return nil, nil
			}
			return nui.InspectUploadFiles(ctx, []string{path})
		}, OnChange: func([]nui.UploadFile) { chooseCalls++ }, OnReject: func(files []nui.UploadFile) { rejected += len(files) }}).Width(450)
	}, 500, 600)
	// The interactive area is the picker's button (D7).
	click(t, tt, "Attachments")
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	if len(s.Files) != 1 || s.Files[0].Status != nui.UploadReady || chooseCalls != 1 {
		t.Fatalf("picker %+v", s.Files)
	}
	click(t, tt, "Attachments")
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	empty = true
	click(t, tt, "Attachments")
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	if chooseCalls != 1 || chooses != 3 {
		t.Fatalf("duplicate/canceled picker emitted change: %d/%d", chooseCalls, chooses)
	}
	empty = false
	dropzone = false
	tt.Frame()
	r, _ := tt.Find("Attachments")
	if tt.DropFiles(r.X+r.W/2, r.Y+10, path) {
		t.Fatal("disabled dropzone registered a target")
	}
	dropzone, inside = true, true
	tt.Frame()
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
	if !tt.Focused("Attachments") {
		t.Fatal("removing a file did not return focus to picker")
	}
	if chooses != 3 {
		t.Fatal("nested file action reopened picker")
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
	rules.MaxFiles = 1
	other := filepath.Join(dir, "extra.txt")
	if err := os.WriteFile(other, []byte("extra"), 0600); err != nil {
		t.Fatal(err)
	}
	tt.DropFiles(r.X+r.W/2, r.Y+10, other)
	awaitCommandResults(t, tt, func() bool { return !s.Pending() })
	if len(s.Files) != 2 || rejected != 1 || s.Files[1].Status != nui.UploadRejected {
		t.Fatalf("count limit/reject: %+v, %d", s.Files, rejected)
	}
	s.Remove(s.Files[1].ID)
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
	var seeded nui.FileUploadState
	seeded.Files = []nui.UploadFile{{ID: "file-1", Name: "existing.txt"}}
	seeded.Add([]nui.UploadFile{{Name: "new.txt"}}, nui.UploadRules{})
	if seeded.Files[0].ID == seeded.Files[1].ID {
		t.Fatal("generated identity collides with supplied metadata")
	}
}
func TestUploadDisposeIgnoresLatePicker(t *testing.T) {
	for _, mode := range []string{"dispose", "disabled", "parent"} {
		t.Run(mode, func(t *testing.T) {
			var s nui.FileUploadState
			started := make(chan struct{})
			release := make(chan struct{})
			returned := make(chan struct{})
			disabled, changes := false, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Disabled(disabled && mode == "parent").Children(func() {
					nui.FileUpload(c, th, &s, nui.FileUploadOptions{Label: "Files", Disabled: disabled && mode == "disabled", OnChange: func([]nui.UploadFile) { changes++ }, Choose: func(ctx context.Context) ([]nui.UploadFile, error) {
						close(started)
						<-release
						defer close(returned)
						return []nui.UploadFile{{Name: "late.txt", Size: 1}}, nil
					}})
				})
			}, 500, 400)
			click(t, tt, "Files")
			<-started
			if mode == "dispose" {
				s.Dispose()
			} else {
				disabled = true
				tt.Frame()
			}
			close(release)
			<-returned
			tt.Frame()
			if s.Pending() || len(s.Files) != 0 || changes != 0 {
				t.Fatal("late picker retained")
			}
		})
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
	snaps, prevented := 0, 0
	dismissible := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		nui.Drawer(c, th, &s, nui.DrawerOptions{Title: "Inspector", Dismissible: nui.Set(dismissible), OnClose: func() { closed++ }, OnClosePrevent: func() { prevented++ }, OnSnap: func(int) { snaps++ }, SnapPoints: []float32{.4, .8}}, func(t nui.Theme) {
			for range 20 {
				ui.Text(c, "Scrollable content")
			}
		})
	}, 800, 600)
	click(t, tt, "Resize Inspector")
	tt.Key(0, ui.KeyHome)
	if snaps != 0 {
		t.Fatal("unchanged drawer snap emitted a notification")
	}
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
	dismissible = false
	tt.Frame()
	r, _ = tt.Find("Resize Inspector")
	tt.Press(r.X+20, r.Y+8)
	tt.Move(r.X+20, r.Y+180)
	tt.Advance(time.Second)
	tt.Release(r.X+20, r.Y+180)
	tt.Advance(time.Second)
	if !s.Open || closed != 0 || prevented != 1 {
		t.Fatalf("prevented drawer drag: open=%v closed=%d prevented=%d", s.Open, closed, prevented)
	}
	dismissible = true
	tt.Frame()
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
	watch, stop := true, true
	changes := 0
	items := []nui.CarouselItem{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}, {ID: "three", Label: "Three"}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		if show {
			nui.Carousel(c, th, &s, items, nui.CarouselOptions{Label: "Guide", Autoplay: time.Second, Active: nui.Set(active), Wrap: true, WatchDrag: nui.Set(watch), StopOnInteraction: nui.Set(stop), OnChange: func(int) { changes++ }}).Width(400)
		}
	}, 500, 400)
	tt.Move(490, 390)
	tt.Frame()
	tt.Advance(1100 * time.Millisecond)
	if s.Index != 1 {
		t.Fatalf("autoplay %d", s.Index)
	}
	viewport, _ := tt.Find("Guide slides")
	slide, _ := tt.Find("Two")
	if changes != 1 || slide.X < viewport.X || slide.X+slide.W > viewport.X+viewport.W+1 {
		t.Fatalf("autoplay model/display mismatch: changes=%d slide=%+v view=%+v", changes, slide, viewport)
	}
	click(t, tt, "Pause Guide")
	index := s.Index
	tt.Move(490, 390)
	tt.Advance(2 * time.Second)
	if s.Index != index {
		t.Fatal("pause")
	}
	s.Paused = false
	click(t, tt, "Guide slides")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyRight)
	if s.Index != 0 {
		t.Fatal("wrapped keyboard")
	}
	if !s.Paused {
		t.Fatal("navigation did not stop autoplay")
	}
	r, _ := tt.Find("Guide slides")
	tt.Press(r.X+300, r.Y+30)
	tt.Move(r.X+50, r.Y+30)
	tt.Advance(time.Second)
	tt.Release(r.X+50, r.Y+30)
	if s.Index != 1 {
		t.Fatalf("drag %d", s.Index)
	}
	tt.Key(0, ui.KeyEnd)
	tt.Press(r.X+300, r.Y+30)
	tt.Move(r.X+50, r.Y+30)
	tt.Advance(time.Second)
	tt.Release(r.X+50, r.Y+30)
	if s.Index != 2 {
		t.Fatalf("drag wrapped at the final slide: %d", s.Index)
	}
	watch = false
	tt.Frame()
	tt.Press(r.X+50, r.Y+30)
	tt.Move(r.X+300, r.Y+30)
	tt.Release(r.X+300, r.Y+30)
	if s.Index != 2 {
		t.Fatal("disabled WatchDrag changed slide")
	}
	watch, stop = true, false
	s.Paused = false
	tt.Frame()
	tt.Key(0, ui.KeyHome)
	if s.Paused {
		t.Fatal("StopOnInteraction=false still stopped autoplay")
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
	draft := "kept"
	items[2].Build = func(c *ui.Context, th nui.Theme) { nui.Input(c, th, &draft, nui.InputOptions{Label: "Third draft"}) }
	s.Index = 2
	tt.Frame()
	click(t, tt, "Third draft")
	beforeRemoval := changes
	items = items[:1]
	tt.Frame()
	if s.Index != 0 {
		t.Fatal("content removal")
	}
	if !tt.Focused("Guide slides") || changes != beforeRemoval {
		t.Fatal("external removal lost focus or emitted a navigation event")
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
	disabled := false
	key := "a"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		if show {
			ui.Box(c).Disabled(disabled).Children(func() {
				nui.Marquee(c, th, &s, nui.MarqueeOptions{Label: "Updates", Active: nui.Set(active), ContentKey: key}, func() { ui.Text(c, strings.Repeat("Build passed   ", 30)).NoWrap() }).Width(300)
			})
		}
	}, 400, 200)
	tt.Move(390, 190)
	tt.Frame()
	tt.Advance(time.Second)
	tt.Frame() // Offset is the phase as of the last build.
	if s.Offset == 0 {
		t.Fatal("marquee did not advance")
	}
	disabled = true
	tt.Frame()
	disabledOffset := s.Offset
	tt.Advance(time.Second)
	if s.Offset != disabledOffset || tt.FramePending() {
		t.Fatal("disabled ancestor did not stop the marquee")
	}
	disabled = false
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
	splitChanges := 0
	items := []nui.NavigationItem{{ID: "one", Label: "Overview"}, {ID: "group", Label: "Manage", Children: []nui.NavigationItem{{ID: "two", Label: "Members"}, {ID: "disabled", Label: "Unavailable", Disabled: true}}}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Children(func() {
			nui.HeaderLayout(c, th, nui.HeaderOptions{Title: "Workspace with a long title"})
			ui.Row(c).Grow(1).MinHeight(0).Children(func() {
				nui.Sidebar(c, th, &collapsed, nui.SidebarOptions{Title: "Workspace"}, func(c *ui.Context, th nui.Theme, compact bool) {
					nui.NavigationMenu(c, th, &active, &nav, items, nui.NavigationOptions{Vertical: true, OnNavigate: func(string) { count++ }})
				})
				nui.Splitter(c, th, &value, nui.SplitterOptions{Percent: true, Min: nui.Set(float32(30)), Max: nui.Set(float32(70)), OnChange: func(float32) { splitChanges++ }}, func() { ui.Text(c, "First") }, func() { ui.Text(c, "Second") }).Grow(1)
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
	click(t, tt, "Members")
	nav.Expanded["group"] = false
	tt.Frame()
	if !tt.Focused("Manage") {
		t.Fatal("external collapse lost child focus")
	}
	click(t, tt, "Toggle sidebar")
	if !collapsed {
		t.Fatal("sidebar")
	}
	// The focusable divider is one DIP wide; pressing its wider grip focuses it.
	clickOver(t, tt, "Divider")
	for range 20 {
		tt.Key(0, ui.KeyRight)
	}
	if value > 70 {
		t.Fatalf("split limit %g", value)
	}
	atLimit := splitChanges
	tt.Key(0, ui.KeyRight)
	if splitChanges != atLimit {
		t.Fatal("splitter notified repeatedly at Max")
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
	t.Run("nested splitter junction", func(t *testing.T) {
		for _, vert := range []bool{false, true} {
			for _, rtl := range []bool{false, true} {
				outer, inner := float32(40), float32(55)
				style := func(label string) func(nui.Slot, ui.Element) {
					return func(slot nui.Slot, e ui.Element) {
						if slot == nui.Indicator {
							e.Label(label)
						}
					}
				}
				p := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					nui.Splitter(c.Key("outer"), th, &outer, nui.SplitterOptions{Percent: true, Vertical: vert, Decoration: nui.Decoration{Style: style("Outer line")}}, nil, func() {
						nui.Splitter(c.Key("inner"), th, &inner, nui.SplitterOptions{Percent: true, Vertical: !vert, Decoration: nui.Decoration{Style: style("Inner line")}}, nil, nil).Fill()
					}).Fill()
				}, 500, 300)
				if rtl {
					p.SetDirection(ui.RightToLeft)
				}
				p.Frame()
				a, b := bounds(t, p, "Outer line"), bounds(t, p, "Inner line")
				gap := b.X - (a.X + a.W)
				if vert {
					gap = b.Y - (a.Y + a.H)
				} else if rtl {
					gap = a.X - (b.X + b.W)
				}
				if math.Abs(float64(gap)) > .01 {
					t.Fatalf("nested lines leave a gap: vertical=%t rtl=%t gap=%g", vert, rtl, gap)
				}
				for _, offset := range []float32{-2, 2} {
					a = bounds(t, p, "Outer line")
					x, y, dx, dy := a.X+a.W/2+offset, a.Y+10, float32(12), float32(0)
					if vert {
						x, y, dx, dy = a.X+10, a.Y+a.H/2+offset, 0, 12
					}
					before := outer
					p.Press(x, y)
					p.Move(x+dx, y+dy)
					p.Release(x+dx, y+dy)
					if outer == before || inner != 55 {
						t.Fatalf("divider hit area failed: vertical=%t rtl=%t offset=%g", vert, rtl, offset)
					}
				}
				parityScreenshot(t, p, fmt.Sprintf("splitter-junction-vertical-%t-rtl-%t", vert, rtl))
			}
		}
	})
	t.Run("splitter contract", func(t *testing.T) {
		value, changes := float32(50), 0
		disabled, parentDisabled, vert := true, false, false
		draft := "draft"
		step := float32(5)
		lo, hi := float32(30), float32(70)
		var editor ui.Handle
		p := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Box(c).Fill().Disabled(parentDisabled).Children(func() {
				nui.Splitter(c.Key("split"), th, &value, nui.SplitterOptions{
					Percent: true, Vertical: vert, Disabled: disabled, Min: nui.Set(lo), Max: nui.Set(hi), KeyboardResizeBy: nui.Set(step),
					OnChange: func(float32) { changes++ },
				}, func() {
					nui.Input(c.Key("editor"), th, &draft, nui.InputOptions{Label: "Pane editor"}).Bind(&editor)
				}, func() { ui.Text(c, "Preview").Label("Second pane") }).Fill()
			})
		}, 500, 300)
		click(t, p, "Pane editor")
		p.Key(ui.Ctrl, ui.KeyA)
		p.Type("editable while locked")
		if draft != "editable while locked" {
			t.Fatalf("Disabled also disabled the panel: %q", draft)
		}
		clickOver(t, p, "Divider") // its grip takes the press
		p.Key(0, ui.KeyRight)
		if value != 50 || changes != 0 {
			t.Fatal("disabled divider moved")
		}
		disabled = false
		p.Frame()
		clickOver(t, p, "Divider")
		p.Key(0, ui.KeyRight)
		if math.Abs(float64(value-55)) > .01 || changes != 1 {
			t.Fatalf("keyboard step: %g, %d changes", value, changes)
		}
		p.Key(0, ui.KeyEnd)
		before := changes
		for range 3 {
			p.Key(0, ui.KeyRight)
		}
		if value != 70 || changes != before {
			t.Fatalf("boundary repeated change: %g, %d -> %d", value, before, changes)
		}
		p.SetDirection(ui.RightToLeft)
		p.Key(0, ui.KeyRight)
		if math.Abs(float64(value-65)) > .01 {
			t.Fatalf("RTL physical right: %g", value)
		}
		vert = true
		p.Frame()
		p.Key(0, ui.KeyUp)
		if math.Abs(float64(value-60)) > .01 {
			t.Fatalf("vertical up: %g", value)
		}
		step = 0
		p.Frame()
		before = changes
		beforeValue := value
		p.Key(0, ui.KeyDown)
		if value != beforeValue || changes != before {
			t.Fatal("explicit zero arrow step moved")
		}
		click(t, p, "Pane editor")
		vert = false
		p.SetSize(300, 200)
		p.SetPreferences(ui.Preferences{TextScale: 1.5})
		if !p.Focused("Pane editor") || draft != "editable while locked" {
			t.Fatal("layout change lost pane draft or focus")
		}
		parentDisabled = true
		p.Frame()
		click(t, p, "Pane editor")
		p.Type("blocked")
		if draft != "editable while locked" {
			t.Fatal("disabled ancestor did not disable contents")
		}
		lo, hi, value = 0, 0, 50
		p.Frame()
		if value != 0 || changes != before {
			t.Fatal("explicit zero range or external clamp emitted a change")
		}
	})
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
