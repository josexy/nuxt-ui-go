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

启用高对比模式后，默认色板的 Solid 正文颜色已通过 4.5:1 对比度检查。该结果不适用于普通模式：例如浅色默认主色与白色文字的对比度约为 2.22:1。自定义色板、其他状态和完整应用的无障碍体验仍需分别验证。

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

Nuxt UI 的 `Theme.props` 可以为组件设置任意公开参数默认值。Go 的 `ComponentTheme` 只包含 Color、Size、Variant 和 UI；加载状态、禁用状态、延迟等行为参数通过各组件 Options 或应用自己的构建函数统一设置。Go 需要显式传递主题，`Themed` 还会创建布局容器，这两点与 Vue 的透明 provider 不同。

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
| Button | 按钮，支持颜色、变体、尺寸、图标、头像、加载和禁用 | Root、Label、Leading、Trailing |
| Link | 打开 URL 或执行应用导航，支持 Active、Color、Disabled | Root、Label |

Container 的默认水平内边距按窗口宽度调整：小于 640 DIP 为 16，640–1023 DIP 为 24，至少 1024 DIP 为 32，再乘主题 Density。断点依据窗口宽度，嵌套在较窄父级中不会另算断点。显式 `Padding` 使用固定 DIP 值，`nui.Set(0)` 去掉内边距；实例 UI 和 Style 仍可覆盖。

Button 的 `Avatar: &nui.AvatarOptions{Alt: "Ada Lovelace"}` 显示前导头像。前导图标或前导加载图标优先；尾部加载图标不隐藏头像。头像未设置 Size 时按按钮 XS–XL 使用 16 / 16 / 20 / 20 / 24 DIP，显式 Avatar.Size 优先。纯头像按钮仍需 AccessibleLabel。Leading 部件在此模式下对应头像根元素。

加载图标使用独立的绘制回调旋转。设置其颜色时使用 `UI[Leading].Foreground` 或 `UI[Trailing].Foreground`；Style 回调中的 TextColor 不会改写绘制回调已解析的颜色。其他布局和尺寸仍可通过 Style 设置。

### 信息展示

| 组件 | 用途与主要选项 | 可用部件 |
| --- | --- | --- |
| Badge | 短标签；支持 Solid/Outline/Soft/Subtle、颜色、尺寸、Square、两侧图标和前导头像 | Root、Label、Leading、Trailing |
| Card | 卡片；默认 Outline，另支持 Solid/Soft/Subtle；标题、说明、Header、正文和 Footer | Root、Header、Body、Footer、Label、Description |
| Alert | 提示信息；默认 Solid，支持四种变体、图标、头像、Actions 及关闭配置 | Root、Body、Leading、Label、Description、Actions、Trailing |
| Empty | 空状态；头像/图标、标题、说明、内容构建入口及 Loading/LoadingIcon；`None` 对应 naked 外观 | Root、Header、Leading、Label、Description、Body、Actions、Footer |
| User | 头像、角标、姓名和说明；支持 Orientation / Vertical、Disabled、OnActivate，长文字省略 | Root、Leading、Body、Label、Description |

Badge 的标签为空时不创建文字，也不保留多余间距，并自动使用四边相同的内边距。在 FieldGroup 或 FormField 内会继承尺寸，并参与组的圆角处理。Avatar 未设置 Size 时，标签 XS–XL 的头像直径为 16 / 16 / 16 / 20 / 20 DIP；显式 Leading 图标优先于头像。

Card 的正文回调为 nil 时不创建正文区域；分隔线只出现在实际存在的区域之间。窗口宽度小于 640 DIP 时，卡片各区域内边距为 16 DIP；达到 640 后，水平内边距和正文垂直内边距为 24，页头和页尾垂直内边距保持 16，均受 Density 影响。只有描述而无标题时，描述仍保留 4 DIP 顶部间距。

页头、正文和页尾使用独立标识。动态移除页头不会重建仍存在的正文或页尾，内部输入可以保留焦点和选区；子控件仍需使用稳定 key 和持久应用数据。

Alert 的 Avatar 优先于 Icon，头像默认直径为 44 DIP；显式 Avatar.Size 覆盖该默认值。Actions 默认位于正文下方，顶部间距为 10 DIP；Horizontal 将操作区放到正文旁边。切换方向或前导内容会保留仍存在的操作控件身份。Body 对应文字容器，Actions 对应操作行，Trailing 对应关闭按钮。

OnClose 或 Close 非 nil 时显示关闭按钮。Close 使用 ButtonOptions 配置，默认 MD、Neutral、LinkVariant 和零内边距；默认外观沿用提示的文字颜色，悬停时变淡。CloseIcon 默认为 X，显式 Close.Leading 优先。每次激活只调用一次 OnClose，组件不会自行隐藏，应用负责更新可见状态。方向设置见[参数与默认值](options.md#orientation)。

User 的 `AvatarOptions` 全为空时不显示头像。需要首字母头像时，设置 `Avatar: nui.AvatarOptions{Alt: name}` 或 `Text`。XS 尺寸的说明与姓名显示在同一行；只有传入 `OnActivate` 时，User 才作为可键盘激活的按钮。

User.Size 决定头像与角标的尺寸，覆盖嵌套 Avatar.Size / Chip.Size；Chip.Inset 固定为 true，其他位置、颜色、文字和 Hidden 配置沿用 Chip。头像不存在时不显示角标。Leading 可替换默认头像区域，Body 可替换姓名与说明；各区域按稳定标识构建。

可激活的 User 在悬停或键盘焦点下强调文字，头像放大至 1.15 倍；缩放过渡遵循 Motion.Transition，减少动态效果时立即显示最终状态。禁用及父级禁用会阻止激活。OnActivate 不负责路由或自动打开 URL，相关操作由应用实现。

Empty 的默认内边距随窗口宽度采用 16 / 24 / 32 DIP，断点为 640 和 1024；页头、正文和页脚最大宽度为 384 DIP。头像默认随 Empty 的 XS–XL 使用 32/36/40/44/48 DIP，显式 Avatar.Size 优先。Avatar.Image 始终优先于图标，显式 Avatar.Icon 优先于 Empty.Icon 和加载图标。

Header 替换整个默认页头，Leading 只替换前导内容，Body 替换默认 Actions；Footer 独立显示。操作区和页脚使用稳定标识，页头变化不会替换其中的输入控件。默认标题提供二级标题语义，自定义 Header 的语义由应用配置。

Loading 只控制前导图标，不会禁用 Actions，也不改变已有头像图片。减少动态效果会停止旋转；原生没有等价的 aria-busy 元数据。加载图标颜色通过 UI.Leading.Foreground 设置，Style 的晚期 TextColor 不会替换已解析的自绘颜色。头像外观通过 Avatar.UI 配置，Leading 部件保留默认图标含义；自定义 Leading 时对应其容器。

### 标识、分隔与加载

| 组件 | 用途与主要选项 | 可用部件 |
| --- | --- | --- |
| Avatar / AvatarGroup | 单个头像或重叠头像组；组项目使用稳定 ID，可限制数量 | Avatar：Root、Body、Label、Leading；组：Root、Item |
| Chip | 角标，支持四角位置、Inset、Hidden、颜色、尺寸和文字；不拦截宿主点击，无正文时可独立显示 | Root、Indicator |
| Separator | 横向或纵向分隔线，可插入文字、图标、头像或自定义内容 | Root、Body、Label、Leading、Indicator |
| Kbd / FormatKey | 显示快捷键提示和平台修饰键，不注册快捷键 | Root、Label |
| Skeleton | 加载占位，默认 160×16 DIP，可设宽高、圆角和动画 | Root |
| Progress / ProgressGroup | 单项或分段进度，支持尺寸、颜色和方向；Progress 支持动画 | Progress：Root、Indicator；ProgressGroup 另有 Segment、List、Item、Leading、Label、Trailing |

Separator 的 `Size` 从 XS 到 XL 对应 1–5 DIP 线宽；`Type` 可选 Solid、Dashed、Dotted 对应常量，`Position` 可选 Center、Start、End。内容优先级为 `Content → Label → Icon → Avatar`，默认标签 14 DIP、图标 20 DIP、头像 20 DIP；显式 Avatar.Size 优先。Leading 对应图标或头像根，纵向标签允许换行。纵向默认高度 48 DIP，可通过返回值修改；长标签和交互内容应给足高度。内容容器的身份不随位置或方向变化，内部控件仍需稳定 key。

Separator 目前绘制装饰性分隔线，保留内部文字和交互内容各自的可访问性。MyGo 没有静态 separator 角色，不能与上游 `decorative=false` 的语义等同；也没有添加一个不生效的 Decorative 参数。虚线、点线的笔画颜色通过 `UI[Indicator].Background` 设置，最终 Style 的 Background 只改变元素背景。

Kbd 将普通键名转为大写，并把 Tab、Backspace、Delete、CapsLock、方向键、PageUp/PageDown、Home/End 显示为相应符号。修饰键按桌面习惯显示，例如非 macOS 平台上的 meta 为 Win、shift 为 Shift。

`command` 与 `win` 明确显示 ⌘、⊞，不随 Platform 改变；`CmdOrCtrl` / `primary` 在 macOS 上显示 ⌘，其他平台显示 Ctrl。Platform 默认为当前 GOOS，可显式指定 darwin、windows 或 linux 来展示目标平台。Kbd 不注册快捷键，也不进入默认 Tab 顺序。标准尺寸为 SM / MD / LG；XS、XL 是本库兼容的 SM、LG 尺寸别名。

Button、Badge、Alert、Kbd 的 Neutral / Solid 背景使用 Theme.InvertedBackground，文字使用 Inverted。Button 悬停时为该背景的 90% 透明度，高对比度模式保持不透明；实例 UI / Style 仍可覆盖默认外观。

### 操作组件

Button 支持 `Solid`、`Outline`、`Soft`、`Subtle`、`Ghost`、`LinkVariant` 六种外观。颜色可选 `Primary`、`Secondary`、`Success`、`Info`、`Warning`、`Error`、`Neutral`；旧的 `Neutral bool` 字段保留兼容，新代码使用 `Color: nui.Neutral`。

| Size | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| 默认基准高度（DIP） | 24 | 28 | 32 | 36 | 40 |

当字体行高较大时，按钮会随文字增高。`Block` 控制填满宽度，`Square` 控制方形布局；仅图标按钮会自动采用方形布局，且必须提供 `AccessibleLabel`，否则 panic。实例 `UI` 可以继续覆盖 Square 的内边距。

**加载与禁用：** `Loading` 阻止点击和键盘激活，保留标签，默认在前侧替换或新增旋转图标。`LoadingTrailing` 可将它放到后侧，但显式 Leading 优先；`LoadingIcon` 可更换图标。原来没有图标时，进入加载状态可能增加按钮宽度。

加载图标颜色从 Root / Leading 的前景样式读取，需要改色时设置 `UI`；最后执行的 `Style(...leading).TextColor` 无法改变已解析的自绘图标颜色。`Active`、`ActiveColor`、`ActiveVariant` 控制活动状态。

MyGo 将禁用控件绘制为 50% 透明度，本库不再叠加透明度。子组件设置 `Disabled: false` 无法解除父级禁用。Chip 只有装饰作用，业务上有意义的数量或在线状态应补充到宿主的 `AccessibleLabel` 或 `Alt`。

Chip 的四角位置采用物理左右，RTL 不镜像；Inset 将角标放在主体边界内，内容回调为 nil 时独立显示且不应用角点偏移。文字宽度根据 Indicator.UI 的字体、字号、字重和水平内边距计算。Style 在测量之后执行，通过它改变文字尺寸时应同时设置宽高；自定义大字号通常还需增大角标高度。

**链接：** Link 的 URL 非空时，在激活后通过 `Services.OpenURL` 打开地址。URL 为空时，可用返回元素的 `.OnClick` 实现应用内导航；同时设置 URL 和 OnClick 时要注意 MyGo 的动作覆盖规则。

### 图片、标识与资源

Avatar 按 `Bitmap → Icon → Text → Alt 首字母` 的顺序选择内容。图片要在视图构建之外解码，再传入 `ui.DecodeBitmap` 的结果；失败时传 nil，组件会继续使用后面的内容。下载、缓存、取消和错误处理由应用负责。

Body 部件对应位图，默认填满根容器并按 Cover 居中裁剪；通过 Style 改变根尺寸后，图片仍填满可用区域。XS–XL 对应 24、28、32、36、40 DIP，其他尺寸可通过同步 Style 调整，未增加上游 3xs / 2xs / 2xl / 3xl 的共享枚举。需要头像角标时用 `Chip(..., ChipOptions{Inset: true}, func() { Avatar(...) })` 组合，颜色、位置和隐藏状态由 Chip 控制。

`Text` 原样显示，例如 `+12`。没有 Text 时，Avatar 从 Alt 的前两个空格分词中各取首字符，保留大小写；没有 Alt 时留空。文字会缩小以适应圆形，但未实现完整的 Unicode 字素缩写算法。User 的 Name 默认作为头像 Alt。

`Initials` 是独立的原生辅助函数：忽略连续空白、取首末单词并转为大写，单个中文姓名取前两个字符；Avatar 没有调用它，因此两者的空值、多词姓名和中文结果可能不同。需要这一规则时，显式传 `Text: nui.Initials(name)`。

AvatarGroup 和 ProgressGroup 的每项都需要唯一、非空 ID。AvatarGroup 未设置 Max 或 Max≤0 时显示全部头像；正数限制可见数量，并用同尺寸 `+n` 头像表示剩余人数。组的 Size / Color 只给未显式配置的子头像提供默认值，`Overlap: nui.Set(0)` 可关闭重叠。

AvatarGroup 的 Item 部件统一配置普通头像和溢出头像，应用于各头像自身样式之后。Item.BorderWidth / BorderColor 控制头像外侧分隔环，显式 BorderWidth=0 可移除；重排顺序按稳定的 AvatarItem.ID 保留身份。

Lucide 图标使用独立函数，例如 `lucide.House()`、`lucide.GitBranch()`、`lucide.FileJson2()`。本库收录固定提交中的 2,015 个 SVG，每个图标独立嵌入和缓存，未引用的图标可由 Go 链接器移除。返回的 SVG 是共享只读对象，颜色和尺寸通过组件设置。

完整名称、固定提交、导入方式和离线生成步骤见[图标文档](../icons/lucide/README.md)及 [manifest.json](../icons/lucide/manifest.json)。图标升级和生成属于库维护工作，普通应用直接调用图标函数即可。

### 进度、动画与边界

**Skeleton** 的 Width / Height / Animate 是原生扩展。默认尺寸为 160×16 DIP，默认高度遵循文字缩放，显式尺寸不再缩放；显式 0 保持为 0。需要适应父级宽度时，在返回元素上调用 FillWidth。Root.Radius 与最后执行的 Style 均可设置圆角，最终圆角用于裁剪占位背景及附加内容。

脉冲只改变自绘背景的透明度，颜色通过 UI.Root.Background 设置。Style.Background 配置的是底层表面，不能替换已解析的脉冲颜色。Skeleton 提供本地化加载名称和原生 status 角色，没有 Web 的 aria-busy / alert 完整通知语义；应用负责何时显示、隐藏占位和呈现加载结果。

**Progress** 的 Max 默认 100。显式设置非正数、NaN 或 Inf 会 panic；进度值的 NaN/Inf 按 0 处理，再限制在 0 到 Max 之间。

确定进度按四舍五入后的整数百分比填充，辅助技术数值保留实际完成量。不确定进度指示器长为轨道的一半，默认以两秒周期移动；减少动态效果或 Animate=false 时居中静止。底层使用 `Now=-1 < Min=0` 表示未知进度，并提供 In progress 的值或描述。Windows UIA 会将未知数值限制到下界，具体读屏呈现尚未验证。

水平方向从文字起始侧填充，纵向默认高 160 DIP、从顶部填充；`Inverted` 反转填充方向。`Orientation` 和 `Vertical` 均可控制方向。Neutral 使用 InvertedBackground；Indicator 的 Background 可覆盖自绘填充色，其他 Indicator 字段不生效。Root 对应轨道，其样式不改变自绘指示器的固定胶囊形状。

Label 是可访问名称；返回元素的 Value 可提供工作单位等值说明。需要可见百分比或步骤文字时，应用可组合 Text；当前未提供上游 status 插槽、Max 步骤数组和其余三种 CSS 动画预设。Progress 只展示传入数值，不修改业务状态，也没有值变化回调。

**ProgressGroup** 将多段进度放在同一轨道上。每段先限制在 0 到 Max，非有限值按 0 处理；总和不足时保留空余，超出时按比例缩短，辅助技术总值也限制在 Max 内。

说明列表显示各段相对于 Max 的四舍五入百分比。只要某项有 Label、Icon 或 Build 就创建列表；Icon 未设置时显示色点。Build 同步构建该项目的整行图例，替换默认图标、标签和百分比，内部控件按稳定 ID 保留身份；HideLegend 会卸载图例内容。

每项 UI 可配置 Segment、Item、Leading、Label、Trailing，覆盖组级同名 UI；组级 List 配置列表，Indicator 配置整个轨道。Segment.Background 可指定任意颜色。Neutral 分段和色点使用 InvertedBackground，Neutral 图标使用 Highlighted。XS–XL 图例字号为 12/14/14/14/16，图标为 12/16/16/16/20，色点为 6/8/8/8/10 DIP，遵循文字缩放。

水平方向随 RTL 调整；Orientation 默认 Horizontal，纵向 Height≤0 使用 160 DIP。分段静态显示，提供整体可访问进度；没有逐分段进度角色、内置状态文字或分段点击事件。应用可通过图例 Build 添加操作或说明，任务调度仍由应用处理。

**动画何时停止：** 持续动画只在可见组件实际绘制时请求下一帧。移出滚动可见区、卸载、关闭 Animate 或启用减少动态效果后停止。Button 的背景过渡由 `Motion.Transition/Reduce` 控制，`Animate: nui.Set(false)` 控制加载图标；过渡结束后不再持续请求帧。组件没有用于动画的定时 goroutine。

## 使用边界

样式通过 Go 选项和部件 API 设置，不支持 CSS 选择器或 Vue 插槽。平台状态、输入法和辅助技术的验证范围见[兼容性说明](compatibility.md)。
