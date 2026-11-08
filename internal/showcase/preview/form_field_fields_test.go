package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestFormFieldPreviewLayoutAndErrorsPreserveEditing(t *testing.T) {
	tester, cfg := exampleTester(t, "form-field")
	tester.Key(0, ui.KeyTab)
	if !tester.Focused("Project") {
		t.Fatal("preview input was not reachable by Tab")
	}
	tester.Type("Original")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["label"], cfg.Values["horizontal"], cfg.Values["regionStyles"] = "", true, true
	tester.Frame()
	if !tester.Focused("Project input") {
		t.Fatal("removing the preview label or changing regions lost focus")
	}
	tester.Type("Updated")
	if !tester.HasText("Value: Updated") {
		t.Fatal("field preview discarded its selection")
	}
	cfg.Values["error"] = "Preview error"
	tester.Frame()
	if !tester.HasText("Preview error") || tester.HasText("Use a name your team will recognize.") {
		t.Fatal("field preview did not replace help with the error")
	}
}

func TestFieldGroupPreviewReorderAndAction(t *testing.T) {
	tester, cfg := exampleTester(t, "field-group")
	clickExample(t, tester, "Search text")
	tester.Type("Original")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["vertical"], cfg.Values["reversed"] = true, true
	tester.Frame()
	if !tester.Focused("Search text") {
		t.Fatal("group controls discarded editor focus")
	}
	tester.Type("Updated")
	clickExample(t, tester, "Search")
	if !tester.HasText("Submitted: Updated") {
		t.Fatal("group preview did not submit the selected replacement")
	}
}
