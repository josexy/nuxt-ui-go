<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Carousel

Drag or page through a set of cards.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Carousel`
- Types: `CarouselOptions`, `CarouselItem`, `CarouselState`
- Theme key: `nui.ComponentCarousel` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep CarouselState persistent and control position through Index. Give items stable IDs and synchronous Build callbacks. Arrows and Dots default to false; this example explicitly shows dots. Wrap loops buttons, keys and autoplay while drags remain bounded.

## Things to know

This native single-page carousel has no Embla drag-free, slide grouping, media query, physics duration or plugin API. Active gates external activity for autoplay and does not disable manual navigation, unlike Embla active. Build runs for the current item and its neighbors; neighbors are disabled. Start business work from actions. Autoplay respects window activity, visibility and reduced motion.

## API

```go
func Carousel(c *ui.Context, t Theme, s *CarouselState, items []CarouselItem, o CarouselOptions) ui.Element
```

### CarouselOptions

```go
type CarouselOptions struct {
	Label    string
	Height   float32
	Active   Optional[bool]
	Autoplay time.Duration
	// Wrap wraps explicit/automatic next and previous actions; drag stays bounded.
	Wrap                         bool
	Orientation                  Orientation
	Arrows                       Optional[bool]
	Dots                         bool
	Prev, Next                   *ButtonOptions
	PrevIcon, NextIcon           *ui.SVG
	WatchDrag, StopOnInteraction Optional[bool]
	OnChange                     func(int)
	Decoration
}
```

### CarouselItem

```go
type CarouselItem struct {
	ID, Label string
	Build     func(*ui.Context, Theme)
	Decoration
}
```

### CarouselState

```go
type CarouselState struct {
	Index  int
	Paused bool
	// contains filtered or unexported fields
}
```

Types documented in other references: `ButtonOptions` → [button.md](button.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### Carousel and Marquee

##### Carousel: switch content page by page

CarouselItem contains a stable ID, a Label, and a synchronous Build; your app can control CarouselState.Index and Paused. The component builds only the current item and its neighbors. Neighboring items are not interactive, but their Build still runs, so never start business operations unconditionally in a build callback.

| Action or option | Behavior |
| --- | --- |
| Dragging | Drags along the Orientation. Moving more than 20% of the viewport, or faster than 600 DIP/s, changes the page; on release, it snaps into place |
| Previous / Next, arrow keys, Home / End | Navigate by page, with ←→ horizontally and ↑↓ vertically |
| Arrows / Dots | Control the previous/next buttons and dot navigation |
| Wrap=true | Buttons, keyboard, and autoplay can wrap from the last item to the first; dragging still stops at the ends |
| WatchDrag | Default true. false registers no dragging while other navigation keeps working. Changing the orientation cancels the current drag |
| Prev / Next, PrevIcon / NextIcon | Button and icon options. Defaults to Outline / Neutral; the default arrow follows the orientation, and explicit icons are shown as is |
| StopOnInteraction | Default true. With autoplay on, the arrows, keyboard, and a completed drag set Paused=true, and Play resumes it. false keeps automatic resumption; dots only change position |

The current position is exposed only as the root element's assistive technology value, such as "Slide 1 of 3", using the slidePosition language key; no page counter is shown. Index is positional, so after reordering your app decides the index, and removing items clamps it to the valid range. Outside updates do not fire OnChange. If the focused element on the current page is removed, focus returns to the viewport. CarouselItem.Decoration can override the Item part; your app still stores business data such as drafts.

Autoplay is off by default; a positive Autoplay sets the interval. Hovering, focus inside the component, Disabled, Paused, Active=false, being invisible, an inactive window, or reduced motion all stop scheduling. Active only indicates whether native autoplay is active; it does not disable manual navigation, which differs from Embla's active. With Wrap=false, autoplay stops at the last item. When a temporary pause ends, the timer restarts. The play/pause button is a native extension.

OnChange fires once when user navigation or autoplay actually changes the index; drag and timer results produced during a build are reported after the build ends. Canceling a drag clears the offset, and an empty list also clears the drag state. Buttons are centered based on their actual size, and disabling at the bounds takes precedence over button options.

Go builds only the current page and its neighbors. It has no Embla free dragging, slide groups, media queries, physics-based animation duration, auto height, or plugin objects. Switching pages positions the content immediately. Dots use native selected buttons and do not claim to implement a full ARIA tablist.

##### Stopping animation when hidden

Neither component builds the view on every frame, and neither uses timer goroutines: Carousel autoplay asks for one build per interval, and Marquee moves its content in frames that only repaint, building about once a second. Visibility is what showed in the previous frame within the window and the scroll containers and clips around the component, so content scrolled out of a ScrollArea, clipped away, or kept mounted in a hidden panel (Tabs or Collapsible with KeepMounted) stops by itself. While a window is out of sight, MyGo draws no frames for it, so both stand still and continue from where the clock puts them when it shows again. Since MyGo v0.9.0 that covers a hidden, minimized, other-Space, or fully covered window on macOS and a minimized window on Windows and on Linux under X11. A window covered by other windows on Windows or Linux, or any window under Wayland, keeps drawing, and Marquee keeps moving there. Set Active=false for content your app hides by other means, such as covering it with another element, or stop building the component.

After unmounting, no more frames are requested and no paused time accumulates.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| Carousel | CarouselState | Root, Body, Item, Previous, Next, Actions, List, Indicator, Trigger, EmptySlot |

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| Carousel, FileUpload | Carousel: Arrows=false. FileUpload: Layout=UploadGrid |

#### Common options by component › Sliders, tabs, and accordions (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Carousel | Arrows shows the previous/next buttons and Dots enables dot navigation. The dots' page names follow the current language |

### Forms, dates, and data components

#### Table › Filtering and scope (excerpt)

EmptyText and BuildEmpty set the empty content below the header. Loading shows a loading bar and keeps the existing data. With no data, BuildLoading's content is shown if provided; otherwise the empty content is still shown. LoadingColor defaults to Primary. LoadingAnimation can be Carousel, CarouselInverse, Swing, or Elastic, and a static bar is used with reduced motion. Caption is the table's description for assistive technology, and BuildCaption can show a visible title. Height defaults to 320 DIP, including the title; the header stays fixed and the body scrolls inside.

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
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Carousel nui.CarouselState
	Drafts   [3]string
	Hidden   [3]bool
	Changes  int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	ids := []string{"one", "two", "three"}
	titles := []string{"Design", "Build", "Share"}
	descriptions := []string{"Start with a focused idea.", "Compose native Go components.", "Publish your work."}
	items := []nui.CarouselItem{}
	for i, id := range ids {
		if s.Hidden[i] {
			continue
		}
		items = append(items, nui.CarouselItem{ID: id, Label: titles[i], Build: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, titles[i]).FontSize(18)
			ui.Text(c, descriptions[i])
			nui.Input(c.Key("draft"), t, &s.Drafts[i], nui.InputOptions{Label: "Preview " + titles[i] + " draft", Placeholder: "Keep a draft"})
			nui.Button(c.Key("remove"), t, "Remove "+titles[i], nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { s.Hidden[i] = true })
		}})
	}
	o := nui.CarouselOptions{Label: "Slides", Height: 220, Orientation: nui.Horizontal, Dots: true, WatchDrag: nui.Set(true), StopOnInteraction: nui.Set(true), Active: nui.Set(true), OnChange: func(int) { s.Changes++ }}
	o.UI = nui.Styles{}
	if o.Orientation == nui.Vertical {
		o.Height = 280
	}
	if o.Arrows.Set {
		if o.Arrows.Value {
			if o.Orientation == nui.Vertical {
				o.UI[nui.Item] = nui.PartStyle{PaddingY: nui.Set(float32(64))}
			} else {
				o.UI[nui.Item] = nui.PartStyle{PaddingX: nui.Set(float32(64))}
			}
		}
	}

	direction := ui.LeftToRight

	ui.Column(c.Key("direction")).Direction(direction).Disabled(false).Children(func() { nui.Carousel(c.Key("carousel"), t, &s.Carousel, items, o) })
	nui.Button(c.Key("restore"), t, "Restore slides", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Hidden = [3]bool{}; s.Carousel.Index = 0 })
	ui.Textf(c, "Index: %d · Changes: %d · Paused: %t", s.Carousel.Index, s.Changes, s.Carousel.Paused)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Carousel · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
