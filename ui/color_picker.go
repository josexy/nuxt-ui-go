package nui

import (
	"fmt"
	"strings"
	"time"

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
	Size     Size
	// Throttle limits OnChange notifications, not the bound color's immediate
	// updates. Unset uses 50 ms; an explicit zero reports every actual change.
	Throttle Optional[time.Duration]
	OnChange func(ui.Color)
	Decoration
}

// Pending notifications belong to a mounted picker, not the application's
// persistent drafts. Unmounting discards them without a background callback.
type colorPickerActivity struct {
	seen, notified ui.Color
	last           time.Time
	pending        bool
}

// ColorPicker reuses the native SV square, hue/alpha sliders and HEX editor.
// RGBA fields are library inputs. Internal SV, track and HEX parts use the
// scoped native theme; the panel and RGBA containers accept slot overrides.
func ColorPicker(c *ui.Context, t Theme, value *ui.Color, s *ColorPickerState, o ColorPickerOptions) ui.Element {
	if value == nil || s == nil {
		panic("nui: ColorPicker requires value and state")
	}
	delay := valueOr(o.Throttle, 50*time.Millisecond)
	if delay < 0 {
		panic("nui: ColorPicker throttle must be non-negative")
	}
	if o.Label == "" {
		o.Label = t.field.Label
	}
	config := selected(defaults(t, ComponentColorPicker, ComponentTheme{Size: MD}), ComponentTheme{Size: inheritedSize(t, o.Size)})
	side := t.spacing(t.text([5]float32{152, 160, 168, 176, 184}[sizeIndex(config.Size)]))
	native := t.Native()
	// The pinned native picker sizes its SV region in 40 theme units. Keep
	// that region square at its default width without touching shared themes.
	native.Spacing, native.FontSize = side/40, buttonMetrics(t, config.Size).Font
	root := ui.Themed(c, native).Background(ui.Transparent).MinWidth(0).Width(side).MaxWidthPercent(100).Disabled(o.Disabled || t.field.Disabled).Role(ui.RoleGroup).Label(o.Label).Value(fmt.Sprintf("#%02x%02x%02x%02x", value.R, value.G, value.B, value.A))
	activity := ui.Local(root, "notifications", func() colorPickerActivity { return colorPickerActivity{seen: *value, notified: *value} })
	if activity.seen != *value {
		*activity = colorPickerActivity{seen: *value, notified: *value}
	}
	emit := func() {
		activity.pending, activity.last, activity.notified = false, c.Now(), *value
		if o.OnChange != nil {
			o.OnChange(activity.notified)
		}
		if *value != activity.notified {
			*activity = colorPickerActivity{seen: *value, notified: *value}
		}
	}
	disabled := root.IsDisabled()
	if disabled || o.OnChange == nil {
		*activity = colorPickerActivity{seen: *value, notified: *value}
	} else if activity.pending {
		if remaining := delay - c.Now().Sub(activity.last); remaining > 0 {
			c.After(remaining)
		} else {
			emit()
		}
	}
	if !s.initialized || s.seen != *value {
		s.channels = [4]Optional[float64]{Set(float64(value.R)), Set(float64(value.G)), Set(float64(value.B)), Set(float64(value.A))}
		s.seen = *value
		s.initialized = true
	}
	services := c.Services()
	notify := func() {
		if disabled || *value == activity.seen {
			return
		}
		activity.seen = *value
		if o.OnChange == nil || *value == activity.notified {
			activity.pending = false
			return
		}
		activity.pending = true
		// Input callbacks can run outside a build. The next build processes
		// the notification with its valid window clock and schedules trailing work.
		services.Invalidate()
	}
	channelTheme := t
	// The four numeric drafts make one form field. Binding each draft to the
	// same FormFieldState would mark an untouched field dirty every build.
	channelTheme.field.State = nil
	root.Children(func() {
		// MyGo draws its color axes left-to-right, while Slider input normally
		// mirrors in RTL. Scope the native panel before building its controls so
		// their hit testing, keys, gradients and thumbs use the same direction.
		ui.Column(c.Key("native")).Direction(ui.LeftToRight).FillWidth().Children(func() {
			picker := ui.ColorPicker(c.Key("picker"), value).FillWidth().Label(o.Label).OnChange(notify)
			finish(t, ComponentColorPicker, Panel, picker, PartStyle{}, o.Decoration)
		})
		channels := ui.Column(c.Key("channels")).MinWidth(0)
		finish(t, ComponentColorPicker, Body, channels, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
		channels.Children(func() {
			for i, label := range []string{"R", "G", "B", "A"} {
				row := ui.Row(c.Key(label)).MinWidth(0).AlignItems(ui.Center)
				finish(t, ComponentColorPicker, Item, row, PartStyle{Gap: Set(t.spacing(8))}, o.Decoration)
				row.Children(func() {
					part := ui.Text(c, label).Width(t.text(18)).Shrink(0)
					finish(t, ComponentColorPicker, Label, part, PartStyle{Foreground: Set(t.Text), FontSize: Set(buttonMetrics(t, config.Size).Font)}, o.Decoration)
					InputNumber(c, channelTheme, &s.channels[i], &s.inputs[i], InputNumberOptions{Label: strings.TrimSpace(o.Label + " " + label), Size: config.Size, Min: Set(float64(0)), Max: Set(float64(255)), Step: 1, OnValueChange: func(v Optional[float64]) {
						if !v.Set {
							return
						}
						channel := uint8(v.Value)
						next := ui.Color{R: value.R, G: value.G, B: value.B, A: value.A}
						previous := []uint8{value.R, value.G, value.B, value.A}[i]
						if channel == previous {
							return
						}
						switch i {
						case 0:
							next.R = channel
						case 1:
							next.G = channel
						case 2:
							next.B = channel
						case 3:
							next.A = channel
						}
						*value = next
						notify()
					}}).Grow(1)
				})
			}
		})
	})
	root.Value(fmt.Sprintf("#%02x%02x%02x%02x", value.R, value.G, value.B, value.A))
	finish(t, ComponentColorPicker, Root, root, PartStyle{Gap: Set(t.spacing(10))}, o.Decoration)
	var drafts [4]string
	err, key := "", ""
	for i := range s.inputs {
		drafts[i] = s.inputs[i].Text
		if s.inputs[i].Error != "" {
			err, key = s.inputs[i].Error, "invalidNumber"
		}
	}
	trackFieldEdit(t, fmt.Sprintf("%02x%02x%02x%02x|%s", value.R, value.G, value.B, value.A, strings.Join(drafts[:], "|")), err, key)
	return root
}
