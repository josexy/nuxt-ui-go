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
	Active            string
	Query             string
	QueryHandle       ui.Handle
	List              ui.ListState
	prefix            string
	typed             time.Time
	horizontal        ui.ScrollState
	scrollTo          string
	selectionAnchor   string
	data              listboxData
	selection         *SelectionState[string]
	selectionRevision uint64
	preferred         string
	source            listboxItems
	filter            listboxFilter
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
	SearchExecution    SearchExecution
	IgnoreFilter       bool
	SearchTerm         *string // nil uses ListboxState.Query.
	OnSearchTermChange func(string)
	Height             float32
	Orientation        Optional[Orientation] // Defaults to Vertical.
	ItemWidth          float32               // Horizontal item width, default 160 DIP.
	Loop               bool
	HighlightOnHover   Optional[bool]
	SelectionBehavior  SelectionBehavior
	OnChange           func()
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

func nextListIndex(items *listboxItems, at, delta int) int {
	for i := at + delta; i >= 0 && i < items.count(); i += delta {
		if !items.item(i).Disabled && items.item(i).Kind == ListboxOption {
			return i
		}
	}
	return at
}
func Listbox(c *ui.Context, t Theme, selected *SelectionState[string], s *ListboxState, o ListboxOptions) ui.Element {
	if s == nil {
		panic("nui: Listbox requires state")
	}
	return listbox(c, t, selected, s, &s.source, o)
}

func listbox(c *ui.Context, t Theme, selected *SelectionState[string], s *ListboxState, source *listboxItems, o ListboxOptions) ui.Element {
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
	notifySelection := func(changed bool) {
		if changed {
			s.observeSelection(selected, false)
			if o.OnChange != nil {
				o.OnChange()
			}
		}
	}
	publish := func(next []string) { notifySelection(selected.Set(next)) }
	choose := func(id string, toggle bool) {
		if disabled || o.Loading || !s.filter.search.ready(*query) {
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
		if options := s.data.item(index).Options; options != nil && options.OnSelect != nil {
			event := &ListboxEvent{ID: id, Selected: selected.Has(id)}
			options.OnSelect(event)
			if event.prevented {
				return
			}
		}
		if toggle && (o.Multiple || selected.Has(id)) {
			if selected.Has(id) {
				notifySelection(selected.Remove(id))
			} else {
				notifySelection(selected.Add(id))
			}
		} else {
			s.selectionAnchor = id
			publish([]string{id})
		}
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
	s.filter.search.bind(c, root, s.Close)
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
					// Input methods' filters withhold these keys; the guard covers a backend that lets one through.
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
		visible := source
		if !o.IgnoreFilter && *query != "" {
			visible = s.filter.visible(source, *query, o.FilterFields, o.SearchExecution)
		} else {
			s.filter.search.pause()
		}
		loading := o.Loading || s.filter.search.pending
		dataChanged := s.data.bind(visible)
		items := &s.data.listboxItems
		rows, positions := s.data.rows, s.data.positions
		external := s.observeSelection(selected, dataChanged)
		at := s.data.enabledIndex(s.Active)
		queryChanged := activity.searchChanged || o.resetHighlight
		if at < 0 || external || queryChanged {
			at = -1
			if !queryChanged {
				at = s.data.enabledIndex(s.preferred)
			}
			if at < 0 && len(s.data.enabled) > 0 {
				at = s.data.enabled[0]
			}
			active := ""
			if at >= 0 {
				active = items.item(at).ID
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
		if dataChanged && at >= 0 {
			if vert {
				s.List.ScrollIntoView(positions[at])
			} else {
				s.scrollTo = s.Active
			}
		}
		s.Query = *query
		activity.searchChanged = false
		if loading {
			listboxLoading(c, t, config.Size, o)
		} else if items.count() == 0 {
			text := o.EmptyText
			if text == "" {
				key := "noData"
				if *query != "" {
					key = "noMatch"
				}
				text = emptyMessage(c, key, "noResults")
			}
			pad := t.spacing([5]float32{12, 16, 24, 28, 32}[sizeIndex(config.Size)])
			finish(t, ComponentListbox, EmptySlot, statusText(c, "empty", text).TextAlign(ui.Center), PartStyle{Foreground: Set(t.Muted), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font), PaddingX: Set(m.Pad), PaddingY: Set(pad)}, o.Decoration)
		}
		itemWidth := o.ItemWidth
		if itemWidth <= 0 {
			itemWidth = 160
		}
		first, last := 0, len(rows)
		count := len(rows)
		if loading {
			count, first, last = 0, 0, 0
		}
		if vert {
			list = ui.List(c.Key("list"), &s.List, count).ItemKey(func(i int) any { return rows[i].key })
		} else {
			// Before its first layout, or while an ancestor is Hidden, the
			// viewport is unknown: build the window's width of rows, and let
			// the first paint with a real box ask for one more build. A hidden
			// list paints nothing, so it never asks.
			viewport := s.horizontal.Bounds(c).W
			guessed := viewport <= 0
			if guessed {
				viewport, _ = c.Size()
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
			// The window's width is at least the real viewport's, so rows for a
			// guess still cover what shows; only a measured one may clamp the
			// offset kept while hidden.
			x := max(0, min(s.horizontal.X, float32(len(rows))*itemWidth-viewport))
			if !guessed {
				s.horizontal.X = x
			}
			first = max(0, int(x/itemWidth)-2)
			last = min(count, int(math.Ceil(float64((x+viewport)/itemWidth)))+2)
			first = min(first, last)
			list = ui.ScrollHorizontal(c.Key("list")).TrackScroll(&s.horizontal).Bind(&s.List.Handle).Role(ui.RoleList)
			if guessed {
				services := c.Services()
				list.Draw(func(_ *ui.Painter, r ui.Rect) {
					if guessed && r.W > 0 && r.H > 0 {
						guessed = false // once per build: Draw may run more than once
						services.Invalidate()
					}
				})
			}
		}
		list.Fill().MinHeight(0).Focusable().FocusRing(false).Label(o.Label).InputMethod(false)
		if vert {
			list.Padding(t.spacing(4))
		}
		bindField(c, list, t)
		finish(t, ComponentListbox, List, list, PartStyle{}, o.Decoration)
		move := func(i int) {
			if i >= 0 && i < items.count() && !items.item(i).Disabled && items.item(i).Kind == ListboxOption {
				highlight(items.item(i).ID)
				if vert {
					s.List.ScrollIntoView(positions[i])
				} else {
					s.scrollTo = items.item(i).ID
				}
			}
		}
		next := func(at, delta int) int {
			i := nextListIndex(items, at, delta)
			if i == at && o.Loop {
				if delta > 0 {
					i = nextListIndex(items, -1, 1)
				} else {
					i = nextListIndex(items, items.count(), -1)
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
			if loading || disabled {
				return false
			}
			at := s.data.enabledIndex(s.Active)
			if ev.Kind == ui.InputText && ev.Text != "" {
				if ev.Time.Sub(s.typed) > time.Second {
					s.prefix = ""
				}
				s.typed = ev.Time
				s.prefix += strings.ToLower(ev.Text)
				if i := typeaheadIndex(s.prefix, at, items.count(), func(i int) (string, bool) {
					return items.item(i).Label, !items.item(i).Disabled && items.item(i).Kind == ListboxOption
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
				for i := range items.count() {
					it := items.item(i)
					if !it.Disabled && it.Kind == ListboxOption && it.ID != o.actionID {
						all = append(all, it.ID)
					}
				}
				publish(all)
				move(nextListIndex(items, items.count(), -1))
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
				move(nextListIndex(items, items.count(), -1))
			case ui.KeySpace:
				if at >= 0 && ev.Mods == 0 {
					choose(items.item(at).ID, o.SelectionBehavior != SelectionReplace)
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
						if !items.item(i).Disabled && items.item(i).Kind == ListboxOption && items.item(i).ID != o.actionID {
							rangeIDs = append(rangeIDs, items.item(i).ID)
						}
					}
					publish(rangeIDs)
				}
			}
			return true
		}
		list.HandleInput(navigate)
		list.OnShortcut(0, ui.KeyEnter, func() {
			if i := s.data.enabledIndex(s.Active); i >= 0 && !loading {
				choose(items.item(i).ID, o.SelectionBehavior != SelectionReplace)
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
			it := items.item(entry.item)
			if it.Kind != ListboxOption {
				listboxStructure(r.Context, t, config.Size, it, itemWidth, !vert, o.Decoration)
				return
			}
			chosen := it.ID != o.actionID && selected.Has(it.ID)
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
			t.fadeColors(r.Context, row)
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
func listboxContentHeight(t Theme, size Size, items *listboxItems, limit float32) float32 {
	m := optionRecipe(t, size)
	label := optionLabelStyle(t, size)
	heading := 2*m.Pad + label.FontSize.Value*label.LineHeight.Value
	if items.count() == 0 {
		return 2*t.spacing([5]float32{12, 16, 24, 28, 32}[sizeIndex(size)]) + m.Line
	}
	h, group := 2*t.spacing(4), ""
	for i := range items.count() {
		it := items.item(i)
		if h >= limit {
			return limit
		}
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
	Open                                      bool
	Query                                     string
	Handle                                    ui.Handle
	QueryHandle                               ui.Handle
	Listbox                                   ListboxState
	clear                                     ui.Handle
	wasOpen                                   bool
	labels                                    map[string]string
	data                                      listboxItems
	labelSelection                            *SelectionState[string]
	labelSelectionRevision, labelDataRevision uint64
	creation                                  listboxCreation
	label                                     string
	filterCache                               listboxFilter
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
}

func SelectMenu(c *ui.Context, t Theme, value *SelectionState[string], s *SelectMenuState, o SelectMenuOptions) ui.Element {
	if s == nil || value == nil {
		panic("nui: SelectMenu requires value and state")
	}
	opened := &s.Open
	previousActive := s.Listbox.Active
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
	items := &s.data
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
	s.filterCache.search.bind(c, root, s.Close)
	if !*opened {
		s.filterCache.search.pause()
	}
	activity := ui.Local(root, "activity", func() listboxActivity { return listboxActivity{} })
	if !*opened {
		activity.submitted = 0
	}
	finish(t, ComponentSelectMenu, Root, root, PartStyle{Gap: Set(t.spacing(4))}, o.Decoration)
	root.Children(func() {
		label := s.selectionLabel(value)
		placeholder := label == ""
		if placeholder {
			label = o.Placeholder
			if label == "" {
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
		if o.Highlight && !fieldInvalid(t, field) {
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
		t.fadeColors(c, anchor)
		if anchor.FocusVisible() && config.Variant != None {
			color := config.Color
			if fieldInvalid(t, field) {
				color = Error
			}
			anchor.Shadow(0, 0, 0, 3, inputFocusColor(t, color))
		}
		showClear := o.Clearable && value.Len() > 0 && !anchor.IsDisabled() && valueOr(o.ShowTrailing, true)
		if !showClear || o.ClearOptions != nil && (o.ClearOptions.Disabled || o.ClearOptions.Loading) {
			if s.clear.Focused(c) && !anchor.IsDisabled() {
				s.Handle.Focus(c)
			}
			s.clear.CancelFocus(c)
		}
		var clearAnchor ui.Element
		anchor.Children(func() {
			triggerOptions := SelectOptions{Leading: o.Leading, Trailing: o.Trailing, Avatar: o.Avatar, Loading: o.Loading, LoadingTrailing: o.LoadingTrailing, LoadingIcon: o.LoadingIcon}
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
				if valueOr(o.ResetModelValueOnClear, true) && value.Clear() {
					if o.OnChange != nil {
						o.OnChange()
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
						// Input methods' filters withhold these keys; the guard covers a backend that lets one through.
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
			visible := &s.data
			if !o.ExternalFilter && !o.IgnoreFilter && *query != "" {
				visible = s.filterCache.visible(&s.data, *query, o.FilterFields, o.SearchExecution)
			} else {
				s.filterCache.search.pause()
			}
			filtered := visible
			status := func(key, text string, color ui.Color) {
				e := statusText(c, key, text).TextColor(color).Shrink(0)
				optionItemStyle(t, m, false).apply(e)
				e.TextColor(color)
				used += m.Line + 2*m.Pad
			}
			if o.Loading && !s.filterCache.search.pending {
				status("status-loading", message(c, "loading"), t.Muted)
			}
			if o.Error != "" {
				status("status-error", o.Error, t.Error)
			}
			opts := o.ListboxOptions
			opts.Decoration = selectMenuListDecoration(t, o)
			opts.Color, opts.Size = config.Color, config.Size
			opts.Filter, opts.Loading, opts.IgnoreFilter = false, s.filterCache.search.pending, true
			opts.Highlight = false
			opts.SearchTerm = nil
			opts.submit, activity.submitted = activity.submitted, 0
			opts.resetHighlight, activity.searchChanged = activity.searchChanged, false
			opts.previousActive = activity.previousActive
			s.Listbox.Query = *query
			opts.Label = namedMessage(c, "optionsNamed", o.Label)
			if opts.EmptyText == "" && items.count() > 0 && filtered.count() == 0 {
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
			create := !s.filterCache.search.pending && o.OnCreate != nil && strings.TrimSpace(*query) != "" && (o.CreateWhenAlways || filtered.count() == 0)
			if at, exists := visible.index[*query]; exists && visible.item(at).Kind == ListboxOption {
				create = false
			}
			if create {
				// The create entry is a row of the list, which the arrow keys reach,
				// as in Nuxt; activating it requests creation instead of selecting.
				visible = s.creation.prepare(&s.data, visible, namedMessage(c, "createItem", *query), o.CreateAtTop)
				filtered = visible
				createID := s.creation.id
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
				opts.Height = max(0, min(listboxContentHeight(t, config.Size, filtered, popup.MaxHeight.Value-used), popup.MaxHeight.Value-used))
				opts.heightSet = true
			}
			listbox(c.Key("choices"), inner, value, &s.Listbox, visible, opts).FillWidth()
		})
		if panel.Valid() && panel.Dismissed() && !*opened && o.OnOpenChange != nil {
			o.OnOpenChange(false)
		}
		if s.wasOpen && !*opened && valueOr(o.ResetSearchTermOnBlur, true) {
			resetQuery()
		}
		if !*opened {
			s.filterCache.search.pause()
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

// Searching reports whether local matching is pending. Application-owned
// remote loading remains the caller's responsibility.
func (s *ListboxState) Searching() bool { return s.filter.search.pending }

// Close cancels local search and outstanding focus requests, retaining the
// submitted data, query and app-owned selection. A later mounted view can search
// again. Owners may call Close during disposal; unmount also cancels work.
func (s *ListboxState) Close() {
	s.filter.reset()
	s.scrollTo, s.prefix = "", ""
	s.QueryHandle.CancelFocus()
	s.List.CancelFocus()
}

func (s *SelectMenuState) Searching() bool { return s.filterCache.search.pending }

// Close dismisses the popup and cancels local search/focus requests without
// changing submitted data, the draft or selection.
func (s *SelectMenuState) Close() {
	s.Open, s.wasOpen = false, false
	s.filterCache.reset()
	s.Listbox.Close()
	s.Handle.CancelFocus()
	s.QueryHandle.CancelFocus()
}
