package nui

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// Repeated characters cycle, while a longer prefix can keep the current match.
func typeaheadIndex(prefix string, current, count int, label func(int) (string, bool)) int {
	chars := []rune(prefix)
	if len(chars) == 0 {
		return -1
	}
	repeated := true
	for _, ch := range chars[1:] {
		if ch != chars[0] {
			repeated = false
			break
		}
	}
	start := max(current, 0)
	if repeated {
		prefix = string(chars[0])
		start = current + 1
	}
	for j := range count {
		i := (start + j) % count
		text, enabled := label(i)
		if enabled && strings.HasPrefix(strings.ToLower(text), prefix) {
			return i
		}
	}
	return -1
}

// A popup's keyboard candidates belong to its data, independently of the rows
// built by ui.List. Native SelectBase/ComboboxBase derive them from Item calls
// and therefore cannot navigate a virtualized set of items.
type optionIdentity[T comparable] struct {
	value              T
	label, description string
	disabled           bool
}

type optionListState[T comparable] struct {
	list         ui.ListState
	identities   []optionIdentity[T]
	index        map[T]int
	keys         []any
	active       Optional[T]
	descriptions int
	enabledCount int
}

func (s *optionListState[T]) update(options []Option[T]) bool {
	return s.updateRows(len(options), func(i int) Option[T] { return options[i] })
}

func (s *optionListState[T]) updateRows(count int, option func(int) Option[T]) bool {
	changed := count != len(s.identities)
	if !changed {
		for i := range count {
			o := option(i)
			if s.identities[i] != (optionIdentity[T]{o.Value, o.Label, o.Description, o.Disabled}) {
				changed = true
				break
			}
		}
	}
	if !changed {
		return false
	}
	s.identities = make([]optionIdentity[T], count)
	s.index = make(map[T]int, count)
	s.keys = make([]any, count)
	s.descriptions, s.enabledCount = 0, 0
	for i := range count {
		o := option(i)
		if _, exists := s.index[o.Value]; o.Value != o.Value || exists {
			panic("nui: selection values must be unique and reflexive")
		}
		s.identities[i] = optionIdentity[T]{o.Value, o.Label, o.Description, o.Disabled}
		s.index[o.Value] = i
		s.keys[i] = optionKey[T]{o.Value}
		if o.Description != "" {
			s.descriptions++
		}
		if !o.Disabled {
			s.enabledCount++
		}
	}
	return true
}

func (s *optionListState[T]) enabled(value T) (int, bool) {
	i, ok := s.index[value]
	return i, ok && !s.identities[i].disabled
}

func (s *optionListState[T]) highlight(value T, scroll bool) {
	if i, ok := s.enabled(value); ok {
		s.active = Set(value)
		if scroll {
			s.list.ScrollIntoView(i)
		}
	}
}

func (s *optionListState[T]) move(key ui.Key) {
	at, ok := s.enabled(s.active.Value)
	if !s.active.Set || !ok {
		at = -1
	}
	delta := 1
	switch key {
	case ui.KeyHome:
		at = -1
	case ui.KeyEnd:
		at, delta = len(s.identities), -1
	case ui.KeyUp:
		delta = -1
		if at < 0 {
			at = len(s.identities)
		}
	}
	for i := at + delta; i >= 0 && i < len(s.identities); i += delta {
		if !s.identities[i].disabled {
			s.highlight(s.identities[i].value, true)
			return
		}
	}
}

func (s *optionListState[T]) moveWrapped(key ui.Key, loop bool) {
	previous := s.active
	s.move(key)
	if loop && s.active == previous {
		switch key {
		case ui.KeyDown:
			s.move(ui.KeyHome)
		case ui.KeyUp:
			s.move(ui.KeyEnd)
		}
	}
}

func optionList[T comparable](c *ui.Context, t Theme, s *optionListState[T], owner ui.Element, count int, option func(int) Option[T], height float32, component Component, decoration Decoration, size Size, selectedIcon *ui.SVG, highlightOnHover bool, chosen func(T) bool, choose func(T)) {
	m := optionRecipe(t, size)
	// Rows are a line and their padding; a description adds a line of the same size.
	estimate := float32(count)*(m.Line+2*m.Pad) + float32(s.descriptions)*m.Line
	list := ui.List(c.Key("options"), &s.list, count).ItemKey(func(i int) any { return s.keys[i] })
	list.MinWidth(0).FillWidth().Height(min(height, estimate)).KeepFocus()
	list.Rows(func(r ui.ListRow) {
		o := option(r.Index)
		item := ui.Row(r.Context).KeepFocus().Role(ui.RoleListItem).Label(o.Label).Description(o.Description).Disabled(o.Disabled).Selected(chosen(o.Value))
		highlighted := s.active.Set && s.active.Value == o.Value && !o.Disabled
		if highlighted {
			owner.ActiveDescendant(item)
		}
		optionStyle(item, t, component, decoration, size, highlighted)
		item.HandleInput(func(ev ui.InputEvent) bool {
			if highlightOnHover && ev.Kind == ui.InputPointerMove && !o.Disabled && (!s.active.Set || s.active.Value != o.Value) {
				s.highlight(o.Value, false)
				return true
			}
			return false
		})
		item.OnClick(func() { choose(o.Value) })
		item.Children(func() { optionContent(r.Context, t, o, chosen(o.Value), highlighted, size, selectedIcon) })
	})
}
