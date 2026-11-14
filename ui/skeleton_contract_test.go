package nui_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSkeletonStyleRadiusOverridesUI(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).AlignItems(ui.Start).Children(func() {
			nui.Skeleton(c, th, nui.SkeletonOptions{Width: nui.Set(float32(40)), Height: nui.Set(float32(40)), Animate: nui.Set(false), UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(20))}}, Style: func(_ nui.Slot, e ui.Element) { e.Radius(0) }})
		})
	}, 100, 100)
	r := bounds(t, tt, "Loading…")
	if got := tt.Image().RGBAAt(int(r.X)+2, int(r.Y)+2); got != rgba(nui.Light().Elevated) {
		t.Fatalf("Style radius did not override UI radius: corner=%v", got)
	}
}

func TestSkeletonExplicitZeroAndTransparentBackground(t *testing.T) {
	w, h := float32(0), float32(16)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).AlignSelf(ui.Start).AlignItems(ui.Start).Label("Skeleton layout").Children(func() {
			nui.Skeleton(c, th, nui.SkeletonOptions{Width: nui.Set(w), Height: nui.Set(h), Animate: nui.Set(false), UI: nui.Styles{nui.Root: {Background: nui.Set(ui.Transparent)}}})
			ui.Box(c).Size(4, 4).Label("Following marker")
		})
	}, 200, 100)
	// Empty painted rectangles have no hit bounds; inspect their layout contribution.
	if r := bounds(t, tt, "Skeleton layout"); r.W != 4 || r.H != 20 {
		t.Fatalf("explicit zero width was replaced: %+v", r)
	}
	w, h = 80, 0
	tt.Frame()
	if r := bounds(t, tt, "Skeleton layout"); r.W != 80 || r.H != 4 {
		t.Fatalf("explicit zero height was replaced: %+v", r)
	}
	h = 20
	tt.Frame()
	r := bounds(t, tt, "Loading…")
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)); got != rgba(nui.Light().Background) {
		t.Fatalf("transparent background was not preserved: %v", got)
	}
}

func TestSkeletonPulseChangesPixelsAndStops(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Skeleton(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode}), nui.SkeletonOptions{})
	}, 200, 60)
	first := bytes.Clone(tt.Image().Pix)
	changed := false
	for range 4 {
		tt.Advance(250 * time.Millisecond)
		changed = changed || !bytes.Equal(first, tt.Image().Pix)
	}
	if !changed {
		t.Fatal("pulse requested frames without changing pixels")
	}
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	still := bytes.Clone(tt.Image().Pix)
	tt.Advance(time.Second)
	if !bytes.Equal(still, tt.Image().Pix) || tt.FramePending() {
		t.Fatal("reduced motion did not stop pulse")
	}
}

func TestSkeletonThemesAndParentWidth(t *testing.T) {
	for name, mode := range map[string]nui.Mode{"light": nui.LightMode, "dark": nui.DarkMode} {
		t.Run(name, func(t *testing.T) {
			var th nui.Theme
			tt := ui.NewTester(func(c *ui.Context) {
				th = nui.App(c, nui.AppConfig{Mode: mode})
				ui.Column(c).Width(100).Padding(8).Children(func() {
					nui.Skeleton(c, th, nui.SkeletonOptions{Animate: nui.Set(false)}).FillWidth()
				})
			}, 180, 80)
			tt.SetPreferences(ui.Preferences{TextScale: 1.5})
			r := bounds(t, tt, "Loading…")
			if r.W != 84 || r.H != 24 {
				t.Fatalf("parent width or text scaling: %+v", r)
			}
			if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)); got != rgba(th.Elevated) {
				t.Fatalf("theme background=%v want=%v", got, th.Elevated)
			}
		})
	}
}
