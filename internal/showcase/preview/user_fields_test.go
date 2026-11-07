package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestUserPreviewActivationChipAndContent(t *testing.T) {
	tester, cfg := exampleTester(t, "user")
	cfg.Values["interactive"], cfg.Values["chip"] = true, true
	tester.Frame()
	if !tester.HasText("3") {
		t.Fatal("chip control did not create a marker")
	}
	clickExample(t, tester, "Ada Lovelace")
	if !tester.HasText("Activated: 1") {
		t.Fatal("activation control did not enable the callback")
	}
	cfg.Values["disabled"] = true
	tester.Frame()
	clickExample(t, tester, "Ada Lovelace")
	tester.Key(0, ui.KeySpace)
	if !tester.HasText("Activated: 1") {
		t.Fatal("disabled preview activated")
	}
	cfg.Values["avatar"], cfg.Values["customBody"] = false, true
	tester.Frame()
	if tester.HasText("3") || tester.HasText("Maintainer") || !tester.HasText("Custom user content") {
		t.Fatal("avatar condition or body override ignored")
	}
}
