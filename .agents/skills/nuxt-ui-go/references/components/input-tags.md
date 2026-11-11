<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputTags

Turn typed text into removable tags.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `InputTags`, `AddTags`
- Types: `InputTagsOptions`
- Theme key: `nui.ComponentInputTags` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Bind a string slice and keep the draft in TagsState.Query. Tags and the editor wrap together. ConvertValue transforms new values; DisplayValue changes labels without replacing their identity.

## Things to know

Deduplication is case-sensitive after conversion. MaxTags<=0 is unlimited; MaxLength limits draft UTF-16 units, with 0 unlimited. A successful paste batch emits OnTagsChange once and reports each added or rejected non-empty candidate separately. Paste preserves the existing draft; rejected candidates are not inserted as ordinary text. Values are strings, without object models or regex delimiters.

## API

```go
func InputTags(c *ui.Context, t Theme, value *[]string, s *TagsState, o InputTagsOptions) ui.Element

// AddTags trims and deduplicates without changing the order of existing tags.
func AddTags(tags []string, text, separators string, limit int) []string
```

### InputTagsOptions

```go
type InputTagsOptions struct {
	InputOptions
	MaxTags                          int
	Separators                       string
	AddOnTab, AddOnBlur, Duplicate   bool
	AddOnPaste                       Optional[bool]
	MaxLength                        int // Native editing limit in UTF-16 units; 0 is unlimited.
	OnTagsChange                     func([]string)
	DeleteIcon                       *ui.SVG
	ConvertValue, DisplayValue       func(string) string
	OnAddTag, OnRemoveTag, OnInvalid func(string)
}
```

### Related types

```go
type TagsState struct {
	Query  string
	Handle ui.Handle
	// contains filtered or unexported fields
}
```

Types documented in other references: `InputOptions` → [input.md](input.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Numbers, tags, PINs, and ratings › InputTags

InputTags binds a `*[]string`, and TagsState.Query stores the draft that has not been added yet. Enter adds a tag, and tags and the editor wrap together on the same lines. With the cursor at the start of the draft, Backspace first selects the last tag, and a second Backspace or Delete removes it. Left and Right move the tag selection, and Home and End go to the first and last tags while a tag is selected. The delete button uses DeleteIcon, × by default; after a deletion, focus returns to the input.

The component trims leading and trailing whitespace from a candidate, converts the string with ConvertValue, then removes case-sensitive duplicates and checks MaxTags. DisplayValue changes only the text of the tag and its delete button; the value and stable identity are unchanged. OnTagsChange returns a copy of the slice. OnAddTag, OnRemoveTag, and OnInvalid report each added, removed, or rejected candidate. A successful batch paste fires OnTagsChange at most once, while the per-item events are still reported separately.

Pasting is not split into tags by default. With `AddOnPaste: nui.Set(true)`, pasted text is split on Separators, which by default contains only a comma. To also split on semicolons or newlines, set `Separators: ",;\n"`. A batch paste keeps the existing draft. For more about when tags are added and how duplicates are handled, see [Switches, tags, and ratings](switch.md).

Tags use the Elevated background with an Accented inner border, and a selected tag uses the Accented background. The font size follows the field size, and long tags can wrap. When the editor or a tag button has focus, the frame shows the accent color and a focus ring; in the error state it uses the error color, and None shows no focus ring. Root is the whole field and Body is the wrapping area. Item, Label, Close, and Indicator are the tag container, its text, the delete button, and its icon. Style's root also refers to the whole field. Object models, regular expression separators, drag-and-drop reordering, and remote entity matching are not supported.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| InputTags | Root, Editor, Leading, Trailing, Body, Item, Label, Close, Indicator |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](theme.md).

#### Examples (excerpt)

Text input in InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, and ColorPicker is affected by the known IME cancel issue.

### Options and defaults

#### Unset, zero, and default values › Frequently checked defaults (excerpt)

| Component | Defaults |
| --- | --- |
| InputTags, InputRating, Slider | InputTags: AddOnPaste=false. InputRating: Hoverable=false. Slider: Step=1 |

#### Common options by component › Switches, tags, and ratings (excerpt)

| Component | Common options and behavior |
| --- | --- |
| InputTags | AddOnTab, AddOnBlur, and AddOnPaste control when tags are added. Duplicate, ConvertValue, and DisplayValue control values and display. MaxLength limits the text being edited |

InputTags' MaxLength limits what the user can insert, counted in UTF-16 code units; 0 means no limit. It limits the draft being edited and does not validate tag length on submit. With AddOnPaste on, pasted text is always handled as tags; text rejected as a duplicate or over the limit is not inserted as plain text either. With it off, pasting inserts plain text. With AddOnTab on, Tab and Shift+Tab commit a non-empty draft and keep focus; with an empty draft they move focus as usual.

### Inputs, selection, and fields

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
	Values                            []string
	Tags                              nui.TagsState
	Changes, Added, Removed, Rejected int
	LastRejected                      string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.InputTags(c.Key("tags"), t, &s.Values, &s.Tags, nui.InputTagsOptions{InputOptions: nui.InputOptions{Label: "Project name", Placeholder: "Type something…"}, MaxTags: 6, OnTagsChange: func([]string) { s.Changes++ }, OnAddTag: func(string) { s.Added++ }, OnRemoveTag: func(string) { s.Removed++ }, OnInvalid: func(value string) { s.Rejected++; s.LastRejected = value }})
	nui.Button(c.Key("leave"), t, "Leave tag input", nui.ButtonOptions{Variant: nui.Outline})
	ui.Textf(c, "Changes: %d · Added: %d · Removed: %d · Rejected: %d", s.Changes, s.Added, s.Removed, s.Rejected)
	if s.LastRejected != "" {
		ui.Text(c, "Last rejected: "+s.LastRejected)
	}
}

func main() {
	s := &example{}

	s.Values = []string{"go", "vue"}
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "InputTags · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
