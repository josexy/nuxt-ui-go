package nui_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestPinInputSequentialReplacement(t *testing.T) {
	for _, mask := range []bool{false, true} {
		t.Run(fmt.Sprintf("mask=%v", mask), func(t *testing.T) {
			value := "123456"
			var state nui.PinState
			changes, completes := 0, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				index := 0
				nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Label: "Code", Length: 6, Numeric: true, Mask: mask, SeparatorEvery: 2,
					OnChange: func(string) { changes++ }, OnComplete: func(string) { completes++ },
					Style: func(part nui.Slot, e ui.Element) {
						if part == nui.Item {
							index++
							e.Label(fmt.Sprintf("Digit %d", index))
						}
					},
				})
			}, 360, 100)
			click(t, tt, "Digit 2")
			tt.Type("8")
			tt.Type("9")
			if value != "189456" || changes != 2 || completes != 2 {
				t.Fatalf("sequential replacement shifted later slots: %q, changes=%d completes=%d", value, changes, completes)
			}
			tt.Command("undo")
			if value != "183456" {
				t.Fatalf("undo replacement = %q", value)
			}
			tt.Command("redo")
			if value != "189456" {
				t.Fatalf("redo replacement = %q", value)
			}
			click(t, tt, "Digit 3")
			tt.SetClipboard("72")
			tt.Command("paste")
			if value != "187256" {
				t.Fatalf("pasting over slots shifted the suffix: %q", value)
			}
			tt.Command("undo")
			if value != "189456" {
				t.Fatalf("undo paste = %q", value)
			}
			click(t, tt, "Digit 6")
			tt.Type("7")
			tt.Type("8")
			if value != "189458" {
				t.Fatalf("typing at the final slot = %q", value)
			}
			tt.Key(0, ui.KeyHome)
			tt.Key(0, ui.KeyRight)
			tt.Key(0, ui.KeyLeft)
			tt.Type("6")
			if value != "689458" {
				t.Fatalf("arrow navigation did not replace the first slot: %q", value)
			}
			click(t, tt, "Digit 3")
			tt.Type("72")
			if value != "687258" {
				t.Fatalf("multi-character input shifted the suffix: %q", value)
			}
			click(t, tt, "Digit 3")
			tt.Type("7") // The value stays the same, but focus must advance.
			tt.Type("3")
			if value != "687358" {
				t.Fatalf("unchanged digit failed to advance: %q", value)
			}
			click(t, tt, "Digit 3")
			tt.Type("x")
			tt.Type("4")
			if value != "684358" {
				t.Fatalf("invalid input changed a slot or advanced: %q", value)
			}
		})
	}
}

func TestPinInputVisibleCaret(t *testing.T) {
	value := ""
	var state nui.PinState
	readOnly, disabled, shown, reduceMotion := false, false, true, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = reduceMotion
		ui.Column(c).Padding(24).Gap(16).Children(func() {
			if shown {
				index := 0
				nui.PinInput(c.Key("pin"), th, &value, &state, nui.PinInputOptions{Label: "Code", Length: 4, Placeholder: "○", ReadOnly: readOnly, Disabled: disabled,
					Style: func(part nui.Slot, e ui.Element) {
						if part == nui.Item {
							index++
							e.Label(fmt.Sprintf("Digit %d", index))
						}
					},
				})
			}
			nui.Button(c, th, "After", nui.ButtonOptions{})
		})
	}, 280, 150)
	caretPixels := func(slot int) int {
		r := bounds(t, tt, fmt.Sprintf("Digit %d", slot))
		r = ui.Rect{X: r.X + r.W/2 - 1, Y: r.Y + 6, W: 3, H: r.H - 12}
		return countColor(tt.Image(), r, rgba(nui.Light().Highlighted))
	}
	placeholderPixels := caretPixels(2)
	click(t, tt, "Digit 1")
	parityScreenshot(t, tt, "pin-visible-caret")
	if caretPixels(1) < 10 || caretPixels(2) != placeholderPixels {
		t.Fatalf("focused empty slot needs a visible caret: %d, inactive: %d", caretPixels(1), caretPixels(2))
	}
	tt.Advance(600 * time.Millisecond)
	if caretPixels(1) != 0 {
		t.Fatal("caret did not blink off")
	}
	tt.Type("1")
	if caretPixels(2) < 10 {
		t.Fatal("typing did not advance and reset the visible caret")
	}
	readOnly = true
	tt.Frame()
	if caretPixels(2) != 0 {
		t.Fatal("read-only PIN shows an editing caret")
	}
	readOnly, disabled = false, true
	tt.Frame()
	if caretPixels(2) >= 10 {
		t.Fatal("disabled PIN shows a caret")
	}
	shown = false
	tt.Frame()
	tt.Advance(time.Second)
	shown, disabled = true, false
	tt.Frame()
	click(t, tt, "Digit 2")
	if caretPixels(2) < 10 {
		t.Fatal("remounted PIN did not restore its caret")
	}
	reduceMotion = true
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	if caretPixels(2) < 10 {
		t.Fatal("reduced-motion caret should stay visible")
	}
	click(t, tt, "After")
	if caretPixels(2) >= 10 {
		t.Fatal("blurred PIN shows a caret")
	}
}

func TestPinInputUnicodeReplacement(t *testing.T) {
	value := "A中😀bc"
	var state nui.PinState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		index := 0
		nui.PinInput(c, th, &value, &state, nui.PinInputOptions{Label: "Code", Style: func(part nui.Slot, e ui.Element) {
			if part == nui.Item {
				index++
				e.Label(fmt.Sprintf("Digit %d", index))
			}
		}})
	}, 300, 100)
	click(t, tt, "Digit 2")
	tt.Type("文🎉")
	if value != "A文🎉bc" {
		t.Fatalf("Unicode replacement shifted the suffix: %q", value)
	}
	tt.Type("x")
	if value != "A文🎉xc" {
		t.Fatalf("Unicode replacement advanced by bytes instead of runes: %q", value)
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("中")
	if value != "中" {
		t.Fatalf("select-all replacement retained old slots: %q", value)
	}
}

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
