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
	Month, Active Date
	Handle        ui.Handle
	seen          DateSelection
	initialized   bool
}
type CalendarOptions struct {
	Label           string
	Range, Disabled bool
	Min, Max        Optional[Date]
	IsDateDisabled  func(Date) bool
	WeekStart       time.Weekday
	Today           Optional[Date]
	OnChange        func(DateSelection)
	Decoration
}

func dateEnabled(d Date, o CalendarOptions) bool {
	return d.Valid() && (!o.Min.Set || d.Compare(o.Min.Value) >= 0) && (!o.Max.Set || d.Compare(o.Max.Value) <= 0) && (o.IsDateDisabled == nil || !o.IsDateDisabled(d))
}
func validateCalendar(o CalendarOptions) {
	if o.Min.Set && !o.Min.Value.Valid() || o.Max.Set && !o.Max.Value.Valid() || o.Min.Set && o.Max.Set && o.Min.Value.Compare(o.Max.Value) > 0 || o.WeekStart < 0 || o.WeekStart > 6 {
		panic("nui: invalid calendar bounds")
	}
}

// SelectDate rejects ranges crossing a disabled day; reverse selection is normalized.
func SelectDate(value DateSelection, date Date, o CalendarOptions) (DateSelection, bool) {
	validateCalendar(o)
	if !dateEnabled(date, o) || o.Disabled {
		return value, false
	}
	if !o.Range {
		return DateSelection{Start: Set(date)}, true
	}
	if !value.Start.Set || !value.Start.Value.Valid() || value.End.Set {
		return DateSelection{Start: Set(date)}, true
	}
	start, end := value.Start.Value, date
	if start.Compare(end) > 0 {
		start, end = end, start
	}
	for day := start; day.Compare(end) <= 0; day = day.AddDays(1) {
		if !dateEnabled(day, o) {
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
	today := DateFromTime(time.Now())
	if o.Today.Set && o.Today.Value.Valid() {
		today = o.Today.Value
	}
	if s.initialized && s.seen != *value {
		// Follow an externally changed endpoint without resetting ordinary
		// month browsing or a range selection made inside this calendar.
		target := value.Start
		if value.Start == s.seen.Start && value.End.Set {
			target = value.End
		}
		if target.Set && target.Value.Valid() {
			s.Active = target.Value
			s.Month = Date{s.Active.Year, s.Active.Month, 1}
		}
	}
	s.seen, s.initialized = *value, true
	if !s.Active.Valid() {
		s.Active = today
		if value.Start.Set && value.Start.Value.Valid() {
			s.Active = value.Start.Value
		}
		if o.Min.Set && s.Active.Compare(o.Min.Value) < 0 {
			s.Active = o.Min.Value
		}
		if o.Max.Set && s.Active.Compare(o.Max.Value) > 0 {
			s.Active = o.Max.Value
		}
	}
	if !s.Month.Valid() {
		s.Month = Date{s.Active.Year, s.Active.Month, 1}
	}
	root := ui.Column(c).MinWidth(0).Width(t.text(294)).Disabled(o.Disabled || t.field.Disabled)
	finish(t, ComponentCalendar, Root, root, PartStyle{Gap: Set(t.spacing(8)), PaddingX: Set(t.spacing(8)), PaddingY: Set(t.spacing(8)), Background: Set(t.Background), Radius: Set(t.radius(8))}, o.Decoration)
	choose := func(d Date) {
		if next, ok := SelectDate(*value, d, o); ok {
			*value = next
			s.seen = next
			s.Active = d
			if o.OnChange != nil {
				o.OnChange(next)
			}
		}
	}
	root.Children(func() {
		ui.Row(c.Key("header")).AlignItems(ui.Center).Children(func() {
			previous, next := s.Month.AddMonths(-1), s.Month.AddMonths(1)
			// Earlier months are toward the start of the line: the right, right to left.
			previousIcon, nextIcon := lucide.ChevronLeft(), lucide.ChevronRight()
			if c.Direction() == ui.RightToLeft {
				previousIcon, nextIcon = nextIcon, previousIcon
			}
			Button(c.Key("previous"), t, "", ButtonOptions{AccessibleLabel: message(c, "previousMonth"), Leading: previousIcon, Variant: Ghost, Size: SM, Disabled: !previous.Valid() || o.Min.Set && previous.AddMonths(1).AddDays(-1).Compare(o.Min.Value) < 0}).OnClick(func() { s.Month = previous; s.Active = previous })
			ui.Text(c, i18n.FormatDate(c.Locale(), s.Month.dayTime(), i18n.DateOptions{Skeleton: "yMMMM"})).Grow(1).TextAlign(ui.Center).FontWeight(600)
			Button(c.Key("next"), t, "", ButtonOptions{AccessibleLabel: message(c, "nextMonth"), Leading: nextIcon, Variant: Ghost, Size: SM, Disabled: !next.Valid() || o.Max.Set && next.Compare(o.Max.Value) > 0}).OnClick(func() { s.Month = next; s.Active = next })
		})
		ui.Row(c.Key("weekdays")).Children(func() {
			weekdays := i18n.Weekdays(c.Locale(), i18n.Short, true)
			for i := range 7 {
				day := time.Weekday((i + int(o.WeekStart)) % 7)
				ui.Text(c, weekdays[day]).Width(t.text(38)).TextAlign(ui.Center).TextColor(t.Primary).FontSize(t.text(12))
			}
		})
		grid := ui.Column(c.Key("grid")).Focusable().Bind(&s.Handle).Role(ui.RoleTable).Label(o.Label).InputMethod(false)
		direction := 1
		if c.Direction() == ui.RightToLeft {
			direction = -1
		}
		grid.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind != ui.InputKeyDown || ev.Mods != 0 {
				return false
			}
			next := s.Active
			switch ev.Key {
			case ui.KeyLeft:
				next = next.AddDays(-direction)
			case ui.KeyRight:
				next = next.AddDays(direction)
			case ui.KeyUp:
				next = next.AddDays(-7)
			case ui.KeyDown:
				next = next.AddDays(7)
			case ui.KeyPageUp:
				next = next.AddMonths(-1)
			case ui.KeyPageDown:
				next = next.AddMonths(1)
			case ui.KeyHome:
				next = Date{next.Year, next.Month, 1}
			case ui.KeyEnd:
				next = Date{next.Year, next.Month, 1}.AddMonths(1).AddDays(-1)
			case ui.KeyEnter, ui.KeySpace:
				choose(next)
				return true
			default:
				return false
			}
			if next.Valid() {
				s.Active = next
				s.Month = Date{next.Year, next.Month, 1}
			}
			return true
		})
		grid.Children(func() {
			offset := (int(s.Month.dayTime().Weekday()) - int(o.WeekStart) + 7) % 7
			first := s.Month.AddDays(-offset)
			for week := range 6 {
				ui.Row(c.Key(week)).Role(ui.RoleRow).Children(func() {
					for day := range 7 {
						date := first.AddDays(week*7 + day)
						enabled := dateEnabled(date, o)
						selected := value.Start.Set && date == value.Start.Value || value.End.Set && date == value.End.Value
						between := value.Start.Set && value.End.Set && date.Compare(value.Start.Value) > 0 && date.Compare(value.End.Value) < 0
						cell := ui.Box(c.Key(date)).Size(t.text(32), t.text(32)).Margin(t.text(1), t.text(3)).Center().Role(ui.RoleCell).Label(strings.TrimSpace(o.Label + " " + date.String())).Selected(selected).Disabled(!enabled).OnClick(func() { choose(date); s.Handle.Focus(c) })
						style := PartStyle{Radius: Set(float32(999)), Foreground: Set(t.Text)}
						if date.Month != s.Month.Month {
							style.Foreground = Set(t.Dimmed)
						}
						if between {
							style.Background = Set(t.Primary.Alpha(.1))
						}
						if selected {
							style.Background = Set(t.Primary)
							style.Foreground = Set(t.Inverted)
						}
						if date == today && !selected {
							style.Foreground = Set(t.Primary)
							style.Weight = Set(600)
						}
						if cell.Hovered() && !selected && !between {
							style.Background = Set(t.Primary.Alpha(.1))
						}
						if date == s.Active {
							grid.ActiveDescendant(cell)
							if grid.FocusVisible() {
								style.BorderWidth = Set(float32(1))
								style.BorderColor = Set(t.Primary)
							}
						}
						finish(t, ComponentCalendar, Item, cell, style, o.Decoration)
						cell.Children(func() { ui.Textf(c, "%d", date.Day) })
					}
				})
			}
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
