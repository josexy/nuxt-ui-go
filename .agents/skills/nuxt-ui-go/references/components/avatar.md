<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Avatar

Display an image, icon or name initials as an avatar.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Avatar`, `Initials`
- Types: `AvatarOptions`
- Theme key: `nui.ComponentAvatar` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Avatar chooses an image, icon, Text, then initials from Alt. Load and decode images outside the component; pass nil on failure to use fallback content. Wrap it in Chip with Inset set to true to place a status chip inside the avatar edge.

## Things to know

Alt supplies the accessible image name, falling back to Text. XS–XL diameters range from 24 to 40 DIP. Network loading, caching, retries and HTML image attributes belong to the application or platform.

## API

```go
func Avatar(c *ui.Context, t Theme, opts AvatarOptions) ui.Element

// Initials returns up to two Unicode initials; a one-word CJK name uses its first two runes.
func Initials(name string) string
```

### AvatarOptions

```go
type AvatarOptions struct {
	Image     *ui.Bitmap
	Icon      *ui.SVG
	Text, Alt string
	Color     Color
	Size      Size
	Decoration
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Page structure and actions (excerpt)

Button's `Avatar: &nui.AvatarOptions{Alt: "Ada Lovelace"}` shows a leading avatar. A leading icon or a leading loading spinner takes precedence over it; a trailing loading spinner does not hide the avatar. If Avatar.Size is unset, button sizes XS–XL use avatars of 16, 16, 20, 20, and 24 DIP; an explicit Avatar.Size takes precedence. An avatar-only button still needs an AccessibleLabel. In this mode, the Leading part is the avatar's root element.

#### Components and parts › Displaying information (excerpt)

When a Badge's label is empty, no text is created, no extra gap is left, and the padding becomes equal on all sides. Inside a FieldGroup or FormField, a Badge inherits the size and takes part in the group's corner rounding. If Avatar.Size is unset, badge sizes XS–XL use avatars of 16, 16, 16, 20, and 20 DIP. An explicit Leading icon takes precedence over the avatar.

Alert's Avatar takes precedence over Icon. The avatar defaults to a 44 DIP diameter; an explicit Avatar.Size overrides it. Actions sit below the body by default with a 10 DIP top margin; Horizontal places them beside the body. Switching the orientation or the leading content keeps the identity of the action controls that remain. Body is the text container, Actions is the action row, and Trailing is the close button.

If every field of User's `AvatarOptions` is empty, no avatar is shown. To show initials, set `Avatar: nui.AvatarOptions{Alt: name}` or `Text`. At size XS, the description appears on the same line as the name. User acts as a keyboard-activatable button only when you pass `OnActivate`.

User.Size sets the size of the avatar and its chip, overriding the nested Avatar.Size and Chip.Size. Chip.Inset is always true; the other chip settings (position, color, text, and Hidden) are passed through. Without an avatar, no chip is shown. Leading replaces the default avatar area and Body replaces the name and description. Each region is built with a stable identity.

Empty's default padding is 16, 24, or 32 DIP depending on the window width, with breakpoints at 640 and 1024 DIP. The header, body, and footer are at most 384 DIP wide. The avatar defaults to 32, 36, 40, 44, or 48 DIP for Empty sizes XS–XL; an explicit Avatar.Size takes precedence. Avatar.Image always takes precedence over an icon, and an explicit Avatar.Icon takes precedence over Empty.Icon and the loading icon.

Loading affects only the leading icon. It does not disable Actions or change an existing avatar image. Reduced motion stops the rotation. There is no native equivalent of aria-busy. Set the loading icon's color with UI.Leading.Foreground; a late TextColor from Style does not replace the resolved custom-drawn color. Style the avatar through Avatar.UI. The Leading part keeps its default meaning (the icon); with a custom Leading, it refers to that content's container.

#### Components and parts › Indicators, separators, and loading (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| Avatar / AvatarGroup | A single avatar or a group of overlapping avatars. Group items use stable IDs, and the visible count can be limited | Avatar: Root, Body, Label, Leading. Group: Root, Item |

Separator's `Size` maps XS–XL to line widths of 1–5 DIP. Painting rounds the width to whole device pixels before aligning its position, so equal widths remain equal at fractional DPI and UI zoom. Dashed and dotted segments use whole-pixel lengths and gaps; the last segment may be shorter. Card section dividers and default menu/Listbox separators paint one device pixel while preserving DIP layout spacing. `Type` is one of the Solid, Dashed, and Dotted constants, and `Position` is Center, Start, or End. Content precedence is `Content → Label → Icon → Avatar`. The default label is 14 DIP, the icon 20 DIP, and the avatar 20 DIP; an explicit Avatar.Size takes precedence. Leading refers to the icon or avatar root. Vertical labels can wrap. A vertical separator is 48 DIP tall by default, which you can change on the returned element; give long labels and interactive content enough height. The content container keeps its identity when the position or orientation changes, but controls inside it still need stable keys.

#### Components and parts › Images, avatars, and icons

Avatar chooses its content in the order `Bitmap → Icon → Text → initials from Alt`. Decode images outside the view build and pass the result of `ui.DecodeBitmap`. On failure, pass nil and the component falls back to the next option. Your app handles downloading, caching, cancellation, and errors.

The Body part is the bitmap. By default it fills the root container and is center-cropped (Cover). If Style changes the root size, the image still fills the available area. Sizes XS–XL map to 24, 28, 32, 36, and 40 DIP; adjust other sizes with a synchronous Style. The library does not add upstream's 3xs, 2xs, 2xl, and 3xl sizes to the shared enum. To add a badge to an avatar, combine them as `Chip(..., ChipOptions{Inset: true}, func() { Avatar(...) })`; Chip controls the color, position, and visibility.

`Text` is shown as is, for example `+12`. Without Text, Avatar takes the first character of each of the first two space-separated words in Alt, preserving case; without Alt, it stays empty. The text shrinks to fit the circle, but a full Unicode grapheme-aware abbreviation algorithm is not implemented. User's Name is the avatar's default Alt.

`Initials` is a separate native helper: it ignores runs of whitespace, takes the first and last words, and converts them to uppercase. For a single Chinese name, it takes the first two characters. Avatar does not call it, so the two can produce different results for empty values, multi-word names, and Chinese names. To use this rule, pass `Text: nui.Initials(name)` explicitly.

Every item in AvatarGroup and ProgressGroup needs a unique, non-empty ID. If AvatarGroup's Max is unset or ≤ 0, all avatars are shown. A positive Max limits the visible count and adds a same-sized `+n` avatar for the rest. The group's Size and Color only provide defaults for avatars that do not set their own. `Overlap: nui.Set(0)` turns off the overlap.

AvatarGroup's Item part styles both ordinary and overflow avatars and is applied after each avatar's own style. Item.BorderWidth and BorderColor control the ring that separates avatars; an explicit BorderWidth of 0 removes it. Reordering keeps each avatar's identity through its stable AvatarItem.ID.

Lucide icons are separate functions, such as `lucide.House()`, `lucide.GitBranch()`, and `lucide.FileJson2()`. The library includes all 2,015 SVGs from a pinned commit. Each icon is embedded and cached separately, and the Go linker can remove icons you do not reference. The returned SVG is a shared, read-only object; set its color and size through the component.

For the full list of names, the pinned commit, the import path, and the offline generation steps and manifest.json. Upgrading and generating icons is library maintenance work; apps simply call the icon functions.

### Advanced interaction, charts, and layout

#### CommandPalette › Shortcuts and search (excerpt)

| Option or behavior | Description |
| --- | --- |
| Kbds / Prefix / Suffix / Avatar / Chip / Loading / Active | Command content and appearance. Kbds only display shortcuts, and Loading does not disable the command |

#### FileUpload and resource ownership › The component handles selection and state; your app handles transfer (excerpt)

FileUploadState.Files stores file metadata, not open file handles. UploadFile contains Path, Name, MIME, Size, a status, a Progress from 0 to 1, Error, and an app-supplied Avatar. Add assigns stable IDs that do not collide with existing entries and ignores entries with an identical Path. If you assign Files directly, your app must provide unique IDs. The component does not load or decode preview images itself.

| Option | Purpose |
| --- | --- |
| FileIcon / FileImage / UploadFile.Avatar | The fallback file icon and the app-supplied preview accessory. FileImage=false shows only FileIcon |

### Forms, dates, and data components

#### Listbox, SelectMenu, and async search › Listbox: a selectable list shown inline (excerpt)

Icon, Avatar, and Chip are checked in that order for the leading content, which aligns with the first line of text. SelectedIcon overrides the default check icon. Per-item options provide UI, Style, and OnSelect. The callback receives the ID and the selected state before the change, and can call PreventDefault to cancel this selection and the component's OnSelect. For the row appearance, see [Option row appearance](select.md). An MD row is about 32 DIP tall, plus one line when it has a description.

#### Listbox, SelectMenu, and async search › SelectMenu: searchable popup (excerpt)

Leading and Avatar configure the trigger's leading content, Trailing sets the trailing arrow, and Highlight emphasizes the border. Loading shows a spinner using LoadingIcon, and LoadingTrailing controls its position. The native loading status row is also kept, and existing results can still be selected. Error and EmptyText set the status text and the no-results message. ShowTrailing=false hides the whole trailing area.

#### Calendar, InputDate, and InputTime › Date and time styles and limits (excerpt)

The date and time inputs share Input's five sizes, five variants, colors, avatar, icons, loading state, disabled and read-only states, and Clearable. A leading icon or leading spinner takes precedence over Avatar. LoadingTrailing controls the spinner position, and an explicit Leading keeps a leading spinner. With Neutral, focus and Highlight use the inverted background color, and Highlight also works with None. Shadow overrides on Root from the theme or the instance are kept, and Style runs last.

### Inputs, selection, and fields

#### Input and Textarea › Common options (excerpt)

| Option | Purpose |
| --- | --- |
| Avatar | A leading avatar; Leading or a leading loading spinner takes precedence |

If Avatar.Size is unset, input sizes XS–XL use avatars of 16, 16, 20, 20, and 24 DIP; an explicit Avatar.Size takes precedence. A trailing spinner can appear together with a leading avatar. Changing the avatar or the spinner position does not reset the text, focus, or selection.

#### Select and InputMenu › Select: single selection and empty values (excerpt)

SelectOptions includes Label, Placeholder, ErrorMessage, Disabled, Invalid, Color, Variant, Size, Popup, OnChange, and Decoration. Leading and Trailing set the trigger icons; Trailing defaults to ChevronDown. Avatar is shown when there is no leading icon. SelectedIcon replaces the check mark on the selected item, not the item's own Icon.

Loading only shows a spinner; it does not disable selection. By default, LoaderCircle appears at the start; LoadingTrailing=true moves it to the end, but an explicit Leading keeps it at the start. LoadingIcon replaces the spinner icon, which uses the corresponding part's foreground color. If Avatar.Size is unset, the avatar diameter is 16/16/20/20/24 DIP by select size; an explicit avatar size takes precedence.

#### Select and InputMenu › InputMenu: drafts versus selected values (excerpt)

InputMenuOptions embeds InputOptions and adds options such as EmptyText, Popup, AllowCustom, and OnSelect. By default, clicking only focuses the input; typing, arrow keys, or the trailing arrow open the panel. Set OpenOnClick=true to open on click. The input keeps focus. Trailing can change the icon. Loading does not disable editing or selection. Avatar, LoadingTrailing, and Highlight work as in Input.

### Overlays, menus, and navigation

#### DropdownMenu and ContextMenu › Menu items (excerpt)

| MenuItem field | Behavior |
| --- | --- |
| Avatar / Loading | Loading takes precedence, then Icon, then Avatar. The loading icon does not disable the item |

MenuOptions.Size defaults to MD, can be inherited from the component theme, or can be set explicitly from XS to XL. Menus use the [option row style](select.md): at MD, the padding is 6 DIP and icons are 20 DIP. Descriptions use the same font size as labels, and icons and trailing content align with the first line. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. CheckedIcon defaults to Check and LoadingIcon to LoaderCircle. The submenu arrow follows RTL.

#### Toast › IDs, duration, and actions (excerpt)

| Option or action | Behavior |
| --- | --- |
| Avatar / Icon | The avatar takes precedence. The default avatar diameter is 44 DIP and the icon is 20 DIP. An explicit avatar size uses Avatar's own size |

You can use the single action and Actions together. Each button in Actions can set ID, Label, OnAction, ButtonOptions, and Close; buttons default to XS and Solid, and inherit the toast's color. Provide stable IDs when actions are updated or reordered; non-empty IDs must be unique within a toast. Without IDs, actions are identified only by their position in a static list. Changing the orientation keeps focus on the current action button. Without an Icon or Avatar, no status icon is added automatically.

#### Tabs, Collapsible, and Accordion › Choose a component and its state (excerpt)

TabItem provides ID, Label, Icon, Avatar, Badge, BadgeOptions, Disabled, Content, and an item UI. Icon takes precedence over Avatar. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. Badge is a string, and `"0"` is shown. It defaults to SM / Neutral / Outline and can be overridden with BadgeOptions. The body callbacks of Tabs and Accordion receive a Context and a Theme; Collapsible's body receives a Theme and uses the outer `c`.

#### Breadcrumb, Pagination, Stepper, and Timeline › Breadcrumb: show the current location (excerpt)

BreadcrumbItem contains ID, Label, Icon, Avatar, Active, Disabled, OnSelect, and UI. The options provide Label, Color, SeparatorIcon, and MaxItems. Items with OnSelect are actionable links, including the last one: the pointer and Enter activate them, while Space, as on the web, does not. Items with a nil callback are static text. Your app provides the path and the navigation actions.

When Active is unset, only the last item is active; an explicit false turns off the active style. Color defaults to Primary, and with Neutral the active text uses Highlighted. Inactive items are 14 DIP Muted text without padding, and actionable items change to the Text color on hover. The separator arrow is 20 DIP, and items are 6 DIP apart. In RTL, the default arrow is reversed, while a custom SeparatorIcon is kept as is. Icon takes precedence over Avatar, and the default avatar diameter is 20 DIP. The path stays on one line, and long text is truncated to fit.

#### Breadcrumb, Pagination, Stepper, and Timeline › Timeline: display a timeline (excerpt)

TimelineItem contains ID, Title, Description, Time, Icon, Color, and Content, and can also provide Avatar and UI. Timelines are vertical by default; Horizontal places the indicator icon and connecting line above the text.

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
	nui.Chip(c.Key("chip"), t, nui.ChipOptions{Hidden: !false, Inset: true, Color: nui.Success}, func() {
		nui.Avatar(c.Key("avatar"), t, nui.AvatarOptions{Alt: "Ada Lovelace", Style: func(slot nui.Slot, e ui.Element) {
		}})
	}).AlignSelf(ui.Start)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Avatar · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
