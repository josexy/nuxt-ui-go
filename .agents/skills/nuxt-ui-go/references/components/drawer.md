<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Drawer

Open a bottom panel and adjust its height with a handle.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Drawer`
- Types: `DrawerOptions`, `DrawerState`
- Theme key: `nui.ComponentDrawer` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep DrawerState persistent: Open controls visibility and Snap selects a stop. SnapPoints are ascending fractions of available window height; Inset subtracts the margins first. Render while closed for exit animation and return focus. Handle controls the independent handle; CloseButton defaults to false.

## Things to know

This native bottom sheet is dragged only by its handle. SnapPoints accepts fractions in (0,1], not px strings; defaults are 0.5/0.85. Snaps change immediately; entrance/exit use Slideover. DismissDistance/DismissVelocity are native thresholds with 0 selecting defaults, not Vaul closeThreshold equivalents. Overlay=false only hides the backdrop, retaining modal interaction rules. Width, MaxHeight and Fullscreen do not apply.

## API

```go
// Drawer is a bottom sheet. Only its handle drags; scrolling body content cannot dismiss it.
// Call every frame, including while closed, as for Slideover.
func Drawer(c *ui.Context, t Theme, s *DrawerState, o DrawerOptions, content func(Theme)) ui.Element
```

### DrawerOptions

```go
type DrawerOptions struct {
	ModalOptions
	Handle Optional[bool]
	Inset  bool
	// SnapPoints are ascending fractions of available window height, in (0,1].
	SnapPoints                       []float32
	DismissDistance, DismissVelocity float32
	OnSnap                           func(int)
}
```

### DrawerState

```go
type DrawerState struct {
	Open bool
	Snap int
	// contains filtered or unexported fields
}
```

Types documented in other references: `ModalOptions` → [modal.md](modal.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### Drawer

##### Control the open state and height

Drawer is a modal bottom sheet with a separate drag handle. It reuses Slideover's exit animation and focus restoration. The title, body, and footer scroll together inside ContainerSlot, so the close actions can be reached even at the lowest snap height. Your app stores a DrawerState, controls opening with Open, and controls the snap position with Snap.

The default snap heights are 50% and 85% of the available window height. SnapPoints must be strictly increasing, and each value must be in `(0, 1]`; px strings are not accepted. With Inset=true, a 16 DIP margin is subtracted before the height is computed. To open from another edge, use Slideover. The shared Width, MaxHeight, and Fullscreen fields have no effect on Drawer.

##### Dragging and closing

Only the handle drags; the body keeps scrolling. Dragging down lowers the panel, and on release a snap point is chosen from the current height and velocity. The drawer closes when its height falls below half the lowest snap point, or when it moves more than 80 DIP (the default) at more than 900 DIP/s. DismissDistance and DismissVelocity adjust these thresholds; 0 keeps the default, and values must be finite and non-negative. These are native thresholds and are not equivalent to Vaul's closeThreshold.

Dismissible=false prevents closing by dragging, clicking the overlay, or pressing Escape, and fires OnClosePrevent; an explicit close button still works. CloseButton defaults to false, but passing Close options shows the button by default, and an explicit false takes precedence. Handle defaults to true and is independent of Header; hiding it cancels any drag in progress.

The handle is focusable. Up and Down switch snap points, and Home and End go to the first and last. Losing pointer capture or window focus cancels the drag and restores the previous snap position. OnSnap reports only actual index changes; repeated keys at the bounds and outside changes are silent. OnClose reports close actions performed by the component; changing Open directly or a custom Done button does not report.

The visible handle is 48×6 DIP by default, rounded, in the Accented color, and styled with Indicator. Thumb is the 64×16 DIP interactive area. An empty title creates no empty heading, and the panel and handle have localized names. Snap height changes are immediate, and opening and closing use Slideover's animation. Call Drawer on every build, including when it is closed. Dragging from any edge, spring snapping, and touch gestures linked to body scrolling are not supported.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| Drawer | DrawerState | Backdrop, Panel, ContainerSlot, Thumb, Indicator, Header, Label, Description, Body, Footer, Close |

### Overlays, menus, and navigation

#### Modal and Slideover › Slideover: open from an edge (excerpt)

Slideover cannot be stretched from its edge. For dragging by a bottom handle with multiple snap heights, use Drawer.

### Compatibility and known limitations

#### Feature scope › Layout and navigation (excerpt)

| Area | Current behavior and limits |
| --- | --- |
| Drawer | Dragged by a separate handle at the bottom; the content scrolls as a whole. No dragging from other edges, Vaul background scaling, or gestures linked to body scrolling. SnapPoints accept only height fractions; native close thresholds and animation are documented separately |

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
	Drawer                   nui.DrawerState
	Opener                   ui.Handle
	Draft                    string
	Snaps, Closed, Prevented int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var parts nui.Styles

	nui.Button(c.Key("open"), t, "Open drawer", nui.ButtonOptions{}).Bind(&s.Opener).OnClick(func() { s.Drawer.Open = true })
	nui.Drawer(c.Key("drawer"), t, &s.Drawer, nui.DrawerOptions{ModalOptions: nui.ModalOptions{Title: "Quick actions", CloseButton: nui.Set(false), ReturnFocus: &s.Opener, OnClose: func() { s.Closed++ }, OnClosePrevent: func() { s.Prevented++ }, Decoration: nui.Decoration{UI: parts}, Footer: func(c *ui.Context, t nui.Theme) {
		nui.Button(c.Key("close"), t, "Done", nui.ButtonOptions{}).OnClick(func() { s.Drawer.Open = false })
		ui.Textf(c, "Snap: %d · Changes: %d · Closed: %d · Prevented: %d", s.Drawer.Snap, s.Snaps, s.Closed, s.Prevented)
	}}, Handle: nui.Set(true), SnapPoints: []float32{0.5, 0.8}, OnSnap: func(int) { s.Snaps++ }}, func(t nui.Theme) {
		nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Drawer draft"})
		ui.Text(c, "Drag the bottom drawer handle.")
		for i := range 8 {
			ui.Textf(c, "Review item %02d", i+1)
		}
	})
	ui.Textf(c, "Snap: %d · Changes: %d · Closed: %d · Prevented: %d", s.Drawer.Snap, s.Snaps, s.Closed, s.Prevented)
}

func main() {
	s := &example{}

	s.Draft = "Review build"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Drawer · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
