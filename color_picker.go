package nui

import (
	"github.com/egoist/mygo/ui"
)

type ColorPickerState struct {
	channels    [4]Optional[float64]
	inputs      [4]NumberState
	seen        ui.Color
	initialized bool
}
type ColorPickerOptions struct {
	Label    string
	Disabled bool
	OnChange func(ui.Color)
	Decoration
}

// ColorPicker reuses the native SV square, hue/alpha sliders and HEX editor.
// RGB fields are library inputs. Internal native parts follow Theme.Native, not slot overrides.
func ColorPicker(c *ui.Context, t Theme, value *ui.Color, s *ColorPickerState, o ColorPickerOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: ColorPicker requires value and state")
	}
	if !s.initialized || s.seen != *value {
		s.channels = [4]Optional[float64]{Set(float64(value.R)), Set(float64(value.G)), Set(float64(value.B)), Set(float64(value.A))}
		s.seen = *value
		s.initialized = true
	}
	emit := func() {
		if o.OnChange != nil {
			o.OnChange(*value)
		}
	}
	root := ui.Themed(c, t.Native()).Background(ui.Transparent).MinWidth(0).Width(280).Disabled(o.Disabled || t.field.Disabled)
	finish(t, ComponentColorPicker, Root, root, PartStyle{Gap: Set(t.spacing(10))}, o.Decoration)
	root.Children(func() {
		ui.ColorPicker(c.Key("picker"), value).FillWidth().Label(o.Label).OnChange(emit)
		ui.Column(c.Key("channels")).Gap(6).Children(func() {
			for i, label := range []string{"R", "G", "B", "A"} {
				ui.Row(c.Key(label)).AlignItems(ui.Center).Gap(8).Children(func() {
					ui.Text(c, label).Width(18)
					InputNumber(c, t, &s.channels[i], &s.inputs[i], InputNumberOptions{Label: o.Label + " " + label, Size: SM, Min: Set(float64(0)), Max: Set(float64(255)), Step: 1, OnValueChange: func(v Optional[float64]) {
						if !v.Set {
							return
						}
						channel := uint8(v.Value)
						switch i {
						case 0:
							value.R = channel
						case 1:
							value.G = channel
						case 2:
							value.B = channel
						case 3:
							value.A = channel
						}
						emit()
					}}).Grow(1)
				})
			}
		})
	})
	return root
}
