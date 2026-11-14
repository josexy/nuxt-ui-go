<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Checkbox

Toggle an option or represent a partially selected state.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Checkbox`, `CheckboxTri`
- Types: `CheckboxOptions`, `CheckboxTriOptions`
- Theme key: `nui.ComponentCheckbox` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Checkbox binds a `*bool`; CheckboxTri binds CheckState with Unchecked, Checked and Indeterminate values. CheckedIcon customizes the checked icon and IndeterminateIcon customizes the mixed icon. With a hidden indicator, an explicit CheckedIcon appears above the label; no default icon is inserted.

## Things to know

Keep selection values outside the view function. The application computes relationships between parent and child checkboxes. Loading is retained in the shared options for compatibility and is not used by Checkbox; it applies to Switch. ChoiceTable is a Go native extension for a standalone Checkbox.

## API

```go
func Checkbox(c *ui.Context, t Theme, value *bool, o CheckboxOptions) ui.Element

func CheckboxTri(c *ui.Context, t Theme, value *CheckState, o CheckboxTriOptions) ui.Element
```

### CheckboxOptions

```go
type CheckboxOptions = ToggleOptions
```

### CheckboxTriOptions

```go
type CheckboxTriOptions struct {
	CheckboxOptions
	// IndeterminateIcon replaces the minus in a visible mixed-state indicator.
	IndeterminateIcon *ui.SVG
	OnStateChange     func(CheckState)
}
```

### Related types

```go
type CheckState uint8

const (
	Unchecked CheckState = iota
	Checked
	Indeterminate
)
```

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

For a tri-state checkbox with Unchecked, Checked, and Indeterminate, use CheckboxTri.

##### Options and keyboard

Checkbox and Switch share ToggleOptions; CheckboxOptions and SwitchOptions are aliases of it. Common fields include Label, Description, ErrorMessage, Disabled, Invalid, Color, Size, OnChange, and Decoration.

A shared type does not mean every field applies to both controls. Checkbox does not use Loading, LoadingIcon, or UncheckedIcon; Switch does. RadioGroup always selects one item, so Max has no effect on it; CheckboxGroup's Max limits how many items can be selected.

ChoiceTable on a single Checkbox and Variant/Indicator on Switch are Go-native extensions, whereas ChoiceTable on CheckboxGroup and RadioGroup corresponds to an existing Nuxt UI group style. The native extensions remain available.

Both group types use ChoiceGroupOptions, which adds Horizontal, Loop, and a parameterless OnChange. Groups are vertical by default; Horizontal lets items wrap, except with ChoiceTable. Each option's Label and Description provide its text. In checkbox groups, Option.Icon replaces the check icon; in both group types with the indicator hidden, it appears above the label. Option.Build has no effect in choice groups; it is only used by Select and InputMenu.

Checkbox and Switch toggle on click or Space. Each group is a single Tab stop. Use Left/Right arrows in horizontal layout and Up/Down arrows in vertical layout; Home and End go to the first and last items, and disabled items are skipped. In CheckboxGroup, arrow keys only move focus and Space toggles the item; in RadioGroup, moving also selects. Loop is an Optional[bool]: it defaults to false for checkbox groups and true for radio groups, and an explicit `nui.Set(false)` stops wrapping. State updates immediately; there is currently no toggle animation.

A group's visible title can come from the enclosing FormField.Label. To use one icon for all checkboxes in a group, set the same SVG as Option.Icon on every option that has no icon of its own; items with their own icons keep them.

### Options and defaults

#### Common options by component › Switches, tags, and ratings (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Checkbox / CheckboxGroup / RadioGroup | Variant is ChoiceList, ChoiceCard, or ChoiceTable. Indicator is IndicatorStart, IndicatorEnd, or IndicatorHidden. Highlight emphasizes the border or indicator, even when unchecked |

Checkbox uses only CheckedIcon; it does not use Switch's UncheckedIcon. A ChoiceCard is tinted when selected, so the selected state stays visible even with the indicator hidden.

Checkbox's Loading, LoadingIcon, and UncheckedIcon, and RadioGroup's Max, come from shared option types and have no effect on those components; they remain for compatibility with existing code. ChoiceTable on a single Checkbox and Variant/Indicator on Switch are Go-native extensions.

### Forms, dates, and data components

#### Numbers, tags, PINs, and ratings › CheckboxTri

CheckboxTri uses a CheckState: Unchecked, Checked, or Indeterminate. The indeterminate state is reported to assistive technology as Mixed. The first activation changes it to Checked, and later activations toggle between Checked and Unchecked.

`IndeterminateIcon` sets the icon for the indeterminate state and defaults to a minus sign; `CheckedIcon` sets the ordinary checked icon. With the indicator hidden, only an explicit `CheckedIcon` is shown above the label; the indeterminate icon is not used and no default icon is added. `OnChange` reports whether the box is checked, and `OnStateChange` reports the full CheckState. Each effective action fires each callback once; changing an icon fires neither.

Your business logic computes parent-child check relationships. If you only need on/off, use a Checkbox bound to a `*bool`.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Value                 bool
	Mixed                 nui.CheckState
	Changes, MixedChanges int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.FormField(c.Key("field"), t, nui.FormFieldOptions{Label: "Notification permission", Help: "Applies to this workspace"}, func(field nui.Theme) {
		nui.Checkbox(c.Key("toggle"), field, &s.Value, nui.CheckboxOptions{Label: "Enable notifications", Description: "Updates are saved locally in this example.", Variant: nui.ChoiceList, Indicator: nui.IndicatorStart, OnChange: func(bool) { s.Changes++ }})
	})
	ui.Textf(c, "Enabled: %t", s.Value)
	nui.CheckboxTri(c.Key("tri"), t, &s.Mixed, nui.CheckboxTriOptions{CheckboxOptions: nui.CheckboxOptions{Label: "Mixed selection", Variant: nui.ChoiceList, Indicator: nui.IndicatorStart}, OnStateChange: func(nui.CheckState) { s.MixedChanges++ }})
	nui.Button(c.Key("reset-mixed"), t, "Reset mixed state", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Mixed = nui.Indeterminate })
	ui.Textf(c, "Changes: %d / %d", s.Changes, s.MixedChanges)
}

func main() {
	s := &example{}

	s.Mixed = nui.Indeterminate
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Checkbox · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
