package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func (a *State) formStates(c *ui.Context, t nui.Theme) {
	heading(c, t, "Form states", "Field-specific variants, sizing, disabled states and grouped controls.")
	section(c, t, "Input variants", func() {
		ui.Row(c).Gap(10).Wrap().Children(func() {
			for _, sample := range []struct {
				name    string
				variant nui.Variant
			}{{"Outline", nui.Outline}, {"Soft", nui.Soft}, {"Subtle", nui.Subtle}, {"Ghost", nui.Ghost}, {"None", nui.None}} {
				nui.Input(c.Key(sample.name), t, &a.Name, nui.InputOptions{Label: sample.name + " input", Variant: sample.variant, Leading: lucide.Search()}).Width(145)
			}
		})
	})
	section(c, t, "Five sizes", func() {
		ui.Row(c).Gap(10).Wrap().AlignItems(ui.Center).Children(func() {
			for i, name := range []string{"XS", "SM", "MD", "LG", "XL"} {
				nui.Input(c.Key(name), t, &a.Name, nui.InputOptions{Label: name + " input", Size: nui.Size(i + 1)}).Width(130)
			}
		})
	})
	section(c, t, "Editing states", func() {
		ui.Row(c).Gap(12).Wrap().Children(func() {
			nui.Input(c, t, &a.Invalid, nui.InputOptions{Label: "Invalid email", Invalid: true, ErrorMessage: "Use a valid email address", Clearable: true}).Width(230)
			nui.Input(c, t, &a.ReadOnly, nui.InputOptions{Label: "Read-only ID", ReadOnly: true}).Width(230)
			nui.Input(c, t, &a.Name, nui.InputOptions{Label: "Disabled input", Disabled: true}).Width(230)
		})
	})
	section(c, t, "Layout and empty choices", func() {
		ui.Row(c).Gap(24).Wrap().AlignItems(ui.Start).Children(func() {
			nui.FieldGroup(c, t, nui.FieldGroupOptions{Vertical: true}, []nui.FieldGroupItem{{ID: "input", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Input(c, t, &a.Name, nui.InputOptions{Label: "Joined name"})
			}}, {ID: "button", Build: func(c *ui.Context, t nui.Theme) ui.Element {
				return nui.Button(c, t, "Validate", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { a.Message = "Name checked" })
			}}}).Width(230)
			nui.Select(c, t, &a.Plan, nil, nui.SelectOptions{Label: "Empty select", Placeholder: "No options available"}).Width(230)
			nui.InputMenu(c, t, &a.Settings.Owner, nil, nui.InputMenuOptions{Label: "Empty search", Placeholder: "Search an empty list"}).Width(230)
		})
	})
	section(c, t, "Vertical slider", func() {
		ui.Row(c).Gap(16).Children(func() {
			nui.Slider(c, t, &a.Percent, nui.SliderOptions{Label: "Vertical progress", Vertical: true, Step: 5}).Height(100)
			ui.Text(c, fmt.Sprintf("%.0f%%", a.Percent)).TextColor(t.Muted)
		})
	})
	ui.Text(c, a.Message).TextColor(t.Muted)
}
