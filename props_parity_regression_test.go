package nui_test

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func parityScreenshot(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("close screenshot %s: %v", f.Name(), err)
		}
	}()
	if err = png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

// Regressions against the pinned Nuxt UI / Reka UI behavior.
func TestParityInputMenuEscapeReset(t *testing.T) {
	var s nui.SearchState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Search"}).Width(300)
	}, 500, 400)
	click(t, tt, "Search")
	tt.Type("Al")
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if s.Query != "" {
		t.Fatalf("Escape left Query=%q, focus=%v; expected reset", s.Query, tt.Focused("Search"))
	}
}

func TestParityInputMenuDeselectionIndicator(t *testing.T) {
	var s nui.SearchState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, []nui.Option[string]{{Value: "b", Label: "Beta"}}, nui.InputMenuOptions{Label: "Search", Multiple: true}).Width(300)
	}, 500, 400)
	click(t, tt, "Search")
	tt.Key(0, ui.KeyDown)
	click(t, tt, "Beta")
	click(t, tt, "Remove Beta")
	click(t, tt, "Search")
	tt.Key(0, ui.KeyDown)
	tt.Move(490, 390)
	tt.Advance(time.Second)
	if len(s.SelectedValues) != 0 {
		t.Fatal(s.SelectedValues)
	}
	r := bounds(t, tt, "Beta")
	green := func() int {
		n := 0
		im := tt.Image()
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			for x := int(r.X); x < int(r.X+r.W); x++ {
				rr, g, b, _ := im.At(x, y).RGBA()
				if g > rr+12000 && g > b+4000 {
					n++
				}
			}
		}
		return n
	}
	before := green()
	parityScreenshot(t, tt, "inputmenu-deselected")
	s.Selected = ""
	tt.Frame()
	after := green()
	parityScreenshot(t, tt, "inputmenu-deselected-cleared")
	if before != after {
		t.Fatalf("empty SelectedValues still had checkmark: green pixels=%d; clearing stale Selected removes it (%d)", before, after)
	}
}

func TestParityInputMenuCreateAlwaysExactMatch(t *testing.T) {
	var s nui.SearchState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, []nui.Option[string]{{Value: "Alpha", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Search", CreateWhenAlways: true, OnCreate: func(string) {}}).Width(300)
	}, 500, 400)
	click(t, tt, "Search")
	tt.Type("Alpha")
	if tt.HasText("Create") {
		t.Fatalf("exact existing value still offers creation: %v", tt.Texts())
	}
}

func TestParitySelectMenuReselectCloses(t *testing.T) {
	s := nui.SelectMenuState{}
	selected := []string{"a"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.SelectMenu(c, th, &selected, &s, []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, nui.SelectMenuOptions{Label: "Choose"}).Width(300)
	}, 500, 500)
	click(t, tt, "Choose")
	tt.Type("Al")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(time.Second)
	if s.Open || s.Query != "" {
		t.Fatalf("reselect: Open=%v Query=%q selected=%v", s.Open, s.Query, selected)
	}
}

func TestParityTagsTabRetainsFocus(t *testing.T) {
	var s nui.TagsState
	var tags []string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags", AddOnTab: true})
			nui.Button(c, th, "Next", nui.ButtonOptions{})
		})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.Type("alpha")
	tt.Key(0, ui.KeyTab)
	if !slices.Equal(tags, []string{"alpha"}) || !tt.Focused("Tags") {
		t.Fatalf("tags=%v inputFocused=%v nextFocused=%v", tags, tt.Focused("Tags"), tt.Focused("Next"))
	}
}

func TestParityTagsMaxLengthCapsDraft(t *testing.T) {
	var s nui.TagsState
	var tags []string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags", MaxLength: 3})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.Type("abcd")
	if s.Query != "abc" {
		t.Fatalf("maxlength=3 draft=%q; expected abc", s.Query)
	}
}

func TestParityTagsRejectedPasteDoesNotInsertDraft(t *testing.T) {
	var s nui.TagsState
	tags := []string{"alpha"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags", AddOnPaste: nui.Set(true)})
	}, 500, 300)
	click(t, tt, "Tags")
	tt.SetClipboard("alpha")
	tt.Command("paste")
	if s.Query != "" {
		t.Fatalf("duplicate rejected but pasted into draft: %q", s.Query)
	}
}

func TestParityListboxSingleToggle(t *testing.T) {
	var s nui.ListboxState
	selected := []string{"a"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Listbox(c, th, &selected, &s, []nui.ListboxItem{{ID: "a", Label: "Alpha"}}, nui.ListboxOptions{Label: "Choices", SelectionBehavior: nui.SelectionToggle})
	}, 500, 300)
	click(t, tt, "Alpha")
	if len(selected) != 0 {
		t.Fatalf("toggle on selected single item retained %v", selected)
	}
}

func TestParitySliderLargeKeyboardSteps(t *testing.T) {
	for _, inverted := range []bool{false, true} {
		for _, key := range []ui.Key{ui.KeyPageUp, ui.KeyRight} {
			t.Run(fmt.Sprintf("inverted=%v/key=%v", inverted, key), func(t *testing.T) {
				value := float64(40)
				var h ui.Handle
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Min: nui.Set(float64(0)), Max: nui.Set(float64(100)), Step: 2, Inverted: inverted}).Bind(&h)
					h.Focus(c)
				}, 500, 300)
				mods := ui.Modifiers(0)
				want := float64(60)
				if key == ui.KeyRight {
					mods = ui.Shift
					if inverted {
						want = 20
					}
				}
				tt.Key(mods, key)
				if value != want {
					t.Fatalf("value=%v want=%v", value, want)
				}
			})
		}
	}
}

func TestParityRangeCanCrossThumbs(t *testing.T) {
	low, high := float64(20), float64(40)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{Label: "Budget", Min: 0, Max: 100, Step: 1}).Width(400)
	}, 500, 300)
	r := bounds(t, tt, "Budget")
	x := func(v float32) float32 { return r.X + 8 + (r.W-16)*v/100 }
	y := r.Y + r.H/2
	tt.Press(x(20), y)
	tt.Move(x(60), y)
	tt.Release(x(60), y)
	if low != 40 || high != 60 {
		t.Fatalf("drag 20 past 40 to 60 produced [%v,%v]; upstream sorts to [40,60]", low, high)
	}
}

func TestParityRatingVerticalDown(t *testing.T) {
	value := 2
	var h ui.Handle
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Rating", Orientation: nui.Vertical}).Bind(&h)
		h.Focus(c)
	}, 500, 300)
	tt.Key(0, ui.KeyDown)
	if value != 3 {
		t.Fatalf("vertical ArrowDown gives %v; expected next rating 3", value)
	}
}

func TestParitySelectMultipleTheme(t *testing.T) {
	single := "a"
	multiple := []string{"a"}
	options := []nui.Option[string]{{Value: "a", Label: "Alpha"}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{}).WithComponent(nui.ComponentSelect, nui.ComponentTheme{Size: nui.XL})
		ui.Column(c).Gap(20).Children(func() {
			nui.Select(c, th, &single, options, nui.SelectOptions{Label: "Single"})
			nui.SelectMultiple(c, th, &multiple, options, nui.SelectOptions{Label: "Multiple"})
		})
	}, 500, 300)
	a, b := bounds(t, tt, "Single"), bounds(t, tt, "Multiple")
	if a.H != b.H {
		t.Fatalf("ComponentSelect.Size=XL ignored by multiple: single height=%v multiple=%v", a.H, b.H)
	}
}

func TestParityChoiceHiddenCardHasSelectedAppearance(t *testing.T) {
	checked := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Option", Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden}).Width(250)
	}, 500, 300)
	before := slices.Clone(tt.Image().Pix)
	checked = true
	tt.Frame()
	parityScreenshot(t, tt, "checkbox-hidden-selected")
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("card with hidden indicator has identical unchecked and checked render")
	}
}

func TestParityChoiceHighlightUnchecked(t *testing.T) {
	checked, highlight := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Option", Highlight: highlight}).Width(250)
	}, 500, 300)
	before := slices.Clone(tt.Image().Pix)
	highlight = true
	tt.Frame()
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("highlight=true does not change unchecked indicator")
	}
}

func TestParityFormFieldHorizontalCopyPlacement(t *testing.T) {
	value := "abc"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.FormField(c, th, nui.FormFieldOptions{Label: "Workspace", Description: "Description", Help: "Help", Hint: "Hint", Orientation: nui.Set(nui.Horizontal)}, func(th nui.Theme) { nui.Input(c, th, &value, nui.InputOptions{Label: "Editor"}).Width(250) }).Width(450)
	}, 600, 300)
	parityScreenshot(t, tt, "formfield-horizontal")
	a, b := bounds(t, tt, "Description"), bounds(t, tt, "Editor")
	t.Logf("description+help=%+v editor=%+v texts=%v", a, b, tt.Texts())
	if a.X >= b.X || bounds(t, tt, "Help").Y <= b.Y {
		t.Fatal("description belongs below the label; it is merged with help below the editor")
	}
}

func TestParityPopoverNonDismissibleKeepsFocus(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Open", nui.ButtonOptions{}).OnClick(func() { open = true })
		nui.Popover(c, th, a, &open, nui.PopoverOptions{Dismissible: nui.Set(false)}, func(th nui.Theme) { nui.Button(c, th, "Inside", nui.ButtonOptions{}) })
	}, 500, 300)
	click(t, tt, "Open")
	click(t, tt, "Inside")
	tt.Key(0, ui.KeyEscape)
	tt.Advance(time.Second)
	if !open || !tt.Focused("Inside") {
		t.Fatalf("Escape: open=%v insideFocused=%v triggerFocused=%v", open, tt.Focused("Inside"), tt.Focused("Open"))
	}
}

func TestParityManualTabsInitialTabStop(t *testing.T) {
	selected := "b"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.Tabs(c, th, &selected, []nui.TabItem{{ID: "a", Label: "Tab A"}, {ID: "b", Label: "Tab B"}}, nui.TabsOptions{ActivationMode: nui.ActivationManual})
			nui.Button(c, th, "After", nui.ButtonOptions{})
		})
	}, 500, 300)
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Tab B") {
		t.Fatalf("Tab entered A=%v B=%v selected=%s", tt.Focused("Tab A"), tt.Focused("Tab B"), selected)
	}
}

func TestParityInvertedSliderOtherAxis(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(fmt.Sprint(vertical), func(t *testing.T) {
			value := float64(40)
			var h ui.Handle
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Slider(c, th, &value, nui.SliderOptions{Label: "Amount", Step: 2, Inverted: true, Vertical: vertical}).Bind(&h)
				h.Focus(c)
			}, 500, 300)
			key := ui.KeyUp
			if vertical {
				key = ui.KeyRight
			}
			tt.Key(0, key)
			if value != 42 {
				t.Fatalf("inverted=%v vertical=%v key=%v gives %v; expected 42", true, vertical, key, value)
			}
		})
	}
}

func TestParityControlledTooltipHoverAndClick(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Help", nui.ButtonOptions{})
		nui.Tooltip(c, th, a, "Tip", nui.TooltipOptions{Open: &open})
	}, 500, 300)
	r := bounds(t, tt, "Help")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(2 * time.Second)
	if !open {
		t.Error("controlled tooltip does not publish open on hover")
	}
	open = true
	tt.Frame()
	click(t, tt, "Help")
	if open {
		t.Error("controlled tooltip does not close on trigger click")
	}
}

func TestParityInputMenuCreateKeepsSearch(t *testing.T) {
	var s nui.SearchState
	created := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputMenu(c, th, &s, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.InputMenuOptions{Label: "Search", OnCreate: func(v string) { created = v }}).Width(300)
	}, 500, 300)
	click(t, tt, "Search")
	tt.Type("new")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if created != "new" || s.Query != "new" || !tt.HasText("Create") {
		t.Fatalf("created=%q query=%q texts=%v; upstream create prevents selection/close", created, s.Query, tt.Texts())
	}
}
