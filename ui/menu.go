package nui

import (
	"slices"
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
	SearchExecution    SearchExecution
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
	SearchExecution          Optional[SearchExecution]
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
	search  menuSearchCache
	list    ui.ListState
	panel   ui.Handle
	view    menuView
	active  string
	heights map[string]float32
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
	source       submittedMenu
	dataChanged  bool
	shown        bool
	path         []string
	handles      map[string]*ui.Handle
	levels       map[string]*menuLevel
	trigger      ui.Handle
	panel        ui.Handle
	focus        *ui.Handle
	focusID      string
	activeLevels []string
	x, y         float32
	point        bool
	repositioned bool
	pop          popState // the root panel's, for popIn
}

func (s *MenuState) resetHover() {
	for _, parent := range s.activeLevels {
		level := s.levels[parent]
		level.hover = ""
	}
}
func (s *MenuState) close(restore bool, notify func(bool)) {
	if !s.Open && !s.shown && s.focus == nil && len(s.path) == 0 {
		return
	}
	changed := s.Open
	if s.focus != nil {
		s.focus.CancelFocus()
	}
	for _, parent := range s.activeLevels {
		level := s.levels[parent]
		level.search.search.pause()
		level.filter.CancelFocus()
	}
	s.resetHover()
	s.focus, s.focusID = nil, ""
	s.activeLevels = nil
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
func (s *MenuState) prepare(c *ui.Context) {
	if s.handles == nil {
		s.handles = map[string]*ui.Handle{}
		s.levels = map[string]*menuLevel{}
	}

	if !s.dataChanged {
		return
	}
	s.dataChanged = false
	focusWithin := (s.Open || s.shown) && s.menuFocused(c)
	seen := s.source.byID
	for _, parent := range s.activeLevels {
		level := s.levels[parent]
		if handle := s.handles[level.active]; handle != nil && handle.Focused(c) {
			if location, ok := seen[level.active]; ok && location.parent == parent && menuEnabled(*location.item) {
				s.focusItem(level.active)
			}
		}
	}
	removedFocus := false
	for id, handle := range s.handles {
		if _, exists := seen[id]; !exists {
			removedFocus = removedFocus || handle.Focused(c)
			delete(s.handles, id)
			if level := s.levels[id]; level != nil {
				level.search.search.reset()
			}
			delete(s.levels, id)
		}
	}
	for id := range seen {
		if s.handles[id] == nil {
			s.handles[id] = &ui.Handle{}
		}
	}
	s.activeLevels = slices.DeleteFunc(s.activeLevels, func(parent string) bool { return s.levels[parent] == nil })
	if removedFocus {
		s.focusEdge(s.source.levels[""].view(), false)
	}
	parent := ""
	for depth, id := range s.path {
		location, exists := s.source.byID[id]
		if !exists || location.parent != parent || !menuEnabled(*location.item) || len(location.item.Children) == 0 {
			s.path = s.path[:depth]
			if focusWithin {
				s.focus, s.focusID = nil, ""
				s.focusEdge(s.source.levels[parent].view(), false)
			}
			break
		}
		parent = id
	}
}
func menuEnabled(it MenuItem) bool { return it.Kind == MenuAction && !it.Disabled }
func (s *MenuState) focusItem(id string) {
	location, ok := s.source.byID[id]
	if !ok || !menuEnabled(*location.item) {
		return
	}
	if s.focus != nil && s.focus != s.handles[id] {
		s.focus.CancelFocus()
	}
	s.focus, s.focusID = s.handles[id], id
	if level := s.levels[location.parent]; level != nil {
		if at := level.view.enabledAt(id); at >= 0 {
			level.list.ScrollIntoView(at)
		}
	}
}

func (s *MenuState) focusEdge(items menuView, last bool) {
	if len(items.enabled) == 0 {
		return
	}
	at := items.enabled[0]
	if last {
		at = items.enabled[len(items.enabled)-1]
	}
	s.focusItem(items.item(at).ID)
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
			s.focus, s.focusID = &s.levels[it.ID].filter, ""
		} else {
			s.focusEdge(s.source.levels[it.ID].view(), false)
		}
	}
}
func DropdownMenu(c *ui.Context, t Theme, anchor ui.Element, state *MenuState, o MenuOptions) ui.Element {
	return menu(c, t, anchor, state, o, false, nil)
}
func ContextMenu(c *ui.Context, t Theme, anchor ui.Element, state *MenuState, o MenuOptions) ui.Element {
	return menu(c, t, anchor, state, o, true, nil)
}

type menuDecoration func(MenuItem, Decoration) (MenuItem, Decoration)

func menu(c *ui.Context, t Theme, anchor ui.Element, s *MenuState, o MenuOptions, context bool, decorate menuDecoration) ui.Element {
	// Consume the component key even when the popup is absent.
	ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
	if s == nil {
		panic("nui: menu requires state")
	}
	s.prepare(c)
	s.pop.update(c, t, s.Open, s)
	for _, parent := range s.activeLevels {
		level := s.levels[parent]
		if !s.Open || parent != "" && !slices.Contains(s.path, parent) {
			level.search.search.pause()
			level.filter.CancelFocus()
			if parent != "" && !t.fades(c) {
				cancelOverlayExit(c, s.handles[parent])
			}
		}
	}
	defer func() { s.pop.note(c, s.Open) }()
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
		for _, parent := range s.activeLevels {
			s.levels[parent].prefix = ""
		}
		changed := !s.Open
		s.Open = true
		s.path = nil
		s.focusEdge(s.source.levels[""].view(), last)
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
		for _, parent := range s.activeLevels {
			s.levels[parent].prefix = ""
		}
		s.focusEdge(s.source.levels[""].view(), false)
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
	s.activeLevels = s.activeLevels[:0]
	panel := ui.PopoverBase(c.Key(s.pop.key(s)), target, &s.Open, func(panel ui.Element) {
		panel.Bind(&s.panel)
		menuPanel(c, t, panel, target, s, o, id, "", 0, decorate)
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
	s.focus, s.focusID = nil, ""
	return panel
}
func menuPanel(c *ui.Context, t Theme, panel, anchor ui.Element, s *MenuState, o MenuOptions, id Component, parent string, depth int, decorate menuDecoration) {
	if s.levels[parent] == nil {
		s.levels[parent] = &menuLevel{}
	}
	level := s.levels[parent]
	s.activeLevels = append(s.activeLevels, parent)
	panel.Bind(&level.panel)
	level.search.search.bind(c, panel, func() {
		level.filter.CancelFocus()
		level.list = ui.ListState{}
		level.view, level.heights = menuView{}, nil
		if parent == "" && s.levels[parent] == level {
			s.Close()
		}
	})
	size := selected(defaults(t, id, ComponentTheme{Size: MD}), ComponentTheme{Size: o.Size}).Size
	m := optionRecipe(t, size)
	query := &level.query
	if depth == 0 {
		query = &s.Query
	}
	if o.SearchTerm != nil {
		query = o.SearchTerm
	}
	items := level.search.apply(s.source.levels[parent], *query, o)
	level.view = items
	if level.active != "" && items.enabledAt(level.active) < 0 {
		if handle := s.handles[level.active]; handle != nil && handle.Focused(c) {
			if o.Filter {
				s.focus, s.focusID = &level.filter, ""
			} else {
				s.focusEdge(items, false)
			}
		}
		level.active = ""
	}
	if len(s.path) > depth && items.enabledAt(s.path[depth]) < 0 {
		s.path = s.path[:depth]
	}
	if s.focusID != "" {
		if at := items.enabledAt(s.focusID); at >= 0 {
			level.list.ScrollIntoView(at)
		} else if location, exists := s.source.byID[s.focusID]; exists && location.parent == parent {
			s.focus.CancelFocus()
			s.focus, s.focusID = nil, ""
			if o.Filter {
				s.focus = &level.filter
			} else {
				s.focusEdge(items, false)
			}
		}
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
	// A submenu opened again as it leaves enters anew.
	var pop *popState
	if parent == "" {
		pop = &s.pop
	}
	shown := popIn(c, t, pop, panel, popupMotion, originX, originY)
	popOut(c, t, panel)
	finishPanel(t, id, panel, recipe, o.Decoration, shown)
	if len(items.enabled) == 0 && !o.Filter {
		panel.Focusable().Focus()
	}
	if o.Filter {
		menuFilter(c, t, panel, s, level, query, items, o, size, id, parent, depth)
	} else {
		*ui.Local(panel, "nui-menu-filter", func() bool { return false }) = false
	}
	if items.count() == 0 {
		text := message(c, "noActions")
		if level.search.search.pending {
			text = message(c, "loading")
		} else if o.Filter && *query != "" {
			text = message(c, "noMatch")
		}
		empty := statusText(c, "empty", text).TextAlign(ui.Center)
		finish(t, id, EmptySlot, empty, optionEmptyStyle(t, size), o.Decoration)
		return
	}
	list := ui.List(c.Key("viewport"), &level.list, items.count()).ItemKey(func(i int) any { return items.item(i).ID })
	limit := max(0, maxHeight-2*t.spacing(4))
	height := menuViewportHeight(t, m, items, level.heights, limit)
	list.Height(height).Basis(height).MaxHeight(limit).MinWidth(0).MinHeight(0)
	list.Rows(func(r ui.ListRow) {
		c, i := r.Context, r.Index
		it := items.item(i)
		d := o.Decoration
		d.UI = mergeStyles(o.UI, it.UI)
		if decorate != nil {
			it, d = decorate(it, d)
		}
		if it.Kind == MenuSeparator {
			line := separator(c.Key(it.ID), t, SeparatorOptions{}, true).Margin(t.spacing(4), 0)
			finish(t, id, SeparatorSlot, line, PartStyle{}, d)
			level.measure(c, line, it.ID, 2*t.spacing(4))
			return
		}
		if it.Kind == MenuLabel {
			label := ui.Row(c.Key(it.ID)).MinWidth(0).AlignItems(ui.Start)
			style := optionItemStyle(t, m, false)
			style.Weight, style.Foreground = Set(600), Set(t.Highlighted)
			finish(t, id, Header, label, style, d)
			label.Children(func() { menuItemContent(c, t, it, o, d, id, size, m, t.Highlighted) })
			level.measure(c, label, it.ID, 0)
			return
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
		if row.Focused() {
			level.active = it.ID
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
		t.fadeColors(c, row)
		row.TextCaret(ui.Rect{H: 1}).InputMethod(false)
		handle := s.handles[it.ID]
		rtl := c.Direction() == ui.RightToLeft
		row.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerMove {
				if level.hover != it.ID {
					level.hover = it.ID
					level.hoverAt = ev.Time
					if !filterFocused {
						level.active = it.ID
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
				at := candidates.index[it.ID]
				find := func(prefix string) bool {
					for offset := 1; offset <= candidates.count(); offset++ {
						candidate := candidates.item((at + offset) % candidates.count())
						if menuEnabled(candidate) && strings.HasPrefix(strings.ToLower(candidate.Label), prefix) {
							s.focusItem(candidate.ID)
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
					s.focusItem(parent)
				} else {
					s.close(true, o.OnOpenChange)
				}
				return true
			case back:
				if depth > 0 {
					s.path = s.path[:depth-1]
					s.focusItem(parent)
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
			if level.search.search.ready(*query) {
				s.activate(it, depth, o.OnOpenChange)
			}
		})
		row.Children(func() { menuItemContent(c, t, it, o, d, id, size, m, leading) })
		level.measure(c, row, it.ID, 0)
		if childOpen {
			childVisible := true
			childOptions := submenuOptions(o, it)
			child := ui.PopoverBase(c.Key(s.handles[it.ID]), row, &childVisible, func(child ui.Element) {
				menuPanel(c, t, child, row, s, childOptions, id, it.ID, depth+1, decorate)
			})
			if child.Dismissed() && len(s.path) > depth {
				s.path = s.path[:depth]
				s.focusItem(it.ID)
			}
		}
	})
}

// Searching reports pending local filtering in any mounted menu level.
func (s *MenuState) Searching() bool {
	for _, level := range s.levels {
		if level.search.search.pending {
			return true
		}
	}
	return false
}

// Close dismisses every level and cancels local search/focus requests. It does
// not emit interaction callbacks or discard submitted command descriptors.
func (s *MenuState) Close() {
	s.close(false, nil)
	for _, level := range s.levels {
		level.search.search.reset()
		level.filter.CancelFocus()
	}
	s.trigger.CancelFocus()
	s.panel.CancelFocus()
}
