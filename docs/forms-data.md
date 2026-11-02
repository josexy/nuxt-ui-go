# 表单、日期与数据组件

提供 Form、复杂输入、日期、颜色、Listbox、Table 和 Tree，以及 CheckboxTri、RangeSlider、QueryState。InputTags、PinInput 文本模式、SelectMenu 和 ColorPicker 的文本编辑受 [输入法取消问题](compatibility.md#输入法取消) 影响。

## 生命周期与主题

组件入口接收 `c *ui.Context`、`t nui.Theme`，返回当前构建的 `ui.Element`。调用方按窗口持有状态，不复制已使用的状态，不保存 Element/Context。`FormTicket`、`QueryTicket`、`TreeTicket` 带实例身份和版本，不能跨实例完成，也不能把旧结果应用到新编辑。后台工作只处理数据快照，通过 `Window.Update` 发布；窗口关闭时取消调用方工作。

- InputNumber/InputDate/InputTime 继承 Input 的五种尺寸、outline/soft/subtle/ghost/none、颜色、只读、禁用、前后图标、错误与同步 Style 回调。
- 其余组件使用各自的原生配方，支持明暗主题、密度/字体比例、实例 `UI`/`Style` 和全局 `WithComponent` 的 **已列出 slots**；尺寸与变体支持范围见各组件 Options。
- Form：Root；Listbox：Root/Label（组标题）/Item/EmptySlot；SelectMenu：Root/Trigger，选择面板通过嵌入 ListboxOptions 和 Popup 定制。
- PinInput：Root/Item；InputRating：Root/Item；Calendar：Root/Item；Table、Tree：Root/Item/EmptySlot；ColorPicker、RangeSlider：Root。
- Table、Tree、ColorPicker、RangeSlider 的内部原生区域由 `Theme.Native()` 控制，未暴露的原生部件不能通过不存在的 slot 覆盖。Table 单元格可完全自定义。

## Form

`FormState` 保存命名字段、初始值、dirty/touched、可见性、错误和持久 Handle。`FormFieldOptions.Name/Value` 向包围的 Form 注册一个明确序列化的字符串值。Name 必须非空且在该 Form 的单次构建中唯一；一个注册字段对应一个编辑值。

```go
// 保存在窗口状态中：form nui.FormState; name string
nui.Form(c, theme, &form, nui.FormOptions{
    Validate: func(v nui.FormValues) nui.FormErrors {
        if v["name"] == "" { return nui.FormErrors{"name": "请输入名称"} }
        return nil
    },
    OnSubmit: func(values nui.FormValues) { save(values["name"]) },
}, func(t nui.Theme) {
    nui.FormField(c, t, nui.FormFieldOptions{
        Name: "name", Value: name, Label: "名称", Required: true,
    }, func(t nui.Theme) { nui.Input(c, t, &name, nui.InputOptions{}) })
    nui.Button(c, t, "保存", nui.ButtonOptions{}).OnClick(form.Submit)
})
```

`FormValues`/`FormErrors` 都是 `map[string]string` 的命名类型；不是反射 Schema。跨字段校验读取整个 Values。控件的非法编辑文本/错误也会阻止提交，不会悄悄提交旧的有效值。字段首次注册形成 dirty 基线，失焦或 `Touch(name)` 标记 touched；Submit/Validate 标记当前字段 touched，并聚焦构建顺序中的首个错误。同步校验由 Submit 或显式 Validate 触发，不自动在每次输入时运行昂贵业务校验。

隐藏/禁用字段保留模型与 dirty/touched，但不进入 Values、校验失败判定或首错聚焦。隐藏字段错误仍可保留，重新显示后继续检查。`Reset()` 将**当前值**设为新基线并清理错误，不回写应用模型；恢复业务初值由应用执行。

异步校验：在 UI 线程调用 `ticket := form.BeginAsync()`，只把 `ticket.Values` 的快照交给校验任务；结果在 UI 线程用 `CompleteAsync(ticket, errors)` 应用，返回 true 才表示当前结果有效且无错误。改值、改原始编辑文本、改变字段可见性、Cancel 或新请求都会使旧票据失效。`CompleteAsync` 不自动再次调用 OnSubmit；应用显式完成保存。`Cancel()` 使结果失效，实际 I/O 的 context 由应用取消。自定义控件需要把 `form.Field(name).Handle` 绑定到真实可聚焦控件；内置输入、选择、单值开关等会自动关联，组合组只绑定第一个控件。

设置页的 Project setup 展示两步注册、条件隐藏、跨字段日期/预算错误、首错聚焦和可取消的本地待确认提交。Confirm project 是可控结果模拟，不发起网络请求。

## 数值、标签、PIN、评分

| 入口 / 状态 | 值与行为 | 边界 |
| --- | --- | --- |
| `InputNumber` / `NumberState` | `*Optional[float64]`；Text 与有效数值分离；空值、负号中间态、有限数值、Min/Max、Step（默认 1）、上下箭头及按钮；horizontal/vertical 布局；Enter 恢复/格式化最后有效值 | 小数点为 `.`，无货币/区域格式；NaN/Inf 模型及非法边界拒绝；不把 float64 当精确货币类型 |
| `InputTags` / `TagsState` | `*[]string`；Query 独立；Enter、空输入 Backspace 删除末项、标签上的 × 删除、trim/去重、MaxTags；标签为 Nuxt 的 Elevated 底加 Accented 内描边小芯片，字号随字段尺寸，只有 × 能删除；设置 `AddOnPaste: nui.Set(true)` 启用批量粘贴 | 批量粘贴默认关闭；开启后按 Separators 分隔（默认与 Nuxt 一样只有 `,`；需要分号或换行时设 `Separators: ",;\n"`），保留已有草稿；按大小写区分去重；不提供拖拽排序与远程实体匹配；受已知输入法取消问题影响 |
| `PinInput` / `PinState` | `*string`；默认 5 格文本，Numeric=true 为 ASCII 数字；1–32 格、按 rune 截断、退格、遮罩、ReadOnly、Placeholder、Separator/SeparatorEvery、OnChange/OnComplete | 一个真实编辑器配多个可视槽位，Tab 只进入一次；长度满且值变化才触发完成；不做短信读取或自动提交 |
| `InputRating` | `*int`，默认 Max=5、0 表示未评分；整数星级、只读/禁用、方向键/Home/End/空格 | 不支持半星、任意精度或拖拽评分 |

InputNumber、InputDate、InputTime 和 Numeric=true 的 PinInput 使用固定格式并关闭 IME；PinInput 文本模式及其他普通文本入口启用 IME。PIN 可视槽位跟随编辑器光标显示焦点，编辑、粘贴和撤销由单个原生编辑器承担。放在有 Error 的 FormField 中时，各槽位改用错误色边框。

InputNumber/InputDate/InputTime 的文本编辑器与 Input 共用鼠标、Tab 聚焦高亮。InputTags 的外框使用字段变体配方：编辑器聚焦或标签按钮键盘聚焦时显示强调色和外圈，支持 Color/Variant 与错误色；None 不显示焦点外圈。

`InputNumberOptions.Orientation` 使用 `nui.Orientation`：零值 `nui.Horizontal`（默认）或 `nui.Vertical`。对应 Nuxt UI 4.12.0 的 InputNumber 方向：horizontal 为左侧减号、居中数值、右侧加号；vertical 为数值靠左、右侧上增下减的 chevron。RTL 下按逻辑 start/end 镜像。按钮使用 link 外观，颜色和尺寸继承字段；两种方向保持相同字段高度。未设置 Orientation 时使用水平布局。

```go
// 按窗口持有 value、numberState 和 orientation。
nui.InputNumber(c, theme, &value, &numberState, nui.InputNumberOptions{
    InputOptions: nui.InputOptions{Label: "数量"},
    Orientation: nui.Vertical, // nui.Horizontal 为默认值
    Min: nui.Set(0.0), Max: nui.Set(100.0), Step: 1,
})
```

方向变化只调整布局，编辑器和增减按钮保持 key 与父级，保留有效值、未完成草稿、选区、撤销历史及当前焦点，不触发 OnValueChange。点击方向选择器会按普通控件交互把焦点移到选择器，组件不会主动抢回焦点。两种方向均支持 Min/Max、只读、禁用、键盘 ↑/↓、前后图标和清空按钮，操作名称随中英文切换；水平布局的文本在可编辑区域居中。

`Style` 接收 root/editor/leading/trailing；这些 leading/trailing 是调用方图标，增减按钮没有独立的公开 slot。组件先预留按钮空间，最后调用实例 Style；定制 root padding 时需保留增减按钮的空间。Gallery 的 **Advanced forms → Orientation** 可在同一个 Seats 字段上切换两种布局。

六位数字验证码使用 `PinInputOptions{Numeric: true, Length: 6}`；文本模式启用 IME，受已知输入法取消问题影响。InputRating 五档星标尺寸为 12/16/20/24/28 DIP，间距 2 DIP。

PIN 点击已有字符会选中该位，随后输入替换该位、Backspace 删除该位；点击空槽或槽间空隙将光标放在已有内容末尾。槽位不会增加 Tab 停靠点，文字缩放与遮罩不影响点击映射。InputRating 横向左右箭头随 LTR/RTL 镜像，纵向 ↓ 下一项、↑ 上一项，忽略另一轴。与 Nuxt 一样，方向键到首尾停止、不循环，← 不会把评分降到 0（只有清零能回到 0）；空评分按 → 得 1。空格作用于当前星：空评分得 1，Clearable 时清零当前评分；回车不做处理。Home/End（及 PageUp/PageDown）直接设为 1/Max：Nuxt 只把焦点移到首尾的星，需再按空格确认，这里保持滑块的按键习惯。键盘焦点环画在当前星上（空评分时为第 1 颗），组件按星星大小排布，不随容器拉伸。可交互时，悬停的星放大到 1.1 倍；Hoverable 的预览划过星间空隙时保持，指针离开组件或点击清零后消失。放在出错的 FormField 中时，实心星和焦点环使用 Error 色。只读评分仍可通过 Tab 聚焦，按键无效（Nuxt 会跳过只读评分）。

`CheckboxTri` 使用 `CheckState`：Unchecked / Checked / Indeterminate。混合态通过原生 Mixed 语义暴露；首次激活混合态变为 Checked，之后在 Checked/Unchecked 间切换。父子级联由业务计算；普通二态选择使用绑定 `*bool` 的 `Checkbox`。

`RangeSlider` 接收两个 `*float64`，默认 0–100、Step=1；拇指交叉后排序并转移活动焦点，保持 low≤high，不满足最小间距的移动被拒绝。外部越界值在构建时归入合法范围。支持 Orientation、Inverted、每拇指 Tooltip、MinStepsBetweenThumbs、Color/Size、实时 OnChange 和结束 OnCommit；使用公开 Range/Step 与作用域快捷键，两个拇指保留独立焦点和实际可访问值。放在 FormField 中时，点击字段标签聚焦第一个拇指，字段的描述和错误也落在该拇指上；两个拇指保留各自的"最小值/最大值"名称。纵向元数据限制及参数说明见 [参数与默认值](options.md)。

## Listbox、SelectMenu 与异步搜索

`ListboxItem` 包含 ID、Label、Description、Group、Disabled、Icon。ID 不依赖标签；空/重复 ID 报错。选择使用 `*[]string`，空切片表示空选；默认 SelectionToggle：单选与多选均可再次激活以取消；SelectionReplace 表示激活替换，SelectMenu 的单选模式使用后者。数据暂时不包含的已选 ID 保留（用于异步搜索），清理失效业务值由调用方负责。

`ListboxState` 持有 Active 和列表状态；相邻同组项目有不可选择的组标题。行使用[选项行配方](inputs.md#select-与-inputmenu)，MD 下约 32 DIP 高，有描述时增加一行；根圆角 8，键盘聚焦时整个列表显示组件颜色的焦点环；空状态居中。Listbox 还没有公开的 Size/Color 字段：尺寸取外层 FormField，其次 `WithComponent(ComponentListbox, …)` 的默认值；在 SelectMenu 中使用菜单的 Size 和 Color。纵向使用虚拟列表和 Up/Down；横向按 ItemWidth 裁剪可见项目，使用 Left/Right，组标题也占一个 ItemWidth。两种方向均只构建可见项目及少量预留项目。Home/End 到两端，Enter/Space 激活，字符前缀检索，高亮随稳定 ID 恢复；被删除的高亮回退到第一个可用项。

状态缓存数据快照、ID 索引、分组及选择集合；未变更数据不再逐帧重建这些结构。调用方原地修改项目或选择切片也会更新缓存，不需要显式版本号。每次构建仍线性比较输入与快照，数据变更时重新验证 ID 和整理分组；字符前缀检索仍可能遍历全部项目，不宣称整个渲染过程为常量时间。

`SelectMenuState` 保存 Open/Query/Listbox。SelectMenu 的 Size 和 Color 同时作用于触发器、弹层里的搜索框和列表。弹层与 Nuxt 的 combobox 一样是平铺的：无外边距，搜索框为 None 变体，下方一条分隔线，列表无边框、行内缩 4，默认最大高 240，列表高度随行数收缩。未选择时占位文字为 Dimmed 色；Subtle 变体悬停时加深底色。Clearable 在有选择时把触发器尾部的 chevron 换成清空按钮。OnCreate 的创建项是列表中的一行，方向键可以选到，Enter 或点击发出创建请求，不选中任何项。触发器的 ↑、↓ 都可打开；默认聚焦搜索，SearchInput=false 时聚焦列表。搜索框 ↑↓ 把焦点移至列表，Escape 使用 Popover 关闭并返回触发器。单选后关闭，多选保留打开；默认选择或关闭后清空搜索词，两个 ResetSearchTerm 选项可分别关闭。Clearable 清空，OnCreate 只发送创建请求并保留搜索词/打开状态；不会替应用添加记录。默认只搜索 Label；FilterFields 可指定 label/description/value(id)/group。默认无匹配结果才显示创建；CreateWhenAlways 提供非完全相同 ID 的创建项，CreateAtTop 调整位置。Loading/Error/EmptyText 显示状态；ExternalFilter=true 时不再执行本地 contains 筛选。

SelectMenu 通过 ID 索引解析已选标签，并缓存拼接后的触发器文本；临时缺失的已选项保留上次标签。数据、查询或 FilterFields 改变时才重新执行本地过滤。首次处理或更新大量已选标签仍有线性工作，远程搜索应使用 ExternalFilter 与 QueryState。

```go
// 均为窗口模型；不要在 worker 中修改 queryState。
ctx, ticket := queryState.Start(parentContext, query)
// worker 使用 ctx 执行请求；Window.Update 的结果回调中：
if queryState.Complete(ticket, fetchedItems, err) {
    // 下一次构建将 Items/Loading/Error 传给 SelectMenu。
}
// 离开视图或关窗：queryState.Cancel()
```

`QueryState[T]` 适用于 SelectMenu 也适用于 InputMenu 的应用侧数据源；Start 取消前一个 context，Complete 克隆结果切片并拒绝旧/跨实例票据。库不创建网络 goroutine，也不预设 API 或 debounce 时间。

Listbox 支持同步 Build、SelectionBehavior、Loop 和 HighlightOnHover；`SelectionReplace` 多选模式支持 Shift 连续范围选择。选择角色和所选行通过原生语义提供，但 MyGo 未开放独立的 multiselectable 标志，完整多选读屏体验尚未验证。模糊匹配与远程分页由应用处理。

## Calendar、InputDate、InputTime

`Date{Year, Month time.Month, Day}` 为 1–9999 年公历民用日期；Valid/Compare/AddDays/AddMonths/String/ParseDate 可用于业务边界。AddMonths 将 1 月 31 日移至目标月最后一天。DateFromTime 读取传入 time.Time **所在 Location 的年月日**；没有 UTC 来回转换。业务若需要瞬时值，必须自己选择时区并处理夏令时。

`Calendar` 接收 `*DateSelection{Start, End Optional[Date]}` 与 `CalendarState`。Range=false 单日；Range=true 首次点击产生未完成范围，第二次点击规范化反向端点，完整范围后再点开始新范围。Min/Max/IsDateDisabled 限制可选日期，跨越禁用日的范围被拒绝。WeekStart 可配置；Today 可注入确定性日期。导航：左右按日、上下按周、PageUp/PageDown 按月、Home/End 到月首/月末、Enter/Space 选择；不可选日期可以获得网格高亮但不能激活。单月六周网格，标题与星期使用 MyGo i18n 日期格式化，导航文案随中英文切换；左右方向键已随 RTL 镜像；其他历法尚不支持，完整读屏体验尚未验证。

首次构建可显式设置 CalendarState.Month/Active。之后外部更新 Start 会同步高亮及显示月份；Start 不变而 End 更新时跟随有效 End。外部赋值不触发 OnChange；清空选择保留浏览位置。普通翻月和组件内部完成范围选择不会在后续构建中跳回范围起点。

`InputDate` 接收 `*Optional[Date]` 与 `DateInputState`，ISO `YYYY-MM-DD` 文本与单日 Calendar 共用业务值。非法/未完成文本保留在 Text/Error，最后有效值不变；空文本清空，弹层选择回写文本并关闭。HideCalendar 对单日与范围输入均生效，隐藏触发器并关闭现有面板。范围输入使用同组件的 `InputDateRange` 入口、`DateRangeInputState` 和 `*DateSelection`：两个 ISO 编辑器共享范围面板；起点改到终点之后会清空终点，终点不能早于起点，第二次日历点击完成范围后关闭。作为一个 FormField 使用时错误聚焦到起点；需要独立端点错误聚焦时分别注册两个 InputDate。

`ClockTime{Hour,Minute,Second}` 为不含日期、时区的墙上时间；小时 0–23。`InputTime` 使用 `*Optional[ClockTime]` 与 `TimeInputState`，支持 24 小时或 `hh:mm AM/PM`、可选秒。`TwelveHour` 是 `Optional[bool]`：未设置时与 Nuxt 一样随语言决定（en 为 12 小时，zh-Hans 为 24 小时），并优先采用用户系统区域格式里的小时制（`Context.Region().HourCycle`）；`nui.Set(true/false)` 固定格式。12/24 格式切换（包括切换语言）保留值并按新格式重写；正在输入的无效草稿和错误提示保持不变。隐藏秒后编辑小时/分钟仍保留原秒。无时区选择器、夏令时推算、分段旋钮或时间弹层。

尚不支持日历多选、年/月选择面板、多月并排、周序号、其他历法及时间范围输入。

## ColorPicker

`ColorPicker(c,t,*ui.Color,*ColorPickerState,ColorPickerOptions)` 提供原生 ColorPicker 的饱和度/亮度平面、色相、透明度、HEX/HEXA 与色板，以及 R/G/B/A 数值输入（0–255）。修改任一表示后从同一个 `ui.Color` 同步，其余通道不丢失。平面和滑条保留原生键盘行为。RGB 小数会按 uint8 截断，建议按钮按整数步进。

原生色板/内部 HEX 字段仅支持 Native 主题，不支持任意 slot builder；HEX 的编辑事务仍受共享 IME 缺口影响。不支持 OKLCH/HSL 编辑、屏幕取色、色彩管理或完整 Nuxt ColorPicker API。

## Table

`TableState[T].SetRows(rows,key)` 在数据变更时克隆行值并验证非空唯一 ID；不要在每次渲染里重复调用。泛型 T 内部如含指针/切片仍由调用方遵守所有权。Selection 是 `ui.Selection[string]`；数据替换会移除不再存在的已选 ID。Sort 为 `ui.SortOrder`；列宽和顺序位于 `state.List.Columns`，可显式持久化。

`TableColumn[T]`：ID/Title、Width/MinWidth/MaxWidth、Align、Fixed、Value、Compare、Cell。Compare 提供本地稳定排序，Cell 同步接收 Context/Theme/行值/是否选中。Row/RowID/Len 读取当前视图；CellBuilds 用于测量当前构建的单元格调用数。排序后编辑一定按 RowID/Selection 中的 ID 回查模型。

| 能力 | 当前范围 |
| --- | --- |
| 排序、选择、列宽/重排、滚动虚拟化 | 公开 Table 行为；排序按列 Compare，稳定 ID 选择；宽度/顺序由原生 TableLayout 保存 |
| 筛选、分页、服务端数据 | 应用准备当前 rows 后 SetRows；data-browser 给出搜索、团队/状态筛选、100 条分页。服务端请求/总量/缓存由应用管理 |
| 空状态、单元格 | EmptyText、EmptySlot、Cell builder；加载/失败状态可在表格外组合 |
| 固定列 | `Fixed` 仅禁止移动/调整宽度，**不是冻结滚动列**；sticky/frozen columns 未提供 |
| 多层表头、整行/表头任意 builder | 未提供 |
| 行展开、分组聚合、完整 TanStack 语义 | 未提供 |
| 可变高度 | 底层 List 有测量机制；当前面向稳定行高，任意动态复杂单元格需由应用验证 |

虚拟化限制可见单元格的构建量；数据生成、ID 验证、排序与应用筛选仍随数据量变化。

## Tree

`TreeState.SetNodes([]TreeNode)` 建立 ID 模型；TreeNode 含 ID/Label/Icon/Children/Lazy，空/重复 ID 拒绝。复用公开 `ui.Outline`，不复制上游树引擎；`state.Outline.Open` 管展开，`state.Selection` 管 ID 选择，保留原生左右展开/父子移动、上下选择、Ctrl/Shift 多选与虚拟化。自定义图标为 Lucide，原生展开指示符仍由控件绘制。

Lazy=true 且 Children=nil 的节点在展开时只调用一次 `OnLoad(id,ticket)`；应用通过 Window.Update 调用 `CompleteLoad(ticket,children,err)`。错误行有 Retry 按钮；折叠、替换数据或重试使旧票据失效。CompleteLoad 返回 true 表示结果已应用（包括错误），不表示成功加载。空根显示 EmptyText。SetNodes 替换会清理不存在的选择/展开 ID。

不提供拖拽搬移、级联 Checkbox、节点级禁用或完整业务权限模型。懒加载检查只遍历已展开节点与进行中的请求，不扫描所有已加载节点；折叠后、下次构建之前到达的结果也会被拒绝。底层 Outline 仍在构建时扁平化根节点与展开的子节点，成本随这些节点数增长，虚拟化限制的是实际行构建数量。

## 示例

- Gallery 的 Advanced forms、Dates & color、Task browser 展示输入和数据交互。
- `go run ./examples/settings`：Project setup 表单与可取消提交。
- `go run ./examples/data-browser`：任务搜索、筛选、Table/Tree、按 ID 编辑与 Toast。
