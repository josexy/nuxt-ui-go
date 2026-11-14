package nui

import (
	"runtime"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

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
	plain        bool                                 // Omit dividers when composed by Drawer or Sidebar.
	drawerLayout bool                                 // Scroll the Drawer content as one region.
	beforeHeader func(*ui.Context, Theme, ui.Element) // Build-local adornment used by Drawer.
}

// Modal is a convenience wrapper; use ModalWithOptions for additional options.
func Modal(c *ui.Context, t Theme, open *bool, title string, content func(Theme)) ui.Element {
	return ModalWithOptions(c, t, open, ModalOptions{Title: title}, content)
}
func ModalWithOptions(c *ui.Context, t Theme, open *bool, o ModalOptions, content func(Theme)) ui.Element {
	if open == nil {
		panic("nui: Modal requires open state")
	}
	host := ui.Themed(c, t.Native()).Background(ui.Transparent).Absolute().Size(0, 0).IgnoreInput(true)
	wasOpen := ui.Local(host, "modal-presence", func() bool { return false })
	if *wasOpen && !*open && o.ReturnFocus != nil {
		o.ReturnFocus.Focus(c)
	}
	*wasOpen = *open
	if !*open {
		return ui.Element{}
	}
	var panel ui.Element
	host.Children(func() {
		panel = ui.DialogBase(c.Key(wasOpen), open, func(back, panel ui.Element) {
			backdrop := t.Elevated.Alpha(.75)
			if !valueOr(o.Overlay, true) {
				backdrop = ui.Transparent
			}
			// Nuxt fades the overlay in while the panel scales in, both over 200 ms.
			motion := t
			if !valueOr(o.Transition, true) {
				motion.Motion.Reduce = true
			}
			shown := popIn(c, motion, panel, dialogMotion, .5, .5)
			finishBackdrop(t, ComponentModal, back, backdrop, o.Decoration, shown)
			w, h := c.Size()
			width := float32(512)
			if o.Width.Set {
				width = o.Width.Value
			}
			maximum := h - 64
			if o.MaxHeight.Set {
				maximum = min(maximum, o.MaxHeight.Value)
			}
			panel.Column().Width(min(width, max(0, w-32))).MaxHeight(max(0, maximum)).MinWidth(0).Clip()
			recipe := dialogStyle(t)
			if o.Fullscreen {
				maximum = h
				panel.Width(w).Height(h).MaxHeight(h)
				recipe.Radius, recipe.BorderWidth = Set(float32(0)), Set(float32(0))
			}
			recipe.PaddingX, recipe.PaddingY = Set(float32(0)), Set(float32(0))
			recipe.Gap = Set(float32(0))
			finishPanel(t, ComponentModal, panel, recipe, o.Decoration, shown)
			dialogContent(c, t, panel, open, o, ComponentModal, max(0, maximum), content)
		})
	})
	if panel.Valid() && panel.Dismissed() {
		if o.Dismissible.Set && !o.Dismissible.Value {
			*open = true
			if o.OnClosePrevent != nil {
				panel.SetChanged().OnChange(o.OnClosePrevent)
			}
		} else if o.OnClose != nil {
			panel.SetChanged().OnChange(o.OnClose)
		}
	}
	return panel
}
func dialogContent(c *ui.Context, t Theme, panel ui.Element, open *bool, o ModalOptions, id Component, available float32, content func(Theme)) {
	panel.Label(o.Title).Description(o.Description)
	if o.beforeHeader != nil {
		o.beforeHeader(c, t, panel)
	}
	t.group = groupPart{}
	if o.drawerLayout {
		container := ui.Scroll(c.Key("drawer-container")).MinWidth(0).MinHeight(0).Grow(1)
		finish(t, id, ContainerSlot, container, PartStyle{PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(16))}, o.Decoration)
		container.Children(func() { dialogSections(c, t, open, o, id, available, content) })
		return
	}
	dialogSections(c, t, open, o, id, available, content)
}

func dialogSections(c *ui.Context, t Theme, open *bool, o ModalOptions, id Component, available float32, content func(Theme)) {
	hasHeader := o.Header != nil || o.Title != "" || o.Description != "" || valueOr(o.CloseButton, true)
	if hasHeader {
		// Nuxt's header: p-4 sm:px-6, at least --ui-header-height (64) tall,
		// with divide-y lines between the regions that render.
		header := ui.Row(c.Key("header")).MinWidth(0).MinHeight(t.spacing(64)).AlignItems(ui.Center).Shrink(0)
		if !o.plain && (content != nil || o.Footer != nil) {
			header.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
		}
		paddingX, paddingY := t.spacing(24), t.spacing(16)
		if o.drawerLayout {
			header.MinHeight(t.spacing(32))
			paddingX, paddingY = 0, 0
		}
		finish(t, id, Header, header, PartStyle{PaddingX: Set(paddingX), PaddingY: Set(paddingY), Gap: Set(t.spacing(6))}, o.Decoration)
		header.Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(t.spacing(4)).Children(func() {
				if o.Header != nil {
					o.Header(c, t)
				} else {
					if o.Title != "" {
						finish(t, id, Label, ui.Text(c, o.Title).Role(ui.RoleHeading).Level(2), PartStyle{Foreground: Set(t.Highlighted), Weight: Set(600), FontSize: Set(t.text(16)), LineHeight: Set(float32(24.0 / 16))}, o.Decoration)
					}
					if o.Description != "" {
						finish(t, id, Description, ui.Text(c, o.Description), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
					}
				}
			})
			if !o.CloseButton.Set || o.CloseButton.Value {
				close := Button(c.Key("close"), t, "", dialogCloseOptions(c, o)).OnClick(func() {
					*open = false
					if o.OnClose != nil {
						o.OnClose()
					}
				})
				finish(t, id, Close, close, PartStyle{}, o.Decoration)
			}
		})
	}
	// Like Nuxt's, the body exists only with content.
	if content != nil {
		var body ui.Element
		padding := t.spacing(24)
		if o.drawerLayout {
			body = ui.Column(c.Key("body")).MinWidth(0)
			padding = 0
		} else {
			body = ui.Scroll(c.Key("body")).MinWidth(0).MinHeight(0).MaxHeight(available).Shrink(1)
			if id == ComponentSlideover || o.Fullscreen {
				body.Grow(1)
			}
		}
		finish(t, id, Body, body, PartStyle{PaddingX: Set(padding), PaddingY: Set(padding), Gap: Set(t.spacing(16))}, o.Decoration)
		body.Children(func() { content(t) })
	}
	if o.Footer != nil {
		footer := ui.Row(c.Key("footer")).Wrap().AlignItems(ui.Center).Justify(ui.End).Shrink(0)
		paddingX, paddingY := t.spacing(24), t.spacing(16)
		if o.drawerLayout {
			footer.Column().AlignItems(ui.Stretch)
			paddingX, paddingY = 0, 0
		}
		if !o.plain && (hasHeader || content != nil) {
			footer.BorderWidth(1, 0, 0, 0).BorderColor(t.Border)
		}
		finish(t, id, Footer, footer, PartStyle{PaddingX: Set(paddingX), PaddingY: Set(paddingY), Gap: Set(t.spacing(6))}, o.Decoration)
		footer.Children(func() { o.Footer(c, t) })
	}
}

func dialogCloseOptions(c *ui.Context, o ModalOptions) ButtonOptions {
	close := ButtonOptions{}
	if o.Close != nil {
		close = *o.Close
	}
	if close.Size == SizeDefault {
		close.Size = MD
	}
	if close.Color == ColorDefault {
		close.Color = Neutral
	}
	if close.Variant == VariantDefault {
		close.Variant = Ghost
	}
	if o.CloseLabel != "" {
		close.AccessibleLabel = o.CloseLabel
	}
	if close.AccessibleLabel == "" {
		close.AccessibleLabel = namedMessage(c, "closeNamed", o.Title)
	}
	if close.Leading == nil {
		close.Leading = o.CloseIcon
		if close.Leading == nil {
			close.Leading = lucide.X()
		}
	}
	return close
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
	Inset  bool
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
	host := ui.Themed(c, t.Native()).Background(ui.Transparent).Absolute().Size(0, 0).IgnoreInput(true)
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
	if !t.motion(c) || !valueOr(o.Transition, true) {
		presence.value = goal
		presence.from, presence.started = goal, c.Now()
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
	host.Children(func() {
		ui.Overlay(c, func() {
			back := ui.Box(c.Key(presence)).Absolute().Left(0).Top(0).Right(0).Bottom(0).Modal()
			backdrop := t.Elevated.Alpha(.75)
			if !valueOr(o.Overlay, true) {
				backdrop = ui.Transparent
			}
			finishBackdrop(t, ComponentSlideover, back, backdrop, o.Decoration, presence.value)
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
				} else if *open && o.OnClosePrevent != nil {
					o.OnClosePrevent()
				}
			})
			if !*open {
				back.Focusable().Focus()
			}
			back.Children(func() {
				w, h := c.Size()
				insetX, insetY := float32(0), float32(0)
				if o.Inset {
					insetX, insetY = min(t.spacing(16), max(0, w/2)), min(t.spacing(16), max(0, h/2))
				}
				availableW, availableH := max(0, w-2*insetX), max(0, h-2*insetY)
				extent := float32(448)
				if o.Extent.Set {
					extent = max(0, o.Extent.Value)
				}
				panel := ui.Column(c).Absolute().MinWidth(0).MinHeight(0).Clip().Role(ui.RoleDialog).Disabled(!*open).OnClick(func() {})
				vertical := o.Side == SideTop || o.Side == SideBottom
				if vertical {
					// Without an Extent, Nuxt's top and bottom panels fit their content.
					// They slide by their last measured height; at the first frame,
					// fully hidden either way, by the window's.
					if o.Extent.Set {
						extent = min(availableH, extent)
						panel.Height(extent)
					} else {
						panel.MaxHeight(availableH)
						extent = panel.Bounds().H
						if extent <= 0 {
							extent = availableH
						}
					}
				}
				x, y := float32(0), float32(0)
				switch o.Side {
				case SideLeft:
					extent = min(availableW, extent)
					panel.Left(insetX).Top(insetY).Bottom(insetY).Width(extent)
					x = -extent - insetX
				case SideTop:
					panel.Left(insetX).Right(insetX).Top(insetY)
					y = -extent - insetY
				case SideBottom:
					panel.Left(insetX).Right(insetX).Bottom(insetY)
					y = extent + insetY
				default:
					extent = min(availableW, extent)
					panel.Right(insetX).Top(insetY).Bottom(insetY).Width(extent)
					x = extent + insetX
				}
				panel.Translate(x*(1-presence.value), y*(1-presence.value))
				recipe := dialogStyle(t)
				recipe.Radius = Set(float32(0))
				if o.Inset {
					recipe.Radius = Set(t.radius(8))
				}
				recipe.PaddingX = Set(float32(0))
				recipe.PaddingY = Set(float32(0))
				recipe.Gap = Set(float32(0))
				finish(t, ComponentSlideover, Panel, panel, recipe, o.Decoration)
				result = panel
				available := availableH
				if vertical && o.Extent.Set {
					available = extent
				}
				panel.Children(func() {
					if content == nil && !o.drawerLayout {
						content = func(Theme) {}
					}
					dialogContent(c, t, panel, open, o.ModalOptions, ComponentSlideover, max(0, available), content)
				})
			})
			if *open && back.OverlayShortcut(0, ui.KeyEscape) {
				if valueOr(o.Dismissible, true) {
					close()
				} else if o.OnClosePrevent != nil {
					o.OnClosePrevent()
				}
			}
		})
	})
	return result
}

// tooltipShadow is Nuxt's shadow-sm, lighter than the popup panels' shadow-lg.
var tooltipShadow = ShadowTokens{Color: ui.RGBA(0, 0, 0, .10), Blur: 3, OffsetY: 1}

type TooltipOptions struct {
	Disabled      bool
	MaxWidth      Optional[float32]
	Placement     Optional[Placement]
	Offset        Optional[float32]
	Open          *bool
	Kbds          []string
	DelayDuration Optional[time.Duration]
	Decoration
}

type tooltipMemory struct {
	anchor                                                     ui.Handle
	lastOpen, nativeOpen, suppressed, open, hovered, warmHover bool
	hoverSince                                                 time.Time
}
type tooltipOwner struct {
	open        *bool
	identity    *tooltipMemory
	lastShown   time.Time
	customDelay bool
}

func Tooltip(c *ui.Context, t Theme, anchor ui.Element, text string, o TooltipOptions) ui.Element {
	host := ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	identity := ui.Local(host, "tooltip", func() tooltipMemory { return tooltipMemory{} })
	owner := ui.Local(c.Root(), "nui-tooltip-owner", func() tooltipOwner { return tooltipOwner{} })
	if owner.identity == identity && identity.lastOpen {
		owner.lastShown = c.Now()
	}
	if !anchor.Valid() || o.Disabled || anchor.IsDisabled() || (text == "" && len(o.Kbds) == 0) {
		*identity = tooltipMemory{}
		if o.Open != nil {
			*o.Open = false
		}
		return ui.Element{}
	}
	description := []string{}
	if text != "" {
		description = append(description, text)
	}
	for _, key := range o.Kbds {
		description = append(description, FormatKey(key, runtime.GOOS))
	}
	anchor.Description(strings.Join(description, " "))
	anchor.Bind(&identity.anchor)
	opened := o.Open
	if opened == nil {
		opened = &identity.open
	}
	dismiss := func() {
		*opened = false
		identity.lastOpen = false
		identity.suppressed = true
		c.After(time.Nanosecond)
	}
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
		tip.Row().Height(t.text(24)).AlignItems(ui.Center)
		originX, originY := popupOrigin(placement, c.Direction() == ui.RightToLeft)
		shown := popIn(c, t, tip, popupMotion, originX, originY)
		finishPanel(t, ComponentTooltip, tip, PartStyle{Background: Set(t.Background), Foreground: Set(t.Highlighted), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border), Radius: Set(t.radius(4)), PaddingX: Set(t.spacing(10)), PaddingY: Set(t.spacing(4)), Gap: Set(t.spacing(4)), Font: Set(t.Font), FontSize: Set(t.text(12)), LineHeight: Set(float32(16.0 / 12)), Shadow: Set(tooltipShadow)}, o.Decoration, shown)
		tip.Children(func() {
			if text != "" {
				finish(t, ComponentTooltip, Label, ui.Text(c, text).SingleLine().MinWidth(0), PartStyle{}, o.Decoration)
			}
			if len(o.Kbds) > 0 {
				keys := ui.Row(c.Key("kbds")).Shrink(0).AlignItems(ui.Center)
				finish(t, ComponentTooltip, Trailing, keys, PartStyle{Gap: Set(t.spacing(2))}, o.Decoration)
				keys.Children(func() {
					if text != "" {
						ui.Text(c, "·").MarginEnd(t.spacing(2)).Role(ui.RoleNone)
					}
					Kbd(c, t, o.Kbds, KbdOptions{Size: SM}).Gap(t.spacing(2))
				})
			}
		})
		if tip.PressedOutside() {
			dismiss()
		}
	}
	var native ui.Element
	sourceOpen := false
	hovered := anchor.Hovered()
	warm := !owner.lastShown.IsZero() && c.Now().Sub(owner.lastShown) <= 400*time.Millisecond
	if hovered != identity.hovered {
		identity.hovered, identity.hoverSince = hovered, c.Now()
		identity.warmHover = hovered && owner.customDelay && warm
	}
	if !o.DelayDuration.Set {
		native = ui.TooltipBase(c, anchor, func(ui.Element) {})
		if hovered && owner.customDelay && warm {
			identity.warmHover = true
		}
		sourceOpen = native.Valid() || identity.warmHover && !identity.suppressed
		if native.Valid() {
			identity.warmHover = false
		}
	} else {
		// TooltipBase has a fixed delay. Keep its default path, and use the
		// public clock/overlay APIs when an instance explicitly overrides it.
		sourceOpen = anchor.FocusVisible()
		if hovered && !sourceOpen && !identity.suppressed {
			remaining := max(0, o.DelayDuration.Value) - c.Now().Sub(identity.hoverSince)
			if remaining <= 0 || warm {
				sourceOpen = true
			} else {
				c.After(remaining)
			}
		}
	}
	claim := func() {
		if owner.identity != identity && owner.open != nil && *owner.open {
			*owner.open = false
			owner.identity.suppressed = true
			if hovered && tooltipContains(owner.identity.anchor.Bounds(c), anchor.Bounds()) {
				// Entering a nested trigger does not dismiss its enclosing one.
				owner.identity.lastOpen, owner.identity.nativeOpen, owner.identity.suppressed = false, false, false
			}
			c.After(time.Nanosecond)
		}
		owner.identity, owner.open, owner.lastShown = identity, opened, c.Now()
		owner.customDelay = o.DelayDuration.Set
	}
	if hovered && !anchor.FocusVisible() && owner.identity != nil && owner.identity != identity && owner.open != nil && *owner.open {
		previous := c.Resolve(owner.identity.anchor)
		if previous.Valid() && previous.Hovered() && tooltipContains(anchor.Bounds(), previous.Bounds()) {
			sourceOpen = false
		}
	}
	if !anchor.Hovered() && !anchor.FocusVisible() {
		identity.suppressed = false
	}
	if *opened != identity.lastOpen {
		identity.suppressed = !*opened
	}
	if sourceOpen != identity.nativeOpen {
		if !sourceOpen || !identity.suppressed {
			*opened = sourceOpen
		}
	}
	if anchor.Clicked() || anchor.Pressed() {
		*opened = false
		identity.suppressed = true
	}
	if identity.lastOpen && !*opened {
		c.After(time.Nanosecond)
	}
	identity.nativeOpen = sourceOpen
	identity.lastOpen = *opened
	if !*opened {
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
			dismiss()
		}
	})
	return tip
}

// Both elements must currently be hovered: their strict geometric containment
// then distinguishes an enclosing trigger from the deeper pointer target.
func tooltipContains(outer, inner ui.Rect) bool {
	return outer.W*outer.H > inner.W*inner.H && inner.X >= outer.X && inner.Y >= outer.Y && inner.X+inner.W <= outer.X+outer.W && inner.Y+inner.H <= outer.Y+outer.H
}
