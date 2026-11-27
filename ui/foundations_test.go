package nui_test

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func bounds(t *testing.T, tt *ui.Tester, label string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatal("missing", label)
	}
	return r
}

func TestAppSystemPreferencesAndIsolation(t *testing.T) {
	config := nui.DefaultApp()
	var resolved nui.Theme
	tt := ui.NewTester(func(c *ui.Context) {
		resolved = nui.App(c, config)
		nui.Button(c, resolved, "System", nui.ButtonOptions{})
	}, 240, 100)
	other := ui.NewTester(func(c *ui.Context) { nui.App(c, nui.AppConfig{Mode: nui.LightMode}) }, 240, 100)
	tt.SetDark(true)
	if !resolved.Dark {
		t.Fatal("system dark ignored")
	}
	if tt.Image().At(230, 90) == other.Image().At(230, 90) {
		t.Fatal("window theme isolation")
	}
	config.Mode = nui.LightMode
	tt.Frame()
	if resolved.Dark {
		t.Fatal("explicit light ignored")
	}
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true, HighContrast: true})
	if !resolved.Motion.Reduce || !resolved.HighContrast || bounds(t, tt, "System").H != 42 {
		t.Fatalf("preferences not applied: %+v", resolved)
	}
	tt.Frame()
	if bounds(t, tt, "System").H != 42 {
		t.Fatal("text scale compounded on rebuild")
	}
	config.Mode = nui.SystemMode
	tt.Frame()
	if !resolved.Dark {
		t.Fatal("system setting lost after explicit override")
	}
}

func TestHighContrastDefaultAccentText(t *testing.T) {
	luminance := func(c ui.Color) float64 {
		linear := func(v uint8) float64 {
			x := float64(v) / 255
			if x <= .04045 {
				return x / 12.92
			}
			return math.Pow((x+.055)/1.055, 2.4)
		}
		return .2126*linear(c.R) + .7152*linear(c.G) + .0722*linear(c.B)
	}
	for _, theme := range []nui.Theme{nui.Light(), nui.Dark()} {
		contrast := func(a, b ui.Color) float64 {
			x, y := luminance(a), luminance(b)
			return (math.Max(x, y) + .05) / (math.Min(x, y) + .05)
		}
		t.Logf("dark=%v default primary/inverted contrast %.2f:1", theme.Dark, contrast(theme.Primary, theme.Inverted))
		theme = theme.FollowPreferences(ui.Preferences{HighContrast: true})
		for _, accent := range []ui.Color{theme.Primary, theme.Secondary, theme.Success, theme.Info, theme.Warning, theme.Error} {
			if ratio := contrast(accent, theme.Inverted); ratio < 4.5 {
				t.Fatalf("high contrast solid text %.2f:1", ratio)
			}
		}
	}
}

func TestBlackPrimaryTheme(t *testing.T) {
	palettes := nui.DefaultPalettes()
	palettes.Neutral = nui.MauvePalette()
	for _, dark := range []bool{false, true} {
		original := nui.NewTheme(dark, palettes).WithComponent(nui.ComponentButton, nui.ComponentTheme{Size: nui.XL})
		original.Radius, original.Density = 10, .8
		theme := original.WithBlackPrimary()
		want := ui.Hex("#000000")
		if dark {
			want = ui.Hex("#ffffff")
		}
		for _, highContrast := range []bool{false, true} {
			prefs := ui.Preferences{HighContrast: highContrast}
			got, baseline := theme.FollowPreferences(prefs), original.FollowPreferences(prefs)
			if got.Primary != want || got.Native().Accent != want || got.Native().AccentText != baseline.Inverted {
				t.Fatalf("incorrect Black primary or foreground: dark=%v highContrast=%v", dark, highContrast)
			}
			if got.Background != baseline.Background || got.Success != baseline.Success || got.Warning != baseline.Warning || got.Palettes.Neutral != palettes.Neutral {
				t.Fatal("Black primary changed other theme colors")
			}
			if got.Radius != 10 || got.Density != .8 || got.Component(nui.ComponentButton).Size != nui.XL {
				t.Fatal("Black primary discarded theme customization")
			}
		}
		if original.Palettes != palettes || original.Primary == want {
			t.Fatal("Black primary mutated the original snapshot")
		}
	}
}

func TestButtonSizesDensityAndLoadingGeometry(t *testing.T) {
	for i, wantHeight := range []float32{24, 28, 32, 36, 40} {
		for _, icons := range []int{0, 1, 2, 3} {
			t.Run(fmt.Sprintf("size-%d-icons-%d", i, icons), func(t *testing.T) {
				loading := false
				clicks := 0
				tt := ui.NewTester(func(c *ui.Context) {
					theme := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
					ui.Column(c).AlignItems(ui.Start).Children(func() {
						o := nui.ButtonOptions{Size: nui.Size(i + 1), Loading: loading, Animate: nui.Set(false)}
						if icons&1 != 0 {
							o.Leading = lucide.Check()
						}
						if icons&2 != 0 {
							o.Trailing = lucide.ArrowRight()
						}
						nui.Button(c, theme, "Save changes", o).OnClick(func() { clicks++ })
					})
				}, 400, 100)
				before := bounds(t, tt, "Save changes")
				if before.H != wantHeight {
					t.Fatalf("height=%g want=%g", before.H, wantHeight)
				}
				loading = true
				tt.Frame()
				after := bounds(t, tt, "Save changes")
				if before.H != after.H || (icons&1 != 0 && before.W != after.W) || (icons&1 == 0 && before.W >= after.W) {
					t.Fatalf("loading must replace a leading icon or add one beside the visible label: %v -> %v", before, after)
				}
				click(t, tt, "Save changes")
				if clicks != 0 {
					t.Fatal("loading activated")
				}
				loading = false
				tt.Frame()
				click(t, tt, "Save changes")
				if clicks != 1 {
					t.Fatal("activation after loading")
				}
			})
		}
	}
	theme := nui.Light()
	theme.Density = .5
	tt := ui.NewTester(func(c *ui.Context) { nui.Window(c, theme); nui.Button(c, theme, "Compact", nui.ButtonOptions{}) }, 240, 80)
	if bounds(t, tt, "Compact").H != 26 {
		t.Fatal("density should scale padding, not font")
	}
}

func TestAnimationLifecycle(t *testing.T) {
	for _, kind := range []string{"button", "skeleton", "progress", "empty", "input"} {
		t.Run(kind, func(t *testing.T) {
			shown, offscreen, enabled, off := true, false, true, false
			text := ""
			tt := ui.NewTester(func(c *ui.Context) {
				theme := nui.App(c, nui.AppConfig{DisableAnimations: off})
				ui.Scroll(c).Fill().Children(func() {
					if offscreen {
						ui.Box(c).Height(600)
					}
					if !shown {
						return
					}
					switch kind {
					case "button":
						nui.Button(c, theme, "Loading", nui.ButtonOptions{Loading: true, Animate: nui.Set(enabled)})
					case "skeleton":
						nui.Skeleton(c, theme, nui.SkeletonOptions{Animate: nui.Set(enabled)})
					case "progress":
						nui.Progress(c, theme, 0, nui.ProgressOptions{Indeterminate: true, Animate: nui.Set(enabled)})
					case "empty":
						nui.Empty(c, theme, nui.EmptyOptions{Title: "Loading", Loading: enabled})
					case "input":
						nui.Input(c, theme, &text, nui.InputOptions{Label: "Loading", Loading: enabled})
					}
				})
			}, 300, 120)
			tt.Advance(50 * time.Millisecond)
			if !tt.FramePending() {
				t.Fatal("visible animation stopped")
			}
			offscreen = true
			tt.Frame()
			tt.Advance(300 * time.Millisecond)
			if tt.FramePending() {
				t.Fatal("offscreen animation did not stop")
			}
			offscreen = false
			tt.Frame()
			if !tt.FramePending() {
				t.Fatal("returning animation did not start")
			}
			enabled = false
			tt.Frame()
			tt.Advance(300 * time.Millisecond)
			if tt.FramePending() {
				t.Fatal("explicit false ignored")
			}
			// Spinners and pulses go on under reduced motion, as Nuxt's
			// animate-spin and animate-pulse do; an indeterminate bar pulses
			// in place.
			enabled = true
			tt.SetPreferences(ui.Preferences{ReduceMotion: true})
			tt.Advance(300 * time.Millisecond)
			if !tt.FramePending() {
				t.Fatal("reduced motion stopped the animation")
			}
			off = true
			tt.Frame()
			tt.Advance(300 * time.Millisecond)
			if tt.FramePending() {
				t.Fatal("disabled animations ignored")
			}
			off = false
			tt.SetPreferences(ui.Preferences{})
			shown = false
			tt.Frame()
			tt.Advance(300 * time.Millisecond)
			if tt.FramePending() {
				t.Fatal("unmounted animation did not stop")
			}
		})
	}
}

func TestKeyboardActionsChipAndDisabledParent(t *testing.T) {
	calls := 0
	notice := true
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Gap(16).AlignItems(ui.Start).Children(func() {
			nui.Link(c, theme, "Documentation", nui.LinkOptions{}).OnClick(func() { calls++ })
			nui.Chip(c, theme, nui.ChipOptions{Text: "3"}, func() { nui.Button(c, theme, "Inbox", nui.ButtonOptions{}).OnClick(func() { calls++ }) })
			nui.User(c, theme, nui.UserOptions{Name: "Member", OnActivate: func() { calls++ }})
			ui.Column(c).Disabled(true).Children(func() { nui.Button(c, theme, "Blocked", nui.ButtonOptions{}).OnClick(func() { calls += 100 }) })
			if notice {
				nui.Alert(c, theme, nui.AlertOptions{Title: "Saved", OnClose: func() { notice = false }})
			}
		})
	}, 400, 360)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Documentation") {
		t.Fatal("link not keyboard focusable")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeySpace)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if calls != 3 {
		t.Fatalf("keyboard calls %d", calls)
	}
	click(t, tt, "Blocked")
	if calls != 3 {
		t.Fatal("parent disabled ignored")
	}
	click(t, tt, "Dismiss Saved")
	tt.Frame()
	tt.Frame()
	if notice {
		t.Fatal("dismiss failed")
	}
	// Chip overlaps the host's upper-right corner but must never intercept its hit.
	r := bounds(t, tt, "Inbox")
	tt.ClickAt(r.X+r.W-2, r.Y+2)
	if calls != 4 {
		t.Fatal("chip swallowed host click")
	}
}

// Links follow the web and WAI-ARIA: Enter activates them, while Space is
// left to the container around them (MyGo 0.8). Buttons keep both keys.
func TestLinkRolesActivateOnEnterOnly(t *testing.T) {
	calls := map[string]int{}
	tt := ui.NewTester(func(c *ui.Context) {
		theme := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(16).Gap(16).AlignItems(ui.Start).Children(func() {
			nui.Link(c.Key("link"), theme, "Documentation", nui.LinkOptions{}).OnClick(func() { calls["Documentation"]++ })
			nui.Breadcrumb(c.Key("breadcrumb"), theme, []nui.BreadcrumbItem{
				{ID: "home", Label: "Home", OnSelect: func() { calls["Home"]++ }},
				{ID: "current", Label: "Current"},
			}, nui.BreadcrumbOptions{})
			nui.FooterColumns(c.Key("footer"), theme, []nui.FooterColumn{
				{ID: "company", Title: "Company", Items: []nui.NavigationItem{{ID: "about", Label: "About"}}},
			}, nui.FooterColumnsOptions{OnNavigate: func(id string) { calls[id]++ }})
		})
	}, 600, 400)
	for _, link := range []struct{ name, id string }{{"Documentation", "Documentation"}, {"Home", "Home"}, {"About", "about"}} {
		tt.Key(0, ui.KeyTab)
		if !tt.Focused(link.name) {
			t.Fatalf("%s not reached by Tab", link.name)
		}
		tt.Key(0, ui.KeySpace)
		if calls[link.id] != 0 {
			t.Fatalf("Space activated link %s", link.name)
		}
		tt.Key(0, ui.KeyEnter)
		if calls[link.id] != 1 {
			t.Fatalf("Enter activated link %s %d times", link.name, calls[link.id])
		}
	}
}

func TestIdentityAndProgressEdgeCases(t *testing.T) {
	for in, want := range map[string]string{"": "?", "  Ada Lovelace ": "AL", "林晓": "林晓", "éva smith": "ÉS", "😀": "😀"} {
		if got := nui.Initials(in); got != want {
			t.Fatalf("initials %q: %q", in, got)
		}
	}
	if nui.FormatKey("CmdOrCtrl", "darwin") != "⌘" || nui.FormatKey("CmdOrCtrl", "windows") != "Ctrl" {
		t.Fatal("platform keys")
	}
	bitmap, err := ui.DecodeBitmap([]byte("invalid"))
	if err == nil {
		t.Fatal("invalid image unexpectedly decoded")
	}
	for _, value := range []float64{-12, 0, 100, 140, math.NaN(), math.Inf(1)} {
		tt := ui.NewTester(func(c *ui.Context) {
			theme := nui.App(c, nui.AppConfig{})
			ui.Column(c).FillWidth().AlignItems(ui.Start).Children(func() {
				nui.Avatar(c, theme, nui.AvatarOptions{Image: bitmap, Alt: "Fallback Name"})
				nui.AvatarGroup(c, theme, []nui.AvatarItem{{ID: "a", Avatar: nui.AvatarOptions{Text: "A"}}, {ID: "b", Avatar: nui.AvatarOptions{Text: "B"}}, {ID: "c", Avatar: nui.AvatarOptions{Text: "C"}}}, nui.AvatarGroupOptions{Max: nui.Set(2)}).Label("Group")
				nui.Progress(c, theme, value, nui.ProgressOptions{Label: "Progress"}).Width(200)
				nui.ProgressGroup(c, theme, []nui.ProgressSegment{{ID: "a", Value: math.MaxFloat64, Color: nui.Success}, {ID: "b", Value: math.MaxFloat64, Color: nui.Error}}, nui.ProgressGroupOptions{Label: "Huge", HideLegend: true}).Width(200)
			})
		}, 300, 200)
		bounds(t, tt, "FN")
		bounds(t, tt, "1 more person")
		if w := bounds(t, tt, "Group").W; w != 84 { // 3*32 - 2*6: Nuxt MD overlap is 6 DIP.
			t.Fatalf("avatar overlap width %g, want 84", w)
		}
		r := bounds(t, tt, "Progress")
		im := tt.Image()
		left, right := im.At(int(r.X+30), int(r.Y+4)), im.At(int(r.X+170), int(r.Y+4))
		if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) || value >= 100 {
			if left != right {
				t.Fatalf("invalid/clamped progress edge mismatch: %v", value)
			}
		}
		r = bounds(t, tt, "Huge")
		if im.At(int(r.X+30), int(r.Y+4)) == im.At(int(r.X+150), int(r.Y+4)) {
			t.Fatal("large values lost distinct segments")
		}
	}
}

func TestButtonLoadingIconAnimationChangesPixels(t *testing.T) {
	off := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{DisableAnimations: off})
		nui.Button(c, th, "Loading", nui.ButtonOptions{Loading: true})
	}, 240, 100)
	before := slices.Clone(tt.Image().Pix)
	tt.Advance(125 * time.Millisecond)
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("loading icon scheduled frames without rotating")
	}
	// A spinner is no movement: it turns under reduced motion, as Nuxt's
	// animate-spin and MyGo's Loop do.
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	before = slices.Clone(tt.Image().Pix)
	tt.Advance(125 * time.Millisecond)
	if bytes.Equal(before, tt.Image().Pix) || !tt.FramePending() {
		t.Fatal("reduced motion stopped the loading icon")
	}
	off = true
	tt.Frame()
	before = slices.Clone(tt.Image().Pix)
	tt.Advance(125 * time.Millisecond)
	if !bytes.Equal(before, tt.Image().Pix) || tt.FramePending() {
		t.Fatal("disabled animations did not stop the loading icon")
	}
}

func TestButtonRestoringMotionKeepsCurrentColor(t *testing.T) {
	color, reduce := nui.Primary, false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		th.Motion.Disabled = reduce
		ui.Column(c).Padding(16).Children(func() { nui.Button(c, th, "Motion button", nui.ButtonOptions{Color: color}).Width(160) })
	}, 220, 90)
	color, reduce = nui.Error, true
	tt.Frame()
	r := bounds(t, tt, "Motion button")
	before := tt.Image().RGBAAt(int(r.X)+2, int(r.Y+r.H/2))
	if before != rgba(nui.Light().Error) {
		t.Fatal("disabled animations did not show the current button color immediately")
	}
	reduce = false
	tt.Frame()
	if after := tt.Image().RGBAAt(int(r.X)+2, int(r.Y+r.H/2)); after != before {
		t.Fatalf("restoring motion reverted the current button color: %v -> %v", before, after)
	}
}

func TestButtonLoadingIconUsesPartForeground(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.Button(c, th, "Loading", nui.ButtonOptions{Loading: true, Animate: nui.Set(false), UI: nui.Styles{nui.Leading: {Foreground: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(root, label, leading, trailing ui.Element) {
			leading.Label("Loading icon")
		}})
	}, 240, 100)
	r := bounds(t, tt, "Loading icon")
	if countWhere(tt.Image(), r, func(c color.RGBA) bool { return int(c.R) > int(c.G)+20 && int(c.B) > int(c.G)+20 }) == 0 {
		t.Fatal("part foreground was ignored by the loading icon")
	}
}

func TestButtonAvatarDimensionsAndPriority(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		avatar := nui.AvatarOptions{Text: "AL", Alt: "Ada Lovelace"}
		opts := nui.ButtonOptions{Size: size, Avatar: &avatar, Animate: nui.Set(false)}
		calls := 0
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{})
			nui.Button(c.Key("account"), th, "Account", opts).OnClick(func() { calls++ })
		}, 320, 120)
		want := []float32{16, 16, 20, 20, 24}[i]
		r := bounds(t, tt, "Ada Lovelace")
		if r.W != want || r.H != want {
			t.Fatalf("size=%v avatar=%+v, want %g", size, r, want)
		}
		click(t, tt, "Account")
		if calls != 1 {
			t.Fatal("avatar button did not activate once")
		}
		opts.Leading = lucide.Check()
		tt.Frame()
		if tt.HasText("Ada Lovelace") || tt.HasText("AL") || calls != 1 || !tt.Focused("Account") {
			t.Fatal("leading icon did not replace avatar while preserving state")
		}
		opts.Leading = nil
		opts.Loading = true
		tt.Frame()
		if tt.HasText("Ada Lovelace") {
			t.Fatal("leading loader did not replace avatar")
		}
		click(t, tt, "Account")
		tt.Key(0, ui.KeySpace)
		if calls != 1 {
			t.Fatal("loading avatar button activated")
		}
		opts.LoadingTrailing = true
		tt.Frame()
		if !tt.HasText("Ada Lovelace") {
			t.Fatal("trailing loader hid the leading avatar")
		}
		opts.Loading = false
		avatar.Size = nui.XS
		tt.Frame()
		if r = bounds(t, tt, "Ada Lovelace"); r.W != 24 || r.H != 24 {
			t.Fatalf("explicit avatar size was overridden: %+v", r)
		}
		if avatar.Size != nui.XS || avatar.Text != "AL" {
			t.Fatal("button mutated avatar options")
		}
	}
}

func TestButtonAvatarOnlyRTLAndPartStyles(t *testing.T) {
	avatar := nui.AvatarOptions{Icon: lucide.User(), Alt: "Account image"}
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Gap(12).Children(func() {
			nui.Button(c.Key("only"), th, "", nui.ButtonOptions{Avatar: &avatar, AccessibleLabel: "Account", Style: func(root, label, leading, trailing ui.Element) { leading.Label("Styled avatar") }}).OnClick(func() { calls++ })
			nui.Button(c.Key("both"), th, "Profile", nui.ButtonOptions{Avatar: &avatar, Trailing: lucide.ChevronDown(), Style: func(root, label, leading, trailing ui.Element) {
				leading.Label("Leading avatar")
				trailing.Label("Trailing icon")
			}})
		})
	}, 400, 120)
	if r := bounds(t, tt, "Account"); r.W != r.H {
		t.Fatalf("avatar-only button not square: %+v", r)
	}
	if bounds(t, tt, "Leading avatar").X >= bounds(t, tt, "Trailing icon").X {
		t.Fatal("LTR avatar order")
	}
	tt.SetDirection(ui.RightToLeft)
	if bounds(t, tt, "Leading avatar").X <= bounds(t, tt, "Trailing icon").X {
		t.Fatal("RTL avatar order")
	}
	click(t, tt, "Account")
	tt.Key(0, ui.KeySpace)
	if calls != 2 || !tt.Focused("Account") {
		t.Fatal("avatar-only mouse/Space activation or focus")
	}
}

func TestBuiltInPalettesMatchPinnedTailwind(t *testing.T) {
	primary := strings.Fields("red orange amber yellow lime green emerald teal cyan sky blue indigo violet purple fuchsia pink rose")
	neutral := strings.Fields("slate gray zinc neutral stone taupe mauve mist olive")
	if !slices.Equal(nui.PrimaryPaletteNames(), primary) || !slices.Equal(nui.NeutralPaletteNames(), neutral) {
		t.Fatal("palette choices differ from the Nuxt UI theme menu")
	}
	css, err := os.ReadFile("../third_party/tailwind/theme.css")
	if err != nil {
		t.Fatal(err)
	}
	// Compare every generated shade with the independent vendored CSS source,
	// including the achromatic `none` hue used by Zinc, Neutral and Mauve.
	declarations := regexp.MustCompile(`--color-([a-z]+)-(\d+): oklch\(([^)]+)\)`).FindAllStringSubmatch(string(css), -1)
	counts := map[string]int{}
	for _, declaration := range declarations {
		name := declaration[1]
		palette, ok := nui.PaletteByName(name)
		if !ok {
			t.Fatalf("missing Tailwind palette %q", name)
		}
		shade, err := strconv.Atoi(declaration[2])
		if err != nil {
			t.Fatal(err)
		}
		components := strings.Fields(declaration[3])
		var values [3]float64
		for i, component := range components {
			if component == "none" {
				continue
			}
			values[i], err = strconv.ParseFloat(strings.TrimSuffix(component, "%"), 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		want := ui.Oklch(float32(values[0]/100), float32(values[1]), float32(values[2]))
		if got := palette.Shade(shade); got != want {
			t.Errorf("%s-%d: got %v, want %v", name, shade, got, want)
		}
		counts[name]++
	}
	for _, name := range append(primary, neutral...) {
		if counts[name] != 11 {
			t.Errorf("%s: checked %d shades, want 11", name, counts[name])
		}
	}
	for _, name := range []string{"", "black", "white", "Blue", "unknown"} {
		if p, ok := nui.PaletteByName(name); ok || p != (nui.Palette{}) {
			t.Errorf("unexpected palette for %q", name)
		}
	}
}
