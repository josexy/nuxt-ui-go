package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func iconFixture(t *testing.T, names ...string) (string, iconManifest) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "svg"), 0755); err != nil {
		t.Fatal(err)
	}
	m := iconManifest{Commit: strings.Repeat("a", 40)}
	for _, name := range names {
		data := []byte(`<svg width="24" height="24"><path d="M4 4L20 20"/></svg>`)
		if err := os.WriteFile(filepath.Join(dir, "svg", name+".svg"), data, 0644); err != nil {
			t.Fatal(err)
		}
		m.Icons = append(m.Icons, iconEntry{name, fmt.Sprintf("%x", sha256.Sum256(data))})
	}
	writeManifest(t, dir, m)
	return dir, m
}

func writeManifest(t *testing.T, dir string, m iconManifest) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func generatedFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		out[filepath.Base(file)] = string(data)
	}
	return out
}

func TestGenerationIsCompleteAndDeterministic(t *testing.T) {
	dir, manifest := iconFixture(t, "robot-arm", "file-json-2", "check")
	if err := generate(dir); err != nil {
		t.Fatal(err)
	}
	first := generatedFiles(t, dir)
	if len(first) != 4 {
		t.Fatal("missing getters or test catalog", len(first))
	}
	if !strings.Contains(first["robot_arm_icon.go"], "func RobotArm()") || !strings.Contains(first["file_json_2.go"], "func FileJson2()") {
		t.Fatal("platform-safe filename or numeric name mapping")
	}
	if _, present := first["robot_arm.go"]; present {
		t.Fatal("architecture-constrained getter")
	}
	if !strings.Contains(first["icons_catalog_test.go"], `{"robot-arm", RobotArm}`) {
		t.Fatal("missing test coverage")
	}
	// Manifest ordering does not change output or require a runtime registry.
	manifest.Icons[0], manifest.Icons[2] = manifest.Icons[2], manifest.Icons[0]
	writeManifest(t, dir, manifest)
	if err := generate(dir); err != nil {
		t.Fatal(err)
	}
	second := generatedFiles(t, dir)
	for name, data := range first {
		if second[name] != data {
			t.Fatal("nondeterministic generation", name)
		}
		if !strings.HasSuffix(name, "_test.go") && strings.Contains(data, "generatedIconCases") {
			t.Fatal("full registry retained in production")
		}
	}
}

func TestInvalidAssetsDoNotPartiallyWrite(t *testing.T) {
	tests := []struct {
		name, want string
		change     func(*testing.T, string, *iconManifest)
	}{
		{"hash", "hash mismatch", func(t *testing.T, dir string, m *iconManifest) { m.Icons[1].SHA256 = strings.Repeat("0", 64) }},
		{"duplicate", "duplicate icon", func(t *testing.T, dir string, m *iconManifest) { m.Icons = append(m.Icons, m.Icons[0]) }},
		{"path", "invalid or duplicate", func(t *testing.T, dir string, m *iconManifest) { m.Icons[1].Name = "../escape" }},
		{"unlisted", "missing from manifest", func(t *testing.T, dir string, m *iconManifest) { m.Icons = m.Icons[:1] }},
		{"handwritten", "hand-written file", func(t *testing.T, dir string, m *iconManifest) {
			if err := os.WriteFile(filepath.Join(dir, "check.go"), []byte("package lucide\n// handwritten\n"), 0644); err != nil {
				t.Fatal(err)
			}
		}},
		{"stale", "stale generated file", func(t *testing.T, dir string, m *iconManifest) {
			if err := os.WriteFile(filepath.Join(dir, "removed.go"), []byte(generatedPrefix+m.Commit+". DO NOT EDIT.\npackage lucide\n"), 0644); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, manifest := iconFixture(t, "check", "sun")
			test.change(t, dir, &manifest)
			writeManifest(t, dir, manifest)
			before := generatedFiles(t, dir)
			err := generate(dir)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v; want %q", err, test.want)
			}
			after := generatedFiles(t, dir)
			if len(before) != len(after) {
				t.Fatal("partial generated output after validation failure")
			}
			for name, data := range before {
				if !bytes.Equal([]byte(after[name]), []byte(data)) {
					t.Fatal("file changed on validation failure", name)
				}
			}
		})
	}
}

func TestNameCollisionsAreRejected(t *testing.T) {
	for _, test := range []struct {
		names []string
		want  string
	}{
		{[]string{"a-1-2", "a-12"}, "getter A12 collides"},
		{[]string{"robot-arm", "robot-arm-icon"}, "filename collision"},
		{[]string{"icons"}, "conflicts"},
		{[]string{"sample-test"}, "conflicts"},
	} {
		t.Run(strings.Join(test.names, "+"), func(t *testing.T) {
			dir, _ := iconFixture(t, test.names...)
			if err := generate(dir); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v; want %q", err, test.want)
			}
			if len(generatedFiles(t, dir)) != 0 {
				t.Fatal("partial output after collision")
			}
		})
	}
}
