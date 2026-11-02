package nui

import (
	"context"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type ListboxItem struct {
	ID, Label, Description, Group string
	Disabled                      bool
	Icon                          *ui.SVG
}
type ListboxState struct {
	Active          string
	List            ui.ListState
	prefix          string
	typed           time.Time
	horizontal      ui.ScrollState
	scrollTo        string
	selectionAnchor string
	data            listboxData
	selection       stringSelection
}
type SelectionBehavior uint8

const (
	SelectionToggle SelectionBehavior = iota
	SelectionReplace
)

type ListboxOptions struct {
	// Build draws non-interactive row content synchronously; Label still supplies semantics.
	Build              func(*ui.Context, Theme, ListboxItem)
	Label, EmptyText   string
	Multiple, Disabled bool
	Height             float32
	Orientation        Optional[Orientation] // Defaults to Vertical.
	ItemWidth          float32               // Horizontal item width, default 160 DIP.
	Loop               bool
	HighlightOnHover   Optional[bool]
	SelectionBehavior  SelectionBehavior
	OnChange           func([]string)
	OnSelect           func(string) // Activation, including selecting an unchanged value.
	Decoration
}

func nextListIndex(items []ListboxItem, at, delta int) int {
	for i := at + delta; i >= 0 && i < len(items); i += delta {
		if !items[i].Disabled {
			return i
		}
	}
	return at
}
func Listbox(c *ui.Context, t Theme, selected *[]string, s *ListboxState, items []ListboxItem, o ListboxOptions) ui.Element {
	vert := vertical(valueOr(o.Orientation, Vertical))
	if o.SelectionBehavior > SelectionReplace {
		panic("nui: invalid selection behavior")
	}
	if selected == nil || s == nil {
		panic("nui: Listbox requires selection and state")
	}
	s.data.update(items)
	s.selection.update(*selected)
	items = s.data.items
	rows, positions := s.data.rows, s.data.positions
	at := s.data.enabledIndex(s.Active)
	if at < 0 {
		at = nextListIndex(items, -1, 1)
		if at >= 0 {
			s.Active = items[at].ID
		} else {
			s.Active = ""
		}
	}
	publish := func(next []string) {
		if !slices.Equal(next, *selected) {
			*selected = next
			s.selection.update(next)
			if o.OnChange != nil {
				o.OnChange(slices.Clone(next))
			}
		}
	}
	choose := func(id string, toggle bool) {
		if o.Disabled || t.field.Disabled {
			return
		}
		next := slices.Clone(*selected)
		if toggle && (o.Multiple || slices.Contains(next, id)) {
			if slices.Contains(next, id) {
				next = slices.DeleteFunc(next, func(v string) bool { return v == id })
			} else {
				next = append(next, id)
			}
		} else {
			next = []string{id}
			s.selectionAnchor = id
		}
		publish(next)
		if o.OnSelect != nil {
			o.OnSelect(id)
		}
	}
	height := o.Height
	if height <= 0 {
		height = 224
		if !vert {
			height = 64
		}
	}
	root := ui.Column(c).MinWidth(0).Height(height).Disabled(o.Disabled || t.field.Disabled)
	finish(t, ComponentListbox, Root, root, PartStyle{Background: Set(t.Background), Radius: Set(t.radius(6)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}, o.Decoration)
	root.Children(func() {
		if len(items) == 0 {
			text := o.EmptyText
			if text == "" {
				text = message(c, "noResults")
			}
			finish(t, ComponentListbox, EmptySlot, ui.Text(c, text).Role(ui.RoleStatus).Padding(12), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
		}
		var list ui.Element
		itemWidth := o.ItemWidth
		if itemWidth <= 0 {
			itemWidth = 160
		}
		first, last := 0, len(rows)
		if vert {
			list = ui.List(c.Key("list"), &s.List, len(rows)).ItemKey(func(i int) any { return rows[i].key })
		} else {
			viewport := s.horizontal.Bounds(c).W
			if viewport <= 0 {
				viewport, _ = c.Size()
				c.After(time.Nanosecond)
			}
			if i := s.data.enabledIndex(s.scrollTo); s.scrollTo != "" && i >= 0 {
				x := float32(positions[i]) * itemWidth
				if x < s.horizontal.X {
					s.horizontal.X = x
				} else if x+itemWidth > s.horizontal.X+viewport {
					s.horizontal.X = max(0, x+itemWidth-viewport)
				}
			}
			s.scrollTo = ""
			s.horizontal.X = max(0, min(s.horizontal.X, float32(len(rows))*itemWidth-viewport))
			first = max(0, int(s.horizontal.X/itemWidth)-2)
			last = min(len(rows), int(math.Ceil(float64((s.horizontal.X+viewport)/itemWidth)))+2)
			list = ui.ScrollHorizontal(c.Key("list")).TrackScroll(&s.horizontal).Bind(&s.List.Handle).Role(ui.RoleList)
		}
		list.Fill().Focusable().Label(o.Label).InputMethod(false)
		bindField(c, list, t)
		move := func(i int) {
			if i >= 0 && i < len(items) && !items[i].Disabled {
				s.Active = items[i].ID
				if vert {
					s.List.ScrollIntoView(positions[i])
				} else {
					s.scrollTo = items[i].ID
				}
			}
		}
		next := func(at, delta int) int {
			i := nextListIndex(items, at, delta)
			if i == at && o.Loop {
				if delta > 0 {
					i = nextListIndex(items, -1, 1)
				} else {
					i = nextListIndex(items, len(items), -1)
				}
			}
			return i
		}
		back, forth := ui.KeyUp, ui.KeyDown
		if !vert {
			back, forth = ui.KeyLeft, ui.KeyRight
			if c.Direction() == ui.RightToLeft {
				back, forth = forth, back
			}
		}
		list.HandleInput(func(ev ui.InputEvent) bool {
			at := s.data.enabledIndex(s.Active)
			if ev.Kind == ui.InputText && ev.Text != "" {
				if ev.Time.Sub(s.typed) > time.Second {
					s.prefix = ""
				}
				s.typed = ev.Time
				s.prefix += strings.ToLower(ev.Text)
				for j := 1; j <= len(items); j++ {
					i := (max(at, 0) + j) % len(items)
					if !items[i].Disabled && strings.HasPrefix(strings.ToLower(items[i].Label), s.prefix) {
						move(i)
						break
					}
				}
				return true
			}
			if ev.Kind != ui.InputKeyDown {
				return false
			}
			if o.Multiple && ev.Key == ui.KeyA && (ev.Mods == ui.Ctrl || ev.Mods == ui.Super) {
				all := []string{}
				for _, it := range items {
					if !it.Disabled {
						all = append(all, it.ID)
					}
				}
				publish(all)
				move(nextListIndex(items, len(items), -1))
				return true
			}
			if ev.Mods & ^ui.Shift != 0 {
				return false
			}
			before := s.Active
			switch ev.Key {
			case forth:
				move(next(at, 1))
			case back:
				move(next(at, -1))
			case ui.KeyHome:
				move(nextListIndex(items, -1, 1))
			case ui.KeyEnd:
				move(nextListIndex(items, len(items), -1))
			case ui.KeySpace:
				if at >= 0 && ev.Mods == 0 {
					choose(items[at].ID, o.SelectionBehavior != SelectionReplace)
				}
				return ev.Mods == 0
			default:
				return false
			}
			if ev.Mods == ui.Shift && o.Multiple && o.SelectionBehavior == SelectionReplace {
				anchor := s.data.enabledIndex(s.selectionAnchor)
				if anchor < 0 {
					s.selectionAnchor = before
					anchor = s.data.enabledIndex(before)
				}
				to := s.data.enabledIndex(s.Active)
				if anchor >= 0 && to >= 0 {
					rangeIDs := []string{}
					for i := min(anchor, to); i <= max(anchor, to); i++ {
						if !items[i].Disabled {
							rangeIDs = append(rangeIDs, items[i].ID)
						}
					}
					publish(rangeIDs)
				}
			}
			return true
		})
		list.OnShortcut(0, ui.KeyEnter, func() {
			if i := s.data.enabledIndex(s.Active); i >= 0 {
				choose(items[i].ID, o.SelectionBehavior != SelectionReplace)
			}
		})
		buildRow := func(r ui.ListRow) {
			entry := rows[r.Index]
			if entry.item < 0 {
				header := ui.Text(r.Context, entry.group).Height(t.text(28)).Padding(6, 8).FontWeight(600)
				if !vert {
					header.Width(itemWidth).MinWidth(itemWidth).MaxWidth(itemWidth).Shrink(0).SingleLine()
				}
				finish(t, ComponentListbox, Label, header, PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(12))}, o.Decoration)
				return
			}
			it := items[entry.item]
			chosen := s.selection.set[it.ID]
			// Rows with a description grow; the list measures each row it shows.
			row := ui.Row(r.Context).MinWidth(0).MinHeight(t.text(40)).AlignItems(ui.Center).Gap(t.spacing(8)).Padding(6, 8).Role(ui.RoleListItem).Label(it.Label).Description(it.Description).Selected(chosen).Background(ui.Transparent).Disabled(it.Disabled)
			if !vert {
				row.Width(itemWidth).MinWidth(itemWidth).MaxWidth(itemWidth).Shrink(0).Height(height)
			}
			if it.ID == s.Active {
				row.Background(t.Elevated)
				list.ActiveDescendant(row)
			}
			row.HandleInput(func(ev ui.InputEvent) bool {
				if ev.Kind == ui.InputPointerMove && !it.Disabled && valueOr(o.HighlightOnHover, true) && s.Active != it.ID {
					s.Active = it.ID
					return true
				}
				return false
			})
			row.OnClick(func() {
				s.Active = it.ID
				choose(it.ID, o.SelectionBehavior != SelectionReplace)
			})
			finish(t, ComponentListbox, Item, row, PartStyle{Radius: Set(t.radius(4)), Foreground: Set(t.Text)}, o.Decoration)
			row.Children(func() {
				if it.Icon != nil {
					ui.Icon(r.Context, it.Icon).Size(16, 16)
				}
				if o.Build != nil {
					o.Build(r.Context, t, it)
				} else {
					ui.Column(r.Context).Grow(1).MinWidth(0).Children(func() {
						ui.Text(r.Context, it.Label).SingleLine()
						if it.Description != "" {
							ui.Text(r.Context, it.Description).FontSize(t.text(11)).TextColor(t.Muted)
						}
					})
				}
				if chosen {
					ui.Icon(r.Context, lucide.Check()).Size(16, 16).TextColor(t.Primary)
				}
			})
		}
		if vert {
			list.Rows(buildRow)
		} else {
			list.Children(func() {
				ui.Row(c).Width(float32(len(rows)) * itemWidth).Height(height).Shrink(0).Children(func() {
					ui.Box(c.Key(listboxSpacer(false))).Width(float32(first) * itemWidth).Shrink(0)
					for i := first; i < last; i++ {
						entry := rows[i]
						buildRow(ui.ListRow{Context: c.Key(entry.key), Index: i})
					}
					ui.Box(c.Key(listboxSpacer(true))).Width(float32(len(rows)-last) * itemWidth).Shrink(0)
				})
			})
		}
	})
	return root
}

type SelectMenuState struct {
	Open           bool
	Query          string
	Listbox        ListboxState
	queryHandle    ui.Handle
	wasOpen        bool
	labels         map[string]string
	data           listboxItems
	labelIDs       []string
	label          string
	filtered       []ListboxItem
	filterQuery    string
	filterFields   []string
	filterRevision uint64
}
type SelectMenuOptions struct {
	ListboxOptions
	Color                                          Color
	Variant                                        Variant
	Size                                           Size
	SearchInput                                    Optional[bool]
	ResetSearchTermOnBlur, ResetSearchTermOnSelect Optional[bool]
	Placeholder                                    string
	Invalid                                        bool
	ErrorMessage                                   string
	SearchPlaceholder                              string
	Clearable, Loading, ExternalFilter             bool
	Error                                          string
	OnQuery                                        func(string)
	OnCreate                                       func(string)
	FilterFields                                   []string // label (default), description, value/id, group.
	CreateWhenAlways, CreateAtTop                  bool
	Popup                                          PopoverOptions
}

func SelectMenu(c *ui.Context, t Theme, value *[]string, s *SelectMenuState, items []ListboxItem, o SelectMenuOptions) ui.Element {
	if s == nil || value == nil {
		panic("nui: SelectMenu requires value and state")
	}
	dataChanged := s.data.update(items)
	items = s.data.items
	resetQuery := func() {
		if s.Query == "" {
			return
		}
		s.Query = ""
		if o.OnQuery != nil {
			o.OnQuery("")
		}
	}
	if s.wasOpen && !s.Open && valueOr(o.ResetSearchTermOnBlur, true) {
		resetQuery()
	}
	search := valueOr(o.SearchInput, true)
	config := selected(defaults(t, ComponentSelectMenu, ComponentTheme{Color: Primary, Variant: Outline, Size: MD}), ComponentTheme{Color: o.Color, Variant: o.Variant, Size: inheritedSize(t, o.Size)})
	root := ui.Column(c).MinWidth(0)
	finish(t, ComponentSelectMenu, Root, root, PartStyle{Gap: Set(t.spacing(4))}, o.Decoration)
	root.Children(func() {
		label := s.selectionLabel(*value, dataChanged)
		if label == "" {
			label = o.Placeholder
			if label == "" {
				label = message(c, "chooseOptions")
			}
		}
		anchor := ui.ButtonBase(c.Key("trigger")).FillWidth().MinWidth(0).FocusRing(false).Disabled(o.Disabled || t.field.Disabled)
		name := o.Label
		if name == "" {
			name = label
		}
		anchor.Label(name)
		field := InputOptions{Invalid: o.Invalid, ErrorMessage: o.ErrorMessage}
		hover := anchor.Hovered() && !anchor.IsDisabled()
		styles := inputRecipe(t, config, anchor.FocusVisible(), hover, fieldInvalid(t, field))
		style := styles[Root]
		if hover && config.Variant == Outline {
			style.Background = Set(t.Elevated)
		}
		style.apply(anchor)
		if anchor.FocusVisible() && config.Variant != None {
			color := config.Color
			if fieldInvalid(t, field) {
				color = Error
			}
			anchor.Shadow(0, 0, 0, 3, t.focusColor(color))
		}
		anchor.Children(func() {
			finish(t, ComponentSelectMenu, Label, ui.Text(c, label).SingleLine().MinWidth(0).Grow(1), PartStyle{}, o.Decoration)
			m := buttonMetrics(t, config.Size)
			finish(t, ComponentSelectMenu, Trailing, ui.Icon(c, lucide.ChevronDown()).Size(m.Icon, m.Icon).TextColor(t.Dimmed).Shrink(0), PartStyle{}, o.Decoration)
		})
		fieldSemantics(c, anchor, t, field)
		anchor.OnClick(func() { s.Open = !s.Open }).OnShortcut(0, ui.KeyDown, func() { s.Open = true }).OnShortcut(0, ui.KeyUp, func() { s.Open = true })
		// Group corners follow part styles and precede the caller's Style callback, as for Input.
		trigger := mergeStyle(t.components[ComponentSelectMenu].UI[Trigger], o.UI[Trigger])
		trigger.apply(anchor)
		applyGroup(c, anchor, t, mergeValue(style.Radius, trigger.Radius).Value)
		if o.Style != nil {
			o.Style(Trigger, anchor)
		}
		popup := o.Popup
		popup.MatchAnchor = true
		if !popup.MaxHeight.Set {
			popup.MaxHeight = Set(float32(400))
		}
		Popover(c.Key("popup"), t, anchor, &s.Open, popup, func(inner Theme) {
			if search {
				in := Input(c.Key("search"), inner, &s.Query, InputOptions{Label: namedMessage(c, "searchNamed", o.Label), Placeholder: o.SearchPlaceholder, Leading: lucide.Search(), OnChange: o.OnQuery, Style: func(root, editor, leading, trailing ui.Element) {
					editor.Bind(&s.queryHandle).OnTextKey(func(k ui.TextKey) bool {
						if k.Mods == 0 && (k.Key == ui.KeyDown || k.Key == ui.KeyUp) {
							s.Listbox.List.Handle.Focus(c)
							return true
						}
						return false
					})
				}})
				_ = in
				if !s.wasOpen {
					s.queryHandle.Focus(c)
				}
			} else if !s.wasOpen {
				s.Listbox.List.Handle.Focus(c)
			}
			filtered := items
			if search && !o.ExternalFilter && s.Query != "" {
				filtered = s.filter(s.Query, o.FilterFields)
			}
			if o.Loading {
				ui.Text(c, message(c, "loading")).Role(ui.RoleStatus)
			}
			if o.Error != "" {
				ui.Text(c, o.Error).Role(ui.RoleStatus).TextColor(t.Error)
			}
			opts := o.ListboxOptions
			opts.Label = namedMessage(c, "optionsNamed", o.Label)
			if !o.Multiple {
				opts.SelectionBehavior = SelectionReplace
			}
			opts.OnSelect = func(id string) {
				if valueOr(o.ResetSearchTermOnSelect, true) {
					resetQuery()
				}
				if !o.Multiple {
					s.Open = false
				}
				if o.OnSelect != nil {
					o.OnSelect(id)
				}
			}
			create := o.OnCreate != nil && strings.TrimSpace(s.Query) != "" && (o.CreateWhenAlways || len(filtered) == 0)
			create = create && !slices.ContainsFunc(filtered, func(it ListboxItem) bool { return it.ID == s.Query })
			buildCreate := func() {
				if create {
					Button(c.Key("create"), inner, namedMessage(c, "createItem", s.Query), ButtonOptions{Variant: Ghost}).OnClick(func() { o.OnCreate(strings.TrimSpace(s.Query)) })
				}
			}
			if o.CreateAtTop {
				buildCreate()
			}
			Listbox(c.Key("choices"), inner, value, &s.Listbox, filtered, opts).FillWidth()
			if !o.CreateAtTop {
				buildCreate()
			}
			if o.Clearable && len(*value) > 0 {
				Button(c.Key("clear"), inner, message(c, "clearSelection"), ButtonOptions{Variant: Ghost}).OnClick(func() {
					*value = nil
					if o.OnChange != nil {
						o.OnChange(nil)
					}
				})
			}
		})
		if s.wasOpen && !s.Open && valueOr(o.ResetSearchTermOnBlur, true) {
			resetQuery()
		}
		s.wasOpen = s.Open
	})
	return root
}

// QueryState is UI-thread-owned. Start cancels obsolete work; Complete rejects stale tickets.
// Workers must publish with Window.Update, never mutate this state directly.
type QueryTicket struct {
	owner   *byte
	version uint64
	Query   string
}
type QueryState[T any] struct {
	identity *byte
	Items    []T
	Loading  bool
	Error    string
	version  uint64
	cancel   context.CancelFunc
}

func (s *QueryState[T]) Start(parent context.Context, query string) (context.Context, QueryTicket) {
	s.Cancel()
	s.Loading = true
	s.Error = ""
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	if s.identity == nil {
		s.identity = new(byte)
	}
	return ctx, QueryTicket{s.identity, s.version, query}
}
func (s *QueryState[T]) Cancel() {
	s.version++
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.Loading = false
}
func (s *QueryState[T]) Complete(ticket QueryTicket, items []T, err error) bool {
	if !s.Loading || ticket.owner != s.identity || ticket.version != s.version {
		return false
	}
	s.Loading = false
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.Items = slices.Clone(items)
	s.Error = ""
	if err != nil {
		s.Error = err.Error()
	}
	return true
}
