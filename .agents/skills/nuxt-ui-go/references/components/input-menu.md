<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputMenu

Search in an input and select from its suggestions.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `InputMenu`, `SearchSelect`
- Types: `InputMenuOptions`
- Theme key: `nui.ComponentInputMenu` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

SearchState.Query stores the draft, Selected the single ID and SelectedValues the multiple IDs. An initial single value fills an empty draft; external selection updates refresh idle display while active drafts survive until closing or blur. Open is an optional persistent bool pointer.

## Things to know

OnChange reports draft edits, OnSelect item activation, OnValuesChange multiple-value changes and OnClear the clear action. Tag deletion does not emit OnSelect. Native filtering keeps input order and ignores case; upstream relevance sorting and locale-sensitive equivalence are not implemented. Changed item data closes the old list; cancellation and stale-result handling belong to the application. SearchSelect retains legacy click-to-open and label-plus-description filtering.

## API

```go
func InputMenu(c *ui.Context, t Theme, state *SearchState, options []Option[string], o InputMenuOptions) ui.Element

// SearchSelect is a convenience wrapper; use InputMenu for additional options.
func SearchSelect(c *ui.Context, t Theme, state *SearchState, options []Option[string], label string) ui.Element
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
	Mode                                           InputMenuMode
	OpenOnClick, OpenOnFocus                       Optional[bool]
	ResetSearchTermOnBlur, ResetSearchTermOnSelect Optional[bool]
	ResetModelValueOnClear                         Optional[bool]
	HighlightOnHover                               Optional[bool]
	Loop                                           bool
	Open                                           *bool
	SelectedIcon, DeleteIcon, ClearIcon            *ui.SVG
	OnValuesChange                                 func([]string)
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
	SelectedValues  []string // Used when InputMenuOptions.Multiple is true.
}
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

```go
type InputMenuMode uint8

const (
	InputMenuCombobox InputMenuMode = iota
	InputMenuAutocomplete
)
```

Types documented in other references: `InputOptions` → [input.md](input.md), `Mode` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `PopoverOptions` → [popover.md](popover.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Input, Textarea, Select, and InputMenu support these variants:

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

#### Checkbox, CheckboxGroup, RadioGroup, and Switch › Options and keyboard (excerpt)

Both group types use ChoiceGroupOptions, which adds Horizontal, Loop, and a parameterless OnChange. Groups are vertical by default; Horizontal lets items wrap, except with ChoiceTable. Each option's Label and Description provide its text. In checkbox groups, Option.Icon replaces the check icon; in both group types with the indicator hidden, it appears above the label. Option.Build has no effect in choice groups; it is only used by Select and InputMenu.

#### Select and InputMenu

##### Choose an interaction

| What you need | Use |
| --- | --- |
| Pick one item from a list after clicking | `Select`, bound to `*T` |
| Pick several items after clicking | `SelectMultiple`, bound to `*[]T` |
| Show suggestions while typing | `InputMenu`, bound to `*SearchState` |
| A trigger with a separate search field in the popup | [SelectMenu](select-menu.md) |

Select and InputMenu share `Option[T comparable]`, which has Value, Label, Description, Disabled, Icon, and Build. Value is the stable identity: it must be unique and equal to itself (so NaN is not allowed). Labels can repeat. Even with a custom Build, fill in Label for search and assistive technology.

Disabled items are still shown but cannot be reached with the keyboard. The list is virtualized and calls Build only for the visible rows plus a few extra, so do not run business logic that must see all the data inside a row builder. Comparing input data still costs more as the data grows; your app handles remote paging and async requests.

##### InputMenu: drafts versus selected values

SearchState.Query holds the text being edited and Selected holds the single selected ID. An empty Selected means nothing is selected, so InputMenu option IDs cannot be empty strings. In multiple selection, SelectedValues is the source of truth. Removing an option from the data source does not clean up the app's selection. For ordinary single selection, if the initial draft is empty, the selected name is shown. If Selected changes from outside while the user is not editing, the display updates; a draft being edited is kept until the menu closes or loses focus.

By default, the search is a case-insensitive substring match on Label. FilterFields changes which fields are searched; the older SearchSelect convenience function searches Label and Description. The local filter result is recomputed only when the data, the query, or the filter fields change.

InputMenuOptions embeds InputOptions and adds options such as EmptyText, Popup, AllowCustom, and OnSelect. By default, clicking only focuses the input; typing, arrow keys, or the trailing arrow open the panel. Set OpenOnClick=true to open on click. The input keeps focus. Trailing can change the icon. Loading does not disable editing or selection. Avatar, LoadingTrailing, and Highlight work as in Input.

Clearable shows a clear button whenever there is a draft or a selection, so you can clear all tags in multiple selection even with an empty draft. ResetModelValueOnClear defaults to true; an explicit false clears only the draft and keeps the selection. Deleting ordinary text does not clear the selected ID. SelectedIcon, DeleteIcon, and ClearIcon replace the check mark, the X for deleting a tag, and the X for clearing. Autocomplete has no selection mark and no multi-select tags. After a tag is deleted with the keyboard, focus returns to the input.

Open takes a persistent bool pointer; nil uses internal state. Opening it from outside focuses the input, and becoming disabled or read-only closes it. Loop defaults to false and HighlightOnHover to true; both skip disabled items. Popup.Dismissible=false prevents closing with Escape or when focus moves outside, though setting Open to false still closes it. Popup's hover mode and delays do not apply to InputMenu.

| Action | Callbacks and state |
| --- | --- |
| Editing the draft | OnChange reports the text change |
| Choosing a suggestion | Calls OnSelect once and does not also submit free text |
| Pressing Enter with AllowCustom=true and no suggestion chosen | Calls InputOptions.OnSubmit(Query). Does not set Selected and does not fire OnSelect |
| Clear button | Clears Query, clears the selection depending on ResetModelValueOnClear, and calls OnClear. OnValuesChange fires if the multiple selection actually changed |
| Removing in multiple selection | Click a tag's × or press Backspace on an empty draft. Selection changes are reported through OnValuesChange |

By default, only suggestions can be selected and free text is not submitted. In multiple selection, values appear as small tags. Only the × button removes a tag, and the × is still in the Tab order. There is currently no arrow-key navigation between tags.

When nothing matches, or all suggestions are disabled, EmptyText is shown; the default messages distinguish "no data" from "no matching data". Up and Down can move to suggestions that are scrolled out of view, and focus stays in the input after a selection. Escape or blur closes the panel. The empty state cannot be selected, does not take focus, and does not intercept outside clicks.

When an option's ID, Label, Description, or Disabled changes, the open panel closes, and it uses the new data when reopened. This prevents accepting a stale suggestion. The component does not cancel network requests itself. For multiple selection, creating options, autocomplete, and reset policies, see InputMenu values and search.

InputMenu supports Input's four parts plus Panel, Item, and EmptySlot. Icons in rows with a description align with the first line. The empty message appears only when nothing matches; suggestions that are all disabled are still shown normally. The inherited Style receives only the input and its icons; style the panel with Popup.Decoration.Style. New code should use InputMenu with its full Options directly. SearchSelect keeps the older behavior of filtering by name and description and opening on click.

##### Option row appearance

Select, InputMenu, SelectMenu, Listbox, menus, and CommandPalette share the same option row style:

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| Padding and gap (DIP) | 4 | 6 | 6 | 8 | 8 |
| Font size / line height (DIP) | 12/16 | 12/16 | 14/20 | 14/20 | 16/24 |
| Icon size (DIP) | 16 | 16 | 20 | 20 | 24 |

Rows have a 6 DIP corner radius. A row highlighted by keyboard or pointer uses Elevated at 50% as its background and Highlighted for its text, and its leading icon changes from Dimmed to Text. Selected items are marked with a check mark at the end of the row, without an extra background.

The description appears on the line below the label, at the same font size, in the Muted color. Group headings are semibold and one size smaller than items. The empty state is centered and uses the Muted color.

#### Popover and ScrollArea › Popover: show content next to a control (excerpt)

By default, a generic Popover has the theme background, a 1 DIP border, a 6 DIP corner radius, an 8 DIP content gap, a shadow, and 16 DIP of padding. To make content reach the edges, set the padding of UI[Panel] to 0. Select, InputMenu, and menus use the same base style but usually with a 240 DIP maximum height, 4 DIP of padding, no gap, and at least the trigger's width. For the menus' own height options, see [DropdownMenu and ContextMenu](dropdown-menu.md).

### Options and defaults

#### InputMenu values and search

##### Choose what to store

| Mode | State fields |
| --- | --- |
| Single selection | `SearchState.Selected` stores the selected option ID; Query stores the edit draft |
| Multiple selection | `SelectedValues` stores every selected ID. Remove items with their delete buttons or with Backspace on an empty draft |
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
        queryState.Complete(ticket, items, err)
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

	nui.InputMenu(c.Key("menu"), t, &s.Search, []nui.Option[string]{{Value: "go", Label: "Go", Description: "Backend services", Icon: lucide.Code()}, {Value: "rust", Label: "Rust", Description: "Native applications", Icon: lucide.Code()}, {Value: "blocked", Label: "Disabled language", Disabled: true}, {Value: "typescript", Label: "TypeScript", Description: "Web applications", Icon: lucide.Code()}}, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Project name", Placeholder: "Type something…", LoadingTrailing: true, Avatar: avatar, Clearable: true, OnChange: func(string) { s.Queries++ }, OnClear: func() { s.Clears++ }, OnSubmit: func(string) { s.Submits++ }}, Mode: nui.InputMenuCombobox, Open: &s.Open, SelectedIcon: selectedIcon, DeleteIcon: deleteIcon, ClearIcon: clearIcon, OnSelect: func(string) { s.Selections++ }, OnValuesChange: func([]string) { s.Values++ }})
	ui.Textf(c, "Query: %s · Selected: %s · IDs: %v", s.Search.Query, s.Search.Selected, s.Search.SelectedValues)
	ui.Textf(c, "Events: query %d · select %d · values %d · clear %d · submit %d", s.Queries, s.Selections, s.Values, s.Clears, s.Submits)
	ui.Textf(c, "Open: %v · Create request: %s", s.Open, s.Created)
	nui.Button(c.Key("toggle"), t, "Toggle search list", nui.ButtonOptions{Variant: nui.Outline, Disabled: disabled || readOnly}).OnClick(func() { s.Open = !s.Open })
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
