package nui_test

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// While an input method composes, keys without Ctrl or Super go to the input
// method, as each platform's filter has them, and reach no component action.
// Each test also checks the committed or canceled path, so that a MyGo change
// that kept withholding keys after the composition ends fails too.

func TestCompositionInputEnterAndEscapeDoNotSubmit(t *testing.T) {
	value := ""
	submits, changes := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Input(c.Key("name"), th, &value, nui.InputOptions{Label: "Name", OnChange: func(string) { changes++ }, OnSubmit: func(string) { submits++ }}).Width(260)
	}, 400, 150)
	click(t, tt, "Name")
	tt.Compose("lin", 3)
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeyEscape)
	if submits != 0 || value != "" || changes != 0 {
		t.Fatalf("composing keys reached the field: submits=%d value=%q changes=%d", submits, value, changes)
	}
	tt.Type("林")
	if value != "林" || changes != 1 || submits != 0 {
		t.Fatalf("commit: value=%q changes=%d submits=%d", value, changes, submits)
	}
	tt.Key(0, ui.KeyEnter)
	if submits != 1 {
		t.Fatalf("Enter after the commit submitted %d times", submits)
	}
}

func TestCompositionInputTagsKeysChangeNothing(t *testing.T) {
	var tags []string
	var state nui.TagsState
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputTags(c.Key("tags"), th, &tags, &state, nui.InputTagsOptions{Label: "Tags", AddOnTab: true, OnTagsChange: func([]string) { changes++ }})
			nui.Button(c.Key("next"), th, "Next", nui.ButtonOptions{})
		})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.Compose("abc", 3)
	for _, key := range []ui.Key{ui.KeyEnter, ui.KeyTab, ui.KeyBackspace} {
		tt.Key(0, key)
		if len(tags) != 0 || changes != 0 || state.Query != "" || !tt.Focused("Tags") {
			t.Fatalf("key %v while composing: tags=%v changes=%d query=%q focused=%v", key, tags, changes, state.Query, tt.Focused("Tags"))
		}
	}
	tt.Type("阿")
	if state.Query != "阿" || len(tags) != 0 {
		t.Fatalf("commit: query=%q tags=%v", state.Query, tags)
	}
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(tags, []string{"阿"}) || changes != 1 || state.Query != "" {
		t.Fatalf("Enter after the commit: tags=%v changes=%d query=%q", tags, changes, state.Query)
	}
}

func TestCompositionInputTagsBackspaceKeepsExistingTag(t *testing.T) {
	tags := []string{"one"}
	var state nui.TagsState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTags(c.Key("tags"), th, &tags, &state, nui.InputTagsOptions{Label: "Tags"})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.Compose("abc", 3)
	tt.Key(0, ui.KeyBackspace)
	tt.Key(0, ui.KeyBackspace)
	if !slices.Equal(tags, []string{"one"}) {
		t.Fatalf("Backspace while composing removed a tag: %v", tags)
	}
	// Canceled: the first Backspace marks the last tag, the second removes it.
	tt.Compose("", 0)
	tt.Key(0, ui.KeyBackspace)
	tt.Key(0, ui.KeyBackspace)
	if len(tags) != 0 {
		t.Fatalf("Backspace after canceling: %v", tags)
	}
}

func TestCompositionCommandPaletteSearchKeepsHighlight(t *testing.T) {
	var s nui.CommandPaletteState
	s.SetIndex(nui.NewCommandIndex([]nui.CommandAction{{ID: "alpha", Label: "Alpha"}, {ID: "beta", Label: "Beta"}, {ID: "gamma", Label: "Gamma"}}))
	selected := ""
	var highlights []string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{OnSelect: func(id string) { selected = id }, OnHighlight: func(id string) { highlights = append(highlights, id) }})
	}, 700, 600)
	s.Open = true
	tt.Frame()
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading && len(s.Results.Items) == 3 })
	click(t, tt, "Search Command palette")
	before := slices.Clone(highlights)
	tt.Compose("be", 2)
	// The first Down shows the first result's highlight; the second moves it.
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(highlights, before) || selected != "" || !s.Open || s.Query != "" {
		t.Fatalf("composing keys reached the palette: highlights=%v (was %v) selected=%q open=%v query=%q", highlights, before, selected, s.Open, s.Query)
	}
	// Canceled: the same keys move to the second result and choose it.
	tt.Compose("", 0)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	if !slices.Equal(highlights, append(before, "beta")) {
		t.Fatalf("Down after canceling: %v", highlights)
	}
	tt.Key(0, ui.KeyEnter)
	if selected != "beta" || s.Open {
		t.Fatalf("Enter after canceling: selected=%q open=%v", selected, s.Open)
	}
	// Committed: the text becomes the query, and Enter chooses its match.
	selected = ""
	s.Open = true
	tt.Frame()
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading })
	click(t, tt, "Search Command palette")
	tt.Compose("gam", 3)
	tt.Type("Gamma")
	awaitCommandResults(t, tt, func() bool { return !s.Results.Loading && len(s.Results.Items) == 1 })
	if s.Query != "Gamma" || selected != "" {
		t.Fatalf("commit: query=%q selected=%q", s.Query, selected)
	}
	tt.Key(0, ui.KeyEnter)
	if selected != "gamma" || s.Open {
		t.Fatalf("Enter after the commit: selected=%q open=%v", selected, s.Open)
	}
}

func TestCompositionModalEscapeKeepsItOpen(t *testing.T) {
	open := true
	value := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Modal(c.Key("dialog"), th, &open, "Rename", func(inner nui.Theme) {
			nui.Input(c.Key("name"), inner, &value, nui.InputOptions{Label: "Name"}).FillWidth()
		})
	}, 600, 420)
	click(t, tt, "Name")
	tt.Compose("ni", 2)
	tt.Key(0, ui.KeyEscape)
	if !open || value != "" {
		t.Fatalf("Escape while composing: open=%v value=%q", open, value)
	}
	tt.Compose("", 0)
	tt.Key(0, ui.KeyEscape)
	if open || value != "" {
		t.Fatalf("Escape after canceling: open=%v value=%q", open, value)
	}
}

func TestCompositionShortcutChordRunsAction(t *testing.T) {
	var s nui.CommandPaletteState
	runs := 0
	s.SetIndex(nui.NewCommandIndex([]nui.CommandAction{{ID: "quick", Label: "Quick", Key: ui.KeyJ, Mods: ui.Ctrl, OnSelect: func() { runs++ }}}))
	value := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Input(c.Key("search"), th, &value, nui.InputOptions{Label: "Search"}).Width(260)
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{})
	}, 700, 600)
	click(t, tt, "Search")
	tt.Compose("q", 1)
	tt.Key(ui.Ctrl, ui.KeyJ)
	if runs != 1 || value != "" || s.Open {
		t.Fatalf("Ctrl+J while composing: runs=%d value=%q open=%v", runs, value, s.Open)
	}
	// The composition stayed: committing it still writes the value once.
	tt.Type("去")
	if value != "去" || runs != 1 {
		t.Fatalf("commit after the chord: value=%q runs=%d", value, runs)
	}
}
