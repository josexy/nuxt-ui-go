package gallery

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestGalleryAppPreviewsDoNotChangeOuterTheme(t *testing.T) {
	a := newStaticGallery()
	t.Cleanup(a.Dispose)
	for _, id := range []string{"app", "theme"} {
		for _, s := range a.page(id).Scenes {
			tt := ui.NewTester(func(c *ui.Context) {
				theme := nui.App(c, a.themeConfig())
				outer := c.Theme()
				ui.Column(c).Fill().Children(func() { s.Build(c, theme) })
				if c.Theme() != outer {
					t.Fatalf("%s/%s changed the outer theme scope", id, s.ID)
				}
			}, 1080, 820)
			tt.Frame()
			if a.Dark || a.Compact || a.Primary != "green" {
				t.Fatalf("%s/%s changed window appearance", id, s.ID)
			}
		}
	}
}

func TestGalleryPaletteMenusApplyThemeAndModes(t *testing.T) {
	a := newStaticGallery()
	var applied ui.Theme
	tt := ui.NewTester(func(c *ui.Context) {
		a.View(c)
		applied = *c.Theme()
	}, 1080, 820)
	initial := applied
	press(t, tt, "Theme")
	press(t, tt, "Primary palette")
	press(t, tt, "Orange")
	if a.Primary != "orange" || applied.Accent != nui.OrangePalette().Shade(500) || applied.Text != initial.Text || a.ThemeMenu.Open {
		t.Fatal("primary selection did not update the native theme independently")
	}
	if !tt.Focused("Theme") {
		t.Fatal("palette selection did not restore trigger focus")
	}

	// Enter the neutral submenu and choose a new ramp using only the keyboard.
	press(t, tt, "Theme")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyRight)
	tt.Type("mauve")
	tt.Key(0, ui.KeyEnter)
	if a.Neutral != "mauve" || applied.Text != nui.MauvePalette().Shade(700) || applied.Accent != nui.OrangePalette().Shade(500) {
		t.Fatal("neutral selection did not update text independently of the primary")
	}
	selectMode := func(label string) {
		press(t, tt, "Theme")
		press(t, tt, label)
	}
	selectMode("Dark")
	if !applied.Dark || applied.Accent != nui.OrangePalette().Shade(400) || applied.Background != nui.MauvePalette().Shade(900) {
		t.Fatal("dark mode lost the selected palettes")
	}
	selectMode("System")
	for _, dark := range []bool{false, true, false} {
		tt.SetDark(dark)
		shade := 500
		if dark {
			shade = 400
		}
		if applied.Dark != dark || applied.Accent != nui.OrangePalette().Shade(shade) || a.Neutral != "mauve" {
			t.Fatal("system appearance change lost the palette selection")
		}
	}

	press(t, tt, "Theme")
	press(t, tt, "Primary palette")
	press(t, tt, "Black")
	tt.SetPreferences(ui.Preferences{HighContrast: true})
	for _, dark := range []bool{false, true} {
		tt.SetDark(dark)
		want := ui.Hex("#000000")
		if dark {
			want = ui.Hex("#ffffff")
		}
		if applied.Accent != want {
			t.Fatalf("Black must remain monochrome in high contrast, dark=%v: %v", dark, applied.Accent)
		}
	}
	tt.SetPreferences(ui.Preferences{})
	selectMode("Light")
	press(t, tt, "Theme")
	press(t, tt, "Primary palette")
	tt.Key(0, ui.KeyEnd)
	if !tt.Focused("Rose") {
		t.Fatal("last primary color is not keyboard reachable")
	}
	rose, ok := tt.Find("Rose")
	if !ok || rose.Y < 0 || rose.Y+rose.H > 820 {
		t.Fatalf("focused last color was not scrolled into view: %v", rose)
	}
	tt.Key(0, ui.KeyEnter)
	if a.Primary != "rose" || applied.Accent != nui.RosePalette().Shade(500) || a.Neutral != "mauve" {
		t.Fatal("switching away from Black did not restore the selected primary ramp")
	}
	other := newStaticGallery()
	ui.NewTester(func(c *ui.Context) {
		other.View(c)
		if c.Theme().Accent != nui.GreenPalette().Shade(500) || c.Theme().Text != nui.SlatePalette().Shade(700) {
			t.Fatal("theme selection leaked into a second window")
		}
	}, 1080, 820)
}

func TestGalleryPaletteMenuHoverResizeAndReopen(t *testing.T) {
	a := newStaticGallery()
	tt := ui.NewTester(a.View, 1080, 820)
	press(t, tt, "Theme")
	primary, ok := tt.Find("Primary palette")
	if !ok {
		t.Fatal("missing Primary submenu")
	}
	tt.Move(primary.X+8, primary.Y+primary.H/2)
	tt.Advance(200 * time.Millisecond)
	if !tt.HasText("Emerald") {
		t.Fatal("hover did not open the primary submenu")
	}
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEscape)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyRight)
	for _, width := range []int{720, 1440, 1080} {
		tt.SetSize(width, 600)
		tt.Key(0, ui.KeyEnd)
		if !tt.Focused("Olive") {
			t.Fatal("neutral submenu lost focus across resize")
		}
		item, ok := tt.Find("Olive")
		if !ok || item.X < 0 || item.Y < 0 || item.X+item.W > float32(width) || item.Y+item.H > 600 {
			t.Fatalf("neutral menu escaped the resized window: %v", item)
		}
	}
	tt.Key(0, ui.KeyEnter)
	if a.Neutral != "olive" || a.ThemeMenu.Open {
		t.Fatal("neutral selection failed after resize")
	}
	a.Navigate("theme")
	tt.Frame()
	press(t, tt, "Choose neutral color")
	tt.Type("zinc")
	tt.Key(0, ui.KeyEnter)
	if a.Neutral != "zinc" || !tt.HasText("Neutral: Zinc") {
		t.Fatal("foundation control did not share the selected palette")
	}
	a.Navigate("button")
	tt.Frame()
	press(t, tt, "Theme")
	press(t, tt, "Neutral palette")
	tt.Type("mist")
	tt.Key(0, ui.KeyEnter)
	if a.Neutral != "mist" {
		t.Fatalf("theme menu did not reopen after page changes: neutral=%s open=%v focusedMist=%v texts=%v", a.Neutral, a.ThemeMenu.Open, tt.Focused("Mist"), tt.Texts())
	}
	tt.Advance(time.Second)
	if tt.FramePending() {
		t.Fatal("palette menu did not settle")
	}
}

func TestGalleryThemeMenuArtifacts(t *testing.T) {
	directory := os.Getenv("NUI_ARTIFACTS")
	if directory == "" {
		t.Skip("NUI_ARTIFACTS not set")
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	for _, dark := range []bool{false, true} {
		for _, role := range []string{"Primary", "Neutral"} {
			a := newStaticGallery()
			a.Dark, a.Neutral = dark, "mauve"
			a.Navigate("theme")
			tt := ui.NewTester(a.View, 1080, 820)
			press(t, tt, "Theme")
			press(t, tt, role+" palette")
			f, err := os.Create(filepath.Join(directory, fmt.Sprintf("theme-dark-%v-%s.png", dark, role)))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, tt.Image())
			closeErr := f.Close()
			if err != nil || closeErr != nil {
				t.Fatalf("write image: %v, close: %v", err, closeErr)
			}
		}
	}
}
