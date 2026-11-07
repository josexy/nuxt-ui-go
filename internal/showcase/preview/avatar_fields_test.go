package preview

import "testing"

func TestAvatarPreviewFallbacksAndImageResize(t *testing.T) {
	tester, cfg := exampleTester(t, "avatar")
	if !tester.HasText("AL") {
		t.Fatal("initials missing")
	}
	cfg.Values["text"] = "+12"
	tester.Frame()
	if !tester.HasText("+12") || tester.HasText("AL") {
		t.Fatal("literal fallback did not replace initials")
	}
	cfg.Values["icon"] = "user"
	tester.Frame()
	if tester.HasText("+12") {
		t.Fatal("icon did not replace fallback text")
	}
	cfg.Values["chip"], cfg.Values["enlarge"] = true, true
	cfg.Options = map[string]any{"AvatarOptions": map[string]any{"Image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR4nGP4z8DwHwAFAAH/iZk9HQAAAABJRU5ErkJggg=="}}
	tester.Frame()
	r, ok := tester.Find(cfg.String("label"))
	if !ok || r.W != 64 || r.H != 64 {
		t.Fatalf("custom root dimensions=%+v", r)
	}
	// The sample is opaque red. It must reach beyond the old fixed 32 DIP image.
	for _, offset := range []int{8, 56} {
		pixel := tester.Image().RGBAAt(int(r.X)+offset, int(r.Y)+32)
		if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
			t.Fatalf("image does not fill resized root at %d: %v", offset, pixel)
		}
	}
	cfg.Values["enlarge"], cfg.Values["chip"] = false, false
	tester.Frame()
	if r, _ := tester.Find(cfg.String("label")); r.W != 32 || r.H != 32 {
		t.Fatalf("reset did not restore MD size: %+v", r)
	}
}

func TestAvatarGroupPreviewReorderAndZeroLimit(t *testing.T) {
	tester, cfg := exampleTester(t, "avatar-group")
	if !tester.HasText("1 more person") || tester.HasText("Lin Chen") {
		t.Fatal("default maximum did not show two avatars and overflow")
	}
	clickExample(t, tester, "Reverse avatars")
	if !tester.HasText("Lin Chen") || tester.HasText("Ada Lovelace") {
		t.Fatal("reorder did not change visible members")
	}
	cfg.Values["maximum"], cfg.Values["memberStyle"] = float64(0), true
	tester.Frame()
	if !tester.HasText("Ada Lovelace") || tester.HasText("1 more person") {
		t.Fatal("maximum zero did not restore all members")
	}
	a, _ := tester.Find("Ada Lovelace")
	l, _ := tester.Find("Lin Chen")
	if l.X >= a.X {
		t.Fatal("live option changes discarded reversed order")
	}
}
