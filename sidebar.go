package nui

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type SidebarMode uint8

const (
	SidebarSlideover SidebarMode = iota
	SidebarModal
	SidebarDrawer
)

// SidebarState belongs to one sidebar in one window. Width is the remembered
// expanded width in DIP; unset uses SidebarOptions.Width. Applications may save
// Width and their collapsed value and restore them before building the sidebar.
// Open controls the narrow-window overlay independently of desktop collapse.
type SidebarState struct {
	Width Optional[float32]
	Open  bool
}

// SidebarScope is supplied synchronously to custom regions. Its actions may be
// used in input callbacks; never retain Context or Elements from a builder.
type SidebarScope struct {
	Collapsed, Open, Mobile bool
	Collapse                func(bool)
	SetOpen                 func(bool)
}

type SidebarOptions struct {
	Title, ToggleLabel, ResizeLabel string
	Description                     string
	Width, CollapsedWidth           Optional[float32]
	MinWidth, MaxWidth              Optional[float32]
	// SideLeft (default) is the leading side; SideRight is trailing. Both mirror in RTL.
	Side Optional[Side]
	// Breakpoint enables the overlay below this window width in DIP. Unset/zero
	// keeps the original desktop-only behavior. Mode defaults to SidebarSlideover.
	Breakpoint          Optional[float32]
	Mode                SidebarMode
	State               *SidebarState
	Resizable, Disabled bool
	Collapsible, Toggle Optional[bool]
	// NavigationKey is an application-owned route/selection ID. Changes close the
	// overlay when AutoClose is true (the default), without coupling to a router.
	NavigationKey  string
	AutoClose      Optional[bool]
	Header, Footer func(*ui.Context, Theme)
	// HeaderContent replaces the default title; FooterContent replaces Footer.
	// ToggleContent replaces the desktop collapse / narrow-window open button.
	HeaderContent, FooterContent, ToggleContent func(*ui.Context, Theme, SidebarScope)
	// Actions adds header actions while expanded or in the narrow-window menu.
	Actions           func(*ui.Context, Theme, SidebarScope)
	OnResize          func(float32)
	OnCollapsedChange func(bool)
	OnOpenChange      func(bool)
	// ReturnFocus overrides the default built-in toggle when using an external opener.
	ReturnFocus *ui.Handle
	Decoration
}

type sidebarMemory struct {
	state               SidebarState
	trigger, resize     ui.Handle
	root                ui.Handle
	drawer              DrawerState
	dragging            bool
	startX, startWidth  float32
	navigation          string
	initialized, mobile bool
	right               bool
}

// Sidebar preserves the original collapsed-pointer API. Supply State to control
// the overlay externally or retain width across unmounts. Otherwise local widget
// state remembers it for the lifetime of this keyed Sidebar. Call every build,
// including while its overlay is closed, to finish exits and restore focus.
func Sidebar(c *ui.Context, t Theme, collapsed *bool, o SidebarOptions, content func(*ui.Context, Theme, bool)) ui.Element {
	if collapsed == nil {
		panic("nui: Sidebar requires collapsed state")
	}
	side := valueOr(o.Side, SideLeft)
	if (side != SideLeft && side != SideRight) || o.Mode > SidebarDrawer {
		panic("nui: invalid Sidebar side or mode")
	}
	initial, compact := valueOr(o.Width, float32(256)), valueOr(o.CollapsedWidth, float32(64))
	minimum, maximum := valueOr(o.MinWidth, float32(160)), valueOr(o.MaxWidth, float32(480))
	breakpoint := valueOr(o.Breakpoint, float32(0))
	for _, v := range []float32{initial, compact, minimum, maximum, breakpoint} {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) || v < 0 {
			panic("nui: Sidebar dimensions must be finite and nonnegative")
		}
	}
	if maximum < minimum || (o.Resizable && compact > minimum) {
		panic("nui: invalid Sidebar width constraints")
	}
	root := ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).MinHeight(0).Shrink(0).Role(ui.RoleGroup).Label(o.Title).Disabled(o.Disabled)
	m := ui.Local(root, "sidebar", func() sidebarMemory { return sidebarMemory{} })
	root.Bind(&m.root)
	containedFocus := m.root.FocusWithin(c)
	s := o.State
	if s == nil {
		s = &m.state
	}
	expanded := valueOr(s.Width, initial)
	if math.IsNaN(float64(expanded)) || math.IsInf(float64(expanded), 0) || expanded < 0 {
		panic("nui: Sidebar state width must be finite and nonnegative")
	}
	if o.Resizable || o.MinWidth.Set || o.MaxWidth.Set {
		expanded = max(minimum, min(maximum, expanded))
	}
	windowWidth, _ := c.Size()
	mobile := breakpoint > 0 && windowWidth < breakpoint
	wasMobile := m.mobile
	disabled := root.IsDisabled()
	collapsible := valueOr(o.Collapsible, true)
	changeCollapse := func(v bool) {
		if disabled || !collapsible || v == *collapsed {
			return
		}
		*collapsed = v
		if o.OnCollapsedChange != nil {
			o.OnCollapsedChange(v)
		}
	}
	changeOpen := func(v bool) {
		if disabled || s.Open == v {
			return
		}
		s.Open = v
		if o.OnOpenChange != nil {
			o.OnOpenChange(v)
		}
	}
	// Route and breakpoint changes are actions, so callbacks run after building.
	if m.initialized && ((valueOr(o.AutoClose, true) && m.navigation != o.NavigationKey) || (m.mobile && !mobile)) && s.Open {
		root.SetChanged().OnChange(func() { changeOpen(false) })
	}
	m.navigation, m.mobile, m.initialized = o.NavigationKey, mobile, true
	if disabled || mobile || !o.Resizable {
		m.dragging = false
	}
	scope := SidebarScope{Collapsed: *collapsed && !mobile, Open: s.Open, Mobile: mobile, Collapse: changeCollapse, SetOpen: changeOpen}
	physicalRight := (side == SideRight) != (c.Direction() == ui.RightToLeft)
	if m.right != physicalRight {
		m.dragging = false
	}
	m.right = physicalRight
	w := expanded
	if scope.Collapsed {
		w = compact
	}
	restoreFocus := containedFocus && ((mobile && !wasMobile && !s.Open) || (!mobile && w == 0 && m.root.Bounds(c).W > 0))
	root.Width(w)
	padding := t.spacing(16)
	if scope.Collapsed {
		padding = t.spacing(8)
	}
	if mobile {
		root.Width(0).Height(0).AlignSelf(ui.Start)
		// The external trigger can live in an application header instead.
		if valueOr(o.Toggle, true) {
			root.Width(t.spacing(48)).Height(t.spacing(48))
		}
	} else if w == 0 {
		root.BorderWidth(0)
	} else if side == SideRight {
		root.BorderStartWidth(1).BorderColor(t.Border)
	} else {
		root.BorderEndWidth(1).BorderColor(t.Border)
	}
	finish(t, ComponentSidebar, Root, root, PartStyle{Background: Set(t.Background)}, o.Decoration)
	root.Children(func() {
		if mobile {
			if valueOr(o.Toggle, true) {
				sidebarToggle(c.Key("toggle"), t, m, o, scope, physicalRight)
			}
			return
		}
		if w == 0 {
			if o.Resizable {
				sidebarResize(c.Key("resize"), t, m, s, collapsed, o, physicalRight, w, initial, compact, minimum, maximum, changeCollapse)
			}
			return
		}
		hasHeading := o.HeaderContent != nil || (!scope.Collapsed && (o.Title != "" || o.Description != ""))
		hasActions := o.Actions != nil && !scope.Collapsed
		showToggle := valueOr(o.Toggle, true) && collapsible
		if hasHeading || hasActions || showToggle {
			header := ui.Row(c.Key("header")).MinWidth(0).MinHeight(t.spacing(64)).Shrink(0).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
			finish(t, ComponentSidebar, Header, header, PartStyle{PaddingX: Set(padding), Gap: Set(t.spacing(6))}, o.Decoration)
			header.Children(func() {
				if hasHeading {
					ui.Column(c.Key("heading")).Grow(1).MinWidth(0).Children(func() {
						if o.HeaderContent != nil {
							o.HeaderContent(c, t, scope)
						} else {
							sidebarHeading(c, t, o)
						}
					})
				} else {
					header.Justify(ui.Center)
				}
				if hasActions {
					sidebarActions(c.Key("actions"), t, o, scope)
				}
				if showToggle {
					sidebarToggle(c.Key("toggle"), t, m, o, scope, physicalRight)
				}
			})
		}
		if o.Header != nil && o.HeaderContent == nil {
			ui.Column(c.Key("legacy-header")).MinWidth(0).Padding(0, 8).Children(func() { o.Header(c, t) })
		}
		body := ui.Scroll(c.Key("body")).Grow(1).MinWidth(0).MinHeight(0)
		finish(t, ComponentSidebar, Body, body, PartStyle{PaddingX: Set(padding), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(16))}, o.Decoration)
		if content != nil {
			body.Children(func() { content(c, t, scope.Collapsed) })
		}
		if o.Footer != nil || o.FooterContent != nil {
			footer := ui.Column(c.Key("footer")).MinWidth(0).Shrink(0).BorderWidth(1, 0, 0, 0).BorderColor(t.Border)
			finish(t, ComponentSidebar, Footer, footer, PartStyle{PaddingX: Set(padding), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(6))}, o.Decoration)
			footer.Children(func() { sidebarFooter(c, t, o, scope) })
		}
		if o.Resizable {
			sidebarResize(c.Key("resize"), t, m, s, collapsed, o, physicalRight, w, initial, compact, minimum, maximum, changeCollapse)
		}
	})
	// Overlay hosts remain under the same keyed root through open/close frames.
	root.Children(func() {
		sidebarOverlay(c.Key("overlay"), t, m, s, o, scope, expanded, physicalRight, disabled, content)
	})
	if restoreFocus && !disabled {
		switch {
		case o.ReturnFocus != nil:
			o.ReturnFocus.Focus(c)
		case valueOr(o.Toggle, true) && o.ToggleContent == nil && (mobile || (w > 0 && collapsible)):
			m.trigger.Focus(c)
		case !mobile && o.Resizable:
			m.resize.Focus(c)
		}
	}
	return root
}

func sidebarHeading(c *ui.Context, t Theme, o SidebarOptions) {
	if o.Title != "" {
		finish(t, ComponentSidebar, Label, ui.Text(c.Key("title"), o.Title).SingleLine().MinWidth(0), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(16)), LineHeight: Set(float32(24.0 / 16)), Weight: Set(600)}, o.Decoration)
	}
	if o.Description != "" {
		finish(t, ComponentSidebar, Description, ui.Text(c.Key("description"), o.Description).SingleLine().MinWidth(0), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, o.Decoration)
	}
}

func sidebarActions(c *ui.Context, t Theme, o SidebarOptions, scope SidebarScope) {
	actions := ui.Row(c).MinWidth(0).Shrink(0).AlignItems(ui.Center)
	finish(t, ComponentSidebar, Actions, actions, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
	actions.Children(func() { o.Actions(c, t, scope) })
}

func sidebarToggle(c *ui.Context, t Theme, m *sidebarMemory, o SidebarOptions, scope SidebarScope, right bool) {
	if o.ToggleContent != nil {
		o.ToggleContent(c, t, scope)
		return
	}
	label := o.ToggleLabel
	if label == "" {
		label = message(c, "toggleSidebar")
	}
	icon := lucide.PanelLeftClose()
	switch {
	case scope.Mobile:
		icon = lucide.Menu()
	case scope.Collapsed && right:
		icon = lucide.PanelRightOpen()
	case scope.Collapsed:
		icon = lucide.PanelLeftOpen()
	case right:
		icon = lucide.PanelRightClose()
	}
	button := Button(c, t, "", ButtonOptions{Leading: icon, AccessibleLabel: label, Square: true, Variant: Ghost, Color: Neutral}).Bind(&m.trigger)
	if scope.Mobile {
		button.Expanded(scope.Open).OnClick(func() { scope.SetOpen(!scope.Open) })
	} else {
		button.Expanded(!scope.Collapsed).OnClick(func() { scope.Collapse(!scope.Collapsed) })
	}
	finish(t, ComponentSidebar, Trigger, button, PartStyle{}, o.Decoration)
}

func sidebarFooter(c *ui.Context, t Theme, o SidebarOptions, scope SidebarScope) {
	if o.FooterContent != nil {
		o.FooterContent(c, t, scope)
	} else if o.Footer != nil {
		o.Footer(c, t)
	}
}

func sidebarResize(c *ui.Context, t Theme, m *sidebarMemory, s *SidebarState, collapsed *bool, o SidebarOptions, right bool, shown, initial, compact, minimum, maximum float32, collapse func(bool)) {
	label := o.ResizeLabel
	if label == "" {
		label = message(c, "resizeSidebar")
	}
	handle := ui.Box(c).Absolute().Top(0).Bottom(0).Width(8).Focusable().FocusRing(true).Role(ui.RoleSplitter).Label(label).Cursor(ui.CursorResizeEW).Bind(&m.resize)
	if right {
		handle.Left(-4)
	} else {
		handle.Right(-4)
	}
	low := minimum
	if valueOr(o.Collapsible, true) {
		low = compact
	}
	handle.Range(float64(low), float64(maximum), float64(shown)).Step(10)
	set := func(v float32) {
		if valueOr(o.Collapsible, true) && v < compact+4 {
			collapse(true)
			return
		}
		v = max(minimum, min(maximum, v))
		collapse(false)
		if valueOr(s.Width, initial) != v {
			s.Width = Set(v)
			if o.OnResize != nil {
				o.OnResize(v)
			}
		}
	}
	direction := float32(1)
	if right {
		direction = -1
	}
	handle.HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind == ui.InputPointerUp {
			was := m.dragging
			m.dragging = false
			return was
		}
		if ev.Button != 0 || (ev.Kind != ui.InputPointerDown && ev.Kind != ui.InputPointerMove) {
			return false
		}
		x := m.resize.Bounds(c).X + ev.X
		if ev.Kind == ui.InputPointerDown {
			m.resize.Focus(c)
			if ev.Clicks == 2 {
				m.dragging = false
				set(initial)
				return true
			}
			m.startX, m.startWidth, m.dragging = x, shown, true
			return true
		}
		if !m.dragging {
			return false
		}
		set(m.startWidth + direction*(x-m.startX))
		return true
	})
	for _, mods := range []ui.Modifiers{0, ui.Shift} {
		step := float32(10)
		if mods == ui.Shift {
			step = 50
		}
		handle.OnShortcut(mods, ui.KeyLeft, func() { set(shown - direction*step) }).OnShortcut(mods, ui.KeyRight, func() { set(shown + direction*step) })
	}
	handle.OnShortcut(0, ui.KeyHome, func() { set(low) }).OnShortcut(0, ui.KeyEnd, func() { set(maximum) }).OnShortcut(0, ui.KeyEnter, func() { collapse(!*collapsed) })
	color := t.Border
	if handle.Hovered() || handle.FocusVisible() || m.dragging {
		color = t.Primary
	}
	handle.Children(func() {
		// The handle is placed from inside the root's one-DIP edge border, so
		// mirror the line with it to draw over that border, not beside it.
		line := ui.Box(c).Absolute().Top(0).Bottom(0).Width(1).Background(color).IgnoreInput(true)
		if right {
			line.Left(3)
		} else {
			line.Right(3)
		}
	})
	finish(t, ComponentSidebar, SeparatorSlot, handle, PartStyle{}, o.Decoration)
}

func sidebarOverlay(c *ui.Context, t Theme, m *sidebarMemory, s *SidebarState, o SidebarOptions, scope SidebarScope, width float32, right, disabled bool, content func(*ui.Context, Theme, bool)) {
	scope.Collapsed = false
	if !scope.Mobile {
		t.Motion.Reduce = true
	}
	// Keep the desktop content as the only active copy after a breakpoint change.
	open := s.Open && scope.Mobile && !disabled
	mo := ModalOptions{Title: o.Title, Description: o.Description, Fullscreen: true, CloseLabel: message(c, "closeSidebar"), CloseButton: Set(true), OnClose: func() { scope.SetOpen(false) }, plain: true}
	if valueOr(o.Toggle, true) && o.ToggleContent == nil {
		mo.ReturnFocus = &m.trigger
	}
	if o.ReturnFocus != nil {
		mo.ReturnFocus = o.ReturnFocus
	}
	mo.Header = func(c *ui.Context, t Theme) {
		ui.Row(c.Key("sidebar-header-content")).MinWidth(0).AlignItems(ui.Center).Gap(t.spacing(6)).Children(func() {
			ui.Column(c.Key("heading")).Grow(1).MinWidth(0).Children(func() {
				if o.HeaderContent != nil {
					o.HeaderContent(c, t, scope)
				} else if o.Header != nil {
					o.Header(c, t)
				} else {
					sidebarHeading(c, t, o)
				}
			})
			if o.Actions != nil {
				sidebarActions(c.Key("actions"), t, o, scope)
			}
		})
	}
	if o.Footer != nil || o.FooterContent != nil {
		mo.Footer = func(c *ui.Context, t Theme) { sidebarFooter(c, t, o, scope) }
	}
	mo.Style = func(slot Slot, e ui.Element) {
		base := PartStyle{}
		if slot == Panel {
			e.Disabled(disabled)
		}
		switch slot {
		case Header:
			e.BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
			base = PartStyle{PaddingX: Set(t.spacing(16)), PaddingY: Set(float32(0))}
		case Body:
			base = PartStyle{PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(16))}
		case Footer:
			e.Justify(ui.Start).BorderWidth(1, 0, 0, 0).BorderColor(t.Border)
			base = PartStyle{PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(6))}
		}
		finish(t, ComponentSidebar, slot, e, base, o.Decoration)
	}
	build := func(t Theme) {
		if content != nil && scope.Mobile {
			content(c, t, false)
		}
	}
	// These components own their modal input scope, Escape/outside dismissal,
	// scrolling and exit animation. Only their model is mirrored back here.
	switch o.Mode {
	case SidebarModal:
		ModalWithOptions(c, t, &open, mo, build)
	case SidebarDrawer:
		m.drawer.Open = open
		Drawer(c, t, &m.drawer, DrawerOptions{ModalOptions: mo, SnapPoints: []float32{.85}, pinnedRegions: true}, build)
	default:
		edge := SideLeft
		if right {
			edge = SideRight
		}
		Slideover(c, t, &open, SlideoverOptions{ModalOptions: mo, Side: edge, Extent: Set(width)}, build)
	}
}
