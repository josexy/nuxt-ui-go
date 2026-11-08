package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

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
