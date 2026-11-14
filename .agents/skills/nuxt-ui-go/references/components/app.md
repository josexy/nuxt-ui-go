<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# App

Apply a window theme and share it with its components.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `App`, `DefaultApp`, `Window`, `Themed`, `SupportedLocales`
- Types: `AppConfig`
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Call `nui.App(c, config)` at the start of the view and pass its theme to buttons, cards and other components. Initialize `config` when creating the window to keep appearance and palette choices. The complete example includes window and language setup.

## Things to know

App configures the current window theme. Create the window with `mygo.NewWindow`. Set the application language with `mygo.App.SetLocale`; AppConfig does not provide a per-window locale.

## API

```go
// Themed builds a column with a local theme. It intentionally has layout and a
// background, just like ui.Themed. The callback executes in the current build.
func Themed(c *ui.Context, theme Theme, children func(Theme)) ui.Element

// SupportedLocales returns the bundled component languages, English first.
// An app without its own catalogs can declare these with
// mygo.App.SetLanguages(SupportedLocales()...) and switch all its windows with
// mygo.App.SetLocale("zh-Hans"), SetLocale("en"), or SetLocale("") for the system.
// Applications with catalogs declare their own languages instead.
func SupportedLocales() []string
```

### AppConfig

```go
// AppConfig owns no window or transient UI values. Copy it for independent windows.
type AppConfig struct {
	Mode        Mode
	Light, Dark Theme
}

func DefaultApp() AppConfig

func (a AppConfig) Resolve(systemDark bool, prefs ui.Preferences) Theme
```

Types documented in other references: `Mode` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### App and theme

##### Create the config once, use it in the view

Initialize the config when you create the window and keep it in your app model. This example sets the primary color to blue and the neutral color to zinc:

```go
config := nui.DefaultApp()
config.Mode = nui.SystemMode // or LightMode / DarkMode

palettes := nui.DefaultPalettes()
palettes.Primary = nui.BluePalette()
palettes.Neutral = nui.ZincPalette()
config.Light = nui.NewTheme(false, palettes)
config.Dark = nui.NewTheme(true, palettes)
config.Light.Radius, config.Dark.Radius = 8, 8
config.Light.Density, config.Dark.Density = .9, .9
```

Call `App` at the start of the view function on every build, but do not re-run the setup above there; doing so would overwrite changes the user made at runtime.

```go
theme := nui.App(c, config)
nui.Button(c.Key("save"), theme, "Save", nui.ButtonOptions{
    Color: nui.Primary,
    Variant: nui.Solid,
    Size: nui.MD,
    Leading: lucide.Check(),
}).OnClick(save)
```

`App` sets the current window's native theme from the config and the system preferences. Call it before any `Window` or `Themed` call in the same build so it can read the system appearance. You still create MyGo windows with `mygo.NewWindow`, and notifications need a separate `ToastViewport`.

`App` returns a theme with the system preferences already applied. **Do not write it back to `config.Light` or `config.Dark`**, or text scaling may be applied twice. If you want full control over the theme, use the lower-level `Window(c, theme)`, which does not apply system preferences.

##### Common theme fields

`NewTheme` derives colors for each purpose, such as primary, text, and background colors, from a complete `Palettes` set. After changing a palette, regenerate both the light and the dark theme. Changing `Theme.Primary` directly changes only that field; it does not update the palette or the related colors.

| Field | Default and purpose |
| --- | --- |
| `Mode` | `SystemMode` follows the system; you can also choose `LightMode` or `DarkMode`. A zero-value `AppConfig` also follows the system |
| `FontSize` / `TextScale` | Base font size 14. Component font sizes scale from this base, and `App` then applies the system text scale |
| `Density` | Defaults to 1. Adjusts spacing and padding without shrinking text. Values of 0 or less are treated as 1 |
| `Radius` | Base corner radius of 6 DIP. Set it to 0 for square corners. Each component derives its radius from this base |
| `ContainerWidth` | Maximum content width, 1280 DIP by default |
| `Shadow` | `ShadowTokens{Color, Blur, OffsetY}`. Apply it to a part with `Shadow: nui.Set(theme.Shadow)`. Card has no outer shadow by default |
| `Motion` | Background transitions take 150 ms by default, a loading spinner rotation takes 1 s, and a Skeleton pulse takes 2 s. `Reduce` turns animation off |

Windows uses Segoe UI and Consolas by default; other platforms use system-ui and monospace. The library does not bundle fonts, and fonts and fallback fonts can affect the actual control height.

##### Color fields

| Purpose | Fields or default palettes |
| --- | --- |
| Semantic colors | Primary and Success default to green; Secondary and Info to blue; Warning is yellow; Error is red; Neutral is slate |
| Accent shades | Usually shade 500 in light mode and 400 in dark mode. Neutral has its own shade rules |
| Text | `Dimmed`, `Muted`, `Toned`, `Text`, `Highlighted`, `Inverted` |
| Backgrounds | `Background`, `MutedBackground`, `Elevated`, `Accented`, `InvertedBackground` |
| Borders | `Border`, `BorderMuted`, `BorderAccented`, `BorderInverted` |

If either the system or the theme enables reduced motion, animation is off; a component's `Animate: nui.Set(true)` cannot override it. High contrast mode strengthens text, borders, and focus indicators, and switches accent colors to shade 700 in light mode and 300 in dark mode.

With high contrast mode on, the default palettes' Solid text colors pass a 4.5:1 contrast check. This does not hold in normal mode: for example, white text on the default light primary color has a contrast ratio of about 2.22:1. Verify custom palettes, other states, and the accessibility of your complete app separately.

##### Palettes and runtime switching

The 26 built-in palettes come from pinned Tailwind CSS v4.3.3 data and keep their OKLCH values. Each palette has 11 shades: 50, 100, 200, …, 900, and 950. `Palette.Shade` accepts only these values and panics on any other.

Call functions such as `OrangePalette()` and `MauvePalette()` directly, or look a palette up with `PaletteByName("violet")`. Names are lowercase. The function returns `(Palette, bool)`; an unknown name returns the zero value and `false`.

| Function | Names returned |
| --- | --- |
| `PrimaryPaletteNames()` | red, orange, amber, yellow, lime, green, emerald, teal, cyan, sky, blue, indigo, violet, purple, fuchsia, pink, rose |
| `NeutralPaletteNames()` | slate, gray, zinc, neutral, stone, taupe, mauve, mist, olive |

Each function returns a new slice in menu order. The groups only help users pick colors; they do not restrict which color field a palette can be assigned to. `NeutralPalette()` returns the gray palette named "neutral", whereas `Palettes.Neutral` is the theme field used for neutral colors.

To switch colors at runtime:

1. Store the user's chosen palette in your app state.
2. Regenerate `config.Light` and `config.Dark` from the new palette.
3. Keep calling `App(c, config)` in later builds.

Primary and Neutral can be changed independently; the other semantic colors use their own palettes. The menus support arrow keys, Home, End, type-to-select, and Escape, and return focus to the trigger button after a selection.

##### Black and white primary color

`WithBlackPrimary()` makes the primary color black in light mode and white in dark mode, including in high contrast mode. Create both themes first, then call it on each:

```go
config := nui.DefaultApp()
config.Light = config.Light.WithBlackPrimary()
config.Dark = config.Dark.WithBlackPrimary()
```

The method returns a new theme with Primary and its shades updated; all other colors and styles are unchanged.

Black changes with the color mode and has no 11-shade palette, so it does not appear in `PrimaryPaletteNames()`, and `PaletteByName("black")` returns no palette. To switch back to a color, regenerate the light and dark themes from the chosen palette.

#### Components and parts › Page structure and actions (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| App / Theme | `AppConfig` manages light/dark mode; `Theme` holds colors, sizes, and component defaults | Configured per component |

### Internationalization

#### Setting up your app › Switch languages (excerpt)

For automated tests, set the environment variable `MYGO_LOCALE=en` to pin the language without saving a preference.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct{}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Card(c.Key("card"), t, nui.CardOptions{Title: "NuxtUI Go", Description: "Window theme and component defaults"}, func() {
		nui.Button(c.Key("save"), t, "Save changes", nui.ButtonOptions{AccessibleLabel: "Save changes"})
	})
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
		mygo.NewWindow(mygo.WindowOptions{Title: "App · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
