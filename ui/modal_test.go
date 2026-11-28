package nui_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestModalBackdropOpacityDoesNotJumpAtAnimationEnd(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Dialog", UI: nui.Styles{nui.Backdrop: {Background: nui.Set(ui.RGBA(0, 0, 0, .5))}}}, nil)
	}, 600, 400)
	tt.Advance(199 * time.Millisecond)
	before := tt.Image().RGBAAt(4, 4).R
	tt.Advance(2 * time.Millisecond)
	after := tt.Image().RGBAAt(4, 4).R
	if math.Abs(float64(before)-float64(after)) > 2 {
		t.Fatalf("backdrop jumps at completion: %d -> %d", before, after)
	}
}

func TestModalEmptyHeaderAndInheritedPanelText(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Disabled = true
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{CloseButton: nui.Set(false), UI: nui.Styles{nui.Panel: {Foreground: nui.Set(ui.RGB(255, 0, 255)), FontSize: nui.Set(float32(24)), LineHeight: nui.Set(float32(1.5))}}, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Header {
				e.Label("Empty header")
			}
		}}, func(nui.Theme) { ui.Text(c, "Plain body text") }).Label("Dialog panel")
	}, 600, 400)
	if _, ok := tt.Find("Empty header"); ok {
		t.Fatal("empty title/description/close left a header")
	}
	r := bounds(t, tt, "Plain body text")
	if r.H != 36 || countColor(tt.Image(), r, rgba(ui.RGB(255, 0, 255))) == 0 {
		t.Fatalf("Panel text style lost in body: %v", r)
	}
}

func TestModalClipsContentToPanelCorners(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Reduce = true
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{CloseButton: nui.Set(false), Overlay: nui.Set(false), UI: nui.Styles{nui.Panel: {Background: nui.Set(ui.Transparent), Radius: nui.Set(float32(24)), BorderWidth: nui.Set(float32(0)), Shadow: nui.Set(nui.ShadowTokens{})}, nui.Body: {PaddingX: nui.Set(float32(0)), PaddingY: nui.Set(float32(0)), Background: nui.Set(ui.RGB(255, 0, 0))}}}, func(nui.Theme) { ui.Box(c).FillWidth().Height(80).Background(ui.RGB(255, 0, 0)) }).Label("Rounded dialog")
	}, 600, 400)
	r := bounds(t, tt, "Rounded dialog")
	if got := tt.Image().RGBAAt(int(r.X)+1, int(r.Y+r.H)-2); got == rgba(ui.RGB(255, 0, 0)) {
		t.Fatal("body painted outside rounded panel")
	}
}

func TestModalClosePoliciesAndCustomButton(t *testing.T) {
	open, dismissible := false, false
	closed, prevented := 0, 0
	custom := nui.ButtonOptions{Size: nui.XL, Color: nui.Error, Variant: nui.Outline, AccessibleLabel: "Configured close", Style: func(_, _, leading, _ ui.Element) { leading.Label("Close glyph") }}
	var opener ui.Handle
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Button(c.Key("open"), th, "Open dialog", nui.ButtonOptions{}).Bind(&opener).OnClick(func() { open = true })
		nui.ModalWithOptions(c.Key("dialog"), th, &open, nui.ModalOptions{Title: "Configured dialog", Transition: nui.Set(false), Overlay: nui.Set(false), Dismissible: nui.Set(dismissible), Close: &custom, CloseIcon: lucide.ArrowRight(), ReturnFocus: &opener, OnClose: func() { closed++ }, OnClosePrevent: func() { prevented++ }}, func(nui.Theme) { ui.Text(c, "Body") })
	}, 600, 400)
	click(t, tt, "Open dialog")
	if r := bounds(t, tt, "Configured dialog"); r.W != 512 {
		t.Fatalf("Transition=false still scales: %v", r)
	}
	if r := bounds(t, tt, "Close glyph"); r.W != 24 {
		t.Fatalf("custom close icon size=%v", r)
	}
	if custom.Leading != nil || custom.UI != nil {
		t.Fatal("close configuration was mutated")
	}
	tt.Key(0, ui.KeyEscape)
	tt.ClickAt(5, 5)
	if !open || closed != 0 || prevented != 2 {
		t.Fatalf("prevented dismissal open=%v close=%d prevented=%d", open, closed, prevented)
	}
	custom.Loading = true
	tt.Frame()
	click(t, tt, "Configured close")
	if !open || closed != 0 {
		t.Fatal("loading close button dismissed dialog")
	}
	custom.Loading = false
	tt.Frame()
	click(t, tt, "Configured close")
	if open || closed != 1 || !tt.Focused("Open dialog") {
		t.Fatalf("close/focus: %v/%d", open, closed)
	}
	dismissible = true
	click(t, tt, "Open dialog")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if open || closed != 2 || prevented != 2 {
		t.Fatal("Escape callback counts are incorrect")
	}
}

func TestModalHeaderAndFullscreenPreserveDraftSelection(t *testing.T) {
	open, fullscreen := true, false
	title, value := "Edit", "Selected draft"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ModalWithOptions(c.Key("dialog"), th, &open, nui.ModalOptions{Title: title, CloseButton: nui.Set(false), Transition: nui.Set(false), Fullscreen: fullscreen}, func(th nui.Theme) { nui.Input(c.Key("draft"), th, &value, nui.InputOptions{Label: "Draft"}) })
	}, 600, 400)
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	title, fullscreen = "", true
	tt.Frame()
	if !tt.Focused("Draft") {
		t.Fatal("removing header/fullscreen lost editor focus")
	}
	tt.Type("Replacement")
	if value != "Replacement" {
		t.Fatalf("selection lost: %q", value)
	}
	open = false
	tt.Frame()
	open = true
	tt.Frame()
	if value != "Replacement" {
		t.Fatal("application draft did not survive unmount")
	}
}

func TestModalFooterOnlyHasNoOrphanDivider(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{CloseButton: nui.Set(false), Transition: nui.Set(false), Footer: func(c *ui.Context, th nui.Theme) { nui.Button(c, th, "Done", nui.ButtonOptions{}) }, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Panel {
				e.BorderWidth(0)
			}
			if slot == nui.Footer {
				e.Label("Only footer")
			}
		}}, nil).Label("Footer dialog")
	}, 600, 400)
	panel, footer := bounds(t, tt, "Footer dialog"), bounds(t, tt, "Only footer")
	if panel.H != footer.H || panel.Y != footer.Y {
		t.Fatalf("empty regions left space: panel=%v footer=%v", panel, footer)
	}
	if footer.H != 64 {
		t.Fatalf("lone footer includes more than the 32 DIP button and 32 DIP padding: %v", footer)
	}
}

func TestModalBodyScrollKeepsHeaderAndFooterFixed(t *testing.T) {
	open := true
	var scroll ui.ScrollState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Scrolling dialog", Transition: nui.Set(false), MaxHeight: nui.Set(float32(260)), Footer: func(c *ui.Context, th nui.Theme) { nui.Button(c, th, "Pinned footer", nui.ButtonOptions{}) }, Style: func(slot nui.Slot, e ui.Element) {
			switch slot {
			case nui.Header:
				e.Label("Pinned header")
			case nui.Body:
				e.Label("Scroll body").TrackScroll(&scroll)
			}
		}}, func(nui.Theme) {
			for i := range 20 {
				ui.Text(c.Key(i), fmt.Sprintf("Scrollable line %d", i)).Height(30)
			}
		})
	}, 600, 400)
	header, footer, body := bounds(t, tt, "Pinned header"), bounds(t, tt, "Pinned footer"), bounds(t, tt, "Scroll body")
	tt.Scroll(body.X+30, body.Y+30, 0, 150)
	if scroll.Y <= 0 || bounds(t, tt, "Pinned header") != header || bounds(t, tt, "Pinned footer") != footer {
		t.Fatalf("body scroll moved fixed regions, offset=%v", scroll.Y)
	}
}

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
	action := bounds(t, tt, "Panel action")
	tt.Key(0, ui.KeyEscape)
	// The leaving panel is inert: focus returns as it closes, and a click
	// where its action was does not reach it.
	if closed != 1 || !tt.Focused("Open panel") {
		t.Fatalf("closing did not return focus at once: closed=%d", closed)
	}
	tt.Advance(30 * time.Millisecond)
	tt.ClickAt(action.X+action.W/2, action.Y+action.H/2)
	if actions != 0 {
		t.Fatal("the leaving panel accepted input")
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

func TestPopoverMotionRestorationDoesNotRestartEntry(t *testing.T) {
	for _, initialReduced := range []bool{false, true} {
		reduced, open := initialReduced, false
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			th.Motion.Disabled = reduced
			ui.Column(c).Padding(24).Children(func() {
				a := nui.Button(c, th, "Details", nui.ButtonOptions{}).OnClick(func() { open = !open })
				nui.Popover(c, th, a, &open, nui.PopoverOptions{Label: "Panel", Width: nui.Set(float32(220))}, func(nui.Theme) { ui.Text(c, "Panel content") })
			})
		}, 360, 300)
		click(t, tt, "Details")
		tt.Advance(30 * time.Millisecond)
		reduced = true
		tt.Frame()
		settled := bounds(t, tt, "Panel")
		reduced = false
		tt.Frame()
		resumed := bounds(t, tt, "Panel")
		if resumed != settled {
			t.Fatalf("initial reduced=%v restarted entry: %v -> %v", initialReduced, settled, resumed)
		}
		tt.Advance(10 * time.Millisecond)
		if tt.FramePending() {
			t.Fatal("restoring motion animated an already visible panel")
		}
	}
}

func TestPopoverReanchorAndSizingPreserveEditor(t *testing.T) {
	open, second, draft := false, false, "Original draft"
	width, offset := float32(180), float32(0)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		ui.Column(c).Fill().Padding(24).Children(func() {
			a := nui.Button(c.Key("first"), th, "First anchor", nui.ButtonOptions{}).OnClick(func() { open = !open })
			b := nui.Button(c.Key("second"), th, "Second anchor", nui.ButtonOptions{}).Absolute().Left(300).Top(160)
			anchor := a
			if second {
				anchor = b
			}
			nui.Popover(c.Key("popup"), th, anchor, &open, nui.PopoverOptions{Label: "Panel", Width: nui.Set(width), Offset: nui.Set(offset), MatchAnchor: true}, func(inner nui.Theme) {
				nui.Input(c.Key("draft"), inner, &draft, nui.InputOptions{Label: "Popup editor"})
			})
		})
	}, 640, 400)
	click(t, tt, "First anchor")
	click(t, tt, "Popup editor")
	tt.Key(ui.Ctrl, ui.KeyA)
	second = true
	width = 240
	offset = 18
	tt.Frame()
	a, p := bounds(t, tt, "Second anchor"), bounds(t, tt, "Panel")
	if p.X != a.X || p.Y != a.Y+a.H+18 || p.W != 240 || !tt.Focused("Popup editor") {
		t.Fatalf("reanchor/layout/focus: anchor=%v panel=%v", a, p)
	}
	tt.Type("Preserved selection")
	if draft != "Preserved selection" {
		t.Fatal("reanchoring discarded editor selection")
	}
	tt.Key(0, ui.KeyEscape)
	if open || !tt.Focused("Second anchor") {
		t.Fatal("closing did not return to current anchor")
	}
}

func TestPopoverInitialControlledOpenReturnsToAnchor(t *testing.T) {
	open := false
	draft := ""
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(24).Children(func() {
			a := nui.Button(c, th, "Anchor", nui.ButtonOptions{})
			nui.Button(c, th, "External open", nui.ButtonOptions{}).OnClick(func() { open = true })
			nui.Popover(c, th, a, &open, nui.PopoverOptions{Label: "Panel"}, func(inner nui.Theme) { nui.Input(c, inner, &draft, nui.InputOptions{Label: "Draft"}) })
		})
	}, 400, 320)
	click(t, tt, "External open")
	click(t, tt, "Draft")
	tt.Key(0, ui.KeyEscape)
	if open || !tt.Focused("Anchor") {
		t.Fatal("controlled popup restored an unrelated opener")
	}
}

func TestPopoverPanelTypographyReachesPlainContent(t *testing.T) {
	open := true
	callback := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Disabled = true
		a := nui.Button(c, th, "Anchor", nui.ButtonOptions{})
		nui.Popover(c, th, a, &open, nui.PopoverOptions{UI: nui.Styles{nui.Panel: {Foreground: nui.Set(ui.RGB(255, 0, 255)), FontSize: nui.Set(float32(24)), LineHeight: nui.Set(float32(1.5))}}, Style: func(_ nui.Slot, e ui.Element) {
			if callback {
				e.TextColor(ui.RGB(0, 180, 0)).FontSize(28).LineHeight(2).Gap(24)
			}
		}}, func(nui.Theme) { ui.Text(c, "Popup text"); ui.Text(c, "Second line") })
	}, 400, 260)
	r := bounds(t, tt, "Popup text")
	if r.H != 36 || countColor(tt.Image(), r, rgba(ui.RGB(255, 0, 255))) == 0 {
		t.Fatalf("panel typography was overridden: %+v", r)
	}
	if gap := bounds(t, tt, "Second line").Y - r.Y - r.H; gap != 8 {
		t.Fatalf("default panel gap=%v", gap)
	}
	callback = true
	tt.Frame()
	r = bounds(t, tt, "Popup text")
	if r.H != 56 || countColor(tt.Image(), r, rgba(ui.RGB(0, 180, 0))) == 0 || bounds(t, tt, "Second line").Y-r.Y-r.H != 24 {
		t.Fatalf("final panel Style did not override text or gap: text=%v second=%v green=%d", r, bounds(t, tt, "Second line"), countColor(tt.Image(), r, rgba(ui.RGB(0, 180, 0))))
	}
}

func TestPopoverLocalNativeThemeDoesNotLeak(t *testing.T) {
	open := true
	inside, outside := ui.RGB(126, 34, 206), ui.RGB(0, 160, 120)
	tt := ui.NewTester(func(c *ui.Context) {
		app := nui.AppConfig{Mode: nui.LightMode}
		app.Light = nui.Light()
		app.Light.Primary = outside
		th := nui.App(c, app)
		th.Motion.Disabled = true
		ui.Column(c).Padding(24).Children(func() {
			a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Width(120)
			local := th
			local.Primary = inside
			nui.Popover(c, local, a, &open, nui.PopoverOptions{Placement: nui.RightStart, Width: nui.Set(float32(200))}, func(nui.Theme) { ui.PrimaryButton(c, "Native inside").Label("Inside button") })
			ui.PrimaryButton(c, "Native outside").Label("Outside button")
		})
	}, 520, 260)
	for label, want := range map[string]ui.Color{"Inside button": inside, "Outside button": outside} {
		r := bounds(t, tt, label)
		if got := tt.Image().RGBAAt(int(r.X+12), int(r.Y+r.H/2)); got != rgba(want) {
			t.Fatalf("%s theme=%v want %v bounds=%v", label, got, rgba(want), r)
		}
	}
}
