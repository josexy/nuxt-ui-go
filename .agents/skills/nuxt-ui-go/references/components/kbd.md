<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Kbd

Display a keyboard shortcut as key labels.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Kbd`, `FormatKey`
- Types: `KbdOptions`
- Theme key: `nui.ComponentKbd` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Pass key names in order, such as `[]string{"CmdOrCtrl", "K"}`. CmdOrCtrl displays Ctrl or ⌘ for the target platform; command and win always display ⌘ and ⊞. Kbd displays hints without registering shortcuts.

## Things to know

Kbd displays a hint; it does not register the shortcut. Bind keyboard actions separately in the application.

## API

```go
func Kbd(c *ui.Context, t Theme, keys []string, opts KbdOptions) ui.Element

// FormatKey formats familiar key tokens for a target GOOS without registering a shortcut.
func FormatKey(key, goos string) string
```

### KbdOptions

```go
type KbdOptions struct {
	Size     Size
	Color    Color
	Variant  Variant
	Platform string
	Decoration
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Kbd / FormatKey | Displays shortcut hints and platform modifier keys. Does not register shortcuts | Root, Label |

Kbd shows ordinary key names in uppercase and shows Tab, Backspace, Delete, CapsLock, the arrow keys, PageUp/PageDown, and Home/End as symbols. Modifier keys follow desktop conventions; for example, on platforms other than macOS, meta is shown as Win and shift as Shift.

`command` and `win` always show ⌘ and ⊞, regardless of Platform. `CmdOrCtrl` and `primary` show ⌘ on macOS and Ctrl elsewhere. Platform defaults to the current GOOS; set it to darwin, windows, or linux to show a specific platform. Kbd does not register shortcuts and is not in the default Tab order. The standard sizes are SM, MD, and LG; XS and XL are compatibility aliases for SM and LG.

The Neutral/Solid background of Button, Badge, Alert, and Kbd uses Theme.InvertedBackground, and the text uses Inverted. On hover, Button uses that background at 90% opacity; in high contrast mode it stays opaque. Instance UI and Style can still override the default look.

### Advanced interaction, charts, and layout

#### CommandPalette › Shortcuts and search (excerpt)

| Option or behavior | Description |
| --- | --- |
| Kbds / Prefix / Suffix / Avatar / Chip / Loading / Active | Command content and appearance. Kbds only display shortcuts, and Loading does not disable the command |

### Overlays, menus, and navigation

#### DropdownMenu and ContextMenu › Menu items (excerpt)

Shortcut is split on `+` into Kbd keys; in `Ctrl++`, the final `+` is the plus key. FormatKey converts key names to each platform's display form. XS and SM menus use SM Kbds, MD and LG use MD, and XL uses LG; the gaps between keys are 2/2/2/4/4 DIP. Go shows shortcuts even in narrow windows.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct{}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Kbd(c.Key("shortcut"), t, []string{"CmdOrCtrl", "K"}, nui.KbdOptions{}).AlignSelf(ui.Start)
}

func main() {
	s := &example{}

	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Kbd · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
