<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Tooltip

Show a short explanation on hover or keyboard focus.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Tooltip`
- Types: `TooltipOptions`
- Theme key: `nui.ComponentTooltip` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Pass an anchor from the current build, text and optional Kbds. Bind Open to persistent state. DelayDuration is an Optional[time.Duration]; explicit zero shows immediately. Text and formatted keys append to the anchor description.

## Things to know

Unset delay preserves MyGo’s 600ms default and a 400ms warm-switch window; Nuxt UI defaults are 700ms and 300ms. Go tips do not receive pointer input or focus and close when the pointer leaves the anchor. Hoverable content and arbitrary content slots are unavailable; use Popover for multiline or interactive content. Kbds remain visible at every window width.

## API

```go
func Tooltip(c *ui.Context, t Theme, anchor ui.Element, text string, o TooltipOptions) ui.Element
```

### TooltipOptions

```go
type TooltipOptions struct {
	Disabled      bool
	MaxWidth      Optional[float32]
	Placement     Optional[Placement]
	Offset        Optional[float32]
	Open          *bool
	Kbds          []string
	DelayDuration Optional[time.Duration]
	Decoration
}
```

### Related types

```go
type Placement uint8

const (
	BottomStart Placement = iota
	BottomEnd
	TopStart
	TopEnd
	RightStart
	LeftStart
	BottomCenter
	TopCenter
	RightCenter
	RightEnd
	LeftCenter
	LeftEnd
)
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Tooltip

Tooltip shows a short, non-interactive hint: `Tooltip(c, t, anchor, text, options)`. It responds to mouse hover and keyboard focus, and appends its text and the formatted Kbds to the anchor's Description, keeping any existing description. Kbds only display shortcut hints; they do not register shortcut actions.

| Behavior | Default |
| --- | --- |
| Position and offset | Below the anchor, centered, 8 DIP away |
| MaxWidth | 320 DIP; text longer than one line is truncated with an ellipsis |
| Appearance | 24 DIP tall, 4 DIP corner radius, 10 DIP horizontal padding, small shadow |
| Mouse | 600 ms delay when DelayDuration is unset. When moving between tooltips, a 400 ms window lets the next one appear quickly |
| Keyboard | Shown immediately on focus |

`DelayDuration` is an `Optional[time.Duration]`, for example `nui.Set(200 * time.Millisecond)`. An explicit 0 or a negative value shows the tooltip immediately. Leaving the anchor cancels the wait, and keyboard focus does not wait for the hover delay. Nuxt UI defaults to a 700 ms delay and a 300 ms skip window; Go keeps MyGo's native defaults.

The tooltip hides when the pointer leaves, the trigger is clicked, the user presses outside, or Escape is pressed. It also closes when it is Disabled, when the anchor is invalid or disabled, or when both the text and Kbds are empty. It opens with a 100 ms scale-and-fade animation, appears immediately with reduced motion, and disappears immediately on close. A Go Tooltip does not receive pointer input or focus, and it has no upstream hoverable content, arrow, or arbitrary content slot. For multiline or clickable content, use Popover.

Kbds use the SM size by default. The text and the shortcut are separated by `·`, and keys are 2 DIP apart. Unlike upstream, which hides shortcuts on narrow screens, Go always shows them. Panel.Gap sets the space between the text and the shortcut container, and Trailing styles the container of the shortcut and separator dot. The text is truncated to one line, but the shortcut does not shrink, so leave room for it in MaxWidth.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Tooltip | Panel, Label, Trailing |

### Options and defaults

#### Common options by component › Overlays and notifications (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Tooltip | Placement and Offset set the position. Open can bind the open state. Kbds displays shortcut keys and adds them to the anchor's description. DelayDuration sets the hover delay explicitly; 0 shows the tooltip immediately |

Even when Open is bound, Tooltip still responds to the hover delay, keyboard focus, blur, clicks, and Escape. When the app sets Open to false, the tooltip does not reopen just because the pointer or focus is still on the anchor. A new tooltip closes any previous controlled tooltip.

### Advanced interaction, charts, and layout

#### NavigationMenu and Sidebar › What navigation and layout each handle (excerpt)

In compact mode, Tooltip and the side Popover are on by default, as in the native implementation. Turn either off with `NavigationOptions.Tooltip/Popover = nui.Set(false)`. Per-item settings override the global ones, and TooltipOptions and PopoverOptions configure their appearance and behavior. In horizontal menus, tooltips are enabled per item through Tooltip.

#### Chart › Legends and data selection (excerpt)

Cartesian charts support pointing at the nearest X with the mouse, and Left, Right, Home, End, and Escape; Home and End go to the logical first and last data points. The description line at the bottom and the Tooltip show the raw values, and Tooltip=false hides only the tooltip. TooltipOptions configures settings such as position; the default is TopCenter with a 12 DIP offset and no delay. Pie and donut charts support pointing at slices and arrow keys, and show percentages of the visible total. The donut's center does not respond to slice pointing. Selection on dense data always looks up the raw values.

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns › FooterColumns (excerpt)

FooterColumns takes columns with stable, non-empty IDs, and link IDs must be unique within each column. A non-empty Title is shown as a level-3 heading. Items use NavigationItem's ID, Label, Icon, Active, Disabled, OnSelect, and Decoration. The other fields (groups, descriptions, nested children, avatars, badges, Tooltip, and Popover) are for NavigationMenu and are not used by this component.

### Forms, dates, and data components

#### Numbers, tags, PINs, and ratings › RangeSlider (excerpt)

Common options include Orientation, Inverted, Tooltip, MinStepsBetweenThumbs, Color, Size, OnChange, and OnCommit. Each thumb has its own focus and reports its own value and minimum/maximum names to assistive technology.

### Inputs, selection, and fields

#### Slider (excerpt)

Tooltip is a bool. When on, hovering or focusing the thumb shows the current value immediately, and Escape hides it. It has no full TooltipOptions or ComponentTooltip part configuration. On a single slider, changing UI.Root.PaddingX/PaddingY changes the active track area, and drawing, pointer math, and tooltip positioning all follow. RangeSlider keeps a fixed track inset of half a thumb.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct{ open bool }

func (s *example) view(c *ui.Context, t nui.Theme) {
	var delay nui.Optional[time.Duration]

	var keys []string

	keys = []string{"Ctrl", "K"}

	styles := nui.Styles{}

	options := nui.TooltipOptions{Open: &s.open, DelayDuration: delay, MaxWidth: nui.Set(float32(240)), Kbds: keys, Decoration: nui.Decoration{UI: styles}}
	nui.Button(c.Key("toggle"), t, "Toggle tooltip", nui.ButtonOptions{Disabled: options.Open == nil || options.Disabled, Variant: nui.Outline}).OnClick(func() {
		if options.Open != nil {
			*options.Open = !*options.Open
		}
	})
	anchor := nui.Button(c.Key("trigger"), t, "Hover or focus", nui.ButtonOptions{Variant: nui.Outline}).Description("Workspace action")
	nui.Tooltip(c.Key("tooltip"), t, anchor, "Search this workspace", options)
	if options.Open != nil {
		ui.Text(c, fmt.Sprintf("Tooltip open: %t", *options.Open))
	} else {
		ui.Text(c, "Uncontrolled tooltip")
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Tooltip · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
