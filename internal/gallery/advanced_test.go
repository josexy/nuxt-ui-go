package gallery

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestGalleryFilesAreIndependentAndRulesApply(t *testing.T) {
	a, b := newStaticGallery(), newStaticGallery()
	t.Cleanup(a.Dispose)
	t.Cleanup(b.Dispose)
	ta, tb := sceneTester(t, a, "file-upload", "rules"), sceneTester(t, b, "file-upload", "rules")
	press(t, ta, "Add rules sample")
	requireText(t, ta, "review-1.md")
	if tb.HasText("review-1.md") {
		t.Fatal("windows share file state")
	}
	press(t, ta, "Add rejected executable")
	requireTextContains(t, ta, "application.exe")
	press(t, ta, "Remove review-1.md")
	requireText(t, ta, "File removed")
	if ta.HasText("review-1.md") {
		t.Fatal("file was not removed")
	}
	tb.Frame()
	if tb.HasText("application.exe") {
		t.Fatal("rule rejection leaked to another window")
	}
}
func TestGalleryMotionAndSplitterControls(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "carousel", "autoplay")
	tt.Advance(10 * time.Second)
	requireText(t, tt, "Slide 1 · 0 changes · paused false")
	press(t, tt, "Pause / resume autoplay")
	requireTextContains(t, tt, "paused true")
	tt = sceneTester(t, a, "splitter", "horizontal")
	focus(t, tt, "Divider")
	tt.Key(0, ui.KeyRight)
	requireTextContains(t, tt, "1 changes")
	press(t, tt, "horizontal pane draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Pane 中文 edit")
	tt.SetSize(720, 820)
	requireInput(t, tt, "horizontal pane draft", "Pane 中文 edit")
	tt = sceneTester(t, a, "marquee", "static")
	tt.Advance(10 * time.Second)
	if tt.FramePending() {
		t.Fatal("static marquee kept requesting frames")
	}
	press(t, tt, "Pause / resume static")
	requireText(t, tt, "Paused: true")
}

func TestGalleryNestedSplitters(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "splitter", "horizontal")
	bounds := func(label string) ui.Rect {
		t.Helper()
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("missing %q", label)
		}
		return r
	}
	checkLayout := func() {
		t.Helper()
		left, top, bottom := bounds("Left pane"), bounds("Right · top"), bounds("Right · bottom")
		if top.X <= left.X+left.W || top.Y != left.Y || bottom.X != top.X || bottom.Y <= top.Y+top.H {
			t.Fatalf("want a left pane beside stacked right panes: left=%v top=%v bottom=%v", left, top, bottom)
		}
	}
	checkLayout()
	outer, inner := bounds("horizontal pane widths"), bounds("horizontal pane heights")
	x, y := inner.X-.5, outer.Y+outer.H/2
	tt.Press(x, y)
	tt.Move(x+40, y)
	tt.Release(x+40, y)
	wider := bounds("horizontal pane heights")
	if wider.X <= inner.X || wider.Y != inner.Y || wider.H != inner.H {
		t.Fatal("outer divider did not resize the left pane independently")
	}
	requireTextContains(t, tt, "Right top height 55%")
	x, y = wider.X+wider.W/2, wider.Y+(wider.H-1)*.55+.5
	before := bounds("Right · bottom")
	tt.Press(x, y)
	tt.Move(x, y+25)
	tt.Release(x, y+25)
	tt.Key(0, ui.KeyDown)
	after := bounds("Right · bottom")
	if after.Y <= before.Y || bounds("horizontal pane heights").X != wider.X {
		t.Fatal("inner divider did not resize the stacked panes independently")
	}
	requireTextContains(t, tt, "3 changes")
	press(t, tt, "horizontal pane draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Nested preview 中文")
	requireText(t, tt, "Nested preview 中文")
	press(t, tt, "horizontal pane notes")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Separate notes")
	for _, width := range []int{1080, 720} {
		tt.SetSize(width, 820)
		checkLayout()
		requireInput(t, tt, "horizontal pane draft", "Nested preview 中文")
		requireInput(t, tt, "horizontal pane notes", "Separate notes")
		captureGallery(t, tt, fmt.Sprintf("splitter-nested-%d", width))
	}

	tt = sceneTester(t, a, "splitter", "disabled")
	outer, inner = bounds("disabled pane widths"), bounds("disabled pane heights")
	x, y = inner.X-.5, outer.Y+outer.H/2
	tt.Press(x, y)
	tt.Move(x+40, y)
	tt.Release(x+40, y)
	x, y = inner.X+inner.W/2, inner.Y+(inner.H-1)*.55+.5
	tt.Press(x, y)
	tt.Move(x, y+25)
	tt.Release(x, y+25)
	tt.Key(0, ui.KeyDown)
	requireText(t, tt, "Left width 40% · Right top height 55% · 0 changes")
}

func waitCommandText(t *testing.T, tt *ui.Tester, text string) {
	t.Helper()
	until := time.Now().Add(2 * time.Second)
	for !tt.HasText(text) && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
		tt.Frame()
	}
	requireText(t, tt, text)
}
func TestGalleryCommandNavigationAndClose(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "command-palette", "search")
	press(t, tt, "Open search commands")
	waitCommandText(t, tt, "Appearance")
	press(t, tt, "Appearance")
	waitCommandText(t, tt, "Use dark theme")
	press(t, tt, "Use dark theme")
	if !a.Dark {
		t.Fatal("command theme")
	}
	press(t, tt, "Open search commands")
	waitCommandText(t, tt, "Use dark theme")
	press(t, tt, "Back")
	waitCommandText(t, tt, "Open activity")
	press(t, tt, "Open activity")
	if a.ActiveID != "chart" {
		t.Fatal("command route")
	}
	tt.Frame()
	if tt.HasText("Search search workspace commands") {
		t.Fatal("command did not close")
	}
}
func TestGalleryDrawerCarouselFilesAndNavigation(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := sceneTester(t, a, "drawer", "basic")
	press(t, tt, "Open basic drawer")
	press(t, tt, "Resize basic inspector")
	tt.Key(0, ui.KeyEnd)
	requireText(t, tt, "Snap 1")
	tt.Key(0, ui.KeyEscape)
	if !tt.Focused("Open basic drawer") {
		t.Fatal("drawer focus")
	}
	tt = sceneTester(t, a, "carousel", "horizontal")
	press(t, tt, "Next horizontal review guide")
	requireText(t, tt, "Slide 2")
	press(t, tt, "Previous horizontal review guide")
	requireText(t, tt, "Slide 1")
	tt = sceneTester(t, a, "file-upload", "simulation")
	press(t, tt, "Add simulation sample")
	press(t, tt, "Start simulated upload")
	press(t, tt, "Advance progress")
	requireText(t, tt, "Simulated progress 25%")
	a.Deactivate()
	tt.Frame()
	press(t, tt, "Advance progress")
	requireText(t, tt, "Stale upload result rejected")
	a.WindowInactive = false
	press(t, tt, "Start simulated upload")
	press(t, tt, "Fail simulated upload")
	requireText(t, tt, "Simulation failed")
	tt = sceneTester(t, a, "navigation-menu", "vertical")
	press(t, tt, "Workspace settings")
	press(t, tt, "Members")
	requireText(t, tt, "Opened members")
}
