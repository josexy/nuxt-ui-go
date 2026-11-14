package nui_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

func TestCheckboxHiddenIndicatorKeepsExplicitIcon(t *testing.T) {
	checked, changes := false, 0
	icon := lucide.Heart()
	label, description := "Receive updates", "Only for this workspace"
	size := nui.MD
	width := float32(260)
	name := func(slot nui.Slot, e ui.Element) {
		switch slot {
		case nui.Root:
			e.Label("Choice")
		case nui.Leading:
			e.Label("Choice icon")
		case nui.Label:
			e.Label("Choice label")
		case nui.Description:
			e.Label("Choice description")
		}
	}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c.Key("choice"), th, &checked, nui.CheckboxOptions{Label: label, Description: description, Variant: nui.ChoiceCard, Indicator: nui.IndicatorHidden, CheckedIcon: icon, Size: size, Style: name, OnChange: func(bool) { changes++ }}).Width(width)
		})
	}, 320, 180)
	glyph := bounds(t, tt, "Choice icon")
	caption := bounds(t, tt, "Choice label")
	detail := bounds(t, tt, "Choice description")
	if glyph.W != 16 || glyph.H != 16 || caption.Y != glyph.Y+glyph.H+4 || detail.Y != caption.Y+caption.H+4 {
		t.Fatalf("hidden indicator layout: icon=%+v label=%+v description=%+v", glyph, caption, detail)
	}
	if glyph.X+glyph.W/2 != caption.X+caption.W/2 {
		t.Fatal("icon and text are not centered together")
	}
	parityScreenshot(t, tt, "checkbox-hidden-label-icon")
	click(t, tt, "Choice")
	if !checked || changes != 1 || !tt.Focused("Choice") {
		t.Fatal("hidden indicator lost its toggle or focus behavior")
	}
	size = nui.XL
	tt.Frame()
	if bounds(t, tt, "Choice icon").W != 20 || changes != 1 || !tt.Focused("Choice") {
		t.Fatal("changing size lost the icon, focus or selection")
	}
	tt.Key(0, ui.KeySpace)
	if checked || changes != 2 {
		t.Fatal("hidden control did not toggle exactly once with Space")
	}
	width, label, description = 140, "接收工作区更新通知", "只应用于当前工作区的更新设置"
	tt.SetSize(320, 320)
	tt.SetDirection(ui.RightToLeft)
	tt.SetDark(true)
	tt.SetZoom(1.25)
	tt.Frame()
	root := bounds(t, tt, "Choice")
	for _, name := range []string{"Choice icon", "Choice label", "Choice description"} {
		r := bounds(t, tt, name)
		if r.X < root.X || r.X+r.W > root.X+root.W || r.Y+r.H > root.Y+root.H {
			t.Fatalf("%s escapes the narrow scaled card: %+v outside %+v", name, r, root)
		}
	}
	if !tt.Focused("Choice") || changes != 2 {
		t.Fatal("direction, theme or text scaling lost focus or changed the value")
	}
	parityScreenshot(t, tt, "checkbox-hidden-narrow-rtl")
	label, description = "", ""
	tt.Frame()
	if _, found := tt.Find("Choice icon"); !found {
		t.Fatal("explicit icon disappeared when text was cleared")
	}
	icon = nil
	tt.Frame()
	if _, found := tt.Find("Choice icon"); found {
		t.Fatal("hidden indicator inserted a fallback icon")
	}
	if changes != 2 {
		t.Fatal("presentation-only changes emitted selection events")
	}
}

func TestCheckboxCardFocusDoesNotOutlineBothCardAndIndicator(t *testing.T) {
	checked := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Focus card", Variant: nui.ChoiceCard, Style: func(slot nui.Slot, e ui.Element) {
				if slot == nui.Indicator {
					e.Label("Indicator")
				}
			}}).Width(240)
		})
	}, 280, 100)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Focus card") {
		t.Fatal("card did not receive keyboard focus")
	}
	r := bounds(t, tt, "Indicator")
	im := tt.Image()
	x, y := int(r.X+r.W), int(r.Y+r.H/2)
	if im.RGBAAt(x+1, y) != im.RGBAAt(x+5, y) {
		t.Fatal("card focus also draws an outline around its inner indicator")
	}
	parityScreenshot(t, tt, "checkbox-card-focus")
}

func TestCheckboxCardPreservesPartOverrides(t *testing.T) {
	checked := true
	global, instance := ui.RGB(220, 80, 120), ui.RGB(60, 100, 180)
	var explicit nui.Styles
	var color nui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode}).WithComponent(nui.ComponentCheckbox, nui.ComponentTheme{UI: nui.Styles{nui.Root: {Background: nui.Set(global), PaddingY: nui.Set(float32(22))}}})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Styled card", Variant: nui.ChoiceCard, Color: color, UI: explicit}).Width(240)
		})
	}, 280, 120)
	pixel := func() [4]uint32 {
		t.Helper()
		r := bounds(t, tt, "Styled card")
		rr, g, b, a := tt.Image().At(int(r.X+30), int(r.Y+8)).RGBA()
		return [4]uint32{rr, g, b, a}
	}
	rgba32 := func(c ui.Color) [4]uint32 { r, g, b, a := rgba(c).RGBA(); return [4]uint32{r, g, b, a} }
	if pixel() != rgba32(global) || bounds(t, tt, "Styled card").H != 64 {
		t.Fatalf("card state overwrote the component's root background or padding: color=%v want=%v bounds=%+v", pixel(), rgba32(global), bounds(t, tt, "Styled card"))
	}
	color = nui.Error
	tt.Frame()
	if pixel() == rgba32(global) {
		t.Fatal("explicit Color did not take precedence over global appearance")
	}
	explicit = nui.Styles{nui.Root: {Background: nui.Set(instance)}}
	tt.Frame()
	if pixel() != rgba32(instance) {
		t.Fatal("instance root style did not take precedence over Color")
	}
	checked = false
	tt.Frame()
	if pixel() != rgba32(instance) {
		t.Fatal("changing selection discarded the instance root style")
	}
}

func TestCheckboxCardHighlightFollowsIndicatorPlacement(t *testing.T) {
	checked, highlight := false, false
	placement := nui.IndicatorStart
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Highlighted card", Variant: nui.ChoiceCard, Highlight: highlight, Indicator: placement}).Width(240)
		})
	}, 280, 100)
	border := func() [4]uint32 {
		r := bounds(t, tt, "Highlighted card")
		rr, g, b, a := tt.Image().At(int(r.X+r.W/2), int(r.Y)).RGBA()
		return [4]uint32{rr, g, b, a}
	}
	before := border()
	highlight = true
	tt.Frame()
	if border() != before {
		t.Fatal("visible indicator highlight also recolored the card border")
	}
	placement = nui.IndicatorHidden
	tt.Frame()
	if border() == before {
		t.Fatal("hidden indicator highlight did not recolor the card border")
	}
}

func TestCheckboxCardHoverAndFocusBorder(t *testing.T) {
	checked := false
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{Mode: nui.LightMode})
		ui.Column(c).Padding(20).Children(func() {
			nui.Checkbox(c, th, &checked, nui.CheckboxOptions{Label: "Card", Variant: nui.ChoiceCard}).Width(240)
		})
	}, 280, 100)
	r := bounds(t, tt, "Card")
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	th := nui.Light()
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)); got != rgba(th.BorderAccented) {
		t.Fatalf("hovered card border = %v, want %v", got, rgba(th.BorderAccented))
	}
	tt.Key(0, ui.KeyTab)
	if got := tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y)); got != rgba(th.Primary) {
		t.Fatalf("focused card border = %v, want %v", got, rgba(th.Primary))
	}
}

func TestCheckboxTriCardMixedKeepsUnselectedCardBackground(t *testing.T) {
	value := nui.Unchecked
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.CheckboxTri(c, th, &value, nui.CheckboxTriOptions{Label: "All", Variant: nui.ChoiceCard}).Width(240)
		})
	}, 280, 100)
	r := bounds(t, tt, "All")
	x, y := int(r.X+r.W/2), int(r.Y+8)
	unchecked := tt.Image().RGBAAt(x, y)
	value = nui.Indeterminate
	tt.Frame()
	if tt.Image().RGBAAt(x, y) != unchecked {
		t.Fatal("mixed indicator incorrectly applies the checked card background")
	}
	value = nui.Checked
	tt.Frame()
	if tt.Image().RGBAAt(x, y) == unchecked {
		t.Fatal("checked card did not receive its selected background")
	}
}

func TestChoiceDescriptionFollowsLabelLine(t *testing.T) {
	for _, kind := range []string{"checkbox", "switch", "radio"} {
		t.Run(kind, func(t *testing.T) {
			on, selected := false, "a"
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				name := func(slot nui.Slot, e ui.Element) {
					if slot == nui.Label {
						e.Label("Caption line")
					}
					if slot == nui.Description {
						e.Label("Description line")
					}
				}
				o := nui.ToggleOptions{Label: "Caption", Description: "Description", Style: name}
				switch kind {
				case "checkbox":
					nui.Checkbox(c, th, &on, o)
				case "switch":
					nui.Switch(c, th, &on, o)
				default:
					nui.RadioGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Caption", Description: "Description"}}, nui.RadioGroupOptions{Style: name})
				}
			}, 300, 100)
			label, description := bounds(t, tt, "Caption line"), bounds(t, tt, "Description line")
			if description.Y != label.Y+label.H {
				t.Fatalf("unexpected gap between normal label and description: label=%+v description=%+v", label, description)
			}
		})
	}
}

func TestChoiceIndicatorPartRadiusOverride(t *testing.T) {
	for _, kind := range []string{"radio", "switch"} {
		t.Run(kind, func(t *testing.T) {
			on, selected := true, "a"
			marker := ui.RGB(255, 0, 255)
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				style := nui.Decoration{UI: nui.Styles{nui.Indicator: {Radius: nui.Set(float32(0)), BorderWidth: nui.Set(float32(0)), Background: nui.Set(marker)}}, Style: func(slot nui.Slot, e ui.Element) {
					if slot == nui.Indicator {
						e.Label("Shape")
					}
				}}
				ui.Column(c).Padding(20).Children(func() {
					if kind == "radio" {
						nui.RadioGroup(c, th, &selected, []nui.Option[string]{{Value: "a", Label: "Alpha"}}, nui.RadioGroupOptions{Decoration: style})
					} else {
						nui.Switch(c, th, &on, nui.SwitchOptions{Label: "Sync", Decoration: style})
					}
				})
			}, 300, 100)
			r := bounds(t, tt, "Shape")
			if tt.Image().RGBAAt(int(r.X+1), int(r.Y+1)) != rgba(marker) {
				t.Fatal("default rounding overwrote explicit indicator part styling")
			}
		})
	}
}

func TestCheckboxTriCustomIconsKeepStateAndEvents(t *testing.T) {
	value := nui.Indeterminate
	var custom *ui.SVG
	placement := nui.IndicatorStart
	disabled := false
	changes, states := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(20).Children(func() {
			nui.CheckboxTri(c.Key("tri"), th, &value, nui.CheckboxTriOptions{
				Label: "Select all", CheckedIcon: lucide.Heart(), Indicator: placement, Disabled: disabled, OnChange: func(bool) { changes++ },
				IndeterminateIcon: custom,
				OnStateChange:     func(nui.CheckState) { states++ },
			}).Width(220)
		})
	}, 280, 150)
	minus := slices.Clone(tt.Image().Pix)
	custom = lucide.X()
	tt.Frame()
	if bytes.Equal(minus, tt.Image().Pix) {
		t.Fatal("custom indeterminate icon did not replace the default minus")
	}
	parityScreenshot(t, tt, "checkbox-custom-indeterminate")
	value = nui.Checked
	tt.Frame()
	checked := slices.Clone(tt.Image().Pix)
	custom = nil
	tt.Frame()
	if !bytes.Equal(checked, tt.Image().Pix) {
		t.Fatal("indeterminate icon affected the checked state")
	}
	value = nui.Indeterminate
	placement = nui.IndicatorHidden
	tt.Frame()
	hidden := slices.Clone(tt.Image().Pix)
	custom = lucide.X()
	tt.Frame()
	if !bytes.Equal(hidden, tt.Image().Pix) || changes != 0 || states != 0 {
		t.Fatal("mixed icon replaced the hidden label icon or emitted an event")
	}
	click(t, tt, "Select all")
	tt.Frame()
	tt.Frame()
	if value != nui.Checked || changes != 1 || states != 1 {
		t.Fatalf("mixed activation: state=%v changes=%d states=%d", value, changes, states)
	}
	tt.Key(0, ui.KeySpace)
	if value != nui.Unchecked || changes != 2 || states != 2 {
		t.Fatal("Space did not change the checked state exactly once")
	}
	disabled = true
	tt.Frame()
	click(t, tt, "Select all")
	tt.Key(0, ui.KeySpace)
	if value != nui.Unchecked || changes != 2 || states != 2 {
		t.Fatal("disabled custom-icon checkbox changed state")
	}
}
