package nui_test

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

// Shared recipes aligned with Nuxt UI v4.12: overlay surfaces and motion.

func TestPopoverPaddingAndScaleIn(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			anchor := nui.Button(c, th, "Details", nui.ButtonOptions{}).OnClick(func() { open = !open })
			nui.Popover(c, th, anchor, &open, nui.PopoverOptions{Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Panel {
					e.Label("Popover panel")
				}
			}}}, func(nui.Theme) { ui.Text(c, "Inside") })
		})
	}, 400, 300)
	click(t, tt, "Details")
	opening := bounds(t, tt, "Popover panel")
	tt.Advance(50 * time.Millisecond)
	middle := bounds(t, tt, "Popover panel")
	tt.Advance(time.Second)
	settled := bounds(t, tt, "Popover panel")
	if !(opening.W < middle.W && middle.W < settled.W) {
		t.Fatalf("popover did not scale in: widths %v, %v, %v", opening.W, middle.W, settled.W)
	}
	// Scaling from the anchor side keeps the edge facing the anchor in place.
	if opening.Y != settled.Y || opening.X != settled.X {
		t.Fatalf("scale-in moved the anchored corner: %v -> %v", opening, settled)
	}
	if tt.FramePending() {
		t.Fatal("settled popover keeps requesting frames")
	}
	inside := bounds(t, tt, "Inside")
	if inside.X-settled.X != 16 || inside.Y-settled.Y != 16 {
		t.Fatalf("popover content offset (%v, %v), want 16 DIP padding", inside.X-settled.X, inside.Y-settled.Y)
	}

	click(t, tt, "Details")
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	click(t, tt, "Details")
	if got := bounds(t, tt, "Popover panel"); got != settled {
		t.Fatalf("reduced motion still animates: %v, want %v", got, settled)
	}
}

func TestTooltipIsOneTruncatedLine(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(150).Top(100)
		nui.Tooltip(c, th, a, "A rather long explanation that cannot fit on one tooltip line", nui.TooltipOptions{Open: &open, MaxWidth: nui.Set(float32(160)), Decoration: nui.Decoration{Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Panel {
				e.Label("Tip")
			}
		}}})
	}, 500, 300)
	tt.Advance(time.Second)
	tip := bounds(t, tt, "Tip")
	if tip.H != 24 || tip.W > 160 {
		t.Fatalf("tooltip box %v, want one 24 DIP line within its 160 DIP maximum", tip)
	}
	if text := bounds(t, tt, "A rather long explanation that cannot fit on one tooltip line"); text.H != 16 || text.X+text.W > tip.X+tip.W-10 {
		t.Fatalf("tooltip text %v in %v is not one truncated line inside the 10 DIP padding", text, tip)
	}
}

func TestModalScalesInOverFadingBackdrop(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Confirm"}, func(nui.Theme) { ui.Text(c, "Body") })
	}, 600, 400)
	// The backdrop is Elevated at 75% over white; it starts transparent.
	corner := func() uint8 { return tt.Image().RGBAAt(2, 2).G }
	first, panel := corner(), bounds(t, tt, "Confirm")
	tt.Advance(time.Second)
	settled := bounds(t, tt, "Confirm")
	if first <= corner() {
		t.Fatalf("backdrop did not fade in: corner green %d then %d", first, corner())
	}
	if panel.W >= settled.W || settled.W != 512 {
		t.Fatalf("modal did not scale in to 512 DIPs: %v -> %v", panel, settled)
	}
}
