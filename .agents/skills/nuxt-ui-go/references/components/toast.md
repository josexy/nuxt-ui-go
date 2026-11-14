<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Toast

Briefly show the result of an action.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Toast`, `ToastViewport`, `CloseToast`
- Types: `ToastOptions`, `ToastViewportOptions`, `ToastActionOptions`, `ToastProgressOptions`
- Theme key: `nui.ComponentToast` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Build ToastViewport once per window and call Toast from actions or UI handling of completed work. CloseToast accepts the returned ID. Publishing the same ID replaces its content and restarts its timeout. Stable action IDs preserve focus across replacement and reordering.

## Things to know

This demo uses four seconds. Go Timeout=0 means the five-second default and negative values persist; Nuxt UI duration=0 persists. A notice snapshots its theme and configuration, while its generated close name follows locale changes. MyGo supplies at most three visible notices; older queued notices keep timing and may reappear after newer ones close. Upstream max defaults to five and drops older entries.

## API

```go
// Toast adds or replaces a notification in this window's native queue. Call from an action.
func Toast(c *ui.Context, t Theme, o ToastOptions) string

func ToastViewport(c *ui.Context, t Theme, o ToastViewportOptions) ui.Element

// CloseToast closes the ID, or all notifications for an empty ID.
func CloseToast(c *ui.Context, id string)
```

### ToastOptions

```go
type ToastOptions struct {
	ID, Title, Description, Action string
	Color                          Color
	Icon                           *ui.SVG
	Avatar                         *AvatarOptions
	Timeout                        time.Duration
	OnAction, OnClose              func()
	UI                             Styles
	Orientation                    Optional[Orientation]
	Close, Progress                Optional[bool]
	CloseOptions                   *ButtonOptions
	CloseIcon                      *ui.SVG
	ProgressOptions                *ToastProgressOptions
	Actions                        []ToastActionOptions
}
```

### ToastViewportOptions

```go
type ToastViewportOptions struct {
	Position ToastPosition
	Width    Optional[float32]
	Limit    int
	Decoration
}
```

### ToastActionOptions

```go
type ToastActionOptions struct {
	ID       string
	Label    string
	OnAction func()
	Close    Optional[bool]
	ButtonOptions
}
```

### ToastProgressOptions

```go
type ToastProgressOptions struct {
	Color Color
	UI    Styles
}
```

### Related types

```go
type ToastPosition uint8

const (
	ToastBottomRight ToastPosition = iota
	ToastBottomLeft
	ToastBottomCenter
	ToastTopRight
	ToastTopLeft
	ToastTopCenter
)
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `ButtonOptions` → [button.md](button.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `ProgressOptions` → [progress.md](progress.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Shared conventions › Styles and callbacks (excerpt)

Most components return a `ui.Element`. Toast returns the notification ID, and CloseToast returns nothing. Style configures parts synchronously during the current build; use click or selection callbacks for business actions.

#### Toast

##### Add a notification area and post notifications

Build one ToastViewport per window, and call Toast from user actions:

```go
t := nui.App(c, config)
// Put this before other overlays so modal overlays can cover notifications.
nui.ToastViewport(c, t, nui.ToastViewportOptions{})
nui.Button(c, t, "Save", nui.ButtonOptions{}).OnClick(func() {
    previous := state.Value
    state.Value = state.Draft
    nui.Toast(c, t, nui.ToastOptions{
        ID: "saved",
        Title: "Saved",
        Color: nui.Success,
        Action: "Undo",
        OnAction: func() { state.Value = previous },
    })
})
```

Do not call Toast unconditionally during the view build, or a notification may be posted on every rebuild. Write background results back to the model through Window.Update, then post the notification from a window UI action or an explicitly controlled event handler.

##### IDs, duration, and actions

| Option or action | Behavior |
| --- | --- |
| ID | Generated automatically when empty. Posting with an existing ID replaces that toast and restarts its timer, without calling the old toast's OnClose |
| Timeout | 5 seconds by default; a negative value makes the toast persistent |
| Action / OnAction | A single action. Running it closes the toast and calls OnClose |
| Actions | Several ToastActionOptions. Each closes the toast by default; Close=false on an action keeps it open |
| ToastOptions.Close | `nui.Set(false)` hides the close button without changing the timeout or the rules for closing on action |
| Avatar / Icon | The avatar takes precedence. The default avatar diameter is 44 DIP and the icon is 20 DIP. An explicit avatar size uses Avatar's own size |
| CloseOptions / CloseIcon | Optional button options and icon. Defaults to MD / Neutral / Link, no padding, and X. CloseOptions.Leading takes precedence |
| ProgressOptions | ToastProgressOptions sets Color and the Progress UI. Defaults to size SM, with the color inherited from the toast |
| OnClose / CloseToast | The close button, the timeout, and CloseToast all call OnClose. `CloseToast(c, "")` closes all toasts |

You can use the single action and Actions together. Each button in Actions can set ID, Label, OnAction, ButtonOptions, and Close; buttons default to XS and Solid, and inherit the toast's color. Provide stable IDs when actions are updated or reordered; non-empty IDs must be unique within a toast. Without IDs, actions are identified only by their position in a static list. Changing the orientation keeps focus on the current action button. Without an Icon or Avatar, no status icon is added automatically.

Vertical is the default orientation: the actions sit below the body with a 10 DIP gap. Horizontal centers the content and actions vertically side by side. An empty title leaves no blank line. ToastOptions.UI.Actions styles the action container, such as its spacing; style individual buttons with ButtonOptions.UI on each Actions item. Code that used UI.Actions to style the single compatibility Action button should switch to an Actions list with button UI.

The queue stores OnAction, OnClose, and the ButtonOptions.Style of each action, so their closures should capture only the business model, data, or services. Do not capture Elements, Parts, or Contexts produced by a build; the elements Style receives are valid only during that build.

##### Position, pausing, and style

ToastViewport defaults to the bottom right and a width of 384 DIP. It also supports bottom left, bottom center, top right, top left, and top center. These are physical positions and are not swapped in RTL. Top positions put the newest toast at the top, and bottom positions put it at the bottom. The width is further limited by the viewport's actual padding; an explicit PaddingX=0 removes that extra limit.

Limit accepts 1–3; 0 or an out-of-range value is treated as 3. MyGo shows at most the 3 newest toasts. Older queued toasts keep counting down and may reappear after newer ones close; Limit is not the queue capacity. Nuxt UI's max defaults to 5 and discards older toasts beyond the limit, so the rules differ.

The queue timer pauses while the pointer is over the notification area, or while focus is inside a toast and the window has focus. It keeps running while the window is in the background or hidden, even with focus left inside a toast. A window out of sight draws no frames (since MyGo v0.9.0 this includes a minimized window on Windows and on Linux under X11), so a toast whose time runs out meanwhile closes on the frame the window shows again, running its OnClose then, without showing. The default countdown bar pauses with it; hide it with `Progress: nui.Set(false)`. Persistent toasts have no countdown. With reduced motion, the remaining time is still updated, but it is shown in steps at least 250 ms apart without interpolation. Toasts enter and exit with a 150 ms, 8 DIP fade and slide toward the matching screen edge; with reduced motion, they switch immediately. An empty queue draws no panel.

Toasts use RoleStatus, and the action and close buttons are keyboard accessible. Actual screen reader announcements have not been verified.

Each toast stores a copy of the visual Theme and options at the time it was created, without the source form's disabled state or field-group information. If the global theme changes while a toast is visible, existing toasts keep their original theme; to update a toast, post it again with the same ID. Automatically generated close names still follow the current language, while explicit app text is updated by your app. Viewport.Decoration affects only Root; style each toast through ToastOptions.UI or the ComponentToast configuration at creation time.

Updating with the same ID replaces the toast entirely: fields you do not pass again return to their defaults, the old OnClose is not called, and the timer restarts. There is no upstream duplicate-pulse effect, collapsed stacking, swipe to dismiss, arbitrary content slot, or foreground/background announcement policy.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Toast | Root (viewport only), Panel, Body, Leading, Label, Description, Actions, Close. ProgressOptions.UI uses Progress's Root and Indicator |

Tabs, Stepper, and Timeline accept global and instance Color and Size; Banner and Toast accept Color. Color on menu items and Timeline items is a per-item setting. Other styles are configured through the parts listed above.

### Options and defaults

#### Unset, zero, and default values (excerpt)

For example, to hide the Toast close button, write `Close: nui.Set(false)`. To show no sibling pages in Pagination, write `SiblingCount: nui.Set(0)`. Writing only `Siblings: 0` uses the default of 2.

#### Common options by component › Overlays and notifications (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Toast | Close and Progress control visibility. CloseOptions, CloseIcon, Avatar, and ProgressOptions configure appearance. Stable IDs on Actions keep focus when actions are reordered. Close=false on an action keeps the toast open |

By default, running a Toast action closes the toast; set `Close: nui.Set(false)` on that action to keep it. While hovering or focus pauses the toast, its countdown bar also stops. Persistent toasts have no countdown bar.

#### Limitations (excerpt)

FormField associates its label with this library's controls. If you use a custom MyGo control, your app must associate the label itself. Toast action callbacks and ButtonOptions.Style are stored with the toast, so their closures must not capture temporary Elements, Parts, or Contexts; the Style arguments are valid only during the build in which they are called.

### Themes and basic components

#### App and theme › Create the config once, use it in the view (excerpt)

`App` sets the current window's native theme from the config and the system preferences. Call it before any `Window` or `Themed` call in the same build so it can read the system appearance. You still create MyGo windows with `mygo.NewWindow`, and notifications need a separate `ToastViewport`.

### Internationalization

#### What gets translated › Text your app translates (excerpt)

Text you pass explicitly, such as `Label`, `Description`, `Placeholder`, `ErrorMessage`, `EmptyText`, `ChooseLabel`, and `CloseLabel`, is displayed as is. Your app also translates business validation errors, server errors, Toast titles, menu items, table data, and Timeline dates.

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
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Used, Closed, Published int
	Last                    string
	Reversed                bool
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.ToastViewport(c.Key("viewport"), t, nui.ToastViewportOptions{Position: nui.ToastBottomRight, Width: nui.Set(float32(384)), Limit: int(3)})
	publish := func() {
		var progressUI nui.Styles

		progress := &nui.ToastProgressOptions{UI: progressUI}
		var actions []nui.ToastActionOptions

		actions = []nui.ToastActionOptions{{ID: "keep", Label: "Keep", ButtonOptions: nui.ButtonOptions{AccessibleLabel: "Keep notification"}, Close: nui.Set(false), OnAction: func() { s.Used++ }}, {ID: "done", Label: "Done", ButtonOptions: nui.ButtonOptions{AccessibleLabel: "Done notification"}, OnAction: func() { s.Used++ }}}

		horizontal := false
		if s.Reversed {
			horizontal = !horizontal
			if len(actions) == 2 {
				actions[0], actions[1] = actions[1], actions[0]
			}
		}
		orientation := nui.Vertical
		if horizontal {
			orientation = nui.Horizontal
		}
		id := ""

		id = "example-notification"

		timeout := 4 * time.Second

		s.Last = nui.Toast(c, t, nui.ToastOptions{ID: id, Title: "Saved successfully", Description: "Changes have been saved.", Icon: lucide.Info(), CloseIcon: lucide.X(), Progress: nui.Set(true), ProgressOptions: progress, Orientation: nui.Set(orientation), Actions: actions, Timeout: timeout, OnClose: func() { s.Closed++ }})
		s.Published++
	}
	nui.Button(c.Key("notify"), t, "Show notification", nui.ButtonOptions{}).OnClick(publish)
	nui.Button(c.Key("update"), t, "Update focused notification", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { s.Reversed = !s.Reversed; publish() })
	nui.Button(c.Key("close"), t, "Close latest notification", nui.ButtonOptions{Variant: nui.Outline, Disabled: s.Last == ""}).OnClick(func() { nui.CloseToast(c, s.Last) })
	nui.Button(c.Key("clear"), t, "Clear notifications", nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { nui.CloseToast(c, "") })
	ui.Text(c, fmt.Sprintf("Published: %d · Actions: %d · Closed: %d", s.Published, s.Used, s.Closed))
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Toast · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
