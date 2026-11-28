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

func TestCardResponsivePaddingKeepsContent(t *testing.T) {
	value := "Persistent draft"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Card(c.Key("card"), th, nui.CardOptions{Title: "Title", Description: "Description", Footer: func() { ui.Text(c, "Footer text") }, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Header || slot == nui.Body || slot == nui.Footer {
				e.Label(string(slot))
			}
		}}, func() {
			ui.Box(c.Key("marker")).Height(1).FillWidth().Label("Body marker")
			nui.Input(c.Key("draft"), th, &value, nui.InputOptions{Label: "Draft"})
		}).FillWidth()
	}, 320, 320)
	click(t, tt, "Draft")
	for _, width := range []int{320, 639, 640, 1024} {
		tt.SetSize(width, 320)
		want := float32(16)
		if width >= 640 {
			want = 24
		}
		for _, entry := range []struct{ part, text string }{{"header", "Title"}, {"body", "Body marker"}, {"footer", "Footer text"}} {
			outer, inner := bounds(t, tt, entry.part), bounds(t, tt, entry.text)
			if inner.X-outer.X != want {
				t.Fatalf("width=%d %s horizontal padding=%g, want %g", width, entry.part, inner.X-outer.X, want)
			}
			if entry.part == "body" && inner.Y-outer.Y != want {
				t.Fatalf("width=%d body vertical padding=%g, want %g", width, inner.Y-outer.Y, want)
			}
		}
		if value != "Persistent draft" || !tt.Focused("Draft") {
			t.Fatal("responsive card reset draft or focus")
		}
	}
}

func TestCardDescriptionOnlyKeepsTopMargin(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Card(c, th, nui.CardOptions{Description: "Description", Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Header {
				e.Label("Header")
			}
		}}, nil)
	}, 320, 100)
	if got := bounds(t, tt, "Description").Y - bounds(t, tt, "Header").Y; got != 20 {
		t.Fatalf("description-only top offset=%g, want 16+4", got)
	}
}

func TestCardHeaderRemovalPreservesBodyEditor(t *testing.T) {
	showHeader, value, changes := true, "Persistent draft", 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.CardOptions{Footer: func() { ui.Text(c, "Footer") }}
		if showHeader {
			opts.Title = "Header"
		}
		nui.Card(c.Key("card"), th, opts, func() {
			nui.Input(c.Key("draft"), th, &value, nui.InputOptions{Label: "Draft", OnChange: func(string) { changes++ }})
		})
	}, 320, 240)
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	showHeader = false
	tt.Frame()
	if !tt.Focused("Draft") {
		t.Fatal("removing header discarded body focus")
	}
	tt.Type("Replacement")
	if value != "Replacement" || changes != 1 {
		t.Fatal("header removal discarded editor selection")
	}
	showHeader = true
	tt.Frame()
	if !tt.Focused("Draft") {
		t.Fatal("restoring header discarded body focus")
	}
}

func TestSeparatorAvatarDefaultAndLeadingStyle(t *testing.T) {
	avatar := nui.AvatarOptions{Alt: "Separator avatar"}
	leading := false
	tt := ui.NewTester(func(c *ui.Context) {
		leading = false
		nui.Separator(c, nui.App(c, nui.AppConfig{}), nui.SeparatorOptions{Avatar: &avatar, Style: func(slot nui.Slot, _ ui.Element) {
			if slot == nui.Leading {
				leading = true
			}
		}})
	}, 320, 100)
	if r := bounds(t, tt, "Separator avatar"); r.W != 20 || r.H != 20 || !leading {
		t.Fatalf("default avatar or Leading style mismatch: bounds=%+v leading=%v", r, leading)
	}
	avatar.Size = nui.LG
	tt.Frame()
	if r := bounds(t, tt, "Separator avatar"); r.W != 36 || r.H != 36 {
		t.Fatalf("explicit avatar size ignored: %+v", r)
	}
}

func TestSeparatorPositionKeepsContentFocus(t *testing.T) {
	value, changes := "Persistent draft", 0
	position, orientation := nui.SeparatorCenter, nui.Horizontal
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Separator(c.Key("separator"), th, nui.SeparatorOptions{Position: position, Orientation: orientation, Content: func() {
			nui.Input(c.Key("draft"), th, &value, nui.InputOptions{Label: "Separator draft", OnChange: func(string) { changes++ }}).Width(160)
		}}).Height(160)
	}, 400, 220)
	click(t, tt, "Separator draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	for _, next := range []nui.SeparatorPosition{nui.SeparatorStart, nui.SeparatorEnd, nui.SeparatorCenter} {
		position = next
		tt.Frame()
		if !tt.Focused("Separator draft") {
			t.Fatalf("position %d discarded content focus", next)
		}
	}
	orientation = nui.Vertical
	tt.Frame()
	tt.Type("Replacement")
	if value != "Replacement" || changes != 1 || !tt.Focused("Separator draft") {
		t.Fatal("layout change discarded selection or repeated change notification")
	}
}

func TestSeparatorVerticalLongLabelWraps(t *testing.T) {
	label := "这是一段需要在窄容器中换行的分隔说明文字"
	tt := ui.NewTester(func(c *ui.Context) {
		nui.Separator(c, nui.App(c, nui.AppConfig{}), nui.SeparatorOptions{Label: label, Orientation: nui.Vertical}).Size(100, 200).Label("Separator")
	}, 200, 250)
	root, text := bounds(t, tt, "Separator"), bounds(t, tt, label)
	if text.W > root.W || text.H <= 20 || text.X < root.X || text.X+text.W > root.X+root.W {
		t.Fatalf("vertical label did not wrap within root: root=%+v text=%+v", root, text)
	}
}
