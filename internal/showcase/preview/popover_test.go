package preview

import (
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func configuredPopover(t *testing.T, fields map[string]any) (*ui.Tester, *showcase.Config) {
	t.Helper()
	fields["Label"] = "Configured popup"
	demo, _ := showcase.Find("popover")
	cfg, err := demo.Normalize(showcase.Config{Options: map[string]any{"PopoverOptions": fields}})
	if err != nil {
		t.Fatal(err)
	}
	example, _ := New("popover")
	t.Cleanup(example.Dispose)
	tester := ui.NewTester(func(c *ui.Context) {
		app := cfg.Theme()
		app.Light.Motion.Reduce = true
		app.Dark.Motion.Reduce = true
		theme := nui.App(c, app)
		ui.Column(c.Key("preview")).Fill().Padding(24).Gap(16).Children(func() { example.View(c, theme, cfg) })
	}, 760, demo.Height)
	return tester, &cfg
}

func popupBounds(t *testing.T, tester *ui.Tester, label string) ui.Rect {
	t.Helper()
	r, ok := tester.Find(label)
	if !ok {
		t.Fatalf("missing %s", label)
	}
	return r
}

func TestPopoverConfiguredPlacementAndSizing(t *testing.T) {
	fields := map[string]any{"Width": float64(180), "Offset": float64(18)}
	tester, cfg := configuredPopover(t, fields)
	if err := tester.Click("Open details"); err != nil {
		t.Fatal(err)
	}
	fields = cfg.Options["PopoverOptions"].(map[string]any)
	for _, placement := range []struct {
		name                   string
		ax, ay, px, py, dx, dy float32
	}{
		{"BottomStart", 0, 1, 0, 0, 0, 18}, {"BottomCenter", .5, 1, .5, 0, 0, 18}, {"BottomEnd", 1, 1, 1, 0, 0, 18},
		{"TopStart", 0, 0, 0, 1, 0, -18}, {"TopCenter", .5, 0, .5, 1, 0, -18}, {"TopEnd", 1, 0, 1, 1, 0, -18},
		{"LeftStart", 0, 0, 1, 0, -18, 0}, {"LeftCenter", 0, .5, 1, .5, -18, 0}, {"LeftEnd", 0, 1, 1, 1, -18, 0},
		{"RightStart", 1, 0, 0, 0, 18, 0}, {"RightCenter", 1, .5, 0, .5, 18, 0}, {"RightEnd", 1, 1, 0, 1, 18, 0},
	} {
		t.Run(placement.name, func(t *testing.T) {
			fields["Placement"] = "nui." + placement.name
			tester.Frame()
			a := popupBounds(t, tester, "Open details")
			p := popupBounds(t, tester, "Configured popup")
			x := a.X + placement.ax*a.W - placement.px*p.W + placement.dx
			y := a.Y + placement.ay*a.H - placement.py*p.H + placement.dy
			if math.Abs(float64(p.X-x)) > 1 || math.Abs(float64(p.Y-y)) > 1 || p.W != 180 {
				t.Fatalf("placement did not take effect: anchor=%v panel=%v want=(%v,%v)", a, p, x, y)
			}
		})
	}
	fields["Placement"] = "nui.BottomStart"
	fields["Width"] = float64(260)
	fields["MaxHeight"] = float64(60)
	tester.Frame()
	p := popupBounds(t, tester, "Configured popup")
	if p.W != 260 || p.H > 60 {
		t.Fatalf("width/maximum height did not take effect: %v", p)
	}
	fields["Width"] = nil // Clear the demo's explicit width to test content sizing.
	delete(fields, "MaxHeight")
	cfg.Values["label"] = "Open additional workspace configuration details"
	tester.Frame()
	fields["MatchAnchor"] = true
	tester.Frame()
	a := popupBounds(t, tester, cfg.Values["label"].(string))
	p = popupBounds(t, tester, "Configured popup")
	if p.W < a.W {
		t.Fatalf("MatchAnchor did not set minimum width: anchor=%v panel=%v", a, p)
	}
	fields["MatchAnchor"] = false
	tester.Frame()
	width := popupBounds(t, tester, "Configured popup").W
	fields["UI"] = map[string]any{"panel": map[string]any{"PaddingX": float64(24)}}
	tester.Frame()
	if popupBounds(t, tester, "Configured popup").W != width+16 {
		t.Fatal("panel UI did not apply")
	}
	fields["Style"] = map[string]any{"mode": "style", "parts": map[string]any{"panel": map[string]any{"PaddingX": float64(32)}}}
	tester.Frame()
	if popupBounds(t, tester, "Configured popup").W != width+32 {
		t.Fatal("final panel Style did not override UI")
	}
}

func TestPopoverConfiguredHoverDelays(t *testing.T) {
	tester, _ := configuredPopover(t, map[string]any{"Mode": "nui.PopoverHover", "OpenDelay": float64(2000), "CloseDelay": float64(500)})
	a := popupBounds(t, tester, "Open details")
	tester.Move(a.X+a.W/2, a.Y+a.H/2)
	if err := tester.Click("Open details"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(1999 * time.Millisecond)
	if tester.HasText("Additional information") {
		t.Fatal("opened before configured 2000 ms")
	}
	tester.Advance(2 * time.Millisecond)
	if !tester.HasText("Additional information") {
		t.Fatal("did not open after configured delay")
	}
	p := popupBounds(t, tester, "Configured popup")
	tester.Move(p.X+8, p.Y+8)
	tester.Advance(time.Second)
	if !tester.HasText("Additional information") {
		t.Fatal("closed while hovering the panel")
	}
	tester.Move(740, 400)
	tester.Advance(499 * time.Millisecond)
	if !tester.HasText("Additional information") {
		t.Fatal("closed before configured 500 ms")
	}
	tester.Advance(2 * time.Millisecond)
	if tester.HasText("Additional information") {
		t.Fatal("did not close after configured delay")
	}
}

func TestPopoverPreviewReanchorPreservesDraft(t *testing.T) {
	tester, cfg := exampleTester(t, "popover")
	clickExample(t, tester, "Open details")
	clickExample(t, tester, "Popover draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["alternate"], cfg.Values["styled"], cfg.Values["reduced"] = true, true, true
	tester.Frame()
	if !tester.Focused("Popover draft") {
		t.Fatal("panel controls discarded editor focus")
	}
	tester.Type("Reanchored draft")
	clickExample(t, tester, "Close")
	if !tester.Focused("Alternate anchor") {
		t.Fatal("closing did not focus the configured anchor")
	}
	clickExample(t, tester, "Alternate anchor")
	clickExample(t, tester, "Popover draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Command("copy")
	if tester.Clipboard() != "Reanchored draft" {
		t.Fatal("closing lost the persistent draft")
	}
}
