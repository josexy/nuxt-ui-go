package nui_test

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestInputNumberOrientationGeometry(t *testing.T) {
	for _, orientation := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
		for _, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
			for _, rtl := range []bool{false, true} {
				t.Run(fmt.Sprintf("orientation=%d/size=%d/rtl=%v", orientation, size, rtl), func(t *testing.T) {
					value := nui.Set(125.0)
					var state nui.NumberState
					text := "125"
					tt := ui.NewTester(func(c *ui.Context) {
						th := nui.App(c, nui.AppConfig{})
						ui.Column(c).Padding(20).Gap(20).Children(func() {
							nui.InputNumber(c.Key("number"), th, &value, &state, nui.InputNumberOptions{Label: "Amount", Size: size, Orientation: orientation}).Width(320).Label("Number shell")
							nui.Input(c.Key("reference"), th, &text, nui.InputOptions{Label: "Reference", Size: size}).Width(320).Label("Reference shell")
						})
					}, 420, 260)
					if rtl {
						tt.SetDirection(ui.RightToLeft)
					}
					for _, textScale := range []float32{1, 1.5} {
						tt.SetPreferences(ui.Preferences{TextScale: textScale})
						r, editor := bounds(t, tt, "Number shell"), bounds(t, tt, "Amount")
						minus, plus := bounds(t, tt, "Decrease Amount"), bounds(t, tt, "Increase Amount")
						if r.H != bounds(t, tt, "Reference shell").H {
							t.Fatalf("step controls inflated input height: %v", r)
						}
						for _, button := range []ui.Rect{minus, plus} {
							if button.X < r.X || button.X+button.W > r.X+r.W || button.Y < r.Y || button.Y+button.H > r.Y+r.H || button.W <= 0 || button.H <= 0 {
								t.Fatalf("button outside shell: button=%v shell=%v", button, r)
							}
						}
						if orientation == nui.Horizontal {
							left, right := minus, plus
							if rtl {
								left, right = plus, minus
							}
							if left.X+left.W > editor.X+.1 || right.X < editor.X+editor.W-.1 || minus.Y != plus.Y || minus.H != plus.H {
								t.Fatalf("horizontal buttons must flank editor: %v %v %v", left, editor, right)
							}
							if math.Abs(float64(editor.X+editor.W/2-r.X-r.W/2)) > 1 {
								t.Fatal("horizontal editor is not centered")
							}
						} else {
							if plus.X != minus.X || plus.W != minus.W || plus.Y+plus.H > minus.Y+.1 || math.Abs(float64(minus.Y-plus.Y-plus.H)) > 1 {
								t.Fatalf("vertical controls must be stacked without overlap: up=%v down=%v", plus, minus)
							}
							if (!rtl && plus.X < editor.X+editor.W-.1) || (rtl && plus.X+plus.W > editor.X+.1) {
								t.Fatal("vertical controls overlap the editor")
							}
						}
					}
				})
			}
		}
	}
}

func TestInputNumberOrientationKeepsEditorAndButtons(t *testing.T) {
	orientation := nui.Horizontal
	value := nui.Set(12.5)
	var state nui.NumberState
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(24).Children(func() {
			nui.InputNumber(c, th, &value, &state, nui.InputNumberOptions{Label: "Amount", Orientation: orientation, Step: .5, FocusOnChange: nui.Set(false), OnValueChange: func(nui.Optional[float64]) { changes++ }}).Width(320)
		})
	}, 400, 160)
	click(t, tt, "Amount")
	tt.Key(ui.Ctrl, ui.KeyA)
	for _, direction := range []nui.Orientation{nui.Vertical, nui.Horizontal, nui.Vertical} {
		orientation = direction
		tt.Frame()
		tt.Command("copy")
		if !tt.Focused("Amount") || tt.Clipboard() != "12.5" || state.Text != "12.5" || changes != 0 {
			t.Fatal("switching orientation changed focus, selection, draft or value")
		}
	}
	tt.Type("-")
	orientation = nui.Horizontal
	tt.Frame()
	if !tt.Focused("Amount") || state.Text != "-" || value.Value != 12.5 || changes != 0 || state.Error == "" {
		t.Fatal("switching orientation discarded an invalid draft")
	}
	tt.Key(ui.Ctrl, ui.KeyZ)
	if state.Text != "12.5" {
		t.Fatal("orientation change reset the editor's undo history")
	}
	click(t, tt, "Increase Amount")
	if !tt.Focused("Increase Amount") || value.Value != 13 || changes != 1 {
		t.Fatal("horizontal increase button did not activate once")
	}
	orientation = nui.Vertical
	tt.Frame()
	if !tt.Focused("Increase Amount") {
		t.Fatal("orientation change replaced the focused button")
	}
	tt.Key(0, ui.KeySpace)
	if value.Value != 13.5 || changes != 2 {
		t.Fatal("focused button lost keyboard activation")
	}
	tt.SetLocale("zh-Hans")
	click(t, tt, "减少Amount")
	if value.Value != 13 || changes != 3 {
		t.Fatal("vertical decrease button did not activate once after locale switch")
	}
}

func TestInputNumberOrientationDisabledBoundsAndKeyboard(t *testing.T) {
	for _, orientation := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
		t.Run(fmt.Sprint(orientation), func(t *testing.T) {
			var value nui.Optional[float64]
			var state nui.NumberState
			readOnly, disabled, ancestorDisabled := false, false, false
			changes := 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(20).Disabled(ancestorDisabled).Children(func() {
					nui.InputNumber(c, th, &value, &state, nui.InputNumberOptions{Label: "Amount", ReadOnly: readOnly, Disabled: disabled, Orientation: orientation, Min: nui.Set(0.0), Max: nui.Set(1.0), Step: .5, OnValueChange: func(nui.Optional[float64]) { changes++ }}).Width(320)
				})
			}, 400, 180)
			click(t, tt, "Increase Amount")
			click(t, tt, "Increase Amount")
			click(t, tt, "Increase Amount")
			click(t, tt, "Increase Amount")
			if value.Value != 1 || changes != 3 {
				t.Fatal("upper-bound button changed value or emitted a duplicate callback")
			}
			click(t, tt, "Decrease Amount")
			click(t, tt, "Decrease Amount")
			click(t, tt, "Decrease Amount")
			if value.Value != 0 || changes != 5 {
				t.Fatal("lower-bound button changed value or emitted a duplicate callback")
			}
			click(t, tt, "Amount")
			tt.Key(0, ui.KeyUp)
			tt.Key(0, ui.KeyDown)
			if value.Value != 0 || changes != 7 {
				t.Fatal("editor's arrow keys changed with orientation")
			}
			for mode := range 3 {
				readOnly, disabled, ancestorDisabled = mode == 0, mode == 1, mode == 2
				tt.Frame()
				click(t, tt, "Increase Amount")
				click(t, tt, "Amount")
				tt.Key(0, ui.KeyUp)
				tt.Type("9")
				if value.Value != 0 || state.Text != "0" || changes != 7 {
					t.Fatalf("mode %d allowed editing", mode)
				}
			}
		})
	}
}

func TestInputNumberStepPrecisionAndPendingPress(t *testing.T) {
	for _, sample := range []struct {
		name    string
		value   nui.Optional[float64]
		options nui.InputNumberOptions
		key     ui.Key
		want    float64
	}{
		{"empty", nui.Optional[float64]{}, nui.InputNumberOptions{Min: nui.Set(5.0), Step: 2}, ui.KeyUp, 5},
		{"starting", nui.Optional[float64]{}, nui.InputNumberOptions{StartingValue: nui.Set(40.0), Step: 3}, ui.KeyDown, 40},
		{"off-grid up", nui.Set(18.98), nui.InputNumberOptions{}, ui.KeyUp, 19},
		{"off-grid down", nui.Set(18.98), nui.InputNumberOptions{}, ui.KeyDown, 18},
		{"page", nui.Set(18.98), nui.InputNumberOptions{}, ui.KeyPageUp, 28},
		{"tiny step", nui.Set(1e-13), nui.InputNumberOptions{Step: 1e-13}, ui.KeyUp, 2e-13},
		{"large decimal", nui.Set(9000.1), nui.InputNumberOptions{Step: .1}, ui.KeyUp, 9000.2},
		{"decimal maximum", nui.Set(.2), nui.InputNumberOptions{Min: nui.Set(.1), Max: nui.Set(.3), Step: .1}, ui.KeyUp, .3},
		{"off-grid maximum", nui.Set(9.0), nui.InputNumberOptions{Max: nui.Set(9.5)}, ui.KeyUp, 9},
		{"free step", nui.Set(1.4), nui.InputNumberOptions{StepSnapping: nui.Set(false)}, ui.KeyUp, 2.4},
		{"out-of-range direction", nui.Set(20.0), nui.InputNumberOptions{Max: nui.Set(10.0), AllowInvalid: true}, ui.KeyUp, 20},
	} {
		t.Run(sample.name, func(t *testing.T) {
			value, state := sample.value, &nui.NumberState{}
			calls := 0
			sample.options.Label = "Precision"
			sample.options.OnValueChange = func(nui.Optional[float64]) { calls++ }
			tt := ui.NewTester(func(c *ui.Context) { nui.InputNumber(c, nui.App(c, nui.AppConfig{}), &value, state, sample.options) }, 360, 100)
			click(t, tt, "Precision")
			tt.Key(0, sample.key)
			wantCalls := 1
			if sample.value == nui.Set(sample.want) {
				wantCalls = 0
			}
			if value != nui.Set(sample.want) || state.Text != strconv.FormatFloat(sample.want, 'f', -1, 64) || calls != wantCalls {
				t.Fatalf("value=%v text=%q calls=%d", value, state.Text, calls)
			}
		})
	}
	value, state := nui.Set(5.0), &nui.NumberState{}
	disabled, shown, mounted := false, true, true
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		if mounted {
			nui.InputNumber(c, nui.App(c, nui.AppConfig{}), &value, state, nui.InputNumberOptions{Label: "Repeat", Increment: nui.Set(shown), IncrementDisabled: disabled, OnValueChange: func(nui.Optional[float64]) { calls++ }}).Width(300)
		}
	}, 360, 100)
	r := bounds(t, tt, "Increase Repeat")
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(400 * time.Millisecond)
	tt.Advance(60 * time.Millisecond)
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	if value.Value != 8 || calls != 3 || !tt.Focused("Repeat") {
		t.Fatal("press repeat, release count or editor focus")
	}
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	disabled = true
	tt.Frame()
	tt.Advance(time.Second)
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	if value.Value != 9 || calls != 4 {
		t.Fatal("disabled hold kept repeating")
	}
	shown = false
	tt.Frame()
	if _, ok := tt.Find("Increase Repeat"); ok {
		t.Fatal("hidden increment still present")
	}
	editor := bounds(t, tt, "Repeat")
	tt.Scroll(editor.X+5, editor.Y+5, 0, 20)
	if value.Value != 10 || calls != 5 {
		t.Fatal("focused wheel did not step once")
	}
	mounted = false
	tt.Frame()
	tt.Advance(time.Second)
	if calls != 5 {
		t.Fatal("unmounted number continued its work")
	}
}

func TestInputNumberOrientationAccessories(t *testing.T) {
	value := nui.Set(5.0)
	var state nui.NumberState
	orientation := nui.Horizontal
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputNumber(c, th, &value, &state, nui.InputNumberOptions{Label: "Amount", Leading: lucide.Search(), Trailing: lucide.Check(), Clearable: true, Style: func(root, editor, leading, trailing ui.Element) {
				leading.Label("Leading icon")
				trailing.Label("Trailing icon")
			}, Orientation: orientation}).Width(340)
		})
	}, 400, 160)
	for _, direction := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
		orientation = direction
		tt.Frame()
		for _, part := range []string{"Amount", "Leading icon", "Trailing icon", "Clear Amount"} {
			r := bounds(t, tt, part)
			for _, action := range []string{"Decrease Amount", "Increase Amount"} {
				b := bounds(t, tt, action)
				if r.X < b.X+b.W && b.X < r.X+r.W && r.Y < b.Y+b.H && b.Y < r.Y+r.H {
					t.Fatalf("%s overlaps %s in orientation %d", part, action, orientation)
				}
			}
		}
	}
	click(t, tt, "Clear Amount")
	if value.Set || state.Text != "" || !tt.Focused("Amount") {
		t.Fatal("clear accessory lost its behavior")
	}
}

func TestInputNumberOrientationVisuals(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		values := [5][2]nui.Optional[float64]{}
		states := [5][2]nui.NumberState{}
		for i := range values {
			values[i] = [2]nui.Optional[float64]{nui.Set(5.0), nui.Set(5.0)}
		}
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.Light()
			if dark {
				th = nui.Dark()
			}
			nui.Window(c, th)
			ui.Column(c).Padding(24).Gap(16).Children(func() {
				ui.Text(c, "InputNumber · horizontal / vertical").FontSize(20)
				for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
					ui.Row(c.Key(i)).Gap(20).AlignItems(ui.Center).Children(func() {
						ui.Text(c, []string{"XS", "SM", "MD", "LG", "XL"}[i]).Width(24)
						for j, orientation := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
							nui.InputNumber(c.Key(j), th, &values[i][j], &states[i][j], nui.InputNumberOptions{Label: fmt.Sprintf("Number %d %d", i, j), Size: size, Orientation: orientation}).Width(280)
						}
					})
				}
			})
		}, 720, 380)
		f, err := os.Create(filepath.Join(dir, fmt.Sprintf("input-number-orientation-dark-%v.png", dark)))
		if err != nil {
			t.Fatal(err)
		}
		encodeErr := png.Encode(f, tt.Image())
		closeErr := f.Close()
		if encodeErr != nil || closeErr != nil {
			t.Fatalf("save orientation image: encode=%v close=%v", encodeErr, closeErr)
		}
	}
}
