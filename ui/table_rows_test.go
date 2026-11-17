package nui_test

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

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
			Label:             "Rows",
			Height:            height,
			OnChange:          func() { h.counts.change++ },
			OnSortChange:      func([]ui.SortOrder) { h.counts.sort++ },
			OnSelectionChange: func([]string) { h.counts.selection++ },
			OnActivate:        func(string) { h.counts.activate++ },
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
