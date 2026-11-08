package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSelectPreviewAccessoriesOpenAndCounts(t *testing.T) {
	tester, cfg := exampleTester(t, "select")
	tester.Key(0, ui.KeyTab)
	tester.Type("r")
	if !tester.HasText("Selected: rust · Changes: 1 · Open: false") {
		t.Fatal("closed prefix did not change the preview value once")
	}
	cfg.Values["loading"], cfg.Values["loadingTrailing"] = true, true
	cfg.Values["highlight"], cfg.Values["customNeutral"] = true, true
	tester.Frame()
	if !tester.Focused("Language") || !tester.HasText("Language avatar") {
		t.Fatal("changing accessories lost focus or avatar")
	}
	tester.Key(0, ui.KeySpace)
	tester.Type("t")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Selected: typescript · Changes: 2 · Open: false") {
		t.Fatal("loading selection or controlled closing failed")
	}
	clickExample(t, tester, "Toggle multiple select")
	tester.Type("g")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Selected IDs: [go] · Changes: 1 · Open: true") {
		t.Fatalf("multiple preview selection failed: %v", tester.Texts())
	}
	cfg.Values["customIcons"] = true
	tester.Frame()
	tester.Key(0, ui.KeyEscape)
	if !tester.Focused("Multiple languages") || !tester.HasText("Selected IDs: [go] · Changes: 1 · Open: false") {
		t.Fatalf("custom icons changed selection or closed focus: focused=%v texts=%v", tester.Focused("Multiple languages"), tester.Texts())
	}
	if tester.HasText("Language avatar") {
		t.Fatal("leading icon did not replace both avatars")
	}
	clickExample(t, tester, "Toggle single select")
	tester.Key(0, ui.KeyEscape)
	if !tester.Focused("Language") {
		t.Fatal("controlled single open did not return focus to trigger")
	}
	cfg.Values["disabled"] = true
	tester.Frame()
	clickExample(t, tester, "Toggle single select")
	if !tester.HasText("Selected: typescript · Changes: 2 · Open: false") {
		t.Fatal("disabled external open button remained active")
	}
}
