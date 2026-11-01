# Lucide SVG 图标

本包完整收录 [Lucide 固定提交的 icons 目录](https://github.com/lucide-icons/lucide/tree/a630cec77eac6568b989e87e4947a8bfe45caacd/icons) 中的 **2,015 个 SVG**。原始文件保存在 `svg/`，提交和逐文件 SHA-256 保存在 [manifest.json](manifest.json)，许可见 [LICENSE](LICENSE)。不自动追踪上游最新提交。

## 使用

```go
import (
    "github.com/egoist/mygo/ui"
    nui "github.com/josexy/nuxt-ui-go"
    "github.com/josexy/nuxt-ui-go/icons/lucide"
)

// 在 view 中使用当前 Context 和 Theme。
ui.Icon(c, lucide.House()).Size(20, 20).TextColor(theme.Primary)
nui.Button(c, theme, "Repository", nui.ButtonOptions{Leading: lucide.GitBranch()})
```

每个 getter 返回共享、只读的 `*ui.SVG`，首次调用时解析一次。颜色与尺寸通过当前 UI 元素配置，不修改共享 SVG。常用入口包括 `Check()`、`Search()`、`LoaderCircle()`。

| SVG 名称 | Go 入口 |
| --- | --- |
| `house` | `lucide.House()` |
| `git-branch` | `lucide.GitBranch()` |
| `file-json-2` | `lucide.FileJson2()` |
| `circle-arrow-up-right` | `lucide.CircleArrowUpRight()` |
| `robot-arm` | `lucide.RobotArm()` |

命名规则为按 `-` 分段，各段首字母大写，其余字符和数字保留。可以在 IDE 中补全 `lucide.`，或用 `go doc github.com/josexy/nuxt-ui-go/icons/lucide` 查看全部函数。

每个图标独立 embed 和缓存，包初始化不会解析全量图标，也没有生产代码中的全量注册表。未调用的 getter 和 SVG 可由 Go 链接器移除；新增图标不会强制扩大仅使用少量图标的可执行文件。

## 生成与验证

```powershell
$env:CGO_ENABLED = '0'
go generate ./icons/lucide
go test ./icons/lucide ./internal/cmd/genicons
```

从项目根目录调用上述命令。生成过程完全离线，根据 manifest 验证所有 SVG，再生成独立 getter 和仅供测试使用的全量清单。它会拒绝未登记 SVG、哈希不匹配、重复或冲突名称、覆盖手写代码以及残留的旧生成文件。可能触发 Go 平台约束的文件名使用 `_icon.go` 后缀，例如 `robot_arm_icon.go`。

升级资源时，从明确的新提交获取完整 `icons/*.svg` 和 LICENSE，更新本地集合与 manifest；审核上游删除的图标并移除对应旧资源和生成文件，然后重新生成并测试。无需修改消费者的导入路径。
