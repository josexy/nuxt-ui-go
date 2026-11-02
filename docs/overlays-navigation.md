# 弹层、菜单与导航

使用持久状态控制弹层与选中项，使用稳定 ID 连接条目与业务数据。完整平台支持和已知问题见 [兼容性说明](compatibility.md)。

## 公共约定

Tooltip 受控状态、Toast 多动作、Tabs 激活方式、Accordion 折叠规则与分页选项见 [参数与默认值](options.md)。

导入 `nui "github.com/josexy/nuxt-ui-go"`。入口沿用 `(c *ui.Context, t nui.Theme, ...) ui.Element`；Toast 返回通知 ID，CloseToast 不返回元素。每个实例在构建前使用 `c.Key(...)`，每帧重建，应用保存数据与 `ui.Handle`。构建得到的 Element、Parts、Context 不存入模型或后台任务。返回的零 Element 表示本帧未显示。可选弹层通过不可见宿主消费实例 key，避免显隐时把 key 泄漏给后续控件；弹层节点采用稳定身份，其他通知卸载不应丢失编辑状态。

`Decoration{UI, Style}` 继续使用内置配方 → Theme 组件配置 → 实例 UI → 同步 Style 回调的顺序。Style 只在当前构建期设置部件，不负责持久状态。组件通过 `Component...` 常量配置默认样式，并提供 `Backdrop/Trigger/List/Close/Actions` slots；不是每个组件都支持每个 slot 或 Color/Size/Variant。

所有条目 ID 必须非空且唯一，重复会 panic。Tabs、Accordion、Breadcrumb、Stepper、Timeline 在各自列表中唯一；菜单在整棵树中唯一。标签可重复，不能用标签代替 ID。必需的状态指针传 nil 会 panic；Banner 的 nil 是明确例外。

## Modal 与 Slideover

```go
// 每帧调用，包括 open 为 false 时；state 为窗口独立的业务模型。
edit := nui.Button(c, t, "编辑", nui.ButtonOptions{}).
    Bind(&state.EditTrigger).
    OnClick(func() { state.EditOpen = true })
nui.Tooltip(c, t, edit, "编辑当前资料", nui.TooltipOptions{})
nui.ModalWithOptions(c.Key("edit"), t, &state.EditOpen, nui.ModalOptions{
    Title: "编辑资料",
    Description: "更改只在保存后应用。",
    ReturnFocus: &state.EditTrigger,
    Footer: func(c *ui.Context, t nui.Theme) {
        nui.Button(c, t, "完成", nui.ButtonOptions{}).
            OnClick(func() { state.EditOpen = false })
    },
}, func(t nui.Theme) {
    nui.Input(c, t, &state.Draft, nui.InputOptions{Label: "名称"})
})
```

`Modal(c,t,&open,title,func(Theme))` 是便捷入口；需要完整配置时使用 `ModalWithOptions`。

| ModalOptions | 行为 |
| --- | --- |
| Title、Description | dialog 名称与描述，默认标题区同步显示 |
| Header、Footer `func(*ui.Context, Theme)` | 自定义标题内容和操作区；正文使用独立 `func(Theme)` 参数 |
| Width、MaxHeight `Optional[float32]` | 默认宽 512 DIP；限制在窗口宽减 32、高减 64 内；正文滚动，标题和操作区不参与正文滚动 |
| Dismissible `Optional[bool]` | 默认 true；false 拦截遮罩点击与 Escape |
| CloseButton `Optional[bool]`、CloseLabel | 默认有关闭按钮；与 Dismissible 独立。完全禁止用户关闭需同时设为 false |
| OnClose `func()` | 组件关闭按钮、允许的遮罩/Escape 关闭时调用一次；应用直接写 open=false 不触发 |
| ReturnFocus `*ui.Handle` | 可选显式恢复目标。尤其适合“菜单项打开对话框”，菜单项关闭后已卸载，应回到仍存在的菜单按钮 |

Modal 使用 DialogBase 的模态作用域，背景不可操作，Tab 约束在内容中；Select/Popover 等子层先消费 Escape。默认焦点进入第一个可聚焦控件，通常为关闭按钮。需要首字段焦点时，在应用打开动作中请求该字段 Handle。自定义 Header/Footer 过大时，应用仍需控制内容尺寸，不能用无限高度操作区挤占正文。

ModalOptions 提供 Overlay（默认 true）与 Fullscreen。Overlay=false 仅隐藏遮罩颜色，保留模态焦点；Fullscreen 使用整窗尺寸、零圆角和零边框。默认遮罩为 Elevated 的 75% 透明度配色。Slideover 不使用 Fullscreen，仍由 Side/Extent 控制。

`Slideover(c,t,&open,SlideoverOptions,body)` 复用相同标题、正文、关闭与 Footer 选项。`Side` 为 `SideRight`（零值）、`SideLeft`、`SideTop`、`SideBottom`；`Extent` 默认 448 DIP，表示左右宽度或上下高度，最多占满对应窗口轴。Slideover 使用 Extent 和窗口尺寸；嵌入 ModalOptions 的 Width/MaxHeight 仅供 Modal 使用。

Slideover 采用 200ms cubic ease-out 位移与遮罩过渡；`Theme.Motion.Reduce` 时即时开关，当前持续时间固定。退出动画结束前保留模态作用域、禁止内容操作，结束后卸载并恢复焦点；反向切换从当前进度继续。请持续调用组件，不能用 `if open` 包住整次构建。正文伸展并滚动，Footer 位于面板底部。需要底部手柄拖拽与停靠点时使用 [Drawer](advanced.md#drawer)。Slideover 不提供拉伸边缘。

## Tooltip

`Tooltip(c,t,anchor,text,TooltipOptions)`：无交互提示，设置 anchor.Description，使用 TooltipBase 的悬停与键盘焦点规则。默认下方居中、边距 8 DIP，MaxWidth 默认 320；Disabled 或无效/禁用锚点不显示。

当前 MyGo 的鼠标延迟为 600ms，连续提示有 400ms warm window；键盘焦点即时显示。离开、按下或 Escape 隐藏。Options 未提供延迟时长配置。提示不接收点击或额外焦点；需要交互内容使用 Popover。Description 会替换锚点已有描述，组合时由调用者整理所需文本。

## DropdownMenu 与 ContextMenu

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

`MenuState` 每窗口、每独立触发器一份，可由应用读写 Open；内部只保存数据和 Handle。`DropdownMenu` 接收可点击锚点，设置 menuButton/expanded；`ContextMenu` 接收普通可聚焦区域，支持右键点位、Menu 键与 Shift+F10（键盘方式放在锚点下方）。两者返回当前根面板，关闭时返回零值。

| MenuItem | 行为 |
| --- | --- |
| ID、Label、Description、Icon | 稳定身份、语义与显示；图标使用 Lucide |
| Kind | MenuAction（零值）、MenuLabel、MenuSeparator；分组标题和分隔符也要求 ID |
| Shortcut | 右侧显示字符串；不自动注册全局快捷键 |
| Disabled、Color | 禁用阻止聚焦/激活；语义色可表达危险动作 |
| Checked `Optional[bool]`、Radio | 设置后分别使用 menuItemCheckBox/menuItemRadio；应用在 OnSelect 更新值及同组互斥关系 |
| Children | 嵌套条目；键盘和鼠标打开；递归超过 32 层拒绝构建 |
| OnSelect `func()` | 叶项激活后先关闭并请求恢复触发器，再调用一次；子菜单父项不执行叶项动作 |
| Build `func(*ui.Context, Theme)` | 自定义展示内容。Label 仍必需；不在其中嵌套输入框或独立按钮 |

MenuOptions 提供 Label、Width（默认 240）、MaxHeight（默认 320）与 Decoration。面板宽高限制在窗口内，长菜单滚动，子菜单锚定父行并由公开 AttachTo 避让窗口边缘。非模态菜单外部点击关闭后继续传到目标。

键盘：上下/Home/End 在可用项间移动；右箭头打开子菜单，左箭头退一级（RTL 时左右互换）；Enter/Space 激活；Escape 只退当前层，根层关闭并恢复锚点；Tab/Shift+Tab 关闭菜单并交给 MyGo 正常遍历。直接文本输入进行不区分大小写的 1 秒前缀匹配，重复首字母可循环；它不启用系统 IME 候选输入，也不是模糊搜索。

指针移入改变高亮，子菜单悬停 180ms 打开；键盘动作清除悬停待打开状态，避免 Escape 后原地重开。重排按 ID 保留焦点，删除条目清理 Handle 与无效子路径。不支持指针轨迹三角容错或触屏长按。

DropdownMenu 接管锚点 OnClick；ContextMenu 接管锚点 HandleInput，不能假定与原有同类处理器链式组合。菜单不能与另一个独立弹层共享同一触发器状态。窗口失焦时 MyGo 不显示焦点且不执行 focus-group 方向键；焦点不会在每次后台重建时被重新请求到第一项。

## Toast

```go
t := nui.App(c, config)
// 每个窗口一次。放在其他弹层构建前，让模态遮罩盖住通知，避免遮挡面板操作。
nui.ToastViewport(c, t, nui.ToastViewportOptions{})
nui.Button(c, t, "保存", nui.ButtonOptions{}).OnClick(func() {
    previous := state.Value
    state.Value = state.Draft
    nui.Toast(c, t, nui.ToastOptions{
        ID: "saved", Title: "已保存", Color: nui.Success,
        Action: "撤销", OnAction: func() { state.Value = previous },
    })
})
```

Toast 从动作回调加入窗口内置队列，不能无条件在视图构建中调用。ToastOptions 包含 ID、Title、Description、Action、Actions、Color、Icon、Timeout、OnAction、OnClose、Orientation、Close、Progress 和 UI。空 ID 自动生成；同 ID 替换并重置时长，不执行被替换旧项的 OnClose。默认超时 5 秒，负值常驻。关闭按钮、到时和 CloseToast 会调用 OnClose；`CloseToast(c, "")` 关闭所有项。

Action / OnAction 定义一个动作按钮，执行后关闭通知并调用 OnClose。Actions 可添加多个 `ToastActionOptions`，各自配置 Label、OnAction、ButtonOptions 和 Close；默认执行后关闭，`Close: nui.Set(false)` 可保留通知。两种动作入口可同时使用。`ToastOptions.Close: nui.Set(false)` 隐藏通知关闭按钮，不影响超时或动作的关闭规则。

OnAction/OnClose 被队列保存，必须只捕获业务模型、数据或服务，不能捕获构建 Element、Parts、Context。Toast 本身应从仍有效的 UI 动作调用；后台完成通知通过 Window.Update 发布结果，再由窗口 UI 动作/受控事件桥接。

ToastViewportOptions：Position（右下默认，另有左下/中下/右上/左上/中上）、Width（默认 356）、Limit、Decoration。Limit 1–3，0 或越界归为 3；固定 MyGo 最多提供最新 3 个，较早排队项仍计时。Limit 不是无限队列容量。无队列不绘制面板，进入/退出 150ms（减少动态效果时无过渡）。有超时的通知默认绘制倒计时条，`ToastOptions.Progress: nui.Set(false)` 可关闭；悬停/焦点暂停时同步停止推进。

队列在悬停整个 viewport 内的通知、焦点位于通知内或窗口失焦时暂停，由 MyGo 管理计时和窗口清理。通知为 RoleStatus，操作和关闭按钮可通过键盘访问。屏幕阅读器实际播报尚未验证。

**主题策略**：Toast 保存创建时 Theme 值和 UI 副本，保持局部主题；显示期间切换全局模式不会改写已有通知，后续通知使用新主题。需要即时重绘同一通知时用同 ID 再发布。Viewport 的 Decoration 只作用于 Root；每条通知用 ToastOptions.UI 或创建时 Theme 的 ComponentToast 配置。Actions 中的 ButtonOptions.Style 会随通知保存，闭包只捕获模型或服务；传给 Style 的 Element 仅在当次构建有效。

## Tabs、Collapsible 与 Accordion

| 入口与状态 | Options / 条目 | 行为 |
| --- | --- | --- |
| `Tabs(c,t,*string,[]TabItem,TabsOptions)` | TabItem：ID、Label、Icon、Disabled、Content；Options：Label、Color、Size、Underline、Vertical、KeepMounted、OnChange | 默认 pill，支持 underline/纵向；ID 映射为 Base 内部索引，方向键跳过禁用项；选中项删除/禁用后归到首个可用项，无可用项时保留业务 ID 且不显示激活正文 |
| `Collapsible(c,t,*bool,CollapsibleOptions,func(Theme))` | Label、Icon、Disabled、KeepMounted、OnChange | Enter/Space 或点击切换，expanded 语义和旋转箭头；默认 Base 200ms 展开/收起 |
| `Accordion(c,t,*[]string,[]AccordionItem,AccordionOptions)` | Item：ID、Label、Icon、Disabled、Content；Options：Multiple、Disabled、KeepMounted、OnChange | 默认单开，可再次关闭；Multiple 可多开；标题之间上下/Home/End 跳过禁用项，Tab 保持标题与正文正常顺序 |

正文 builder 使用 `(c,Theme)`，Collapsible 的正文仅接收 Theme 并沿用当前 c。Tabs OnChange 接收新 ID，Collapsible 接收 bool，Accordion 无参数、由调用方读取 ID 列表。用户变更触发回调；应用直接赋值、Tabs 归一化无效选中值、Accordion 单开模式截断多余 ID 不触发用户变更回调。Accordion 不自动清理业务 open 中已删除的 ID，便于同 ID 重现；永久删除时应用自行清理。

默认只构建活动 Tabs 正文；默认折叠正文在退出过渡完成后卸载。应用数据始终由模型持有，卸载不承诺保留编辑器选区、撤销栈或局部 UI 状态。`KeepMounted=true` 每帧构建所有正文，以 Hidden 隐藏；保留 MyGo 局部状态，隐藏项不参与布局/焦点，代价是继续执行 builder。折叠保留模式的正文即时隐藏，箭头仍遵循 Base 进度。删除整个条目仍会卸载，不等于缓存已删条目。

Tabs 的 pill 活动项使用语义色背景/反色文字，水平列表填充宽度；underline 为 1 DIP，纵向位置随 RTL 镜像。Banner 使用对应语义色的 solid 表面。Toast 未传 Icon 时不自动添加状态图标；Actions 中的按钮默认 Neutral/Outline/SM。

## Breadcrumb、Pagination、Stepper、Timeline

- **Breadcrumb**：`[]BreadcrumbItem{ID,Label,Icon,Disabled,OnSelect}`，Options 为 Label、MaxItems。最后一项作为当前页文本，不可点击，同样显示 Icon；祖先为 Link 语义。分隔符为 chevron-right，右到左时换成 chevron-left。MaxItems≥2 时保留第一项和末尾若干项，中间显示不可交互省略号；不实现省略项下拉或路由系统。当前路径需由应用提供。
- **Pagination**：绑定 `*int`，页码从 1 开始。Total 为总条数、PageSize≤0 默认 10、Siblings≤0 默认 2（最大 10），`SiblingCount: nui.Set(0)` 可显式指定零邻页。ShowEdges 默认 false，Size 默认 MD；总数减少会静默夹紧 page。边界按钮禁用；OnChange 仅在用户切到不同页时执行，不执行查询。按钮与可访问名称跟随 MyGo 全局语言，支持中英文，见 [i18n](i18n.md)。
- **Stepper**：绑定当前 `*string`，`[]StepItem{ID,Title,Description,Icon,Disabled}`；Options：Label、Vertical、Linear、Disabled、Color、Size、CanSelect、OnChange。横向默认等宽排列，圆点居中，标题与描述在下方，圆点之间有连接线；Vertical 将文字放在圆点旁，并改为纵向连线。Size 默认 MD，XS/SM/MD/LG/XL 的圆点直径为 24/32/40/48/56 DIP（随 Density 缩放）。当前位置以前的圆点及其后连接线显示完成色，当前位置圆点高亮；图标或数字保留，不自动替换成勾号。点击圆点激活步骤。默认允许任意可用步骤；Linear 允许返回和前进至下一个位置，CanSelect 在实际激活时可否决。方向键只移动焦点，Enter/Space 才激活；不自动跳过线性流程中的禁用业务步骤，应用应提供正确流程。未知 ID 表示尚未开始，不静默选中第一项。
- **Timeline**：`[]TimelineItem{ID,Title,Description,Time,Icon,Color,Content}`；Options：Label、Horizontal、Color、Size、Current、CurrentID、Reverse、OnSelect。默认纵向；横向将指标/连接线放在上方，文字放在下方。Current 为可选零基步骤索引，CurrentID 优先；未设置时无活动/完成项；Reverse 改变进度解释而不重排。点击只通知 OnSelect，不改受控进度。条目可带 Avatar/UI。Time 为应用格式化字符串。RoleList/ListItem，不自动排序、格式化时区或调度动画。

## Banner 与 ErrorView

`Banner(c,t,open *bool,BannerOptions)` 是容器内状态通知。open=false 不显示；nil 表示常驻且无关闭按钮；非 nil 显示可关闭按钮，悬停或键盘聚焦时铺 10% 的背景色（Nuxt 的 `bg-default/10`），反色图标保持可读。Options：Title、Description、CloseLabel、Color、Icon、Actions builder、OnClose、Decoration。应用安排其位置，不自动固定在窗口顶部。长文本换行，操作区支持多按钮；极窄容器应由应用减少操作或改为纵向组合。

Nuxt 的 Error 入口在 Go 命名为 `ErrorView`，保留既有语义色 `nui.Error`。`ErrorOptions` 提供 Status、Title、Description、RetryLabel、Icon、OnRetry、Actions、Decoration。默认警告图标与居中标题，Actions 优先于默认重试按钮；没有 OnRetry 时不生成重试按钮。不接管 panic、请求、路由或重试状态机。

## 部件索引

| Component 常量后缀 | 可用 slots |
| --- | --- |
| Tooltip | Panel、Label |
| Modal / Slideover | Backdrop、Panel、Header、Label、Description、Close、Body、Footer |
| Toast | Root（仅 viewport）、Panel、Leading、Label、Description、Actions、Close |
| DropdownMenu / ContextMenu | Panel、Header（分组）、Item、Leading、Label、Description、Trailing；自定义 Build 替代 Label/Description 内容 |
| Tabs | Root、List、Trigger、Leading、Label、Body |
| Collapsible | Root、Trigger、Leading、Label、Trailing、Body |
| Accordion | Root、Item、Trigger、Leading、Label、Trailing、Body |
| Breadcrumb / Pagination | Root、Item |
| Stepper | Root、Item、Header（圆点与连线容器）、Trigger（圆形按钮）、Indicator、Leading（图标）、SeparatorSlot（连接线）、Body（文字容器）、Label、Description |
| Timeline | Root、Item、Indicator、SeparatorSlot、Body、Label、Description |
| Banner | Root、Leading、Label、Description、Actions、Close |
| Error | Root、Leading、Label、Description、Actions |

Tabs、Stepper、Timeline 接受全局/实例 Color、Size；Banner、Toast 接受 Color；其他组件主要通过 UI slots 定制。菜单每个 Item.Color、Timeline 每个 Item.Color 是条目级配置，不借用其他组件默认语义色。几何选项使用 DIP，不按物理像素缩放二次计算。

## 示例

Gallery 的 Overlays & menus、Navigation & disclosure 页面展示菜单、弹层与导航。Settings 示例包含编辑、校验、重置确认、Toast 撤销和通知筛选。示例保存到窗口模型，业务持久化由应用实现。
