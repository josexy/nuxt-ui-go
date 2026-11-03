# NuxtUI Go website / 双语组件站

Nuxt 4、Nuxt UI 4 和 Nuxt Content 构建文档界面；预览运行本仓库 Go 组件编译出的 WASM。站点包含 73 项组件与布局入口、146 个中英文组件页面、完整 Options 字段面板、API 表和可复制的完整 Go 示例。

## 本地开发

安装根目录 `go.mod` 要求的 Go、Node.js 24 和 pnpm 10.33.3。在仓库根目录执行：

```sh
pnpm --dir website install --frozen-lockfile
pnpm --dir website dev
```

打开终端输出的本地地址。Vue/CSS 修改会热更新。修改 Go 库、示例注册表或 API 说明后，在另一个终端运行 `pnpm --dir website prepare:site`，然后刷新页面。

网站工具不会改变普通 Go 用户的安装方式，也不需要常驻 Go 或 Node 服务。

## 代码格式化

在仓库根目录执行 `task fmt:website` 格式化前端，执行 `task fmt:go` 用 goimports 格式化 Go 代码。前端也可以直接使用：

```sh
pnpm --dir website format
pnpm --dir website format:check
```

Prettier 统一 Vue、TypeScript、JavaScript、CSS、HTML 和 JSON 的缩进与换行，多属性标签按属性分行。生成的页面、索引、预览资源和构建产物不参与格式化；应修改其源模板。`check` 包含格式检查。

## 检查与静态构建

```sh
pnpm --dir website check
pnpm --dir website generate
pnpm --dir website preview
```

- `check`：检查源码格式、生成资料和 WASM、检查 TypeScript，并编译 73 个默认示例和全部字段配置生成的独立 Go 示例。
- `generate`：输出 `website/.output/public/`，可以交给静态托管服务。
- `preview`：在 `http://127.0.0.1:4173/` 预览静态产物。

Go 示例回归测试：`CGO_ENABLED=0 go test ./internal/showcase/...`。PowerShell 使用 `$env:CGO_ENABLED='0'`。原生查看单个示例：`go run ./internal/cmd/sitepreview button`。库的完整检查仍使用 `task lint:go` 和 `task verify`。

示例回归会逐项渲染面板提供的枚举、开关、数值边界与空文字，并检查输入保留、操作次数和弹窗重建。参数范围需要依据实际组件能力维护，例如输入类组件支持 `outline / soft / subtle / ghost / none`，不能使用 Button 的 `solid / link`。

Go 交互回归还覆盖菜单的每项操作、表单校验与提交、选择与展开状态、弹窗关闭与重开、日期范围的禁用/只读，以及窄屏内容边界。`test:snippets` 调用 Go 测试并编译独立示例；前端保留格式、类型和静态构建检查。

## 示例与文档的来源

`internal/showcase/catalog.json` 是示例的唯一来源，维护双语用途、说明、参数定义、持久状态、初始化代码和视图模板。模板中的 `{{parameter}}` 在运行版本中替换为配置读取表达式，在复制版本中替换为经过验证的 Go 字面量。

`internal/cmd/genshowcase` 从源码和示例生成配置类型、枚举注册表；`fields.go` 自动遍历所有公开 Options 和嵌套字段，`values.go` / `functions.go` 同时生成预览值与公开 Go 代码。字段覆盖测试会检测遗漏。数据初始化只在数据配置变化时执行，不随绘制反复重置。

`internal/showcase/api-notes.json` 维护 API 字段的双语说明。`internal/cmd/genwebsite` 从根包源码提取公开函数、Options 和嵌入字段；API 表区分 Go 零值与实际组件默认行为，不从类型猜测有效默认值。

增加示例时：

1. 添加具有稳定 ID 的注册项，填写中英文文案及全部参数允许值。
2. 将应用状态放在 `state`，集合初始化放在 `setup`，同步视图放在 `body`；不保留 `ui.Element` 或 `*ui.Context`。
3. 运行 `pnpm --dir website prepare:site` 更新编译视图、页面和 API，再执行检查。
4. 有后台任务的示例需在生成器的 `Dispose` 中取消任务，补充对应回归测试。

`internal/showcase/preview/generated.go` 和 `internal/showcase/field_types_gen.go` 是生成的 Go 文件，需要随示例修改一起提交；测试会检查它是否与模板一致。网站的 Markdown 页面、JSON 索引、WASM、Go 加载器和构建产物均由命令生成，不手工编辑、不提交。

## GitHub Pages

仓库提供 **Publish component website** 手动工作流。先在仓库 Settings → Pages 中将构建来源设为 GitHub Actions，再由维护者手动运行工作流。代码验证工作流只检查，不发布。

项目站点默认基路径为 `/<repository-name>/`。工作流中的 `base_path` 可设为 `/` 以部署到根域名；所有页面、WASM、语言文件和脚本都跟随此路径。

本地验证同一路径（PowerShell）：

```powershell
$env:NUXT_APP_BASE_URL='/nuxt-ui-go/'
pnpm --dir website generate
pnpm --dir website preview
```

打开 `http://127.0.0.1:4173/nuxt-ui-go/`。环境变量需要同时用于构建和预览。WASM 文件使用内容哈希，宿主需要以 `application/wasm` 提供该文件；首次下载比普通文档页大，后续页面共享缓存。

## 已知边界

- 全部公开 Options 字段可配置，包含嵌套对象、数据列表、UI/Style、图标/SVG、图片、持久指针与回调。回调提供事件记录、同步内容、样式、格式化、谓词、返回值、搜索和懒加载等预设；不提供任意 Go 程序在线编译。
- 字段未覆盖时保留示例行为，重置按钮移除覆盖；显式 `null` 清空或恢复 Optional 的未设置状态，`false` 与 `0` 保持显式设置。
- 单列表最多 100 项，图片限制 2 MB / 4096 像素。无效参数会显示具体错误并保留上一次有效预览；未同步时不允许复制过期代码。
- 改颜色、尺寸或语言会更新同一预览实例，保留输入与选择；切换到另一个组件或重置会创建新示例。参数面板在本次浏览会话内保留各组件的配置。
- 预览使用 MyGo v0.7.4。文件选择框、辅助技术语义、字体与系统输入等存在浏览器限制；FileUpload 使用样例文件并关闭原生选择，网络上传由应用实现。
- 动画遵守系统减少动态效果偏好。界面和参数表使用普通 HTML，可通过键盘访问；Canvas 预览不等同于桌面的可访问性验证。
- 前端只加载一份当前示例的 WASM 实例。预览加载失败时，说明、API 和默认完整代码仍可阅读，并提供重试。

## English

The site uses Nuxt 4, Nuxt UI 4 and Nuxt Content. Live previews run the actual Go library through MyGo WASM. It is a static site with 73 entries and 146 localized component pages; visitors do not need Go or Node.

Use Go from the repository's `go.mod`, Node 24 and pnpm 10.33.3. Run `pnpm --dir website install --frozen-lockfile`, then `pnpm --dir website dev`. Run `check` and `generate` before publishing. Frontend checks cover formatting, types and the static build; Go regression tests and standalone snippet compilation are retained. Generated output is `website/.output/public/`.

Every public Options field has an editor, including nested styles, data and callback presets. The preview and standalone Go code share typed conversion and validation. Unset, explicit false/zero and cleared values remain distinct. The shared example catalog owns controls, templates and code generation. API signatures and embedded option fields come from the current Go source. Run `prepare:site` after changing Go sources and refresh the browser. Generated preview Go code is committed and checked for drift; generated website pages and binaries are ignored.

Publishing is manual through **Publish component website**. Set `NUXT_APP_BASE_URL` to the project path for local production verification. Native dialogs and full assistive-technology semantics are not supported by the browser backend; previews complement desktop testing.
