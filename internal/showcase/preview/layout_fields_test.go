package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestContainerPreviewDefaultAndExplicitPadding(t *testing.T) {
	tester, cfg := exampleTester(t, "container")
	check := func(want float32) {
		t.Helper()
		tester.Frame()
		outer, outerOK := tester.Find("Container preview")
		inner, innerOK := tester.Find("Content boundary")
		if !outerOK || !innerOK || inner.X-outer.X != want {
			t.Fatalf("padding=%g, want %g", inner.X-outer.X, want)
		}
	}
	check(24)
	clickExample(t, tester, "Container draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Type("Edited draft")
	cfg.Values["defaultPadding"], cfg.Values["padding"] = false, float64(0)
	check(0)
	cfg.Values["defaultPadding"] = true
	tester.SetSize(500, 360)
	check(16)
	if !tester.Focused("Container draft") {
		t.Fatal("padding update discarded input focus")
	}
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Command("copy")
	if tester.Clipboard() != "Edited draft" {
		t.Fatalf("padding update discarded the draft: %q", tester.Clipboard())
	}
}

func TestCardPreviewResponsiveHeaderAndBody(t *testing.T) {
	tester, cfg := exampleTester(t, "card")
	cfg.Values["label"] = ""
	tester.SetSize(500, 360)
	tester.Frame()
	desc, ok := tester.Find(cfg.String("description"))
	if !ok || desc.X != 40 || desc.Y != 44 {
		t.Fatalf("description-only narrow layout=%+v", desc)
	}
	cfg.Values["description"] = ""
	tester.Frame()
	body, ok := tester.Find("A focused space for related content.")
	if !ok || body.X != 40 || body.Y != 40 {
		t.Fatalf("empty header retained spacing: %+v", body)
	}
	tester.SetSize(760, 360)
	body, _ = tester.Find("A focused space for related content.")
	if body.X != 48 || body.Y != 48 {
		t.Fatalf("wide body padding=%+v", body)
	}
}

func TestCardPreviewEditingSurvivesHeaderRemoval(t *testing.T) {
	tester, cfg := exampleTester(t, "card")
	clickExample(t, tester, "Card draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["label"], cfg.Values["description"] = "", ""
	tester.Frame()
	if !tester.Focused("Card draft") {
		t.Fatal("clearing the header reset editor focus")
	}
	tester.Type("Replaced draft")
	if !tester.HasText("Draft: Replaced draft") {
		t.Fatal("clearing the header reset the selection")
	}
	cfg.Values["label"] = "Restored header"
	tester.Frame()
	if !tester.Focused("Card draft") {
		t.Fatal("restoring the header reset focus")
	}
}
