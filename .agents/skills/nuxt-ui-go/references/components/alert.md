<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Alert

Display an inline message, warning or error.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `Alert`
- Types: `AlertOptions`
- Theme key: `nui.ComponentAlert` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Title, Description and Icon describe the status. Avatar takes precedence over Icon and defaults to 44 DIP. Actions builds controls below the body by default; Horizontal moves them beside it.

## Things to know

OnClose reports a close request; the application owns visibility. This example hides the notice on request and retains action counts when reopened. Close accepts ButtonOptions; its defaults are MD, Neutral and LinkVariant, with an X icon.

## API

```go
func Alert(c *ui.Context, t Theme, opts AlertOptions) ui.Element
```

### AlertOptions

```go
type AlertOptions struct {
	Title, Description string
	Color              Color
	Variant            Variant
	Icon               *ui.SVG
	Avatar             *AvatarOptions // Takes precedence over Icon; defaults to 44 DIP.
	Actions            func()
	Orientation        Optional[Orientation] // Defaults to Vertical.
	OnClose            func()
	Close              *ButtonOptions // A non-nil value also displays the close button without OnClose.
	CloseIcon          *ui.SVG
	Decoration
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `ButtonOptions` → [button.md](button.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Alert | A message. Defaults to Solid; supports four variants, an icon, an avatar, Actions, and close options | Root, Body, Leading, Label, Description, Actions, Trailing |

Alert's Avatar takes precedence over Icon. The avatar defaults to a 44 DIP diameter; an explicit Avatar.Size overrides it. Actions sit below the body by default with a 10 DIP top margin; Horizontal places them beside the body. Switching the orientation or the leading content keeps the identity of the action controls that remain. Body is the text container, Actions is the action row, and Trailing is the close button.

#### Components and parts › Indicators, separators, and loading (excerpt)

The Neutral/Solid background of Button, Badge, Alert, and Kbd uses Theme.InvertedBackground, and the text uses Inverted. On hover, Button uses that background at 90% opacity; in high contrast mode it stays opaque. Instance UI and Style can still override the default look.

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
	Hidden           bool
	Reviewed, Closed int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var closeOptions *nui.ButtonOptions
	var onClose func()

	closeOptions = &nui.ButtonOptions{}
	onClose = func() { s.Hidden = true; s.Closed++ }

	nui.Button(c.Key("reopen"), t, "Reopen notice", nui.ButtonOptions{Variant: nui.Outline, Size: nui.XS}).OnClick(func() { s.Hidden = false })
	if !s.Hidden {
		nui.Alert(c.Key("alert"), t, nui.AlertOptions{Title: "Changes saved", Description: "Your changes are ready to review.", Icon: lucide.Info(), Orientation: nui.Set(nui.Vertical), Close: closeOptions, OnClose: onClose, Actions: func() {
			nui.Button(c.Key("review"), t, "Review changes", nui.ButtonOptions{Size: nui.XS, Color: nui.Neutral, Variant: nui.Subtle}).OnClick(func() { s.Reviewed++ })
		}})
	}
	ui.Textf(c, "Reviewed %d times; closed %d times", s.Reviewed, s.Closed)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Alert · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
