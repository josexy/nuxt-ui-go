# 高级交互、图表与布局

提供命令面板、文件选择、图表、媒体交互与桌面布局组件。状态按窗口持有，异步操作和文件资源的所有权在各节说明；平台范围见 [兼容性说明](compatibility.md)。

## 入口与定制

| 目录条目 | Go 入口 / 持久状态 | 可定制 slots |
| --- | --- | --- |
| CommandPalette | `CommandPalette` / `CommandPaletteState`、`CommandIndex` | Root、Editor、Listbox 转交的 Root/Label/Item/EmptySlot |
| NavigationMenu | `NavigationMenu` / `NavigationState`、`*string` active ID | Root、Label、Item |
| Drawer | `Drawer` / `DrawerState` | Slideover/Modal 的各 slots、Thumb |
| FileUpload | `FileUpload` / `FileUploadState` | Root、Body、Item |
| Chart | `Chart` / `ChartState` | Root、Body、Item（图例）、Label、Description、EmptySlot |
| Carousel | `Carousel` / `CarouselState` | Root、Body、Item |
| Marquee | `Marquee` / `MarqueeState` | Root |
| Splitter | `Splitter` / `*float32` | Root、Leading（第一面板）、Trailing（第二面板） |
| Sidebar | `Sidebar` / `*bool` collapsed | Root、Body；Header/Footer 是内容 builder |
| Header | `HeaderLayout` / `HeaderOptions` | Root、Label、Description、Actions |
| Footer | `FooterLayout` / `FooterOptions` | Root、Label；Actions 是内容 builder |
| FooterColumns | `FooterColumns` / `[]FooterColumn` | Root、Label、Item |

每个 Options 的 Decoration 支持组件主题 defaults → 实例 UI → 同步 Style 回调。Header、Footer 是 slot 常量，对应布局函数命名为 `HeaderLayout`、`FooterLayout`。布局尺寸仍可通过返回的当前帧 Element 调整；不得保存 Element 或 Context 到模型/工作线程。

Gallery 的 Advanced interactions、Activity & attachments、Application layout 展示这些组件，Props & behavior 和 Orientation 展示可选配置。Carousel / FileUpload 参数见 [参数与默认值](options.md)。

## CommandPalette

`NewCommandIndex([]CommandAction)` 在数据变更时创建，不要每帧创建。索引复制搜索数据，并按父动作分层；全树 ID 必须唯一、非空，最多 32 层。状态使用 `SetIndex` 更换索引。

- 默认 Ctrl+K 切换；`ShortcutKey/ShortcutMods` 可改为平台习惯，`DisableShortcut` 可关闭。动作的 `Key/Mods` 是可执行快捷键，`Shortcut` 仅负责显示。业务快捷键由应用避免冲突。
- 大小写不敏感的 Unicode 子序列搜索；连续匹配与词首加权，Keywords 可提供别名。返回 rune 索引用 RichText 高亮，不按 UTF-8 字节切片。默认不做拼音、分词或重音折叠。
- 分组、描述、图标、快捷键、禁用项、子动作与 Back。描述以 Muted 色、与标签同字号显示在标签下一行。列表用 Listbox 虚拟化；默认最多 200 个结果，配置最大 1,000。不对全部结果建立每帧 UI。
- 默认搜索在工作线程扫描不可变索引，使用至多一项的 mailbox；UI 构建轮询并发布。每 128 条检查取消，过期任务不回写。排序保存有界 top-k，不全量排序。
- 输入 Down 进入列表，Up/Down/Home/End 在列表移动，Enter 激活。正在搜索或查询与结果不一致时，Enter 不执行旧动作。OnTextKey 沿用底层组合输入屏蔽规则。
- `Open=false` 的下一次构建取消任务；`Close()` 立即取消。组件应每帧调用，包括关闭时；不再调用或窗口关闭时显式 `Close()`。无永久 goroutine/ticker。
- `Search(ctx,ticket,query,parent)` 可替换检索策略。启动应用自己的任务，并在 `Window.Update` 内调用 `state.Results.Complete(ticket, matches, err)`；外部回调自行调度重绘。返回 ID 和父层级由应用负责。状态属于一个窗口，不可跨窗口共享。

```go
state.Commands.SetIndex(nui.NewCommandIndex([]nui.CommandAction{
    {ID: "save", Label: "Save workspace", Group: "Workspace",
        Key: ui.KeyS, Mods: ui.Ctrl, Shortcut: "Ctrl S", OnSelect: save},
}))
// view: 始终调用；调用点的父作用域决定快捷键范围。
nui.CommandPalette(c, theme, &state.Commands,
    nui.CommandPaletteOptions{Title: "Workspace commands"})
// window.OnClosed(state.Commands.Close)
```

查询输入受 [输入法取消问题](compatibility.md#输入法取消) 影响；完整读屏体验尚未验证。

## NavigationMenu 与 Sidebar

`NavigationItem` 包含 ID、Label、Description、Group、Icon、Disabled、Children。`NavigationMenu(c,t,&active,&state,items,options)` 将实际路由交给 `OnNavigate(id)`；只更新业务 ID，不猜测 URL。

水平导航的子级使用已有 DropdownMenu（多层菜单、Escape、焦点恢复）；纵向采用内联展开层级。方向键、Home/End 和 Tab 使用 MyGo FocusGroup；禁用项不激活。父项用于展开，不同时跳转。全树 ID 唯一，删除项目后清理菜单状态。

Sidebar 是侧栏内容区，有标题、Header/Footer builder、可滚动 body 和收起按钮，默认宽 224/64 DIP。收起按钮使用 Nuxt 的面板图标（panel-left-close / open），分隔边框在结束侧；右到左时图标和边框一起镜像。`content(c,t,compact)` 中可组合 NavigationMenu；紧凑模式隐藏有图标项目的文字，但保留可访问名称。没有图标的项仍显示标签。长文本按可用宽度省略；应用应在更窄的窗口主动收起，或把导航放进 Drawer。

不包含 DashboardSidebar、路由匹配、Web mega menu、移动端断点系统，也不要求 URL。导航的原生无障碍角色为 Group（MyGo 没有公开 Navigation landmark 枚举）。

## Drawer

基于公开 Slideover 的底部面板、Modal 作用域、滚动正文、退出与焦点恢复。`DrawerState.Open/Snap` 可由应用控制；默认停靠点是窗口高度的 0.5/0.85，`SnapPoints` 必须严格递增且在 (0,1]。

只在独立手柄拖动，因此正文滚动不会被抽屉抢走。向下拖动减小高度，松开后按高度与速度投影选择停靠点；高度低于最小停靠点的一半，或移动超过默认 80 DIP 且速度超过 900 DIP/s，可关闭。`DismissDistance/DismissVelocity` 可配置；`Dismissible=false` 禁止拖拽、遮罩与 Escape 关闭，仍保留显式关闭操作。

手柄可聚焦：Up/Down 切换停靠点，Home/End 到首尾；关闭按钮与 Escape 提供替代路径。失去捕获、失焦取消恢复原停靠点。尺寸停靠立即完成，打开/关闭继承 Slideover 过渡及 reduced motion。组件应包括关闭状态每帧调用。

Drawer 仅支持底部手柄拖拽，不提供四边拖拽、连续弹簧吸附或与正文滚动联动的触屏手势。需要从其他方向打开面板时使用 Slideover。

## FileUpload 与资源所有权

FileUpload 继承 FormField.Disabled/Size；禁用字段会取消尚未完成的文件检查并丢弃晚到结果。默认颜色/尺寸取 ComponentFileUpload，选择按钮同样使用 ComponentFileUpload 默认值。

`FileUploadState.Files` 保存元数据，不保存文件句柄。`UploadFile` 包含 Path/Name/MIME/Size、状态、0..1 Progress、Error；Add 为每次加入分配稳定 ID，忽略重复 Path。`UploadRules` 支持扩展名（`.png`）、MIME（`text/plain`）、类别（`image/*`）、单文件 MaxBytes、已接受文件 MaxFiles。超限或读取失败的项作为 Rejected 显示原因，可移除。

- 默认 Choose 按钮在 worker 调用 `mygo.Dialog.Open`，接着用 `os.Stat` 检查普通文件。`Dialog` 可配置父窗口、多选、过滤器等；单选/多选以原生配置为准。`Choose(ctx)` 可注入自己的文件来源。
- 外部 DroppedFiles 走相同 worker 元数据检查与规则。`InspectUploadFiles` 和 `PickUploadFiles` 也是公开 helper。
- 只根据扩展名推导 MIME，不读取正文；Accept 是选择体验约束，不是可信的内容校验或上传安全检查。具体业务需要时自行验证内容。
- 不发起网络传输，也不自动把 Ready 改为上传中。应用调用 `ctx,ticket := state.Begin(id)` 开始自己的任务，在 `Window.Update` 内用 `Report(ticket,progress,done,err)` 发布进度。
- `Cancel(id)` 取消任务 context；旧票据、已完成、取消、删除、重试前的结果不能回写。`Remove(id)` 同时取消。`OnCancel/OnRemove` 可做额外业务清理。应用拥有它打开的流、临时文件和网络连接，并负责关闭。
- `Dispose()` 在隐藏/卸载或窗口关闭时调用，取消元数据检查和任务。原生对话框自身不能被 context 强制关闭；取消后丢弃其最终结果。状态属于 UI 线程；worker 不能直接改 Files。
- UI 自带 Ready/Uploading/Complete/Failed/Canceled/Rejected 展示及进度；没有内建上传重试协议。调用 Begin 可重试业务失败。

```go
// 窗口创建后配置 Parent，关闭时清理；Gallery 已示范。
options.Dialog = mygo.OpenDialogOptions{Parent: window, Multiple: true}
window.OnClosed(state.Files.Dispose)
// 由应用的上传按钮/队列调用 Begin，再启动自己的传输。
ctx, ticket := state.Files.Begin(fileID)
go func() {
    err := upload(ctx, path)
    window.Update(func() { state.Files.Report(ticket, 1, true, err) })
}()
```

原生对话框、跨应用拖放及文件筛选的平台验证范围见 [兼容性说明](compatibility.md#环境与平台)。macOS sandbox 权限与 bookmark 由应用管理。

## Chart

基于 MyGo `ui/chart` 的 Linear/Band/Time、Line/Area、Stack、Pie/Sector、Select 与 MinMax，提供 NuxtUI Go 主题、持久状态、图例交互和动画。组件由原生绘制构成，没有 WebView 或额外图表依赖。

| 能力 | API |
| --- | --- |
| 折线、柱状、面积、散点、饼图、环形图 | `ChartLine / ChartBar / ChartArea / ChartScatter / ChartPie / ChartDonut` |
| 线性、单调平滑、三种阶梯曲线 | `ChartCurveLinear / Monotone / StepAfter / StepBefore / Step`（完整常量均以 `ChartCurve` 开头） |
| 不堆叠、标准、归一化、居中堆叠 | `ChartStackNone / Standard / Normalized / Center`（完整常量均以 `ChartStack` 开头）；用于柱图和面积图 |
| 数值、分类、时间 X 轴 | `SetData`、`SetCategories`、`SetTimeData` |
| 独立的饼图／环形图数据 | `SetSlices([]ChartSlice)` |
| 图例与系列显隐 | `Legend`、`InteractiveLegend`；`SeriesVisible / SetSeriesVisible`、`SliceVisible / SetSliceVisible` |
| 格式与尺寸 | `FormatX / FormatY / FormatTime`、`Height`、`Grid`、`BarPadding`、`DonutInnerRadius`、`SlicePadding`、`LineWidth / PointSize` |

所有数据设置方法先验证再复制；错误不改动原数据。`SetData(xs, series)` 要求有限、严格递增的共享 X；每条序列与 X 等长，系列 ID 唯一且非空。NaN/Inf Y 表示缺值，不转为 0。`SetCategories(labels, series)` 使用唯一非空标签的 Band 轴；`SetTimeData(times, series)` 要求时间严格递增，刻度使用第一项的时区。散点图沿用共享、排序的 X 数据模型。

`SetSlices` 独立保存 `ChartSlice{ID, Label, Value, Color}`，可在同一状态中保留直角坐标与径向两组数据，再切换 Kind。切片 ID 唯一非空；Value 必须有限、非负，总和有限。0 值不绘制扇区，但保留图例。语义颜色随 Primary／Neutral 与明暗主题刷新；默认系列颜色与 Nuxt 的 `--ui-chart-1..6` 一致，依次为 Primary、blue、violet、amber、rose、cyan（浅色主题取 500 色阶，深色取 400），六条之后循环，也可显式指定 `Color`。

Y 轴包含 0。Standard 将正负值分别向上／下累积；Normalized 按每个 X 的可见值**绝对值总和**归一化，纯正值合计 100%，混合正负时两侧共享分母；Center 将代数和居中到 0。隐藏系列会重新计算坐标范围和堆叠；隐藏扇区后重算可见比例。隐藏状态按 ID 跨更新和重排保留，删除 ID 后清理。

图例默认可点击，使用可聚焦的复选框语义，Space / Enter 同样可切换，`OnLegendChange(id, visible)` 收到用户动作。`InteractiveLegend: nui.Set(false)` 只展示图例；`Legend: nui.Set(false)` 完全隐藏。所有项目隐藏后仍保留图例用于恢复。

直角坐标图支持鼠标最近 X、Left/Right/Home/End/Escape、底部原始数值与 Tooltip。饼／环形图支持扇区命中、方向键/Home/End/Escape 和可见总和百分比；环形中心不命中。稠密数据的选择仍查原始数据，不用采样值冒充原始值。

`Animate` 默认开启，遵循主题和系统 reduced motion，`AnimationDuration` 默认 350ms。首次显示直接呈现；数据或参数变化时，同图型、同系列及缺值结构进行坐标插值，Pie ↔ Donut 还会插值内半径。跨图型、增删系列／点、显隐或缺值结构变化使用淡入淡出。中途更新从当时画面继续，最多保留 3 层退出画面和 1 层目标画面；轴标签和提示对应目标数据。动画完成、不可见或卸载后不持续申请帧，无后台 ticker。`Animate: nui.Set(false)` 立即采用目标值。

```go
// 数据变化时调用；worker 的结果通过 Window.Update 发布到该窗口的状态。
err := state.Activity.SetCategories([]string{"Jan", "Feb", "Mar"}, []nui.ChartSeries{
    {ID: "build", Label: "Builds", Values: []float64{30, 45, 38}},
    {ID: "review", Label: "Reviews", Values: []float64{12, 19, 16}},
})
if err != nil { return err }
// view
nui.Chart(c, theme, &state.Activity, nui.ChartOptions{
    Label: "Activity", Kind: nui.ChartArea,
    Curve: nui.ChartCurveMonotone, Stack: nui.ChartStackStandard,
    Height: 240, AnimationDuration: 400 * time.Millisecond,
})
```

`Prepare(pixelWidth)` 可预热默认折线缓存；其他选项由 Chart 准备。采样宽度上限 2,048，仅在数据／显隐、尺寸或几何选项变化时重做。绘制和插值只处理有界采样几何。未堆叠数据使用 MinMax；缺值片段过多时改用固定桶，含缺值的桶保持断开。密集堆叠共用桶的端点和总体包络极值，保持边界 X 对齐；单个系列的像素内细节可能被省略。稠密柱图不做求和／平均，业务聚合应先完成。

`ChartState` 属于单个窗口、单个挂载图表，不复制已渲染状态，不在 worker 与 UI 间共享可变数据。当前没有双 Y 轴、对数轴、刷选缩放或触屏手势；低层自定义标记和布局可直接组合 MyGo `ui/chart`。Gallery 的 Activity & attachments 展示全部图型、轴、曲线、堆叠、数据更新、缺值／负值与图例显隐。

## Carousel 与 Marquee

CarouselItem 使用稳定 ID、Label 和同步 Build。`CarouselState.Index/Paused` 可控；每次只构建当前及相邻项，裁剪内沿 Orientation 拖拽，位移超过视口 20% 或速度超过 600 DIP/s 时换页，释放后吸附到页。Previous/Next、水平 Left/Right 或纵向 Up/Down、Home/End 可操作；Arrows 可隐藏按钮，Dots 可开启圆点导航。`Wrap=true` 让按钮、键盘与自动播放到尾后回到首项；拖拽边缘有界，不是无缝克隆循环。索引以位置为控制值；重排业务数据时由应用决定新的 Index。

默认无自动播放；设置 Autoplay 才启用。悬停、焦点在幻灯片、手动暂停、Active=false、窗口外不可见、reduced motion 都停止推进。移除项会修正索引，拖拽取消归零。相邻项禁用交互，但其 builder 仍会执行；不要在 builder 启动副作用。

Marquee 支持水平/垂直、Reverse、DIP/s Speed、Gap、Height、内容重复和 Space 暂停。ContentKey 改变重置相位；内容短于容器时不滚动；PauseOnHover=true 时悬停暂停（默认 false）；焦点、Paused、Active=false、reduced motion 始终停止。Orientation 与 Vertical 都可选择方向，水平运动随 RTL 镜像。内容 builder 只放展示元素，会被调用两次，不能放业务按钮、输入或有副作用的内容。底层没有整棵子树的 accessibility-hidden 接口，重复内容的读屏呈现仍需在目标读屏工具中验证。

两者没有定时 goroutine。可见性判断使用已提交窗口矩形；对仍挂载但被应用自定义裁剪/隐藏的容器，调用方必须同步 `Active=false`，或者不构建组件；窗口隐藏/最小化可由应用事件传入 Active。卸载后不再申请帧，不累积暂停时间。不要将停止申请帧误解为可随意持续构建隐藏动画。

## Splitter、Header、Footer、FooterColumns

Splitter 使用 native Split/SplitVertical/SplitPercent/SplitVerticalPercent。值单位默认为 DIP，Percent=true 为 0..100，Vertical=true 上下分割。Min/Max 为相同单位的模型约束；指针和键盘操作后会钳制并通知 OnChange。应用直接改值会在下次构建钳制。

原生面板保留主题 40 DIP 最小值，窗口小于约束时按原生 fit 行为收缩；无障碍 divider Range 使用原生 DIP 可用范围，不反映自定义百分比约束。样式支持 Root、Leading、Trailing，分隔线和手柄继承 Theme.Native。未提供多面板、折叠停靠或布局序列化协议。

HeaderLayout 包含 Leading、标题/说明与 Actions，FooterLayout 包含 Text 与 Actions；可换行、伸缩、跟随字体，不改系统标题栏。FooterColumns 接收稳定列 ID、标题和简单导航 Items，按最小列宽自动换列；业务导航由 OnNavigate 接管，嵌套导航使用 NavigationMenu。窄窗口的总体布局优先级仍由应用决定。
