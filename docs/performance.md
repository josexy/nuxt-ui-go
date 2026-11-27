# Performance evidence

For public API changes, see the [choice migration](choice-migration.md)
and [Table migration](table-migration.md) guides. The [FileUpload guidance](advanced.md#fileupload-and-resource-ownership)
documents its new bounded preview layout and explicit zero height, and
[motion guidance](foundations.md) explains color transitions and native scale poses.
InputTags/Tabs/Accordion/CheckboxGroup/Form retain their nonvirtual layouts and
existing public contracts.

Set `CGO_ENABLED=0` (PowerShell: `$env:CGO_ENABLED='0'`) and run the native Go
benchmark command from the repository root:

```sh
go test ./ui -run '^$' -bench '^(BenchmarkDataComponents|BenchmarkDataAnimations|BenchmarkPerformance)' -benchmem -benchtime=100ms -count=5 -cpu=1
```

Save raw output under `.work/` together with the command, toolchain, dependency,
Git revision and working-tree changes or a source snapshot. Do not edit sources
or run other verification jobs while measuring. Keep the machine, power mode,
CPU setting, viewport, motion settings and background load the same between runs.
Record the environment and calculate statistical summaries alongside the raw output.

For a quick fixture check, use `-benchtime=1x -count=1`. This checks execution,
not comparative performance. Narrow a repeat with `-bench '^BenchmarkPerformanceTable$'`.
Compare five-sample medians and dispersion, inspect the raw samples, and repeat
unexplained changes under the same conditions.

| Workload | Fixture and interpretation |
| --- | --- |
| Settled default rendering | `BenchmarkDataComponents`: motion Off before mounting, fixed Tester clock, one view build per operation asserted; 100/10k/100k datasets allocated separately for each component |
| Selection size | `BenchmarkPerformanceSelection`: S=0/1/N, open and closed multi-selects and an always-visible Listbox; a Style counter records constructed rows without replacing their content |
| Initial data preparation | `BenchmarkPerformanceSubmission`: new public Select/InputMenu/Listbox/SelectMenu/Menu states and SelectionState submissions; ownership, validation and metadata are timed, with source generation excluded |
| Search | `BenchmarkPerformanceQuery`: first, alternating changed, and identical cached queries over the same submitted items; `last-results` is the final query's result count, alongside time/allocation |
| Pending query dispatch | `BenchmarkPerformanceSearchDispatch`: a held Background worker isolates the UI's query-change/build cost at 100/10k/100k candidates; one active job and one latest request. This is dispatch cost, not completed-search latency |
| Popup sizing | `BenchmarkPerformancePopupHeight`: unchanged 240-DIP cap at 100/10k/100k candidates; stops scanning once the viewport fills |
| Bulk user action | `BenchmarkPerformanceBulkSelection`: native Ctrl+A, acceptance check, then external clear and rebuild; this whole cycle is measured separately from steady builds |
| Tree shape | `BenchmarkPerformanceTree`: N flat nodes, N folded children, or N open parents with one child each; selected count and open-branch count are explicit, including linked selection |
| Table width | `BenchmarkPerformanceTable`: 100k rows, 5/50/200 columns, empty/all selection, with/without the native combined OnChange callback; actual cells and rows built are reported |
| Cursor restoration | `BenchmarkPerformanceTableCursor`: mounted same-index restoration and complete owned/borrowed rebinding, with stable typed IDs |
| Rich menus | `BenchmarkPerformanceMenu`: 100/10k/100k total items spread over one or three open levels, plus closed controls |
| File admission | `BenchmarkPerformanceUploadAdmission`: existing plus incoming batches at N/2N, with duplicate paths, rejected entries and count limits; includes creating the owned existing working copy |
| File previews | `BenchmarkPerformanceUploadPreviews`: measured List and fixed GridView viewports at 100/10k/100k files; rows/op counts the constructed removal controls in the settled tree, without a Style callback that would request measured fallback |
| Other collections | `BenchmarkPerformanceBookkeeping`: InputTags, Accordion, Tabs, CheckboxGroup and Form at 100/1k/10k; their nonvirtual child construction remains part of the cost |
| Tag batches / command text | `BenchmarkPerformanceTagBatch` measures ordered deduplication; `BenchmarkPerformanceCommandText` separates plain and Unicode highlighted text at 100/1k/10k runes |
| Motion | `BenchmarkDataAnimations`: Full and Reduced popup open/close cycles; input builds, later tick builds, and 600 ms of simulated time are separate metrics |

`ns/op`, `B/op`, and `allocs/op` are Go benchmark measurements. Fixture setup and
the first layout are excluded from steady timings; data preparation has its own
cases. Instrumented row-count fixtures should be compared to the same fixtures,
not to custom-content or default fixtures with different work. Closed controls
legitimately report zero rows. Rich menus and file previews construct visible
rows/cells and native overscan. The native Table still builds all columns of
visible rows. InputTags, Accordion, Tabs, CheckboxGroup and Form remain nonvirtual;
removing nested membership scans does not make their complete layouts constant-time.

Tester freezes its clock until `Advance`. `Frame` forces a build and may settle
up to twenty builds at that timestamp. Reduced motion still permits fades, so it
does not settle an entering popup. Steady fixtures turn component animations Off
before creation and assert one build/op. Animation fixtures advance time and
report all builds. Public Tester does not expose a paint counter; these results
are neither native frame counts nor GPU FPS.

Run the focused correctness regressions with
`CGO_ENABLED=0 go test ./ui -run '^Test(InputLimitCaretAndSelectedReplacement|TreeSummaryModesWithCustomParts|TableCursorChangeKeepsSelection|NativeTransitionVsPoseScheduling)$' -v -count=1`
(PowerShell: set `$env:CGO_ENABLED='0'` first). They preserve the following evidence:

- A rejected middle insertion at `ab|cd` leaves the next Backspace producing
  `acd` in both Input and Textarea, preserving the native caret and undo history.
- Linked Tree mixed state must clear before custom content on mode
  exit and recompute from current selection on re-entry. Narrowing the summary
  condition alone breaks this regression.
- Native Ctrl+Down changes the bound cursor and combined notification
  once while the chosen-key set remains unchanged.
- ElementTransition entry causes later view builds; AnimatePose ticks
  do not. This does not validate popup/sheet presence semantics or authorize
  replacing their animation without the separate focus/interruption regressions.

Record growth/shrink retention separately from allocation per operation: retain
the live state after submission, force GC before each heap snapshot, grow to the
largest dataset, replace it with a small/empty dataset, force GC again, and keep
source arrays out of the retained model. Report HeapAlloc/HeapObjects and setup
cost independently; allocation totals are not retained heap. Background search
also needs input-to-present and query-to-result latency, cancellation, held-worker,
unmount and shutdown evidence; synchronous search timings do not prove worker
responsiveness.

The optional js/wasm regression `TestChoiceSearchWorkerYieldsToJavaScript` can
run through Go's `wasm_exec_node.js` using Node.js with a minimal environment.
A JavaScript timer must cancel a background
normalization job before it finishes. Browser goroutines share the JavaScript
event loop, so background choice search yields between time-bounded batches;
Immediate mode stays synchronous. This checks event-loop cooperation, not browser
rendering latency or native FPS.

Local acceptance compares fixed-viewport choice growth from 100 to 100k
candidates against budgets of 2x median time, 16 KiB/op and 32 allocs/op;
empty/all-selected Table builds use analogous comparisons with fixed columns. Submission, searching,
bulk selection, explicit value snapshots, and wide columns are separate costs.
These are local engineering gates, not portable frame-rate promises or measured
speedup claims. Retain the before source archives and inspect the final evidence
report rather than inferring improvements from a single timing or widget count.

## Host-event and native verification

An idle `Advance` after `Tester.Move` proves only that no later build was needed.
`Move` calls the ordinary Tester `send` path, which unconditionally settles a
build after delivering the event. It cannot prove zero builds for the initial
host pointer event. The pinned dependency has `HostMove` and `Paints` helpers in
`ui/export_test.go`, used by `TestHoverReadAsBuiltBuilds` and
`TestHoverLookPaintsAsBuilt`. They are test-only exports, unavailable to importing
libraries. The upstream harness delivers the host event and counts paints and
view builds separately. Its native button is a paint-only control; its box that
reads `Hovered` during construction is the positive build control.

Resolve the dependency directory from `go mod download -json github.com/egoist/mygo`
(`Dir`), then run the following there with CGO_ENABLED=0:

```text
go test ./ui -run '^TestHover(ReadAsBuiltBuilds|LookPaintsAsBuilt)$' -v -count=1
```

Both tests passed on the Windows amd64 baseline environment. This is an
instrumented upstream host-event check, not an interactive desktop check of
NuxtUI Go controls. Our transition/pose probe separately records the initial
forced operation and subsequent timed builds. The public dependency API cannot
expose host-event paint counts to our fixtures, and no copied internal helper or
private-field access is used to bypass that boundary.

NuxtUI Go resolves hover targets during builds and fades surface colors under Full and Reduced motion; Off or a nonpositive Motion.Transition shows them immediately. Color animation can rebuild the view on later ticks; record those builds separately from settled rendering. Zero-build hover is deferred until a safe native paint-state API supports retained exit copies, with exit/unmount/remount tests and actual host-event instrumentation. A Valid guard, private-field access, or material retaining an Element does not solve that ownership problem. User and InputRating scale poses are separately verified to paint without tick builds.

`MYGO_FRAME_STATS=all` logs native frame phases and marks frames with zero build
passes as `painted again`; a rounded `build 0.0` duration alone is not proof of
zero builds. Keep the logs under `.work/`, and leave `MYGO_UNCAPPED` unset when
checking ordinary host scheduling. Capture/forced-frame helpers, a hidden
first-frame smoke test, and the Gallery socket's input actions do not alone
establish real OS pointer delivery. Validate changed controls on the native path
and explicitly record any desktop/backend that was not exercised.
