package nui

//go:generate go run ./internal/cmd/genpalettes

import (
	"runtime"
	"time"

	"github.com/egoist/mygo/ui"
)

// Palette stores shades 50,100,200,...,900,950 in that order.
type Palette [11]ui.Color

func (p Palette) Shade(shade int) ui.Color {
	if shade == 50 {
		return p[0]
	}
	if shade == 950 {
		return p[10]
	}
	if shade >= 100 && shade <= 900 && shade%100 == 0 {
		return p[shade/100]
	}
	panic("nui: palette shade must be 50, 100..900 or 950")
}

type Palettes struct{ Primary, Secondary, Success, Info, Warning, Error, Neutral Palette }

func DefaultPalettes() Palettes {
	return Palettes{GreenPalette(), BluePalette(), GreenPalette(), BluePalette(), YellowPalette(), RedPalette(), SlatePalette()}
}

type Mode uint8

const (
	SystemMode Mode = iota
	LightMode
	DarkMode
)

type MotionTokens struct {
	Reduce                  bool
	Transition, Spin, Pulse time.Duration
}
type ShadowTokens struct {
	Color         ui.Color
	Blur, OffsetY float32
}

// Theme is a value snapshot. The private component table is immutable; WithComponent copies it.
type Theme struct {
	Dark                                                                bool
	Background, MutedBackground, Elevated, Accented, InvertedBackground ui.Color
	Dimmed, Muted, Toned, Text, Highlighted, Inverted                   ui.Color
	Border, BorderMuted, BorderAccented, BorderInverted                 ui.Color
	Primary, Secondary, Success, Info, Warning, Error                   ui.Color
	Radius, FontSize, Density, TextScale, ContainerWidth                float32
	Font, MonoFont                                                      string
	Motion                                                              MotionTokens
	Shadow                                                              ShadowTokens
	HighContrast                                                        bool
	Palettes                                                            Palettes
	components                                                          map[Component]ComponentTheme
	field                                                               fieldDefaults
	form                                                                *FormState
	group                                                               groupPart
}

func Light() Theme { return NewTheme(false, DefaultPalettes()) }
func Dark() Theme  { return NewTheme(true, DefaultPalettes()) }

// WithBlackPrimary returns a copy with a black primary in light mode and a white
// primary in dark mode. It also updates the primary palette so high contrast and
// components that use individual shades retain the monochrome primary.
// Apply it to both AppConfig snapshots after constructing their themes.
func (t Theme) WithBlackPrimary() Theme {
	t.Primary = ui.Hex("#000000")
	if t.Dark {
		t.Primary = ui.Hex("#ffffff")
	}
	for i := range t.Palettes.Primary {
		t.Palettes.Primary[i] = t.Primary
	}
	return t
}

func NewTheme(dark bool, p Palettes) Theme {
	t := Theme{Dark: dark, Palettes: p, Radius: 6, FontSize: 14, Density: 1, TextScale: 1, ContainerWidth: 1280, Font: "system-ui", MonoFont: "monospace", Motion: MotionTokens{Transition: 150 * time.Millisecond, Spin: time.Second, Pulse: 2 * time.Second}, Shadow: ShadowTokens{Color: ui.RGBA(0, 0, 0, .12), Blur: 12, OffsetY: 4}}
	if runtime.GOOS == "windows" {
		t.Font = "Segoe UI"
		t.MonoFont = "Consolas"
	}
	n := p.Neutral
	white := ui.Hex("#ffffff")
	t.Background, t.MutedBackground, t.Elevated, t.Accented, t.InvertedBackground = white, n.Shade(50), n.Shade(100), n.Shade(200), n.Shade(900)
	t.Dimmed, t.Muted, t.Toned, t.Text, t.Highlighted, t.Inverted = n.Shade(400), n.Shade(500), n.Shade(600), n.Shade(700), n.Shade(900), white
	t.Border, t.BorderMuted, t.BorderAccented, t.BorderInverted = n.Shade(200), n.Shade(200), n.Shade(300), n.Shade(900)
	shade := 500
	if dark {
		shade = 400
		t.Background, t.MutedBackground, t.Elevated, t.Accented, t.InvertedBackground = n.Shade(900), n.Shade(800), n.Shade(800), n.Shade(700), white
		t.Dimmed, t.Muted, t.Toned, t.Text, t.Highlighted, t.Inverted = n.Shade(500), n.Shade(400), n.Shade(300), n.Shade(200), white, n.Shade(900)
		t.Border, t.BorderMuted, t.BorderAccented, t.BorderInverted = n.Shade(800), n.Shade(700), n.Shade(700), white
	}
	t.Primary, t.Secondary, t.Success, t.Info, t.Warning, t.Error = p.Primary.Shade(shade), p.Secondary.Shade(shade), p.Success.Shade(shade), p.Info.Shade(shade), p.Warning.Shade(shade), p.Error.Shade(shade)
	return t
}
func (t Theme) spacing(v float32) float32 {
	if t.Density <= 0 {
		return v
	}
	return v * t.Density
}
func (t Theme) text(v float32) float32 {
	scale := t.TextScale
	if scale <= 0 {
		scale = 1
	}
	base := t.FontSize
	if base <= 0 {
		base = 14
	}
	return v * (base / 14) * scale
}
func (t Theme) radius(v float32) float32  { return v * t.Radius / 6 }
func (t Theme) motion(c *ui.Context) bool { return !t.Motion.Reduce && !c.Preferences().ReduceMotion }
func (t Theme) focusColor(color Color) ui.Color {
	c := t.color(color)
	if !t.HighContrast {
		c = c.Alpha(.25)
	}
	return c
}

// FollowPreferences applies desktop accessibility preferences once to an unscaled snapshot.
func (t Theme) FollowPreferences(p ui.Preferences) Theme {
	t.Motion.Reduce = t.Motion.Reduce || p.ReduceMotion
	t.HighContrast = t.HighContrast || p.HighContrast
	if p.TextScale > 0 {
		if t.TextScale <= 0 {
			t.TextScale = 1
		}
		t.TextScale *= p.TextScale
	}
	if t.HighContrast {
		// Stronger accents preserve readable inverted text in the default palettes.
		shade := 700
		if t.Dark {
			shade = 300
		}
		p := t.Palettes
		t.Primary, t.Secondary, t.Success, t.Info, t.Warning, t.Error = p.Primary.Shade(shade), p.Secondary.Shade(shade), p.Success.Shade(shade), p.Info.Shade(shade), p.Warning.Shade(shade), p.Error.Shade(shade)
		t.Muted = t.Muted.Mix(t.Text, .5)
		t.Dimmed = t.Muted
		t.Border = t.Border.Mix(t.Text, .45)
		t.BorderAccented = t.BorderAccented.Mix(t.Text, .45)
	}
	return t
}
func (t Theme) Native() *ui.Theme {
	n := ui.LightTheme()
	if t.Dark {
		n = ui.DarkTheme()
	}
	n.Dark, n.Background = t.Dark, t.Background
	n.Surface, n.SurfaceHover, n.SurfacePressed = t.Background, t.Elevated, t.Accented
	n.Text, n.TextMuted, n.Border = t.Text, t.Muted, t.BorderAccented
	n.Accent, n.AccentHover, n.AccentPressed, n.AccentText = t.Primary, t.Primary.Alpha(.75), t.Primary.Alpha(.75), t.Inverted
	n.Danger, n.Warning, n.Success = t.Error, t.Warning, t.Success
	n.Focus, n.Selection = t.focusColor(Primary), t.Primary.Alpha(.20)
	n.Radius, n.Spacing, n.FontSize, n.Font = t.Radius, t.spacing(4), t.text(14), t.Font
	return n
}

// AppConfig owns no window or transient UI values. Copy it for independent windows.
type AppConfig struct {
	Mode        Mode
	Light, Dark Theme
}

func DefaultApp() AppConfig { return AppConfig{Light: Light(), Dark: Dark()} }
func (a AppConfig) Resolve(systemDark bool, prefs ui.Preferences) Theme {
	// A zero config follows the system with the built-in palettes.
	if a.Light.Palettes == (Palettes{}) {
		a.Light = Light()
	}
	if a.Dark.Palettes == (Palettes{}) {
		a.Dark = Dark()
	}
	dark := a.Mode == DarkMode || (a.Mode == SystemMode && systemDark)
	t := a.Light
	if dark {
		t = a.Dark
	}
	return t.FollowPreferences(prefs)
}

// App must be called at the start of a view, before Window or Themed replace the system theme.
func App(c *ui.Context, config AppConfig) Theme {
	return Window(c, config.Resolve(c.Theme().Dark, c.Preferences()))
}
