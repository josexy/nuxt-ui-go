<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Button

Run an action with support for icons, loading and disabled states.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Button`
- Types: `ButtonOptions`
- Theme key: `nui.ComponentButton` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Set button text and appearance with `Button`, then handle activation with `.OnClick(...)`. Keep values such as a click count outside the view function. Loading and Disabled both prevent activation.

## Things to know

An icon-only or avatar-only button requires an `AccessibleLabel`. Leading icons and leading loaders replace Avatar. Set loader colors through `UI[Leading].Foreground` or `UI[Trailing].Foreground`; TextColor in Style does not change their painted color.

## API

```go
func Button(c *ui.Context, t Theme, label string, opts ButtonOptions) ui.Element
```

### ButtonOptions

```go
type ButtonOptions struct {
	Variant Variant
	Color   Color
	Size    Size
	// Neutral selects the neutral palette; Color: Neutral is preferred.
	Neutral, Disabled, Loading, Block, Square bool
	AccessibleLabel                           string
	Leading, Trailing                         *ui.SVG
	Avatar                                    *AvatarOptions // Used when no leading icon or leading loading indicator is shown.
	LoadingIcon                               *ui.SVG
	LoadingTrailing                           bool // Place loading at the trailing edge unless Leading is set.
	Active                                    bool
	ActiveColor                               Color
	ActiveVariant                             Variant
	Animate                                   Optional[bool]
	UI                                        Styles
	Style                                     func(root, label, leading, trailing ui.Element)
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Style overrides › Styling a single part (excerpt)

Button has a `UI` field and `Style(root, label, leading, trailing ui.Element)` directly. Parts that were not created are passed as zero-value Elements. The callback configures only the elements of the current build; do not keep Elements, parts, or the Context for later. If you customize interaction, keep the control's name, role, and focus behavior.

#### Components and parts › Page structure and actions (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Button | A button with colors, variants, sizes, icons, an avatar, and loading and disabled states | Root, Label, Leading, Trailing |

Button's `Avatar: &nui.AvatarOptions{Alt: "Ada Lovelace"}` shows a leading avatar. A leading icon or a leading loading spinner takes precedence over it; a trailing loading spinner does not hide the avatar. If Avatar.Size is unset, button sizes XS–XL use avatars of 16, 16, 20, 20, and 24 DIP; an explicit Avatar.Size takes precedence. An avatar-only button still needs an AccessibleLabel. In this mode, the Leading part is the avatar's root element.

#### Components and parts › Indicators, separators, and loading (excerpt)

The Neutral/Solid background of Button, Badge, Alert, and Kbd uses Theme.InvertedBackground, and the text uses Inverted. On hover, Button uses that background at 90% opacity; in high contrast mode it stays opaque. Instance UI and Style can still override the default look.

#### Components and parts › Buttons and links

Button has six variants: `Solid`, `Outline`, `Soft`, `Subtle`, `Ghost`, and `LinkVariant`. Colors are `Primary`, `Secondary`, `Success`, `Info`, `Warning`, `Error`, and `Neutral`. The older `Neutral bool` field remains for compatibility; new code should use `Color: nui.Neutral`.

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| Default base height (DIP) | 24 | 28 | 32 | 36 | 40 |

Buttons grow taller when the font's line height is large. `Block` makes a button fill the available width and `Square` gives it a square layout. Icon-only buttons are square automatically and must have an `AccessibleLabel`, or they panic. Instance `UI` can still override the padding of a square button.

**Loading and disabled:** `Loading` blocks clicks and keyboard activation and keeps the label. By default it replaces the leading icon with a spinner, or adds one. `LoadingTrailing` moves the spinner to the trailing side, although an explicit Leading still takes precedence. `LoadingIcon` changes the icon. A button that had no icon may become wider while loading.

The spinner's color comes from the Root/Leading foreground style, so set it through `UI`. A `Style(...leading).TextColor` that runs last cannot change the already resolved color of the custom-drawn icon. `Active`, `ActiveColor`, and `ActiveVariant` control the active state.

MyGo draws disabled controls at 50% opacity, and the library adds no further transparency. A child with `Disabled: false` cannot override a disabled parent. Chip is purely decorative; if a count or online status matters, include it in the host's `AccessibleLabel` or `Alt`.

Chip's four corner positions are physical left and right and are not mirrored in RTL. Inset places the chip inside the host's bounds. When the content builder is nil, the chip is shown on its own without corner offsets. The text width is computed from the font, font size, font weight, and horizontal padding in Indicator.UI. Style runs after measurement, so if you change the text size there, set the width and height as well; a large custom font size usually also needs a taller chip.

**Links:** When Link's URL is non-empty, activating the link opens the address through `Services.OpenURL`. When the URL is empty, use `.OnClick` on the returned element for in-app navigation. If you set both URL and OnClick, keep MyGo's action override rules in mind. A focused link activates with Enter; as on the web, Space does not activate it and is left to the container around it. Use Button where Space should activate too.

#### Components and parts › Progress and animation (excerpt)

**When animations stop:** Continuous animations request the next frame only while a visible component actually draws. They stop when the component scrolls out of view, unmounts, has Animate turned off, or when reduced motion is enabled. Button's background transition follows `Motion.Transition/Reduce`, and `Animate: nui.Set(false)` controls the loading spinner; once a transition ends, no more frames are requested. Components do not use timer goroutines for animation.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Fields do not accept Button's Solid or LinkVariant. Color mainly controls the focus and selection accent; font size, corner radius, density, and light/dark colors come from the Theme.

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

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

type example struct{ Clicks int }

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Button(c.Key("button"), t, "Save changes", nui.ButtonOptions{AccessibleLabel: "Save changes"}).AlignSelf(ui.Start).OnClick(func() { s.Clicks++ })
	ui.Textf(c, "Clicked %d times", s.Clicks)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Button · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
