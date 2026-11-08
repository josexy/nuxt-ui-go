# Advanced interaction, charts, and layout

This page covers command search, app sidebars, bottom sheets, file selection, charts, and resizable layouts. Component state is stored per window; each section explains where to initialize data and start background tasks.

| What you want to build | Where to look |
| --- | --- |
| Command search opened with a shortcut | [CommandPalette](#commandpalette) |
| App navigation, a resizable sidebar, or navigation for narrow windows | [NavigationMenu and Sidebar](#navigationmenu-and-sidebar) |
| A draggable bottom sheet, file selection, or upload status | [Drawer](#drawer), [FileUpload](#fileupload-and-resource-ownership) |
| Charts or rotating content | [Chart](#chart), [Carousel and Marquee](#carousel-and-marquee) |
| Split panes, page headers and footers, and footer links | [Layout components](#splitter-headerlayout-footerlayout-and-footercolumns) |

## Usage and customization

Usually you declare state in the window model and pass it to the component in the view. The Decoration field in Options sets UI and the synchronous Style; for precedence, see [Style overrides](foundations.md#style-overrides). You can set the layout of the returned element right away, but do not keep Elements or Contexts for later builds or background tasks.

Each component's state types and available parts are collected in the [state and part index](#state-and-part-index) at the end. For platform coverage, see [Compatibility and known limitations](compatibility.md).

## CommandPalette

### Create commands and show the palette

CommandPalette provides a searchable list of commands and opens or closes with Ctrl+K by default. Create a CommandIndex first, then hand it to a CommandPaletteState that you keep between builds:

```go
// Run on initialization or when the commands change; do not create the index on every build.
state.Commands.SetIndex(nui.NewCommandIndex([]nui.CommandAction{
    {
        ID: "save",
        Label: "Save workspace",
        Group: "Workspace",
        Key: ui.KeyS,
        Mods: ui.Ctrl,
        Shortcut: "Ctrl S",
        OnSelect: save,
    },
}))
```

Always call the component in the view, including when it is closed:

```go
nui.CommandPalette(c, theme, &state.Commands, nui.CommandPaletteOptions{
    Title: "Workspace commands",
})
```

Clean up tasks when the window closes with `window.OnClosed(state.Commands.Close)`. If the component stops appearing in the view, call Close explicitly as well.

Every ID in the command tree must be non-empty and unique, and the tree can be at most 32 levels deep. NewCommandIndex copies the children, key caps, style maps, and search data, and organizes the hierarchy by parent action. Treat the returned Action metadata as read-only, and replace it with SetIndex when the data changes. Your app still manages icons, bitmap resources, and callbacks.

### Shortcuts and search

| Option or behavior | Description |
| --- | --- |
| ShortcutKey / ShortcutMods | Change the shortcut that opens the palette; DisableShortcut turns it off |
| An action's Key / Mods | Register a shortcut that runs the action; Shortcut is for display only |
| Label / Description / Suffix / Keywords | Matched by default; FilterFields chooses which ones. Matching is case-insensitive and finds Unicode characters as a subsequence |
| Groups and child actions | Results from the same group are shown together. PreserveGroupOrder keeps the index's group order. Children can set a Placeholder, and Backspace on an empty query goes back |
| SearchDelay / IgnoreFilter | Default 0 / false. A delay postpones the query but cancels the old request immediately. IgnoreFilter skips only the built-in matching |
| Kbds / Prefix / Suffix / Avatar / Chip / Loading / Active | Command content and appearance. Kbds only display shortcuts, and Loading does not disable the command |
| Result count | Up to 200 results by default, configurable up to 1,000. The list is virtualized |

The parent scope where you call the component determines the shortcut's scope; your app resolves conflicts with its own shortcuts. Search favors contiguous matches and matches at word starts, and highlights by rune index. It does not handle pinyin, word segmentation, or accent folding.

In the input, Down moves into the list; then the arrow keys, Home, and End move, and Enter activates. If a search is still running or the results do not match the current query, Enter does not run a stale action. The description appears on the line below the label, at the same font size, in Muted. SearchInput=false hides the input but keeps the query. Size, SearchInputOptions, and BackOptions configure the size, the input, and the back button.

Before a leaf command runs, OnSelectEvent is called, with the per-item handler first; PreventDefault cancels the command. After it runs, the older Action.OnSelect and the root OnSelect are called, and the palette closes first by default. CloseOnSelect=false keeps it open so the command can run repeatedly. If a callback for the same click changes the query, index, level, or request, the old command does not continue. Disabled restricts input, results, and registered shortcuts together.

### Background search and closing

The default search scans an immutable index in the background, keeps at most one pending result, and reads it during the UI build. It checks for cancellation every 128 items and keeps only a limited number of top-scoring results, so it never sorts every match. Stale tasks cannot write back.

The build after Open becomes false cancels the task, and Close cancels it immediately. The component has no long-running goroutines or timers, and its state can be used in only one window.

To use your own search service, provide `Search(ctx, ticket, query, parent)`. Your app starts the task and calls `state.Results.Complete(ticket, matches, err)` in `Window.Update`. It is also responsible for result IDs, parent relationships, and triggering redraws. FilterFields and IgnoreFilter only affect the built-in search. After replacing a Search closure, call Refresh. Searching() covers both the delay and the running request.

This component is a native modal command launcher. It does not provide Vue's embedded multi-select model, route fields, or Fuse-specific options.

The query field is affected by the [IME cancel issue](compatibility.md#canceling-ime-composition), and the full screen reader experience has not been verified.

## NavigationMenu and Sidebar

### What navigation and layout each handle

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

### Sidebar state and size

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

### Toggles, side, and callbacks

| Option | Default behavior |
| --- | --- |
| Collapsible | true; false prevents the user from collapsing the sidebar |
| Toggle | true; false hides the built-in button so your app can provide its own control elsewhere |
| Disabled | Disables built-in interaction; also inherits disabled ancestors |
| Side | Defaults to `nui.Set(nui.SideLeft)`, meaning the start side in the reading direction; SideRight means the end side |

In RTL, the border, button icon, handle, and Slideover all adjust together. Your app still has to place a right-side sidebar after the main content in the Row. Widths are in DIP; web percentages, rem, and DashboardGroup units are not used.

OnResize, OnCollapsedChange, and OnOpenChange fire only when a user action or an automatic close by the component changes the value; writing the state directly from the app does not call them. OnResize reports the expanded width, and collapse changes are reported separately by OnCollapsedChange.

### Content regions and narrow windows

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

## Drawer

### Control the open state and height

Drawer is a modal bottom sheet with a separate drag handle. It reuses Slideover's exit animation and focus restoration. The title, body, and footer scroll together inside ContainerSlot, so the close actions can be reached even at the lowest snap height. Your app stores a DrawerState, controls opening with Open, and controls the snap position with Snap.

The default snap heights are 50% and 85% of the available window height. SnapPoints must be strictly increasing, and each value must be in `(0, 1]`; px strings are not accepted. With Inset=true, a 16 DIP margin is subtracted before the height is computed. To open from another edge, use Slideover. The shared Width, MaxHeight, and Fullscreen fields have no effect on Drawer.

### Dragging and closing

Only the handle drags; the body keeps scrolling. Dragging down lowers the panel, and on release a snap point is chosen from the current height and velocity. The drawer closes when its height falls below half the lowest snap point, or when it moves more than 80 DIP (the default) at more than 900 DIP/s. DismissDistance and DismissVelocity adjust these thresholds; 0 keeps the default, and values must be finite and non-negative. These are native thresholds and are not equivalent to Vaul's closeThreshold.

Dismissible=false prevents closing by dragging, clicking the overlay, or pressing Escape, and fires OnClosePrevent; an explicit close button still works. CloseButton defaults to false, but passing Close options shows the button by default, and an explicit false takes precedence. Handle defaults to true and is independent of Header; hiding it cancels any drag in progress.

The handle is focusable. Up and Down switch snap points, and Home and End go to the first and last. Losing pointer capture or window focus cancels the drag and restores the previous snap position. OnSnap reports only actual index changes; repeated keys at the bounds and outside changes are silent. OnClose reports close actions performed by the component; changing Open directly or a custom Done button does not report.

The visible handle is 48×6 DIP by default, rounded, in the Accented color, and styled with Indicator. Thumb is the 64×16 DIP interactive area. An empty title creates no empty heading, and the panel and handle have localized names. Snap height changes are immediate, and opening and closing use Slideover's animation. Call Drawer on every build, including when it is closed. Dragging from any edge, spring snapping, and touch gestures linked to body scrolling are not supported.

## FileUpload and resource ownership

### The component handles selection and state; your app handles transfer

FileUpload provides file selection, rule checks, a file list, and progress display. Your app implements the network upload; selecting files does not start a transfer.

FileUploadState.Files stores file metadata, not open file handles. UploadFile contains Path, Name, MIME, Size, a status, a Progress from 0 to 1, Error, and an app-supplied Avatar. Add assigns stable IDs that do not collide with existing entries and ignores entries with an identical Path. If you assign Files directly, your app must provide unique IDs. The component does not load or decode preview images itself.

| Option | Purpose |
| --- | --- |
| UploadRules.Accept | Extensions, MIME types, or categories, such as `.png`, `text/plain`, or `image/*`. Supports `*`, `*/*`, comma-separated lists, compound extensions, and case-insensitive matching. An empty or whitespace-only value allows all types |
| MaxBytes / MaxFiles | Maximum size per file and maximum number of accepted files |
| Dialog | The native file picker's parent window, multiple selection, and filters |
| Choose(ctx) | Replaces the default file source |
| Variant / Layout / Position | Default UploadArea / UploadGrid / UploadOutside. UploadInside places only the UploadList inside the selection area |
| Icon / ShowIcon / Highlight | The upload entry icon, its visibility, and border highlighting. Neutral uses the inverted color, and a field error uses the error color first |
| FileIcon / FileImage / UploadFile.Avatar | The fallback file icon and the app-supplied preview accessory. FileImage=false shows only FileIcon |
| FileDelete / FileDeleteOptions / FileDeleteIcon | Whether the delete button is shown, its options, and its icon. Defaults to Neutral / LinkVariant / component size in the list, and Neutral / Solid / XS in the grid |
| Preview | Default true; false hides the file list but keeps the data and tasks. See [Sidebar, navigation, and file selection](options.md#sidebar-navigation-and-file-selection) |

Files that exceed a limit or cannot be read are shown as Rejected with the reason, and the user can remove them. A non-positive MaxBytes or MaxFiles means no limit. Rejected files do not count toward the limit, while accepted files that later fail or are canceled still count. The component inherits Disabled from its parent container and FormField, as well as FormField's Size, label, and error message. The default color and size come from ComponentFileUpload.

OnChange fires once, with a copy of the list, only when a selection or removal actually changes the list. Duplicate paths, a canceled selection, a failed selection, and direct Add, Remove, or Report calls from your app do not fire it. OnReject reports the newly rejected files before OnChange and does not clear rejected entries from the queue. After a removal or a canceled transfer, focus returns to the file selection entry point.

Go keeps an upload queue model. Dialog.Multiple only controls whether one native dialog accepts several files; you cannot use it to limit the queue size. Nuxt UI's single-file modelValue replacement, `multiple` forcing the area variant, and the web input reset have no direct equivalent. The grid shows file names, status, and progress in wrapping cards and does not claim to match upstream's image grid exactly.

### How files are checked

By default, the selection action calls `mygo.Dialog.Open` in the background and then checks for regular files with `os.Stat`. External DroppedFiles go through the same checks and rules. You can also call InspectUploadFiles and PickUploadFiles directly.

MIME types are derived only from the extension; file contents are not read. Accept checks selected and dropped files but is not converted into native Dialog.Filters automatically. If your business needs a trustworthy file type, validate the content yourself. Dropzone=false registers no drop target, and Interactive=false keeps a separate selection button. When Choose returns an empty result, the selection was canceled; when it returns an error, State.Error is set and the file queue is unchanged.

### Start, cancel, and retry uploads

```go
// After creating the window, set the parent window and register cleanup.
options.Dialog = mygo.OpenDialogOptions{Parent: window, Multiple: true}
window.OnClosed(state.Files.Dispose)

// Start the task in your app's upload action; your app implements upload.
ctx, ticket := state.Files.Begin(fileID)
go func() {
    err := upload(ctx, path)
    window.Update(func() {
        state.Files.Report(ticket, 1, true, err)
    })
}()
```

Begin starts a task, and your app publishes progress, completion, or errors with Report in Window.Update. Background code must not change Files directly. The UI shows Ready, Uploading, Complete, Failed, Canceled, and Rejected. After a failure, you can call Begin again; the library has no additional retry protocol.

Cancel(id) cancels the task's context, and Remove(id) cancels and removes the file. Old tickets from before a cancel, removal, or retry, or from a task that has already completed, can no longer write back. OnCancel and OnRemove can perform extra cleanup. Your app also closes any file streams, temporary files, and network connections it opened.

### Cleaning up on hide and close

Call Dispose when the component is hidden or unmounted, or when the window closes, to cancel file checks and upload tasks. Disabling the field also cancels pending file checks and discards late results.

Canceling the context cannot force an open native file dialog to close, but its result is discarded. For platform verification of native selection, multiple selection, cancellation, drag and drop, and filters, see [Requirements and platforms](compatibility.md#requirements-and-platforms). Your app manages macOS sandbox permissions and bookmarks.

## Chart

### Prepare the data, then draw the chart

Chart draws with MyGo's `ui/chart` and adds theming, interactive legends, and animation. A ChartState should be owned by a single chart instance in a single window. Do not copy a rendered state or share mutable data with background code.

Call the setter methods when the data is initialized or changes, and check the error:

```go
err := state.Activity.SetCategories([]string{"Jan", "Feb", "Mar"}, []nui.ChartSeries{
    {ID: "build", Label: "Builds", Values: []float64{30, 45, 38}},
    {ID: "review", Label: "Reviews", Values: []float64{12, 19, 16}},
})
if err != nil {
    return err
}
```

Put this fragment in a data initialization function that returns an error. Write background results to the window state through Window.Update. Create the chart in the view function:

```go
nui.Chart(c, theme, &state.Activity, nui.ChartOptions{
    Label: "Activity",
    Kind: nui.ChartArea,
    Curve: nui.ChartCurveMonotone,
    Stack: nui.ChartStackStandard,
    Height: 240,
    AnimationDuration: 400 * time.Millisecond,
})
```

### Chart types and data APIs

| Need | API |
| --- | --- |
| Line, bar, area, scatter, pie, and donut charts | ChartLine, ChartBar, ChartArea, ChartScatter, ChartPie, ChartDonut |
| Numeric / category / time X axis | SetData, SetCategories, SetTimeData |
| Pie and donut charts | SetSlices([]ChartSlice) |
| Curves | ChartCurveLinear, ChartCurveMonotone, ChartCurveStepAfter, ChartCurveStepBefore, ChartCurveStep |
| Stacking | ChartStackNone, ChartStackStandard, ChartStackNormalized, ChartStackCenter; for bar and area charts |

Every setter validates before copying, so an error leaves the existing data unchanged. The data must meet these requirements:

| Method | Data requirements |
| --- | --- |
| SetData(xs, series) | X values are finite and strictly increasing; each series has the same length as X; series IDs are unique and non-empty. A Y value of NaN or Inf marks a missing value |
| SetCategories(labels, series) | Category labels are unique and non-empty; uses a band category axis |
| SetTimeData(times, series) | Times are strictly increasing; ticks use the time zone of the first entry |
| SetSlices(slices) | Slice IDs are unique and non-empty; values are finite and non-negative, and the total is finite. A value of 0 keeps its legend entry but draws no slice |

Scatter charts also use shared, ordered X data. ChartSlice contains ID, Label, Value, and Color. The same state can hold both Cartesian data and slice data, and Kind chooses which one to draw.

### Axes, colors, and stacking

The Y axis always includes 0. Standard stacks positive values upward and negative values downward separately. Normalized divides by the sum of the absolute visible values at each X, so all-positive values add up to 100%, and mixed positive and negative values share that denominator. Center centers the algebraic sum on 0.

The default series colors are, in order, Primary, blue, violet, amber, rose, and cyan. After the first round, each additional round mixes the colors with the background by a factor of 0.65. Light mode uses shade 500 and dark mode uses shade 400. Colors can supply a list of native colors, and a series' explicit Color takes precedence; Neutral uses the inverted color. Mixing uses MyGo's native color interpolation and does not claim to match CSS oklab mixing exactly.

FormatX, FormatY, and FormatTime set the formatting. Numbers are formatted with Go by default, not with browser Intl number formatting. XAxis and YAxis default to true and control each axis's ticks and title; Grid independently controls the horizontal grid lines. XLabel is centered below the chart, and YLabel is shown horizontally above it. In RTL, the X axis data and ticks are reversed, the Y axis moves to the right, and physical arrow keys and pointer positions are mapped accordingly.

Height defaults to 300 DIP, and a positive value takes precedence; otherwise, a positive AspectRatio computes the height from the actual width. A positive Width fixes the width; otherwise, the chart follows its parent and uses 640 before the first measurement. Height applies to the plot area; the legend, axis titles, and status text take extra space. All sizes must be finite. BarPadding, SlicePadding, LineWidth, and PointSize remain as native geometry options. Points defaults to false; when on, line charts show the sampled data points. PointSize does not affect the separate focus marker.

A donut's Thickness defaults to 24 DIP. An explicit DonutInnerRadius can still be given as a ratio and takes precedence over Thickness; it is clamped to 0–0.95. Thickness is clamped to non-negative values, and 0 makes the inner and outer edges meet. Center synchronously builds content in the center of the plot area without intercepting the pointer; the callback must not keep Elements or Contexts.

### Legends and data selection

When Legend is unset, the legend appears only if there is more than one series or slice in total, counting hidden and zero-value items; an explicit true or false takes precedence. The legend is clickable by default, and focused entries can toggle a series with Space or Enter. OnLegendChange(id, visible) reports only user actions. Your app can query and change visibility with SeriesVisible/SetSeriesVisible and SliceVisible/SetSliceVisible. An empty Label defaults to the corresponding ID.

InteractiveLegend=false only displays the legend, and Legend=false hides it entirely. Hiding a series recomputes the axes and stacking, and hiding a slice recomputes the visible proportions. Visibility is kept by ID across updates and reordering, and is cleared when an ID is removed. Even when everything is hidden, the legend stays so the user can restore items.

Cartesian charts support pointing at the nearest X with the mouse, and Left, Right, Home, End, and Escape; Home and End go to the logical first and last data points. The status text at the bottom and the Tooltip show the raw values, and Tooltip=false hides only the tooltip. TooltipOptions configures settings such as position; the default is TopCenter with a 12 DIP offset and no delay. Pie and donut charts support pointing at slices and arrow keys, and show percentages of the visible total. The donut's center does not respond to slice pointing. Selection on dense data always looks up the raw values.

OnFocusChange fires when the pointer or keyboard changes the focused row or slice; repeated keys at the bounds do not fire it again. Clicking or pressing Enter/Space fires OnSelect, and activating again can fire it again. ChartSelection.Index is the raw X row or slice index, or -1 for no focus; SliceID is used only by radial charts. Go works in whole X rows and has no separate per-series point, focusGroupChange object, or DOM event from the web engine. The chart uses focusable group semantics, and native assistive technology can still read the center text and the empty state.

### Animation

Animate is on by default, AnimationDuration defaults to 350 ms, and both theme and system reduced motion settings are respected. The first display appears directly; only later data or option changes animate.

When the chart type, series, and missing-value structure stay the same, coordinates are interpolated, and switching between Pie and Donut also animates the inner radius. When the chart type, series, point count, visibility, or missing-value structure changes, the chart cross-fades. An update in the middle of an animation continues from the current frame, keeping at most 3 exiting frames and 1 target frame. Axis labels and tooltips use the target data.

When the animation finishes, or the chart is invisible or unmounted, it stops requesting frames; there is no background timer. `Animate: nui.Set(false)` shows the target values immediately.

### Large data and limits

Prepare(pixelWidth) can prepare the default line cache in advance; Chart prepares everything else. The sampling width is capped at 2,048 and is recomputed only after the data, visibility, size, or geometry options change. Drawing and interpolation work on the sampled shapes.

Unstacked data uses min-max sampling. When there are too many gaps in the data, fixed buckets are used instead, and buckets containing missing values stay disconnected. Dense stacked data shares bucket endpoints and overall envelope extremes so the X boundaries stay aligned, and details of a single series within one pixel may be dropped. Dense bar charts are not summed or averaged automatically; aggregate your business data first if you need that.

The chart is modeled on the TanStack Charts used by Nuxt UI 4.12.0, but Go actually uses MyGo's ui/chart and depends on no web charting engine. Native capabilities such as scatter and pie charts and numeric and time axes are kept. Arbitrary TanStack definitions, dual Y axes, logarithmic axes, brush zoom, and touch gestures are not provided. For low-level custom marks and layouts, compose ui/chart directly. The Gallery's Chart page and **Activity & attachments** page show chart types, axes, RTL, events, center content, data updates, and interactive legends.

## Carousel and Marquee

### Carousel: switch content page by page

CarouselItem contains a stable ID, a Label, and a synchronous Build; your app can control CarouselState.Index and Paused. The component builds only the current item and its neighbors. Neighboring items are not interactive, but their Build still runs, so never start business operations unconditionally in a build callback.

| Action or option | Behavior |
| --- | --- |
| Dragging | Drags along the Orientation. Moving more than 20% of the viewport, or faster than 600 DIP/s, changes the page; on release, it snaps into place |
| Previous / Next, arrow keys, Home / End | Navigate by page, with ←→ horizontally and ↑↓ vertically |
| Arrows / Dots | Control the previous/next buttons and dot navigation |
| Wrap=true | Buttons, keyboard, and autoplay can wrap from the last item to the first; dragging still stops at the ends |
| WatchDrag | Default true. false registers no dragging while other navigation keeps working. Changing the orientation cancels the current drag |
| Prev / Next, PrevIcon / NextIcon | Button and icon options. Defaults to Outline / Neutral; the default arrow follows the orientation, and explicit icons are shown as is |
| StopOnInteraction | Default true. With autoplay on, the arrows, keyboard, and a completed drag set Paused=true, and Play resumes it. false keeps automatic resumption; dots only change position |

The current position is exposed only as the root element's assistive technology value, such as "Slide 1 of 3", using the slidePosition language key; no page counter is shown. Index is positional, so after reordering your app decides the index, and removing items clamps it to the valid range. Outside updates do not fire OnChange. If the focused element on the current page is removed, focus returns to the viewport. CarouselItem.Decoration can override the Item part; your app still stores business data such as drafts.

Autoplay is off by default; a positive Autoplay sets the interval. Hovering, focus inside the component, Disabled, Paused, Active=false, being invisible, an inactive window, or reduced motion all stop scheduling. Active only indicates whether native autoplay is active; it does not disable manual navigation, which differs from Embla's active. With Wrap=false, autoplay stops at the last item. When a temporary pause ends, the timer restarts. The play/pause button is a native extension.

OnChange fires once when user navigation or autoplay actually changes the index; drag and timer results produced during a build are reported after the build ends. Canceling a drag clears the offset, and an empty list also clears the drag state. Buttons are centered based on their actual size, and disabling at the bounds takes precedence over button options. The default buttons sit 16 DIP inside the viewport, so leave room for them in your content; the Gallery and website show an input layout using Item padding.

Go builds only the current page and its neighbors. It has no Embla free dragging, slide groups, media queries, physics-based animation duration, auto height, or plugin objects. Switching pages positions the content immediately. Dots use native selected buttons and do not claim to implement a full ARIA tablist.

### Marquee: continuously scrolling content

Marquee is for decorative content. It supports Orientation/Vertical, Reverse, Speed in DIP/s, Gap, Height, and pausing with Space. Speed defaults to 32 DIP/s, Gap to 64 DIP scaled by the theme, and Height to 44 DIP. These floating-point values must be finite, and non-positive values use the defaults. Speed differs from upstream's CSS cycle of 20 seconds. To pause, use Paused or Active.

When Repeat is unset, the content is repeated 4 times; an explicit `nui.Set(0)` builds no content, and negative values are invalid. Each copy is at least as long as the container in the direction of motion, centered on the other axis, with its items evenly distributed, so even short content scrolls; a single copy may show a gap as it loops. Overlay fades to the background color over the outer third at each end by default; `Overlay: nui.Set(false)` turns the fade off.

When ContentKey changes, scrolling restarts from the beginning. PauseOnHover defaults to false; when enabled, hovering pauses the motion. Focus, Paused, Active=false, a disabled parent, and reduced motion also stop it. Horizontal motion follows RTL. An offset set from outside is normalized to the current content cycle, so the position stays valid when the container shrinks; a non-finite offset resets to 0.

The content moves in frames that only repaint, without building the view: Marquee builds again about once a second, to wrap the content back within its copies and to check whether it still shows. MarqueeState.Offset is therefore the phase as of the last build; read it in a build. Assistive technology also sees the copies where the last build placed them.

Root styles the container, and Body styles each copy. Body.Gap changes only the spacing between items inside a copy; copies are still separated by Gap. The usual override order of theme, instance UI, and the final Style applies.

**The content callback runs Repeat times on every build, so put only display elements in it.** Do not put business buttons, inputs, or side effects there. The underlying layer has no API for hiding a whole subtree from assistive technology, so verify how repeated content is read with the screen readers you target. To limit resource use, the website preview allows at most 100 copies; the Go API sets no such limit.

### Stopping animation when hidden

Neither component builds the view on every frame, and neither uses timer goroutines: Carousel autoplay asks for one build per interval, and Marquee moves its content in frames that only repaint, building about once a second. Visibility is what showed in the previous frame within the window and the scroll containers and clips around the component, so content scrolled out of a ScrollArea, clipped away, or kept mounted in a hidden panel (Tabs or Collapsible with KeepMounted) stops by itself, as it does in a minimized or covered window. Set Active=false for content your app hides by other means, such as covering it with another element, or stop building the component.

After unmounting, no more frames are requested and no paused time accumulates.

## Splitter, HeaderLayout, FooterLayout, and FooterColumns

### Splitter: split two panes

Splitter is built on MyGo's layout, focus, dragging, and shortcuts, and binds a `*float32`. The default unit is DIP. With Percent=true, the value is 0–100, a share of the space the two panes can use, not counting the divider. Orientation=Vertical or Vertical=true splits top and bottom. Horizontal layout and arrow keys are mirrored in RTL.

Min and Max use the same unit as the bound value. The default lower bound is 0, and the upper bound is 1e7 DIP or 100%. An explicit 0 is valid. OnChange fires once after a pointer or keyboard action actually changes the value; repeated input at a bound does not fire it. When your app assigns the value directly, it is silently clamped on the next build, and a non-finite model value is treated as 0 first. Constraints and steps must be finite, and the maximum cannot be lower than the effective minimum.

When KeyboardResizeBy is unset, each key press moves 10 DIP scaled by the theme. An explicit value uses the bound value's unit, and 0 turns off arrow-key resizing. Home and End go to the reachable bounds of the range. Disabled locks only the divider, and the panes stay editable; to disable the whole subtree, disable the parent container.

Both panes keep a native minimum size of 40 DIP; in DIP mode, the first pane can explicitly be set smaller. When space runs out, the layout shrinks but the model value is kept, so a model value of 0 does not mean a collapsed pane. The layout reserves only 1 DIP for the divider line, and an 8 DIP drag area overlaps the pane edges without pushing the panes apart; when nested, horizontal and vertical lines meet without gaps. Root, Leading, Trailing, SeparatorSlot, and Indicator style the container, the two panes, the drag area, and the center line. The accessible Range and Step use the bound value's unit and the configured constraints rather than always using the underlying DIP.

Two panes and nesting are supported. Multi-pane coordination, collapsible snapping, autosave, and layout serialization are not. Your app manages the persistent model, and layout and orientation changes keep input state in the panes.

### HeaderLayout and FooterLayout

HeaderLayout is a native app header with Title, the native-extension Description, Leading, and Actions. The title is empty by default and, when non-empty, is a level-1 heading; an empty title creates no empty heading node. The minimum height is 64 DIP scaled by the theme, the horizontal and vertical padding is 24 and 8 DIP, the gap between regions is 12 DIP, and the gap inside the leading and action areas is 6 DIP. Root, Leading, Body (the title text container), Label, Description, and Actions all have working style entry points. Changing Leading keeps the identity and input state of the controls in Actions.

HeaderLayout does not create a home link, web sticky positioning, a breakpoint menu, or a route listener. The Gallery and website examples combine a menu button in Actions with Sidebar, and manage opening, closing on navigation, and focus through a persistent SidebarState, NavigationKey, and ReturnFocus. The examples use a 2000 DIP breakpoint so the menu is easy to try at common preview widths; this is not a HeaderOptions default. Header's Body part refers to the title text container, which differs from the web Header's menu body slot.

FooterLayout shows a description with Text and builds actions with Actions. Root, Label, and Actions style the whole footer, the text, and the action container. The default horizontal and vertical padding is 20 and 12 DIP, the gap between regions is 12 DIP, the gap inside the action area is 6 DIP, and there is no top border. Text and actions can wrap within the available width, and action styles are not mistakenly applied to individual buttons.

Both functions support text scaling, RTL, and composition with native layouts. They do not change the system title bar and do not provide web DOM element types or every slot. Your app composes multi-region footers, extra content above or below, and breakpoint-specific arrangements. In Go, Header and Footer are already part constants, so the functions are named HeaderLayout and FooterLayout.

### FooterColumns

FooterColumns takes columns with stable, non-empty IDs, and link IDs must be unique within each column. A non-empty Title is shown as a level-3 heading. Items use NavigationItem's ID, Label, Icon, Active, Disabled, OnSelect, and Decoration. The other fields (groups, descriptions, nested children, avatars, badges, Tooltip, and Popover) are for NavigationMenu and are not used by this component.

When the user activates an available link, the item's OnSelect is called first, with NavigationEvent.ID set to the target ID; PreventDefault cancels the root OnNavigate. Disabled links do not run callbacks. Your app controls Active; active links use Primary with a medium font weight by default, inactive links use Muted, and hover uses Text. Natively, links take the pointer and Enter; as on the web, Space does not activate a link but is left to the container around it. There are no URLs, automatic route matching, or external link markers.

MinColumnWidth defaults to 160 DIP. Non-positive values use the default, and non-finite values are invalid. This value is the initial column width in the flexible layout and drives wrapping; columns and links are constrained by the available width. Columns are top-aligned. By default, columns are 32 DIP apart, the title is 24 DIP above the list, links are 16 DIP apart, and long links are truncated to one line. RTL adjusts the order, and reordering columns keeps their stable IDs.

Root, Body (each column), Label (the column title), List, Item (a link), ItemLeadingSlot, and ItemLabelSlot can each be styled. Item UI overrides component UI, and the item's Style runs after the component's Style. Style links with Item, ItemLeadingSlot, and ItemLabelSlot. The website's full data panel keeps reference information for the other NavigationItem fields and marks where they apply.

## State and part index

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| CommandPalette | CommandPaletteState, CommandIndex | Root (modal), Editor, BackSlot, Panel, List, Label, Item, Body, ItemLeadingSlot, ItemLabelSlot, ItemLabelPrefixSlot, ItemLabelSuffixSlot, Description, ItemTrailingSlot, Indicator, EmptySlot, LoadingSlot, LoadingIconSlot |
| NavigationMenu | NavigationState, active ID `*string` | Root, List, Panel, Body, Label, Item, ItemLeadingSlot, ItemLabelSlot, Description, ItemTrailingSlot, Indicator |
| Drawer | DrawerState | Backdrop, Panel, ContainerSlot, Thumb, Indicator, Header, Label, Description, Body, Footer, Close |
| FileUpload | FileUploadState | Root, Body, Leading, Label, Description, Trigger, List, Item, ItemLeadingSlot, ItemLabelSlot, StatusSlot, ErrorSlot, ItemTrailingSlot |
| Chart | ChartState | Root, Body, Item (legend), Label, Description, EmptySlot, List, Indicator, XAxisSlot, YAxisSlot, CenterSlot |
| Carousel | CarouselState | Root, Body, Item, Previous, Next, Actions, List, Indicator, Trigger, EmptySlot |
| Marquee | MarqueeState | Root, Body |
| Splitter | `*float32` | Root, Leading, Trailing, SeparatorSlot, Indicator |
| Sidebar | `*bool` collapsed, optional SidebarState | Root, Header, Label, Description, Actions, Trigger, Body, Footer, SeparatorSlot. The overlay also has Panel, Backdrop, Close |
| HeaderLayout | HeaderOptions | Root, Leading, Body, Label, Description, Actions |
| FooterLayout | FooterOptions | Root, Label, Actions |
| FooterColumns | `[]FooterColumn` | Root, Body, Label, List, Item, ItemLeadingSlot, ItemLabelSlot |

## Examples

The Gallery's **Advanced interactions**, **Activity & attachments**, and **Application layout** pages show these components, and **Props & behavior** and **Orientation** demonstrate option changes.

The Gallery's own leftmost navigation also uses Sidebar with NavigationMenu. It is 256 DIP wide by default, can be dragged between 200 and 360 DIP, and collapses to a 64 DIP icon bar. Switching pages keeps the width and collapsed state, which are stored per window; this sidebar is independent of the demo sidebar on the Application layout page. Automated tests can find its button and handle by the names "Toggle gallery sidebar" and "Resize gallery sidebar".
