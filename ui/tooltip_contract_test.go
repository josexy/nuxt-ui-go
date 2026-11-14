package nui_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestTooltipConfiguredDelayAndSuppression(t *testing.T) {
	for _, delay := range []time.Duration{0, 150 * time.Millisecond, time.Second} {
		t.Run(delay.String(), func(t *testing.T) {
			open := false
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				a := nui.Button(c.Key("anchor"), th, "Anchor", nui.ButtonOptions{}).Absolute().Left(80).Top(60)
				nui.Tooltip(c.Key("tip"), th, a, "Delayed help", nui.TooltipOptions{Open: &open, DelayDuration: nui.Set(delay)})
			}, 500, 300)
			r := bounds(t, tt, "Anchor")
			tt.Move(r.X+4, r.Y+4)
			if delay > 0 {
				tt.Advance(delay - time.Millisecond)
				if open || tt.HasText("Delayed help") {
					t.Fatal("tooltip opened before configured delay")
				}
				tt.Advance(time.Millisecond)
				tt.Frame()
			}
			if !open || !tt.HasText("Delayed help") {
				t.Fatal("tooltip did not open at configured delay")
			}
			tt.Key(0, ui.KeyEscape)
			tt.Advance(2 * time.Second)
			if open || tt.HasText("Delayed help") {
				t.Fatal("Escape did not suppress the current hover")
			}
			tt.Move(490, 290)
			tt.Key(0, ui.KeyTab)
			if !open || !tt.HasText("Delayed help") {
				t.Fatal("keyboard focus waited for hover delay")
			}
			open = false
			tt.Frame()
			tt.Advance(2 * time.Second)
			if open {
				t.Fatal("external false reopened while focused")
			}
		})
	}
}

func TestTooltipKeysAndPanelGap(t *testing.T) {
	open := true
	gap := float32(4)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(150).Top(100)
		nui.Tooltip(c, th, a, "Save", nui.TooltipOptions{Open: &open, Kbds: []string{"A", "B"}, UI: nui.Styles{nui.Panel: {Gap: nui.Set(gap)}}})
	}, 500, 300)
	tt.Advance(time.Second)
	a, b := bounds(t, tt, "A"), bounds(t, tt, "B")
	// Account for centered glyphs in keys with a 16 DIP minimum width.
	want := (max(float32(16), a.W+8)+max(float32(16), b.W+8)+a.W-b.W)/2 + 2
	if actual := b.X - a.X; math.Abs(float64(actual-want)) > .1 {
		t.Fatalf("key pitch=%v, want %v", actual, want)
	}
	text, dot := bounds(t, tt, "Save"), bounds(t, tt, "·")
	before := dot.X - text.X - text.W
	gap = 20
	tt.Frame()
	text, dot = bounds(t, tt, "Save"), bounds(t, tt, "·")
	if after := dot.X - text.X - text.W; after-before != 16 {
		t.Fatalf("Panel.Gap did not change text/key spacing: %v -> %v", before, after)
	}
}

func TestTooltipCustomDelayCancellationDisabledAndWarmSwitch(t *testing.T) {
	first, second, disabled := false, false, false
	delay := nui.Set(200 * time.Millisecond)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(40).Gap(20).Children(func() {
			a := nui.Button(c.Key("first"), th, "First anchor", nui.ButtonOptions{})
			nui.Tooltip(c.Key("first-tip"), th, a, "First help", nui.TooltipOptions{Open: &first, DelayDuration: delay, Disabled: disabled})
			b := nui.Button(c.Key("second"), th, "Second anchor", nui.ButtonOptions{})
			nui.Tooltip(c.Key("second-tip"), th, b, "Second help", nui.TooltipOptions{Open: &second})
		})
	}, 500, 300)
	a, b := bounds(t, tt, "First anchor"), bounds(t, tt, "Second anchor")
	tt.Move(a.X+5, a.Y+5)
	tt.Advance(100 * time.Millisecond)
	tt.Move(490, 290)
	tt.Advance(time.Second)
	if first {
		t.Fatal("leaving during delay did not cancel opening")
	}
	tt.Move(a.X+5, a.Y+5)
	tt.Advance(210 * time.Millisecond)
	if !first {
		t.Fatal("re-enter did not restart the delay")
	}
	tt.Advance(time.Second)
	tt.Move(b.X+5, b.Y+5)
	tt.Frame()
	if first || !second || tt.HasText("First help") || !tt.HasText("Second help") {
		t.Fatal("custom-to-default warm switch did not transfer ownership immediately")
	}
	tt.Move(a.X+5, a.Y+5)
	tt.Frame()
	if !first || second {
		t.Fatal("default-to-custom warm switch did not transfer ownership")
	}
	disabled = true
	tt.Frame()
	if first || tt.HasText("First help") {
		t.Fatal("disabled tooltip stayed open")
	}
	disabled = false
	delay = nui.Set(time.Duration(0))
	tt.Frame()
	if !first {
		t.Fatal("enabled zero-delay tooltip did not open")
	}
	click(t, tt, "First anchor")
	tt.Advance(time.Second)
	if first {
		t.Fatal("trigger click reopened the custom tooltip")
	}
}

func TestTooltipLongTextKeysLayoutAndDisable(t *testing.T) {
	open, disabled := true, false
	for _, rtl := range []bool{false, true} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("rtl=%v/dark=%v", rtl, dark), func(t *testing.T) {
				open = true
				tt := ui.NewTester(func(c *ui.Context) {
					mode := nui.LightMode
					if dark {
						mode = nui.DarkMode
					}
					th := nui.App(c, nui.AppConfig{Mode: mode})
					th.Motion.Reduce = true
					if rtl {
						c.Root().Direction(ui.RightToLeft)
					}
					a := nui.Button(c, th, "Anchor", nui.ButtonOptions{Disabled: disabled}).Absolute().Left(280).Top(155).Size(100, 32)
					nui.Tooltip(c, th, a, "这是需要截断的很长中文提示文字", nui.TooltipOptions{Open: &open, MaxWidth: nui.Set(float32(150)), Placement: nui.Set(nui.BottomStart), Kbds: []string{"Ctrl", "S"}, Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Panel {
							e.Label("Tip panel")
						}
					}})
				}, 400, 210)
				r := bounds(t, tt, "Tip panel")
				if r.W > 150 || r.X < 0 || r.X+r.W > 400 || r.Y+r.H > 210 {
					t.Fatalf("tooltip outside window: %v", r)
				}
				if !tt.HasText("S") || !tt.HasText("·") {
					t.Fatal("narrow tooltip lost shortcut keys")
				}
				disabled = true
				tt.Frame()
				if open || tt.HasText("Tip panel") {
					t.Fatal("disabled anchor did not close tooltip")
				}
				disabled = false
			})
		}
	}
}

func TestTooltipExternalOpenDismissesOnOutsidePress(t *testing.T) {
	open, clicks := false, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(30).Gap(24).Children(func() {
			a := nui.Button(c.Key("anchor"), th, "Anchor", nui.ButtonOptions{})
			nui.Button(c.Key("open"), th, "Open help", nui.ButtonOptions{}).OnClick(func() { open = true })
			nui.Button(c.Key("outside"), th, "Outside action", nui.ButtonOptions{}).OnClick(func() { clicks++ })
			nui.Tooltip(c.Key("tip"), th, a, "Manual help", nui.TooltipOptions{Open: &open})
		})
	}, 500, 300)
	click(t, tt, "Open help")
	if !open || !tt.HasText("Manual help") {
		t.Fatal("opening press immediately dismissed the tip")
	}
	click(t, tt, "Outside action")
	if open || tt.HasText("Manual help") || clicks != 1 {
		t.Fatalf("outside press: open=%v clicks=%d", open, clicks)
	}
}

func TestTooltipCustomDelayPrefersInnerAnchor(t *testing.T) {
	innerOpen, outerOpen := false, false
	showInnerTip := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		outer := ui.Column(c.Key("outer")).Width(300).Height(160).Padding(30)
		outer.Children(func() {
			inner := nui.Button(c.Key("inner"), th, "Inner anchor", nui.ButtonOptions{})
			if showInnerTip {
				nui.Tooltip(c.Key("inner-tip"), th, inner, "Inner help", nui.TooltipOptions{Open: &innerOpen, DelayDuration: nui.Set(time.Duration(0))})
			}
		})
		nui.Tooltip(c.Key("outer-tip"), th, outer, "Outer help", nui.TooltipOptions{Open: &outerOpen, DelayDuration: nui.Set(time.Duration(0))})
	}, 500, 300)
	r := bounds(t, tt, "Inner anchor")
	tt.Move(r.X+5, r.Y+5)
	tt.Advance(time.Millisecond)
	if !innerOpen || outerOpen || !tt.HasText("Inner help") || tt.HasText("Outer help") {
		t.Fatal("outer tooltip replaced the hovered inner tooltip")
	}
	tt.Move(r.X-15, r.Y+5)
	tt.Advance(time.Millisecond)
	if innerOpen || !outerOpen {
		t.Fatal("leaving the inner trigger did not restore outer hover")
	}
	tt.Move(r.X+5, r.Y+5)
	tt.Advance(time.Millisecond)
	if !innerOpen || outerOpen {
		t.Fatal("re-entering inner trigger did not transfer ownership")
	}
	showInnerTip = false
	tt.Frame()
	tt.Advance(time.Millisecond)
	if innerOpen || !outerOpen {
		t.Fatalf("removed inner tooltip still owned the hover: inner=%v outer=%v texts=%v", innerOpen, outerOpen, tt.Texts())
	}
}
