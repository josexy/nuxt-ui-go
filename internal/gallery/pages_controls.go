package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

var semanticColors = []struct {
	Name  string
	Value nui.Color
}{{"Primary", nui.Primary}, {"Secondary", nui.Secondary}, {"Success", nui.Success}, {"Info", nui.Info}, {"Warning", nui.Warning}, {"Error", nui.Error}, {"Neutral", nui.Neutral}}
var fiveSizes = []struct {
	Name  string
	Value nui.Size
}{{"XS", nui.XS}, {"SM", nui.SM}, {"MD", nui.MD}, {"LG", nui.LG}, {"XL", nui.XL}}
var buttonVariants = []struct {
	Name  string
	Value nui.Variant
}{{"Solid", nui.Solid}, {"Outline", nui.Outline}, {"Soft", nui.Soft}, {"Subtle", nui.Subtle}, {"Ghost", nui.Ghost}, {"Link", nui.LinkVariant}}

func buttonPage() *componentPage {
	clicks, searches, blocked := 0, 0, 0
	return &componentPage{Scenes: []exampleScene{
		scene("usage", "Basic usage", "Click, press Enter or Space. Each activation updates the count once.", func(c *ui.Context, t nui.Theme) {
			nui.Button(c.Key("save"), t, "Save changes", nui.ButtonOptions{Leading: lucide.Check()}).OnClick(func() { clicks++ })
			ui.Textf(c, "Saved %d times", clicks).Role(ui.RoleStatus).TextColor(t.Muted)
		}),
		scene("variants", "Variants", "Compare all six button variants.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(10).Children(func() {
				for _, v := range buttonVariants {
					nui.Button(c.Key(v.Name), t, v.Name, nui.ButtonOptions{Variant: v.Value})
				}
			})
		}),
		scene("colors", "Colors", "Semantic colors communicate the action's purpose.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(10).Children(func() {
				for _, v := range semanticColors {
					nui.Button(c.Key(v.Name), t, v.Name, nui.ButtonOptions{Color: v.Value, Variant: nui.Soft})
				}
			})
		}),
		scene("sizes", "Sizes", "Compare XS through XL, with the same leading icon.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(10).AlignItems(ui.Center).Children(func() {
				for _, v := range fiveSizes {
					nui.Button(c.Key(v.Name), t, v.Name, nui.ButtonOptions{Size: v.Value, Leading: lucide.Plus()})
				}
			})
		}),
		scene("icons", "Icons and custom parts", "An icon-only button has an accessible name. Part styling affects this instance.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(10).Children(func() {
				nui.Button(c, t, "Next", nui.ButtonOptions{Trailing: lucide.ArrowRight()})
				nui.Button(c, t, "", nui.ButtonOptions{AccessibleLabel: "Search library", Leading: lucide.Search(), Square: true}).OnClick(func() { searches++ })
				nui.Button(c, t, "Custom label", nui.ButtonOptions{Variant: nui.Soft, Style: func(root, label, leading, trailing ui.Element) { label.FontWeight(700) }})
			})
			ui.Textf(c, "Searches: %d", searches)
		}),
		scene("states", "Disabled and loading", "Neither control accepts activation while blocked.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(10).Children(func() {
				nui.Button(c, t, "Disabled", nui.ButtonOptions{Disabled: true}).OnClick(func() { blocked++ })
				nui.Button(c, t, "Loading", nui.ButtonOptions{Loading: true}).OnClick(func() { blocked++ })
			})
			ui.Textf(c, "Blocked callbacks: %d", blocked)
		}),
	}}
}

func inputPage() *componentPage {
	name, invalid, readonly := "Nuxt workspace", "team@", "workspace_01"
	disabled, loading := "disabled_01", "loading_01"
	variants, sizes := make([]string, 5), make([]string, 5)
	return &componentPage{Scenes: []exampleScene{
		scene("usage", "Basic usage", "Edit English or Chinese text. The draft survives navigation and theme changes.", func(c *ui.Context, t nui.Theme) {
			nui.Input(c.Key("workspace-name"), t, &name, nui.InputOptions{Label: "Workspace name", Leading: lucide.Search(), Trailing: lucide.Check(), Clearable: true}).FillWidth()
			ui.Text(c, "Current value: "+name).Role(ui.RoleStatus).TextColor(t.Muted)
		}),
		scene("variants", "Variants", "Each input has its own value. Input supports Outline, Soft, Subtle, Ghost and None.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(10).Children(func() {
				for i, v := range []struct {
					Name  string
					Value nui.Variant
				}{{"Outline", nui.Outline}, {"Soft", nui.Soft}, {"Subtle", nui.Subtle}, {"Ghost", nui.Ghost}, {"None", nui.None}} {
					nui.Input(c.Key(v.Name), t, &variants[i], nui.InputOptions{Label: v.Name + " input", Placeholder: v.Name, Variant: v.Value}).Width(160)
				}
			})
		}),
		scene("sizes", "Sizes", "Five sizes keep independent drafts.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(10).Children(func() {
				for i, v := range fiveSizes {
					nui.Input(c.Key(v.Name), t, &sizes[i], nui.InputOptions{Label: v.Name + " input", Placeholder: v.Name, Size: v.Value}).Width(150)
				}
			})
		}),
		scene("states", "Editing states", "Invalid fields expose an error. Read-only fields remain selectable; disabled and loading fields cannot edit.", func(c *ui.Context, t nui.Theme) {
			nui.Input(c.Key("invalid-email"), t, &invalid, nui.InputOptions{Label: "Email", Invalid: true, ErrorMessage: "Enter a complete email address"}).FillWidth()
			ui.Text(c, "Enter a complete email address").TextColor(t.Error)
			nui.Input(c.Key("readonly"), t, &readonly, nui.InputOptions{Label: "Workspace ID", ReadOnly: true}).FillWidth()
			nui.Input(c, t, &disabled, nui.InputOptions{Label: "Disabled input", Disabled: true}).FillWidth()
			nui.Input(c, t, &loading, nui.InputOptions{Label: "Loading input", Loading: true}).FillWidth()
			ui.Text(c, fmt.Sprintf("Read-only ID: %s", readonly)).TextColor(t.Muted)
		}),
	}}
}
