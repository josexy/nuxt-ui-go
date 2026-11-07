package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestInputMenuPreviewClearAndAutocomplete(t *testing.T) {
	tester, cfg := exampleTester(t, "input-menu")
	cfg.Values["label"] = "Language search"
	cfg.Values["multiple"], cfg.Values["loading"] = true, true
	tester.Frame()
	clickExample(t, tester, "Language search")
	if tester.HasText("Backend services") {
		t.Fatal("default click opened the preview")
	}
	tester.Type("Go")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Query:  · Selected: go · IDs: [go]") || !tester.HasText("Events: query 1 · select 1 · values 1 · clear 0 · submit 0") {
		t.Fatalf("multiple preview selection: %v", tester.Texts())
	}
	clickExample(t, tester, "Clear Language search")
	if !tester.HasText("Query:  · Selected:  · IDs: []") || !tester.HasText("Events: query 1 · select 1 · values 2 · clear 1 · submit 0") {
		t.Fatal("empty-draft clear failed or reported input edit")
	}
	cfg.Values["mode"] = "autocomplete"
	cfg.Values["resetOnSelect"] = false
	tester.Frame()
	tester.Type("Ru")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Query: Rust · Selected: rust · IDs: []") {
		t.Fatal("autocomplete did not fill the selected suggestion")
	}
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Type("Custom text")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Events: query 3 · select 2 · values 2 · clear 1 · submit 1") {
		t.Fatal("free-text submit events failed")
	}
	cfg.Values["readOnly"] = true
	tester.Frame()
	clickExample(t, tester, "Toggle search list")
	if !tester.HasText("Open: false · Create request: ") {
		t.Fatal("read-only input accepted external opening")
	}
}

func TestInputMenuPreviewCreateRetainsDraft(t *testing.T) {
	tester, cfg := exampleTester(t, "input-menu")
	cfg.Values["label"] = "Language search"
	cfg.Values["create"] = true
	cfg.Values["createAtTop"] = true
	tester.Frame()
	clickExample(t, tester, "Language search")
	tester.Type("New language")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Open: true · Create request: New language") || !tester.HasText("Query: New language · Selected:  · IDs: []") {
		t.Fatal("create did not retain draft and selection")
	}
}
