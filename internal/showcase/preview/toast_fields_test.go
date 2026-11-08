package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestToastPreviewUsesLibraryViewport(t *testing.T) {
	tester, _ := exampleTester(t, "toast")
	clickExample(t, tester, "Show notification")
	if !tester.HasText("Dismiss Saved successfully") {
		t.Fatal("toast preview did not build the NuxtUI Go close action")
	}
}

func TestToastPreviewNestedOptionsLiveViewportAndStableActions(t *testing.T) {
	tester, cfg := exampleTester(t, "toast")
	cfg.Options = map[string]any{"ToastOptions": map[string]any{"Avatar": map[string]any{"Alt": "Preview avatar", "Text": "PA"}, "CloseOptions": map[string]any{"AccessibleLabel": "Close configured toast", "Size": "nui.XL"}, "ProgressOptions": map[string]any{"UI": map[string]any{"indicator": map[string]any{"Background": "#ff00ff"}}}}}
	tester.Frame()
	clickExample(t, tester, "Show notification")
	if !tester.HasText("PA") || !tester.HasText("Close configured toast") {
		t.Fatal("nested toast options did not render")
	}
	panel, _ := tester.Find("Saved successfully")
	if pixel := tester.Image().RGBAAt(int(panel.X+panel.W/2), int(panel.Y+panel.H-2)); pixel.R != 255 || pixel.B != 255 || pixel.G != 0 {
		t.Fatalf("nested progress styling did not apply: %v", pixel)
	}
	clickExample(t, tester, "Keep notification")
	clickExample(t, tester, "Update focused notification")
	if !tester.Focused("Keep notification") {
		t.Fatal("preview update lost the focused action")
	}
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Published: 2 · Actions: 2 · Closed: 0") {
		t.Fatal("updated focused action executed the wrong callback")
	}
	cfg.Options["ToastViewportOptions"] = map[string]any{"Position": "nui.ToastTopLeft", "Width": float64(240), "Limit": float64(1)}
	tester.Frame()
	panel, _ = tester.Find("Saved successfully")
	if panel.X != 16 || panel.Y != 16 || panel.W != 240 {
		t.Fatalf("live viewport override did not apply: %v", panel)
	}
	clickExample(t, tester, "Close configured toast")
	if !tester.HasText("Published: 2 · Actions: 2 · Closed: 1") {
		t.Fatal("configured close callback count is incorrect")
	}
}
