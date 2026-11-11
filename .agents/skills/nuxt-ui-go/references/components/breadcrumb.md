<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Breadcrumb

Show the current page within a navigation hierarchy.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `Breadcrumb`
- Types: `BreadcrumbOptions`, `BreadcrumbItem`
- Theme key: `nui.ComponentBreadcrumb` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Give each item a stable ID and use OnSelect for application navigation. Items without callbacks are display-only; the final item is actionable when a callback is supplied. Active defaults to the final item, while explicit false removes its active style.

## Things to know

MaxItems is a native extension. Values below 2 do not truncate; larger values retain the first and trailing items with a non-interactive ellipsis. Color defaults to Primary; Neutral uses Highlighted for active text. Routing, URL opening and history remain application-owned.

## API

```go
func Breadcrumb(c *ui.Context, t Theme, items []BreadcrumbItem, o BreadcrumbOptions) ui.Element
```

### BreadcrumbOptions

```go
type BreadcrumbOptions struct {
	Label         string
	MaxItems      int
	Color         Color
	SeparatorIcon *ui.SVG
	Decoration
}
```

### BreadcrumbItem

```go
type BreadcrumbItem struct {
	ID, Label string
	Icon      *ui.SVG
	Avatar    *AvatarOptions
	Active    Optional[bool]
	Disabled  bool
	OnSelect  func()
	UI        Styles
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Shared conventions › Open state and stable IDs (excerpt)

List item IDs must be non-empty and unique, or the component panics. Tabs, Accordion, Breadcrumb, Stepper, and Timeline require uniqueness within their own list; menus require it across the whole menu tree. Display names can repeat. Required state pointers must not be nil, except for Banner, which explicitly accepts nil to mean "always shown".

#### Breadcrumb, Pagination, Stepper, and Timeline

##### Breadcrumb: show the current location

BreadcrumbItem contains ID, Label, Icon, Avatar, Active, Disabled, OnSelect, and UI. The options provide Label, Color, SeparatorIcon, and MaxItems. Items with OnSelect are actionable links, including the last one: the pointer and Enter activate them, while Space, as on the web, does not. Items with a nil callback are static text. Your app provides the path and the navigation actions.

When Active is unset, only the last item is active; an explicit false turns off the active style. Color defaults to Primary, and with Neutral the active text uses Highlighted. Inactive items are 14 DIP Muted text without padding, and actionable items change to the Text color on hover. The separator arrow is 20 DIP, and items are 6 DIP apart. In RTL, the default arrow is reversed, while a custom SeparatorIcon is kept as is. Icon takes precedence over Avatar, and the default avatar diameter is 20 DIP. The path stays on one line, and long text is truncated to fit.

When MaxItems is 2 or more, the first item and the last few items are kept, and the middle is replaced by a non-interactive ellipsis with separators on both sides. The ellipsis does not count toward the limit. MaxItems is a native extension; there is no dropdown for the collapsed items and no routing system. Item UI takes precedence over component UI, and SeparatorSlot uses the style of the preceding item. An empty Label uses the localized name of the navigation group.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Breadcrumb | Root/List, Item/Trigger, Leading, Label, SeparatorSlot, Ellipsis. Parts joined by a slash share one node |

### Options and defaults

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Breadcrumb | Color and SeparatorIcon configure appearance. Each item's Avatar, Active, and UI can be set individually. The last item is also actionable if it has OnSelect |

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
	Active string
	Clicks int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var avatar *nui.AvatarOptions

	avatar = &nui.AvatarOptions{Alt: "Ada Lovelace"}

	var styles nui.Styles

	nui.Breadcrumb(c.Key("breadcrumb"), t, []nui.BreadcrumbItem{
		{ID: "home", Label: "Home", Avatar: avatar, OnSelect: func() { s.Active = "Home"; s.Clicks++ }},
		{ID: "projects", Label: "Projects", OnSelect: func() { s.Active = "Projects"; s.Clicks++ }},
		{ID: "desktop", Label: "Desktop", OnSelect: func() { s.Active = "Desktop"; s.Clicks++ }},
		{ID: "current", Label: "NuxtUI Go", Active: nui.Set(true), UI: styles, OnSelect: func() { s.Active = "NuxtUI Go"; s.Clicks++ }},
	}, nui.BreadcrumbOptions{Label: "Location", MaxItems: 3})
	ui.Textf(c, "Selected: %s · Clicks: %d", s.Active, s.Clicks)
}

func main() {
	s := &example{}

	s.Active = "No selection"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Breadcrumb · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
