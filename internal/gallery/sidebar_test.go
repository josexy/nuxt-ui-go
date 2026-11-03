package gallery

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
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
		t.Fatalf("initial sidebar: %v", r)
	}
	r := galleryBounds(t, tt, "Resize gallery sidebar")
	x, y := r.X+r.W/2, r.Y+r.H/2
	tt.Press(x, y)
	tt.Move(x+44, y)
	tt.Release(x+44, y)
	if a.Sidebar.Width.Value != 300 {
		t.Fatalf("drag did not resize Gallery: %v", a.Sidebar.Width)
	}
	press(t, tt, "Toggle gallery sidebar")
	if !a.SidebarCollapsed || galleryBounds(t, tt, "Gallery sidebar").W != 64 {
		t.Fatal("Gallery did not collapse")
	}
	press(t, tt, "Components")
	tt.Key(0, ui.KeyDown)
	if !tt.Focused("Inputs & overlays") {
		t.Fatal("compact navigation lost arrow-key focus")
	}
	tt.Key(0, ui.KeyEnter)
	if a.Page != 1 {
		t.Fatal("compact keyboard navigation did not switch page")
	}
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Resize gallery sidebar") {
		t.Fatal("Tab did not leave the navigation group for the resize handle")
	}
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Workspace name") {
		t.Fatal("Tab did not enter the page's first input")
	}
	tt.Key(ui.Ctrl, ui.KeyA)
	tt.Type("Preserved draft")
	press(t, tt, "Components")
	press(t, tt, "Inputs & overlays")
	if a.Name != "Preserved draft" || !a.SidebarCollapsed || a.Sidebar.Width.Value != 300 {
		t.Fatalf("page switch: draft=%q collapsed=%v width=%v page=%d", a.Name, a.SidebarCollapsed, a.Sidebar.Width, a.Page)
	}
	press(t, tt, "Toggle gallery sidebar")
	if galleryBounds(t, tt, "Gallery sidebar").W != 300 {
		t.Fatal("expansion did not restore resized width")
	}
	// Commands and other examples also navigate by assigning Page directly.
	a.Page = 3
	tt.Frame()
	if a.navigationActive != "theme-buttons" {
		t.Fatal("external navigation left stale selection")
	}
	// Every item remains reachable through the shared keyboard focus group.
	press(t, tt, "Components")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if a.Page != 18 {
		t.Fatal("last Gallery page is not keyboard reachable")
	}
	other := newStaticGallery()
	t.Cleanup(other.Dispose)
	otherTester := ui.NewTester(other.View, 1080, 820)
	if other.SidebarCollapsed || galleryBounds(t, otherTester, "Gallery sidebar").W != 256 {
		t.Fatal("sidebar state leaked between windows")
	}
}

func TestGallerySidebarIndependentOfLayoutExample(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	a.Page = 16
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Toggle gallery sidebar")
	if !a.SidebarCollapsed || a.Advanced.Collapsed {
		t.Fatal("outer toggle changed the example sidebar")
	}
	press(t, tt, "Toggle sidebar")
	if !a.SidebarCollapsed || !a.Advanced.Collapsed {
		t.Fatal("example toggle changed the Gallery sidebar")
	}
	press(t, tt, "Toggle gallery sidebar")
	if a.SidebarCollapsed || !a.Advanced.Collapsed {
		t.Fatal("Gallery expansion lost example state")
	}
	tt.SetSize(720, 820)
	press(t, tt, "Toggle gallery sidebar")
	press(t, tt, "Components")
	if a.Page != 0 || galleryBounds(t, tt, "Gallery sidebar").W != 64 {
		t.Fatal("narrow-window compact navigation")
	}
}

func TestGallerySidebarVisualArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, scene := range []string{"expanded", "compact", "tooltip", "narrow-dark", "rtl"} {
		a := newStaticGallery()
		t.Cleanup(a.Dispose)
		a.SidebarCollapsed = scene == "compact" || scene == "tooltip" || scene == "narrow-dark"
		a.Dark = scene == "narrow-dark"
		tt := ui.NewTester(a.View, 1080, 820)
		if scene == "narrow-dark" {
			tt.SetSize(720, 820)
		}
		if scene == "rtl" {
			tt.SetDirection(ui.RightToLeft)
			tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
		}
		if scene == "tooltip" {
			r := galleryBounds(t, tt, "Inputs & overlays")
			tt.Move(r.X+r.W/2, r.Y+r.H/2)
			tt.Advance(time.Second)
			if !tt.HasText("Inputs & overlays") {
				t.Fatal("compact navigation tooltip missing")
			}
		}
		f, err := os.Create(filepath.Join(dir, fmt.Sprintf("gallery-sidebar-%s.png", scene)))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, tt.Image())
		closeErr := f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
}
