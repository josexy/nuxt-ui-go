package nui

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

type fieldDefaults struct {
	Size               Size
	Disabled, Required bool
	Error              string
	Focus              *ui.Handle
	State              *FormFieldState
	Label, Description string
	Activate           *func()
}
type groupPart struct{ active, first, last, vertical bool }

type FormFieldOptions struct {
	// Name and Value register a serialized value with the enclosing Form.
	Name, Value                           string
	Label, Description, Help, Hint, Error string
	Required, Disabled                    bool
	Size                                  Size
	Orientation                           Optional[Orientation] // Defaults to Vertical.
	Decoration
}

// FormField keeps the label and supporting text separate from the editor column.
// Only model data and the first control's handle cross build passes.
func FormField(c *ui.Context, t Theme, opts FormFieldOptions, content func(Theme)) ui.Element {
	var registered *FormFieldState
	if t.form != nil && opts.Name != "" {
		registered = t.form.Register(opts.Name, opts.Value)
		if registered.controlErrorKey != "" {
			// Translate before the editor compares its error: a locale change
			// updates visible validation without acting like a new form edit.
			translated := message(c, registered.controlErrorKey)
			if registered.Error == registered.controlError {
				registered.Error = translated
			}
			registered.controlError = translated
		}
		if opts.Error == "" && registered.Touched {
			opts.Error = registered.Error
		}
	}
	config := defaults(t, ComponentFormField, ComponentTheme{Size: MD})
	if opts.Size != SizeDefault {
		config.Size = opts.Size
	}
	inner := t
	inner.field = fieldDefaults{Size: config.Size, Disabled: opts.Disabled || t.field.Disabled, Required: opts.Required, Error: opts.Error}
	e := ui.Column(c).MinWidth(0).Disabled(inner.field.Disabled).Role(ui.RoleGroup)
	control := ui.Local(e, "control", func() ui.Handle { return ui.Handle{} })
	inner.field.Focus = control
	if registered != nil {
		inner.field.Focus = &registered.Handle
		inner.field.State = registered
		registered.Disabled = inner.field.Disabled
	}
	label := opts.Label
	if opts.Required && label != "" {
		label += " *"
	}
	hasHeader := label != "" || opts.Description != "" || opts.Hint != ""
	e.Label(label)
	inner.field.Label = label
	inner.field.Description = strings.TrimSpace(strings.Join([]string{opts.Description, opts.Hint, opts.Help}, "\n"))
	// Nuxt's text-xs/sm/base: an explicit line height keeps the horizontal
	// layout's baseline alignment independent of the platform font's metrics.
	size := sizeIndex(config.Size)
	fontSize := t.text([5]float32{12, 12, 14, 14, 16}[size])
	lineHeight := Set([5]float32{16, 16, 20, 20, 24}[size] / [5]float32{12, 12, 14, 14, 16}[size])
	copyStyle := PartStyle{Foreground: Set(t.Muted), FontSize: Set(fontSize), LineHeight: lineHeight}
	activate := func() { inner.field.Focus.Focus(c) }
	inner.field.Activate = &activate
	if !vertical(valueOr(opts.Orientation, Vertical)) {
		e.Row().Justify(ui.SpaceBetween).AlignItems(ui.Baseline)
	}
	e.Children(func() {
		if hasHeader {
			header := ui.Column(c.Key("label")).MinWidth(0)
			finish(t, ComponentFormField, Header, header, PartStyle{}, opts.Decoration)
			header.Children(func() {
				if label != "" || opts.Hint != "" {
					ui.Row(c.Key("heading")).Justify(ui.SpaceBetween).AlignItems(ui.Center).Gap(t.spacing(4)).Children(func() {
						if label != "" {
							// The accessible name keeps " *"; the visible asterisk is a
							// separate red mark 2 DIPs after the label, as Nuxt's.
							labelPart := ui.Row(c.Key("label-text")).MinWidth(0).AlignItems(ui.Center).FontWeight(500).OnClick(func() {
								if !inner.field.Disabled {
									activate()
								}
							}).Children(func() {
								if opts.Label != "" {
									ui.Text(c, opts.Label).MinWidth(0)
								}
								if opts.Required {
									ui.Text(c, "*").TextColor(t.Error).MarginStart(t.spacing(2)).Role(ui.RoleNone)
								}
							})
							finish(t, ComponentFormField, Label, labelPart, PartStyle{Foreground: Set(t.Text), FontSize: Set(fontSize), LineHeight: lineHeight}, opts.Decoration)
						}
						if opts.Hint != "" {
							finish(t, ComponentFormField, Hint, ui.Text(c, opts.Hint), copyStyle, opts.Decoration)
						}
					})
				}
				if opts.Description != "" {
					finish(t, ComponentFormField, Description, ui.Text(c, opts.Description), copyStyle, opts.Decoration)
				}
			})
		}
		body := ui.Column(c.Key("content")).MinWidth(0)
		finish(t, ComponentFormField, Body, body, PartStyle{Gap: Set(t.spacing(8))}, opts.Decoration)
		body.Children(func() {
			if content != nil {
				content(inner)
			}
			if opts.Error != "" {
				finish(t, ComponentFormField, ErrorSlot, ui.Text(c, opts.Error), PartStyle{Foreground: Set(t.Error), FontSize: Set(fontSize), LineHeight: lineHeight}, opts.Decoration)
			} else if opts.Help != "" {
				finish(t, ComponentFormField, Help, ui.Text(c, opts.Help), copyStyle, opts.Decoration)
			}
		})
	})
	if registered != nil {
		// A composite input can move focus between its segments without
		// leaving the field. The handle remains the first-error focus target.
		focused := e.FocusWithin()
		if registered.focused && !focused {
			registered.Touched = true
		}
		registered.focused = focused
	}
	c.Resolve(*inner.field.Focus).Error(opts.Error)
	// Content is 4 DIPs under the label; side by side, the columns are 8 apart.
	gap := float32(0)
	if hasHeader {
		gap = t.spacing(4)
		if !vertical(valueOr(opts.Orientation, Vertical)) {
			gap = t.spacing(8)
		}
	}
	finish(t, ComponentFormField, Root, e, PartStyle{Gap: Set(gap)}, opts.Decoration)
	return e
}

type FieldGroupItem struct {
	ID    string
	Build func(*ui.Context, Theme) ui.Element
}
type FieldGroupOptions struct {
	Size               Size
	Disabled, Vertical bool
	Decoration
}
type fieldGroupKey string

func FieldGroup(c *ui.Context, t Theme, opts FieldGroupOptions, items []FieldGroupItem) ui.Element {
	seen := map[string]bool{}
	for _, item := range items {
		if item.ID == "" || seen[item.ID] || item.Build == nil {
			panic("nui: FieldGroup needs unique non-empty IDs and builders")
		}
		seen[item.ID] = true
	}
	// Like Nuxt, an unsized group keeps the size of an enclosing FormField.
	size := t.field.Size
	if global := t.components[ComponentFieldGroup].Size; global != SizeDefault {
		size = global
	}
	if opts.Size != SizeDefault {
		size = opts.Size
	}
	// Items stretch like Nuxt's inline-flex group, so a Badge matches its neighbors' height.
	e := ui.Row(c).MinWidth(0).AlignItems(ui.Stretch).Disabled(opts.Disabled || t.field.Disabled)
	if opts.Vertical {
		e.Column()
	}
	finish(t, ComponentFieldGroup, Root, e, PartStyle{Gap: Set(float32(-1))}, opts.Decoration)
	e.Children(func() {
		for i, item := range items {
			inner := t
			inner.field.Size = size
			inner.field.Disabled = inner.field.Disabled || opts.Disabled
			inner.group = groupPart{true, i == 0, i == len(items)-1, opts.Vertical}
			item.Build(c.Key(fieldGroupKey(item.ID)), inner)
		}
	})
	return e
}
func applyGroup(c *ui.Context, e ui.Element, t Theme, radius float32) {
	g := t.group
	if !g.active {
		return
	}
	first, last := float32(0), float32(0)
	if g.first {
		first = radius
	}
	if g.last {
		last = radius
	}
	if g.vertical {
		e.Radius(first, first, last, last)
		return
	}
	// Radius corners are physical; right to left, the first item is at the right.
	if c.Direction() == ui.RightToLeft {
		first, last = last, first
	}
	e.Radius(first, last, last, first)
}
func inheritedSize(t Theme, size Size) Size {
	if size == SizeDefault {
		return t.field.Size
	}
	return size
}

type ScrollDirection uint8

const (
	ScrollVertical ScrollDirection = iota
	ScrollHorizontal
	ScrollBoth
)

type ScrollAreaOptions struct {
	Direction      ScrollDirection
	ScrollbarColor Optional[ui.Color]
	ScrollbarWidth Optional[float32]
	State          *ui.ScrollState
	Decoration
}

// ScrollArea uses a themed wrapper because MyGo reads scrollbar colors from its theme.
// Size the returned wrapper; it is intentionally not a transparent provider.
func ScrollArea(c *ui.Context, t Theme, opts ScrollAreaOptions, content func()) ui.Element {
	native := t.Native()
	native.Scrollbar = t.BorderAccented
	if opts.ScrollbarColor.Set {
		native.Scrollbar = opts.ScrollbarColor.Value
	}
	if opts.ScrollbarWidth.Set {
		native.ScrollbarWidth = opts.ScrollbarWidth.Value
	}
	root := ui.Themed(c, native).MinWidth(0).MinHeight(0).Clip()
	finish(t, ComponentScrollArea, Root, root, PartStyle{}, opts.Decoration)
	root.Children(func() {
		var scroll ui.Element
		switch opts.Direction {
		case ScrollHorizontal:
			scroll = ui.ScrollHorizontal(c)
		case ScrollBoth:
			scroll = ui.ScrollBoth(c)
		default:
			scroll = ui.Scroll(c)
		}
		scroll.Fill().MinWidth(0).MinHeight(0)
		if opts.State != nil {
			scroll.TrackScroll(opts.State)
		}
		finish(t, ComponentScrollArea, Body, scroll, PartStyle{}, opts.Decoration)
		if content != nil {
			scroll.Children(content)
		}
	})
	return root
}

// Bind only the first focusable field control (groups can have several).
func bindField(c *ui.Context, e ui.Element, t Theme) {
	if t.field.Focus != nil && !c.Resolve(*t.field.Focus).Valid() {
		e.Bind(t.field.Focus)
		e.Label(t.field.Label).Description(t.field.Description)
	}
}

func fieldActivation(c *ui.Context, e ui.Element, t Theme, activate func() bool) {
	if t.field.Activate != nil && t.field.Focus != nil && c.Resolve(*t.field.Focus) == e {
		pending := ui.Local(e, "field-change", func() bool { return false })
		if *pending {
			*pending = false
			if activate() {
				e.SetChanged()
			}
		}
		*t.field.Activate = func() {
			t.field.Focus.Focus(c)
			// Recompute derived values (such as a group item's checked bool)
			// before activating, then let the current build observe the change.
			*pending = true
		}
	}
}
