package nui_test

import (
	"cmp"
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type tableSnapshotRow struct {
	id    string
	value int
}

func tableSnapshotKey(row tableSnapshotRow) string { return row.id }

func TestTableSnapshotsUpdateValuesAndRestoreInputOrder(t *testing.T) {
	rows := []tableSnapshotRow{{"a", 2}, {"b", 1}, {"c", 3}}
	var state nui.TableState[tableSnapshotRow]
	state.SetRows(rows, tableSnapshotKey)
	state.Sort = ui.SortOrder{Column: "value"}
	state.Selection.Add("a")
	notices := 0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Table(c, nui.App(c, nui.AppConfig{}), &state, []nui.TableColumn[tableSnapshotRow]{{
			ID: "value", Title: "Value", Value: func(row tableSnapshotRow) string { return row.id },
			Compare: func(a, b tableSnapshotRow) int { return cmp.Compare(a.value, b.value) },
		}}, nui.TableOptions{Label: "Rows", OnChange: func() { notices++ }}).Width(300)
	}, 400, 300)
	if state.RowID(0) != "b" {
		t.Fatal("initial sort")
	}
	for _, value := range []int{0, 4, 0} {
		rows[0].value = value
		state.SetRows(rows, tableSnapshotKey)
		// Updating the caller's slice after SetRows must not change its copy.
		rows[0].value = 99
		tt.Frame()
		if value == 0 && state.RowID(0) != "a" || value == 4 && state.RowID(2) != "a" {
			t.Fatal("same-ID value update did not invalidate sorting")
		}
		if !state.Selection.Has("a") || notices != 0 {
			t.Fatal("external snapshot lost selection or emitted input notices")
		}
	}
	state.Sort = ui.SortOrder{}
	tt.Frame()
	for index, id := range []string{"a", "b", "c"} {
		if state.RowID(index) != id {
			t.Fatal("clearing sort did not restore the latest input order")
		}
	}
	state.SetRows(rows[1:], tableSnapshotKey)
	tt.Frame()
	_, present := tt.Find("a")
	if state.Selection.Len() != 0 || present {
		t.Fatal("removed row or its selection survived")
	}
}

func TestTableRejectedSnapshotLeavesPreviousRowsAndSelectionIntact(t *testing.T) {
	for _, failure := range []string{"duplicate", "empty", "key-panic", "nil-key"} {
		t.Run(failure, func(t *testing.T) {
			rows := []tableSnapshotRow{{"a", 1}, {"b", 2}, {"c", 3}}
			var state nui.TableState[tableSnapshotRow]
			state.SetRows(rows, tableSnapshotKey)
			state.Selection.Add("b")
			bad := []tableSnapshotRow{{"a", 11}, {"b", 12}, {"new", 13}}
			key := tableSnapshotKey
			switch failure {
			case "duplicate":
				bad[2].id = "a"
			case "empty":
				bad[2].id = ""
			case "key-panic":
				key = func(row tableSnapshotRow) string {
					if row.id == "new" {
						panic("key failed")
					}
					return row.id
				}
			case "nil-key":
				key = nil
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("invalid snapshot was accepted")
					}
				}()
				state.SetRows(bad, key)
			}()
			for i, want := range rows {
				if state.Row(i) != want || state.RowID(i) != want.id {
					t.Fatal("invalid snapshot partially overwrote the visible rows")
				}
			}
			if state.Selection.Len() != 1 || !state.Selection.Has("b") {
				t.Fatal("invalid snapshot changed selection")
			}
			// A failed validation must not poison a later unchanged-ID update.
			state.Selection.Add("missing")
			rows[1].value = 20
			state.SetRows(rows, tableSnapshotKey)
			if state.Row(1).value != 20 || state.Selection.Len() != 1 || !state.Selection.Has("b") {
				t.Fatal("snapshot or membership cache did not recover")
			}
			state.Selection.Add("also-missing")
			state.SetRows(rows, tableSnapshotKey)
			if state.Selection.Has("also-missing") {
				t.Fatal("unchanged IDs stopped pruning outside selection edits")
			}
		})
	}
}

func TestTableSnapshotStorageDoesNotAllocatePerUpdate(t *testing.T) {
	rows := make([]tableSnapshotRow, 5000)
	for i := range rows {
		rows[i] = tableSnapshotRow{strconv.Itoa(i), i}
	}
	var state nui.TableState[tableSnapshotRow]
	state.SetRows(rows, tableSnapshotKey)
	if allocations := testing.AllocsPerRun(20, func() {
		rows[0].value++
		state.SetRows(rows, tableSnapshotKey)
	}); allocations != 0 {
		t.Fatalf("same-ID update allocated %.0f times", allocations)
	}
	if allocations := testing.AllocsPerRun(20, func() {
		rows[0], rows[len(rows)-1] = rows[len(rows)-1], rows[0]
		state.SetRows(rows, tableSnapshotKey)
	}); allocations > 4 {
		t.Fatalf("reordered snapshot did not reuse validation storage: %.0f allocations", allocations)
	}
}

func TestTableGrowingSnapshotsAmortizeStorage(t *testing.T) {
	rows := make([]tableSnapshotRow, 5200)
	for i := range rows {
		rows[i] = tableSnapshotRow{strconv.Itoa(i), i}
	}
	var state nui.TableState[tableSnapshotRow]
	state.SetRows(rows[:5000], tableSnapshotKey)
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for count := 5001; count <= len(rows); count++ {
		state.SetRows(rows[:count], tableSnapshotKey)
	}
	runtime.ReadMemStats(&after)
	if bytes := after.TotalAlloc - before.TotalAlloc; bytes > 8<<20 {
		t.Fatalf("200 single-row appends allocated %d bytes", bytes)
	}
	if state.RowID(5199) != "5199" {
		t.Fatal("last appended row is missing")
	}
}

func BenchmarkTableSetRows(b *testing.B) {
	for _, count := range []int{5000, 50000, 100000} {
		for _, mode := range []string{"Stable", "Reorder", "Churn", "Grow", "FormatIDs"} {
			b.Run(fmt.Sprintf("Rows%d/%s", count, mode), func(b *testing.B) {
				rows := make([]tableSnapshotRow, count)
				for i := range rows {
					rows[i] = tableSnapshotRow{strconv.Itoa(i), i}
				}
				key := tableSnapshotKey
				if mode == "FormatIDs" {
					key = func(row tableSnapshotRow) string { return strconv.Itoa(row.value) }
				}
				var state nui.TableState[tableSnapshotRow]
				state.SetRows(rows, key)
				next := count
				b.ReportAllocs()
				for b.Loop() {
					switch mode {
					case "Stable":
						rows[0].value++
					case "Reorder":
						rows[0], rows[count-1] = rows[count-1], rows[0]
					case "Churn":
						copy(rows, rows[100:])
						for i := count - 100; i < count; i++ {
							rows[i] = tableSnapshotRow{strconv.Itoa(next), next}
							next++
						}
					case "Grow":
						rows = append(rows, tableSnapshotRow{strconv.Itoa(next), next})
						next++
					}
					state.SetRows(rows, key)
				}
				if state.Len() != len(rows) || state.Row(0) != rows[0] || state.Row(len(rows)-1) != rows[len(rows)-1] {
					b.Fatal("latest snapshot was not applied")
				}
			})
		}
	}
}
