package nui_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, &selected, &state, []nui.ListboxItem{{ID: "a", Label: "Alpha"}}, nui.SelectMenuOptions{Label: "Menu", OnChange: func([]string) { changes++ }, OnSelect: func(string) { activations++ }, OnQuery: func(string) { queries++ }})
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{}).WithComponent(nui.ComponentSelect, nui.ComponentTheme{UI: nui.Styles{nui.Root: {PaddingY: nui.Set(float32(12))}}}).WithComponent(nui.ComponentSelectMenu, nui.ComponentTheme{Size: nui.XS})
		ui.Column(c).Gap(16).Children(func() {
			d := nui.Decoration{}
			if instance {
				d.UI = nui.Styles{nui.Root: {PaddingY: nui.Set(float32(15))}}
			}
			nui.Select(c, th, &single, options, nui.SelectOptions{Label: "Single", Decoration: d})
			d.Style = func(slot nui.Slot, e ui.Element) {
				if slot == nui.Root {
					e.Bind(&root)
				}
			}
			nui.SelectMultiple(c, th, &multiple, options, nui.SelectOptions{Label: "Multiple", Decoration: d})
			nui.Button(c, th, "Focus trigger", nui.ButtonOptions{}).OnClick(func() { root.Focus(c) })
		})
	}, 500, 500)
	previous := float32(32)
	for range 2 {
		tt.Frame()
		a, b := bounds(t, tt, "Single"), bounds(t, tt, "Multiple")
		if a.H != b.H || b.H <= previous {
			t.Fatalf("Select root padding: single=%v multiple=%v previous=%v", a.H, b.H, previous)
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
