package nui_test

import (
	"fmt"
	"image"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// redPixels counts the pixels of r tinted red, as a red panel fading out
// over a white window leaves them.
func redPixels(im *image.RGBA, r ui.Rect) int {
	n := 0
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if redSurfacePixel(im, x, y) {
				n++
			}
		}
	}
	return n
}

// A solid panel tints an area, whereas Windows subpixel text rendering can
// tint individual glyph-edge pixels even in a closed, never-opened control.
// Require the surrounding pixels too, so text fringes cannot masquerade as
// a surviving exit. This trims one pixel from each measured panel edge.
func redSurfacePixel(im *image.RGBA, x, y int) bool {
	for _, at := range [][2]int{{x, y}, {x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
		p := im.RGBAAt(at[0], at[1])
		if int(p.R) <= int(p.G)+12 || int(p.R) <= int(p.B)+12 {
			return false
		}
	}
	return true
}

// The overlay exits rely on MyGo copying a PopoverBase panel, built
// directly in the overlay layer, as it leaves: the copy paints for its
// ExitDuration and takes no input (design D6).
func TestPopoverBaseExitCopySpike(t *testing.T) {
	open, behind := true, 0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Padding(20).Gap(8).Children(func() {
			anchor := ui.PrimaryButton(c, "Anchor")
			ui.PrimaryButton(c, "Behind").OnClick(func() { behind++ })
			ui.PopoverBase(c, anchor, &open, func(panel ui.Element) {
				panel.Size(200, 120).Background(ui.RGB(220, 0, 0)).Label("Panel").
					Transition(ui.ElementTransition{Exit: &ui.Motion{Scale: .95}, ExitDuration: 100 * time.Millisecond})
			})
		})
	}, 400, 300)
	panel := bounds(t, tt, "Panel")
	under := bounds(t, tt, "Behind")
	if cx, cy := under.X+under.W/2, under.Y+under.H/2; cx < panel.X || cx > panel.X+panel.W || cy < panel.Y || cy > panel.Y+panel.H {
		t.Fatalf("the button %v is not under the panel %v", under, panel)
	}
	if redPixels(tt.Image(), panel) == 0 {
		t.Fatal("the open panel painted nothing")
	}
	open = false
	tt.Frame()
	tt.Advance(30 * time.Millisecond)
	tt.ClickAt(under.X+under.W/2, under.Y+under.H/2)
	if behind != 1 {
		t.Fatalf("a click 30 ms into the exit did not reach the button behind: %d", behind)
	}
	tt.Advance(20 * time.Millisecond)
	if n := redPixels(tt.Image(), panel); n == 0 {
		t.Fatal("the exit copy was not painted at 50 ms")
	}
	tt.Advance(50 * time.Millisecond)
	if n := redPixels(tt.Image(), panel); n != 0 {
		t.Fatalf("the exit copy was still painted at 100 ms: %d pixels", n)
	}
}

// redExtent returns the box of the red-tinted pixels of r and how red the
// reddest is (0 to 255): a red panel fading shows less, and one scaling a
// smaller box.
func redExtent(im *image.RGBA, r ui.Rect) (box ui.Rect, strength int) {
	x0, y0, x1, y1 := 1<<30, 1<<30, -1, -1
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			p := im.RGBAAt(x, y)
			if tint := int(p.R) - int(p.G); redSurfacePixel(im, x, y) {
				x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
				strength = max(strength, tint)
			}
		}
	}
	if x1 < 0 {
		return ui.Rect{}, 0
	}
	return ui.Rect{X: float32(x0), Y: float32(y0), W: float32(x1 - x0 + 1), H: float32(y1 - y0 + 1)}, strength
}

var redPanel = nui.Styles{nui.Panel: {Background: nui.Set(ui.RGB(220, 0, 0))}}

// popoverView is a red Popover on an anchor, over a button behind it.
func popoverView(open *bool, behind *int, inside *int) func(c *ui.Context, th nui.Theme) {
	return func(c *ui.Context, th nui.Theme) {
		a := nui.Button(c.Key("anchor"), th, "Details", nui.ButtonOptions{}).OnClick(func() { *open = !*open })
		nui.Button(c.Key("behind"), th, "Behind", nui.ButtonOptions{}).OnClick(func() { *behind++ })
		nui.Popover(c.Key("popover"), th, a, open, nui.PopoverOptions{Width: nui.Set(float32(220)), UI: redPanel, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Panel {
				e.Label("Panel")
			}
		}}, func(th nui.Theme) {
			ui.Box(c).Height(60)
			nui.Button(c.Key("inside"), th, "Inside", nui.ButtonOptions{}).OnClick(func() { *inside++ })
		})
	}
}

func TestPopoverExitAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			open, behind, inside := true, 0, 0
			m := newMotionTester(l, 500, 400, popoverView(&open, &behind, &inside))
			m.Advance(300 * time.Millisecond)
			region := ui.Rect{W: 500, H: 400}
			rest, full := redExtent(m.Image(), region)
			m.Key(0, ui.KeyEscape)
			if open {
				t.Fatal("Escape did not close the Popover")
			}
			if l.off {
				if box, _ := redExtent(m.Image(), region); box.W != 0 {
					t.Fatalf("at Off the panel stayed: %v", box)
				}
				m.idle(t, time.Second, "a closed Popover at Off")
				return
			}
			m.Advance(50 * time.Millisecond)
			box, strength := redExtent(m.Image(), region)
			if box.W == 0 || strength >= full {
				t.Fatalf("at 50 ms the panel was not painted, partly transparent: %v strength %d of %d", box, strength, full)
			}
			if smaller := box.W < rest.W-2; smaller != (l == motionFull) {
				t.Fatalf("at 50 ms the panel was %v, at rest %v", box, rest)
			}
			m.Advance(50 * time.Millisecond)
			if box, _ := redExtent(m.Image(), region); box.W != 0 {
				t.Fatalf("at 100 ms the panel was still painted: %v", box)
			}
		})
	}
}

func TestClickThroughALeavingPopover(t *testing.T) {
	open, behind, inside := true, 0, 0
	m := newMotionTester(motionFull, 500, 400, popoverView(&open, &behind, &inside))
	m.Advance(300 * time.Millisecond)
	panel, under := bounds(t, m.Tester, "Panel"), bounds(t, m.Tester, "Behind")
	// A point of the button the panel covers.
	cx, cy := panel.X+20, under.Y+under.H/2
	if cx < under.X || cx > under.X+under.W || cy < panel.Y || cy > panel.Y+panel.H {
		t.Fatalf("the button %v is not under the panel %v", under, panel)
	}
	m.ClickAt(480, 380) // outside, to close it
	if open {
		t.Fatal("pressing outside did not close the Popover")
	}
	m.Advance(30 * time.Millisecond)
	m.ClickAt(cx, cy)
	if behind != 1 || inside != 0 {
		t.Fatalf("the click 30 ms into the exit went to behind=%d inside=%d", behind, inside)
	}
}

func TestReopenDuringAPopoverExit(t *testing.T) {
	open, behind, inside := true, 0, 0
	m := newMotionTester(motionFull, 500, 400, popoverView(&open, &behind, &inside))
	m.Advance(300 * time.Millisecond)
	region := ui.Rect{W: 500, H: 400}
	rest, full := redExtent(m.Image(), region)
	open = false
	m.Frame()
	m.Advance(40 * time.Millisecond)
	_, leaving := redExtent(m.Image(), region)
	open = true
	m.Frame()
	for step := 0; step <= 100; step += 20 {
		box, strength := redExtent(m.Image(), region)
		// It goes on from where the exit left it, not from transparent.
		if strength < leaving {
			t.Fatalf("%d ms after reopening the panel faded to %d, below the %d it left at", step, strength, leaving)
		}
		// One panel shows, where it rests: a second, entering at 95% over
		// the leaving copy, would show an inner edge of its own.
		if box.W > rest.W || box.H > rest.H {
			t.Fatalf("%d ms after reopening the panels covered %v, more than %v", step, box, rest)
		}
		if step == 100 && (box != rest || strength != full) {
			t.Fatalf("100 ms after reopening the panel was %v at %d, want %v at %d", box, strength, rest, full)
		}
		m.Advance(20 * time.Millisecond)
	}
	click(t, m.Tester, "Inside")
	if inside != 1 {
		t.Fatal("the reopened panel took no input")
	}
}

func TestSubmenuLeavesWithItsParent(t *testing.T) {
	var state nui.MenuState
	chosen := 0
	state.SetItems([]nui.MenuItem{{ID: "edit", Label: "Edit"}, {ID: "more", Label: "More", Children: []nui.MenuItem{{ID: "archive", Label: "Archive", OnSelect: func() {
		chosen++
	}}}}})
	m := newMotionTester(motionFull, 600, 400, func(c *ui.Context, th nui.Theme) {
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c, th, a, &state, nui.MenuOptions{UI: redPanel})
	})
	click(t, m.Tester, "Actions")
	m.Advance(300 * time.Millisecond)
	click(t, m.Tester, "More")
	m.Advance(300 * time.Millisecond)
	region := ui.Rect{W: 600, H: 400}
	before, _ := redExtent(m.Image(), region)
	click(t, m.Tester, "Archive")
	if chosen != 1 || state.Open {
		t.Fatalf("choosing the submenu item: chosen=%d open=%v", chosen, state.Open)
	}
	m.Advance(40 * time.Millisecond)
	if box, _ := redExtent(m.Image(), region); box.W == 0 || box.X > before.X+30 || box.X+box.W < before.X+before.W-30 {
		t.Fatalf("the menus did not leave together: %v, open %v", box, before)
	}
	m.Advance(70 * time.Millisecond)
	if box, _ := redExtent(m.Image(), region); box.W != 0 {
		t.Fatalf("a menu was still painted after 100 ms: %v", box)
	}
}

func TestMenuChoiceDoesNotWaitForTheExit(t *testing.T) {
	var state nui.MenuState
	selects := 0
	state.SetItems([]nui.MenuItem{{ID: "edit", Label: "Edit", OnSelect: func() {
		selects++
	}}})
	m := newMotionTester(motionFull, 500, 400, func(c *ui.Context, th nui.Theme) {
		a := nui.Button(c, th, "Actions", nui.ButtonOptions{})
		nui.DropdownMenu(c, th, a, &state, nui.MenuOptions{UI: redPanel})
	})
	click(t, m.Tester, "Actions")
	m.Advance(300 * time.Millisecond)
	m.Key(0, ui.KeyEnter)
	if selects != 1 || !m.Focused("Actions") {
		t.Fatalf("Enter: OnSelect ran %d times, the button focused=%v", selects, m.Focused("Actions"))
	}
	if box, _ := redExtent(m.Image(), ui.Rect{W: 500, H: 400}); box.W == 0 {
		t.Fatal("the menu was not leaving as focus returned")
	}
	m.Advance(time.Second)
	if selects != 1 {
		t.Fatalf("OnSelect ran %d times", selects)
	}
}

func TestTooltipLeavesAtOnceAtOff(t *testing.T) {
	m := newMotionTester(motionOff, 500, 300, func(c *ui.Context, th nui.Theme) {
		a := nui.Button(c, th, "Save", nui.ButtonOptions{})
		nui.Tooltip(c, th, a, "Save the draft", nui.TooltipOptions{DelayDuration: nui.Set(time.Duration(0))})
	})
	r := bounds(t, m.Tester, "Save")
	m.Move(r.X+r.W/2, r.Y+r.H/2)
	m.Advance(100 * time.Millisecond)
	if !m.HasText("Save the draft") {
		t.Fatal("the tooltip did not show")
	}
	m.Move(495, 295)
	if m.HasText("Save the draft") {
		t.Fatal("the tooltip was still shown in the frame after the pointer left")
	}
}

// modalView is a red Modal over a button behind it.
func modalView(open *bool, closes, behind *int, transition nui.Optional[bool]) func(c *ui.Context, th nui.Theme) {
	return func(c *ui.Context, th nui.Theme) {
		nui.Button(c.Key("behind"), th, "Behind", nui.ButtonOptions{}).OnClick(func() { *behind++ })
		nui.ModalWithOptions(c.Key("modal"), th, open, nui.ModalOptions{Title: "Edit", Transition: transition, UI: redPanel, OnClose: func() { *closes++ }}, func(nui.Theme) {
			ui.Text(c, "Modal body")
			ui.Box(c).Height(100)
		})
	}
}

func TestModalExitAtEachLevel(t *testing.T) {
	type run struct {
		level      motionLevel
		transition nui.Optional[bool]
	}
	for _, r := range []run{{motionFull, nui.Optional[bool]{}}, {motionReducedDesktop, nui.Optional[bool]{}}, {motionReducedTheme, nui.Optional[bool]{}}, {motionOff, nui.Optional[bool]{}}, {motionFull, nui.Set(false)}} {
		t.Run(fmt.Sprintf("%s/transition=%v", r.level.name, r.transition), func(t *testing.T) {
			open, closes, behind := true, 0, 0
			m := newMotionTester(r.level, 700, 500, modalView(&open, &closes, &behind, r.transition))
			m.UseScreenReader()
			m.Advance(400 * time.Millisecond)
			region := ui.Rect{W: 700, H: 500}
			rest, full := redExtent(m.Image(), region)
			under := bounds(t, m.Tester, "Behind")
			m.Key(0, ui.KeyEscape)
			if open || closes != 1 {
				t.Fatalf("Escape: open=%v, OnClose ran %d times", open, closes)
			}
			for _, info := range m.AccessAll() {
				if info.Role == ui.RoleDialog || info.Name == "Modal body" {
					t.Fatalf("the leaving Modal is exposed to assistive technology: %+v", info)
				}
			}
			box, _ := redExtent(m.Image(), region)
			still := r.level.off || r.transition.Set
			if still {
				if box.W != 0 {
					t.Fatalf("the Modal left a copy: %v", box)
				}
			} else {
				if box.W == 0 || box.W != rest.W {
					t.Fatalf("the exit did not keep the panel's size: %v, at rest %v", box, rest)
				}
				// The window behind takes input from the first frame of the exit.
				m.ClickAt(under.X+under.W/2, under.Y+under.H/2)
				if behind != 1 {
					t.Fatal("the window behind took no click during the exit")
				}
				m.Advance(100 * time.Millisecond)
				if box, strength := redExtent(m.Image(), region); box.W == 0 || strength >= full {
					t.Fatalf("at 100 ms the panel was not fading: %v at %d of %d", box, strength, full)
				}
				m.Advance(120 * time.Millisecond)
			}
			if box, _ := redExtent(m.Image(), region); box.W != 0 {
				t.Fatalf("the Modal was still painted after its exit: %v", box)
			}
			if closes != 1 {
				t.Fatalf("OnClose ran %d times", closes)
			}
		})
	}
}

func TestSlideoverAndDrawerFadeInPlaceAtReduced(t *testing.T) {
	for _, kind := range []string{"slideover", "drawer"} {
		t.Run(kind, func(t *testing.T) {
			open := false
			drawer := nui.DrawerState{}
			m := newMotionTester(motionReducedTheme, 700, 500, func(c *ui.Context, th nui.Theme) {
				nui.Button(c, th, "Open", nui.ButtonOptions{}).OnClick(func() { open, drawer.Open = true, true })
				mo := nui.ModalOptions{Title: "Sheet", UI: redPanel}
				if kind == "drawer" {
					nui.Drawer(c, th, &drawer, nui.DrawerOptions{ModalOptions: mo}, func(nui.Theme) { ui.Box(c).Height(80) })
				} else {
					nui.Slideover(c, th, &open, nui.SlideoverOptions{ModalOptions: mo, Side: nui.SideRight}, func(nui.Theme) { ui.Box(c).Height(80) })
				}
			})
			click(t, m.Tester, "Open")
			region := ui.Rect{W: 700, H: 500}
			var boxes []ui.Rect
			var strengths []int
			for range 4 {
				m.Advance(50 * time.Millisecond)
				box, strength := redExtent(m.Image(), region)
				boxes, strengths = append(boxes, box), append(strengths, strength)
			}
			for i := range boxes {
				if boxes[i] != boxes[len(boxes)-1] {
					t.Fatalf("the panel moved while it faded in: %v", boxes)
				}
			}
			if strengths[0] >= strengths[1] || strengths[1] >= strengths[3] {
				t.Fatalf("the panel's opacity did not rise over 200 ms: %v", strengths)
			}
		})
	}
}

func TestSlideoverSlidesOutAtFull(t *testing.T) {
	open := true
	m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
		nui.Slideover(c, th, &open, nui.SlideoverOptions{Title: "Sheet", UI: redPanel, Side: nui.SideLeft}, func(nui.Theme) { ui.Box(c).Height(80) })
	})
	m.Advance(400 * time.Millisecond)
	region := ui.Rect{W: 700, H: 500}
	rest, _ := redExtent(m.Image(), region)
	open = false
	m.Frame()
	m.Advance(100 * time.Millisecond)
	if mid, _ := redExtent(m.Image(), region); mid.W == 0 || mid.X+mid.W >= rest.X+rest.W {
		t.Fatalf("at 100 ms the panel was not sliding toward the left edge: %v (at rest %v)", mid, rest)
	}
	m.Advance(120 * time.Millisecond)
	if box, _ := redExtent(m.Image(), region); box.W != 0 {
		t.Fatalf("the panel stayed after 200 ms: %v", box)
	}
}

func toastView(th nui.Theme, c *ui.Context) {
	nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
		nui.Toast(c, th, nui.ToastOptions{ID: "note", Title: "Saved", Timeout: time.Hour, Progress: nui.Set(false)})
	})
	nui.ToastViewport(c, th, nui.ToastViewportOptions{})
}

func TestToastFadesInPlaceAtReduced(t *testing.T) {
	m := newMotionTester(motionReducedDesktop, 600, 400, func(c *ui.Context, th nui.Theme) { toastView(th, c) })
	click(t, m.Tester, "Notify")
	m.Move(5, 5)
	var ys []float32
	for range 5 {
		ys = append(ys, bounds(t, m.Tester, "Saved").Y)
		m.Advance(50 * time.Millisecond)
	}
	for _, y := range ys {
		if y != ys[len(ys)-1] {
			t.Fatalf("a toast at Reduced slid: %v", ys)
		}
	}
}

func TestToastExitTakes200msAtFull(t *testing.T) {
	m := newMotionTester(motionFull, 600, 400, func(c *ui.Context, th nui.Theme) {
		toastView(th, c)
		nui.Button(c, th, "Dismiss", nui.ButtonOptions{}).OnClick(func() { nui.CloseToast(c, "note") })
	})
	click(t, m.Tester, "Notify")
	m.Advance(400 * time.Millisecond)
	panel := bounds(t, m.Tester, "Saved")
	// The pixels of the toast's title that differ from the white window.
	at := func() int {
		n := 0
		im := m.Image()
		for y := int(panel.Y); y < int(panel.Y+panel.H); y++ {
			for x := int(panel.X); x < int(panel.X+panel.W); x++ {
				if p := im.RGBAAt(x, y); p.R < 250 || p.G < 250 || p.B < 250 {
					n++
				}
			}
		}
		return n
	}
	shown := at()
	click(t, m.Tester, "Dismiss")
	m.Move(5, 5)
	m.Advance(100 * time.Millisecond)
	if mid := at(); mid == 0 || mid >= shown {
		t.Fatalf("at 100 ms the toast was not leaving: %d of %d pixels", mid, shown)
	}
	m.Advance(120 * time.Millisecond)
	if end := at(); end != 0 {
		t.Fatalf("the toast was still painted after 200 ms: %d pixels", end)
	}
}

// Every popup leaves with Nuxt's 100 ms scale-out at Full and a fade at
// Reduced, painted for its exit and gone after it.
func TestPopupExitsAtEachLevel(t *testing.T) {
	type popup struct {
		name  string
		view  func(c *ui.Context, th nui.Theme)
		open  func(t *testing.T, m *motionTester)
		close func(t *testing.T, m *motionTester)
	}
	escape := func(t *testing.T, m *motionTester) { m.Key(0, ui.KeyEscape) }
	options := []nui.Option[string]{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
	popups := func() []popup {
		choice := "a"
		var search nui.SearchState
		var menu nui.SelectMenuState
		var picked []string
		var context nui.MenuState
		selectState11 := submittedSelect(options)
		listSelection1 := chosenSelection(picked)
		menu.SetItems([]nui.ListboxItem{{ID: "a", Label: "Alpha"}})
		search.SetOptions(options)
		context.SetItems([]nui.MenuItem{{ID: "a", Label: "Alpha"}})
		return []popup{
			{"Select", func(c *ui.Context, th nui.Theme) {
				nui.Select(c, th, &choice, selectState11, nui.SelectOptions{Label: "Choice", Popup: nui.PopoverOptions{UI: redPanel}}).Width(240)
			}, func(t *testing.T, m *motionTester) { click(t, m.Tester, "Choice") }, escape},
			{"SelectMenu", func(c *ui.Context, th nui.Theme) {
				nui.SelectMenu(c, th, listSelection1, &menu, nui.SelectMenuOptions{Label: "Menu", Popup: nui.PopoverOptions{UI: redPanel}}).Width(240)
			}, func(t *testing.T, m *motionTester) { click(t, m.Tester, "Menu") }, escape},
			{"InputMenu", func(c *ui.Context, th nui.Theme) {
				nui.InputMenu(c, th, &search, nui.InputMenuOptions{Label: "Search", Popup: nui.PopoverOptions{UI: redPanel}}).Width(240)
			}, func(t *testing.T, m *motionTester) { click(t, m.Tester, "Search"); m.Key(0, ui.KeyDown) }, escape},
			{"ContextMenu", func(c *ui.Context, th nui.Theme) {
				a := ui.Box(c).Size(200, 60).Label("Area")
				nui.ContextMenu(c, th, a, &context, nui.MenuOptions{UI: redPanel})
			}, func(t *testing.T, m *motionTester) {
				if err := m.RightClick("Area"); err != nil {
					t.Fatal(err)
				}
			}, escape},
			{"Tooltip", func(c *ui.Context, th nui.Theme) {
				a := nui.Button(c, th, "Save", nui.ButtonOptions{})
				nui.Tooltip(c, th, a, "Save the draft before closing", nui.TooltipOptions{DelayDuration: nui.Set(time.Duration(0)), UI: redPanel})
			}, func(t *testing.T, m *motionTester) {
				r := bounds(t, m.Tester, "Save")
				m.Move(r.X+r.W/2, r.Y+r.H/2)
			}, func(t *testing.T, m *motionTester) { m.Move(695, 495) }},
		}
	}
	for i := range len(popups()) {
		for _, l := range []motionLevel{motionFull, motionReducedTheme} {
			p := popups()[i]
			t.Run(p.name+"/"+l.name, func(t *testing.T) {
				m := newMotionTester(l, 700, 500, p.view)
				p.open(t, m)
				m.Advance(300 * time.Millisecond)
				region := ui.Rect{W: 700, H: 500}
				rest, full := redExtent(m.Image(), region)
				if rest.W == 0 {
					t.Fatal("the popup did not open")
				}
				p.close(t, m)
				m.Advance(40 * time.Millisecond)
				box, strength := redExtent(m.Image(), region)
				if box.W == 0 || strength >= full {
					t.Fatalf("40 ms into the exit the popup was not leaving: %v at %d of %d", box, strength, full)
				}
				// At 40 ms a scale-out is about 4% smaller; a fade alone loses
				// at most the faintest pixel of each edge.
				if smaller := box.W < rest.W*.98-1; smaller != (l == motionFull) {
					t.Fatalf("40 ms into the exit the popup was %v, at rest %v", box, rest)
				}
				m.Advance(80 * time.Millisecond)
				if box, _ := redExtent(m.Image(), region); box.W != 0 {
					t.Fatalf("the popup was still painted after its exit: %v", box)
				}
			})
		}
	}
}

func TestModalReturnsFocusAsItLeaves(t *testing.T) {
	open := false
	m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
		nui.Button(c, th, "Open dialog", nui.ButtonOptions{}).OnClick(func() { open = true })
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Edit", UI: redPanel}, func(nui.Theme) { ui.Box(c).Height(80) })
	})
	m.Key(0, ui.KeyTab)
	m.Key(0, ui.KeyEnter)
	m.Advance(300 * time.Millisecond)
	if m.Focused("Open dialog") {
		t.Fatal("focus stayed on the opener behind the Modal")
	}
	m.Key(0, ui.KeyEscape)
	if !m.Focused("Open dialog") {
		t.Fatal("focus did not return to the opener in the frame the Modal closed")
	}
	if box, _ := redExtent(m.Image(), ui.Rect{W: 700, H: 500}); box.W == 0 {
		t.Fatal("the Modal was not leaving as focus returned")
	}
}

// A Slideover or Drawer leaving is inert: what is behind takes input from
// the first frame of the exit, and focus returns as it closes.
func TestLeavingSheetsAreInert(t *testing.T) {
	for _, kind := range []string{"slideover", "drawer"} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/returnFocus=%v", kind, explicit), func(t *testing.T) {
				open, closes, behind := false, 0, 0
				drawer := nui.DrawerState{}
				var other ui.Handle
				m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
					nui.Button(c.Key("open"), th, "Open sheet", nui.ButtonOptions{}).OnClick(func() { open, drawer.Open = true, true })
					nui.Button(c.Key("other"), th, "Other", nui.ButtonOptions{}).Bind(&other).OnClick(func() { behind++ })
					mo := nui.ModalOptions{Title: "Sheet", OnClose: func() { closes++ }}
					if explicit {
						mo.ReturnFocus = &other
					}
					if kind == "drawer" {
						nui.Drawer(c, th, &drawer, nui.DrawerOptions{ModalOptions: mo}, func(nui.Theme) { ui.Text(c, "Sheet body") })
						open = drawer.Open
					} else {
						nui.Slideover(c, th, &open, nui.SlideoverOptions{ModalOptions: mo}, func(nui.Theme) { ui.Text(c, "Sheet body") })
					}
				})
				m.UseScreenReader()
				m.Key(0, ui.KeyTab)
				if !m.Focused("Open sheet") {
					t.Fatal("Tab did not focus the opener")
				}
				m.Key(0, ui.KeyEnter)
				m.Advance(400 * time.Millisecond)
				if m.Focused("Open sheet") {
					t.Fatal("focus stayed behind the open sheet")
				}
				under := bounds(t, m.Tester, "Other")
				m.Key(0, ui.KeyEscape)
				if open || closes != 1 {
					t.Fatalf("Escape: open=%v, OnClose ran %d times", open, closes)
				}
				want := "Open sheet"
				if explicit {
					want = "Other"
				}
				if !m.Focused(want) {
					t.Fatalf("focus did not return to %q in the closing frame", want)
				}
				for _, info := range m.AccessAll() {
					if info.Role == ui.RoleDialog || info.Name == "Sheet body" {
						t.Fatalf("the leaving sheet is exposed to assistive technology: %+v", info)
					}
				}
				// Still leaving, it takes no input: the click goes behind it.
				// This point is outside both the right Slideover and the
				// bottom Drawer, so it measures the backdrop, not the panel.
				leavingCorner := m.at(240, 10)
				m.ClickAt(under.X+under.W/2, under.Y+under.H/2)
				if behind != 1 {
					t.Fatal("the window behind took no click in the exit's first frame")
				}
				m.Advance(400 * time.Millisecond)
				if m.at(240, 10) == leavingCorner {
					t.Fatal("the sheet's backdrop was not showing as the click went behind it")
				}
				if closes != 1 {
					t.Fatalf("OnClose ran %d times", closes)
				}
			})
		}
	}
}
