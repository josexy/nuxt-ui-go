package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestContainerResponsivePaddingPreservesEditing(t *testing.T) {
	value, changes := "中文 draft", 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Children(func() {
			nui.Container(c.Key("container"), th, nui.ContainerOptions{}, func() {
				ui.Box(c.Key("marker")).FillWidth().Height(1).Label("Padding marker")
				nui.Input(c.Key("draft"), th, &value, nui.InputOptions{Label: "Draft", OnChange: func(string) { changes++ }}).FillWidth()
			}).Label("Container")
		})
	}, 320, 160)
	click(t, tt, "Draft")
	for _, test := range []struct {
		width   int
		padding float32
	}{{320, 16}, {639, 16}, {640, 24}, {1023, 24}, {1024, 32}, {1600, 32}} {
		tt.SetSize(test.width, 160)
		outer, inner := bounds(t, tt, "Container"), bounds(t, tt, "Padding marker")
		if inner.X-outer.X != test.padding || outer.X+outer.W-inner.X-inner.W != test.padding {
			t.Fatalf("window=%d padding=%g: outer=%+v inner=%+v", test.width, test.padding, outer, inner)
		}
		if outer.W != min(float32(test.width), 1280) || outer.X != (float32(test.width)-outer.W)/2 {
			t.Fatalf("container is not centered and constrained: %+v", outer)
		}
		if value != "中文 draft" || changes != 0 || !tt.Focused("Draft") {
			t.Fatal("responsive layout changed draft, focus or callbacks")
		}
	}
	tt.Type("!")
	if changes != 1 {
		t.Fatal("resizing lost editable focus")
	}
}

func TestContainerPaddingPriorityAndExplicitZero(t *testing.T) {
	opts := nui.ContainerOptions{}
	th := nui.Light().WithComponent(nui.ComponentContainer, nui.ComponentTheme{UI: nui.Styles{nui.Root: {PaddingX: nui.Set(float32(40))}}})
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Window(c, th)
		nui.Container(c.Key("container"), th, opts, func() { ui.Text(c, "Content").FillWidth() }).Label("Container")
	}, 500, 120)
	check := func(want float32) {
		t.Helper()
		tt.Frame()
		outer, inner := bounds(t, tt, "Container"), bounds(t, tt, "Content")
		if inner.X-outer.X != want {
			t.Fatalf("padding=%g, want %g", inner.X-outer.X, want)
		}
	}
	check(40)
	opts.Padding = nui.Set(float32(0))
	check(0)
	opts.Padding = nui.Set(float32(12))
	check(12)
	opts.UI = nui.Styles{nui.Root: {PaddingX: nui.Set(float32(0))}}
	check(0)
	opts.Style = func(slot nui.Slot, e ui.Element) {
		if slot == nui.Root {
			e.PaddingX(7)
		}
	}
	check(7)
}
