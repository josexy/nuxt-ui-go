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
| Sidebar | Resizable=false；Collapsible、Toggle、AutoClose=true；展开/收起宽度 256/64 DIP |
| 弹层与通知 | Popover.Dismissible=true；Toast.Close、Toast.Progress=true |
| Listbox / Accordion | HighlightOnHover=true；Collapsible=true |
| InputDate | ShowCalendar=false；HideCalendar 优先关闭日历 |

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

`SearchInput` 默认 true。设为 false 后不创建搜索框，也不会用保留的 Query 筛选列表。触发器上的 ↑、↓ 均可打开菜单：有搜索框时聚焦搜索框，否则聚焦列表；禁用触发器不能打开。

默认在选择后和关闭后清空 Query，分别由 `ResetSearchTermOnSelect`、`ResetSearchTermOnBlur` 控制。关闭包括 Escape、外部点击和应用修改打开状态。

再次选择同一个值也会关闭、重置并触发 OnSelect；OnChange 只在选中值实际变化时触发。Invalid / ErrorMessage 会同步影响字段错误语义和边框。

### Pagination 的相邻页和外观

`SiblingCount` 优先于旧的 `Siblings`，范围为 0–10。例如当前第 5 页、共 10 页、相邻数为 0 时：

| 配置 | 显示结果 |
| --- | --- |
| ShowEdges=true | `1 … 5 … 10` |
| ShowEdges=false | `5` |

接近首尾时会扩大边缘页码范围；总页数较少时不显示省略号。ShowControls 控制翻页按钮，ShowEdges 控制首尾页码。Size、Color、Variant 设置普通页码外观，ActiveColor、ActiveVariant 单独设置当前页。

## 常用参数入口

### 开关、标签和评分

| 组件 | 常用选项与行为 |
| --- | --- |
| Switch | Loading 阻止鼠标和键盘切换；LoadingIcon、CheckedIcon、UncheckedIcon 配置图标；减少动态效果时停止旋转 |
| Checkbox / CheckboxGroup / RadioGroup | Variant 可选 ChoiceList、ChoiceCard、ChoiceTable；Indicator 可选 IndicatorStart、IndicatorEnd、IndicatorHidden；Highlight 强调边框或指示器，未选中时也生效 |
| InputTags | AddOnTab、AddOnBlur、AddOnPaste 控制何时添加标签；Duplicate 控制重复项；MaxLength 限制编辑文本 |
| PinInput | Color、Variant、Size 设置外观；`Separator: []int{3}` 在第三格后分组，仍使用一个编辑器和一个 Tab 停靠点 |
| InputRating | Clearable 允许点击当前评分或按空格清零；Hoverable 开启悬停预览；未设置 EmptyIcon 时沿用 Icon |

Checkbox 只使用 CheckedIcon，不使用 Switch 的 UncheckedIcon。ChoiceCard 选中后会着色，隐藏指示器时仍能识别选中状态。

InputTags 的 MaxLength 按 UTF-16 单元限制用户插入，0 表示不限；它限制的是编辑草稿，不负责提交时的标签长度校验。启用 AddOnPaste 后，粘贴会按标签处理，即使因重复或数量上限被拒绝，也不会再作为普通文本插入；关闭时按普通文本粘贴。启用 AddOnTab 后，非空草稿会在 Tab / Shift+Tab 时提交并保留焦点，空草稿则正常移动焦点。

InputRating 的横向方向键随 RTL 调整，纵向使用 ↑↓，另一轴按键无效。评分到首尾停止；Home/PageUp 到 1，End/PageDown 到 Max。清零后悬停预览随即消失。

### 滑块和内容切换

| 组件 | 常用选项与行为 |
| --- | --- |
| Slider / RangeSlider | PageUp/PageDown、Shift+方向键移动 10 个 Step；OnChange 报告实时变化，OnCommit 在有变化的按键或拖拽结束时通知 |
| RangeSlider | 两个滑块可交叉，随后自动排序并转移活动焦点；小于 `Step × MinStepsBetweenThumbs` 的移动被拒绝；Tooltip 分别显示两个值 |
| Tabs | ActivationMode 默认 ActivationAutomatic；ActivationManual 下方向键只移动焦点，Enter/Space 才激活并通知一次 |
| Accordion | Collapsible 默认 true；单开模式设为 false 后不能再次点击收起当前项，仍可切到另一项；多开模式不受此选项限制 |
| Carousel | Arrows 控制翻页按钮，Dots 开启圆点导航；圆点的页码名称支持语言切换 |

### 弹层和通知

| 组件 | 常用选项与行为 |
| --- | --- |
| Popover | Mode 默认 PopoverClick，可改为 PopoverHover；OpenDelay / CloseDelay 只用于悬停模式，类型为 time.Duration，默认 0 |
| Tooltip | Placement、Offset 控制位置；Open 可绑定打开状态；Kbds 显示快捷键 |
| Toast | Close 控制关闭按钮，Progress 控制倒计时条；Actions 可配置多个动作，每项含 Label、OnAction、ButtonOptions、Close |

Popover 的 Dismissible 默认 true，控制 Escape 和外部点击关闭，应用始终可以直接修改 open。悬停模式允许指针跨过锚点与面板之间的间隙；Escape 关闭后，指针停在原处不会立即重新打开。网站参数面板中的延迟以毫秒填写。

Tooltip 即使绑定了 Open，也继续响应悬停延迟、键盘焦点、失焦、点击和 Escape。应用将 Open 改为 false 后，不会因为仍悬停或聚焦而立即重开；新提示出现时会关闭旧的受控提示。

Toast 动作默认执行后关闭通知，单项设置 `Close: nui.Set(false)` 可保留。悬停或聚焦使通知暂停时，倒计时条也停止；常驻通知没有倒计时条。

### 侧栏、导航和文件选择

| 组件 | 常用选项与行为 |
| --- | --- |
| Sidebar | State 保存展开宽度和窄窗口打开状态；MinWidth/MaxWidth 默认 160/480 DIP；Breakpoint 未设置或为 0 时始终使用桌面布局 |
| NavigationMenu | 纵向 Compact 模式默认提供 Tooltip 和子级 Popover，可分别用 `nui.Set(false)` 关闭；PopupPlacement 修改子级弹层位置 |
| FileUpload | Variant 使用 UploadArea / UploadButton；Layout 使用 UploadList / UploadGrid；Dropzone、Interactive、FileDelete、Preview 控制交互与显示 |

可交互的 UploadArea 整个区域都是按钮，点击或按 Enter/Space 打开文件选择器，名称取 Label。Choose 按钮只在 UploadButton 或 `Interactive: nui.Set(false)` 时出现。Preview=false 只隐藏预览，Files 数据仍保留。Sidebar 的区域回调、尺寸和键盘操作见[侧栏说明](advanced.md#sidebar-状态与尺寸)。

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

OpenOnClick 默认 true，OpenOnFocus 默认 false。ResetSearchTermOnSelect 和 ResetSearchTermOnBlur 默认均为 true，分别控制选择后和关闭后的草稿。单选恢复选中标签，多选清空草稿，Autocomplete 保留自由文本。

## 示例

Gallery 的 **Props & behavior** 页面演示加载、选择上限、标签添加、手动 Tabs、搜索重置、分页、弹层和文件预览；**Orientation** 页面演示方向切换。启动命令见[运行示例](../README.md#运行示例)。

## 使用限制

Listbox 两种方向都只构建可见项目及少量预留项目。横向项目和组标题使用固定 ItemWidth，纵向行高由原生 List 管理。

Slider / RangeSlider 的实际值、范围和步长可以提供给辅助技术，但 MyGo 缺少纵向方向元数据的设置接口；这一限制适用于全部纵向模式，包括没有 Inverted 的情况。滑块提示可按每个滑块定位，目前没有独立的 Tooltip 延迟配置。InputRating 使用单焦点 slider 角色。

FormField 会关联本库控件的标签；使用自定义 MyGo 控件时，应用需要自行处理标签关联。Toast 的动作回调和 ButtonOptions.Style 会随通知保存，闭包不能捕获临时 Element、Parts 或 Context；Style 参数仅在当次构建中有效。

Go Options 的字段和类型以本库 API 为准，不能直接照搬 Nuxt Vue Props。日期时间操作与迁移见[日期组件](forms-data.md#calendarinputdateinputtime)，平台、输入法和读屏验证范围见[兼容性说明](compatibility.md)。
