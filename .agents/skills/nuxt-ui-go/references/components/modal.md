<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Modal

Complete an edit or confirmation in a modal dialog.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Modal`, `ModalWithOptions`
- Types: `ModalOptions`
- Theme key: `nui.ComponentModal` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep a persistent open bool and call ModalWithOptions on every build, even while closed. Title and Description name the dialog, the content callback builds a scrolling body, and Footer builds actions.

## Things to know

Overlay=false hides the backdrop while keeping modal focus containment. CloseButton=false hides the close button. Dismissible=false blocks Escape and outside clicks and calls OnClosePrevent. OnClose counts component dismissal only; Save changes application state directly. Closing unmounts immediately, with no exit animation or keep-mounted option; persistent application drafts remain.

## API

```go
// Modal is a convenience wrapper; use ModalWithOptions for additional options.
func Modal(c *ui.Context, t Theme, open *bool, title string, content func(Theme)) ui.Element

func ModalWithOptions(c *ui.Context, t Theme, open *bool, o ModalOptions, content func(Theme)) ui.Element
```

### ModalOptions

```go
type ModalOptions struct {
	Title, Description, CloseLabel string
	Width, MaxHeight               Optional[float32]
	Dismissible, CloseButton       Optional[bool]
	Transition                     Optional[bool]
	Close                          *ButtonOptions
	CloseIcon                      *ui.SVG
	Overlay                        Optional[bool] // Defaults to true; false hides the backdrop, not the modal focus scope.
	Fullscreen                     bool
	Header, Footer                 func(*ui.Context, Theme)
	OnClose                        func()
	OnClosePrevent                 func()
	ReturnFocus                    *ui.Handle
	Decoration
}
```

Types documented in other references: `ButtonOptions` → [button.md](button.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

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

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Modal / Slideover | Backdrop, Panel, Header, Label, Description, Close, Body, Footer |

### Options and defaults

#### Common options by component › Overlays and notifications (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Modal / Slideover | CloseButton controls whether the close button appears; Close and CloseIcon configure it. Transition=false skips the transition. OnClosePrevent reports close attempts blocked by Dismissible=false |

### Inputs, selection, and fields

#### State, themes, and events › Focus and callbacks (excerpt)

Since MyGo v0.9.0, focus moved by the keyboard or by a component scrolls the scroll containers around the newly focused control as little as needed to show it, as the web's `element.focus()` does. This covers roving focus in Tabs and menus, Form focusing its first invalid field, and focus returning to a trigger when Modal or DropdownMenu closes. A target that is already fully visible does not scroll. A control that a component keeps focused across builds does not undo the user's scrolling, and a pointer press focuses a control where it is without scrolling. To move focus from app code without scrolling, pass `ui.FocusOptions` to the element's Focus:

```go
cell := ui.Box(c.Key(id)).Focusable()
if app.restoreFocus == id {
	cell.Focus(ui.FocusOptions{PreventScroll: true})
}
```

### Compatibility and known limitations

#### Accessibility and motion (excerpt)

Programmatic focus scrolls the newly focused control into view as little as needed, as described in Focus and callbacks. `ui.Element.Focus` accepts `ui.FocusOptions{PreventScroll: true}`, but MyGo v0.9.0's `ui.Handle.Focus` takes no options, so focus moved through a persistent Handle, such as a Modal's or Sidebar's ReturnFocus, always reveals a target that is not fully visible.

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
	Open              bool
	Value             string
	Closed, Prevented int
	trigger           ui.Handle
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Button(c.Key("open"), t, "Open dialog", nui.ButtonOptions{}).Bind(&s.trigger).OnClick(func() { s.Open = true })
	title, description := "Edit project", "Edit and save the project name."

	nui.ModalWithOptions(c.Key("modal"), t, &s.Open, nui.ModalOptions{Title: title, Description: description, ReturnFocus: &s.trigger, OnClose: func() { s.Closed++ }, OnClosePrevent: func() { s.Prevented++ }, Footer: func(c *ui.Context, t nui.Theme) {
		nui.Button(c.Key("save"), t, "Save", nui.ButtonOptions{}).OnClick(func() { s.Open = false })
	}}, func(t nui.Theme) {
		nui.Input(c.Key("name"), t, &s.Value, nui.InputOptions{Label: "Project name"})
		ui.Text(c.Key("explanation"), "The body scrolls while the header and footer stay visible.")
	})
	ui.Text(c, fmt.Sprintf("Close notifications: %d · Prevented: %d", s.Closed, s.Prevented))
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Modal · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
