package main

import (
	"context"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestStableSortOwnsResultsAndLeavesInputOrder(t *testing.T) {
	rows := []record{{0, "b", 2}, {1, "a", 1}, {2, "a", 2}, {3, "a", 2}}
	original := slices.Clone(rows)
	got, err := sortedRecords(context.Background(), rows, []ui.SortOrder{{Column: "name"}, {Column: "score", Descending: true}})
	if err != nil || !slices.Equal(got, []record{rows[2], rows[3], rows[1], rows[0]}) || !slices.Equal(rows, original) {
		t.Fatalf("sort=%v error=%v input=%v", got, err, rows)
	}
	got[0].Name = "edited"
	if !slices.Equal(rows, original) {
		t.Fatal("output aliases immutable input")
	}
	_, _ = sortedRecords(context.Background(), rows, []ui.SortOrder{{Column: "score"}})
	if got[0].Name != "edited" {
		t.Fatal("a later job reused published output")
	}
}

type cancelDuringWork struct {
	context.Context
	checks int
}

func (c *cancelDuringWork) Err() error {
	c.checks++
	if c.checks >= 6 {
		return context.Canceled
	}
	return nil
}

func TestSortCancellationAndObsoletePublication(t *testing.T) {
	ctx := &cancelDuringWork{Context: context.Background()}
	if rows, err := sortedRecords(ctx, make([]record, 1024), []ui.SortOrder{{Column: "score"}}); err != context.Canceled || rows != nil {
		t.Fatal("canceled merge published rows")
	}
	a := &application{version: 2, busy: true}
	a.table.SetRows([]record{{0, "Current", 0}}, func(r record) int { return r.ID })
	a.accept(sortRequest{ctx: context.Background(), generation: 1}, []record{{1, "Old", 0}})
	if a.table.RowID(0) != 0 || !a.busy {
		t.Fatal("obsolete generation published")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	a.accept(sortRequest{ctx: canceled, generation: 2}, []record{{1, "Canceled", 0}})
	if a.table.RowID(0) != 0 {
		t.Fatal("canceled result published")
	}
	a.accept(sortRequest{ctx: context.Background(), generation: 2}, []record{{2, "Latest", 0}})
	if a.table.RowID(0) != 2 || a.busy {
		t.Fatal("latest result was not accepted")
	}
	a.close()
	a.accept(sortRequest{ctx: context.Background(), generation: 2}, nil)
	if a.table.Len() != 1 {
		t.Fatal("closed owner accepted a delayed result")
	}
}
