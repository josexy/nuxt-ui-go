package preview

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestSidebarExampleNavigationResizeAndFooter(t *testing.T) {
	tt, cfg := exampleTester(t, "sidebar")
	if !tt.HasText("Current page: overview") {
		t.Fatal("initial workspace not shown")
	}
	clickExample(t, tt, "Resize sidebar")
	tt.Key(0, ui.KeyRight)
	if !tt.HasText("Resized to 266 DIP") {
		t.Fatal("resize feedback missing")
	}
	clickExample(t, tt, "Toggle sidebar")
	if !tt.HasText("Collapsed: true") {
		t.Fatal("collapse feedback missing")
	}
	clickExample(t, tt, "Settings")
	clickExample(t, tt, "Members")
	if !tt.HasText("Current page: members") {
		t.Fatal("compact submenu did not navigate")
	}
	clickExample(t, tt, "Account")
	if !tt.HasText("Current page: account") {
		t.Fatal("compact footer did not work")
	}
	clickExample(t, tt, "Toggle sidebar")
	if r, ok := tt.Find("Workspace"); !ok || r.W != 266 {
		t.Fatalf("remembered width lost: %v", r)
	}
	cfg.Values["width"] = float64(300)
	tt.Frame()
	clickExample(t, tt, "Reset width")
	if r, ok := tt.Find("Workspace"); !ok || r.W != 300 {
		t.Fatalf("reset did not use configured width: %v", r)
	}
	cfg.Values["resizable"] = false
	tt.Frame()
	if _, ok := tt.Find("Resize sidebar"); ok {
		t.Fatal("resize switch left the handle enabled")
	}
	cfg.Values["collapsible"] = false
	tt.Frame()
	if _, ok := tt.Find("Toggle sidebar"); ok {
		t.Fatal("collapse switch left the toggle enabled")
	}
	cfg.Values["side"] = "right"
	tt.Frame()
	sidebar, ok := tt.Find("Workspace")
	workspace, found := tt.Find("Workspace overview")
	if !ok || !found || sidebar.X <= workspace.X {
		t.Fatal("right-side quick control did not move the sidebar")
	}
	cfg.Mode = "dark"
	tt.Frame()
	if !tt.HasText("Current page: account") {
		t.Fatal("configuration changes lost navigation state")
	}
}

func TestSidebarExampleResponsiveModes(t *testing.T) {
	for _, mode := range []string{"slideover", "modal", "drawer"} {
		t.Run(mode, func(t *testing.T) {
			tt, cfg := exampleTester(t, "sidebar")
			clickExample(t, tt, "Toggle sidebar")
			cfg.Values["mode"] = mode
			cfg.Values["breakpoint"] = float64(1000)
			tt.SetSize(600, 440)
			clickExample(t, tt, "Toggle sidebar")
			if _, ok := tt.Find("Close sidebar"); !ok {
				t.Fatal("narrow preview did not open")
			}
			clickExample(t, tt, "Projects")
			tt.Advance(300 * time.Millisecond)
			if _, ok := tt.Find("Close sidebar"); ok {
				t.Fatal("navigation did not close the overlay")
			}
			if !tt.HasText("Current page: projects") || !tt.HasText("Collapsed: true") {
				t.Fatal("overlay lost selected page or desktop collapse")
			}
			cfg.Values["breakpoint"] = float64(0)
			tt.Frame()
			if r, ok := tt.Find("Workspace"); !ok || r.W != 64 {
				t.Fatalf("desktop collapse not restored: %v", r)
			}
		})
	}
}
