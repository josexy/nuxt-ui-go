package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func (a *State) foundations(c *ui.Context, t nui.Theme) {
	heading(c, t, "Theme & buttons", "Semantic colors, sizes and component overrides. All controls support keyboard navigation.")
	ui.Row(c).Gap(8).Wrap().Children(func() {
		mode := "light"
		if a.Dark {
			mode = "dark"
		}
		if a.System {
			mode = "system"
		}
		nui.Button(c, t, "Mode: "+mode, nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral}).OnClick(func() {
			if a.System {
				a.System = false
				a.Dark = false
			} else if a.Dark {
				a.System = true
			} else {
				a.Dark = true
			}
		})
		a.paletteControl(c, t, true)
		a.paletteControl(c, t, false)
		nui.Button(c, t, "Toggle density", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { a.Compact = !a.Compact })
	})
	section(c, t, "Six variants", func() {
		ui.Row(c).Gap(8).Wrap().Children(func() {
			for _, v := range []struct {
				name  string
				value nui.Variant
			}{{"Solid", nui.Solid}, {"Outline", nui.Outline}, {"Soft", nui.Soft}, {"Subtle", nui.Subtle}, {"Ghost", nui.Ghost}, {"Link", nui.LinkVariant}} {
				nui.Button(c.Key(v.name), t, v.name, nui.ButtonOptions{Variant: v.value}).OnClick(func() { a.Message = v.name + " activated" })
			}
		})
	})
	section(c, t, "Five sizes", func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Wrap().Children(func() {
			for i, name := range []string{"XS", "SM", "MD", "LG", "XL"} {
				nui.Button(c.Key(name), t, name, nui.ButtonOptions{Size: nui.Size(i + 1), Leading: lucide.Plus()})
			}
		})
	})
	section(c, t, "Semantic colors", func() {
		ui.Row(c).Gap(8).Wrap().Children(func() {
			for i, name := range []string{"Primary", "Secondary", "Success", "Info", "Warning", "Error", "Neutral"} {
				nui.Button(c.Key(name), t, name, nui.ButtonOptions{Color: nui.Color(i + 1), Variant: nui.Soft})
			}
		})
	})
	section(c, t, "States and actions", func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Wrap().Children(func() {
			nui.Button(c, t, "Save", nui.ButtonOptions{Leading: lucide.Check()}).OnClick(func() { a.Message = "Saved" })
			nui.Button(c, t, "Unavailable", nui.ButtonOptions{Disabled: true})
			nui.Button(c, t, "Saving changes", nui.ButtonOptions{Loading: true, Leading: lucide.Check()})
			nui.Button(c, t, "", nui.ButtonOptions{Square: true, Leading: lucide.Search(), AccessibleLabel: "Search library"}).OnClick(func() { a.Message = "Search activated" })
			nui.Link(c, t, "View component guide", nui.LinkOptions{}).OnClick(func() { a.Message = "See docs/foundations.md" })
		})
	})
	section(c, t, "Local component defaults and explicit zero values", func() {
		local := t.WithComponent(nui.ComponentButton, nui.ComponentTheme{Color: nui.Secondary, Variant: nui.Soft, UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999)), PaddingX: nui.Set(float32(20))}}})
		nui.Themed(c, local, func(inner nui.Theme) {
			ui.Row(c).Gap(12).Wrap().Children(func() {
				nui.Button(c, inner, "Rounded default", nui.ButtonOptions{})
				nui.Button(c, inner, "Explicit square corners", nui.ButtonOptions{Color: nui.Error, Variant: nui.Outline, UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(0))}}})
			})
		})
	})
	ui.Text(c, a.Message).TextColor(t.Muted)
}

func (a *State) display(c *ui.Context, t nui.Theme) {
	heading(c, t, "Display components", "Reusable identity, status and surface components.")
	ui.Row(c).Gap(8).AlignItems(ui.Center).Wrap().Children(func() {
		nui.Badge(c, t, "Active", nui.BadgeOptions{Color: nui.Success, Variant: nui.Soft, Leading: lucide.CircleCheck()})
		nui.Badge(c, t, "Needs review", nui.BadgeOptions{Color: nui.Warning, Variant: nui.Subtle})
		nui.Badge(c, t, "Archived", nui.BadgeOptions{Color: nui.Neutral, Variant: nui.Outline})
		nui.Badge(c, t, "Small", nui.BadgeOptions{Size: nui.XS})
		nui.Badge(c, t, "Large", nui.BadgeOptions{Size: nui.XL})
	})
	nui.Separator(c, t, nui.SeparatorOptions{Label: "Workspace members", Icon: lucide.UserRound()})
	ui.Row(c).Gap(24).AlignItems(ui.Center).Wrap().Children(func() {
		nui.Avatar(c, t, nui.AvatarOptions{Image: a.AvatarImage, Alt: "Gradient fixture", Size: nui.XL})
		nui.Avatar(c, t, nui.AvatarOptions{Text: "林晓", Color: nui.Secondary, Size: nui.XL})
		nui.Avatar(c, t, nui.AvatarOptions{Icon: lucide.UserRound(), Alt: "Guest", Size: nui.XL})
		nui.AvatarGroup(c, t, []nui.AvatarItem{{ID: "lin", Avatar: nui.AvatarOptions{Text: "林晓"}}, {ID: "ada", Avatar: nui.AvatarOptions{Alt: "Ada Lovelace"}}, {ID: "yang", Avatar: nui.AvatarOptions{Text: "杨明"}}, {ID: "sam", Avatar: nui.AvatarOptions{Alt: "Sam Lee"}}}, nui.AvatarGroupOptions{Max: nui.Set(3), Size: nui.LG})
		nui.Chip(c, t, nui.ChipOptions{Color: nui.Success}, func() { nui.Avatar(c, t, nui.AvatarOptions{Text: "AL", Alt: "Ada online", Size: nui.LG}) })
		nui.Chip(c, t, nui.ChipOptions{Color: nui.Error, Text: "3"}, func() {
			nui.Button(c, t, "Notifications", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { a.Message = "3 unread notifications" })
		})
	})
	nui.Card(c, t, nui.CardOptions{Title: "Component library review", Description: "Shared with the desktop team", Footer: func() {
		nui.Button(c, t, "Open review", nui.ButtonOptions{}).OnClick(func() { a.Message = "Review opened" })
		nui.Kbd(c, t, []string{"CmdOrCtrl", "K"}, nui.KbdOptions{})
	}}, func() {
		nui.User(c, t, nui.UserOptions{Name: "Lin Chen / 林晓", Description: "Design systems engineer", Avatar: nui.AvatarOptions{Image: a.AvatarImage}, OnActivate: func() { a.Message = "Profile opened" }})
		ui.Text(c, "组件应能在真实业务页面中复用。这里使用较长的中英文内容检查换行、头像裁剪、键盘焦点，以及主题切换后的文字和边框表现。 Resize the window to check wrapping.").TextColor(t.Muted)
	})
	ui.Text(c, a.Message).TextColor(t.Muted)
}

func (a *State) feedback(c *ui.Context, t nui.Theme) {
	heading(c, t, "Feedback", "Try loading indicators, progress and notifications. Pause or resume motion at any time.")
	ui.Row(c).Gap(8).Children(func() {
		label := "Start animations"
		if a.Animations {
			label = "Pause animations"
		}
		nui.Button(c, t, label, nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { a.Animations = !a.Animations })
		nui.Button(c, t, "Reset notice", nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { a.Notice = true })
	})
	if a.Notice {
		nui.Alert(c, t, nui.AlertOptions{Title: "Your changes are saved", Description: "The theme is ready to use across your workspace.", Color: nui.Success, Variant: nui.Soft, Icon: lucide.CircleCheck(), OnClose: func() { a.Notice = false }, Actions: func() {
			nui.Button(c, t, "View changes", nui.ButtonOptions{Variant: nui.Outline, Size: nui.SM}).OnClick(func() { a.Message = "Changes opened" })
		}})
	}
	section(c, t, "Progress", func() {
		ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
			ui.Textf(c, "Preparing files · %.0f%%", a.Percent).Grow(1)
			nui.Button(c, t, "Advance progress", nui.ButtonOptions{Variant: nui.Soft, Size: nui.SM}).OnClick(func() {
				a.Percent += 17
				if a.Percent > 100 {
					a.Percent = 0
				}
			})
		})
		nui.Progress(c, t, a.Percent, nui.ProgressOptions{Label: "Preparing files"})
		nui.Progress(c, t, 0, nui.ProgressOptions{Label: "Synchronizing", Indeterminate: true, Size: nui.SM})
		nui.ProgressGroup(c, t, []nui.ProgressSegment{{ID: "done", Label: "Complete", Value: 42, Color: nui.Success}, {ID: "review", Label: "Review", Value: 18, Color: nui.Warning}, {ID: "planned", Label: "Planned", Value: 15, Color: nui.Info}}, nui.ProgressGroupOptions{Label: "Project completion"})
	})
	section(c, t, "Skeleton", func() {
		ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
			nui.Skeleton(c, t, nui.SkeletonOptions{Width: nui.Set(float32(48)), Height: nui.Set(float32(48)), UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999))}}})
			ui.Column(c).Gap(8).Children(func() {
				nui.Skeleton(c, t, nui.SkeletonOptions{Width: nui.Set(float32(240))})
				nui.Skeleton(c, t, nui.SkeletonOptions{Width: nui.Set(float32(180)), Height: nui.Set(float32(12))})
			})
		})
	})
	ui.Text(c, a.Message).TextColor(t.Muted)
}

func (a *State) workspace(c *ui.Context, t nui.Theme) {
	nui.Main(c, t, func() {
		nui.Container(c, t, nui.ContainerOptions{MaxWidth: nui.Set(float32(920)), Padding: nui.Set(float32(0))}, func() {
			heading(c, t, "Workspace overview", "A workspace built with reusable native components.")
			nui.Card(c, t, nui.CardOptions{Title: "NuxtUI Go", Description: "Reusable native components for everyday desktop tools.", Footer: func() {
				nui.Button(c.Key("workspace-save"), t, "Save workspace", nui.ButtonOptions{Leading: lucide.Check()}).OnClick(func() { a.Saved++; a.Message = fmt.Sprintf("Workspace saved %d times", a.Saved) })
				nui.Button(c, t, "Preferences", nui.ButtonOptions{Variant: nui.Outline, Leading: lucide.Settings()}).OnClick(func() { a.Page = 3 })
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
