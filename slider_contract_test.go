package nui_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

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
