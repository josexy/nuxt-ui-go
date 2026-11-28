package nui

import (
	"cmp"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
)

type borrowedRow struct {
	id    uint64
	value int
}

func borrowedKey(row borrowedRow) uint64 { return row.id }

func TestBorrowedTableOwnsOnlyIdentitiesAndSortOrder(t *testing.T) {
	rows := []borrowedRow{{0, 3}, {1 << 63, 1}, {^uint64(0), 2}}
	original := slices.Clone(rows)
	var state KeyedTableState[borrowedRow, uint64]
	if !state.SetBorrowedRows(rows, borrowedKey) {
		t.Fatal("first publication did not report new identities")
	}
	if &state.borrowed[0] != &rows[0] || state.rows != nil || state.ids != nil || state.pendingIDs != nil {
		t.Fatal("borrowed binding copied rows or retained validation storage")
	}
	columns := []TableColumn[borrowedRow]{{ID: "value", Compare: func(a, b borrowedRow) int { return cmp.Compare(a.value, b.value) }}}
	state.Sort = ui.SortOrder{Column: "value"}
	state.sort(columns, false)
	if state.RowID(0) != 1<<63 || state.RowID(2) != 0 || !slices.Equal(rows, original) {
		t.Fatal("local sorting changed input or numeric identities")
	}
	rows[0].value = 0
	if state.SetBorrowedRows(rows, borrowedKey) {
		t.Fatal("same-ID payload update changed input identities")
	}
	if state.Row(0).value != 0 || state.RowID(0) != 0 {
		t.Fatal("publication did not immediately expose updated input order")
	}
	state.sort(columns, false)
	if state.RowID(0) != 0 {
		t.Fatal("same-ID update did not invalidate local sorting")
	}
	state.Sort.Descending = true
	state.sort(columns, false)
	if state.RowID(0) != ^uint64(0) {
		t.Fatal("descending numeric sort lost a high identity")
	}
	state.sort(columns, true)
	if state.RowID(0) != 0 || state.borrowedOrder != nil {
		t.Fatal("manual sorting retained a permutation or changed input order")
	}
	state.sort(columns, false)
	state.Sort = ui.SortOrder{}
	state.sort(columns, false)
	if state.RowID(0) != 0 || state.borrowedOrder != nil {
		t.Fatal("clearing sort did not restore input order and release permutation")
	}
	state.SetRows(rows, borrowedKey)
	rows[0].value = 99
	if state.Row(0).value == 99 || state.borrowed != nil || state.borrowedIDs != nil {
		t.Fatal("copy mode retained the borrowed source")
	}
	state.SetBorrowedRows(rows, borrowedKey)
	state.SetBorrowedRows(nil, borrowedKey)
	if state.Len() != 0 || state.borrowed != nil || state.borrowedIDs != nil || rows[0].value != 99 {
		t.Fatal("release retained borrowed storage or cleared caller data")
	}
}

func TestBorrowedTableReusedArrayReconcilesSelectionAndShrinks(t *testing.T) {
	rows := make([]borrowedRow, 100000)
	for i := range rows {
		rows[i].id = uint64(i)
	}
	var state KeyedTableState[borrowedRow, uint64]
	state.SetBorrowedRows(rows, borrowedKey)
	for _, row := range rows {
		state.Selection.Add(row.id)
	}
	state.List.Selection = &state.Selection
	bound := state.List.Selection
	rows[0].id = ^uint64(0)
	if !state.SetBorrowedRows(rows[:2], borrowedKey) {
		t.Fatal("in-place ID replacement went unnoticed")
	}
	if state.Selection.Len() != 1 || !state.Selection.Has(1) || state.selectionPeak != 1 || cap(state.borrowedIDs) > 64 {
		t.Fatal("shrink retained removed choices or oversized identity storage")
	}
	state.Selection.Add(1234)
	state.SetBorrowedRows(rows[:2], borrowedKey)
	if state.Selection.Has(1234) {
		t.Fatal("unchanged IDs did not prune an outside selection edit")
	}
	state.SetBorrowedRows(nil, borrowedKey)
	if state.Selection.Len() != 0 || state.selectionPeak != 0 || bound != state.List.Selection {
		t.Fatal("release changed native binding or retained selection")
	}
}

func TestBorrowedTableStableUpdatesAllocateNoRowStorage(t *testing.T) {
	rows := make([]borrowedRow, 100000)
	for i := range rows {
		rows[i].id = uint64(i)
	}
	var state KeyedTableState[borrowedRow, uint64]
	state.SetBorrowedRows(rows, borrowedKey)
	if got := testing.AllocsPerRun(20, func() {
		rows[0].value++
		state.SetBorrowedRows(rows, borrowedKey)
	}); got != 0 {
		t.Fatalf("stable borrowed update allocated %g times", got)
	}
	// Reconciling a selection whose IDs are all still bound needs no scratch set.
	for _, row := range rows {
		state.Selection.Add(row.id)
	}
	if got := testing.AllocsPerRun(5, func() {
		rows[0].value++
		state.SetBorrowedRows(rows, borrowedKey)
	}); got != 0 {
		t.Fatalf("stable borrowed update with every row chosen allocated %g times", got)
	}
	if state.Selection.Len() != len(rows) {
		t.Fatalf("stable borrowed update changed the selection to %d rows", state.Selection.Len())
	}
}

func TestBorrowedTableRetainedStorageExcludesCallerRows(t *testing.T) {
	rows := make([]borrowedRow, 100000)
	for i := range rows {
		rows[i].id = uint64(i)
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&before)
	var state KeyedTableState[borrowedRow, uint64]
	state.SetBorrowedRows(rows, borrowedKey)
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&after)
	if retained := int64(after.HeapAlloc) - int64(before.HeapAlloc); retained > 2<<20 {
		t.Fatalf("100k borrowed IDs retained %d B beyond source rows", retained)
	}
	runtime.KeepAlive(rows)
	runtime.KeepAlive(&state)
}

func TestKeyedCopyValidationPreservesPriorBorrowedData(t *testing.T) {
	rows := []borrowedRow{{0, 1}, {^uint64(0), 2}}
	var state KeyedTableState[borrowedRow, uint64]
	state.SetBorrowedRows(rows, borrowedKey)
	state.Selection.Add(0)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("copy mode accepted duplicate numeric IDs")
			}
		}()
		state.SetRows([]borrowedRow{{8, 3}, {8, 4}}, borrowedKey)
	}()
	if state.RowID(0) != 0 || state.RowID(1) != ^uint64(0) || !state.Selection.Has(0) {
		t.Fatal("failed copying publication changed the previous borrowed state")
	}
}

func TestTableShrinkingReleasesRemovedValues(t *testing.T) {
	values := make([]*string, 5000)
	for i := range values {
		value := strconv.Itoa(i)
		values[i] = &value
	}
	key := func(value *string) string { return *value }
	var state TableState[*string]
	state.SetRows(values, key)
	state.SetRows(values[:4000], key)
	for _, record := range state.rows[:cap(state.rows)][len(state.rows):] {
		if record.id != "" || record.value != nil {
			t.Fatal("truncated capacity retains removed row data")
		}
	}
	for _, id := range state.pendingIDs[:cap(state.pendingIDs)] {
		if id != "" {
			t.Fatal("ID scratch storage retains row identities")
		}
	}
	state.SetRows(values[:2], key)
	if cap(state.rows) > 64 || cap(state.pendingIDs) > 64 || state.idsCapacity > 64 {
		t.Fatal("a small table retained storage for the large snapshot")
	}
	if len(state.ids) != 2 {
		t.Fatal("membership cache retained removed IDs")
	}
	state.SetRows(nil, key)
	if state.rows != nil || state.pendingIDs != nil || state.ids != nil {
		t.Fatal("empty table retained row storage")
	}
}
