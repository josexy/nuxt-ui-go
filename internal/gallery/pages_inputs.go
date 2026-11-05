package gallery

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func inputComponentsPage(id string) *componentPage {
	p := &componentPage{}
	add := func(id, title, description string, build func(*ui.Context, nui.Theme)) {
		p.Scenes = append(p.Scenes, scene(id, title, description, build))
	}
	switch id {
	case "textarea":
		body, fixed, auto := "Hello / 你好\nA persistent multiline draft.", "Line one\nLine two\nLine three", "Type several lines to grow this editor."
		readonly, disabled, invalid := "Read-only text\nSelectable but not editable.", "Disabled content", "Please explain the issue."
		add("edit", "Multiline editing", "Enter inserts a line. Chinese input follows the native platform's input method support.", func(c *ui.Context, t nui.Theme) {
			nui.Textarea(c, t, &body, nui.TextareaOptions{InputOptions: nui.InputOptions{Label: "Project description"}, Rows: 3}).FillWidth()
			ui.Textf(c, "Characters: %d", len([]rune(body)))
		})
		add("height", "Fixed rows and automatic growth", "The second editor grows from two to six rows.", func(c *ui.Context, t nui.Theme) {
			nui.Textarea(c.Key("fixed"), t, &fixed, nui.TextareaOptions{InputOptions: nui.InputOptions{Label: "Fixed rows"}, Rows: 3}).FillWidth()
			nui.Textarea(c.Key("auto"), t, &auto, nui.TextareaOptions{InputOptions: nui.InputOptions{Label: "Auto resize"}, Rows: 2, MaxRows: 6, AutoResize: true}).FillWidth()
		})
		add("states", "Read-only, disabled and error", "Each field has its own value.", func(c *ui.Context, t nui.Theme) {
			nui.Textarea(c, t, &readonly, nui.TextareaOptions{InputOptions: nui.InputOptions{Label: "Read-only notes", ReadOnly: true}, Rows: 2}).FillWidth()
			nui.Textarea(c, t, &disabled, nui.TextareaOptions{InputOptions: nui.InputOptions{Label: "Disabled notes", Disabled: true}, Rows: 2}).FillWidth()
			nui.Textarea(c, t, &invalid, nui.TextareaOptions{InputOptions: nui.InputOptions{Label: "Invalid notes", Invalid: true, ErrorMessage: "Add more detail"}, Rows: 2}).FillWidth()
		})
	case "form-field":
		name, email, locked := "Desktop team", "team@", "Locked workspace"
		add("label", "Label and supporting text", "The label focuses its editor. Required is exposed as field metadata.", func(c *ui.Context, t nui.Theme) {
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Team name", Description: "Visible to everyone", Help: "You can rename it later.", Hint: "Required", Required: true}, func(inner nui.Theme) { nui.Input(c, inner, &name, nui.InputOptions{}).FillWidth() })
		})
		add("validation", "Error and horizontal orientation", "An explicit error is associated with this field.", func(c *ui.Context, t nui.Theme) {
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Contact email", Error: "Enter a complete address", Orientation: nui.Set(nui.Horizontal)}, func(inner nui.Theme) { nui.Input(c, inner, &email, nui.InputOptions{}).FillWidth() })
		})
		add("inheritance", "Disabled and size inheritance", "The child inherits LG size and the disabled state from its field.", func(c *ui.Context, t nui.Theme) {
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Locked team", Disabled: true, Size: nui.LG}, func(inner nui.Theme) { nui.Input(c, inner, &locked, nui.InputOptions{}).FillWidth() })
			ui.Text(c, "Value: "+locked)
		})
	case "field-group":
		for _, v := range []struct {
			id, title          string
			vertical, disabled bool
			size               nui.Size
		}{{"horizontal", "Horizontal group", false, false, nui.MD}, {"vertical", "Vertical group", true, false, nui.MD}, {"inheritance", "Large disabled group", false, true, nui.LG}} {
			value, result := "Workspace", "Not submitted"
			add(v.id, v.title, "The input and action share borders and inherited size. Each group owns its draft.", func(c *ui.Context, t nui.Theme) {
				nui.FieldGroup(c, t, nui.FieldGroupOptions{Vertical: v.vertical, Disabled: v.disabled, Size: v.size}, []nui.FieldGroupItem{{ID: "input", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
					return nui.Input(c, inner, &value, nui.InputOptions{Label: v.title + " name"}).MinWidth(0)
				}}, {ID: "action", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
					return nui.Button(c, inner, "Validate "+v.id, nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { result = "Checked: " + value })
				}}}).FillWidth()
				ui.Text(c, result)
				ui.Text(c, "Value: "+value)
			})
		}
	case "checkbox":
		checked, changes, tri := false, 0, nui.Indeterminate
		locked, loading := true, false
		add("boolean", "Controlled checkbox", "Space toggles the value. Each change invokes the callback once.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c, t, &checked, nui.CheckboxOptions{Label: "Accept terms", Description: "Required before continuing", OnChange: func(bool) { changes++ }})
			ui.Textf(c, "Checked: %t · Changes: %d", checked, changes)
		})
		add("tri-state", "Three states", "The initial mixed state is distinct from checked and unchecked.", func(c *ui.Context, t nui.Theme) {
			nui.CheckboxTri(c, t, &tri, nui.CheckboxTriOptions{Label: "Select all files", Description: "Mixed selection"})
			ui.Textf(c, "Check state: %d", tri)
		})
		add("states", "Disabled and loading", "Blocked controls preserve their value and expose descriptions.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c, t, &locked, nui.CheckboxOptions{Label: "Locked checkbox", Disabled: true, Description: "Managed by policy"})
			nui.Checkbox(c, t, &loading, nui.CheckboxOptions{Label: "Loading checkbox", Loading: true, LoadingIcon: lucide.LoaderCircle()})
		})
	case "switch":
		enabled, changes := true, 0
		values := make([]bool, 5)
		disabled, loading := true, false
		add("controlled", "Controlled switch", "Toggle with Space and observe the result.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c, t, &enabled, nui.SwitchOptions{Label: "Email notifications", OnChange: func(bool) { changes++ }})
			ui.Textf(c, "Notifications: %t · Changes: %d", enabled, changes)
		})
		add("appearance", "Colors, sizes and icons", "Each switch has a separate controlled value.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for i, v := range fiveSizes {
					nui.Switch(c.Key(v.Name), t, &values[i], nui.SwitchOptions{Label: v.Name + " switch", Size: v.Value, Color: semanticColors[i].Value, CheckedIcon: lucide.Check(), UncheckedIcon: lucide.X()})
				}
			})
		})
		add("states", "Disabled and loading", "Neither state accepts input.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c, t, &disabled, nui.SwitchOptions{Label: "Disabled switch", Disabled: true})
			nui.Switch(c, t, &loading, nui.SwitchOptions{Label: "Loading switch", Loading: true, LoadingIcon: lucide.LoaderCircle()})
			ui.Textf(c, "Disabled: %t · Loading: %t", disabled, loading)
		})
	case "checkbox-group", "radio-group":
		options := []nui.Option[string]{{Value: "build", Label: "Build", Description: "Compile projects"}, {Value: "test", Label: "Test", Description: "Run checks"}, {Value: "deploy", Label: "Deploy", Description: "Unavailable for this workspace", Disabled: true}}
		for _, sample := range []struct {
			id, title  string
			horizontal bool
			variant    nui.ChoiceVariant
			indicator  nui.IndicatorPlacement
		}{{"selection", "Selection and result", false, nui.ChoiceList, nui.IndicatorStart}, {"horizontal", "Horizontal cards", true, nui.ChoiceCard, nui.IndicatorEnd}, {"table", "Joined choices with limit", false, nui.ChoiceTable, nui.IndicatorHidden}} {
			selected := []string{"build"}
			single := "build"
			changes := 0
			add(sample.id, sample.title, "Use arrows and Space. Deploy is disabled; the checkbox group allows at most two values.", func(c *ui.Context, t nui.Theme) {
				o := nui.ChoiceGroupOptions{Label: sample.title, Description: "Independent scenario selection", Horizontal: sample.horizontal, Variant: sample.variant, Indicator: sample.indicator, Highlight: true, Max: nui.Set(2), OnChange: func() { changes++ }}
				if id == "checkbox-group" {
					nui.CheckboxGroup(c, t, &selected, options, o)
					ui.Textf(c, "Selected IDs: %v · Changes: %d", selected, changes)
				} else {
					nui.RadioGroup(c, t, &single, options, o)
					ui.Textf(c, "Selected ID: %s · Changes: %d", single, changes)
				}
			})
		}
	case "slider":
		value, lo, hi, vertical, disabled := float64(40), float64(20), float64(70), float64(30), float64(50)
		committed := "Not committed"
		inverted := false
		rangeDirection := nui.Horizontal
		add("single", "Step and bounds", "Arrow keys step by 5 within 10–90. Home and End reach the bounds.", func(c *ui.Context, t nui.Theme) {
			nui.Slider(c, t, &value, nui.SliderOptions{Label: "Progress slider", Min: nui.Set(float64(10)), Max: nui.Set(float64(90)), Step: 5, Tooltip: true})
			ui.Textf(c, "Value: %.0f", value)
		})
		add("range", "Range with minimum gap", "The two thumbs stay at least two 5-unit steps apart. Commit feedback updates on release or a key step.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c, t, &rangeDirection, []nui.Option[nui.Orientation]{{Value: nui.Horizontal, Label: "horizontal"}, {Value: nui.Vertical, Label: "vertical"}}, nui.SelectOptions{Label: "Range orientation"})
			nui.Switch(c, t, &inverted, nui.SwitchOptions{Label: "Reverse range"})
			rangeElement := nui.RangeSlider(c, t, &lo, &hi, nui.RangeSliderOptions{Label: "Workspace range", Min: 0, Max: 100, Step: 5, MinStepsBetweenThumbs: 2, Inverted: inverted, Orientation: rangeDirection, Tooltip: true, OnCommit: func(l, h float64) { committed = fmt.Sprintf("Confirmed: %g – %g", l, h) }})
			if rangeDirection == nui.Vertical {
				rangeElement.Height(160)
			}
			ui.Textf(c, "%g – %g · %s", lo, hi, committed)
		})
		add("states", "Vertical and disabled", "Vertical uses Up and Down. Disabled preserves its value.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Slider(c, t, &vertical, nui.SliderOptions{Label: "Vertical progress", Vertical: true, Step: 10}).Height(140)
				nui.Slider(c, t, &disabled, nui.SliderOptions{Label: "Disabled slider", Disabled: true}).Width(200)
			})
		})
	case "select":
		basic, duplicate, empty := "personal", "team-eu", ""
		multiple := []string{"alpha"}
		options := append(append([]nui.Option[string]{}, plans...), nui.Option[string]{Value: "locked", Label: "Enterprise", Disabled: true})
		add("single", "Single selection", "Arrow keys and Enter select a value; Escape returns focus to the trigger.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c, t, &basic, plans, nui.SelectOptions{Label: "Plan"}).FillWidth()
			ui.Text(c, "Selected ID: "+basic)
		})
		add("identity", "Repeated labels and disabled options", "Both Team rows have different values. Enterprise cannot be selected.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c, t, &duplicate, options, nui.SelectOptions{Label: "Regional plan"}).FillWidth()
			ui.Text(c, "Regional ID: "+duplicate)
		})
		add("multiple", "Typed multiple selection", "The selected string IDs are independent of single selection.", func(c *ui.Context, t nui.Theme) {
			nui.SelectMultiple(c, t, &multiple, propsOptions, nui.SelectOptions{Label: "Multiple teams", Placeholder: "Select teams"}).FillWidth()
			ui.Textf(c, "Team IDs: %v", multiple)
		})
		add("empty", "Empty selection and list", "An empty value shows a placeholder; an empty list has no selectable items.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c, t, &empty, nil, nui.SelectOptions{Label: "Empty select", Placeholder: "No options available"}).FillWidth()
		})
	case "input-menu":
		basic, multiple, empty := nui.SearchState{}, nui.SearchState{}, nui.SearchState{}
		external := false
		result := "No create request"
		locked, readonly, loading := nui.SearchState{}, nui.SearchState{}, nui.SearchState{}
		add("filter", "Editing and filtering", "Type a name to filter. Escape closes the results without leaving the field.", func(c *ui.Context, t nui.Theme) {
			nui.InputMenu(c, t, &basic, plans, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Search workspaces", Placeholder: "Type Personal or Team"}}).FillWidth()
			ui.Text(c, "Selected ID: "+basic.Selected)
		})
		add("multiple", "Multiple values and creation", "External filtering leaves supplied results unchanged; creation is a local callback.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c, t, &external, nui.SwitchOptions{Label: "Use external filtering"})
			nui.InputMenu(c, t, &multiple, propsOptions, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Team search", Placeholder: "Search teams or create one"}, Multiple: true, ExternalFilter: external, OnCreate: func(v string) { result = "Create requested: " + v }}).FillWidth()
			ui.Text(c, result)
		})
		add("states", "Empty, disabled, read-only and loading", "Each field has its own search state. No background request is started.", func(c *ui.Context, t nui.Theme) {
			nui.InputMenu(c, t, &empty, nil, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Empty search"}, EmptyText: "No matching teams"}).FillWidth()
			nui.InputMenu(c, t, &locked, plans, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Disabled search", Disabled: true}}).FillWidth()
			nui.InputMenu(c, t, &readonly, plans, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Read-only search", ReadOnly: true}}).FillWidth()
			nui.InputMenu(c, t, &loading, plans, nui.InputMenuOptions{InputOptions: nui.InputOptions{Label: "Loading search", Loading: true}}).FillWidth()
		})
	case "popover":
		for _, v := range []struct {
			id, title   string
			dismissible bool
			mode        nui.PopoverMode
		}{{"click", "Click popover", true, nui.PopoverClick}, {"policy", "Explicit close only", false, nui.PopoverClick}, {"hover", "Hover and nested controls", true, nui.PopoverHover}} {
			open := false
			plan := "personal"
			add(v.id, v.title, "The trigger regains focus on close. The policy scene requires its close button; hover uses a short delay.", func(c *ui.Context, t nui.Theme) {
				anchor := nui.Button(c, t, "Open "+v.id+" popover", nui.ButtonOptions{Variant: nui.Outline})
				if v.mode == nui.PopoverClick {
					anchor.OnClick(func() { open = !open })
				}
				nui.Popover(c, t, anchor, &open, nui.PopoverOptions{Label: v.title, Mode: v.mode, Dismissible: nui.Set(v.dismissible), OpenDelay: 200 * time.Millisecond, CloseDelay: 150 * time.Millisecond, Width: nui.Set(float32(300))}, func(inner nui.Theme) {
					nui.Select(c, inner, &plan, plans, nui.SelectOptions{Label: v.id + " popover plan"}).FillWidth()
					nui.Button(c, inner, "Close "+v.id+" popover", nui.ButtonOptions{}).OnClick(func() { open = false })
				})
			})
			p.Deactivate = append(p.Deactivate, func() { open = false })
		}
	case "scroll-area":
		for _, v := range []struct {
			id, title string
			width     float32
			direction nui.ScrollDirection
		}{{"list", "Long vertical list", 6, nui.ScrollVertical}, {"wide", "Wide scrollbar", 14, nui.ScrollVertical}, {"nested", "Bounded two-axis content", 8, nui.ScrollBoth}} {
			state := &ui.ScrollState{}
			add(v.id, v.title, "Use the wheel over this area. Scrolling is scoped to this independent persistent model.", func(c *ui.Context, t nui.Theme) {
				nui.ScrollArea(c, t, nui.ScrollAreaOptions{Direction: v.direction, ScrollbarWidth: nui.Set(v.width), State: state}, func() {
					for i := range 30 {
						ui.Textf(c, "%s · Row %02d / 独立滚动区域", v.title, i+1).Width(600)
					}
				}).Height(180).FillWidth()
			})
		}
	}
	return p
}
