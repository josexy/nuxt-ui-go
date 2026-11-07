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
| Close / CloseIcon | 配置关闭按钮的颜色、尺寸、加载及图标；默认 MD、Neutral、Ghost 和 X，Close.Leading 优先于 CloseIcon |
| OnClose | 用户通过组件关闭按钮、允许的遮罩或 Escape 关闭时调用一次；应用直接设 open=false 不调用 |
| OnClosePrevent | Dismissible=false 阻止 Escape 或外部点击时调用一次；关闭按钮不受 Dismissible 限制 |
| ReturnFocus | 指定关闭后恢复焦点的 Handle |

Dismissible 与 CloseButton 独立。需要禁止用户关闭时，两项都设为 `nui.Set(false)`。从菜单项打开对话框时，ReturnFocus 通常应指向仍存在的菜单触发按钮，因为菜单项会随菜单关闭而卸载。

Tab 焦点限制在模态内容中，内部 Select / Popover 会先处理 Escape。默认聚焦第一个可聚焦控件，通常是关闭按钮；需要先聚焦字段时，在打开动作中请求该字段 Handle。Header / Footer 过高时，应用需要控制内容尺寸，为正文保留空间。

### 内容、尺寸和外观

| 选项 | 行为 |
| --- | --- |
| Title / Description | 对话框名称和描述，也显示在默认标题区 |
| Header / Footer | `func(*ui.Context, Theme)`；Header 替换标题与说明的内容，关闭按钮仍独立创建；Footer 自定义操作区 |
| Width / MaxHeight | Optional[float32]；默认宽 512 DIP，限制在窗口宽减 32、高减 64 内 |
| Overlay | 默认 true；false 只隐藏遮罩颜色，仍保留模态焦点限制 |
| Fullscreen | 使用整窗尺寸，去掉圆角和边框；仅用于 Modal |
| Transition | Optional[bool]，默认 true；false 跳过入场动效，恢复 true 不重播已经完成的入场 |

正文可滚动，标题和 Footer 留在原位置。标题区上下内边距 16、左右 24、最低高度 64 DIP；正文内边距 24 DIP，Footer 间距 6 DIP。Title、Description、Header 均为空且 CloseButton=false 时不创建标题区；body 为 nil 时不创建正文区；仅有 Footer 时没有孤立分隔线。

Panel 的文字样式直接继承到普通正文元素，Panel.Gap 控制实际区域之间的间距，默认 0。内部内容按面板圆角裁剪。标题区、正文和 Footer 使用稳定身份；移除标题区或切换 Fullscreen 会保留正文编辑器的焦点与选区。关闭时会卸载内容，应用持久保存的值仍保留，不保证临时控件状态在重新打开后延续。

Footer 默认右对齐。PartStyle 不能设置对齐方式，需要其他排列时，在 Footer 回调内自行创建布局。

Modal 面板圆角为 8 DIP，遮罩使用 Elevated 色的 75% 不透明度。打开时，面板在 200ms 内从 95% 大小淡入，遮罩同步淡入；Transition=false 或减少动态效果时直接显示。Modal 关闭立即完成，原因与 [Popover](inputs.md#popover-与-scrollarea) 相同。没有上游的整层滚动模式、非模态开关或保持挂载选项。

### Slideover：从边缘打开

Slideover 复用 Modal 的标题、正文、Footer 和关闭配置，通过 Side / Extent 决定位置和大小：

| Side | Extent 的含义和默认值 |
| --- | --- |
| SideRight（默认）/ SideLeft | 面板宽度，默认 448 DIP |
| SideTop / SideBottom | 面板高度；未设置时按内容高度，最多占满窗口 |

显式 Extent 最多占满对应窗口轴。嵌入 ModalOptions 的 Width、MaxHeight、Fullscreen 不用于 Slideover；正文伸展并滚动，Footer 位于面板底部。

Inset 默认 false，启用后在窗口四周保留 16 DIP 边距并使用 8 DIP 圆角；边距随主题间距比例变化，Extent 限制到扣除边距后的空间。Side 指定物理边缘，不随 RTL 对调。左右默认宽 448 DIP，Go 不使用上游小屏全宽的断点规则。

Slideover 始终创建正文区域，即使 content=nil 也保留默认的 24 DIP 内边距。顶部和底部按内容高度时，这会影响面板高度；与 Modal 的 nil 正文不创建区域的规则不同。Body、Header 和 Footer 使用稳定身份，切换方向或 Inset 保留编辑器焦点与选区。

打开和关闭使用固定 200ms 的 cubic ease-out 位移及遮罩过渡；Transition=false、Theme.Motion.Reduce 或系统减少动态效果时立即切换。退出过程中保留模态范围并禁止内容操作，结束后卸载和恢复焦点；中途反向切换会从当前进度继续。

Backdrop 的主题或实例 Background 覆盖也参与渐入渐出，最终 Style 仍有最高优先级。原生 Backdrop 包含面板，设置它的 Opacity 会同时影响子面板；只隐藏遮罩应使用 Overlay=false 或透明背景。继承字段的适用范围见[参数说明](options.md#弹层和通知)。

Slideover 不提供边缘拉伸。需要底部手柄拖动和多个停靠高度时使用 [Drawer](advanced.md#drawer)。

## Tooltip

Tooltip 用于无交互的短提示：`Tooltip(c, t, anchor, text, options)`。它响应鼠标悬停和键盘焦点，并把文字与格式化后的 Kbds 追加到锚点 Description，保留已有说明。Kbds 只显示提示，不注册快捷键动作。

| 行为 | 默认值 |
| --- | --- |
| 位置和间距 | 下方居中，距锚点 8 DIP |
| MaxWidth | 320 DIP，单行超长时省略 |
| 外观 | 高 24 DIP、圆角 4 DIP、水平内边距 10 DIP、小阴影 |
| 鼠标触发 | 未设置 DelayDuration 时延迟 600ms；连续查看提示时有 400ms 的快速切换时间窗口 |
| 键盘触发 | 聚焦后立即显示 |

`DelayDuration` 是 `Optional[time.Duration]`，例如 `nui.Set(200 * time.Millisecond)`；显式 0 或负值立即显示。离开锚点会取消等待，键盘聚焦不等待悬停延迟。Nuxt UI 的默认延迟为 700ms、快速切换时间窗口为 300ms；Go 保留 MyGo 的原生默认值。

指针离开、点击触发器、外部按下或 Escape 会隐藏提示；Disabled、无效或禁用锚点，以及文字和 Kbds 同时为空时也会关闭。打开动画为 100ms 缩放淡入，减少动态效果时立即显示，关闭时立即消失。Go Tooltip 不接收指针输入或焦点，没有上游的可悬停内容、箭头和任意内容插槽；需要多行或可点击内容时使用 Popover。

Kbds 默认使用 SM 尺寸，文字和快捷键间以 `·` 分隔，各键间距 2 DIP；与上游在较窄屏幕隐藏快捷键的规则不同，Go 始终显示它们。Panel.Gap 控制正文与快捷键容器的间距，Trailing 配置快捷键及分隔点的容器。正文单行省略，快捷键不收缩，因此 MaxWidth 需要为它们留出空间。受控 Open 和位置选项见[参数说明](options.md#弹层和通知)。

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

每个窗口、每个独立触发器各保存一份 MenuState，应用可读写 Open。DropdownMenu 设置锚点的菜单按钮和展开语义。ContextMenu 将右键位置保存为窗口内坐标，再次右键只重新定位，不重复通知打开；目标布局移动时菜单保持原点击位置。Menu 键和 Shift+F10 使用目标左下角作为定位点，应用直接打开则使用当前目标区域。关闭时返回零值面板，并清除旧指针位置。

### 菜单项配置

| MenuItem 字段 | 行为 |
| --- | --- |
| ID / Label / Description / Icon | 稳定标识、名称、说明和图标；ID 在整棵菜单树中唯一 |
| Kind | MenuAction（默认）、MenuLabel、MenuSeparator；分组标题和分隔符也要有 ID |
| Shortcut | 在行尾显示快捷键，不自动注册快捷键 |
| Avatar / Loading | Loading 优先，其次 Icon，再次 Avatar；加载图标不自动禁用项目 |
| Disabled | 禁止聚焦和激活 |
| Color | 非 Neutral 时，为文字和图标着色，高亮背景使用该颜色的 10% 不透明度 |
| Checked / Radio | Checked 是 Optional[bool]；显式 false/true 使用复选语义。Radio 为原生扩展，须配合 Checked，互斥关系由应用维护 |
| OnCheckedChange / Close | 已设置 Checked 的叶项激活时通知新值一次，Radio 通知 true；Close 默认 true，设为 `nui.Set(false)` 可保留菜单 |
| Children | 子菜单；支持鼠标和键盘打开，最多 32 层 |
| Content | 单个子菜单的 MenuContentOptions，可覆盖宽高、位置、循环、筛选和部件；尺寸和图标继承根菜单 |
| UI | 覆盖本项目部件，优先于 MenuOptions.UI；子菜单面板使用 Content.UI |
| OnSelect | 叶项默认先关闭并请求恢复触发器焦点，再通知勾选变化和选择；Close=false 时保留菜单。父项只打开子菜单 |
| Build | 自定义展示内容，仍需 Label；内容中不要嵌入独立输入框或按钮 |

MenuOptions.Size 默认 MD，可继承组件主题，也可显式设置 XS–XL。菜单使用[选项行样式](inputs.md#选项行的外观)，MD 内边距为 6 DIP、图标为 20 DIP；描述与标签同字号，图标和行尾内容对齐首行。头像默认直径随尺寸为 16 / 16 / 20 / 20 / 24 DIP，显式 Avatar.Size 优先。CheckedIcon 默认 Check，LoadingIcon 默认 LoaderCircle；子菜单箭头随 RTL 调整。

MenuOptions.Label 设置面板名称。Width 默认按内容，至少 128 DIP；MaxHeight 默认 320 DIP，长菜单滚动，显式 0 收起可见内容。尺寸限制到窗口范围。DropdownMenu 根菜单默认 BottomStart、间隔 8 DIP；ContextMenu 默认 RightStart、间隔 2 DIP。子菜单按阅读方向向外打开，默认间隔 3 DIP。Placement 和 Offset 都允许显式覆盖，其中 ContextMenu 的覆盖能力为原生扩展；菜单通过 AttachTo 避让窗口边缘，没有自定义碰撞区域或箭头配置。

Shortcut 按 `+` 拆分为 Kbd，`Ctrl++` 的末尾 `+` 表示加号键。FormatKey 将键名转换为对应平台的显示形式。XS/SM 菜单使用 SM Kbd，MD/LG 使用 MD，XL 使用 LG；键间距分别为 2 / 2 / 2 / 4 / 4 DIP。Go 在窄窗口仍显示快捷键。

### 筛选与受控状态

Filter 默认 false。开启后显示原生 Input，根菜单打开时聚焦它；SearchTerm 可绑定持久字符串，nil 时使用 MenuState.Query。FilterFields 默认仅 `label`，还支持 `description`、`id`、`shortcut`；按不区分大小写的包含关系匹配，并保留输入顺序。IgnoreFilter=true 时保留传入的列表，由应用结合查询提供结果。分组标题和分隔符不参与匹配；没有动作匹配时显示本地化空状态，不递归搜索未打开的子菜单。

FilterInput 提供 InputOptions，可配置名称、占位文字、清空按钮和外观。OnSearchTermChange 在用户编辑或清空查询时通知；外部赋值和无输入重建不通知。筛选输入中的 Escape 先清空非空查询，再按层关闭；Enter 激活当前悬停项，否则激活第一个可用结果。上下方向键把实际焦点移到结果，此后文字输入执行菜单前缀导航；这与上游通过 aria-activedescendant 保持输入焦点的实现不同。

子菜单通过 MenuItem.Content.Filter 单独启用筛选，不继承根 Filter。各层按稳定 ID 保存独立查询，也可绑定 Content.SearchTerm。鼠标悬停打开子菜单不会抢走当前筛选输入的焦点；点击或键盘进入子菜单时聚焦其筛选输入。查询使当前焦点项目消失时，焦点返回本层筛选框。

Disabled 关闭菜单；DropdownMenu 同时禁用触发器，ContextMenu 保留目标自身的点击、输入和焦点能力。目标本身已经禁用时，两种菜单都不能打开。OnOpenChange 只报告用户引起的根菜单开关，直接修改 MenuState.Open 不通知；外部关闭时若焦点仍在菜单中，则返回触发器。筛选能力对 DropdownMenu 对应上游参数，对 ContextMenu 属于原生扩展。

### 键盘和指针

| 操作 | 行为 |
| --- | --- |
| ↑↓ / Home / End | 在可用项中移动；Loop 默认 false，在首尾停止，true 时循环 |
| → / ← | 打开子菜单 / 返回父菜单，RTL 下交换 |
| Enter / Space | 激活当前项 |
| Escape | 关闭当前层；根层关闭后恢复锚点焦点 |
| Tab / Shift+Tab | 关闭菜单，继续 MyGo 的正常焦点遍历 |

直接输入字符可做忽略大小写的前缀查找，连续输入间隔为 1 秒，重复首字母可循环匹配。此功能不启用 IME 候选，也不做模糊搜索。

指针移入项目会改变高亮，悬停 180ms 打开子菜单。键盘操作会取消待打开的悬停任务，避免 Escape 后马上重开。项目重排按 ID 保留焦点，删除后清理相关状态。当前没有指针移动轨迹容错或触屏长按打开。

上游 ContextMenu 的 pressOpenDelay 默认 700ms，用于触屏或笔的长按。当前 MyGo InputEvent 不区分鼠标、触屏和笔，Go 未添加会误将鼠标长按当作触屏的延迟字段；应用可提供显式菜单按钮或设置 MenuState.Open。

### 与现有事件组合

DropdownMenu 在锚点上追加 OnClick，原有回调仍会执行；避免再由原有回调切换同一 MenuState.Open。ContextMenu 接管 HandleInput，不会自动组合已有的同类处理器。不同弹层不要共享同一个触发器状态。

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
| Avatar / Icon | 头像优先；默认头像直径 44 DIP，图标 20 DIP，显式头像尺寸沿用 Avatar |
| CloseOptions / CloseIcon | 可选按钮配置和图标，默认 MD / Neutral / Link、无内边距和 X；CloseOptions.Leading 优先 |
| ProgressOptions | 使用 ToastProgressOptions 配置 Color 和 Progress 的 UI；默认 SM 尺寸，颜色继承通知 |
| OnClose / CloseToast | 关闭按钮、到时或 CloseToast 会调用 OnClose；`CloseToast(c, "")` 关闭全部通知 |

单动作和 Actions 可以同时使用。Actions 中每个按钮可设置 ID、Label、OnAction、ButtonOptions 和 Close；默认 XS、Solid，颜色继承通知。动态更新或重排时应提供稳定 ID，非空 ID 在一条通知内必须唯一；省略时仅按静态列表位置识别。更换方向时也会保留当前操作按钮焦点。没有传 Icon 或 Avatar 时，不自动添加状态图标。

Vertical 为默认方向，操作区位于正文下方，间距 10 DIP；Horizontal 将内容与操作垂直居中排列。空标题不留占位行。ToastOptions.UI.Actions 配置操作容器的间距等样式；单个按钮的样式使用 Actions 项中的 ButtonOptions.UI。此前通过 UI.Actions 配置单个兼容 Action 按钮的代码，应改用 Actions 列表和按钮 UI。

OnAction、OnClose 及 Actions 中的 ButtonOptions.Style 会由队列保存，闭包只应捕获业务模型、数据或服务。不要捕获构建产生的 Element、Parts 或 Context；Style 接收到的元素也只在当次构建有效。

### 位置、暂停和样式

ToastViewport 默认右下角、宽 384 DIP，还支持左下、中下、右上、左上、中上；这些是物理位置，不因 RTL 对调。顶部位置将最新通知放在最上方，底部位置放在最下方。宽度进一步受视口实际内边距约束，显式 PaddingX=0 不会保留额外的宽度限制。

Limit 支持 1–3，0 或越界按 3 处理。MyGo 最多提供最新 3 条可见通知，较早的排队项仍计时，较新通知关闭后可能重新出现；Limit 不表示队列容量。Nuxt UI 的 max 默认为 5，超过限制时丢弃较早项，规则不同。

悬停通知区域、焦点在通知内或窗口失焦时，队列计时暂停。默认倒计时条同步暂停，`Progress: nui.Set(false)` 可隐藏；常驻通知不显示倒计时。减少动态效果时仍更新剩余值，按至少 250ms 的间隔直接显示，不插值。进入和退出为朝向对应屏幕边缘的 8 DIP、150ms 淡化位移；减少动态效果时直接切换，空队列不绘制面板。

通知使用 RoleStatus，动作和关闭按钮可用键盘访问；实际读屏播报尚未验证。

每条 Toast 保存创建时的视觉 Theme 和配置副本，清除来源表单的禁用状态及拼接区域信息。显示期间全局主题改变，已有通知仍使用原主题；需要更新同一条通知时，用相同 ID 再次发布。自动生成的关闭名称仍随当前语言变化，显式应用文字由应用更新。Viewport.Decoration 只影响 Root，每条通知通过 ToastOptions.UI 或创建时的 ComponentToast 配置。

同 ID 更新为完整替换，未再次传入的字段回到默认，不调用旧 OnClose，并重新计时。没有上游的重复提示脉冲、折叠堆叠、滑动关闭、任意内容插槽或 foreground / background 播报策略。网站示例和复制程序均显式创建 ToastViewport，避免使用 MyGo 默认通知外观。

## Tabs、Collapsible 与 Accordion

### 选择组件和状态

| 组件 | 绑定状态 | 默认行为 |
| --- | --- | --- |
| Tabs | `*string` 当前项 ID | 胶囊样式，只显示当前内容；可选 Underline、Vertical |
| Collapsible | `*bool` 是否展开 | 点击或 Enter/Space 切换，展开/收起默认 200ms |
| Accordion | `*[]string` 展开项 ID | 默认只展开一项，可再次关闭；Multiple 开启多项 |

TabItem 提供 ID、Label、Icon、Avatar、Badge、BadgeOptions、Disabled、Content 和项目 UI。Icon 优先于 Avatar；默认头像直径随尺寸为 16/16/20/20/24 DIP，显式 Avatar.Size 优先。Badge 使用字符串，`"0"` 会显示；默认 SM / Neutral / Outline，也可由 BadgeOptions 覆盖。Tabs / Accordion 的正文回调接收 Context、Theme，Collapsible 正文接收 Theme 并使用外层 c。

Tabs 按方向使用左右或上下方向键，循环并跳过禁用项；横向左右随 RTL 交换，另一方向的按键不改变选择。当前项被删或禁用后转到首个可用项，没有可用项时保留业务 ID，但不显示正文。当前焦点所在的正文被切换、隐藏或删除后，焦点返回有效标签。正文是可聚焦的原生 group，Tab 可先进入面板，再进入其内部控件；MyGo 暂无独立 tabpanel 角色。Accordion 标题间用 ↑↓、Home、End 移动，Tab 保留标题与正文的正常顺序。

Tabs.OnChange 接收新 ID，Collapsible.OnChange 接收 bool，Accordion.OnChange 无参数，由应用读取 ID 列表。应用直接改值、Tabs 自动修正无效 ID、Accordion 单开模式截断多余 ID，都不会触发用户变更回调。

Accordion 保留应用 open 列表中暂时消失的 ID，便于同 ID 再出现；永久删除时由应用清理。手动 Tabs 激活和 Accordion 的 Collapsible 选项见[参数说明](options.md#滑块和内容切换)。

Collapsible.Disabled、Accordion.Disabled 和 AccordionItem.Disabled 只控制标题，已展开的正文仍使用内部控件自身的禁用设置。外部关闭正在编辑的正文时，焦点返回可用标题；Accordion 删除聚焦项后返回首个可用标题。关闭动画中的正文立即停止接收输入，避免内容收起时继续触发操作。

### 是否保留隐藏内容

默认仅构建活动 Tabs 的正文，折叠内容在退出动画结束后卸载。应用模型保留，但编辑器选区、撤销历史和局部 UI 状态可能在卸载后丢失。

KeepMounted=true 会在每次构建中保留所有正文，通过 Hidden 隐藏非活动内容，因而保留 MyGo 局部状态。隐藏项不参与布局和焦点，但内容回调仍执行。折叠保留模式即时隐藏正文，箭头仍按原生动画进度变化；删除整个条目仍会卸载。

Collapsible 和 Accordion 在打开状态切换 KeepMounted 时沿用相同正文身份，不重建编辑器的局部状态。Theme.Motion.Reduce 或系统减少动态效果会立即完成高度和图标变化；默认动画使用 MyGo 的 200ms 插值，与上游 CSS 动画并非逐帧一致。

TabsOptions.Content 默认 true，显式 false 不构建任何面板，即使 KeepMounted=true 也会卸载局部 UI 状态；应用保存的草稿不受影响。重新开启后从应用数据构建。不要在内容回调中无条件启动 I/O 或提交操作。

### 外观

Tabs 的胶囊活动项使用语义色背景和反色文字，Neutral 使用 InvertedBackground；水平列表填满宽度。Underline 对应 Nuxt 的 link 样式，列表有 1 DIP 基线和活动指示线，横向在底部、纵向在起始侧，随 RTL 调整。纵向标签宽度一致，列表与正文居中对齐。XS–XL 触发器间距为 4/6/6/8/8 DIP，文字行高为 16/16/20/20/24 DIP。

Indicator 可配置活动背景或指示线，Trailing 配置徽标，TabItem.UI 覆盖本项部件。现有 Trigger.Background 覆盖继续有效，Indicator.Background 是更具体的活动背景配置。当前指示器直接随选择更新，没有上游跨标签的 200ms 滑动效果；未提供任意 leading/default/trailing 或 list-leading/list-trailing 插槽，其他操作可放在 Tabs 外部的工具栏中。

pill 的 Indicator 与活动 Trigger 共用原生元素，使背景保持在文字和头像后方；underline 的 Indicator 是独立指示线。Style 回调应按这些实际部件配置，不假定所有样式入口都对应独立节点。

Accordion / Collapsible 的触发器无悬停底色和水平内边距，上下内边距 14 DIP，图标 20 DIP、间距 6 DIP。Accordion 正文使用默认文字色，下边距 14 DIP。

TrailingIcon 默认 ChevronDown。AccordionItem.TrailingIcon 优先于 AccordionOptions.TrailingIcon；AccordionItem.UI 可覆盖本项 Item、Header、Trigger、Leading、Label、Trailing 和 Body。空 Content 不创建正文或底部留白；正文布局与动画裁剪容器不提供两个独立样式入口。

Nuxt Collapsible 使用应用提供的触发器插槽；Go 的 Label、Icon、TrailingIcon、默认标题和正文间距属于原生扩展。需要任意触发器组合时可使用 MyGo Base API。

## Breadcrumb、Pagination、Stepper、Timeline

### Breadcrumb：显示当前位置

BreadcrumbItem 包含 ID、Label、Icon、Avatar、Active、Disabled、OnSelect 和 UI，Options 提供 Label、Color、SeparatorIcon 和 MaxItems。提供 OnSelect 的项目可以操作，包括末项；nil 回调的项目为静态文字。应用负责提供路径和导航动作。

Active 未设置时仅末项激活，显式 false 可以关闭激活样式。Color 默认 Primary，Neutral 的活动文字使用 Highlighted；非活动项为 14 DIP、无内边距的 Muted 文字，可操作项悬停变为 Text 色。分隔箭头 20 DIP、项间距 6 DIP，RTL 下默认箭头反向，自定义 SeparatorIcon 保持原样。Icon 优先于 Avatar，默认头像直径 20 DIP。路径保持单行，长文字按可用空间截断。

MaxItems≥2 时保留第一项及末尾若干项，中间用不可交互的省略号代替，并保留两侧分隔符；省略号不计入显示上限。MaxItems 是原生扩展，没有省略项下拉菜单或路由系统。项目 UI 优先于组件 UI，SeparatorSlot 使用前一项的样式。空 Label 使用本地化的导航组名称。

### Pagination：控制页码

Pagination 绑定 `*int`，页码从 1 开始。Total 是总条数，PageSize≤0 时默认 10。总数减少时会自动把页码限制到有效范围；这类修正不触发 OnChange，只有用户切到不同页时才通知。

组件不查询数据，应用收到变化后自行加载当前页。Siblings≤0 默认 2，最大 10；显式零相邻页使用 SiblingCount。ShowEdges 默认 false，Size 默认 MD，更多外观参数见[分页选项](options.md#pagination-的相邻页和外观)。

翻页控件和页码采用方形按钮，边界按钮禁用，箭头随 RTL 调整。省略号不可交互。首页、上一页、下一页、末页和第 N 页的辅助技术名称随语言切换。

FirstIcon、PrevIcon、NextIcon、LastIcon、EllipsisIcon 可覆盖默认图标，显式图标不随 RTL 自动镜像。ActiveColor / ActiveVariant 默认 Primary / Solid，普通按钮默认 Neutral / Outline。Root / List 共用布局元素；为兼容已有用法，Item 仍覆盖所有按钮，First / Previous / Next / Last 可进一步区分控制按钮，Ellipsis 配置省略区域。页面链接、数据查询和浏览历史由应用在 OnChange 中处理。

### Stepper：选择流程步骤

Stepper 绑定当前步骤 `*string`，StepItem 包含 ID、Title、Description、Icon、Disabled、UI 和 Content。未知 ID 表示尚未开始，不自动选择第一项。Content 同步构建当前步骤正文，切换步骤会卸载旧正文；业务草稿由应用保存，同一 ID 的重排或方向切换保留正文身份。

Linear 是 Optional[bool]，默认 true，允许返回或前进到紧邻的下一步；设为 `nui.Set(false)` 可选择任意可用步骤。CanSelect 能在实际激活时否决操作。方向键只移动焦点，Enter/Space 或点击圆点才激活。组件不会自动跳过线性流程中的禁用步骤，业务流程需由应用安排。

横向默认等宽，圆点居中，文字在下方；Vertical 把文字放在圆点旁，连接线改为纵向。Size 默认 MD，XS 到 XL 的圆点直径为 24/32/40/48/56 DIP，随 Density 缩放。之前的步骤和连接线使用完成色，当前步骤高亮；原图标或数字保留，不自动换成对勾。

步骤正文在横向列表下方、纵向列表旁边。List 配置步骤列表，Panel 配置当前正文；Neutral 的圆点和完成线使用 InvertedBackground。Disabled 只禁用步骤标题，正文控件保留自身的禁用设置。移除或禁用聚焦标题、切换聚焦正文时，将焦点移回可用标题；应用直接改值不触发 OnChange。Nuxt 暴露的 next/prev 和 hasNext/hasPrev 由原生应用通过 ID 与项目列表自行实现。

### Timeline：展示时间线

TimelineItem 包含 ID、Title、Description、Time、Icon、Color、Content，还可提供 Avatar 和 UI。默认纵向，Horizontal 将指示图标及连接线放在文字上方。

Current 是从 0 开始的可选索引，CurrentID 优先；均未设置时没有活动或完成项。Reverse 改变进度解释，不重排数据。点击只调用 OnSelect，不自动改进度。

Time 是应用格式化后的字符串，默认 12/20 DIP 字号与行高，可通过 DateSlot 配置；标题和说明默认 14/20 DIP。Header 配置标记及连接线容器，横向项目等分宽度。Neutral 的活动标记和完成线使用 InvertedBackground。提供 OnSelect 后可通过点击或 Tab、Enter/Space 操作；没有回调时不进入键盘焦点序列。

CurrentID 无匹配时不回退 Current，无效索引不高亮任何项目。组件使用列表语义，不自动排序、换算时区或调度动画；单项 Color、附加 Content 和五档尺寸是明确的原生配置边界。

## Banner 与 ErrorView

### Banner：页面内通知

Banner 放在正常布局中，由应用决定位置。open=false 隐藏；绑定 open 或提供 Close 配置时默认显示关闭按钮，CloseButton=false 可显式隐藏。nil 状态默认常驻，显式启用关闭后使用组件内状态，卸载后重新显示。跨启动的持久化由应用保存，不使用浏览器 localStorage。

Color 默认 Primary，Neutral 背景使用 InvertedBackground，文字使用 Inverted。标题和原生扩展 Description 都保持单行并截断。窗口宽度达到 1024 DIP 后，内容在两侧区域间居中，较窄时从起始边排列；Body 容器按窗口宽度使用 16/24/32 DIP 水平内边距，并受 Theme.ContainerWidth 限制。默认最小高度 48 DIP，字号和图标随文字缩放。

Actions 为同步构建回调，其中 Button 默认 Neutral / XS。OnSelect 接收整条通知的点击或键盘激活，内部按钮独立执行，由应用决定导航动作。Close / CloseIcon 配置关闭按钮，CloseLabel 配置名称；按钮默认 MD / Neutral / Ghost，悬停或聚焦显示 10% 背景色。OnClose 只在用户关闭时通知一次，外部开关不通知。极窄区域需由应用减少操作，Banner 不会自动固定到窗口顶部。

### ErrorView：错误状态页面

Go 中 `nui.Error` 已用于语义颜色，因此错误展示函数名为 ErrorView。ErrorOptions 提供 Status、Title、Description、RetryLabel、Icon、Clear、Retry、OnRetry、Actions 和 Decoration。

只有传入 Icon 时才显示 40 DIP 图标。默认状态码为 16/24 DIP 半粗体 Primary，使用 StatusSlot 配置；标题在窗口小于 640 DIP 时使用 36 DIP 字号，否则为 48 DIP，说明为 18/28 DIP Muted。空标题不创建文字，说明与标题相同时省略。

默认重试按钮为 LG / Primary / Solid，文案为重试；仍需提供 OnRetry 才会创建。Clear 默认 true，false 隐藏默认操作；Retry 配置 ButtonOptions，Actions 替代整个操作区且不受 Clear=false 限制。操作区上方间距 32 DIP，内部默认间距 24 DIP，没有操作时不创建空区域。

嵌入页面局部时可通过 UI[Label] 等降低字号。原生实现保留内边距、480 DIP 说明宽度上限及二级标题语义，不强制占满整个窗口；指定高度后内容居中。组件不调用 Nuxt 的 clearError 或 redirect，panic 处理、请求、导航及错误清理由应用负责。

## 部件索引

| Component 常量后缀 | 可用部件 |
| --- | --- |
| Tooltip | Panel、Label、Trailing |
| Modal / Slideover | Backdrop、Panel、Header、Label、Description、Close、Body、Footer |
| Toast | Root（仅 viewport）、Panel、Body、Leading、Label、Description、Actions、Close；ProgressOptions.UI 使用 Progress 的 Root / Indicator |
| DropdownMenu / ContextMenu | Panel、Header（分组）、Item、Body、Leading、Label、Description、Trailing（行尾容器）、Indicator、SeparatorSlot、Editor（筛选根部件）、EmptySlot；Build 替代 Body 内容 |
| Tabs | Root、List、Trigger、Leading、Label、Trailing、Indicator、Body；TabItem.UI 配置单项部件 |
| Collapsible | Root、Trigger、Leading、Label、Trailing、Body |
| Accordion | Root、Item、Header、Trigger、Leading、Label、Trailing、Body |
| Breadcrumb | Root/List、Item/Trigger、Leading、Label、SeparatorSlot、Ellipsis；斜线分隔的部件共用节点 |
| Pagination | Root/List、Item、First、Previous、Next、Last、Ellipsis、Leading、Label；Root/List 共用节点 |
| Stepper | Root、List、Item、Header、Trigger、Indicator、Leading、SeparatorSlot、Body、Label、Description、Panel |
| Timeline | Root、Item、Header、Indicator、SeparatorSlot、Body、DateSlot、Label、Description |
| Banner | Root、Body、Header、Leading、Label、Description、Actions、Trailing、Close |
| Error | Root、Leading、StatusSlot、Label、Description、Actions |

Stepper 的 Header 是圆点和连线容器，Trigger 是圆形按钮，Leading 为图标，SeparatorSlot 为连接线，Body 为文字区域。

Tabs、Stepper、Timeline 接受全局和实例 Color / Size；Banner、Toast 接受 Color。菜单项和 Timeline 条目的 Color 是各条目自己的设置，其他样式通过已列出的部件配置。

## 示例

Gallery 的 **Overlays & menus** 和 **Navigation & disclosure** 展示弹层、菜单和导航。Settings 示例包含编辑、校验、重置确认、Toast 撤销和通知筛选。示例状态保存在窗口模型中，长期存储仍由应用实现。

平台和输入法限制见[兼容性说明](compatibility.md)。
