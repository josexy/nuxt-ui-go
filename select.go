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
	values      []T
	highlighted Optional[T]
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
	validateOptions(options)
	parts := ui.SelectBase(c, selected)
	e := parts.Trigger.Disabled(o.Disabled || t.field.Disabled).FocusRing(false).MinWidth(0)
	memory := ui.Local(e, "selection-memory", func() selectionMemory[T] { return selectionMemory[T]{} })
	enabled := make([]T, 0, len(options))
	for _, item := range options {
		if !item.Disabled {
			enabled = append(enabled, item.Value)
		}
	}
	changed := !slices.Equal(memory.values, enabled)
	previous := memory.highlighted
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
	for _, item := range options {
		if item.Value == *selected {
			label = item.Label
			found = true
			break
		}
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
	parts.Popup(func(panel ui.Element) {
		// The overlay is outside the trigger's disabled subtree. Propagate its
		// effective state so disabling an open select also disables its options.
		panel.Disabled(e.IsDisabled())
		selectionPanel(c, panel, e, t, o.Popup)
		finish(t, ComponentSelect, Panel, panel, PartStyle{}, o.Decoration)
		ui.Scroll(c).MaxHeight(selectionHeight(o.Popup)).MinWidth(0).Children(func() {
			if len(options) == 0 {
				finish(t, ComponentSelect, EmptySlot, ui.Text(c, message(c, "noOptions")).Role(ui.RoleStatus).Padding(8), PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
			}
			for _, option := range options {
				var item ui.Element
				if option.Disabled {
					item = ui.Row(c.Key(optionKey[T]{option.Value})).Disabled(true)
				} else {
					c.Key(optionKey[T]{option.Value})
					item = parts.Item(option.Value)
				}
				item.Role(ui.RoleListItem).Label(option.Label).Description(option.Description)
				optionStyle(item, t, ComponentSelect, o.Decoration)
				if item.Highlighted() && !option.Disabled {
					memory.highlighted = Set(option.Value)
				}
				item.Children(func() { optionContent(c, t, option, option.Value == *selected, config.Size) })
			}
		})
	})
	if parts.Open() && changed && len(enabled) > 0 {
		target := enabled[0]
		if previous.Set && slices.Contains(enabled, previous.Value) {
			target = previous.Value
		} else if slices.Contains(enabled, *selected) {
			target = *selected
		}
		parts.Highlight(target)
		memory.highlighted = Set(target)
	}
	memory.values = append(memory.values[:0], enabled...)
	applyGroup(e, t, styles[Root].Radius.Value)
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
type menuIdentity struct {
	ID, Label, Description string
	Disabled               bool
}
type menuMemory struct {
	identities  []menuIdentity
	initialized bool
	focused     bool
	open        bool
	keepFilter  bool
}

func InputMenu(c *ui.Context, t Theme, state *SearchState, options []Option[string], o InputMenuOptions) ui.Element {
	if o.Mode > InputMenuAutocomplete {
		panic("nui: invalid input menu mode")
	}
	if o.Mode == InputMenuAutocomplete {
		o.Multiple = false
		o.AllowCustom = true
	}
	if state == nil {
		panic("nui: InputMenu requires state")
	}
	validateOptions(options)
	for _, item := range options {
		if item.Value == "" {
			panic("nui: InputMenu IDs cannot be empty")
		}
	}
	input := o.InputOptions
	if o.Multiple {
		style := input.Style
		input.Style = func(root, editor, leading, trailing ui.Element) {
			editor.OnTextKey(func(k ui.TextKey) bool {
				if !o.ReadOnly && k.Mods == 0 && k.Key == ui.KeyBackspace && k.Text() == "" && len(state.SelectedValues) > 0 {
					state.SelectedValues = slices.Clone(state.SelectedValues[:len(state.SelectedValues)-1])
					if o.OnValuesChange != nil {
						o.OnValuesChange(slices.Clone(state.SelectedValues))
					}
					return true
				}
				return false
			})
			if style != nil {
				style(root, editor, leading, trailing)
			}
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
	var parts ui.ComboboxParts
	var memory *menuMemory
	createID := "\x00nui:create"
	for slices.ContainsFunc(options, func(it Option[string]) bool { return it.Value == createID }) {
		createID += "_"
	}
	root := inputShell(c, t, &state.Query, input, ComponentInputMenu, 0, 0, func() ui.Element {
		parts = ui.ComboboxBase(c.Key("input"), &state.Query)
		return parts.Input
	}, func(anchor ui.Element) {
		if o.Multiple {
			anchor.Wrap()
			seen := map[string]bool{}
			for _, id := range state.SelectedValues {
				if seen[id] {
					continue
				}
				seen[id] = true
				label := id
				for _, it := range options {
					if it.Value == id {
						label = it.Label
						break
					}
				}
				Button(c.Key("selected:"+id), t, label, ButtonOptions{AccessibleLabel: namedMessage(c, "removeNamed", label), Size: XS, Variant: Subtle, Color: Neutral, Trailing: lucide.X(), Disabled: o.ReadOnly}).OnClick(func() {
					state.SelectedValues = slices.DeleteFunc(slices.Clone(state.SelectedValues), func(v string) bool { return v == id })
					if o.OnValuesChange != nil {
						o.OnValuesChange(slices.Clone(state.SelectedValues))
					}
				})
			}
		}
		if input.Disabled || input.ReadOnly || t.field.Disabled {
			parts.SetOpen(false)
		}
		memory = ui.Local(parts.Input, "menu-identity", func() menuMemory { return menuMemory{} })
		focused := parts.Input.Focused()
		if parts.Input.Clicked() && !valueOr(o.OpenOnClick, true) {
			parts.SetOpen(false)
		}
		if focused && !memory.focused && valueOr(o.OpenOnFocus, false) && !input.Disabled && !input.ReadOnly && !t.field.Disabled {
			parts.SetOpen(true)
		}
		identities := make([]menuIdentity, len(options))
		for i, item := range options {
			identities[i] = menuIdentity{item.Value, item.Label, item.Description, item.Disabled}
		}
		// Base highlights by index. A changed data set dismisses before accepting another
		// choice, so a highlighted position cannot silently become a different identity.
		if memory.initialized && !slices.Equal(memory.identities, identities) {
			parts.SetOpen(false)
		}
		memory.identities = identities
		memory.initialized = true
		query := strings.ToLower(strings.TrimSpace(state.Query))
		visible := make([]Option[string], 0, len(options))
		enabled := 0
		for _, option := range options {
			if o.ExternalFilter || (!parts.Filtering() && !memory.keepFilter) || matchOption(option, query, o.FilterFields) {
				visible = append(visible, option)
				if !option.Disabled {
					enabled++
				}
			}
		}
		config := inputDefaults(t, ComponentInputMenu, input)
		create := o.OnCreate != nil && query != "" && (o.CreateWhenAlways || len(visible) == 0)
		if create && slices.ContainsFunc(visible, func(it Option[string]) bool { return it.Value == state.Query }) {
			create = false
		}
		build := func(panel ui.Element, empty bool) {
			selectionPanel(c, panel, anchor, t, o.Popup)
			finish(t, ComponentInputMenu, Panel, panel, PartStyle{}, Decoration{UI: input.UI})
			ui.Scroll(c).MaxHeight(selectionHeight(o.Popup)).MinWidth(0).Children(func() {
				buildCreate := func() {
					if create {
						parts.Item(createID).Padding(8).Label(namedMessage(c, "createItem", state.Query)).Children(func() { ui.Text(c, namedMessage(c, "createItem", state.Query)) })
					}
				}
				if o.CreateAtTop {
					buildCreate()
				}
				for _, option := range visible {
					var item ui.Element
					if option.Disabled {
						item = ui.Row(c.Key(optionKey[string]{option.Value})).Disabled(true)
					} else {
						c.Key(optionKey[string]{option.Value})
						item = parts.Item(option.Value)
					}
					item.Role(ui.RoleListItem).Label(option.Label).Description(option.Description)
					optionStyle(item, t, ComponentInputMenu, Decoration{UI: input.UI})
					item.Children(func() {
						chosen := option.Value == state.Selected
						if o.Multiple {
							chosen = slices.Contains(state.SelectedValues, option.Value)
						}
						optionContent(c, t, option, chosen, config.Size)
					})
				}
				if !o.CreateAtTop {
					buildCreate()
				}
				if empty && !create {
					text := o.EmptyText
					if text == "" {
						text = message(c, "noResultsFound")
					}
					finish(t, ComponentInputMenu, EmptySlot, ui.Text(c, text).Padding(8).Role(ui.RoleStatus).Label(text), PartStyle{Foreground: Set(t.Muted)}, Decoration{UI: input.UI})
				}
			})
		}
		parts.Popup(func(panel ui.Element) {
			if enabled > 0 || create {
				build(panel, false)
			}
		})
		if enabled == 0 && !create && parts.Open() {
			ui.Overlay(c, func() {
				// Display-only fallback: no fake selectable item and no focus-stealing surface.
				panel := ui.Column(c).IgnoreInput(true).Role(ui.RoleStatus)
				panel.Children(func() { build(panel, true) })
			})
		}
	})
	if id, chosen := parts.Chosen(); chosen {
		if id == createID {
			query := strings.TrimSpace(state.Query)
			if o.OnCreate != nil {
				root.SetChanged().OnChange(func() { o.OnCreate(query) })
			}
			parts.SetOpen(true)
			memory.open, memory.keepFilter = true, true
			memory.focused = parts.Input.Focused()
			return root
		}
		for _, option := range options {
			if option.Value == id && !option.Disabled {
				state.Selected = id
				if o.Multiple {
					if slices.Contains(state.SelectedValues, id) {
						state.SelectedValues = slices.DeleteFunc(slices.Clone(state.SelectedValues), func(v string) bool { return v == id })
					} else {
						state.SelectedValues = append(slices.Clone(state.SelectedValues), id)
					}
					if valueOr(o.ResetSearchTermOnSelect, true) {
						state.Query = ""
					}
					parts.SetOpen(true)
				} else if valueOr(o.ResetSearchTermOnSelect, true) {
					state.Query = option.Label
				}
				if o.OnSelect != nil {
					root.SetChanged().OnChange(func() { o.OnSelect(id) })
				}
				if o.Multiple && o.OnValuesChange != nil {
					root.SetChanged().OnChange(func() { o.OnValuesChange(slices.Clone(state.SelectedValues)) })
				}
				break
			}
		}
	}
	if !parts.Open() {
		memory.keepFilter = false
	}
	if ((memory.open && !parts.Open()) || (memory.focused && !parts.Input.Focused())) && valueOr(o.ResetSearchTermOnBlur, true) && !o.AllowCustom {
		state.Query = ""
		if !o.Multiple {
			for _, it := range options {
				if it.Value == state.Selected {
					state.Query = it.Label
					break
				}
			}
		}
	}
	memory.open, memory.focused = parts.Open(), parts.Input.Focused()
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
