package nui

import (
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type InputOptions struct {
	Label, Placeholder, ErrorMessage                 string
	Disabled, ReadOnly, Invalid, Password, Clearable bool
	ClearLabel                                       string
	Color                                            Color
	Variant                                          Variant
	Size                                             Size
	Leading, Trailing                                *ui.SVG
	Loading                                          bool
	LoadingIcon                                      *ui.SVG
	Highlight                                        bool
	MaxLength                                        int // UTF-16 units; 0 is unlimited.
	OnChange, OnSubmit                               func(string)
	OnClear                                          func()
	UI                                               Styles
	// Style receives current-build parts; editor is where handles and editing actions bind.
	Style    func(root, editor, leading, trailing ui.Element)
	errorKey string // Identifies an internal validation message, independent of locale.
}

type TextareaOptions struct {
	InputOptions
	Rows, MaxRows int
	AutoResize    bool
}

func inputRecipe(t Theme, v ComponentTheme, focused, hover, invalid bool) Styles {
	m := buttonMetrics(t, v.Size)
	root := controlStyle(t, m)
	root.Weight = Set(400)
	root.Foreground = Set(t.Highlighted)
	root.Background = Set(ui.Transparent)
	root.BorderWidth = Set(float32(0))
	root.BorderColor = Set(t.BorderAccented)
	switch v.Variant {
	case Outline:
		root.Background = Set(t.Background)
		root.BorderWidth = Set(float32(1))
	case Soft:
		root.Background = Set(t.Elevated.Alpha(.5))
		if focused || hover {
			root.Background = Set(t.Elevated)
		}
	case Subtle:
		root.Background = Set(t.Elevated)
		root.BorderWidth = Set(float32(1))
	case Ghost:
		if focused || hover {
			root.Background = Set(t.Elevated)
		}
	case None:
	default:
		panic("nui: fields support outline, soft, subtle, ghost and none")
	}
	if invalid {
		root.BorderWidth = Set(float32(1))
		root.BorderColor = Set(t.Error)
	} else if focused && root.BorderWidth.Value > 0 {
		root.BorderColor = Set(t.color(v.Color))
	}
	return Styles{Root: root, Editor: {Foreground: Set(t.Highlighted), Font: Set(t.Font), FontSize: Set(m.Font), LineHeight: Set(m.Line / m.Font)}, Leading: {Foreground: Set(t.Dimmed)}, Trailing: {Foreground: Set(t.Dimmed)}}
}
func inputDefaults(t Theme, id Component, o InputOptions) ComponentTheme {
	return selected(defaults(t, id, ComponentTheme{Color: Primary, Size: MD, Variant: Outline}), ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size), Variant: o.Variant})
}
func fieldInvalid(t Theme, o InputOptions) bool {
	return o.Invalid || o.ErrorMessage != "" || t.field.Error != ""
}
func fieldSemantics(c *ui.Context, e ui.Element, t Theme, o InputOptions) {
	bindField(c, e, t)
	if t.field.Required {
		e.Description(message(c, "required"))
	}
	if o.Label != "" {
		e.Label(o.Label)
	}
	if o.ErrorMessage != "" {
		e.Error(o.ErrorMessage)
	} else if o.Invalid && t.field.Error == "" {
		e.Error(message(c, "checkValue"))
	}
}

func Input(c *ui.Context, t Theme, value *string, o InputOptions) ui.Element {
	return inputShell(c, t, value, o, ComponentInput, 0, 0, func() ui.Element { return ui.TextInputBase(c.Key("input"), value) }, nil)
}
func Textarea(c *ui.Context, t Theme, value *string, o TextareaOptions) ui.Element {
	rows := o.Rows
	if rows <= 0 {
		rows = 3
	}
	maximum := rows
	if o.AutoResize {
		maximum = o.MaxRows
		if maximum <= 0 {
			maximum = int(^uint(0) >> 1) // MyGo requires a positive upper bound to honor the minimum.
		}
		if maximum < rows {
			maximum = rows
		}
	}
	return inputShell(c, t, value, o.InputOptions, ComponentTextarea, rows, maximum, func() ui.Element { return ui.TextAreaBase(c.Key("input"), value) }, nil)
}

// makeEditor and afterEditor execute synchronously inside the shell, never retained.
func inputShell(c *ui.Context, t Theme, value *string, o InputOptions, id Component, minRows, maxRows int, makeEditor func() ui.Element, afterEditor func(ui.Element)) ui.Element {
	if value == nil {
		panic("nui: input requires a value pointer")
	}
	config := inputDefaults(t, id, o)
	m := buttonMetrics(t, config.Size)
	root := ui.Row(c).MinWidth(0).AlignItems(ui.Center).Cursor(ui.CursorText).Disabled(o.Disabled || t.field.Disabled)
	if minRows > 0 {
		root.AlignItems(ui.Start)
	}
	var editor, leading, trailing, clear ui.Element
	root.Children(func() {
		previous := *value // Capture before TextInputBase applies this frame's bound edit.
		if o.Loading {
			icon := o.LoadingIcon
			if icon == nil {
				icon = lucide.LoaderCircle()
			}
			leading = loadingSVG(c, icon, m.Icon, t.Dimmed, t.motion(c), t.Motion.Spin)
		} else if o.Leading != nil {
			leading = ui.Icon(c, o.Leading).Size(m.Icon, m.Icon).Shrink(0)
		}
		editor = makeEditor().Grow(1).MinWidth(0).Font(t.Font).FontSize(m.Font).LineHeight(m.Line / m.Font).Placeholder(o.Placeholder).ReadOnly(o.ReadOnly).FocusRing(false)
		if minRows > 0 {
			editor.Lines(minRows, maxRows)
		}
		if o.Password {
			editor.Password()
		}
		fieldSemantics(c, editor, t, o)

		if o.OnChange != nil || o.MaxLength > 0 {
			editor.OnChange(func() {
				if o.MaxLength > 0 {
					*value = limitTextEdit(previous, *value, o.MaxLength)
				}
				if o.OnChange != nil {
					o.OnChange(*value)
				}
			})
		}
		if o.OnSubmit != nil {
			editor.OnSubmit(func() { o.OnSubmit(*value) })
		}
		if o.Clearable && *value != "" && !o.ReadOnly {
			name := o.ClearLabel
			if name == "" {
				name = message(c, "clear")
				if o.Label != "" {
					name = namedMessage(c, "clearNamed", o.Label)
				}
			}
			clear = ui.ButtonBase(c.Key("clear")).Label(name).Size(m.Icon, m.Icon).FocusRing(true).Children(func() { ui.Icon(c, lucide.X()).Size(m.Icon, m.Icon).TextColor(t.Muted) })
		}
		if o.Trailing != nil {
			trailing = ui.Icon(c, o.Trailing).Size(m.Icon, m.Icon).Shrink(0)
		}
		if afterEditor != nil {
			afterEditor(root)
		}
	})
	// Browsers indicate focus for text editors even after a pointer click.
	// MyGo's FocusVisible only tracks keyboard navigation, so use actual focus.
	focused := editor.Focused()
	config, styles := resolve(t, id, ComponentTheme{Color: Primary, Size: MD, Variant: Outline}, ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size), Variant: o.Variant, UI: o.UI}, func(v ComponentTheme) Styles {
		styles := inputRecipe(t, v, focused, root.Hovered() && !root.IsDisabled(), fieldInvalid(t, o))
		if o.Highlight && !fieldInvalid(t, o) && v.Variant != None {
			s := styles[Root]
			s.BorderWidth, s.BorderColor = Set(float32(1)), Set(t.color(v.Color))
			styles[Root] = s
		}
		return styles
	})
	styles[Root].apply(root)
	styles[Editor].apply(editor)
	styles[Leading].apply(leading)
	styles[Trailing].apply(trailing)
	if focused && config.Variant != None {
		color := config.Color
		if fieldInvalid(t, o) {
			color = Error
		}
		root.Shadow(0, 0, 0, 3, t.focusColor(color))
	}
	if clear.Valid() && clear.Clicked() {
		*value = ""
		editor.SetChanged().Focus()
	}
	if clear.Valid() && o.OnClear != nil {
		clear.OnClick(o.OnClear)
	}
	if root.Clicked() && !root.IsDisabled() {
		editor.Focus()
	}
	applyGroup(root, t, styles[Root].Radius.Value)
	if o.Style != nil {
		o.Style(root, editor, leading, trailing)
	}
	err := o.ErrorMessage
	errKey := o.errorKey
	if o.Invalid && err == "" {
		err = message(c, "checkValue")
		errKey = "checkValue"
	}
	trackFieldEdit(t, *value, err, errKey)
	return root
}

func trackFieldEdit(t Theme, text, message, messageKey string) {
	if f := t.field.State; f != nil {
		if !f.hasEditor {
			f.initialEditing = text
			f.editing = text
			f.hasEditor = true
		}
		if f.editing != text || f.controlError != message {
			if t.form != nil {
				t.form.version++
				t.form.Submitting = false
			}
		}
		f.editing = text
		f.editorDirty = f.editing != f.initialEditing
		f.Dirty = f.Value != f.Initial || f.editorDirty
		f.controlError = message
		f.controlErrorKey = messageKey
	}
}
