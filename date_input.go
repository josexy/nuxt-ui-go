package nui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type DateInputState struct {
	// Text is the canonical ISO draft (or rejected pasted text), independent
	// of the locale's visual segment order. Incomplete drafts retain the value.
	Text, Error string
	Calendar    CalendarState
	Open        bool
	seen        Optional[Date]
	initialized bool
	segments    segmentState
}

type InputDateOptions struct {
	InputOptions
	// ShowCalendar adds a calendar button and popup. The default is a date field.
	ShowCalendar bool
	// HideCalendar overrides ShowCalendar; retained for existing callers.
	HideCalendar  bool
	Calendar      CalendarOptions
	OnValueChange func(Optional[Date])
}

func (s *DateInputState) set(v Optional[Date]) {
	s.segments.values = [partCount]Optional[int]{}
	s.segments.typed = [partCount]string{}
	if v.Set && v.Value.Valid() {
		s.segments.values[partYear] = Set(v.Value.Year)
		s.segments.values[partMonth] = Set(int(v.Value.Month))
		s.segments.values[partDay] = Set(v.Value.Day)
	}
	s.Text, s.Error = dateDraft(&s.segments), ""
	s.seen, s.initialized = v, true
}

func dateDraft(s *segmentState) string {
	y, m, d := s.values[partYear], s.values[partMonth], s.values[partDay]
	if !y.Set && !m.Set && !d.Set {
		return ""
	}
	return draftNumber(y, 4) + "-" + draftNumber(m, 2) + "-" + draftNumber(d, 2)
}

func draftNumber(v Optional[int], width int) string {
	if !v.Set {
		return strings.Repeat("-", width)
	}
	return fmt.Sprintf("%0*d", width, v.Value)
}

func dateField(c *ui.Context, value *Optional[Date], s *DateInputState, o InputDateOptions) segmentField {
	validateCalendar(o.Calendar)
	if !s.initialized || s.seen != *value {
		s.set(*value)
	}
	invalid := message(c, "invalidDate")
	if s.Error != "" {
		s.Error = invalid
	}
	f := segmentField{state: &s.segments, parts: segmentPattern(c.Region().FormatLocale(c.Locale()), false, false, false), label: o.Label, text: &s.Text, err: &s.Error}
	commit := func(v Optional[Date]) {
		s.Error = ""
		if v.Set {
			s.Calendar.Active = v.Value
			s.Calendar.Month = Date{v.Value.Year, v.Value.Month, 1}
		}
		if v != *value {
			*value, s.seen = v, v
			if o.OnValueChange != nil {
				o.OnValueChange(v)
			}
		}
	}
	notify := func(before string) {
		if before != s.Text && o.OnChange != nil {
			o.OnChange(s.Text)
		}
	}
	f.change = func(p datePart) {
		before := s.Text
		// A year is typed as four digits. Do not clamp February while the
		// user is still replacing 2024 with 2028 one digit at a time.
		pendingYear := len(s.segments.typed[partYear]) > 0 && len(s.segments.typed[partYear]) < 4
		if (p == partYear || p == partMonth) && !pendingYear {
			_, last := f.limits(partDay)
			if d := &s.segments.values[partDay]; d.Set && d.Value > last {
				d.Value = last
			}
		}
		s.Text = dateDraft(&s.segments)
		v, ok := ParseDate(s.Text)
		if ok && !pendingYear && (!v.Set || dateEnabled(v.Value, o.Calendar)) {
			commit(v)
		} else {
			s.Error = invalid
		}
		notify(before)
	}
	f.paste = func(text string) {
		before := s.Text
		v, ok := parseDatePaste(text, f.parts)
		if ok && (!v.Set || dateEnabled(v.Value, o.Calendar)) {
			commit(v)
			s.set(v)
		} else {
			s.Text, s.Error = text, invalid
		}
		notify(before)
	}
	return f
}

func parseDatePaste(text string, parts []dateSegment) (Optional[Date], bool) {
	if v, ok := ParseDate(text); ok {
		return v, true
	}
	// Accept the visible locale order as well as ISO. Require three fields;
	// arbitrary letters or extra digits must not silently become a date.
	numbers := strings.FieldsFunc(strings.TrimSpace(text), func(r rune) bool {
		return r == '/' || r == '.' || r == '-' || r == '年' || r == '月' || r == '日' || unicode.IsSpace(r)
	})
	if len(numbers) != 3 {
		return Optional[Date]{}, false
	}
	var values [partCount]int
	i := 0
	for _, part := range parts {
		if part.part == partLiteral {
			continue
		}
		if i >= len(numbers) {
			return Optional[Date]{}, false
		}
		n, err := strconv.Atoi(numbers[i])
		if err != nil {
			return Optional[Date]{}, false
		}
		values[part.part] = n
		i++
	}
	d := Date{values[partYear], time.Month(values[partMonth]), values[partDay]}
	return Set(d), i == 3 && d.Valid()
}

func segmentOptions(o InputOptions, fields ...segmentField) InputOptions {
	for _, f := range fields {
		if *f.err != "" && o.ErrorMessage == "" {
			o.ErrorMessage = *f.err
			o.errorKey = "invalidDate"
			if f.clock {
				o.errorKey = "invalidTime"
			}
		}
	}
	return o
}

func InputDate(c *ui.Context, t Theme, value *Optional[Date], s *DateInputState, o InputDateOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputDate requires value and state")
	}
	if o.Label == "" {
		o.Label = t.field.Label
	}
	f := dateField(c, value, s, o)
	show := o.ShowCalendar && !o.HideCalendar
	if !show || o.ReadOnly || o.Disabled || t.field.Disabled {
		s.Open = false
	}
	root := ui.Column(c).MinWidth(0)
	root.Children(func() {
		anchor := segmentedInput(c.Key("date"), t, []segmentField{f}, segmentOptions(o.InputOptions, f), ComponentInputDate, func() {
			if show {
				Button(c.Key("calendar-button"), t, "", ButtonOptions{AccessibleLabel: namedMessage(c, "chooseNamed", o.Label), Leading: lucide.Calendar(), Variant: Ghost, Size: XS, Disabled: o.ReadOnly}).OnClick(func() { s.Open = !s.Open })
			}
		})
		Popover(c.Key("calendar"), t, anchor, &s.Open, PopoverOptions{MaxHeight: Set(float32(400))}, func(inner Theme) {
			selection := DateSelection{Start: *value}
			opts := o.Calendar
			opts.Range = false
			opts.Label = namedMessage(c, "calendarNamed", o.Label)
			opts.OnChange = func(v DateSelection) { f.paste(v.Start.Value.String()); s.Open = false }
			Calendar(c, inner, &selection, &s.Calendar, opts)
		})
	})
	return root
}

type TimeInputState struct {
	Text, Error                  string
	seen                         Optional[ClockTime]
	initialized, twelve, seconds bool
	segments                     segmentState
}

type InputTimeOptions struct {
	InputOptions
	// TwelveHour defaults to the locale and the user's regional hour cycle.
	TwelveHour    Optional[bool]
	Seconds       bool
	OnValueChange func(Optional[ClockTime])
}

func (s *TimeInputState) set(v Optional[ClockTime], twelve, seconds bool) {
	s.segments.values = [partCount]Optional[int]{}
	s.segments.typed = [partCount]string{}
	if v.Set && v.Value.Valid() {
		h := v.Value.Hour
		if twelve {
			h %= 12
			if h == 0 {
				h = 12
			}
		}
		s.segments.values[partHour] = Set(h)
		s.segments.values[partMinute] = Set(v.Value.Minute)
		s.segments.values[partSecond] = Set(v.Value.Second)
		s.segments.values[partPeriod] = Set(v.Value.Hour / 12)
	}
	s.twelve, s.seconds = twelve, seconds
	s.Text, s.Error = timeDraft(&s.segments, twelve, seconds), ""
	s.seen, s.initialized = v, true
}

func timeDraft(s *segmentState, twelve, seconds bool) string {
	h, m, sec := s.values[partHour], s.values[partMinute], s.values[partSecond]
	if !h.Set && !m.Set && (!seconds || !sec.Set) {
		return ""
	}
	text := draftNumber(h, 2) + ":" + draftNumber(m, 2)
	if seconds {
		text += ":" + draftNumber(sec, 2)
	}
	if twelve {
		period := s.values[partPeriod]
		if !period.Set {
			text += " --"
		} else if period.Value == 0 {
			text += " AM"
		} else {
			text += " PM"
		}
	}
	return text
}

func timeField(c *ui.Context, value *Optional[ClockTime], s *TimeInputState, o InputTimeOptions) segmentField {
	twelve := o.TwelveHour.Value
	if !o.TwelveHour.Set {
		twelve = localeTwelveHour(c)
	}
	if !s.initialized || s.seen != *value {
		s.set(*value, twelve, o.Seconds)
	}
	if s.twelve != twelve {
		if !twelve && s.segments.handles[partPeriod].Focused(c) {
			s.segments.handles[partHour].Focus(c)
		}
		h := &s.segments.values[partHour]
		if h.Set {
			if twelve {
				s.segments.values[partPeriod] = Set(h.Value / 12)
				h.Value %= 12
				if h.Value == 0 {
					h.Value = 12
				}
			} else {
				h.Value %= 12
				h.Value += 12 * s.segments.values[partPeriod].Value
			}
		}
		s.twelve = twelve
		if s.Error == "" {
			s.Text = timeDraft(&s.segments, twelve, o.Seconds)
		}
	}
	if s.seconds != o.Seconds {
		if !o.Seconds && s.segments.handles[partSecond].Focused(c) {
			s.segments.handles[partMinute].Focus(c)
		}
		s.seconds = o.Seconds
		if o.Seconds && !s.segments.values[partSecond].Set && value.Set {
			s.segments.values[partSecond] = Set(value.Value.Second)
		}
		if s.Error == "" {
			s.Text = timeDraft(&s.segments, twelve, o.Seconds)
		}
	}
	invalid := message(c, "invalidTime")
	if s.Error != "" {
		s.Error = invalid
	}
	f := segmentField{state: &s.segments, parts: segmentPattern(c.Region().FormatLocale(c.Locale()), true, twelve, o.Seconds), label: o.Label, text: &s.Text, err: &s.Error, twelve: twelve, clock: true}
	commit := func(v Optional[ClockTime]) {
		s.Error = ""
		if v.Set && !o.Seconds && value.Set {
			v.Value.Second = value.Value.Second
		}
		if v.Set {
			s.segments.values[partSecond] = Set(v.Value.Second)
		}
		if v != *value {
			*value, s.seen = v, v
			if o.OnValueChange != nil {
				o.OnValueChange(v)
			}
		}
	}
	notify := func(before string) {
		if before != s.Text && o.OnChange != nil {
			o.OnChange(s.Text)
		}
	}
	f.change = func(p datePart) {
		before := s.Text
		if twelve && p != partPeriod && !s.segments.values[partPeriod].Set {
			s.segments.values[partPeriod] = Set(0)
		}
		s.Text = timeDraft(&s.segments, twelve, o.Seconds)
		v, ok := ParseClockTime(s.Text, twelve, o.Seconds)
		if ok {
			commit(v)
		} else {
			s.Error = invalid
		}
		notify(before)
	}
	f.paste = func(text string) {
		before := s.Text
		v, ok := ParseClockTime(text, twelve, o.Seconds)
		if ok {
			commit(v)
			s.set(*value, twelve, o.Seconds)
		} else {
			s.Text, s.Error = text, invalid
		}
		notify(before)
	}
	return f
}

func InputTime(c *ui.Context, t Theme, value *Optional[ClockTime], s *TimeInputState, o InputTimeOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputTime requires value and state")
	}
	if o.Label == "" {
		o.Label = t.field.Label
	}
	f := timeField(c, value, s, o)
	return segmentedInput(c, t, []segmentField{f}, segmentOptions(o.InputOptions, f), ComponentInputTime, nil)
}

// DateRangeInputState owns two segmented endpoints and an optional shared calendar.
type DateRangeInputState struct {
	Start, End DateInputState
	Calendar   CalendarState
	Open       bool
}

func InputDateRange(c *ui.Context, t Theme, value *DateSelection, s *DateRangeInputState, o InputDateOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputDateRange requires value and state")
	}
	if o.Label == "" {
		o.Label = t.field.Label
	}
	start := o
	start.Label = namedMessage(c, "rangeStart", o.Label)
	start.OnValueChange = func(v Optional[Date]) {
		if value.End.Set && v.Set && v.Value.Compare(value.End.Value) > 0 {
			value.End = Optional[Date]{}
		}
		if o.Calendar.OnChange != nil {
			o.Calendar.OnChange(*value)
		}
	}
	end := o
	end.Label = namedMessage(c, "rangeEnd", o.Label)
	if value.Start.Set && (!end.Calendar.Min.Set || value.Start.Value.Compare(end.Calendar.Min.Value) > 0) {
		end.Calendar.Min = value.Start
	}
	end.OnValueChange = func(Optional[Date]) {
		if o.Calendar.OnChange != nil {
			o.Calendar.OnChange(*value)
		}
	}
	fields := []segmentField{dateField(c, &value.Start, &s.Start, start), dateField(c, &value.End, &s.End, end)}
	show := o.ShowCalendar && !o.HideCalendar
	if !show || o.ReadOnly || o.Disabled || t.field.Disabled {
		s.Open = false
	}
	if !s.Open && value.Start.Set && value.Start.Value.Valid() {
		s.Calendar.Month = Date{value.Start.Value.Year, value.Start.Value.Month, 1}
		s.Calendar.Active = value.Start.Value
	}
	root := ui.Column(c).MinWidth(0)
	root.Children(func() {
		anchor := segmentedInput(c.Key("range"), t, fields, segmentOptions(o.InputOptions, fields...), ComponentInputDate, func() {
			if show {
				Button(c.Key("calendar-button"), t, "", ButtonOptions{AccessibleLabel: namedMessage(c, "chooseNamed", o.Label), Leading: lucide.Calendar(), Variant: Ghost, Size: XS, Disabled: o.ReadOnly}).OnClick(func() { s.Open = !s.Open })
			}
		})
		Popover(c.Key("calendar"), t, anchor, &s.Open, PopoverOptions{MaxHeight: Set(float32(400))}, func(inner Theme) {
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
	return root
}
