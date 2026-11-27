package nui_test

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestBorrowedNumericTableCallbacksAndFocus(t *testing.T) {
	type row struct {
		id   uint64
		name string
	}
	const high = uint64(1)<<63 + 17
	rows := []row{{0, "Zero"}, {high, "High"}, {^uint64(0), "Last"}}
	key := func(r row) uint64 { return r.id }
	var state nui.KeyedTableState[row, uint64]
	state.SetBorrowedRows(rows, key)
	var activated, selected []uint64
	notices := 0
	tester := ui.NewTester(func(c *ui.Context) {
		nui.Table(c.Key("typed"), nui.App(c, nui.AppConfig{}), &state, []nui.TableColumn[row]{{
			ID: "name", Title: "Name", Value: func(r row) string { return r.name }, Compare: func(a, b row) int { return cmp.Compare(a.name, b.name) },
		}}, nui.KeyedTableOptions[uint64]{
			Label:      "Typed rows",
			OnActivate: func(id uint64) { activated = append(activated, id) },
			OnChange: func() {
				selected = slices.Sorted(state.Selection.All())
				notices++
			},
		}).Width(300)
	}, 400, 360)
	click(t, tester, "Zero")
	tester.Key(0, ui.KeyEnter)
	if !slices.Equal(activated, []uint64{0}) || !slices.Equal(selected, []uint64{0}) {
		t.Fatalf("numeric zero callbacks: activated=%v selected=%v", activated, selected)
	}
	before := notices
	slices.Reverse(rows)
	state.SetBorrowedRows(rows, key)
	rows[2].name = "Updated zero"
	state.SetBorrowedRows(rows, key)
	tester.Frame()
	tester.Key(0, ui.KeyEnter)
	if activated[len(activated)-1] != 0 || notices != before || !tester.HasText("Updated zero") {
		t.Fatal("external reorder/value update changed focus, emitted selection, or showed stale values")
	}
	click(t, tester, "High")
	tester.Key(0, ui.KeyEnter)
	if activated[len(activated)-1] != high || !state.Selection.Has(high) {
		t.Fatal("high numeric identity was formatted or truncated")
	}
	bound := state.List.Selection
	state.Selection.Clear()
	tester.Frame()
	if bound != state.List.Selection {
		t.Fatal("selection clear changed the native bound address")
	}
	tester.Key(0, ui.KeyDown)
	if state.Selection.Len() != 1 {
		t.Fatal("selection compaction broke keyboard navigation")
	}
	before = len(activated)
	state.SetBorrowedRows(nil, key)
	tester.Frame()
	tester.Key(0, ui.KeyEnter)
	if len(activated) != before {
		t.Fatal("empty table activated an obsolete record")
	}
}

func TestBorrowedStringTableRetainsLegacyAPI(t *testing.T) {
	rows := []tableSnapshotRow{{"a", 2}, {"b", 1}}
	var state nui.TableState[tableSnapshotRow]
	state.SetBorrowedRows(rows, tableSnapshotKey)
	state.Sort = ui.SortOrder{Column: "value"}
	tester := ui.NewTester(func(c *ui.Context) {
		nui.Table[tableSnapshotRow](c, nui.App(c, nui.AppConfig{}), &state, []nui.TableColumn[tableSnapshotRow]{{
			ID: "value", Value: func(row tableSnapshotRow) string { return row.id }, Compare: func(a, b tableSnapshotRow) int { return cmp.Compare(a.value, b.value) },
		}}, nui.TableOptions{Label: "Borrowed strings"})
	}, 400, 300)
	tester.Frame()
	if state.RowID(0) != "b" || rows[0].id != "a" {
		t.Fatal("legacy string API failed borrowed sorting")
	}
}

func TestTableBuildCellUsesDisplayedIndexAndExistingContainer(t *testing.T) {
	rows := []tableSnapshotRow{{"a", 2}, {"b", 1}}
	var state nui.TableState[tableSnapshotRow]
	state.SetBorrowedRows(rows, tableSnapshotKey)
	state.Selection.Add("b")
	state.Sort = ui.SortOrder{Column: "value"}
	tester := ui.NewTester(func(c *ui.Context) {
		nui.Table(c, nui.App(c, nui.AppConfig{}), &state, []nui.TableColumn[tableSnapshotRow]{{
			ID: "value", Compare: func(a, b tableSnapshotRow) int { return cmp.Compare(a.value, b.value) },
			Cell: func(*ui.Context, nui.Theme, tableSnapshotRow, bool) { t.Fatal("BuildCell did not take precedence") },
			BuildCell: func(c *ui.Context, theme nui.Theme, element ui.Element, cell nui.TableCell[tableSnapshotRow]) {
				element.Background(theme.Elevated)
				ui.Text(c, fmt.Sprintf("%d:%s:%t", cell.Index, cell.Value.id, cell.Selected))
			},
		}}, nui.TableOptions{Label: "Custom cells"})
	}, 400, 300)
	tester.Frame()
	if !tester.HasText("0:b:true") || !tester.HasText("1:a:false") {
		t.Fatal("custom cell lost sorted index, value, or selected state")
	}
	state.Sort = ui.SortOrder{}
	tester.Frame()
	if !tester.HasText("0:a:false") || !tester.HasText("1:b:true") {
		t.Fatal("custom cell did not follow restored input order")
	}
}
