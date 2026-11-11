<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# RadioGroup

Choose one item from a visible set of options.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `RadioGroup`
- Types: `RadioGroupOptions`
- Theme key: `nui.ComponentRadioGroup` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

RadioGroup binds a `*T` and selects one item. Option.Value must be unique; labels may repeat. Option.Icon customizes item icons, placing them above the label when the indicator is hidden. FormField.Label provides a visible group heading.

## Things to know

Max is retained in the shared options but is not used by RadioGroup. Visible indicators remain radio dots; item icons appear only when the indicator is hidden. Loop uses `Optional[bool]` and accepts explicit false. Option.Build is not used by choice groups.

## API

```go
func RadioGroup[T comparable](c *ui.Context, t Theme, value *T, options []Option[T], o RadioGroupOptions) ui.Element
```

### RadioGroupOptions

```go
type RadioGroupOptions = ChoiceGroupOptions
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

```go
type ChoiceGroupOptions struct {
	Label, Description, ErrorMessage string
	Disabled, Invalid, Horizontal    bool
	Color                            Color
	Size                             Size
	Max                              Optional[int]
	// Loop controls arrow-key wrapping; defaults to false for CheckboxGroup and true for RadioGroup.
	Loop      Optional[bool]
	Variant   ChoiceVariant
	Indicator IndicatorPlacement
	Highlight bool
	OnChange  func()
	Decoration
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

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

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

#### Common options by component › Switches, tags, and ratings (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Checkbox / CheckboxGroup / RadioGroup | Variant is ChoiceList, ChoiceCard, or ChoiceTable. Indicator is IndicatorStart, IndicatorEnd, or IndicatorHidden. Highlight emphasizes the border or indicator, even when unchecked |

Checkbox's Loading, LoadingIcon, and UncheckedIcon, and RadioGroup's Max, come from shared option types and have no effect on those components; they remain for compatibility with existing code. ChoiceTable on a single Checkbox and Variant/Indicator on Switch are Go-native extensions.

Option.Icon works in CheckboxGroup and RadioGroup. With the indicator hidden, the icon appears above the item label; an ordinary Radio indicator is still a dot. Loop controls whether arrow keys wrap around; it defaults to false for checkbox groups and true for radio groups, and `nui.Set(false)` turns wrapping off. Both group types use a single Tab stop, with Left/Right arrows in horizontal layout and Up/Down arrows in vertical layout.

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
	Value   string
	Reverse bool
	Changes int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	items := []nui.Option[string]{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}, {Value: "typescript", Label: "TypeScript"}}

	for i := range items {
		if items[i].Icon == nil {
			switch items[i].Value {
			case "go":
				items[i].Icon = lucide.Code()
			case "rust":
				items[i].Icon = lucide.Flame()
			default:
				items[i].Icon = lucide.FileCode()
			}
		}
	}

	if s.Reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}
	nui.FormField(c.Key("field"), t, nui.FormFieldOptions{Label: "Language preferences"}, func(field nui.Theme) {
		nui.RadioGroup(c.Key("choices"), field, &s.Value, items, nui.RadioGroupOptions{Label: "Languages", Variant: nui.ChoiceList, Indicator: nui.IndicatorStart, Loop: nui.Set(true), OnChange: func() { s.Changes++ }})
	})
	ui.Textf(c, "Selected: %v", s.Value)
	nui.Button(c.Key("reverse"), t, "Reverse choices", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Reverse = !s.Reverse })
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
		mygo.NewWindow(mygo.WindowOptions{Title: "RadioGroup · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
