<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Skeleton

Show a placeholder while content is loading.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Skeleton`
- Types: `SkeletonOptions`
- Theme key: `nui.ComponentSkeleton` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Width and Height are native Go dimensions. Unset values default to 160×16 DIP; default height follows text scaling. Explicit zero remains zero. Call FillWidth on the returned element to use the parent width. The application decides when to show the placeholder.

## Things to know

Animation changes the placeholder background opacity. System or theme reduced motion keeps it still; leaving the visible area or unmounting stops animation frames. Set the pulse color with UI.Root.Background; Style.Background sets the underlying surface without replacing the drawn pulse color. The native status role does not reproduce Web aria-busy / alert announcements.

## API

```go
func Skeleton(c *ui.Context, t Theme, opts SkeletonOptions) ui.Element
```

### SkeletonOptions

```go
type SkeletonOptions struct {
	Width, Height Optional[float32]
	Animate       Optional[bool]
	Decoration
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Skeleton | A loading placeholder, 160×16 DIP by default, with configurable size, corner radius, and animation | Root |

#### Components and parts › Progress and animation (excerpt)

**Skeleton**'s Width, Height, and Animate are native extensions. The default size is 160×16 DIP. The default height follows the text scale, but explicit sizes do not scale, and an explicit 0 stays 0. To fit the parent's width, call FillWidth on the returned element. Both Root.Radius and the final Style can set the corner radius, and the resulting radius clips the placeholder background and any attached content.

The pulse changes only the opacity of the custom-drawn background; set its color with UI.Root.Background. Style.Background configures the underlying surface and cannot replace the resolved pulse color. Skeleton provides a localized loading name and a native status role, but not the full aria-busy/alert announcement semantics of the web. Your app decides when to show and hide placeholders and how to present the loaded result.

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
	nui.Skeleton(c.Key("skeleton"), t, nui.SkeletonOptions{Width: nui.Set(float32(240)), Height: nui.Set(float32(72)), Decoration: nui.Decoration{UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(6))}}, Style: func(_ nui.Slot, e ui.Element) {
	}}})
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Skeleton · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
