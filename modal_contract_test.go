package nui_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestModalBackdropOpacityDoesNotJumpAtAnimationEnd(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Dialog", Decoration: nui.Decoration{UI: nui.Styles{nui.Backdrop: {Background: nui.Set(ui.RGBA(0, 0, 0, .5))}}}}, nil)
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
		th.Motion.Reduce = true
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{CloseButton: nui.Set(false), Decoration: nui.Decoration{UI: nui.Styles{nui.Panel: {Foreground: nui.Set(ui.RGB(255, 0, 255)), FontSize: nui.Set(float32(24)), LineHeight: nui.Set(float32(1.5))}}, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Header {
				e.Label("Empty header")
			}
		}}}, func(nui.Theme) { ui.Text(c, "Plain body text") }).Label("Dialog panel")
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
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{CloseButton: nui.Set(false), Overlay: nui.Set(false), Decoration: nui.Decoration{UI: nui.Styles{nui.Panel: {Background: nui.Set(ui.Transparent), Radius: nui.Set(float32(24)), BorderWidth: nui.Set(float32(0)), Shadow: nui.Set(nui.ShadowTokens{})}, nui.Body: {PaddingX: nui.Set(float32(0)), PaddingY: nui.Set(float32(0)), Background: nui.Set(ui.RGB(255, 0, 0))}}}}, func(nui.Theme) { ui.Box(c).FillWidth().Height(80).Background(ui.RGB(255, 0, 0)) }).Label("Rounded dialog")
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
