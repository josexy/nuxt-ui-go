package nui_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestContextMenuPointerRepositionAndControlledReopen(t *testing.T) {
	var state nui.MenuState
	var changes []bool
	actions := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		target := nui.Button(c.Key("target"), th, "Target", nui.ButtonOptions{}).Absolute().Left(40).Top(30).Width(300).Height(180)
		nui.ContextMenu(c.Key("menu"), th, target, &state, []nui.MenuItem{{ID: "first", Label: "First"}, {ID: "second", Label: "Second", OnSelect: func() { actions++ }}}, nui.MenuOptions{Label: "Context panel", Width: nui.Set(float32(180)), OnOpenChange: func(v bool) { changes = append(changes, v) }})
	}, 700, 500)
	tt.RightClickAt(70, 60)
	first := bounds(t, tt, "Context panel")
	if first.X != 72 || first.Y != 60 || !tt.Focused("First") {
		t.Fatalf("pointer placement/focus: %v", first)
	}
	tt.Key(0, ui.KeyDown)
	tt.RightClickAt(300, 190)
	second := bounds(t, tt, "Context panel")
	if !state.Open || second.X != 302 || second.Y != 190 || !tt.Focused("Second") || !slices.Equal(changes, []bool{true}) {
		t.Fatalf("reposition open=%v bounds=%v notifications=%v focus=%v", state.Open, second, changes, tt.Focused("Second"))
	}
	tt.Key(0, ui.KeyEnter)
	if actions != 1 || state.Open || !tt.Focused("Target") {
		t.Fatal("repositioned menu did not retain keyboard action and return focus")
	}
	state.Open = true
	tt.Frame()
	opened := bounds(t, tt, "Context panel")
	if opened.X != 342 || opened.Y != 30 || len(changes) != 2 {
		t.Fatalf("controlled open retained stale point or emitted: bounds=%v events=%v", opened, changes)
	}
	state.Open = false
	tt.Frame()
	if !tt.Focused("Target") || len(changes) != 2 {
		t.Fatal("controlled close emitted or lost focus")
	}
}

func TestContextMenuDisabledPreservesTargetAction(t *testing.T) {
	var state nui.MenuState
	disabled, targetDisabled := true, false
	clicks, changes := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		target := nui.Button(c, th, "Target", nui.ButtonOptions{Disabled: targetDisabled}).Width(200).Height(100).OnClick(func() { clicks++ })
		nui.ContextMenu(c, th, target, &state, []nui.MenuItem{{ID: "a", Label: "Action"}}, nui.MenuOptions{Disabled: disabled, OnOpenChange: func(bool) { changes++ }})
	}, 500, 400)
	click(t, tt, "Target")
	tt.Key(ui.Shift, ui.KeyF10)
	_ = tt.RightClick("Target")
	if clicks != 1 || state.Open || changes != 0 {
		t.Fatal("disabled context menu changed the target's own action")
	}
	disabled = false
	tt.Frame()
	_ = tt.RightClick("Target")
	if !state.Open || changes != 1 {
		t.Fatal("context menu did not recover from disabled")
	}
	disabled = true
	tt.Frame()
	if state.Open || changes != 1 {
		t.Fatal("external disabled change did not close silently")
	}
	click(t, tt, "Target")
	if clicks != 2 {
		t.Fatal("target action stayed disabled after closing the menu")
	}
	disabled = false
	targetDisabled = true
	tt.Frame()
	_ = tt.RightClick("Target")
	if state.Open {
		t.Fatal("disabled target opened a menu")
	}
}

func TestContextMenuKeyboardAndPointerWindowEdges(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		for _, width := range []float32{260, 650} {
			t.Run(fmt.Sprintf("%v/%v", rtl, width), func(t *testing.T) {
				var state nui.MenuState
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{Mode: nui.DarkMode})
					th.Motion.Reduce = true
					target := nui.Button(c, th, "Target", nui.ButtonOptions{}).Absolute().Left(20).Top(20).Width(width - 40).Height(250)
					nui.ContextMenu(c, th, target, &state, []nui.MenuItem{{ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}, {ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "json", Label: "JSON"}}}}, nui.MenuOptions{Label: "Context panel", Width: nui.Set(float32(200))})
				}, int(width), 300)
				if rtl {
					tt.SetDirection(ui.RightToLeft)
				}
				tt.RightClickAt(width-25, 265)
				r := bounds(t, tt, "Context panel")
				if r.X < 0 || r.Y < 0 || r.X+r.W > width || r.Y+r.H > 300 {
					t.Fatalf("point popup exceeds viewport: %v", r)
				}
				forward, back := ui.KeyRight, ui.KeyLeft
				if rtl {
					forward, back = back, forward
				}
				tt.Key(0, forward)
				if !tt.Focused("CSV") {
					t.Fatal("directional submenu opening failed")
				}
				tt.Key(0, ui.KeyDown)
				if !tt.Focused("JSON") {
					t.Fatal("submenu did not skip disabled item")
				}
				tt.Key(0, back)
				if !tt.Focused("Export") {
					t.Fatal("submenu did not return to parent")
				}
				tt.Key(0, ui.KeyEscape)
				if !tt.Focused("Target") {
					t.Fatal("Escape did not restore context target")
				}
				tt.Key(ui.Shift, ui.KeyF10)
				if !state.Open || !tt.Focused("Export") {
					t.Fatal("Shift+F10 did not open")
				}
				tt.Key(0, ui.KeyEscape)
				tt.Key(0, ui.KeyContextMenu)
				if !state.Open || !tt.Focused("Export") {
					t.Fatal("Menu key did not open")
				}
			})
		}
	}
}

func TestContextMenuPointerAnchorStaysAtWindowCoordinates(t *testing.T) {
	var state nui.MenuState
	x := float32(40)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c.Key("target"), th, "Target", nui.ButtonOptions{}).Absolute().Left(x).Top(30).Width(300).Height(180)
		nui.ContextMenu(c.Key("menu"), th, a, &state, []nui.MenuItem{{ID: "one", Label: "Action"}}, nui.MenuOptions{Label: "Context panel"})
	}, 700, 500)
	tt.RightClickAt(70, 60)
	before := bounds(t, tt, "Context panel")
	x = 140
	tt.Frame()
	if after := bounds(t, tt, "Context panel"); after.X != before.X || after.Y != before.Y {
		t.Fatalf("pointer menu moved with its target: before=%v after=%v", before, after)
	}
}

func TestContextMenuKeyboardUsesCommittedTargetBounds(t *testing.T) {
	var state nui.MenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c, th, "Target", nui.ButtonOptions{}).Absolute().Left(40).Top(30).Width(300).Height(100)
		nui.ContextMenu(c, th, a, &state, []nui.MenuItem{{ID: "a", Label: "Action"}}, nui.MenuOptions{Label: "Context panel"})
	}, 700, 500)
	click(t, tt, "Target")
	for _, key := range []ui.Key{ui.KeyContextMenu, ui.KeyF10} {
		mods := ui.Modifiers(0)
		if key == ui.KeyF10 {
			mods = ui.Shift
		}
		tt.Key(mods, key)
		if r := bounds(t, tt, "Context panel"); r.X != 42 || r.Y != 130 {
			t.Fatalf("keyboard point was not at the committed target bottom: %v", r)
		}
		tt.Key(0, ui.KeyEscape)
	}
}
