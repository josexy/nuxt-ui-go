package gallery

import (
	"slices"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

func addLayoutContract(p *componentPage, id string) {
	if id == "footer-columns" {
		addFooterColumnsContract(p)
		return
	}
	draft := ""
	var sidebar nui.SidebarState
	var opener ui.Handle
	collapsed, page := false, "overview"
	leading, styled, rtl, longText := false, false, false, false
	actions := 0
	p.Scenes = append(p.Scenes, scene("contract", "Layout parts and persistent input", "Toggle leading content while editing the header draft. Action regions expose their own spacing and background. Compare RTL and long Chinese text in a narrow composition.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			if id == "header" {
				nui.Button(c, t, "Toggle header leading", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { leading = !leading })
			}
			nui.Checkbox(c, t, &styled, nui.CheckboxOptions{Label: "Style layout parts"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL layout parts"})
			nui.Checkbox(c, t, &longText, nui.CheckboxOptions{Label: "Long layout text"})
		})
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("layout")).Direction(direction).MinWidth(0).MaxWidth(520).Border(1, t.Border).Children(func() {
			label := "Native workspace"
			if longText {
				label = "本地工作区的组件使用说明与待处理事项，较长的文字可以在可用空间内换行。"
			}
			styles := nui.Styles{}
			if styled {
				styles[nui.Actions] = nui.PartStyle{PaddingX: nui.Set(float32(12)), PaddingY: nui.Set(float32(8)), Background: nui.Set(t.Elevated), Radius: nui.Set(float32(6))}
				styles[nui.Leading] = nui.PartStyle{Background: nui.Set(t.Accented), PaddingX: nui.Set(float32(8))}
			}
			if id == "header" {
				o := nui.HeaderOptions{Title: label, Description: "Stable action editor", Decoration: nui.Decoration{UI: styles}, Actions: func(c *ui.Context, t nui.Theme) {
					nui.Input(c.Key("draft"), t, &draft, nui.InputOptions{Label: "Layout header draft"}).Width(150)
					nui.Button(c.Key("menu"), t, "", nui.ButtonOptions{Leading: lucide.Menu(), AccessibleLabel: "Open composed header menu", Variant: nui.Ghost, Color: nui.Neutral, Square: true}).Bind(&opener).OnClick(func() { sidebar.Open = true })
				}}
				if leading {
					o.Leading = func(c *ui.Context, t nui.Theme) { ui.Icon(c, lucide.Folder()).Size(24, 24) }
				}
				nui.HeaderLayout(c.Key("header"), t, o)
				ui.Row(c.Key("content")).Height(160).MinWidth(0).Children(func() {
					nui.Sidebar(c.Key("menu-sidebar"), t, &collapsed, nui.SidebarOptions{Title: "Header navigation", Description: "Composed with Sidebar", State: &sidebar, Toggle: nui.Set(false), Breakpoint: nui.Set(float32(2000)), NavigationKey: page, ReturnFocus: &opener}, func(c *ui.Context, t nui.Theme, compact bool) {
						for _, target := range []string{"overview", "settings"} {
							nui.Button(c.Key(target), t, "Header menu "+target, nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { page = target })
						}
					})
					ui.Text(c, "Header page: "+page).Padding(16)
				})
			} else {
				nui.FooterLayout(c.Key("footer"), t, nui.FooterOptions{Text: label, Decoration: nui.Decoration{UI: styles}, Actions: func(c *ui.Context, t nui.Theme) {
					nui.Button(c, t, "Contract footer action", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { actions++ })
				}})
			}
		})
		if id == "header" {
			ui.Text(c, "Header draft: "+draft)
		} else {
			ui.Textf(c, "Layout actions: %d", actions)
		}
	}))
}

func addFooterColumnsContract(p *componentPage) {
	active := "docs"
	cancel, reverse, styled, rtl, narrow := false, false, false, false, false
	selected, navigated := 0, 0
	p.Scenes = append(p.Scenes, scene("contract", "Footer link states and order", "Active links use the primary color. Per-link OnSelect can cancel navigation; disabled links stay inactive. Reverse the columns to check stable identities, and compare item styling and narrow wrapping.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Checkbox(c, t, &cancel, nui.CheckboxOptions{Label: "Cancel footer navigation"})
			nui.Button(c, t, "Reverse footer columns", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { reverse = !reverse })
			nui.Checkbox(c, t, &styled, nui.CheckboxOptions{Label: "Style footer links"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL footer columns"})
			nui.Checkbox(c, t, &narrow, nui.CheckboxOptions{Label: "Narrow footer columns"})
		})
		selectLink := func(e *nui.NavigationEvent) {
			selected++
			if cancel {
				e.PreventDefault()
			}
		}
		columns := []nui.FooterColumn{
			{ID: "product", Title: "Product", Items: []nui.NavigationItem{{ID: "docs", Label: "Contract documentation", Icon: lucide.BookOpen(), Active: active == "docs", OnSelect: selectLink}, {ID: "examples", Label: "Contract examples", Icon: lucide.Folder(), Active: active == "examples", OnSelect: selectLink}}},
			{ID: "resources", Title: "Resources", Items: []nui.NavigationItem{{ID: "guide", Label: "组件使用说明与完整导航链接", Active: active == "guide", OnSelect: selectLink}, {ID: "disabled", Label: "Contract unavailable", Disabled: true, OnSelect: selectLink}}},
		}
		if reverse {
			slices.Reverse(columns)
		}
		var styles nui.Styles
		if styled {
			styles = nui.Styles{nui.Body: {Background: nui.Set(t.Elevated), PaddingX: nui.Set(float32(12)), PaddingY: nui.Set(float32(12)), Radius: nui.Set(float32(8))}, nui.Item: {Gap: nui.Set(float32(10))}}
			for i := range columns {
				for j := range columns[i].Items {
					if columns[i].Items[j].ID == "examples" {
						columns[i].Items[j].UI = nui.Styles{nui.ItemLabelSlot: {Foreground: nui.Set(t.Error)}}
					}
				}
			}
		}
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		box := ui.Column(c.Key("footer-layout")).Direction(direction).MinWidth(0)
		if narrow {
			box.MaxWidth(140)
		}
		box.Children(func() {
			nui.FooterColumns(c.Key("columns"), t, columns, nui.FooterColumnsOptions{MinColumnWidth: 160, Decoration: nui.Decoration{UI: styles}, OnNavigate: func(id string) { active = id; navigated++ }})
		})
		ui.Textf(c, "Footer active: %s · Selects: %d · Navigations: %d", active, selected, navigated)
	}))
}
