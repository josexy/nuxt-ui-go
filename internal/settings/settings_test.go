package settings

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func press(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
}
func TestSettingsSaveValidationAndSections(t *testing.T) {
	a := New()
	tt := ui.NewTester(a.View, 960, 900)
	press(t, tt, "Clear display name")
	tt.Scroll(800, 400, 0, 800)
	press(t, tt, "Save changes")
	if a.Saved != 0 || !a.Attempted || a.Message != "Review the highlighted fields." {
		t.Fatal("invalid form was saved")
	}
	tt.Scroll(800, 400, 0, -2000)
	press(t, tt, "Display name *")
	tt.Type("Updated name")
	tt.Scroll(800, 400, 0, 1000)
	press(t, tt, "Save changes")
	if a.Saved != 1 {
		t.Fatal("valid settings not saved")
	}
	press(t, tt, "Notifications")
	press(t, tt, "I accept the workspace policy")
	tt.Scroll(800, 400, 0, 800)
	press(t, tt, "Save changes")
	if a.Saved != 1 || !tt.HasText("Accept the workspace policy to save.") {
		t.Fatal("required checkbox was not validated")
	}
	press(t, tt, "I accept the workspace policy")
	tt.Scroll(800, 400, 0, -2000)
	press(t, tt, "Enable notifications")
	if a.Notify {
		t.Fatal("switch state")
	}
	press(t, tt, "Desktop")
	if len(a.Channels) != 1 {
		t.Fatal("dependent disabled channel changed")
	}
	press(t, tt, "Workspace")
	press(t, tt, "Region")
	tt.Key(0, ui.KeyEnd)
	tt.Key(0, ui.KeyEnter)
	if a.Region != nui.Set("ap") {
		t.Fatalf("region %v", a.Region)
	}
	press(t, tt, "Workspace owner")
	tt.Type("Taipei")
	tt.Key(0, ui.KeyEnter)
	if a.Owner.Selected != "alex-taipei" {
		t.Fatal("owner identity")
	}
	tt.Scroll(800, 400, 0, 900)
	press(t, tt, "About workspace access")
	if !tt.HasText("Got it") {
		t.Fatal("help popup")
	}
	tt.Key(0, ui.KeyEscape)
}

func TestSettingsNarrowTextScaleAndIsolation(t *testing.T) {
	a, b := New(), New()
	first, second := ui.NewTester(a.View, 720, 900), ui.NewTester(b.View, 720, 900)
	first.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	for _, label := range []string{"Notifications", "Workspace", "Profile"} {
		press(t, first, label)
		first.Scroll(600, 400, 0, 2000)
		press(t, first, "Save changes")
	}
	if a.Saved != 3 || b.Saved != 0 {
		t.Fatal("window model leaked")
	}
	second.Frame()
	for _, width := range []int{1440, 720, 960} {
		first.SetSize(width, 900)
		first.Scroll(float32(width-100), 400, 0, -4000)
		press(t, first, "Display name *")
		first.Key(ui.Ctrl, ui.KeyA)
		first.Type(fmt.Sprintf("Name %d", width))
	}
	if a.Name != "Name 960" || b.Name != "Lin Chen / 林晓" {
		t.Fatal("responsive editor state")
	}
}

func TestSettingsVisualArtifacts(t *testing.T) {
	dir := os.Getenv("NUI_ARTIFACTS")
	if dir == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, scale := range []float32{1, 1.25, 1.5, 2} {
			for page := range 4 {
				a := New()
				a.Page = page
				if dark {
					a.Config.Mode = nui.DarkMode
				} else {
					a.Config.Mode = nui.LightMode
				}
				tt := ui.NewTester(a.View, 960, 1000)
				tt.SetScale(scale)
				f, err := os.Create(filepath.Join(dir, fmt.Sprintf("settings-dark-%v-page-%d-scale-%g.png", dark, page, scale)))
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
	}
	a := New()
	tt := ui.NewTester(a.View, 720, 900)
	tt.SetPreferences(ui.Preferences{TextScale: 1.5, ReduceMotion: true})
	f, err := os.Create(filepath.Join(dir, "settings-narrow-text-150.png"))
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
