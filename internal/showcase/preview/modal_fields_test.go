package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestModalPreviewPoliciesAndHeaderChanges(t *testing.T) {
	tester, cfg := exampleTester(t, "modal")
	cfg.Values["transition"], cfg.Values["overlay"], cfg.Values["dismissible"], cfg.Values["customClose"] = false, false, false, true
	tester.Frame()
	clickExample(t, tester, "Open dialog")
	if !tester.HasText("Configured dialog close") {
		t.Fatal("custom close was not applied")
	}
	tester.Key(0, ui.KeyEscape)
	if !tester.HasText("Project name") {
		t.Fatal("non-dismissible dialog closed")
	}
	clickExample(t, tester, "Project name")
	tester.Type("Retained draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["headerless"], cfg.Values["fullscreen"], cfg.Values["longBody"] = true, true, true
	tester.Frame()
	if !tester.Focused("Project name") || tester.HasText("Configured dialog close") {
		t.Fatal("header removal lost focus or kept the close button")
	}
	tester.Type("Selected replacement")
	clickExample(t, tester, "Save")
	if !tester.HasText("Close notifications: 0 · Prevented: 1") {
		t.Fatal("prevented/application-close counts are incorrect")
	}
	cfg.Values["headerless"], cfg.Values["dismissible"], cfg.Values["fullscreen"] = false, true, false
	tester.Frame()
	clickExample(t, tester, "Open dialog")
	clickExample(t, tester, "Project name")
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Key(ui.Ctrl, ui.KeyC)
	if tester.Clipboard() != "Selected replacement" {
		t.Fatalf("persistent draft or selection was lost: %q", tester.Clipboard())
	}
	clickExample(t, tester, "Configured dialog close")
	if !tester.Focused("Open dialog") || !tester.HasText("Close notifications: 1 · Prevented: 1") {
		t.Fatal("custom close did not restore focus/count once")
	}
}
