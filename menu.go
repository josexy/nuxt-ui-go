package nui

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type MenuItemKind uint8

const (
	MenuAction MenuItemKind = iota
	MenuLabel
	MenuSeparator
)

type MenuItem struct {
	ID, Label, Description, Shortcut string
	Kind                             MenuItemKind
	Icon                             *ui.SVG
	Color                            Color
	Disabled, Radio                  bool
	Checked                          Optional[bool]
	Children                         []MenuItem
	OnSelect                         func()
	// Build supplies non-interactive visual content; Label is still required for semantics/typeahead.
	Build func(*ui.Context, Theme)
}
type MenuOptions struct {
	Label            string
	Width, MaxHeight Optional[float32]
	Decoration
}
type menuLevel struct {
	prefix  string
	typed   time.Time
	hover   string
	hoverAt time.Time
}

// MenuState is window-owned data. Never share one instance between independent menu triggers.
// Open can be set by the application; all other fields preserve keyboard/hover identity.
type MenuState struct {
	Open    bool
	shown   bool
	path    []string
	handles map[string]*ui.Handle
	levels  map[string]*menuLevel
	trigger ui.Handle
	focus   *ui.Handle
	x, y    float32
	point   bool
}

func (s *MenuState) resetHover() {
	for _, level := range s.levels {
		level.hover = ""
	}
}
func (s *MenuState) close(restore bool) {
	s.resetHover()
	s.focus = nil
	s.Open = false
	s.shown = false
	s.path = nil
	if restore {
		s.trigger.Focus()
	}
}
func (s *MenuState) prepare(c *ui.Context, items []MenuItem) {
	seen := map[string]bool{}
	var visit func([]MenuItem, int)
	visit = func(items []MenuItem, depth int) {
		if depth > 32 {
			panic("nui: menu nesting exceeds 32 levels")
		}
		for _, it := range items {
			checkID(seen, it.ID, "Menu")
			visit(it.Children, depth+1)
		}
	}
	visit(items, 0)
	if s.handles == nil {
		s.handles = map[string]*ui.Handle{}
		s.levels = map[string]*menuLevel{}
	}
	removedFocus := false
	for id, handle := range s.handles {
		if !seen[id] {
			removedFocus = removedFocus || handle.Focused(c)
			delete(s.handles, id)
			delete(s.levels, id)
		}
	}
	for id := range seen {
		if s.handles[id] == nil {
			s.handles[id] = &ui.Handle{}
		}
	}
	if removedFocus {
		s.focusEdge(items, false)
	}
	current := items
	for depth, id := range s.path {
		found := false
		for _, it := range current {
			if it.ID == id && !it.Disabled && len(it.Children) > 0 {
				current = it.Children
				found = true
				break
			}
		}
		if !found {
			s.path = s.path[:depth]
			break
		}
	}
}
func menuEnabled(it MenuItem) bool { return it.Kind == MenuAction && !it.Disabled }
func (s *MenuState) focusEdge(items []MenuItem, last bool) {
	for i := range items {
		j := i
		if last {
			j = len(items) - 1 - i
		}
		if menuEnabled(items[j]) {
			s.focus = s.handles[items[j].ID]
			return
		}
	}
}
func (s *MenuState) child(depth int, it MenuItem, focus bool) {
	if len(s.path) > depth {
		s.path = s.path[:depth]
	}
	if len(s.path) == depth {
		s.path = append(s.path, it.ID)
	}
	if focus {
		s.focusEdge(it.Children, false)
	}
}
func DropdownMenu(c *ui.Context, t Theme, anchor ui.Element, state *MenuState, items []MenuItem, o MenuOptions) ui.Element {
	return menu(c, t, anchor, state, items, o, false)
}
func ContextMenu(c *ui.Context, t Theme, anchor ui.Element, state *MenuState, items []MenuItem, o MenuOptions) ui.Element {
	return menu(c, t, anchor, state, items, o, true)
}
func menu(c *ui.Context, t Theme, anchor ui.Element, s *MenuState, items []MenuItem, o MenuOptions, context bool) ui.Element {
	// Consume the component key even when the popup is absent.
	ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	if s == nil {
		panic("nui: menu requires state")
	}
	s.prepare(c, items)
	if !anchor.Valid() || anchor.IsDisabled() {
		s.close(false)
		return ui.Element{}
	}
	anchor.Bind(&s.trigger)
	id := ComponentDropdownMenu
	if context {
		id = ComponentContextMenu
	}
	open := func(last bool) {
		s.resetHover()
		for _, level := range s.levels {
			level.prefix = ""
		}
		s.Open = true
		s.path = nil
		s.focusEdge(items, last)
	}
	if context {
		anchor.Focusable()
		height := anchor.Bounds().H
		anchor.OnShortcut(0, ui.KeyContextMenu, func() { s.point = true; s.x = 0; s.y = height; open(false) }).
			OnShortcut(ui.Shift, ui.KeyF10, func() { s.point = true; s.x = 0; s.y = height; open(false) })
		anchor.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 1 {
				s.point = true
				s.x = ev.X
				s.y = ev.Y
				open(false)
				return true
			}
			return false
		})
	} else {
		anchor.Role(ui.RoleMenuButton).Expanded(s.Open)
		anchor.OnClick(func() {
			if s.Open {
				s.close(true)
			} else {
				open(false)
			}
		})
		anchor.OnShortcut(0, ui.KeyDown, func() { open(false) }).OnShortcut(0, ui.KeyUp, func() { open(true) })
	}
	if !s.Open {
		s.shown = false
		s.resetHover()
		s.path = nil
		s.focus = nil
		return ui.Element{}
	}
	if !s.shown && s.focus == nil {
		for _, level := range s.levels {
			level.prefix = ""
		}
		s.focusEdge(items, false)
	}
	s.shown = true
	target := anchor
	if context && s.point {
		target = ui.Box(c.Key(&s.point)).Size(0, 0).AttachTo(anchor, ui.AnchorTopLeft, ui.AnchorTopLeft).Margin(s.y, 0, 0, s.x).IgnoreInput(true)
	}
	panel := ui.PopoverBase(c.Key(s), target, &s.Open, func(panel ui.Element) {
		menuPanel(c, t, panel, target, s, items, o, id, "", 0)
	})
	if !s.Open {
		s.path = nil
	}
	if s.Open && s.focus != nil {
		s.focus.Focus()
	}
	s.focus = nil
	return panel
}
func menuPanel(c *ui.Context, t Theme, panel, anchor ui.Element, s *MenuState, items []MenuItem, o MenuOptions, id Component, parent string, depth int) {
	if s.levels[parent] == nil {
		s.levels[parent] = &menuLevel{}
	}
	level := s.levels[parent]
	maxHeight := float32(320)
	if o.MaxHeight.Set {
		maxHeight = o.MaxHeight.Value
	}
	w, h := c.Size()
	maxHeight = min(maxHeight, h-24)
	width := float32(240)
	if o.Width.Set {
		width = o.Width.Value
	}
	width = min(width, w-24)
	popup := PopoverOptions{Width: Set(width), MaxHeight: Set(maxHeight)}
	if depth > 0 {
		popup.Placement = RightStart
		if c.Direction() == ui.RightToLeft {
			popup.Placement = LeftStart
		}
	}
	recipe := panelStyle(t)
	recipe.PaddingX = Set(t.spacing(4))
	recipe.PaddingY = Set(t.spacing(4))
	recipe.Gap = Set(float32(0))
	positionPanel(panel, anchor, t, popup)
	panel.Role(ui.RoleMenu).Label(o.Label).FocusGroup(ui.Vertical).MinWidth(0)
	finish(t, id, Panel, panel, recipe, o.Decoration)
	enabled := 0
	for _, it := range items {
		if menuEnabled(it) {
			enabled++
		}
	}
	if enabled == 0 {
		panel.Focusable().Focus()
	}
	ui.Scroll(c).MaxHeight(max(0, maxHeight-8)).MinWidth(0).Children(func() {
		for _, it := range items {
			if it.Kind == MenuSeparator {
				Separator(c.Key(it.ID), t, SeparatorOptions{}).Margin(4, 0)
				continue
			}
			if it.Kind == MenuLabel {
				finish(t, id, Header, ui.Text(c.Key(it.ID), it.Label).Padding(6), PartStyle{Weight: Set(600), Foreground: Set(t.Highlighted)}, o.Decoration)
				continue
			}
			row := ui.ButtonBase(c.Key(it.ID)).Role(ui.RoleMenuItem).Label(it.Label).Description(it.Description).Disabled(it.Disabled).MinWidth(0).AlignItems(ui.Center).FocusRing(false).Bind(s.handles[it.ID])
			if it.Checked.Set {
				row.Role(ui.RoleMenuItemCheckBox).Checked(it.Checked.Value)
				if it.Radio {
					row.Role(ui.RoleMenuItemRadio)
				}
			}
			childOpen := len(s.path) > depth && s.path[depth] == it.ID
			if len(it.Children) > 0 {
				row.Expanded(childOpen)
			}
			if row.Focused() && len(s.path) > depth && !childOpen {
				s.path = s.path[:depth]
			}
			if level.hover == it.ID && !row.Hovered() && !childOpen {
				level.hover = ""
			}
			if level.hover == it.ID && len(it.Children) > 0 && !it.Disabled && !childOpen {
				remaining := 180*time.Millisecond - c.Now().Sub(level.hoverAt)
				if remaining <= 0 {
					s.child(depth, it, false)
					childOpen = true
				} else {
					c.After(remaining)
				}
			}
			highlighted := row.Focused() || childOpen
			bg := ui.Transparent
			if highlighted {
				bg = t.Elevated
			}
			fg := t.Text
			if it.Color != ColorDefault && it.Color != Neutral {
				fg = t.color(it.Color)
			}
			finish(t, id, Item, row, PartStyle{Background: Set(bg), Foreground: Set(fg), Font: Set(t.Font), FontSize: Set(t.text(14)), PaddingX: Set(t.spacing(6)), PaddingY: Set(t.spacing(6)), Gap: Set(t.spacing(8)), Radius: Set(t.radius(4))}, o.Decoration)
			row.TextCaret(ui.Rect{H: 1}).InputMethod(false)
			handle := s.handles[it.ID]
			rtl := c.Direction() == ui.RightToLeft
			row.HandleInput(func(ev ui.InputEvent) bool {
				if ev.Kind == ui.InputPointerMove {
					if level.hover != it.ID {
						level.hover = it.ID
						level.hoverAt = ev.Time
						handle.Focus()
						if len(s.path) > depth && s.path[depth] != it.ID {
							s.path = s.path[:depth]
						}
						return true
					}
					return false
				}
				if ev.Kind == ui.InputKeyDown || ev.Kind == ui.InputText {
					s.resetHover()
				}
				if ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyTab {
					s.close(false)
					return false
				}
				if ev.Kind == ui.InputText {
					text := strings.ToLower(strings.TrimSpace(ev.Text))
					if text == "" {
						return false
					}
					if ev.Time.Sub(level.typed) > time.Second {
						level.prefix = ""
					}
					level.typed = ev.Time
					level.prefix += text
					candidates := items
					at := -1
					for i, candidate := range candidates {
						if candidate.ID == it.ID {
							at = i
							break
						}
					}
					find := func(prefix string) bool {
						for offset := 1; offset <= len(candidates); offset++ {
							candidate := candidates[(at+offset)%len(candidates)]
							if menuEnabled(candidate) && strings.HasPrefix(strings.ToLower(candidate.Label), prefix) {
								s.focus = s.handles[candidate.ID]
								return true
							}
						}
						return false
					}
					if !find(level.prefix) {
						level.prefix = text
						find(text)
					}
					return true
				}
				if ev.Kind != ui.InputKeyDown || ev.Mods != 0 {
					return false
				}
				forward, back := ui.KeyRight, ui.KeyLeft
				if rtl {
					forward, back = back, forward
				}
				switch ev.Key {
				case ui.KeyEscape:
					if depth > 0 {
						s.path = s.path[:depth-1]
						s.focus = s.handles[parent]
					} else {
						s.close(true)
					}
					return true
				case back:
					if depth > 0 {
						s.path = s.path[:depth-1]
						s.focus = s.handles[parent]
						return true
					}
				case forward:
					if len(it.Children) > 0 {
						s.child(depth, it, true)
						return true
					}
				case ui.KeyTab:
					s.close(false)
					return false
				}
				return false
			})
			row.OnClick(func() {
				if len(it.Children) > 0 {
					s.child(depth, it, true)
				} else {
					s.close(true)
					if it.OnSelect != nil {
						it.OnSelect()
					}
				}
			})
			row.Children(func() {
				if it.Checked.Set {
					box := ui.Box(c).Size(16, 16)
					if it.Checked.Value {
						box.Children(func() { ui.Icon(c, lucide.Check()).Size(16, 16).TextColor(t.Primary) })
					}
				} else if it.Icon != nil {
					finish(t, id, Leading, ui.Icon(c, it.Icon).Size(18, 18).Shrink(0), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
				}
				if it.Build != nil {
					it.Build(c, t)
				} else {
					ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
						finish(t, id, Label, ui.Text(c, it.Label).SingleLine(), PartStyle{}, o.Decoration)
						if it.Description != "" {
							finish(t, id, Description, ui.Text(c, it.Description).SingleLine(), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(12))}, o.Decoration)
						}
					})
				}
				if len(it.Children) > 0 {
					angle := float32(-90)
					if rtl {
						angle = 90
					}
					finish(t, id, Trailing, ui.Icon(c, lucide.ChevronDown()).Size(16, 16).Rotate(angle), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
				} else if it.Shortcut != "" {
					finish(t, id, Trailing, ui.Text(c, it.Shortcut).SingleLine(), PartStyle{Foreground: Set(t.Dimmed), FontSize: Set(t.text(12))}, o.Decoration)
				}
			})
			if childOpen {
				childVisible := true
				childOptions := o
				childOptions.Label = it.Label
				child := ui.PopoverBase(c.Key(s.handles[it.ID]), row, &childVisible, func(child ui.Element) { menuPanel(c, t, child, row, s, it.Children, childOptions, id, it.ID, depth+1) })
				if child.Dismissed() && len(s.path) > depth {
					s.path = s.path[:depth]
					s.focus = s.handles[it.ID]
				}
			}
		}
		if len(items) == 0 {
			ui.Text(c, message(c, "noActions")).Padding(8).TextColor(t.Muted).Role(ui.RoleStatus)
		}
	})
}
