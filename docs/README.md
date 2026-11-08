# Component guide

If you are new to NuxtUI Go, first follow [Getting started](../README.md#getting-started) to open a window. After that, use this guide to find what your current task needs; you do not have to read every component in order.

## Start from your task

| What you want to do | Read |
| --- | --- |
| Set the window theme, or add buttons, cards, or status messages | [Themes and basic components](foundations.md) |
| Accept text, or add labels, checkboxes, or dropdown selects | [Inputs, selection, and fields](inputs.md) |
| Open dialogs or menus, switch views, or show notifications | [Overlays, menus, and navigation](overlays-navigation.md) |
| Validate forms, handle dates, or display tables and trees | [Forms, dates, and data components](forms-data.md) |
| Add a sidebar, command palette, file picker, or chart | [Advanced interaction, charts, and layout](advanced.md) |

Once you know which component you need, look up its common options in [Options and defaults](options.md). The [bilingual component website](../website/README.md) also offers per-component option editors, Go/WASM previews, and complete programs you can copy.

## How component calls work

Most components take the same kinds of arguments: the current UI context, the theme, the data to edit, and an options struct. The code below runs inside a view function; `app.Name` is a string field that the app keeps between builds:

```go
nui.Input(c.Key("name"), theme, &app.Name, nui.InputOptions{
    Label: "Name",
    Placeholder: "Enter your name",
    Clearable: true,
}).FillWidth()
```

| Code | Meaning |
| --- | --- |
| `c *ui.Context` | The context for the window's current build; MyGo passes it in |
| `theme nui.Theme` | The current theme, usually returned by `nui.App(c, config)` |
| `c.Key("name")` | A stable identity that lets MyGo find the same control across rebuilds |
| `&app.Name` | A pointer to app-owned data; the component updates it when the user types |
| `nui.InputOptions{...}` | The component's options; `FillWidth()` on the result makes the control fill the available width |

Throughout the docs, `c`, `theme`, and `t` have these meanings. `app` and `state` stand in for your own app model, so replace them with your variables. Code blocks without `package main` are fragments; see the [minimal example](../examples/minimal/main.go) for the full imports and startup code. Icon examples use the `lucide` package; see the [icon docs](../icons/lucide/README.md) for the import.

### Where to keep state

MyGo calls the view function repeatedly: during input, after data changes, and while animating. **Keep input values and component state outside the view function**, or every build resets them.

Your app can keep business data, state objects such as `AppConfig`, `SearchState`, and `FormState`, and `ui.Handle` values used for focus and geometry queries. Each window should keep its own editing and interaction state. Do not copy a component state value after a component has used it.

The `ui.Element` values and parts returned by components are valid only during the current build. You can call methods such as `FillWidth()` on them right away or pass them to a Tooltip or Popover in the same build, but do not store them in your model for later builds. The parent scope of `*ui.Context` changes during a build, so do not store the context in your model or background tasks either. To get window services you need to keep, call `c.Services()`.

In dynamic lists, call `c.Key(businessID)` before creating each control. The ID must be stable and unique; do not use indexes that change with sorting or display names that users can edit.

### Changing data and running work

Handle clicks, input changes, and form submission in callbacks such as `OnClick`, `OnChange`, and `OnSubmit`. These callbacks are the right place to change data and start business operations.

Build callbacks such as `Build`, `Content`, and `Style` may run many times. Use them only to create or configure the current UI. Never submit a form, change a collection you are iterating over, or send a request from them unconditionally.

Run slow computation and I/O in a background task. Return the result to the UI thread with `Window.Update` before writing it to the model. Cancel the task when the user leaves the view or closes the window. Components that use async tickets also discard stale results; see [Form and QueryState](forms-data.md).

### Unset versus explicit values

Some options have the type `Optional[T]` so the component can tell the default behavior apart from an explicit value. For example, to turn off close behavior that is on by default:

```go
options := nui.ModalOptions{
    Dismissible: nui.Set(false),
    CloseButton: nui.Set(false),
}
```

`Optional[T]{}` means unset. `nui.Set(false)` and `nui.Set(0)` mean "use `false`" and "use `0`". Plain `bool`, numeric, and enum fields are assigned directly without `Set`. See [Unset, zero, and default values](options.md#unset-zero-and-default-values).

## Terminology

| Term | Meaning in this library |
| --- | --- |
| Part (slot) | A region of a component that can be styled on its own, such as the outer `Root`, the `Label` text, or a popup `Panel`. Available parts vary by component |
| Controlled state | A value the app owns and can change directly, such as `open *bool`, the current page, or a selected ID |
| Builder | A function called during a build to supply content such as a body, a table cell, or other custom UI |
| DIP | Device-independent pixel, a logical size unit the system maps to screen pixels. You normally do not multiply by the display scale yourself |
| LTR / RTL | Left-to-right and right-to-left writing direction. Components that support direction adjust their layout or key handling |

## Component catalog

The catalog lists 73 components and layout concepts by name. To see a full function signature, run `go doc` in the repository, for example `go doc . Input`.

### Themes and basic components

[Themes and basic components](foundations.md): App, Theme, Container, Main, Icon, Button, Link, Badge, Card, Alert, Avatar, AvatarGroup, Chip, Separator, Kbd, Skeleton, Progress, ProgressGroup, Empty, User.

### Inputs and selection

[Inputs, selection, and fields](inputs.md): Input, Textarea, FormField, FieldGroup, Checkbox, CheckboxGroup, RadioGroup, Switch, Slider, Select, InputMenu, Popover, ScrollArea.

### Overlays and navigation

[Overlays, menus, and navigation](overlays-navigation.md): Tooltip, Modal, Slideover, Toast, DropdownMenu, ContextMenu, Tabs, Collapsible, Accordion, Breadcrumb, Pagination, Stepper, Timeline, Banner, ErrorView.

### Forms and data

[Forms, dates, and data components](forms-data.md): Form, InputNumber, InputTags, PinInput, InputRating, Listbox, SelectMenu, Calendar, InputDate, InputTime, ColorPicker, Table, Tree.

### Advanced interaction and layout

[Advanced interaction, charts, and layout](advanced.md): CommandPalette, NavigationMenu, Drawer, FileUpload, Chart, Carousel, Marquee, Splitter, Sidebar, HeaderLayout, FooterLayout, FooterColumns.

In Go, `Error` is already the name of a semantic color, so the error display component is called `ErrorView`. `Header` and `Footer` are part names, so the layout functions are called `HeaderLayout` and `FooterLayout`. Related functions such as `CheckboxTri`, `RangeSlider`, `SelectMultiple`, and `InputDateRange` are covered in the sections for their base components.

## Other references

| What you need | Document |
| --- | --- |
| Switching between English and Chinese, and overriding messages | [Internationalization](i18n.md) |
| Lucide icon names and usage | [Icon docs](../icons/lucide/README.md) |
| Platform support, differences from Nuxt UI, IME, and accessibility limits | [Compatibility and known limitations](compatibility.md) |
| Runnable examples and development checks | [Run the examples](../README.md#run-the-examples), [Taskfile.yml](../Taskfile.yml) |
| Version changes | [Changelog](../CHANGELOG.md) |
