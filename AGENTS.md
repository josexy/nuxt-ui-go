# AGENTS.md

Guidance for coding agents working anywhere in this repository. Instructions in the current task override this file, and a more specific instruction file in a subdirectory overrides it within that subdirectory.

## Project overview

NuxtUI Go is a native Go component library built on MyGo, styled after Nuxt UI v4. The module path is `github.com/josexy/nuxt-ui-go` and the root package is `nui`. [go.mod](go.mod) is the source of truth for the required Go toolchain, the MyGo version, and other dependencies.

Build and test with `CGO_ENABLED=0`. Taskfile commands already set it; set it yourself when you run `go` directly.

Read only the files relevant to the task; you do not need to scan the whole repository for every change.

| Location | Contents |
| --- | --- |
| Root Go files and their `*_test.go` files | Public components, themes, state models, and regression tests |
| [docs/](docs/README.md) | Component APIs, options, internationalization, and compatibility |
| [website/](website/README.md) | Bilingual component website, option editors, and live WASM previews |
| [internal/showcase/](internal/showcase/) | Website examples, API descriptions, configuration, and Go preview tests |
| [examples/](examples/) | Runnable minimal app, Gallery, settings app, and data browser |
| [internal/](internal/) | Example implementations, generators, and native test tools; not public API |
| [locales/](locales/) | English and Simplified Chinese component messages |
| [icons/lucide/](icons/lucide/README.md), [third_party/](third_party/) | Icon assets, generator inputs, and third-party licenses |
| [Taskfile.yml](Taskfile.yml), [scripts/](scripts/), [.github/workflows/](.github/workflows/) | Development commands and automated checks |
| [.agents/skills/](.agents/skills/) | Agent skills, including [nuxt-ui-go](.agents/skills/nuxt-ui-go/SKILL.md) (per-component references for using the components) |

See the [README](README.md) for installation and complete examples, and [docs/compatibility.md](docs/compatibility.md) for supported behavior and known limitations.

## Scope and dependencies

- Inspect the working tree and the relevant code before changing anything. Preserve existing user changes; never reset the workspace or delete untracked files to get a clean start.
- Follow the existing package layout, Options types, state types, and naming conventions.
- Change public APIs, defaults, or dependencies only when the task requires it, and update the affected examples and documentation in the same change.
- Manage dependencies with Go modules. Never commit configuration that points to a personal directory, such as a local `replace` directive or a `go.work` file; users must be able to install with plain `go get`.
- To check whether a MyGo API exists, find the pinned module's directory in the `Dir` field of `go mod download -json github.com/egoist/mygo` (unlike `go list -m`, it downloads the module if it is not cached yet), then read its `ui/doc.go` and the relevant source. Do not assume other MyGo versions have the same interfaces.
- When upgrading a dependency, review its migration notes, lifecycle assumptions, and platform behavior. Keep third-party copyright and license notices intact.
- Before deleting a file, check for references in code, scripts, and documentation, and back it up outside the repository.
- Commits, pushes, tags, and releases require the user's explicit permission. Permission to edit and verify locally does not include publishing.

## MyGo lifecycle and interaction

- Keep app data, widget state, and `ui.Handle` values in persistent storage. `ui.Element` values, Parts, and custom parts are valid only during the current build; never keep them across builds or capture them in deferred callbacks.
- The parent scope of `*ui.Context` changes during a build, so never store a Context in an app model or a background task. Get long-lived services from `c.Services()`.
- Do not hide lifecycle errors behind validity checks or panic recovery; fix the lifecycle problem instead.
- Call `c.Key(stableID)` before creating each widget. In dynamic lists, use unique business IDs, never editable labels, filtered indexes, or display order.
- Track focus and geometry across builds with persistent Handles and APIs such as `Focused`, `Focus`, and `Bounds`. Keep widget state separate for each window and each component instance.
- Views, `Build` and `Style` callbacks, and draw callbacks can run many times, and they configure only the current build. Trigger I/O, collection changes, and submissions from actions or explicitly controlled jobs, never unconditionally while building or painting.
- Run slow operations and expensive computation in the background, and publish results on the UI thread with `Window.Update`. When data changes, a view unmounts, or a window closes, cancel work that is no longer needed or reject its stale tickets and results.
- Prefer MyGo widgets and Base APIs so that keyboard interaction, focus restoration, inherited disabled state, and accessible names keep working.

## Components, themes, and internationalization

- Style components with the existing Theme, ComponentTheme, Decoration, and part mechanisms, and keep the precedence of global, local, and instance overrides.
- Use the existing `Optional[T]` and `Set` APIs to distinguish an unset value from an explicit `false` or `0`.
- Express sizes in DIP, and account for text scaling, light and dark themes, writing direction, and reduced motion.
- When changing a control, check mouse and keyboard behavior, read-only and disabled states, controlled state, callback counts, and stable identity. Layout changes must not lose focus, drafts, or selections.
- Add every built-in message to both [locales/en.json](locales/en.json) and [locales/zh-Hans.json](locales/zh-Hans.json), keeping ICU parameter names and types identical. Text that an app passes in explicitly belongs to the app; the library does not translate it. See [docs/i18n.md](docs/i18n.md) for how the library catalog and app overrides work.
- Document differences from Nuxt UI based on the actual native API. Looking the same does not mean behaving the same or being equally accessible. When verification is incomplete, describe platform limitations accurately rather than claiming support.

## Website synchronization

- When a component's API, defaults, behavior, or appearance changes, update the affected website demos, option editors, bilingual documentation, and copyable Go examples. If no website change is needed, say why in your summary.
- Edit the sources in `internal/showcase/` and `website/`, and follow [website/README.md](website/README.md) to regenerate output and to decide which generated files to commit. Never edit generated files by hand; see [Generated files and assets](#generated-files-and-assets).
- For component or website code changes, run `task website:check` and `task website:build` in addition to the Go checks, then verify the affected interactions in a browser.

## Formatting and verification

Run commands through [Taskfile.yml](Taskfile.yml). The scripts behind `task verify` and the native test tasks need Python 3.9 or later, formatting needs goimports, and linting needs golangci-lint v2. The system packages that Linux checks need are listed in the [CI configuration](.github/workflows/check.yml).

| Command | Purpose |
| --- | --- |
| `task` / `task --list` | List available tasks |
| `task fmt` (alias of `task fmt:go`) | Format Go code and organize imports with goimports; edits files |
| `task fmt:website` | Format website sources with Prettier; edits files |
| `task lint:go` | Validate the golangci-lint configuration and lint all Go code |
| `task lint:go:fix` | Apply supported lint and formatting fixes; edits files |
| `task modernize` | Preview Go modernization changes; exits nonzero if any are suggested |
| `task modernize:fix` | Apply Go modernization changes; edits files |
| `task verify` | Check formatting, tests, `go vet`, MyGo lifecycle rules, generated source consistency, and cross-platform builds |
| `task test:native` | Run native Gallery interaction tests; needs a desktop session |
| `task test:native:smoke` | Check the first frame of a hidden native window; needs a desktop session |
| `task website:dev` | Run the website locally with live Go previews |
| `task website:check` | Check website formatting, types, Go previews, and standalone examples |
| `task website:build` | Generate the static website, API pages, and WASM previews |

Which checks to run depends on the change:

- **Code changes:** Run the relevant tests first, then `task lint:go` and `task verify`. Regression tests should catch observable failures, not restate implementation details.
- **UI behavior:** Prefer `ui.NewTester` for UI regression tests, and cover the relevant focus, input, shortcut, and unmount/rebuild behavior.
- **Layout or platform interaction:** Also inspect the rendered output and the native behavior on the platform. A first-frame check alone does not validate interactions.
- **Shared state or concurrency:** Add race checks when the environment supports them.
- **Scripts:** The scripts in `scripts/` are written in Python so one copy runs on Windows, Linux, and macOS; CI runs on Linux. Use only the standard library and keep them compatible with Python 3.9. Account for case-sensitive paths, `.exe` suffixes, Unix socket path limits, and signal differences; pass commands as argument lists instead of shell strings; and keep `.py` files in LF line endings. `task verify` parses every script with the Python 3.9 grammar; also run `python3 scripts/test-native.py --help`.
- **Documentation or comments only:** Check links, commands, examples, and consistency with the configuration. Full test runs and native windows are not needed.

Formatters and automatic fixes can touch many files. Review the diff and keep only changes relevant to the task; do not rewrite unrelated code.

Report every check you could not run and why. Never report an unexecuted check as passing.

## Generated files and assets

Never edit generated files by hand. Change the generator or its pinned inputs, then run `go generate ./...`. `task verify` fails if committed generated files are out of date.

| Generated files | Generator |
| --- | --- |
| `palette_gen.go` | `internal/cmd/genpalettes` |
| Go files in `icons/lucide/` marked `Code generated` | `internal/cmd/genicons` |
| `internal/showcase/field_types_gen.go`, `internal/showcase/preview/generated.go` | `internal/cmd/genshowcase`, also run by `pnpm --dir website prepare:site` |
| `.agents/skills/nuxt-ui-go/references/components/*.md` | `internal/cmd/genskill` (`go run ./internal/cmd/genskill`); not part of `go generate` or `task verify`, so run it with `-check` after changing a component's API, docs or catalog entry |

- The nuxt-ui-go skill is given to library users as a standalone folder. Keep it self-contained: no repository paths, and links only to files inside the skill.
- Palette and icon generation reads pinned inputs stored in the repository. Keep generation offline, deterministic, and gofmt-clean. After changing assets, review the generated diff and run the relevant tests.
- When upgrading icons, update the SVG files, the manifest, and the license together. Update [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) whenever a source or asset version changes.
- Write screenshots, logs, and other verification artifacts to the Git-ignored `.work/` directory or outside the repository. Keep backups outside the repository so they never become public source files.
