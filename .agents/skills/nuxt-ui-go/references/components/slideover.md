<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Slideover

Open a detail or editing panel from a window edge.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Slideover`
- Types: `SlideoverOptions`
- Theme key: `nui.ComponentSlideover` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep a persistent open bool and keep calling Slideover during closing. Side selects a physical edge and Extent sizes its axis. Inset adds a 16 DIP margin and 8 DIP corners. Top and bottom panels fit content without Extent; an empty body still retains its padding.

## Things to know

Inherited Width, MaxHeight and Fullscreen do not affect Slideover layout. Exit retains modal focus and disables panel actions until unmount and focus restoration. Keep-mounted and non-modal modes are unavailable. Side panels default to 448 DIP within available space; Go does not apply the upstream small-screen full-width breakpoint.

## API

```go
// Slideover retains its modal scope through the exit, then unmounts to restore focus.
// Call it every frame, including while open is false.
func Slideover(c *ui.Context, t Theme, open *bool, o SlideoverOptions, content func(Theme)) ui.Element
```

### SlideoverOptions

```go
type SlideoverOptions struct {
	ModalOptions
	Side   Side
	Extent Optional[float32]
	Inset  bool
}
```

### Related types

```go
type Side uint8

const (
	SideRight Side = iota
	SideLeft
	SideTop
	SideBottom
)
```

Types documented in other references: `ModalOptions` → [modal.md](modal.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Modal and Slideover

##### Open an edit dialog

Modal appears in the center, and Slideover slides in from an edge of the window. Both are modal: the background cannot be used while they are open.

```go
// state stores EditOpen, Draft, and EditTrigger (a ui.Handle) per window.
edit := nui.Button(c, t, "Edit", nui.ButtonOptions{}).
    Bind(&state.EditTrigger).
    OnClick(func() { state.EditOpen = true })
nui.Tooltip(c, t, edit, "Edit the current profile", nui.TooltipOptions{})

// Call this on every build, including when EditOpen is false.
nui.ModalWithOptions(c.Key("edit"), t, &state.EditOpen, nui.ModalOptions{
    Title: "Edit profile",
    Description: "Changes apply only after you save.",
    ReturnFocus: &state.EditTrigger,
    Footer: func(c *ui.Context, inner nui.Theme) {
        nui.Button(c, inner, "Done", nui.ButtonOptions{}).
            OnClick(func() { state.EditOpen = false })
    },
}, func(inner nui.Theme) {
    nui.Input(c, inner, &state.Draft, nui.InputOptions{Label: "Name"})
})
```

This example only shows opening, editing a draft, and closing. To actually save the data, handle it in the app's Done action. For a simple dialog, you can also use `Modal(c, t, &open, title, body)`; use ModalWithOptions for the full set of options.

##### Closing and focus

| ModalOptions | Behavior |
| --- | --- |
| Dismissible | Optional[bool], default true. false prevents closing by clicking the overlay or pressing Escape |
| CloseButton / CloseLabel | The close button is shown by default; you can hide it or rename it separately |
| Close / CloseIcon | Configure the close button's color, size, loading state, and icon. Defaults to MD, Neutral, Ghost, and X. Close.Leading takes precedence over CloseIcon |
| OnClose | Called once when the user closes the dialog with the close button, an allowed overlay click, or Escape. Not called when the app sets open=false directly |
| OnClosePrevent | Called once when Dismissible=false blocks Escape or an outside click. Dismissible does not affect the close button |
| ReturnFocus | The Handle that receives focus after the dialog closes |

Dismissible and CloseButton are independent. To prevent the user from closing the dialog, set both to `nui.Set(false)`. When a menu item opens a dialog, ReturnFocus should usually point to the menu's trigger button, which still exists after the menu closes; the menu item itself unmounts with the menu.

Tab focus is trapped inside the modal content, and an inner Select or Popover handles Escape first. By default, the first focusable control receives focus, usually the close button. To focus a field first, request that field's Handle in the action that opens the dialog. If the Header or Footer is very tall, your app must constrain the content size to leave room for the body.

##### Content, size, and appearance

| Option | Behavior |
| --- | --- |
| Title / Description | The dialog's name and description, also shown in the default header |
| Header / Footer | `func(*ui.Context, Theme)`. Header replaces the title and description content, while the close button is still created separately. Footer customizes the action area |
| Width / MaxHeight | Optional[float32]. The default width is 512 DIP, limited to the window width minus 32 and the window height minus 64 |
| Overlay | Default true. false only hides the overlay color; focus is still trapped |
| Fullscreen | Uses the full window size without rounded corners or a border. Modal only |
| Transition | Optional[bool], default true. false skips the entrance animation; switching back to true does not replay an entrance that already finished |

The body scrolls while the title and Footer stay in place. The header has 16 DIP of vertical padding, 24 DIP of horizontal padding, and a minimum height of 64 DIP. The body has 24 DIP of padding, and the Footer has a 6 DIP gap. If Title, Description, and Header are all empty and CloseButton=false, no header is created. If body is nil, no body region is created. With only a Footer, there is no stray separator.

The Panel's text styles are inherited directly by ordinary body elements. Panel.Gap sets the spacing between the regions that exist and defaults to 0. Inner content is clipped to the panel's rounded corners. The header, body, and Footer have stable identities, so removing the header or toggling Fullscreen keeps focus and selection in a body editor. Closing unmounts the content. Values your app stores persistently remain, but temporary control state is not guaranteed to survive reopening.

The Footer is right-aligned by default. PartStyle cannot set alignment; for a different arrangement, create the layout yourself inside the Footer callback.

The Modal panel has an 8 DIP corner radius, and the overlay uses the Elevated color at 75% opacity. On open, the panel fades in from 95% scale over 200 ms while the overlay fades in with it. With Transition=false or reduced motion, it appears immediately. Like [Popover](popover.md), Modal closes immediately without an exit animation. There is no upstream full-overlay scrolling mode, non-modal switch, or keep-mounted option.

##### Slideover: open from an edge

Slideover reuses Modal's title, body, Footer, and close options. Side and Extent determine its position and size:

| Side | What Extent means, and its default |
| --- | --- |
| SideRight (default) / SideLeft | Panel width, 448 DIP by default |
| SideTop / SideBottom | Panel height. When unset, it fits the content, up to the full window |

An explicit Extent is limited to the window's size along that axis. The Width, MaxHeight, and Fullscreen fields embedded from ModalOptions are not used by Slideover. The body stretches and scrolls, and the Footer sits at the bottom of the panel.

Inset defaults to false. When enabled, it leaves a 16 DIP margin on all sides of the window and uses an 8 DIP corner radius. The margin scales with the theme's spacing, and Extent is limited to the space left after the margin. Side is a physical edge and is not swapped in RTL. Left and right panels default to 448 DIP wide; Go does not use upstream's breakpoint rule that makes the panel full width on small screens.

Slideover always creates a body region; even with content=nil, it keeps the default 24 DIP of padding. For top and bottom panels sized to their content, this affects the panel height. This differs from Modal, which creates no body region when the body is nil. Body, Header, and Footer have stable identities, so changing the side or Inset keeps focus and selection in an editor.

Opening and closing use a fixed 200 ms cubic ease-out slide and overlay transition. With Transition=false, Theme.Motion.Reduce, or the system's reduced motion setting, the change is immediate. During the exit, the modal scope is kept and the content cannot be used; afterward, the panel unmounts and focus is restored. If the direction reverses midway, the transition continues from its current progress.

A Background override on the Backdrop, from the theme or the instance, also fades in and out, and the final Style still has the highest precedence. The native Backdrop contains the panel, so setting its Opacity affects the panel too. To hide only the overlay, use Overlay=false or a transparent background.

Slideover cannot be stretched from its edge. For dragging by a bottom handle with multiple snap heights, use [Drawer](drawer.md).

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Modal / Slideover | Backdrop, Panel, Header, Label, Description, Close, Body, Footer |

### Options and defaults

#### Common options by component › Overlays and notifications (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Modal / Slideover | CloseButton controls whether the close button appears; Close and CloseIcon configure it. Transition=false skips the transition. OnClosePrevent reports close attempts blocked by Dismissible=false |
| Slideover | Inset adds edge spacing and rounded corners. Extent sets only the size along the relevant axis; Width, MaxHeight, and Fullscreen do not affect the layout |

### Advanced interaction, charts, and layout

#### NavigationMenu and Sidebar › Toggles, side, and callbacks (excerpt)

In RTL, the border, button icon, handle, and Slideover all adjust together. Your app still has to place a right-side sidebar after the main content in the Row. Widths are in DIP; web percentages, rem, and DashboardGroup units are not used.

#### Drawer › Control the open state and height (excerpt)

Drawer is a modal bottom sheet with a separate drag handle. It reuses Slideover's exit animation and focus restoration. The title, body, and footer scroll together inside ContainerSlot, so the close actions can be reached even at the lowest snap height. Your app stores a DrawerState, controls opening with Open, and controls the snap position with Snap.

The default snap heights are 50% and 85% of the available window height. SnapPoints must be strictly increasing, and each value must be in `(0, 1]`; px strings are not accepted. With Inset=true, a 16 DIP margin is subtracted before the height is computed. To open from another edge, use Slideover. The shared Width, MaxHeight, and Fullscreen fields have no effect on Drawer.

#### Drawer › Dragging and closing (excerpt)

The visible handle is 48×6 DIP by default, rounded, in the Accented color, and styled with Indicator. Thumb is the 64×16 DIP interactive area. An empty title creates no empty heading, and the panel and handle have localized names. Snap height changes are immediate, and opening and closing use Slideover's animation. Call Drawer on every build, including when it is closed. Dragging from any edge, spring snapping, and touch gestures linked to body scrolling are not supported.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Open              bool
	Draft             string
	Closed, Prevented int
	trigger           ui.Handle
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Button(c.Key("open"), t, "Open panel", nui.ButtonOptions{}).Bind(&s.trigger).OnClick(func() { s.Open = true })
	var extent nui.Optional[float32]

	styles := nui.Styles{}

	var body func(nui.Theme)

	body = func(t nui.Theme) {
		nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Panel draft"})
		ui.Text(c, "The footer stays visible while the body scrolls.")
	}

	nui.Slideover(c.Key("panel"), t, &s.Open, nui.SlideoverOptions{ModalOptions: nui.ModalOptions{Title: "Project details", ReturnFocus: &s.trigger, Decoration: nui.Decoration{UI: styles}, OnClose: func() { s.Closed++ }, OnClosePrevent: func() { s.Prevented++ }, Footer: func(c *ui.Context, t nui.Theme) {
		nui.Button(c.Key("close"), t, "Done", nui.ButtonOptions{}).OnClick(func() { s.Open = false })
	}}, Side: nui.SideRight, Extent: extent}, body)
	ui.Text(c, fmt.Sprintf("Panel notifications: closed=%d prevented=%d", s.Closed, s.Prevented))
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Slideover · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
