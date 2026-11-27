package nui_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestDropdownMenuSizesAndDescriptionAlignment(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			var state nui.MenuState
			state.SetItems([]nui.MenuItem{{ID: "heading", Kind: nui.MenuLabel, Label: "Heading", Description: "Group description", Avatar: &nui.AvatarOptions{Alt: "Group avatar"}}, {ID: "rich", Label: "Rich action", Description: "Second line", Icon: lucide.Info(), Shortcut: "Ctrl+K"}, {ID: "plain", Label: "Plain action"}})
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				th.Motion.Reduce = true
				a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(20)
				leadingIndex := 0
				nui.DropdownMenu(c, th, a, &state, nui.MenuOptions{Size: size, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Leading {
						e.Label(fmt.Sprintf("Leading %d", leadingIndex))
						leadingIndex++
					}
					if slot == nui.Trailing {
						e.Label("Keys")
					}
				}})
			}, 460, 380)
			click(t, tt, "Actions")
			row, icon, keys := bounds(t, tt, "Rich action"), bounds(t, tt, "Leading 1"), bounds(t, tt, "Keys")
			pad := []float32{4, 6, 6, 8, 8}[i]
			line := []float32{16, 16, 20, 20, 24}[i]
			if row.H != 2*line+2*pad || keys.Y != row.Y+pad || keys.H != line || icon.Y != row.Y+pad {
				t.Fatalf("description alignment: row=%+v keys=%+v icon=%+v", row, keys, icon)
			}
			if got := bounds(t, tt, "Plain action").H; got != line+2*pad {
				t.Fatalf("plain row height %v", got)
			}
			if got := bounds(t, tt, "Leading 0").W; got != line {
				t.Fatalf("heading avatar diameter %v", got)
			}
		})
	}
}

func TestDropdownMenuLoopCheckedAndCallbacks(t *testing.T) {
	var state nui.MenuState
	checked, loop, swap, disabled := false, false, false, false
	selected, updated, triggerClicks := 0, 0, 0
	var opens []bool
	var submit func()
	submit = func() {
		items := []nui.MenuItem{{ID: "first", Label: "First"}, {ID: "disabled", Label: "Disabled item", Disabled: true}, {ID: "check", Label: "Pinned", Checked: nui.Set(checked), Close: nui.Set(false), OnCheckedChange: func(v bool) { checked = v; updated++; submit() }, OnSelect: func() { selected++ }}, {ID: "last", Label: "Last"}}
		if swap {
			items[0], items[2] = items[2], items[0]
		}
		state.SetItems(items)
	}
	submit()
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			a := nui.Button(c.Key("anchor"), th, "Actions", nui.ButtonOptions{}).OnClick(func() { triggerClicks++ })
			nui.DropdownMenu(c.Key("menu"), th, a, &state, nui.MenuOptions{Loop: loop, Disabled: disabled, OnOpenChange: func(v bool) { opens = append(opens, v) }})
			nui.Button(c.Key("after"), th, "After", nui.ButtonOptions{})
		})
	}, 500, 400)
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyUp)
	if !tt.Focused("First") {
		t.Fatal("default navigation wrapped")
	}
	loop = true
	tt.Frame()
	tt.Key(0, ui.KeyUp)
	if !tt.Focused("Last") {
		t.Fatal("Loop did not wrap")
	}
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeySpace)
	if !checked || !state.Open || updated != 1 || selected != 1 {
		t.Fatalf("checked/open/events: %v/%v/%d/%d", checked, state.Open, updated, selected)
	}
	swap = true
	submit()
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if checked || updated != 2 || selected != 2 || !tt.Focused("Pinned") {
		t.Fatal("stable checked action lost focus or emitted twice")
	}
	tt.Key(0, ui.KeyEscape)
	if state.Open || !tt.Focused("Actions") || !slices.Equal(opens, []bool{true, false}) || triggerClicks != 1 {
		t.Fatalf("close/trigger notifications: %v %d", opens, triggerClicks)
	}
	state.Open = true
	tt.Frame()
	state.Open = false
	tt.Frame()
	if !tt.Focused("Actions") || len(opens) != 2 {
		t.Fatal("external close did not restore focus or emitted interaction")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Actions")
	if state.Open || triggerClicks != 1 {
		t.Fatal("disabled trigger activated")
	}
}

func TestDropdownMenuFilterAndExternalUpdates(t *testing.T) {
	var state nui.MenuState
	query := ""
	ignored, filter := false, true
	changes, selections := 0, 0
	var opens []bool
	state.SetItems([]nui.MenuItem{{ID: "label", Label: "Workspace", Kind: nui.MenuLabel}, {ID: "alpha", Label: "Alpha", Description: "first project"}, {ID: "blocked", Label: "Blocked", Description: "copy", Disabled: true}, {ID: "copy", Label: "Duplicate", Description: "copy project", OnSelect: func() {
		selections++
	}}, {ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}, {ID: "json", Label: "JSON"}}, Content: &nui.MenuContentOptions{Filter: true, FilterInput: &nui.InputOptions{Label: "Export search"}}}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			a := nui.Button(c.Key("anchor"), th, "Actions", nui.ButtonOptions{})
			nui.DropdownMenu(c.Key("menu"), th, a, &state, nui.MenuOptions{Filter: filter, SearchTerm: &query, FilterFields: []string{"label", "description"}, IgnoreFilter: ignored, FilterInput: &nui.InputOptions{Label: "Action search", Clearable: true}, OnSearchTermChange: func(string) { changes++ }, OnOpenChange: func(v bool) { opens = append(opens, v) }})
			nui.Button(c.Key("after"), th, "After", nui.ButtonOptions{})
		})
	}, 520, 440)
	click(t, tt, "Actions")
	if !tt.Focused("Action search") {
		t.Fatal("filter did not receive initial focus")
	}
	tt.Type("copy")
	if query != "copy" || changes != 1 || tt.HasText("Alpha") || !tt.HasText("Duplicate") {
		t.Fatalf("filter query=%q changes=%d texts=%v", query, changes, tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	if selections != 1 || state.Open {
		t.Fatal("Enter did not skip disabled match")
	}
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyEscape)
	if query != "" || !state.Open || changes != 2 {
		t.Fatal("Escape did not clear only the query")
	}
	tt.Type("unmatched")
	if !tt.HasText("No matching data") || tt.HasText("Workspace") {
		t.Fatal("filtered empty state retained structural items")
	}
	ignored = true
	tt.Frame()
	if !tt.HasText("Alpha") {
		t.Fatal("external filtering ignored the supplied collection")
	}
	query = ""
	ignored = false
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Type("exp")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Export search") {
		t.Fatal("submenu filter did not receive keyboard focus")
	}
	tt.Type("json")
	tt.Key(0, ui.KeyEscape)
	if !state.Open || !tt.HasText("CSV") {
		t.Fatal("submenu Escape failed to clear query")
	}
	tt.Key(0, ui.KeyEscape)
	if !tt.Focused("Export") || tt.HasText("JSON") {
		t.Fatal("submenu Escape did not return focus")
	}
	query = "first"
	tt.Frame()
	if !tt.Focused("Action search") || tt.HasText("Export") {
		t.Fatal("external filter removal did not repair focus")
	}
	filter = false
	tt.Frame()
	filter = true
	tt.Frame()
	if !tt.Focused("Action search") {
		t.Fatal("filter re-enable did not restore input focus")
	}
	tt.Key(0, ui.KeyTab)
	if state.Open || !tt.Focused("After") {
		t.Fatal("filter Tab did not leave the menu")
	}
	if !slices.Equal(opens, []bool{true, false, true, false}) {
		t.Fatalf("open notifications=%v", opens)
	}
}

func TestDropdownMenuOutsideAndEmptyDismissal(t *testing.T) {
	var state nui.MenuState
	outside, changes := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			a := nui.Button(c.Key("trigger"), th, "Actions", nui.ButtonOptions{})
			nui.DropdownMenu(c.Key("menu"), th, a, &state, nui.MenuOptions{OnOpenChange: func(bool) { changes++ }})
			nui.Button(c.Key("outside"), th, "Outside", nui.ButtonOptions{}).Margin(180, 0, 0, 0).OnClick(func() { outside++ })
		})
	}, 500, 400)
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyEscape)
	if state.Open || !tt.Focused("Actions") || changes != 2 {
		t.Fatalf("empty dismissal %v/%d", state.Open, changes)
	}
	state.SetItems([]nui.MenuItem{{ID: "first", Label: "First"}})
	click(t, tt, "Actions")
	click(t, tt, "Outside")
	if state.Open || outside != 1 || changes != 4 || !tt.Focused("Outside") {
		t.Fatalf("outside dismissal %v/%d/%d", state.Open, outside, changes)
	}
	tt.Advance(time.Second)
	if changes != 4 || tt.FramePending() {
		t.Fatal("closed menu did not settle")
	}
}

func TestDropdownMenuFilterHoverKeepsEditorAndChildFocus(t *testing.T) {
	var state nui.MenuState
	state.SetItems([]nui.MenuItem{{ID: "first", Label: "Alpha"}, {ID: "child", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}, {ID: "json", Label: "JSON"}}, Content: &nui.MenuContentOptions{Filter: true, FilterInput: &nui.InputOptions{Label: "Child search"}}}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(20)
		nui.DropdownMenu(c, th, a, &state, nui.MenuOptions{Filter: true, FilterInput: &nui.InputOptions{Label: "Root search"}})
	}, 560, 400)
	click(t, tt, "Actions")
	r := bounds(t, tt, "Export")
	tt.Move(r.X+3, r.Y+3)
	tt.Advance(200 * time.Millisecond)
	if !tt.HasText("Child search") || !tt.Focused("Root search") {
		t.Fatal("hover opening stole filter focus")
	}
	tt.Type("al")
	if state.Query != "al" || tt.HasText("Export") || tt.HasText("Child search") {
		t.Fatalf("hover disrupted filtering: query=%q texts=%v", state.Query, tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	tt.Key(0, ui.KeyDown)
	tt.Type("exp")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Child search") {
		t.Fatal("keyboard submenu did not focus its filter")
	}
}

func TestDropdownMenuLoadingIconsStylesAndDarkRTL(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			var state nui.MenuState
			loading := true
			calls := 0
			submit := func() {
				state.SetItems([]nui.MenuItem{{ID: "action", Label: "项目操作与很长的中文名称", Description: "描述需要在有限的宽度中省略", Loading: loading, Icon: lucide.Info(), Avatar: &nui.AvatarOptions{Alt: "Hidden avatar"}, Checked: nui.Set(true), Close: nui.Set(false), OnSelect: func() {
					calls++
				}, UI: nui.Styles{nui.Indicator: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}}})
			}
			submit()
			tt := ui.NewTester(func(c *ui.Context) {
				mode := nui.LightMode
				if dark {
					mode = nui.DarkMode
				}
				th := nui.App(c, nui.AppConfig{Mode: mode})
				th.Motion.Disabled = true
				a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Right(12).Top(12)
				nui.DropdownMenu(c, th, a, &state, nui.MenuOptions{Size: nui.XL, Width: nui.Set(float32(500)), CheckedIcon: lucide.Heart(), LoadingIcon: lucide.RefreshCw(), Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Panel {
						e.Label("Panel")
					}
					if slot == nui.Indicator {
						e.Label("Checked icon")
					}
					if slot == nui.Leading {
						e.Label("Leading icon")
					}
				}})
			}, 280, 280)
			tt.SetDirection(ui.RightToLeft)
			tt.SetPreferences(ui.Preferences{TextScale: 1.5})
			click(t, tt, "Actions")
			panel, row, icon := bounds(t, tt, "Panel"), bounds(t, tt, "项目操作与很长的中文名称"), bounds(t, tt, "Leading icon")
			if panel.X < 0 || panel.X+panel.W > 280 || panel.W > 256 || icon.W != 36 || row.H != 88 || tt.HasText("Hidden avatar") {
				t.Fatalf("narrow RTL scale: panel=%v row=%v icon=%v", panel, row, icon)
			}
			checked := bounds(t, tt, "Checked icon")
			if checked.X >= icon.X {
				t.Fatal("RTL did not put trailing indicator on the left")
			}
			pink := false
			for y := int(checked.Y); y < int(checked.Y+checked.H); y++ {
				for x := int(checked.X); x < int(checked.X+checked.W); x++ {
					p := tt.Image().RGBAAt(x, y)
					if p.R == 255 && p.G == 0 && p.B == 255 {
						pink = true
					}
				}
			}
			if !pink {
				t.Fatal("per-item indicator color did not render")
			}
			tt.Key(0, ui.KeyEnter)
			if calls != 1 || !state.Open {
				t.Fatal("loading item incorrectly blocked selection")
			}
			loading = false
			submit()
			tt.Frame()
			if !tt.Focused("项目操作与很长的中文名称") {
				t.Fatal("loading update lost action focus")
			}
		})
	}
}

func TestDropdownMenuDisablingOpenAncestorRepairsFocus(t *testing.T) {
	var state nui.MenuState
	items := []nui.MenuItem{{ID: "child", Label: "Export", Disabled: false, Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}}}, {ID: "other", Label: "Other"}}
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c, th, a, &state, nui.MenuOptions{})
	}, 500, 400)
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("CSV") {
		t.Fatal("child did not open")
	}
	items[0].Disabled = true
	state.SetItems(items)
	tt.Frame()
	if !state.Open || tt.HasText("CSV") || !tt.Focused("Other") {
		t.Fatal("disabling an open ancestor did not repair focus within the remaining menu")
	}
}

func TestContextMenuPointerRepositionAndControlledReopen(t *testing.T) {
	var state nui.MenuState
	var changes []bool
	actions := 0
	state.SetItems([]nui.MenuItem{{ID: "first", Label: "First"}, {ID: "second", Label: "Second", OnSelect: func() {
		actions++
	}}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		target := nui.Button(c.Key("target"), th, "Target", nui.ButtonOptions{}).Absolute().Left(40).Top(30).Width(300).Height(180)
		nui.ContextMenu(c.Key("menu"), th, target, &state, nui.MenuOptions{Label: "Context panel", Width: nui.Set(float32(180)), OnOpenChange: func(v bool) { changes = append(changes, v) }})
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
	state.SetItems([]nui.MenuItem{{ID: "a", Label: "Action"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		target := nui.Button(c, th, "Target", nui.ButtonOptions{Disabled: targetDisabled}).Width(200).Height(100).OnClick(func() { clicks++ })
		nui.ContextMenu(c, th, target, &state, nui.MenuOptions{Disabled: disabled, OnOpenChange: func(bool) { changes++ }})
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
				state.SetItems([]nui.MenuItem{{ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}, {ID: "blocked", Label: "Blocked", Disabled: true}, {ID: "json", Label: "JSON"}}}})
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{Mode: nui.DarkMode})
					th.Motion.Reduce = true
					target := nui.Button(c, th, "Target", nui.ButtonOptions{}).Absolute().Left(20).Top(20).Width(width - 40).Height(250)
					nui.ContextMenu(c, th, target, &state, nui.MenuOptions{Label: "Context panel", Width: nui.Set(float32(200))})
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
	state.SetItems([]nui.MenuItem{{ID: "one", Label: "Action"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c.Key("target"), th, "Target", nui.ButtonOptions{}).Absolute().Left(x).Top(30).Width(300).Height(180)
		nui.ContextMenu(c.Key("menu"), th, a, &state, nui.MenuOptions{Label: "Context panel"})
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
	state.SetItems([]nui.MenuItem{{ID: "a", Label: "Action"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c, th, "Target", nui.ButtonOptions{}).Absolute().Left(40).Top(30).Width(300).Height(100)
		nui.ContextMenu(c, th, a, &state, nui.MenuOptions{Label: "Context panel"})
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

func TestMenuVirtualRowsAndOffscreenNavigation(t *testing.T) {
	const count = 10000
	items := make([]nui.MenuItem, count)
	chosen, actions, built := "", 0, 0
	for i := range items {
		id := fmt.Sprintf("Action %06d", i)
		items[i] = nui.MenuItem{ID: id, Label: id, OnSelect: func() { chosen = id; actions++ }}
	}
	items[count-1].Disabled = true
	items[count-2].Label = "Zebra target"
	items[count-3].Kind = nui.MenuLabel
	var state nui.MenuState
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: true})
		anchor := nui.Button(c.Key("trigger"), th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c.Key("menu"), th, anchor, &state, nui.MenuOptions{Width: nui.Set(float32(300)), MaxHeight: nui.Set(float32(200)), Decoration: nui.Decoration{Style: func(slot nui.Slot, _ ui.Element) {
			if slot == nui.Item {
				built++
			}
		}}})
	}, 500, 350)
	bounded := func() {
		t.Helper()
		built = 0
		tt.Frame()
		if built < 1 || built > 16 {
			t.Fatalf("built %d rows for %d commands", built, count)
		}
	}
	click(t, tt, "Actions")
	bounded()
	tt.Key(0, ui.KeyEnd)
	if !tt.Focused("Zebra target") {
		t.Fatal("End did not reveal the last enabled action")
	}
	bounded()
	tt.Key(0, ui.KeyUp)
	if !tt.Focused(items[count-4].Label) {
		t.Fatal("Up did not skip the structural row")
	}
	tt.Key(0, ui.KeyHome)
	if !tt.Focused(items[0].Label) {
		t.Fatal("Home did not reveal the first action")
	}
	tt.Type("z")
	if !tt.Focused("Zebra target") {
		t.Fatal("typeahead did not reveal an offscreen match")
	}
	bounded()
	tt.Key(0, ui.KeyEnter)
	if chosen != items[count-2].ID || actions != 1 || state.Open || !tt.Focused("Actions") {
		t.Fatal("offscreen activation or trigger restoration failed")
	}
	built = 0
	tt.Frame()
	if built != 0 {
		t.Fatal("closed menu built command rows")
	}
}

func TestMenuVirtualNestedRebindAndFilter(t *testing.T) {
	items := make([]nui.MenuItem, 10000)
	children := make([]nui.MenuItem, 10000)
	for i := range items {
		items[i] = nui.MenuItem{ID: fmt.Sprintf("root-%d", i), Label: fmt.Sprintf("Root %05d", i)}
		children[i] = nui.MenuItem{ID: fmt.Sprintf("child-%d", i), Label: fmt.Sprintf("Child %05d", i)}
	}
	items[len(items)-1].Children = children
	var state nui.MenuState
	state.SetItems(items)
	query := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: true})
		anchor := nui.Button(c.Key("trigger"), th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c.Key("menu"), th, anchor, &state, nui.MenuOptions{Width: nui.Set(float32(200)), Filter: true, SearchTerm: &query, SearchExecution: nui.SearchImmediate, FilterInput: &nui.InputOptions{Label: "Find action"}})
	}, 700, 400)
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEnd)
	if !tt.Focused("Child 09999") {
		t.Fatal("nested End did not reach the last child")
	}
	tt.Key(0, ui.KeyLeft)
	if !tt.Focused("Root 09999") {
		t.Fatal("return did not restore the offscreen parent")
	}
	slices.Reverse(items)
	state.SetItems(items)
	tt.Frame()
	if !tt.Focused("Root 09999") {
		t.Fatal("large same-ID reorder lost focus")
	}
	tt.Key(0, ui.KeyRight)
	items[0].Disabled = true
	state.SetItems(items)
	tt.Frame()
	if tt.HasText("Child 00000") || !tt.Focused("Root 09998") {
		t.Fatal("disabled submenu anchor retained its child or lost repaired focus")
	}
	query = "Root 00042"
	tt.Frame()
	if !tt.Focused("Find action") || !tt.HasText("Root 00042") {
		t.Fatal("filter removal failed to return focus to the editor")
	}
	tt.Key(0, ui.KeyDown)
	if !tt.Focused("Root 00042") {
		t.Fatal("filtered logical navigation used source indexes")
	}
	state.SetItems(nil)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if state.Open || !tt.Focused("Actions") {
		t.Fatal("empty replacement lost dismissal")
	}
}

func TestMenuVirtualMeasuredCustomRows(t *testing.T) {
	var state nui.MenuState
	state.SetItems([]nui.MenuItem{{ID: "tall", Label: "Tall action", Build: func(c *ui.Context, _ nui.Theme) {
		ui.Text(c, "Custom content").Height(120)
	}}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: true})
		anchor := nui.Button(c.Key("trigger"), th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c.Key("menu"), th, anchor, &state, nui.MenuOptions{Width: nui.Set(float32(250))})
	}, 500, 400)
	click(t, tt, "Actions")
	tt.Frame()
	if got := bounds(t, tt, "Tall action").H; got != 132 {
		t.Fatalf("measured custom height = %g, want 132", got)
	}
	tt.Key(0, ui.KeyEscape)
	click(t, tt, "Actions")
	if !tt.Focused("Tall action") {
		t.Fatal("measured row lost focus after reopening")
	}
}

func TestMenuOwnedUpdatesAndEmptySubmenuDismissal(t *testing.T) {
	var state nui.MenuState
	before, after := 0, 0
	items := []nui.MenuItem{{ID: "group", Label: "Export", Children: []nui.MenuItem{{ID: "action", Label: "Original", OnSelect: func() { before++ }}}}}
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: true})
		anchor := nui.Button(c.Key("trigger"), th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c.Key("menu"), th, anchor, &state, nui.MenuOptions{})
	}, 500, 350)
	items[0].Children[0].Label = "Replacement"
	items[0].Children[0].OnSelect = func() { after++ }
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Original") || tt.HasText("Replacement") {
		t.Fatal("source edits bypassed owned submission")
	}
	tt.Key(0, ui.KeyEnter)
	if before != 1 || after != 0 {
		t.Fatal("unsubmitted callback replacement leaked")
	}
	state.SetItems(items)
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Replacement") {
		t.Fatal("same-ID label/callback resubmission did not reach the submenu")
	}
	tt.Key(0, ui.KeyEnter)
	if before != 1 || after != 1 {
		t.Fatal("replacement callback did not run exactly once")
	}
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyRight)
	state.SetItems(nil)
	tt.Frame()
	if tt.HasText("Replacement") || tt.HasText("Export") {
		t.Fatal("empty submission retained old rows")
	}
	tt.Key(0, ui.KeyEscape)
	if state.Open || !tt.Focused("Actions") {
		t.Fatal("emptying an open submenu lost keyboard dismissal or trigger focus")
	}
}
