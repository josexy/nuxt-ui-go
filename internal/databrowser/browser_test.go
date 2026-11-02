package databrowser

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestDataBrowserFilterPageEditIdentity(t *testing.T) {
	s := New(1000)
	tt := ui.NewTester(s.View, 1100, 800)
	click := func(label string) {
		t.Helper()
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	click("Task 00001")
	click("Task")
	click("Task")
	click("Edit task")
	click("Task name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Renamed original")
	click("Save task")
	if s.All[0].Name != "Renamed original" || s.All[99].Name == "Renamed original" {
		t.Fatal("edit used display index")
	}
	// The top-right notification includes a progress bar and overlaps the toolbar.
	click("Dismiss Task updated")
	tt.Advance(200 * time.Millisecond)
	click("Next page")
	if s.Page != 2 || s.Table.Len() != 100 {
		t.Fatal("page")
	}
	click("Search tasks")
	tt.Type("00001")
	if s.Total != 0 {
		t.Fatal("filter")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	click("Filters")
	tt.Advance(300 * time.Millisecond)
	click("Active tasks only")
	click("Apply filters")
	tt.Advance(300 * time.Millisecond)
	if s.Total != 666 {
		t.Fatalf("active filter %d", s.Total)
	}
	click("Engineering")
	if s.Total != 333 {
		t.Fatalf("team filter %d", s.Total)
	}
}
