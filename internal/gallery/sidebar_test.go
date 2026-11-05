package gallery

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func galleryBounds(t *testing.T, tt *ui.Tester, label string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("missing %q", label)
	}
	return r
}
func TestGallerySidebarNavigationAndWidth(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	tt := ui.NewTester(a.View, 1080, 820)
	if r := galleryBounds(t, tt, "Gallery sidebar"); r.W != 256 {
		t.Fatalf("initial width: %v", r)
	}
	r := galleryBounds(t, tt, "Resize gallery sidebar")
	x, y := r.X+r.W/2, r.Y+r.H/2
	tt.Press(x, y)
	tt.Move(x+44, y)
	tt.Release(x+44, y)
	if a.Sidebar.Width.Value != 300 {
		t.Fatal("resize")
	}
	press(t, tt, "Toggle gallery sidebar")
	if !a.SidebarCollapsed || galleryBounds(t, tt, "Gallery sidebar").W != 64 {
		t.Fatal("collapse")
	}
	press(t, tt, "App")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if a.ActiveID != "example/task-browser" {
		t.Fatalf("last item: %s", a.ActiveID)
	}
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyEnter)
	if a.ActiveID != "app" {
		t.Fatal("first item")
	}
	press(t, tt, "Toggle gallery sidebar")
	if galleryBounds(t, tt, "Gallery sidebar").W != 300 {
		t.Fatal("restore width")
	}
	a.Navigate("input")
	tt.Frame()
	press(t, tt, "Workspace name")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Saved draft")
	press(t, tt, "Search components")
	tt.Type("BUTTON")
	if a.ActiveID != "input" || !tt.Focused("Search components") {
		t.Fatal("filter navigated or stole focus")
	}
	requireText(t, tt, "Current value: Saved draft")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("加载")
	if !tt.HasText("Button") {
		t.Fatal("Chinese description search")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("no-such-component")
	requireText(t, tt, "No matching components")
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	if a.ActiveID != "input" {
		t.Fatal("clearing search changed route")
	}
	b := newStaticGallery()
	t.Cleanup(b.Dispose)
	bt := ui.NewTester(b.View, 1080, 820)
	if b.SidebarCollapsed || galleryBounds(t, bt, "Gallery sidebar").W != 256 {
		t.Fatal("window isolation")
	}
}
func TestGallerySidebarIndependentOfLayoutExample(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	a.Navigate("sidebar")
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Toggle gallery sidebar")
	press(t, tt, "Toggle inner slideover sidebar")
	if !a.SidebarCollapsed {
		t.Fatal("inner toggle changed outer sidebar")
	}
	requireText(t, tt, "Collapsed true")
	press(t, tt, "Toggle gallery sidebar")
	requireText(t, tt, "Collapsed true")
	tt.SetSize(720, 820)
	press(t, tt, "Open component directory")
	if !a.Sidebar.Open {
		t.Fatal("narrow opener")
	}
	press(t, tt, "Button")
	if a.ActiveID != "button" || a.Sidebar.Open {
		t.Fatal("narrow navigation did not close directory")
	}
	if !tt.Focused("Open component directory") {
		t.Fatal("directory focus restoration")
	}
}
func TestGallerySidebarVisualArtifacts(t *testing.T) {
	for _, kind := range []string{"compact", "tooltip", "rtl", "narrow-dark"} {
		a := newStaticGallery()
		t.Cleanup(a.Dispose)
		a.SidebarCollapsed = kind != "rtl"
		a.Dark = kind == "narrow-dark"
		tt := ui.NewTester(a.View, 1080, 820)
		if kind == "rtl" {
			tt.SetDirection(ui.RightToLeft)
			tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
		}
		if kind == "narrow-dark" {
			tt.SetSize(720, 820)
			press(t, tt, "Open component directory")
		}
		if kind == "tooltip" {
			r := galleryBounds(t, tt, "Button")
			tt.Move(r.X+r.W/2, r.Y+r.H/2)
			tt.Advance(time.Second)
			requireText(t, tt, "Button")
		}
		captureGallery(t, tt, "sidebar-"+kind)
	}
}
