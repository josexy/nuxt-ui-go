package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func (a *State) workspace(c *ui.Context, t nui.Theme) {
	nui.Main(c, t, func() {
		nui.Container(c, t, nui.ContainerOptions{MaxWidth: nui.Set(float32(920)), Padding: nui.Set(float32(0))}, func() {
			heading(c, t, "Workspace overview", "A workspace built with reusable native components.")
			nui.Card(c, t, nui.CardOptions{Title: "NuxtUI Go", Description: "Reusable native components for everyday desktop tools.", Footer: func() {
				nui.Button(c.Key("workspace-save"), t, "Save workspace", nui.ButtonOptions{Leading: lucide.Check()}).OnClick(func() { a.Saved++; a.Message = fmt.Sprintf("Workspace saved %d times", a.Saved) })
				nui.Button(c, t, "Preferences", nui.ButtonOptions{Variant: nui.Outline, Leading: lucide.Settings()}).OnClick(func() { a.Navigate("theme") })
			}}, func() {
				nui.User(c, t, nui.UserOptions{Name: "Desktop team", Description: "4 members · updated today", Avatar: nui.AvatarOptions{Text: "DT"}})
				ui.Row(c).Gap(8).Children(func() {
					nui.Badge(c, t, "Private", nui.BadgeOptions{Color: nui.Neutral, Variant: nui.Soft})
					nui.Badge(c, t, "Ready for review", nui.BadgeOptions{Color: nui.Success, Variant: nui.Soft})
				})
				nui.ProgressGroup(c, t, []nui.ProgressSegment{{ID: "design", Label: "Design", Value: 30, Color: nui.Secondary}, {ID: "implementation", Label: "Implementation", Value: 35, Color: nui.Primary}, {ID: "review", Label: "Review", Value: 15, Color: nui.Warning}}, nui.ProgressGroupOptions{Label: "Milestone completion"})
			})
			if !a.HasDocument {
				nui.Empty(c, t, nui.EmptyOptions{Title: "No documents yet", Description: "Create an example document to see the empty state become a useful workspace.", Icon: lucide.FolderOpen(), Actions: func() {
					nui.Button(c, t, "Add example document", nui.ButtonOptions{Leading: lucide.Plus()}).OnClick(func() { a.HasDocument = true })
				}})
			} else {
				nui.Card(c, t, nui.CardOptions{Title: "Documents"}, func() {
					ui.Row(c).Gap(12).AlignItems(ui.Center).Wrap().Children(func() {
						nui.Icon(c, t, lucide.FolderOpen(), nui.IconOptions{})
						nui.Link(c, t, "Foundation component notes", nui.LinkOptions{}).OnClick(func() { a.Message = "Document opened" })
						nui.Button(c, t, "Remove document", nui.ButtonOptions{Color: nui.Error, Variant: nui.Ghost, Size: nui.SM}).OnClick(func() { a.HasDocument = false })
					})
				})
			}
			ui.Text(c, a.Message).TextColor(t.Muted)
		})
	})
}
