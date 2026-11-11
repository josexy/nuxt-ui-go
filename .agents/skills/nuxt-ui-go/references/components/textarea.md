<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Textarea

Edit multiple lines and optionally grow with the content.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `Textarea`
- Types: `TextareaOptions`
- Theme key: `nui.ComponentTextarea` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Textarea binds a string and takes shared input settings through InputOptions. Rows sets the base line count, AutoResize grows with new lines, and MaxRows can cap that growth.

## Things to know

Keep text outside the view function. Changes to rows, avatar or loader placement preserve editing state. Rows≤0 uses 3. MaxRows applies only with AutoResize: ≤0 means unlimited and values below Rows use Rows. Enter and Shift+Enter insert lines without OnSubmit; the shared Password field has no effect on multiline input. There is no drag-to-resize handle.

## API

```go
func Textarea(c *ui.Context, t Theme, value *string, o TextareaOptions) ui.Element
```

### TextareaOptions

```go
type TextareaOptions struct {
	InputOptions
	Rows, MaxRows int
	AutoResize    bool
}
```

Types documented in other references: `InputOptions` → [input.md](input.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

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

##### Textarea rows

Rows defaults to 3. Without AutoResize, the textarea always shows that many rows; with it, the textarea grows with the actual line count. A MaxRows of 0 or less means no limit, and a positive value less than Rows is treated as Rows. Once the limit is reached, the editor scrolls internally.

Textarea provides plain multiline editing, without a character counter, syntax highlighting, or rich text.

The avatar and loading spinner in a Textarea align with the first line. Enter and Shift+Enter insert a newline and do not fire OnSubmit; submit with a separate button or an editor shortcut instead. The shared Password field has no effect on multiline input. There is no native resize handle; adjust the height with Rows, AutoResize, or the surrounding layout.

##### Styles and the editor Handle

Input and Textarea have the parts Root, Editor, Leading, and Trailing. `Style(root, editor, leading, trailing)` runs last and synchronously; parts that do not exist are zero-value Elements. Leading can be an icon, a loading spinner, or an avatar. Set the spinner's color with `UI.Leading.Foreground` or `UI.Trailing.Foreground`; Style.TextColor does not change the custom-drawn color. The clear button is named with ClearLabel and has no style part of its own yet.

The returned element is the outer container of the input. To bind editor focus or run editing commands, work with the editor in Style or bind a persistent Handle to it. Do not treat the outer container as the text editor.

#### FormField and FieldGroup › FieldGroup: join several controls (excerpt)

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Value           string
	Changes, Clears int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var parts nui.Styles

	nui.Textarea(c.Key("textarea"), t, &s.Value, nui.TextareaOptions{InputOptions: nui.InputOptions{Label: "Project name", Placeholder: "Type something…", Clearable: true, UI: parts, OnChange: func(string) { s.Changes++ }, OnClear: func() { s.Clears++ }}, Rows: 3, MaxRows: 4, AutoResize: true}).FillWidth()
	ui.Textf(c, "Changes: %d · Clears: %d", s.Changes, s.Clears)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Textarea · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
