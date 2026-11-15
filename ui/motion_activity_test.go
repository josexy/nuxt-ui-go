package nui_test

import (
	"fmt"
	"image"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// spinnerMarker is a loading icon of one bar above its centre, whose turn
// shows where the bar is. Components tint it with their colour.
var spinnerMarker = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24"><rect x="9" y="1" width="6" height="9" fill="#000"/></svg>`))

// markerAngle returns the direction of the marker's pixels, those unlike the
// background at r's corner, from the centre of r, in degrees clockwise from
// the positive x axis.
func markerAngle(t *testing.T, im *image.RGBA, r ui.Rect) float64 {
	t.Helper()
	var sx, sy, n float64
	cx, cy := float64(r.X+r.W/2), float64(r.Y+r.H/2)
	bg := im.RGBAAt(int(r.X), int(r.Y))
	far := func(a, b uint8) bool { return int(a)-int(b) > 60 || int(b)-int(a) > 60 }
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if p := im.RGBAAt(x, y); far(p.R, bg.R) || far(p.G, bg.G) || far(p.B, bg.B) {
				sx, sy, n = sx+float64(x)+.5-cx, sy+float64(y)+.5-cy, n+1
			}
		}
	}
	if n == 0 {
		t.Fatal("the loading icon drew nothing")
	}
	return math.Atan2(sy/n, sx/n) * 180 / math.Pi
}

// turned returns the clockwise turn from a to b, in (-180, 180].
func turned(a, b float64) float64 {
	d := math.Mod(b-a+540, 360) - 180
	if d == -180 {
		d = 180
	}
	return d
}

func TestSpinnerTurnsAQuarterAtReduced(t *testing.T) {
	m := newMotionTester(motionReducedDesktop, 300, 100, func(c *ui.Context, th nui.Theme) {
		nui.Button(c, th, "Saving", nui.ButtonOptions{Loading: true, LoadingIcon: spinnerMarker, Style: func(_, _, leading, _ ui.Element) {
			leading.Label("Spinner")
		}})
	})
	r := bounds(t, m.Tester, "Spinner")
	start := markerAngle(t, m.Image(), r)
	m.Advance(250 * time.Millisecond)
	if d := turned(start, markerAngle(t, m.Image(), r)); math.Abs(d-90) > 12 {
		t.Fatalf("a spinner at Reduced turned %.1f degrees in 250 ms, want 90", d)
	}
}

func TestSpinnerStillAtOff(t *testing.T) {
	text := ""
	m := newMotionTester(motionOff, 300, 100, func(c *ui.Context, th nui.Theme) {
		nui.Input(c, th, &text, nui.InputOptions{Label: "Search", Loading: true, LoadingIcon: spinnerMarker, Style: func(_, _, leading, _ ui.Element) {
			leading.Label("Spinner")
		}})
	})
	r := bounds(t, m.Tester, "Spinner")
	start := markerAngle(t, m.Image(), r)
	for range 8 {
		m.idle(t, 250*time.Millisecond, "a loading Input at Off")
		if a := markerAngle(t, m.Image(), r); a != start {
			t.Fatalf("a spinner at Off moved from %.1f to %.1f degrees", start, a)
		}
	}
}

func TestSkeletonPulsesAtReduced(t *testing.T) {
	m := newMotionTester(motionReducedDesktop, 300, 100, func(c *ui.Context, th nui.Theme) {
		nui.Skeleton(c, th, nui.SkeletonOptions{})
	})
	start := slices.Clone(m.Image().Pix)
	changed := false
	for range 4 {
		m.Advance(250 * time.Millisecond)
		changed = changed || !slices.Equal(start, m.Image().Pix)
	}
	if !changed || !m.FramePending() {
		t.Fatalf("a Skeleton at Reduced did not pulse over 1 s (pending=%v)", m.FramePending())
	}
}

func TestMarqueeStillAtOff(t *testing.T) {
	var state nui.MarqueeState
	m := newMotionTester(motionOff, 500, 200, func(c *ui.Context, th nui.Theme) {
		nui.Marquee(c, th, &state, nui.MarqueeOptions{Label: "Logos", Overlay: nui.Set(false)}, func() {
			for i := range 3 {
				ui.Box(c.Key(i)).Size(40, 24).Background(ui.RGB(200, 30, 30))
			}
		}).Width(300)
	})
	region := bounds(t, m.Tester, "Logos")
	before := redEdges(m.Image(), region, false)
	b, _ := m.span(2 * time.Second)
	if b > 2 || m.FramePending() || state.Offset != 0 {
		t.Fatalf("a Marquee at Off built %d times in 2 s, offset %v (pending=%v)", b, state.Offset, m.FramePending())
	}
	if after := redEdges(m.Image(), region, false); fmt.Sprint(before) != fmt.Sprint(after) {
		t.Fatalf("a Marquee at Off moved: %v -> %v", before, after)
	}
}

func TestCarouselAutoplayOnlyAtFull(t *testing.T) {
	items := []nui.CarouselItem{
		{ID: "a", Label: "Slide A", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "A") }},
		{ID: "b", Label: "Slide B", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "B") }},
		{ID: "c", Label: "Slide C", Build: func(c *ui.Context, _ nui.Theme) { ui.Text(c, "C") }},
	}
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			var s nui.CarouselState
			m := newMotionTester(l, 500, 300, func(c *ui.Context, th nui.Theme) {
				nui.Carousel(c, th, &s, items, nui.CarouselOptions{Label: "Slides", Autoplay: 500 * time.Millisecond, Height: 120, Arrows: nui.Set(true)}).Width(400)
			})
			m.Move(490, 290) // off the carousel, which pauses under the pointer
			m.Advance(1200 * time.Millisecond)
			if played := s.Index != 0; played != (l == motionFull) {
				t.Fatalf("autoplay advanced to %d", s.Index)
			}
			if l == motionFull {
				return
			}
			if m.FramePending() {
				t.Fatal("a carousel without autoplay asked for frames")
			}
			click(t, m.Tester, "Next Slides")
			if s.Index != 1 {
				t.Fatalf("the Next arrow went to %d", s.Index)
			}
		})
	}
}

func TestToastCountdownAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			m := newMotionTester(l, 600, 400, func(c *ui.Context, th nui.Theme) {
				nui.Button(c, th, "Notify", nui.ButtonOptions{}).OnClick(func() {
					nui.Toast(c, th, nui.ToastOptions{Title: "Countdown", Color: nui.Error, Timeout: time.Second})
				})
				nui.ToastViewport(c, th, nui.ToastViewportOptions{})
			})
			click(t, m.Tester, "Notify")
			m.Move(5, 5)
			lengths := map[int]bool{}
			for range 66 { // 1.1 s
				m.Advance(time.Second / 60)
				panel, ok := m.Find("Countdown")
				if !ok {
					break
				}
				im, n := m.Image(), 0
				for x := int(panel.X); x < int(panel.X+panel.W); x++ {
					if p := im.RGBAAt(x, int(panel.Y+panel.H-2)); p.R > 150 && p.G < 120 {
						n++
					}
				}
				lengths[n] = true
			}
			if m.HasText("Countdown") {
				t.Fatal("the toast did not close once its second ran out")
			}
			if stepped := len(lengths) <= 5; stepped == (l == motionFull) {
				t.Fatalf("the countdown bar was painted at %d lengths", len(lengths))
			}
		})
	}
}
