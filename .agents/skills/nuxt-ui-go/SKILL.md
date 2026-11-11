---
name: nuxt-ui-go
description: Use NuxtUI Go (github.com/josexy/nuxt-ui-go, Go package `nui`), a native MyGo component library styled after Nuxt UI v4. Use it whenever Go code calls or should call `nui.*` components (Button, Input, Select, SelectMenu, InputMenu, Modal, Toast, DropdownMenu, Tabs, Form, Table, Tree, Calendar, Sidebar, CommandPalette, Chart, FileUpload, …), when building screens, forms, dialogs or data views in a MyGo desktop app with this library, when choosing between similar components, and when theming or styling parts. It has one reference per component (73) with exact Go signatures, Options and state types, pitfalls and a runnable example. Use it even for small component tasks, because the Go API differs from Nuxt UI's Vue props and slots in ways memory gets wrong.
---

# NuxtUI Go components

NuxtUI Go is a native Go library built on MyGo (`github.com/egoist/mygo`). It borrows Nuxt UI v4's component names and looks, but its API is Go: options structs instead of props, builder callbacks instead of slots, app-owned state objects and pointers instead of `v-model`, and `Optional[T]` for "unset versus explicit zero". Nuxt UI knowledge tells you what a component is for, not how to call it, so look up the real API before writing code.

## How to work

1. **Pick the components.** Use the [component index](#component-index) and [choosing guide](#choosing-between-similar-components) below.
2. **Read each component's reference before writing code:** `references/components/<id>.md`, for example `references/components/select-menu.md`. Each reference contains:
   - **Usage** and **Things to know**: the intended pattern and its pitfalls.
   - **API**: exact function signatures, Options and state types with field comments, and related types. Types documented in other references are named at the end of this section.
   - **Documentation excerpts**: the library documentation for the component (defaults, keyboard behavior, style parts, limits). Links go to other component references.
   - **Complete example**: a runnable program. Copy its structure: where state lives, how the component is bound, and which callbacks it uses.

   For a long reference, read Usage, Things to know and API first, then search the excerpts for the option you are using.
3. **Wire the app correctly** with `references/app-patterns.md`: window setup, model and state ownership, keys in lists, callbacks versus builders, background work, overlays, styling and i18n.
4. **Theming and per-part styling:** `references/components/theme.md` (palettes, `WithComponent`, `Styles`, `Slot`, `PartStyle`, override order) and `references/components/app.md`.

## Setup and versions

- Add the library with `go get github.com/josexy/nuxt-ui-go@latest`, import it as `nui "github.com/josexy/nuxt-ui-go"` together with `github.com/egoist/mygo` and `github.com/egoist/mygo/ui`, and build and run with `CGO_ENABLED=0`. A desktop graphics session is required to open windows.
- Each reference has a **Version** line naming the nuxt-ui-go and MyGo versions it describes. If the app's `go.mod` pins other versions, confirm signatures and defaults with `go doc github.com/josexy/nuxt-ui-go <Name>` before relying on them. When a reference and `go doc` disagree, `go doc` wins.

## Rules that prevent most bugs

Every component call has the same shape:

```go
nui.Input(c.Key("name"), theme, &app.Name, nui.InputOptions{Label: "Name", Clearable: true}).FillWidth()
```

- **Theme:** keep `config := nui.DefaultApp()` (or a customized `nui.AppConfig`) in the app model and call `theme := nui.App(c, config)` at the top of every build. Do not rebuild the config inside the view, and do not write the returned theme back into the config.
- **State:** bound values (`&app.Name`), open flags, and state objects (`TableState`, `MenuState`, `FormState`, `SearchState`, `SelectMenuState`, …) are fields of a model that outlives the view. Use one per window and per component instance, and never copy one after a component has used it. A local variable inside the view resets on every build.
- **Identity:** call `c.Key(stableID)` for each widget. In loops, use business IDs (never indexes, filtered positions or editable labels), and bind through the slice element (`&app.Items[i].Done`), not through a copy made by `range`.
- **Build lifetime:** `ui.Element` values, parts and `*ui.Context` are valid only during the current build. Never store them or capture them in goroutines. For long-lived window services, keep `c.Services()`.
- **Side effects:** change data, submit, start I/O and mutate collections in action callbacks (`OnClick`, `OnChange`, `OnSubmit`, `OnSelect`, …). Builders (body/content callbacks, `Style`, `Build`, `Cell`, `Footer`) run many times per interaction and must only build UI.
- **Background work:** run slow work in a goroutine, then publish results on the UI thread with `window.Update(func() { ... })`. Cancel stale work, or reject its results with a ticket or sequence number. The library has helpers: `FormState.BeginAsync`/`CompleteAsync`, `QueryState`, `CommandIndex` and FileUpload tickets.
- **Unset versus zero:** `Optional[T]` fields distinguish "use the default" (omit the field) from an explicit value (`nui.Set(false)`, `nui.Set(0)`). Check the field's type in the reference. Writing `Dismissible: false` does not compile, and leaving an Optional unset keeps the default (often `true`).
- **Overlays are controlled:** Modal, Slideover, Drawer, Popover and the menus take an app-owned open flag or state. Call them on every build, even while closed, and never wrap the call in `if open`; they need those builds to clean up and restore focus. Toasts need a `nui.ToastViewport` in the view.
- **Icons and accessibility:** icons come from `github.com/josexy/nuxt-ui-go/icons/lucide` (`lucide.Check()`). An icon-only button without `AccessibleLabel` panics, and avatar-only buttons need one too. Keep `Label` set even on items with a custom builder, so search and assistive technology still work.
- **List IDs:** items in Tabs, Accordion, Breadcrumb, Stepper, Timeline, menus and option lists need non-empty, unique IDs or values (menus: across the whole tree), or the component panics. Display labels may repeat.
- **Text:** the library translates its own messages (Clear, Close, empty states, …) in `en` and `zh-Hans` once the app calls `mygo.App.SetLanguages(nui.SupportedLocales()...)`. Text the app passes in (Label, Title, Placeholder, …) is shown as is.
- **Sizes** are in DIP. Colors and spacing come from the theme. Do not hard-code pixel scaling or light/dark colors.

## Component index

Reference file: `references/components/<id>.md`, where the id is the component name in kebab case.

| Category | Components → reference id |
| --- | --- |
| Themes and basics | App `app` · Theme, palettes, styling `theme` · Container `container` · Main `main` · Icon `icon` · Button `button` · Link `link` · Badge `badge` · Card `card` · Alert `alert` · Avatar `avatar` · AvatarGroup `avatar-group` · Chip `chip` · Separator `separator` · Kbd `kbd` · Skeleton `skeleton` · Progress `progress` · ProgressGroup `progress-group` · Empty `empty` · User `user` |
| Inputs and selection | Input `input` · Textarea `textarea` · FormField `form-field` · FieldGroup `field-group` · Checkbox `checkbox` · Switch `switch` · CheckboxGroup `checkbox-group` · RadioGroup `radio-group` · Slider `slider` · Select `select` · InputMenu `input-menu` · Popover `popover` · ScrollArea `scroll-area` |
| Overlays and navigation | Tooltip `tooltip` · Modal `modal` · Slideover `slideover` · Toast `toast` · DropdownMenu `dropdown-menu` · ContextMenu `context-menu` · Tabs `tabs` · Collapsible `collapsible` · Accordion `accordion` · Breadcrumb `breadcrumb` · Pagination `pagination` · Stepper `stepper` · Timeline `timeline` · Banner `banner` · ErrorView `error` |
| Forms and data | Form `form` · InputNumber `input-number` · InputTags `input-tags` · PinInput `pin-input` · InputRating `input-rating` · Listbox `listbox` · SelectMenu `select-menu` · Calendar `calendar` · InputDate `input-date` · InputTime `input-time` · ColorPicker `color-picker` · Table `table` · Tree `tree` |
| Advanced and layout | CommandPalette `command-palette` · NavigationMenu `navigation-menu` · Drawer `drawer` · FileUpload `file-upload` · Chart `chart` · Carousel `carousel` · Marquee `marquee` · Splitter `splitter` · Sidebar `sidebar` · HeaderLayout `header` · FooterLayout `footer` · FooterColumns `footer-columns` |

Functions documented inside another component's reference: `CheckboxTri` → `checkbox`, `RangeSlider` → `slider`, `SelectMultiple` → `select`, `SearchSelect` → `input-menu`, `InputDateRange` → `input-date`, `ModalWithOptions` → `modal`, `ToastViewport` and `CloseToast` → `toast`, `ParseNumber` → `input-number`, `AddTags` → `input-tags`, `Initials` → `avatar`, `FormatKey` → `kbd`, `Window`, `Themed` and `SupportedLocales` → `app`, palettes and `NewTheme` → `theme`.

## Choosing between similar components

| Need | Use |
| --- | --- |
| Pick one or several values from a list after clicking | `select` (`Select` binds `*T`, `SelectMultiple` binds `*[]T`) |
| Type with suggestions, or accept free text | `input-menu` (binds `*SearchState`) |
| A trigger with a search field in the popup, for large or async lists | `select-menu` (builds on `listbox`) |
| An always-visible selectable list | `listbox` |
| A single on/off value (`*bool`) | `checkbox` or `switch`. Tri-state: `CheckboxTri` in `checkbox` |
| Several values from visible options (`*[]T`), or exactly one (`*T`) | `checkbox-group`, `radio-group` |
| Numbers, ranges, stars, codes, tags | `input-number`, `slider` (`RangeSlider`), `input-rating`, `pin-input`, `input-tags` |
| Dates and times | Click a date: `calendar`. Type a date or range: `input-date`. Type a time: `input-time` |
| Label, help text and error for a control; validation and submission | `form-field` around each control, `form` around the set |
| Visually join controls (an input plus a button) | `field-group` |
| Blocking dialog / side panel / bottom sheet with drag handle | `modal` / `slideover` / `drawer` |
| Content anchored to a control / hover hint | `popover` / `tooltip` |
| Action menu from a button / right-click menu / Ctrl+K command search | `dropdown-menu` / `context-menu` / `command-palette` |
| Transient notice / inline message / page-wide bar / error page / nothing to show | `toast` / `alert` / `banner` / `error` / `empty` |
| Switch between views (`*string`) / one expandable region (`*bool`) / expandable list (`*[]string`) | `tabs` / `collapsible` / `accordion` |
| App navigation and its frame | `navigation-menu` inside `sidebar`; `header`, `footer`, `footer-columns`; two resizable panes: `splitter` |
| Location, pages, steps, history | `breadcrumb`, `pagination`, `stepper`, `timeline` |
| Rows and columns / hierarchy / charts | `table` / `tree` / `chart` |

Some Nuxt UI features have no native equivalent: the web mega menu, CSS breakpoints, DOM slots, NuxtLink, and the Page, Dashboard, AI Chat, Editor, Content, color mode and LocaleSelect components. Call the theme and language APIs directly, build the rest from the components above, and do not invent options for them.

