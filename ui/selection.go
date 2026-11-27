package nui

import "slices"

// SelectionState owns an ordered, unique selection. Its zero value is empty.
// Keep it in persistent application state and do not copy it after use. While
// mounted, mutate it on the UI thread. Methods never emit component callbacks.
// Values makes an explicit O(S) snapshot; Len, Has and Revision do not copy it.
type SelectionState[K comparable] struct {
	values    []K
	positions map[K]int
	revision  uint64
}

func checkSelectionKey[K comparable](key K) {
	if key != key {
		panic("nui: selection values must be reflexive")
	}
}

// Set owns a snapshot, deduplicating in first-occurrence order. It returns true
// and advances Revision only when the ordered selection changes. Invalid keys
// panic before replacing the previous selection. Nil and empty both clear it.
func (s *SelectionState[K]) Set(values []K) bool {
	if len(values) == 0 {
		return s.Clear()
	}
	positions := make(map[K]int, len(values))
	next := make([]K, 0, len(values))
	for _, key := range values {
		checkSelectionKey(key)
		if _, exists := positions[key]; !exists {
			positions[key] = len(next)
			next = append(next, key)
		}
	}
	if slices.Equal(s.values, next) {
		return false
	}
	s.values, s.positions = next, positions
	s.revision++
	return true
}

// Add appends a missing key, returning whether the selection changed.
func (s *SelectionState[K]) Add(key K) bool {
	checkSelectionKey(key)
	if _, exists := s.positions[key]; exists {
		return false
	}
	if s.positions == nil {
		s.positions = make(map[K]int)
	}
	s.positions[key] = len(s.values)
	s.values = append(s.values, key)
	s.revision++
	return true
}

// Remove preserves the order of surviving keys, returning whether it changed
// the selection. Moving the remaining ordered keys can take O(S) work.
func (s *SelectionState[K]) Remove(key K) bool {
	checkSelectionKey(key)
	at, exists := s.positions[key]
	if !exists {
		return false
	}
	if len(s.values) == 1 {
		return s.Clear()
	}
	delete(s.positions, key)
	s.values = slices.Delete(s.values, at, at+1)
	for i := at; i < len(s.values); i++ {
		s.positions[s.values[i]] = i
	}
	s.revision++
	return true
}

// Clear releases the selection's storage and reports an effective change.
func (s *SelectionState[K]) Clear() bool {
	changed := len(s.values) != 0
	s.values, s.positions = nil, nil
	if changed {
		s.revision++
	}
	return changed
}

// Has reports membership without enumerating selected values.
func (s *SelectionState[K]) Has(key K) bool {
	checkSelectionKey(key)
	_, exists := s.positions[key]
	return exists
}

func (s *SelectionState[K]) Len() int         { return len(s.values) }
func (s *SelectionState[K]) Revision() uint64 { return s.revision }

// Values returns an owned ordered snapshot. Mutating it cannot change s.
func (s *SelectionState[K]) Values() []K { return slices.Clone(s.values) }
