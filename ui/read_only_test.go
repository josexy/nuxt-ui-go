package nui_test

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// focusLabel tabs through the view until the element labeled name has the
// keyboard focus.
func focusLabel(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	for range 20 {
		if tt.Focused(name) {
			return
		}
		tt.Key(0, ui.KeyTab)
	}
	t.Fatalf("could not focus %q", name)
}

func TestCalendarGridIsAnInteractiveGrid(t *testing.T) {
	day := nui.Date{Year: 2026, Month: time.November, Day: 13}
	for _, readOnly := range []bool{false, true} {
		name := "editable"
		if readOnly {
			name = "read-only"
		}
		t.Run(name, func(t *testing.T) {
			selection := nui.DateSelection{Start: nui.Set(day)}
			var state nui.CalendarState
			changes := 0
			r := newReaderTester(t, func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Calendar(c.Key("calendar"), th, &selection, &state, nui.CalendarOptions{Label: "Due", ReadOnly: readOnly, OnChange: func(nui.DateSelection) { changes++ }})
			}, 640, 560)

			grid := r.role(ui.RoleGrid, "Due")
			if grid.ReadOnly != readOnly {
				t.Fatalf("grid read-only %v, want %v", grid.ReadOnly, readOnly)
			}
			if _, ok := r.AccessByRole(ui.RoleTable, ""); ok {
				t.Fatal("the day grid is still reported as a table")
			}
			for _, weekday := range []string{"Sunday", "Monday", "Saturday"} {
				r.role(ui.RoleColumnHeader, weekday)
			}
			cell := r.role(ui.RoleCell, "Due 2026-11-13")
			if !cell.Selected || cell.ReadOnly != readOnly || !inside(cell.Bounds, grid.Bounds) {
				t.Fatalf("13 November cell %+v in grid %v", cell, grid.Bounds)
			}
			if other := r.role(ui.RoleCell, "Due 2026-11-14"); other.Selected || other.ReadOnly != readOnly {
				t.Fatalf("14 November cell %+v", other)
			}

			focusLabel(t, r.Tester, "Due")
			if active := r.role(ui.RoleCell, "Due 2026-11-13"); !active.Focused {
				t.Fatalf("the grid does not point assistive technology at the active day: %+v", active)
			}
			r.Key(0, ui.KeyRight)
			if want := (nui.Date{Year: 2026, Month: time.November, Day: 14}); state.Active != want || !r.role(ui.RoleCell, "Due 2026-11-14").Focused {
				t.Fatalf("Right moved the focused date to %v, want %v", state.Active, want)
			}
			r.Key(0, ui.KeyEnter)
			if readOnly {
				if selection.Start.Value != day || changes != 0 {
					t.Fatalf("a read-only calendar changed its value to %+v (%d changes)", selection, changes)
				}
			} else if selection.Start.Value != state.Active || changes != 1 {
				t.Fatalf("Enter chose %+v (%d changes)", selection, changes)
			}
		})
	}
}

func TestCalendarMonthAndYearGridsAreGrids(t *testing.T) {
	selection := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 13})}
	state := nui.CalendarState{View: nui.CalendarMonthView}
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Calendar(c.Key("calendar"), th, &selection, &state, nui.CalendarOptions{Label: "Due", ReadOnly: true})
	}, 640, 560)
	if grid := r.role(ui.RoleGrid, "Due"); !grid.ReadOnly {
		t.Fatalf("month grid %+v", grid)
	}
	if cell := r.role(ui.RoleCell, "Due Nov 2026"); !cell.ReadOnly {
		t.Fatalf("month cell %+v", cell)
	}
	state.View = nui.CalendarYearView
	r.Frame()
	if grid := r.role(ui.RoleGrid, "Due"); !grid.ReadOnly {
		t.Fatalf("year grid %+v", grid)
	}
	r.role(ui.RoleCell, "Due 2026")
}

func TestReadOnlyControlsReportReadOnly(t *testing.T) {
	rating := 3
	date := nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 13})
	clock := nui.Set(nui.ClockTime{Hour: 9, Minute: 30})
	var dateState nui.DateInputState
	var timeState nui.TimeInputState
	r := newReaderTester(t, func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Gap(12).Children(func() {
			nui.InputRating(c.Key("rating"), th, &rating, nui.InputRatingOptions{Label: "Score", ReadOnly: true})
			nui.InputDate(c.Key("date"), th, &date, &dateState, nui.InputDateOptions{Label: "Due", ReadOnly: true})
			nui.InputTime(c.Key("time"), th, &clock, &timeState, nui.InputTimeOptions{Label: "Start"})
		})
	}, 640, 400)

	score := r.role(ui.RoleSlider, "Score")
	if !score.ReadOnly || score.Now != 3 || score.Value != "3" {
		t.Fatalf("read-only rating %+v", score)
	}
	if month := r.role(ui.RoleStepper, "Due month"); !month.ReadOnly {
		t.Fatalf("read-only InputDate month segment %+v", month)
	}
	if hour := r.role(ui.RoleStepper, "Start hour"); hour.ReadOnly {
		t.Fatalf("editable InputTime hour segment reports read-only: %+v", hour)
	}
}
