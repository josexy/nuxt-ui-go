<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Icon

Display a Lucide vector icon.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Icon`
- Types: `IconOptions`
- Theme key: `nui.ComponentIcon` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Choose a function from `icons/lucide`, such as `lucide.Check()`, and pass its SVG to `nui.Icon`. Size controls the displayed dimensions. Give a meaningful standalone icon a Label for assistive technology.

## Things to know

Icon size uses DIP. Icon functions return shared SVGs; set color and size on the component rather than modifying the SVG.

## API

```go
func Icon(c *ui.Context, t Theme, svg *ui.SVG, opts IconOptions) ui.Element
```

### IconOptions

```go
type IconOptions struct {
	Size  Optional[float32]
	Color Optional[ui.Color]
	Label string
	Decoration
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Page structure and actions (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Icon | Displays a `*ui.SVG` with a size, color, and Label. Without a Label, it is decorative | Root |

#### Components and parts › Displaying information (excerpt)

Alert's Avatar takes precedence over Icon. The avatar defaults to a 44 DIP diameter; an explicit Avatar.Size overrides it. Actions sit below the body by default with a 10 DIP top margin; Horizontal places them beside the body. Switching the orientation or the leading content keeps the identity of the action controls that remain. Body is the text container, Actions is the action row, and Trailing is the close button.

Empty's default padding is 16, 24, or 32 DIP depending on the window width, with breakpoints at 640 and 1024 DIP. The header, body, and footer are at most 384 DIP wide. The avatar defaults to 32, 36, 40, 44, or 48 DIP for Empty sizes XS–XL; an explicit Avatar.Size takes precedence. Avatar.Image always takes precedence over an icon, and an explicit Avatar.Icon takes precedence over Empty.Icon and the loading icon.

#### Components and parts › Indicators, separators, and loading (excerpt)

Separator's `Size` maps XS–XL to line widths of 1–5 DIP. Painting rounds the width to whole device pixels before aligning its position, so equal widths remain equal at fractional DPI and UI zoom. Dashed and dotted segments use whole-pixel lengths and gaps; the last segment may be shorter. Card section dividers and default menu/Listbox separators paint one device pixel while preserving DIP layout spacing. `Type` is one of the Solid, Dashed, and Dotted constants, and `Position` is Center, Start, or End. Content precedence is `Content → Label → Icon → Avatar`. The default label is 14 DIP, the icon 20 DIP, and the avatar 20 DIP; an explicit Avatar.Size takes precedence. Leading refers to the icon or avatar root. Vertical labels can wrap. A vertical separator is 48 DIP tall by default, which you can change on the returned element; give long labels and interactive content enough height. The content container keeps its identity when the position or orientation changes, but controls inside it still need stable keys.

#### Components and parts › Buttons and links (excerpt)

Buttons grow taller when the font's line height is large. `Block` makes a button fill the available width and `Square` gives it a square layout. Icon-only buttons are square automatically and must have an `AccessibleLabel`, or they panic. Instance `UI` can still override the padding of a square button.

#### Components and parts › Images, avatars, and icons

Avatar chooses its content in the order `Bitmap → Icon → Text → initials from Alt`. Decode images outside the view build and pass the result of `ui.DecodeBitmap`. On failure, pass nil and the component falls back to the next option. Your app handles downloading, caching, cancellation, and errors.

The Body part is the bitmap. By default it fills the root container and is center-cropped (Cover). If Style changes the root size, the image still fills the available area. Sizes XS–XL map to 24, 28, 32, 36, and 40 DIP; adjust other sizes with a synchronous Style. The library does not add upstream's 3xs, 2xs, 2xl, and 3xl sizes to the shared enum. To add a badge to an avatar, combine them as `Chip(..., ChipOptions{Inset: true}, func() { Avatar(...) })`; Chip controls the color, position, and visibility.

`Text` is shown as is, for example `+12`. Without Text, Avatar takes the first character of each of the first two space-separated words in Alt, preserving case; without Alt, it stays empty. The text shrinks to fit the circle, but a full Unicode grapheme-aware abbreviation algorithm is not implemented. User's Name is the avatar's default Alt.

`Initials` is a separate native helper: it ignores runs of whitespace, takes the first and last words, and converts them to uppercase. For a single Chinese name, it takes the first two characters. Avatar does not call it, so the two can produce different results for empty values, multi-word names, and Chinese names. To use this rule, pass `Text: nui.Initials(name)` explicitly.

Every item in AvatarGroup and ProgressGroup needs a unique, non-empty ID. If AvatarGroup's Max is unset or ≤ 0, all avatars are shown. A positive Max limits the visible count and adds a same-sized `+n` avatar for the rest. The group's Size and Color only provide defaults for avatars that do not set their own. `Overlap: nui.Set(0)` turns off the overlap.

AvatarGroup's Item part styles both ordinary and overflow avatars and is applied after each avatar's own style. Item.BorderWidth and BorderColor control the ring that separates avatars; an explicit BorderWidth of 0 removes it. Reordering keeps each avatar's identity through its stable AvatarItem.ID.

Lucide icons are separate functions, such as `lucide.House()`, `lucide.GitBranch()`, and `lucide.FileJson2()`. The library includes all 2,015 SVGs from a pinned commit. Each icon is embedded and cached separately, and the Go linker can remove icons you do not reference. The returned SVG is a shared, read-only object; set its color and size through the component.

For the full list of names, the pinned commit, the import path, and the offline generation steps and manifest.json. Upgrading and generating icons is library maintenance work; apps simply call the icon functions.

#### Components and parts › Progress and animation (excerpt)

The legend shows each segment's rounded percentage of Max. A legend is created as soon as any item has a Label, Icon, or Build. Without an Icon, an item shows a color dot. Build synchronously builds the whole legend row for that item and replaces the default icon, label, and percentage; controls inside it keep their identity through the stable ID. HideLegend unmounts the legend content.

### Options and defaults

#### Common options by component › Switches, tags, and ratings (excerpt)

Option.Icon works in CheckboxGroup and RadioGroup. With the indicator hidden, the icon appears above the item label; an ordinary Radio indicator is still a dot. Loop controls whether arrow keys wrap around; it defaults to false for checkbox groups and true for radio groups, and `nui.Set(false)` turns wrapping off. Both group types use a single Tab stop, with Left/Right arrows in horizontal layout and Up/Down arrows in vertical layout.

### Advanced interaction, charts, and layout

#### FileUpload and resource ownership › The component handles selection and state; your app handles transfer (excerpt)

| Option | Purpose |
| --- | --- |
| Icon / ShowIcon / Highlight | The upload entry icon, its visibility, and border highlighting. Neutral uses the inverted color, and a field error uses the error color first |

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns › FooterColumns (excerpt)

FooterColumns takes columns with stable, non-empty IDs, and link IDs must be unique within each column. A non-empty Title is shown as a level-3 heading. Items use NavigationItem's ID, Label, Icon, Active, Disabled, OnSelect, and Decoration. The other fields (groups, descriptions, nested children, avatars, badges, Tooltip, and Popover) are for NavigationMenu and are not used by this component.

### Forms, dates, and data components

#### Listbox, SelectMenu, and async search › Listbox: a selectable list shown inline (excerpt)

Icon, Avatar, and Chip are checked in that order for the leading content, which aligns with the first line of text. SelectedIcon overrides the default check icon. Per-item options provide UI, Style, and OnSelect. The callback receives the ID and the selected state before the change, and can call PreventDefault to cancel this selection and the component's OnSelect. For the row appearance, see [Option row appearance](select.md). An MD row is about 32 DIP tall, plus one line when it has a description.

#### Tree › Migrating from the old Outline wrapper (excerpt)

A loaded empty branch keeps its expansion state but does not get the default folder icon and arrow; set Icon explicitly if you want a folder look. Type-ahead search cycles through matches when the same character is repeated, and keeps the current match while the user continues typing a prefix.

### Inputs, selection, and fields

#### Checkbox, CheckboxGroup, RadioGroup, and Switch › Options and keyboard (excerpt)

Both group types use ChoiceGroupOptions, which adds Horizontal, Loop, and a parameterless OnChange. Groups are vertical by default; Horizontal lets items wrap, except with ChoiceTable. Each option's Label and Description provide its text. In checkbox groups, Option.Icon replaces the check icon; in both group types with the indicator hidden, it appears above the label. Option.Build has no effect in choice groups; it is only used by Select and InputMenu.

A group's visible title can come from the enclosing FormField.Label. To use one icon for all checkboxes in a group, set the same SVG as Option.Icon on every option that has no icon of its own; items with their own icons keep them.

#### Checkbox, CheckboxGroup, RadioGroup, and Switch › Icons, styles, and spacing

Checkbox's CheckedIcon replaces the default check mark, and CheckboxTri's IndeterminateIcon replaces the minus sign shown for the indeterminate state. IndicatorHidden places an explicit CheckedIcon above the label, with the icon and text centered; without an icon, only the text is shown. The indeterminate icon is used only in a visible indicator; it does not replace the label icon in hidden mode. UncheckedIcon applies only to Switch.

Checkbox's Leading part styles the check icon, or the label icon in hidden mode. ChoiceCard and ChoiceTable show the keyboard focus ring on the outer frame, while ChoiceList shows it on the visible indicator. The Switch track has no outline by default; an accent or error ring appears only with Highlight or in the error state.

A card with a visible indicator shows Highlight only on the indicator; with the indicator hidden, the card's frame shows the highlight. Card styles follow the usual override order: component defaults, the instance's semantic options, the instance's UI, and then Style.

Mark a field as required with FormField.Required. It shows an asterisk and gives assistive technology a "required" hint along with the field's description and the control's description. It does not validate the selection; configure FormState validation rules to enforce it on submit.

Items in a group are 8 DIP apart horizontally. Vertically, the gap is 2/2/4/4/6 DIP for sizes XS to XL. Adjacent ChoiceTable items share a 1 DIP border, the outer corners have an 8 DIP radius, and the corners follow the writing direction. The border after a selected item is omitted so the selected item's full border is visible.

The Checkbox/Radio indicator container is as tall as the text line: 16/16/20/20/24 DIP by size. The inner Radio dot is 4/4/6/6/8 DIP.

A regular label and description are stacked using their own line heights, with no extra gap; the centered layout used when the indicator is hidden has a 4 DIP gap. In CheckboxTri's indeterminate state, the indicator uses the accent color but the card background stays the same as unchecked; only Checked applies the selected background.

| Component | Parts |
| --- | --- |
| Checkbox | Root, Indicator, Leading, Label, Description |
| Switch | Root, Indicator, Thumb, Label, Description |
| CheckboxGroup | Root and Item apply to the group and its rows. Other parts and Style are passed to each Checkbox; you can also customize them with ComponentCheckbox |
| RadioGroup | Root, Item, Indicator, Leading (the icon in hidden mode), Label, Description |

#### Select and InputMenu › Choose an interaction (excerpt)

Select and InputMenu share `Option[T comparable]`, which has Value, Label, Description, Disabled, Icon, and Build. Value is the stable identity: it must be unique and equal to itself (so NaN is not allowed). Labels can repeat. Even with a custom Build, fill in Label for search and assistive technology.

#### Select and InputMenu › Select: single selection and empty values (excerpt)

SelectOptions includes Label, Placeholder, ErrorMessage, Disabled, Invalid, Color, Variant, Size, Popup, OnChange, and Decoration. Leading and Trailing set the trigger icons; Trailing defaults to ChevronDown. Avatar is shown when there is no leading icon. SelectedIcon replaces the check mark on the selected item, not the item's own Icon.

#### Select and InputMenu › Option row appearance (excerpt)

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| Icon size (DIP) | 16 | 16 | 20 | 20 | 24 |

### Overlays, menus, and navigation

#### DropdownMenu and ContextMenu › Menu items (excerpt)

| MenuItem field | Behavior |
| --- | --- |
| ID / Label / Description / Icon | Stable identity, name, description, and icon. The ID is unique across the whole menu tree |

#### Toast › IDs, duration, and actions (excerpt)

| Option or action | Behavior |
| --- | --- |
| Avatar / Icon | The avatar takes precedence. The default avatar diameter is 44 DIP and the icon is 20 DIP. An explicit avatar size uses Avatar's own size |

You can use the single action and Actions together. Each button in Actions can set ID, Label, OnAction, ButtonOptions, and Close; buttons default to XS and Solid, and inherit the toast's color. Provide stable IDs when actions are updated or reordered; non-empty IDs must be unique within a toast. Without IDs, actions are identified only by their position in a static list. Changing the orientation keeps focus on the current action button. Without an Icon or Avatar, no status icon is added automatically.

#### Tabs, Collapsible, and Accordion › Choose a component and its state (excerpt)

TabItem provides ID, Label, Icon, Avatar, Badge, BadgeOptions, Disabled, Content, and an item UI. Icon takes precedence over Avatar. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. Badge is a string, and `"0"` is shown. It defaults to SM / Neutral / Outline and can be overridden with BadgeOptions. The body callbacks of Tabs and Accordion receive a Context and a Theme; Collapsible's body receives a Theme and uses the outer `c`.

#### Tabs, Collapsible, and Accordion › Appearance (excerpt)

Nuxt UI's Collapsible uses a trigger slot supplied by the app. Go's Label, Icon, TrailingIcon, default header, and body spacing are native extensions. For an arbitrary trigger composition, use MyGo's Base APIs.

#### Breadcrumb, Pagination, Stepper, and Timeline › Breadcrumb: show the current location (excerpt)

BreadcrumbItem contains ID, Label, Icon, Avatar, Active, Disabled, OnSelect, and UI. The options provide Label, Color, SeparatorIcon, and MaxItems. Items with OnSelect are actionable links, including the last one: the pointer and Enter activate them, while Space, as on the web, does not. Items with a nil callback are static text. Your app provides the path and the navigation actions.

When Active is unset, only the last item is active; an explicit false turns off the active style. Color defaults to Primary, and with Neutral the active text uses Highlighted. Inactive items are 14 DIP Muted text without padding, and actionable items change to the Text color on hover. The separator arrow is 20 DIP, and items are 6 DIP apart. In RTL, the default arrow is reversed, while a custom SeparatorIcon is kept as is. Icon takes precedence over Avatar, and the default avatar diameter is 20 DIP. The path stays on one line, and long text is truncated to fit.

#### Breadcrumb, Pagination, Stepper, and Timeline › Stepper: choose a step in a process (excerpt)

Stepper binds the current step as a `*string`. StepItem contains ID, Title, Description, Icon, Disabled, UI, and Content. An unknown ID means the process has not started, and the first item is not selected automatically. Content synchronously builds the current step's body; switching steps unmounts the old body. Your app stores business drafts. Reordering or changing orientation keeps the body's identity for the same ID.

#### Breadcrumb, Pagination, Stepper, and Timeline › Timeline: display a timeline (excerpt)

TimelineItem contains ID, Title, Description, Time, Icon, Color, and Content, and can also provide Avatar and UI. Timelines are vertical by default; Horizontal places the indicator icon and connecting line above the text.

#### Banner and ErrorView › ErrorView: an error page (excerpt)

In Go, `nui.Error` is already a semantic color, so the error display function is called ErrorView. ErrorOptions provides Status, Title, Description, RetryLabel, Icon, Clear, Retry, OnRetry, Actions, and Decoration.

A 40 DIP icon is shown only when you pass Icon. The status code defaults to 16/24 DIP semibold Primary and is styled with StatusSlot. The title uses a 36 DIP font when the window is narrower than 640 DIP and 48 DIP otherwise. The description is 18/28 DIP in Muted. An empty title creates no text, and a description identical to the title is omitted.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type example struct{}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Icon(c.Key("icon"), t, lucide.Check(), nui.IconOptions{Size: nui.Set(float32(40)), Label: "Example icon"}).AlignSelf(ui.Start)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Icon · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
