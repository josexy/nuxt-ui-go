<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Listbox

Filter and select one or more items in an always-visible list.

- Category: Forms and data
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `Listbox`
- Types: `ListboxOptions`, `ListboxEvent`
- Theme key: `nui.ComponentListbox` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Store selected IDs in `[]string` and keep query, highlight, and scroll state in `ListboxState`. Every item—including labels and separators—requires a unique `ID`. `Filter` adds a built-in search box; `Loading` shows a spinner and pauses item actions.

#### Large lists

Items passed to `Listbox` are compared with the last ones in every build, so that in-place edits show; that takes about a millisecond a build per 100,000 items. For a list that large, call `ListboxState.SetItems(items)` when the items change, outside the view, and pass `nil` items: builds then compare no items, however many there are. Edits to the slice show after the next `SetItems`, and `ReleaseItems` returns to passing items in each build.

## Things to know

- Built-in filtering matches `Label` by default (ignoring case and accents) and ranks exact, prefix, then substring matches.
- Call `PreventDefault` inside an item's `OnSelect` to cancel selection.
- Hiding the filter keeps an existing query active; set `IgnoreFilter: true` to bypass local filtering.
- `SetItems` copies and validates the items once; while it is in effect, pass `nil` items (anything else panics). Do not call it on the `ListboxState` inside a SelectMenu or CommandPalette.

## API

```go
func Listbox(c *ui.Context, t Theme, selected *[]string, s *ListboxState, items []ListboxItem, o ListboxOptions) ui.Element
```

### ListboxOptions

```go
type ListboxOptions struct {
	// Build draws non-interactive row content synchronously; Label still supplies semantics.
	Build              func(*ui.Context, Theme, ListboxItem)
	Label, EmptyText   string
	Multiple, Disabled bool
	Color              Color
	Size               Size
	Highlight          bool
	Loading            bool
	LoadingIcon        *ui.SVG
	SelectedIcon       *ui.SVG
	Filter             bool
	FilterInput        *InputOptions
	FilterFields       []string // label (default), description, value/id, group.
	IgnoreFilter       bool
	SearchTerm         *string // nil uses ListboxState.Query.
	OnSearchTermChange func(string)
	Height             float32
	Orientation        Optional[Orientation] // Defaults to Vertical.
	ItemWidth          float32               // Horizontal item width, default 160 DIP.
	Loop               bool
	HighlightOnHover   Optional[bool]
	SelectionBehavior  SelectionBehavior
	OnChange           func([]string)
	OnSelect           func(string) // Activation, including selecting an unchanged value.
	OnHighlight        func(string) // A user action changed the highlighted item.
	Decoration
}
```

### ListboxEvent

```go
// ListboxEvent describes the item before activation. PreventDefault cancels
// selection and the component's OnSelect notification for that activation.
type ListboxEvent struct {
	ID       string
	Selected bool
	// contains filtered or unexported fields
}

func (e *ListboxEvent) PreventDefault()
```

### Related types

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
type SelectionBehavior uint8

const (
	SelectionToggle SelectionBehavior = iota
	SelectionReplace
)
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

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `ChipOptions` → [chip.md](chip.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `InputOptions` → [input.md](input.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Lifecycle and themes (excerpt)

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. InputNumber, InputTags, PinInput cells, the SelectMenu trigger, Calendar cells and card-variant CheckboxGroup items fade their background, border and focus ring over `Motion.Transition` (150 ms) at the Full and Reduced motion levels, as Nuxt UI's `transition-colors`; text colours change at once, and the Off motion level shows every change at once. InputDate and InputTime fields do not fade. The rows of Listbox, Tree and the SelectMenu list change their highlight at once at every motion level, as MyGo's own list rows do: in MyGo a colour fade builds the whole view in each frame it moves, and the pointer going down a long list crosses a row every few frames. Other components are configured through their own Options at the end of this page.

#### Listbox, SelectMenu, and async search

##### Listbox: a selectable list shown inline

Each ListboxItem's ID must be non-empty and unique. Label and Description provide the text, and Group adds a group heading to adjacent items. Kind can be ListboxOption, ListboxLabel, or ListboxSeparator; structural rows also need IDs. The selected value is a `*[]string`, where an empty slice means nothing is selected. ListboxState stores the query, the active item, scroll state, and a persistent QueryHandle, and for long lists the items themselves (see Caching and large data sets).

The default is SelectionToggle: in both single and multiple selection, activating a selected item again deselects it. SelectionReplace replaces the selection on activation, and in multiple selection it also supports Shift range selection. Selected IDs that are temporarily missing from the data source are kept, which helps with async search; your app cleans up IDs that are gone for good.

Icon, Avatar, and Chip are checked in that order for the leading content, which aligns with the first line of text. SelectedIcon overrides the default check icon. Per-item options provide UI, Style, and OnSelect. The callback receives the ID and the selected state before the change, and can call PreventDefault to cancel this selection and the component's OnSelect. For the row appearance, see [Option row appearance](select.md). An MD row is about 32 DIP tall, plus one line when it has a description.

Size and Color default to MD and Primary. Instance options take precedence over the component theme, and the size can also be inherited from FormField. Highlight keeps a colored border visible. Keyboard focus shows a focus ring, which uses InvertedBackground with Neutral. When embedded in SelectMenu, the list follows the menu's size and color.

Filter defaults to false. When enabled, configure the search input through FilterInput. SearchTerm takes a persistent string pointer; when unset, State.Query is used. FilterFields defaults to matching label only and also supports description, value/id, and group; an explicit empty list matches nothing for a non-empty query. Filtering ignores case and combining diacritical marks. Within each run of options between groups and structural rows, results are ordered by exact match, then prefix match, then substring match. Full locale collation rules are not implemented.

With the filter field hidden, a retained query still applies; IgnoreFilter skips local filtering. Loading replaces the options with an icon and pauses list interaction, while the filter field stays editable. LoadingIcon overrides the icon. The spinner keeps turning under reduced motion and stops at the Off motion level.

The empty message and the loading indicator are polite status regions with stable identities, as are the empty messages of Select, InputMenu, SelectMenu, and filtered DropdownMenu and ContextMenu, and SelectMenu's loading and error rows. A status is not read when it first appears. If its text changes while it stays shown, for example from "No data" to "No matching data" as the user types, the new text is read once; rebuilding without a change reads nothing.

| Orientation or key | Behavior |
| --- | --- |
| Vertical | A native virtual list; ↑↓ move |
| Horizontal | Items are shown and clipped at a fixed ItemWidth; ←→ move. Group headings also take one item width |
| Home/End, Enter/Space | Go to the first or last item; activate the item |
| Typing characters | Finds items by prefix |
| Enter in the filter field | Keeps focus in the input and activates the highlighted item; during IME composition, the IME handles it |
| Reordered data / outside selection changes | Reordering keeps stable IDs. On initialization or when the selection changes from outside, the first available selected item is highlighted, or else the first available item |

In both orientations, only the visible items plus a few extra are built. Height is the total height including the filter field; when unset, it is 224 DIP vertically and 64 DIP horizontally. ItemWidth defaults to 160 DIP in horizontal layout. A synchronous Build can customize non-interactive row content. Loop defaults to false and HighlightOnHover to true. OnHighlight reports only highlight changes caused by the user, not initialization or outside value updates.

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
| Listbox | Root, Editor, LoadingSlot, LoadingIconSlot, List, Label (group heading), SeparatorSlot, Item, Leading, Body, ItemLabelSlot, Description, Trailing, Indicator, EmptySlot |

#### Examples (excerpt)

Text input in InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, and ColorPicker is affected by the known IME cancel issue.

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| Listbox, Accordion | Listbox: HighlightOnHover=true. Accordion: Collapsible=true |

#### Selection and pagination › Single selection, multiple selection, and reselecting (excerpt)

Single and multiple Select both inherit Color, Variant, Size, and UI from ComponentSelect; instance options take precedence. Listbox also supports Loop and HighlightOnHover.

#### Selection and pagination › SelectMenu search and reset (excerpt)

Listbox configures built-in filtering with Filter and FilterInput; SelectMenu uses SearchInput and FilterInput. Both support a SearchTerm pointer, FilterFields, IgnoreFilter, and OnSearchTermChange. When SearchTerm is unset, each uses its own State.Query. Assigning the term from outside does not fire a search notification.

#### Limitations (excerpt)

Listbox builds only the visible items plus a few extra in either orientation. Horizontal items and group headers use a fixed ItemWidth; in vertical layout, the native List manages row heights.

### Themes and basic components

#### Components and parts › Indicators, separators, and loading (excerpt)

Separator's `Size` maps XS–XL to line widths of 1–5 DIP. Painting rounds the width to whole device pixels before aligning its position, so equal widths remain equal at fractional DPI and UI zoom. Dashed and dotted segments use whole-pixel lengths and gaps; the last segment may be shorter. Card section dividers and default menu/Listbox separators paint one device pixel while preserving DIP layout spacing. `Type` is one of the Solid, Dashed, and Dotted constants, and `Position` is Center, Start, or End. Content precedence is `Content → Label → Icon → Avatar`. The default label is 14 DIP, the icon 20 DIP, and the avatar 20 DIP; an explicit Avatar.Size takes precedence. Leading refers to the icon or avatar root. Vertical labels can wrap. A vertical separator is 48 DIP tall by default, which you can change on the returned element; give long labels and interactive content enough height. The content container keeps its identity when the position or orientation changes, but controls inside it still need stable keys.

#### Components and parts › Progress and animation (excerpt)

**When animations stop:** Continuous animations request the next frame only while a visible component actually draws. They stop when the component scrolls out of view, unmounts, has Animate turned off, or at the Off motion level; spinners and the Skeleton pulse go on under reduced motion. Button's colour transition follows `Motion.Transition` at Full and Reduced and is immediate at Off, and `Animate: nui.Set(false)` controls the loading spinner. Once a transition ends, no more frames are requested. Indeterminate Progress and Table's loading bar pulse under reduced motion. In MyGo a colour transition builds the view in each frame while the colour moves (about ten builds for 150 ms), so an app with an expensive view can set `Motion.Transition` to 0 or turn animation off. The rows of the virtual lists (Listbox, Select, SelectMenu, InputMenu, CommandPalette and Tree) never fade, so that moving the pointer down a long list does not build the view in every frame; Tree paints its row hover without building the view at all. Components do not use timer goroutines for animation.

### Inputs, selection, and fields

#### Select and InputMenu › Option row appearance (excerpt)

Select, InputMenu, SelectMenu, Listbox, menus, and CommandPalette share the same option row style:

Rows have a 6 DIP corner radius. A row highlighted by keyboard or pointer uses Elevated at 50% as its background and Highlighted for its text, and its leading icon changes from Dimmed to Text. In the virtual lists of Select, InputMenu, SelectMenu, Listbox and CommandPalette the highlight moves at once at every motion level, unlike Nuxt UI's `transition-colors`, as MyGo's own list rows do: in MyGo a colour fade builds the whole view in each frame it moves, and the pointer going down a long list crosses a row every few frames. A held arrow key highlights, scrolls and reports the same rows at every motion level. The rows of DropdownMenu and ContextMenu, which are built whole, fade their background over `Motion.Transition`, also under reduced motion, so moving the highlight cross-fades the two rows; their text colour changes at once. Selected items are marked with a check mark at the end of the row, without an extra background.

#### Popover and ScrollArea › ScrollArea: give content a scrolling region (excerpt)

ScrollArea supports nesting and two-way scrolling. It does not provide upstream's items/virtualize, edge fade shadows, infinite loading, or custom scrollbar interaction. All content is built every time, so use the virtual lists in Listbox or Table for long lists. Put long content in a Popover inside a ScrollArea yourself.

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
	Values                                             []string
	List                                               nui.ListboxState
	Changes, Selections, Queries, Attempts, Highlights int
	Reversed                                           bool
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
	nui.Listbox(c.Key("list"), t, &s.Values, &s.List, items, nui.ListboxOptions{Label: "Languages", Filter: true, FilterInput: &nui.InputOptions{Label: "Filter languages", Placeholder: "Search languages or cafe"}, SelectedIcon: lucide.Check(), Height: 320, OnChange: func([]string) { s.Changes++ }, OnSelect: func(string) { s.Selections++ }, OnSearchTermChange: func(string) { s.Queries++ }, OnHighlight: func(string) { s.Highlights++ }})
	nui.Button(c.Key("reorder"), t, "Reorder options", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { s.Reversed = !s.Reversed })
	ui.Textf(c, "Selected: %v · Changes: %d · Activations: %d", s.Values, s.Changes, s.Selections)
	ui.Textf(c, "Queries: %d · Highlights: %d · Attempts: %d", s.Queries, s.Highlights, s.Attempts)
}

func main() {
	s := &example{}

	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Listbox · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
