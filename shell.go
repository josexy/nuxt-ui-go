package nui

import (
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
}
type NavigationOptions struct {
	Label             string
	Vertical, Compact bool
	OnNavigate        func(string)
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
	}
	for id := range s.menus {
		if !seen[id] {
			delete(s.menus, id)
			delete(s.expanded, id)
		}
	}
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
		root.Column().AlignItems(ui.Stretch).FocusGroup(ui.Vertical)
	}
	finish(t, ComponentNavigationMenu, Root, root, PartStyle{Gap: Set(t.spacing(4))}, o.Decoration)
	var render func([]NavigationItem, int)
	render = func(items []NavigationItem, depth int) {
		group := ""
		for _, it := range items {
			if it.Group != "" && it.Group != group && !o.Compact {
				finish(t, ComponentNavigationMenu, Label, ui.Text(c, it.Group).Padding(8, 10).FontSize(t.text(12)).FontWeight(600), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
			}
			group = it.Group
			if s.menus[it.ID] == nil {
				s.menus[it.ID] = &MenuState{}
			}
			ui.Column(c.Key(it.ID)).MinWidth(0).Children(func() {
				variant := Ghost
				if *active == it.ID {
					variant = Soft
				}
				label := it.Label
				if o.Compact && it.Icon != nil {
					label = ""
				}
				b := Button(c, t, label, ButtonOptions{AccessibleLabel: it.Label, Leading: it.Icon, Variant: variant, Color: Neutral, Disabled: it.Disabled}).Description(it.Description)
				if o.Vertical {
					b.FillWidth().Justify(ui.Start)
				}
				finish(t, ComponentNavigationMenu, Item, b, PartStyle{}, o.Decoration)
				if len(it.Children) == 0 {
					b.OnClick(func() { choose(it.ID) })
					return
				}
				if o.Vertical {
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
	return root
}

type SidebarOptions struct {
	Title, ToggleLabel    string
	Width, CollapsedWidth Optional[float32]
	Header, Footer        func(*ui.Context, Theme)
	Decoration
}

func Sidebar(c *ui.Context, t Theme, collapsed *bool, o SidebarOptions, content func(*ui.Context, Theme, bool)) ui.Element {
	if collapsed == nil {
		panic("nui: Sidebar requires collapsed state")
	}
	w := float32(224)
	if o.Width.Set {
		w = o.Width.Value
	}
	if *collapsed {
		w = 64
		if o.CollapsedWidth.Set {
			w = o.CollapsedWidth.Value
		}
	}
	root := ui.Column(c).Width(max(0, w)).AlignSelf(ui.Stretch).MinWidth(0).MinHeight(0).Shrink(0).Role(ui.RoleGroup).Label(o.Title).BorderEndWidth(1).BorderColor(t.Border)
	finish(t, ComponentSidebar, Root, root, PartStyle{Background: Set(t.Background), PaddingX: Set(t.spacing(8)), PaddingY: Set(t.spacing(12)), Gap: Set(t.spacing(12))}, o.Decoration)
	root.Children(func() {
		ui.Row(c).MinWidth(0).AlignItems(ui.Center).Children(func() {
			if !*collapsed {
				ui.Text(c, o.Title).Grow(1).MinWidth(0).SingleLine().FontSize(t.text(16)).LineHeight(24.0 / 16).FontWeight(600).TextColor(t.Highlighted)
			}
			label := o.ToggleLabel
			if label == "" {
				label = message(c, "toggleSidebar")
			}
			// Panel icons point toward the sidebar's edge, which is the right edge right to left.
			icon := lucide.PanelLeftClose()
			switch rtl := c.Direction() == ui.RightToLeft; {
			case *collapsed && rtl:
				icon = lucide.PanelRightOpen()
			case *collapsed:
				icon = lucide.PanelLeftOpen()
			case rtl:
				icon = lucide.PanelRightClose()
			}
			Button(c, t, "", ButtonOptions{Leading: icon, AccessibleLabel: label, Square: true, Variant: Ghost, Color: Neutral}).Expanded(!*collapsed).OnClick(func() { *collapsed = !*collapsed })
		})
		if o.Header != nil {
			o.Header(c, t)
		}
		body := ui.Scroll(c).Grow(1).MinWidth(0).MinHeight(0)
		finish(t, ComponentSidebar, Body, body, PartStyle{}, o.Decoration)
		if content != nil {
			body.Children(func() { content(c, t, *collapsed) })
		}
		if o.Footer != nil {
			o.Footer(c, t)
		}
	})
	return root
}

type HeaderOptions struct {
	Title, Description string
	Leading, Actions   func(*ui.Context, Theme)
	Decoration
}

// HeaderLayout is the Header component; Header remains the public slot constant.
func HeaderLayout(c *ui.Context, t Theme, o HeaderOptions) ui.Element {
	root := ui.Row(c).Wrap().MinWidth(0).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
	finish(t, ComponentHeader, Root, root, PartStyle{PaddingX: Set(t.spacing(20)), PaddingY: Set(t.spacing(16)), Gap: Set(t.spacing(12)), Background: Set(t.Background)}, o.Decoration)
	root.Children(func() {
		if o.Leading != nil {
			o.Leading(c, t)
		}
		ui.Column(c).MinWidth(0).Grow(1).Children(func() {
			finish(t, ComponentHeader, Label, ui.Text(c, o.Title).Role(ui.RoleHeading).Level(1), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(18)), Weight: Set(600)}, o.Decoration)
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
	root := ui.Row(c).Wrap().MinWidth(0).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(t.Border)
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
			ui.Column(c.Key(col.ID)).Width(w).Grow(1).Gap(8).Children(func() {
				finish(t, ComponentFooterColumns, Label, ui.Text(c, col.Title).FontWeight(600), PartStyle{Foreground: Set(t.Highlighted)}, o.Decoration)
				ids := map[string]bool{}
				for _, it := range col.Items {
					checkID(ids, it.ID, "FooterColumns item")
					b := Button(c.Key(it.ID), t, it.Label, ButtonOptions{Variant: LinkVariant, Color: Neutral, Leading: it.Icon, Disabled: it.Disabled}).Justify(ui.Start)
					finish(t, ComponentFooterColumns, Item, b, PartStyle{}, o.Decoration)
					b.OnClick(func() {
						if o.OnNavigate != nil {
							o.OnNavigate(it.ID)
						}
					})
				}
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
