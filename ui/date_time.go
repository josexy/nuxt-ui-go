package nui

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// Date is a civil Gregorian date; it has no timezone or time of day.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func (d Date) Valid() bool {
	if d.Year < 1 || d.Year > 9999 || d.Month < 1 || d.Month > 12 || d.Day < 1 {
		return false
	}
	v := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
	return v.Year() == d.Year && v.Month() == d.Month && v.Day() == d.Day
}
func (d Date) String() string {
	if !d.Valid() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}
func DateFromTime(v time.Time) Date { y, m, d := v.Date(); return Date{y, m, d} }
func (d Date) dayTime() time.Time   { return time.Date(d.Year, d.Month, d.Day, 12, 0, 0, 0, time.UTC) }
func (d Date) AddDays(n int) Date   { return DateFromTime(d.dayTime().AddDate(0, 0, n)) }
func (d Date) AddMonths(n int) Date {
	first := time.Date(d.Year, d.Month, 1, 12, 0, 0, 0, time.UTC).AddDate(0, n, 0)
	last := first.AddDate(0, 1, -1).Day()
	return Date{first.Year(), first.Month(), min(d.Day, last)}
}
func (d Date) Compare(other Date) int {
	a, b := d.Year*10000+int(d.Month)*100+d.Day, other.Year*10000+int(other.Month)*100+other.Day
	return a - b
}
func ParseDate(text string) (Optional[Date], bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Optional[Date]{}, true
	}
	v, err := time.Parse("2006-01-02", text)
	if err != nil {
		return Optional[Date]{}, false
	}
	d := DateFromTime(v)
	return Set(d), d.Valid()
}

type DateSelection struct{ Start, End Optional[Date] }
type CalendarState struct {
	Month, Active        Date
	View                 CalendarView
	Handle               ui.Handle
	seen                 DateSelection
	initialized          bool
	displayedMonth       Date
	yearStart, yearCount int
}
type CalendarOptions struct {
	Label                                                            string
	Range, Disabled                                                  bool
	ReadOnly                                                         bool
	Color                                                            Color
	Variant                                                          Variant
	Size                                                             Size
	Min, Max                                                         Optional[Date]
	IsDateDisabled                                                   func(Date) bool
	IsDateUnavailable                                                func(Date) bool
	IsDateHighlightable                                              func(Date) bool
	AllowNonContiguousRanges                                         bool
	PreventDeselect                                                  bool
	MaximumDays                                                      int // Inclusive range length; 0 is unlimited.
	FixedDate                                                        CalendarFixedDate
	WeekStart                                                        time.Weekday
	WeekdayFormat                                                    CalendarWeekdayFormat
	WeekNumbers                                                      bool
	FixedWeeks                                                       Optional[bool]
	NumberOfMonths                                                   int // 0 uses one month; at most 12.
	PagedNavigation                                                  bool
	PreviousPage, NextPage                                           func(Date) Date
	DisableOutsideDays                                               bool
	MonthControls, YearControls                                      Optional[bool]
	ViewControl                                                      Optional[bool]
	ViewControlOptions                                               *ButtonOptions
	YearsPerPage                                                     int // 0 uses 12; at most 120.
	IsMonthDisabled, IsMonthUnavailable                              func(Date) bool
	IsYearDisabled, IsYearUnavailable                                func(Date) bool
	PreviousMonth, NextMonth, PreviousYear, NextYear                 *ButtonOptions
	PreviousMonthIcon, NextMonthIcon, PreviousYearIcon, NextYearIcon *ui.SVG
	Today                                                            Optional[Date]
	OnChange                                                         func(DateSelection)
	OnMonthChange                                                    func(Date)
	OnViewChange                                                     func(CalendarView)
	BuildDay, BuildMonth, BuildYear                                  func(*ui.Context, Theme, Date)
	BuildWeekday                                                     func(*ui.Context, Theme, time.Weekday)
	BuildHeading                                                     func(*ui.Context, Theme, Date, CalendarView)
	Decoration
}

type CalendarView uint8

const (
	CalendarDayView CalendarView = iota
	CalendarMonthView
	CalendarYearView
)

type CalendarFixedDate uint8

const (
	CalendarNoFixedDate CalendarFixedDate = iota
	CalendarFixedStart
	CalendarFixedEnd
)

type CalendarWeekdayFormat uint8

const (
	CalendarWeekdayNarrow CalendarWeekdayFormat = iota
	CalendarWeekdayShort
	CalendarWeekdayLong
)

func dateInBounds(d Date, o CalendarOptions) bool {
	return d.Valid() && (!o.Min.Set || d.Compare(o.Min.Value) >= 0) && (!o.Max.Set || d.Compare(o.Max.Value) <= 0)
}

func dateNavigable(d Date, o CalendarOptions) bool {
	return dateInBounds(d, o) && (o.IsDateDisabled == nil || !o.IsDateDisabled(d))
}

func dateEnabled(d Date, o CalendarOptions) bool {
	return dateNavigable(d, o) && (o.IsDateUnavailable == nil || !o.IsDateUnavailable(d))
}
func validateCalendar(o CalendarOptions) {
	if o.Min.Set && !o.Min.Value.Valid() || o.Max.Set && !o.Max.Value.Valid() || o.Min.Set && o.Max.Set && o.Min.Value.Compare(o.Max.Value) > 0 || o.WeekStart < 0 || o.WeekStart > 6 {
		panic("nui: invalid calendar bounds")
	}
	if o.WeekdayFormat > CalendarWeekdayLong {
		panic("nui: invalid calendar weekday format")
	}
	if o.NumberOfMonths < 0 || o.NumberOfMonths > 12 {
		panic("nui: Calendar supports 1 to 12 displayed months")
	}
	if o.YearsPerPage < 0 || o.YearsPerPage > 120 {
		panic("nui: Calendar supports 1 to 120 years per page")
	}
	if o.MaximumDays < 0 || o.FixedDate > CalendarFixedEnd {
		panic("nui: invalid calendar range options")
	}
}

// SelectDate rejects ranges crossing a disabled day; reverse selection is normalized.
func SelectDate(value DateSelection, date Date, o CalendarOptions) (DateSelection, bool) {
	validateCalendar(o)
	if !dateEnabled(date, o) || o.Disabled || o.ReadOnly {
		return value, false
	}
	if !o.Range {
		if !o.PreventDeselect && value.Start == Set(date) {
			return DateSelection{}, true
		}
		return DateSelection{Start: Set(date)}, true
	}
	if !o.PreventDeselect && value.Start == Set(date) && (!value.End.Set || value.End == value.Start) {
		return DateSelection{}, true
	}
	if !value.Start.Set || !value.Start.Value.Valid() || value.End.Set && (!value.End.Value.Valid() || o.FixedDate == CalendarNoFixedDate) {
		return DateSelection{Start: Set(date)}, true
	}
	start, end := value.Start.Value, date
	if value.End.Set && o.FixedDate == CalendarFixedEnd {
		start, end = date, value.End.Value
	}
	if value.End.Set && o.FixedDate == CalendarFixedStart && date.Compare(start) < 0 {
		start, end = date, value.End.Value
	}
	if value.End.Set && o.FixedDate == CalendarFixedEnd && date.Compare(end) > 0 {
		start, end = value.Start.Value, date
	}
	if start.Compare(end) > 0 {
		start, end = end, start
	}
	if !dateEnabled(start, o) || !dateEnabled(end, o) {
		return value, false
	}
	if o.MaximumDays > 0 && calendarSpanDays(start, end) > o.MaximumDays {
		return value, false
	}
	if o.IsDateDisabled == nil && (o.IsDateUnavailable == nil || o.AllowNonContiguousRanges) {
		return DateSelection{Set(start), Set(end)}, true
	}
	for day := start; day.Compare(end) <= 0; day = day.AddDays(1) {
		if day != start && day != end && o.IsDateHighlightable != nil && o.IsDateHighlightable(day) {
			continue
		}
		if !dateNavigable(day, o) || !o.AllowNonContiguousRanges && o.IsDateUnavailable != nil && o.IsDateUnavailable(day) {
			return value, false
		}
	}
	return DateSelection{Set(start), Set(end)}, true
}
func Calendar(c *ui.Context, t Theme, value *DateSelection, s *CalendarState, o CalendarOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: Calendar requires value and state")
	}
	validateCalendar(o)
	if s.View > CalendarYearView {
		panic("nui: invalid calendar view")
	}
	config := selected(defaults(t, ComponentCalendar, ComponentTheme{Color: Primary, Variant: Solid, Size: MD}), ComponentTheme{Color: o.Color, Variant: o.Variant, Size: inheritedSize(t, o.Size)})
	if config.Variant != Solid && config.Variant != Outline && config.Variant != Soft && config.Variant != Subtle {
		panic("nui: Calendar supports solid, outline, soft and subtle")
	}
	if t.field.Error != "" {
		config.Color = Error
	}
	if o.Label == "" {
		o.Label = t.field.Label
	}
	m := calendarRecipe(t, config.Size)
	months := max(1, o.NumberOfMonths)
	if s.View != CalendarDayView {
		months = 1
	}
	columns := 7
	if o.WeekNumbers && s.View == CalendarDayView {
		columns++
	}
	today := DateFromTime(time.Now())
	if o.Today.Set && o.Today.Value.Valid() {
		today = o.Today.Value
	}
	if s.initialized && s.seen != *value {
		target := value.Start
		if value.Start == s.seen.Start && value.End.Set {
			target = value.End
		}
		if target.Set && target.Value.Valid() {
			s.Active = target.Value
			s.Month = calendarMonth(s.Active)
		}
	} else if s.initialized && s.Month.Valid() && calendarMonth(s.Month) != s.displayedMonth {
		s.Active = calendarMonth(s.Month)
	}
	s.seen, s.initialized = *value, true
	if !s.Active.Valid() {
		s.Active = today
		if value.Start.Set && value.Start.Value.Valid() {
			s.Active = value.Start.Value
		}
		if s.Month.Valid() && calendarMonth(s.Active) != calendarMonth(s.Month) {
			s.Active = calendarMonth(s.Month)
		}
	}
	if o.Min.Set && s.Active.Compare(o.Min.Value) < 0 {
		s.Active = o.Min.Value
	}
	if o.Max.Set && s.Active.Compare(o.Max.Value) > 0 {
		s.Active = o.Max.Value
	}
	if !s.Month.Valid() {
		s.Month = calendarMonth(s.Active)
	}
	s.Month = calendarMonth(s.Month)
	if s.View == CalendarDayView {
		if s.Active.Compare(s.Month) < 0 {
			s.Month = calendarMonth(s.Active)
		}
		if s.Active.Compare(s.Month.AddMonths(months)) >= 0 {
			s.Month = calendarMonth(s.Active).AddMonths(1 - months)
		}
	}
	s.displayedMonth = s.Month
	months = min(months, (9999-s.Month.Year)*12+13-int(s.Month.Month))
	gridGap := valueOr(mergeStyle(t.components[ComponentCalendar].UI[GridSlot], o.UI[GridSlot]).Gap, t.spacing(16))
	rootPadding := valueOr(mergeStyle(t.components[ComponentCalendar].UI[Root], o.UI[Root]).PaddingX, float32(0))
	bodyPadding := valueOr(mergeStyle(t.components[ComponentCalendar].UI[Body], o.UI[Body]).PaddingX, float32(0))
	root := ui.Column(c).MinWidth(0).MaxWidthPercent(100).Width(m.column*float32(columns*months) + gridGap*float32(months-1) + 2*(rootPadding+bodyPadding)).Disabled(o.Disabled || t.field.Disabled)
	finish(t, ComponentCalendar, Root, root, PartStyle{Font: Set(t.Font), FontSize: Set(m.font), LineHeight: Set(m.line / m.font)}, o.Decoration)
	disabled := root.IsDisabled()
	controls := ui.Local(root, "header-focus", func() [5]ui.Handle { return [5]ui.Handle{} })
	setView := func(view CalendarView) {
		if disabled || s.View == view {
			return
		}
		s.View = view
		s.Handle.Focus(c)
		if o.OnViewChange != nil {
			o.OnViewChange(view)
		}
	}
	changeMonth := func(month, date Date) {
		if disabled || !month.Valid() || !date.Valid() {
			return
		}
		before := s.Month
		s.Active, s.Month = date, calendarMonth(month)
		s.displayedMonth = s.Month
		if before != s.Month && o.OnMonthChange != nil {
			o.OnMonthChange(s.Month)
		}
	}
	move := func(date Date) {
		month := s.Month
		if date.Compare(month) < 0 {
			month = calendarMonth(date)
		}
		if date.Compare(month.AddMonths(months)) >= 0 {
			month = calendarMonth(date).AddMonths(1 - months)
		}
		changeMonth(month, date)
	}
	choose := func(date Date) {
		if disabled {
			return
		}
		if next, ok := SelectDate(*value, date, o); ok {
			changed := next != *value
			*value, s.seen = next, next
			move(date)
			if changed && o.OnChange != nil {
				o.OnChange(next)
			}
		}
	}
	root.Children(func() {
		var headerParts [5]ui.Element
		header := ui.Row(c.Key("header")).MinWidth(0).AlignItems(ui.Center)
		finish(t, ComponentCalendar, Header, header, PartStyle{}, o.Decoration)
		header.Children(func() {
			step := 1
			if o.PagedNavigation {
				step = months
			}
			if s.View == CalendarMonthView {
				step = 12
			}
			if s.View == CalendarYearView {
				step = calendarYearsPerPage(o) * 12
				calendarYearPage(s, o)
			}
			previousIcon, nextIcon := lucide.ChevronLeft(), lucide.ChevronRight()
			previousYearIcon, nextYearIcon := lucide.ChevronsLeft(), lucide.ChevronsRight()
			if c.Direction() == ui.RightToLeft {
				previousIcon, nextIcon = nextIcon, previousIcon
				previousYearIcon, nextYearIcon = nextYearIcon, previousYearIcon
			}
			if o.PreviousMonthIcon != nil {
				previousIcon = o.PreviousMonthIcon
			}
			if o.NextMonthIcon != nil {
				nextIcon = o.NextMonthIcon
			}
			if o.PreviousYearIcon != nil {
				previousYearIcon = o.PreviousYearIcon
			}
			if o.NextYearIcon != nil {
				nextYearIcon = o.NextYearIcon
			}
			control := func(key, name string, months int, icon *ui.SVG, opts *ButtonOptions) {
				target := s.Month.AddMonths(months)
				custom := false
				if key == "previous" && o.PreviousPage != nil {
					target, custom = o.PreviousPage(s.Month), true
				}
				if key == "next" && o.NextPage != nil {
					target, custom = o.NextPage(s.Month), true
				}
				enabled := calendarMonthEnabled(target, o)
				if s.View != CalendarDayView {
					enabled = target.Valid() && calendarYearPageEnabled(target.Year, 1, o)
				}
				targetYear := s.yearStart
				if s.View == CalendarYearView {
					targetYear = max(1, min(9999, s.yearStart+months/12))
					if custom {
						targetYear = target.Year
					} else {
						target = Date{max(1, min(9999, s.Month.Year+months/12)), s.Month.Month, 1}
					}
					enabled = target.Valid() && targetYear != s.yearStart && calendarYearPageEnabled(targetYear, calendarYearsPerPage(o), o)
				}
				index := map[string]int{"previous-year": 0, "previous": 1, "next": 3, "next-year": 4}[key]
				headerParts[index] = calendarNavigationButton(c.Key(key), t, config.Size, name, icon, opts, enabled, func() {
					if s.View == CalendarYearView {
						s.yearStart = targetYear
					}
					active := target
					if o.Min.Set && active.Compare(o.Min.Value) < 0 {
						active = o.Min.Value
					}
					changeMonth(target, active)
				}).Bind(&controls[index])
			}
			previousName, nextName := "previousMonth", "nextMonth"
			if s.View != CalendarDayView {
				previousName, nextName = "previousYear", "nextYear"
			}
			if s.View == CalendarDayView && valueOr(o.YearControls, true) {
				control("previous-year", "previousYear", -12, previousYearIcon, o.PreviousYear)
			}
			if s.View != CalendarDayView || valueOr(o.MonthControls, true) {
				control("previous", previousName, -step, previousIcon, o.PreviousMonth)
			}
			title := i18n.FormatDate(c.Locale(), s.Month.dayTime(), i18n.DateOptions{Skeleton: "yMMMM"})
			if months > 1 {
				title += " – " + i18n.FormatDate(c.Locale(), s.Month.AddMonths(months-1).dayTime(), i18n.DateOptions{Skeleton: "yMMMM"})
			}
			if s.View == CalendarMonthView {
				title = i18n.FormatDate(c.Locale(), s.Month.dayTime(), i18n.DateOptions{Skeleton: "y"})
			}
			if s.View == CalendarYearView {
				title = fmt.Sprintf("%d – %d", s.yearStart, min(9999, s.yearStart+calendarYearsPerPage(o)-1))
			}
			if s.View == CalendarYearView && calendarYearsPerPage(o) == 1 {
				title = fmt.Sprint(s.yearStart)
			}
			if o.BuildHeading != nil {
				heading := ui.Row(c.Key("heading")).Grow(1).MinWidth(0).AlignItems(ui.Center)
				finish(t, ComponentCalendar, HeadingSlot, heading, PartStyle{}, o.Decoration)
				heading.Children(func() { o.BuildHeading(c, t, s.Month, s.View) })
			} else if valueOr(o.ViewControl, true) {
				options := ButtonOptions{}
				if o.ViewControlOptions != nil {
					options = *o.ViewControlOptions
				}
				if options.Color == ColorDefault {
					options.Color = Neutral
				}
				if options.Variant == VariantDefault {
					options.Variant = Ghost
				}
				if options.Size == SizeDefault {
					options.Size = config.Size
				}
				style := options.Style
				options.Style = func(root, label, leading, trailing ui.Element) {
					finish(t, ComponentCalendar, HeadingSlot, root, PartStyle{}, o.Decoration)
					if style != nil {
						style(root, label, leading, trailing)
					}
				}
				headerParts[2] = Button(c.Key("heading"), t, title, options).Bind(&controls[2]).Grow(1).MinWidth(0).Justify(ui.Center).OnClick(func() { setView((s.View + 1) % 3) })
			} else {
				heading := ui.Text(c.Key("heading"), title).SingleLine().MinWidth(0).Grow(1).TextAlign(ui.Center)
				finish(t, ComponentCalendar, HeadingSlot, heading, PartStyle{Weight: Set(500), PaddingX: Set(t.spacing(6)), PaddingY: Set(t.spacing(6))}, o.Decoration)
			}
			if s.View != CalendarDayView || valueOr(o.MonthControls, true) {
				control("next", nextName, step, nextIcon, o.NextMonth)
			}
			if s.View == CalendarDayView && valueOr(o.YearControls, true) {
				control("next-year", "nextYear", 12, nextYearIcon, o.NextYear)
			}
		})
		for i, part := range headerParts {
			if !part.Valid() || part.IsDisabled() {
				if controls[i].Focused(c) && !disabled {
					s.Handle.Focus(c)
				}
				controls[i].CancelFocus(c)
			}
		}
		body := ui.Column(c.Key("body")).MinWidth(0)
		padding := t.spacing(16)
		if config.Size == XS {
			padding = t.spacing(8)
		}
		body.Padding(padding, 0, 0, 0)
		finish(t, ComponentCalendar, Body, body, PartStyle{}, o.Decoration)
		body.Children(func() {
			if s.View != CalendarDayView {
				calendarPicker(c, t, s, o, config, today, disabled, changeMonth, setView)
				return
			}
			// A grid, as Reka UI's CalendarGrid: its cells take the arrow keys
			// and report read-only, which a table's cells do not.
			grid := ui.Row(c.Key("grid")).Wrap().AlignItems(ui.Start).Focusable().FocusRing(false).Bind(&s.Handle).Role(ui.RoleGrid).Label(o.Label).ReadOnly(o.ReadOnly).InputMethod(false)
			bindField(c, grid, t)
			finish(t, ComponentCalendar, GridSlot, grid, PartStyle{Gap: Set(t.spacing(16))}, o.Decoration)
			hover := ui.Local(grid, "hover-date", func() Date { return Date{} })
			if !grid.Hovered() {
				*hover = Date{}
			}
			direction := 1
			if c.Direction() == ui.RightToLeft {
				direction = -1
			}
			grid.HandleInput(func(ev ui.InputEvent) bool {
				if disabled || ev.Kind != ui.InputKeyDown || ev.Mods & ^ui.Shift != 0 {
					return false
				}
				if ev.Mods == ui.Shift && ev.Key != ui.KeyPageUp && ev.Key != ui.KeyPageDown {
					return false
				}
				next := s.Active
				months := 1
				if ev.Mods == ui.Shift {
					months = 12
				}
				switch ev.Key {
				case ui.KeyLeft:
					next = calendarNextDate(next, -direction, o)
				case ui.KeyRight:
					next = calendarNextDate(next, direction, o)
				case ui.KeyUp:
					next = calendarNextDate(next, -7, o)
				case ui.KeyDown:
					next = calendarNextDate(next, 7, o)
				case ui.KeyPageUp:
					next = calendarPageDate(next, -months, o)
				case ui.KeyPageDown:
					next = calendarPageDate(next, months, o)
				case ui.KeyHome:
					next = calendarWeekBoundary(next, false, o)
				case ui.KeyEnd:
					next = calendarWeekBoundary(next, true, o)
				case ui.KeyEnter, ui.KeySpace:
					*hover = Date{}
					choose(next)
					return true
				default:
					return false
				}
				*hover = Date{}
				move(next)
				return true
			})
			preview := *value
			if o.Range && value.Start.Set && (!value.End.Set || o.FixedDate != CalendarNoFixedDate) && !disabled && !o.ReadOnly {
				date := s.Active
				if hover.Valid() {
					date = *hover
				}
				if next, ok := SelectDate(*value, date, o); ok && next.End.Set {
					preview = next
				}
			}
			grid.Children(func() {
				for monthIndex := range months {
					month := s.Month.AddMonths(monthIndex)
					ui.Column(c.Key(month)).Width(m.column * float32(columns)).Shrink(0).Gap(m.gap).Children(func() {
						weekdays := i18n.Weekdays(c.Locale(), calendarWeekdayWidth(o.WeekdayFormat), true)
						weekdayNames := i18n.Weekdays(c.Locale(), i18n.Wide, true)
						ui.Row(c.Key("weekdays")).Role(ui.RoleRow).Children(func() {
							if o.WeekNumbers {
								ui.Box(c.Key("week-number")).Width(m.column)
							}
							for i := range 7 {
								day := time.Weekday((i + int(o.WeekStart)) % 7)
								var cell ui.Element
								if o.BuildWeekday == nil {
									cell = ui.Text(c.Key(day), weekdays[day]).SingleLine().TextAlign(ui.Center)
								} else {
									cell = ui.Box(c.Key(day)).Center()
								}
								cell.Width(m.column).MinWidth(0).Clip().Role(ui.RoleColumnHeader).Label(weekdayNames[day])
								finish(t, ComponentCalendar, WeekDaySlot, cell, PartStyle{Foreground: Set(t.color(config.Color)), FontSize: Set(m.weekday)}, o.Decoration)
								if o.BuildWeekday != nil {
									cell.Children(func() { o.BuildWeekday(c, t, day) })
								}
							}
						})
						offset := (int(month.dayTime().Weekday()) - int(o.WeekStart) + 7) % 7
						first := month.AddDays(-offset)
						for week := range calendarWeeks(month, o.WeekStart, valueOr(o.FixedWeeks, true)) {
							row := ui.Row(c.Key(week)).Role(ui.RoleRow)
							finish(t, ComponentCalendar, GridRowSlot, row, PartStyle{}, o.Decoration)
							row.Children(func() {
								if o.WeekNumbers {
									weekNumber := calendarWeekNumber(first.AddDays(week*7), o.WeekStart)
									part := ui.Box(c.Key("week-number")).Width(m.column).Height(m.cell + t.spacing(4)).Center().Role(ui.RoleCell).Label(formatMessage(c, "calendarWeekNumber", map[string]any{"number": weekNumber}))
									finish(t, ComponentCalendar, WeekNumberSlot, part, PartStyle{Foreground: Set(t.Muted), FontSize: Set(m.font)}, o.Decoration)
									part.Children(func() { ui.Textf(c, "%d", weekNumber) })
								}
								for day := range 7 {
									date := first.AddDays(week*7 + day)
									if !date.Valid() {
										ui.Box(c.Key(date)).Size(m.cell, m.cell).Margin(t.spacing(2))
										continue
									}
									outside := date.Month != month.Month || date.Year != month.Year
									unavailable := o.IsDateUnavailable != nil && o.IsDateUnavailable(date)
									enabled := dateNavigable(date, o) && !unavailable && (!o.DisableOutsideDays || !outside) && calendarWithinRangeLimit(date, *value, o)
									chosen := value.Start == Set(date) || o.Range && (value.End == Set(date) || value.Start.Set && value.End.Set && date.Compare(value.Start.Value) >= 0 && date.Compare(value.End.Value) <= 0)
									between := o.Range && preview.Start.Set && preview.End.Set && date.Compare(preview.Start.Value) >= 0 && date.Compare(preview.End.Value) <= 0
									cell := ui.Box(c.Key(date)).Size(m.cell, m.cell).Margin(t.spacing(2)).Center().Role(ui.RoleCell).Label(strings.TrimSpace(o.Label + " " + date.String())).Selected(chosen).ReadOnly(o.ReadOnly).Disabled(!enabled).OnClick(func() {
										*hover = Date{}
										if o.ReadOnly {
											move(date)
										} else {
											choose(date)
										}
										s.Handle.Focus(c)
									})
									if o.ReadOnly {
										cell.Cursor(ui.CursorDefault)
									}
									cell.HandleInput(func(ev ui.InputEvent) bool {
										if ev.Kind == ui.InputPointerMove && o.Range && value.Start.Set && (!value.End.Set || o.FixedDate != CalendarNoFixedDate) && !disabled && !o.ReadOnly && enabled && *hover != date {
											*hover = date
											return true
										}
										return false
									})
									focused := date == s.Active && grid.FocusVisible()
									style := calendarCellStyle(t, config, chosen, between, cell.Hovered() && enabled && !disabled, date == today, outside, focused)
									if !enabled {
										style.Foreground = Set(t.Muted)
									}
									if date == s.Active && (!outside || months == 1) {
										grid.ActiveDescendant(cell)
									}
									if focused {
										cell.Shadow(0, 0, 0, 3, inputFocusColor(t, config.Color))
									}
									finish(t, ComponentCalendar, Item, cell, style, o.Decoration)
									cell.Children(func() {
										if o.BuildDay != nil {
											o.BuildDay(c, t, date)
											return
										}
										text := ui.Textf(c, "%d", date.Day)
										if unavailable {
											text.Strikethrough()
										}
									})
								}
							})
						}
					})
				}
			})
		})
	})
	return root
}

// ClockTime is a wall-clock value, not an instant. Conversion to a timezone is application-owned.
type ClockTime struct{ Hour, Minute, Second int }

func (v ClockTime) Valid() bool {
	return v.Hour >= 0 && v.Hour < 24 && v.Minute >= 0 && v.Minute < 60 && v.Second >= 0 && v.Second < 60
}
func (v ClockTime) String() string { return fmt.Sprintf("%02d:%02d:%02d", v.Hour, v.Minute, v.Second) }
func ParseClockTime(text string, twelve, seconds bool) (Optional[ClockTime], bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Optional[ClockTime]{}, true
	}
	layout := "15:04"
	if seconds {
		layout += ":05"
	}
	if twelve {
		layout = "03:04"
		if seconds {
			layout += ":05"
		}
		layout += " PM"
	}
	v, err := time.Parse(layout, strings.ToUpper(text))
	if err != nil {
		return Optional[ClockTime]{}, false
	}
	return Set(ClockTime{v.Hour(), v.Minute(), v.Second()}), true
}

// localeTwelveHour reports whether the current locale, with the user's
// regional formats, writes times with AM and PM: whether its preferred hour
// ("j") is written as the 12-hour cycle ("h") is.
func localeTwelveHour(c *ui.Context) bool {
	locale := c.Region().FormatLocale(c.Locale())
	at := time.Date(2000, 1, 1, 13, 0, 0, 0, time.UTC)
	return i18n.FormatDate(locale, at, i18n.DateOptions{Skeleton: "j"}) == i18n.FormatDate(locale, at, i18n.DateOptions{Skeleton: "h"})
}
