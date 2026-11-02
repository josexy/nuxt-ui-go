package nui_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestNumberIntermediateAndBounds(t *testing.T) {
	var value nui.Optional[float64]
	var s nui.NumberState
	changes := 0
	readOnly := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputNumber(c, th, &value, &s, nui.InputNumberOptions{Label: "Amount", ReadOnly: readOnly, Min: nui.Set(-10.0), Max: nui.Set(10.0), Step: .25, OnValueChange: func(nui.Optional[float64]) { changes++ }}).Width(320)
	}, 400, 180)
	click(t, tt, "Amount")
	tt.Type("-")
	if value.Set || s.Text != "-" {
		t.Fatalf("incomplete edit %v %q", value, s.Text)
	}
	tt.Type("2.5")
	if !value.Set || value.Value != -2.5 {
		t.Fatalf("number %v", value)
	}
	click(t, tt, "Increase Amount")
	if value.Value != -2.25 {
		t.Fatalf("step %v", value)
	}
	click(t, tt, "Amount")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("99")
	if value.Value != -2.25 || s.Text != "99" {
		t.Fatal("out of range committed")
	}
	tt.Key(0, ui.KeyEnter)
	if s.Text != "-2.25" {
		t.Fatal("commit did not restore")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	if value.Set {
		t.Fatal("empty value")
	}
	readOnly = true
	tt.Frame()
	before := changes
	click(t, tt, "Increase Amount")
	tt.Type("4")
	if changes != before || value.Set {
		t.Fatal("read-only changed")
	}
	for _, text := range []string{"NaN", "Inf", "1e999", "11"} {
		if _, ok := nui.ParseNumber(text, nui.InputNumberOptions{Max: nui.Set(10.0)}); ok {
			t.Fatalf("accepted %s", text)
		}
	}
}
func TestTagsPinRatingTriAndRange(t *testing.T) {
	tags := []string{"Go"}
	var ts nui.TagsState
	pin := ""
	var ps nui.PinState
	complete := 0
	rating := 0
	tri := nui.Indeterminate
	lo, hi := 20.0, 70.0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(12).Gap(12).Children(func() {
			nui.InputTags(c, th, &tags, &ts, nui.InputTagsOptions{Label: "Tags", AddOnPaste: nui.Set(true), MaxTags: 4})
			nui.PinInput(c, th, &pin, &ps, nui.PinInputOptions{Numeric: true, Label: "Code", Length: 4, OnComplete: func(string) { complete++ }})
			nui.InputRating(c, th, &rating, nui.InputRatingOptions{Label: "Rating"})
			nui.CheckboxTri(c, th, &tri, nui.CheckboxTriOptions{Label: "All members"})
			nui.RangeSlider(c, th, &lo, &hi, nui.RangeSliderOptions{Label: "Budget", Min: 0, Max: 100, Step: 5})
		})
	}, 500, 600)
	click(t, tt, "Tags")
	tt.SetClipboard("Go, Vue\n原生")
	tt.Command("paste")
	if !slices.Equal(tags, []string{"Go", "Vue", "原生"}) {
		t.Fatalf("tags %v", tags)
	}
	tt.Key(0, ui.KeyBackspace)
	if len(tags) != 2 {
		t.Fatal("empty backspace")
	}
	click(t, tt, "Remove Vue")
	if len(tags) != 1 {
		t.Fatal("remove tag")
	}
	click(t, tt, "Code")
	tt.SetClipboard("12a345")
	tt.Command("paste")
	tt.Frame()
	if pin != "1234" || complete != 1 {
		t.Fatalf("PIN %q complete=%d", pin, complete)
	}
	tt.Key(0, ui.KeyBackspace)
	if pin != "123" {
		t.Fatalf("backspace PIN %q", pin)
	}
	click(t, tt, "Rating")
	tt.Key(0, ui.KeyEnd)
	if rating != 5 {
		t.Fatalf("rating %d", rating)
	}
	tt.Key(0, ui.KeyLeft)
	if rating != 4 {
		t.Fatal("rating keyboard")
	}
	click(t, tt, "All members")
	if tri != nui.Checked {
		t.Fatalf("tri %v", tri)
	}
	tt.Key(0, ui.KeySpace)
	if tri != nui.Unchecked {
		t.Fatal("tri toggle")
	}
	if lo > hi || lo < 0 || hi > 100 {
		t.Fatal("range bounds")
	}
}
func TestFormRegistrationHiddenFieldsAndStaleValidation(t *testing.T) {
	var state nui.FormState
	name, email := "", ""
	hide := false
	submitted := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Form(c, th, &state, nui.FormOptions{Validate: func(v nui.FormValues) nui.FormErrors {
			e := nui.FormErrors{}
			if v["name"] == "" {
				e["name"] = "Required"
			}
			if email, ok := v["email"]; ok && email == v["name"] {
				e["email"] = "Must differ"
			}
			return e
		}, OnSubmit: func(nui.FormValues) { submitted++ }}, func(inner nui.Theme) {
			nui.FormField(c.Key("name"), inner, nui.FormFieldOptions{Name: "name", Value: name, Label: "Name"}, func(ft nui.Theme) { nui.Input(c, ft, &name, nui.InputOptions{Label: "Name input"}) })
			if !hide {
				nui.FormField(c.Key("email"), inner, nui.FormFieldOptions{Name: "email", Value: email, Label: "Email"}, func(ft nui.Theme) { nui.Input(c, ft, &email, nui.InputOptions{}) })
			}
			nui.Button(c, inner, "Submit", nui.ButtonOptions{}).OnClick(state.Submit)
		})
	}, 500, 400)
	click(t, tt, "Submit")
	if submitted != 0 || !tt.Focused("Name input") {
		t.Fatal("first invalid focus")
	}
	tt.Type("Alice")
	tt.Frame()
	if !state.Field("name").Dirty {
		t.Fatal("dirty")
	}
	click(t, tt, "Submit")
	if submitted != 1 {
		t.Fatal("valid submit")
	}
	old := state.BeginAsync()
	name = "Bob"
	tt.Frame()
	if state.CompleteAsync(old, nui.FormErrors{"name": "Old result"}) {
		t.Fatal("stale validation accepted")
	}
	current := state.BeginAsync()
	state.Cancel()
	if state.CompleteAsync(current, nil) {
		t.Fatal("cancelled submit accepted")
	}
	hide = true
	tt.Frame()
	if _, ok := state.Values()["email"]; ok {
		t.Fatal("hidden value submitted")
	}
	hide = false
	tt.Frame()
	if !state.Field("email").Visible {
		t.Fatal("reappearing field")
	}
}
func TestQueryCancellationAndListbox(t *testing.T) {
	var query nui.QueryState[nui.ListboxItem]
	ctx, old := query.Start(context.Background(), "a")
	_, newer := query.Start(context.Background(), "ab")
	if ctx.Err() == nil {
		t.Fatal("old work not cancelled")
	}
	if query.Complete(old, nil, nil) {
		t.Fatal("stale results")
	}
	if !query.Complete(newer, nil, errors.New("offline")) || query.Error != "offline" {
		t.Fatal("error state")
	}
	items := []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Blocked", Disabled: true}, {ID: "c", Label: "Charlie"}}
	var s nui.ListboxState
	selected := []string{}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, &selected, &s, items, nui.ListboxOptions{Label: "Options", Multiple: true})
	}, 400, 300)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if !slices.Contains(selected, "c") {
		t.Fatalf("disabled skip %v active=%s", selected, s.Active)
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeySpace)
	if len(selected) != 2 {
		t.Fatal("multiple selection")
	}
	items = items[1:]
	tt.Frame()
	if s.Active == "a" {
		t.Fatal("removed active identity")
	}
}
func TestCalendarCivilBoundariesAndTime(t *testing.T) {
	d := nui.Date{Year: 2024, Month: time.January, Day: 31}
	if got := d.AddMonths(1); got.Day != 29 || got.Month != time.February {
		t.Fatal(got)
	}
	if (nui.Date{Year: 2023, Month: time.February, Day: 29}).Valid() {
		t.Fatal("non leap date")
	}
	if got := d.AddMonths(12); got.Year != 2025 || got.Day != 31 {
		t.Fatal(got)
	}
	o := nui.CalendarOptions{Range: true}
	first, _ := nui.SelectDate(nui.DateSelection{}, nui.Date{Year: 2024, Month: 2, Day: 29}, o)
	if first.End.Set {
		t.Fatal("unfinished range")
	}
	v, ok := nui.SelectDate(first, nui.Date{Year: 2024, Month: 2, Day: 20}, o)
	if !ok || v.Start.Value.Day != 20 || v.End.Value.Day != 29 {
		t.Fatal(v)
	}
	o.IsDateDisabled = func(d nui.Date) bool { return d.Day == 25 }
	if _, ok := nui.SelectDate(first, nui.Date{Year: 2024, Month: 2, Day: 20}, o); ok {
		t.Fatal("disabled interior")
	}
	for _, text := range []string{"12:00 AM", "12:00 PM", "01:15 PM"} {
		v, ok := nui.ParseClockTime(text, true, false)
		if !ok || !v.Value.Valid() {
			t.Fatal(text)
		}
	}
	var value nui.DateSelection
	var state nui.CalendarState
	state.Month = nui.Date{Year: 2024, Month: 2, Day: 1}
	state.Active = state.Month
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Calendar(c, th, &value, &state, nui.CalendarOptions{Label: "Calendar", Range: true, Today: nui.Set(state.Month)})
	}, 400, 420)
	click(t, tt, "Calendar 2024-02-29")
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEnter)
	if value.End.Value != (nui.Date{Year: 2024, Month: 3, Day: 1}) {
		t.Fatalf("month boundary %v", value)
	}
}
func TestDateTimeEditorsPreserveValues(t *testing.T) {
	date := nui.Set(nui.Date{Year: 2024, Month: 2, Day: 29})
	var ds nui.DateInputState
	tm := nui.Set(nui.ClockTime{Hour: 13, Minute: 15, Second: 30})
	var ts nui.TimeInputState
	twelve := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Gap(12).Children(func() {
			nui.InputDate(c, th, &date, &ds, nui.InputDateOptions{Label: "Date"})
			nui.InputTime(c, th, &tm, &ts, nui.InputTimeOptions{Label: "Time", TwelveHour: nui.Set(twelve)})
		})
	}, 420, 300)
	click(t, tt, "Date")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("2023-02-29")
	if date.Value.Year != 2024 || ds.Error == "" {
		t.Fatal("invalid date accepted")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("2024-03-01")
	if date.Value.Day != 1 || ds.Error != "" {
		t.Fatal("valid date rejected")
	}
	twelve = true
	tt.Frame()
	if ts.Text != "01:15 PM" || tm.Value.Second != 30 {
		t.Fatalf("time format %s %v", ts.Text, tm)
	}
	click(t, tt, "Time")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("12:00 AM")
	if tm.Value.Hour != 0 || tm.Value.Second != 30 {
		t.Fatal("12-hour conversion")
	}
}
func TestTableVirtualizationAndTreeLazy(t *testing.T) {
	for _, count := range []int{100, 10000, 100000} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var state nui.TableState[string]
			rows := make([]string, count)
			for i := range rows {
				rows[i] = fmt.Sprintf("Row %06d", i)
			}
			state.SetRows(rows, func(v string) string { return v })
			cols := []nui.TableColumn[string]{{ID: "name", Title: "Name", Width: 300, Value: func(v string) string { return v }, Compare: strings.Compare}}
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Table(c, th, &state, cols, nui.TableOptions{Label: "Records", Height: 320}).Width(400)
			}, 500, 400)
			if state.CellBuilds > 32 {
				t.Fatalf("built %d for %d rows", state.CellBuilds, count)
			}
			t.Logf("rows=%d cells=%d", count, state.CellBuilds)
			click(t, tt, "Row 000000")
			click(t, tt, "Name")
			click(t, tt, "Name")
			if !state.Sort.Descending || state.Row(0) != rows[count-1] || !state.Selection.Has(rows[0]) {
				t.Fatal("sort/identity")
			}
			state.SetRows(rows[1:], func(v string) string { return v })
			tt.Frame()
			if state.Selection.Has(rows[0]) {
				t.Fatal("removed selection")
			}
		})
	}
	var tree nui.TreeState
	tree.SetNodes([]nui.TreeNode{{ID: "root", Label: "Root", Lazy: true}})
	requests := 0
	var ticket nui.TreeTicket
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Tree(c, th, &tree, nui.TreeOptions{Label: "Files", OnLoad: func(id string, v nui.TreeTicket) { requests++; ticket = v }})
	}, 420, 300)
	click(t, tt, "Root")
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if requests != 1 {
		t.Fatalf("lazy requests %d", requests)
	}
	old := ticket
	tt.Key(0, ui.KeyLeft)
	tt.Frame()
	if tree.CompleteLoad(old, nil, nil) {
		t.Fatal("collapsed request accepted")
	}
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if !tree.CompleteLoad(ticket, nil, errors.New("offline")) {
		t.Fatal("load failure")
	}
	tt.Frame()
	click(t, tt, "Retry Root")
	tt.Frame()
	if requests != 3 {
		t.Fatalf("retry %d", requests)
	}
	if !tree.CompleteLoad(ticket, []nui.TreeNode{{ID: "child", Label: "Child"}}, nil) {
		t.Fatal("loaded data")
	}
	tt.Frame()
	if !tt.HasText("Child") {
		t.Fatal("lazy child missing")
	}
}

func TestInvalidEditorBlocksFormAndCancelsTicket(t *testing.T) {
	value := nui.Set(2.0)
	var ns nui.NumberState
	var form nui.FormState
	submitted := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Form(c, th, &form, nui.FormOptions{OnSubmit: func(nui.FormValues) { submitted++ }}, func(ft nui.Theme) {
			nui.FormField(c, ft, nui.FormFieldOptions{Name: "amount", Value: fmt.Sprint(value), Label: "Amount"}, func(it nui.Theme) {
				nui.InputNumber(c, it, &value, &ns, nui.InputNumberOptions{Label: "Budget input", Min: nui.Set(0.0), Max: nui.Set(10.0)})
			})
			nui.Button(c, ft, "Submit amount", nui.ButtonOptions{}).OnClick(form.Submit)
		})
	}, 420, 220)
	ticket := form.BeginAsync()
	click(t, tt, "Budget input")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("-")
	tt.Frame()
	if form.CompleteAsync(ticket, nil) {
		t.Fatal("invalid edit did not cancel result")
	}
	click(t, tt, "Submit amount")
	if submitted != 0 || !tt.Focused("Budget input") {
		t.Fatal("invalid text submitted or lost first-error focus")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("3")
	tt.Key(0, ui.KeyUp)
	if value.Value != 4 {
		t.Fatalf("keyboard number step: %v", value)
	}
	click(t, tt, "Submit amount")
	if submitted != 1 {
		t.Fatal("valid number not submitted")
	}
}
func TestLargeListboxKeyboardVirtualization(t *testing.T) {
	for _, count := range []int{100, 10000, 100000} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			items := make([]nui.ListboxItem, count)
			for i := range items {
				items[i] = nui.ListboxItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Choice %06d", i)}
			}
			var state nui.ListboxState
			var selected []string
			calls := 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				calls = 0
				nui.Listbox(c, th, &selected, &state, items, nui.ListboxOptions{Label: "Large choices", Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Item {
						calls++
					}
				}})
			}, 400, 300)
			tt.Key(0, ui.KeyTab)
			tt.Key(0, ui.KeyEnd)
			tt.Key(0, ui.KeyEnter)
			if !slices.Equal(selected, []string{fmt.Sprint(count - 1)}) {
				t.Fatalf("offscreen selected %v", selected)
			}
			if calls > 15 {
				t.Fatalf("built %d rows", calls)
			}
			t.Logf("options=%d rows=%d", count, calls)
		})
	}
}
func TestRangeKeyboardEndpoints(t *testing.T) {
	low, high := 20.0, 70.0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{Label: "Range", Max: 100, Step: 5}).Width(240)
	}, 320, 120)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyRight)
	if low != 25 {
		t.Fatalf("minimum %g", low)
	}
	tt.Key(0, ui.KeyEnd)
	if low != 25 || high != 100 || !tt.Focused("Range maximum") {
		t.Fatal("End must move the last thumb to maximum")
	}
	tt.Key(0, ui.KeyHome)
	if low != 0 || high != 100 || !tt.Focused("Range minimum") {
		t.Fatal("Home must move the first thumb to minimum")
	}
}
func TestTableInsertionResizeAndScrollIdentity(t *testing.T) {
	var state nui.TableState[string]
	rows := make([]string, 200)
	for i := range rows {
		rows[i] = fmt.Sprintf("Record %03d", i)
	}
	state.SetRows(rows, func(s string) string { return s })
	cols := []nui.TableColumn[string]{{ID: "name", Title: "Name", Width: 250, MinWidth: 150, MaxWidth: 400, Value: func(s string) string { return s }, Compare: strings.Compare}, {ID: "id", Title: "ID", Width: 150, Value: func(s string) string { return s }}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Table(c, th, &state, cols, nui.TableOptions{Label: "Table", Height: 300}).Width(420)
	}, 500, 400)
	click(t, tt, "Record 000")
	state.SetRows(append([]string{"Inserted"}, rows...), func(s string) string { return s })
	tt.Frame()
	if !state.Selection.Has("Record 000") {
		t.Fatal("insertion lost selection")
	}
	r := bounds(t, tt, "Name")
	table := bounds(t, tt, "Table")
	x, y := table.X+247, r.Y+r.H/2
	tt.Press(x, y)
	tt.Move(x+30, y)
	tt.Release(x+30, y)
	if state.List.Columns.Widths["name"] <= 250 {
		t.Fatalf("column resize: header=%+v table=%+v widths=%+v", r, table, state.List.Columns.Widths)
	}
	tt.Scroll(300, 200, 0, 3000)
	tt.Frame()
	if !state.Selection.Has("Record 000") {
		t.Fatal("scroll changed business selection")
	}
}

func TestTicketsBelongToTheirState(t *testing.T) {
	var a, b nui.FormState
	ta := a.BeginAsync()
	b.BeginAsync()
	if b.CompleteAsync(ta, nil) {
		t.Fatal("foreign form ticket")
	}
	var x, y nui.QueryState[int]
	_, tx := x.Start(context.Background(), "a")
	y.Start(context.Background(), "a")
	defer x.Cancel()
	defer y.Cancel()
	if y.Complete(tx, []int{1}, nil) {
		t.Fatal("foreign search ticket")
	}
}

func TestColorRepresentationsPreserveOtherChannels(t *testing.T) {
	value := ui.Hex("#22c55e")
	var state nui.ColorPickerState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ColorPicker(c, th, &value, &state, nui.ColorPickerOptions{Label: "Color"})
	}, 400, 650)
	click(t, tt, "Hex")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("#11223344")
	tt.Frame()
	if value != (ui.Color{R: 17, G: 34, B: 51, A: 68}) {
		t.Fatalf("HEX %v", value)
	}
	click(t, tt, "Color R")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("100")
	if value != (ui.Color{R: 100, G: 34, B: 51, A: 68}) {
		t.Fatalf("RGB %v", value)
	}
	click(t, tt, "Opacity")
	tt.Key(0, ui.KeyEnd)
	tt.Frame()
	if value != (ui.Color{R: 100, G: 34, B: 51, A: 255}) {
		t.Fatalf("alpha %v", value)
	}
}
func TestTreeVisibleRowWork(t *testing.T) {
	for _, count := range []int{100, 10000, 100000} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			nodes := make([]nui.TreeNode, count)
			for i := range nodes {
				nodes[i] = nui.TreeNode{ID: fmt.Sprint(i), Label: fmt.Sprintf("File %06d", i)}
			}
			var state nui.TreeState
			state.SetNodes(nodes)
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Tree(c, th, &state, nui.TreeOptions{Label: "Files", Height: 300})
			}, 400, 400)
			tt.Key(0, ui.KeyTab)
			tt.Key(0, ui.KeyEnd)
			tt.Frame()
			if state.RowBuilds > 24 {
				t.Fatalf("built %d rows", state.RowBuilds)
			}
			if !state.Selection.Has(fmt.Sprint(count - 1)) {
				t.Fatal("tree end navigation")
			}
			t.Logf("nodes=%d rows=%d", count, state.RowBuilds)
		})
	}
}

func TestDateRangeInputSynchronizesEndpoints(t *testing.T) {
	value := nui.DateSelection{Start: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 20}), End: nui.Set(nui.Date{Year: 2024, Month: 2, Day: 29})}
	var state nui.DateRangeInputState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputDateRange(c, th, &value, &state, nui.InputDateOptions{Label: "Period"})
	}, 600, 500)
	click(t, tt, "Period start")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("2024-03-01")
	if value.End.Set {
		t.Fatal("end before new start survived")
	}
	click(t, tt, "Choose Period")
	click(t, tt, "Period calendar 2024-03-10")
	if value.End.Value.Day != 10 || state.Open {
		t.Fatal("range popup did not complete")
	}
	if state.End.Text != "2024-03-10" {
		t.Fatalf("end text %q", state.End.Text)
	}
}
