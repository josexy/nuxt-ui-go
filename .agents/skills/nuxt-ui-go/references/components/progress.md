<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Progress

Show completed progress or indicate that work is ongoing.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.1.0 plus unreleased changes (MyGo v0.9.0)
- Functions: `Progress`
- Types: `ProgressOptions`
- Theme key: `nui.ComponentProgress` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Pass the completed amount; Max defaults to 100. Values are clamped to 0–Max and visual percentages are rounded to integers. Indeterminate ignores the amount and displays a moving half-track indicator. Use Value on the returned element for accessible work-unit text; compose a Text element when an on-screen percentage is needed.

## Things to know

An explicit Max must be finite and positive or the component panics. Non-finite amounts become zero. Unknown progress uses one of Nuxt's four two-second presets (Animation), shared with Table's loading bar; an out-of-range preset panics. Step arrays are not provided. The custom background affects Neutral only; Indicator.Background directly sets the fill color. This example does not start a task.

## API

```go
func Progress(c *ui.Context, t Theme, value float64, opts ProgressOptions) ui.Element
```

### ProgressOptions

```go
type ProgressOptions struct {
	Label                   string
	Color                   Color
	Size                    Size
	Max                     Optional[float64]
	Indeterminate, Vertical bool
	Orientation             Orientation
	Inverted                bool
	Animate                 Optional[bool]
	// Animation is how an indeterminate indicator moves: LoadingCarousel by
	// default, LoadingCarouselInverse, LoadingSwing or LoadingElastic.
	Animation LoadingAnimation
	Decoration
}
```

### Related types

```go
// LoadingAnimation is how an indeterminate bar moves, as Nuxt UI's Progress
// animation presets: Progress's indicator and Table's loading bar.
type LoadingAnimation uint8

const (
	// LoadingCarousel slides a half-length indicator from the start to the end.
	LoadingCarousel LoadingAnimation = iota
	// LoadingCarouselInverse slides it from the end to the start.
	LoadingCarouselInverse
	// LoadingSwing swings it to the end and back.
	LoadingSwing
	// LoadingElastic grows it from the middle to the whole track and back.
	LoadingElastic
)
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Progress / ProgressGroup | Single or segmented progress with sizes, colors, and orientation. Progress supports animation | Progress: Root, Indicator. ProgressGroup also has Segment, List, Item, Leading, Label, Trailing |

#### Components and parts › Progress and animation

**Skeleton**'s Width, Height, and Animate are native extensions. The default size is 160×16 DIP. The default height follows the text scale, but explicit sizes do not scale, and an explicit 0 stays 0. To fit the parent's width, call FillWidth on the returned element. Both Root.Radius and the final Style can set the corner radius, and the resulting radius clips the placeholder background and any attached content.

The pulse changes only the opacity of the custom-drawn background, once per `Motion.Pulse`; it goes on under reduced motion, as Nuxt UI's `animate-pulse` does, and stops with `Animate: nui.Set(false)` or at the Off motion level. Set its color with UI.Root.Background. Style.Background configures the underlying surface and cannot replace the resolved pulse color. Skeleton provides a localized loading name and a native status role, but not the full aria-busy/alert announcement semantics of the web. Your app decides when to show and hide placeholders and how to present the loaded result.

**Progress**'s Max defaults to 100. Explicitly setting it to a non-positive value, NaN, or Inf panics. A progress value of NaN or Inf is treated as 0 and then clamped to the range 0 to Max.

Determinate progress fills to the rounded integer percentage, while assistive technology receives the exact amount completed. The indeterminate indicator moves in a two-second cycle along one of Nuxt UI's presets, set with `Animation`: `LoadingCarousel` (the default, a half-length indicator sliding from the start to the end), `LoadingCarouselInverse`, `LoadingSwing` or `LoadingElastic`. The carousels run from the other end right to left and when Inverted, and all presets run along a vertical track; Table's loading bar uses the same presets. Under reduced motion the indicator spans the whole track and pulses its opacity instead, as Nuxt's `motion-reduce:animate-pulse`; with Animate=false or at the Off motion level it stays still, half the track's length, in the center. A determinate value moves to its new amount over 200 ms only at the Full motion level. An out-of-range Animation panics, as Table's LoadingAnimation does. Internally, `Now=-1 < Min=0` represents unknown progress, with an "In progress" value or description. Windows UIA clamps the unknown value to the minimum; how screen readers announce it has not been verified.

Horizontal progress fills from the start of the text direction. Vertical progress is 160 DIP tall by default and fills from the top. `Inverted` reverses the fill direction. Both `Orientation` and `Vertical` control orientation. Neutral uses InvertedBackground. The Indicator's Background overrides the custom-drawn fill color; other Indicator fields have no effect. Root is the track, and its style does not change the indicator's fixed pill shape.

Label is the accessible name, and Value on the returned element can describe the value in work units. To show a visible percentage or step text, compose Text yourself. The upstream status slot, the Max array of steps, and the other three CSS animation presets are not provided. Progress only displays the value you pass in; it does not change business state and has no value-change callback.

**ProgressGroup** shows several segments on a single track. Each segment is first clamped to 0 through Max, and non-finite values are treated as 0. If the total is less than Max, the rest of the track stays empty; if it is more, segments are shortened proportionally, and the total reported to assistive technology is also capped at Max.

The legend shows each segment's rounded percentage of Max. A legend is created as soon as any item has a Label, Icon, or Build. Without an Icon, an item shows a color dot. Build synchronously builds the whole legend row for that item and replaces the default icon, label, and percentage; controls inside it keep their identity through the stable ID. HideLegend unmounts the legend content.

Each item's UI can configure Segment, Item, Leading, Label, and Trailing, overriding the group-level UI of the same name. The group-level List styles the legend list and Indicator styles the whole track. Segment.Background can be any color. Neutral segments and dots use InvertedBackground, and Neutral icons use Highlighted. For sizes XS–XL, the legend font size is 12/14/14/14/16, icons are 12/16/16/16/20, and dots are 6/8/8/8/10 DIP, all following the text scale.

Horizontal groups follow RTL. Orientation defaults to Horizontal; in vertical layout, a Height of 0 or less uses 160 DIP. Segments are static and the group exposes overall progress to assistive technology. There is no per-segment progress role, built-in status text, or segment click event. You can add actions or descriptions with the legend's Build; scheduling the underlying work remains up to your app.

**When animations stop:** Continuous animations request the next frame only while a visible component actually draws. They stop when the component scrolls out of view, unmounts, has Animate turned off, or at the Off motion level; spinners and the Skeleton pulse go on under reduced motion. Button's colour transition follows `Motion.Transition` at Full and Reduced and is immediate at Off, and `Animate: nui.Set(false)` controls the loading spinner. Once a transition ends, no more frames are requested. Indeterminate Progress and Table's loading bar pulse under reduced motion. In MyGo v0.9.0 a colour transition builds the view in each frame while the colour moves (about ten builds for 150 ms), so an app with an expensive view can set `Motion.Transition` to 0 or turn animation off. Components do not use timer goroutines for animation.

### Advanced interaction, charts, and layout

#### FileUpload and resource ownership › The component handles selection and state; your app handles transfer (excerpt)

FileUploadState.Files stores file metadata, not open file handles. UploadFile contains Path, Name, MIME, Size, a status, a Progress from 0 to 1, Error, and an app-supplied Avatar. Add assigns stable IDs that do not collide with existing entries and ignores entries with an identical Path. If you assign Files directly, your app must provide unique IDs. The component does not load or decode preview images itself.

### Forms, dates, and data components

#### Table › Filtering and scope (excerpt)

EmptyText and BuildEmpty set the empty content below the header. Loading shows a loading bar and keeps the existing data. With no data, BuildLoading's content is shown if provided; otherwise the empty content is still shown. LoadingColor defaults to Primary. LoadingAnimation can be Carousel, CarouselInverse, Swing, or Elastic, the same presets as Progress's Animation (`TableLoadingAnimation` is an alias of `LoadingAnimation`). Under reduced motion the bar spans the header and pulses its opacity, as Nuxt's `motion-reduce:animate-pulse`; at the Off motion level it is a static full-width bar. Caption is the table's description for assistive technology, and BuildCaption can show a visible title. Height defaults to 320 DIP, including the title; the header stays fixed and the body scrolls inside.

### Overlays, menus, and navigation

#### Toast › Position, pausing, and style (excerpt)

The queue timer pauses while the pointer is over the notification area, or while focus is inside a toast and the window has focus. It keeps running while the window is in the background or hidden, even with focus left inside a toast. A window out of sight draws no frames (since MyGo v0.9.0 this includes a minimized window on Windows and on Linux under X11), so a toast whose time runs out meanwhile closes on the frame the window shows again, running its OnClose then, without showing. The default countdown bar pauses with it; hide it with `Progress: nui.Set(false)`. Persistent toasts have no countdown. Under reduced motion and at the Off motion level, the remaining time is still updated, but the bar shows it rounded to 250 ms steps, painted once per step, without interpolation. Toasts enter and exit with a 200 ms, 8 DIP fade and slide toward the matching screen edge; under reduced motion they fade in place, and at the Off motion level they switch immediately. An empty queue draws no panel.

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
	nui.Progress(c.Key("progress"), t, 60, nui.ProgressOptions{Label: "Progress", Max: nui.Set(float64(100)), Orientation: nui.Horizontal, Animation: nui.LoadingCarousel})
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Progress · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
