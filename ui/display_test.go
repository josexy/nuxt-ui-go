package nui_test

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestAvatarImageFillsStyledRootAndExposesPart(t *testing.T) {
	pixels := image.NewRGBA(image.Rect(0, 0, 20, 30))
	draw.Draw(pixels, pixels.Bounds(), image.NewUniform(color.RGBA{R: 210, G: 40, B: 50, A: 255}), image.Point{}, draw.Src)
	bitmap := ui.NewBitmap(pixels)
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Avatar(c, th, nui.AvatarOptions{Image: bitmap, Alt: "Photo", Style: func(slot nui.Slot, e ui.Element) {
			switch slot {
			case nui.Root:
				e.Size(48, 48)
			case nui.Body:
				e.Label("Avatar image")
			}
		}})
	}, 100, 100)
	root, img := bounds(t, tt, "Photo"), bounds(t, tt, "Avatar image")
	if img != root {
		t.Fatalf("image did not fill customized root: image=%+v root=%+v", img, root)
	}
}

func TestAvatarFallbackPriorityAndNames(t *testing.T) {
	bitmap := ui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 20, 30)))
	opts := nui.AvatarOptions{Image: bitmap, Icon: lucide.User(), Text: "+12", Alt: "Ada Lovelace"}
	parts := map[nui.Slot]bool{}
	opts.Style = func(slot nui.Slot, _ ui.Element) { parts[slot] = true }
	tt := ui.NewTester(func(c *ui.Context) {
		clear(parts)
		nui.Avatar(c, nui.App(c, nui.AppConfig{}), opts)
	}, 100, 100)
	check := func(part nui.Slot, text string) {
		t.Helper()
		if !parts[part] || len(parts) != 2 || (text != "" && !tt.HasText(text)) {
			t.Fatalf("fallback priority: parts=%v, text=%q", parts, text)
		}
		bounds(t, tt, opts.Alt)
	}
	check(nui.Body, "")
	opts.Image = nil
	tt.Frame()
	check(nui.Leading, "")
	opts.Icon = nil
	tt.Frame()
	check(nui.Label, "+12")
	opts.Text = ""
	tt.Frame()
	check(nui.Label, "AL")
	opts.Alt = "林晓"
	tt.Frame()
	check(nui.Label, "林")
	opts.Alt, opts.Text = "", "+5"
	tt.Frame()
	bounds(t, tt, "+5")
	if bounds(t, tt, "+5").W != 32 {
		t.Fatal("literal Text was not used as the image name when Alt was empty")
	}
}

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

func TestBadgeIconOnlyUsesAutomaticSquarePadding(t *testing.T) {
	for _, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Row(c).Children(func() { nui.Badge(c, th, "", nui.BadgeOptions{Size: size, Leading: lucide.Check()}).Label("Badge") })
		}, 200, 100)
		r := bounds(t, tt, "Badge")
		if r.W != r.H {
			t.Fatalf("size %v icon-only badge=%+v", size, r)
		}
	}
}

func TestBadgeAvatarDimensionsAndPriority(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		avatar := nui.AvatarOptions{Alt: "Ada Lovelace"}
		opts := nui.BadgeOptions{Size: size, Avatar: &avatar}
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			ui.Row(c).Children(func() { nui.Badge(c.Key("badge"), th, "Online", opts).Label("Badge") })
		}, 260, 100)
		want := []float32{16, 16, 16, 20, 20}[i]
		if r := bounds(t, tt, "Ada Lovelace"); r.W != want || r.H != want {
			t.Fatalf("size=%v avatar=%+v, want %g", size, r, want)
		}
		opts.Leading = lucide.Check()
		tt.Frame()
		if tt.HasText("Ada Lovelace") {
			t.Fatal("leading icon did not replace badge avatar")
		}
		opts.Leading = nil
		avatar.Size = nui.LG
		tt.Frame()
		if r := bounds(t, tt, "Ada Lovelace"); r.W != 36 || r.H != 36 {
			t.Fatal("explicit avatar size ignored")
		}
		tt.Key(0, ui.KeyTab)
		if tt.Focused("Badge") {
			t.Fatal("display badge became keyboard-interactive")
		}
	}
}

func TestBadgeSquareKeepsInstancePadding(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Children(func() {
			nui.Badge(c, th, "", nui.BadgeOptions{Leading: lucide.Check(), UI: nui.Styles{nui.Root: {PaddingX: nui.Set(float32(20))}}}).Label("Badge")
		})
	}, 200, 100)
	if r := bounds(t, tt, "Badge"); r.W != 56 {
		t.Fatalf("automatic square overrode explicit padding: %+v", r)
	}
}

func TestChipCornersInsetAndDirection(t *testing.T) {
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		for _, pos := range []nui.ChipPosition{nui.TopLeft, nui.TopRight, nui.BottomLeft, nui.BottomRight} {
			for _, inset := range []bool{false, true} {
				t.Run(fmt.Sprintf("%v/%v/%v", dir, pos, inset), func(t *testing.T) {
					tt := ui.NewTester(func(c *ui.Context) {
						th := nui.App(c, nui.AppConfig{})
						ui.Column(c).Padding(32).AlignItems(ui.Start).Children(func() {
							nui.Chip(c, th, nui.ChipOptions{Position: pos, Inset: inset, Style: func(slot nui.Slot, e ui.Element) {
								if slot == nui.Indicator {
									e.Label("Marker")
								}
							}}, func() { ui.Box(c).Size(80, 40).Label("Host") })
						})
					}, 180, 120)
					tt.SetDirection(dir)
					host, marker := bounds(t, tt, "Host"), bounds(t, tt, "Marker")
					x, y := host.X+host.W-4, host.Y-4
					if pos == nui.TopLeft || pos == nui.BottomLeft {
						x = host.X - 4
					}
					if pos == nui.BottomLeft || pos == nui.BottomRight {
						y = host.Y + host.H - 4
					}
					if inset {
						if pos == nui.TopLeft || pos == nui.BottomLeft {
							x += 4
						} else {
							x -= 4
						}
						if pos == nui.TopLeft || pos == nui.TopRight {
							y += 4
						} else {
							y -= 4
						}
					}
					if marker.X != x || marker.Y != y || marker.W != 8 || marker.H != 8 {
						t.Fatalf("position mismatch: marker=%+v want=(%g,%g,8,8)", marker, x, y)
					}
				})
			}
		}
	}
}

func TestChipTextWidthIncludesIndicatorStyles(t *testing.T) {
	var want float32
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		want, _ = c.MeasureText(1000, ui.Span{Text: "99+", Font: th.MonoFont, Size: 24, Weight: 700})
		nui.Chip(c, th, nui.ChipOptions{Text: "99+", UI: nui.Styles{nui.Indicator: {Font: nui.Set(th.MonoFont), FontSize: nui.Set(float32(24)), Weight: nui.Set(700), PaddingX: nui.Set(float32(4))}}}, nil).AlignSelf(ui.Start).Label("Standalone")
	}, 180, 100)
	if got := bounds(t, tt, "Standalone").W; math.Abs(float64(got-want-8)) > .1 {
		t.Fatalf("styled text width=%g, want glyphs %g + padding 8", got, want)
	}
}

func TestChipHiddenKeepsHostFocusAndClick(t *testing.T) {
	hidden, calls := false, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).AlignItems(ui.Start).Children(func() {
			nui.Chip(c.Key("chip"), th, nui.ChipOptions{Text: "3", Hidden: hidden}, func() {
				nui.Button(c.Key("host"), th, "Host", nui.ButtonOptions{}).OnClick(func() { calls++ })
			})
		})
	}, 180, 100)
	r := bounds(t, tt, "Host")
	tt.ClickAt(r.X+r.W-1, r.Y+1)
	if calls != 1 {
		t.Fatal("chip intercepted host activation")
	}
	hidden = true
	tt.Frame()
	if !tt.Focused("Host") || bounds(t, tt, "Host") != r || tt.HasText("3") {
		t.Fatal("hiding chip altered host layout, focus or indicator visibility")
	}
	tt.Key(0, ui.KeyEnter)
	if calls != 2 {
		t.Fatal("host did not remain keyboard operable")
	}
}

func TestChipNeutralUsesInvertedBackgroundToken(t *testing.T) {
	th := nui.Light()
	th.InvertedBackground = ui.Hex("#e879f9")
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, Light: th})
		nui.Chip(c, th, nui.ChipOptions{Color: nui.Neutral}, nil).AlignSelf(ui.Start).Label("Chip")
	}, 100, 50)
	r := bounds(t, tt, "Chip")
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)); got != rgba(th.InvertedBackground) {
		t.Fatalf("neutral background=%v, want customized inverted token %v", got, rgba(th.InvertedBackground))
	}
}

func TestKbdExplicitPlatformSymbols(t *testing.T) {
	for _, platform := range []string{"windows", "linux", "darwin"} {
		for key, want := range map[string]string{"command": "⌘", "win": "⊞", "COMMAND": "⌘", "WIN": "⊞", "F12": "F12", "Custom key": "Custom key"} {
			if got := nui.FormatKey(key, platform); got != want {
				t.Errorf("FormatKey(%q, %q)=%q, want %q", key, platform, got, want)
			}
		}
	}
}

func TestKbdCombinationAndNoninteractiveBehavior(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Gap(8).Children(func() {
			nui.Button(c.Key("before"), th, "Before", nui.ButtonOptions{})
			nui.Kbd(c.Key("keys"), th, []string{"command", "win", "Ctrl", "k", "custom"}, nui.KbdOptions{Platform: "windows"})
			nui.Button(c.Key("after"), th, "After", nui.ButtonOptions{})
		})
	}, 400, 160)
	for _, text := range []string{"⌘", "⊞", "CTRL", "K", "CUSTOM"} {
		if !tt.HasText(text) {
			t.Errorf("key label %q missing", text)
		}
	}
	click(t, tt, "Before")
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("After") {
		t.Fatal("Kbd entered the tab order")
	}
}

func TestNeutralSolidUsesInvertedBackground(t *testing.T) {
	for _, highContrast := range []bool{false, true} {
		th := nui.Light()
		th.InvertedBackground = ui.Hex("#7e22ce")
		th.HighContrast = highContrast
		th.Motion.Reduce = true
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode, Light: th})
			ui.Column(c).Padding(16).Gap(16).Children(func() {
				nui.Button(c, th, "Button", nui.ButtonOptions{Color: nui.Neutral, Variant: nui.Solid})
				nui.Badge(c, th, "Badge", nui.BadgeOptions{Color: nui.Neutral, Variant: nui.Solid}).Label("Badge")
				nui.Alert(c, th, nui.AlertOptions{Title: "Alert", Color: nui.Neutral, Variant: nui.Solid})
				nui.Kbd(c, th, []string{"K"}, nui.KbdOptions{Color: nui.Neutral, Variant: nui.Solid}).Label("Kbd")
			})
		}, 300, 260)
		for _, label := range []string{"Button", "Badge", "Alert", "Kbd"} {
			r := bounds(t, tt, label)
			if got := tt.Image().RGBAAt(int(r.X+2), int(r.Y+r.H/2)); got != rgba(th.InvertedBackground) {
				t.Errorf("highContrast=%v %s background=%v, want %v", highContrast, label, got, rgba(th.InvertedBackground))
			}
		}
		r := bounds(t, tt, "Button")
		tt.Move(r.X+2, r.Y+r.H/2)
		tt.Advance(time.Second)
		got := tt.Image().RGBAAt(int(r.X+2), int(r.Y+r.H/2))
		if highContrast {
			if got != rgba(th.InvertedBackground) {
				t.Fatal("high-contrast hover lost the opaque inverted background")
			}
		} else if got.R < 125 || got.R > 141 || got.G < 34 || got.G > 58 || got.B < 206 || got.B > 213 {
			t.Fatalf("hover did not use inverted background at 90%% opacity: %v", got)
		}
	}
}

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
	// Tester accessibility bounds refresh on builds, while poses advance on paint.
	tt.Frame()
	after := bounds(t, tt, "User avatar")
	if after.W <= before.W || after.H <= before.H {
		t.Fatalf("hover did not scale the avatar: %+v -> %+v", before, after)
	}
}
