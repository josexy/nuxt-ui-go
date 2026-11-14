<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Link

Open a URL or navigate with a link.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Link`
- Types: `LinkOptions`
- Theme key: `nui.ComponentLink` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Set URL to open an external address, or use `.OnClick(...)` on the returned element for application navigation. Active marks the current location and Disabled prevents activation.

## Things to know

This example only increments a count. Use URL or an application callback for navigation. Color controls active text and the focus ring; inactive text uses Muted and hovered text uses Text.

## API

```go
// Link opens URL only on activation. With URL empty, attach an application OnClick.
func Link(c *ui.Context, t Theme, label string, opts LinkOptions) ui.Element
```

### LinkOptions

```go
type LinkOptions struct {
	URL              string
	Color            Color
	Active, Disabled bool
	Decoration
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Page structure and actions (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Link | Opens a URL or runs app navigation. Supports Active, Color, and Disabled | Root, Label |

#### Components and parts › Buttons and links

Button has six variants: `Solid`, `Outline`, `Soft`, `Subtle`, `Ghost`, and `LinkVariant`. Colors are `Primary`, `Secondary`, `Success`, `Info`, `Warning`, `Error`, and `Neutral`. The older `Neutral bool` field remains for compatibility; new code should use `Color: nui.Neutral`.

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| Default base height (DIP) | 24 | 28 | 32 | 36 | 40 |

Buttons grow taller when the font's line height is large. `Block` makes a button fill the available width and `Square` gives it a square layout. Icon-only buttons are square automatically and must have an `AccessibleLabel`, or they panic. Instance `UI` can still override the padding of a square button.

**Loading and disabled:** `Loading` blocks clicks and keyboard activation and keeps the label. By default it replaces the leading icon with a spinner, or adds one. `LoadingTrailing` moves the spinner to the trailing side, although an explicit Leading still takes precedence. `LoadingIcon` changes the icon. A button that had no icon may become wider while loading.

The spinner's color comes from the Root/Leading foreground style, so set it through `UI`. A `Style(...leading).TextColor` that runs last cannot change the already resolved color of the custom-drawn icon. `Active`, `ActiveColor`, and `ActiveVariant` control the active state.

MyGo draws disabled controls at 50% opacity, and the library adds no further transparency. A child with `Disabled: false` cannot override a disabled parent. Chip is purely decorative; if a count or online status matters, include it in the host's `AccessibleLabel` or `Alt`.

Chip's four corner positions are physical left and right and are not mirrored in RTL. Inset places the chip inside the host's bounds. When the content builder is nil, the chip is shown on its own without corner offsets. The text width is computed from the font, font size, font weight, and horizontal padding in Indicator.UI. Style runs after measurement, so if you change the text size there, set the width and height as well; a large custom font size usually also needs a taller chip.

**Links:** When Link's URL is non-empty, activating the link opens the address through `Services.OpenURL`. When the URL is empty, use `.OnClick` on the returned element for in-app navigation. If you set both URL and OnClick, keep MyGo's action override rules in mind. A focused link activates with Enter; as on the web, Space does not activate it and is left to the container around it. Use Button where Space should activate too.

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

type example struct{ Clicks int }

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Link(c.Key("link"), t, "View documentation", nui.LinkOptions{}).OnClick(func() { s.Clicks++ })
	ui.Textf(c, "Activated %d times", s.Clicks)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Link · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
