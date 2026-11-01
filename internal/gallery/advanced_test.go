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

func waitCommands(t *testing.T, tt *ui.Tester, a *State) {
	t.Helper()
	until := time.Now().Add(time.Second)
	for a.Advanced.Command.Results.Loading && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
		tt.Frame()
	}
	if a.Advanced.Command.Results.Loading {
		t.Fatal("command search timeout")
	}
}
func TestGalleryComposition(t *testing.T) {
	a := newStaticGallery()
	a.Page = 14
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Search commands")
	waitCommands(t, tt, a)
	press(t, tt, "Appearance")
	waitCommands(t, tt, a)
	press(t, tt, "Use dark theme")
	if !a.Dark {
		t.Fatal("command theme")
	}
	press(t, tt, "Open inspector")
	press(t, tt, "Resize Activity inspector")
	tt.Key(0, ui.KeyEnd)
	if a.Advanced.Drawer.Snap != 1 {
		t.Fatal("drawer keyboard")
	}
	tt.Key(0, ui.KeyEscape)
	press(t, tt, "Next Review guide")
	if a.Advanced.Carousel.Index != 1 {
		t.Fatal("carousel next")
	}
	a.Page = 15
	tt.Frame()
	press(t, tt, "Area")
	tt.Scroll(900, 600, 0, 500)
	press(t, tt, "Add sample attachment")
	if len(a.Advanced.Files.Files) != 1 {
		t.Fatal("sample attachment")
	}
	a.Page = 16
	tt.Frame()
	press(t, tt, "Workspace settings")
	press(t, tt, "Members")
	if a.Advanced.Active != "members" {
		t.Fatal("shell navigation")
	}
}
func TestGalleryShell(t *testing.T) {
	a := newStaticGallery()
	a.Page = 16
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Workspace settings")
	press(t, tt, "Members")
	if a.Advanced.Active != "members" {
		t.Fatal("nested navigation")
	}
	press(t, tt, "Toggle sidebar")
	if !a.Advanced.Collapsed {
		t.Fatal("collapse")
	}
	press(t, tt, "Workspace actions")
	waitCommands(t, tt, a)
	tt.Key(0, ui.KeyEscape)
	for _, width := range []int{720, 1280} {
		tt.SetSize(width, 900)
		tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
		_ = tt.Image()
	}
}
func TestAdvancedVisualArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, scale := range []float32{1, 1.25, 1.5, 2} {
			for _, scene := range []string{"interactions", "commands", "drawer", "chart-files", "shell", "narrow"} {
				a := newStaticGallery()
				a.Dark = dark
				a.Page = 14
				tt := ui.NewTester(a.View, 1080, 820)
				tt.SetScale(scale)
				switch scene {
				case "commands":
					press(t, tt, "Search commands")
					waitCommands(t, tt, a)
					press(t, tt, "Search Workspace commands")
					tt.Type("act")
					waitCommands(t, tt, a)
				case "drawer":
					press(t, tt, "Open inspector")
				case "chart-files":
					a.Page = 15
					tt.Frame()
					tt.Scroll(900, 600, 0, 500)
					press(t, tt, "Add sample attachment")
				case "shell":
					a.Page = 16
					tt.Frame()
				case "narrow":
					a.Page = 16
					a.Advanced.Collapsed = true
					tt.SetSize(720, 960)
					tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
				}
				tt.Move(0, 0)
				f, err := os.Create(filepath.Join(dir, fmt.Sprintf("advanced-%s-dark-%v-scale-%g.png", scene, dark, scale)))
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
				a.Dispose()
			}
		}
	}
}
