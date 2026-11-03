package nui

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type NavigationItem struct {
	ID, Label, Description, Group string
	Icon                          *ui.SVG
	Disabled                      bool
	Children                      []NavigationItem
}
type NavigationState struct {
	menus    map[string]*MenuState
	expanded map[string]bool
	popovers map[string]*navigationPopover
	handles  map[string]*ui.Handle
	compact  bool
}
type navigationPopover struct {
	open, pinned, focus   bool
	suppressed, restoring bool
	state                 NavigationState
	panel                 ui.Handle
}
type NavigationOptions struct {
	Label             string
	Vertical, Compact bool
	// Tooltip and Popover default to true for compact vertical navigation.
	Tooltip, Popover Optional[bool]
	// PopupPlacement defaults to the reading direction's trailing side.
	PopupPlacement Optional[Placement]
	OnNavigate     func(string)
	Decoration
}

// NavigationMenu leaves routing to OnNavigate. IDs are unique across the whole tree.
func NavigationMenu(c *ui.Context, t Theme, active *string, s *NavigationState, items []NavigationItem, o NavigationOptions) ui.Element {
	if active == nil || s == nil {
		panic("nui: NavigationMenu requires active and state")
	}
	seen := map[string]bool{}
	var visit func([]NavigationItem, int)
	visit = func(items []NavigationItem, depth int) {
		if depth > 32 {
			panic("nui: NavigationMenu nesting exceeds 32")
		}
		for _, it := range items {
			checkID(seen, it.ID, "NavigationMenu")
			visit(it.Children, depth+1)
		}
	}
	visit(items, 0)
	if s.menus == nil {
		s.menus = map[string]*MenuState{}
		s.expanded = map[string]bool{}
		s.popovers = map[string]*navigationPopover{}
		s.handles = map[string]*ui.Handle{}
	}
	for id := range s.menus {
		if !seen[id] {
			delete(s.menus, id)
			delete(s.expanded, id)
			delete(s.popovers, id)
			delete(s.handles, id)
		}
	}
	compact := o.Vertical && o.Compact
	if compact != s.compact {
		for id, popup := range s.popovers {
			if popup.open && popup.panel.FocusWithin(c) {
				s.handles[id].Focus(c)
			}
			popup.open = false
			popup.pinned = false
			popup.suppressed, popup.restoring = false, false
		}
	}
	var containsActive func(NavigationItem) bool
	containsActive = func(it NavigationItem) bool {
		if it.ID == *active {
			return true
		}
		for _, child := range it.Children {
			if containsActive(child) {
				return true
			}
		}
		return false
	}
	var focusedChild func([]NavigationItem) bool
	focusedChild = func(items []NavigationItem) bool {
		for _, it := range items {
			if handle := s.handles[it.ID]; handle != nil && handle.Focused(c) {
				return true
			}
			if focusedChild(it.Children) {
				return true
			}
		}
		return false
	}
	placement := RightStart
	if c.Direction() == ui.RightToLeft {
		placement = LeftStart
	}
	placement = valueOr(o.PopupPlacement, placement)
	choose := func(id string) {
		*active = id
		if o.OnNavigate != nil {
			o.OnNavigate(id)
		}
	}
	var menuItems func([]NavigationItem) []MenuItem
	menuItems = func(items []NavigationItem) []MenuItem {
		out := make([]MenuItem, 0, len(items))
		for _, it := range items {
			out = append(out, MenuItem{ID: it.ID, Label: it.Label, Description: it.Description, Icon: it.Icon, Disabled: it.Disabled, Checked: Set(*active == it.ID), Children: menuItems(it.Children), OnSelect: func() { choose(it.ID) }})
		}
		return out
	}
	root := ui.Row(c).Wrap().MinWidth(0).Role(ui.RoleGroup).Label(o.Label)
	if o.Vertical {
		root.Column().FillWidth().AlignItems(ui.Stretch).FocusGroup(ui.Vertical)
	}
	finish(t, ComponentNavigationMenu, Root, root, PartStyle{Gap: Set(t.spacing(4))}, o.Decoration)
	var render func([]NavigationItem, int)
	render = func(items []NavigationItem, depth int) {
		group := ""
		for _, it := range items {
			if it.Group != "" && it.Group != group && !o.Compact {
				finish(t, ComponentNavigationMenu, Label, ui.Text(c.Key(struct{ Group string }{it.ID}), it.Group).Padding(8, 10).FontSize(t.text(12)).FontWeight(600), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
			}
			group = it.Group
			if s.menus[it.ID] == nil {
				s.menus[it.ID] = &MenuState{}
				s.popovers[it.ID] = &navigationPopover{}
				s.handles[it.ID] = &ui.Handle{}
			}
			if compact && !s.compact && focusedChild(it.Children) {
				s.handles[it.ID].Focus(c)
			}
			item := ui.Column(c.Key(it.ID)).MinWidth(0)
			if o.Vertical {
				item.FillWidth()
			}
			item.Children(func() {
				variant := Ghost
				if containsActive(it) {
					variant = Soft
				}
				label := it.Label
				if o.Compact && it.Icon != nil {
					label = ""
				}
				var trailing *ui.SVG
				if o.Vertical && !compact && len(it.Children) > 0 {
					trailing = lucide.ChevronRight()
					if c.Direction() == ui.RightToLeft {
						trailing = lucide.ChevronLeft()
					}
					if s.expanded[it.ID] {
						trailing = lucide.ChevronDown()
					}
				}
				b := Button(c.Key("trigger"), t, label, ButtonOptions{AccessibleLabel: it.Label, Leading: it.Icon, Trailing: trailing, Block: o.Vertical, Variant: variant, Color: Neutral, Disabled: it.Disabled}).Description(it.Description).Bind(s.handles[it.ID])
				if o.Vertical {
					b.FillWidth().Justify(ui.Start)
					if compact {
						b.Justify(ui.Center)
					}
				}
				finish(t, ComponentNavigationMenu, Item, b, PartStyle{}, o.Decoration)
				if compact && valueOr(o.Tooltip, true) && (len(it.Children) == 0 || !valueOr(o.Popover, true)) {
					Tooltip(c.Key("tooltip"), t, b, it.Label, TooltipOptions{Placement: Set(placement)})
				}
				if len(it.Children) == 0 {
					b.OnClick(func() { choose(it.ID) })
					return
				}
				if compact && valueOr(o.Popover, true) {
					popup := s.popovers[it.ID]
					if popup.restoring && s.handles[it.ID].Focused(c) {
						popup.restoring = false
					}
					if !popup.restoring && !b.Hovered() && !b.FocusVisible() {
						popup.suppressed = false
					}
					open := func() {
						popup.open, popup.pinned, popup.focus = true, true, true
						popup.suppressed, popup.restoring = false, false
					}
					close := func() {
						popup.open, popup.pinned, popup.focus, popup.suppressed = false, false, false, true
						popup.restoring = true
						s.handles[it.ID].Focus()
					}
					b.Expanded(popup.open).OnClick(func() {
						if popup.open && popup.pinned {
							close()
						} else {
							open()
						}
					})
					forward := ui.KeyRight
					if placement == LeftStart || placement == LeftCenter || placement == LeftEnd {
						forward = ui.KeyLeft
					}
					b.OnShortcut(0, forward, open)
					mode := PopoverHover
					if popup.pinned || popup.suppressed {
						mode = PopoverClick
					}
					panel := Popover(c.Key("popover"), t, b, &popup.open, PopoverOptions{Label: it.Label, Placement: placement, Mode: mode, CloseDelay: 150 * time.Millisecond, Width: Set(float32(224))}, func(t Theme) {
						ui.Text(c.Key("label"), it.Label).FontWeight(600).TextColor(t.Muted)
						NavigationMenu(c.Key("children"), t, active, &popup.state, it.Children, NavigationOptions{Label: it.Label, Vertical: true, OnNavigate: func(id string) {
							choose(id)
							close()
						}, Decoration: o.Decoration})
						if popup.focus {
							for _, child := range it.Children {
								if !child.Disabled {
									popup.state.handles[child.ID].Focus(c)
									break
								}
							}
							popup.focus = false
						}
					})
					panel.Bind(&popup.panel)
					back := ui.KeyLeft
					if forward == ui.KeyLeft {
						back = ui.KeyRight
					}
					panel.OnShortcut(0, back, close)
					if panel.Dismissed() && popup.panel.FocusWithin(c) {
						s.handles[it.ID].Focus(c)
					}
					if !popup.open {
						popup.pinned, popup.focus = false, false
					}
				} else if o.Vertical {
					b.Expanded(s.expanded[it.ID]).OnClick(func() { s.expanded[it.ID] = !s.expanded[it.ID] })
					if s.expanded[it.ID] {
						ui.Column(c).Padding(0, 0, 0, 12).Gap(4).Disabled(it.Disabled).Children(func() { render(it.Children, depth+1) })
					}
				} else {
					DropdownMenu(c, t, b, s.menus[it.ID], menuItems(it.Children), MenuOptions{Label: it.Label})
				}
			})
		}
	}
	if !o.Vertical {
		root.FocusGroup(ui.Horizontal)
	}
	root.Children(func() { render(items, 0) })
	s.compact = compact
	return root
}

type HeaderOptions struct {
	Title, Description string
	Leading, Actions   func(*ui.Context, Theme)
	Decoration
}

// HeaderLayout is the Header component; Header remains the public slot constant.
func HeaderLayout(c *ui.Context, t Theme, o HeaderOptions) ui.Element {
	root := ui.Row(c).Wrap().MinWidth(0).MinHeight(t.spacing(64)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
	finish(t, ComponentHeader, Root, root, PartStyle{PaddingX: Set(t.spacing(24)), PaddingY: Set(t.spacing(8)), Gap: Set(t.spacing(12)), Background: Set(t.Background)}, o.Decoration)
	root.Children(func() {
		if o.Leading != nil {
			o.Leading(c, t)
		}
		ui.Column(c).MinWidth(0).Grow(1).Children(func() {
			finish(t, ComponentHeader, Label, ui.Text(c, o.Title).Role(ui.RoleHeading).Level(1), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(20)), LineHeight: Set(float32(28.0 / 20)), Weight: Set(700)}, o.Decoration)
			if o.Description != "" {
				finish(t, ComponentHeader, Description, ui.Text(c, o.Description), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(12))}, o.Decoration)
			}
		})
		if o.Actions != nil {
			a := ui.Row(c).Wrap().Gap(8).MinWidth(0)
			finish(t, ComponentHeader, Actions, a, PartStyle{}, o.Decoration)
			a.Children(func() { o.Actions(c, t) })
		}
	})
	return root
}

type FooterOptions struct {
	Text    string
	Actions func(*ui.Context, Theme)
	Decoration
}

func FooterLayout(c *ui.Context, t Theme, o FooterOptions) ui.Element {
	root := ui.Row(c).Wrap().MinWidth(0).AlignItems(ui.Center)
	finish(t, ComponentFooter, Root, root, PartStyle{PaddingX: Set(t.spacing(20)), PaddingY: Set(t.spacing(12)), Gap: Set(t.spacing(12)), Foreground: Set(t.Muted)}, o.Decoration)
	root.Children(func() {
		finish(t, ComponentFooter, Label, ui.Text(c, o.Text).Grow(1).MinWidth(0), PartStyle{}, o.Decoration)
		if o.Actions != nil {
			o.Actions(c, t)
		}
	})
	return root
}

type FooterColumn struct {
	ID, Title string
	Items     []NavigationItem
}
type FooterColumnsOptions struct {
	MinColumnWidth float32
	OnNavigate     func(string)
	Decoration
}

func FooterColumns(c *ui.Context, t Theme, columns []FooterColumn, o FooterColumnsOptions) ui.Element {
	w := o.MinColumnWidth
	if w <= 0 {
		w = 160
	}
	seen := map[string]bool{}
	root := ui.Row(c).Wrap().MinWidth(0)
	finish(t, ComponentFooterColumns, Root, root, PartStyle{Gap: Set(t.spacing(24))}, o.Decoration)
	root.Children(func() {
		for _, col := range columns {
			checkID(seen, col.ID, "FooterColumns")
			ui.Column(c.Key(col.ID)).Width(w).Grow(1).Gap(t.spacing(24)).Children(func() {
				finish(t, ComponentFooterColumns, Label, ui.Text(c, col.Title).FontWeight(600), PartStyle{Foreground: Set(t.Highlighted)}, o.Decoration)
				ids := map[string]bool{}
				list := ui.Column(c.Key("links")).MinWidth(0).Gap(t.spacing(16)).AlignItems(ui.Start)
				list.Children(func() {
					for _, it := range col.Items {
						checkID(ids, it.ID, "FooterColumns item")
						b := Button(c.Key(it.ID), t, it.Label, ButtonOptions{Variant: LinkVariant, Color: Neutral, Leading: it.Icon, Disabled: it.Disabled, UI: Styles{Root: {PaddingX: Set(float32(0)), PaddingY: Set(float32(0))}}}).Justify(ui.Start)
						finish(t, ComponentFooterColumns, Item, b, PartStyle{}, o.Decoration)
						b.OnClick(func() {
							if o.OnNavigate != nil {
								o.OnNavigate(it.ID)
							}
						})
					}
				})
			})
		}
	})
	return root
}

type SplitterOptions struct {
	Vertical, Percent, Disabled bool
	Min, Max                    Optional[float32]
	OnChange                    func(float32)
	Decoration
}

// Splitter binds DIPs or a percentage (0..100) to value. Native panes retain their 40-DIP minimum.
func Splitter(c *ui.Context, t Theme, value *float32, o SplitterOptions, first, second func()) ui.Element {
	if value == nil {
		panic("nui: Splitter requires value")
	}
	clamp := func(v float32) float32 {
		lo := float32(0)
		hi := float32(1e7)
		if o.Percent {
			hi = 100
		}
		if o.Min.Set {
			lo = max(lo, o.Min.Value)
		}
		if o.Max.Set {
			hi = min(hi, o.Max.Value)
		}
		if hi < lo {
			panic("nui: invalid Splitter constraints")
		}
		return max(lo, min(hi, v))
	}
	*value = clamp(*value)
	root := ui.Themed(c, t.Native()).Background(ui.Transparent).AlignSelf(ui.Stretch).MinWidth(0).MinHeight(0).Disabled(o.Disabled)
	finish(t, ComponentSplitter, Root, root, PartStyle{}, o.Decoration)
	root.Children(func() {
		pane := func(slot Slot, fn func()) func() {
			return func() {
				e := ui.Column(c).Fill().MinWidth(0).MinHeight(0)
				finish(t, ComponentSplitter, slot, e, PartStyle{}, o.Decoration)
				if fn != nil {
					e.Children(fn)
				}
			}
		}
		var e ui.Element
		switch {
		case o.Vertical && o.Percent:
			e = ui.SplitVerticalPercent(c, value, pane(Leading, first), pane(Trailing, second))
		case o.Vertical:
			e = ui.SplitVertical(c, value, pane(Leading, first), pane(Trailing, second))
		case o.Percent:
			e = ui.SplitPercent(c, value, pane(Leading, first), pane(Trailing, second))
		default:
			e = ui.Split(c, value, pane(Leading, first), pane(Trailing, second))
		}
		e.Fill().OnChange(func() {
			*value = clamp(*value)
			if o.OnChange != nil {
				o.OnChange(*value)
			}
		})
	})
	return root
}
