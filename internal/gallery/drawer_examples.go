package gallery

import (
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func addDrawerContract(p *componentPage) {
	state := nui.DrawerState{Snap: 1}
	var opener ui.Handle
	dismissible, handle, inset, closeButton, blank, custom := true, true, false, false, false, false
	snaps, closed, prevented := 0, 0, 0
	draft := "Review build"
	p.Deactivate = append(p.Deactivate, func() { state.Open = false })
	p.Scenes = append(p.Scenes, scene("contract", "Drawer dismissal and handle", "Home/End and Up/Down change snap only when the target differs. Protected dismissal reports Escape, outside clicks and closing drags. The handle stays centered independently of the title and close button; inset and custom indicator styles can be changed while open.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(10).Children(func() {
			nui.Checkbox(c, t, &dismissible, nui.CheckboxOptions{Label: "Allow contract drawer dismissal"})
			nui.Checkbox(c, t, &closeButton, nui.CheckboxOptions{Label: "Show contract drawer close"})
			nui.Checkbox(c, t, &custom, nui.CheckboxOptions{Label: "Style contract drawer indicator"})
		})
		nui.Button(c, t, "Open contract drawer", nui.ButtonOptions{}).Bind(&opener).OnClick(func() { state.Open = true })
		title := "Contract drawer"
		if blank {
			title = ""
		}
		var parts nui.Styles
		if custom {
			parts = nui.Styles{nui.Indicator: {Background: nui.Set(t.Secondary)}}
		}
		nui.Drawer(c.Key("contract-drawer"), t, &state, nui.DrawerOptions{Title: title, Description: "Edit a draft, resize the sheet and inspect notification counts.", Dismissible: nui.Set(dismissible), CloseButton: nui.Set(closeButton), ReturnFocus: &opener, OnClose: func() { closed++ }, OnClosePrevent: func() { prevented++ }, UI: parts, Footer: func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c, t, "Toggle drawer handle", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { handle = !handle })
				nui.Button(c, t, "Toggle drawer inset", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { inset = !inset })
				nui.Button(c, t, "Toggle drawer title", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { blank = !blank })
				nui.Button(c, t, "Close drawer from app", nui.ButtonOptions{}).OnClick(func() { state.Open = false })
			})
			ui.Textf(c, "Snap: %d · Changes: %d · Closed: %d · Prevented: %d", state.Snap, snaps, closed, prevented)
		}, Handle: nui.Set(handle), Inset: inset, SnapPoints: []float32{.3, .55, .85}, OnSnap: func(int) { snaps++ }}, func(t nui.Theme) {
			nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Contract drawer draft"})
			for i := range 12 {
				ui.Textf(c, "Review item %02d", i+1)
			}
		})
		ui.Textf(c, "Snap: %d · Changes: %d · Closed: %d · Prevented: %d", state.Snap, snaps, closed, prevented)
	}))
}
