package nui_test

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

// Sample away from text and rounded corners, including the outside focus ring.
func fieldPixels(t *testing.T, tt *ui.Tester, label string) [2]color.RGBA {
	t.Helper()
	r := bounds(t, tt, label)
	x, y := int(r.X+r.W/2), int(r.Y)
	return [2]color.RGBA{tt.Image().RGBAAt(x, y), tt.Image().RGBAAt(x, y-2)}
}

func TestTextFieldPointerAndKeyboardFocus(t *testing.T) {
	for _, kind := range []string{"input", "textarea", "number", "date", "time", "menu", "tags"} {
		for _, dark := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dark=%v", kind, dark), func(t *testing.T) {
				text := ""
				var number nui.Optional[float64]
				var numberState nui.NumberState
				var date nui.Optional[nui.Date]
				var dateState nui.DateInputState
				var clock nui.Optional[nui.ClockTime]
				var timeState nui.TimeInputState
				var menu nui.SearchState
				var tags []string
				var tagsState nui.TagsState
				shown, disabled := true, false
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.Light()
					if dark {
						th = nui.Dark()
					}
					nui.Window(c, th)
					ui.Column(c).Padding(24).Gap(16).Children(func() {
						nui.Button(c.Key("before"), th, "Before", nui.ButtonOptions{})
						if shown {
							o := nui.InputOptions{Label: "Editor", Disabled: disabled, Style: func(root, editor, leading, trailing ui.Element) {
								if kind != "tags" {
									root.Label("Field")
								}
							}}
							c.Key("field")
							switch kind {
							case "input":
								nui.Input(c, th, &text, o)
							case "textarea":
								nui.Textarea(c, th, &text, nui.TextareaOptions{InputOptions: o})
							case "number":
								nui.InputNumber(c, th, &number, &numberState, nui.InputNumberOptions{InputOptions: o})
							case "date":
								nui.InputDate(c, th, &date, &dateState, nui.InputDateOptions{InputOptions: o})
							case "time":
								nui.InputTime(c, th, &clock, &timeState, nui.InputTimeOptions{InputOptions: o})
							case "menu":
								nui.InputMenu(c, th, &menu, []nui.Option[string]{{Value: "go", Label: "Go"}}, nui.InputMenuOptions{InputOptions: o})
							case "tags":
								nui.InputTags(c, th, &tags, &tagsState, nui.InputTagsOptions{InputOptions: o}).Label("Field")
							}
						}
						nui.Button(c.Key("after"), th, "After", nui.ButtonOptions{})
					})
				}, 420, 360)
				idle := fieldPixels(t, tt, "Field")
				click(t, tt, "Before")
				tt.Key(0, ui.KeyTab)
				if !tt.Focused("Editor") {
					t.Fatal("Tab did not focus the text editor")
				}
				keyboard := fieldPixels(t, tt, "Field")
				if keyboard[0] == idle[0] || keyboard[1] == idle[1] {
					t.Fatalf("keyboard focus did not highlight border and ring: idle=%v focused=%v", idle, keyboard)
				}
				tt.Key(0, ui.KeyEscape)
				click(t, tt, "After")
				click(t, tt, "Editor")
				tt.Frame()
				if !tt.Focused("Editor") || fieldPixels(t, tt, "Field") != keyboard {
					t.Fatalf("pointer focus should match keyboard focus: pointer=%v keyboard=%v", fieldPixels(t, tt, "Field"), keyboard)
				}
				if dir := os.Getenv("NUI_ARTIFACTS"); dir != "" && (kind == "input" || kind == "tags") {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					f, err := os.Create(filepath.Join(dir, fmt.Sprintf("pointer-focus-%s-dark-%v.png", kind, dark)))
					if err != nil {
						t.Fatal(err)
					}
					encodeErr := png.Encode(f, tt.Image())
					closeErr := f.Close()
					if encodeErr != nil || closeErr != nil {
						t.Fatalf("save focus image: encode=%v close=%v", encodeErr, closeErr)
					}
				}
				disabled = true
				tt.Frame()
				click(t, tt, "Editor")
				if tt.Focused("Editor") {
					t.Fatal("disabled editor retained focus")
				}
				shown = false
				tt.Frame()
				tt.Frame()
				shown, disabled = true, false
				tt.Frame()
				click(t, tt, "Editor")
				if !tt.Focused("Editor") || fieldPixels(t, tt, "Field") != keyboard {
					t.Fatal("remounted editor lost pointer focus styling")
				}
			})
		}
	}
}

func TestInputTagsReadOnlyDraft(t *testing.T) {
	tags := []string{"Go"}
	s := nui.TagsState{Query: "draft"}
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		nui.InputTags(c, th, &tags, &s, nui.InputTagsOptions{
			Label: "Tags", ReadOnly: true,
			OnTagsChange: func([]string) { changes++ },
		})
	}, 400, 160)
	click(t, tt, "Tags")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !slices.Equal(tags, []string{"Go"}) || s.Query != "draft" || changes != 0 {
		t.Fatalf("read-only Enter changed tags or draft: tags=%v query=%q changes=%d", tags, s.Query, changes)
	}
}

func TestSelectionPopupDisabledWhileOpen(t *testing.T) {
	for _, kind := range []string{"select", "input-menu", "select-menu"} {
		t.Run(kind, func(t *testing.T) {
			disabled := false
			value := "go"
			values := []string{"go"}
			search := nui.SearchState{}
			state := nui.SelectMenuState{}
			changes := 0
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(24).Children(func() {
					switch kind {
					case "select":
						nui.Select(c, th, &value, []nui.Option[string]{{Value: "go", Label: "Go"}, {Value: "vue", Label: "Vue"}}, nui.SelectOptions{Label: "Choice", Disabled: disabled, OnChange: func() { changes++ }})
					case "input-menu":
						nui.InputMenu(c, th, &search, []nui.Option[string]{{Value: "go", Label: "Go"}, {Value: "vue", Label: "Vue"}}, nui.InputMenuOptions{Label: "Choice", Disabled: disabled, OnSelect: func(string) { changes++ }})
					case "select-menu":
						nui.SelectMenu(c, th, &values, &state, []nui.ListboxItem{{ID: "go", Label: "Go"}, {ID: "vue", Label: "Vue"}}, nui.SelectMenuOptions{Label: "Choice", Disabled: disabled, OnChange: func([]string) { changes++ }, Clearable: true})
					}
				})
			}, 420, 420)
			click(t, tt, "Choice")
			if kind == "input-menu" {
				tt.Key(0, ui.KeyDown)
			}
			if !tt.HasText("Vue") {
				t.Fatal("popup did not open")
			}
			disabled = true
			tt.Frame()
			if _, ok := tt.Find("Vue"); ok {
				click(t, tt, "Vue")
			}
			tt.Key(0, ui.KeyDown)
			tt.Key(0, ui.KeyEnter)
			if changes != 0 || value != "go" || search.Selected != "" || !slices.Equal(values, []string{"go"}) {
				t.Fatalf("disabled popup remained interactive: changes=%d select=%s input=%s menu=%v", changes, value, search.Selected, values)
			}
		})
	}
}

func TestFieldFocusVariantsAndReadOnly(t *testing.T) {
	for _, tagsField := range []bool{false, true} {
		for _, variant := range []nui.Variant{nui.Outline, nui.Subtle, nui.Soft, nui.Ghost, nui.None} {
			for _, invalid := range []bool{false, true} {
				t.Run(fmt.Sprintf("tags=%v/variant=%d/invalid=%v", tagsField, variant, invalid), func(t *testing.T) {
					value := "Draft"
					tags := []string{"Go"}
					state := nui.TagsState{Query: "Draft"}
					tt := ui.NewTester(func(c *ui.Context) {
						th := nui.App(c, nui.AppConfig{})
						ui.Column(c).Padding(24).Gap(16).Children(func() {
							o := nui.InputOptions{Label: "Editor", ReadOnly: true, Invalid: invalid, Variant: variant}
							if tagsField {
								nui.InputTags(c, th, &tags, &state, nui.InputTagsOptions{InputOptions: o}).Label("Field")
							} else {
								nui.Input(c, th, &value, o).Label("Field")
							}
							nui.Button(c, th, "After", nui.ButtonOptions{})
						})
					}, 420, 240)
					idle := fieldPixels(t, tt, "Field")
					click(t, tt, "Editor")
					tt.Move(410, 230) // Focus must persist without hover.
					if !tt.Focused("Editor") {
						t.Fatal("read-only editor must remain focusable")
					}
					focused := fieldPixels(t, tt, "Field")
					if invalid && focused[0] != idle[0] {
						t.Fatal("focus replaced the error border")
					}
					if (focused[1] == idle[1]) != (variant == nui.None) {
						t.Fatalf("unexpected focus ring for variant %d: idle=%v focused=%v", variant, idle, focused)
					}
					tt.Key(ui.Ctrl, ui.KeyA)
					tt.Command("copy")
					if tt.Clipboard() != "Draft" {
						t.Fatal("read-only field lost selection or copy support")
					}
					tt.Type("blocked")
					tt.Key(0, ui.KeyBackspace)
					if value != "Draft" || state.Query != "Draft" || !slices.Equal(tags, []string{"Go"}) {
						t.Fatal("read-only editing changed the value")
					}
					click(t, tt, "After")
					if fieldPixels(t, tt, "Field") != idle {
						t.Fatal("blur did not restore the idle border and ring")
					}
				})
			}
		}
	}
}

func TestSelectHoverVariants(t *testing.T) {
	for _, variant := range []nui.Variant{nui.Outline, nui.Subtle} {
		t.Run(fmt.Sprint(variant), func(t *testing.T) {
			value := "go"
			disabled := false
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(24).Children(func() {
					nui.Select(c, th, &value, []nui.Option[string]{{Value: "go", Label: "Go"}}, nui.SelectOptions{Label: "Choice", Variant: variant, Disabled: disabled}).Width(300)
				})
			}, 400, 160)
			r := bounds(t, tt, "Choice")
			background := func() color.RGBA { return tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)) }
			idle := background()
			tt.Move(r.X+r.W/2, r.Y+r.H/2)
			if background() == idle {
				t.Fatal("select should highlight its background on hover")
			}
			disabled = true
			tt.Frame()
			hoverDisabled := background()
			tt.Move(390, 150)
			if background() != hoverDisabled {
				t.Fatal("disabled select should ignore hover")
			}
		})
	}
}

func TestToggleErrorFocusIndication(t *testing.T) {
	for _, kind := range []string{"checkbox", "switch"} {
		t.Run(kind, func(t *testing.T) {
			value, messageOnly := false, false
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(24).Gap(16).Children(func() {
					nui.Button(c, th, "Before", nui.ButtonOptions{})
					o := nui.ToggleOptions{Label: "Toggle", Invalid: !messageOnly, Style: func(slot nui.Slot, e ui.Element) {
						if slot == nui.Indicator {
							e.Label("Indicator")
						}
					}}
					if messageOnly {
						o.ErrorMessage = "Required"
					}
					if kind == "checkbox" {
						nui.Checkbox(c, th, &value, o)
					} else {
						nui.Switch(c, th, &value, o)
					}
				})
			}, 400, 180)
			idle := fieldPixels(t, tt, "Indicator")
			click(t, tt, "Before")
			tt.Key(0, ui.KeyTab)
			explicitError := fieldPixels(t, tt, "Indicator")
			if explicitError[1] == idle[1] {
				t.Fatal("keyboard focus ring missing")
			}
			messageOnly = true
			tt.Frame()
			if fieldPixels(t, tt, "Indicator") != explicitError {
				t.Fatal("ErrorMessage must use the same error focus ring as Invalid")
			}
			click(t, tt, "Before")
			click(t, tt, "Toggle")
			if fieldPixels(t, tt, "Indicator")[1] != idle[1] {
				t.Fatal("pointer activation should not add a keyboard ring to a toggle")
			}
		})
	}
}

func TestSelectOpenPopupInheritsDisabledAncestor(t *testing.T) {
	value := "go"
	disabled := false
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(24).Disabled(disabled).Children(func() {
			nui.Select(c, th, &value, []nui.Option[string]{{Value: "go", Label: "Go"}, {Value: "vue", Label: "Vue"}}, nui.SelectOptions{Label: "Choice", OnChange: func() { changes++ }})
		})
	}, 420, 240)
	click(t, tt, "Choice")
	disabled = true
	tt.Frame()
	click(t, tt, "Vue")
	if value != "go" || changes != 0 {
		t.Fatal("an open select escaped its disabled ancestor")
	}
	disabled = false
	tt.Frame()
	click(t, tt, "Vue")
	tt.Frame()
	if value != "vue" || changes != 1 {
		t.Fatalf("re-enabled select did not recover: value=%q changes=%d", value, changes)
	}
}

func TestInputTagsErrorUsesOuterBorder(t *testing.T) {
	tags := []string{"Go"}
	state := nui.TagsState{Query: "Draft"}
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(24).Children(func() {
			nui.InputTags(c, th, &tags, &state, nui.InputTagsOptions{
				Label: "Editor", ErrorMessage: "Check tags", Style: func(root, editor, leading, trailing ui.Element) {
					root.Label("Editor shell")
				}}).Label("Field")
		})
	}, 420, 200)
	click(t, tt, "Editor")
	if fieldPixels(t, tt, "Editor shell")[0] != tt.Image().RGBAAt(0, 0) {
		t.Fatalf("tag editor drew a second border inside the outer error border: inner=%v outer=%v pixels=%v background=%v", bounds(t, tt, "Editor shell"), bounds(t, tt, "Field"), fieldPixels(t, tt, "Editor shell"), tt.Image().RGBAAt(0, 0))
	}
	if fieldPixels(t, tt, "Field")[0] == tt.Image().RGBAAt(0, 0) {
		t.Fatal("outer error border missing")
	}
}
