package nui_test

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestTabsOnlyNavigateTheirOrientation(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, manual := range []bool{false, true} {
			t.Run(fmt.Sprintf("vertical=%v/manual=%v", vertical, manual), func(t *testing.T) {
				value, calls := "a", 0
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					mode := nui.ActivationAutomatic
					if manual {
						mode = nui.ActivationManual
					}
					nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}}, nui.TabsOptions{Vertical: vertical, ActivationMode: mode, OnChange: func(string) { calls++ }})
				}, 500, 300)
				click(t, tt, "First")
				wrong, right := ui.KeyDown, ui.KeyRight
				if vertical {
					wrong, right = right, wrong
				}
				tt.Key(0, wrong)
				if value != "a" || calls != 0 || !tt.Focused("First") {
					t.Fatal("the inactive arrow axis changed tabs")
				}
				tt.Key(0, right)
				if manual {
					if value != "a" || calls != 0 || !tt.Focused("Second") {
						t.Fatal("manual navigation activated content")
					}
					tt.Key(0, ui.KeySpace)
				}
				if value != "b" || calls != 1 {
					t.Fatalf("navigation value/callback=%q/%d", value, calls)
				}
			})
		}
	}
}

func TestTabsPlainTriggerMetrics(t *testing.T) {
	for i, size := range []nui.Size{nui.XS, nui.SM, nui.MD, nui.LG, nui.XL} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			value := "a"
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}}, nui.TabsOptions{Size: size, Underline: true, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Label {
						e.Label("Tab text")
					}
				}})
			}, 500, 300)
			row, text := bounds(t, tt, "First"), bounds(t, tt, "Tab text")
			if row.H != []float32{24, 28, 32, 36, 40}[i] || text.X-row.X != []float32{8, 10, 12, 12, 12}[i] {
				t.Fatalf("size %v row=%v text=%v", size, row, text)
			}
		})
	}
}

func TestTabsLongLabelsStayInsideTheList(t *testing.T) {
	value := "a"
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "Very long first project settings"}, {ID: "b", Label: "Very long second account settings"}}, nui.TabsOptions{Label: "List"}).Width(260)
	}, 300, 200)
	list := bounds(t, tt, "List")
	for _, name := range []string{"Very long first project settings", "Very long second account settings"} {
		r := bounds(t, tt, name)
		if r.X < list.X || r.X+r.W > list.X+list.W {
			t.Fatalf("tab %q exceeds list: tab=%v list=%v", name, r, list)
		}
	}
}

func TestTabsAvatarBadgeItemStylesAndContentSwitch(t *testing.T) {
	value, draft := "a", "Draft"
	content, icon := true, false
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		th.Motion.Reduce = true
		var leading *ui.SVG
		if icon {
			leading = lucide.Info()
		}
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First", Icon: leading, Avatar: &nui.AvatarOptions{Alt: "Tab avatar", Text: "TA"}, Badge: "0", BadgeOptions: &nui.BadgeOptions{Color: nui.Success, Variant: nui.Soft}, UI: nui.Styles{nui.Trigger: {Background: nui.Set(ui.RGB(255, 0, 255))}}, Content: func(c *ui.Context, th nui.Theme) { nui.Input(c, th, &draft, nui.InputOptions{Label: "Tab draft"}) }}, {ID: "b", Label: "Second", Content: func(c *ui.Context, th nui.Theme) { ui.Text(c, "Second panel") }}}, nui.TabsOptions{Content: nui.Set(content), KeepMounted: true, OnChange: func(string) { changes++ }, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Trailing {
				e.Label("Tab badge")
			}
		}}).Width(460)
		nui.Badge(c, th, "0", nui.BadgeOptions{Size: nui.SM, Color: nui.Success, Variant: nui.Soft}).Label("Standalone badge")
	}, 500, 320)
	if bounds(t, tt, "Tab avatar").W != 20 || !tt.HasText("0") || bounds(t, tt, "Tab badge").H != bounds(t, tt, "Standalone badge").H {
		t.Fatalf("avatar=%v badge=%v texts=%v", bounds(t, tt, "Tab avatar"), bounds(t, tt, "Tab badge"), tt.Texts())
	}
	r := bounds(t, tt, "First")
	if p := tt.Image().RGBAAt(int(r.X+8), int(r.Y+r.H-4)); p.R != 255 || p.G != 0 || p.B != 255 {
		t.Fatalf("item trigger background=%v", p)
	}
	click(t, tt, "Tab draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Edited")
	icon = true
	tt.Frame()
	if tt.HasText("Tab avatar") || !tt.Focused("Tab draft") {
		t.Fatal("icon priority update lost content focus")
	}
	content = false
	tt.Frame()
	if tt.HasText("Tab draft") || !tt.Focused("First") || value != "a" || draft != "Edited" || changes != 0 {
		t.Fatal("content switch did not unmount panels and repair focus without changing value")
	}
	content = true
	tt.Frame()
	if !tt.HasText("Tab draft") || draft != "Edited" {
		t.Fatal("content recreation lost application data")
	}
}

func TestTabsPanelFocusControlledSelectionAndRemoval(t *testing.T) {
	value, draft := "a", "Draft"
	remove, disabled, vertical := false, false, false
	changes, notices := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		items := []nui.TabItem{{ID: "a", Label: "First", Disabled: disabled, Content: func(c *ui.Context, th nui.Theme) { nui.Input(c, th, &draft, nui.InputOptions{Label: "Draft"}) }}, {ID: "b", Label: "Second", Content: func(c *ui.Context, th nui.Theme) { ui.Text(c, "Static panel") }}}
		if remove {
			items = items[1:]
		}
		nui.Tabs(c, th, &value, items, nui.TabsOptions{Vertical: vertical, KeepMounted: true, OnChange: func(string) { changes++ }}).OnChange(func() { notices++ })
	}, 500, 300)
	click(t, tt, "First")
	tt.Key(0, ui.KeyTab)
	// The panel is named after its tab, so find it by role, not by name alone.
	if panel, ok := tt.AccessByRole(ui.RoleTabPanel, "First"); !ok || !panel.Focused {
		t.Fatalf("Tab did not reach the content panel: %+v", panel)
	}
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Draft") {
		t.Fatal("panel input was not the next focus stop")
	}
	value = "b"
	tt.Frame()
	if !tt.Focused("Second") || changes != 0 || notices != 0 {
		t.Fatal("external switch did not restore trigger focus silently")
	}
	click(t, tt, "First")
	if changes != 1 || notices != 1 {
		t.Fatalf("click notifications=%d/%d", changes, notices)
	}
	disabled = true
	tt.Frame()
	if value != "b" || !tt.Focused("Second") || changes != 1 {
		t.Fatal("disabled selection was not normalized with focus repair")
	}
	disabled = false
	tt.Frame()
	click(t, tt, "First")
	remove = true
	tt.Frame()
	if value != "b" || !tt.Focused("Second") {
		t.Fatal("removed trigger did not restore focus")
	}
	vertical = true
	tt.Frame()
	if !tt.Focused("Second") {
		t.Fatal("orientation changed trigger identity")
	}
}

func TestTabsNeutralIndicatorAndPartOverride(t *testing.T) {
	for _, underline := range []bool{false, true} {
		t.Run(fmt.Sprint(underline), func(t *testing.T) {
			value := "a"
			override := false
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				th.InvertedBackground = ui.RGB(255, 0, 255)
				styles := nui.Styles{}
				if override {
					styles[nui.Indicator] = nui.PartStyle{Background: nui.Set(ui.RGB(0, 190, 255))}
				}
				nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}}, nui.TabsOptions{Color: nui.Neutral, Underline: underline, UI: styles, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Indicator {
						e.Label("Indicator")
					}
				}}).Width(300)
			}, 340, 200)
			r := bounds(t, tt, "Indicator")
			x, y := int(r.X+r.W/2), int(r.Y+r.H/2)
			if !underline {
				y = int(r.Y + r.H - 4)
			}
			if p := tt.Image().RGBAAt(x, y); p.R != 255 || p.G != 0 || p.B != 255 {
				t.Fatalf("neutral indicator=%v bounds=%v", p, r)
			}
			if !underline && countColor(tt.Image(), bounds(t, tt, "First"), rgba(nui.Light().Inverted)) == 0 {
				t.Fatal("pill indicator covers the active tab text")
			}
			override = true
			tt.Frame()
			if p := tt.Image().RGBAAt(x, y); p.R != 0 || p.G != 190 || p.B != 255 {
				t.Fatalf("indicator override=%v", p)
			}
		})
	}
}

func TestTabsBackgroundUpdatesDoNotActivateFocusedID(t *testing.T) {
	value := "a"
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.Tabs(c, th, &value, []nui.TabItem{{ID: "a", Label: "First"}, {ID: "b", Label: "Second"}}, nui.TabsOptions{OnChange: func(string) { changes++ }})
	}, 400, 240)
	click(t, tt, "First")
	tt.SetFocused(false)
	value = "b"
	tt.Frame()
	tt.SetFocused(true)
	if value != "b" || changes != 0 {
		t.Fatal("window reactivation overwrote controlled selection")
	}
}

func TestTabsVerticalCenterRTLAndRetainedDraftSelection(t *testing.T) {
	value, draft := "a", "Selected text"
	vertical := true
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.DarkMode})
		nui.Tabs(c.Key("tabs"), th, &value, []nui.TabItem{{ID: "a", Label: "First", Content: func(c *ui.Context, th nui.Theme) {
			nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Draft"})
			ui.Box(c).Height(80)
		}}, {ID: "b", Label: "Second"}}, nui.TabsOptions{Label: "List", Vertical: vertical, Underline: true, KeepMounted: true, UI: nui.Styles{nui.Indicator: {Background: nui.Set(ui.RGB(255, 0, 255))}}, Style: func(slot nui.Slot, e ui.Element) {
			if slot == nui.Indicator {
				e.Label("Indicator")
			}
		}}).Width(440)
	}, 500, 360)
	first, ok := tt.AccessByRole(ui.RoleTabPanel, "First")
	if !ok {
		t.Fatal("missing the First tab panel")
	}
	list, panel := bounds(t, tt, "List"), first.Bounds
	if list.Y+list.H/2 != panel.Y+panel.H/2 {
		t.Fatalf("vertical regions not centered: list=%v panel=%v", list, panel)
	}
	click(t, tt, "Draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.SetDirection(ui.RightToLeft)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5})
	line := bounds(t, tt, "Indicator")
	list = bounds(t, tt, "List")
	if line.X+line.W != list.X+list.W {
		t.Fatalf("RTL underline did not use start edge: line=%v list=%v", line, list)
	}
	vertical = false
	tt.Frame()
	tt.Type("Replacement")
	if draft != "Replacement" || !tt.Focused("Draft") {
		t.Fatalf("layout/scaling lost editor selection: %q", draft)
	}
}
