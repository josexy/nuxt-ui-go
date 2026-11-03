package nui_test

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

var sidebarItems = []nui.NavigationItem{
	{ID: "home", Label: "Home", Group: "Workspace", Icon: lucide.House()},
	{ID: "settings", Label: "Settings", Icon: lucide.Settings(), Children: []nui.NavigationItem{
		{ID: "members", Label: "Members"},
		{ID: "blocked", Label: "Unavailable", Disabled: true},
		{ID: "team", Label: "Team", Children: []nui.NavigationItem{{ID: "roles", Label: "Roles"}}},
	}},
}

func TestSidebarResizeCollapseAndPersistence(t *testing.T) {
	for _, dir := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
		for _, side := range []nui.Side{nui.SideLeft, nui.SideRight} {
			t.Run(fmt.Sprint(dir, side), func(t *testing.T) {
				var state nui.SidebarState
				collapsed, mounted, disabled := false, true, false
				resizes, collapses := 0, 0
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{})
					ui.Row(c).Fill().Children(func() {
						if side == nui.SideRight {
							ui.Box(c.Key("spacer")).Grow(1)
						}
						if mounted {
							nui.Sidebar(c.Key("sidebar"), th, &collapsed, nui.SidebarOptions{Title: "Files", State: &state, Side: nui.Set(side), Width: nui.Set(float32(220)), MinWidth: nui.Set(float32(160)), MaxWidth: nui.Set(float32(360)), Resizable: true, Disabled: disabled, OnResize: func(float32) { resizes++ }, OnCollapsedChange: func(bool) { collapses++ }}, func(c *ui.Context, th nui.Theme, compact bool) {
								nui.Button(c, th, "Entry", nui.ButtonOptions{})
							})
						}
					})
				}, 900, 600)
				tt.SetDirection(dir)
				right := (side == nui.SideRight) != (dir == ui.RightToLeft)
				grow := ui.KeyRight
				delta := float32(1)
				if right {
					grow, delta = ui.KeyLeft, -1
				}
				click(t, tt, "Resize sidebar")
				tt.Key(0, grow)
				if state.Width.Value != 230 || resizes != 1 {
					t.Fatalf("keyboard resize: %+v, callbacks %d", state, resizes)
				}
				click(t, tt, "Toggle sidebar")
				if !collapsed || bounds(t, tt, "Files").W != 64 {
					t.Fatal("collapse width")
				}
				click(t, tt, "Toggle sidebar")
				if collapsed || bounds(t, tt, "Files").W != 230 || collapses != 2 {
					t.Fatal("restore width")
				}
				r := bounds(t, tt, "Resize sidebar")
				x, y := r.X+r.W/2, r.Y+r.H/2
				tt.Press(x, y)
				tt.Move(x+40*delta, y)
				tt.Move(x+80*delta, y)
				tt.Release(x+80*delta, y)
				if state.Width.Value != 310 {
					t.Fatalf("pointer resize: %v", state.Width)
				}
				// Drag through the collapse threshold, then restore the remembered size.
				r = bounds(t, tt, "Resize sidebar")
				x = r.X + r.W/2
				tt.Press(x, y)
				tt.Move(x-260*delta, y)
				tt.Release(x-260*delta, y)
				if !collapsed || state.Width.Value != 310 {
					t.Fatal("drag collapse discarded width")
				}
				click(t, tt, "Toggle sidebar")
				if bounds(t, tt, "Files").W != 310 {
					t.Fatal("drag restore width")
				}
				click(t, tt, "Resize sidebar")
				tt.Key(0, ui.KeyEnd)
				if state.Width.Value != 360 {
					t.Fatal("max constraint")
				}
				before := resizes
				tt.Key(0, grow)
				if resizes != before {
					t.Fatal("callback at unchanged maximum")
				}
				tt.Key(0, ui.KeyHome)
				if !collapsed {
					t.Fatal("Home should collapse")
				}
				tt.Key(0, ui.KeyEnter)
				if collapsed || state.Width.Value != 360 {
					t.Fatal("Enter should restore")
				}
				mounted = false
				tt.Frame()
				mounted = true
				tt.Frame()
				if bounds(t, tt, "Files").W != 360 {
					t.Fatal("unmount lost external width")
				}
				disabled = true
				tt.Frame()
				before = resizes
				click(t, tt, "Resize sidebar")
				tt.Key(0, ui.KeyHome)
				click(t, tt, "Toggle sidebar")
				if collapsed || resizes != before {
					t.Fatal("disabled sidebar changed")
				}
			})
		}
	}
}

func TestSidebarCompactNavigation(t *testing.T) {
	var state nui.SidebarState
	var nav nui.NavigationState
	active := "home"
	collapsed := false
	selections := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Fill().Children(func() {
			nui.Sidebar(c.Key("sidebar"), th, &collapsed, nui.SidebarOptions{Title: "Files", State: &state}, func(c *ui.Context, th nui.Theme, compact bool) {
				nui.NavigationMenu(c.Key("nav"), th, &active, &nav, sidebarItems, nui.NavigationOptions{Vertical: true, Compact: compact, OnNavigate: func(string) { selections++ }})
			})
			nui.Button(c, th, "Outside", nui.ButtonOptions{})
		})
	}, 900, 600)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1})
	click(t, tt, "Settings")
	click(t, tt, "Members")
	// An application-controlled collapse while a child has focus moves focus to its parent.
	collapsed = true
	tt.Frame()
	if !tt.Focused("Settings") {
		t.Fatal("collapse lost child focus")
	}
	if tt.HasText("Members") {
		t.Fatal("compact children should not be inline")
	}
	tt.Key(0, ui.KeyRight)
	if !tt.HasText("Members") {
		t.Fatal("keyboard failed to open compact popover")
	}
	if !tt.Focused("Members") {
		t.Fatal("keyboard opening did not focus first child")
	}
	tt.Key(0, ui.KeyEscape)
	if !tt.Focused("Settings") {
		t.Fatal("Escape lost trigger focus")
	}
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEnter)
	tt.Advance(250 * time.Millisecond)
	if tt.HasText("Members") || !tt.Focused("Settings") {
		t.Fatalf("keyboard selection popup visible=%v, trigger focused=%v, selections=%d", tt.HasText("Members"), tt.Focused("Settings"), selections)
	}
	// Choosing the already active leaf still emits exactly one navigation action.
	if selections != 2 {
		t.Fatal("keyboard navigation callback count")
	}
	tt.Key(0, ui.KeyRight)
	click(t, tt, "Unavailable")
	if active != "members" || selections != 2 {
		t.Fatal("disabled navigation activated")
	}
	click(t, tt, "Team")
	click(t, tt, "Roles")
	if active != "roles" || selections != 3 {
		t.Fatal("nested popover navigation")
	}
	if !tt.Focused("Settings") {
		t.Fatal("selection failed to restore trigger focus")
	}
	tt.Key(0, ui.KeyEscape)
	click(t, tt, "Outside")
	// Hover opens a side-facing submenu and allows crossing the anchor/panel gap.
	r := bounds(t, tt, "Settings")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(200 * time.Millisecond)
	if !tt.HasText("Members") {
		t.Fatal("hover failed to open submenu")
	}
	members := bounds(t, tt, "Members")
	if members.X < r.X+r.W {
		t.Fatalf("popover not beside sidebar: %v %v", r, members)
	}
	tt.Move(members.X+4, members.Y+4)
	tt.Advance(200 * time.Millisecond)
	if !tt.HasText("Members") {
		t.Fatal("crossing popup gap dismissed it")
	}
	tt.Key(0, ui.KeyEscape)
	click(t, tt, "Outside")
	r = bounds(t, tt, "Home")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(time.Second)
	if !tt.HasText("Home") {
		t.Fatal("compact item has no tooltip")
	}
	collapsed = false
	tt.Frame()
	if !tt.HasText("Members") {
		t.Fatal("expansion state lost after compact navigation")
	}
}

func TestSidebarControlledSlotsAndZeroWidth(t *testing.T) {
	state := nui.SidebarState{}
	collapsed, toggle, collapsible := false, true, true
	var draft string
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Row(c).Fill().Children(func() {
			nui.Sidebar(c.Key("sidebar"), th, &collapsed, nui.SidebarOptions{Title: "Files", State: &state, Toggle: nui.Set(toggle), Collapsible: nui.Set(collapsible), CollapsedWidth: nui.Set(float32(0)), Resizable: true, HeaderContent: func(c *ui.Context, th nui.Theme, scope nui.SidebarScope) {
				nui.Input(c.Key("search"), th, &draft, nui.InputOptions{Label: "Search"})
			}}, nil)
		})
	}, 900, 600)
	click(t, tt, "Search")
	tt.Type("retained")
	state.Width = nui.Set(float32(320))
	tt.Frame()
	if !tt.Focused("Search") || draft != "retained" {
		t.Fatal("width change lost editor state")
	}
	collapsed = true
	tt.Frame()
	if r := bounds(t, tt, "Files"); r.W != 0 {
		t.Fatal("explicit zero width ignored")
	}
	if _, ok := tt.Find("Search"); ok {
		t.Fatal("zero-width sidebar still builds content")
	}
	collapsed, toggle, collapsible = false, false, false
	tt.Frame()
	if _, ok := tt.Find("Toggle sidebar"); ok {
		t.Fatal("toggle disabled")
	}
	click(t, tt, "Resize sidebar")
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	if collapsed || state.Width.Value != 160 {
		t.Fatal("non-collapsible resize should stop at minimum")
	}
}

func TestSidebarResponsiveModes(t *testing.T) {
	for _, mode := range []nui.SidebarMode{nui.SidebarSlideover, nui.SidebarModal, nui.SidebarDrawer} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			var state nui.SidebarState
			collapsed := true
			active, draft := "home", "draft"
			changes := 0
			var scopes []nui.SidebarScope
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Row(c).Fill().Children(func() {
					nui.Sidebar(c.Key("sidebar"), th, &collapsed, nui.SidebarOptions{Title: "Files", State: &state, Breakpoint: nui.Set(float32(700)), Mode: mode, NavigationKey: active, OnOpenChange: func(bool) { changes++ }, HeaderContent: func(c *ui.Context, th nui.Theme, scope nui.SidebarScope) {
						scopes = append(scopes, scope)
						ui.Text(c, "Header")
					}, FooterContent: func(c *ui.Context, th nui.Theme, scope nui.SidebarScope) {
						nui.Button(c, th, "Close from footer", nui.ButtonOptions{}).OnClick(func() { scope.SetOpen(false) })
					}}, func(c *ui.Context, th nui.Theme, compact bool) {
						if !compact {
							nui.Input(c.Key("draft"), th, &draft, nui.InputOptions{Label: "Draft"})
							nui.Button(c.Key("navigate"), th, "Navigate", nui.ButtonOptions{}).OnClick(func() { active = "members" })
						}
					})
					nui.Button(c.Key("outside"), th, "Outside", nui.ButtonOptions{})
				})
			}, 600, 600)
			tt.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1})
			click(t, tt, "Toggle sidebar")
			if !state.Open || changes != 1 || !tt.HasText("Navigate") {
				t.Fatalf("open: %+v callbacks %d", state, changes)
			}
			if footer := bounds(t, tt, "Close from footer"); footer.Y < 500 {
				t.Fatalf("footer is not pinned below the scrolling body: %v", footer)
			}
			if last := scopes[len(scopes)-1]; !last.Mobile || last.Collapsed {
				t.Fatal("overlay scope must be expanded")
			}
			click(t, tt, "Draft")
			tt.Key(ui.Ctrl, ui.KeyA)
			tt.Type("kept")
			click(t, tt, "Navigate")
			tt.Frame()
			if state.Open || changes != 2 || !tt.Focused("Toggle sidebar") {
				t.Fatal("navigation should close and restore focus")
			}
			click(t, tt, "Toggle sidebar")
			tt.Key(0, ui.KeyEscape)
			if state.Open || changes != 4 {
				t.Fatalf("Escape: open %v callbacks %d", state.Open, changes)
			}
			click(t, tt, "Toggle sidebar")
			click(t, tt, "Close from footer")
			if state.Open || changes != 6 {
				t.Fatal("footer close")
			}
			click(t, tt, "Toggle sidebar")
			tt.SetSize(900, 600)
			tt.Frame()
			if state.Open || !collapsed || draft != "kept" {
				t.Fatal("breakpoint lost model state or left overlay open")
			}
			if _, ok := tt.Find("Close sidebar"); ok {
				t.Fatal("overlay remained mounted on desktop")
			}
			click(t, tt, "Outside")
			if !tt.Focused("Outside") {
				t.Fatal("modal focus scope leaked")
			}
		})
	}
}

func TestSidebarExternalOpenerAndAnimatedClose(t *testing.T) {
	var state nui.SidebarState
	var opener ui.Handle
	collapsed, mounted := false, true
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Fill().Children(func() {
			nui.Button(c.Key("opener"), th, "Open navigation", nui.ButtonOptions{}).Bind(&opener).OnClick(func() { state.Open = true })
			if mounted {
				nui.Sidebar(c.Key("sidebar"), th, &collapsed, nui.SidebarOptions{State: &state, Title: "Files", Breakpoint: nui.Set(float32(700)), Toggle: nui.Set(false), ReturnFocus: &opener, OnOpenChange: func(bool) { changes++ }}, func(c *ui.Context, th nui.Theme, compact bool) {
					nui.Button(c, th, "Link", nui.ButtonOptions{})
				})
			}
		})
	}, 600, 600)
	click(t, tt, "Open navigation")
	tt.Advance(250 * time.Millisecond)
	click(t, tt, "Link")
	tt.Key(0, ui.KeyEscape)
	if state.Open || changes != 1 {
		t.Fatal("animated dismissal did not update external state exactly once")
	}
	tt.Advance(300 * time.Millisecond)
	if !tt.Focused("Open navigation") {
		t.Fatal("animated close did not restore external opener")
	}
	if tt.FramePending() {
		t.Fatal("closed sidebar continues scheduling frames")
	}
	click(t, tt, "Open navigation")
	tt.Advance(250 * time.Millisecond)
	mounted = false
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	click(t, tt, "Open navigation")
	if !tt.Focused("Open navigation") {
		t.Fatal("unmount left a modal input scope")
	}
}

func TestSidebarVisualArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, scene := range []string{"expanded", "compact", "popover", "mobile", "rtl"} {
			var nav nui.NavigationState
			state := nui.SidebarState{Open: scene == "mobile"}
			collapsed := scene == "compact" || scene == "popover"
			active := "members"
			tt := ui.NewTester(func(c *ui.Context) {
				mode := nui.LightMode
				if dark {
					mode = nui.DarkMode
				}
				th := nui.App(c, nui.AppConfig{Mode: mode})
				ui.Row(c).Fill().Children(func() {
					nui.Sidebar(c.Key("sidebar"), th, &collapsed, nui.SidebarOptions{Title: "Workspace", State: &state, Resizable: true, Breakpoint: nui.Set(float32(700)), FooterContent: func(c *ui.Context, th nui.Theme, scope nui.SidebarScope) {
						label := "Alex Morgan"
						if scope.Collapsed {
							label = ""
						}
						nui.Button(c, th, label, nui.ButtonOptions{Leading: lucide.User(), AccessibleLabel: "Account", Variant: nui.Ghost}).FillWidth()
					}}, func(c *ui.Context, th nui.Theme, compact bool) {
						nui.NavigationMenu(c.Key("nav"), th, &active, &nav, sidebarItems, nui.NavigationOptions{Vertical: true, Compact: compact})
					})
					ui.Column(c).Grow(1).Padding(32).Gap(16).Children(func() {
						ui.Text(c, "Workspace overview").FontSize(24).FontWeight(600)
						ui.Text(c, "Projects, people, and settings.").TextColor(th.Muted)
					})
				})
			}, 1000, 640)
			tt.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1})
			if scene == "rtl" {
				tt.SetDirection(ui.RightToLeft)
				tt.SetPreferences(ui.Preferences{ReduceMotion: true, TextScale: 1.5})
			}
			if scene == "mobile" {
				tt.SetSize(600, 640)
			}
			if scene == "popover" {
				click(t, tt, "Settings")
			}
			tt.Frame()
			file, err := os.Create(filepath.Join(dir, fmt.Sprintf("sidebar-%s-dark-%v.png", scene, dark)))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, tt.Image())
			closeErr := file.Close()
			if err != nil {
				t.Fatal(err)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
		}
	}
}
