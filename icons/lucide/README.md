# Lucide SVG icons

This package includes all **2,015 SVGs** from the [icons directory at a pinned Lucide commit](https://github.com/lucide-icons/lucide/tree/a630cec77eac6568b989e87e4947a8bfe45caacd/icons). The original files are in `svg/`, the commit and per-file SHA-256 hashes are recorded in [manifest.json](manifest.json), and the license is in [LICENSE](LICENSE). The package does not track upstream changes automatically.

## Usage

```go
import (
    "github.com/egoist/mygo/ui"
    nui "github.com/josexy/nuxt-ui-go/ui"
    "github.com/josexy/nuxt-ui-go/icons/lucide"
)

// Inside a view, using the current Context and Theme.
ui.Icon(c, lucide.House()).Size(20, 20).TextColor(theme.Primary)
nui.Button(c, theme, "Repository", nui.ButtonOptions{Leading: lucide.GitBranch()})
```

Each getter returns a shared, read-only `*ui.SVG` that is parsed once, on the first call. Set the color and size on the UI element that displays it; the shared SVG itself is never modified. Commonly used icons include `Check()`, `Search()`, and `LoaderCircle()`.

| SVG name | Go function |
| --- | --- |
| `house` | `lucide.House()` |
| `git-branch` | `lucide.GitBranch()` |
| `file-json-2` | `lucide.FileJson2()` |
| `circle-arrow-up-right` | `lucide.CircleArrowUpRight()` |
| `robot-arm` | `lucide.RobotArm()` |

To get the function name, split the SVG name on `-` and capitalize the first letter of each part, keeping the other characters and digits as they are. Use your IDE's completion on `lucide.`, or run `go doc github.com/josexy/nuxt-ui-go/icons/lucide`, to see every function.

Each icon is embedded and cached separately. Initializing the package does not parse every icon, and production code has no registry of all icons. The Go linker can remove getters and SVGs you never call, so adding new icons does not grow executables that use only a few of them.

## Generation and verification

```powershell
$env:CGO_ENABLED = '0'
go generate ./icons/lucide
go test ./icons/lucide ./internal/cmd/genicons
```

Run these commands from the repository root. Generation is fully offline: it verifies every SVG against the manifest, then generates one getter per icon and a full list used only by tests. It rejects unregistered SVGs, hash mismatches, duplicate or conflicting names, overwriting handwritten code, and leftover generated files. File names that could trigger Go's platform build constraints get an `_icon.go` suffix, for example `robot_arm_icon.go`.

To upgrade the icons, fetch the complete `icons/*.svg` set and LICENSE from a specific new commit, and update the local collection and the manifest. Review icons removed upstream and delete their old assets and generated files, then regenerate and run the tests. Consumers do not need to change their import paths.
