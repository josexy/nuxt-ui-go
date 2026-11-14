package nui_test

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestFormFieldWithoutHeadingHasNoGap(t *testing.T) {
	for _, required := range []bool{false, true} {
		value := "Value"
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.FormField(c, th, nui.FormFieldOptions{Required: required}, func(ft nui.Theme) {
				nui.Input(c, ft, &value, nui.InputOptions{Label: "Editor"}).Width(200).Label("Editor shell")
			}).Label("Field root")
		}, 300, 100)
		root, input := bounds(t, tt, "Field root"), bounds(t, tt, "Editor shell")
		if root.Y != input.Y || root.H != input.H || tt.HasText("*") {
			t.Fatalf("required=%v empty heading left visible content or a gap: field=%+v input=%+v", required, root, input)
		}
	}
}

func TestFormFieldHorizontalBaseline(t *testing.T) {
	value := "Same baseline"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Name", Orientation: nui.Set(nui.Horizontal)}, func(ft nui.Theme) {
			nui.Input(c, ft, &value, nui.InputOptions{Label: "Editor"}).Width(200)
		}).Width(340).Label("Field root")
	}, 400, 100)
	label, editor := bounds(t, tt, "Name"), bounds(t, tt, "Editor")
	if math.Abs(float64(label.Y-editor.Y)) > 1 {
		t.Fatalf("horizontal text does not align: label=%+v editor=%+v", label, editor)
	}
}

func TestFormFieldHeaderAndBodyStyles(t *testing.T) {
	value := "Value"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Name", UI: nui.Styles{nui.Header: {PaddingX: nui.Set(float32(12))}, nui.Body: {PaddingX: nui.Set(float32(16))}}}, func(ft nui.Theme) {
			nui.Input(c, ft, &value, nui.InputOptions{Label: "Editor"}).Width(200).Label("Editor shell")
		}).Width(300).Label("Field root")
	}, 360, 120)
	root, label, shell := bounds(t, tt, "Field root"), bounds(t, tt, "Name"), bounds(t, tt, "Editor shell")
	if label.X-root.X != 12 || shell.X-root.X != 16 {
		t.Fatalf("header/body styles ignored: field=%+v label=%+v editor=%+v", root, label, shell)
	}
}

func TestFormFieldChangesPreserveEditingAndFirstControl(t *testing.T) {
	first, second := "Original first", "Original second"
	showLabel, horizontal, hideFirst, disabled := false, false, false, false
	errorText := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.FormFieldOptions{Help: "Supporting help", Error: errorText, Disabled: disabled}
		if showLabel {
			opts.Label, opts.Description, opts.Required = "Field name", "Field description", true
		}
		if horizontal {
			opts.Orientation = nui.Set(nui.Horizontal)
		}
		nui.FormField(c, th, opts, func(ft nui.Theme) {
			if !hideFirst {
				nui.Input(c.Key("first"), ft, &first, nui.InputOptions{Label: "First editor"}).Width(180)
			}
			nui.Input(c.Key("second"), ft, &second, nui.InputOptions{Label: "Second editor"}).Width(180)
		}).Width(380).Label("Field root")
	}, 420, 240)
	click(t, tt, "First editor")
	tt.Key(ui.Ctrl, ui.KeyA)
	showLabel, horizontal, errorText = true, true, "Field error"
	tt.Frame()
	if !tt.Focused("First editor") || tt.HasText("Supporting help") || !tt.HasText("Field error") {
		t.Fatal("field changes lost focus or failed error precedence")
	}
	tt.Type("Replacement")
	if first != "Replacement" {
		t.Fatal("field changes discarded the selection")
	}
	click(t, tt, "Second editor")
	click(t, tt, "Field name")
	if !tt.Focused("First editor") {
		t.Fatal("field label did not focus the first control")
	}
	hideFirst, errorText = true, ""
	tt.Frame()
	click(t, tt, "Field name")
	if !tt.Focused("Second editor") || !tt.HasText("Supporting help") || tt.HasText("Field error") {
		t.Fatal("label target or help was not restored when the first editor disappeared")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Field name")
	tt.Type("Blocked")
	if second != "Original second" {
		t.Fatal("disabled field allowed label activation or editing")
	}
}

func TestFormFieldRequiredDoesNotAddValidation(t *testing.T) {
	state := nui.FormState{}
	value, submits := "", 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Form(c, th, &state, nui.FormOptions{OnSubmit: func(nui.FormValues) { submits++ }}, func(ft nui.Theme) {
			nui.FormField(c.Key("field"), ft, nui.FormFieldOptions{Name: "name", Value: value, Label: "Name", Required: true}, func(field nui.Theme) {
				nui.Input(c, field, &value, nui.InputOptions{})
			})
			nui.Button(c, ft, "Submit", nui.ButtonOptions{}).OnClick(state.Submit)
		})
	}, 320, 200)
	click(t, tt, "Submit")
	if submits != 1 || state.Field("name").Error != "" {
		t.Fatal("required display added an implicit validator")
	}
}
