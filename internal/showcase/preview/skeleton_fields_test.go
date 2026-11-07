package preview

import "testing"

func TestSkeletonPreviewRadiusOverride(t *testing.T) {
	tester, cfg := exampleTester(t, "skeleton")
	cfg.Values["radius"] = float64(30)
	tester.Frame()
	r, ok := tester.Find("Loading…")
	if !ok {
		t.Fatal("missing skeleton")
	}
	x, y := int(r.X)+3, int(r.Y)+3
	before := tester.Image().RGBAAt(x, y)
	center := tester.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2))
	if before == center {
		t.Fatal("radius control did not round the placeholder")
	}
	cfg.Values["square"] = true
	tester.Frame()
	if got := tester.Image().RGBAAt(x, y); got != center {
		t.Fatalf("Style radius did not override the UI radius: %v != %v", got, center)
	}
}
