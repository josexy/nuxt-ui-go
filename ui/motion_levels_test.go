package nui_test

import (
	"fmt"
	"image/color"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// motionLevel is one way a window resolves nui's motion: Full, Reduced
// through the desktop's preference or through the theme, or Off.
type motionLevel struct {
	name    string
	desktop bool // the desktop asks for reduced motion
	reduce  bool // the theme's Motion.Reduce
	off     bool // AppConfig.DisableAnimations
}

var (
	motionFull           = motionLevel{name: "full"}
	motionReducedDesktop = motionLevel{name: "reduced-desktop", desktop: true}
	motionReducedTheme   = motionLevel{name: "reduced-theme", reduce: true}
	motionOff            = motionLevel{name: "off", off: true}
	motionLevels         = []motionLevel{motionFull, motionReducedDesktop, motionReducedTheme, motionOff}
)

// reduced reports whether the level is one of the Reduced ones.
func (l motionLevel) reduced() bool { return l.desktop || l.reduce }

// config returns the light AppConfig of the level.
func (l motionLevel) config() nui.AppConfig {
	light := nui.Light()
	light.Motion.Reduce = l.reduce
	return nui.AppConfig{Mode: nui.LightMode, Light: light, Dark: nui.Dark(), DisableAnimations: l.off}
}

// motionTester runs a view at a motion level, counting the builds of the
// view and the frames painted whole. A frame that only paints parts again,
// as a spinner's, is not a whole paint: check FramePending for those.
type motionTester struct {
	*ui.Tester
	builds, paints int
}

// newMotionTester runs view, given the theme the level resolves, in a
// width×height window.
func newMotionTester(l motionLevel, width, height int, view func(c *ui.Context, th nui.Theme)) *motionTester {
	m := &motionTester{}
	config := l.config()
	m.Tester = ui.NewTesterWith(func(c *ui.Context) {
		m.builds++
		th := nui.App(c, config)
		ui.Column(c).Key("motion-frame").Fill().Padding(8).Gap(8).
			Draw(func(*ui.Painter, ui.Rect) { m.paints++ }).
			Children(func() { view(c, th) })
	}, width, height, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: l.desktop}})
	return m
}

// span advances the clock by d and returns how many builds and whole paints
// it took.
func (m *motionTester) span(d time.Duration) (builds, paints int) {
	b, p := m.builds, m.paints
	m.Advance(d)
	return m.builds - b, m.paints - p
}

// at returns the colour painted at x, y.
func (m *motionTester) at(x, y float32) color.RGBA { return m.Image().RGBAAt(int(x), int(y)) }

// idle fails when the window builds, paints or asks for a frame while the
// clock advances by d.
func (m *motionTester) idle(t *testing.T, d time.Duration, what string) {
	t.Helper()
	if b, p := m.span(d); b != 0 || p != 0 || m.FramePending() {
		t.Fatalf("%s: %d builds and %d paints in %v (pending=%v)", what, b, p, d, m.FramePending())
	}
}

// between reports whether every channel of v lies between those of a and b,
// and v is neither.
func between(v, a, b color.RGBA) bool {
	in := func(v, a, b uint8) bool { return min(a, b) <= v && v <= max(a, b) }
	return v != a && v != b && in(v.R, a.R, b.R) && in(v.G, a.G, b.G) && in(v.B, a.B, b.B)
}

func TestMotionTesterButtonHoverAtOff(t *testing.T) {
	m := newMotionTester(motionOff, 300, 120, func(c *ui.Context, th nui.Theme) {
		nui.Button(c, th, "Hover me", nui.ButtonOptions{})
	})
	r := bounds(t, m.Tester, "Hover me")
	x, y := r.X+3, r.Y+r.H/2
	idle := m.at(x, y)
	m.Move(r.X+r.W/2, y)
	hover := m.at(x, y)
	if hover == idle {
		t.Fatalf("the hover colour did not show in the next frame: %v", hover)
	}
	m.idle(t, time.Second, "a hovered Button at Off")
	if got := m.at(x, y); got != hover {
		t.Fatalf("the hover colour changed after its frame: %v -> %v", hover, got)
	}
}

// offIdleView builds every continuous activity indicator.
func offIdleView(table *nui.TableState[string], marquee *nui.MarqueeState, carousel *nui.CarouselState) func(c *ui.Context, th nui.Theme) {
	cols := []nui.TableColumn[string]{{ID: "name", Title: "Name", Width: 200, Value: func(v string) string { return v }}}
	slides := []nui.CarouselItem{
		{ID: "a", Label: "Slide A", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "A") }},
		{ID: "b", Label: "Slide B", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "B") }},
	}
	return func(c *ui.Context, th nui.Theme) {
		nui.Button(c.Key("button"), th, "Saving", nui.ButtonOptions{Loading: true})
		nui.Skeleton(c.Key("skeleton"), th, nui.SkeletonOptions{})
		nui.Progress(c.Key("progress"), th, 0, nui.ProgressOptions{Label: "Busy", Indeterminate: true})
		nui.Table(c.Key("table"), th, table, cols, nui.TableOptions{Label: "Rows", Loading: true, Height: 120}).Width(300)
		nui.Marquee(c.Key("marquee"), th, marquee, nui.MarqueeOptions{Label: "Logos"}, func() { ui.Text(c, "Partner") }).Width(300)
		nui.Carousel(c.Key("carousel"), th, carousel, slides, nui.CarouselOptions{Label: "Slides", Autoplay: 300 * time.Millisecond, Height: 60}).Width(300)
	}
}

func TestOffIdleWindowStaysIdle(t *testing.T) {
	var table nui.TableState[string]
	table.SetRows([]string{"One", "Two"}, func(v string) string { return v })
	var marquee nui.MarqueeState
	var carousel nui.CarouselState
	m := newMotionTester(motionOff, 600, 700, offIdleView(&table, &marquee, &carousel))
	b, p := m.span(2 * time.Second)
	if b > 2 || p > b || m.FramePending() {
		t.Fatalf("an idle window at Off built %d times and painted %d frames in 2s (pending=%v)", b, p, m.FramePending())
	}
	m.idle(t, time.Second, "an idle window at Off")
	if carousel.Index != 0 || marquee.Offset != 0 {
		t.Fatalf("content travelled at Off: carousel %d, marquee %v", carousel.Index, marquee.Offset)
	}
}

func TestOffStateChangesShowAtOnce(t *testing.T) {
	tab, open, modal := "first", []string{}, true
	tabs := []nui.TabItem{{ID: "first", Label: "First"}, {ID: "second", Label: "Second"}, {ID: "third", Label: "Third"}}
	items := []nui.AccordionItem{{ID: "q", Label: "Question", Content: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "Answer text").Label("Answer") }}}
	m := newMotionTester(motionOff, 700, 500, func(c *ui.Context, th nui.Theme) {
		nui.Tabs(c.Key("tabs"), th, &tab, tabs, nui.TabsOptions{Label: "Sections", Content: nui.Set(false)}).Width(400)
		nui.Accordion(c.Key("accordion"), th, &open, items, nui.AccordionOptions{}).Width(400)
		ui.Text(c.Key("below"), "Below")
		nui.ModalWithOptions(c.Key("modal"), th, &modal, nui.ModalOptions{Title: "Dialog"}, func(nui.Theme) { ui.Text(c, "Dialog body") })
	})
	if !m.HasText("Dialog body") {
		t.Fatal("the Modal did not show")
	}
	third := bounds(t, m.Tester, "Third")
	idle := m.at(third.X+3, third.Y+third.H/2)
	tab, open, modal = "third", []string{"q"}, false
	m.Frame()
	if m.HasText("Dialog body") {
		t.Fatal("the Modal stayed after its close")
	}
	if got := m.at(third.X+3, third.Y+third.H/2); got == idle || got != rgba(nui.Light().Primary) {
		t.Fatalf("the pill did not move to the third trigger at once: %v -> %v", idle, got)
	}
	answer, below := bounds(t, m.Tester, "Answer"), bounds(t, m.Tester, "Below")
	if below.Y < answer.Y+answer.H {
		t.Fatalf("the item was not fully open in the next frame: answer %v, below %v", answer, below)
	}
	m.idle(t, time.Second, "state changes at Off")
}

func TestTurningAnimationsOffMidEntrance(t *testing.T) {
	open, off, draft := false, false, ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, DisableAnimations: off})
		nui.Button(c, th, "Open", nui.ButtonOptions{}).OnClick(func() { open = true })
		if p := nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Edit"}, func(inner nui.Theme) {
			nui.Input(c.Key("draft"), inner, &draft, nui.InputOptions{Label: "Draft"})
		}); p.Valid() {
			p.Label("Dialog panel")
		}
	}, 800, 500)
	click(t, tt, "Open")
	tt.Advance(20 * time.Millisecond)
	click(t, tt, "Draft")
	tt.Type("kept")
	tt.Advance(30 * time.Millisecond)
	if w := bounds(t, tt, "Dialog panel").W; w >= 512 {
		t.Fatalf("the Modal was not still entering at 50 ms: width %v", w)
	}
	off = true
	tt.Frame()
	tt.Advance(150 * time.Millisecond) // 200 ms after it opened
	if w := bounds(t, tt, "Dialog panel").W; w != 512 {
		t.Fatalf("the Modal was not fully shown by 200 ms: width %v", w)
	}
	if !tt.Focused("Draft") || draft != "kept" {
		t.Fatalf("turning animations off lost focus (%v) or the draft (%q)", tt.Focused("Draft"), draft)
	}
	if tt.FramePending() {
		t.Fatal("frames were still requested at Off")
	}
}

func TestTurningAnimationsBackOn(t *testing.T) {
	open, off := false, true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, DisableAnimations: off})
		ui.Column(c).Width(300).Children(func() {
			nui.Collapsible(c.Key("details"), th, &open, nui.CollapsibleOptions{Label: "Details"}, func(nui.Theme) {
				ui.Box(c).Size(200, 100)
			})
			ui.Text(c.Key("below"), "Below")
		})
	}, 400, 400)
	closed := bounds(t, tt, "Below").Y
	off = false
	tt.Frame()
	open = true
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	mid := bounds(t, tt, "Below").Y
	tt.Advance(200 * time.Millisecond)
	if end := bounds(t, tt, "Below").Y; mid <= closed || mid >= end || end < closed+100 {
		t.Fatalf("the content height did not animate open: closed at %v, at 100 ms %v, open at %v", closed, mid, end)
	}
}

func TestThemedSubtreeOptsOutOfMotion(t *testing.T) {
	inside, outside := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Row(c).Gap(16).Children(func() {
			ui.Column(c.Key("outside")).Width(200).Children(func() {
				nui.Collapsible(c, th, &outside, nui.CollapsibleOptions{Label: "Outside"}, func(nui.Theme) { ui.Box(c).Size(100, 100) })
				ui.Text(c, "Below outside")
			})
			still := th
			still.Motion.Disabled = true
			nui.Themed(c.Key("inside"), still, func(th nui.Theme) {
				nui.Collapsible(c, th, &inside, nui.CollapsibleOptions{Label: "Inside"}, func(nui.Theme) { ui.Box(c).Size(100, 100) })
				ui.Text(c, "Below inside")
			}).Width(200)
		})
	}, 500, 400)
	startOut, startIn := bounds(t, tt, "Below outside").Y, bounds(t, tt, "Below inside").Y
	inside, outside = true, true
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	midIn, midOut := bounds(t, tt, "Below inside").Y, bounds(t, tt, "Below outside").Y
	tt.Advance(300 * time.Millisecond)
	endIn, endOut := bounds(t, tt, "Below inside").Y, bounds(t, tt, "Below outside").Y
	if endIn < startIn+100 || midIn != endIn {
		t.Fatalf("the subtree with motion disabled animated: %v, %v at 100 ms, %v", startIn, midIn, endIn)
	}
	if midOut <= startOut || midOut >= endOut {
		t.Fatalf("the window outside the subtree did not animate: %v, %v at 100 ms, %v", startOut, midOut, endOut)
	}
}

func TestSkeletonAnimatePrecedence(t *testing.T) {
	for _, l := range motionLevels {
		for _, animate := range []nui.Optional[bool]{{}, nui.Set(true), nui.Set(false)} {
			t.Run(fmt.Sprintf("%s/%v", l.name, animate), func(t *testing.T) {
				m := newMotionTester(l, 300, 100, func(c *ui.Context, th nui.Theme) {
					nui.Skeleton(c, th, nui.SkeletonOptions{Animate: animate})
				})
				m.Advance(50 * time.Millisecond)
				before, pulses := slices.Clone(m.Image().Pix), false
				for range 4 {
					m.Advance(250 * time.Millisecond)
					pulses = pulses || !slices.Equal(before, m.Image().Pix)
				}
				want := !l.off && (!animate.Set || animate.Value)
				if pulses != want || m.FramePending() != want {
					t.Fatalf("pulses=%v pending=%v, want %v", pulses, m.FramePending(), want)
				}
			})
		}
	}
}

func TestModalTransitionPrecedence(t *testing.T) {
	red := ui.RGB(220, 0, 0)
	for _, l := range motionLevels {
		for _, transition := range []nui.Optional[bool]{{}, nui.Set(true), nui.Set(false)} {
			t.Run(fmt.Sprintf("%s/%v", l.name, transition), func(t *testing.T) {
				open := false
				m := newMotionTester(l, 800, 500, func(c *ui.Context, th nui.Theme) {
					nui.Button(c, th, "Open", nui.ButtonOptions{}).OnClick(func() { open = true })
					if p := nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Edit", Transition: transition, UI: nui.Styles{nui.Panel: {Background: nui.Set(red)}}}, func(nui.Theme) {
						ui.Box(c).Size(10, 120)
					}); p.Valid() {
						p.Label("Dialog panel")
					}
				})
				click(t, m.Tester, "Open")
				m.Advance(50 * time.Millisecond)
				r := bounds(t, m.Tester, "Dialog panel")
				pixel := m.at(r.X+4, r.Y+r.H-4)
				off := l.off || (transition.Set && !transition.Value)
				switch {
				case off:
					if r.W != 512 || pixel != rgba(red) {
						t.Fatalf("not shown at once: width %v, pixel %v", r.W, pixel)
					}
				case l.reduced():
					if r.W != 512 || pixel == rgba(red) {
						t.Fatalf("Reduced did not fade without scaling: width %v, pixel %v", r.W, pixel)
					}
				default:
					if r.W >= 512 || pixel == rgba(red) {
						t.Fatalf("Full did not scale and fade in: width %v, pixel %v", r.W, pixel)
					}
				}
				m.Advance(300 * time.Millisecond)
				if r := bounds(t, m.Tester, "Dialog panel"); r.W != 512 || m.at(r.X+4, r.Y+r.H-4) != rgba(red) {
					t.Fatalf("the Modal did not settle: %v", r)
				}
				if off {
					open = false
					m.Frame()
					if _, ok := m.Find("Dialog panel"); ok {
						t.Fatal("the Modal did not disappear in a single frame")
					}
					if l.off {
						// Elsewhere the button taking focus back fades its colours.
						m.idle(t, time.Second, "a closed Modal at Off")
					}
				}
			})
		}
	}
}

func TestPopoverEntranceAtEachLevel(t *testing.T) {
	red := ui.RGB(220, 0, 0)
	// Disabled wins when the theme and the desktop both ask for reduced motion.
	offEverywhere := motionLevel{name: "off-reduced", desktop: true, reduce: true, off: true}
	for _, l := range append(slices.Clone(motionLevels), offEverywhere) {
		t.Run(l.name, func(t *testing.T) {
			open := false
			m := newMotionTester(l, 500, 300, func(c *ui.Context, th nui.Theme) {
				a := nui.Button(c, th, "Details", nui.ButtonOptions{}).OnClick(func() { open = true })
				nui.Popover(c, th, a, &open, nui.PopoverOptions{Width: nui.Set(float32(200)), UI: nui.Styles{nui.Panel: {Background: nui.Set(red)}}, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Panel {
						e.Label("Panel")
					}
				}}, func(nui.Theme) { ui.Box(c).Size(10, 40) })
			})
			click(t, m.Tester, "Details")
			if l.off {
				r := bounds(t, m.Tester, "Panel")
				if r.W != 200 || m.at(r.X+3, r.Y+r.H-3) != rgba(red) {
					t.Fatalf("not fully shown in its first frame: %v %v", r, m.at(r.X+3, r.Y+r.H-3))
				}
				m.idle(t, time.Second, "an open Popover at Off")
				return
			}
			m.Advance(50 * time.Millisecond)
			r := bounds(t, m.Tester, "Panel")
			pixel := m.at(r.X+3, r.Y+r.H-3)
			if pixel == rgba(red) {
				t.Fatalf("the panel did not fade in: %v", pixel)
			}
			if scaled := r.W < 200; scaled != (l == motionFull) {
				t.Fatalf("scaled=%v at %s: %v", scaled, l.name, r)
			}
		})
	}
}

// Content expanding moves what follows it, so it opens at once under reduced
// motion as at Off (design D9).
func TestCollapsibleHeightAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			open := false
			m := newMotionTester(l, 400, 400, func(c *ui.Context, th nui.Theme) {
				nui.Collapsible(c.Key("details"), th, &open, nui.CollapsibleOptions{Label: "Details"}, func(nui.Theme) {
					ui.Box(c).Size(200, 100)
				})
				ui.Text(c.Key("below"), "Below")
			})
			closed := bounds(t, m.Tester, "Below").Y
			open = true
			m.Frame()
			first := bounds(t, m.Tester, "Below").Y
			m.Advance(300 * time.Millisecond)
			end := bounds(t, m.Tester, "Below").Y
			if end < closed+100 {
				t.Fatalf("the content did not open: %v -> %v", closed, end)
			}
			if animated := first != end; animated != (l == motionFull) {
				t.Fatalf("animated=%v: closed %v, next frame %v, open %v", animated, closed, first, end)
			}
		})
	}
}
