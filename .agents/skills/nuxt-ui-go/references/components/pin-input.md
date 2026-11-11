<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# PinInput

Enter a verification code or other fixed-length text.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `PinInput`
- Types: `PinInputOptions`
- Theme key: `nui.ComponentPinInput` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

PinInput binds a string and uses one real editor, with a single Tab stop. It defaults to five text cells; Numeric retains ASCII digits only. Completion fires when a user change reaches the required length.

## Things to know

Length is 1–32, default 5, counted in runes. Separator and SeparatorEvery combine without duplicate positions. Root gap/padding and SeparatorSlot font, size and padding contribute to total geometry; coordinate raw Style geometry overrides in the application. Native code does not read SMS, provide browser OTP autofill or create one input node per cell; use PinState.Handle for focus.

## API

```go
// PinInput uses one real editor, preserving paste, selection and Backspace across all slots.
func PinInput(c *ui.Context, t Theme, value *string, s *PinState, o PinInputOptions) ui.Element
```

### PinInputOptions

```go
type PinInputOptions struct {
	Label                string
	Length               int
	Disabled, Mask       bool
	Numeric, ReadOnly    bool // Text is the default; Numeric keeps only ASCII digits.
	Highlight            bool
	Placeholder          string
	Color                Color
	Variant              Variant
	Size                 Size
	Separator            []int // Insert a separator after these 1-based slots.
	SeparatorEvery       int   // Insert after every N slots; Separator also applies.
	OnChange, OnComplete func(string)
	Decoration
}
```

### Related types

```go
type PinState struct {
	Handle ui.Handle
	// contains filtered or unexported fields
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Numbers, tags, PINs, and ratings › PinInput

PinInput binds a `*string` and keeps a separate PinState. It defaults to 5 text cells; for a six-digit verification code, use `PinInputOptions{Numeric: true, Length: 6}`. Length ranges from 1 to 32 and is counted in runes. Numeric accepts only ASCII digits.

Highlight defaults to false; when enabled, all cells show a border in the current color, and a field error takes precedence. UI and Style support Root, Item, Label, and SeparatorSlot. Root.Gap/Padding and the separator's font, font size, and padding are included in the overall size calculation, and an explicit 0 is honored. If a raw Style callback changes the width or height directly, your app coordinates the parent's size.

PinInput uses one real editor behind several visual cells, so Tab enters it only once. That editor handles editing, pasting, selection, copying, and undo/redo. Numeric mode does not enable IME; text mode does. OnComplete fires only when the required length is reached and the value has changed. The component does not read SMS messages or submit automatically.

| Action or option | Behavior |
| --- | --- |
| Click an existing character | Selects that position; typing replaces it and moves to the next. Typing at the last position replaces the last character |
| Paste several characters | Overwrites the corresponding positions without pushing later characters back |
| Backspace | Deletes the selected character. Clicking an empty cell or the gap between cells moves the cursor to the end of the existing content |
| SeparatorEvery / Separator | Groups cells at a fixed interval or at given positions, for example `SeparatorEvery: 2` or `Separator: []int{3}` |
| ReadOnly / Placeholder / masking | Control read-only state, text in empty cells, and whether the content is shown |

Cells are 6 DIP apart, scaled by Density. The separator is a centered dot in the Dimmed color with equal space on both sides. The current cell shows a 3 DIP focus ring in the theme color at 25% opacity. In the error state it uses Error, in high contrast mode the ring is opaque, and None shows no ring.

A focused empty cell hides its Placeholder and shows a blinking cursor. With reduced motion, the cursor stays solid; it is hidden when read-only, disabled, or unfocused. When FormField has an Error, every cell shows the error border. Text scaling and masking do not affect hit testing.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| PinInput | Root, Item, Label, SeparatorSlot |

#### Examples (excerpt)

Text input in InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, and ColorPicker is affected by the known IME cancel issue.

### Options and defaults

#### Common options by component › Switches, tags, and ratings (excerpt)

| Component | Common options and behavior |
| --- | --- |
| PinInput | Color, Variant, and Size set the appearance. `Separator: []int{3}` adds a group break after the third cell; the control still uses one editor and one Tab stop |

InputNumber's StepSnapping, FocusOnChange, Increment, and Decrement default to true, and the mouse wheel is enabled by default. The first increment or decrement from an empty value uses StartingValue instead of adding an extra step. PinInput.Highlight defaults to false; SeparatorSlot and Root.Gap/Padding adjust the separator and spacing.

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

### Internationalization

#### What gets translated › Defaults that are easy to confuse (excerpt)

| Situation | Message and key |
| --- | --- |
| PinInput's default accessible name | `pinCharacters` in text mode and `pinDigits` in numeric mode; the count follows `Length` |

For example, a 5-cell PinInput is named `5 characters` / `5 个字符`, or `5 digits` / `5 位数字` in numeric mode. This is the editor's accessible name; the placeholder shown in each cell still comes from `Placeholder`. An explicit `Label` or FormField label takes precedence.

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
	Value             string
	Pin               nui.PinState
	Changes, Complete int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var parts nui.Styles

	nui.PinInput(c.Key("pin"), t, &s.Value, &s.Pin, nui.PinInputOptions{Label: "Verification code", Length: 6, Numeric: true, OnChange: func(string) { s.Changes++ }, OnComplete: func(string) { s.Complete++ }, Decoration: nui.Decoration{UI: parts}})
	ui.Textf(c, "Characters: %d · Changes: %d · Complete: %d", len([]rune(s.Value)), s.Changes, s.Complete)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "PinInput · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
