<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# AvatarGroup

Arrange several avatars and show how many remain hidden.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `AvatarGroup`
- Types: `AvatarGroupOptions`
- Theme key: `nui.ComponentAvatarGroup` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Give each AvatarItem a stable, unique user ID and avatar options. Max limits the visible avatars and represents the rest as `+n`. Size and Color supply defaults for items without their own settings.

## Things to know

An unset Max or a value at or below zero shows all avatars. Use user IDs rather than positions in a sortable list. Item targets each avatar root, including overflow. Item.BorderWidth and BorderColor control the outer separator ring; explicit BorderWidth=0 removes it.

## API

```go
func AvatarGroup(c *ui.Context, t Theme, items []AvatarItem, opts AvatarGroupOptions) ui.Element
```

### AvatarGroupOptions

```go
type AvatarGroupOptions struct {
	Max     Optional[int]
	Size    Size
	Color   Color
	Overlap Optional[float32]
	Decoration
}
```

### Related types

```go
type AvatarItem struct {
	ID     string
	Avatar AvatarOptions
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Avatar / AvatarGroup | A single avatar or a group of overlapping avatars. Group items use stable IDs, and the visible count can be limited | Avatar: Root, Body, Label, Leading. Group: Root, Item |

#### Components and parts › Images, avatars, and icons (excerpt)

Every item in AvatarGroup and ProgressGroup needs a unique, non-empty ID. If AvatarGroup's Max is unset or ≤ 0, all avatars are shown. A positive Max limits the visible count and adds a same-sized `+n` avatar for the rest. The group's Size and Color only provide defaults for avatars that do not set their own. `Overlap: nui.Set(0)` turns off the overlap.

AvatarGroup's Item part styles both ordinary and overflow avatars and is applied after each avatar's own style. Item.BorderWidth and BorderColor control the ring that separates avatars; an explicit BorderWidth of 0 removes it. Reordering keeps each avatar's identity through its stable AvatarItem.ID.

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

type example struct{ Reverse bool }

func (s *example) view(c *ui.Context, t nui.Theme) {
	items := []nui.AvatarItem{{ID: "ada", Avatar: nui.AvatarOptions{Alt: "Ada Lovelace"}}, {ID: "grace", Avatar: nui.AvatarOptions{Alt: "Grace Hopper"}}, {ID: "lin", Avatar: nui.AvatarOptions{Alt: "Lin Chen"}}}
	if s.Reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}
	var styles nui.Styles

	nui.AvatarGroup(c.Key("avatars"), t, items, nui.AvatarGroupOptions{Max: nui.Set(2), UI: styles})
	nui.Button(c.Key("reverse"), t, "Reverse avatars", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).OnClick(func() { s.Reverse = !s.Reverse })
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
		mygo.NewWindow(mygo.WindowOptions{Title: "AvatarGroup · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
