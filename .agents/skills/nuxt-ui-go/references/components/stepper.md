<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# Stepper

Display a process and select its current step.

- Category: Overlays and navigation
- Version: nuxt-ui-go v0.0.2 (MyGo v0.8.3)
- Functions: `Stepper`
- Types: `StepperOptions`, `StepItem`
- Theme key: `nui.ComponentStepper` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Store the current stable step ID. Linear defaults to true, allowing previous steps or the immediately following step. Arrows move focus; Enter/Space activates. StepItem.Content builds the active body, and UI styles individual items.

## Things to know

Changing steps unmounts the previous body; application drafts remain caller-owned. Reordering or changing orientation with the same active ID preserves body identity. Unknown IDs mean not started and do not select the first item automatically. There are no Nuxt-style exposed next/prev methods; update the bound ID from the application.

## API

```go
func Stepper(c *ui.Context, t Theme, value *string, items []StepItem, o StepperOptions) ui.Element
```

### StepperOptions

```go
type StepperOptions struct {
	Label              string
	Vertical, Disabled bool
	// Linear defaults to true, as Nuxt's: a step can be reached only after
	// the one before it. Set(false) allows any available step.
	Linear    Optional[bool]
	Color     Color
	Size      Size
	CanSelect func(string) bool
	OnChange  func(string)
	Decoration
}
```

### StepItem

```go
type StepItem struct {
	ID, Title, Description string
	Icon                   *ui.SVG
	Disabled               bool
	UI                     Styles
	Content                func(*ui.Context, Theme)
}
```

Types documented in other references: `Color` → [theme.md](theme.md), `Decoration` → [theme.md](theme.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Styles` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Overlays, menus, and navigation

#### Shared conventions › Open state and stable IDs (excerpt)

List item IDs must be non-empty and unique, or the component panics. Tabs, Accordion, Breadcrumb, Stepper, and Timeline require uniqueness within their own list; menus require it across the whole menu tree. Display names can repeat. Required state pointers must not be nil, except for Banner, which explicitly accepts nil to mean "always shown".

#### Breadcrumb, Pagination, Stepper, and Timeline

##### Stepper: choose a step in a process

Stepper binds the current step as a `*string`. StepItem contains ID, Title, Description, Icon, Disabled, UI, and Content. An unknown ID means the process has not started, and the first item is not selected automatically. Content synchronously builds the current step's body; switching steps unmounts the old body. Your app stores business drafts. Reordering or changing orientation keeps the body's identity for the same ID.

Linear is an Optional[bool] that defaults to true, which allows going back or forward only to the next adjacent step. Set it to `nui.Set(false)` to allow selecting any available step. CanSelect can veto an activation. Arrow keys only move focus; Enter/Space or clicking the dot activates a step. The component does not skip disabled steps in a linear process automatically; your app arranges the flow.

Horizontal steppers give each step equal width, with the dot centered and the text below. Vertical places the text next to the dot and makes the connecting line vertical. Size defaults to MD, and the dot diameter is 24/32/40/48/56 DIP for XS to XL, scaled by Density. Earlier steps and their lines use the completed color, and the current step is highlighted. The original icon or number is kept and is not replaced by a check mark.

The step body is below a horizontal list or beside a vertical one. List styles the step list and Panel styles the current body. With Neutral, the dots and completed lines use InvertedBackground. Disabled disables only the step headers; controls in the body keep their own disabled settings. If the focused header is removed or disabled, or the focused body is switched, focus moves back to an available header. Changing the value from the app does not fire OnChange. Nuxt UI's next/prev and hasNext/hasPrev are implemented in native apps from the ID and the item list.

#### Part index (excerpt)

| Component constant suffix | Parts |
| --- | --- |
| Stepper | Root, List, Item, Header, Trigger, Indicator, Leading, SeparatorSlot, Body, Label, Description, Panel |

In Stepper, Header is the container for the dot and the connecting line, Trigger is the round button, Leading is the icon, SeparatorSlot is the connecting line, and Body is the text area.

Tabs, Stepper, and Timeline accept global and instance Color and Size; Banner and Toast accept Color. Color on menu items and Timeline items is a per-item setting. Other styles are configured through the parts listed above.

### Options and defaults

#### Common options by component › Sidebar, navigation, and file selection (excerpt)

| Component | Common options and behavior |
| --- | --- |
| Stepper | StepItem.Content builds the current step's body and UI overrides a single step. Disabled disables only the step headers. List and Panel configure the list and the body |

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type example struct {
	Step, Draft string
	Reversed    bool
	Changes     int
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	var styles nui.Styles

	items := []nui.StepItem{
		{ID: "details", Title: "Details", Description: "Project name", Icon: lucide.Folder(), UI: styles, Content: func(c *ui.Context, inner nui.Theme) {
			nui.Input(c, inner, &s.Draft, nui.InputOptions{Label: "Step draft"}).FillWidth()
		}},
		{ID: "team", Title: "Team", Description: "Invite members", Content: func(c *ui.Context, inner nui.Theme) { ui.Text(c, "Choose the people who can access the project.") }},
		{ID: "review", Title: "Review", Description: "Confirm changes", Content: func(c *ui.Context, inner nui.Theme) { ui.Text(c, "Review the project before continuing.") }},
	}

	if s.Reversed {
		slices.Reverse(items)
	}
	nui.Stepper(c.Key("stepper"), t, &s.Step, items, nui.StepperOptions{Label: "Setup", CanSelect: func(id string) bool { return true }, OnChange: func(string) { s.Changes++ }})
	nui.Button(c, t, "Reverse steps", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { s.Reversed = !s.Reversed })
	nui.Button(c, t, "Select review externally", nui.ButtonOptions{Variant: nui.Outline}).KeepFocus().OnClick(func() { s.Step = "review" })
	ui.Textf(c, "Step: %s · Changes: %d", s.Step, s.Changes)
}

func main() {
	s := &example{}

	s.Step = "details"
	s.Draft = "Retained step draft"
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "Stepper · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
