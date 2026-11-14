<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Container

Limit content width and add space around it.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Container`
- Types: `ContainerOptions`
- Theme key: `nui.ComponentContainer` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Build content inside `Container` and use `MaxWidth` to limit its width. Unset `Padding` follows the window width; an explicit value uses fixed DIP spacing.

## Things to know

Default horizontal padding is 16, 24 or 32 DIP for window widths below 640, from 640 to 1023, or at least 1024 DIP, scaled by theme Density. Breakpoints use the window width. `Padding: nui.Set(0)` removes padding. Unset maximum width uses Theme.ContainerWidth.

## API

```go
func Container(c *ui.Context, t Theme, opts ContainerOptions, children func()) ui.Element
```

### ContainerOptions

```go
type ContainerOptions struct {
	MaxWidth, Padding Optional[float32]
	Decoration
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Page structure and actions (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Container | Centers content and limits its width. `MaxWidth` and `Padding` are `Optional[float32]` | Root |

Container's default horizontal padding depends on the window width: 16 DIP below 640 DIP, 24 DIP from 640 to 1023 DIP, and 32 DIP at 1024 DIP or more, multiplied by the theme's Density. The breakpoints use the window width, not the width of a narrower parent. An explicit `Padding` is a fixed DIP value, and `nui.Set(0)` removes the padding; instance UI and Style can still override it.

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
	nui.Container(c.Key("container"), t, nui.ContainerOptions{MaxWidth: nui.Set(float32(480))}, func() {
		ui.Box(c.Key("boundary")).FillWidth().Height(3).Background(t.Primary).Label("Content boundary")
		nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Container draft"}).FillWidth()
	}).Label("Container preview")
}

func main() {
	s := &example{}

	s.Draft = "A persistent draft"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Container · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
