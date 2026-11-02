package nui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// Option identity is independent of its label. Build is an optional synchronous content builder.
type Option[T comparable] struct {
	Value              T
	Label, Description string
	Disabled           bool
	Icon               *ui.SVG
	Build              func(*ui.Context, Theme)
}
type SelectOptions struct {
	Label, Placeholder, ErrorMessage string
	Disabled, Invalid                bool
	Color                            Color
	Variant                          Variant
	Size                             Size
	Popup                            PopoverOptions
	OnChange                         func()
	Decoration
}
type selectionMemory[T comparable] struct {
	list    optionListState[T]
	open    bool
	pending Optional[T]
}
type optionKey[T comparable] struct{ Value T }

func validateOptions[T comparable](options []Option[T]) {
	seen := make(map[T]bool, len(options))
	for _, o := range options {
		if o.Value != o.Value || seen[o.Value] {
			panic("nui: selection values must be unique and reflexive")
		}
		seen[o.Value] = true
	}
}
func optionContent[T comparable](c *ui.Context, t Theme, o Option[T], chosen bool, size Size) {
	m := buttonMetrics(t, size)
	if o.Build != nil {
		o.Build(c, t)
	} else {
		if o.Icon != nil {
			ui.Icon(c, o.Icon).Size(m.Icon, m.Icon).TextColor(t.Muted).Shrink(0)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(t.spacing(2)).Children(func() {
			ui.Text(c, o.Label).SingleLine().TextColor(t.Highlighted)
			if o.Description != "" {
				ui.Text(c, o.Description).SingleLine().FontSize(t.text(12)).TextColor(t.Muted)
			}
		})
	}
	if chosen {
		ui.Icon(c, lucide.Check()).Size(t.text(16), t.text(16)).TextColor(t.Primary).Shrink(0)
	}
}
func optionStyle(e ui.Element, t Theme, id Component, d Decoration) {
	s := PartStyle{PaddingX: Set(t.spacing(8)), PaddingY: Set(t.spacing(6)), Gap: Set(t.spacing(8)), Radius: Set(t.radius(4)), Foreground: Set(t.Text)}
	if e.Highlighted() && !e.IsDisabled() {
		s.Background = Set(t.Elevated)
	}
	e.MinWidth(0).AlignItems(ui.Center)
	finish(t, id, Item, e, s, d)
}

// Select never reserves a T zero value. Optional[T] or an explicit option models absence.
func Select[T comparable](c *ui.Context, t Theme, selected *T, options []Option[T], o SelectOptions) ui.Element {
	if selected == nil {
		panic("nui: Select requires a value pointer")
	}
	e := ui.ButtonBase(c).Role(ui.RolePopUpButton).Disabled(o.Disabled || t.field.Disabled).FocusRing(false).MinWidth(0)
	memory := ui.Local(e, "selection-memory", func() selectionMemory[T] { return selectionMemory[T]{} })
	changed := memory.list.update(options)
	if memory.pending.Set {
		v := memory.pending.Value
		memory.pending = Optional[T]{}
		if _, ok := memory.list.enabled(v); ok && !e.IsDisabled() && *selected != v {
			*selected = v
			e.SetChanged()
		}
	}
	open := func() {
		memory.open = true
		memory.list.active = Optional[T]{}
		memory.list.highlight(*selected, true)
		if !memory.list.active.Set {
			memory.list.move(ui.KeyHome)
		}
	}
	choose := func(value T) { memory.pending, memory.open = Set(value), false }
	e.OnClick(func() {
		if memory.open {
			memory.open = false
		} else {
			open()
		}
	})
	for _, key := range []ui.Key{ui.KeyDown, ui.KeyUp, ui.KeyHome, ui.KeyEnd} {
		if !memory.open && (key == ui.KeyHome || key == ui.KeyEnd) {
			continue
		}
		e.OnShortcut(0, key, func() {
			if !memory.open {
				open()
			} else {
				memory.list.move(key)
			}
		})
	}
	for _, key := range []ui.Key{ui.KeyEnter, ui.KeySpace} {
		e.OnShortcut(0, key, func() {
			if !memory.open {
				open()
			} else if memory.list.active.Set {
				choose(memory.list.active.Value)
			}
		})
	}
	if memory.open && changed {
		if _, ok := memory.list.enabled(memory.list.active.Value); !memory.list.active.Set || !ok {
			memory.list.active = Optional[T]{}
			memory.list.highlight(*selected, true)
			if !memory.list.active.Set {
				memory.list.move(ui.KeyHome)
			}
		}
	}
	e.Expanded(memory.open)
	field := InputOptions{Color: o.Color, Variant: o.Variant, Size: o.Size, Invalid: o.Invalid, ErrorMessage: o.ErrorMessage, Label: o.Label}
	config, styles := resolve(t, ComponentSelect, ComponentTheme{Color: Primary, Size: MD, Variant: Outline}, ComponentTheme{Color: o.Color, Variant: o.Variant, Size: inheritedSize(t, o.Size), UI: o.UI}, func(v ComponentTheme) Styles {
		hover := e.Hovered() && !e.IsDisabled()
		styles := inputRecipe(t, v, e.FocusVisible(), hover, fieldInvalid(t, field))
		if hover {
			root := styles[Root]
			switch v.Variant {
			case Outline:
				root.Background = Set(t.Elevated)
			case Subtle:
				root.Background = Set(t.Accented.Alpha(.75))
			}
			styles[Root] = root
		}
		return styles
	})
	styles[Root].apply(e)
	fieldSemantics(c, e, t, field)
	label := o.Placeholder
	found := false
	if i, ok := memory.list.index[*selected]; ok {
		label, found = options[i].Label, true
	}
	e.Value(label).Children(func() {
		value := ui.Text(c, label).SingleLine().Grow(1).MinWidth(0)
		if !found {
			value.TextColor(t.Dimmed)
		}
		decorated(Label, value, styles, o.Style)
		decorated(Trailing, ui.Icon(c, lucide.ChevronDown()).Size(buttonMetrics(t, config.Size).Icon, buttonMetrics(t, config.Size).Icon).TextColor(t.Dimmed).Shrink(0), styles, o.Style)
	})
	if e.FocusVisible() && config.Variant != None {
		color := config.Color
		if fieldInvalid(t, field) {
			color = Error
		}
		e.Shadow(0, 0, 0, 3, t.focusColor(color))
	}
	if o.OnChange != nil {
		e.OnChange(o.OnChange)
	}
	if memory.open {
		// Select dismisses and consumes an outside press, like the native menu.
		ui.Overlay(c, func() {
			ui.Box(c.Key("select-backdrop")).Absolute().Left(0).Top(0).Right(0).Bottom(0).KeepFocus().OnClick(func() { memory.open = false })
		})
	}
	ui.PopoverBase(c.Key(memory), e, &memory.open, func(panel ui.Element) {
		// The overlay is outside the trigger's disabled subtree. Propagate its
		// effective state so disabling an open select also disables its options.
		panel.Disabled(e.IsDisabled()).KeepFocus()
		selectionPanel(c, panel, e, t, o.Popup)
		finish(t, ComponentSelect, Panel, panel, PartStyle{}, o.Decoration)
		if len(options) == 0 {
			finish(t, ComponentSelect, EmptySlot, ui.Text(c, message(c, "noOptions")).Role(ui.RoleStatus).Padding(8), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
		}
		optionList(c, t, &memory.list, e, len(options), func(i int) Option[T] { return options[i] }, selectionHeight(o.Popup), ComponentSelect, o.Decoration, config.Size, func(v T) bool { return v == *selected }, choose)
	})
	applyGroup(c, e, t, styles[Root].Radius.Value)
	if o.Style != nil {
		o.Style(Root, e)
	}
	return e
}
func selectionHeight(o PopoverOptions) float32 {
	if o.MaxHeight.Set {
		return max(0, o.MaxHeight.Value-8)
	}
	return 232
}

// SearchState separates a draft query from its last committed stable ID. Empty Selected means no choice.
type SearchState struct {
	Query, Selected string
	SelectedValues  []string // Used when InputMenuOptions.Multiple is true.
}
type InputMenuMode uint8

const (
	InputMenuCombobox InputMenuMode = iota
	InputMenuAutocomplete
)

type InputMenuOptions struct {
	InputOptions
	EmptyText                                      string
	Popup                                          PopoverOptions
	AllowCustom                                    bool
	Multiple, ExternalFilter                       bool
	FilterFields                                   []string // label, description, value; nil uses label.
	Mode                                           InputMenuMode
	OpenOnClick, OpenOnFocus                       Optional[bool]
	ResetSearchTermOnBlur, ResetSearchTermOnSelect Optional[bool]
	OnValuesChange                                 func([]string)
	OnCreate                                       func(string)
	CreateWhenAlways                               bool
	CreateAtTop                                    bool
	OnSelect                                       func(string)
}
type menuMemory struct {
	data         optionListState[string]
	list         optionListState[string]
	selection    stringSelection
	filtered     []int
	filterQuery  string
	filterFields []string
	filterKnown  bool
	initialized  bool
	focused      bool
	open         bool
	wasOpen      bool
	filtering    bool
	keepFilter   bool
	pending      Optional[string]
	openingKey   ui.Key
}

func InputMenu(c *ui.Context, t Theme, state *SearchState, options []Option[string], o InputMenuOptions) ui.Element {
	if o.Mode > InputMenuAutocomplete {
		panic("nui: invalid input menu mode")
	}
	if o.Mode == InputMenuAutocomplete {
		o.Multiple, o.AllowCustom = false, true
	}
	if state == nil {
		panic("nui: InputMenu requires state")
	}
	input := o.InputOptions
	var editor ui.Element
	var memory *menuMemory
	style := input.Style
	input.Style = func(root, in, leading, trailing ui.Element) {
		in.OnTextKey(func(k ui.TextKey) bool {
			if o.ReadOnly || k.Mods != 0 {
				return false
			}
			switch k.Key {
			case ui.KeyDown, ui.KeyUp:
				if !memory.open {
					memory.open, memory.filtering = true, false
					memory.openingKey = ui.KeyHome
					if k.Key == ui.KeyUp {
						memory.openingKey = ui.KeyEnd
					}
				} else {
					memory.list.move(k.Key)
				}
				return true
			case ui.KeyEnter:
				if memory.open && memory.list.active.Set {
					memory.pending = memory.list.active
					return true
				}
			case ui.KeyBackspace:
				if o.Multiple && k.Text() == "" && len(state.SelectedValues) > 0 {
					state.SelectedValues = slices.Clone(state.SelectedValues[:len(state.SelectedValues)-1])
					if o.OnValuesChange != nil {
						o.OnValuesChange(slices.Clone(state.SelectedValues))
					}
					return true
				}
			}
			return false
		})
		if style != nil {
			style(root, in, leading, trailing)
		}
	}
	input.OnClear = func() {
		state.Selected = ""
		if o.Multiple {
			state.SelectedValues = nil
			if o.OnValuesChange != nil {
				o.OnValuesChange(nil)
			}
		}
		if o.OnClear != nil {
			o.OnClear()
		}
	}
	if !o.AllowCustom {
		input.OnSubmit = nil
	}
	createID := "\x00nui:create"
	root := inputShell(c, t, &state.Query, input, ComponentInputMenu, 0, 0, func() ui.Element {
		editor = ui.TextInputBase(c.Key("input"), &state.Query).Role(ui.RoleComboBox)
		memory = ui.Local(editor, "menu-identity", func() menuMemory { return menuMemory{} })
		changed := memory.data.update(options)
		if changed {
			memory.filterKnown = false
		}
		if _, exists := memory.data.index[""]; exists {
			panic("nui: InputMenu IDs cannot be empty")
		}
		// A changed data set dismisses before accepting another choice.
		if changed && memory.initialized {
			memory.open, memory.pending = false, Optional[string]{}
		}
		memory.initialized = true
		for {
			if _, exists := memory.data.index[createID]; !exists {
				break
			}
			createID += "_"
		}
		return editor
	}, func(anchor ui.Element) {
		memory.selection.update(state.SelectedValues)
		if o.Multiple {
			anchor.Wrap()
			seen := map[string]bool{}
			for _, id := range state.SelectedValues {
				if seen[id] {
					continue
				}
				seen[id] = true
				label := id
				if i, ok := memory.data.index[id]; ok {
					label = options[i].Label
				}
				Button(c.Key("selected:"+id), t, label, ButtonOptions{AccessibleLabel: namedMessage(c, "removeNamed", label), Size: XS, Variant: Subtle, Color: Neutral, Trailing: lucide.X(), Disabled: o.ReadOnly}).OnClick(func() {
					state.SelectedValues = slices.DeleteFunc(slices.Clone(state.SelectedValues), func(v string) bool { return v == id })
					if o.OnValuesChange != nil {
						o.OnValuesChange(slices.Clone(state.SelectedValues))
					}
				})
			}
		}
		focused := editor.Focused()
		if editor.Changed() {
			memory.open, memory.filtering = true, true
			memory.openingKey = ui.KeyHome
		}
		editor.OnClick(func() {
			if valueOr(o.OpenOnClick, true) && !o.ReadOnly {
				memory.open, memory.filtering = true, false
			}
		})
		if focused && !memory.focused && valueOr(o.OpenOnFocus, false) {
			memory.open, memory.filtering = true, false
		}
		if !focused || editor.IsDisabled() || input.ReadOnly {
			memory.open = false
		}
		query := strings.ToLower(strings.TrimSpace(state.Query))
		filter := !o.ExternalFilter && (memory.filtering || memory.keepFilter)
		if filter && (!memory.filterKnown || memory.filterQuery != query ||
			(memory.filterFields == nil) != (o.FilterFields == nil) || !slices.Equal(memory.filterFields, o.FilterFields)) {
			memory.filtered = memory.filtered[:0]
			for i, option := range options {
				if matchOption(option, query, o.FilterFields) {
					memory.filtered = append(memory.filtered, i)
				}
			}
			memory.filterQuery, memory.filterFields, memory.filterKnown = query, slices.Clone(o.FilterFields), true
		}
		count := len(options)
		if filter {
			count = len(memory.filtered)
		}
		create := o.OnCreate != nil && query != "" && (o.CreateWhenAlways || count == 0)
		if at, exists := memory.data.index[state.Query]; create && exists && (!filter || slices.Contains(memory.filtered, at)) {
			create = false
		}
		var entry Option[string]
		if create {
			entry = Option[string]{Value: createID, Label: namedMessage(c, "createItem", state.Query)}
			count++
		}
		optionAt := func(i int) Option[string] {
			if create {
				if (o.CreateAtTop && i == 0) || (!o.CreateAtTop && i == count-1) {
					return entry
				}
				if o.CreateAtTop {
					i--
				}
			}
			if filter {
				i = memory.filtered[i]
			}
			return options[i]
		}
		if memory.list.updateRows(count, optionAt) {
			if _, ok := memory.list.enabled(memory.list.active.Value); !memory.list.active.Set || !ok {
				memory.list.active = Optional[string]{}
				if memory.filtering {
					memory.list.move(ui.KeyHome)
				}
			}
		}
		if memory.openingKey != 0 {
			memory.list.move(memory.openingKey)
			memory.openingKey = 0
		}
		editor.Expanded(memory.open)
		config := inputDefaults(t, ComponentInputMenu, input)
		ui.PopoverBase(c.Key(memory), anchor, &memory.open, func(panel ui.Element) {
			panel.KeepFocus().Disabled(editor.IsDisabled() || input.ReadOnly).Role(ui.RoleList)
			selectionPanel(c, panel, anchor, t, o.Popup)
			finish(t, ComponentInputMenu, Panel, panel, PartStyle{}, Decoration{UI: input.UI})
			ui.Column(c).MinWidth(0).Children(func() {
				optionList(c, t, &memory.list, editor, count, optionAt, selectionHeight(o.Popup), ComponentInputMenu, Decoration{UI: input.UI}, config.Size, func(id string) bool {
					if o.Multiple {
						return memory.selection.set[id]
					}
					return id == state.Selected
				}, func(id string) { memory.pending = Set(id) })
				if memory.list.enabledCount == 0 && !create {
					text := o.EmptyText
					if text == "" {
						text = message(c, "noResultsFound")
					}
					finish(t, ComponentInputMenu, EmptySlot, ui.Text(c, text).Padding(8).Role(ui.RoleStatus), PartStyle{Foreground: Set(t.Muted)}, Decoration{UI: input.UI})
				}
			})
		})
	})
	if memory.pending.Set {
		id := memory.pending.Value
		memory.pending = Optional[string]{}
		c.Invalidate()
		if id == createID && o.OnCreate != nil {
			query := strings.TrimSpace(state.Query)
			root.SetChanged().OnChange(func() { o.OnCreate(query) })
			memory.open, memory.keepFilter = true, true
		} else if i, ok := memory.data.enabled(id); ok {
			state.Selected = id
			memory.open, memory.filtering = o.Multiple, false
			if o.Multiple {
				if slices.Contains(state.SelectedValues, id) {
					state.SelectedValues = slices.DeleteFunc(slices.Clone(state.SelectedValues), func(v string) bool { return v == id })
				} else {
					state.SelectedValues = append(slices.Clone(state.SelectedValues), id)
				}
				if valueOr(o.ResetSearchTermOnSelect, true) {
					state.Query = ""
				}
			} else if valueOr(o.ResetSearchTermOnSelect, true) {
				state.Query = options[i].Label
			}
			if o.OnSelect != nil {
				root.SetChanged().OnChange(func() { o.OnSelect(id) })
			}
			if o.Multiple && o.OnValuesChange != nil {
				root.SetChanged().OnChange(func() { o.OnValuesChange(slices.Clone(state.SelectedValues)) })
			}
		}
	}
	if !memory.open {
		memory.keepFilter, memory.filtering = false, false
		memory.list.active = Optional[string]{}
	}
	if ((memory.wasOpen && !memory.open) || (memory.focused && !editor.Focused())) && valueOr(o.ResetSearchTermOnBlur, true) && !o.AllowCustom {
		before := state.Query
		state.Query = ""
		if !o.Multiple {
			if i, ok := memory.data.index[state.Selected]; ok {
				state.Query = options[i].Label
			}
		}
		if before != state.Query {
			c.Invalidate()
		}
	}
	memory.wasOpen, memory.focused = memory.open, editor.Focused()
	return root
}
func matchOption[T comparable](option Option[T], query string, fields []string) bool {
	if fields == nil {
		fields = []string{"label"}
	}
	for _, field := range fields {
		var text string
		switch field {
		case "label":
			text = option.Label
		case "description":
			text = option.Description
		case "value":
			text = fmt.Sprint(option.Value)
		default:
			panic("nui: unknown option filter field")
		}
		if strings.Contains(strings.ToLower(text), query) {
			return true
		}
	}
	return false
}

// SearchSelect is a convenience wrapper; use InputMenu for additional options.
func SearchSelect(c *ui.Context, t Theme, state *SearchState, options []Option[string], label string) ui.Element {
	return InputMenu(c, t, state, options, InputMenuOptions{Label: label, Placeholder: message(c, "searchWorkspaces"), FilterFields: []string{"label", "description"}})
}
