package lucide

import (
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAssetsParseAndReuseIdentity(t *testing.T) {
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct{ Icons []struct{ Name string } }
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Icons) != len(generatedIconCases) {
		t.Fatal("generated getter coverage differs from manifest")
	}
	seen := map[string]bool{}
	for _, icon := range generatedIconCases {
		if seen[icon.name] {
			t.Fatal("duplicate generated getter case", icon.name)
		}
		seen[icon.name] = true
		t.Run(icon.name, func(t *testing.T) {
			a, b := icon.get(), icon.get()
			if a != b {
				t.Fatal("cached SVG identity changed")
			}
			w, h := a.Size()
			if w != 24 || h != 24 {
				t.Fatalf("unexpected viewbox %v x %v", w, h)
			}
		})
	}
	for _, icon := range manifest.Icons {
		if !seen[icon.Name] {
			t.Fatal("missing getter test", icon.Name)
		}
	}
}

// Parsing alone cannot detect an SVG whose shapes are unsupported by the renderer.
func TestEveryIconRenders(t *testing.T) {
	const columns, cell, size = 32, 40, 24
	rows := (len(generatedIconCases) + columns - 1) / columns
	for _, dark := range []bool{false, true} {
		name := "light"
		foreground, background := ui.RGB(0, 0, 0), ui.RGB(255, 255, 255)
		if dark {
			name, foreground, background = "dark", background, foreground
		}
		t.Run(name, func(t *testing.T) {
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Box(c).Fill().Background(background).Children(func() {
					for i, icon := range generatedIconCases {
						ui.Icon(c.Key(icon.name), icon.get()).Absolute().Left(float32(i%columns*cell+8)).Top(float32(i/columns*cell+8)).Size(size, size).TextColor(foreground)
					}
				})
			}, columns*cell, rows*cell)
			im := tt.Image()
			for i, icon := range generatedIconCases {
				ink := 0
				for y := i / columns * cell; y < (i/columns+1)*cell; y++ {
					for x := i % columns * cell; x < (i%columns+1)*cell; x++ {
						pixel := im.RGBAAt(x, y)
						if pixel.A > 0 && ((!dark && pixel.R < 128) || (dark && pixel.R > 128)) {
							ink++
						}
					}
				}
				if ink == 0 {
					t.Errorf("%s produced no visible pixels", icon.name)
				}
			}
			if dir := os.Getenv("NUI_ICON_ARTIFACTS"); dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				file, err := os.Create(filepath.Join(dir, "lucide-all-"+name+".png"))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(file, im)
				closeErr := file.Close()
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

func BenchmarkParseSVG(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if _, err := ui.ParseSVG(checkSource); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCachedSVG(b *testing.B) {
	Check()
	b.ReportAllocs()
	for b.Loop() {
		Check()
	}
}
