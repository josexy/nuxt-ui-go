package nui

import "github.com/egoist/mygo/ui"

func carouselButtonOptions(custom *ButtonOptions, override, fallback *ui.SVG, label string, disabled bool) ButtonOptions {
	o := ButtonOptions{Square: true}
	if custom != nil {
		o = *custom
	}
	if o.Color == ColorDefault {
		o.Color = Neutral
	}
	if o.Variant == VariantDefault {
		o.Variant = Outline
	}
	if o.AccessibleLabel == "" {
		o.AccessibleLabel = label
	}
	if o.Leading == nil && o.Trailing == nil {
		o.Leading = override
		if o.Leading == nil && o.Avatar == nil {
			o.Leading = fallback
		}
	}
	o.Disabled = o.Disabled || disabled
	o.UI = mergeStyles(Styles{Root: {Radius: Set(float32(999))}}, o.UI)
	return o
}
