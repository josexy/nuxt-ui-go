package nui_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestProgressNeutralUsesInvertedBackground(t *testing.T) {
	color := ui.Hex("#7e22ce")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.InvertedBackground = color
		nui.Progress(c, th, 50, nui.ProgressOptions{Label: "Progress", Color: nui.Neutral, Animate: nui.Set(false)}).Width(200)
	}, 240, 60)
	r := bounds(t, tt, "Progress")
	if got := tt.Image().RGBAAt(int(r.X)+30, int(r.Y)+4); got != rgba(color) {
		t.Fatalf("Neutral progress ignored InvertedBackground: %v", got)
	}
}

func TestProgressRejectsInvalidMaximum(t *testing.T) {
	for _, maximum := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run(fmt.Sprint(maximum), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid maximum was silently accepted")
				}
			}()
			ui.NewTester(func(c *ui.Context) {
				nui.Progress(c, nui.App(c, nui.AppConfig{}), 20, nui.ProgressOptions{Max: nui.Set(maximum)})
			}, 240, 60)
		})
	}
}

func TestProgressValueTransitionSettles(t *testing.T) {
	value := float64(25)
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Progress(c, nui.App(c, nui.AppConfig{}), value, nui.ProgressOptions{Label: "Progress"}).Width(200)
	}, 240, 60)
	value = 75
	tt.Frame()
	if !tt.FramePending() {
		t.Fatal("value change did not start a transition")
	}
	tt.Advance(250 * time.Millisecond)
	if tt.FramePending() {
		t.Fatal("completed value transition kept requesting frames")
	}
	r := bounds(t, tt, "Progress")
	if got := tt.Image().RGBAAt(int(r.X)+140, int(r.Y)+4); got != rgba(nui.Light().Primary) {
		t.Fatal("value transition did not reach its target")
	}
}

func TestProgressRestoringAnimationKeepsCurrentValue(t *testing.T) {
	value, animate := float64(25), true
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Progress(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode}), value, nui.ProgressOptions{Label: "Progress", Animate: nui.Set(animate)}).Width(200)
	}, 240, 60)
	value, animate = 75, false
	tt.Frame()
	r := bounds(t, tt, "Progress")
	before := tt.Image().RGBAAt(int(r.X)+140, int(r.Y)+4)
	if before != rgba(nui.Light().Primary) {
		t.Fatal("disabled animation did not show the current value immediately")
	}
	animate = true
	tt.Frame()
	if after := tt.Image().RGBAAt(int(r.X)+140, int(r.Y)+4); after != before {
		t.Fatalf("restoring animation reverted the current progress: %v -> %v", before, after)
	}
}

func TestProgressIndeterminateHalfTrack(t *testing.T) {
	animated, vertical := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		e := nui.Progress(c, th, 0, nui.ProgressOptions{Label: "Progress", Indeterminate: true, Animate: nui.Set(animated), Vertical: vertical})
		if vertical {
			e.Height(200)
		} else {
			e.Width(200)
		}
	}, 240, 240)
	assertHalf := func() {
		t.Helper()
		r := bounds(t, tt, "Progress")
		for _, offset := range []int{60, 140} {
			x, y := int(r.X)+offset, int(r.Y)+4
			if vertical {
				x, y = int(r.X)+4, int(r.Y)+offset
			}
			if got := tt.Image().RGBAAt(x, y); got != rgba(nui.Light().Primary) {
				t.Fatalf("half-track indicator missing at %d: %v", offset, got)
			}
		}
	}
	assertHalf()
	animated = true
	tt.Frame()
	phase := time.Duration(tt.Now().UnixNano() % int64(2*time.Second))
	tt.Advance(3*time.Second - phase)
	tt.Frame()
	assertHalf()
	vertical = true
	tt.Frame()
	assertHalf()
}
