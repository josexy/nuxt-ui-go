<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# SelectMenu

Search and select one or more options inside a popup list.

- Category: Forms and data
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `SelectMenu`
- Types: `SelectMenuOptions`
- Theme key: `nui.ComponentSelectMenu` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

`SelectMenu` combines a trigger, search field, and `Listbox`. Store selected IDs in a slice and keep open state, query, and focus handles in `SelectMenuState`. Customize the search input via `FilterInput` and initial focus via `FocusSearch`.

#### Large lists

As with `Listbox`, items passed to `SelectMenu` are compared with the last ones in every build, open or closed; that takes about a millisecond a build per 100,000 items. For a list that large, call `SelectMenuState.SetItems(items)` when the items change, outside the view, and pass `nil` items: the trigger's label, the popup's rows and the filter then read that snapshot, and builds compare no items. Edits to the slice show after the next `SetItems`, and `ReleaseItems` returns to passing items in each build.

## Things to know

- Selection and closing clear the search query by default. The clear button always clears the query; `ResetModelValueOnClear` controls whether selections clear too.
- `Loading` shows a trigger spinner and status row while existing items stay selectable.
- For remote search, use `ExternalFilter`/`IgnoreFilter` with `QueryState` to drop stale responses, and publish results on the UI thread via `Window.Update`.
- `SelectMenuState.SetItems` copies and validates the items once; while it is in effect, pass `nil` items (anything else panics). Call it on `SelectMenuState`, not on the `Listbox` state it holds.

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

// ReleaseItems drops the items SetItems took: SelectMenu shows the items
// passed to it in each build again.
func (s *SelectMenuState) ReleaseItems()

// SetItems copies items and checks their IDs once, outside rendering, for
// SelectMenu to show in place of the items passed to it, which must then be
// nil, as ListboxState.SetItems does for Listbox: a build then compares no
// items, open or closed, however many there are.
//
// Call it again when the items change: edits to the slice after SetItems do
// not show. Items equal to the last ones keep the snapshot, the selected
// labels and the filter's results. A duplicate or empty ID, or an unknown
// Kind, panics and keeps the items shown before. Selected IDs missing from
// the items keep their last known label, as with items passed in each build.
// Call it on the window's UI thread, between builds, and not in every build.
// ReleaseItems returns to the items passed in each build.
func (s *SelectMenuState) SetItems(items []ListboxItem)
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

// ReleaseItems drops the items SetItems took: Listbox shows the items passed
// to it in each build again.
func (s *ListboxState) ReleaseItems()

// SetItems copies items and checks their IDs once, outside rendering, for
// Listbox to show in place of the items passed to it, which must then be nil.
// A build then compares no items, however many there are, where items passed
// in each build are compared with the last ones to see in-place edits.
//
// Call it again when the items change: edits to the slice after SetItems do
// not show. Items equal to the last ones keep the snapshot and the filter's
// results. A duplicate or empty ID, or an unknown Kind, panics and keeps the
// items shown before. Call it on the window's UI thread, between builds, and
// not in every build, where it compares as much as passing the items would.
// ReleaseItems returns to the items passed in each build.
//
// SelectMenu and CommandPalette set the items of their Listbox: do not call
// SetItems on the ListboxState they hold. SelectMenuState.SetItems gives a
// SelectMenu its items.
func (s *ListboxState) SetItems(items []ListboxItem)
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

#### Lifecycle and themes (excerpt)

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. InputNumber, InputTags, PinInput cells, the SelectMenu trigger, Calendar cells and card-variant CheckboxGroup items fade their background, border and focus ring over `Motion.Transition` (150 ms) at the Full and Reduced motion levels, as Nuxt UI's `transition-colors`; text colours change at once, and the Off motion level shows every change at once. InputDate and InputTime fields do not fade. The rows of Listbox, Tree and the SelectMenu list change their highlight at once at every motion level, as MyGo's own list rows do: in MyGo a colour fade builds the whole view in each frame it moves, and the pointer going down a long list crosses a row every few frames. Other components are configured through their own Options at the end of this page.

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

Listbox caches a snapshot of the data, an ID index, the groups, and the selection set. These structures are not rebuilt when the input has not changed, and in-place changes to the slice are also detected. However, each build still compares the items passed with the snapshot linearly, about a millisecond per 100,000 items, and prefix search may scan every item.

For long lists, give the items to the state instead, as Table's SetRows does. `ListboxState.SetItems(items)` copies the items and checks their IDs once, and Listbox, given nil items, shows that snapshot without comparing any items in its builds: with 100,000 items a build takes about as long as with 100. Call SetItems on the UI thread when the items change, not in every build; edits to the slice show only after the next call. Items equal to the last ones keep the snapshot and the filter's cached results, and an invalid snapshot (a duplicate or empty ID, an unknown Kind) panics in SetItems and leaves the items shown before. While the state holds items, passing non-nil items to Listbox panics; ReleaseItems returns to passing them in each build. SelectMenu and CommandPalette set the items of their list, so do not call SetItems on the ListboxState they hold; SelectMenu has a SetItems of its own.

```go
// When the data arrives, on the UI thread:
app.members.SetItems(items)

// In the view:
nui.Listbox(c, t, &app.selected, &app.members, nil, nui.ListboxOptions{Label: "Members"})
```

SelectMenu caches filter results and selected labels; options that are temporarily missing keep their last label. Processing a large number of selected items for the first time still takes linear work. `SelectMenuState.SetItems(items)` gives a SelectMenu its items in the same way: pass nil items, and the trigger's label, the popup's rows and its filter read the snapshot, with no items compared in a build, open or closed. ReleaseItems returns to passing them in each build, and passing items while the state holds its own panics. With the create entry shown (OnCreate), the list still copies and compares the filtered items in each build. Virtual lists reduce the amount of UI built, but the cost of processing data still depends on its size.

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

### Themes and basic components

#### Components and parts › Progress and animation (excerpt)

**When animations stop:** Continuous animations request the next frame only while a visible component actually draws. They stop when the component scrolls out of view, unmounts, has Animate turned off, or at the Off motion level; spinners and the Skeleton pulse go on under reduced motion. Button's colour transition follows `Motion.Transition` at Full and Reduced and is immediate at Off, and `Animate: nui.Set(false)` controls the loading spinner. Once a transition ends, no more frames are requested. Indeterminate Progress and Table's loading bar pulse under reduced motion. In MyGo a colour transition builds the view in each frame while the colour moves (about ten builds for 150 ms), so an app with an expensive view can set `Motion.Transition` to 0 or turn animation off. The rows of the virtual lists (Listbox, Select, SelectMenu, InputMenu, CommandPalette and Tree) never fade, so that moving the pointer down a long list does not build the view in every frame; Tree paints its row hover without building the view at all. Components do not use timer goroutines for animation.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Input, Textarea, InputNumber, InputTags, PinInput cells, and the Select, SelectMenu and InputMenu triggers fade their background, border and focus ring over `Motion.Transition` (150 ms) when hover, focus or invalid state changes, as Nuxt UI's `transition-colors`. The fade also runs under reduced motion; `Motion.Transition = 0` or the Off motion level shows the change at once. Text and icon colours change at once.

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

Disabled is passed down to the items. Size comes from the instance first, then the ComponentFieldGroup default, then the enclosing FormField; if none is set, each child uses its own default size. Popups opened by child controls keep their own corner radius, such as SelectMenu's search field and InputDate's calendar.

#### Select and InputMenu › Select: single selection and empty values (excerpt)

The parts are Root, Label, Leading, Trailing, Panel, Item, and EmptySlot; SelectMultiple maps the Root style to the trigger. Option.Build can compose row content such as avatars and chips, and must not keep build-time elements. The current model has no label or separator item types and no nested groups. Popup has no arrow, no DOM portal, and no upstream item-aligned positioning. Single selection does not use Popup.Mode, OpenDelay, or CloseDelay; multiple selection keeps SelectMenu's native hover extension.

#### Select and InputMenu › Option row appearance (excerpt)

Select, InputMenu, SelectMenu, Listbox, menus, and CommandPalette share the same option row style:

Rows have a 6 DIP corner radius. A row highlighted by keyboard or pointer uses Elevated at 50% as its background and Highlighted for its text, and its leading icon changes from Dimmed to Text. In the virtual lists of Select, InputMenu, SelectMenu, Listbox and CommandPalette the highlight moves at once at every motion level, unlike Nuxt UI's `transition-colors`, as MyGo's own list rows do: in MyGo a colour fade builds the whole view in each frame it moves, and the pointer going down a long list crosses a row every few frames. A held arrow key highlights, scrolls and reports the same rows at every motion level. The rows of DropdownMenu and ContextMenu, which are built whole, fade their background over `Motion.Transition`, also under reduced motion, so moving the highlight cross-fades the two rows; their text colour changes at once. Selected items are marked with a check mark at the end of the row, without an extra background.

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
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
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
