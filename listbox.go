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
	// embedded lists draw no focus ring of their own, as in SelectMenu's popup.
	embedded bool
	// actionID identifies a focusable action row, excluded from selection and Build.
	actionID string
	activate func()
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
		if id == o.actionID && o.activate != nil {
			o.activate()
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
	// Listbox has no Size or Color option yet: they come from the component
	// defaults, an enclosing FormField, or SelectMenu's own Size and Color.
	config := defaults(t, ComponentListbox, ComponentTheme{Color: Primary, Size: MD})
	if t.field.Size != SizeDefault {
		config.Size = t.field.Size
	}
	m := optionRecipe(t, config.Size)
	root := ui.Column(c).MinWidth(0).Height(height).Disabled(o.Disabled || t.field.Disabled)
	rootStyle := PartStyle{Background: Set(t.Background), Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}
	var list ui.Element
	root.Children(func() {
		if len(items) == 0 {
			text := o.EmptyText
			if text == "" {
				text = emptyMessage(c, "noData", "noResults")
			}
			pad := t.spacing([5]float32{12, 16, 24, 28, 32}[sizeIndex(config.Size)])
			finish(t, ComponentListbox, EmptySlot, ui.Text(c, text).Role(ui.RoleStatus).TextAlign(ui.Center), PartStyle{Foreground: Set(t.Muted), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font), PaddingX: Set(m.Pad), PaddingY: Set(pad)}, o.Decoration)
		}
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
		list.Fill().Focusable().FocusRing(false).Label(o.Label).InputMethod(false)
		if vert {
			list.Padding(t.spacing(4))
		}
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
					if !it.Disabled && it.ID != o.actionID {
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
						if !items[i].Disabled && items[i].ID != o.actionID {
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
				header := ui.Text(r.Context, entry.group).SingleLine()
				if !vert {
					header.Width(itemWidth).MinWidth(itemWidth).MaxWidth(itemWidth).Shrink(0)
				}
				finish(t, ComponentListbox, Label, header, optionLabelStyle(t, config.Size), o.Decoration)
				return
			}
			it := items[entry.item]
			chosen := it.ID != o.actionID && s.selection.set[it.ID]
			highlighted := it.ID == s.Active
			// Rows take their content's height; the list measures each row it shows.
			row := ui.Row(r.Context).MinWidth(0).AlignItems(ui.Center).Role(ui.RoleListItem).Label(it.Label).Description(it.Description).Selected(chosen).Disabled(it.Disabled)
			if !vert {
				row.Width(itemWidth).MinWidth(itemWidth).MaxWidth(itemWidth).Shrink(0).Height(height)
			}
			if highlighted {
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
			finish(t, ComponentListbox, Item, row, optionItemStyle(t, m, highlighted), o.Decoration)
			row.Children(func() {
				if it.Icon != nil {
					ui.Icon(r.Context, it.Icon).Size(m.Icon, m.Icon).TextColor(optionLeadingColor(t, highlighted)).Shrink(0)
				}
				if o.Build != nil && it.ID != o.actionID {
					o.Build(r.Context, t, it)
				} else {
					optionTexts(r.Context, t, it.Label, it.Description)
				}
				if chosen {
					optionCheck(r.Context, m)
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
	// Nuxt rings the list in its color while it has keyboard focus.
	if list.FocusVisible() && !o.embedded {
		rootStyle.BorderColor = Set(t.color(config.Color))
		root.Shadow(0, 0, 0, 3, t.focusColor(config.Color))
	}
	finish(t, ComponentListbox, Root, root, rootStyle, o.Decoration)
	return root
}

// listboxContentHeight is the height a vertical Listbox lays its rows out in:
// items and their descriptions, group headings, the p-1 inset, or the empty state.
func listboxContentHeight(t Theme, size Size, items []ListboxItem) float32 {
	m := optionRecipe(t, size)
	label := optionLabelStyle(t, size)
	heading := 2*m.Pad + label.FontSize.Value*label.LineHeight.Value
	if len(items) == 0 {
		return 2*t.spacing([5]float32{12, 16, 24, 28, 32}[sizeIndex(size)]) + m.Line
	}
	h, group := 2*t.spacing(4), ""
	for _, it := range items {
		if it.Group != "" && it.Group != group {
			h += heading
		}
		group = it.Group
		h += m.Line + 2*m.Pad
		if it.Description != "" {
			h += m.Line
		}
	}
	return h
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
		placeholder := label == ""
		if placeholder {
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
		if hover {
			switch config.Variant {
			case Outline:
				style.Background = Set(t.Elevated)
			case Subtle:
				style.Background = Set(t.Accented.Alpha(.75))
			}
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
			text := ui.Text(c, label).SingleLine().MinWidth(0).Grow(1)
			if placeholder {
				text.TextColor(t.Dimmed)
			}
			finish(t, ComponentSelectMenu, Label, text, PartStyle{}, o.Decoration)
			m := buttonMetrics(t, config.Size)
			if o.Clearable && len(*value) > 0 && !anchor.IsDisabled() {
				// Nuxt's trailing clear: an unpadded neutral link button in place of the chevron.
				clear := ui.ButtonBase(c.Key("clear")).Label(message(c, "clearSelection")).KeepFocus().FocusRing(true).Shrink(0).OnClick(func() {
					*value = nil
					if o.OnChange != nil {
						o.OnChange(nil)
					}
				})
				color := t.Muted
				if clear.Hovered() {
					color = t.Text
				}
				clear.Children(func() { ui.Icon(c, lucide.X()).Size(m.Icon, m.Icon).TextColor(color) })
			} else {
				finish(t, ComponentSelectMenu, Trailing, ui.Icon(c, lucide.ChevronDown()).Size(m.Icon, m.Icon).TextColor(t.Dimmed).Shrink(0), PartStyle{}, o.Decoration)
			}
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
			popup.MaxHeight = Set(float32(240)) // Nuxt's max-h-60
		}
		// Nuxt's combobox content is flat: no padding around the search field and list.
		recipe := popupStyle(t)
		recipe.PaddingX, recipe.PaddingY, recipe.Gap = Set(float32(0)), Set(float32(0)), Set(float32(0))
		popover(c.Key("popup"), t, anchor, &s.Open, popup, recipe, func(inner Theme) {
			// The search field and list take the menu's Size and Color.
			inner.field.Size = config.Size
			inner = inner.WithComponent(ComponentListbox, ComponentTheme{Color: config.Color})
			m := optionRecipe(t, config.Size)
			used := float32(0)
			if search {
				bar := ui.Column(c.Key("search-bar")).MinWidth(0).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
				bar.Children(func() {
					Input(c.Key("search"), inner, &s.Query, InputOptions{Label: namedMessage(c, "searchNamed", o.Label), Placeholder: o.SearchPlaceholder, Leading: lucide.Search(), Variant: None, OnChange: o.OnQuery, Style: func(root, editor, leading, trailing ui.Element) {
						root.FillWidth()
						editor.Bind(&s.queryHandle).OnTextKey(func(k ui.TextKey) bool {
							if k.Mods == 0 && (k.Key == ui.KeyDown || k.Key == ui.KeyUp) {
								s.Listbox.List.Handle.Focus(c)
								return true
							}
							return false
						})
					}})
				})
				bm := buttonMetrics(t, config.Size)
				used += bm.Line + 2*bm.Y + 1
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
			status := func(text string, color ui.Color) {
				e := ui.Text(c, text).Role(ui.RoleStatus).TextColor(color).Shrink(0)
				optionItemStyle(t, m, false).apply(e)
				e.TextColor(color)
				used += m.Line + 2*m.Pad
			}
			if o.Loading {
				status(message(c, "loading"), t.Muted)
			}
			if o.Error != "" {
				status(o.Error, t.Error)
			}
			opts := o.ListboxOptions
			opts.Label = namedMessage(c, "optionsNamed", o.Label)
			if opts.EmptyText == "" && len(items) > 0 && len(filtered) == 0 {
				opts.EmptyText = emptyMessage(c, "noMatch", "noResults")
			}
			opts.embedded = true
			opts.UI = cloneStyles(opts.UI)
			opts.UI[Root] = mergeStyle(PartStyle{BorderWidth: Set(float32(0)), Radius: Set(float32(0)), Background: Set(ui.Transparent)}, opts.UI[Root])
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
			if create {
				// The create entry is a row of the list, which the arrow keys reach,
				// as in Nuxt; activating it requests creation instead of selecting.
				createID := "\x00nui:create"
				for slices.ContainsFunc(items, func(it ListboxItem) bool { return it.ID == createID }) {
					createID += "_"
				}
				entry := ListboxItem{ID: createID, Label: namedMessage(c, "createItem", s.Query)}
				if o.CreateAtTop {
					filtered = append([]ListboxItem{entry}, filtered...)
				} else {
					filtered = append(slices.Clone(filtered), entry)
				}
				opts.actionID = createID
				opts.activate = func() {
					o.OnCreate(strings.TrimSpace(s.Query))
				}
			}
			if opts.Height <= 0 {
				// The list fits its rows within the popup's maximum height.
				opts.Height = max(0, min(listboxContentHeight(t, config.Size, filtered), popup.MaxHeight.Value-used))
			}
			Listbox(c.Key("choices"), inner, value, &s.Listbox, filtered, opts).FillWidth()
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
