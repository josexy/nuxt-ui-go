<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Chip

Add a count or status marker to the corner of other content.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Chip`
- Types: `ChipOptions`
- Theme key: `nui.ComponentChip` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Build the host in Chip’s content callback and configure Text, Color, Position and Inset. Hidden removes only the marker. A nil content callback renders a standalone chip. Left and right are physical positions and do not mirror in RTL.

## Things to know

Text width uses Indicator UI font, size, weight and horizontal padding. Style runs after measurement; size changes made there also need explicit dimensions. This example increases height for Styled text. Include meaningful counts or status in the host’s accessible name; the marker itself does not handle clicks.

## API

```go
func Chip(c *ui.Context, t Theme, opts ChipOptions, content func()) ui.Element
```

### ChipOptions

```go
type ChipOptions struct {
	Color         Color
	Size          Size
	Text          string
	Position      ChipPosition
	Inset, Hidden bool
	Decoration
}
```

### Related types

```go
type ChipPosition uint8

const (
	TopRight ChipPosition = iota
	TopLeft
	BottomRight
	BottomLeft
)
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

User.Size sets the size of the avatar and its chip, overriding the nested Avatar.Size and Chip.Size. Chip.Inset is always true; the other chip settings (position, color, text, and Hidden) are passed through. Without an avatar, no chip is shown. Leading replaces the default avatar area and Body replaces the name and description. Each region is built with a stable identity.

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Chip | A badge indicator with four corner positions, Inset, Hidden, colors, sizes, and text. Does not intercept clicks on its host; without content, it can be shown on its own | Root, Indicator |

#### Components and parts › Buttons and links (excerpt)

MyGo draws disabled controls at 50% opacity, and the library adds no further transparency. A child with `Disabled: false` cannot override a disabled parent. Chip is purely decorative; if a count or online status matters, include it in the host's `AccessibleLabel` or `Alt`.

Chip's four corner positions are physical left and right and are not mirrored in RTL. Inset places the chip inside the host's bounds. When the content builder is nil, the chip is shown on its own without corner offsets. The text width is computed from the font, font size, font weight, and horizontal padding in Indicator.UI. Style runs after measurement, so if you change the text size there, set the width and height as well; a large custom font size usually also needs a taller chip.

#### Components and parts › Images, avatars, and icons (excerpt)

The Body part is the bitmap. By default it fills the root container and is center-cropped (Cover). If Style changes the root size, the image still fills the available area. Sizes XS–XL map to 24, 28, 32, 36, and 40 DIP; adjust other sizes with a synchronous Style. The library does not add upstream's 3xs, 2xs, 2xl, and 3xl sizes to the shared enum. To add a badge to an avatar, combine them as `Chip(..., ChipOptions{Inset: true}, func() { Avatar(...) })`; Chip controls the color, position, and visibility.

### Advanced interaction, charts, and layout

#### CommandPalette › Shortcuts and search (excerpt)

| Option or behavior | Description |
| --- | --- |
| Kbds / Prefix / Suffix / Avatar / Chip / Loading / Active | Command content and appearance. Kbds only display shortcuts, and Loading does not disable the command |

### Forms, dates, and data components

#### Listbox, SelectMenu, and async search › Listbox: a selectable list shown inline (excerpt)

Icon, Avatar, and Chip are checked in that order for the leading content, which aligns with the first line of text. SelectedIcon overrides the default check icon. Per-item options provide UI, Style, and OnSelect. The callback receives the ID and the selected state before the change, and can call PreventDefault to cancel this selection and the component's OnSelect. For the row appearance, see [Option row appearance](select.md). An MD row is about 32 DIP tall, plus one line when it has a description.

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

type example struct{ Count int }

func (s *example) view(c *ui.Context, t nui.Theme) {
	var content func()

	content = func() {
		nui.Button(c.Key("host"), t, "Notifications", nui.ButtonOptions{Variant: nui.Outline, AccessibleLabel: "Notifications " + "3"}).OnClick(func() { s.Count++ })
	}

	var styles nui.Styles

	nui.Chip(c.Key("chip"), t, nui.ChipOptions{Text: "3", Position: nui.TopRight, UI: styles, Style: func(slot nui.Slot, e ui.Element) {
	}}, content).AlignSelf(ui.Start)
	ui.Textf(c, "Host actions: %d", s.Count)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Chip · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
