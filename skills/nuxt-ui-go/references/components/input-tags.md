<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# InputTags

Turn typed or pasted text into removable tags.

- Category: Forms and data
- Version: nuxt-ui-go v0.1.1 plus unreleased changes (MyGo v0.10.0)
- Functions: `InputTags`, `AddTags`
- Types: `InputTagsOptions`
- Theme key: `nui.ComponentInputTags` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Bind a `*[]string` for tags and use `TagsState.Query` for the active input draft; tags and editor wrap together. `ConvertValue` normalizes new values, while `DisplayValue` customizes display labels without changing underlying values.

## Things to know

- Deduplication is case-sensitive and runs after `ConvertValue`. `MaxTags <= 0` is unlimited; `MaxLength` limits draft UTF-16 units.
- Batch paste emits `OnTagsChange` once; duplicate or over-limit items are rejected instead of inserted as plain text.
- Supports string slices only (no object models or regex delimiters).

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

#### Lifecycle and themes (excerpt)

InputNumber, InputDate, and InputTime share Input's sizes, five field variants, colors, read-only and disabled states, icons, and error settings. InputNumber, InputTags, PinInput cells, the SelectMenu trigger, Listbox and Tree rows, Calendar cells and card-variant CheckboxGroup items fade their background, border and focus ring over `Motion.Transition` (150 ms) at the Full and Reduced motion levels, as Nuxt UI's `transition-colors`; text colours change at once, and the Off motion level shows every change at once. InputDate and InputTime fields do not fade. Other components are configured through their own Options at the end of this page.

#### Numbers, tags, PINs, and ratings › InputTags

InputTags binds a `*[]string`, and TagsState.Query stores the draft that has not been added yet. Enter adds a tag, and tags and the editor wrap together on the same lines. With the cursor at the start of the draft, Backspace first selects the last tag, and a second Backspace or Delete removes it. Left and Right move the tag selection, and Home and End go to the first and last tags while a tag is selected. The delete button uses DeleteIcon, × by default; after a deletion, focus returns to the input.

The component trims leading and trailing whitespace from a candidate, converts the string with ConvertValue, then removes case-sensitive duplicates and checks MaxTags. DisplayValue changes only the text of the tag and its delete button; the value and stable identity are unchanged. OnTagsChange returns a copy of the slice. OnAddTag, OnRemoveTag, and OnInvalid report each added, removed, or rejected candidate. A successful batch paste fires OnTagsChange at most once, while the per-item events are still reported separately.

Pasting is not split into tags by default. With `AddOnPaste: nui.Set(true)`, pasted text is split on Separators, which by default contains only a comma. To also split on semicolons or newlines, set `Separators: ",;\n"`. A batch paste keeps the existing draft. For more about when tags are added and how duplicates are handled, see [Switches, tags, and ratings](switch.md).

Tags use the Elevated background with an Accented inner border, and a selected tag uses the Accented background. The font size follows the field size, and long tags can wrap. When the editor or a tag button has focus, the frame shows the accent color and a focus ring; in the error state it uses the error color, and None shows no focus ring. Root is the whole field and Body is the wrapping area. Item, Label, Close, and Indicator are the tag container, its text, the delete button, and its icon. Style's root also refers to the whole field. Object models, regular expression separators, drag-and-drop reordering, and remote entity matching are not supported.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| InputTags | Root, Editor, Leading, Trailing, Body, Item, Label, Close, Indicator |

InputNumber's and InputTags' `UI` support every part in the table, while the embedded `InputOptions.Style` still receives only the Root, Editor, Leading, and Trailing nodes. Configure InputNumber's button callbacks through IncrementOptions and DecrementOptions. Table's header divider paints one device pixel and takes its stroke color from `UI[SeparatorSlot].Background` (default BorderAccented). Its replacement covers MyGo's one-DIP divider using the theme background or an opaque Root/Body/Panel UI background; custom translucent surfaces or raw Style background changes should provide their own SeparatorSlot material. Table's other internal native regions and ColorPicker's unexposed internal parts use the Native theme. RangeSlider draws its track and thumbs from the Theme and Color; apart from Indicator.Background and Thumb.Background, other Indicator and Thumb fields have no effect. Parts that are not exposed cannot be changed by writing their names yourself. For the precedence of global WithComponent settings and instance UI/Style, see [Style overrides](theme.md).

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

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Input, Textarea, InputNumber, InputTags, PinInput cells, and the Select, SelectMenu and InputMenu triggers fade their background, border and focus ring over `Motion.Transition` (150 ms) when hover, focus or invalid state changes, as Nuxt UI's `transition-colors`. The fade also runs under reduced motion; `Motion.Transition = 0` or the Off motion level shows the change at once. Text and icon colours change at once.

#### Input and Textarea › Editing and clearing (excerpt)

MyGo's TextInputBase and TextAreaBase handle text editing, selection, copy and paste, undo, and IME. The native editor applies MaxLength while it inserts text, before OnChange runs. Replacing a selection can use the freed space, and a partly accepted paste never splits a character. An insertion that is rejected entirely leaves the caret and undo history unchanged and fires no OnChange. Values assigned from outside are not truncated and do not fire input notifications. Changing the limit applies to the next input without resetting the draft or caret. InputTags and InputMenu use the same editor.

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
	nui "github.com/josexy/nuxt-ui-go/ui"
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
