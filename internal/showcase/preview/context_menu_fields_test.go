package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestContextMenuPreviewPointerStateAndDisabledTarget(t *testing.T) {
	tester, cfg := exampleTester(t, "context-menu")
	target, _ := tester.Find("Right click here")
	tester.RightClickAt(target.X+target.W-4, target.Y+4)
	first, ok := tester.Find("Project actions")
	if !ok {
		t.Fatal("pointer menu did not open")
	}
	tester.RightClickAt(target.X+4, target.Y+target.H-4)
	second, ok := tester.Find("Project actions")
	if !ok || first == second {
		t.Fatal("second right click did not reposition menu")
	}
	clickExample(t, tester, "Move target")
	after, _ := tester.Find("Project actions")
	if after != second || !tester.HasText("Target clicks: 0 · Moved: true") {
		t.Fatal("moving the target moved the pointer popup")
	}
	tester.Key(0, ui.KeyEscape)
	clickExample(t, tester, "Open at target")
	target, _ = tester.Find("Right click here")
	opened, _ := tester.Find("Project actions")
	if opened.X != target.X+target.W+2 || opened.Y != target.Y {
		t.Fatalf("controlled anchor: target=%v panel=%v", target, opened)
	}
	cfg.Values["disabledMenu"] = true
	tester.Frame()
	clickExample(t, tester, "Right click here")
	if tester.HasText("Project actions") || !tester.HasText("Target clicks: 1 · Moved: true") {
		t.Fatal("disabled menu blocked the target action")
	}
}

func TestContextMenuPreviewSizeAndFilterExtension(t *testing.T) {
	tester, cfg := exampleTester(t, "context-menu")
	cfg.Values["size"] = "xs"
	cfg.Values["avatar"] = true
	cfg.Values["filter"] = true
	tester.Frame()
	if err := tester.RightClick("Right click here"); err != nil {
		t.Fatal(err)
	}
	avatar, ok := tester.Find("Editor avatar")
	if !ok || avatar.W != 16 || !tester.Focused("Action search") {
		t.Fatal("context size/avatar/filter did not apply")
	}
	tester.Type("copy")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Duplicate") || !tester.Focused("Right click here") {
		t.Fatal("filtered context action did not select and restore focus")
	}
}
