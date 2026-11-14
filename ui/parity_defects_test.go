package nui_test

import (
	"image"
	"image/color"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// Regressions for parameters that were accepted but ignored, broken rendering
// and right-to-left defects found by the Nuxt UI v4.12 parity review.

func rgba(c ui.Color) color.RGBA { return color.RGBA{c.R, c.G, c.B, 255} }

var white = color.RGBA{255, 255, 255, 255}

// countColor counts the pixels of r that match want exactly.
func countColor(im *image.RGBA, r ui.Rect, want color.RGBA) int {
	n := 0
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if im.RGBAAt(x, y) == want {
				n++
			}
		}
	}
	return n
}

// topCorners samples the top row one pixel inside each side: a rounded corner
// shows the background there, while FieldGroup's 1 DIP overlap stays outside.
func topCorners(im *image.RGBA, r ui.Rect) (left, right color.RGBA) {
	y := int(r.Y)
	return im.RGBAAt(int(r.X)+1, y), im.RGBAAt(int(r.X+r.W)-2, y)
}

// chevronPointsRight reports whether the chevron drawn on a white background
// in r points right: its open side, at the left, spreads wider than its tip.
func chevronPointsRight(t *testing.T, im *image.RGBA, r ui.Rect) bool {
	t.Helper()
	ink := func(x, y int) bool { c := im.RGBAAt(x, y); return c.R < 235 || c.G < 235 || c.B < 235 }
	x0, x1, y0, y1 := int(r.X), int(r.X+r.W), int(r.Y), int(r.Y+r.H)
	minX, maxX := math.MaxInt, -1
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if ink(x, y) {
				minX, maxX = min(minX, x), max(maxX, x)
			}
		}
	}
	if maxX < 0 {
		t.Fatal("no chevron drawn in", r)
	}
	spread := func(x int) int {
		lo, hi := math.MaxInt, -1
		for y := y0; y < y1; y++ {
			if ink(x, y) {
				lo, hi = min(lo, y), max(hi, y)
			}
		}
		return hi - lo
	}
	return spread(minX) > spread(maxX)
}

func TestCheckboxGroupForwardsItemPartStyles(t *testing.T) {
	marker := ui.RGB(255, 0, 255)
	var selected []string
	calls := map[nui.Slot]int{}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		clear(calls)
		ui.Column(c).Padding(16).Children(func() {
			nui.CheckboxGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}, nui.CheckboxGroupOptions{Label: "Letters",
				UI:    nui.Styles{nui.Label: {Foreground: nui.Set(marker)}},
				Style: func(slot nui.Slot, e ui.Element) { calls[slot]++ }})
		})
	}, 300, 160)
	if calls[nui.Root] != 1 || calls[nui.Item] != 2 || calls[nui.Label] != 2 || calls[nui.Indicator] != 2 {
		t.Fatalf("group Style calls = %v; want one Root and per-item Item, Label and Indicator", calls)
	}
	if countColor(tt.Image(), bounds(t, tt, "Letters"), rgba(marker)) == 0 {
		t.Fatal("group UI[Label] did not color the item labels")
	}
	click(t, tt, "Beta")
	if len(selected) != 1 || selected[0] != "b" {
		t.Fatalf("selection after styling = %v", selected)
	}
}

func TestCheckboxUsesCheckedIcon(t *testing.T) {
	render := func(icon *ui.SVG) *image.RGBA {
		on := true
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(16).Children(func() { nui.Checkbox(c, th, &on, nui.CheckboxOptions{Label: "Agree", CheckedIcon: icon}) })
		}, 160, 60)
		return tt.Image()
	}
	plain, custom := render(nil), render(lucide.X())
	differs := false
	for y := 16; y < 36 && !differs; y++ {
		for x := 16; x < 36; x++ {
			if plain.RGBAAt(x, y) != custom.RGBAAt(x, y) {
				differs = true
				break
			}
		}
	}
	if !differs {
		t.Fatal("CheckedIcon did not replace the default check mark")
	}
}

func TestSwitchTrackRingOnlyForHighlightOrError(t *testing.T) {
	on := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Gap(20).Children(func() {
			nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Plain"})
			nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Lit", Highlight: true})
			nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Broken", Invalid: true})
		})
	}, 300, 200)
	th := nui.Light()
	for label, want := range map[string]ui.Color{"Plain": th.Accented, "Lit": th.Primary, "Broken": th.Error} {
		// The middle of the track's top edge, away from its rounded ends and the thumb.
		r := bounds(t, tt, label)
		if got := tt.Image().RGBAAt(int(r.X)+18, int(r.Y)); got != rgba(want) {
			t.Errorf("%s switch track edge = %v, want %v", label, got, rgba(want))
		}
	}
}

func TestFieldGroupKeepsFormFieldSize(t *testing.T) {
	var a, b string
	var groupSize nui.Size
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Children(func() {
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Range", Size: nui.LG}, func(th nui.Theme) {
				nui.FieldGroup(c, th, nui.FieldGroupOptions{Size: groupSize}, []nui.FieldGroupItem{
					{ID: "a", Build: func(c *ui.Context, th nui.Theme) ui.Element {
						return nui.Input(c, th, &a, nui.InputOptions{Label: "From"})
					}},
					{ID: "b", Build: func(c *ui.Context, th nui.Theme) ui.Element {
						return nui.Input(c, th, &b, nui.InputOptions{Label: "To"})
					}},
				})
			})
		})
	}, 400, 200)
	// The labeled editor sits below its root's vertical padding: 8 for LG, 6 for SM and MD.
	editorY := func() float32 { return bounds(t, tt, "From").Y }
	lg := editorY()
	groupSize = nui.SM
	tt.Frame()
	if sm := editorY(); sm != lg-2 {
		t.Fatalf("explicit group size: SM editor at %v, LG at %v", sm, lg)
	}
	groupSize = nui.SizeDefault
	tt.Frame()
	if got := editorY(); got != lg {
		t.Fatalf("unsized FieldGroup replaced FormField LG: editor at %v, want %v", got, lg)
	}
}

func TestFieldGroupCornersFollowDirection(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.FieldGroup(c, th, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
				{ID: "a", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Button(c, th, "First", nui.ButtonOptions{}).Width(80)
				}},
				{ID: "b", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Button(c, th, "Second", nui.ButtonOptions{}).Width(80)
				}},
			})
		})
	}, 400, 100)
	green := rgba(nui.Light().Primary)
	rounded := func(label string) (left, right bool) {
		l, r := topCorners(tt.Image(), bounds(t, tt, label))
		return l != green, r != green
	}
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(dir)
		tt.Frame()
		firstLeft, firstRight := rounded("First")
		lastLeft, lastRight := rounded("Second")
		// Left to right, the first item is leftmost; right to left, rightmost.
		ltr := dir == ui.LeftToRight
		if firstLeft != ltr || firstRight == ltr || lastLeft == ltr || lastRight != ltr {
			t.Fatalf("%v: rounded corners first=(%v,%v) last=(%v,%v); only the group's outer corners should be rounded", dir, firstLeft, firstRight, lastLeft, lastRight)
		}
	}
}

func TestBadgeAndSelectMenuJoinFieldGroup(t *testing.T) {
	var value []string
	var s nui.SelectMenuState
	var text string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Gap(20).Children(func() {
			nui.FieldGroup(c, th, nui.FieldGroupOptions{Size: nui.LG}, []nui.FieldGroupItem{
				{ID: "badge", Build: func(c *ui.Context, th nui.Theme) ui.Element { return nui.Badge(c, th, "https://", nui.BadgeOptions{}) }},
				{ID: "text", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Input(c, th, &text, nui.InputOptions{Label: "Host"}).Width(120)
				}},
			})
			nui.FieldGroup(c, th, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
				{ID: "text", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.Input(c, th, &text, nui.InputOptions{Label: "Name"}).Width(120)
				}},
				{ID: "menu", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					return nui.SelectMenu(c, th, &value, &s, []nui.ListboxItem{{ID: "x", Label: "Xray"}}, nui.SelectMenuOptions{Label: "Domain"}).Width(160)
				}},
			})
		})
	}, 400, 200)
	// Badge has no accessible label; Find returns its text, one LG line tall.
	label := bounds(t, tt, "https://")
	if label.H != 20 {
		t.Fatalf("badge in an LG group has a %v DIP label line, want LG's 20", label.H)
	}
	green := rgba(nui.Light().Primary)
	im := tt.Image()
	// Group items stretch: the badge's label line is centered in the 36 DIP input height.
	host := bounds(t, tt, "Host")
	badge := ui.Rect{X: label.X - 8, Y: host.Y - 8, W: label.W + 16, H: 36}
	if label.Y != host.Y {
		t.Fatalf("badge label at %v is not aligned with its neighbor's text at %v", label.Y, host.Y)
	}
	if l, r := topCorners(im, badge); l == green || r != green {
		t.Fatalf("first badge corners: start %v end %v; want only the start rounded", l, r)
	}
	if l, r := topCorners(im, bounds(t, tt, "Domain")); l == white || r != white {
		t.Fatalf("last SelectMenu trigger corners: start %v end %v; want only the end rounded", l, r)
	}
}

func TestPopoverContentLeavesFieldGroup(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.FieldGroup(c, th, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
				{ID: "a", Build: func(c *ui.Context, th nui.Theme) ui.Element {
					anchor := nui.Button(c, th, "Open", nui.ButtonOptions{})
					nui.Popover(c, th, anchor, &open, nui.PopoverOptions{}, func(inner nui.Theme) { nui.Button(c, inner, "Inside", nui.ButtonOptions{}).Width(80) })
					return anchor
				}},
				{ID: "b", Build: func(c *ui.Context, th nui.Theme) ui.Element { return nui.Button(c, th, "Next", nui.ButtonOptions{}) }},
			})
		})
	}, 400, 200)
	green := rgba(nui.Light().Primary)
	if l, r := topCorners(tt.Image(), bounds(t, tt, "Inside")); l == green || r == green {
		t.Fatal("a control in a group item's popover was given group corners")
	}
}

func TestRangeSliderBindsFormField(t *testing.T) {
	low, high := 20.0, 80.0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Gap(12).Children(func() {
			nui.Button(c, th, "Before", nui.ButtonOptions{})
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Price", Error: "Too wide"}, func(th nui.Theme) {
				nui.RangeSlider(c, th, &low, &high, nui.RangeSliderOptions{}).Width(240)
			})
		})
	}, 400, 200)
	click(t, tt, "Before")
	label := bounds(t, tt, "Price")
	tt.ClickAt(label.X+8, label.Y+6)
	if !tt.Focused("minimum") {
		t.Fatalf("FormField label did not focus the first thumb; texts %v", tt.Texts())
	}
	tt.Key(0, ui.KeyRight)
	if low != 21 || high != 80 {
		t.Fatalf("keyboard after label focus moved the range to %v-%v", low, high)
	}
}

func TestPinInputShowsFormFieldError(t *testing.T) {
	pin := ""
	var s nui.PinState
	blue := ui.RGB(0, 0, 255)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).Children(func() {
			// Recolor the error text so that only error-colored slots can match.
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Code", Error: "Wrong code", UI: nui.Styles{nui.ErrorSlot: {Foreground: nui.Set(blue)}}}, func(th nui.Theme) {
				nui.PinInput(c, th, &pin, &s, nui.PinInputOptions{Length: 4})
			})
		})
	}, 300, 160)
	if countColor(tt.Image(), ui.Rect{W: 300, H: 160}, rgba(nui.Light().Error)) == 0 {
		t.Fatal("PIN slots in an invalid FormField kept their normal border")
	}
}

func TestBreadcrumbAndFooterColumnsShowItemIcons(t *testing.T) {
	icons := false
	cancel := true
	selected, navigated := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		var icon *ui.SVG
		if icons {
			icon = lucide.House()
		}
		ui.Column(c).Padding(16).Gap(16).AlignItems(ui.Start).Children(func() {
			nui.Breadcrumb(c, th, []nui.BreadcrumbItem{{ID: "home", Label: "Home", OnSelect: func() {}}, {ID: "here", Label: "Here", Icon: icon}}, nui.BreadcrumbOptions{})
			nui.FooterColumns(c, th, []nui.FooterColumn{{ID: "docs", Title: "Docs", Items: []nui.NavigationItem{{ID: "start", Label: "Start", Icon: icon, OnSelect: func(e *nui.NavigationEvent) {
				selected++
				if cancel {
					e.PreventDefault()
				}
			}}}}}, nui.FooterColumnsOptions{OnNavigate: func(string) { navigated++ }})
		})
	}, 500, 200)
	// Footer items fill their column, so compare the item's pixels instead of its width.
	region := func(r ui.Rect) []color.RGBA {
		var out []color.RGBA
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			for x := int(r.X); x < int(r.X+r.W); x++ {
				out = append(out, tt.Image().RGBAAt(x, y))
			}
		}
		return out
	}
	here, start := bounds(t, tt, "Here").X, region(bounds(t, tt, "Start"))
	icons = true
	tt.Frame()
	if got := bounds(t, tt, "Here").X; got <= here {
		t.Fatalf("current breadcrumb icon missing: label at %v, was %v", got, here)
	}
	if slices.Equal(region(bounds(t, tt, "Start")), start) {
		t.Fatal("footer item icon missing")
	}
	click(t, tt, "Start")
	if selected != 1 || navigated != 0 {
		t.Fatalf("footer link cancellation: selects=%d navigations=%d", selected, navigated)
	}
	cancel = false
	tt.Key(0, ui.KeyEnter)
	if selected != 2 || navigated != 1 {
		t.Fatalf("footer link activation: selects=%d navigations=%d", selected, navigated)
	}
}

func TestEmptyCardAndBadgeParts(t *testing.T) {
	footer := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).AlignItems(ui.Start).Children(func() {
			o := nui.CardOptions{Title: "Only"}
			if footer {
				o.Footer = func() { ui.Text(c, "Foot") }
			}
			nui.Card(c, th, o, nil).Width(240)
			ui.Text(c, "After card")
			ui.Row(c).Children(func() {
				nui.Badge(c, th, "", nui.BadgeOptions{Leading: lucide.Check()})
				ui.Text(c, "After badge")
			})
		})
	}, 400, 400)
	// Header padding 16 + the title line + 16; no body region or trailing divider.
	header := bounds(t, tt, "Only").H + 32
	if got := bounds(t, tt, "After card").Y - 20; got != header {
		t.Fatalf("card without body is %v DIP tall, want its %v DIP header", got, header)
	}
	// Automatic square padding: 4 + 16 DIP icon + 4, with no label gap.
	if got := bounds(t, tt, "After badge").X - 20; got != 24 {
		t.Fatalf("icon-only badge is %v DIP wide, want 24", got)
	}
	footer = true
	tt.Frame()
	// The header, a single divider, and the footer's padded line.
	want := header + 1 + bounds(t, tt, "Foot").H + 32
	if got := bounds(t, tt, "After card").Y - 20; got != want {
		t.Fatalf("card with header and footer but no body is %v DIP tall, want %v", got, want)
	}
}

func TestCommandPaletteShowsDescriptions(t *testing.T) {
	var s nui.CommandPaletteState
	s.SetIndex(nui.NewCommandIndex([]nui.CommandAction{{ID: "new", Label: "New file", Description: "Create an empty document"}}))
	s.Open = true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.CommandPalette(c, th, &s, nui.CommandPaletteOptions{})
	}, 700, 500)
	awaitCommandResults(t, tt, func() bool { return tt.HasText("New file") })
	if !tt.HasText("Create an empty document") {
		t.Fatalf("command description missing: %v", tt.Texts())
	}
	// Find returns the labeled row, which must grow to hold both lines.
	row, description := bounds(t, tt, "New file"), bounds(t, tt, "Create an empty document")
	if description.Y < row.Y+row.H/2 || description.Y+description.H > row.Y+row.H {
		t.Fatalf("description %v is not the lower line of row %v", description, row)
	}
}

func TestDirectionalIconsMirrorRightToLeft(t *testing.T) {
	collapsed := false
	var value nui.DateSelection
	var cal nui.CalendarState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Row(c).Children(func() {
			nui.Sidebar(c, th, &collapsed, nui.SidebarOptions{Title: "Files"}, nil)
			ui.Column(c).Padding(12).Gap(12).Children(func() {
				nui.Breadcrumb(c, th, []nui.BreadcrumbItem{{ID: "a", Label: "Root", OnSelect: func() {}}, {ID: "b", Label: "Leaf"}}, nui.BreadcrumbOptions{})
				nui.Calendar(c, th, &value, &cal, nui.CalendarOptions{Label: "Due"})
			})
		})
	}, 700, 500)
	separator := func() ui.Rect {
		root, leaf := bounds(t, tt, "Root"), bounds(t, tt, "Leaf")
		start, end := root.X+root.W, leaf.X
		if leaf.X < root.X {
			start, end = leaf.X+leaf.W, root.X
		}
		return ui.Rect{X: start, Y: leaf.Y - 2, W: end - start, H: leaf.H + 4}
	}
	pixels := func(r ui.Rect) []color.RGBA {
		var out []color.RGBA
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			for x := int(r.X); x < int(r.X+r.W); x++ {
				out = append(out, tt.Image().RGBAAt(x, y))
			}
		}
		return out
	}
	sidebar := map[ui.Direction][]color.RGBA{}
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(dir)
		tt.Frame()
		im := tt.Image()
		forward := dir == ui.LeftToRight
		if chevronPointsRight(t, im, separator()) != forward {
			t.Errorf("%v: breadcrumb separator should point toward the line's end", dir)
		}
		if chevronPointsRight(t, im, bounds(t, tt, "Previous month")) == forward || chevronPointsRight(t, im, bounds(t, tt, "Next month")) != forward {
			t.Errorf("%v: calendar paging chevrons should point toward the earlier and later months", dir)
		}
		sidebar[dir] = pixels(bounds(t, tt, "Toggle sidebar"))
	}
	if slices.Equal(sidebar[ui.LeftToRight], sidebar[ui.RightToLeft]) {
		t.Error("sidebar toggle icon does not mirror right to left")
	}
}

func TestNoticeCloseButtonsKeepContrastOnHover(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).Gap(16).Children(func() {
			nui.Banner(c, th, &open, nui.BannerOptions{Title: "Release notes", CloseLabel: "Close banner"})
			nui.Alert(c, th, nui.AlertOptions{Title: "Saved", OnClose: func() {}})
		})
	}, 500, 260)
	th := nui.Light()
	elevated := rgba(th.Elevated)
	for _, label := range []string{"Close banner", "Dismiss Saved"} {
		r := bounds(t, tt, label)
		tt.Move(r.X+r.W/2, r.Y+r.H/2)
		tt.Advance(time.Second) // Let the hover color transition finish.
		// The button's corner holds no icon stroke; it shows only the hover surface.
		got := tt.Image().RGBAAt(int(r.X)+2, int(r.Y)+2)
		if got == elevated || got.G < got.R+100 {
			t.Errorf("%s hover surface %v hides the inverted icon on the primary color", label, got)
		}
	}
}

func TestTextInputPlaceholdersUseDimmed(t *testing.T) {
	for _, dark := range []bool{false, true} {
		text := ""
		var number nui.Optional[float64]
		var numberState nui.NumberState
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: map[bool]nui.Mode{false: nui.LightMode, true: nui.DarkMode}[dark]})
			ui.Column(c).Padding(16).Gap(16).Children(func() {
				nui.Input(c, th, &text, nui.InputOptions{Label: "Plain", Placeholder: "Search everything"}).Width(260)
				nui.InputNumber(c, th, &number, &numberState, nui.InputNumberOptions{Label: "Count", Placeholder: "Enter a number"}).Width(260)
			})
		}, 320, 160)
		th := nui.Light()
		if dark {
			th = nui.Dark()
		}
		for _, label := range []string{"Plain", "Count"} {
			r := bounds(t, tt, label)
			if countColor(tt.Image(), r, rgba(th.Dimmed)) == 0 || countColor(tt.Image(), r, rgba(th.Muted)) != 0 {
				t.Errorf("dark=%v %s placeholder is not drawn in Dimmed", dark, label)
			}
		}
	}
}
