package nui

import (
	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type BannerOptions struct {
	Title, Description, CloseLabel string
	Color                          Color
	Icon                           *ui.SVG
	Actions                        func(*ui.Context, Theme)
	OnClose                        func()
	Decoration
}

func Banner(c *ui.Context, t Theme, open *bool, o BannerOptions) ui.Element {
	if open != nil && !*open {
		ui.Box(c).Absolute().Size(0, 0).IgnoreInput(true)
		return ui.Element{}
	}
	color := selected(defaults(t, ComponentBanner, ComponentTheme{Color: Primary}), ComponentTheme{Color: o.Color}).Color
	root := ui.Row(c).MinWidth(0).MinHeight(t.spacing(48)).AlignItems(ui.Center).Role(ui.RoleStatus).Label(o.Title)
	finish(t, ComponentBanner, Root, root, PartStyle{Background: Set(t.color(color)), Foreground: Set(t.Inverted), PaddingX: Set(t.spacing(16)), PaddingY: Set(t.spacing(6)), Gap: Set(t.spacing(12))}, o.Decoration)
	root.Children(func() {
		// Nuxt's banner centers its content between equal side regions.
		ui.Box(c.Key("start")).Grow(1).Shrink(1).Basis(0)
		ui.Row(c.Key("center")).MinWidth(0).Shrink(1).AlignItems(ui.Center).Gap(t.spacing(6)).Children(func() {
			if o.Icon != nil {
				finish(t, ComponentBanner, Leading, ui.Icon(c, o.Icon).Size(20, 20).Shrink(0), PartStyle{Foreground: Set(t.Inverted)}, o.Decoration)
			}
			finish(t, ComponentBanner, Label, ui.Text(c, o.Title).SingleLine().MinWidth(0), PartStyle{Foreground: Set(t.Inverted), Weight: Set(500), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, o.Decoration)
			if o.Description != "" {
				finish(t, ComponentBanner, Description, ui.Text(c, o.Description).SingleLine().MinWidth(0), PartStyle{Foreground: Set(t.Inverted.Alpha(.9))}, o.Decoration)
			}
			if o.Actions != nil {
				actions := ui.Row(c).Shrink(0).Gap(t.spacing(6))
				finish(t, ComponentBanner, Actions, actions, PartStyle{}, o.Decoration)
				// Actions default to Nuxt's neutral XS buttons.
				actionTheme := t.WithComponent(ComponentButton, ComponentTheme{Color: Neutral, Size: XS})
				actions.Children(func() { o.Actions(c, actionTheme) })
			}
		})
		ui.Row(c.Key("end")).Grow(1).Shrink(1).Basis(0).Justify(ui.End).Children(func() {
			if open != nil {
				label := o.CloseLabel
				if label == "" {
					label = message(c, "dismissBanner")
				}
				// Nuxt's hover:bg-default/10 keeps the inverted icon readable on the banner color.
				close := Button(c, t, "", ButtonOptions{Leading: lucide.X(), Variant: Ghost, Color: Neutral, Square: true, AccessibleLabel: label, UI: Styles{Root: {Foreground: Set(t.Inverted)}}, Style: func(root, label, leading, trailing ui.Element) {
					if !root.IsDisabled() && (root.Hovered() || root.Pressed() || root.FocusVisible()) {
						root.Background(t.Background.Alpha(.10))
					} else {
						root.Background(ui.Transparent)
					}
				}}).OnClick(func() {
					*open = false
					if o.OnClose != nil {
						o.OnClose()
					}
				})
				finish(t, ComponentBanner, Close, close, PartStyle{}, o.Decoration)
			}
		})
	})
	return root
}

type ErrorOptions struct {
	Status, Title, Description, RetryLabel string
	Icon                                   *ui.SVG
	OnRetry                                func()
	Actions                                func(*ui.Context, Theme)
	Decoration
}

// ErrorView is the Error component; Error is already the library's semantic color.
func ErrorView(c *ui.Context, t Theme, o ErrorOptions) ui.Element {
	root := ui.Column(c).MinWidth(0).AlignItems(ui.Center).Role(ui.RoleGroup).Label(o.Title)
	finish(t, ComponentError, Root, root, PartStyle{PaddingX: Set(t.spacing(24)), PaddingY: Set(t.spacing(32)), Gap: Set(t.spacing(12))}, o.Decoration)
	root.Children(func() {
		if o.Icon != nil {
			finish(t, ComponentError, Leading, ui.Icon(c, o.Icon).Size(36, 36), PartStyle{Foreground: Set(t.Primary)}, o.Decoration)
		}
		if o.Status != "" {
			ui.Text(c, o.Status).FontSize(t.text(16)).LineHeight(24.0 / 16).FontWeight(600).TextColor(t.Primary)
		}
		// Nuxt's text-4xl sm:text-5xl title, at desktop width.
		finish(t, ComponentError, Label, ui.Text(c, o.Title).Role(ui.RoleHeading).Level(2).TextAlign(ui.Center), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(48)), LineHeight: Set(float32(1)), Weight: Set(700)}, o.Decoration)
		if o.Description != "" {
			finish(t, ComponentError, Description, ui.Text(c, o.Description).MaxWidth(480).TextAlign(ui.Center), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(18)), LineHeight: Set(float32(28.0 / 18))}, o.Decoration)
		}
		actions := ui.Row(c).Wrap().Justify(ui.Center).Gap(8).Margin(t.spacing(12), 0, 0, 0)
		finish(t, ComponentError, Actions, actions, PartStyle{}, o.Decoration)
		actions.Children(func() {
			if o.Actions != nil {
				o.Actions(c, t)
			} else if o.OnRetry != nil {
				label := o.RetryLabel
				if label == "" {
					label = message(c, "tryAgain")
				}
				Button(c, t, label, ButtonOptions{Color: Primary, Variant: Solid, Size: LG}).OnClick(o.OnRetry)
			}
		})
	})
	return root
}
