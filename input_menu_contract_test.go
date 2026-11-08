package nui_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestInputMenuDefaultClickAndExplicitOpening(t *testing.T) {
	var state nui.SearchState
	openOnClick := nui.Optional[bool]{}
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Query", OpenOnClick: openOnClick}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	if tt.HasText("Alpha") {
		t.Fatal("default click unexpectedly opened InputMenu")
	}
	tt.Key(0, ui.KeyDown)
	if !tt.HasText("Alpha") {
		t.Fatal("Down did not open InputMenu")
	}
	tt.Key(0, ui.KeyEscape)
	openOnClick = nui.Set(true)
	tt.Frame()
	click(t, tt, "Query")
	if !tt.HasText("Alpha") {
		t.Fatal("explicit openOnClick failed")
	}
}

func TestInputMenuControlledOpenLoopAndHover(t *testing.T) {
	state, open := nui.SearchState{}, false
	loop, hover := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "x", Label: "Blocked", Disabled: true}, {Value: "b", Label: "Beta"}}, nui.InputMenuOptions{Label: "Query", Open: &open, Loop: loop, HighlightOnHover: nui.Set(hover)}).Width(260)
	}, 360, 320)
	open = true
	tt.Frame()
	if !open || !tt.Focused("Query") || !tt.HasText("Beta") {
		t.Fatal("controlled open did not focus and show input")
	}
	r := bounds(t, tt, "Beta")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "a" || open {
		t.Fatal("hover=false changed highlight")
	}
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" {
		t.Fatal("loop=false did not stop at last enabled item")
	}
	loop = true
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "a" {
		t.Fatal("loop did not wrap past disabled item")
	}
	hover = true
	open = true
	tt.Frame()
	r = bounds(t, tt, "Beta")
	tt.Move(r.X+2, r.Y+2)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" {
		t.Fatal("hover=true failed to move highlight")
	}
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if open || !tt.Focused("Query") {
		t.Fatal("Escape did not update open state or retain input focus")
	}
}

func TestInputMenuClearResetPolicyAndReadOnly(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(fmt.Sprint(multiple), func(t *testing.T) {
			state := nui.SearchState{Selected: "a", Query: "Draft", SelectedValues: []string{"a"}}
			reset, readonly, clears, changes := false, false, 0, 0
			tt := ui.NewTester(func(c *ui.Context) {
				nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Query", Multiple: multiple, Clearable: true, ReadOnly: readonly, ResetModelValueOnClear: nui.Set(reset), OnClear: func() { clears++ }, OnValuesChange: func([]string) { changes++ }}).Width(280)
			}, 380, 300)
			click(t, tt, "Clear Query")
			if state.Selected != "a" || !slices.Equal(state.SelectedValues, []string{"a"}) || state.Query != "" || clears != 1 || changes != 0 {
				t.Fatalf("retained clear: %+v %d/%d", state, clears, changes)
			}
			reset = true
			readonly = true
			tt.Frame()
			if _, ok := tt.Find("Clear Query"); ok {
				t.Fatal("read-only model exposed clear")
			}
			readonly = false
			tt.Frame()
			click(t, tt, "Clear Query")
			if state.Selected != "" || multiple && len(state.SelectedValues) != 0 || clears != 2 || multiple && changes != 1 {
				t.Fatalf("reset clear: %+v %d/%d", state, clears, changes)
			}
		})
	}
}

func TestInputMenuIconsAndLoadingDoNotChangeState(t *testing.T) {
	state := nui.SearchState{SelectedValues: []string{"a"}}
	queries, values, selected := 0, 0, 0
	loader := false
	marker := ui.MustParseSVG([]byte(`<svg viewBox="0 0 10 10"><rect width="10" height="10"/></svg>`))
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		nui.InputMenu(c, th, &state, []nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First choice"}, {Value: "b", Label: "Beta", Description: "Second choice"}}, nui.InputMenuOptions{Label: "Query", Multiple: true, Clearable: true, SelectedIcon: marker, DeleteIcon: marker, ClearIcon: marker, Avatar: &nui.AvatarOptions{Text: "AL", Alt: "Query avatar"}, Loading: loader, LoadingTrailing: true, LoadingIcon: lucide.Heart(), OnChange: func(string) { queries++ }, OnSelect: func(string) { selected++ }, OnValuesChange: func([]string) { values++ }, UI: nui.Styles{nui.Item: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}}).Width(280)
	}, 380, 330)
	delete := bounds(t, tt, "Remove Alpha")
	if color := tt.Image().RGBAAt(int(delete.X+1), int(delete.Y+1)); color != rgba(nui.Light().Dimmed) {
		t.Fatalf("custom square delete icon was not used: %v", color)
	}
	clear := bounds(t, tt, "Clear Query")
	if color := tt.Image().RGBAAt(int(clear.X+1), int(clear.Y+1)); color != rgba(nui.Light().Muted) {
		t.Fatalf("custom square clear icon was not used: %v", color)
	}
	click(t, tt, "Query")
	loader = true
	tt.Frame()
	tt.Type("Be")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(state.SelectedValues, []string{"a", "b"}) || values != 1 || selected != 1 || queries != 1 || !tt.HasText("Query avatar") {
		t.Fatalf("loading select state=%+v values/select/query=%d/%d/%d", state, values, selected, queries)
	}
	click(t, tt, "Remove Alpha")
	if !slices.Equal(state.SelectedValues, []string{"b"}) || values != 2 || selected != 1 || !tt.Focused("Query") {
		t.Fatal("tag removal changed focus or selection event")
	}
	click(t, tt, "Clear Query")
	if len(state.SelectedValues) != 0 || values != 3 || queries != 1 {
		t.Fatal("model-only clear reported an input edit")
	}
}

func TestInputMenuExplicitDismissalAndDisabledOpen(t *testing.T) {
	state, open, disabled := nui.SearchState{}, false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputMenu(c, th, &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Query", Open: &open, Disabled: disabled, Popup: nui.PopoverOptions{Dismissible: nui.Set(false)}}).Width(260)
			nui.Button(c, th, "Outside", nui.ButtonOptions{}).Absolute().Top(260)
		})
	}, 360, 320)
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if !open {
		t.Fatal("Dismissible=false allowed Escape")
	}
	click(t, tt, "Outside")
	if !open || !tt.Focused("Query") {
		t.Fatal("Dismissible=false allowed an outside focus to close the list")
	}
	open = false
	tt.Frame()
	if open || tt.HasText("Alpha") {
		t.Fatal("explicit close was blocked")
	}
	disabled = true
	open = true
	tt.Frame()
	if open || tt.HasText("Alpha") {
		t.Fatal("disabled input accepted an external open")
	}
}

func TestInputMenuClearSelectedTagsWithoutDraft(t *testing.T) {
	state := nui.SearchState{SelectedValues: []string{"a", "b"}}
	clears, changes, queries := 0, 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.InputMenuOptions{Label: "Query", Multiple: true, Clearable: true, OnClear: func() { clears++ }, OnValuesChange: func([]string) { changes++ }, OnChange: func(string) { queries++ }}).Width(300)
	}, 400, 300)
	if _, ok := tt.Find("Clear Query"); !ok {
		t.Fatal("selected tags with an empty draft cannot be cleared")
	}
	click(t, tt, "Clear Query")
	if len(state.SelectedValues) != 0 || clears != 1 || changes != 1 || queries != 0 || !tt.Focused("Query") {
		t.Fatalf("clear result: %+v callbacks %d/%d/%d", state, clears, changes, queries)
	}
	tt.Frame()
	if clears != 1 || changes != 1 {
		t.Fatal("clear callback repeated")
	}
}

func TestInputMenuAutocompleteIgnoresSelectionReset(t *testing.T) {
	state := nui.SearchState{}
	selected, submitted := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "alpha", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Query", Mode: nui.InputMenuAutocomplete, Multiple: true, ResetSearchTermOnSelect: nui.Set(false), ResetSearchTermOnBlur: nui.Set(true), OnSelect: func(string) { selected++ }, OnSubmit: func(string) { submitted++ }}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	tt.Type("Al")
	tt.Key(0, ui.KeyEnter)
	if state.Query != "Alpha" || state.Selected != "alpha" || selected != 1 || submitted != 0 || len(state.SelectedValues) != 0 {
		t.Fatalf("autocomplete suggestion: %+v %d/%d", state, selected, submitted)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Free text")
	tt.Key(0, ui.KeyEnter)
	if state.Query != "Free text" || selected != 1 || submitted != 1 {
		t.Fatal("autocomplete free submit changed the suggestion or duplicated events")
	}
}

func TestInputMenuDisabledResultsAreNotEmpty(t *testing.T) {
	var state nui.SearchState
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "a", Label: "Unavailable", Disabled: true}}, nui.InputMenuOptions{Label: "Query"}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	tt.Type("Unavailable")
	if !tt.HasText("Unavailable") || tt.HasText("No matching data") {
		t.Fatal("visible disabled result was reported as no results")
	}
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "" {
		t.Fatal("disabled result was selected")
	}
}

func TestInputMenuControlledSelectionPreservesActiveDraft(t *testing.T) {
	state := nui.SearchState{Selected: "a"}
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputMenu(c, th, &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.InputMenuOptions{Label: "Query", OnChange: func(string) { changes++ }}).Width(260)
			nui.Button(c, th, "Outside", nui.ButtonOptions{}).Absolute().Top(260)
		})
	}, 360, 320)
	if state.Query != "Alpha" {
		t.Fatalf("initial selection display=%q", state.Query)
	}
	state.Selected = "b"
	tt.Frame()
	if state.Query != "Beta" || changes != 0 {
		t.Fatal("external selected value did not update idle display silently")
	}
	click(t, tt, "Query")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Draft")
	state.Selected = "a"
	tt.Frame()
	if state.Query != "Draft" || changes != 1 {
		t.Fatal("external value discarded the active draft")
	}
	click(t, tt, "Outside")
	if state.Query != "Alpha" || changes != 1 {
		t.Fatal("blur did not restore the current selected label silently")
	}
}

func TestInputMenuSelectionReportsBothCallbacksOnce(t *testing.T) {
	state := nui.SearchState{}
	var events []string
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Query", Multiple: true, OnSelect: func(id string) { events = append(events, "select:"+id) }, OnValuesChange: func([]string) { events = append(events, "values") }}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	tt.Type("Al")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !slices.Equal(events, []string{"select:a", "values"}) {
		t.Fatalf("selection events=%v", events)
	}
}

func TestInputMenuKeyboardTagDeleteReturnsToEditor(t *testing.T) {
	state := nui.SearchState{SelectedValues: []string{"a", "b"}}
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.InputMenuOptions{Label: "Query", Multiple: true, OnValuesChange: func([]string) { changes++ }}).Width(300)
	}, 400, 300)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Remove Alpha") {
		t.Fatal("tag delete is not reachable by keyboard")
	}
	tt.Key(0, ui.KeySpace)
	if !slices.Equal(state.SelectedValues, []string{"b"}) || changes != 1 || !tt.Focused("Query") {
		t.Fatal("keyboard deletion lost editor focus or notified twice")
	}
}
