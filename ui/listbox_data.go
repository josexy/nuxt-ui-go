package nui

import (
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
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
		if item.Kind > ListboxSeparator {
			panic("nui: invalid listbox item kind")
		}
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
	if i, ok := s.index[id]; ok && !s.items[i].Disabled && s.items[i].Kind == ListboxOption {
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
	labels := make([]string, 0, len(values))
	cache := make(map[string]string, len(values))
	for _, id := range values {
		label, known := s.labels[id]
		if at, ok := s.data.index[id]; ok {
			label = s.data.items[at].Label
		} else if !known {
			label = id
		}
		cache[id] = label
		if label != "" {
			labels = append(labels, label)
		}
	}
	s.labelIDs, s.labels = slices.Clone(values), cache
	s.label = strings.Join(labels, ", ")
	return s.label
}

type listboxFilter struct {
	items    []ListboxItem
	query    string
	fields   []string
	revision uint64
}

func foldListboxSearch(text string) string {
	text = norm.NFKD.String(cases.Fold().String(text))
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, text)
}

func listboxMatch(item ListboxItem, query string, fields []string) int {
	best := 3
	for _, field := range fields {
		var value string
		switch field {
		case "label":
			value = item.Label
		case "description":
			value = item.Description
		case "id", "value":
			value = item.ID
		case "group":
			value = item.Group
		}
		value = foldListboxSearch(value)
		switch {
		case value == query:
			best = 0
		case strings.HasPrefix(value, query):
			best = min(best, 1)
		case strings.Contains(value, query):
			best = min(best, 2)
		}
	}
	return best
}

// Filter each consecutive group independently. Structural entries keep their
// position between runs; a group with no matching selectable item disappears.
func (s *listboxFilter) apply(data *listboxItems, query string, fields []string) []ListboxItem {
	query = foldListboxSearch(query)
	if s.revision == data.revision && s.query == query &&
		(fields == nil) == (s.fields == nil) && slices.Equal(s.fields, fields) {
		return s.items
	}
	s.revision, s.query, s.fields = data.revision, query, slices.Clone(fields)
	if fields == nil {
		fields = []string{"label"}
	}
	for _, field := range fields {
		if !slices.Contains([]string{"label", "description", "id", "value", "group"}, field) {
			panic("nui: unknown option filter field")
		}
	}
	s.items = nil
	for start := 0; start < len(data.items); {
		end := start + 1
		for end < len(data.items) && data.items[end].Group == data.items[start].Group {
			end++
		}
		groupStart, matches := len(s.items), 0
		for run := start; run < end; {
			if data.items[run].Kind != ListboxOption {
				s.items = append(s.items, data.items[run])
				run++
				continue
			}
			last := run
			for last < end && data.items[last].Kind == ListboxOption {
				last++
			}
			var buckets [3][]ListboxItem
			for _, item := range data.items[run:last] {
				if score := listboxMatch(item, query, fields); score < len(buckets) {
					buckets[score] = append(buckets[score], item)
					matches++
				}
			}
			for _, bucket := range buckets {
				s.items = append(s.items, bucket...)
			}
			run = last
		}
		if matches == 0 {
			s.items = s.items[:groupStart]
		}
		start = end
	}
	return s.items
}

func (s *SelectMenuState) filter(query string, fields []string) []ListboxItem {
	return s.filterCache.apply(&s.data, query, fields)
}
