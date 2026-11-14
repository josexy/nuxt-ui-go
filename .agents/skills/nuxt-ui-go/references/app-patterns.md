# App patterns for NuxtUI Go

Cross-cutting patterns every screen needs. Component-specific details are in `components/<id>.md`.

Contents: 1 Window and startup · 2 Model, view and state ownership · 3 Identity and lists · 4 Callbacks versus builders · 5 Unset versus explicit values · 6 Background work · 7 Overlays and notifications · 8 Theming and styling · 9 Language

## 1. Window and startup

Add the library with `go get github.com/josexy/nuxt-ui-go/ui@latest` and build and run with `CGO_ENABLED=0`.

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go/ui"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

type app struct {
	config nui.AppConfig // created once; the view only reads it
	window *mygo.Window  // for Window.Update from background work
	Name   string
	Saved  int
}

func (a *app) view(c *ui.Context) {
	t := nui.App(c, a.config) // first, on every build
	nui.Container(c, t, nui.ContainerOptions{}, func() {
		nui.Input(c.Key("name"), t, &a.Name, nui.InputOptions{Label: "Name"}).FillWidth()
		nui.Button(c.Key("save"), t, "Save", nui.ButtonOptions{Color: nui.Primary, Leading: lucide.Check()}).
			OnClick(func() { a.Saved++ })
		ui.Textf(c, "Saved %d times", a.Saved)
	}).PaddingY(24)
}

func main() {
	a := &app{config: nui.DefaultApp()}
	mygo.App.SetLanguages(nui.SupportedLocales()...) // before any window
	mygo.App.WhenReady(func() {
		a.window = mygo.NewWindow(mygo.WindowOptions{Title: "Demo", Width: 640, Height: 400, Content: ui.View(a.view)})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- `nui.App(c, config)` applies the window theme and the system preferences (dark mode, text scale, contrast, reduced motion) and returns the `Theme` to pass to every component. Call it before any `nui.Window`/`nui.Themed` call in the same build. Never assign its result back to `config.Light` or `config.Dark`, or text scaling is applied twice.
- Build `config` (mode, palettes, radius, density) when you create the window. Rebuilding it in the view overwrites runtime changes. See `components/app.md` and `components/theme.md`.
- Layout uses MyGo containers: `ui.Column(c)`, `ui.Row(c)`, `.Gap`, `.Padding`, `.Fill`, `.FillWidth`, `ui.Scroll`. `nui.Container` centers content and caps its width. `nui.Main` is a flexible main area.

## 2. Model, view and state ownership

MyGo calls the view again after input, after `Window.Update`, and while something animates, so anything declared inside the view starts over on every build.

| Keep in the model (persistent) | Keep only within one build |
| --- | --- |
| Bound values (`string`, `bool`, `[]string`, IDs, `Optional[Date]`, …) | `ui.Element` values returned by components |
| Open flags (`bool`) and state objects (`nui.TableState[T]`, `nui.MenuState`, `nui.FormState`, `nui.SearchState`, `nui.SelectMenuState`, `nui.TreeState`, `nui.CommandPaletteState`, …) | Parts passed to `Style` callbacks |
| `ui.Handle` values for focus and geometry (`.Bind(&a.handle)`, `ReturnFocus: &a.handle`) | `*ui.Context` (`c`) |
| `c.Services()` captured for clipboard/URL/redraw outside the build | Builder closures |

- Keep each window's editing and interaction state separate, and give every component instance its own state object. Do not copy a state value after a component has used it; pass pointers.
- Components bind through pointers to model fields (`&a.Name`). Binding to a local variable loses edits, as does binding to the copy that `for _, item := range` makes.
- Elements may be used right away in the same build, for example `.FillWidth()` or passing an anchor element to `nui.Popover` or `nui.Tooltip`, but they expire afterwards.

## 3. Identity and lists

Call `c.Key(id)` immediately before creating each widget. MyGo uses it to keep focus, drafts, scroll position and animation on the right control across rebuilds. In dynamic lists, use a unique business ID, not an index, display position or editable label:

```go
type task struct {
	ID    string
	Title string
	Done  bool
}

func (a *todoApp) list(c *ui.Context, t nui.Theme) {
	for i := range a.Tasks {
		task := &a.Tasks[i] // pointer into the model, not a range copy
		nui.Checkbox(c.Key("task-"+task.ID), t, &task.Done, nui.CheckboxOptions{Label: task.Title})
	}
}
```

Components with item lists (Tabs, Accordion, Breadcrumb, Stepper, Timeline, menus, Select/Listbox options, Table rows, Tree nodes) need non-empty, unique item IDs or values, or they panic. Removing an item does not clean up the app's selection or open lists; the app decides when to do that.

## 4. Callbacks versus builders

| Runs once per user action, after the build: put side effects here | Runs on every build, possibly several times: build UI only |
| --- | --- |
| `.OnClick(fn)` on returned elements, `OnChange`, `OnSubmit`, `OnSelect`, `OnActivate`, `OnClose`, `OnNavigate`, `OnSortChange`, `OnSelectionChange`, … | Content/body callbacks (`func(nui.Theme)`, `func(*ui.Context, nui.Theme)`), `Header`, `Footer`, `Actions`, `Build`, `Cell`, `Style` |

- Change data, mutate collections, start requests and show toasts from callbacks. A builder that appends to the slice it is iterating, sends a request, or calls `nui.Toast` unconditionally repeats that work on every frame.
- Callbacks report user changes. When the app writes a bound value directly, the `OnChange`-style callbacks usually do not fire (see each component's "Things to know").
- Within one build, controls apply their input as they are constructed, so a statement after an `Input` sees the edited value. Bind straight to persistent fields instead of deriving values that you write back.

## 5. Unset versus explicit values

```go
nui.ModalOptions{
	Dismissible: nui.Set(false), // explicitly off; omitting it keeps the default (true)
	Width:       nui.Set[float32](480),
}
nui.PaginationOptions{SiblingCount: nui.Set(0)} // explicit 0, not "use the default"
```

`Optional[T]{}` (or omitting the field) means "use the component's default". Plain `bool`, number and enum fields are assigned directly, but a Go zero value is not always the effective default (for example, `PageSize: 0` means 10). The **API** and **Documentation excerpts** sections of each reference show the field types and defaults.

## 6. Background work

Run slow I/O and expensive computation off the UI thread. Write results into the model only inside `window.Update`, which runs on the UI thread and rebuilds the view. Cancel or ignore results that are no longer wanted:

```go
type searchApp struct {
	window  *mygo.Window
	Query   string
	Results []string
	Loading bool
	seq     int                // rejects stale results
	cancel  context.CancelFunc // cancels the running request
}

// Called from an action such as OnChange or a button's OnClick, never from the view body.
func (a *searchApp) search(query string) {
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.seq++
	seq := a.seq
	a.Loading = true
	go func() {
		results, err := fetch(ctx, query) // your I/O
		a.window.Update(func() {
			if seq != a.seq { // a newer search started
				return
			}
			a.Loading = false
			if err == nil {
				a.Results = results
			}
		})
	}()
}
```

- Library helpers do this bookkeeping. Form: `ticket := form.BeginAsync()`, a worker reads `ticket.Values`, then `form.CompleteAsync(ticket, errs)` inside `Window.Update` (`components/form.md`). Async option lists: `QueryState` (`components/select-menu.md`). Command search: `nui.NewCommandIndex` is immutable and safe to search on a worker (`components/command-palette.md`). Uploads: FileUpload tickets (`components/file-upload.md`).
- When a view unmounts or its window closes, stop pending work: `FormState.Cancel()`, `QueryState.Cancel()`, `CommandPaletteState.Close()`, `TreeState.Close()`, `FileUploadState.Dispose()`, and cancel your own contexts. Those calls make the state reject late results; they do not stop your I/O.
- Never capture `c`, elements or parts in a goroutine. To redraw from outside a build without changing data, use `c.Services().Invalidate()` captured during a build, or `window.Invalidate()`.

## 7. Overlays and notifications

```go
type dialogApp struct {
	Open    bool
	Draft   string
	trigger ui.Handle
}

func (a *dialogApp) view(c *ui.Context, t nui.Theme) {
	nui.Button(c.Key("edit"), t, "Edit", nui.ButtonOptions{}).Bind(&a.trigger).OnClick(func() { a.Open = true })
	// Called on every build, open or not; never wrapped in `if a.Open`.
	nui.ModalWithOptions(c.Key("edit-dialog"), t, &a.Open, nui.ModalOptions{
		Title:       "Edit project",
		ReturnFocus: &a.trigger,
		Footer: func(c *ui.Context, t nui.Theme) {
			nui.Button(c.Key("save"), t, "Save", nui.ButtonOptions{Color: nui.Primary}).OnClick(func() { a.Open = false })
		},
	}, func(t nui.Theme) {
		nui.Input(c.Key("draft"), t, &a.Draft, nui.InputOptions{Label: "Name"}).FillWidth()
	})
}
```

- The same pattern applies to Slideover, Drawer, Popover (anchor element plus `*bool`), DropdownMenu and ContextMenu (`*nui.MenuState`, one per trigger and window), and CommandPalette (`*nui.CommandPaletteState`).
- Toasts: put `nui.ToastViewport(c, t, nui.ToastViewportOptions{})` in the view once per window, then call `nui.Toast(c, t, nui.ToastOptions{...})` from an action. It returns the toast ID; `nui.CloseToast` removes it. See `components/toast.md`.
- Tooltip wraps an anchor element built in the same build. Keep tooltip text short and do not put the only copy of important information in it.

## 8. Theming and styling

Styles are applied in this order, later winning: built-in style → global theme (`config.Light`/`config.Dark` and their component settings) → local theme (`theme.WithComponent(...)` applied with `nui.Themed`) → instance options (`Color`, `Size`, `Variant`, `UI`) → the instance `Style` callback.

```go
// Instance: one part of one component.
nui.Button(c.Key("danger"), t, "Delete", nui.ButtonOptions{
	Color:   nui.Error,
	Variant: nui.Soft,
	UI:      nui.Styles{nui.Label: {Weight: nui.Set(600)}},
})

// Local: every Button inside this subtree.
local := t.WithComponent(nui.ComponentButton, nui.ComponentTheme{Size: nui.SM, Variant: nui.Outline})
nui.Themed(c, local, func(t nui.Theme) {
	nui.Button(c.Key("a"), t, "Small outline", nui.ButtonOptions{})
})
```

- Palettes: `nui.DefaultPalettes()`, then set `palettes.Primary = nui.BluePalette()` (or `nui.PaletteByName("blue")`) and regenerate both `config.Light = nui.NewTheme(false, palettes)` and `config.Dark = nui.NewTheme(true, palettes)`. `theme.WithBlackPrimary()` gives a black/white primary color.
- Each component supports only certain parts (`Slot` constants such as `nui.Root`, `nui.Label`, `nui.Panel`, `nui.Item`); parts that are not listed have no effect. The component reference's documentation excerpts list its parts.
- Use theme colors (`t.Primary`, `t.Muted`, `t.Elevated`, `t.Border`, …) rather than fixed colors so light, dark and high-contrast modes keep working. Sizes are DIP.

## 9. Language

- `mygo.App.SetLanguages(nui.SupportedLocales()...)` before creating windows, unless the app registers its own catalog. In that case, keep declaring the app's own languages.
- Switch languages from an action: `mygo.App.SetLocale("zh-Hans")`, `"en"`, or `""` to follow the system. All windows refresh, and drafts and selections are kept.
- Components translate only their built-in messages. The app translates the text it passes in, business errors, toast text and data.

