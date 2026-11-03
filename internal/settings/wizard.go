package settings

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type WizardState struct {
	Form                 nui.FormState
	Step                 int
	Name, Key            string
	CustomKey            bool
	Minimum, Maximum     nui.Optional[float64]
	MinInput, MaxInput   nui.NumberState
	Start, End           nui.Optional[nui.Date]
	StartInput, EndInput nui.DateInputState
	Ticket               nui.FormTicket
	Pending              bool
	Status               string
}

func (a *State) wizard(c *ui.Context, t nui.Theme) {
	s := &a.Wizard
	title(c, t, "Project setup", "Create a project in two steps. Review errors before confirming.")
	step := []string{"details", "schedule"}[s.Step]
	nui.Stepper(c, t, &step, []nui.StepItem{{ID: "details", Title: "Project"}, {ID: "schedule", Title: "Schedule"}}, nui.StepperOptions{CanSelect: func(string) bool { return false }})
	nui.Form(c, t, &s.Form, nui.FormOptions{Label: "Project setup form", Validate: func(v nui.FormValues) nui.FormErrors {
		e := nui.FormErrors{}
		if s.Step == 0 {
			if strings.TrimSpace(v["name"]) == "" {
				e["name"] = "Enter a project name"
			}
			if key, ok := v["key"]; ok && (key == "" || strings.Contains(key, " ")) {
				e["key"] = "Use a key without spaces"
			}
		} else {
			if !s.Start.Set {
				e["start"] = "Choose a start date"
			}
			if !s.End.Set || s.Start.Set && s.End.Value.Compare(s.Start.Value) < 0 {
				e["end"] = "End must be on or after start"
			}
			if s.Minimum.Set && s.Maximum.Set && s.Maximum.Value < s.Minimum.Value {
				e["maximum"] = "Maximum must be at least the minimum"
			}
		}
		return e
	}, OnSubmit: func(nui.FormValues) {
		if s.Step == 0 {
			s.Step = 1
			return
		}
		s.Ticket = s.Form.BeginAsync()
		s.Pending = true
		s.Status = "Review the project before confirming"
	}}, func(inner nui.Theme) {
		if s.Step == 0 {
			nui.FormField(c.Key("project-name"), inner, nui.FormFieldOptions{Name: "name", Value: s.Name, Label: "Project name", Required: true}, func(ft nui.Theme) {
				nui.Input(c, ft, &s.Name, nui.InputOptions{Label: "Project name input", Placeholder: "Desktop workspace"})
			})
			nui.Switch(c, inner, &s.CustomKey, nui.SwitchOptions{Label: "Custom project key"})
			if s.CustomKey {
				nui.FormField(c.Key("project-key"), inner, nui.FormFieldOptions{Name: "key", Value: s.Key, Label: "Project key"}, func(ft nui.Theme) { nui.Input(c, ft, &s.Key, nui.InputOptions{Placeholder: "desktop-tools"}) })
			}
		} else {
			for i, label := range []string{"Minimum budget", "Maximum budget"} {
				value, input, name := &s.Minimum, &s.MinInput, "minimum"
				if i == 1 {
					value, input, name = &s.Maximum, &s.MaxInput, "maximum"
				}
				nui.FormField(c.Key(name), inner, nui.FormFieldOptions{Name: name, Value: fmt.Sprint(*value), Label: label}, func(ft nui.Theme) {
					nui.InputNumber(c, ft, value, input, nui.InputNumberOptions{Label: label, Min: nui.Set(0.0), Step: 100})
				})
			}
			for i, label := range []string{"Start date", "End date"} {
				value, input, name := &s.Start, &s.StartInput, "start"
				if i == 1 {
					value, input, name = &s.End, &s.EndInput, "end"
				}
				nui.FormField(c.Key(name), inner, nui.FormFieldOptions{Name: name, Value: value.Value.String(), Label: label, Required: true}, func(ft nui.Theme) {
					nui.InputDate(c, ft, value, input, nui.InputDateOptions{ShowCalendar: true, Label: label, Calendar: nui.CalendarOptions{Today: nui.Set(nui.Date{Year: 2026, Month: time.November, Day: 1})}})
				})
			}
		}
		ui.Row(c).Gap(8).Children(func() {
			if s.Step > 0 {
				nui.Button(c, inner, "Back", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Form.Cancel(); s.Pending = false; s.Step = 0 })
			}
			label := "Continue"
			if s.Step == 1 {
				label = "Create project"
			}
			nui.Button(c, inner, label, nui.ButtonOptions{Loading: s.Form.Submitting}).OnClick(s.Form.Submit)
			if s.Pending && s.Form.Submitting {
				nui.Button(c, inner, "Confirm project", nui.ButtonOptions{Variant: nui.Soft}).OnClick(func() {
					if s.Form.CompleteAsync(s.Ticket, nil) {
						s.Pending = false
						s.Status = "Project created"
						nui.Toast(c, t, nui.ToastOptions{Title: "Project created", Description: s.Name, Color: nui.Success})
					}
				})
				nui.Button(c, inner, "Cancel creation", nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { s.Form.Cancel(); s.Pending = false; s.Status = "Creation cancelled" })
			}
		})
		if s.Status != "" {
			ui.Text(c, s.Status).Role(ui.RoleStatus).TextColor(t.Muted)
		}
	})
}
