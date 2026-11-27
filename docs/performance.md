# Performance benchmarks

The `ui` package has benchmarks that measure how much work components do per build with small and large data sets. Use them to compare a change against its base revision on the same machine.

## Run the benchmarks

Run from the repository root with `CGO_ENABLED=0` (PowerShell: `$env:CGO_ENABLED='0'`):

```sh
go test ./ui -run '^$' -bench '^(BenchmarkDataComponents|BenchmarkDataAnimations|BenchmarkPerformance)' -benchmem -benchtime=100ms -count=5 -cpu=1
```

To check only that every benchmark runs, use `-benchtime=1x -count=1`. To repeat one benchmark, narrow `-bench`, for example `-bench '^BenchmarkPerformanceTable$'`.

Compare the median and spread of the five samples, not a single run. Keep the machine, power mode, and background load the same for both revisions, and do not run other jobs while measuring. Save the raw output under `.work/` together with the command, Go version, and Git revision.

## What each benchmark measures

| Benchmark | Measures |
| --- | --- |
| `BenchmarkDataComponents` | One settled build of each data component with 100, 10k, and 100k items, animations off |
| `BenchmarkDataAnimations` | Popup open and close cycles at Full and Reduced motion |
| `BenchmarkPerformanceSelection` | Builds with no, one, or every item selected, for open and closed multi-selects and a Listbox |
| `BenchmarkPerformanceSubmission` | `SetOptions`, `SetItems`, and `SelectionState.Set` on new states |
| `BenchmarkPerformanceQuery` | A first, a changed, and a repeated query over the same items |
| `BenchmarkPerformanceSearchDispatch` | The UI cost of starting a background search while the worker is busy; not the time to a result |
| `BenchmarkPerformancePopupHeight` | Sizing a popup over 100 to 100k items |
| `BenchmarkPerformanceBulkSelection` | Ctrl+A, then clearing the selection from code |
| `BenchmarkPerformanceTree`, `BenchmarkPerformanceTreeFoldedSelection` | Flat, folded, and expanded trees, with and without linked selection |
| `BenchmarkPerformanceTable`, `BenchmarkPerformanceTableCursor` | 100k rows with 5, 50, or 200 columns; restoring the cursor after new rows |
| `BenchmarkPerformanceMenu` | Rich menus with 100 to 100k items over one or three open levels |
| `BenchmarkPerformanceUploadAdmission`, `BenchmarkPerformanceUploadPreviews` | Adding files with duplicates and limits; building the List and Grid previews |
| `BenchmarkPerformanceBookkeeping` | InputTags, Accordion, Tabs, CheckboxGroup, and Form, which are not virtualized |
| `BenchmarkPerformanceTagBatch`, `BenchmarkPerformanceCommandText` | Adding many tags at once; highlighting CommandPalette matches |

Some benchmarks also report custom metrics, such as the rows or cells built per operation. Compare those only with the same benchmark: closed controls build no rows, and instrumented fixtures do more work than default ones.

## Reading the results

`ns/op`, `B/op`, and `allocs/op` are Go benchmark measurements, not frame rates. Fixture setup and the first layout are not timed. `ui.Tester` keeps its clock still until `Advance`, so the steady-state benchmarks turn animations off and check that each operation takes exactly one build. Allocation per operation is not retained memory; measure retained heap separately with a forced GC before each snapshot.

Color transitions rebuild the view on each animation frame. The User and InputRating hover scale only repaints, so it costs no builds after the first one.

## Background search in the browser

On js/wasm, background search shares the JavaScript event loop and yields between short slices of work. This test checks that a JavaScript timer can cancel a running search; it needs Node.js:

```sh
GOOS=js GOARCH=wasm go test ./ui -run '^TestChoiceSearchWorkerYieldsToJavaScript$' -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec"
```

## Native frames

`ui.Tester` does not count paints or native frames. To see them, run a native app with `MYGO_FRAME_STATS=all`: it logs each frame's phases and marks frames that only repaint as `painted again`. Keep `MYGO_UNCAPPED` unset, save the log under `.work/`, and record which desktop backends you did not check.
