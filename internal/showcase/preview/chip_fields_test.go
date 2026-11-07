package preview

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestChipPreviewHostVisibilityAndStandalone(t *testing.T) {
	tester, cfg := exampleTester(t, "chip")
	host := "Notifications " + cfg.String("label")
	clickExample(t, tester, host)
	cfg.Values["hidden"] = true
	tester.Frame()
	if !tester.Focused(host) {
		t.Fatal("hiding the chip discarded host focus")
	}
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Host actions: 2") {
		t.Fatal("host action did not remain active")
	}
	cfg.Values["hidden"], cfg.Values["standalone"], cfg.Values["styledText"] = false, true, true
	cfg.Values["label"] = "999+"
	tester.Frame()
	if tester.HasText(host) || !tester.HasText("999+") || !tester.HasText("Host actions: 2") {
		t.Fatal("standalone chip did not replace host or preserve action count")
	}
	cfg.Values["hidden"] = true
	tester.Frame()
	if tester.HasText("999+") {
		t.Fatal("hidden standalone chip is still visible")
	}
}
