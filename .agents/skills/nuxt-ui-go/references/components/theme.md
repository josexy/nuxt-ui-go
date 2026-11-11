<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Theme (`nui.Light`)

Set shared colors, sizes and component styles.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `Light`, `Dark`, `NewTheme`, `Set`, `DefaultPalettes`, `PaletteByName`, `PrimaryPaletteNames`, `NeutralPaletteNames`
- Types: `Theme`, `ComponentTheme`, `PartStyle`, `Decoration`, `Optional`, `Color`, `Size`, `Variant`, `ButtonVariant`, `Orientation`, `Slot`, `Styles`, `Component`, `ShadowTokens`, `Palette`, `Palettes`, `Mode`, `MotionTokens`
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

`Theme` holds colors, typography, spacing and component defaults. Use `WithComponent` to set defaults for a component type, Options to override them on one instance, and `UI` to customize individual parts.

## Things to know

Apply `WithBlackPrimary()` to both themes for black in light mode and white in dark mode, including high contrast. ComponentTheme supplies Color, Size, Variant and UI defaults; set other parameters through component Options. The 4.5:1 contrast check for default Solid text covers high-contrast mode. The normal light primary with white text is approximately 2.22:1.

## API

```go
// PrimaryPaletteNames returns the built-in primary choices in display order.
// The returned slice belongs to the caller. Black is a mode-dependent override, not a palette.
func PrimaryPaletteNames() []string

// NeutralPaletteNames returns the built-in neutral choices in display order.
// The returned slice belongs to the caller. Black is a mode-dependent override, not a palette.
func NeutralPaletteNames() []string
```

### Theme

```go
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
	// contains filtered or unexported fields
}

// App must be called at the start of a view, before Window or Themed replace the system theme.
func App(c *ui.Context, config AppConfig) Theme

func Dark() Theme

func Light() Theme

func NewTheme(dark bool, p Palettes) Theme

// Window applies a value snapshot to this window for the current build.
// Pass the returned theme and the live Context explicitly to constructors.
func Window(c *ui.Context, theme Theme) Theme

func (t Theme) Component(component Component) ComponentTheme

// FollowPreferences applies desktop accessibility preferences once to an unscaled snapshot.
func (t Theme) FollowPreferences(p ui.Preferences) Theme

func (t Theme) Native() *ui.Theme

// WithBlackPrimary returns a copy with a black primary in light mode and a white
// primary in dark mode. It also updates the primary palette so high contrast and
// components that use individual shades retain the monochrome primary.
// Apply it to both AppConfig snapshots after constructing their themes.
func (t Theme) WithBlackPrimary() Theme

// WithComponent returns an isolated theme snapshot. Caller maps are copied.
func (t Theme) WithComponent(component Component, patch ComponentTheme) Theme
```

### ComponentTheme

```go
type ComponentTheme struct {
	Color   Color
	Size    Size
	Variant Variant
	UI      Styles
}
```

### PartStyle

```go
// PartStyle contains appearance and spacing only. It cannot remove roles or input behavior.
type PartStyle struct {
	Background, Foreground, BorderColor           Optional[ui.Color]
	Radius, BorderWidth, Opacity                  Optional[float32]
	FontSize, LineHeight, PaddingX, PaddingY, Gap Optional[float32]
	Font                                          Optional[string]
	Weight                                        Optional[int]
	Shadow                                        Optional[ShadowTokens]
}
```

### Decoration

```go
// Decoration is shared by components with several synchronous styling slots.
type Decoration struct {
	UI    Styles
	Style func(Slot, ui.Element)
}
```

### Optional

```go
// Optional distinguishes inheritance from an explicit zero, false or transparent value.
type Optional[T any] struct {
	Value T
	Set   bool
}

func ParseClockTime(text string, twelve, seconds bool) (Optional[ClockTime], bool)

func ParseDate(text string) (Optional[Date], bool)

// ParseNumber accepts empty input and finite numbers; AllowInvalid bypasses bounds.
func ParseNumber(text string, o InputNumberOptions) (Optional[float64], bool)

func Set[T any](v T) Optional[T]
```

### Color

```go
type Color uint8

const (
	ColorDefault Color = iota
	Primary
	Secondary
	Success
	Info
	Warning
	Error
	Neutral
)
```

### Size

```go
type Size uint8

const (
	SizeDefault Size = iota
	XS
	SM
	MD
	LG
	XL
)
```

### Variant

```go
type Variant uint8

const (
	VariantDefault Variant = iota
	Solid
	Outline
	Soft
	Subtle
	Ghost
	LinkVariant
	None
)
```

### ButtonVariant

```go
type ButtonVariant = Variant
```

### Orientation

```go
// Orientation describes the arrangement of a component's controls.
type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)
```

### Slot

```go
type Slot string

const (
	Root                Slot = "root"
	IncrementSlot       Slot = "increment"
	DecrementSlot       Slot = "decrement"
	Label               Slot = "label"
	Leading             Slot = "leading"
	Trailing            Slot = "trailing"
	Header              Slot = "header"
	Body                Slot = "body"
	Footer              Slot = "footer"
	Description         Slot = "description"
	Hint                Slot = "hint"
	Help                Slot = "help"
	ErrorSlot           Slot = "error"
	Indicator           Slot = "indicator"
	IconSlot            Slot = "icon"
	EmptyIconSlot       Slot = "emptyIcon"
	Editor              Slot = "editor"
	Segment             Slot = "segment"
	Panel               Slot = "panel"
	ContainerSlot       Slot = "container"
	CenterSlot          Slot = "center"
	XAxisSlot           Slot = "xAxis"
	YAxisSlot           Slot = "yAxis"
	Item                Slot = "item"
	EmptySlot           Slot = "empty"
	Thumb               Slot = "thumb"
	Backdrop            Slot = "backdrop"
	Trigger             Slot = "trigger"
	List                Slot = "list"
	Close               Slot = "close"
	BackSlot            Slot = "back"
	Actions             Slot = "actions"
	SeparatorSlot       Slot = "separator"
	First               Slot = "first"
	Previous            Slot = "previous"
	Next                Slot = "next"
	Last                Slot = "last"
	Ellipsis            Slot = "ellipsis"
	DateSlot            Slot = "date"
	StatusSlot          Slot = "status"
	LoadingSlot         Slot = "loading"
	LoadingIconSlot     Slot = "loadingIcon"
	ItemLabelSlot       Slot = "itemLabel"
	ItemLabelPrefixSlot Slot = "itemLabelPrefix"
	ItemLabelSuffixSlot Slot = "itemLabelSuffix"
	ItemLeadingSlot     Slot = "itemLeading"
	ItemTrailingSlot    Slot = "itemTrailing"
	GridSlot            Slot = "grid"
	GridRowSlot         Slot = "gridRow"
	HeadingSlot         Slot = "heading"
	CaptionSlot         Slot = "caption"
	WeekDaySlot         Slot = "weekDay"
	WeekNumberSlot      Slot = "weekNumber"
)
```

### Styles

```go
type Styles map[Slot]PartStyle
```

### Component

```go
type Component string

const (
	ComponentButton         Component = "button"
	ComponentLink           Component = "link"
	ComponentBadge          Component = "badge"
	ComponentCard           Component = "card"
	ComponentAlert          Component = "alert"
	ComponentAvatar         Component = "avatar"
	ComponentAvatarGroup    Component = "avatar-group"
	ComponentChip           Component = "chip"
	ComponentSeparator      Component = "separator"
	ComponentKbd            Component = "kbd"
	ComponentSkeleton       Component = "skeleton"
	ComponentProgress       Component = "progress"
	ComponentProgressGroup  Component = "progress-group"
	ComponentEmpty          Component = "empty"
	ComponentUser           Component = "user"
	ComponentContainer      Component = "container"
	ComponentMain           Component = "main"
	ComponentIcon           Component = "icon"
	ComponentInput          Component = "input"
	ComponentTextarea       Component = "textarea"
	ComponentSelect         Component = "select"
	ComponentInputMenu      Component = "input-menu"
	ComponentPopover        Component = "popover"
	ComponentFormField      Component = "form-field"
	ComponentFieldGroup     Component = "field-group"
	ComponentCheckbox       Component = "checkbox"
	ComponentCheckboxGroup  Component = "checkbox-group"
	ComponentRadioGroup     Component = "radio-group"
	ComponentSwitch         Component = "switch"
	ComponentSlider         Component = "slider"
	ComponentScrollArea     Component = "scroll-area"
	ComponentTooltip        Component = "tooltip"
	ComponentModal          Component = "modal"
	ComponentToast          Component = "toast"
	ComponentDropdownMenu   Component = "dropdown-menu"
	ComponentContextMenu    Component = "context-menu"
	ComponentSlideover      Component = "slideover"
	ComponentTabs           Component = "tabs"
	ComponentCollapsible    Component = "collapsible"
	ComponentAccordion      Component = "accordion"
	ComponentBreadcrumb     Component = "breadcrumb"
	ComponentPagination     Component = "pagination"
	ComponentStepper        Component = "stepper"
	ComponentTimeline       Component = "timeline"
	ComponentBanner         Component = "banner"
	ComponentError          Component = "error"
	ComponentForm           Component = "form"
	ComponentInputNumber    Component = "input-number"
	ComponentInputTags      Component = "input-tags"
	ComponentPinInput       Component = "pin-input"
	ComponentInputRating    Component = "input-rating"
	ComponentListbox        Component = "listbox"
	ComponentSelectMenu     Component = "select-menu"
	ComponentCalendar       Component = "calendar"
	ComponentInputDate      Component = "input-date"
	ComponentInputTime      Component = "input-time"
	ComponentColorPicker    Component = "color-picker"
	ComponentTable          Component = "table"
	ComponentTree           Component = "tree"
	ComponentCommandPalette Component = "command-palette"
	ComponentNavigationMenu Component = "navigation-menu"
	ComponentDrawer         Component = "drawer"
	ComponentFileUpload     Component = "file-upload"
	ComponentChart          Component = "chart"
	ComponentCarousel       Component = "carousel"
	ComponentMarquee        Component = "marquee"
	ComponentSplitter       Component = "splitter"
	ComponentSidebar        Component = "sidebar"
	ComponentHeader         Component = "header"
	ComponentFooter         Component = "footer"
	ComponentFooterColumns  Component = "footer-columns"
)
```

### ShadowTokens

```go
type ShadowTokens struct {
	Color         ui.Color
	Blur, OffsetY float32
}
```

### Palette

```go
// Palette stores shades 50,100,200,...,900,950 in that order.
type Palette [11]ui.Color

func AmberPalette() Palette

func BluePalette() Palette

func CyanPalette() Palette

func EmeraldPalette() Palette

func FuchsiaPalette() Palette

func GrayPalette() Palette

func GreenPalette() Palette

func IndigoPalette() Palette

func LimePalette() Palette

func MauvePalette() Palette

func MistPalette() Palette

func NeutralPalette() Palette

func OlivePalette() Palette

func OrangePalette() Palette

// PaletteByName looks up a built-in palette by its lowercase name.
// Unknown names return a zero palette and false.
func PaletteByName(name string) (Palette, bool)

func PinkPalette() Palette

func PurplePalette() Palette

func RedPalette() Palette

func RosePalette() Palette

func SkyPalette() Palette

func SlatePalette() Palette

func StonePalette() Palette

func TaupePalette() Palette

func TealPalette() Palette

func VioletPalette() Palette

func YellowPalette() Palette

func ZincPalette() Palette

func (p Palette) Shade(shade int) ui.Color
```

### Palettes

```go
type Palettes struct{ Primary, Secondary, Success, Info, Warning, Error, Neutral Palette }

func DefaultPalettes() Palettes
```

### Mode

```go
type Mode uint8

const (
	SystemMode Mode = iota
	LightMode
	DarkMode
)
```

### MotionTokens

```go
type MotionTokens struct {
	Reduce                  bool
	Transition, Spin, Pulse time.Duration
}
```

### Related types

```go
// ClockTime is a wall-clock value, not an instant. Conversion to a timezone is application-owned.
type ClockTime struct{ Hour, Minute, Second int }

func (v ClockTime) String() string

func (v ClockTime) Valid() bool
```

```go
// Date is a civil Gregorian date; it has no timezone or time of day.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func DateFromTime(v time.Time) Date

func (d Date) AddDays(n int) Date

func (d Date) AddMonths(n int) Date

func (d Date) Compare(other Date) int

func (d Date) String() string

func (d Date) Valid() bool
```

Types documented in other references: `AppConfig` → [app.md](app.md), `InputNumberOptions` → [input-number.md](input-number.md).

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

#### Style overrides

##### Choose the scope

To change one component, set its color, size, or `UI` in that call's options. To give a group of components a shared style, create a local theme with `WithComponent` and apply it with `Themed`:

```go
local := theme.WithComponent(nui.ComponentButton, nui.ComponentTheme{
    Color: nui.Secondary,
    Variant: nui.Soft,
    UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999))}},
})
nui.Themed(c, local, func(inner nui.Theme) {
    nui.Button(c, inner, "Local default", nui.ButtonOptions{})
    nui.Button(c, inner, "Square corners", nui.ButtonOptions{
        Color: nui.Error,
        Variant: nui.Outline,
        UI: nui.Styles{nui.Root: {
            Radius: nui.Set(float32(0)),
            PaddingX: nui.Set(float32(0)),
        }},
    })
})
```

`WithComponent` copies the configuration map and `Component` returns a copy of the configuration, so changing a local theme does not affect sibling regions or other windows. `Themed` creates a Column with a background, whose layout you can configure further; after the children are built, the outer native theme is restored.

In Nuxt UI, `Theme.props` can set a default for any public prop of a component. Go's `ComponentTheme` contains only Color, Size, Variant, and UI. Set behavioral options such as loading, disabled, and delays through each component's Options or your own builder functions. Unlike Vue's transparent providers, Go passes the theme explicitly, and `Themed` also creates a layout container.

##### Override order

Each later setting overrides the same field from earlier ones:

```text
Built-in component style
  → Component defaults and UI from the window or local theme
  → Color / Variant / Size set explicitly on the instance
  → The instance's UI
  → The synchronous Style callback
  → Later changes on the returned element
```

`Color` and `Variant` adjust the related backgrounds, foregrounds, and borders. `Size` adjusts font size, line height, font weight, corner radius, padding, and gaps. They override only those fields and leave other part settings intact. Set the overall width and height through the options or on the returned element.

`PartStyle` supports foreground, background, border, corner radius, opacity, font, font size, font weight, line height, padding, gap, and shadow. `Optional[T]{}` means unset, while `Set(0)`, `Set(false)`, and `Set(ui.Transparent)` are all valid overrides. In enums, Default usually means inherit.

##### Styling a single part

Most components accept part styles through `Decoration{UI, Style}`. `UI` sets properties by part name; use `Style` when you need to configure the MyGo element directly:

```go
nui.Card(c, theme, nui.CardOptions{
    Title: "Project",
    Decoration: nui.Decoration{
        UI: nui.Styles{nui.Body: {PaddingY: nui.Set(float32(12))}},
        Style: func(slot nui.Slot, part ui.Element) {
            if slot == nui.Footer {
                part.Wrap()
            }
        },
    },
}, func() {
    ui.Text(c, "Project description")
})
```

Button has a `UI` field and `Style(root, label, leading, trailing ui.Element)` directly. Parts that were not created are passed as zero-value Elements. The callback configures only the elements of the current build; do not keep Elements, parts, or the Context for later. If you customize interaction, keep the control's name, role, and focus behavior.

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
	local := t.WithComponent(nui.ComponentButton, nui.ComponentTheme{})
	nui.Button(c.Key("local"), local, "Local theme", nui.ButtonOptions{})
	nui.Button(c.Key("override"), local, "Instance override", nui.ButtonOptions{Color: nui.Warning})
	nui.Button(c.Key("parts"), t, "Part style override", nui.ButtonOptions{UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999)), PaddingX: nui.Set(float32(24))}}}).AlignSelf(ui.Start)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Theme · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
