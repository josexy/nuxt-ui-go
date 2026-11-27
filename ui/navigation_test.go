package nui_test

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestTabsOnlyNavigateTheirOrientation(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, manual := range []bool{false, true} {
			t.Run(fmt.Sprintf("vertical=%v/manual=%v", vertical, manual), func(t *testing.T) {
				value, calls := "a", 0
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					mode := nui.ActivationAutomatic
					if manual {
						mode = nui.ActivationManual
					}
					nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}}, nui.TabsOptions{Vertical: vertical, ActivationMode: mode, OnChange: func(string) { calls++ }})
				}, 500, 300)
				click(t, tt, "First")
				wrong, right := ui.KeyDown, ui.KeyRight
				if vertical {
					wrong, right = right, wrong
				}
				tt.Key(0, wrong)
				if value != "a" || calls != 0 || !tt.Focused("First") {
					t.Fatal("the inactive arrow axis changed tabs")
				}
				tt.Key(0, right)
				if manual {
					if value != "a" || calls != 0 || !tt.Focused("Second") {
						t.Fatal("manual navigation activated content")
					}
					tt.Key(0, ui.KeySpace)
				}
				if value != "b" || calls != 1 {
					t.Fatalf("navigation value/callback=%q/%d", value, calls)
				}
			})
		}
	}
}

func TestTabsPlainTriggerMetrics(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			value := "a"
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}}, nui.TabsOptions{Size: size, Underline: true, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Label {
						e.Label("Tab text")
					}
				}})
			}, 500, 300)
			row, text := bounds(t, tt, "First"), bounds(t, tt, "Tab text")
			if row.H != []float32{24, 28, 32, 36, 40}[i] || text.X-row.X != []float32{8, 10, 12, 12, 12}[i] {
				t.Fatalf("size %v row=%v text=%v", size, row, text)
			}
		})
	}
}

func TestTabsLongLabelsStayInsideTheList(t *testing.T) {
	value := "a"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "Very long first project settings"}, {ID: "b", Label: "Very long second account settings"}}, nui.TabsOptions{Label: "List"}).Width(260)
	}, 300, 200)
	list := bounds(t, tt, "List")
	for _, name := range []string{"Very long first project settings", "Very long second account settings"} {
		r := bounds(t, tt, name)
		if r.X < list.X || r.X+r.W > list.X+list.W {
			t.Fatalf("tab %q exceeds list: tab=%v list=%v", name, r, list)
		}
	}
}

func TestTabsAvatarBadgeItemStylesAndContentSwitch(t *testing.T) {
	value, draft := "a", "Draft"
	content, icon := true, false
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		var leading *ui.SVG
		if icon {
			leading = lucide.Info()
		}
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First", Icon: leading, Avatar: &nui.AvatarOptions{Alt: "Tab avatar", Text: "TA"}, Badge: "0", BadgeOptions: &nui.BadgeOptions{Color: nui.Success, Variant: nui.Soft}, UI: nui.Styles{nui.Trigger: {Background: nui.Set(ui.RGB(255, 0, 255))}}, Content: func(c *ui.Context, th nui.Theme) { nui.Input(c, th, &draft, nui.InputOptions{Label: "Tab draft"}) }}, {ID: "b", Label: "Second", Content: func(c *ui.Context, th nui.Theme) { ui.Text(c, "Second panel") }}}, nui.TabsOptions{Content: nui.Set(content), KeepMounted: true, OnChange: func(string) { changes++ }, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Trailing {
				e.Label("Tab badge")
			}
		}}).Width(460)
		nui.Badge(c, th, "0", nui.BadgeOptions{Size: nui.SM, Color: nui.Success, Variant: nui.Soft}).Label("Standalone badge")
	}, 500, 320)
	if bounds(t, tt, "Tab avatar").W != 20 || !tt.HasText("0") || bounds(t, tt, "Tab badge").H != bounds(t, tt, "Standalone badge").H {
		t.Fatalf("avatar=%v badge=%v texts=%v", bounds(t, tt, "Tab avatar"), bounds(t, tt, "Tab badge"), tt.Texts())
	}
	r := bounds(t, tt, "First")
	if p := tt.Image().RGBAAt(int(r.X+8), int(r.Y+r.H-4)); p.R != 255 || p.G != 0 || p.B != 255 {
		t.Fatalf("item trigger background=%v", p)
	}
	click(t, tt, "Tab draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Edited")
	icon = true
	tt.Frame()
	if tt.HasText("Tab avatar") || !tt.Focused("Tab draft") {
		t.Fatal("icon priority update lost content focus")
	}
	content = false
	tt.Frame()
	if tt.HasText("Tab draft") || !tt.Focused("First") || value != "a" || draft != "Edited" || changes != 0 {
		t.Fatal("content switch did not unmount panels and repair focus without changing value")
	}
	content = true
	tt.Frame()
	if !tt.HasText("Tab draft") || draft != "Edited" {
		t.Fatal("content recreation lost application data")
	}
}

func TestTabsPanelFocusControlledSelectionAndRemoval(t *testing.T) {
	value, draft := "a", "Draft"
	remove, disabled, vertical := false, false, false
	changes, notices := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		items := []nui.TabItem{{ID: "a", Label: "First", Disabled: disabled, Content: func(c *ui.Context, th nui.Theme) { nui.Input(c, th, &draft, nui.InputOptions{Label: "Draft"}) }}, {ID: "b", Label: "Second", Content: func(c *ui.Context, th nui.Theme) { ui.Text(c, "Static panel") }}}
		if remove {
			items = items[1:]
		}
		nui.Tabs(c, th, &value, items, nui.TabsOptions{Vertical: vertical, KeepMounted: true, OnChange: func(string) { changes++ }}).OnChange(func() { notices++ })
	}, 500, 300)
	click(t, tt, "First")
	tt.Key(0, ui.KeyTab)
	// The panel is named after its tab, so find it by role, not by name alone.
	if panel, ok := tt.AccessByRole(ui.RoleTabPanel, "First"); !ok || !panel.Focused {
		t.Fatalf("Tab did not reach the content panel: %+v", panel)
	}
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Draft") {
		t.Fatal("panel input was not the next focus stop")
	}
	value = "b"
	tt.Frame()
	if !tt.Focused("Second") || changes != 0 || notices != 0 {
		t.Fatal("external switch did not restore trigger focus silently")
	}
	click(t, tt, "First")
	if changes != 1 || notices != 1 {
		t.Fatalf("click notifications=%d/%d", changes, notices)
	}
	disabled = true
	tt.Frame()
	if value != "b" || !tt.Focused("Second") || changes != 1 {
		t.Fatal("disabled selection was not normalized with focus repair")
	}
	disabled = false
	tt.Frame()
	click(t, tt, "First")
	remove = true
	tt.Frame()
	if value != "b" || !tt.Focused("Second") {
		t.Fatal("removed trigger did not restore focus")
	}
	vertical = true
	tt.Frame()
	if !tt.Focused("Second") {
		t.Fatal("orientation changed trigger identity")
	}
}

func TestTabsNeutralIndicatorAndPartOverride(t *testing.T) {
	for _, underline := range []bool{false, true} {
		t.Run(fmt.Sprint(underline), func(t *testing.T) {
			value := "a"
			override := false
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				th.InvertedBackground = ui.RGB(255, 0, 255)
				styles := nui.Styles{}
				if override {
					styles[nui.Indicator] = nui.PartStyle{Background: nui.Set(ui.RGB(0, 190, 255))}
				}
				nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}}, nui.TabsOptions{Color: nui.Neutral, Underline: underline, UI: styles, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Indicator {
						e.Label("Indicator")
					}
				}}).Width(300)
			}, 340, 200)
			r := bounds(t, tt, "Indicator")
			x, y := int(r.X+r.W/2), int(r.Y+r.H/2)
			if !underline {
				y = int(r.Y + r.H - 4)
			}
			if p := tt.Image().RGBAAt(x, y); p.R != 255 || p.G != 0 || p.B != 255 {
				t.Fatalf("neutral indicator=%v bounds=%v", p, r)
			}
			if !underline && countColor(tt.Image(), bounds(t, tt, "First"), rgba(nui.Light().Inverted)) == 0 {
				t.Fatal("pill indicator covers the active tab text")
			}
			override = true
			tt.Frame()
			if p := tt.Image().RGBAAt(x, y); p.R != 0 || p.G != 190 || p.B != 255 {
				t.Fatalf("indicator override=%v", p)
			}
		})
	}
}

func TestTabsBackgroundUpdatesDoNotActivateFocusedID(t *testing.T) {
	value := "a"
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}}, nui.TabsOptions{OnChange: func(string) { changes++ }})
	}, 400, 240)
	click(t, tt, "First")
	tt.SetFocused(false)
	value = "b"
	tt.Frame()
	tt.SetFocused(true)
	if value != "b" || changes != 0 {
		t.Fatal("window reactivation overwrote controlled selection")
	}
}

func TestTabsVerticalCenterRTLAndRetainedDraftSelection(t *testing.T) {
	value, draft := "a", "Selected text"
	vertical := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.DarkMode})
		nui.Tabs(c.Key("tabs"), th, &value, []nui.TabItem{{ID: "a", Label: "First", Content: func(c *ui.Context, th nui.Theme) {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Draft"})
			ui.Box(c).Height(80)
		}}, {ID: "b", Label: "Second"}}, nui.TabsOptions{Label: "List", Vertical: vertical, Underline: true, KeepMounted: true, UI: nui.Styles{nui.Indicator: {Background: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Indicator {
				e.Label("Indicator")
			}
		}}).Width(440)
	}, 500, 360)
	first, ok := tt.AccessByRole(ui.RoleTabPanel, "First")
	if !ok {
		t.Fatal("missing the First tab panel")
	}
	list, panel := bounds(t, tt, "List"), first.Bounds
	if list.Y+list.H/2 != panel.Y+panel.H/2 {
		t.Fatalf("vertical regions not centered: list=%v panel=%v", list, panel)
	}
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.SetDirection(ui.RightToLeft)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5})
	line := bounds(t, tt, "Indicator")
	list = bounds(t, tt, "List")
	if line.X+line.W != list.X+list.W {
		t.Fatalf("RTL underline did not use start edge: line=%v list=%v", line, list)
	}
	vertical = false
	tt.Frame()
	tt.Type("Replacement")
	if draft != "Replacement" || !tt.Focused("Draft") {
		t.Fatalf("layout/scaling lost editor selection: %q", draft)
	}
}

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
					nui.Stepper(c.Key("steps"), theme, &value, []nui.StepItem{{ID: "one", Title: "One"}, {ID: "two", Title: "Two"}, {ID: "three", Title: "Three", Disabled: true}}, nui.StepperOptions{Vertical: vertical, OnChange: func(string) { changes++ }})
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
