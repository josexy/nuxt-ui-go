# Third-party notices

## Nuxt UI

The component styling recipes follow [nuxt/ui](https://github.com/nuxt/ui/tree/200b0bfb7a5666c870742f3f43c83cf09234aed9) version 4.12.0, commit `200b0bfb7a5666c870742f3f43c83cf09234aed9`.

Copyright (c) 2023 Nuxt, MIT License. The full license is in [nuxt-ui.txt](third_party/licenses/nuxt-ui.txt). Vue and Tailwind are not included at runtime.

## Lucide

All **2,015 SVGs** from the `icons/` directory of [lucide-icons/lucide](https://github.com/lucide-icons/lucide/tree/a630cec77eac6568b989e87e4947a8bfe45caacd/icons) at pinned commit `a630cec77eac6568b989e87e4947a8bfe45caacd` are included. File names and per-file SHA-256 hashes are listed in [manifest.json](icons/lucide/manifest.json).

The full ISC license and the MIT notice for the Feather icons it derives from are in [icons/lucide/LICENSE](icons/lucide/LICENSE). The SVGs are kept unmodified; the Go package adds a lazily parsed cache for each icon.

## Tailwind CSS palettes

The color scales come from [Tailwind CSS v4.3.3 theme.css](https://github.com/tailwindlabs/tailwindcss/blob/v4.3.3/packages/tailwindcss/theme.css). The source data used by the generator is stored in [theme.css](third_party/tailwind/theme.css) with SHA-256 `443d7af364200f8ec8352dc78e39485a118a1840d498c97967bf0ae402b39167`. A Go generator extracts 17 color palettes and 9 neutral palettes from it, for 26 OKLCH color scales in total. Black is this library's own override that switches with light and dark mode; it is not an additional extracted palette.

MIT License, Copyright (c) Tailwind Labs, Inc. The full license is in [tailwindcss.txt](third_party/licenses/tailwindcss.txt). Only the generated Go color values are included at runtime; there is no CSS engine.

## MyGo

Native layout, input, text, drawing, focus, selection, and overlay capabilities come from [egoist/mygo v0.9.0](https://github.com/egoist/mygo/tree/v0.9.0) as a Go module dependency. See [go.mod](go.mod) and [go.sum](go.sum) for the exact version and transitive dependencies, each of which keeps its own license.

This project's own code is released under the [MIT License](LICENSE), copyright NuxtUI Go contributors. The third-party materials above keep their own copyright and license notices, and this project's license does not replace them.
