package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestSplitterPointerHighlightClearsAfterDrag(t *testing.T) {
	theme := nui.Light()
	size, changes := float32(120), 0
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Splitter(c.Key("splitter"), theme, &size, nui.SplitterOptions{
			Label:    "Panel divider",
			Max:      nui.Set(float32(180)),
			OnChange: func(float32) { changes++ },
		}, nil, nil).Fill()
	}, 320, 120)

	checkHighlight := func(stage string, want bool) {
		t.Helper()
		r := bounds(t, tt, "Panel divider")
		if got := strokePixels(tt.Image(), r, 1, true, rgba(theme.Primary)) > 0; got != want {
			t.Fatalf("%s: highlighted = %t, want %t", stage, got, want)
		}
	}
	checkHighlight("idle", false)
	r := bounds(t, tt, "Panel divider")
	x, y := r.X+r.W/2, r.Y+r.H/2
	tt.Move(x, y)
	checkHighlight("hover", true)
	tt.Press(x, y)
	tt.Move(x+100, y)
	checkHighlight("held beyond the resize limit", true)
	tt.Release(x+100, y)
	checkHighlight("released outside the drag target", false)
	if size != 180 || changes != 1 {
		t.Fatalf("drag size = %g, changes = %d", size, changes)
	}

	tt.Key(0, ui.KeyLeft)
	if size != 170 || changes != 2 {
		t.Fatalf("keyboard resize after drag: size = %g, changes = %d", size, changes)
	}
	tt.ClickAt(x+100, y)
	tt.Key(0, ui.KeyTab)
	checkHighlight("keyboard focus with pointer outside", true)
}
