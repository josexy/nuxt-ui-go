package nui

import "github.com/egoist/mygo/ui"

// FormValues are explicit serialized values, independent of widget representation.
type FormValues map[string]string
type FormErrors map[string]string
type FormValidator func(FormValues) FormErrors
type FormFieldState struct {
	Disabled                              bool
	Value, Initial, Error                 string
	controlError, editing, initialEditing string
	controlErrorKey                       string
	hasEditor, editorDirty                bool
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
	if !f.initialized {
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
func (s *FormState) Reset() {
	for _, f := range s.Fields {
		f.Initial = f.Value
		f.Dirty = false
		f.initialEditing = f.editing
		f.editorDirty = false
		f.Touched = false
		f.Error = ""
	}
	s.Cancel()
}
func (s *FormState) Submit() { s.requested = true }
func (s *FormState) Cancel() { s.version++; s.Submitting = false; s.requested = false }
func (s *FormState) apply(errors FormErrors) bool {
	valid := true
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
	return s.apply(errors)
}

type FormOptions struct {
	Label    string
	Disabled bool
	Validate FormValidator
	OnSubmit func(FormValues)
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
	inner := t
	inner.form = s
	inner.field.Disabled = inner.field.Disabled || o.Disabled
	root := ui.Column(c).MinWidth(0).Role(ui.RoleGroup).Label(o.Label).Disabled(o.Disabled).Children(func() {
		if content != nil {
			content(inner)
		}
	})
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
	if s.requested && !o.Disabled {
		s.requested = false
		if s.Validate(o.Validate) && o.OnSubmit != nil {
			o.OnSubmit(s.Values())
		}
	}
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
	finish(t, ComponentForm, Root, root, PartStyle{Gap: Set(t.spacing(16))}, o.Decoration)
	return root
}
