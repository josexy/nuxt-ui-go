package nui_test

import (
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestBuiltInPalettesMatchPinnedTailwind(t *testing.T) {
	primary := strings.Fields("red orange amber yellow lime green emerald teal cyan sky blue indigo violet purple fuchsia pink rose")
	neutral := strings.Fields("slate gray zinc neutral stone taupe mauve mist olive")
	if !slices.Equal(nui.PrimaryPaletteNames(), primary) || !slices.Equal(nui.NeutralPaletteNames(), neutral) {
		t.Fatal("palette choices differ from the Nuxt UI theme menu")
	}
	css, err := os.ReadFile("../third_party/tailwind/theme.css")
	if err != nil {
		t.Fatal(err)
	}
	// Compare every generated shade with the independent vendored CSS source,
	// including the achromatic `none` hue used by Zinc, Neutral and Mauve.
	declarations := regexp.MustCompile(`--color-([a-z]+)-(\d+): oklch\(([^)]+)\)`).FindAllStringSubmatch(string(css), -1)
	counts := map[string]int{}
	for _, declaration := range declarations {
		name := declaration[1]
		palette, ok := nui.PaletteByName(name)
		if !ok {
			t.Fatalf("missing Tailwind palette %q", name)
		}
		shade, err := strconv.Atoi(declaration[2])
		if err != nil {
			t.Fatal(err)
		}
		components := strings.Fields(declaration[3])
		var values [3]float64
		for i, component := range components {
			if component == "none" {
				continue
			}
			values[i], err = strconv.ParseFloat(strings.TrimSuffix(component, "%"), 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		want := ui.Oklch(float32(values[0]/100), float32(values[1]), float32(values[2]))
		if got := palette.Shade(shade); got != want {
			t.Errorf("%s-%d: got %v, want %v", name, shade, got, want)
		}
		counts[name]++
	}
	for _, name := range append(primary, neutral...) {
		if counts[name] != 11 {
			t.Errorf("%s: checked %d shades, want 11", name, counts[name])
		}
	}
	for _, name := range []string{"", "black", "white", "Blue", "unknown"} {
		if p, ok := nui.PaletteByName(name); ok || p != (nui.Palette{}) {
			t.Errorf("unexpected palette for %q", name)
		}
	}
}

// A zero AppConfig resolves both built-in themes in every build: their
// palettes are converted once, and stay the generated ones.
func TestDefaultPalettesAreConvertedOnce(t *testing.T) {
	want := nui.Palettes{nui.GreenPalette(), nui.BluePalette(), nui.GreenPalette(), nui.BluePalette(), nui.YellowPalette(), nui.RedPalette(), nui.SlatePalette()}
	got := nui.DefaultPalettes()
	if got != want {
		t.Fatal("the default palettes differ from the generated ones")
	}
	got.Primary[0] = ui.RGB(1, 2, 3) // a copy: no caller changes another's
	if nui.DefaultPalettes() != want {
		t.Fatal("changing a returned copy changed the defaults")
	}
	if n := testing.AllocsPerRun(10, func() { _ = nui.AppConfig{}.Resolve(false, ui.Preferences{}) }); n != 0 {
		t.Fatalf("resolving a zero AppConfig allocated %.0f times", n)
	}
}

// BenchmarkZeroAppConfig is the theme a zero AppConfig resolves in each build.
func BenchmarkZeroAppConfig(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = nui.AppConfig{}.Resolve(false, ui.Preferences{})
	}
}
