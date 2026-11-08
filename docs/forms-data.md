# Forms, dates, and data components

This page covers forms with validation, specialized inputs, dates and times, and lists, tables, and trees. For ordinary text input, field labels, and dropdown selects, see [Inputs, selection, and fields](inputs.md).

| Task | Component |
| --- | --- |
| Validate several fields and submit them | [Form](#form) |
| Enter numbers, tags, verification codes, or ratings | [Specialized inputs](#numbers-tags-pins-and-ratings) |
| Search and select list items, including remote search | [Listbox, SelectMenu, QueryState](#listbox-selectmenu-and-async-search) |
| Enter a date, time, or color | [Date and time components](#calendar-inputdate-and-inputtime), [ColorPicker](#colorpicker) |
| Display a sortable table or hierarchical data | [Table](#table), [Tree](#tree) |

## Lifecycle and themes

Keep business values and component state per window. Do not reinitialize them on every build, and do not copy state that a component has already used. A returned `ui.Element` is valid only during the current build; do not put Elements or Contexts in your app model or background tasks.

FormTicket, QueryTicket, and TreeTicket identify async operations and carry the instance identity and a version. Return each result to the state object that started the request; old results are rejected if the data has changed or the request was canceled. Background work should operate only on data snapshots and return to the UI thread through `Window.Update`. Cancel business tasks when the window closes.

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. Other components are configured through their own Options; see the [style part index](#style-part-index) at the end of this page.

## Form

### Register fields and submit

Form manages field validation, error display, and focusing the first field with an error. FormState stores each field's initial value, whether it has changed (Dirty), whether the user has interacted with it (Touched), its visibility, its error, and its Handle.

Register a field through FormField's Name and Value. Within one form build, Name must be non-empty and unique, and Value is a string your app converts explicitly. Each field corresponds to one edited value.

```go
// form nui.FormState and name string are stored in the window model.
nui.Form(c, theme, &form, nui.FormOptions{
    Validate: func(values nui.FormValues) nui.FormErrors {
        if values["name"] == "" {
            return nui.FormErrors{"name": "Enter a name"}
        }
        return nil
    },
    OnSubmit: func(values nui.FormValues) { save(values["name"]) },
}, func(inner nui.Theme) {
    nui.FormField(c, inner, nui.FormFieldOptions{
        Name: "name",
        Value: name,
        Label: "Name",
        Required: true,
    }, func(field nui.Theme) {
        nui.Input(c, field, &name, nui.InputOptions{})
    })
    nui.Button(c, inner, "Save", nui.ButtonOptions{}).OnClick(form.Submit)
})
```

FormValues and FormErrors are both named `map[string]string` types. The form does not use reflection to derive a structure; cross-field validation reads the complete Values directly. In this example, your app provides `save`; run slow saves in the background.

Errors for empty names or names that have not been registered also block submission. Read them with `form.Errors()` to display global errors. An invalid submission calls `OnError` once and a valid one calls `OnSubmit` once; each callback receives its own snapshot. Explicit Validate, SetErrors, and CompleteAsync calls do not trigger either callback.

### Validating, hiding, and resetting

| Action or state | Behavior |
| --- | --- |
| First registration of a field | The current value becomes the initial value for Dirty comparisons |
| Blur or `Touch(name)` | Marks the field as Touched |
| `Submit()` / `Validate(...)` | Runs validation, marks the current fields as Touched, and focuses the first field with an error in build order |
| Hidden or disabled field | Keeps its model, Dirty, Touched, and error, but is excluded from Values, submit validation, and error focus |
| `Reset()` | Cancels pending requests, clears errors and interaction state, and uses the values from the next build as the new initial values. Does not change the data bound by your app |
| `Errors()` / `SetErrors(errors)` | Gets a separate snapshot of the errors / replaces the validation errors. Invalid edit drafts still take precedence and block submission |
| `ClearErrors(names...)` | Clears validation errors for the given names, or all errors if no names are passed. Does not make an invalid draft valid |

Invalid number, date, or other edit drafts also block submission, so the last valid value is never submitted by mistake. Business validation runs on Submit or an explicit Validate, not automatically on every keystroke.

When a hidden field reappears, its existing error is checked again. To restore the business's initial values, your app first restores the bound data and then resets the form state as needed; calling Reset alone does not restore the input contents.

Dirty is cleared for the fields that can currently be submitted when a synchronous OnSubmit returns without starting async work, or when CompleteAsync succeeds. Touched means the field has lost focus or taken part in validation; this is not exactly the same set of events as in Nuxt UI. Nested Forms manage their own fields and are not automatically linked to the parent form's values and validation; your app coordinates submissions across forms.

### Async validation

1. On the UI thread, call `ticket := form.BeginAsync()` to get a snapshot of the current fields.
2. The background task reads only `ticket.Values` and performs validation.
3. In a `Window.Update` callback, call `form.CompleteAsync(ticket, errors)`.
4. A return value of true means the result is still current and has no errors; your app then saves.

CompleteAsync does not call OnSubmit again. A change to a field value, the raw edit text, or visibility invalidates the old ticket, as do Cancel and a new request. `Cancel()` only makes the form reject old results; your app must still cancel the context of any actual I/O.

LoadingAuto defaults to true: after BeginAsync, inputs and buttons inside the form are disabled, and the field snapshot is kept. To allow editing while waiting, set `LoadingAuto: nui.Set(false)`; editing then invalidates the old ticket. Both modes ignore repeated Submit calls, and a submission made while disabled does not run after the form is re-enabled. Put cancel and finish buttons outside the Form.

Validate, SetErrors, ClearErrors, and Reset cancel the old ticket so an older result cannot overwrite newer validation or manually set errors. Call Cancel before unmounting the component or closing the window; the state object cannot detect every unmount path on its own. Background work may only use ticket.Values, and results must be received in Window.Update.

Root is the only style part, with a default vertical gap of 16 DIP; its font and text color are set before the body is built. There is no HTML form submission, schema conversion, automatic debounced validation, or Nuxt nested path handling.

Built-in inputs, selects, and single-value toggles are automatically associated with the field Handle; in a composite group, only the first control is associated. For a custom control, bind `form.Field(name).Handle` to the actual focusable control.

The settings example's **Project setup** demonstrates step-by-step registration, conditionally hidden fields, cross-field date and budget validation, error focus, and a cancelable submission. Its Confirm project button only simulates local confirmation and sends no network request.

## Numbers, tags, PINs, and ratings

### InputNumber

InputNumber binds an `*Optional[float64]` and keeps the input draft in a separate NumberState. This distinguishes an empty value, an incomplete input (such as a lone minus sign), and a valid number.

```go
// value nui.Optional[float64] and numberState nui.NumberState are stored per window.
nui.InputNumber(c, theme, &value, &numberState, nui.InputNumberOptions{
    InputOptions: nui.InputOptions{Label: "Quantity"},
    Orientation: nui.Vertical,
    Min: nui.Set(0.0),
    Max: nui.Set(100.0),
    Step: 1,
})
```

Step defaults to 1 and must be a finite positive number; 0 uses the default. StepSnapping defaults to true: for example, with Step=1, stepping up from 18.98 gives 19 and stepping down gives 18. With snapping off, the step is added or subtracted directly. Min is the origin of the step grid, or 0 when unset. The first increment or decrement from an empty or unparseable draft gives StartingValue, which defaults to Min or 0. An explicit starting value is only clamped to the bounds; it is snapped to the step grid on the next operation.

Up and Down change the value by one step and PageUp and PageDown by ten steps. When Min or Max is set, Home and End go to that bound. Pressing a button changes the value once immediately; holding it for 400 ms then repeats every 60 ms. Releasing it does not add another change, and repetition stops when the control is disabled or unmounted. When the input is focused and the vertical scroll amount is large enough, the mouse wheel changes the value by default, with scrolling down increasing it. DisableWheelChange turns this off and InvertWheelChange reverses the direction.

Valid edits update the model as you type. Enter snaps a valid value to the grid, or restores an invalid draft from the bound value; blur keeps the draft. AllowInvalid lets finite values go outside the bounds and keeps manual input without marking it as a range error; your app handles business validation. Numbers use ASCII decimal and scientific notation. Go's hexadecimal and underscore forms are not accepted, and IME is not enabled. NaN/Inf models, starting values, and invalid bounds are rejected. The component provides no locale, currency, or Intl formatting, and float64 is not suitable as an exact currency type.

| Orientation | Layout |
| --- | --- |
| Horizontal (default) | Minus on the left, value in the center, plus on the right |
| Vertical | Value on the left, with up and down arrows stacked on the right |

Both layouts have the same field height. The buttons inherit the field's color and size by default, use the link style, and swap start and end sides in RTL. Increment and Decrement show their buttons by default and can be turned off explicitly through Optional. IncrementOptions and DecrementOptions configure the buttons, and IncrementIcon and DecrementIcon override the icons. Explicit icons are not rotated in vertical layout. Disabling one side disables only that button; the keyboard and wheel still follow the field's rules.

FocusOnChange defaults to true, which returns focus to the input after incrementing or decrementing; false lets the button keep focus. Changing the orientation or display keeps the editor's identity, draft, and selection. Assigning a value from outside or clicking another control does not focus the input. Style receives the whole field, the editor, and the leading and trailing accessories. UI.IncrementSlot and UI.DecrementSlot style the buttons, and a custom button Style runs last. The buttons keep their native Tab stops; there is no complete mapping to the web spinbutton role.

### InputTags

InputTags binds a `*[]string`, and TagsState.Query stores the draft that has not been added yet. Enter adds a tag, and tags and the editor wrap together on the same lines. With the cursor at the start of the draft, Backspace first selects the last tag, and a second Backspace or Delete removes it. Left and Right move the tag selection, and Home and End go to the first and last tags while a tag is selected. The delete button uses DeleteIcon, × by default; after a deletion, focus returns to the input.

The component trims leading and trailing whitespace from a candidate, converts the string with ConvertValue, then removes case-sensitive duplicates and checks MaxTags. DisplayValue changes only the text of the tag and its delete button; the value and stable identity are unchanged. OnTagsChange returns a copy of the slice. OnAddTag, OnRemoveTag, and OnInvalid report each added, removed, or rejected candidate. A successful batch paste fires OnTagsChange at most once, while the per-item events are still reported separately.

Pasting is not split into tags by default. With `AddOnPaste: nui.Set(true)`, pasted text is split on Separators, which by default contains only a comma. To also split on semicolons or newlines, set `Separators: ",;\n"`. A batch paste keeps the existing draft. For more about when tags are added and how duplicates are handled, see [Switches, tags, and ratings](options.md#switches-tags-and-ratings).

Tags use the Elevated background with an Accented inner border, and a selected tag uses the Accented background. The font size follows the field size, and long tags can wrap. When the editor or a tag button has focus, the frame shows the accent color and a focus ring; in the error state it uses the error color, and None shows no focus ring. Root is the whole field and Body is the wrapping area. Item, Label, Close, and Indicator are the tag container, its text, the delete button, and its icon. Style's root also refers to the whole field. Object models, regular expression separators, drag-and-drop reordering, and remote entity matching are not supported.

### PinInput

PinInput binds a `*string` and keeps a separate PinState. It defaults to 5 text cells; for a six-digit verification code, use `PinInputOptions{Numeric: true, Length: 6}`. Length ranges from 1 to 32 and is counted in runes. Numeric accepts only ASCII digits.

Highlight defaults to false; when enabled, all cells show a border in the current color, and a field error takes precedence. UI and Style support Root, Item, Label, and SeparatorSlot. Root.Gap/Padding and the separator's font, font size, and padding are included in the overall size calculation, and an explicit 0 is honored. If a raw Style callback changes the width or height directly, your app coordinates the parent's size.

PinInput uses one real editor behind several visual cells, so Tab enters it only once. That editor handles editing, pasting, selection, copying, and undo/redo. Numeric mode does not enable IME; text mode does. OnComplete fires only when the required length is reached and the value has changed. The component does not read SMS messages or submit automatically.

| Action or option | Behavior |
| --- | --- |
| Click an existing character | Selects that position; typing replaces it and moves to the next. Typing at the last position replaces the last character |
| Paste several characters | Overwrites the corresponding positions without pushing later characters back |
| Backspace | Deletes the selected character. Clicking an empty cell or the gap between cells moves the cursor to the end of the existing content |
| SeparatorEvery / Separator | Groups cells at a fixed interval or at given positions, for example `SeparatorEvery: 2` or `Separator: []int{3}` |
| ReadOnly / Placeholder / masking | Control read-only state, text in empty cells, and whether the content is shown |

Cells are 6 DIP apart, scaled by Density. The separator is a centered dot in the Dimmed color with equal space on both sides. The current cell shows a 3 DIP focus ring in the theme color at 25% opacity. In the error state it uses Error, in high contrast mode the ring is opaque, and None shows no ring.

A focused empty cell hides its Placeholder and shows a blinking cursor. With reduced motion, the cursor stays solid; it is hidden when read-only, disabled, or unfocused. When FormField has an Error, every cell shows the error border. Text scaling and masking do not affect hit testing.

### InputRating

InputRating binds an `*int`. Max defaults to 5, and 0 means not rated. Only whole-number ratings are supported; there are no half stars, arbitrary precision, or drag-to-rate.

| Action | Behavior |
| --- | --- |
| ←→ (horizontal) / ↑↓ (vertical) | Move one level and stop at the ends. Horizontal keys follow RTL, and keys for the other axis do nothing |
| Home / PageUp | Set to 1 |
| End / PageDown | Set to Max |
| Space | Sets an empty rating to 1. With Clearable, resets the current rating to 0 |
| Enter | Not handled |

Arrow keys never lower the rating from 1 to 0; clearing requires Clearable. The keyboard focus ring is on the current star, or on the first star when there is no rating. A read-only rating can still receive focus with Tab, but ignores keys that would change it.

Stars are 12/16/20/24/28 DIP for sizes XS to XL, 2 DIP apart, and the component is as wide as its content. When interactive, the hovered star scales to 1.1×. With Hoverable on, hovering only previews the rating; the preview stays while the pointer crosses gaps between stars and is canceled when the pointer leaves or the rating is cleared. When FormField has an error, filled stars and the focus ring use the Error color.

UI and Style support Root, Item, Indicator, IconSlot, and EmptyIconSlot. The empty icon is drawn as the base layer, and Indicator and IconSlot are layered on top within the rated or hovered range. If the two icons differ, the base layer may still show around the overlay. IconSlot inherits Indicator's foreground color and can override it. Root's Gap/Padding are included in the overall size calculation, and Item's Shadow can override the default focus ring. With Neutral, rated icons use Highlighted and the focus ring uses InvertedBackground. A read-only rating keeps its normal colors and the default cursor.

Max accepts 1–20, and 0 uses the default of 5. When the rating or Max changes from outside, the rating is silently clamped to 0–Max; OnChange reports only actual user changes. Hoverable and Clearable default to false. Nuxt UI's step can represent fractional ratings such as half stars, but this library's integer binding has no corresponding field. There is also no arbitrary item content builder, HTML form field, or web radio node.

### CheckboxTri

CheckboxTri uses a CheckState: Unchecked, Checked, or Indeterminate. The indeterminate state is reported to assistive technology as Mixed. The first activation changes it to Checked, and later activations toggle between Checked and Unchecked.

`IndeterminateIcon` sets the icon for the indeterminate state and defaults to a minus sign; `CheckedIcon` sets the ordinary checked icon. With the indicator hidden, only an explicit `CheckedIcon` is shown above the label; the indeterminate icon is not used and no default icon is added. `OnChange` reports whether the box is checked, and `OnStateChange` reports the full CheckState. Each effective action fires each callback once; changing an icon fires neither.

Your business logic computes parent-child check relationships. If you only need on/off, use a Checkbox bound to a `*bool`.

### RangeSlider

RangeSlider takes two `*float64` values. By default, the range is 0–100 and Step=1. If the two thumbs cross, they are reordered and active focus moves with them so that low ≤ high always holds. Moves that would violate the minimum distance are rejected. Out-of-range values passed by your app are clamped during the build.

Common options include Orientation, Inverted, Tooltip, MinStepsBetweenThumbs, Color, Size, OnChange, and OnCommit. Each thumb has its own focus and reports its own value and minimum/maximum names to assistive technology.

Indicator.Background and Thumb.Background work the same way as on a single Slider. Floating-point noise is removed from fractional steps, and the minimum-distance comparison uses a tolerance that scales with the distance; even a very small step cannot make the thumbs overlap and then push the other thumb along. Disabling a parent container cancels the current drag without a commit.

Inside a FormField, clicking the label focuses the first thumb, and the field's description and error are associated with it. For orientation and keyboard behavior, see [Sliders, tabs, and accordions](options.md#sliders-tabs-and-accordions). For vertical assistive technology limits, see [Accessibility and motion](compatibility.md#accessibility-and-motion).

## Listbox, SelectMenu, and async search

### Listbox: a selectable list shown inline

Each ListboxItem's ID must be non-empty and unique. Label and Description provide the text, and Group adds a group heading to adjacent items. Kind can be ListboxOption, ListboxLabel, or ListboxSeparator; structural rows also need IDs. The selected value is a `*[]string`, where an empty slice means nothing is selected. ListboxState stores the query, the active item, scroll state, and a persistent QueryHandle.

The default is SelectionToggle: in both single and multiple selection, activating a selected item again deselects it. SelectionReplace replaces the selection on activation, and in multiple selection it also supports Shift range selection. Selected IDs that are temporarily missing from the data source are kept, which helps with async search; your app cleans up IDs that are gone for good.

Icon, Avatar, and Chip are checked in that order for the leading content, which aligns with the first line of text. SelectedIcon overrides the default check icon. Per-item options provide UI, Style, and OnSelect. The callback receives the ID and the selected state before the change, and can call PreventDefault to cancel this selection and the component's OnSelect. For the row appearance, see [Option row appearance](inputs.md#option-row-appearance). An MD row is about 32 DIP tall, plus one line when it has a description.

Size and Color default to MD and Primary. Instance options take precedence over the component theme, and the size can also be inherited from FormField. Highlight keeps a colored border visible. Keyboard focus shows a focus ring, which uses InvertedBackground with Neutral. When embedded in SelectMenu, the list follows the menu's size and color.

Filter defaults to false. When enabled, configure the search input through FilterInput. SearchTerm takes a persistent string pointer; when unset, State.Query is used. FilterFields defaults to matching label only and also supports description, value/id, and group; an explicit empty list matches nothing for a non-empty query. Filtering ignores case and combining diacritical marks. Within each run of options between groups and structural rows, results are ordered by exact match, then prefix match, then substring match. Full locale collation rules are not implemented.

With the filter field hidden, a retained query still applies; IgnoreFilter skips local filtering. Loading replaces the options with an icon and pauses list interaction, while the filter field stays editable. LoadingIcon overrides the icon, and the spinner stops with reduced motion.

| Orientation or key | Behavior |
| --- | --- |
| Vertical | A native virtual list; ↑↓ move |
| Horizontal | Items are shown and clipped at a fixed ItemWidth; ←→ move. Group headings also take one item width |
| Home/End, Enter/Space | Go to the first or last item; activate the item |
| Typing characters | Finds items by prefix |
| Enter in the filter field | Keeps focus in the input and activates the highlighted item; during IME composition, the IME handles it |
| Reordered data / outside selection changes | Reordering keeps stable IDs. On initialization or when the selection changes from outside, the first available selected item is highlighted, or else the first available item |

In both orientations, only the visible items plus a few extra are built. Height is the total height including the filter field; when unset, it is 224 DIP vertically and 64 DIP horizontally. ItemWidth defaults to 160 DIP in horizontal layout. A synchronous Build can customize non-interactive row content. Loop defaults to false and HighlightOnHover to true. OnHighlight reports only highlight changes caused by the user, not initialization or outside value updates.

### SelectMenu: searchable popup

SelectMenu combines a trigger with a search field and a Listbox inside a popup. SelectMenuState stores Open, Query, the Listbox state, the trigger Handle, and the search field's QueryHandle. Size and Color apply to all three areas. SearchTerm can bind an external query pointer; assigning it from outside does not fire OnQuery or OnSearchTermChange.

By default, both ↑ and ↓ open the menu from the trigger and focus the search field. With FocusSearch=false, SearchInput=false, or a disabled search field, the list receives focus instead. In the search field, Enter activates the highlighted item directly, ↑↓ move into the list, and Escape closes the menu and returns to the trigger. Single selection closes the menu; multiple selection keeps it open. If your selection callback moves focus deliberately, that target is kept.

The default filtering is the same as Listbox. SearchInput only controls whether the search field is shown; when hidden, the existing query still applies. When ExternalFilter or IgnoreFilter is true, the results your app provides are used as is. FilterInput configures the search field. SearchPlaceholder remains as a shortcut, but FilterInput.Placeholder takes precedence. The inherited Filter field has no effect on SelectMenu.

Leading and Avatar configure the trigger's leading content, Trailing sets the trailing arrow, and Highlight emphasizes the border. Loading shows a spinner using LoadingIcon, and LoadingTrailing controls its position. The native loading status row is also kept, and existing results can still be selected. Error and EmptyText set the status text and the no-results message. ShowTrailing=false hides the whole trailing area.

| Option | Behavior |
| --- | --- |
| ResetSearchTermOnSelect / ResetSearchTermOnBlur | Both on by default. Clear the search term after a selection and after closing, respectively |
| Clearable / ClearOptions / ClearIcon | When something is selected, replaces the trailing arrow with a configurable clear button. Hidden when ShowTrailing=false |
| ResetModelValueOnClear | Default true. Controls whether the clear button resets the selected value. The button always clears the query, and deleting search text by itself never resets the selection |
| OnClear / OnOpenChange | Fire once per user clear or actual open/close change. Changing State.Open directly does not report an open/close change |
| OnCreate | Shows a create row that can be clicked or activated from the keyboard. Only sends a create request; it does not add or select anything automatically |
| CreateWhenAlways / CreateAtTop | By default, the create row appears only when nothing matches. These options show it whenever no item has the same ID, and change its position |

A create request keeps the query and the open state. OnCreate receives the query with leading and trailing whitespace trimmed and does not change the options or the selection. On the create row, OnHighlight returns the query text instead of exposing an internal row ID. OnChange fires only when the value changes; OnSelect also fires when the selected value is activated again. Clearing or resetting the query fires OnQuery and OnSearchTermChange only if the query actually changed.

With nothing selected, the trigger shows placeholder text in Dimmed. A Subtle trigger's background darkens on hover. The search field in the popup uses the None variant with no leading icon by default and has a separator below it. The list has no border, rows are inset by 4 DIP, and the default maximum height is 240 DIP, shrinking when there is little content. If you customize Build or part heights, set Height explicitly; the automatic estimate does not inspect arbitrary custom layouts. Root, Trigger, Leading, Label, and Trailing style only the trigger. The list uses separate parts such as List, Header, ItemLeadingSlot, ItemLabelSlot, and ItemTrailingSlot, and the panel uses Popup.UI.

### Caching and large data sets

Listbox caches a snapshot of the data, an ID index, the groups, and the selection set. These structures are not rebuilt when the input has not changed, and in-place changes to the slice are also detected. However, each build still compares the snapshot linearly, and prefix search may scan every item.

SelectMenu caches filter results and selected labels; options that are temporarily missing keep their last label. Processing a large number of selected items for the first time still takes linear work. Virtual lists reduce the amount of UI built, but the cost of processing data still depends on its size.

### QueryState: reject stale search results

QueryState[T] can serve as the app data source for SelectMenu or InputMenu. Your app sends requests; QueryState handles cancellation tokens, the loading state, and checking whether a result is still valid.

```go
// Start on the UI thread; queryState is stored per window.
ctx, ticket := queryState.Start(parentContext, query)
go func() {
    items, err := fetchItems(ctx, ticket.Query)
    window.Update(func() {
        queryState.Complete(ticket, items, err)
    })
}()
```

`fetchItems` is your app's request function. In the next build, pass queryState's Items, Loading, and Error to the menu. Remote search is usually combined with ExternalFilter.

Start cancels the previous context. Complete clones the result slice and rejects tickets that are stale or come from another instance. Call `queryState.Cancel()` when leaving the view or closing the window. The library creates no network tasks and does not prescribe debounce timing, fuzzy matching, or remote paging.

Multi-select lists provide a selection role and selected-row semantics, but MyGo does not yet expose a separate multiselectable flag, so the full screen reader experience still needs verification.

## Calendar, InputDate, and InputTime

### Choose a value type

| Type | Represents |
| --- | --- |
| `Date{Year, Month, Day}` | A Gregorian date in the years 1–9999, without a time zone or time of day |
| `Optional[Date]` | A single date that can be empty |
| `DateSelection{Start, End}` | A single date or a date range; both ends are Optional[Date] |
| `ClockTime{Hour, Minute, Second}` | A time without a date or time zone; Hour is 0–23 |

Date provides Valid, Compare, AddDays, AddMonths, String, and ParseDate. When AddMonths lands on a day that does not exist in the target month, it uses the last day of that month. DateFromTime reads the year, month, and day in the Location of the given time.Time without converting to UTC. To represent a specific instant, your app chooses the time zone and handles daylight saving time.

### Calendar: pick a date by clicking

Calendar binds a `*DateSelection`. CalendarState stores the first displayed month (Month), the keyboard's active date (Active), the navigation view (View), and a persistent Handle. With Range=false, only Start is used. With Range=true, the user picks the start and then the end, and a reversed selection is put in order. By default, selecting the same single date, or the start of an incomplete range, again deselects it; PreventDeselect=true keeps the value instead. Enable it too if users need to pick a range that starts and ends on the same day.

ReadOnly allows focusing, browsing, and paging but prevents the user from changing the value. OnChange fires only when the value actually changes, not for outside assignments or rebuilds caused by option changes. OnMonthChange and OnViewChange report changes to the first month and the navigation view caused by the user; changing the state directly from outside is silent.

| Option | Behavior |
| --- | --- |
| Min / Max / IsDateDisabled | Restrict dates. Arrow keys skip disabled dates, and range ends must be selectable |
| IsDateUnavailable | Shows a strikethrough. The keyboard can move onto it, but it cannot be chosen as a range end |
| AllowNonContiguousRanges | Default false; true allows a range to span unavailable dates |
| IsDateHighlightable | Explicitly allows a range to span the given inner dates. Does not lift the restriction on range ends |
| MaximumDays | Maximum range length in days, counting both ends; 0 means no limit. Candidates beyond the limit are disabled |
| FixedDate | Fixes the start or end of a complete range. Moving past the fixed end updates that end and keeps the dates in order |
| Color / Variant / Size | Default Primary / Solid / MD. Supports Solid, Outline, Soft, and Subtle. The size is inherited from the field or the component theme |
| WeekStart / WeekdayFormat | WeekStart keeps the native default of Sunday. Weekdays use narrow names by default, and abbreviated or full names are also available |
| FixedWeeks | Default true, which always shows six weeks; false shows four to six weeks depending on the month |
| NumberOfMonths / PagedNavigation | One month by default, up to 12, wrapping when space runs out. PagedNavigation makes the month buttons page by the number of months shown |
| MonthControls / YearControls | Default true, which shows the month and year buttons in the day view. Buttons and icons can be configured individually |
| ViewControl / ViewControlOptions | The heading button is on by default and switches between the day, month, and year views. Month and year cells only navigate; they do not commit a date |
| YearsPerPage | The year view shows 12 years per page by default, from 1 to 120. IsMonthDisabled/Unavailable and IsYearDisabled/Unavailable configure the matching views |
| PreviousPage / NextPage | Custom target dates for paging back and forward. Used by the month buttons and the paging buttons of the month and year views; does not change the separate previous/next year buttons or the keyboard rules |
| Today | Pins today's date, which is useful for examples and tests. When unset, the current local date is used |

| Key | Behavior in the day view |
| --- | --- |
| ←→ / ↑↓ | Move by day / by week. Left and Right follow RTL; disabled dates are skipped and bounds are respected |
| PageUp / PageDown | Previous / next month, keeping the day of the month where possible |
| Shift+PageUp / PageDown | Previous / next year |
| Home / End | First / last available date of the current week |
| Enter / Space | Select or deselect the current date. Read-only, disabled, or unavailable dates do not change the value |

The month and year views use a four-column grid. Arrow keys move by column or row, and PageUp/PageDown move by year or by page of years. Choosing a cell returns to the next view down without changing the date value. This navigation also works in read-only mode.

WeekNumbers are computed from the explicit WeekStart. With Sunday, week 1 is the week containing January 1; with any other start day, it is the week containing January 4, so Monday gives ISO week numbers. This differs from Nuxt UI, which derives the week start and week numbers from the locale. Changing the language updates the heading, weekdays, and button names but does not change an explicit week setting.

Month is normalized to the first of the month. When the selected start or end changes from outside, the view follows the valid date; clearing the selection keeps the current view. Changing the number of months keeps the active date visible, and at the end of the year 9999 only the remaining months are shown. Multiple months share one native keyboard focus and wrap to fit narrow containers.

BuildDay, BuildMonth, and BuildYear build cell content synchronously, and BuildWeekday configures the weekday content. The container still provides the default names, selection, disabled state, and focus. Keep cell content non-interactive; the calendar determines its size. BuildHeading replaces the default heading and receives the current month and view; custom controls manage their state through the persistent CalendarState. Builders and date predicates may be called many times, so never perform I/O or submit anything from them.

Dates and ranges are always complete Gregorian dates. There is no discrete multi-date model and no separate whole-month or whole-year value model. The month and year navigation views do not implement Nuxt UI's `type="month"` or `type="year"`; manage those values with app controls such as Select or InputNumber. Other calendars, time zones, and browser DOM slots are outside this native model.

### InputDate: type a date

InputDate binds an `*Optional[Date]`, and DateInputState stores the segments, the draft, and the optional calendar state. Dates are Gregorian, and the segment order and separators follow the window's locale. PlaceholderValue provides a reference date for the first adjustment of an empty segment; it does not set the business value and does not replace the placeholder text.

| Action | Behavior |
| --- | --- |
| ←→, Tab / Shift+Tab | Move between segments. Arrow keys follow RTL; Tab enters or leaves the field |
| Typing digits | Advances to the next segment once the digit count or the maximum value is reached. Years take four digits |
| ↑↓, Home / End | Cycle by DateStep, or go to the segment's minimum or maximum. A step of 0 uses 1, and changes do not carry into neighboring segments |
| Backspace / Delete | Clear the current segment. After Ctrl/Cmd+A, deleting clears the current date endpoint |
| Ctrl/Cmd+C/X/V | Copy, cut, or paste the current endpoint; system edit commands also work |

Changing the year or month clamps a day past the end of the month to the last day of the target month. Pasting accepts ISO dates or numeric dates in the current locale's order. Calendar.Min/Max and IsDateDisabled/IsDateUnavailable restrict both typing and the calendar, and the bounds are inclusive. Outside assignments and option changes do not rewrite the model; if the current value violates a constraint, an error is shown.

DateInputState.Text stores an ISO draft, with `-` for unfilled segments. An invalid paste keeps the pasted text and the Error, while the segments on screen and the last valid business value stay unchanged. The business value is cleared only when every date segment is empty. OnChange reports the draft; OnValueChange fires once only when the valid single date, or the empty value, actually changes. Outside assignments are silent. Changing the language keeps the draft and focus. Read-only mode allows navigation and copying but not editing.

ShowCalendar defaults to false. When enabled, it adds a calendar button and a Popover that closes after a date is chosen. HideCalendar takes precedence over ShowCalendar. The Calendar's appearance, navigation, and part options apply to the popup, while Range is determined by which input function you call; a custom Label is kept. For a single date, Calendar.OnChange reports only calendar actions; segment editing is reported through OnValueChange. The popup closes if the input, its field, or an ancestor becomes disabled.

### InputDateRange: type a start and end date

InputDateRange binds a `*DateSelection` and keeps a DateRangeInputState. Both ends share one frame, navigation can cross between them, and they wrap when space runs out. SeparatorIcon defaults to a minus sign. BuildSeparator can replace it synchronously with non-interactive content, and SeparatorSlot styles it.

Moving the start after the end clears the end, and the end cannot be earlier than the start. Range rules such as MaximumDays, unavailable dates, and AllowNonContiguousRanges also apply to segment input. If the start value set from outside is past the calendar's upper bound, an error is shown instead of producing conflicting calendar bounds.

OnRangeChange reports actual range changes caused by segment editing or the calendar. Clearable clears both ends at once with a single notification. Calendar.OnChange remains as a compatible range notification and is also called only once per change. Ranges do not call the single-date OnValueChange. OnChange still reports each endpoint's draft, so clearing or a calendar change to both ends can fire it once per end; OnClear reports the clear action as a whole. The popup closes after the second calendar endpoint is chosen.

Inside a single FormField, an error focuses the first segment of the start date. To register errors and focus for each end separately, combine two InputDates.

### InputTime: type a time

InputTime binds an `*Optional[ClockTime]`, and TimeInputState stores the draft. When TwelveHour is unset, the window locale's clock is used; `nui.Set(true)` and `nui.Set(false)` fix it to 12-hour and 24-hour respectively. Up and Down cycle within the current segment. When the hour crosses noon or midnight, AM/PM updates, and A or P selects the period directly.

Granularity can be TimeHour, TimeMinute, or TimeSecond. When unset, minutes are shown by default, and Seconds=true adds seconds; an explicit Granularity takes precedence. Changing the clock or granularity does not rewrite the business value or fire notifications, and hidden units are kept. If the focused minute or second segment is hidden, focus moves to a visible segment; if the AM/PM segment is hidden, focus moves to the hour.

TimeStep sets the step for hours, minutes, and seconds separately. 0 uses 1, negative values are invalid, and changes do not carry into other segments. StepSnapping defaults to false. When enabled, leaving a non-empty segment rounds it to the nearest valid step counted from 0, rounding up when exactly halfway; in 12-hour mode, the rounding uses the full 24-hour time. No snapping happens when read-only or disabled, for rejected drafts, or when only display options change.

Min and Max are inclusive bounds on the complete time and include the hidden minutes and seconds in the comparison. Reversed bounds for a range that crosses midnight are not supported. IsTimeUnavailable can reject specific times. Invalid input keeps the draft and the last valid value; outside assignments are silent, and an out-of-range outside value shows an error. PlaceholderValue provides a reference value for the first adjustment of an empty segment, and at hour granularity, a new value's hidden units can also come from it.

The Text and paste formats are `HH` and `HH:MM[:SS]`, with ` AM` or ` PM` appended in 12-hour mode. The displayed text and segment order use the locale format. Currently, only ASCII digits and the canonical time text above are accepted; localized AM/PM words are not parsed.

### Date and time styles and limits

The date and time inputs share Input's five sizes, five variants, colors, avatar, icons, loading state, disabled and read-only states, and Clearable. A leading icon or leading spinner takes precedence over Avatar. LoadingTrailing controls the spinner position, and an explicit Leading keeps a leading spinner. With Neutral, focus and Highlight use the inverted background color, and Highlight also works with None. Shadow overrides on Root from the theme or the instance are kept, and Style runs last.

Root styles the frame, Editor sets the base style of the segment text, Segment styles placeholders, the focus background, and corner radius, and Leading and Trailing style the accessories. Style's editor is the first focusable segment in the current order and can bind a Handle. It is not a TextInputBase, so text selection APIs do not apply. Placeholder, Password, and MaxLength do not apply to segmented input; placeholder text comes from the language catalog.

Date represents a complete Gregorian date and ClockTime a single time of day with second precision; neither carries another calendar, milliseconds, or a time zone. Dates and times are edited separately. For a time range, combine two InputTimes and validate them in your app. Segmented input has no undo/redo, and the full screen reader experience has not been verified. Calendar's month and year views are for navigation and do not change the date value type.

## ColorPicker

ColorPicker binds a `*ui.Color` and keeps a ColorPickerState. The saturation/value area, the horizontal hue and alpha sliders, HEX/HEXA, the swatches, and the R/G/B/A inputs (0–255) all edit this color. Your app initializes it; unlike Vue's defaultValue, the zero value is not replaced with white automatically.

Size defaults to MD. For XS/SM/MD/LG/XL, the default selection area is 152/160/168/176/184 DIP square. Size also adjusts the native panel spacing and the numeric field sizes, and everything scales with density and text size. The root's maximum width is constrained by its container; an explicit width, padding, or a very narrow container can change the selection area's proportions. Changing the size does not reset the value, drafts, or focus.

When Throttle is unset, 50 ms is used; `nui.Set(time.Duration(0))` disables throttling, and negative values are invalid. The bound color always updates immediately, while OnChange calls are merged per interval and the last color is kept. Outside assignments are silent, and re-entering the same channel value does not notify again. Disabling, unmounting, removing the callback, or replacing the color from outside cancels a pending notification. Scheduling is handled by the window build and persistent services; no Element, Context, or background callback is stored.

An incomplete RGBA draft keeps the last valid color. The numbers use InputNumber's integer stepping. Editing an actual channel rebuilds the color from sRGB bytes, so stale wide-gamut data is not kept. A form tracks these drafts as a single color field, so the initial display and rebuilds without edits are not mistakenly marked as changed by writes from the four sub-inputs.

When Label is empty, it is inherited from FormField. Arrow keys adjust saturation and value in the SV area, the sliders keep their native keyboard behavior, and the internal names follow the window language. Each RGBA channel is an integer up to 255. HEX accepts 3-, 4-, 6-, or 8-digit forms. An invalid HEX value keeps the last valid color, and the canonical text is restored on blur following the underlying native rules. Form labels and error focus use the first RGBA number input; there is no public persistent Handle for the internal SV area or HEX field.

UI and Style support Root, Panel (the native panel container), Body (the RGBA container), Item (a channel row), and Label (a channel letter). The internal SV area, hue/alpha tracks, HEX field, and swatches have no public parts and use a local Native theme; you cannot style them with arbitrary part names. The native layout keeps horizontal tracks, alpha, and swatches, which differs from Nuxt UI's vertical hue track. In an RTL form, the native color panel is fixed to LTR so that click positions, arrow keys, gradients, and thumbs agree; the RGBA rows still follow the form's direction.

The color value is a ui.Color; your app handles serialization to HEX, RGB, or any other string format. This component provides no HSL/CMYK/Lab text editing, wide-gamut editing, screen color picking, or color management. HEX editing is affected by the [IME cancel issue](compatibility.md#canceling-ime-composition), and the full screen reader experience has not been verified.

## Table

### Prepare rows and columns

TableState[T] stores the current row data, sorting, selection, and column layout. Call SetRows when the data changes; do not set the rows again on every view build.

```go
// Declare the type and state in your app model.
type TaskRow struct {
    ID string
    Title string
}

var tasks nui.TableState[TaskRow]

// Run on initialization or when the data changes.
tasks.SetRows([]TaskRow{
    {ID: "task-1", Title: "Collect requirements"},
    {ID: "task-2", Title: "Review the docs"},
}, func(row TaskRow) string { return row.ID })
```

Define the displayed columns and create the table in the view:

```go
nui.Table(c, theme, &tasks, []nui.TableColumn[TaskRow]{
    {
        ID: "title",
        Title: "Task",
        Width: 240,
        Value: func(row TaskRow) string { return row.Title },
    },
}, nui.TableOptions{Label: "Task list", Height: 320})
```

SetRows copies the row values and checks that each ID is non-empty and unique. If T contains pointers or slices, your app still owns the data they point to. Replacing the data removes selected IDs that no longer exist.

### Sorting, selection, and custom cells

| Field or method | Purpose |
| --- | --- |
| `TableColumn.Compare` | Stable local sorting; returns the comparison of two rows |
| `TableColumn.Sortable` | When unset, determined by Compare. An explicit false prevents sorting the column; true without Compare requires ManualSorting |
| `TableColumn.Cell` | A custom cell that receives the Context, the Theme, the row value, and whether it is selected |
| `Selection` / `Sort` | A `ui.Selection[string]` and a `ui.SortOrder`, respectively |
| `state.List.SortBy` | Point it at a persistent `[]ui.SortOrder` to enable multi-column sorting; Sort mirrors the first entry |
| `ManualSorting` | Keeps the input order from SetRows; your app loads or computes the data based on the sort state |
| `InvalidateSort` | Requests a re-sort when data captured by a comparator changes. SetRows invalidates the sort automatically |
| `state.List.Columns` | Stores column widths and order; your app can persist it |
| `Row` / `RowID` / `Len` | Read a row, a row ID, and the row count in the current view |

Columns can also set Width, MinWidth, MaxWidth, and Align. Sorting changes row positions, so when editing data, look up the business model by RowID or by the IDs in Selection. CellBuilds counts the cell builder calls in the current build.

Clearing the sort restores the input order from the latest SetRows, which is also kept for rows that compare as equal. With multi-column sorting, Shift-click or Ctrl-click (Cmd-click on macOS) on a header adds a sort column; further clicks switch it to descending and then remove it. An ordinary click toggles that column between ascending and descending. Removing a column or disabling its sorting removes the invalid sort entries.

OnSortChange and OnSelectionChange report sort and selection changes caused by the user; outside model assignments and cleanup by SetRows do not trigger them. Selection notifications return stable IDs sorted as strings. OnActivate returns a row ID on Enter or double-click, and OnChange keeps the native combined notification. These semantics differ from Nuxt UI's select event on row click. List.Selected is an optional active-row index; when it is not bound, OnActivate uses an internal index. Still use IDs when acting on business data.

Table sets List.Key, Selection, Sort, and Label; do not override these bindings in your app. List.SortBy, Columns, Selected, KeepOffset, FollowEnd, HeaderMenu, CellMenu, and similar fields follow MyGo's rules. List.CellSelect is a native extension, and this wrapper has no RowHeader column. Full cell selection semantics differ from row selection notifications.

### Filtering and scope

For search, filtering, and pagination, your app computes the current rows first and then calls SetRows. The data browser example shows search, team and status filters, and pages of 100 rows; your app manages server requests, totals, and caching.

EmptyText and BuildEmpty set the empty content below the header. Loading shows a loading bar and keeps the existing data. With no data, BuildLoading's content is shown if provided; otherwise the empty content is still shown. LoadingColor defaults to Primary. LoadingAnimation can be Carousel, CarouselInverse, Swing, or Elastic, and a static bar is used with reduced motion. Caption is the table's description for assistive technology, and BuildCaption can show a visible title. Height defaults to 320 DIP, including the title; the header stays fixed and the body scrolls inside.

Virtualization is always on. It reduces only the number of cells built; generating data, validating IDs, sorting, and filtering still scale with the data size. MyGo handles row layout. Cells can override Item and Label through a column's UI and Style, and column overrides are applied after table-level overrides.

`Fixed` **only prevents moving and resizing the column**; it does not freeze the column during scrolling. Frozen columns, multi-level headers, arbitrary row or header builders, expandable rows, grouping and aggregation, and full TanStack semantics are not yet supported. Native headers have no keyboard focus of their own; to support keyboard sorting, provide buttons that change Sort or SortBy. The underlying List measures row heights, so custom cells can be taller; your app should still verify complex dynamic content.

## Tree

### Store nodes, selection, and expansion

TreeState.SetNodes takes a list of TreeNodes and checks that every ID in the tree is non-empty and unique. The state fields are:

| State | Purpose |
| --- | --- |
| `state.Outline.Open` | IDs of expanded nodes |
| `state.Selection` | IDs of selected nodes |
| `state.Active` | ID of the node highlighted by the keyboard |

The default is single selection, and clicking a row both selects it and toggles its expansion. Arrow keys only navigate; Enter and Space select. Multiple enables multiple selection, and SelectionBehavior chooses between toggling and replacing the selection. PropagateSelect and BubbleSelect control parent-child linking, and Loop controls whether navigation wraps around.

Nodes can be Disabled, and OnSelect and OnToggle can cancel the current action. Parent-child linking applies only to loaded nodes, and disabled nodes block it. Children loaded later do not inherit the selection automatically.

### Indentation and virtualization

Nested defaults to true, which shows ancestor guide lines and indents each child's background and click area by level. With `nui.Set(false)`, the background and click area span the full row while the content stays indented. Switching keeps the selection, expansion, focus, and scroll position.

Both modes use a native virtual list that builds only the rows near the viewport. The index of visible nodes is rebuilt when the data or the expansion state changes. Because apps may edit the public collections directly, every build still checks the expanded IDs, and also the selected IDs when indeterminate states are needed. Customize row content with Build.

### Lazy loading

A node with Lazy=true and Children=nil calls `OnLoad(id, ticket)` once when it is visible and expanded. Your app starts loading in the background and returns the result in `Window.Update` with `CompleteLoad(ticket, children, err)`.

A failed load shows Retry. Collapsing the node or an ancestor with SetExpanded, replacing the data, retrying, or calling Close all invalidate the old ticket. A true result from CompleteLoad only means the result was accepted; it may still contain an error.

When SetNodes replaces the data, it removes selected and expanded IDs that no longer exist. An empty root list shows EmptyText.

### Migrating from the old Outline wrapper

| Behavior you need | Current code |
| --- | --- |
| Multiple selection | Set `Multiple: true` explicitly |
| Multiple selection that replaces | `SelectionBehavior: nui.SelectionReplace` |
| Click selects without expanding | `ExpandOnClick: nui.Set(false)` |
| Listen for selection / expansion | Use OnChange / OnToggle |
| Query visible rows | Use `state.Rows()`, `Item()`, and `Depth()` instead of the methods of the same names on `state.Outline` |

Outline.List keeps the scroll state and Handle, while Tree manages Selected, Selection, and Label. Drag-and-drop moving and a full business permission model are not provided. For the limit on assistive technology expand actions, see [Accessibility and motion](compatibility.md#accessibility-and-motion).

OnExpandedChange fires after a user action actually changes the expansion state, with the expanded IDs sorted as strings. Cancellation, initialization, and outside SetExpanded calls do not trigger it. OnChange still reports only selection changes. If a selection handler replaces the nodes, the earlier click no longer expands a newly created node with the same ID.

A TreeNode's UI and Style can override Item, Leading, Label, and Trailing; node overrides are applied after the tree instance's overrides. BuildLeading and BuildTrailing replace the leading and trailing icons, and Build replaces the main content; these callbacks should build display content only. TreeItemState.Indeterminate is used only for multiple-selection linking. With BubbleSelect, a node is indeterminate when its available descendants are a mix of selected and unselected; with only PropagateSelect, the parent must also be selected. In single selection, or with both linking options off, it is false.

A loaded empty branch keeps its expansion state but does not get the default folder icon and arrow; set Icon explicitly if you want a folder look. Type-ahead search cycles through matches when the same character is repeated, and keeps the current match while the user continues typing a prefix.

## Style part index

| Component | Parts |
| --- | --- |
| Form | Root |
| Listbox | Root, Editor, LoadingSlot, LoadingIconSlot, List, Label (group heading), SeparatorSlot, Item, Leading, Body, ItemLabelSlot, Description, Trailing, Indicator, EmptySlot |
| SelectMenu | Root, Trigger, Leading, Label, Trailing, Close, Editor, List, Header (group heading), SeparatorSlot, Item, ItemLeadingSlot, Body, ItemLabelSlot, Description, ItemTrailingSlot, Indicator, EmptySlot. The panel uses Popup.UI |
| InputDate / InputDateRange / InputTime | Root, Editor, Segment, Leading, Trailing. The date range also has SeparatorSlot |
| InputNumber | Root, Editor, Leading, Trailing, IncrementSlot, DecrementSlot |
| InputTags | Root, Editor, Leading, Trailing, Body, Item, Label, Close, Indicator |
| PinInput | Root, Item, Label, SeparatorSlot |
| InputRating | Root, Item, Indicator, IconSlot, EmptyIconSlot |
| Calendar | Root, Header, HeadingSlot, Body, GridSlot, GridRowSlot, WeekDaySlot, WeekNumberSlot, Item |
| Table | Root, CaptionSlot, Body, Panel, Item, Label, EmptySlot, LoadingSlot, Indicator |
| Tree | Root, List, Item, Leading, Label, Trailing, Indicator, EmptySlot |
| ColorPicker | Root, Panel, Body, Item, Label |
| RangeSlider | Root, plus Indicator.Background and Thumb.Background |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](foundations.md#style-overrides).

## Examples

| Command or page | Content |
| --- | --- |
| The Gallery's Advanced forms, Dates & color, and Task browser pages | Specialized inputs, dates, colors, and data interaction |
| `go run ./examples/settings` | The Project setup form and a cancelable submission |
| `go run ./examples/data-browser` | Search, filtering, Table/Tree, editing by ID, and Toast |

Text input in InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, and ColorPicker is affected by the known IME cancel issue; see [Canceling IME composition](compatibility.md#canceling-ime-composition).
