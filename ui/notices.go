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
	CloseButton                    Optional[bool]
	Close                          *ButtonOptions
	CloseIcon                      *ui.SVG
	OnSelect                       func()
	OnClose                        func()
	Decoration
}

func Banner(c *ui.Context, t Theme, open *bool, o BannerOptions) ui.Element {
	root := ui.Column(c).FillWidth().MinWidth(0).Role(ui.RoleStatus).Label(o.Title)
	localOpen := ui.Local(root, "open", func() bool { return true })
	closable := valueOr(o.CloseButton, open != nil || o.Close != nil)
	if open == nil {
		open = localOpen
	}
	if !*open {
		root.Hidden(true)
		return ui.Element{}
	}
	color := selected(defaults(t, ComponentBanner, ComponentTheme{Color: Primary}), ComponentTheme{Color: o.Color}).Color
	bg := t.color(color)
	if color == Neutral {
		bg = t.InvertedBackground
	}
	if o.OnSelect != nil {
		root.Role(ui.RoleGroup).Focusable().OnClick(o.OnSelect).FocusRing(false)
		if root.Hovered() {
			bg = bg.Alpha(.9)
		}
		if root.FocusVisible() {
			root.Shadow(0, 0, 0, 3, t.Background.Alpha(.25))
		}
	}
	finish(t, ComponentBanner, Root, root, PartStyle{Background: Set(bg), Foreground: Set(t.Inverted), Font: Set(t.Font), FontSize: Set(t.text(14)), LineHeight: Set(float32(20.0 / 14))}, o.Decoration)
	width, _ := c.Size()
	padding := float32(16)
	if width >= 1024 {
		padding = 32
	} else if width >= 640 {
		padding = 24
	}
	root.Children(func() {
		container := ui.Row(c.Key("container")).FillWidth().MaxWidth(t.ContainerWidth).MinWidth(0).MinHeight(t.spacing(48)).AlignSelf(ui.Center).AlignItems(ui.Center).Justify(ui.SpaceBetween)
		finish(t, ComponentBanner, Body, container, PartStyle{PaddingX: Set(t.spacing(padding)), PaddingY: Set(t.spacing(6)), Gap: Set(t.spacing(12))}, o.Decoration)
		container.Children(func() {
			ui.Box(c.Key("start")).Grow(1).Shrink(1).Basis(0).Hidden(width < 1024)
			center := ui.Row(c.Key("center")).MinWidth(0).Shrink(1).AlignItems(ui.Center)
			finish(t, ComponentBanner, Header, center, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
			center.Children(func() {
				if o.Icon != nil {
					finish(t, ComponentBanner, Leading, ui.Icon(c.Key("icon"), o.Icon).Size(t.text(20), t.text(20)).Shrink(0), PartStyle{Foreground: Set(t.Inverted)}, o.Decoration)
				}
				if o.Title != "" {
					finish(t, ComponentBanner, Label, ui.Text(c.Key("title"), o.Title).SingleLine().MinWidth(0), PartStyle{Foreground: Set(t.Inverted), Weight: Set(500)}, o.Decoration)
				}
				if o.Description != "" {
					finish(t, ComponentBanner, Description, ui.Text(c.Key("description"), o.Description).SingleLine().MinWidth(0), PartStyle{Foreground: Set(t.Inverted.Alpha(.9))}, o.Decoration)
				}
				if o.Actions != nil {
					actions := ui.Row(c.Key("actions")).Shrink(0)
					finish(t, ComponentBanner, Actions, actions, PartStyle{Gap: Set(t.spacing(6))}, o.Decoration)
					// Actions default to Nuxt's neutral XS buttons.
					actionTheme := t.WithComponent(ComponentButton, ComponentTheme{Color: Neutral, Size: XS})
					actions.Children(func() { o.Actions(c, actionTheme) })
				}
			})
			end := ui.Row(c.Key("end")).Shrink(0).Justify(ui.End)
			if width >= 1024 {
				end.Grow(1).Shrink(1).Basis(0)
			}
			finish(t, ComponentBanner, Trailing, end, PartStyle{}, o.Decoration)
			end.Children(func() {
				if closable {
					close := Button(c.Key("close"), t, "", bannerCloseOptions(c, t, o)).OnClick(func() {
						if !*open {
							return
						}
						*open = false
						if o.OnClose != nil {
							o.OnClose()
						}
					})
					if width < 1024 {
						if c.Direction() == ui.RightToLeft {
							close.Margin(0, 0, 0, -t.spacing(6))
						} else {
							close.Margin(0, -t.spacing(6), 0, 0)
						}
					}
					finish(t, ComponentBanner, Close, close, PartStyle{}, o.Decoration)
				}
			})
		})
	})
	return root
}

func bannerCloseOptions(c *ui.Context, t Theme, o BannerOptions) ButtonOptions {
	close := ButtonOptions{}
	if o.Close != nil {
		close = *o.Close
	}
	if close.Size == SizeDefault {
		close.Size = MD
	}
	if close.Color == ColorDefault {
		close.Color = Neutral
	}
	if close.Variant == VariantDefault {
		close.Variant = Ghost
	}
	if close.AccessibleLabel == "" {
		close.AccessibleLabel = o.CloseLabel
		if close.AccessibleLabel == "" {
			close.AccessibleLabel = message(c, "dismissBanner")
		}
	}
	if close.Leading == nil {
		close.Leading = o.CloseIcon
		if close.Leading == nil {
			close.Leading = lucide.X()
		}
	}
	close.Square = true
	if close.Color == Neutral && close.Variant == Ghost && !close.Active {
		close.UI = mergeStyles(Styles{Root: {Foreground: Set(t.Inverted)}}, close.UI)
		custom := close.Style
		if !close.UI[Root].Background.Set {
			close.Style = func(root, label, leading, trailing ui.Element) {
				if !root.IsDisabled() && (root.Hovered() || root.Pressed() || root.FocusVisible()) {
					root.Background(t.Background.Alpha(.1))
				} else {
					root.Background(ui.Transparent)
				}
				if custom != nil {
					custom(root, label, leading, trailing)
				}
			}
		}
	}
	return close
}

type ErrorOptions struct {
	Status, Title, Description, RetryLabel string
	Icon                                   *ui.SVG
	OnRetry                                func()
	Clear                                  Optional[bool]
	Retry                                  *ButtonOptions
	Actions                                func(*ui.Context, Theme)
	Decoration
}

// ErrorView is the Error component; Error is already the library's semantic color.
func ErrorView(c *ui.Context, t Theme, o ErrorOptions) ui.Element {
	root := ui.Column(c).MinWidth(0).AlignItems(ui.Center).Justify(ui.Center).Role(ui.RoleGroup).Label(o.Title)
	finish(t, ComponentError, Root, root, PartStyle{PaddingX: Set(t.spacing(24)), PaddingY: Set(t.spacing(32)), Font: Set(t.Font)}, o.Decoration)
	width, _ := c.Size()
	titleSize, titleLine := float32(36), float32(40.0/36)
	if width >= 640 {
		titleSize, titleLine = 48, 1
	}
	root.Children(func() {
		if o.Icon != nil {
			finish(t, ComponentError, Leading, ui.Icon(c.Key("leading"), o.Icon).Size(t.text(40), t.text(40)).Shrink(0).Margin(0, 0, t.spacing(16), 0), PartStyle{Foreground: Set(t.Primary)}, o.Decoration)
		}
		if o.Status != "" {
			finish(t, ComponentError, StatusSlot, ui.Text(c.Key("status"), o.Status), PartStyle{FontSize: Set(t.text(16)), LineHeight: Set(float32(24.0 / 16)), Weight: Set(600), Foreground: Set(t.Primary)}, o.Decoration)
		}
		if o.Title != "" {
			finish(t, ComponentError, Label, ui.Text(c.Key("title"), o.Title).Role(ui.RoleHeading).Level(2).TextAlign(ui.Center).Margin(t.spacing(8), 0, 0, 0), PartStyle{Foreground: Set(t.Highlighted), FontSize: Set(t.text(titleSize)), LineHeight: Set(titleLine), Weight: Set(700)}, o.Decoration)
		}
		if o.Description != "" && o.Description != o.Title {
			finish(t, ComponentError, Description, ui.Text(c.Key("description"), o.Description).MaxWidth(480).TextAlign(ui.Center).Margin(t.spacing(16), 0, 0, 0), PartStyle{Foreground: Set(t.Muted), FontSize: Set(t.text(18)), LineHeight: Set(float32(28.0 / 18))}, o.Decoration)
		}
		if o.Actions != nil || (o.OnRetry != nil && valueOr(o.Clear, true)) {
			actions := ui.Row(c.Key("actions")).MinWidth(0).Wrap().Justify(ui.Center).Margin(t.spacing(32), 0, 0, 0)
			finish(t, ComponentError, Actions, actions, PartStyle{Gap: Set(t.spacing(24))}, o.Decoration)
			actions.Children(func() {
				if o.Actions != nil {
					o.Actions(c, t)
				} else if o.OnRetry != nil {
					label := o.RetryLabel
					if label == "" {
						label = message(c, "tryAgain")
					}
					opts := ButtonOptions{}
					if o.Retry != nil {
						opts = *o.Retry
					}
					if opts.Color == ColorDefault {
						opts.Color = Primary
					}
					if opts.Variant == VariantDefault {
						opts.Variant = Solid
					}
					if opts.Size == SizeDefault {
						opts.Size = LG
					}
					Button(c.Key("retry"), t, label, opts).OnClick(o.OnRetry)
				}
			})
		}
	})
	return root
}
