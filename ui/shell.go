package nui

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

type NavigationItem struct {
	ID, Label, Description, Group  string
	Icon                           *ui.SVG
	Avatar                         *AvatarOptions
	Badge                          string
	BadgeOptions                   *BadgeOptions
	Chip                           *ChipOptions
	TrailingIcon                   *ui.SVG
	Active, DefaultOpen, LabelOnly bool
	Tooltip, Popover               Optional[bool]
	TooltipOptions                 *TooltipOptions
	PopoverOptions                 *PopoverOptions
	Disabled                       bool
	Children                       []NavigationItem
	OnSelect                       func(*NavigationEvent)
	Decoration
}

type NavigationEvent struct {
	ID        string
	prevented bool
}

func (e *NavigationEvent) PreventDefault() { e.prevented = true }

type NavigationState struct {
	menus    map[string]*MenuState
	Expanded map[string]bool
	popovers map[string]*navigationPopover
	handles  map[string]*ui.Handle
	compact  bool
	vertical bool
	known    map[string]bool
	parents  map[string]string
}
type navigationPopover struct {
	open, pinned, focus   bool
	suppressed, restoring bool
	state                 NavigationState
	panel                 ui.Handle
}
type NavigationOptions struct {
	Label                 string
	Vertical, Compact     bool
	Color, HighlightColor Color
	// Soft is the pill appearance; LinkVariant omits its active background.
	Variant                     Variant
	Highlight, Disabled, Single bool
	Collapsible                 Optional[bool]
	TrailingIcon                *ui.SVG
	// Tooltip and Popover default to true for compact vertical navigation.
	Tooltip, Popover Optional[bool]
	TooltipOptions   *TooltipOptions
	PopoverOptions   *PopoverOptions
	// PopupPlacement defaults to the reading direction's trailing side.
	PopupPlacement   Optional[Placement]
	OnNavigate       func(string)
	OnSelect         func(*NavigationEvent)
	OnExpandedChange func([]string)
	Decoration
}

// NavigationMenu leaves routing to OnNavigate. IDs are unique across the whole tree.
func NavigationMenu(c *ui.Context, t Theme, active *string, s *NavigationState, items []NavigationItem, o NavigationOptions) ui.Element {
	return navigationMenu(c, t, active, s, items, o, false)
}

func navigationMenu(c *ui.Context, t Theme, active *string, s *NavigationState, items []NavigationItem, o NavigationOptions, nested bool) ui.Element {
	if active == nil || s == nil {
		panic("nui: NavigationMenu requires active and state")
	}
	config := selected(defaults(t, ComponentNavigationMenu, ComponentTheme{Color: Primary, Variant: Soft}), ComponentTheme{Color: o.Color, Variant: o.Variant})
	if config.Variant != Soft && config.Variant != LinkVariant {
		panic("nui: NavigationMenu supports Soft and LinkVariant")
	}
	if s.known == nil {
		s.known = map[string]bool{}
		s.parents = map[string]string{}
	}
	if s.Expanded == nil {
		s.Expanded = map[string]bool{}
	}
	seen := map[string]bool{}
	nodes := map[string]NavigationItem{}
	parents := map[string]string{}
	var visit func([]NavigationItem, int)
	visit = func(items []NavigationItem, depth int) {
		if depth > 32 {
			panic("nui: NavigationMenu nesting exceeds 32")
		}
		opened := false
		for _, it := range items {
			checkID(seen, it.ID, "NavigationMenu")
			nodes[it.ID] = it
			if _, exists := s.Expanded[it.ID]; !exists && !s.known[it.ID] && it.DefaultOpen && len(it.Children) > 0 && !it.LabelOnly {
				s.Expanded[it.ID] = true
			}
			if o.Vertical && o.Single && len(it.Children) > 0 && !it.LabelOnly && s.Expanded[it.ID] {
				if opened {
					s.Expanded[it.ID] = false
				} else {
					opened = true
				}
			}
			s.known[it.ID] = true
			for _, child := range it.Children {
				parents[child.ID] = it.ID
			}
			visit(it.Children, depth+1)
		}
	}
	visit(items, 0)
	if !nested {
		for id := range s.Expanded {
			node, ok := nodes[id]
			if !ok || node.LabelOnly || len(node.Children) == 0 {
				delete(s.Expanded, id)
			}
		}
	}
	if s.menus == nil {
		s.menus = map[string]*MenuState{}
		s.popovers = map[string]*navigationPopover{}
		s.handles = map[string]*ui.Handle{}
	}
	restore := ""
	lostFocus := false
	for id := range s.known {
		if !seen[id] {
			if handle := s.handles[id]; handle != nil {
				focused := handle.Focused(c)
				if menu := s.menus[id]; menu != nil {
					focused = focused || menu.menuFocused(c)
				}
				if popup := s.popovers[id]; popup != nil {
					focused = focused || popup.panel.FocusWithin(c)
				}
				if focused {
					lostFocus = true
					for parent := s.parents[id]; parent != ""; parent = s.parents[parent] {
						if seen[parent] {
							restore = parent
							break
						}
					}
				}
				handle.CancelFocus(c)
			}
			delete(s.menus, id)
			delete(s.Expanded, id)
			delete(s.popovers, id)
			delete(s.handles, id)
			delete(s.known, id)
		}
	}
	s.parents = parents
	for id := range seen {
		if s.handles[id] == nil {
			s.handles[id] = &ui.Handle{}
		}
	}
	if lostFocus && restore == "" {
		for _, it := range items {
			if !it.Disabled && !it.LabelOnly {
				restore = it.ID
				break
			}
		}
	}
	if restore != "" {
		s.handles[restore].Focus(c)
	}
	compact := o.Vertical && o.Compact
	if compact != s.compact || o.Vertical != s.vertical {
		for _, menu := range s.menus {
			if menu.Open {
				menu.close(menu.menuFocused(c), nil)
			}
		}
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
		if it.ID == *active || it.Active {
			return true
		}
		return slices.ContainsFunc(it.Children, containsActive)
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
	choose := func(it NavigationItem) bool {
		if it.Disabled || it.LabelOnly {
			return false
		}
		previous := *active
		event := NavigationEvent{ID: it.ID}
		if it.OnSelect != nil {
			it.OnSelect(&event)
		} else if o.OnSelect != nil {
			o.OnSelect(&event)
		}
		if event.prevented || *active != previous {
			return false
		}
		*active = it.ID
		if o.OnNavigate != nil {
			o.OnNavigate(it.ID)
		}
		return true
	}
	var menuItems func([]NavigationItem, *MenuState) []MenuItem
	menuItems = func(items []NavigationItem, menu *MenuState) []MenuItem {
		out := make([]MenuItem, 0, len(items))
		for _, it := range items {
			if it.LabelOnly {
				continue
			}
			kind := MenuAction
			out = append(out, MenuItem{ID: it.ID, Label: it.Label, Description: it.Description, Kind: kind, Icon: it.Icon, Avatar: it.Avatar, Disabled: it.Disabled, Checked: Set(*active == it.ID || it.Active), Close: Set(false), Children: menuItems(it.Children, menu), OnSelect: func() {
				if choose(it) {
					menu.close(true, nil)
				}
			}})
		}
		return out
	}
	name := o.Label
	if name == "" {
		name = t.field.Label
	}
	root := ui.Row(c).Wrap().MinWidth(0).Role(ui.RoleGroup).Label(name).Disabled(o.Disabled || t.field.Disabled)
	if o.Vertical {
		root.Column().FillWidth().AlignItems(ui.Stretch).FocusGroup(ui.Vertical)
	}
	rootSlot := Root
	if nested {
		rootSlot = List
	}
	finish(t, ComponentNavigationMenu, rootSlot, root, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
	expandedChanged := func() {
		if o.OnExpandedChange != nil {
			ids := []string{}
			for id, open := range s.Expanded {
				if open {
					ids = append(ids, id)
				}
			}
			slices.Sort(ids)
			o.OnExpandedChange(ids)
		}
	}
	horizontalIDs := []string{}
	if !o.Vertical {
		for _, item := range items {
			if !item.Disabled && !item.LabelOnly {
				horizontalIDs = append(horizontalIDs, item.ID)
			}
		}
	}
	var render func([]NavigationItem, int)
	render = func(items []NavigationItem, depth int) {
		group := ""
		for _, it := range items {
			if it.Group != "" && it.Group != group && o.Vertical && !compact {
				finishNavigationPart(t, Label, ui.Text(c.Key(struct{ Group string }{it.ID}), it.Group), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(12)), Weight: Set(600), PaddingX: Set(t.spacing(10)), PaddingY: Set(t.spacing(6))}, o.Decoration, it.Decoration)
			}
			group = it.Group
			if s.menus[it.ID] == nil {
				s.menus[it.ID] = &MenuState{}
				s.popovers[it.ID] = &navigationPopover{}
			}
			if compact && !s.compact && focusedChild(it.Children) {
				s.handles[it.ID].Focus(c)
			}
			item := ui.Column(c.Key(it.ID)).MinWidth(0)
			finishNavigationPart(t, Body, item, PartStyle{}, o.Decoration, it.Decoration)
			if o.Vertical {
				item.FillWidth()
			}
			item.Children(func() {
				if it.LabelOnly {
					if o.Vertical && !compact {
						finishNavigationPart(t, Label, ui.Text(c, it.Label), PartStyle{FontSize: Set(t.text(12)), Weight: Set(600), Foreground: Set(t.Highlighted), PaddingX: Set(t.spacing(10)), PaddingY: Set(t.spacing(6))}, o.Decoration, it.Decoration)
					}
					return
				}
				label := it.Label
				if compact && (it.Icon != nil || it.Avatar != nil) {
					label = ""
				}
				open := s.Expanded[it.ID]
				if compact {
					open = s.popovers[it.ID].open
				} else if !o.Vertical {
					open = s.menus[it.ID].Open
				}
				b := navigationButton(c.Key("trigger"), t, it, o, config, label, containsActive(it), open, compact, depth).Bind(s.handles[it.ID])
				if !o.Vertical && !b.IsDisabled() {
					move := func(delta int, edge bool) {
						at := slices.Index(horizontalIDs, it.ID)
						if at < 0 {
							return
						}
						next := delta
						if !edge {
							next = at + delta
						}
						next = max(0, min(len(horizontalIDs)-1, next))
						s.handles[horizontalIDs[next]].Focus()
					}
					back, forward := ui.KeyLeft, ui.KeyRight
					if c.Direction() == ui.RightToLeft {
						back, forward = forward, back
					}
					b.OnShortcut(0, back, func() { move(-1, false) }).OnShortcut(0, forward, func() { move(1, false) }).OnShortcut(0, ui.KeyHome, func() { move(0, true) }).OnShortcut(0, ui.KeyEnd, func() { move(len(horizontalIDs)-1, true) })
				}
				if o.Vertical {
					b.FillWidth().Justify(ui.Start)
					if compact {
						b.Justify(ui.Center)
					}
				}
				popover := valueOr(it.Popover, valueOr(o.Popover, true))
				tooltip := valueOr(it.Tooltip, compact && valueOr(o.Tooltip, true))
				if tooltip && (compact || !o.Vertical) && (len(it.Children) == 0 || !compact || !popover) {
					to := TooltipOptions{Placement: Set(placement), DelayDuration: Set(time.Duration(0))}
					if !o.Vertical {
						to = TooltipOptions{}
					}
					if o.TooltipOptions != nil {
						to = *o.TooltipOptions
					}
					if it.TooltipOptions != nil {
						to = *it.TooltipOptions
					}
					if compact && !to.Placement.Set {
						to.Placement = Set(placement)
					}
					if compact && !to.DelayDuration.Set {
						to.DelayDuration = Set(time.Duration(0))
					}
					Tooltip(c.Key("tooltip"), t, b, it.Label, to)
				}
				if len(it.Children) == 0 {
					b.OnClick(func() { choose(it) })
					return
				}
				if compact && popover {
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
					po := PopoverOptions{Label: it.Label, Placement: placement, Mode: PopoverHover, CloseDelay: 150 * time.Millisecond, Width: Set(float32(224))}
					if o.PopoverOptions != nil {
						po = *o.PopoverOptions
					}
					if it.PopoverOptions != nil {
						po = *it.PopoverOptions
					}
					if po.Label == "" {
						po.Label = it.Label
					}
					if o.PopupPlacement.Set {
						po.Placement = o.PopupPlacement.Value
					}
					forward := ui.KeyRight
					if po.Placement == LeftStart || po.Placement == LeftCenter || po.Placement == LeftEnd {
						forward = ui.KeyLeft
					}
					b.OnShortcut(0, forward, open)
					if popup.pinned || popup.suppressed {
						po.Mode = PopoverClick
					}
					panel := Popover(c.Key("popover"), t, b, &popup.open, po, func(t Theme) {
						ui.Text(c.Key("label"), it.Label).FontWeight(600).TextColor(t.Muted)
						childOptions := o
						childOptions.Label = it.Label
						childOptions.Vertical = true
						childOptions.Compact = false
						childOptions.OnNavigate = func(id string) {
							if o.OnNavigate != nil {
								o.OnNavigate(id)
							}
							close()
						}
						popup.state.Expanded = s.Expanded
						navigationMenu(c.Key("children"), t, active, &popup.state, it.Children, childOptions, true)
						if popup.focus {
							for _, child := range it.Children {
								if !child.Disabled && !child.LabelOnly {
									popup.state.handles[child.ID].Focus(c)
									break
								}
							}
							popup.focus = false
						}
					})
					if panel.Valid() {
						finishNavigationPart(t, Panel, panel, PartStyle{}, o.Decoration, it.Decoration)
					}
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
					b.Expanded(s.Expanded[it.ID]).OnClick(func() {
						if o.Single && s.Expanded[it.ID] && !valueOr(o.Collapsible, true) {
							return
						}
						next := !s.Expanded[it.ID]
						if next && o.Single {
							for _, sibling := range items {
								s.Expanded[sibling.ID] = false
							}
						}
						s.Expanded[it.ID] = next
						expandedChanged()
					})
					if s.Expanded[it.ID] {
						children := ui.Column(c.Key("children")).MarginStart(t.spacing(20)).PaddingStart(t.spacing(6)).BorderStartWidth(1).BorderColor(t.Border).Gap(t.spacing(4)).Disabled(it.Disabled)
						finishNavigationPart(t, List, children, PartStyle{}, o.Decoration, it.Decoration)
						children.Children(func() { render(it.Children, depth+1) })
					} else if focusedChild(it.Children) {
						s.handles[it.ID].Focus(c)
					}
				} else {
					panel := menu(c, t, b, s.menus[it.ID], menuItems(it.Children, s.menus[it.ID]), MenuOptions{Label: it.Label, Disabled: root.IsDisabled()}, false, func(item MenuItem, base Decoration) Decoration {
						return navigationMenuDecoration(t, o, nodes[item.ID], base)
					})
					if panel.Valid() {
						finishNavigationPart(t, Panel, panel, PartStyle{}, o.Decoration, it.Decoration)
					}
				}
			})
		}
	}
	if !o.Vertical {
		root.FocusGroup(ui.Horizontal)
	}
	root.Children(func() { render(items, 0) })
	s.compact = compact
	s.vertical = o.Vertical
	return root
}

type HeaderOptions struct {
	Title, Description string
	Leading, Actions   func(*ui.Context, Theme)
	Decoration
}

// HeaderLayout is the Header component; Header remains the public slot constant.
func HeaderLayout(c *ui.Context, t Theme, o HeaderOptions) ui.Element {
	// The window's banner landmark, as Nuxt UI's <header>. A nested instance
	// opts out by replacing the role on the returned element.
	root := ui.Row(c).Wrap().MinWidth(0).MinHeight(t.spacing(64)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Role(ui.RoleBanner)
	finish(t, ComponentHeader, Root, root, PartStyle{PaddingX: Set(t.spacing(24)), PaddingY: Set(t.spacing(8)), Gap: Set(t.spacing(12)), Background: Set(t.Background)}, o.Decoration)
	root.Children(func() {
		if o.Leading != nil {
			leading := ui.Row(c.Key("leading")).MinWidth(0).Shrink(0).AlignItems(ui.Center)
			finish(t, ComponentHeader, Leading, leading, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
			leading.Children(func() { o.Leading(c, t) })
		}
		heading := ui.Column(c.Key("heading")).MinWidth(0).Grow(1)
		finish(t, ComponentHeader, Body, heading, PartStyle{}, o.Decoration)
		heading.Children(func() {
			if o.Title != "" {
				finish(t, ComponentHeader, Label, ui.Text(c.Key("title"), o.Title).Role(ui.RoleHeading).Level(1), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(20)), LineHeight: Set(float32(28.0 / 20)), Weight: Set(700)}, o.Decoration)
			}
			if o.Description != "" {
				finish(t, ComponentHeader, Description, ui.Text(c.Key("description"), o.Description), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(12))}, o.Decoration)
			}
		})
		if o.Actions != nil {
			a := ui.Row(c.Key("actions")).Wrap().MinWidth(0)
			finish(t, ComponentHeader, Actions, a, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
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
	// The window's content-information landmark, as Nuxt UI's <footer>. A
	// nested instance opts out by replacing the role on the returned element.
	root := ui.Row(c).Wrap().MinWidth(0).AlignItems(ui.Center).Role(ui.RoleContentInfo)
	finish(t, ComponentFooter, Root, root, PartStyle{PaddingX: Set(t.spacing(20)), PaddingY: Set(t.spacing(12)), Gap: Set(t.spacing(12)), Foreground: Set(t.Muted)}, o.Decoration)
	root.Children(func() {
		finish(t, ComponentFooter, Label, ui.Text(c.Key("text"), o.Text).Grow(1).MinWidth(0), PartStyle{}, o.Decoration)
		if o.Actions != nil {
			actions := ui.Row(c.Key("actions")).Wrap().MinWidth(0).AlignItems(ui.Center)
			finish(t, ComponentFooter, Actions, actions, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
			actions.Children(func() { o.Actions(c, t) })
		}
	})
	return root
}
