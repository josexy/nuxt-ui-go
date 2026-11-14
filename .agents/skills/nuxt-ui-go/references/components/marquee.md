<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Marquee

Continuously scroll text or other display-only content.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Marquee`
- Types: `MarqueeOptions`
- Theme key: `nui.ComponentMarquee` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Repeat defaults to four copies; explicit zero builds no content and a single copy may leave an empty interval. Speed defaults to 32 DIP/s and Gap to 64 DIP. Keep the callback presentation-only; Body styles each content copy.

## Things to know

Repeat is the callback count: avoid inputs, buttons and business side effects. Focus, pause, Active=false, disabled ancestors, invisibility (including being scrolled or clipped out of view, or a hidden kept-mounted panel) and reduced motion stop animation. The content moves in frames that only repaint, so the view builds about once a second, and MarqueeState.Offset is the phase as of the last build. Native speed uses DIP/s, while upstream CSS defaults to 20 seconds per cycle. Repeated content still requires target screen-reader validation.

## API

```go
func Marquee(c *ui.Context, t Theme, s *MarqueeState, o MarqueeOptions, content func()) ui.Element
```

### MarqueeOptions

```go
type MarqueeOptions struct {
	Label              string
	Vertical, Reverse  bool
	Orientation        Orientation
	PauseOnHover       bool
	Speed, Gap, Height float32
	Active             Optional[bool]
	// Repeat is the number of content copies. Unset uses four; explicit zero is empty.
	Repeat Optional[int]
	// Overlay fades the content in and out at both edges; defaults to true.
	Overlay Optional[bool]
	// ContentKey resets the phase when content changes. Children must be presentation-only.
	ContentKey string
	Decoration
}
```

### Related types

```go
type MarqueeState struct {
	// Offset is the content's phase within one period of it and its gap, as
	// of the last build: between builds the content moves in frames that only
	// paint. Set it to move the content there.
	Offset float32
	Paused bool
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### Carousel and Marquee

##### Marquee: continuously scrolling content

Marquee is for decorative content. It supports Orientation/Vertical, Reverse, Speed in DIP/s, Gap, Height, and pausing with Space. Speed defaults to 32 DIP/s, Gap to 64 DIP scaled by the theme, and Height to 44 DIP. These floating-point values must be finite, and non-positive values use the defaults. Speed differs from upstream's CSS cycle of 20 seconds. To pause, use Paused or Active.

When Repeat is unset, the content is repeated 4 times; an explicit `nui.Set(0)` builds no content, and negative values are invalid. Each copy is at least as long as the container in the direction of motion, centered on the other axis, with its items evenly distributed, so even short content scrolls; a single copy may show a gap as it loops. Overlay fades to the background color over the outer third at each end by default; `Overlay: nui.Set(false)` turns the fade off.

When ContentKey changes, scrolling restarts from the beginning. PauseOnHover defaults to false; when enabled, hovering pauses the motion. Focus, Paused, Active=false, a disabled parent, and reduced motion also stop it. Horizontal motion follows RTL. An offset set from outside is normalized to the current content cycle, so the position stays valid when the container shrinks; a non-finite offset resets to 0.

The content moves in frames that only repaint, without building the view: Marquee builds again about once a second, to wrap the content back within its copies and to check whether it still shows. MarqueeState.Offset is therefore the phase as of the last build; read it in a build. Assistive technology also sees the copies where the last build placed them.

Root styles the container, and Body styles each copy. Body.Gap changes only the spacing between items inside a copy; copies are still separated by Gap. The usual override order of theme, instance UI, and the final Style applies.

**The content callback runs Repeat times on every build, so put only display elements in it.** Do not put business buttons, inputs, or side effects there. The underlying layer has no API for hiding a whole subtree from assistive technology, so verify how repeated content is read with the screen readers you target.

##### Stopping animation when hidden

Neither component builds the view on every frame, and neither uses timer goroutines: Carousel autoplay asks for one build per interval, and Marquee moves its content in frames that only repaint, building about once a second. Visibility is what showed in the previous frame within the window and the scroll containers and clips around the component, so content scrolled out of a ScrollArea, clipped away, or kept mounted in a hidden panel (Tabs or Collapsible with KeepMounted) stops by itself, as it does in a minimized or covered window. Set Active=false for content your app hides by other means, such as covering it with another element, or stop building the component.

After unmounting, no more frames are requested and no paused time accumulates.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| Marquee | MarqueeState | Root, Body |

### Compatibility and known limitations

#### Accessibility and motion (excerpt)

| Area | Current limitation |
| --- | --- |
| Marquee | How repeated content is presented to screen readers still needs verification |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Marquee    nui.MarqueeState
	Generation int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	ui.Row(c).Wrap().Gap(8).Children(func() {
		nui.Button(c, t, "Pause / resume", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Marquee.Paused = !s.Marquee.Paused })
		nui.Button(c, t, "Replace content", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Generation++ })
	})
	direction := ui.LeftToRight

	var styles nui.Styles

	ui.Column(c.Key("direction")).Direction(direction).Disabled(false).Children(func() {
		nui.Marquee(c.Key("marquee"), t, &s.Marquee, nui.MarqueeOptions{Label: "Technologies", Height: 160, Speed: float32(32), Repeat: nui.Set(int(4)), Orientation: nui.Horizontal, Overlay: nui.Set(true), Active: nui.Set(true), ContentKey: fmt.Sprint(s.Generation), Decoration: nui.Decoration{UI: styles}}, func() {
			nui.Badge(c.Key("release"), t, fmt.Sprintf("Release %d", s.Generation), nui.BadgeOptions{Variant: nui.Soft, Size: nui.LG})
			nui.Badge(c.Key("go"), t, "Go / MyGo", nui.BadgeOptions{Variant: nui.Outline, Size: nui.LG})
		})
	})
	ui.Textf(c, "Paused: %t · Content: %d", s.Marquee.Paused, s.Generation)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Marquee · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
