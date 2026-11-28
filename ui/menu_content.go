package nui

import (
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// Small menus shrink to their content. Inspect only a bounded prefix: long
// levels fill the viewport, and List measures their visible variable rows.
func menuViewportHeight(t Theme, m optionMetrics, items menuView, measured map[string]float32, limit float32) float32 {
	height := float32(0)
	for i := range min(items.count(), 64) {
		it := items.item(i)
		h, known := measured[it.ID]
		if !known {
			h = m.Line + 2*m.Pad
			if it.Description != "" {
				h += m.Line
			}
			if it.Kind == MenuSeparator {
				h = 1 + 2*t.spacing(4)
			} else if it.Avatar != nil && it.Avatar.Size != SizeDefault {
				h = max(h, avatarSize(t, it.Avatar.Size)+2*m.Pad)
			}
		}
		height += h
		if height >= limit {
			return limit
		}
	}
	if items.count() > 64 {
		return limit
	}
	return height
}

func (level *menuLevel) measure(c *ui.Context, row ui.Element, id string, margin float32) {
	services := c.Services()
	mount := level.search.search.mount
	row.Draw(func(_ *ui.Painter, r ui.Rect) {
		if level.search.search.mount != mount {
			return
		}
		height := r.H + margin
		if previous, known := level.heights[id]; !known || previous != height {
			if level.heights == nil {
				level.heights = make(map[string]float32)
			}
			level.heights[id] = height
			services.Invalidate()
		}
	})
}

func submenuOptions(root MenuOptions, item MenuItem) MenuOptions {
	o := root
	o.Label = item.Label
	o.Placement, o.Offset = Optional[Placement]{}, Optional[float32]{}
	o.Filter, o.SearchTerm, o.FilterInput, o.OnSearchTermChange = false, nil, nil, nil
	if v := item.Content; v != nil {
		if v.Label != "" {
			o.Label = v.Label
		}
		o.Width, o.MaxHeight = mergeValue(o.Width, v.Width), mergeValue(o.MaxHeight, v.MaxHeight)
		o.Placement, o.Offset = v.Placement, v.Offset
		o.Loop = valueOr(v.Loop, o.Loop)
		o.Filter, o.SearchTerm, o.FilterInput, o.OnSearchTermChange = v.Filter, v.SearchTerm, v.FilterInput, v.OnSearchTermChange
		if v.FilterFields != nil {
			o.FilterFields = v.FilterFields
		}
		o.SearchExecution = valueOr(v.SearchExecution, o.SearchExecution)
		o.IgnoreFilter = valueOr(v.IgnoreFilter, o.IgnoreFilter)
		o.UI = mergeStyles(o.UI, v.UI)
		if v.Style != nil {
			parent := o.Style
			o.Style = func(slot Slot, e ui.Element) {
				if parent != nil {
					parent(slot, e)
				}
				v.Style(slot, e)
			}
		}
	}
	return o
}

func menuItemContent(c *ui.Context, t Theme, it MenuItem, o MenuOptions, d Decoration, id Component, size Size, m optionMetrics, leading ui.Color) {
	if it.Loading || it.Icon != nil {
		icon := it.Icon
		if it.Loading {
			icon = o.LoadingIcon
			if icon == nil {
				icon = lucide.LoaderCircle()
			}
		}
		var e ui.Element
		if it.Loading {
			e = ui.Box(c.Key("leading")).Size(m.Icon, m.Icon).Shrink(0)
		} else {
			e = ui.Icon(c.Key("leading"), icon).Size(m.Icon, m.Icon).Shrink(0)
		}
		finish(t, id, Leading, e, PartStyle{Foreground: Set(leading)}, d)
		if it.Loading {
			color := mergeValue(Set(leading), t.components[id].UI[Leading].Foreground)
			color = mergeValue(color, d.UI[Leading].Foreground)
			drawLoadingSVG(e, icon, color.Value, t.fades(c), t.Motion.Spin)
		}
	} else if it.Avatar != nil {
		config := *it.Avatar
		diameter := Optional[float32]{}
		if config.Size == SizeDefault {
			diameter = Set(t.text([5]float32{16, 16, 20, 20, 24}[sizeIndex(size)]))
		}
		finish(t, id, Leading, avatar(c.Key("leading"), t, config, diameter).Shrink(0), PartStyle{}, d)
	}
	if it.Build != nil {
		it.Build(c, t)
	} else {
		body := ui.Column(c.Key("body")).Grow(1).MinWidth(0)
		finish(t, id, Body, body, PartStyle{Gap: Set(float32(0))}, d)
		body.Children(func() {
			if it.Label != "" {
				finish(t, id, Label, ui.Text(c.Key("label"), it.Label).SingleLine().MinWidth(0), PartStyle{}, d)
			}
			if it.Description != "" {
				finish(t, id, Description, ui.Text(c.Key("description"), it.Description).SingleLine().MinWidth(0), PartStyle{Foreground: Set(t.Muted)}, d)
			}
		})
	}
	if len(it.Children) == 0 && it.Shortcut == "" && !valueOr(it.Checked, false) {
		return
	}
	trailing := ui.Row(c.Key("trailing")).Shrink(0).AlignItems(ui.Center)
	finish(t, id, Trailing, trailing, PartStyle{Gap: Set(t.spacing(6))}, d)
	trailing.Children(func() {
		if len(it.Children) > 0 {
			chevron := lucide.ChevronRight()
			if c.Direction() == ui.RightToLeft {
				chevron = lucide.ChevronLeft()
			}
			ui.Icon(c.Key("children"), chevron).Size(m.Icon, m.Icon).Shrink(0)
		} else if it.Shortcut != "" {
			kbdSize, gap := MD, float32(2)
			switch size {
			case XS, SM:
				kbdSize = SM
			case LG:
				gap = 4
			case XL:
				kbdSize, gap = LG, 4
			}
			Kbd(c.Key("keys"), t, shortcutKeys(it.Shortcut), KbdOptions{Size: kbdSize}).Gap(t.spacing(gap))
		}
		if valueOr(it.Checked, false) {
			icon := o.CheckedIcon
			if icon == nil {
				icon = lucide.Check()
			}
			finish(t, id, Indicator, ui.Icon(c.Key("checked"), icon).Size(m.Icon, m.Icon).Shrink(0), PartStyle{}, d)
		}
	})
}

func (s *MenuState) menuFocused(c *ui.Context) bool {
	for _, parent := range s.activeLevels {
		if level := s.levels[parent]; level != nil && level.panel.FocusWithin(c) {
			return true
		}
	}
	return false
}

func (s *MenuState) moveFocus(items menuView, current string, back, loop bool) {
	if len(items.enabled) == 0 {
		return
	}
	at := items.enabledAt(current)
	position, found := slices.BinarySearch(items.enabled, at)
	if back {
		position--
	} else if found {
		position++
	}
	if position < 0 || position >= len(items.enabled) {
		if !loop {
			return
		}
		position = (position + len(items.enabled)) % len(items.enabled)
	}
	s.focusItem(items.item(items.enabled[position]).ID)
}

func (s *MenuState) activate(it MenuItem, depth int, notify func(bool)) {
	if !menuEnabled(it) {
		return
	}
	if len(it.Children) > 0 {
		s.child(depth, it, true)
		return
	}
	if valueOr(it.Close, true) {
		s.close(true, notify)
	}
	if it.Checked.Set && it.OnCheckedChange != nil {
		next := !it.Checked.Value
		if it.Radio {
			next = true
		}
		it.OnCheckedChange(next)
	}
	if it.OnSelect != nil {
		it.OnSelect()
	}
}

func menuFilter(c *ui.Context, t Theme, panel ui.Element, s *MenuState, level *menuLevel, query *string, items menuView, o MenuOptions, size Size, id Component, parent string, depth int) {
	input := InputOptions{}
	if o.FilterInput != nil {
		input = *o.FilterInput
	}
	if input.Label == "" {
		input.Label = message(c, "searchActions")
	}
	if input.Placeholder == "" {
		input.Placeholder = message(c, "searchActions")
	}
	if input.Size == SizeDefault {
		input.Size = size
	}
	if input.Variant == VariantDefault {
		input.Variant = None
	}
	notify := input.OnChange
	input.OnChange = func(value string) {
		s.resetHover()
		if notify != nil {
			notify(value)
		}
		if o.OnSearchTermChange != nil {
			o.OnSearchTermChange(value)
		}
	}
	submit := input.OnSubmit
	input.OnSubmit = func(value string) {
		if !level.search.search.ready(*query) {
			return
		}
		if submit != nil {
			submit(value)
		}
		if at := items.enabledAt(level.hover); at >= 0 {
			s.activate(items.item(at), depth, o.OnOpenChange)
		} else if len(items.enabled) > 0 {
			s.activate(items.item(items.enabled[0]), depth, o.OnOpenChange)
		}
	}
	style := input.Style
	input.Style = func(root, editor, leading, trailing ui.Element) {
		editor.Bind(&level.filter)
		editor.OnShortcut(0, ui.KeyDown, func() { s.resetHover(); s.focusEdge(items, false) }).
			OnShortcut(0, ui.KeyUp, func() { s.resetHover(); s.focusEdge(items, true) }).
			OnShortcut(0, ui.KeyEscape, func() {
				s.resetHover()
				if *query != "" {
					*query = ""
					input.OnChange("")
				} else if depth > 0 {
					s.path = s.path[:depth-1]
					s.focusItem(parent)
				} else {
					s.close(true, o.OnOpenChange)
				}
			})
		editor.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputKeyDown && ev.Key == ui.KeyTab {
				s.close(false, o.OnOpenChange)
			}
			return false
		})
		finish(t, id, Editor, root, PartStyle{}, o.Decoration)
		if style != nil {
			style(root, editor, leading, trailing)
		}
	}
	Input(c.Key("filter"), t, query, input).Margin(0, 0, t.spacing(4), 0)
	started := ui.Local(panel, "nui-menu-filter", func() bool { return false })
	if !*started && !input.Disabled {
		*started = true
		if depth == 0 {
			s.focus, s.focusID = &level.filter, ""
		}
	}
}
