package nui_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

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
