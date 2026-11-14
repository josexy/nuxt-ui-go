<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Select

Choose one or more values from a dropdown list.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Select`, `SelectMultiple`
- Types: `SelectOptions`
- Theme key: `nui.ComponentSelect` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Bind one value with Select or a slice with SelectMultiple. Option.Value is the stable ID, Label is display text and Description adds detail. Open points to a persistent bool writable by both the component and application; omit it for internal state.

## Things to know

Empty strings and zero values can be valid option values. Use Optional[T] or an explicit business value for absence. A missing single value shows the placeholder; multiple selections retain labels for temporarily missing items and omit empty labels from joined text. External assignments, appearance changes and opening do not call OnChange. Closing the popup restores focus to the current trigger. If the highlighted item is removed, the list moves to a valid selected item or the first available item.

## API

```go
// Select never reserves a T zero value. Optional[T] or an explicit option models absence.
func Select[T comparable](c *ui.Context, t Theme, selected *T, options []Option[T], o SelectOptions) ui.Element

// SelectMultiple is Select's slice-valued counterpart. It reuses the searchable
// list implementation with search disabled, preserving typed option identities.
func SelectMultiple[T comparable](c *ui.Context, t Theme, value *[]T, options []Option[T], o SelectOptions) ui.Element
```

### SelectOptions

```go
type SelectOptions struct {
	Label, Placeholder, ErrorMessage string
	Disabled, Invalid                bool
	Color                            Color
	Variant                          Variant
	Size                             Size
	Leading, Trailing                *ui.SVG
	Avatar                           *AvatarOptions
	Loading, LoadingTrailing         bool
	LoadingIcon, SelectedIcon        *ui.SVG
	Highlight                        bool
	Open                             *bool // Optional persistent state; nil keeps opening internal.
	Popup                            PopoverOptions
	OnChange                         func()
	Decoration
}
```

### Related types

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

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `PopoverOptions` → [popover.md](popover.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Input, Textarea, Select, and InputMenu support these variants:

#### State, themes, and events › Focus and callbacks (excerpt)

Text inputs show an accent border and focus ring when focused by mouse or Tab, and return to normal on blur. Read-only fields can still be focused, selected, and copied, and an error border keeps its error color. Triggers such as Select keep their keyboard focus indicator. The hover background of Outline and Subtle does not apply when disabled.

An open Select panel inherits the disabled state of its trigger and ancestors; once disabled, the selection cannot be changed. The current MyGo version may keep the panel visible until it is closed.

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

##### Select: single selection and empty values

Select reserves no special zero value. To represent "nothing selected" explicitly, wrap the business value in an Optional:

```go
// app.Region has type nui.Optional[string].
nui.Select(c, theme, &app.Region, []nui.Option[nui.Optional[string]]{
    {Value: nui.Optional[string]{}, Label: "No region"},
    {Value: nui.Set("ap"), Label: "Asia Pacific", Description: "Singapore"},
    {Value: nui.Set("us"), Label: "United States", Disabled: true},
}, nui.SelectOptions{
    Label: "Region",
    Placeholder: "Choose a region",
}).FillWidth()
```

This example lists the Optional zero value as an option, so the select initially shows "No region". A value with no matching option is kept as is and Placeholder is shown; Placeholder defaults to an empty string. With no options at all, the built-in "No data" message is shown.

SelectOptions includes Label, Placeholder, ErrorMessage, Disabled, Invalid, Color, Variant, Size, Popup, OnChange, and Decoration. Leading and Trailing set the trigger icons; Trailing defaults to ChevronDown. Avatar is shown when there is no leading icon. SelectedIcon replaces the check mark on the selected item, not the item's own Icon.

Loading only shows a spinner; it does not disable selection. By default, LoaderCircle appears at the start; LoadingTrailing=true moves it to the end, but an explicit Leading keeps it at the start. LoadingIcon replaces the spinner icon, which uses the corresponding part's foreground color. If Avatar.Size is unset, the avatar diameter is 16/16/20/20/24 DIP by select size; an explicit avatar size takes precedence.

Highlight keeps the colored border when the select is not focused, including with the None variant; the error color takes precedence. With Neutral, the border and keyboard focus ring use InvertedBackground.

Open is an optional persistent `*bool`. Without it, Select uses internal state. With it, app assignments, trigger actions, and closing all update the same value, and focus returns to the trigger after closing. Opening the select, or changing the selection from outside, does not fire OnChange; only a user action that changes the selection fires it, once. Popup.Dismissible defaults to true; an explicit false prevents closing with Escape or an outside click, though setting Open to false still closes it.

When the select is open, use the arrow keys, Home, End, or type a prefix to move the highlight; press Enter or Space to select and Escape to close. Repeating a letter cycles through matches and skips disabled items. When a single select is closed, typing a prefix changes the selection directly, as a native convenience. In multiple selection, prefix navigation works inside the open list. Reordering keeps the highlight by ID. If the highlighted item is removed, the highlight moves to the current valid selection or the first available item. Multiple selection shows the selected names joined by commas, skipping empty labels; selected items that are temporarily missing keep their previous names.

The parts are Root, Label, Leading, Trailing, Panel, Item, and EmptySlot; SelectMultiple maps the Root style to the trigger. Option.Build can compose row content such as avatars and chips, and must not keep build-time elements. The current model has no label or separator item types and no nested groups. Popup has no arrow, no DOM portal, and no upstream item-aligned positioning. Single selection does not use Popup.Mode, OpenDelay, or CloseDelay; multiple selection keeps SelectMenu's native hover extension.

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

#### Selection and pagination › Single selection, multiple selection, and reselecting (excerpt)

| Component or option | Behavior |
| --- | --- |
| `Select(*T)` / `SelectMultiple(*[]T)` | Bind a single value or a slice of values. Both share SelectOptions, stable IDs, custom Option.Build, and selected-item tags |

Single and multiple Select both inherit Color, Variant, Size, and UI from ComponentSelect; instance options take precedence. Listbox also supports Loop and HighlightOnHover.

### Forms, dates, and data components

#### Listbox, SelectMenu, and async search › Listbox: a selectable list shown inline (excerpt)

The empty message and the loading indicator are polite status regions with stable identities, as are the empty messages of Select, InputMenu, SelectMenu, and filtered DropdownMenu and ContextMenu, and SelectMenu's loading and error rows. A status is not read when it first appears. If its text changes while it stays shown, for example from "No data" to "No matching data" as the user types, the new text is read once; rebuilding without a change reads nothing.

#### Calendar, InputDate, and InputTime › Calendar: pick a date by clicking (excerpt)

Dates and ranges are always complete Gregorian dates. There is no discrete multi-date model and no separate whole-month or whole-year value model. The month and year navigation views do not implement Nuxt UI's `type="month"` or `type="year"`; manage those values with app controls such as Select or InputNumber. Other calendars, time zones, and browser DOM slots are outside this native model.

### Overlays, menus, and navigation

#### Modal and Slideover › Closing and focus (excerpt)

Tab focus is trapped inside the modal content, and an inner Select or Popover handles Escape first. By default, the first focusable control receives focus, usually the close button. To focus a field first, request that field's Handle in the action that opens the dialog. If the Header or Footer is very tall, your app must constrain the content size to leave room for the body.

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
	Value                string
	Many                 []string
	Open, ManyOpen       bool
	Changes, ManyChanges int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var avatar *nui.AvatarOptions

	avatar = &nui.AvatarOptions{Text: "GO", Alt: "Language avatar"}

	items := []nui.Option[string]{{Value: "go", Label: "Go", Description: "Backend services", Icon: lucide.Code()}, {Value: "rust", Label: "Rust", Description: "Native applications", Icon: lucide.Code()}, {Value: "blocked", Label: "Disabled language", Disabled: true}, {Value: "typescript", Label: "TypeScript", Description: "Web applications", Icon: lucide.Code()}}
	nui.Select(c.Key("select"), t, &s.Value, items, nui.SelectOptions{Label: "Language", Placeholder: "Choose a language", Avatar: avatar, LoadingTrailing: true, Open: &s.Open, OnChange: func() { s.Changes++ }})
	ui.Textf(c, "Selected: %s · Changes: %d · Open: %v", s.Value, s.Changes, s.Open)
	nui.SelectMultiple(c.Key("multiple"), t, &s.Many, items, nui.SelectOptions{Label: "Multiple languages", Placeholder: "Choose a language", Avatar: avatar, LoadingTrailing: true, Open: &s.ManyOpen, OnChange: func() { s.ManyChanges++ }})
	ui.Textf(c, "Selected IDs: %v · Changes: %d · Open: %v", s.Many, s.ManyChanges, s.ManyOpen)
	nui.Button(c.Key("open-one"), t, "Toggle single select", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Open = !s.Open })
	nui.Button(c.Key("open-many"), t, "Toggle multiple select", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.ManyOpen = !s.ManyOpen })
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Select · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
