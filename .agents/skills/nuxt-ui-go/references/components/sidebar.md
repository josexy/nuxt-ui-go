<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Sidebar

Add a collapsible, resizable application sidebar.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Sidebar`
- Types: `SidebarOptions`, `SidebarState`, `SidebarScope`
- Theme key: `nui.ComponentSidebar` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Sidebar provides a title, description, header actions, scrollable body and footer. Keep collapsed in application state and use SidebarState for width and the narrow-window menu. Actions appear while expanded or mobile; HeaderContent replaces the default title and description.

The resize highlight aligns to the same physical edge as the sidebar border, avoiding adjacent lines at fractional scales.

## Things to know

The component defaults to no breakpoint and no resizing; this example enables 160–360 DIP resizing and a 64-DIP compact width. Desktop collapsed and mobile Open are independent; native icon-width collapse differs from upstream offcanvas. External assignments are silent. Breakpoint remounts retain application models, but not widget-local scrolling or selections. Supply an external opener and ReturnFocus when hiding every internal control. Native floating/inset variants, a click-to-collapse rail and arbitrary menu option objects are not provided.

## API

```go
// Sidebar preserves the original collapsed-pointer API. Supply State to control
// the overlay externally or retain width across unmounts. Otherwise local widget
// state remembers it for the lifetime of this keyed Sidebar. Call every build,
// including while its overlay is closed, to finish exits and restore focus.
func Sidebar(c *ui.Context, t Theme, collapsed *bool, o SidebarOptions, content func(*ui.Context, Theme, bool)) ui.Element
```

### SidebarOptions

```go
type SidebarOptions struct {
	Title, ToggleLabel, ResizeLabel string
	Description                     string
	Width, CollapsedWidth           Optional[float32]
	MinWidth, MaxWidth              Optional[float32]
	// SideLeft (default) is the leading side; SideRight is trailing. Both mirror in RTL.
	Side Optional[Side]
	// Breakpoint enables the overlay below this window width in DIP. Unset/zero
	// keeps the original desktop-only behavior. Mode defaults to SidebarSlideover.
	Breakpoint          Optional[float32]
	Mode                SidebarMode
	State               *SidebarState
	Resizable, Disabled bool
	Collapsible, Toggle Optional[bool]
	// NavigationKey is an application-owned route/selection ID. Changes close the
	// overlay when AutoClose is true (the default), without coupling to a router.
	NavigationKey  string
	AutoClose      Optional[bool]
	Header, Footer func(*ui.Context, Theme)
	// HeaderContent replaces the default title; FooterContent replaces Footer.
	// ToggleContent replaces the desktop collapse / narrow-window open button.
	HeaderContent, FooterContent, ToggleContent func(*ui.Context, Theme, SidebarScope)
	// Actions adds header actions while expanded or in the narrow-window menu.
	Actions           func(*ui.Context, Theme, SidebarScope)
	OnResize          func(float32)
	OnCollapsedChange func(bool)
	OnOpenChange      func(bool)
	// ReturnFocus overrides the default built-in toggle when using an external opener.
	ReturnFocus *ui.Handle
	Decoration
}
```

### SidebarState

```go
// SidebarState belongs to one sidebar in one window. Width is the remembered
// expanded width in DIP; unset uses SidebarOptions.Width. Applications may save
// Width and their collapsed value and restore them before building the sidebar.
// Open controls the narrow-window overlay independently of desktop collapse.
type SidebarState struct {
	Width Optional[float32]
	Open  bool
}
```

### SidebarScope

```go
// SidebarScope is supplied synchronously to custom regions. Its actions may be
// used in input callbacks; never retain Context or Elements from a builder.
type SidebarScope struct {
	Collapsed, Open, Mobile bool
	Collapse                func(bool)
	SetOpen                 func(bool)
}
```

### Related types

```go
type Side uint8

const (
	SideRight Side = iota
	SideLeft
	SideTop
	SideBottom
)
```

```go
type SidebarMode uint8

const (
	SidebarSlideover SidebarMode = iota
	SidebarModal
	SidebarDrawer
)
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Mode` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

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

##### Sidebar state and size

By default, Sidebar is 256 DIP wide when expanded and 64 DIP wide when collapsed. `collapsed *bool` controls the desktop collapsed state, and Options.State can take a SidebarState:

| State | Purpose |
| --- | --- |
| Width (Optional[float32]) | Remembers the expanded width. When unset, Options.Width is used; it is updated when the user resizes |
| Open (bool) | Whether the overlay is open in narrow windows; independent of collapsed |
| No State passed | Uses local state under the current stable Key. Pass State explicitly to keep it across unmounts or to save it to a file |

Collapsing does not change the remembered expanded width. Double-clicking the handle restores Options.Width, and your app can also reset State.Width to unset. Your app handles persistence; the component never reads or writes the disk during a build.

`Resizable: true` enables resizing, with a default minimum and maximum width of 160 and 480 DIP, which you can change with MinWidth and MaxWidth. Dragging below CollapsedWidth + 4 DIP collapses the sidebar, and dragging outward expands it again.

| Handle action | Behavior |
| --- | --- |
| Left / Right | Moves the edge 10 DIP in that screen direction |
| Shift+Left / Right | Moves 50 DIP |
| Home / End | Collapsed or minimum width / maximum width |
| Enter / double-click | Toggles collapsed / restores the default width |

These actions work for sidebars on either side and in RTL. Sizes must be finite and non-negative, and MaxWidth must be at least MinWidth. When resizing is enabled, the collapsed width cannot exceed the minimum expanded width.

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

#### Splitter, HeaderLayout, FooterLayout, and FooterColumns › HeaderLayout and FooterLayout (excerpt)

HeaderLayout does not create a home link, web sticky positioning, a breakpoint menu, or a route listener. Header's Body part refers to the title text container, which differs from the web Header's menu body slot.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| Sidebar | `*bool` collapsed, optional SidebarState | Root, Header, Label, Description, Actions, Trigger, Body, Footer, SeparatorSlot. The overlay also has Panel, Backdrop, Close |

#### Examples (excerpt)

It is 256 DIP wide by default, can be dragged between 200 and 360 DIP, and collapses to a 64 DIP icon bar. Switching pages keeps the width and collapsed state, which are stored per window; this sidebar is independent of the demo sidebar on the Application layout page.

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| Sidebar | Resizable=false; Collapsible, Toggle, and AutoClose=true. Expanded and collapsed widths are 256 and 64 DIP. Description is empty. Actions appear only when expanded or in the mobile menu |

#### Common options by component › Sidebar, navigation, and file selection

| Component | Common options and behavior |
| --- | --- |
| Sidebar | State stores the expanded width and the open state in narrow windows. MinWidth and MaxWidth default to 160 and 480 DIP. When Breakpoint is unset or 0, the desktop layout is always used |
| NavigationMenu | Vertical Compact mode adds tooltips and child popovers by default; turn either off with `nui.Set(false)`. PopupPlacement changes where child popups appear |
| Breadcrumb | Color and SeparatorIcon configure appearance. Each item's Avatar, Active, and UI can be set individually. The last item is also actionable if it has OnSelect |
| Pagination | FirstIcon, PrevIcon, NextIcon, LastIcon, and EllipsisIcon override the default icons. ActiveColor and ActiveVariant style the current page |
| Stepper | StepItem.Content builds the current step's body and UI overrides a single step. Disabled disables only the step headers. List and Panel configure the list and the body |
| Timeline | CurrentID takes precedence over Current. Reverse does not reorder the data. DateSlot styles Time. OnSelect does not change progress by itself |
| Banner | CloseButton controls visibility; Close and CloseIcon configure the close button. OnSelect is independent of the inner actions. OnClose fires only when the user closes the banner |
| ErrorView | Clear defaults to true, but the default button also requires OnRetry. Retry configures the button, Actions always takes precedence, and StatusSlot styles the status code |
| FileUpload | Variant is UploadArea or UploadButton. Layout defaults to UploadGrid. Position affects only the list inside the area. ShowIcon, FileImage, FileDelete, and Preview default to true. Supports icon and delete button configuration and Highlight |

Clicking an interactive UploadArea, or pressing Enter or Space on it, opens the file picker; its accessible name comes from Label first. In the built-in list, file actions are separate from the title/icon entry point so native assistive technology can reach each action. The Choose button appears only with UploadButton or `Interactive: nui.Set(false)`. Dropzone=false removes the drop target. Preview=false hides the whole list, and FileImage=false hides only the app-supplied preview accessories; the Files data is kept in both cases. OnChange fires when a selection or removal changes the queue, and OnReject reports newly rejected files. After a removal, focus returns to the selection entry point. For full defaults and responsibilities, see [FileUpload](file-upload.md). For Sidebar's region builders, sizes, and keyboard controls, see Sidebar state and size.

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
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type example struct {
	Collapsed             bool
	Sidebar               nui.SidebarState
	Navigation            nui.NavigationState
	Active, Result, Draft string
	Actions               int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var headerActions func(*ui.Context, nui.Theme, nui.SidebarScope)
	var styles nui.Styles

	headerActions = func(c *ui.Context, t nui.Theme, scope nui.SidebarScope) {
		nui.Button(c, t, "", nui.ButtonOptions{Leading: lucide.RefreshCw(), AccessibleLabel: "Sidebar header action", Variant: nui.Ghost, Color: nui.Neutral, Square: true}).OnClick(func() { s.Actions++; s.Result = fmt.Sprintf("Header actions: %d", s.Actions) })
	}

	options := nui.SidebarOptions{
		Title: "Workspace", Description: "Project navigation", Actions: headerActions, Decoration: nui.Decoration{UI: styles}, State: &s.Sidebar, Resizable: true, Collapsible: nui.Set(true),
		Width: nui.Set(float32(256)), MinWidth: nui.Set(float32(160)), MaxWidth: nui.Set(float32(360)),
		Side: nui.Set(nui.SideLeft), Breakpoint: nui.Set(float32(0)), Mode: nui.SidebarSlideover, NavigationKey: s.Active,
		FooterContent: func(c *ui.Context, t nui.Theme, scope nui.SidebarScope) {
			label := "Account"
			if scope.Collapsed {
				label = ""
			}
			nui.Button(c.Key("account"), t, label, nui.ButtonOptions{Leading: lucide.User(), AccessibleLabel: "Account", Variant: nui.Ghost, Color: nui.Neutral, Block: true}).OnClick(func() { s.Active = "account"; s.Result = "Account selected" })
		},
		OnResize: func(width float32) { s.Result = fmt.Sprintf("Resized to %.0f DIP", width) },
	}
	right := options.Side.Set && options.Side.Value == nui.SideRight
	workspace := func() {
		ui.Column(c.Key("workspace")).Grow(1).MinWidth(0).Padding(20).Gap(12).Children(func() {
			ui.Text(c, "Workspace overview").FontSize(20).FontWeight(600)
			ui.Textf(c, "Current page: %s", s.Active)
			ui.Textf(c, "Collapsed: %t", s.Collapsed).TextColor(t.Muted)
			ui.Text(c, s.Result).TextColor(t.Muted)
			nui.Button(c.Key("reset-width"), t, "Reset width", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { s.Sidebar.Width = nui.Optional[float32]{}; s.Result = "Configured width restored" })
		})
	}
	direction := ui.LeftToRight

	ui.Row(c.Key("layout")).Direction(direction).FillWidth().Height(360).MinWidth(0).MinHeight(0).Border(1, t.Border).Radius(t.Radius).Children(func() {
		if right {
			workspace()
		}
		nui.Sidebar(c.Key("sidebar"), t, &s.Collapsed, options, func(c *ui.Context, t nui.Theme, collapsed bool) {
			if !collapsed {
				nui.Input(c.Key("draft"), t, &s.Draft, nui.InputOptions{Label: "Sidebar draft"})
			}
			placement := nui.RightStart
			if right != (c.Direction() == ui.RightToLeft) {
				placement = nui.LeftStart
			}
			nui.NavigationMenu(c.Key("navigation"), t, &s.Active, &s.Navigation, []nui.NavigationItem{
				{ID: "overview", Label: "Overview", Icon: lucide.LayoutDashboard()},
				{ID: "projects", Label: "Projects", Icon: lucide.Folder()},
				{ID: "settings", Label: "Settings", Icon: lucide.Settings(), Children: []nui.NavigationItem{{ID: "members", Label: "Members"}, {ID: "permissions", Label: "Permissions"}}},
			}, nui.NavigationOptions{Vertical: true, Compact: collapsed, PopupPlacement: nui.Set(placement)})
		})
		if !right {
			workspace()
		}
	})
}

func main() {
	s := &example{}

	s.Active = "overview"
	s.Result = "Drag the sidebar edge to resize."
	s.Draft = "Persistent sidebar draft"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Sidebar · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
