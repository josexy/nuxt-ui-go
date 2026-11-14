<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Tabs

Switch between related content panels.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Tabs`
- Types: `TabsOptions`, `TabItem`
- Theme key: `nui.ComponentTabs` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Stable TabItem IDs identify both selection and widget state. Build panels with Content, and configure labels using Avatar, Badge, BadgeOptions and UI. TabsOptions.Content=false omits panels; KeepMounted controls local state in inactive panels.

Underline widths round from DIPs to whole device pixels and align with the list border edge.

## Things to know

Empty, missing or disabled selection falls back to the first enabled item without OnChange. An all-disabled list preserves the business ID. Content=false unmounts all panels even with KeepMounted=true. Badges default to SM / Neutral / Outline, with height depending on font metrics. Panels are focusable tab panels (ui.RoleTabPanel) named after their tab’s Label, and each tab points at its built panel with Controls. Sliding indicator animation is not provided.

## API

```go
func Tabs(c *ui.Context, t Theme, value *string, items []TabItem, o TabsOptions) ui.Element
```

### TabsOptions

```go
type TabsOptions struct {
	Label                            string
	Color                            Color
	Size                             Size
	Underline, Vertical, KeepMounted bool
	Content                          Optional[bool] // Defaults to true.
	ActivationMode                   ActivationMode
	OnChange                         func(string)
	Decoration
}
```

### TabItem

```go
type TabItem struct {
	ID, Label    string
	Icon         *ui.SVG
	Avatar       *AvatarOptions
	Badge        string
	BadgeOptions *BadgeOptions
	Disabled     bool
	Content      func(*ui.Context, Theme)
	UI           Styles
}
```

### Related types

```go
type ActivationMode uint8

const (
	ActivationAutomatic ActivationMode = iota
	ActivationManual
)
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `BadgeOptions` → [badge.md](badge.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Shared conventions › Open state and stable IDs (excerpt)

List item IDs must be non-empty and unique, or the component panics. Tabs, Accordion, Breadcrumb, Stepper, and Timeline require uniqueness within their own list; menus require it across the whole menu tree. Display names can repeat. Required state pointers must not be nil, except for Banner, which explicitly accepts nil to mean "always shown".

#### Tabs, Collapsible, and Accordion

##### Choose a component and its state

| Component | Bound state | Default behavior |
| --- | --- | --- |
| Tabs | `*string`, the current item ID | Pill style; only the current content is shown. Underline and Vertical are available |
| Collapsible | `*bool`, whether it is expanded | Toggles on click or Enter/Space; expanding and collapsing take 200 ms by default |
| Accordion | `*[]string`, the IDs of expanded items | Only one item can be open at a time by default, and it can be closed again. Multiple allows several |

TabItem provides ID, Label, Icon, Avatar, Badge, BadgeOptions, Disabled, Content, and an item UI. Icon takes precedence over Avatar. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. Badge is a string, and `"0"` is shown. It defaults to SM / Neutral / Outline and can be overridden with BadgeOptions. The body callbacks of Tabs and Accordion receive a Context and a Theme; Collapsible's body receives a Theme and uses the outer `c`.

Tabs use Left/Right or Up/Down arrows depending on the orientation, wrapping around and skipping disabled items. In horizontal tabs, Left and Right swap in RTL, and keys for the other axis do not change the selection. If the current item is removed or disabled, the selection moves to the first available item. If no item is available, the business ID is kept but no content is shown. If the content that has focus is switched, hidden, or removed, focus returns to a valid tab. The content is a focusable tab panel (`ui.RoleTabPanel`) named after its tab's Label, and the tab points at it with `Controls`, as in ARIA's Tabs pattern, so Tab first enters the panel and then its inner controls. Hidden kept-mounted panels are left out of what assistive technology reads, and tabs whose panel is not built (no Content, or Content=false) control nothing. Earlier versions named the panel "{name} content"; an app that overrides the `nuxtui.contentNamed` message keeps its own panel name. In Accordion, ↑↓, Home, and End move between headers, and Tab keeps the normal order of headers and content.

Tabs.OnChange receives the new ID, Collapsible.OnChange receives a bool, and Accordion.OnChange has no parameters; your app reads the ID list. The user-change callbacks do not fire when the app changes the value directly, when Tabs corrects an invalid ID, or when Accordion's single mode trims extra IDs.

Accordion keeps IDs in the app's open list that are temporarily missing, so the item opens again if the same ID returns. Your app removes IDs that are gone for good. For manual Tabs activation and Accordion's Collapsible option, see [Sliders, tabs, and accordions](slider.md).

Collapsible.Disabled, Accordion.Disabled, and AccordionItem.Disabled affect only the headers; expanded content keeps the disabled settings of its own controls. If content being edited is closed from outside, focus returns to an available header. If Accordion removes the focused item, focus returns to the first available header. Content that is animating closed stops receiving input immediately, so it cannot trigger actions while collapsing.

##### Keeping hidden content mounted

By default, only the active Tabs content is built, and collapsed content unmounts after its exit animation. The app model is kept, but editor selections, undo history, and local UI state may be lost on unmount.

KeepMounted=true keeps all content in every build and hides inactive content with Hidden, so MyGo's local state is preserved. Hidden items take no part in layout or focus, but their content callbacks still run. In keep-mounted collapsible content, the body hides immediately while the arrow still follows the native animation. Removing an entire item still unmounts it.

When KeepMounted changes while a Collapsible or Accordion is open, the content keeps the same identity and the editor's local state is not rebuilt. Theme.Motion.Reduce or the system's reduced motion setting completes height and icon changes immediately. The default animation uses MyGo's 200 ms interpolation, which is not frame-for-frame identical to upstream's CSS animation.

TabsOptions.Content defaults to true. An explicit false builds no panels, which unmounts local UI state even with KeepMounted=true; drafts your app stores are unaffected. When re-enabled, the panels are built from app data. Do not start I/O or submit anything unconditionally from content callbacks.

##### Appearance

In the pill style, the active Tabs item uses the semantic color as its background with inverted text; Neutral uses InvertedBackground. A horizontal list fills the width. Underline corresponds to Nuxt UI's link style: the list has a 1 DIP baseline and an active indicator line, at the bottom when horizontal and on the start side when vertical, following RTL. Vertical tabs have equal widths, and the list and content are center-aligned. For XS–XL, the trigger gap is 4/6/6/8/8 DIP and the text line height is 16/16/20/20/24 DIP.

Indicator styles the active background or indicator line, Trailing styles the badge, and TabItem.UI overrides that item's parts. Existing Trigger.Background overrides still work; Indicator.Background is the more specific setting for the active background. The indicator currently updates directly with the selection, without upstream's 200 ms slide between tabs. There are no arbitrary leading/default/trailing or list-leading/list-trailing slots; place other actions in a toolbar outside Tabs.

In the pill style, Indicator and the active Trigger share one native element so the background stays behind the text and avatar. In the underline style, Indicator is a separate line. Configure the Style callback for these actual parts; do not assume every style entry point is a separate node.

Accordion and Collapsible triggers have no hover background and no horizontal padding. They have 14 DIP of vertical padding, a 20 DIP icon, and a 6 DIP gap. Accordion content uses the default text color with a 14 DIP bottom margin.

TrailingIcon defaults to ChevronDown. AccordionItem.TrailingIcon takes precedence over AccordionOptions.TrailingIcon. AccordionItem.UI can override that item's Item, Header, Trigger, Leading, Label, Trailing, and Body. Empty Content creates no body and no bottom margin. The body layout container and the animation clipping container do not have separate style entry points.

Nuxt UI's Collapsible uses a trigger slot supplied by the app. Go's Label, Icon, TrailingIcon, default header, and body spacing are native extensions. For an arbitrary trigger composition, use MyGo's Base APIs.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Tabs | Root, List, Trigger, Leading, Label, Trailing, Indicator, Body. TabItem.UI styles a single item |

Tabs, Stepper, and Timeline accept global and instance Color and Size; Banner and Toast accept Color. Color on menu items and Timeline items is a per-item setting. Other styles are configured through the parts listed above.

### Options and defaults

#### Common options by component › Sliders, tabs, and accordions (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Tabs | ActivationMode defaults to ActivationAutomatic. With ActivationManual, arrow keys only move focus, and Enter or Space activates a tab and fires a single notification |

#### Common options by component › Overlays and notifications (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Tabs | Content defaults to true; false unmounts all panels. KeepMounted keeps only the inactive panels mounted. TabItem supports Avatar, Badge/BadgeOptions, and UI. Indicator styles the active background or underline |

#### Examples (excerpt)

The **Orientation** page demonstrates orientation switching.

### Advanced interaction, charts, and layout

#### Carousel and Marquee › Stopping animation when hidden (excerpt)

Neither component builds the view on every frame, and neither uses timer goroutines: Carousel autoplay asks for one build per interval, and Marquee moves its content in frames that only repaint, building about once a second. Visibility is what showed in the previous frame within the window and the scroll containers and clips around the component, so content scrolled out of a ScrollArea, clipped away, or kept mounted in a hidden panel (Tabs or Collapsible with KeepMounted) stops by itself. While a window is out of sight, MyGo draws no frames for it, so both stand still and continue from where the clock puts them when it shows again. Since MyGo v0.9.0 that covers a hidden, minimized, other-Space, or fully covered window on macOS and a minimized window on Windows and on Linux under X11. A window covered by other windows on Windows or Linux, or any window under Wayland, keeps drawing, and Marquee keeps moving there. Set Active=false for content your app hides by other means, such as covering it with another element, or stop building the component.

### Inputs, selection, and fields

#### State, themes, and events › Focus and callbacks (excerpt)

Since MyGo v0.9.0, focus moved by the keyboard or by a component scrolls the scroll containers around the newly focused control as little as needed to show it, as the web's `element.focus()` does. This covers roving focus in Tabs and menus, Form focusing its first invalid field, and focus returning to a trigger when Modal or DropdownMenu closes. A target that is already fully visible does not scroll. A control that a component keeps focused across builds does not undo the user's scrolling, and a pointer press focuses a control where it is without scrolling. To move focus from app code without scrolling, pass `ui.FocusOptions` to the element's Focus:

```go
cell := ui.Box(c.Key(id)).Focusable()
if app.restoreFocus == id {
	cell.Focus(ui.FocusOptions{PreventScroll: true})
}
```

### Compatibility and known limitations

#### Feature scope › Layout and navigation (excerpt)

| Area | Current behavior and limits |
| --- | --- |
| Tabs | Panels are tab panels named after their tab and linked from it with `Controls`. Real screen-reader output for this relation has not been verified |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Value, Draft               string
	Reversed, OverviewDisabled bool
	Changes                    int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var itemUI nui.Styles

	var parts nui.Styles

	mode := nui.ActivationAutomatic

	items := []nui.TabItem{{ID: "overview", Label: "Overview", Disabled: s.OverviewDisabled, Content: func(c *ui.Context, t nui.Theme) {
		ui.Text(c, "Project overview")
		nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Tab draft"})
	}}, {ID: "settings", Label: "Settings", UI: itemUI, Content: func(c *ui.Context, t nui.Theme) { ui.Text(c, "Project settings") }}}
	if s.Reversed {
		slices.Reverse(items)
	}
	nui.Tabs(c.Key("tabs"), t, &s.Value, items, nui.TabsOptions{Label: "Project tabs", Content: nui.Set(true), ActivationMode: mode, OnChange: func(string) { s.Changes++ }, Decoration: nui.Decoration{UI: parts}})
	nui.Button(c.Key("reverse"), t, "Reverse tabs", nui.ButtonOptions{Variant: nui.Outline}).AlignSelf(ui.Start).KeepFocus().OnClick(func() { s.Reversed = !s.Reversed })
	nui.Button(c.Key("disable-overview"), t, "Toggle Overview disabled", nui.ButtonOptions{Variant: nui.Soft}).AlignSelf(ui.Start).KeepFocus().OnClick(func() { s.OverviewDisabled = !s.OverviewDisabled })
	ui.Textf(c, "Selected: %s · Changes: %d", s.Value, s.Changes)
}

func main() {
	s := &example{}

	s.Draft = "Retained tab draft"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Tabs · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
