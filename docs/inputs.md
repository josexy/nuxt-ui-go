# Inputs, selection, and fields

Your app stores input values and selections; components display them and handle user interaction. Pick a component by need below, then read its section. For where to keep state and what the example variables mean, see [How component calls work](README.md#how-component-calls-work).

| What you need | Component |
| --- | --- |
| Single-line or multiline text | [Input / Textarea](#input-and-textarea) |
| Labels, descriptions, and error text, or several controls joined into one group | [FormField / FieldGroup](#formfield-and-fieldgroup) |
| A toggle, checkboxes, or a set of mutually exclusive options | [Checkbox / RadioGroup / Switch](#checkbox-checkboxgroup-radiogroup-and-switch) |
| Adjusting a number by dragging | [Slider](#slider); for a two-value range, see [RangeSlider](forms-data.md#rangeslider) |
| A dropdown select, or search-as-you-type | [Select / InputMenu](#select-and-inputmenu); for a separate search field and complex lists, see [SelectMenu](forms-data.md#listbox-selectmenu-and-async-search) |

For custom popup panels and scroll areas, see [Popover and ScrollArea](#popover-and-scrollarea). For a runnable example that combines these components, run `go run ./examples/settings`.

## State, themes, and events

### Storing values and control state

Components take the current `*ui.Context` and a `nui.Theme`. Keep input values, selected IDs, SearchState, ScrollState, open states, and Handles in your app model, separately for each window. Do not keep Elements, Contexts, or parts across builds.

Call `c.Key(stableID)` before creating each dynamic control. The same business data can appear in several views, but each window still needs its own editing state.

### Sizes, variants, and disabled state

Input controls support XS, SM, MD, LG, and XL, with MD as the default. At the default font size and density, the base single-line heights are 24, 28, 32, 36, and 40 DIP. Controls grow taller when the font's line height is larger; for example, Chinese fallback fonts on Linux can make an MD Input taller than 32 DIP.

Input, Textarea, Select, and InputMenu support these variants:

| Variant | Effect |
| --- | --- |
| Outline | The default bordered style |
| Soft / Subtle | A field with a background |
| Ghost | A subtler, low-emphasis look |
| None | No regular background or focus ring; the error border is still shown |

Fields do not accept Button's Solid or LinkVariant. Color mainly controls the focus and selection accent; font size, corner radius, density, and light/dark colors come from the Theme.

Size precedence is: the instance's Size → the FormField/FieldGroup Size → the component's default Size. Set component defaults with calls such as `WithComponent(ComponentInput, ...)`. When a parent field or any ancestor is disabled, a child cannot re-enable itself with `Disabled: false`.

### Focus and callbacks

Text inputs show an accent border and focus ring when focused by mouse or Tab, and return to normal on blur. Read-only fields can still be focused, selected, and copied, and an error border keeps its error color. Triggers such as Select keep their keyboard focus indicator. The hover background of Outline and Subtle does not apply when disabled.

An open Select panel inherits the disabled state of its trigger and ancestors; once disabled, the selection cannot be changed. The current MyGo version may keep the panel visible until it is closed.

MyGo processes queued input with the configuration from the previous build. `Changed` and `Submitted` report the input notifications received in the current build. `OnClick`, `OnChange`, and `OnSubmit` run after the build. Put business logic in these callbacks; `Build` and `Style` should only handle the current UI.

## Input and Textarea

### Bind a string

In this example, `app.Name`, `app.Bio`, and `app.NameError` are string fields in the window's model:

```go
nui.FormField(c, theme, nui.FormFieldOptions{
    Label: "Name",
    Required: true,
    Error: app.NameError,
}, func(field nui.Theme) {
    nui.Input(c, field, &app.Name, nui.InputOptions{
        Placeholder: "Enter your name",
        Clearable: true,
        ClearLabel: "Clear name",
        OnChange: func(value string) { app.NameError = "" },
    }).FillWidth()
})
nui.Textarea(c, theme, &app.Bio, nui.TextareaOptions{
    InputOptions: nui.InputOptions{Label: "Bio", Placeholder: "Tell us about yourself"},
    Rows: 3,
    MaxRows: 6,
    AutoResize: true,
}).FillWidth()
```

The value argument must be a non-nil `*string`. An empty string is a valid value.

### Common options

| Option | Purpose |
| --- | --- |
| Label / Placeholder | The control's name and placeholder text |
| Invalid / ErrorMessage | Mark the control as invalid. To show error text below the field, use FormField.Error |
| Disabled / ReadOnly | Disabled blocks interaction. ReadOnly allows selecting and copying but not editing, and turns off IME for editing |
| Password | Hides the text, single-line Input only; Textarea does not support password mode |
| Leading / Trailing | SVG icons on either side of the input |
| Avatar | A leading avatar; Leading or a leading loading spinner takes precedence |
| Loading / LoadingIcon / LoadingTrailing | Shows a loading spinner; the input stays editable. LoadingTrailing moves the spinner to the end, but an explicit Leading keeps it at the start |
| Clearable / ClearLabel | Enables the clear button and sets its name |
| MaxLength | Limits how many UTF-16 code units the user can insert; 0 means no limit |
| Highlight / Color / Variant / Size | Highlight state, color, variant, and size |
| OnChange / OnSubmit / OnClear | Callbacks for input changes, submission, and clearing |

Placeholder text uses the Dimmed color by default. To change it, call `PlaceholderColor` on the editor in the Style callback. Password inputs expose only a password state in the native assistive technology tree, never the plain-text value.

If Avatar.Size is unset, input sizes XS–XL use avatars of 16, 16, 20, 20, and 24 DIP; an explicit Avatar.Size takes precedence. A trailing spinner can appear together with a leading avatar. Changing the avatar or the spinner position does not reset the text, focus, or selection.

With the Neutral color, the focus and highlight borders use InvertedBackground. Highlight also works with the None variant; the error state takes precedence.

### Editing and clearing

MyGo's TextInputBase and TextAreaBase handle text editing, selection, copy and paste, undo, and IME. MaxLength limits insertions before OnChange runs; when replacing a selection, it keeps the unchanged suffix. An insertion that is rejected entirely and leaves the value unchanged does not fire OnChange. Values assigned from outside are not truncated and do not fire input notifications.

The clear button appears only when there is content and the input is not read-only, and it inherits the disabled state. Each clear updates the bound value, calls OnChange and OnClear once each, and returns focus to the editor. Clearing changes the value programmatically, so it is not guaranteed to be a single undoable edit. If your app needs to undo a clear, save and restore the data yourself.

**Known IME issue:** If you select text, start an IME composition, and then cancel it, the selected text may be lost; undo restores it. For the affected components, see [Canceling IME composition](compatibility.md#canceling-ime-composition).

### Textarea rows

Rows defaults to 3. Without AutoResize, the textarea always shows that many rows; with it, the textarea grows with the actual line count. A MaxRows of 0 or less means no limit, and a positive value less than Rows is treated as Rows. Once the limit is reached, the editor scrolls internally.

Textarea provides plain multiline editing, without a character counter, syntax highlighting, or rich text.

The avatar and loading spinner in a Textarea align with the first line. Enter and Shift+Enter insert a newline and do not fire OnSubmit; submit with a separate button or an editor shortcut instead. The shared Password field has no effect on multiline input. There is no native resize handle; adjust the height with Rows, AutoResize, or the surrounding layout.

### Styles and the editor Handle

Input and Textarea have the parts Root, Editor, Leading, and Trailing. `Style(root, editor, leading, trailing)` runs last and synchronously; parts that do not exist are zero-value Elements. Leading can be an icon, a loading spinner, or an avatar. Set the spinner's color with `UI.Leading.Foreground` or `UI.Trailing.Foreground`; Style.TextColor does not change the custom-drawn color. The clear button is named with ClearLabel and has no style part of its own yet.

The returned element is the outer container of the input. To bind editor focus or run editing commands, work with the editor in Style or bind a persistent Handle to it. Do not treat the outer container as the text editor.

## FormField and FieldGroup

### FormField: labels, descriptions, and errors

FormField associates its label, description, error, and Handle with the first control inside it that supports field binding. The main options are Label, Description, Help, Hint, Error, Required, Disabled, Size, Orientation, and Decoration.

`Required` shows a red asterisk and adds a "required" description to supported controls. **Your app or Form still performs the actual validation**; the underlying controls currently have no native required flag of their own.

| Region | Content |
| --- | --- |
| Label area | Label, Hint, Description |
| Content area | The input control, followed by Help or Error |
| Default layout | The label area above the content area |
| Horizontal | The label area and the content area side by side, aligned on the baseline of the first line of text |

For sizes XS–XL, the label, hint, description, help, and error text use font sizes of 12/12/14/14/16 DIP with line heights of 16/16/20/20/24 DIP. The description sits directly below the label. In vertical layout, the content area is 4 DIP below the label area. Help and Error are 8 DIP below the input, and in horizontal layout the two columns are also 8 DIP apart. Without a Label, Hint, or Description, no label area is created and no top gap is left. The asterisk is 2 DIP from the label, and the accessible name includes both the label and `*`. Without a label, no standalone asterisk is shown, but the required description is still passed to the control.

Clicking the label activates the first associated control. Checkbox, Switch, and CheckboxGroup keep their toggle and notification behavior, and a tri-state Checkbox goes from indeterminate to checked. Use a separate FormField for each independent input. Your app must associate labels with custom MyGo controls itself.

The parts are Root, Header, Body, Label, Hint, Description, Help, and ErrorSlot. Header is the label area, and Body is the input plus the help/error area. Required and Error reach child controls through the inner Theme, which stores only data and no UI references.

Error is a string and takes precedence over Help. When it is empty, the Form error for a touched field can be used instead. There is no equivalent of upstream `error=true` (highlight only), `error=false` (hide errors), `errorPattern` regex matching, or field-level deferred validation. Field names must match exactly, and Form.Validate or your app performs validation. In this native implementation, Hint can be shown without a Label.

### FieldGroup: join several controls

Use FieldGroup to join an input and a button into one continuous row:

```go
nui.FieldGroup(c, theme, nui.FieldGroupOptions{Size: nui.MD}, []nui.FieldGroupItem{
    {ID: "name", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Input(c, inner, &app.Name, nui.InputOptions{Label: "Name"}).Grow(1)
    }},
    {ID: "check", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Button(c, inner, "Check", nui.ButtonOptions{
            Variant: nui.Outline,
        }).OnClick(app.Check)
    }},
}).FillWidth()
```

Each item's ID must be unique and non-empty, and Build must not be nil; otherwise FieldGroup panics. Build creates the current content synchronously.

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

Disabled is passed down to the items. Size comes from the instance first, then the ComponentFieldGroup default, then the enclosing FormField; if none is set, each child uses its own default size. Popups opened by child controls keep their own corner radius, such as SelectMenu's search field and InputDate's calendar.

The Root part adjusts the spacing. Inside the group, each control keeps its own Tab order, and the group itself adds no focus stop. MyGo has no public z-index yet, so a neighboring control drawn later may cover the focus ring along a shared edge.

## Checkbox, CheckboxGroup, RadioGroup, and Switch

### Choose a binding type

| Component | Bound value | Use for |
| --- | --- | --- |
| Checkbox / Switch | Non-nil `*bool` | A single on/off state |
| CheckboxGroup | Non-nil `*[]T`, where T is comparable | A set of items where several can be selected; a nil slice means nothing is selected |
| RadioGroup | Non-nil `*T` | A set of mutually exclusive items; for an empty selection, use Optional[T] or an explicit "none" option |

Option values must be unique, but display names can repeat. CheckboxGroup appends a value when it is checked and removes it when unchecked. Reordering does not change the selected values, and selected values that are temporarily missing from the options are kept; your app decides when to clean them up. RadioGroup does not change a missing value to the first item, and it reserves no special zero value.

For a tri-state checkbox with Unchecked, Checked, and Indeterminate, use [CheckboxTri](forms-data.md#checkboxtri).

### Options and keyboard

Checkbox and Switch share ToggleOptions; CheckboxOptions and SwitchOptions are aliases of it. Common fields include Label, Description, ErrorMessage, Disabled, Invalid, Color, Size, OnChange, and Decoration.

A shared type does not mean every field applies to both controls. Checkbox does not use Loading, LoadingIcon, or UncheckedIcon; Switch does. RadioGroup always selects one item, so Max has no effect on it; CheckboxGroup's Max limits how many items can be selected. The website keeps these public fields and marks those that do not apply to the component being shown.

ChoiceTable on a single Checkbox and Variant/Indicator on Switch are Go-native extensions, whereas ChoiceTable on CheckboxGroup and RadioGroup corresponds to an existing Nuxt UI group style. The native extensions remain available.

Both group types use ChoiceGroupOptions, which adds Horizontal, Loop, and a parameterless OnChange. Groups are vertical by default; Horizontal lets items wrap, except with ChoiceTable. Each option's Label and Description provide its text. In checkbox groups, Option.Icon replaces the check icon; in both group types with the indicator hidden, it appears above the label. Option.Build has no effect in choice groups; it is only used by Select and InputMenu.

Checkbox and Switch toggle on click or Space. Each group is a single Tab stop. Use Left/Right arrows in horizontal layout and Up/Down arrows in vertical layout; Home and End go to the first and last items, and disabled items are skipped. In CheckboxGroup, arrow keys only move focus and Space toggles the item; in RadioGroup, moving also selects. Loop is an Optional[bool]: it defaults to false for checkbox groups and true for radio groups, and an explicit `nui.Set(false)` stops wrapping. State updates immediately; there is currently no toggle animation.

A group's visible title can come from the enclosing FormField.Label. To use one icon for all checkboxes in a group, set the same SVG as Option.Icon on every option that has no icon of its own; items with their own icons keep them.

### Icons, styles, and spacing

Checkbox's CheckedIcon replaces the default check mark, and CheckboxTri's IndeterminateIcon replaces the minus sign shown for the indeterminate state. IndicatorHidden places an explicit CheckedIcon above the label, with the icon and text centered; without an icon, only the text is shown. The indeterminate icon is used only in a visible indicator; it does not replace the label icon in hidden mode. UncheckedIcon applies only to Switch.

Checkbox's Leading part styles the check icon, or the label icon in hidden mode. ChoiceCard and ChoiceTable show the keyboard focus ring on the outer frame, while ChoiceList shows it on the visible indicator. The Switch track has no outline by default; an accent or error ring appears only with Highlight or in the error state.

A card with a visible indicator shows Highlight only on the indicator; with the indicator hidden, the card's frame shows the highlight. Card styles follow the usual override order: component defaults, the instance's semantic options, the instance's UI, and then Style.

Mark a field as required with FormField.Required. It shows an asterisk and gives assistive technology a "required" hint along with the field's description and the control's description. It does not validate the selection; configure FormState validation rules to enforce it on submit.

Items in a group are 8 DIP apart horizontally. Vertically, the gap is 2/2/4/4/6 DIP for sizes XS to XL. Adjacent ChoiceTable items share a 1 DIP border, the outer corners have an 8 DIP radius, and the corners follow the writing direction. The border after a selected item is omitted so the selected item's full border is visible.

The Checkbox/Radio indicator container is as tall as the text line: 16/16/20/20/24 DIP by size. The inner Radio dot is 4/4/6/6/8 DIP.

A regular label and description are stacked using their own line heights, with no extra gap; the centered layout used when the indicator is hidden has a 4 DIP gap. In CheckboxTri's indeterminate state, the indicator uses the accent color but the card background stays the same as unchecked; only Checked applies the selected background.

| Component | Parts |
| --- | --- |
| Checkbox | Root, Indicator, Leading, Label, Description |
| Switch | Root, Indicator, Thumb, Label, Description |
| CheckboxGroup | Root and Item apply to the group and its rows. Other parts and Style are passed to each Checkbox; you can also customize them with ComponentCheckbox |
| RadioGroup | Root, Item, Indicator, Leading (the icon in hidden mode), Label, Description |

## Slider

Slider binds a non-nil `*float64`. By default, its range is 0–100, it is horizontal, and its size is MD. Min and Max are Optional[float64]; the bounds must be finite with Max > Min, and Step must be finite and non-negative.

An explicit Label is used as the name first, then the FormField label; if both are empty, the localized "Slider" name is used. The label is not drawn as visible text; Description adds a supplementary description.

Values set from outside are clamped to the range; NaN and Inf are treated as 0 before clamping. Slider has no empty-value model. Step=0 uses the default step of 1.

User actions snap to multiples of Step from Min, and floating-point noise is removed based on the decimal precision of Min and Step. For example, adding 0.1 to 0.2 shows 0.3. Disabling the slider or a parent cancels any drag in progress, and OnCommit does not fire.

| Action | Behavior |
| --- | --- |
| Arrow keys | Move one Step, following the orientation, Inverted, and RTL |
| PageUp/PageDown, Shift+arrow keys | Move 10 Steps |
| Home/End | Go to the bounds |
| OnChange | Reports value changes live |
| OnCommit | Fires when a key press or drag that changed the value ends |

The default height is 24 DIP horizontally and 160 DIP vertically; change it on the returned element. The track is 6/7/8/9/10 DIP thick for sizes XS to XL. The thumb is a dot in the background color with a 2 DIP border in the theme color. The keyboard focus ring is 3 DIP wide and 2 DIP away from the thumb. RangeSlider looks the same.

For Slider and RangeSlider, Root styles the whole control, Indicator.Background sets the filled track and the thumb border, and Thumb.Background sets the thumb fill. Neutral uses InvertedBackground. A FormField error uses the Error color first, although explicit instance part colors can still override it. The unfilled track uses Theme.Accented. The track and thumb are drawn by a draw callback, so other Thumb and Indicator style fields have no effect, and Style receives only Root.

Tooltip is a bool. When on, hovering or focusing the thumb shows the current value immediately, and Escape hides it. It has no full TooltipOptions or ComponentTooltip part configuration. On a single slider, changing UI.Root.PaddingX/PaddingY changes the active track area, and drawing, pointer math, and tooltip positioning all follow. RangeSlider keeps a fixed track inset of half a thumb.

For dual thumbs, inversion, value tooltips, and vertical assistive technology limits, see [Sliders, tabs, and accordions](options.md#sliders-tabs-and-accordions).

## Select and InputMenu

### Choose an interaction

| What you need | Use |
| --- | --- |
| Pick one item from a list after clicking | `Select`, bound to `*T` |
| Pick several items after clicking | `SelectMultiple`, bound to `*[]T` |
| Show suggestions while typing | `InputMenu`, bound to `*SearchState` |
| A trigger with a separate search field in the popup | [SelectMenu](forms-data.md#selectmenu-searchable-popup) |

Select and InputMenu share `Option[T comparable]`, which has Value, Label, Description, Disabled, Icon, and Build. Value is the stable identity: it must be unique and equal to itself (so NaN is not allowed). Labels can repeat. Even with a custom Build, fill in Label for search and assistive technology.

Disabled items are still shown but cannot be reached with the keyboard. The list is virtualized and calls Build only for the visible rows plus a few extra, so do not run business logic that must see all the data inside a row builder. Comparing input data still costs more as the data grows; your app handles remote paging and async requests.

### Select: single selection and empty values

Select reserves no special zero value. To represent "nothing selected" explicitly, wrap the business value in an Optional:

```go
// app.Region has type nui.Optional[string].
nui.Select(c, theme, &app.Region, []nui.Option[nui.Optional[string]]{
    {Value: nui.Optional[string]{}, Label: "No region"},
    {Value: nui.Set("ap"), Label: "Asia Pacific", Description: "Singapore"},
    {Value: nui.Set("us"), Label: "United States", Disabled: true},
}, nui.SelectOptions{
    Label: "Region",
    Placeholder: "Choose a region",
}).FillWidth()
```

This example lists the Optional zero value as an option, so the select initially shows "No region". A value with no matching option is kept as is and Placeholder is shown; Placeholder defaults to an empty string. With no options at all, the built-in "No data" message is shown.

SelectOptions includes Label, Placeholder, ErrorMessage, Disabled, Invalid, Color, Variant, Size, Popup, OnChange, and Decoration. Leading and Trailing set the trigger icons; Trailing defaults to ChevronDown. Avatar is shown when there is no leading icon. SelectedIcon replaces the check mark on the selected item, not the item's own Icon.

Loading only shows a spinner; it does not disable selection. By default, LoaderCircle appears at the start; LoadingTrailing=true moves it to the end, but an explicit Leading keeps it at the start. LoadingIcon replaces the spinner icon, which uses the corresponding part's foreground color. If Avatar.Size is unset, the avatar diameter is 16/16/20/20/24 DIP by select size; an explicit avatar size takes precedence.

Highlight keeps the colored border when the select is not focused, including with the None variant; the error color takes precedence. With Neutral, the border and keyboard focus ring use InvertedBackground.

Open is an optional persistent `*bool`. Without it, Select uses internal state. With it, app assignments, trigger actions, and closing all update the same value, and focus returns to the trigger after closing. Opening the select, or changing the selection from outside, does not fire OnChange; only a user action that changes the selection fires it, once. Popup.Dismissible defaults to true; an explicit false prevents closing with Escape or an outside click, though setting Open to false still closes it.

When the select is open, use the arrow keys, Home, End, or type a prefix to move the highlight; press Enter or Space to select and Escape to close. Repeating a letter cycles through matches and skips disabled items. When a single select is closed, typing a prefix changes the selection directly, as a native convenience. In multiple selection, prefix navigation works inside the open list. Reordering keeps the highlight by ID. If the highlighted item is removed, the highlight moves to the current valid selection or the first available item. Multiple selection shows the selected names joined by commas, skipping empty labels; selected items that are temporarily missing keep their previous names.

The parts are Root, Label, Leading, Trailing, Panel, Item, and EmptySlot; SelectMultiple maps the Root style to the trigger. Option.Build can compose row content such as avatars and chips, and must not keep build-time elements. The current model has no label or separator item types and no nested groups. Popup has no arrow, no DOM portal, and no upstream item-aligned positioning. Single selection does not use Popup.Mode, OpenDelay, or CloseDelay; multiple selection keeps SelectMenu's native hover extension.

### InputMenu: drafts versus selected values

SearchState.Query holds the text being edited and Selected holds the single selected ID. An empty Selected means nothing is selected, so InputMenu option IDs cannot be empty strings. In multiple selection, SelectedValues is the source of truth. Removing an option from the data source does not clean up the app's selection. For ordinary single selection, if the initial draft is empty, the selected name is shown. If Selected changes from outside while the user is not editing, the display updates; a draft being edited is kept until the menu closes or loses focus.

By default, the search is a case-insensitive substring match on Label. FilterFields changes which fields are searched; the older SearchSelect convenience function searches Label and Description. The local filter result is recomputed only when the data, the query, or the filter fields change.

InputMenuOptions embeds InputOptions and adds options such as EmptyText, Popup, AllowCustom, and OnSelect. By default, clicking only focuses the input; typing, arrow keys, or the trailing arrow open the panel. Set OpenOnClick=true to open on click. The input keeps focus. Trailing can change the icon. Loading does not disable editing or selection. Avatar, LoadingTrailing, and Highlight work as in Input.

Clearable shows a clear button whenever there is a draft or a selection, so you can clear all tags in multiple selection even with an empty draft. ResetModelValueOnClear defaults to true; an explicit false clears only the draft and keeps the selection. Deleting ordinary text does not clear the selected ID. SelectedIcon, DeleteIcon, and ClearIcon replace the check mark, the X for deleting a tag, and the X for clearing. Autocomplete has no selection mark and no multi-select tags. After a tag is deleted with the keyboard, focus returns to the input.

Open takes a persistent bool pointer; nil uses internal state. Opening it from outside focuses the input, and becoming disabled or read-only closes it. Loop defaults to false and HighlightOnHover to true; both skip disabled items. Popup.Dismissible=false prevents closing with Escape or when focus moves outside, though setting Open to false still closes it. Popup's hover mode and delays do not apply to InputMenu.

| Action | Callbacks and state |
| --- | --- |
| Editing the draft | OnChange reports the text change |
| Choosing a suggestion | Calls OnSelect once and does not also submit free text |
| Pressing Enter with AllowCustom=true and no suggestion chosen | Calls InputOptions.OnSubmit(Query). Does not set Selected and does not fire OnSelect |
| Clear button | Clears Query, clears the selection depending on ResetModelValueOnClear, and calls OnClear. OnValuesChange fires if the multiple selection actually changed |
| Removing in multiple selection | Click a tag's × or press Backspace on an empty draft. Selection changes are reported through OnValuesChange |

By default, only suggestions can be selected and free text is not submitted. In multiple selection, values appear as small tags. Only the × button removes a tag, and the × is still in the Tab order. There is currently no arrow-key navigation between tags.

When nothing matches, or all suggestions are disabled, EmptyText is shown; the default messages distinguish "no data" from "no matching data". Up and Down can move to suggestions that are scrolled out of view, and focus stays in the input after a selection. Escape or blur closes the panel. The empty state cannot be selected, does not take focus, and does not intercept outside clicks.

When an option's ID, Label, Description, or Disabled changes, the open panel closes, and it uses the new data when reopened. This prevents accepting a stale suggestion. The component does not cancel network requests itself. For multiple selection, creating options, autocomplete, and reset policies, see [InputMenu values and search](options.md#inputmenu-values-and-search).

InputMenu supports Input's four parts plus Panel, Item, and EmptySlot. Icons in rows with a description align with the first line. The empty message appears only when nothing matches; suggestions that are all disabled are still shown normally. The inherited Style receives only the input and its icons; style the panel with Popup.Decoration.Style. New code should use InputMenu with its full Options directly. SearchSelect keeps the older behavior of filtering by name and description and opening on click.

### Option row appearance

Select, InputMenu, SelectMenu, Listbox, menus, and CommandPalette share the same option row style:

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| Padding and gap (DIP) | 4 | 6 | 6 | 8 | 8 |
| Font size / line height (DIP) | 12/16 | 12/16 | 14/20 | 14/20 | 16/24 |
| Icon size (DIP) | 16 | 16 | 20 | 20 | 24 |

Rows have a 6 DIP corner radius. A row highlighted by keyboard or pointer uses Elevated at 50% as its background and Highlighted for its text, and its leading icon changes from Dimmed to Text. Selected items are marked with a check mark at the end of the row, without an extra background.

The description appears on the line below the label, at the same font size, in the Muted color. Group headings are semibold and one size smaller than items. The empty state is centered and uses the Muted color.

## Popover and ScrollArea

### Popover: show content next to a control

```go
anchor := nui.Button(c, theme, "Help", nui.ButtonOptions{Variant: nui.Outline}).
    OnClick(func() { app.Help = !app.Help })
nui.Popover(c, theme, anchor, &app.Help, nui.PopoverOptions{
    Label: "Help panel",
    Width: nui.Set(float32(300)),
    Placement: nui.BottomStart,
}, func(inner nui.Theme) {
    ui.Text(c, "Explanation").TextColor(inner.Muted)
    nui.Button(c, inner, "Close", nui.ButtonOptions{}).
        OnClick(func() { app.Help = false })
})
```

The anchor must be an element from the current build, and open must be a non-nil pointer you keep between builds. The popover closes if the anchor is invalid or disabled. The return value is this build's panel, or a zero-value Element when closed; do not keep it across builds.

Popover is non-modal: an outside click closes it and still reaches its target. Escape closes it and returns focus to where it was opened. Tab is not trapped inside the panel. Build the content with the Theme passed to the callback so the local theme is preserved.

| Option | Default behavior |
| --- | --- |
| Placement | BottomStart. Supports start, center, and end on each of the four sides, for 12 positions |
| Width / MatchAnchor | Width follows the content by default; you can set a width or match the anchor |
| MaxHeight | 320 DIP. Limits only the height and does not add scrolling |
| Offset | 8 DIP from the anchor |
| Decoration | Use the Panel part to style the appearance |

For top and bottom placements, start and end follow the text direction; left and right placements use physical screen directions. MyGo's AttachTo keeps the popover clear of the window edges.

By default, a generic Popover has the theme background, a 1 DIP border, a 6 DIP corner radius, an 8 DIP content gap, a shadow, and 16 DIP of padding. To make content reach the edges, set the padding of UI[Panel] to 0. Select, InputMenu, and menus use the same base style but usually with a 240 DIP maximum height, 4 DIP of padding, no gap, and at least the trigger's width. For the menus' own height options, see [DropdownMenu and ContextMenu](overlays-navigation.md#dropdownmenu-and-contextmenu).

The Panel's font, text color, line height, and Gap apply directly to plain content. Native child controls keep using the local theme passed to them, and their own explicit text styles take precedence. Changing the anchor or panel size keeps the content's identity, drafts, and selection. If focus is inside the popover when it closes, focus returns to the current anchor; if it closes because the user clicked another control, that control keeps focus.

The entrance animation respects reduced motion, and changing that preference while the popover is open does not replay it. Closing unmounts the popover immediately; there is no exit animation as in upstream. Collision handling uses MyGo's 4 DIP edge margin, while upstream defaults to 8 px. When there is no room on either side, a side placement can still overflow the window; constrain the content size or use a top or bottom placement.

On open, the panel fades in from 95% scale within 100 ms, scaling from the side nearest the anchor. With reduced motion, it appears immediately. Closing unmounts the panel at once and restores focus, with no exit animation. A selection list's own Panel override is applied last, and explicit padding in Popup.UI also takes effect.

### ScrollArea: give content a scrolling region

ScrollArea scrolls in one of three directions: ScrollVertical (the default), ScrollHorizontal, or ScrollBoth. You can pass a `*ui.ScrollState` to read the scroll position.

Each instance uses its own persistent state. X and Y are offsets in DIP; assign them directly or call ScrollTo. MaxX and MaxY are computed by layout. Out-of-range values are clamped, and when the direction changes, any axis that is no longer available resets to zero. In RTL, X is measured from the right edge. Scrolling() reports only programmatic smooth scrolling, not every wheel action.

**Set Height, Fill, or Grow on the returned outer element** to constrain the scroll area's size. Put content padding and gaps on the Body part. The parts are Root and Body.

Horizontal scrolling uses a row layout; vertical and two-way scrolling use a column layout. Root's corner radius clips the content, and changing the scrollbar style does not reset the offset. The root container draws the theme background by default; for a transparent background, set Root.Background explicitly.

ScrollbarColor defaults to BorderAccented, and ScrollbarWidth defaults to MyGo's 6 DIP. A width of 0 or less also uses the default and does not hide the scrollbar. An explicit transparent color only hides the drawing; the hit area remains. The native Scroll handles the wheel, dragging, and scroll bounds. When an inner area reaches its bound, unconsumed wheel input can reach the outer area.

ScrollArea supports nesting and two-way scrolling. It does not provide upstream's items/virtualize, edge fade shadows, infinite loading, or custom scrollbar interaction. All content is built every time, so use the virtual lists in [Listbox or Table](forms-data.md) for long lists. Put long content in a Popover inside a ScrollArea yourself.

## Related components

For complex forms, tri-state checkboxes, range sliders, dates, and data display, see [Forms, dates, and data components](forms-data.md). For dialogs, menus, and notifications, see [Overlays, menus, and navigation](overlays-navigation.md). For theme settings, see [Themes and basic components](foundations.md).
