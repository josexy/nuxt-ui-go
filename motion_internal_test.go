package nui

import (
	"fmt"
	"image"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// marqueeEdges returns where red content starts along the middle of r.
func marqueeEdges(im *image.RGBA, r ui.Rect) []int {
	var at []int
	inside := false
	for x := int(r.X); x < int(r.X+r.W); x++ {
		p := im.RGBAAt(x, int(r.Y+r.H/2))
		red := p.R > 150 && p.G < 120
		if red && !inside {
			at = append(at, x)
		}
		inside = red
	}
	return at
}

// Every build re-expresses the motion of a Marquee: it wraps the strip back
// within its copies, starts a new run of its pose, or a strip of a new
// generation. None of them may move what shows from one frame to the next by
// more than the speed does.
func TestMarqueeMotionIsContinuousAcrossBuilds(t *testing.T) {
	travel, renew := marqueeTravel, marqueeRenew
	defer func() { marqueeTravel, marqueeRenew = travel, renew }()
	marqueeTravel, marqueeRenew = 120, 300 // a run every 3.75s, a new strip every ~9.4s
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprint("reverse=", reverse), func(t *testing.T) {
			var state MarqueeState
			builds := 0
			tt := ui.NewTester(func(c *ui.Context) {
				builds++
				th := App(c, AppConfig{Mode: LightMode})
				Marquee(c, th, &state, MarqueeOptions{Label: "Logos", Overlay: Set(false), Reverse: reverse}, func() {
					for i := range 3 {
						ui.Box(c.Key(i)).Size(40, 24).Background(ui.RGB(200, 30, 30))
					}
				}).Width(300)
			}, 400, 300)
			tt.Advance(100 * time.Millisecond)
			r, _ := tt.Find("Logos")
			previous := marqueeEdges(tt.Image(), r)
			b0, moved := builds, 0
			// 4ms steps see each 60 Hz frame; at 32 DIP/s a frame moves the
			// content half a DIP.
			for range 12 * 250 { // 12s: three runs and a new strip
				tt.Advance(4 * time.Millisecond)
				edges := marqueeEdges(tt.Image(), r)
				if fmt.Sprint(edges) != fmt.Sprint(previous) {
					moved++
				}
				for _, e := range edges {
					if e < int(r.X)+4 || e > int(r.X+r.W)-4 {
						continue // entering or leaving the marquee
					}
					near := false
					for _, p := range previous {
						near = near || e-p >= -2 && e-p <= 2
					}
					if !near {
						t.Fatalf("at %v the content jumped: %v -> %v", tt.Now(), previous, edges)
					}
				}
				previous = edges
			}
			if moved < 12*16 {
				t.Fatalf("the content moved in only %d of 12s of frames", moved)
			}
			if n := builds - b0; n > 12*3 {
				t.Fatalf("12s of motion built the view %d times", n)
			}
		})
	}
}
