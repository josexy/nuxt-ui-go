package gallery

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// Simulated completions are explicit UI actions: no network or timer owns a
// window after unmount. Tickets still exercise cancellation and stale results.
type formExample struct {
	Form    nui.FormState
	Ticket  nui.FormTicket
	Name    string
	Success int
	Result  string
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
			if kind != "validation" {
				s.Name = "Native workspace"
			}
			p.Deactivate = append(p.Deactivate, s.Form.Cancel)
			add(kind, map[string]string{"validation": "Validation and synchronous submit", "async": "Simulated asynchronous request", "disabled": "Disabled form"}[kind], "Enter a name, then submit. The simulated request completes only when you press Complete; Cancel, leaving or hiding this page invalidates its ticket.", func(c *ui.Context, t nui.Theme) {
				nui.Form(c, t, &s.Form, nui.FormOptions{Label: kind + " workspace form", Disabled: kind == "disabled", Validate: func(values nui.FormValues) nui.FormErrors {
					if strings.TrimSpace(values["name"]) == "" {
						return nui.FormErrors{"name": "Workspace name is required"}
					}
					return nil
				}, OnSubmit: func(nui.FormValues) {
					if kind == "async" {
						s.Ticket = s.Form.BeginAsync()
						s.Result = "Simulated request pending"
					} else {
						s.Success++
						s.Result = "Workspace submitted"
					}
				}}, func(ft nui.Theme) {
					nui.FormField(c, ft, nui.FormFieldOptions{Name: "name", Value: s.Name, Label: kind + " workspace name", Required: true, Help: "Application validation runs on submit."}, func(ft nui.Theme) { nui.Input(c, ft, &s.Name, nui.InputOptions{Label: kind + " workspace name"}) })
					nui.Button(c, t, "Submit "+kind, nui.ButtonOptions{Disabled: kind == "disabled" || s.Form.Submitting}).OnClick(s.Form.Submit)
				})
				if kind == "async" {
					ui.Row(c).Wrap().Gap(8).Children(func() {
						nui.Button(c, t, "Complete simulated request", nui.ButtonOptions{}).OnClick(s.finish)
						nui.Button(c, t, "Cancel request", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Form.Cancel(); s.Result = "Request cancelled" })
					})
				}
				ui.Textf(c, "%s · %d successful submissions", s.Result, s.Success)
			})
		}
	case "input-number":
		for _, kind := range []string{"bounded", "vertical", "states"} {
			value := nui.Optional[float64]{}
			state := &nui.NumberState{}
			disabledValue, disabledState := nui.Set(4.0), &nui.NumberState{}
			direction := nui.Horizontal
			changes := 0
			if kind != "bounded" {
				value = nui.Set(3.0)
			}
			add(kind, "Number · "+kind, "Use arrow keys or the step buttons. Empty is distinct from zero; allowed values are 0–10 in steps of 0.5.", func(c *ui.Context, t nui.Theme) {
				if kind == "bounded" {
					nui.Select(c, t, &direction, []nui.Option[nui.Orientation]{{Value: nui.Horizontal, Label: "horizontal"}, {Value: nui.Vertical, Label: "vertical"}}, nui.SelectOptions{Label: "Number orientation"})
				}
				o := nui.InputNumberOptions{InputOptions: nui.InputOptions{Label: kind + " seats", ReadOnly: kind == "states"}, Orientation: direction, Min: nui.Set(0.0), Max: nui.Set(10.0), Step: .5, OnValueChange: func(nui.Optional[float64]) { changes++ }}
				if kind == "vertical" {
					o.Orientation = nui.Vertical
				}
				nui.InputNumber(c, t, &value, state, o)
				if kind == "states" {
					nui.InputNumber(c, t, &disabledValue, disabledState, nui.InputNumberOptions{InputOptions: nui.InputOptions{Label: "Disabled seats", Disabled: true, Invalid: true}})
				}
				ui.Textf(c, "Value: %v · %d changes", value, changes)
			})
		}
	case "input-tags":
		for _, kind := range []string{"editable", "paste", "states"} {
			tags := []string{"Go", "Desktop"}
			state := &nui.TagsState{}
			changes := 0
			disabledTags := []string{"Locked"}
			disabledState := &nui.TagsState{}
			if kind == "paste" {
				tags = nil
			}
			add(kind, "Tags · "+kind, "Enter adds a tag; Backspace removes. Paste comma/semicolon-separated values. The policy sample permits duplicates and adds on Tab/blur; maximum 6 tags, 12 UTF-16 units per draft.", func(c *ui.Context, t nui.Theme) {
				nui.InputTags(c, t, &tags, state, nui.InputTagsOptions{InputOptions: nui.InputOptions{Label: kind + " project tags", ReadOnly: kind == "states"}, Separators: ",;", MaxTags: 6, MaxLength: 12, AddOnPaste: nui.Set(true), Duplicate: kind == "paste", AddOnTab: kind == "paste", AddOnBlur: kind == "paste", OnTagsChange: func([]string) { changes++ }})
				if kind == "states" {
					nui.InputTags(c, t, &disabledTags, disabledState, nui.InputTagsOptions{InputOptions: nui.InputOptions{Label: "Disabled tags", Disabled: true}})
				}
				ui.Textf(c, "Tags: %v · %d changes", tags, changes)
			})
		}
	case "pin-input":
		for _, kind := range []string{"numeric", "masked", "states"} {
			value := ""
			state := &nui.PinState{}
			complete := "Not complete"
			locked := "1234"
			lockedState := &nui.PinState{}
			add(kind, "Verification · "+kind, "Paste a complete code or move between cells with arrows. Numeric uses six cells with a separator; text accepts four masked characters.", func(c *ui.Context, t nui.Theme) {
				length := 6
				if kind != "numeric" {
					length = 4
				}
				nui.PinInput(c, t, &value, state, nui.PinInputOptions{Label: kind + " verification code", Length: length, Numeric: kind == "numeric", Mask: kind == "masked", ReadOnly: kind == "states", Separator: []int{3}, OnComplete: func(v string) { complete = "Completed: " + v }})
				if kind == "states" {
					nui.PinInput(c, t, &locked, lockedState, nui.PinInputOptions{Label: "Disabled verification", Length: 4, Disabled: true})
				}
				ui.Text(c, complete)
			})
		}
	case "input-rating":
		for _, kind := range []string{"clearable", "vertical", "states"} {
			value := 3
			locked := 2
			changes := 0
			add(kind, "Rating · "+kind, "Use arrows, Home/End or click a star. Click the current star again to clear the first sample; the vertical sample uses hearts and an upper limit of 3.", func(c *ui.Context, t nui.Theme) {
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
		items := append([]nui.ListboxItem{}, projectMembers...)
		items = append(items, nui.ListboxItem{ID: "team-eu", Label: "Team", Group: "Regions"}, nui.ListboxItem{ID: "team-us", Label: "Team", Group: "Regions"})
		for _, kind := range []string{"single", "multiple", "states"} {
			chosen := []string{}
			list := &nui.ListboxState{}
			menu := &nui.SelectMenuState{}
			p.Deactivate = append(p.Deactivate, func() { menu.Open = false })
			changes := 0
			disabled := []string{}
			disabledState := &nui.SelectMenuState{}
			search := true
			reset := true
			add(kind, "Members · "+kind, "Arrow keys move, Enter/Space selects; disabled Guest cannot be selected. Duplicate Team labels have different IDs. Empty/filter states display their own message.", func(c *ui.Context, t nui.Theme) {
				if id == "listbox" {
					visible := items
					if kind == "states" {
						visible = nil
					}
					nui.Listbox(c, t, &chosen, list, visible, nui.ListboxOptions{Label: kind + " members list", Multiple: kind == "multiple", Height: 200, EmptyText: "No members", OnChange: func([]string) { changes++ }})
				} else {
					nui.Switch(c, t, &search, nui.SwitchOptions{Label: kind + " search input"})
					nui.Switch(c, t, &reset, nui.SwitchOptions{Label: kind + " reset search"})
					nui.SelectMenu(c, t, &chosen, menu, items, nui.SelectMenuOptions{ListboxOptions: nui.ListboxOptions{Label: kind + " members menu", Multiple: kind == "multiple", OnChange: func([]string) { changes++ }}, Placeholder: "Choose members", SearchPlaceholder: "Search people", SearchInput: nui.Set(search), ResetSearchTermOnBlur: nui.Set(reset), ResetSearchTermOnSelect: nui.Set(reset), Clearable: true, Loading: kind == "states"})
					if kind == "states" {
						nui.SelectMenu(c, t, &disabled, disabledState, nil, nui.SelectMenuOptions{ListboxOptions: nui.ListboxOptions{Label: "Disabled member menu", Disabled: true}, Placeholder: "No available members"})
					}
				}
				ui.Textf(c, "Selected IDs: %v · %d changes", chosen, changes)
			})
		}
	case "calendar":
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
			add(kind, "Date · "+kind, "Edit date segments with arrows or paste ISO date text. Calendar constrains selection to November 2026; empty values remain distinct.", func(c *ui.Context, t nui.Theme) {
				o := nui.InputDateOptions{InputOptions: nui.InputOptions{Label: kind + " release date", ReadOnly: kind == "states"}, ShowCalendar: kind != "segments"}
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
					nui.InputDate(c, t, &disabledValue, disabledState, nui.InputDateOptions{InputOptions: nui.InputOptions{Label: "Disabled release date", Disabled: true}})
				}
			})
		}
	case "input-time":
		for _, kind := range []string{"24-hour", "12-hour", "states"} {
			value := nui.Optional[nui.ClockTime]{}
			state := &nui.TimeInputState{}
			locked := nui.Set(nui.ClockTime{Hour: 9, Minute: 30})
			lockedState := &nui.TimeInputState{}
			add(kind, "Time · "+kind, "Use arrows within each time segment. The 12-hour sample includes seconds and AM/PM; no date or time-zone conversion is performed.", func(c *ui.Context, t nui.Theme) {
				nui.InputTime(c, t, &value, state, nui.InputTimeOptions{InputOptions: nui.InputOptions{Label: kind + " reminder", ReadOnly: kind == "states"}, TwelveHour: nui.Set(kind == "12-hour"), Seconds: kind == "12-hour"})
				if kind == "states" {
					nui.InputTime(c, t, &locked, lockedState, nui.InputTimeOptions{InputOptions: nui.InputOptions{Label: "Disabled reminder", Disabled: true}})
				}
				ui.Textf(c, "Time: %v", value)
			})
		}
	case "color-picker":
		for _, kind := range []string{"palette", "hex", "disabled"} {
			value := ui.Hex("#22c55e")
			state := &nui.ColorPickerState{}
			changes := 0
			add(kind, "Color · "+kind, "Drag the saturation/value area or hue slider, or edit HEX. Keyboard arrows adjust the focused control. Each picker owns its value.", func(c *ui.Context, t nui.Theme) {
				nui.ColorPicker(c, t, &value, state, nui.ColorPickerOptions{Label: kind + " accent", Disabled: kind == "disabled", OnChange: func(ui.Color) { changes++ }})
				if kind == "hex" {
					nui.Button(c, t, "Set blue accent", nui.ButtonOptions{}).OnClick(func() { value = ui.Hex("#3b82f6"); changes++ })
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
			add(sample.id, sample.title, "Click Workspace/Members to sort and drag a header boundary to resize. Selection follows record IDs, including after sorting. Wheel scrolls the bounded table; only visible cells build.", func(c *ui.Context, t nui.Theme) {
				ui.Row(c).Wrap().Gap(8).Children(func() {
					nui.Button(c, t, "Select first record", nui.ButtonOptions{Disabled: len(model.Rows) == 0}).OnClick(func() { model.Selection.Add(model.Rows[0].ID) })
					nui.Button(c, t, "Clear selection", nui.ButtonOptions{Variant: nui.Outline}).OnClick(model.Selection.Clear)
				})
				model.Table(c, t)
				ui.Textf(c, "%d rows · %d selected · %d cells built", len(model.Rows), model.Selection.Len(), model.CellCalls)
			})
		}
	case "tree":
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
