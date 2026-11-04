# 高级交互、图表与布局

本页介绍命令搜索、应用侧栏、底部面板、文件选择、图表和可拖动布局。组件状态按窗口保存，初始化数据和启动后台任务的位置在各节单独说明。

| 想实现什么 | 阅读位置 |
| --- | --- |
| 用快捷键打开命令搜索 | [CommandPalette](#commandpalette) |
| 应用导航、可调宽侧栏、窄窗口导航 | [NavigationMenu 与 Sidebar](#navigationmenu-与-sidebar) |
| 可拖动的底部面板、文件选择和上传状态 | [Drawer](#drawer)、[FileUpload](#fileupload-与资源所有权) |
| 展示图表或轮播内容 | [Chart](#chart)、[Carousel 与 Marquee](#carousel-与-marquee) |
| 分割面板、页面头尾和页脚链接 | [布局组件](#splitterheaderfooterfootercolumns) |

## 入口与定制

通常先在窗口模型中声明状态，再在视图中传入该状态。Options 中的 Decoration 用于设置 UI 和同步 Style，优先级见[主题文档](foundations.md#样式覆盖)。返回元素可立即设置布局，Element 和 Context 不保留到之后的构建或后台任务。

各组件的状态类型和可用部件集中在末尾的[状态与部件索引](#状态与部件索引)。平台范围见[兼容性说明](compatibility.md)。

## CommandPalette

### 创建命令并显示面板

CommandPalette 提供可搜索的命令列表，默认用 Ctrl+K 打开或关闭。应用先创建 CommandIndex，再交给长期保存的 CommandPaletteState：

```go
// 初始化或命令数据变化时执行，不要每次构建都创建索引。
state.Commands.SetIndex(nui.NewCommandIndex([]nui.CommandAction{
    {
        ID: "save",
        Label: "保存工作区",
        Group: "工作区",
        Key: ui.KeyS,
        Mods: ui.Ctrl,
        Shortcut: "Ctrl S",
        OnSelect: save,
    },
}))
```

视图中始终调用组件，包括关闭时：

```go
nui.CommandPalette(c, theme, &state.Commands, nui.CommandPaletteOptions{
    Title: "工作区命令",
})
```

窗口关闭时通过 `window.OnClosed(state.Commands.Close)` 清理任务。若组件不再出现在视图中，也要显式调用 Close。

命令全树的 ID 必须非空且唯一，最多 32 层。NewCommandIndex 复制搜索数据，按父动作组织层级；数据变化时用 SetIndex 替换。

### 快捷键与搜索

| 选项或行为 | 说明 |
| --- | --- |
| ShortcutKey / ShortcutMods | 修改打开面板的快捷键；DisableShortcut 可关闭 |
| 动作的 Key / Mods | 注册可执行快捷键；Shortcut 只用于显示 |
| Label / Keywords | 搜索名称和别名；不区分大小写，按 Unicode 字符子序列匹配 |
| 分组与子动作 | 支持描述、图标、禁用、子级和返回上一层 |
| 结果数量 | 默认最多 200 条，可配置到 1,000 条，列表使用虚拟化 |

调用组件时所在的父级作用域决定快捷键范围，业务快捷键冲突由应用处理。搜索会优先考虑连续匹配和词首匹配，使用 rune 索引高亮，不做拼音、分词或重音折叠。

输入框按 Down 进入列表，再用方向键、Home、End 移动，Enter 激活。搜索未完成或当前查询与结果不一致时，Enter 不会执行旧动作。描述显示在标签下一行，字号相同、颜色为 Muted。

### 后台搜索和关闭

默认搜索在后台扫描不可变索引，最多保留一份待交付结果，在 UI 构建时读取。每处理 128 条检查取消，只保留评分最高的有限结果，无需对全量匹配项排序。过期任务不能回写。

Open=false 后的下一次构建会取消任务；Close 立即取消。组件没有常驻 goroutine 或定时器，状态只能用于一个窗口。

需要自己的搜索服务时，提供 `Search(ctx, ticket, query, parent)`。应用启动任务，并在 `Window.Update` 中调用 `state.Results.Complete(ticket, matches, err)`；同时负责结果 ID、父级关系和触发重绘。

查询框受[输入法取消问题](compatibility.md#输入法取消)影响，完整读屏体验尚未验证。

## NavigationMenu 与 Sidebar

### 导航与布局分别负责什么

NavigationMenu 显示导航项，Sidebar 提供头部、可滚动正文、底部和收起按钮。应用用 active ID 保存当前位置，并在 OnNavigate(id) 中执行实际页面切换；组件不解析 URL。

NavigationItem 包含 ID、Label、Description、Group、Icon、Disabled、Children。ID 在全树唯一，删除项目后会清理相关菜单状态。父项用于展开子级，不同时导航。

| 模式 | 子级显示方式 |
| --- | --- |
| 水平导航 | 使用 DropdownMenu 展示多层菜单，支持 Escape 和焦点恢复 |
| 纵向导航 | 在原位置展开层级，支持方向键、Home、End 和 Tab |
| 纵向紧凑导航 | 有图标的项目隐藏文字并居中，保留辅助技术名称；无图标项目仍显示标签 |

选中子项时祖先也会高亮，禁用项不能激活。紧凑模式默认有 Tooltip 和侧向 Popover，可分别通过 `NavigationOptions.Tooltip/Popover = nui.Set(false)` 关闭。

侧向 Popover 支持悬停、跨锚点间隙、点击或朝向面板的方向键打开；键盘打开时聚焦第一个可用子项。Escape、反向键或选择项目会关闭并恢复触发项焦点。内部子级可继续展开，PopupPlacement 可修改默认位置，适合右侧栏。

收起导航时，隐藏子项上的焦点移回父项，展开状态和活动 ID 保留。

### Sidebar 状态与尺寸

Sidebar 默认展开宽 256 DIP，收起宽 64 DIP。`collapsed *bool` 控制桌面收起状态，Options.State 可传入一份 SidebarState：

| 状态 | 用途 |
| --- | --- |
| Width（Optional[float32]） | 记录展开宽度；未设置时取 Options.Width，用户调宽后更新 |
| Open（bool） | 窄窗口弹层是否打开，与 collapsed 独立 |
| 不传 State | 使用当前稳定 Key 的局部状态；跨卸载保留或写入文件时应显式传 State |

收起不会改写记住的展开宽度。双击手柄恢复 Options.Width，应用也可把 State.Width 改为未设置。持久化由应用执行，组件不会在构建中读写磁盘。

`Resizable: true` 开启调宽，默认最小 / 最大宽度 160 / 480 DIP，可用 MinWidth / MaxWidth 修改。拖到 CollapsedWidth + 4 DIP 以下收起，向外拖动重新展开。

| 手柄操作 | 行为 |
| --- | --- |
| Left / Right | 按屏幕方向移动边缘 10 DIP |
| Shift+Left / Right | 移动 50 DIP |
| Home / End | 收起或最小宽度 / 最大宽度 |
| Enter / 双击 | 切换收起 / 恢复默认宽度 |

左右侧栏和 RTL 都支持这些操作。尺寸必须有限且非负，MaxWidth≥MinWidth；启用调宽时，收起宽度不能大于最小展开宽度。

### 开关、位置和回调

| 选项 | 默认行为 |
| --- | --- |
| Collapsible | true；false 禁止用户收起 |
| Toggle | true；false 隐藏内置按钮，应用可在别处提供控制 |
| Disabled | 禁用内置交互，也继承祖先禁用状态 |
| Side | 默认 `nui.Set(nui.SideLeft)`，表示阅读方向起始侧；SideRight 表示结束侧 |

RTL 下边框、按钮图标、手柄和 Slideover 一起调整。应用仍要把右侧栏放到 Row 的主内容之后。宽度使用 DIP，不使用网页的百分比、rem 或 DashboardGroup 单位。

OnResize、OnCollapsedChange、OnOpenChange 只在用户操作或组件自动关闭导致值变化时通知，应用直接写状态不回调。OnResize 报告展开宽度，收起变化由 OnCollapsedChange 单独报告。

### 内容区域和窄窗口

HeaderContent、FooterContent、ToggleContent 回调接收 `SidebarScope`，可读取 Collapsed、Open、Mobile，并调用 Collapse(bool) / SetOpen(bool)。HeaderContent / FooterContent 优先于旧的 Header / Footer 回调。

头尾内容需要自行适应紧凑宽度；完全自定义头部时可隐藏内置 Toggle。这些回调同步构建内容，不能保存 Element 或 Context。

设置 `Breakpoint: nui.Set(float32(900))` 后，窗口宽度小于 900 DIP 时隐藏桌面侧栏，改为显示打开按钮。未设置或为 0 时始终采用桌面布局。

| 窄窗口 Mode | 展示方式 |
| --- | --- |
| SidebarSlideover（默认） | 从侧面打开 |
| SidebarModal | 全屏模态面板 |
| SidebarDrawer | 底部面板，高度为窗口的 85% |

弹层内容按展开模式构建，沿用对应组件的 Escape、遮罩关闭、滚动和焦点限制。回到桌面宽度时立即移除弹层焦点约束；普通关闭保留退出动画并遵守减少动态效果设置。

NavigationKey 传入当前页面 ID，变化时默认关闭窄窗口面板；AutoClose=false 可关闭这个行为。外部打开按钮可通过 ReturnFocus 指定关闭后的焦点目标。Sidebar 应持续调用，包括关闭期间。

```go
// state 保存 Collapsed bool、Sidebar nui.SidebarState、
// Active string、Navigation nui.NavigationState。
nui.Sidebar(c.Key("workspace-sidebar"), theme, &state.Collapsed,
    nui.SidebarOptions{
        Title: "工作区",
        State: &state.Sidebar,
        Resizable: true,
        Breakpoint: nui.Set(float32(900)),
        NavigationKey: state.Active,
        FooterContent: func(c *ui.Context, t nui.Theme, scope nui.SidebarScope) {
            label := "账户"
            if scope.Collapsed {
                label = ""
            }
            nui.Button(c, t, label, nui.ButtonOptions{
                Leading: lucide.User(),
                AccessibleLabel: "账户",
                Variant: nui.Ghost,
                Color: nui.Neutral,
            })
        },
    }, func(c *ui.Context, t nui.Theme, compact bool) {
        nui.NavigationMenu(c.Key("navigation"), t, &state.Active,
            &state.Navigation, items,
            nui.NavigationOptions{Vertical: true, Compact: compact})
    })
```

跨断点会卸载并重建控件树。应用模型中的草稿和导航状态可保留，控件局部滚动状态不会自动迁移。当前没有 Web mega menu、CSS 断点、NuxtLink 兼容层或 DashboardGroup 多面板协调。导航使用原生 Group 角色，MyGo 尚无公开的 Navigation landmark 枚举。

## Drawer

### 控制打开状态和高度

Drawer 是带独立拖动手柄的底部模态面板，复用 Slideover 的正文滚动、退出动画和焦点恢复。应用保存 DrawerState，通过 Open 控制打开，通过 Snap 控制停靠位置。

默认停靠高度为窗口的 50% 和 85%。SnapPoints 必须严格递增，每项在 `(0, 1]` 范围内。需要从其他边缘打开时，使用 Slideover。

### 拖动和关闭

只有手柄能拖动，正文区域继续用于滚动。向下拖动降低面板，松开后根据当前高度和移动速度选择停靠点。以下情况可关闭：高度低于最小停靠点的一半，或移动超过默认 80 DIP 且速度超过 900 DIP/s。DismissDistance / DismissVelocity 可调整阈值。

Dismissible=false 禁止通过拖动、遮罩和 Escape 关闭，显式关闭按钮仍可使用。手柄可聚焦，Up/Down 切换停靠点，Home/End 到首尾；失去指针捕获或窗口焦点时取消拖动，恢复原停靠位置。

手柄默认 48×6 DIP，Accented 色、圆角。停靠高度立即切换，打开和关闭沿用 Slideover 动画。应在每次构建中调用 Drawer，包括关闭时。当前不支持四边拖拽、弹簧吸附或正文滚动联动的触屏手势。

## FileUpload 与资源所有权

### 组件负责选择和状态，应用负责传输

FileUpload 提供文件选择、规则检查、列表和进度展示。网络上传由应用实现；选中文件不会自动开始传输。

FileUploadState.Files 保存文件元数据，不保存打开的文件句柄。UploadFile 包含 Path、Name、MIME、Size、状态、0–1 的 Progress 和 Error。Add 为每次加入分配稳定 ID，忽略重复 Path。

| 配置 | 用途 |
| --- | --- |
| UploadRules.Accept | 扩展名、MIME 或类别，例如 `.png`、`text/plain`、`image/*` |
| MaxBytes / MaxFiles | 单文件大小、已接受文件数量上限 |
| Dialog | 原生文件选择器的父窗口、多选和过滤器 |
| Choose(ctx) | 替换默认文件来源 |
| Variant / Layout / Preview 等 | 选择区域、列表或网格、预览开关，见[参数说明](options.md#侧栏导航和文件选择) |

超限或读取失败的文件以 Rejected 显示原因，可由用户移除。组件继承 FormField.Disabled / Size，默认颜色、尺寸和选择按钮外观取 ComponentFileUpload。

### 文件检查过程

默认选择操作在后台调用 `mygo.Dialog.Open`，再用 `os.Stat` 检查普通文件。外部 DroppedFiles 使用相同的检查和规则；InspectUploadFiles、PickUploadFiles 也可单独调用。

MIME 只根据扩展名推导，不读取文件正文。Accept 用于筛选选择结果，业务需要可信的文件类型时，还应自行验证内容。

### 启动、取消和重试上传

```go
// 创建窗口后设置父窗口，并注册清理。
options.Dialog = mygo.OpenDialogOptions{Parent: window, Multiple: true}
window.OnClosed(state.Files.Dispose)

// 在应用的上传动作中开始任务；upload 由应用实现。
ctx, ticket := state.Files.Begin(fileID)
go func() {
    err := upload(ctx, path)
    window.Update(func() {
        state.Files.Report(ticket, 1, true, err)
    })
}()
```

Begin 开始任务，应用在 Window.Update 中通过 Report 发布进度、完成或错误。后台不能直接修改 Files。界面会展示 Ready、Uploading、Complete、Failed、Canceled、Rejected；失败后可再次 Begin，库没有额外的上传重试协议。

Cancel(id) 取消任务 context，Remove(id) 同时取消并移除。取消、删除、重试前或已经完成的旧票据不能继续回写。OnCancel / OnRemove 可执行额外清理；应用打开的文件流、临时文件和网络连接也由应用关闭。

### 隐藏和关闭时清理

隐藏、卸载或关窗时调用 Dispose，取消文件检查和上传任务。字段被禁用时也会取消未完成的文件检查，并丢弃晚到结果。

context 取消无法强制关闭已打开的原生文件对话框，但返回结果会被丢弃。原生选择、多选、取消、拖放和过滤器的平台验证见[兼容性说明](compatibility.md#环境与平台)。macOS 沙盒权限和 bookmark 由应用管理。

## Chart

### 准备数据，再绘制图表

Chart 使用 MyGo 的 `ui/chart` 绘制，提供主题、交互图例和动画。ChartState 应由一个窗口中的一个图表实例持有，不复制已渲染状态，也不与后台共享可变数据。

数据初始化或变化时调用设置方法，检查错误：

```go
err := state.Activity.SetCategories([]string{"一月", "二月", "三月"}, []nui.ChartSeries{
    {ID: "build", Label: "构建次数", Values: []float64{30, 45, 38}},
    {ID: "review", Label: "审查次数", Values: []float64{12, 19, 16}},
})
if err != nil {
    return err
}
```

上述片段放在返回 error 的数据初始化函数中。后台结果通过 Window.Update 写入窗口状态。在视图函数中创建图表：

```go
nui.Chart(c, theme, &state.Activity, nui.ChartOptions{
    Label: "工作记录",
    Kind: nui.ChartArea,
    Curve: nui.ChartCurveMonotone,
    Stack: nui.ChartStackStandard,
    Height: 240,
    AnimationDuration: 400 * time.Millisecond,
})
```

### 图型和数据接口

| 需求 | API |
| --- | --- |
| 折线、柱状、面积、散点、饼图、环形图 | ChartLine、ChartBar、ChartArea、ChartScatter、ChartPie、ChartDonut |
| 数值 / 分类 / 时间 X 轴 | SetData、SetCategories、SetTimeData |
| 饼图和环形图 | SetSlices([]ChartSlice) |
| 曲线 | ChartCurveLinear、ChartCurveMonotone、ChartCurveStepAfter、ChartCurveStepBefore、ChartCurveStep |
| 堆叠 | ChartStackNone、ChartStackStandard、ChartStackNormalized、ChartStackCenter，用于柱状图和面积图 |

所有设置方法先验证再复制，错误不会改动已有数据。数据需要满足以下条件：

| 方法 | 数据要求 |
| --- | --- |
| SetData(xs, series) | X 有限且严格递增，每个系列与 X 等长，系列 ID 唯一且非空；NaN/Inf 的 Y 表示缺值 |
| SetCategories(labels, series) | 分类标签唯一且非空，使用 Band 分类轴 |
| SetTimeData(times, series) | 时间严格递增，刻度使用第一项的时区 |
| SetSlices(slices) | 扇区 ID 唯一非空，值有限且非负，总和有限；0 值保留图例但不绘制扇区 |

散点图也使用共享且有序的 X 数据。ChartSlice 包含 ID、Label、Value、Color；同一个状态可分别保存直角坐标数据和扇区数据，切换 Kind 时选择对应数据。

### 坐标、颜色和堆叠

Y 轴包含 0。Standard 将正值向上、负值向下分别累加；Normalized 使用每个 X 的可见值绝对值总和作为分母，纯正值合计为 100%，混合正负值共用该分母；Center 将代数和居中到 0。

默认系列颜色依次为 Primary、blue、violet、amber、rose、cyan，之后循环。浅色取 500 色阶，深色取 400，也可显式配置 Color。语义颜色会随主色、中性色和明暗主题更新。

FormatX、FormatY、FormatTime 设置格式；Height、Grid、BarPadding、DonutInnerRadius、SlicePadding、LineWidth、PointSize 控制尺寸和几何外观。

### 图例和数据选择

图例默认可点击，也能聚焦后用 Space / Enter 切换系列。OnLegendChange(id, visible) 只报告用户操作；应用可通过 SeriesVisible / SetSeriesVisible、SliceVisible / SetSliceVisible 查询和修改。

InteractiveLegend=false 只显示图例，Legend=false 完全隐藏。隐藏系列后会重算坐标和堆叠，隐藏扇区后重算可见比例。显隐状态按 ID 跨更新、重排保留，删除 ID 后清理；全部隐藏时仍保留图例供用户恢复。

直角坐标图支持鼠标选择最近 X，以及 Left、Right、Home、End、Escape，底部和 Tooltip 显示原始数值。饼图和环形图支持点击扇区、方向键和相同的首尾/退出操作，显示相对可见总和的百分比；环形中心不接受扇区选择。密集数据选择始终查询原始值。

### 动画

Animate 默认开启，AnimationDuration 默认 350ms，遵守主题和系统的减少动态效果设置。首次显示直接呈现，后续数据或参数变化才使用过渡。

同图型、同系列、相同缺值结构时做坐标插值，Pie 与 Donut 之间还会过渡内半径。图型、系列、点数、显隐或缺值结构变化时使用淡入淡出。中途更新从当前画面继续，最多保留 3 层退出画面和 1 层目标画面；轴标签和提示使用目标数据。

动画结束、不可见或卸载后停止请求帧，没有后台定时器。`Animate: nui.Set(false)` 立即显示目标值。

### 大量数据与限制

Prepare(pixelWidth) 可提前准备默认折线缓存，其余配置由 Chart 准备。采样宽度上限 2,048，只在数据、显隐、尺寸或几何选项变化后重算，绘制和插值处理采样后的图形。

未堆叠数据使用 MinMax 采样；缺值片段过多时改用固定桶，含缺值的桶保持断开。密集堆叠共用桶端点和整体包络极值，以保持边界 X 对齐，单系列在一个像素内的细节可能省略。密集柱状图不自动求和或平均，需要聚合时先处理业务数据。

当前没有双 Y 轴、对数轴、刷选缩放或触屏手势。需要低层自定义标记和布局时可直接组合 MyGo 的 ui/chart。Gallery 的 **Activity & attachments** 展示各图型、轴、曲线、堆叠、数据更新、缺值、负值和图例。

## Carousel 与 Marquee

### Carousel：按页切换内容

CarouselItem 包含稳定 ID、Label 和同步 Build，CarouselState.Index / Paused 可由应用控制。组件只构建当前及相邻项，相邻项禁用交互，但其 Build 仍执行，因此构建回调中不能无条件启动业务操作。

| 操作或选项 | 行为 |
| --- | --- |
| 拖动 | 沿 Orientation 方向拖动，超过视口 20% 或速度 600 DIP/s 时换页，释放后吸附 |
| Previous / Next、方向键、Home / End | 按页导航，水平用 ←→，纵向用 ↑↓ |
| Arrows / Dots | 控制翻页按钮和圆点导航 |
| Wrap=true | 按钮、键盘和自动播放可从末项回到首项；拖动仍有边界 |

当前位置只写入根元素的辅助技术值，例如第 1 张，共 3 张，使用 slidePosition 语言键；不额外显示页码文字。Index 按位置控制，业务重排后由应用决定新的索引，移除项目时组件会修正越界索引。

默认不自动播放，设置 Autoplay 后才启用。悬停、焦点位于幻灯片、Paused、Active=false、位于窗口可见区外或减少动态效果，都会停止播放。取消拖动会重置位移。

### Marquee：连续滚动展示

Marquee 用于展示性内容，支持方向、Reverse、以 DIP/s 表示的 Speed、Gap、Height 和 Space 暂停。Gap 默认 64 DIP。

内容会重复四份，每份至少与容器等长，内部均匀分布，即使内容短也能滚动。Overlay 默认在两端各三分之一处渐隐到背景色，`Overlay: nui.Set(false)` 关闭渐隐。

ContentKey 变化时从头开始。PauseOnHover 默认 false，启用后悬停暂停；焦点、Paused、Active=false、减少动态效果始终会停止运动。Orientation 或 Vertical 设置方向，水平运动随 RTL 调整。

**内容回调每次构建会执行四次，只放展示元素。** 不要放业务按钮、输入框或副作用。底层没有对整棵子树隐藏辅助技术内容的接口，重复内容的读屏效果需在目标工具中验证。

### 隐藏时停止动画

两者都通过绘制请求后续帧，没有定时 goroutine。默认可见性判断使用已经提交的窗口矩形。应用自定义裁剪、隐藏、最小化或仍保留已隐藏容器时，应同步 Active=false，或停止构建组件。

卸载后不会继续请求帧，也不会累积暂停时间；保留隐藏内容时，应用仍应明确管理 Active。

## Splitter、Header、Footer、FooterColumns

### Splitter：分割两个面板

Splitter 使用 MyGo 的 Split / SplitVertical / SplitPercent / SplitVerticalPercent。绑定 `*float32`，默认单位 DIP；Percent=true 时为 0–100，Vertical=true 时上下分割。

Min / Max 与绑定值使用相同单位。指针或键盘修改后，组件把值限制到范围内并调用 OnChange；应用直接赋值后，在下一次构建中限制范围。

原生面板默认保留主题的 40 DIP 最小尺寸，窗口更小时按底层布局规则收缩。辅助技术分隔条的 Range 使用原生 DIP 范围，不反映自定义百分比约束。Root、Leading、Trailing 可分别设置整体和两个面板的样式，分隔线与手柄使用 Theme.Native。

当前只支持双面板，不含多面板协调、折叠停靠或布局序列化。

### HeaderLayout 和 FooterLayout

HeaderLayout 提供 Leading、标题、说明和 Actions；FooterLayout 提供 Text 和 Actions。两者支持换行、伸缩和字体缩放，显示在应用内容区，不修改系统标题栏。

Go 中 Header、Footer 已作为部件名，因此布局函数使用 HeaderLayout、FooterLayout。

### FooterColumns

FooterColumns 接收有稳定 ID 的列，每列包含标题和简单导航项，按最小列宽自动换列。OnNavigate 处理业务跳转，需要嵌套导航时使用 NavigationMenu。窄窗口下哪些区域优先显示，仍由应用安排。

## 状态与部件索引

| 组件 | 持久状态或数据 | 可用部件 |
| --- | --- | --- |
| CommandPalette | CommandPaletteState、CommandIndex | Root、Editor，以及传给 Listbox 的 Root、Label、Item、EmptySlot |
| NavigationMenu | NavigationState、活动 ID `*string` | Root、Label、Item |
| Drawer | DrawerState | Modal / Slideover 的部件，另有 Thumb；默认标题使用 Label / Description |
| FileUpload | FileUploadState | Root、Body、Item |
| Chart | ChartState | Root、Body、Item（图例）、Label、Description、EmptySlot |
| Carousel / Marquee | CarouselState / MarqueeState | Carousel：Root、Body、Item；Marquee：Root |
| Splitter | `*float32` | Root、Leading、Trailing |
| Sidebar | `*bool` collapsed，可选 SidebarState | Root、Header、Label、Trigger、Body、Footer、SeparatorSlot；弹层另有 Panel、Backdrop、Close |
| HeaderLayout | HeaderOptions | Root、Label、Description、Actions |
| FooterLayout | FooterOptions | Root、Label；Actions 使用内容回调 |
| FooterColumns | `[]FooterColumn` | Root、Label、Item |

## 示例

Gallery 的 **Advanced interactions**、**Activity & attachments**、**Application layout** 展示这些组件，**Props & behavior** 和 **Orientation** 演示配置变化。

Gallery 最左侧目录本身也使用 Sidebar + NavigationMenu，默认宽 256 DIP，可在 200–360 DIP 间拖动，收起为 64 DIP 图标栏。页面切换保留宽度和收起状态，每个窗口分别保存；它与 Application layout 页面里的演示侧栏相互独立。自动化测试可通过 Toggle gallery sidebar 和 Resize gallery sidebar 名称定位按钮与手柄。
