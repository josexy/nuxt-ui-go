package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSeparatorPreviewContentAndState(t *testing.T) {
	tester, cfg := exampleTester(t, "separator")
	cfg.Values["avatar"] = true
	tester.Frame()
	if tester.HasText("Separator member") {
		t.Fatal("avatar replaced higher-priority label")
	}
	cfg.Values["label"] = ""
	tester.Frame()
	if r, ok := tester.Find("Separator member"); !ok || r.W != 20 || r.H != 20 {
		t.Fatalf("default separator avatar=%+v", r)
	}
	cfg.Values["editor"] = true
	tester.Frame()
	if tester.HasText("Separator member") {
		t.Fatal("avatar replaced higher-priority content")
	}
	clickExample(t, tester, "Separator draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["position"] = "start"
	tester.Frame()
	if !tester.Focused("Separator draft") {
		t.Fatal("position update discarded editor focus")
	}
	tester.Type("Replaced draft")
	if !tester.HasText("Draft: Replaced draft") {
		t.Fatal("position update discarded editor selection")
	}
	cfg.Values["orientation"], cfg.Values["position"] = "vertical", "end"
	tester.Frame()
	if !tester.Focused("Separator draft") || !tester.HasText("Draft: Replaced draft") {
		t.Fatal("vertical layout discarded the draft or focus")
	}
}
