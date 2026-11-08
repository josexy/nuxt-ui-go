package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestDropdownMenuPreviewFilterAndCheckedActions(t *testing.T) {
	tester, cfg := exampleTester(t, "dropdown-menu")
	cfg.Values["filter"] = true
	cfg.Values["subFilter"] = true
	cfg.Values["avatar"] = true
	cfg.Values["customIcons"] = true
	tester.Frame()
	clickExample(t, tester, "Actions")
	if !tester.Focused("Action search") || !tester.HasText("Editor avatar") {
		t.Fatal("filter or avatar control did not apply")
	}
	tester.Type("copy")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Duplicate") || tester.HasText("Project actions") {
		t.Fatal("filter Enter did not select the matching action")
	}
	clickExample(t, tester, "Actions")
	tester.Key(0, ui.KeyEscape)
	clickExample(t, tester, "Pinned")
	if !tester.HasText("Actions: 2 · Checked changes: 1 · Open changes: 3") || !tester.HasText("Project actions") {
		t.Fatal("checked keep-open action did not update counts")
	}
	clickExample(t, tester, "Reverse actions")
	tester.Key(0, ui.KeyEnter)
	if !tester.Focused("Reverse actions") {
		t.Fatal("reordered action lost focus")
	}
	clickExample(t, tester, "Export")
	if !tester.Focused("Export search") {
		t.Fatal("nested filter did not open")
	}
	tester.Type("json")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("JSON") || !tester.Focused("Actions") {
		t.Fatal("submenu selection did not return to trigger")
	}
}

func TestDropdownMenuPreviewOptionsBoundsAndSize(t *testing.T) {
	tester, cfg := exampleTester(t, "dropdown-menu")
	cfg.Options = map[string]any{"MenuOptions": map[string]any{"Size": "nui.XS", "Width": float64(210), "Offset": float64(0), "Filter": true, "FilterInput": map[string]any{"Label": "Configured search"}}}
	tester.Frame()
	clickExample(t, tester, "Actions")
	panel, _ := tester.Find("Project actions")
	anchor, _ := tester.Find("Actions")
	row, _ := tester.Find("Edit")
	if panel.W != 210 || panel.Y != anchor.Y+anchor.H || row.H != 40 || !tester.Focused("Configured search") {
		t.Fatalf("options not reflected: panel=%v row=%v anchor=%v", panel, row, anchor)
	}
	cfg.Options["MenuOptions"].(map[string]any)["Disabled"] = true
	tester.Frame()
	if tester.HasText("Project actions") {
		t.Fatal("configured disabled state left the menu open")
	}
}
