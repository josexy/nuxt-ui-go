package nui

import (
	"fmt"
	"math"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type ToggleOptions struct {
	mixed                                   bool
	Label, Description, ErrorMessage        string
	Disabled, Invalid                       bool
	Color                                   Color
	Size                                    Size
	Loading                                 bool
	LoadingIcon, CheckedIcon, UncheckedIcon *ui.SVG
	Variant                                 ChoiceVariant
	Indicator                               IndicatorPlacement
	Highlight                               bool
	OnChange                                func(bool)
	Decoration
}
type CheckboxOptions = ToggleOptions
type SwitchOptions = ToggleOptions

type ChoiceVariant uint8

const (
	ChoiceList ChoiceVariant = iota
	ChoiceCard
	ChoiceTable
)

type IndicatorPlacement uint8

const (
	IndicatorStart IndicatorPlacement = iota
	IndicatorEnd
	IndicatorHidden
)

func Checkbox(c *ui.Context, t Theme, value *bool, o CheckboxOptions) ui.Element {
	if value == nil {
		panic("nui: Checkbox requires a bool pointer")
	}
	e := ui.CheckboxBase(c, value).Disabled(o.Disabled || t.field.Disabled).FocusRing(false)
	togglePresentation(c, t, e, *value, ComponentCheckbox, false, false, o)
	fieldActivation(c, e, t, func() bool {
		if !o.Disabled && !t.field.Disabled {
			*value = !*value
			return true
		}
		return false
	})
	if o.OnChange != nil {
		e.OnChange(func() { o.OnChange(*value) })
	}
	return e
}
func Switch(c *ui.Context, t Theme, value *bool, o SwitchOptions) ui.Element {
	if value == nil {
		panic("nui: Switch requires a bool pointer")
	}
	e := ui.SwitchBase(c, value).Disabled(o.Disabled || o.Loading || t.field.Disabled).FocusRing(false)
	togglePresentation(c, t, e, *value, ComponentSwitch, true, false, o)
	fieldActivation(c, e, t, func() bool {
		if !o.Disabled && !o.Loading && !t.field.Disabled {
			*value = !*value
			return true
		}
		return false
	})
	if o.OnChange != nil {
		e.OnChange(func() { o.OnChange(*value) })
	}
	return e
}
func togglePresentation(c *ui.Context, t Theme, e ui.Element, on bool, id Component, isSwitch, radio bool, o ToggleOptions) {
	bindField(c, e, t)
	rowSlot := Root
	if radio {
		rowSlot = Item
	}
	config, styles := resolve(t, id, ComponentTheme{Color: Primary, Size: MD}, ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size), UI: o.UI}, func(v ComponentTheme) Styles {
		bg, border := t.Background, t.BorderAccented
		if on {
			bg = t.color(v.Color)
			border = bg
		}
		if o.Highlight {
			border = t.color(v.Color)
		}
		invalid := o.Invalid || o.ErrorMessage != "" || t.field.Error != ""
		if invalid {
			border = t.Error
		}
		indicator := PartStyle{Background: Set(bg), BorderWidth: Set(float32(1)), BorderColor: Set(border), Radius: Set(t.radius(4))}
		if isSwitch && !o.Highlight && !invalid {
			// Nuxt's switch track border is transparent; only highlight and errors draw a ring.
			indicator.BorderWidth = Set(float32(0))
		}
		return Styles{rowSlot: {Gap: Set(t.spacing(8)), Font: Set(t.Font), FontSize: Set(t.text([5]float32{12, 12, 14, 14, 16}[sizeIndex(v.Size)])), LineHeight: Set(float32(20.0 / 14))}, Indicator: indicator, Label: {Foreground: Set(t.Text), Weight: Set(500)}, Description: {Foreground: Set(t.Muted), FontSize: Set(t.text([5]float32{12, 12, 14, 14, 16}[sizeIndex(v.Size)]))}, Thumb: {Background: Set(t.Background)}}
	})
	e.MinWidth(0).AlignItems(ui.Start)
	styles[rowSlot].apply(e)
	if o.Variant > ChoiceTable || o.Indicator > IndicatorHidden {
		panic("nui: invalid choice appearance")
	}
	if o.Variant != ChoiceList {
		e.Padding(t.spacing([5]float32{10, 12, 14, 16, 18}[sizeIndex(config.Size)])).BorderWidth(1).BorderColor(t.Border).Radius(t.radius(8))
		if o.Variant == ChoiceTable {
			e.Radius(0)
		}
		if on {
			e.BorderColor(t.color(config.Color).Alpha(.5)).Background(t.color(config.Color).Alpha(.10))
			if config.Color == Neutral {
				e.Background(t.Elevated)
			}
		} else if e.Hovered() && !e.IsDisabled() {
			e.Background(t.Elevated.Alpha(.5))
		}
		if o.Highlight {
			e.BorderColor(t.color(config.Color))
		}
	}
	if (o.Variant != ChoiceList || o.Indicator == IndicatorHidden) && e.FocusVisible() {
		e.Shadow(0, 0, 0, 3, t.focusColor(config.Color))
	}
	if o.Label != "" {
		e.Label(o.Label)
	}
	e.Description(o.Description)
	if t.field.Required {
		e.Description(message(c, "required"))
	}
	if o.ErrorMessage != "" {
		e.Error(o.ErrorMessage)
	} else if o.Invalid && t.field.Error == "" {
		e.Error(message(c, "checkValue"))
	}
	boxSize := t.text([5]float32{12, 14, 16, 18, 20}[sizeIndex(config.Size)])
	line := t.text([5]float32{16, 16, 20, 20, 24}[sizeIndex(config.Size)])
	e.Children(func() {
		indicator := func() {
			ui.Box(c.Key("indicator")).Height(max(line, boxSize+4)).Shrink(0).Center().Children(func() {
				box := ui.Box(c).Size(boxSize, boxSize).Center().Shrink(0)
				styles[Indicator].apply(box)
				if radio {
					box.Radius(999)
				}
				if isSwitch {
					width := t.text([5]float32{28, 32, 36, 40, 44}[sizeIndex(config.Size)])
					height := boxSize + 4
					bg := t.Accented
					if on {
						bg = t.color(config.Color)
					}
					bg = mergeValue(Set(bg), t.components[id].UI[Indicator].Background).Value
					if o.Color != ColorDefault {
						if on {
							bg = t.color(o.Color)
						} else {
							bg = t.Accented
						}
					}
					bg = mergeValue(Set(bg), o.UI[Indicator].Background).Value
					box.Size(width, height).Radius(999).Background(bg)
					offset := float32(2)
					if on {
						offset = width - boxSize - 2
					}
					if c.Direction() == ui.RightToLeft {
						offset = width - boxSize - offset
					}
					box.Children(func() {
						thumb := ui.Box(c).Size(boxSize, boxSize).Radius(999).Absolute().Left(offset).Top(2).Center()
						decorated(Thumb, thumb, styles, o.Style)
						thumb.Children(func() {
							iconColor := t.Dimmed
							if on {
								iconColor = t.color(config.Color)
							}
							icon := o.UncheckedIcon
							if on {
								icon = o.CheckedIcon
							}
							if o.Loading {
								icon = o.LoadingIcon
								if icon == nil {
									icon = lucide.LoaderCircle()
								}
								loadingSVG(c, icon, boxSize*10/12, iconColor, t.motion(c), t.Motion.Spin)
							} else if icon != nil {
								ui.Icon(c, icon).Size(boxSize*10/12, boxSize*10/12).TextColor(iconColor)
							}
						})
					})
				} else if on {
					box.Children(func() {
						if radio {
							ui.Box(c).Size(boxSize*.4, boxSize*.4).Radius(999).Background(t.Background)
						} else {
							icon := lucide.Check()
							if o.CheckedIcon != nil {
								icon = o.CheckedIcon
							}
							if o.mixed {
								icon = lucide.Minus()
							}
							ui.Icon(c, icon).Size(boxSize-2, boxSize-2).TextColor(t.Inverted)
						}
					})
				}
				if e.FocusVisible() {
					color := config.Color
					if o.Invalid || o.ErrorMessage != "" || t.field.Error != "" {
						color = Error
					}
					box.Shadow(0, 0, 0, 3, t.focusColor(color))
				}
				if o.Style != nil {
					o.Style(Indicator, box)
				}
			})
		}
		label := func() {
			if o.Label != "" || o.Description != "" {
				ui.Column(c.Key("label")).MinWidth(0).Grow(1).Gap(t.spacing(2)).Children(func() {
					if o.Label != "" {
						decorated(Label, ui.Text(c, o.Label), styles, o.Style)
					}
					if o.Description != "" {
						decorated(Description, ui.Text(c, o.Description), styles, o.Style)
					}
				})
			}
		}
		if o.Indicator == IndicatorEnd {
			label()
			indicator()
		} else {
			if o.Indicator != IndicatorHidden {
				indicator()
			}
			label()
		}
	})
	if o.Style != nil {
		o.Style(rowSlot, e)
	}
}

type ChoiceGroupOptions struct {
	Label, Description, ErrorMessage string
	Disabled, Invalid, Horizontal    bool
	Color                            Color
	Size                             Size
	Max                              Optional[int]
	Variant                          ChoiceVariant
	Indicator                        IndicatorPlacement
	Highlight                        bool
	OnChange                         func()
	Decoration
}
type CheckboxGroupOptions = ChoiceGroupOptions
type RadioGroupOptions = ChoiceGroupOptions

func CheckboxGroup[T comparable](c *ui.Context, t Theme, selected *[]T, options []Option[T], o CheckboxGroupOptions) ui.Element {
	if selected == nil {
		panic("nui: CheckboxGroup requires selection")
	}
	validateOptions(options)
	config := selectedDefaults(t, ComponentCheckboxGroup, o)
	e := ui.Column(c).MinWidth(0).Role(ui.RoleGroup).Label(o.Label).Description(o.Description).Disabled(o.Disabled || t.field.Disabled)
	if o.Horizontal {
		e.Row().Wrap()
	}
	finish(t, ComponentCheckboxGroup, Root, e, PartStyle{Gap: Set(t.spacing(10))}, o.Decoration)
	// The group styles its own Root and each row's Item; the remaining parts
	// (Label, Description, Indicator, ...) style every Checkbox, as in RadioGroup.
	child := Decoration{UI: cloneStyles(t.components[ComponentCheckboxGroup].UI)}
	for slot, style := range o.UI {
		child.UI[slot] = mergeStyle(child.UI[slot], style)
	}
	delete(child.UI, Root)
	delete(child.UI, Item)
	if o.Style != nil {
		child.Style = func(slot Slot, part ui.Element) {
			if slot != Root {
				o.Style(slot, part)
			}
		}
	}
	e.Children(func() {
		for _, option := range options {
			checked := slices.Contains(*selected, option.Value)
			limited := o.Max.Set && !checked && len(*selected) >= max(0, o.Max.Value)
			item := Checkbox(c.Key(optionKey[T]{option.Value}), t, &checked, CheckboxOptions{Label: option.Label, Description: option.Description, Disabled: o.Disabled || option.Disabled || limited, Invalid: o.Invalid || o.ErrorMessage != "", Color: config.Color, Size: config.Size, Variant: o.Variant, Indicator: o.Indicator, Highlight: o.Highlight, Decoration: child})
			finish(t, ComponentCheckboxGroup, Item, item, PartStyle{}, o.Decoration)
			if item.Changed() {
				next := slices.Clone(*selected)
				next = slices.DeleteFunc(next, func(v T) bool { return v == option.Value })
				if checked {
					next = append(next, option.Value)
				}
				*selected = next
				e.SetChanged()
			}
		}
	})
	if o.Variant == ChoiceTable {
		e.Gap(0)
	}
	if o.ErrorMessage != "" {
		e.Error(o.ErrorMessage)
	}
	if o.OnChange != nil {
		e.OnChange(o.OnChange)
	}
	return e
}
func selectedDefaults(t Theme, id Component, o ChoiceGroupOptions) ComponentTheme {
	return selected(defaults(t, id, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size)})
}
func RadioGroup[T comparable](c *ui.Context, t Theme, value *T, options []Option[T], o RadioGroupOptions) ui.Element {
	if value == nil {
		panic("nui: RadioGroup requires selection")
	}
	validateOptions(options)
	changed := false
	e := ui.RadioGroup(c, func() {
		for _, option := range options {
			item := ui.RadioBase(c.Key(optionKey[T]{option.Value}), value, option.Value).Disabled(o.Disabled || option.Disabled || t.field.Disabled).FocusRing(false)
			togglePresentation(c, t, item, *value == option.Value, ComponentRadioGroup, false, true, ToggleOptions{Label: option.Label, Description: option.Description, Disabled: o.Disabled || option.Disabled, Invalid: o.Invalid || o.ErrorMessage != "", Color: o.Color, Size: o.Size, Variant: o.Variant, Indicator: o.Indicator, Highlight: o.Highlight, Decoration: o.Decoration})
			changed = changed || item.Changed()
		}
	}).MinWidth(0).Label(o.Label).Description(o.Description).Disabled(o.Disabled || t.field.Disabled)
	if o.Horizontal {
		e.Row().Wrap()
	}
	finish(t, ComponentRadioGroup, Root, e, PartStyle{Gap: Set(t.spacing(10))}, o.Decoration)
	if o.Variant == ChoiceTable {
		e.Gap(0)
	}
	if changed {
		e.SetChanged()
	}
	if o.ErrorMessage != "" {
		e.Error(o.ErrorMessage)
	}
	if o.OnChange != nil {
		e.OnChange(o.OnChange)
	}
	return e
}

type SliderOptions struct {
	Label, Description string
	Min, Max           Optional[float64]
	Step               float64
	Color              Color
	Size               Size
	Disabled, Vertical bool
	Inverted, Tooltip  bool
	OnChange           func(float64)
	OnCommit           func(float64) // Release or a keyboard step; OnChange continues to report live changes.
	Decoration
}

func Slider(c *ui.Context, t Theme, value *float64, o SliderOptions) ui.Element {
	if value == nil {
		panic("nui: Slider requires a value pointer")
	}
	lo, hi := float64(0), float64(100)
	if o.Min.Set {
		lo = o.Min.Value
	}
	if o.Max.Set {
		hi = o.Max.Value
	}
	if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) || hi <= lo || math.IsNaN(o.Step) || math.IsInf(o.Step, 0) || o.Step < 0 {
		panic("nui: invalid Slider range or step")
	}
	*value = math.Max(lo, math.Min(hi, finite(*value)))
	e := ui.Box(c).Focusable().Role(ui.RoleSlider).Range(lo, hi, *value)
	s := ui.Local(e, "interaction", func() sliderInputMemory { return sliderInputMemory{} })
	if o.Disabled || t.field.Disabled {
		s.dragging = false
	}
	e.Bind(&s.handle)
	if s.changed {
		e.SetChanged()
		s.changed = false
	}
	step := o.Step
	if step == 0 {
		step = 1
	}
	set := func(v float64) bool {
		v = max(lo, min(hi, lo+math.Round((v-lo)/step)*step))
		if v == *value {
			return false
		}
		*value = v
		s.changed = true
		if o.OnChange != nil {
			o.OnChange(v)
		}
		return true
	}
	for _, key := range sliderKeys {
		for _, mods := range []ui.Modifiers{0, ui.Shift} {
			e.OnShortcut(mods, key, func() {
				if set(sliderKeyValue(key, mods == ui.Shift, *value, lo, hi, step, o.Vertical, o.Inverted, c.Direction() == ui.RightToLeft)) && o.OnCommit != nil {
					o.OnCommit(*value)
				}
			})
		}
	}
	e.Label(o.Label).Description(o.Description).Disabled(o.Disabled || t.field.Disabled).FocusRing(false).Step(o.Step)
	bindField(c, e, t)
	config := selected(defaults(t, ComponentSlider, ComponentTheme{Color: Primary, Size: MD}), ComponentTheme{Color: o.Color, Size: inheritedSize(t, o.Size)})
	thumb := t.text([5]float32{12, 14, 16, 18, 20}[sizeIndex(config.Size)])
	thickness := t.spacing([5]float32{2, 3, 4, 5, 6}[sizeIndex(config.Size)])
	rootStyle := PartStyle{PaddingX: Set(thumb / 2)}
	if o.Vertical {
		e.Vertical().Width(max(24, thumb)).Height(160)
		rootStyle = PartStyle{PaddingY: Set(thumb / 2)}
	} else {
		e.FillWidth().MinWidth(64).Height(max(24, thumb))
	}
	rootStyle = mergeStyle(rootStyle, t.components[ComponentSlider].UI[Root])
	rootStyle = mergeStyle(rootStyle, o.UI[Root])
	rootStyle.apply(e)
	{
		e.Step(step)
		input := func(ev ui.InputEvent) bool {
			if o.Disabled || t.field.Disabled {
				return false
			}
			if ev.Kind == ui.InputPointerUp {
				was := s.dragging
				s.dragging = false
				if was && *value != s.start && o.OnCommit != nil {
					o.OnCommit(*value)
				}
				return was
			}
			if ev.Button != 0 || (ev.Kind != ui.InputPointerDown && ev.Kind != ui.InputPointerMove) || (ev.Kind == ui.InputPointerMove && !s.dragging) {
				return false
			}
			if ev.Kind == ui.InputPointerDown {
				s.start = *value
				s.dragging = true
				s.handle.Focus(c)
			}
			r := s.handle.Bounds(c)
			axis, length, pad := ev.X, r.W, rootStyle.PaddingX.Value
			if o.Vertical {
				axis, length, pad = ev.Y, r.H, rootStyle.PaddingY.Value
			}
			f := max(float32(0), min(float32(1), (axis-pad)/max(float32(1), length-2*pad)))
			if (o.Vertical || c.Direction() == ui.RightToLeft) != o.Inverted {
				f = 1 - f
			}
			set(lo + float64(f)*(hi-lo))
			return true
		}
		e.HandleInput(input)
	}
	color := t.color(config.Color)
	if global := t.components[ComponentSlider].UI[Indicator].Background; global.Set && o.Color == ColorDefault {
		color = global.Value
	}
	color = mergeValue(Set(color), o.UI[Indicator].Background).Value
	thumbColor := mergeValue(mergeValue(Set(t.Background), t.components[ComponentSlider].UI[Thumb].Background), o.UI[Thumb].Background).Value
	ratio := float32((*value - lo) / (hi - lo))
	rtl := c.Direction() == ui.RightToLeft
	focus := e.FocusVisible()
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		track := r
		position := ratio
		if o.Vertical {
			padding := rootStyle.PaddingY.Value
			track.Y += padding
			track.H -= 2 * padding
			track.X += r.W/2 - thickness/2
			track.W = thickness
			position = 1 - ratio
		} else {
			padding := rootStyle.PaddingX.Value
			track.X += padding
			track.W -= 2 * padding
			track.Y += r.H/2 - thickness/2
			track.H = thickness
			if rtl {
				position = 1 - ratio
			}
		}
		if o.Inverted {
			position = 1 - position
		}
		p.Fill(track, t.Accented, 999)
		filled := track
		dot := ui.Rect{W: thumb, H: thumb}
		if o.Vertical {
			filled.H = track.H * ratio
			filled.Y = track.Y + track.H - filled.H
			if o.Inverted {
				filled.Y = track.Y
			}
			dot.X = r.X + (r.W-thumb)/2
			dot.Y = track.Y + track.H*position - thumb/2
		} else {
			filled.W = track.W * ratio
			if rtl != o.Inverted {
				filled.X = track.X + track.W - filled.W
			}
			dot.X = track.X + track.W*position - thumb/2
			dot.Y = r.Y + (r.H-thumb)/2
		}
		p.Fill(filled, color, 999)
		if focus {
			p.Fill(ui.Rect{X: dot.X - 3, Y: dot.Y - 3, W: dot.W + 6, H: dot.H + 6}, t.focusColor(config.Color), 999)
		}
		p.Fill(dot, color, 999)
		p.Fill(ui.Rect{X: dot.X + 2, Y: dot.Y + 2, W: dot.W - 4, H: dot.H - 4}, thumbColor, 999)
	})
	if o.Tooltip {
		position := ratio
		if (o.Vertical || rtl) != o.Inverted {
			position = 1 - position
		}
		e.Children(func() { sliderValueTooltip(c, t, e.Focused(), fmt.Sprint(*value), position, o.Vertical, thumb) })
	}
	if o.Style != nil {
		o.Style(Root, e)
	}
	return e
}
