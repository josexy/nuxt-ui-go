package nui

import (
	"strings"
)

// An immutable submitted dataset and its row metadata. Source revisions and
// explicit derived-result revisions replace per-build descriptor comparisons.
type listboxItems struct {
	// order indexes the immutable source; nil is the unfiltered source.
	order     []int
	action    *ListboxItem
	search    *choiceSearchData
	items     []ListboxItem
	index     map[string]int
	revision  uint64
	rows      []listboxRow
	positions []int
	enabled   []int
}

func (s *listboxItems) count() int {
	if s.order != nil {
		return len(s.order)
	}
	return len(s.items)
}

func (s *listboxItems) item(at int) ListboxItem {
	if s.order != nil {
		at = s.order[at]
		if at < 0 {
			return *s.action
		}
	}
	return s.items[at]
}

func (s *listboxItems) prepareSearch() {
	s.search = &choiceSearchData{kind: searchListbox, records: make([]choiceSearchRecord, len(s.items))}
	for i, item := range s.items {
		if item.Kind == ListboxOption {
			s.search.candidates++
		}
		s.search.records[i] = choiceSearchRecord{id: item.ID, label: item.Label, description: item.Description, group: item.Group, structural: item.Kind != ListboxOption, disabled: item.Disabled}
	}
}

func indexListboxItems(items []ListboxItem) map[string]int {
	if len(items) == 0 {
		return nil
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
	return index
}

func (s *listboxItems) enabledIndex(id string) int {
	if i, ok := s.index[id]; ok && !s.item(i).Disabled && s.item(i).Kind == ListboxOption {
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
	owner *listboxItems
}

func (s *listboxData) bind(source *listboxItems) bool {
	if s.owner == source && s.revision == source.revision {
		return false
	}
	s.listboxItems, s.owner = *source, source
	return true
}

func (s *listboxItems) prepareRows() {
	if s.count() == 0 {
		s.rows, s.positions, s.enabled, s.index = nil, nil, nil, nil
		return
	}
	s.rows = make([]listboxRow, 0, s.count())
	s.positions = make([]int, s.count())
	s.enabled = nil
	indexing := s.index == nil
	if indexing {
		s.index = make(map[string]int, s.count())
	}
	previousGroup, groupNumber := "", 0
	for i := range s.count() {
		item := s.item(i)
		if indexing {
			s.index[item.ID] = i
		}
		if item.Group != "" && item.Group != previousGroup {
			groupNumber++
			s.rows = append(s.rows, listboxRow{item: -1, group: item.Group, key: struct {
				Group  string
				Number int
			}{item.Group, groupNumber}})
		}
		s.positions[i] = len(s.rows)
		s.rows = append(s.rows, listboxRow{item: i, key: item.ID})
		if item.Kind == ListboxOption && !item.Disabled {
			s.enabled = append(s.enabled, i)
		}
		previousGroup = item.Group
	}
}

func (s *ListboxState) observeSelection(selected *SelectionState[string], dataChanged bool) bool {
	external := s.selection != selected || s.selectionRevision != selected.Revision()
	if external || dataChanged {
		s.preferred = ""
		for _, id := range selected.values {
			if s.data.enabledIndex(id) >= 0 {
				s.preferred = id
				break
			}
		}
		s.selection, s.selectionRevision = selected, selected.Revision()
	}
	return external
}

// Creation is a separate action row, cached with its source and presentation.
// It is never inserted into the application's submitted business values.
type listboxCreation struct {
	data                         listboxItems
	source                       *listboxItems
	sourceRevision, fullRevision uint64
	label, id                    string
	top                          bool
}

func (s *listboxCreation) prepare(full, source *listboxItems, label string, top bool) *listboxItems {
	if s.source == source && s.sourceRevision == source.revision && s.fullRevision == full.revision && s.label == label && s.top == top {
		return &s.data
	}
	id := "\x00nui:create"
	for {
		if _, exists := full.index[id]; !exists {
			break
		}
		id += "_"
	}
	entry := ListboxItem{ID: id, Label: label}
	order := make([]int, 0, source.count()+1)
	if top {
		order = append(order, -1)
	}
	if source.order != nil {
		order = append(order, source.order...)
	} else {
		for i := range source.count() {
			order = append(order, i)
		}
	}
	if !top {
		order = append(order, -1)
	}
	s.data = listboxItems{items: source.items, order: order, action: &entry, revision: s.data.revision + 1}
	s.data.prepareRows()
	s.source, s.sourceRevision, s.fullRevision = source, source.revision, full.revision
	s.label, s.id, s.top = label, id, top
	return &s.data
}

func (s *SelectMenuState) selectionLabel(selection *SelectionState[string]) string {
	if s.labelSelection == selection && s.labelDataRevision == s.data.revision && s.labelSelectionRevision == selection.Revision() {
		return s.label
	}
	labels := make([]string, 0, selection.Len())
	cache := make(map[string]string, selection.Len())
	for _, id := range selection.values {
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
	s.labelSelection, s.labels = selection, cache
	s.labelDataRevision, s.labelSelectionRevision = s.data.revision, selection.Revision()
	s.label = strings.Join(labels, ", ")
	return s.label
}

type listboxFilter struct {
	result listboxItems
	search choiceSearchCache
}

func (s *listboxFilter) visible(data *listboxItems, query string, fields []string, execution SearchExecution) *listboxItems {
	source := data.search
	if source == nil {
		source = emptyListboxSearch
	}
	if s.search.execute(source, query, fields, execution) {
		s.result = s.search.result.listbox
		s.result.items = data.items
		s.result.revision = s.search.revision
	}
	if s.search.pending {
		return &emptyChoiceItems
	}
	return &s.result
}

func (s *listboxFilter) reset() { s.search.reset(); s.result = listboxItems{} }

var emptyChoiceItems listboxItems
var emptyListboxSearch = &choiceSearchData{kind: searchListbox}
