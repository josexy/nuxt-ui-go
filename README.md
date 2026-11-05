# NuxtUI Go

用 Go 编写桌面界面的组件库。NuxtUI Go 基于 [MyGo](https://github.com/egoist/mygo)，提供按钮、输入框、表单、表格、树、图表和应用布局，外观采用 Nuxt UI v4 的设计风格。

原生应用不依赖 Vue、Node.js 或 WebView。组件支持浅色、深色和系统外观，内置英文、简体中文、26 套色板，以及 2,015 个按需使用的 Lucide 图标。

**第一次使用，从下面的最小应用开始。** 想先查看组件效果，可以[运行 Gallery](#运行示例)；已有应用需要接入组件，直接查看[组件指南](docs/README.md)。

## 安装与最小应用

需要 **Go 1.27.3 或更高版本**，以及桌面图形会话。构建时设置 `CGO_ENABLED=0`。Windows、Linux、macOS 的验证程度不同，具体见[兼容性说明](docs/compatibility.md#环境与平台)。

### 1. 创建项目

```sh
mkdir nuxt-ui-demo
cd nuxt-ui-demo
go mod init example.com/nuxt-ui-demo
go get github.com/josexy/nuxt-ui-go@latest
```

模块路径是 `github.com/josexy/nuxt-ui-go`，代码中使用包名 `nui`。本仓库使用的 Go 和 MyGo 版本见 [go.mod](go.mod)。

### 2. 创建 main.go

这个程序会打开一个窗口，点击 Save 后更新计数。代码与仓库中的[最小示例](examples/minimal/main.go)一致。

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

这里有三处需要了解：

- `count` 和 `config` 放在视图函数外，界面重新构建时仍保留原值。
- `nui.App(c, config)` 为当前窗口准备主题，后面的组件使用返回的 `theme`。
- `OnClick` 在点击后修改数据；MyGo 根据新数据重新构建界面。

### 3. 运行应用

Windows PowerShell：

```powershell
go mod tidy
$env:CGO_ENABLED = '0'
go run .
```

Linux / macOS shell：

```sh
go mod tidy
CGO_ENABLED=0 go run .
```

应用成功打开后，可以继续阅读[组件指南中的基本写法](docs/README.md#理解组件的基本写法)，为窗口添加输入框和其他控件。

## 运行示例

以下命令在**本仓库根目录**执行。先将 `CGO_ENABLED` 设为 `0`：PowerShell 使用 `$env:CGO_ENABLED = '0'`，Linux / macOS shell 使用 `export CGO_ENABLED=0`。

| 想查看什么 | 命令 | 示例源码 |
| --- | --- | --- |
| 全部组件、主题和交互效果 | `go run ./examples/gallery` | [Gallery](examples/gallery/) |
| 窗口、主题和按钮的最小用法 | `go run ./examples/minimal` | [最小应用](examples/minimal/main.go) |
| 表单校验、编辑对话框、菜单和通知 | `go run ./examples/settings` | [设置页](examples/settings/) |
| 搜索、筛选、分页、表格和树 | `go run ./examples/data-browser` | [数据浏览器](examples/data-browser/) |

## 按任务查文档

| 任务 | 文档 |
| --- | --- |
| 查找组件，了解状态和回调的基本写法 | [组件指南](docs/README.md) |
| 修改颜色、字号、间距或某个组件的样式 | [主题与基础组件](docs/foundations.md) |
| 查询默认值，配置多选、方向、加载状态或弹层 | [参数与默认值](docs/options.md) |
| 切换语言、修改内置文案或使用图标 | [国际化](docs/i18n.md)、[Lucide 图标](icons/lucide/README.md) |
| 确认平台支持、已知问题和版本变化 | [兼容性与已知限制](docs/compatibility.md)、[更新日志](CHANGELOG.md) |

组件网站提供 73 项组件与布局的中英文说明、Go/WASM 实时预览、参数编辑和可复制的完整 Go 示例。按[网站开发说明](website/README.md#本地开发)安装依赖后，运行 `task website:dev` 即可本地查看。静态构建和 GitHub Pages 发布方式也在该文档中。

## 参与开发

安装 [Task](https://taskfile.dev/docs/installation) 后，在仓库根目录运行 `task` 查看可用命令。`task verify` 会选择 Windows 或 Linux 验证脚本，检查格式、测试、生命周期规则和跨平台构建。

完整命令见 [Taskfile.yml](Taskfile.yml)，持续集成配置见 [check.yml](.github/workflows/check.yml)。开发网站还需要 Node.js 和 pnpm，普通 Go 应用不需要安装它们。

## 许可证

本项目采用 [MIT License](LICENSE)。Nuxt UI、Lucide/Feather、Tailwind CSS 和 MyGo 的来源及许可见[第三方声明](THIRD_PARTY_NOTICES.md)。
