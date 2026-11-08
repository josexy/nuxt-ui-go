# Overlays, menus, and navigation

Dialogs and notifications support the task at hand; tabs, pagination, and steppers organize content. Choose a component by what the interaction is for:

| What you need | Component |
| --- | --- |
| Have the user finish a task before returning to the page | [Modal / Slideover](#modal-and-slideover) |
| A short hint, or extra content next to a control | [Tooltip](#tooltip), [Popover](inputs.md#popover-and-scrollarea) |
| A dropdown of actions, a context menu, or a notification about a result | [DropdownMenu / ContextMenu](#dropdownmenu-and-contextmenu), [Toast](#toast) |
| Switch between content, or expand details | [Tabs / Collapsible / Accordion](#tabs-collapsible-and-accordion) |
| A page path, pagination, a process, a timeline, or an error display | [Navigation components](#breadcrumb-pagination-stepper-and-timeline), [Banner / ErrorView](#banner-and-errorview) |

## Shared conventions

### Open state and stable IDs

Your app stores whether an overlay is open, the current selection, and other persistent state. Call `c.Key(...)` before creating each instance. The Elements and parts produced by a build belong to that build only; do not store them in the model or in background tasks. A zero-value Element returned on close means there is currently no panel.

Keep calling overlay components from the view function even when they are closed. This lets them finish exit animations, clean up, and restore focus. Do not wrap the entire component call in `if open`.

List item IDs must be non-empty and unique, or the component panics. Tabs, Accordion, Breadcrumb, Stepper, and Timeline require uniqueness within their own list; menus require it across the whole menu tree. Display names can repeat. Required state pointers must not be nil, except for Banner, which explicitly accepts nil to mean "always shown".

### Styles and callbacks

Most components return a `ui.Element`. Toast returns the notification ID, and CloseToast returns nothing. Style configures parts synchronously during the current build; use click or selection callbacks for business actions.

Styles are applied in this order: built-in style, component configuration in the theme, instance UI, and finally Style. Each component supports a different set of Color, Size, Variant, and parts; see the [part index](#part-index) at the end. Sizes are in DIP, so you do not need to multiply them by the display scale.

## Modal and Slideover

### Open an edit dialog

Modal appears in the center, and Slideover slides in from an edge of the window. Both are modal: the background cannot be used while they are open.

```go
// state stores EditOpen, Draft, and EditTrigger (a ui.Handle) per window.
edit := nui.Button(c, t, "Edit", nui.ButtonOptions{}).
    Bind(&state.EditTrigger).
    OnClick(func() { state.EditOpen = true })
nui.Tooltip(c, t, edit, "Edit the current profile", nui.TooltipOptions{})

// Call this on every build, including when EditOpen is false.
nui.ModalWithOptions(c.Key("edit"), t, &state.EditOpen, nui.ModalOptions{
    Title: "Edit profile",
    Description: "Changes apply only after you save.",
    ReturnFocus: &state.EditTrigger,
    Footer: func(c *ui.Context, inner nui.Theme) {
        nui.Button(c, inner, "Done", nui.ButtonOptions{}).
            OnClick(func() { state.EditOpen = false })
    },
}, func(inner nui.Theme) {
    nui.Input(c, inner, &state.Draft, nui.InputOptions{Label: "Name"})
})
```

This example only shows opening, editing a draft, and closing. To actually save the data, handle it in the app's Done action. For a simple dialog, you can also use `Modal(c, t, &open, title, body)`; use ModalWithOptions for the full set of options.

### Closing and focus

| ModalOptions | Behavior |
| --- | --- |
| Dismissible | Optional[bool], default true. false prevents closing by clicking the overlay or pressing Escape |
| CloseButton / CloseLabel | The close button is shown by default; you can hide it or rename it separately |
| Close / CloseIcon | Configure the close button's color, size, loading state, and icon. Defaults to MD, Neutral, Ghost, and X. Close.Leading takes precedence over CloseIcon |
| OnClose | Called once when the user closes the dialog with the close button, an allowed overlay click, or Escape. Not called when the app sets open=false directly |
| OnClosePrevent | Called once when Dismissible=false blocks Escape or an outside click. Dismissible does not affect the close button |
| ReturnFocus | The Handle that receives focus after the dialog closes |

Dismissible and CloseButton are independent. To prevent the user from closing the dialog, set both to `nui.Set(false)`. When a menu item opens a dialog, ReturnFocus should usually point to the menu's trigger button, which still exists after the menu closes; the menu item itself unmounts with the menu.

Tab focus is trapped inside the modal content, and an inner Select or Popover handles Escape first. By default, the first focusable control receives focus, usually the close button. To focus a field first, request that field's Handle in the action that opens the dialog. If the Header or Footer is very tall, your app must constrain the content size to leave room for the body.

### Content, size, and appearance

| Option | Behavior |
| --- | --- |
| Title / Description | The dialog's name and description, also shown in the default header |
| Header / Footer | `func(*ui.Context, Theme)`. Header replaces the title and description content, while the close button is still created separately. Footer customizes the action area |
| Width / MaxHeight | Optional[float32]. The default width is 512 DIP, limited to the window width minus 32 and the window height minus 64 |
| Overlay | Default true. false only hides the overlay color; focus is still trapped |
| Fullscreen | Uses the full window size without rounded corners or a border. Modal only |
| Transition | Optional[bool], default true. false skips the entrance animation; switching back to true does not replay an entrance that already finished |

The body scrolls while the title and Footer stay in place. The header has 16 DIP of vertical padding, 24 DIP of horizontal padding, and a minimum height of 64 DIP. The body has 24 DIP of padding, and the Footer has a 6 DIP gap. If Title, Description, and Header are all empty and CloseButton=false, no header is created. If body is nil, no body region is created. With only a Footer, there is no stray separator.

The Panel's text styles are inherited directly by ordinary body elements. Panel.Gap sets the spacing between the regions that exist and defaults to 0. Inner content is clipped to the panel's rounded corners. The header, body, and Footer have stable identities, so removing the header or toggling Fullscreen keeps focus and selection in a body editor. Closing unmounts the content. Values your app stores persistently remain, but temporary control state is not guaranteed to survive reopening.

The Footer is right-aligned by default. PartStyle cannot set alignment; for a different arrangement, create the layout yourself inside the Footer callback.

The Modal panel has an 8 DIP corner radius, and the overlay uses the Elevated color at 75% opacity. On open, the panel fades in from 95% scale over 200 ms while the overlay fades in with it. With Transition=false or reduced motion, it appears immediately. Like [Popover](inputs.md#popover-and-scrollarea), Modal closes immediately without an exit animation. There is no upstream full-overlay scrolling mode, non-modal switch, or keep-mounted option.

### Slideover: open from an edge

Slideover reuses Modal's title, body, Footer, and close options. Side and Extent determine its position and size:

| Side | What Extent means, and its default |
| --- | --- |
| SideRight (default) / SideLeft | Panel width, 448 DIP by default |
| SideTop / SideBottom | Panel height. When unset, it fits the content, up to the full window |

An explicit Extent is limited to the window's size along that axis. The Width, MaxHeight, and Fullscreen fields embedded from ModalOptions are not used by Slideover. The body stretches and scrolls, and the Footer sits at the bottom of the panel.

Inset defaults to false. When enabled, it leaves a 16 DIP margin on all sides of the window and uses an 8 DIP corner radius. The margin scales with the theme's spacing, and Extent is limited to the space left after the margin. Side is a physical edge and is not swapped in RTL. Left and right panels default to 448 DIP wide; Go does not use upstream's breakpoint rule that makes the panel full width on small screens.

Slideover always creates a body region; even with content=nil, it keeps the default 24 DIP of padding. For top and bottom panels sized to their content, this affects the panel height. This differs from Modal, which creates no body region when the body is nil. Body, Header, and Footer have stable identities, so changing the side or Inset keeps focus and selection in an editor.

Opening and closing use a fixed 200 ms cubic ease-out slide and overlay transition. With Transition=false, Theme.Motion.Reduce, or the system's reduced motion setting, the change is immediate. During the exit, the modal scope is kept and the content cannot be used; afterward, the panel unmounts and focus is restored. If the direction reverses midway, the transition continues from its current progress.

A Background override on the Backdrop, from the theme or the instance, also fades in and out, and the final Style still has the highest precedence. The native Backdrop contains the panel, so setting its Opacity affects the panel too. To hide only the overlay, use Overlay=false or a transparent background. For which inherited fields apply, see [Overlays and notifications](options.md#overlays-and-notifications).

Slideover cannot be stretched from its edge. For dragging by a bottom handle with multiple snap heights, use [Drawer](advanced.md#drawer).

## Tooltip

Tooltip shows a short, non-interactive hint: `Tooltip(c, t, anchor, text, options)`. It responds to mouse hover and keyboard focus, and appends its text and the formatted Kbds to the anchor's Description, keeping any existing description. Kbds only display shortcut hints; they do not register shortcut actions.

| Behavior | Default |
| --- | --- |
| Position and offset | Below the anchor, centered, 8 DIP away |
| MaxWidth | 320 DIP; text longer than one line is truncated with an ellipsis |
| Appearance | 24 DIP tall, 4 DIP corner radius, 10 DIP horizontal padding, small shadow |
| Mouse | 600 ms delay when DelayDuration is unset. When moving between tooltips, a 400 ms window lets the next one appear quickly |
| Keyboard | Shown immediately on focus |

`DelayDuration` is an `Optional[time.Duration]`, for example `nui.Set(200 * time.Millisecond)`. An explicit 0 or a negative value shows the tooltip immediately. Leaving the anchor cancels the wait, and keyboard focus does not wait for the hover delay. Nuxt UI defaults to a 700 ms delay and a 300 ms skip window; Go keeps MyGo's native defaults.

The tooltip hides when the pointer leaves, the trigger is clicked, the user presses outside, or Escape is pressed. It also closes when it is Disabled, when the anchor is invalid or disabled, or when both the text and Kbds are empty. It opens with a 100 ms scale-and-fade animation, appears immediately with reduced motion, and disappears immediately on close. A Go Tooltip does not receive pointer input or focus, and it has no upstream hoverable content, arrow, or arbitrary content slot. For multiline or clickable content, use Popover.

Kbds use the SM size by default. The text and the shortcut are separated by `·`, and keys are 2 DIP apart. Unlike upstream, which hides shortcuts on narrow screens, Go always shows them. Panel.Gap sets the space between the text and the shortcut container, and Trailing styles the container of the shortcut and separator dot. The text is truncated to one line, but the shortcut does not shrink, so leave room for it in MaxWidth. For controlled Open and position options, see [Overlays and notifications](options.md#overlays-and-notifications).

## DropdownMenu and ContextMenu

### Create an action menu

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

### Menu items

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

MenuOptions.Size defaults to MD, can be inherited from the component theme, or can be set explicitly from XS to XL. Menus use the [option row style](inputs.md#option-row-appearance): at MD, the padding is 6 DIP and icons are 20 DIP. Descriptions use the same font size as labels, and icons and trailing content align with the first line. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. CheckedIcon defaults to Check and LoadingIcon to LoaderCircle. The submenu arrow follows RTL.

MenuOptions.Label sets the panel's name. Width follows the content by default, with a minimum of 128 DIP. MaxHeight defaults to 320 DIP, long menus scroll, and an explicit 0 hides the visible content. Sizes are limited to the window. The DropdownMenu root opens at BottomStart with an 8 DIP offset by default; ContextMenu opens at RightStart with a 2 DIP offset. Submenus open outward in the reading direction with a 3 DIP offset. Placement and Offset can both be overridden explicitly; for ContextMenu, this is a native extension. Menus stay clear of the window edges through AttachTo. There is no custom collision boundary or arrow option.

Shortcut is split on `+` into Kbd keys; in `Ctrl++`, the final `+` is the plus key. FormatKey converts key names to each platform's display form. XS and SM menus use SM Kbds, MD and LG use MD, and XL uses LG; the gaps between keys are 2/2/2/4/4 DIP. Go shows shortcuts even in narrow windows.

### Filtering and controlled state

Filter defaults to false. When enabled, a native Input is shown and receives focus when the root menu opens. SearchTerm can bind a persistent string; when nil, MenuState.Query is used. FilterFields defaults to `label` only and also supports `description`, `id`, and `shortcut`. Matching is a case-insensitive substring match that keeps the original order. With IgnoreFilter=true, the list is shown as passed in and your app supplies results based on the query. Group labels and separators are not matched. When no action matches, a localized empty state is shown. Submenus that are not open are not searched.

FilterInput takes InputOptions for the name, placeholder, clear button, and appearance. OnSearchTermChange fires when the user edits or clears the query, but not for outside assignments or rebuilds without input. In the filter input, Escape first clears a non-empty query and then closes menus one level at a time. Enter activates the hovered item, or else the first available result. Up and Down move real focus to the results, after which typing performs prefix navigation in the menu. This differs from upstream, which keeps focus in the input using aria-activedescendant.

A submenu enables its own filtering with MenuItem.Content.Filter; it does not inherit the root's Filter. Each level keeps its own query by stable ID, or can bind Content.SearchTerm. Opening a submenu by hovering does not take focus away from the current filter input; entering a submenu by click or keyboard focuses its filter input. If the query hides the focused item, focus returns to that level's filter field.

Disabled closes the menu. DropdownMenu also disables its trigger, while ContextMenu keeps the target's own clicks, input, and focus. If the target itself is disabled, neither menu can open. OnOpenChange reports only root menu open/close changes caused by the user; changing MenuState.Open directly does not report. If the menu is closed from outside while focus is still inside it, focus returns to the trigger. Filtering corresponds to an upstream prop for DropdownMenu and is a native extension for ContextMenu.

### Keyboard and pointer

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

### Combining with existing events

DropdownMenu adds an OnClick to the anchor, and the anchor's existing callback still runs; do not toggle the same MenuState.Open from that callback as well. ContextMenu takes over HandleInput and does not chain an existing handler of the same kind. Do not share one trigger state between different overlays.

Menus are non-modal: an outside click closes the menu and still reaches its target. When the window loses focus, no focus indicator is shown and focus-group arrow keys do not run; background rebuilds do not keep pulling focus back to the first item.

## Toast

### Add a notification area and post notifications

Build one ToastViewport per window, and call Toast from user actions:

```go
t := nui.App(c, config)
// Put this before other overlays so modal overlays can cover notifications.
nui.ToastViewport(c, t, nui.ToastViewportOptions{})
nui.Button(c, t, "Save", nui.ButtonOptions{}).OnClick(func() {
    previous := state.Value
    state.Value = state.Draft
    nui.Toast(c, t, nui.ToastOptions{
        ID: "saved",
        Title: "Saved",
        Color: nui.Success,
        Action: "Undo",
        OnAction: func() { state.Value = previous },
    })
})
```

Do not call Toast unconditionally during the view build, or a notification may be posted on every rebuild. Write background results back to the model through Window.Update, then post the notification from a window UI action or an explicitly controlled event handler.

### IDs, duration, and actions

| Option or action | Behavior |
| --- | --- |
| ID | Generated automatically when empty. Posting with an existing ID replaces that toast and restarts its timer, without calling the old toast's OnClose |
| Timeout | 5 seconds by default; a negative value makes the toast persistent |
| Action / OnAction | A single action. Running it closes the toast and calls OnClose |
| Actions | Several ToastActionOptions. Each closes the toast by default; Close=false on an action keeps it open |
| ToastOptions.Close | `nui.Set(false)` hides the close button without changing the timeout or the rules for closing on action |
| Avatar / Icon | The avatar takes precedence. The default avatar diameter is 44 DIP and the icon is 20 DIP. An explicit avatar size uses Avatar's own size |
| CloseOptions / CloseIcon | Optional button options and icon. Defaults to MD / Neutral / Link, no padding, and X. CloseOptions.Leading takes precedence |
| ProgressOptions | ToastProgressOptions sets Color and the Progress UI. Defaults to size SM, with the color inherited from the toast |
| OnClose / CloseToast | The close button, the timeout, and CloseToast all call OnClose. `CloseToast(c, "")` closes all toasts |

You can use the single action and Actions together. Each button in Actions can set ID, Label, OnAction, ButtonOptions, and Close; buttons default to XS and Solid, and inherit the toast's color. Provide stable IDs when actions are updated or reordered; non-empty IDs must be unique within a toast. Without IDs, actions are identified only by their position in a static list. Changing the orientation keeps focus on the current action button. Without an Icon or Avatar, no status icon is added automatically.

Vertical is the default orientation: the actions sit below the body with a 10 DIP gap. Horizontal centers the content and actions vertically side by side. An empty title leaves no blank line. ToastOptions.UI.Actions styles the action container, such as its spacing; style individual buttons with ButtonOptions.UI on each Actions item. Code that used UI.Actions to style the single compatibility Action button should switch to an Actions list with button UI.

The queue stores OnAction, OnClose, and the ButtonOptions.Style of each action, so their closures should capture only the business model, data, or services. Do not capture Elements, Parts, or Contexts produced by a build; the elements Style receives are valid only during that build.

### Position, pausing, and style

ToastViewport defaults to the bottom right and a width of 384 DIP. It also supports bottom left, bottom center, top right, top left, and top center. These are physical positions and are not swapped in RTL. Top positions put the newest toast at the top, and bottom positions put it at the bottom. The width is further limited by the viewport's actual padding; an explicit PaddingX=0 removes that extra limit.

Limit accepts 1–3; 0 or an out-of-range value is treated as 3. MyGo shows at most the 3 newest toasts. Older queued toasts keep counting down and may reappear after newer ones close; Limit is not the queue capacity. Nuxt UI's max defaults to 5 and discards older toasts beyond the limit, so the rules differ.

The queue timer pauses while the pointer is over the notification area, or while focus is inside a toast and the window has focus. It keeps running while the window is in the background or hidden, even with focus left inside a toast, so a toast whose time runs out there closes before the window shows again. The default countdown bar pauses with it; hide it with `Progress: nui.Set(false)`. Persistent toasts have no countdown. With reduced motion, the remaining time is still updated, but it is shown in steps at least 250 ms apart without interpolation. Toasts enter and exit with a 150 ms, 8 DIP fade and slide toward the matching screen edge; with reduced motion, they switch immediately. An empty queue draws no panel.

Toasts use RoleStatus, and the action and close buttons are keyboard accessible. Actual screen reader announcements have not been verified.

Each toast stores a copy of the visual Theme and options at the time it was created, without the source form's disabled state or field-group information. If the global theme changes while a toast is visible, existing toasts keep their original theme; to update a toast, post it again with the same ID. Automatically generated close names still follow the current language, while explicit app text is updated by your app. Viewport.Decoration affects only Root; style each toast through ToastOptions.UI or the ComponentToast configuration at creation time.

Updating with the same ID replaces the toast entirely: fields you do not pass again return to their defaults, the old OnClose is not called, and the timer restarts. There is no upstream duplicate-pulse effect, collapsed stacking, swipe to dismiss, arbitrary content slot, or foreground/background announcement policy. The website examples and copyable programs all create a ToastViewport explicitly so they do not fall back to MyGo's default notification look.

## Tabs, Collapsible, and Accordion

### Choose a component and its state

| Component | Bound state | Default behavior |
| --- | --- | --- |
| Tabs | `*string`, the current item ID | Pill style; only the current content is shown. Underline and Vertical are available |
| Collapsible | `*bool`, whether it is expanded | Toggles on click or Enter/Space; expanding and collapsing take 200 ms by default |
| Accordion | `*[]string`, the IDs of expanded items | Only one item can be open at a time by default, and it can be closed again. Multiple allows several |

TabItem provides ID, Label, Icon, Avatar, Badge, BadgeOptions, Disabled, Content, and an item UI. Icon takes precedence over Avatar. The default avatar diameter is 16/16/20/20/24 DIP by size; an explicit Avatar.Size takes precedence. Badge is a string, and `"0"` is shown. It defaults to SM / Neutral / Outline and can be overridden with BadgeOptions. The body callbacks of Tabs and Accordion receive a Context and a Theme; Collapsible's body receives a Theme and uses the outer `c`.

Tabs use Left/Right or Up/Down arrows depending on the orientation, wrapping around and skipping disabled items. In horizontal tabs, Left and Right swap in RTL, and keys for the other axis do not change the selection. If the current item is removed or disabled, the selection moves to the first available item. If no item is available, the business ID is kept but no content is shown. If the content that has focus is switched, hidden, or removed, focus returns to a valid tab. The content is a focusable native group, so Tab first enters the panel and then its inner controls; MyGo has no separate tabpanel role yet. In Accordion, ↑↓, Home, and End move between headers, and Tab keeps the normal order of headers and content.

Tabs.OnChange receives the new ID, Collapsible.OnChange receives a bool, and Accordion.OnChange has no parameters; your app reads the ID list. The user-change callbacks do not fire when the app changes the value directly, when Tabs corrects an invalid ID, or when Accordion's single mode trims extra IDs.

Accordion keeps IDs in the app's open list that are temporarily missing, so the item opens again if the same ID returns. Your app removes IDs that are gone for good. For manual Tabs activation and Accordion's Collapsible option, see [Sliders, tabs, and accordions](options.md#sliders-tabs-and-accordions).

Collapsible.Disabled, Accordion.Disabled, and AccordionItem.Disabled affect only the headers; expanded content keeps the disabled settings of its own controls. If content being edited is closed from outside, focus returns to an available header. If Accordion removes the focused item, focus returns to the first available header. Content that is animating closed stops receiving input immediately, so it cannot trigger actions while collapsing.

### Keeping hidden content mounted

By default, only the active Tabs content is built, and collapsed content unmounts after its exit animation. The app model is kept, but editor selections, undo history, and local UI state may be lost on unmount.

KeepMounted=true keeps all content in every build and hides inactive content with Hidden, so MyGo's local state is preserved. Hidden items take no part in layout or focus, but their content callbacks still run. In keep-mounted collapsible content, the body hides immediately while the arrow still follows the native animation. Removing an entire item still unmounts it.

When KeepMounted changes while a Collapsible or Accordion is open, the content keeps the same identity and the editor's local state is not rebuilt. Theme.Motion.Reduce or the system's reduced motion setting completes height and icon changes immediately. The default animation uses MyGo's 200 ms interpolation, which is not frame-for-frame identical to upstream's CSS animation.

TabsOptions.Content defaults to true. An explicit false builds no panels, which unmounts local UI state even with KeepMounted=true; drafts your app stores are unaffected. When re-enabled, the panels are built from app data. Do not start I/O or submit anything unconditionally from content callbacks.

### Appearance

In the pill style, the active Tabs item uses the semantic color as its background with inverted text; Neutral uses InvertedBackground. A horizontal list fills the width. Underline corresponds to Nuxt UI's link style: the list has a 1 DIP baseline and an active indicator line, at the bottom when horizontal and on the start side when vertical, following RTL. Vertical tabs have equal widths, and the list and content are center-aligned. For XS–XL, the trigger gap is 4/6/6/8/8 DIP and the text line height is 16/16/20/20/24 DIP.

Indicator styles the active background or indicator line, Trailing styles the badge, and TabItem.UI overrides that item's parts. Existing Trigger.Background overrides still work; Indicator.Background is the more specific setting for the active background. The indicator currently updates directly with the selection, without upstream's 200 ms slide between tabs. There are no arbitrary leading/default/trailing or list-leading/list-trailing slots; place other actions in a toolbar outside Tabs.

In the pill style, Indicator and the active Trigger share one native element so the background stays behind the text and avatar. In the underline style, Indicator is a separate line. Configure the Style callback for these actual parts; do not assume every style entry point is a separate node.

Accordion and Collapsible triggers have no hover background and no horizontal padding. They have 14 DIP of vertical padding, a 20 DIP icon, and a 6 DIP gap. Accordion content uses the default text color with a 14 DIP bottom margin.

TrailingIcon defaults to ChevronDown. AccordionItem.TrailingIcon takes precedence over AccordionOptions.TrailingIcon. AccordionItem.UI can override that item's Item, Header, Trigger, Leading, Label, Trailing, and Body. Empty Content creates no body and no bottom margin. The body layout container and the animation clipping container do not have separate style entry points.

Nuxt UI's Collapsible uses a trigger slot supplied by the app. Go's Label, Icon, TrailingIcon, default header, and body spacing are native extensions. For an arbitrary trigger composition, use MyGo's Base APIs.

## Breadcrumb, Pagination, Stepper, and Timeline

### Breadcrumb: show the current location

BreadcrumbItem contains ID, Label, Icon, Avatar, Active, Disabled, OnSelect, and UI. The options provide Label, Color, SeparatorIcon, and MaxItems. Items with OnSelect are actionable links, including the last one: the pointer and Enter activate them, while Space, as on the web, does not. Items with a nil callback are static text. Your app provides the path and the navigation actions.

When Active is unset, only the last item is active; an explicit false turns off the active style. Color defaults to Primary, and with Neutral the active text uses Highlighted. Inactive items are 14 DIP Muted text without padding, and actionable items change to the Text color on hover. The separator arrow is 20 DIP, and items are 6 DIP apart. In RTL, the default arrow is reversed, while a custom SeparatorIcon is kept as is. Icon takes precedence over Avatar, and the default avatar diameter is 20 DIP. The path stays on one line, and long text is truncated to fit.

When MaxItems is 2 or more, the first item and the last few items are kept, and the middle is replaced by a non-interactive ellipsis with separators on both sides. The ellipsis does not count toward the limit. MaxItems is a native extension; there is no dropdown for the collapsed items and no routing system. Item UI takes precedence over component UI, and SeparatorSlot uses the style of the preceding item. An empty Label uses the localized name of the navigation group.

### Pagination: control the page

Pagination binds an `*int`, and pages start at 1. Total is the total number of items, and a PageSize of 0 or less defaults to 10. When the total shrinks, the page is clamped to the valid range automatically; such corrections do not fire OnChange, which fires only when the user goes to a different page.

The component does not query data; your app loads the current page when it receives a change. A Siblings value of 0 or less defaults to 2, with a maximum of 10; for an explicit zero, use SiblingCount. ShowEdges defaults to false and Size to MD. For more appearance options, see [Pagination siblings and appearance](options.md#pagination-siblings-and-appearance).

Navigation controls and page numbers are square buttons. Controls at the bounds are disabled, and the arrows follow RTL. The ellipsis is not interactive. The accessible names for first, previous, next, last, and page N follow the language.

FirstIcon, PrevIcon, NextIcon, LastIcon, and EllipsisIcon override the default icons; explicit icons are not mirrored in RTL. ActiveColor and ActiveVariant default to Primary / Solid, and ordinary buttons default to Neutral / Outline. Root and List share one layout element. For compatibility, Item still applies to all buttons; First, Previous, Next, and Last can further target the control buttons, and Ellipsis styles the ellipsis. Your app handles page links, data queries, and browser-style history in OnChange.

### Stepper: choose a step in a process

Stepper binds the current step as a `*string`. StepItem contains ID, Title, Description, Icon, Disabled, UI, and Content. An unknown ID means the process has not started, and the first item is not selected automatically. Content synchronously builds the current step's body; switching steps unmounts the old body. Your app stores business drafts. Reordering or changing orientation keeps the body's identity for the same ID.

Linear is an Optional[bool] that defaults to true, which allows going back or forward only to the next adjacent step. Set it to `nui.Set(false)` to allow selecting any available step. CanSelect can veto an activation. Arrow keys only move focus; Enter/Space or clicking the dot activates a step. The component does not skip disabled steps in a linear process automatically; your app arranges the flow.

Horizontal steppers give each step equal width, with the dot centered and the text below. Vertical places the text next to the dot and makes the connecting line vertical. Size defaults to MD, and the dot diameter is 24/32/40/48/56 DIP for XS to XL, scaled by Density. Earlier steps and their lines use the completed color, and the current step is highlighted. The original icon or number is kept and is not replaced by a check mark.

The step body is below a horizontal list or beside a vertical one. List styles the step list and Panel styles the current body. With Neutral, the dots and completed lines use InvertedBackground. Disabled disables only the step headers; controls in the body keep their own disabled settings. If the focused header is removed or disabled, or the focused body is switched, focus moves back to an available header. Changing the value from the app does not fire OnChange. Nuxt UI's next/prev and hasNext/hasPrev are implemented in native apps from the ID and the item list.

### Timeline: display a timeline

TimelineItem contains ID, Title, Description, Time, Icon, Color, and Content, and can also provide Avatar and UI. Timelines are vertical by default; Horizontal places the indicator icon and connecting line above the text.

Current is an optional zero-based index, and CurrentID takes precedence over it. When neither is set, no item is active or completed. Reverse changes how progress is interpreted but does not reorder the data. Clicking an item only calls OnSelect; it does not change the progress.

Time is a string your app formats. By default, its font size and line height are 12/20 DIP, configurable with DateSlot; the title and description default to 14/20 DIP. Header styles the container of the marker and connecting line, and horizontal items share the width equally. With Neutral, the active marker and completed lines use InvertedBackground. With OnSelect, items can be operated by click, or by Tab and Enter/Space; without a callback, they are not in the keyboard focus order.

If CurrentID has no match, Current is not used as a fallback, and an invalid index highlights nothing. The component uses list semantics and does not sort, convert time zones, or schedule animations. Per-item Color, extra Content, and the five sizes are deliberate native configuration options.

## Banner and ErrorView

### Banner: an in-page notice

Banner sits in the normal layout, and your app decides where. open=false hides it. When open is bound or Close options are provided, a close button is shown by default; CloseButton=false hides it explicitly. With a nil state, the banner is always shown; once closing is explicitly enabled, it uses internal state and reappears after being unmounted. To keep it closed across launches, your app stores that state; browser localStorage is not used.

Color defaults to Primary. With Neutral, the background uses InvertedBackground and the text uses Inverted. The title and the native-extension Description each stay on one line and are truncated. When the window is at least 1024 DIP wide, the content is centered between the side regions; when narrower, it starts from the leading edge. The Body container uses 16/24/32 DIP of horizontal padding depending on the window width and is limited by Theme.ContainerWidth. The minimum height is 48 DIP by default, and the font and icons follow the text scale.

Actions is a synchronous builder, and Buttons inside it default to Neutral / XS. OnSelect receives clicks or keyboard activation on the whole banner, while inner buttons act independently; your app decides what navigation to perform. Close and CloseIcon configure the close button, and CloseLabel sets its name. The button defaults to MD / Neutral / Ghost and shows a 10% background on hover or focus. OnClose fires once when the user closes the banner, not when the app opens or closes it. In very narrow areas, your app should reduce the number of actions. Banner does not pin itself to the top of the window.

### ErrorView: an error page

In Go, `nui.Error` is already a semantic color, so the error display function is called ErrorView. ErrorOptions provides Status, Title, Description, RetryLabel, Icon, Clear, Retry, OnRetry, Actions, and Decoration.

A 40 DIP icon is shown only when you pass Icon. The status code defaults to 16/24 DIP semibold Primary and is styled with StatusSlot. The title uses a 36 DIP font when the window is narrower than 640 DIP and 48 DIP otherwise. The description is 18/28 DIP in Muted. An empty title creates no text, and a description identical to the title is omitted.

The default retry button is LG / Primary / Solid with the label "Retry", but it is created only if you provide OnRetry. Clear defaults to true; false hides the default action. Retry configures the ButtonOptions. Actions replaces the entire action area and is not affected by Clear=false. The action area has a 32 DIP top margin and a 24 DIP gap by default, and no empty area is created when there are no actions.

When embedding an ErrorView in part of a page, reduce font sizes through parts such as UI[Label]. The native implementation keeps its padding, a 480 DIP maximum description width, and level-2 heading semantics, and does not force itself to fill the window; with a set height, the content is centered. The component does not call Nuxt's clearError or redirect. Your app handles panics, requests, navigation, and clearing errors.

## Part index

| Component constant suffix | Parts |
| --- | --- |
| Tooltip | Panel, Label, Trailing |
| Modal / Slideover | Backdrop, Panel, Header, Label, Description, Close, Body, Footer |
| Toast | Root (viewport only), Panel, Body, Leading, Label, Description, Actions, Close. ProgressOptions.UI uses Progress's Root and Indicator |
| DropdownMenu / ContextMenu | Panel, Header (group), Item, Body, Leading, Label, Description, Trailing (row-end container), Indicator, SeparatorSlot, Editor (filter root), EmptySlot. Build replaces the Body content |
| Tabs | Root, List, Trigger, Leading, Label, Trailing, Indicator, Body. TabItem.UI styles a single item |
| Collapsible | Root, Trigger, Leading, Label, Trailing, Body |
| Accordion | Root, Item, Header, Trigger, Leading, Label, Trailing, Body |
| Breadcrumb | Root/List, Item/Trigger, Leading, Label, SeparatorSlot, Ellipsis. Parts joined by a slash share one node |
| Pagination | Root/List, Item, First, Previous, Next, Last, Ellipsis, Leading, Label. Root and List share one node |
| Stepper | Root, List, Item, Header, Trigger, Indicator, Leading, SeparatorSlot, Body, Label, Description, Panel |
| Timeline | Root, Item, Header, Indicator, SeparatorSlot, Body, DateSlot, Label, Description |
| Banner | Root, Body, Header, Leading, Label, Description, Actions, Trailing, Close |
| Error | Root, Leading, StatusSlot, Label, Description, Actions |

In Stepper, Header is the container for the dot and the connecting line, Trigger is the round button, Leading is the icon, SeparatorSlot is the connecting line, and Body is the text area.

Tabs, Stepper, and Timeline accept global and instance Color and Size; Banner and Toast accept Color. Color on menu items and Timeline items is a per-item setting. Other styles are configured through the parts listed above.

## Examples

The Gallery's **Overlays & menus** and **Navigation & disclosure** pages show overlays, menus, and navigation. The Settings example includes editing, validation, reset confirmation, Toast undo, and notification filtering. The examples keep state in the window model; long-term storage is left to your app.

For platform and IME limitations, see [Compatibility and known limitations](compatibility.md).
