package main

import (
	"regexp"
	"testing"

	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

func TestMyGoVersionComesFromGoMod(t *testing.T) {
	v, err := mygoVersion("../../..")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(v) {
		t.Fatalf("MyGo version %q, want a semantic version without its v", v)
	}
	if _, err := mygoVersion(t.TempDir()); err == nil {
		t.Fatal("a missing go.mod gave no error")
	}
}

func TestAPISourcePathsAreRepositoryRelative(t *testing.T) {
	e := extract("../../..")
	for _, name := range []string{"Button", "ButtonOptions"} {
		entry := e.declarations[name]
		if entry.File != "ui/button.go" || entry.Line <= 0 {
			t.Errorf("%s source location = %q:%d, want ui/button.go and a positive line", name, entry.File, entry.Line)
		}
	}
}

func TestSidebarDescriptionsDoNotInheritAccordionMeaning(t *testing.T) {
	e := extract("../../..")
	for _, name := range []string{"SidebarOptions", "AccordionOptions"} {
		fields := e.fields(e.types[name], map[string]bool{}, name)
		var description showcase.Text
		for _, f := range fields {
			if f.Name == "Collapsible" {
				description = f.Description
			}
		}
		key := name + ".Collapsible"
		if description.EN == "" || description != e.notes[key] {
			t.Fatalf("%s has the wrong collapse description: %+v", name, description)
		}
	}
}

func TestChoiceAPITablesUseTheDeclaredAlias(t *testing.T) {
	e := extract("../../..")
	for _, tc := range []struct{ owner, field string }{
		{"CheckboxOptions", "Loading"},
		{"CheckboxTriOptions", "Loading"},
		{"SwitchOptions", "Loading"},
		{"CheckboxOptions", "UncheckedIcon"},
		{"SwitchOptions", "UncheckedIcon"},
		{"CheckboxGroupOptions", "Max"},
		{"RadioGroupOptions", "Max"},
		{"CheckboxOptions", "Variant"},
		{"CheckboxGroupOptions", "Variant"},
	} {
		fields := e.fields(e.types[tc.owner], map[string]bool{}, tc.owner)
		found := false
		for _, f := range fields {
			if f.Name == tc.field {
				found = true
				want := e.notes[tc.owner+"."+tc.field]
				if want.EN == "" || want.ZH == "" || f.Description != want {
					t.Errorf("%s.%s lost its component-specific meaning: %+v", tc.owner, tc.field, f.Description)
				}
			}
		}
		if !found {
			t.Errorf("%s.%s missing from API table", tc.owner, tc.field)
		}
	}
}
