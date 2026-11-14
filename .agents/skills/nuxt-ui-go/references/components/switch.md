<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Switch

Turn a setting on or off.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Switch`
- Types: `SwitchOptions`
- Theme key: `nui.ComponentSwitch` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Switch binds a non-nil `*bool`. CheckedIcon and UncheckedIcon customize each state. Loading displays LoadingIcon, using the default loader when unset, and blocks toggling. Completing loading does not change the value.

## Things to know

The application retains the selected value. Variant and Indicator are Go native extensions. The thumb updates immediately; the loader respects reduced motion. FormField provides the required marker, while form validators enforce submission rules.

## API

```go
func Switch(c *ui.Context, t Theme, value *bool, o SwitchOptions) ui.Element
```

### SwitchOptions

```go
type SwitchOptions = ToggleOptions
```

### Related types

```go
type ToggleOptions struct {
	Label, Description, ErrorMessage        string
	Disabled, Invalid                       bool
	Color                                   Color
	Size                                    Size
	Loading                                 bool
	LoadingIcon, CheckedIcon, UncheckedIcon *ui.SVG
	Variant                                 ChoiceVariant
	Indicator                               IndicatorPlacement
	Highlight                               bool
	OnChange                                func(bool)
	Decoration
	// contains filtered or unexported fields
}
```

```go
type ChoiceVariant uint8

const (
	ChoiceList ChoiceVariant = iota
	ChoiceCard
	ChoiceTable
)
```

```go
type IndicatorPlacement uint8

const (
	IndicatorStart IndicatorPlacement = iota
	IndicatorEnd
	IndicatorHidden
)
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### FormField and FieldGroup › FormField: labels, descriptions, and errors (excerpt)

Clicking the label activates the first associated control. Checkbox, Switch, and CheckboxGroup keep their toggle and notification behavior, and a tri-state Checkbox goes from indeterminate to checked. Use a separate FormField for each independent input. Your app must associate labels with custom MyGo controls itself.

#### Checkbox, CheckboxGroup, RadioGroup, and Switch

##### Choose a binding type

| Component | Bound value | Use for |
| --- | --- | --- |
| Checkbox / Switch | Non-nil `*bool` | A single on/off state |
| CheckboxGroup | Non-nil `*[]T`, where T is comparable | A set of items where several can be selected; a nil slice means nothing is selected |
| RadioGroup | Non-nil `*T` | A set of mutually exclusive items; for an empty selection, use Optional[T] or an explicit "none" option |

Option values must be unique, but display names can repeat. CheckboxGroup appends a value when it is checked and removes it when unchecked. Reordering does not change the selected values, and selected values that are temporarily missing from the options are kept; your app decides when to clean them up. RadioGroup does not change a missing value to the first item, and it reserves no special zero value.

For a tri-state checkbox with Unchecked, Checked, and Indeterminate, use [CheckboxTri](checkbox.md).

##### Options and keyboard

Checkbox and Switch share ToggleOptions; CheckboxOptions and SwitchOptions are aliases of it. Common fields include Label, Description, ErrorMessage, Disabled, Invalid, Color, Size, OnChange, and Decoration.

A shared type does not mean every field applies to both controls. Checkbox does not use Loading, LoadingIcon, or UncheckedIcon; Switch does. RadioGroup always selects one item, so Max has no effect on it; CheckboxGroup's Max limits how many items can be selected.

ChoiceTable on a single Checkbox and Variant/Indicator on Switch are Go-native extensions, whereas ChoiceTable on CheckboxGroup and RadioGroup corresponds to an existing Nuxt UI group style. The native extensions remain available.

Both group types use ChoiceGroupOptions, which adds Horizontal, Loop, and a parameterless OnChange. Groups are vertical by default; Horizontal lets items wrap, except with ChoiceTable. Each option's Label and Description provide its text. In checkbox groups, Option.Icon replaces the check icon; in both group types with the indicator hidden, it appears above the label. Option.Build has no effect in choice groups; it is only used by Select and InputMenu.

Checkbox and Switch toggle on click or Space. Each group is a single Tab stop. Use Left/Right arrows in horizontal layout and Up/Down arrows in vertical layout; Home and End go to the first and last items, and disabled items are skipped. In CheckboxGroup, arrow keys only move focus and Space toggles the item; in RadioGroup, moving also selects. Loop is an Optional[bool]: it defaults to false for checkbox groups and true for radio groups, and an explicit `nui.Set(false)` stops wrapping. State updates immediately; there is currently no toggle animation.

A group's visible title can come from the enclosing FormField.Label. To use one icon for all checkboxes in a group, set the same SVG as Option.Icon on every option that has no icon of its own; items with their own icons keep them.

### Options and defaults

#### Common options by component › Switches, tags, and ratings

| Component | Common options and behavior |
| --- | --- |
| Switch | Loading blocks toggling by mouse and keyboard. LoadingIcon, CheckedIcon, and UncheckedIcon set the icons. The spinner stops when reduced motion is on |
| Checkbox / CheckboxGroup / RadioGroup | Variant is ChoiceList, ChoiceCard, or ChoiceTable. Indicator is IndicatorStart, IndicatorEnd, or IndicatorHidden. Highlight emphasizes the border or indicator, even when unchecked |
| InputTags | AddOnTab, AddOnBlur, and AddOnPaste control when tags are added. Duplicate, ConvertValue, and DisplayValue control values and display. MaxLength limits the text being edited |
| PinInput | Color, Variant, and Size set the appearance. `Separator: []int{3}` adds a group break after the third cell; the control still uses one editor and one Tab stop |
| InputRating | Clearable lets clicking the current rating or pressing Space reset it to zero. Hoverable defaults to false. An unset EmptyIcon reuses Icon. Indicator, IconSlot, and EmptyIconSlot style the fill layer and the two icons |

Checkbox uses only CheckedIcon; it does not use Switch's UncheckedIcon. A ChoiceCard is tinted when selected, so the selected state stays visible even with the indicator hidden.

Checkbox's Loading, LoadingIcon, and UncheckedIcon, and RadioGroup's Max, come from shared option types and have no effect on those components; they remain for compatibility with existing code. ChoiceTable on a single Checkbox and Variant/Indicator on Switch are Go-native extensions.

Option.Icon works in CheckboxGroup and RadioGroup. With the indicator hidden, the icon appears above the item label; an ordinary Radio indicator is still a dot. Loop controls whether arrow keys wrap around; it defaults to false for checkbox groups and true for radio groups, and `nui.Set(false)` turns wrapping off. Both group types use a single Tab stop, with Left/Right arrows in horizontal layout and Up/Down arrows in vertical layout.

InputTags' MaxLength limits what the user can insert, counted in UTF-16 code units; 0 means no limit. It limits the draft being edited and does not validate tag length on submit. With AddOnPaste on, pasted text is always handled as tags; text rejected as a duplicate or over the limit is not inserted as plain text either. With it off, pasting inserts plain text. With AddOnTab on, Tab and Shift+Tab commit a non-empty draft and keep focus; with an empty draft they move focus as usual.

InputNumber's StepSnapping, FocusOnChange, Increment, and Decrement default to true, and the mouse wheel is enabled by default. The first increment or decrement from an empty value uses StartingValue instead of adding an extra step. PinInput.Highlight defaults to false; SeparatorSlot and Root.Gap/Padding adjust the separator and spacing.

InputRating's horizontal arrow keys follow RTL; in vertical layout, use ↑ and ↓, and the other axis's keys do nothing. The rating stops at both ends. Home and PageUp go to 1, End and PageDown go to Max. The hover preview disappears as soon as the rating is cleared.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Value   bool
	Changes int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.FormField(c.Key("field"), t, nui.FormFieldOptions{Label: "Sync policy", Help: "Applies to this workspace"}, func(field nui.Theme) {
		nui.Switch(c.Key("toggle"), field, &s.Value, nui.SwitchOptions{Label: "Enable notifications", Description: "Updates are saved locally in this example.", OnChange: func(bool) { s.Changes++ }})
	})
	ui.Textf(c, "Enabled: %t", s.Value)
	ui.Textf(c, "Changes: %d", s.Changes)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Switch · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
