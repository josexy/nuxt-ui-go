# Themes and basic components

Call `App` to set up the window's theme, then pass the returned `Theme` to each component. This page covers theme configuration and basic components such as buttons, cards, avatars, and progress bars. For complete startup code, see the [minimal app](../examples/minimal/main.go).

| What you want to do | Where to look |
| --- | --- |
| Set light/dark mode, the primary color, or the font size | [App and theme](#app-and-theme) |
| Change the look of a component type or a single part | [Style overrides](#style-overrides) |
| Find a basic component | [Components and parts](#components-and-parts) |
| Add buttons, links, or avatars | [Buttons and links](#buttons-and-links), [Images, avatars, and icons](#images-avatars-and-icons) |
| Show progress or loading states | [Progress and animation](#progress-and-animation) |

## App and theme

### Create the config once, use it in the view

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

### Common theme fields

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

### Color fields

| Purpose | Fields or default palettes |
| --- | --- |
| Semantic colors | Primary and Success default to green; Secondary and Info to blue; Warning is yellow; Error is red; Neutral is slate |
| Accent shades | Usually shade 500 in light mode and 400 in dark mode. Neutral has its own shade rules |
| Text | `Dimmed`, `Muted`, `Toned`, `Text`, `Highlighted`, `Inverted` |
| Backgrounds | `Background`, `MutedBackground`, `Elevated`, `Accented`, `InvertedBackground` |
| Borders | `Border`, `BorderMuted`, `BorderAccented`, `BorderInverted` |

If either the system or the theme enables reduced motion, animation is off; a component's `Animate: nui.Set(true)` cannot override it. High contrast mode strengthens text, borders, and focus indicators, and switches accent colors to shade 700 in light mode and 300 in dark mode.

With high contrast mode on, the default palettes' Solid text colors pass a 4.5:1 contrast check. This does not hold in normal mode: for example, white text on the default light primary color has a contrast ratio of about 2.22:1. Verify custom palettes, other states, and the accessibility of your complete app separately.

### Palettes and runtime switching

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

Primary and Neutral can be changed independently; the other semantic colors use their own palettes. The Gallery's **Theme → Primary / Neutral** menus show a complete example, and **Light / Dark / System** switches the mode. The menus support arrow keys, Home, End, type-to-select, and Escape, and return focus to the trigger button after a selection. Choices are kept per window and reset when the Gallery restarts.

### Black and white primary color

`WithBlackPrimary()` makes the primary color black in light mode and white in dark mode, including in high contrast mode. Create both themes first, then call it on each:

```go
config := nui.DefaultApp()
config.Light = config.Light.WithBlackPrimary()
config.Dark = config.Dark.WithBlackPrimary()
```

The method returns a new theme with Primary and its shades updated; all other colors and styles are unchanged. The Primary menus in the Gallery and the website preview both include a Black option.

Black changes with the color mode and has no 11-shade palette, so it does not appear in `PrimaryPaletteNames()`, and `PaletteByName("black")` returns no palette. To switch back to a color, regenerate the light and dark themes from the chosen palette.

## Style overrides

### Choose the scope

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

### Override order

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

### Styling a single part

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

## Components and parts

Choose a component by purpose. The part names below can be used in `UI`; parts that are not listed have no styling effect.

### Page structure and actions

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| App / Theme | `AppConfig` manages light/dark mode; `Theme` holds colors, sizes, and component defaults | Configured per component |
| Container | Centers content and limits its width. `MaxWidth` and `Padding` are `Optional[float32]` | Root |
| Main | A flexible main content area with a minimum width and height of 0. Wrap it in `ui.Scroll` if it needs to scroll | Global Root; the returned element can be configured further |
| Icon | Displays a `*ui.SVG` with a size, color, and Label. Without a Label, it is decorative | Root |
| Button | A button with colors, variants, sizes, icons, an avatar, and loading and disabled states | Root, Label, Leading, Trailing |
| Link | Opens a URL or runs app navigation. Supports Active, Color, and Disabled | Root, Label |

Container's default horizontal padding depends on the window width: 16 DIP below 640 DIP, 24 DIP from 640 to 1023 DIP, and 32 DIP at 1024 DIP or more, multiplied by the theme's Density. The breakpoints use the window width, not the width of a narrower parent. An explicit `Padding` is a fixed DIP value, and `nui.Set(0)` removes the padding; instance UI and Style can still override it.

Button's `Avatar: &nui.AvatarOptions{Alt: "Ada Lovelace"}` shows a leading avatar. A leading icon or a leading loading spinner takes precedence over it; a trailing loading spinner does not hide the avatar. If Avatar.Size is unset, button sizes XS–XL use avatars of 16, 16, 20, 20, and 24 DIP; an explicit Avatar.Size takes precedence. An avatar-only button still needs an AccessibleLabel. In this mode, the Leading part is the avatar's root element.

The loading spinner rotates in its own draw callback. To set its color, use `UI[Leading].Foreground` or `UI[Trailing].Foreground`; a TextColor set in the Style callback does not change a color the draw callback has already resolved. Other layout and size properties can still be set through Style.

### Displaying information

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Badge | A short label. Supports Solid/Outline/Soft/Subtle, colors, sizes, Square, icons on either side, and a leading avatar | Root, Label, Leading, Trailing |
| Card | A card. Defaults to Outline and also supports Solid/Soft/Subtle. Has a title, description, Header, body, and Footer | Root, Header, Body, Footer, Label, Description |
| Alert | A message. Defaults to Solid; supports four variants, an icon, an avatar, Actions, and close options | Root, Body, Leading, Label, Description, Actions, Trailing |
| Empty | An empty state with an avatar or icon, title, description, content builders, and Loading/LoadingIcon. `None` gives the "naked" look | Root, Header, Leading, Label, Description, Body, Actions, Footer |
| User | An avatar, badge chip, name, and description. Supports Orientation/Vertical, Disabled, and OnActivate. Long text is truncated with an ellipsis | Root, Leading, Body, Label, Description |

When a Badge's label is empty, no text is created, no extra gap is left, and the padding becomes equal on all sides. Inside a FieldGroup or FormField, a Badge inherits the size and takes part in the group's corner rounding. If Avatar.Size is unset, badge sizes XS–XL use avatars of 16, 16, 16, 20, and 20 DIP. An explicit Leading icon takes precedence over the avatar.

If a Card's body builder is nil, no body is created, and separators appear only between regions that exist. When the window is narrower than 640 DIP, every region has 16 DIP of padding. From 640 DIP up, the horizontal padding and the body's vertical padding are 24 DIP, while the header and footer keep 16 DIP of vertical padding. All of these scale with Density. A description without a title keeps a 4 DIP top margin.

The header, body, and footer have separate identities. Removing the header at runtime does not rebuild the body or footer that remain, so inputs inside them keep focus and selection. Child controls still need stable keys and persistent app data.

Alert's Avatar takes precedence over Icon. The avatar defaults to a 44 DIP diameter; an explicit Avatar.Size overrides it. Actions sit below the body by default with a 10 DIP top margin; Horizontal places them beside the body. Switching the orientation or the leading content keeps the identity of the action controls that remain. Body is the text container, Actions is the action row, and Trailing is the close button.

A close button appears when OnClose or Close is non-nil. Close takes ButtonOptions and defaults to MD, Neutral, LinkVariant, and zero padding. By default, the button uses the alert's text color and fades on hover. CloseIcon defaults to X, and an explicit Close.Leading takes precedence. OnClose is called once per activation. The component does not hide itself; your app updates the visibility. For orientation settings, see [Orientation](options.md#orientation).

If every field of User's `AvatarOptions` is empty, no avatar is shown. To show initials, set `Avatar: nui.AvatarOptions{Alt: name}` or `Text`. At size XS, the description appears on the same line as the name. User acts as a keyboard-activatable button only when you pass `OnActivate`.

User.Size sets the size of the avatar and its chip, overriding the nested Avatar.Size and Chip.Size. Chip.Inset is always true; the other chip settings (position, color, text, and Hidden) are passed through. Without an avatar, no chip is shown. Leading replaces the default avatar area and Body replaces the name and description. Each region is built with a stable identity.

An activatable User emphasizes its text and scales the avatar to 1.15× on hover or keyboard focus. The scale transition follows Motion.Transition, and with reduced motion the final state appears immediately. Disabling the User or a parent prevents activation. OnActivate does not route or open URLs; your app implements those actions.

Empty's default padding is 16, 24, or 32 DIP depending on the window width, with breakpoints at 640 and 1024 DIP. The header, body, and footer are at most 384 DIP wide. The avatar defaults to 32, 36, 40, 44, or 48 DIP for Empty sizes XS–XL; an explicit Avatar.Size takes precedence. Avatar.Image always takes precedence over an icon, and an explicit Avatar.Icon takes precedence over Empty.Icon and the loading icon.

Header replaces the entire default header, Leading replaces only the leading content, and Body replaces the default Actions. Footer is shown separately. The action area and footer have stable identities, so changes to the header do not replace input controls inside them. The default title has level-2 heading semantics; for a custom Header, your app sets the semantics.

Loading affects only the leading icon. It does not disable Actions or change an existing avatar image. Reduced motion stops the rotation. There is no native equivalent of aria-busy. Set the loading icon's color with UI.Leading.Foreground; a late TextColor from Style does not replace the resolved custom-drawn color. Style the avatar through Avatar.UI. The Leading part keeps its default meaning (the icon); with a custom Leading, it refers to that content's container.

### Indicators, separators, and loading

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Avatar / AvatarGroup | A single avatar or a group of overlapping avatars. Group items use stable IDs, and the visible count can be limited | Avatar: Root, Body, Label, Leading. Group: Root, Item |
| Chip | A badge indicator with four corner positions, Inset, Hidden, colors, sizes, and text. Does not intercept clicks on its host; without content, it can be shown on its own | Root, Indicator |
| Separator | A horizontal or vertical line that can contain text, an icon, an avatar, or custom content | Root, Body, Label, Leading, Indicator |
| Kbd / FormatKey | Displays shortcut hints and platform modifier keys. Does not register shortcuts | Root, Label |
| Skeleton | A loading placeholder, 160×16 DIP by default, with configurable size, corner radius, and animation | Root |
| Progress / ProgressGroup | Single or segmented progress with sizes, colors, and orientation. Progress supports animation | Progress: Root, Indicator. ProgressGroup also has Segment, List, Item, Leading, Label, Trailing |

Separator's `Size` maps XS–XL to line widths of 1–5 DIP. `Type` is one of the Solid, Dashed, and Dotted constants, and `Position` is Center, Start, or End. Content precedence is `Content → Label → Icon → Avatar`. The default label is 14 DIP, the icon 20 DIP, and the avatar 20 DIP; an explicit Avatar.Size takes precedence. Leading refers to the icon or avatar root. Vertical labels can wrap. A vertical separator is 48 DIP tall by default, which you can change on the returned element; give long labels and interactive content enough height. The content container keeps its identity when the position or orientation changes, but controls inside it still need stable keys.

Separator currently draws a decorative line and leaves the accessibility of its text and interactive content to that content. MyGo has no static separator role, so this is not equivalent to upstream `decorative=false`, and the library does not add a Decorative option that would have no effect. Set the stroke color of dashed and dotted lines with `UI[Indicator].Background`; a Background set in the final Style changes only the element's background.

Kbd shows ordinary key names in uppercase and shows Tab, Backspace, Delete, CapsLock, the arrow keys, PageUp/PageDown, and Home/End as symbols. Modifier keys follow desktop conventions; for example, on platforms other than macOS, meta is shown as Win and shift as Shift.

`command` and `win` always show ⌘ and ⊞, regardless of Platform. `CmdOrCtrl` and `primary` show ⌘ on macOS and Ctrl elsewhere. Platform defaults to the current GOOS; set it to darwin, windows, or linux to show a specific platform. Kbd does not register shortcuts and is not in the default Tab order. The standard sizes are SM, MD, and LG; XS and XL are compatibility aliases for SM and LG.

The Neutral/Solid background of Button, Badge, Alert, and Kbd uses Theme.InvertedBackground, and the text uses Inverted. On hover, Button uses that background at 90% opacity; in high contrast mode it stays opaque. Instance UI and Style can still override the default look.

### Buttons and links

Button has six variants: `Solid`, `Outline`, `Soft`, `Subtle`, `Ghost`, and `LinkVariant`. Colors are `Primary`, `Secondary`, `Success`, `Info`, `Warning`, `Error`, and `Neutral`. The older `Neutral bool` field remains for compatibility; new code should use `Color: nui.Neutral`.

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| Default base height (DIP) | 24 | 28 | 32 | 36 | 40 |

Buttons grow taller when the font's line height is large. `Block` makes a button fill the available width and `Square` gives it a square layout. Icon-only buttons are square automatically and must have an `AccessibleLabel`, or they panic. Instance `UI` can still override the padding of a square button.

**Loading and disabled:** `Loading` blocks clicks and keyboard activation and keeps the label. By default it replaces the leading icon with a spinner, or adds one. `LoadingTrailing` moves the spinner to the trailing side, although an explicit Leading still takes precedence. `LoadingIcon` changes the icon. A button that had no icon may become wider while loading.

The spinner's color comes from the Root/Leading foreground style, so set it through `UI`. A `Style(...leading).TextColor` that runs last cannot change the already resolved color of the custom-drawn icon. `Active`, `ActiveColor`, and `ActiveVariant` control the active state.

MyGo draws disabled controls at 50% opacity, and the library adds no further transparency. A child with `Disabled: false` cannot override a disabled parent. Chip is purely decorative; if a count or online status matters, include it in the host's `AccessibleLabel` or `Alt`.

Chip's four corner positions are physical left and right and are not mirrored in RTL. Inset places the chip inside the host's bounds. When the content builder is nil, the chip is shown on its own without corner offsets. The text width is computed from the font, font size, font weight, and horizontal padding in Indicator.UI. Style runs after measurement, so if you change the text size there, set the width and height as well; a large custom font size usually also needs a taller chip.

**Links:** When Link's URL is non-empty, activating the link opens the address through `Services.OpenURL`. When the URL is empty, use `.OnClick` on the returned element for in-app navigation. If you set both URL and OnClick, keep MyGo's action override rules in mind.

### Images, avatars, and icons

Avatar chooses its content in the order `Bitmap → Icon → Text → initials from Alt`. Decode images outside the view build and pass the result of `ui.DecodeBitmap`. On failure, pass nil and the component falls back to the next option. Your app handles downloading, caching, cancellation, and errors.

The Body part is the bitmap. By default it fills the root container and is center-cropped (Cover). If Style changes the root size, the image still fills the available area. Sizes XS–XL map to 24, 28, 32, 36, and 40 DIP; adjust other sizes with a synchronous Style. The library does not add upstream's 3xs, 2xs, 2xl, and 3xl sizes to the shared enum. To add a badge to an avatar, combine them as `Chip(..., ChipOptions{Inset: true}, func() { Avatar(...) })`; Chip controls the color, position, and visibility.

`Text` is shown as is, for example `+12`. Without Text, Avatar takes the first character of each of the first two space-separated words in Alt, preserving case; without Alt, it stays empty. The text shrinks to fit the circle, but a full Unicode grapheme-aware abbreviation algorithm is not implemented. User's Name is the avatar's default Alt.

`Initials` is a separate native helper: it ignores runs of whitespace, takes the first and last words, and converts them to uppercase. For a single Chinese name, it takes the first two characters. Avatar does not call it, so the two can produce different results for empty values, multi-word names, and Chinese names. To use this rule, pass `Text: nui.Initials(name)` explicitly.

Every item in AvatarGroup and ProgressGroup needs a unique, non-empty ID. If AvatarGroup's Max is unset or ≤ 0, all avatars are shown. A positive Max limits the visible count and adds a same-sized `+n` avatar for the rest. The group's Size and Color only provide defaults for avatars that do not set their own. `Overlap: nui.Set(0)` turns off the overlap.

AvatarGroup's Item part styles both ordinary and overflow avatars and is applied after each avatar's own style. Item.BorderWidth and BorderColor control the ring that separates avatars; an explicit BorderWidth of 0 removes it. Reordering keeps each avatar's identity through its stable AvatarItem.ID.

Lucide icons are separate functions, such as `lucide.House()`, `lucide.GitBranch()`, and `lucide.FileJson2()`. The library includes all 2,015 SVGs from a pinned commit. Each icon is embedded and cached separately, and the Go linker can remove icons you do not reference. The returned SVG is a shared, read-only object; set its color and size through the component.

For the full list of names, the pinned commit, the import path, and the offline generation steps, see the [icon docs](../icons/lucide/README.md) and [manifest.json](../icons/lucide/manifest.json). Upgrading and generating icons is library maintenance work; apps simply call the icon functions.

### Progress and animation

**Skeleton**'s Width, Height, and Animate are native extensions. The default size is 160×16 DIP. The default height follows the text scale, but explicit sizes do not scale, and an explicit 0 stays 0. To fit the parent's width, call FillWidth on the returned element. Both Root.Radius and the final Style can set the corner radius, and the resulting radius clips the placeholder background and any attached content.

The pulse changes only the opacity of the custom-drawn background; set its color with UI.Root.Background. Style.Background configures the underlying surface and cannot replace the resolved pulse color. Skeleton provides a localized loading name and a native status role, but not the full aria-busy/alert announcement semantics of the web. Your app decides when to show and hide placeholders and how to present the loaded result.

**Progress**'s Max defaults to 100. Explicitly setting it to a non-positive value, NaN, or Inf panics. A progress value of NaN or Inf is treated as 0 and then clamped to the range 0 to Max.

Determinate progress fills to the rounded integer percentage, while assistive technology receives the exact amount completed. The indeterminate indicator is half the track's length and moves in a two-second cycle by default; with reduced motion or Animate=false, it stays still in the center. Internally, `Now=-1 < Min=0` represents unknown progress, with an "In progress" value or description. Windows UIA clamps the unknown value to the minimum; how screen readers announce it has not been verified.

Horizontal progress fills from the start of the text direction. Vertical progress is 160 DIP tall by default and fills from the top. `Inverted` reverses the fill direction. Both `Orientation` and `Vertical` control orientation. Neutral uses InvertedBackground. The Indicator's Background overrides the custom-drawn fill color; other Indicator fields have no effect. Root is the track, and its style does not change the indicator's fixed pill shape.

Label is the accessible name, and Value on the returned element can describe the value in work units. To show a visible percentage or step text, compose Text yourself. The upstream status slot, the Max array of steps, and the other three CSS animation presets are not provided. Progress only displays the value you pass in; it does not change business state and has no value-change callback.

**ProgressGroup** shows several segments on a single track. Each segment is first clamped to 0 through Max, and non-finite values are treated as 0. If the total is less than Max, the rest of the track stays empty; if it is more, segments are shortened proportionally, and the total reported to assistive technology is also capped at Max.

The legend shows each segment's rounded percentage of Max. A legend is created as soon as any item has a Label, Icon, or Build. Without an Icon, an item shows a color dot. Build synchronously builds the whole legend row for that item and replaces the default icon, label, and percentage; controls inside it keep their identity through the stable ID. HideLegend unmounts the legend content.

Each item's UI can configure Segment, Item, Leading, Label, and Trailing, overriding the group-level UI of the same name. The group-level List styles the legend list and Indicator styles the whole track. Segment.Background can be any color. Neutral segments and dots use InvertedBackground, and Neutral icons use Highlighted. For sizes XS–XL, the legend font size is 12/14/14/14/16, icons are 12/16/16/16/20, and dots are 6/8/8/8/10 DIP, all following the text scale.

Horizontal groups follow RTL. Orientation defaults to Horizontal; in vertical layout, a Height of 0 or less uses 160 DIP. Segments are static and the group exposes overall progress to assistive technology. There is no per-segment progress role, built-in status text, or segment click event. You can add actions or descriptions with the legend's Build; scheduling the underlying work remains up to your app.

**When animations stop:** Continuous animations request the next frame only while a visible component actually draws. They stop when the component scrolls out of view, unmounts, has Animate turned off, or when reduced motion is enabled. Button's background transition follows `Motion.Transition/Reduce`, and `Animate: nui.Set(false)` controls the loading spinner; once a transition ends, no more frames are requested. Components do not use timer goroutines for animation.

## Limitations

Styles are set through Go options and part APIs; CSS selectors and Vue slots are not supported. For platform state, IME, and assistive technology verification, see [Compatibility and known limitations](compatibility.md).
