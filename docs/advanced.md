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

命令全树的 ID 必须非空且唯一，最多 32 层。NewCommandIndex 复制子级、键帽、样式映射及搜索数据，按父动作组织层级；返回的 Action 元数据应只读，数据变化时用 SetIndex 替换。图标、位图资源及回调仍由应用管理。

### 快捷键与搜索

| 选项或行为 | 说明 |
| --- | --- |
| ShortcutKey / ShortcutMods | 修改打开面板的快捷键；DisableShortcut 可关闭 |
| 动作的 Key / Mods | 注册可执行快捷键；Shortcut 只用于显示 |
| Label / Description / Suffix / Keywords | 默认匹配四类字段，FilterFields 可指定；不区分大小写，按 Unicode 字符子序列匹配 |
| 分组与子动作 | 同组结果连续显示；PreserveGroupOrder 保留索引组顺序。子级可设置 Placeholder，空查询 Backspace 返回 |
| SearchDelay / IgnoreFilter | 默认 0 / false；延迟查询编辑但立即取消旧请求。IgnoreFilter 只跳过内置匹配 |
| Kbds / Prefix / Suffix / Avatar / Chip / Loading / Active | 命令内容及外观；Kbds 只显示快捷键，Loading 不自动禁用 |
| 结果数量 | 默认最多 200 条，可配置到 1,000 条，列表使用虚拟化 |

调用组件时所在的父级作用域决定快捷键范围，业务快捷键冲突由应用处理。搜索会优先考虑连续匹配和词首匹配，使用 rune 索引高亮，不做拼音、分词或重音折叠。

输入框按 Down 进入列表，再用方向键、Home、End 移动，Enter 激活。搜索未完成或当前查询与结果不一致时，Enter 不会执行旧动作。描述显示在标签下一行，字号相同、颜色为 Muted。SearchInput=false 隐藏输入但保留查询；Size、SearchInputOptions、BackOptions 配置尺寸、输入与返回按钮。

叶命令执行前调用 OnSelectEvent，单项处理器优先，PreventDefault 可取消执行。执行后调用旧 Action.OnSelect 和根 OnSelect，默认先关闭面板；CloseOnSelect=false 允许保持打开并重复执行。同一点击的回调改变查询、索引、层级或请求后，不继续执行旧命令。Disabled 同时限制输入、结果和注册的快捷键。

### 后台搜索和关闭

默认搜索在后台扫描不可变索引，最多保留一份待交付结果，在 UI 构建时读取。每处理 128 条检查取消，只保留评分最高的有限结果，无需对全量匹配项排序。过期任务不能回写。

Open=false 后的下一次构建会取消任务；Close 立即取消。组件没有常驻 goroutine 或定时器，状态只能用于一个窗口。

需要自己的搜索服务时，提供 `Search(ctx, ticket, query, parent)`。应用启动任务，并在 `Window.Update` 中调用 `state.Results.Complete(ticket, matches, err)`；同时负责结果 ID、父级关系和触发重绘。FilterFields / IgnoreFilter 仅控制内置搜索。替换同类 Search 闭包后调用 Refresh；Searching() 同时包含延时等待与请求执行。

本组件是原生模态命令启动器，不提供 Vue 的嵌入式多选模型、路由字段或 Fuse 专有配置。

查询框受[输入法取消问题](compatibility.md#输入法取消)影响，完整读屏体验尚未验证。

## NavigationMenu 与 Sidebar

### 导航与布局分别负责什么

NavigationMenu 显示导航项，Sidebar 提供标题、描述、标题操作、可滚动正文、底部和收起按钮。应用用 active ID 保存当前位置，并在 OnNavigate(id) 中执行实际页面切换；组件不解析 URL。

NavigationItem 包含稳定 ID、文字、图标/头像、徽章、状态圆点、单项样式与 Children。ID 在全树唯一，删除项目后清理菜单状态并恢复焦点。父项用于展开子级，不同时导航；OnSelect 可用 PreventDefault 取消叶项导航，成功后 OnNavigate 通知一次。

| 模式 | 子级显示方式 |
| --- | --- |
| 水平导航 | 使用 DropdownMenu 展示多层菜单，支持 Escape 和焦点恢复 |
| 纵向导航 | 在原位置展开层级，支持方向键、Home、End 和 Tab |
| 纵向紧凑导航 | 有图标或头像时隐藏文字并居中，保留辅助技术名称；无配件项目仍显示标签，宽度由宿主控制 |

默认 Color=Primary、Variant=Soft（pill），LinkVariant 为 link 样式；Highlight 显示活动指示线，HighlightColor 未设置时继承 Color。活动后代使祖先强调；Disabled 合并根、单项和继承禁用。横向忽略 Compact，方向键在首尾停止；纵向方向键循环。

紧凑模式的 Tooltip 和侧向 Popover 保留原生默认 true，可分别通过 `NavigationOptions.Tooltip/Popover = nui.Set(false)` 关闭。单项开关可覆盖全局，TooltipOptions / PopoverOptions 可配置具体外观和行为；横向提示由单项 Tooltip 启用。

侧向 Popover 支持悬停、跨锚点间隙、点击或朝向面板的方向键打开；键盘打开时聚焦第一个可用子项。Escape、反向键或选择项目会关闭并恢复触发项焦点。内部子级可继续展开，PopupPlacement 可修改默认位置，适合右侧栏。

NavigationState.Expanded 公开纵向展开状态，DefaultOpen 只初始化新节点。Single=true 限制同级分支只展开一项，Collapsible=false 在 Single 模式下禁止收起当前分支。用户变化触发 OnExpandedChange，初始化与应用改值静默。收起导航时，隐藏子项的焦点移回父项；删除根项恢复可用根项，Compact 切换保留展开和活动 ID。

横向使用原生菜单，没有网页的共享 mega menu 视口、路由活动匹配或根悬停延时模型。

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

HeaderContent、FooterContent、ToggleContent 回调接收 `SidebarScope`，可读取 Collapsed、Open、Mobile，并调用 Collapse(bool) / SetOpen(bool)。HeaderContent / FooterContent 优先于旧的 Header / Footer 回调。HeaderContent 替换默认标题和描述；Actions 在展开状态或窄窗口菜单中构建独立操作区。

标题、描述、标题操作和 Toggle 都缺省时，不保留空标题区。正文和底部默认垂直内边距为 16 DIP，实际头尾与正文之间有分隔线。头尾内容需要自行适应紧凑宽度；完全自定义头部时可隐藏内置 Toggle。这些回调同步构建内容，不能保存 Element 或 Context。

设置 `Breakpoint: nui.Set(float32(900))` 后，窗口宽度小于 900 DIP 时隐藏桌面侧栏，改为显示打开按钮。未设置或为 0 时始终采用桌面布局。

| 窄窗口 Mode | 展示方式 |
| --- | --- |
| SidebarSlideover（默认） | 从侧面打开 |
| SidebarModal | 全屏模态面板 |
| SidebarDrawer | 底部面板，高度为窗口的 85% |

弹层内容按展开模式构建，沿用对应组件的 Escape、遮罩关闭、滚动和焦点限制。回到桌面宽度时立即移除弹层焦点约束；普通关闭保留退出动画并遵守减少动态效果设置。

NavigationKey 传入当前页面 ID，变化时默认关闭窄窗口面板；AutoClose=false 可关闭这个行为。禁用状态会阻止内置动作和通知，不会替应用改写受控模型。外部打开按钮可通过 ReturnFocus 指定关闭后的焦点目标。Sidebar 应持续调用，包括关闭期间。零宽或切到关闭的窄窗口模式后，如果原焦点内容被移除，优先返回 ReturnFocus，否则使用仍显示的内置按钮或调宽手柄；完全隐藏内部入口时应提供外部按钮及 ReturnFocus。

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

Drawer 是带独立拖动手柄的底部模态面板，复用 Slideover 的退出动画和焦点恢复。标题、正文与页脚在 ContainerSlot 内整体滚动，最低停靠高度下仍可滚动到关闭操作。应用保存 DrawerState，通过 Open 控制打开，通过 Snap 控制停靠位置。

默认停靠高度为可用窗口的 50% 和 85%。SnapPoints 必须严格递增，每项在 `(0, 1]` 范围内，不接受 px 字符串。Inset=true 先扣除 16 DIP 外边距，再计算高度；需要从其他边缘打开时，使用 Slideover。共享的 Width / MaxHeight / Fullscreen 在 Drawer 中不生效。

### 拖动和关闭

只有手柄能拖动，正文区域继续用于滚动。向下拖动降低面板，松开后根据当前高度和移动速度选择停靠点。以下情况可关闭：高度低于最小停靠点的一半，或移动超过默认 80 DIP 且速度超过 900 DIP/s。DismissDistance / DismissVelocity 可调整阈值，0 沿用默认，参数必须有限且非负。这些是原生阈值，不等同于 Vaul 的 closeThreshold。

Dismissible=false 禁止通过拖动、遮罩和 Escape 关闭，并发送 OnClosePrevent；显式关闭按钮仍可使用。CloseButton 默认 false，传入 Close 配置时默认显示，显式 false 优先。Handle 默认 true，独立于 Header；隐藏时取消进行中的拖动。

手柄可聚焦，Up/Down 切换停靠点，Home/End 到首尾；失去指针捕获或窗口焦点时取消拖动，恢复原停靠位置。OnSnap 仅报告实际索引变化，边界重复按键和外部改值静默。OnClose 报告组件关闭操作，应用直接改 Open 或自定义 Done 不通知。

可见手柄默认 48×6 DIP，Accented 色、圆角，由 Indicator 配置；Thumb 为 64×16 DIP 交互区域。标题为空时不创建空标题，面板和手柄有本地化名称。停靠高度立即切换，打开和关闭沿用 Slideover 动画。应在每次构建中调用 Drawer，包括关闭时。当前不支持四边拖拽、弹簧吸附或正文滚动联动的触屏手势。

## FileUpload 与资源所有权

### 组件负责选择和状态，应用负责传输

FileUpload 提供文件选择、规则检查、列表和进度展示。网络上传由应用实现；选中文件不会自动开始传输。

FileUploadState.Files 保存文件元数据，不保存打开的文件句柄。UploadFile 包含 Path、Name、MIME、Size、状态、0–1 的 Progress、Error 和应用提供的 Avatar。Add 分配不与现有条目冲突的稳定 ID，忽略完全相同的 Path；直接赋值 Files 时，应用应提供唯一 ID。组件不自行加载或解码预览图片。

| 配置 | 用途 |
| --- | --- |
| UploadRules.Accept | 扩展名、MIME 或类别，例如 `.png`、`text/plain`、`image/*`；支持 `*`、`*/*`、逗号分隔、复合后缀及忽略大小写，空或全空白允许所有类型 |
| MaxBytes / MaxFiles | 单文件大小、已接受文件数量上限 |
| Dialog | 原生文件选择器的父窗口、多选和过滤器 |
| Choose(ctx) | 替换默认文件来源 |
| Variant / Layout / Position | 默认 UploadArea / UploadGrid / UploadOutside；UploadInside 只将 UploadList 放进选择区域 |
| Icon / ShowIcon / Highlight | 上传入口图标、显隐和边框高亮；Neutral 使用反色，字段错误优先使用错误色 |
| FileIcon / FileImage / UploadFile.Avatar | 文件回退图标与应用提供的预览配件；FileImage=false 只显示 FileIcon |
| FileDelete / FileDeleteOptions / FileDeleteIcon | 删除按钮开关、配置和图标；列表默认 Neutral / LinkVariant / 组件尺寸，网格默认 Neutral / Solid / XS |
| Preview | 默认 true；false 隐藏文件列表，保留数据和任务，见[参数说明](options.md#侧栏导航和文件选择) |

超限或读取失败的文件以 Rejected 显示原因，可由用户移除。MaxBytes / MaxFiles 非正数不限制；拒绝项不计入数量，已接受但失败或取消的项目仍计入。组件继承父容器与 FormField.Disabled，以及 FormField.Size / 标签 / 错误信息，默认颜色、尺寸取 ComponentFileUpload。

OnChange 仅在选择或移除实际改变列表时通知一次，传入列表副本；重复 Path、取消选择、选择失败及应用直接 Add / Remove / Report 不发送通知。OnReject 先于 OnChange 报告本次新增的拒绝项，不清除队列中的拒绝记录。移除或取消传输后，焦点返回文件选择入口。

Go 保留上传队列模型，Dialog.Multiple 只控制原生对话框单次多选，不能用它限制队列数量；Nuxt 单文件 modelValue 替换、multiple 强制 area 和网页 input.reset 没有直接映射。网格保留文件名、状态与进度，采用可换行卡片，不宣称与上游图片网格完全相同。

### 文件检查过程

默认选择操作在后台调用 `mygo.Dialog.Open`，再用 `os.Stat` 检查普通文件。外部 DroppedFiles 使用相同的检查和规则；InspectUploadFiles、PickUploadFiles 也可单独调用。

MIME 只根据扩展名推导，不读取文件正文。Accept 用于检查选择及拖放结果，不自动转成原生 Dialog.Filters；业务需要可信的文件类型时，还应自行验证内容。Dropzone=false 不注册拖放目标；Interactive=false 保留独立选择按钮。Choose 返回空结果表示取消；返回错误时设置 State.Error，文件队列不变。

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

默认系列颜色依次为 Primary、blue、violet、amber、rose、cyan；超过一轮后按 0.65 的轮次比例与背景混色。浅色取 500 色阶，深色取 400。Colors 可提供原生颜色列表，单项显式 Color 优先，其中 Neutral 使用反色。混色采用 MyGo 的原生颜色插值，不声称与 CSS 的 oklab 混色完全相同。

FormatX、FormatY、FormatTime 设置格式；默认数值使用 Go 格式化，不自动采用浏览器 Intl 数字格式。XAxis / YAxis 默认 true，控制各轴刻度与标题；Grid 独立控制横向网格。XLabel 在图下方居中，YLabel 在上方横向显示。RTL 将 X 轴数据和刻度反向、Y 轴放在右侧，并映射物理方向键与指针位置。

Height 默认 300 DIP，正值优先；否则可按正的 AspectRatio 用实际宽度计算。Width 正值固定宽度，否则跟随父容器，首次未测量时用 640。高度用于坐标区，图例、轴标题与状态文字另占空间。所有尺寸必须有限。BarPadding、SlicePadding、LineWidth、PointSize 保留原生几何配置；Points 默认 false，开启后在折线上显示采样数据点，PointSize 不影响独立的焦点标记。

圆环 Thickness 默认 24 DIP；显式 DonutInnerRadius 仍可使用比例并优先于厚度，范围限制为 0–0.95。Thickness 限制为非负，0 使内外边界重合。Center 同步构建绘图区中心内容，不拦截指针；回调不能保存 Element 或 Context。

### 图例和数据选择

未设置 Legend 时，只有系列或扇区总数大于 1 才显示图例，计数包括隐藏与零值项目；显式 true/false 优先。图例默认可点击，也能聚焦后用 Space / Enter 切换系列。OnLegendChange(id, visible) 只报告用户操作；应用可通过 SeriesVisible / SetSeriesVisible、SliceVisible / SetSliceVisible 查询和修改。空 Label 默认使用对应 ID。

InteractiveLegend=false 只显示图例，Legend=false 完全隐藏。隐藏系列后会重算坐标和堆叠，隐藏扇区后重算可见比例。显隐状态按 ID 跨更新、重排保留，删除 ID 后清理；全部隐藏时仍保留图例供用户恢复。

直角坐标图支持鼠标定位最近 X，以及 Left、Right、Home、End、Escape；Home/End 保持数据的逻辑首尾。底部状态文字和 Tooltip 显示原始数值，Tooltip=false 只隐藏提示。TooltipOptions 可配置位置等参数，默认 TopCenter、偏移 12 DIP、无延时。饼图和环形图支持扇区定位和方向键，显示相对可见总和的百分比；环形中心不接受扇区定位。密集数据选择始终查询原始值。

OnFocusChange 在指针或键盘改变焦点行/扇区时通知，边界按键不会重复通知。点击或 Enter/Space 激活发送 OnSelect，重复激活可以再次发送。参数 ChartSelection.Index 是原始 X 行或扇区索引，-1 表示无焦点；SliceID 仅用于径向图。Go 以整行 X 为单位，没有 Web 引擎的独立系列 point、focusGroupChange 对象或 DOM 事件。图表使用可聚焦的分组语义，中心文字和空状态仍可由原生辅助技术读取。

### 动画

Animate 默认开启，AnimationDuration 默认 350ms，遵守主题和系统的减少动态效果设置。首次显示直接呈现，后续数据或参数变化才使用过渡。

同图型、同系列、相同缺值结构时做坐标插值，Pie 与 Donut 之间还会过渡内半径。图型、系列、点数、显隐或缺值结构变化时使用淡入淡出。中途更新从当前画面继续，最多保留 3 层退出画面和 1 层目标画面；轴标签和提示使用目标数据。

动画结束、不可见或卸载后停止请求帧，没有后台定时器。`Animate: nui.Set(false)` 立即显示目标值。

### 大量数据与限制

Prepare(pixelWidth) 可提前准备默认折线缓存，其余配置由 Chart 准备。采样宽度上限 2,048，只在数据、显隐、尺寸或几何选项变化后重算，绘制和插值处理采样后的图形。

未堆叠数据使用 MinMax 采样；缺值片段过多时改用固定桶，含缺值的桶保持断开。密集堆叠共用桶端点和整体包络极值，以保持边界 X 对齐，单系列在一个像素内的细节可能省略。密集柱状图不自动求和或平均，需要聚合时先处理业务数据。

参考 Nuxt UI 4.12.0 的 TanStack Charts，Go 实际使用 MyGo 的 ui/chart，不依赖 Web 图表引擎。保留散点图、饼图、数值与时间轴等原生能力；不提供任意 TanStack definition、双 Y 轴、对数轴、刷选缩放或触屏手势。需要底层自定义标记和布局时可直接组合 ui/chart。Gallery 的 Chart 组件页和 **Activity & attachments** 展示图形、坐标、RTL、事件、中心内容、数据更新及交互图例。

## Carousel 与 Marquee

### Carousel：按页切换内容

CarouselItem 包含稳定 ID、Label 和同步 Build，CarouselState.Index / Paused 可由应用控制。组件只构建当前及相邻项，相邻项禁用交互，但其 Build 仍执行，因此构建回调中不能无条件启动业务操作。

| 操作或选项 | 行为 |
| --- | --- |
| 拖动 | 沿 Orientation 方向拖动，超过视口 20% 或速度 600 DIP/s 时换页，释放后吸附 |
| Previous / Next、方向键、Home / End | 按页导航，水平用 ←→，纵向用 ↑↓ |
| Arrows / Dots | 控制翻页按钮和圆点导航 |
| Wrap=true | 按钮、键盘和自动播放可从末项回到首项；拖动仍有边界 |
| WatchDrag | 默认 true；false 不注册拖拽，其他导航继续可用；切换方向会取消当前拖拽 |
| Prev / Next、PrevIcon / NextIcon | 按钮与图标配置；默认 Outline / Neutral，默认 Arrow 随方向改变；显式图标按原样显示 |
| StopOnInteraction | 默认 true；启用自动播放时，箭头、键盘和完成拖拽会设置 Paused=true，需 Play 恢复；false 保留自动恢复策略，圆点只定位 |

当前位置只写入根元素的辅助技术值，例如第 1 张，共 3 张，使用 slidePosition 语言键；不额外显示页码文字。Index 按位置控制，重排后由应用决定索引，移除项目时限制到有效范围。外部更新不发送 OnChange；当前页中的焦点元素被移除时，焦点回到视口。CarouselItem.Decoration 可覆盖 Item 部件，应用仍负责保存草稿等业务数据。

默认不自动播放，Autoplay 正值指定间隔。悬停、组件内焦点、禁用、Paused、Active=false、不可见、窗口不活动或减少动态效果都会停止调度。Active 仅表示原生自动播放的活动状态，不禁用手动导航，与 Embla 的 active 不同。Wrap=false 时停在末项；临时暂停解除后重新计时。播放/暂停按钮为原生扩展。

OnChange 在用户导航或自动播放实际改变索引时通知一次；构建期间的拖拽和计时结果在构建结束后通知。取消拖拽清除位移，空列表也清理拖拽状态。按钮按实际尺寸居中，边界禁用优先于按钮配置。默认按钮在视口内侧 16 DIP，内容需要留出空间；Gallery 与网站通过 Item 内边距展示输入框布局。

Go 只构建当前页及相邻页，没有 Embla 的自由拖拽、滑动组、媒体查询、物理动画持续时间、自动高度或插件对象。切换为即时定位；Dots 使用原生选中按钮，不声称实现完整 ARIA tablist。

### Marquee：连续滚动展示

Marquee 用于展示性内容，支持 Orientation / Vertical、Reverse、以 DIP/s 表示的 Speed、Gap、Height 和 Space 暂停。默认 Speed 为 32 DIP/s，Gap 为主题缩放后的 64 DIP，Height 为 44 DIP；这些浮点值必须有限，非正值使用默认值。Speed 与上游 CSS 的每周期 20 秒不同，暂停使用 Paused 或 Active。

Repeat 未设置时为 4，显式 `nui.Set(0)` 不构建内容，负数无效。每份内容在运动方向上至少与容器等长，在另一方向居中，内部均匀分布，即使内容短也能滚动；一份内容可能在循环中出现空白。Overlay 默认在两端各三分之一处渐隐到背景色，`Overlay: nui.Set(false)` 关闭渐隐。

ContentKey 变化时从头开始。PauseOnHover 默认 false，启用后悬停暂停；焦点、Paused、Active=false、父级禁用和减少动态效果也会停止运动。水平运动随 RTL 调整。外部偏移会归一到当前内容周期，容器缩小时仍保持有效位置；非有限偏移恢复到 0。

Root 配置容器，Body 配置每份内容。Body.Gap 只改变副本内部项目间距，副本之间仍使用 Gap。主题、实例 UI 和最后 Style 的覆盖顺序保持一致。

**内容回调每次构建按 Repeat 执行，只放展示元素。** 不要放业务按钮、输入框或副作用。底层没有对整棵子树隐藏辅助技术内容的接口，重复内容的读屏效果需在目标工具中验证。网站预览为控制资源使用，允许最多 100 份；Go API 不设置这个上限。

### 隐藏时停止动画

两者都通过绘制请求后续帧，没有定时 goroutine。默认可见性判断使用已经提交的窗口矩形。应用自定义裁剪、隐藏、最小化或仍保留已隐藏容器时，应同步 Active=false，或停止构建组件。

卸载后不会继续请求帧，也不会累积暂停时间；保留隐藏内容时，应用仍应明确管理 Active。

## Splitter、Header、Footer、FooterColumns

### Splitter：分割两个面板

Splitter 基于 MyGo 布局、焦点、拖拽和快捷键构建，绑定 `*float32`。默认单位 DIP；Percent=true 时为 0–100，占两个面板可分配空间的比例，不包括分隔条。Orientation=Vertical 或 Vertical=true 时上下分割，水平布局及方向键随 RTL 镜像。

Min / Max 与绑定值使用相同单位，默认下界 0，上界为 1e7 DIP 或 100%。显式 0 有效。指针或键盘实际改值后调用一次 OnChange；边界重复输入不通知。应用直接赋值后，在下一次构建中静默限制范围，非有限模型值先按 0 处理。约束和步长必须有限，最大值不能低于有效最小值。

KeyboardResizeBy 未设置时每次移动主题缩放后的 10 DIP；显式设置使用绑定值单位，0 关闭方向键调整。Home / End 定位到可达的范围边界。Disabled 只锁定分隔条，面板仍可编辑；禁用父容器才会禁用整个子树。

两个面板保留原生 40 DIP 最小尺寸；DIP 模式第一面板可显式设置得更小。空间不足时布局会收缩，模型值仍保留，因此模型 0 不等同于折叠面板。布局只为分隔线保留 1 DIP，8 DIP 拖拽区域覆盖面板边缘，不额外分开两个面板，嵌套时横纵线连续相接。Root、Leading、Trailing、SeparatorSlot、Indicator 可分别配置容器、两个面板、拖拽区域和中心线。可访问 Range / Step 使用绑定值单位及配置约束，不再固定为底层 DIP。

当前支持双面板及嵌套，不含多面板协调、折叠停靠、自动保存或布局序列化。持久模型由应用管理；布局与方向变化保留面板输入状态。

### HeaderLayout 和 FooterLayout

HeaderLayout 是原生应用标题区，提供 Title、原生扩展 Description、Leading 和 Actions。标题默认空，非空时为一级标题；空标题不生成空的标题节点。最小高度为主题缩放后的 64 DIP，水平/垂直内边距为 24/8 DIP，区域间距为 12 DIP，前导和动作内部间距为 6 DIP。Root、Leading、Body（标题文字容器）、Label、Description、Actions 均有实际样式入口。切换 Leading 时保留动作中的控件身份和输入状态。

HeaderLayout 不自动创建首页链接、网页粘性定位、断点菜单或路由监听。Gallery 和网站示例把动作中的菜单按钮与 Sidebar 组合，通过持久 SidebarState、NavigationKey 和 ReturnFocus 管理打开、导航关闭及焦点。示例使用 2000 DIP 断点，便于在常见预览宽度下操作菜单；这不属于 HeaderOptions 默认值。Header 的 Body 样式部件指标题文字容器，与网页 Header 的菜单 body 插槽不同。

FooterLayout 用 Text 显示说明，用 Actions 构建操作。Root、Label、Actions 分别配置整体、文字和操作容器。默认水平/垂直内边距为 20/12 DIP、区域间距 12 DIP、动作内部间距 6 DIP，无顶部边线。文字和操作可以在可用宽度内换行，动作样式不会误作用于单个按钮。

这两个函数支持文字缩放、RTL 和原生布局组合，不修改系统标题栏，也不提供网页的 DOM 元素类型或全部插槽。多区页脚、上下附加内容和特定断点排列由应用组合布局。Go 中 Header、Footer 已作为部件常量，因此函数名使用 HeaderLayout、FooterLayout。

### FooterColumns

FooterColumns 接收有稳定非空 ID 的列，列内的链接 ID 在该列中唯一。Title 非空时显示三级标题；Items 使用 NavigationItem 的 ID、Label、Icon、Active、Disabled、OnSelect 和 Decoration。其余分组、描述、嵌套子项、头像、徽章、Tooltip、Popover 配置用于 NavigationMenu，不用于本组件。

用户激活可用链接后先调用单项 OnSelect，NavigationEvent.ID 为目标 ID；PreventDefault 可以取消根 OnNavigate。被禁用的链接不执行回调。Active 由应用控制，默认使用 Primary 和中等字重，未活动链接使用 Muted，悬停使用 Text。原生通过链接角色和 ButtonBase 支持指针、Enter / Space，但不提供 URL、自动路由匹配或外链标记。

MinColumnWidth 默认 160 DIP，非正值采用默认，非有限值无效。此值作为弹性布局的初始列宽参与换行；列和链接受可用宽度约束。各列顶部对齐，默认列间距 32 DIP、标题与列表间距 24 DIP、链接间距 16 DIP，长链接单行省略。RTL 会调整排列顺序，重排列时保留稳定 ID。

Root、Body（每列）、Label（列标题）、List、Item（链接）、ItemLeadingSlot、ItemLabelSlot 可分别配置。单项 UI 覆盖组件 UI，组件 Style 后再执行单项 Style；链接的项目样式使用 Item、ItemLeadingSlot、ItemLabelSlot。网站完整数据面板保留其他 NavigationItem 字段的参考信息，同时标明适用范围。

## 状态与部件索引

| 组件 | 持久状态或数据 | 可用部件 |
| --- | --- | --- |
| CommandPalette | CommandPaletteState、CommandIndex | Root（模态）、Editor、BackSlot、Panel、List、Label、Item、Body、ItemLeadingSlot、ItemLabelSlot、ItemLabelPrefixSlot、ItemLabelSuffixSlot、Description、ItemTrailingSlot、Indicator、EmptySlot、LoadingSlot、LoadingIconSlot |
| NavigationMenu | NavigationState、活动 ID `*string` | Root、List、Panel、Body、Label、Item、ItemLeadingSlot、ItemLabelSlot、Description、ItemTrailingSlot、Indicator |
| Drawer | DrawerState | Backdrop、Panel、ContainerSlot、Thumb、Indicator、Header、Label、Description、Body、Footer、Close |
| FileUpload | FileUploadState | Root、Body、Leading、Label、Description、Trigger、List、Item、ItemLeadingSlot、ItemLabelSlot、StatusSlot、ErrorSlot、ItemTrailingSlot |
| Chart | ChartState | Root、Body、Item（图例）、Label、Description、EmptySlot、List、Indicator、XAxisSlot、YAxisSlot、CenterSlot |
| Carousel | CarouselState | Root、Body、Item、Previous、Next、Actions、List、Indicator、Trigger、EmptySlot |
| Marquee | MarqueeState | Root、Body |
| Splitter | `*float32` | Root、Leading、Trailing、SeparatorSlot、Indicator |
| Sidebar | `*bool` collapsed，可选 SidebarState | Root、Header、Label、Description、Actions、Trigger、Body、Footer、SeparatorSlot；弹层另有 Panel、Backdrop、Close |
| HeaderLayout | HeaderOptions | Root、Leading、Body、Label、Description、Actions |
| FooterLayout | FooterOptions | Root、Label、Actions |
| FooterColumns | `[]FooterColumn` | Root、Body、Label、List、Item、ItemLeadingSlot、ItemLabelSlot |

## 示例

Gallery 的 **Advanced interactions**、**Activity & attachments**、**Application layout** 展示这些组件，**Props & behavior** 和 **Orientation** 演示配置变化。

Gallery 最左侧目录本身也使用 Sidebar + NavigationMenu，默认宽 256 DIP，可在 200–360 DIP 间拖动，收起为 64 DIP 图标栏。页面切换保留宽度和收起状态，每个窗口分别保存；它与 Application layout 页面里的演示侧栏相互独立。自动化测试可通过 Toggle gallery sidebar 和 Resize gallery sidebar 名称定位按钮与手柄。
