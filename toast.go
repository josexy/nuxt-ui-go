package nui

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type ToastOptions struct {
	ID, Title, Description, Action string
	Color                          Color
	Icon                           *ui.SVG
	Timeout                        time.Duration
	OnAction, OnClose              func()
	UI                             Styles
	Orientation                    Optional[Orientation]
	Close, Progress                Optional[bool]
	Actions                        []ToastActionOptions
}
type ToastActionOptions struct {
	Label    string
	OnAction func()
	Close    Optional[bool]
	ButtonOptions
}
type toastAppearance struct {
	Theme           Theme
	Color           Color
	Icon            *ui.SVG
	UI              Styles
	Orientation     Orientation
	Close, Progress bool
	Actions         []ToastActionOptions
}

// Toast adds or replaces a notification in this window's native queue. Call from an action.
func Toast(c *ui.Context, t Theme, o ToastOptions) string {
	if o.Timeout == 0 {
		o.Timeout = 5 * time.Second
	}
	color := selected(defaults(t, ComponentToast, ComponentTheme{Color: Primary}), ComponentTheme{Color: o.Color}).Color
	orientation := valueOr(o.Orientation, Vertical)
	vertical(orientation)
	actions := slices.Clone(o.Actions)
	for i := range actions {
		actions[i].UI = cloneStyles(actions[i].UI)
	}
	return c.AddToast(ui.Toast{ID: o.ID, Title: o.Title, Description: o.Description, Action: o.Action, Timeout: o.Timeout, OnAction: o.OnAction, OnClose: o.OnClose, Data: toastAppearance{Theme: t, Color: color, Icon: o.Icon, UI: cloneStyles(o.UI), Orientation: orientation, Close: valueOr(o.Close, true), Progress: valueOr(o.Progress, true), Actions: actions}})
}

// CloseToast closes the ID, or all notifications for an empty ID.
func CloseToast(c *ui.Context, id string) { c.CloseToast(id) }

type ToastPosition uint8

const (
	ToastBottomRight ToastPosition = iota
	ToastBottomLeft
	ToastBottomCenter
	ToastTopRight
	ToastTopLeft
	ToastTopCenter
)

type ToastViewportOptions struct {
	Position ToastPosition
	Width    Optional[float32]
	Limit    int
	Decoration
}

func ToastViewport(c *ui.Context, t Theme, o ToastViewportOptions) ui.Element {
	ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	return ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Padding(t.spacing(16)).Gap(t.spacing(8)).AlignItems(ui.End)
		if o.Position >= ToastTopRight {
			viewport.Justify(ui.Start)
		}
		switch o.Position {
		case ToastBottomLeft, ToastTopLeft:
			viewport.AlignItems(ui.Start)
		case ToastBottomCenter, ToastTopCenter:
			viewport.AlignItems(ui.Center)
		}
		finish(t, ComponentToast, Root, viewport, PartStyle{}, o.Decoration)
		limit := o.Limit
		if limit <= 0 || limit > 3 {
			limit = 3
		}
		if len(toasts) > limit {
			toasts = toasts[len(toasts)-limit:]
		}
		for _, ts := range toasts {
			appearance, ok := ts.Data.(toastAppearance)
			if !ok {
				appearance = toastAppearance{Theme: t, Color: Primary, Orientation: Vertical, Close: true, Progress: true}
			}
			inner := appearance.Theme
			d := Decoration{UI: appearance.UI}
			color := inner.color(appearance.Color)
			parts := ui.ToastBase(c, ts)
			w, _ := c.Size()
			width := float32(356)
			if o.Width.Set {
				width = o.Width.Value
			}
			root := parts.Root.Column().MinWidth(0).Width(min(width, max(0, w-32))).Label(ts.Title)
			finish(inner, ComponentToast, Panel, root, dialogStyle(inner), d)
			if !inner.Motion.Reduce {
				root.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Enter: &ui.Motion{Y: 8}, Exit: &ui.Motion{Y: 8}})
			}
			root.Children(func() {
				actions := func() {
					ui.Row(c.Key("actions")).Wrap().Gap(inner.spacing(6)).Children(func() {
						if ts.Action != "" {
							action := parts.ActionButton().Label(ts.Action).Padding(4, 8).Radius(inner.radius(4)).TextColor(color).FocusRing(true)
							finish(inner, ComponentToast, Actions, action, PartStyle{}, d)
							action.Children(func() { ui.Text(c, ts.Action).FontWeight(500) })
						}
						for i, action := range appearance.Actions {
							opts := action.ButtonOptions
							if opts.Variant == VariantDefault {
								opts.Variant = Outline
							}
							if opts.Color == ColorDefault {
								opts.Color = Neutral
							}
							if opts.Size == SizeDefault {
								opts.Size = SM
							}
							Button(c.Key(i), inner, action.Label, opts).OnClick(func() {
								if action.OnAction != nil {
									action.OnAction()
								}
								if valueOr(action.Close, true) {
									CloseToast(c, ts.ID)
								}
							})
						}
					})
				}
				hasActions := ts.Action != "" || len(appearance.Actions) > 0
				ui.Row(c).MinWidth(0).AlignItems(ui.Start).Gap(10).Children(func() {
					icon := appearance.Icon
					if icon != nil {
						finish(inner, ComponentToast, Leading, ui.Icon(c, icon).Size(20, 20).Shrink(0), PartStyle{Foreground: Set(color)}, d)
					}
					ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
						finish(inner, ComponentToast, Label, ui.Text(c, ts.Title), PartStyle{Foreground: Set(inner.Highlighted), Weight: Set(500)}, d)
						if ts.Description != "" {
							finish(inner, ComponentToast, Description, ui.Text(c, ts.Description), PartStyle{Foreground: Set(inner.Muted)}, d)
						}
						if hasActions && appearance.Orientation == Vertical {
							actions()
						}
					})
					if hasActions && appearance.Orientation == Horizontal {
						actions()
					}
					if appearance.Close {
						close := parts.CloseButton().Label(namedMessage(c, "dismissNamed", ts.Title)).Size(20, 20).FocusRing(true)
						finish(inner, ComponentToast, Close, close, PartStyle{}, d)
						close.Children(func() { ui.Icon(c, lucide.X()).Size(16, 16).TextColor(inner.Muted) })
					}
				})
				if appearance.Progress && ts.Timeout >= 0 {
					total := ts.Timeout
					if total == 0 {
						total = 4 * time.Second
					}
					left := parts.Left()
					Progress(c.Key("progress"), inner, 100*float64(left)/float64(total), ProgressOptions{Color: appearance.Color, Size: XS})
					if left > 0 && inner.motion(c) && !viewport.Hovered() && !viewport.FocusWithin() {
						c.After(50 * time.Millisecond)
					}
				}
			})
		}
	})
}
