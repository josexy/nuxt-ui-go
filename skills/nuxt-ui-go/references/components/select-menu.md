<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# SelectMenu

Search and select one or more options inside a popup list.

- Category: Forms and data
- Version: nuxt-ui-go v0.2.0 (MyGo v0.10.0)
- Functions: `SelectMenu`
- Types: `SelectMenuOptions`
- Theme key: `nui.ComponentSelectMenu` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

`SelectMenu` combines a trigger, search field, and `Listbox`. Store selected IDs in a `SelectionState[string]` and keep open state, query, and focus handles in `SelectMenuState`. Customize the search input via `FilterInput` and initial focus via `FocusSearch`.

#### Submitted data

Call SetItems during setup or when the data changes, not in the view; it copies the items, so call it again after editing them. Passing nil clears the items but keeps the selection and query. OnChange takes no arguments and fires once per user change; call Values() there when you need the selected IDs. Selection changes made from code are silent.

#### Local search execution

Set Local search execution to background and type into the filter to watch a background search. SearchAuto, the default, searches in the background above 2,048 items. While a search runs, the list shows a loading message and the query stays editable, but Enter and clicks are ignored so that a stale result cannot be chosen. A newer query, closing, or unmounting cancels the search; call Close() when you dispose of the state yourself.

## Things to know

- Selection and closing clear the search query by default. The clear button always clears the query; `ResetModelValueOnClear` controls whether selections clear too.
- `Loading` shows a trigger spinner and status row while existing items stay selectable.
- For remote search, use `ExternalFilter`/`IgnoreFilter` with `QueryState` to drop stale responses, and publish results on the UI thread via `Window.Update`.

## API

```go
func SelectMenu(c *ui.Context, t Theme, value *SelectionState[string], s *SelectMenuState, o SelectMenuOptions) ui.Element
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
// SelectionState owns an ordered, unique selection. Its zero value is empty.
// Keep it in persistent application state and do not copy it after use. While
// mounted, mutate it on the UI thread. Methods never emit component callbacks.
// Values makes an explicit O(S) snapshot; Len, Has and Revision do not copy it.
type SelectionState[K comparable] struct {
	// contains filtered or unexported fields
}

// Add appends a missing key, returning whether the selection changed.
func (s *SelectionState[K]) Add(key K) bool

// Clear releases the selection's storage and reports an effective change.
func (s *SelectionState[K]) Clear() bool

// Has reports membership without enumerating selected values.
func (s *SelectionState[K]) Has(key K) bool

func (s *SelectionState[K]) Len() int

// Remove preserves the order of surviving keys, returning whether it changed
// the selection. Moving the remaining ordered keys can take O(S) work.
func (s *SelectionState[K]) Remove(key K) bool

func (s *SelectionState[K]) Revision() uint64

// Set owns a snapshot, deduplicating in first-occurrence order. It returns true
// and advances Revision only when the ordered selection changes. Invalid keys
// panic before replacing the previous selection. Nil and empty both clear it.
func (s *SelectionState[K]) Set(values []K) bool

// Values returns an owned ordered snapshot. Mutating it cannot change s.
func (s *SelectionState[K]) Values() []K
```

```go
type SelectMenuState struct {
	Open        bool
	Query       string
	Handle      ui.Handle
	QueryHandle ui.Handle
	Listbox     ListboxState
	// contains filtered or unexported fields
}

// Close dismisses the popup and cancels local search/focus requests without
// changing submitted data, the draft or selection.
func (s *SelectMenuState) Close()

func (s *SelectMenuState) Searching() bool

// SetItems submits once to both the trigger and embedded list. They share the
// immutable owned snapshot; no cloning setter is needed during a view build.
func (s *SelectMenuState) SetItems(items []ListboxItem)
```

```go
type ListboxState struct {
	Active      string
	Query       string
	QueryHandle ui.Handle
	List        ui.ListState
	// contains filtered or unexported fields
}

// Close cancels local search and outstanding focus requests, retaining the
// submitted data, query and app-owned selection. A later mounted view can search
// again. Owners may call Close during disposal; unmount also cancels work.
func (s *ListboxState) Close()

// Searching reports whether local matching is pending. Application-owned
// remote loading remains the caller's responsibility.
func (s *ListboxState) Searching() bool

// SetItems owns item descriptors, nested avatar/chip/options and style maps.
// IDs (including structural rows) must be unique and nonempty. Invalid input
// leaves the previous dataset intact. Nil/empty clears data and filter caches.
func (s *ListboxState) SetItems(items []ListboxItem)
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

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `ButtonOptions` → [button.md](button.md), `Color` → [theme.md](theme.md), `ListboxOptions` → [listbox.md](listbox.md), `Optional` → [theme.md](theme.md), `PopoverOptions` → [popover.md](popover.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Lifecycle and themes (excerpt)

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. InputNumber, InputTags, PinInput cells, the SelectMenu trigger, Listbox and Tree rows, Calendar cells and card-variant CheckboxGroup items fade their background, border and focus ring over `Motion.Transition` (150 ms) at the Full and Reduced motion levels, as Nuxt UI's `transition-colors`; text colours change at once, and the Off motion level shows every change at once. InputDate and InputTime fields do not fade. Other components are configured through their own Options at the end of this page.

#### Listbox, SelectMenu, and async search

##### SelectMenu: searchable popup

SelectMenu combines a trigger with a search field and a Listbox inside a popup. Submit items with `state.SetItems(items)` and call `SelectMenu(c, theme, &selected, &state, options)`. SelectMenuState stores Open, Query, the Listbox state, the trigger Handle, and the search field's QueryHandle. Size and Color apply to all three areas. SearchTerm can bind an external query pointer; assigning it from outside does not fire OnQuery or OnSearchTermChange.

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

SetItems prepares the ID index, groups, and row keys once, so a build that changes nothing does not walk the items or the selection, and only visible rows are built. Submitting, filtering, Select All, prefix search, and `Values()` still take time proportional to the data or selection size. SelectMenu also caches its filter results and the trigger's joined labels.

OnChange takes no arguments and fires once per user change to the selection; changes made from code are silent. OnSelect reports activation, including activating the item that is already selected. An item's PreventDefault cancels both.

##### QueryState: reject stale search results

QueryState[T] can serve as the app data source for SelectMenu or InputMenu. Your app sends requests; QueryState handles cancellation tokens, the loading state, and checking whether a result is still valid.

```go
// Start on the UI thread; queryState is stored per window.
ctx, ticket := queryState.Start(parentContext, query)
go func() {
    items, err := fetchItems(ctx, ticket.Query)
    window.Update(func() {
        if queryState.Complete(ticket, items, err) {
            menuState.SetItems(queryState.Items)
        }
    })
}()
```

`fetchItems` is your app's request function. In the view, pass queryState's Loading and Error to the menu. Remote search is usually combined with ExternalFilter.

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

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Input, Textarea, InputNumber, InputTags, PinInput cells, and the Select, SelectMenu and InputMenu triggers fade their background, border and focus ring over `Motion.Transition` (150 ms) when hover, focus or invalid state changes, as Nuxt UI's `transition-colors`. The fade also runs under reduced motion; `Motion.Transition = 0` or the Off motion level shows the change at once. Text and icon colours change at once.

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

Disabled is passed down to the items. Size comes from the instance first, then the ComponentFieldGroup default, then the enclosing FormField; if none is set, each child uses its own default size. Popups opened by child controls keep their own corner radius, such as SelectMenu's search field and InputDate's calendar.

#### Select and InputMenu › Submitted data and SelectionState (excerpt)

| Component | Submit data with | Selection |
| --- | --- | --- |
| Listbox, SelectMenu | `SetItems` on their state | `*SelectionState[string]` |

#### Select and InputMenu › Search in large lists (excerpt)

Listbox, SelectMenu, InputMenu, DropdownMenu, and ContextMenu filter their submitted data locally. SearchExecution chooses where that runs:

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
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Values                                                            nui.SelectionState[string]
	Menu                                                              nui.SelectMenuState
	Changes, Selections, Queries, Attempts, Highlights, Opens, Clears int
	Reversed                                                          bool
	Remote                                                            nui.SelectMenuState

	Results nui.QueryState[nui.ListboxItem]

	Status  string
	Items   []nui.ListboxItem
	Protect bool
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	s.Protect = true
	nui.SelectMenu(c.Key("select"), t, &s.Values, &s.Menu, nui.SelectMenuOptions{ListboxOptions: nui.ListboxOptions{SearchExecution: nui.SearchAuto, Label: "Languages", FilterInput: &nui.InputOptions{Label: "Search languages", Placeholder: "Search languages or cafe"}, OnChange: func() { s.Changes++ }, OnSelect: func(string) { s.Selections++ }, OnSearchTermChange: func(string) { s.Queries++ }, OnHighlight: func(string) { s.Highlights++ }}, Placeholder: "Type something…", Clearable: true, Avatar: &nui.AvatarOptions{Text: "UI", Alt: "Language choice"}, OnOpenChange: func(bool) { s.Opens++ }, OnClear: func() { s.Clears++ }})
	nui.Button(c.Key("reorder"), t, "Reorder options", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() {
		s.Reversed = !s.Reversed
		if len(s.Items) > 3 {
			s.Items[1], s.Items[3] = s.Items[3], s.Items[1]
		}
		s.Menu.SetItems(s.Items)
	})
	ui.Textf(c, "Selected: %v · Changes: %d · Activations: %d", s.Values.Values(), s.Changes, s.Selections)
	ui.Textf(c, "Queries: %d · Highlights: %d · Attempts: %d", s.Queries, s.Highlights, s.Attempts)

	ui.Textf(c, "Open changes: %d · Clears: %d", s.Opens, s.Clears)
}

func main() {
	s := &example{}

	s.Status = "No request"
	s.Items = []nui.ListboxItem{
		{ID: "languages", Label: "Languages", Kind: nui.ListboxLabel},
		{ID: "go", Label: "Go", Description: "Native applications", Icon: lucide.Code()},
		{ID: "rust", Label: "Rust", Description: "Systems programming", Chip: &nui.ChipOptions{Color: nui.Warning}},
		{ID: "typescript", Label: "TypeScript", Description: "Web applications", Avatar: &nui.AvatarOptions{Text: "TS", Alt: "TypeScript"}},
		{ID: "separator", Kind: nui.ListboxSeparator},
		{ID: "prefix", Label: "Cafeteria", Description: "Prefix match"},
		{ID: "contains", Label: "A café", Description: "Contains the query"},
		{ID: "exact", Label: "Café", Description: "Exact match; activation can be cancelled", Options: &nui.ListboxItemOptions{OnSelect: func(e *nui.ListboxEvent) {
			s.Attempts++
			if s.Protect {
				e.PreventDefault()
			}
		}, Decoration: nui.Decoration{UI: nui.Styles{nui.ItemLabelSlot: {Weight: nui.Set(600)}}}}},
	}
	if s.Reversed && len(s.Items) > 3 {
		s.Items[1], s.Items[3] = s.Items[3], s.Items[1]
	}
	s.Menu.SetItems(s.Items)
	defer func() { s.Results.Cancel(); s.Menu.Close(); s.Remote.Close() }()
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
