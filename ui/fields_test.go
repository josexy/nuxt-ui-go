package nui_test

import (
	"math"
	"testing"
	"time"

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

func TestFieldGroupReorderOrientationAndDisabledPreserveDraft(t *testing.T) {
	value, reversed, vertical, disabled := "Original", false, false, false
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		items := []nui.FieldGroupItem{{ID: "draft", Build: func(c *ui.Context, t nui.Theme) ui.Element {
			return nui.Input(c, t, &value, nui.InputOptions{Label: "Group draft"}).Width(160)
		}}, {ID: "action", Build: func(c *ui.Context, t nui.Theme) ui.Element {
			return nui.Button(c, t, "Apply", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { calls++ })
		}}}
		if reversed {
			items[0], items[1] = items[1], items[0]
		}
		nui.FieldGroup(c.Key("group"), th, nui.FieldGroupOptions{Vertical: vertical, Disabled: disabled}, items).Width(240)
	}, 300, 160)
	click(t, tt, "Group draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	reversed, vertical = true, true
	tt.Frame()
	if !tt.Focused("Group draft") || bounds(t, tt, "Apply").Y >= bounds(t, tt, "Group draft").Y {
		t.Fatal("reorder or orientation did not preserve focus and item order")
	}
	tt.Type("Replacement")
	if value != "Replacement" {
		t.Fatal("group layout changes discarded the selection")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Apply")
	click(t, tt, "Group draft")
	tt.Type("Blocked")
	if calls != 0 || value != "Replacement" {
		t.Fatal("group disabled state did not reach its children")
	}
	disabled = false
	tt.Frame()
	click(t, tt, "Apply")
	tt.Key(0, ui.KeyEnter)
	if calls != 2 {
		t.Fatal("group action did not resume after enabling")
	}
}

func TestScrollAreaRoundedRootClipsContent(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).Children(func() {
			nui.ScrollArea(c, th, nui.ScrollAreaOptions{UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(16))}}}, func() { ui.Box(c).Width(200).Height(400).Background(ui.RGB(255, 0, 0)) }).Width(200).Height(120).Label("Rounded scroll")
		})
	}, 260, 170)
	r := bounds(t, tt, "Rounded scroll")
	if tt.Image().RGBAAt(int(r.X+1), int(r.Y+1)) == rgba(ui.RGB(255, 0, 0)) {
		t.Fatal("scroll content escaped Root radius")
	}
	if tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+6)) != rgba(ui.RGB(255, 0, 0)) {
		t.Fatal("rounded clipping removed visible content")
	}
}

func TestScrollAreaAxesAndStableState(t *testing.T) {
	state := ui.ScrollState{}
	direction := nui.ScrollBoth
	width := float32(6)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ScrollArea(c.Key("area"), nui.App(c, nui.AppConfig{}), nui.ScrollAreaOptions{Direction: direction, State: &state, ScrollbarWidth: nui.Set(width)}, func() { ui.Box(c.Key("canvas")).Width(800).Height(800) }).Width(240).Height(160).Label("Area")
	}, 320, 240)
	r := bounds(t, tt, "Area")
	tt.Scroll(r.X+80, r.Y+80, 100, 120)
	tt.Advance(time.Second)
	if state.X <= 0 || state.Y <= 0 || state.MaxX <= 0 || state.MaxY <= 0 {
		t.Fatalf("both axes not scrollable: %+v", state)
	}
	x, y := state.X, state.Y
	width = 14
	tt.Frame()
	if state.X != x || state.Y != y {
		t.Fatal("scrollbar styling reset offsets")
	}
	direction = nui.ScrollHorizontal
	tt.Frame()
	if state.X != x || state.Y != 0 {
		t.Fatalf("horizontal switch=%+v", state)
	}
	direction = nui.ScrollBoth
	state.X = 9000
	state.Y = 9000
	tt.Frame()
	if state.X != state.MaxX || state.Y != state.MaxY {
		t.Fatal("programmatic end was not clamped")
	}
	state.X, state.Y = 0, 0
	tt.Frame()
	if state.X != 0 || state.Y != 0 {
		t.Fatal("programmatic start failed")
	}
}

func TestScrollAreaNestedBoundaryPassesWheel(t *testing.T) {
	var outer, inner ui.ScrollState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c, th, nui.ScrollAreaOptions{State: &outer}, func() {
			nui.ScrollArea(c, th, nui.ScrollAreaOptions{State: &inner}, func() { ui.Box(c).Height(500) }).Height(100).Label("Inner")
			ui.Box(c).Height(800)
		}).Width(240).Height(220)
	}, 300, 280)
	r := bounds(t, tt, "Inner")
	tt.Scroll(r.X+70, r.Y+50, 0, 60)
	tt.Advance(time.Second)
	if inner.Y <= 0 || outer.Y != 0 {
		t.Fatal("inner wheel leaked to outer before its boundary")
	}
	inner.Y = inner.MaxY
	tt.Frame()
	tt.Scroll(r.X+70, r.Y+50, 0, 60)
	tt.Advance(time.Second)
	if outer.Y <= 0 {
		t.Fatal("inner boundary swallowed wheel instead of reaching outer")
	}
}

func TestScrollAreaThumbColorWidthAndZeroFallback(t *testing.T) {
	width := float32(14)
	state := ui.ScrollState{}
	marker := ui.RGB(255, 0, 255)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ScrollArea(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode}), nui.ScrollAreaOptions{State: &state, ScrollbarColor: nui.Set(marker), ScrollbarWidth: nui.Set(width)}, func() { ui.Box(c).Height(600) }).Width(240).Height(160).Label("Area")
	}, 300, 220)
	tt.Move(220, 20)
	strip := ui.Rect{X: 210, Y: 20, W: 30, H: 1}
	if got := countColor(tt.Image(), strip, rgba(marker)); got != 14 {
		t.Fatalf("custom thumb width/color=%d", got)
	}
	width = 0
	tt.Frame()
	if got := countColor(tt.Image(), strip, rgba(marker)); got != 6 {
		t.Fatalf("zero thumb width=%d, want native 6", got)
	}
	tt.Press(234, 20)
	tt.Move(234, 120)
	tt.Release(234, 120)
	if state.Y <= 0 {
		t.Fatal("dragging zero-fallback thumb did not scroll")
	}
}

func TestScrollAreaDirectionKeepsEditorIdentity(t *testing.T) {
	draft, direction := "Original", nui.ScrollBoth
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c.Key("area"), th, nui.ScrollAreaOptions{Direction: direction}, func() {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Draft"}).Width(200)
			ui.Box(c.Key("canvas")).Width(800).Height(600)
		}).Width(260).Height(180)
	}, 340, 240)
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	direction = nui.ScrollHorizontal
	tt.Frame()
	tt.SetDirection(ui.RightToLeft)
	if !tt.Focused("Draft") {
		t.Fatal("direction change lost the editor")
	}
	tt.Type("Replaced")
	if draft != "Replaced" {
		t.Fatal("direction change lost the selected text")
	}
}
