<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputNumber

Enter a number or adjust it by a fixed increment.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `InputNumber`, `ParseNumber`
- Types: `InputNumberOptions`
- Theme key: `nui.ComponentInputNumber` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Optional[float64] distinguishes empty from zero; NumberState retains incomplete text. This example explicitly bounds values to 0–100; the library has no default bounds. Valid edits update immediately, while Enter snaps or restores the last valid value.

## Things to know

StepSnapping and FocusOnChange default to true; wheel changes are enabled. AllowInvalid permits finite out-of-range numbers, not NaN, Infinity or invalid text; validate business rules in the application. Input uses ASCII decimal or scientific notation without locale, currency or Intl formatting. Native blur retains drafts and controls remain Tab stops, differing from the web commit timing and focus structure.

## API

```go
func InputNumber(c *ui.Context, t Theme, value *Optional[float64], s *NumberState, o InputNumberOptions) ui.Element
```

### InputNumberOptions

```go
type InputNumberOptions struct {
	InputOptions
	// Horizontal (the default) places minus/plus around centered text.
	// Vertical stacks up/down chevrons at the trailing edge.
	Orientation Orientation
	Min, Max    Optional[float64]
	Step        float64
	// StepSnapping and FocusOnChange default to true.
	StepSnapping, FocusOnChange           Optional[bool]
	StartingValue                         Optional[float64]
	AllowInvalid                          bool
	DisableWheelChange, InvertWheelChange bool
	Increment, Decrement                  Optional[bool]
	IncrementOptions, DecrementOptions    *ButtonOptions
	IncrementIcon, DecrementIcon          *ui.SVG
	IncrementDisabled, DecrementDisabled  bool
	OnValueChange                         func(Optional[float64])
}
```

### Related types

```go
// NumberState keeps incomplete text independent of the optional numeric value.
type NumberState struct {
	Text, Error string

	Handle ui.Handle
	// contains filtered or unexported fields
}
```

Types documented in other references: `ButtonOptions` → [button.md](button.md), `InputOptions` → [input.md](input.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Lifecycle and themes (excerpt)

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. Other components are configured through their own Options at the end of this page.

#### Numbers, tags, PINs, and ratings › InputNumber

InputNumber binds an `*Optional[float64]` and keeps the input draft in a separate NumberState. This distinguishes an empty value, an incomplete input (such as a lone minus sign), and a valid number.

```go
// value nui.Optional[float64] and numberState nui.NumberState are stored per window.
nui.InputNumber(c, theme, &value, &numberState, nui.InputNumberOptions{
    InputOptions: nui.InputOptions{Label: "Quantity"},
    Orientation: nui.Vertical,
    Min: nui.Set(0.0),
    Max: nui.Set(100.0),
    Step: 1,
})
```

Step defaults to 1 and must be a finite positive number; 0 uses the default. StepSnapping defaults to true: for example, with Step=1, stepping up from 18.98 gives 19 and stepping down gives 18. With snapping off, the step is added or subtracted directly. Min is the origin of the step grid, or 0 when unset. The first increment or decrement from an empty or unparseable draft gives StartingValue, which defaults to Min or 0. An explicit starting value is only clamped to the bounds; it is snapped to the step grid on the next operation.

Up and Down change the value by one step and PageUp and PageDown by ten steps. When Min or Max is set, Home and End go to that bound. Pressing a button changes the value once immediately; holding it for 400 ms then repeats every 60 ms. Releasing it does not add another change, and repetition stops when the control is disabled or unmounted. When the input is focused and the vertical scroll amount is large enough, the mouse wheel changes the value by default, with scrolling down increasing it. DisableWheelChange turns this off and InvertWheelChange reverses the direction.

Valid edits update the model as you type. Enter snaps a valid value to the grid, or restores an invalid draft from the bound value; blur keeps the draft. AllowInvalid lets finite values go outside the bounds and keeps manual input without marking it as a range error; your app handles business validation. Numbers use ASCII decimal and scientific notation. Go's hexadecimal and underscore forms are not accepted, and IME is not enabled. NaN/Inf models, starting values, and invalid bounds are rejected. The component provides no locale, currency, or Intl formatting, and float64 is not suitable as an exact currency type.

| Orientation | Layout |
| --- | --- |
| Horizontal (default) | Minus on the left, value in the center, plus on the right |
| Vertical | Value on the left, with up and down arrows stacked on the right |

Both layouts have the same field height. The buttons inherit the field's color and size by default, use the link style, and swap start and end sides in RTL. Increment and Decrement show their buttons by default and can be turned off explicitly through Optional. IncrementOptions and DecrementOptions configure the buttons, and IncrementIcon and DecrementIcon override the icons. Explicit icons are not rotated in vertical layout. Disabling one side disables only that button; the keyboard and wheel still follow the field's rules.

FocusOnChange defaults to true, which returns focus to the input after incrementing or decrementing; false lets the button keep focus. Changing the orientation or display keeps the editor's identity, draft, and selection. Assigning a value from outside or clicking another control does not focus the input. Style receives the whole field, the editor, and the leading and trailing accessories. UI.IncrementSlot and UI.DecrementSlot style the buttons, and a custom button Style runs last. The buttons keep their native Tab stops; there is no complete mapping to the web spinbutton role.

#### Calendar, InputDate, and InputTime › Calendar: pick a date by clicking (excerpt)

Dates and ranges are always complete Gregorian dates. There is no discrete multi-date model and no separate whole-month or whole-year value model. The month and year navigation views do not implement Nuxt UI's `type="month"` or `type="year"`; manage those values with app controls such as Select or InputNumber. Other calendars, time zones, and browser DOM slots are outside this native model.

#### ColorPicker (excerpt)

An incomplete RGBA draft keeps the last valid color. The numbers use InputNumber's integer stepping. Editing an actual channel rebuilds the color from sRGB bytes, so stale wide-gamut data is not kept. A form tracks these drafts as a single color field, so the initial display and rebuilds without edits are not mistakenly marked as changed by writes from the four sub-inputs.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| InputNumber | Root, Editor, Leading, Trailing, IncrementSlot, DecrementSlot |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](theme.md).

### Options and defaults

#### Common options by component › Switches, tags, and ratings (excerpt)

InputNumber's StepSnapping, FocusOnChange, Increment, and Decrement default to true, and the mouse wheel is enabled by default. The first increment or decrement from an empty value uses StartingValue instead of adding an extra step. PinInput.Highlight defaults to false; SeparatorSlot and Root.Gap/Padding adjust the separator and spacing.

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

`PinInputOptions{Numeric: true}`, InputNumber, InputDate, and InputTime accept only numeric or date/time formats and do not enable IME. Other input methods still need testing on each target platform.

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
	Value   nui.Optional[float64]
	Number  nui.NumberState
	Changes int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.InputNumber(c.Key("number"), t, &s.Value, &s.Number, nui.InputNumberOptions{InputOptions: nui.InputOptions{Label: "Project name", Placeholder: "Type something…"}, Orientation: nui.Horizontal, Min: nui.Set(0.0), Max: nui.Set(100.0), Step: 1, StepSnapping: nui.Set(true), StartingValue: nui.Set(float64(0)), FocusOnChange: nui.Set(true), Increment: nui.Set(true), Decrement: nui.Set(true), OnValueChange: func(nui.Optional[float64]) { s.Changes++ }})
	ui.Textf(c, "Value: %v · Changes: %d", s.Value, s.Changes)
}

func main() {
	s := &example{}

	s.Value = nui.Set(3.0)
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "InputNumber · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
