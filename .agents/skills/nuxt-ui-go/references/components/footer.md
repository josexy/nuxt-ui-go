<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Footer (`nui.FooterLayout`)

Show explanatory text and secondary actions below the content.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `FooterLayout`
- Types: `FooterOptions`
- Theme key: `nui.ComponentFooter` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

FooterLayout uses Text for copy and Actions for controls. Root, Label and Actions style the container, text and action region. Actions use a six-DIP gap and wrap when needed.

## Things to know

The native API uses Text and Actions rather than web left/default/right/top/bottom slots or UContainer breakpoint reordering. Compose additional regions with native layouts. The footer has no default top border and uses 20-DIP horizontal and 12-DIP vertical root padding.

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
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Actions int
	Last    string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := ui.LeftToRight

	var styles nui.Styles

	ui.Column(c.Key("layout")).Direction(direction).MinWidth(0).Children(func() {
		nui.FooterLayout(c.Key("footer"), t, nui.FooterOptions{Text: "由 NuxtUI Go 构建的本地工作区，可以显示较长的状态说明。", Decoration: nui.Decoration{UI: styles}, Actions: func(c *ui.Context, t nui.Theme) {
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
