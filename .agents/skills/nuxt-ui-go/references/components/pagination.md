<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Pagination

Change data pages and show the current page number.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `Pagination`
- Types: `PaginationOptions`
- Theme key: `nui.ComponentPagination` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Store the one-based page in a `*int`. PageSize defaults to 10 and Total is the item count. SiblingCount supports explicit zero. OnChange reports user selection only; clamping after a total change is silent.

## Things to know

PageSize≤0 uses 10, and Total≤0 retains page 1. Siblings are bounded to 0–10. Ellipses are inert. Root/List share a layout element; Item continues styling every button for compatibility, with First, Previous, Next and Last providing more specific control styles.

## API

```go
func Pagination(c *ui.Context, t Theme, page *int, o PaginationOptions) ui.Element
```

### PaginationOptions

```go
type PaginationOptions struct {
	Label                     string
	Total, PageSize, Siblings int
	// SiblingCount overrides legacy Siblings and can explicitly request zero.
	SiblingCount                                          Optional[int]
	ShowControls, ShowEdges                               Optional[bool]
	Color, ActiveColor                                    Color
	Variant, ActiveVariant                                Variant
	Size                                                  Size
	Disabled                                              bool
	FirstIcon, PrevIcon, NextIcon, LastIcon, EllipsisIcon *ui.SVG
	OnChange                                              func(int)
	Decoration
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Breadcrumb, Pagination, Stepper, and Timeline

##### Pagination: control the page

Pagination binds an `*int`, and pages start at 1. Total is the total number of items, and a PageSize of 0 or less defaults to 10. When the total shrinks, the page is clamped to the valid range automatically; such corrections do not fire OnChange, which fires only when the user goes to a different page.

The component does not query data; your app loads the current page when it receives a change. A Siblings value of 0 or less defaults to 2, with a maximum of 10; for an explicit zero, use SiblingCount. ShowEdges defaults to false and Size to MD. For more appearance options, see Pagination siblings and appearance.

Navigation controls and page numbers are square buttons. Controls at the bounds are disabled, and the arrows follow RTL. The ellipsis is not interactive. The accessible names for first, previous, next, last, and page N follow the language.

FirstIcon, PrevIcon, NextIcon, LastIcon, and EllipsisIcon override the default icons; explicit icons are not mirrored in RTL. ActiveColor and ActiveVariant default to Primary / Solid, and ordinary buttons default to Neutral / Outline. Root and List share one layout element. For compatibility, Item still applies to all buttons; First, Previous, Next, and Last can further target the control buttons, and Ellipsis styles the ellipsis. Your app handles page links, data queries, and browser-style history in OnChange.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Pagination | Root/List, Item, First, Previous, Next, Last, Ellipsis, Leading, Label. Root and List share one node |

### Options and defaults

#### Unset, zero, and default values (excerpt)

A Go zero value is not always the value a component ends up using. For example, Pagination treats `PageSize: 0` as the default of 10. Fields that must distinguish "unset" from an explicit `false` or `0` use `Optional[T]`.

For example, to hide the Toast close button, write `Close: nui.Set(false)`. To show no sibling pages in Pagination, write `SiblingCount: nui.Set(0)`. Writing only `Siblings: 0` uses the default of 2.

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| Pagination | PageSize=10, Siblings=2, ShowEdges=false, ShowControls=true, Size=MD |

#### Selection and pagination › Pagination siblings and appearance

`SiblingCount` takes precedence over the older `Siblings` and accepts 0–10. For example, on page 5 of 10 with zero siblings:

| Setting | Display |
| --- | --- |
| ShowEdges=true | `1 … 5 … 10` |
| ShowEdges=false | `5` |

Near the first or last page, the range of edge pages widens. When there are only a few pages, no ellipsis is shown. ShowControls controls the previous/next buttons and ShowEdges controls the first and last page numbers. Size, Color, and Variant style the ordinary page buttons; ActiveColor and ActiveVariant style the current page.

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Pagination | FirstIcon, PrevIcon, NextIcon, LastIcon, and EllipsisIcon override the default icons. ActiveColor and ActiveVariant style the current page |

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

type example struct{ Page, Changes int }

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Pagination(c.Key("pagination"), t, &s.Page, nui.PaginationOptions{Label: "Pages", Total: 100, PageSize: 10, SiblingCount: nui.Set(1), OnChange: func(int) { s.Changes++ }})
	ui.Textf(c, "Page %d", s.Page)
	ui.Textf(c, "Changes: %d", s.Changes)
}

func main() {
	s := &example{}

	s.Page = 5
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Pagination · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
