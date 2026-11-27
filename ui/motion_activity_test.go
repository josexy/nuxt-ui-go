package nui_test

import (
	"fmt"
	"image"
	"image/color"
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

// nearColor reports whether p is within 40 of c in every channel.
func nearColor(p color.RGBA, c ui.Color) bool {
	d := func(a, b uint8) bool { return math.Abs(float64(a)-float64(b)) <= 40 }
	return d(p.R, c.R) && d(p.G, c.G) && d(p.B, c.B)
}

// span returns the first and last pixel, from r's start, of c along the
// middle of r, across it or, when vertical, down it, and how many there are.
func span(im *image.RGBA, r ui.Rect, c ui.Color, vertical bool) (first, last, n int) {
	first, last = -1, -1
	length := int(r.W)
	if vertical {
		length = int(r.H)
	}
	for i := range length {
		x, y := int(r.X)+i, int(r.Y+r.H/2)
		if vertical {
			x, y = int(r.X+r.W/2), int(r.Y)+i
		}
		if nearColor(im.RGBAAt(x, y), c) {
			if first < 0 {
				first = i
			}
			last, n = i, n+1
		}
	}
	return first, last, n
}

func labelled(name string) func(nui.Slot, ui.Element) {
	return func(slot nui.Slot, e ui.Element) {
		if slot == nui.Indicator {
			e.Label(name)
		}
	}
}

func TestProgressSharesTablePresets(t *testing.T) {
	var table nui.TableState[string]
	table.SetRows([]string{"One"}, func(v string) string { return v })
	cols := []nui.TableColumn[string]{{ID: "name", Title: "Name", Width: 300, Value: func(v string) string { return v }}}
	primary := nui.Light().Primary
	for _, preset := range []nui.LoadingAnimation{nui.LoadingCarousel, nui.LoadingCarouselInverse, nui.LoadingSwing, nui.LoadingElastic} {
		t.Run(fmt.Sprint(preset), func(t *testing.T) {
			m := newMotionTester(motionFull, 400, 400, func(c *ui.Context, th nui.Theme) {
				nui.Table(c.Key("table"), th, &table, cols, nui.TableOptions{Label: "Rows", Loading: true, LoadingAnimation: preset, Height: 120, Style: labelled("Table bar")}).Width(300)
				ui.Column(c.Key("bar")).Width(300).Children(func() {
					nui.Progress(c, th, 0, nui.ProgressOptions{Label: "Progress bar", Indeterminate: true, Animation: preset})
				})
			})
			for range 5 {
				m.Advance(170 * time.Millisecond)
				tb, pb := bounds(t, m.Tester, "Table bar"), bounds(t, m.Tester, "Progress bar")
				_, _, tn := span(m.Image(), tb, primary, false)
				_, _, pn := span(m.Image(), pb, primary, false)
				if tf, pf := float64(tn)/float64(tb.W), float64(pn)/float64(pb.W); math.Abs(tf-pf) > .02 {
					t.Fatalf("preset %d covers %.3f of the table's track and %.3f of the progress track", preset, tf, pf)
				}
			}
		})
	}
	if nui.TableLoadingElastic != nui.LoadingElastic || nui.TableLoadingCarouselInverse != nui.LoadingCarouselInverse {
		t.Fatal("the Table presets are not the shared ones")
	}
}

// progressSamples returns where the indicator of a Progress labelled "Bar"
// starts and ends, every 100 ms over one 2 s cycle and once more at its end.
func progressSamples(t *testing.T, m *motionTester, vertical bool) (starts, ends []int) {
	t.Helper()
	primary := nui.Light().Primary
	for range 21 {
		first, last, _ := span(m.Image(), bounds(t, m.Tester, "Bar"), primary, vertical)
		starts, ends = append(starts, first), append(ends, last)
		m.Advance(100 * time.Millisecond)
	}
	return starts, ends
}

func TestProgressSwingCycle(t *testing.T) {
	m := newMotionTester(motionFull, 400, 100, func(c *ui.Context, th nui.Theme) {
		ui.Column(c).Width(300).Children(func() {
			nui.Progress(c, th, 0, nui.ProgressOptions{Label: "Bar", Indeterminate: true, Animation: nui.LoadingSwing})
		})
	})
	starts, ends := progressSamples(t, m, false)
	w := int(bounds(t, m.Tester, "Bar").W)
	lo, hi := w, 0
	for i := range starts {
		lo, hi = min(lo, starts[i]), max(hi, ends[i])
	}
	if lo > 2 || hi < w-3 {
		t.Fatalf("the swing did not reach both ends of %d DIPs: from %d to %d (%v)", w, lo, hi, starts)
	}
	if d := starts[20] - starts[0]; d < -1 || d > 1 {
		t.Fatalf("the swing was not back where it started after 2 s: %d -> %d", starts[0], starts[20])
	}
}

func TestProgressElasticVertical(t *testing.T) {
	m := newMotionTester(motionFull, 200, 300, func(c *ui.Context, th nui.Theme) {
		nui.Progress(c, th, 0, nui.ProgressOptions{Label: "Bar", Indeterminate: true, Animation: nui.LoadingElastic, Orientation: nui.Vertical})
	})
	starts, ends := progressSamples(t, m, true)
	h := int(bounds(t, m.Tester, "Bar").H)
	shortest, longest := h, 0
	for i := range starts {
		length := ends[i] - starts[i] + 1
		shortest, longest = min(shortest, length), max(longest, length)
		if mid := (starts[i] + ends[i]) / 2; mid < h/2-2 || mid > h/2+2 {
			t.Fatalf("the elastic left the middle of the %d DIP track: %d to %d", h, starts[i], ends[i])
		}
	}
	if shortest > h/2+4 || longest < h-4 {
		t.Fatalf("the elastic did not grow from half to all of %d DIPs: %d to %d", h, shortest, longest)
	}
}

func TestProgressCarouselRightToLeft(t *testing.T) {
	for _, rtl := range []bool{false, true} {
		t.Run(fmt.Sprint("rtl=", rtl), func(t *testing.T) {
			m := newMotionTester(motionFull, 400, 100, func(c *ui.Context, th nui.Theme) {
				ui.Column(c).Width(300).Children(func() {
					nui.Progress(c, th, 0, nui.ProgressOptions{Label: "Bar", Indeterminate: true})
				})
			})
			if rtl {
				m.SetDirection(ui.RightToLeft)
			}
			// Over 50 ms steps the indicator moves about 11 DIPs, and once a
			// cycle jumps back to where it starts.
			primary, forward, back := nui.Light().Primary, 0, 0
			r := bounds(t, m.Tester, "Bar")
			before, _, _ := span(m.Image(), r, primary, false)
			for range 40 {
				m.Advance(50 * time.Millisecond)
				first, _, _ := span(m.Image(), r, primary, false)
				_, last, _ := span(m.Image(), r, primary, false)
				at := first
				if rtl {
					at = last
				}
				if before >= 0 && at >= 0 {
					if at > before {
						forward++
					} else if at < before {
						back++
					}
				}
				before = at
			}
			if rtl == (forward > back) {
				t.Fatalf("rtl=%v: the indicator moved right %d times and left %d times", rtl, forward, back)
			}
		})
	}
}

func TestLoadingAnimationValidation(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		view       func(c *ui.Context, th nui.Theme)
	}{
		{"progress", "nui: invalid progress loading animation", func(c *ui.Context, th nui.Theme) {
			nui.Progress(c, th, 0, nui.ProgressOptions{Indeterminate: true, Animation: nui.LoadingElastic + 1})
		}},
		{"table", "nui: invalid table loading animation", func(c *ui.Context, th nui.Theme) {
			var s nui.TableState[string]
			nui.Table(c, th, &s, nil, nui.TableOptions{Label: "Rows", LoadingAnimation: nui.TableLoadingElastic + 1})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if got := fmt.Sprint(recover()); got != tc.want {
					t.Fatalf("panic %q, want %q", got, tc.want)
				}
			}()
			newMotionTester(motionFull, 300, 200, tc.view)
		})
	}
}

func TestTableLoadingPulsesAtReduced(t *testing.T) {
	var table nui.TableState[string]
	table.SetRows([]string{"One"}, func(v string) string { return v })
	cols := []nui.TableColumn[string]{{ID: "name", Title: "Name", Width: 300, Value: func(v string) string { return v }}}
	m := newMotionTester(motionReducedDesktop, 400, 300, func(c *ui.Context, th nui.Theme) {
		nui.Table(c, th, &table, cols, nui.TableOptions{Label: "Rows", Loading: true, Height: 120, Style: labelled("Table bar")}).Width(300)
	})
	r := bounds(t, m.Tester, "Table bar")
	first := m.at(r.X+r.W/2, r.Y+r.H/2)
	changed := false
	for range 10 {
		m.Advance(100 * time.Millisecond)
		im := m.Image()
		for _, x := range []float32{r.X + 1, r.X + r.W/2, r.X + r.W - 2} {
			if p := im.RGBAAt(int(x), int(r.Y+r.H/2)); p == im.RGBAAt(int(x), int(r.Y+r.H/2)+3) {
				t.Fatalf("the loading bar did not cover the header at x=%v: %v", x, p)
			}
		}
		changed = changed || m.at(r.X+r.W/2, r.Y+r.H/2) != first
	}
	if !changed || !m.FramePending() {
		t.Fatalf("the loading bar did not pulse at Reduced (pending=%v)", m.FramePending())
	}
}

func TestIndeterminateProgressAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			m := newMotionTester(l, 400, 100, func(c *ui.Context, th nui.Theme) {
				ui.Column(c).Width(300).Children(func() {
					nui.Progress(c, th, 0, nui.ProgressOptions{Label: "Bar", Indeterminate: true})
				})
			})
			r := bounds(t, m.Tester, "Bar")
			primary := nui.Light().Primary
			first, last, _ := span(m.Image(), r, primary, false)
			switch {
			case l.off:
				if w := int(r.W); first < w/4-2 || first > w/4+2 || last < 3*w/4-3 || last > 3*w/4+1 {
					t.Fatalf("at Off the indicator is not centred at half length: %d to %d of %d", first, last, w)
				}
				m.idle(t, 2*time.Second, "an indeterminate Progress at Off")
			case l.reduced():
				// A pulse fades the full-length indicator toward the track:
				// its start, middle and end show the same colour.
				y, track := r.Y+r.H/2, rgba(nui.Light().Accented)
				if a, b, c := m.at(r.X+2, y), m.at(r.X+r.W/2, y), m.at(r.X+r.W-3, y); a != b || b != c || b == track {
					t.Fatalf("at Reduced the indicator does not span the track: %v %v %v", a, b, c)
				}
				before, changed := m.at(r.X+r.W/2, r.Y+r.H/2), false
				for range 4 {
					m.Advance(250 * time.Millisecond)
					changed = changed || m.at(r.X+r.W/2, r.Y+r.H/2) != before
				}
				if !changed || !m.FramePending() {
					t.Fatal("at Reduced the indicator did not pulse")
				}
			default:
				// The indicator may be clipped at an end: compare both edges.
				m.Advance(500 * time.Millisecond)
				if f, l, _ := span(m.Image(), r, primary, false); f == first && l == last {
					t.Fatal("at Full the indicator did not travel")
				}
			}
		})
	}
}

func TestProgressValueChangeAtEachLevel(t *testing.T) {
	for _, l := range motionLevels {
		t.Run(l.name, func(t *testing.T) {
			value := float64(20)
			m := newMotionTester(l, 400, 100, func(c *ui.Context, th nui.Theme) {
				ui.Column(c).Width(300).Children(func() {
					nui.Progress(c, th, value, nui.ProgressOptions{Label: "Bar"})
				})
			})
			r := bounds(t, m.Tester, "Bar")
			primary := nui.Light().Primary
			value = 80
			m.Frame()
			_, last, _ := span(m.Image(), r, primary, false)
			at := float64(last+1) / float64(r.W)
			if l == motionFull {
				if at > .7 {
					t.Fatalf("at Full the indicator jumped to %.2f", at)
				}
				m.Advance(250 * time.Millisecond)
				_, last, _ = span(m.Image(), r, primary, false)
				at = float64(last+1) / float64(r.W)
			}
			if math.Abs(at-.8) > .02 {
				t.Fatalf("the indicator shows %.2f, want 0.80", at)
			}
		})
	}
}
