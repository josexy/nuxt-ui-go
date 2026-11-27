# Migrating submitted choices

Choice controls now keep owned data and revisioned selection in persistent state. Submit data during initialization or an accepted data update. The view only renders that state. Keep separate states for each control and window; do not copy a state after use or retain a build Context in a setter callback.

## Before and after

The following **before** fragments describe the removed API; they are not current examples:

```go
// Before: slices passed on every build, selected IDs in a writable slice.
nui.Select(c.Key("single"), theme, &app.Value, options, nui.SelectOptions{})
nui.SelectMultiple(c.Key("multiple"), theme, &app.Values, options, nui.SelectOptions{})
nui.Listbox(c.Key("list"), theme, &app.IDs, &app.List, items,
    nui.ListboxOptions{OnChange: func(ids []string) { app.Save(ids) }})
nui.InputMenu(c.Key("search"), theme, &app.Search, options,
    nui.InputMenuOptions{Multiple: true, OnValuesChange: func(ids []string) { app.Save(ids) }})
nui.DropdownMenu(c.Key("actions"), theme, anchor, &app.Menu, commands, nui.MenuOptions{})
```

**After**, initialize persistent state outside the view:

```go
type Choices struct {
    Value    int
    Single   nui.SelectState[int]
    Multiple nui.SelectState[int]
    Values   nui.SelectionState[int]
    IDs      nui.SelectionState[string]
    List     nui.ListboxState
    Popup    nui.SelectMenuState
    Search   nui.SearchState
    Menu     nui.MenuState
}

func newChoices() *Choices {
    app := &Choices{}
    options := []nui.Option[int]{{Value: 0, Label: "All"}, {Value: 7, Label: "Active"}}
    app.Single.SetOptions(options)
    app.Multiple.SetOptions(options)
    app.Values.Set([]int{0})
    items := []nui.ListboxItem{{ID: "go", Label: "Go"}, {ID: "rust", Label: "Rust"}}
    app.List.SetItems(items)
    app.Popup.SetItems(items)
    app.Search.SetOptions([]nui.Option[string]{{Value: "go", Label: "Go"}})
    app.Search.SelectedValues.Set([]string{"go"})
    app.Menu.SetItems([]nui.MenuItem{{ID: "clear", Label: "Clear", OnSelect: func() { app.IDs.Clear() }}})
    return app
}
```

The view uses the submitted state. A snapshot is taken only where the app needs all selected keys:

```go
nui.Select(c.Key("single"), theme, &app.Value, &app.Single, nui.SelectOptions{})
nui.SelectMultiple(c.Key("multiple"), theme, &app.Values, &app.Multiple, nui.SelectOptions{})
nui.Listbox(c.Key("list"), theme, &app.IDs, &app.List,
    nui.ListboxOptions{OnChange: func() { save(app.IDs.Values()) }})
nui.SelectMenu(c.Key("popup"), theme, &app.IDs, &app.Popup, nui.SelectMenuOptions{})
nui.InputMenu(c.Key("search"), theme, &app.Search,
    nui.InputMenuOptions{Multiple: true, OnValuesChange: func() { save(app.Search.SelectedValues.Values()) }})
nui.DropdownMenu(c.Key("actions"), theme, anchor, &app.Menu, nui.MenuOptions{})
```

`ContextMenu` uses the same `MenuState.SetItems` pattern. `SearchSelect(c, theme, &app.Search, label)` also reads options from SearchState. A single Select still binds `*T`; `0`, `false`, and empty strings are valid typed Select values. InputMenu, Listbox, SelectMenu, and rich menus require nonempty string IDs. Label/separator rows also require IDs. All submitted IDs must be unique; typed values must equal themselves, so NaN is rejected.

## Ownership and notifications

| Operation | Contract |
| --- | --- |
| `SetOptions` / `SetItems` | Own a snapshot and validate before publication. Invalid data panics without replacing the previous dataset. Every successful call advances the data revision, even for equivalent descriptors. |
| Later source edits | Have no effect until resubmitted, including changes to labels, disabled flags, callbacks, and styles. Nested menu children, mutable descriptor structs, and UI maps are copied. Callbacks, immutable assets, and explicitly app-owned pointers remain references. |
| `SelectionState.Set(values)` | Owns the input, removes duplicates in first-occurrence order, and returns whether the ordered selection changed. Nil and empty clear it. Invalid nonreflexive keys leave the old state intact. |
| `Add` / `Remove` / `Clear` | Return whether selection changed. An effective change advances `Revision`; reapplying the same ordered selection does not. `Remove` preserves survivor order. |
| `Has` / `Len` / `Revision` | Read membership, count, or revision without copying or enumerating the selected keys. |
| `Values()` | Returns an independent ordered snapshot. Editing that snapshot cannot update the state. |
| External setters and selection mutations | Are silent: they never emit component change, query, or activation callbacks. Update mounted state on the UI thread. |
| Accepted selection action | Parameterless `SelectOptions.OnChange`, `ListboxOptions.OnChange` (also embedded in SelectMenu), and `InputMenuOptions.OnValuesChange` notify once for an accepted effective selection change. No automatic selection snapshot is made. |
| Activation and query | Listbox/SelectMenu `OnSelect(id)` still reports accepted activation, including an unchanged selected ID. InputMenu `OnSelect(id)` is separate from `OnValuesChange` and the editor's query `OnChange`. Disabled or cancelled actions do not emit successful selection notifications. Creation requests do not insert a selected value for the app. |

Do not capture an Element or build Context in a submitted callback. Capture application state or long-lived services instead. When a callback itself changes, resubmit its descriptor. For remote data, validate the request ticket on the UI thread before calling the setter.

An empty data update preserves selection and the query draft. Clear them separately when the app intends to do so:

```go
// Run in an action or an accepted UI-thread data update, not every view build.
app.Single.SetOptions(nil)
app.List.SetItems(nil)
app.Popup.SetItems(nil)
app.Search.SetOptions(nil)
app.Menu.SetItems(nil)
// app.IDs and app.Search.Query are still retained.
app.IDs.Clear() // explicit, silent selection reset
app.Search.SelectedValues.Clear()
app.Search.Query = ""
```

Temporarily missing selected values stay selected; multiple-choice displays retain their last known labels. InputMenu reconciles changed data by dismissing an open suggestion panel; data submission does not promise to keep a removed row or submenu open.

## Layout and cost

Multi-select InputMenu uses a scrollable selected-tag viewport, capped at **96 text-scaled DIP** and shrinking for small selections. Labels wrap in full; the draft editor stays visible below the tags. Arrow keys, Home/End, and Tab navigate logical tags and scroll before focusing an offscreen removal control. Shift+Tab from the editor reaches the last tag. Removing a tag returns focus to the editor and preserves its draft. When the draft is empty, Backspace removes the final value. Disabled and read-only states prevent removal.

Let **N** be submitted candidates, **S** selected keys, and **V** visible rows plus bounded overscan. Unchanged revision checks do not enumerate N or S, and virtual choice rows/tags are built in proportion to V. Submission and full filtering remain O(N); selection-label reconciliation and `Values()` can cost O(S); preserving order during `Remove` can cost O(S). Select All and other bulk changes are data-dependent operations. Calling setters or `Values()` unconditionally in a view reintroduces those costs. Rich menu open-level virtualization is documented with the menu implementation; data ownership alone does not imply bounded rendering for every component.

## Local search execution and cleanup

Listbox, SelectMenu, InputMenu, DropdownMenu and ContextMenu expose `SearchExecution`. The default `SearchAuto` runs locally on the UI thread for at most **2,048 candidates**, and uses background work above that count. `SearchImmediate` always runs synchronously; `SearchBackground` always uses a worker for a nonempty local query. Menu sublevels inherit the root policy unless `MenuContentOptions.SearchExecution` supplies an explicit Optional override. Empty queries, `IgnoreFilter`, and `ExternalFilter` bypass local matching.

```go
// Initialize once, outside the view.
app.Popup.SetItems(items)
// In the view; Auto is also the zero/default policy.
nui.SelectMenu(c.Key("popup"), theme, &app.IDs, &app.Popup,
    nui.SelectMenuOptions{SearchExecution: nui.SearchAuto})
// During owner disposal, cancel requests while retaining data and selection.
app.Popup.Close()
```

The first normalization and candidate scan both run in the worker in background mode. Native builds use a background goroutine. On the browser backend, Go shares the JavaScript event loop: background jobs cooperatively yield between time-bounded batches, allowing input and cancellation; this is not a dedicated Web Worker. Immediate mode intentionally stays synchronous. Normalized fields are reused until submitted data or searchable fields change; results carry immutable source indices and prepared row metadata. Each mounted search surface, including each filtered rich-menu level, has at most one active worker and one latest pending request. Query, data, field or policy changes cancel obsolete work. Results publish on the UI thread only for the current request and mount. They never call app selection callbacks from a worker.

`Searching()` reports local pending work. While pending, a loading status replaces actionable results; the query stays editable, and Enter, Space, pointer actions and creation cannot accept an obsolete result. Closing, unmounting, `Close()`, and window shutdown cancel work and outstanding focus requests. Reopening can search the retained data again. If InputMenu binds an external Open pointer, its owner sets that pointer false as well when disposing. Remote application requests still need their own cancellation, for example QueryState.Cancel.

Listbox/SelectMenu keep case/accent folding and exact/prefix/substring ordering within structural groups. InputMenu and rich menus keep case-insensitive substring matching in source order without accent folding. All local choice searches distinguish nil fields (default label) from an explicit empty list (no match for a nonempty query). For rich menus this replaces the old empty-list fallback to label; omit FilterFields to keep the default. Search initialization and each full query still depend on N; submission and explicit bulk/snapshot work retain their separate costs. `SearchImmediate` does not promise responsiveness for large datasets. Creation-row composition and app-provided builders can add their own work; a background matching policy does not move app callbacks or arbitrary UI construction off the UI thread.
