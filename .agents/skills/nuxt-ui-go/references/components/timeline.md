<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Timeline

Display activities or milestones in sequence.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Timeline`
- Types: `TimelineOptions`, `TimelineItem`
- Theme key: `nui.ComponentTimeline` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Timeline is vertical by default. Current is an optional zero-based index, overridden by a non-empty CurrentID. Reverse changes progress interpretation without reordering items. Time is an application-formatted string, styled independently through DateSlot.

Connectors retain their 2 DIP width, rounded consistently to whole device pixels when painted.

## Things to know

OnSelect does not change progress automatically. An unmatched CurrentID does not fall back to Current, and invalid indexes highlight no items. Per-item Color and Content are native extensions; sizes are limited to XS–XL.

## API

```go
func Timeline(c *ui.Context, t Theme, items []TimelineItem, o TimelineOptions) ui.Element
```

### TimelineOptions

```go
type TimelineOptions struct {
	Label      string
	Horizontal bool
	Color      Color
	Size       Size
	Current    Optional[int] // Zero-based step; unset means no active/completed steps.
	CurrentID  string        // Takes precedence over Current when non-empty.
	Reverse    bool          // Changes progress interpretation, not item order.
	OnSelect   func(string)
	Decoration
}
```

### TimelineItem

```go
type TimelineItem struct {
	ID, Title, Description, Time string
	Icon                         *ui.SVG
	Avatar                       *AvatarOptions
	Color                        Color
	UI                           Styles
	Content                      func(*ui.Context, Theme)
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Shared conventions › Open state and stable IDs (excerpt)

List item IDs must be non-empty and unique, or the component panics. Tabs, Accordion, Breadcrumb, Stepper, and Timeline require uniqueness within their own list; menus require it across the whole menu tree. Display names can repeat. Required state pointers must not be nil, except for Banner, which explicitly accepts nil to mean "always shown".

#### Breadcrumb, Pagination, Stepper, and Timeline

##### Timeline: display a timeline

TimelineItem contains ID, Title, Description, Time, Icon, Color, and Content, and can also provide Avatar and UI. Timelines are vertical by default; Horizontal places the indicator icon and connecting line above the text.

Current is an optional zero-based index, and CurrentID takes precedence over it. When neither is set, no item is active or completed. Reverse changes how progress is interpreted but does not reorder the data. Clicking an item only calls OnSelect; it does not change the progress.

Time is a string your app formats. By default, its font size and line height are 12/20 DIP, configurable with DateSlot; the title and description default to 14/20 DIP. Header styles the container of the marker and connecting line, and horizontal items share the width equally. With Neutral, the active marker and completed lines use InvertedBackground. With OnSelect, items can be operated by click, or by Tab and Enter/Space; without a callback, they are not in the keyboard focus order.

If CurrentID has no match, Current is not used as a fallback, and an invalid index highlights nothing. The component uses list semantics and does not sort, convert time zones, or schedule animations. Per-item Color, extra Content, and the five sizes are deliberate native configuration options.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Timeline | Root, Item, Header, Indicator, SeparatorSlot, Body, DateSlot, Label, Description |

Tabs, Stepper, and Timeline accept global and instance Color and Size; Banner and Toast accept Color. Color on menu items and Timeline items is a per-item setting. Other styles are configured through the parts listed above.

### Options and defaults

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Timeline | CurrentID takes precedence over Current. Reverse does not reorder the data. DateSlot styles Time. OnSelect does not change progress by itself |

### Internationalization

#### What gets translated › Text your app translates (excerpt)

Text you pass explicitly, such as `Label`, `Description`, `Placeholder`, `ErrorMessage`, `EmptyText`, `ChooseLabel`, and `CloseLabel`, is displayed as is. Your app also translates business validation errors, server errors, Toast titles, menu items, table data, and Timeline dates.

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
	Selected   string
	Selections int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var styles nui.Styles

	nui.Timeline(c.Key("timeline"), t, []nui.TimelineItem{
		{ID: "create", Title: "Created", Time: "09:00", Description: "Project initialized", UI: styles},
		{ID: "build", Title: "Built", Time: "09:10", Description: "Checks passed", Icon: lucide.Check()},
		{ID: "ship", Title: "Published", Time: "09:20", Description: "Ready to share", Icon: lucide.Send()},
	}, nui.TimelineOptions{Label: "Activity", Current: nui.Set(1), OnSelect: func(id string) { s.Selected = id; s.Selections++ }})
	ui.Textf(c, "Selected: %s · Selections: %d", s.Selected, s.Selections)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Timeline · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
