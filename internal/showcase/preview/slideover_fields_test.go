package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSlideoverPreviewPlacementPolicyAndEmptyBody(t *testing.T) {
	tester, cfg := exampleTester(t, "slideover")
	cfg.Values["transition"] = false
	tester.Frame()
	clickExample(t, tester, "Open panel")
	clickExample(t, tester, "Panel draft")
	tester.Type("Selected panel draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["side"], cfg.Values["inset"], cfg.Values["dismissible"] = "left", true, false
	tester.Frame()
	if !tester.Focused("Panel draft") {
		t.Fatal("side/inset change lost focus")
	}
	tester.Type("Panel replacement")
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Key(ui.Ctrl, ui.KeyC)
	if tester.Clipboard() != "Panel replacement" {
		t.Fatal("placement change lost selection")
	}
	tester.Key(0, ui.KeyEscape)
	if !tester.HasText("Panel draft") {
		t.Fatal("blocked Escape closed the panel")
	}
	clickExample(t, tester, "Done")
	if !tester.Focused("Open panel") || !tester.HasText("Panel notifications: closed=0 prevented=1") {
		t.Fatal("Done count or focus is incorrect")
	}
	cfg.Values["side"], cfg.Values["emptyBody"] = "top", true
	tester.Frame()
	clickExample(t, tester, "Open panel")
	r, ok := tester.Find("Project details")
	if !ok || r.H != 178 || r.Y != 16 || tester.HasText("Panel draft") {
		t.Fatalf("empty inset top panel=%v", r)
	}
	cfg.Options = map[string]any{"SlideoverOptions": map[string]any{"Width": float64(10), "MaxHeight": float64(20), "Fullscreen": true}}
	tester.Frame()
	if after, _ := tester.Find("Project details"); after != r {
		t.Fatal("unused inherited Modal fields changed preview layout")
	}
}
