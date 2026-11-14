package nui_test

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSelectMenuCreateExcludedFromBulkSelection(t *testing.T) {
	for _, top := range []bool{false, true} {
		for _, key := range []struct {
			name string
			mods ui.Modifiers
			key  ui.Key
		}{{"Ctrl+A", ui.Ctrl, ui.KeyA}, {"Cmd+A", ui.Super, ui.KeyA}, {"Shift+End", ui.Shift, ui.KeyEnd}} {
			t.Run(key.name+map[bool]string{false: "/bottom", true: "/top"}[top], func(t *testing.T) {
				var values, notified []string
				s := nui.SelectMenuState{Open: true, Query: "Al"}
				changes, selects, creates := 0, 0, 0
				// A business ID resembling the internal action ID must remain selectable.
				items := []nui.ListboxItem{{ID: "\x00nui:create", Label: "Alpha"}, {ID: "b", Label: "Alpine"}, {ID: "c", Label: "Albatross", Disabled: true}}
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					nui.SelectMenu(c, th, &values, &s, items, nui.SelectMenuOptions{
						Label: "Letters", Multiple: true, SelectionBehavior: nui.SelectionReplace, CreateAtTop: top, CreateWhenAlways: true,
						OnChange: func(ids []string) { changes++; notified = slices.Clone(ids) },
						OnSelect: func(string) { selects++ }, OnCreate: func(string) { creates++ },
					}).Width(260)
				}, 400, 400)
				tt.Advance(time.Second)
				tt.Key(0, ui.KeyDown)
				if !tt.Focused("Letters options") {
					t.Fatal("the list did not receive focus")
				}
				tt.Key(0, ui.KeyHome)
				tt.Key(key.mods, key.key)
				want := []string{"\x00nui:create", "b"}
				if !slices.Equal(values, want) || !slices.Equal(notified, want) || changes != 1 || selects != 0 || creates != 0 {
					t.Fatalf("bulk selection: values=%q notified=%q changes=%d selects=%d creates=%d", values, notified, changes, selects, creates)
				}
				// The action remains reachable and activates once without changing selection.
				end := ui.KeyEnd
				if top {
					end = ui.KeyHome
				}
				tt.Key(0, end)
				for _, activate := range []ui.Key{ui.KeyEnter, ui.KeySpace} {
					tt.Key(0, activate)
				}
				if creates != 2 || changes != 1 || selects != 0 || !slices.Equal(values, want) || !s.Open || s.Query != "Al" {
					t.Fatalf("create activation: values=%q changes=%d selects=%d creates=%d open=%v query=%q", values, changes, selects, creates, s.Open, s.Query)
				}
			})
		}
	}
}

func TestSelectMenuCreateKeepsBusinessBuilderSeparate(t *testing.T) {
	var values []string
	s := nui.SelectMenuState{Open: true, Query: "Al"}
	var unexpected []string
	created := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, &values, &s, []nui.ListboxItem{{ID: "a", Label: "Alpha"}}, nui.SelectMenuOptions{
			Label: "Letters", CreateWhenAlways: true, OnCreate: func(query string) { created = query },
			Build: func(c *ui.Context, th nui.Theme, it nui.ListboxItem) {
				if it.ID != "a" {
					unexpected = append(unexpected, it.ID)
					return
				}
				ui.Text(c, "Custom Alpha")
			},
		}).Width(260)
	}, 400, 400)
	tt.Advance(time.Second)
	if len(unexpected) != 0 || !tt.HasText("Custom Alpha") {
		t.Fatalf("builder received %q; texts=%v", unexpected, tt.Texts())
	}
	click(t, tt, "Create “Al”")
	if created != "Al" || len(values) != 0 || !s.Open || s.Query != "Al" {
		t.Fatalf("create: query=%q created=%q values=%q open=%v", s.Query, created, values, s.Open)
	}
	// Rebuild with only the action row, then restore the business row.
	s.Query = "New"
	tt.Frame()
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if created != "New" || len(unexpected) != 0 || len(values) != 0 {
		t.Fatalf("unmatched create: created=%q builder IDs=%q values=%q", created, unexpected, values)
	}
	s.Query = "Al"
	tt.Frame()
	click(t, tt, "Custom Alpha")
	if !slices.Equal(values, []string{"a"}) || s.Open {
		t.Fatalf("business row no longer selects: values=%q open=%v", values, s.Open)
	}
}

func TestSelectMenuClearAndSearchSubmit(t *testing.T) {
	state := nui.SelectMenuState{}
	values := []string{"b"}
	reset, search, focusSearch, focusElsewhere := true, true, true, false
	changes, clears := 0, 0
	var outside ui.Handle
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Button(c.Key("outside"), th, "Outside target", nui.ButtonOptions{}).Bind(&outside)
			nui.SelectMenu(c.Key("menu"), th, &values, &state, []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, nui.SelectMenuOptions{
				Label: "Menu", Clearable: true, ResetModelValueOnClear: nui.Set(reset), SearchInput: nui.Set(search), FocusSearch: nui.Set(focusSearch),
				ResetSearchTermOnBlur: nui.Set(false), FilterInput: &nui.InputOptions{Label: "Filter choices"},
				OnChange: func([]string) { changes++ }, OnClear: func() { clears++ },
				OnSelect: func(string) {
					if focusElsewhere {
						outside.Focus(c)
					}
				},
			}).Width(300)
		})
	}, 450, 400)
	click(t, tt, "Menu")
	tt.Type("alp")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(values, []string{"a"}) || changes != 1 || state.Open || state.Query != "" {
		t.Fatalf("search Enter: values=%v changes=%d open=%v query=%q", values, changes, state.Open, state.Query)
	}
	reset = false
	state.Query = "draft"
	tt.Frame()
	click(t, tt, "Clear selection")
	if !slices.Equal(values, []string{"a"}) || state.Query != "" || changes != 1 || clears != 1 || !tt.Focused("Menu") {
		t.Fatal("clear without model reset changed the value, kept the query, or lost focus")
	}
	reset = true
	tt.Frame()
	click(t, tt, "Clear selection")
	if len(values) != 0 || changes != 2 || clears != 2 || !tt.Focused("Menu") {
		t.Fatal("clear did not reset once and return trigger focus")
	}
	search, focusSearch = false, false
	state.Open, state.Query = true, "bet"
	tt.Frame()
	if tt.HasText("Alpha") || !tt.HasText("Beta") || !tt.Focused("Menu options") {
		t.Fatal("hidden search ignored its controlled query or initial list focus")
	}
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(values, []string{"b"}) || changes != 3 || state.Open {
		t.Fatal("searchless selection did not commit")
	}
	search, focusElsewhere = true, true
	state.Open = true
	tt.Frame()
	if !tt.Focused("Menu options") {
		t.Fatal("FocusSearch=false did not focus the list")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Advance(time.Second)
	if state.Open || !tt.Focused("Outside target") || changes != 3 {
		t.Fatal("closing a reselected value discarded application focus or emitted a value change")
	}
}

func TestInputMenuTrailingStylesAndToggle(t *testing.T) {
	var state nui.SearchState
	var instance, callback nui.Optional[ui.Color]
	readOnly, disabled := false, false
	valid := false
	red, green, blue := ui.RGB(240, 0, 0), ui.RGB(0, 180, 0), ui.RGB(0, 0, 240)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode}).WithComponent(nui.ComponentInputMenu, nui.ComponentTheme{UI: nui.Styles{nui.Trailing: {Foreground: nui.Set(red)}}})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputMenu(c, th, &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.InputMenuOptions{
				Label: "Letters", Trailing: lucide.Search(), Clearable: true,
				ReadOnly: readOnly, Disabled: disabled,
				UI: nui.Styles{nui.Trailing: {Foreground: instance}},
				Style: func(root, editor, leading, trailing ui.Element) {
					valid = trailing.Valid()
					trailing.Label("Styled trailing")
					if callback.Set {
						trailing.TextColor(callback.Value)
					}
				},
			}).Width(260)
		})
	}, 400, 300)
	check := func(want ui.Color) {
		t.Helper()
		if !valid {
			t.Fatal("Style received an empty trailing element")
		}
		if countColor(tt.Image(), bounds(t, tt, "Styled trailing"), rgba(want)) == 0 {
			t.Fatalf("trailing icon does not use %v", want)
		}
	}
	check(red)
	instance = nui.Set(green)
	tt.Frame()
	check(green)
	callback = nui.Set(blue)
	tt.Frame()
	check(blue)
	click(t, tt, "Styled trailing")
	if !tt.HasText("Alpha") || !tt.Focused("Letters") {
		t.Fatal("styled trailing icon did not open the list and focus the editor")
	}
	click(t, tt, "Styled trailing")
	if tt.HasText("Alpha") {
		t.Fatal("styled trailing icon did not close the list")
	}
	tt.Type("Al")
	if _, shown := tt.Find("Styled trailing"); shown {
		t.Fatal("the trailing icon should give way to the clear button")
	}
	click(t, tt, "Clear Letters")
	if state.Query != "" {
		t.Fatalf("clear left query %q", state.Query)
	}
	check(blue)
	click(t, tt, "Styled trailing")
	if !tt.HasText("Alpha") || !tt.Focused("Letters") {
		t.Fatal("the trailing icon did not restore its behavior after clearing")
	}
	for _, mode := range []string{"readonly", "disabled"} {
		readOnly, disabled = mode == "readonly", mode == "disabled"
		tt.Frame()
		click(t, tt, "Styled trailing")
		if tt.HasText("Alpha") {
			t.Fatalf("%s trailing icon opened the list", mode)
		}
	}
}

func TestOverlayPanelOpacityOverrides(t *testing.T) {
	for _, kind := range []string{"popover", "menu", "tooltip", "modal"} {
		for _, source := range []string{"theme", "instance", "callback", "zero"} {
			t.Run(kind+"/"+source, func(t *testing.T) {
				open := true
				menu := nui.MenuState{Open: true}
				want := float32(.25)
				if source == "zero" {
					want = 0
				}
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
					component := map[string]nui.Component{"popover": nui.ComponentPopover, "menu": nui.ComponentDropdownMenu, "tooltip": nui.ComponentTooltip, "modal": nui.ComponentModal}[kind]
					th = th.WithComponent(component, nui.ComponentTheme{UI: nui.Styles{nui.Panel: {Background: nui.Set(ui.RGB(255, 0, 0)), Opacity: nui.Set(float32(.25)), Radius: nui.Set(float32(0)), BorderWidth: nui.Set(float32(0)), Shadow: nui.Set(nui.ShadowTokens{})}}})
					d := nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Panel {
							e.Label("Panel")
							if source == "callback" {
								e.Opacity(want)
							}
						}
					}}
					if source != "theme" {
						th = th.WithComponent(component, nui.ComponentTheme{UI: nui.Styles{nui.Panel: {Opacity: nui.Set(float32(.75))}}})
						d.UI = nui.Styles{nui.Panel: {Opacity: nui.Set(want)}}
						if source == "callback" {
							d.UI[nui.Panel] = nui.PartStyle{Opacity: nui.Set(float32(.5))}
						}
					}
					a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(50).Top(20)
					switch kind {
					case "popover":
						nui.Popover(c, th, a, &open, nui.PopoverOptions{Decoration: d, Width: nui.Set(float32(200))}, func(nui.Theme) { ui.Text(c, "Contents") })
					case "menu":
						nui.DropdownMenu(c, th, a, &menu, []nui.MenuItem{{ID: "a", Label: "Action"}}, nui.MenuOptions{Decoration: d})
					case "tooltip":
						nui.Tooltip(c, th, a, "Contents", nui.TooltipOptions{Open: &open, Decoration: d})
					case "modal":
						nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Panel", Overlay: nui.Set(false), Decoration: d}, func(nui.Theme) { ui.Text(c, "Contents") })
					}
				}, 600, 400)
				pixel := func() uint8 {
					r := bounds(t, tt, "Panel")
					return tt.Image().RGBAAt(int(r.X+2), int(r.Y+r.H/2)).G
				}
				first := pixel()
				tt.Advance(50 * time.Millisecond)
				middle := pixel()
				tt.Advance(time.Second)
				last := pixel()
				expected := int(255 * (1 - want))
				if difference := int(last) - expected; difference < -2 || difference > 2 {
					t.Fatalf("final opacity %v: green=%d, want about %d", want, last, expected)
				}
				if source != "callback" && want > 0 && (first <= middle || middle <= last) {
					t.Fatalf("opacity override should still fade in: green=%d, %d, %d", first, middle, last)
				}
				open, menu.Open = false, false
				tt.Frame()
				if _, visible := tt.Find("Panel"); visible {
					t.Fatal("closed panel is still visible")
				}
				tt.SetPreferences(ui.Preferences{ReduceMotion: true})
				open, menu.Open = true, true
				tt.Frame()
				if got := pixel(); got != last {
					t.Fatalf("reduced motion changed explicit opacity: %d -> %d", last, got)
				}
			})
		}
	}
}
