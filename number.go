package nui

import (
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

// NumberState keeps incomplete text independent of the optional numeric value.
type NumberState struct {
	Text, Error string
	seen        Optional[float64]
	initialized bool
	Handle      ui.Handle
}
type InputNumberOptions struct {
	InputOptions
	// Horizontal (the default) places minus/plus around centered text.
	// Vertical stacks up/down chevrons at the trailing edge.
	Orientation   Orientation
	Min, Max      Optional[float64]
	Step          float64
	OnValueChange func(Optional[float64])
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
	return lo, hi, step
}
func numberText(v Optional[float64]) string {
	if !v.Set {
		return ""
	}
	return strconv.FormatFloat(v.Value, 'f', -1, 64)
}

// ParseNumber accepts empty input, rejects non-finite and out-of-range values.
func ParseNumber(text string, o InputNumberOptions) (Optional[float64], bool) {
	lo, hi, _ := numberBounds(o)
	text = strings.TrimSpace(text)
	if text == "" {
		return Optional[float64]{}, true
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < lo || v > hi {
		return Optional[float64]{}, false
	}
	return Set(v), true
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
	lo, hi, step := numberBounds(o)
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
			changed(v)
		}
		s.Text = numberText(*value)
		if priorSubmit != nil {
			priorSubmit(s.Text)
		}
	}
	style := in.Style
	in.Style = func(root, editor, leading, trailing ui.Element) {
		editor.Bind(&s.Handle).InputMethod(false)
		root.PaddingEnd(buttonWidth + inset)
		if vertical {
			editor.TextAlign(ui.Start)
		} else {
			root.PaddingStart(buttonWidth + inset)
			editor.TextAlign(ui.Center)
		}
		if style != nil {
			style(root, editor, leading, trailing)
		}
	}
	adjust := func(delta float64) {
		if o.ReadOnly || o.Disabled || t.field.Disabled {
			return
		}
		v := float64(0)
		if value.Set {
			v = value.Value
		}
		v = math.Max(lo, math.Min(hi, v+delta*step))
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}
		if math.Abs(v) < 9000 {
			v = math.Round(v*1e12) / 1e12
		}
		changed(Set(v))
		s.Text = numberText(*value)
	}
	return inputShell(c, t, &s.Text, in, ComponentInputNumber, 0, 0, func() ui.Element {
		return ui.TextInputBase(c.Key("number"), &s.Text).OnShortcut(0, ui.KeyUp, func() { adjust(1) }).OnShortcut(0, ui.KeyDown, func() { adjust(-1) })
	}, func(root ui.Element) {
		// Absolute siblings keep the editor and both buttons at stable keys and
		// parents when orientation changes. Their space is reserved above.
		for _, increase := range []bool{false, true} {
			key, label, icon, delta := "decrease", "decreaseNamed", lucide.Minus(), -1.0
			atBound := value.Set && value.Value <= lo
			if increase {
				key, label, icon, delta = "increase", "increaseNamed", lucide.Plus(), 1
				atBound = value.Set && value.Value >= hi
			}
			if vertical {
				icon = lucide.ChevronDown()
			}
			Button(c.Key(key), t, "", ButtonOptions{
				AccessibleLabel: namedMessage(c, label, o.Label), Leading: icon,
				Variant: LinkVariant, Color: config.Color, Size: config.Size,
				Disabled: o.ReadOnly || atBound,
				Style: func(button, label, leading, trailing ui.Element) {
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
							leading.Rotate(180)
						} else {
							button.Bottom(0)
						}
						size := min(m.Icon*.8, (m.Line+2*m.Y)/2)
						leading.Size(size, size)
					} else {
						button.Top(0).Bottom(0)
					}
				},
			}).OnClick(func() { adjust(delta) })
		}
	})
}
