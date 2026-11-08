package nui_test

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestAvatarGroupItemStylesIncludeOverflow(t *testing.T) {
	seen := 0
	items := []nui.AvatarItem{{ID: "a", Avatar: nui.AvatarOptions{Alt: "A"}}, {ID: "b", Avatar: nui.AvatarOptions{Alt: "B"}}, {ID: "c", Avatar: nui.AvatarOptions{Alt: "C"}}}
	tt := ui.NewTester(func(c *ui.Context) {
		seen = 0
		th := nui.App(c, nui.AppConfig{})
		nui.AvatarGroup(c, th, items, nui.AvatarGroupOptions{Max: nui.Set(2), Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Item {
				seen++
				e.Size(40, 40)
			}
		}})
	}, 240, 100)
	if seen != 3 {
		t.Fatalf("Item style called %d times, want 2 visible + overflow", seen)
	}
	for _, label := range []string{"A", "B", "1 more person"} {
		if r := bounds(t, tt, label); r.W != 40 || r.H != 40 {
			t.Fatalf("item %s style ignored: %+v", label, r)
		}
	}
}

func TestAvatarGroupReorderPreservesIdentity(t *testing.T) {
	items := []nui.AvatarItem{{ID: "a", Avatar: nui.AvatarOptions{Alt: "A"}}, {ID: "b", Avatar: nui.AvatarOptions{Alt: "B"}}}
	for i := range items {
		items[i].Avatar.Style = func(slot nui.Slot, e ui.Element) {
			if slot == nui.Root {
				e.Focusable()
			}
		}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		nui.AvatarGroup(c.Key("group"), nui.App(c, nui.AppConfig{}), items, nui.AvatarGroupOptions{})
	}, 240, 100)
	click(t, tt, "B")
	slices.Reverse(items)
	tt.Frame()
	if !tt.Focused("B") || bounds(t, tt, "B").X >= bounds(t, tt, "A").X {
		t.Fatal("reordering avatars changed focused identity or caller order")
	}
	tt.SetDirection(ui.RightToLeft)
	if !tt.Focused("B") || bounds(t, tt, "B").X <= bounds(t, tt, "A").X {
		t.Fatal("RTL changed focused identity or caller order")
	}
}

func TestAvatarGroupExplicitZeroRing(t *testing.T) {
	zero := false
	parent := ui.Hex("#ff00ff")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(16).Background(parent).Children(func() {
			opts := nui.AvatarGroupOptions{}
			if zero {
				opts.UI = nui.Styles{nui.Item: {BorderWidth: nui.Set(float32(0))}}
			}
			nui.AvatarGroup(c, th, []nui.AvatarItem{{ID: "a", Avatar: nui.AvatarOptions{Alt: "A"}}}, opts)
		})
	}, 100, 100)
	r := bounds(t, tt, "A")
	x, y := int(r.X)-1, int(r.Y+r.H/2)
	if tt.Image().RGBAAt(x, y) != rgba(nui.Light().Background) {
		t.Fatal("default outer ring was not drawn")
	}
	zero = true
	tt.Frame()
	if tt.Image().RGBAAt(x, y) != rgba(parent) {
		t.Fatal("Item.BorderWidth=0 did not remove outer ring")
	}
}
