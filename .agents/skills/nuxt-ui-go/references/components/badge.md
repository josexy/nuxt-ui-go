<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Badge

Show a status, category or count as a small label.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Badge`
- Types: `BadgeOptions`
- Theme key: `nui.ComponentBadge` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Pass a short label to `Badge` and choose Color, Variant and Size. Leading and Trailing add icons before or after the text.

## Things to know

Badge displays information; use Button for actions. An empty label enables Square automatically. Unset avatar Size follows the badge; explicit avatar size takes precedence.

## API

```go
func Badge(c *ui.Context, t Theme, label string, opts BadgeOptions) ui.Element
```

### BadgeOptions

```go
type BadgeOptions struct {
	Color             Color
	Variant           Variant
	Size              Size
	Square            bool
	Leading, Trailing *ui.SVG
	Avatar            *AvatarOptions // Used when Leading is nil.
	Decoration
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Badge | A short label. Supports Solid/Outline/Soft/Subtle, colors, sizes, Square, icons on either side, and a leading avatar | Root, Label, Leading, Trailing |

When a Badge's label is empty, no text is created, no extra gap is left, and the padding becomes equal on all sides. Inside a FieldGroup or FormField, a Badge inherits the size and takes part in the group's corner rounding. If Avatar.Size is unset, badge sizes XS–XL use avatars of 16, 16, 16, 20, and 20 DIP. An explicit Leading icon takes precedence over the avatar.

#### Components and parts › Indicators, separators, and loading (excerpt)

The Neutral/Solid background of Button, Badge, Alert, and Kbd uses Theme.InvertedBackground, and the text uses Inverted. On hover, Button uses that background at 90% opacity; in high contrast mode it stays opaque. Instance UI and Style can still override the default look.

### Inputs, selection, and fields

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

### Overlays, menus, and navigation

#### Tabs, Collapsible, and Accordion › Choose a component and its state (excerpt)

TabItem provides ID, Label, Icon, Avatar, Badge, BadgeOptions, Disabled, Content, and an item UI. Icon takes precedence over Avatar. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. Badge is a string, and `"0"` is shown. It defaults to SM / Neutral / Outline and can be overridden with BadgeOptions. The body callbacks of Tabs and Accordion receive a Context and a Theme; Collapsible's body receives a Theme and uses the outer `c`.

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

type example struct{}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Badge(c.Key("badge"), t, "Ready", nui.BadgeOptions{}).AlignSelf(ui.Start).Label("Badge preview")
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Badge · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
