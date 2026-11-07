package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSliderPreviewDecimalCommitAndDirection(t *testing.T) {
	tester, cfg := exampleTester(t, "slider")
	cfg.Values["step"] = .1
	tester.Frame()
	tester.Key(0, ui.KeyTab)
	tester.Key(0, ui.KeyRight)
	if !tester.HasText("Value: 40.1 · Changes: 1 · Commits: 1") {
		t.Fatalf("single slider controls or callbacks were not applied: %v", tester.Texts())
	}
	tester.Key(0, ui.KeyTab)
	tester.Key(0, ui.KeyRight)
	if !tester.HasText("Range: 20.1 – 75 · Changes: 1 · Commits: 1") {
		t.Fatal("range callbacks were not applied")
	}
	cfg.Values["vertical"] = true
	tester.Frame()
	if !tester.Focused("Range minimum") {
		t.Fatal("changing orientation lost range focus")
	}
	tester.Key(0, ui.KeyUp)
	if !tester.HasText("Range: 20.2 – 75 · Changes: 2 · Commits: 2") {
		t.Fatal("vertical range was not connected")
	}
	cfg.Values["parentDisabled"] = true
	tester.Frame()
	tester.Key(0, ui.KeyUp)
	if !tester.HasText("Range: 20.2 – 75 · Changes: 2 · Commits: 2") {
		t.Fatal("disabled preview parent did not block range editing")
	}
}
