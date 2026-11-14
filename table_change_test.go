package nui_test

import (
	"cmp"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type tableFile struct {
	ID, Name, Modified string
	Size               int
}

var tableFiles = []tableFile{{"a", "Alpha", "Mon", 3}, {"b", "Beta", "Tue", 2}, {"c", "Gamma", "Wed", 1}}

type tableCounts struct{ change, sort, selection, activate int }

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
			Label:             "Files",
			OnChange:          func() { h.counts.change++ },
			OnSortChange:      func([]ui.SortOrder) { h.counts.sort++ },
			OnSelectionChange: func([]string) { h.counts.selection++ },
			OnActivate:        func(id string) { h.counts.activate++; h.activated = append(h.activated, id) },
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
	h.expect("choosing Beta", tableCounts{change: 1, selection: 1})
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
