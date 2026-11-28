package nui_test

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSelectNeutralFocusUsesInvertedBackground(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(fmt.Sprint(multiple), func(t *testing.T) {
			value, values := "a", []string{"a"}
			accent := ui.RGB(126, 34, 206)
			selectState1 := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
			selectedState := chosenSelection(values)
			selectState2 := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				th.InvertedBackground, th.HighContrast = accent, true
				ui.Column(c).Padding(16).Children(func() {
					opts := nui.SelectOptions{Label: "Choice", Color: nui.Neutral, Style: func(part nui.Slot, e ui.Element) {
						if part == nui.Root {
							e.Label("Select trigger")
						}
					}}

					if multiple {
						nui.SelectMultiple(c, th, selectedState, selectState1, opts).Width(240)
					} else {
						nui.Select(c, th, &value, selectState2, opts).Width(240)
					}
				})
			}, 300, 120)
			tt.Key(0, ui.KeyTab)
			tt.Advance(200 * time.Millisecond) // the colour transition
			r := bounds(t, tt, "Select trigger")
			if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)-2); got != rgba(accent) {
				t.Fatalf("neutral focus=%v, want %v", got, rgba(accent))
			}
		})
	}
}

func TestSelectAccessoriesControlledOpenAndLoading(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(fmt.Sprint(multiple), func(t *testing.T) {
			value, values, changes, open := "a", []string{"a"}, 0, false
			avatar := nui.AvatarOptions{Alt: "Choice avatar", Text: "AL"}
			o := nui.SelectOptions{Label: "Choice", Avatar: &avatar, Open: &open, OnChange: func() { changes++ }, UI: nui.Styles{nui.Trailing: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(part nui.Slot, e ui.Element) {
				if part == nui.Trailing {
					e.Label("Choice trailing")
				}
			}}
			items := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "blocked", Label: "Blocked", Disabled: true}, {Value: "b", Label: "Beta"}}
			selectState1 := submittedSelect(items)
			selectedState := chosenSelection(values)
			selectState2 := submittedSelect(items)
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				th.Motion.Reduce = true
				nui.FormField(c, th, nui.FormFieldOptions{Label: "Outer field"}, func(ft nui.Theme) {
					if multiple {
						nui.SelectMultiple(c, ft, selectedState, selectState1, o).Width(260)
					} else {
						nui.Select(c, ft, &value, selectState2, o).Width(260)
					}
				})
			}, 360, 300)
			for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
				o.Size = size
				tt.Frame()
				if r := bounds(t, tt, "Choice avatar"); r.W != []float32{16, 16, 20, 20, 24}[i] {
					t.Fatalf("size %d avatar=%v", size, r)
				}
			}
			o.Size = nui.MD
			tt.Key(0, ui.KeyTab)
			o.Loading, o.LoadingTrailing = true, true
			tt.Frame()
			if !tt.Focused("Choice") || !tt.HasText("Choice avatar") {
				t.Fatal("loading change lost trigger focus or avatar")
			}
			if countWhere(tt.Image(), bounds(t, tt, "Choice trailing"), func(c color.RGBA) bool { return int(c.R) > int(c.G)+20 && int(c.B) > int(c.G)+20 }) == 0 {
				t.Fatal("loader ignored Trailing foreground")
			}
			tt.Key(0, ui.KeySpace)
			if !open || !tt.HasText("Beta") {
				t.Fatal("loading prevented opening")
			}
			click(t, tt, "Beta")
			if changes != 1 || !multiple && (value != "b" || open) || multiple && (!slices.Equal(selectedState.Values(), []string{"a", "b"}) || !open) {
				t.Fatalf("loading choice: %q %v open=%v changes=%d", value, selectedState.Values(), open, changes)
			}
			if multiple {
				tt.Key(0, ui.KeyEscape)
			}
			if open || !tt.Focused("Choice") {
				t.Fatal("close did not restore the named trigger")
			}
			o.Leading = lucide.Star()
			tt.Frame()
			if tt.HasText("Choice avatar") {
				t.Fatal("leading loader did not replace avatar")
			}
			o.Loading, o.Leading = false, nil
			avatar.Size = nui.XS
			open = true
			tt.Frame()
			if r := bounds(t, tt, "Choice avatar"); r.W != 24 {
				t.Fatal("explicit avatar size was ignored")
			}
			if !tt.HasText("Beta") || changes != 1 {
				t.Fatal("controlled open failed or notified selection")
			}
			open = false
			tt.Frame()
			if !tt.Focused("Choice") || tt.HasText("Blocked") {
				t.Fatal("controlled close did not restore focus")
			}
		})
	}
}

func TestSelectHighlightRTLAndSelectedIcon(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			t.Run(fmt.Sprint(multiple, rtl), func(t *testing.T) {
				value, values, open := "a", []string{"a"}, false
				accent := ui.RGB(126, 34, 206)
				marker := ui.MustParseSVG([]byte(`<svg viewBox="0 0 10 10"><rect width="10" height="10" fill="#ff00ff"/></svg>`))
				selectState1 := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First choice"}})
				selectedState := chosenSelection(values)
				selectState2 := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First choice"}})
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
					th.Motion.Disabled = true
					th.InvertedBackground = accent
					ui.Column(c).Padding(16).Children(func() {
						o := nui.SelectOptions{Label: "Choice", Color: nui.Neutral, Variant: nui.None, Highlight: true, Open: &open, Leading: lucide.Star(), Trailing: lucide.Plus(), SelectedIcon: marker, UI: nui.Styles{nui.Item: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(p nui.Slot, e ui.Element) {
							if p == nui.Leading {
								e.Label("Leading icon")
							}
							if p == nui.Trailing {
								e.Label("Trailing icon")
							}
							if p == nui.Item {
								e.Label("Popup row")
							}
						}}

						if multiple {
							nui.SelectMultiple(c, th, selectedState, selectState1, o).Width(260)
						} else {
							nui.Select(c, th, &value, selectState2, o).Width(260)
						}
					})
				}, 360, 280)
				if rtl {
					tt.SetDirection(ui.RightToLeft)
				}
				r := bounds(t, tt, "Choice")
				if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)); got != rgba(accent) {
					t.Fatalf("None highlight=%v", got)
				}
				leading, trailing := bounds(t, tt, "Leading icon"), bounds(t, tt, "Trailing icon")
				if (!rtl && leading.X >= trailing.X) || (rtl && leading.X <= trailing.X) {
					t.Fatal("adornments ignored direction")
				}
				click(t, tt, "Choice")
				row := bounds(t, tt, "Popup row")
				x := row.X + row.W - 16
				if rtl {
					x = row.X + 16
				}
				if pixel := tt.Image().RGBAAt(int(x), int(row.Y+8)); pixel != rgba(ui.RGB(255, 0, 255)) {
					t.Fatalf("selected icon not aligned with first text line: %v, row=%+v", pixel, row)
				}
			})
		}
	}
}

func TestSelectMultiplePrefixKeepsCurrentMatchAndCycles(t *testing.T) {
	var values []string
	changes, open := 0, false
	items := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "x", Label: "Blocked", Disabled: true}, {Value: "c", Label: "Boreal"}}
	selectState := submittedSelect(items)
	selectedState := chosenSelection(values)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.SelectMultiple(c, nui.App(c, nui.AppConfig{}), selectedState, selectState, nui.SelectOptions{Label: "Choice", Open: &open, OnChange: func() { changes++ }}).Width(260)
	}, 360, 320)
	click(t, tt, "Choice")
	tt.Type("be")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(selectedState.Values(), []string{"b"}) || changes != 1 {
		t.Fatalf("prefix be: %v/%d", selectedState.Values(), changes)
	}
	tt.Advance(2 * time.Second)
	tt.Type("b")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(selectedState.Values(), []string{"b", "c"}) || changes != 2 {
		t.Fatalf("cycle b: %v/%d", selectedState.Values(), changes)
	}
	items = []nui.Option[string]{items[3], items[0], items[2], items[1]}
	selectState.SetOptions(items)
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(selectedState.Values(), []string{"b"}) || changes != 3 {
		t.Fatal("reorder changed active identity")
	}
	tt.Key(0, ui.KeyEscape)
	if open || !tt.Focused("Choice") {
		t.Fatal("Escape did not restore focus")
	}
}

func TestSelectMultipleEmptyLabelsAndInheritedDisable(t *testing.T) {
	values, changes, open, disabled := []string{"a", "b"}, 0, false, false
	items := []nui.Option[string]{{Value: "a", Label: ""}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}}
	selectState := submittedSelect(items)
	selectedState := chosenSelection(values)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Box(c).Disabled(disabled).Children(func() {
			nui.SelectMultiple(c, th, selectedState, selectState, nui.SelectOptions{Label: "Choice", Placeholder: "Pick a choice", Open: &open, OnChange: func() { changes++ }}).Width(260)
		})
	}, 360, 320)
	if !tt.HasText("Beta") || tt.HasText(", Beta") {
		t.Fatalf("empty label left a separator: %v", tt.Texts())
	}
	click(t, tt, "Choice")
	disabled = true
	tt.Frame()
	if _, ok := tt.Find("Gamma"); ok {
		click(t, tt, "Gamma")
	} else {
		click(t, tt, "Choice")
		tt.Key(0, ui.KeyEnter)
	}
	if changes != 0 || !slices.Equal(selectedState.Values(), []string{"a", "b"}) {
		t.Fatal("inherited disable allowed an open row selection")
	}
	disabled = false
	selectedState.Set([]string{"a"})

	tt.Frame()
	open = false
	tt.Frame()
	if !tt.HasText("Pick a choice") || changes != 0 {
		t.Fatal("empty display did not use placeholder or external value notified")
	}
}

func TestSelectKeyboardPrefixClosedAndOpen(t *testing.T) {
	value, changes, disabled := "a", 0, false
	items := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}, {Value: "blocked", Label: "Blocked", Disabled: true}, {Value: "c", Label: "Boreal"}, {Value: "d", Label: "Delta"}}
	selectState := submittedSelect(items)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Select(c, nui.App(c, nui.AppConfig{}), &value, selectState, nui.SelectOptions{Label: "Choice", Disabled: disabled, OnChange: func() { changes++ }}).Width(240)
	}, 320, 300)
	tt.Key(0, ui.KeyTab)
	tt.Type("b")
	if value != "b" || changes != 1 {
		t.Fatalf("closed prefix value=%q changes=%d", value, changes)
	}
	tt.Type("b")
	if value != "c" || changes != 2 {
		t.Fatal("repeating a prefix did not cycle past the disabled item")
	}
	tt.Advance(2 * time.Second)
	tt.Key(0, ui.KeySpace)
	tt.Type("d")
	if value != "c" || changes != 2 {
		t.Fatal("open prefix committed before Enter")
	}
	tt.Key(0, ui.KeyEnter)
	if value != "d" || changes != 3 || !tt.Focused("Choice") {
		t.Fatal("open prefix did not select once and restore trigger focus")
	}
	disabled = true
	tt.Frame()
	tt.Type("a")
	if value != "d" || changes != 3 {
		t.Fatal("disabled select accepted a prefix")
	}
}

func TestSelectPopupDismissibleAndExplicitClose(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(fmt.Sprint(multiple), func(t *testing.T) {
			value, values, open, changes := "a", []string{"a"}, false, 0
			selectState1 := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
			selectedState := chosenSelection(values)
			selectState2 := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				o := nui.SelectOptions{Label: "Choice", Open: &open, Popup: nui.PopoverOptions{Dismissible: nui.Set(false)}, OnChange: func() { changes++ }}

				if multiple {
					nui.SelectMultiple(c, th, selectedState, selectState1, o).Width(240)
				} else {
					nui.Select(c, th, &value, selectState2, o).Width(240)
				}
			}, 320, 300)
			click(t, tt, "Choice")
			tt.Key(0, ui.KeyEscape)
			if !open || !tt.HasText("Beta") {
				t.Fatal("Dismissible=false allowed Escape")
			}
			open = false
			tt.Frame()
			if open || tt.HasText("Beta") || changes != 0 || !tt.Focused("Choice") {
				t.Fatal("explicit close failed or changed value")
			}
		})
	}
}

func TestSelectMultiplePanelBackgroundOverride(t *testing.T) {
	values, open := []string{"a"}, true
	accent := ui.RGB(255, 0, 255)
	selectState := submittedSelect([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
	selectedState := chosenSelection(values)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Disabled = true
		ui.Column(c).Padding(16).Children(func() {
			nui.SelectMultiple(c, th, selectedState, selectState, nui.SelectOptions{Label: "Choice", Open: &open, Popup: nui.PopoverOptions{Label: "Selection panel"}, UI: nui.Styles{nui.Panel: {Background: nui.Set(accent)}}}).Width(240)
		})
	}, 320, 240)
	r := bounds(t, tt, "Selection panel")
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+2)); got != rgba(accent) {
		t.Fatalf("panel background was covered: %v, panel=%v", got, r)
	}
}

func TestSubmittedSelectMultipleHoverAndOutsideAction(t *testing.T) {
	state := submittedSelect([]nui.Option[int]{{Value: 0, Label: "Zero"}})
	var selected nui.SelectionState[int]
	actions := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: true})
		nui.SelectMultiple(c.Key("choices"), th, &selected, state, nui.SelectOptions{Label: "Choices", Popup: nui.PopoverOptions{Mode: nui.PopoverHover, OpenDelay: 100 * time.Millisecond}}).Width(200)
		nui.Button(c.Key("outside"), th, "Outside action", nui.ButtonOptions{}).Absolute().Top(250).OnClick(func() { actions++ })
	}, 320, 300)
	r := bounds(t, tt, "Choices")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(50 * time.Millisecond)
	if state.Open {
		t.Fatal("hover ignored OpenDelay")
	}
	tt.Advance(60 * time.Millisecond)
	if !state.Open || !tt.HasText("Zero") {
		t.Fatal("hover did not open submitted options")
	}
	tt.Key(0, ui.KeyEscape)
	if state.Open {
		t.Fatal("Escape did not suppress hover reopening")
	}
	click(t, tt, "Choices")
	click(t, tt, "Outside action")
	if state.Open || actions != 1 {
		t.Fatal("outside dismissal swallowed the target action")
	}
}

func TestSubmittedSelectMultipleTypedBulkAndExternalUpdates(t *testing.T) {
	type key struct{ Number int }
	options := []nui.Option[key]{{Value: key{0}, Label: "Zero"}, {Value: key{1}, Label: "Blocked", Disabled: true}, {Value: key{2}, Label: "Two"}}
	var state nui.SelectState[key]
	var selected nui.SelectionState[key]
	state.SetOptions(options)
	selected.Add(key{0})
	open, disabled, changes := false, false, 0
	state.Open = true // The explicit Open binding takes precedence.
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: true})
		ui.Column(c).Disabled(disabled).Children(func() {
			nui.SelectMultiple(c.Key("choices"), th, &selected, &state, nui.SelectOptions{Label: "Choices", Open: &open, OnChange: func() { changes++ }}).Width(260)
		})
	}, 320, 300)
	if state.Open || tt.HasText("Blocked") {
		t.Fatal("internal opening overrode the explicit binding")
	}
	click(t, tt, "Choices")
	tt.Key(ui.Ctrl, ui.KeyA)
	if !slices.Equal(selected.Values(), []key{{0}, {2}}) || changes != 1 || !open {
		t.Fatalf("select all: values=%v, changes=%d, open=%v", selected.Values(), changes, open)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	if changes != 1 {
		t.Fatal("unchanged Select All notified")
	}
	tt.Key(0, ui.KeySpace) // Select All leaves the last enabled value active.
	if !slices.Equal(selected.Values(), []key{{0}}) || changes != 2 {
		t.Fatal("typed keyboard toggle lost the zero key or active item")
	}
	disabled = true
	tt.Frame()
	tt.Key(ui.Ctrl, ui.KeyA)
	if selected.Len() != 1 || changes != 2 {
		t.Fatal("inherited disabling permitted bulk selection")
	}
	disabled = false
	selected.Set([]key{{2}})
	options[2].Label = "Changed two"
	state.SetOptions(options)
	open = false
	tt.Frame()
	if changes != 2 || !tt.HasText("Changed two") || !tt.Focused("Choices") {
		t.Fatal("external data/selection changed notification or failed to refresh the label/focus")
	}
	options[2].Label = "Unsubmitted edit"
	tt.Frame()
	if !tt.HasText("Changed two") || tt.HasText("Unsubmitted edit") {
		t.Fatal("source mutation bypassed submission")
	}
}

func TestSubmittedSelectMultipleVirtualIdentity(t *testing.T) {
	const count = 10000
	options := make([]nui.Option[int], count)
	for i := range options {
		options[i] = nui.Option[int]{Value: i, Label: fmt.Sprintf("Option %05d", i)}
	}
	var state nui.SelectState[int]
	var selected nui.SelectionState[int]
	state.SetOptions(options)
	rows, changes := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		rows = 0
		nui.SelectMultiple(c.Key("choices"), nui.App(c, nui.AppConfig{DisableAnimations: true}), &selected, &state, nui.SelectOptions{Label: "Choices", OnChange: func() { changes++ }, Decoration: nui.Decoration{Style: func(slot nui.Slot, _ ui.Element) {
			if slot == nui.Item {
				rows++
			}
		}}}).Width(280)
	}, 340, 320)
	click(t, tt, "Choices")
	tt.Key(0, ui.KeyEnd)
	if rows > 24 || !tt.HasText("Option 09999") {
		t.Fatalf("end navigation built %d rows without exposing the last choice", rows)
	}
	slices.Reverse(options)
	state.SetOptions(options)
	tt.Frame()
	if rows > 24 || !tt.HasText("Option 09999") {
		t.Fatal("resubmission did not scroll to the surviving active identity")
	}
	tt.Key(0, ui.KeyEnter)
	if !selected.Has(count-1) || changes != 1 {
		t.Fatal("reorder selected an index rather than a typed key")
	}
	state.SetOptions(nil)
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if selected.Len() != 1 || changes != 1 || rows != 0 {
		t.Fatal("empty submission retained selectable rows or changed selection")
	}
	tt.Key(0, ui.KeyEscape)
	if state.Open || !tt.Focused("Choices") {
		t.Fatal("empty popup did not return focus to its trigger")
	}
}

func TestInputMenuDefaultClickAndExplicitOpening(t *testing.T) {
	var state nui.SearchState
	openOnClick := nui.Optional[bool]{}
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query", OpenOnClick: openOnClick}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	if tt.HasText("Alpha") {
		t.Fatal("default click unexpectedly opened InputMenu")
	}
	tt.Key(0, ui.KeyDown)
	if !tt.HasText("Alpha") {
		t.Fatal("Down did not open InputMenu")
	}
	tt.Key(0, ui.KeyEscape)
	openOnClick = nui.Set(true)
	tt.Frame()
	click(t, tt, "Query")
	if !tt.HasText("Alpha") {
		t.Fatal("explicit openOnClick failed")
	}
}

func TestInputMenuControlledOpenLoopAndHover(t *testing.T) {
	state, open := nui.SearchState{}, false
	loop, hover := false, false
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "x", Label: "Blocked", Disabled: true}, {Value: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query", Open: &open, Loop: loop, HighlightOnHover: nui.Set(hover)}).Width(260)
	}, 360, 320)
	open = true
	tt.Frame()
	if !open || !tt.Focused("Query") || !tt.HasText("Beta") {
		t.Fatal("controlled open did not focus and show input")
	}
	r := bounds(t, tt, "Beta")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "a" || open {
		t.Fatal("hover=false changed highlight")
	}
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" {
		t.Fatal("loop=false did not stop at last enabled item")
	}
	loop = true
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "a" {
		t.Fatal("loop did not wrap past disabled item")
	}
	hover = true
	open = true
	tt.Frame()
	r = bounds(t, tt, "Beta")
	tt.Move(r.X+2, r.Y+2)
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "b" {
		t.Fatal("hover=true failed to move highlight")
	}
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if open || !tt.Focused("Query") {
		t.Fatal("Escape did not update open state or retain input focus")
	}
}

func TestInputMenuClearResetPolicyAndReadOnly(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(fmt.Sprint(multiple), func(t *testing.T) {
			state := nui.SearchState{Selected: "a", Query: "Draft"}
			state.SelectedValues.Set([]string{"a"})
			reset, readonly, clears, changes := false, false, 0, 0
			state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
			tt := ui.NewTester(func(c *ui.Context) {
				nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query", Multiple: multiple, Clearable: true, ReadOnly: readonly, ResetModelValueOnClear: nui.Set(reset), OnClear: func() { clears++ }, OnValuesChange: func() { changes++ }}).Width(280)
			}, 380, 300)
			click(t, tt, "Clear Query")
			if state.Selected != "a" || !slices.Equal(state.SelectedValues.Values(), []string{"a"}) || state.Query != "" || clears != 1 || changes != 0 {
				t.Fatalf("retained clear: %+v %d/%d", state, clears, changes)
			}
			reset = true
			readonly = true
			tt.Frame()
			if _, ok := tt.Find("Clear Query"); ok {
				t.Fatal("read-only model exposed clear")
			}
			readonly = false
			tt.Frame()
			click(t, tt, "Clear Query")
			if state.Selected != "" || multiple && state.SelectedValues.Len() != 0 || clears != 2 || multiple && changes != 1 {
				t.Fatalf("reset clear: %+v %d/%d", state, clears, changes)
			}
		})
	}
}

func TestInputMenuIconsAndLoadingDoNotChangeState(t *testing.T) {
	state := nui.SearchState{}
	state.SelectedValues.Set([]string{"a"})
	queries, values, selected := 0, 0, 0
	loader := false
	marker := ui.MustParseSVG([]byte(`<svg viewBox="0 0 10 10"><rect width="10" height="10"/></svg>`))
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First choice"}, {Value: "b", Label: "Beta", Description: "Second choice"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		nui.InputMenu(c, th, &state, nui.InputMenuOptions{Label: "Query", Multiple: true, Clearable: true, SelectedIcon: marker, DeleteIcon: marker, ClearIcon: marker, Avatar: &nui.AvatarOptions{Text: "AL", Alt: "Query avatar"}, Loading: loader, LoadingTrailing: true, LoadingIcon: lucide.Heart(), OnChange: func(string) { queries++ }, OnSelect: func(string) { selected++ }, OnValuesChange: func() { values++ }, UI: nui.Styles{nui.Item: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}}).Width(280)
	}, 380, 330)
	delete := bounds(t, tt, "Remove Alpha")
	if color := tt.Image().RGBAAt(int(delete.X+1), int(delete.Y+1)); color != rgba(nui.Light().Dimmed) {
		t.Fatalf("custom square delete icon was not used: %v", color)
	}
	clear := bounds(t, tt, "Clear Query")
	if color := tt.Image().RGBAAt(int(clear.X+1), int(clear.Y+1)); color != rgba(nui.Light().Muted) {
		t.Fatalf("custom square clear icon was not used: %v", color)
	}
	click(t, tt, "Query")
	loader = true
	tt.Frame()
	tt.Type("Be")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(state.SelectedValues.Values(), []string{"a", "b"}) || values != 1 || selected != 1 || queries != 1 || !tt.HasText("Query avatar") {
		t.Fatalf("loading select state=%+v values/select/query=%d/%d/%d", state, values, selected, queries)
	}
	click(t, tt, "Remove Alpha")
	if !slices.Equal(state.SelectedValues.Values(), []string{"b"}) || values != 2 || selected != 1 || !tt.Focused("Query") {
		t.Fatal("tag removal changed focus or selection event")
	}
	click(t, tt, "Clear Query")
	if state.SelectedValues.Len() != 0 || values != 3 || queries != 1 {
		t.Fatal("model-only clear reported an input edit")
	}
}

func TestInputMenuExplicitDismissalAndDisabledOpen(t *testing.T) {
	state, open, disabled := nui.SearchState{}, false, false
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputMenu(c, th, &state, nui.InputMenuOptions{Label: "Query", Open: &open, Disabled: disabled, Popup: nui.PopoverOptions{Dismissible: nui.Set(false)}}).Width(260)
			nui.Button(c, th, "Outside", nui.ButtonOptions{}).Absolute().Top(260)
		})
	}, 360, 320)
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if !open {
		t.Fatal("Dismissible=false allowed Escape")
	}
	click(t, tt, "Outside")
	if !open || !tt.Focused("Query") {
		t.Fatal("Dismissible=false allowed an outside focus to close the list")
	}
	open = false
	tt.Frame()
	if open || tt.HasText("Alpha") {
		t.Fatal("explicit close was blocked")
	}
	disabled = true
	open = true
	tt.Frame()
	if open || tt.HasText("Alpha") {
		t.Fatal("disabled input accepted an external open")
	}
}

func TestInputMenuClearSelectedTagsWithoutDraft(t *testing.T) {
	state := nui.SearchState{}
	state.SelectedValues.Set([]string{"a", "b"})
	clears, changes, queries := 0, 0, 0
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query", Multiple: true, Clearable: true, OnClear: func() { clears++ }, OnValuesChange: func() { changes++ }, OnChange: func(string) { queries++ }}).Width(300)
	}, 400, 300)
	if _, ok := tt.Find("Clear Query"); !ok {
		t.Fatal("selected tags with an empty draft cannot be cleared")
	}
	click(t, tt, "Clear Query")
	if state.SelectedValues.Len() != 0 || clears != 1 || changes != 1 || queries != 0 || !tt.Focused("Query") {
		t.Fatalf("clear result: %+v callbacks %d/%d/%d", state, clears, changes, queries)
	}
	tt.Frame()
	if clears != 1 || changes != 1 {
		t.Fatal("clear callback repeated")
	}
}

func TestInputMenuAutocompleteIgnoresSelectionReset(t *testing.T) {
	state := nui.SearchState{}
	selected, submitted := 0, 0
	state.SetOptions([]nui.Option[string]{{Value: "alpha", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query", Mode: nui.InputMenuAutocomplete, Multiple: true, ResetSearchTermOnSelect: nui.Set(false), ResetSearchTermOnBlur: nui.Set(true), OnSelect: func(string) { selected++ }, OnSubmit: func(string) { submitted++ }}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	tt.Type("Al")
	tt.Key(0, ui.KeyEnter)
	if state.Query != "Alpha" || state.Selected != "alpha" || selected != 1 || submitted != 0 || state.SelectedValues.Len() != 0 {
		t.Fatalf("autocomplete suggestion: %+v %d/%d", state, selected, submitted)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Free text")
	tt.Key(0, ui.KeyEnter)
	if state.Query != "Free text" || selected != 1 || submitted != 1 {
		t.Fatal("autocomplete free submit changed the suggestion or duplicated events")
	}
}

func TestInputMenuDisabledResultsAreNotEmpty(t *testing.T) {
	var state nui.SearchState
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Unavailable", Disabled: true}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query"}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	tt.Type("Unavailable")
	if !tt.HasText("Unavailable") || tt.HasText("No matching data") {
		t.Fatal("visible disabled result was reported as no results")
	}
	tt.Key(0, ui.KeyEnter)
	if state.Selected != "" {
		t.Fatal("disabled result was selected")
	}
}

func TestInputMenuControlledSelectionPreservesActiveDraft(t *testing.T) {
	state := nui.SearchState{Selected: "a"}
	changes := 0
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputMenu(c, th, &state, nui.InputMenuOptions{Label: "Query", OnChange: func(string) { changes++ }}).Width(260)
			nui.Button(c, th, "Outside", nui.ButtonOptions{}).Absolute().Top(260)
		})
	}, 360, 320)
	if state.Query != "Alpha" {
		t.Fatalf("initial selection display=%q", state.Query)
	}
	state.Selected = "b"
	tt.Frame()
	if state.Query != "Beta" || changes != 0 {
		t.Fatal("external selected value did not update idle display silently")
	}
	click(t, tt, "Query")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Draft")
	state.Selected = "a"
	tt.Frame()
	if state.Query != "Draft" || changes != 1 {
		t.Fatal("external value discarded the active draft")
	}
	click(t, tt, "Outside")
	if state.Query != "Alpha" || changes != 1 {
		t.Fatal("blur did not restore the current selected label silently")
	}
}

func TestInputMenuSelectionReportsBothCallbacksOnce(t *testing.T) {
	state := nui.SearchState{}
	var events []string
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query", Multiple: true, OnSelect: func(id string) { events = append(events, "select:"+id) }, OnValuesChange: func() { events = append(events, "values") }}).Width(260)
	}, 360, 300)
	click(t, tt, "Query")
	tt.Type("Al")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !slices.Equal(events, []string{"select:a", "values"}) {
		t.Fatalf("selection events=%v", events)
	}
}

func TestInputMenuKeyboardTagDeleteReturnsToEditor(t *testing.T) {
	state := nui.SearchState{}
	state.SelectedValues.Set([]string{"a", "b"})
	changes := 0
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c, nui.App(c, nui.AppConfig{}), &state, nui.InputMenuOptions{Label: "Query", Multiple: true, OnValuesChange: func() { changes++ }}).Width(300)
	}, 400, 300)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Remove Alpha") {
		t.Fatal("tag delete is not reachable by keyboard")
	}
	tt.Key(0, ui.KeySpace)
	if !slices.Equal(state.SelectedValues.Values(), []string{"b"}) || changes != 1 || !tt.Focused("Query") {
		t.Fatal("keyboard deletion lost editor focus or notified twice")
	}
}

func TestInputMenuBoundedTagsKeyboardAndDraft(t *testing.T) {
	const count = 10000
	ids := make([]string, count)
	options := make([]nui.Option[string], count)
	for i := range ids {
		ids[i] = fmt.Sprintf("id-%05d", i)
		options[i] = nui.Option[string]{Value: ids[i], Label: fmt.Sprintf("Tag %05d", i)}
	}
	state := nui.SearchState{Query: "draft"}
	state.SetOptions(options)
	state.SelectedValues.Set(ids)
	changes, activations := 0, 0
	open, disabled, readOnly := false, false, false
	scale := float32(1)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: true})
		th.TextScale = scale
		ui.Column(c).Padding(16).Disabled(disabled).Children(func() {
			nui.InputMenu(c.Key("choices"), th, &state, nui.InputMenuOptions{Label: "Query", Multiple: true, ReadOnly: readOnly, Open: &open, ExternalFilter: true, OnValuesChange: func() { changes++ }, OnSelect: func(string) { activations++ }}).Width(320)
		})
	}, 400, 420)
	bounded := func() {
		t.Helper()
		built := 0
		for _, text := range tt.Texts() {
			if strings.HasPrefix(text, "Remove Tag ") {
				built++
			}
		}
		if built == 0 || built > 16 {
			t.Fatalf("built %d tag removal controls for %d selections", built, state.SelectedValues.Len())
		}
		viewport, editor := bounds(t, tt, "Query selected options"), bounds(t, tt, "Query")
		if viewport.H <= 0 || viewport.H > 96*scale || editor.Y < viewport.Y+viewport.H || editor.H <= 0 {
			t.Fatalf("viewport/editor geometry: %v, %v at scale %g", viewport, editor, scale)
		}
	}
	bounded()
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Remove Tag 00000") {
		t.Fatal("first tag is not a keyboard stop")
	}
	tt.Key(0, ui.KeyEnd)
	if !tt.Focused("Remove Tag 09999") || state.Query != "draft" || changes != 0 {
		t.Fatal("offscreen End lost tag focus or changed the draft/selection")
	}
	bounded()
	tt.Key(0, ui.KeyDelete)
	if !tt.Focused("Query") || state.Query != "draft" || state.SelectedValues.Len() != count-1 || changes != 1 || activations != 0 {
		t.Fatalf("removal lost editor/draft or notified incorrectly: query=%q size=%d changes=%d activations=%d", state.Query, state.SelectedValues.Len(), changes, activations)
	}
	tt.Key(ui.Shift, ui.KeyTab)
	if !tt.Focused("Remove Tag 09998") {
		t.Fatal("reverse Tab did not reveal the last tag")
	}
	tt.Key(0, ui.KeyHome)
	if !tt.Focused("Remove Tag 00000") {
		t.Fatal("Home did not reveal the first tag")
	}
	readOnly = true
	tt.Frame()
	tt.Key(0, ui.KeyDelete)
	if changes != 1 {
		t.Fatal("read-only tag removal changed the selection")
	}
	readOnly, disabled = false, true
	tt.Frame()
	if err := tt.Click("Remove Tag 00000"); err != nil {
		t.Fatal(err)
	}
	if changes != 1 {
		t.Fatal("inherited disabling allowed removal")
	}
	disabled, scale = false, 2
	tt.Frame()
	bounded()
	state.SetOptions(nil)
	tt.Frame()
	if state.Query != "draft" || !tt.HasText("Tag 00000") || changes != 1 {
		t.Fatal("empty submission lost the draft or selected labels")
	}
	state.SelectedValues.Set([]string{ids[0]})
	tt.Frame()
	if viewport := bounds(t, tt, "Query selected options"); viewport.H >= 96*scale {
		t.Fatal("a short selection did not shrink its viewport")
	}
	state.SelectedValues.Clear()
	tt.Frame()
	if _, found := tt.Find("Query selected options"); found || state.Query != "draft" || changes != 1 {
		t.Fatal("external clear retained tags or changed the draft/notifications")
	}
}

func TestInputMenuTagFullLabelAndResubmission(t *testing.T) {
	label := strings.TrimSpace(strings.Repeat("A complete selected label with full wrapping. ", 12))
	state := nui.SearchState{Query: "uncommitted draft"}
	state.SelectedValues.Add("a")
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: label}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.InputMenu(c.Key("choices"), nui.App(c, nui.AppConfig{DisableAnimations: true}), &state, nui.InputMenuOptions{Label: "Query", Multiple: true}).Width(180)
	}, 240, 300)
	if !tt.HasText(label) || bounds(t, tt, label).H <= 16 {
		t.Fatal("long selected label was truncated instead of wrapping")
	}
	tt.Key(0, ui.KeyTab)
	viewport, removal := bounds(t, tt, "Query selected options"), bounds(t, tt, "Remove "+label)
	if !tt.Focused("Remove "+label) || removal.H <= 0 || removal.Y < viewport.Y || removal.Y+removal.H > viewport.Y+viewport.H {
		t.Fatalf("long-label removal control is inaccessible: viewport=%v removal=%v", viewport, removal)
	}
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Renamed"}})
	tt.Frame()
	if !tt.HasText("Renamed") || state.Query != "uncommitted draft" {
		t.Fatal("resubmission failed to refresh the label while keeping the draft")
	}
}

func TestSelectMenuCreateExcludedFromBulkSelection(t *testing.T) {
	for _, top := range []bool{false, true} {
		for _, key := range []struct {
			name string
			mods ui.Modifiers
			key  ui.Key
		}{{"Ctrl+A", ui.Ctrl, ui.KeyA}, {"Cmd+A", ui.Super, ui.KeyA}, {"Shift+End", ui.Shift, ui.KeyEnd}} {
			t.Run(key.name+map[bool]string{false: "/bottom", true: "/top"}[top], func(t *testing.T) {
				var values, notified []string
				s := nui.SelectMenuState{Open: true, Query: "Al"}
				changes, selects, creates := 0, 0, 0
				// A business ID resembling the internal action ID must remain selectable.
				items := []nui.ListboxItem{{ID: "\x00nui:create", Label: "Alpha"}, {ID: "b", Label: "Alpine"}, {ID: "c", Label: "Albatross", Disabled: true}}
				listSelection := chosenSelection(values)
				s.SetItems(items)
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					nui.SelectMenu(c, th, listSelection, &s, nui.SelectMenuOptions{
						Label: "Letters", Multiple: true, SelectionBehavior: nui.SelectionReplace, CreateAtTop: top, CreateWhenAlways: true,
						OnChange: func() { ids := listSelection.Values(); changes++; notified = slices.Clone(ids) },
						OnSelect: func(string) { selects++ }, OnCreate: func(string) { creates++ },
					}).Width(260)
				}, 400, 400)
				tt.Advance(time.Second)
				tt.Key(0, ui.KeyDown)
				if !tt.Focused("Letters options") {
					t.Fatal("the list did not receive focus")
				}
				tt.Key(0, ui.KeyHome)
				tt.Key(key.mods, key.key)
				want := []string{"\x00nui:create", "b"}
				if !slices.Equal(listSelection.Values(), want) || !slices.Equal(notified, want) || changes != 1 || selects != 0 || creates != 0 {
					t.Fatalf("bulk selection: values=%q notified=%q changes=%d selects=%d creates=%d", listSelection.Values(), notified, changes, selects, creates)
				}
				// The action remains reachable and activates once without changing selection.
				end := ui.KeyEnd
				if top {
					end = ui.KeyHome
				}
				tt.Key(0, end)
				for _, activate := range []ui.Key{ui.KeyEnter, ui.KeySpace} {
					tt.Key(0, activate)
				}
				if creates != 2 || changes != 1 || selects != 0 || !slices.Equal(listSelection.Values(), want) || !s.Open || s.Query != "Al" {
					t.Fatalf("create activation: values=%q changes=%d selects=%d creates=%d open=%v query=%q", listSelection.Values(), changes, selects, creates, s.Open, s.Query)
				}
			})
		}
	}
}

func TestSelectMenuCreateKeepsBusinessBuilderSeparate(t *testing.T) {
	var values []string
	s := nui.SelectMenuState{Open: true, Query: "Al"}
	var unexpected []string
	created := ""
	listSelection := chosenSelection(values)
	s.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, listSelection, &s, nui.SelectMenuOptions{
			Label: "Letters", CreateWhenAlways: true, OnCreate: func(query string) { created = query },
			Build: func(c *ui.Context, th nui.Theme, it nui.ListboxItem) {
				if it.ID != "a" {
					unexpected = append(unexpected, it.ID)
					return
				}
				ui.Text(c, "Custom Alpha")
			},
		}).Width(260)
	}, 400, 400)
	tt.Advance(time.Second)
	if len(unexpected) != 0 || !tt.HasText("Custom Alpha") {
		t.Fatalf("builder received %q; texts=%v", unexpected, tt.Texts())
	}
	click(t, tt, "Create “Al”")
	if created != "Al" || listSelection.Len() != 0 || !s.Open || s.Query != "Al" {
		t.Fatalf("create: query=%q created=%q values=%q open=%v", s.Query, created, listSelection.Values(), s.Open)
	}

	// Rebuild with only the action row, then restore the business row.
	s.Query = "New"
	tt.Frame()
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if created != "New" || len(unexpected) != 0 || listSelection.Len() != 0 {
		t.Fatalf("unmatched create: created=%q builder IDs=%q values=%q", created, unexpected, listSelection.Values())
	}
	s.Query = "Al"
	tt.Frame()
	click(t, tt, "Custom Alpha")
	if !slices.Equal(listSelection.Values(), []string{"a"}) || s.Open {
		t.Fatalf("business row no longer selects: values=%q open=%v", listSelection.Values(), s.Open)
	}
}

func TestSelectMenuClearAndSearchSubmit(t *testing.T) {
	state := nui.SelectMenuState{}
	values := []string{"b"}
	reset, search, focusSearch, focusElsewhere := true, true, true, false
	changes, clears := 0, 0
	var outside ui.Handle
	listSelection := chosenSelection(values)
	state.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Button(c.Key("outside"), th, "Outside target", nui.ButtonOptions{}).Bind(&outside)
			nui.SelectMenu(c.Key("menu"), th, listSelection, &state, nui.SelectMenuOptions{
				Label: "Menu", Clearable: true, ResetModelValueOnClear: nui.Set(reset), SearchInput: nui.Set(search), FocusSearch: nui.Set(focusSearch),
				ResetSearchTermOnBlur: nui.Set(false), FilterInput: &nui.InputOptions{Label: "Filter choices"},
				OnChange: func() { changes++ }, OnClear: func() { clears++ },
				OnSelect: func(string) {
					if focusElsewhere {
						outside.Focus(c)
					}
				},
			}).Width(300)
		})
	}, 450, 400)
	click(t, tt, "Menu")
	tt.Type("alp")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(listSelection.Values(), []string{"a"}) || changes != 1 || state.Open || state.Query != "" {
		t.Fatalf("search Enter: values=%v changes=%d open=%v query=%q", listSelection.Values(), changes, state.Open, state.Query)
	}
	reset = false
	state.Query = "draft"
	tt.Frame()
	click(t, tt, "Clear selection")
	if !slices.Equal(listSelection.Values(), []string{"a"}) || state.Query != "" || changes != 1 || clears != 1 || !tt.Focused("Menu") {
		t.Fatal("clear without model reset changed the value, kept the query, or lost focus")
	}
	reset = true
	tt.Frame()
	click(t, tt, "Clear selection")
	if listSelection.Len() != 0 || changes != 2 || clears != 2 || !tt.Focused("Menu") {
		t.Fatal("clear did not reset once and return trigger focus")
	}
	search, focusSearch = false, false
	state.Open, state.Query = true, "bet"
	tt.Frame()
	if tt.HasText("Alpha") || !tt.HasText("Beta") || !tt.Focused("Menu options") {
		t.Fatal("hidden search ignored its controlled query or initial list focus")
	}
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(listSelection.Values(), []string{"b"}) || changes != 3 || state.Open {
		t.Fatal("searchless selection did not commit")
	}
	search, focusElsewhere = true, true
	state.Open = true
	tt.Frame()
	if !tt.Focused("Menu options") {
		t.Fatal("FocusSearch=false did not focus the list")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Advance(time.Second)
	if state.Open || !tt.Focused("Outside target") || changes != 3 {
		t.Fatal("closing a reselected value discarded application focus or emitted a value change")
	}
}

func TestInputMenuTrailingStylesAndToggle(t *testing.T) {
	var state nui.SearchState
	var instance, callback nui.Optional[ui.Color]
	readOnly, disabled := false, false
	valid := false
	red, green, blue := ui.RGB(240, 0, 0), ui.RGB(0, 180, 0), ui.RGB(0, 0, 240)
	state.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode}).WithComponent(nui.ComponentInputMenu, nui.ComponentTheme{UI: nui.Styles{nui.Trailing: {Foreground: nui.Set(red)}}})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputMenu(c, th, &state, nui.InputMenuOptions{
				Label: "Letters", Trailing: lucide.Search(), Clearable: true,
				ReadOnly: readOnly, Disabled: disabled,
				UI: nui.Styles{nui.Trailing: {Foreground: instance}},
				Style: func(root, editor, leading, trailing ui.Element) {
					valid = trailing.Valid()
					trailing.Label("Styled trailing")
					if callback.Set {
						trailing.TextColor(callback.Value)
					}
				},
			}).Width(260)
		})
	}, 400, 300)
	check := func(want ui.Color) {
		t.Helper()
		if !valid {
			t.Fatal("Style received an empty trailing element")
		}
		if countColor(tt.Image(), bounds(t, tt, "Styled trailing"), rgba(want)) == 0 {
			t.Fatalf("trailing icon does not use %v", want)
		}
	}
	check(red)
	instance = nui.Set(green)
	tt.Frame()
	check(green)
	callback = nui.Set(blue)
	tt.Frame()
	check(blue)
	click(t, tt, "Styled trailing")
	if !tt.HasText("Alpha") || !tt.Focused("Letters") {
		t.Fatal("styled trailing icon did not open the list and focus the editor")
	}
	click(t, tt, "Styled trailing")
	if tt.HasText("Alpha") {
		t.Fatal("styled trailing icon did not close the list")
	}
	tt.Type("Al")
	if _, shown := tt.Find("Styled trailing"); shown {
		t.Fatal("the trailing icon should give way to the clear button")
	}
	click(t, tt, "Clear Letters")
	if state.Query != "" {
		t.Fatalf("clear left query %q", state.Query)
	}
	check(blue)
	click(t, tt, "Styled trailing")
	if !tt.HasText("Alpha") || !tt.Focused("Letters") {
		t.Fatal("the trailing icon did not restore its behavior after clearing")
	}
	for _, mode := range []string{"readonly", "disabled"} {
		readOnly, disabled = mode == "readonly", mode == "disabled"
		tt.Frame()
		click(t, tt, "Styled trailing")
		if tt.HasText("Alpha") {
			t.Fatalf("%s trailing icon opened the list", mode)
		}
	}
}

func TestOverlayPanelOpacityOverrides(t *testing.T) {
	for _, kind := range []string{"popover", "menu", "tooltip", "modal"} {
		for _, source := range []string{"theme", "instance", "callback", "zero"} {
			t.Run(kind+"/"+source, func(t *testing.T) {
				open, off := true, false
				menu := nui.MenuState{Open: true}
				want := float32(.25)
				if source == "zero" {
					want = 0
				}
				menu.SetItems([]nui.MenuItem{{ID: "a", Label: "Action"}})
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, DisableAnimations: off})
					component := map[string]nui.Component{"popover": nui.ComponentPopover, "menu": nui.ComponentDropdownMenu, "tooltip": nui.ComponentTooltip, "modal": nui.ComponentModal}[kind]
					th = th.WithComponent(component, nui.ComponentTheme{UI: nui.Styles{nui.Panel: {Background: nui.Set(ui.RGB(255, 0, 0)), Opacity: nui.Set(float32(.25)), Radius: nui.Set(float32(0)), BorderWidth: nui.Set(float32(0)), Shadow: nui.Set(nui.ShadowTokens{})}}})
					d := nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Panel {
							e.Label("Panel")
							if source == "callback" {
								e.Opacity(want)
							}
						}
					}}
					if source != "theme" {
						th = th.WithComponent(component, nui.ComponentTheme{UI: nui.Styles{nui.Panel: {Opacity: nui.Set(float32(.75))}}})
						d.UI = nui.Styles{nui.Panel: {Opacity: nui.Set(want)}}
						if source == "callback" {
							d.UI[nui.Panel] = nui.PartStyle{Opacity: nui.Set(float32(.5))}
						}
					}
					a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(50).Top(20)
					switch kind {
					case "popover":
						nui.Popover(c, th, a, &open, nui.PopoverOptions{Decoration: d, Width: nui.Set(float32(200))}, func(nui.Theme) { ui.Text(c, "Contents") })
					case "menu":
						nui.DropdownMenu(c, th, a, &menu, nui.MenuOptions{Decoration: d})
					case "tooltip":
						nui.Tooltip(c, th, a, "Contents", nui.TooltipOptions{Open: &open, Decoration: d})
					case "modal":
						nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Panel", Overlay: nui.Set(false), Decoration: d}, func(nui.Theme) { ui.Text(c, "Contents") })
					}
				}, 600, 400)
				pixel := func() uint8 {
					r := bounds(t, tt, "Panel")
					return tt.Image().RGBAAt(int(r.X+2), int(r.Y+r.H/2)).G
				}
				first := pixel()
				tt.Advance(50 * time.Millisecond)
				middle := pixel()
				tt.Advance(time.Second)
				last := pixel()
				expected := int(255 * (1 - want))
				if difference := int(last) - expected; difference < -2 || difference > 2 {
					t.Fatalf("final opacity %v: green=%d, want about %d", want, last, expected)
				}
				if source != "callback" && want > 0 && (first <= middle || middle <= last) {
					t.Fatalf("opacity override should still fade in: green=%d, %d, %d", first, middle, last)
				}
				open, menu.Open = false, false
				tt.Frame()
				if _, visible := tt.Find("Panel"); visible {
					t.Fatal("closed panel is still visible")
				}
				off = true
				open, menu.Open = true, true
				tt.Frame()
				if got := pixel(); got != last {
					t.Fatalf("disabled animations changed explicit opacity: %d -> %d", last, got)
				}
			})
		}
	}
}

func TestSelectionStateOrderAndRevisions(t *testing.T) {
	type key int
	var s nui.SelectionState[key]
	if s.Len() != 0 || s.Revision() != 0 || s.Has(0) || s.Clear() || s.Set(nil) {
		t.Fatal("zero selection is not an unchanged empty set")
	}
	if !s.Set([]key{2, 0, 2, 1, 0}) || !slices.Equal(s.Values(), []key{2, 0, 1}) || s.Revision() != 1 {
		t.Fatal("ordered deduplication lost a typed zero or its first position")
	}
	if s.Add(0) || s.Remove(4) || s.Set([]key{2, 0, 1, 2}) || s.Revision() != 1 {
		t.Fatal("no-op mutation advanced the revision")
	}
	if !s.Remove(0) || !s.Add(0) || !slices.Equal(s.Values(), []key{2, 1, 0}) || s.Revision() != 3 {
		t.Fatal("remove/add did not preserve ordered membership")
	}
	if !s.Set([]key{0, 1, 2}) || s.Revision() != 4 {
		t.Fatal("ordered replacement was not a change")
	}
	if !s.Clear() || s.Clear() || s.Revision() != 5 || s.Len() != 0 {
		t.Fatal("clear revision or contents")
	}
	var strings nui.SelectionState[string]
	if !strings.Add("") || !strings.Has("") {
		t.Fatal("selection state reserved a valid typed zero")
	}
}

func TestSelectionStateOwnsSnapshots(t *testing.T) {
	values := []string{"a", "b"}
	var s nui.SelectionState[string]
	s.Set(values)
	values[0] = "source changed"
	snapshot := s.Values()
	snapshot[1] = "snapshot changed"
	if !slices.Equal(s.Values(), []string{"a", "b"}) || s.Has("source changed") || s.Has("snapshot changed") {
		t.Fatal("a writable alias escaped selection ownership")
	}
	s.Remove("a")
	if snapshot[0] != "a" {
		t.Fatal("a later mutation rewrote an earlier snapshot")
	}
}

func TestSelectionStateRejectsNonreflexiveKeysAtomically(t *testing.T) {
	var s nui.SelectionState[float64]
	s.Set([]float64{1, 2})
	for _, mutate := range []func(){
		func() { s.Set([]float64{3, math.NaN(), 4}) },
		func() { s.Add(math.NaN()) },
		func() { s.Remove(math.NaN()) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("non-reflexive key was accepted")
				}
			}()
			mutate()
		}()
		if !slices.Equal(s.Values(), []float64{1, 2}) || s.Revision() != 1 {
			t.Fatal("invalid mutation damaged the previous snapshot")
		}
	}
}
