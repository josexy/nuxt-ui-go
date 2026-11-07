# 参数与默认值

本页用于查询常用选项。第一次使用某个组件时，先看[组件指南](README.md)中的示例，再回来查默认值和行为开关。

| 需要调整什么 | 阅读位置 |
| --- | --- |
| 显式关闭默认开启的功能，或设置 0 | [未设置、零值和默认值](#未设置零值和默认值) |
| 横向和纵向布局 | [Orientation](#orientation) |
| 多选、搜索、查询词重置和分页 | [选择与分页](#选择与分页)、[InputMenu](#inputmenu-值与搜索策略) |
| 加载、标签、弹层、侧栏或上传区域 | [常用参数入口](#常用参数入口) |
| 键盘和辅助技术限制 | [使用限制](#使用限制) |

## 未设置、零值和默认值

Go 的零值不一定是组件最终使用的值。例如 Pagination 的 `PageSize: 0` 会采用默认值 10。需要区分未设置与显式 `false` / `0` 的字段，使用 `Optional[T]`。

| 写法 | 含义 |
| --- | --- |
| 不填 Optional 字段，或填 `nui.Optional[T]{}` | 使用组件默认行为 |
| `nui.Set(false)` | 明确关闭该选项 |
| `nui.Set(0)` | 明确使用 0；仍需满足该字段的取值限制 |
| 普通 bool、数值、枚举字段 | 直接赋值；具体零值行为看组件说明 |

例如隐藏 Toast 关闭按钮，用 `Close: nui.Set(false)`；将 Pagination 的相邻页数量设为零，用 `SiblingCount: nui.Set(0)`。仅写 `Siblings: 0` 会使用默认值 2。

### 常查的默认值

| 组件 | 默认值 |
| --- | --- |
| Pagination | PageSize=10、Siblings=2、ShowEdges=false、ShowControls=true、Size=MD |
| InputTags / InputRating / Slider | AddOnPaste=false；Hoverable=false；Step=1 |
| SelectMenu | SearchInput=true；选择后和关闭后均重置搜索词 |
| Carousel / FileUpload | Arrows=false；Layout=UploadGrid |
| Sidebar | Resizable=false；Collapsible、Toggle、AutoClose=true；展开/收起宽度 256/64 DIP；Description 为空，Actions 仅展开或移动菜单显示 |
| Header / Footer / FooterColumns | Header Title 默认为空；Footer 无顶部边线；FooterColumns 首选最小列宽 160 DIP、列间距 32 DIP |
| 弹层与通知 | Popover.Dismissible=true；Toast.Close、Toast.Progress=true |
| Listbox / Accordion | HighlightOnHover=true；Collapsible=true |
| Calendar | Primary / Solid / MD；FixedWeeks、MonthControls、YearControls、ViewControl=true；NumberOfMonths 默认 1，YearsPerPage 默认 12 |
| ColorPicker | Size=MD，默认 SV 边长 168 DIP；Throttle 未设置为 50 毫秒，显式 0 不节流；模型即时更新 |
| InputDate / InputDateRange | ShowCalendar=false；HideCalendar 优先关闭；DateStep 各项 0 使用 1；范围分隔默认为减号 |
| InputTime | TwelveHour 未设置时使用区域小时制；Granularity 未设置时由 Seconds 决定，默认分钟；TimeStep 各项 0 使用 1；StepSnapping=false |

FileUpload 的 `Dropzone`、`Interactive`、`FileDelete`、`Preview` 默认均为 true。`UploadLayout` 的零值是 `UploadLayoutDefault`，填写选项时使用枚举名称，不依赖底层数字。

## Orientation

横向使用 `nui.Horizontal`，纵向使用 `nui.Vertical`。默认纵向的组件通常采用 `Optional[Orientation]`，需要写成 `nui.Set(nui.Horizontal)`；普通 Orientation 字段直接赋值。

| Options | 字段类型与默认值 | 改变的内容 |
| --- | --- | --- |
| FormFieldOptions | Optional；Vertical | 标签和输入内容上下排列或左右分列，保留标签关联和编辑器状态 |
| AlertOptions / ToastOptions | Optional；Vertical | 操作区在正文下方或旁边 |
| ListboxOptions | Optional；Vertical | 纵向虚拟列表或固定 ItemWidth 的横向虚拟列表；使用对应轴的方向键 |
| InputNumberOptions | Orientation；Horizontal | 减号/数值/加号横排，或数值右侧上下排列增减按钮 |
| InputRatingOptions / RangeSliderOptions | Orientation；Horizontal | 星级或滑块的排列及交互方向；范围滑块保留两个独立焦点和值 |
| ProgressGroupOptions | Orientation；Horizontal | 水平或纵向分段轨道；纵向 Height 默认 160 DIP |
| CarouselOptions | Orientation；Horizontal | 布局、拖拽和方向键同步切换；切换时取消未完成的拖拽 |

```go
nui.FormField(c, t, nui.FormFieldOptions{
    Label: "工作区",
    Orientation: nui.Set(nui.Horizontal),
}, func(inner nui.Theme) {
    nui.Input(c, inner, &state.Name, nui.InputOptions{})
})

nui.RangeSlider(c, t, &state.Low, &state.High, nui.RangeSliderOptions{
    Label: "预算",
    Orientation: nui.Vertical,
    Min: 0, Max: 100, Step: 5, MinStepsBetweenThumbs: 2,
})
```

部分组件另有 `Vertical` 或 `Horizontal` 布尔字段，类型以对应 Options 为准。支持 RTL 的横向交互会按书写方向调整。

## 选择与分页

### 单选、多选和再次点击

| 组件或选项 | 行为 |
| --- | --- |
| `Select(*T)` / `SelectMultiple(*[]T)` | 分别绑定单个值和多个值，共用 SelectOptions、稳定 ID、自定义 Option.Build 和已选标签 |
| `SelectionToggle` | Listbox 默认策略；单选和多选都可再次激活当前项以取消选择 |
| `SelectionReplace` | 激活时替换选择；SelectMenu 单选固定使用此策略 |
| 替换式多选 | Shift+方向键/Home/End 扩展连续范围并跳过禁用项；Ctrl/Cmd+A 选择可用项 |
| `CheckboxGroup.Max` | 达到上限后，未选项仍可聚焦，但拒绝勾选，并向辅助技术报告只读状态 |

Select 单选和多选都继承 ComponentSelect 的 Color、Variant、Size 和 UI，实例配置优先。Listbox 还支持 Loop 和 HighlightOnHover。

### SelectMenu 的搜索与重置

`SearchInput` 默认 true。设为 false 后隐藏搜索框，保留的 Query 仍参与筛选；ExternalFilter 或 IgnoreFilter 跳过本地筛选。触发器上的 ↑、↓ 均可打开菜单：默认聚焦可用搜索框，FocusSearch=false 时聚焦列表；禁用触发器不能打开。

默认在选择后和关闭后清空 Query，分别由 `ResetSearchTermOnSelect`、`ResetSearchTermOnBlur` 控制。关闭包括 Escape、外部点击和应用修改打开状态。

再次选择同一个值也会关闭、重置并触发 OnSelect；OnChange 只在选中值实际变化时触发。Invalid / ErrorMessage 会同步影响字段错误语义和边框。

Clearable 显示清空按钮，ClearOptions / ClearIcon 配置按钮。按钮始终清除查询，ResetModelValueOnClear 默认 true，设为 false 时保留选中值。OnClear 在重置后通知一次，OnOpenChange 只通知用户操作造成的实际开关变化。

Listbox 使用 Filter / FilterInput 配置内置筛选，SelectMenu 使用 SearchInput / FilterInput。两者都支持 SearchTerm 指针、FilterFields、IgnoreFilter 和 OnSearchTermChange。SearchTerm 未设置时使用各自 State.Query；外部赋值不发出搜索通知。

### Pagination 的相邻页和外观

`SiblingCount` 优先于旧的 `Siblings`，范围为 0–10。例如当前第 5 页、共 10 页、相邻数为 0 时：

| 配置 | 显示结果 |
| --- | --- |
| ShowEdges=true | `1 … 5 … 10` |
| ShowEdges=false | `5` |

接近首尾时会扩大边缘页码范围；总页数较少时不显示省略号。ShowControls 控制翻页按钮，ShowEdges 控制首尾页码。Size、Color、Variant 设置普通页码外观，ActiveColor、ActiveVariant 单独设置当前页。

## 常用参数入口

Table 的 Sortable、ManualSorting 和 List.SortBy 分别控制列是否可排序、由应用排序、多列排序。Loading 保留现有行并显示加载线，BuildLoading 只在空表加载时替换空内容；Caption 与 BuildCaption 分别提供描述和可见标题。排序、选择和激活通知以及原生限制见 [Table 文档](forms-data.md#table)。

Tree 的 OnSelect / OnToggle 可以取消操作，OnChange / OnExpandedChange 报告实际选择和展开变化。TreeNode.UI / Style 控制单项部件，BuildLeading / BuildTrailing 替换非交互首尾内容。部分选中状态与多选联动、懒加载和空分支的规则见 [Tree 文档](forms-data.md#tree)。

### 开关、标签和评分

| 组件 | 常用选项与行为 |
| --- | --- |
| Switch | Loading 阻止鼠标和键盘切换；LoadingIcon、CheckedIcon、UncheckedIcon 配置图标；减少动态效果时停止旋转 |
| Checkbox / CheckboxGroup / RadioGroup | Variant 可选 ChoiceList、ChoiceCard、ChoiceTable；Indicator 可选 IndicatorStart、IndicatorEnd、IndicatorHidden；Highlight 强调边框或指示器，未选中时也生效 |
| InputTags | AddOnTab、AddOnBlur、AddOnPaste 控制添加时机；Duplicate、ConvertValue、DisplayValue 控制值与显示；MaxLength 限制编辑文本 |
| PinInput | Color、Variant、Size 设置外观；`Separator: []int{3}` 在第三格后分组，仍使用一个编辑器和一个 Tab 停靠点 |
| InputRating | Clearable 允许点击当前评分或按空格清零；Hoverable 默认 false；EmptyIcon 未设置时沿用 Icon，Indicator / IconSlot / EmptyIconSlot 分别配置填充层及两种图标 |

Checkbox 只使用 CheckedIcon，不使用 Switch 的 UncheckedIcon。ChoiceCard 选中后会着色，隐藏指示器时仍能识别选中状态。

Checkbox 的 Loading、LoadingIcon 和 UncheckedIcon，以及 RadioGroup 的 Max，来自共享选项类型，对相应组件不生效；保留这些字段以兼容已有代码。单个 Checkbox 的 ChoiceTable、Switch 的 Variant / Indicator 是 Go 原生扩展。网站参数面板和 API 表会分别说明适用范围。

CheckboxGroup / RadioGroup 的 Option.Icon 已生效；隐藏指示器时显示于项目标签上方，普通 Radio 指示器仍为圆点。Loop 控制方向键循环，多选组默认 false，单选组默认 true；Set(false) 明确关闭循环。两类组均使用一个 Tab 停靠点，横纵布局分别使用左右、上下方向键。

InputTags 的 MaxLength 按 UTF-16 单元限制用户插入，0 表示不限；它限制的是编辑草稿，不负责提交时的标签长度校验。启用 AddOnPaste 后，粘贴会按标签处理，即使因重复或数量上限被拒绝，也不会再作为普通文本插入；关闭时按普通文本粘贴。启用 AddOnTab 后，非空草稿会在 Tab / Shift+Tab 时提交并保留焦点，空草稿则正常移动焦点。

InputNumber 的 StepSnapping、FocusOnChange、Increment 和 Decrement 默认 true，滚轮默认开启。空值首次增减使用 StartingValue，不会先额外加减一步。PinInput.Highlight 默认 false；SeparatorSlot 和 Root.Gap / Padding 可调整分隔符及间距。完整行为见[数值、标签、PIN](forms-data.md#数值标签pin评分)。

InputRating 的横向方向键随 RTL 调整，纵向使用 ↑↓，另一轴按键无效。评分到首尾停止；Home/PageUp 到 1，End/PageDown 到 Max。清零后悬停预览随即消失。

### 滑块和内容切换

| 组件 | 常用选项与行为 |
| --- | --- |
| Slider / RangeSlider | PageUp/PageDown、Shift+方向键移动 10 个 Step；OnChange 报告实时变化，OnCommit 在有变化的按键或拖拽结束时通知 |
| RangeSlider | 两个滑块可交叉，随后自动排序并转移活动焦点；小于 `Step × MinStepsBetweenThumbs` 的移动被拒绝；Tooltip 分别显示两个值 |
| Tabs | ActivationMode 默认 ActivationAutomatic；ActivationManual 下方向键只移动焦点，Enter/Space 才激活并通知一次 |
| Accordion | Collapsible 默认 true；单开模式设为 false 后不能再次点击收起当前项，仍可切到另一项；多开模式不受此选项限制 |
| Collapsible / Accordion | Disabled 只禁用标题；KeepMounted 默认 false，true 保留隐藏控件。TrailingIcon 默认 ChevronDown，Accordion 单项图标优先；AccordionItem.UI 可覆盖本项样式 |
| Carousel | Arrows 控制翻页按钮，Dots 开启圆点导航；圆点的页码名称支持语言切换 |

### 弹层和通知

| 组件 | 常用选项与行为 |
| --- | --- |
| Popover | Mode 默认 PopoverClick，可改为 PopoverHover；OpenDelay / CloseDelay 只用于悬停模式，类型为 time.Duration，默认 0 |
| Tooltip | Placement、Offset 控制位置；Open 可绑定打开状态；Kbds 显示并补充锚点说明；DelayDuration 可显式设置悬停延迟，0 表示立即显示 |
| Modal / Slideover | CloseButton 控制关闭按钮是否显示，Close / CloseIcon 配置按钮；Transition=false 跳过过渡，OnClosePrevent 通知被 Dismissible=false 阻止的关闭尝试 |
| Slideover | Inset 增加边缘间距和圆角；Extent 只设置对应轴尺寸，Width / MaxHeight / Fullscreen 不参与布局 |
| Toast | Close / Progress 控制是否显示；CloseOptions / CloseIcon、Avatar、ProgressOptions 配置外观；Actions 的稳定 ID 保留重排焦点，单项 Close=false 可保留通知 |
| DropdownMenu / ContextMenu | Size 默认 MD，Loop 默认 false；Filter / SearchTerm 配置筛选，MenuItem.Content 配置子级；Close=false 保留叶项选择后的菜单，OnCheckedChange 和 OnOpenChange 分别通知值与开关变化 |
| Tabs | Content 默认 true；false 卸载全部面板，KeepMounted 只保留未激活面板。TabItem 支持 Avatar、Badge / BadgeOptions 和 UI，Indicator 配置当前活动背景或指示线 |

Popover 的 Dismissible 默认 true，控制 Escape 和外部点击关闭，应用始终可以直接修改 open。悬停模式允许指针跨过锚点与面板之间的间隙；Escape 关闭后，指针停在原处不会立即重新打开。网站参数面板中的延迟以毫秒填写。

Tooltip 即使绑定了 Open，也继续响应悬停延迟、键盘焦点、失焦、点击和 Escape。应用将 Open 改为 false 后，不会因为仍悬停或聚焦而立即重开；新提示出现时会关闭旧的受控提示。

Toast 动作默认执行后关闭通知，单项设置 `Close: nui.Set(false)` 可保留。悬停或聚焦使通知暂停时，倒计时条也停止；常驻通知没有倒计时条。

### 侧栏、导航和文件选择

| 组件 | 常用选项与行为 |
| --- | --- |
| Sidebar | State 保存展开宽度和窄窗口打开状态；MinWidth/MaxWidth 默认 160/480 DIP；Breakpoint 未设置或为 0 时始终使用桌面布局 |
| NavigationMenu | 纵向 Compact 模式默认提供 Tooltip 和子级 Popover，可分别用 `nui.Set(false)` 关闭；PopupPlacement 修改子级弹层位置 |
| Breadcrumb | Color、SeparatorIcon 配置外观；项目 Avatar、Active、UI 可单独设置；提供 OnSelect 的末项也可操作 |
| Pagination | FirstIcon、PrevIcon、NextIcon、LastIcon、EllipsisIcon 覆盖默认图标；ActiveColor / ActiveVariant 独立设置当前页 |
| Stepper | StepItem.Content 构建当前正文，UI 覆盖单项；Disabled 只禁用步骤标题；List / Panel 分别配置列表和正文 |
| Timeline | CurrentID 优先于 Current；Reverse 不重排数据；DateSlot 配置 Time，OnSelect 不自动改进度 |
| Banner | CloseButton 控制显示，Close / CloseIcon 配置关闭按钮；OnSelect 与内部动作独立，OnClose 只通知用户关闭 |
| ErrorView | Clear 默认 true，但默认按钮仍需 OnRetry；Retry 配置按钮，Actions 始终优先，StatusSlot 配置状态码 |
| FileUpload | Variant 使用 UploadArea / UploadButton；Layout 默认 UploadGrid；Position 仅影响区域内的列表。ShowIcon、FileImage、FileDelete、Preview 默认 true，支持图标、删除按钮配置及 Highlight |

可交互的 UploadArea 点击或按 Enter/Space 打开文件选择器，名称优先取 Label。内置列表时，文件动作与标题/图标选择入口分开，保证原生辅助技术可以访问各动作。Choose 按钮只在 UploadButton 或 `Interactive: nui.Set(false)` 时出现；Dropzone=false 取消文件拖放目标。Preview=false 隐藏整个列表，FileImage=false 只隐藏应用提供的预览配件，Files 数据均保留。选择或移除改变队列时发送 OnChange，新拒绝项由 OnReject 通知；移除后返回选择入口焦点。完整默认值与资源职责见 [FileUpload](advanced.md#fileupload-与资源所有权)。Sidebar 的区域回调、尺寸和键盘操作见[侧栏说明](advanced.md#sidebar-状态与尺寸)。

## InputMenu 值与搜索策略

### 先决定保存什么值

| 模式 | 状态字段 |
| --- | --- |
| 单选 | `SearchState.Selected` 保存选项 ID，Query 保存编辑草稿 |
| 多选 | `SelectedValues` 保存全部选中 ID；删除按钮或空草稿 Backspace 可移除项 |
| InputMenuAutocomplete | Query 是可自由输入的文本；Selected 保留最后选中的 ID，强制单选并允许自定义提交 |

多选变化通过 OnValuesChange 通知，OnSelect 报告本次激活的 ID。

### 本地或外部搜索

默认只按 Label 做不区分大小写的包含匹配。FilterFields 可指定 `label`、`description`、`value`；nil 表示只搜索 label。旧的 SearchSelect 入口保留 label + description。

`ExternalFilter=true` 直接展示应用传入的结果。请求、取消和过期结果处理由 QueryState 或应用服务完成。选项的身份、标签、描述或禁用状态变化时会关闭旧面板；ExternalFilter 不会改变这一规则，也不会自动保持面板打开。

### 创建候选项

提供 OnCreate 后，列表可显示一个能通过键盘激活的创建行。默认只在无匹配结果时显示；CreateWhenAlways 在查询非空且没有相同 value 时提供创建行。

创建只发送请求，保留 Query、筛选和打开状态，不触发普通选择，也不会替应用添加选项。应用应在创建成功后更新数据源。

### 打开和重置

OpenOnClick 和 OpenOnFocus 默认均为 false。输入文字、方向键或尾部按钮仍可打开；需要点击即打开时显式设置 OpenOnClick: Set(true)。SearchSelect 保留旧的点击打开行为。Open 可绑定持久 bool；Loop 默认 false，HighlightOnHover 默认 true。

ResetSearchTermOnSelect 和 ResetSearchTermOnBlur 默认均为 true，分别控制选择后和关闭后的草稿。单选恢复选中标签，多选清空草稿。Autocomplete 忽略 Multiple、ResetSearchTermOnSelect 和 ResetModelValueOnClear，选中建议时始终回填完整文字，普通失焦保留自由文本。

ResetModelValueOnClear 只控制清空按钮，默认 true；false 保留选中值。它不控制键盘删除普通文字。多选可以在空草稿时清空全部标签，或通过 DeleteIcon 按钮逐项移除；实际多选值变化只通知一次 OnValuesChange。键盘删除标签后焦点返回输入框。

## 示例

FormOptions.LoadingAuto 默认 true，异步处理期间禁用表单控件且保留字段快照；设置 Set(false) 可继续编辑并使旧结果失效。OnError 在提交校验失败时调用一次，包含全局错误；详细的提交、取消和重置规则见 [Form](forms-data.md#form)。

Gallery 的 **Props & behavior** 页面演示加载、选择上限、标签添加、手动 Tabs、搜索重置、分页、弹层和文件预览；**Orientation** 页面演示方向切换。启动命令见[运行示例](../README.md#运行示例)。

## 使用限制

Listbox 两种方向都只构建可见项目及少量预留项目。横向项目和组标题使用固定 ItemWidth，纵向行高由原生 List 管理。

Slider / RangeSlider 的实际值、范围和步长可以提供给辅助技术，但 MyGo 缺少纵向方向元数据的设置接口；这一限制适用于全部纵向模式，包括没有 Inverted 的情况。滑块提示可按每个滑块定位，目前没有独立的 Tooltip 延迟配置。InputRating 使用单焦点 slider 角色。

FormField 会关联本库控件的标签；使用自定义 MyGo 控件时，应用需要自行处理标签关联。Toast 的动作回调和 ButtonOptions.Style 会随通知保存，闭包不能捕获临时 Element、Parts 或 Context；Style 参数仅在当次构建中有效。

Go Options 的字段和类型以本库 API 为准，不能直接照搬 Nuxt Vue Props。日期时间操作与迁移见[日期组件](forms-data.md#calendarinputdateinputtime)，平台、输入法和读屏验证范围见[兼容性说明](compatibility.md)。
