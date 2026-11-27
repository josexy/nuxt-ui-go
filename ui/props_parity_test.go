package nui_test

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestParitySliderKeyboardDirections(t *testing.T) {
	// Independent expected directions from Reka's horizontal/vertical BACK_KEYS.
	cases := []struct {
		vertical, inverted, rtl bool
		deltas                  [4]float64 // Left, Right, Up, Down
	}{
		{false, false, false, [4]float64{-1, 1, 1, -1}},
		{false, true, false, [4]float64{1, -1, 1, -1}},
		{false, false, true, [4]float64{1, -1, 1, -1}},
		{false, true, true, [4]float64{-1, 1, 1, -1}},
		{true, false, false, [4]float64{-1, 1, 1, -1}},
		{true, true, false, [4]float64{-1, 1, -1, 1}},
		{true, false, true, [4]float64{-1, 1, 1, -1}},
		{true, true, true, [4]float64{-1, 1, -1, 1}},
	}
	for _, tc := range cases {
		for _, multiple := range []bool{false, true} {
			t.Run(fmt.Sprintf("vertical=%v/inverted=%v/rtl=%v/range=%v", tc.vertical, tc.inverted, tc.rtl, multiple), func(t *testing.T) {
				value, high := 40.0, 90.0
				changes, commits := 0, 0
				observers := 0
				disabled := false
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					if multiple {
						axis := nui.Horizontal
						if tc.vertical {
							axis = nui.Vertical
						}
						nui.RangeSlider(c, th, &value, &high, nui.RangeSliderOptions{Label: "Range", Step: 2, Orientation: axis, Inverted: tc.inverted, Disabled: disabled, OnChange: func(float64, float64) { changes++ }, OnCommit: func(float64, float64) { commits++ }}).OnChange(func() { observers++ })
					} else {
						nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Step: 2, Vertical: tc.vertical, Inverted: tc.inverted, Disabled: disabled, OnChange: func(float64) { changes++ }, OnCommit: func(float64) { commits++ }}).OnChange(func() { observers++ })
					}
				}, 500, 300)
				if tc.rtl {
					tt.SetDirection(ui.RightToLeft)
				}
				tt.Key(0, ui.KeyTab)
				keys := []ui.Key{ui.KeyLeft, ui.KeyRight, ui.KeyUp, ui.KeyDown}
				for i, key := range keys {
					for _, mods := range []ui.Modifiers{0, ui.Shift} {
						value = 40
						tt.Frame()
						changes, commits = 0, 0
						observers = 0
						tt.Key(mods, key)
						step := 2.0
						if mods == ui.Shift {
							step = 20
						}
						want := 40 + tc.deltas[i]*step
						if value != want || high != 90 || changes != 1 || commits != 1 {
							t.Fatalf("key=%v mods=%v value=%v want=%v high=%v changes=%d commits=%d", key, mods, value, want, high, changes, commits)
						}
						if observers != 1 {
							t.Fatalf("element OnChange observers=%d", observers)
						}
					}
				}
				disabled = true
				tt.Frame()
				before := value
				tt.Key(0, ui.KeyUp)
				tt.Key(ui.Shift, ui.KeyRight)
				tt.Frame()
				if value != before || changes != 1 || commits != 1 {
					t.Fatal("disabled slider changed or repeated callback")
				}
			})
		}
	}
}

func TestParityRangeGapCommitAndThumbTooltip(t *testing.T) {
	low, high := 20.0, 40.0
	changes, commits := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(40).Children(func() {
			nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{Label: "Range", Step: 5, MinStepsBetweenThumbs: 2, Tooltip: true, OnChange: func(float64, float64) { changes++ }, OnCommit: func(float64, float64) { commits++ }}).Width(400)
		})
	}, 500, 300)
	r := bounds(t, tt, "Range")
	x := func(v float32) float32 { return r.X + 8 + (r.W-16)*v/100 }
	y := r.Y + r.H/2
	tt.Move(x(90), y)
	tt.Advance(time.Second)
	if tt.HasText("20") || tt.HasText("40") {
		t.Fatal("track hover opened thumb tooltip")
	}
	tt.Move(x(20), y)
	tt.Advance(time.Second)
	if !tt.HasText("20") || tt.HasText("40") {
		t.Fatal("hover should show only the addressed thumb value")
	}
	tt.Press(x(20), y)
	tt.Move(x(35), y)
	if low != 20 || high != 40 || changes != 0 {
		t.Fatal("minimum gap violation should reject the move")
	}
	tt.Move(x(60), y)
	if low != 40 || high != 60 || changes != 1 || commits != 0 {
		t.Fatalf("crossing [%v,%v], changes=%d commits=%d", low, high, changes, commits)
	}
	tt.Release(x(60), y)
	if commits != 1 {
		t.Fatalf("release commits=%d", commits)
	}
	tt.Move(490, 290)
	tt.Key(0, ui.KeyRight)
	if low != 40 || high != 65 || commits != 2 {
		t.Fatal("focus did not follow crossing thumb")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if tt.HasText("65") {
		t.Fatal("Escape should dismiss thumb tooltip")
	}
}

func TestParityTagsLengthCompositionAndSelection(t *testing.T) {
	var state nui.TagsState
	var tags []string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputTags(c, th, &tags, &state, nui.InputTagsOptions{Label: "Tags", MaxLength: 4, AddOnTab: true})
			nui.Button(c, th, "Next", nui.ButtonOptions{})
		})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.Type("😀a")
	tt.Key(0, ui.KeyHome)
	tt.Type("中文")
	if state.Query != "中😀a" {
		t.Fatalf("UTF16 insertion should keep suffix: %q", state.Query)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Compose("nihao", 5)
	tt.Type("你好世界啊")
	if state.Query != "你好世界" {
		t.Fatalf("composition commit limit: %q", state.Query)
	}
	tt.Key(ui.Shift, ui.KeyTab)
	if !slices.Equal(tags, []string{"你好世界"}) || !tt.Focused("Tags") {
		t.Fatalf("Shift+Tab tags=%v focused=%v", tags, tt.Focused("Tags"))
	}
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Next") {
		t.Fatal("empty draft Tab should leave")
	}
}

func TestParityFormFieldLabelActivation(t *testing.T) {
	for _, toggle := range []bool{false, true} {
		t.Run(fmt.Sprint(toggle), func(t *testing.T) {
			checked, disabled := false, false
			optionCalls, observerCalls := 0, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.FormField(c, th, nui.FormFieldOptions{Label: "Field label", Disabled: disabled, Orientation: nui.Set(nui.Horizontal)}, func(th nui.Theme) {
					o := nui.CheckboxOptions{Label: "Control", OnChange: func(bool) { optionCalls++ }}
					if toggle {
						nui.Switch(c, th, &checked, o).OnChange(func() { observerCalls++ })
					} else {
						nui.Checkbox(c, th, &checked, o).OnChange(func() { observerCalls++ })
					}
				})
			}, 500, 300)
			// The group and visible label share a name; click the label's left edge.
			r := bounds(t, tt, "Field label")
			tt.Press(r.X+5, r.Y+8)
			tt.Release(r.X+5, r.Y+8)
			tt.Frame()
			if !checked || optionCalls != 1 || observerCalls != 1 {
				t.Fatalf("checked=%v callbacks=%d/%d", checked, optionCalls, observerCalls)
			}
			disabled = true
			tt.Frame()
			tt.Press(r.X+5, r.Y+8)
			tt.Release(r.X+5, r.Y+8)
			if !checked || optionCalls != 1 || observerCalls != 1 {
				t.Fatal("disabled label activated control")
			}
		})
	}
}

func TestParityControlledTooltipFocusAndExternalClose(t *testing.T) {
	open, second := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			a := nui.Button(c, th, "First", nui.ButtonOptions{})
			nui.Tooltip(c, th, a, "First tip", nui.TooltipOptions{Open: &open})
			b := nui.Button(c, th, "Second", nui.ButtonOptions{})
			nui.Tooltip(c, th, b, "Second tip", nui.TooltipOptions{Open: &second})
		})
	}, 500, 300)
	tt.Key(0, ui.KeyTab)
	if !open || !tt.HasText("First tip") {
		t.Fatal("keyboard focus should open controlled tooltip")
	}
	open = false
	tt.Frame()
	tt.Advance(time.Second)
	if open || tt.HasText("First tip") {
		t.Fatal("external false should stay closed while focused")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if open || !second || tt.HasText("First tip") || !tt.HasText("Second tip") {
		t.Fatal("blur and next focus should transfer tooltip ownership")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if second || tt.HasText("Second tip") {
		t.Fatal("Escape should publish closed state")
	}
}

func TestParityFormFieldCheckboxGroupLabel(t *testing.T) {
	var selected []string
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Features"}, func(th nui.Theme) {
			nui.CheckboxGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.CheckboxGroupOptions{OnChange: func() { changes++ }})
		})
	}, 500, 300)
	tt.Press(5, 8)
	tt.Release(5, 8)
	tt.Frame()
	if !slices.Equal(selected, []string{"a"}) || changes != 1 {
		t.Fatalf("label selected=%v changes=%d", selected, changes)
	}
}

func TestParityReselectCallbacks(t *testing.T) {
	selected := []string{"a"}
	var state nui.SelectMenuState
	changes, activations, queries := 0, 0, 0
	listSelection := chosenSelection(selected)
	state.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, listSelection, &state, nui.SelectMenuOptions{Label: "Menu", OnChange: func() { changes++ }, OnSelect: func(string) { activations++ }, OnQuery: func(string) { queries++ }})
	}, 500, 300)
	click(t, tt, "Menu")
	tt.Type("Al")
	queries = 0
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if state.Open || state.Query != "" || changes != 0 || activations != 1 || queries != 1 {
		t.Fatalf("open=%v query=%q callbacks=%d/%d/%d", state.Open, state.Query, changes, activations, queries)
	}
}

func TestParityCarouselRTL(t *testing.T) {
	var state nui.CarouselState
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Carousel(c, th, &state, []nui.CarouselItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Gamma"}}, nui.CarouselOptions{Label: "Guide", OnChange: func(int) { changes++ }}).Width(400)
	}, 500, 300)
	tt.SetDirection(ui.RightToLeft)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyLeft)
	if state.Index != 1 || changes != 1 {
		t.Fatalf("RTL left: index=%d changes=%d", state.Index, changes)
	}
	tt.Key(0, ui.KeyRight)
	if state.Index != 0 || changes != 2 {
		t.Fatal("RTL right should return to first")
	}
	r := bounds(t, tt, "Guide")
	tt.Press(r.X+120, r.Y+50)
	tt.Move(r.X+320, r.Y+50)
	tt.Release(r.X+320, r.Y+50)
	tt.Advance(time.Second)
	if state.Index != 1 || changes != 3 {
		t.Fatalf("RTL drag right: index=%d changes=%d", state.Index, changes)
	}
}

func TestParitySelectMultipleSlotMapping(t *testing.T) {
	single := ""
	var multiple []string
	instance := false
	options := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
	var root ui.Handle
	selectState1 := submittedSelect(options)
	selectState2 := submittedSelect(options)
	selectedState := chosenSelection(multiple)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{}).WithComponent(nui.ComponentSelect, nui.ComponentTheme{Size: nui.XL, UI: nui.Styles{nui.Root: {PaddingY: nui.Set(float32(12))}}}).WithComponent(nui.ComponentSelectMenu, nui.ComponentTheme{Size: nui.XS})
		ui.Column(c).Gap(16).Children(func() {
			d := nui.Decoration{}
			if instance {
				d.UI = nui.Styles{nui.Root: {PaddingY: nui.Set(float32(15))}}
			}
			nui.Select(c, th, &single, selectState1, nui.SelectOptions{Label: "Single", Decoration: d})
			d.Style = func(slot nui.Slot, e ui.Element) {
				if slot == nui.Root {
					e.Bind(&root)
				}
			}
			nui.SelectMultiple(c, th, selectedState, selectState2, nui.SelectOptions{Label: "Multiple", Decoration: d})
			nui.Button(c, th, "Focus trigger", nui.ButtonOptions{}).OnClick(func() { root.Focus(c) })
		})
	}, 500, 500)
	previous := float32(32)
	for range 2 {
		tt.Frame()
		a, b := bounds(t, tt, "Single"), bounds(t, tt, "Multiple")
		if a.H != b.H || b.H <= previous {
			t.Fatalf("ComponentSelect size and root padding: single=%v multiple=%v previous=%v", a.H, b.H, previous)
		}
		previous = b.H
		instance = true
	}
	click(t, tt, "Focus trigger")
	if !tt.Focused("Multiple") {
		t.Fatal("Style(Root) did not address the focusable trigger")
	}
	tt.Key(0, ui.KeyDown)
	if !tt.HasText("Alpha") || !tt.HasText("Beta") {
		t.Fatal("trigger height constrained option panel")
	}
}

func parityScreenshot(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("close screenshot %s: %v", f.Name(), err)
		}
	}()
	if err = png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

// Regressions against the pinned Nuxt UI / Reka UI behavior.
func TestParityInputMenuEscapeReset(t *testing.T) {
	var s nui.SearchState
	s.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, nui.InputMenuOptions{Label: "Search"}).Width(300)
	}, 500, 400)
	click(t, tt, "Search")
	tt.Type("Al")
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if s.Query != "" {
		t.Fatalf("Escape left Query=%q, focus=%v; expected reset", s.Query, tt.Focused("Search"))
	}
}

func TestParityInputMenuDeselectionIndicator(t *testing.T) {
	var s nui.SearchState
	s.SetOptions([]nui.Option[string]{{Value: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, nui.InputMenuOptions{Label: "Search", Multiple: true}).Width(300)
	}, 500, 400)
	click(t, tt, "Search")
	tt.Key(0, ui.KeyDown)
	click(t, tt, "Beta")
	click(t, tt, "Remove Beta")
	click(t, tt, "Search")
	tt.Key(0, ui.KeyDown)
	tt.Move(490, 390)
	tt.Advance(time.Second)
	if s.SelectedValues.Len() != 0 {
		t.Fatal(s.SelectedValues.Values())
	}
	r := bounds(t, tt, "Beta")
	green := func() int {
		n := 0
		im := tt.Image()
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			for x := int(r.X); x < int(r.X+r.W); x++ {
				rr, g, b, _ := im.At(x, y).RGBA()
				if g > rr+12000 && g > b+4000 {
					n++
				}
			}
		}
		return n
	}
	before := green()
	parityScreenshot(t, tt, "inputmenu-deselected")
	s.Selected = ""
	tt.Frame()
	after := green()
	parityScreenshot(t, tt, "inputmenu-deselected-cleared")
	if before != after {
		t.Fatalf("empty SelectedValues still had checkmark: green pixels=%d; clearing stale Selected removes it (%d)", before, after)
	}
}

func TestParityInputMenuCreateAlwaysExactMatch(t *testing.T) {
	var s nui.SearchState
	s.SetOptions([]nui.Option[string]{{Value: "Alpha", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, nui.InputMenuOptions{Label: "Search", CreateWhenAlways: true, OnCreate: func(string) {}}).Width(300)
	}, 500, 400)
	click(t, tt, "Search")
	tt.Type("Alpha")
	if tt.HasText("Create") {
		t.Fatalf("exact existing value still offers creation: %v", tt.Texts())
	}
}

func TestParitySelectMenuReselectCloses(t *testing.T) {
	s := nui.SelectMenuState{}
	selected := []string{"a"}
	listSelection := chosenSelection(selected)
	s.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, listSelection, &s, nui.SelectMenuOptions{Label: "Choose"}).Width(300)
	}, 500, 500)
	click(t, tt, "Choose")
	tt.Type("Al")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(time.Second)
	if s.Open || s.Query != "" {
		t.Fatalf("reselect: Open=%v Query=%q selected=%v", s.Open, s.Query, listSelection.Values())
	}
}

func TestParityTagsTabRetainsFocus(t *testing.T) {
	var s nui.TagsState
	var tags []string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags", AddOnTab: true})
			nui.Button(c, th, "Next", nui.ButtonOptions{})
		})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.Type("alpha")
	tt.Key(0, ui.KeyTab)
	if !slices.Equal(tags, []string{"alpha"}) || !tt.Focused("Tags") {
		t.Fatalf("tags=%v inputFocused=%v nextFocused=%v", tags, tt.Focused("Tags"), tt.Focused("Next"))
	}
}

func TestParityTagsMaxLengthCapsDraft(t *testing.T) {
	var s nui.TagsState
	var tags []string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags", MaxLength: 3})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.Type("abcd")
	if s.Query != "abc" {
		t.Fatalf("maxlength=3 draft=%q; expected abc", s.Query)
	}
}

func TestParityTagsRejectedPasteDoesNotInsertDraft(t *testing.T) {
	var s nui.TagsState
	tags := []string{"alpha"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags", AddOnPaste: nui.Set(true)})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.SetClipboard("alpha")
	tt.Command("paste")
	if s.Query != "" {
		t.Fatalf("duplicate rejected but pasted into draft: %q", s.Query)
	}
}

func TestParityListboxSingleToggle(t *testing.T) {
	var s nui.ListboxState
	selected := []string{"a"}
	listSelection := chosenSelection(selected)
	s.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, listSelection, &s, nui.ListboxOptions{Label: "Choices", SelectionBehavior: nui.SelectionToggle})
	}, 500, 300)
	click(t, tt, "Alpha")
	if listSelection.Len() != 0 {
		t.Fatalf("toggle on selected single item retained %v", listSelection.Values())
	}
}

func TestParitySliderLargeKeyboardSteps(t *testing.T) {
	for _, inverted := range []bool{false, true} {
		for _, key := range []ui.Key{ui.KeyPageUp, ui.KeyRight} {
			t.Run(fmt.Sprintf("inverted=%v/key=%v", inverted, key), func(t *testing.T) {
				value := float64(40)
				var h ui.Handle
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Min: nui.Set(float64(0)), Max: nui.Set(float64(100)), Step: 2, Inverted: inverted}).Bind(&h)
					h.Focus(c)
				}, 500, 300)
				mods := ui.Modifiers(0)
				want := float64(60)
				if key == ui.KeyRight {
					mods = ui.Shift
					if inverted {
						want = 20
					}
				}
				tt.Key(mods, key)
				if value != want {
					t.Fatalf("value=%v want=%v", value, want)
				}
			})
		}
	}
}

func TestParityRangeCanCrossThumbs(t *testing.T) {
	low, high := float64(20), float64(40)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{Label: "Budget", Min: 0, Max: 100, Step: 1}).Width(400)
	}, 500, 300)
	r := bounds(t, tt, "Budget")
	x := func(v float32) float32 { return r.X + 8 + (r.W-16)*v/100 }
	y := r.Y + r.H/2
	tt.Press(x(20), y)
	tt.Move(x(60), y)
	tt.Release(x(60), y)
	if low != 40 || high != 60 {
		t.Fatalf("drag 20 past 40 to 60 produced [%v,%v]; upstream sorts to [40,60]", low, high)
	}
}

func TestParityRatingVerticalDown(t *testing.T) {
	value := 2
	var h ui.Handle
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Rating", Orientation: nui.Vertical}).Bind(&h)
		h.Focus(c)
	}, 500, 300)
	tt.Key(0, ui.KeyDown)
	if value != 3 {
		t.Fatalf("vertical ArrowDown gives %v; expected next rating 3", value)
	}
}

func TestParityChoiceHiddenCardHasSelectedAppearance(t *testing.T) {
	checked := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Option", Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden}).Width(250)
	}, 500, 300)
	before := slices.Clone(tt.Image().Pix)
	checked = true
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	parityScreenshot(t, tt, "checkbox-hidden-selected")
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("card with hidden indicator has identical unchecked and checked render")
	}
}

func TestParityChoiceHighlightUnchecked(t *testing.T) {
	checked, highlight := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Option", Highlight: highlight}).Width(250)
	}, 500, 300)
	before := slices.Clone(tt.Image().Pix)
	highlight = true
	tt.Frame()
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("highlight=true does not change unchecked indicator")
	}
}

func TestParityFormFieldHorizontalCopyPlacement(t *testing.T) {
	value := "abc"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Workspace", Description: "Description", Help: "Help", Hint: "Hint", Orientation: nui.Set(nui.Horizontal)}, func(th nui.Theme) { nui.Input(c, th, &value, nui.InputOptions{Label: "Editor"}).Width(250) }).Width(450)
	}, 600, 300)
	parityScreenshot(t, tt, "formfield-horizontal")
	a, b := bounds(t, tt, "Description"), bounds(t, tt, "Editor")
	t.Logf("description+help=%+v editor=%+v texts=%v", a, b, tt.Texts())
	if a.X >= b.X || bounds(t, tt, "Help").Y <= b.Y {
		t.Fatal("description belongs below the label; it is merged with help below the editor")
	}
}

func TestParityPopoverNonDismissibleKeepsFocus(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Open", nui.ButtonOptions{}).OnClick(func() { open = true })
		nui.Popover(c, th, a, &open, nui.PopoverOptions{Dismissible: nui.Set(false)}, func(th nui.Theme) { nui.Button(c, th, "Inside", nui.ButtonOptions{}) })
	}, 500, 300)
	click(t, tt, "Open")
	click(t, tt, "Inside")
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if !open || !tt.Focused("Inside") {
		t.Fatalf("Escape: open=%v insideFocused=%v triggerFocused=%v", open, tt.Focused("Inside"), tt.Focused("Open"))
	}
}

func TestParityManualTabsInitialTabStop(t *testing.T) {
	selected := "b"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.Tabs(c, th, &selected, []nui.TabItem{{ID: "a", Label: "Tab A"}, {ID: "b", Label: "Tab B"}}, nui.TabsOptions{ActivationMode: nui.ActivationManual})
			nui.Button(c, th, "After", nui.ButtonOptions{})
		})
	}, 500, 300)
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Tab B") {
		t.Fatalf("Tab entered A=%v B=%v selected=%s", tt.Focused("Tab A"), tt.Focused("Tab B"), selected)
	}
}

func TestParityInvertedSliderOtherAxis(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(fmt.Sprint(vertical), func(t *testing.T) {
			value := float64(40)
			var h ui.Handle
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Step: 2, Inverted: true, Vertical: vertical}).Bind(&h)
				h.Focus(c)
			}, 500, 300)
			key := ui.KeyUp
			if vertical {
				key = ui.KeyRight
			}
			tt.Key(0, key)
			if value != 42 {
				t.Fatalf("inverted=%v vertical=%v key=%v gives %v; expected 42", true, vertical, key, value)
			}
		})
	}
}

func TestParityControlledTooltipHoverAndClick(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Help", nui.ButtonOptions{})
		nui.Tooltip(c, th, a, "Tip", nui.TooltipOptions{Open: &open})
	}, 500, 300)
	r := bounds(t, tt, "Help")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(2 * time.Second)
	if !open {
		t.Error("controlled tooltip does not publish open on hover")
	}
	open = true
	tt.Frame()
	click(t, tt, "Help")
	if open {
		t.Error("controlled tooltip does not close on trigger click")
	}
}

func TestParityInputMenuCreateKeepsSearch(t *testing.T) {
	var s nui.SearchState
	created := ""
	s.SetOptions([]nui.Option[string]{{Value: "a", Label: "Alpha"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, nui.InputMenuOptions{Label: "Search", OnCreate: func(v string) { created = v }}).Width(300)
	}, 500, 300)
	click(t, tt, "Search")
	tt.Type("new")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if created != "new" || s.Query != "new" || !tt.HasText("Create") {
		t.Fatalf("created=%q query=%q texts=%v; upstream create prevents selection/close", created, s.Query, tt.Texts())
	}
}
