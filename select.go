package nui

import (
	"fmt"
	"slices"
	"strings"
	"time"

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
	Leading, Trailing                *ui.SVG
	Avatar                           *AvatarOptions
	Loading, LoadingTrailing         bool
	LoadingIcon, SelectedIcon        *ui.SVG
	Highlight                        bool
	Open                             *bool // Optional persistent state; nil keeps opening internal.
	Popup                            PopoverOptions
	OnChange                         func()
	Decoration
}
type selectionMemory[T comparable] struct {
	list    optionListState[T]
	open    bool
	wasOpen bool
	prefix  string
	typed   time.Time
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

// optionMetrics is Nuxt's item recipe shared by Select, InputMenu, SelectMenu,
// Listbox, menus and CommandPalette: p-1…p-2 with a matching gap, text-xs…
// text-base, and size-4…size-6 icons, which are as tall as the text line.
type optionMetrics struct{ Pad, Gap, Font, Line, Icon float32 }

func optionRecipe(t Theme, size Size) optionMetrics {
	i := sizeIndex(size)
	return optionMetrics{
		Pad:  t.spacing([5]float32{4, 6, 6, 8, 8}[i]),
		Gap:  t.spacing([5]float32{4, 6, 6, 8, 8}[i]),
		Font: t.text([5]float32{12, 12, 14, 14, 16}[i]),
		Line: t.text([5]float32{16, 16, 20, 20, 24}[i]),
		Icon: t.text([5]float32{16, 16, 20, 20, 24}[i]),
	}
}

// optionItemStyle is the row: rounded-md; a highlighted row has an
// elevated/50 surface and highlighted text, which its icons follow.
// The surface is always set: MyGo's Selected fills a selected row with the
// accent color, where Nuxt shows a selected row only by its trailing check.
func optionItemStyle(t Theme, m optionMetrics, highlighted bool) PartStyle {
	s := PartStyle{Background: Set(ui.Transparent), PaddingX: Set(m.Pad), PaddingY: Set(m.Pad), Gap: Set(m.Gap), Radius: Set(t.radius(6)), Foreground: Set(t.Text), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font)}
	if highlighted {
		s.Background = Set(t.Elevated.Alpha(.5))
		s.Foreground = Set(t.Highlighted)
	}
	return s
}

// optionLeadingColor dims a leading icon until its row is highlighted.
func optionLeadingColor(t Theme, highlighted bool) ui.Color {
	if highlighted {
		return t.Text
	}
	return t.Dimmed
}

// optionLabelStyle is a group heading: semibold highlighted text, one step smaller.
func optionLabelStyle(t Theme, size Size) PartStyle {
	m := optionRecipe(t, size)
	font := t.text([5]float32{10, 10, 12, 12, 14}[sizeIndex(size)])
	return PartStyle{PaddingX: Set(m.Pad), PaddingY: Set(m.Pad), FontSize: Set(font), LineHeight: Set(float32(12.0 / 10)), Weight: Set(600), Foreground: Set(t.Highlighted)}
}

// optionTexts are a row's label and muted description, at the row's size.
func optionTexts(c *ui.Context, t Theme, label, description string) {
	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		ui.Text(c, label).SingleLine().MinWidth(0)
		if description != "" {
			ui.Text(c, description).SingleLine().MinWidth(0).TextColor(t.Muted)
		}
	})
}

// optionCheck marks a selected row at its end, in the row's text color.
func optionCheck(c *ui.Context, m optionMetrics, icon *ui.SVG) {
	if icon == nil {
		icon = lucide.Check()
	}
	ui.Icon(c.Key("selected"), icon).Size(m.Icon, m.Icon).Shrink(0)
}

func optionContent[T comparable](c *ui.Context, t Theme, o Option[T], chosen, highlighted bool, size Size, selectedIcon *ui.SVG) {
	m := optionRecipe(t, size)
	if o.Build != nil {
		o.Build(c, t)
	} else {
		if o.Icon != nil {
			ui.Icon(c, o.Icon).Size(m.Icon, m.Icon).TextColor(optionLeadingColor(t, highlighted)).Shrink(0)
		}
		optionTexts(c, t, o.Label, o.Description)
	}
	if chosen {
		optionCheck(c, m, selectedIcon)
	}
}

// tagChip is Nuxt's tag: an elevated chip with an accented inset ring, sized
// with its field, whose x alone removes it. Pressing the x keeps the focus
// in the editor; it stays a keyboard stop, as MyGo has no arrow keys between tags.
func tagChip(c *ui.Context, t Theme, label, removeLabel string, size Size, disabled bool, deleteIcon *ui.SVG, remove func()) (chip, del ui.Element) {
	return tagChipStyled(c, t, label, removeLabel, size, disabled, deleteIcon, remove, nil)
}

func tagChipStyled(c *ui.Context, t Theme, label, removeLabel string, size Size, disabled bool, deleteIcon *ui.SVG, remove func(), decorate func(Slot, ui.Element)) (chip, del ui.Element) {
	i := sizeIndex(size)
	font := t.text([5]float32{10, 10, 12, 12, 14}[i])
	line := t.text([5]float32{12, 12, 16, 16, 20}[i])
	icon := t.text([5]float32{12, 12, 14, 14, 16}[i])
	chip = ui.Row(c).MinWidth(0).AlignItems(ui.Center).Gap(t.spacing(2)).PaddingX(t.spacing(6)).PaddingY(t.spacing(2)).Radius(t.radius(4)).
		Background(t.Elevated).TextColor(t.Text).FontSize(font).LineHeight(line/font).FontWeight(500).InsetShadow(0, 0, 0, 1, t.BorderAccented)
	chip.Children(func() {
		text := ui.Text(c, label).MinWidth(0)
		if decorate == nil {
			text.SingleLine()
		} else {
			decorate(Label, text)
		}
		del = ui.ButtonBase(c.Key("remove")).Label(removeLabel).Disabled(disabled).KeepFocus().FocusRing(true).Radius(t.radius(2)).Shrink(0).OnClick(remove)
		color := t.Dimmed
		if del.Hovered() && !del.IsDisabled() {
			color = t.Text
			del.Background(t.Accented.Alpha(.75))
		}
		if deleteIcon == nil {
			deleteIcon = lucide.X()
		}
		del.Children(func() {
			part := ui.Icon(c, deleteIcon).Size(icon, icon).TextColor(color)
			if decorate != nil {
				decorate(Indicator, part)
			}
		})
		if decorate != nil {
			decorate(Close, del)
		}
	})
	if decorate != nil {
		decorate(Item, chip)
	}
	return chip, del
}

// optionEmptyStyle is a popup list's centered, muted empty state.
func optionEmptyStyle(t Theme, size Size) PartStyle {
	m := optionRecipe(t, size)
	pad := t.spacing([5]float32{8, 10, 10, 12, 12}[sizeIndex(size)])
	return PartStyle{Foreground: Set(t.Muted), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font), PaddingX: Set(pad), PaddingY: Set(pad)}
}
func optionStyle(e ui.Element, t Theme, id Component, d Decoration, size Size, highlighted bool) {
	e.MinWidth(0).AlignItems(ui.Center)
	if id == ComponentSelect || id == ComponentInputMenu {
		e.AlignItems(ui.Start)
	}
	finish(t, id, Item, e, optionItemStyle(t, optionRecipe(t, size), highlighted), d)
}

// Select never reserves a T zero value. Optional[T] or an explicit option models absence.
func Select[T comparable](c *ui.Context, t Theme, selected *T, options []Option[T], o SelectOptions) ui.Element {
	if selected == nil {
		panic("nui: Select requires a value pointer")
	}
	e := ui.ButtonBase(c).Role(ui.RolePopUpButton).Disabled(o.Disabled || t.field.Disabled).FocusRing(false).MinWidth(0)
	memory := ui.Local(e, "selection-memory", func() selectionMemory[T] { return selectionMemory[T]{} })
	opened := &memory.open
	if o.Open != nil {
		opened = o.Open
	}
	changed := memory.list.update(options)
	if memory.pending.Set {
		v := memory.pending.Value
		memory.pending = Optional[T]{}
		if _, ok := memory.list.enabled(v); ok && !e.IsDisabled() && *selected != v {
			*selected = v
			e.SetChanged()
		}
	}
	initialize := func() {
		memory.prefix = ""
		memory.list.active = Optional[T]{}
		memory.list.highlight(*selected, true)
		if !memory.list.active.Set {
			memory.list.move(ui.KeyHome)
		}
	}
	if *opened && !memory.wasOpen {
		initialize()
	}
	open := func() {
		*opened = true
		initialize()
		memory.wasOpen = true
	}
	choose := func(value T) { memory.pending, *opened = Set(value), false }
	e.OnClick(func() {
		if *opened {
			*opened = false
		} else {
			open()
		}
	})
	for _, key := range []ui.Key{ui.KeyDown, ui.KeyUp, ui.KeyHome, ui.KeyEnd} {
		if !*opened && (key == ui.KeyHome || key == ui.KeyEnd) {
			continue
		}
		e.OnShortcut(0, key, func() {
			if !*opened {
				open()
			} else {
				memory.list.move(key)
			}
		})
	}
	for _, key := range []ui.Key{ui.KeyEnter, ui.KeySpace} {
		e.OnShortcut(0, key, func() {
			if !*opened {
				open()
			} else if memory.list.active.Set {
				choose(memory.list.active.Value)
			}
		})
	}
	disabled := e.IsDisabled()
	e.HandleInput(func(ev ui.InputEvent) bool {
		if disabled || ev.Kind != ui.InputText || ev.Text == "" || ev.Mods&(ui.Ctrl|ui.Alt|ui.Super) != 0 {
			return false
		}
		if ev.Time.Sub(memory.typed) > time.Second {
			memory.prefix = ""
		}
		memory.typed, memory.prefix = ev.Time, memory.prefix+strings.ToLower(ev.Text)
		current := *selected
		if *opened && memory.list.active.Set {
			current = memory.list.active.Value
		}
		at, ok := memory.list.index[current]
		if !ok {
			at = -1
		}
		matched := typeaheadIndex(memory.prefix, at, len(options), func(i int) (string, bool) {
			return options[i].Label, !options[i].Disabled
		})
		if matched >= 0 {
			if *opened {
				memory.list.highlight(options[matched].Value, true)
			} else {
				memory.pending = Set(options[matched].Value)
			}
		}
		return true
	})
	if *opened && changed {
		if _, ok := memory.list.enabled(memory.list.active.Value); !memory.list.active.Set || !ok {
			memory.list.active = Optional[T]{}
			memory.list.highlight(*selected, true)
			if !memory.list.active.Set {
				memory.list.move(ui.KeyHome)
			}
		}
	}
	e.Expanded(*opened)
	field := InputOptions{Color: o.Color, Variant: o.Variant, Size: o.Size, Invalid: o.Invalid, ErrorMessage: o.ErrorMessage, Label: o.Label}
	config, styles := resolve(t, ComponentSelect, ComponentTheme{Color: Primary, Size: MD, Variant: Outline}, ComponentTheme{Color: o.Color, Variant: o.Variant, Size: inheritedSize(t, o.Size), UI: o.UI}, func(v ComponentTheme) Styles {
		hover := e.Hovered() && !e.IsDisabled()
		styles := inputRecipe(t, v, e.FocusVisible(), hover, fieldInvalid(t, field))
		if o.Highlight && !fieldInvalid(t, field) {
			root := styles[Root]
			root.BorderWidth, root.BorderColor = Set(float32(1)), Set(inputColor(t, v.Color))
			styles[Root] = root
		}
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
		selectTriggerContent(c, t, label, !found, o, config.Size, styles)
	})
	if e.FocusVisible() && config.Variant != None {
		color := config.Color
		if fieldInvalid(t, field) {
			color = Error
		}
		e.Shadow(0, 0, 0, 3, inputFocusColor(t, color))
	}
	if o.OnChange != nil {
		e.OnChange(o.OnChange)
	}
	if *opened {
		// Select dismisses and consumes an outside press, like the native menu.
		ui.Overlay(c, func() {
			ui.Box(c.Key("select-backdrop")).Absolute().Left(0).Top(0).Right(0).Bottom(0).KeepFocus().OnClick(func() {
				if valueOr(o.Popup.Dismissible, true) {
					*opened = false
				}
			})
		})
	}
	panel := ui.PopoverBase(c.Key(memory), e, opened, func(panel ui.Element) {
		// The overlay is outside the trigger's disabled subtree. Propagate its
		// effective state so disabling an open select also disables its options.
		panel.Disabled(e.IsDisabled()).KeepFocus()
		selectionPanel(c, panel, e, t, o.Popup)
		finish(t, ComponentSelect, Panel, panel, PartStyle{}, o.Decoration)
		if len(options) == 0 {
			finish(t, ComponentSelect, EmptySlot, ui.Text(c, emptyMessage(c, "noData", "noOptions")).Role(ui.RoleStatus).TextAlign(ui.Center), optionEmptyStyle(t, config.Size), o.Decoration)
		}
		optionList(c, t, &memory.list, e, len(options), func(i int) Option[T] { return options[i] }, selectionHeight(o.Popup), ComponentSelect, o.Decoration, config.Size, o.SelectedIcon, true, func(v T) bool { return v == *selected }, choose)
	})
	if panel.Valid() && panel.Dismissed() && !valueOr(o.Popup.Dismissible, true) {
		*opened = true
	}
	if memory.wasOpen && !*opened && !disabled {
		e.Focus()
	}
	memory.open, memory.wasOpen = *opened, *opened
	applyGroup(c, e, t, styles[Root].Radius.Value)
	if o.Style != nil {
		o.Style(Root, e)
	}
	return e
}

// selectTriggerContent is shared by the typed single and multiple adapters.
// Parts are created with stable keys so changing adornments preserves the trigger.
func selectTriggerContent(c *ui.Context, t Theme, label string, placeholder bool, o SelectOptions, size Size, styles Styles) {
	selectTriggerContentWithTrailing(c, t, label, placeholder, o, size, styles, nil)
}

func selectTriggerContentWithTrailing(c *ui.Context, t Theme, label string, placeholder bool, o SelectOptions, size Size, styles Styles, customTrailing func() ui.Element) {
	m := buttonMetrics(t, size)
	loadingLeading := o.Loading && (!o.LoadingTrailing || o.Leading != nil)
	var leading, trailing, loader ui.Element
	if loadingLeading {
		leading = ui.Box(c.Key("leading")).Size(m.Icon, m.Icon).Shrink(0)
		loader = leading
	} else if o.Leading != nil {
		leading = ui.Icon(c.Key("leading"), o.Leading).Size(m.Icon, m.Icon).Shrink(0)
	} else if o.Avatar != nil {
		diameter := Optional[float32]{}
		if o.Avatar.Size == SizeDefault {
			diameter = Set(m.Icon)
		}
		leading = avatar(c.Key("leading"), t, *o.Avatar, diameter)
	}
	decorated(Leading, leading, styles, o.Style)
	value := ui.Text(c.Key("value"), label).SingleLine().Grow(1).MinWidth(0)
	if placeholder {
		value.TextColor(t.Dimmed)
	}
	decorated(Label, value, styles, o.Style)
	if customTrailing != nil {
		trailing = customTrailing()
	} else if o.Loading && !loadingLeading {
		trailing = ui.Box(c.Key("trailing")).Size(m.Icon, m.Icon).Shrink(0)
		loader = trailing
	} else {
		icon := o.Trailing
		if icon == nil {
			icon = lucide.ChevronDown()
		}
		trailing = ui.Icon(c.Key("trailing"), icon).Size(m.Icon, m.Icon).Shrink(0)
	}
	decorated(Trailing, trailing, styles, o.Style)
	if loader.Valid() {
		icon := o.LoadingIcon
		if icon == nil {
			icon = lucide.LoaderCircle()
		}
		part := Trailing
		if loadingLeading {
			part = Leading
		}
		drawLoadingSVG(loader, icon, mergeValue(styles[Root].Foreground, styles[part].Foreground).Value, t.motion(c), t.Motion.Spin)
	}
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
	ResetModelValueOnClear                         Optional[bool]
	HighlightOnHover                               Optional[bool]
	Loop                                           bool
	Open                                           *bool
	SelectedIcon, DeleteIcon, ClearIcon            *ui.SVG
	OnValuesChange                                 func([]string)
	OnCreate                                       func(string)
	CreateWhenAlways                               bool
	CreateAtTop                                    bool
	OnSelect                                       func(string)
}
type menuMemory struct {
	data           optionListState[string]
	list           optionListState[string]
	selection      stringSelection
	filtered       []int
	filterQuery    string
	filterFields   []string
	filterKnown    bool
	initialized    bool
	focused        bool
	open           bool
	wasOpen        bool
	filtering      bool
	keepFilter     bool
	pending        Optional[string]
	openingKey     ui.Key
	editor         ui.Handle
	focusRequested bool
	lastSelected   string
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
	input.clearWhenEmpty = state.Selected != ""
	if o.Multiple {
		input.clearWhenEmpty = len(state.SelectedValues) > 0
	}
	input.clearIcon = o.ClearIcon
	var editor ui.Element
	var memory *menuMemory
	var opened *bool
	style := input.Style
	input.Style = func(root, in, leading, trailing ui.Element) {
		in.OnTextKey(func(k ui.TextKey) bool {
			if o.ReadOnly || k.Mods != 0 {
				return false
			}
			switch k.Key {
			case ui.KeyDown, ui.KeyUp:
				if !*opened {
					*opened, memory.filtering = true, false
					memory.openingKey = ui.KeyHome
					if k.Key == ui.KeyUp {
						memory.openingKey = ui.KeyEnd
					}
				} else {
					memory.list.moveWrapped(k.Key, o.Loop)
				}
				return true
			case ui.KeyEnter:
				if *opened && memory.list.active.Set {
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
		// Keep the toggle in the normal Trailing styling and callback pipeline.
		trailing.Role(ui.RoleButton).Label(message(c, "showOptions")).KeepFocus().Cursor(ui.CursorDefault).
			Hidden(o.Clearable && (state.Query != "" || input.clearWhenEmpty) && !o.ReadOnly).
			OnClick(func() {
				if !o.ReadOnly {
					*opened, memory.filtering = !*opened, false
					memory.focusRequested = true
				}
			})
		if style != nil {
			style(root, in, leading, trailing)
		}
	}
	input.OnClear = func() {
		if o.Mode == InputMenuAutocomplete || valueOr(o.ResetModelValueOnClear, true) {
			state.Selected = ""
			if o.Multiple && len(state.SelectedValues) > 0 {
				state.SelectedValues = nil
				if o.OnValuesChange != nil {
					o.OnValuesChange(nil)
				}
			}
		}
		if o.OnClear != nil {
			o.OnClear()
		}
	}
	if !o.AllowCustom {
		input.OnSubmit = nil
	}
	if input.Trailing == nil {
		input.Trailing = lucide.ChevronDown()
	}
	createID := "\x00nui:create"
	root := inputShell(c, t, &state.Query, input, ComponentInputMenu, 0, 0, func() ui.Element {
		// Tags and the draft share the available space, while shell icons stay
		// at the field edges even when the selected tags wrap onto more lines.
		content := ui.Row(c.Key("content")).Grow(1).MinWidth(0).Basis(0).AlignItems(ui.Center).Gap(t.spacing(4))
		if o.Multiple {
			content.Wrap()
		}
		memory = ui.Local(content, "menu-identity", func() menuMemory { return menuMemory{} })
		opened = &memory.open
		if o.Open != nil {
			opened = o.Open
		}
		if !o.Multiple && o.Mode != InputMenuAutocomplete && valueOr(o.ResetSearchTermOnSelect, true) && !memory.editor.Focused(c) && ((!memory.initialized && state.Query == "") || (memory.initialized && state.Selected != memory.lastSelected)) {
			state.Query = ""
			for _, option := range options {
				if option.Value == state.Selected {
					state.Query = option.Label
					break
				}
			}
		}
		changed := memory.data.update(options)
		if changed {
			memory.filterKnown = false
		}
		if _, exists := memory.data.index[""]; exists {
			panic("nui: InputMenu IDs cannot be empty")
		}
		// A changed data set dismisses before accepting another choice.
		if changed && memory.initialized {
			*opened, memory.pending = false, Optional[string]{}
		}
		memory.initialized = true
		for {
			if _, exists := memory.data.index[createID]; !exists {
				break
			}
			createID += "_"
		}
		memory.selection.update(state.SelectedValues)
		size := inputDefaults(t, ComponentInputMenu, input).Size
		content.Children(func() {
			if o.Multiple {
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
					c.Key("selected:" + id)
					tagChip(c, t, label, namedMessage(c, "removeNamed", label), size, o.ReadOnly, o.DeleteIcon, func() {
						state.SelectedValues = slices.DeleteFunc(slices.Clone(state.SelectedValues), func(v string) bool { return v == id })
						memory.focusRequested = true
						if o.OnValuesChange != nil {
							o.OnValuesChange(slices.Clone(state.SelectedValues))
						}
					})
				}
			}
			editor = ui.TextInputBase(c.Key("input"), &state.Query).Role(ui.RoleComboBox).Bind(&memory.editor)
			if o.Multiple {
				// Leave room for a caret before moving the draft to the next line.
				editor.Basis(t.text(40))
			}
		})
		return editor
	}, func(anchor ui.Element) {
		if memory.focusRequested || *opened && (!memory.wasOpen || !valueOr(o.Popup.Dismissible, true)) {
			editor.Focus()
			memory.focusRequested = false
		}
		focused := editor.Focused()
		if editor.Changed() {
			*opened, memory.filtering = true, true
			memory.openingKey = ui.KeyHome
		}
		editor.OnClick(func() {
			if valueOr(o.OpenOnClick, false) && !o.ReadOnly {
				*opened, memory.filtering = true, false
			}
		})
		if focused && !memory.focused && valueOr(o.OpenOnFocus, false) {
			*opened, memory.filtering = true, false
		}
		if !focused || editor.IsDisabled() || input.ReadOnly {
			*opened = false
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
		if *opened && !memory.wasOpen && !memory.list.active.Set {
			memory.list.highlight(state.Selected, true)
			if !memory.list.active.Set {
				memory.list.move(ui.KeyHome)
			}
		}
		editor.Expanded(*opened)
		config := inputDefaults(t, ComponentInputMenu, input)
		panel := ui.PopoverBase(c.Key(memory), anchor, opened, func(panel ui.Element) {
			panel.KeepFocus().Disabled(editor.IsDisabled() || input.ReadOnly).Role(ui.RoleList)
			selectionPanel(c, panel, anchor, t, o.Popup)
			finish(t, ComponentInputMenu, Panel, panel, PartStyle{}, Decoration{UI: input.UI})
			ui.Column(c).MinWidth(0).Children(func() {
				optionList(c, t, &memory.list, editor, count, optionAt, selectionHeight(o.Popup), ComponentInputMenu, Decoration{UI: input.UI}, config.Size, o.SelectedIcon, valueOr(o.HighlightOnHover, true), func(id string) bool {
					if o.Mode == InputMenuAutocomplete {
						return false
					}
					if o.Multiple {
						return memory.selection.set[id]
					}
					return id == state.Selected
				}, func(id string) { memory.pending = Set(id) })
				if count == 0 && !create {
					text := o.EmptyText
					if text == "" {
						key := "noData"
						if len(options) > 0 && strings.TrimSpace(state.Query) != "" {
							key = "noMatch"
						}
						text = emptyMessage(c, key, "noResultsFound")
					}
					finish(t, ComponentInputMenu, EmptySlot, ui.Text(c, text).Role(ui.RoleStatus).TextAlign(ui.Center), optionEmptyStyle(t, config.Size), Decoration{UI: input.UI})
				}
			})
		})
		if panel.Valid() && panel.Dismissed() && !valueOr(o.Popup.Dismissible, true) && !editor.IsDisabled() && !o.ReadOnly {
			*opened = true
		}
	})
	if memory.pending.Set {
		id := memory.pending.Value
		memory.pending = Optional[string]{}
		c.Invalidate()
		if editor.IsDisabled() || o.ReadOnly {
			*opened = false
		} else if id == createID && o.OnCreate != nil {
			query := strings.TrimSpace(state.Query)
			root.SetChanged().OnChange(func() { o.OnCreate(query) })
			*opened, memory.keepFilter = true, true
		} else if i, ok := memory.data.enabled(id); ok {
			state.Selected = id
			*opened, memory.filtering = o.Multiple, false
			if o.Multiple {
				if slices.Contains(state.SelectedValues, id) {
					state.SelectedValues = slices.DeleteFunc(slices.Clone(state.SelectedValues), func(v string) bool { return v == id })
				} else {
					state.SelectedValues = append(slices.Clone(state.SelectedValues), id)
				}
				if valueOr(o.ResetSearchTermOnSelect, true) {
					state.Query = ""
				}
			} else if o.Mode == InputMenuAutocomplete || valueOr(o.ResetSearchTermOnSelect, true) {
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
	if !*opened {
		memory.keepFilter, memory.filtering = false, false
		memory.list.active = Optional[string]{}
	}
	if ((memory.wasOpen && !*opened) || (memory.focused && !editor.Focused())) && valueOr(o.ResetSearchTermOnBlur, true) && !o.AllowCustom {
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
	memory.wasOpen, memory.focused = *opened, editor.Focused()
	memory.open, memory.lastSelected = *opened, state.Selected
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
	return InputMenu(c, t, state, options, InputMenuOptions{Label: label, Placeholder: message(c, "searchWorkspaces"), FilterFields: []string{"label", "description"}, OpenOnClick: Set(true)})
}
