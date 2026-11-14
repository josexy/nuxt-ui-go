package nui

import (
	"strconv"
	"testing"
)

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
