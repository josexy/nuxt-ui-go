package nui_test

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

// InputRating behavior aligned with Nuxt UI v4.12 (Reka UI 2.11 Rating).

// labelStars names each star "Star i" for lookups; the counter restarts every build.
func labelStars() func(nui.Slot, ui.Element) {
	star := 0
	return func(slot nui.Slot, e ui.Element) {
		switch slot {
		case nui.Root:
			star = 0
		case nui.Item:
			star++
			e.Label(fmt.Sprintf("Star %d", star))
		}
	}
}

func TestInputRatingKeyboardMatchesNuxt(t *testing.T) {
	for _, tc := range []struct {
		name                string
		start               int
		clearable, readOnly bool
		keys                []ui.Key
		want, changes       int
	}{
		{"right from empty rates one", 0, false, false, []ui.Key{ui.KeyRight}, 1, 1},
		{"left from empty stays empty", 0, false, false, []ui.Key{ui.KeyLeft}, 0, 0},
		{"left at one stops", 1, false, false, []ui.Key{ui.KeyLeft}, 1, 0},
		{"right at max stops", 5, false, false, []ui.Key{ui.KeyRight}, 5, 0},
		{"space rates an empty rating one", 0, false, false, []ui.Key{ui.KeySpace}, 1, 1},
		{"space keeps the rating without clearable", 3, false, false, []ui.Key{ui.KeySpace}, 3, 0},
		{"space clears the current star", 3, true, false, []ui.Key{ui.KeySpace}, 0, 1},
		// Nuxt's focus stays on the cleared star; Go's current star of an
		// empty rating is the first, so Space then rates it 1.
		{"space after clearing rates one", 3, true, false, []ui.Key{ui.KeySpace, ui.KeySpace}, 1, 2},
		{"enter does nothing", 3, true, false, []ui.Key{ui.KeyEnter}, 3, 0},
		{"read-only ignores space", 3, true, true, []ui.Key{ui.KeySpace}, 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, changes := tc.start, 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(10).Children(func() {
					nui.Button(c, th, "Before", nui.ButtonOptions{})
					nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", Clearable: tc.clearable, ReadOnly: tc.readOnly, OnChange: func(int) { changes++ }})
				})
			}, 300, 120)
			click(t, tt, "Before")
			tt.Key(0, ui.KeyTab)
			if !tt.Focused("Stars") {
				t.Fatal("the rating did not take focus")
			}
			for _, k := range tc.keys {
				tt.Key(0, k)
			}
			if value != tc.want || changes != tc.changes {
				t.Fatalf("value=%d changes=%d, want %d and %d", value, changes, tc.want, tc.changes)
			}
		})
	}
}

func TestInputRatingRingsCurrentStarAndHugsStars(t *testing.T) {
	value := 3
	var orientation nui.Orientation
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Gap(10).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", Orientation: orientation, Style: labelStars()})
		})
	}, 300, 260)
	// Five 20 DIP stars 2 apart, instead of the column's 260 DIP width.
	if r := bounds(t, tt, "Stars"); r.W != 108 || r.H != 20 {
		t.Fatalf("horizontal rating %v, want 108x20", r)
	}
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	ring := func(label string) bool {
		r := bounds(t, tt, label)
		c := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)-2)
		return c != white
	}
	if !ring("Star 3") || ring("Star 1") || ring("Star 5") {
		t.Fatal("the focus ring is not on the current star alone")
	}
	root := bounds(t, tt, "Stars")
	if c := tt.Image().RGBAAt(int(root.X+root.W)+3, int(root.Y+root.H/2)); c != white {
		t.Fatalf("a ring is still drawn around the whole rating: %v", c)
	}
	value = 0
	tt.Frame()
	if !ring("Star 1") || ring("Star 3") {
		t.Fatal("an empty rating should ring its first star")
	}
	orientation = nui.Vertical
	tt.Frame()
	if r := bounds(t, tt, "Stars"); r.W != 20 || r.H != 108 {
		t.Fatalf("vertical rating %v, want 20x108", r)
	}
}

func TestInputRatingHoverPreview(t *testing.T) {
	value := 3
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			// A 10 DIP gap leaves room between the stars for the pointer.
			nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", Clearable: true, Hoverable: nui.Set(true), Style: labelStars(), UI: nui.Styles{nui.Root: {Gap: nui.Set(float32(10))}}})
		})
	}, 300, 120)
	// A filled star is green (primary), an empty one slate (muted). Judge by
	// hue: the hovered star is scaled, so its edges blend with the background.
	green := func(c color.RGBA) bool { return int(c.G) > int(c.R)+40 && int(c.G) > int(c.B)+20 }
	filled := func(i int) bool {
		return countWhere(tt.Image(), bounds(t, tt, fmt.Sprintf("Star %d", i)), green) > 10
	}
	moveTo := func(x, y float32) {
		tt.Move(x, y)
		tt.Advance(time.Second)
	}
	center := func(label string) (float32, float32) {
		r := bounds(t, tt, label)
		return r.X + r.W/2, r.Y + r.H/2
	}
	moveTo(center("Star 4"))
	if !filled(4) || filled(5) {
		t.Fatal("hovering star 4 does not preview four stars")
	}
	// Between stars 4 and 5 the preview stays, without falling back to the value.
	four := bounds(t, tt, "Star 4")
	moveTo(four.X+four.W+5, four.Y+four.H/2)
	if !filled(4) {
		t.Fatal("the preview flickered back while crossing the gap")
	}
	moveTo(150, 100)
	if filled(4) || !filled(3) {
		t.Fatal("leaving the rating did not restore the value")
	}
	// Clearing the current star shows an empty rating while the pointer stays.
	x, y := center("Star 3")
	moveTo(x, y)
	tt.ClickAt(x, y)
	tt.Advance(time.Second)
	if value != 0 || filled(1) || filled(3) {
		t.Fatalf("after clearing: value %d, star 1 filled %v, star 3 filled %v", value, filled(1), filled(3))
	}
	moveTo(center("Star 2"))
	if !filled(2) || filled(3) {
		t.Fatal("entering another star after clearing does not preview it")
	}
}

func TestInputRatingStarsGrowOnHover(t *testing.T) {
	value := 0
	readOnly := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputRating(c, th, &value, nui.InputRatingOptions{Label: "Stars", ReadOnly: readOnly, Style: labelStars(), UI: nui.Styles{nui.Root: {Gap: nui.Set(float32(10))}}})
		})
	}, 300, 120)
	// extent is the width of the star's drawing, inked pixels included.
	extent := func() int {
		r := bounds(t, tt, "Star 2")
		lo, hi := 1<<30, -1
		for y := int(r.Y) - 3; y < int(r.Y+r.H)+3; y++ {
			for x := int(r.X) - 3; x < int(r.X+r.W)+3; x++ {
				if tt.Image().RGBAAt(x, y) != white {
					lo, hi = min(lo, x), max(hi, x)
				}
			}
		}
		return hi - lo
	}
	rest := extent()
	r := bounds(t, tt, "Star 2")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(time.Second)
	if grown := extent(); grown <= rest {
		t.Fatalf("hovered star spans %d DIPs, unhovered %d; want it scaled up", grown, rest)
	}
	readOnly = true
	tt.Frame()
	tt.Advance(time.Second)
	if got := extent(); got != rest {
		t.Fatalf("a read-only star grows on hover: %d, want %d", got, rest)
	}
}

func TestInputRatingFieldErrorColor(t *testing.T) {
	value := 3
	blue := ui.RGB(0, 0, 255)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			// Recolor the error text so that only the stars can use the error color.
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Score", Error: "Too low", UI: nui.Styles{nui.ErrorSlot: {Foreground: nui.Set(blue)}}}, func(th nui.Theme) {
				nui.InputRating(c, th, &value, nui.InputRatingOptions{})
			})
		})
	}, 300, 200)
	th := nui.Light()
	all := ui.Rect{W: 300, H: 200}
	if countColor(tt.Image(), all, rgba(th.Error)) == 0 || countColor(tt.Image(), all, rgba(th.Primary)) != 0 {
		t.Fatal("the rated stars of an invalid field are not drawn in the error color")
	}
}
