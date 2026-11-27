package nui_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

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
