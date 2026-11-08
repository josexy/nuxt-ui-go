package preview

import "testing"

func TestKbdPreviewPlatformAndExplicitSymbols(t *testing.T) {
	tester, cfg := exampleTester(t, "kbd")
	cfg.Values["platform"] = "macos"
	tester.Frame()
	if !tester.HasText("⌘") || tester.HasText("CTRL") {
		t.Fatal("Platform control did not change the primary modifier")
	}
	cfg.Options = map[string]any{"KbdItems": []any{"command", "win", "f12"}}
	tester.Frame()
	for _, text := range []string{"⌘", "⊞", "F12"} {
		if !tester.HasText(text) {
			t.Fatalf("configured key %q missing", text)
		}
	}
	cfg.Values["platform"] = "windows"
	tester.Frame()
	if !tester.HasText("⌘") || !tester.HasText("⊞") {
		t.Fatal("explicit symbols changed with Platform")
	}
}

func TestNeutralSolidPreviewThemeControl(t *testing.T) {
	for _, id := range []string{"button", "badge", "alert", "kbd"} {
		t.Run(id, func(t *testing.T) {
			tester, cfg := exampleTester(t, id)
			cfg.Values["color"], cfg.Values["variant"], cfg.Values["customNeutral"] = "neutral", "solid", true
			for _, mode := range []string{"light", "dark"} {
				cfg.Mode = mode
				tester.Frame()
				pixels := tester.Image().Pix
				r, g, b := byte(126), byte(34), byte(206)
				if mode == "dark" {
					r, g, b = 192, 132, 252
				}
				count := 0
				for i := 0; i < len(pixels); i += 4 {
					if pixels[i] == r && pixels[i+1] == g && pixels[i+2] == b && pixels[i+3] == 255 {
						count++
					}
				}
				if count == 0 {
					t.Fatalf("%s preview ignored its custom inverted background", mode)
				}
			}
		})
	}
}
