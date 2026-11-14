package nui_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestSlideoverInsetFitsEveryEdge(t *testing.T) {
	for _, side := range []nui.Side{nui.SideLeft, nui.SideRight, nui.SideTop, nui.SideBottom} {
		for _, size := range [][2]int{{640, 480}, {240, 180}} {
			t.Run(fmt.Sprintf("%d/%v", side, size), func(t *testing.T) {
				open := true
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					nui.Slideover(c, th, &open, nui.SlideoverOptions{Title: "Inset panel", Transition: nui.Set(false), Side: side, Extent: nui.Set(float32(300)), Inset: true}, func(nui.Theme) { ui.Text(c, "Body") })
				}, size[0], size[1])
				r := bounds(t, tt, "Inset panel")
				w, h := float32(size[0]), float32(size[1])
				if r.X < 16 || r.Y < 16 || r.X+r.W > w-16 || r.Y+r.H > h-16 {
					t.Fatalf("inset bounds outside 16 DIP margin: %v in %vx%v", r, w, h)
				}
				if (side == nui.SideLeft || side == nui.SideRight) && r.H != h-32 {
					t.Fatalf("side panel height=%v", r.H)
				}
				if (side == nui.SideTop || side == nui.SideBottom) && r.W != w-32 {
					t.Fatalf("top/bottom panel width=%v", r.W)
				}
			})
		}
	}
}

func TestSlideoverCustomBackdropFades(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Slideover(c, th, &open, nui.SlideoverOptions{Title: "Panel", UI: nui.Styles{nui.Backdrop: {Background: nui.Set(ui.RGBA(0, 0, 0, .5))}}, Extent: nui.Set(float32(180))}, nil)
	}, 600, 400)
	first := tt.Image().RGBAAt(4, 4).R
	tt.Advance(250 * time.Millisecond)
	last := tt.Image().RGBAAt(4, 4).R
	if first != 255 || last > 129 || last < 125 {
		t.Fatalf("custom backdrop did not fade from transparent to 50%%: %d -> %d", first, last)
	}
	open = false
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	if middle := tt.Image().RGBAAt(4, 4).R; middle <= last || middle == 255 {
		t.Fatalf("custom backdrop exit did not fade: %d", middle)
	}
}

func TestSlideoverEmptyTopKeepsBodyPadding(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Slideover(c, th, &open, nui.SlideoverOptions{Title: "Empty panel", Transition: nui.Set(false), Side: nui.SideTop}, nil)
	}, 600, 400)
	if r := bounds(t, tt, "Empty panel"); r.H != 113 {
		t.Fatalf("empty slideover=%v, want header 64 + divider 1 + body padding 48", r)
	}
}

func TestSlideoverTransitionExitReversalAndFocus(t *testing.T) {
	open, transition, block := false, true, false
	closed, prevented, actions := 0, 0, 0
	var opener ui.Handle
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c.Key("open"), th, "Open panel", nui.ButtonOptions{}).Bind(&opener).OnClick(func() { open = true })
		nui.Slideover(c.Key("panel"), th, &open, nui.SlideoverOptions{Title: "Panel", Transition: nui.Set(transition), Dismissible: nui.Set(!block), ReturnFocus: &opener, OnClose: func() { closed++ }, OnClosePrevent: func() { prevented++ }, Inset: true, Extent: nui.Set(float32(240))}, func(th nui.Theme) {
			nui.Button(c, th, "Panel action", nui.ButtonOptions{}).OnClick(func() { actions++ })
		})
	}, 600, 400)
	click(t, tt, "Open panel")
	tt.Advance(60 * time.Millisecond)
	transition = false
	tt.Frame()
	settled := bounds(t, tt, "Panel")
	if settled.X != 344 {
		t.Fatalf("disabled transition not settled: %v", settled)
	}
	transition = true
	tt.Frame()
	if r := bounds(t, tt, "Panel"); r != settled {
		t.Fatalf("re-enabled transition moved backwards: %v", r)
	}
	block = true
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.ClickAt(5, 5)
	if !open || prevented != 2 || closed != 0 {
		t.Fatalf("prevented counts open=%v / %d / %d", open, prevented, closed)
	}
	block = false
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Advance(30 * time.Millisecond)
	click(t, tt, "Panel action")
	if actions != 0 || closed != 1 || tt.Focused("Open panel") {
		t.Fatal("exiting panel accepted input or restored focus too early")
	}
	open = true
	tt.Frame()
	tt.Advance(250 * time.Millisecond)
	if bounds(t, tt, "Panel") != settled {
		t.Fatal("exit reversal did not reach the same inset position")
	}
	click(t, tt, "Panel action")
	if actions != 1 {
		t.Fatal("reopened panel stayed disabled")
	}
	transition = false
	open = false
	tt.Frame()
	if _, ok := tt.Find("Panel"); ok || !tt.Focused("Open panel") {
		t.Fatal("immediate exit did not unmount and restore focus")
	}
}

func TestSlideoverDirectionInsetAndIgnoredModalFieldsPreserveSelection(t *testing.T) {
	open, inset, ignored := true, false, false
	side := nui.SideRight
	value := "Selected draft"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		o := nui.SlideoverOptions{Title: "Panel", Transition: nui.Set(false), Side: side, Inset: inset, Extent: nui.Set(float32(300))}
		if ignored {
			o.Width = nui.Set(float32(10))
			o.MaxHeight = nui.Set(float32(20))
			o.Fullscreen = true
		}
		nui.Slideover(c.Key("panel"), th, &open, o, func(th nui.Theme) { nui.Input(c.Key("draft"), th, &value, nui.InputOptions{Label: "Panel draft"}) })
	}, 600, 500)
	before := bounds(t, tt, "Panel")
	ignored = true
	tt.Frame()
	if bounds(t, tt, "Panel") != before {
		t.Fatal("inherited Modal-only fields changed Slideover geometry")
	}
	click(t, tt, "Panel draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	inset, side = true, nui.SideLeft
	tt.SetDirection(ui.RightToLeft)
	tt.Frame()
	if !tt.Focused("Panel draft") || bounds(t, tt, "Panel").X != 16 {
		t.Fatal("physical side/RTL layout lost editor focus")
	}
	tt.Type("Changed")
	if value != "Changed" {
		t.Fatalf("side/inset update discarded selection: %q", value)
	}
}

func TestSlideoverDarkInsetAndTextScaling(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.DarkMode})
		nui.Slideover(c, th, &open, nui.SlideoverOptions{Inset: true, Extent: nui.Set(float32(300)), Title: "Dark panel", Transition: nui.Set(false), Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Label {
				e.Label("Scaled heading")
			}
		}}, nil)
	}, 600, 400)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5})
	r := bounds(t, tt, "Dark panel")
	if r.X != 284 || r.Y != 16 || r.H != 368 || bounds(t, tt, "Scaled heading").H != 36 {
		t.Fatalf("dark scaled inset layout=%v", r)
	}
	if got := tt.Image().RGBAAt(int(r.X+8), int(r.Y+30)); got != rgba(nui.Dark().Background) {
		t.Fatalf("dark panel background=%v", got)
	}
}
