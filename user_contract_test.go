package nui_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func TestUserEmptyNameDoesNotReserveSpace(t *testing.T) {
	for _, size := range []nui.Size{nui.XS, nui.MD} {
		tt := ui.NewTester(func(c *ui.Context) {
			nui.User(c, nui.App(c, nui.AppConfig{}), nui.UserOptions{Description: "Description only", Size: size}).Label("User root")
		}, 300, 80)
		root, description := bounds(t, tt, "User root"), bounds(t, tt, "Description only")
		if description.X != root.X || description.Y != root.Y || description.H != root.H {
			t.Fatalf("size %d: empty name reserved layout space: root=%+v description=%+v", size, root, description)
		}
	}
}

func TestUserReducedMotionKeepsFocusFeedback(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Children(func() {
			nui.User(c, th, nui.UserOptions{Name: "Profile", Avatar: nui.AvatarOptions{Text: "AU", Alt: "User avatar"}, OnActivate: func() {}})
		})
	}, 260, 110)
	before := bounds(t, tt, "User avatar")
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Key(0, ui.KeyTab)
	after := bounds(t, tt, "User avatar")
	if !tt.Focused("Profile") || math.Abs(float64(after.W/before.W-1.15)) > .001 || tt.FramePending() {
		t.Fatalf("reduced motion lost immediate focus feedback: before=%+v after=%+v pending=%t", before, after, tt.FramePending())
	}
	tt.SetPreferences(ui.Preferences{})
	if resumed := bounds(t, tt, "User avatar"); math.Abs(float64(resumed.W-after.W)) > .01 {
		t.Fatalf("restoring motion changed the settled scale: %+v -> %+v", after, resumed)
	}
}

func TestUserChipFollowsAvatarAndSize(t *testing.T) {
	showAvatar := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.UserOptions{Name: "Member", Size: nui.XS, Chip: &nui.ChipOptions{Text: "9", Size: nui.XL, Inset: false, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Indicator {
				e.Label("User marker")
			}
		}}}
		if showAvatar {
			opts.Avatar = nui.AvatarOptions{Text: "AU", Alt: "Member avatar", Size: nui.XL}
		}
		nui.User(c, th, opts)
	}, 240, 90)
	av, chip := bounds(t, tt, "Member avatar"), bounds(t, tt, "User marker")
	if av.W != 24 || chip.H != 6 || chip.X < av.X || chip.Y < av.Y || chip.X+chip.W > av.X+av.W {
		t.Fatalf("user size or inset did not override nested options: avatar=%+v chip=%+v", av, chip)
	}
	showAvatar = false
	tt.Frame()
	if tt.HasText("9") {
		t.Fatal("chip rendered without an avatar")
	}
}

func TestUserBodyKeepsFocusWhenAvatarAndOrientationChange(t *testing.T) {
	draft, showAvatar, vertical := "Original", false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		opts := nui.UserOptions{Body: func() { nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "User draft"}).Width(180) }}
		if showAvatar {
			opts.Avatar = nui.AvatarOptions{Text: "AU"}
			opts.Chip = &nui.ChipOptions{}
		}
		if vertical {
			opts.Orientation = nui.Vertical
		}
		nui.User(c, th, opts).Width(240)
	}, 300, 180)
	click(t, tt, "User draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	showAvatar, vertical = true, true
	tt.Frame()
	if !tt.Focused("User draft") {
		t.Fatal("avatar or orientation change discarded body focus")
	}
	tt.Type("Replacement")
	if draft != "Replacement" {
		t.Fatal("avatar or orientation change discarded body selection")
	}
}

func TestUserActivationAndDisabledParent(t *testing.T) {
	calls := 0
	disabled, parentDisabled := false, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Disabled(parentDisabled).Children(func() {
			nui.User(c, th, nui.UserOptions{Name: "Profile", Avatar: nui.AvatarOptions{Text: "AU"}, Chip: &nui.ChipOptions{Text: "9"}, Disabled: disabled, OnActivate: func() { calls++ }})
		})
	}, 260, 100)
	clickOver(t, tt, "9") // the chip ignores input; the press activates the user
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeySpace)
	if calls != 3 {
		t.Fatalf("activation count=%d", calls)
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Profile")
	tt.Key(0, ui.KeyEnter)
	if calls != 3 {
		t.Fatal("disabled user activated")
	}
	disabled, parentDisabled = false, true
	tt.Frame()
	click(t, tt, "Profile")
	if calls != 3 {
		t.Fatal("user cleared inherited disabled state")
	}
}

func TestUserLongNamesStayWithinWidth(t *testing.T) {
	name := strings.Repeat("用户姓名", 20)
	for _, size := range []nui.Size{nui.XS, nui.MD} {
		tt := ui.NewTester(func(c *ui.Context) {
			nui.User(c, nui.App(c, nui.AppConfig{}), nui.UserOptions{Name: name, Description: "Maintainer", Avatar: nui.AvatarOptions{Text: "AU"}, Size: size}).Width(180).Label("User root")
		}, 240, 100)
		root := bounds(t, tt, "User root")
		for _, label := range []string{name, "Maintainer"} {
			r := bounds(t, tt, label)
			if r.X+r.W > root.X+root.W+.1 {
				t.Fatalf("size %d: text escaped user width: %+v > %+v", size, r, root)
			}
		}
	}
}

func TestUserInteractiveAvatarScalesOnHover(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		nui.User(c, nui.App(c, nui.AppConfig{}), nui.UserOptions{Name: "Profile", Avatar: nui.AvatarOptions{Text: "AU", Alt: "User avatar"}, OnActivate: func() {}}).Width(200)
	}, 260, 90)
	before, root := bounds(t, tt, "User avatar"), bounds(t, tt, "Profile")
	tt.Move(root.X+root.W/2, root.Y+root.H/2)
	tt.Advance(time.Second)
	after := bounds(t, tt, "User avatar")
	if after.W <= before.W || after.H <= before.H {
		t.Fatalf("hover did not scale the avatar: %+v -> %+v", before, after)
	}
}
