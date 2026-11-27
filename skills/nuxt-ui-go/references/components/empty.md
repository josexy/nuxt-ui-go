<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Empty

Explain an empty state and guide the user to the next action.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `Empty`
- Types: `EmptyOptions`
- Theme key: `nui.ComponentEmpty` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Use `Title` and `Description` to explain the empty state, add an `Icon` or `Avatar`, and provide buttons via `Actions` and `Footer`. For deeper customization, `Header`, `Leading`, and `Body` replace their respective regions.

## Things to know

- Leading priority: `Avatar.Image` > `Avatar.Icon` > `Loading` spinner > `Empty.Icon` / text avatar.
- Set spinner color via `UI[Leading].Foreground`. `Loading` does not disable action buttons automatically.
- Default padding adapts to window width (16 / 24 / 32 DIP), and swapping headers preserves action controls.

## API

```go
func Empty(c *ui.Context, t Theme, opts EmptyOptions) ui.Element
```

### EmptyOptions

```go
type EmptyOptions struct {
	Title, Description string
	Icon               *ui.SVG
	Avatar             *AvatarOptions
	Size               Size
	Variant            Variant
	Loading            bool
	LoadingIcon        *ui.SVG
	Header, Leading    func()
	Body               func()
	Actions, Footer    func()
	Decoration
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Empty | An empty state with an avatar or icon, title, description, content builders, and Loading/LoadingIcon. `None` gives the "naked" look | Root, Header, Leading, Label, Description, Body, Actions, Footer |

Empty's default padding is 16, 24, or 32 DIP depending on the window width, with breakpoints at 640 and 1024 DIP. The header, body, and footer are at most 384 DIP wide. The avatar defaults to 32, 36, 40, 44, or 48 DIP for Empty sizes XS–XL; an explicit Avatar.Size takes precedence. Avatar.Image always takes precedence over an icon, and an explicit Avatar.Icon takes precedence over Empty.Icon and the loading icon.

### Inputs, selection, and fields

#### Select and InputMenu › Search in large lists (excerpt)

While a background search runs, the list shows a loading message and the query stays editable; Enter, clicks, and creation are ignored so that a stale result cannot be chosen. A new query cancels the previous search. The state's `Searching()` reports pending work. Closing the popup, removing the control, or closing the window cancels the search; call the state's `Close()` when you dispose of it yourself, and set an external Open pointer to false. Empty queries, IgnoreFilter, and ExternalFilter skip local search. On the browser backend, background search runs on the JavaScript event loop in short slices, not in a Web Worker.

### Overlays, menus, and navigation

#### Tabs, Collapsible, and Accordion › Appearance (excerpt)

TrailingIcon defaults to ChevronDown. AccordionItem.TrailingIcon takes precedence over AccordionOptions.TrailingIcon. AccordionItem.UI can override that item's Item, Header, Trigger, Leading, Label, Trailing, and Body. Empty Content creates no body and no bottom margin. The body layout container and the animation clipping container do not have separate style entry points.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct{ Created int }

func (s *example) view(c *ui.Context, t nui.Theme) {
	icon := lucide.Folder()

	var parts nui.Styles

	nui.Empty(c.Key("empty"), t, nui.EmptyOptions{Title: "No projects yet", Description: "Create a project to get started.", Icon: icon, Actions: func() {
		nui.Button(c.Key("create"), t, "Create project", nui.ButtonOptions{}).OnClick(func() { s.Created++ })
	}, Footer: func() { ui.Textf(c, "Created: %d", s.Created) }, Decoration: nui.Decoration{UI: parts}})
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Empty · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
