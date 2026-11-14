<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# User

Show a user’s avatar, name and short description.

- Category: Themes and basic components
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `User`
- Types: `UserOptions`
- Theme key: `nui.ComponentUser` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Name and Description present user information; Avatar configures the image or fallback. Empty Avatar options show no avatar. Chip appears only with the default avatar, follows User size and stays inset. Leading replaces the avatar area; Body replaces the default name and description.

OnActivate makes User a native button for pointer, Enter and Space activation. Hover or keyboard focus emphasizes text and scales the avatar; reduced motion removes the scale transition. Handle navigation in the application callback.

## Things to know

User.Size overrides Avatar.Size and Chip.Size; Chip.Inset is always true. The chip does not intercept clicks; include meaningful status in the name, description or accessible name. Interactive custom content should still provide Name. Native User has no Nuxt routing, target or automatic URL opening.

## API

```go
func User(c *ui.Context, t Theme, opts UserOptions) ui.Element
```

### UserOptions

```go
type UserOptions struct {
	Name, Description  string
	Avatar             AvatarOptions
	Chip               *ChipOptions
	Size               Size
	Orientation        Orientation
	Vertical, Disabled bool
	OnActivate         func()
	Leading, Body      func()
	Decoration
}
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `ChipOptions` → [chip.md](chip.md), `Decoration` → [theme.md](theme.md), `Orientation` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Themes and basic components

#### Components and parts › Displaying information (excerpt)

| Component | Purpose and main options | Parts |
| --- | --- | --- |
| User | An avatar, badge chip, name, and description. Supports Orientation/Vertical, Disabled, and OnActivate. Long text is truncated with an ellipsis | Root, Leading, Body, Label, Description |

If every field of User's `AvatarOptions` is empty, no avatar is shown. To show initials, set `Avatar: nui.AvatarOptions{Alt: name}` or `Text`. At size XS, the description appears on the same line as the name. User acts as a keyboard-activatable button only when you pass `OnActivate`.

User.Size sets the size of the avatar and its chip, overriding the nested Avatar.Size and Chip.Size. Chip.Inset is always true; the other chip settings (position, color, text, and Hidden) are passed through. Without an avatar, no chip is shown. Leading replaces the default avatar area and Body replaces the name and description. Each region is built with a stable identity.

An activatable User emphasizes its text and scales the avatar to 1.15× on hover or keyboard focus. The scale transition follows Motion.Transition, and with reduced motion the final state appears immediately. Disabling the User or a parent prevents activation. OnActivate does not route or open URLs; your app implements those actions.

#### Components and parts › Images, avatars, and icons (excerpt)

`Text` is shown as is, for example `+12`. Without Text, Avatar takes the first character of each of the first two space-separated words in Alt, preserving case; without Alt, it stays empty. The text shrinks to fit the circle, but a full Unicode grapheme-aware abbreviation algorithm is not implemented. User's Name is the avatar's default Alt.

### Advanced interaction, charts, and layout

#### NavigationMenu and Sidebar › What navigation and layout each handle (excerpt)

NavigationState.Expanded exposes the vertical expansion state, and DefaultOpen only initializes new nodes. Single=true lets only one branch per level be open; with Single, Collapsible=false prevents collapsing the current branch. User changes fire OnExpandedChange; initialization and app changes are silent. When a branch collapses, focus on a hidden child moves to its parent. Removing a root item moves focus to an available root item, and switching Compact keeps the expansion and the active ID.

### Inputs, selection, and fields

#### Slider (excerpt)

User actions snap to multiples of Step from Min, and floating-point noise is removed based on the decimal precision of Min and Step. For example, adding 0.1 to 0.2 shows 0.3. Disabling the slider or a parent cancels any drag in progress, and OnCommit does not fire.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
)

type example struct{ Activated int }

func (s *example) view(c *ui.Context, t nui.Theme) {
	avatar := nui.AvatarOptions{}

	avatar.Alt = "Ada Lovelace"

	nui.User(c.Key("user"), t, nui.UserOptions{Name: "Ada Lovelace", Description: "Maintainer", Avatar: avatar})
	ui.Textf(c, "Activated: %d", s.Activated)
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
		mygo.NewWindow(mygo.WindowOptions{Title: "User · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
