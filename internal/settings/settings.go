// Package settings composes forms, menus and dialogs around an application-owned model.
package settings

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type State struct {
	Wizard                                                       WizardState
	EditOpen, ConfirmReset, FiltersOpen, EditAttempted, Advanced bool
	EditName, EditEmail                                          string
	FilterTab                                                    string
	ActionsMenu                                                  nui.MenuState
	EditTrigger, ActionsTrigger                                  ui.Handle
	ReturnFocus                                                  *ui.Handle

	Config                           nui.AppConfig
	Page                             int
	Name, Email, Bio, Password, Slug string
	Notify, Marketing, Sync, Policy  bool
	Channels                         []string
	Plan                             string
	Region                           nui.Optional[string]
	Owner                            nui.SearchState
	Volume                           float64
	Help, Attempted                  bool
	Saved                            int
	Message                          string
}

func New() *State {
	return &State{Config: nui.DefaultApp(), Name: "Lin Chen / 林晓", Email: "lin@example.com", Bio: "为团队构建可复用的原生桌面工具。", Slug: "desktop-team", Notify: true, Sync: true, Policy: true, Channels: []string{"email"}, Plan: "team", Volume: 40}
}
func (a *State) View(c *ui.Context) {
	t := nui.App(c, a.Config)
	nui.ToastViewport(c, t, nui.ToastViewportOptions{})

	ui.Column(c).Fill().Children(func() {
		ui.Row(c).Padding(16, 24).Gap(12).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			ui.Text(c, "Workspace settings").FontSize(t.FontSize + 4).FontWeight(600).Grow(1)
			nui.Button(c, t, "Theme", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral, Leading: lucide.Moon()}).OnClick(func() {
				if t.Dark {
					a.Config.Mode = nui.LightMode
				} else {
					a.Config.Mode = nui.DarkMode
				}
			})
		})
		a.Panel(c, t).Grow(1)
	})
}
func (a *State) fields(c *ui.Context, t nui.Theme) ui.Element {
	return ui.Row(c).MinHeight(0).MinWidth(0).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Width(150).Padding(20, 12).Gap(6).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
			for i, label := range []string{"Profile", "Notifications", "Workspace", "Project setup"} {
				variant := nui.Ghost
				if a.Page == i {
					variant = nui.Soft
				}
				nui.Button(c.Key(label), t, label, nui.ButtonOptions{Variant: variant, Color: nui.Neutral, AccessibleLabel: "Settings: " + label}).FillWidth().Justify(ui.Start).OnClick(func() { a.Page = i })
			}
		})
		nui.ScrollArea(c.Key(fmt.Sprintf("settings-%d", a.Page)), t, nui.ScrollAreaOptions{UI: nui.Styles{nui.Body: {PaddingX: nui.Set(float32(28)), PaddingY: nui.Set(float32(24)), Gap: nui.Set(float32(20))}}}, func() {
			nui.Container(c, t, nui.ContainerOptions{MaxWidth: nui.Set(float32(720)), Padding: nui.Set(float32(0))}, func() {
				switch a.Page {
				case 0:
					a.profile(c, t)
				case 1:
					a.notifications(c, t)
				case 2:
					a.workspace(c, t)
				case 3:
					a.wizard(c, t)
				}
				if a.Page == 3 {
					return
				}
				nui.Separator(c, t, nui.SeparatorOptions{})
				ui.Row(c).Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
					nui.Button(c, t, "Save changes", nui.ButtonOptions{Leading: lucide.Check()}).OnClick(func() {
						before := a.Saved
						a.save()
						if a.Saved > before {
							nui.Toast(c, t, nui.ToastOptions{Title: "Settings saved", Description: "Your changes are saved in this window.", Color: nui.Success})
						}
					})
					if a.Message != "" {
						ui.Text(c, a.Message).TextColor(t.Muted).Role(ui.RoleStatus).MinWidth(0)
					}
				})
			})
		}).Grow(1)
	})
}
func title(c *ui.Context, t nui.Theme, name, description string) {
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, name).FontSize(24).FontWeight(600).TextColor(t.Highlighted)
		ui.Text(c, description).TextColor(t.Muted)
	})
}
func (a *State) nameError() string {
	if a.Attempted && strings.TrimSpace(a.Name) == "" {
		return "Enter your display name."
	}
	return ""
}
func (a *State) emailError() string {
	if a.Attempted && !strings.Contains(a.Email, "@") {
		return "Enter an email address with @."
	}
	return ""
}
func (a *State) policyError() string {
	if a.Attempted && !a.Policy {
		return "Accept the workspace policy to save."
	}
	return ""
}
func (a *State) save() {
	a.Attempted = true
	if a.nameError() != "" || a.emailError() != "" {
		a.Page = 0
		a.Message = "Review the highlighted fields."
		return
	}
	if a.policyError() != "" {
		a.Page = 1
		a.Message = "Review the highlighted fields."
		return
	}
	a.Saved++
	a.Message = fmt.Sprintf("Changes saved · %d", a.Saved)
}
func (a *State) profile(c *ui.Context, t nui.Theme) {
	title(c, t, "Profile", "Manage the details your teammates see.")
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Display name", Required: true, Error: a.nameError()}, func(inner nui.Theme) {
		nui.Input(c, inner, &a.Name, nui.InputOptions{Placeholder: "Your name", Clearable: true, ClearLabel: "Clear display name"}).FillWidth()
	})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Email address", Required: true, Help: "Used for invitations and account updates.", Error: a.emailError()}, func(inner nui.Theme) {
		nui.Input(c, inner, &a.Email, nui.InputOptions{Placeholder: "you@example.com"}).FillWidth()
	})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "About you", Help: "A short introduction to your role and current work."}, func(inner nui.Theme) {
		nui.Textarea(c, inner, &a.Bio, nui.TextareaOptions{Placeholder: "Tell your team about your work", Rows: 3, MaxRows: 6, AutoResize: true}).FillWidth()
	})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Password", Hint: "Optional"}, func(inner nui.Theme) {
		nui.Input(c, inner, &a.Password, nui.InputOptions{Password: true, Placeholder: "Enter a new password"}).FillWidth()
	})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Workspace address"}, func(inner nui.Theme) {
		nui.FieldGroup(c, inner, nui.FieldGroupOptions{}, []nui.FieldGroupItem{
			{ID: "slug", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Input(c, t, &a.Slug, nui.InputOptions{Placeholder: "workspace-name"}).Grow(1)
			}},
			{ID: "check", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Button(c, t, "Check", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() {
					if strings.TrimSpace(a.Slug) == "" {
						a.Message = "Enter a workspace address."
					} else {
						a.Message = "Workspace address checked locally."
					}
				})
			}},
		}).FillWidth()
	})
}
func (a *State) notifications(c *ui.Context, t nui.Theme) {
	title(c, t, "Notifications", "Choose what reaches you and how it arrives.")
	nui.Switch(c, t, &a.Notify, nui.SwitchOptions{Label: "Enable notifications", Description: "Receive updates about the work you follow."})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Notification channels", Disabled: !a.Notify}, func(inner nui.Theme) {
		nui.CheckboxGroup(c, inner, &a.Channels, []nui.Option[string]{{Value: "email", Label: "Email", Description: "A summary in your inbox."}, {Value: "desktop", Label: "Desktop", Description: "Show a notification on this computer."}, {Value: "sms", Label: "SMS", Description: "Unavailable in this workspace.", Disabled: true}}, nui.CheckboxGroupOptions{Label: "Notification channels"})
	})
	nui.Checkbox(c, t, &a.Marketing, nui.CheckboxOptions{Label: "Product announcements", Description: "Occasional news about features and releases."})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Notification volume", Help: fmt.Sprintf("Current volume: %.0f%%", a.Volume), Disabled: !a.Notify}, func(inner nui.Theme) {
		nui.Slider(c, inner, &a.Volume, nui.SliderOptions{Label: "Notification volume", Step: 5})
	})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Workspace policy", Required: true, Error: a.policyError()}, func(inner nui.Theme) {
		nui.Checkbox(c, inner, &a.Policy, nui.CheckboxOptions{Label: "I accept the workspace policy"})
	})
}
func (a *State) workspace(c *ui.Context, t nui.Theme) {
	title(c, t, "Workspace", "Set defaults for this team's desktop tools.")
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Plan"}, func(inner nui.Theme) {
		nui.RadioGroup(c, inner, &a.Plan, []nui.Option[string]{{Value: "personal", Label: "Personal", Description: "For individual work."}, {Value: "team", Label: "Team", Description: "Share projects with your teammates."}, {Value: "enterprise", Label: "Enterprise", Description: "Contact an administrator.", Disabled: true}}, nui.RadioGroupOptions{Label: "Plan"})
	})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Region", Help: "Choose the default region for new projects."}, func(inner nui.Theme) {
		nui.Select(c, inner, &a.Region, []nui.Option[nui.Optional[string]]{{Value: nui.Optional[string]{}, Label: "No region"}, {Value: nui.Set("eu"), Label: "Europe", Description: "Frankfurt"}, {Value: nui.Set("ap"), Label: "Asia Pacific", Description: "Singapore"}, {Value: nui.Set("us"), Label: "United States", Disabled: true}}, nui.SelectOptions{Label: "Region"}).FillWidth()
	})
	nui.FormField(c, t, nui.FormFieldOptions{Label: "Workspace owner", Help: "Search by name or team. A typed draft does not change the selected owner."}, func(inner nui.Theme) {
		nui.InputMenu(c, inner, &a.Owner, []nui.Option[string]{{Value: "alex-berlin", Label: "Alex Chen", Description: "Berlin team"}, {Value: "alex-taipei", Label: "Alex Chen", Description: "Taipei team"}, {Value: "lin", Label: "林晓", Description: "Desktop team", Icon: lucide.UserRound()}, {Value: "guest", Label: "Guest", Description: "Cannot own a workspace", Disabled: true}}, nui.InputMenuOptions{Label: "Workspace owner", Placeholder: "Search people…", Leading: lucide.Search(), EmptyText: "No people match your search", FilterFields: []string{"label", "description"}}).FillWidth()
	})
	nui.Switch(c, t, &a.Sync, nui.SwitchOptions{Label: "Sync preferences", Description: "Keep these preferences available on your other computers."})
	help := nui.Button(c, t, "About workspace access", nui.ButtonOptions{Variant: nui.LinkVariant, Leading: lucide.Info()}).OnClick(func() { a.Help = !a.Help })
	nui.Popover(c, t, help, &a.Help, nui.PopoverOptions{Label: "Workspace access", Width: nui.Set(float32(300))}, func(inner nui.Theme) {
		ui.Text(c, "Workspace access").FontWeight(600)
		ui.Text(c, "Your administrator controls membership. Changing the owner here only changes this example's local settings.").TextColor(inner.Muted)
		nui.Button(c, inner, "Got it", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { a.Help = false })
	})
}
