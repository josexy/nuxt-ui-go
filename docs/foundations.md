# 主题与基础组件

使用 `App` 建立窗口主题，再将 Theme 传给组件。最小应用见 [minimal](../examples/minimal/main.go)，平台与功能范围见 [兼容性说明](compatibility.md)。

## 应用与主题

持久状态保存 `AppConfig`、业务值和 `ui.Handle`。每次构建先调用 `App`，随后把返回的 Theme 和当前 Context 传给组件：

```go
// 窗口模型持有 config；不要在每次构建时覆盖用户修改。
config := nui.DefaultApp()
config.Mode = nui.SystemMode // 也支持 LightMode、DarkMode；零 AppConfig 跟随系统
p := nui.DefaultPalettes()
p.Primary = nui.BluePalette()
p.Neutral = nui.ZincPalette()
config.Light = nui.NewTheme(false, p)
config.Dark = nui.NewTheme(true, p)
config.Light.Radius, config.Dark.Radius = 8, 8
config.Light.Density, config.Dark.Density = .9, .9

// 在 func(c *ui.Context) 内：
theme := nui.App(c, config)
nui.Button(c.Key("save"), theme, "保存", nui.ButtonOptions{
    Color: nui.Primary, Variant: nui.Solid, Size: nui.MD,
    Leading: lucide.Check(),
}).OnClick(save)
```

`App` 必须在每次构建中的其他 `Window` / `Themed` 调用之前执行，才能读取 MyGo 的系统外观。它设置当前窗口的原生主题，不创建 MyGo 窗口，也不提供 Toast 队列。`Window(c, theme)` 是底层显式快照入口，不自动叠加系统偏好。已有原生服务通过 `c.Services()` 使用，异步业务仍由应用管理。

`NewTheme` 从完整 `Palettes` 生成角色；改变色板后重新生成明暗主题。直接修改 `Theme.Primary` 只修改该角色，不会反向修改 Palette 或其他同色角色。不要将 `App` 返回的已处理偏好的快照重新存入 config，否则文本缩放会被重复应用。

| 配置 | 默认及行为 |
| --- | --- |
| `Palettes` | primary/success=green，secondary/info=blue，warning=yellow，error=red，neutral=slate |
| `Palette.Shade` | 50、100…900、950；内置 26 套（17 套彩色、9 套中性色），每套 11 级；非法色阶 panic |
| 强调角色 | 浅色 500、深色 400；另有 Neutral 独立配方 |
| 文字 | `Dimmed/Muted/Toned/Text/Highlighted/Inverted` |
| 表面 | `Background/MutedBackground/Elevated/Accented/InvertedBackground` |
| 边框 | `Border/BorderMuted/BorderAccented/BorderInverted` |
| 字体 | Windows Segoe UI / Consolas；其他平台 system-ui / monospace，不捆绑字体 |
| `FontSize/TextScale` | 基础字号 14；组件独立字号表按基准比例缩放；App 再乘系统 TextScale |
| `Density` | 默认 1，只缩放相关间距/内边距；文字不随密度缩小；<=0 作为 1 |
| `Radius` | 基准 6 DIP；0 可得到直角；组件按自身配方比例取值 |
| `ContainerWidth` | 默认 1280 DIP |
| `Shadow` | 可复用 `ShadowTokens{Color, Blur, OffsetY}`；通过部件 `Shadow: nui.Set(theme.Shadow)` 应用，Card 默认无外阴影 |
| `Motion` | 150ms 背景过渡、1s loading 旋转、2s Skeleton 周期；`Reduce` 关闭动效 |

系统 ReduceMotion 与主题 Reduce 按 OR 合并，实例 `Animate: Set(true)` 不会覆盖系统的减少动态效果设置。高对比模式强化边框、文字、焦点，并将强调色改为浅色 700 / 深色 300；默认色板的 solid 正文对比通过 4.5:1 检查。这不代表任意自定义色板、状态、屏幕阅读器或整个应用已通过无障碍认证。

### 色板与运行时切换

内置色板从固定的 Tailwind CSS v4.3.3 数据生成，保留 OKLCH 色值。每套提供同名构造函数（如 `OrangePalette()`、`MauvePalette()`）；也可以通过 `PaletteByName("violet")` 查询，返回 `(Palette, bool)`，名称为小写，未知名称返回零值和 `false`。

- `PrimaryPaletteNames()`：red、orange、amber、yellow、lime、green、emerald、teal、cyan、sky、blue、indigo、violet、purple、fuchsia、pink、rose。
- `NeutralPaletteNames()`：slate、gray、zinc、neutral、stone、taupe、mauve、mist、olive。

这两个列表按菜单展示顺序返回调用方独有的切片；分组用于选色，不限制色板可以赋给哪些语义角色。`NeutralPalette()` 是名为 neutral 的灰色色板，与 `Palettes.Neutral` 这个语义角色不同。

运行时在应用状态中保存所选色板，修改后重新生成 `config.Light` 和 `config.Dark`，再由 `App` 应用。Primary 与 Neutral 独立变化，其他语义角色保持自己的色板。Gallery 顶部的 **Theme → Primary / Neutral** 提供色点、当前选项和可滚动子菜单，**Light / Dark / System** 切换外观；“Theme & buttons”页提供同一组选色入口。菜单支持方向键、Home / End、键入名称和 Escape，选择后恢复触发按钮焦点。选择按窗口保存，重启 Gallery 后恢复默认。

Gallery 还提供与 Nuxt UI 一致的 **Black** 主色：浅色为黑、深色为白，在系统高对比设置下也保持黑白。它是模式相关的覆盖，不是 11 级 Tailwind 色板，因此不在 `PrimaryPaletteNames()` 中，`PaletteByName("black")` 也不会返回色板。示例实现见 `internal/gallery/theme.go`。

## 样式覆盖

优先级为：组件配方 → 窗口/局部组件默认值和 UI → 实例显式 Color/Variant/Size → 实例 UI → 同步 Style 回调 → 返回元素上的后置修改。

Color/Variant 覆盖相关背景、前景、边框；Size 覆盖配方的字号、行高、字重、圆角、padding、gap。显式语义选项不会清除无关的全局部件字段。尺寸之外的宽高通过 Options 或返回值调整。`PartStyle` 支持前景、背景、边框、圆角、透明度、字体、字号、字重、行高、padding、gap、阴影。

```go
local := theme.WithComponent(nui.ComponentButton, nui.ComponentTheme{
    Color: nui.Secondary, Variant: nui.Soft,
    UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999))}},
})
nui.Themed(c, local, func(inner nui.Theme) {
    nui.Button(c, inner, "局部默认", nui.ButtonOptions{})
    nui.Button(c, inner, "直角", nui.ButtonOptions{
        Color: nui.Error, Variant: nui.Outline,
        UI: nui.Styles{nui.Root: {
            Radius: nui.Set(float32(0)),
            PaddingX: nui.Set(float32(0)),
        }},
    })
})
```

`WithComponent` 复制映射，`Component` 返回防御性副本；局部修改不泄漏到兄弟或其他窗口。`Themed` 会创建一个有背景的 Column，可继续设置布局，并在构建后恢复外层原生主题。

`Optional[T]` 的零值代表未设置，`Set(0)`、`Set(false)`、`Set(ui.Transparent)` 代表真实覆盖；枚举中的 Default 代表继承。大部分组件通过嵌入的 `Decoration{UI, Style}` 接收部件覆盖：

```go
nui.Card(c, theme, nui.CardOptions{
    Title: "项目",
    Decoration: nui.Decoration{
        UI: nui.Styles{nui.Body: {PaddingY: nui.Set(float32(12))}},
        Style: func(slot nui.Slot, part ui.Element) {
            if slot == nui.Footer { part.Wrap() }
        },
    },
}, func() { /* 当前构建中的正文 */ })
```

Button 提供 `Style(root, label, leading, trailing ui.Element)` 和直接 `UI` 字段。不存在的部件是零 Element；不要保存任何 Element、Context 或部件以跨帧使用。低层回调能修改 MyGo 元素，调用方有责任保留角色、名称和交互语义。

## 组件与部件

表中 slots 仅列实际支持的部件；没有列出的 slot 不产生效果。

| 条目 / 入口 | 选项、范围与 slots | 示例 |
| --- | --- | --- |
| App / `App`、`AppConfig` | 窗口配置、light/dark/system、系统偏好；不创建窗口 | 每个 Gallery 页面、minimal |
| Theme / `Theme`、`Window`、`Themed` | 色阶、语义、参数、组件默认、局部作用域 | Theme & buttons |
| Container / `Container` | `MaxWidth/Padding Optional[float32]`，居中、填充；Root | Workspace、minimal |
| Main / `Main` | Grow 主内容区、最小宽高 0；不是滚动器，外层用 `ui.Scroll`；全局 Root 覆盖，返回值可继续配置 | Workspace |
| Icon / `Icon` | `*ui.SVG`，可选尺寸/颜色/Label；无 Label 时装饰性；Root | Workspace、所有操作 |
| Button / `Button` | 六变体、五尺寸、七语义色、图标、Loading/Disabled/Block/Square；Root/Label/Leading/Trailing | Theme & buttons |
| Link / `Link` | URL 或应用 OnClick，Active、Color、Disabled，原生链接角色和键盘焦点；Root/Label | Workspace、Display |
| Badge / `Badge` | solid/outline/soft/subtle，五尺寸、颜色、Square、两侧图标；Root/Label/Leading/Trailing | Display、Workspace |
| Card / `Card` | solid/outline/soft/subtle；默认 outline；Title/Description 或 Header builder、正文、Footer builder；Root/Header/Body/Footer/Label/Description | Display、Workspace |
| Alert / `Alert` | 四种同上变体、默认 solid；图标、标题、长说明、Actions、可选 OnClose；Root/Leading/Label/Description/Trailing | Feedback |
| Avatar / `Avatar` | 已解码 Bitmap → SVG → Text 字面文字或 Alt 首字母回退，Alt、Color、五尺寸；Root/Label/Leading | Display、Workspace |
| AvatarGroup / `AvatarGroup` | 稳定 ID、统一尺寸、重叠、Max 未设置时不限数量、剩余数量；Root；单项 AvatarOptions 可配置 | Display |
| Chip / `Chip` | 四角、Inset、Hidden、颜色/尺寸、可选文字；Root/Indicator；忽略输入不抢宿主点击；无 content 时独立显示 | Display |
| Separator / `Separator` | 横/纵、Label、Icon；Root/Label/Indicator；纵向默认高 48，可链式修改 | Display |
| Kbd / `Kbd`、`FormatKey` | 多个字符串键、平台修饰键、尺寸/色/变体；Root/Label；只显示提示，不注册快捷键 | Display Card footer |
| Skeleton / `Skeleton` | Width/Height、圆角、可选动画；Root；默认 160×16 | Feedback |
| Progress / `Progress` | 确定/不确定、横/纵、五档粗细、颜色、Max、Label、动画；Root；Indicator 的 Background 定义自绘填色 | Feedback |
| ProgressGroup / `ProgressGroup` | 横向多段进度、稳定 ID、段颜色、Max、尺寸、说明列表可隐藏；Root/Indicator（轨道）；不承担任务调度 | Feedback、Workspace |
| Empty / `Empty` | 四种表面变体、尺寸、图标/标题/说明/Actions/Footer；Root/Leading/Label/Description | Workspace |
| User / `User` | Avatar、Name/Description、尺寸、Vertical、Disabled、OnActivate；Root/Label/Description；长文本省略 | Display、Workspace |

Separator 支持 Color、Size（XS..XL 对应 1..5 DIP）、Type（SeparatorSolid/Dashed/Dotted）、Position（Center/Start/End）、Avatar、Content、Orientation。内容优先级为 Content → Label → Icon → Avatar；Body 为内容容器，Indicator 为分隔线。默认标签 14 DIP、图标 20 DIP；Avatar 默认映射到本库最小 XS（上游 2xs 无同名尺寸）。

Empty 支持 Loading/LoadingIcon，None 对应无表面样式。User 的 XS 尺寸使用内联描述。

### 操作组件

Button 变体：`Solid/Outline/Soft/Subtle/Ghost/LinkVariant`；尺寸 `XS/SM/MD/LG/XL` 默认基准高度为 **24/28/32/36/40 DIP**，原生字体或回退字体的行高较大时会随文字增高。`Neutral` bool 保留兼容，推荐 `Color: nui.Neutral`。Color 取 `Primary/Secondary/Success/Info/Warning/Error/Neutral`。

Loading 阻止点击与键盘激活。默认在 leading 位置替换或新增旋转图标，标签始终显示；`LoadingTrailing` 可放到 trailing（显式 Leading 优先），`LoadingIcon` 自定义图标。没有原图标时按钮宽度会增加。`Active/ActiveColor/ActiveVariant` 控制活动态；仅图标按钮自动使用 square，Square 的 padding 允许实例 UI 最后覆盖。图标按钮要求 `AccessibleLabel`，否则 panic。动画 spinner 的颜色由 Root / Leading 前景字段解析，需覆盖颜色时使用 UI；低层 `Style(...leading).TextColor` 不能重新着色已解析的自绘 spinner。

MyGo 当前将 Disabled 绘为 50% 透明，本库不再叠乘 75%。disabled/loading 不重新启用父级禁用。装饰 Chip 没有操作角色；有业务意义的数量或在线状态应加入宿主的 AccessibleLabel/Alt。

Link URL 非空时在激活后调用 `Services.OpenURL`，不会在构建中打开网页。URL 为空时使用返回元素 `.OnClick`，由应用实现导航；同时再设置 OnClick 会遵循 MyGo 的动作覆盖规则。User 只有 OnActivate 非空时才是可键盘激活的按钮。

### 图片、标识与资源

Avatar 不执行磁盘或网络 I/O。调用方在构建之外执行 `ui.DecodeBitmap`；失败后传 nil，即使用 Icon 或 Text。远程下载、取消、缓存和错误状态属于应用资源层。`Text` 按字面显示（例如 `+12`）；没有 Text 时从 Alt 前两个空格分词各取首字符，保留大小写；没有 Alt 时为空。会缩小文字以适应圆形内部，未实现完整 Unicode grapheme 缩写算法。姓名自动缩写通过 Alt 配置，User 的 Name 默认作为头像 Alt。

AvatarGroup 和 ProgressGroup 要求唯一、非空 ID，不用 label 作为身份；AvatarGroup 未设置 Max 或 Max≤0 时全部显示；正数限制可见数量并显示同尺寸的 `+n` Avatar。组的 Size/Color 只为未显式指定的子头像提供默认值。`Overlap: Set(0)` 可关闭重叠。

Lucide 已完整收录固定提交 `a630cec77eac6568b989e87e4947a8bfe45caacd` 的 **2,015 个 SVG**，各自提供独立 getter，例如 `lucide.House()`、`lucide.GitBranch()`、`lucide.FileJson2()`、`lucide.RobotArm()`。命名由连字符分段后各段首字母大写，数字保持不变；完整名称见 `icons/lucide/manifest.json` 和 Go 文档。

每个 SVG 独立 embed 和 `sync.Once` 缓存，未引用 getter 可由 Go 链接器移除；没有在运行时注册或预解析全部图标。返回的共享 SVG 只读使用，尺寸／颜色通过 `ui.Icon` 或组件选项设置。全量 getter 清单仅生成在 `_test.go` 内，避免验证代码影响消费者体积。

升级图标时获取指定 Lucide 提交的完整 `icons/*.svg`，原样收录并更新 `icons/lucide/manifest.json` 的提交、名称和 SHA-256；同步 LICENSE 后运行 `go generate ./icons/lucide`。生成器离线校验全部文件、未登记 SVG、重复名称、Go 名称／文件冲突和旧生成文件，验证通过后再写出结果。平台后缀会自动规避，例如 `robot-arm` 生成 `robot_arm_icon.go`，保证 amd64 也能使用。`go generate ./...` 不联网；色板也由固定 Tailwind 4.3.3 文件离线生成。更多用法见 [图标包说明](../icons/lucide/README.md)。

### 进度、动画与边界

Progress 的 Max 默认为 100，显式非正/非有限 Max 为程序配置错误并 panic；值为 NaN/Inf 时按 0，负数归 0，超过 Max 归 Max。不确定状态使用 MyGo 的 `Now=-1 < Min=0` 哨兵，另提供 “In progress” 值/描述；减少动态效果时显示静态短段。Windows UIA 的 range getter 会将未知值夹到下界，这是当前 MyGo 的平台行为，未知进度的屏幕阅读器呈现尚未验证。纵向默认高 160 DIP，正常从顶部填充，Inverted 从底部填充；水平从逻辑 start 填充，Inverted 反向。Orientation 和 Vertical 均可设置方向。

ProgressGroup 将各段绘制在同一条多段轨道上。总和未满 Max 时显示余量；超出时按比例缩到同一轨道，避免溢出；可访问总值限制在 Max 内。先将每段限制在 0..Max，再按比例缩短超出的总宽度，避免浮点溢出；非有限值归 0。列表显示各段相对于 Max 的四舍五入百分比；没有标签时不创建说明列表。水平方向随 RTL 镜像。支持 Orientation（默认 Horizontal，纵向 Height 默认 160）；分段交互和任意段 builder 仍未提供。Alert 的方向参数等扩展见 [参数与默认值](options.md)。

持续动画只在 Painter 实际绘制可见组件时请求帧；移出滚动可见区、卸载、关闭 Animate 或 ReduceMotion 后应静止。Button 背景过渡单独受 `Motion.Transition/Reduce` 控制，`Animate:false` 控制 spinner，有限过渡完成后仍会静止。没有定时 goroutine。

## 使用边界

采用原生 Go 配方与部件 API；不支持 CSS selector 或 Vue 插槽。组件间的功能差异、原生无障碍边界和平台状态见 [兼容性说明](compatibility.md)。
