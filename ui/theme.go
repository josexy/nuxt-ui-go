package nui

//go:generate go run ../internal/cmd/genpalettes -root ..

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

// MotionTokens set how much nui's components animate, and for how long.
//
// Components animate at one of three levels, resolved for each build:
//   - Off, when Disabled is set: nothing animates, transitions, pulses or
//     spins, and no component asks for frames to animate.
//   - Reduced, when Reduce is set or the desktop asks for reduced motion:
//     nothing moves, scales, slides or rotates, while fades, colour
//     transitions, loading spinners and pulses go on.
//   - Full otherwise.
//
// A component's own Animate or Transition option can turn its animation down,
// never up past the level.
type MotionTokens struct {
	// Reduce asks for reduced motion, as the desktop's preference does.
	Reduce bool
	// Disabled turns nui's animation off. AppConfig.DisableAnimations sets it
	// in both themes.
	Disabled bool
	// Transition controls surface color transitions and avatar/rating scale
	// poses; nonpositive values are immediate. Spin is a spinner's turn and
	// Pulse a skeleton's or a reduced loading bar's pulse.
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
func (t Theme) radius(v float32) float32 { return v * t.Radius / 6 }

// motionLevel is how much a component animates in this build (MotionTokens).
type motionLevel uint8

const (
	motionFull motionLevel = iota
	motionReduced
	motionOff
)

func (t Theme) motionLevel(c *ui.Context) motionLevel {
	switch {
	case t.Motion.Disabled:
		return motionOff
	case t.Motion.Reduce || c.Preferences().ReduceMotion:
		return motionReduced
	}
	return motionFull
}

// moves reports whether components may move: scale, slide, rotate, reveal
// height, change width, travel and autoplay. Full only.
func (t Theme) moves(c *ui.Context) bool { return t.motionLevel(c) == motionFull }

// fades reports whether components may change without moving: opacity,
// colour transitions, spinners and pulses. Full and Reduced.
func (t Theme) fades(c *ui.Context) bool { return t.motionLevel(c) != motionOff }

// colorTransition returns the transition of an interactive surface's
// background, border and shadow colours, as Nuxt UI's transition-colors, and
// false when colour changes are immediate: at Off, or when Motion.Transition
// is not positive.
func (t Theme) colorTransition(c *ui.Context) (ui.ElementTransition, bool) {
	if t.Motion.Transition <= 0 || !t.fades(c) {
		return ui.ElementTransition{}, false
	}
	return ui.ElementTransition{Colors: true, Shadows: true, Duration: t.Motion.Transition}, true
}

// fadeColors gives the surface e the colour transition (colorTransition),
// when colour changes fade.
func (t Theme) fadeColors(c *ui.Context, e ui.Element) {
	if tr, ok := t.colorTransition(c); ok {
		e.Transition(tr)
	}
}

// focusRing gives e Nuxt's 3 DIP focus ring in color while shown, and a
// transparent one otherwise, added before the part's style so that Style
// still runs last. The ring is always there, so that it fades with the
// colour transition: MyGo changes a shadow list at once where an outer
// shadow takes the place of the inset one that draws the border.
func focusRing(e ui.Element, shown bool, color ui.Color) {
	if !shown {
		color = color.Alpha(0)
	}
	e.Shadow(0, 0, 0, 3, color)
}

// focusRingColor is the colour a field's focus ring takes: Error when invalid.
func focusRingColor(color Color, invalid bool) Color {
	if invalid {
		return Error
	}
	return color
}

// overlayTransition returns the enter and exit of an overlay element, over d
// and exitD. A nil motion plays none. At Reduced the motions keep their
// opacity and blur and lose everything that moves, since MyGo strips motion
// only for the desktop's preference, not for Motion.Reduce. At Off every
// change is immediate and the element leaves without a copy.
func (t Theme) overlayTransition(c *ui.Context, enter, exit *ui.Motion, d, exitD time.Duration) ui.ElementTransition {
	switch t.motionLevel(c) {
	case motionOff:
		return ui.ElementTransition{Duration: -1, ExitDuration: -1}
	case motionReduced:
		enter, exit = fadeOnly(enter), fadeOnly(exit)
	}
	return ui.ElementTransition{Duration: d, ExitDuration: exitD, Enter: enter, Exit: exit}
}

// fadeOnly returns m without what moves.
func fadeOnly(m *ui.Motion) *ui.Motion {
	if m == nil {
		return nil
	}
	return &ui.Motion{Opacity: m.Opacity, Blur: m.Blur}
}
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
	// DisableAnimations turns nui's animation off in both themes, whatever
	// the mode (MotionTokens.Disabled). It does not stop MyGo's own motion,
	// such as scroll glides, which follow the desktop's preference; call
	// c.DisableAnimations for the app's own.
	DisableAnimations bool
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
	a.Light.Motion.Disabled = a.Light.Motion.Disabled || a.DisableAnimations
	a.Dark.Motion.Disabled = a.Dark.Motion.Disabled || a.DisableAnimations
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

// Immediate transforms discard an in-flight native pose when motion is
// disabled or its duration is nonpositive. No view timer is needed in any mode.
func (t Theme) scaleFeedback(c *ui.Context, e ui.Element, scale float32) {
	if t.moves(c) && t.Motion.Transition > 0 {
		e.AnimatePose(ui.Pose{Scale: scale}, ui.Animation{Duration: t.Motion.Transition, Ease: ui.EaseOut})
	} else {
		e.Scale(scale, scale)
	}
}
