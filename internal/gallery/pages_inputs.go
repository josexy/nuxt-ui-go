package gallery

import (
	"fmt"
	"maps"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
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
		accessoryText, showAvatar, autoGrow := "First line\nSecond line\nThird line\nFourth line", false, false
		changes := 0
		add("edit", "Multiline editing", "Enter inserts a line. Chinese input follows the native platform's input method support.", func(c *ui.Context, t nui.Theme) {
			nui.Textarea(c, t, &body, nui.TextareaOptions{Label: "Project description", Rows: 3}).FillWidth()
			ui.Textf(c, "Characters: %d", len([]rune(body)))
		})
		add("height", "Fixed rows and automatic growth", "The second editor grows from two to six rows.", func(c *ui.Context, t nui.Theme) {
			nui.Textarea(c.Key("fixed"), t, &fixed, nui.TextareaOptions{Label: "Fixed rows", Rows: 3}).FillWidth()
			nui.Textarea(c.Key("auto"), t, &auto, nui.TextareaOptions{Label: "Auto resize", Rows: 2, MaxRows: 6, AutoResize: true}).FillWidth()
		})
		add("accessories", "Avatar, loading and row changes", "Loading remains editable. The avatar and loader align with the first line, and changing row policy preserves selection.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c.Key("avatar"), t, "Toggle textarea avatar", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { showAvatar = !showAvatar })
				nui.Button(c.Key("growth"), t, "Toggle textarea growth", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { autoGrow = !autoGrow })
			})
			var avatar *nui.AvatarOptions
			if showAvatar {
				avatar = &nui.AvatarOptions{Alt: "Textarea account avatar", Text: "AL"}
			}
			nui.Textarea(c.Key("notes"), t, &accessoryText, nui.TextareaOptions{Label: "Accessory notes", Avatar: avatar, Loading: true, LoadingTrailing: true, Clearable: true, ClearLabel: "Clear accessory notes", OnChange: func(string) { changes++ }, UI: nui.Styles{nui.Trailing: {Foreground: nui.Set(t.Error)}}, Rows: 2, AutoResize: autoGrow, MaxRows: 4}).FillWidth()
			ui.Textf(c, "Notes changes: %d", changes).Role(ui.RoleStatus)
		})
		add("states", "Read-only, disabled and error", "Each field has its own value.", func(c *ui.Context, t nui.Theme) {
			nui.Textarea(c, t, &readonly, nui.TextareaOptions{Label: "Read-only notes", ReadOnly: true, Rows: 2}).FillWidth()
			nui.Textarea(c, t, &disabled, nui.TextareaOptions{Label: "Disabled notes", Disabled: true, Rows: 2}).FillWidth()
			nui.Textarea(c, t, &invalid, nui.TextareaOptions{Label: "Invalid notes", Invalid: true, ErrorMessage: "Add more detail", Rows: 2}).FillWidth()
		})
	case "form-field":
		name, email, locked := "Desktop team", "team@", "Locked workspace"
		draft, showLabel, horizontal, showError := "Keep this selection", false, false, false
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
		add("layout", "Optional heading and field regions", "Toggle the heading, direction and error without losing the draft. Header and Body styles apply to their separate regions.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c.Key("label"), t, "Toggle field label", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { showLabel = !showLabel })
				nui.Button(c.Key("direction"), t, "Toggle field direction", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { horizontal = !horizontal })
				nui.Button(c.Key("error"), t, "Toggle field error", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { showError = !showError })
			})
			opts := nui.FormFieldOptions{Help: "State field help", Required: true, UI: nui.Styles{nui.Header: {PaddingX: nui.Set(float32(8)), Background: nui.Set(t.MutedBackground)}, nui.Body: {PaddingX: nui.Set(float32(8)), Background: nui.Set(t.MutedBackground)}}}
			if showLabel {
				opts.Label = "State field"
			}
			if horizontal {
				opts.Orientation = nui.Set(nui.Horizontal)
			}
			if showError {
				opts.Error = "State field error"
			}
			nui.FormField(c.Key("field"), t, opts, func(ft nui.Theme) {
				nui.Input(c.Key("draft"), ft, &draft, nui.InputOptions{Label: "Field draft"}).Width(240)
			}).FillWidth().Label("Field composition")
			ui.Text(c, "Field value: "+draft).Role(ui.RoleStatus)
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
		draft, result, reversed, vertical := "Group workspace", "Not checked", false, false
		add("dynamic", "Stable items and direction", "Reorder the items or switch direction while editing; item IDs preserve the input and its selection.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c.Key("order"), t, "Reorder field group", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { reversed = !reversed })
				nui.Button(c.Key("direction"), t, "Rotate field group", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { vertical = !vertical })
			})
			items := []nui.FieldGroupItem{{ID: "draft", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Input(c, t, &draft, nui.InputOptions{Label: "Group draft"}).Grow(1).MinWidth(0)
			}}, {ID: "action", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Button(c, t, "Check group draft", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { result = "Checked group: " + draft })
			}}}
			if reversed {
				items[0], items[1] = items[1], items[0]
			}
			nui.FieldGroup(c.Key("group"), t, nui.FieldGroupOptions{Vertical: vertical}, items).FillWidth()
			ui.Text(c, result).Role(ui.RoleStatus)
		})
	case "checkbox":
		checked, changes, tri := false, 0, nui.Indeterminate
		locked, loading := true, false
		customMixed := false
		hiddenValues := make([]bool, 3)
		add("boolean", "Controlled checkbox", "Space toggles the value. Each change invokes the callback once.", func(c *ui.Context, t nui.Theme) {
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Permission", Required: true, Help: "Applies to this workspace"}, func(field nui.Theme) {
				nui.Checkbox(c, field, &checked, nui.CheckboxOptions{Label: "Accept terms", Description: "Required before continuing", OnChange: func(bool) { changes++ }})
			})
			ui.Textf(c, "Checked: %t · Changes: %d", checked, changes)
		})
		add("tri-state", "Three states and custom icons", "Choose a custom mixed icon, then reset to the mixed state to compare it with the default minus.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("custom-icon"), t, &customMixed, nui.SwitchOptions{Label: "Custom mixed icon"})
			var mixedIcon *ui.SVG
			if customMixed {
				mixedIcon = lucide.Plus()
			}
			nui.CheckboxTri(c.Key("tri"), t, &tri, nui.CheckboxTriOptions{Label: "Select all files", Description: "Mixed selection", IndeterminateIcon: mixedIcon})
			nui.Button(c.Key("reset-mixed"), t, "Reset mixed state", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { tri = nui.Indeterminate })
			ui.Textf(c, "Check state: %d", tri)
		})
		add("hidden-icons", "Hidden indicators and label icons", "An explicit icon stays above the label in both states. Clearing the icon does not insert a default check mark.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Checkbox(c.Key("with-icon"), t, &hiddenValues[0], nui.CheckboxOptions{Label: "Workspace alerts", Description: "Explicit icon", Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden, CheckedIcon: lucide.Heart()}).Width(200)
				nui.Checkbox(c.Key("without-icon"), t, &hiddenValues[1], nui.CheckboxOptions{Label: "No automatic icon", Description: "Text only", Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden}).Width(200)
				nui.Checkbox(c.Key("icon-only"), t, &hiddenValues[2], nui.CheckboxOptions{Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden, CheckedIcon: lucide.Bell()}).Label("Icon only choice").Width(80)
			})
			ui.Textf(c, "Hidden choices: %v", hiddenValues)
		})
		add("states", "Disabled and shared options", "Disabled blocks input. Loading belongs to the shared Switch options and has no effect on Checkbox.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c, t, &locked, nui.CheckboxOptions{Label: "Locked checkbox", Disabled: true, Description: "Managed by policy"})
			nui.Checkbox(c, t, &loading, nui.CheckboxOptions{Label: "Checkbox ignores Loading", Loading: true, LoadingIcon: lucide.LoaderCircle()})
			ui.Textf(c, "Disabled: %t · Shared loading value: %t", locked, loading)
		})
	case "switch":
		enabled, changes := true, 0
		values := make([]bool, 5)
		disabled, loading := true, false
		busy, loadingChanges := true, 0
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
		add("states", "Disabled and loading", "Loading blocks the switch and its field label. Finish loading to restore input; changing loading itself does not change the selected value.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c, t, &disabled, nui.SwitchOptions{Label: "Disabled switch", Disabled: true})
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Background sync"}, func(field nui.Theme) {
				nui.Switch(c, field, &loading, nui.SwitchOptions{Label: "Loading switch", Loading: busy, LoadingIcon: lucide.RefreshCw(), CheckedIcon: lucide.Check(), UncheckedIcon: lucide.X(), OnChange: func(bool) { loadingChanges++ }})
			})
			nui.Button(c, t, "Toggle loading state", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { busy = !busy })
			ui.Textf(c, "Disabled: %t · Loading: %t", disabled, loading)
			ui.Textf(c, "Busy: %t · Changes: %d", busy, loadingChanges)
		})
	case "checkbox-group", "radio-group":
		options := []nui.Option[string]{{Value: "build", Label: "Build", Description: "Compile projects"}, {Value: "test", Label: "Test", Description: "Run checks"}, {Value: "deploy", Label: "Deploy", Description: "Unavailable for this workspace", Disabled: true}}
		if id == "checkbox-group" {
			options = append(options, nui.Option[string]{Value: "preview", Label: "Preview", Description: "Review local changes"})
		}
		for _, sample := range []struct {
			id, title  string
			horizontal bool
			variant    nui.ChoiceVariant
			indicator  nui.IndicatorPlacement
		}{{"selection", "Selection and result", false, nui.ChoiceList, nui.IndicatorStart}, {"horizontal", "Horizontal cards", true, nui.ChoiceCard, nui.IndicatorEnd}, {"table", "Joined choices with limit", false, nui.ChoiceTable, nui.IndicatorHidden}} {
			selected := []string{"build"}
			single := "build"
			changes := 0
			title := sample.title
			description := "Use arrows and Space. Select up to two of Build, Test and Preview; a third selection is blocked until one is cleared. Deploy is disabled."
			if id == "radio-group" {
				description = "Use arrows or Space to choose Build or Test. Deploy is disabled."
				if sample.id == "table" {
					title = "Joined single choice"
				}
			}
			add(sample.id, title, description, func(c *ui.Context, t nui.Theme) {
				o := nui.ChoiceGroupOptions{Label: title, Description: "Independent scenario selection", Horizontal: sample.horizontal, Variant: sample.variant, Indicator: sample.indicator, Highlight: true, Max: nui.Set(2), OnChange: func() { changes++ }}
				if id == "checkbox-group" {
					nui.CheckboxGroup(c, t, &selected, options, o)
					ui.Textf(c, "Selected IDs: %v · Changes: %d", selected, changes)
				} else {
					nui.RadioGroup(c, t, &single, options, o)
					ui.Textf(c, "Selected ID: %s · Changes: %d", single, changes)
				}
			})
		}
		iconItems := []nui.Option[string]{{Value: "build", Label: "Build task", Icon: lucide.Hammer()}, {Value: "test", Label: "Run tests", Icon: lucide.FlaskConical()}, {Value: "deploy", Label: "Deploy task", Icon: lucide.Rocket(), Disabled: true}}
		withIcons, loop := true, id == "radio-group"
		iconSelection, iconSingle, iconChanges := []string{"build"}, "build", 0
		add("icons", "Icons and keyboard navigation", "Each item keeps its value and focus when reordered. Checkbox arrows move focus; radio arrows also select. Loop controls wrapping at either end.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("icons"), t, &withIcons, nui.SwitchOptions{Label: "Show choice icons"})
			nui.Switch(c.Key("loop"), t, &loop, nui.SwitchOptions{Label: "Loop choices"})
			nui.Button(c.Key("reverse"), t, "Reverse choices", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
				for i, j := 0, len(iconItems)-1; i < j; i, j = i+1, j-1 {
					iconItems[i], iconItems[j] = iconItems[j], iconItems[i]
				}
			})
			items := make([]nui.Option[string], len(iconItems))
			copy(items, iconItems)
			if !withIcons {
				for i := range items {
					items[i].Icon = nil
				}
			}
			nui.FormField(c, t, nui.FormFieldOptions{Label: "Task choices", Help: "Deploy is disabled"}, func(field nui.Theme) {
				o := nui.ChoiceGroupOptions{Label: "Task actions", Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden, Horizontal: true, Loop: nui.Set(loop), OnChange: func() { iconChanges++ }}
				if id == "checkbox-group" {
					nui.CheckboxGroup(c.Key("choices"), field, &iconSelection, items, o)
					ui.Textf(c, "Icon choices: %v · Changes: %d", iconSelection, iconChanges)
				} else {
					nui.RadioGroup(c.Key("choices"), field, &iconSingle, items, o)
					ui.Textf(c, "Icon choice: %s · Changes: %d", iconSingle, iconChanges)
				}
			})
		})
	case "slider":
		value, lo, hi, vertical, disabled := float64(40), float64(20), float64(70), float64(30), float64(50)
		committed := "Not committed"
		inverted := false
		rangeDirection := nui.Horizontal
		decimal, decimalLow, decimalHigh := .2, .2, .8
		changes, commits, rangeChanges, rangeCommits := 0, 0, 0, 0
		customNeutral, customParts, fieldError, padded, parentDisabled := true, false, false, false, false
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
		add("parity", "Decimal steps and shared styling", "Step by 0.1, compare neutral and part colors, and inspect field errors. Padding moves the single slider's track and tooltip together.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(12).Children(func() {
				nui.Checkbox(c.Key("neutral"), t, &customNeutral, nui.CheckboxOptions{Label: "Custom slider neutral"})
				nui.Checkbox(c.Key("parts"), t, &customParts, nui.CheckboxOptions{Label: "Style slider parts"})
				nui.Checkbox(c.Key("error"), t, &fieldError, nui.CheckboxOptions{Label: "Mark sliders invalid"})
				nui.Checkbox(c.Key("padding"), t, &padded, nui.CheckboxOptions{Label: "Pad slider track"})
				nui.Checkbox(c.Key("disabled"), t, &parentDisabled, nui.CheckboxOptions{Label: "Disable slider parent"})
			})
			if customNeutral {
				t.InvertedBackground = ui.Hex("#7e22ce")
				if t.Dark {
					t.InvertedBackground = ui.Hex("#c084fc")
				}
			}
			parts, singleParts := nui.Styles{}, nui.Styles{}
			if customParts {
				parts[nui.Indicator] = nui.PartStyle{Background: nui.Set(t.Info)}
				parts[nui.Thumb] = nui.PartStyle{Background: nui.Set(t.Elevated)}
				maps.Copy(singleParts, parts)
			}
			if padded {
				singleParts[nui.Root] = nui.PartStyle{PaddingX: nui.Set(float32(48))}
			}
			errorText := ""
			if fieldError {
				errorText = "Choose another slider value"
			}
			ui.Column(c).Disabled(parentDisabled).Gap(10).Children(func() {
				nui.FormField(c.Key("single"), t, nui.FormFieldOptions{Error: errorText}, func(ft nui.Theme) {
					nui.Slider(c, ft, &decimal, nui.SliderOptions{Label: "Decimal slider", Max: nui.Set(1.0), Step: .1, Color: nui.Neutral, Tooltip: true, OnChange: func(float64) { changes++ }, OnCommit: func(float64) { commits++ }, UI: singleParts})
				})
				ui.Textf(c, "Decimal: %g · Changes: %d · Commits: %d", decimal, changes, commits).Role(ui.RoleStatus)
				nui.FormField(c.Key("range"), t, nui.FormFieldOptions{Error: errorText}, func(ft nui.Theme) {
					nui.RangeSlider(c, ft, &decimalLow, &decimalHigh, nui.RangeSliderOptions{Label: "Decimal range", Max: 1, Step: .1, MinStepsBetweenThumbs: 1, Color: nui.Neutral, Tooltip: true, OnChange: func(float64, float64) { rangeChanges++ }, OnCommit: func(float64, float64) { rangeCommits++ }, UI: parts}).FillWidth()
				})
				ui.Textf(c, "Decimal range: %g – %g · Changes: %d · Commits: %d", decimalLow, decimalHigh, rangeChanges, rangeCommits).Role(ui.RoleStatus)
			})
		})
	case "select":
		basic, duplicate, unselected, empty := "personal", "team-eu", "", ""
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
			ui.Text(c, "Available options, no initial selection").TextColor(t.Muted)
			nui.Select(c, t, &unselected, plans, nui.SelectOptions{Label: "Unselected plan", Placeholder: "Choose a plan"}).FillWidth()
			ui.Text(c, "No available options").TextColor(t.Muted)
			nui.Select(c, t, &empty, nil, nui.SelectOptions{Label: "Empty select", Placeholder: "No options available"}).FillWidth()
		})
		accessory, accessoryMany := "alpha", []string{"alpha"}
		loading, trailing, highlight, customIcons, parentDisabled := false, true, true, false, false
		open, manyOpen, changes, manyChanges := false, false, 0, 0
		add("accessories", "Trigger content and controlled opening", "Loading remains selectable. Type a prefix in the single trigger; in an open list, type a prefix and press Enter. Escape restores trigger focus. Description icons align with the first line.", func(c *ui.Context, t nui.Theme) {
			t.InvertedBackground = ui.Hex("#7e22ce")
			if t.Dark {
				t.InvertedBackground = ui.Hex("#c084fc")
			}
			nui.Switch(c.Key("loading"), t, &loading, nui.SwitchOptions{Label: "Load select choices"})
			nui.Switch(c.Key("trailing"), t, &trailing, nui.SwitchOptions{Label: "Select trailing loader"})
			nui.Switch(c.Key("highlight"), t, &highlight, nui.SwitchOptions{Label: "Highlight select trigger"})
			nui.Switch(c.Key("icons"), t, &customIcons, nui.SwitchOptions{Label: "Customize select icons"})
			nui.Switch(c.Key("disabled"), t, &parentDisabled, nui.SwitchOptions{Label: "Disable select parent"})
			items := []nui.Option[string]{{Value: "alpha", Label: "Alpha team", Description: "Design and research", Icon: lucide.Users()}, {Value: "blocked", Label: "Blocked team", Description: "Unavailable", Disabled: true}, {Value: "beta", Label: "Beta team", Description: "Application development", Icon: lucide.Code()}}
			o := nui.SelectOptions{Color: nui.Neutral, Highlight: highlight, Avatar: &nui.AvatarOptions{Text: "AL", Alt: "Select team avatar"}, Loading: loading, LoadingTrailing: trailing}
			if customIcons {
				o.Leading, o.Trailing, o.SelectedIcon = lucide.Users(), lucide.ChevronsUpDown(), lucide.Heart()
			}
			ui.Column(c.Key("accessory-fields")).FillWidth().Gap(8).Disabled(parentDisabled).Children(func() {
				o.Label, o.Open, o.OnChange = "Accessory team", &open, func() { changes++ }
				nui.Select(c.Key("one"), t, &accessory, items, o).FillWidth()
				ui.Textf(c, "Accessory: %s · Changes: %d · Open: %v", accessory, changes, open).Role(ui.RoleStatus)
				o.Label, o.Open, o.OnChange = "Accessory teams", &manyOpen, func() { manyChanges++ }
				nui.SelectMultiple(c.Key("many"), t, &accessoryMany, items, o).FillWidth()
				ui.Textf(c, "Accessories: %v · Changes: %d · Open: %v", accessoryMany, manyChanges, manyOpen).Role(ui.RoleStatus)
			})
			nui.Button(c.Key("open-one"), t, "Toggle single select", nui.ButtonOptions{Variant: nui.Outline, Disabled: parentDisabled}).OnClick(func() { open = !open })
			nui.Button(c.Key("open-many"), t, "Toggle multiple select", nui.ButtonOptions{Variant: nui.Outline, Disabled: parentDisabled}).OnClick(func() { manyOpen = !manyOpen })
		})
	case "input-menu":
		basic, multiple, empty := nui.SearchState{}, nui.SearchState{}, nui.SearchState{}
		external := false
		result := "No create request"
		locked, readonly, loading := nui.SearchState{}, nui.SearchState{}, nui.SearchState{}
		locked.Query, locked.Selected = "Personal", "personal"
		readonly.Query, readonly.Selected = "Team", "team"
		add("filter", "Editing and filtering", "Type a name to filter. Escape closes the results without leaving the field.", func(c *ui.Context, t nui.Theme) {
			nui.InputMenu(c, t, &basic, plans, nui.InputMenuOptions{Label: "Search workspaces", Placeholder: "Type Personal or Team"}).FillWidth()
			ui.Text(c, "Selected ID: "+basic.Selected)
		})
		add("multiple", "Multiple values and creation", "External filtering leaves supplied results unchanged; creation is a local callback.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c, t, &external, nui.SwitchOptions{Label: "Use external filtering"})
			nui.InputMenu(c, t, &multiple, propsOptions, nui.InputMenuOptions{Label: "Team search", Placeholder: "Search teams or create one", Multiple: true, ExternalFilter: external, OnCreate: func(v string) { result = "Create requested: " + v }}).FillWidth()
			ui.Text(c, result)
		})
		add("states", "Empty, disabled, read-only and loading", "The empty search has no options. Disabled Personal and read-only Team retain their values. Loading shows an indicator without starting a background request; each field owns its search state.", func(c *ui.Context, t nui.Theme) {
			nui.InputMenu(c, t, &empty, nil, nui.InputMenuOptions{Label: "Empty search", EmptyText: "No matching teams"}).FillWidth()
			nui.InputMenu(c, t, &locked, plans, nui.InputMenuOptions{Label: "Disabled search", Disabled: true}).FillWidth()
			nui.InputMenu(c, t, &readonly, plans, nui.InputMenuOptions{Label: "Read-only search", ReadOnly: true}).FillWidth()
			nui.InputMenu(c, t, &loading, plans, nui.InputMenuOptions{Label: "Loading search", Loading: true}).FillWidth()
		})
		policy := nui.SearchState{SelectedValues: []string{"alpha", "beta"}}
		policyOpen, resetModel, policyReadonly, policyLoading, loop, hover, autocomplete := false, true, false, false, false, true, false
		queries, selections, valueChanges, clears, submits := 0, 0, 0, 0, 0
		add("policies", "Clear, navigation and autocomplete", "Clicking the editor only focuses it by default. Type or use the arrow to open. Clear can retain selected IDs; loading remains editable. Autocomplete ignores Multiple and selection-reset settings.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("reset"), t, &resetModel, nui.SwitchOptions{Label: "Clear resets selected IDs"})
			nui.Switch(c.Key("readonly"), t, &policyReadonly, nui.SwitchOptions{Label: "InputMenu read only"})
			nui.Switch(c.Key("loading"), t, &policyLoading, nui.SwitchOptions{Label: "InputMenu loading"})
			nui.Switch(c.Key("loop"), t, &loop, nui.SwitchOptions{Label: "Loop input menu"})
			nui.Switch(c.Key("hover"), t, &hover, nui.SwitchOptions{Label: "Hover input menu items"})
			nui.Switch(c.Key("mode"), t, &autocomplete, nui.SwitchOptions{Label: "InputMenu autocomplete"})
			mode := nui.InputMenuCombobox
			if autocomplete {
				mode = nui.InputMenuAutocomplete
			}
			nui.InputMenu(c.Key("policy"), t, &policy, []nui.Option[string]{{Value: "alpha", Label: "Alpha", Description: "Design and research", Icon: lucide.Users()}, {Value: "blocked", Label: "Blocked", Disabled: true}, {Value: "beta", Label: "Beta", Description: "Application development", Icon: lucide.Code()}}, nui.InputMenuOptions{Label: "Policy search", Clearable: true, ReadOnly: policyReadonly, Loading: policyLoading, LoadingTrailing: true, Avatar: &nui.AvatarOptions{Text: "AL", Alt: "Search team avatar"}, OnChange: func(string) { queries++ }, OnClear: func() { clears++ }, OnSubmit: func(string) { submits++ }, Multiple: true, Mode: mode, Open: &policyOpen, Loop: loop, HighlightOnHover: nui.Set(hover), ResetModelValueOnClear: nui.Set(resetModel), ResetSearchTermOnSelect: nui.Set(!autocomplete), SelectedIcon: lucide.Heart(), DeleteIcon: lucide.Minus(), ClearIcon: lucide.Eraser(), OnSelect: func(string) { selections++ }, OnValuesChange: func([]string) { valueChanges++ }}).FillWidth()
			ui.Textf(c, "Policy query: %s · ID: %s · IDs: %v", policy.Query, policy.Selected, policy.SelectedValues).Role(ui.RoleStatus)
			ui.Textf(c, "Policy events: query %d · select %d · values %d · clear %d · submit %d", queries, selections, valueChanges, clears, submits).Role(ui.RoleStatus)
			nui.Button(c.Key("open"), t, "Toggle policy search", nui.ButtonOptions{Variant: nui.Outline, Disabled: policyReadonly}).OnClick(func() { policyOpen = !policyOpen })
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
		controlledOpen, alternate, styled := false, false, false
		draft := "Keep this draft"
		add("controlled", "Current anchor and panel styles", "Move the open panel to the other anchor while editing. Closing returns to that anchor. Panel typography and spacing apply to plain content without replacing the local native theme.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("style"), t, &styled, nui.SwitchOptions{Label: "Style controlled popup"})
			var first, second ui.Element
			ui.Row(c.Key("anchors")).Gap(12).Wrap().Children(func() {
				first = nui.Button(c.Key("first"), t, "Open controlled popup", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { controlledOpen = !controlledOpen })
				second = nui.Button(c.Key("second"), t, "Alternate popup anchor", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { controlledOpen = !controlledOpen })
			})
			anchor := first
			if alternate {
				anchor = second
			}
			parts := nui.Styles{}
			if styled {
				parts[nui.Panel] = nui.PartStyle{Background: nui.Set(t.Elevated), Foreground: nui.Set(t.Info), FontSize: nui.Set(float32(18)), LineHeight: nui.Set(float32(1.5)), Gap: nui.Set(float32(12)), PaddingX: nui.Set(float32(12)), PaddingY: nui.Set(float32(12))}
			}
			nui.Popover(c.Key("panel"), t, anchor, &controlledOpen, nui.PopoverOptions{Label: "Controlled popup panel", Width: nui.Set(float32(280)), UI: parts}, func(inner nui.Theme) {
				ui.Text(c, "Panel heading")
				nui.Input(c.Key("draft"), inner, &draft, nui.InputOptions{Label: "Controlled popup draft"}).FillWidth()
				nui.Button(c.Key("move"), inner, "Move popup anchor", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { alternate = !alternate })
				nui.Button(c.Key("close"), inner, "Close controlled popup", nui.ButtonOptions{}).OnClick(func() { controlledOpen = false })
			})
			ui.Text(c, "Controlled popup value: "+draft).Role(ui.RoleStatus)
		})
		p.Deactivate = append(p.Deactivate, func() { controlledOpen = false })
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
		var position ui.ScrollState
		wide, hidden := false, false
		direction := nui.ScrollBoth
		add("state", "Rounded viewport and retained offsets", "Scroll in both directions, change thumb styling, or set the stored position. Changing direction preserves the available axis and clamps the other. Root corners clip the content.", func(c *ui.Context, t nui.Theme) {
			nui.Select(c.Key("direction"), t, &direction, []nui.Option[nui.ScrollDirection]{{Value: nui.ScrollVertical, Label: "Vertical"}, {Value: nui.ScrollHorizontal, Label: "Horizontal"}, {Value: nui.ScrollBoth, Label: "Both axes"}}, nui.SelectOptions{Label: "Scroll direction"})
			nui.Switch(c.Key("width"), t, &wide, nui.SwitchOptions{Label: "Wide scroll thumb"})
			nui.Switch(c.Key("color"), t, &hidden, nui.SwitchOptions{Label: "Transparent scroll thumb"})
			width := float32(6)
			if wide {
				width = 14
			}
			color := t.Info
			if hidden {
				color = ui.Transparent
			}
			nui.ScrollArea(c.Key("area"), t, nui.ScrollAreaOptions{Direction: direction, State: &position, ScrollbarWidth: nui.Set(width), ScrollbarColor: nui.Set(color), UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(16))}}, Style: func(part nui.Slot, e ui.Element) {
				if part == nui.Body {
					e.Label("State scroll body")
				}
			}}, func() {
				for i := range 24 {
					row := ui.Textf(c.Key(i), "Record %02d · Wide scrollable content", i+1).PaddingX(12).Background(t.Elevated)
					if direction == nui.ScrollHorizontal {
						row.Width(180).Height(120)
					} else {
						row.Width(900).Height(40)
					}
				}
			}).Height(180).FillWidth()
			ui.Textf(c, "Scroll position: x=%.0f y=%.0f", position.X, position.Y).Role(ui.RoleStatus)
			nui.Button(c.Key("end"), t, "Scroll to end", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { position.X, position.Y = position.MaxX, position.MaxY })
			nui.Button(c.Key("start"), t, "Scroll to start", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { position.X, position.Y = 0, 0 })
		})
	}
	return p
}
