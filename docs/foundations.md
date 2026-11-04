# 主题与基础组件

先用 `App` 为窗口设置主题，再将返回的 `Theme` 传给组件。本页介绍主题配置，以及按钮、卡片、头像、进度等基础组件。完整启动代码见[最小应用](../examples/minimal/main.go)。

| 需要做什么 | 阅读位置 |
| --- | --- |
| 设置明暗模式、主色和字号 | [应用与主题](#应用与主题) |
| 修改某类组件或单个部件的外观 | [样式覆盖](#样式覆盖) |
| 查找基础组件 | [组件与部件](#组件与部件) |
| 添加按钮、链接或头像 | [操作组件](#操作组件)、[图片标识与资源](#图片标识与资源) |
| 展示进度和加载状态 | [进度动画与边界](#进度动画与边界) |

## 应用与主题

### 创建配置，再在视图中使用

在创建窗口时初始化配置，并将它保存在应用模型中。下面把主色设为蓝色，中性色设为 zinc：

```go
config := nui.DefaultApp()
config.Mode = nui.SystemMode // 也可使用 LightMode 或 DarkMode

palettes := nui.DefaultPalettes()
palettes.Primary = nui.BluePalette()
palettes.Neutral = nui.ZincPalette()
config.Light = nui.NewTheme(false, palettes)
config.Dark = nui.NewTheme(true, palettes)
config.Light.Radius, config.Dark.Radius = 8, 8
config.Light.Density, config.Dark.Density = .9, .9
```

在视图函数开头调用 `App`。每次构建都要调用，但不要重新初始化上面的配置，否则用户运行时所做的修改会被覆盖。

```go
theme := nui.App(c, config)
nui.Button(c.Key("save"), theme, "保存", nui.ButtonOptions{
    Color: nui.Primary,
    Variant: nui.Solid,
    Size: nui.MD,
    Leading: lucide.Check(),
}).OnClick(save)
```

`App` 根据配置和系统偏好设置当前窗口的原生主题。它要在本次构建的 `Window`、`Themed` 调用之前执行，以便读取系统外观。MyGo 窗口仍通过 `mygo.NewWindow` 创建；通知需要另设 `ToastViewport`。

`App` 返回的是已应用系统偏好的主题。**不要再把它写回 `config.Light` 或 `config.Dark`**，否则文本缩放可能被重复应用。需要完全自行控制主题时，可使用底层 `Window(c, theme)`；它不会自动应用系统偏好。

### 常用主题参数

`NewTheme` 根据整套 `Palettes` 生成不同用途的颜色，例如主色、文字色和背景色。修改色板后，应重新生成浅色和深色主题。直接改 `Theme.Primary` 只会改变该字段，不会同步修改色板和其他相关颜色。

| 参数 | 默认值和用途 |
| --- | --- |
| `Mode` | `SystemMode` 跟随系统；也可选 `LightMode` 或 `DarkMode`。零值 `AppConfig` 同样跟随系统 |
| `FontSize` / `TextScale` | 基础字号 14；组件字号按基准比例缩放，`App` 再应用系统文本缩放 |
| `Density` | 默认 1，调整相关间距和内边距，不缩小文字；小于等于 0 时按 1 处理 |
| `Radius` | 基准圆角 6 DIP；设为 0 可使用直角，各组件按自身样式比例取值 |
| `ContainerWidth` | 默认最大内容宽度 1280 DIP |
| `Shadow` | `ShadowTokens{Color, Blur, OffsetY}`；通过部件的 `Shadow: nui.Set(theme.Shadow)` 使用，Card 默认没有外阴影 |
| `Motion` | 默认背景过渡 150ms、加载图标旋转周期 1s、Skeleton 周期 2s；`Reduce` 关闭动画 |

Windows 默认使用 Segoe UI / Consolas，其他平台使用 system-ui / monospace。本库不捆绑字体；字体和回退字体可能影响实际控件高度。

### 颜色字段

| 用途 | 字段或默认色板 |
| --- | --- |
| 语义颜色 | Primary、Success 默认 green；Secondary、Info 默认 blue；Warning 为 yellow；Error 为 red；Neutral 为 slate |
| 强调色 | 浅色通常使用 500 色阶，深色使用 400；Neutral 有独立的取色规则 |
| 文字 | `Dimmed`、`Muted`、`Toned`、`Text`、`Highlighted`、`Inverted` |
| 背景 | `Background`、`MutedBackground`、`Elevated`、`Accented`、`InvertedBackground` |
| 边框 | `Border`、`BorderMuted`、`BorderAccented`、`BorderInverted` |

系统或主题任一处启用减少动态效果，动画都会关闭，组件的 `Animate: nui.Set(true)` 无法覆盖这一设置。高对比模式会增强文字、边框和焦点提示，并将强调色改为浅色 700 / 深色 300。

默认色板的 Solid 正文颜色已通过 4.5:1 对比度检查。自定义色板、其他状态和完整应用的无障碍体验仍需分别验证。

### 色板与运行时切换

内置 26 套色板来自固定的 Tailwind CSS v4.3.3 数据，保留 OKLCH 色值。每套有 11 个色阶：50、100、200……900、950。`Palette.Shade` 接收这些色阶，传入其他值会 panic。

可直接调用 `OrangePalette()`、`MauvePalette()` 等函数，也可用 `PaletteByName("violet")` 查询。名称使用小写，函数返回 `(Palette, bool)`；未知名称返回零值和 `false`。

| 查询函数 | 返回的名称 |
| --- | --- |
| `PrimaryPaletteNames()` | red、orange、amber、yellow、lime、green、emerald、teal、cyan、sky、blue、indigo、violet、purple、fuchsia、pink、rose |
| `NeutralPaletteNames()` | slate、gray、zinc、neutral、stone、taupe、mauve、mist、olive |

名称列表按菜单顺序返回独立切片。分组仅用于选色，不限制色板能赋给哪个颜色字段。`NeutralPalette()` 返回名为 neutral 的灰色色板；`Palettes.Neutral` 则是主题中用于中性色的字段。

运行时切换颜色的步骤是：

1. 在应用状态中保存用户选择的色板。
2. 用新色板重新生成 `config.Light` 和 `config.Dark`。
3. 在后续构建中继续调用 `App(c, config)`。

Primary 和 Neutral 可以分别修改，其他语义颜色使用各自的色板。Gallery 的 **Theme → Primary / Neutral** 提供完整示例，**Light / Dark / System** 切换明暗模式。菜单支持方向键、Home、End、输入名称和 Escape，选择后恢复触发按钮焦点。选择按窗口保存，重启 Gallery 后恢复默认。

### 使用黑白主色

`WithBlackPrimary()` 让主色在浅色模式下为黑、深色模式下为白，高对比模式也保持黑白。先创建两种主题，再分别调用：

```go
config := nui.DefaultApp()
config.Light = config.Light.WithBlackPrimary()
config.Dark = config.Dark.WithBlackPrimary()
```

方法返回新主题，同时更新 Primary 及其色阶，其他颜色和样式配置保持原值。Gallery 和网站预览的 Primary 菜单均有 Black 选项。

Black 随明暗模式变化，没有独立的 11 级色板，因此不出现在 `PrimaryPaletteNames()` 中，`PaletteByName("black")` 也不会返回色板。切回彩色时，使用所选色板重新生成明暗主题。

## 样式覆盖

### 先选择修改范围

只改一个组件时，在该次调用的 Options 中设置颜色、尺寸或 `UI`。需要让一组组件共用样式时，用 `WithComponent` 创建局部主题，再通过 `Themed` 使用它：

```go
local := theme.WithComponent(nui.ComponentButton, nui.ComponentTheme{
    Color: nui.Secondary,
    Variant: nui.Soft,
    UI: nui.Styles{nui.Root: {Radius: nui.Set(float32(999))}},
})
nui.Themed(c, local, func(inner nui.Theme) {
    nui.Button(c, inner, "局部默认", nui.ButtonOptions{})
    nui.Button(c, inner, "直角", nui.ButtonOptions{
        Color: nui.Error,
        Variant: nui.Outline,
        UI: nui.Styles{nui.Root: {
            Radius: nui.Set(float32(0)),
            PaddingX: nui.Set(float32(0)),
        }},
    })
})
```

`WithComponent` 复制配置映射，`Component` 返回配置副本，修改局部主题不会影响兄弟区域或其他窗口。`Themed` 自身会创建一个有背景的 Column；可继续设置它的布局，子内容构建结束后会恢复外层原生主题。

### 覆盖顺序

后应用的设置覆盖前面同一字段的值：

```text
组件内置样式
  → 窗口或局部主题中的组件默认值和 UI
  → 当前实例显式设置的 Color / Variant / Size
  → 当前实例的 UI
  → 同步 Style 回调
  → 返回元素上的后续修改
```

`Color`、`Variant` 调整相关背景、前景和边框；`Size` 调整字号、行高、字重、圆角、内边距和间距。它们只覆盖相关字段，不会清除其他部件设置。整体宽高可通过 Options 或返回元素调整。

`PartStyle` 支持前景、背景、边框、圆角、透明度、字体、字号、字重、行高、内边距、间距和阴影。`Optional[T]{}` 表示未设置，`Set(0)`、`Set(false)`、`Set(ui.Transparent)` 都是有效覆盖；枚举中的 Default 通常表示继承。

### 修改单个部件

多数组件通过 `Decoration{UI, Style}` 接收部件样式。`UI` 按部件名设置属性；需要直接配置 MyGo 元素时再使用 `Style`：

```go
nui.Card(c, theme, nui.CardOptions{
    Title: "项目",
    Decoration: nui.Decoration{
        UI: nui.Styles{nui.Body: {PaddingY: nui.Set(float32(12))}},
        Style: func(slot nui.Slot, part ui.Element) {
            if slot == nui.Footer {
                part.Wrap()
            }
        },
    },
}, func() {
    ui.Text(c, "项目说明")
})
```

Button 直接提供 `UI` 字段和 `Style(root, label, leading, trailing ui.Element)`。未创建的部件以零值 Element 传入。回调只配置当前构建中的元素，不要保存 Element、部件或 Context 供之后使用；自定义交互时还需保留控件名称、角色和焦点行为。

## 组件与部件

先按用途选组件。下面的部件名可用于 `UI`；未列出的部件不会产生样式效果。

### 页面结构与操作

| 组件 | 用途与主要选项 | 可用部件 |
| --- | --- | --- |
| App / Theme | `AppConfig` 管理明暗模式；`Theme` 保存颜色、尺寸与组件默认值 | 通过各组件配置 |
| Container | 居中并限制内容宽度；`MaxWidth`、`Padding` 为 `Optional[float32]` | Root |
| Main | 可伸展的主内容区，最小宽高为 0；需要滚动时在外层使用 `ui.Scroll` | 全局 Root，返回值可继续配置 |
| Icon | 显示 `*ui.SVG`，支持尺寸、颜色和 Label；无 Label 时作为装饰图标 | Root |
| Button | 按钮，支持颜色、变体、尺寸、图标、加载和禁用 | Root、Label、Leading、Trailing |
| Link | 打开 URL 或执行应用导航，支持 Active、Color、Disabled | Root、Label |

### 信息展示

| 组件 | 用途与主要选项 | 可用部件 |
| --- | --- | --- |
| Badge | 短标签；支持 Solid/Outline/Soft/Subtle、颜色、尺寸、Square 和两侧图标 | Root、Label、Leading、Trailing |
| Card | 卡片；默认 Outline，另支持 Solid/Soft/Subtle；标题、说明、Header、正文和 Footer | Root、Header、Body、Footer、Label、Description |
| Alert | 提示信息；默认 Solid，支持与 Card 相同的四种变体、图标、Actions 和 OnClose | Root、Leading、Label、Description、Trailing |
| Empty | 空状态；图标、标题、说明、Actions、Footer、Loading/LoadingIcon；`None` 去掉表面样式 | Root、Leading、Label、Description |
| User | 头像、姓名和说明；支持 Vertical、Disabled、OnActivate，长文字省略 | Root、Label、Description |

Badge 的标签为空时不创建文字，也不保留多余间距。在 FieldGroup 或 FormField 内会继承尺寸，并参与组的圆角处理。

Card 的正文回调为 nil 时不创建正文区域；分隔线只出现在实际存在的区域之间。Alert 的关闭按钮使用 MD、LinkVariant，无内边距和悬停底色，沿用提示的文字颜色，悬停时变淡。方向设置见[参数与默认值](options.md#orientation)。

User 的 `AvatarOptions` 全为空时不显示头像。需要首字母头像时，设置 `Avatar: nui.AvatarOptions{Alt: name}` 或 `Text`。XS 尺寸的说明与姓名显示在同一行；只有传入 `OnActivate` 时，User 才作为可键盘激活的按钮。

### 标识、分隔与加载

| 组件 | 用途与主要选项 | 可用部件 |
| --- | --- | --- |
| Avatar / AvatarGroup | 单个头像或重叠头像组；组项目使用稳定 ID，可限制数量 | Avatar：Root、Label、Leading；组：Root |
| Chip | 角标，支持四角位置、Inset、Hidden、颜色、尺寸和文字；不拦截宿主点击，无正文时可独立显示 | Root、Indicator |
| Separator | 横向或纵向分隔线，可插入文字、图标、头像或自定义内容 | Root、Body、Label、Indicator |
| Kbd / FormatKey | 显示快捷键提示和平台修饰键，不注册快捷键 | Root、Label |
| Skeleton | 加载占位，默认 160×16 DIP，可设宽高、圆角和动画 | Root |
| Progress / ProgressGroup | 单项或分段进度，支持尺寸、颜色、方向和动画 | Root、Indicator，详见进度说明 |

Separator 的 `Size` 从 XS 到 XL 对应 1–5 DIP 线宽；`Type` 可选 Solid、Dashed、Dotted 对应常量，`Position` 可选 Center、Start、End。内容优先级为 `Content → Label → Icon → Avatar`，默认标签 14 DIP、图标 20 DIP、头像 XS。纵向默认高度 48 DIP，可通过返回值修改。

Kbd 将普通键名转为大写，并把 Tab、Backspace、Delete、CapsLock、方向键、PageUp/PageDown、Home/End 显示为相应符号。修饰键按桌面习惯显示，例如非 macOS 平台上的 meta 为 Win、shift 为 Shift。

### 操作组件

Button 支持 `Solid`、`Outline`、`Soft`、`Subtle`、`Ghost`、`LinkVariant` 六种外观。颜色可选 `Primary`、`Secondary`、`Success`、`Info`、`Warning`、`Error`、`Neutral`；旧的 `Neutral bool` 字段保留兼容，新代码使用 `Color: nui.Neutral`。

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| 默认基准高度（DIP） | 24 | 28 | 32 | 36 | 40 |

当字体行高较大时，按钮会随文字增高。`Block` 控制填满宽度，`Square` 控制方形布局；仅图标按钮会自动采用方形布局，且必须提供 `AccessibleLabel`，否则 panic。实例 `UI` 可以继续覆盖 Square 的内边距。

**加载与禁用：** `Loading` 阻止点击和键盘激活，保留标签，默认在前侧替换或新增旋转图标。`LoadingTrailing` 可将它放到后侧，但显式 Leading 优先；`LoadingIcon` 可更换图标。原来没有图标时，进入加载状态可能增加按钮宽度。

加载图标颜色从 Root / Leading 的前景样式读取，需要改色时设置 `UI`；最后执行的 `Style(...leading).TextColor` 无法改变已解析的自绘图标颜色。`Active`、`ActiveColor`、`ActiveVariant` 控制活动状态。

MyGo 将禁用控件绘制为 50% 透明度，本库不再叠加透明度。子组件设置 `Disabled: false` 无法解除父级禁用。Chip 只有装饰作用，业务上有意义的数量或在线状态应补充到宿主的 `AccessibleLabel` 或 `Alt`。

**链接：** Link 的 URL 非空时，在激活后通过 `Services.OpenURL` 打开地址。URL 为空时，可用返回元素的 `.OnClick` 实现应用内导航；同时设置 URL 和 OnClick 时要注意 MyGo 的动作覆盖规则。

### 图片、标识与资源

Avatar 按 `Bitmap → Icon → Text → Alt 首字母` 的顺序选择内容。图片要在视图构建之外解码，再传入 `ui.DecodeBitmap` 的结果；失败时传 nil，组件会继续使用后面的内容。下载、缓存、取消和错误处理由应用负责。

`Text` 原样显示，例如 `+12`。没有 Text 时，Avatar 从 Alt 的前两个空格分词中各取首字符，保留大小写；没有 Alt 时留空。文字会缩小以适应圆形，但未实现完整的 Unicode 字素缩写算法。User 的 Name 默认作为头像 Alt。

AvatarGroup 和 ProgressGroup 的每项都需要唯一、非空 ID。AvatarGroup 未设置 Max 或 Max≤0 时显示全部头像；正数限制可见数量，并用同尺寸 `+n` 头像表示剩余人数。组的 Size / Color 只给未显式配置的子头像提供默认值，`Overlap: nui.Set(0)` 可关闭重叠。

Lucide 图标使用独立函数，例如 `lucide.House()`、`lucide.GitBranch()`、`lucide.FileJson2()`。本库收录固定提交中的 2,015 个 SVG，每个图标独立嵌入和缓存，未引用的图标可由 Go 链接器移除。返回的 SVG 是共享只读对象，颜色和尺寸通过组件设置。

完整名称、固定提交、导入方式和离线生成步骤见[图标文档](../icons/lucide/README.md)及 [manifest.json](../icons/lucide/manifest.json)。图标升级和生成属于库维护工作，普通应用直接调用图标函数即可。

### 进度、动画与边界

**Progress** 的 Max 默认 100。显式设置非正数、NaN 或 Inf 会 panic；进度值的 NaN/Inf 按 0 处理，再限制在 0 到 Max 之间。

确定进度按数值填充，不确定进度显示移动短段；减少动态效果时短段静止。底层使用 `Now=-1 < Min=0` 表示未知进度，并提供 In progress 的值或描述。Windows UIA 会将未知数值限制到下界，具体读屏呈现尚未验证。

水平方向从文字起始侧填充，纵向默认高 160 DIP、从顶部填充；`Inverted` 反转填充方向。`Orientation` 和 `Vertical` 均可控制方向。Indicator 的 `Background` 设置自绘进度颜色。

**ProgressGroup** 将多段进度放在同一轨道上。每段先限制在 0 到 Max，非有限值按 0 处理；总和不足时保留空余，超出时按比例缩短，辅助技术总值也限制在 Max 内。

说明列表显示各段相对于 Max 的四舍五入百分比，没有标签时不创建列表。水平方向随 RTL 调整；Orientation 默认 Horizontal，纵向 Height 默认 160 DIP。当前不提供分段点击和任意分段内容回调，也不负责调度任务。

**动画何时停止：** 持续动画只在可见组件实际绘制时请求下一帧。移出滚动可见区、卸载、关闭 Animate 或启用减少动态效果后停止。Button 的背景过渡由 `Motion.Transition/Reduce` 控制，`Animate: nui.Set(false)` 控制加载图标；过渡结束后不再持续请求帧。组件没有用于动画的定时 goroutine。

## 使用边界

样式通过 Go 选项和部件 API 设置，不支持 CSS 选择器或 Vue 插槽。平台状态、输入法和辅助技术的验证范围见[兼容性说明](compatibility.md)。
