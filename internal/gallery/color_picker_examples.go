package gallery

import (
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func addColorPickerContract(p *componentPage) {
	value := ui.Hex("#11223344")
	last := value
	var state nui.ColorPickerState
	size := nui.MD
	disabled, throttled, styled, rtl := false, false, false, false
	changes := 0
	p.Scenes = append(p.Scenes, scene("contract", "Color values, size and notifications", "Edit HEX or RGBA and compare the current color with the last notification. This example starts with immediate notifications; enable the library's default 50 ms throttle to combine rapid changes. External replacement stays quiet.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(12).Children(func() {
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Disable contract color"})
			nui.Checkbox(c, t, &throttled, nui.CheckboxOptions{Label: "Use default color throttle"})
			nui.Checkbox(c, t, &styled, nui.CheckboxOptions{Label: "Style color parts"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL color form"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c.Key("size"), t, "Cycle color size", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() {
				if size == nui.XL {
					size = nui.XS
				} else {
					size++
				}
			})
			nui.Button(c.Key("external"), t, "Set translucent blue", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { value = ui.Hex("#11223344") })
		})
		o := nui.ColorPickerOptions{Label: "Contract color", Size: size, Throttle: nui.Set(time.Duration(0)), OnChange: func(v ui.Color) { last = v; changes++ }}
		if throttled {
			o.Throttle = nui.Optional[time.Duration]{}
		}
		if styled {
			o.UI = nui.Styles{nui.Panel: {Radius: nui.Set(float32(8))}, nui.Body: {Gap: nui.Set(float32(10))}, nui.Label: {Foreground: nui.Set(t.Info)}, nui.Root: {BorderWidth: nui.Set(float32(1)), BorderColor: nui.Set(t.Border)}}
		}
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("scope")).Direction(direction).MinWidth(0).Disabled(disabled).Children(func() { nui.ColorPicker(c.Key("picker"), t, &value, &state, o) })
		ui.Textf(c, "Color value: #%02X%02X%02X%02X · Size: %s", value.R, value.G, value.B, value.A, [...]string{"", "XS", "SM", "MD", "LG", "XL"}[size])
		ui.Textf(c, "Color notified: #%02X%02X%02X%02X · Changes: %d", last.R, last.G, last.B, last.A, changes)
	}))
}
