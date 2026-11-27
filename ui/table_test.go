package nui_test

import (
	"bytes"
	"cmp"
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type tableFile struct {
	ID, Name, Modified string
	Size               int
}

var tableFiles = []tableFile{{"a", "Alpha", "Mon", 3}, {"b", "Beta", "Tue", 2}, {"c", "Gamma", "Wed", 1}}

type tableCounts struct{ change, sort, activate int }

type tableHarness struct {
	*ui.Tester
	t         *testing.T
	state     *nui.TableState[tableFile]
	counts    tableCounts
	activated []string
}

func newTableHarness(t *testing.T) *tableHarness {
	h := &tableHarness{t: t, state: &nui.TableState[tableFile]{}}
	h.state.SetRows(tableFiles, func(f tableFile) string { return f.ID })
	columns := []nui.TableColumn[tableFile]{
		{ID: "name", Title: "Name", Width: 160, MinWidth: 40, Value: func(f tableFile) string { return f.Name }},
		{ID: "size", Title: "Size", Width: 100, Value: func(f tableFile) string { return strings.Repeat("■", f.Size) }, Compare: func(a, b tableFile) int { return cmp.Compare(a.Size, b.Size) }},
		{ID: "modified", Title: "Modified", Width: 120, Value: func(f tableFile) string { return f.Modified }},
		{ID: "lock", Title: "Lock", Width: 70, Fixed: true, Value: func(tableFile) string { return "no" }},
	}
	h.Tester = ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Table(c.Key("files"), th, h.state, columns, nui.TableOptions{
			Label:        "Files",
			OnChange:     func() { h.counts.change++ },
			OnSortChange: func([]ui.SortOrder) { h.counts.sort++ },
			OnActivate:   func(id string) { h.counts.activate++; h.activated = append(h.activated, id) },
		})
	}, 700, 360)
	return h
}

// header is the box of a column's header, as assistive technology sees it.
func (h *tableHarness) header(title string) ui.Rect {
	h.t.Helper()
	for _, info := range h.AccessAll() {
		if info.Role == ui.RoleColumnHeader && strings.HasPrefix(info.Name, title) {
			return info.Bounds
		}
	}
	h.t.Fatalf("no %q header among %v", title, accessSummary(h.AccessAll()))
	return ui.Rect{}
}

func (h *tableHarness) expect(step string, want tableCounts) {
	h.t.Helper()
	if h.counts != want {
		h.t.Fatalf("%s: callbacks %+v, want %+v (columns %+v)", step, h.counts, want, h.state.List.Columns)
	}
}

// drag presses at x, moves through the offsets and lets go at the last.
func (h *tableHarness) drag(x, y float32, offsets ...float32) {
	h.Press(x, y)
	for _, d := range offsets {
		h.Move(x+d, y)
	}
	h.Release(x+offsets[len(offsets)-1], y)
}

func (h *tableHarness) order() []string {
	order := h.state.List.Columns.Order
	if len(order) == 0 {
		return []string{"name", "size", "modified", "lock"}
	}
	return order
}

func TestTableColumnGesturesReportOneChangeEach(t *testing.T) {
	h := newTableHarness(t)
	h.Frame()
	h.expect("first builds", tableCounts{})

	name := h.header("Name")
	y := name.Y + name.H/2
	edge := name.X + name.W - 2
	h.Press(edge, y)
	h.Move(edge+30, y)
	h.Move(edge+60, y)
	h.expect("resizing, held", tableCounts{})
	h.Release(edge+60, y)
	h.expect("resized", tableCounts{change: 1})
	if w := h.state.List.Columns.Widths["name"]; w != 220 {
		t.Fatalf("Name is %v wide, want 220", w)
	}

	// Modified dragged before Name.
	modified := h.header("Modified")
	start := modified.X + 20
	distance := start - (h.header("Name").X + 5)
	h.drag(start, y, -5, -distance/3, -2*distance/3, -distance)
	h.expect("moved", tableCounts{change: 2})
	if got := h.order(); !slices.Equal(got[:3], []string{"modified", "name", "size"}) {
		t.Fatalf("the order is %q", got)
	}

	// A header let go where it started changes nothing, and a click still sorts.
	size := h.header("Size")
	h.drag(size.X+20, y, -5, -10, 0)
	h.expect("dragged in place", tableCounts{change: 2})
	h.ClickAt(size.X+20, y)
	h.expect("sorted by a click", tableCounts{change: 3, sort: 1})
	for range 3 {
		h.Frame()
	}
	h.expect("rebuilds after sorting", tableCounts{change: 3, sort: 1})

	// A double click on Name's edge fits it to its cells.
	name = h.header("Name")
	edge = name.X + name.W - 2
	h.ClickAt(edge, y)
	h.ClickAt(edge, y)
	h.expect("fitted", tableCounts{change: 4, sort: 1})
	if w := h.state.List.Columns.Widths["name"]; w >= 220 || w <= 0 {
		t.Fatalf("fitting left Name %v wide", w)
	}

	// The fixed column neither resizes nor moves.
	lock := h.header("Lock")
	h.drag(lock.X+lock.W-1, y, 20, 40)
	h.drag(lock.X+20, y, -5, -60, -200)
	h.Frame()
	if got := h.header("Lock"); got.W != lock.W || h.order()[3] != "lock" {
		t.Fatalf("the fixed column changed: %v, order %q", got, h.order())
	}
	h.expect("fixed column", tableCounts{change: 4, sort: 1})
}

func TestTableActiveRowFollowsItsRecord(t *testing.T) {
	h := newTableHarness(t)
	click(t, h.Tester, "Beta")
	h.expect("choosing Beta", tableCounts{change: 1})
	size := h.header("Size")
	h.ClickAt(size.X+20, size.Y+size.H/2)
	if got := []string{h.state.RowID(0), h.state.RowID(1), h.state.RowID(2)}; !slices.Equal(got, []string{"c", "b", "a"}) {
		t.Fatalf("sorted order %q", got)
	}
	h.Key(0, ui.KeyEnter)
	if h.counts.activate != 1 || !slices.Equal(h.activated, []string{"b"}) {
		t.Fatalf("Enter after sorting activated %q", h.activated)
	}

	// A record that moves: Alpha goes from last to first.
	click(t, h.Tester, "Alpha")
	h.ClickAt(size.X+20, size.Y+size.H/2)
	if h.state.RowID(0) != "a" {
		t.Fatalf("descending order starts with %q", h.state.RowID(0))
	}
	h.Key(0, ui.KeyEnter)
	if !slices.Equal(h.activated, []string{"b", "a"}) {
		t.Fatalf("Enter after re-sorting activated %q", h.activated)
	}

	// The chosen record leaves: nothing is active.
	click(t, h.Tester, "Beta")
	h.state.SetRows([]tableFile{tableFiles[0], tableFiles[2]}, func(f tableFile) string { return f.ID })
	h.Frame()
	h.Key(0, ui.KeyEnter)
	if !slices.Equal(h.activated, []string{"b", "a"}) {
		t.Fatalf("Enter with the chosen record removed activated %q", h.activated)
	}

	// A double click activates once.
	alpha, ok := h.Find("Alpha")
	if !ok {
		t.Fatal("no Alpha row")
	}
	h.ClickAt(alpha.X+alpha.W/2, alpha.Y+alpha.H/2)
	h.ClickAt(alpha.X+alpha.W/2, alpha.Y+alpha.H/2)
	if !slices.Equal(h.activated, []string{"b", "a", "a"}) {
		t.Fatalf("double click activated %q", h.activated)
	}
}

func TestTableActiveRowSurvivesBatchedSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name string
		last []tableFile
		want []string
	}{
		{"remove another record", []tableFile{tableFiles[1], tableFiles[2]}, []string{"b"}},
		{"remove the active record", []tableFile{tableFiles[0], tableFiles[2]}, nil},
		{"clear all records", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newTableHarness(t)
			click(t, h.Tester, "Beta")
			before := h.counts
			key := func(f tableFile) string { return f.ID }
			// No build separates these snapshots: the chosen record moves
			// before a later update removes a record from its old position.
			h.state.SetRows([]tableFile{tableFiles[1], tableFiles[0], tableFiles[2]}, key)
			h.state.SetRows(tc.last, key)
			h.Frame()
			h.Frame()
			h.expect("external snapshots", before)
			if h.state.Selection.Has("b") != (len(tc.want) > 0) {
				t.Fatal("selection did not follow Beta's membership")
			}
			h.Key(0, ui.KeyEnter)
			if !slices.Equal(h.activated, tc.want) {
				t.Fatalf("Enter activated %q, want %q", h.activated, tc.want)
			}
		})
	}
}

// The documented way to keep a column layout: save List.Columns from
// OnChange, and restore it before the first build.
func TestTableColumnLayoutPersistsThroughOnChange(t *testing.T) {
	type settings struct{ Columns ui.TableLayout }
	var saved settings
	saves := 0
	build := func(app *nui.TableState[tableFile]) *ui.Tester {
		columns := []nui.TableColumn[tableFile]{
			{ID: "name", Title: "Name", Width: 160, Value: func(f tableFile) string { return f.Name }},
			{ID: "modified", Title: "Modified", Width: 120, Value: func(f tableFile) string { return f.Modified }},
		}
		return ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Table(c.Key("files"), th, app, columns, nui.TableOptions{
				Label: "Files",
				OnChange: func() {
					saved.Columns = app.List.Columns
					saves++
				},
			})
		}, 600, 300)
	}
	var first nui.TableState[tableFile]
	first.SetRows(tableFiles, func(f tableFile) string { return f.ID })
	h := &tableHarness{Tester: build(&first), t: t, state: &first}
	name := h.header("Name")
	edge, y := name.X+name.W-2, name.Y+name.H/2
	h.drag(edge, y, 20, 40)
	if saves != 1 || saved.Columns.Widths["name"] != 200 {
		t.Fatalf("saved %d times: %+v", saves, saved.Columns)
	}

	var second nui.TableState[tableFile]
	second.SetRows(tableFiles, func(f tableFile) string { return f.ID })
	second.List.Columns = saved.Columns
	restored := &tableHarness{Tester: build(&second), t: t, state: &second}
	if w := restored.header("Name").W; w != 200 {
		t.Fatalf("restored Name is %v wide", w)
	}
}

func TestTableRowHeadersKeepNativeCellSelection(t *testing.T) {
	var state nui.KeyedTableState[int, int]
	state.SetRows([]int{0, 1, 2}, func(v int) int { return v })
	state.List.CellSelect, state.List.FollowEnd, state.List.RowHeight = true, true, 42
	state.List.NearEnd = 7
	changes, sorts, headers := 0, 0, 0
	state.List.HeaderCell = func(e ui.Element, col int) { headers++; e.Description(fmt.Sprintf("Custom header %d", col)) }
	columns := []nui.TableColumn[int]{
		{ID: "row", Title: "#", Width: 60, RowHeader: true, Sortable: nui.Set(true), Value: func(v int) string { return fmt.Sprintf("Row %d", v) }},
		{ID: "name", Title: "Name", Width: 140, Value: func(v int) string { return fmt.Sprintf("Name %d", v) }},
		{ID: "value", Title: "Value", Width: 140, Value: func(v int) string { return fmt.Sprintf("Value %d", v) }},
	}
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Table(c.Key("table"), nui.App(c, nui.AppConfig{DisableAnimations: true}), &state, columns,
			nui.KeyedTableOptions[int]{Label: "Rows", OnChange: func() { changes++ }, OnSortChange: func([]ui.SortOrder) { sorts++ }}).Width(400)
	}, 450, 400)
	click(t, tt, "Name 1")
	if state.List.Cells.All || state.List.Cells.Cursor != "name" || !state.Selection.Has(1) {
		t.Fatal("ordinary cell did not select its column")
	}
	header, ok := tt.AccessByRole(ui.RoleRowHeader, "Row 0")
	if !ok {
		t.Fatal("row header has no native row-header role")
	}
	tt.ClickAt(header.Bounds.X+header.Bounds.W/2, header.Bounds.Y+header.Bounds.H/2)
	if !state.List.Cells.All || state.List.Cells.Cursor == "row" || !slices.Equal(slices.Sorted(state.Selection.All()), []int{0}) || !slices.Equal(state.List.SelectedColumns(), []string{"name", "value"}) {
		t.Fatalf("row header lost whole-row selection: cells=%+v selection=%v", state.List.Cells, slices.Sorted(state.Selection.All()))
	}
	click(t, tt, "#")
	if state.Selection.Len() != 3 || state.Sort != (ui.SortOrder{}) || sorts != 0 || changes != 3 {
		t.Fatalf("row-header heading sorted or emitted wrong changes: selection=%d sort=%+v events=%d/%d", state.Selection.Len(), state.Sort, changes, sorts)
	}
	state.Sort = ui.SortOrder{Column: "row", Descending: true}
	tt.Frame()
	if state.Sort != (ui.SortOrder{}) || state.RowID(0) != 0 || changes != 3 {
		t.Fatal("external row-header sort was not silently discarded")
	}
	if !state.List.CellSelect || !state.List.FollowEnd || state.List.RowHeight != 42 || state.List.NearEnd != 7 || state.List.HeaderCell == nil || headers == 0 {
		t.Fatal("rendering overwrote native ListState configuration")
	}
}

func TestTablePreservesNativeHeaderAndCellMenus(t *testing.T) {
	var state nui.TableState[string]
	state.SetRows([]string{"Alpha", "Beta"}, func(v string) string { return v })
	header, row, column := -1, -1, -1
	state.List.HeaderMenu = func(m *ui.Menu, col int) {
		if m.Item("Header action").Chosen() {
			header = col
		}
	}
	state.List.CellMenu = func(m *ui.Menu, r, col int) {
		if m.Item("Cell action").Chosen() {
			row, column = r, col
		}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Table(c.Key("table"), nui.App(c, nui.AppConfig{DisableAnimations: true}), &state,
			[]nui.TableColumn[string]{{ID: "name", Title: "Name", Width: 160, Value: func(v string) string { return v }}, {ID: "extra", Title: "Extra", Width: 160, Value: func(v string) string { return "Extra " + v }}}, nui.TableOptions{}).Width(350)
	}, 450, 350)
	if err := tt.RightClick("Extra"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Header action"); err != nil {
		t.Fatal(err)
	}
	if err := tt.RightClick("Extra Beta"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Cell action"); err != nil {
		t.Fatal(err)
	}
	if header != 1 || row != 1 || column != 1 || !state.Selection.Has("Beta") {
		t.Fatalf("native menu mapping: header=%d cell=%d/%d selected=%v", header, row, column, state.Selection.Has("Beta"))
	}
}

func TestPlainTableCellsMatchDecoratedRendering(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, scale := range []float32{1, 1.25, 2} {
			for _, align := range []ui.Align{ui.Start, ui.Center, ui.End} {
				t.Run(fmt.Sprintf("dark=%v/scale=%g/align=%d", dark, scale, align), func(t *testing.T) {
					var state nui.TableState[string]
					state.SetRows([]string{"A short cell", "中文与 café — a long line that exceeds the column width"}, func(v string) string { return v })
					decorated := false
					var styles nui.Styles
					tt := ui.NewTester(func(c *ui.Context) {
						th := nui.Light()
						if dark {
							th = nui.Dark()
						}
						th.TextScale = scale
						th = nui.App(c, nui.AppConfig{Mode: nui.LightMode, Light: th, DisableAnimations: true})
						o := nui.TableOptions{Label: "Rows", Decoration: nui.Decoration{UI: styles}}
						if decorated {
							o.Style = func(nui.Slot, ui.Element) {}
						}
						nui.Table(c.Key("table"), th, &state, []nui.TableColumn[string]{{ID: "text", Title: "Text", Width: 190, Align: align, Value: func(v string) string { return v }}}, o).Width(220)
					}, 300, 400)
					tt.SetScale(1.25)
					plain := append([]byte(nil), tt.Image().Pix...)
					decorated = true
					tt.Frame()
					if !bytes.Equal(plain, tt.Image().Pix) {
						t.Fatal("plain value cells changed pixels versus the decorated path")
					}
					// Label styling alone keeps the direct text, styled as decorated text.
					styles, decorated = nui.Styles{nui.Label: {Foreground: nui.Set(ui.RGB(220, 0, 0)), Weight: nui.Set(700)}}, false
					tt.Frame()
					label := append([]byte(nil), tt.Image().Pix...)
					if bytes.Equal(plain, label) || redPixels(tt.Image(), ui.Rect{W: 300, H: 400}) == 0 {
						t.Fatal("a Label override did not style direct value cells")
					}
					decorated = true
					tt.Frame()
					if !bytes.Equal(label, tt.Image().Pix) {
						t.Fatal("Label-styled value cells changed pixels versus the decorated path")
					}
					styles, decorated = nui.Styles{nui.Label: {Foreground: nui.Set(ui.RGB(220, 0, 0))}, nui.Item: {PaddingY: nui.Set(float32(3))}}, false
					tt.Frame()
					if bytes.Equal(plain, tt.Image().Pix) || redPixels(tt.Image(), ui.Rect{W: 300, H: 400}) == 0 {
						t.Fatal("adding a part override did not restore decorated rendering")
					}
				})
			}
		}
	}
}

func TestTableCursorSurvivesSameIndexAndMovedRebindings(t *testing.T) {
	for _, borrowed := range []bool{false, true} {
		t.Run(fmt.Sprint(borrowed), func(t *testing.T) {
			var state nui.KeyedTableState[int, int]
			set := func(rows []int) {
				if borrowed {
					state.SetBorrowedRows(rows, func(v int) int { return v })
				} else {
					state.SetRows(rows, func(v int) int { return v })
				}
			}
			set([]int{0, 1, 2})
			var activated []int
			tt := ui.NewTester(func(c *ui.Context) {
				nui.Table(c.Key("table"), nui.App(c, nui.AppConfig{DisableAnimations: true}), &state,
					[]nui.TableColumn[int]{{ID: "id", Title: "ID", Value: func(v int) string { return fmt.Sprintf("Record %d", v) }}},
					nui.KeyedTableOptions[int]{OnActivate: func(id int) { activated = append(activated, id) }}).Width(200)
			}, 300, 350)
			click(t, tt, "Record 0")
			set([]int{0, 2, 1}) // same index, including identity zero
			set([]int{2, 1, 0}) // another update before a build moves that identity
			tt.Frame()
			tt.Key(0, ui.KeyEnter)
			set([]int{2, 1})
			tt.Frame()
			tt.Key(0, ui.KeyEnter)
			if !slices.Equal(activated, []int{0}) {
				t.Fatalf("cursor activated %v", activated)
			}
		})
	}
}

func TestTableCursorChangeKeepsSelection(t *testing.T) {
	var state nui.TableState[string]
	state.SetRows([]string{"a", "b", "c"}, func(v string) string { return v })
	cursor := 0
	state.List.Selected = &cursor
	state.Selection.Add("a")
	changes := 0
	columns := []nui.TableColumn[string]{{ID: "id", Title: "ID", Width: 200, Value: func(v string) string { return v }}}
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Table(c.Key("table"), nui.App(c, nui.AppConfig{DisableAnimations: true}), &state, columns, nui.TableOptions{
			OnChange: func() { changes++ },
		}).Width(300)
	}, 500, 400)
	tt.Key(0, ui.KeyTab)
	changes = 0
	tt.Key(ui.Ctrl, ui.KeyDown)
	if cursor != 1 || state.Selection.Len() != 1 || !state.Selection.Has("a") || changes != 1 {
		t.Fatalf("cursor=%d selected=%v combined=%d", cursor, state.Selection.Has("a"), changes)
	}
}

// RowHeight, EndReached and NearEnd are ui.ListState fields that Table hands
// to the native table through TableState.List; these tests hold Table to them.

type pageRow struct{ ID, Name string }

// pageRows returns rows "Row 00" to "Row n-1", or the reverse with reversed.
func pageRows(n int, reversed bool) []pageRow {
	rows := make([]pageRow, n)
	for i := range rows {
		v := i
		if reversed {
			v = n - 1 - i
		}
		rows[i] = pageRow{strconv.Itoa(v), fmt.Sprintf("Row %02d", v)}
	}
	return rows
}

func pageRowID(r pageRow) string { return r.ID }

var pageColumns = []nui.TableColumn[pageRow]{{ID: "name", Title: "Name", Width: 200, Value: func(r pageRow) string { return r.Name }, Compare: func(a, b pageRow) int { return cmp.Compare(a.Name, b.Name) }}}

type pageTable struct {
	*ui.Tester
	t      *testing.T
	state  *nui.TableState[pageRow]
	calls  []int
	counts tableCounts
}

// newPageTable shows rows in a Table named "Rows" with RowHeight rowHeight;
// height 0 keeps Table's default. With loadMore, it counts EndReached calls
// with NearEnd 5.
func newPageTable(t *testing.T, rows []pageRow, height, rowHeight float32, loadMore bool) *pageTable {
	h := &pageTable{t: t, state: &nui.TableState[pageRow]{}}
	h.state.SetRows(rows, pageRowID)
	h.state.List.RowHeight = rowHeight
	if loadMore {
		h.state.List.NearEnd = 5
		h.state.List.EndReached = func(last int) { h.calls = append(h.calls, last) }
	}
	h.Tester = ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Table(c.Key("rows"), th, h.state, pageColumns, nui.TableOptions{
			Label:        "Rows",
			Height:       height,
			OnChange:     func() { h.counts.change++ },
			OnSortChange: func([]ui.SortOrder) { h.counts.sort++ },
			OnActivate:   func(string) { h.counts.activate++ },
		})
	}, 500, 400)
	h.Frame()
	return h
}

func (h *pageTable) box() ui.Rect {
	h.t.Helper()
	info, ok := h.AccessByRole(ui.RoleTable, "Rows")
	if !ok {
		h.t.Fatalf("no Rows table among %v", accessSummary(h.AccessAll()))
	}
	return info.Bounds
}

func (h *pageTable) header() ui.Rect {
	h.t.Helper()
	for _, info := range h.AccessAll() {
		if info.Role == ui.RoleColumnHeader && strings.HasPrefix(info.Name, "Name") {
			return info.Bounds
		}
	}
	h.t.Fatalf("no Name header among %v", accessSummary(h.AccessAll()))
	return ui.Rect{}
}

// rows are the body rows the table built, in order, including those built
// past its bottom edge.
func (h *pageTable) rows() []ui.AccessInfo {
	var rows []ui.AccessInfo
	for _, info := range h.AccessAll() {
		if info.Role == ui.RoleRow && strings.HasPrefix(info.Name, "Row ") {
			rows = append(rows, info)
		}
	}
	if len(rows) == 0 {
		h.t.Fatalf("no body rows among %v", accessSummary(h.AccessAll()))
	}
	return rows
}

// visible are the rows whose top is inside the table.
func (h *pageTable) visible() []ui.AccessInfo {
	box := h.box()
	var rows []ui.AccessInfo
	for _, r := range h.rows() {
		if r.Bounds.Y+r.Bounds.H > box.Y && r.Bounds.Y < box.Y+box.H {
			rows = append(rows, r)
		}
	}
	return rows
}

// lastVisible is the index of the last row in view, as its name gives it.
func (h *pageTable) lastVisible() int {
	h.t.Helper()
	rows := h.visible()
	n, err := strconv.Atoi(strings.TrimPrefix(rows[len(rows)-1].Name, "Row "))
	if err != nil {
		h.t.Fatal(err)
	}
	return n
}

// scrollDown scrolls as the user does, 100 DIPs at a time, until the row at
// index last is in view. Each new EndReached call must name the last row in
// view as that frame laid it out.
func (h *pageTable) scrollDown(last int) {
	h.t.Helper()
	box := h.box()
	for range 100 {
		if h.lastVisible() == last {
			return
		}
		before := len(h.calls)
		h.Scroll(box.X+box.W/2, box.Y+box.H/2, 0, 100)
		if len(h.calls) > before {
			if got, want := h.calls[len(h.calls)-1], h.lastVisible(); got != want {
				h.t.Fatalf("EndReached(%d), but row %d is the last in view", got, want)
			}
		}
	}
	h.t.Fatalf("scrolling never showed row %d", last)
}

func (h *pageTable) expectCalls(step string, want int) {
	h.t.Helper()
	if len(h.calls) != want {
		h.t.Fatalf("%s: EndReached calls %v, want %d", step, h.calls, want)
	}
}

func TestTableRowHeightSizesBodyRows(t *testing.T) {
	type layout struct {
		header, first ui.Rect
		visible       int
	}
	got := map[float32]layout{}
	for _, height := range []float32{0, 24, 48} {
		h := newPageTable(t, pageRows(50, false), 300, height, false)
		rows := h.rows()
		want := height
		if want == 0 {
			want = rows[0].Bounds.H
		}
		for i, r := range rows {
			if r.Bounds.H != want {
				t.Fatalf("RowHeight %v: %s is %v high, want %v", height, r.Name, r.Bounds.H, want)
			}
			if i > 0 && r.Bounds.Y-rows[i-1].Bounds.Y != want {
				t.Fatalf("RowHeight %v: %s is %v below the row above, want %v", height, r.Name, r.Bounds.Y-rows[i-1].Bounds.Y, want)
			}
		}
		// The cell holds its text in the middle of the row.
		text, ok := h.Find(rows[0].Name)
		if !ok {
			t.Fatalf("RowHeight %v: no %s text", height, rows[0].Name)
		}
		if d := (text.Y + text.H/2) - (rows[0].Bounds.Y + rows[0].Bounds.H/2); d < -1 || d > 1 {
			t.Fatalf("RowHeight %v: text %v is %v off the middle of row %v", height, text, d, rows[0].Bounds)
		}
		got[height] = layout{h.header(), rows[0].Bounds, len(h.visible())}
	}
	base := got[0]
	if gap := base.first.Y - (base.header.Y + base.header.H); gap < 0 || gap > 1 {
		t.Fatalf("the first row starts %v below the header %v", gap, base.header)
	}
	for _, height := range []float32{24, 48} {
		if got[height].header != base.header || got[height].first.Y != base.first.Y {
			t.Fatalf("RowHeight %v moved the header to %v and the first row to %v, want %v and %v", height, got[height].header, got[height].first.Y, base.header, base.first.Y)
		}
	}
	if !(got[24].visible > base.visible && base.visible > got[48].visible) {
		t.Fatalf("rows in view for RowHeight 24/0/48: %d/%d/%d", got[24].visible, base.visible, got[48].visible)
	}
}

func TestTableRowHeightSortKeepsRowsAndCallbacks(t *testing.T) {
	h := newPageTable(t, pageRows(50, true), 300, 24, false)
	if first := h.rows()[0].Name; first != "Row 49" {
		t.Fatalf("unsorted first row %q", first)
	}
	header := h.header()
	h.ClickAt(header.X+header.W/2, header.Y+header.H/2)
	for range 3 {
		h.Frame()
	}
	// The list keeps the row that was at its top in view, so check the order
	// rather than which row comes first.
	rows := h.rows()
	for i, r := range rows {
		if r.Bounds.H != 24 {
			t.Fatalf("after sorting %s is %v high", r.Name, r.Bounds.H)
		}
		if i > 0 && r.Name <= rows[i-1].Name {
			t.Fatalf("after sorting %s follows %s", r.Name, rows[i-1].Name)
		}
	}
	if h.counts != (tableCounts{change: 1, sort: 1}) {
		t.Fatalf("callbacks %+v, want one change and one sort", h.counts)
	}
}

func TestTableEndReachedLoadsOncePerApproach(t *testing.T) {
	h := newPageTable(t, pageRows(40, false), 0, 0, true)
	if n := len(h.visible()); n < 8 || n > 10 {
		t.Fatalf("%d rows in view, want about 9", n)
	}
	h.expectCalls("shown", 0)

	h.scrollDown(39)
	h.expectCalls("scrolled to the end", 1)
	if h.calls[0] < 39-5 {
		t.Fatalf("EndReached(%d) outside NearEnd 5", h.calls[0])
	}
	for range 3 {
		h.Frame()
	}
	h.expectCalls("rebuilt at the end", 1)
	box := h.box()
	h.Scroll(box.X+box.W/2, box.Y+box.H/2, 0, 100)
	h.expectCalls("scrolled past the end", 1)

	h.state.SetRows(pageRows(60, false), pageRowID)
	for range 3 {
		h.Frame()
	}
	h.expectCalls("appended 20 rows", 1)

	h.scrollDown(59)
	h.expectCalls("scrolled to the new end", 2)
	if h.calls[1] < 59-5 {
		t.Fatalf("EndReached(%d) outside NearEnd 5 of the new end", h.calls[1])
	}
	if h.counts != (tableCounts{}) {
		t.Fatalf("loading more called Table's callbacks: %+v", h.counts)
	}
}

func TestTableEndReachedIgnoresScrollToEnd(t *testing.T) {
	h := newPageTable(t, pageRows(40, false), 0, 0, true)
	h.state.List.ScrollToEnd()
	for range 3 {
		h.Frame()
	}
	if last := h.lastVisible(); last != 39 {
		t.Fatalf("ScrollToEnd shows row %d last", last)
	}
	h.expectCalls("ScrollToEnd", 0)
}

func TestTableEndReachedShortFirstPage(t *testing.T) {
	h := newPageTable(t, pageRows(5, false), 0, 0, true)
	for range 3 {
		h.Frame()
	}
	box := h.box()
	h.Scroll(box.X+box.W/2, box.Y+box.H/2, 0, 100)
	if last := h.lastVisible(); last != 4 {
		t.Fatalf("short page shows row %d last", last)
	}
	h.expectCalls("short first page", 0)
}

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
