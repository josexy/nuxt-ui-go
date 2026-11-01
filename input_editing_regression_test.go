package nui_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Calendar(c, th, &value, &state, nui.CalendarOptions{Label: "Date", Range: true, OnChange: func(nui.DateSelection) { changes++ }})
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
			}{{increase, 4}, {decrease, 3}, {ui.KeyUp, 3}, {ui.KeyDown, 3}, {ui.KeyHome, 1}, {ui.KeyEnd, 5}, {increase, 1}} {
				tt.Key(0, step.key)
				if value != step.want {
					t.Fatalf("key %v: value=%d, want=%d", step.key, value, step.want)
				}
			}
			if changes != 5 {
				t.Fatalf("change callbacks=%d, want 5", changes)
			}
			readOnly = true
			tt.Frame()
			tt.Key(0, decrease)
			readOnly, disabled = false, true
			tt.Frame()
			tt.Key(0, decrease)
			if value != 1 || changes != 5 {
				t.Fatal("read-only or disabled rating changed")
			}
		})
	}
}
