package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func comparison(c *ui.Context, children func()) {
	ui.Row(c).FillWidth().MinWidth(0).Wrap().Gap(12).AlignItems(ui.Center).Children(children)
}

func (a *State) foundationPage(id string) *componentPage {
	p := &componentPage{}
	add := func(id, title, description string, build func(*ui.Context, nui.Theme)) {
		p.Scenes = append(p.Scenes, scene(id, title, description, build))
	}
	switch id {
	case "app", "theme":
		add("modes", "Light and dark", "App resolves a window theme once. These previews use local scopes and leave the Gallery theme intact.", func(c *ui.Context, t nui.Theme) {
			for _, dark := range []bool{false, true} {
				local := (nui.AppConfig{Mode: nui.LightMode}).Resolve(dark, c.Preferences())
				if dark {
					local = nui.Dark().FollowPreferences(c.Preferences())
				}
				nui.Themed(c.Key(dark), local, func(inner nui.Theme) {
					nui.Card(c, inner, nui.CardOptions{Title: fmt.Sprintf("Dark: %t", dark)}, func() { nui.Button(c, inner, "Local action", nui.ButtonOptions{}) })
				}).FillWidth()
			}
		})
		add("palettes", "Semantic colors and density", "A blue palette and compact spacing apply only inside this preview.", func(c *ui.Context, t nui.Theme) {
			palettes := nui.DefaultPalettes()
			palettes.Primary, _ = nui.PaletteByName("blue")
			local := nui.NewTheme(t.Dark, palettes).FollowPreferences(c.Preferences())
			local.Density = .8
			nui.Themed(c, local, func(inner nui.Theme) {
				comparison(c, func() {
					for _, color := range semanticColors {
						nui.Button(c.Key(color.Name), inner, color.Name, nui.ButtonOptions{Color: color.Value, Variant: nui.Soft})
					}
				})
			})
		})
		add("defaults", "Component defaults and instance overrides", "The rounded soft default is overridden by an explicit zero radius and outline on the second button.", func(c *ui.Context, t nui.Theme) {
			local := t.WithComponent(nui.ComponentButton, nui.ComponentTheme{Color: nui.Secondary, Variant: nui.Soft, UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999)), PaddingX: nui.Set(float32(20))}}})
			nui.Themed(c, local, func(inner nui.Theme) {
				comparison(c, func() {
					nui.Button(c, inner, "Rounded default", nui.ButtonOptions{})
					nui.Button(c, inner, "Explicit square corners", nui.ButtonOptions{Color: nui.Error, Variant: nui.Outline, UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(0))}}})
				})
			})
		})
		if id == "theme" {
			add("window", "Gallery appearance", "These controls change this window; local previews above remain independent.", func(c *ui.Context, t nui.Theme) {
				comparison(c, func() { a.paletteControl(c, t, true); a.paletteControl(c, t, false) })
				comparison(c, func() {
					nui.Button(c, t, "Toggle density", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { a.Compact = !a.Compact })
					nui.Button(c, t, "Cycle theme mode", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() {
						if a.System {
							a.System = false
							a.Dark = false
						} else if !a.Dark {
							a.Dark = true
						} else {
							a.System = true
						}
					})
					nui.Switch(c, t, &a.Animations, nui.SwitchOptions{Label: "Gallery animations"})
				})
				ui.Textf(c, "Compact: %t · System: %t · Dark: %t", a.Compact, a.System, a.Dark)
			})
			p.Scenes = append([]exampleScene{p.Scenes[len(p.Scenes)-1]}, p.Scenes[:len(p.Scenes)-1]...)
			open, plan := false, "personal"
			add("overlay", "Local theme and overlay", "Open the dialog and its Select. Escape closes each focus scope in order.", func(c *ui.Context, t nui.Theme) {
				local := nui.Dark()
				if t.Dark {
					local = nui.Light()
				}
				local = local.FollowPreferences(c.Preferences())
				nui.Themed(c, local, func(inner nui.Theme) {
					nui.Button(c, inner, "Local dialog", nui.ButtonOptions{}).OnClick(func() { open = true })
					nui.Modal(c, inner, &open, "Local theme dialog", func(mt nui.Theme) {
						nui.Select(c, mt, &plan, plans, nui.SelectOptions{Label: "Local plan"}).FillWidth()
						nui.Button(c, mt, "Close local dialog", nui.ButtonOptions{}).OnClick(func() { open = false })
					})
				}).Padding(20)
			})
			p.Deactivate = append(p.Deactivate, func() { open = false })
		}
	case "container":
		add("default", "Default container", "Content is centered within the theme's maximum width.", func(c *ui.Context, t nui.Theme) {
			nui.Container(c, t, nui.ContainerOptions{}, func() {
				nui.Card(c, t, nui.CardOptions{Title: "Default width"}, func() { ui.Text(c, "A readable centered column that shrinks with its parent.") })
			})
		})
		add("widths", "Maximum widths", "The same content is constrained to 320 and 640 DIP.", func(c *ui.Context, t nui.Theme) {
			for _, width := range []float32{320, 640} {
				nui.Container(c.Key(width), t, nui.ContainerOptions{MaxWidth: nui.Set(width)}, func() {
					nui.Card(c, t, nui.CardOptions{Title: fmt.Sprintf("Max %.0f DIP", width)}, func() { ui.Text(c, "Resize the window to compare the two bounds.") })
				})
			}
		})
		add("padding", "Padding and nesting", "An explicit zero padding wraps an inner container with 24 DIP padding.", func(c *ui.Context, t nui.Theme) {
			nui.Container(c, t, nui.ContainerOptions{Padding: nui.Set(float32(0))}, func() {
				nui.Card(c, t, nui.CardOptions{Title: "Outer container"}, func() {
					nui.Container(c, t, nui.ContainerOptions{Padding: nui.Set(float32(24))}, func() { ui.Text(c, "Inner content / 内层内容") })
				})
			})
		})
	case "main":
		add("body", "Simple body", "Main lets the body grow inside a bounded parent.", func(c *ui.Context, t nui.Theme) {
			ui.Column(c).Height(140).FillWidth().Children(func() {
				nui.Main(c, t, func() {
					ui.Text(c, "Workspace content")
					ui.Text(c, "A body can contain native controls and long text.").TextColor(t.Muted)
				})
			})
		})
		add("shell", "Header, Main and Footer", "The middle region owns scrolling; the header and footer stay visible.", func(c *ui.Context, t nui.Theme) {
			ui.Column(c).Height(260).FillWidth().Children(func() {
				nui.HeaderLayout(c, t, nui.HeaderOptions{Title: "Project"})
				nui.Main(c, t, func() {
					ui.Scroll(c).Fill().Children(func() {
						for i := range 12 {
							ui.Textf(c, "Document section %d", i+1)
						}
					})
				})
				nui.FooterLayout(c, t, nui.FooterOptions{Text: "All changes saved"})
			})
		})
		add("sidebar", "Long content beside a sidebar", "This static side panel belongs to this preview, independent of the Gallery directory.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Height(220).FillWidth().Gap(12).Children(func() {
				ui.Column(c).Width(100).Children(func() { ui.Text(c, "Overview"); ui.Text(c, "Documents") })
				nui.Main(c, t, func() {
					ui.Scroll(c).Fill().Children(func() {
						for i := range 20 {
							ui.Textf(c, "Long body line %d · 原生布局", i+1)
						}
					})
				})
			})
		})
	case "icon":
		add("collection", "Lucide icons", "Bundled vector assets render without network access.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for i, svg := range []*ui.SVG{lucide.Check(), lucide.ChevronDown(), lucide.Search(), lucide.X(), lucide.Save(), lucide.Sun(), lucide.Moon(), lucide.Settings()} {
					nui.Icon(c.Key(i), t, svg, nui.IconOptions{Label: fmt.Sprintf("Lucide icon %d", i+1)})
				}
			})
		})
		add("size-color", "Size and color", "Icon dimensions use DIP and explicit theme colors.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, size := range []float32{16, 24, 32, 48} {
					nui.Icon(c.Key(size), t, lucide.Heart(), nui.IconOptions{Size: nui.Set(size), Color: nui.Set(t.Error), Label: fmt.Sprintf("Heart %.0f", size)})
				}
			})
		})
		add("alignment", "Inline composition", "The icon aligns with a label and with a button's leading part.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Icon(c, t, lucide.Check(), nui.IconOptions{Color: nui.Set(t.Success)})
				ui.Text(c, "Ready to publish")
				nui.Button(c, t, "Save document", nui.ButtonOptions{Leading: lucide.Save()})
			})
		})
	case "link":
		count := 0
		activeResult, customResult := "No navigation", "No navigation"
		add("action", "Local action", "Activate with a pointer or Enter. This demonstration keeps navigation local.", func(c *ui.Context, t nui.Theme) {
			nui.Link(c, t, "Open component guide", nui.LinkOptions{}).OnClick(func() { count++ })
			ui.Textf(c, "Opened %d times", count).Role(ui.RoleStatus)
		})
		add("states", "Active and disabled", "Disabled links cannot change the result.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Link(c, t, "Current page", nui.LinkOptions{Active: true}).OnClick(func() { activeResult = "Current page activated" })
				nui.Link(c, t, "Unavailable link", nui.LinkOptions{Disabled: true}).OnClick(func() { activeResult = "Unexpected activation" })
			})
			ui.Text(c, activeResult)
		})
		add("composition", "Icon and custom label", "Link has no icon option; compose an Icon beside it using native layout.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Icon(c, t, lucide.ArrowRight(), nui.IconOptions{})
				nui.Link(c, t, "View details", nui.LinkOptions{Color: nui.Secondary, Decoration: nui.Decoration{UI: nui.Styles{nui.Label: {Weight: nui.Set(700)}}}}).OnClick(func() { customResult = "Details opened" })
			})
			ui.Text(c, customResult)
		})
	case "badge":
		add("variants", "Variants", "Badge supports Solid, Outline, Soft and Subtle.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, v := range buttonVariants[:4] {
					nui.Badge(c.Key(v.Name), t, v.Name, nui.BadgeOptions{Variant: v.Value})
				}
			})
		})
		add("colors", "Semantic colors", "Status labels use each semantic color.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, v := range semanticColors {
					nui.Badge(c.Key(v.Name), t, v.Name, nui.BadgeOptions{Color: v.Value, Variant: nui.Soft})
				}
			})
		})
		add("sizes-icons", "Sizes, icons and square badges", "Compare XS–XL; the icon-only label has square padding.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, v := range fiveSizes {
					nui.Badge(c.Key(v.Name), t, v.Name, nui.BadgeOptions{Size: v.Value, Leading: lucide.Check(), Trailing: lucide.ChevronDown()})
				}
				nui.Badge(c, t, "3", nui.BadgeOptions{Square: true, Color: nui.Error})
			})
		})
	case "card":
		name := "Desktop team"
		saved := "Not saved"
		add("variants", "Surfaces", "Compare solid, outline, soft and subtle cards.", func(c *ui.Context, t nui.Theme) {
			for _, v := range buttonVariants[:4] {
				nui.Card(c.Key(v.Name), t, nui.CardOptions{Title: v.Name, Variant: v.Value}, func() { ui.Text(c, "A reusable content surface.") })
			}
		})
		add("slots", "Header, description and footer", "The footer can contain controls and keyboard hints.", func(c *ui.Context, t nui.Theme) {
			nui.Card(c, t, nui.CardOptions{Title: "Component review", Description: "Shared with the desktop team", Footer: func() { nui.Kbd(c, t, []string{"CmdOrCtrl", "S"}, nui.KbdOptions{}); ui.Text(c, "Review ready") }}, func() {
				ui.Text(c, "Long bilingual content checks wrapping. 组件可以包含正文、描述和操作。")
			})
		})
		add("form", "Interactive content", "Editing here does not change another scene.", func(c *ui.Context, t nui.Theme) {
			nui.Card(c, t, nui.CardOptions{Title: "Team settings", Footer: func() {
				nui.Button(c, t, "Save card", nui.ButtonOptions{}).OnClick(func() { saved = "Saved: " + name })
			}}, func() { nui.Input(c, t, &name, nui.InputOptions{Label: "Card team name"}).FillWidth() })
			ui.Text(c, saved).Role(ui.RoleStatus)
		})
	case "alert":
		visible, actions := true, 0
		add("colors", "Semantic notifications", "Each alert has a distinct status and icon.", func(c *ui.Context, t nui.Theme) {
			for _, v := range semanticColors {
				nui.Alert(c.Key(v.Name), t, nui.AlertOptions{Title: v.Name, Description: "A local notification", Color: v.Value, Variant: nui.Soft, Icon: lucide.CircleCheck()})
			}
		})
		add("orientation", "Horizontal and vertical", "Descriptions and actions wrap inside the available width.", func(c *ui.Context, t nui.Theme) {
			for _, o := range []nui.Orientation{nui.Horizontal, nui.Vertical} {
				nui.Alert(c.Key(o), t, nui.AlertOptions{Title: "Review changes", Description: "Long descriptions remain readable in a narrow window.", Orientation: nui.Set(o), Variant: nui.Outline, Actions: func() { nui.Badge(c, t, "Ready", nui.BadgeOptions{}) }})
			}
		})
		add("actions", "Actions and dismissal", "Close the notice, then reopen it. Action count persists.", func(c *ui.Context, t nui.Theme) {
			nui.Button(c, t, "Reopen notice", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { visible = true })
			if visible {
				nui.Alert(c, t, nui.AlertOptions{Title: "Your changes are saved", Color: nui.Success, Icon: lucide.Check(), OnClose: func() { visible = false }, Actions: func() {
					nui.Button(c, t, "View changes", nui.ButtonOptions{Size: nui.SM}).OnClick(func() { actions++ })
				}})
			}
			ui.Textf(c, "Viewed %d times", actions)
		})
	case "avatar", "avatar-group":
		if id == "avatar" {
			add("image", "Bitmap image", "A local non-square fixture demonstrates cropping without a remote image.", func(c *ui.Context, t nui.Theme) {
				nui.Avatar(c, t, nui.AvatarOptions{Image: a.avatarImage(), Alt: "Gradient fixture", Size: nui.XL})
			})
			add("fallback", "Unicode initials", "Names fall back to their initials, including Chinese names.", func(c *ui.Context, t nui.Theme) {
				comparison(c, func() {
					for _, name := range []string{"Ada Lovelace", "林晓", "Sam Lee"} {
						nui.Avatar(c.Key(name), t, nui.AvatarOptions{Alt: name})
					}
				})
			})
			add("sizes", "Sizes and icon fallback", "Five sizes and semantic colors can be used without an image.", func(c *ui.Context, t nui.Theme) {
				comparison(c, func() {
					for i, v := range fiveSizes {
						nui.Avatar(c.Key(v.Name), t, nui.AvatarOptions{Icon: lucide.UserRound(), Alt: "Guest " + v.Name, Size: v.Value, Color: semanticColors[i].Value})
					}
				})
			})
		} else {
			items := []nui.AvatarItem{{ID: "lin", Avatar: nui.AvatarOptions{Text: "林晓"}}, {ID: "ada", Avatar: nui.AvatarOptions{Alt: "Ada Lovelace"}}, {ID: "yang", Avatar: nui.AvatarOptions{Text: "杨明"}}, {ID: "sam", Avatar: nui.AvatarOptions{Alt: "Sam Lee"}}}
			add("stack", "Overlapping members", "Stable member IDs preserve order.", func(c *ui.Context, t nui.Theme) { nui.AvatarGroup(c, t, items, nui.AvatarGroupOptions{}) })
			add("overflow", "Maximum and overflow", "Four members with Max=2 show an overflow count.", func(c *ui.Context, t nui.Theme) {
				nui.AvatarGroup(c, t, items, nui.AvatarGroupOptions{Max: nui.Set(2)})
			})
			add("sizes", "Sizes and color", "Compare all five group sizes.", func(c *ui.Context, t nui.Theme) {
				comparison(c, func() {
					for _, v := range fiveSizes {
						nui.AvatarGroup(c.Key(v.Name), t, items, nui.AvatarGroupOptions{Max: nui.Set(3), Size: v.Value, Color: nui.Secondary})
					}
				})
			})
		}
	case "chip":
		hidden := false
		add("counts", "Text and counts", "Chips annotate controls without taking keyboard focus.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, text := range []string{"3", "99+", "New"} {
					nui.Chip(c.Key(text), t, nui.ChipOptions{Text: text}, func() { nui.Button(c, t, "Notifications "+text, nui.ButtonOptions{Variant: nui.Outline}) })
				}
			})
		})
		add("colors", "Colors and sizes", "Compare colors on avatar status chips.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for i, v := range fiveSizes {
					nui.Chip(c.Key(v.Name), t, nui.ChipOptions{Color: semanticColors[i].Value, Size: v.Value}, func() { nui.Avatar(c, t, nui.AvatarOptions{Text: v.Name, Size: nui.XL}) })
				}
			})
		})
		add("visibility", "Dynamic visibility", "Toggle only this avatar's status indicator.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Chip(c, t, nui.ChipOptions{Hidden: hidden, Color: nui.Success}, func() { nui.Avatar(c, t, nui.AvatarOptions{Alt: "Ada online", Size: nui.LG}) })
				nui.Button(c, t, "Toggle status chip", nui.ButtonOptions{}).OnClick(func() { hidden = !hidden })
			})
			ui.Textf(c, "Chip hidden: %t", hidden)
		})
	case "separator":
		add("horizontal", "Horizontal rule", "A separator fills the available width.", func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "First section")
			nui.Separator(c, t, nui.SeparatorOptions{})
			ui.Text(c, "Second section")
		})
		add("vertical", "Vertical rule", "Vertical separators need a bounded height.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				ui.Text(c, "Before")
				nui.Separator(c, t, nui.SeparatorOptions{Orientation: nui.Vertical})
				ui.Text(c, "After")
			})
		})
		add("content", "Labels and custom content", "Text, icons, and custom content can occupy the middle.", func(c *ui.Context, t nui.Theme) {
			nui.Separator(c, t, nui.SeparatorOptions{Label: "Workspace members", Icon: lucide.UserRound()})
			nui.Separator(c, t, nui.SeparatorOptions{Type: nui.SeparatorDashed, Content: func() { nui.Badge(c, t, "OR", nui.BadgeOptions{Variant: nui.Soft}) }})
		})
	case "kbd":
		add("single", "Single keys", "Key labels are formatted for the active platform.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, key := range []string{"Enter", "Escape", "Tab", "Space"} {
					nui.Kbd(c.Key(key), t, []string{key}, nui.KbdOptions{})
				}
			})
		})
		add("shortcuts", "Shortcut combinations", "CmdOrCtrl follows the operating system.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Kbd(c, t, []string{"CmdOrCtrl", "K"}, nui.KbdOptions{})
				nui.Kbd(c, t, []string{"Shift", "ArrowDown"}, nui.KbdOptions{})
			})
		})
		add("appearance", "Sizes and variants", "Compare small outlines with large soft and subtle keys.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for i, v := range fiveSizes {
					nui.Kbd(c.Key(v.Name), t, []string{"Ctrl", v.Name}, nui.KbdOptions{Size: v.Value, Variant: buttonVariants[i%4].Value})
				}
			})
		})
	case "skeleton":
		animate := false
		add("text", "Text placeholders", "Widths approximate a title and two lines of body text.", func(c *ui.Context, t nui.Theme) {
			for i, w := range []float32{240, 280, 180} {
				nui.Skeleton(c.Key(i), t, nui.SkeletonOptions{Width: nui.Set(w), Animate: nui.Set(false)})
			}
		})
		add("card", "Avatar and card placeholder", "A rounded avatar and text compose a static loading card.", func(c *ui.Context, t nui.Theme) {
			nui.Card(c, t, nui.CardOptions{}, func() {
				comparison(c, func() {
					nui.Skeleton(c, t, nui.SkeletonOptions{Width: nui.Set(float32(48)), Height: nui.Set(float32(48)), Animate: nui.Set(false), Decoration: nui.Decoration{UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999))}}}})
					ui.Column(c).Gap(8).Children(func() {
						nui.Skeleton(c, t, nui.SkeletonOptions{Width: nui.Set(float32(180)), Animate: nui.Set(false)})
						nui.Skeleton(c, t, nui.SkeletonOptions{Width: nui.Set(float32(120)), Animate: nui.Set(false)})
					})
				})
			})
		})
		add("motion", "Animation and static mode", "Animation also follows ReduceMotion and the Gallery static flag.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c, t, &animate, nui.SwitchOptions{Label: "Animate skeleton"})
			nui.Skeleton(c, t, nui.SkeletonOptions{Animate: nui.Set(animate)})
			ui.Textf(c, "Animation requested: %t", animate)
		})
	case "progress":
		percent := float64(42)
		add("determinate", "Determinate progress", "Advance by 17; the next step after 100 resets to zero.", func(c *ui.Context, t nui.Theme) {
			nui.Progress(c, t, percent, nui.ProgressOptions{Label: "Preparing files"})
			nui.Button(c, t, "Advance progress", nui.ButtonOptions{}).OnClick(func() {
				percent += 17
				if percent > 100 {
					percent = 0
				}
			})
			ui.Textf(c, "Preparing files · %.0f%%", percent)
		})
		add("indeterminate", "Indeterminate progress", "No numeric percentage is available. Reduced motion keeps the indicator static.", func(c *ui.Context, t nui.Theme) {
			nui.Progress(c, t, 0, nui.ProgressOptions{Label: "Synchronizing", Indeterminate: true, Size: nui.SM})
		})
		add("appearance", "Orientation, sizes and colors", "Vertical and horizontal indicators use independent fixed values.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Progress(c, t, 65, nui.ProgressOptions{Label: "Vertical progress", Orientation: nui.Vertical, Color: nui.Success, Size: nui.LG})
				ui.Column(c).Grow(1).Gap(12).Children(func() {
					for i, v := range fiveSizes {
						nui.Progress(c.Key(v.Name), t, float64(20+i*15), nui.ProgressOptions{Label: v.Name + " progress", Size: v.Value, Color: semanticColors[i].Value})
					}
				})
			})
		})
	case "progress-group":
		segments := []nui.ProgressSegment{{ID: "done", Label: "Complete", Value: 42, Color: nui.Success}, {ID: "review", Label: "Review", Value: 18, Color: nui.Warning}, {ID: "planned", Label: "Planned", Value: 15, Color: nui.Info}}
		add("segments", "Multiple segments", "Values sum to 75 out of 100.", func(c *ui.Context, t nui.Theme) {
			nui.ProgressGroup(c, t, segments, nui.ProgressGroupOptions{Label: "Project completion"})
		})
		add("legend", "Hidden legend", "The visual segmentation remains while labels are hidden.", func(c *ui.Context, t nui.Theme) {
			nui.ProgressGroup(c, t, segments, nui.ProgressGroupOptions{Label: "Compact completion", HideLegend: true, Size: nui.SM})
		})
		add("vertical", "Vertical orientation", "An explicit height bounds the vertical track.", func(c *ui.Context, t nui.Theme) {
			nui.ProgressGroup(c, t, segments, nui.ProgressGroupOptions{Label: "Vertical completion", Orientation: nui.Vertical, Height: 180})
		})
	case "empty":
		hasDocument, created := false, 0
		add("basic", "Simple empty state", "A title, description and icon explain why no content is shown.", func(c *ui.Context, t nui.Theme) {
			nui.Empty(c, t, nui.EmptyOptions{Title: "No results", Description: "Try a different search.", Icon: lucide.Search()})
		})
		add("transition", "Create content", "Add and remove a document to switch between empty and populated content.", func(c *ui.Context, t nui.Theme) {
			if !hasDocument {
				nui.Empty(c, t, nui.EmptyOptions{Title: "No documents yet", Icon: lucide.FolderOpen(), Actions: func() {
					nui.Button(c, t, "Create example document", nui.ButtonOptions{}).OnClick(func() { hasDocument = true; created++ })
				}})
			} else {
				nui.Card(c, t, nui.CardOptions{Title: "Example document"}, func() {
					nui.Button(c, t, "Remove example document", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { hasDocument = false })
				})
			}
			ui.Textf(c, "Created %d documents", created)
		})
		add("loading", "Loading and surfaces", "A loading empty state and a soft empty state communicate different conditions.", func(c *ui.Context, t nui.Theme) {
			nui.Empty(c, t, nui.EmptyOptions{Title: "Loading projects", Loading: true})
			nui.Empty(c, t, nui.EmptyOptions{Title: "No archived projects", Variant: nui.Soft, Size: nui.SM})
		})
	case "user":
		count := 0
		add("profile", "Name, description and avatar", "A bitmap is optional; Unicode initials are also supported.", func(c *ui.Context, t nui.Theme) {
			nui.User(c, t, nui.UserOptions{Name: "Lin Chen / 林晓", Description: "Design systems engineer", Avatar: nui.AvatarOptions{Image: a.avatarImage()}})
		})
		add("orientation", "Horizontal and vertical", "Independent profile presentations use the same read-only fixture.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, vert := range []bool{false, true} {
					nui.User(c.Key(vert), t, nui.UserOptions{Name: "Ada Lovelace", Description: "Desktop team", Avatar: nui.AvatarOptions{Alt: "Ada Lovelace"}, Vertical: vert, Size: nui.LG})
				}
			})
		})
		add("interaction", "Interactive and disabled", "Enter and Space activate a profile. The disabled profile cannot update the count.", func(c *ui.Context, t nui.Theme) {
			nui.User(c, t, nui.UserOptions{Name: "Open profile", Description: "Keyboard accessible", Avatar: nui.AvatarOptions{Text: "AL"}, OnActivate: func() { count++ }})
			nui.User(c, t, nui.UserOptions{Name: "Disabled profile", Disabled: true, OnActivate: func() { count++ }})
			ui.Textf(c, "Profile opened %d times", count)
		})
	}
	return p
}
