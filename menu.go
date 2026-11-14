package nui

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
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
	Avatar                           *AvatarOptions
	Color                            Color
	Disabled, Radio, Loading         bool
	Checked                          Optional[bool]
	Close                            Optional[bool]
	Children                         []MenuItem
	OnSelect                         func()
	OnCheckedChange                  func(bool)
	UI                               Styles
	Content                          *MenuContentOptions
	// Build supplies non-interactive visual content; Label is still required for semantics/typeahead.
	Build func(*ui.Context, Theme)
}
type MenuOptions struct {
	Label              string
	Size               Size
	Disabled, Loop     bool
	CheckedIcon        *ui.SVG
	LoadingIcon        *ui.SVG
	Width, MaxHeight   Optional[float32]
	Placement          Optional[Placement]
	Offset             Optional[float32]
	Filter             bool
	FilterFields       []string // label (default), description, id, shortcut.
	IgnoreFilter       bool
	SearchTerm         *string
	FilterInput        *InputOptions
	OnSearchTermChange func(string)
	OnOpenChange       func(bool)
	Decoration
}

// MenuContentOptions overrides one submenu. Size and icons inherit from the root.
type MenuContentOptions struct {
	Label                    string
	Width, MaxHeight, Offset Optional[float32]
	Placement                Optional[Placement]
	Loop                     Optional[bool]
	Filter                   bool
	FilterFields             []string
	IgnoreFilter             Optional[bool]
	SearchTerm               *string
	FilterInput              *InputOptions
	OnSearchTermChange       func(string)
	Decoration
}

// shortcutKeys splits a displayed shortcut such as "Ctrl+Shift+P" into its
// keys; a "+" that cannot separate two keys, as in "Ctrl++", is a key.
func shortcutKeys(s string) []string {
	var keys []string
	for s != "" {
		i := strings.Index(s[1:], "+")
		if i < 0 || i == len(s)-2 {
			keys = append(keys, strings.TrimSpace(s))
			break
		}
		keys = append(keys, strings.TrimSpace(s[:i+1]))
		s = s[i+2:]
	}
	return keys
}

type menuLevel struct {
	prefix  string
	typed   time.Time
	hover   string
	hoverAt time.Time
	query   string
	filter  ui.Handle
}

// MenuState is window-owned data. Never share one instance between independent menu triggers.
// Open can be set by the application; all other fields preserve keyboard/hover identity.
type MenuState struct {
	Open         bool
	Query        string // Root filter draft when MenuOptions.SearchTerm is nil.
	shown        bool
	path         []string
	handles      map[string]*ui.Handle
	levels       map[string]*menuLevel
	trigger      ui.Handle
	panel        ui.Handle
	focus        *ui.Handle
	x, y         float32
	point        bool
	repositioned bool
}

func (s *MenuState) resetHover() {
	for _, level := range s.levels {
		level.hover = ""
	}
}
func (s *MenuState) close(restore bool, notify func(bool)) {
	changed := s.Open
	s.resetHover()
	s.focus = nil
	s.Open = false
	s.shown = false
	s.path = nil
	s.point, s.repositioned = false, false
	if restore {
		s.trigger.Focus()
	}
	if changed && notify != nil {
		notify(false)
	}
}
func (s *MenuState) prepare(c *ui.Context, items []MenuItem) {
	focusWithin := s.menuFocused(c)
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
			if it.ID == id && menuEnabled(it) && len(it.Children) > 0 {
				current = it.Children
				found = true
				break
			}
		}
		if !found {
			s.path = s.path[:depth]
			if focusWithin {
				s.focus = nil
				s.focusEdge(current, false)
			}
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
		if it.Content != nil && it.Content.Filter && (it.Content.FilterInput == nil || !it.Content.FilterInput.Disabled) {
			if s.levels[it.ID] == nil {
				s.levels[it.ID] = &menuLevel{}
			}
			s.focus = &s.levels[it.ID].filter
		} else {
			s.focusEdge(it.Children, false)
		}
	}
}
func DropdownMenu(c *ui.Context, t Theme, anchor ui.Element, state *MenuState, items []MenuItem, o MenuOptions) ui.Element {
	return menu(c, t, anchor, state, items, o, false, nil)
}
func ContextMenu(c *ui.Context, t Theme, anchor ui.Element, state *MenuState, items []MenuItem, o MenuOptions) ui.Element {
	return menu(c, t, anchor, state, items, o, true, nil)
}

type menuDecoration func(MenuItem, Decoration) Decoration

func menu(c *ui.Context, t Theme, anchor ui.Element, s *MenuState, items []MenuItem, o MenuOptions, context bool, decorate menuDecoration) ui.Element {
	// Consume the component key even when the popup is absent.
	ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	if s == nil {
		panic("nui: menu requires state")
	}
	s.prepare(c, items)
	if anchor.Valid() && !context {
		anchor.Role(ui.RoleMenuButton).Expanded(s.Open && !o.Disabled)
	}
	if anchor.Valid() && o.Disabled && !context {
		anchor.Disabled(true)
	}
	if !anchor.Valid() || anchor.IsDisabled() || o.Disabled {
		s.close(false, nil)
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
		changed := !s.Open
		s.Open = true
		s.path = nil
		s.focusEdge(items, last)
		if changed && o.OnOpenChange != nil {
			o.OnOpenChange(true)
		}
	}
	if context {
		anchor.Focusable()
		openFromKeyboard := func() {
			bounds := s.trigger.Bounds(c)
			s.point, s.x, s.y = true, bounds.X, bounds.Y+bounds.H
			open(false)
		}
		anchor.OnShortcut(0, ui.KeyContextMenu, openFromKeyboard).
			OnShortcut(ui.Shift, ui.KeyF10, openFromKeyboard)
		anchor.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 1 {
				bounds := s.trigger.Bounds(c)
				s.repositioned = s.Open
				s.point = true
				s.x = bounds.X + ev.X
				s.y = bounds.Y + ev.Y
				if !s.Open {
					open(false)
				}
				return true
			}
			return false
		})
	} else {
		anchor.Role(ui.RoleMenuButton).Expanded(s.Open)
		anchor.OnClick(func() {
			if s.Open {
				s.close(true, o.OnOpenChange)
			} else {
				open(false)
			}
		})
		anchor.OnShortcut(0, ui.KeyDown, func() { open(false) }).OnShortcut(0, ui.KeyUp, func() { open(true) })
	}
	if !s.Open {
		// Restore focus only when an application closes an active menu.
		restore := s.shown && s.menuFocused(c)
		s.close(restore, nil)
		return ui.Element{}
	}
	if !s.shown && s.focus == nil {
		for _, level := range s.levels {
			level.prefix = ""
		}
		s.focusEdge(items, false)
	}
	s.shown = true
	// A menu is independent of the field that supplied its trigger.
	t.field, t.group, t.form = fieldDefaults{}, groupPart{}, nil
	target := anchor
	if context && s.point {
		ui.Overlay(c.Key(&s.point), func() {
			target = ui.Box(c).Absolute().Left(s.x).Top(s.y).Size(0, 0).IgnoreInput(true)
		})
	}
	panel := ui.PopoverBase(c.Key(s), target, &s.Open, func(panel ui.Element) {
		panel.Bind(&s.panel)
		menuPanel(c, t, panel, target, s, items, o, id, "", 0, decorate)
	})
	// A context-menu press on its trigger moves the existing popup. The
	// virtual point is not the whole trigger, so PopoverBase sees that press
	// as outside; retain the existing open state for this one input only.
	if context && s.repositioned {
		s.Open, s.repositioned = true, false
	}
	if !s.Open {
		s.path = nil
		s.point = false
		panel.SetChanged().OnChange(func() {
			if o.OnOpenChange != nil {
				o.OnOpenChange(false)
			}
		})
	}
	if s.Open && s.focus != nil {
		s.focus.Focus()
	}
	s.focus = nil
	return panel
}
func menuPanel(c *ui.Context, t Theme, panel, anchor ui.Element, s *MenuState, items []MenuItem, o MenuOptions, id Component, parent string, depth int, decorate menuDecoration) {
	if s.levels[parent] == nil {
		s.levels[parent] = &menuLevel{}
	}
	level := s.levels[parent]
	size := selected(defaults(t, id, ComponentTheme{Size: MD}), ComponentTheme{Size: o.Size}).Size
	m := optionRecipe(t, size)
	query := &level.query
	if depth == 0 {
		query = &s.Query
	}
	if o.SearchTerm != nil {
		query = o.SearchTerm
	}
	original := items
	items = filterMenuItems(items, *query, o)
	visible := make(map[string]bool, len(items))
	for _, it := range items {
		visible[it.ID] = menuEnabled(it)
	}
	for _, it := range original {
		if !visible[it.ID] && s.handles[it.ID].Focused(c) {
			if o.Filter {
				s.focus = &level.filter
			} else {
				s.focusEdge(items, false)
			}
		}
	}
	if len(s.path) > depth && !visible[s.path[depth]] {
		s.path = s.path[:depth]
	}
	maxHeight := float32(320)
	if o.MaxHeight.Set {
		maxHeight = o.MaxHeight.Value
	}
	w, h := c.Size()
	maxHeight = max(0, min(maxHeight, h-24))
	// Nuxt's min-w-32 menu takes its content's width unless Width is set.
	popup := PopoverOptions{MaxHeight: Set(maxHeight)}
	if id == ComponentContextMenu && depth == 0 {
		popup.Placement, popup.Offset = RightStart, Set(t.spacing(2))
	}
	if o.Width.Set {
		popup.Width = Set(min(o.Width.Value, w-24))
	}
	if depth > 0 {
		popup.Placement = RightStart
		if c.Direction() == ui.RightToLeft {
			popup.Placement = LeftStart
		}
		popup.Offset = Set(t.spacing(3))
	}
	if o.Placement.Set {
		popup.Placement = o.Placement.Value
	}
	popup.Offset = mergeValue(popup.Offset, o.Offset)
	recipe := popupStyle(t)
	recipe.PaddingX = Set(t.spacing(4))
	recipe.PaddingY = Set(t.spacing(4))
	recipe.Gap = Set(float32(0))
	positionPanel(panel, anchor, t, popup)
	panel.Role(ui.RoleMenu).Label(o.Label).FocusGroup(ui.Vertical).MinWidth(max(0, min(t.spacing(128), w-24))).MaxWidth(max(0, w-24)).Clip()
	originX, originY := popupOrigin(popup.Placement, c.Direction() == ui.RightToLeft)
	shown := popIn(c, t, panel, popupMotion, originX, originY)
	finishPanel(t, id, panel, recipe, o.Decoration, shown)
	enabled := 0
	for _, it := range items {
		if menuEnabled(it) {
			enabled++
		}
	}
	if enabled == 0 && !o.Filter {
		panel.Focusable().Focus()
	}
	if o.Filter {
		menuFilter(c, t, panel, s, level, query, items, o, size, id, parent, depth)
	} else {
		*ui.Local(panel, "nui-menu-filter", func() bool { return false }) = false
	}
	ui.Scroll(c.Key("viewport")).MaxHeight(max(0, maxHeight-8)).MinWidth(0).Children(func() {
		for _, it := range items {
			d := o.Decoration
			d.UI = mergeStyles(o.UI, it.UI)
			if decorate != nil {
				d = decorate(it, d)
			}
			if it.Kind == MenuSeparator {
				finish(t, id, SeparatorSlot, separator(c.Key(it.ID), t, SeparatorOptions{}, true).Margin(t.spacing(4), 0), PartStyle{}, d)
				continue
			}
			if it.Kind == MenuLabel {
				label := ui.Row(c.Key(it.ID)).MinWidth(0).AlignItems(ui.Start)
				style := optionItemStyle(t, m, false)
				style.Weight, style.Foreground = Set(600), Set(t.Highlighted)
				finish(t, id, Header, label, style, d)
				label.Children(func() { menuItemContent(c, t, it, o, d, id, size, m, t.Highlighted) })
				continue
			}
			row := ui.ButtonBase(c.Key(it.ID)).Role(ui.RoleMenuItem).Label(it.Label).Description(it.Description).Disabled(it.Disabled).MinWidth(0).AlignItems(ui.Start).FocusRing(false).Bind(s.handles[it.ID])
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
			filterFocused := o.Filter && level.filter.Focused(c)
			highlighted := row.Focused() || childOpen || filterFocused && level.hover == it.ID
			// Menus share the option row recipe; colored items tint
			// their highlight and keep their color, as Nuxt's do.
			style := optionItemStyle(t, m, highlighted)
			style.Font = Set(t.Font)
			leading := optionLeadingColor(t, highlighted)
			if it.Color != ColorDefault && it.Color != Neutral {
				color := t.color(it.Color)
				style.Foreground = Set(color)
				leading = color.Alpha(.75)
				if highlighted {
					style.Background = Set(color.Alpha(.10))
					leading = color
				}
			}
			finish(t, id, Item, row, style, d)
			row.TextCaret(ui.Rect{H: 1}).InputMethod(false)
			handle := s.handles[it.ID]
			rtl := c.Direction() == ui.RightToLeft
			row.HandleInput(func(ev ui.InputEvent) bool {
				if ev.Kind == ui.InputPointerMove {
					if level.hover != it.ID {
						level.hover = it.ID
						level.hoverAt = ev.Time
						if !filterFocused {
							handle.Focus()
						}
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
					s.close(false, o.OnOpenChange)
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
				case ui.KeyDown, ui.KeyUp:
					s.moveFocus(items, it.ID, ev.Key == ui.KeyUp, o.Loop)
					return true
				case ui.KeyHome, ui.KeyEnd:
					s.focusEdge(items, ev.Key == ui.KeyEnd)
					return true
				case ui.KeyEscape:
					if depth > 0 {
						s.path = s.path[:depth-1]
						s.focus = s.handles[parent]
					} else {
						s.close(true, o.OnOpenChange)
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
					s.close(false, o.OnOpenChange)
					return false
				}
				return false
			})
			row.OnClick(func() {
				s.activate(it, depth, o.OnOpenChange)
			})
			row.Children(func() { menuItemContent(c, t, it, o, d, id, size, m, leading) })
			if childOpen {
				childVisible := true
				childOptions := submenuOptions(o, it)
				child := ui.PopoverBase(c.Key(s.handles[it.ID]), row, &childVisible, func(child ui.Element) {
					menuPanel(c, t, child, row, s, it.Children, childOptions, id, it.ID, depth+1, decorate)
				})
				if child.Dismissed() && len(s.path) > depth {
					s.path = s.path[:depth]
					s.focus = s.handles[it.ID]
				}
			}
		}
		if len(items) == 0 {
			text := message(c, "noActions")
			if o.Filter && *query != "" {
				text = message(c, "noMatch")
			}
			empty := ui.Text(c, text).TextAlign(ui.Center).Role(ui.RoleStatus)
			finish(t, id, EmptySlot, empty, optionEmptyStyle(t, size), o.Decoration)
		}
	})
}
