<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Banner

Display a persistent, dismissible message within a page.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Banner`
- Types: `BannerOptions`
- Theme key: `nui.ComponentBanner` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Control visibility with an application-owned `*bool`. A bound state shows the close button by default; CloseButton can hide it, while Close and CloseIcon customize it. OnSelect handles banner activation independently from inner actions.

## Things to know

Content remains on one line. It is centered from a 1024 DIP window width and start-aligned below that breakpoint. A nil state has no close button by default; explicitly enabling it uses local state that resets on unmount. Cross-launch persistence and navigation are application-owned, without localStorage.

## API

```go
func Banner(c *ui.Context, t Theme, open *bool, o BannerOptions) ui.Element
```

### BannerOptions

```go
type BannerOptions struct {
	Title, Description, CloseLabel string
	Color                          Color
	Icon                           *ui.SVG
	Actions                        func(*ui.Context, Theme)
	CloseButton                    Optional[bool]
	Close                          *ButtonOptions
	CloseIcon                      *ui.SVG
	OnSelect                       func()
	OnClose                        func()
	Decoration
}
```

Types documented in other references: `ButtonOptions` → [button.md](button.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Shared conventions › Open state and stable IDs (excerpt)

List item IDs must be non-empty and unique, or the component panics. Tabs, Accordion, Breadcrumb, Stepper, and Timeline require uniqueness within their own list; menus require it across the whole menu tree. Display names can repeat. Required state pointers must not be nil, except for Banner, which explicitly accepts nil to mean "always shown".

#### Banner and ErrorView

##### Banner: an in-page notice

Banner sits in the normal layout, and your app decides where. open=false hides it. When open is bound or Close options are provided, a close button is shown by default; CloseButton=false hides it explicitly. With a nil state, the banner is always shown; once closing is explicitly enabled, it uses internal state and reappears after being unmounted. To keep it closed across launches, your app stores that state; browser localStorage is not used.

Color defaults to Primary. With Neutral, the background uses InvertedBackground and the text uses Inverted. The title and the native-extension Description each stay on one line and are truncated. When the window is at least 1024 DIP wide, the content is centered between the side regions; when narrower, it starts from the leading edge. The Body container uses 16/24/32 DIP of horizontal padding depending on the window width and is limited by Theme.ContainerWidth. The minimum height is 48 DIP by default, and the font and icons follow the text scale.

Actions is a synchronous builder, and Buttons inside it default to Neutral / XS. OnSelect receives clicks or keyboard activation on the whole banner, while inner buttons act independently; your app decides what navigation to perform. Close and CloseIcon configure the close button, and CloseLabel sets its name. The button defaults to MD / Neutral / Ghost and shows a 10% background on hover or focus. OnClose fires once when the user closes the banner, not when the app opens or closes it. In very narrow areas, your app should reduce the number of actions. Banner does not pin itself to the top of the window.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Banner | Root, Body, Header, Leading, Label, Description, Actions, Trailing, Close |

Tabs, Stepper, and Timeline accept global and instance Color and Size; Banner and Toast accept Color. Color on menu items and Timeline items is a per-item setting. Other styles are configured through the parts listed above.

### Options and defaults

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Banner | CloseButton controls visibility; Close and CloseIcon configure the close button. OnSelect is independent of the inner actions. OnClose fires only when the user closes the banner |

### Compatibility and known limitations

#### Accessibility and motion (excerpt)

Since MyGo v0.9.0, `ui.RoleStatus` is a polite, atomic live region. The library's own progress, error, empty, and loading messages keep stable identities: such a status is read once when its text changes and is silent when it first appears, clears, or rebuilds unchanged. Alert and Banner take their identity from the key the app gives them. CommandPalette also announces its progress and error messages once when they appear. FileUpload keeps empty status regions mounted, so adding a message to an exposed region is read once; hidden updates are silent, and mounting or exposing an existing message does not announce it as a change. Chart announces keyboard selection once and never reads pointer hover. Alert and Banner read their new title and description once when either changes. These rules are checked headlessly with `Tester.UseScreenReader`; what VoiceOver, NVDA, Narrator, and Orca actually say has not been verified.

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
	Open                        bool
	Closed, Actions, Selections int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var icon *ui.SVG

	icon = lucide.Info()

	var actions func(*ui.Context, nui.Theme)

	actions = func(c *ui.Context, inner nui.Theme) {
		nui.Button(c, inner, "Release notes", nui.ButtonOptions{}).OnClick(func() { s.Actions++ })
	}

	nui.Banner(c.Key("banner"), t, &s.Open, nui.BannerOptions{Title: "An update is available", Icon: icon, Actions: actions, OnClose: func() { s.Closed++ }})
	nui.Button(c.Key("show"), t, "Show banner", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Open = true })
	ui.Textf(c, "Selections: %d · Actions: %d · Closed: %d", s.Selections, s.Actions, s.Closed)
}

func main() {
	s := &example{}

	s.Open = true
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Banner · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
