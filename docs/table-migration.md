# Table migration and native extensions

Table now exposes the native combined `OnChange func()` notification. `OnSelectionChange` has been removed. Cursor movement (for example Ctrl+Down), selected rows/cells, effective sorting and completed column-layout changes can all notify. Rebuilding and external assignments remain silent. `OnSortChange` still reports only effective user sort changes; `OnActivate` preserves typed IDs on Enter/double-click.

Before:

```go
nui.Table(c, t, &state, columns, nui.TableOptions{
    OnSelectionChange: func(ids []string) { useIDs(ids) },
})
```

After, when an action needs the current selected keys:

```go
nui.Table(c, t, &state, columns, nui.TableOptions{
    OnChange: func() { useIDs(slices.Sorted(state.Selection.All())) },
})
```

The explicit snapshot costs O(S) traversal/storage and sorting costs O(S log S). Prefer reading `Selection.Len()` or `Selection.Has(id)` when those answer the application question. Do not infer a selected-set change from OnChange, even if sort and column layout stayed the same. A strict selection-only observer must maintain and compare an application-owned snapshot at its chosen event/data-update boundaries, including external edits; that cost is explicit rather than imposed on every table build. No private native revision is accessed.

`TableColumn.RowHeader` defaults to false. With `state.List.CellSelect = true`, its cells select whole rows, its header selects the whole table, and it never becomes the cell cursor or a sortable data column. Even an explicit Sortable=true or Compare does not enable row-header sorting. Other columns retain normal sort behavior.

Configure native extensions on `state.List`: RowHeight (0 keeps native automatic height), CellSelect/Cells, HeaderCell, HeaderMenu, CellMenu, EndReached/NearEnd, FollowEnd and applicable drag settings. The wrapper does not overwrite them. Native restrictions remain: CellSelect/DragSelect cannot combine with row reordering or external row dragging; FollowAnimation/MoveAnimation follow native motion preferences, and a wrapper motion flag does not promise to override every raw native option. The wrapper owns Key/Selection/Sort bindings; custom selected-row pointers retain native cursor semantics.

Plain Value-only cells skip the extra content Row only when no Item/Label theme, table or column decoration and no Style/Cell/BuildCell callback applies. Adding any relevant override restores the decorated path. BuildCell still wins over Cell and Value, custom child identity stays on that path, and the fractional-scale header divider remains.

The table virtualizes rows, not columns. A 640-DIP fixture still builds 55/550/2200 cells for 5/50/200 columns across 11 visible/overscan rows. More columns remain more work; no horizontal-windowing claim is made. Local sorting is synchronous. For large data, use the complete [cancellable manual sorting example](../examples/table-manual-sorting/main.go): one worker, one latest pending request, immutable snapshots, cancellation during copying/merging, generation validation and Window.Update publication. It preserves input order and typed IDs, supports multiple sort keys, and cancels on window closure. Run with `CGO_ENABLED=0 go run ./examples/table-manual-sorting`. ManualSorting consumes the published order without repeating a local sort.
