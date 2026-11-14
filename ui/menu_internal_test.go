package nui

import (
	"slices"
	"testing"
)

func TestShortcutKeys(t *testing.T) {
	for in, want := range map[string][]string{
		"Ctrl+O":       {"Ctrl", "O"},
		"Ctrl+Shift+P": {"Ctrl", "Shift", "P"},
		"Ctrl + O":     {"Ctrl", "O"},
		"Ctrl++":       {"Ctrl", "+"},
		"+":            {"+"},
		"⌘K":           {"⌘K"},
		"Ctrl S":       {"Ctrl S"},
	} {
		if got := shortcutKeys(in); !slices.Equal(got, want) {
			t.Errorf("shortcutKeys(%q) = %q, want %q", in, got, want)
		}
	}
}
