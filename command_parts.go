package nui

import (
	"maps"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func activateCommand(s *CommandPaletteState, a CommandAction, o CommandPaletteOptions) bool {
	version, index := s.Results.version, s.index
	query, parent, open := s.Query, s.parent(), s.Open
	event := CommandEvent{ID: a.ID}
	if a.OnSelectEvent != nil {
		a.OnSelectEvent(&event)
	} else if o.OnSelectEvent != nil {
		o.OnSelectEvent(&event)
	}
	if event.prevented || s.Results.version != version || s.index != index || s.Query != query || s.parent() != parent || s.Open != open {
		return false
	}
	if valueOr(o.CloseOnSelect, true) {
		s.Close()
	}
	if a.OnSelect != nil {
		a.OnSelect()
	}
	if o.OnSelect != nil {
		o.OnSelect(a.ID)
	}
	return true
}

func commandSlot(slot Slot) Slot {
	switch slot {
	case Root:
		return Panel
	case Leading:
		return ItemLeadingSlot
	case Trailing:
		return ItemTrailingSlot
	default:
		return slot
	}
}

func commandBack(s *CommandPaletteState, o CommandPaletteOptions) {
	before := s.Query
	s.Back()
	if before != s.Query && o.OnSearchTermChange != nil {
		o.OnSearchTermChange(s.Query)
	}
}

func commandInput(c *ui.Context, t Theme, s *CommandPaletteState, o CommandPaletteOptions, title string, disabled bool, size Size, activate func(string) bool) {
	if len(s.level) > 0 && valueOr(o.Back, true) {
		back := ButtonOptions{}
		if o.BackOptions != nil {
			back = *o.BackOptions
		}
		if back.Size == SizeDefault {
			back.Size = size
		}
		if back.Color == ColorDefault {
			back.Color = Neutral
		}
		if back.Variant == VariantDefault {
			back.Variant = LinkVariant
		}
		if back.Leading == nil {
			back.Leading = o.BackIcon
			if back.Leading == nil {
				back.Leading = lucide.ArrowLeft()
				if c.Direction() == ui.RightToLeft {
					back.Leading = lucide.ArrowRight()
				}
			}
		}
		back.Disabled = back.Disabled || disabled
		part := Button(c.Key("back"), t, message(c, "back"), back).OnClick(func() { commandBack(s, o) })
		finish(t, ComponentCommandPalette, BackSlot, part, PartStyle{}, o.Decoration)
	}
	if !valueOr(o.SearchInput, true) {
		if s.editor.Focused(c) && !disabled {
			s.list.List.Handle.Focus(c)
		}
		s.editor.CancelFocus(c)
		return
	}
	input := InputOptions{}
	if o.SearchInputOptions != nil {
		input = *o.SearchInputOptions
	}
	if input.Label == "" {
		input.Label = namedMessage(c, "searchNamed", title)
	}
	if input.Placeholder == "" {
		input.Placeholder = o.Placeholder
		if len(s.level) > 0 && s.level[len(s.level)-1].placeholder != "" {
			input.Placeholder = s.level[len(s.level)-1].placeholder
		}
		if input.Placeholder == "" {
			input.Placeholder = message(c, "searchActions")
		}
	}
	if input.Size == SizeDefault {
		input.Size = size
	}
	if input.Variant == VariantDefault {
		input.Variant = None
	}
	if input.Leading == nil && input.Avatar == nil && valueOr(o.ShowIcon, true) {
		input.Leading = o.Icon
		if input.Leading == nil {
			input.Leading = lucide.Search()
		}
	}
	input.Disabled = input.Disabled || disabled
	input.Loading = input.Loading || o.Loading || s.waiting || s.Results.Loading
	if input.LoadingIcon == nil {
		input.LoadingIcon = o.LoadingIcon
	}
	onChange, onSubmit, style := input.OnChange, input.OnSubmit, input.Style
	input.OnChange = func(value string) {
		// Cancel immediately; schedule debounce from the next valid build clock.
		s.Results.Cancel()
		s.mailbox = nil
		if onChange != nil {
			onChange(value)
		}
		if o.OnSearchTermChange != nil {
			o.OnSearchTermChange(value)
		}
	}
	input.OnSubmit = func(value string) {
		if onSubmit != nil {
			onSubmit(value)
		}
		activate(s.list.Active)
	}
	input.Style = func(root, editor, leading, trailing ui.Element) {
		editor.Bind(&s.editor).OnTextKey(func(key ui.TextKey) bool {
			if key.Mods != 0 || disabled {
				return false
			}
			if key.Key == ui.KeyDown {
				s.list.List.Handle.Focus(c)
				return true
			}
			if key.Key == ui.KeyBackspace && key.Text() == "" && len(s.level) > 0 {
				commandBack(s, o)
				return true
			}
			return false
		})
		m := optionRecipe(t, input.Size)
		root.Height(max(t.spacing([5]float32{40, 44, 48, 52, 56}[sizeIndex(input.Size)]), m.Line+2*m.Pad))
		finish(t, ComponentCommandPalette, Editor, root, PartStyle{}, o.Decoration)
		if style != nil {
			style(root, editor, leading, trailing)
		}
	}
	Input(c.Key("search"), t, &s.Query, input)
}

func commandListUI(styles Styles) Styles {
	out := Styles{}
	for _, slot := range []Slot{Root, List, Item, Body, Leading, Trailing, Label, ItemLabelSlot, Description, Indicator, EmptySlot, LoadingSlot, LoadingIconSlot, SeparatorSlot} {
		if style, ok := styles[commandSlot(slot)]; ok {
			out[slot] = style
		}
	}
	return out
}

// commandRowsCache keeps the rows made from the palette's results: only data,
// and callbacks that reach the state and the results' actions, never UI
// values of a build. Its results keep the slice they were made from alive,
// so a new slice of Results never takes its address.
type commandRowsCache struct {
	results               []CommandMatch
	index                 *CommandIndex
	parent, group         string
	limit                 int
	preserve, made        bool
	styles                Styles
	elevated, highlighted ui.Color
	items                 []ListboxItem
	matches               map[string]CommandMatch
	listUI                Styles
}

func (r *commandRowsCache) current(results []CommandMatch, s *CommandPaletteState, group string, limit int, preserve bool, styles Styles, t Theme) bool {
	return r.made && len(r.results) == len(results) && (len(results) == 0 || &r.results[0] == &results[0]) &&
		r.index == s.index && r.parent == s.parent() && r.group == group && r.limit == limit && r.preserve == preserve &&
		r.elevated == t.Elevated && r.highlighted == t.Highlighted && maps.Equal(r.styles, styles)
}

// commandItems returns the Listbox rows of the results, made again only when
// the results, the level or the look they were made with change: a frame
// with the palette open then neither groups nor copies them.
func commandItems(t Theme, s *CommandPaletteState, o CommandPaletteOptions) *commandRowsCache {
	limit := commandLimit(o.Limit)
	results := s.Results.Items
	group := ""
	if len(s.level) > 0 {
		group = s.level[len(s.level)-1].label
	}
	styles := mergeStyles(Styles{Panel: {BorderWidth: Set(float32(0)), Radius: Set(float32(0))}}, t.components[ComponentCommandPalette].UI)
	styles = mergeStyles(styles, o.UI)
	r := &s.rows
	if r.current(results, s, group, limit, o.PreserveGroupOrder, styles, t) {
		return r
	}
	ordered := groupCommands(results[:min(limit, len(results))], s.index, s.parent(), o.PreserveGroupOrder)
	*r = commandRowsCache{results: results, index: s.index, parent: s.parent(), group: group, limit: limit, preserve: o.PreserveGroupOrder, made: true,
		styles: styles, elevated: t.Elevated, highlighted: t.Highlighted, items: make([]ListboxItem, 0, len(ordered)), matches: make(map[string]CommandMatch, len(ordered)), listUI: commandListUI(styles)}
	for _, match := range ordered {
		a := match.Action
		item := ListboxItem{ID: a.ID, Label: commandLabel(a), Description: a.Description, Group: a.Group, Icon: a.Icon, Avatar: a.Avatar, Chip: a.Chip, Disabled: a.Disabled}
		if item.Group == "" {
			item.Group = group
		}
		if a.Loading {
			item.Icon, item.Avatar, item.Chip = nil, nil, nil
		}
		nodeUI := a.UI
		if a.Active {
			nodeUI = mergeStyles(mergeStyles(Styles{Item: {Background: Set(t.Elevated), Foreground: Set(t.Highlighted)}}, styles), a.UI)
		}
		item.Options = &ListboxItemOptions{OnSelect: func(*ListboxEvent) { s.previous = slices.Clone(s.selected) }, Decoration: Decoration{UI: commandListUI(nodeUI), Style: func(slot Slot, e ui.Element) {
			if a.Style != nil {
				a.Style(commandSlot(slot), e)
			}
		}}}
		r.items = append(r.items, item)
		r.matches[a.ID] = match
	}
	return r
}

func commandRows(c *ui.Context, t Theme, s *CommandPaletteState, o CommandPaletteOptions, title string, disabled bool, size Size, activate func(string) bool) {
	rows := commandItems(t, s, o)
	items, matches := rows.items, rows.matches
	pending := s.waiting || s.Results.Loading
	if len(items) == 0 && !pending && s.Results.Error == "" && o.BuildEmpty != nil {
		part := ui.Column(c.Key("empty")).FillWidth().MinWidth(0).Padding(t.spacing(24)).Role(ui.RoleStatus)
		finish(t, ComponentCommandPalette, EmptySlot, part, PartStyle{Foreground: Set(t.Muted)}, o.Decoration)
		part.Children(func() { o.BuildEmpty(c, t, s.Query) })
		return
	}
	empty := o.EmptyText
	if empty == "" && strings.TrimSpace(s.Query) != "" {
		empty = emptyMessage(c, "noMatch", "noResults")
	}
	Listbox(c.Key("results"), t, &s.selected, &s.list, items, ListboxOptions{Label: namedMessage(c, "resultsNamed", title), Size: size, Height: 280, Disabled: disabled, Loading: pending, LoadingIcon: o.LoadingIcon, EmptyText: empty, Loop: o.Loop, HighlightOnHover: o.HighlightOnHover, SelectionBehavior: SelectionReplace, SelectedIcon: o.SelectedIcon, OnSelect: func(id string) {
		// The row's OnSelect kept the choice from before this selection.
		previous := s.previous
		s.previous = nil
		if !activate(id) {
			s.selected = previous
		}
	}, OnHighlight: o.OnHighlight, Decoration: Decoration{UI: rows.listUI, Style: func(slot Slot, e ui.Element) {
		if slot == Item {
			e.AlignItems(ui.Start)
		}
		if o.Style != nil {
			o.Style(commandSlot(slot), e)
		}
	}}, Build: func(c *ui.Context, t Theme, item ListboxItem) {
		commandRow(c, t, matches[item.ID], o, size)
	}})
}

func commandText(c *ui.Context, t Theme, text string, marks []int, slot Slot, base PartStyle, o CommandPaletteOptions, a CommandAction) ui.Element {
	runes := []rune(text)
	spans := []ui.Span{}
	for start := 0; start < len(runes); {
		marked := slices.Contains(marks, start)
		end := start + 1
		for end < len(runes) && slices.Contains(marks, end) == marked {
			end++
		}
		span := ui.Span{Text: string(runes[start:end])}
		if marked {
			span.Color = t.Primary
			span.Background = t.Primary.Alpha(.15)
			span.Radius = 2
		}
		spans = append(spans, span)
		start = end
	}
	e := ui.RichText(c, spans...).MinWidth(0).SingleLine()
	finishCommandPart(t, slot, e, base, o.Decoration, a.Decoration)
	return e
}

func commandRow(c *ui.Context, t Theme, match CommandMatch, o CommandPaletteOptions, size Size) {
	a, m := match.Action, optionRecipe(t, size)
	ui.Row(c).MinWidth(0).Gap(m.Gap).AlignItems(ui.Start).Children(func() {
		if a.Loading {
			icon := o.LoadingIcon
			if icon == nil {
				icon = lucide.LoaderCircle()
			}
			part := ui.Box(c.Key("loading")).Size(m.Icon, m.Icon).Shrink(0)
			color := mergeStyle(mergeStyle(mergeStyle(PartStyle{Foreground: Set(t.Muted)}, t.components[ComponentCommandPalette].UI[ItemLeadingSlot]), o.UI[ItemLeadingSlot]), a.UI[ItemLeadingSlot]).Foreground.Value
			finishCommandPart(t, ItemLeadingSlot, part, PartStyle{}, o.Decoration, a.Decoration)
			drawLoadingSVG(part, icon, color, t.motion(c), t.Motion.Spin)
		}
		ui.Column(c.Key("body")).Grow(1).MinWidth(0).Children(func() {
			ui.Row(c.Key("label")).MinWidth(0).Gap(t.spacing(4)).Children(func() {
				if a.Prefix != "" {
					commandText(c.Key("prefix"), t, a.Prefix, nil, ItemLabelPrefixSlot, PartStyle{Foreground: Set(t.Text)}, o, a)
				}
				if a.Label != "" {
					commandText(c.Key("base"), t, a.Label, match.Highlights, ItemLabelSlot, PartStyle{Foreground: Set(t.Highlighted)}, o, a)
				}
				if a.Suffix != "" {
					commandText(c.Key("suffix"), t, a.Suffix, match.SuffixHighlights, ItemLabelSuffixSlot, PartStyle{Foreground: Set(t.Dimmed)}, o, a)
				}
			})
			if a.Description != "" {
				commandText(c.Key("description"), t, a.Description, match.DescriptionHighlights, Description, PartStyle{Foreground: Set(t.Muted)}, o, a)
			}
		})
		trailing := ui.Row(c.Key("trailing")).Shrink(0).Gap(t.spacing(4)).AlignItems(ui.Center)
		finishCommandPart(t, ItemTrailingSlot, trailing, PartStyle{}, o.Decoration, a.Decoration)
		trailing.Children(func() {
			if len(a.Children) > 0 {
				icon := o.ChildrenIcon
				if icon == nil {
					icon = lucide.ChevronRight()
					if c.Direction() == ui.RightToLeft {
						icon = lucide.ChevronLeft()
					}
				}
				ui.Icon(c, icon).Size(m.Icon, m.Icon)
			} else {
				keys := a.Kbds
				if len(keys) == 0 && a.Shortcut != "" {
					keys = []string{a.Shortcut}
				}
				if len(keys) > 0 {
					kbdSize := MD
					if size <= SM {
						kbdSize = SM
					}
					Kbd(c, t, keys, KbdOptions{Size: kbdSize})
				}
			}
		})
	})
}

func finishCommandPart(t Theme, slot Slot, e ui.Element, base PartStyle, palette, node Decoration) {
	base = mergeStyle(base, t.components[ComponentCommandPalette].UI[slot])
	base = mergeStyle(base, palette.UI[slot])
	base = mergeStyle(base, node.UI[slot])
	base.apply(e)
	if palette.Style != nil {
		palette.Style(slot, e)
	}
	if node.Style != nil {
		node.Style(slot, e)
	}
}
