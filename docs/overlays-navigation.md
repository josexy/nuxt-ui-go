# 弹层、菜单与导航

对话框和通知用于处理当前操作，Tabs、分页和步骤条用于组织内容。先根据交互目的选组件：

| 需要什么 | 组件 |
| --- | --- |
| 让用户先完成当前操作，再回到页面 | [Modal / Slideover](#modal-与-slideover) |
| 简短提示，或在控件旁补充内容 | [Tooltip](#tooltip)、[Popover](inputs.md#popover-与-scrollarea) |
| 下拉操作、右键菜单或操作结果通知 | [DropdownMenu / ContextMenu](#dropdownmenu-与-contextmenu)、[Toast](#toast) |
| 切换内容、展开详情 | [Tabs / Collapsible / Accordion](#tabscollapsible-与-accordion) |
| 页面路径、分页、流程、时间线或错误展示 | [导航组件](#breadcrumbpaginationsteppertimeline)、[Banner / ErrorView](#banner-与-errorview) |

## 公共约定

### 打开状态和稳定 ID

应用保存弹层是否打开、当前选中项及其他持久状态。每个实例在创建前使用 `c.Key(...)`，构建所得 Element 和部件只用于当次界面，不存入模型或后台任务。关闭时返回的零值 Element 表示当前没有可用面板。

弹层组件应持续出现在视图函数中，包括关闭状态。这样组件才能完成退出动画、清理和焦点恢复，避免用 `if open` 包住整个组件调用。

列表项 ID 必须非空且唯一，重复会 panic。Tabs、Accordion、Breadcrumb、Stepper、Timeline 要求各自列表内唯一，菜单要求整棵树唯一。显示名称可以重复。必需的状态指针不能为 nil；Banner 明确支持 nil，表示常驻。

### 样式和回调

多数组件返回 `ui.Element`，Toast 返回通知 ID，CloseToast 不返回元素。Style 在当前构建中同步配置部件，业务动作使用点击或选择回调。

样式从内置样式、主题组件配置、实例 UI 依次应用，最后执行 Style。各组件支持的 Color、Size、Variant 和部件不同，见末尾[部件索引](#部件索引)。尺寸使用 DIP，无需再乘屏幕缩放比例。

## Modal 与 Slideover

### 打开一个编辑对话框

Modal 居中显示，Slideover 从窗口边缘滑入。两者都属于模态弹层，打开时背景不可操作。

```go
// state 按窗口保存 EditOpen、Draft 和 EditTrigger（ui.Handle）。
edit := nui.Button(c, t, "编辑", nui.ButtonOptions{}).
    Bind(&state.EditTrigger).
    OnClick(func() { state.EditOpen = true })
nui.Tooltip(c, t, edit, "编辑当前资料", nui.TooltipOptions{})

// 每次构建都调用，包括 EditOpen=false 时。
nui.ModalWithOptions(c.Key("edit"), t, &state.EditOpen, nui.ModalOptions{
    Title: "编辑资料",
    Description: "更改只在保存后应用。",
    ReturnFocus: &state.EditTrigger,
    Footer: func(c *ui.Context, inner nui.Theme) {
        nui.Button(c, inner, "完成", nui.ButtonOptions{}).
            OnClick(func() { state.EditOpen = false })
    },
}, func(inner nui.Theme) {
    nui.Input(c, inner, &state.Draft, nui.InputOptions{Label: "名称"})
})
```

上例只演示打开、编辑草稿和关闭。真正保存数据时，由应用在完成动作中处理。简单对话框也可使用 `Modal(c, t, &open, title, body)`；完整配置使用 ModalWithOptions。

### 关闭和焦点

| ModalOptions | 行为 |
| --- | --- |
| Dismissible | Optional[bool]，默认 true；false 阻止遮罩点击和 Escape 关闭 |
| CloseButton / CloseLabel | 默认显示关闭按钮，可单独隐藏或更换名称 |
| OnClose | 用户通过组件关闭按钮、允许的遮罩或 Escape 关闭时调用一次；应用直接设 open=false 不调用 |
| ReturnFocus | 指定关闭后恢复焦点的 Handle |

Dismissible 与 CloseButton 独立。需要禁止用户关闭时，两项都设为 `nui.Set(false)`。从菜单项打开对话框时，ReturnFocus 通常应指向仍存在的菜单触发按钮，因为菜单项会随菜单关闭而卸载。

Tab 焦点限制在模态内容中，内部 Select / Popover 会先处理 Escape。默认聚焦第一个可聚焦控件，通常是关闭按钮；需要先聚焦字段时，在打开动作中请求该字段 Handle。Header / Footer 过高时，应用需要控制内容尺寸，为正文保留空间。

### 内容、尺寸和外观

| 选项 | 行为 |
| --- | --- |
| Title / Description | 对话框名称和描述，也显示在默认标题区 |
| Header / Footer | `func(*ui.Context, Theme)`，分别自定义标题区和操作区 |
| Width / MaxHeight | Optional[float32]；默认宽 512 DIP，限制在窗口宽减 32、高减 64 内 |
| Overlay | 默认 true；false 只隐藏遮罩颜色，仍保留模态焦点限制 |
| Fullscreen | 使用整窗尺寸，去掉圆角和边框；仅用于 Modal |

正文可滚动，标题和 Footer 留在原位置。标题区上下内边距 16、左右 24、最低高度 64 DIP；正文内边距 24 DIP，Footer 间距 6 DIP。只在实际存在的区域之间画分隔线，body 为 nil 时不创建正文区。

Footer 默认右对齐。PartStyle 不能设置对齐方式，需要其他排列时，在 Footer 回调内自行创建布局。

Modal 面板圆角为 8 DIP，遮罩使用 Elevated 色的 75% 不透明度。打开时，面板在 200ms 内从 95% 大小淡入，遮罩同步淡入；减少动态效果时直接显示。Modal 关闭立即完成，原因与 [Popover](inputs.md#popover-与-scrollarea) 相同。

### Slideover：从边缘打开

Slideover 复用 Modal 的标题、正文、Footer 和关闭配置，通过 Side / Extent 决定位置和大小：

| Side | Extent 的含义和默认值 |
| --- | --- |
| SideRight（默认）/ SideLeft | 面板宽度，默认 448 DIP |
| SideTop / SideBottom | 面板高度；未设置时按内容高度，最多占满窗口 |

显式 Extent 最多占满对应窗口轴。嵌入 ModalOptions 的 Width、MaxHeight、Fullscreen 不用于 Slideover；正文伸展并滚动，Footer 位于面板底部。

打开和关闭使用固定 200ms 的 cubic ease-out 位移及遮罩过渡；Theme.Motion.Reduce 时立即切换。退出过程中保留模态范围并禁止内容操作，结束后卸载和恢复焦点；中途反向切换会从当前进度继续。

Slideover 不提供边缘拉伸。需要底部手柄拖动和多个停靠高度时使用 [Drawer](advanced.md#drawer)。

## Tooltip

Tooltip 用于无交互的短提示：`Tooltip(c, t, anchor, text, options)`。它把文字设为锚点 Description，响应鼠标悬停和键盘焦点。该设置会替换已有 Description，应用需要提前合并所需描述。

| 行为 | 默认值 |
| --- | --- |
| 位置和间距 | 下方居中，距锚点 8 DIP |
| MaxWidth | 320 DIP，单行超长时省略 |
| 外观 | 高 24 DIP、圆角 4 DIP、水平内边距 10 DIP、小阴影 |
| 鼠标触发 | 延迟 600ms；连续查看提示时有 400ms 的快速切换时间窗口 |
| 键盘触发 | 聚焦后立即显示 |

指针离开、按下或 Escape 会隐藏提示；Disabled 或无效、禁用锚点不会显示。打开动画为 100ms 缩放淡入，当前没有延迟时长选项。需要多行或可点击内容时，使用 Popover。

Kbds 显示快捷键，文字和快捷键间以 `·` 分隔，各键间距 2 DIP。受控 Open 和位置选项见[参数说明](options.md#弹层和通知)。

## DropdownMenu 与 ContextMenu

### 创建操作菜单

DropdownMenu 从按钮打开；ContextMenu 可从可聚焦区域的右键、Menu 键或 Shift+F10 打开。

```go
anchor := nui.Button(c, t, "操作", nui.ButtonOptions{})
nui.DropdownMenu(c.Key("actions"), t, anchor, &state.Menu, []nui.MenuItem{
    {ID: "pin", Label: "置顶", Checked: nui.Set(state.Pinned),
        OnSelect: func() { state.Pinned = !state.Pinned }},
    {ID: "export", Label: "导出", Children: []nui.MenuItem{
        {ID: "csv", Label: "CSV", OnSelect: state.ExportCSV},
        {ID: "pdf", Label: "PDF", Disabled: true},
    }},
}, nui.MenuOptions{Label: "文档操作"})
```

每个窗口、每个独立触发器各保存一份 MenuState，应用可读写 Open。DropdownMenu 设置锚点的菜单按钮和展开语义；ContextMenu 的鼠标打开位置为右键点，键盘打开位置在锚点下方。关闭时返回零值面板。

### 菜单项配置

| MenuItem 字段 | 行为 |
| --- | --- |
| ID / Label / Description / Icon | 稳定标识、名称、说明和图标；ID 在整棵菜单树中唯一 |
| Kind | MenuAction（默认）、MenuLabel、MenuSeparator；分组标题和分隔符也要有 ID |
| Shortcut | 在行尾显示快捷键，不自动注册快捷键 |
| Disabled | 禁止聚焦和激活 |
| Color | 非 Neutral 时，为文字和图标着色，高亮背景使用该颜色的 10% 不透明度 |
| Checked / Radio | Checked 是 Optional[bool]；使用复选或单选菜单项语义，值和互斥关系由应用在 OnSelect 中更新 |
| Children | 子菜单；支持鼠标和键盘打开，最多 32 层 |
| OnSelect | 叶项激活后先关闭并请求恢复触发器焦点，再调用一次；父项只打开子菜单 |
| Build | 自定义展示内容，仍需 Label；内容中不要嵌入独立输入框或按钮 |

菜单行采用 MD 的[选项行样式](inputs.md#选项行的外观)：内边距 6 DIP、图标 20 DIP，描述与标签同字号。选中对勾在行尾，Icon 仍在行首，子菜单箭头随 RTL 调整。

MenuOptions 提供 Label、Width、MaxHeight 和 Decoration。宽度默认按内容，至少 128 DIP；最大高度默认 320 DIP，长菜单滚动。面板限制在窗口内，子菜单由 AttachTo 避让边缘。

Shortcut 按 `+` 拆分为 MD Kbd，`Ctrl++` 的末尾 `+` 表示加号键。FormatKey 将键名转换为对应平台的显示形式。

### 键盘和指针

| 操作 | 行为 |
| --- | --- |
| ↑↓ / Home / End | 在可用项中移动 |
| → / ← | 打开子菜单 / 返回父菜单，RTL 下交换 |
| Enter / Space | 激活当前项 |
| Escape | 关闭当前层；根层关闭后恢复锚点焦点 |
| Tab / Shift+Tab | 关闭菜单，继续 MyGo 的正常焦点遍历 |

直接输入字符可做忽略大小写的前缀查找，连续输入间隔为 1 秒，重复首字母可循环匹配。此功能不启用 IME 候选，也不做模糊搜索。

指针移入项目会改变高亮，悬停 180ms 打开子菜单。键盘操作会取消待打开的悬停任务，避免 Escape 后马上重开。项目重排按 ID 保留焦点，删除后清理相关状态。当前没有指针移动轨迹容错或触屏长按打开。

### 与现有事件组合

DropdownMenu 接管锚点 OnClick，ContextMenu 接管 HandleInput，不会自动组合已有的同类处理器。不同弹层不要共享同一个触发器状态。

菜单是非模态的，外部点击会关闭并继续传给目标。窗口失焦时不显示焦点，也不执行焦点组方向键；后台重建不会反复将焦点拉回第一项。

## Toast

### 添加通知区域并发布通知

每个窗口构建一次 ToastViewport；在用户动作中调用 Toast：

```go
t := nui.App(c, config)
// 放在其他弹层前，模态遮罩才能盖住通知。
nui.ToastViewport(c, t, nui.ToastViewportOptions{})
nui.Button(c, t, "保存", nui.ButtonOptions{}).OnClick(func() {
    previous := state.Value
    state.Value = state.Draft
    nui.Toast(c, t, nui.ToastOptions{
        ID: "saved",
        Title: "已保存",
        Color: nui.Success,
        Action: "撤销",
        OnAction: func() { state.Value = previous },
    })
})
```

不要在视图构建中无条件调用 Toast，否则每次重建都可能发布通知。后台结果通过 Window.Update 写回模型，再由窗口 UI 动作或明确控制的事件处理发布通知。

### 标识、时长和动作

| 选项或操作 | 行为 |
| --- | --- |
| ID | 空值自动生成；同 ID 替换通知并重置时长，不调用旧项 OnClose |
| Timeout | 默认 5 秒，负值表示常驻 |
| Action / OnAction | 单个动作，执行后关闭通知并调用 OnClose |
| Actions | 多个 ToastActionOptions；默认执行后关闭，单项 Close=false 可保留 |
| ToastOptions.Close | `nui.Set(false)` 隐藏关闭按钮，不改变超时或动作关闭规则 |
| OnClose / CloseToast | 关闭按钮、到时或 CloseToast 会调用 OnClose；`CloseToast(c, "")` 关闭全部通知 |

单动作和 Actions 可以同时使用。Actions 中每个按钮可设置 Label、OnAction、ButtonOptions 和 Close；默认 XS、Solid，颜色继承通知。没有传 Icon 时，不自动添加状态图标。

OnAction、OnClose 及 Actions 中的 ButtonOptions.Style 会由队列保存，闭包只应捕获业务模型、数据或服务。不要捕获构建产生的 Element、Parts 或 Context；Style 接收到的元素也只在当次构建有效。

### 位置、暂停和样式

ToastViewport 默认右下角、宽 384 DIP，还支持左下、中下、右上、左上、中上。Limit 支持 1–3，0 或越界按 3 处理。MyGo 最多提供最新 3 条可见通知，较早的排队项仍计时，Limit 不表示无限队列容量。

悬停通知区域、焦点在通知内或窗口失焦时，队列计时暂停。默认倒计时条同步暂停，`Progress: nui.Set(false)` 可隐藏；常驻通知不显示倒计时。进入和退出过渡为 150ms，减少动态效果时直接切换，空队列不绘制面板。

通知使用 RoleStatus，动作和关闭按钮可用键盘访问；实际读屏播报尚未验证。

每条 Toast 保存创建时的 Theme 和 UI 副本。显示期间全局主题改变，已有通知仍使用原主题；需要更新同一条通知时，用相同 ID 再次发布。Viewport.Decoration 只影响 Root，每条通知通过 ToastOptions.UI 或创建时的 ComponentToast 配置。

## Tabs、Collapsible 与 Accordion

### 选择组件和状态

| 组件 | 绑定状态 | 默认行为 |
| --- | --- | --- |
| Tabs | `*string` 当前项 ID | 胶囊样式，只显示当前内容；可选 Underline、Vertical |
| Collapsible | `*bool` 是否展开 | 点击或 Enter/Space 切换，展开/收起默认 200ms |
| Accordion | `*[]string` 展开项 ID | 默认只展开一项，可再次关闭；Multiple 开启多项 |

TabItem 包含 ID、Label、Icon、Disabled、Content；AccordionItem 包含相应的标题和内容字段。Tabs / Accordion 的内容回调接收 Context、Theme，Collapsible 正文接收 Theme 并使用外层 c。

Tabs 用方向键跳过禁用项；当前项被删或禁用后转到首个可用项，没有可用项时保留业务 ID，但不显示正文。Accordion 标题间用 ↑↓、Home、End 移动，Tab 保留标题与正文的正常顺序。

Tabs.OnChange 接收新 ID，Collapsible.OnChange 接收 bool，Accordion.OnChange 无参数，由应用读取 ID 列表。应用直接改值、Tabs 自动修正无效 ID、Accordion 单开模式截断多余 ID，都不会触发用户变更回调。

Accordion 保留应用 open 列表中暂时消失的 ID，便于同 ID 再出现；永久删除时由应用清理。手动 Tabs 激活和 Accordion 的 Collapsible 选项见[参数说明](options.md#滑块和内容切换)。

### 是否保留隐藏内容

默认仅构建活动 Tabs 的正文，折叠内容在退出动画结束后卸载。应用模型保留，但编辑器选区、撤销历史和局部 UI 状态可能在卸载后丢失。

KeepMounted=true 会在每次构建中保留所有正文，通过 Hidden 隐藏非活动内容，因而保留 MyGo 局部状态。隐藏项不参与布局和焦点，但内容回调仍执行。折叠保留模式即时隐藏正文，箭头仍按原生动画进度变化；删除整个条目仍会卸载。

### 外观

Tabs 的胶囊活动项使用语义色背景和反色文字，水平列表填满宽度。Underline 列表有 1 DIP 基线和活动指示线，横向在底部、纵向在起始侧，随 RTL 调整。触发器间距按 XS 到 XL 为 4/6/6/8/8 DIP。

Accordion / Collapsible 的触发器无悬停底色和水平内边距，上下内边距 14 DIP，图标 20 DIP、间距 6 DIP。Accordion 正文使用默认文字色，下边距 14 DIP。

## Breadcrumb、Pagination、Stepper、Timeline

### Breadcrumb：显示当前位置

BreadcrumbItem 包含 ID、Label、Icon、Disabled、OnSelect，Options 提供 Label 和 MaxItems。最后一项是不可点击的当前页，其余项作为链接；应用负责提供路径和导航动作。

当前项使用 Primary 色和半粗体，可通过 ComponentBreadcrumb.Color 改色。祖先项为 14 DIP、无内边距的 Muted 链接，悬停变为 Text 色。分隔箭头 20 DIP、项间距 6 DIP，RTL 下箭头反向。

MaxItems≥2 时保留第一项及末尾若干项，中间用不可交互的省略号代替。目前没有省略项下拉菜单或路由系统。

### Pagination：控制页码

Pagination 绑定 `*int`，页码从 1 开始。Total 是总条数，PageSize≤0 时默认 10。总数减少时会自动把页码限制到有效范围；这类修正不触发 OnChange，只有用户切到不同页时才通知。

组件不查询数据，应用收到变化后自行加载当前页。Siblings≤0 默认 2，最大 10；显式零相邻页使用 SiblingCount。ShowEdges 默认 false，Size 默认 MD，更多外观参数见[分页选项](options.md#pagination-的相邻页和外观)。

翻页控件和页码采用方形按钮，边界按钮禁用，箭头随 RTL 调整。省略号不可交互。首页、上一页、下一页、末页和第 N 页的辅助技术名称随语言切换。

### Stepper：选择流程步骤

Stepper 绑定当前步骤 `*string`，StepItem 包含 ID、Title、Description、Icon、Disabled。未知 ID 表示尚未开始，不自动选择第一项。

Linear 是 Optional[bool]，默认 true，允许返回或前进到紧邻的下一步；设为 `nui.Set(false)` 可选择任意可用步骤。CanSelect 能在实际激活时否决操作。方向键只移动焦点，Enter/Space 或点击圆点才激活。组件不会自动跳过线性流程中的禁用步骤，业务流程需由应用安排。

横向默认等宽，圆点居中，文字在下方；Vertical 把文字放在圆点旁，连接线改为纵向。Size 默认 MD，XS 到 XL 的圆点直径为 24/32/40/48/56 DIP，随 Density 缩放。之前的步骤和连接线使用完成色，当前步骤高亮；原图标或数字保留，不自动换成对勾。

### Timeline：展示时间线

TimelineItem 包含 ID、Title、Description、Time、Icon、Color、Content，还可提供 Avatar 和 UI。默认纵向，Horizontal 将指示图标及连接线放在文字上方。

Current 是从 0 开始的可选索引，CurrentID 优先；均未设置时没有活动或完成项。Reverse 改变进度解释，不重排数据。点击只调用 OnSelect，不自动改进度。

Time 是应用格式化后的字符串。组件使用列表语义，不自动排序、换算时区或调度动画。

## Banner 与 ErrorView

### Banner：页面内通知

Banner 放在正常布局中，由应用决定位置。open=false 隐藏，nil 表示常驻且无关闭按钮，非 nil 的打开状态可被关闭。Options 包含 Title、Description、CloseLabel、Color、Icon、Actions、OnClose 和 Decoration。

背景使用语义色 Solid 样式，文字可换行，操作区支持多个按钮。关闭按钮悬停或键盘聚焦时显示 10% 背景色。极窄区域需要由应用减少操作，或改为纵向布局；Banner 不会自动固定在窗口顶部。

### ErrorView：错误状态页面

Go 中 `nui.Error` 已用于语义颜色，因此错误展示函数名为 ErrorView。ErrorOptions 提供 Status、Title、Description、RetryLabel、Icon、OnRetry、Actions 和 Decoration。

只有传入 Icon 时才显示图标。默认状态码为 16 DIP 半粗体 Primary，标题 48 DIP 粗体，说明 18 DIP Muted。默认重试按钮为 LG、Primary、Solid，文案为重试；没有 OnRetry 时不创建按钮，Actions 可替代默认操作区。

嵌入页面局部时可通过 UI[Label] 等降低字号。组件只展示错误，panic 处理、请求、导航及重试状态由应用负责。

## 部件索引

| Component 常量后缀 | 可用部件 |
| --- | --- |
| Tooltip | Panel、Label |
| Modal / Slideover | Backdrop、Panel、Header、Label、Description、Close、Body、Footer |
| Toast | Root（仅 viewport）、Panel、Leading、Label、Description、Actions、Close |
| DropdownMenu / ContextMenu | Panel、Header（分组）、Item、Leading、Label、Description、Trailing；Build 替代 Label / Description 内容 |
| Tabs | Root、List、Trigger、Leading、Label、Body |
| Collapsible | Root、Trigger、Leading、Label、Trailing、Body |
| Accordion | Root、Item、Trigger、Leading、Label、Trailing、Body |
| Breadcrumb / Pagination | Root、Item |
| Stepper | Root、Item、Header、Trigger、Indicator、Leading、SeparatorSlot、Body、Label、Description |
| Timeline | Root、Item、Indicator、SeparatorSlot、Body、Label、Description |
| Banner | Root、Leading、Label、Description、Actions、Close |
| Error | Root、Leading、Label、Description、Actions |

Stepper 的 Header 是圆点和连线容器，Trigger 是圆形按钮，Leading 为图标，SeparatorSlot 为连接线，Body 为文字区域。

Tabs、Stepper、Timeline 接受全局和实例 Color / Size；Banner、Toast 接受 Color。菜单项和 Timeline 条目的 Color 是各条目自己的设置，其他样式通过已列出的部件配置。

## 示例

Gallery 的 **Overlays & menus** 和 **Navigation & disclosure** 展示弹层、菜单和导航。Settings 示例包含编辑、校验、重置确认、Toast 撤销和通知筛选。示例状态保存在窗口模型中，长期存储仍由应用实现。

平台和输入法限制见[兼容性说明](compatibility.md)。
