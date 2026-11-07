package preview

import (
	"fmt"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func scrollOffsets(t *testing.T, tester *ui.Tester) (float64, float64) {
	t.Helper()
	for _, text := range tester.Texts() {
		if strings.HasPrefix(text, "Scroll position:") {
			var x, y float64
			if _, err := fmt.Sscanf(text, "Scroll position: x=%f y=%f", &x, &y); err != nil {
				t.Fatal(err)
			}
			return x, y
		}
	}
	t.Fatal("missing scroll offsets")
	return 0, 0
}

func TestScrollAreaPreviewAxesAndPositionControls(t *testing.T) {
	tester, cfg := exampleTester(t, "scroll-area")
	cfg.Values["direction"] = "both"
	tester.Frame()
	clickExample(t, tester, "Scroll to end")
	x, y := scrollOffsets(t, tester)
	if x <= 0 || y <= 0 {
		t.Fatalf("both-axis end=(%v,%v)", x, y)
	}
	cfg.Values["width"], cfg.Values["customColor"], cfg.Values["rounded"] = float64(14), true, false
	tester.Frame()
	if a, b := scrollOffsets(t, tester); a != x || b != y {
		t.Fatal("thumb styling reset scroll state")
	}
	cfg.Values["direction"] = "horizontal"
	tester.Frame()
	if a, b := scrollOffsets(t, tester); a != x || b != 0 {
		t.Fatalf("horizontal offsets=(%v,%v)", a, b)
	}
	clickExample(t, tester, "Scroll to start")
	if a, b := scrollOffsets(t, tester); a != 0 || b != 0 {
		t.Fatal("start action did not reset scroll state")
	}
}

func TestScrollAreaPreviewUsesOverriddenStateAndDirection(t *testing.T) {
	tester, cfg := exampleTester(t, "scroll-area")
	cfg.Options = map[string]any{"ScrollAreaOptions": map[string]any{"State": nil}}
	tester.Frame()
	if !tester.HasText("Attach State to inspect offsets") || tester.HasText("Scroll to end") {
		t.Fatal("nil State left misleading position controls")
	}
	cfg.Options["ScrollAreaOptions"] = map[string]any{"Direction": "nui.ScrollBoth", "State": map[string]any{"X": float64(60), "Y": float64(80)}}
	tester.Frame()
	if x, y := scrollOffsets(t, tester); x != 60 || y != 80 {
		t.Fatalf("overridden state display=(%v,%v)", x, y)
	}
	clickExample(t, tester, "Scroll to start")
	if x, y := scrollOffsets(t, tester); x != 0 || y != 0 {
		t.Fatal("position buttons edited the unused default state")
	}
}
