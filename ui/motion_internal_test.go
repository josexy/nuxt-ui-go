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

func TestNativeTransitionVsPoseScheduling(t *testing.T) {
	for _, pose := range []bool{false, true} {
		t.Run(fmt.Sprintf("Pose=%v", pose), func(t *testing.T) {
			show, started, builds := false, false, 0
			tt := ui.NewTester(func(c *ui.Context) {
				builds++
				ui.Column(c.Key("root")).Fill().Children(func() {
					if !show {
						return
					}
					e := ui.Box(c.Key("panel")).Size(100, 60).Background(ui.RGB(200, 0, 0))
					if !pose {
						e.Transition(ui.ElementTransition{Duration: 150 * time.Millisecond, Enter: &ui.Motion{Scale: .95}})
						return
					}
					p := ui.Pose{}
					if !started {
						started = true
						p = ui.Pose{Fade: 1, Scale: .95}
						c.Invalidate()
					}
					e.AnimatePose(p, ui.Animation{Duration: 150 * time.Millisecond})
				})
			}, 500, 400)
			before := builds
			show = true
			tt.Frame()
			inputBuilds := builds - before
			before = builds
			tt.Advance(300 * time.Millisecond)
			ticks := builds - before
			t.Logf("group 5.4: input builds=%d, subsequent builds=%d, simulated duration=300ms; paint count unavailable", inputBuilds, ticks)
			if pose && ticks != 0 || !pose && ticks == 0 {
				t.Fatalf("unexpected animation scheduling: pose=%v tick builds=%d; update docs/performance.md", pose, ticks)
			}
			if tt.FramePending() {
				t.Fatal("animation did not settle")
			}
		})
	}
}

// motionCase is a theme on a desktop, and what each helper resolves for it.
type motionCase struct {
	name    string
	theme   func(Theme) Theme
	desktop bool // the desktop asks for reduced motion
	level   motionLevel
}

func motionCases() []motionCase {
	return []motionCase{
		{name: "default", theme: func(t Theme) Theme { return t }, level: motionFull},
		{name: "theme reduce", theme: func(t Theme) Theme { t.Motion.Reduce = true; return t }, level: motionReduced},
		{name: "desktop reduce", theme: func(t Theme) Theme { return t }, desktop: true, level: motionReduced},
		{name: "disabled", theme: func(t Theme) Theme { t.Motion.Disabled = true; return t }, level: motionOff},
		{name: "disabled and reduce", theme: func(t Theme) Theme { t.Motion.Disabled, t.Motion.Reduce = true, true; return t }, desktop: true, level: motionOff},
	}
}

// withMotion runs fn in a build of a Tester whose desktop asks for reduced
// motion when desktop is set.
func withMotion(t *testing.T, desktop bool, fn func(c *ui.Context)) {
	t.Helper()
	ran := false
	ui.NewTesterWith(func(c *ui.Context) {
		if !ran {
			ran = true
			fn(c)
		}
	}, 100, 100, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: desktop}})
	if !ran {
		t.Fatal("the view did not build")
	}
}

func TestMotionLevelHelpers(t *testing.T) {
	for _, tc := range motionCases() {
		t.Run(tc.name, func(t *testing.T) {
			withMotion(t, tc.desktop, func(c *ui.Context) {
				th := tc.theme(Light())
				if got := th.motionLevel(c); got != tc.level {
					t.Fatalf("motionLevel = %d, want %d", got, tc.level)
				}
				if got, want := th.moves(c), tc.level == motionFull; got != want {
					t.Fatalf("moves = %v, want %v", got, want)
				}
				if got, want := th.fades(c), tc.level != motionOff; got != want {
					t.Fatalf("fades = %v, want %v", got, want)
				}
			})
		})
	}
}

func TestColorTransitionHelper(t *testing.T) {
	for _, tc := range motionCases() {
		for _, d := range []time.Duration{150 * time.Millisecond, 0, -time.Millisecond} {
			t.Run(tc.name+"/"+d.String(), func(t *testing.T) {
				withMotion(t, tc.desktop, func(c *ui.Context) {
					th := tc.theme(Light())
					th.Motion.Transition = d
					tr, ok := th.colorTransition(c)
					if want := d > 0 && tc.level != motionOff; ok != want {
						t.Fatalf("colorTransition ok = %v, want %v", ok, want)
					}
					if !ok {
						return
					}
					if !tr.Colors || !tr.Shadows || tr.Duration != d || tr.Position || tr.Size || tr.Width || tr.Height || tr.Enter != nil || tr.Exit != nil {
						t.Fatalf("colorTransition = %+v, want colours and shadows over %v", tr, d)
					}
				})
			})
		}
	}
}

func TestOverlayTransitionHelper(t *testing.T) {
	enter := &ui.Motion{Y: -8, X: 4, XPercent: 10, YPercent: 20, Scale: .95, Rotate: 3, Collapse: true, Blur: 2}
	exit := &ui.Motion{Y: 8, Scale: .95, Opacity: .5}
	for _, tc := range motionCases() {
		t.Run(tc.name, func(t *testing.T) {
			withMotion(t, tc.desktop, func(c *ui.Context) {
				th := tc.theme(Light())
				tr := th.overlayTransition(c, enter, exit, 200*time.Millisecond, 100*time.Millisecond)
				switch tc.level {
				case motionOff:
					if tr.Duration >= 0 || tr.ExitDuration >= 0 {
						t.Fatalf("at Off the transition takes time: %+v", tr)
					}
				case motionReduced:
					if tr.Duration != 200*time.Millisecond || tr.ExitDuration != 100*time.Millisecond {
						t.Fatalf("at Reduced the durations changed: %+v", tr)
					}
					if *tr.Enter != (ui.Motion{Blur: 2}) || *tr.Exit != (ui.Motion{Opacity: .5}) {
						t.Fatalf("at Reduced the motions still move: enter %+v exit %+v", *tr.Enter, *tr.Exit)
					}
				default:
					if tr.Enter != enter || tr.Exit != exit || tr.Duration != 200*time.Millisecond || tr.ExitDuration != 100*time.Millisecond {
						t.Fatalf("at Full the transition changed: %+v", tr)
					}
				}
				if none := th.overlayTransition(c, nil, exit, 0, 100*time.Millisecond); none.Enter != nil {
					t.Fatalf("a nil enter became %+v", *none.Enter)
				}
			})
		})
	}
}

func TestDisableAnimationsCoversBothModes(t *testing.T) {
	for _, mode := range []Mode{SystemMode, LightMode, DarkMode} {
		for _, systemDark := range []bool{false, true} {
			a := AppConfig{Mode: mode, DisableAnimations: true}
			if th := a.Resolve(systemDark, ui.Preferences{}); !th.Motion.Disabled {
				t.Fatalf("mode %d systemDark=%v: resolved theme animates", mode, systemDark)
			}
		}
	}
	if th := (AppConfig{}).Resolve(false, ui.Preferences{}); th.Motion.Disabled || th.Motion.Reduce {
		t.Fatalf("a zero AppConfig does not animate fully: %+v", th.Motion)
	}
	// A theme's own Disabled stays without the app-wide switch.
	dark := Dark()
	dark.Motion.Disabled = true
	if th := (AppConfig{Mode: DarkMode, Light: Light(), Dark: dark}).Resolve(true, ui.Preferences{}); !th.Motion.Disabled {
		t.Fatal("the dark theme's Disabled was dropped")
	}
}
