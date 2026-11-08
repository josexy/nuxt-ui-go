package preview

import "testing"

func TestButtonPreviewAvatarAndLoaderPriority(t *testing.T) {
	tester, cfg := exampleTester(t, "button")
	cfg.Values["avatar"] = true
	tester.Frame()
	if !tester.HasText("Ada Lovelace") {
		t.Fatal("avatar control did not display an avatar")
	}
	clickExample(t, tester, cfg.String("label"))
	if !tester.HasText("Clicked 1 times") {
		t.Fatal("avatar action was not reported")
	}
	cfg.Values["icon"] = "check"
	tester.Frame()
	if tester.HasText("Ada Lovelace") {
		t.Fatal("leading icon did not replace avatar")
	}
	cfg.Values["icon"], cfg.Values["loading"] = "none", true
	tester.Frame()
	if tester.HasText("Ada Lovelace") {
		t.Fatal("leading loader did not replace avatar")
	}
	cfg.Values["loadingTrailing"] = true
	tester.Frame()
	if !tester.HasText("Ada Lovelace") {
		t.Fatal("trailing loader removed avatar")
	}
	clickExample(t, tester, cfg.String("label"))
	if !tester.HasText("Clicked 1 times") {
		t.Fatal("loading action changed click count")
	}
	cfg.Values["loading"] = false
	tester.Frame()
	clickExample(t, tester, cfg.String("label"))
	if !tester.HasText("Clicked 2 times") {
		t.Fatal("completed loading did not restore action")
	}
}
