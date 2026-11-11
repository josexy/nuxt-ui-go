<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# FormField

Add a label, help text and errors to an input.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `FormField`
- Types: `FormFieldOptions`
- Theme key: `nui.ComponentFormField` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Build the input inside the `FormField` callback and use the theme passed to that callback. This connects the label and passes size, disabled and error state to the input.

## Things to know

Required displays a required marker without validating the value. Use Form.Validate or application validation to check input. Without Label, no isolated asterisk is drawn, while the required description still reaches the control. Error takes priority over Help; the label targets the first bound control. A standalone Hint is a native extension. Field names match exactly; errorPattern and per-field delayed validation are not provided.

## API

```go
// FormField keeps the label and supporting text separate from the editor column.
// Only model data and the first control's handle cross build passes.
func FormField(c *ui.Context, t Theme, opts FormFieldOptions, content func(Theme)) ui.Element
```

### FormFieldOptions

```go
type FormFieldOptions struct {
	// Name and Value register a serialized value with the enclosing Form.
	Name, Value                           string
	Label, Description, Help, Hint, Error string
	Required, Disabled                    bool
	Size                                  Size
	Orientation                           Optional[Orientation] // Defaults to Vertical.
	Decoration
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Size precedence is: the instance's Size → the FormField/FieldGroup Size → the component's default Size. Set component defaults with calls such as `WithComponent(ComponentInput, ...)`. When a parent field or any ancestor is disabled, a child cannot re-enable itself with `Disabled: false`.

#### FormField and FieldGroup

##### FormField: labels, descriptions, and errors

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

#### Checkbox, CheckboxGroup, RadioGroup, and Switch › Options and keyboard (excerpt)

A group's visible title can come from the enclosing FormField.Label. To use one icon for all checkboxes in a group, set the same SVG as Option.Icon on every option that has no icon of its own; items with their own icons keep them.

#### Checkbox, CheckboxGroup, RadioGroup, and Switch › Icons, styles, and spacing (excerpt)

Mark a field as required with FormField.Required. It shows an asterisk and gives assistive technology a "required" hint along with the field's description and the control's description. It does not validate the selection; configure FormState validation rules to enforce it on submit.

#### Slider (excerpt)

An explicit Label is used as the name first, then the FormField label; if both are empty, the localized "Slider" name is used. The label is not drawn as visible text; Description adds a supplementary description.

For Slider and RangeSlider, Root styles the whole control, Indicator.Background sets the filled track and the thumb border, and Thumb.Background sets the thumb fill. Neutral uses InvertedBackground. A FormField error uses the Error color first, although explicit instance part colors can still override it. The unfilled track uses Theme.Accented. The track and thumb are drawn by a draw callback, so other Thumb and Indicator style fields have no effect, and Style receives only Root.

### Options and defaults

#### Limitations (excerpt)

FormField associates its label with this library's controls. If you use a custom MyGo control, your app must associate the label itself. Toast action callbacks and ButtonOptions.Style are stored with the toast, so their closures must not capture temporary Elements, Parts, or Contexts; the Style arguments are valid only during the build in which they are called.

### Advanced interaction, charts, and layout

#### FileUpload and resource ownership › The component handles selection and state; your app handles transfer (excerpt)

Files that exceed a limit or cannot be read are shown as Rejected with the reason, and the user can remove them. A non-positive MaxBytes or MaxFiles means no limit. Rejected files do not count toward the limit, while accepted files that later fail or are canceled still count. The component inherits Disabled from its parent container and FormField, as well as FormField's Size, label, and error message. The default color and size come from ComponentFileUpload.

### Forms, dates, and data components

#### Form › Register fields and submit (excerpt)

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

#### Numbers, tags, PINs, and ratings › PinInput (excerpt)

A focused empty cell hides its Placeholder and shows a blinking cursor. With reduced motion, the cursor stays solid; it is hidden when read-only, disabled, or unfocused. When FormField has an Error, every cell shows the error border. Text scaling and masking do not affect hit testing.

#### Numbers, tags, PINs, and ratings › InputRating (excerpt)

Stars are 12/16/20/24/28 DIP for sizes XS to XL, 2 DIP apart, and the component is as wide as its content. When interactive, the hovered star scales to 1.1×. With Hoverable on, hovering only previews the rating; the preview stays while the pointer crosses gaps between stars and is canceled when the pointer leaves or the rating is cleared. When FormField has an error, filled stars and the focus ring use the Error color.

#### Numbers, tags, PINs, and ratings › RangeSlider (excerpt)

Inside a FormField, clicking the label focuses the first thumb, and the field's description and error are associated with it. For orientation and keyboard behavior, see [Sliders, tabs, and accordions](slider.md).

#### Listbox, SelectMenu, and async search › Listbox: a selectable list shown inline (excerpt)

Size and Color default to MD and Primary. Instance options take precedence over the component theme, and the size can also be inherited from FormField. Highlight keeps a colored border visible. Keyboard focus shows a focus ring, which uses InvertedBackground with Neutral. When embedded in SelectMenu, the list follows the menu's size and color.

#### Calendar, InputDate, and InputTime › InputDateRange: type a start and end date (excerpt)

Inside a single FormField, an error focuses the first segment of the start date. To register errors and focus for each end separately, combine two InputDates.

#### ColorPicker (excerpt)

When Label is empty, it is inherited from FormField. Arrow keys adjust saturation and value in the SV area, the sliders keep their native keyboard behavior, and the internal names follow the window language. Each RGBA channel is an integer up to 255. HEX accepts 3-, 4-, 6-, or 8-digit forms. An invalid HEX value keeps the last valid color, and the canonical text is restored on blur following the underlying native rules. Form labels and error focus use the first RGBA number input; there is no public persistent Handle for the internal SV area or HEX field.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

When a Badge's label is empty, no text is created, no extra gap is left, and the padding becomes equal on all sides. Inside a FieldGroup or FormField, a Badge inherits the size and takes part in the group's corner rounding. If Avatar.Size is unset, badge sizes XS–XL use avatars of 16, 16, 16, 20, and 20 DIP. An explicit Leading icon takes precedence over the avatar.

### Internationalization

#### What gets translated › Defaults that are easy to confuse (excerpt)

For example, a 5-cell PinInput is named `5 characters` / `5 个字符`, or `5 digits` / `5 位数字` in numeric mode. This is the editor's accessible name; the placeholder shown in each cell still comes from `Placeholder`. An explicit `Label` or FormField label takes precedence.

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

type example struct{ Value string }

func (s *example) view(c *ui.Context, t nui.Theme) {
	orientation := nui.Vertical

	var parts nui.Styles

	nui.FormField(c.Key("field"), t, nui.FormFieldOptions{Label: "Project", Help: "Use a name your team will recognize.", Orientation: nui.Set(orientation), Decoration: nui.Decoration{UI: parts}}, func(ft nui.Theme) {
		nui.Input(c.Key("input"), ft, &s.Value, nui.InputOptions{Label: "Project", Placeholder: "Project name"}).Width(240)
	}).FillWidth()
	ui.Text(c, "Value: "+s.Value)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "FormField · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
