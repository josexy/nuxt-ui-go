package showcase

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	nui "github.com/josexy/nuxt-ui-go"
)

func TestPopoverDurationUnitsMatchPreviewAndSource(t *testing.T) {
	demo, _ := Find("popover")
	cfg, err := demo.Normalize(Config{Options: map[string]any{"PopoverOptions": map[string]any{"Mode": "nui.PopoverHover", "OpenDelay": float64(2000), "CloseDelay": float64(2.5)}}})
	if err != nil {
		t.Fatal(err)
	}
	opts := Options(cfg, "PopoverOptions", nui.PopoverOptions{}, nil, nui.Theme{}, &Runtime{})
	if opts.Mode != nui.PopoverHover || opts.OpenDelay != 2*time.Second || opts.CloseDelay != 2500*time.Microsecond {
		t.Fatalf("duration units differ from editor: %+v", opts)
	}
	_, source, err := demo.Source(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"OpenDelay = 2000 * time.Millisecond", "CloseDelay = time.Duration(2500000)"} {
		if !strings.Contains(source, expected) {
			t.Fatalf("source does not preserve duration: %s", source)
		}
	}
}

func TestEveryPublicOptionsFieldHasAnEditor(t *testing.T) {
	schemas := Schemas()
	for _, demo := range Catalog() {
		bindings := demo.Bindings()
		for _, name := range demo.Types {
			if !strings.HasSuffix(name, "Options") {
				continue
			}
			typ := optionTypes["nui."+name]
			found := false
			for _, binding := range bindings {
				if binding.typ == typ {
					found = true
					for _, field := range publicFields(typ) {
						matched := false
						for _, editor := range binding.Fields {
							if editor.Name == field.Name && editor.Type == goType(field.Type) {
								matched = true
								if _, ok := schemas[editor.Type]; !ok {
									t.Errorf("%s.%s has no editor schema", name, field.Name)
								}
							}
						}
						if !matched {
							t.Errorf("%s: %s.%s missing", demo.ID, name, field.Name)
						}
					}
				}
			}
			if !found {
				t.Errorf("%s: %s not connected to preview", demo.ID, name)
			}
		}
	}
	for name, schema := range schemas {
		if schema.Kind == "object" {
			for _, field := range schema.Fields {
				if _, ok := schemas[field.Type]; !ok {
					t.Errorf("%s.%s nested schema missing", name, field.Name)
				}
			}
		}
	}
}

func TestWorkerEventsAndCallbackResults(t *testing.T) {
	runtime := &Runtime{}
	var events atomic.Int64
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			for range 50 {
				runtime.SetReport(func(string, []string) { events.Add(1) })
				runtime.emit("worker", nil)
			}
		})
	}
	workers.Wait()
	if events.Load() != 400 {
		t.Fatal("events lost")
	}
	typ := reflect.TypeFor[func(nui.FormValues) nui.FormErrors]()
	fn, _, err := functionValue(typ, map[string]any{"mode": "result", "results": []any{map[string]any{"name": "Required"}}}, "FormOptions.Validate", nil, nil, nil, nui.Theme{})
	if err != nil {
		t.Fatal(err)
	}
	call := fn.Interface().(func(nui.FormValues) nui.FormErrors)
	first := call(nil)
	first["name"] = "changed"
	if call(nil)["name"] != "Required" {
		t.Fatal("callback result was reused across invocations")
	}
}

func TestExplicitFieldOverridesAndPersistentPointers(t *testing.T) {
	demo, _ := Find("modal")
	cfg, err := demo.Normalize(Config{Options: map[string]any{"ModalOptions": map[string]any{"Dismissible": false, "Width": float64(0), "ReturnFocus": "target", "UI": map[string]any{"panel": map[string]any{"Radius": float64(0)}}}}})
	if err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{}
	original := nui.ModalOptions{Dismissible: nui.Set(true), Width: nui.Set(float32(400))}
	got := Options(cfg, "ModalOptions", original, nil, nui.Theme{}, runtime)
	if !got.Dismissible.Set || got.Dismissible.Value || !got.Width.Set || got.Width.Value != 0 || !got.UI[nui.Panel].Radius.Set || got.UI[nui.Panel].Radius.Value != 0 || got.ReturnFocus != &runtime.Focus {
		t.Fatalf("lost explicit override: %+v", got)
	}
	_, source, err := demo.Source(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"Dismissible = nui.Set(false)", "Width = nui.Set(float32(", "ReturnFocus = &s.ReturnFocus", "Radius = nui.Set(float32("} {
		if !strings.Contains(source, part) {
			t.Fatalf("source missing %q", part)
		}
	}
	typ := reflect.TypeFor[*bool]()
	first, _, err := value(typ, true, "Visible", runtime, nil, nil, nui.Theme{})
	if err != nil {
		t.Fatal(err)
	}
	first.Elem().SetBool(false)
	again, _, err := value(typ, true, "Visible", runtime, nil, nil, nui.Theme{})
	if err != nil || first.Pointer() != again.Pointer() || again.Elem().Bool() {
		t.Fatal("rebuild reset persistent boolean")
	}
}

func TestRecordedCallbacksKeepExampleBehavior(t *testing.T) {
	cfg := Config{Options: map[string]any{"InputOptions": map[string]any{"OnChange": map[string]any{"mode": "record"}}}}
	calls, events := 0, 0
	runtime := &Runtime{}
	runtime.SetReport(func(string, []string) { events++ })
	original := nui.InputOptions{OnChange: func(string) { calls++ }}
	updated := Options(cfg, "InputOptions", original, nil, nui.Theme{}, runtime)
	updated.OnChange("edited")
	if calls != 1 || events != 1 {
		t.Fatalf("callback counts: %d/%d", calls, events)
	}
}

func TestEditedItemsKeepCallbacksByIdentity(t *testing.T) {
	called := ""
	base := []nui.MenuItem{
		{ID: "edit", Label: "Edit", OnSelect: func() { called = "edit" }},
		{ID: "copy", Label: "Copy", OnSelect: func() { called = "copy" }},
	}
	patch := []any{
		map[string]any{"ID": "copy", "Label": "Duplicate"},
		map[string]any{"ID": "new", "Label": "New action"},
		map[string]any{"ID": "edit", "Label": "Rename"},
	}
	cfg := Config{Options: map[string]any{"items": patch}}
	items := Options(cfg, "items", base, nil, nui.Theme{}, &Runtime{})
	items[0].OnSelect()
	if called != "copy" {
		t.Fatal("reordering changed action identity")
	}
	if items[1].OnSelect != nil {
		t.Fatal("new item inherited an unrelated callback")
	}
	items[2].OnSelect()
	if called != "edit" {
		t.Fatal("moving an item lost its callback")
	}
	if base[0].Label != "Edit" || base[1].Label != "Copy" {
		t.Fatal("configuration mutated defaults")
	}
}

func TestInvalidFieldsDoNotReachComponentBuild(t *testing.T) {
	demo, _ := Find("input")
	for _, options := range []map[string]any{{"Unknown": map[string]any{}}, {"InputOptions": map[string]any{"Unknown": true}}, {"InputOptions": map[string]any{"Variant": "nui.Solid"}}, {"InputOptions": map[string]any{"UI": map[string]any{"root": map[string]any{"Opacity": float64(2)}}}}} {
		if _, err := demo.Normalize(Config{Options: options}); err == nil {
			t.Errorf("accepted invalid fields: %v", options)
		}
	}
}

func TestClearedNestedStylesGenerateQualifiedTypes(t *testing.T) {
	demo, _ := Find("button")
	_, source, err := demo.Source(Config{Options: map[string]any{"ButtonOptions": map[string]any{"UI": map[string]any{"root": map[string]any{"Foreground": nil, "Radius": nil}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "nui.Optional[ui.Color]{}") || !strings.Contains(source, "nui.Optional[float32]{}") {
		t.Fatal("cleared optional style has the wrong Go type")
	}
}
