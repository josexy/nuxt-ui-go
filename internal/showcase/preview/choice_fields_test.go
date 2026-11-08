package preview

import (
	"bytes"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

func TestCheckboxIconControlsPreserveSelectionAndCallbackCounts(t *testing.T) {
	tester, cfg := exampleTester(t, "checkbox")
	before := slices.Clone(tester.Image().Pix)
	cfg.Values["customIcons"] = true
	tester.Frame()
	if bytes.Equal(before, tester.Image().Pix) || !tester.HasText("Changes: 0 / 0") {
		t.Fatal("custom mixed icon did not change the preview independently of selection")
	}
	cfg.Values["variant"], cfg.Values["indicator"] = "card", "hidden"
	tester.Frame()
	clickExample(t, tester, cfg.String("label"))
	clickExample(t, tester, "Mixed selection")
	if !tester.HasText("Enabled: true") || !tester.HasText("Changes: 1 / 1") {
		t.Fatal("hidden indicator selection was not reported exactly once")
	}
	cfg.Values["customIcons"] = false
	tester.Frame()
	if !tester.Focused("Mixed selection") || !tester.HasText("Changes: 1 / 1") {
		t.Fatal("icon update reset focus or emitted selection callbacks")
	}
	tester.Key(0, ui.KeySpace)
	if !tester.HasText("Changes: 1 / 2") {
		t.Fatal("Space did not change the mixed control once")
	}
	clickExample(t, tester, "Reset mixed state")
	if !tester.HasText("Changes: 1 / 2") {
		t.Fatal("external state reset emitted a user callback")
	}
	demo, _ := showcase.Find("checkbox")
	if _, _, err := demo.Source(*cfg); err != nil {
		t.Fatalf("configured preview cannot produce its standalone source: %v", err)
	}
}

func TestSharedChoiceFieldsMatchTheirDocumentedEffects(t *testing.T) {
	for _, id := range []string{"checkbox", "switch"} {
		t.Run(id+"/Loading", func(t *testing.T) {
			tester, cfg := exampleTester(t, id)
			cfg.Options = map[string]any{"ToggleOptions": map[string]any{"Loading": true}}
			tester.Frame()
			clickExample(t, tester, cfg.String("label"))
			want := "Enabled: true"
			if id == "switch" {
				want = "Enabled: false"
			}
			if !tester.HasText(want) {
				t.Fatalf("Loading does not match its component description: %v", tester.Texts())
			}
			if id == "switch" {
				cfg.Options["ToggleOptions"] = map[string]any{"Loading": false}
				tester.Frame()
				clickExample(t, tester, cfg.String("label"))
				if !tester.HasText("Enabled: true") {
					t.Fatal("explicit false did not restore switching")
				}
			}
		})
	}
	for _, id := range []string{"checkbox-group", "radio-group"} {
		t.Run(id+"/Max", func(t *testing.T) {
			tester, cfg := exampleTester(t, id)
			cfg.Options = map[string]any{"ChoiceGroupOptions": map[string]any{"Max": float64(0)}}
			tester.Frame()
			clickExample(t, tester, "Go")
			want := "Selected: []"
			if id == "radio-group" {
				want = "Selected: go"
			}
			if !tester.HasText(want) {
				t.Fatalf("Max does not match its component description: %v", tester.Texts())
			}
		})
	}
}

func TestSwitchPreviewLoadingAndIconsKeepValueAndEvents(t *testing.T) {
	tester, cfg := exampleTester(t, "switch")
	before := slices.Clone(tester.Image().Pix)
	cfg.Values["customIcons"] = true
	tester.Frame()
	if bytes.Equal(before, tester.Image().Pix) || !tester.HasText("Changes: 0") {
		t.Fatal("unchecked icon did not update independently of the value")
	}
	clickExample(t, tester, cfg.String("label"))
	if !tester.HasText("Enabled: true") || !tester.HasText("Changes: 1") {
		t.Fatal("Switch did not report the selection once")
	}
	cfg.Values["loading"] = true
	tester.Frame()
	clickExample(t, tester, cfg.String("label"))
	r, ok := tester.Find("Sync policy")
	if !ok {
		t.Fatal("missing field label")
	}
	tester.ClickAt(r.X+3, r.Y+5)
	tester.Key(0, ui.KeySpace)
	if !tester.HasText("Enabled: true") || !tester.HasText("Changes: 1") {
		t.Fatal("loading accepted a switch or label activation")
	}
	cfg.Values["loading"], cfg.Values["highlight"] = false, true
	tester.Frame()
	clickExample(t, tester, cfg.String("label"))
	if !tester.HasText("Enabled: false") || !tester.HasText("Changes: 2") {
		t.Fatal("completed loading did not restore switching exactly once")
	}
}

func TestChoiceGroupPreviewAllowsConfiguredCollections(t *testing.T) {
	for _, id := range []string{"checkbox-group", "radio-group"} {
		t.Run(id, func(t *testing.T) {
			tester, cfg := exampleTester(t, id)
			binding := "CheckboxGroupItems"
			if id == "radio-group" {
				binding = "RadioGroupItems"
			}
			cfg.Options = map[string]any{binding: []any{}}
			tester.Frame()
			clickExample(t, tester, "Reverse choices")
			cfg.Options[binding] = []any{map[string]any{"Value": "one", "Label": "One", "Icon": "heart"}}
			cfg.Values["variant"], cfg.Values["indicator"] = "card", "hidden"
			tester.Frame()
			withPreset := slices.Clone(tester.Image().Pix)
			cfg.Values["customIcons"] = false
			tester.Frame()
			if !bytes.Equal(withPreset, tester.Image().Pix) {
				t.Fatal("preset icon switch replaced the explicitly configured item icon")
			}
			clickExample(t, tester, "One")
			if !tester.HasText("Changes: 1") {
				t.Fatal("configured item did not update selection exactly once")
			}
			clickExample(t, tester, "Reverse choices")
			if !tester.HasText("Changes: 1") {
				t.Fatal("reordering emitted a selection change")
			}
		})
	}
}

func TestChoiceGroupVerticalCardsKeepActionsAndResultVisible(t *testing.T) {
	for _, id := range []string{"checkbox-group", "radio-group"} {
		t.Run(id, func(t *testing.T) {
			tester, cfg := exampleTester(t, id)
			demo, _ := showcase.Find(id)
			cfg.Values["variant"], cfg.Values["indicator"], cfg.Values["horizontal"] = "card", "hidden", false
			for _, size := range []string{"xs", "sm", "md", "lg", "xl"} {
				cfg.Values["size"] = size
				tester.Frame()
				for _, label := range []string{"Reverse choices", "Changes: 0"} {
					r, found := tester.Find(label)
					if !found || r.H < 16 || r.Y+r.H > float32(demo.Height)-24 {
						t.Fatalf("%s %s is clipped in %d-DIP preview: %+v", size, label, demo.Height, r)
					}
				}
			}
		})
	}
}
