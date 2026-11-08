package preview

import "testing"

func TestBadgePreviewAvatarAndAutomaticSquare(t *testing.T) {
	tester, cfg := exampleTester(t, "badge")
	cfg.Values["avatar"] = true
	tester.Frame()
	if !tester.HasText("Ada Lovelace") {
		t.Fatal("avatar control did not display badge avatar")
	}
	cfg.Values["icon"] = "check"
	tester.Frame()
	if tester.HasText("Ada Lovelace") {
		t.Fatal("leading icon did not replace avatar")
	}
	cfg.Values["label"] = ""
	tester.Frame()
	r, ok := tester.Find("Badge preview")
	if !ok || r.W != 24 || r.H != 24 {
		t.Fatalf("empty MD badge is not square: %+v", r)
	}
}
