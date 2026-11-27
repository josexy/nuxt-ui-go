<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Footer (`nui.FooterLayout`)

Show explanatory copy and secondary actions at the bottom of a page.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `FooterLayout`
- Types: `FooterOptions`
- Theme key: `nui.ComponentFooter` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

`FooterLayout` renders copy via `Text` and controls via `Actions` (6 DIP default gap, wrapping when space is tight). Customize `Root`, `Label`, and `Actions` through `UI`.

## Things to know

- Exposes `Text` and `Actions` regions (no default top border, 20×12 DIP padding); compose native layouts for multi-zone footers.
- Sets the window content-info landmark (`ui.RoleContentInfo`), so a window should have at most one; call `.Role(ui.RoleGroup)` on nested instances.

## API

```go
func FooterLayout(c *ui.Context, t Theme, o FooterOptions) ui.Element
```

### FooterOptions

```go
type FooterOptions struct {
	Text    string
	Actions func(*ui.Context, Theme)
	Decoration
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns

##### HeaderLayout and FooterLayout

HeaderLayout is a native app header with Title, the native-extension Description, Leading, and Actions. The title is empty by default and, when non-empty, is a level-1 heading; an empty title creates no empty heading node. The minimum height is 64 DIP scaled by the theme, the horizontal and vertical padding is 24 and 8 DIP, the gap between regions is 12 DIP, and the gap inside the leading and action areas is 6 DIP. Root, Leading, Body (the title text container), Label, Description, and Actions all have working style entry points. Changing Leading keeps the identity and input state of the controls in Actions.

HeaderLayout does not create a home link, web sticky positioning, a breakpoint menu, or a route listener. Header's Body part refers to the title text container, which differs from the web Header's menu body slot.

FooterLayout shows a description with Text and builds actions with Actions. Root, Label, and Actions style the whole footer, the text, and the action container. The default horizontal and vertical padding is 20 and 12 DIP, the gap between regions is 12 DIP, the gap inside the action area is 6 DIP, and there is no top border. Text and actions can wrap within the available width, and action styles are not mistakenly applied to individual buttons.

HeaderLayout is the window's banner landmark (`ui.RoleBanner`) and FooterLayout its content-information landmark (`ui.RoleContentInfo`), as Nuxt UI's `<header>` and `<footer>`; the header's title stays a level-1 heading inside the banner. A window has at most one of each. When you build another Header or Footer inside the page, as a preview or inside a card, replace its role on the returned element in the same build, for example `nui.HeaderLayout(c, t, o).Role(ui.RoleGroup)`; its layout, styles, focus, and children stay the same.

Both functions support text scaling, RTL, and composition with native layouts. They do not change the system title bar and do not provide web DOM element types or every slot. Your app composes multi-region footers, extra content above or below, and breakpoint-specific arrangements. In Go, Header and Footer are already part constants, so the functions are named HeaderLayout and FooterLayout.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| FooterLayout | FooterOptions | Root, Label, Actions |

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
	Actions int
	Last    string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := ui.LeftToRight

	var styles nui.Styles

	ui.Column(c.Key("layout")).Direction(direction).MinWidth(0).Children(func() {
		nui.FooterLayout(c.Key("footer"), t, nui.FooterOptions{Text: "A native workspace built with NuxtUI Go, with room for a longer status message.", Decoration: nui.Decoration{UI: styles}, Actions: func(c *ui.Context, t nui.Theme) {
			nui.Button(c.Key("review"), t, "Review", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Actions++; s.Last = "Review selected" })
			nui.Button(c.Key("help"), t, "Help", nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { s.Actions++; s.Last = "Help selected" })
		}})
	})
	ui.Textf(c, "Actions: %d · %s", s.Actions, s.Last)
}

func main() {
	s := &example{}

	s.Last = "No action"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Footer · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
