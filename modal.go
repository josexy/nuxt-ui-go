package nui

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type ModalOptions struct {
	Title, Description, CloseLabel string
	Width, MaxHeight               Optional[float32]
	Dismissible, CloseButton       Optional[bool]
	Overlay                        Optional[bool] // Defaults to true; false hides the backdrop, not the modal focus scope.
	Fullscreen                     bool
	Header, Footer                 func(*ui.Context, Theme)
	OnClose                        func()
	ReturnFocus                    *ui.Handle
	Decoration
	plain bool // Drawer: Nuxt's drawer has no dividers between its regions.
}

// Modal is a convenience wrapper; use ModalWithOptions for additional options.
func Modal(c *ui.Context, t Theme, open *bool, title string, content func(Theme)) ui.Element {
	return ModalWithOptions(c, t, open, ModalOptions{Title: title}, content)
}
func ModalWithOptions(c *ui.Context, t Theme, open *bool, o ModalOptions, content func(Theme)) ui.Element {
	if open == nil {
		panic("nui: Modal requires open state")
	}
	host := ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	wasOpen := ui.Local(host, "modal-presence", func() bool { return false })
	if *wasOpen && !*open && o.ReturnFocus != nil {
		o.ReturnFocus.Focus(c)
	}
	*wasOpen = *open
	if !*open {
		return ui.Element{}
	}
	panel := ui.DialogBase(c.Key(wasOpen), open, func(back, panel ui.Element) {
		backdrop := t.Elevated.Alpha(.75)
		if !valueOr(o.Overlay, true) {
			backdrop = ui.Transparent
		}
		// Nuxt fades the overlay in while the panel scales in, both over 200 ms.
		shown := popIn(c, t, panel, dialogMotion, .5, .5)
		style := mergeStyle(mergeStyle(PartStyle{Background: Set(backdrop)}, t.components[ComponentModal].UI[Backdrop]), o.UI[Backdrop])
		if bg := style.Background.Value; style.Background.Set && shown < 1 {
			style.Background = Set(bg.Alpha(float32(bg.A) / 255 * shown))
		}
		style.apply(back)
		if o.Style != nil {
			o.Style(Backdrop, back)
		}
		w, h := c.Size()
		width := float32(512)
		if o.Width.Set {
			width = o.Width.Value
		}
		maximum := h - 64
		if o.MaxHeight.Set {
			maximum = min(maximum, o.MaxHeight.Value)
		}
		panel.Column().Width(min(width, max(0, w-32))).MaxHeight(max(0, maximum)).MinWidth(0)
		recipe := dialogStyle(t)
		if o.Fullscreen {
			maximum = h
			panel.Width(w).Height(h).MaxHeight(h)
			recipe.Radius, recipe.BorderWidth = Set(float32(0)), Set(float32(0))
		}
		recipe.PaddingX, recipe.PaddingY = Set(float32(0)), Set(float32(0))
		finish(t, ComponentModal, Panel, panel, recipe, o.Decoration)
		dialogContent(c, t, panel, open, o, ComponentModal, max(0, maximum), content)
	})
	if panel.Valid() && panel.Dismissed() {
		if o.Dismissible.Set && !o.Dismissible.Value {
			*open = true
		} else if o.OnClose != nil {
			panel.SetChanged().OnChange(o.OnClose)
		}
	}
	return panel
}
func dialogContent(c *ui.Context, t Theme, panel ui.Element, open *bool, o ModalOptions, id Component, available float32, content func(Theme)) {
	panel.Label(o.Title).Description(o.Description)
	container := ui.Themed(c, t.Native()).Background(ui.Transparent).MinWidth(0).MinHeight(0).MaxHeight(available)
	if id == ComponentSlideover {
		container.Grow(1)
	}
	container.Children(func() {
		// Nuxt's header: p-4 sm:px-6, at least --ui-header-height (64) tall,
		// with divide-y lines between the regions that render.
		header := ui.Row(c).MinWidth(0).MinHeight(t.spacing(64)).AlignItems(ui.Center).Shrink(0)
		if !o.plain && (content != nil || o.Footer != nil) {
			header.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
		}
		finish(t, id, Header, header, PartStyle{PaddingX: Set(t.spacing(24)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(6))}, o.Decoration)
		header.Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(t.spacing(4)).Children(func() {
				if o.Header != nil {
					o.Header(c, t)
				} else {
					finish(t, id, Label, ui.Text(c, o.Title).Role(ui.RoleHeading).Level(2), PartStyle{Foreground: Set(t.Highlighted), Weight: Set(600), FontSize: Set(t.text(16)), LineHeight: Set(float32(24.0 / 16))}, o.Decoration)
					if o.Description != "" {
						finish(t, id, Description, ui.Text(c, o.Description), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
					}
				}
			})
			if !o.CloseButton.Set || o.CloseButton.Value {
				label := o.CloseLabel
				if label == "" {
					label = namedMessage(c, "closeNamed", o.Title)
				}
				close := Button(c, t, "", ButtonOptions{Leading: lucide.X(), AccessibleLabel: label, Variant: Ghost, Color: Neutral, Square: true, Size: MD}).OnClick(func() {
					*open = false
					if o.OnClose != nil {
						o.OnClose()
					}
				})
				finish(t, id, Close, close, PartStyle{}, o.Decoration)
			}
		})
		// Like Nuxt's, the body exists only with content.
		if content != nil {
			body := ui.Scroll(c).MinWidth(0).MinHeight(0).MaxHeight(available).Shrink(1)
			if id == ComponentSlideover {
				body.Grow(1)
			}
			finish(t, id, Body, body, PartStyle{PaddingX: Set(t.spacing(24)), PaddingY: Set(t.spacing(24)), Gap: Set(t.spacing(16))}, o.Decoration)
			body.Children(func() { content(t) })
		}
		if o.Footer != nil {
			footer := ui.Row(c).Wrap().AlignItems(ui.Center).Justify(ui.End).Shrink(0)
			if !o.plain {
				footer.BorderWidth(1, 0, 0, 0).BorderColor(t.Border)
			}
			finish(t, id, Footer, footer, PartStyle{PaddingX: Set(t.spacing(24)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(6))}, o.Decoration)
			footer.Children(func() { o.Footer(c, t) })
		}
	})
}

type Side uint8

const (
	SideRight Side = iota
	SideLeft
	SideTop
	SideBottom
)

type SlideoverOptions struct {
	ModalOptions
	Side   Side
	Extent Optional[float32]
}

type slidePresence struct {
	target, mounted bool
	from, value     float32
	started         time.Time
}

// Slideover retains its modal scope through the exit, then unmounts to restore focus.
// Call it every frame, including while open is false.
func Slideover(c *ui.Context, t Theme, open *bool, o SlideoverOptions, content func(Theme)) ui.Element {
	if open == nil {
		panic("nui: Slideover requires open state")
	}
	host := ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	presence := ui.Local(host, "presence", func() slidePresence { return slidePresence{} })
	if presence.target != *open {
		presence.from = presence.value
		presence.target = *open
		presence.started = c.Now()
	}
	goal := float32(0)
	if *open {
		goal = 1
	}
	duration := 200 * time.Millisecond
	fraction := float32(c.Now().Sub(presence.started)) / float32(duration)
	fraction = max(0, min(1, fraction))
	eased := 1 - (1-fraction)*(1-fraction)*(1-fraction)
	presence.value = presence.from + (goal-presence.from)*eased
	if t.Motion.Reduce {
		presence.value = goal
	}
	if presence.value != goal {
		c.AnimationFrame()
	}
	if !*open && presence.value == 0 {
		if presence.mounted && o.ReturnFocus != nil {
			o.ReturnFocus.Focus(c)
		}
		presence.mounted = false
		return ui.Element{}
	}
	presence.mounted = true
	var result ui.Element
	ui.Overlay(c, func() {
		back := ui.Box(c.Key(presence)).Absolute().Left(0).Top(0).Right(0).Bottom(0).Modal()
		backdrop := t.Elevated.Alpha(.75 * presence.value)
		if !valueOr(o.Overlay, true) {
			backdrop = ui.Transparent
		}
		finish(t, ComponentSlideover, Backdrop, back, PartStyle{Background: Set(backdrop)}, o.Decoration)
		close := func() {
			if !*open {
				return
			}
			*open = false
			if o.OnClose != nil {
				o.OnClose()
			}
		}
		back.OnClick(func() {
			if !o.Dismissible.Set || o.Dismissible.Value {
				close()
			}
		})
		if !*open {
			back.Focusable().Focus()
		}
		back.Children(func() {
			w, h := c.Size()
			extent := float32(448)
			if o.Extent.Set {
				extent = max(0, o.Extent.Value)
			}
			panel := ui.Column(c).Absolute().MinWidth(0).MinHeight(0).Role(ui.RoleDialog).Disabled(!*open).OnClick(func() {})
			vertical := o.Side == SideTop || o.Side == SideBottom
			if vertical {
				// Without an Extent, Nuxt's top and bottom panels fit their content.
				// They slide by their last measured height; at the first frame,
				// fully hidden either way, by the window's.
				if o.Extent.Set {
					extent = min(h, extent)
					panel.Height(extent)
				} else {
					panel.MaxHeight(h)
					extent = panel.Bounds().H
					if extent <= 0 {
						extent = h
					}
				}
			}
			x, y := float32(0), float32(0)
			switch o.Side {
			case SideLeft:
				extent = min(w, extent)
				panel.Left(0).Top(0).Bottom(0).Width(extent)
				x = -extent
			case SideTop:
				panel.Left(0).Right(0).Top(0)
				y = -extent
			case SideBottom:
				panel.Left(0).Right(0).Bottom(0)
				y = extent
			default:
				extent = min(w, extent)
				panel.Right(0).Top(0).Bottom(0).Width(extent)
				x = extent
			}
			panel.Translate(x*(1-presence.value), y*(1-presence.value))
			recipe := dialogStyle(t)
			recipe.Radius = Set(float32(0))
			recipe.PaddingX = Set(float32(0))
			recipe.PaddingY = Set(float32(0))
			finish(t, ComponentSlideover, Panel, panel, recipe, o.Decoration)
			result = panel
			available := h
			if vertical && o.Extent.Set {
				available = extent
			}
			panel.Children(func() {
				dialogContent(c, t, panel, open, o.ModalOptions, ComponentSlideover, max(0, available), content)
			})
		})
		if *open && back.OverlayShortcut(0, ui.KeyEscape) && (!o.Dismissible.Set || o.Dismissible.Value) {
			close()
		}
	})
	return result
}

// tooltipShadow is Nuxt's shadow-sm, lighter than the popup panels' shadow-lg.
var tooltipShadow = ShadowTokens{Color: ui.RGBA(0, 0, 0, .10), Blur: 3, OffsetY: 1}

type TooltipOptions struct {
	Disabled  bool
	MaxWidth  Optional[float32]
	Placement Optional[Placement]
	Offset    Optional[float32]
	Open      *bool
	Kbds      []string
	Decoration
}

type tooltipMemory struct {
	lastOpen, nativeOpen, suppressed bool
}
type tooltipOwner struct {
	open     *bool
	identity *tooltipMemory
}

func Tooltip(c *ui.Context, t Theme, anchor ui.Element, text string, o TooltipOptions) ui.Element {
	host := ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	identity := ui.Local(host, "tooltip", func() tooltipMemory { return tooltipMemory{} })
	owner := ui.Local(c.Root(), "nui-tooltip-owner", func() tooltipOwner { return tooltipOwner{} })
	if !anchor.Valid() || o.Disabled || anchor.IsDisabled() || (text == "" && len(o.Kbds) == 0) {
		*identity = tooltipMemory{}
		if o.Open != nil {
			*o.Open = false
		}
		return ui.Element{}
	}
	anchor.Description(text)
	build := func(tip ui.Element) {
		tip.Key(identity)
		width := float32(320)
		if o.MaxWidth.Set {
			width = o.MaxWidth.Value
		}
		tip.MaxWidth(width).IgnoreInput(true)
		placement := valueOr(o.Placement, BottomCenter)
		positionPanel(tip, anchor, t, PopoverOptions{Placement: placement, Offset: o.Offset})
		// Nuxt's tooltip: one truncated 24 DIP line, rounded-sm, px-2.5, shadow-sm.
		tip.Height(t.text(24)).Justify(ui.Center)
		finish(t, ComponentTooltip, Panel, tip, PartStyle{Background: Set(t.Background), Foreground: Set(t.Highlighted), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border), Radius: Set(t.radius(4)), PaddingX: Set(t.spacing(10)), PaddingY: Set(t.spacing(4)), FontSize: Set(t.text(12)), LineHeight: Set(float32(16.0 / 12)), Shadow: Set(tooltipShadow)}, o.Decoration)
		originX, originY := popupOrigin(placement, c.Direction() == ui.RightToLeft)
		popIn(c, t, tip, popupMotion, originX, originY)
		tip.Children(func() {
			ui.Row(c).MinWidth(0).Gap(t.spacing(4)).AlignItems(ui.Center).Children(func() {
				if text != "" {
					finish(t, ComponentTooltip, Label, ui.Text(c, text).SingleLine().MinWidth(0), PartStyle{}, o.Decoration)
				}
				if len(o.Kbds) > 0 {
					ui.Row(c.Key("kbds")).Shrink(0).AlignItems(ui.Center).Gap(t.spacing(2)).Children(func() {
						if text != "" {
							ui.Text(c, "·").MarginEnd(t.spacing(2)).Role(ui.RoleNone)
						}
						Kbd(c, t, o.Kbds, KbdOptions{Size: XS})
					})
				}
			})
		})
	}
	native := ui.TooltipBase(c, anchor, func(ui.Element) {})
	claim := func() {
		if owner.identity != identity && owner.open != nil && *owner.open {
			*owner.open = false
			c.After(time.Nanosecond)
		}
		owner.identity, owner.open = identity, o.Open
	}
	if o.Open == nil {
		if native.Valid() {
			claim()
			build(native)
		}
		return native
	}
	if !anchor.Hovered() && !anchor.FocusVisible() {
		identity.suppressed = false
	}
	if *o.Open != identity.lastOpen {
		identity.suppressed = !*o.Open
	}
	if native.Valid() != identity.nativeOpen {
		if !native.Valid() || !identity.suppressed {
			*o.Open = native.Valid()
		}
	}
	if anchor.Clicked() || anchor.Pressed() {
		*o.Open = false
		identity.suppressed = true
	}
	identity.nativeOpen = native.Valid()
	identity.lastOpen = *o.Open
	if !*o.Open {
		if native.Valid() {
			native.Hidden(true)
		}
		return ui.Element{}
	}
	claim()
	if native.Valid() {
		build(native)
		return native
	}
	var tip ui.Element
	ui.Overlay(c.Key(identity), func() {
		tip = ui.Box(c).Role(ui.RoleTooltip)
		build(tip)
		if tip.OverlayShortcut(0, ui.KeyEscape) {
			*o.Open = false
			identity.lastOpen = false
			identity.suppressed = true
			c.After(time.Nanosecond)
		}
	})
	return tip
}
