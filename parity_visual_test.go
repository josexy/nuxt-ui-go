package nui_test

import (
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

// Per-component visual alignment with Nuxt UI v4.12.

// near reports whether two colors differ by at most tol in every channel,
// allowing for anti-aliased edges.
func near(a, b color.RGBA, tol int) bool {
	d := func(x, y uint8) bool { v := int(x) - int(y); return v >= -tol && v <= tol }
	return d(a.R, b.R) && d(a.G, b.G) && d(a.B, b.B)
}

// whiteRun counts the nearly white pixels of row y between x0 and x1.
func whiteRun(im *image.RGBA, y, x0, x1 int) int {
	n := 0
	for x := x0; x < x1; x++ {
		if near(im.RGBAAt(x, y), white, 24) {
			n++
		}
	}
	return n
}

// countWhere counts the pixels of r that satisfy ok.
func countWhere(im *image.RGBA, r ui.Rect, ok func(color.RGBA) bool) int {
	n := 0
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if ok(im.RGBAAt(x, y)) {
				n++
			}
		}
	}
	return n
}

func TestSliderTrackThumbAndFocusRecipe(t *testing.T) {
	value, low, high := 50.0, 20.0, 80.0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.Slider(c, th, &value, nui.SliderOptions{Label: "Volume"}).Width(300)
			nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{Label: "Price"}).Width(300)
		})
	}, 400, 200)
	th := nui.Light()
	track := func(r ui.Rect) int {
		im, x, n := tt.Image(), int(r.X+r.W*.9), 0
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			if im.RGBAAt(x, y) == rgba(th.Accented) {
				n++
			}
		}
		return n
	}
	slider, price := bounds(t, tt, "Volume"), bounds(t, tt, "Price")
	if track(slider) != 8 || track(price) != 8 {
		t.Fatalf("MD tracks are %d and %d DIPs thick, want Nuxt's 8", track(slider), track(price))
	}
	// At 50% the thumb is centered: a 16 DIP white disc inside a 2 DIP primary ring.
	cx, cy := int(slider.X+slider.W/2), int(slider.Y+slider.H/2)
	im := tt.Image()
	if !near(im.RGBAAt(cx+7, cy), white, 8) || !near(im.RGBAAt(cx+9, cy), rgba(th.Primary), 8) || im.RGBAAt(cx+12, cy) != rgba(th.Accented) {
		t.Fatalf("thumb edge pixels %v %v %v; want white disc, outer ring, then track", im.RGBAAt(cx+7, cy), im.RGBAAt(cx+9, cy), im.RGBAAt(cx+12, cy))
	}
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	// The 3 DIP outline starts 2 DIPs out, where the ring is, and reaches 5 DIPs out.
	if im := tt.Image(); !near(im.RGBAAt(cx+9, cy), rgba(th.Primary), 8) || im.RGBAAt(cx+12, cy) == rgba(th.Accented) {
		t.Fatalf("focused thumb pixels %v %v; want the ring, then the outline over the track", im.RGBAAt(cx+9, cy), im.RGBAAt(cx+12, cy))
	}
}

func TestToggleLinesMatchIndicatorContainer(t *testing.T) {
	for size, want := range map[nui.Size]float32{nui.XS: 16, nui.SM: 16, nui.MD: 20, nui.LG: 20, nui.XL: 24} {
		on := false
		name := func(slot nui.Slot, e ui.Element) {
			if slot == nui.Root {
				e.Label("Box")
			}
		}
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Column(c).Padding(10).AlignItems(ui.Start).Children(func() {
				nui.Checkbox(c, th, &on, nui.CheckboxOptions{Size: size, Decoration: nui.Decoration{Style: name}})
				nui.Checkbox(c, th, &on, nui.CheckboxOptions{Size: size, Label: "Labeled"})
			})
		}, 200, 120)
		if h := bounds(t, tt, "Box").H; h != want {
			t.Errorf("size %v: indicator container is %v DIPs, want %v", size, h, want)
		}
		if h := bounds(t, tt, "Labeled").H; h != want {
			t.Errorf("size %v: labeled checkbox is %v DIPs, want one %v DIP line", size, h, want)
		}
	}
}

func TestRadioDotUsesFixedSize(t *testing.T) {
	for size, want := range map[nui.Size]int{nui.SM: 4, nui.LG: 6} {
		value := "a"
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(10).Children(func() {
				nui.RadioGroup(c, th, &value, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.RadioGroupOptions{Size: size})
			})
		}, 200, 80)
		r := bounds(t, tt, "Alpha")
		box := [5]float32{12, 14, 16, 18, 20}[size-nui.XS]
		line := [5]float32{16, 16, 20, 20, 24}[size-nui.XS]
		y := int(r.Y + line/2)
		if got := whiteRun(tt.Image(), y, int(r.X), int(r.X+box)); got < want-1 || got > want {
			t.Errorf("size %v: radio dot is about %d DIPs wide, want %d", size, got, want)
		}
	}
}

func TestChoiceGroupGaps(t *testing.T) {
	for _, tc := range []struct {
		size       nui.Size
		horizontal bool
		want       float32
	}{{nui.SM, false, 2}, {nui.MD, false, 4}, {nui.XL, false, 6}, {nui.MD, true, 8}} {
		var picked []string
		radio := "a"
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Column(c).Padding(10).Gap(40).AlignItems(ui.Start).Children(func() {
				nui.CheckboxGroup(c, th, &picked, []nui.Option[string]{{Value: "a", Label: "One"}, {Value: "b", Label: "Two"}}, nui.CheckboxGroupOptions{Size: tc.size, Horizontal: tc.horizontal})
				nui.RadioGroup(c, th, &radio, []nui.Option[string]{{Value: "a", Label: "Three"}, {Value: "b", Label: "Four"}}, nui.RadioGroupOptions{Size: tc.size, Horizontal: tc.horizontal})
			})
		}, 400, 300)
		for _, pair := range [][2]string{{"One", "Two"}, {"Three", "Four"}} {
			a, b := bounds(t, tt, pair[0]), bounds(t, tt, pair[1])
			gap := b.Y - (a.Y + a.H)
			if tc.horizontal {
				gap = b.X - (a.X + a.W)
			}
			if gap != tc.want {
				t.Errorf("size %v horizontal=%v: %s to %s gap %v, want %v", tc.size, tc.horizontal, pair[0], pair[1], gap, tc.want)
			}
		}
	}
}

func TestChoiceTableJoinsItems(t *testing.T) {
	value := "two"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.RadioGroup(c, th, &value, []nui.Option[string]{{Value: "one", Label: "One"}, {Value: "two", Label: "Two"}, {Value: "three", Label: "Three"}}, nui.RadioGroupOptions{Variant: nui.ChoiceTable}).Width(240)
		})
	}, 300, 300)
	th := nui.Light()
	one, two, three := bounds(t, tt, "One"), bounds(t, tt, "Two"), bounds(t, tt, "Three")
	if two.Y != one.Y+one.H-1 {
		t.Fatalf("unchecked neighbors do not share a border: %v then %v", one, two)
	}
	// The item after the checked one leaves its border out, keeping the checked edge.
	if three.Y != two.Y+two.H {
		t.Fatalf("item after the checked one: %v then %v", two, three)
	}
	im := tt.Image()
	if line := im.RGBAAt(int(two.X+two.W/2), int(two.Y+two.H)-1); line == rgba(th.Border) {
		t.Fatalf("the checked item's bottom edge is covered by the next item's border (%v)", line)
	}
	// Only the group's outer corners are rounded.
	if im.RGBAAt(int(one.X), int(one.Y)) != white || im.RGBAAt(int(two.X), int(two.Y)) == white || im.RGBAAt(int(three.X), int(three.Y+three.H)-1) != white {
		t.Fatal("table corners: want rounded outer corners and square inner ones")
	}
}

func TestFormFieldRequiredMarkAndSpacing(t *testing.T) {
	value := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Name", Required: true, Description: "Shown to others", Error: "Too short"}, func(th nui.Theme) {
				nui.Input(c, th, &value, nui.InputOptions{Style: func(root, editor, leading, trailing ui.Element) { root.Label("Field shell") }}).Width(240)
			})
		})
	}, 400, 300)
	label, star := bounds(t, tt, "Name"), bounds(t, tt, "*")
	if gap := star.X - (label.X + label.W); gap != 2 {
		t.Fatalf("asterisk is %v DIPs after the label, want 2", gap)
	}
	red := func(c color.RGBA) bool { return int(c.R) > int(c.G)+60 && int(c.R) > int(c.B)+60 }
	if countWhere(tt.Image(), star, red) == 0 {
		t.Fatal("the required asterisk is not drawn in the error color")
	}
	if _, found := tt.Find("Name *"); !found {
		t.Fatal("the accessible name lost its required marker")
	}
	description, shell, errText := bounds(t, tt, "Shown to others"), bounds(t, tt, "Field shell"), bounds(t, tt, "Too short")
	if description.Y != label.Y+label.H || shell.Y-(description.Y+description.H) != 4 || errText.Y-(shell.Y+shell.H) != 8 {
		t.Fatalf("label %v, description %v, field %v, error %v; want gaps 0, 4 and 8", label, description, shell, errText)
	}
}

func TestHorizontalChoiceTableFollowsDirection(t *testing.T) {
	value := "a"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).AlignItems(ui.Start).Children(func() {
			nui.RadioGroup(c, th, &value, []nui.Option[string]{{Value: "a", Label: "Left"}, {Value: "b", Label: "Middle"}, {Value: "c", Label: "Right"}}, nui.RadioGroupOptions{Variant: nui.ChoiceTable, Horizontal: true})
		})
	}, 500, 200)
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(dir)
		tt.Frame()
		first, second := bounds(t, tt, "Left"), bounds(t, tt, "Middle")
		ltr := dir == ui.LeftToRight
		// After the checked first item, the second leaves its start border out.
		if ltr && second.X != first.X+first.W || !ltr && second.X+second.W != first.X {
			t.Fatalf("%v: checked %v then %v; want the next item to keep the checked edge", dir, first, second)
		}
		im := tt.Image()
		outer, inner := int(first.X), int(first.X+first.W)-1
		if !ltr {
			outer, inner = inner, outer
		}
		y := int(first.Y)
		if im.RGBAAt(outer, y) != white || im.RGBAAt(inner, y) == white {
			t.Fatalf("%v: the first item should round only its outer corner", dir)
		}
	}
}

func TestModalRegionsAndDividers(t *testing.T) {
	open, withBody := true, true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		var body func(nui.Theme)
		if withBody {
			body = func(nui.Theme) { ui.Text(c, "Body text") }
		}
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Settings", Footer: func(c *ui.Context, th nui.Theme) {
			nui.Button(c, th, "Cancel", nui.ButtonOptions{})
			nui.Button(c, th, "Save", nui.ButtonOptions{})
		}}, body)
	}, 800, 600)
	tt.Advance(time.Second)
	th := nui.Light()
	panel, close, text := bounds(t, tt, "Settings"), bounds(t, tt, "Close Settings"), bounds(t, tt, "Body text")
	if close.W != 32 || close.H != 32 {
		t.Fatalf("close button %v, want MD 32x32", close)
	}
	// The 64 DIP header is followed by a 1 DIP divider, as Nuxt's divide-y
	// border on the next region; the body is padded by 24.
	header := float32(64 + 1)
	if line := tt.Image().RGBAAt(int(panel.X+panel.W/2), int(panel.Y+header)-1); line != rgba(th.Border) {
		t.Fatalf("no divider under the 64 DIP header: %v", line)
	}
	if text.Y-(panel.Y+header) != 24 || text.X-panel.X != 24 {
		t.Fatalf("body text at %v in panel %v; want 24 DIP padding", text, panel)
	}
	if gap := bounds(t, tt, "Save").X - (bounds(t, tt, "Cancel").X + bounds(t, tt, "Cancel").W); gap != 6 {
		t.Fatalf("footer gap %v, want 6", gap)
	}
	height := panel.H
	withBody = false
	tt.Frame()
	if got := bounds(t, tt, "Settings").H; got >= height-24 {
		t.Fatalf("a modal without content keeps a body region: height %v, was %v", got, height)
	}
}

func TestTopSlideoverFitsContent(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Slideover(c, th, &open, nui.SlideoverOptions{ModalOptions: nui.ModalOptions{Title: "Filters"}, Side: nui.SideTop}, func(nui.Theme) { ui.Text(c, "One line") })
	}, 800, 700)
	tt.Advance(time.Second)
	// Header 64, the divider, and the body: 24 + one 20 DIP line + 24.
	if r := bounds(t, tt, "Filters"); r.Y != 0 || r.H != 64+1+68 {
		t.Fatalf("top slideover %v; want its content's 133 DIP height", r)
	}
}

func TestToastRecipe(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() { nui.Toast(c, th, nui.ToastOptions{Title: "Saved", Timeout: time.Hour}) })
		nui.ToastViewport(c, th, nui.ToastViewportOptions{})
	}, 800, 600)
	click(t, tt, "Notify")
	tt.Advance(time.Second)
	panel, close := bounds(t, tt, "Saved"), bounds(t, tt, "Dismiss Saved")
	if panel.W != 384 {
		t.Fatalf("toast is %v DIPs wide, want 384", panel.W)
	}
	if close.Y-panel.Y != 16 || panel.X+panel.W-(close.X+close.W) != 16 {
		t.Fatalf("close button %v in toast %v; want 16 DIP padding", close, panel)
	}
	// The SM progress bar runs along the toast's bottom edge.
	im := tt.Image()
	y := int(panel.Y+panel.H) - 2
	if c := im.RGBAAt(int(panel.X+panel.W/2), y); c == white || c == rgba(nui.Light().Border) {
		t.Fatalf("no progress bar at the toast's bottom edge: %v", c)
	}
}

func TestTooltipSeparatesKeys(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(150).Top(100)
		nui.Tooltip(c, th, a, "Save", nui.TooltipOptions{Open: &open, Kbds: []string{"Ctrl", "S"}})
	}, 500, 300)
	tt.Advance(time.Second)
	dot, save, key := bounds(t, tt, "·"), bounds(t, tt, "Save"), bounds(t, tt, "S")
	if !(save.X+save.W <= dot.X && dot.X+dot.W < key.X) {
		t.Fatalf("want text, middle dot, then keys: %v %v %v", save, dot, key)
	}
}

func TestMenuWidthChecksAndKeys(t *testing.T) {
	var state nui.MenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			anchor := nui.Button(c, th, "View", nui.ButtonOptions{})
			nui.DropdownMenu(c, th, anchor, &state, []nui.MenuItem{{ID: "grid", Label: "Grid", Checked: nui.Set(true)}, {ID: "open", Label: "Open", Shortcut: "Ctrl+O"}}, nui.MenuOptions{Label: "View menu"})
		})
	}, 500, 400)
	click(t, tt, "View")
	tt.Advance(time.Second)
	panel := bounds(t, tt, "View menu")
	if panel.W >= 240 || panel.W < 128 {
		t.Fatalf("menu is %v DIPs wide; want its content's width, at least 128", panel.W)
	}
	grid := bounds(t, tt, "Grid")
	th := nui.Light()
	if countColor(tt.Image(), grid, rgba(th.Primary)) != 0 {
		t.Fatal("the checked item's mark still uses the primary color")
	}
	// Each key of the shortcut is its own chip.
	if !tt.HasText("O") || tt.HasText("Ctrl+O") {
		t.Fatalf("shortcut not split into keys: %v", tt.Texts())
	}
}
