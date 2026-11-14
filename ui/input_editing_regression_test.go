package nui_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestPinInputClickSelectsVisibleDigit(t *testing.T) {
	for _, mask := range []bool{false, true} {
		for slot := range 6 {
			t.Run(fmt.Sprintf("mask=%v/slot=%d", mask, slot+1), func(t *testing.T) {
				value := "123456"
				var state nui.PinState
				changes := 0
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					// Non-default spacing and font size must not affect hit mapping.
					th.TextScale, th.Density = 1.5, 1.25
					ui.Column(c).Children(func() {
						index := 0
						nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Numeric: true, Length: 6,
							Label: "Code", Mask: mask, OnChange: func(string) { changes++ },
							Style: func(part nui.Slot, e ui.Element) {
								if part == nui.Item {
									index++
									e.Label(fmt.Sprintf("Digit %d", index))
								}
							},
						})
						nui.Button(c, th, "Next", nui.ButtonOptions{})
					})
				}, 500, 160)
				click(t, tt, fmt.Sprintf("Digit %d", slot+1))
				if !tt.Focused("Code") || changes != 0 {
					t.Fatal("click should focus the editor without changing its value")
				}
				tt.Frame()
				tt.Key(0, ui.KeyBackspace)
				if want := "123456"[:slot] + "123456"[slot+1:]; value != want {
					t.Fatalf("Backspace = %q, want %q", value, want)
				}
				tt.Command("undo")
				if value != "123456" {
					t.Fatalf("undo = %q", value)
				}
				click(t, tt, fmt.Sprintf("Digit %d", slot+1))
				tt.Type("9")
				if want := "123456"[:slot] + "9" + "123456"[slot+1:]; value != want {
					t.Fatalf("replacement = %q, want %q", value, want)
				}
				tt.Key(0, ui.KeyTab)
				if !tt.Focused("Next") {
					t.Fatal("PIN slots must not add tab stops")
				}
			})
		}
	}
}

func TestPinInputEmptySlotsDisabledAndRemount(t *testing.T) {
	value := "12"
	var state nui.PinState
	disabled, shown := false, true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Disabled(disabled).Children(func() {
			if shown {
				index := 0
				nui.PinInput(c.Key("pin"), th, &value, &state, nui.PinInputOptions{Numeric: true, Length: 6,
					Label: "Code", Style: func(part nui.Slot, e ui.Element) {
						if part == nui.Item {
							index++
							e.Label(fmt.Sprintf("Digit %d", index))
						}
					},
				})
			}
		})
	}, 400, 120)
	click(t, tt, "Digit 6")
	tt.SetClipboard("3456")
	tt.Command("paste")
	if value != "123456" {
		t.Fatalf("empty slot paste = %q", value)
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Digit 2")
	tt.Key(0, ui.KeyBackspace)
	if value != "123456" {
		t.Fatal("inherited disabled state allowed editing")
	}
	shown = false
	tt.Frame()
	tt.Frame()
	shown, disabled = true, false
	tt.Frame()
	click(t, tt, "Digit 2")
	tt.Type("9")
	if value != "193456" {
		t.Fatalf("remounted replacement = %q", value)
	}
}

func TestCalendarExternalSelectionAndBrowsing(t *testing.T) {
	date := func(y int, m time.Month, d int) nui.Date { return nui.Date{Year: y, Month: m, Day: d} }
	value := nui.DateSelection{Start: nui.Set(date(2026, 11, 15))}
	var state nui.CalendarState
	changes := 0
	monthChanges := 0
	options := nui.CalendarOptions{Label: "Date", Range: true, OnChange: func(nui.DateSelection) { changes++ }, OnMonthChange: func(nui.Date) { monthChanges++ }}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Calendar(c, th, &value, &state, options)
	}, 400, 400)
	value.Start = nui.Set(date(2027, 2, 10))
	tt.Frame()
	if !tt.HasText("February 2027") || state.Active != value.Start.Value || changes != 0 {
		t.Fatalf("external date did not synchronize silently: %+v, changes=%d", state, changes)
	}
	click(t, tt, "Next month")
	tt.Frame()
	if !tt.HasText("March 2027") || value.Start.Value != date(2027, 2, 10) || changes != 0 {
		t.Fatal("browsing should persist without changing the selected date")
	}
	value.End = nui.Set(date(2027, 4, 20))
	tt.Frame()
	if !tt.HasText("April 2027") || state.Active != value.End.Value || changes != 0 {
		t.Fatal("external range endpoint did not synchronize silently")
	}
	value = nui.DateSelection{}
	tt.Frame()
	if !tt.HasText("April 2027") {
		t.Fatal("clearing the selection should preserve the browsed month")
	}
	value.Start = nui.Set(date(2027, 2, 10))
	tt.Frame()
	click(t, tt, "Next month")
	click(t, tt, "Date 2027-03-15")
	tt.Frame()
	if !tt.HasText("March 2027") || state.Active != date(2027, 3, 15) || value.End.Value != state.Active || changes != 1 {
		t.Fatal("completing a range jumped back to its start")
	}
	// Start another range, then complete it in reverse chronological order.
	click(t, tt, "Date 2027-03-20")
	click(t, tt, "Previous month")
	click(t, tt, "Date 2027-02-12")
	tt.Frame()
	if !tt.HasText("February 2027") || state.Active != date(2027, 2, 12) || value.Start.Value != state.Active {
		t.Fatal("reverse range selection moved the cursor away from the clicked date")
	}
	options.Range, options.WeekStart = false, time.Monday
	options.Min, options.Max = nui.Set(date(2027, 2, 10)), nui.Set(date(2027, 2, 28))
	options.IsDateDisabled = func(d nui.Date) bool { return d.Day == 22 }
	value = nui.DateSelection{Start: nui.Set(date(2027, 2, 15))}
	tt.Frame()
	before := changes
	click(t, tt, "Date 2027-02-15")
	if value.Start.Set || changes != before+1 {
		t.Fatal("same date did not deselect once")
	}
	click(t, tt, "Date 2027-02-15")
	options.PreventDeselect = true
	tt.Frame()
	click(t, tt, "Date 2027-02-15")
	if !value.Start.Set || changes != before+2 {
		t.Fatal("prevented deselection notified a model change")
	}
	tt.Key(0, ui.KeyEnd)
	if state.Active != date(2027, 2, 21) {
		t.Fatalf("End did not reach the end of the week: %v", state.Active)
	}
	tt.Key(0, ui.KeyRight)
	if state.Active != date(2027, 2, 23) {
		t.Fatalf("arrow did not skip disabled date: %v", state.Active)
	}
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyRight)
	if state.Active != date(2027, 2, 28) {
		t.Fatal("arrow crossed maximum date")
	}
	options.Min, options.Max = nui.Set(date(2026, 1, 1)), nui.Set(date(2028, 12, 31))
	tt.Frame()
	tt.Key(ui.Shift, ui.KeyPageUp)
	if state.Active != date(2026, 2, 28) {
		t.Fatalf("Shift+PageUp did not page by year: %v", state.Active)
	}
	beforeMonths := monthChanges
	state.Month = date(2028, 6, 15)
	options.ReadOnly = true
	tt.Frame()
	if state.Month != date(2028, 6, 1) || state.Active != state.Month || monthChanges != beforeMonths {
		t.Fatal("controlled month was not normalized silently")
	}
	before = changes
	click(t, tt, "Date 2028-06-15")
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeyRight)
	if changes != before || state.Active != date(2028, 6, 16) {
		t.Fatal("read-only must navigate without selecting")
	}
	click(t, tt, "Next month")
	if state.Month != date(2028, 7, 1) || monthChanges != beforeMonths+1 {
		t.Fatal("read-only month navigation failed")
	}
	options.NumberOfMonths, options.PagedNavigation = 2, true
	tt.Frame()
	if !tt.HasText("Date 2028-08-15") {
		t.Fatal("second calendar month is missing")
	}
	click(t, tt, "Next month")
	if state.Month != date(2028, 9, 1) {
		t.Fatalf("paged month navigation: %v", state.Month)
	}
	options.NumberOfMonths, options.ReadOnly = 1, false
	options.Min = nui.Set(date(2023, 1, 1))
	state.Month = date(2024, 2, 1)
	tt.Frame()
	before = changes
	previousValue := value
	click(t, tt, "February 2024")
	if state.View != nui.CalendarMonthView {
		t.Fatal("heading did not open month view")
	}
	click(t, tt, "Date Mar 2024")
	if state.View != nui.CalendarDayView || state.Month != date(2024, 3, 1) {
		t.Fatal("month view did not navigate back to days")
	}
	click(t, tt, "March 2024")
	click(t, tt, "2024")
	if state.View != nui.CalendarYearView {
		t.Fatal("heading did not open year view")
	}
	click(t, tt, "Date 2025")
	click(t, tt, "Date Feb 2025")
	if state.View != nui.CalendarDayView || state.Month != date(2025, 2, 1) || value != previousValue || changes != before {
		t.Fatal("view navigation committed a date")
	}
}

func TestInputRatingKeyboardDirection(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			value := 3
			readOnly, disabled := false, false
			changes := 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Rating", ReadOnly: readOnly, Disabled: disabled, OnChange: func(int) { changes++ }})
			}, 400, 100)
			tt.SetDirection(direction)
			tt.Key(0, ui.KeyTab)
			increase, decrease := ui.KeyRight, ui.KeyLeft
			if direction == ui.RightToLeft {
				increase, decrease = decrease, increase
			}
			for _, step := range []struct {
				key  ui.Key
				want int
			}{{increase, 4}, {decrease, 3}, {ui.KeyUp, 3}, {ui.KeyDown, 3}, {ui.KeyHome, 1}, {decrease, 1}, {ui.KeyEnd, 5}, {increase, 5}} {
				tt.Key(0, step.key)
				if value != step.want {
					t.Fatalf("key %v: value=%d, want=%d", step.key, value, step.want)
				}
			}
			// The arrows stop at the first and last star without wrapping, as Nuxt's.
			if changes != 4 {
				t.Fatalf("change callbacks=%d, want 4", changes)
			}
			readOnly = true
			tt.Frame()
			tt.Key(0, decrease)
			readOnly, disabled = false, true
			tt.Frame()
			tt.Key(0, decrease)
			if value != 5 || changes != 4 {
				t.Fatal("read-only or disabled rating changed")
			}
		})
	}
}

// Re-verifies the selected-text IME cancellation known issue against the
// pinned MyGo (v0.9.0): docs/compatibility.md describes what this asserts.
// When MyGo keeps the selection, this fails; update the docs with it.
func TestIMECancelOverSelectionProbe(t *testing.T) {
	for _, kind := range []string{"Input", "Textarea"} {
		t.Run(kind, func(t *testing.T) {
			value := "Draft"
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				if kind == "Input" {
					nui.Input(c.Key("field"), th, &value, nui.InputOptions{Label: "Field"})
				} else {
					nui.Textarea(c.Key("field"), th, &value, nui.TextareaOptions{Label: "Field"})
				}
			}, 400, 200)
			click(t, tt, "Field")
			tt.Key(ui.Ctrl, ui.KeyA)
			tt.Compose("x", 1)
			tt.Compose("", 0)
			if value != "" {
				t.Fatalf("canceling a composition over a selection now leaves %q: the known issue is gone, update docs/compatibility.md", value)
			}
			tt.Command("undo")
			if value != "Draft" {
				t.Fatalf("undo restored %q", value)
			}
		})
	}
}
