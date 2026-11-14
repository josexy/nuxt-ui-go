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
	rejected    bool
	segments    segmentState
}

type InputDateOptions struct {
	InputOptions
	Step             DateStep
	PlaceholderValue Optional[Date]
	SeparatorIcon    *ui.SVG
	BuildSeparator   func(*ui.Context, Theme)
	// ShowCalendar adds a calendar button and popup. The default is a date field.
	ShowCalendar bool
	// HideCalendar overrides ShowCalendar; retained for existing callers.
	HideCalendar  bool
	Calendar      CalendarOptions
	OnValueChange func(Optional[Date])
	OnRangeChange func(DateSelection)
}

// DateStep configures segment cycling. Zero uses one; negative steps are invalid.
type DateStep struct{ Year, Month, Day int }

func (s *DateInputState) set(v Optional[Date]) {
	s.segments.values = [partCount]Optional[int]{}
	s.segments.typed = [partCount]string{}
	if v.Set && v.Value.Valid() {
		s.segments.values[partYear] = Set(v.Value.Year)
		s.segments.values[partMonth] = Set(int(v.Value.Month))
		s.segments.values[partDay] = Set(v.Value.Day)
	}
	s.Text, s.Error, s.rejected = dateDraft(&s.segments), "", false
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

func dateField(c *ui.Context, value *Optional[Date], s *DateInputState, o InputDateOptions, accept func(Date) bool) segmentField {
	validateCalendar(o.Calendar)
	if o.Step.Year < 0 || o.Step.Month < 0 || o.Step.Day < 0 || o.PlaceholderValue.Set && !o.PlaceholderValue.Value.Valid() {
		panic("nui: invalid date field step or placeholder")
	}
	enabled := func(d Date) bool { return dateEnabled(d, o.Calendar) && (accept == nil || accept(d)) }
	if !s.initialized || s.seen != *value {
		s.set(*value)
	}
	invalid := message(c, "invalidDate")
	s.Error = ""
	if s.rejected || value.Set && !enabled(value.Value) {
		s.Error = invalid
	}
	f := segmentField{state: &s.segments, parts: segmentPattern(c.Region().FormatLocale(c.Locale()), false, false, false, false), label: o.Label, text: &s.Text, err: &s.Error}
	f.steps[partYear], f.steps[partMonth], f.steps[partDay] = o.Step.Year, o.Step.Month, o.Step.Day
	seed := DateFromTime(time.Now())
	if o.PlaceholderValue.Set {
		seed = o.PlaceholderValue.Value
	}
	f.seed[partYear], f.seed[partMonth], f.seed[partDay] = Set(seed.Year), Set(int(seed.Month)), Set(seed.Day)
	commit := func(v Optional[Date]) {
		s.Error, s.rejected = "", false
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
		if ok && !pendingYear && (!v.Set || enabled(v.Value)) {
			commit(v)
		} else {
			s.Error, s.rejected = invalid, true
		}
		notify(before)
	}
	f.paste = func(text string) {
		before := s.Text
		v, ok := parseDatePaste(text, f.parts)
		if ok && (!v.Set || enabled(v.Value)) {
			commit(v)
			s.set(v)
		} else {
			s.Text, s.Error, s.rejected = text, invalid, true
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
	f := dateField(c, value, s, o, nil)
	if !s.Calendar.Month.Valid() && o.PlaceholderValue.Set {
		s.Calendar.Month, s.Calendar.Active = calendarMonth(o.PlaceholderValue.Value), o.PlaceholderValue.Value
	}
	show := o.ShowCalendar && !o.HideCalendar
	if !show || o.ReadOnly || o.Disabled || t.field.Disabled {
		s.Open = false
	}
	root := ui.Column(c).MinWidth(0).Disabled(o.Disabled || t.field.Disabled)
	if root.IsDisabled() {
		s.Open = false
	}
	root.Children(func() {
		anchor := segmentedInput(c.Key("date"), t, []segmentField{f}, segmentOptions(o.InputOptions, f), ComponentInputDate, func() {
			if show {
				Button(c.Key("calendar-button"), t, "", ButtonOptions{AccessibleLabel: namedMessage(c, "chooseNamed", o.Label), Leading: lucide.Calendar(), Variant: Ghost, Size: XS, Disabled: o.ReadOnly}).OnClick(func() { s.Open = !s.Open })
			}
		}, segmentExtras{})
		Popover(c.Key("calendar"), t, anchor, &s.Open, PopoverOptions{MaxHeight: Set(float32(400))}, func(inner Theme) {
			selection := DateSelection{Start: *value}
			opts := o.Calendar
			opts.Range = false
			if opts.Label == "" {
				opts.Label = namedMessage(c, "calendarNamed", o.Label)
			}
			opts.OnChange = func(v DateSelection) {
				f.paste(v.Start.Value.String())
				s.Open = false
				if o.Calendar.OnChange != nil {
					o.Calendar.OnChange(v)
				}
			}
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
	granularity                  TimeGranularity
	rejected                     bool
}

type InputTimeOptions struct {
	InputOptions
	// TwelveHour defaults to the locale and the user's regional hour cycle.
	TwelveHour Optional[bool]
	Seconds    bool
	// Granularity overrides Seconds when set. Hidden units retain their value.
	Granularity       TimeGranularity
	Min, Max          Optional[ClockTime]
	PlaceholderValue  Optional[ClockTime]
	Step              TimeStep
	StepSnapping      bool
	IsTimeUnavailable func(ClockTime) bool
	OnValueChange     func(Optional[ClockTime])
}

type TimeGranularity uint8

const (
	TimeGranularityDefault TimeGranularity = iota
	TimeHour
	TimeMinute
	TimeSecond
)

// TimeStep configures segment cycling. Zero uses one; negative steps are invalid.
type TimeStep struct{ Hour, Minute, Second int }

func (s *TimeInputState) set(v Optional[ClockTime], twelve bool, precision TimeGranularity) {
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
	s.twelve, s.seconds, s.granularity = twelve, precision == TimeSecond, precision
	s.Text, s.Error, s.rejected = timeDraft(&s.segments, twelve, precision), "", false
	s.seen, s.initialized = v, true
}

func timeDraft(s *segmentState, twelve bool, precision TimeGranularity) string {
	h, m, sec := s.values[partHour], s.values[partMinute], s.values[partSecond]
	if !h.Set && (precision < TimeMinute || !m.Set) && (precision < TimeSecond || !sec.Set) {
		return ""
	}
	text := draftNumber(h, 2)
	if precision >= TimeMinute {
		text += ":" + draftNumber(m, 2)
	}
	if precision == TimeSecond {
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

func timeValueOrder(v ClockTime) int { return v.Hour*3600 + v.Minute*60 + v.Second }

func timeField(c *ui.Context, value *Optional[ClockTime], s *TimeInputState, o InputTimeOptions) segmentField {
	if o.Min.Set && !o.Min.Value.Valid() || o.Max.Set && !o.Max.Value.Valid() || o.Min.Set && o.Max.Set && timeValueOrder(o.Min.Value) > timeValueOrder(o.Max.Value) || o.PlaceholderValue.Set && !o.PlaceholderValue.Value.Valid() {
		panic("nui: invalid time field bounds or placeholder")
	}
	if o.Step.Hour < 0 || o.Step.Minute < 0 || o.Step.Second < 0 || o.Granularity > TimeSecond {
		panic("nui: invalid time field step or granularity")
	}
	precision := o.Granularity
	if precision == TimeGranularityDefault {
		precision = TimeMinute
		if o.Seconds {
			precision = TimeSecond
		}
	}
	twelve := o.TwelveHour.Value
	if !o.TwelveHour.Set {
		twelve = localeTwelveHour(c)
	}
	external := !s.initialized || s.seen != *value
	if external {
		s.set(*value, twelve, precision)
	}
	formatChanged := s.twelve != twelve || s.granularity != precision
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
				h.Value = h.Value%12 + 12*s.segments.values[partPeriod].Value
			}
		}
		s.twelve = twelve
	}
	if s.granularity != precision {
		if precision == TimeHour && (s.segments.handles[partMinute].Focused(c) || s.segments.handles[partSecond].Focused(c)) {
			s.segments.handles[partHour].Focus(c)
		} else if precision == TimeMinute && s.segments.handles[partSecond].Focused(c) {
			s.segments.handles[partMinute].Focus(c)
		}
		if value.Set {
			if precision >= TimeMinute && !s.segments.values[partMinute].Set {
				s.segments.values[partMinute] = Set(value.Value.Minute)
			}
			if precision == TimeSecond && !s.segments.values[partSecond].Set {
				s.segments.values[partSecond] = Set(value.Value.Second)
			}
		}
		s.seconds, s.granularity = precision == TimeSecond, precision
	}
	if formatChanged && !s.rejected {
		s.Text = timeDraft(&s.segments, twelve, precision)
	}
	invalid := message(c, "invalidTime")
	enabled := func(v ClockTime) bool {
		return v.Valid() && (!o.Min.Set || timeValueOrder(v) >= timeValueOrder(o.Min.Value)) && (!o.Max.Set || timeValueOrder(v) <= timeValueOrder(o.Max.Value)) && (o.IsTimeUnavailable == nil || !o.IsTimeUnavailable(v))
	}
	s.Error = ""
	if s.rejected || value.Set && !enabled(value.Value) {
		s.Error = invalid
	}
	f := segmentField{state: &s.segments, parts: segmentPattern(c.Region().FormatLocale(c.Locale()), true, twelve, precision == TimeSecond, precision == TimeHour), label: o.Label, text: &s.Text, err: &s.Error, twelve: twelve, clock: true}
	f.steps[partHour], f.steps[partMinute], f.steps[partSecond] = o.Step.Hour, o.Step.Minute, o.Step.Second
	if o.PlaceholderValue.Set {
		seed := o.PlaceholderValue.Value
		h := seed.Hour
		if twelve {
			h %= 12
			if h == 0 {
				h = 12
			}
		}
		f.seed[partHour], f.seed[partMinute], f.seed[partSecond], f.seed[partPeriod] = Set(h), Set(seed.Minute), Set(seed.Second), Set(seed.Hour/12)
	}
	parse := func(text string) (Optional[ClockTime], bool) {
		if precision == TimeHour && strings.TrimSpace(text) != "" {
			chunks := strings.Fields(text)
			chunks[0] += ":00"
			text = strings.Join(chunks, " ")
		}
		v, ok := ParseClockTime(text, twelve, precision == TimeSecond)
		if v.Set && ok {
			retained := ClockTime{}
			if o.PlaceholderValue.Set {
				retained = o.PlaceholderValue.Value
			}
			if value.Set {
				retained = value.Value
			}
			if precision == TimeHour {
				v.Value.Minute = retained.Minute
			}
			if precision < TimeSecond {
				v.Value.Second = retained.Second
			}
		}
		return v, ok
	}
	commit := func(v Optional[ClockTime]) {
		s.Error, s.rejected = "", false
		if v.Set {
			s.segments.values[partSecond] = Set(v.Value.Second)
			if precision == TimeHour {
				s.segments.values[partMinute] = Set(v.Value.Minute)
			}
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
			period := 0
			if o.PlaceholderValue.Set {
				period = o.PlaceholderValue.Value.Hour / 12
			}
			s.segments.values[partPeriod] = Set(period)
		}
		s.Text = timeDraft(&s.segments, twelve, precision)
		v, ok := parse(s.Text)
		if ok && (!v.Set || enabled(v.Value)) {
			commit(v)
		} else {
			s.Error, s.rejected = invalid, true
		}
		notify(before)
	}
	f.paste = func(text string) {
		before := s.Text
		v, ok := parse(text)
		if ok && (!v.Set || enabled(v.Value)) {
			commit(v)
			s.set(*value, twelve, precision)
		} else {
			s.Text, s.Error, s.rejected = text, invalid, true
		}
		notify(before)
	}
	if o.StepSnapping && !external && !formatChanged {
		f.blur = func(p datePart) {
			if s.rejected {
				return
			}
			if p < partHour || p > partSecond || !s.segments.values[p].Set || p == partMinute && precision < TimeMinute || p == partSecond && precision < TimeSecond {
				return
			}
			step := max(1, f.steps[p])
			if step == 1 {
				return
			}
			n := s.segments.values[p].Value
			hi := 59
			if p == partHour {
				hi = 23
				if twelve {
					if !s.segments.values[partPeriod].Set {
						return
					}
					n = n%12 + 12*s.segments.values[partPeriod].Value
				}
			}
			rounded := min((n+step/2)/step, hi/step) * step
			if rounded == n {
				return
			}
			if p == partHour && twelve {
				s.segments.values[partPeriod] = Set(rounded / 12)
				rounded %= 12
				if rounded == 0 {
					rounded = 12
				}
			}
			s.segments.values[p], s.segments.typed[p] = Set(rounded), ""
			f.change(p)
		}
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
	return segmentedInput(c, t, []segmentField{f}, segmentOptions(o.InputOptions, f), ComponentInputTime, nil, segmentExtras{})
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
	batch := false
	publish := func() {
		if batch {
			return
		}
		next := *value
		if o.OnRangeChange != nil {
			o.OnRangeChange(next)
		}
		if o.Calendar.OnChange != nil {
			o.Calendar.OnChange(next)
		}
	}
	start := o
	start.Label = namedMessage(c, "rangeStart", o.Label)
	start.OnValueChange = func(v Optional[Date]) {
		if value.End.Set && v.Set && v.Value.Compare(value.End.Value) > 0 {
			value.End = Optional[Date]{}
		}
		publish()
	}
	end := o
	end.Label = namedMessage(c, "rangeEnd", o.Label)
	end.OnValueChange = func(Optional[Date]) { publish() }
	fields := []segmentField{
		dateField(c, &value.Start, &s.Start, start, func(d Date) bool {
			return !value.End.Set || d.Compare(value.End.Value) > 0 || dateInputRangeAllowed(d, value.End.Value, o.Calendar)
		}),
		dateField(c, &value.End, &s.End, end, func(d Date) bool {
			return !value.Start.Set || dateInputRangeAllowed(value.Start.Value, d, o.Calendar)
		}),
	}
	if !s.Calendar.Month.Valid() && o.PlaceholderValue.Set {
		s.Calendar.Month, s.Calendar.Active = calendarMonth(o.PlaceholderValue.Value), o.PlaceholderValue.Value
	}
	extras := segmentExtras{separator: o.SeparatorIcon, buildSeparator: o.BuildSeparator, clear: func() {
		before := *value
		batch = true
		for _, f := range fields {
			f.paste("")
		}
		batch = false
		if before != *value {
			publish()
		}
	}}
	show := o.ShowCalendar && !o.HideCalendar
	if !show || o.ReadOnly || o.Disabled || t.field.Disabled {
		s.Open = false
	}
	if !s.Open && value.Start.Set && value.Start.Value.Valid() {
		s.Calendar.Month = Date{value.Start.Value.Year, value.Start.Value.Month, 1}
		s.Calendar.Active = value.Start.Value
	}
	root := ui.Column(c).MinWidth(0).Disabled(o.Disabled || t.field.Disabled)
	if root.IsDisabled() {
		s.Open = false
	}
	root.Children(func() {
		anchor := segmentedInput(c.Key("range"), t, fields, segmentOptions(o.InputOptions, fields...), ComponentInputDate, func() {
			if show {
				Button(c.Key("calendar-button"), t, "", ButtonOptions{AccessibleLabel: namedMessage(c, "chooseNamed", o.Label), Leading: lucide.Calendar(), Variant: Ghost, Size: XS, Disabled: o.ReadOnly}).OnClick(func() { s.Open = !s.Open })
			}
		}, extras)
		Popover(c.Key("calendar"), t, anchor, &s.Open, PopoverOptions{MaxHeight: Set(float32(400))}, func(inner Theme) {
			opts := o.Calendar
			opts.Range = true
			if opts.Label == "" {
				opts.Label = namedMessage(c, "calendarNamed", o.Label)
			}
			opts.OnChange = func(v DateSelection) {
				if v.End.Set {
					s.Open = false
				}
				beforeStart, beforeEnd := s.Start.Text, s.End.Text
				s.Start.set(v.Start)
				s.End.set(v.End)
				if o.OnChange != nil {
					if beforeStart != s.Start.Text {
						o.OnChange(s.Start.Text)
					}
					if beforeEnd != s.End.Text {
						o.OnChange(s.End.Text)
					}
				}
				publish()
			}
			Calendar(c, inner, value, &s.Calendar, opts)
		})
	})
	return root
}

func dateInputRangeAllowed(start, end Date, o CalendarOptions) bool {
	if !start.Valid() || !end.Valid() || start.Compare(end) > 0 {
		return false
	}
	o.Range, o.PreventDeselect, o.Disabled, o.ReadOnly = true, true, false, false
	o.FixedDate = CalendarNoFixedDate
	_, accepted := SelectDate(DateSelection{Start: Set(start)}, end, o)
	return accepted
}
