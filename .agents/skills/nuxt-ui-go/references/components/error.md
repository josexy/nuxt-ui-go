<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Error (`nui.ErrorView`)

Explain a failure and provide a retry action.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `ErrorView`
- Types: `ErrorOptions`
- Theme key: `nui.ComponentError` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

ErrorView presents Status, Title and Description. The default action requires OnRetry. Clear=false hides it, Retry configures its button, and Actions replaces the action area. The component does not clear Nuxt error state or redirect automatically.

## Things to know

The icon is 40 DIP. Status, title, description and actions have separate style parts. Native embedding retains padding, a 480 DIP description limit and level-two heading semantics, without forcing full-window height. Error cleanup, retrying and navigation belong to the application.

## API

```go
// ErrorView is the Error component; Error is already the library's semantic color.
func ErrorView(c *ui.Context, t Theme, o ErrorOptions) ui.Element
```

### ErrorOptions

```go
type ErrorOptions struct {
	Status, Title, Description, RetryLabel string
	Icon                                   *ui.SVG
	OnRetry                                func()
	Clear                                  Optional[bool]
	Retry                                  *ButtonOptions
	Actions                                func(*ui.Context, Theme)
	Decoration
}
```

Types documented in other references: `ButtonOptions` → [button.md](button.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Banner and ErrorView

##### ErrorView: an error page

In Go, `nui.Error` is already a semantic color, so the error display function is called ErrorView. ErrorOptions provides Status, Title, Description, RetryLabel, Icon, Clear, Retry, OnRetry, Actions, and Decoration.

A 40 DIP icon is shown only when you pass Icon. The status code defaults to 16/24 DIP semibold Primary and is styled with StatusSlot. The title uses a 36 DIP font when the window is narrower than 640 DIP and 48 DIP otherwise. The description is 18/28 DIP in Muted. An empty title creates no text, and a description identical to the title is omitted.

The default retry button is LG / Primary / Solid with the label "Retry", but it is created only if you provide OnRetry. Clear defaults to true; false hides the default action. Retry configures the ButtonOptions. Actions replaces the entire action area and is not affected by Clear=false. The action area has a 32 DIP top margin and a 24 DIP gap by default, and no empty area is created when there are no actions.

When embedding an ErrorView in part of a page, reduce font sizes through parts such as UI[Label]. The native implementation keeps its padding, a 480 DIP maximum description width, and level-2 heading semantics, and does not force itself to fill the window; with a set height, the content is centered. The component does not call Nuxt's clearError or redirect. Your app handles panics, requests, navigation, and clearing errors.

### Options and defaults

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| ErrorView | Clear defaults to true, but the default button also requires OnRetry. Retry configures the button, Actions always takes precedence, and StatusSlot styles the status code |

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

type example struct{ Retries, Recoveries int }

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.ErrorView(c.Key("error"), t, nui.ErrorOptions{Status: "404", Title: "Project not found", Description: "This example does not make network requests.", OnRetry: func() { s.Retries++ }})
	ui.Textf(c, "Retry attempts: %d · Recoveries: %d", s.Retries, s.Recoveries)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Error · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
