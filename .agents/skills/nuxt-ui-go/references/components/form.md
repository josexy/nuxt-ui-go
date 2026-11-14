<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Form

Validate fields and handle a successful submission.

- Category: Forms and data
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `Form`
- Types: `FormOptions`, `FormState`
- Theme key: `nui.ComponentForm` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Keep a FormState and register fields through FormField.Name / Value. Validate returns errors keyed by field name; an empty name represents a form-wide error. Valid submissions call OnSubmit; invalid submissions call OnError once. Submit buttons call form.Submit.

## Things to know

Explicit buttons simulate completion without network requests. Deliver CompleteAsync on the UI thread. Value, draft, visibility or disabled-state changes reject stale results. Call Cancel before unmounting or closing and cancel actual I/O in the application. Business validation runs on submit or explicit Validate; Required does not validate by itself, and nested forms do not merge validation automatically.

## API

```go
// Hidden fields retain values/dirty state but do not validate or submit. Reappearing fields retain identity.
func Form(c *ui.Context, t Theme, s *FormState, o FormOptions, content func(Theme)) ui.Element
```

### FormOptions

```go
type FormOptions struct {
	Label    string
	Disabled bool
	// LoadingAuto defaults to true; pending async work disables the form's controls.
	LoadingAuto Optional[bool]
	Validate    FormValidator
	OnSubmit    func(FormValues)
	OnError     func(FormErrors)
	Decoration
}
```

### FormState

```go
type FormState struct {
	Fields map[string]*FormFieldState

	Submitting bool
	// contains filtered or unexported fields
}

// BeginAsync and CompleteAsync run on the UI thread. Workers receive only ticket.Values.
func (s *FormState) BeginAsync() FormTicket

func (s *FormState) Cancel()

// ClearErrors clears all validation errors, or just the named fields.
// It does not make an invalid editor draft valid.
func (s *FormState) ClearErrors(names ...string)

func (s *FormState) CompleteAsync(ticket FormTicket, errors FormErrors) bool

// Errors returns a copy of errors for enabled, visible fields and form-wide errors.
func (s *FormState) Errors() FormErrors

func (s *FormState) Field(name string) *FormFieldState

func (s *FormState) Register(name, value string) *FormFieldState

func (s *FormState) Reset()

// SetErrors replaces validation errors; invalid editor drafts still take precedence.
func (s *FormState) SetErrors(errors FormErrors)

func (s *FormState) Submit()

func (s *FormState) Touch(name string)

func (s *FormState) Validate(validate FormValidator) bool

func (s *FormState) Values() FormValues
```

### Related types

```go
type FormValidator func(FormValues) FormErrors
```

```go
// FormValues are explicit serialized values, independent of widget representation.
type FormValues map[string]string
```

```go
type FormErrors map[string]string
```

```go
type FormFieldState struct {
	Disabled              bool
	Value, Initial, Error string

	Dirty, Touched, Visible bool
	Handle                  ui.Handle
	// contains filtered or unexported fields
}
```

```go
type FormTicket struct {
	Values FormValues
	// contains filtered or unexported fields
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Forms, dates, and data components

#### Form

##### Register fields and submit

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

##### Validating, hiding, and resetting

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

##### Async validation

1. On the UI thread, call `ticket := form.BeginAsync()` to get a snapshot of the current fields.
2. The background task reads only `ticket.Values` and performs validation.
3. In a `Window.Update` callback, call `form.CompleteAsync(ticket, errors)`.
4. A return value of true means the result is still current and has no errors; your app then saves.

CompleteAsync does not call OnSubmit again. A change to a field value, the raw edit text, or visibility invalidates the old ticket, as do Cancel and a new request. `Cancel()` only makes the form reject old results; your app must still cancel the context of any actual I/O.

LoadingAuto defaults to true: after BeginAsync, inputs and buttons inside the form are disabled, and the field snapshot is kept. To allow editing while waiting, set `LoadingAuto: nui.Set(false)`; editing then invalidates the old ticket. Both modes ignore repeated Submit calls, and a submission made while disabled does not run after the form is re-enabled. Put cancel and finish buttons outside the Form.

Validate, SetErrors, ClearErrors, and Reset cancel the old ticket so an older result cannot overwrite newer validation or manually set errors. Call Cancel before unmounting the component or closing the window; the state object cannot detect every unmount path on its own. Background work may only use ticket.Values, and results must be received in Window.Update.

Root is the only style part, with a default vertical gap of 16 DIP; its font and text color are set before the body is built. There is no HTML form submission, schema conversion, automatic debounced validation, or Nuxt nested path handling.

Built-in inputs, selects, and single-value toggles are automatically associated with the field Handle; in a composite group, only the first control is associated. For a custom control, bind `form.Field(name).Handle` to the actual focusable control.

Its Confirm project button only simulates local confirmation and sends no network request.

#### ColorPicker (excerpt)

When Label is empty, it is inherited from FormField. Arrow keys adjust saturation and value in the SV area, the sliders keep their native keyboard behavior, and the internal names follow the window language. Each RGBA channel is an integer up to 255. HEX accepts 3-, 4-, 6-, or 8-digit forms. An invalid HEX value keeps the last valid color, and the canonical text is restored on blur following the underlying native rules. Form labels and error focus use the first RGBA number input; there is no public persistent Handle for the internal SV area or HEX field.

#### Style part index (excerpt)

| Component | Parts |
| --- | --- |
| Form | Root |

### Options and defaults

#### Common options by component (excerpt)

FormOptions.LoadingAuto defaults to true: while an async operation runs, form controls are disabled and a snapshot of the fields is kept. Set it to `nui.Set(false)` to allow editing during the operation, which invalidates the earlier result. OnError is called once when submit validation fails, including global errors. For submit, cancel, and reset rules, see Form.

### Inputs, selection, and fields

#### State, themes, and events › Focus and callbacks (excerpt)

Since MyGo v0.9.0, focus moved by the keyboard or by a component scrolls the scroll containers around the newly focused control as little as needed to show it, as the web's `element.focus()` does. This covers roving focus in Tabs and menus, Form focusing its first invalid field, and focus returning to a trigger when Modal or DropdownMenu closes. A target that is already fully visible does not scroll. A control that a component keeps focused across builds does not undo the user's scrolling, and a pointer press focuses a control where it is without scrolling. To move focus from app code without scrolling, pass `ui.FocusOptions` to the element's Focus:

```go
cell := ui.Box(c.Key(id)).Focusable()
if app.restoreFocus == id {
	cell.Focus(ui.FocusOptions{PreventScroll: true})
}
```

#### FormField and FieldGroup › FormField: labels, descriptions, and errors (excerpt)

`Required` shows a red asterisk and adds a "required" description to supported controls. **Your app or Form still performs the actual validation**; the underlying controls currently have no native required flag of their own.

Error is a string and takes precedence over Help. When it is empty, the Form error for a touched field can be used instead. There is no equivalent of upstream `error=true` (highlight only), `error=false` (hide errors), `errorPattern` regex matching, or field-level deferred validation. Field names must match exactly, and Form.Validate or your app performs validation. In this native implementation, Hint can be shown without a Label.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct {
	Form                  nui.FormState
	Ticket                nui.FormTicket
	Name, Status          string
	Requests, ErrorEvents int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	nui.Form(c.Key("form"), t, &s.Form, nui.FormOptions{Label: "Project form", LoadingAuto: nui.Set(true), Validate: func(values nui.FormValues) nui.FormErrors {
		if strings.TrimSpace(values["name"]) == "" {
			return nui.FormErrors{"name": "Enter a project name"}
		}
		return nil
	}, OnError: func(nui.FormErrors) { s.ErrorEvents++; s.Status = "Validation failed" }, OnSubmit: func(values nui.FormValues) {
		s.Requests++

		s.Status = "Saved: " + values["name"]
	}}, func(t nui.Theme) {
		nui.FormField(c.Key("field"), t, nui.FormFieldOptions{Name: "name", Value: s.Name, Label: "Project name", Required: true}, func(t nui.Theme) {
			nui.Input(c.Key("name"), t, &s.Name, nui.InputOptions{Placeholder: "A new project"})
		})
		nui.Button(c.Key("submit"), t, "Save project", nui.ButtonOptions{}).OnClick(s.Form.Submit)
	})
	if s.Form.Submitting {
		ui.Row(c.Key("async-actions")).Wrap().Gap(8).Children(func() {
			nui.Button(c.Key("complete"), t, "Complete request", nui.ButtonOptions{}).OnClick(func() {
				if s.Form.CompleteAsync(s.Ticket, nil) {
					s.Status = "Saved: " + s.Ticket.Values["name"]
				} else {
					s.Status = "Stale result rejected"
				}
			})
			nui.Button(c.Key("cancel"), t, "Cancel request", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Form.Cancel(); s.Status = "Request cancelled" })
			nui.Button(c.Key("repeat"), t, "Repeat submit", nui.ButtonOptions{Variant: nui.Outline}).OnClick(s.Form.Submit)
		})
	}
	if message := s.Form.Errors()[""]; message != "" {
		ui.Text(c.Key("global-error"), message).TextColor(t.Error)
	}
	nui.Button(c.Key("reset"), t, "Restore project name", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Name = "Restored project"; s.Form.Reset(); s.Status = "Form reset" })
	ui.Text(c.Key("status"), s.Status)
	ui.Textf(c, "Requests: %d · Errors: %d · Dirty: %t", s.Requests, s.ErrorEvents, s.Form.Field("name").Dirty)
}

func main() {
	s := &example{}

	defer s.Form.Cancel()
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Form · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
