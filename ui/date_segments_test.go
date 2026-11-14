package nui_test

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestDateSegmentsTypingNavigationAndDraft(t *testing.T) {
	var value nui.Optional[nui.Date]
	var state nui.DateInputState
	changes, submits := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputDate(c, th, &value, &state, nui.InputDateOptions{Label: "Date", OnValueChange: func(nui.Optional[nui.Date]) { changes++ }, OnSubmit: func(string) { submits++ }})
	}, 400, 150)
	if _, ok := tt.Find("Choose Date"); ok {
		t.Fatal("default date field includes a calendar")
	}
	click(t, tt, "Date month")
	tt.Type("02")
	if !tt.Focused("Date day") || value.Set {
		t.Fatal("month did not advance or incomplete value committed")
	}
	tt.Type("29")
	if !tt.Focused("Date year") {
		t.Fatal("day did not advance")
	}
	for _, digit := range []string{"2", "0", "2", "4"} {
		tt.Type(digit)
	}
	if value.Value != (nui.Date{Year: 2024, Month: 2, Day: 29}) || state.Error != "" || changes != 1 {
		t.Fatalf("leap date: value=%v error=%q changes=%d", value, state.Error, changes)
	}
	tt.Key(0, ui.KeyUp)
	if value.Value != (nui.Date{Year: 2025, Month: 2, Day: 28}) {
		t.Fatalf("year increment did not clamp leap day: %v", value)
	}
	tt.Key(0, ui.KeyLeft)
	if !tt.Focused("Date day") {
		t.Fatal("left did not move to day")
	}
	tt.Key(0, ui.KeyDelete)
	tt.Key(0, ui.KeyEnter)
	if state.Error == "" || value.Value.Day != 28 || submits != 0 {
		t.Fatal("incomplete date changed committed value or submitted")
	}
	tt.Type("15")
	tt.Key(0, ui.KeyEnter)
	if value.Value.Day != 15 || submits != 1 {
		t.Fatal("completed date did not submit")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	if value.Set || state.Text != "" || state.Error != "" {
		t.Fatal("clear all did not unset the date")
	}
}

func TestDateSegmentsLocaleIdentityAndBounds(t *testing.T) {
	value := nui.Set(nui.Date{Year: 2024, Month: 2, Day: 20})
	var state nui.DateInputState
	shown := true
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		if shown {
			nui.InputDate(c.Key("date"), th, &value, &state, nui.InputDateOptions{Label: "Date", Calendar: nui.CalendarOptions{Min: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 10}), Max: nui.Set(nui.Date{Year: 2024, Month: 3, Day: 10})}, OnValueChange: func(nui.Optional[nui.Date]) { changes++ }})
		}
	}, 450, 150)
	if bounds(t, tt, "Date month").X >= bounds(t, tt, "Date year").X {
		t.Fatal("English date order")
	}
	click(t, tt, "Date day")
	tt.Key(0, ui.KeyDelete)
	draft := state.Text
	tt.SetLocale("zh-Hans")
	if !tt.Focused("Date 日") || bounds(t, tt, "Date 年").X >= bounds(t, tt, "Date 月").X || state.Text != draft || changes != 0 {
		t.Fatal("locale change lost focus, order, draft, or fired callback")
	}
	shown = false
	tt.Frame()
	tt.Frame()
	shown = true
	tt.Frame()
	if state.Text != draft {
		t.Fatal("unmount lost draft")
	}
	click(t, tt, "Date 日")
	tt.SetClipboard("2024-04-01")
	tt.Key(ui.Ctrl, ui.KeyV)
	if value.Value.Month != 2 || state.Error == "" || changes != 0 {
		t.Fatal("out-of-bounds paste committed")
	}
	tt.SetClipboard("2024/3/5")
	tt.Key(ui.Ctrl, ui.KeyV)
	if value.Value.Month != 3 || value.Value.Day != 5 || changes != 1 {
		t.Fatalf("localized paste: %v changes=%d", value, changes)
	}
	value = nui.Set(nui.Date{Year: 2024, Month: 2, Day: 12})
	tt.Frame()
	if state.Text != "2024-02-12" || changes != 1 || !tt.Focused("Date 日") {
		t.Fatal("controlled update lost focus or emitted change")
	}
}

func TestTimeSegmentsHourCycleAndHiddenSeconds(t *testing.T) {
	value := nui.Set(nui.ClockTime{Hour: 23, Minute: 59, Second: 42})
	var state nui.TimeInputState
	twelve, seconds := true, true
	changes := 0
	step := nui.TimeStep{}
	precision := nui.TimeGranularityDefault
	snap := false
	var minimum nui.Optional[nui.ClockTime]
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTime(c, th, &value, &state, nui.InputTimeOptions{Label: "Time", TwelveHour: nui.Set(twelve), Seconds: seconds, Granularity: precision, Step: step, StepSnapping: snap, Min: minimum, OnValueChange: func(nui.Optional[nui.ClockTime]) { changes++ }})
	}, 450, 150)
	click(t, tt, "Time AM/PM")
	tt.Type("a")
	if value.Value.Hour != 11 || changes != 1 {
		t.Fatal("AM change")
	}
	tt.Key(0, ui.KeyUp)
	if value.Value.Hour != 23 {
		t.Fatal("period increment")
	}
	twelve = false
	tt.Frame()
	if !tt.Focused("Time hour") || value.Value.Hour != 23 || changes != 2 {
		t.Fatal("hidden period lost focus or changed value")
	}
	click(t, tt, "Time second")
	seconds = false
	tt.Frame()
	if !tt.Focused("Time minute") || value.Value.Second != 42 {
		t.Fatal("hidden second lost focus or value")
	}
	tt.Key(0, ui.KeyUp)
	if value.Value.Minute != 0 || value.Value.Second != 42 {
		t.Fatal("minute wrap or hidden seconds")
	}
	tt.SetRegion(i18n.Region{Locale: "en-GB"})
	tt.Frame()
	if state.Text != "23:00" {
		t.Fatalf("24-hour text %q", state.Text)
	}
	tt.SetClipboard("06:45")
	tt.Key(ui.Ctrl, ui.KeyV)
	if value.Value != (nui.ClockTime{Hour: 6, Minute: 45, Second: 42}) {
		t.Fatalf("paste lost hidden seconds: %v", value)
	}
	value = nui.Set(nui.ClockTime{Hour: 11, Minute: 55, Second: 42})
	twelve, snap, step.Minute = true, true, 5
	tt.Frame()
	click(t, tt, "Time hour")
	tt.Key(0, ui.KeyUp)
	if value.Value.Hour != 12 {
		t.Fatal("hour increment did not cross noon")
	}
	click(t, tt, "Time minute")
	tt.Key(0, ui.KeyUp)
	if value.Value.Minute != 0 || value.Value.Hour != 12 {
		t.Fatal("minute step did not wrap without carrying")
	}
	tt.Type("08")
	tt.Frame()
	if value.Value.Minute != 10 || value.Value.Second != 42 {
		t.Fatalf("blur snapping: %v", value)
	}
	click(t, tt, "Time minute")
	before := changes
	precision = nui.TimeHour
	tt.Frame()
	if !tt.Focused("Time hour") || changes != before || value.Value.Minute != 10 {
		t.Fatal("hour precision lost focus or changed hidden minutes")
	}
	tt.SetClipboard("03 PM")
	tt.Key(ui.Ctrl, ui.KeyV)
	if value.Value != (nui.ClockTime{Hour: 15, Minute: 10, Second: 42}) {
		t.Fatalf("hour-only paste: %v", value)
	}
	minimum = nui.Set(nui.ClockTime{Hour: 16})
	tt.Frame()
	if state.Error == "" {
		t.Fatal("new bound did not invalidate existing value")
	}
	before = changes
	tt.SetClipboard("02 PM")
	tt.Key(ui.Ctrl, ui.KeyV)
	if changes != before || value.Value.Hour != 15 || state.Error == "" {
		t.Fatal("out-of-range time committed")
	}
	minimum = nui.Optional[nui.ClockTime]{}
	value = nui.Set(nui.ClockTime{Hour: 17, Minute: 12, Second: 42})
	tt.Frame()
	if state.Error != "" || changes != before {
		t.Fatal("controlled time update notified or retained error")
	}
}

func TestSegmentsReplaceAllAndFormFormatting(t *testing.T) {
	value := nui.Set(nui.ClockTime{Hour: 13, Minute: 30, Second: 42})
	var state nui.TimeInputState
	var form nui.FormState
	twelve, seconds := false, false
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Form(c, th, &form, nui.FormOptions{LoadingAuto: nui.Set(false)}, func(ft nui.Theme) {
			nui.FormField(c, ft, nui.FormFieldOptions{Name: "time", Label: "Time", Value: value.Value.String()}, func(it nui.Theme) {
				nui.InputTime(c, it, &value, &state, nui.InputTimeOptions{Label: "Time", TwelveHour: nui.Set(twelve), Seconds: seconds, OnValueChange: func(nui.Optional[nui.ClockTime]) { changes++ }})
			})
		})
	}, 450, 160)
	form.BeginAsync()
	twelve, seconds = true, true
	tt.Frame()
	if form.Field("time").Dirty || !form.Submitting || changes != 0 {
		t.Fatal("display options acted as an edit")
	}
	twelve, seconds = false, false
	tt.Frame()
	click(t, tt, "Time minute")
	tt.Key(ui.Ctrl, ui.KeyA)
	if !tt.Focused("Time hour") {
		t.Fatal("select all did not return to the first segment")
	}
	for _, digit := range []string{"0", "6", "4", "5"} {
		tt.Type(digit)
	}
	if value.Value != (nui.ClockTime{Hour: 6, Minute: 45, Second: 42}) || state.Error != "" {
		t.Fatalf("replacement %v %q", value, state.Error)
	}
	form.BeginAsync()
	tt.SetLocale("zh-Hans")
	if !form.Submitting {
		t.Fatal("locale canceled form work")
	}
	seconds = true
	tt.Frame()
	if !form.Submitting {
		t.Fatal("revealing retained seconds canceled form work")
	}
}

func TestSegmentNavigationDoesNotBlurFormField(t *testing.T) {
	var form nui.FormState
	var state nui.DateInputState
	var value nui.Optional[nui.Date]
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Form(c, th, &form, nui.FormOptions{}, func(ft nui.Theme) {
				nui.FormField(c, ft, nui.FormFieldOptions{Name: "date", Label: "Date"}, func(it nui.Theme) { nui.InputDate(c, it, &value, &state, nui.InputDateOptions{Label: "Date"}) })
			})
			nui.Button(c, th, "Outside", nui.ButtonOptions{})
		})
	}, 400, 180)
	click(t, tt, "Date month")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Date day") || form.Field("date").Touched {
		t.Fatal("segment navigation marked field touched")
	}
	click(t, tt, "Outside")
	if !form.Field("date").Touched {
		t.Fatal("leaving composite field did not mark it touched")
	}
}

func TestSegmentsScopedRTLNavigation(t *testing.T) {
	var state nui.DateInputState
	value := nui.Set(nui.Date{Year: 2024, Month: 2, Day: 20})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Direction(ui.RightToLeft).Children(func() { nui.InputDate(c, th, &value, &state, nui.InputDateOptions{Label: "Date"}) })
	}, 400, 140)
	if bounds(t, tt, "Date month").X <= bounds(t, tt, "Date day").X {
		t.Fatal("scoped RTL did not reverse layout")
	}
	click(t, tt, "Date month")
	tt.Key(0, ui.KeyLeft)
	if !tt.Focused("Date day") {
		t.Fatal("left did not follow scoped RTL")
	}
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Date month") {
		t.Fatal("right did not follow scoped RTL")
	}
}

func TestSegmentFieldGroupAndErrorFocus(t *testing.T) {
	var form nui.FormState
	var state nui.DateRangeInputState
	var value nui.DateSelection
	var first ui.Handle
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Form(c, th, &form, nui.FormOptions{}, func(ft nui.Theme) {
			nui.FormField(c, ft, nui.FormFieldOptions{Name: "range", Label: "Period"}, func(it nui.Theme) {
				nui.FieldGroup(c, it, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
					{ID: "range", Build: func(c *ui.Context, gt nui.Theme) ui.Element {
						return nui.InputDateRange(c, gt, &value, &state, nui.InputDateOptions{Label: "Period", Style: func(root, editor, leading, trailing ui.Element) { editor.Bind(&first) }})
					}},
					{ID: "next", Build: func(c *ui.Context, gt nui.Theme) ui.Element { return nui.Button(c, gt, "Next", nui.ButtonOptions{}) }},
				})
			})
		})
	}, 600, 160)
	click(t, tt, "Period end day")
	tt.Type("1")
	if form.Validate(nil) {
		t.Fatal("partial range passed validation")
	}
	form.Field("range").Handle.Focus()
	tt.Frame()
	if !tt.Focused("Period start month") {
		t.Fatal("form error did not focus first segment")
	}
	first.Focus()
	tt.Frame()
	if !tt.Focused("Period start month") {
		t.Fatal("Style editor handle is not a segment")
	}
	click(t, tt, "Period end year")
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Next") {
		t.Fatal("Tab did not leave the range for its group sibling")
	}
}

func TestSegmentsReadOnlyDisabledAndRangeNavigation(t *testing.T) {
	value := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 10}), End: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 20})}
	var state nui.DateRangeInputState
	readonly, disabled := false, false
	changes := 0
	rangeChanges, clears := 0, 0
	var maximum nui.Optional[nui.Date]
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Range", Disabled: disabled}, func(ft nui.Theme) {
			nui.InputDateRange(c, ft, &value, &state, nui.InputDateOptions{Label: "Range", ReadOnly: readonly, Clearable: true, OnClear: func() { clears++ }, OnRangeChange: func(nui.DateSelection) { rangeChanges++ }, Calendar: nui.CalendarOptions{Max: maximum, OnChange: func(nui.DateSelection) { changes++ }}})
		})
	}, 500, 180)
	click(t, tt, "Range start year")
	tt.Key(0, ui.KeyRight)
	if !tt.Focused("Range end month") {
		t.Fatal("range navigation stopped between endpoints")
	}
	tt.Key(ui.Shift, ui.KeyTab)
	if !tt.Focused("Range start year") {
		t.Fatal("reverse Tab did not cross endpoints")
	}
	readonly = true
	tt.Frame()
	before := value
	tt.Key(0, ui.KeyUp)
	tt.Type("2025")
	tt.SetClipboard("2024-03-01")
	tt.Key(ui.Ctrl, ui.KeyV)
	tt.Key(0, ui.KeyDelete)
	if value != before || changes != 0 {
		t.Fatal("read-only changed value")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Range start month")
	tt.Type("3")
	if value != before || tt.Focused("Range start month") {
		t.Fatal("inherited disabled allowed input")
	}
	disabled, readonly = false, false
	tt.Frame()
	click(t, tt, "Range start month")
	tt.SetClipboard("2024-03-01")
	tt.Key(ui.Ctrl, ui.KeyV)
	if value.End.Set || changes != 1 {
		t.Fatal("start edit did not invalidate earlier end exactly once")
	}
	click(t, tt, "Range end month")
	tt.SetClipboard("2024-02-15")
	tt.Key(ui.Ctrl, ui.KeyV)
	if value.End.Set || state.End.Error == "" || changes != 1 {
		t.Fatal("reverse range committed")
	}
	value = nui.DateSelection{Start: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 10}), End: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 20})}
	tt.Frame()
	click(t, tt, "Clear Range")
	if value.Start.Set || value.End.Set || changes != 2 || rangeChanges != 2 || clears != 1 {
		t.Fatalf("range clear was not atomic: %v %d %d %d", value, changes, rangeChanges, clears)
	}
	maximum = nui.Set(nui.Date{Year: 2024, Month: 2, Day: 29})
	value = nui.DateSelection{Start: nui.Set(nui.Date{Year: 2025, Month: 1, Day: 1})}
	tt.Frame()
	if state.Start.Error == "" || changes != 2 {
		t.Fatal("external out-of-bounds start was not reported quietly")
	}
}

func TestSegmentVisualArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		var date nui.DateInputState
		var clock nui.TimeInputState
		var interval nui.DateRangeInputState
		value := nui.Set(nui.Date{Year: 2024, Month: 2, Day: 29})
		tm := nui.Set(nui.ClockTime{Hour: 13, Minute: 5, Second: 42})
		rangeValue := nui.DateSelection{Start: value, End: nui.Set(nui.Date{Year: 2024, Month: 3, Day: 15})}
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.Light()
			if dark {
				th = nui.Dark()
			}
			nui.Window(c, th)
			ui.Column(c).Padding(24).Gap(18).Children(func() {
				nui.InputDate(c, th, &value, &date, nui.InputDateOptions{Label: "Date", ShowCalendar: true})
				nui.InputTime(c, th, &tm, &clock, nui.InputTimeOptions{Label: "Time", Seconds: true, TwelveHour: nui.Set(true)})
				nui.InputDateRange(c, th, &rangeValue, &interval, nui.InputDateOptions{Label: "Range"})
			})
		}, 540, 230)
		click(t, tt, "Date day")
		name := "segments-light.png"
		if dark {
			name = "segments-dark.png"
		}
		file, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(file, tt.Image())
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
	}
}
