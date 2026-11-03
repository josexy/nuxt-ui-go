# NuxtUI Go

基于 [MyGo](https://github.com/egoist/mygo) 的原生 Go 组件库，采用 Nuxt UI v4 的视觉风格，适合构建桌面表单、数据工具与应用界面。

**[MIT License](LICENSE)**

- 73 项组件与布局入口，覆盖主题、输入、表单、弹层、导航、日期、表格、树和图表。
- 浅色、深色与系统模式，26 套色板，部件样式定制及 2,015 个按需使用的 Lucide 图标。
- 英文、简体中文与全局语言切换；原生布局和绘制，不依赖 Vue、Node 或 WebView。

平台支持、运行要求和已知输入法限制见 [兼容性与已知限制](docs/compatibility.md)。

## 安装与最小应用

需要 Go **1.27.3 或更高版本**，原生应用使用 `CGO_ENABLED=0` 构建。模块路径为 `github.com/josexy/nuxt-ui-go`，根包名为 `nui`，依赖版本见 [go.mod](go.mod)。

创建应用目录，初始化模块并添加依赖：

```sh
mkdir nuxt-ui-demo
cd nuxt-ui-demo
go mod init example.com/nuxt-ui-demo
go get github.com/josexy/nuxt-ui-go@latest
```

在应用目录创建 `main.go`，内容与仓库的 [最小示例](examples/minimal/main.go) 一致：

```go
package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func main() {
	config := nui.DefaultApp()
	count := 0
	mygo.App.SetLanguages(nui.SupportedLocales()...)
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title: "NuxtUI Go minimal", Width: 480, Height: 240,
			Content: ui.View(func(c *ui.Context) {
				theme := nui.App(c, config)
				nui.Container(c, theme, nui.ContainerOptions{}, func() {
					ui.Textf(c, "Saved %d times", count)
					nui.Button(c.Key("save"), theme, "Save", nui.ButtonOptions{}).
						OnClick(func() { count++ })
				}).PaddingY(24)
			}),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
```

Windows PowerShell：

```powershell
go mod tidy
$env:CGO_ENABLED = '0'
go run .
```

macOS/Linux shell：`go mod tidy` 后执行 `CGO_ENABLED=0 go run .`。应用需要桌面图形会话；这两个平台的完整桌面交互尚未验证，具体范围见 [兼容性说明](docs/compatibility.md)。

## 运行示例

在本仓库根目录执行（PowerShell）：

```powershell
$env:CGO_ENABLED = '0'
go run ./examples/gallery
```

Linux/macOS shell：

```sh
CGO_ENABLED=0 go run ./examples/gallery
```

安装 [Task](https://taskfile.dev/docs/installation) 后可用 `task` 查看命令，`task verify` 会自动选择 Windows/Linux 验证脚本。完整任务见 [Taskfile.yml](Taskfile.yml)，CI 配置见 [check.yml](.github/workflows/check.yml)。

| 示例 | 命令 | 内容 |
| --- | --- | --- |
| Gallery | `go run ./examples/gallery` | 组件、主题、状态、参数与方向切换 |
| 最小应用 | `go run ./examples/minimal` | 创建窗口、应用主题和处理按钮动作 |
| 设置页 | `go run ./examples/settings` | 表单校验、编辑对话框、菜单、通知与侧边筛选 |
| 数据浏览器 | `go run ./examples/data-browser` | 搜索、筛选、分页、Table/Tree 与按 ID 编辑 |

Gallery 的 Theme 菜单可以切换 Primary、Neutral 及 Light/Dark/System。English / 简体中文 切换组件内置语言；示例标题与业务数据由应用提供，不随组件文案自动翻译。图表页面展示图型、坐标轴、曲线、堆叠与图例交互；需要静态展示时运行 `go run ./examples/gallery -static`。

## 文档

组件网站源码位于 [website/](website/README.md)，提供 73 项组件的中英文说明、实时 Go/WASM 预览、常用参数配置和可复制的 Go 示例。安装网站依赖后可用 `task website:dev` 本地查看；构建和手动 GitHub Pages 发布见网站说明。

- [组件指南](docs/README.md)：按用途查找组件和 Go API。
- [主题与样式](docs/foundations.md)：色板、系统模式、局部主题和部件覆盖。
- [参数与默认值](docs/options.md)：方向、选择策略、加载态和弹层配置。
- [国际化](docs/i18n.md)：语言切换、应用目录与组件文案覆盖。
- [Lucide 图标](icons/lucide/README.md)：类型化 getter、按需嵌入及使用方式。
- [兼容性与已知限制](docs/compatibility.md)：平台状态、输入法问题与功能边界。
- [更新日志](CHANGELOG.md)。

## 许可证

本项目采用 [MIT License](LICENSE)。Nuxt UI、Lucide/Feather、Tailwind CSS 与 MyGo 的来源及各自许可见 [第三方声明](THIRD_PARTY_NOTICES.md)。
