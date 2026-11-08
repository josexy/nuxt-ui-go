# Options and defaults

Use this page to look up common options. When you use a component for the first time, start with the examples in the [component guide](README.md), then come back here for defaults and behavior switches.

| What you want to change | Where to look |
| --- | --- |
| Turn off a feature that is on by default, or set a value to 0 | [Unset, zero, and default values](#unset-zero-and-default-values) |
| Horizontal or vertical layout | [Orientation](#orientation) |
| Multiple selection, search, query reset, and pagination | [Selection and pagination](#selection-and-pagination), [InputMenu](#inputmenu-values-and-search) |
| Loading, tags, overlays, sidebars, or upload areas | [Common options by component](#common-options-by-component) |
| Keyboard and assistive technology limits | [Limitations](#limitations) |

## Unset, zero, and default values

A Go zero value is not always the value a component ends up using. For example, Pagination treats `PageSize: 0` as the default of 10. Fields that must distinguish "unset" from an explicit `false` or `0` use `Optional[T]`.

| What you write | Meaning |
| --- | --- |
| Omit the Optional field, or write `nui.Optional[T]{}` | Use the component's default behavior |
| `nui.Set(false)` | Explicitly turn the option off |
| `nui.Set(0)` | Explicitly use 0; the field's range limits still apply |
| Plain bool, numeric, or enum field | Assign directly; check the component's notes for how it treats the zero value |

For example, to hide the Toast close button, write `Close: nui.Set(false)`. To show no sibling pages in Pagination, write `SiblingCount: nui.Set(0)`. Writing only `Siblings: 0` uses the default of 2.

### Frequently checked defaults

| Component | Defaults |
| --- | --- |
| Pagination | PageSize=10, Siblings=2, ShowEdges=false, ShowControls=true, Size=MD |
| InputTags, InputRating, Slider | InputTags: AddOnPaste=false. InputRating: Hoverable=false. Slider: Step=1 |
| SelectMenu | SearchInput=true. The search term is reset both after a selection and after closing |
| Carousel, FileUpload | Carousel: Arrows=false. FileUpload: Layout=UploadGrid |
| Sidebar | Resizable=false; Collapsible, Toggle, and AutoClose=true. Expanded and collapsed widths are 256 and 64 DIP. Description is empty. Actions appear only when expanded or in the mobile menu |
| Header, Footer, FooterColumns | Header Title is empty. Footer has no top border. FooterColumns uses a preferred minimum column width of 160 DIP and a column gap of 32 DIP |
| Overlays and notifications | Popover.Dismissible=true. Toast.Close and Toast.Progress=true |
| Listbox, Accordion | Listbox: HighlightOnHover=true. Accordion: Collapsible=true |
| Calendar | Color=Primary, Variant=Solid, Size=MD. FixedWeeks, MonthControls, YearControls, and ViewControl=true. NumberOfMonths=1, YearsPerPage=12 |
| ColorPicker | Size=MD, with a 168 DIP SV area. An unset Throttle means 50 ms; an explicit 0 disables throttling. The bound model updates immediately |
| InputDate, InputDateRange | ShowCalendar=false, and HideCalendar wins when both are set. Each DateStep field treats 0 as 1. The range separator defaults to a minus sign |
| InputTime | An unset TwelveHour follows the locale's clock. An unset Granularity is derived from Seconds and defaults to minutes. Each TimeStep field treats 0 as 1. StepSnapping=false |

FileUpload's `Dropzone`, `Interactive`, `FileDelete`, and `Preview` all default to true. The zero value of `UploadLayout` is `UploadLayoutDefault`; use the enum names in options rather than relying on the underlying numbers.

## Orientation

Use `nui.Horizontal` for horizontal and `nui.Vertical` for vertical layout. Components that default to vertical usually take an `Optional[Orientation]`, which you write as `nui.Set(nui.Horizontal)`. Plain `Orientation` fields are assigned directly.

| Options | Field type and default | What changes |
| --- | --- | --- |
| FormFieldOptions | Optional; Vertical | The label sits above or beside the input. Label association and editor state are kept |
| AlertOptions / ToastOptions | Optional; Vertical | The actions sit below or beside the body |
| ListboxOptions | Optional; Vertical | A vertical virtual list, or a horizontal one with a fixed ItemWidth. Arrow keys follow the axis |
| InputNumberOptions | Orientation; Horizontal | Minus, value, and plus in a row, or increment and decrement buttons stacked to the right of the value |
| InputRatingOptions / RangeSliderOptions | Orientation; Horizontal | Layout and interaction direction of the stars or slider. A range slider keeps two separate focus targets and values |
| ProgressGroupOptions | Orientation; Horizontal | A horizontal or vertical segmented track. Vertical Height defaults to 160 DIP |
| CarouselOptions | Orientation; Horizontal | Layout, dragging, and arrow keys switch together. Switching cancels any drag in progress |

```go
nui.FormField(c, t, nui.FormFieldOptions{
    Label: "Workspace",
    Orientation: nui.Set(nui.Horizontal),
}, func(inner nui.Theme) {
    nui.Input(c, inner, &state.Name, nui.InputOptions{})
})

nui.RangeSlider(c, t, &state.Low, &state.High, nui.RangeSliderOptions{
    Label: "Budget",
    Orientation: nui.Vertical,
    Min: 0, Max: 100, Step: 5, MinStepsBetweenThumbs: 2,
})
```

Some components also have a `Vertical` or `Horizontal` bool field; check the component's Options type. Horizontal interaction in components that support RTL follows the writing direction.

## Selection and pagination

### Single selection, multiple selection, and reselecting

| Component or option | Behavior |
| --- | --- |
| `Select(*T)` / `SelectMultiple(*[]T)` | Bind a single value or a slice of values. Both share SelectOptions, stable IDs, custom Option.Build, and selected-item tags |
| `SelectionToggle` | Listbox's default policy. In both single and multiple selection, activating the current item again deselects it |
| `SelectionReplace` | Activating an item replaces the selection. SelectMenu always uses this policy for single selection |
| Multiple selection with replace | Shift+arrow keys, Home, and End extend a contiguous range and skip disabled items. Ctrl/Cmd+A selects all available items |
| `CheckboxGroup.Max` | Once the limit is reached, unchecked items can still receive focus but refuse to be checked, and report a read-only state to assistive technology |

Single and multiple Select both inherit Color, Variant, Size, and UI from ComponentSelect; instance options take precedence. Listbox also supports Loop and HighlightOnHover.

### SelectMenu search and reset

`SearchInput` defaults to true. Setting it to false hides the search field, but any retained Query still filters the list. ExternalFilter or IgnoreFilter skips local filtering. ↑ or ↓ on the trigger opens the menu and focuses the search field if there is one; with FocusSearch=false, the list gets focus instead. A disabled trigger cannot open the menu.

By default, Query is cleared after a selection and after closing. `ResetSearchTermOnSelect` and `ResetSearchTermOnBlur` control these two cases. Closing includes Escape, clicking outside, and the app changing the open state.

Selecting the current value again still closes the menu, resets the query, and fires OnSelect. OnChange fires only when the selected value actually changes. Invalid and ErrorMessage update both the field's error semantics and its border.

Clearable shows a clear button, configured with ClearOptions and ClearIcon. The button always clears the query. ResetModelValueOnClear defaults to true; set it to false to keep the selected value. OnClear fires once after the reset. OnOpenChange reports only actual open/close changes caused by the user.

Listbox configures built-in filtering with Filter and FilterInput; SelectMenu uses SearchInput and FilterInput. Both support a SearchTerm pointer, FilterFields, IgnoreFilter, and OnSearchTermChange. When SearchTerm is unset, each uses its own State.Query. Assigning the term from outside does not fire a search notification.

### Pagination siblings and appearance

`SiblingCount` takes precedence over the older `Siblings` and accepts 0–10. For example, on page 5 of 10 with zero siblings:

| Setting | Display |
| --- | --- |
| ShowEdges=true | `1 … 5 … 10` |
| ShowEdges=false | `5` |

Near the first or last page, the range of edge pages widens. When there are only a few pages, no ellipsis is shown. ShowControls controls the previous/next buttons and ShowEdges controls the first and last page numbers. Size, Color, and Variant style the ordinary page buttons; ActiveColor and ActiveVariant style the current page.

## Common options by component

Table's Sortable, ManualSorting, and List.SortBy control, respectively, whether a column can be sorted, whether the app sorts the data, and multi-column sorting. Loading keeps the existing rows and shows a loading bar; BuildLoading replaces the empty content only while an empty table is loading. Caption provides a description and BuildCaption a visible title. For sort, selection, and activation notifications and native limits, see [Table](forms-data.md#table).

Tree's OnSelect and OnToggle can cancel an operation; OnChange and OnExpandedChange report actual changes to selection and expansion. TreeNode.UI and Style style an individual node's parts, and BuildLeading and BuildTrailing replace the non-interactive leading and trailing content. For indeterminate states with multiple selection, lazy loading, and empty branches, see [Tree](forms-data.md#tree).

FormOptions.LoadingAuto defaults to true: while an async operation runs, form controls are disabled and a snapshot of the fields is kept. Set it to `nui.Set(false)` to allow editing during the operation, which invalidates the earlier result. OnError is called once when submit validation fails, including global errors. For submit, cancel, and reset rules, see [Form](forms-data.md#form).

### Switches, tags, and ratings

| Component | Common options and behavior |
| --- | --- |
| Switch | Loading blocks toggling by mouse and keyboard. LoadingIcon, CheckedIcon, and UncheckedIcon set the icons. The spinner stops when reduced motion is on |
| Checkbox / CheckboxGroup / RadioGroup | Variant is ChoiceList, ChoiceCard, or ChoiceTable. Indicator is IndicatorStart, IndicatorEnd, or IndicatorHidden. Highlight emphasizes the border or indicator, even when unchecked |
| InputTags | AddOnTab, AddOnBlur, and AddOnPaste control when tags are added. Duplicate, ConvertValue, and DisplayValue control values and display. MaxLength limits the text being edited |
| PinInput | Color, Variant, and Size set the appearance. `Separator: []int{3}` adds a group break after the third cell; the control still uses one editor and one Tab stop |
| InputRating | Clearable lets clicking the current rating or pressing Space reset it to zero. Hoverable defaults to false. An unset EmptyIcon reuses Icon. Indicator, IconSlot, and EmptyIconSlot style the fill layer and the two icons |

Checkbox uses only CheckedIcon; it does not use Switch's UncheckedIcon. A ChoiceCard is tinted when selected, so the selected state stays visible even with the indicator hidden.

Checkbox's Loading, LoadingIcon, and UncheckedIcon, and RadioGroup's Max, come from shared option types and have no effect on those components; they remain for compatibility with existing code. ChoiceTable on a single Checkbox and Variant/Indicator on Switch are Go-native extensions. The website's option editor and API tables state where each field applies.

Option.Icon works in CheckboxGroup and RadioGroup. With the indicator hidden, the icon appears above the item label; an ordinary Radio indicator is still a dot. Loop controls whether arrow keys wrap around; it defaults to false for checkbox groups and true for radio groups, and `nui.Set(false)` turns wrapping off. Both group types use a single Tab stop, with Left/Right arrows in horizontal layout and Up/Down arrows in vertical layout.

InputTags' MaxLength limits what the user can insert, counted in UTF-16 code units; 0 means no limit. It limits the draft being edited and does not validate tag length on submit. With AddOnPaste on, pasted text is always handled as tags; text rejected as a duplicate or over the limit is not inserted as plain text either. With it off, pasting inserts plain text. With AddOnTab on, Tab and Shift+Tab commit a non-empty draft and keep focus; with an empty draft they move focus as usual.

InputNumber's StepSnapping, FocusOnChange, Increment, and Decrement default to true, and the mouse wheel is enabled by default. The first increment or decrement from an empty value uses StartingValue instead of adding an extra step. PinInput.Highlight defaults to false; SeparatorSlot and Root.Gap/Padding adjust the separator and spacing. See [Numbers, tags, PINs, and ratings](forms-data.md#numbers-tags-pins-and-ratings) for full details.

InputRating's horizontal arrow keys follow RTL; in vertical layout, use ↑ and ↓, and the other axis's keys do nothing. The rating stops at both ends. Home and PageUp go to 1, End and PageDown go to Max. The hover preview disappears as soon as the rating is cleared.

### Sliders, tabs, and accordions

| Component | Common options and behavior |
| --- | --- |
| Slider / RangeSlider | PageUp/PageDown and Shift+arrow keys move by 10 steps. OnChange reports live changes; OnCommit fires when a key press or drag that changed the value ends |
| RangeSlider | The thumbs can cross; they are then reordered and active focus moves with them. Moves that would bring them closer than `Step × MinStepsBetweenThumbs` are rejected. The tooltip shows each value separately |
| Tabs | ActivationMode defaults to ActivationAutomatic. With ActivationManual, arrow keys only move focus, and Enter or Space activates a tab and fires a single notification |
| Accordion | Collapsible defaults to true. In single mode, setting it to false prevents collapsing the open item by clicking it again, although another item can still be opened. Multiple mode ignores this option |
| Collapsible / Accordion | Disabled disables only the header. KeepMounted defaults to false; true keeps hidden controls mounted. TrailingIcon defaults to ChevronDown, and an Accordion item's own icon takes precedence. AccordionItem.UI overrides that item's style |
| Carousel | Arrows shows the previous/next buttons and Dots enables dot navigation. The dots' page names follow the current language |

### Overlays and notifications

| Component | Common options and behavior |
| --- | --- |
| Popover | Mode defaults to PopoverClick and can be PopoverHover. OpenDelay and CloseDelay apply only in hover mode; they are `time.Duration` values that default to 0 |
| Tooltip | Placement and Offset set the position. Open can bind the open state. Kbds displays shortcut keys and adds them to the anchor's description. DelayDuration sets the hover delay explicitly; 0 shows the tooltip immediately |
| Modal / Slideover | CloseButton controls whether the close button appears; Close and CloseIcon configure it. Transition=false skips the transition. OnClosePrevent reports close attempts blocked by Dismissible=false |
| Slideover | Inset adds edge spacing and rounded corners. Extent sets only the size along the relevant axis; Width, MaxHeight, and Fullscreen do not affect the layout |
| Toast | Close and Progress control visibility. CloseOptions, CloseIcon, Avatar, and ProgressOptions configure appearance. Stable IDs on Actions keep focus when actions are reordered. Close=false on an action keeps the toast open |
| DropdownMenu / ContextMenu | Size defaults to MD and Loop to false. Filter and SearchTerm configure filtering; MenuItem.Content defines a submenu. Close=false keeps the menu open after a leaf item is selected. OnCheckedChange reports value changes and OnOpenChange reports open/close changes |
| Tabs | Content defaults to true; false unmounts all panels. KeepMounted keeps only the inactive panels mounted. TabItem supports Avatar, Badge/BadgeOptions, and UI. Indicator styles the active background or underline |

Popover's Dismissible defaults to true and controls closing with Escape or an outside click; the app can always change `open` directly. In hover mode, the pointer can cross the gap between the anchor and the panel. After Escape closes the popover, leaving the pointer where it is does not reopen it right away. The website's option editor takes delays in milliseconds.

Even when Open is bound, Tooltip still responds to the hover delay, keyboard focus, blur, clicks, and Escape. When the app sets Open to false, the tooltip does not reopen just because the pointer or focus is still on the anchor. A new tooltip closes any previous controlled tooltip.

By default, running a Toast action closes the toast; set `Close: nui.Set(false)` on that action to keep it. While hovering or focus pauses the toast, its countdown bar also stops. Persistent toasts have no countdown bar.

### Sidebar, navigation, and file selection

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

Clicking an interactive UploadArea, or pressing Enter or Space on it, opens the file picker; its accessible name comes from Label first. In the built-in list, file actions are separate from the title/icon entry point so native assistive technology can reach each action. The Choose button appears only with UploadButton or `Interactive: nui.Set(false)`. Dropzone=false removes the drop target. Preview=false hides the whole list, and FileImage=false hides only the app-supplied preview accessories; the Files data is kept in both cases. OnChange fires when a selection or removal changes the queue, and OnReject reports newly rejected files. After a removal, focus returns to the selection entry point. For full defaults and responsibilities, see [FileUpload](advanced.md#fileupload-and-resource-ownership). For Sidebar's region builders, sizes, and keyboard controls, see [Sidebar state and size](advanced.md#sidebar-state-and-size).

## InputMenu values and search

### Choose what to store

| Mode | State fields |
| --- | --- |
| Single selection | `SearchState.Selected` stores the selected option ID; Query stores the edit draft |
| Multiple selection | `SelectedValues` stores every selected ID. Remove items with their delete buttons or with Backspace on an empty draft |
| InputMenuAutocomplete | Query is free text. Selected keeps the last selected ID. Forces single selection and allows submitting custom text |

OnValuesChange reports changes to multiple selection, and OnSelect reports the ID that was just activated.

### Local or external search

By default, the search is a case-insensitive substring match on Label. FilterFields can list `label`, `description`, and `value`; nil means label only. The older SearchSelect entry point searches label and description.

`ExternalFilter=true` shows the results the app passes in without filtering them. QueryState or an app service handles requests, cancellation, and stale results. A change to an option's identity, label, description, or disabled state closes the open panel; ExternalFilter does not change this rule and does not keep the panel open automatically.

### Creating options

When you provide OnCreate, the list can show a create row that is reachable from the keyboard. By default it appears only when nothing matches; CreateWhenAlways shows it whenever the query is non-empty and no option has the same value.

Creating only sends a request: it keeps the query, filter, and open state, does not trigger an ordinary selection, and does not add the option for you. Update the data source after the app finishes creating the option.

### Opening and resetting

OpenOnClick and OpenOnFocus both default to false. Typing, arrow keys, and the trailing button still open the menu. To open it on click, set `OpenOnClick: nui.Set(true)`. SearchSelect keeps its older open-on-click behavior. Open can bind a persistent bool. Loop defaults to false and HighlightOnHover to true.

ResetSearchTermOnSelect and ResetSearchTermOnBlur both default to true and control the draft after a selection and after closing. Single selection restores the selected label; multiple selection clears the draft. Autocomplete ignores Multiple, ResetSearchTermOnSelect, and ResetModelValueOnClear: selecting a suggestion always fills in its full text, and an ordinary blur keeps the free text.

ResetModelValueOnClear controls only the clear button and defaults to true; false keeps the selected value. It does not affect deleting ordinary text with the keyboard. In multiple selection, you can clear all tags from an empty draft or remove them one by one with the DeleteIcon buttons; each actual change to the selected values fires OnValuesChange once. After a tag is deleted with the keyboard, focus returns to the input.

## Examples

The Gallery's **Props & behavior** page demonstrates loading, selection limits, adding tags, manual Tabs, search reset, pagination, overlays, and file previews. The **Orientation** page demonstrates orientation switching. See [Run the examples](../README.md#run-the-examples) for the command.

## Limitations

Listbox builds only the visible items plus a few extra in either orientation. Horizontal items and group headers use a fixed ItemWidth; in vertical layout, the native List manages row heights.

Slider and RangeSlider expose their value, range, and step to assistive technology, but MyGo has no API for setting vertical orientation metadata. This affects every vertical mode, including sliders without Inverted. Slider tooltips can be positioned per thumb, but they have no separate delay setting. InputRating uses a single-focus slider role.

FormField associates its label with this library's controls. If you use a custom MyGo control, your app must associate the label itself. Toast action callbacks and ButtonOptions.Style are stored with the toast, so their closures must not capture temporary Elements, Parts, or Contexts; the Style arguments are valid only during the build in which they are called.

Go Options fields and types follow this library's API; you cannot copy Nuxt Vue props directly. For date and time operations and migration, see [Calendar, InputDate, and InputTime](forms-data.md#calendar-inputdate-and-inputtime). For platform, IME, and screen reader verification, see [Compatibility and known limitations](compatibility.md).
