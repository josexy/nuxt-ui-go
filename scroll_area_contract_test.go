package nui_test

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestScrollAreaRoundedRootClipsContent(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).Children(func() {
			nui.ScrollArea(c, th, nui.ScrollAreaOptions{UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(16))}}}, func() { ui.Box(c).Width(200).Height(400).Background(ui.RGB(255, 0, 0)) }).Width(200).Height(120).Label("Rounded scroll")
		})
	}, 260, 170)
	r := bounds(t, tt, "Rounded scroll")
	if tt.Image().RGBAAt(int(r.X+1), int(r.Y+1)) == rgba(ui.RGB(255, 0, 0)) {
		t.Fatal("scroll content escaped Root radius")
	}
	if tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+6)) != rgba(ui.RGB(255, 0, 0)) {
		t.Fatal("rounded clipping removed visible content")
	}
}

func TestScrollAreaAxesAndStableState(t *testing.T) {
	state := ui.ScrollState{}
	direction := nui.ScrollBoth
	width := float32(6)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ScrollArea(c.Key("area"), nui.App(c, nui.AppConfig{}), nui.ScrollAreaOptions{Direction: direction, State: &state, ScrollbarWidth: nui.Set(width)}, func() { ui.Box(c.Key("canvas")).Width(800).Height(800) }).Width(240).Height(160).Label("Area")
	}, 320, 240)
	r := bounds(t, tt, "Area")
	tt.Scroll(r.X+80, r.Y+80, 100, 120)
	tt.Advance(time.Second)
	if state.X <= 0 || state.Y <= 0 || state.MaxX <= 0 || state.MaxY <= 0 {
		t.Fatalf("both axes not scrollable: %+v", state)
	}
	x, y := state.X, state.Y
	width = 14
	tt.Frame()
	if state.X != x || state.Y != y {
		t.Fatal("scrollbar styling reset offsets")
	}
	direction = nui.ScrollHorizontal
	tt.Frame()
	if state.X != x || state.Y != 0 {
		t.Fatalf("horizontal switch=%+v", state)
	}
	direction = nui.ScrollBoth
	state.X = 9000
	state.Y = 9000
	tt.Frame()
	if state.X != state.MaxX || state.Y != state.MaxY {
		t.Fatal("programmatic end was not clamped")
	}
	state.X, state.Y = 0, 0
	tt.Frame()
	if state.X != 0 || state.Y != 0 {
		t.Fatal("programmatic start failed")
	}
}

func TestScrollAreaNestedBoundaryPassesWheel(t *testing.T) {
	var outer, inner ui.ScrollState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c, th, nui.ScrollAreaOptions{State: &outer}, func() {
			nui.ScrollArea(c, th, nui.ScrollAreaOptions{State: &inner}, func() { ui.Box(c).Height(500) }).Height(100).Label("Inner")
			ui.Box(c).Height(800)
		}).Width(240).Height(220)
	}, 300, 280)
	r := bounds(t, tt, "Inner")
	tt.Scroll(r.X+70, r.Y+50, 0, 60)
	tt.Advance(time.Second)
	if inner.Y <= 0 || outer.Y != 0 {
		t.Fatal("inner wheel leaked to outer before its boundary")
	}
	inner.Y = inner.MaxY
	tt.Frame()
	tt.Scroll(r.X+70, r.Y+50, 0, 60)
	tt.Advance(time.Second)
	if outer.Y <= 0 {
		t.Fatal("inner boundary swallowed wheel instead of reaching outer")
	}
}

func TestScrollAreaThumbColorWidthAndZeroFallback(t *testing.T) {
	width := float32(14)
	state := ui.ScrollState{}
	marker := ui.RGB(255, 0, 255)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ScrollArea(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode}), nui.ScrollAreaOptions{State: &state, ScrollbarColor: nui.Set(marker), ScrollbarWidth: nui.Set(width)}, func() { ui.Box(c).Height(600) }).Width(240).Height(160).Label("Area")
	}, 300, 220)
	tt.Move(220, 20)
	strip := ui.Rect{X: 210, Y: 20, W: 30, H: 1}
	if got := countColor(tt.Image(), strip, rgba(marker)); got != 14 {
		t.Fatalf("custom thumb width/color=%d", got)
	}
	width = 0
	tt.Frame()
	if got := countColor(tt.Image(), strip, rgba(marker)); got != 6 {
		t.Fatalf("zero thumb width=%d, want native 6", got)
	}
	tt.Press(234, 20)
	tt.Move(234, 120)
	tt.Release(234, 120)
	if state.Y <= 0 {
		t.Fatal("dragging zero-fallback thumb did not scroll")
	}
}

func TestScrollAreaDirectionKeepsEditorIdentity(t *testing.T) {
	draft, direction := "Original", nui.ScrollBoth
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c.Key("area"), th, nui.ScrollAreaOptions{Direction: direction}, func() {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Draft"}).Width(200)
			ui.Box(c.Key("canvas")).Width(800).Height(600)
		}).Width(260).Height(180)
	}, 340, 240)
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	direction = nui.ScrollHorizontal
	tt.Frame()
	tt.SetDirection(ui.RightToLeft)
	if !tt.Focused("Draft") {
		t.Fatal("direction change lost the editor")
	}
	tt.Type("Replaced")
	if draft != "Replaced" {
		t.Fatal("direction change lost the selected text")
	}
}
