package nui_test

import (
	"image"
	"image/color"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
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
				nui.Checkbox(c, th, &on, nui.CheckboxOptions{Size: size, Style: name})
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
		nui.Slideover(c, th, &open, nui.SlideoverOptions{Title: "Filters", Side: nui.SideTop}, func(nui.Theme) { ui.Text(c, "One line") })
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

func TestPaginationIconControlsAndSquares(t *testing.T) {
	page := 5
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).AlignItems(ui.Start).Children(func() {
			nui.Pagination(c, th, &page, nui.PaginationOptions{Total: 200, PageSize: 10, SiblingCount: nui.Set(1), ShowEdges: nui.Set(true)})
		})
	}, 700, 120)
	for _, label := range []string{"First page", "Previous page", "Next page", "Last page", "Page 5"} {
		if r := bounds(t, tt, label); r.W != 32 || r.H != 32 {
			t.Errorf("%s is %v; want a 32 DIP square", label, r)
		}
	}
	if texts := tt.Texts(); slices.Contains(texts, "First") || slices.Contains(texts, "Next") {
		t.Fatalf("controls still show text labels: %v", tt.Texts())
	}
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(dir)
		tt.Frame()
		im := tt.Image()
		inner := func(r ui.Rect) ui.Rect { return ui.Rect{X: r.X + 6, Y: r.Y + 6, W: r.W - 12, H: r.H - 12} }
		if chevronPointsRight(t, im, inner(bounds(t, tt, "Next page"))) != (dir == ui.LeftToRight) {
			t.Errorf("%v: the next page chevron should point toward the line's end", dir)
		}
	}
}

func TestBreadcrumbRecipe(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).AlignItems(ui.Start).Children(func() {
			nui.Breadcrumb(c, th, []nui.BreadcrumbItem{{ID: "home", Label: "Home", OnSelect: func() {}}, {ID: "here", Label: "Here"}}, nui.BreadcrumbOptions{})
		})
	}, 400, 100)
	home, here := bounds(t, tt, "Home"), bounds(t, tt, "Here")
	if home.H != 20 {
		t.Fatalf("ancestor link is %v DIPs tall; want an unpadded 14/20 line", home.H)
	}
	// A 20 DIP separator with 6 DIP gaps on either side.
	if gap := here.X - (home.X + home.W); gap != 6+20+6 {
		t.Fatalf("link to current item distance %v, want 32", gap)
	}
	green := func(c color.RGBA) bool { return int(c.G) > int(c.R)+60 && int(c.G) > int(c.B)+30 }
	if countWhere(tt.Image(), here, green) == 0 {
		t.Fatal("the current item is not drawn in the primary color")
	}
}

func TestUnderlineTabsBaseline(t *testing.T) {
	value := "one"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.Tabs(c, th, &value, []nui.TabItem{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}}, nui.TabsOptions{Label: "Sections", Underline: true}).Width(300)
		})
	}, 400, 150)
	th := nui.Light()
	one, two := bounds(t, tt, "One"), bounds(t, tt, "Two")
	im := tt.Image()
	// Below the triggers and the list's 4 DIP padding, a baseline runs under
	// every trigger, and the active one's indicator covers it.
	y := -1
	for row := int(one.Y + one.H); row < int(one.Y+one.H)+8; row++ {
		if im.RGBAAt(int(two.X+two.W/2), row) == rgba(th.Border) {
			y = row
			break
		}
	}
	if y < int(one.Y+one.H)+3 {
		t.Fatalf("no baseline 4 DIPs under the inactive tab (row %d)", y)
	}
	if c := im.RGBAAt(int(one.X+one.W/2), y); !near(c, rgba(th.Primary), 8) {
		t.Fatalf("active indicator %v is not on the list's border line", c)
	}
	if c := im.RGBAAt(int(two.X+two.W/2), int(two.Y+two.H)-1); c == rgba(th.Primary) {
		t.Fatal("the indicator is still drawn at the trigger's own edge")
	}
}

func TestAccordionTriggerAndBodyRecipe(t *testing.T) {
	open := []string{"a"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.Accordion(c, th, &open, []nui.AccordionItem{{ID: "a", Label: "Shipping", Content: func(c *ui.Context, th nui.Theme) { ui.Text(c, "Ships in two days") }}, {ID: "b", Label: "Returns"}}, nui.AccordionOptions{}).Width(300)
		})
	}, 400, 300)
	tt.Advance(time.Second)
	trigger, text, next := bounds(t, tt, "Shipping"), bounds(t, tt, "Ships in two days"), bounds(t, tt, "Returns")
	if trigger.H != 48 || trigger.X != 20 {
		t.Fatalf("trigger %v; want py-3.5 (48 DIPs) without side padding", trigger)
	}
	if text.Y != trigger.Y+trigger.H || next.Y-(text.Y+text.H) != 14+1 {
		t.Fatalf("body text %v between %v and %v; want no top padding and 14 below", text, trigger, next)
	}
	tt.Move(trigger.X+trigger.W/2, trigger.Y+trigger.H/2)
	tt.Advance(time.Second)
	if c := tt.Image().RGBAAt(int(trigger.X)+2, int(trigger.Y)+2); c != white {
		t.Fatalf("hovered trigger has a surface: %v", c)
	}
}

func TestSelectMenuFlatPopup(t *testing.T) {
	var values []string
	var s nui.SelectMenuState
	created := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.SelectMenu(c, th, &values, &s, []nui.ListboxItem{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, nui.SelectMenuOptions{Label: "Letters", Clearable: true, OnCreate: func(v string) { created = v }, Popup: nui.PopoverOptions{Label: "Letters popup"}}).Width(260)
		})
	}, 400, 400)
	th := nui.Light()
	// No selection: the placeholder is dimmed, without the value's dark text.
	dark := func(c color.RGBA) bool { return c.R < 120 && c.G < 120 && c.B < 120 }
	if countWhere(tt.Image(), bounds(t, tt, "Choose options"), dark) != 0 {
		t.Fatal("the placeholder is not dimmed")
	}
	click(t, tt, "Letters")
	tt.Advance(time.Second)
	panel, alpha := bounds(t, tt, "Letters popup"), bounds(t, tt, "Alpha")
	// The search field is a 32 DIP None field on a 1 DIP divider; the
	// unbordered list insets its two 32 DIP rows by 4: 33 + 4 + 64 + 4.
	if panel.H != 33+72 || alpha.Y != panel.Y+33+4 || alpha.X != panel.X+4 {
		t.Fatalf("popup %v with first row %v; want a flat 105 DIP popup", panel, alpha)
	}
	im := tt.Image()
	if c := im.RGBAAt(int(panel.X+panel.W/2), int(panel.Y)+32); c != rgba(th.Border) {
		t.Fatalf("no divider under the search field: %v", c)
	}
	// An unmatched query offers creation as a list row the arrow keys reach.
	tt.Type("Gamma")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if created != "Gamma" || len(values) != 0 {
		t.Fatalf("create row: created %q, values %v", created, values)
	}
	values = []string{"a"}
	s.Open = false
	tt.Frame()
	click(t, tt, "Clear selection")
	if len(values) != 0 || s.Open {
		t.Fatalf("trigger clear: values %v, open %v", values, s.Open)
	}
}

func TestInputMenuToggleAndTagChips(t *testing.T) {
	state := nui.SearchState{SelectedValues: []string{"b"}}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.InputMenu(c, th, &state, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.InputMenuOptions{Label: "Letters", Multiple: true}).Width(300)
		})
	}, 400, 400)
	click(t, tt, "Show options")
	if !tt.HasText("Alpha") || !tt.Focused("Letters") {
		t.Fatalf("the chevron did not open the list: %v", tt.Texts())
	}
	click(t, tt, "Show options")
	if _, open := tt.Find("Alpha"); open {
		t.Fatal("the chevron did not close the list")
	}
	// MD chips: a 12/16 line with py-0.5; only the x removes.
	beta := bounds(t, tt, "Beta")
	if beta.H != 16 {
		t.Fatalf("chip text line %v, want 16", beta.H)
	}
	tt.ClickAt(beta.X+2, beta.Y+beta.H/2)
	if len(state.SelectedValues) != 1 {
		t.Fatal("clicking the chip's text removed it")
	}
	click(t, tt, "Remove Beta")
	if len(state.SelectedValues) != 0 {
		t.Fatal("the chip's x did not remove it")
	}
}

func TestInputMenuMultipleTagsPrecedeEditor(t *testing.T) {
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		t.Run(map[ui.Direction]string{ui.LeftToRight: "ltr", ui.RightToLeft: "rtl"}[direction], func(t *testing.T) {
			state := nui.SearchState{SelectedValues: []string{"rust", "ts", "go"}}
			width := float32(440)
			changes := 0

			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				ui.Column(c).Padding(20).Children(func() {
					nui.InputMenu(c, th, &state, []nui.Option[string]{{Value: "rust", Label: "Rust"}, {Value: "ts", Label: "TypeScript"}, {Value: "go", Label: "Go"}}, nui.InputMenuOptions{
						Label: "Languages", Placeholder: "Type something...", Multiple: true,
						OnValuesChange: func([]string) { changes++ },
						Style:          func(root, editor, leading, trailing ui.Element) { root.Label("Language field") },
					}).Width(width)
				})
			}, 500, 300)
			tt.SetDirection(direction)
			arrow := bounds(t, tt, "Show options")
			previous := bounds(t, tt, "Rust")
			for _, label := range []string{"TypeScript", "Go", "Languages", "Show options"} {
				next := bounds(t, tt, label)
				if direction == ui.LeftToRight && previous.X+previous.W > next.X || direction == ui.RightToLeft && next.X+next.W > previous.X {
					t.Fatalf("%s is not after the preceding tag/editor: previous %v, next %v", label, previous, next)
				}
				if next.Y+next.H/2 != arrow.Y+arrow.H/2 {
					t.Fatalf("%s unexpectedly wrapped in a wide field: %v, arrow %v", label, next, arrow)
				}
				previous = next
			}
			click(t, tt, "Languages")
			tt.Type("draft")
			width = 200
			tt.Frame()
			outer := bounds(t, tt, "Language field")
			editor := bounds(t, tt, "Languages")
			arrow = bounds(t, tt, "Show options")
			if editor.Y <= bounds(t, tt, "Rust").Y || editor.W < 20 || editor.H <= 0 {
				t.Fatalf("narrow field did not wrap tags and retain usable editor: %v", editor)
			}
			for _, label := range []string{"Rust", "TypeScript", "Go", "Languages", "Show options"} {
				r := bounds(t, tt, label)
				if r.X < outer.X || r.X+r.W > outer.X+outer.W || r.Y < outer.Y || r.Y+r.H > outer.Y+outer.H {
					t.Fatalf("%s escaped the field: %v outside %v", label, r, outer)
				}
			}
			if direction == ui.LeftToRight && editor.X+editor.W > arrow.X || direction == ui.RightToLeft && arrow.X+arrow.W > editor.X {
				t.Fatalf("arrow is not at the trailing edge: editor %v, arrow %v", editor, arrow)
			}
			if state.Query != "draft" || !tt.Focused("Languages") {
				t.Fatalf("wrapping discarded the draft or focus: %+v", state)
			}
			click(t, tt, "Remove TypeScript")
			if !slices.Equal(state.SelectedValues, []string{"rust", "go"}) || changes != 1 || !tt.Focused("Languages") {
				t.Fatalf("tag removal lost focus or emitted incorrect selection: %+v, changes %d", state, changes)
			}
			tt.Key(ui.Ctrl, ui.KeyA)
			tt.Key(0, ui.KeyBackspace)
			tt.Key(0, ui.KeyBackspace)
			if !slices.Equal(state.SelectedValues, []string{"rust"}) || changes != 2 {
				t.Fatalf("empty editor Backspace did not remove the last tag once: %+v, changes %d", state, changes)
			}
		})
	}
}

func TestInputTagsChipsFollowSize(t *testing.T) {
	// Nuxt's 10/12/14 chip text on 12/16/20 lines. The native font cannot draw
	// 10 DIP text on a line shorter than its own metrics, so SM only grows.
	previous := float32(0)
	for _, size := range []nui.Size{nui.SM, nui.MD, nui.XL} {
		tags := []string{"go"}
		var s nui.TagsState
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{Label: "Tags", Size: size})
		}, 400, 200)
		h := bounds(t, tt, "go").H
		if want := map[nui.Size]float32{nui.MD: 16, nui.XL: 20}[size]; want != 0 && h != want || h <= previous {
			t.Errorf("size %v: chip line %v after %v", size, h, previous)
		}
		previous = h
		click(t, tt, "Remove go")
		if len(tags) != 0 {
			t.Errorf("size %v: the x did not remove the tag", size)
		}
	}
}

func TestButtonLinkAndBlockRecipe(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			nui.Button(c, th, "Docs", nui.ButtonOptions{Variant: nui.LinkVariant})
			nui.Button(c, th, "Continue", nui.ButtonOptions{Block: true, Trailing: lucide.ArrowRight(), AccessibleLabel: "Wide"}).Width(300)
		})
	}, 400, 200)
	r := bounds(t, tt, "Docs")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(time.Second)
	// The hover color changes, but no underline appears below the text.
	text := bounds(t, tt, "Docs")
	under := ui.Rect{X: text.X, Y: text.Y + text.H - 3, W: text.W, H: 3}
	if countWhere(tt.Image(), under, func(c color.RGBA) bool { return c.R < 200 }) != 0 {
		t.Fatal("the link button is underlined on hover")
	}
	wide, label := bounds(t, tt, "Wide"), bounds(t, tt, "Continue")
	// Label centered in the remaining space; the icon pinned to the end padding.
	icon := ui.Rect{X: wide.X + wide.W - 10 - 20, Y: wide.Y + 6, W: 20, H: 20}
	if countWhere(tt.Image(), icon, func(c color.RGBA) bool { return c.R > 200 && c.G > 200 && c.B > 200 && c != rgba(nui.Light().Primary) }) == 0 || label.X+label.W > icon.X-20 {
		t.Fatalf("block trailing icon is not at the end: button %v label %v", wide, label)
	}
}

func TestSkeletonEmptyAndKbdRecipes(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Gap(12).AlignItems(ui.Start).Children(func() {
			nui.Skeleton(c, th, nui.SkeletonOptions{Animate: nui.Set(false)})
			nui.Empty(c, th, nui.EmptyOptions{Title: "No projects", Description: "Create one to start.", Icon: lucide.Folder()}).Width(300)
			nui.Kbd(c, th, []string{"tab", "k", "pageup"}, nui.KbdOptions{Platform: "linux"})
		})
	}, 500, 500)
	th := nui.Light()
	skeleton := bounds(t, tt, "Loading…")
	if c := tt.Image().RGBAAt(int(skeleton.X+skeleton.W/2), int(skeleton.Y+skeleton.H/2)); c != rgba(th.Elevated) {
		t.Fatalf("skeleton is %v, want elevated", c)
	}
	title, description := bounds(t, tt, "No projects"), bounds(t, tt, "Create one to start.")
	if description.Y-(title.Y+title.H) != 8 {
		t.Fatalf("empty title to description gap %v, want 8", description.Y-(title.Y+title.H))
	}
	// The icon sits on an elevated 40 DIP circle above the title.
	if c := tt.Image().RGBAAt(int(title.X+title.W/2)-12, int(title.Y)-16-20); c != rgba(th.Elevated) {
		t.Fatalf("no elevated circle behind the empty icon: %v", c)
	}
	for _, key := range []string{"⇥", "K", "⇞"} {
		if _, found := tt.Find(key); !found {
			t.Errorf("Kbd does not show %q: %v", key, tt.Texts())
		}
	}
}

func TestLayoutChromeRecipes(t *testing.T) {
	editor, leading, activeLink, narrowColumns := false, false, false, false
	draft, footerLabel := "", "Intro"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Children(func() {
			header := nui.HeaderOptions{Title: "Workspace"}
			footer := nui.FooterOptions{Text: "Ready"}
			if editor {
				header.Actions = func(c *ui.Context, t nui.Theme) {
					nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Header draft"}).Width(180)
				}
				footer.Actions = func(c *ui.Context, t nui.Theme) { nui.Button(c, t, "Footer sample action", nui.ButtonOptions{}) }
				footer.UI = nui.Styles{nui.Actions: {PaddingX: nui.Set(float32(10))}}
				footer.Style = func(slot nui.Slot, e ui.Element) {
					if slot == nui.Actions {
						e.Label("Footer actions")
					}
				}
			}
			if leading {
				header.Leading = func(c *ui.Context, t nui.Theme) { ui.Text(c, "Mark") }
			}
			nui.HeaderLayout(c.Key("header"), th, header)
			nui.FooterLayout(c.Key("footer"), th, footer)
			columns := []nui.FooterColumn{{ID: "docs", Title: "Docs", Items: []nui.NavigationItem{{ID: "a", Label: footerLabel, Active: activeLink}, {ID: "b", Label: "Guide"}}}}
			if narrowColumns {
				columns = append(columns, nui.FooterColumn{ID: "more", Title: "More", Items: []nui.NavigationItem{{ID: "extra", Label: "Extra"}}})
			}
			footerHost := ui.Column(c.Key("footer-host")).FillWidth().MinWidth(0)
			var columnStyles nui.Styles
			if narrowColumns {
				footerHost.Width(120)
				columnStyles = nui.Styles{nui.Body: {PaddingX: nui.Set(float32(12))}}
			}
			footerHost.Children(func() {
				nui.FooterColumns(c.Key("columns"), th, columns, nui.FooterColumnsOptions{Decoration: nui.Decoration{UI: columnStyles}})
			})
		})
	}, 600, 400)
	title, ready := bounds(t, tt, "Workspace"), bounds(t, tt, "Ready")
	if title.H != 28 {
		t.Fatalf("header title line %v, want the 20/28 title", title.H)
	}
	// The 64 DIP header with its bottom border, and no border above the footer.
	if c := tt.Image().RGBAAt(300, 63); c != rgba(nui.Light().Border) {
		t.Fatalf("header is not 64 DIPs with a bottom border: %v", c)
	}
	if ready.Y < 64 || countColor(tt.Image(), ui.Rect{X: 0, Y: 64, W: 600, H: ready.Y - 64}, rgba(nui.Light().Border)) != 0 {
		t.Fatal("the footer still has a top border")
	}
	docs, intro, guide := bounds(t, tt, "Docs"), bounds(t, tt, "Intro"), bounds(t, tt, "Guide")
	if intro.Y-(docs.Y+docs.H) != 24 || guide.Y-(intro.Y+intro.H) != 16 || intro.X != docs.X {
		t.Fatalf("footer columns: title %v, links %v %v; want gaps 24 and 16, aligned", docs, intro, guide)
	}
	editor = true
	tt.Frame()
	click(t, tt, "Header draft")
	tt.Type("persistent")
	leading = true
	tt.Frame()
	if !tt.Focused("Header draft") || draft != "persistent" {
		t.Fatal("header leading content changed the action editor identity")
	}
	actions, action := bounds(t, tt, "Footer actions"), bounds(t, tt, "Footer sample action")
	if action.X-actions.X != 10 || actions.W-action.W != 20 {
		t.Fatalf("footer Actions padding did not reach its container: %v / %v", actions, action)
	}
	editor, leading, activeLink, narrowColumns = false, false, true, true
	footerLabel = "较长的组件使用说明与完整导航链接"
	tt.SetSize(120, 400)
	link := bounds(t, tt, footerLabel)
	if link.X < 0 || link.X+link.W > 120.1 {
		t.Fatalf("footer column overflowed a narrow view: %v", link)
	}
	if bounds(t, tt, "More").Y <= bounds(t, tt, "Docs").Y {
		t.Fatal("footer columns ignored their minimum width when wrapping")
	}
	if countColor(tt.Image(), link, rgba(nui.Light().Primary)) == 0 {
		t.Fatal("active footer link did not use the primary color")
	}
}

func TestBannerCentersOneLine(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Children(func() {
			nui.Banner(c, th, &open, nui.BannerOptions{Title: "Maintenance tonight", CloseLabel: "Close banner", Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Root {
					e.Label("Banner region")
				}
			}}).FillWidth()
		})
	}, 1200, 200)
	banner, title := bounds(t, tt, "Banner region"), bounds(t, tt, "Maintenance tonight")
	if banner.H != 48 || title.H != 20 {
		t.Fatalf("banner %v with title %v; want a 48 DIP row with one 14/20 line", banner, title)
	}
	if center := title.X + title.W/2; center < 590 || center > 610 {
		t.Fatalf("banner title %v is not centered above the 1024 DIP breakpoint", title)
	}
}

func TestFileUploadDashedZoneAndSizes(t *testing.T) {
	var s nui.FileUploadState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.FileUpload(c, th, &s, nui.FileUploadOptions{Label: "Attachments", Description: "Drop files here"}).Width(300)
		})
	}, 400, 400)
	zone := bounds(t, tt, "Attachments")
	th := nui.Light()
	// A dashed edge alternates the border color with the background.
	border, gaps := 0, 0
	for x := int(zone.X) + 20; x < int(zone.X+zone.W)-20; x++ {
		switch tt.Image().RGBAAt(x, int(zone.Y)) {
		case rgba(th.Border):
			border++
		case white:
			gaps++
		}
	}
	if border == 0 || gaps == 0 {
		t.Fatalf("drop zone edge has %d border and %d gap pixels; want dashes", border, gaps)
	}
}
