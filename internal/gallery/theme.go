package gallery

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func (a *State) paletteName(primary bool) string {
	name, fallback := a.Neutral, "slate"
	if primary {
		name, fallback = a.Primary, "green"
		if name == "black" {
			return name
		}
	}
	if _, ok := nui.PaletteByName(name); !ok {
		return fallback
	}
	return name
}

func paletteLabel(name string) string { return strings.ToUpper(name[:1]) + name[1:] }

func (a *State) themeConfig() nui.AppConfig {
	p := nui.DefaultPalettes()
	primary := a.paletteName(true)
	if primary != "black" {
		p.Primary, _ = nui.PaletteByName(primary)
	}
	p.Neutral, _ = nui.PaletteByName(a.paletteName(false))
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, p), Dark: nui.NewTheme(true, p)}
	if primary == "black" {
		config.Light = config.Light.WithBlackPrimary()
		config.Dark = config.Dark.WithBlackPrimary()
	}
	if a.Dark {
		config.Mode = nui.DarkMode
	}
	if a.System {
		config.Mode = nui.SystemMode
	}
	if a.Compact {
		config.Light.Density, config.Dark.Density = .8, .8
	}
	config.Light.Motion.Reduce, config.Dark.Motion.Reduce = !a.Animations, !a.Animations
	return config
}

func paletteSwatch(c *ui.Context, t nui.Theme, name string) {
	color := t.Highlighted
	if p, ok := nui.PaletteByName(name); ok {
		color = p.Shade(500)
	}
	ui.Box(c).Size(10, 10).Radius(5).Background(color).Shrink(0)
}

func (a *State) paletteItems(primary bool) []nui.MenuItem {
	names, role := nui.NeutralPaletteNames(), "neutral"
	if primary {
		names, role = append([]string{"black"}, nui.PrimaryPaletteNames()...), "primary"
	}
	selected := a.paletteName(primary)
	items := make([]nui.MenuItem, 0, len(names))
	for _, name := range names {
		label := paletteLabel(name)
		items = append(items, nui.MenuItem{
			ID: role + "-" + name, Label: label, Radio: true, Checked: nui.Set(selected == name),
			OnSelect: func() {
				if primary {
					a.Primary = name
				} else {
					a.Neutral = name
				}
			},
			Build: func(c *ui.Context, t nui.Theme) {
				paletteSwatch(c, t, name)
				ui.Text(c, label).SingleLine().Grow(1)
			},
		})
	}
	return items
}

func (a *State) themeMenu(c *ui.Context, t nui.Theme) {
	items := make([]nui.MenuItem, 0, 7)
	for _, primary := range []bool{true, false} {
		label := "Neutral"
		if primary {
			label = "Primary"
		}
		name := a.paletteName(primary)
		items = append(items, nui.MenuItem{
			ID: "theme-" + strings.ToLower(label), Label: label + " palette", Description: paletteLabel(name),
			Children: a.paletteItems(primary),
			Build: func(c *ui.Context, t nui.Theme) {
				ui.Text(c, label).SingleLine().Grow(1)
				paletteSwatch(c, t, name)
				ui.Text(c, paletteLabel(name)).SingleLine().TextColor(t.Muted)
			},
		})
	}
	items = append(items, nui.MenuItem{ID: "theme-divider", Kind: nui.MenuSeparator}, nui.MenuItem{ID: "theme-appearance", Label: "Appearance", Kind: nui.MenuLabel})
	for _, mode := range []struct {
		label        string
		dark, system bool
	}{{"Light", false, false}, {"Dark", true, false}, {"System", false, true}} {
		checked := a.System == mode.system && (mode.system || a.Dark == mode.dark)
		items = append(items, nui.MenuItem{
			ID: "theme-" + strings.ToLower(mode.label), Label: mode.label, Radio: true, Checked: nui.Set(checked),
			OnSelect: func() { a.Dark, a.System = mode.dark, mode.system },
		})
	}
	anchor := nui.Button(c.Key("theme-menu-trigger"), t, "Theme", nui.ButtonOptions{Variant: nui.Outline, Color: nui.Neutral, Leading: lucide.Palette(), Trailing: lucide.ChevronDown()})
	nui.DropdownMenu(c.Key("theme-menu"), t, anchor, &a.ThemeMenu, items, nui.MenuOptions{Label: "Theme settings", Width: nui.Set(float32(240)), MaxHeight: nui.Set(float32(400))})
}

func (a *State) paletteControl(c *ui.Context, t nui.Theme, primary bool) {
	role, menu := "Neutral", &a.NeutralMenu
	if primary {
		role, menu = "Primary", &a.PrimaryMenu
	}
	anchor := nui.Button(c.Key("choose-"+role), t, role+": "+paletteLabel(a.paletteName(primary)), nui.ButtonOptions{
		AccessibleLabel: "Choose " + strings.ToLower(role) + " color", Variant: nui.Outline, Color: nui.Neutral, Trailing: lucide.ChevronDown(),
	})
	nui.DropdownMenu(c.Key("choose-"+role+"-menu"), t, anchor, menu, a.paletteItems(primary), nui.MenuOptions{Label: role + " colors", Width: nui.Set(float32(220)), MaxHeight: nui.Set(float32(400))})
}
