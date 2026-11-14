package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func addMarqueeContract(p *componentPage, a *State) {
	var state nui.MarqueeState
	copies, generation := 4, 0
	vertical, reverse, rtl, disabled, styled := false, false, false, false, false
	overlay, hover := true, false
	p.Scenes = append(p.Scenes, scene("contract", "Marquee copies and motion", "Repeat defaults to four; zero builds no content and one may leave an empty interval. Only presentation content belongs in the repeated callback. Compare directions, overlay and content styling; Space pauses the focused region.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			for _, n := range []int{0, 1, 2, 4, 6} {
				nui.Button(c.Key(n), t, fmt.Sprintf("Marquee copies %d", n), nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { copies = n })
			}
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Checkbox(c, t, &vertical, nui.CheckboxOptions{Label: "Vertical contract marquee"})
			nui.Checkbox(c, t, &reverse, nui.CheckboxOptions{Label: "Reverse contract marquee"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL contract marquee"})
			nui.Checkbox(c, t, &overlay, nui.CheckboxOptions{Label: "Fade contract marquee"})
			nui.Checkbox(c, t, &hover, nui.CheckboxOptions{Label: "Pause marquee on hover"})
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Disable marquee parent"})
			nui.Checkbox(c, t, &styled, nui.CheckboxOptions{Label: "Style marquee content"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c, t, "Pause / resume contract marquee", nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { state.Paused = !state.Paused })
			nui.Button(c, t, "Replace marquee content", nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { generation++ })
		})
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		o := nui.MarqueeOptions{Label: "Contract marquee", Height: 150, Repeat: nui.Set(copies), Reverse: reverse, Vertical: vertical, PauseOnHover: hover, Overlay: nui.Set(overlay), ContentKey: fmt.Sprint(generation), Active: nui.Set(a.Animations && !a.WindowInactive)}
		if styled {
			o.UI = nui.Styles{nui.Body: {Background: nui.Set(t.Elevated), PaddingX: nui.Set(float32(12)), PaddingY: nui.Set(float32(8)), Radius: nui.Set(float32(8)), Gap: nui.Set(float32(12))}}
		}
		ui.Column(c.Key("direction")).Direction(direction).Disabled(disabled).Children(func() {
			nui.Marquee(c.Key("marquee"), t, &state, o, func() {
				nui.Badge(c.Key("release"), t, fmt.Sprintf("Marquee release %d", generation), nui.BadgeOptions{Variant: nui.Soft})
				nui.Badge(c.Key("ready"), t, "Ready for review", nui.BadgeOptions{Variant: nui.Outline})
			})
		})
		ui.Textf(c, "Copies: %d · Paused: %t · Content: %d", copies, state.Paused, generation)
	}))
}

func addSplitterContract(p *componentPage) {
	value, changes := float32(40), 0
	draft := "Editable panel"
	disabled, parentDisabled, vertical, rtl, styled, zeroStep := false, false, false, false, false, false
	p.Scenes = append(p.Scenes, scene("contract", "Splitter limits and panel editing", "Disabled locks the divider while the editor remains usable. Disabling the parent locks both. Arrow keys move five percentage points; Home/End reach the allowed range. Repeated boundary input does not emit changes.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Lock contract divider"})
			nui.Checkbox(c, t, &parentDisabled, nui.CheckboxOptions{Label: "Disable splitter parent"})
			nui.Checkbox(c, t, &vertical, nui.CheckboxOptions{Label: "Vertical contract splitter"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL contract splitter"})
			nui.Checkbox(c, t, &styled, nui.CheckboxOptions{Label: "Style contract divider"})
			nui.Checkbox(c, t, &zeroStep, nui.CheckboxOptions{Label: "Zero splitter arrow step"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c, t, "Set splitter to maximum", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { value = 70 })
			nui.Button(c, t, "Reset contract splitter", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { value = 40; changes = 0 })
		})
		step := float32(5)
		if zeroStep {
			step = 0
		}
		o := nui.SplitterOptions{Label: "Contract divider", Percent: true, Vertical: vertical, Disabled: disabled, Min: nui.Set(float32(30)), Max: nui.Set(float32(70)), KeyboardResizeBy: nui.Set(step), OnChange: func(float32) { changes++ }}
		if styled {
			o.UI = nui.Styles{nui.SeparatorSlot: {Background: nui.Set(t.Accented)}, nui.Indicator: {Background: nui.Set(t.Primary)}, nui.Leading: {Background: nui.Set(t.Elevated)}}
		}
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("direction")).Direction(direction).Disabled(parentDisabled).Children(func() {
			nui.Splitter(c.Key("splitter"), t, &value, o, func() {
				ui.Column(c).Fill().Padding(12).Gap(8).Children(func() {
					ui.Text(c, "First contract pane")
					nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Contract pane editor"})
				})
			}, func() {
				ui.Column(c).Fill().Padding(12).Gap(8).Children(func() { ui.Text(c, "Second contract pane"); ui.Text(c, draft) })
			}).Height(280).Border(1, t.Border).Label("Contract panes")
		})
		ui.Textf(c, "Splitter value: %.0f · Changes: %d", value, changes)
	}))
}
