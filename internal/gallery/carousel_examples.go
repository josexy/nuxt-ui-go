package gallery

import (
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func addCarouselContract(p *componentPage, a *State) {
	var state nui.CarouselState
	var items []nui.CarouselItem
	drafts := map[string]*string{"one": new(string), "two": new(string), "three": new(string)}
	for _, id := range []string{"one", "two", "three"} {
		items = append(items, nui.CarouselItem{ID: id, Label: "Contract slide " + id, Build: func(c *ui.Context, t nui.Theme) {
			ui.Text(c, "Contract slide "+id).FontSize(18)
			nui.Input(c.Key("draft"), t, drafts[id], nui.InputOptions{Label: "Contract " + id + " draft", Placeholder: "Keep a draft while changing layout"})
			nui.Button(c.Key("remove"), t, "Remove contract slide "+id, nui.ButtonOptions{Variant: nui.Ghost}).OnClick(func() { items = slices.DeleteFunc(items, func(it nui.CarouselItem) bool { return it.ID == id }) })
		}})
	}
	original := slices.Clone(items)
	vertical, rtl, disabled, custom, avatar, autoplay := false, false, false, false, false, false
	watch, stop, wrap, arrows, dots := true, true, true, true, true
	changes := 0
	p.Scenes = append(p.Scenes, scene("contract", "Carousel navigation and drafts", "Looping arrows and keys wrap; dragging stays at the first/last slide. Disable dragging without losing keyboard navigation. Custom arrow sizes remain centered. Removing the focused slide returns focus to the viewport.", func(c *ui.Context, t nui.Theme) {
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Checkbox(c, t, &vertical, nui.CheckboxOptions{Label: "Vertical contract carousel"})
			nui.Checkbox(c, t, &rtl, nui.CheckboxOptions{Label: "RTL contract carousel"})
			nui.Checkbox(c, t, &disabled, nui.CheckboxOptions{Label: "Disable contract carousel"})
			nui.Checkbox(c, t, &watch, nui.CheckboxOptions{Label: "Allow contract dragging"})
			nui.Checkbox(c, t, &wrap, nui.CheckboxOptions{Label: "Loop contract navigation"})
			nui.Checkbox(c, t, &arrows, nui.CheckboxOptions{Label: "Show contract arrows"})
			nui.Checkbox(c, t, &dots, nui.CheckboxOptions{Label: "Show contract dots"})
			nui.Checkbox(c, t, &custom, nui.CheckboxOptions{Label: "Customize carousel buttons"})
			nui.Checkbox(c, t, &avatar, nui.CheckboxOptions{Label: "Previous button avatar"})
			nui.Checkbox(c, t, &autoplay, nui.CheckboxOptions{Label: "Autoplay contract carousel"})
			nui.Checkbox(c, t, &stop, nui.CheckboxOptions{Label: "Stop after carousel interaction"})
		})
		ui.Row(c).Wrap().Gap(8).Children(func() {
			nui.Button(c, t, "Restore contract slides", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { items = slices.Clone(original); state.Index = 0 })
			nui.Button(c, t, "Empty contract slides", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { items = nil })
		})
		o := nui.CarouselOptions{Label: "Contract tour", Height: 220, Wrap: wrap, Arrows: nui.Set(arrows), Dots: dots, WatchDrag: nui.Set(watch), StopOnInteraction: nui.Set(stop), Active: nui.Set(!a.WindowInactive && a.Animations), OnChange: func(int) { changes++ },
			UI: nui.Styles{}}
		if vertical {
			o.Orientation = nui.Vertical
			o.Height = 280
		}
		if arrows {
			if vertical {
				o.UI[nui.Item] = nui.PartStyle{PaddingY: nui.Set(float32(64))}
			} else {
				o.UI[nui.Item] = nui.PartStyle{PaddingX: nui.Set(float32(64))}
			}
		}
		if autoplay {
			o.Autoplay = 2 * time.Second
		}
		if custom {
			o.Prev = &nui.ButtonOptions{Size: nui.XS, Color: nui.Info}
			o.Next = &nui.ButtonOptions{Size: nui.XL, Color: nui.Warning}
			o.PrevIcon = lucide.ChevronLeft()
			o.NextIcon = lucide.ChevronRight()
			if vertical {
				o.PrevIcon = lucide.ChevronUp()
				o.NextIcon = lucide.ChevronDown()
			} else if rtl {
				o.PrevIcon = lucide.ChevronRight()
				o.NextIcon = lucide.ChevronLeft()
			}
			o.UI[nui.Indicator] = nui.PartStyle{Radius: nui.Set(float32(2))}
		}
		if avatar {
			o.PrevIcon = nil
			o.Prev = &nui.ButtonOptions{Avatar: &nui.AvatarOptions{Text: "P", Color: nui.Info}, Variant: nui.Outline}
		}
		direction := ui.LeftToRight
		if rtl {
			direction = ui.RightToLeft
		}
		ui.Column(c.Key("direction")).Direction(direction).Disabled(disabled).Children(func() { nui.Carousel(c.Key("carousel"), t, &state, items, o) })
		ui.Textf(c, "Carousel index: %d · Changes: %d · Paused: %t", state.Index, changes, state.Paused)
	}))
}
