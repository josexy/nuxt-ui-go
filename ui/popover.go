package nui

import (
	"time"

	"github.com/egoist/mygo/ui"
)

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
type PopoverMode uint8

const (
	PopoverClick PopoverMode = iota
	PopoverHover
)

type popoverMemory struct {
	since                time.Time
	intent, suppressed   bool
	wasOpen, focusWithin bool
}

// popupStyle is Nuxt's popup surface for Popover, Select, InputMenu, SelectMenu
// and menus: rounded-md, a default ring and shadow-lg. Popover keeps a 16 DIP
// padding, as Nuxt's documented examples add; lists and menus tighten it.
func popupStyle(t Theme) PartStyle {
	return PartStyle{Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14)), Foreground: Set(t.Text), Background: Set(t.Background), Radius: Set(t.radius(6)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border), PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(8)), Shadow: Set(t.Shadow)}
}

// dialogStyle is Nuxt's dialog surface for Modal, Slideover and Toast: rounded-lg.
func dialogStyle(t Theme) PartStyle {
	s := popupStyle(t)
	s.Radius = Set(t.radius(8))
	s.PaddingX, s.PaddingY = Set(t.spacing(12)), Set(t.spacing(12))
	return s
}

// Popups take Nuxt's 100 ms scale-in; dialogs take 200 ms.
const (
	popupMotion  = 100 * time.Millisecond
	dialogMotion = 200 * time.Millisecond
)

type popInKey struct{}

// popIn plays Nuxt's scale-in on a panel as it appears: from 95% and
// transparent to its place, scaled around (originX, originY), the side
// toward its anchor. Each newly shown panel plays it once. Closing is
// immediate: a popup's overlay leaves with it, so no exit can be drawn.
// It returns the progress for finishPanel's fade, 1 once settled or while motion is reduced.
func popIn(c *ui.Context, t Theme, panel ui.Element, d time.Duration, originX, originY float32) float32 {
	if !panel.Valid() {
		return 1
	}
	started := ui.Local(panel, "nui-pop-in", func() bool { return false })
	if !t.motion(c) {
		*started = true
		panel.Jump(popInKey{}, 1)
		return 1
	}
	if !*started {
		*started = true
		panel.Jump(popInKey{}, 0)
	}
	v := panel.AnimateWith(popInKey{}, 1, d, ui.EaseOut)
	scale := .95 + .05*v
	panel.TransformOrigin(originX, originY).Scale(scale, scale)
	return v
}

// finishPanel fades toward the resolved opacity and keeps Style as the final override.
func finishPanel(t Theme, id Component, panel ui.Element, base PartStyle, d Decoration, shown float32) {
	base = mergeStyle(base, t.components[id].UI[Panel])
	base = mergeStyle(base, d.UI[Panel])
	base.Opacity = Set(valueOr(base.Opacity, float32(1)) * shown)
	base.apply(panel)
	if d.Style != nil {
		d.Style(Panel, panel)
	}
}

func finishBackdrop(t Theme, id Component, back ui.Element, background ui.Color, d Decoration, shown float32) {
	style := mergeStyle(mergeStyle(PartStyle{Background: Set(background)}, t.components[id].UI[Backdrop]), d.UI[Backdrop])
	if style.Background.Set && shown < 1 {
		style.Background = Set(style.Background.Value.Alpha(shown))
	}
	style.apply(back)
	if d.Style != nil {
		d.Style(Backdrop, back)
	}
}

// popupOrigin is the panel's point nearest its anchor for a placement.
// Start and End of top and bottom placements follow the reading direction.
func popupOrigin(p Placement, rtl bool) (x, y float32) {
	switch p {
	case BottomEnd:
		x, y = 1, 0
	case BottomCenter:
		x, y = .5, 0
	case TopStart:
		x, y = 0, 1
	case TopEnd:
		x, y = 1, 1
	case TopCenter:
		x, y = .5, 1
	case RightStart:
		return 0, 0
	case RightCenter:
		return 0, .5
	case RightEnd:
		return 0, 1
	case LeftStart:
		return 1, 0
	case LeftCenter:
		return 1, .5
	case LeftEnd:
		return 1, 1
	}
	if rtl {
		x = 1 - x
	}
	return x, y
}

func stylePanel(c *ui.Context, panel, anchor ui.Element, t Theme, opts PopoverOptions, recipe PartStyle) {
	positionPanel(panel, anchor, t, opts)
	x, y := popupOrigin(opts.Placement, c.Direction() == ui.RightToLeft)
	shown := popIn(c, t, panel, popupMotion, x, y)
	finishPanel(t, ComponentPopover, panel, recipe, opts.Decoration, shown)
}
func positionPanel(panel, anchor ui.Element, t Theme, opts PopoverOptions) {
	at, self := ui.AnchorBottomStart, ui.AnchorTopStart
	switch opts.Placement {
	case BottomEnd:
		at, self = ui.AnchorBottomEnd, ui.AnchorTopEnd
	case TopStart:
		at, self = ui.AnchorTopStart, ui.AnchorBottomStart
	case TopEnd:
		at, self = ui.AnchorTopEnd, ui.AnchorBottomEnd
	case RightStart:
		at, self = ui.AnchorTopRight, ui.AnchorTopLeft
	case LeftStart:
		at, self = ui.AnchorTopLeft, ui.AnchorTopRight
	case BottomCenter:
		at, self = ui.AnchorBottom, ui.AnchorTop
	case TopCenter:
		at, self = ui.AnchorTop, ui.AnchorBottom
	case RightCenter:
		at, self = ui.AnchorRight, ui.AnchorLeft
	case RightEnd:
		at, self = ui.AnchorBottomRight, ui.AnchorBottomLeft
	case LeftCenter:
		at, self = ui.AnchorLeft, ui.AnchorRight
	case LeftEnd:
		at, self = ui.AnchorBottomLeft, ui.AnchorBottomRight
	}
	offset := t.spacing(8)
	if opts.Offset.Set {
		offset = opts.Offset.Value
	}
	panel.Column().MinWidth(0).AttachTo(anchor, at, self)
	// AttachTo uses opposing-margin differences, so symmetric margins cancel.
	// A single margin preserves the gap when MyGo flips at window edges.
	switch opts.Placement {
	case RightStart, RightCenter, RightEnd:
		panel.Margin(0, 0, 0, offset)
	case LeftStart, LeftCenter, LeftEnd:
		panel.Margin(0, offset, 0, 0)
	case TopStart, TopCenter, TopEnd:
		panel.Margin(0, 0, offset, 0)
	default:
		panel.Margin(offset, 0, 0, 0)
	}
	maxHeight := float32(320)
	if opts.MaxHeight.Set {
		maxHeight = opts.MaxHeight.Value
	}
	panel.MaxHeight(maxHeight)
	if opts.MatchAnchor {
		panel.MinWidth(anchor.Bounds().W)
	}
	if opts.Width.Set {
		panel.Width(opts.Width.Value)
	}
	if opts.Label != "" {
		panel.Label(opts.Label)
	}
}

// Popover is non-modal: outside presses dismiss it and continue to their target.
// The anchor and panel are build values; open is caller-owned window state.
func Popover(c *ui.Context, t Theme, anchor ui.Element, open *bool, opts PopoverOptions, content func(Theme)) ui.Element {
	return popover(c, t, anchor, open, opts, popupStyle(t), content)
}

// popover is Popover on a base recipe of the caller's, before theme and instance parts.
func popover(c *ui.Context, t Theme, anchor ui.Element, open *bool, opts PopoverOptions, recipe PartStyle, content func(Theme)) ui.Element {
	// Scope the native theme around the overlay, so the panel's final text and
	// layout styles inherit directly into its content instead of being reset by
	// a themed container inside the panel.
	host := ui.Themed(c, t.Native()).Background(ui.Transparent).Absolute().Size(0, 0).IgnoreInput(true)
	identity := ui.Local(host, "popover", func() popoverMemory { return popoverMemory{} })
	if opts.Mode > PopoverHover {
		panic("nui: invalid popover mode")
	}
	if open == nil {
		panic("nui: Popover requires open state")
	}
	if !anchor.Valid() || anchor.IsDisabled() {
		*open = false
		*identity = popoverMemory{}
		return ui.Element{}
	}
	anchorIntent := anchor.Hovered() || anchor.FocusVisible()
	if !anchorIntent {
		identity.suppressed = false
	}
	updateHover := func(intent bool) {
		if identity.suppressed {
			return
		}
		if identity.intent != intent {
			identity.intent = intent
			identity.since = c.Now()
		}
		if intent == *open {
			return
		}
		delay := opts.CloseDelay
		if intent {
			delay = opts.OpenDelay
		}
		remaining := delay - c.Now().Sub(identity.since)
		if remaining <= 0 {
			*open = intent
			c.After(time.Nanosecond)
		} else {
			c.After(remaining)
		}
	}
	if opts.Mode == PopoverHover && !*open {
		updateHover(anchorIntent)
	}
	if !*open {
		if identity.wasOpen && identity.focusWithin {
			anchor.Focus()
		}
		identity.wasOpen, identity.focusWithin = false, false
		return ui.Element{}
	}
	// Popup content is not part of the anchor's FieldGroup; its controls keep full corners.
	inner := t
	inner.group = groupPart{}
	var panel ui.Element
	host.Children(func() {
		panel = ui.PopoverBase(c.Key(identity), anchor, open, func(panel ui.Element) {
			stylePanel(c, panel, anchor, t, opts, recipe)
			if content != nil {
				content(inner)
			}
		})
	})
	if panel.Valid() && panel.Dismissed() {
		if !valueOr(opts.Dismissible, true) {
			*open = true
		} else {
			identity.suppressed = anchorIntent
		}
	}
	if opts.Mode == PopoverHover && *open {
		intent := anchorIntent || panel.Hovered() || panel.FocusWithin()
		// Include the small gap between anchor and panel, so crossing it does not close a hover card.
		a, b := anchor.Bounds(), panel.Bounds()
		x, y, _ := anchor.PointerPosition()
		x += a.X
		y += a.Y
		bridge := ui.Rect{X: min(a.X, b.X), Y: min(a.Y, b.Y), W: max(a.X+a.W, b.X+b.W) - min(a.X, b.X), H: max(a.Y+a.H, b.Y+b.H) - min(a.Y, b.Y)}
		intent = intent || (x >= bridge.X && x <= bridge.X+bridge.W && y >= bridge.Y && y <= bridge.Y+bridge.H)
		updateHover(intent)
	}
	if !*open && panel.FocusWithin() {
		anchor.Focus()
	}
	identity.wasOpen, identity.focusWithin = *open, panel.FocusWithin()
	return panel
}

func selectionPanel(c *ui.Context, panel, anchor ui.Element, t Theme, opts PopoverOptions) {
	if !opts.MaxHeight.Set {
		opts.MaxHeight = Set(float32(240))
	}
	opts.MatchAnchor = true
	// Tight viewport spacing is a recipe default, before theme/instance overrides.
	recipe := popupStyle(t)
	recipe.PaddingX, recipe.PaddingY, recipe.Gap = Set(t.spacing(4)), Set(t.spacing(4)), Set(float32(0))
	stylePanel(c, panel, anchor, t, opts, recipe)
}
