package preview

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestTooltipPreviewTimingStateAndEffectiveOptions(t *testing.T) {
	tester, cfg := exampleTester(t, "tooltip")
	cfg.Values["customDelay"], cfg.Values["delay"] = true, float64(1000)
	tester.Frame()
	r, _ := tester.Find("Hover or focus")
	tester.Move(r.X+4, r.Y+4)
	tester.Advance(500 * time.Millisecond)
	if !tester.HasText("Tooltip open: false") {
		t.Fatal("custom delay opened too early")
	}
	tester.Advance(510 * time.Millisecond)
	if !tester.HasText("Tooltip open: true") || !tester.HasText("Search this workspace") {
		t.Fatal("custom delay did not open preview")
	}
	tester.Key(0, ui.KeyEscape)
	if !tester.HasText("Tooltip open: false") {
		t.Fatal("Escape did not publish closed state")
	}
	tester.Move(740, 340)
	clickExample(t, tester, "Toggle tooltip")
	if !tester.HasText("Tooltip open: true") {
		t.Fatal("toggle did not control tooltip")
	}
	cfg.Values["label"], cfg.Values["maxWidth"], cfg.Values["styled"] = "", float64(160), true
	tester.Frame()
	if !tester.HasText("K") || tester.HasText("·") {
		t.Fatal("keys-only content is incorrect")
	}
	cfg.Values["disabled"] = true
	tester.Frame()
	if !tester.HasText("Tooltip open: false") || tester.HasText("K") {
		t.Fatal("disabled tooltip stayed visible")
	}
	cfg.Values["disabled"] = false
	cfg.Options = map[string]any{"TooltipOptions": map[string]any{"Open": nil, "DelayDuration": float64(0)}}
	tester.Frame()
	if !tester.HasText("Uncontrolled tooltip") {
		t.Fatal("cleared Open still displayed the unused default pointer")
	}
	r, _ = tester.Find("Hover or focus")
	tester.Move(r.X+4, r.Y+4)
	if !tester.HasText("K") {
		t.Fatal("explicit zero-delay override did not apply")
	}
}
