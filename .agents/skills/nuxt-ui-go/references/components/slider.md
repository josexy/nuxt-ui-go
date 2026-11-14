<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Slider

Adjust a value or range by dragging sliders.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Slider`, `RangeSlider`
- Types: `SliderOptions`, `RangeSliderOptions`
- Theme key: `nui.ComponentSlider` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Slider binds one `*float64`; RangeSlider binds two pointers for the lower and upper values. Min, Max and Step set the limits and increment. Keep the values outside the view function.

## Things to know

This example uses a 0–100 range. OnChange reports live changes; use OnCommit to handle a completed adjustment. Disabling a parent cancels an active drag. Tooltip is a boolean switch rather than full TooltipOptions. Only the single slider uses UI.Root.PaddingX / PaddingY to set its active track area. Models with three or more interactive thumbs are not provided.

## API

```go
func Slider(c *ui.Context, t Theme, value *float64, o SliderOptions) ui.Element

// RangeSlider shares a pointer track and keeps two independent keyboard and
// accessibility targets. Public Range/Step semantics expose actual model values.
func RangeSlider(c *ui.Context, t Theme, low, high *float64, o RangeSliderOptions) ui.Element
```

### SliderOptions

```go
type SliderOptions struct {
	Label, Description string
	Min, Max           Optional[float64]
	Step               float64
	Color              Color
	Size               Size
	Disabled, Vertical bool
	Inverted, Tooltip  bool
	OnChange           func(float64)
	OnCommit           func(float64) // Release or a keyboard step; OnChange continues to report live changes.
	Decoration
}
```

### RangeSliderOptions

```go
type RangeSliderOptions struct {
	Label                 string
	Min, Max, Step        float64
	Disabled              bool
	Orientation           Orientation
	Inverted, Tooltip     bool
	MinStepsBetweenThumbs int
	Color                 Color
	Size                  Size
	OnChange              func(float64, float64)
	OnCommit              func(float64, float64)
	Decoration
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### Slider

Slider binds a non-nil `*float64`. By default, its range is 0–100, it is horizontal, and its size is MD. Min and Max are Optional[float64]; the bounds must be finite with Max > Min, and Step must be finite and non-negative.

An explicit Label is used as the name first, then the FormField label; if both are empty, the localized "Slider" name is used. The label is not drawn as visible text; Description adds a supplementary description.

Values set from outside are clamped to the range; NaN and Inf are treated as 0 before clamping. Slider has no empty-value model. Step=0 uses the default step of 1.

User actions snap to multiples of Step from Min, and floating-point noise is removed based on the decimal precision of Min and Step. For example, adding 0.1 to 0.2 shows 0.3. Disabling the slider or a parent cancels any drag in progress, and OnCommit does not fire.

| Action | Behavior |
| --- | --- |
| Arrow keys | Move one Step, following the orientation, Inverted, and RTL |
| PageUp/PageDown, Shift+arrow keys | Move 10 Steps |
| Home/End | Go to the bounds |
| OnChange | Reports value changes live |
| OnCommit | Fires when a key press or drag that changed the value ends |

The default height is 24 DIP horizontally and 160 DIP vertically; change it on the returned element. The track is 6/7/8/9/10 DIP thick for sizes XS to XL. The thumb is a dot in the background color with a 2 DIP border in the theme color. The keyboard focus ring is 3 DIP wide and 2 DIP away from the thumb. RangeSlider looks the same.

For Slider and RangeSlider, Root styles the whole control, Indicator.Background sets the filled track and the thumb border, and Thumb.Background sets the thumb fill. Neutral uses InvertedBackground. A FormField error uses the Error color first, although explicit instance part colors can still override it. The unfilled track uses Theme.Accented. The track and thumb are drawn by a draw callback, so other Thumb and Indicator style fields have no effect, and Style receives only Root.

Tooltip is a bool. When on, hovering or focusing the thumb shows the current value immediately, and Escape hides it. It has no full TooltipOptions or ComponentTooltip part configuration. On a single slider, changing UI.Root.PaddingX/PaddingY changes the active track area, and drawing, pointer math, and tooltip positioning all follow. RangeSlider keeps a fixed track inset of half a thumb.

For dual thumbs, inversion, value tooltips, and vertical assistive technology limits, see Sliders, tabs, and accordions.

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| InputTags, InputRating, Slider | InputTags: AddOnPaste=false. InputRating: Hoverable=false. Slider: Step=1 |

#### Common options by component › Sliders, tabs, and accordions

| Component | Common options and behavior |
| --- | --- |
| Slider / RangeSlider | PageUp/PageDown and Shift+arrow keys move by 10 steps. OnChange reports live changes; OnCommit fires when a key press or drag that changed the value ends |
| RangeSlider | The thumbs can cross; they are then reordered and active focus moves with them. Moves that would bring them closer than `Step × MinStepsBetweenThumbs` are rejected. The tooltip shows each value separately |
| Tabs | ActivationMode defaults to ActivationAutomatic. With ActivationManual, arrow keys only move focus, and Enter or Space activates a tab and fires a single notification |
| Accordion | Collapsible defaults to true. In single mode, setting it to false prevents collapsing the open item by clicking it again, although another item can still be opened. Multiple mode ignores this option |
| Collapsible / Accordion | Disabled disables only the header. KeepMounted defaults to false; true keeps hidden controls mounted. TrailingIcon defaults to ChevronDown, and an Accordion item's own icon takes precedence. AccordionItem.UI overrides that item's style |
| Carousel | Arrows shows the previous/next buttons and Dots enables dot navigation. The dots' page names follow the current language |

#### Limitations (excerpt)

Slider and RangeSlider expose their value, range, and step to assistive technology, but MyGo has no API for setting vertical orientation metadata. This affects every vertical mode, including sliders without Inverted. Slider tooltips can be positioned per thumb, but they have no separate delay setting. InputRating uses a single-focus slider role.

### Forms, dates, and data components

#### Numbers, tags, PINs, and ratings › RangeSlider

RangeSlider takes two `*float64` values. By default, the range is 0–100 and Step=1. If the two thumbs cross, they are reordered and active focus moves with them so that low ≤ high always holds. Moves that would violate the minimum distance are rejected. Out-of-range values passed by your app are clamped during the build.

Common options include Orientation, Inverted, Tooltip, MinStepsBetweenThumbs, Color, Size, OnChange, and OnCommit. Each thumb has its own focus and reports its own value and minimum/maximum names to assistive technology.

Indicator.Background and Thumb.Background work the same way as on a single Slider. Floating-point noise is removed from fractional steps, and the minimum-distance comparison uses a tolerance that scales with the distance; even a very small step cannot make the thumbs overlap and then push the other thumb along. Disabling a parent container cancels the current drag without a commit.

Inside a FormField, clicking the label focuses the first thumb, and the field's description and error are associated with it. For orientation and keyboard behavior, see Sliders, tabs, and accordions.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| RangeSlider | Root, plus Indicator.Background and Thumb.Background |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's header divider paints one device pixel and takes its stroke color from `UI[SeparatorSlot].Background` (default BorderAccented). Its replacement covers MyGo's one-DIP divider using the theme background or an opaque Root/Body/Panel UI background; custom translucent surfaces or raw Style background changes should provide their own SeparatorSlot material. Table's other internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](theme.md).

### Compatibility and known limitations

#### Accessibility and motion (excerpt)

| Area | Current limitation |
| --- | --- |
| Vertical Slider / RangeSlider | Orientation metadata is missing, including for non-inverted vertical sliders |

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
	Value, Low, High                             float64
	Changes, Commits, RangeChanges, RangeCommits int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	parts, singleParts := nui.Styles{}, nui.Styles{}

	orientation := nui.Horizontal

	ui.Column(c.Key("controls")).Disabled(false).Gap(16).Children(func() {
		nui.FormField(c.Key("single-field"), t, nui.FormFieldOptions{}, func(ft nui.Theme) {
			nui.Slider(c.Key("slider"), ft, &s.Value, nui.SliderOptions{Label: "Volume", Min: nui.Set(0.0), Max: nui.Set(100.0), Step: 5, Tooltip: true, OnChange: func(float64) { s.Changes++ }, OnCommit: func(float64) { s.Commits++ }, Decoration: nui.Decoration{UI: singleParts}})
		})
		ui.Textf(c, "Value: %g · Changes: %d · Commits: %d", s.Value, s.Changes, s.Commits)
		nui.FormField(c.Key("range-field"), t, nui.FormFieldOptions{}, func(ft nui.Theme) {
			nui.RangeSlider(c.Key("range"), ft, &s.Low, &s.High, nui.RangeSliderOptions{Label: "Range", Max: 100, Step: 5, Orientation: orientation, Tooltip: true, OnChange: func(float64, float64) { s.RangeChanges++ }, OnCommit: func(float64, float64) { s.RangeCommits++ }, Decoration: nui.Decoration{UI: parts}})
		})
		ui.Textf(c, "Range: %g – %g · Changes: %d · Commits: %d", s.Low, s.High, s.RangeChanges, s.RangeCommits)
	})
}

func main() {
	s := &example{}

	s.Value = 40
	s.Low = 20
	s.High = 75
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Slider · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
