package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestProgressPreviewMaximumDirectionAndNeutral(t *testing.T) {
	tester, cfg := exampleTester(t, "progress")
	cfg.Values["max"], cfg.Values["value"] = float64(200), float64(50)
	cfg.Values["color"], cfg.Values["customNeutral"] = "neutral", true
	for _, mode := range []string{"light", "dark"} {
		cfg.Mode = mode
		cfg.Values["inverted"] = false
		tester.Frame()
		r, ok := tester.Find("Progress")
		if !ok {
			t.Fatal("missing progress")
		}
		x1, x2, y := int(r.X+r.W*.1), int(r.X+r.W*.5), int(r.Y+r.H/2)
		fill, empty := tester.Image().RGBAAt(x1, y), tester.Image().RGBAAt(x2, y)
		want := ui.Hex("#7e22ce")
		if mode == "dark" {
			want = ui.Hex("#c084fc")
		}
		if fill.R != want.R || fill.G != want.G || fill.B != want.B || fill == empty {
			t.Fatalf("maximum or neutral color control ignored: %v / %v", fill, empty)
		}
		cfg.Values["inverted"] = true
		tester.Frame()
		if tester.Image().RGBAAt(x1, y) != empty || tester.Image().RGBAAt(int(r.X+r.W*.9), y) != fill {
			t.Fatal("inverted control did not reverse progress")
		}
	}
}
