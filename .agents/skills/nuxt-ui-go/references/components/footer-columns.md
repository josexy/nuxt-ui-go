<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# FooterColumns

Arrange footer links into columns.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `FooterColumns`
- Types: `FooterColumnsOptions`
- Theme key: `nui.ComponentFooterColumns` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Give columns and links stable IDs. Title names a column; Items use Label, Icon, Active, Disabled, OnSelect and item Decoration. OnSelect runs before OnNavigate and PreventDefault cancels navigation. Default gaps are 32 DIP between columns, 24 between titles and lists, and 16 between links.

## Things to know

MinColumnWidth defaults to 160 DIP; nonpositive values use the default and nonfinite values are invalid. Columns and links fit available width and long link labels truncate. NavigationItem children, groups, descriptions, avatars, badges, Tooltip and Popover do not apply here. Native links notify the application by ID; URL routing, external-link indicators and extra left/right regions are not included.

## API

```go
func FooterColumns(c *ui.Context, t Theme, columns []FooterColumn, o FooterColumnsOptions) ui.Element
```

### FooterColumnsOptions

```go
type FooterColumnsOptions struct {
	MinColumnWidth float32
	OnNavigate     func(string)
	Decoration
}
```

### Related types

```go
type FooterColumn struct {
	ID, Title string
	Items     []NavigationItem
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `NavigationItem` → [navigation-menu.md](navigation-menu.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns

##### FooterColumns

FooterColumns takes columns with stable, non-empty IDs, and link IDs must be unique within each column. A non-empty Title is shown as a level-3 heading. Items use NavigationItem's ID, Label, Icon, Active, Disabled, OnSelect, and Decoration. The other fields (groups, descriptions, nested children, avatars, badges, Tooltip, and Popover) are for NavigationMenu and are not used by this component.

When the user activates an available link, the item's OnSelect is called first, with NavigationEvent.ID set to the target ID; PreventDefault cancels the root OnNavigate. Disabled links do not run callbacks. Your app controls Active; active links use Primary with a medium font weight by default, inactive links use Muted, and hover uses Text. Natively, links take the pointer and Enter; as on the web, Space does not activate a link but is left to the container around it. There are no URLs, automatic route matching, or external link markers.

MinColumnWidth defaults to 160 DIP. Non-positive values use the default, and non-finite values are invalid. This value is the initial column width in the flexible layout and drives wrapping; columns and links are constrained by the available width. Columns are top-aligned. By default, columns are 32 DIP apart, the title is 24 DIP above the list, links are 16 DIP apart, and long links are truncated to one line. RTL adjusts the order, and reordering columns keeps their stable IDs.

Root, Body (each column), Label (the column title), List, Item (a link), ItemLeadingSlot, and ItemLabelSlot can each be styled. Item UI overrides component UI, and the item's Style runs after the component's Style. Style links with Item, ItemLeadingSlot, and ItemLabelSlot.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| FooterColumns | `[]FooterColumn` | Root, Body, Label, List, Item, ItemLeadingSlot, ItemLabelSlot |

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| Header, Footer, FooterColumns | Header Title is empty. Footer has no top border. FooterColumns uses a preferred minimum column width of 160 DIP and a column gap of 32 DIP |

### Compatibility and known limitations

#### Feature scope › Layout and navigation (excerpt)

| Area | Current behavior and limits |
| --- | --- |
| Header / Footer / FooterColumns | Native header and footer areas use builders. They do not include every web slot, routing, or sticky positioning. Combine the header menu from components such as Sidebar. FooterColumns supports Active, per-item cancellable navigation, and styles; other NavigationItem fields have their own applicability notes |

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

type example struct {
	Active               string
	Selects, Navigations int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := ui.LeftToRight

	var styles, itemStyles nui.Styles

	onSelect := func(e *nui.NavigationEvent) {
		s.Selects++
	}
	ui.Column(c.Key("direction")).Direction(direction).MinWidth(0).Children(func() {
		nui.FooterColumns(c.Key("footer"), t, []nui.FooterColumn{
			{ID: "product", Title: "Product", Items: []nui.NavigationItem{{ID: "components", Label: "Components", Icon: lucide.Layers(), Active: s.Active == "components", OnSelect: onSelect}, {ID: "examples", Label: "Examples", Active: s.Active == "examples", OnSelect: onSelect, Decoration: nui.Decoration{UI: itemStyles}}}},
			{ID: "resources", Title: "Resources", Items: []nui.NavigationItem{{ID: "guide", Label: "完整组件说明与使用指南", Active: s.Active == "guide", OnSelect: onSelect}, {ID: "unavailable", Label: "Unavailable", Disabled: true, OnSelect: onSelect}}},
		}, nui.FooterColumnsOptions{MinColumnWidth: float32(160), Decoration: nui.Decoration{UI: styles}, OnNavigate: func(id string) { s.Active = id; s.Navigations++ }})
	})
	ui.Textf(c, "Active: %s · Selects: %d · Navigations: %d", s.Active, s.Selects, s.Navigations)
}

func main() {
	s := &example{}

	s.Active = "components"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "FooterColumns · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
