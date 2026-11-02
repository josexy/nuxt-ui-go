package nui

import (
	"slices"
	"strings"
)

// Keep an owned snapshot: callers may edit a slice in place between builds.
// Comparing it is linear, but unchanged data needs no maps, row keys or slices
// allocated again. Only data is cached, never current-build UI values.
type listboxItems struct {
	items    []ListboxItem
	index    map[string]int
	revision uint64
}

func (s *listboxItems) update(items []ListboxItem) bool {
	if slices.Equal(s.items, items) {
		return false
	}
	index := make(map[string]int, len(items))
	for i, item := range items {
		if _, exists := index[item.ID]; item.ID == "" || exists {
			panic("nui: Listbox requires unique non-empty IDs")
		}
		index[item.ID] = i
	}
	s.items, s.index = slices.Clone(items), index
	s.revision++
	return true
}

func (s *listboxItems) enabledIndex(id string) int {
	if i, ok := s.index[id]; ok && !s.items[i].Disabled {
		return i
	}
	return -1
}

type listboxRow struct {
	item  int
	group string
	key   any
}

type listboxSpacer bool

type listboxData struct {
	listboxItems
	rows      []listboxRow
	positions []int
}

func (s *listboxData) update(items []ListboxItem) {
	if !s.listboxItems.update(items) {
		return
	}
	s.rows = make([]listboxRow, 0, len(items))
	s.positions = make([]int, len(items))
	previousGroup, groupNumber := "", 0
	for i, item := range items {
		if item.Group != "" && item.Group != previousGroup {
			groupNumber++
			s.rows = append(s.rows, listboxRow{item: -1, group: item.Group, key: struct {
				Group  string
				Number int
			}{item.Group, groupNumber}})
		}
		s.positions[i] = len(s.rows)
		s.rows = append(s.rows, listboxRow{item: i, key: item.ID})
		previousGroup = item.Group
	}
}

type stringSelection struct {
	values []string
	set    map[string]bool
}

func (s *stringSelection) update(values []string) {
	if slices.Equal(s.values, values) {
		return
	}
	s.values = slices.Clone(values)
	s.set = make(map[string]bool, len(values))
	for _, id := range values {
		s.set[id] = true
	}
}

func (s *SelectMenuState) selectionLabel(values []string, dataChanged bool) string {
	if !dataChanged && slices.Equal(s.labelIDs, values) {
		return s.label
	}
	labels := make([]string, len(values))
	cache := make(map[string]string, len(values))
	for i, id := range values {
		label, known := s.labels[id]
		if at, ok := s.data.index[id]; ok {
			label = s.data.items[at].Label
		} else if !known {
			label = id
		}
		labels[i], cache[id] = label, label
	}
	s.labelIDs, s.labels = slices.Clone(values), cache
	s.label = strings.Join(labels, ", ")
	return s.label
}

func (s *SelectMenuState) filter(query string, fields []string) []ListboxItem {
	query = strings.ToLower(query)
	if s.filterRevision == s.data.revision && s.filterQuery == query &&
		(fields == nil) == (s.filterFields == nil) && slices.Equal(s.filterFields, fields) {
		return s.filtered
	}
	s.filterRevision, s.filterQuery = s.data.revision, query
	s.filterFields = slices.Clone(fields)
	if fields == nil {
		fields = []string{"label"}
	}
	s.filtered = nil
	for _, item := range s.data.items {
		for _, field := range fields {
			var text string
			switch field {
			case "label":
				text = item.Label
			case "description":
				text = item.Description
			case "id", "value":
				text = item.ID
			case "group":
				text = item.Group
			default:
				panic("nui: unknown option filter field")
			}
			if strings.Contains(strings.ToLower(text), query) {
				s.filtered = append(s.filtered, item)
				break
			}
		}
	}
	return s.filtered
}
