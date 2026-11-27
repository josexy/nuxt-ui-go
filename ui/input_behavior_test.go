package nui_test

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
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
				menu.SetOptions([]nui.Option[string]{{Value: "go", Label: "Go"}})
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
								nui.InputMenu(c, th, &menu, nui.InputMenuOptions{InputOptions: o})
							case "tags":
								nui.InputTags(c, th, &tags, &tagsState, nui.InputTagsOptions{InputOptions: o}).Label("Field")
							}
						}
						nui.Button(c.Key("after"), th, "After", nui.ButtonOptions{})
					})
				}, 420, 360)
				editorLabel := "Editor"
				if kind == "date" {
					editorLabel += " month"
				}
				if kind == "time" {
					editorLabel += " hour"
				}
				// Fields fade their colours over Motion.Transition: read them
				// once the fade has arrived.
				settle := func() { tt.Advance(200 * time.Millisecond) }
				idle := fieldPixels(t, tt, "Field")
				click(t, tt, "Before")
				tt.Key(0, ui.KeyTab)
				if !tt.Focused(editorLabel) {
					t.Fatal("Tab did not focus the text editor")
				}
				settle()
				keyboard := fieldPixels(t, tt, "Field")
				if keyboard[0] == idle[0] || keyboard[1] == idle[1] {
					t.Fatalf("keyboard focus did not highlight border and ring: idle=%v focused=%v", idle, keyboard)
				}
				tt.Key(0, ui.KeyEscape)
				click(t, tt, "After")
				click(t, tt, editorLabel)
				tt.Frame()
				settle()
				if !tt.Focused(editorLabel) || fieldPixels(t, tt, "Field") != keyboard {
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
				click(t, tt, editorLabel)
				if tt.Focused(editorLabel) {
					t.Fatal("disabled editor retained focus")
				}
				shown = false
				tt.Frame()
				tt.Frame()
				shown, disabled = true, false
				tt.Frame()
				click(t, tt, editorLabel)
				settle()
				if !tt.Focused(editorLabel) || fieldPixels(t, tt, "Field") != keyboard {
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
			selectState1 := submittedSelect([]nui.Option[string]{{Value: "go", Label: "Go"}, {Value: "vue", Label: "Vue"}})
			listSelection1 := chosenSelection(values)
			state.SetItems([]nui.ListboxItem{{ID: "go", Label: "Go"}, {ID: "vue", Label: "Vue"}})
			search.SetOptions([]nui.Option[string]{{Value: "go", Label: "Go"}, {Value: "vue", Label: "Vue"}})
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(24).Children(func() {
					switch kind {
					case "select":
						nui.Select(c, th, &value, selectState1, nui.SelectOptions{Label: "Choice", Disabled: disabled, OnChange: func() { changes++ }})
					case "input-menu":
						nui.InputMenu(c, th, &search, nui.InputMenuOptions{Label: "Choice", Disabled: disabled, OnSelect: func(string) { changes++ }})
					case "select-menu":
						nui.SelectMenu(c, th, listSelection1, &state, nui.SelectMenuOptions{Label: "Choice", Disabled: disabled, OnChange: func() { changes++ }, Clearable: true})
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
			if changes != 0 || value != "go" || search.Selected != "" || !slices.Equal(listSelection1.Values(), []string{"go"}) {
				t.Fatalf("disabled popup remained interactive: changes=%d select=%s input=%s menu=%v", changes, value, search.Selected, listSelection1.Values())
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
					tt.Advance(200 * time.Millisecond) // the colour transition
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
					tt.Advance(200 * time.Millisecond)
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
			selectState2 := submittedSelect([]nui.Option[string]{{Value: "go", Label: "Go"}})
			tt := ui.NewTester(func(c *ui.Context) {
				th := nui.App(c, nui.AppConfig{})
				ui.Column(c).Padding(24).Children(func() {
					nui.Select(c, th, &value, selectState2, nui.SelectOptions{Label: "Choice", Variant: variant, Disabled: disabled}).Width(300)
				})
			}, 400, 160)
			r := bounds(t, tt, "Choice")
			background := func() color.RGBA { return tt.Image().RGBAAt(int(r.X+r.W/2), int(r.Y+r.H/2)) }
			idle := background()
			tt.Move(r.X+r.W/2, r.Y+r.H/2)
			tt.Advance(200 * time.Millisecond) // the colour transition
			if background() == idle {
				t.Fatal("select should highlight its background on hover")
			}
			disabled = true
			tt.Frame()
			tt.Advance(200 * time.Millisecond)
			hoverDisabled := background()
			tt.Move(390, 150)
			tt.Advance(200 * time.Millisecond)
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
			tt.Advance(250 * time.Millisecond) // a Switch track's ring fades in
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
			tt.Advance(250 * time.Millisecond) // the keyboard ring fading out
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
	selectState3 := submittedSelect([]nui.Option[string]{{Value: "go", Label: "Go"}, {Value: "vue", Label: "Vue"}})
	tt := ui.NewTester(func(c *ui.Context) {
		th := nui.App(c, nui.AppConfig{})
		ui.Column(c).Padding(24).Disabled(disabled).Children(func() {
			nui.Select(c, th, &value, selectState3, nui.SelectOptions{Label: "Choice", OnChange: func() { changes++ }})
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
					root.Label("Field")
				}})
		})
	}, 420, 200)
	click(t, tt, "Editor")
	if p := fieldPixels(t, tt, "Editor")[0]; int(p.R) > 2*int(p.G) && int(p.R) > 2*int(p.B) {
		t.Fatal("tag editor drew a second error border")
	}
	if fieldPixels(t, tt, "Field")[0] == tt.Image().RGBAAt(0, 0) {
		t.Fatal("outer error border missing")
	}
}

func limitedEditor(c *ui.Context, th nui.Theme, value *string, o nui.InputOptions, multiline bool) {
	if multiline {
		nui.Textarea(c.Key("limited"), th, value, nui.TextareaOptions{InputOptions: o, Rows: 2}).Width(300)
	} else {
		nui.Input(c.Key("limited"), th, value, o).Width(300)
	}
}

func TestInputLimitCaretAndSelectedReplacement(t *testing.T) {
	for _, multiline := range []bool{false, true} {
		for _, replace := range []bool{false, true} {
			t.Run(fmt.Sprintf("Multiline=%v/Replace=%v", multiline, replace), func(t *testing.T) {
				value, changes := "abcd", 0
				tt := ui.NewTester(func(c *ui.Context) {
					limitedEditor(c, nui.App(c, nui.AppConfig{DisableAnimations: true}), &value,
						nui.InputOptions{Label: "Limited", MaxLength: 4, OnChange: func(string) { changes++ }}, multiline)
				}, 400, 180)
				tt.Key(0, ui.KeyTab)
				tt.Key(0, ui.KeyHome)
				tt.Key(0, ui.KeyRight)
				if replace {
					tt.Key(ui.Shift, ui.KeyRight)
					tt.Key(ui.Shift, ui.KeyRight)
					tt.Type("XYZ")
					if value != "aXYd" || changes != 1 {
						t.Fatalf("selected replacement: %q, changes=%d", value, changes)
					}
					tt.Key(0, ui.KeyBackspace)
					if value != "aXd" || changes != 2 {
						t.Fatalf("replacement caret: %q, changes=%d", value, changes)
					}
				} else {
					tt.Key(0, ui.KeyRight)
					tt.Type("X")
					if value != "abcd" || changes != 0 {
						t.Fatalf("rejected insertion: %q, changes=%d", value, changes)
					}
					tt.Key(0, ui.KeyBackspace)
					if value != "acd" || changes != 1 {
						t.Fatalf("rejected insertion caret: %q, changes=%d", value, changes)
					}
				}
				before := changes
				tt.Frame()
				tt.Frame()
				if changes != before || !tt.Focused("Limited") {
					t.Fatal("rebuilding changed focus or emitted another edit")
				}
			})
		}
	}
}

func TestInputLimitPasteHistoryAndComposition(t *testing.T) {
	for _, multiline := range []bool{false, true} {
		for _, scenario := range []string{"paste", "rejected-history", "composition", "surrogate"} {
			t.Run(fmt.Sprintf("Multiline=%v/%s", multiline, scenario), func(t *testing.T) {
				value, changes := "", 0
				switch scenario {
				case "paste":
					value = "abcd"
				case "composition":
					value = "a"
				case "surrogate":
					value = "abc"
				}
				tt := ui.NewTester(func(c *ui.Context) {
					limitedEditor(c, nui.App(c, nui.AppConfig{DisableAnimations: true}), &value,
						nui.InputOptions{Label: "Limited", MaxLength: 4, OnChange: func(got string) {
							changes++
							if got != value {
								t.Errorf("callback %q differs from accepted value %q", got, value)
							}
						}}, multiline)
				}, 400, 180)
				tt.Key(0, ui.KeyTab)
				tt.Key(0, ui.KeyEnd)
				switch scenario {
				case "paste":
					tt.Key(0, ui.KeyHome)
					tt.Key(0, ui.KeyRight)
					tt.Key(ui.Shift, ui.KeyRight)
					tt.Key(ui.Shift, ui.KeyRight)
					tt.SetClipboard("😀X")
					tt.Command("paste")
					if value != "a😀d" || changes != 1 {
						t.Fatalf("UTF-16 partial paste: %q, changes=%d", value, changes)
					}
					tt.Command("undo")
					if value != "abcd" || changes != 2 {
						t.Fatalf("paste undo: %q, changes=%d", value, changes)
					}
					tt.Command("redo")
					if value != "a😀d" || changes != 3 {
						t.Fatalf("paste redo: %q, changes=%d", value, changes)
					}
					tt.Key(0, ui.KeyBackspace)
					if value != "ad" || changes != 4 {
						t.Fatalf("redo caret: %q, changes=%d", value, changes)
					}
				case "rejected-history":
					tt.Type("abcd")
					tt.Type("X")
					tt.Command("undo")
					if value != "" || changes != 2 {
						t.Fatalf("rejection added an undo entry: %q, changes=%d", value, changes)
					}
					tt.Command("redo")
					if value != "abcd" || changes != 3 {
						t.Fatalf("accepted insertion redo: %q, changes=%d", value, changes)
					}
				case "composition":
					tt.Compose("😀XY", 4)
					if value != "a" || changes != 0 {
						t.Fatal("composition committed prematurely")
					}
					tt.Type("😀XY")
					if value != "a😀X" || changes != 1 {
						t.Fatalf("native composition limit: %q, changes=%d", value, changes)
					}
				case "surrogate":
					tt.Type("😀")
					if value != "abc" || changes != 0 {
						t.Fatalf("surrogate rejection: %q, changes=%d", value, changes)
					}
					tt.Type("中")
					if value != "abc中" || changes != 1 {
						t.Fatalf("one-unit insertion: %q, changes=%d", value, changes)
					}
				}
			})
		}
	}
}

func TestInputLimitExternalValuesAndConfiguration(t *testing.T) {
	for _, multiline := range []bool{false, true} {
		t.Run(fmt.Sprintf("Multiline=%v", multiline), func(t *testing.T) {
			value, limit, changes := "abcd", 4, 0
			tt := ui.NewTester(func(c *ui.Context) {
				limitedEditor(c, nui.App(c, nui.AppConfig{DisableAnimations: true}), &value,
					nui.InputOptions{Label: "Limited", MaxLength: limit, OnChange: func(string) { changes++ }}, multiline)
			}, 400, 180)
			tt.Key(0, ui.KeyTab)
			value = "abcdef"
			tt.Frame()
			if value != "abcdef" || changes != 0 {
				t.Fatal("external oversized assignment was truncated or notified")
			}
			tt.Key(0, ui.KeyEnd)
			tt.Type("X")
			tt.Key(0, ui.KeyBackspace)
			if value != "abcde" || changes != 1 {
				t.Fatalf("editing an oversized value: %q, changes=%d", value, changes)
			}
			limit = 2
			tt.SetDark(true)
			tt.SetLocale("zh-Hans")
			tt.Frame()
			if value != "abcde" || changes != 1 || !tt.Focused("Limited") {
				t.Fatal("appearance or limit change reset the editor")
			}
			tt.Key(0, ui.KeyBackspace)
			if value != "abcd" || changes != 2 {
				t.Fatal("limit update lost the caret")
			}
			for _, unlimited := range []int{0, -1} {
				limit = unlimited
				tt.Frame()
				before := value
				tt.Type("😀")
				if value != before+"😀" {
					t.Fatalf("limit %d was not unlimited: %q", limit, value)
				}
			}
		})
	}
}

func TestInputLimitDisabledClearAndRemount(t *testing.T) {
	for _, multiline := range []bool{false, true} {
		for _, mode := range []string{"read-only", "disabled", "inherited-disabled"} {
			t.Run(fmt.Sprintf("Multiline=%v/%s", multiline, mode), func(t *testing.T) {
				value, shown, locked := "a", true, false
				changes, clears := 0, 0
				var editor ui.Handle
				tt := ui.NewTester(func(c *ui.Context) {
					th := nui.App(c, nui.AppConfig{DisableAnimations: true})
					ui.Column(c.Key("parent")).Disabled(locked && mode == "inherited-disabled").Children(func() {
						if shown {
							limitedEditor(c, th, &value, nui.InputOptions{
								Label: "Limited", MaxLength: 2, Clearable: true, ClearLabel: "Clear limited",
								ReadOnly: locked && mode == "read-only", Disabled: locked && mode == "disabled",
								OnChange: func(string) { changes++ }, OnClear: func() { clears++ },
								Style: func(_, input, _, _ ui.Element) { input.Bind(&editor) },
							}, multiline)
						}
					})
				}, 400, 180)
				tt.Key(0, ui.KeyTab)
				tt.Key(0, ui.KeyEnd)
				locked = true
				tt.Frame()
				tt.Type("b")
				tt.SetClipboard("c")
				tt.Command("paste")
				tt.Key(0, ui.KeyBackspace)
				if value != "a" || changes != 0 || clears != 0 {
					t.Fatal("locked editor accepted an edit")
				}
				shown = false
				tt.Frame()
				tt.Frame()
				shown, locked = true, false
				editor.Focus()
				tt.Frame()
				if value != "a" || !tt.Focused("Limited") || changes != 0 {
					t.Fatal("remount lost the draft or pending focus")
				}
				click(t, tt, "Clear limited")
				if value != "" || changes != 1 || clears != 1 || !tt.Focused("Limited") {
					t.Fatalf("clear: value=%q changes=%d clears=%d", value, changes, clears)
				}
				tt.Type("😀")
				tt.Type("x")
				if value != "😀" || changes != 2 || clears != 1 {
					t.Fatal("remounted native limit or callbacks changed")
				}
			})
		}
	}
}

func TestAddTagsLargeBatchPreservesOrderAndExistingDuplicates(t *testing.T) {
	previous := []string{"existing", "existing"}
	var batch strings.Builder
	want := slices.Clone(previous)
	for i := range 10000 {
		value := fmt.Sprintf("tag-%d", i)
		batch.WriteString(" " + value + "," + value + ", existing,")
		want = append(want, value)
	}
	if got := nui.AddTags(previous, batch.String(), ",", 0); !slices.Equal(got, want) {
		t.Fatal("batch deduplication changed order or existing duplicate occurrences")
	}
	if got := nui.AddTags(previous, batch.String(), ",", 7); !slices.Equal(got, want[:7]) {
		t.Fatal("batch limit changed accepted order")
	}
	if !slices.Equal(previous, []string{"existing", "existing"}) {
		t.Fatal("batch addition mutated the caller's slice")
	}
}
