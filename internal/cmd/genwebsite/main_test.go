package main

import (
	"testing"

	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

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
		key := "Collapsible"
		if name == "SidebarOptions" {
			key = "SidebarOptions.Collapsible"
		}
		if description.EN == "" || description != e.notes[key] {
			t.Fatalf("%s has the wrong collapse description: %+v", name, description)
		}
	}
}
