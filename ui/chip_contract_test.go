package nui_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestChipCornersInsetAndDirection(t *testing.T) {
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		for _, pos := range []nui.ChipPosition{nui.TopLeft, nui.TopRight, nui.BottomLeft, nui.BottomRight} {
			for _, inset := range []bool{false, true} {
				t.Run(fmt.Sprintf("%v/%v/%v", dir, pos, inset), func(t *testing.T) {
					tt := ui.NewTester(func(c *ui.Context) {
						th := nui.App(c, nui.AppConfig{})
						ui.Column(c).Padding(32).AlignItems(ui.Start).Children(func() {
							nui.Chip(c, th, nui.ChipOptions{Position: pos, Inset: inset, Style: func(slot nui.Slot, e ui.Element) {
								if slot == nui.Indicator {
									e.Label("Marker")
								}
							}}, func() { ui.Box(c).Size(80, 40).Label("Host") })
						})
					}, 180, 120)
					tt.SetDirection(dir)
					host, marker := bounds(t, tt, "Host"), bounds(t, tt, "Marker")
					x, y := host.X+host.W-4, host.Y-4
					if pos == nui.TopLeft || pos == nui.BottomLeft {
						x = host.X - 4
					}
					if pos == nui.BottomLeft || pos == nui.BottomRight {
						y = host.Y + host.H - 4
					}
					if inset {
						if pos == nui.TopLeft || pos == nui.BottomLeft {
							x += 4
						} else {
							x -= 4
						}
						if pos == nui.TopLeft || pos == nui.TopRight {
							y += 4
						} else {
							y -= 4
						}
					}
					if marker.X != x || marker.Y != y || marker.W != 8 || marker.H != 8 {
						t.Fatalf("position mismatch: marker=%+v want=(%g,%g,8,8)", marker, x, y)
					}
				})
			}
		}
	}
}

func TestChipTextWidthIncludesIndicatorStyles(t *testing.T) {
	var want float32
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		want, _ = c.MeasureText(1000, ui.Span{Text: "99+", Font: th.MonoFont, Size: 24, Weight: 700})
		nui.Chip(c, th, nui.ChipOptions{Text: "99+", UI: nui.Styles{nui.Indicator: {Font: nui.Set(th.MonoFont), FontSize: nui.Set(float32(24)), Weight: nui.Set(700), PaddingX: nui.Set(float32(4))}}}, nil).AlignSelf(ui.Start).Label("Standalone")
	}, 180, 100)
	if got := bounds(t, tt, "Standalone").W; math.Abs(float64(got-want-8)) > .1 {
		t.Fatalf("styled text width=%g, want glyphs %g + padding 8", got, want)
	}
}

func TestChipHiddenKeepsHostFocusAndClick(t *testing.T) {
	hidden, calls := false, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).AlignItems(ui.Start).Children(func() {
			nui.Chip(c.Key("chip"), th, nui.ChipOptions{Text: "3", Hidden: hidden}, func() {
				nui.Button(c.Key("host"), th, "Host", nui.ButtonOptions{}).OnClick(func() { calls++ })
			})
		})
	}, 180, 100)
	r := bounds(t, tt, "Host")
	tt.ClickAt(r.X+r.W-1, r.Y+1)
	if calls != 1 {
		t.Fatal("chip intercepted host activation")
	}
	hidden = true
	tt.Frame()
	if !tt.Focused("Host") || bounds(t, tt, "Host") != r || tt.HasText("3") {
		t.Fatal("hiding chip altered host layout, focus or indicator visibility")
	}
	tt.Key(0, ui.KeyEnter)
	if calls != 2 {
		t.Fatal("host did not remain keyboard operable")
	}
}

func TestChipNeutralUsesInvertedBackgroundToken(t *testing.T) {
	th := nui.Light()
	th.InvertedBackground = ui.Hex("#e879f9")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, Light: th})
		nui.Chip(c, th, nui.ChipOptions{Color: nui.Neutral}, nil).AlignSelf(ui.Start).Label("Chip")
	}, 100, 50)
	r := bounds(t, tt, "Chip")
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)); got != rgba(th.InvertedBackground) {
		t.Fatalf("neutral background=%v, want customized inverted token %v", got, rgba(th.InvertedBackground))
	}
}
