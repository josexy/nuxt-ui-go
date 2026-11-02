package nui_test

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// Shared recipes aligned with Nuxt UI v4.12: overlay surfaces and motion.

func TestPopoverPaddingAndScaleIn(t *testing.T) {
	open := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			anchor := nui.Button(c, th, "Details", nui.ButtonOptions{}).OnClick(func() { open = !open })
			nui.Popover(c, th, anchor, &open, nui.PopoverOptions{Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Panel {
					e.Label("Popover panel")
				}
			}}, func(nui.Theme) { ui.Text(c, "Inside") })
		})
	}, 400, 300)
	click(t, tt, "Details")
	opening := bounds(t, tt, "Popover panel")
	tt.Advance(50 * time.Millisecond)
	middle := bounds(t, tt, "Popover panel")
	tt.Advance(time.Second)
	settled := bounds(t, tt, "Popover panel")
	if !(opening.W < middle.W && middle.W < settled.W) {
		t.Fatalf("popover did not scale in: widths %v, %v, %v", opening.W, middle.W, settled.W)
	}
	// Scaling from the anchor side keeps the edge facing the anchor in place.
	if opening.Y != settled.Y || opening.X != settled.X {
		t.Fatalf("scale-in moved the anchored corner: %v -> %v", opening, settled)
	}
	if tt.FramePending() {
		t.Fatal("settled popover keeps requesting frames")
	}
	inside := bounds(t, tt, "Inside")
	if inside.X-settled.X != 16 || inside.Y-settled.Y != 16 {
		t.Fatalf("popover content offset (%v, %v), want 16 DIP padding", inside.X-settled.X, inside.Y-settled.Y)
	}

	click(t, tt, "Details")
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	click(t, tt, "Details")
	if got := bounds(t, tt, "Popover panel"); got != settled {
		t.Fatalf("reduced motion still animates: %v, want %v", got, settled)
	}
}

func TestTooltipIsOneTruncatedLine(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		a := nui.Button(c, th, "Anchor", nui.ButtonOptions{}).Absolute().Left(150).Top(100)
		nui.Tooltip(c, th, a, "A rather long explanation that cannot fit on one tooltip line", nui.TooltipOptions{Open: &open, MaxWidth: nui.Set(float32(160)), Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Panel {
				e.Label("Tip")
			}
		}})
	}, 500, 300)
	tt.Advance(time.Second)
	tip := bounds(t, tt, "Tip")
	if tip.H != 24 || tip.W > 160 {
		t.Fatalf("tooltip box %v, want one 24 DIP line within its 160 DIP maximum", tip)
	}
	if text := bounds(t, tt, "A rather long explanation that cannot fit on one tooltip line"); text.H != 16 || text.X+text.W > tip.X+tip.W-10 {
		t.Fatalf("tooltip text %v in %v is not one truncated line inside the 10 DIP padding", text, tip)
	}
}

func TestModalScalesInOverFadingBackdrop(t *testing.T) {
	open := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Confirm"}, func(nui.Theme) { ui.Text(c, "Body") })
	}, 600, 400)
	// The backdrop is Elevated at 75% over white; it starts transparent.
	corner := func() uint8 { return tt.Image().RGBAAt(2, 2).G }
	first, panel := corner(), bounds(t, tt, "Confirm")
	tt.Advance(time.Second)
	settled := bounds(t, tt, "Confirm")
	if first <= corner() {
		t.Fatalf("backdrop did not fade in: corner green %d then %d", first, corner())
	}
	if panel.W >= settled.W || settled.W != 512 {
		t.Fatalf("modal did not scale in to 512 DIPs: %v -> %v", panel, settled)
	}
}

func TestOptionRowsFollowSizeAndHighlight(t *testing.T) {
	for _, tc := range []struct {
		size      nui.Size
		row, line float32
	}{{nui.XS, 24, 16}, {nui.MD, 32, 20}, {nui.XL, 40, 24}} {
		value := "c"
		tt := ui.NewTester(func(c *ui.Context) {
			th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
			ui.Column(c).Padding(20).Children(func() {
				options := []nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First letter"}, {Value: "b", Label: "Beta"}, {Value: "c", Label: "Gamma"}}
				nui.Select(c, th, &value, options, nui.SelectOptions{Label: "Letter", Size: tc.size, Popup: nui.PopoverOptions{Label: "Letter options"}}).Width(240)
			})
		}, 400, 400)
		click(t, tt, "Letter")
		tt.Advance(time.Second)
		// Beta and the description appear only in the popup; Gamma is also the trigger's value.
		beta, description := bounds(t, tt, "Beta"), bounds(t, tt, "First letter")
		if beta.H != tc.row || description.H != tc.line {
			t.Fatalf("size %v: row %v, description line %v; want %v and the row's own %v", tc.size, beta.H, description.H, tc.row, tc.line)
		}
		th := nui.Light()
		im := tt.Image()
		// The selected row, after Beta, is highlighted with elevated/50 and checked in the text color.
		gamma := ui.Rect{X: beta.X, Y: beta.Y + beta.H, W: beta.W, H: beta.H}
		got, want := im.RGBAAt(int(gamma.X)+2, int(gamma.Y+gamma.H/2)), rgba(th.Elevated.Mix(th.Background, .5))
		if d := int(got.B) - int(want.B); d < -1 || d > 1 || got == rgba(th.Elevated) {
			t.Fatalf("size %v: highlighted row surface %v, want elevated/50 %v", tc.size, got, want)
		}
		if n := countColor(im, bounds(t, tt, "Letter options"), rgba(th.Primary)); n != 0 {
			t.Fatalf("size %v: the selected check uses the primary color (%d pixels)", tc.size, n)
		}
	}
}

func TestListboxRowsRadiusAndEmptyState(t *testing.T) {
	var selected []string
	var s, empty nui.ListboxState
	var menu nui.SelectMenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Gap(12).Children(func() {
			nui.Listbox(c, th, &selected, &s, []nui.ListboxItem{{ID: "a", Label: "Alpha"}}, nui.ListboxOptions{Label: "Letters", Height: 120}).Width(240)
			nui.Listbox(c, th, &selected, &empty, nil, nui.ListboxOptions{Label: "Nothing", EmptyText: "Nothing here", Height: 120}).Width(240)
			nui.SelectMenu(c, th, &selected, &menu, []nui.ListboxItem{{ID: "x", Label: "Xray"}}, nui.SelectMenuOptions{Label: "Small menu", Size: nui.XS}).Width(240)
		})
	}, 400, 600)
	if h := bounds(t, tt, "Alpha").H; h != 32 {
		t.Fatalf("listbox row is %v DIPs tall, want Nuxt's 32", h)
	}
	text, root := bounds(t, tt, "Nothing here"), bounds(t, tt, "Nothing")
	if d := (text.X + text.W/2) - (root.X + root.W/2); d < -1 || d > 1 {
		t.Fatalf("empty state %v is not centered in %v", text, root)
	}
	click(t, tt, "Small menu")
	tt.Advance(time.Second)
	if h := bounds(t, tt, "Xray").H; h != 24 {
		t.Fatalf("an XS SelectMenu lists %v DIP rows, want 24", h)
	}
}

func TestMenuRowsUseOptionRecipe(t *testing.T) {
	var state nui.MenuState
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			anchor := nui.Button(c, th, "Actions", nui.ButtonOptions{})
			nui.DropdownMenu(c, th, anchor, &state, []nui.MenuItem{{ID: "edit", Label: "Edit", Description: "Change the title"}, {ID: "delete", Label: "Delete", Color: nui.Error}}, nui.MenuOptions{})
		})
	}, 400, 400)
	click(t, tt, "Actions")
	tt.Advance(time.Second)
	th := nui.Light()
	edit, del := bounds(t, tt, "Edit"), bounds(t, tt, "Delete")
	if del.H != 32 || edit.H != 52 {
		t.Fatalf("menu rows %v and %v; want 32, and 52 with a same-size description", del, edit)
	}
	// The first row has focus: elevated/50, not the opaque elevated surface.
	if got := tt.Image().RGBAAt(int(edit.X)+2, int(edit.Y+edit.H/2)); got == rgba(th.Elevated) || got == white {
		t.Fatalf("highlighted menu row surface %v", got)
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	// Error at 10% over white: a light red, not elevated/50.
	if got := tt.Image().RGBAAt(int(del.X)+2, int(del.Y+del.H/2)); got == white || got.R <= got.G || got.R <= got.B {
		t.Fatalf("highlighted destructive row %v is not tinted with its color", got)
	}
}

func TestSmallRadiusTokens(t *testing.T) {
	same := func(t *testing.T, name string, build func(c *ui.Context, th nui.Theme, explicit bool)) {
		t.Helper()
		render := func(explicit bool) []byte {
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
				ui.Column(c).Padding(10).Children(func() { build(c, th, explicit) })
			}, 200, 80)
			return tt.Image().Pix
		}
		if !slices.Equal(render(false), render(true)) {
			t.Errorf("%s default radius is not Nuxt's rounded-sm (4)", name)
		}
	}
	four := nui.Set(float32(4))
	same(t, "Checkbox", func(c *ui.Context, th nui.Theme, explicit bool) {
		on := true
		o := nui.CheckboxOptions{Label: "Box"}
		if explicit {
			o.UI = nui.Styles{nui.Indicator: {Radius: four}}
		}
		nui.Checkbox(c, th, &on, o)
	})
	for _, size := range []nui.Size{nui.XS, nui.SM} {
		same(t, "Badge", func(c *ui.Context, th nui.Theme, explicit bool) {
			o := nui.BadgeOptions{Size: size}
			if explicit {
				o.UI = nui.Styles{nui.Root: {Radius: four}}
			}
			nui.Badge(c, th, "New", o)
		})
	}
	same(t, "Kbd", func(c *ui.Context, th nui.Theme, explicit bool) {
		o := nui.KbdOptions{}
		if explicit {
			o.UI = nui.Styles{nui.Label: {Radius: four}}
		}
		nui.Kbd(c, th, []string{"K"}, o)
	})
}

func TestTitlesUseNuxtsPageSize(t *testing.T) {
	open := true
	collapsed := false
	drawer := nui.DrawerState{Open: true}
	// Containers named after their title come first; rename them or the heading itself.
	rename := func(slot nui.Slot, name string) func(nui.Slot, ui.Element) {
		return func(s nui.Slot, e ui.Element) {
			if s == slot {
				e.Label(name)
			}
		}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Children(func() {
			nui.Sidebar(c, th, &collapsed, nui.SidebarOptions{Title: "Side title", Style: rename(nui.Root, "Sidebar region")}, nil)
			ui.Column(c).Padding(12).Children(func() {
				nui.Card(c, th, nui.CardOptions{Title: "Card title"}, nil).Width(240)
			})
		})
		nui.ModalWithOptions(c, th, &open, nui.ModalOptions{Title: "Modal title", Style: rename(nui.Label, "Modal heading")}, nil)
	}, 900, 600)
	tt.Advance(time.Second)
	for _, title := range []string{"Side title", "Card title", "Modal heading"} {
		if h := bounds(t, tt, title).H; h != 24 {
			t.Errorf("%s line is %v DIPs, want the 16/24 page size", title, h)
		}
	}
	open = false
	tt2 := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Drawer(c, th, &drawer, nui.DrawerOptions{Title: "Drawer title", Style: rename(nui.Label, "Drawer heading")}, nil)
	}, 600, 600)
	tt2.Advance(time.Second)
	if h := bounds(t, tt2, "Drawer heading").H; h != 24 {
		t.Errorf("Drawer title line is %v DIPs, want 24", h)
	}
}

func TestSelectMultipleDrawsOneIcon(t *testing.T) {
	var value []string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.SelectMultiple(c, th, &value, []nui.Option[string]{{Value: "a", Label: "Alpha", Description: "First", Icon: lucide.Star()}}, nui.SelectOptions{Label: "Letters"}).Width(240)
		})
	}, 400, 300)
	click(t, tt, "Letters")
	tt.Advance(time.Second)
	// Padding 6, one 20 DIP icon and a 6 DIP gap precede the texts.
	if offset := bounds(t, tt, "First").X - bounds(t, tt, "Alpha").X; offset != 32 {
		t.Fatalf("option text starts %v DIPs into its row, want 32 after a single icon", offset)
	}
}
