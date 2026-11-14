<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# ContextMenu

Show actions for content on right-click.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `ContextMenu`
- Types: `MenuOptions`, `MenuItem`, `MenuContentOptions`
- Theme key: `nui.ComponentContextMenu` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep an independent MenuState for each trigger and stable, unique item IDs. Right-click the target or use the Menu key or Shift+F10. MenuOptions controls size, icons, navigation and filtering. Configure individual submenus through MenuItem.Content.

Default dividers in menus and submenus paint one device pixel; a larger Separator theme Size retains its DIP width.

## Things to know

Menus use non-modal native overlays; outside clicks dismiss and continue to their targets. Loop defaults to false. Shortcut displays keys without registering commands. Filtering uses case-insensitive substring matching and preserves item order; IgnoreFilter leaves results to the application. MenuState.Query stores an unbound root query, with separate drafts per submenu. These actions only update example state. Filtering and Radio items are native extensions; Nuxt ContextMenu has no built-in filter input. Root placement defaults to the right with a 2 DIP gap; Placement and Offset are optional native extensions. Pointer coordinates stay fixed in the window when the target moves.

## API

```go
func ContextMenu(c *ui.Context, t Theme, anchor ui.Element, state *MenuState, items []MenuItem, o MenuOptions) ui.Element
```

### MenuOptions

```go
type MenuOptions struct {
	Label              string
	Size               Size
	Disabled, Loop     bool
	CheckedIcon        *ui.SVG
	LoadingIcon        *ui.SVG
	Width, MaxHeight   Optional[float32]
	Placement          Optional[Placement]
	Offset             Optional[float32]
	Filter             bool
	FilterFields       []string // label (default), description, id, shortcut.
	IgnoreFilter       bool
	SearchTerm         *string
	FilterInput        *InputOptions
	OnSearchTermChange func(string)
	OnOpenChange       func(bool)
	Decoration
}
```

### MenuItem

```go
type MenuItem struct {
	ID, Label, Description, Shortcut string
	Kind                             MenuItemKind
	Icon                             *ui.SVG
	Avatar                           *AvatarOptions
	Color                            Color
	Disabled, Radio, Loading         bool
	Checked                          Optional[bool]
	Close                            Optional[bool]
	Children                         []MenuItem
	OnSelect                         func()
	OnCheckedChange                  func(bool)
	UI                               Styles
	Content                          *MenuContentOptions
	// Build supplies non-interactive visual content; Label is still required for semantics/typeahead.
	Build func(*ui.Context, Theme)
}
```

### MenuContentOptions

```go
// MenuContentOptions overrides one submenu. Size and icons inherit from the root.
type MenuContentOptions struct {
	Label                    string
	Width, MaxHeight, Offset Optional[float32]
	Placement                Optional[Placement]
	Loop                     Optional[bool]
	Filter                   bool
	FilterFields             []string
	IgnoreFilter             Optional[bool]
	SearchTerm               *string
	FilterInput              *InputOptions
	OnSearchTermChange       func(string)
	Decoration
}
```

### Related types

```go
// MenuState is window-owned data. Never share one instance between independent menu triggers.
// Open can be set by the application; all other fields preserve keyboard/hover identity.
type MenuState struct {
	Open  bool
	Query string // Root filter draft when MenuOptions.SearchTerm is nil.
}
```

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

```go
type MenuItemKind uint8

const (
	MenuAction MenuItemKind = iota
	MenuLabel
	MenuSeparator
)
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `InputOptions` → [input.md](input.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### DropdownMenu and ContextMenu

##### Create an action menu

DropdownMenu opens from a button. ContextMenu opens from a focusable area with a right-click, the Menu key, or Shift+F10.

```go
anchor := nui.Button(c, t, "Actions", nui.ButtonOptions{})
nui.DropdownMenu(c.Key("actions"), t, anchor, &state.Menu, []nui.MenuItem{
    {ID: "pin", Label: "Pin", Checked: nui.Set(state.Pinned),
        OnSelect: func() { state.Pinned = !state.Pinned }},
    {ID: "export", Label: "Export", Children: []nui.MenuItem{
        {ID: "csv", Label: "CSV", OnSelect: state.ExportCSV},
        {ID: "pdf", Label: "PDF", Disabled: true},
    }},
}, nui.MenuOptions{Label: "Document actions"})
```

Keep one MenuState per window and per trigger; your app can read and write Open. DropdownMenu gives the anchor menu-button and expanded semantics. ContextMenu stores the right-click position in window coordinates. Another right-click only repositions the menu and does not report a second open. If the target's layout moves, the menu stays at the original click position. The Menu key and Shift+F10 use the target's bottom-left corner as the anchor point, and opening from the app uses the current target area. On close, it returns a zero-value panel and clears the old pointer position.

##### Menu items

| MenuItem field | Behavior |
| --- | --- |
| ID / Label / Description / Icon | Stable identity, name, description, and icon. The ID is unique across the whole menu tree |
| Kind | MenuAction (default), MenuLabel, or MenuSeparator. Group labels and separators also need IDs |
| Shortcut | Shows a shortcut at the end of the row; does not register it |
| Avatar / Loading | Loading takes precedence, then Icon, then Avatar. The loading icon does not disable the item |
| Disabled | Prevents focus and activation |
| Color | Any color other than Neutral tints the text and icon, and the highlight background uses that color at 10% opacity |
| Checked / Radio | Checked is an Optional[bool]; an explicit false or true gives checkbox semantics. Radio is a native extension that must be used with Checked, and your app maintains mutual exclusion |
| OnCheckedChange / Close | When a leaf item with Checked is activated, the new value is reported once (Radio reports true). Close defaults to true; `nui.Set(false)` keeps the menu open |
| Children | A submenu that opens by mouse or keyboard; up to 32 levels deep |
| Content | MenuContentOptions for a single submenu. Can override size, position, looping, filtering, and parts; size and icons are inherited from the root menu |
| UI | Overrides this item's parts and takes precedence over MenuOptions.UI. Submenu panels use Content.UI |
| OnSelect | By default, a leaf item first closes the menu and requests focus back on the trigger, then reports the checked change and the selection. With Close=false, the menu stays open. Parent items only open their submenu |
| Build | Custom display content; Label is still required. Do not embed separate inputs or buttons in it |

MenuOptions.Size defaults to MD, can be inherited from the component theme, or can be set explicitly from XS to XL. Menus use the [option row style](select.md): at MD, the padding is 6 DIP and icons are 20 DIP. Descriptions use the same font size as labels, and icons and trailing content align with the first line. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. CheckedIcon defaults to Check and LoadingIcon to LoaderCircle. The submenu arrow follows RTL.

MenuOptions.Label sets the panel's name. Width follows the content by default, with a minimum of 128 DIP. MaxHeight defaults to 320 DIP, long menus scroll, and an explicit 0 hides the visible content. Sizes are limited to the window. The DropdownMenu root opens at BottomStart with an 8 DIP offset by default; ContextMenu opens at RightStart with a 2 DIP offset. Submenus open outward in the reading direction with a 3 DIP offset. Placement and Offset can both be overridden explicitly; for ContextMenu, this is a native extension. Menus stay clear of the window edges through AttachTo. There is no custom collision boundary or arrow option.

Shortcut is split on `+` into Kbd keys; in `Ctrl++`, the final `+` is the plus key. FormatKey converts key names to each platform's display form. XS and SM menus use SM Kbds, MD and LG use MD, and XL uses LG; the gaps between keys are 2/2/2/4/4 DIP. Go shows shortcuts even in narrow windows.

##### Filtering and controlled state

Filter defaults to false. When enabled, a native Input is shown and receives focus when the root menu opens. SearchTerm can bind a persistent string; when nil, MenuState.Query is used. FilterFields defaults to `label` only and also supports `description`, `id`, and `shortcut`. Matching is a case-insensitive substring match that keeps the original order. With IgnoreFilter=true, the list is shown as passed in and your app supplies results based on the query. Group labels and separators are not matched. When no action matches, a localized empty state is shown. Submenus that are not open are not searched.

FilterInput takes InputOptions for the name, placeholder, clear button, and appearance. OnSearchTermChange fires when the user edits or clears the query, but not for outside assignments or rebuilds without input. In the filter input, Escape first clears a non-empty query and then closes menus one level at a time. Enter activates the hovered item, or else the first available result. Up and Down move real focus to the results, after which typing performs prefix navigation in the menu. This differs from upstream, which keeps focus in the input using aria-activedescendant.

A submenu enables its own filtering with MenuItem.Content.Filter; it does not inherit the root's Filter. Each level keeps its own query by stable ID, or can bind Content.SearchTerm. Opening a submenu by hovering does not take focus away from the current filter input; entering a submenu by click or keyboard focuses its filter input. If the query hides the focused item, focus returns to that level's filter field.

Disabled closes the menu. DropdownMenu also disables its trigger, while ContextMenu keeps the target's own clicks, input, and focus. If the target itself is disabled, neither menu can open. OnOpenChange reports only root menu open/close changes caused by the user; changing MenuState.Open directly does not report. If the menu is closed from outside while focus is still inside it, focus returns to the trigger. Filtering corresponds to an upstream prop for DropdownMenu and is a native extension for ContextMenu.

##### Keyboard and pointer

| Action | Behavior |
| --- | --- |
| ↑↓ / Home / End | Move among available items. Loop defaults to false, which stops at the ends; true wraps around |
| → / ← | Open a submenu / return to the parent menu; swapped in RTL |
| Enter / Space | Activate the current item |
| Escape | Close the current level; closing the root level restores focus to the anchor |
| Tab / Shift+Tab | Close the menu and continue MyGo's normal focus order |

Typing characters performs a case-insensitive prefix search; keystrokes less than 1 second apart are combined, and repeating the first letter cycles through matches. This does not use IME candidates and is not a fuzzy search.

Moving the pointer onto an item changes the highlight, and hovering for 180 ms opens a submenu. Keyboard actions cancel a pending hover-open so that a submenu does not reopen right after Escape. Reordering items keeps focus by ID, and removing items clears their state. There is currently no pointer-path tolerance for diagonal movement and no long-press to open on touch screens.

Upstream ContextMenu's pressOpenDelay defaults to 700 ms for long presses on touch screens or pens. The current MyGo InputEvent does not distinguish mouse, touch, and pen, so Go does not add a delay field that would mistake a mouse long-press for touch. Your app can provide an explicit menu button or set MenuState.Open instead.

##### Combining with existing events

DropdownMenu adds an OnClick to the anchor, and the anchor's existing callback still runs; do not toggle the same MenuState.Open from that callback as well. ContextMenu takes over HandleInput and does not chain an existing handler of the same kind. Do not share one trigger state between different overlays.

Menus are non-modal: an outside click closes the menu and still reaches its target. When the window loses focus, no focus indicator is shown and focus-group arrow keys do not run; background rebuilds do not keep pulling focus back to the first item.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| DropdownMenu / ContextMenu | Panel, Header (group), Item, Body, Leading, Label, Description, Trailing (row-end container), Indicator, SeparatorSlot, Editor (filter root), EmptySlot. Build replaces the Body content |

### Options and defaults

#### Common options by component › Overlays and notifications (excerpt)

| Component | Common options and behavior |
| --- | --- |
| DropdownMenu / ContextMenu | Size defaults to MD and Loop to false. Filter and SearchTerm configure filtering; MenuItem.Content defines a submenu. Close=false keeps the menu open after a leaf item is selected. OnCheckedChange reports value changes and OnOpenChange reports open/close changes |

### Forms, dates, and data components

#### Listbox, SelectMenu, and async search › Listbox: a selectable list shown inline (excerpt)

The empty message and the loading indicator are polite status regions with stable identities, as are the empty messages of Select, InputMenu, SelectMenu, and filtered DropdownMenu and ContextMenu, and SelectMenu's loading and error rows. A status is not read when it first appears. If its text changes while it stays shown, for example from "No data" to "No matching data" as the user types, the new text is read once; rebuilding without a change reads nothing.

### Inputs, selection, and fields

#### Popover and ScrollArea › Popover: show content next to a control (excerpt)

By default, a generic Popover has the theme background, a 1 DIP border, a 6 DIP corner radius, an 8 DIP content gap, a shadow, and 16 DIP of padding. To make content reach the edges, set the padding of UI[Panel] to 0. Select, InputMenu, and menus use the same base style but usually with a 240 DIP maximum height, 4 DIP of padding, no gap, and at least the trigger's width. For the menus' own height options, see DropdownMenu and ContextMenu.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Menu                                          nui.MenuState
	Action, Query, SubQuery                       string
	Pinned, Reverse                               bool
	Actions, CheckedChanges, OpenChanges, Queries int
	TargetClicks                                  int
	TargetMoved                                   bool
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	icon := lucide.Pencil()

	var width nui.Optional[float32]

	items := []nui.MenuItem{
		{ID: "edit", Label: "Edit", Description: "Change project details", Icon: icon, Shortcut: "Ctrl+E", OnSelect: func() { s.Action = "Edit"; s.Actions++ }},
		{ID: "copy", Label: "Duplicate", Description: "Copy this project", Icon: lucide.Copy(), OnSelect: func() { s.Action = "Duplicate"; s.Actions++ }},
		{ID: "archive", Label: "Archive", Color: nui.Warning, OnSelect: func() { s.Action = "Archive"; s.Actions++ }},
		{ID: "separator", Kind: nui.MenuSeparator},
		{ID: "pin", Label: "Pinned", Checked: nui.Set(s.Pinned), Close: nui.Set(false), OnCheckedChange: func(v bool) { s.Pinned = v; s.CheckedChanges++ }, OnSelect: func() { s.Action = "Pinned"; s.Actions++ }},
		{ID: "export", Label: "Export", Children: []nui.MenuItem{{ID: "csv", Label: "CSV", OnSelect: func() { s.Action = "CSV"; s.Actions++ }}, {ID: "json", Label: "JSON", OnSelect: func() { s.Action = "JSON"; s.Actions++ }}}, Content: &nui.MenuContentOptions{SearchTerm: &s.SubQuery, FilterInput: &nui.InputOptions{Label: "Export search"}}},
		{ID: "reorder", Label: "Reverse actions", Close: nui.Set(false), OnSelect: func() { s.Reverse = !s.Reverse; s.Action = "Reordered" }},
		{ID: "move-target", Label: "Move target", Close: nui.Set(false), OnSelect: func() { s.TargetMoved = !s.TargetMoved }},
	}
	if s.Reverse {
		slices.Reverse(items)
	}
	left := float32(0)
	if s.TargetMoved {
		left = 32
	}
	anchor := nui.Button(c.Key("trigger"), t, "Right click here", nui.ButtonOptions{AccessibleLabel: "Right click here", Variant: nui.Outline}).Width(280).Height(64).AlignSelf(ui.Start).Margin(0, 0, 0, left).OnClick(func() { s.TargetClicks++ })
	nui.ContextMenu(c.Key("menu"), t, anchor, &s.Menu, items, nui.MenuOptions{Label: "Project actions", Width: width, Placement: nui.Set(nui.RightStart), FilterFields: []string{"label", "description"}, SearchTerm: &s.Query, FilterInput: &nui.InputOptions{Label: "Action search", Clearable: true}, OnSearchTermChange: func(string) { s.Queries++ }, OnOpenChange: func(bool) { s.OpenChanges++ }})
	ui.Text(c, s.Action)
	ui.Textf(c, "Actions: %d · Checked changes: %d · Open changes: %d", s.Actions, s.CheckedChanges, s.OpenChanges)
	ui.Textf(c, "Query: %s · Changes: %d · Pinned: %v", s.Query, s.Queries, s.Pinned)
	nui.Button(c.Key("open-at-target"), t, "Open at target", nui.ButtonOptions{Variant: nui.Soft}).AlignSelf(ui.Start).OnClick(func() { s.Menu.Open = true })
	ui.Textf(c, "Target clicks: %d · Moved: %v", s.TargetClicks, s.TargetMoved)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "ContextMenu · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
