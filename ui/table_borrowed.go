package nui

import (
	"cmp"
	"slices"

	"github.com/egoist/mygo/ui"
)

// SetBorrowedRows binds caller-owned rows without copying their values. It keeps
// one identity snapshot, and reports whether input identities or order changed.
// Call it outside rendering when data changes, including same-ID value changes.
//
// Unlike SetRows, this is a trusted-data binding: the caller must supply unique,
// stable keys (nonempty for strings), and a pure key function. Validate untrusted
// data before binding it. No full membership map or validation scratch is kept.
// The slice must remain valid until the next binding; update it only on the UI
// thread between builds, then rebind before rendering. The table never changes
// the caller's array. Local sorting uses a separate permutation; manual sorting
// reads the supplied order directly. Binding nil releases all borrowed data.
func (s *KeyedTableState[T, K]) SetBorrowedRows(rows []T, key func(T) K) bool {
	if key == nil {
		panic("nui: Table requires row IDs")
	}
	activeID, hasActive := s.activeRowID()
	n := len(rows)
	changed := !s.borrowing || n != len(s.borrowedIDs)
	ids := s.borrowedIDs
	if n == 0 || cap(ids) > max(64, 4*n) {
		ids = nil
	}
	clear(ids[min(n, len(ids)):])
	ids = slices.Grow(ids[:0], n)[:n]
	for i, row := range rows {
		id := key(row)
		if ids[i] != id {
			changed = true
		}
		ids[i] = id
	}
	s.rows, s.pendingIDs, s.ids = nil, nil, nil
	s.idsCapacity, s.idsValid = 0, false
	s.borrowed, s.borrowedIDs, s.borrowing = rows, ids, true
	// Publish input order immediately; a later local sort can reuse its previous
	// permutation storage without mutating the caller's array.
	s.inputOrder = true
	if n == 0 || cap(s.borrowedOrder) > max(64, 4*n) {
		s.borrowedOrder = nil
	}
	s.revision++
	s.selectionPeak = max(s.selectionPeak, s.Selection.Len())
	// Usually every chosen ID is still bound: count them without allocating,
	// stopping once all are found, and look for the missing ones only then.
	found, chosen := 0, s.Selection.Len()
	for _, id := range ids {
		if found == chosen {
			break
		}
		if s.Selection.Has(id) {
			found++
		}
	}
	if found < chosen {
		missing := make(map[K]struct{}, chosen)
		for id := range s.Selection.All() {
			missing[id] = struct{}{}
		}
		for _, id := range ids {
			delete(missing, id)
		}
		for id := range missing {
			s.Selection.Remove(id)
		}
	}
	s.compactSelection()
	s.restoreActiveRow(activeID, hasActive)
	return changed
}

func (s *KeyedTableState[T, K]) sortBorrowed(orders []ui.SortOrder, columns map[string]TableColumn[T], manual bool) {
	if manual || len(orders) == 0 {
		s.borrowedOrder = nil
		return
	}
	n := len(s.borrowed)
	order := s.borrowedOrder
	if cap(order) > max(64, 4*n) {
		order = nil
	}
	order = slices.Grow(order[:0], n)[:n]
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int {
		for _, order := range orders {
			left, right := s.borrowed[a], s.borrowed[b]
			if order.Descending {
				left, right = right, left
			}
			if result := columns[order.Column].Compare(left, right); result != 0 {
				return result
			}
		}
		return cmp.Compare(a, b)
	})
	s.borrowedOrder = order
	s.inputOrder = false
}

// ClearSelection releases all chosen identities without changing the native
// bound address, cursor or range anchor. It does not emit user-input callbacks.
func (s *KeyedTableState[T, K]) ClearSelection() {
	s.Selection = ui.Selection[K]{}
	s.selectionPeak = 0
}

// Drop oversized selection maps without changing the address bound to MyGo or
// disturbing its cursor, range anchor and current gesture.
func (s *KeyedTableState[T, K]) compactSelection() {
	n := s.Selection.Len()
	if n == 0 {
		s.ClearSelection()
	} else if s.selectionPeak > max(64, 4*n) {
		var compact ui.Selection[K]
		for id := range s.Selection.All() {
			compact.Add(id)
		}
		s.Selection = compact
		s.selectionPeak = n
	} else {
		s.selectionPeak = max(s.selectionPeak, n)
	}
}
