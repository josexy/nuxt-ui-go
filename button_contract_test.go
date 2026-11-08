package nui_test

import (
	"bytes"
	"image/color"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestButtonLoadingIconAnimationChangesPixels(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c, th, "Loading", nui.ButtonOptions{Loading: true})
	}, 240, 100)
	before := slices.Clone(tt.Image().Pix)
	tt.Advance(125 * time.Millisecond)
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("loading icon scheduled frames without rotating")
	}
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	before = slices.Clone(tt.Image().Pix)
	tt.Advance(125 * time.Millisecond)
	if !bytes.Equal(before, tt.Image().Pix) || tt.FramePending() {
		t.Fatal("reduced motion did not stop the loading icon")
	}
}

func TestButtonRestoringMotionKeepsCurrentColor(t *testing.T) {
	color, reduce := nui.Primary, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = reduce
		ui.Column(c).Padding(16).Children(func() { nui.Button(c, th, "Motion button", nui.ButtonOptions{Color: color}).Width(160) })
	}, 220, 90)
	color, reduce = nui.Error, true
	tt.Frame()
	r := bounds(t, tt, "Motion button")
	before := tt.Image().RGBAAt(int(r.X)+2, int(r.Y+r.H/2))
	if before != rgba(nui.Light().Error) {
		t.Fatal("reduced motion did not show the current button color immediately")
	}
	reduce = false
	tt.Frame()
	if after := tt.Image().RGBAAt(int(r.X)+2, int(r.Y+r.H/2)); after != before {
		t.Fatalf("restoring motion reverted the current button color: %v -> %v", before, after)
	}
}

func TestButtonLoadingIconUsesPartForeground(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Button(c, th, "Loading", nui.ButtonOptions{Loading: true, Animate: nui.Set(false), UI: nui.Styles{nui.Leading: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(root, label, leading, trailing ui.Element) {
			leading.Label("Loading icon")
		}})
	}, 240, 100)
	r := bounds(t, tt, "Loading icon")
	if countWhere(tt.Image(), r, func(c color.RGBA) bool { return int(c.R) > int(c.G)+20 && int(c.B) > int(c.G)+20 }) == 0 {
		t.Fatal("part foreground was ignored by the loading icon")
	}
}

func TestButtonAvatarDimensionsAndPriority(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		avatar := nui.AvatarOptions{Text: "AL", Alt: "Ada Lovelace"}
		opts := nui.ButtonOptions{Size: size, Avatar: &avatar, Animate: nui.Set(false)}
		calls := 0
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Button(c.Key("account"), th, "Account", opts).OnClick(func() { calls++ })
		}, 320, 120)
		want := []float32{16, 16, 20, 20, 24}[i]
		r := bounds(t, tt, "Ada Lovelace")
		if r.W != want || r.H != want {
			t.Fatalf("size=%v avatar=%+v, want %g", size, r, want)
		}
		click(t, tt, "Account")
		if calls != 1 {
			t.Fatal("avatar button did not activate once")
		}
		opts.Leading = lucide.Check()
		tt.Frame()
		if tt.HasText("Ada Lovelace") || tt.HasText("AL") || calls != 1 || !tt.Focused("Account") {
			t.Fatal("leading icon did not replace avatar while preserving state")
		}
		opts.Leading = nil
		opts.Loading = true
		tt.Frame()
		if tt.HasText("Ada Lovelace") {
			t.Fatal("leading loader did not replace avatar")
		}
		click(t, tt, "Account")
		tt.Key(0, ui.KeySpace)
		if calls != 1 {
			t.Fatal("loading avatar button activated")
		}
		opts.LoadingTrailing = true
		tt.Frame()
		if !tt.HasText("Ada Lovelace") {
			t.Fatal("trailing loader hid the leading avatar")
		}
		opts.Loading = false
		avatar.Size = nui.XS
		tt.Frame()
		if r = bounds(t, tt, "Ada Lovelace"); r.W != 24 || r.H != 24 {
			t.Fatalf("explicit avatar size was overridden: %+v", r)
		}
		if avatar.Size != nui.XS || avatar.Text != "AL" {
			t.Fatal("button mutated avatar options")
		}
	}
}

func TestButtonAvatarOnlyRTLAndPartStyles(t *testing.T) {
	avatar := nui.AvatarOptions{Icon: lucide.User(), Alt: "Account image"}
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Gap(12).Children(func() {
			nui.Button(c.Key("only"), th, "", nui.ButtonOptions{Avatar: &avatar, AccessibleLabel: "Account", Style: func(root, label, leading, trailing ui.Element) { leading.Label("Styled avatar") }}).OnClick(func() { calls++ })
			nui.Button(c.Key("both"), th, "Profile", nui.ButtonOptions{Avatar: &avatar, Trailing: lucide.ChevronDown(), Style: func(root, label, leading, trailing ui.Element) {
				leading.Label("Leading avatar")
				trailing.Label("Trailing icon")
			}})
		})
	}, 400, 120)
	if r := bounds(t, tt, "Account"); r.W != r.H {
		t.Fatalf("avatar-only button not square: %+v", r)
	}
	if bounds(t, tt, "Leading avatar").X >= bounds(t, tt, "Trailing icon").X {
		t.Fatal("LTR avatar order")
	}
	tt.SetDirection(ui.RightToLeft)
	if bounds(t, tt, "Leading avatar").X <= bounds(t, tt, "Trailing icon").X {
		t.Fatal("RTL avatar order")
	}
	click(t, tt, "Account")
	tt.Key(0, ui.KeySpace)
	if calls != 2 || !tt.Focused("Account") {
		t.Fatal("avatar-only mouse/Space activation or focus")
	}
}
