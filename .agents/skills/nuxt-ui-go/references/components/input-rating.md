<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputRating

Choose a whole-number star rating.

- Category: Forms and data
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `InputRating`
- Types: `InputRatingOptions`
- Theme key: `nui.ComponentInputRating` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

InputRating binds an integer; 0 means unrated. Max defaults to 5 and accepts 1–20. External values and maximum changes clamp silently. Clearable and Hoverable default to false.

#### Interaction and appearance

Arrow keys follow orientation and RTL. Home / PageUp selects 1, End / PageDown selects Max, and Space can select the first item or clear the current value. Read-only ratings retain Tab focus, normal colors and the default cursor.

The empty icon remains behind the filled overlay; IconSlot can override the foreground inherited from Indicator. Neutral uses Highlighted for filled icons and InvertedBackground for its focus ring.

Reduced motion keeps hover scaling without a transition. Restoring animation does not revert the scale.

## Things to know

The integer binding has no fractional step or arbitrary item builder. Native ratings use one slider focus target while the web version uses a radio group; read-only focus and Space after clearing differ. Applications own validation, persistence and submission.

## API

```go
func InputRating(c *ui.Context, t Theme, value *int, o InputRatingOptions) ui.Element
```

### InputRatingOptions

```go
type InputRatingOptions struct {
	Label              string
	Max                int
	Disabled, ReadOnly bool
	Color              Color
	Size               Size
	Orientation        Orientation
	Clearable          bool
	Hoverable          Optional[bool]
	Icon, EmptyIcon    *ui.SVG
	OnChange           func(int)
	Decoration
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Numbers, tags, PINs, and ratings › InputRating

InputRating binds an `*int`. Max defaults to 5, and 0 means not rated. Only whole-number ratings are supported; there are no half stars, arbitrary precision, or drag-to-rate.

| Action | Behavior |
| --- | --- |
| ←→ (horizontal) / ↑↓ (vertical) | Move one level and stop at the ends. Horizontal keys follow RTL, and keys for the other axis do nothing |
| Home / PageUp | Set to 1 |
| End / PageDown | Set to Max |
| Space | Sets an empty rating to 1. With Clearable, resets the current rating to 0 |
| Enter | Not handled |

Arrow keys never lower the rating from 1 to 0; clearing requires Clearable. The keyboard focus ring is on the current star, or on the first star when there is no rating. A read-only rating can still receive focus with Tab, but ignores keys that would change it. Assistive technology sees the rating as a slider with its value, reported read-only while ReadOnly is set, so screen readers neither set nor step it.

Stars are 12/16/20/24/28 DIP for sizes XS to XL, 2 DIP apart, and the component is as wide as its content. When interactive, the hovered star scales to 1.1×. With Hoverable on, hovering only previews the rating; the preview stays while the pointer crosses gaps between stars and is canceled when the pointer leaves or the rating is cleared. When FormField has an error, filled stars and the focus ring use the Error color.

UI and Style support Root, Item, Indicator, IconSlot, and EmptyIconSlot. The empty icon is drawn as the base layer, and Indicator and IconSlot are layered on top within the rated or hovered range. If the two icons differ, the base layer may still show around the overlay. IconSlot inherits Indicator's foreground color and can override it. Root's Gap/Padding are included in the overall size calculation, and Item's Shadow can override the default focus ring. With Neutral, rated icons use Highlighted and the focus ring uses InvertedBackground. A read-only rating keeps its normal colors and the default cursor.

Max accepts 1–20, and 0 uses the default of 5. When the rating or Max changes from outside, the rating is silently clamped to 0–Max; OnChange reports only actual user changes. Hoverable and Clearable default to false. Nuxt UI's step can represent fractional ratings such as half stars, but this library's integer binding has no corresponding field. There is also no arbitrary item content builder, HTML form field, or web radio node.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| InputRating | Root, Item, Indicator, IconSlot, EmptyIconSlot |

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| InputTags, InputRating, Slider | InputTags: AddOnPaste=false. InputRating: Hoverable=false. Slider: Step=1 |

#### Common options by component › Switches, tags, and ratings (excerpt)

| Component | Common options and behavior |
| --- | --- |
| InputRating | Clearable lets clicking the current rating or pressing Space reset it to zero. Hoverable defaults to false. An unset EmptyIcon reuses Icon. Indicator, IconSlot, and EmptyIconSlot style the fill layer and the two icons |

InputRating's horizontal arrow keys follow RTL; in vertical layout, use ↑ and ↓, and the other axis's keys do nothing. The rating stops at both ends. Home and PageUp go to 1, End and PageDown go to Max. The hover preview disappears as soon as the rating is cleared.

#### Limitations (excerpt)

Slider and RangeSlider expose their value, range, and step to assistive technology, but MyGo has no API for setting vertical orientation metadata. This affects every vertical mode, including sliders without Inverted. Slider tooltips can be positioned per thumb, but they have no separate delay setting. InputRating uses a single-focus slider role.

### Compatibility and known limitations

#### Accessibility and motion (excerpt)

| Area | Current limitation |
| --- | --- |
| InputRating | Uses a single-focus slider role, which differs from the radio semantics used on the web |

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
	Value   int
	Changes int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var parts nui.Styles

	nui.InputRating(c.Key("rating"), t, &s.Value, nui.InputRatingOptions{Label: "Rating", Max: 5, Orientation: nui.Horizontal, Hoverable: nui.Set(false), UI: parts, OnChange: func(int) { s.Changes++ }})
	ui.Textf(c, "Rating: %d · Changes: %d", s.Value, s.Changes)
}

func main() {
	s := &example{}

	s.Value = 3
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "InputRating · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
