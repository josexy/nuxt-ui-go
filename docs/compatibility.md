# Compatibility and known limitations

Before choosing a component, check how thoroughly your target platform has been verified and whether your app needs any of the unsupported features listed here.

NuxtUI Go uses MyGo for native layout, input, and drawing, and follows the visual style of Nuxt UI v4. You write against this library's Go API; Vue component imports, CSS classes, Nuxt routing, and browser plugins do not work in native apps. APIs and defaults may change between versions; changes are recorded in the [changelog](../CHANGELOG.md).

## Requirements and platforms

You need Go **1.27.3 or later**, and you build with `CGO_ENABLED=0`. This repository pins the MyGo runtime and tools to `v0.9.0`; see [go.mod](../go.mod). Native apps do not need Node.js, Vue, or a WebView at runtime, but they do need a desktop graphics session.

The table records what the project has verified. Platforms have not all been tested to the same depth.

| Platform | Verified | Not yet verified |
| --- | --- | --- |
| Windows amd64 | Builds, automated component tests, real Gallery window | Full system IME workflows, DPI changes across monitors, screen reader experience |
| Linux amd64 | Builds, automated component tests, and first frame of a hidden native window on WSL Ubuntu 22.04 | Full desktop interaction, system IME, other distributions |
| macOS arm64 | Cross-compilation | Running and interacting on the desktop |
| Browser WASM preview | The project provides [website previews and checks](../website/README.md) | Native file dialogs and the full assistive technology experience are outside the browser preview's scope |
| Other architectures | No verification records | Check in the target environment |

Check native file selection on each target system: cancellation, multiple selection, filters, drag and drop between apps, and results that arrive after the window closes. Your app is responsible for macOS sandbox permissions and file access bookmarks.

## Canceling IME composition

MyGo v0.9.0, the pinned dependency, still has this known issue: **if you select text, start an IME composition, and then cancel it, the selected text is lost.** Undo restores it. Headless `ui.Tester` checks reproduce it with Input and Textarea, and will flag the change once MyGo keeps the selection.

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

`PinInputOptions{Numeric: true}`, InputNumber, InputDate, and InputTime accept only numeric or date/time formats and do not enable IME. Other input methods still need testing on each target platform.

## Feature scope

### Appearance, input, and overlays

| Area | Supported | Notes |
| --- | --- | --- |
| Theme and sizing | DIP units, native fonts, XS/SM/MD/LG/XL sizes | Fonts and fallback fonts can increase line height. There is no Tailwind or CSS runtime. Native disabled controls use 50% opacity. Some sizes, states, and animations differ from Nuxt UI |
| Language and input | English, Simplified Chinese, Gregorian dates, locale-specific segment order, 12- and 24-hour clocks | Digits are ASCII and the decimal separator is `.`. There is no full locale-aware number input, no alternative calendars, no time ranges, and no per-segment undo/redo |
| Segmented date and time | Single dates and date ranges, locale order, steps, date/time bounds and unavailable predicates, time granularity, optional native calendar | Date carries no time of day. ClockTime carries no date, milliseconds, or time zone. Build a time range from two InputTime controls. The full screen reader experience is unverified |
| ColorPicker | sRGB byte channels, HEX/HEXA, native keyboard support, alpha, swatches, sizes, change throttling | Uses MyGo's horizontal tracks. The internal SV area, tracks, and HEX parts are not public. There is no HSL/CMYK/Lab text editing or wide-gamut editing. The full screen reader experience is unverified |
| Overlays | Tooltip, Popover, Modal, Slideover, and others | Tooltip's default hover delay is 600 ms; override it with DelayDuration. Popover, Select, menus, Tooltip, and Modal disappear immediately on close; Slideover has an exit animation |
| Notifications | Toast queue, action buttons, timeouts, pausing | At most 3 toasts show at once. Queued toasts keep counting down |

### Data and files

| Area | Supported | Not provided, or left to the app |
| --- | --- | --- |
| Calendar | Gregorian single dates and ranges, day/month/year navigation, wrapping multi-month layout, range limits | No discrete multi-date selection and no separate year-month value model. WeekStart defaults to Sunday. Week numbers are computed from the explicit week start; full Intl locale week rules are not guaranteed |
| Table / Tree | Native virtual lists and stable IDs. Table supports multi-column and manual sorting, separate model notifications, and loading content. Tree links parent and child selection for loaded nodes | Table headers have no keyboard focus of their own. No full TanStack configuration, frozen columns, multi-level headers, row grouping and aggregation, or tree drag and drop. See [Table](forms-data.md#table) and [Tree](forms-data.md#tree) |
| Listbox / SelectMenu | Stable string IDs, structural rows, avatars and dots, filtering and sorting, per-item cancellable actions, native virtual lists | Does not accept arbitrary Vue objects or dynamic property paths. Search uses Unicode case folding and diacritic normalization, which is not full Intl locale collation. Virtualization does not expose overscan or estimateSize |
| Chart | Six native chart types; category, value, and time axes; stacking; interactive legends; axis and tooltip options; RTL; row and sector selection notifications | Modeled on TanStack Charts. No arbitrary definitions, standalone web point/group objects, dual Y axes, logarithmic axes, or brush zoom. See [Chart](advanced.md#chart) |
| FileUpload | Native selection and queue; type, size, and count rules; rejection notifications; app-supplied preview accessories; progress and cancel states | The app handles network upload, content validation, image loading, file streams, and persistence. Dialog.Multiple only controls whether one dialog accepts several files. The website uses a simulated Choose. See [FileUpload](advanced.md#fileupload-and-resource-ownership) |

### Layout and navigation

| Area | Current behavior and limits |
| --- | --- |
| Drawer | Dragged by a separate handle at the bottom; the content scrolls as a whole. No dragging from other edges, Vaul background scaling, or gestures linked to body scrolling. SnapPoints accept only height fractions; native close thresholds and animation are documented separately |
| Carousel / Splitter | Carousel buttons, keyboard, and autoplay can loop, but dragging still stops at the ends. Active only gates autoplay. No full Embla plugin set or ARIA tablist. Splitter splits into exactly two panels |
| NavigationMenu / Sidebar | Icon navigation, tooltips, child popovers, resizing, and switching to an overlay based on the window width in DIP. The app implements routing and persistence |
| Layout switching | When Sidebar crosses a breakpoint, its controls are rebuilt but the app model is kept. If the content being edited is removed, focus returns to an available entry point. Local scroll position and selection inside controls are not carried over |
| Main / Header / Footer landmarks | Main, HeaderLayout, and FooterLayout are the main, banner, and content-information landmarks, as Nuxt UI's `<main>`, `<header>`, and `<footer>`. A window should have at most one of each, so opt nested instances out with `.Role(ui.RoleGroup)` on the returned element. NavigationMenu and Sidebar have no navigation or complementary landmark role yet |
| Header / Footer / FooterColumns | Native header and footer areas use builders. They do not include every web slot, routing, or sticky positioning. Combine the header menu from components such as Sidebar. FooterColumns supports Active, per-item cancellable navigation, and styles; other NavigationItem fields have their own applicability notes |
| Tabs | Panels are tab panels named after their tab and linked from it with `Controls`. Real screen-reader output for this relation has not been verified |
| Other Nuxt UI features | No web mega menu, CSS breakpoints, DOM slots, or NuxtLink compatibility layer. No full Page, Dashboard, AI Chat, Editor, Content, color mode, or LocaleSelect components; call the theme and language APIs directly instead |

For each component's options, defaults, keyboard behavior, and style parts, the [component guide](README.md) is the reference.

## Accessibility and motion

Components use MyGo's focus, keyboard, and native assistive technology semantics, and they respect the system theme and reduced motion preferences. The full screen reader experience has not been verified.

Since MyGo v0.9.0, `ui.RoleStatus` is a polite, atomic live region. The library's own progress, error, empty, and loading messages keep stable identities: such a status is read once when its text changes and is silent when it first appears, clears, or rebuilds unchanged. Alert and Banner take their identity from the key the app gives them. CommandPalette also announces its progress and error messages once when they appear. FileUpload keeps empty status regions mounted, so adding a message to an exposed region is read once; hidden updates are silent, and mounting or exposing an existing message does not announce it as a change. Chart announces keyboard selection once and never reads pointer hover. Alert and Banner read their new title and description once when either changes. These rules are checked headlessly with `Tester.UseScreenReader`; what VoiceOver, NVDA, Narrator, and Orca actually say has not been verified.

Pay particular attention to these areas:

| Area | Current limitation |
| --- | --- |
| Multi-select lists | The underlying layer does not expose complete multi-selection metadata |
| Vertical Slider / RangeSlider | Orientation metadata is missing, including for non-inverted vertical sliders |
| InputRating | Uses a single-focus slider role, which differs from the radio semantics used on the web |
| Marquee | How repeated content is presented to screen readers still needs verification |
| Tree | Provides Tree/TreeItem roles, levels, selection, and active item semantics. The current virtual list lacks the assistive technology expand/collapse actions built into `ui.Outline` |

Programmatic focus scrolls the newly focused control into view as little as needed, as described in [Focus and callbacks](inputs.md#focus-and-callbacks). `ui.Element.Focus` accepts `ui.FocusOptions{PreventScroll: true}`, but MyGo v0.9.0's `ui.Handle.Focus` takes no options, so focus moved through a persistent Handle, such as a Modal's or Sidebar's ReturnFocus, always reveals a target that is not fully visible.

When you supply custom content, keep the control names, focus, and interaction semantics intact. Check the contrast of custom colors separately. Mouse, keyboard, and Canvas WASM preview tests cover only those interactions; they do not replace screen reader testing on the target system.
