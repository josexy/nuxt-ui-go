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
type ToastProgressOptions struct {
	Color Color
	UI    Styles
}
type ToastActionOptions struct {
	ID       string
	Label    string
	OnAction func()
	Close    Optional[bool]
	ButtonOptions
}
type toastAppearance struct {
	Theme           Theme
	Color           Color
	Icon            *ui.SVG
	Avatar          *AvatarOptions
	UI              Styles
	Orientation     Orientation
	Close, Progress bool
	CloseOptions    *ButtonOptions
	CloseIcon       *ui.SVG
	ProgressOptions *ToastProgressOptions
	Actions         []ToastActionOptions
}

// Toast adds or replaces a notification in this window's native queue. Call from an action.
func Toast(c *ui.Context, t Theme, o ToastOptions) string {
	// A window notification keeps the visual theme, not the field/form scope
	// from which the application published it.
	t.field, t.group, t.form = fieldDefaults{}, groupPart{}, nil
	if o.Timeout == 0 {
		o.Timeout = 5 * time.Second
	}
	color := selected(defaults(t, ComponentToast, ComponentTheme{Color: Primary}), ComponentTheme{Color: o.Color}).Color
	orientation := valueOr(o.Orientation, Vertical)
	vertical(orientation)
	actions := slices.Clone(o.Actions)
	ids := map[string]bool{}
	for i := range actions {
		if actions[i].ID != "" {
			if ids[actions[i].ID] {
				panic("nui: Toast action IDs must be unique")
			}
			ids[actions[i].ID] = true
		}
		actions[i].ButtonOptions = *cloneToastButton(&actions[i].ButtonOptions)
	}
	var progress *ToastProgressOptions
	if o.ProgressOptions != nil {
		copied := *o.ProgressOptions
		copied.UI = cloneStyles(copied.UI)
		progress = &copied
	}
	return c.AddToast(ui.Toast{ID: o.ID, Title: o.Title, Description: o.Description, Action: o.Action, Timeout: o.Timeout, OnAction: o.OnAction, OnClose: o.OnClose, Data: toastAppearance{Theme: t, Color: color, Icon: o.Icon, Avatar: cloneToastAvatar(o.Avatar), UI: cloneStyles(o.UI), Orientation: orientation, Close: valueOr(o.Close, true), CloseOptions: cloneToastButton(o.CloseOptions), CloseIcon: o.CloseIcon, Progress: valueOr(o.Progress, true), ProgressOptions: progress, Actions: actions}})
}

func cloneToastAvatar(src *AvatarOptions) *AvatarOptions {
	if src == nil {
		return nil
	}
	value := *src
	value.UI = cloneStyles(value.UI)
	return &value
}
func cloneToastButton(src *ButtonOptions) *ButtonOptions {
	if src == nil {
		return nil
	}
	value := *src
	value.UI, value.Avatar = cloneStyles(value.UI), cloneToastAvatar(value.Avatar)
	return &value
}

type toastActionKey struct {
	id    string
	index int
}
type toastFocus struct {
	orientation Orientation
	actions     map[toastActionKey]*ui.Handle
	legacy      ui.Handle
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
		top := o.Position >= ToastTopRight
		start, end := ui.Start, ui.End
		if c.Direction() == ui.RightToLeft {
			start, end = end, start
		}
		viewport.Padding(t.spacing(16)).Gap(t.spacing(8)).AlignItems(end)
		if top {
			viewport.Justify(ui.Start)
		}
		switch o.Position {
		case ToastBottomLeft, ToastTopLeft:
			viewport.AlignItems(start)
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
		for i := range toasts {
			index := i
			if top {
				index = len(toasts) - 1 - i
			}
			ts := toasts[index]
			appearance, ok := ts.Data.(toastAppearance)
			if !ok {
				appearance = toastAppearance{Theme: t, Color: Primary, Orientation: Vertical, Close: true, Progress: true}
			}
			inner := appearance.Theme
			d := Decoration{UI: appearance.UI}
			color := inner.color(appearance.Color)
			parts := ui.ToastBase(c, ts)
			focus := ui.Local(parts.Root, "nui-toast-focus", func() toastFocus {
				return toastFocus{orientation: appearance.Orientation, actions: map[toastActionKey]*ui.Handle{}}
			})
			relocated := focus.orientation != appearance.Orientation
			focus.orientation = appearance.Orientation
			bindAction := func(button ui.Element, handle *ui.Handle) {
				restore := relocated && handle.Focused(c) && !button.IsDisabled()
				button.Bind(handle)
				if restore {
					handle.Focus(c)
				}
			}
			// Nuxt's sm:w-96 toast with p-4; the progress bar is clipped to its corners.
			width := float32(384)
			if o.Width.Set {
				width = o.Width.Value
			}
			root := parts.Root.Column().MinWidth(0).Width(max(0, width)).MaxWidthPercent(100).Label(ts.Title).Clip()
			recipe := dialogStyle(inner)
			recipe.PaddingX, recipe.PaddingY = Set(inner.spacing(16)), Set(inner.spacing(16))
			finish(inner, ComponentToast, Panel, root, recipe, d)
			if inner.motion(c) {
				shift := float32(8)
				if top {
					shift = -shift
				}
				root.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Enter: &ui.Motion{Y: shift}, Exit: &ui.Motion{Y: shift}})
			}
			root.Children(func() {
				actions := func() {
					group := ui.Row(c.Key("actions")).Wrap()
					if appearance.Orientation == Vertical {
						group.Margin(inner.spacing(10), 0, 0, 0)
					}
					finish(inner, ComponentToast, Actions, group, PartStyle{Gap: Set(inner.spacing(6))}, d)
					group.Children(func() {
						if ts.Action != "" {
							action := parts.ActionButton().Label(ts.Action).Padding(4, 8).Radius(inner.radius(4)).TextColor(color).FocusRing(true)
							bindAction(action, &focus.legacy)
							action.Children(func() { ui.Text(c, ts.Action).FontWeight(500) })
						}
						for i, action := range appearance.Actions {
							// Nuxt's toast actions: xs solid buttons in the toast's color.
							opts := action.ButtonOptions
							if opts.Variant == VariantDefault {
								opts.Variant = Solid
							}
							if opts.Color == ColorDefault {
								opts.Color = appearance.Color
							}
							if opts.Size == SizeDefault {
								opts.Size = XS
							}
							key := toastActionKey{id: action.ID}
							if action.ID == "" {
								key.index = i
							}
							button := Button(c.Key(key), inner, action.Label, opts).OnClick(func() {
								if action.OnAction != nil {
									action.OnAction()
								}
								if valueOr(action.Close, true) {
									CloseToast(c, ts.ID)
								}
							})
							if focus.actions[key] == nil {
								focus.actions[key] = &ui.Handle{}
							}
							bindAction(button, focus.actions[key])
						}
					})
				}
				hasActions := ts.Action != "" || len(appearance.Actions) > 0
				row := ui.Row(c).MinWidth(0).AlignItems(ui.Start).Gap(inner.spacing(10))
				if appearance.Orientation == Horizontal {
					row.AlignItems(ui.Center)
				}
				row.Children(func() {
					icon := appearance.Icon
					if appearance.Avatar != nil {
						diameter := Optional[float32]{}
						if appearance.Avatar.Size == SizeDefault {
							diameter = Set(inner.text(44))
						}
						finish(inner, ComponentToast, Leading, avatar(c.Key("avatar"), inner, *appearance.Avatar, diameter), PartStyle{}, d)
					} else if icon != nil {
						finish(inner, ComponentToast, Leading, ui.Icon(c.Key("icon"), icon).Size(inner.text(20), inner.text(20)).Shrink(0), PartStyle{Foreground: Set(color)}, d)
					}
					body := ui.Column(c.Key("body")).Grow(1).MinWidth(0)
					finish(inner, ComponentToast, Body, body, PartStyle{}, d)
					body.Children(func() {
						if ts.Title != "" {
							finish(inner, ComponentToast, Label, ui.Text(c, ts.Title), PartStyle{Foreground: Set(inner.Highlighted), Weight: Set(500)}, d)
						}
						if ts.Description != "" {
							text := ui.Text(c, ts.Description)
							if ts.Title != "" {
								text.Margin(inner.spacing(4), 0, 0, 0)
							}
							finish(inner, ComponentToast, Description, text, PartStyle{Foreground: Set(inner.Muted)}, d)
						}
						if hasActions && appearance.Orientation == Vertical {
							actions()
						}
					})
					if hasActions && appearance.Orientation == Horizontal {
						actions()
					}
					if appearance.Close {
						close := Button(c.Key("close"), inner, "", toastCloseOptions(c, appearance, ts.Title)).OnClick(func() { CloseToast(c, ts.ID) })
						finish(inner, ComponentToast, Close, close, PartStyle{}, d)
					}
				})
				if appearance.Progress && ts.Timeout >= 0 {
					total := ts.Timeout
					if total == 0 {
						total = 4 * time.Second
					}
					left := parts.Left()
					// Pinned to the bottom edge across the full width, as Nuxt's.
					opts := ProgressOptions{Color: appearance.Color, Size: SM}
					if appearance.ProgressOptions != nil {
						if appearance.ProgressOptions.Color != ColorDefault {
							opts.Color = appearance.ProgressOptions.Color
						}
						opts.UI = appearance.ProgressOptions.UI
					}
					// The toasts' time stops as MyGo's ToastViewportBase stops it. The
					// bar empties in frames that only paint; reduced motion steps
					// it four times a second instead of moving it smoothly.
					stopped := viewport.Hovered() || viewport.FocusWithin() && c.WindowActive()
					opts.countdown = &progressCountdown{left: left, total: total, at: c.Now(), running: left > 0 && !stopped}
					if !inner.motion(c) {
						opts.countdown.fps = 4
					}
					Progress(c.Key("progress"), inner, 100*float64(left)/float64(total), opts).Absolute().Left(0).Right(0).Bottom(0)
					if opts.countdown.running {
						// A build closes the toast on time, and keeps the bar's
						// accessible value current once a second.
						c.After(min(left, time.Second))
					}
				}
			})
		}
	})
}

func toastCloseOptions(c *ui.Context, appearance toastAppearance, title string) ButtonOptions {
	opts := ButtonOptions{}
	if appearance.CloseOptions != nil {
		opts = *appearance.CloseOptions
	}
	if opts.Size == SizeDefault {
		opts.Size = MD
	}
	if opts.Color == ColorDefault {
		opts.Color = Neutral
	}
	if opts.Variant == VariantDefault {
		opts.Variant = LinkVariant
	}
	if opts.AccessibleLabel == "" {
		opts.AccessibleLabel = namedMessage(c, "dismissNamed", title)
	}
	if opts.Leading == nil {
		opts.Leading = appearance.CloseIcon
		if opts.Leading == nil {
			opts.Leading = lucide.X()
		}
	}
	opts.UI = cloneStyles(opts.UI)
	opts.UI[Root] = mergeStyle(PartStyle{PaddingX: Set(float32(0)), PaddingY: Set(float32(0))}, opts.UI[Root])
	return opts
}
