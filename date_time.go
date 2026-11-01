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
			Button(c.Key("previous"), t, "", ButtonOptions{AccessibleLabel: message(c, "previousMonth"), Leading: lucide.ChevronDown(), Variant: Ghost, Size: SM, Disabled: !previous.Valid() || o.Min.Set && previous.AddMonths(1).AddDays(-1).Compare(o.Min.Value) < 0, Style: func(root, label, leading, trailing ui.Element) { leading.Rotate(90) }}).OnClick(func() { s.Month = previous; s.Active = previous })
			ui.Text(c, i18n.FormatDate(c.Locale(), s.Month.dayTime(), i18n.DateOptions{Skeleton: "yMMMM"})).Grow(1).TextAlign(ui.Center).FontWeight(600)
			Button(c.Key("next"), t, "", ButtonOptions{AccessibleLabel: message(c, "nextMonth"), Leading: lucide.ChevronDown(), Variant: Ghost, Size: SM, Disabled: !next.Valid() || o.Max.Set && next.Compare(o.Max.Value) > 0, Style: func(root, label, leading, trailing ui.Element) { leading.Rotate(-90) }}).OnClick(func() { s.Month = next; s.Active = next })
		})
		ui.Row(c.Key("weekdays")).Children(func() {
			weekdays := i18n.Weekdays(c.Locale(), i18n.Short, true)
			for i := range 7 {
				day := time.Weekday((i + int(o.WeekStart)) % 7)
				ui.Text(c, weekdays[day]).Width(t.text(38)).TextAlign(ui.Center).TextColor(t.Muted).FontSize(t.text(12))
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

type DateInputState struct {
	Text        string
	Error       string
	Calendar    CalendarState
	Open        bool
	seen        Optional[Date]
	initialized bool
}
type InputDateOptions struct {
	HideCalendar bool
	InputOptions
	Calendar      CalendarOptions
	OnValueChange func(Optional[Date])
}

func InputDate(c *ui.Context, t Theme, value *Optional[Date], s *DateInputState, o InputDateOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputDate requires value and state")
	}
	validateCalendar(o.Calendar)
	if !s.initialized || s.seen != *value {
		s.Text = ""
		if value.Set {
			s.Text = value.Value.String()
			s.Calendar.Month = Date{value.Value.Year, value.Value.Month, 1}
			s.Calendar.Active = value.Value
		}
		s.seen = *value
		s.initialized = true
		s.Error = ""
	}
	update := func(v Optional[Date]) {
		s.Error = ""
		if v.Set {
			s.Calendar.Month = Date{v.Value.Year, v.Value.Month, 1}
			s.Calendar.Active = v.Value
		}
		if v != *value {
			*value = v
			s.seen = v
			if o.OnValueChange != nil {
				o.OnValueChange(v)
			}
		}
	}
	in := o.InputOptions
	if in.Placeholder == "" {
		in.Placeholder = "YYYY-MM-DD"
	}
	invalidMessage := message(c, "invalidDate")
	if s.Error != "" {
		s.Error = invalidMessage
		in.ErrorMessage = s.Error
		in.errorKey = "invalidDate"
	}
	prior := in.OnChange
	in.OnChange = func(text string) {
		v, ok := ParseDate(text)
		if ok && (!v.Set || dateEnabled(v.Value, o.Calendar)) {
			update(v)
		} else {
			s.Error = invalidMessage
		}
		if prior != nil {
			prior(text)
		}
	}
	in.OnSubmit = func(text string) {
		if s.Error == "" && o.OnSubmit != nil {
			o.OnSubmit(text)
		}
	}
	style := in.Style
	in.Style = func(root, editor, leading, trailing ui.Element) {
		editor.InputMethod(false)
		if style != nil {
			style(root, editor, leading, trailing)
		}
	}
	root := ui.Column(c).MinWidth(0)
	root.Children(func() {
		anchor := inputShell(c.Key("date"), t, &s.Text, in, ComponentInputDate, 0, 0, func() ui.Element { return ui.TextInputBase(c.Key("editor"), &s.Text) }, func(root ui.Element) {
			if o.HideCalendar {
				return
			}
			Button(c.Key("calendar-button"), t, "", ButtonOptions{AccessibleLabel: namedMessage(c, "chooseNamed", o.Label), Trailing: lucide.ChevronDown(), Variant: Ghost, Size: XS, Disabled: o.ReadOnly}).OnClick(func() { s.Open = !s.Open })
		})
		if o.HideCalendar || o.ReadOnly || o.Disabled || t.field.Disabled {
			s.Open = false
		}
		Popover(c.Key("calendar"), t, anchor, &s.Open, PopoverOptions{MaxHeight: Set(float32(400))}, func(inner Theme) {
			selection := DateSelection{Start: *value}
			opts := o.Calendar
			opts.Range = false
			opts.Label = namedMessage(c, "calendarNamed", o.Label)
			opts.OnChange = func(v DateSelection) {
				update(v.Start)
				s.Text = value.Value.String()
				s.Calendar.Month = Date{value.Value.Year, value.Value.Month, 1}
				s.Calendar.Active = value.Value
				s.Open = false
			}
			Calendar(c, inner, &selection, &s.Calendar, opts)
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
func clockFormat(v ClockTime, twelve, seconds bool) string {
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
	return time.Date(2000, 1, 1, v.Hour, v.Minute, v.Second, 0, time.UTC).Format(layout)
}
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

type TimeInputState struct {
	Text, Error                  string
	seen                         Optional[ClockTime]
	initialized, twelve, seconds bool
}
type InputTimeOptions struct {
	InputOptions
	TwelveHour, Seconds bool
	OnValueChange       func(Optional[ClockTime])
}

func InputTime(c *ui.Context, t Theme, value *Optional[ClockTime], s *TimeInputState, o InputTimeOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputTime requires value and state")
	}
	if !s.initialized || s.seen != *value || s.twelve != o.TwelveHour || s.seconds != o.Seconds {
		s.Text = ""
		if value.Set {
			s.Text = clockFormat(value.Value, o.TwelveHour, o.Seconds)
		}
		s.seen = *value
		s.initialized = true
		s.twelve = o.TwelveHour
		s.seconds = o.Seconds
		s.Error = ""
	}
	in := o.InputOptions
	if in.Placeholder == "" {
		in.Placeholder = "HH:MM"
		if o.Seconds {
			in.Placeholder += ":SS"
		}
		if o.TwelveHour {
			in.Placeholder += " AM/PM"
		}
	}
	invalidMessage := message(c, "invalidTime")
	if s.Error != "" {
		s.Error = invalidMessage
		in.ErrorMessage = s.Error
		in.errorKey = "invalidTime"
	}
	prior := in.OnChange
	in.OnChange = func(text string) {
		v, ok := ParseClockTime(text, o.TwelveHour, o.Seconds)
		if ok {
			s.Error = ""
			if v.Set && !o.Seconds && value.Set {
				v.Value.Second = value.Value.Second
			}
			if v != *value {
				*value = v
				s.seen = v
				if o.OnValueChange != nil {
					o.OnValueChange(v)
				}
			}
		} else {
			s.Error = invalidMessage
		}
		if prior != nil {
			prior(text)
		}
	}
	submit := in.OnSubmit
	in.OnSubmit = func(text string) {
		if s.Error == "" && submit != nil {
			submit(text)
		}
	}
	style := in.Style
	in.Style = func(root, editor, leading, trailing ui.Element) {
		editor.InputMethod(false)
		if style != nil {
			style(root, editor, leading, trailing)
		}
	}
	return inputShell(c, t, &s.Text, in, ComponentInputTime, 0, 0, func() ui.Element { return ui.TextInputBase(c.Key("time"), &s.Text) }, nil)
}

// DateRangeInputState owns two editors and a shared range calendar.
type DateRangeInputState struct {
	Start, End DateInputState
	Calendar   CalendarState
	Open       bool
}

// InputDateRange uses the InputDate recipe. Start changes beyond End clear End;
// a reverse calendar click is normalized by SelectDate. Register each endpoint separately
// with FormField when per-endpoint validation and first-error focus are required.
func InputDateRange(c *ui.Context, t Theme, value *DateSelection, s *DateRangeInputState, o InputDateOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputDateRange requires value and state")
	}
	inner := t
	inner.field.State = nil
	if !s.Open && value.Start.Set && value.Start.Value.Valid() {
		s.Calendar.Month = Date{value.Start.Value.Year, value.Start.Value.Month, 1}
		s.Calendar.Active = value.Start.Value
	}
	root := ui.Row(c).MinWidth(0).Gap(t.spacing(8)).Disabled(o.Disabled || t.field.Disabled)
	root.Children(func() {
		start := o
		start.HideCalendar = true
		start.Label = namedMessage(c, "rangeStart", o.Label)
		start.OnValueChange = func(v Optional[Date]) {
			if value.End.Set && v.Set && v.Value.Compare(value.End.Value) > 0 {
				value.End = Optional[Date]{}
			}
			if o.Calendar.OnChange != nil {
				o.Calendar.OnChange(*value)
			}
		}
		InputDate(c.Key("start"), inner, &value.Start, &s.Start, start).Grow(1)
		ui.Text(c, "–").TextColor(t.Muted)
		end := o
		end.HideCalendar = true
		end.Label = namedMessage(c, "rangeEnd", o.Label)
		if value.Start.Set && (!end.Calendar.Min.Set || value.Start.Value.Compare(end.Calendar.Min.Value) > 0) {
			end.Calendar.Min = value.Start
		}
		end.OnValueChange = func(Optional[Date]) {
			if o.Calendar.OnChange != nil {
				o.Calendar.OnChange(*value)
			}
		}
		// A shared field binds only its first editor; independent fields can be used for two focus targets.
		InputDate(c.Key("end"), inner, &value.End, &s.End, end).Grow(1)
		if o.HideCalendar {
			s.Open = false
			return
		}
		anchor := Button(c.Key("range-trigger"), t, "", ButtonOptions{AccessibleLabel: namedMessage(c, "chooseNamed", o.Label), Leading: lucide.ChevronDown(), Variant: Outline, Disabled: o.ReadOnly}).OnClick(func() { s.Open = !s.Open })
		if o.Disabled || o.ReadOnly || t.field.Disabled {
			s.Open = false
		}
		Popover(c.Key("range-calendar"), t, anchor, &s.Open, PopoverOptions{MaxHeight: Set(float32(400))}, func(inner Theme) {
			opts := o.Calendar
			opts.Range = true
			opts.Label = namedMessage(c, "calendarNamed", o.Label)
			opts.OnChange = func(v DateSelection) {
				if v.End.Set {
					s.Open = false
				}
				if o.Calendar.OnChange != nil {
					o.Calendar.OnChange(v)
				}
			}
			Calendar(c, inner, value, &s.Calendar, opts)
		})
	})
	message := s.Start.Error
	if message == "" {
		message = s.End.Error
	}
	key := ""
	if message != "" {
		key = "invalidDate"
	}
	trackFieldEdit(t, s.Start.Text+"\x00"+s.End.Text, message, key)
	return root
}
