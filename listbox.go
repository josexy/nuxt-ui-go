package nui

import (
	"context"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

type ListboxItemKind uint8

const (
	ListboxOption ListboxItemKind = iota
	ListboxLabel
	ListboxSeparator
)

// ListboxItem remains comparable; optional per-item behavior and styles live
// behind Options so existing maps and equality-based data caches remain usable.
type ListboxItem struct {
	ID, Label, Description, Group string
	Kind                          ListboxItemKind
	Disabled                      bool
	Icon                          *ui.SVG
	Avatar                        *AvatarOptions
	Chip                          *ChipOptions
	Options                       *ListboxItemOptions
}

type ListboxItemOptions struct {
	OnSelect func(*ListboxEvent)
	Decoration
}

// ListboxEvent describes the item before activation. PreventDefault cancels
// selection and the component's OnSelect notification for that activation.
type ListboxEvent struct {
	ID        string
	Selected  bool
	prevented bool
}

func (e *ListboxEvent) PreventDefault() { e.prevented = true }

type ListboxState struct {
	Active          string
	Query           string
	QueryHandle     ui.Handle
	List            ui.ListState
	prefix          string
	typed           time.Time
	horizontal      ui.ScrollState
	scrollTo        string
	selectionAnchor string
	data            listboxData
	selection       stringSelection
	source          listboxItems
	filter          listboxFilter
}

type listboxActivity struct {
	searchChanged  bool
	previousActive string
	submitted      int
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
	Color              Color
	Size               Size
	Highlight          bool
	Loading            bool
	LoadingIcon        *ui.SVG
	SelectedIcon       *ui.SVG
	Filter             bool
	FilterInput        *InputOptions
	FilterFields       []string // label (default), description, value/id, group.
	IgnoreFilter       bool
	SearchTerm         *string // nil uses ListboxState.Query.
	OnSearchTermChange func(string)
	Height             float32
	Orientation        Optional[Orientation] // Defaults to Vertical.
	ItemWidth          float32               // Horizontal item width, default 160 DIP.
	Loop               bool
	HighlightOnHover   Optional[bool]
	SelectionBehavior  SelectionBehavior
	OnChange           func([]string)
	OnSelect           func(string) // Activation, including selecting an unchanged value.
	OnHighlight        func(string) // A user action changed the highlighted item.
	Decoration
	// embedded lists draw no focus ring of their own, as in SelectMenu's popup.
	embedded bool
	// actionID identifies a focusable action row, excluded from selection and Build.
	actionID       string
	activate       func()
	submit         int
	resetHighlight bool
	previousActive string
	heightSet      bool
}

func nextListIndex(items []ListboxItem, at, delta int) int {
	for i := at + delta; i >= 0 && i < len(items); i += delta {
		if !items[i].Disabled && items[i].Kind == ListboxOption {
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
	query := &s.Query
	previousActive := s.Active
	if o.SearchTerm != nil {
		query = o.SearchTerm
	}
	disabled := o.Disabled || t.field.Disabled
	var navigate func(ui.InputEvent) bool
	var searchEditor ui.Element
	highlight := func(id string) {
		if s.Active != id {
			s.Active = id
			if o.OnHighlight != nil {
				o.OnHighlight(id)
			}
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
		if disabled || o.Loading {
			return
		}
		index := s.data.enabledIndex(id)
		if index < 0 {
			return
		}
		if id == o.actionID && o.activate != nil {
			o.activate()
			return
		}
		if options := s.data.items[index].Options; options != nil && options.OnSelect != nil {
			event := &ListboxEvent{ID: id, Selected: slices.Contains(*selected, id)}
			options.OnSelect(event)
			if event.prevented {
				return
			}
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
	if height <= 0 && !o.heightSet {
		height = 224
		if !vert {
			height = 64
		}
	}
	config := defaults(t, ComponentListbox, ComponentTheme{Color: Primary, Size: MD})
	if o.Color != ColorDefault {
		config.Color = o.Color
	}
	if size := inheritedSize(t, o.Size); size != SizeDefault {
		config.Size = size
	}
	if t.field.Error != "" {
		config.Color = Error
	}
	m := optionRecipe(t, config.Size)
	root := ui.Column(c).MinWidth(0).Height(height).Clip().Disabled(o.Disabled || t.field.Disabled)
	disabled = root.IsDisabled()
	activity := ui.Local(root, "activity", func() listboxActivity { return listboxActivity{} })
	rootStyle := PartStyle{Background: Set(t.Background), Radius: Set(t.radius(8)), BorderWidth: Set(float32(1)), BorderColor: Set(t.Border)}
	if o.Highlight || t.field.Error != "" {
		rootStyle.BorderColor = Set(inputColor(t, config.Color))
	}
	mergeStyle(mergeStyle(rootStyle, t.components[ComponentListbox].UI[Root]), o.UI[Root]).apply(root)
	if !o.Filter {
		if s.QueryHandle.Focused(c) && !root.IsDisabled() {
			s.List.Handle.Focus(c)
		}
		s.QueryHandle.CancelFocus(c)
	}
	var list ui.Element
	root.Children(func() {
		listHeight := height
		if o.Filter {
			input := InputOptions{}
			if o.FilterInput != nil {
				input = *o.FilterInput
			}
			if input.Label == "" {
				input.Label = namedMessage(c, "searchNamed", o.Label)
			}
			if input.Placeholder == "" {
				input.Placeholder = message(c, "searchOptions")
			}
			if input.Size == SizeDefault {
				input.Size = config.Size
			}
			if input.Color == ColorDefault {
				input.Color = config.Color
			}
			if input.Variant == VariantDefault {
				input.Variant = None
			}
			if input.Disabled && s.QueryHandle.Focused(c) && !disabled {
				s.List.Handle.Focus(c)
				s.QueryHandle.CancelFocus(c)
			}
			notify, submit, style := input.OnChange, input.OnSubmit, input.Style
			input.OnChange = func(value string) {
				activity.searchChanged = true
				activity.previousActive = previousActive
				if notify != nil {
					notify(value)
				}
				if o.OnSearchTermChange != nil {
					o.OnSearchTermChange(value)
				}
			}
			input.OnSubmit = func(value string) {
				activity.submitted++
				if submit != nil {
					submit(value)
				}
			}
			input.Style = func(frame, editor, leading, trailing ui.Element) {
				searchEditor = editor
				editor.Bind(&s.QueryHandle).OnTextKey(func(k ui.TextKey) bool {
					if c.Resolve(s.QueryHandle).Composing() || k.Mods & ^ui.Shift != 0 {
						return false
					}
					if k.Key == ui.KeyUp || k.Key == ui.KeyDown || k.Key == ui.KeyHome || k.Key == ui.KeyEnd {
						return navigate(ui.InputEvent{Kind: ui.InputKeyDown, Key: k.Key, Mods: k.Mods})
					}
					return false
				})
				finish(t, ComponentListbox, Editor, frame, PartStyle{}, o.Decoration)
				if style != nil {
					style(frame, editor, leading, trailing)
				}
			}
			ui.Column(c.Key("filter")).MinWidth(0).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
				Input(c.Key("input"), t, query, input).FillWidth()
			})
			bm := buttonMetrics(t, input.Size)
			listHeight = max(0, height-bm.Line-2*bm.Y-1)
		}
		if !o.IgnoreFilter && *query != "" {
			s.source.update(items)
			items = s.filter.apply(&s.source, *query, o.FilterFields)
		}
		s.data.update(items)
		items = s.data.items
		rows, positions := s.data.rows, s.data.positions
		external := !slices.Equal(s.selection.values, *selected)
		s.selection.update(*selected)
		at := s.data.enabledIndex(s.Active)
		queryChanged := activity.searchChanged || o.resetHighlight
		if at < 0 || external || queryChanged {
			at = -1
			if !queryChanged {
				for _, id := range *selected {
					if at = s.data.enabledIndex(id); at >= 0 {
						break
					}
				}
			}
			if at < 0 {
				at = nextListIndex(items, -1, 1)
			}
			active := ""
			if at >= 0 {
				active = items[at].ID
			}
			if queryChanged {
				before := o.previousActive
				if activity.searchChanged {
					before = activity.previousActive
				}
				s.Active = active
				if before != active && o.OnHighlight != nil {
					o.OnHighlight(active)
				}
			} else {
				s.Active = active
			}
		}
		s.Query = *query
		activity.searchChanged = false
		if o.Loading {
			listboxLoading(c.Key("loading"), t, config.Size, o)
		} else if len(items) == 0 {
			text := o.EmptyText
			if text == "" {
				key := "noData"
				if *query != "" {
					key = "noMatch"
				}
				text = emptyMessage(c, key, "noResults")
			}
			pad := t.spacing([5]float32{12, 16, 24, 28, 32}[sizeIndex(config.Size)])
			finish(t, ComponentListbox, EmptySlot, ui.Text(c, text).Role(ui.RoleStatus).TextAlign(ui.Center), PartStyle{Foreground: Set(t.Muted), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font), PaddingX: Set(m.Pad), PaddingY: Set(pad)}, o.Decoration)
		}
		itemWidth := o.ItemWidth
		if itemWidth <= 0 {
			itemWidth = 160
		}
		first, last := 0, len(rows)
		count := len(rows)
		if o.Loading {
			count, first, last = 0, 0, 0
		}
		if vert {
			list = ui.List(c.Key("list"), &s.List, count).ItemKey(func(i int) any { return rows[i].key })
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
			last = min(count, int(math.Ceil(float64((s.horizontal.X+viewport)/itemWidth)))+2)
			first = min(first, last)
			list = ui.ScrollHorizontal(c.Key("list")).TrackScroll(&s.horizontal).Bind(&s.List.Handle).Role(ui.RoleList)
		}
		list.Fill().MinHeight(0).Focusable().FocusRing(false).Label(o.Label).InputMethod(false)
		if vert {
			list.Padding(t.spacing(4))
		}
		bindField(c, list, t)
		finish(t, ComponentListbox, List, list, PartStyle{}, o.Decoration)
		move := func(i int) {
			if i >= 0 && i < len(items) && !items[i].Disabled && items[i].Kind == ListboxOption {
				highlight(items[i].ID)
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
		navigate = func(ev ui.InputEvent) bool {
			if o.Loading || disabled {
				return false
			}
			at := s.data.enabledIndex(s.Active)
			if ev.Kind == ui.InputText && ev.Text != "" {
				if ev.Time.Sub(s.typed) > time.Second {
					s.prefix = ""
				}
				s.typed = ev.Time
				s.prefix += strings.ToLower(ev.Text)
				if i := typeaheadIndex(s.prefix, at, len(items), func(i int) (string, bool) {
					return items[i].Label, !items[i].Disabled && items[i].Kind == ListboxOption
				}); i >= 0 {
					move(i)
				}
				return true
			}
			if ev.Kind != ui.InputKeyDown {
				return false
			}
			if o.Multiple && ev.Key == ui.KeyA && (ev.Mods == ui.Ctrl || ev.Mods == ui.Super) {
				all := []string{}
				for _, it := range items {
					if !it.Disabled && it.Kind == ListboxOption && it.ID != o.actionID {
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
						if !items[i].Disabled && items[i].Kind == ListboxOption && items[i].ID != o.actionID {
							rangeIDs = append(rangeIDs, items[i].ID)
						}
					}
					publish(rangeIDs)
				}
			}
			return true
		}
		list.HandleInput(navigate)
		list.OnShortcut(0, ui.KeyEnter, func() {
			if i := s.data.enabledIndex(s.Active); i >= 0 && !o.Loading {
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
			if it.Kind != ListboxOption {
				listboxStructure(r.Context, t, config.Size, it, itemWidth, !vert, o.Decoration)
				return
			}
			chosen := it.ID != o.actionID && s.selection.set[it.ID]
			highlighted := it.ID == s.Active
			// Rows take their content's height; the list measures each row it shows.
			row := ui.Row(r.Context).MinWidth(0).AlignItems(ui.Start).Role(ui.RoleListItem).Label(it.Label).Description(it.Description).Selected(chosen).Disabled(it.Disabled)
			if !vert {
				row.Width(itemWidth).MinWidth(itemWidth).MaxWidth(itemWidth).Shrink(0).Height(listHeight)
			}
			if highlighted {
				list.ActiveDescendant(row)
				if o.Filter {
					searchEditor.ActiveDescendant(row)
				}
			}
			row.HandleInput(func(ev ui.InputEvent) bool {
				if ev.Kind == ui.InputPointerMove && !disabled && !it.Disabled && valueOr(o.HighlightOnHover, true) && s.Active != it.ID {
					highlight(it.ID)
					return true
				}
				return false
			})
			row.OnClick(func() {
				highlight(it.ID)
				choose(it.ID, o.SelectionBehavior != SelectionReplace)
			})
			d := listboxItemDecoration(o.Decoration, it.Options)
			finish(t, ComponentListbox, Item, row, optionItemStyle(t, m, highlighted), d)
			row.Children(func() { listboxItemContent(r.Context, t, it, o, d, config.Size, chosen, highlighted) })
		}
		if vert {
			list.Rows(buildRow)
		} else {
			list.Children(func() {
				ui.Row(c).Width(float32(count) * itemWidth).Height(listHeight).Shrink(0).Children(func() {
					ui.Box(c.Key(listboxSpacer(false))).Width(float32(first) * itemWidth).Shrink(0)
					for i := first; i < last; i++ {
						entry := rows[i]
						buildRow(ui.ListRow{Context: c.Key(entry.key), Index: i})
					}
					ui.Box(c.Key(listboxSpacer(true))).Width(float32(count-last) * itemWidth).Shrink(0)
				})
			})
		}
	})
	if submitted := activity.submitted + o.submit; submitted > 0 {
		activity.submitted = 0
		if !root.IsDisabled() {
			for range submitted {
				choose(s.Active, o.SelectionBehavior != SelectionReplace)
			}
		}
		c.Services().Invalidate()
	}
	// Nuxt rings the list in its color while it has keyboard focus.
	if (list.FocusVisible() || o.Filter && searchEditor.FocusVisible()) && !o.embedded {
		rootStyle.BorderColor = Set(inputColor(t, config.Color))
		root.Shadow(0, 0, 0, 3, inputFocusColor(t, config.Color))
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
		if it.Kind == ListboxLabel {
			h += heading
			continue
		}
		if it.Kind == ListboxSeparator {
			h += 1 + 2*t.spacing(4)
			continue
		}
		body := m.Line
		if it.Description != "" {
			line := m.Line
			if size == XL {
				line = t.text(20)
			}
			if it.Label == "" {
				body = line
			} else {
				body += line
			}
		}
		if it.Icon == nil && it.Avatar != nil && it.Avatar.Size != SizeDefault {
			body = max(body, avatarSize(t, it.Avatar.Size))
		}
		h += body + 2*m.Pad
	}
	return h
}

type SelectMenuState struct {
	Open        bool
	Query       string
	Handle      ui.Handle
	QueryHandle ui.Handle
	Listbox     ListboxState
	clear       ui.Handle
	wasOpen     bool
	labels      map[string]string
	data        listboxItems
	labelIDs    []string
	label       string
	filterCache listboxFilter
}
type SelectMenuOptions struct {
	ListboxOptions
	Color                                          Color
	Variant                                        Variant
	Size                                           Size
	SearchInput                                    Optional[bool]
	FocusSearch                                    Optional[bool]
	ResetSearchTermOnBlur, ResetSearchTermOnSelect Optional[bool]
	ResetModelValueOnClear                         Optional[bool]
	Placeholder                                    string
	Invalid                                        bool
	ErrorMessage                                   string
	SearchPlaceholder                              string
	Clearable, Loading, ExternalFilter             bool
	Leading, Trailing, LoadingIcon, ClearIcon      *ui.SVG
	Avatar                                         *AvatarOptions
	LoadingTrailing                                bool
	ShowTrailing                                   Optional[bool]
	ClearOptions                                   *ButtonOptions
	Error                                          string
	OnQuery                                        func(string)
	OnCreate                                       func(string)
	OnClear                                        func()
	OnOpenChange                                   func(bool)
	FilterFields                                   []string // label (default), description, value/id, group.
	CreateWhenAlways, CreateAtTop                  bool
	Popup                                          PopoverOptions
	selectOptions                                  *SelectOptions
}

func SelectMenu(c *ui.Context, t Theme, value *[]string, s *SelectMenuState, items []ListboxItem, o SelectMenuOptions) ui.Element {
	if s == nil || value == nil {
		panic("nui: SelectMenu requires value and state")
	}
	opened := &s.Open
	previousActive := s.Listbox.Active
	if o.selectOptions != nil && o.selectOptions.Open != nil {
		opened = o.selectOptions.Open
	}
	query := &s.Query
	if o.SearchTerm != nil {
		query = o.SearchTerm
	}
	setOpen := func(next bool) {
		if *opened != next {
			*opened = next
			if o.OnOpenChange != nil {
				o.OnOpenChange(next)
			}
		}
	}
	notifyQuery := func(value string) {
		s.Query = *query
		if o.OnQuery != nil {
			o.OnQuery(value)
		}
		if o.OnSearchTermChange != nil {
			o.OnSearchTermChange(value)
		}
	}
	dataChanged := s.data.update(items)
	items = s.data.items
	resetQuery := func() {
		if *query == "" {
			return
		}
		*query = ""
		notifyQuery("")
	}
	if s.wasOpen && !*opened && valueOr(o.ResetSearchTermOnBlur, true) {
		resetQuery()
	}
	search := valueOr(o.SearchInput, true)
	config := selected(defaults(t, ComponentSelectMenu, ComponentTheme{Color: Primary, Variant: Outline, Size: MD}), ComponentTheme{Color: o.Color, Variant: o.Variant, Size: inheritedSize(t, o.Size)})
	root := ui.Column(c).MinWidth(0)
	activity := ui.Local(root, "activity", func() listboxActivity { return listboxActivity{} })
	if !*opened {
		activity.submitted = 0
	}
	finish(t, ComponentSelectMenu, Root, root, PartStyle{Gap: Set(t.spacing(4))}, o.Decoration)
	root.Children(func() {
		label := s.selectionLabel(*value, dataChanged)
		placeholder := label == ""
		if placeholder {
			label = o.Placeholder
			if label == "" && o.selectOptions == nil {
				label = message(c, "chooseOptions")
			}
		}
		anchor := ui.ButtonBase(c.Key("trigger")).FillWidth().MinWidth(0).FocusRing(false).Disabled(o.Disabled || t.field.Disabled).Bind(&s.Handle)
		name := o.Label
		if name == "" {
			name = label
		}
		anchor.Label(name)
		field := InputOptions{Label: o.Label, Invalid: o.Invalid, ErrorMessage: o.ErrorMessage}
		hover := anchor.Hovered() && !anchor.IsDisabled()
		styles := inputRecipe(t, config, anchor.FocusVisible(), hover, fieldInvalid(t, field))
		style := styles[Root]
		if (o.Highlight || o.selectOptions != nil && o.selectOptions.Highlight) && !fieldInvalid(t, field) {
			style.BorderWidth, style.BorderColor = Set(float32(1)), Set(inputColor(t, config.Color))
		}
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
			anchor.Shadow(0, 0, 0, 3, inputFocusColor(t, color))
		}
		showClear := o.selectOptions == nil && o.Clearable && len(*value) > 0 && !anchor.IsDisabled() && valueOr(o.ShowTrailing, true)
		if !showClear || o.ClearOptions != nil && (o.ClearOptions.Disabled || o.ClearOptions.Loading) {
			if s.clear.Focused(c) && !anchor.IsDisabled() {
				s.Handle.Focus(c)
			}
			s.clear.CancelFocus(c)
		}
		var clearAnchor ui.Element
		anchor.Children(func() {
			triggerOptions := SelectOptions{Leading: o.Leading, Trailing: o.Trailing, Avatar: o.Avatar, Loading: o.Loading, LoadingTrailing: o.LoadingTrailing, LoadingIcon: o.LoadingIcon}
			if o.selectOptions != nil {
				triggerOptions = *o.selectOptions
			}
			triggerOptions.Style = o.Style
			for _, part := range []Slot{Leading, Label, Trailing} {
				styles[part] = mergeStyle(mergeStyle(styles[part], t.components[ComponentSelectMenu].UI[part]), o.UI[part])
			}
			var trailing func() ui.Element
			if !valueOr(o.ShowTrailing, true) {
				trailing = func() ui.Element { return ui.Box(c.Key("trailing")).Hidden(true) }
			} else if showClear {
				trailing = func() ui.Element {
					size := config.Size
					if o.ClearOptions != nil && o.ClearOptions.Size != SizeDefault {
						size = o.ClearOptions.Size
					}
					m := buttonMetrics(t, size)
					clearAnchor = ui.Box(c.Key("trailing")).Size(m.Icon, m.Icon).Shrink(0)
					return clearAnchor
				}
			}
			selectTriggerContentWithTrailing(c, t, label, placeholder, triggerOptions, config.Size, styles, trailing)
		})
		if showClear {
			// A popup button is an accessibility leaf. Keep its clear action a
			// sibling, attached to the trailing slot, so it remains discoverable.
			selectMenuClear(c.Key("clear-control"), t, config.Size, o, &s.clear, func() {
				if valueOr(o.ResetModelValueOnClear, true) && len(*value) > 0 {
					*value = nil
					if o.OnChange != nil {
						o.OnChange(nil)
					}
				}
				resetQuery()
				if *opened {
					if search && valueOr(o.FocusSearch, true) && (o.FilterInput == nil || !o.FilterInput.Disabled) {
						s.QueryHandle.Focus(c)
					} else {
						s.Listbox.List.Handle.Focus(c)
					}
				} else {
					s.Handle.Focus(c)
				}
				if o.OnClear != nil {
					o.OnClear()
				}
			}).AttachTo(clearAnchor, ui.AnchorCenter, ui.AnchorCenter)
		}
		fieldSemantics(c, anchor, t, field)
		anchor.Role(ui.RolePopUpButton).Value(label).Expanded(*opened)
		anchor.OnClick(func() { setOpen(!*opened) }).OnShortcut(0, ui.KeyDown, func() { setOpen(true) }).OnShortcut(0, ui.KeyUp, func() { setOpen(true) })
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
		panel := popover(c.Key("popup"), t, anchor, opened, popup, recipe, func(inner Theme) {
			// The search field and list take the menu's Size and Color.
			inner.field.Size = config.Size
			inner = inner.WithComponent(ComponentListbox, ComponentTheme{Color: config.Color})
			m := optionRecipe(t, config.Size)
			used := float32(0)
			if search {
				input := InputOptions{Placeholder: o.SearchPlaceholder, Variant: None}
				if o.FilterInput != nil {
					input = *o.FilterInput
				}
				if input.Label == "" {
					input.Label = namedMessage(c, "searchNamed", o.Label)
				}
				if input.Placeholder == "" {
					input.Placeholder = o.SearchPlaceholder
					if input.Placeholder == "" {
						input.Placeholder = message(c, "searchOptions")
					}
				}
				if input.Variant == VariantDefault {
					input.Variant = None
				}
				if input.Size == SizeDefault {
					input.Size = config.Size
				}
				if input.Color == ColorDefault {
					input.Color = config.Color
				}
				if input.Disabled && s.QueryHandle.Focused(c) && !anchor.IsDisabled() {
					s.Listbox.List.Handle.Focus(c)
					s.QueryHandle.CancelFocus(c)
				}
				changed, submit, inputStyle := input.OnChange, input.OnSubmit, input.Style
				input.OnChange = func(value string) {
					activity.searchChanged = true
					activity.previousActive = previousActive
					if changed != nil {
						changed(value)
					}
					notifyQuery(value)
				}
				input.OnSubmit = func(value string) {
					activity.submitted++
					if submit != nil {
						submit(value)
					}
				}
				input.Style = func(frame, editor, leading, trailing ui.Element) {
					frame.FillWidth()
					editor.Bind(&s.QueryHandle).OnTextKey(func(k ui.TextKey) bool {
						if !c.Resolve(s.QueryHandle).Composing() && k.Mods == 0 && (k.Key == ui.KeyDown || k.Key == ui.KeyUp) {
							s.Listbox.List.Handle.Focus(c)
							return true
						}
						return false
					})
					finish(t, ComponentSelectMenu, Editor, frame, PartStyle{}, o.Decoration)
					if inputStyle != nil {
						inputStyle(frame, editor, leading, trailing)
					}
				}
				bar := ui.Column(c.Key("search-bar")).MinWidth(0).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(t.Border)
				bar.Children(func() {
					Input(c.Key("search"), inner, query, input)
				})
				bm := buttonMetrics(t, input.Size)
				used += bm.Line + 2*bm.Y + 1
				if !s.wasOpen {
					if valueOr(o.FocusSearch, true) && !input.Disabled {
						s.QueryHandle.Focus(c)
					} else {
						s.Listbox.List.Handle.Focus(c)
					}
				}
			} else {
				if !s.wasOpen || s.QueryHandle.Focused(c) {
					s.Listbox.List.Handle.Focus(c)
				}
				s.QueryHandle.CancelFocus(c)
			}
			filtered := items
			if !o.ExternalFilter && !o.IgnoreFilter && *query != "" {
				filtered = s.filter(*query, o.FilterFields)
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
			opts.Decoration = selectMenuListDecoration(t, o)
			opts.Color, opts.Size = config.Color, config.Size
			opts.Filter, opts.Loading, opts.IgnoreFilter = false, false, true
			opts.Highlight = false
			opts.SearchTerm = nil
			opts.submit, activity.submitted = activity.submitted, 0
			opts.resetHighlight, activity.searchChanged = activity.searchChanged, false
			opts.previousActive = activity.previousActive
			s.Listbox.Query = *query
			if o.selectOptions != nil {
				opts.SelectedIcon = o.selectOptions.SelectedIcon
			}
			opts.Label = namedMessage(c, "optionsNamed", o.Label)
			if opts.EmptyText == "" && len(items) > 0 && len(filtered) == 0 {
				opts.EmptyText = emptyMessage(c, "noMatch", "noResults")
			}
			opts.embedded = true
			if !o.Multiple {
				opts.SelectionBehavior = SelectionReplace
			}
			opts.OnSelect = func(id string) {
				if valueOr(o.ResetSearchTermOnSelect, true) {
					resetQuery()
				}
				if !o.Multiple {
					setOpen(false)
				}
				if o.OnSelect != nil {
					o.OnSelect(id)
				}
			}
			create := o.OnCreate != nil && strings.TrimSpace(*query) != "" && (o.CreateWhenAlways || len(filtered) == 0)
			create = create && !slices.ContainsFunc(filtered, func(it ListboxItem) bool { return it.ID == *query && it.Kind == ListboxOption })
			if create {
				// The create entry is a row of the list, which the arrow keys reach,
				// as in Nuxt; activating it requests creation instead of selecting.
				createID := "\x00nui:create"
				for slices.ContainsFunc(items, func(it ListboxItem) bool { return it.ID == createID }) {
					createID += "_"
				}
				entry := ListboxItem{ID: createID, Label: namedMessage(c, "createItem", *query)}
				if o.CreateAtTop {
					filtered = append([]ListboxItem{entry}, filtered...)
				} else {
					filtered = append(slices.Clone(filtered), entry)
				}
				opts.actionID = createID
				opts.activate = func() {
					o.OnCreate(strings.TrimSpace(*query))
				}
				if o.OnHighlight != nil {
					opts.OnHighlight = func(id string) {
						if id == createID {
							id = *query
						}
						o.OnHighlight(id)
					}
				}
			}
			if opts.Height <= 0 {
				// The list fits its rows within the popup's maximum height.
				opts.Height = max(0, min(listboxContentHeight(t, config.Size, filtered), popup.MaxHeight.Value-used))
				opts.heightSet = true
			}
			Listbox(c.Key("choices"), inner, value, &s.Listbox, filtered, opts).FillWidth()
		})
		if panel.Valid() && panel.Dismissed() && !*opened && o.OnOpenChange != nil {
			o.OnOpenChange(false)
		}
		if s.wasOpen && !*opened && valueOr(o.ResetSearchTermOnBlur, true) {
			resetQuery()
		}
		if o.selectOptions != nil && s.wasOpen && !*opened && !anchor.IsDisabled() {
			anchor.Focus()
		}
		s.Open, s.wasOpen, s.Query = *opened, *opened, *query
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
