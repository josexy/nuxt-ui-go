package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func addDateInputContract(p *componentPage) {
	value := nui.Set(nui.Date{Year: 2024, Month: 2, Day: 15})
	ranged := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 18}), End: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 22})}
	var state nui.DateInputState
	var rangeState nui.DateRangeInputState
	readonly, loading, blocked, styled := false, false, false, false
	changes, rangeChanges, clears := 0, 0, 0
	p.Scenes = append(p.Scenes, scene("contract", "Date input values and callbacks", "Use the calendar or edit segments. Clear the whole range to receive one range notification. The 20th can be made unavailable; existing values then show an error without being changed.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Checkbox(c, t, &readonly, nui.CheckboxOptions{Label: "Read-only date inputs"})
			nui.Checkbox(c, t, &loading, nui.CheckboxOptions{Label: "Trailing date loader"})
			nui.Checkbox(c, t, &blocked, nui.CheckboxOptions{Label: "Unavailable date 20"})
			nui.Checkbox(c, t, &styled, nui.CheckboxOptions{Label: "Custom date separator"})
		})
		o := nui.InputDateOptions{Label: "Contract date", Color: nui.Neutral, ReadOnly: readonly, Clearable: true, Loading: loading, LoadingTrailing: true, Avatar: &nui.AvatarOptions{Text: "Ada"}, OnClear: func() { clears++ }, ShowCalendar: true,
			Calendar: nui.CalendarOptions{Today: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 15}), Min: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 1}), Max: nui.Set(nui.Date{Year: 2024, Month: 3, Day: 31})},
			Step:     nui.DateStep{Day: 2}, OnValueChange: func(nui.Optional[nui.Date]) { changes++ }, OnRangeChange: func(nui.DateSelection) { rangeChanges++ }}
		if blocked {
			o.Calendar.IsDateUnavailable = func(d nui.Date) bool { return d.Day == 20 }
		}
		nui.InputDate(c.Key("date"), t, &value, &state, o)
		o.Label = "Contract date range"
		o.SeparatorIcon = lucide.ArrowRight()
		if styled {
			o.BuildSeparator = func(c *ui.Context, t nui.Theme) { ui.Text(c, "to").TextColor(t.Info) }
			o.UI = nui.Styles{nui.Segment: {Radius: nui.Set(float32(4))}, nui.SeparatorSlot: {Foreground: nui.Set(t.Info)}}
		}
		nui.InputDateRange(c.Key("range"), t, &ranged, &rangeState, o)
		ui.Textf(c, "Date input: %s · Changes: %d", value.Value.String(), changes)
		ui.Textf(c, "Date input range: %s to %s · Changes: %d · Clears: %d", ranged.Start.Value.String(), ranged.End.Value.String(), rangeChanges, clears)
	}))
}

func addTimeInputContract(p *componentPage) {
	value := nui.Set(nui.ClockTime{Hour: 11, Minute: 55, Second: 42})
	var state nui.TimeInputState
	twelve, seconds, hourOnly, snap, readonly, limited, loading := true, true, false, true, false, false, false
	changes, clears := 0, 0
	p.Scenes = append(p.Scenes, scene("contract", "Time input steps and precision", "Hours cross noon and midnight. Minutes cycle in steps of five; leaving a typed minute snaps it to the nearest step. Changing the display keeps the hidden units and does not notify.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Checkbox(c, t, &twelve, nui.CheckboxOptions{Label: "12-hour contract time"})
			nui.Checkbox(c, t, &seconds, nui.CheckboxOptions{Label: "Contract time seconds"})
			nui.Checkbox(c, t, &hourOnly, nui.CheckboxOptions{Label: "Contract time hour only"})
			nui.Checkbox(c, t, &snap, nui.CheckboxOptions{Label: "Snap contract time steps"})
			nui.Checkbox(c, t, &readonly, nui.CheckboxOptions{Label: "Read-only contract time"})
			nui.Checkbox(c, t, &limited, nui.CheckboxOptions{Label: "Limit contract time to working hours"})
			nui.Checkbox(c, t, &loading, nui.CheckboxOptions{Label: "Trailing time loader"})
		})
		o := nui.InputTimeOptions{Label: "Contract time", Color: nui.Neutral, ReadOnly: readonly, Clearable: true, Avatar: &nui.AvatarOptions{Text: "Ada"}, Loading: loading, LoadingTrailing: true, OnClear: func() { clears++ }, TwelveHour: nui.Set(twelve), Seconds: seconds, Step: nui.TimeStep{Minute: 5}, StepSnapping: snap, OnValueChange: func(nui.Optional[nui.ClockTime]) { changes++ }}
		if hourOnly {
			o.Granularity = nui.TimeHour
		}
		if limited {
			o.Min, o.Max = nui.Set(nui.ClockTime{Hour: 9}), nui.Set(nui.ClockTime{Hour: 17, Minute: 30})
		}
		nui.InputTime(c.Key("time"), t, &value, &state, o)
		nui.Button(c.Key("reset"), t, "Reset contract time", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { value = nui.Set(nui.ClockTime{Hour: 11, Minute: 55, Second: 42}) })
		ui.Textf(c, "Time input: %s · Set: %t · Changes: %d · Clears: %d", value.Value.String(), value.Set, changes, clears)
		ui.Textf(c, "Time draft: %s · Error: %s", state.Text, state.Error)
	}))
}
