package nui_test

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// popupDocs names the docs that describe MyGo's AttachTo room rule, which
// these tests pin: a failure after a MyGo upgrade means they need updating.
const popupDocs = "if MyGo's AttachTo room rule changed, update docs/inputs.md (Popover, Select, InputMenu) and docs/overlays-navigation.md (DropdownMenu, ContextMenu)"

const shortW, shortH = 400, 240

func placementItems(n int) []nui.MenuItem {
	items := make([]nui.MenuItem, n)
	for i := range items {
		items[i] = nui.MenuItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Item %d", i)}
	}
	return items
}

func placementOptions(n int) []nui.Option[string] {
	options := make([]nui.Option[string], n)
	for i := range options {
		options[i] = nui.Option[string]{Value: fmt.Sprint(i), Label: fmt.Sprintf("Option %d", i)}
	}
	return options
}

func placementListboxItems(n int) []nui.ListboxItem {
	items := make([]nui.ListboxItem, n)
	for i := range items {
		items[i] = nui.ListboxItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Option %d", i)}
	}
	return items
}

func placementTheme(c *ui.Context) nui.Theme {
	th := nui.App(c, nui.AppConfig{})
	th.Motion.Reduce = true
	return th
}

// insideWindow fails unless r lies within a w×h window.
func insideWindow(t *testing.T, what string, r ui.Rect, w, h float32) {
	t.Helper()
	if r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 || r.X+r.W > w || r.Y+r.H > h {
		t.Fatalf("%s %+v is not inside the %vx%v window; %s", what, r, w, h, popupDocs)
	}
}

// fullyVisible fails unless the entry named label shows its full height
// inside panel and the window. Find returns the box clipped to what shows.
func fullyVisible(t *testing.T, tt *ui.Tester, label string, rowH float32, panel ui.Rect, w, h float32) {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok || r.H < rowH || r.Y < panel.Y || r.Y+r.H > panel.Y+panel.H || r.Y < 0 || r.Y+r.H > h || r.X+r.W > w {
		t.Fatalf("%s is not fully visible: %+v (row height %v, panel %+v); %s", label, r, rowH, panel, popupDocs)
	}
}

// scrollToEnd scrolls the list under the entry named first to its end.
func scrollToEnd(t *testing.T, tt *ui.Tester, first string) {
	t.Helper()
	r := bounds(t, tt, first)
	for range 6 {
		tt.Scroll(r.X+r.W/2, r.Y+r.H/2, 0, 2000)
		tt.Frame()
	}
}

func TestPopupPlacementDropdownMenuFitsBelow(t *testing.T) {
	var state nui.MenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := placementTheme(c)
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(20)
		nui.DropdownMenu(c, th, a, &state, placementItems(20), nui.MenuOptions{Label: "Menu panel"})
	}, shortW, shortH)
	click(t, tt, "Actions")
	panel := bounds(t, tt, "Menu panel")
	insideWindow(t, "menu panel", panel, shortW, shortH)
	trigger := bounds(t, tt, "Actions")
	if panel.Y < trigger.Y+trigger.H {
		t.Fatalf("panel %+v covers its trigger %+v; %s", panel, trigger, popupDocs)
	}
	rowH := bounds(t, tt, "Item 0").H
	scrollToEnd(t, tt, "Item 0")
	fullyVisible(t, tt, "Item 19", rowH, bounds(t, tt, "Menu panel"), shortW, shortH)
}

func TestPopupPlacementDropdownMenuOpensAboveNearBottom(t *testing.T) {
	var state nui.MenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := placementTheme(c)
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(170)
		nui.DropdownMenu(c, th, a, &state, placementItems(20), nui.MenuOptions{Label: "Menu panel"})
	}, shortW, shortH)
	click(t, tt, "Actions")
	panel, trigger := bounds(t, tt, "Menu panel"), bounds(t, tt, "Actions")
	insideWindow(t, "menu panel", panel, shortW, shortH)
	if panel.Y+panel.H > trigger.Y {
		t.Fatalf("panel %+v does not end above its trigger %+v; %s", panel, trigger, popupDocs)
	}
	rowH := float32(32) // MD: a 20 DIP line and 6 DIP padding on each side
	fullyVisible(t, tt, "Item 0", rowH, panel, shortW, shortH)
	scrollToEnd(t, tt, "Item 0")
	fullyVisible(t, tt, "Item 19", rowH, bounds(t, tt, "Menu panel"), shortW, shortH)
}

func TestPopupPlacementFilteredDropdownMenu(t *testing.T) {
	var state nui.MenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := placementTheme(c)
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(90)
		nui.DropdownMenu(c, th, a, &state, placementItems(20), nui.MenuOptions{Label: "Menu panel", Filter: true, FilterInput: &nui.InputOptions{Label: "Filter items"}})
	}, shortW, shortH)
	click(t, tt, "Actions")
	panel := bounds(t, tt, "Menu panel")
	insideWindow(t, "filtered menu panel", panel, shortW, shortH)
	rowH := bounds(t, tt, "Item 0").H
	scrollToEnd(t, tt, "Item 0")
	panel = bounds(t, tt, "Menu panel")
	if f, ok := tt.Find("Filter items"); !ok || f.H <= 0 || f.Y < panel.Y || f.Y+f.H > panel.Y+panel.H {
		t.Fatalf("filter field %+v left the panel %+v; %s", f, panel, popupDocs)
	}
	fullyVisible(t, tt, "Item 19", rowH, panel, shortW, shortH)
}

func TestPopupPlacementContextMenu(t *testing.T) {
	var state nui.MenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := placementTheme(c)
		area := ui.Box(c).Absolute().Left(20).Top(20).Size(200, 200).Label("Area")
		nui.ContextMenu(c, th, area, &state, placementItems(20), nui.MenuOptions{Label: "Menu panel"})
	}, shortW, shortH)
	tt.RightClickAt(60, 150)
	tt.Frame()
	panel := bounds(t, tt, "Menu panel")
	insideWindow(t, "context menu panel", panel, shortW, shortH)
	rowH := bounds(t, tt, "Item 0").H
	scrollToEnd(t, tt, "Item 0")
	fullyVisible(t, tt, "Item 19", rowH, bounds(t, tt, "Menu panel"), shortW, shortH)
}

func TestPopupPlacementSelectionLists(t *testing.T) {
	popup := nui.PopoverOptions{Label: "List panel"}
	// Each case returns the view of one component, over state of its own.
	cases := []struct {
		name string
		view func() func(c *ui.Context, th nui.Theme)
		open string
	}{
		{"Select", func() func(*ui.Context, nui.Theme) {
			value := ""
			return func(c *ui.Context, th nui.Theme) {
				nui.Select(c, th, &value, placementOptions(20), nui.SelectOptions{Placeholder: "Pick one", Popup: popup})
			}
		}, "Pick one"},
		{"SelectMenu", func() func(*ui.Context, nui.Theme) {
			var value []string
			var state nui.SelectMenuState
			return func(c *ui.Context, th nui.Theme) {
				nui.SelectMenu(c, th, &value, &state, placementListboxItems(20), nui.SelectMenuOptions{Placeholder: "Pick some", Popup: popup})
			}
		}, "Pick some"},
		{"InputMenu", func() func(*ui.Context, nui.Theme) {
			var state nui.SearchState
			return func(c *ui.Context, th nui.Theme) {
				nui.InputMenu(c, th, &state, placementOptions(20), nui.InputMenuOptions{Label: "Search", OpenOnClick: nui.Set(true), Popup: popup})
			}
		}, "Search"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			view := tc.view()
			tt := ui.NewTester(func(c *ui.Context) {
				th := placementTheme(c)
				ui.Box(c).Absolute().Left(20).Top(90).Width(200).Children(func() { view(c, th) })
			}, shortW, shortH)
			click(t, tt, tc.open)
			panel := bounds(t, tt, "List panel")
			insideWindow(t, tc.name+" panel", panel, shortW, shortH)
			if trigger := bounds(t, tt, tc.open); panel.Y < trigger.Y+trigger.H && panel.Y+panel.H > trigger.Y {
				t.Fatalf("panel %+v covers its trigger %+v; %s", panel, trigger, popupDocs)
			}
			rowH := bounds(t, tt, "Option 0").H
			scrollToEnd(t, tt, "Option 0")
			fullyVisible(t, tt, "Option 19", rowH, bounds(t, tt, "List panel"), shortW, shortH)
		})
	}
}

func TestPopupPlacementInputMenuEmptyMessage(t *testing.T) {
	var state nui.SearchState
	tt := ui.NewTester(func(c *ui.Context) {
		th := placementTheme(c)
		ui.Box(c).Absolute().Left(20).Top(90).Width(200).Children(func() {
			nui.InputMenu(c, th, &state, placementOptions(20), nui.InputMenuOptions{Label: "Search", OpenOnClick: nui.Set(true), Popup: nui.PopoverOptions{Label: "List panel"}})
		})
	}, shortW, shortH)
	click(t, tt, "Search")
	tt.Type("missing")
	tt.Frame()
	panel, empty := bounds(t, tt, "List panel"), bounds(t, tt, "No matching data")
	// The message sits below the empty list, inside the panel's 4 DIP padding.
	if empty.H <= 0 || empty.Y != panel.Y+4 || empty.Y+empty.H > panel.Y+panel.H || empty.X < panel.X || empty.X+empty.W > panel.X+panel.W {
		t.Fatalf("empty message %+v moved within panel %+v", empty, panel)
	}
}

func TestPopupPlacementPopover(t *testing.T) {
	for _, top := range []float32{20, 170} {
		t.Run(fmt.Sprint(top), func(t *testing.T) {
			open := false
			tt := ui.NewTester(func(c *ui.Context) {
				th := placementTheme(c)
				a := nui.Button(c, th, "Open", nui.ButtonOptions{}).Absolute().Left(20).Top(top).OnClick(func() { open = !open })
				nui.Popover(c, th, a, &open, nui.PopoverOptions{Label: "Popover panel"}, func(nui.Theme) {
					ui.Column(c).Children(func() {
						for i := range 12 {
							ui.Text(c.Key(i), fmt.Sprintf("Line %d", i)).Height(24)
						}
					})
				})
			}, shortW, shortH)
			click(t, tt, "Open")
			panel, trigger := bounds(t, tt, "Popover panel"), bounds(t, tt, "Open")
			insideWindow(t, "popover panel", panel, shortW, shortH)
			if top == 170 {
				if panel.Y+panel.H > trigger.Y {
					t.Fatalf("panel %+v does not end above its trigger %+v; %s", panel, trigger, popupDocs)
				}
				fullyVisible(t, tt, "Line 0", 24, panel, shortW, shortH)
			} else if panel.Y < trigger.Y+trigger.H {
				t.Fatalf("panel %+v covers its trigger %+v; %s", panel, trigger, popupDocs)
			}
		})
	}
}

func TestPopupPlacementFittingPopupKeepsItsSize(t *testing.T) {
	t.Run("short menu", func(t *testing.T) {
		var state nui.MenuState
		tt := ui.NewTester(func(c *ui.Context) {
			th := placementTheme(c)
			a := nui.Button(c, th, "Actions", nui.ButtonOptions{}).Absolute().Left(20).Top(20)
			nui.DropdownMenu(c, th, a, &state, placementItems(3), nui.MenuOptions{Label: "Menu panel"})
		}, 500, 400)
		click(t, tt, "Actions")
		panel, first, last := bounds(t, tt, "Menu panel"), bounds(t, tt, "Item 0"), bounds(t, tt, "Item 2")
		top, bottom := first.Y-panel.Y, panel.Y+panel.H-(last.Y+last.H)
		if top <= 0 || top != bottom || last.H != first.H || last.Y != first.Y+2*first.H {
			t.Fatalf("panel %+v does not hug its rows %+v..%+v; %s", panel, first, last, popupDocs)
		}
	})
	t.Run("MaxHeight in a tall window", func(t *testing.T) {
		value := ""
		tt := ui.NewTester(func(c *ui.Context) {
			th := placementTheme(c)
			ui.Box(c).Absolute().Left(20).Top(20).Width(200).Children(func() {
				nui.Select(c, th, &value, placementOptions(40), nui.SelectOptions{Placeholder: "Pick one", Popup: nui.PopoverOptions{Label: "List panel"}})
			})
		}, 500, 800)
		click(t, tt, "Pick one")
		// docs/inputs.md: selection popups default to a 240 DIP maximum height.
		if panel := bounds(t, tt, "List panel"); panel.H > 240 || panel.H < 200 {
			t.Fatalf("panel %+v ignores the default 240 DIP MaxHeight", panel)
		}
	})
}
