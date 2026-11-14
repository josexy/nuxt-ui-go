<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Splitter

Resize two panels by dragging their divider.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Splitter`
- Types: `SplitterOptions`
- Theme key: `nui.ComponentSplitter` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Splitter binds the first pane size as float32 in DIP, or a percentage with Percent=true. Orientation or Vertical selects the axis. Min, Max and explicit KeyboardResizeBy use the bound unit; the unset keyboard step moves ten DIPs.

The visible 1 DIP stroke rounds to whole device pixels; the 8 DIP drag target is unchanged.

## Things to know

Disabled locks only the divider. Two panels and nested splitters are supported; coordinated multi-panel layouts, collapse/docking and auto-save are not provided. Value constraints apply to the model; panes retain the native 40-DIP minimum and shrink in very small layouts. Only the one-DIP line occupies layout space. An eight-DIP hit area overlaps the pane edges without separating them. SeparatorSlot and Indicator style the hit area and line.

## API

```go
// Splitter binds the first pane's size in DIPs, or a percentage when Percent is true.
// Disabled locks only the divider; disabling an ancestor also disables pane contents.
func Splitter(c *ui.Context, t Theme, value *float32, o SplitterOptions, first, second func()) ui.Element
```

### SplitterOptions

```go
type SplitterOptions struct {
	Label                       string
	Orientation                 Orientation
	Vertical, Percent, Disabled bool
	Min, Max                    Optional[float32]
	// KeyboardResizeBy uses the bound value's unit. Unset moves ten DIPs; zero locks arrow keys.
	KeyboardResizeBy Optional[float32]
	OnChange         func(float32)
	Decoration
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns

##### Splitter: split two panes

Splitter is built on MyGo's layout, focus, dragging, and shortcuts, and binds a `*float32`. The default unit is DIP. With Percent=true, the value is 0–100, a share of the space the two panes can use, not counting the divider. Orientation=Vertical or Vertical=true splits top and bottom. Horizontal layout and arrow keys are mirrored in RTL.

Min and Max use the same unit as the bound value. The default lower bound is 0, and the upper bound is 1e7 DIP or 100%. An explicit 0 is valid. OnChange fires once after a pointer or keyboard action actually changes the value; repeated input at a bound does not fire it. When your app assigns the value directly, it is silently clamped on the next build, and a non-finite model value is treated as 0 first. Constraints and steps must be finite, and the maximum cannot be lower than the effective minimum.

When KeyboardResizeBy is unset, each key press moves 10 DIP scaled by the theme. An explicit value uses the bound value's unit, and 0 turns off arrow-key resizing. Home and End go to the reachable bounds of the range. Disabled locks only the divider, and the panes stay editable; to disable the whole subtree, disable the parent container.

The divider highlights on hover or during a drag. Once the drag ends and the pointer is outside the hit area, the highlight clears and the divider still accepts keyboard shortcuts. Tabbing to the divider shows its focus ring.

Both panes keep a native minimum size of 40 DIP; in DIP mode, the first pane can explicitly be set smaller. When space runs out, the layout shrinks but the model value is kept, so a model value of 0 does not mean a collapsed pane. The layout reserves only 1 DIP for the divider line, and an 8 DIP drag area overlaps the pane edges without pushing the panes apart; when nested, horizontal and vertical lines meet without gaps. Root, Leading, Trailing, SeparatorSlot, and Indicator style the container, the two panes, the drag area, and the center line. The accessible Range and Step use the bound value's unit and the configured constraints rather than always using the underlying DIP.

Two panes and nesting are supported. Multi-pane coordination, collapsible snapping, autosave, and layout serialization are not. Your app manages the persistent model, and layout and orientation changes keep input state in the panes.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| Splitter | `*float32` | Root, Leading, Trailing, SeparatorSlot, Indicator |

### Compatibility and known limitations

#### Feature scope › Layout and navigation (excerpt)

| Area | Current behavior and limits |
| --- | --- |
| Carousel / Splitter | Carousel buttons, keyboard, and autoplay can loop, but dragging still stops at the ends. Active only gates autoplay. No full Embla plugin set or ARIA tablist. Splitter splits into exactly two panels |

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

type example struct {
	Ratio, InnerRatio     float32
	Draft, Notes          string
	Changes, InnerChanges int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	ui.Row(c).Wrap().Gap(8).Children(func() {
		nui.Button(c, t, "External maximum", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Ratio = 70 })
		nui.Button(c, t, "Reset splitter", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Ratio = 40; s.InnerRatio = 55; s.Changes = 0; s.InnerChanges = 0 })
	})
	direction := ui.LeftToRight

	var styles nui.Styles

	ui.Column(c.Key("direction")).Direction(direction).Disabled(false).Children(func() {
		nui.Splitter(c.Key("split"), t, &s.Ratio, nui.SplitterOptions{Label: "Panel divider", Percent: true, Orientation: nui.Horizontal, Min: nui.Set(float32(30)), Max: nui.Set(float32(70)), KeyboardResizeBy: nui.Set(float32(5)), Decoration: nui.Decoration{UI: styles}, OnChange: func(float32) { s.Changes++ }}, func() {
			ui.Column(c).Fill().Padding(12).Gap(8).Children(func() {
				ui.Text(c, "Files")
				nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Pane draft"})
			})
		}, func() {
			preview := func() {
				ui.Column(c).Fill().Padding(12).Gap(8).Children(func() { ui.Text(c, "Preview"); ui.Text(c, s.Draft) })
			}

			var inner nui.SplitterOptions
			inner.Label = "Nested panel divider"
			inner.Percent = true
			inner.Vertical = nui.Horizontal != nui.Vertical
			inner.Disabled = false
			inner.Min = nui.Set(float32(30))
			inner.Max = nui.Set(float32(70))
			inner.KeyboardResizeBy = nui.Set(float32(5))
			inner.UI = styles
			inner.OnChange = func(float32) { s.InnerChanges++ }
			nui.Splitter(c.Key("nested"), t, &s.InnerRatio, inner, preview, func() {
				ui.Column(c).Fill().Padding(12).Gap(8).Children(func() {
					ui.Text(c, "Notes")
					nui.Input(c.Key("notes"), t, &s.Notes, nui.InputOptions{Label: "Pane notes"})
				})
			}).Fill()
		}).Height(280).Border(1, t.Border)
	})
	ui.Textf(c, "Value: %.0f · Changes: %d", s.Ratio, s.Changes)

	ui.Textf(c, "Nested value: %.0f · Changes: %d", s.InnerRatio, s.InnerChanges)
}

func main() {
	s := &example{}

	s.Ratio = 40
	s.InnerRatio = 55
	s.Draft = "Editable panel"
	s.Notes = "Independent notes"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Splitter · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
