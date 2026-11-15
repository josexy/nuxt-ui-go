package nui_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSheetExitPreservesScrollAndLocalState(t *testing.T) {
	for _, drawer := range []bool{false, true} {
		t.Run(fmt.Sprintf("drawer=%v", drawer), func(t *testing.T) {
			open, initialized := true, 0
			state := nui.DrawerState{Open: true}
			m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
				body := func(nui.Theme) {
					first := ui.Box(c.Key("first")).Height(400).Shrink(0).Background(ui.RGB(220, 0, 0))
					ui.Local(first, "draft", func() int { initialized++; return initialized })
					ui.Box(c.Key("second")).Height(600).Shrink(0).Background(ui.RGB(0, 180, 0))
				}
				if drawer {
					state.Open = open
					nui.Drawer(c, th, &state, nui.DrawerOptions{Title: "Sheet"}, body)
				} else {
					nui.Slideover(c, th, &open, nui.SlideoverOptions{Title: "Sheet"}, body)
				}
			})
			m.Advance(300 * time.Millisecond)
			m.Scroll(500, 400, 0, 450)
			m.Advance(300 * time.Millisecond)
			before := m.at(500, 400)
			if before.G <= before.R || initialized != 1 {
				t.Fatalf("sheet did not scroll to its second block: %v, initialized %d", before, initialized)
			}
			open = false
			m.Frame()
			if got := m.at(500, 400); got != before || initialized != 1 {
				t.Fatalf("the first exit frame reset its content: %v -> %v, initialized %d", before, got, initialized)
			}
			m.Advance(40 * time.Millisecond)
			open = true
			m.Frame()
			m.Advance(300 * time.Millisecond)
			if got := m.at(500, 400); got != before || initialized != 1 {
				t.Fatalf("reversing the exit reset its content: %v, initialized %d", got, initialized)
			}
			open = false
			m.Frame()
			m.Advance(300 * time.Millisecond)
			open = true
			m.Frame()
			m.Advance(300 * time.Millisecond)
			if initialized != 2 {
				t.Fatalf("the completed exit did not unmount the contents: initialized %d", initialized)
			}
		})
	}
}

func TestDrawerDragInterruptsSnapAtVisibleHeight(t *testing.T) {
	for _, start := range []int{0, 1} {
		t.Run(fmt.Sprint(start), func(t *testing.T) {
			state := nui.DrawerState{Open: true, Snap: start}
			changes := 0
			m := newMotionTester(motionFull, 500, 600, func(c *ui.Context, th nui.Theme) {
				nui.Drawer(c, th, &state, nui.DrawerOptions{Title: "Sheet", OnSnap: func(int) { changes++ }}, func(nui.Theme) { ui.Text(c, "Content") })
			})
			m.Advance(300 * time.Millisecond)
			state.Snap = 1 - start
			m.Frame()
			m.Advance(50 * time.Millisecond)
			r := bounds(t, m.Tester, "Resize Sheet")
			x, y := r.X+r.W/2, r.Y+r.H/2
			m.Press(x, y)
			m.Move(x, y-10)
			m.Frame()
			if got := drawerTop(t, m); math.Abs(float64(got-(r.Y-10))) > 1 {
				t.Fatalf("a 10 DIP drag from %v jumped to %v", r.Y, got)
			}
			if changes != 0 {
				t.Fatalf("an external snap and held drag fired OnSnap %d times", changes)
			}
			m.Release(x, y-10)
			m.Advance(300 * time.Millisecond)
			if !state.Open || changes > 1 {
				t.Fatalf("releasing the interrupted snap: open=%v, callbacks=%d", state.Open, changes)
			}
		})
	}
}

func TestCarouselJumpKeepsOutgoingPage(t *testing.T) {
	for _, dir := range []string{"ltr", "rtl", "vertical"} {
		for _, jump := range []struct {
			from, to int
			action   string
		}{{4, 0, "Next Slides"}, {0, 4, "Previous Slides"}, {0, 4, "Go to slide 5"}, {4, 0, "Go to slide 1"}} {
			t.Run(fmt.Sprintf("%s/%s", dir, jump.action), func(t *testing.T) {
				state := nui.CarouselState{Index: jump.from}
				items := make([]nui.CarouselItem, 5)
				for i := range items {
					color := ui.RGB(0, 0, 220)
					if i == jump.from {
						color = ui.RGB(220, 0, 0)
					}
					items[i] = nui.CarouselItem{ID: fmt.Sprint(i), Label: fmt.Sprint(i), UI: nui.Styles{nui.Item: {Background: nui.Set(color)}}}
				}
				m := newMotionTester(motionFull, 600, 400, func(c *ui.Context, th nui.Theme) {
					o := nui.CarouselOptions{Label: "Slides", Height: 120, Wrap: true, Arrows: nui.Set(true), Dots: true}
					if dir == "vertical" {
						o.Orientation = nui.Vertical
					}
					nui.Carousel(c, th, &state, items, o).Width(400)
				})
				if dir == "rtl" {
					m.SetDirection(ui.RightToLeft)
				}
				m.Advance(300 * time.Millisecond)
				region := ui.Rect{W: 410, H: 140}
				if redPixels(m.Image(), region) < 1000 {
					t.Fatal("the starting page did not paint")
				}
				click(t, m.Tester, jump.action)
				m.Advance(20 * time.Millisecond)
				if state.Index != jump.to || redPixels(m.Image(), region) < 100 {
					t.Fatalf("the outgoing page vanished during the jump: index=%d", state.Index)
				}
				m.Advance(250 * time.Millisecond)
				if redPixels(m.Image(), region) != 0 {
					t.Fatal("the old page remained after the slide")
				}
				m.Move(590, 390)
				m.Advance(300 * time.Millisecond)
				m.idle(t, time.Second, "a completed carousel jump")
			})
		}
	}
}

func TestOffCancelsPopupExits(t *testing.T) {
	for _, kind := range []string{"popover", "modal", "select", "select-menu", "menu", "tooltip"} {
		for _, delay := range []time.Duration{0, 20 * time.Millisecond} {
			t.Run(fmt.Sprintf("%s/%v", kind, delay), func(t *testing.T) {
				open, off := false, false
				choice := "a"
				var selected []string
				var selectMenu nui.SelectMenuState
				var menu nui.MenuState
				m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
					th.Motion.Disabled = off
					a := nui.Button(c.Key("open"), th, "Open", nui.ButtonOptions{}).OnClick(func() { open = true })
					switch kind {
					case "popover":
						nui.Popover(c, th, a, &open, nui.PopoverOptions{UI: redPanel}, func(nui.Theme) { ui.Box(c).Size(200, 100) })
					case "modal":
						nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Panel", UI: redPanel}, func(nui.Theme) { ui.Box(c).Height(100) })
					case "select":
						nui.Select(c, th, &choice, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.SelectOptions{Label: "Choice", Open: &open, Popup: nui.PopoverOptions{UI: redPanel}}).Width(240)
					case "select-menu":
						selectMenu.Open = open
						nui.SelectMenu(c, th, &selected, &selectMenu, []nui.ListboxItem{{ID: "a", Label: "Alpha"}}, nui.SelectMenuOptions{Label: "Choice", Popup: nui.PopoverOptions{UI: redPanel}}).Width(240)
					case "menu":
						menu.Open = open
						nui.DropdownMenu(c, th, a, &menu, []nui.MenuItem{{ID: "a", Label: "Alpha"}}, nui.MenuOptions{UI: redPanel})
					case "tooltip":
						nui.Tooltip(c, th, a, "Explanation of this setting", nui.TooltipOptions{Open: &open, DelayDuration: nui.Set(time.Hour), UI: redPanel})
					}
				})
				m.Move(690, 490)
				open = true
				m.Frame()
				m.Advance(300 * time.Millisecond)
				region := ui.Rect{W: 700, H: 500}
				if redPixels(m.Image(), region) == 0 {
					t.Fatal("the popup did not open")
				}
				open = false
				if delay > 0 {
					m.Frame()
					m.Advance(delay)
				}
				off = true
				m.Frame()
				if redPixels(m.Image(), region) != 0 {
					t.Fatal("Off kept the exit copy")
				}
				// Tooltip may request one cleanup build when its open state
				// changes. No animation may remain after that build.
				m.Advance(time.Millisecond)
				m.idle(t, 300*time.Millisecond, "a cancelled popup exit")
				// Reusing a cancelled surface must not suppress the next entrance.
				off, open = false, true
				m.Frame()
				m.Advance(300 * time.Millisecond)
				if redPixels(m.Image(), region) == 0 {
					t.Fatal("the popup did not reopen after cancellation")
				}
			})
		}
	}
}

func TestClosedSelectDoesNotLookLikeAnExitCopy(t *testing.T) {
	choice := "a"
	m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
		nui.Select(c, th, &choice, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.SelectOptions{Label: "Choice", Popup: nui.PopoverOptions{UI: redPanel}}).Width(240)
	})
	if box, _ := redExtent(m.Image(), ui.Rect{W: 700, H: 500}); box.W != 0 {
		t.Fatalf("a closed, never-opened select looks like an exit copy: %v", box)
	}
}

func TestCancellingPopupExitKeepsOtherMotion(t *testing.T) {
	open, off, on := true, false, false
	m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
		a := nui.Button(c.Key("anchor"), th, "Details", nui.ButtonOptions{})
		nui.Switch(c.Key("switch"), th, &on, nui.SwitchOptions{Label: "Other motion", Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Thumb {
				e.Label("Thumb")
			}
		}})
		th.Motion.Disabled = off
		nui.Popover(c.Key("popover"), th, a, &open, nui.PopoverOptions{UI: redPanel}, func(nui.Theme) { ui.Box(c).Size(200, 100) })
	})
	m.Advance(300 * time.Millisecond)
	start := bounds(t, m.Tester, "Thumb").X
	open, on = false, true
	m.Frame()
	m.Advance(20 * time.Millisecond)
	off = true
	m.Frame()
	m.Advance(50 * time.Millisecond)
	mid := bounds(t, m.Tester, "Thumb").X
	m.Advance(200 * time.Millisecond)
	end := bounds(t, m.Tester, "Thumb").X
	if !(start < mid && mid < end) {
		t.Fatalf("cancelling one exit interrupted another component: %v -> %v -> %v", start, mid, end)
	}
}

func TestModalAfterCancelledExitRestoresNewOpener(t *testing.T) {
	open, off := false, false
	m := newMotionTester(motionFull, 700, 500, func(c *ui.Context, th nui.Theme) {
		for _, label := range []string{"First opener", "Second opener"} {
			nui.Button(c.Key(label), th, label, nui.ButtonOptions{}).OnClick(func() { open = true })
		}
		th.Motion.Disabled = off
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Dialog"}, func(nui.Theme) { ui.Text(c, "Body") })
	})
	click(t, m.Tester, "First opener")
	m.Advance(300 * time.Millisecond)
	open, off = false, true
	m.Frame()
	if !m.Focused("First opener") {
		t.Fatal("cancellation did not return focus")
	}
	click(t, m.Tester, "Second opener")
	open, off = false, false
	m.Frame()
	if !m.Focused("Second opener") {
		t.Fatal("the next close restored the obsolete opener")
	}
}

func TestCancelledPopupDoesNotRekeyFollowingControl(t *testing.T) {
	open, choice, draft := false, "a", "kept"
	var id uint64
	m := newMotionTester(motionOff, 500, 300, func(c *ui.Context, th nui.Theme) {
		nui.Select(c.Key("select"), th, &choice, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.SelectOptions{Label: "Choice", Open: &open})
		// Positional identity must not inherit a key from a closed popup.
		id = ui.TextInput(c, &draft).Label("Draft").ID()
	})
	before := id
	for range 2 {
		open = true
		m.Frame()
		open = false
		m.Frame()
	}
	if before != id || draft != "kept" {
		t.Fatalf("opening and cancelling a popup rekeyed the next control: %d -> %d", before, id)
	}
}
