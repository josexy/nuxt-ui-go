package nui_test

import (
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

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
