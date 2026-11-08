package showcase

import (
	"slices"
	"strings"
	"testing"
)

func TestSharedOptionFieldsFollowCalledComponent(t *testing.T) {
	// All five calls share only two underlying options types. The call site,
	// rather than a type name or page ID, determines what their fields mean.
	demo := Demo{Symbols: []string{"Checkbox", "CheckboxTri", "Switch", "CheckboxGroup", "RadioGroup"}, Body: `
nui.Checkbox(c, t, &checked, nui.SwitchOptions{})
nui.Switch(c, t, &enabled, nui.CheckboxOptions{})
nui.CheckboxTri(c, t, &mixed, nui.CheckboxTriOptions{})
nui.CheckboxGroup(c, t, &values, items, nui.RadioGroupOptions{})
nui.RadioGroup(c, t, &value, items, nui.CheckboxGroupOptions{})
`}
	bindings := demo.Bindings()
	for _, tc := range []struct {
		id, field, owner string
		unused, native   bool
	}{
		{"ToggleOptions", "Loading", "CheckboxOptions", true, false},
		{"ToggleOptions", "UncheckedIcon", "CheckboxOptions", true, false},
		{"ToggleOptions", "Variant", "CheckboxOptions", false, true},
		{"ToggleOptions2", "Loading", "SwitchOptions", false, false},
		{"ToggleOptions2", "UncheckedIcon", "SwitchOptions", false, false},
		{"ToggleOptions2", "Variant", "SwitchOptions", false, true},
		{"CheckboxTriOptions", "Loading", "CheckboxOptions", true, false},
		{"ChoiceGroupOptions", "Max", "CheckboxGroupOptions", false, false},
		{"ChoiceGroupOptions2", "Max", "RadioGroupOptions", true, false},
	} {
		t.Run(tc.id+"/"+tc.field, func(t *testing.T) {
			i := slices.IndexFunc(bindings, func(b Binding) bool { return b.ID == tc.id })
			if i < 0 {
				t.Fatalf("existing configuration ID %s changed", tc.id)
			}
			j := slices.IndexFunc(bindings[i].Fields, func(f ValueField) bool { return f.Name == tc.field })
			if j < 0 {
				t.Fatal("public field was removed from the reference")
			}
			f := bindings[i].Fields[j]
			if f.Description != APINotes()[tc.owner+"."+tc.field] || f.Description.ZH == "" || f.Description.EN == "" {
				t.Fatalf("wrong component meaning: %+v", f.Description)
			}
			unused := f.Hint != nil && f.Hint.EN == "Not used"
			native := f.Hint != nil && strings.Contains(strings.ToLower(f.Hint.EN), "native extension")
			if unused != tc.unused || native != tc.native {
				t.Fatalf("wrong applicability: %+v", f.Hint)
			}
		})
	}
}

func TestChoiceExtensionsKeepConfigurationAndSourceCompatibility(t *testing.T) {
	for _, id := range []string{"checkbox", "switch", "checkbox-group", "radio-group"} {
		demo, _ := Find(id)
		found := false
		for _, binding := range demo.Bindings() {
			if !slices.Contains([]string{"ToggleOptions", "ChoiceGroupOptions"}, binding.ID) {
				continue
			}
			found = true
			cfg, err := demo.Normalize(Config{Options: map[string]any{binding.ID: map[string]any{"Variant": "nui.ChoiceTable"}}})
			if err != nil {
				t.Fatalf("%s rejected a supported extension: %v", id, err)
			}
			_, source, err := demo.Source(cfg)
			if err != nil || !strings.Contains(source, ".Variant = nui.ChoiceTable") {
				t.Fatalf("%s lost configured variant: %v", id, err)
			}
		}
		if !found {
			t.Fatalf("%s: missing existing shared-options binding", id)
		}
	}
}
