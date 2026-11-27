package nui

import (
	"math"
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
)

func expectSubmissionPanic(t *testing.T, submit func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("invalid submission was accepted")
		}
	}()
	submit()
}

func TestSelectSubmissionOwnsDataAndValidatesBeforeCommit(t *testing.T) {
	var s SelectState[float64]
	called := ""
	options := []Option[float64]{{Value: 0, Label: "Zero", Build: func(*ui.Context, Theme) { called = "original" }}}
	s.SetOptions(options)
	options[0].Label = "Changed"
	options[0].Build = func(*ui.Context, Theme) { called = "replacement" }
	s.source.items[0].Build(nil, Theme{})
	if s.source.items[0].Label != "Zero" || called != "original" || s.source.metadata.index[0] != 0 {
		t.Fatal("submission aliased descriptors or lost a typed zero")
	}
	for _, invalid := range [][]Option[float64]{
		{{Value: 1}, {Value: 1}}, {{Value: 1}, {Value: math.NaN()}},
	} {
		expectSubmissionPanic(t, func() { s.SetOptions(invalid) })
		if s.source.revision != 1 || len(s.source.items) != 1 || s.source.items[0].Label != "Zero" {
			t.Fatal("invalid update replaced the usable option model")
		}
	}
	s.SetOptions(options)
	s.source.items[0].Build(nil, Theme{})
	if s.source.items[0].Label != "Changed" || called != "replacement" || s.source.revision != 2 {
		t.Fatal("same-ID descriptor replacement was ignored")
	}
	s.SetOptions(options[:0])
	if s.source.items != nil || s.source.metadata.index != nil || s.source.metadata.keys != nil {
		t.Fatal("empty submission retained option storage")
	}
}

func TestInputMenuSubmissionKeepsDraftAndRejectsEmptyIDs(t *testing.T) {
	s := SearchState{Query: "draft", Selected: "a", Open: true}
	s.SetOptions([]Option[string]{{Value: "a", Label: "Alpha"}})
	expectSubmissionPanic(t, func() { s.SetOptions([]Option[string]{{Value: "", Label: "Invalid"}}) })
	if s.source.items[0].Label != "Alpha" || s.source.revision != 1 {
		t.Fatal("invalid input-menu data was committed")
	}
	s.SetOptions(nil)
	if s.Query != "draft" || s.Selected != "a" || !s.Open || s.source.items != nil || s.source.metadata.index != nil {
		t.Fatal("empty options reset independent input state or retained data")
	}
}

func TestListboxSubmissionOwnsNestedDescriptors(t *testing.T) {
	color := ui.RGB(10, 20, 30)
	styles := Styles{Item: {Background: Set(color)}}
	called := ""
	items := []ListboxItem{{ID: "a", Label: "Alpha",
		Avatar:  &AvatarOptions{Text: "A", Decoration: Decoration{UI: styles}},
		Chip:    &ChipOptions{Text: "Chip", Decoration: Decoration{UI: styles}},
		Options: &ListboxItemOptions{Decoration: Decoration{UI: styles}, OnSelect: func(*ListboxEvent) { called = "original" }},
	}}
	var s ListboxState
	s.SetItems(items)
	styles[Item] = PartStyle{Background: Set(ui.Transparent)}
	items[0].Label, items[0].Avatar.Text, items[0].Chip.Text = "Changed", "B", "Other"
	items[0].Options.OnSelect = func(*ListboxEvent) { called = "replacement" }
	owned := s.source.items[0]
	owned.Options.OnSelect(&ListboxEvent{})
	if owned.Label != "Alpha" || owned.Avatar.Text != "A" || owned.Chip.Text != "Chip" || called != "original" {
		t.Fatal("listbox submission aliased a nested descriptor")
	}
	for _, style := range []Styles{owned.Options.UI, owned.Avatar.UI, owned.Chip.UI} {
		if style[Item].Background.Value != color {
			t.Fatal("listbox submission aliased a style map")
		}
	}
	s.SetItems(items)
	s.source.items[0].Options.OnSelect(&ListboxEvent{})
	if called != "replacement" || s.source.items[0].Label != "Changed" {
		t.Fatal("listbox same-ID replacement was ignored")
	}
	for _, invalid := range [][]ListboxItem{{{ID: ""}}, {{ID: "a"}, {ID: "a"}}, {{ID: "a", Kind: 255}}} {
		expectSubmissionPanic(t, func() { s.SetItems(invalid) })
		if s.source.revision != 2 || s.source.items[0].Label != "Changed" {
			t.Fatal("invalid listbox update damaged the previous data")
		}
	}
	s.Query = "retained query"
	s.SetItems(items[:0])
	if s.source.items != nil || s.source.index != nil || s.Query != "retained query" {
		t.Fatal("empty listbox submission retained data or reset the query")
	}
}

func TestSelectMenuSubmissionUpdatesBothModels(t *testing.T) {
	s := SelectMenuState{Query: "draft", Open: true}
	items := []ListboxItem{{ID: "a", Label: "Alpha"}}
	s.SetItems(items)
	items[0].Label = "Replacement"
	if s.data.items[0].Label != "Alpha" || s.Listbox.source.items[0].Label != "Alpha" {
		t.Fatal("trigger or list aliased the caller's data")
	}
	s.SetItems(items)
	if s.data.items[0].Label != "Replacement" || s.Listbox.source.items[0].Label != "Replacement" {
		t.Fatal("trigger and list did not receive the same revision")
	}
	s.SetItems(nil)
	if s.data.items != nil || s.Listbox.source.items != nil || s.Query != "draft" || !s.Open {
		t.Fatal("empty submission retained data or reset independent popup state")
	}
}

func TestMenuSubmissionOwnsTreeAndBindings(t *testing.T) {
	query, called := "draft", ""
	styles := Styles{Item: {Radius: Set(float32(9))}}
	fields := []string{"label"}
	items := []MenuItem{{ID: "parent", Label: "Parent", UI: styles, Avatar: &AvatarOptions{Text: "P", Decoration: Decoration{UI: styles}},
		Content:  &MenuContentOptions{SearchTerm: &query, FilterFields: fields, Decoration: Decoration{UI: styles}, FilterInput: &InputOptions{Label: "Search", UI: styles, Avatar: &AvatarOptions{Text: "Q", Decoration: Decoration{UI: styles}}}},
		Children: []MenuItem{{ID: "child", Label: "Child", OnSelect: func() { called = "original" }}},
	}}
	var s MenuState
	s.SetItems(items)
	items[0].Children[0].Label = "Changed"
	items[0].Children[0].OnSelect = func() { called = "replacement" }
	items[0].Avatar.Text, items[0].Content.FilterInput.Label = "Other", "Other search"
	fields[0] = "description"
	styles[Item] = PartStyle{Radius: Set(float32(42))}
	owned := s.source.items[0]
	owned.Children[0].OnSelect()
	if owned.Children[0].Label != "Child" || called != "original" || owned.Avatar.Text != "P" || owned.Content.FilterInput.Label != "Search" || owned.Content.FilterFields[0] != "label" {
		t.Fatal("menu submission aliased nested data")
	}
	for _, style := range []Styles{owned.UI, owned.Avatar.UI, owned.Content.UI, owned.Content.FilterInput.UI, owned.Content.FilterInput.Avatar.UI} {
		if style[Item].Radius.Value != 9 {
			t.Fatal("menu submission aliased a style map")
		}
	}
	query = "live binding"
	if *owned.Content.SearchTerm != query {
		t.Fatal("submission detached an intentional app-owned query binding")
	}
	s.SetItems(items)
	s.source.byID["child"].item.OnSelect()
	if called != "replacement" || s.source.byID["child"].item.Label != "Changed" {
		t.Fatal("same-ID menu callback/label replacement was ignored")
	}
	for _, invalid := range [][]MenuItem{
		{{ID: ""}}, {{ID: "x", Children: []MenuItem{{ID: "x"}}}}, {{ID: "bad", Kind: 255}},
	} {
		expectSubmissionPanic(t, func() { s.SetItems(invalid) })
		if s.source.revision != 2 || s.source.byID["child"].item.Label != "Changed" {
			t.Fatal("invalid menu update damaged the previous tree")
		}
	}
	deep := []MenuItem{{ID: "leaf"}}
	for i := range 33 {
		deep = []MenuItem{{ID: strconv.Itoa(i), Children: deep}}
	}
	expectSubmissionPanic(t, func() { s.SetItems(deep) })
	s.SetItems(nil)
	if s.source.items != nil || s.source.byID != nil || s.source.levels != nil {
		t.Fatal("empty menu submission retained data")
	}
}
