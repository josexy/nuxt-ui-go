<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Popover

Open related content beside a button or other control.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Popover`
- Types: `PopoverOptions`
- Theme key: `nui.ComponentPopover` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Create an anchor and pass it with persistent open state to Popover. The content callback receives the local theme. Plain text inherits final Panel typography, color and spacing while native widgets keep that theme. Bind click-mode toggling in the caller.

## Things to know

Unset Width follows content; default MaxHeight is 320 DIP without automatic scrolling. MatchAnchor sets a minimum width and Label is an accessible name. MyGo handles collisions with a 4 DIP window margin, versus upstream collisionPadding 8 px. Modal, arrow, portal and virtual anchors are not exposed. An outside button keeps its focus; closing from within returns to the current anchor.

## API

```go
// Popover is non-modal: outside presses dismiss it and continue to their target.
// The anchor and panel are build values; open is caller-owned window state.
func Popover(c *ui.Context, t Theme, anchor ui.Element, open *bool, opts PopoverOptions, content func(Theme)) ui.Element
```

### PopoverOptions

```go
type PopoverOptions struct {
	Label                    string
	Placement                Placement
	Width, MaxHeight, Offset Optional[float32]
	MatchAnchor              bool
	Mode                     PopoverMode
	OpenDelay, CloseDelay    time.Duration
	Dismissible              Optional[bool]
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

```go
type PopoverMode uint8

const (
	PopoverClick PopoverMode = iota
	PopoverHover
)
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Mode` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### Popover and ScrollArea

##### Popover: show content next to a control

```go
anchor := nui.Button(c, theme, "Help", nui.ButtonOptions{Variant: nui.Outline}).
    OnClick(func() { app.Help = !app.Help })
nui.Popover(c, theme, anchor, &app.Help, nui.PopoverOptions{
    Label: "Help panel",
    Width: nui.Set(float32(300)),
    Placement: nui.BottomStart,
}, func(inner nui.Theme) {
    ui.Text(c, "Explanation").TextColor(inner.Muted)
    nui.Button(c, inner, "Close", nui.ButtonOptions{}).
        OnClick(func() { app.Help = false })
})
```

The anchor must be an element from the current build, and open must be a non-nil pointer you keep between builds. The popover closes if the anchor is invalid or disabled. The return value is this build's panel, or a zero-value Element when closed; do not keep it across builds.

Popover is non-modal: an outside click closes it and still reaches its target. Escape closes it and returns focus to where it was opened. Tab is not trapped inside the panel. Build the content with the Theme passed to the callback so the local theme is preserved.

| Option | Default behavior |
| --- | --- |
| Placement | BottomStart. Supports start, center, and end on each of the four sides, for 12 positions |
| Width / MatchAnchor | Width follows the content by default; you can set a width or match the anchor |
| MaxHeight | 320 DIP. Limits only the height and does not add scrolling |
| Offset | 8 DIP from the anchor |
| Decoration | Use the Panel part to style the appearance |

For top and bottom placements, start and end follow the text direction; left and right placements use physical screen directions. MyGo's AttachTo keeps the popover clear of the window edges.

By default, a generic Popover has the theme background, a 1 DIP border, a 6 DIP corner radius, an 8 DIP content gap, a shadow, and 16 DIP of padding. To make content reach the edges, set the padding of UI[Panel] to 0. Select, InputMenu, and menus use the same base style but usually with a 240 DIP maximum height, 4 DIP of padding, no gap, and at least the trigger's width. For the menus' own height options, see [DropdownMenu and ContextMenu](dropdown-menu.md).

The Panel's font, text color, line height, and Gap apply directly to plain content. Native child controls keep using the local theme passed to them, and their own explicit text styles take precedence. Changing the anchor or panel size keeps the content's identity, drafts, and selection. If focus is inside the popover when it closes, focus returns to the current anchor; if it closes because the user clicked another control, that control keeps focus.

The entrance animation respects reduced motion, and changing that preference while the popover is open does not replay it. Closing unmounts the popover immediately; there is no exit animation as in upstream. Collision handling uses MyGo's 4 DIP edge margin, while upstream defaults to 8 px. When there is no room on either side, a side placement can still overflow the window; constrain the content size or use a top or bottom placement.

On open, the panel fades in from 95% scale within 100 ms, scaling from the side nearest the anchor. With reduced motion, it appears immediately. Closing unmounts the panel at once and restores focus, with no exit animation. A selection list's own Panel override is applied last, and explicit padding in Popup.UI also takes effect.

### Options and defaults

#### Common options by component › Overlays and notifications (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Popover | Mode defaults to PopoverClick and can be PopoverHover. OpenDelay and CloseDelay apply only in hover mode; they are `time.Duration` values that default to 0 |

Popover's Dismissible defaults to true and controls closing with Escape or an outside click; the app can always change `open` directly. In hover mode, the pointer can cross the gap between the anchor and the panel. After Escape closes the popover, leaving the pointer where it is does not reopen it right away.

### Advanced interaction, charts, and layout

#### NavigationMenu and Sidebar › What navigation and layout each handle (excerpt)

In compact mode, Tooltip and the side Popover are on by default, as in the native implementation. Turn either off with `NavigationOptions.Tooltip/Popover = nui.Set(false)`. Per-item settings override the global ones, and TooltipOptions and PopoverOptions configure their appearance and behavior. In horizontal menus, tooltips are enabled per item through Tooltip.

The side Popover opens on hover (including across the gap from the anchor), on click, or with the arrow key pointing toward the panel; opening it with the keyboard focuses the first available child. Escape, the opposite arrow key, or choosing an item closes it and restores focus to the trigger. Children inside it can expand further. PopupPlacement changes the default position, which is useful for a sidebar on the right.

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns › FooterColumns (excerpt)

FooterColumns takes columns with stable, non-empty IDs, and link IDs must be unique within each column. A non-empty Title is shown as a level-3 heading. Items use NavigationItem's ID, Label, Icon, Active, Disabled, OnSelect, and Decoration. The other fields (groups, descriptions, nested children, avatars, badges, Tooltip, and Popover) are for NavigationMenu and are not used by this component.

### Forms, dates, and data components

#### Calendar, InputDate, and InputTime › InputDate: type a date (excerpt)

ShowCalendar defaults to false. When enabled, it adds a calendar button and a Popover that closes after a date is chosen. HideCalendar takes precedence over ShowCalendar. The Calendar's appearance, navigation, and part options apply to the popup, while Range is determined by which input function you call; a custom Label is kept. For a single date, Calendar.OnChange reports only calendar actions; segment editing is reported through OnValueChange. The popup closes if the input, its field, or an ancestor becomes disabled.

### Overlays, menus, and navigation

#### Modal and Slideover › Closing and focus (excerpt)

Tab focus is trapped inside the modal content, and an inner Select or Popover handles Escape first. By default, the first focusable control receives focus, usually the close button. To focus a field first, request that field's Handle in the action that opens the dialog. If the Header or Footer is very tall, your app must constrain the content size to leave room for the body.

#### Modal and Slideover › Content, size, and appearance (excerpt)

The Modal panel has an 8 DIP corner radius, and the overlay uses the Elevated color at 75% opacity. On open, the panel fades in from 95% scale over 200 ms while the overlay fades in with it. With Transition=false or reduced motion, it appears immediately. Like Popover, Modal closes immediately without an exit animation. There is no upstream full-overlay scrolling mode, non-modal switch, or keep-mounted option.

#### Tooltip (excerpt)

The tooltip hides when the pointer leaves, the trigger is clicked, the user presses outside, or Escape is pressed. It also closes when it is Disabled, when the anchor is invalid or disabled, or when both the text and Kbds are empty. It opens with a 100 ms scale-and-fade animation, appears immediately with reduced motion, and disappears immediately on close. A Go Tooltip does not receive pointer input or focus, and it has no upstream hoverable content, arrow, or arbitrary content slot. For multiline or clickable content, use Popover.

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
	Open  bool
	Draft string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	t.Motion.Reduce = t.Motion.Reduce || false
	parts := nui.Styles{}

	options := nui.PopoverOptions{Width: nui.Set(float32(240)), Decoration: nui.Decoration{UI: parts}}
	ui.Column(c.Key("popover-stage")).Fill().Center().Children(func() {
		var first ui.Element
		ui.Row(c.Key("anchors")).Gap(8).Wrap().Children(func() {
			first = nui.Button(c.Key("trigger"), t, "Open details", nui.ButtonOptions{AccessibleLabel: "Open details"}).OnClick(func() {
				if options.Mode == nui.PopoverClick {
					s.Open = !s.Open
				}
			})
			nui.Button(c.Key("alternate"), t, "Alternate anchor", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				if options.Mode == nui.PopoverClick {
					s.Open = !s.Open
				}
			})
		})
		anchor := first

		nui.Popover(c.Key("popover"), t, anchor, &s.Open, options, func(inner nui.Theme) {
			ui.Text(c, "Additional information")
			nui.Input(c.Key("draft"), inner, &s.Draft, nui.InputOptions{Label: "Popover draft"}).FillWidth()
			nui.Button(c.Key("close"), inner, "Close", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Open = false })
		})
	})
}

func main() {
	s := &example{}

	s.Draft = "Editable content"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Popover · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
