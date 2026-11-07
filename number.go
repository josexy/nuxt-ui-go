package nui

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// NumberState keeps incomplete text independent of the optional numeric value.
type NumberState struct {
	Text, Error          string
	seen                 Optional[float64]
	initialized          bool
	Handle               ui.Handle
	increment, decrement ui.Handle
}
type InputNumberOptions struct {
	InputOptions
	// Horizontal (the default) places minus/plus around centered text.
	// Vertical stacks up/down chevrons at the trailing edge.
	Orientation Orientation
	Min, Max    Optional[float64]
	Step        float64
	// StepSnapping and FocusOnChange default to true.
	StepSnapping, FocusOnChange           Optional[bool]
	StartingValue                         Optional[float64]
	AllowInvalid                          bool
	DisableWheelChange, InvertWheelChange bool
	Increment, Decrement                  Optional[bool]
	IncrementOptions, DecrementOptions    *ButtonOptions
	IncrementIcon, DecrementIcon          *ui.SVG
	IncrementDisabled, DecrementDisabled  bool
	OnValueChange                         func(Optional[float64])
}

type numberPress struct {
	pressed, handled bool
	next             time.Time
}

func numberBounds(o InputNumberOptions) (float64, float64, float64) {
	lo, hi, step := math.Inf(-1), math.Inf(1), o.Step
	if o.Min.Set {
		lo = o.Min.Value
	}
	if o.Max.Set {
		hi = o.Max.Value
	}
	if step == 0 {
		step = 1
	}
	if math.IsNaN(lo) || math.IsNaN(hi) || o.Min.Set && math.IsInf(lo, 0) || o.Max.Set && math.IsInf(hi, 0) || lo > hi || math.IsNaN(step) || math.IsInf(step, 0) || step <= 0 {
		panic("nui: invalid number bounds or step")
	}
	if o.StartingValue.Set && (math.IsNaN(o.StartingValue.Value) || math.IsInf(o.StartingValue.Value, 0)) {
		panic("nui: number starting value must be finite")
	}
	return lo, hi, step
}
func numberText(v Optional[float64]) string {
	if !v.Set {
		return ""
	}
	return strconv.FormatFloat(v.Value, 'f', -1, 64)
}

// ParseNumber accepts empty input and finite numbers; AllowInvalid bypasses bounds.
func ParseNumber(text string, o InputNumberOptions) (Optional[float64], bool) {
	lo, hi, _ := numberBounds(o)
	v, ok := parseNumberText(text)
	if !ok || v.Set && !o.AllowInvalid && (v.Value < lo || v.Value > hi) {
		return Optional[float64]{}, false
	}
	return v, true
}

func parseNumberText(text string) (Optional[float64], bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Optional[float64]{}, true
	}
	for _, r := range text {
		if (r < '0' || r > '9') && !strings.ContainsRune(".+-eE", r) {
			return Optional[float64]{}, false
		}
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return Optional[float64]{}, false
	}
	return Set(v), true
}

func roundNumber(value, base, step float64) float64 {
	precision := sliderPrecision(base, step)
	if rounded, err := strconv.ParseFloat(strconv.FormatFloat(value, 'f', precision, 64), 64); err == nil {
		return rounded
	}
	return value
}

func normalizeNumber(value float64, o InputNumberOptions) float64 {
	lo, hi, step := numberBounds(o)
	base := valueOr(o.Min, 0)
	if valueOr(o.StepSnapping, true) {
		index := (value - base) / step
		if !math.IsInf(index, 0) && !math.IsNaN(index) {
			value = roundNumber(base+math.Round(index)*step, base, step)
		}
		if value > hi {
			index = (hi - base) / step
			if !math.IsInf(index, 0) {
				index = math.Floor(index)
				value = roundNumber(base+index*step, base, step)
				// Decimal bounds such as (0.3 - 0.1) / 0.1 can land
				// just below an integer in binary floating point.
				if next := roundNumber(base+(index+1)*step, base, step); next <= hi {
					value = next
				}
			}
		}
	}
	return max(lo, min(hi, value))
}

func numberStep(value Optional[float64], delta float64, o InputNumberOptions) Optional[float64] {
	lo, hi, step := numberBounds(o)
	base := valueOr(o.Min, 0)
	if !value.Set {
		if o.StartingValue.Set {
			return Set(max(lo, min(hi, o.StartingValue.Value)))
		}
		return Set(normalizeNumber(base, o))
	}
	from := value.Value
	next := from + delta*step
	precisionBase := from
	if valueOr(o.StepSnapping, true) {
		precisionBase = base
		index := (from - base) / step
		if !math.IsInf(index, 0) && !math.IsNaN(index) {
			nearest := roundNumber(base+math.Round(index)*step, base, step)
			if nearest != from {
				if delta > 0 {
					next = base + (math.Ceil(index)+delta-1)*step
				} else {
					next = base + (math.Floor(index)+delta+1)*step
				}
			}
		}
	}
	next = roundNumber(next, precisionBase, step)
	next = normalizeNumber(next, o)
	if math.IsNaN(next) || math.IsInf(next, 0) || delta > 0 && next < from || delta < 0 && next > from {
		return value
	}
	return Set(next)
}
func InputNumber(c *ui.Context, t Theme, value *Optional[float64], s *NumberState, o InputNumberOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: InputNumber requires value and state")
	}
	if o.Orientation > Vertical {
		panic("nui: invalid InputNumber orientation")
	}
	config := inputDefaults(t, ComponentInputNumber, o.InputOptions)
	m := buttonMetrics(t, config.Size)
	buttonWidth, inset := m.Icon+2*m.Y, t.spacing(4)
	vertical := o.Orientation == Vertical
	lo, hi, _ := numberBounds(o)
	showIncrement, showDecrement := valueOr(o.Increment, true), valueOr(o.Decrement, true)
	if value.Set && (math.IsNaN(value.Value) || math.IsInf(value.Value, 0)) {
		panic("nui: number value must be finite")
	}
	if !s.initialized || s.seen != *value {
		s.Text = numberText(*value)
		s.seen = *value
		s.initialized = true
	}
	changed := func(v Optional[float64]) {
		if v != *value {
			*value = v
			s.seen = v
			if o.OnValueChange != nil {
				o.OnValueChange(v)
			}
		}
	}
	s.Error = ""
	if _, ok := ParseNumber(s.Text, o); !ok {
		s.Error = message(c, "invalidNumber")
	}
	in := o.InputOptions
	if s.Error != "" {
		in.ErrorMessage = s.Error
		in.errorKey = "invalidNumber"
	}
	priorChange, priorSubmit := in.OnChange, in.OnSubmit
	in.OnChange = func(text string) {
		if v, ok := ParseNumber(text, o); ok {
			changed(v)
		}
		if priorChange != nil {
			priorChange(text)
		}
	}
	in.OnSubmit = func(text string) {
		if v, ok := ParseNumber(text, o); ok {
			if v.Set && !o.AllowInvalid {
				v.Value = normalizeNumber(v.Value, o)
			}
			changed(v)
		}
		s.Text = numberText(*value)
		if priorSubmit != nil {
			priorSubmit(s.Text)
		}
	}
	adjust := func(delta float64) {
		if o.ReadOnly || o.Disabled || t.field.Disabled {
			return
		}
		if valueOr(o.FocusOnChange, true) {
			s.Handle.Focus(c)
		}
		v, _ := parseNumberText(s.Text)
		next := numberStep(v, delta, o)
		if v.Set && next == v && v != *value {
			return // A bound must not move an invalid draft in the opposite direction.
		}
		changed(next)
		s.Text = numberText(*value)
	}
	style := in.Style
	in.Style = func(root, editor, leading, trailing ui.Element) {
		editor.Bind(&s.Handle).InputMethod(false).HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind != ui.InputScroll || o.DisableWheelChange || o.ReadOnly || !s.Handle.Focused(c) || math.Abs(float64(ev.DY)) <= math.Abs(float64(ev.DX)) {
				return false
			}
			delta := float64(1)
			if (ev.DY < 0) != o.InvertWheelChange {
				delta = -1
			}
			adjust(delta)
			return true
		})
		if showIncrement || vertical && showDecrement {
			root.PaddingEnd(buttonWidth + inset)
		}
		if !vertical && showDecrement {
			root.PaddingStart(buttonWidth + inset)
			editor.TextAlign(ui.Center)
		} else {
			editor.TextAlign(ui.Start)
		}
		if style != nil {
			style(root, editor, leading, trailing)
		}
	}
	return inputShell(c, t, &s.Text, in, ComponentInputNumber, 0, 0, func() ui.Element {
		return ui.TextInputBase(c.Key("number"), &s.Text).OnTextKey(func(k ui.TextKey) bool {
			if k.Mods != 0 || o.ReadOnly {
				return false
			}
			step := func(delta float64) { s.Text = k.Text(); adjust(delta) }
			switch k.Key {
			case ui.KeyUp:
				step(1)
			case ui.KeyDown:
				step(-1)
			case ui.KeyPageUp:
				step(10)
			case ui.KeyPageDown:
				step(-10)
			case ui.KeyHome:
				if !o.Min.Set {
					return false
				}
				changed(Set(normalizeNumber(lo, o)))
				s.Text = numberText(*value)
			case ui.KeyEnd:
				if !o.Max.Set {
					return false
				}
				changed(Set(normalizeNumber(hi, o)))
				s.Text = numberText(*value)
			default:
				return false
			}
			k.SetText(s.Text)
			return true
		})
	}, func(root ui.Element) {
		// Absolute siblings keep the editor and both buttons at stable keys and
		// parents when orientation changes. Their space is reserved above.
		for _, increase := range []bool{false, true} {
			key, label, icon, delta := "decrease", "decreaseNamed", lucide.Minus(), -1.0
			shown, disabled, custom, explicit, handle, slot := showDecrement, o.DecrementDisabled, o.DecrementOptions, o.DecrementIcon, &s.decrement, DecrementSlot
			if increase {
				key, label, icon, delta = "increase", "increaseNamed", lucide.Plus(), 1
				shown, disabled, custom, explicit, handle, slot = showIncrement, o.IncrementDisabled, o.IncrementOptions, o.IncrementIcon, &s.increment, IncrementSlot
			}
			if vertical {
				icon = lucide.ChevronDown()
				if increase {
					icon = lucide.ChevronUp()
				}
			}
			if explicit != nil {
				icon = explicit
			}
			opts := ButtonOptions{}
			if custom != nil {
				opts = *custom
			}
			if opts.AccessibleLabel == "" {
				opts.AccessibleLabel = namedMessage(c, label, o.Label)
			}
			if opts.Leading == nil {
				opts.Leading = icon
			}
			if opts.Color == ColorDefault {
				opts.Color = config.Color
			}
			if opts.Size == SizeDefault {
				opts.Size = config.Size
			}
			if opts.Variant == VariantDefault {
				opts.Variant = LinkVariant
			}
			atBound := false
			if value.Set {
				next := numberStep(*value, delta, o)
				atBound = increase && next.Value <= value.Value || !increase && next.Value >= value.Value
			}
			opts.Disabled = opts.Disabled || o.ReadOnly || disabled || atBound
			if !shown || opts.Disabled || opts.Loading {
				if !root.IsDisabled() && handle.Focused(c) {
					s.Handle.Focus(c)
				}
				handle.CancelFocus(c)
			}
			if !shown {
				continue
			}
			buttonStyle := opts.Style
			opts.Style = func(button, label, leading, trailing ui.Element) {
				button.Bind(handle)
				button.Absolute().Width(buttonWidth).Padding(0).Center()
				// Nuxt uses logical start/end: mirror placement in RTL.
				atEnd := vertical || increase
				if atEnd != (c.Direction() == ui.RightToLeft) {
					button.Right(inset)
				} else {
					button.Left(inset)
				}
				if vertical {
					button.HeightPercent(50)
					if increase {
						button.Top(0)
					} else {
						button.Bottom(0)
					}
					size := min(m.Icon*.8, (m.Line+2*m.Y)/2)
					leading.Size(size, size)
				} else {
					button.Top(0).Bottom(0)
				}
				finish(t, ComponentInputNumber, slot, button, PartStyle{}, Decoration{UI: o.UI})
				if buttonStyle != nil {
					buttonStyle(button, label, leading, trailing)
				}
			}
			button := Button(c.Key(key), t, "", opts)
			if valueOr(o.FocusOnChange, true) {
				button.KeepFocus()
			}
			press := ui.Local(button, "repeat", func() numberPress { return numberPress{} })
			if button.Pressed() && !button.IsDisabled() {
				if !press.pressed {
					press.pressed, press.handled = true, true
					press.next = c.Now().Add(400 * time.Millisecond)
					adjust(delta)
					c.Services().Invalidate()
				} else if !c.Now().Before(press.next) {
					press.next = c.Now().Add(60 * time.Millisecond)
					adjust(delta)
					c.Services().Invalidate()
				}
				c.After(press.next.Sub(c.Now()))
			} else {
				press.pressed = false
				if !button.Clicked() {
					press.handled = false
				}
			}
			button.OnClick(func() {
				if press.handled {
					press.handled = false
					return
				}
				adjust(delta)
			})
		}
	})
}
