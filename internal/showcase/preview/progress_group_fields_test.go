package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestProgressGroupPreviewColorsMaximumAndEmptyItems(t *testing.T) {
	tester, cfg := exampleTester(t, "progress-group")
	cfg.Values["color"], cfg.Values["customNeutral"] = "neutral", true
	cfg.Values["customSegment"] = true
	cfg.Values["max"] = float64(200)
	tester.Frame()
	if !tester.HasText("23%") || !tester.HasText("15%") || !tester.HasText("8%") {
		t.Fatal("Max control did not update legend percentages")
	}
	r, ok := tester.Find("Project progress")
	if !ok {
		t.Fatal("missing group track")
	}
	for _, sample := range []struct {
		at    float32
		color ui.Color
	}{{.1, ui.Hex("#7e22ce")}, {.3, ui.Hex("#f97316")}} {
		p := tester.Image().RGBAAt(int(r.X+r.W*sample.at), int(r.Y+r.H/2))
		if p.R != sample.color.R || p.G != sample.color.G || p.B != sample.color.B {
			t.Fatalf("fill at %g ignored its configuration: %v", sample.at, p)
		}
	}
	cfg.Values["hideLegend"] = true
	tester.Frame()
	if tester.HasText("Code") || tester.HasText("23%") {
		t.Fatal("HideLegend control left legend content")
	}
	cfg.Options = map[string]any{"ProgressGroupItems": []any{}}
	tester.Frame()
	if _, ok := tester.Find("Project progress"); !ok {
		t.Fatal("empty items removed the unused track")
	}
}
