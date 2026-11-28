<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputMenu

Type to filter suggestions inside an input and choose or submit a value.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `InputMenu`, `SearchSelect`
- Types: `InputMenuOptions`
- Theme key: `nui.ComponentInputMenu` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

`SearchState.Query` holds the input draft, `Selected` / `SelectedValues` store single or multiple IDs, and `Open` optionally controls the popup. Supports both combobox (`InputMenuCombobox`) and autocomplete (`InputMenuAutocomplete`) modes.

#### Submitted options and selected tags

Keep one SearchState per control and call SetOptions during setup or when the options change; the view passes only the state. SelectedValues is a SelectionState[string]. OnValuesChange takes no arguments, and changes made from code are silent.

Selected tags scroll inside an area up to 96 DIP tall, with the editor below. Tab, Shift+Tab, the arrow keys, Home, and End move between tags; removing a tag returns focus to the editor and keeps the draft. Read-only and disabled states prevent removal.

#### Local search execution

Set Local search execution to background and type into the filter to watch a background search. SearchAuto, the default, searches in the background above 2,048 items. While a search runs, the list shows a loading message and the query stays editable, but Enter and clicks are ignored so that a stale result cannot be chosen. A newer query, closing, or unmounting cancels the search; call Close() when you dispose of the state yourself.

## Things to know

- Callbacks: `OnChange` tracks draft edits, `OnSelect` item activation, `OnValuesChange` multi-select changes, and `OnClear` clear clicks (removing tags does not fire `OnSelect`).
- Built-in filtering is case-insensitive and preserves item order without relevance ranking. Updating options closes the old popup.
- Creating an item only notifies your callback; your app decides whether to add it. `SearchSelect` keeps legacy click-to-open behavior.

## API

```go
func InputMenu(c *ui.Context, t Theme, state *SearchState, o InputMenuOptions) ui.Element

// SearchSelect is a convenience wrapper; use InputMenu for additional options.
func SearchSelect(c *ui.Context, t Theme, state *SearchState, label string) ui.Element
```

### InputMenuOptions

```go
type InputMenuOptions struct {
	InputOptions
	EmptyText                                      string
	Popup                                          PopoverOptions
	AllowCustom                                    bool
	Multiple, ExternalFilter                       bool
	FilterFields                                   []string // label, description, value; nil uses label.
	SearchExecution                                SearchExecution
	Mode                                           InputMenuMode
	OpenOnClick, OpenOnFocus                       Optional[bool]
	ResetSearchTermOnBlur, ResetSearchTermOnSelect Optional[bool]
	ResetModelValueOnClear                         Optional[bool]
	HighlightOnHover                               Optional[bool]
	Loop                                           bool
	Open                                           *bool
	SelectedIcon, DeleteIcon, ClearIcon            *ui.SVG
	OnValuesChange                                 func()
	OnCreate                                       func(string)
	CreateWhenAlways                               bool
	CreateAtTop                                    bool
	OnSelect                                       func(string)
}
```

### Related types

```go
// SearchState separates a draft query from its last committed stable ID. Empty Selected means no choice.
type SearchState struct {
	Query, Selected string
	SelectedValues  SelectionState[string] // Used when InputMenuOptions.Multiple is true.
	Open            bool
}

// Close dismisses InputMenu and cancels search/activation/focus requests. When
// an external Open pointer is bound, its owner must also set that binding false.
// Submitted data, the draft and selection remain available on remount.
func (s *SearchState) Close()

func (s *SearchState) Searching() bool

// SetOptions owns an InputMenu dataset with unique, nonempty string IDs.
// Query, selection and opening are independent of the submitted descriptors.
func (s *SearchState) SetOptions(options []Option[string])
```

```go
// SearchExecution controls where local choice filtering runs. Empty queries
// and application-filtered data bypass local search under every policy.
type SearchExecution uint8

const (
	// SearchAuto runs normalization and matching in the background above
	// 2,048 submitted candidates, and immediately for smaller datasets.
	SearchAuto SearchExecution = iota
	SearchImmediate
	SearchBackground
)
```

```go
type InputMenuMode uint8

const (
	InputMenuCombobox InputMenuMode = iota
	InputMenuAutocomplete
)
```

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
// Option identity is independent of its label. Build is an optional synchronous content builder.
type Option[T comparable] struct {
	Value              T
	Label, Description string
	Disabled           bool
	Icon               *ui.SVG
	Build              func(*ui.Context, Theme)
}
```

Types documented in other references: `InputOptions` → [input.md](input.md), `Mode` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `PopoverOptions` → [popover.md](popover.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Input, Textarea, Select, and InputMenu support these variants:

Input, Textarea, InputNumber, InputTags, PinInput cells, and the Select, SelectMenu and InputMenu triggers fade their background, border and focus ring over `Motion.Transition` (150 ms) when hover, focus or invalid state changes, as Nuxt UI's `transition-colors`. The fade also runs under reduced motion; `Motion.Transition = 0` or the Off motion level shows the change at once. Text and icon colours change at once.

#### Input and Textarea › Editing and clearing (excerpt)

MyGo's TextInputBase and TextAreaBase handle text editing, selection, copy and paste, undo, and IME. The native editor applies MaxLength while it inserts text, before OnChange runs. Replacing a selection can use the freed space, and a partly accepted paste never splits a character. An insertion that is rejected entirely leaves the caret and undo history unchanged and fires no OnChange. Values assigned from outside are not truncated and do not fire input notifications. Changing the limit applies to the next input without resetting the draft or caret. InputTags and InputMenu use the same editor.

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

#### Checkbox, CheckboxGroup, RadioGroup, and Switch › Options and keyboard (excerpt)

Both group types use ChoiceGroupOptions, which adds Horizontal, Loop, and a parameterless OnChange. Groups are vertical by default; Horizontal lets items wrap, except with ChoiceTable. Each option's Label and Description provide its text. In checkbox groups, Option.Icon replaces the check icon; in both group types with the indicator hidden, it appears above the label. Option.Build has no effect in choice groups; it is only used by Select and InputMenu.

#### Select and InputMenu

##### Choose an interaction

| What you need | Use |
| --- | --- |
| Pick one item from a list after clicking | `Select`, bound to `*T` |
| Pick several items after clicking | `SelectMultiple`, bound to `*SelectionState[T]` |
| Show suggestions while typing | `InputMenu`, bound to `*SearchState` |
| A trigger with a separate search field in the popup | [SelectMenu](select-menu.md) |

Select and InputMenu share `Option[T comparable]`, which has Value, Label, Description, Disabled, Icon, and Build. Value is the stable identity: it must be unique and equal to itself (so NaN is not allowed). Labels can repeat. Even with a custom Build, fill in Label for search and assistive technology.

Disabled items are still shown but cannot be reached with the keyboard. The list is virtualized and calls Build only for the visible rows plus a few extra, so do not run business logic that must see all the data inside a row builder. Your app handles remote paging and network requests.

##### Submitted data and SelectionState

Choice controls read their data from a persistent state instead of taking a slice on every build:

| Component | Submit data with | Selection |
| --- | --- | --- |
| Select | `SelectState[T].SetOptions` | `*T` |
| SelectMultiple | `SelectState[T].SetOptions` | `*SelectionState[T]` |
| InputMenu, SearchSelect | `SearchState.SetOptions` | `SearchState.Selected`, or `SelectedValues` with Multiple |
| [Listbox, SelectMenu](listbox.md) | `SetItems` on their state | `*SelectionState[string]` |
| [DropdownMenu, ContextMenu](dropdown-menu.md) | `MenuState.SetItems` | None |

Call the setter during setup, or in the action or request completion that changes the data, never unconditionally in the view. The setter copies what you pass, so later edits to your slice have no effect until you call it again. Callbacks and icons are kept as references; they run later, so capture app state in them, never an Element or the build Context. Invalid data (a duplicate ID, a NaN value, or an empty string ID in the string-keyed components) panics and leaves the previous data in place. Passing nil clears the data but keeps the selection and the query. Keep one state per control and per window, and do not copy a state after it has been used.

`SelectionState[K]` is an ordered set of selected keys whose zero value is empty. Set, Add, Remove, and Clear return whether the selection changed; Has and Len read it without copying; Values returns a copy. Selected keys that are missing from the current data stay selected and keep their last known labels. Changing a selection from code never fires component callbacks. A user change fires the parameterless OnChange (OnValuesChange for InputMenu) once; call `Values()` there only when you need the keys.

```go
// Persistent app state: Status nui.SelectionState[int], StatusOptions nui.SelectState[int].
// In app setup, or an action that replaces the options:
app.StatusOptions.SetOptions([]nui.Option[int]{{Value: 0, Label: "Draft"}, {Value: 1, Label: "Published"}})

// In the view:
nui.SelectMultiple(c, theme, &app.Status, &app.StatusOptions, nui.SelectOptions{
    Label: "Status",
    OnChange: func() { app.SaveStatus(app.Status.Values()) },
})
```

##### Search in large lists

Listbox, SelectMenu, InputMenu, DropdownMenu, and ContextMenu filter their submitted data locally. SearchExecution chooses where that runs:

| Value | Behavior |
| --- | --- |
| `SearchAuto` (default) | In the background for more than 2,048 items, otherwise immediately |
| `SearchImmediate` | Always during the build, on the UI thread |
| `SearchBackground` | Always in the background |

While a background search runs, the list shows a loading message and the query stays editable; Enter, clicks, and creation are ignored so that a stale result cannot be chosen. A new query cancels the previous search. The state's `Searching()` reports pending work. Closing the popup, removing the control, or closing the window cancels the search; call the state's `Close()` when you dispose of it yourself, and set an external Open pointer to false. Empty queries, IgnoreFilter, and ExternalFilter skip local search. On the browser backend, background search runs on the JavaScript event loop in short slices, not in a Web Worker.

For FilterFields, nil means the default (label only), and an explicit empty list matches nothing for a non-empty query.

##### InputMenu: drafts versus selected values

SearchState.Query holds the text being edited and Selected holds the single selected ID. An empty Selected means nothing is selected, so InputMenu option IDs cannot be empty strings. Submit options with `SearchState.SetOptions` (see Submitted data and SelectionState), then call `InputMenu(c, theme, &state, options)` or `SearchSelect(c, theme, &state, label)`. In multiple selection, `SelectedValues` is a `SelectionState[string]`. Removing an option from the data does not remove it from the selection. For ordinary single selection, if the initial draft is empty, the selected name is shown. If Selected changes from outside while the user is not editing, the display updates; a draft being edited is kept until the menu closes or loses focus.

By default, the search is a case-insensitive substring match on Label. FilterFields changes which fields are searched; the older SearchSelect convenience function searches Label and Description. The local filter result is recomputed only when the data, the query, or the filter fields change.

InputMenuOptions embeds InputOptions and adds options such as EmptyText, Popup, AllowCustom, and OnSelect. By default, clicking only focuses the input; typing, arrow keys, or the trailing arrow open the panel. Set OpenOnClick=true to open on click. The input keeps focus. Trailing can change the icon. Loading does not disable editing or selection. Avatar, LoadingTrailing, and Highlight work as in Input.

In multiple selection, the selected values appear as tags above the editor. The tag area scrolls once it is taller than 96 DIP (scaled with the text size), and labels wrap in full. Tab, Shift+Tab, the arrow keys, Home, and End move between tags and scroll to them; Shift+Tab from the editor goes to the last tag. Enter, Space, Delete, or Backspace on a tag removes it and returns focus to the editor, and Backspace in an empty editor removes the last value. Moving between and removing tags keeps the draft. Disabled and read-only tags cannot be removed.

Clearable shows a clear button whenever there is a draft or a selection, so you can clear all tags in multiple selection even with an empty draft. ResetModelValueOnClear defaults to true; an explicit false clears only the draft and keeps the selection. Deleting ordinary text does not clear the selected ID. SelectedIcon, DeleteIcon, and ClearIcon replace the check mark, the X for deleting a tag, and the X for clearing. Autocomplete has no selection mark and no multi-select tags. After a tag is deleted with the keyboard, focus returns to the input.

Open takes a persistent bool pointer; nil uses internal state. Opening it from outside focuses the input, and becoming disabled or read-only closes it. Loop defaults to false and HighlightOnHover to true; both skip disabled items. Popup.Dismissible=false prevents closing with Escape or when focus moves outside, though setting Open to false still closes it. Popup's hover mode and delays do not apply to InputMenu.

| Action | Callbacks and state |
| --- | --- |
| Editing the draft | OnChange reports the text change |
| Choosing a suggestion | Calls OnSelect once and does not also submit free text |
| Pressing Enter with AllowCustom=true and no suggestion chosen | Calls InputOptions.OnSubmit(Query). Does not set Selected and does not fire OnSelect |
| Clear button | Clears Query, clears the selection depending on ResetModelValueOnClear, and calls OnClear. OnValuesChange fires if the multiple selection actually changed |
| Removing in multiple selection | Click a tag's ×, press Delete or Backspace on a tag, or press Backspace on an empty draft. OnValuesChange reports the change |

By default, only suggestions can be selected and free text is not submitted.

When nothing matches, or all suggestions are disabled, EmptyText is shown; the default messages distinguish "no data" from "no matching data". Up and Down can move to suggestions that are scrolled out of view, and focus stays in the input after a selection. Escape or blur closes the panel. The empty state cannot be selected, does not take focus, and does not intercept outside clicks.

Submitting new options while the panel is open closes it, so a stale suggestion cannot be accepted; it shows the new data when reopened. The component does not cancel network requests itself. For multiple selection, creating options, autocomplete, and reset policies, see InputMenu values and search.

InputMenu supports Input's four parts plus Panel, Item, and EmptySlot. Icons in rows with a description align with the first line. The empty message appears only when nothing matches; suggestions that are all disabled are still shown normally. The inherited Style receives only the input and its icons; style the panel with Popup.Decoration.Style. New code should use InputMenu with its full Options directly. SearchSelect keeps the older behavior of filtering by name and description and opening on click.

##### Option row appearance

Select, InputMenu, SelectMenu, Listbox, menus, and CommandPalette share the same option row style:

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| Padding and gap (DIP) | 4 | 6 | 6 | 8 | 8 |
| Font size / line height (DIP) | 12/16 | 12/16 | 14/20 | 14/20 | 16/24 |
| Icon size (DIP) | 16 | 16 | 20 | 20 | 24 |

Rows have a 6 DIP corner radius. A row highlighted by keyboard or pointer uses Elevated at 50% as its background and Highlighted for its text, and its leading icon changes from Dimmed to Text. The background fades in and out over `Motion.Transition`, also under reduced motion, so moving the highlight cross-fades the two rows; the text colour changes at once. In virtual lists the fade is on each row's content, so it never moves rows, and a held arrow key highlights, scrolls and reports the same rows as with animations off. Selected items are marked with a check mark at the end of the row, without an extra background.

The description appears on the line below the label, at the same font size, in the Muted color. Group headings are semibold and one size smaller than items. The empty state is centered and uses the Muted color.

#### Popover and ScrollArea › Popover: show content next to a control (excerpt)

By default, a generic Popover has the theme background, a 1 DIP border, a 6 DIP corner radius, an 8 DIP content gap, a shadow, and 16 DIP of padding. To make content reach the edges, set the padding of UI[Panel] to 0. Select, InputMenu, and menus use the same base style but usually with a 240 DIP maximum height, 4 DIP of padding, no gap, and at least the trigger's width. Their lists shrink with the panel and scroll, so in a short window every option can still be scrolled into view. For the menus' own height options, see [DropdownMenu and ContextMenu](dropdown-menu.md).

### Options and defaults

#### InputMenu values and search

##### Choose what to store

| Mode | State fields |
| --- | --- |
| Single selection | `SearchState.Selected` stores the selected option ID; Query stores the edit draft |
| Multiple selection | `SelectedValues`, a `SelectionState[string]`, stores the selected IDs, shown as tags above the editor. Remove a tag with its ×, with Delete or Backspace on the tag, or with Backspace on an empty draft |
| InputMenuAutocomplete | Query is free text. Selected keeps the last selected ID. Forces single selection and allows submitting custom text |

OnValuesChange reports changes to multiple selection, and OnSelect reports the ID that was just activated.

##### Local or external search

By default, the search is a case-insensitive substring match on Label. FilterFields can list `label`, `description`, and `value`; nil means label only. The older SearchSelect entry point searches label and description.

`ExternalFilter=true` shows the results the app passes in without filtering them. QueryState or an app service handles requests, cancellation, and stale results. A change to an option's identity, label, description, or disabled state closes the open panel; ExternalFilter does not change this rule and does not keep the panel open automatically.

##### Creating options

When you provide OnCreate, the list can show a create row that is reachable from the keyboard. By default it appears only when nothing matches; CreateWhenAlways shows it whenever the query is non-empty and no option has the same value.

Creating only sends a request: it keeps the query, filter, and open state, does not trigger an ordinary selection, and does not add the option for you. Update the data source after the app finishes creating the option.

##### Opening and resetting

OpenOnClick and OpenOnFocus both default to false. Typing, arrow keys, and the trailing button still open the menu. To open it on click, set `OpenOnClick: nui.Set(true)`. SearchSelect keeps its older open-on-click behavior. Open can bind a persistent bool. Loop defaults to false and HighlightOnHover to true.

ResetSearchTermOnSelect and ResetSearchTermOnBlur both default to true and control the draft after a selection and after closing. Single selection restores the selected label; multiple selection clears the draft. Autocomplete ignores Multiple, ResetSearchTermOnSelect, and ResetModelValueOnClear: selecting a suggestion always fills in its full text, and an ordinary blur keeps the free text.

ResetModelValueOnClear controls only the clear button and defaults to true; false keeps the selected value. It does not affect deleting ordinary text with the keyboard. In multiple selection, you can clear all tags from an empty draft or remove them one by one with the DeleteIcon buttons; each actual change to the selected values fires OnValuesChange once. After a tag is deleted with the keyboard, focus returns to the input.

### Forms, dates, and data components

#### Listbox, SelectMenu, and async search › Listbox: a selectable list shown inline (excerpt)

The empty message and the loading indicator are polite status regions with stable identities, as are the empty messages of Select, InputMenu, SelectMenu, and filtered DropdownMenu and ContextMenu, and SelectMenu's loading and error rows. A status is not read when it first appears. If its text changes while it stays shown, for example from "No data" to "No matching data" as the user types, the new text is read once; rebuilding without a change reads nothing.

#### Listbox, SelectMenu, and async search › QueryState: reject stale search results (excerpt)

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

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

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
	Search                                       nui.SearchState
	Open                                         bool
	Queries, Selections, Values, Clears, Submits int
	Created                                      string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	disabled, readOnly := false, false
	var avatar *nui.AvatarOptions

	avatar = &nui.AvatarOptions{Text: "GO", Alt: "Search avatar"}

	var selectedIcon, deleteIcon, clearIcon *ui.SVG

	selectedIcon = lucide.Heart()
	deleteIcon = lucide.Minus()
	clearIcon = lucide.Eraser()

	nui.InputMenu(c.Key("menu"), t, &s.Search, nui.InputMenuOptions{SearchExecution: nui.SearchAuto, InputOptions: nui.InputOptions{Label: "Project name", Placeholder: "Type something…", LoadingTrailing: true, Avatar: avatar, Clearable: true, OnChange: func(string) { s.Queries++ }, OnClear: func() { s.Clears++ }, OnSubmit: func(string) { s.Submits++ }}, Mode: nui.InputMenuCombobox, Open: &s.Open, SelectedIcon: selectedIcon, DeleteIcon: deleteIcon, ClearIcon: clearIcon, OnSelect: func(string) { s.Selections++ }, OnValuesChange: func() { s.Values++ }})
	ui.Textf(c, "Query: %s · Selected: %s · IDs: %v", s.Search.Query, s.Search.Selected, s.Search.SelectedValues.Values())
	ui.Textf(c, "Events: query %d · select %d · values %d · clear %d · submit %d", s.Queries, s.Selections, s.Values, s.Clears, s.Submits)
	ui.Textf(c, "Open: %v · Create request: %s", s.Open, s.Created)
	nui.Button(c.Key("toggle"), t, "Toggle search list", nui.ButtonOptions{Variant: nui.Outline, Disabled: disabled || readOnly}).OnClick(func() { s.Open = !s.Open })
}

func main() {
	s := &example{}

	s.Search.SetOptions([]nui.Option[string]{{Value: "go", Label: "Go", Description: "Backend services", Icon: lucide.Code()}, {Value: "rust", Label: "Rust", Description: "Native applications", Icon: lucide.Code()}, {Value: "blocked", Label: "Disabled language", Disabled: true}, {Value: "typescript", Label: "TypeScript", Description: "Web applications", Icon: lucide.Code()}})
	defer func() { s.Open = false; s.Search.Close() }()
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "InputMenu · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
