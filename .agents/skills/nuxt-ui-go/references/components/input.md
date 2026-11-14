<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Input

Edit one line of text with read-only, clear and loading options.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.1.0 (MyGo v0.9.0)
- Functions: `Input`
- Types: `InputOptions`
- Theme key: `nui.ComponentInput` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Pass a string pointer to `Input`; the component updates that value. Keep the string outside the view function so UI updates do not reset it. Label names the control and Placeholder gives an input hint.

## Things to know

ReadOnly allows selection and copying but prevents editing. Disabled prevents interaction. Loading shows an indicator while allowing editing. Switching avatar or loader placement preserves text, focus and selection. An explicit Leading icon takes priority over LoadingTrailing. MaxLength limits inserted UTF-16 units; external assignments are neither truncated nor reported through OnChange.

## API

```go
func Input(c *ui.Context, t Theme, value *string, o InputOptions) ui.Element
```

### InputOptions

```go
type InputOptions struct {
	Label, Placeholder, ErrorMessage                 string
	Disabled, ReadOnly, Invalid, Password, Clearable bool
	ClearLabel                                       string
	Color                                            Color
	Variant                                          Variant
	Size                                             Size
	Leading, Trailing                                *ui.SVG
	Avatar                                           *AvatarOptions // Used when there is no leading icon or leading loader.
	Loading                                          bool
	LoadingIcon                                      *ui.SVG
	LoadingTrailing                                  bool
	Highlight                                        bool
	MaxLength                                        int // UTF-16 units; 0 is unlimited.
	OnChange, OnSubmit                               func(string)
	OnClear                                          func()
	UI                                               Styles
	// Style receives current-build parts; editor is where handles and editing actions bind.
	Style func(root, editor, leading, trailing ui.Element)
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `Color` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md), `Variant` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Input controls support XS, SM, MD, LG, and XL, with MD as the default. At the default font size and density, the base single-line heights are 24, 28, 32, 36, and 40 DIP. Controls grow taller when the font's line height is larger; for example, Chinese fallback fonts on Linux can make an MD Input taller than 32 DIP.

Input, Textarea, Select, and InputMenu support these variants:

#### Input and Textarea

##### Bind a string

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

##### Common options

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

##### Editing and clearing

MyGo's TextInputBase and TextAreaBase handle text editing, selection, copy and paste, undo, and IME. MaxLength limits insertions before OnChange runs; when replacing a selection, it keeps the unchanged suffix. An insertion that is rejected entirely and leaves the value unchanged does not fire OnChange. Values assigned from outside are not truncated and do not fire input notifications.

The clear button appears only when there is content and the input is not read-only, and it inherits the disabled state. Each clear updates the bound value, calls OnChange and OnClear once each, and returns focus to the editor. Clearing changes the value programmatically, so it is not guaranteed to be a single undoable edit. If your app needs to undo a clear, save and restore the data yourself.

**Known IME issue:** If you select text, start an IME composition, and then cancel it, the selected text may be lost; undo restores it.

##### Styles and the editor Handle

Input and Textarea have the parts Root, Editor, Leading, and Trailing. `Style(root, editor, leading, trailing)` runs last and synchronously; parts that do not exist are zero-value Elements. Leading can be an icon, a loading spinner, or an avatar. Set the spinner's color with `UI.Leading.Foreground` or `UI.Trailing.Foreground`; Style.TextColor does not change the custom-drawn color. The clear button is named with ClearLabel and has no style part of its own yet.

The returned element is the outer container of the input. To bind editor focus or run editing commands, work with the editor in Style or bind a persistent Handle to it. Do not treat the outer container as the text editor.

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

#### Select and InputMenu › InputMenu: drafts versus selected values (excerpt)

InputMenuOptions embeds InputOptions and adds options such as EmptyText, Popup, AllowCustom, and OnSelect. By default, clicking only focuses the input; typing, arrow keys, or the trailing arrow open the panel. Set OpenOnClick=true to open on click. The input keeps focus. Trailing can change the icon. Loading does not disable editing or selection. Avatar, LoadingTrailing, and Highlight work as in Input.

InputMenu supports Input's four parts plus Panel, Item, and EmptySlot. Icons in rows with a description align with the first line. The empty message appears only when nothing matches; suggestions that are all disabled are still shown normally. The inherited Style receives only the input and its icons; style the panel with Popup.Decoration.Style. New code should use InputMenu with its full Options directly. SearchSelect keeps the older behavior of filtering by name and description and opening on click.

### Forms, dates, and data components

#### Lifecycle and themes (excerpt)

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. Other components are configured through their own Options at the end of this page.

#### Calendar, InputDate, and InputTime › Date and time styles and limits (excerpt)

The date and time inputs share Input's five sizes, five variants, colors, avatar, icons, loading state, disabled and read-only states, and Clearable. A leading icon or leading spinner takes precedence over Avatar. LoadingTrailing controls the spinner position, and an explicit Leading keeps a leading spinner. With Neutral, focus and Highlight use the inverted background color, and Highlight also works with None. Shadow overrides on Root from the theme or the instance are kept, and Style runs last.

### Overlays, menus, and navigation

#### DropdownMenu and ContextMenu › Filtering and controlled state (excerpt)

Filter defaults to false. When enabled, a native Input is shown and receives focus when the root menu opens. SearchTerm can bind a persistent string; when nil, MenuState.Query is used. FilterFields defaults to `label` only and also supports `description`, `id`, and `shortcut`. Matching is a case-insensitive substring match that keeps the original order. With IgnoreFilter=true, the list is shown as passed in and your app supplies results based on the query. Group labels and separators are not matched. When no action matches, a localized empty state is shown. Submenus that are not open are not searched.

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

MyGo v0.9.0, the pinned dependency, still has this known issue: **if you select text, start an IME composition, and then cancel it, the selected text is lost.** Undo restores it. Headless `ui.Tester` checks reproduce it with Input and Textarea, and will flag the change once MyGo keeps the selection.

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Value                    string
	Changes, Submits, Clears int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var parts nui.Styles

	nui.Input(c.Key("input"), t, &s.Value, nui.InputOptions{Label: "Project name", Placeholder: "Type something…", Clearable: true, UI: parts, OnChange: func(string) { s.Changes++ }, OnSubmit: func(string) { s.Submits++ }, OnClear: func() { s.Clears++ }}).FillWidth()
	ui.Textf(c, "Value: %s", s.Value)
	ui.Textf(c, "Changes: %d · Submits: %d · Clears: %d", s.Changes, s.Submits, s.Clears)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Input · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
