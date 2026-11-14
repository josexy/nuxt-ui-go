package nui

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/egoist/mygo/i18n"
)

func TestComponentCatalogs(t *testing.T) {
	languages := SupportedLocales()
	if !slices.Equal(languages, []string{"en", "zh-Hans"}) {
		t.Fatalf("languages: %v", languages)
	}
	languages[0] = "changed"
	if SupportedLocales()[0] != "en" {
		t.Fatal("caller changed the catalog's languages")
	}
	catalogs := map[string]map[string]string{}
	for _, locale := range SupportedLocales() {
		data, err := localeFiles.ReadFile("locales/" + locale + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var entries map[string]string
		if err := json.Unmarshal(data, &entries); err != nil {
			t.Fatal(err)
		}
		if entries["@@locale"] != locale {
			t.Fatalf("wrong locale metadata in %s", locale)
		}
		delete(entries, "@@locale")
		catalogs[locale] = entries
	}
	for locale, entries := range catalogs {
		if len(entries) != len(catalogs["en"]) {
			t.Fatalf("%s has missing or extra keys", locale)
		}
		for key, source := range catalogs["en"] {
			translated, ok := entries[key]
			if !ok || strings.TrimSpace(translated) == "" {
				t.Fatalf("%s: missing translation for %s", locale, key)
			}
			original, err := i18n.ParseMessage(source)
			if err != nil {
				t.Fatalf("en.%s: %v", key, err)
			}
			m, err := i18n.ParseMessage(translated)
			if err != nil {
				t.Fatalf("%s.%s: %v", locale, key, err)
			}
			args, sourceArgs := m.Args(), original.Args()
			less := func(a, b i18n.Arg) int { return strings.Compare(a.Name, b.Name) }
			slices.SortFunc(args, less)
			slices.SortFunc(sourceArgs, less)
			if !slices.Equal(args, sourceArgs) {
				t.Fatalf("%s.%s changed message arguments: %v != %v", locale, key, args, sourceArgs)
			}
		}
	}
	for _, tc := range []struct {
		locale, key string
		args        map[string]any
		want        string
	}{
		{"en-US", "clear", nil, "Clear"},
		{"zh-CN", "clear", nil, "清空"},
		{"fr", "clear", nil, "Clear"},
		{"en", "tooManyFiles", map[string]any{"count": 1}, "Choose at most 1 file"},
		{"en", "tooManyFiles", map[string]any{"count": 2}, "Choose at most 2 files"},
		{"zh-Hans", "tooManyFiles", map[string]any{"count": 2}, "最多选择 2 个文件"},
		{"zh-Hans", "createItem", map[string]any{"name": "{draft}"}, "创建“{draft}”"},
		{"en", "pinCharacters", map[string]any{"count": 1}, "1 character"},
		{"en", "pinCharacters", map[string]any{"count": 5}, "5 characters"},
		{"zh-Hans", "pinCharacters", map[string]any{"count": 5}, "5 个字符"},
		{"en", "pinDigits", map[string]any{"count": 1}, "1 digit"},
		{"en", "pinDigits", map[string]any{"count": 6}, "6 digits"},
		{"zh-Hans", "pinDigits", map[string]any{"count": 6}, "6 位数字"},
	} {
		if got := componentMessages().Format(tc.locale, tc.key, tc.args); got != tc.want {
			t.Errorf("%s.%s = %q, want %q", tc.locale, tc.key, got, tc.want)
		}
	}
}

// The app-oriented MyGo catalog checker does not inspect our library wrappers.
// Check their literal keys as well as the bilingual catalog's internal parity.
func TestComponentMessageReferencesExist(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	positions := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(positions, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) < 2 {
				return true
			}
			fn, ok := call.Fun.(*ast.Ident)
			if !ok || (fn.Name != "message" && fn.Name != "namedMessage" && fn.Name != "formatMessage") {
				return true
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true // Dynamic keys are covered by their component regressions.
			}
			key, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, found := componentMessages().Message("en", key); !found {
				t.Errorf("%s: missing component message %q", positions.Position(literal.Pos()), key)
			}
			return true
		})
	}
}
