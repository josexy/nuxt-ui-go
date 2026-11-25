<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# ProgressGroup

Show multiple progress segments together on one track.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `ProgressGroup`
- Types: `ProgressGroupOptions`
- Theme key: `nui.ComponentProgressGroup` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Give each `ProgressSegment` a unique ID, `Value`, and legend `Label` (optional `Icon`, otherwise a colored dot). Use `Build` for custom legend rows, and `UI` on the group or individual segments for colors.

## Things to know

- Each segment clamps to `0–Max`; when the sum exceeds `Max`, segments scale proportionally while legend percentages stay relative to `Max`.
- `Max` defaults to `100` (explicit values must be finite and positive); vertical tracks default to 160 DIP when `Height <= 0`.
- Hiding the legend unmounts custom `Build` controls.

## API

```go
func ProgressGroup(c *ui.Context, t Theme, segments []ProgressSegment, opts ProgressGroupOptions) ui.Element
```

### ProgressGroupOptions

```go
type ProgressGroupOptions struct {
	Label       string
	Max         Optional[float64]
	Color       Color
	Size        Size
	HideLegend  bool
	Orientation Orientation
	Height      float32
	Decoration
}
```

### Related types

```go
type ProgressSegment struct {
	ID, Label string
	Value     float64
	Color     Color
	Icon      *ui.SVG
	UI        Styles
	Build     func(*ui.Context, Theme) // Optional synchronous legend content builder.
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Progress / ProgressGroup | Single or segmented progress with sizes, colors, and orientation. Progress supports animation | Progress: Root, Indicator. ProgressGroup also has Segment, List, Item, Leading, Label, Trailing |

#### Components and parts › Images, avatars, and icons (excerpt)

Every item in AvatarGroup and ProgressGroup needs a unique, non-empty ID. If AvatarGroup's Max is unset or ≤ 0, all avatars are shown. A positive Max limits the visible count and adds a same-sized `+n` avatar for the rest. The group's Size and Color only provide defaults for avatars that do not set their own. `Overlap: nui.Set(0)` turns off the overlap.

#### Components and parts › Progress and animation (excerpt)

**ProgressGroup** shows several segments on a single track. Each segment is first clamped to 0 through Max, and non-finite values are treated as 0. If the total is less than Max, the rest of the track stays empty; if it is more, segments are shortened proportionally, and the total reported to assistive technology is also capped at Max.

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

type example struct{}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var codeIcon, testIcon, docIcon *ui.SVG

	codeIcon, testIcon, docIcon = lucide.Settings(), lucide.Check(), lucide.Folder()

	var segmentUI nui.Styles

	nui.ProgressGroup(c.Key("progress"), t, []nui.ProgressSegment{{ID: "code", Label: "Code", Value: 45, Icon: codeIcon}, {ID: "tests", Label: "Tests", Value: 30, Color: nui.Info, Icon: testIcon, UI: segmentUI}, {ID: "docs", Label: "Docs", Value: 15, Color: nui.Warning, Icon: docIcon}}, nui.ProgressGroupOptions{Label: "Project progress", Max: nui.Set(float64(100)), Orientation: nui.Horizontal, Height: 160})
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
		mygo.NewWindow(mygo.WindowOptions{Title: "ProgressGroup · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
