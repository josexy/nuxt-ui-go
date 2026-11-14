<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->

# CommandPalette

Search grouped commands, enter nested actions, and run a result.

- Category: Advanced interaction and layout
- Version: nuxt-ui-go v0.0.2 plus unreleased changes (MyGo v0.9.0)
- Functions: `CommandPalette`
- Types: `CommandPaletteOptions`, `CommandAction`, `CommandMatch`, `CommandPaletteState`, `CommandEvent`, `CommandIndex`
- Theme key: `nui.ComponentCommandPalette` (for `Theme.WithComponent`)
- Sections: Usage · Things to know · API · Documentation excerpts · Complete example

## Usage

Build an immutable index with NewCommandIndex and assign it to persistent CommandPaletteState. Call SetIndex when data changes. Keep rendering CommandPalette while closed so shortcuts and return focus work. OnSelect receives executed IDs; OnSelectEvent can cancel first.

## Things to know

This native command launcher closes after execution by default and always virtualizes results. Embedded Vue multiple selection and routing props are not provided. Built-in search covers label, description, suffix and keywords, with up to 200 results. FilterFields/IgnoreFilter affect only built-in search; custom Search handles filtering. SearchDelay delays query edits while canceling old requests immediately. Call Refresh after replacing a custom Search callback, and Close on unmount or window close. The searching message and errors are polite statuses, read once when they appear and again when their text changes; clearing or an unchanged rebuild reads nothing.

## API

```go
func CommandPalette(c *ui.Context, t Theme, s *CommandPaletteState, o CommandPaletteOptions) ui.Element
```

### CommandPaletteOptions

```go
type CommandPaletteOptions struct {
	Title, Placeholder                                      string
	Limit                                                   int
	ShortcutKey                                             ui.Key
	ShortcutMods                                            ui.Modifiers
	DisableShortcut                                         bool
	Disabled                                                bool
	Size                                                    Size
	SearchDelay                                             time.Duration
	FilterFields                                            []string
	IgnoreFilter, PreserveGroupOrder                        bool
	Autofocus, SearchInput, Back                            Optional[bool]
	SearchInputOptions                                      *InputOptions
	Icon, ChildrenIcon, BackIcon, LoadingIcon, SelectedIcon *ui.SVG
	ShowIcon                                                Optional[bool]
	Loading, Loop                                           bool
	HighlightOnHover                                        Optional[bool]
	BackOptions                                             *ButtonOptions
	EmptyText                                               string
	BuildEmpty                                              func(*ui.Context, Theme, string)
	OnSearchTermChange, OnHighlight                         func(string)
	OnSelectEvent                                           func(*CommandEvent)
	CloseOnSelect                                           Optional[bool]
	// Search starts application work; publish Results.Complete on the UI thread.
	// Nil uses the immutable index on a worker and a bounded mailbox.
	Search   func(context.Context, QueryTicket, string, string)
	OnSelect func(string)
	Modal    ModalOptions
	Decoration
}
```

### CommandAction

```go
type CommandAction struct {
	Key                                               ui.Key
	Mods                                              ui.Modifiers
	ID, Label, Description, Group, Keywords, Shortcut string
	Prefix, Suffix, Placeholder                       string
	Kbds                                              []string
	Icon                                              *ui.SVG
	Avatar                                            *AvatarOptions
	Chip                                              *ChipOptions
	Loading, Active                                   bool
	Disabled                                          bool
	Children                                          []CommandAction
	OnSelect                                          func()
	// OnSelectEvent can cancel a leaf activation before closing or callbacks.
	OnSelectEvent func(*CommandEvent)
	Decoration
}
```

### CommandMatch

```go
type CommandMatch struct {
	Action                                  CommandAction
	Highlights                              []int
	DescriptionHighlights, SuffixHighlights []int
	Score                                   int
	// contains filtered or unexported fields
}
```

### CommandPaletteState

```go
type CommandPaletteState struct {
	Open    bool
	Query   string
	Results QueryState[CommandMatch]
	// contains filtered or unexported fields
}

func (s *CommandPaletteState) Back()

func (s *CommandPaletteState) Close()

// Refresh cancels pending work and searches again on the next build. Use it
// after replacing a custom Search closure without changing Query or SetIndex.
func (s *CommandPaletteState) Refresh()

// Searching includes the debounce interval before a worker request starts.
func (s *CommandPaletteState) Searching() bool

func (s *CommandPaletteState) SetIndex(index *CommandIndex)
```

### CommandEvent

```go
type CommandEvent struct {
	ID string
	// contains filtered or unexported fields
}

func (e *CommandEvent) PreventDefault()
```

### CommandIndex

```go
// CommandIndex is immutable after construction and safe to search on a worker.
// Rebuild it when actions change, outside rendering. Callbacks run only on activation.
type CommandIndex struct {
	// contains filtered or unexported fields
}

func NewCommandIndex(actions []CommandAction) *CommandIndex

// Search checks cancellation while scanning and retains at most limit matches.
// Highlight offsets are Unicode rune indexes, not UTF-8 byte offsets.
func (x *CommandIndex) Search(ctx context.Context, query, parent string, limit int) []CommandMatch
```

### Related types

```go
// QueryState is UI-thread-owned. Start cancels obsolete work; Complete rejects stale tickets.
// Workers must publish with Window.Update, never mutate this state directly.
type QueryTicket struct {
	Query string
	// contains filtered or unexported fields
}
```

```go
type QueryState[T any] struct {
	Items   []T
	Loading bool
	Error   string
	// contains filtered or unexported fields
}

func (s *QueryState[T]) Cancel()

func (s *QueryState[T]) Complete(ticket QueryTicket, items []T, err error) bool

func (s *QueryState[T]) Start(parent context.Context, query string) (context.Context, QueryTicket)
```

Types documented in other references: `AvatarOptions` → [avatar.md](avatar.md), `ButtonOptions` → [button.md](button.md), `ChipOptions` → [chip.md](chip.md), `Decoration` → [theme.md](theme.md), `InputOptions` → [input.md](input.md), `ModalOptions` → [modal.md](modal.md), `Optional` → [theme.md](theme.md), `Size` → [theme.md](theme.md), `Theme` → [theme.md](theme.md).

## Documentation excerpts

From the library documentation. Links point to other component references in this folder.

### Advanced interaction, charts, and layout

#### CommandPalette

##### Create commands and show the palette

CommandPalette provides a searchable list of commands and opens or closes with Ctrl+K by default. Create a CommandIndex first, then hand it to a CommandPaletteState that you keep between builds:

```go
// Run on initialization or when the commands change; do not create the index on every build.
state.Commands.SetIndex(nui.NewCommandIndex([]nui.CommandAction{
    {
        ID: "save",
        Label: "Save workspace",
        Group: "Workspace",
        Key: ui.KeyS,
        Mods: ui.Ctrl,
        Shortcut: "Ctrl S",
        OnSelect: save,
    },
}))
```

Always call the component in the view, including when it is closed:

```go
nui.CommandPalette(c, theme, &state.Commands, nui.CommandPaletteOptions{
    Title: "Workspace commands",
})
```

Clean up tasks when the window closes with `window.OnClosed(state.Commands.Close)`. If the component stops appearing in the view, call Close explicitly as well.

Every ID in the command tree must be non-empty and unique, and the tree can be at most 32 levels deep. NewCommandIndex copies the children, key caps, style maps, and search data, and organizes the hierarchy by parent action. Treat the returned Action metadata as read-only, and replace it with SetIndex when the data changes. Your app still manages icons, bitmap resources, and callbacks.

##### Shortcuts and search

| Option or behavior | Description |
| --- | --- |
| ShortcutKey / ShortcutMods | Change the shortcut that opens the palette; DisableShortcut turns it off |
| An action's Key / Mods | Register a shortcut that runs the action; Shortcut is for display only |
| Label / Description / Suffix / Keywords | Matched by default; FilterFields chooses which ones. Matching is case-insensitive and finds Unicode characters as a subsequence |
| Groups and child actions | Results from the same group are shown together. PreserveGroupOrder keeps the index's group order. Children can set a Placeholder, and Backspace on an empty query goes back |
| SearchDelay / IgnoreFilter | Default 0 / false. A delay postpones the query but cancels the old request immediately. IgnoreFilter skips only the built-in matching |
| Kbds / Prefix / Suffix / Avatar / Chip / Loading / Active | Command content and appearance. Kbds only display shortcuts, and Loading does not disable the command |
| Result count | Up to 200 results by default, configurable up to 1,000. The list is virtualized |

The parent scope where you call the component determines the shortcut's scope; your app resolves conflicts with its own shortcuts. Search favors contiguous matches and matches at word starts, and highlights by rune index. It does not handle pinyin, word segmentation, or accent folding.

In the input, Down moves into the list; then the arrow keys, Home, and End move, and Enter activates. If a search is still running or the results do not match the current query, Enter does not run a stale action. The description appears on the line below the label, at the same font size, in Muted. SearchInput=false hides the input but keeps the query. Size, SearchInputOptions, and BackOptions configure the size, the input, and the back button.

Before a leaf command runs, OnSelectEvent is called, with the per-item handler first; PreventDefault cancels the command. After it runs, the older Action.OnSelect and the root OnSelect are called, and the palette closes first by default. CloseOnSelect=false keeps it open so the command can run repeatedly. If a callback for the same click changes the query, index, level, or request, the old command does not continue. Disabled restricts input, results, and registered shortcuts together.

##### Background search and closing

The default search scans an immutable index in the background, keeps at most one pending result, and reads it during the UI build. It checks for cancellation every 128 items and keeps only a limited number of top-scoring results, so it never sorts every match. Stale tasks cannot write back.

The build after Open becomes false cancels the task, and Close cancels it immediately. The component has no long-running goroutines or timers, and its state can be used in only one window.

To use your own search service, provide `Search(ctx, ticket, query, parent)`. Your app starts the task and calls `state.Results.Complete(ticket, matches, err)` in `Window.Update`. It is also responsible for result IDs, parent relationships, and triggering redraws. FilterFields and IgnoreFilter only affect the built-in search. After replacing a Search closure, call Refresh. Searching() covers both the delay and the running request.

This component is a native modal command launcher. It does not provide Vue's embedded multi-select model, route fields, or Fuse-specific options.

The searching message and a result error are polite status regions with stable identities. Each is read once when it appears and again only when its text changes; clearing it, and rebuilding without a change, read nothing. The empty-results message is a status too, but it is not read when it first appears.

The query field is affected by the IME cancel issue, and the full screen reader experience has not been verified.

#### State and part index (excerpt)

| Component | Persistent state or data | Parts |
| --- | --- | --- |
| CommandPalette | CommandPaletteState, CommandIndex | Root (modal), Editor, BackSlot, Panel, List, Label, Item, Body, ItemLeadingSlot, ItemLabelSlot, ItemLabelPrefixSlot, ItemLabelSuffixSlot, Description, ItemTrailingSlot, Indicator, EmptySlot, LoadingSlot, LoadingIconSlot |

### Inputs, selection, and fields

#### Select and InputMenu › Option row appearance (excerpt)

Select, InputMenu, SelectMenu, Listbox, menus, and CommandPalette share the same option row style:

### Compatibility and known limitations

#### Canceling IME composition (excerpt)

Affected inputs: Input, Textarea, InputMenu, InputTags, PinInput in text mode, the Listbox filter field, SelectMenu, the HEX field inside ColorPicker, and CommandPalette. If your app edits important text, test this flow with the IMEs on your target systems.

#### Accessibility and motion (excerpt)

Since MyGo v0.9.0, `ui.RoleStatus` is a polite, atomic live region. The library's own progress, error, empty, and loading messages keep stable identities: such a status is read once when its text changes and is silent when it first appears, clears, or rebuilds unchanged. Alert and Banner take their identity from the key the app gives them. CommandPalette and FileUpload also announce their progress and error messages once when they appear. Chart announces keyboard selection once and never reads pointer hover. Alert and Banner read their new title and description once when either changes. These rules are checked headlessly with `Tester.UseScreenReader`; what VoiceOver, NVDA, Narrator, and Orca actually say has not been verified.

## Complete example

A runnable program. App state lives in `example`, outside the view, and the view only builds UI.

```go
package main

import (
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type example struct {
	Commands               nui.CommandPaletteState
	Index                  *nui.CommandIndex
	Opener                 ui.Handle
	Selected, Status       string
	Calls, Edits, Requests int
	Protect                bool
}

func (s *example) view(c *ui.Context, t nui.Theme) {
	s.Protect = true

	direction := ui.LeftToRight

	ui.Column(c.Key("direction")).Direction(direction).Children(func() {
		nui.Button(c.Key("open"), t, "Open commands", nui.ButtonOptions{}).Bind(&s.Opener).OnClick(func() { s.Commands.Open = true })
		nui.CommandPalette(c.Key("commands"), t, &s.Commands, nui.CommandPaletteOptions{Title: "Command palette", SearchDelay: time.Duration(0), CloseOnSelect: nui.Set(!false), SearchInput: nui.Set(true), ShowIcon: nui.Set(true), DisableShortcut: true,
			OnSelect: func(id string) { s.Selected = id; s.Calls++; s.Status = "Command completed" }, OnSearchTermChange: func(string) { s.Edits++ },
			Modal: nui.ModalOptions{ReturnFocus: &s.Opener, Footer: func(c *ui.Context, t nui.Theme) {
				ui.Textf(c, "Command: %s · Calls: %d · Edits: %d", s.Selected, s.Calls, s.Edits)
				ui.Textf(c, "%s · Requests: %d", s.Status, s.Requests)
			}}})
	})
	ui.Textf(c, "Command: %s · Calls: %d", s.Selected, s.Calls)
}

func main() {
	s := &example{}

	s.Index = nui.NewCommandIndex([]nui.CommandAction{
		{ID: "create", Label: "Create document", Prefix: "File", Suffix: "Local", Description: "Start a blank invoice", Group: "Files", Keywords: "new", Icon: lucide.FilePlus(), Kbds: []string{"Ctrl", "N"}},
		{ID: "appearance", Label: "Appearance options", Placeholder: "Choose a theme…", Group: "Preferences", Children: []nui.CommandAction{{ID: "light", Label: "Light workspace", Icon: lucide.Sun()}, {ID: "dark", Label: "Dark workspace", Icon: lucide.Moon()}}},
		{ID: "profile", Label: "Personal workspace", Description: "Your account and documents", Group: "Files", Avatar: &nui.AvatarOptions{Text: "LC", Alt: "Lin Chen"}, Decoration: nui.Decoration{UI: nui.Styles{nui.ItemLabelSlot: {Weight: nui.Set(600)}}}},
		{ID: "status", Label: "Service status", Group: "Status", Chip: &nui.ChipOptions{Color: nui.Success}},
		{ID: "sync", Label: "Background sync", Group: "Status", Loading: true, Disabled: true},
		{ID: "protected", Label: "Protected command", Group: "Actions", OnSelectEvent: func(e *nui.CommandEvent) {
			if s.Protect {
				e.PreventDefault()
				s.Status = "Activation prevented"
			}
		}},
	})
	s.Commands.SetIndex(s.Index)
	s.Status = "Ready"
	defer s.Commands.Close()
	palettes := nui.DefaultPalettes()
	palettes.Primary, _ = nui.PaletteByName("green")
	palettes.Neutral, _ = nui.PaletteByName("slate")
	config := nui.AppConfig{Mode: nui.LightMode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.App.SetLocale("en")
		mygo.NewWindow(mygo.WindowOptions{Title: "CommandPalette · NuxtUI Go", Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
			t := nui.App(c, config)
			ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```
