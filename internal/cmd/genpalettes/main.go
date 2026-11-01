// Genpalettes extracts the pinned Tailwind palette data; no network is used.
package main

import (
	"crypto/sha256"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var primaryNames = []string{"red", "orange", "amber", "yellow", "lime", "green", "emerald", "teal", "cyan", "sky", "blue", "indigo", "violet", "purple", "fuchsia", "pink", "rose"}
var neutralNames = []string{"slate", "gray", "zinc", "neutral", "stone", "taupe", "mauve", "mist", "olive"}

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func generate() error {
	data, err := os.ReadFile("third_party/tailwind/theme.css")
	if err != nil {
		return err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "443d7af364200f8ec8352dc78e39485a118a1840d498c97967bf0ae402b39167" {
		return fmt.Errorf("source hash mismatch for Tailwind v4.3.3")
	}
	var out strings.Builder
	out.WriteString("// Code generated from Tailwind CSS v4.3.3 theme.css (MIT). DO NOT EDIT.\n// Source: https://github.com/tailwindlabs/tailwindcss/blob/v4.3.3/packages/tailwindcss/theme.css\npackage nui\nimport \"github.com/egoist/mygo/ui\"\n")
	names := append(append([]string{}, primaryNames...), neutralNames...)
	for _, name := range names {
		fmt.Fprintf(&out, "func %sPalette() Palette { return Palette{", strings.ToUpper(name[:1])+name[1:])
		for _, shade := range []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950} {
			re := regexp.MustCompile(fmt.Sprintf(`--color-%s-%d: oklch\(([\d.]+)%% ([\d.]+) ([\d.]+|none)\);`, name, shade))
			m := re.FindSubmatch(data)
			if len(m) != 4 {
				return fmt.Errorf("missing %s-%d", name, shade)
			}
			l, err := strconv.ParseFloat(string(m[1]), 64)
			if err != nil {
				return err
			}
			hue := string(m[3])
			if hue == "none" {
				// CSS uses a missing hue for achromatic shades.
				hue = "0"
			}
			fmt.Fprintf(&out, "ui.Oklch(%.5f, %s, %s),", l/100, m[2], hue)
		}
		out.WriteString("}}\n")
	}
	for _, group := range []struct {
		name  string
		names []string
	}{{"Primary", primaryNames}, {"Neutral", neutralNames}} {
		fmt.Fprintf(&out, "// %sPaletteNames returns the built-in %s choices in display order.\n", group.name, strings.ToLower(group.name))
		fmt.Fprintf(&out, "// The returned slice belongs to the caller. Black is a mode-dependent override, not a palette.\nfunc %sPaletteNames() []string { return []string{", group.name)
		for _, name := range group.names {
			fmt.Fprintf(&out, "%q,", name)
		}
		out.WriteString("}}\n")
	}
	out.WriteString("// PaletteByName looks up a built-in palette by its lowercase name.\n// Unknown names return a zero palette and false.\nfunc PaletteByName(name string) (Palette, bool) { switch name {\n")
	for _, name := range names {
		fmt.Fprintf(&out, "case %q: return %sPalette(), true\n", name, strings.ToUpper(name[:1])+name[1:])
	}
	out.WriteString("default: return Palette{}, false\n}}\n")
	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		return err
	}
	return os.WriteFile("palette_gen.go", formatted, 0644)
}
