package preview

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

func TestInputPreviewAccessoriesAndCallbacks(t *testing.T) {
	tester, cfg := exampleTester(t, "input")
	clickExample(t, tester, "Project name")
	tester.Type("Original")
	tester.Key(ui.Ctrl, ui.KeyA)
	cfg.Values["avatar"], cfg.Values["loading"], cfg.Values["loadingTrailing"] = true, true, true
	tester.Frame()
	if !tester.HasText("Input preview avatar") || !tester.Focused("Project name") {
		t.Fatal("input preview failed to preserve its avatar or focus")
	}
	tester.Type("Updated")
	tester.Key(0, ui.KeyEnter)
	if !tester.HasText("Value: Updated") || !tester.HasText("Changes: 2 · Submits: 1 · Clears: 0") {
		t.Fatal("input preview values or callbacks are not connected")
	}
	clickExample(t, tester, "Clear Project name")
	if !tester.HasText("Changes: 3 · Submits: 1 · Clears: 1") || !tester.Focused("Project name") {
		t.Fatal("input preview clear failed")
	}
	cfg.Values["loadingTrailing"] = false
	tester.Frame()
	if tester.HasText("Input preview avatar") {
		t.Fatal("leading loader did not replace the preview avatar")
	}
}

func TestInputPreviewNativeLengthCaret(t *testing.T) {
	for _, id := range []string{"input", "textarea"} {
		t.Run(id, func(t *testing.T) {
			tester, cfg := exampleTester(t, id)
			binding := "InputOptions"
			if id == "textarea" {
				binding = "TextareaOptions"
			}
			cfg.Options = map[string]any{binding: map[string]any{"MaxLength": float64(4)}}
			tester.Frame()
			clickExample(t, tester, "Project name")
			tester.Type("abcd")
			tester.Key(0, ui.KeyHome)
			tester.Key(0, ui.KeyRight)
			tester.Key(0, ui.KeyRight)
			tester.Type("X")
			tester.Key(0, ui.KeyBackspace)
			want := "Changes: 2 · Clears: 0"
			if id == "input" {
				want = "Changes: 2 · Submits: 0 · Clears: 0"
			}
			if !tester.HasText(want) || !tester.Focused("Project name") {
				t.Fatalf("native limit preview: %v", tester.Texts())
			}
			tester.Key(ui.Ctrl, ui.KeyA)
			tester.Command("copy")
			if tester.Clipboard() != "acd" {
				t.Fatalf("preview caret edit produced %q, want acd", tester.Clipboard())
			}
			demo, _ := showcase.Find(id)
			_, source, err := demo.Source(*cfg)
			if err != nil || !strings.Contains(strings.Join(strings.Fields(source), " "), "o.MaxLength = 4") {
				t.Fatalf("copied program lost the explicit native limit: %v\n%s", err, source)
			}
		})
	}
}

func TestTextareaPreviewRowsAccessoriesAndClear(t *testing.T) {
	tester, cfg := exampleTester(t, "textarea")
	clickExample(t, tester, "Project name")
	tester.Type("First\nSecond\nThird\nFourth\nFifth")
	cfg.Values["rows"], cfg.Values["maxRows"] = float64(2), float64(4)
	cfg.Values["autoResize"], cfg.Values["avatar"], cfg.Values["loading"], cfg.Values["loadingTrailing"] = true, true, true, true
	tester.Frame()
	if !tester.HasText("Textarea preview avatar") || !tester.Focused("Project name") {
		t.Fatal("textarea controls lost the avatar or editor focus")
	}
	before, ok := tester.Find("Project name")
	if !ok {
		t.Fatal("missing textarea editor")
	}
	cfg.Values["maxRows"] = float64(2)
	tester.Frame()
	after, _ := tester.Find("Project name")
	if after.H >= before.H {
		t.Fatal("maximum row control did not constrain the editor")
	}
	clickExample(t, tester, "Clear Project name")
	if !tester.HasText("Changes: 2 · Clears: 1") || !tester.Focused("Project name") {
		t.Fatal("textarea clear control is not connected")
	}
}
