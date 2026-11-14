<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Separator

Separate content with a line, label or icon.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Separator`
- Types: `SeparatorOptions`
- Theme key: `nui.ComponentSeparator` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Use Orientation for a horizontal or vertical line and Label to insert text. Size adjusts line thickness and Color sets its color. The container and component dimensions determine its length.

XS–XL retain 1–5 DIP widths, rounded to whole device pixels when painted. Dash lengths and gaps also align to pixels. Set stroke colors with Indicator.Background.

## Things to know

Content takes precedence over Label, Icon and Avatar. Vertical height defaults to 48 DIP; this example explicitly uses 160 DIP for long labels and editor content. This is a decorative divider: MyGo has no static separator accessibility role, so no ineffective Decorative option is exposed.

## API

```go
func Separator(c *ui.Context, t Theme, opts SeparatorOptions) ui.Element
```

### SeparatorOptions

```go
type SeparatorOptions struct {
	Label       string
	Icon        *ui.SVG
	Avatar      *AvatarOptions
	Content     func()
	Color       Color
	Size        Size // XS..XL correspond to 1..5 DIP.
	Type        SeparatorType
	Position    SeparatorPosition
	Vertical    bool // Kept for existing callers; either this or Orientation can select vertical.
	Orientation Orientation
	Decoration
}
```

### Related types

```go
type SeparatorType uint8

const (
	SeparatorSolid SeparatorType = iota
	SeparatorDashed
	SeparatorDotted
)
```

```go
type SeparatorPosition uint8

const (
	SeparatorCenter SeparatorPosition = iota
	SeparatorStart
	SeparatorEnd
)
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Separator | A horizontal or vertical line that can contain text, an icon, an avatar, or custom content | Root, Body, Label, Leading, Indicator |

Separator's `Size` maps XS–XL to line widths of 1–5 DIP. Painting rounds the width to whole device pixels before aligning its position, so equal widths remain equal at fractional DPI and UI zoom. Dashed and dotted segments use whole-pixel lengths and gaps; the last segment may be shorter. Card section dividers and default menu/Listbox separators paint one device pixel while preserving DIP layout spacing. `Type` is one of the Solid, Dashed, and Dotted constants, and `Position` is Center, Start, or End. Content precedence is `Content → Label → Icon → Avatar`. The default label is 14 DIP, the icon 20 DIP, and the avatar 20 DIP; an explicit Avatar.Size takes precedence. Leading refers to the icon or avatar root. Vertical labels can wrap. A vertical separator is 48 DIP tall by default, which you can change on the returned element; give long labels and interactive content enough height. The content container keeps its identity when the position or orientation changes, but controls inside it still need stable keys.

Separator currently draws a decorative line and leaves the accessibility of its text and interactive content to that content. MyGo has no static separator role, so this is not equivalent to upstream `decorative=false`, and the library does not add a Decorative option that would have no effect. Set every stroke type's color with `UI[Indicator].Background`. The final Style can replace the stroke with its own Background or Material; that replacement owns pixel alignment.

### Forms, dates, and data components

#### Numbers, tags, PINs, and ratings › PinInput (excerpt)

| Action or option | Behavior |
| --- | --- |
| SeparatorEvery / Separator | Groups cells at a fixed interval or at given positions, for example `SeparatorEvery: 2` or `Separator: []int{3}` |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct{ Draft string }

func (s *example) view(c *ui.Context, t nui.Theme) {
	options := nui.SeparatorOptions{Label: "Continue", Orientation: nui.Horizontal, Position: nui.SeparatorCenter, Type: nui.SeparatorSolid}
	part := nui.Separator(c.Key("separator"), t, options)
	if options.Vertical || options.Orientation == nui.Vertical {
		part.Width(180).Height(160)
	}
	ui.Text(c, "Draft: "+s.Draft)
	ui.Text(c, "Pixel alignment comparison")
	comparison := options
	comparison.Label = ""
	comparison.Content = nil
	comparison.Icon = nil
	comparison.Avatar = nil
	comparison.Vertical = false
	comparison.Orientation = nui.Horizontal
	for i := range 3 {
		ui.Column(c.Key(fmt.Sprintf("comparison-%d", i))).PaddingY(float32(i)*.3 + 2).Children(func() { nui.Separator(c, t, comparison) })
	}
}

func main() {
	s := &example{}

	s.Draft = "Persistent draft"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Separator · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
