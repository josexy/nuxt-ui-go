package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func addCalendarContract(p *componentPage) {
	value := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 15})}
	state := nui.CalendarState{Month: nui.Date{Year: 2024, Month: 2, Day: 1}}
	readonly, ranged, prevent, variable, multiple, weeks := false, false, false, false, false, false
	unavailable, gaps, limited, styled := true, false, false, false
	size, variant := nui.MD, nui.Solid
	changes, months, views := 0, 0, 0
	p.Scenes = append(p.Scenes, scene("contract", "Calendar values and navigation", "Choose a day twice to clear it. The 20th is unavailable; range selection can include it only when gaps are allowed. The heading switches day, month and year views without committing a date.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Switch(c.Key("readonly"), t, &readonly, nui.SwitchOptions{Label: "Read-only contract calendar"})
			nui.Switch(c.Key("range"), t, &ranged, nui.SwitchOptions{Label: "Contract calendar range"})
			nui.Switch(c.Key("prevent"), t, &prevent, nui.SwitchOptions{Label: "Prevent calendar deselection"})
			nui.Switch(c.Key("variable"), t, &variable, nui.SwitchOptions{Label: "Variable calendar weeks"})
			nui.Switch(c.Key("multiple"), t, &multiple, nui.SwitchOptions{Label: "Show two calendar months"})
			nui.Switch(c.Key("weeks"), t, &weeks, nui.SwitchOptions{Label: "Show calendar week numbers"})
			nui.Switch(c.Key("unavailable"), t, &unavailable, nui.SwitchOptions{Label: "Mark the 20th unavailable"})
			nui.Switch(c.Key("gaps"), t, &gaps, nui.SwitchOptions{Label: "Allow unavailable range dates"})
			nui.Switch(c.Key("limit"), t, &limited, nui.SwitchOptions{Label: "Limit calendar to seven days"})
			nui.Switch(c.Key("styles"), t, &styled, nui.SwitchOptions{Label: "Custom calendar parts"})
		})
		o := nui.CalendarOptions{Label: "Contract calendar", ReadOnly: readonly, Range: ranged, Color: nui.Neutral, Variant: variant, Size: size,
			WeekStart: time.Monday, WeekNumbers: weeks, FixedWeeks: nui.Set(!variable), PreventDeselect: prevent, AllowNonContiguousRanges: gaps, PagedNavigation: true,
			Min: nui.Set(nui.Date{Year: 2023, Month: 1, Day: 1}), Max: nui.Set(nui.Date{Year: 2028, Month: 12, Day: 31}), Today: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 15}),
			PreviousMonthIcon: lucide.ArrowLeft(), NextMonthIcon: lucide.ArrowRight(),
			ViewControlOptions: &nui.ButtonOptions{AccessibleLabel: "Change contract calendar view"},
			OnChange:           func(nui.DateSelection) { changes++ }, OnMonthChange: func(nui.Date) { months++ }, OnViewChange: func(nui.CalendarView) { views++ }}
		if multiple {
			o.NumberOfMonths = 2
		}
		if limited {
			o.MaximumDays = 7
		}
		if unavailable {
			o.IsDateUnavailable = func(d nui.Date) bool { return d.Day == 20 }
		}
		if styled {
			o.UI = nui.Styles{nui.WeekDaySlot: {Foreground: nui.Set(t.Info)}, nui.HeadingSlot: {Foreground: nui.Set(t.Warning)}}
		}
		nui.Calendar(c.Key("calendar"), t, &value, &state, o)
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c.Key("size"), t, "Toggle calendar size", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() {
				if size == nui.MD {
					size = nui.LG
				} else {
					size = nui.MD
				}
			})
			nui.Button(c.Key("variant"), t, "Toggle calendar variant", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() {
				if variant == nui.Solid {
					variant = nui.Outline
				} else {
					variant = nui.Solid
				}
			})
			nui.Button(c.Key("jump"), t, "Jump calendar to February", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { state.Month = nui.Date{Year: 2024, Month: 2, Day: 15}; state.View = nui.CalendarDayView })
			nui.Button(c.Key("clear"), t, "Clear contract calendar", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { value = nui.DateSelection{} })
		})
		ui.Textf(c, "Calendar date: %s to %s · Changes: %d", value.Start.Value.String(), value.End.Value.String(), changes)
		ui.Textf(c, "Calendar month: %s · Month changes: %d", state.Month, months)
		ui.Textf(c, "Calendar active: %s · View: %d · View changes: %d", state.Active, state.View, views)
	}))
}
