<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# ScrollArea

Scroll content that exceeds the visible area.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `ScrollArea`
- Types: `ScrollAreaOptions`
- Theme key: `nui.ComponentScrollArea` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Set Height, Fill or Grow on the returned wrapper to define the viewport. Direction selects vertical, horizontal or both axes. Give each instance its own ui.ScrollState: X / Y are offsets and MaxX / MaxY are the laid-out bounds.

## Things to know

All content is built. Nuxt ScrollArea items / virtualization, fading edge shadows and infinite loading are not exposed; use Listbox or Table for virtual lists. Upstream scroll events come from the virtualizer; ui.ScrollState.Scrolling() only reports programmatic smooth scrolling. A transparent thumb retains its hit area.

## API

```go
// ScrollArea uses a themed wrapper because MyGo reads scrollbar colors from its theme.
// Size the returned wrapper; it is intentionally not a transparent provider.
func ScrollArea(c *ui.Context, t Theme, opts ScrollAreaOptions, content func()) ui.Element
```

### ScrollAreaOptions

```go
type ScrollAreaOptions struct {
	Direction      ScrollDirection
	ScrollbarColor Optional[ui.Color]
	ScrollbarWidth Optional[float32]
	State          *ui.ScrollState
	Decoration
}
```

### Related types

```go
type ScrollDirection uint8

const (
	ScrollVertical ScrollDirection = iota
	ScrollHorizontal
	ScrollBoth
)
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### Popover and ScrollArea

##### ScrollArea: give content a scrolling region

ScrollArea scrolls in one of three directions: ScrollVertical (the default), ScrollHorizontal, or ScrollBoth. You can pass a `*ui.ScrollState` to read the scroll position.

Each instance uses its own persistent state. X and Y are offsets in DIP; assign them directly or call ScrollTo. MaxX and MaxY are computed by layout. Out-of-range values are clamped, and when the direction changes, any axis that is no longer available resets to zero. In RTL, X is measured from the right edge. Scrolling() reports only programmatic smooth scrolling, not every wheel action.

**Set Height, Fill, or Grow on the returned outer element** to constrain the scroll area's size. Put content padding and gaps on the Body part. The parts are Root and Body.

Horizontal scrolling uses a row layout; vertical and two-way scrolling use a column layout. Root's corner radius clips the content, and changing the scrollbar style does not reset the offset. The root container draws the theme background by default; for a transparent background, set Root.Background explicitly.

ScrollbarColor defaults to BorderAccented, and ScrollbarWidth defaults to MyGo's 6 DIP. A width of 0 or less also uses the default and does not hide the scrollbar. An explicit transparent color only hides the drawing; the hit area remains. The native Scroll handles the wheel, dragging, and scroll bounds. When an inner area reaches its bound, unconsumed wheel input can reach the outer area.

ScrollArea supports nesting and two-way scrolling. It does not provide upstream's items/virtualize, edge fade shadows, infinite loading, or custom scrollbar interaction. All content is built every time, so use the virtual lists in Listbox or Table for long lists. Put long content in a Popover inside a ScrollArea yourself.

### Advanced interaction, charts, and layout

#### Carousel and Marquee › Stopping animation when hidden (excerpt)

Neither component builds the view on every frame, and neither uses timer goroutines: Carousel autoplay asks for one build per interval, and Marquee moves its content in frames that only repaint, building about once a second. Visibility is what showed in the previous frame within the window and the scroll containers and clips around the component, so content scrolled out of a ScrollArea, clipped away, or kept mounted in a hidden panel (Tabs or Collapsible with KeepMounted) stops by itself, as it does in a minimized or covered window. Set Active=false for content your app hides by other means, such as covering it with another element, or stop building the component.

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

type example struct{ Scroll ui.ScrollState }

func (s *example) view(c *ui.Context, t nui.Theme) {
	direction := nui.ScrollVertical
	radius := float32(0)

	radius = 16

	options := nui.ScrollAreaOptions{Direction: direction, State: &s.Scroll, ScrollbarWidth: nui.Set(float32(6)), Decoration: nui.Decoration{UI: nui.Styles{nui.Root: {Radius: nui.Set(radius)}}, Style: func(part nui.Slot, e ui.Element) {
		if part == nui.Body {
			e.Label("Scrollable records")
		}
	}}}
	nui.ScrollArea(c.Key("scroll"), t, options, func() {
		for i := range 30 {
			row := ui.Textf(c.Key(i), "Activity item %02d · Scrollable content", i+1).PaddingX(12).Background(t.Elevated)
			if options.Direction == nui.ScrollHorizontal {
				row.Width(180).Height(140)
			} else {
				row.Width(900).Height(40)
			}
		}
	}).Height(240).FillWidth()
	if position := options.State; position != nil {
		ui.Textf(c, "Scroll position: x=%.0f y=%.0f", position.X, position.Y)
		nui.Button(c.Key("end"), t, "Scroll to end", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { position.X, position.Y = position.MaxX, position.MaxY })
		nui.Button(c.Key("start"), t, "Scroll to start", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { position.X, position.Y = 0, 0 })
	} else {
		ui.Text(c, "Attach State to inspect offsets")
	}
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
		mygo.NewWindow(mygo.WindowOptions{Title: "ScrollArea · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
