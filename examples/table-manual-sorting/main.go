// Table manual sorting keeps expensive work outside the UI thread.
// Copy this directory as a standalone program; it uses only public APIs.
package main

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type record struct {
	ID    int
	Name  string
	Score int
}

type sortRequest struct {
	ctx        context.Context
	generation uint64
	orders     []ui.SortOrder
}

type application struct {
	table    nui.KeyedTableState[record, int]
	original []record // immutable after setup, including while jobs read it
	orders   []ui.SortOrder
	busy     bool
	closed   bool
	version  uint64
	lifetime context.Context
	stop     context.CancelFunc
	cancel   context.CancelFunc
	pending  chan sortRequest // one latest pending request
	done     chan struct{}
}

func (a *application) start(window *mygo.Window) {
	a.lifetime, a.stop = context.WithCancel(context.Background())
	a.pending, a.done = make(chan sortRequest, 1), make(chan struct{})
	original, pending, lifetime, done := a.original, a.pending, a.lifetime, a.done
	go func() {
		defer close(done)
		for {
			select {
			case <-lifetime.Done():
				return
			case request := <-pending:
				rows, err := sortedRecords(request.ctx, original, request.orders)
				if err != nil {
					continue
				}
				// Only immutable output crosses this boundary. No Context,
				// Element, callback builder or native selection runs in the job.
				window.Update(func() { a.accept(request, rows) })
			}
		}
	}()
}

// request runs only from a UI action. Old work is canceled and queued work is
// replaced; the one worker never runs two sorts concurrently.
func (a *application) request(orders []ui.SortOrder) {
	if a.closed {
		return
	}
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(a.lifetime)
	a.cancel = cancel
	a.version++
	a.busy = true
	request := sortRequest{ctx, a.version, slices.Clone(orders)}
	select {
	case a.pending <- request:
	default:
		select {
		case <-a.pending:
		default:
		}
		a.pending <- request
	}
}

// accept is called on the UI thread, even if a job finished just before a newer
// request canceled it. A closed window or obsolete generation cannot publish.
func (a *application) accept(request sortRequest, rows []record) {
	if a.closed || request.ctx.Err() != nil || request.generation != a.version {
		return
	}
	a.table.SetBorrowedRows(rows, func(r record) int { return r.ID })
	a.busy = false
}

func (a *application) close() {
	a.closed = true
	if a.cancel != nil {
		a.cancel()
	}
	if a.stop != nil {
		a.stop()
	}
}

func (a *application) view(c *ui.Context) {
	t := nui.App(c, nui.DefaultApp())
	ui.Column(c.Key("root")).Fill().Padding(24).Gap(12).Children(func() {
		nui.Button(c.Key("restore"), t, "Restore original order", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
			a.orders, a.table.Sort = nil, ui.SortOrder{}
			a.request(nil)
		})
		nui.Table(c.Key("records"), t, &a.table, []nui.TableColumn[record]{
			{ID: "id", Title: "#", Width: 70, RowHeader: true, Value: func(r record) string { return fmt.Sprint(r.ID) }},
			{ID: "name", Title: "Name", Width: 300, Sortable: nui.Set(true), Value: func(r record) string { return r.Name }},
			{ID: "score", Title: "Score", Width: 120, Sortable: nui.Set(true), Value: func(r record) string { return fmt.Sprint(r.Score) }},
		}, nui.KeyedTableOptions[int]{Label: "Records", ManualSorting: true, Loading: a.busy, OnSortChange: a.request}).Grow(1)
		ui.Textf(c.Key("selection"), "Selected: %d. Shift-click headers for multiple sort keys.", a.table.Selection.Len())
	})
}

// A stable merge sort permits cooperative cancellation throughout copying and
// merging. Standard slices.SortFunc is synchronous and has no cancellation API.
// Every call owns its buffers; a published slice is never reused by another job.
func sortedRecords(ctx context.Context, source []record, orders []ui.SortOrder) ([]record, error) {
	rows := make([]record, len(source))
	for i := range source {
		if i%256 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		rows[i] = source[i]
	}
	if len(orders) == 0 {
		return rows, ctx.Err()
	}
	compare := func(a, b record) int {
		for _, order := range orders {
			left, right := a, b
			if order.Descending {
				left, right = right, left
			}
			result := 0
			switch order.Column {
			case "name":
				result = cmp.Compare(left.Name, right.Name)
			case "score":
				result = cmp.Compare(left.Score, right.Score)
			}
			if result != 0 {
				return result
			}
		}
		return 0
	}
	buffer := make([]record, len(rows))
	for width := 1; width < len(rows); width *= 2 {
		for start := 0; start < len(rows); start += 2 * width {
			middle, end := min(start+width, len(rows)), min(start+2*width, len(rows))
			i, j := start, middle
			for at := start; at < end; at++ {
				if at%256 == 0 {
					if err := ctx.Err(); err != nil {
						return nil, err
					}
				}
				if i < middle && (j == end || compare(rows[i], rows[j]) <= 0) {
					buffer[at] = rows[i]
					i++
				} else {
					buffer[at] = rows[j]
					j++
				}
			}
		}
		rows, buffer = buffer, rows
	}
	return rows, ctx.Err()
}

func main() {
	a := &application{original: make([]record, 50000)}
	for i := range a.original {
		a.original[i] = record{i, fmt.Sprintf("Record %06d", len(a.original)-i), (i * 7919) % 1000}
	}
	a.table.SetBorrowedRows(a.original, func(r record) int { return r.ID })
	a.table.List.SortBy = &a.orders
	a.table.List.CellSelect = true
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{Title: "Cancellable manual sorting", Width: 820, Height: 620, Content: ui.View(a.view)})
		a.start(window)
		window.OnClosed(a.close)
	})
	err := mygo.App.Run()
	a.close()
	if a.done != nil {
		<-a.done
	}
	if err != nil {
		log.Fatal(err)
	}
}
