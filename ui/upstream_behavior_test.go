package nui_test

import (
	"reflect"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

// docs/inputs.md and docs/compatibility.md say that MyGo's ui.Handle.Focus
// takes no options, so focus moved through a Handle always reveals its
// target. This fails once Focus takes options; update the docs with it.
func TestUpstreamHandleFocusTakesNoOptions(t *testing.T) {
	got := reflect.TypeOf((*ui.Handle).Focus)
	if want := reflect.TypeOf(func(*ui.Handle, ...*ui.Context) {}); got != want {
		t.Fatalf("ui.Handle.Focus is now %v: update docs/inputs.md (Focus and callbacks) and docs/compatibility.md (Accessibility and motion)", got)
	}
}

// docs/compatibility.md documents that cancelling a composition over selected
// text loses the selection, and that undo restores it. This fails once MyGo
// keeps the selection; update that section with it.
func TestUpstreamSelectedIMECancellationParity(t *testing.T) {
	for _, kind := range []string{"TextInputBase", "Input", "Textarea", "PinInput"} {
		t.Run(kind, func(t *testing.T) {
			value := "Draft"
			var pin nui.PinState
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				switch kind {
				case "TextInputBase":
					ui.TextInputBase(c, &value).Label("Field")
				case "Input":
					nui.Input(c, th, &value, nui.InputOptions{Label: "Field"})
				case "Textarea":
					nui.Textarea(c, th, &value, nui.TextareaOptions{Label: "Field"})
				case "PinInput":
					nui.PinInput(c, th, &value, &pin, nui.PinInputOptions{Label: "PIN", Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Item {
							e.Label("Field")
						}
					}})
				}
			}, 400, 200)
			click(t, tt, "Field")
			tt.Key(ui.Ctrl, ui.KeyA)
			tt.Compose("lin", 3)
			tt.Compose("", 0)
			if value != "" {
				t.Fatalf("cancelling a composition over a selection now leaves %q: the known issue is gone, update docs/compatibility.md", value)
			}
			tt.Command("undo")
			if value != "Draft" {
				t.Fatalf("undo restored %q", value)
			}
			if kind == "PinInput" {
				return
			}
			tt.Key(0, ui.KeyEnd)
			tt.Compose("lin", 3)
			if value != "Draft" {
				t.Fatalf("composition committed early: %q", value)
			}
			tt.Compose("", 0)
			if value != "Draft" {
				t.Fatalf("cancellation without a selection changed the value to %q", value)
			}
		})
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
