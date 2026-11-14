package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestFieldGroupReorderOrientationAndDisabledPreserveDraft(t *testing.T) {
	value, reversed, vertical, disabled := "Original", false, false, false
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		items := []nui.FieldGroupItem{{ID: "draft", Build: func(c *ui.Context, t nui.Theme) ui.Element {
			return nui.Input(c, t, &value, nui.InputOptions{Label: "Group draft"}).Width(160)
		}}, {ID: "action", Build: func(c *ui.Context, t nui.Theme) ui.Element {
			return nui.Button(c, t, "Apply", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { calls++ })
		}}}
		if reversed {
			items[0], items[1] = items[1], items[0]
		}
		nui.FieldGroup(c.Key("group"), th, nui.FieldGroupOptions{Vertical: vertical, Disabled: disabled}, items).Width(240)
	}, 300, 160)
	click(t, tt, "Group draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	reversed, vertical = true, true
	tt.Frame()
	if !tt.Focused("Group draft") || bounds(t, tt, "Apply").Y >= bounds(t, tt, "Group draft").Y {
		t.Fatal("reorder or orientation did not preserve focus and item order")
	}
	tt.Type("Replacement")
	if value != "Replacement" {
		t.Fatal("group layout changes discarded the selection")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Apply")
	click(t, tt, "Group draft")
	tt.Type("Blocked")
	if calls != 0 || value != "Replacement" {
		t.Fatal("group disabled state did not reach its children")
	}
	disabled = false
	tt.Frame()
	click(t, tt, "Apply")
	tt.Key(0, ui.KeyEnter)
	if calls != 2 {
		t.Fatal("group action did not resume after enabling")
	}
}
