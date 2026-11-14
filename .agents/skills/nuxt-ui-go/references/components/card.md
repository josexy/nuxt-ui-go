<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Card

Group a title, description and related content in a card.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Card`
- Types: `CardOptions`
- Theme key: `nui.ComponentCard` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Use Title and Description for the heading, and the content callback for the body and buttons. Header and Footer callbacks customize the heading area and footer.

Section dividers paint one device pixel while retaining 1 DIP of layout space.

## Things to know

Default padding is 16 DIP below a window width of 640 DIP. At 640 and above, horizontal and body vertical padding are 24; header/footer vertical padding stays 16. A description without a title retains its 4 DIP top margin. Removing the header does not remount the body editor.

## API

```go
func Card(c *ui.Context, t Theme, opts CardOptions, body func()) ui.Element
```

### CardOptions

```go
type CardOptions struct {
	Variant            Variant
	Title, Description string
	Header, Footer     func()
	Decoration
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Card | A card. Defaults to Outline and also supports Solid/Soft/Subtle. Has a title, description, Header, body, and Footer | Root, Header, Body, Footer, Label, Description |

If a Card's body builder is nil, no body is created, and separators appear only between regions that exist. When the window is narrower than 640 DIP, every region has 16 DIP of padding. From 640 DIP up, the horizontal padding and the body's vertical padding are 24 DIP, while the header and footer keep 16 DIP of vertical padding. All of these scale with Density. A description without a title keeps a 4 DIP top margin.

#### Components and parts › Indicators, separators, and loading (excerpt)

Separator's `Size` maps XS–XL to line widths of 1–5 DIP. Painting rounds the width to whole device pixels before aligning its position, so equal widths remain equal at fractional DPI and UI zoom. Dashed and dotted segments use whole-pixel lengths and gaps; the last segment may be shorter. Card section dividers and default menu/Listbox separators paint one device pixel while preserving DIP layout spacing. `Type` is one of the Solid, Dashed, and Dotted constants, and `Position` is Center, Start, or End. Content precedence is `Content → Label → Icon → Avatar`. The default label is 14 DIP, the icon 20 DIP, and the avatar 20 DIP; an explicit Avatar.Size takes precedence. Leading refers to the icon or avatar root. Vertical labels can wrap. A vertical separator is 48 DIP tall by default, which you can change on the returned element; give long labels and interactive content enough height. The content container keeps its identity when the position or orientation changes, but controls inside it still need stable keys.

### Inputs, selection, and fields

#### Checkbox, CheckboxGroup, RadioGroup, and Switch › Icons, styles, and spacing (excerpt)

A card with a visible indicator shows Highlight only on the indicator; with the indicator hidden, the card's frame shows the highlight. Card styles follow the usual override order: component defaults, the instance's semantic options, the instance's UI, and then Style.

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

type example struct{ Draft string }

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Card(c.Key("card"), t, nui.CardOptions{Title: "Project overview", Description: "A native Go component"}, func() {
		ui.Text(c, "A focused space for related content.")
		nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Card draft"}).FillWidth()
	})
	ui.Text(c, "Draft: "+s.Draft)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Card · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
