package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestInputPreviewAccessoriesAndCallbacks(t *testing.T) {
	tester, cfg := exampleTester(t, "input")
	clickExample(t, tester, "Project name")
	tester.Type("Original")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["avatar"], cfg.Values["loading"], cfg.Values["loadingTrailing"] = true, true, true
	tester.Frame()
	if !tester.HasText("Input preview avatar") || !tester.Focused("Project name") {
		t.Fatal("input preview failed to preserve its avatar or focus")
	}
	tester.Type("Updated")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Value: Updated") || !tester.HasText("Changes: 2 · Submits: 1 · Clears: 0") {
		t.Fatal("input preview values or callbacks are not connected")
	}
	clickExample(t, tester, "Clear Project name")
	if !tester.HasText("Changes: 3 · Submits: 1 · Clears: 1") || !tester.Focused("Project name") {
		t.Fatal("input preview clear failed")
	}
	cfg.Values["loadingTrailing"] = false
	tester.Frame()
	if tester.HasText("Input preview avatar") {
		t.Fatal("leading loader did not replace the preview avatar")
	}
}

func TestTextareaPreviewRowsAccessoriesAndClear(t *testing.T) {
	tester, cfg := exampleTester(t, "textarea")
	clickExample(t, tester, "Project name")
	tester.Type("First\nSecond\nThird\nFourth\nFifth")
	cfg.Values["rows"], cfg.Values["maxRows"] = float64(2), float64(4)
	cfg.Values["autoResize"], cfg.Values["avatar"], cfg.Values["loading"], cfg.Values["loadingTrailing"] = true, true, true, true
	tester.Frame()
	if !tester.HasText("Textarea preview avatar") || !tester.Focused("Project name") {
		t.Fatal("textarea controls lost the avatar or editor focus")
	}
	before, ok := tester.Find("Project name")
	if !ok {
		t.Fatal("missing textarea editor")
	}
	cfg.Values["maxRows"] = float64(2)
	tester.Frame()
	after, _ := tester.Find("Project name")
	if after.H >= before.H {
		t.Fatal("maximum row control did not constrain the editor")
	}
	clickExample(t, tester, "Clear Project name")
	if !tester.HasText("Changes: 2 · Clears: 1") || !tester.Focused("Project name") {
		t.Fatal("textarea clear control is not connected")
	}
}
