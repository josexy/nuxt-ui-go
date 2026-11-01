package nui_test

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

// The independent target is the desktop CSS fixture: 20px line + 2*6px padding.
func TestReferenceGeometry(t *testing.T) {
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("dark-%v-scale-%g", dark, scale), func(t *testing.T) {
				value := "Workspace 你好"
				tt := ui.NewTester(func(c *ui.Context) {
					theme := nui.Light()
					if dark {
						theme = nui.Dark()
					}
					nui.Window(c, theme)
					ui.Column(c).Padding(32).Gap(24).AlignItems(ui.Start).Children(func() {
						ui.Text(c, "Nuxt UI 4.12.0 · resolved desktop styles").LineHeight(20.0 / 14)
						ui.Row(c).Gap(16).Children(func() {
							nui.Button(c, theme, "Save changes", nui.ButtonOptions{}).Label("Reference button")
							nui.Button(c, theme, "Outline", nui.ButtonOptions{Variant: nui.Outline})
							nui.Button(c, theme, "Soft", nui.ButtonOptions{Variant: nui.Soft})
							nui.Button(c, theme, "Focus", nui.ButtonOptions{})
						})
						nui.Input(c, theme, &value, nui.InputOptions{Label: "Reference input"}).Width(300).Label("Reference input root")
					})
				}, 800, 220)
				tt.SetScale(scale)
				for _, label := range []string{"Reference button", "Reference input root"} {
					r, ok := tt.Find(label)
					if !ok {
						t.Fatal(label)
					}
					t.Logf("%s: %.3fx%.3f DIP", label, r.W, r.H)
					if r.H != 32 {
						t.Fatalf("%s height %.3f, reference 32", label, r.H)
					}
				}
				for range 4 {
					tt.Key(0, ui.KeyTab)
				}
				if !tt.Focused("Focus") {
					t.Fatal("focus sample")
				}
				if dir := os.Getenv("NUI_ARTIFACTS"); dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					f, err := os.Create(filepath.Join(dir, fmt.Sprintf("reference-dark-%v-scale-%g.png", dark, scale)))
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
			})
		}
	}
}
