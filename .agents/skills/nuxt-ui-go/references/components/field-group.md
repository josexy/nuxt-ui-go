<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# FieldGroup

Join related inputs and buttons into one group.

- Category: Inputs and selection
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `FieldGroup`
- Types: `FieldGroupOptions`
- Theme key: `nui.ComponentFieldGroup` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Provide FieldGroupItems with stable IDs and Build callbacks. The group coordinates size, disabled state and outer corners while keeping each control’s keyboard order.

## Things to know

Search displays the submitted text without connecting to a search service. Each item needs a unique nonempty ID and a Build function. Size follows the group instance, group theme and enclosing field; explicit child sizes remain independent. Later-painted neighbors can cover the shared edge of a focus ring; MyGo has no public z-index.

## API

```go
func FieldGroup(c *ui.Context, t Theme, opts FieldGroupOptions, items []FieldGroupItem) ui.Element
```

### FieldGroupOptions

```go
type FieldGroupOptions struct {
	Size               Size
	Disabled, Vertical bool
	Decoration
}
```

### Related types

```go
type FieldGroupItem struct {
	ID    string
	Build func(*ui.Context, Theme) ui.Element
}
```

Types documented in other references: `Decoration` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Inputs, selection, and fields

#### State, themes, and events › Sizes, variants, and disabled state (excerpt)

Size precedence is: the instance's Size → the FormField/FieldGroup Size → the component's default Size. Set component defaults with calls such as `WithComponent(ComponentInput, ...)`. When a parent field or any ancestor is disabled, a child cannot re-enable itself with `Disabled: false`.

#### FormField and FieldGroup

##### FieldGroup: join several controls

Use FieldGroup to join an input and a button into one continuous row:

```go
nui.FieldGroup(c, theme, nui.FieldGroupOptions{Size: nui.MD}, []nui.FieldGroupItem{
    {ID: "name", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Input(c, inner, &app.Name, nui.InputOptions{Label: "Name"}).Grow(1)
    }},
    {ID: "check", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Button(c, inner, "Check", nui.ButtonOptions{
            Variant: nui.Outline,
        }).OnClick(app.Check)
    }},
}).FillWidth()
```

Each item's ID must be unique and non-empty, and Build must not be nil; otherwise FieldGroup panics. Build creates the current content synchronously.

FieldGroup rounds the outer corners of Input, Textarea, InputMenu, InputTags, Select, SelectMultiple, SelectMenu, Button, and Badge. By default, adjacent borders overlap by 1 DIP and the items stretch to the same height. The outer corners follow the writing direction. Vertical stacks the items vertically.

Disabled is passed down to the items. Size comes from the instance first, then the ComponentFieldGroup default, then the enclosing FormField; if none is set, each child uses its own default size. Popups opened by child controls keep their own corner radius, such as SelectMenu's search field and InputDate's calendar.

The Root part adjusts the spacing. Inside the group, each control keeps its own Tab order, and the group itself adds no focus stop. MyGo has no public z-index yet, so a neighboring control drawn later may cover the focus ring along a shared edge.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

When a Badge's label is empty, no text is created, no extra gap is left, and the padding becomes equal on all sides. Inside a FieldGroup or FormField, a Badge inherits the size and takes part in the group's corner rounding. If Avatar.Size is unset, badge sizes XS–XL use avatars of 16, 16, 16, 20, and 20 DIP. An explicit Leading icon takes precedence over the avatar.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type example struct {
	Value  string
	Result string
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	items := []nui.FieldGroupItem{{ID: "input", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
		return nui.Input(c, inner, &s.Value, nui.InputOptions{Label: "Search text", Placeholder: "Search"}).Grow(1).MinWidth(0)
	}}, {ID: "button", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
		return nui.Button(c, inner, "Search", nui.ButtonOptions{Variant: nui.Outline}).OnClick(func() { s.Result = "Submitted: " + s.Value })
	}}}

	nui.FieldGroup(c.Key("group"), t, nui.FieldGroupOptions{}, items).FillWidth()
	ui.Text(c, s.Result)
}

func main() {
	s := &example{}

	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "FieldGroup · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
