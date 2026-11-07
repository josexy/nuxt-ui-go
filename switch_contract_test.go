package nui_test

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func TestSwitchLoadingBlocksControlAndFieldActivation(t *testing.T) {
	on, loading, disabled := false, true, false
	changes, observed := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Disabled(disabled).Children(func() {
			nui.FormField(c, th, nui.FormFieldOptions{Label: "Sync policy"}, func(field nui.Theme) {
				nui.Switch(c.Key("sync"), field, &on, nui.SwitchOptions{Label: "Sync", Loading: loading, OnChange: func(bool) { changes++ }}).OnChange(func() { observed++ })
			})
		})
	}, 320, 140)
	label := func() {
		r := bounds(t, tt, "Sync policy")
		tt.ClickAt(r.X+4, r.Y+8)
		tt.Frame()
	}
	click(t, tt, "Sync")
	tt.Key(0, ui.KeySpace)
	label()
	if on || changes != 0 || observed != 0 {
		t.Fatal("loading Switch accepted pointer, keyboard or field-label activation")
	}
	loading = false
	tt.Frame()
	label()
	if !on || !tt.Focused("Sync") || changes != 1 || observed != 1 {
		t.Fatal("completed loading did not restore field activation exactly once")
	}
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	tt.Frame()
	if on || changes != 2 || observed != 2 {
		t.Fatal("Space did not toggle once after loading")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Sync")
	label()
	tt.Key(0, ui.KeySpace)
	if on || changes != 2 || observed != 2 {
		t.Fatal("Switch escaped ancestor disabled state")
	}
}

func TestSwitchIconsDirectionAndReducedMotion(t *testing.T) {
	on, loading, custom := false, false, false
	size := nui.MD
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			o := nui.SwitchOptions{Label: "Sync", Size: size, Loading: loading, OnChange: func(bool) { changes++ }, Style: func(slot nui.Slot, e ui.Element) {
				switch slot {
				case nui.Indicator:
					e.Label("Track")
				case nui.Thumb:
					e.Label("Thumb")
				}
			}}
			if custom {
				o.CheckedIcon, o.UncheckedIcon, o.LoadingIcon = lucide.Check(), lucide.X(), lucide.RefreshCw()
			}
			nui.Switch(c.Key("sync"), th, &on, o)
		})
	}, 300, 100)
	for _, value := range []bool{false, true} {
		on, custom = value, false
		tt.Frame()
		plain := slices.Clone(tt.Image().Pix)
		custom = true
		tt.Frame()
		if bytes.Equal(plain, tt.Image().Pix) {
			t.Fatalf("custom icon is missing when checked=%v", value)
		}
	}
	for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		tt.SetDirection(direction)
		for i, choiceSize := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
			size = choiceSize
			for _, value := range []bool{false, true} {
				on = value
				tt.Frame()
				track, thumb := bounds(t, tt, "Track"), bounds(t, tt, "Thumb")
				want := track.X + 2
				if value != (direction == ui.RightToLeft) {
					want = track.X + track.W - thumb.W - 2
				}
				if track.W != float32(28+4*i) || thumb.W != float32(12+2*i) || thumb.X != want || thumb.Y != track.Y+2 {
					t.Fatalf("size=%v direction=%v checked=%v: track=%+v thumb=%+v", size, direction, value, track, thumb)
				}
			}
		}
	}
	parityScreenshot(t, tt, "switch-icons-rtl")
	loading = true
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()
	before := slices.Clone(tt.Image().Pix)
	tt.Advance(250 * time.Millisecond)
	if tt.FramePending() || !bytes.Equal(before, tt.Image().Pix) || changes != 0 {
		t.Fatal("reduced-motion loading continued animating or appearance updates emitted changes")
	}
	custom = false
	tt.Frame()
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("LoadingIcon did not differ from the default loader")
	}
}
