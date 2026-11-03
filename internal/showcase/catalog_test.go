package showcase

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestCatalogAndStandaloneCode(t *testing.T) {
	demos := Catalog()
	if len(demos) != 73 {
		t.Fatalf("component coverage: %d", len(demos))
	}
	seen := map[string]bool{}
	for _, demo := range demos {
		t.Run(demo.ID, func(t *testing.T) {
			if seen[demo.ID] {
				t.Fatal("duplicate component")
			}
			seen[demo.ID] = true
			if demo.Description.ZH == "" || demo.Description.EN == "" || demo.Guide.ZH == "" || demo.Guide.EN == "" {
				t.Fatal("missing translation")
			}
			cfg, err := demo.Normalize(Config{})
			if err != nil {
				t.Fatal(err)
			}
			_, code, err := demo.Source(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parser.ParseFile(token.NewFileSet(), "main.go", code, parser.AllErrors); err != nil {
				t.Fatal(err)
			}
			for _, control := range demo.Controls {
				values := []any{control.Default}
				switch control.Kind {
				case "select":
					for _, choice := range control.Choices {
						values = append(values, choice.Value)
					}
				case "boolean":
					values = append(values, true, false)
				case "optionalBoolean":
					values = append(values, nil, true, false)
				case "text":
					values = append(values, "中文 \"quoted\"\ntext ` $()")
				case "number", "integer":
					values = append(values, control.Min, control.Max)
				}
				for _, value := range values {
					cfg.Values[control.Key] = value
					_, code, err = demo.Source(cfg)
					if err != nil {
						t.Fatalf("%s=%v: %v", control.Key, value, err)
					}
					if strings.Contains(code, "{{"+control.Key+"}}") {
						t.Fatalf("unexpanded %s", control.Key)
					}
				}
				cfg.Values[control.Key] = control.Default
			}
		})
	}
}

func TestConfigurationBoundaryAndOptionalValues(t *testing.T) {
	demo, _ := Find("modal")
	cfg, err := demo.Normalize(Config{Values: map[string]any{"dismissible": false}})
	if err != nil {
		t.Fatal(err)
	}
	if value := cfg.OptionalBool("dismissible"); !value.Set || value.Value {
		t.Fatal("explicit false lost")
	}
	if cfg.OptionalBool("closeButton").Set {
		t.Fatal("unset value became explicit")
	}
	_, code, err := demo.Source(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "Dismissible: nui.Set(false)") || !strings.Contains(code, "CloseButton: nui.Optional[bool]{}") {
		t.Fatal("copied code lost optional state")
	}
	for _, input := range []Config{{Values: map[string]any{"dismissible": "false"}}, {Values: map[string]any{"unknown": true}}, {Primary: "unknown"}, {Mode: "unknown"}} {
		if _, err := demo.Normalize(input); err == nil {
			t.Fatalf("accepted invalid config: %+v", input)
		}
	}
	page, _ := Find("pagination")
	_, code, err = page.Source(Config{Values: map[string]any{"siblings": float64(0)}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "SiblingCount: nui.Set(0)") {
		t.Fatal("explicit zero lost")
	}
}

func TestCompiledViewsMatchSourceTemplates(t *testing.T) {
	want, err := PreviewSource()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("preview/generated.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatal("generated preview drift: run go run ./internal/cmd/genwebsite")
	}
}

func TestCodePreservesTemplateLikeUserText(t *testing.T) {
	demo, _ := Find("input")
	_, source, err := demo.Source(Config{Values: map[string]any{
		"label": "{{placeholder}}", "placeholder": "中文 \"quoted\"",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, `Label: "{{placeholder}}"`) || !strings.Contains(source, `Placeholder: "中文 \"quoted\""`) {
		t.Fatal("code generation expanded or changed user-entered text")
	}
}
