package nui

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
)

func calendarYearsPerPage(o CalendarOptions) int {
	if o.YearsPerPage == 0 {
		return 12
	}
	return o.YearsPerPage
}

func calendarYearPage(s *CalendarState, o CalendarOptions) {
	count := calendarYearsPerPage(o)
	if s.yearCount != count || s.yearStart == 0 || s.Active.Year < s.yearStart || s.Active.Year >= s.yearStart+count {
		s.yearStart = max(1, s.Active.Year/10*10)
		if s.Active.Year >= s.yearStart+count {
			s.yearStart = s.Active.Year
		}
		s.yearCount = count
	}
}

func calendarYearPageEnabled(start, count int, o CalendarOptions) bool {
	return (!o.Min.Set || start+count-1 >= o.Min.Value.Year) && (!o.Max.Set || start <= o.Max.Value.Year)
}

func calendarPeriodEnabled(date Date, view CalendarView, o CalendarOptions, available bool) bool {
	if !date.Valid() {
		return false
	}
	start, end := calendarMonth(date), calendarMonth(date).AddMonths(1).AddDays(-1)
	disabled, unavailable := o.IsMonthDisabled, o.IsMonthUnavailable
	if view == CalendarYearView {
		start, end = Date{date.Year, 1, 1}, Date{date.Year, 12, 31}
		disabled, unavailable = o.IsYearDisabled, o.IsYearUnavailable
	}
	return (!o.Min.Set || end.Compare(o.Min.Value) >= 0) && (!o.Max.Set || start.Compare(o.Max.Value) <= 0) && (disabled == nil || !disabled(start)) && (!available || unavailable == nil || !unavailable(start))
}

func calendarPicker(c *ui.Context, t Theme, s *CalendarState, o CalendarOptions, config ComponentTheme, today Date, disabled bool, changeMonth func(Date, Date), setView func(CalendarView)) {
	view := s.View
	m := calendarRecipe(t, config.Size)
	start, count := Date{s.Month.Year, 1, 1}, 12
	if view == CalendarYearView {
		calendarYearPage(s, o)
		start, count = Date{s.yearStart, 1, 1}, min(calendarYearsPerPage(o), 10000-s.yearStart)
	}
	dateAt := func(i int) Date {
		if view == CalendarYearView {
			return Date{start.Year + i, 1, 1}
		}
		return Date{start.Year, time.Month(i + 1), 1}
	}
	choose := func(date Date) {
		if disabled || !calendarPeriodEnabled(date, view, o, true) {
			return
		}
		active := date
		if o.Min.Set && active.Compare(o.Min.Value) < 0 {
			active = o.Min.Value
		}
		if o.Max.Set && active.Compare(o.Max.Value) > 0 {
			active = o.Max.Value
		}
		changeMonth(date, active)
		setView(view - 1)
	}
	grid := ui.Column(c.Key("grid")).Focusable().FocusRing(false).Bind(&s.Handle).Role(ui.RoleGrid).Label(o.Label).ReadOnly(o.ReadOnly).InputMethod(false)
	bindField(c, grid, t)
	finish(t, ComponentCalendar, GridSlot, grid, PartStyle{Gap: Set(m.gap)}, o.Decoration)
	direction := 1
	if c.Direction() == ui.RightToLeft {
		direction = -1
	}
	grid.HandleInput(func(ev ui.InputEvent) bool {
		if disabled || ev.Kind != ui.InputKeyDown || ev.Mods != 0 {
			return false
		}
		current := calendarMonth(s.Active)
		if view == CalendarYearView {
			current = Date{current.Year, 1, 1}
		}
		step := 0
		switch ev.Key {
		case ui.KeyLeft:
			step = -direction
		case ui.KeyRight:
			step = direction
		case ui.KeyUp:
			step = -4
		case ui.KeyDown:
			step = 4
		case ui.KeyPageUp:
			step = -12
			if view == CalendarYearView {
				step = -calendarYearsPerPage(o)
			}
		case ui.KeyPageDown:
			step = 12
			if view == CalendarYearView {
				step = calendarYearsPerPage(o)
			}
		case ui.KeyEnter, ui.KeySpace:
			choose(current)
			return true
		default:
			return false
		}
		multiplier := 1
		if view == CalendarYearView {
			multiplier = 12
		}
		for attempt := 1; attempt <= 49; attempt++ {
			next := current.AddMonths(step * multiplier * attempt)
			if !next.Valid() {
				break
			}
			if calendarPeriodEnabled(next, view, o, false) {
				if view == CalendarYearView && (ev.Key == ui.KeyPageUp || ev.Key == ui.KeyPageDown) {
					s.yearStart = max(1, min(9999, s.yearStart+step))
				}
				changeMonth(next, next)
				break
			}
		}
		return true
	})
	grid.Children(func() {
		for rowIndex := range (count + 3) / 4 {
			row := ui.Row(c.Key(rowIndex)).Role(ui.RoleRow)
			finish(t, ComponentCalendar, GridRowSlot, row, PartStyle{}, o.Decoration)
			row.Children(func() {
				for column := range 4 {
					index := rowIndex*4 + column
					if index >= count {
						break
					}
					date := dateAt(index)
					label := i18n.Months(c.Locale(), i18n.Abbreviated, true)[date.Month-1]
					name := fmt.Sprintf("%s %d", label, date.Year)
					active, isToday := s.Active.Year == date.Year && s.Active.Month == date.Month, today.Year == date.Year && today.Month == date.Month
					if view == CalendarYearView {
						label, name = fmt.Sprint(date.Year), fmt.Sprint(date.Year)
						active, isToday = s.Active.Year == date.Year, today.Year == date.Year
					}
					enabled := calendarPeriodEnabled(date, view, o, true)
					unavailable := o.IsMonthUnavailable
					if view == CalendarYearView {
						unavailable = o.IsYearUnavailable
					}
					cell := ui.Box(c.Key(date)).Width(m.column*7/4 - t.spacing(4)).Height(m.cell).Margin(t.spacing(2)).MinWidth(0).Center().Role(ui.RoleCell).Label(strings.TrimSpace(o.Label + " " + name)).ReadOnly(o.ReadOnly).Disabled(!enabled).OnClick(func() { choose(date) })
					style := calendarCellStyle(t, config, false, false, cell.Hovered() && enabled && !disabled, isToday, false, active && grid.FocusVisible())
					style.Radius = Set(t.radius(6))
					if !enabled {
						style.Foreground = Set(t.Muted)
					}
					if active {
						grid.ActiveDescendant(cell)
						if grid.FocusVisible() {
							cell.Shadow(0, 0, 0, 3, inputFocusColor(t, config.Color))
						}
					}
					finish(t, ComponentCalendar, Item, cell, style, o.Decoration)
					cell.Children(func() {
						build := o.BuildMonth
						if view == CalendarYearView {
							build = o.BuildYear
						}
						if build != nil {
							build(c, t, date)
							return
						}
						text := ui.Text(c, label).SingleLine().MinWidth(0)
						if unavailable != nil && unavailable(date) {
							text.Strikethrough()
						}
					})
				}
			})
		}
	})
}
