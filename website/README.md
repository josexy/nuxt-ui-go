# NuxtUI Go website

The bilingual component website is built with Nuxt 4, Nuxt UI 4, and Nuxt Content. Its live previews run this repository's Go components compiled to WASM. The site has entries for 73 components and layouts, 146 component pages in English and Chinese, editors for every Options field, API tables, and complete Go examples you can copy.

The website tooling does not change how ordinary Go users install the library. The result is a static site that does not need a running Go or Node service, and visitors do not need Go or Node.

## Local development

Install the Go version required by the root `go.mod`, Node.js 24, and pnpm 10.33.3. Then, from the repository root, run:

```sh
pnpm --dir website install --frozen-lockfile
pnpm --dir website dev
```

Open the local address printed in the terminal. Vue and CSS changes reload automatically. After changing the Go library, the example registry, or the API descriptions, run `pnpm --dir website prepare:site` in another terminal and refresh the page.

## Formatting

From the repository root, run `task fmt:website` to format the frontend and `task fmt:go` to format Go code with goimports. You can also format the frontend directly:

```sh
pnpm --dir website format
pnpm --dir website format:check
```

Prettier normalizes indentation and line breaks in Vue, TypeScript, JavaScript, CSS, HTML, and JSON, and puts each attribute of a multi-attribute tag on its own line. Generated pages, indexes, preview assets, and build output are not formatted; edit their source templates instead. `check` includes the format check.

## Checks and static builds

```sh
pnpm --dir website check
pnpm --dir website generate
pnpm --dir website preview
```

- `check`: checks source formatting, generated data, and WASM; type-checks TypeScript; and compiles the 73 default examples and the standalone Go examples generated from every field configuration.
- `generate`: writes `website/.output/public/`, which you can deploy to any static host.
- `preview`: serves the static output at `http://127.0.0.1:4173/`.

To run the Go example regression tests: `CGO_ENABLED=0 go test ./internal/showcase/...` (in PowerShell, set `$env:CGO_ENABLED='0'` first). To view a single example natively: `go run ./internal/cmd/sitepreview button`. Full library checks still use `task lint:go` and `task verify`.

The example regression tests render every enum, toggle, numeric bound, and empty text the editors offer, and check that input is preserved, actions are counted correctly, and dialogs rebuild correctly. Keep option ranges in line with what each component actually supports. For example, input components support `outline / soft / subtle / ghost / none`, but not Button's `solid / link`.

The Go interaction tests also cover every menu action, form validation and submission, selection and expansion state, closing and reopening dialogs, disabled and read-only date ranges, and content bounds in narrow layouts. `test:snippets` runs the Go tests and compiles the standalone examples, while the frontend keeps its format, type, and static build checks.

## Where examples and docs come from

`internal/showcase/catalog.json` is the single source for examples. It holds the bilingual purpose, descriptions, option definitions, persistent state, setup code, and view templates. In a template, `{{parameter}}` is replaced with an expression that reads the configuration in the running version, and with a validated Go literal in the copyable version.

Each component's `description` briefly states its purpose. `guide` is Markdown with basic usage and concrete preview steps. `note` holds necessary limitations or behavior that is easy to misuse; when non-empty, it appears under "Things to know". When you edit this text, update both languages and check button names, options, and preview results. Do not present example settings as component defaults.

`internal/cmd/genshowcase` generates the configuration types and the enum registry from the source code and examples. `fields.go` walks every public Options type and nested field, and `values.go` and `functions.go` generate both the preview values and the public Go code. A field coverage test detects anything missing. Data setup runs only when the data configuration changes, so drawing does not keep resetting it.

`internal/showcase/api-notes.json` holds bilingual descriptions of API fields. `internal/cmd/genwebsite` extracts public functions, Options, and embedded fields from the root package source. The API tables distinguish Go zero values from the component's actual default behavior, and never guess an effective default from the type.

To add an example:

1. Add a registry entry with a stable ID, fill in the English and Chinese text, and list every allowed value for each option.
2. Put app state in `state`, collection setup in `setup`, and the synchronous view in `body`. Do not keep `ui.Element` or `*ui.Context` values.
3. Run `pnpm --dir website prepare:site` to update the compiled views, pages, and API, then run the checks.
4. If the example runs background tasks, cancel them in the generator's `Dispose` and add matching regression tests.

`internal/showcase/preview/generated.go` and `internal/showcase/field_types_gen.go` are generated Go files; commit them together with the example changes. A test checks that they match the templates. The website's Markdown pages, JSON indexes, WASM, Go loader, and build output are all generated by commands; do not edit or commit them.

Fields from shared Options that apply only to other components keep their reference information and show a "does not apply" hint. Go-native extensions are labeled separately. Each field description follows the component that actually uses it, not just the type alias, so aliased types keep their meaning without changing existing configuration IDs.

## GitHub Pages

The repository has a manual **Publish component website** workflow. First, in the repository's Settings → Pages, set the build source to GitHub Actions; a maintainer then runs the workflow manually. The code verification workflow only checks and never publishes.

The project site's base path defaults to `/<repository-name>/`. Set the workflow's `base_path` to `/` to deploy at the domain root. All pages, WASM, language files, and scripts follow this path.

To verify the same path locally (PowerShell):

```powershell
$env:NUXT_APP_BASE_URL='/nuxt-ui-go/'
pnpm --dir website generate
pnpm --dir website preview
```

Then open `http://127.0.0.1:4173/nuxt-ui-go/`. The environment variable must be set for both the build and the preview. WASM file names include a content hash, and the host must serve them as `application/wasm`. The first download is larger than an ordinary docs page, and later pages share the cache.

## Known limits

- Every public Options field is configurable, including nested objects, data lists, UI/Style, icons/SVGs, images, persistent pointers, and callbacks. Callbacks offer presets for event logging, synchronous content, styles, formatting, predicates, return values, search, and lazy loading. Compiling arbitrary Go programs online is not supported.
- When a field is not overridden, the example's own behavior is kept, and the reset button removes the override. An explicit `null` clears the value or restores an Optional to unset, while `false` and `0` stay explicit values.
- Each list holds at most 100 items, and images are limited to 2 MB and 4096 pixels. Invalid options show a specific error and keep the last valid preview. While the preview is out of sync, outdated code cannot be copied.
- Changing the color, size, or language updates the same preview instance and keeps input and selections. Switching to another component or resetting creates a new example. The option editors keep each component's configuration for the current browsing session.
- Previews use MyGo v0.7.4. File dialogs, assistive technology semantics, fonts, and system input are limited by the browser. FileUpload uses sample files with native selection turned off; your app implements network uploads.
- Animations respect the system's reduced motion preference. The page and option tables are ordinary HTML and keyboard accessible, but the Canvas preview is no substitute for accessibility testing on the desktop.
- The frontend loads only one WASM instance, for the current example. If a preview fails to load, the description, API, and default complete code remain readable, and a retry option is offered.
