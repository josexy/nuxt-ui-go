package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestTabsPreviewManualAccessoriesAndContent(t *testing.T) {
	tester, cfg := exampleTester(t, "tabs")
	cfg.Values["manual"] = true
	cfg.Values["avatar"] = true
	cfg.Values["showBadge"] = true
	cfg.Values["keepMounted"] = true
	tester.Frame()
	if !tester.HasText("Account avatar") || !tester.HasText("0") {
		t.Fatal("tab accessories did not render")
	}
	clickExample(t, tester, "Overview")
	tester.Key(0, ui.KeyDown)
	if !tester.HasText("Selected: overview · Changes: 0") {
		t.Fatal("wrong arrow axis changed selection")
	}
	tester.Key(0, ui.KeyRight)
	if !tester.Focused("Settings") || !tester.HasText("Selected: overview · Changes: 0") {
		t.Fatal("manual focus navigation changed content")
	}
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Project settings") || !tester.HasText("Selected: settings · Changes: 1") {
		t.Fatal("manual Enter did not activate")
	}
	clickExample(t, tester, "Overview")
	clickExample(t, tester, "Tab draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Type("Edited preview")
	clickExample(t, tester, "Reverse tabs")
	if !tester.Focused("Tab draft") {
		t.Fatal("preview reorder did not keep editor focus")
	}
	cfg.Values["content"] = false
	tester.Frame()
	if tester.HasText("Tab draft") || !tester.Focused("Overview") {
		t.Fatal("content=false left panels mounted or focus hidden")
	}
	cfg.Values["content"] = true
	tester.Frame()
	clickExample(t, tester, "Tab draft")
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Key(ui.Ctrl, ui.KeyC)
	if tester.Clipboard() != "Edited preview" {
		t.Fatal("preview content toggle lost application data")
	}
}

func TestTabsPreviewOptionsAndItemDataBindings(t *testing.T) {
	tester, cfg := exampleTester(t, "tabs")
	cfg.Options = map[string]any{"TabsOptions": map[string]any{"Content": false, "UI": map[string]any{"indicator": map[string]any{"Background": "#ff00ff"}}}}
	tester.Frame()
	if tester.HasText("Project overview") {
		t.Fatal("configured Content override did not apply")
	}
	r, _ := tester.Find("Overview")
	if p := tester.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H-4)); p.R != 255 || p.G != 0 || p.B != 255 {
		t.Fatalf("configured indicator color=%v", p)
	}
}
