<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# SelectMenu

Search and choose items in a popup list.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `SelectMenu`
- Types: `SelectMenuOptions`
- Theme key: `nui.ComponentSelectMenu` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

SelectMenu combines a trigger, search input and Listbox. Selected values are IDs; SelectMenuState stores open state, query and persistent focus handles. FilterInput configures the search field, and FocusSearch chooses its initial focus.

## Things to know

Selection and dismissal reset the query by default. Clear always clears the query; ResetModelValueOnClear only controls selection. Loading shows a trigger spinner and native status row while existing choices remain usable. ExternalFilter or IgnoreFilter bypasses local filtering. QueryState rejects stale results; publish real background results through Window.Update.

## API

```go
func SelectMenu(c *ui.Context, t Theme, value *[]string, s *SelectMenuState, items []ListboxItem, o SelectMenuOptions) ui.Element
```

### SelectMenuOptions

```go
type SelectMenuOptions struct {
	ListboxOptions
	Color                                          Color
	Variant                                        Variant
	Size                                           Size
	SearchInput                                    Optional[bool]
	FocusSearch                                    Optional[bool]
	ResetSearchTermOnBlur, ResetSearchTermOnSelect Optional[bool]
	ResetModelValueOnClear                         Optional[bool]
	Placeholder                                    string
	Invalid                                        bool
	ErrorMessage                                   string
	SearchPlaceholder                              string
	Clearable, Loading, ExternalFilter             bool
	Leading, Trailing, LoadingIcon, ClearIcon      *ui.SVG
	Avatar                                         *AvatarOptions
	LoadingTrailing                                bool
	ShowTrailing                                   Optional[bool]
	ClearOptions                                   *ButtonOptions
	Error                                          string
	OnQuery                                        func(string)
	OnCreate                                       func(string)
	OnClear                                        func()
	OnOpenChange                                   func(bool)
	FilterFields                                   []string // label (default), description, value/id, group.
	CreateWhenAlways, CreateAtTop                  bool
	Popup                                          PopoverOptions
}
```

### Related types

```go
type SelectMenuState struct {
	Open        bool
	Query       string
	Handle      ui.Handle
	QueryHandle ui.Handle
	Listbox     ListboxState
	// contains filtered or unexported fields
}
```

```go
// ListboxItem remains comparable; optional per-item behavior and styles live
// behind Options so existing maps and equality-based data caches remain usable.
type ListboxItem struct {
	ID, Label, Description, Group string
	Kind                          ListboxItemKind
	Disabled                      bool
	Icon                          *ui.SVG
	Avatar                        *AvatarOptions
	Chip                          *ChipOptions
	Options                       *ListboxItemOptions
}
```

```go
type ListboxState struct {
	Active      string
	Query       string
	QueryHandle ui.Handle
	List        ui.ListState
	// contains filtered or unexported fields
}
```

```go
type ListboxItemKind uint8

const (
	ListboxOption ListboxItemKind = iota
	ListboxLabel
	ListboxSeparator
)
```

```go
type ListboxItemOptions struct {
	OnSelect func(*ListboxEvent)
	Decoration
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `ButtonOptions` → [button.md](button.md), `ChipOptions` → [chip.md](chip.md), `Color` → [theme.md](theme.md), `ListboxOptions` → [listbox.md](listbox.md), `Optional` → [theme.md](theme.md), `PopoverOptions` → [popover.md](popover.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Listbox, SelectMenu, and async search

##### SelectMenu: searchable popup

SelectMenu combines a trigger with a search field and a Listbox inside a popup. SelectMenuState stores Open, Query, the Listbox state, the trigger Handle, and the search field's QueryHandle. Size and Color apply to all three areas. SearchTerm can bind an external query pointer; assigning it from outside does not fire OnQuery or OnSearchTermChange.

By default, both ↑ and ↓ open the menu from the trigger and focus the search field. With FocusSearch=false, SearchInput=false, or a disabled search field, the list receives focus instead. In the search field, Enter activates the highlighted item directly, ↑↓ move into the list, and Escape closes the menu and returns to the trigger. Single selection closes the menu; multiple selection keeps it open. If your selection callback moves focus deliberately, that target is kept.

The default filtering is the same as Listbox. SearchInput only controls whether the search field is shown; when hidden, the existing query still applies. When ExternalFilter or IgnoreFilter is true, the results your app provides are used as is. FilterInput configures the search field. SearchPlaceholder remains as a shortcut, but FilterInput.Placeholder takes precedence. The inherited Filter field has no effect on SelectMenu.

Leading and Avatar configure the trigger's leading content, Trailing sets the trailing arrow, and Highlight emphasizes the border. Loading shows a spinner using LoadingIcon, and LoadingTrailing controls its position. The native loading status row is also kept, and existing results can still be selected. Error and EmptyText set the status text and the no-results message. ShowTrailing=false hides the whole trailing area.

| Option | Behavior |
| --- | --- |
| ResetSearchTermOnSelect / ResetSearchTermOnBlur | Both on by default. Clear the search term after a selection and after closing, respectively |
| Clearable / ClearOptions / ClearIcon | When something is selected, replaces the trailing arrow with a configurable clear button. Hidden when ShowTrailing=false |
| ResetModelValueOnClear | Default true. Controls whether the clear button resets the selected value. The button always clears the query, and deleting search text by itself never resets the selection |
| OnClear / OnOpenChange | Fire once per user clear or actual open/close change. Changing State.Open directly does not report an open/close change |
| OnCreate | Shows a create row that can be clicked or activated from the keyboard. Only sends a create request; it does not add or select anything automatically |
| CreateWhenAlways / CreateAtTop | By default, the create row appears only when nothing matches. These options show it whenever no item has the same ID, and change its position |

A create request keeps the query and the open state. OnCreate receives the query with leading and trailing whitespace trimmed and does not change the options or the selection. On the create row, OnHighlight returns the query text instead of exposing an internal row ID. OnChange fires only when the value changes; OnSelect also fires when the selected value is activated again. Clearing or resetting the query fires OnQuery and OnSearchTermChange only if the query actually changed.

With nothing selected, the trigger shows placeholder text in Dimmed. A Subtle trigger's background darkens on hover. The search field in the popup uses the None variant with no leading icon by default and has a separator below it. The list has no border, rows are inset by 4 DIP, and the default maximum height is 240 DIP, shrinking when there is little content. If you customize Build or part heights, set Height explicitly; the automatic estimate does not inspect arbitrary custom layouts. Root, Trigger, Leading, Label, and Trailing style only the trigger. The list uses separate parts such as List, Header, ItemLeadingSlot, ItemLabelSlot, and ItemTrailingSlot, and the panel uses Popup.UI.

##### Caching and large data sets

Listbox caches a snapshot of the data, an ID index, the groups, and the selection set. These structures are not rebuilt when the input has not changed, and in-place changes to the slice are also detected. However, each build still compares the snapshot linearly, and prefix search may scan every item.

SelectMenu caches filter results and selected labels; options that are temporarily missing keep their last label. Processing a large number of selected items for the first time still takes linear work. Virtual lists reduce the amount of UI built, but the cost of processing data still depends on its size.

##### QueryState: reject stale search results

QueryState[T] can serve as the app data source for SelectMenu or InputMenu. Your app sends requests; QueryState handles cancellation tokens, the loading state, and checking whether a result is still valid.

```go
// Start on the UI thread; queryState is stored per window.
ctx, ticket := queryState.Start(parentContext, query)
go func() {
    items, err := fetchItems(ctx, ticket.Query)
    window.Update(func() {
        queryState.Complete(ticket, items, err)
    })
}()
```

`fetchItems` is your app's request function. In the next build, pass queryState's Items, Loading, and Error to the menu. Remote search is usually combined with ExternalFilter.

Start cancels the previous context. Complete clones the result slice and rejects tickets that are stale or come from another instance. Call `queryState.Cancel()` when leaving the view or closing the window. The library creates no network tasks and does not prescribe debounce timing, fuzzy matching, or remote paging.

Multi-select lists provide a selection role and selected-row semantics, but MyGo does not yet expose a separate multiselectable flag, so the full screen reader experience still needs verification.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| SelectMenu | Root, Trigger, Leading, Label, Trailing, Close, Editor, List, Header (group heading), SeparatorSlot, Item, ItemLeadingSlot, Body, ItemLabelSlot, Description, ItemTrailingSlot, Indicator, EmptySlot. The panel uses Popup.UI |

#### Examples (excerpt)

Text input in InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, and ColorPicker is affected by the known IME cancel issue.

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| SelectMenu | SearchInput=true. The search term is reset both after a selection and after closing |

#### Selection and pagination › SelectMenu search and reset

`SearchInput` defaults to true. Setting it to false hides the search field, but any retained Query still filters the list. ExternalFilter or IgnoreFilter skips local filtering. ↑ or ↓ on the trigger opens the menu and focuses the search field if there is one; with FocusSearch=false, the list gets focus instead. A disabled trigger cannot open the menu.

By default, Query is cleared after a selection and after closing. `ResetSearchTermOnSelect` and `ResetSearchTermOnBlur` control these two cases. Closing includes Escape, clicking outside, and the app changing the open state.

Selecting the current value again still closes the menu, resets the query, and fires OnSelect. OnChange fires only when the selected value actually changes. Invalid and ErrorMessage update both the field's error semantics and its border.

Clearable shows a clear button, configured with ClearOptions and ClearIcon. The button always clears the query. ResetModelValueOnClear defaults to true; set it to false to keep the selected value. OnClear fires once after the reset. OnOpenChange reports only actual open/close changes caused by the user.

Listbox configures built-in filtering with Filter and FilterInput; SelectMenu uses SearchInput and FilterInput. Both support a SearchTerm pointer, FilterFields, IgnoreFilter, and OnSearchTermChange. When SearchTerm is unset, each uses its own State.Query. Assigning the term from outside does not fire a search notification.

### Inputs, selection, and fields

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

Disabled is passed down to the items. Size comes from the instance first, then the ComponentFieldGroup default, then the enclosing FormField; if none is set, each child uses its own default size. Popups opened by child controls keep their own corner radius, such as SelectMenu's search field and InputDate's calendar.

#### Select and InputMenu › Select: single selection and empty values (excerpt)

The parts are Root, Label, Leading, Trailing, Panel, Item, and EmptySlot; SelectMultiple maps the Root style to the trigger. Option.Build can compose row content such as avatars and chips, and must not keep build-time elements. The current model has no label or separator item types and no nested groups. Popup has no arrow, no DOM portal, and no upstream item-aligned positioning. Single selection does not use Popup.Mode, OpenDelay, or CloseDelay; multiple selection keeps SelectMenu's native hover extension.

#### Select and InputMenu › Option row appearance (excerpt)

Select, InputMenu, SelectMenu, Listbox, menus, and CommandPalette share the same option row style:

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

#### Feature scope › Data and files (excerpt)

| Area | Supported | Not provided, or left to the app |
| --- | --- | --- |
| Listbox / SelectMenu | Stable string IDs, structural rows, avatars and dots, filtering and sorting, per-item cancellable actions, native virtual lists | Does not accept arbitrary Vue objects or dynamic property paths. Search uses Unicode case folding and diacritic normalization, which is not full Intl locale collation. Virtualization does not expose overscan or estimateSize |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type example struct {
	Values                                                            []string
	Menu                                                              nui.SelectMenuState
	Changes, Selections, Queries, Attempts, Highlights, Opens, Clears int
	Reversed                                                          bool

	Results nui.QueryState[nui.ListboxItem]

	Status string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	items := []nui.ListboxItem{
		{ID: "languages", Label: "Languages", Kind: nui.ListboxLabel},
		{ID: "go", Label: "Go", Description: "Native applications", Icon: lucide.Code()},
		{ID: "rust", Label: "Rust", Description: "Systems programming", Chip: &nui.ChipOptions{Color: nui.Warning}},
		{ID: "typescript", Label: "TypeScript", Description: "Web applications", Avatar: &nui.AvatarOptions{Text: "TS", Alt: "TypeScript"}},
		{ID: "separator", Kind: nui.ListboxSeparator},
		{ID: "prefix", Label: "Cafeteria", Description: "Prefix match"},
		{ID: "contains", Label: "A café", Description: "Contains the query"},
		{ID: "exact", Label: "Café", Description: "Exact match; activation can be cancelled", Options: &nui.ListboxItemOptions{OnSelect: func(e *nui.ListboxEvent) {
			s.Attempts++

			e.PreventDefault()
		}, Decoration: nui.Decoration{UI: nui.Styles{nui.ItemLabelSlot: {Weight: nui.Set(600)}}}}},
	}
	if s.Reversed {
		items[1], items[3] = items[3], items[1]
	}
	nui.SelectMenu(c.Key("select"), t, &s.Values, &s.Menu, items, nui.SelectMenuOptions{ListboxOptions: nui.ListboxOptions{Label: "Languages", FilterInput: &nui.InputOptions{Label: "Search languages", Placeholder: "Search languages or cafe"}, OnChange: func([]string) { s.Changes++ }, OnSelect: func(string) { s.Selections++ }, OnSearchTermChange: func(string) { s.Queries++ }, OnHighlight: func(string) { s.Highlights++ }}, Placeholder: "Type something…", Clearable: true, Avatar: &nui.AvatarOptions{Text: "UI", Alt: "Language choice"}, OnOpenChange: func(bool) { s.Opens++ }, OnClear: func() { s.Clears++ }})
	nui.Button(c.Key("reorder"), t, "Reorder options", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { s.Reversed = !s.Reversed })
	ui.Textf(c, "Selected: %v · Changes: %d · Activations: %d", s.Values, s.Changes, s.Selections)
	ui.Textf(c, "Queries: %d · Highlights: %d · Attempts: %d", s.Queries, s.Highlights, s.Attempts)

	ui.Textf(c, "Open changes: %d · Clears: %d", s.Opens, s.Clears)
}

func main() {
	s := &example{}

	s.Status = "No request"
	defer s.Results.Cancel()
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "SelectMenu · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
