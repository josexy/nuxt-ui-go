package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// This is an upstream parity probe, NOT an acceptance of lossless cancellation.
// Replacing a selection with composition and cancelling keeps the deletion.
// Keeping this executable makes an upstream upgrade's behavior change visible.
func TestUpstreamSelectedIMECancellationParity(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		value := "Draft"
		tt := ui.NewTester(func(c *ui.Context) {
			theme := nui.App(c, nui.AppConfig{})
			if wrapped {
				nui.Input(c, theme, &value, nui.InputOptions{Label: "Name"})
			} else {
				ui.TextInputBase(c, &value).Label("Name")
			}
		}, 400, 100)
		click(t, tt, "Name")
		tt.Key(ui.Ctrl, ui.KeyA)
		tt.Compose("lin", 3)
		tt.Compose("", 0)
		if value != "" {
			t.Fatalf("wrapped=%v upstream behavior changed to %q; reassess selected-text IME cancellation", wrapped, value)
		}
		tt.Command("undo")
		if value != "Draft" {
			t.Fatalf("wrapped=%v undo could not restore selected text", wrapped)
		}
		tt.Key(0, ui.KeyEnd)
		tt.Compose("lin", 3)
		if value != "Draft" {
			t.Fatalf("wrapped=%v composition unexpectedly changed value to %q", wrapped, value)
		}
		tt.Compose("", 0)
		if value != "Draft" {
			t.Fatal("cancellation without a selection lost text")
		}
	}
}

func TestNestedScrollAreasAndHorizontalScroll(t *testing.T) {
	var outer, inner, horizontal ui.ScrollState
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		nui.ScrollArea(c, theme, nui.ScrollAreaOptions{State: &outer}, func() {
			nui.ScrollArea(c, theme, nui.ScrollAreaOptions{State: &inner}, func() {
				for i := range 30 {
					ui.Textf(c, "Inner row %d", i)
				}
			}).Height(120).Label("Inner scroller")
			nui.ScrollArea(c, theme, nui.ScrollAreaOptions{Direction: nui.ScrollHorizontal, State: &horizontal}, func() { ui.Text(c, "Wide content").Width(1200) }).Height(50).Label("Horizontal scroller")
			ui.Box(c).Height(1200)
		}).Fill()
	}, 400, 350)
	r := bounds(t, tt, "Inner scroller")
	tt.Scroll(r.X+100, r.Y+60, 0, 80)
	if inner.Y <= 0 || outer.Y != 0 {
		t.Fatalf("nested offsets inner=%g outer=%g", inner.Y, outer.Y)
	}
	r = bounds(t, tt, "Horizontal scroller")
	tt.Scroll(r.X+100, r.Y+20, 100, 0)
	if horizontal.X <= 0 {
		t.Fatal("horizontal scroll did not move")
	}
	tt.Scroll(350, 300, 0, 100)
	if outer.Y <= 0 {
		t.Fatal("outer scroll did not move")
	}
}
