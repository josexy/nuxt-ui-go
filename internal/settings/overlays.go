package settings

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func (a *State) Panel(c *ui.Context, t nui.Theme) ui.Element {
	return ui.Column(c).MinHeight(0).MinWidth(0).Children(func() {
		ui.Row(c).Wrap().Padding(12, 20).Gap(8).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			edit := nui.Button(c, t, "Edit profile", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral, Size: nui.SM}).Bind(&a.EditTrigger).OnClick(func() { a.beginEdit(&a.EditTrigger) })
			nui.Tooltip(c, t, edit, "Edit your name and email together", nui.TooltipOptions{})
			actions := nui.Button(c, t, "Actions", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral, Size: nui.SM, Trailing: lucide.ChevronDown()}).Bind(&a.ActionsTrigger)
			nui.DropdownMenu(c, t, actions, &a.ActionsMenu, []nui.MenuItem{
				{ID: "edit", Label: "Edit profile details", Icon: lucide.UserRound(), OnSelect: func() { a.beginEdit(&a.ActionsTrigger) }},
				{ID: "notify", Label: "Notification presets", Children: []nui.MenuItem{
					{ID: "quiet", Label: "Quiet mode", Description: "Pause notification channels", OnSelect: func() {
						a.Notify = false
						nui.Toast(c, t, nui.ToastOptions{Title: "Quiet mode enabled", Color: nui.Info})
					}},
					{ID: "standard", Label: "Standard mode", Description: "Email and desktop notifications", OnSelect: func() { a.Notify = true; a.Channels = []string{"email", "desktop"} }},
				}},
				{ID: "divider", Kind: nui.MenuSeparator},
				{ID: "reset", Label: "Reset notification preferences", Icon: lucide.CircleAlert(), Color: nui.Error, OnSelect: func() { a.ConfirmReset = true; a.ReturnFocus = &a.ActionsTrigger }},
			}, nui.MenuOptions{Label: "Settings actions", Width: nui.Set(float32(300))})
			nui.Button(c, t, "Notification filters", nui.ButtonOptions{Variant: nui.Ghost, Color: nui.Neutral, Size: nui.SM, Leading: lucide.Settings()}).OnClick(func() { a.FiltersOpen = true })
		})
		a.fields(c, t).Grow(1)
		a.overlays(c, t)
	})
}
func (a *State) beginEdit(opener *ui.Handle) {
	a.EditName, a.EditEmail = a.Name, a.Email
	a.EditAttempted = false
	a.ReturnFocus = opener
	a.EditOpen = true
}
func (a *State) overlays(c *ui.Context, t nui.Theme) {
	nui.ModalWithOptions(c.Key("edit-profile"), t, &a.EditOpen, nui.ModalOptions{
		Title: "Edit profile details", Description: "Update how your teammates find you.", ReturnFocus: a.ReturnFocus,
		Footer: func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Cancel edit", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { a.EditOpen = false })
			nui.Button(c, t, "Apply profile", nui.ButtonOptions{}).OnClick(func() {
				a.EditAttempted = true
				if strings.TrimSpace(a.EditName) == "" || !strings.Contains(a.EditEmail, "@") {
					return
				}
				a.Name, a.Email = a.EditName, a.EditEmail
				a.EditOpen = false
				nui.Toast(c, t, nui.ToastOptions{Title: "Profile updated", Color: nui.Success})
			})
		},
	}, func(t nui.Theme) {
		nameError, emailError := "", ""
		if a.EditAttempted {
			if strings.TrimSpace(a.EditName) == "" {
				nameError = "Enter your display name."
			}
			if !strings.Contains(a.EditEmail, "@") {
				emailError = "Enter an email address with @."
			}
		}
		nui.FormField(c, t, nui.FormFieldOptions{Label: "Profile name", Required: true, Error: nameError}, func(t nui.Theme) { nui.Input(c, t, &a.EditName, nui.InputOptions{}).FillWidth() })
		nui.FormField(c, t, nui.FormFieldOptions{Label: "Profile email", Required: true, Error: emailError}, func(t nui.Theme) { nui.Input(c, t, &a.EditEmail, nui.InputOptions{}).FillWidth() })
	})
	nui.ModalWithOptions(c.Key("reset-notifications"), t, &a.ConfirmReset, nui.ModalOptions{
		Title: "Reset notifications?", Description: "Restore the notification defaults for this window. You can undo this change.", ReturnFocus: a.ReturnFocus,
		Footer: func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Keep preferences", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() { a.ConfirmReset = false })
			nui.Button(c, t, "Reset notifications", nui.ButtonOptions{Color: nui.Error}).OnClick(func() {
				channels, notify, volume := slices.Clone(a.Channels), a.Notify, a.Volume
				a.Channels = []string{"email"}
				a.Notify = true
				a.Volume = 40
				a.ConfirmReset = false
				nui.Toast(c, t, nui.ToastOptions{Title: "Notifications reset", Action: "Undo reset", Color: nui.Success, OnAction: func() { a.Channels = channels; a.Notify = notify; a.Volume = volume }})
			})
		},
	}, func(t nui.Theme) {
		ui.Text(c, "Email notifications will be enabled and the volume will return to 40%.").TextColor(t.Muted)
	})
	nui.Slideover(c.Key("notification-filters"), t, &a.FiltersOpen, nui.SlideoverOptions{
		Title: "Notification filters", Description: "Choose which updates reach you.", Footer: func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Done filtering", nui.ButtonOptions{}).OnClick(func() { a.FiltersOpen = false })
		}, Extent: nui.Set(float32(420))}, func(t nui.Theme) {
		nui.Tabs(c, t, &a.FilterTab, []nui.TabItem{
			{ID: "delivery", Label: "Delivery", Content: func(c *ui.Context, t nui.Theme) {
				nui.Switch(c, t, &a.Notify, nui.SwitchOptions{Label: "Deliver notifications"})
				nui.Checkbox(c, t, &a.Marketing, nui.CheckboxOptions{Label: "Include product news"})
			}},
			{ID: "volume", Label: "Sound", Content: func(c *ui.Context, t nui.Theme) {
				nui.Slider(c, t, &a.Volume, nui.SliderOptions{Label: "Filter volume", Step: 5})
			}},
		}, nui.TabsOptions{Label: "Filter category"})
		nui.Collapsible(c, t, &a.Advanced, nui.CollapsibleOptions{Label: "Advanced delivery"}, func(t nui.Theme) { nui.Switch(c, t, &a.Sync, nui.SwitchOptions{Label: "Sync notification filters"}) })
	})
}
