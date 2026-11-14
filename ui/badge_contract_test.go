package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestBadgeIconOnlyUsesAutomaticSquarePadding(t *testing.T) {
	for _, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Row(c).Children(func() { nui.Badge(c, th, "", nui.BadgeOptions{Size: size, Leading: lucide.Check()}).Label("Badge") })
		}, 200, 100)
		r := bounds(t, tt, "Badge")
		if r.W != r.H {
			t.Fatalf("size %v icon-only badge=%+v", size, r)
		}
	}
}

func TestBadgeAvatarDimensionsAndPriority(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		avatar := nui.AvatarOptions{Alt: "Ada Lovelace"}
		opts := nui.BadgeOptions{Size: size, Avatar: &avatar}
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Row(c).Children(func() { nui.Badge(c.Key("badge"), th, "Online", opts).Label("Badge") })
		}, 260, 100)
		want := []float32{16, 16, 16, 20, 20}[i]
		if r := bounds(t, tt, "Ada Lovelace"); r.W != want || r.H != want {
			t.Fatalf("size=%v avatar=%+v, want %g", size, r, want)
		}
		opts.Leading = lucide.Check()
		tt.Frame()
		if tt.HasText("Ada Lovelace") {
			t.Fatal("leading icon did not replace badge avatar")
		}
		opts.Leading = nil
		avatar.Size = nui.LG
		tt.Frame()
		if r := bounds(t, tt, "Ada Lovelace"); r.W != 36 || r.H != 36 {
			t.Fatal("explicit avatar size ignored")
		}
		tt.Key(0, ui.KeyTab)
		if tt.Focused("Badge") {
			t.Fatal("display badge became keyboard-interactive")
		}
	}
}

func TestBadgeSquareKeepsInstancePadding(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Children(func() {
			nui.Badge(c, th, "", nui.BadgeOptions{Leading: lucide.Check(), UI: nui.Styles{nui.Root: {PaddingX: nui.Set(float32(20))}}}).Label("Badge")
		})
	}, 200, 100)
	if r := bounds(t, tt, "Badge"); r.W != 56 {
		t.Fatalf("automatic square overrode explicit padding: %+v", r)
	}
}
