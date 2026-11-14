package nui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestThemeOverridesAreIsolatedAndPreserveZero(t *testing.T) {
	patch := Styles{Root: {PaddingX: Set(float32(30)), Radius: Set(float32(20)), Background: Set(ui.Hex("#ff0000")), BorderWidth: Set(float32(2))}}
	base := Light()
	local := base.WithComponent(ComponentButton, ComponentTheme{Color: Secondary, Size: LG, Variant: Outline, UI: patch})
	patch[Root] = PartStyle{Radius: Set(float32(500))}
	exported := local.Component(ComponentButton)
	exported.UI[Root] = PartStyle{}
	if base.Component(ComponentButton).Color != ColorDefault || local.Component(ComponentButton).UI[Root].Radius.Value != 20 {
		t.Fatal("theme maps aliased")
	}
	recipe := func(v ComponentTheme) Styles {
		return Styles{Root: mergeStyle(controlStyle(local, buttonMetrics(local, v.Size)), appearance(local, v.Color, v.Variant, false, false))}
	}
	config, s := resolve(local, ComponentButton, ComponentTheme{Color: Primary, Size: MD, Variant: Solid}, ComponentTheme{Color: Success, Size: XS, Variant: Soft, UI: Styles{Root: {Radius: Set(float32(0)), PaddingX: Set(float32(0)), Background: Set(ui.Transparent)}}}, recipe)
	if config.Color != Success || config.Size != XS || config.Variant != Soft {
		t.Fatal("instance semantics ignored")
	}
	got := s[Root]
	if got.Radius.Value != 0 || got.PaddingX.Value != 0 || got.Background.Value != ui.Transparent || got.BorderWidth.Value != 0 || got.FontSize.Value != 12 {
		t.Fatalf("zero/semantic priority: %+v", got)
	}
	_, s = resolve(local, ComponentButton, ComponentTheme{Color: Primary, Size: MD, Variant: Solid}, ComponentTheme{}, recipe)
	if s[Root].PaddingX.Value != 30 || s[Root].Background.Value != ui.Hex("#ff0000") {
		t.Fatal("global parts ignored")
	}
}
