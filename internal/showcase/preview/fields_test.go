package preview

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

func editorSample(schemas map[string]showcase.ValueSchema, typ, name string, choices []string) any {
	s := schemas[typ]
	switch s.Kind {
	case "optional", "pointer":
		return editorSample(schemas, s.Item, name, nil)
	case "object":
		return map[string]any{}
	case "array":
		items := []any{}
		for range s.Length {
			items = append(items, editorSample(schemas, s.Item, "", nil))
		}
		return items
	case "map":
		if typ == "nui.Styles" {
			return map[string]any{"root": map[string]any{"Radius": float64(4), "Foreground": "#15803d"}}
		}
		return map[string]any{}
	case "boolean":
		return true
	case "number", "integer":
		if name == "Max" {
			return float64(20)
		}
		return float64(1)
	case "enum":
		if len(choices) > 0 {
			return choices[0]
		}
		return s.Choices[0]
	case "text":
		return "Configured"
	case "image":
		return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR4nGP4z8DwHwAFAAH/iZk9HQAAAABJRU5ErkJggg=="
	case "function":
		modes := s.Choices
		if len(choices) > 0 {
			modes = choices
		}
		mode := modes[0]
		result := map[string]any{"mode": mode}
		switch mode {
		case "content":
			result["text"] = "Configured content"
		case "style":
			result["parts"] = map[string]any{"root": map[string]any{"Radius": float64(4)}}
		case "format":
			result["text"] = "% "
			result["precision"] = float64(1)
		case "result":
			var values []any
			for _, r := range s.Returns {
				if r == "error" {
					values = append(values, "")
				} else {
					values = append(values, editorSample(schemas, r, "", nil))
				}
			}
			result["results"] = values
		}
		return result
	}
	return s.Sample
}

func TestAllFieldEditorsUpdateRealViews(t *testing.T) {
	schemas := showcase.Schemas()
	directory := os.Getenv("SHOWCASE_SNIPPETS_DIR")
	count := 0
	for _, demo := range showcase.Catalog() {
		t.Run(demo.ID, func(t *testing.T) {
			example, err := New(demo.ID)
			if err != nil {
				t.Fatal(err)
			}
			defer example.Dispose()
			cfg, err := demo.Normalize(showcase.Config{})
			if err != nil {
				t.Fatal(err)
			}
			tester := ui.NewTester(func(c *ui.Context) {
				theme := cfg.Theme()
				theme.Light.Motion.Reduce = true
				theme.Dark.Motion.Reduce = true
				t := nui.App(c, theme)
				ui.Column(c.Key("preview")).Fill().Padding(24).Children(func() { example.View(c, t, cfg) })
			}, 760, demo.Height)
			tester.Frame()
			for _, binding := range demo.Bindings() {
				if len(binding.Fields) == 0 && strings.HasPrefix(binding.Type, "[]") {
					t.Run(binding.ID, func(t *testing.T) {
						encoded, _ := json.Marshal(binding.Defaults)
						var items []any
						if err := json.Unmarshal(encoded, &items); err != nil {
							t.Fatal(err)
						}
						if len(items) == 0 {
							return
						}
						if item, ok := items[0].(map[string]any); ok {
							if _, ok := item["Label"]; ok {
								item["Label"] = "Configured item"
							} else if _, ok := item["Title"]; ok {
								item["Title"] = "Configured item"
							}
						}
						next, err := demo.Normalize(showcase.Config{Options: map[string]any{binding.ID: items}})
						if err != nil {
							t.Fatal(err)
						}
						_, source, err := demo.Source(next)
						if err != nil {
							t.Fatal(err)
						}
						if directory != "" {
							target := filepath.Join(directory, demo.ID, "data_"+binding.ID, "main.go")
							if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
								t.Fatal(err)
							}
							if err = os.WriteFile(target, []byte(source), 0644); err != nil {
								t.Fatal(err)
							}
						}
						if err = example.Configure(next); err != nil {
							t.Fatal(err)
						}
						cfg = next
						tester.Frame()
						_ = tester.Image()
						cfg, err = demo.Normalize(showcase.Config{})
						if err != nil {
							t.Fatal(err)
						}
						if err = example.Configure(cfg); err != nil {
							t.Fatal(err)
						}
						tester.Frame()
					})
				}
				for _, field := range binding.Fields {
					name := binding.ID + "." + field.Name
					t.Run(name, func(t *testing.T) {
						raw := editorSample(schemas, field.Type, field.Name, field.Choices)
						if binding.ID == "SplitterOptions" && field.Name == "Max" {
							raw = float64(90) // Keep the sample above this demo's 30% minimum.
						}
						if binding.ID == "SidebarOptions" {
							switch field.Name {
							case "MinWidth":
								raw = float64(160)
							case "MaxWidth":
								raw = float64(480)
							case "CollapsedWidth":
								raw = float64(64)
							}
						}
						input := showcase.Config{Options: map[string]any{binding.ID: map[string]any{field.Name: raw}}}
						next, err := demo.Normalize(input)
						if err != nil {
							t.Fatal(err)
						}
						_, source, err := demo.Source(next)
						if err != nil {
							t.Fatal(err)
						}
						if strings.Contains(source, "nuxt-ui-go/internal/") {
							t.Fatal("standalone source imports internal code")
						}
						if directory != "" {
							target := filepath.Join(directory, demo.ID, fmt.Sprintf("field_%03d", count), "main.go")
							if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
								t.Fatal(err)
							}
							if err = os.WriteFile(target, []byte(source), 0644); err != nil {
								t.Fatal(err)
							}
						}
						count++
						cfg = next
						tester.Frame()
						tester.Frame()
						_ = tester.Image()
						cfg, err = demo.Normalize(showcase.Config{})
						if err != nil {
							t.Fatal(err)
						}
						tester.Frame()
					})
				}
			}
		})
	}
	t.Logf("Verified %d public field editors against compiled Go views", count)
}

func TestClearedFieldSources(t *testing.T) {
	directory := os.Getenv("SHOWCASE_SNIPPETS_DIR")
	for _, demo := range showcase.Catalog() {
		t.Run(demo.ID, func(t *testing.T) {
			options := map[string]any{}
			for _, binding := range demo.Bindings() {
				if len(binding.Fields) == 0 {
					continue
				}
				fields := map[string]any{}
				for _, field := range binding.Fields {
					fields[field.Name] = nil
				}
				options[binding.ID] = fields
			}
			_, source, err := demo.Source(showcase.Config{Options: options})
			if err != nil {
				t.Fatal(err)
			}
			if directory != "" {
				target := filepath.Join(directory, demo.ID, "cleared", "main.go")
				if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(target, []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestNestedFieldSources(t *testing.T) {
	schemas := showcase.Schemas()
	directory := os.Getenv("SHOWCASE_SNIPPETS_DIR")
	count := 0
	for _, demo := range showcase.Catalog() {
		for _, binding := range demo.Bindings() {
			for _, field := range binding.Fields {
				if schemas[field.Type].Kind != "object" && field.Type != "nui.Styles" {
					continue
				}
				var visit func(string, []string)
				visit = func(typ string, path []string) {
					schema := schemas[typ]
					if schema.Kind == "object" {
						for _, nested := range schema.Fields {
							visit(nested.Type, append(append([]string{}, path...), nested.Name))
						}
						return
					}
					if schema.Kind == "map" && typ == "nui.Styles" {
						visit(schema.Item, append(path, "root"))
						return
					}
					name := strings.Join(path, ".")
					t.Run(demo.ID+"/"+binding.ID+"/"+name, func(t *testing.T) {
						patch := editorSample(schemas, typ, path[len(path)-1], nil)
						for _, p := range slices.Backward(path) {
							patch = map[string]any{p: patch}
						}
						_, source, err := demo.Source(showcase.Config{Options: map[string]any{binding.ID: patch}})
						if err != nil {
							t.Fatal(err)
						}
						if directory != "" {
							target := filepath.Join(directory, demo.ID, fmt.Sprintf("nested_%04d", count), "main.go")
							if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
								t.Fatal(err)
							}
							if err = os.WriteFile(target, []byte(source), 0644); err != nil {
								t.Fatal(err)
							}
						}
						count++
					})
				}
				visit(field.Type, []string{field.Name})
			}
		}
	}
	t.Logf("Verified %d nested field source configurations", count)
}
