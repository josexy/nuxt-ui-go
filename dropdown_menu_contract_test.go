package nui_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestDropdownMenuSizesAndDescriptionAlignment(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			var state nui.MenuState
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				th.Motion.Reduce = true
				a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(20)
				leadingIndex := 0
				nui.DropdownMenu(c, th, a, &state, []nui.MenuItem{
					{ID: "heading", Kind: nui.MenuLabel, Label: "Heading", Description: "Group description", Avatar: &nui.AvatarOptions{Alt: "Group avatar"}},
					{ID: "rich", Label: "Rich action", Description: "Second line", Icon: lucide.Info(), Shortcut: "Ctrl+K"},
					{ID: "plain", Label: "Plain action"},
				}, nui.MenuOptions{Size: size, Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Leading {
						e.Label(fmt.Sprintf("Leading %d", leadingIndex))
						leadingIndex++
					}
					if slot == nui.Trailing {
						e.Label("Keys")
					}
				}}})
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			a := nui.Button(c.Key("anchor"), th, "Actions", nui.ButtonOptions{}).OnClick(func() { triggerClicks++ })
			items := []nui.MenuItem{{ID: "first", Label: "First"}, {ID: "disabled", Label: "Disabled item", Disabled: true}, {ID: "check", Label: "Pinned", Checked: nui.Set(checked), Close: nui.Set(false), OnCheckedChange: func(v bool) { checked = v; updated++ }, OnSelect: func() { selected++ }}, {ID: "last", Label: "Last"}}
			if swap {
				items[0], items[2] = items[2], items[0]
			}
			nui.DropdownMenu(c.Key("menu"), th, a, &state, items, nui.MenuOptions{Loop: loop, Disabled: disabled, OnOpenChange: func(v bool) { opens = append(opens, v) }})
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			a := nui.Button(c.Key("anchor"), th, "Actions", nui.ButtonOptions{})
			nui.DropdownMenu(c.Key("menu"), th, a, &state, []nui.MenuItem{{ID: "label", Label: "Workspace", Kind: nui.MenuLabel}, {ID: "alpha", Label: "Alpha", Description: "first project"}, {ID: "blocked", Label: "Blocked", Description: "copy", Disabled: true}, {ID: "copy", Label: "Duplicate", Description: "copy project", OnSelect: func() { selections++ }}, {ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}, {ID: "json", Label: "JSON"}}, Content: &nui.MenuContentOptions{Filter: true, FilterInput: &nui.InputOptions{Label: "Export search"}}}}, nui.MenuOptions{Filter: filter, SearchTerm: &query, FilterFields: []string{"label", "description"}, IgnoreFilter: ignored, FilterInput: &nui.InputOptions{Label: "Action search", Clearable: true}, OnSearchTermChange: func(string) { changes++ }, OnOpenChange: func(v bool) { opens = append(opens, v) }})
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
	empty := true
	outside, changes := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			a := nui.Button(c.Key("trigger"), th, "Actions", nui.ButtonOptions{})
			var items []nui.MenuItem
			if !empty {
				items = []nui.MenuItem{{ID: "first", Label: "First"}}
			}
			nui.DropdownMenu(c.Key("menu"), th, a, &state, items, nui.MenuOptions{OnOpenChange: func(bool) { changes++ }})
			nui.Button(c.Key("outside"), th, "Outside", nui.ButtonOptions{}).Margin(180, 0, 0, 0).OnClick(func() { outside++ })
		})
	}, 500, 400)
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyEscape)
	if state.Open || !tt.Focused("Actions") || changes != 2 {
		t.Fatalf("empty dismissal %v/%d", state.Open, changes)
	}
	empty = false
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(20)
		nui.DropdownMenu(c, th, a, &state, []nui.MenuItem{{ID: "first", Label: "Alpha"}, {ID: "child", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}, {ID: "json", Label: "JSON"}}, Content: &nui.MenuContentOptions{Filter: true, FilterInput: &nui.InputOptions{Label: "Child search"}}}}, nui.MenuOptions{Filter: true, FilterInput: &nui.InputOptions{Label: "Root search"}})
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
			tt := ui.NewTester(func(c *ui.Context) {
				mode := nui.LightMode
				if dark {
					mode = nui.DarkMode
				}
				th := nui.App(c, nui.AppConfig{Mode: mode})
				th.Motion.Reduce = true
				a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Right(12).Top(12)
				nui.DropdownMenu(c, th, a, &state, []nui.MenuItem{{ID: "action", Label: "项目操作与很长的中文名称", Description: "描述需要在有限的宽度中省略", Loading: loading, Icon: lucide.Info(), Avatar: &nui.AvatarOptions{Alt: "Hidden avatar"}, Checked: nui.Set(true), Close: nui.Set(false), OnSelect: func() { calls++ }, UI: nui.Styles{nui.Indicator: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}}}, nui.MenuOptions{Size: nui.XL, Width: nui.Set(float32(500)), CheckedIcon: lucide.Heart(), LoadingIcon: lucide.RefreshCw(), Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Panel {
						e.Label("Panel")
					}
					if slot == nui.Indicator {
						e.Label("Checked icon")
					}
					if slot == nui.Leading {
						e.Label("Leading icon")
					}
				}}})
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
			tt.Frame()
			if !tt.Focused("项目操作与很长的中文名称") {
				t.Fatal("loading update lost action focus")
			}
		})
	}
}

func TestDropdownMenuDisablingOpenAncestorRepairsFocus(t *testing.T) {
	var state nui.MenuState
	disabled := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c, th, a, &state, []nui.MenuItem{{ID: "child", Label: "Export", Disabled: disabled, Children: []nui.MenuItem{{ID: "csv", Label: "CSV"}}}, {ID: "other", Label: "Other"}}, nui.MenuOptions{})
	}, 500, 400)
	click(t, tt, "Actions")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("CSV") {
		t.Fatal("child did not open")
	}
	disabled = true
	tt.Frame()
	if !state.Open || tt.HasText("CSV") || !tt.Focused("Other") {
		t.Fatal("disabling an open ancestor did not repair focus within the remaining menu")
	}
}
