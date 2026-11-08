package gallery

import (
	"fmt"
	"slices"

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
		defaultPadding, draft := true, "A persistent draft"
		add("default", "Default container", "Default horizontal padding follows the window: 16 DIP below 640, 24 below 1024, and 32 above. Explicit zero removes it.", func(c *ui.Context, t nui.Theme) {
			nui.Button(c.Key("padding-mode"), t, "Toggle container padding", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { defaultPadding = !defaultPadding })
			opts := nui.ContainerOptions{}
			if !defaultPadding {
				opts.Padding = nui.Set(float32(0))
			}
			nui.Container(c.Key("container"), t, opts, func() {
				ui.Box(c.Key("boundary")).FillWidth().Height(3).Background(t.Primary).Label("Container content boundary")
				nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Container draft"}).FillWidth()
			}).Label("Responsive container")
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
			ui.Row(c).Height(220).FillWidth().Gap(12).AlignItems(ui.Stretch).Children(func() {
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
				nui.Link(c, t, "View details", nui.LinkOptions{Color: nui.Secondary, UI: nui.Styles{nui.Label: {Weight: nui.Set(700)}}}).OnClick(func() { customResult = "Details opened" })
			})
			ui.Text(c, customResult)
		})
	case "badge":
		badgeIcon := false
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
		add("avatar", "Avatar and icon-only badges", "Leading icons replace the avatar. An empty label automatically uses equal padding on all sides.", func(c *ui.Context, t nui.Theme) {
			nui.Switch(c.Key("icon"), t, &badgeIcon, nui.SwitchOptions{Label: "Override badge avatar with icon"})
			comparison(c, func() {
				var leading *ui.SVG
				if badgeIcon {
					leading = lucide.Check()
				}
				nui.Badge(c.Key("profile"), t, "Online", nui.BadgeOptions{Avatar: &nui.AvatarOptions{Alt: "Ada Lovelace"}, Leading: leading}).Label("Profile badge")
				nui.Badge(c.Key("only"), t, "", nui.BadgeOptions{Leading: lucide.Check()}).Label("Icon-only badge")
			})
		})
	case "card":
		name := "Desktop team"
		saved := "Not saved"
		showTitle := true
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
			nui.Button(c.Key("toggle-title"), t, "Toggle card title", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { showTitle = !showTitle })
			title := ""
			if showTitle {
				title = "Team settings"
			}
			nui.Card(c.Key("card"), t, nui.CardOptions{Title: title, Footer: func() {
				nui.Button(c, t, "Save card", nui.ButtonOptions{}).OnClick(func() { saved = "Saved: " + name })
			}}, func() { nui.Input(c, t, &name, nui.InputOptions{Label: "Card team name"}).FillWidth() })
			ui.Text(c, saved).Role(ui.RoleStatus)
		})
		add("description-only", "Description without a title", "The description keeps its 4 DIP top margin. Card padding uses 16 DIP below 640 and 24 DIP above; header and footer keep 16 DIP vertical padding.", func(c *ui.Context, t nui.Theme) {
			nui.Card(c.Key("description"), t, nui.CardOptions{Description: "Description without title", Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Header {
					e.Label("Description card header")
				}
			}}, nil).Label("Description-only card")
		})
	case "alert":
		visible, actions := true, 0
		withAvatar, horizontal, closeDisabled := true, false, false
		alertDraft, closeRequests := "Persistent alert draft", 0
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
		add("avatar-close", "Avatar and close options", "Avatar takes precedence over Icon. Layout changes retain the action editor; closing reports a request so the application can decide whether to hide the notice.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c.Key("avatar"), t, "Toggle alert avatar", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).KeepFocus().OnClick(func() { withAvatar = !withAvatar })
				nui.Button(c.Key("direction"), t, "Toggle alert direction", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).KeepFocus().OnClick(func() { horizontal = !horizontal })
			})
			nui.Checkbox(c.Key("disabled"), t, &closeDisabled, nui.CheckboxOptions{Label: "Disable alert close"})
			orientation := nui.Vertical
			if horizontal {
				orientation = nui.Horizontal
			}
			var avatar *nui.AvatarOptions
			if withAvatar {
				avatar = &nui.AvatarOptions{Alt: "Ada Lovelace"}
			}
			nui.Alert(c.Key("notice"), t, nui.AlertOptions{Title: "Avatar notification", Description: "Review the draft before continuing. Changing the layout keeps the editor focused.", Variant: nui.Outline, Icon: lucide.Info(), Avatar: avatar, Orientation: nui.Set(orientation), CloseIcon: lucide.ArrowRight(), Close: &nui.ButtonOptions{AccessibleLabel: "Close custom notice", Disabled: closeDisabled}, OnClose: func() { closeRequests++ }, Actions: func() {
				nui.Input(c.Key("draft"), t, &alertDraft, nui.InputOptions{Label: "Alert action draft"}).Width(150)
			}})
			ui.Textf(c, "Alert close requests: %d", closeRequests).Role(ui.RoleStatus)
		})
	case "avatar", "avatar-group":
		if id == "avatar" {
			largeImage, imageAvailable, fallbackIcon, fallbackText, avatarChip := false, true, true, true, true
			add("image", "Bitmap image", "A local non-square fixture demonstrates cropping without a remote image.", func(c *ui.Context, t nui.Theme) {
				nui.Button(c.Key("resize"), t, "Toggle avatar image size", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).OnClick(func() { largeImage = !largeImage })
				nui.Avatar(c.Key("avatar"), t, nui.AvatarOptions{Image: a.avatarImage(), Alt: "Gradient fixture", Size: nui.XL, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Root && largeImage {
						e.Size(64, 64)
					}
					if slot == nui.Body {
						e.Label("Avatar bitmap")
					}
				}})
			})
			add("fallback-priority", "Image and fallback priority", "Image, then icon, literal text and Alt initials. The inset chip is composed around the avatar.", func(c *ui.Context, t nui.Theme) {
				ui.Row(c).Wrap().Gap(8).Children(func() {
					nui.Button(c.Key("image"), t, "Toggle avatar image", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).OnClick(func() { imageAvailable = !imageAvailable })
					nui.Button(c.Key("icon"), t, "Toggle avatar fallback icon", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).OnClick(func() { fallbackIcon = !fallbackIcon })
					nui.Button(c.Key("text"), t, "Toggle avatar fallback text", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).OnClick(func() { fallbackText = !fallbackText })
				})
				nui.Checkbox(c.Key("chip"), t, &avatarChip, nui.CheckboxOptions{Label: "Show avatar chip"})
				opts := nui.AvatarOptions{Alt: "Fallback Avatar", Size: nui.XL}
				if imageAvailable {
					opts.Image = a.avatarImage()
				}
				if fallbackIcon {
					opts.Icon = lucide.UserRound()
				}
				if fallbackText {
					opts.Text = "+12"
				}
				nui.Chip(c.Key("chip-avatar"), t, nui.ChipOptions{Inset: true, Hidden: !avatarChip, Color: nui.Success}, func() { nui.Avatar(c.Key("avatar"), t, opts) }).AlignSelf(ui.Start)
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
			reverseMembers, styleMembers := false, false
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
			add("item-styles", "Member styles and identity", "Item styles apply to visible and overflow avatars. Reordering keeps business IDs; explicit member sizes take precedence over the group.", func(c *ui.Context, t nui.Theme) {
				ui.Row(c).Wrap().Gap(8).Children(func() {
					nui.Button(c.Key("reverse"), t, "Reverse avatar members", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).OnClick(func() { reverseMembers = !reverseMembers })
					nui.Button(c.Key("style"), t, "Toggle avatar member style", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).OnClick(func() { styleMembers = !styleMembers })
				})
				members := []nui.AvatarItem{{ID: "ada", Avatar: nui.AvatarOptions{Alt: "Group Ada", Size: nui.XL}}, {ID: "grace", Avatar: nui.AvatarOptions{Alt: "Group Grace"}}, {ID: "lin", Avatar: nui.AvatarOptions{Alt: "Group Lin"}}}
				if reverseMembers {
					members[0], members[2] = members[2], members[0]
				}
				opts := nui.AvatarGroupOptions{Max: nui.Set(2), Size: nui.SM, Color: nui.Secondary}
				if styleMembers {
					opts.UI = nui.Styles{nui.Item: {Radius: nui.Set(float32(8)), BorderColor: nui.Set(t.Primary), BorderWidth: nui.Set(float32(2))}}
				}
				nui.AvatarGroup(c.Key("members"), t, members, opts)
			})
		}
	case "chip":
		hidden := false
		inset := false
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
		add("positions", "Four physical corners", "Left and right remain physical positions in either reading direction. Inset keeps the marker inside the host.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("inset"), t, &inset, nui.CheckboxOptions{Label: "Inset chip markers"})
			comparison(c, func() {
				for _, pos := range []nui.ChipPosition{nui.TopLeft, nui.TopRight, nui.BottomLeft, nui.BottomRight} {
					nui.Chip(c.Key(pos), t, nui.ChipOptions{Position: pos, Inset: inset}, func() { nui.Avatar(c, t, nui.AvatarOptions{Text: "AL", Size: nui.XL}) })
				}
			})
		})
		add("standalone", "Standalone and styled text", "A nil content callback renders an inline chip. Text width uses the configured Indicator font and padding.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Chip(c.Key("dot"), t, nui.ChipOptions{Color: nui.Success}, nil)
				nui.Chip(c.Key("styled"), t, nui.ChipOptions{Text: "999+", UI: nui.Styles{nui.Indicator: {Font: nui.Set(t.MonoFont), FontSize: nui.Set(float32(14) * t.TextScale), PaddingX: nui.Set(float32(4))}}, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Indicator {
						e.Height(18 * t.TextScale)
					}
				}}, nil).Label("Styled standalone chip")
			})
		})
		add("neutral-token", "Neutral background token", "Neutral uses InvertedBackground. Compare the default token and a local purple override.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				for _, custom := range []bool{false, true} {
					local := t
					label := "Default neutral chip"
					if custom {
						local.InvertedBackground = ui.Hex("#7e22ce")
						label = "Custom neutral chip"
					}
					nui.Chip(c.Key(custom), local, nui.ChipOptions{Color: nui.Neutral, Text: "3", Size: nui.XL}, nil).Label(label)
				}
			})
		})
	case "separator":
		position, vertical := nui.SeparatorCenter, false
		draft := "Persistent separator draft"
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
			nui.Separator(c, t, nui.SeparatorOptions{Avatar: &nui.AvatarOptions{Alt: "Separator member"}})
		})
		add("state", "Content position and focus", "Move the input between start, center and end without replacing its editor. Vertical labels can wrap inside a constrained width.", func(c *ui.Context, t nui.Theme) {
			ui.Row(c).Wrap().Gap(8).Children(func() {
				nui.Button(c.Key("position"), t, "Cycle separator position", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).KeepFocus().OnClick(func() { position = (position + 1) % 3 })
				nui.Button(c.Key("direction"), t, "Toggle separator direction", nui.ButtonOptions{Size: nui.XS, Variant: nui.Outline}).KeepFocus().OnClick(func() { vertical = !vertical })
			})
			part := nui.Separator(c.Key("editor"), t, nui.SeparatorOptions{Position: position, Vertical: vertical, Content: func() {
				nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Separator draft"}).Width(160)
			}})
			if vertical {
				part.Width(180).Height(160)
			}
			nui.Separator(c.Key("wrapped"), t, nui.SeparatorOptions{Orientation: nui.Vertical, Label: "这是一段需要在窄容器中换行的分隔说明文字"}).Size(100, 160)
		})
	case "kbd":
		macKeys, customNeutral := false, false
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
		add("platform", "Explicit platform symbols", "Command and Win display their symbols on every platform. CmdOrCtrl follows the selected platform; this does not register a shortcut.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("platform"), t, &macKeys, nui.CheckboxOptions{Label: "Show macOS key labels"})
			platform := "windows"
			if macKeys {
				platform = "darwin"
			}
			comparison(c, func() {
				nui.Kbd(c.Key("portable"), t, []string{"CmdOrCtrl", "Shift", "P"}, nui.KbdOptions{Platform: platform})
				nui.Kbd(c.Key("symbols"), t, []string{"command", "win"}, nui.KbdOptions{Platform: platform})
			})
		})
		add("neutral-theme", "Shared neutral backgrounds", "Button, Badge, Alert and Kbd use InvertedBackground for Neutral/Solid. The local override leaves other color and variant combinations unchanged.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("custom"), t, &customNeutral, nui.CheckboxOptions{Label: "Custom inverted background"})
			local := t
			if customNeutral {
				local.InvertedBackground = ui.Hex("#7e22ce")
				if local.Dark {
					local.InvertedBackground = ui.Hex("#c084fc")
				}
			}
			comparison(c, func() {
				nui.Button(c.Key("button"), local, "Neutral button", nui.ButtonOptions{Color: nui.Neutral, Variant: nui.Solid})
				nui.Badge(c.Key("badge"), local, "Neutral badge", nui.BadgeOptions{Color: nui.Neutral, Variant: nui.Solid})
				nui.Kbd(c.Key("kbd"), local, []string{"command", "K"}, nui.KbdOptions{Color: nui.Neutral, Variant: nui.Solid})
			})
			nui.Alert(c.Key("alert"), local, nui.AlertOptions{Title: "Neutral alert", Color: nui.Neutral, Variant: nui.Solid})
		})
	case "skeleton":
		animate, square := false, false
		add("text", "Text placeholders", "Widths approximate a title and two lines of body text.", func(c *ui.Context, t nui.Theme) {
			for i, w := range []float32{240, 280, 180} {
				nui.Skeleton(c.Key(i), t, nui.SkeletonOptions{Width: nui.Set(w), Animate: nui.Set(false)})
			}
		})
		add("card", "Avatar and card placeholder", "A rounded avatar and text compose a static loading card.", func(c *ui.Context, t nui.Theme) {
			nui.Card(c, t, nui.CardOptions{}, func() {
				comparison(c, func() {
					nui.Skeleton(c, t, nui.SkeletonOptions{Width: nui.Set(float32(48)), Height: nui.Set(float32(48)), Animate: nui.Set(false), UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999))}}})
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
		add("shape", "Radius overrides", "UI sets a 24 DIP radius; the final Style callback can make the placeholder square. Explicit zero width or height contributes no area in that dimension.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("square"), t, &square, nui.CheckboxOptions{Label: "Square skeleton corners"})
			nui.Skeleton(c.Key("sample"), t, nui.SkeletonOptions{
				Width: nui.Set(float32(160)), Height: nui.Set(float32(48)), Animate: nui.Set(false),
				UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(24))}}, Style: func(_ nui.Slot, e ui.Element) {
					if square {
						e.Radius(0)
					}
				},
			}).Label("Skeleton shape sample")
		})
	case "progress":
		percent := float64(42)
		animateValue := true
		customNeutral, inverted := false, false
		add("determinate", "Determinate progress", "Advance by 17; the next step after 100 resets to zero.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("animate"), t, &animateValue, nui.CheckboxOptions{Label: "Animate progress changes"})
			nui.Progress(c, t, percent, nui.ProgressOptions{Label: "Preparing files", Animate: nui.Set(animateValue)})
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
		add("neutral", "Neutral color and unknown progress", "Neutral uses InvertedBackground. Unknown progress occupies half the track and stays centered when motion is disabled. Value text can describe work units.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("color"), t, &customNeutral, nui.CheckboxOptions{Label: "Custom progress background"})
			nui.Checkbox(c.Key("direction"), t, &inverted, nui.CheckboxOptions{Label: "Invert progress"})
			local := t
			if customNeutral {
				local.InvertedBackground = ui.Hex("#7e22ce")
				if local.Dark {
					local.InvertedBackground = ui.Hex("#c084fc")
				}
			}
			nui.Progress(c.Key("known"), local, 50, nui.ProgressOptions{Label: "Neutral progress", Color: nui.Neutral, Inverted: inverted, Animate: nui.Set(false)}).Value("50 / 100 files")
			nui.Progress(c.Key("unknown"), local, 0, nui.ProgressOptions{Label: "Unknown neutral progress", Color: nui.Neutral, Indeterminate: true, Inverted: inverted, Animate: nui.Set(false)})
		})
	case "progress-group":
		segments := []nui.ProgressSegment{{ID: "done", Label: "Complete", Value: 42, Color: nui.Success}, {ID: "review", Label: "Review", Value: 18, Color: nui.Warning}, {ID: "planned", Label: "Planned", Value: 15, Color: nui.Info}}
		showIcons, customNeutral, groupVertical := true, false, false
		draft := "Keep this draft"
		styled := []nui.ProgressSegment{
			{ID: "system", Label: "System", Value: 30, Color: nui.Neutral, Icon: lucide.Settings()},
			{ID: "apps", Label: "Applications", Value: 20, Color: nui.Info, Icon: lucide.Folder(), UI: nui.Styles{nui.Segment: {Background: nui.Set(ui.Hex("#f97316"))}}},
			{ID: "draft", Value: 10, Color: nui.Warning, Build: func(c *ui.Context, t nui.Theme) {
				nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Segment draft"}).Width(180)
			}},
		}
		add("segments", "Multiple segments", "Values sum to 75 out of 100.", func(c *ui.Context, t nui.Theme) {
			nui.ProgressGroup(c, t, segments, nui.ProgressGroupOptions{Label: "Project completion"})
		})
		add("legend", "Hidden legend", "The visual segmentation remains while labels are hidden.", func(c *ui.Context, t nui.Theme) {
			nui.ProgressGroup(c, t, segments, nui.ProgressGroupOptions{Label: "Compact completion", HideLegend: true, Size: nui.SM})
		})
		add("vertical", "Vertical orientation", "An explicit height bounds the vertical track.", func(c *ui.Context, t nui.Theme) {
			nui.ProgressGroup(c, t, segments, nui.ProgressGroupOptions{Label: "Vertical completion", Orientation: nui.Vertical, Height: 180})
		})
		add("items", "Icons, styles and stable items", "The orange segment has a local fill override. Neutral dots use InvertedBackground; neutral icons use Highlighted. The custom legend editor keeps its draft and focus through reordering and direction changes.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("icons"), t, &showIcons, nui.CheckboxOptions{Label: "Show progress group icons"})
			nui.Checkbox(c.Key("neutral"), t, &customNeutral, nui.CheckboxOptions{Label: "Custom progress group background"})
			comparison(c, func() {
				nui.Button(c.Key("reverse"), t, "Reverse progress segments", nui.ButtonOptions{}).KeepFocus().OnClick(func() { slices.Reverse(styled) })
				nui.Button(c.Key("direction"), t, "Toggle progress group direction", nui.ButtonOptions{}).KeepFocus().OnClick(func() { groupVertical = !groupVertical })
			})
			local := t
			if customNeutral {
				local.InvertedBackground = ui.Hex("#7e22ce")
				if local.Dark {
					local.InvertedBackground = ui.Hex("#c084fc")
				}
			}
			shown := append([]nui.ProgressSegment(nil), styled...)
			if !showIcons {
				for i := range shown {
					shown[i].Icon = nil
				}
			}
			orientation := nui.Horizontal
			if groupVertical {
				orientation = nui.Vertical
			}
			nui.ProgressGroup(c.Key("styled"), local, shown, nui.ProgressGroupOptions{Label: "Styled progress group", Orientation: orientation, Height: 160})
		})
	case "empty":
		hasDocument, created := false, 0
		showTitle, showAvatar, customHeader, customBody, loading := false, true, false, false, false
		draft, actions := "Keep this draft", 0
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
		add("composition", "Avatar, content and stable actions", "A custom header replaces its title and avatar; a custom body replaces default actions. Header changes preserve the action editor. Loading keeps an existing avatar image visible.", func(c *ui.Context, t nui.Theme) {
			nui.Checkbox(c.Key("avatar"), t, &showAvatar, nui.CheckboxOptions{Label: "Show empty avatar"})
			nui.Checkbox(c.Key("loading"), t, &loading, nui.CheckboxOptions{Label: "Load empty content"})
			nui.Checkbox(c.Key("header"), t, &customHeader, nui.CheckboxOptions{Label: "Custom empty header"})
			nui.Checkbox(c.Key("body"), t, &customBody, nui.CheckboxOptions{Label: "Custom empty body"})
			nui.Button(c.Key("title"), t, "Toggle empty title", nui.ButtonOptions{}).KeepFocus().OnClick(func() { showTitle = !showTitle })
			opts := nui.EmptyOptions{Icon: lucide.Folder(), Loading: loading, Actions: func() {
				nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Empty action draft"}).Width(180)
				nui.Button(c.Key("create"), t, "Create from empty action", nui.ButtonOptions{}).OnClick(func() { actions++ })
			}, Footer: func() { ui.Textf(c, "Empty actions: %d", actions) }, UI: nui.Styles{nui.Leading: {Foreground: nui.Set(t.Error)}}}
			if showTitle {
				opts.Title = "这是一段用于检查窄窗口换行和操作区状态保留的空状态说明"
			}
			if showAvatar {
				opts.Avatar = &nui.AvatarOptions{Image: a.avatarImage(), Alt: "Empty example avatar"}
			}
			if customHeader {
				opts.Header = func() { ui.Text(c, "Custom empty header content") }
			}
			if customBody {
				opts.Body = func() {
					nui.Button(c.Key("custom"), t, "Create from custom body", nui.ButtonOptions{}).OnClick(func() { actions++ })
				}
			}
			nui.Empty(c.Key("composed"), t, opts)
		})
	case "user":
		count := 0
		showAvatar, vertical := false, false
		draft := "Keep this draft"
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
		add("chip-content", "Avatar chip and stable content", "The chip uses the User size and stays inset. Changing the avatar or direction preserves the custom editor. A description can be shown without a name.", func(c *ui.Context, t nui.Theme) {
			comparison(c, func() {
				nui.Button(c.Key("avatar"), t, "Toggle user avatar", nui.ButtonOptions{}).KeepFocus().OnClick(func() { showAvatar = !showAvatar })
				nui.Button(c.Key("orientation"), t, "Toggle user direction", nui.ButtonOptions{}).KeepFocus().OnClick(func() { vertical = !vertical })
			})
			opts := nui.UserOptions{Chip: &nui.ChipOptions{Color: nui.Success, Text: "5"}, Vertical: vertical, Body: func() {
				nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "User draft"}).Width(180)
			}}
			if showAvatar {
				opts.Avatar = nui.AvatarOptions{Text: "AU", Alt: "User state avatar"}
			}
			nui.User(c.Key("state"), t, opts).Width(300)
			ui.Textf(c, "User draft: %s", draft)
			nui.User(c.Key("description-only"), t, nui.UserOptions{Description: "Description without a name", Size: nui.XS})
		})
	}
	return p
}
