<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Main

Place the main application content in the available space.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Main`
- Theme key: `nui.ComponentMain` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Build the main page content in the `Main` callback and combine it with HeaderLayout, Sidebar and FooterLayout. The callback runs again as the UI updates, so keep application data outside the view function.

## Things to know

Main provides layout without scrolling. Add ScrollArea or MyGo Scroll when content needs to scroll.

## API

```go
func Main(c *ui.Context, t Theme, children func()) ui.Element
```

Types documented in other references: `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Page structure and actions (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Main | A flexible main content area with a minimum width and height of 0, exposed as the window's main landmark (`ui.RoleMain`), as Nuxt UI's `<main>`. Wrap it in `ui.Scroll` if it needs to scroll. A window has one main landmark: for a nested or preview instance, replace the role on the returned element, `nui.Main(c, t, body).Role(ui.RoleGroup)`, which keeps its layout | Global Root; the returned element can be configured further |

### Compatibility and known limitations

#### Feature scope › Layout and navigation (excerpt)

| Area | Current behavior and limits |
| --- | --- |
| Main / Header / Footer landmarks | Main, HeaderLayout, and FooterLayout are the main, banner, and content-information landmarks, as Nuxt UI's `<main>`, `<header>`, and `<footer>`. A window should have at most one of each, so opt nested instances out with `.Role(ui.RoleGroup)` on the returned element. NavigationMenu and Sidebar have no navigation or complementary landmark role yet |

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

type example struct{}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Main(c.Key("main"), t, func() {
		nui.Card(c.Key("content"), t, nui.CardOptions{Title: "Workspace"}, func() { ui.Text(c, "Main application content") })
	})
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Main · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
