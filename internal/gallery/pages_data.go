package gallery

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// Simulated completions are explicit UI actions: no network or timer owns a
// window after unmount. Tickets still exercise cancellation and stale results.
type formExample struct {
	Form             nui.FormState
	Ticket           nui.FormTicket
	Name             string
	Success          int
	Requests, Errors int
	Result           string
}

func (s *formExample) finish() {
	if s.Form.CompleteAsync(s.Ticket, nil) {
		s.Success++
		s.Result = "Simulated request completed"
	} else {
		s.Result = "Stale request rejected"
	}
}

func dataPage(id string) *componentPage {
	p := &componentPage{}
	add := func(id, title, description string, build func(*ui.Context, nui.Theme)) {
		p.Scenes = append(p.Scenes, scene(id, title, description, build))
	}
	switch id {
	case "form":
		for _, kind := range []string{"validation", "async", "disabled"} {
			s := &formExample{Result: "Ready"}
			loadingAuto, globalError := true, false
			if kind != "validation" {
				s.Name = "Native workspace"
			}
			p.Deactivate = append(p.Deactivate, s.Form.Cancel)
			description := map[string]string{
				"validation": "Submit an empty name to show validation, then enter a name and submit again. Successful submissions complete immediately.",
				"async":      "Submit to start a simulated request, then press Complete. Cancel, leaving or hiding this page invalidates its ticket.",
				"disabled":   "The populated form is disabled. Its name and submit button ignore input; the submission count stays at zero.",
			}[kind]
			add(kind, map[string]string{"validation": "Validation and synchronous submit", "async": "Simulated asynchronous request", "disabled": "Disabled form"}[kind], description, func(c *ui.Context, t nui.Theme) {
				if kind == "async" {
					nui.Switch(c.Key("loading-auto"), t, &loadingAuto, nui.SwitchOptions{Label: "Disable form while submitting"})
				}
				if kind == "validation" {
					nui.Switch(c.Key("global-error"), t, &globalError, nui.SwitchOptions{Label: "Reject form with a global error"})
				}
				nui.Form(c.Key("form"), t, &s.Form, nui.FormOptions{Label: kind + " workspace form", Disabled: kind == "disabled", LoadingAuto: nui.Set(loadingAuto), Validate: func(values nui.FormValues) nui.FormErrors {
					if globalError {
						return nui.FormErrors{"": "Workspace is unavailable"}
					}
					if strings.TrimSpace(values["name"]) == "" {
						return nui.FormErrors{"name": "Workspace name is required"}
					}
					return nil
				}, OnError: func(nui.FormErrors) { s.Errors++; s.Result = "Validation failed" }, OnSubmit: func(nui.FormValues) {
					s.Requests++
					if kind == "async" {
						s.Ticket = s.Form.BeginAsync()
						s.Result = "Simulated request pending"
					} else {
						s.Success++
						s.Result = "Workspace submitted"
					}
				}}, func(ft nui.Theme) {
					nui.FormField(c, ft, nui.FormFieldOptions{Name: "name", Value: s.Name, Label: kind + " workspace name", Required: true, Help: "Application validation runs on submit."}, func(ft nui.Theme) { nui.Input(c, ft, &s.Name, nui.InputOptions{Label: kind + " workspace name"}) })
					nui.Button(c.Key("submit"), ft, "Submit "+kind, nui.ButtonOptions{}).OnClick(s.Form.Submit)
				})
				if kind == "async" {
					ui.Row(c).Wrap().Gap(8).Children(func() {
						nui.Button(c, t, "Complete simulated request", nui.ButtonOptions{}).OnClick(s.finish)
						nui.Button(c, t, "Cancel request", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Form.Cancel(); s.Result = "Request cancelled" })
						nui.Button(c, t, "Repeat pending submit", nui.ButtonOptions{Variant: nui.Outline}).OnClick(s.Form.Submit)
					})
				}
				if message := s.Form.Errors()[""]; message != "" {
					ui.Text(c.Key("error"), message).TextColor(t.Error)
				}
				if kind == "validation" {
					nui.Button(c.Key("reset"), t, "Restore form name", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Name = "Restored workspace"; s.Form.Reset(); s.Result = "Form reset" })
				}
				ui.Textf(c, "%s · %d successful submissions", s.Result, s.Success)
				ui.Textf(c, "Requests: %d · Errors: %d · Dirty: %t", s.Requests, s.Errors, s.Form.Field("name").Dirty)
			})
		}
	case "input-number":
		contractValue, contractState := nui.Set(18.98), &nui.NumberState{}
		contractChanges := 0
		increment, customIcons, snapping, focus := true, true, true, true
		add("contract", "Starting values and step controls", "Step from 18.98 to the next grid value. Empty input starts at 40. Hold a control to repeat; hiding it keeps the editor intact.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(12).Children(func() {
				nui.Switch(c.Key("snapping"), t, &snapping, nui.SwitchOptions{Label: "Snap number steps"})
				nui.Switch(c.Key("focus"), t, &focus, nui.SwitchOptions{Label: "Focus number after stepping"})
				nui.Switch(c.Key("icons"), t, &customIcons, nui.SwitchOptions{Label: "Custom number control icons"})
			})
			o := nui.InputNumberOptions{Label: "Contract number", Min: nui.Set(0.0), Max: nui.Set(100.0), StartingValue: nui.Set(40.0), StepSnapping: nui.Set(snapping), FocusOnChange: nui.Set(focus), Increment: nui.Set(increment), OnValueChange: func(nui.Optional[float64]) { contractChanges++ }}
			if customIcons {
				o.IncrementIcon = lucide.ArrowUp()
				o.DecrementIcon = lucide.ArrowDown()
				o.IncrementOptions = &nui.ButtonOptions{Variant: nui.Soft, Color: nui.Info}
			}
			nui.InputNumber(c.Key("number"), t, &contractValue, contractState, o)
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c.Key("empty"), t, "Use empty number", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { contractValue = nui.Optional[float64]{} })
				nui.Button(c.Key("limit"), t, "Use number limit", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { contractValue = nui.Set(98.0) })
				nui.Button(c.Key("hide"), t, "Toggle increment control", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { increment = !increment })
			})
			ui.Textf(c, "Contract number: %v · Changes: %d", contractValue, contractChanges)
		})
		for _, kind := range []string{"bounded", "vertical", "states"} {
			value := nui.Optional[float64]{}
			state := &nui.NumberState{}
			disabledValue, disabledState := nui.Set(4.0), &nui.NumberState{}
			direction := nui.Horizontal
			changes := 0
			if kind != "bounded" {
				value = nui.Set(3.0)
			}
			description := "Use arrow keys or the step buttons. Empty is distinct from zero; allowed values are 0–10 in steps of 0.5."
			if kind == "states" {
				description = "The first number is read-only. The second is disabled and marked invalid. Both preserve their initial values."
			}
			add(kind, "Number · "+kind, description, func(c *ui.Context, t nui.Theme) {
				if kind == "bounded" {
					nui.Select(c, t, &direction, []nui.Option[nui.Orientation]{{Value: nui.Horizontal, Label: "horizontal"}, {Value: nui.Vertical, Label: "vertical"}}, nui.SelectOptions{Label: "Number orientation"})
				}
				o := nui.InputNumberOptions{Label: kind + " seats", ReadOnly: kind == "states", Orientation: direction, Min: nui.Set(0.0), Max: nui.Set(10.0), Step: .5, OnValueChange: func(nui.Optional[float64]) { changes++ }}
				if kind == "vertical" {
					o.Orientation = nui.Vertical
				}
				nui.InputNumber(c, t, &value, state, o)
				if kind == "states" {
					nui.InputNumber(c, t, &disabledValue, disabledState, nui.InputNumberOptions{Label: "Disabled seats", Disabled: true, Invalid: true})
				}
				ui.Textf(c, "Value: %v · %d changes", value, changes)
			})
		}
	case "input-tags":
		contractTags, contractTagsState := []string{"go", "vue"}, &nui.TagsState{}
		changes, added, removed, rejected := 0, 0, 0, 0
		uppercase, readOnly := true, false
		add("contract", "Inline tags, keyboard selection and events", "Backspace first selects a tag, then removes it. Values normalize to lowercase; display labels can change without replacing their identity.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(12).Children(func() {
				nui.Switch(c.Key("display"), t, &uppercase, nui.SwitchOptions{Label: "Uppercase tag labels"})
				nui.Switch(c.Key("readonly"), t, &readOnly, nui.SwitchOptions{Label: "Read-only contract tags"})
			})
			o := nui.InputTagsOptions{Label: "Contract tags", Placeholder: "Add a tag", ReadOnly: readOnly, DeleteIcon: lucide.Minus(), MaxTags: 4, AddOnPaste: nui.Set(true), AddOnTab: true, AddOnBlur: true, ConvertValue: strings.ToLower, OnTagsChange: func([]string) { changes++ }, OnAddTag: func(string) { added++ }, OnRemoveTag: func(string) { removed++ }, OnInvalid: func(string) { rejected++ }}
			if uppercase {
				o.DisplayValue = strings.ToUpper
			}
			nui.InputTags(c.Key("tags"), t, &contractTags, contractTagsState, o)
			nui.Button(c.Key("leave"), t, "Leave tags field", nui.ButtonOptions{Variant: nui.Outline})
			nui.Button(c.Key("long"), t, "Add long tag", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() {
				if len(contractTags) < 4 {
					contractTags = append(contractTags, "用于验证窄窗口下标签换行和删除按钮布局的中文内容，用于验证窄窗口下标签换行和删除按钮布局的中文内容")
				}
			})
			ui.Textf(c, "Tag changes: %d · Added: %d · Removed: %d · Rejected: %d", changes, added, removed, rejected)
		})
		for _, kind := range []string{"editable", "paste", "states"} {
			tags := []string{"Go", "Desktop"}
			state := &nui.TagsState{}
			changes := 0
			disabledTags := []string{"Locked"}
			disabledState := &nui.TagsState{}
			if kind == "paste" {
				tags = nil
			}
			description := map[string]string{
				"editable": "Enter adds a tag; Backspace selects then removes. Paste comma/semicolon-separated values. Duplicates are rejected; maximum 6 tags, 12 UTF-16 units per draft.",
				"paste":    "Start with no tags. Paste comma/semicolon-separated values or add on Enter, Tab or blur. Duplicates are allowed; maximum 6 tags, 12 UTF-16 units per draft.",
				"states":   "The first tag list is read-only; the second is disabled. Existing tags stay visible, and neither list allows adding or removing tags.",
			}[kind]
			add(kind, "Tags · "+kind, description, func(c *ui.Context, t nui.Theme) {
				nui.InputTags(c, t, &tags, state, nui.InputTagsOptions{Label: kind + " project tags", ReadOnly: kind == "states", Separators: ",;", MaxTags: 6, MaxLength: 12, AddOnPaste: nui.Set(true), Duplicate: kind == "paste", AddOnTab: kind == "paste", AddOnBlur: kind == "paste", OnTagsChange: func([]string) { changes++ }})
				if kind == "states" {
					nui.InputTags(c, t, &disabledTags, disabledState, nui.InputTagsOptions{Label: "Disabled tags", Disabled: true})
				}
				ui.Textf(c, "Tags: %v · %d changes", tags, changes)
			})
		}
	case "pin-input":
		contractPin, contractPinState := "", &nui.PinState{}
		highlight, styledParts := true, true
		pinChanges, pinComplete := 0, 0
		add("contract", "Highlight and separator parts", "The native PIN uses one editor. Highlight colors every cell; custom separator typography and spacing also reserve the correct width.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("highlight"), t, &highlight, nui.SwitchOptions{Label: "Highlight all PIN cells"})
			nui.Switch(c.Key("parts"), t, &styledParts, nui.SwitchOptions{Label: "Custom PIN separator spacing"})
			o := nui.PinInputOptions{Label: "Contract verification code", Length: 6, Numeric: true, Color: nui.Neutral, Highlight: highlight, SeparatorEvery: 2, OnChange: func(string) { pinChanges++ }, OnComplete: func(string) { pinComplete++ }}
			if styledParts {
				o.UI = nui.Styles{nui.Root: {Gap: nui.Set(float32(10)), PaddingX: nui.Set(float32(4)), PaddingY: nui.Set(float32(4))}, nui.SeparatorSlot: {FontSize: nui.Set(t.TextScale * 20), Foreground: nui.Set(t.Info)}}
			}
			nui.PinInput(c.Key("pin"), t, &contractPin, contractPinState, o)
			ui.Textf(c, "PIN characters: %d · Changes: %d · Complete: %d", len([]rune(contractPin)), pinChanges, pinComplete)
		})
		for _, kind := range []string{"numeric", "masked", "states"} {
			value := ""
			state := &nui.PinState{}
			complete := "Not complete"
			locked := "1234"
			lockedState := &nui.PinState{}
			description := "Paste a six-digit code or move between cells with arrows. A separator divides the numeric code after the third digit."
			switch kind {
			case "masked":
				description = "Enter or paste four text characters. The code is masked; arrow keys move between cells."
			case "states":
				value = "1234"
				description = "The first populated code is read-only; the second is disabled. Both keep their code when typing or pasting."
			}
			add(kind, "Verification · "+kind, description, func(c *ui.Context, t nui.Theme) {
				length := 6
				if kind != "numeric" {
					length = 4
				}
				nui.PinInput(c, t, &value, state, nui.PinInputOptions{Label: kind + " verification code", Length: length, Numeric: kind == "numeric", Mask: kind == "masked", ReadOnly: kind == "states", Separator: []int{3}, OnComplete: func(v string) {
					complete = "Completed: " + v
					if kind == "masked" {
						complete = "Masked code complete"
					}
				}})
				if kind == "states" {
					nui.PinInput(c, t, &locked, lockedState, nui.PinInputOptions{Label: "Disabled verification", Length: 4, Disabled: true})
					ui.Textf(c, "Read-only code: %s · Disabled code: %s", value, locked)
				} else {
					ui.Text(c, complete)
				}
			})
		}
	case "input-rating":
		contractValue, contractChanges, contractMax := 3, 0, 5
		hoverable, customParts, customIcons, readOnly, vertical := false, false, false, false, false
		add("contract", "Rating parts and interaction", "Preview without changing the value. Neutral focus, custom icon layers and orientation keep the same rating; reducing the maximum clamps without a user event.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(12).Children(func() {
				nui.Switch(c.Key("hover"), t, &hoverable, nui.SwitchOptions{Label: "Preview rating on hover"})
				nui.Switch(c.Key("parts"), t, &customParts, nui.SwitchOptions{Label: "Custom rating layer colors"})
				nui.Switch(c.Key("readonly"), t, &readOnly, nui.SwitchOptions{Label: "Read-only contract rating"})
			})
			opts := nui.InputRatingOptions{Label: "Contract rating", Max: contractMax, Color: nui.Neutral, Size: nui.XL, Clearable: true, Hoverable: nui.Set(hoverable), ReadOnly: readOnly, OnChange: func(int) { contractChanges++ }}
			if vertical {
				opts.Orientation = nui.Vertical
			}
			if customIcons {
				opts.Icon, opts.EmptyIcon = lucide.Heart(), lucide.Circle()
			}
			if customParts {
				opts.UI = nui.Styles{nui.Indicator: {Foreground: nui.Set(t.Info)}, nui.EmptyIconSlot: {Foreground: nui.Set(t.Warning)}}
			}
			nui.InputRating(c.Key("rating"), t, &contractValue, opts)
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c.Key("icons"), t, "Toggle rating icons", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { customIcons = !customIcons })
				nui.Button(c.Key("orientation"), t, "Toggle rating orientation", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { vertical = !vertical })
				nui.Button(c.Key("maximum"), t, "Shrink rating maximum", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { contractMax = 3 })
			})
			ui.Textf(c, "Contract rating: %d · Changes: %d", contractValue, contractChanges)
		})
		for _, kind := range []string{"clearable", "vertical", "states"} {
			value := 3
			reduceMotion := false
			locked := 2
			changes := 0
			description := map[string]string{
				"clearable": "Use arrows, Home/End or click a star. Click the current star again to clear the rating.",
				"vertical":  "Use Up/Down or click a heart. This vertical rating has an upper limit of 3 and no hover preview; clicking the current heart clears it.",
				"states":    "The first rating is read-only; the second is disabled. Pointer and keyboard input leave both ratings unchanged.",
			}[kind]
			add(kind, "Rating · "+kind, description, func(c *ui.Context, t nui.Theme) {
				if kind == "clearable" {
					nui.Checkbox(c.Key("reduce-motion"), t, &reduceMotion, nui.CheckboxOptions{Label: "Reduce rating motion"})
					t.Motion.Reduce = t.Motion.Reduce || reduceMotion
				}
				o := nui.InputRatingOptions{Label: kind + " project rating", Clearable: true, ReadOnly: kind == "states", OnChange: func(int) { changes++ }}
				if kind == "vertical" {
					o.Orientation = nui.Vertical
					o.Max = 3
					o.Icon = lucide.Heart()
					o.EmptyIcon = lucide.Heart()
					o.Hoverable = nui.Set(false)
				}
				nui.InputRating(c, t, &value, o)
				if kind == "states" {
					nui.InputRating(c, t, &locked, nui.InputRatingOptions{Label: "Disabled rating", Disabled: true})
				}
				ui.Textf(c, "Rating %d · %d changes", value, changes)
			})
		}
	case "listbox", "select-menu":
		addSelectionContract(p, id == "select-menu")
		items := append([]nui.ListboxItem{}, projectMembers...)
		items = append(items, nui.ListboxItem{ID: "team-eu", Label: "Team", Group: "Regions"}, nui.ListboxItem{ID: "team-us", Label: "Team", Group: "Regions"})
		for _, kind := range []string{"single", "multiple", "states"} {
			if id == "listbox" && kind == "states" {
				var empty, locked nui.ListboxState
				var emptySelection []string
				lockedSelection := []string{"lin"}
				changes := 0
				add(kind, "Members · empty and disabled", "The empty list intentionally has no items and shows a custom message. The disabled list keeps Lin selected and ignores mouse and keyboard input.", func(c *ui.Context, t nui.Theme) {
					ui.Text(c, "Empty list").TextColor(t.Muted)
					nui.Listbox(c.Key("empty"), t, &emptySelection, &empty, nil, nui.ListboxOptions{Label: "Empty members list", Height: 72, EmptyText: "No members"})
					ui.Text(c, "Disabled list").TextColor(t.Muted)
					nui.Listbox(c.Key("disabled"), t, &lockedSelection, &locked, projectMembers[:3], nui.ListboxOptions{Label: "Disabled members list", Disabled: true, Height: 168, OnChange: func([]string) { changes++ }})
					ui.Textf(c, "Disabled selected IDs: %v · %d changes", lockedSelection, changes)
				})
				continue
			}
			chosen := []string{}
			list := &nui.ListboxState{}
			menu := &nui.SelectMenuState{}
			p.Deactivate = append(p.Deactivate, func() { menu.Open = false })
			changes := 0
			disabled := []string{}
			disabledState := &nui.SelectMenuState{}
			search := true
			reset := true
			description := "Arrow keys move, Enter/Space selects; disabled Guest cannot be selected. Duplicate Team labels have different IDs."
			if kind == "states" {
				description = "The first menu shows a loading indicator; the second menu is disabled because no members are available."
			}
			add(kind, "Members · "+kind, description, func(c *ui.Context, t nui.Theme) {
				if id == "listbox" {
					nui.Listbox(c, t, &chosen, list, items, nui.ListboxOptions{Label: kind + " members list", Multiple: kind == "multiple", Height: 200, EmptyText: "No members", OnChange: func([]string) { changes++ }})
				} else {
					nui.Switch(c, t, &search, nui.SwitchOptions{Label: kind + " search input"})
					nui.Switch(c, t, &reset, nui.SwitchOptions{Label: kind + " reset search"})
					nui.SelectMenu(c, t, &chosen, menu, items, nui.SelectMenuOptions{Label: kind + " members menu", Multiple: kind == "multiple", OnChange: func([]string) { changes++ }, Placeholder: "Choose members", SearchPlaceholder: "Search people", SearchInput: nui.Set(search), ResetSearchTermOnBlur: nui.Set(reset), ResetSearchTermOnSelect: nui.Set(reset), Clearable: true, Loading: kind == "states"})
					if kind == "states" {
						nui.SelectMenu(c, t, &disabled, disabledState, nil, nui.SelectMenuOptions{Label: "Disabled member menu", Disabled: true, Placeholder: "No available members"})
					}
				}
				ui.Textf(c, "Selected IDs: %v · %d changes", chosen, changes)
			})
		}
	case "calendar":
		addCalendarContract(p)
		for _, kind := range []string{"single", "range", "constraints"} {
			value := nui.DateSelection{}
			state := &nui.CalendarState{}
			changes := 0
			add(kind, "Calendar · "+kind, "Arrow keys move days; Page Up/Down changes month. The constraint sample allows November 2026 weekdays only. Selection uses civil dates, without time zones.", func(c *ui.Context, t nui.Theme) {
				o := nui.CalendarOptions{Label: kind + " calendar", Range: kind == "range", Today: nui.Set(nui.Date{Year: 2026, Month: 11, Day: 5}), OnChange: func(nui.DateSelection) { changes++ }}
				if kind == "constraints" {
					o.Min = nui.Set(nui.Date{Year: 2026, Month: 11, Day: 1})
					o.Max = nui.Set(nui.Date{Year: 2026, Month: 11, Day: 30})
					o.IsDateDisabled = func(d nui.Date) bool {
						return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).Weekday() == time.Saturday || time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).Weekday() == time.Sunday
					}
				}
				nui.Calendar(c, t, &value, state, o)
				ui.Textf(c, "Selection: %v · %d changes", value, changes)
			})
		}
	case "input-date":
		addDateInputContract(p)
		for _, kind := range []string{"segments", "calendar", "range", "states"} {
			value := nui.Optional[nui.Date]{}
			state := &nui.DateInputState{}
			selection := nui.DateSelection{}
			rangeState := &nui.DateRangeInputState{}
			disabledValue := nui.Set(nui.Date{Year: 2026, Month: 11, Day: 5})
			disabledState := &nui.DateInputState{}
			p.Deactivate = append(p.Deactivate, func() {
				state.Open = false
				rangeState.Open = false
				rangeState.Start.Open = false
				rangeState.End.Open = false
			})
			description := map[string]string{
				"segments": "Edit date segments with arrows or paste ISO date text. This input has no calendar button; an empty value is distinct from a date.",
				"calendar": "Edit segments, paste ISO date text or open the calendar. Selection is constrained to November 2026.",
				"range":    "Edit the start and end dates or choose a range in the shared calendar. This sample has no month constraint.",
				"states":   "The first populated date is read-only; the second is disabled. Neither allows editing or opening a calendar.",
			}[kind]
			if kind == "states" {
				value = disabledValue
			}
			add(kind, "Date · "+kind, description, func(c *ui.Context, t nui.Theme) {
				o := nui.InputDateOptions{Label: kind + " release date", ReadOnly: kind == "states", ShowCalendar: kind != "segments"}
				if kind == "calendar" {
					o.Calendar.Min = nui.Set(nui.Date{Year: 2026, Month: 11, Day: 1})
					o.Calendar.Max = nui.Set(nui.Date{Year: 2026, Month: 11, Day: 30})
				}
				if kind == "range" {
					nui.InputDateRange(c, t, &selection, rangeState, o)
					ui.Textf(c, "Range: %v", selection)
				} else {
					nui.InputDate(c, t, &value, state, o)
					ui.Textf(c, "Date: %v", value)
				}
				if kind == "states" {
					nui.InputDate(c, t, &disabledValue, disabledState, nui.InputDateOptions{Label: "Disabled release date", Disabled: true})
				}
			})
		}
	case "input-time":
		addTimeInputContract(p)
		for _, kind := range []string{"24-hour", "12-hour", "states"} {
			value := nui.Optional[nui.ClockTime]{}
			state := &nui.TimeInputState{}
			locked := nui.Set(nui.ClockTime{Hour: 9, Minute: 30})
			lockedState := &nui.TimeInputState{}
			description := "Use arrows within each time segment or paste a time in 24-hour format. No date or time-zone conversion is performed."
			switch kind {
			case "12-hour":
				description = "Edit hours, minutes, seconds and AM/PM. No date or time-zone conversion is performed."
			case "states":
				value = locked
				description = "The first populated time is read-only; the second is disabled. Both retain 09:30 when typing or using arrow keys."
			}
			add(kind, "Time · "+kind, description, func(c *ui.Context, t nui.Theme) {
				nui.InputTime(c, t, &value, state, nui.InputTimeOptions{Label: kind + " reminder", ReadOnly: kind == "states", TwelveHour: nui.Set(kind == "12-hour"), Seconds: kind == "12-hour"})
				if kind == "states" {
					nui.InputTime(c, t, &locked, lockedState, nui.InputTimeOptions{Label: "Disabled reminder", Disabled: true})
				}
				ui.Textf(c, "Time: %v", value)
			})
		}
	case "color-picker":
		addColorPickerContract(p)
		for _, kind := range []string{"palette", "hex", "disabled"} {
			value := ui.Hex("#22c55e")
			state := &nui.ColorPickerState{}
			changes := 0
			description := "Drag the saturation/value area or hue slider, or edit HEX. Keyboard arrows adjust the focused control. Each picker owns its value."
			if kind == "disabled" {
				description = "The picker and its color fields are disabled. Pointer and keyboard input preserve the initial green color and do not fire change callbacks."
			}
			add(kind, "Color · "+kind, description, func(c *ui.Context, t nui.Theme) {
				nui.ColorPicker(c, t, &value, state, nui.ColorPickerOptions{Label: kind + " accent", Disabled: kind == "disabled", OnChange: func(ui.Color) { changes++ }})
				if kind == "hex" {
					nui.Button(c, t, "Set blue accent", nui.ButtonOptions{}).OnClick(func() { value = ui.Hex("#3b82f6") })
				}
				ui.Textf(c, "Color: %v · %d changes", value, changes)
			})
		}
	case "table":
		for _, sample := range []struct {
			id, title string
			rows      int
		}{{"small", "Small sortable data", 12}, {"empty", "Empty table", 0}, {"virtual", "10,000 virtual rows", 10000}} {
			model := &State{}
			model.SetRows(sample.rows)
			loading := false
			sortChanges, selectionChanges := 0, 0
			activated := "None"
			description := "Click Workspace/Members to sort and drag a header boundary to resize. Selection follows record IDs, including after sorting. Wheel scrolls the bounded table; only visible cells build."
			if sample.rows == 0 {
				description = "This table initially has no records. Its column headers remain visible, and selecting the first record is disabled while empty. Add a sample record to compare the populated state."
			}
			add(sample.id, sample.title, description, func(c *ui.Context, t nui.Theme) {
				ui.Row(c).Wrap().Gap(8).Children(func() {
					if sample.rows == 0 {
						nui.Button(c, t, "Add sample record", nui.ButtonOptions{Disabled: len(model.Rows) != 0}).OnClick(func() { model.SetRows(1) })
						nui.Button(c, t, "Remove sample record", nui.ButtonOptions{Variant: nui.Outline, Disabled: len(model.Rows) == 0}).OnClick(func() { model.SetRows(0); model.Selection.Clear() })
					}
					nui.Button(c, t, "Select first record", nui.ButtonOptions{Disabled: len(model.Rows) == 0}).OnClick(func() { model.Selection.Add(model.Rows[0].ID) })
					nui.Button(c, t, "Clear selection", nui.ButtonOptions{Variant: nui.Outline}).OnClick(model.Selection.Clear)
					nui.Button(c, t, "Clear sorting", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { model.Sort = ui.SortOrder{} })
					nui.Checkbox(c, t, &loading, nui.CheckboxOptions{Label: "Show table loading"})
				})
				model.TableConfig = nui.TableOptions{Loading: loading, Caption: "Workspace records", BuildLoading: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Loading workspaces") }, OnSortChange: func([]ui.SortOrder) { sortChanges++ }, OnSelectionChange: func([]string) { selectionChanges++ }, OnActivate: func(id string) { activated = id }}
				if sample.rows == 0 && len(model.Rows) == 0 {
					model.TableConfig.Height = 140
				}
				model.Table(c, t)
				ui.Textf(c, "%d rows · %d selected · %d cells built", len(model.Rows), model.Selection.Len(), model.CellCalls)
				ui.Textf(c, "Sort changes: %d · Selection changes: %d · Activated: %s", sortChanges, selectionChanges, activated)
			})
		}
		add("contract", "Sorting and loading", "Multi-column and manual sorting, stable IDs, column styles and loading content use the public Table component. Enter or double-click activates a row; sorting headers are mouse controls in MyGo.", tableParityExample())
	case "tree":
		addTreeContract(p)
		for _, kind := range []string{"nested", "linked", "lazy"} {
			state := &nui.TreeState{}
			state.SetNodes([]nui.TreeNode{{ID: "project", Label: "Project", DefaultExpanded: true, Children: []nui.TreeNode{{ID: "source", Label: "Source", Children: []nui.TreeNode{{ID: "main", Label: "main.go"}, {ID: "theme", Label: "theme.go"}}}, {ID: "guide", Label: "Guide"}, {ID: "locked", Label: "Locked", Disabled: true}}}, {ID: "remote", Label: "Load examples", Lazy: kind == "lazy", Children: func() []nui.TreeNode {
				if kind == "lazy" {
					return nil
				}
				return []nui.TreeNode{{ID: "example-button", Label: "Button example"}, {ID: "example-input", Label: "Input example"}}
			}()}})
			ticket := nui.TreeTicket{}
			requested := false
			flat := false
			result := "Ready"
			p.Deactivate = append(p.Deactivate, func() { state.Close(); state.SetExpanded("remote", false) })
			add(kind, "Tree · "+kind, "Arrow Right/Left expands/collapses; Space selects stable IDs. Linked multi-selection propagates through loaded nodes only. Lazy loading is simulated with Complete/Cancel; native drag-and-drop reordering is not implemented.", func(c *ui.Context, t nui.Theme) {
				nui.Switch(c, t, &flat, nui.SwitchOptions{Label: kind + " flat rows"})
				nui.Tree(c, t, state, nui.TreeOptions{Label: kind + " project tree", Height: 240, Nested: nui.Set(!flat), Multiple: kind != "nested", PropagateSelect: kind == "linked", BubbleSelect: kind == "linked", OnChange: func() { result = fmt.Sprintf("%d selected", state.Selection.Len()) }, OnLoad: func(_ string, next nui.TreeTicket) {
					ticket = next
					requested = true
					result = "Simulated load pending"
				}})
				if kind == "lazy" {
					ui.Row(c).Wrap().Gap(8).Children(func() {
						nui.Button(c, t, "Complete tree load", nui.ButtonOptions{Disabled: !requested}).OnClick(func() {
							if state.CompleteLoad(ticket, []nui.TreeNode{{ID: "example-button", Label: "Button example"}, {ID: "example-input", Label: "Input example"}}, nil) {
								result = "Examples loaded"
							} else {
								result = "Stale tree result rejected"
							}
						})
						nui.Button(c, t, "Cancel tree load", nui.ButtonOptions{}).OnClick(func() {
							state.SetExpanded("remote", false)
							state.Close()
							result = "Tree load cancelled"
						})
					})
				}
				ui.Textf(c, "%s · %d visible rows · %d selected", result, state.Rows(), state.Selection.Len())
			})
		}
	}
	return p
}
