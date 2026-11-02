package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/internal/databrowser"
)

type formsDataState struct {
	RangeInput        nui.DateRangeInputState
	Number            nui.Optional[float64]
	NumberState       nui.NumberState
	NumberOrientation nui.Orientation
	Tags              []string
	TagsState         nui.TagsState
	Pin               string
	PinState          nui.PinState
	Rating            int
	Tri               nui.CheckState
	Low, High         float64
	Chosen            []string
	List              nui.ListboxState
	Select            nui.SelectMenuState
	Date              nui.Optional[nui.Date]
	DateInput         nui.DateInputState
	Time              nui.Optional[nui.ClockTime]
	TimeInput         nui.TimeInputState
	Range             nui.DateSelection
	Calendar          nui.CalendarState
	Color             ui.Color
	ColorState        nui.ColorPickerState
	Browser           *databrowser.State
}

func newFormsData() formsDataState {
	return formsDataState{Number: nui.Set(3.0), Tags: []string{"Go", "Desktop"}, Rating: 3, Tri: nui.Indeterminate, Low: 20, High: 70, Date: nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 1}), Time: nui.Set(nui.ClockTime{Hour: 9, Minute: 30}), Color: ui.Hex("#22c55e"), Browser: databrowser.New(10000)}
}

var projectMembers = []nui.ListboxItem{{ID: "lin", Label: "Lin Chen", Group: "Design"}, {ID: "ada", Label: "Ada Lovelace", Group: "Engineering"}, {ID: "grace", Label: "Grace Hopper", Group: "Engineering"}, {ID: "guest", Label: "Guest account", Group: "External", Disabled: true}}

func (a *State) advancedForms(c *ui.Context, t nui.Theme) {
	s := &a.FormsData
	heading(c, t, "Advanced forms", "Reusable values, selection and validation for desktop workflows.")
	layout := ui.Row(c).Gap(24).AlignItems(ui.Start)
	width, _ := c.Size()
	if width < 1000 {
		layout.Column().AlignItems(ui.Stretch).Shrink(0)
	}
	layout.Children(func() {
		var column ui.Element
		column = ui.Column(c).Grow(1).MinWidth(0).Gap(16)
		if width < 1000 {
			column.Grow(0).Shrink(0)
		}
		column.Children(func() {
			nui.FormField(c.Key("number-orientation"), t, nui.FormFieldOptions{Label: "Orientation"}, func(ft nui.Theme) {
				nui.Select(c, ft, &s.NumberOrientation, []nui.Option[nui.Orientation]{
					{Value: nui.Horizontal, Label: "horizontal"},
					{Value: nui.Vertical, Label: "vertical"},
				}, nui.SelectOptions{Label: "Number orientation"}).Width(180)
			})
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Seats"}, func(ft nui.Theme) {
				nui.InputNumber(c, ft, &s.Number, &s.NumberState, nui.InputNumberOptions{Label: "Seats", Orientation: s.NumberOrientation, Min: nui.Set(0.0), Max: nui.Set(100.0)})
			})
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Project tags", Help: "Paste comma-separated tags or press Enter to add."}, func(ft nui.Theme) {
				nui.InputTags(c, ft, &s.Tags, &s.TagsState, nui.InputTagsOptions{Label: "Project tags", MaxTags: 8})
			})
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Verification code"}, func(ft nui.Theme) {
				nui.PinInput(c, ft, &s.Pin, &s.PinState, nui.PinInputOptions{Numeric: true, Label: "Verification code", Length: 6})
			})
			nui.InputRating(c, t, &s.Rating, nui.InputRatingOptions{Label: "Project rating"})
			nui.CheckboxTri(c, t, &s.Tri, nui.CheckboxTriOptions{Label: "Select all permissions", Description: "A mixed value becomes checked on activation."})
			nui.RangeSlider(c, t, &s.Low, &s.High, nui.RangeSliderOptions{Label: "Budget range", Min: 0, Max: 100, Step: 5})
		})
		column = ui.Column(c).Grow(1).MinWidth(0).Gap(16)
		if width < 1000 {
			column.Grow(0).Shrink(0)
		}
		column.Children(func() {
			nui.SelectMenu(c, t, &s.Chosen, &s.Select, projectMembers, nui.SelectMenuOptions{Label: "Project members", Multiple: true, Placeholder: "Choose project members", SearchPlaceholder: "Search members", Clearable: true, OnCreate: func(label string) { a.Message = "Requested invitation for " + label }})
			nui.Listbox(c, t, &s.Chosen, &s.List, projectMembers, nui.ListboxOptions{Label: "Members list", Multiple: true, Height: 230})
			nui.Button(c, t, "Open project setup", nui.ButtonOptions{Variant: nui.Soft}).OnClick(func() { a.Settings.Page = 3; a.Page = 7 })
			ui.Text(c, a.Message).TextColor(t.Muted)
		})
	})
}
func (a *State) dateComponents(c *ui.Context, t nui.Theme) {
	s := &a.FormsData
	heading(c, t, "Dates & color", "Civil dates, wall-clock times and synchronized color controls.")
	layout := ui.Row(c).Gap(24).AlignItems(ui.Start)
	width, _ := c.Size()
	if width < 1000 {
		layout.Column().AlignItems(ui.Stretch).Shrink(0)
	}
	layout.Children(func() {
		column := ui.Column(c).Grow(1).MinWidth(0).Gap(16)
		if width < 1000 {
			column.Grow(0).Shrink(0)
		}
		column.Children(func() {
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Release date"}, func(ft nui.Theme) {
				nui.InputDate(c, ft, &s.Date, &s.DateInput, nui.InputDateOptions{Label: "Release date"})
			})
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Reminder"}, func(ft nui.Theme) {
				nui.InputTime(c, ft, &s.Time, &s.TimeInput, nui.InputTimeOptions{Label: "Reminder"})
			})
			nui.InputDateRange(c, t, &s.Range, &s.RangeInput, nui.InputDateOptions{Label: "Release window"})
			nui.Calendar(c, t, &s.Range, &s.Calendar, nui.CalendarOptions{Label: "Release range", Range: true, Today: nui.Set(nui.Date{Year: 2026, Month: 11, Day: 1})})
		})
		nui.ColorPicker(c, t, &s.Color, &s.ColorState, nui.ColorPickerOptions{Label: "Accent color"})
	})
}
