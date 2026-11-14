package nui

import (
	"maps"

	"github.com/egoist/mygo/ui"
)

// FormValues are explicit serialized values, independent of widget representation.
type FormValues map[string]string
type FormErrors map[string]string
type FormValidator func(FormValues) FormErrors
type FormFieldState struct {
	Disabled                              bool
	Value, Initial, Error                 string
	controlError, editing, initialEditing string
	controlErrorKey                       string
	hasEditor, editorDirty, resetPending  bool
	Dirty, Touched, Visible               bool
	Handle                                ui.Handle
	focused, initialized                  bool
}
type FormTicket struct {
	owner   *byte
	version uint64
	Values  FormValues
}
type FormState struct {
	identity              *byte
	Fields                map[string]*FormFieldState
	order                 []string
	version               uint64
	requested, focusError bool
	dispatching           bool
	unboundErrors         FormErrors
	Submitting            bool
}

func (s *FormState) Field(name string) *FormFieldState {
	if name == "" {
		panic("nui: empty form field name")
	}
	if s.Fields == nil {
		s.Fields = map[string]*FormFieldState{}
	}
	if s.Fields[name] == nil {
		s.Fields[name] = &FormFieldState{}
	}
	return s.Fields[name]
}
func (s *FormState) Register(name, value string) *FormFieldState {
	f := s.Field(name)
	if f.Visible {
		panic("nui: duplicate form field: " + name)
	}
	if message, ok := s.unboundErrors[name]; ok {
		f.Error, f.Touched = message, true
		delete(s.unboundErrors, name)
	}
	if !f.initialized || f.resetPending {
		f.Initial = value
		f.initialized = true
	}
	if f.Value != value {
		f.Value = value
		s.version++
		s.Submitting = false
	}
	f.Dirty = f.Value != f.Initial || f.editorDirty
	f.Visible = true
	s.order = append(s.order, name)
	return f
}
func (s *FormState) Values() FormValues {
	out := FormValues{}
	for name, f := range s.Fields {
		if f.Visible && !f.Disabled {
			out[name] = f.Value
		}
	}
	return out
}
func (s *FormState) Touch(name string) { s.Field(name).Touched = true }

// Errors returns a copy of errors for enabled, visible fields and form-wide errors.
func (s *FormState) Errors() FormErrors {
	errors := FormErrors{}
	maps.Copy(errors, s.unboundErrors)
	for name, f := range s.Fields {
		if f.Visible && !f.Disabled && f.Error != "" {
			errors[name] = f.Error
		}
	}
	return errors
}

// SetErrors replaces validation errors; invalid editor drafts still take precedence.
func (s *FormState) SetErrors(errors FormErrors) {
	s.Cancel()
	s.apply(errors)
}

// ClearErrors clears all validation errors, or just the named fields.
// It does not make an invalid editor draft valid.
func (s *FormState) ClearErrors(names ...string) {
	s.Cancel()
	clearName := func(name string) {
		delete(s.unboundErrors, name)
		if f := s.Fields[name]; f != nil {
			f.Error = f.controlError
		}
	}
	if len(names) == 0 {
		clear(s.unboundErrors)
		for name := range s.Fields {
			clearName(name)
		}
		return
	}
	for _, name := range names {
		clearName(name)
	}
}

func (s *FormState) Reset() {
	for _, f := range s.Fields {
		f.Initial = f.Value
		f.Dirty = false
		f.initialEditing = f.editing
		f.editorDirty = false
		f.Touched = false
		f.Error = ""
		f.resetPending = true
	}
	clear(s.unboundErrors)
	s.Cancel()
}
func (s *FormState) Submit() {
	if !s.Submitting && !s.dispatching {
		s.requested = true
	}
}
func (s *FormState) Cancel() {
	s.version++
	s.Submitting = false
	s.requested = false
	s.focusError = false
}
func (s *FormState) apply(errors FormErrors) bool {
	valid := true
	s.unboundErrors = FormErrors{}
	for name, message := range errors {
		if f := s.Fields[name]; (f == nil || !f.initialized) && message != "" {
			s.unboundErrors[name] = message
			valid = false
		}
	}
	for name, f := range s.Fields {
		f.Error = errors[name]
		if f.controlError != "" {
			f.Error = f.controlError
		}
		if f.Visible && !f.Disabled {
			f.Touched = true
			if f.Error != "" {
				valid = false
			}
		}
	}
	s.focusError = !valid
	return valid
}
func (s *FormState) Validate(validate FormValidator) bool {
	s.Cancel()
	if validate == nil {
		return s.apply(nil)
	}
	return s.apply(validate(s.Values()))
}

// BeginAsync and CompleteAsync run on the UI thread. Workers receive only ticket.Values.
func (s *FormState) BeginAsync() FormTicket {
	s.version++
	s.Submitting = true
	if s.identity == nil {
		s.identity = new(byte)
	}
	return FormTicket{s.identity, s.version, s.Values()}
}
func (s *FormState) CompleteAsync(ticket FormTicket, errors FormErrors) bool {
	if !s.Submitting || ticket.owner != s.identity || ticket.version != s.version {
		return false
	}
	s.Submitting = false
	valid := s.apply(errors)
	if valid {
		s.markClean()
	}
	return valid
}

func (s *FormState) markClean() {
	for _, f := range s.Fields {
		if f.Visible && !f.Disabled {
			f.Initial, f.initialEditing = f.Value, f.editing
			f.Dirty, f.editorDirty = false, false
		}
	}
}

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

// Hidden fields retain values/dirty state but do not validate or submit. Reappearing fields retain identity.
func Form(c *ui.Context, t Theme, s *FormState, o FormOptions, content func(Theme)) ui.Element {
	if s == nil {
		panic("nui: Form requires state")
	}
	previous := s.Values()
	for _, f := range s.Fields {
		f.Visible = false
	}
	s.order = s.order[:0]
	root := ui.Column(c).MinWidth(0).Role(ui.RoleGroup).Label(o.Label).Disabled(o.Disabled || t.field.Disabled)
	disabled := root.IsDisabled()
	root.Disabled(disabled || s.Submitting && valueOr(o.LoadingAuto, true))
	finish(t, ComponentForm, Root, root, PartStyle{Gap: Set(t.spacing(16)), Font: Set(t.Font), Foreground: Set(t.Text)}, o.Decoration)
	inner := t
	inner.form = s
	inner.field.Disabled = disabled
	root.Children(func() {
		if content != nil {
			content(inner)
		}
	})
	for _, f := range s.Fields {
		f.resetPending = false
		if !f.Visible || f.Disabled {
			f.Handle.CancelFocus(c)
		}
	}
	current := s.Values()
	changedVisibility := len(previous) != len(current)
	for k := range previous {
		if _, ok := current[k]; !ok {
			changedVisibility = true
		}
	}
	if changedVisibility {
		s.version++
		s.Submitting = false
	}
	if s.requested {
		s.requested = false
		if !disabled && !s.Submitting && !s.dispatching {
			s.submit(o)
			// Validation runs after the fields were built. Rebuild their error
			// text and loading styles even when focus stays on the same button.
			c.Services().Invalidate()
		}
	}
	root.Disabled(disabled || s.Submitting && valueOr(o.LoadingAuto, true))
	if s.focusError {
		for _, name := range s.order {
			f := s.Fields[name]
			if f.Error != "" && !f.Disabled {
				f.Handle.Focus(c)
				break
			}
		}
		s.focusError = false
	}
	return root
}

func (s *FormState) submit(o FormOptions) {
	s.dispatching = true
	defer func() { s.dispatching = false }()
	if !s.Validate(o.Validate) {
		if o.OnError != nil {
			o.OnError(s.Errors())
		}
		return
	}
	if o.OnSubmit != nil {
		o.OnSubmit(s.Values())
		if !s.Submitting {
			s.markClean()
		}
	}
}
