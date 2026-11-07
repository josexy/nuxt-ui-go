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

空名称或尚未注册名称的错误同样会阻止提交，可通过 `form.Errors()` 获取并显示全局错误。无效提交调用一次 `OnError`，有效提交调用一次 `OnSubmit`；回调收到独立快照。显式 Validate、SetErrors 和 CompleteAsync 不自动触发这两个事件。

### 校验、隐藏和重置

| 操作或状态 | 行为 |
| --- | --- |
| 首次注册字段 | 当前值作为 Dirty 比较的初值 |
| 失焦或 `Touch(name)` | 将字段标记为 Touched |
| `Submit()` / `Validate(...)` | 执行校验，标记当前字段为 Touched，并聚焦构建顺序中的首个错误字段 |
| 隐藏或禁用字段 | 保留模型、Dirty、Touched 和错误，但不进入 Values、提交校验或错误焦点选择 |
| `Reset()` | 取消旧请求、清理错误和交互状态，并以下次构建的当前值作为新初值；不改写应用绑定的数据 |
| `Errors()` / `SetErrors(errors)` | 获取独立错误快照 / 替换校验错误；非法编辑草稿仍优先且会阻止提交 |
| `ClearErrors(names...)` | 清除指定名称的校验错误；不传名称时清除全部，不会把非法草稿改成有效值 |

非法的数字、日期或其他编辑草稿也会阻止提交，避免提交上一次有效值。业务校验由 Submit 或显式 Validate 触发，不会在每次输入时自动运行。

隐藏字段再次出现时继续检查已有错误。需要恢复业务初值时，应用先恢复绑定数据，再按需要重置表单状态；单独调用 Reset 不会恢复输入内容。

同步 OnSubmit 返回且未启动异步工作，或 CompleteAsync 成功时，当前可提交字段的 Dirty 清零。Touched 表示曾经失焦或参与校验；它与 Nuxt 的事件集合并非完全相同。嵌套 Form 各自管理字段，不自动连接父表单的值和校验；跨表单提交由应用协调。

### 异步校验

1. 在 UI 线程调用 `ticket := form.BeginAsync()`，获得当前字段快照。
2. 后台任务只读取 `ticket.Values` 并执行校验。
3. 在 `Window.Update` 回调中调用 `form.CompleteAsync(ticket, errors)`。
4. 返回 true 表示结果仍有效且无错误，再由应用执行保存。

CompleteAsync 不会再次调用 OnSubmit。字段值、原始编辑文字、可见性变化，或 Cancel、新请求，都会使旧票据失效。`Cancel()` 只让表单拒绝旧结果，实际 I/O 的 context 仍由应用取消。

LoadingAuto 默认 true：BeginAsync 后，表单内的输入和按钮禁用，字段快照仍保留。需要在等待期间继续编辑时设置 `LoadingAuto: nui.Set(false)`，编辑会使旧票据失效。两种模式都忽略重复 Submit；禁用期间的提交也不会在重新启用后执行。取消和完成按钮应放在 Form 外部。

Validate、SetErrors、ClearErrors 和 Reset 会取消旧票据，避免较旧结果覆盖最新校验或手工错误。卸载组件或关闭窗口前调用 Cancel；状态对象本身无法自动获知所有卸载路径。后台只能处理 ticket.Values，在 Window.Update 中接收结果。

Root 是唯一样式部件，默认纵向间距 16 DIP；字体和文字颜色在正文构建前设置。这里没有 HTML form 提交、schema 转换、自动输入防抖校验或 Nuxt 嵌套路径处理。

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

Step 默认 1，必须为有限正数，0 使用默认值。StepSnapping 默认 true：例如 Step=1 时从 18.98 向上到 19，向下到 18；关闭后直接加减步长。Min 是步长网格的起点，未设置时使用 0。空值或不能解析的草稿首次增减得到 StartingValue，未指定时使用 Min 或 0；显式起始值只限制到上下界，下一次操作才按步长对齐。

上下方向键增减一步，PageUp / PageDown 增减十步；提供 Min / Max 时，Home / End 到达对应可用边界。鼠标按下立即改变一次，持续按住 400 ms 后每 60 ms 重复；松开不会额外改变一次，禁用或卸载后停止。输入框聚焦且纵向滚动量较大时，滚轮默认改变数值，向下增加；DisableWheelChange 禁用，InvertWheelChange 反转方向。

有效编辑实时更新模型，Enter 对齐有效值或用绑定数值恢复非法草稿；失焦保留草稿。AllowInvalid 允许有限数值超出上下界，并保留手工输入，不自动标为范围错误；业务校验由应用处理。数字使用 ASCII 十进制和科学记数法，不接受 Go 的十六进制或下划线写法，不启用 IME；NaN / Inf 模型、起始值和非法边界会被拒绝。组件不提供地区、货币或 Intl 格式，float64 也不适合作为精确货币类型。

| Orientation | 布局 |
| --- | --- |
| Horizontal（默认） | 左侧减号、居中数值、右侧加号 |
| Vertical | 数值靠左，右侧上下排列增减箭头 |

两种布局字段高度相同，按钮默认继承字段颜色和尺寸，使用链接样式，RTL 下按起始侧和结束侧调整。Increment / Decrement 默认显示按钮，可通过 Optional 显式关闭；IncrementOptions / DecrementOptions 配置按钮，IncrementIcon / DecrementIcon 覆盖图标。显式图标不会因为换成纵向而旋转。单侧 Disabled 只禁用该按钮，键盘和滚轮仍按字段规则工作。

FocusOnChange 默认 true，增减后回到输入框；false 允许按钮保留焦点。方向和显示切换保留编辑器身份、草稿及选区；外部赋值或点击其他控件不会主动聚焦输入框。Style 接收整个字段、编辑器和前后配件；UI.IncrementSlot / DecrementSlot 配置增减按钮，自定义按钮 Style 最后执行。按钮保留原生 Tab 停靠点，没有 Web spinbutton 的完整角色映射。

### InputTags

InputTags 绑定 `*[]string`，TagsState.Query 保存尚未添加的草稿。Enter 添加标签，标签和编辑器在同一行内自动换行。光标在草稿开头时，Backspace 先选中末项，再次 Backspace / Delete 删除；左右方向键移动选中项，已选中时 Home / End 到首尾。删除按钮使用 DeleteIcon，默认 ×，删除后输入框恢复焦点。

组件去除候选的首尾空白，再调用 ConvertValue 转换字符串，最后按大小写区分去重并检查 MaxTags。DisplayValue 只改变标签及删除按钮文字，值和稳定身份不变。OnTagsChange 返回切片副本；OnAddTag、OnRemoveTag、OnInvalid 分别报告每个新增、删除或拒绝的候选。一次成功批量粘贴最多通知一次 OnTagsChange，单项事件仍分别报告。

默认不批量处理粘贴。设置 `AddOnPaste: nui.Set(true)` 后，按 Separators 分割，默认分隔符只有逗号；需要分号或换行时设置 `Separators: ",;\n"`。批量粘贴保留已有草稿。更多添加和重复策略见[参数说明](options.md#开关标签和评分)。

标签使用 Elevated 背景和 Accented 内边框，选中标签使用 Accented 背景；字号随字段尺寸，长标签可以换行。编辑器或标签按钮聚焦时，外框显示强调色与外圈；错误时使用错误色，None 不显示焦点外圈。Root 是整个字段，Body 是换行区域，Item / Label / Close / Indicator 对应标签容器、文字、删除按钮与图标。Style 的 root 现在也指整个字段。当前不支持对象模型、正则分隔符、拖拽排序或远程实体匹配。

### PinInput

PinInput 绑定 `*string`，另保存 PinState。默认是 5 格文本；六位数字验证码可使用 `PinInputOptions{Numeric: true, Length: 6}`。Length 范围为 1–32，按 rune 限长，Numeric 只接受 ASCII 数字。

Highlight 默认为 false，开启后全部格子显示当前颜色边框，字段错误优先。UI / Style 支持 Root、Item、Label 和 SeparatorSlot；Root.Gap / Padding 与分隔符字体、字号、内边距参与整体尺寸计算，显式 0 有效。原始 Style 回调直接改变宽高时，由应用协调父级尺寸。

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

UI / Style 支持 Root、Item、Indicator、IconSlot 和 EmptyIconSlot。未评分图标作为底层，评分或悬停范围内叠加 Indicator 和 IconSlot；两种图标不同时，底层仍可能出现在叠加图标周围。Indicator 的前景色由 IconSlot 继承，后者可以单独覆盖。Root 的 Gap / Padding 参与整体尺寸计算；Item 的 Shadow 可覆盖默认焦点外圈。Neutral 的已评分图标使用 Highlighted，焦点外圈使用 InvertedBackground。只读保持正常颜色和默认光标。

Max 接受 1–20，0 使用默认 5；外部评分或 Max 变化会静默限制到 0–Max，OnChange 只报告实际用户修改。Hoverable、Clearable 默认 false。Nuxt UI 的 step 可表示半星等小数评分，本库的整数绑定没有对应字段；也不提供任意 item 内容构建器、HTML 表单字段和网页 radio 节点。

### CheckboxTri

CheckboxTri 使用 CheckState：Unchecked、Checked、Indeterminate。混合态向辅助技术提供 Mixed 状态，首次激活后变为 Checked，之后在 Checked / Unchecked 间切换。

`IndeterminateIcon` 设置部分选中时的图标，未设置时使用减号；`CheckedIcon` 设置普通选中图标。隐藏指示器时，只把显式 `CheckedIcon` 显示在标签上方，不使用部分选中图标，也不补入默认图标。`OnChange` 报告是否选中，`OnStateChange` 报告完整 CheckState；一次有效操作分别通知一次，修改图标本身不触发这些回调。

父子选中关系由业务计算。只需要开关状态时，使用绑定 `*bool` 的 Checkbox。

### RangeSlider

RangeSlider 接收两个 `*float64`，默认范围 0–100、Step=1。两个滑块交叉后会排序并转移活动焦点，保持 low≤high；不满足最小间距的移动会被拒绝。应用传入的越界值会在构建时限制到合法范围。

常用选项包括 Orientation、Inverted、Tooltip、MinStepsBetweenThumbs、Color、Size、OnChange 和 OnCommit。两个滑块有独立焦点，并向辅助技术提供各自的实际值和最小值/最大值名称。

Indicator.Background 和 Thumb.Background 与单滑块一致生效。小数步长去掉运算尾数，最小间隔比较使用随间隔大小变化的容差；极小步长也不能让两端重合后再推移另一端。禁用父容器会取消当前拖拽，不触发提交。

放在 FormField 中时，点击标签聚焦第一个滑块，字段描述和错误也关联到它。方向和键盘说明见[参数文档](options.md#滑块和内容切换)，纵向辅助技术限制见[兼容性说明](compatibility.md#无障碍与动态效果)。

## Listbox、SelectMenu 与异步搜索

### Listbox：直接展示可选择列表

ListboxItem 的 ID 必须非空且唯一。Label、Description 提供文字，Group 为相邻项目添加组标题；Kind 可以是 ListboxOption、ListboxLabel 或 ListboxSeparator，结构行也需要 ID。选择值为 `*[]string`，空切片表示未选择；ListboxState 保存查询、活动项、滚动和持久 QueryHandle。

默认 SelectionToggle，单选和多选均可再次激活以取消。SelectionReplace 在激活时替换选择；多选时还支持 Shift 连续范围选择。暂时不在数据源中的选中 ID 会保留，便于异步搜索；永久失效的 ID 由应用清理。

Icon、Avatar、Chip 按此顺序选择前导内容，与文字首行对齐。SelectedIcon 覆盖默认勾选图标。单项 Options 提供 UI、Style 和 OnSelect；回调接收改值前的 ID 与选中状态，可调用 PreventDefault 阻止本次选择及组件 OnSelect。选项行外观见[输入组件](inputs.md#选项行的外观)。MD 行高约 32 DIP，有描述时增加一行。

Size / Color 默认 MD / Primary，实例选项优先于组件主题，尺寸还可继承 FormField。Highlight 持续显示颜色边框；键盘聚焦时显示外圈，Neutral 使用 InvertedBackground。嵌入 SelectMenu 时，列表跟随菜单的尺寸和颜色。

Filter 默认 false，开启后可通过 FilterInput 配置搜索输入框。SearchTerm 接受持久字符串指针，未设置时使用 State.Query。FilterFields 默认仅匹配 label，也支持 description、value/id、group；显式空列表不匹配非空查询。筛选忽略大小写和组合附加符号，在各组及结构行之间的选项段内，按完全匹配、前缀、包含排序。它不实现完整的地区排序规则。

隐藏筛选框后仍使用保留的查询词；IgnoreFilter 跳过本地筛选。Loading 用图标替换选项并暂停列表操作，筛选框仍可编辑。LoadingIcon 可覆盖图标；减少动态效果时停止旋转。

| 方向或按键 | 行为 |
| --- | --- |
| 纵向 | 原生虚拟列表，↑↓ 移动 |
| 横向 | 按固定 ItemWidth 显示和裁剪项目，←→ 移动；组标题也占一项宽度 |
| Home/End、Enter/Space | 到首尾、激活项目 |
| 输入字符 | 按前缀检索项目 |
| 筛选框中的 Enter | 保持输入焦点并激活高亮项；组合输入期间由输入法处理 |
| 数据重排 / 外部选中值 | 重排保留稳定 ID；初始化或外部选中值变化时优先高亮可用的已选项，否则使用第一个可用项 |

两种方向都只构建可见及少量预留项目。Height 是包含筛选框的总高度，未设置时纵向 224 DIP、横向 64 DIP；横向 ItemWidth 默认 160 DIP。同步 Build 可自定义非交互行内容，Loop 默认 false，HighlightOnHover 默认 true。OnHighlight 只通知用户操作造成的高亮变化，初始化和外部值更新不通知。

### SelectMenu：搜索弹层

SelectMenu 将触发器与弹层内的搜索框、Listbox 组合起来。SelectMenuState 保存 Open、Query、Listbox 及触发器 Handle、搜索框 QueryHandle，Size / Color 同时作用于三个区域。SearchTerm 可绑定外部查询指针；外部赋值不通知 OnQuery 或 OnSearchTermChange。

默认 ↑、↓ 都能从触发器打开，打开后聚焦搜索框；FocusSearch=false、SearchInput=false 或搜索框禁用时聚焦列表。搜索框中按 Enter 直接激活高亮项，↑↓ 进入列表，Escape 关闭并返回触发器。单选后关闭，多选继续打开；应用的选择回调主动移动焦点时保留该目标。

默认筛选与 Listbox 相同；SearchInput 只控制搜索框显示，隐藏后仍使用已有查询。ExternalFilter 或 IgnoreFilter 为 true 时直接使用应用提供的结果。FilterInput 配置搜索框；SearchPlaceholder 保留为简便入口，FilterInput.Placeholder 优先。继承的 Filter 字段不参与 SelectMenu 行为。

Leading、Avatar 配置触发器前导，Trailing 配置尾部箭头，Highlight 强调边框。Loading 使用 LoadingIcon 显示旋转图标，LoadingTrailing 控制位置；同时保留原生加载状态行，已有结果仍可选择。Error 和 EmptyText 分别配置状态文字与空结果提示。ShowTrailing=false 隐藏整个尾部区域。

| 选项 | 行为 |
| --- | --- |
| ResetSearchTermOnSelect / ResetSearchTermOnBlur | 默认均启用，分别控制选择后和关闭后清空搜索词 |
| Clearable / ClearOptions / ClearIcon | 有选择时，用可配置的清空按钮替换尾部箭头；ShowTrailing=false 时隐藏 |
| ResetModelValueOnClear | 默认 true；控制清空按钮是否重置选中值。按钮始终清除查询，删除搜索文字本身不会重置选中值 |
| OnClear / OnOpenChange | 一次用户清空或实际开关变化通知一次；直接修改 State.Open 不发出开关通知 |
| OnCreate | 显示可点击或键盘激活的创建行，只发出创建请求，不自动新增或选中 |
| CreateWhenAlways / CreateAtTop | 默认仅无匹配时显示创建行；可改为没有相同 ID 时显示，并调整位置 |

创建请求保留查询和打开状态，OnCreate 收到去除首尾空白的查询，不自动修改选项或选中值。OnHighlight 在创建行返回查询文本，不暴露内部行 ID。OnChange 只在值变化时通知；OnSelect 包含重新激活已选值。清空与重置查询只有在查询实际变化时才通知 OnQuery / OnSearchTermChange。

未选择时使用 Dimmed 占位文字；Subtle 触发器悬停时背景加深。弹层搜索框默认 None 外观、无前导图标，下方有分隔线；列表无边框、行内缩 4 DIP，默认最大高度 240 DIP，内容少时自动缩短。自定义 Build 或部件高度时可显式设置 Height，自动估算不读取任意自定义布局。Root、Trigger、Leading、Label、Trailing 只配置触发器；列表使用 List、Header、ItemLeadingSlot、ItemLabelSlot、ItemTrailingSlot 等独立部件，面板使用 Popup.UI。

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

Calendar 绑定 `*DateSelection`，CalendarState 保存首个显示月份 Month、键盘活动日期 Active、导航视图 View 和持久 Handle。Range=false 只使用 Start；Range=true 先选起点，再选终点，反向选择会规范化顺序。默认重复选择单日或未完成范围的起点会取消，PreventDeselect=true 保留原值；需要选择同日起止范围时也应开启此项。

ReadOnly 允许聚焦、浏览和翻页，阻止用户改值。OnChange 只在实际改值时通知；外部赋值和配置重建不通知。OnMonthChange、OnViewChange 分别通知用户操作造成的首月与导航视图变化，外部直接修改状态保持静默。

| 选项 | 行为 |
| --- | --- |
| Min / Max / IsDateDisabled | 限制日期；方向键跳过禁用日期，端点必须可选 |
| IsDateUnavailable | 显示删除线，键盘可以浏览，但不能选作端点 |
| AllowNonContiguousRanges | 默认 false；true 允许范围经过不可用日期 |
| IsDateHighlightable | 显式允许范围经过指定的内部日期；不解除端点限制 |
| MaximumDays | 包含两端的范围天数上限，0 不限制；超出上限的候选禁用 |
| FixedDate | 完整范围可固定起点或终点；越过固定端时更新该端，保持日期顺序 |
| Color / Variant / Size | 默认 Primary / Solid / MD；支持 Solid、Outline、Soft、Subtle，尺寸继承字段或组件主题 |
| WeekStart / WeekdayFormat | WeekStart 保留原生默认周日；星期默认窄名称，也可选缩写或完整名称 |
| FixedWeeks | 默认 true，显示六周；false 按当月显示四至六周 |
| NumberOfMonths / PagedNavigation | 默认一个月，最多 12 个；空间不足时换行。PagedNavigation 让月份按钮按显示月数翻页 |
| MonthControls / YearControls | 默认 true，显示日视图的月份和年份按钮；按钮与图标均可单独配置 |
| ViewControl / ViewControlOptions | 默认启用标题按钮，在日、月、年导航视图间切换；月、年单元格只导航，不提交日期 |
| YearsPerPage | 年导航默认每页 12 年，支持 1–120；IsMonthDisabled / Unavailable、IsYearDisabled / Unavailable 配置对应导航视图 |
| PreviousPage / NextPage | 自定义前后页的目标日期；用于月份按钮及月、年视图的翻页按钮，不改变独立的前后年按钮与键盘规则 |
| Today | 固定今天的位置，便于示例或测试；未设置时使用当前本地日期 |

| 按键 | 日视图行为 |
| --- | --- |
| ←→ / ↑↓ | 按天 / 按周移动，左右随 RTL 调整，跳过禁用日期并遵守边界 |
| PageUp / PageDown | 上个月 / 下个月，尽量保留日号 |
| Shift+PageUp / PageDown | 前一年 / 后一年 |
| Home / End | 当前周的首个 / 最后一个可用日期 |
| Enter / Space | 选择或取消当前日期；只读、禁用或不可用日期不会改值 |

月、年视图使用四列网格，方向键按列或行移动，PageUp / PageDown 按年或年份页移动。选择后返回下一层视图，日期值保持不变。只读模式也允许这种导航。

WeekNumbers 按显式 WeekStart 计算：周日采用包含 1 月 1 日的第一周，其他起点采用包含 1 月 4 日的第一周，周一对应 ISO 周序号。这个规则不等同于 Nuxt 按 locale 推导周起始日和周序号。语言变化会更新标题、星期和按钮名称，不会改变显式周设置。

Month 会规范化为月首。外部修改选中起点或终点时跟随有效日期；清空选择保留浏览位置。改变显示月数时保持活动日期可见，9999 年末只显示剩余月份。多月共用一个原生键盘焦点，窄容器按可用空间换行。

BuildDay、BuildMonth、BuildYear 可同步构建单元格内容，BuildWeekday 配置星期内容；默认名称、选择、禁用和焦点仍由容器提供。单元格内容应保持非交互，尺寸由日历决定。BuildHeading 替换默认标题，接收当前月份和视图；自定义控件通过持久 CalendarState 管理状态。构建器和日期谓词可能多次调用，不能在其中直接执行 I/O 或提交操作。

日期和范围仍是公历的完整日期，不提供离散多日期模型或独立的整月、整年值模型。月、年导航视图不代表实现了 Nuxt 的 `type="month" / "year"`；这些值可由应用的 Select / InputNumber 等控件管理。其他历法、时区以及浏览器 DOM 插槽不在此原生模型中。

### InputDate：输入日期

InputDate 绑定 `*Optional[Date]`，DateInputState 保存分段、草稿和可选日历状态。日期采用公历，显示顺序和分隔符随窗口区域设置变化。可用 PlaceholderValue 为首次调整空分段提供参考日期；它不会直接设置业务值，也不替代字符串占位文字。

| 操作 | 行为 |
| --- | --- |
| ←→、Tab / Shift+Tab | 在分段间移动；方向键遵循 RTL，Tab 进入或离开字段 |
| 输入数字 | 达到位数或数值上限后自动进入下一段；年份按四位输入 |
| ↑↓、Home / End | 按 DateStep 循环，或到当前段最小 / 最大值；步长 0 使用 1，不向相邻段进位 |
| Backspace / Delete | 清空当前段；Ctrl/Cmd+A 后删除清空当前日期端点 |
| Ctrl/Cmd+C/X/V | 复制、剪切、粘贴当前端点，也支持系统编辑命令 |

修改年月会将超出月底的日期限制到目标月末。粘贴接受 ISO 日期或当前区域顺序的数字日期。Calendar.Min / Max、IsDateDisabled / IsDateUnavailable 同时限制输入与日历；边界包含两端。外部赋值或配置变化不改写模型，现有值不符合约束时显示错误。

DateInputState.Text 保存 ISO 草稿，未填完的段用 `-` 表示。非法粘贴保留原文和 Error，屏幕上的上一组分段值及最后有效业务值不变；全部日期段清空时才清空业务值。OnChange 报告草稿，OnValueChange 仅在有效单日或空值实际变化时通知一次；外部赋值静默。切换语言保留草稿和焦点。只读允许导航与复制，禁止编辑。

ShowCalendar 默认 false，开启后提供日历按钮和 Popover，选值后关闭。HideCalendar 优先于 ShowCalendar。Calendar 的外观、导航和部件配置用于弹层，Range 由输入函数决定；自定义 Label 保留。单日的 Calendar.OnChange 只报告日历操作，分段编辑使用 OnValueChange。自身、字段或祖先禁用时关闭已打开的弹层。

### InputDateRange：输入起止日期

InputDateRange 绑定 `*DateSelection` 并保存 DateRangeInputState。两端共用外框，可跨端点导航，空间不足时换行。SeparatorIcon 默认为减号，BuildSeparator 可同步替换为非交互内容，SeparatorSlot 控制其样式。

起点改到终点之后会清空终点；终点不能早于起点。MaximumDays、不可用日期及 AllowNonContiguousRanges 等范围规则也用于分段输入。外部起点超出日历上限时显示错误，不再生成相互冲突的日历边界。

OnRangeChange 报告分段或日历引起的实际范围变化。Clearable 一次清空两端，只通知一次；Calendar.OnChange 保留为兼容的范围通知，每次同样只调用一次。范围不调用单日的 OnValueChange。OnChange 仍报告各端点草稿，清空或日历改变两端时可各通知一次；OnClear 是整次清空动作的通知。第二个日历端点选完后关闭弹层。

放在同一个 FormField 中时，错误聚焦到起点首段。需要两端分别注册错误和焦点时，可组合两个 InputDate。

### InputTime：输入时间

InputTime 绑定 `*Optional[ClockTime]`，TimeInputState 保存草稿。TwelveHour 未设置时使用窗口区域小时制，`nui.Set(true)` / `nui.Set(false)` 固定 12 / 24 小时。上下键在当前段循环，小时跨过中午或午夜会更新 AM/PM，A / P 可直接选择上下午。

Granularity 可选 TimeHour、TimeMinute、TimeSecond。未设置时默认显示分钟，Seconds=true 增加秒；显式 Granularity 优先。改变小时制或精度不改写业务值、不通知，隐藏单位仍保留。隐藏正在聚焦的分钟或秒时回到可见段，隐藏上下午时回到小时。

TimeStep 分别设置小时、分钟、秒步长，0 使用 1，负数无效，不向其他段进位。StepSnapping 默认 false，开启后离开非空分段时按 0 起点取最近的有效步长，正好居中时向上取值；12 小时制按完整 24 小时时间计算。只读、禁用、被拒绝的草稿以及仅改变显示配置时不自动对齐。

Min / Max 为包含两端的完整时间边界，比较隐藏的分钟与秒；不支持上下界倒置的跨午夜区间。IsTimeUnavailable 可拒绝指定时间。无效输入保留草稿与最后有效值，外部赋值静默，越界的外部值显示错误。PlaceholderValue 为未填写分段提供首次调整的参考值，小时精度下新值的隐藏单位也可取自它。

Text 与粘贴格式为 `HH`、`HH:MM[:SS]`，12 小时制追加 ` AM` / ` PM`。显示文字和分隔顺序使用区域格式；当前只接受 ASCII 数字和上述规范时间文本，不解析本地化上下午词语。

### 日期时间的样式与限制

日期时间输入共用 Input 的五种尺寸、五种变体、颜色、头像、图标、加载、禁用、只读和 Clearable。前缀图标或前缀加载优先于 Avatar；LoadingTrailing 控制加载位置，显式 Leading 保留前缀加载。Neutral 焦点和 Highlight 使用反转背景色；Highlight 可用于 None。Root 的主题和实例阴影覆盖保留，Style 最后执行。

Root 控制外框，Editor 控制分段文字基础样式，Segment 控制占位、焦点背景和圆角，Leading / Trailing 控制配件。Style 的 editor 是当前排列的首个可聚焦分段，可绑定 Handle；它不是 TextInputBase，不能使用文本选区 API。Placeholder、Password、MaxLength 不适用于分段输入，占位文字由语言目录配置。

Date 表示完整公历日期，ClockTime 表示单个秒精度时刻，不携带其他历法、毫秒或时区。日期与时刻可分别编辑，时间范围可组合两个 InputTime 并由应用校验。分段输入没有撤销/重做；完整读屏体验尚未验证。Calendar 的月年视图用于导航，不改变日期值类型。

## ColorPicker

ColorPicker 绑定 `*ui.Color`，并保存 ColorPickerState。饱和度/亮度平面、横向色相与透明度滑条、HEX / HEXA、色板以及 0–255 的 R/G/B/A 输入共用这个颜色。应用负责初始化；没有像 Vue defaultValue 那样把零值自动替换为白色。

Size 默认 MD。XS / SM / MD / LG / XL 的默认选择区边长分别为 152 / 160 / 168 / 176 / 184 DIP，同时调整原生面板间距与数值字段尺寸，并随密度和文字缩放变化。根部最大宽度受容器约束；显式宽度、内边距或特别窄的容器可能改变选择区比例。尺寸变化不重置值、草稿和焦点。

Throttle 未设置时使用 50 毫秒，`nui.Set(time.Duration(0))` 表示不节流，负值无效。绑定颜色总是即时更新，OnChange 按间隔合并并保留末次颜色。外部赋值静默，重新填入原通道值不重复通知；禁用、卸载、移除回调或外部替换颜色会取消待发通知。调度由窗口构建与持久服务完成，没有保存 Element、Context 或后台回调。

未完成的 RGBA 草稿保留最后有效颜色。数字按 InputNumber 的整数步长处理；修改实际通道时从 sRGB 字节重建颜色，避免保留失效的广色域数据。表单将这些草稿作为一个颜色字段跟踪，初次显示和未编辑的重建不会因四个子输入反复写入而被误标为已修改。

Label 为空时继承 FormField。SV 区域支持方向键调整饱和度与亮度，滑条保留原生键盘操作，内部名称随窗口语言变化。RGBA 的整数上限为 255，HEX 接受 3 / 4 / 6 / 8 位格式；非法 HEX 保留最后有效颜色，失焦后按底层原生规则恢复规范文本。表单标签和错误焦点使用首个 RGBA 数字输入，未公开内部 SV / HEX 的持久 Handle。

UI / Style 支持 Root、Panel（原生面板容器）、Body（RGBA 容器）、Item（通道行）和 Label（通道字母）。内部 SV、色相/透明度轨道、HEX 和色板没有公开独立部件，仍使用局部 Native 主题；不能通过任意部件名修改它们。原生布局保留横向轨道、透明度和色板，与 Nuxt 的纵向色相轨道不同。RTL 表单中的原生颜色面板固定为 LTR，使点击位置、方向键、渐变和滑块一致；RGBA 行继续继承表单方向。

颜色值采用 ui.Color，HEX、RGB 或其他字符串格式的序列化由应用负责。本组件不提供 HSL / CMYK / Lab 文本编辑、广色域编辑、屏幕取色或色彩管理。HEX 编辑仍受[输入法取消问题](compatibility.md#输入法取消)影响，完整读屏体验尚未验证。

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
| `TableColumn.Sortable` | 未设置时由 Compare 决定；显式 false 禁止该列排序，true 且无 Compare 时需要 ManualSorting |
| `TableColumn.Cell` | 自定义单元格，接收 Context、Theme、行值和是否选中 |
| `Selection` / `Sort` | 分别为 `ui.Selection[string]`、`ui.SortOrder` |
| `state.List.SortBy` | 指向持久的 `[]ui.SortOrder` 启用多列排序；Sort 同步第一项 |
| `ManualSorting` | 保留 SetRows 的输入顺序，应用根据排序状态加载或计算数据 |
| `InvalidateSort` | 比较器捕获的数据改变时请求重新排序；SetRows 已自动执行失效处理 |
| `state.List.Columns` | 保存列宽和顺序，可由应用持久化 |
| `Row` / `RowID` / `Len` | 读取当前视图中的行、行 ID 和数量 |

列还可设置 Width、MinWidth、MaxWidth 和 Align。排序后的行位置会变化，编辑数据时应通过 RowID 或 Selection 中的 ID 回查业务模型。CellBuilds 可用于统计当前构建中的单元格调用次数。

清除排序恢复最近一次 SetRows 的输入顺序；比较值相等时也保持这个顺序。多列排序下 Shift / Ctrl（macOS 为 Cmd）点击表头可添加排序列，再依次切换降序和移除。普通点击切换该列升降序。删除或禁用排序列会移除无效排序项。

OnSortChange 和 OnSelectionChange 分别通知用户引起的排序、选择变化；外部模型赋值和 SetRows 清理不触发它们。选择通知返回按字符串排序的稳定 ID。OnActivate 在 Enter 或双击时返回行 ID，OnChange 保留原生综合通知。它们与 Nuxt 的行点击 select 事件语义不同。List.Selected 为可选的活动行索引，未绑定时 OnActivate 使用内部索引；操作业务数据仍应使用 ID。

Table 会设置 List.Key、Selection、Sort、Label，应用不要覆盖这些绑定。List.SortBy、Columns、Selected、KeepOffset、FollowEnd、HeaderMenu、CellMenu 等保留 MyGo 规则。List.CellSelect 是原生扩展，本包装器没有 RowHeader 列；完整单元格选择语义与行选择通知不同。

### 筛选与功能范围

搜索、筛选和分页由应用先计算当前 rows，再调用 SetRows。数据浏览器示例展示搜索、团队/状态筛选和每页 100 条数据；服务端请求、总量和缓存由应用管理。

EmptyText / BuildEmpty 设置表头下方的空内容。Loading 显示加载线并保留已有数据；无数据且提供 BuildLoading 时显示加载内容，否则仍显示空内容。LoadingColor 默认 Primary，LoadingAnimation 提供 Carousel、CarouselInverse、Swing、Elastic，减少动态效果时使用静态线。Caption 为表格的辅助技术描述，BuildCaption 可显示可见标题。默认 Height 为 320 DIP，包含标题；表头固定，正文在内部滚动。

虚拟化始终开启，只减少可见单元格的构建量，数据生成、ID 校验、排序和筛选仍取决于数据量。整行布局由 MyGo 负责，单元格可通过列的 UI / Style 覆盖 Item、Label；列覆盖在表格级覆盖之后应用。

`Fixed` **只禁止移动列和调整列宽**，没有冻结滚动的效果。当前还不支持冻结列、多层表头、整行或表头的任意构建回调、行展开、分组聚合及完整 TanStack 语义。原生表头没有独立键盘焦点，需要键盘排序时提供按钮修改 Sort / SortBy。底层 List 会测量行高，自定义单元格可以增高；复杂动态内容仍需应用验证。

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

OnExpandedChange 在用户操作实际改变展开状态后通知，参数为按字符串排序的展开 ID；取消、初始化和外部 SetExpanded 不触发。OnChange 仍只通知选择变化。选择处理器替换节点后，旧点击不再对新建的同 ID 节点执行展开。

TreeNode 的 UI / Style 可覆盖 Item、Leading、Label、Trailing，节点覆盖在树实例覆盖之后应用。BuildLeading / BuildTrailing 替换首尾图标，Build 替换正文；这些回调应只构建展示内容。TreeItemState.Indeterminate 仅用于多选联动：BubbleSelect 按可用后代是否混合选中判断，只有 PropagateSelect 时还要求父项已选中。单选或关闭两项联动时为 false。

已加载的空分支保留展开状态，但不补默认文件夹和箭头；需要文件夹外观时显式提供 Icon。字符查找支持重复字符循环，以及继续输入前缀时保留当前匹配。

## 样式部件索引

| 组件 | 可用部件 |
| --- | --- |
| Form | Root |
| Listbox | Root、Editor、LoadingSlot、LoadingIconSlot、List、Label（组标题）、SeparatorSlot、Item、Leading、Body、ItemLabelSlot、Description、Trailing、Indicator、EmptySlot |
| SelectMenu | Root、Trigger、Leading、Label、Trailing、Close、Editor、List、Header（组标题）、SeparatorSlot、Item、ItemLeadingSlot、Body、ItemLabelSlot、Description、ItemTrailingSlot、Indicator、EmptySlot；面板使用 Popup.UI |
| InputDate / InputDateRange / InputTime | Root、Editor、Segment、Leading、Trailing；日期范围另有 SeparatorSlot |
| InputNumber | Root、Editor、Leading、Trailing、IncrementSlot、DecrementSlot |
| InputTags | Root、Editor、Leading、Trailing、Body、Item、Label、Close、Indicator |
| PinInput | Root、Item、Label、SeparatorSlot |
| InputRating | Root、Item、Indicator、IconSlot、EmptyIconSlot |
| Calendar | Root、Header、HeadingSlot、Body、GridSlot、GridRowSlot、WeekDaySlot、WeekNumberSlot、Item |
| Table | Root、CaptionSlot、Body、Panel、Item、Label、EmptySlot、LoadingSlot、Indicator |
| Tree | Root、List、Item、Leading、Label、Trailing、Indicator、EmptySlot |
| ColorPicker | Root、Panel、Body、Item、Label |
| RangeSlider | Root |

InputNumber 和 InputTags 的 `UI` 支持表中全部部件，嵌入的 `InputOptions.Style` 仍接收 Root、Editor、Leading、Trailing 四个节点。InputNumber 的按钮回调通过 IncrementOptions / DecrementOptions 配置。Table 的内部原生区域与 ColorPicker 未公开的内部部件使用 Native 主题；RangeSlider 的轨道和滑块按 Theme 及 Color 绘制，不支持单独覆盖 Indicator / Thumb 部件。未公开的部件无法通过自行填写部件名来修改。全局 WithComponent 和实例 UI / Style 的优先级见[主题文档](foundations.md#样式覆盖)。

## 示例

| 命令或页面 | 内容 |
| --- | --- |
| Gallery 的 Advanced forms、Dates & color、Task browser | 专用输入、日期、颜色和数据交互 |
| `go run ./examples/settings` | Project setup 表单和可取消提交 |
| `go run ./examples/data-browser` | 搜索、筛选、Table / Tree、按 ID 编辑和 Toast |

InputTags、PinInput 文本模式、Listbox 筛选框、SelectMenu 和 ColorPicker 的文本输入受已知 IME 取消问题影响，详见[兼容性说明](compatibility.md#输入法取消)。
