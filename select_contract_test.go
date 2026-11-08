package nui_test

import (
	"fmt"
	"image/color"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestSelectNeutralFocusUsesInvertedBackground(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(fmt.Sprint(multiple), func(t *testing.T) {
			value, values := "a", []string{"a"}
			accent := ui.RGB(126, 34, 206)
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				th.InvertedBackground, th.HighContrast = accent, true
				ui.Column(c).Padding(16).Children(func() {
					opts := nui.SelectOptions{Label: "Choice", Color: nui.Neutral, Decoration: nui.Decoration{Style: func(part nui.Slot, e ui.Element) {
						if part == nui.Root {
							e.Label("Select trigger")
						}
					}}}
					items := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
					if multiple {
						nui.SelectMultiple(c, th, &values, items, opts).Width(240)
					} else {
						nui.Select(c, th, &value, items, opts).Width(240)
					}
				})
			}, 300, 120)
			tt.Key(0, ui.KeyTab)
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
			o := nui.SelectOptions{Label: "Choice", Avatar: &avatar, Open: &open, OnChange: func() { changes++ }, Decoration: nui.Decoration{UI: nui.Styles{nui.Trailing: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(part nui.Slot, e ui.Element) {
				if part == nui.Trailing {
					e.Label("Choice trailing")
				}
			}}}
			items := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "blocked", Label: "Blocked", Disabled: true}, {Value: "b", Label: "Beta"}}
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				th.Motion.Reduce = true
				nui.FormField(c, th, nui.FormFieldOptions{Label: "Outer field"}, func(ft nui.Theme) {
					if multiple {
						nui.SelectMultiple(c, ft, &values, items, o).Width(260)
					} else {
						nui.Select(c, ft, &value, items, o).Width(260)
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
			if changes != 1 || !multiple && (value != "b" || open) || multiple && (!slices.Equal(values, []string{"a", "b"}) || !open) {
				t.Fatalf("loading choice: %q %v open=%v changes=%d", value, values, open, changes)
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
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
					th.Motion.Reduce = true
					th.InvertedBackground = accent
					ui.Column(c).Padding(16).Children(func() {
						o := nui.SelectOptions{Label: "Choice", Color: nui.Neutral, Variant: nui.None, Highlight: true, Open: &open, Leading: lucide.Star(), Trailing: lucide.Plus(), SelectedIcon: marker, Decoration: nui.Decoration{UI: nui.Styles{nui.Item: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(p nui.Slot, e ui.Element) {
							if p == nui.Leading {
								e.Label("Leading icon")
							}
							if p == nui.Trailing {
								e.Label("Trailing icon")
							}
							if p == nui.Item {
								e.Label("Popup row")
							}
						}}}
						items := []nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First choice"}}
						if multiple {
							nui.SelectMultiple(c, th, &values, items, o).Width(260)
						} else {
							nui.Select(c, th, &value, items, o).Width(260)
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
	tt := ui.NewTester(func(c *ui.Context) {
		nui.SelectMultiple(c, nui.App(c, nui.AppConfig{}), &values, items, nui.SelectOptions{Label: "Choice", Open: &open, OnChange: func() { changes++ }}).Width(260)
	}, 360, 320)
	click(t, tt, "Choice")
	tt.Type("be")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(values, []string{"b"}) || changes != 1 {
		t.Fatalf("prefix be: %v/%d", values, changes)
	}
	tt.Advance(2 * time.Second)
	tt.Type("b")
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(values, []string{"b", "c"}) || changes != 2 {
		t.Fatalf("cycle b: %v/%d", values, changes)
	}
	items = []nui.Option[string]{items[3], items[0], items[2], items[1]}
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(values, []string{"b"}) || changes != 3 {
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
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Box(c).Disabled(disabled).Children(func() {
			nui.SelectMultiple(c, th, &values, items, nui.SelectOptions{Label: "Choice", Placeholder: "Pick a choice", Open: &open, OnChange: func() { changes++ }}).Width(260)
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
	if changes != 0 || !slices.Equal(values, []string{"a", "b"}) {
		t.Fatal("inherited disable allowed an open row selection")
	}
	disabled = false
	values = []string{"a"}
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
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Select(c, nui.App(c, nui.AppConfig{}), &value, items, nui.SelectOptions{Label: "Choice", Disabled: disabled, OnChange: func() { changes++ }}).Width(240)
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
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				o := nui.SelectOptions{Label: "Choice", Open: &open, Popup: nui.PopoverOptions{Dismissible: nui.Set(false)}, OnChange: func() { changes++ }}
				items := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
				if multiple {
					nui.SelectMultiple(c, th, &values, items, o).Width(240)
				} else {
					nui.Select(c, th, &value, items, o).Width(240)
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

func TestSelectMultiplePanelBackgroundSurvivesThemeWrapper(t *testing.T) {
	values, open := []string{"a"}, true
	accent := ui.RGB(255, 0, 255)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		ui.Column(c).Padding(16).Children(func() {
			nui.SelectMultiple(c, th, &values, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.SelectOptions{Label: "Choice", Open: &open, Popup: nui.PopoverOptions{Label: "Selection panel"}, Decoration: nui.Decoration{UI: nui.Styles{nui.Panel: {Background: nui.Set(accent)}}}}).Width(240)
		})
	}, 320, 240)
	r := bounds(t, tt, "Selection panel")
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+2)); got != rgba(accent) {
		t.Fatalf("panel background was covered: %v, panel=%v", got, r)
	}
}
