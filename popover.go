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
	since              time.Time
	intent, suppressed bool
}

func panelStyle(t Theme) PartStyle {
	return PartStyle{Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14)), Foreground: Set(t.Text), Background: Set(t.Background), Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border), PaddingX: Set(t.spacing(12)), PaddingY: Set(t.spacing(12)), Gap: Set(t.spacing(8)), Shadow: Set(t.Shadow)}
}
func stylePanel(panel, anchor ui.Element, t Theme, opts PopoverOptions, recipe PartStyle) {
	positionPanel(panel, anchor, t, opts)
	finish(t, ComponentPopover, Panel, panel, recipe, opts.Decoration)
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
	host := ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
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
		return ui.Element{}
	}
	// Popup content is not part of the anchor's FieldGroup; its controls keep full corners.
	inner := t
	inner.group = groupPart{}
	panel := ui.PopoverBase(c.Key(identity), anchor, open, func(panel ui.Element) {
		stylePanel(panel, anchor, t, opts, panelStyle(t))
		ui.Themed(c, t.Native()).MinWidth(0).Children(func() {
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
	return panel
}

func selectionPanel(c *ui.Context, panel, anchor ui.Element, t Theme, opts PopoverOptions) {
	if !opts.MaxHeight.Set {
		opts.MaxHeight = Set(float32(240))
	}
	opts.MatchAnchor = true
	// Tight viewport spacing is a recipe default, before theme/instance overrides.
	recipe := panelStyle(t)
	recipe.PaddingX, recipe.PaddingY, recipe.Gap = Set(t.spacing(4)), Set(t.spacing(4)), Set(float32(0))
	stylePanel(panel, anchor, t, opts, recipe)
}
