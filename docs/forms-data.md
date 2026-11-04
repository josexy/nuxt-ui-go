# 表单、日期与数据组件

本页介绍带校验的表单、专用输入框、日期时间，以及列表、表格和树。普通文字输入、字段标签和下拉选择见[输入、选择与字段](inputs.md)。

| 当前任务 | 建议组件 |
| --- | --- |
| 校验多个字段并提交 | [Form](#form) |
| 输入数字、标签、验证码或评分 | [专用输入组件](#数值标签pin评分) |
| 搜索并选择列表项，接入远程搜索 | [Listbox、SelectMenu、QueryState](#listboxselectmenu-与异步搜索) |
| 输入日期、时间或颜色 | [日期时间组件](#calendarinputdateinputtime)、[ColorPicker](#colorpicker) |
| 展示可排序表格或层级数据 | [Table](#table)、[Tree](#tree) |

## 生命周期与主题

业务值和组件状态按窗口保存，不要在每次构建中重新初始化，也不要复制已使用的状态。返回的 `ui.Element` 仅在当前构建有效，Element 和 Context 不放入应用模型或后台任务。

FormTicket、QueryTicket、TreeTicket 是异步操作的标识，包含实例身份和版本。结果必须交回发起请求的状态对象；数据已变化或请求已取消时，旧结果会被拒绝。后台只处理数据快照，通过 `Window.Update` 返回 UI 线程，关窗时取消业务任务。

InputNumber、InputDate、InputTime 沿用 Input 的尺寸、五种字段变体、颜色、只读、禁用、图标和错误设置。其他组件按各自 Options 配置；本页末尾有[样式部件索引](#样式部件索引)。

## Form

### 注册字段并提交

Form 管理字段校验、错误显示和首个错误字段的焦点。FormState 保存字段初值、是否修改（Dirty）、是否已交互（Touched）、可见性、错误和 Handle。

字段通过 FormField 的 Name / Value 注册。Name 在一次表单构建中必须非空且唯一，Value 是应用明确转换后的字符串。一个字段对应一个编辑值。

```go
// form nui.FormState 和 name string 保存在窗口模型中。
nui.Form(c, theme, &form, nui.FormOptions{
    Validate: func(values nui.FormValues) nui.FormErrors {
        if values["name"] == "" {
            return nui.FormErrors{"name": "请输入名称"}
        }
        return nil
    },
    OnSubmit: func(values nui.FormValues) { save(values["name"]) },
}, func(inner nui.Theme) {
    nui.FormField(c, inner, nui.FormFieldOptions{
        Name: "name",
        Value: name,
        Label: "名称",
        Required: true,
    }, func(field nui.Theme) {
        nui.Input(c, field, &name, nui.InputOptions{})
    })
    nui.Button(c, inner, "保存", nui.ButtonOptions{}).OnClick(form.Submit)
})
```

FormValues 和 FormErrors 都是 `map[string]string` 的命名类型。表单不通过反射推导结构；跨字段校验直接读取完整 Values。上例的 save 由应用提供，耗时保存应放到后台。

### 校验、隐藏和重置

| 操作或状态 | 行为 |
| --- | --- |
| 首次注册字段 | 当前值作为 Dirty 比较的初值 |
| 失焦或 `Touch(name)` | 将字段标记为 Touched |
| `Submit()` / `Validate(...)` | 执行校验，标记当前字段为 Touched，并聚焦构建顺序中的首个错误字段 |
| 隐藏或禁用字段 | 保留模型、Dirty、Touched 和错误，但不进入 Values、提交校验或错误焦点选择 |
| `Reset()` | 把当前值设为新初值并清理错误，不改写应用绑定的数据 |

非法的数字、日期或其他编辑草稿也会阻止提交，避免提交上一次有效值。业务校验由 Submit 或显式 Validate 触发，不会在每次输入时自动运行。

隐藏字段再次出现时继续检查已有错误。需要恢复业务初值时，应用先恢复绑定数据，再按需要重置表单状态；单独调用 Reset 不会恢复输入内容。

### 异步校验

1. 在 UI 线程调用 `ticket := form.BeginAsync()`，获得当前字段快照。
2. 后台任务只读取 `ticket.Values` 并执行校验。
3. 在 `Window.Update` 回调中调用 `form.CompleteAsync(ticket, errors)`。
4. 返回 true 表示结果仍有效且无错误，再由应用执行保存。

CompleteAsync 不会再次调用 OnSubmit。字段值、原始编辑文字、可见性变化，或 Cancel、新请求，都会使旧票据失效。`Cancel()` 只让表单拒绝旧结果，实际 I/O 的 context 仍由应用取消。

内置输入、选择和单值开关会自动关联字段 Handle；组合组只关联第一个控件。自定义控件需把 `form.Field(name).Handle` 绑定到真实可聚焦控件。

设置页的 **Project setup** 演示分步注册、条件隐藏、跨字段日期和预算校验、错误焦点及可取消提交。其中 Confirm project 只模拟本地确认，不发送网络请求。

## 数值、标签、PIN、评分

### InputNumber

InputNumber 绑定 `*Optional[float64]`，另用 NumberState 保存输入草稿。这样可以区分空值、未完成输入（例如单独的负号）和有效数值。

```go
// value nui.Optional[float64]、numberState nui.NumberState 按窗口保存。
nui.InputNumber(c, theme, &value, &numberState, nui.InputNumberOptions{
    InputOptions: nui.InputOptions{Label: "数量"},
    Orientation: nui.Vertical,
    Min: nui.Set(0.0),
    Max: nui.Set(100.0),
    Step: 1,
})
```

默认 Step=1，支持上下方向键、增减按钮、Min / Max、只读、禁用、前后图标和清空。Enter 恢复或格式化最后有效值。数字使用 ASCII 和小数点 `.`，不启用 IME；NaN / Inf 模型和非法边界会被拒绝。它不提供地区或货币格式，float64 也不适合作为精确货币类型。

| Orientation | 布局 |
| --- | --- |
| Horizontal（默认） | 左侧减号、居中数值、右侧加号 |
| Vertical | 数值靠左，右侧上下排列增减箭头 |

两种布局字段高度相同，按钮继承字段颜色和尺寸，使用链接样式，RTL 下按起始侧和结束侧调整。方向切换保留值、未完成草稿、选区、撤销历史和焦点，不触发 OnValueChange。点击外部方向选择器时，焦点正常移到选择器，输入框不会主动抢回。

Style 接收 root、editor、leading、trailing；后两者指应用传入的图标，增减按钮没有单独公开的部件。修改 root 内边距时要保留按钮空间。Gallery 的 **Advanced forms → Orientation** 可切换同一个 Seats 字段的布局。

### InputTags

InputTags 绑定 `*[]string`，TagsState.Query 保存尚未添加的草稿。Enter 添加标签，空草稿时 Backspace 删除末项，标签上的 × 删除对应项。组件会去除首尾空白并按大小写区分去重，MaxTags 控制数量。

默认不批量处理粘贴。设置 `AddOnPaste: nui.Set(true)` 后，按 Separators 分割，默认分隔符只有逗号；需要分号或换行时设置 `Separators: ",;\n"`。批量粘贴保留已有草稿。更多添加和重复策略见[参数说明](options.md#开关标签和评分)。

标签使用 Elevated 背景和 Accented 内边框，字号随字段尺寸，只有 × 可点击删除。编辑器聚焦或标签按钮键盘聚焦时，外框显示强调色与外圈；错误时使用错误色，None 不显示焦点外圈。当前不支持拖拽排序或远程实体匹配。

### PinInput

PinInput 绑定 `*string`，另保存 PinState。默认是 5 格文本；六位数字验证码可使用 `PinInputOptions{Numeric: true, Length: 6}`。Length 范围为 1–32，按 rune 限长，Numeric 只接受 ASCII 数字。

它使用一个真实编辑器和多个可视格子，Tab 只进入一次。编辑、粘贴、选区、复制和撤销/重做都由该编辑器处理。Numeric 模式不启用 IME，文本模式启用。长度达到要求且值变化时才触发 OnComplete，组件不会读取短信或自动提交。

| 操作或选项 | 行为 |
| --- | --- |
| 点击已有字符 | 选中该位，随后输入替换并移至下一位；末位继续输入会替换末位 |
| 粘贴多个字符 | 覆盖对应位置，不把后面的内容向后挤 |
| Backspace | 删除选中的字符；点击空格子或格子间隙时，光标移到现有内容末尾 |
| SeparatorEvery / Separator | 按固定间隔或指定位置分组；例如 `SeparatorEvery: 2` 或 `Separator: []int{3}` |
| ReadOnly / Placeholder / 遮罩 | 分别控制只读、空格子文字和内容显示 |

格子间距为 6 DIP，随 Density 缩放。分隔符是居中的 Dimmed 色圆点，两侧间距相同。当前格子显示 3 DIP、25% 不透明度的主题色焦点外圈；错误时改为 Error，高对比模式使用不透明外圈，None 不显示外圈。

聚焦的空格子隐藏 Placeholder 并显示闪烁光标；减少动态效果时光标常亮，只读、禁用或失焦时隐藏。FormField 有 Error 时，各格子都显示错误边框。文字缩放和遮罩不会改变点击位置的判断。

### InputRating

InputRating 绑定 `*int`，默认 Max=5，0 表示未评分。只支持整数评分，不含半星、任意精度或拖拽评分。

| 操作 | 行为 |
| --- | --- |
| 横向 ←→ / 纵向 ↑↓ | 按一个等级移动，到首尾停止；横向随 RTL 调整，另一轴无效 |
| Home / PageUp | 设为 1 |
| End / PageDown | 设为 Max |
| Space | 空评分设为 1；启用 Clearable 后可清零当前评分 |
| Enter | 不处理 |

方向键不会把评分从 1 降为 0；清零需启用 Clearable。键盘焦点外圈位于当前星上，空评分时位于第一颗星。只读评分仍可 Tab 聚焦，但不响应修改按键。

星标大小按 XS 到 XL 为 12/16/20/24/28 DIP，间距 2 DIP，组件按内容宽度排列。可交互时悬停星标放大到 1.1 倍。Hoverable 开启后，悬停只预览，经过星间空隙保留预览，离开或清零后取消。FormField 出错时，实心星和焦点外圈使用 Error 色。

### CheckboxTri

CheckboxTri 使用 CheckState：Unchecked、Checked、Indeterminate。混合态向辅助技术提供 Mixed 状态，首次激活后变为 Checked，之后在 Checked / Unchecked 间切换。

父子选中关系由业务计算。只需要开关状态时，使用绑定 `*bool` 的 Checkbox。

### RangeSlider

RangeSlider 接收两个 `*float64`，默认范围 0–100、Step=1。两个滑块交叉后会排序并转移活动焦点，保持 low≤high；不满足最小间距的移动会被拒绝。应用传入的越界值会在构建时限制到合法范围。

常用选项包括 Orientation、Inverted、Tooltip、MinStepsBetweenThumbs、Color、Size、OnChange 和 OnCommit。两个滑块有独立焦点，并向辅助技术提供各自的实际值和最小值/最大值名称。

放在 FormField 中时，点击标签聚焦第一个滑块，字段描述和错误也关联到它。方向和键盘说明见[参数文档](options.md#滑块和内容切换)，纵向辅助技术限制见[兼容性说明](compatibility.md#无障碍与动态效果)。

## Listbox、SelectMenu 与异步搜索

### Listbox：直接展示可选择列表

ListboxItem 包含 ID、Label、Description、Group、Disabled、Icon，ID 必须非空且唯一。选择值为 `*[]string`，空切片表示未选择；ListboxState 保存活动项和列表状态。

默认 SelectionToggle，单选和多选均可再次激活以取消。SelectionReplace 在激活时替换选择；多选时还支持 Shift 连续范围选择。暂时不在数据源中的选中 ID 会保留，便于异步搜索；永久失效的 ID 由应用清理。

相邻且同组的项目上方显示不可选择的组标题。选项行外观见[输入组件](inputs.md#选项行的外观)。MD 行高约 32 DIP，有描述时增加一行；列表圆角 8 DIP，键盘聚焦时整体显示焦点外圈，空状态居中。

Listbox 没有直接的 Size / Color 字段，尺寸先取外层 FormField，再取 `WithComponent(ComponentListbox, ...)`；嵌入 SelectMenu 时使用菜单的 Size / Color。

| 方向或按键 | 行为 |
| --- | --- |
| 纵向 | 原生虚拟列表，↑↓ 移动 |
| 横向 | 按固定 ItemWidth 显示和裁剪项目，←→ 移动；组标题也占一项宽度 |
| Home/End、Enter/Space | 到首尾、激活项目 |
| 输入字符 | 按前缀检索项目 |
| 数据重排 | 按稳定 ID 保留高亮；删除高亮项后回到第一个可用项 |

两种方向都只构建可见及少量预留项目。同步 Build 可自定义内容，Loop 和 HighlightOnHover 控制循环导航和悬停高亮。

### SelectMenu：搜索弹层

SelectMenu 将触发器与弹层内的搜索框、Listbox 组合起来。SelectMenuState 保存 Open、Query 和 Listbox，Size / Color 同时作用于三个区域。

默认 ↑、↓ 都能从触发器打开，打开后聚焦搜索框；SearchInput=false 时聚焦列表。搜索框中按 ↑↓ 进入列表，Escape 关闭并返回触发器。单选后关闭，多选继续打开。

默认只搜索 Label；FilterFields 可指定 label、description、value（ID）、group。ExternalFilter=true 时直接使用应用提供的结果。Loading、Error、EmptyText 用于显示加载、失败或空状态。

| 选项 | 行为 |
| --- | --- |
| ResetSearchTermOnSelect / ResetSearchTermOnBlur | 默认均启用，分别控制选择后和关闭后清空搜索词 |
| Clearable | 有选择时，用清空按钮替换触发器尾部箭头 |
| OnCreate | 显示可点击或键盘激活的创建行，只发出创建请求，不自动新增或选中 |
| CreateWhenAlways / CreateAtTop | 默认仅无匹配时显示创建行；可改为没有相同 ID 时显示，并调整位置 |

创建请求保留查询和打开状态。未选择时使用 Dimmed 占位文字；Subtle 触发器悬停时背景加深。弹层搜索框使用 None 外观，下方有分隔线，列表无边框、行内缩 4 DIP，默认最大高度 240 DIP，内容少时自动缩短。

### 缓存与大量数据

Listbox 缓存数据快照、ID 索引、分组和选择集合。输入未变时不重建这些结构，应用原地修改切片也会被识别。但每次构建仍要线性比较快照，字符前缀检索也可能遍历全部项目。

SelectMenu 缓存过滤结果和已选标签，临时缺失的选项保留上一次标签。首次处理大量选中项仍需要线性工作。虚拟列表减少的是 UI 构建量，数据处理成本仍取决于数据规模。

### QueryState：拒绝过期搜索结果

QueryState[T] 可用于 SelectMenu 或 InputMenu 的应用数据源。应用负责发请求，QueryState 负责取消标识、加载状态和结果是否有效的检查。

```go
// 在 UI 线程启动；queryState 按窗口保存。
ctx, ticket := queryState.Start(parentContext, query)
go func() {
    items, err := fetchItems(ctx, ticket.Query)
    window.Update(func() {
        queryState.Complete(ticket, items, err)
    })
}()
```

`fetchItems` 是应用提供的请求函数。下一次构建将 queryState 的 Items、Loading、Error 传给菜单；远程搜索通常配合 ExternalFilter 使用。

Start 会取消前一个 context；Complete 克隆结果切片并拒绝过期或来自其他实例的票据。离开视图或关窗时调用 `queryState.Cancel()`。库不创建网络任务，也不规定防抖时间、模糊匹配或远程分页方式。

多选列表提供选择角色和选中行语义，但 MyGo 尚未开放独立的 multiselectable 标志，完整读屏体验仍需验证。

## Calendar、InputDate、InputTime

### 先选择值类型

| 类型 | 表示什么 |
| --- | --- |
| `Date{Year, Month, Day}` | 1–9999 年的公历日期，不含时区和时刻 |
| `Optional[Date]` | 单个日期，允许为空 |
| `DateSelection{Start, End}` | 单日或日期范围，两端均为 Optional[Date] |
| `ClockTime{Hour, Minute, Second}` | 不含日期和时区的时间，小时为 0–23 |

Date 提供 Valid、Compare、AddDays、AddMonths、String、ParseDate。AddMonths 遇到目标月没有对应日期时，取该月最后一天。DateFromTime 读取传入 time.Time 所在 Location 的年月日，不做 UTC 转换；需要表示具体时刻时，由应用选择时区并处理夏令时。

### Calendar：点击选择日期

Calendar 接收 `*DateSelection` 和 CalendarState。Range=false 为单日，Range=true 为范围：先选起点，再选终点；反向选择会规范化顺序，完整范围后再次点击会开始新范围。

Min、Max、IsDateDisabled 限制可选日期，跨过禁用日期的范围会被拒绝。WeekStart 设置一周起始日，Today 可以传入固定日期用于测试。

| 按键 | 行为 |
| --- | --- |
| ←→ / ↑↓ | 按天 / 按周移动，左右随 RTL 调整 |
| PageUp / PageDown | 上个月 / 下个月 |
| Home / End | 月首 / 月末 |
| Enter / Space | 选择日期；禁用日期可高亮但不能选择 |

日历使用单月六周网格，标题、星期和导航文案支持语言切换。首次构建可设置 CalendarState.Month / Active。之后外部修改 Start 会同步高亮和月份；只改 End 时跟随有效 End。外部赋值不触发 OnChange，清空选择保留浏览月份；正常翻月或完成范围选择后不会自动跳回起点。

### InputDate：分段输入日期

InputDate 绑定 `*Optional[Date]`，DateInputState 保存草稿和日历状态。需要日历按钮时显式启用 ShowCalendar：

```go
// app.Date 为 nui.Optional[nui.Date]，app.DateInput 为 nui.DateInputState。
nui.InputDate(c, theme, &app.Date, &app.DateInput, nui.InputDateOptions{
    InputOptions: nui.InputOptions{Label: "开始日期", Clearable: true},
    ShowCalendar: true,
})
```

年、月、日按窗口语言和区域格式排列，每段可单独聚焦。日期和时间字段都沿用 Input 的鼠标、Tab 聚焦外框。

| 操作 | 行为 |
| --- | --- |
| ←→、Tab / Shift+Tab | 在分段间移动；Tab 按视觉顺序进入或离开字段 |
| 输入数字 | 达到位数或数值上限后自动进入下一段；年份按四位输入 |
| ↑↓、Home / End | 当前段循环增减，或到该段最小 / 最大值 |
| Backspace / Delete | 清空当前段；Ctrl/Cmd+A 后删除清空当前日期端点 |
| Ctrl/Cmd+C/X/V | 复制、剪切、粘贴当前端点，也支持系统编辑命令 |

修改年月时会将超出月底的日期限制到目标月末。粘贴接受 ISO 日期或当前区域顺序的数字日期。

DateInputState.Text 保存 ISO 草稿，未填完的段用 `-` 表示。非法粘贴保留原文和 Error，屏幕上的上一组分段值及最后有效业务值不变；全部日期段清空时才清空业务值。

OnChange 报告规范草稿或被拒绝的粘贴文本。外部赋值会刷新分段，但不触发 OnValueChange；切换语言保留草稿和字段焦点。只读时允许导航与复制，不允许编辑。

ShowCalendar 默认 false，开启后提供日历图标和 Popover，选择日期会更新分段并关闭面板。旧的 HideCalendar 仍保留，优先于 ShowCalendar。旧应用若依赖默认显示日历，需要显式补上 ShowCalendar。

### InputDateRange：输入起止日期

InputDateRange 绑定 `*DateSelection`，并保存 DateRangeInputState。两端分段共用外框，中间用减号分隔，左右键和 Tab 可跨端点移动。

起点改到终点之后会清空终点；终点不能早于起点。可选的范围日历在第二次选择完成后关闭，范围变化通过 Calendar.OnChange 通知。

放在同一个 FormField 中时，出错聚焦到起点首段。需要两端各有错误和焦点时，分别注册两个 InputDate。

### InputTime：输入时间

InputTime 绑定 `*Optional[ClockTime]`，TimeInputState 保存草稿。它分段编辑小时、分钟、可选秒和上下午；上下键在当前段循环，A / P 选择 AM / PM。

TwelveHour 是 Optional[bool]，未设置时由窗口语言和系统区域的 HourCycle 决定；用 `nui.Set(true)` 或 `nui.Set(false)` 固定小时制。显示顺序、分隔符和上下午文字随区域变化，Text 始终保存 `HH:MM[:SS]` 或 `hh:mm[:ss] AM/PM` 草稿；完整时间粘贴使用同一格式。

切换小时制保留业务值和未完成草稿。隐藏秒后再编辑小时或分钟，原秒值仍保留；正在聚焦的秒或上下午段被隐藏时，焦点移到分钟或小时。

### 日期时间的样式与限制

InputDate、InputDateRange、InputTime 共用 Input 的尺寸、变体、颜色、图标、加载、禁用、只读和 Clearable。Root 控制外框，Editor 控制分段文字基础样式，Segment 可改占位、焦点背景和圆角。

Style 的 editor 参数是当前排列中的第一个可聚焦分段，可绑定 Handle；它已不再是 TextInputBase，不能使用原生文本选区 API。Placeholder、Password、MaxLength 不适用于分段输入，占位文字通过语言目录覆盖。

当前只支持公历与 ASCII 数字，分段输入不支持撤销/重做。日历多选、年/月选择面板、多月并排、周序号、其他历法、时间范围、时间弹层和时区选择器尚未提供。完整读屏体验尚未验证。

## ColorPicker

ColorPicker 绑定 `*ui.Color`，并保存 ColorPickerState。它提供饱和度/亮度平面、色相、透明度、HEX / HEXA、色板，以及 0–255 的 R/G/B/A 数值输入。修改任一表示时，从同一个 ui.Color 同步其他通道。

平面和滑条保留原生键盘操作。RGB 小数会截断为 uint8，建议按整数步进。内部色板和 HEX 输入只接受 Native 主题，不提供任意部件构建回调。

HEX 编辑受[输入法取消问题](compatibility.md#输入法取消)影响。当前不含 OKLCH/HSL 编辑、屏幕取色、色彩管理或完整 Nuxt ColorPicker API。

## Table

### 准备行和列

TableState[T] 保存当前行数据、排序、选择和列布局。数据变更时调用 SetRows，不要在每次视图构建中重复设置。

```go
// 在应用模型中声明类型和状态。
type TaskRow struct {
    ID string
    Title string
}

var tasks nui.TableState[TaskRow]

// 初始化或数据变化时执行。
tasks.SetRows([]TaskRow{
    {ID: "task-1", Title: "整理需求"},
    {ID: "task-2", Title: "检查文档"},
}, func(row TaskRow) string { return row.ID })
```

在视图中定义显示列并创建表格：

```go
nui.Table(c, theme, &tasks, []nui.TableColumn[TaskRow]{
    {
        ID: "title",
        Title: "任务",
        Width: 240,
        Value: func(row TaskRow) string { return row.Title },
    },
}, nui.TableOptions{Label: "任务列表", Height: 320})
```

SetRows 复制行值并检查 ID 非空且唯一；T 中如果有指针或切片，内部数据的所有权仍由应用负责。替换数据会移除已不存在的选中 ID。

### 排序、选择和自定义单元格

| 字段或方法 | 用途 |
| --- | --- |
| `TableColumn.Compare` | 本地稳定排序，返回两行的比较结果 |
| `TableColumn.Cell` | 自定义单元格，接收 Context、Theme、行值和是否选中 |
| `Selection` / `Sort` | 分别为 `ui.Selection[string]`、`ui.SortOrder` |
| `state.List.Columns` | 保存列宽和顺序，可由应用持久化 |
| `Row` / `RowID` / `Len` | 读取当前视图中的行、行 ID 和数量 |

列还可设置 Width、MinWidth、MaxWidth 和 Align。排序后的行位置会变化，编辑数据时应通过 RowID 或 Selection 中的 ID 回查业务模型。CellBuilds 可用于统计当前构建中的单元格调用次数。

### 筛选与功能范围

搜索、筛选和分页由应用先计算当前 rows，再调用 SetRows。数据浏览器示例展示搜索、团队/状态筛选和每页 100 条数据；服务端请求、总量和缓存由应用管理。

EmptyText / EmptySlot 设置空状态，加载和失败提示可放在表格外。虚拟化减少可见单元格的构建量，数据生成、ID 校验、排序和筛选仍取决于数据量。

`Fixed` **只禁止移动列和调整列宽**，没有冻结滚动的效果。当前还不支持冻结列、多层表头、整行或表头的任意构建回调、行展开、分组聚合及完整 TanStack 语义。底层 List 有高度测量，但复杂动态单元格仍需应用验证，常规用法应保持稳定行高。

## Tree

### 保存节点、选择和展开状态

TreeState.SetNodes 接收 TreeNode 列表，检查全树 ID 非空且唯一。各状态用途如下：

| 状态 | 用途 |
| --- | --- |
| `state.Outline.Open` | 已展开的节点 ID |
| `state.Selection` | 已选节点 ID |
| `state.Active` | 键盘当前高亮的节点 ID |

默认单选，点击行同时选择并切换展开。方向键只导航，Enter / Space 选择。Multiple 开启多选，SelectionBehavior 决定切换或替换选择；PropagateSelect、BubbleSelect 控制父子联动，Loop 控制循环导航。

节点可设 Disabled，OnSelect / OnToggle 可取消本次操作。父子联动只作用于已加载节点，禁用节点会阻断联动；后续加载的子节点不会自动继承选择。

### 缩进与虚拟化

Nested 默认 true，显示祖先竖线，并让子项背景和点击区域随层级缩进。设为 `nui.Set(false)` 后，背景和点击区域占整行，内容仍缩进。切换保留选择、展开、焦点和滚动位置。

两种模式都使用原生虚拟列表，只构建视口附近的行。可见节点索引在数据或展开状态变化时重建；为兼容应用直接编辑公开集合，每次构建仍会检查展开 ID，需要半选状态时也会检查所选 ID。行内容可通过 Build 自定义。

### 懒加载

Lazy=true 且 Children=nil 的节点在可见并展开时调用一次 `OnLoad(id, ticket)`。应用启动后台加载，在 `Window.Update` 中通过 `CompleteLoad(ticket, children, err)` 返回结果。

加载失败会显示 Retry。SetExpanded 折叠该节点或祖先、替换数据、重试或 Close，都会使旧票据失效。CompleteLoad 返回 true 只表示结果已接受，结果仍可能包含错误。

SetNodes 替换数据时会清理不存在的选择和展开 ID；空根列表显示 EmptyText。

### 从旧 Outline 封装迁移

| 需要的行为 | 当前写法 |
| --- | --- |
| 多选 | 显式设置 `Multiple: true` |
| 替换式多选 | `SelectionBehavior: nui.SelectionReplace` |
| 点击只选择，不展开 | `ExpandOnClick: nui.Set(false)` |
| 监听选择 / 展开 | 分别用 OnChange / OnToggle |
| 查询可见行 | 用 `state.Rows()`、`Item()`、`Depth()`，不再用 `state.Outline` 上的同名方法 |

Outline.List 保留滚动状态和 Handle，Selected、Selection、Label 由 Tree 管理。当前不提供拖拽搬移或完整业务权限模型；辅助技术展开动作的限制见[兼容性说明](compatibility.md#无障碍与动态效果)。

## 样式部件索引

| 组件 | 可用部件 |
| --- | --- |
| Form | Root |
| Listbox | Root、Label（组标题）、Item、EmptySlot |
| SelectMenu | Root、Trigger；面板通过嵌入的 ListboxOptions 和 Popup 配置 |
| InputDate / InputDateRange / InputTime | Root、Editor、Segment、Leading、Trailing |
| PinInput / InputRating / Calendar | Root、Item |
| Table | Root、Item、EmptySlot |
| Tree | Root、Item、Leading、Label、Trailing、Indicator、EmptySlot |
| ColorPicker / RangeSlider | Root |

Table 和 ColorPicker 的内部原生区域由 Theme.Native() 设置外观；RangeSlider 的轨道和滑块按 Theme 及 Color 绘制，不支持单独覆盖 Indicator / Thumb 部件。未公开的部件无法通过自行填写部件名来修改。全局 WithComponent 和实例 UI / Style 的优先级见[主题文档](foundations.md#样式覆盖)。

## 示例

| 命令或页面 | 内容 |
| --- | --- |
| Gallery 的 Advanced forms、Dates & color、Task browser | 专用输入、日期、颜色和数据交互 |
| `go run ./examples/settings` | Project setup 表单和可取消提交 |
| `go run ./examples/data-browser` | 搜索、筛选、Table / Tree、按 ID 编辑和 Toast |

InputTags、PinInput 文本模式、SelectMenu 和 ColorPicker 的文本输入受已知 IME 取消问题影响，详见[兼容性说明](compatibility.md#输入法取消)。
