package nui_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestPinInputSeparatorLayout(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		for _, scale := range []float32{1, 1.5} {
			for _, size := range []nui.Size{nui.XS, nui.MD, nui.XL} {
				t.Run(fmt.Sprintf("rtl=%v/scale=%v/size=%v", rtl, scale, size), func(t *testing.T) {
					value := "223233"
					var state nui.PinState
					grouped := true
					tt := ui.NewTester(func(c *ui.Context) {
						th := nui.App(c, nui.AppConfig{})
						th.TextScale, th.Density = scale, scale
						column := ui.Column(c).Padding(24)
						if rtl {
							column.Direction(ui.RightToLeft)
						}
						column.Children(func() {
							index := 0
							o := nui.PinInputOptions{Label: "Code", Length: 6, Size: size, Style: func(part nui.Slot, e ui.Element) {
								switch part {
								case nui.Item:
									index++
									e.Label(fmt.Sprintf("Digit %d", index))
								case nui.Root:
									e.Label("PIN layout")
								}
							}}
							if grouped {
								o.SeparatorEvery = 2
								// Duplicates and invalid positions must not add gaps.
								o.Separator = []int{0, 2, 2, 6, 8}
							}
							nui.PinInput(c.Key("pin"), th, &value, &state, o)
						})
					}, 600, 180)
					if !tt.HasText("•") || tt.HasText("−") {
						t.Fatal("group separators should be bullets")
					}
					bullet := bounds(t, tt, "•")
					first := bounds(t, tt, "Digit 1")
					if math.Abs(float64(bullet.Y+bullet.H/2-first.Y-first.H/2)) > 1 {
						t.Fatalf("separator is not vertically centered: bullet=%+v slot=%+v", bullet, first)
					}
					for i := 1; i < 6; i++ {
						a, b := bounds(t, tt, fmt.Sprintf("Digit %d", i)), bounds(t, tt, fmt.Sprintf("Digit %d", i+1))
						gap := b.X - a.X - a.W
						if rtl {
							gap = a.X - b.X - b.W
						}
						want := 6 * scale
						if i%2 == 0 {
							want = 12*scale + bullet.W
						}
						if math.Abs(float64(gap-want)) > 1 {
							t.Fatalf("gap after slot %d = %v, want %v", i, gap, want)
						}
					}
					last, root := bounds(t, tt, "Digit 6"), bounds(t, tt, "PIN layout")
					if want := max(first.X+first.W, last.X+last.W) - min(first.X, last.X); math.Abs(float64(root.W-want)) > 1 {
						t.Fatalf("root width %v does not fit slots %v", root.W, want)
					}
					parityScreenshot(t, tt, fmt.Sprintf("pin-separator-rtl-%v-scale-%v-size-%v", rtl, scale, size))
					click(t, tt, "Digit 3")
					grouped = false
					tt.Frame()
					tt.Type("9")
					if value != "229233" || !tt.Focused("Code") {
						t.Fatal("changing grouping lost the editor's focus or selection")
					}
				})
			}
		}
	}
}

func TestPinInputFocusAppearance(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, variant := range []nui.Variant{nui.Outline, nui.Subtle, nui.Soft, nui.Ghost, nui.None} {
			t.Run(fmt.Sprintf("dark=%v/variant=%v", dark, variant), func(t *testing.T) {
				value := "123456"
				var state nui.PinState
				invalid, disabled := false, false
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.Light()
					if dark {
						th = nui.Dark()
					}
					nui.Window(c, th)
					ui.Column(c).Padding(24).Gap(16).Children(func() {
						nui.Button(c, th, "Before", nui.ButtonOptions{})
						errorText := ""
						if invalid {
							errorText = "Wrong code"
						}
						nui.FormField(c, th, nui.FormFieldOptions{Error: errorText}, func(th nui.Theme) {
							index := 0
							ui.Column(c).Disabled(disabled).Children(func() {
								nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Label: "Code", Length: 6, SeparatorEvery: 2, Variant: variant, Style: func(part nui.Slot, e ui.Element) {
									if part == nui.Item {
										index++
										e.Label(fmt.Sprintf("Digit %d", index))
									}
								}})
							})
						})
						nui.Button(c, th, "After", nui.ButtonOptions{})
					})
				}, 380, 240)
				idle := fieldPixels(t, tt, "Digit 3")
				before := bounds(t, tt, "Digit 3")
				click(t, tt, "Digit 3")
				focused := fieldPixels(t, tt, "Digit 3")
				if (focused[1] != idle[1]) != (variant != nui.None) {
					t.Fatalf("pointer focus outline: idle=%v focused=%v", idle, focused)
				}
				if fieldPixels(t, tt, "Digit 2")[1] != idle[1] || bounds(t, tt, "Digit 3") != before {
					t.Fatal("focus changed another slot's outline or shifted the layout")
				}
				parityScreenshot(t, tt, fmt.Sprintf("pin-focus-dark-%v-variant-%v", dark, variant))
				tt.Key(0, ui.KeyTab)
				tt.Move(370, 230)
				if !tt.Focused("After") || fieldPixels(t, tt, "Digit 3") != idle {
					t.Fatal("Tab must leave the PIN and clear its focus appearance")
				}
				tt.Key(ui.Shift, ui.KeyTab)
				tt.Key(0, ui.KeyHome)
				if fieldPixels(t, tt, "Digit 1") != focused {
					t.Fatal("keyboard focus must use the same appearance as pointer focus")
				}
				invalid = true
				tt.Frame()
				if variant != nui.None && fieldPixels(t, tt, "Digit 1")[1] == focused[1] {
					t.Fatal("invalid PIN must use the error focus outline")
				}
				disabled = true
				tt.Frame()
				if fieldPixels(t, tt, "Digit 1")[1] != idle[1] {
					t.Fatal("disabled PIN retained its focus outline")
				}
			})
		}
	}
}
