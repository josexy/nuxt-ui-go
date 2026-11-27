package nui_test

import (
	"bytes"
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSwitchLoadingBlocksControlAndFieldActivation(t *testing.T) {
	on, loading, disabled := false, true, false
	changes, observed := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Disabled(disabled).Children(func() {
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Sync policy"}, func(field nui.Theme) {
				nui.Switch(c.Key("sync"), field, &on, nui.SwitchOptions{Label: "Sync", Loading: loading, OnChange: func(bool) { changes++ }}).OnChange(func() { observed++ })
			})
		})
	}, 320, 140)
	label := func() {
		r := bounds(t, tt, "Sync policy")
		tt.ClickAt(r.X+4, r.Y+8)
		tt.Frame()
	}
	click(t, tt, "Sync")
	tt.Key(0, ui.KeySpace)
	label()
	if on || changes != 0 || observed != 0 {
		t.Fatal("loading Switch accepted pointer, keyboard or field-label activation")
	}
	loading = false
	tt.Frame()
	label()
	if !on || !tt.Focused("Sync") || changes != 1 || observed != 1 {
		t.Fatal("completed loading did not restore field activation exactly once")
	}
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	tt.Frame()
	if on || changes != 2 || observed != 2 {
		t.Fatal("Space did not toggle once after loading")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Sync")
	label()
	tt.Key(0, ui.KeySpace)
	if on || changes != 2 || observed != 2 {
		t.Fatal("Switch escaped ancestor disabled state")
	}
}

func TestSwitchIconsDirectionAndReducedMotion(t *testing.T) {
	on, loading, custom, off := false, false, false, false
	size := nui.MD
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: off})
		ui.Column(c).Padding(20).Children(func() {
			o := nui.SwitchOptions{Label: "Sync", Size: size, Loading: loading, OnChange: func(bool) { changes++ }, Style: func(slot nui.Slot, e ui.Element) {
				switch slot {
				case nui.Indicator:
					e.Label("Track")
				case nui.Thumb:
					e.Label("Thumb")
				}
			}}
			if custom {
				o.CheckedIcon, o.UncheckedIcon, o.LoadingIcon = lucide.Check(), lucide.X(), lucide.RefreshCw()
			}
			nui.Switch(c.Key("sync"), th, &on, o)
		})
	}, 300, 100)
	for _, value := range []bool{false, true} {
		on, custom = value, false
		tt.Frame()
		plain := slices.Clone(tt.Image().Pix)
		custom = true
		tt.Frame()
		if bytes.Equal(plain, tt.Image().Pix) {
			t.Fatalf("custom icon is missing when checked=%v", value)
		}
	}
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(direction)
		for i, choiceSize := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
			size = choiceSize
			for _, value := range []bool{false, true} {
				on = value
				tt.Frame()
				tt.Advance(250 * time.Millisecond) // the thumb's slide
				track, thumb := bounds(t, tt, "Track"), bounds(t, tt, "Thumb")
				want := track.X + 2
				if value != (direction == ui.RightToLeft) {
					want = track.X + track.W - thumb.W - 2
				}
				if track.W != float32(28+4*i) || thumb.W != float32(12+2*i) || thumb.X != want || thumb.Y != track.Y+2 {
					t.Fatalf("size=%v direction=%v checked=%v: track=%+v thumb=%+v", size, direction, value, track, thumb)
				}
			}
		}
	}
	parityScreenshot(t, tt, "switch-icons-rtl")
	// The spinner turns under reduced motion and stands still with
	// animations off.
	loading = true
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()
	before := slices.Clone(tt.Image().Pix)
	tt.Advance(250 * time.Millisecond)
	if !tt.FramePending() || bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("reduced-motion loading stopped the spinner")
	}
	off = true
	tt.Frame()
	before = slices.Clone(tt.Image().Pix)
	tt.Advance(250 * time.Millisecond)
	if tt.FramePending() || !bytes.Equal(before, tt.Image().Pix) || changes != 0 {
		t.Fatal("loading with animations off continued animating or appearance updates emitted changes")
	}
	custom = false
	tt.Frame()
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("LoadingIcon did not differ from the default loader")
	}
}

func TestSliderNeutralUsesInvertedBackground(t *testing.T) {
	for _, rangeMode := range []bool{false, true} {
		t.Run(fmt.Sprint(rangeMode), func(t *testing.T) {
			value, low, high := 50.0, 20.0, 80.0
			accent := ui.RGB(126, 34, 206)
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				th.InvertedBackground = accent
				if rangeMode {
					nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{Label: "Amount", Color: nui.Neutral}).Width(300)
				} else {
					nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Color: nui.Neutral}).Width(300)
				}
			}, 360, 90)
			r := bounds(t, tt, "Amount")
			if got := tt.Image().RGBAAt(int(r.X+r.W*.3), int(r.Y+r.H/2)); got != rgba(accent) {
				t.Fatalf("neutral slider fill=%v, want %v", got, rgba(accent))
			}
		})
	}
}

func TestRangeSliderUsesIndicatorAndThumbStyles(t *testing.T) {
	low, high := 20.0, 80.0
	fill, thumb := ui.RGB(255, 0, 255), ui.RGB(0, 255, 0)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{Label: "Range", UI: nui.Styles{nui.Indicator: {Background: nui.Set(fill)}, nui.Thumb: {Background: nui.Set(thumb)}}}).Width(300)
	}, 360, 90)
	r := bounds(t, tt, "Range")
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)); got != rgba(fill) {
		t.Fatalf("range indicator fill=%v, want %v", got, rgba(fill))
	}
	if got := tt.Image().RGBAAt(int(r.X+8+(r.W-16)*.2), int(r.Y+r.H/2)); got != rgba(thumb) {
		t.Fatalf("range thumb fill=%v, want %v", got, rgba(thumb))
	}
}

func TestSliderDecimalStepValueAndTooltip(t *testing.T) {
	for _, rangeMode := range []bool{false, true} {
		t.Run(fmt.Sprint(rangeMode), func(t *testing.T) {
			value, high, changes, commits := .2, .8, 0, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(30).Children(func() {
					if rangeMode {
						nui.RangeSlider(c, th, &value, &high, nui.RangeSliderOptions{Label: "Amount", Max: 1, Step: .1, Tooltip: true, OnChange: func(float64, float64) { changes++ }, OnCommit: func(float64, float64) { commits++ }}).Width(300)
					} else {
						nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Max: nui.Set(1.0), Step: .1, Tooltip: true, OnChange: func(float64) { changes++ }, OnCommit: func(float64) { commits++ }}).Width(300)
					}
				})
			}, 400, 120)
			tt.Key(0, ui.KeyTab)
			tt.Key(0, ui.KeyRight)
			if value != .3 || changes != 1 || commits != 1 {
				t.Fatalf("decimal step value=%.17g changes=%d commits=%d", value, changes, commits)
			}
			if _, ok := tt.Find("0.3"); !ok {
				t.Fatalf("decimal tooltip did not show the step value: %v", tt.Texts())
			}
		})
	}
}

func TestSliderInheritsFormFieldErrorColor(t *testing.T) {
	for _, rangeMode := range []bool{false, true} {
		t.Run(fmt.Sprint(rangeMode), func(t *testing.T) {
			value, high := 30.0, 80.0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				nui.FormField(c, th, nui.FormFieldOptions{Label: "Amount field", Error: "Choose another value"}, func(ft nui.Theme) {
					if rangeMode {
						nui.RangeSlider(c, ft, &value, &high, nui.RangeSliderOptions{Color: nui.Success}).Width(300).Label("Slider track")
					} else {
						nui.Slider(c, ft, &value, nui.SliderOptions{Color: nui.Success}).Width(300).Label("Slider track")
					}
				})
			}, 360, 160)
			r := bounds(t, tt, "Slider track")
			fraction := float32(.15)
			if rangeMode {
				fraction = .5
			}
			if got := tt.Image().RGBAAt(int(r.X+r.W*fraction), int(r.Y+r.H/2)); got != rgba(nui.Light().Error) {
				t.Fatalf("field error fill=%v, want %v", got, rgba(nui.Light().Error))
			}
		})
	}
}

func TestSliderDefaultAndExplicitAccessibleName(t *testing.T) {
	value := 40.0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Slider(c, nui.App(c, nui.AppConfig{}), &value, nui.SliderOptions{}).Width(240)
	}, 300, 90)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Slider") {
		t.Fatal("unnamed slider has no default accessible name")
	}
	tt.SetLocale("zh-Hans")
	if !tt.Focused("滑块") || value != 40 {
		t.Fatal("translated slider name changed its focus or value")
	}
	ft := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Field label"}, func(inner nui.Theme) {
			nui.Slider(c, inner, &value, nui.SliderOptions{Label: "Custom amount"}).Width(240)
		})
	}, 300, 120)
	ft.Key(0, ui.KeyTab)
	if !ft.Focused("Custom amount") {
		t.Fatal("field label replaced the explicitly named slider")
	}
}

func TestSliderTooltipFollowsTrackPadding(t *testing.T) {
	value := 0.0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(60).Children(func() {
			nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Tooltip: true, UI: nui.Styles{nui.Root: {PaddingX: nui.Set(float32(64))}}}).Width(300)
		})
	}, 440, 180)
	tt.Key(0, ui.KeyTab)
	root, tip := bounds(t, tt, "Amount"), bounds(t, tt, "0")
	if math.Abs(float64(tip.X+tip.W/2-root.X-64)) > 1 {
		t.Fatalf("tooltip is not over the padded thumb: root=%+v tip=%+v", root, tip)
	}
}

func TestSliderDecimalOffsetsAndSmallSteps(t *testing.T) {
	for _, tc := range []struct{ minimum, value, maximum, step, want float64 }{
		{0, .3, .35, .1, .35}, {.05, .05, 1, .1, .15}, {-.3, -.3, 1, .1, -.2}, {0, 2e-7, 1, 1e-7, 3e-7},
	} {
		for _, rangeMode := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/range=%v", tc, rangeMode), func(t *testing.T) {
				value, high := tc.value, tc.maximum
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					if rangeMode {
						nui.RangeSlider(c, th, &value, &high, nui.RangeSliderOptions{Min: tc.minimum, Max: tc.maximum, Step: tc.step}).Width(240)
					} else {
						nui.Slider(c, th, &value, nui.SliderOptions{Min: nui.Set(tc.minimum), Max: nui.Set(tc.maximum), Step: tc.step}).Width(240)
					}
				}, 300, 100)
				tt.Key(0, ui.KeyTab)
				tt.Key(0, ui.KeyRight)
				if value != tc.want {
					t.Fatalf("step value=%.17g want %.17g", value, tc.want)
				}
			})
		}
	}
}

func TestSliderParentDisablingCancelsDrag(t *testing.T) {
	for _, rangeMode := range []bool{false, true} {
		t.Run(fmt.Sprint(rangeMode), func(t *testing.T) {
			value, high, disabled, changes, commits := 20.0, 80.0, false, 0, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Box(c).Disabled(disabled).Children(func() {
					if rangeMode {
						nui.RangeSlider(c, th, &value, &high, nui.RangeSliderOptions{Label: "Amount", OnChange: func(float64, float64) { changes++ }, OnCommit: func(float64, float64) { commits++ }}).Width(300)
					} else {
						nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", OnChange: func(float64) { changes++ }, OnCommit: func(float64) { commits++ }}).Width(300)
					}
				})
			}, 360, 100)
			r := bounds(t, tt, "Amount")
			x := func(v float32) float32 { return r.X + 8 + (r.W-16)*v/100 }
			y := r.Y + r.H/2
			tt.Press(x(20), y)
			tt.Move(x(30), y)
			disabled = true
			tt.Frame()
			tt.Release(x(30), y)
			disabled = false
			tt.Frame()
			tt.Move(x(60), y)
			if value != 30 || changes != 1 || commits != 0 {
				t.Fatalf("disabled parent left a drag active: value=%g changes=%d commits=%d", value, changes, commits)
			}
		})
	}
}

func TestRangeSliderSmallMinimumGapDoesNotCollapse(t *testing.T) {
	low, high, changes := 0.0, 1e-12, 0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.RangeSlider(c, nui.App(c, nui.AppConfig{}), &low, &high, nui.RangeSliderOptions{Label: "Range", Max: 1e-9, Step: 1e-12, MinStepsBetweenThumbs: 1, OnChange: func(float64, float64) { changes++ }}).Width(300)
	}, 360, 100)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyRight)
	if low != 0 || high != 1e-12 || changes != 0 {
		t.Fatalf("minimum gap collapsed: low=%g high=%g changes=%d", low, high, changes)
	}
}

func TestChoiceGroupItemIconsKeepIdentity(t *testing.T) {
	for _, radio := range []bool{false, true} {
		t.Run(map[bool]string{false: "checkbox", true: "radio"}[radio], func(t *testing.T) {
			items := []nui.Option[string]{{Value: "a", Label: "Alpha", Icon: lucide.Heart()}, {Value: "b", Label: "Beta", Icon: lucide.Bell()}}
			selected, single := []string{"a"}, "a"
			placement, horizontal, changes := nui.IndicatorStart, false, 0
			icons := 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				icons = 0
				ui.Column(c).Padding(20).Children(func() {
					o := nui.ChoiceGroupOptions{Label: "Letters", Variant: nui.ChoiceCard, Indicator: placement, Horizontal: horizontal, OnChange: func() { changes++ }, Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Leading {
							icons++
						}
					}}
					if radio {
						nui.RadioGroup(c.Key("group"), th, &single, items, o)
					} else {
						nui.CheckboxGroup(c.Key("group"), th, &selected, items, o)
					}
				})
			}, 420, 260)
			before := slices.Clone(tt.Image().Pix)
			items[0].Icon = lucide.X()
			tt.Frame()
			if radio {
				if icons != 0 || !bytes.Equal(before, tt.Image().Pix) {
					t.Fatal("radio item icon replaced the visible radio dot")
				}
			} else if icons != 1 || bytes.Equal(before, tt.Image().Pix) {
				t.Fatal("checkbox item icon did not replace the checked indicator")
			}
			placement = nui.IndicatorHidden
			tt.Frame()
			if icons != 2 {
				t.Fatalf("hidden indicators rendered %d label icons, want 2", icons)
			}
			click(t, tt, "Beta")
			items[0], items[1] = items[1], items[0]
			items[0].Label = "Beta renamed"
			horizontal = true
			tt.SetDirection(ui.RightToLeft)
			tt.Frame()
			if !tt.Focused("Beta renamed") || icons != 2 || changes != 1 || single != map[bool]string{true: "b", false: "a"}[radio] {
				t.Fatal("reordering or changing layout lost the item's focus, icon or single value")
			}
			if !radio && !slices.Equal(selected, []string{"a", "b"}) {
				t.Fatal("reordering changed the multi-selection")
			}
			parityScreenshot(t, tt, map[bool]string{false: "checkbox-group-icons-rtl", true: "radio-group-icons-rtl"}[radio])
		})
	}
}

func TestCheckboxGroupArrowNavigationDoesNotSelect(t *testing.T) {
	var selected []string
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.CheckboxGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Blocked", Disabled: true}, {Value: "c", Label: "Charlie"}}, nui.CheckboxGroupOptions{OnChange: func() { changes++ }})
			nui.Button(c, th, "After", nui.ButtonOptions{})
		})
	}, 300, 200)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyDown)
	if !tt.Focused("Charlie") || len(selected) != 0 || changes != 0 {
		t.Fatal("Down did not move focus past the disabled checkbox without selecting")
	}
	tt.Key(0, ui.KeyDown)
	if !tt.Focused("Charlie") {
		t.Fatal("default CheckboxGroup navigation wrapped at the end")
	}
	tt.Key(0, ui.KeySpace)
	if !slices.Equal(selected, []string{"c"}) || changes != 1 {
		t.Fatal("Space did not select the focused checkbox exactly once")
	}
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("After") {
		t.Fatal("Tab did not leave the checkbox group")
	}
}

func TestChoiceGroupLoopOrientationAndDisabledItems(t *testing.T) {
	for _, radio := range []bool{false, true} {
		for _, loop := range []bool{false, true} {
			single, selected, changes := "a", []string{}, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				items := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Blocked", Disabled: true}, {Value: "c", Label: "Charlie"}}
				o := nui.ChoiceGroupOptions{Horizontal: true, Loop: nui.Set(loop), OnChange: func() { changes++ }}
				if radio {
					nui.RadioGroup(c, th, &single, items, o)
				} else {
					nui.CheckboxGroup(c, th, &selected, items, o)
				}
			}, 320, 120)
			tt.Key(0, ui.KeyTab)
			tt.Key(0, ui.KeyDown)
			if !tt.Focused("Alpha") || changes != 0 {
				t.Fatal("horizontal group responded to the vertical arrow")
			}
			tt.Key(0, ui.KeyRight)
			if !tt.Focused("Charlie") {
				t.Fatal("horizontal arrow did not skip the disabled option")
			}
			tt.Key(0, ui.KeyRight)
			want, count := "Charlie", 1
			if loop {
				want, count = "Alpha", 2
			}
			if !tt.Focused(want) || (radio && changes != count) || (!radio && (changes != 0 || len(selected) != 0)) {
				t.Fatalf("radio=%t loop=%t boundary focus or events: value=%s changes=%d", radio, loop, single, changes)
			}
			tt.SetDirection(ui.RightToLeft)
			tt.Key(0, ui.KeyHome)
			tt.Key(0, ui.KeyLeft)
			if !tt.Focused("Charlie") {
				t.Fatal("RTL did not reverse horizontal navigation")
			}
		}
	}
}

func TestCheckboxGroupZeroLimitAlsoBlocksFieldLabel(t *testing.T) {
	var selected []string
	limit, changes := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Features"}, func(field nui.Theme) {
			nui.CheckboxGroup(c, field, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.CheckboxGroupOptions{Max: nui.Set(limit), OnChange: func() { changes++ }})
		})
	}, 300, 180)
	r := bounds(t, tt, "Features")
	tt.ClickAt(r.X+3, r.Y+5)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeySpace)
	if !tt.Focused("Beta") || len(selected) != 0 || changes != 0 {
		t.Fatal("Max=0 lost focus or accepted field-label / keyboard activation")
	}
	limit = 1
	tt.Frame()
	tt.Key(0, ui.KeySpace)
	click(t, tt, "Alpha")
	if !slices.Equal(selected, []string{"b"}) || changes != 1 {
		t.Fatal("Max=1 did not allow the first selection and refuse the second")
	}
	click(t, tt, "Beta")
	click(t, tt, "Alpha")
	if !slices.Equal(selected, []string{"a"}) || changes != 3 {
		t.Fatal("unchecking at the limit did not restore another selection")
	}
}

func TestCheckboxHiddenIndicatorKeepsExplicitIcon(t *testing.T) {
	checked, changes := false, 0
	icon := lucide.Heart()
	label, description := "Receive updates", "Only for this workspace"
	size := nui.MD
	width := float32(260)
	name := func(slot nui.Slot, e ui.Element) {
		switch slot {
		case nui.Root:
			e.Label("Choice")
		case nui.Leading:
			e.Label("Choice icon")
		case nui.Label:
			e.Label("Choice label")
		case nui.Description:
			e.Label("Choice description")
		}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c.Key("choice"), th, &checked, nui.CheckboxOptions{Label: label, Description: description, Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden, CheckedIcon: icon, Size: size, Style: name, OnChange: func(bool) { changes++ }}).Width(width)
		})
	}, 320, 180)
	glyph := bounds(t, tt, "Choice icon")
	caption := bounds(t, tt, "Choice label")
	detail := bounds(t, tt, "Choice description")
	if glyph.W != 16 || glyph.H != 16 || caption.Y != glyph.Y+glyph.H+4 || detail.Y != caption.Y+caption.H+4 {
		t.Fatalf("hidden indicator layout: icon=%+v label=%+v description=%+v", glyph, caption, detail)
	}
	if glyph.X+glyph.W/2 != caption.X+caption.W/2 {
		t.Fatal("icon and text are not centered together")
	}
	parityScreenshot(t, tt, "checkbox-hidden-label-icon")
	click(t, tt, "Choice")
	if !checked || changes != 1 || !tt.Focused("Choice") {
		t.Fatal("hidden indicator lost its toggle or focus behavior")
	}
	size = nui.XL
	tt.Frame()
	if bounds(t, tt, "Choice icon").W != 20 || changes != 1 || !tt.Focused("Choice") {
		t.Fatal("changing size lost the icon, focus or selection")
	}
	tt.Key(0, ui.KeySpace)
	if checked || changes != 2 {
		t.Fatal("hidden control did not toggle exactly once with Space")
	}
	width, label, description = 140, "接收工作区更新通知", "只应用于当前工作区的更新设置"
	tt.SetSize(320, 320)
	tt.SetDirection(ui.RightToLeft)
	tt.SetDark(true)
	tt.SetZoom(1.25)
	tt.Frame()
	root := bounds(t, tt, "Choice")
	for _, name := range []string{"Choice icon", "Choice label", "Choice description"} {
		r := bounds(t, tt, name)
		if r.X < root.X || r.X+r.W > root.X+root.W || r.Y+r.H > root.Y+root.H {
			t.Fatalf("%s escapes the narrow scaled card: %+v outside %+v", name, r, root)
		}
	}
	if !tt.Focused("Choice") || changes != 2 {
		t.Fatal("direction, theme or text scaling lost focus or changed the value")
	}
	parityScreenshot(t, tt, "checkbox-hidden-narrow-rtl")
	label, description = "", ""
	tt.Frame()
	if _, found := tt.Find("Choice icon"); !found {
		t.Fatal("explicit icon disappeared when text was cleared")
	}
	icon = nil
	tt.Frame()
	if _, found := tt.Find("Choice icon"); found {
		t.Fatal("hidden indicator inserted a fallback icon")
	}
	if changes != 2 {
		t.Fatal("presentation-only changes emitted selection events")
	}
}

func TestCheckboxCardFocusDoesNotOutlineBothCardAndIndicator(t *testing.T) {
	checked := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Focus card", Variant: nui.ChoiceCard, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Indicator {
					e.Label("Indicator")
				}
			}}).Width(240)
		})
	}, 280, 100)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Focus card") {
		t.Fatal("card did not receive keyboard focus")
	}
	r := bounds(t, tt, "Indicator")
	im := tt.Image()
	x, y := int(r.X+r.W), int(r.Y+r.H/2)
	if im.RGBAAt(x+1, y) != im.RGBAAt(x+5, y) {
		t.Fatal("card focus also draws an outline around its inner indicator")
	}
	parityScreenshot(t, tt, "checkbox-card-focus")
}

func TestCheckboxCardPreservesPartOverrides(t *testing.T) {
	checked := true
	global, instance := ui.RGB(220, 80, 120), ui.RGB(60, 100, 180)
	var explicit nui.Styles
	var color nui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode}).WithComponent(nui.ComponentCheckbox, nui.ComponentTheme{UI: nui.Styles{nui.Root: {Background: nui.Set(global), PaddingY: nui.Set(float32(22))}}})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Styled card", Variant: nui.ChoiceCard, Color: color, UI: explicit}).Width(240)
		})
	}, 280, 120)
	pixel := func() [4]uint32 {
		t.Helper()
		r := bounds(t, tt, "Styled card")
		rr, g, b, a := tt.Image().At(int(r.X+30), int(r.Y+8)).RGBA()
		return [4]uint32{rr, g, b, a}
	}
	rgba32 := func(c ui.Color) [4]uint32 { r, g, b, a := rgba(c).RGBA(); return [4]uint32{r, g, b, a} }
	if pixel() != rgba32(global) || bounds(t, tt, "Styled card").H != 64 {
		t.Fatalf("card state overwrote the component's root background or padding: color=%v want=%v bounds=%+v", pixel(), rgba32(global), bounds(t, tt, "Styled card"))
	}
	color = nui.Error
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	if pixel() == rgba32(global) {
		t.Fatal("explicit Color did not take precedence over global appearance")
	}
	explicit = nui.Styles{nui.Root: {Background: nui.Set(instance)}}
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	if pixel() != rgba32(instance) {
		t.Fatal("instance root style did not take precedence over Color")
	}
	checked = false
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	if pixel() != rgba32(instance) {
		t.Fatal("changing selection discarded the instance root style")
	}
}

func TestCheckboxCardHighlightFollowsIndicatorPlacement(t *testing.T) {
	checked, highlight := false, false
	placement := nui.IndicatorStart
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Highlighted card", Variant: nui.ChoiceCard, Highlight: highlight, Indicator: placement}).Width(240)
		})
	}, 280, 100)
	border := func() [4]uint32 {
		r := bounds(t, tt, "Highlighted card")
		rr, g, b, a := tt.Image().At(int(r.X+r.W/2), int(r.Y)).RGBA()
		return [4]uint32{rr, g, b, a}
	}
	before := border()
	highlight = true
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	if border() != before {
		t.Fatal("visible indicator highlight also recolored the card border")
	}
	placement = nui.IndicatorHidden
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	if border() == before {
		t.Fatal("hidden indicator highlight did not recolor the card border")
	}
}

func TestCheckboxCardHoverAndFocusBorder(t *testing.T) {
	checked := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Card", Variant: nui.ChoiceCard}).Width(240)
		})
	}, 280, 100)
	r := bounds(t, tt, "Card")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(200 * time.Millisecond) // the colour transition
	th := nui.Light()
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)); got != rgba(th.BorderAccented) {
		t.Fatalf("hovered card border = %v, want %v", got, rgba(th.BorderAccented))
	}
	tt.Key(0, ui.KeyTab)
	tt.Advance(200 * time.Millisecond) // the colour transition
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)); got != rgba(th.Primary) {
		t.Fatalf("focused card border = %v, want %v", got, rgba(th.Primary))
	}
}

func TestCheckboxTriCardMixedKeepsUnselectedCardBackground(t *testing.T) {
	value := nui.Unchecked
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.CheckboxTri(c, th, &value, nui.CheckboxTriOptions{Label: "All", Variant: nui.ChoiceCard}).Width(240)
		})
	}, 280, 100)
	r := bounds(t, tt, "All")
	x, y := int(r.X+r.W/2), int(r.Y+8)
	unchecked := tt.Image().RGBAAt(x, y)
	value = nui.Indeterminate
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	if tt.Image().RGBAAt(x, y) != unchecked {
		t.Fatal("mixed indicator incorrectly applies the checked card background")
	}
	value = nui.Checked
	tt.Frame()
	tt.Advance(200 * time.Millisecond) // the colour transition
	if tt.Image().RGBAAt(x, y) == unchecked {
		t.Fatal("checked card did not receive its selected background")
	}
}

func TestChoiceDescriptionFollowsLabelLine(t *testing.T) {
	for _, kind := range []string{"checkbox", "switch", "radio"} {
		t.Run(kind, func(t *testing.T) {
			on, selected := false, "a"
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				name := func(slot nui.Slot, e ui.Element) {
					if slot == nui.Label {
						e.Label("Caption line")
					}
					if slot == nui.Description {
						e.Label("Description line")
					}
				}
				o := nui.ToggleOptions{Label: "Caption", Description: "Description", Style: name}
				switch kind {
				case "checkbox":
					nui.Checkbox(c, th, &on, o)
				case "switch":
					nui.Switch(c, th, &on, o)
				default:
					nui.RadioGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Caption", Description: "Description"}}, nui.RadioGroupOptions{Style: name})
				}
			}, 300, 100)
			label, description := bounds(t, tt, "Caption line"), bounds(t, tt, "Description line")
			if description.Y != label.Y+label.H {
				t.Fatalf("unexpected gap between normal label and description: label=%+v description=%+v", label, description)
			}
		})
	}
}

func TestChoiceIndicatorPartRadiusOverride(t *testing.T) {
	for _, kind := range []string{"radio", "switch"} {
		t.Run(kind, func(t *testing.T) {
			on, selected := true, "a"
			marker := ui.RGB(255, 0, 255)
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				style := nui.Decoration{UI: nui.Styles{nui.Indicator: {Radius: nui.Set(float32(0)), BorderWidth: nui.Set(float32(0)), Background: nui.Set(marker)}}, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Indicator {
						e.Label("Shape")
					}
				}}
				ui.Column(c).Padding(20).Children(func() {
					if kind == "radio" {
						nui.RadioGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.RadioGroupOptions{Decoration: style})
					} else {
						nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Sync", Decoration: style})
					}
				})
			}, 300, 100)
			r := bounds(t, tt, "Shape")
			if tt.Image().RGBAAt(int(r.X+1), int(r.Y+1)) != rgba(marker) {
				t.Fatal("default rounding overwrote explicit indicator part styling")
			}
		})
	}
}

func TestCheckboxTriCustomIconsKeepStateAndEvents(t *testing.T) {
	value := nui.Indeterminate
	var custom *ui.SVG
	placement := nui.IndicatorStart
	disabled := false
	changes, states := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.CheckboxTri(c.Key("tri"), th, &value, nui.CheckboxTriOptions{
				Label: "Select all", CheckedIcon: lucide.Heart(), Indicator: placement, Disabled: disabled, OnChange: func(bool) { changes++ },
				IndeterminateIcon: custom,
				OnStateChange:     func(nui.CheckState) { states++ },
			}).Width(220)
		})
	}, 280, 150)
	minus := slices.Clone(tt.Image().Pix)
	custom = lucide.X()
	tt.Frame()
	if bytes.Equal(minus, tt.Image().Pix) {
		t.Fatal("custom indeterminate icon did not replace the default minus")
	}
	parityScreenshot(t, tt, "checkbox-custom-indeterminate")
	value = nui.Checked
	tt.Frame()
	checked := slices.Clone(tt.Image().Pix)
	custom = nil
	tt.Frame()
	if !bytes.Equal(checked, tt.Image().Pix) {
		t.Fatal("indeterminate icon affected the checked state")
	}
	value = nui.Indeterminate
	placement = nui.IndicatorHidden
	tt.Frame()
	hidden := slices.Clone(tt.Image().Pix)
	custom = lucide.X()
	tt.Frame()
	if !bytes.Equal(hidden, tt.Image().Pix) || changes != 0 || states != 0 {
		t.Fatal("mixed icon replaced the hidden label icon or emitted an event")
	}
	click(t, tt, "Select all")
	tt.Frame()
	tt.Frame()
	if value != nui.Checked || changes != 1 || states != 1 {
		t.Fatalf("mixed activation: state=%v changes=%d states=%d", value, changes, states)
	}
	tt.Key(0, ui.KeySpace)
	if value != nui.Unchecked || changes != 2 || states != 2 {
		t.Fatal("Space did not change the checked state exactly once")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Select all")
	tt.Key(0, ui.KeySpace)
	if value != nui.Unchecked || changes != 2 || states != 2 {
		t.Fatal("disabled custom-icon checkbox changed state")
	}
}
