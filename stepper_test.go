package nui_test

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// Reference geometry: nuxt/ui src/theme/stepper.ts, default md.
func TestStepperReferenceLayout(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, rtl := range []bool{false, true} {
			for _, width := range []int{900, 330} {
				name := fmt.Sprintf("vertical-%v-rtl-%v-width-%d", vertical, rtl, width)
				t.Run(name, func(t *testing.T) {
					value := "address"
					items := []nui.StepItem{
						{ID: "address", Title: "Address", Description: "Add your address here", Icon: lucide.House()},
						{ID: "shipping", Title: "Shipping", Description: "Set your preferred shipping method", Icon: lucide.Truck()},
						{ID: "checkout", Title: "Checkout", Description: "Confirm your order"},
					}
					tt := ui.NewTester(func(c *ui.Context) {
						theme := nui.Light()
						nui.Window(c, theme)
						root := ui.Column(c).Padding(24)
						if rtl {
							root.Direction(ui.RightToLeft)
						}
						root.Children(func() {
							counts := map[nui.Slot]int{}
							nui.Stepper(c, theme, &value, items, nui.StepperOptions{Vertical: vertical, Style: func(slot nui.Slot, e ui.Element) {
								if slot != nui.Trigger {
									e.Label(fmt.Sprintf("%s-%d", slot, counts[slot]))
								}
								counts[slot]++
							}})
						})
					}, width, 420)
					for _, scale := range []float32{1, 1.5} {
						tt.SetScale(scale)
						for i, item := range items {
							circle := bounds(t, tt, item.Title)
							body := bounds(t, tt, fmt.Sprintf("body-%d", i))
							cell := bounds(t, tt, fmt.Sprintf("item-%d", i))
							if circle.W != 40 || circle.H != 40 {
								t.Fatalf("md indicator should be 40 DIP: %v", circle)
							}
							near := func(got, want float32) {
								t.Helper()
								if math.Abs(float64(got-want)) > .1 {
									t.Fatalf("geometry: got %g, want %g (circle=%v, body=%v)", got, want, circle, body)
								}
							}
							if vertical {
								if rtl {
									near(circle.X+circle.W, float32(width-24))
								} else {
									near(circle.X, 24)
								}
								near(body.Y, circle.Y+10)
								if rtl {
									near(body.X+body.W+10, circle.X)
								} else {
									near(body.X, circle.X+circle.W+10)
								}
							} else {
								near(circle.Y, 24)
								near(circle.X+20, cell.X+cell.W/2)
								near(cell.W, float32(width-48)/3)
								near(body.Y, circle.Y+50)
							}
							if i == len(items)-1 {
								if _, ok := tt.Find("separator-2"); ok {
									t.Fatal("last step must not have a connector")
								}
								continue
							}
							line := bounds(t, tt, fmt.Sprintf("separator-%d", i))
							next := bounds(t, tt, items[i+1].Title)
							if vertical {
								near(line.X+1, circle.X+20)
								near(line.Y, circle.Y+46)
								near(line.Y+line.H, next.Y-6)
							} else {
								near(line.Y, circle.Y+18)
								near(line.H, 2)
								left, right := circle, next
								if rtl {
									left, right = next, circle
								}
								near(line.X, left.X+48)
								near(line.X+line.W, right.X-8)
							}
						}
					}
					if dir := os.Getenv("NUI_ARTIFACTS"); dir != "" {
						if err := os.MkdirAll(dir, 0755); err != nil {
							t.Fatal(err)
						}
						f, err := os.Create(filepath.Join(dir, "stepper-"+name+".png"))
						if err != nil {
							t.Fatal(err)
						}
						defer func() {
							if err := f.Close(); err != nil {
								t.Errorf("close screenshot %s: %v", f.Name(), err)
							}
						}()
						if err := png.Encode(f, tt.Image()); err != nil {
							t.Fatal(err)
						}
					}
				})
			}
		}
	}
}

func TestStepperSizesAndCompletedIndicators(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			value := "last"
			icons := 0
			tt := ui.NewTester(func(c *ui.Context) {
				icons = 0
				theme := nui.Light().WithComponent(nui.ComponentStepper, nui.ComponentTheme{Size: size})
				nui.Window(c, theme)
				nui.Stepper(c, theme, &value, []nui.StepItem{{ID: "number", Title: "Number"}, {ID: "icon", Title: "Icon", Icon: lucide.House()}, {ID: "last", Title: "Last"}}, nui.StepperOptions{Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Leading {
						icons++
					}
				}})
			}, 600, 200)
			if circle := bounds(t, tt, "Number"); circle.W != float32(24+i*8) || circle.W != circle.H {
				t.Fatalf("theme size %v: %v", size, circle)
			}
			if !tt.HasText("1") || icons != 1 {
				t.Fatal("completed steps should retain their number or icon")
			}
		})
	}
}

func TestStepperKeyboardAndRemount(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(fmt.Sprint(vertical), func(t *testing.T) {
			value, changes, shown := "one", 0, true
			tt := ui.NewTester(func(c *ui.Context) {
				theme := nui.App(c, nui.AppConfig{})
				if shown {
					nui.Stepper(c.Key("steps"), theme, &value, []nui.StepItem{{ID: "one", Title: "One"}, {ID: "two", Title: "Two"}, {ID: "three", Title: "Three", Disabled: true}}, nui.StepperOptions{Vertical: vertical, Linear: true, OnChange: func(string) { changes++ }})
				}
			}, 600, 300)
			tt.Key(0, ui.KeyTab)
			key := ui.KeyRight
			if vertical {
				key = ui.KeyDown
			}
			tt.Key(0, key)
			if !tt.Focused("Two") || value != "one" {
				t.Fatal("arrow should move focus without selecting")
			}
			tt.Key(0, ui.KeyEnter)
			if value != "two" || changes != 1 {
				t.Fatal("Enter should select the focused step once")
			}
			click(t, tt, "Three")
			if value != "two" {
				t.Fatal("disabled step changed the selection")
			}
			shown = false
			tt.Frame()
			tt.Frame()
			shown = true
			tt.Frame()
			click(t, tt, "One")
			if value != "one" || changes != 2 {
				t.Fatal("selection did not survive remount")
			}
		})
	}
}

func TestStepperConnectorColors(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprint(dark), func(t *testing.T) {
			value := "one"
			theme := nui.Light()
			if dark {
				theme = nui.Dark()
			}
			tt := ui.NewTester(func(c *ui.Context) {
				nui.Window(c, theme)
				ui.Column(c).Padding(24).Children(func() {
					nui.Stepper(c, theme, &value, []nui.StepItem{{ID: "one", Title: "One"}, {ID: "two", Title: "Two"}}, nui.StepperOptions{Color: nui.Secondary, Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.SeparatorSlot {
							e.Label("connector")
						}
					}})
				})
			}, 600, 150)
			checkColor := func(want ui.Color) {
				t.Helper()
				r := bounds(t, tt, "connector")
				got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y))
				if got.R != want.R || got.G != want.G || got.B != want.B {
					t.Fatalf("connector color=%v, want %v", got, want)
				}
			}
			checkColor(theme.Accented)
			click(t, tt, "Two")
			checkColor(theme.Secondary)
			click(t, tt, "One")
			checkColor(theme.Accented)
		})
	}
}
