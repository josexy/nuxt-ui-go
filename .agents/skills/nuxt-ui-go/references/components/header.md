<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Header (`nui.HeaderLayout`)

Show a title, description and primary actions above the content.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `HeaderLayout`
- Types: `HeaderOptions`
- Theme key: `nui.ComponentHeader` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

HeaderLayout renders Title, native Description, Leading and Actions. Leading, Body, Label, Description and Actions are actual style parts. Stable region identities retain action editors when leading content changes.

## Things to know

HeaderLayout is a native application header. Its title defaults to empty; it does not create a home link, sticky positioning or a mobile menu automatically. Mode, menu, open and route auto-close are not HeaderOptions; the composed Sidebar owns menu state.

## API

```go
// HeaderLayout is the Header component; Header remains the public slot constant.
func HeaderLayout(c *ui.Context, t Theme, o HeaderOptions) ui.Element
```

### HeaderOptions

```go
type HeaderOptions struct {
	Title, Description string
	Leading, Actions   func(*ui.Context, Theme)
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

Both functions support text scaling, RTL, and composition with native layouts. They do not change the system title bar and do not provide web DOM element types or every slot. Your app composes multi-region footers, extra content above or below, and breakpoint-specific arrangements. In Go, Header and Footer are already part constants, so the functions are named HeaderLayout and FooterLayout.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| HeaderLayout | HeaderOptions | Root, Leading, Body, Label, Description, Actions |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type example struct {
	Draft, Page string
	Sidebar     nui.SidebarState
	Collapsed   bool
	MenuTrigger ui.Handle
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := ui.LeftToRight

	var styles nui.Styles

	ui.Column(c.Key("layout")).Direction(direction).MinWidth(0).Children(func() {
		options := nui.HeaderOptions{Title: "Native workspace", Description: "Native application header", Decoration: nui.Decoration{UI: styles}, Actions: func(c *ui.Context, t nui.Theme) {
			nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Header draft", Placeholder: "Keep an action draft"}).Width(150)
			nui.Button(c.Key("menu"), t, "", nui.ButtonOptions{Leading: lucide.Menu(), AccessibleLabel: "Open composed header menu", Variant: nui.Ghost, Color: nui.Neutral, Square: true}).Bind(&s.MenuTrigger).OnClick(func() { s.Sidebar.Open = true })
		}}
		nui.HeaderLayout(c.Key("header"), t, options)
		ui.Row(c.Key("content")).Height(190).MinWidth(0).Children(func() {
			nui.Sidebar(c.Key("menu-sidebar"), t, &s.Collapsed, nui.SidebarOptions{Title: "Header navigation", Description: "Composed with Sidebar", State: &s.Sidebar, Toggle: nui.Set(false), Breakpoint: nui.Set(float32(2000)), NavigationKey: s.Page, ReturnFocus: &s.MenuTrigger}, func(c *ui.Context, t nui.Theme, compact bool) {
				for _, target := range []string{"overview", "settings"} {
					nui.Button(c.Key(target), t, "Header menu "+target, nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { s.Page = target })
				}
			})
			ui.Text(c, "Header page: "+s.Page).Padding(16)
		})
	})
	ui.Text(c, "Header draft: "+s.Draft)
}

func main() {
	s := &example{}

	s.Page = "overview"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Header · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
