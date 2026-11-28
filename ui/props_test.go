package nui_test

import (
	"context"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSelectMenuResetAndOpeningKeys(t *testing.T) {
	for _, keep := range []bool{false, true} {
		var state nui.SelectMenuState
		var selected []string
		var queries []string
		var openings []bool
		disabled, search := false, true
		listSelection := chosenSelection(selected)
		state.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}})
		tt := ui.NewTester(func(c *ui.Context) {
			theme := nui.App(c, nui.AppConfig{})
			ui.Column(c).Padding(20).Children(func() {
				nui.SelectMenu(c, theme, listSelection, &state, nui.SelectMenuOptions{
					Label: "Choose", Disabled: disabled, SearchInput: nui.Set(search),
					ResetSearchTermOnBlur: nui.Set(!keep), ResetSearchTermOnSelect: nui.Set(!keep), OnQuery: func(q string) { queries = append(queries, q) },
					OnOpenChange: func(open bool) { openings = append(openings, open) },
				}).Width(300)
				nui.Button(c, theme, "Outside", nui.ButtonOptions{}).Absolute().Top(440).Left(350)
			})
		}, 500, 500)
		click(t, tt, "Choose")
		if len(openings) != 1 || !openings[0] {
			t.Fatalf("opening notifications: %v", openings)
		}
		tt.Type("Alpha")
		click(t, tt, "Alpha")
		tt.Advance(500 * time.Millisecond)
		if state.Open || listSelection.Len() != 1 || listSelection.Values()[0] != "a" {
			t.Fatal("selection did not close")
		}
		if len(openings) != 2 || openings[1] {
			t.Fatalf("selection close notifications: %v", openings)
		}
		if (state.Query != "") != keep {
			t.Fatalf("keep=%v query=%q", keep, state.Query)
		}
		click(t, tt, "Choose")
		if tt.HasText("Beta") == keep {
			t.Fatal("reopen filtering does not reflect reset option")
		}
		tt.Key(ui.Ctrl, ui.KeyA)
		tt.Type("Beta")
		beforeClose := len(openings)
		tt.Key(0, ui.KeyEscape)
		tt.Advance(500 * time.Millisecond)
		if len(openings) != beforeClose+1 || openings[len(openings)-1] {
			t.Fatalf("Escape notifications: %v", openings)
		}
		if !tt.Focused("Choose") {
			t.Fatal("focus did not return")
		}
		if (state.Query != "") != keep {
			t.Fatal("Escape did not apply reset option")
		}
		for _, key := range []ui.Key{ui.KeyUp, ui.KeyDown} {
			tt.Key(0, key)
			if !state.Open {
				t.Fatalf("%v did not open", key)
			}
			tt.Key(0, ui.KeyEscape)
			tt.Advance(500 * time.Millisecond)
		}
		search = false
		tt.Frame()
		tt.Key(0, ui.KeyUp)
		if tt.HasText("Search Choose") || !tt.HasText("Beta") {
			t.Fatal("hidden search still filters/appears")
		}
		tt.Key(0, ui.KeyEnd)
		tt.Key(0, ui.KeyEnter)
		if listSelection.Len() != 1 || listSelection.Values()[0] != "b" {
			t.Fatalf("searchless keyboard selection: %v", listSelection.Values())
		}
		tt.Advance(500 * time.Millisecond)
		click(t, tt, "Choose")
		search = true
		tt.Frame()
		click(t, tt, "Search Choose")
		tt.Type("Alpha")
		click(t, tt, "Outside")
		tt.Advance(500 * time.Millisecond)
		if state.Open || ((state.Query != "") != keep) || !tt.Focused("Outside") {
			t.Fatalf("outside dismissal: open=%v query=%q keep=%v focused=%v", state.Open, state.Query, keep, tt.Focused("Outside"))
		}
		disabled = true
		tt.Frame()
		tt.Key(0, ui.KeyUp)
		if state.Open {
			t.Fatal("disabled trigger opened")
		}
		before := len(queries)
		tt.Frame()
		tt.Frame()
		if len(queries) != before {
			t.Fatal("reset callback repeated on rebuild")
		}
	}
}

func TestPaginationExplicitZeroAndControls(t *testing.T) {
	page := 5
	edges, controls := true, true
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.Pagination(c, theme, &page, nui.PaginationOptions{Total: 100, PageSize: 10, SiblingCount: nui.Set(0), ShowEdges: nui.Set(edges), ShowControls: nui.Set(controls)})
	}, 1000, 150)
	if tt.HasText("4") || tt.HasText("6") || !tt.HasText("5") || !tt.HasText("10") {
		t.Fatal("zero neighbors / edges")
	}
	edges, controls = false, false
	tt.Frame()
	if tt.HasText("1") || tt.HasText("10") || tt.HasText("Next") || !tt.HasText("5") {
		t.Fatal("hidden controls/edges")
	}
	page = 99
	tt.Frame()
	if page != 10 || !tt.HasText("10") {
		t.Fatal("page clamping")
	}
}

func TestPaginationEdgeWindowsAndActiveVariant(t *testing.T) {
	page := 1
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Pagination(c, th, &page, nui.PaginationOptions{Total: 100, PageSize: 10, SiblingCount: nui.Set(0), ShowEdges: nui.Set(true), ActiveColor: nui.Error, ActiveVariant: nui.Outline})
	}, 1000, 150)
	if !tt.HasText("3") || tt.HasText("4") {
		t.Fatal("first edge window")
	}
	page = 10
	tt.Frame()
	if !tt.HasText("8") || tt.HasText("7") {
		t.Fatal("last edge window")
	}
	r := bounds(t, tt, "Page 10")
	red, green, blue, _ := tt.Image().At(int(r.X+6), int(r.Y+6)).RGBA()
	if red < 60000 || green < 60000 || blue < 60000 {
		t.Fatalf("active outline overridden by Selected fill: %d,%d,%d", red, green, blue)
	}
}

func TestPropsLoadingAndSelectionLimit(t *testing.T) {
	loading, checked := true, false
	selected := []string{"a"}
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Gap(12).Children(func() {
			nui.Switch(c, th, &checked, nui.SwitchOptions{Label: "Sync", Loading: loading, OnChange: func(bool) { changes++ }})
			nui.CheckboxGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.CheckboxGroupOptions{Max: nui.Set(1), Variant: nui.ChoiceCard, Indicator: nui.IndicatorEnd, Highlight: true})
		})
	}, 400, 300)
	click(t, tt, "Sync")
	if checked || changes != 0 {
		t.Fatal("loading switch activated")
	}
	click(t, tt, "Beta")
	if !slices.Equal(selected, []string{"a"}) {
		t.Fatal("selection cap ignored")
	}
	click(t, tt, "Alpha")
	click(t, tt, "Beta")
	if !slices.Equal(selected, []string{"b"}) {
		t.Fatal("selected choice could not be removed at cap")
	}
	loading = false
	tt.Frame()
	click(t, tt, "Sync")
	if !checked || changes != 1 {
		t.Fatal("loaded switch did not activate once")
	}
}

func TestPropsTagsCommitPolicies(t *testing.T) {
	var tags []string
	var state nui.TagsState
	readOnly, onPaste := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputTags(c, th, &tags, &state, nui.InputTagsOptions{Label: "Tags", ReadOnly: readOnly, AddOnTab: true, AddOnBlur: true, AddOnPaste: nui.Set(onPaste), Duplicate: true, MaxLength: 3})
			nui.Button(c, th, "Next", nui.ButtonOptions{})
		})
	}, 500, 250)
	click(t, tt, "Tags")
	tt.Type("中文")
	tt.Key(0, ui.KeyTab)
	if !slices.Equal(tags, []string{"中文"}) || !tt.Focused("Tags") {
		t.Fatal("Tab must commit and retain input focus")
	}
	click(t, tt, "Tags")
	tt.Type("中文")
	click(t, tt, "Next")
	if len(tags) != 2 {
		t.Fatal("blur/duplicate policy")
	}
	click(t, tt, "Tags")
	tt.Type("long")
	tt.Key(0, ui.KeyEnter)
	if len(tags) != 3 || tags[2] != "lon" || state.Query != "" {
		t.Fatal("maxLength must cap draft before commit")
	}
	tt.SetClipboard("a,b")
	tt.Command("paste")
	if len(tags) != 3 || state.Query != "a,b" {
		t.Fatalf("paste=false committed: %v %q", tags, state.Query)
	}
	tt.Type("c")
	if len(tags) != 3 || state.Query != "a,b" {
		t.Fatal("typing after plain paste unexpectedly committed")
	}
	tt.Key(0, ui.KeyEnter)
	if len(tags) != 5 {
		t.Fatal("explicit commit after plain paste")
	}
	onPaste = true
	tt.Frame()
	tt.SetClipboard("one")
	tt.Command("paste")
	if len(tags) != 6 || tags[5] != "one" {
		t.Fatal("addOnPaste=true must also accept a single tag")
	}
	readOnly = true
	state.Query = "x"
	tt.Frame()
	click(t, tt, "Next")
	if len(tags) != 6 || state.Query != "x" {
		t.Fatal("readonly blur committed")
	}
}

func TestPropsManualTabsAndAccordion(t *testing.T) {
	selected, changes := "a", 0
	open := []string{"a"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Tabs(c, th, &selected, []nui.TabItem{{ID: "a", Label: "Tab A"}, {ID: "disabled", Label: "Disabled", Disabled: true}, {ID: "b", Label: "Tab B"}}, nui.TabsOptions{ActivationMode: nui.ActivationManual, OnChange: func(string) { changes++ }})
			nui.Accordion(c, th, &open, []nui.AccordionItem{{ID: "a", Label: "Section A", Content: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Body A") }}, {ID: "b", Label: "Section B", Content: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Body B") }}}, nui.AccordionOptions{Collapsible: nui.Set(false)})
		})
	}, 500, 400)
	click(t, tt, "Tab A")
	tt.Key(0, ui.KeyRight)
	if selected != "a" || !tt.Focused("Tab B") || changes != 0 {
		t.Fatal("manual arrows activated tab or failed to skip disabled")
	}
	tt.Key(0, ui.KeySpace)
	if selected != "b" || changes != 1 {
		t.Fatalf("manual Space activation value=%s calls=%d", selected, changes)
	}
	click(t, tt, "Section A")
	tt.Advance(time.Second)
	if !slices.Equal(open, []string{"a"}) || !tt.HasText("Body A") {
		t.Fatal("non-collapsible section closed")
	}
	click(t, tt, "Section B")
	tt.Advance(time.Second)
	if !slices.Equal(open, []string{"b"}) || !tt.HasText("Body B") {
		t.Fatal("non-collapsible accordion cannot switch")
	}
}

func TestPropsInputMenuFilteringCreationAndMultiple(t *testing.T) {
	var state nui.SearchState
	changes, created := 0, ""
	external, multiple := true, true
	opts := []nui.Option[string]{{Value: "a", Label: "Alpha", Description: "One"}, {Value: "b", Label: "Beta", Description: "Two"}}
	state.SetOptions(opts)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputMenu(c, th, &state, nui.InputMenuOptions{Label: "Search", Multiple: multiple, ExternalFilter: external, FilterFields: []string{"description"}, OnValuesChange: func() { changes++ }, OnCreate: func(s string) { created = s }})
			nui.Button(c, th, "Outside", nui.ButtonOptions{})
		})
	}, 500, 450)
	click(t, tt, "Search")
	tt.Type("remote")
	if !tt.HasText("Beta") {
		t.Fatal("external filter still filtered locally")
	}
	click(t, tt, "Beta")
	if !slices.Equal(state.SelectedValues.Values(), []string{"b"}) || state.Query != "" || changes != 1 {
		t.Fatalf("multiple select: %+v changes=%d", state, changes)
	}
	click(t, tt, "Search")
	tt.Key(0, ui.KeyBackspace)
	if state.SelectedValues.Len() != 0 || changes != 2 {
		t.Fatal("empty draft Backspace removal")
	}
	external, multiple = false, false
	tt.Frame()
	click(t, tt, "Search")
	tt.Type("Two")
	if !tt.HasText("Beta") || tt.HasText("Alpha") {
		t.Fatal("filterFields description")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("New item")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if created != "New item" {
		t.Fatalf("create keyboard result %q", created)
	}
	tt.Frame()
	if state.Query != "New item" {
		t.Fatal("create must retain query until caller changes it")
	}
}

func TestPropsTypedMultipleSelectIdentity(t *testing.T) {
	selected := []int{}
	options := []nui.Option[int]{{Value: 0, Label: "Zero"}, {Value: 1, Label: "One"}, {Value: 2, Label: "Disabled", Disabled: true}}
	changes := 0
	selectState := submittedSelect(options)
	selectedState := chosenSelection(selected)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMultiple(c, th, selectedState, selectState, nui.SelectOptions{Label: "Numbers", OnChange: func() { changes++ }}).Width(300)
	}, 500, 450)
	click(t, tt, "Numbers")
	click(t, tt, "Zero")
	click(t, tt, "One")
	if !slices.Equal(selectedState.Values(), []int{0, 1}) || changes != 2 {
		t.Fatalf("typed multiple %v calls=%d", selectedState.Values(), changes)
	}
	options = options[1:]
	selectState.SetOptions(options)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if !tt.HasText("Zero, One") {
		t.Fatal("missing selected label after options update")
	}
	click(t, tt, "Numbers")
	click(t, tt, "One")
	if !slices.Equal(selectedState.Values(), []int{0}) {
		t.Fatal("selection lost stable typed identity")
	}
}

func TestPropsHorizontalListboxAndLoop(t *testing.T) {
	var state nui.ListboxState
	selected := []string{}
	listSelection := chosenSelection(selected)
	state.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "x", Label: "Disabled", Disabled: true}, {ID: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, listSelection, &state, nui.ListboxOptions{Label: "Choices", Multiple: true, Orientation: nui.Set(nui.Horizontal), ItemWidth: 120, Loop: true, HighlightOnHover: nui.Set(false), SelectionBehavior: nui.SelectionReplace}).Width(400)
	}, 500, 200)
	click(t, tt, "Alpha")
	tt.Key(0, ui.KeyRight)
	if state.Active != "b" || !slices.Equal(listSelection.Values(), []string{"a"}) {
		t.Fatalf("horizontal navigation %s %v", state.Active, listSelection.Values())
	}
	tt.Key(0, ui.KeyRight)
	if state.Active != "a" {
		t.Fatal("loop")
	}
	r := bounds(t, tt, "Beta")
	tt.Move(r.X+5, r.Y+5)
	if state.Active != "a" {
		t.Fatal("highlightOnHover=false")
	}
	tt.Key(ui.Shift, ui.KeyEnd)
	if !slices.Equal(listSelection.Values(), []string{"a", "b"}) {
		t.Fatalf("replace range must skip disabled: %v", listSelection.Values())
	}
	tt.Key(0, ui.KeySpace)
	if !slices.Equal(listSelection.Values(), []string{"b"}) {
		t.Fatal("replace activation")
	}
}

func TestListboxFilterAndItemOptions(t *testing.T) {
	var state nui.ListboxState
	selected := []string{"tea"}
	changes, queries, selections, attempts, highlights := 0, 0, 0, 0, 0
	prevent, filter, loading, disabled := true, true, false, false
	items := []nui.ListboxItem{
		{ID: "heading", Label: "Drinks", Kind: nui.ListboxLabel},
		{ID: "prefix", Label: "Cafeteria"},
		{ID: "contains", Label: "A café"},
		{ID: "exact", Label: "Café", Options: &nui.ListboxItemOptions{OnSelect: func(e *nui.ListboxEvent) {
			attempts++
			if prevent {
				e.PreventDefault()
			}
		}}},
		{ID: "separator", Kind: nui.ListboxSeparator},
		{ID: "tea", Label: "Tea"},
	}
	listSelection := chosenSelection(selected)
	state.SetItems(items)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Disabled(disabled).Children(func() {
			nui.Listbox(c, th, listSelection, &state, nui.ListboxOptions{
				Label: "Choices", Filter: filter, Loading: loading, Color: nui.Neutral, Highlight: true,
				OnChange: func() { changes++ }, OnSelect: func(string) { selections++ },
				OnSearchTermChange: func(string) { queries++ },
				OnHighlight:        func(string) { highlights++ },
			}).Width(340)
		})
	}, 400, 300)
	if state.Active != "tea" {
		t.Fatal("selected item was not initially highlighted")
	}
	click(t, tt, "Search Choices")
	tt.Type("cafe")
	if state.Query != "cafe" || state.Active != "exact" || queries != 1 || tt.HasText("Tea") {
		t.Fatalf("filter or relevance mismatch: query=%q active=%q queries=%d", state.Query, state.Active, queries)
	}
	if highlights != 1 {
		t.Fatalf("query highlight notified %d times, want once", highlights)
	}
	tt.Key(0, ui.KeyDown)
	if state.Active != "prefix" || !tt.Focused("Search Choices") {
		t.Fatal("filter navigation did not retain editor focus or relevance order")
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(listSelection.Values(), []string{"tea"}) || changes != 0 || selections != 0 || attempts != 1 {
		t.Fatalf("prevented activation: values=%v changes=%d selections=%d attempts=%d", listSelection.Values(), changes, selections, attempts)
	}
	prevent = false
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(listSelection.Values(), []string{"exact"}) || changes != 1 || selections != 1 || attempts != 2 {
		t.Fatalf("filter Enter: values=%v changes=%d selections=%d attempts=%d", listSelection.Values(), changes, selections, attempts)
	}
	loading = true
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if tt.HasText("Cafeteria") || changes != 1 || attempts != 2 {
		t.Fatal("loading list still exposed or activated its rows")
	}
	loading = false
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	filter = false
	tt.Frame()
	if !tt.Focused("Choices") {
		t.Fatal("hiding the filter did not restore list focus")
	}
	tt.Key(0, ui.KeyHome)
	if state.Active != "prefix" {
		t.Fatal("keyboard navigation did not skip the structural heading")
	}
	disabled = true
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if changes != 1 || selections != 1 {
		t.Fatal("ancestor-disabled list changed selection")
	}
}

func TestPropsSliderOrientationPointerAndGap(t *testing.T) {
	for _, orientation := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
		for _, inverted := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%v", orientation, inverted), func(t *testing.T) {
				lo, hi, single := 20.0, 80.0, 50.0
				changes := 0
				singleChanges := 0
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					ui.Row(c).Padding(20).Gap(30).Children(func() {
						r := nui.RangeSlider(c, th, &lo, &hi, nui.RangeSliderOptions{Label: "Range", Min: 0, Max: 100, Step: 5, MinStepsBetweenThumbs: 2, Orientation: orientation, Inverted: inverted, OnChange: func(float64, float64) { changes++ }})
						if orientation == nui.Horizontal {
							r.Width(240)
						}
						s := nui.Slider(c, th, &single, nui.SliderOptions{Label: "Single", Inverted: inverted, Vertical: orientation == nui.Vertical, Step: 5, OnChange: func(float64) { singleChanges++ }})
						if orientation == nui.Horizontal {
							s.Width(240)
						}
					})
				}, 650, 300)
				r := bounds(t, tt, "Range")
				x, y := r.X+8+(r.W-16)*.3, r.Y+r.H/2
				if orientation == nui.Vertical {
					x, y = r.X+r.W/2, r.Y+8+(r.H-16)*.3
				}
				tt.Press(x, y)
				tt.Release(x, y)
				wantLow := (orientation == nui.Horizontal) != inverted
				if (wantLow && lo != 30) || (!wantLow && hi != 70) || changes != 1 {
					t.Fatalf("pointer range %g %g changes=%d", lo, hi, changes)
				}
				active := &hi
				if wantLow {
					active = &lo
				}
				wantArrow := *active + 5
				if inverted && orientation == nui.Vertical {
					wantArrow = *active - 5
				}
				tt.Key(0, ui.KeyUp)
				if *active != wantArrow || changes != 2 {
					t.Fatalf("range arrow %g want %g calls=%d", *active, wantArrow, changes)
				}
				tt.Key(0, ui.KeyEnd)
				if hi-lo < 10 {
					t.Fatal("thumb gap lost")
				}
				s := bounds(t, tt, "Single")
				x, y = s.X+8+(s.W-16)*.3, s.Y+s.H/2
				if orientation == nui.Vertical {
					x, y = s.X+s.W/2, s.Y+8+(s.H-16)*.3
				}
				tt.Press(x, y)
				tt.Release(x, y)
				want := 70.0
				if wantLow {
					want = 30
				}
				if single != want || singleChanges != 1 {
					t.Fatalf("single pointer %g want %g changes=%d", single, want, singleChanges)
				}
				tt.Key(0, ui.KeyUp)
				if inverted && orientation == nui.Vertical {
					want -= 5
				} else {
					want += 5
				}
				if single != want || singleChanges != 2 {
					t.Fatalf("single arrow value=%g changes=%d", single, singleChanges)
				}
				tt.Key(0, ui.KeyHome)
				if single != 0 {
					t.Fatal("Home must reach actual minimum")
				}
				if inverted {
					before := singleChanges
					back, forward := ui.KeyUp, ui.KeyDown
					if orientation == nui.Horizontal {
						back, forward = ui.KeyRight, ui.KeyLeft
					}
					tt.Key(0, back)
					if single != 0 || singleChanges != before {
						t.Fatal("inverted minimum boundary")
					}
					tt.Key(0, forward)
					if single != 5 || singleChanges != before+1 {
						t.Fatal("inverted step away from minimum")
					}
				}
			})
		}
	}
}

func TestPropsInputMenuOpenAndResetPolicies(t *testing.T) {
	var state nui.SearchState
	openOnFocus := false
	reset := false
	mode := nui.InputMenuCombobox
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputMenu(c, th, &state, nui.InputMenuOptions{Label: "Query", Mode: mode, OpenOnClick: nui.Set(false), OpenOnFocus: nui.Set(openOnFocus), ResetSearchTermOnBlur: nui.Set(reset), ResetSearchTermOnSelect: nui.Set(reset)})
			nui.Button(c, th, "Outside", nui.ButtonOptions{}).Absolute().Top(400).Left(350)
		})
	}, 500, 470)
	click(t, tt, "Query")
	if tt.HasText("Alpha") {
		t.Fatal("openOnClick=false")
	}
	tt.Type("Al")
	click(t, tt, "Alpha")
	if state.Query != "Al" || state.Selected != "a" {
		t.Fatalf("resetSelect=false: %+v", state)
	}
	click(t, tt, "Outside")
	if state.Query != "Al" {
		t.Fatal("resetBlur=false")
	}
	openOnFocus = true
	tt.Frame()
	tt.Key(ui.Shift, ui.KeyTab)
	if !tt.Focused("Query") || !tt.HasText("Beta") {
		t.Fatal("keyboard openOnFocus")
	}
	mode = nui.InputMenuAutocomplete
	reset = true
	tt.Frame()
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Free text")
	click(t, tt, "Outside")
	if state.Query != "Free text" {
		t.Fatal("autocomplete discarded free text")
	}
}

func TestPropsHoverPopoverAndControlledTooltip(t *testing.T) {
	open, tip, dismiss := false, false, true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(30).Gap(12).Children(func() {
			a := nui.Button(c, th, "Hover", nui.ButtonOptions{})
			nui.Popover(c, th, a, &open, nui.PopoverOptions{Label: "Hover panel", Mode: nui.PopoverHover, OpenDelay: 200 * time.Millisecond, CloseDelay: 100 * time.Millisecond, Dismissible: nui.Set(dismiss)}, func(th nui.Theme) { nui.Button(c, th, "Inside", nui.ButtonOptions{}) })
			a = nui.Button(c, th, "Help", nui.ButtonOptions{})
			nui.Tooltip(c, th, a, "Help text", nui.TooltipOptions{Open: &tip, Placement: nui.Set(nui.RightStart), Kbds: []string{"Ctrl", "K"}})
		})
	}, 600, 400)
	r := bounds(t, tt, "Hover")
	tt.Move(r.X+4, r.Y+4)
	tt.Advance(199 * time.Millisecond)
	if open {
		t.Fatal("hover delay ignored")
	}
	tt.Advance(2 * time.Millisecond)
	if !open || !tt.HasText("Inside") {
		t.Fatal("hover did not open")
	}
	p := bounds(t, tt, "Hover panel")
	tt.Move(r.X+4, (r.Y+r.H+p.Y)/2)
	tt.Advance(150 * time.Millisecond)
	if !open {
		t.Fatal("anchor gap closed popup")
	}
	tt.Move(p.X+8, p.Y+8)
	tt.Advance(150 * time.Millisecond)
	if !open {
		t.Fatal("hover panel did not keep popup open")
	}
	tt.Move(590, 390)
	tt.Advance(99 * time.Millisecond)
	if !open {
		t.Fatal("close delay ignored")
	}
	tt.Advance(2 * time.Millisecond)
	if open {
		t.Fatal("hover did not close")
	}
	tt.Move(r.X+4, r.Y+4)
	tt.Advance(201 * time.Millisecond)
	dismiss = false
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if !open {
		t.Fatal("non-dismissible Escape")
	}
	dismiss = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if open {
		t.Fatal("dismissed hover reopened without pointer leaving")
	}
	tip = true
	tt.Frame()
	if !tt.HasText("Help text") || !tt.HasText("K") {
		t.Fatal("controlled tooltip/keys")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if tip || tt.HasText("Help text") {
		t.Fatal("controlled tooltip Escape")
	}
}

func TestPropsOrientationPreservesFieldAndCarousel(t *testing.T) {
	direction := nui.Vertical
	text := "original"
	var carousel nui.CarouselState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(20).Children(func() {
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Name", Orientation: nui.Set(direction)}, func(th nui.Theme) { nui.Input(c, th, &text, nui.InputOptions{Label: "Editor"}) })
			nui.Carousel(c, th, &carousel, []nui.CarouselItem{{ID: "a", Label: "Slide A"}, {ID: "b", Label: "Slide B"}, {ID: "c", Label: "Slide C"}}, nui.CarouselOptions{Label: "Guide", Orientation: direction, Arrows: nui.Set(false), Dots: true}).Width(400)
		})
	}, 600, 500)
	click(t, tt, "Name")
	if !tt.Focused("Editor") {
		t.Fatal("field label association")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	direction = nui.Horizontal
	tt.Frame()
	tt.Type("edited")
	if text != "edited" || !tt.Focused("Editor") {
		t.Fatal("orientation change lost editor identity/selection")
	}
	direction = nui.Vertical
	tt.Frame()
	if tt.HasText("Next Guide") {
		t.Fatal("arrows=false")
	}
	click(t, tt, "Guide slides")
	tt.Key(0, ui.KeyDown)
	if carousel.Index != 1 {
		t.Fatal("vertical carousel keyboard")
	}
	click(t, tt, "Go to slide 3")
	if carousel.Index != 2 {
		t.Fatal("dots navigation")
	}
	click(t, tt, "Guide slides")
	tt.Key(0, ui.KeyHome)
	r := bounds(t, tt, "Guide slides")
	tt.Press(r.X+40, r.Y+140)
	tt.Move(r.X+40, r.Y+20)
	tt.Advance(time.Second)
	tt.Release(r.X+40, r.Y+20)
	if carousel.Index != 1 {
		t.Fatalf("vertical drag %d", carousel.Index)
	}
}

func TestPropsToastActionsAndProgress(t *testing.T) {
	actions, closed := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
			nui.Toast(c, th, nui.ToastOptions{Title: "Ready", Description: "Two actions", Timeout: time.Second, Orientation: nui.Set(nui.Horizontal), Close: nui.Set(false), OnClose: func() { closed++ }, Actions: []nui.ToastActionOptions{{Label: "Keep", Close: nui.Set(false), OnAction: func() { actions++ }}, {Label: "Done", OnAction: func() { actions++ }}}})
		})
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 650, 350)
	click(t, tt, "Notify")
	if tt.HasText("Dismiss Ready") {
		t.Fatal("close=false")
	}
	click(t, tt, "Keep")
	tt.Advance(2 * time.Second)
	if !tt.HasText("Ready") || actions != 1 || closed != 0 {
		t.Fatal("action retention / focused toast pause")
	}
	click(t, tt, "Done")
	tt.Advance(time.Second)
	if tt.HasText("Ready") || actions != 2 || closed != 1 {
		t.Fatalf("toast close actions=%d close=%d", actions, closed)
	}
}

func TestPropsRatingClearableAndPinSeparator(t *testing.T) {
	rating, pin := 3, "123456"
	var state nui.PinState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Gap(24).Children(func() {
			star := 0
			nui.InputRating(c, th, &rating, nui.InputRatingOptions{Label: "Stars", Orientation: nui.Vertical, Size: nui.LG, Clearable: true, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Item {
					star++
					e.Label(fmt.Sprintf("Star %d", star))
				}
			}})
			i := 0
			nui.PinInput(c, th, &pin, &state, nui.PinInputOptions{Numeric: true, Length: 6, Label: "PIN", Size: nui.LG, Separator: []int{3}, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Item {
					i++
					e.Label(fmt.Sprintf("Slot %d", i))
				}
			}})
		})
	}, 600, 350)
	r := bounds(t, tt, "Stars")
	tt.Move(r.X+r.W/2, r.Y+3)
	if rating != 3 {
		t.Fatal("hover changed value")
	}
	click(t, tt, "Star 3")
	if rating != 0 {
		t.Fatalf("clearable rating %d", rating)
	}
	click(t, tt, "Slot 4")
	tt.Type("9")
	if pin != "123956" {
		t.Fatalf("separator hit mapping %s", pin)
	}
}

func TestPropsFileUploadPresentationAndInteraction(t *testing.T) {
	var state nui.FileUploadState
	var calls atomic.Int32
	interactive, preview, remove := false, true, false
	choose := func(context.Context) ([]nui.UploadFile, error) {
		calls.Add(1)
		return []nui.UploadFile{{ID: "one", Name: "one.txt", Size: 12}}, nil
	}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FileUpload(c, th, &state, nui.FileUploadOptions{Label: "Files", Interactive: nui.Set(interactive), Preview: nui.Set(preview), FileDelete: nui.Set(remove), Dropzone: nui.Set(false), Layout: nui.UploadGrid, Choose: choose}).Width(450)
	}, 500, 400)
	t.Cleanup(state.Dispose)
	r := bounds(t, tt, "Files")
	tt.Press(r.X+10, r.Y+10)
	tt.Release(r.X+10, r.Y+10)
	if calls.Load() != 0 || state.Pending() {
		t.Fatal("interactive=false opened picker")
	}
	click(t, tt, "Choose files")
	until := time.Now().Add(time.Second)
	for state.Pending() && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
		tt.Frame()
	}
	if calls.Load() != 1 || !tt.HasText("one.txt") || tt.HasText("Remove one.txt") {
		t.Fatal("choose/fileDelete behavior")
	}
	preview = false
	tt.Frame()
	if tt.HasText("one.txt") || len(state.Files) != 1 {
		t.Fatal("preview=false changed file data")
	}
	preview, remove = true, true
	tt.Frame()
	click(t, tt, "Remove one.txt")
	if len(state.Files) != 0 {
		t.Fatal("remove option")
	}
}
