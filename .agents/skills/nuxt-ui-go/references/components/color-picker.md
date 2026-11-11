<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# ColorPicker

Choose a color from a panel, swatches or numeric values.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `ColorPicker`
- Types: `ColorPickerOptions`
- Theme key: `nui.ComponentColorPicker` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Bind ui.Color and persistent ColorPickerState. The SV area, horizontal hue and opacity tracks, HEX/HEXA, swatches and RGBA fields share one color. Five sizes change the actual picker and fields; MD defaults to a 168 DIP square.

#### Try the controls

Compare the current value with the last notification. Throttle defaults to 50 ms and coalesces OnChange only; the bound color updates immediately. Set zero for every actual change. External assignments and restoring an unchanged channel do not notify. Try the part styles and disabled state.

## Things to know

The native panel keeps MyGo’s horizontal tracks, opacity and swatches. Root, Panel, Body, Item and Label are configurable; internal SV, track and HEX parts are not exposed. ui.Color is the model, so string formatting belongs to the application. Channel edits rebuild an sRGB color; wide-gamut editing, HSL/CMYK/Lab text input and screen picking are not provided. In RTL, the native color panel stays LTR so input matches its gradients and thumbs; RGBA fields follow the form direction.

## API

```go
// ColorPicker reuses the native SV square, hue/alpha sliders and HEX editor.
// RGBA fields are library inputs. Internal SV, track and HEX parts use the
// scoped native theme; the panel and RGBA containers accept slot overrides.
func ColorPicker(c *ui.Context, t Theme, value *ui.Color, s *ColorPickerState, o ColorPickerOptions) ui.Element
```

### ColorPickerOptions

```go
type ColorPickerOptions struct {
	Label    string
	Disabled bool
	Size     Size
	// Throttle limits OnChange notifications, not the bound color's immediate
	// updates. Unset uses 50 ms; an explicit zero reports every actual change.
	Throttle Optional[time.Duration]
	OnChange func(ui.Color)
	Decoration
}
```

### Related types

```go
type ColorPickerState struct {
	// contains filtered or unexported fields
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### ColorPicker

ColorPicker binds a `*ui.Color` and keeps a ColorPickerState. The saturation/value area, the horizontal hue and alpha sliders, HEX/HEXA, the swatches, and the R/G/B/A inputs (0–255) all edit this color. Your app initializes it; unlike Vue's defaultValue, the zero value is not replaced with white automatically.

Size defaults to MD. For XS/SM/MD/LG/XL, the default selection area is 152/160/168/176/184 DIP square. Size also adjusts the native panel spacing and the numeric field sizes, and everything scales with density and text size. The root's maximum width is constrained by its container; an explicit width, padding, or a very narrow container can change the selection area's proportions. Changing the size does not reset the value, drafts, or focus.

When Throttle is unset, 50 ms is used; `nui.Set(time.Duration(0))` disables throttling, and negative values are invalid. The bound color always updates immediately, while OnChange calls are merged per interval and the last color is kept. Outside assignments are silent, and re-entering the same channel value does not notify again. Disabling, unmounting, removing the callback, or replacing the color from outside cancels a pending notification. Scheduling is handled by the window build and persistent services; no Element, Context, or background callback is stored.

An incomplete RGBA draft keeps the last valid color. The numbers use InputNumber's integer stepping. Editing an actual channel rebuilds the color from sRGB bytes, so stale wide-gamut data is not kept. A form tracks these drafts as a single color field, so the initial display and rebuilds without edits are not mistakenly marked as changed by writes from the four sub-inputs.

When Label is empty, it is inherited from FormField. Arrow keys adjust saturation and value in the SV area, the sliders keep their native keyboard behavior, and the internal names follow the window language. Each RGBA channel is an integer up to 255. HEX accepts 3-, 4-, 6-, or 8-digit forms. An invalid HEX value keeps the last valid color, and the canonical text is restored on blur following the underlying native rules. Form labels and error focus use the first RGBA number input; there is no public persistent Handle for the internal SV area or HEX field.

UI and Style support Root, Panel (the native panel container), Body (the RGBA container), Item (a channel row), and Label (a channel letter). The internal SV area, hue/alpha tracks, HEX field, and swatches have no public parts and use a local Native theme; you cannot style them with arbitrary part names. The native layout keeps horizontal tracks, alpha, and swatches, which differs from Nuxt UI's vertical hue track. In an RTL form, the native color panel is fixed to LTR so that click positions, arrow keys, gradients, and thumbs agree; the RGBA rows still follow the form's direction.

The color value is a ui.Color; your app handles serialization to HEX, RGB, or any other string format. This component provides no HSL/CMYK/Lab text editing, wide-gamut editing, screen color picking, or color management. HEX editing is affected by the IME cancel issue, and the full screen reader experience has not been verified.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| ColorPicker | Root, Panel, Body, Item, Label |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](theme.md).

#### Examples (excerpt)

Text input in InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, and ColorPicker is affected by the known IME cancel issue.

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| ColorPicker | Size=MD, with a 168 DIP SV area. An unset Throttle means 50 ms; an explicit 0 disables throttling. The bound model updates immediately |

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

#### Feature scope › Appearance, input, and overlays (excerpt)

| Area | Supported | Notes |
| --- | --- | --- |
| ColorPicker | sRGB byte channels, HEX/HEXA, native keyboard support, alpha, swatches, sizes, change throttling | Uses MyGo's horizontal tracks. The internal SV area, tracks, and HEX parts are not public. There is no HSL/CMYK/Lab text editing or wide-gamut editing. The full screen reader experience is unverified |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Value, Last ui.Color
	Picker      nui.ColorPickerState
	Changes     int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := ui.LeftToRight

	var styles nui.Styles

	ui.Column(c.Key("direction")).Direction(direction).Children(func() {
		nui.ColorPicker(c.Key("color"), t, &s.Value, &s.Picker, nui.ColorPickerOptions{Label: "Accent color", Size: nui.MD, Throttle: nui.Set(50 * time.Millisecond), Decoration: nui.Decoration{UI: styles}, OnChange: func(v ui.Color) { s.Last = v; s.Changes++ }})
	})
	nui.Button(c.Key("external"), t, "Set translucent blue", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Value = ui.Hex("#11223344") })
	ui.Textf(c, "Color: #%02X%02X%02X%02X", s.Value.R, s.Value.G, s.Value.B, s.Value.A)
	ui.Textf(c, "Last notification: #%02X%02X%02X%02X · Changes: %d", s.Last.R, s.Last.G, s.Last.B, s.Last.A, s.Changes)
}

func main() {
	s := &example{}

	s.Value = ui.Hex("#11223344")
	s.Last = s.Value
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "ColorPicker · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
