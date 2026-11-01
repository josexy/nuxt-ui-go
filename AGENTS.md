# AGENTS.md

This file provides repository-wide guidance for coding agents and applies to this directory and its descendants. Explicit task instructions take precedence. Follow more specific instructions in subdirectories within their scope.

## Project and entry points

NuxtUI Go is a native Go component library built on MyGo, with the visual style of Nuxt UI v4. Its module path is `github.com/josexy/nuxt-ui-go`, and its root package is `nui`. Use [go.mod](go.mod) as the source of truth for the Go toolchain requirement, MyGo version, and other dependencies. Build and run ordinary tests with `CGO_ENABLED=0`.

Read the files relevant to the task; a full repository scan is not required for every change.

| Location | Purpose |
| --- | --- |
| Root Go files and corresponding `*_test.go` files | Public components, themes, state models, and regression tests |
| [docs/](docs/README.md) | Component APIs, options, internationalization, and compatibility |
| [examples/](examples/) | Runnable minimal app, Gallery, settings app, and data browser |
| [internal/](internal/) | Example implementations, generators, and native test tools; not public APIs |
| [locales/](locales/) | English and Simplified Chinese component messages |
| [icons/lucide/](icons/lucide/README.md), [third_party/](third_party/) | Icon assets, generation inputs, and third-party licenses |
| [Taskfile.yml](Taskfile.yml), [scripts/](scripts/), [CI configuration](.github/workflows/check.yml) | Development commands and automated verification |

See the [README](README.md) for installation and complete examples, and the [compatibility guide](docs/compatibility.md) for supported behavior and limitations.

## Change scope and dependencies

- Inspect the working tree and relevant implementation first. Preserve existing user changes; do not start by resetting the workspace or deleting untracked files.
- Follow existing Go packages, Options types, state types, and naming conventions. Changes to public APIs, defaults, or dependencies must serve the task and include updates to affected examples and documentation.
- Manage dependencies through Go modules. Keep the public installation workflow usable without configuration tied to personal directories.
- To confirm MyGo API availability, locate the pinned dependency with `go list -m -f '{{.Dir}}' github.com/egoist/mygo`, then read its `ui/doc.go` and relevant implementation. Do not assume another version has the same interfaces.
- When upgrading dependencies, review migration guidance, lifecycle assumptions, and platform behavior. Preserve third-party copyright and license notices.
- Before removing files, check references in code, scripts, and documentation, and back up the materials outside the repository.
- Commits, pushes, tags, and releases require explicit user authorization. Permission to edit and verify locally does not authorize publication.

## MyGo lifecycle and interaction

- Persist application data, widget state, and `ui.Handle` values. `ui.Element`, Parts, and custom parts are valid only during the current build; never retain them across builds or capture them in deferred callbacks.
- The parent scope of `*ui.Context` changes during construction. Keep Context out of application models and background tasks. Obtain persistent services through `c.Services()`; do not hide lifecycle errors with validity checks or panic recovery.
- Set `c.Key(stableID)` before constructing a widget. Dynamic lists must use unique business IDs rather than mutable labels, filtered indexes, or display order.
- Use persistent Handles for focus and geometry across builds through APIs such as `Focused`, `Focus`, and `Bounds`. Manage widget state per window and component instance.
- Views and drawing callbacks may run repeatedly. Trigger I/O, collection mutations, and submissions through actions or explicitly controlled jobs, never unconditionally during construction or painting.
- Run slow operations and expensive computation in the background. Publish results on the UI thread through `Window.Update`. Cancel obsolete work or reject stale tickets and results when data changes, views unmount, or windows close.
- Prefer MyGo widgets and Base APIs to preserve keyboard interaction, focus restoration, inherited disabled state, and accessible names. Synchronous `Build` / `Style` callbacks configure only the current build.

## Components, themes, and internationalization

- Use the existing Theme, ComponentTheme, Decoration, and part styling mechanisms. Preserve the precedence of global, local, and instance overrides.
- Distinguish unset values from explicit `false` / `0` with the existing `Optional[T]` and `Set` APIs. Use DIP for dimensions and account for text scaling, light/dark themes, direction, and reduced motion.
- When changing controls, check mouse and keyboard behavior, read-only and disabled states, controlled state, callback counts, and stable identity. Layout changes must not discard focus, drafts, or selections.
- Add built-in messages to both the [English catalog](locales/en.json) and the [Simplified Chinese catalog](locales/zh-Hans.json), preserving ICU parameter names and types. Explicit application text remains application-owned.
- Components use a library catalog. Follow the [internationalization guide](docs/i18n.md) for language setup and message overrides.
- Document differences from Nuxt UI according to the actual native APIs. Visual similarity does not establish equivalent behavior or accessibility. Keep platform limitations accurate when verification is incomplete.

## Formatting and verification

Use the commands in [Taskfile.yml](Taskfile.yml). Formatting requires goimports; linting requires golangci-lint v2. System packages used by Linux checks are listed in the [CI configuration](.github/workflows/check.yml).

| Command | Purpose |
| --- | --- |
| `task` / `task --list` | List available tasks |
| `task fmt` / `task fmt:go` | Format Go code and organize imports with goimports; modifies files |
| `task lint:go` | Validate golangci-lint configuration and check Go code |
| `task lint:go:fix` | Apply supported lint and formatting fixes |
| `task modernize` | Preview Go modernization changes; exits nonzero when changes are suggested |
| `task modernize:fix` | Apply Go modernization changes |
| `task verify` | Check formatting, tests, Go vet, MyGo lifecycle rules, generated source consistency, and cross-platform builds |
| `task test:native` | Run native Gallery interaction tests; requires a desktop session |
| `task test:native:smoke` | Check the first frame of a hidden window; also requires a desktop session |

- For code changes, run relevant tests first, then `task lint:go` and `task verify`. Regression tests should cover observable failures rather than duplicate implementation details.
- Prefer `ui.NewTester` for UI regressions, covering relevant focus, input, shortcuts, and unmount/rebuild behavior. Layout or platform interaction changes also require inspecting rendered output and the corresponding native behavior. A first-frame check alone does not validate interactions.
- Add race checks for shared-state or concurrency changes when the environment supports them. Report checks that could not run and explain why; never report an unexecuted check as passing.
- For documentation-only or comment-only changes, check links, commands, examples, and consistency with configuration. Full test runs and native windows are unnecessary.
- Formatting and automatic fixes may affect multiple files. Review the diff and keep changes relevant to the task; do not rewrite unrelated code.
- The `verify` and native script tasks support Windows / Linux. CI runs on Linux. For script changes, account for case-sensitive paths and Bash / PowerShell differences, keep `.sh` files in LF format, and run `shellcheck scripts/*.sh`.

## Generated files and assets

- Do not edit `palette_gen.go` or icon Go files marked `Code generated` directly. Update the generator or pinned inputs, then run `go generate ./...`.
- Palette and icon generation uses pinned inputs in the repository. Keep generation offline, deterministic, and consistent with formatting. Review generated diffs and run relevant tests after changing assets.
- When upgrading icons, update SVG files, the manifest, and licenses together. Update [third-party notices](THIRD_PARTY_NOTICES.md) when sources or asset versions change.
- Write screenshots, logs, and verification artifacts to the Git-ignored `.work/` directory or outside the repository. Keep backups outside the repository so they do not become public source files.
