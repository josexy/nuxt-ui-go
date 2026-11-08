package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAlertPreviewActionsCloseAndLayout(t *testing.T) {
	tester, cfg := exampleTester(t, "alert")
	clickExample(t, tester, "Review changes")
	if !tester.HasText("Reviewed 1 times; closed 0 times") {
		t.Fatal("action count did not update")
	}
	cfg.Values["avatar"], cfg.Values["orientation"] = true, "horizontal"
	tester.Frame()
	if !tester.HasText("Ada Lovelace") || !tester.Focused("Review changes") {
		t.Fatal("avatar/direction change did not preserve action focus")
	}
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Reviewed 2 times; closed 0 times") {
		t.Fatal("retained action focus did not activate once")
	}
	label := "Dismiss " + cfg.String("label")
	cfg.Values["closeDisabled"] = true
	tester.Frame()
	clickExample(t, tester, label)
	if !tester.HasText("Review changes") {
		t.Fatal("disabled close hid notice")
	}
	cfg.Values["closeDisabled"], cfg.Values["customCloseIcon"] = false, true
	tester.Frame()
	clickExample(t, tester, label)
	if tester.HasText("Review changes") || !tester.HasText("Reviewed 2 times; closed 1 times") {
		t.Fatal("close did not hide notice and notify once")
	}
	clickExample(t, tester, "Reopen notice")
	if !tester.HasText("Review changes") || !tester.HasText("Reviewed 2 times; closed 1 times") {
		t.Fatal("reopen discarded action state")
	}
	cfg.Values["closable"] = false
	tester.Frame()
	if tester.HasText(label) {
		t.Fatal("disabling close did not remove close button")
	}
}
