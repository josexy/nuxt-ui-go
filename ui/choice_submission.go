package nui

import (
	"maps"
	"slices"
)

// SelectState owns the submitted typed options and popup state for one control
// in one window. Submit options outside views and do not copy a mounted state.
// SelectOptions.Open, when supplied, takes precedence over Open.
type SelectState[T comparable] struct {
	Open                                      bool
	source                                    submittedOptions[T]
	memory                                    selectionMemory[T]
	labels                                    map[T]string
	label                                     string
	labelSelection                            *SelectionState[T]
	labelDataRevision, labelSelectionRevision uint64
	preferred                                 Optional[T]
}

type submittedOptions[T comparable] struct {
	items    []Option[T]
	metadata optionMetadata[T]
	revision uint64
}

func submitOptions[T comparable](options []Option[T], revision uint64) submittedOptions[T] {
	if len(options) == 0 {
		return submittedOptions[T]{revision: revision + 1}
	}
	owned := slices.Clone(options)
	var next optionListState[T]
	// Build and validate a temporary model, so a panic cannot corrupt a
	// previous usable submission or the mounted list's identity/focus state.
	next.update(owned)
	return submittedOptions[T]{items: owned, metadata: next.optionMetadata, revision: revision + 1}
}

// SetOptions owns a snapshot and validates unique, reflexive typed identities.
// Zero values are valid identities. Callbacks and immutable icon assets are
// referenced; changing a descriptor's callback requires another submission.
// Every accepted call advances the data revision; nil/empty releases its data.
func (s *SelectState[T]) SetOptions(options []Option[T]) {
	s.source = submitOptions(options, s.source.revision)
	if len(options) == 0 {
		s.memory.list.optionMetadata = optionMetadata[T]{}
		s.memory.pending = Optional[T]{}
	}
}

// SetOptions owns an InputMenu dataset with unique, nonempty string IDs.
// Query, selection and opening are independent of the submitted descriptors.
func (s *SearchState) SetOptions(options []Option[string]) {
	for _, option := range options {
		if option.Value == "" {
			panic("nui: InputMenu IDs cannot be empty")
		}
	}
	s.source = submitOptions(options, s.source.revision)
	s.searchData = &choiceSearchData{kind: searchOptions, candidates: len(options), records: make([]choiceSearchRecord, len(options))}
	for i, option := range s.source.items {
		s.searchData.records[i] = choiceSearchRecord{id: option.Value, label: option.Label, description: option.Description, disabled: option.Disabled}
	}
	s.memory.filtered, s.memory.filteredSet = nil, nil
	s.memory.search.reset()
	if len(options) == 0 {
		s.memory.data.optionMetadata, s.memory.list.optionMetadata = optionMetadata[string]{}, optionMetadata[string]{}
		s.memory.pending = Optional[string]{}
	}
}

func ownedAvatar(source *AvatarOptions) *AvatarOptions {
	if source == nil {
		return nil
	}
	next := *source
	next.UI = maps.Clone(source.UI)
	return &next
}

func ownedInput(source *InputOptions) *InputOptions {
	if source == nil {
		return nil
	}
	next := *source
	next.UI, next.Avatar = maps.Clone(source.UI), ownedAvatar(source.Avatar)
	return &next
}

func submitListboxItems(items []ListboxItem, revision uint64) listboxItems {
	if len(items) == 0 {
		return listboxItems{revision: revision + 1}
	}
	owned := slices.Clone(items)
	for i := range owned {
		item := &owned[i]
		item.Avatar = ownedAvatar(item.Avatar)
		if item.Chip != nil {
			chip := *item.Chip
			chip.UI = maps.Clone(chip.UI)
			item.Chip = &chip
		}
		if item.Options != nil {
			options := *item.Options
			options.UI = maps.Clone(options.UI)
			item.Options = &options
		}
	}
	next := listboxItems{items: owned, index: indexListboxItems(owned), revision: revision + 1}
	next.prepareRows()
	next.prepareSearch()
	return next
}

func (s *ListboxState) submit(source listboxItems) {
	s.source = source
	s.data = listboxData{}
	s.filter.reset()
	// Keep the app's selection, draft and surviving active identity. The next
	// view reconciles logical rows and scrolling against this revision.
	if len(source.items) == 0 {
		s.scrollTo, s.selectionAnchor = "", ""
	}
}

// SetItems owns item descriptors, nested avatar/chip/options and style maps.
// IDs (including structural rows) must be unique and nonempty. Invalid input
// leaves the previous dataset intact. Nil/empty clears data and filter caches.
func (s *ListboxState) SetItems(items []ListboxItem) {
	s.submit(submitListboxItems(items, s.source.revision))
}

// SetItems submits once to both the trigger and embedded list. They share the
// immutable owned snapshot; no cloning setter is needed during a view build.
func (s *SelectMenuState) SetItems(items []ListboxItem) {
	next := submitListboxItems(items, s.data.revision)
	s.data = next
	s.filterCache.reset()
	s.creation = listboxCreation{}
	s.Listbox.submit(next)
}

type menuItemLocation struct {
	parent string
	index  int
	item   *MenuItem
}

type submittedMenuLevel struct {
	search *choiceSearchData
	items  []MenuItem
	menuMetadata
}

func (s submittedMenuLevel) view() menuView {
	return menuView{items: s.items, menuMetadata: s.menuMetadata}
}

type submittedMenu struct {
	items    []MenuItem
	byID     map[string]menuItemLocation
	levels   map[string]submittedMenuLevel
	revision uint64
}

func (s *submittedMenu) own(items []MenuItem, parent string, depth int) []MenuItem {
	if depth > 32 {
		panic("nui: menu nesting exceeds 32 levels")
	}
	if len(items) == 0 {
		return nil
	}
	owned := slices.Clone(items)
	var enabled []int
	index := make(map[string]int, len(owned))
	for i := range owned {
		item := &owned[i]
		if _, exists := s.byID[item.ID]; item.ID == "" || exists {
			panic("nui: Menu requires unique non-empty IDs")
		}
		if item.Kind > MenuSeparator {
			panic("nui: invalid menu item kind")
		}
		s.byID[item.ID] = menuItemLocation{parent: parent, index: i, item: item}
		index[item.ID] = i
		item.UI, item.Avatar = maps.Clone(item.UI), ownedAvatar(item.Avatar)
		if item.Content != nil {
			content := *item.Content
			content.UI = maps.Clone(content.UI)
			content.FilterFields = slices.Clone(content.FilterFields)
			content.FilterInput = ownedInput(content.FilterInput)
			// SearchTerm is an intentional app-owned binding, not a descriptor
			// container. Preserve the pointer just as callbacks retain app data.
			item.Content = &content
		}
		item.Children = s.own(item.Children, item.ID, depth+1)
		if menuEnabled(*item) {
			enabled = append(enabled, i)
		}
	}
	search := &choiceSearchData{kind: searchMenu, records: make([]choiceSearchRecord, len(owned))}
	for i, item := range owned {
		if item.Kind == MenuAction {
			search.candidates++
		}
		search.records[i] = choiceSearchRecord{id: item.ID, label: item.Label, description: item.Description, shortcut: item.Shortcut, structural: item.Kind != MenuAction, disabled: item.Disabled}
	}
	s.levels[parent] = submittedMenuLevel{items: owned, menuMetadata: menuMetadata{index: index, enabled: enabled}, search: search}
	return owned
}

// SetItems owns and validates the complete rich-menu tree. IDs are unique
// across levels; nested descriptor containers and style maps are copied.
// Assets, callbacks and explicit app-state bindings remain references. Call
// only at initialization or an actual data change, on the owning UI thread.
func (s *MenuState) SetItems(items []MenuItem) {
	next := submittedMenu{revision: s.source.revision + 1}
	if len(items) > 0 {
		next.byID = make(map[string]menuItemLocation)
		next.levels = make(map[string]submittedMenuLevel)
		next.items = next.own(items, "", 0)
	}
	for _, level := range s.levels {
		level.search.search.reset()
		level.heights = nil
		level.view = menuView{}
	}
	s.source, s.dataChanged = next, true
	if len(items) == 0 {
		if s.focus != nil {
			s.focus.CancelFocus()
		}
		s.focus, s.handles, s.levels, s.path = nil, nil, nil, nil
		s.focusID, s.activeLevels = "", nil
	}
}
