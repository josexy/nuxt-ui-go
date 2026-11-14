<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# NavigationMenu

Provide primary navigation and nested destinations.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `NavigationMenu`
- Types: `NavigationOptions`, `NavigationItem`, `NavigationState`, `NavigationEvent`
- Theme key: `nui.ComponentNavigationMenu` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep the current destination in an active ID and vertical expansion in NavigationState.Expanded. OnNavigate switches pages, OnSelect can cancel leaf navigation, and OnExpandedChange reports user expansion changes. The component does not resolve URLs; parents open their children.

Active indicator widths round from DIPs to whole device pixels so moving selection does not vary thickness.

## Things to know

Defaults to horizontal, Primary and Soft (pill). Compact only affects vertical navigation. Compact Tooltip/Popover retain the native true defaults, while upstream defaults to false. DefaultOpen initializes new nodes only; external Expanded edits are silent. Single limits sibling expansion; Collapsible only restricts closing in Single mode. Horizontal children use native DropdownMenu rather than a shared web mega-menu viewport, automatic route matching or its hover-delay model.

## API

```go
// NavigationMenu leaves routing to OnNavigate. IDs are unique across the whole tree.
func NavigationMenu(c *ui.Context, t Theme, active *string, s *NavigationState, items []NavigationItem, o NavigationOptions) ui.Element
```

### NavigationOptions

```go
type NavigationOptions struct {
	Label                 string
	Vertical, Compact     bool
	Color, HighlightColor Color
	// Soft is the pill appearance; LinkVariant omits its active background.
	Variant                     Variant
	Highlight, Disabled, Single bool
	Collapsible                 Optional[bool]
	TrailingIcon                *ui.SVG
	// Tooltip and Popover default to true for compact vertical navigation.
	Tooltip, Popover Optional[bool]
	TooltipOptions   *TooltipOptions
	PopoverOptions   *PopoverOptions
	// PopupPlacement defaults to the reading direction's trailing side.
	PopupPlacement   Optional[Placement]
	OnNavigate       func(string)
	OnSelect         func(*NavigationEvent)
	OnExpandedChange func([]string)
	Decoration
}
```

### NavigationItem

```go
type NavigationItem struct {
	ID, Label, Description, Group  string
	Icon                           *ui.SVG
	Avatar                         *AvatarOptions
	Badge                          string
	BadgeOptions                   *BadgeOptions
	Chip                           *ChipOptions
	TrailingIcon                   *ui.SVG
	Active, DefaultOpen, LabelOnly bool
	Tooltip, Popover               Optional[bool]
	TooltipOptions                 *TooltipOptions
	PopoverOptions                 *PopoverOptions
	Disabled                       bool
	Children                       []NavigationItem
	OnSelect                       func(*NavigationEvent)
	Decoration
}
```

### NavigationState

```go
type NavigationState struct {
	Expanded map[string]bool
	// contains filtered or unexported fields
}
```

### NavigationEvent

```go
type NavigationEvent struct {
	ID string
	// contains filtered or unexported fields
}

func (e *NavigationEvent) PreventDefault()
```

### Related types

```go
type Placement uint8

const (
	BottomStart Placement = iota
	BottomEnd
	TopStart
	TopEnd
	RightStart
	LeftStart
	BottomCenter
	TopCenter
	RightCenter
	RightEnd
	LeftCenter
	LeftEnd
)
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `BadgeOptions` → [badge.md](badge.md), `ChipOptions` → [chip.md](chip.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `PopoverOptions` → [popover.md](popover.md), `Theme` → [theme.md](theme.md), `TooltipOptions` → [tooltip.md](tooltip.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### NavigationMenu and Sidebar

##### What navigation and layout each handle

NavigationMenu displays navigation items. Sidebar provides a title, description, title actions, a scrollable body, a footer, and a collapse button. Your app stores the current location as the active ID and performs the actual page switch in OnNavigate(id); the components do not parse URLs.

NavigationItem contains a stable ID, text, an icon or avatar, a badge, a status dot, per-item styles, and Children. IDs are unique across the whole tree; removing an item clears its menu state and restores focus. A parent item expands its children and does not navigate. OnSelect can cancel navigation to a leaf item with PreventDefault; on success, OnNavigate fires once.

| Mode | How children are shown |
| --- | --- |
| Horizontal | A DropdownMenu shows multi-level menus, with Escape and focus restoration |
| Vertical | Levels expand in place, with arrow keys, Home, End, and Tab |
| Vertical compact | Items with an icon or avatar hide their text and are centered, keeping their accessible names. Items without one still show their label. The host controls the width |

The defaults are Color=Primary and Variant=Soft (pill); LinkVariant gives the link style. Highlight shows an active indicator line, and an unset HighlightColor inherits Color. An active descendant emphasizes its ancestors. Disabled combines the root, item, and inherited disabled states. Horizontal menus ignore Compact, and their arrow keys stop at the ends; vertical arrow keys wrap around.

In compact mode, Tooltip and the side Popover are on by default, as in the native implementation. Turn either off with `NavigationOptions.Tooltip/Popover = nui.Set(false)`. Per-item settings override the global ones, and TooltipOptions and PopoverOptions configure their appearance and behavior. In horizontal menus, tooltips are enabled per item through Tooltip.

The side Popover opens on hover (including across the gap from the anchor), on click, or with the arrow key pointing toward the panel; opening it with the keyboard focuses the first available child. Escape, the opposite arrow key, or choosing an item closes it and restores focus to the trigger. Children inside it can expand further. PopupPlacement changes the default position, which is useful for a sidebar on the right.

NavigationState.Expanded exposes the vertical expansion state, and DefaultOpen only initializes new nodes. Single=true lets only one branch per level be open; with Single, Collapsible=false prevents collapsing the current branch. User changes fire OnExpandedChange; initialization and app changes are silent. When a branch collapses, focus on a hidden child moves to its parent. Removing a root item moves focus to an available root item, and switching Compact keeps the expansion and the active ID.

Horizontal menus use native menus. There is no shared web mega menu viewport, route-based active matching, or root hover delay model.

##### Toggles, side, and callbacks

| Option | Default behavior |
| --- | --- |
| Collapsible | true; false prevents the user from collapsing the sidebar |
| Toggle | true; false hides the built-in button so your app can provide its own control elsewhere |
| Disabled | Disables built-in interaction; also inherits disabled ancestors |
| Side | Defaults to `nui.Set(nui.SideLeft)`, meaning the start side in the reading direction; SideRight means the end side |

In RTL, the border, button icon, handle, and Slideover all adjust together. Your app still has to place a right-side sidebar after the main content in the Row. Widths are in DIP; web percentages, rem, and DashboardGroup units are not used.

OnResize, OnCollapsedChange, and OnOpenChange fire only when a user action or an automatic close by the component changes the value; writing the state directly from the app does not call them. OnResize reports the expanded width, and collapse changes are reported separately by OnCollapsedChange.

##### Content regions and narrow windows

The HeaderContent, FooterContent, and ToggleContent callbacks receive a `SidebarScope`, which exposes Collapsed, Open, and Mobile, and can call Collapse(bool) and SetOpen(bool). HeaderContent and FooterContent take precedence over the older Header and Footer callbacks. HeaderContent replaces the default title and description. Actions builds a separate action area when the sidebar is expanded or shown as a menu in a narrow window.

When the title, description, title actions, and Toggle are all absent, no empty header is left. The body and footer have 16 DIP of vertical padding by default, with separators between the body and any header or footer that exists. Header and footer content must adapt to the compact width themselves; if you fully customize the header, you can hide the built-in Toggle. These callbacks build content synchronously and must not keep Elements or Contexts.

With `Breakpoint: nui.Set(float32(900))`, the desktop sidebar is hidden and an open button is shown whenever the window is narrower than 900 DIP. When Breakpoint is unset or 0, the desktop layout is always used.

| Narrow-window Mode | Presentation |
| --- | --- |
| SidebarSlideover (default) | Opens from the side |
| SidebarModal | A full-screen modal panel |
| SidebarDrawer | A bottom sheet 85% of the window height |

The overlay content is built in expanded mode and uses the matching component's Escape, overlay-click closing, scrolling, and focus trapping. When the window returns to desktop width, the overlay's focus trap is removed immediately; an ordinary close keeps the exit animation and respects reduced motion.

NavigationKey takes the current page ID; when it changes, the narrow-window panel closes by default. AutoClose=false turns this off. The disabled state blocks built-in actions and notifications but does not rewrite your controlled model. An external open button can set the focus target after closing through ReturnFocus. Keep calling Sidebar, including while it is closed. If the focused content is removed because the width becomes zero or the sidebar switches to a closed narrow-window mode, focus goes to ReturnFocus first, or else to a built-in button or the resize handle that is still visible. If you hide all internal entry points, provide an external button and ReturnFocus.

```go
// state stores Collapsed bool, Sidebar nui.SidebarState,
// Active string, and Navigation nui.NavigationState.
nui.Sidebar(c.Key("workspace-sidebar"), theme, &state.Collapsed,
    nui.SidebarOptions{
        Title: "Workspace",
        State: &state.Sidebar,
        Resizable: true,
        Breakpoint: nui.Set(float32(900)),
        NavigationKey: state.Active,
        FooterContent: func(c *ui.Context, t nui.Theme, scope nui.SidebarScope) {
            label := "Account"
            if scope.Collapsed {
                label = ""
            }
            nui.Button(c, t, label, nui.ButtonOptions{
                Leading: lucide.User(),
                AccessibleLabel: "Account",
                Variant: nui.Ghost,
                Color: nui.Neutral,
            })
        },
    }, func(c *ui.Context, t nui.Theme, compact bool) {
        nui.NavigationMenu(c.Key("navigation"), t, &state.Active,
            &state.Navigation, items,
            nui.NavigationOptions{Vertical: true, Compact: compact})
    })
```

Crossing the breakpoint unmounts and rebuilds the control tree. Drafts and navigation state in your app model are kept, but local scroll state inside controls is not carried over. There is currently no web mega menu, CSS breakpoints, NuxtLink compatibility layer, or DashboardGroup multi-panel coordination. Navigation uses the native Group role because MyGo has no public Navigation landmark enum yet.

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns › FooterColumns (excerpt)

FooterColumns takes columns with stable, non-empty IDs, and link IDs must be unique within each column. A non-empty Title is shown as a level-3 heading. Items use NavigationItem's ID, Label, Icon, Active, Disabled, OnSelect, and Decoration. The other fields (groups, descriptions, nested children, avatars, badges, Tooltip, and Popover) are for NavigationMenu and are not used by this component.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| NavigationMenu | NavigationState, active ID `*string` | Root, List, Panel, Body, Label, Item, ItemLeadingSlot, ItemLabelSlot, Description, ItemTrailingSlot, Indicator |

#### Examples (excerpt)

It is 256 DIP wide by default, can be dragged between 200 and 360 DIP, and collapses to a 64 DIP icon bar. Switching pages keeps the width and collapsed state, which are stored per window; this sidebar is independent of the demo sidebar on the Application layout page.

### Options and defaults

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| NavigationMenu | Vertical Compact mode adds tooltips and child popovers by default; turn either off with `nui.Set(false)`. PopupPlacement changes where child popups appear |

### Compatibility and known limitations

#### Feature scope › Layout and navigation (excerpt)

| Area | Current behavior and limits |
| --- | --- |
| NavigationMenu / Sidebar | Icon navigation, tooltips, child popovers, resizing, and switching to an overlay based on the window width in DIP. The app implements routing and persistence |

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

type example struct {
	Active, Result      string
	Navigation          nui.NavigationState
	Items               []nui.NavigationItem
	Protect             bool
	Actions, Expansions int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	s.Protect = true
	direction := ui.LeftToRight

	frame := ui.Column(c.Key("direction")).Direction(direction)

	frame.Children(func() {
		nui.NavigationMenu(c.Key("navigation"), t, &s.Active, &s.Navigation, s.Items, nui.NavigationOptions{Label: "Workspace", Variant: nui.Soft, Collapsible: nui.Set(true), Tooltip: nui.Set(true), Popover: nui.Set(true), OnNavigate: func(id string) { s.Actions++; s.Result = "Opened " + id }, OnExpandedChange: func([]string) { s.Expansions++ }})
	})
	nui.Button(c.Key("collapse"), t, "Collapse Settings externally", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
		if s.Navigation.Expanded != nil {
			s.Navigation.Expanded["settings"] = false
		}
	})
	ui.Textf(c, "Active: %s", s.Active)
	ui.Textf(c, "Actions: %d · Expansion changes: %d", s.Actions, s.Expansions)
	ui.Text(c, s.Result)
}

func main() {
	s := &example{}

	s.Active = "home"
	s.Result = "Ready"
	s.Items = []nui.NavigationItem{
		{ID: "home", Label: "Overview", Icon: lucide.House(), Group: "Workspace", Badge: "4"},
		{ID: "projects", Label: "Projects", Icon: lucide.Folder(), Chip: &nui.ChipOptions{Color: nui.Success, Inset: true}},
		{ID: "settings", Label: "Settings", Icon: lucide.Settings(), DefaultOpen: true, Children: []nui.NavigationItem{
			{ID: "members", Label: "Members", Description: "Invite people and assign permissions", Avatar: &nui.AvatarOptions{Text: "LC", Alt: "Lin Chen"}},
			{ID: "protected", Label: "Protected destination", Icon: lucide.Lock(), OnSelect: func(e *nui.NavigationEvent) {
				if s.Protect {
					e.PreventDefault()
					s.Result = "Navigation prevented"
				}
			}},
			{ID: "disabled", Label: "Unavailable destination", Disabled: true},
		}},
		{ID: "help-label", Label: "Help section", LabelOnly: true},
		{ID: "help", Label: "Help center", Icon: lucide.Info(), Tooltip: nui.Set(true), Children: []nui.NavigationItem{{ID: "quickstart", Label: "Quickstart guide"}, {ID: "releases", Label: "Release notes"}}, Decoration: nui.Decoration{UI: nui.Styles{nui.ItemLabelSlot: {Weight: nui.Set(600)}}}},
	}
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "NavigationMenu · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
