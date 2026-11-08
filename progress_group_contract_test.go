package nui_test

import (
	"image/color"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestProgressGroupNeutralUsesInvertedBackground(t *testing.T) {
	color := ui.Hex("#7e22ce")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.InvertedBackground = color
		nui.ProgressGroup(c, th, []nui.ProgressSegment{{ID: "first", Label: "First", Value: 50, Color: nui.Neutral}}, nui.ProgressGroupOptions{Label: "Group"}).Width(200)
	}, 240, 100)
	r := bounds(t, tt, "Group")
	if got := tt.Image().RGBAAt(int(r.X)+30, int(r.Y)+4); got != rgba(color) {
		t.Fatalf("Neutral segment ignored InvertedBackground: %v", got)
	}
}

func TestProgressGroupIconAndItemStyles(t *testing.T) {
	showIcon := true
	size := nui.XS
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		segment := nui.ProgressSegment{ID: "first", Value: 50, UI: nui.Styles{nui.Segment: {Background: nui.Set(ui.Hex("#dc2626"))}}}
		if showIcon {
			segment.Icon = lucide.Check()
		}
		nui.ProgressGroup(c, th, []nui.ProgressSegment{segment}, nui.ProgressGroupOptions{Label: "Group", Size: size, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Leading {
				e.Label("Legend marker")
			}
		}}).Width(200)
	}, 240, 140)
	if r := bounds(t, tt, "Legend marker"); r.W != 12 || r.H != 12 {
		t.Fatalf("icon-only item did not create an XS legend: %+v", r)
	}
	r := bounds(t, tt, "Group")
	if got := tt.Image().RGBAAt(int(r.X)+30, int(r.Y)+1); got != rgba(ui.Hex("#dc2626")) {
		t.Fatalf("per-segment fill style ignored: %v", got)
	}
	size = nui.XL
	tt.Frame()
	if r := bounds(t, tt, "Legend marker"); r.W != 20 || r.H != 20 {
		t.Fatalf("XL icon size: %+v", r)
	}
	showIcon = false
	tt.Frame()
	if tt.HasText("Legend marker") {
		t.Fatal("empty item left an unnecessary legend")
	}
}

func TestProgressGroupBuilderKeepsFocusAcrossReorderAndOrientation(t *testing.T) {
	draft := "Original"
	items := []nui.ProgressSegment{
		{ID: "draft", Value: 50, Build: func(c *ui.Context, th nui.Theme) {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Segment draft"}).Width(140)
		}},
		{ID: "other", Label: "Other", Value: 20},
	}
	orientation := nui.Horizontal
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ProgressGroup(c, nui.App(c, nui.AppConfig{}), items, nui.ProgressGroupOptions{Orientation: orientation}).Width(240)
	}, 300, 240)
	click(t, tt, "Segment draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	slices.Reverse(items)
	orientation = nui.Vertical
	tt.Frame()
	if !tt.Focused("Segment draft") {
		t.Fatal("segment reorder or orientation discarded focus")
	}
	tt.Type("Replacement")
	if draft != "Replacement" {
		t.Fatalf("segment reorder discarded selection: %q", draft)
	}
}

func TestProgressGroupLegendFollowsSize(t *testing.T) {
	size := nui.XS
	tt := ui.NewTester(func(c *ui.Context) {
		nui.ProgressGroup(c, nui.App(c, nui.AppConfig{}), []nui.ProgressSegment{{ID: "first", Label: "First", Value: 50}}, nui.ProgressGroupOptions{Size: size}).Width(200)
	}, 240, 100)
	small := bounds(t, tt, "First")
	size = nui.XL
	tt.Frame()
	large := bounds(t, tt, "First")
	if large.H <= small.H {
		t.Fatalf("legend size unchanged: XS=%+v XL=%+v", small, large)
	}
}

func TestProgressGroupVerticalAndUnusedSpace(t *testing.T) {
	for _, orientation := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
		items := []nui.ProgressSegment{{ID: "used", Value: 25}, {ID: "zero", Value: 0, Color: nui.Error}}
		tt := ui.NewTester(func(c *ui.Context) {
			nui.ProgressGroup(c, nui.App(c, nui.AppConfig{Mode: nui.LightMode}), items, nui.ProgressGroupOptions{Label: "Group", Orientation: orientation, Height: 200}).Width(200)
		}, 240, 240)
		sample := func(at float32) color.RGBA {
			r := bounds(t, tt, "Group")
			x, y := r.X+at, r.Y+4
			if orientation == nui.Vertical {
				x, y = r.X+4, r.Y+at
			}
			p := tt.Image().RGBAAt(int(x), int(y))
			return p
		}
		if sample(20) != rgba(nui.Light().Primary) || sample(150) != rgba(nui.Light().Accented) {
			t.Fatalf("orientation %d: wrong fill or unused space", orientation)
		}
		items = nil
		tt.Frame()
		if sample(20) != rgba(nui.Light().Accented) || sample(150) != rgba(nui.Light().Accented) {
			t.Fatalf("orientation %d: empty segments left a visible fill", orientation)
		}
	}
}
