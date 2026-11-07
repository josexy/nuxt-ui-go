package nui_test

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestKbdExplicitPlatformSymbols(t *testing.T) {
	for _, platform := range []string{"windows", "linux", "darwin"} {
		for key, want := range map[string]string{"command": "⌘", "win": "⊞", "COMMAND": "⌘", "WIN": "⊞", "F12": "F12", "Custom key": "Custom key"} {
			if got := nui.FormatKey(key, platform); got != want {
				t.Errorf("FormatKey(%q, %q)=%q, want %q", key, platform, got, want)
			}
		}
	}
}

func TestKbdCombinationAndNoninteractiveBehavior(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Gap(8).Children(func() {
			nui.Button(c.Key("before"), th, "Before", nui.ButtonOptions{})
			nui.Kbd(c.Key("keys"), th, []string{"command", "win", "Ctrl", "k", "custom"}, nui.KbdOptions{Platform: "windows"})
			nui.Button(c.Key("after"), th, "After", nui.ButtonOptions{})
		})
	}, 400, 160)
	for _, text := range []string{"⌘", "⊞", "CTRL", "K", "CUSTOM"} {
		if !tt.HasText(text) {
			t.Errorf("key label %q missing", text)
		}
	}
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("After") {
		t.Fatal("Kbd entered the tab order")
	}
}

func TestNeutralSolidUsesInvertedBackground(t *testing.T) {
	for _, highContrast := range []bool{false, true} {
		th := nui.Light()
		th.InvertedBackground = ui.Hex("#7e22ce")
		th.HighContrast = highContrast
		th.Motion.Reduce = true
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, Light: th})
			ui.Column(c).Padding(16).Gap(16).Children(func() {
				nui.Button(c, th, "Button", nui.ButtonOptions{Color: nui.Neutral, Variant: nui.Solid})
				nui.Badge(c, th, "Badge", nui.BadgeOptions{Color: nui.Neutral, Variant: nui.Solid}).Label("Badge")
				nui.Alert(c, th, nui.AlertOptions{Title: "Alert", Color: nui.Neutral, Variant: nui.Solid})
				nui.Kbd(c, th, []string{"K"}, nui.KbdOptions{Color: nui.Neutral, Variant: nui.Solid}).Label("Kbd")
			})
		}, 300, 260)
		for _, label := range []string{"Button", "Badge", "Alert", "Kbd"} {
			r := bounds(t, tt, label)
			if got := tt.Image().RGBAAt(int(r.X+2), int(r.Y+r.H/2)); got != rgba(th.InvertedBackground) {
				t.Errorf("highContrast=%v %s background=%v, want %v", highContrast, label, got, rgba(th.InvertedBackground))
			}
		}
		r := bounds(t, tt, "Button")
		tt.Move(r.X+2, r.Y+r.H/2)
		tt.Advance(time.Second)
		got := tt.Image().RGBAAt(int(r.X+2), int(r.Y+r.H/2))
		if highContrast {
			if got != rgba(th.InvertedBackground) {
				t.Fatal("high-contrast hover lost the opaque inverted background")
			}
		} else if got.R < 125 || got.R > 141 || got.G < 34 || got.G > 58 || got.B < 206 || got.B > 213 {
			t.Fatalf("hover did not use inverted background at 90%% opacity: %v", got)
		}
	}
}
