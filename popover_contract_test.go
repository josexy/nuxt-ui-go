package nui_test

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestPopoverReducedMotionRestorationDoesNotRestartEntry(t *testing.T) {
	for _, initialReduced := range []bool{false, true} {
		reduced, open := initialReduced, false
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			th.Motion.Reduce = reduced
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
		th.Motion.Reduce = true
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
		th.Motion.Reduce = true
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
