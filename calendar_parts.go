package nui

import (
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
)

type calendarMetrics struct{ cell, column, font, line, weekday, gap float32 }

func calendarRecipe(t Theme, size Size) calendarMetrics {
	i := sizeIndex(size)
	cell := t.text([5]float32{24, 28, 32, 36, 40}[i])
	return calendarMetrics{cell: cell, column: cell + t.spacing(4), font: t.text([5]float32{12, 12, 14, 16, 18}[i]), line: t.text([5]float32{16, 16, 20, 24, 28}[i]), weekday: t.text([5]float32{10, 12, 12, 16, 18}[i]), gap: t.spacing(4)}
}

func calendarCellStyle(t Theme, config ComponentTheme, selected, highlighted, hovered, today, outside, focused bool) PartStyle {
	style := PartStyle{Radius: Set(float32(999)), Foreground: Set(t.Text)}
	if outside {
		style.Foreground = Set(t.Muted)
	}
	if selected {
		style = mergeStyle(style, appearance(t, config.Color, config.Variant, false, focused))
		if config.Color == Neutral && focused && (config.Variant == Outline || config.Variant == Subtle) {
			style.BorderColor = Set(t.InvertedBackground)
		}
	} else {
		if today {
			style.Foreground, style.Weight = Set(t.color(config.Color)), Set(600)
		}
		if highlighted || hovered {
			alpha := float32(.2)
			if config.Variant == Outline || config.Color == Neutral && !highlighted {
				alpha = .1
			}
			color := t.color(config.Color)
			if config.Color == Neutral {
				color = t.InvertedBackground
			}
			style.Background = Set(color.Alpha(alpha))
		}
	}
	return style
}

func calendarMonth(date Date) Date { return Date{date.Year, date.Month, 1} }

func calendarMonthEnabled(month Date, o CalendarOptions) bool {
	return month.Valid() && (!o.Min.Set || month.AddMonths(1).AddDays(-1).Compare(o.Min.Value) >= 0) && (!o.Max.Set || month.Compare(o.Max.Value) <= 0)
}

func calendarNavigationButton(c *ui.Context, t Theme, size Size, name string, icon *ui.SVG, options *ButtonOptions, enabled bool, action func()) ui.Element {
	o := ButtonOptions{}
	if options != nil {
		o = *options
	}
	if o.AccessibleLabel == "" {
		o.AccessibleLabel = message(c, name)
	}
	if o.Leading == nil {
		o.Leading = icon
	}
	if o.Color == ColorDefault {
		o.Color = Neutral
	}
	if o.Variant == VariantDefault {
		o.Variant = Ghost
	}
	if o.Size == SizeDefault {
		o.Size = size
	}
	o.Disabled = o.Disabled || !enabled
	return Button(c, t, "", o).OnClick(action)
}

func calendarWeekdayWidth(format CalendarWeekdayFormat) i18n.Width {
	switch format {
	case CalendarWeekdayShort:
		return i18n.Abbreviated
	case CalendarWeekdayLong:
		return i18n.Wide
	default:
		return i18n.Narrow
	}
}

// Disabled-day navigation follows the same stride as the key (one day or one
// week). Date validity and the configured bounds provide the finite domain.
func calendarNextDate(from Date, step int, o CalendarOptions) Date {
	for next := from.AddDays(step); dateInBounds(next, o); next = next.AddDays(step) {
		if dateNavigable(next, o) {
			return next
		}
	}
	return from
}

func calendarWeekBoundary(from Date, end bool, o CalendarOptions) Date {
	start := from.AddDays(-(int(from.dayTime().Weekday()) - int(o.WeekStart) + 7) % 7)
	for i := range 7 {
		offset := i
		if end {
			offset = 6 - i
		}
		date := start.AddDays(offset)
		if dateNavigable(date, o) {
			return date
		}
	}
	return from
}

func calendarPageDate(from Date, months int, o CalendarOptions) Date {
	target := from.AddMonths(months)
	if o.Min.Set && target.Compare(o.Min.Value) < 0 {
		target = o.Min.Value
	}
	if o.Max.Set && target.Compare(o.Max.Value) > 0 {
		target = o.Max.Value
	}
	if !target.Valid() {
		return from
	}
	if dateNavigable(target, o) {
		return target
	}
	step := 1
	if months < 0 {
		step = -1
	}
	if o.Min.Set && target == o.Min.Value {
		step = 1
	}
	if o.Max.Set && target == o.Max.Value {
		step = -1
	}
	result := calendarNextDate(target, step, o)
	if result == target {
		return from
	}
	return result
}

func calendarWeeks(month Date, weekStart time.Weekday, fixed bool) int {
	if fixed {
		return 6
	}
	offset := (int(month.dayTime().Weekday()) - int(weekStart) + 7) % 7
	days := month.AddMonths(1).AddDays(-1).Day
	return (offset + days + 6) / 7
}

func calendarSpanDays(start, end Date) int {
	days := int((end.dayTime().Unix() - start.dayTime().Unix()) / 86400)
	if days < 0 {
		days = -days
	}
	return days + 1
}

func calendarWithinRangeLimit(date Date, value DateSelection, o CalendarOptions) bool {
	if !o.Range || o.MaximumDays == 0 || !value.Start.Set || !value.Start.Value.Valid() || value.End.Set && (!value.End.Value.Valid() || o.FixedDate == CalendarNoFixedDate) {
		return true
	}
	anchor := value.Start.Value
	if value.End.Set && (o.FixedDate == CalendarFixedEnd && date.Compare(value.End.Value) <= 0 || o.FixedDate == CalendarFixedStart && date.Compare(value.Start.Value) < 0) {
		anchor = value.End.Value
	}
	return calendarSpanDays(anchor, date) <= o.MaximumDays
}

// WeekStart is an explicit native setting. Sunday uses Jan 1 as week one's
// reference; other starts use Jan 4, including the ISO Monday convention.
func calendarWeekNumber(date Date, weekStart time.Weekday) int {
	startOfWeek := func(date Date) Date { return date.AddDays(-(int(date.dayTime().Weekday()) - int(weekStart) + 7) % 7) }
	reference := 4
	if weekStart == time.Sunday {
		reference = 1
	}
	first := startOfWeek(Date{date.Year, 1, reference})
	current := startOfWeek(date)
	if current.Compare(first) < 0 {
		first = startOfWeek(Date{date.Year - 1, 1, reference})
	}
	if next := startOfWeek(Date{date.Year + 1, 1, reference}); current.Compare(next) >= 0 {
		first = next
	}
	return int((current.dayTime().Unix()-first.dayTime().Unix())/(7*86400)) + 1
}
