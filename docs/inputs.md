# 输入、选择与字段

输入值和选中项由应用保存，组件负责显示和处理用户操作。下面先按需求选择组件，再查具体写法。状态放置和示例变量的约定见[组件指南](README.md#理解组件的基本写法)。

| 需要什么 | 组件 |
| --- | --- |
| 单行或多行文字 | [Input / Textarea](#input-与-textarea) |
| 标签、说明、错误文字，或将多个控件连成一组 | [FormField / FieldGroup](#formfield-与-fieldgroup) |
| 开关、复选或一组互斥选项 | [Checkbox / RadioGroup / Switch](#checkboxcheckboxgroupradiogroupswitch) |
| 拖动调整数值 | [Slider](#slider)，双值范围见 [RangeSlider](forms-data.md#rangeslider) |
| 下拉选择或边输入边搜索 | [Select / InputMenu](#select-与-inputmenu)，独立搜索框和复杂列表见 [SelectMenu](forms-data.md#listboxselectmenu-与异步搜索) |

自定义弹出面板和滚动区域见 [Popover 与 ScrollArea](#popover-与-scrollarea)。可运行的组合示例是 `go run ./examples/settings`。

## 状态、主题与事件

### 保存值和控件状态

组件接收当前 `*ui.Context` 和 `nui.Theme`。输入值、选中 ID、SearchState、ScrollState、打开状态和 Handle 放在应用模型中，按窗口分别保存；Element、Context 和部件不跨构建保存。

动态控件要在创建前使用 `c.Key(stableID)`。同一业务数据可以显示在多个视图中，各窗口的独立编辑状态仍需分别创建。

### 尺寸、变体和禁用

输入控件支持 XS、SM、MD、LG、XL，默认 MD。默认字号和密度下，单行外框基准高度依次为 24、28、32、36、40 DIP。字体行高更大时会自动增高，例如 Linux 的中文回退字体可能使 MD Input 高于 32 DIP。

Input、Textarea、Select、InputMenu 支持以下外观：

| Variant | 效果 |
| --- | --- |
| Outline | 默认的边框样式 |
| Soft / Subtle | 带背景的字段样式 |
| Ghost | 较弱的常规外观 |
| None | 去掉普通背景和焦点外圈；错误边框仍保留 |

字段不接受 Button 的 Solid、LinkVariant。Color 主要控制焦点和选中强调色，字号、圆角、密度及明暗颜色来自 Theme。

尺寸的优先级是：实例 Size → FormField / FieldGroup 的 Size → 组件默认 Size。组件默认值可通过 `WithComponent(ComponentInput, ...)` 等设置。父字段或祖先禁用后，子组件不能通过 `Disabled: false` 重新启用。

### 焦点和回调

文本输入框在鼠标或 Tab 聚焦时显示强调边框及外圈，失焦后恢复。只读字段仍可聚焦、选中和复制；错误边框保持错误色。Select 等触发器保留键盘焦点提示，Outline / Subtle 的悬停背景在禁用时不生效。

已打开的 Select 面板会继承触发器及祖先的禁用状态，禁用后不能修改选中值；当前 MyGo 可能保留面板，直到它被关闭。

MyGo 根据上一构建的配置处理排队输入，`Changed` / `Submitted` 报告本次构建收到的输入通知。`OnClick`、`OnChange`、`OnSubmit` 在构建后执行。业务操作放在这些回调中，`Build` 和 `Style` 只负责当前界面。

## Input 与 Textarea

### 先绑定字符串

下面的 `app.Name`、`app.Bio`、`app.NameError` 都是窗口模型中的字符串字段：

```go
nui.FormField(c, theme, nui.FormFieldOptions{
    Label: "姓名",
    Required: true,
    Error: app.NameError,
}, func(field nui.Theme) {
    nui.Input(c, field, &app.Name, nui.InputOptions{
        Placeholder: "输入姓名",
        Clearable: true,
        ClearLabel: "清空姓名",
        OnChange: func(value string) { app.NameError = "" },
    }).FillWidth()
})
nui.Textarea(c, theme, &app.Bio, nui.TextareaOptions{
    InputOptions: nui.InputOptions{Label: "简介", Placeholder: "介绍一下自己"},
    Rows: 3,
    MaxRows: 6,
    AutoResize: true,
}).FillWidth()
```

值参数必须是非 nil 的 `*string`，空字符串是有效值。

### 常用选项

| 选项 | 用途 |
| --- | --- |
| Label / Placeholder | 控件名称和占位文字 |
| Invalid / ErrorMessage | 标记控件错误；要在字段下方显示错误文字，使用 FormField.Error |
| Disabled / ReadOnly | Disabled 阻止交互；ReadOnly 允许选择、复制，禁止修改并关闭编辑 IME |
| Password | 仅单行 Input 隐藏文字；Textarea 不支持密码显示 |
| Leading / Trailing | 在输入框两侧显示 SVG 图标 |
| Avatar | 前导头像；Leading 或前导加载图标优先 |
| Loading / LoadingIcon / LoadingTrailing | 显示加载图标，输入框仍可编辑；LoadingTrailing 将加载图标放在尾部，显式 Leading 仍优先使用前导加载 |
| Clearable / ClearLabel | 启用清空按钮及其名称 |
| MaxLength | 限制用户插入的 UTF-16 单元数，0 为不限 |
| Highlight / Color / Variant / Size | 强调状态、颜色、外观和尺寸 |
| OnChange / OnSubmit / OnClear | 输入变化、提交和清空回调 |

占位文字默认使用 Dimmed 色；需要改色时，可在 Style 回调中对 editor 调用 `PlaceholderColor`。密码输入在原生辅助技术树中只提供 password 状态，不暴露明文值。

Avatar.Size 未设置时，头像按输入尺寸 XS–XL 使用 16 / 16 / 20 / 20 / 24 DIP；显式 Avatar.Size 优先。尾部加载可以与前导头像同时显示。切换头像或加载位置不会重置文字、焦点和选区。

Neutral 的焦点和强调边框使用 InvertedBackground。Highlight 也适用于 None 样式；错误状态优先。

### 编辑与清空

文本编辑、选区、复制粘贴、撤销和 IME 由 MyGo 的 TextInputBase / TextAreaBase 处理。MaxLength 在 OnChange 之前限制插入，替换选区时会保留未修改的后缀；完全拒绝且值未改变的插入不触发 OnChange。外部赋值不会自动截断，也不会触发输入通知。

清空按钮只在有内容且非只读时出现，并继承禁用状态。一次清空更新绑定值，分别调用一次 OnChange 和 OnClear，然后恢复编辑器焦点。清空通过程序改值实现，不能保证作为一次可撤销的编辑命令；应用需要撤销清空时，应另行保存和恢复数据。

**已知输入法问题：** 选中文字后开始组合输入，再取消组合，原选中文字可能丢失，撤销可以恢复。受影响范围见[兼容性说明](compatibility.md#输入法取消)。

### Textarea 的行数

Rows 默认 3。未启用 AutoResize 时固定显示相应行数；启用后按实际换行增长。MaxRows≤0 表示不限，正值小于 Rows 时按 Rows 处理。达到上限后在编辑器内部滚动。

Textarea 提供普通多行文本编辑，不含字数计数器、语法高亮和富文本功能。

Textarea 的头像和加载图标与第一行对齐。Enter 和 Shift+Enter 用于换行，不触发 OnSubmit；应用可通过独立按钮或编辑器快捷键提交。共享 Password 字段对多行输入不生效。原生没有拖拽调整高度手柄，调整高度应使用 Rows / AutoResize 或外部布局。

### 样式与编辑器 Handle

Input / Textarea 的可用部件是 Root、Editor、Leading、Trailing。`Style(root, editor, leading, trailing)` 最后同步执行，不存在的部件为零值 Element。Leading 可对应图标、加载图标或头像。加载图标颜色使用相应 `UI.Leading.Foreground` / `UI.Trailing.Foreground`；Style.TextColor 不改变自绘颜色。清空按钮使用 ClearLabel 命名，尚无独立样式部件。

返回元素是输入框外层容器。需要绑定编辑焦点或执行编辑命令时，在 Style 中操作 editor，或为 editor 绑定持久 Handle，不要把外层容器当作文本编辑器。

## FormField 与 FieldGroup

### FormField：标签、说明和错误

FormField 将标签、描述、错误和 Handle 关联到其中第一个支持字段绑定的控件。主要选项为 Label、Description、Help、Hint、Error、Required、Disabled、Size、Orientation 和 Decoration。

`Required` 显示红色星号，并为支持的控件追加必填描述。**实际校验仍需由应用或 Form 完成**，底层控件目前没有独立的原生 required 标志。

| 区域 | 内容 |
| --- | --- |
| 标签区 | Label、Hint、Description |
| 内容区 | 输入控件，以及下方的 Help 或 Error |
| 默认排列 | 标签区在上，内容区在下 |
| Horizontal | 标签区和内容区左右排列，并按首行文字基线对齐 |

描述紧贴标签，纵向内容区与标签区间距为 4 DIP；Help / Error 与输入间距为 8 DIP，横向两列间距也为 8 DIP。没有 Label、Hint 和 Description 时不创建标签区，也不保留顶部间距。星号距标签 2 DIP，辅助技术名称保留标签和 `*`；无标签时不单独显示星号，必填说明仍传给控件。

点击标签会激活首个关联控件；Checkbox、Switch、CheckboxGroup 保留切换和通知行为，三态 Checkbox 从混合态变为 Checked。多个独立输入应分别使用 FormField。自定义 MyGo 控件需要应用自行建立标签关联。

可用部件为 Root、Header、Body、Label、Hint、Description、Help、ErrorSlot。Header 对应标签区，Body 对应输入与帮助/错误区域。Required 和 Error 通过内部 Theme 传给子控件，只保存数据，不保存 UI 引用。

Error 是字符串，优先于 Help；为空时可使用已触碰字段的 Form 错误。没有上游 error=true 的仅强调、error=false 的隐藏错误、errorPattern 正则匹配或字段级延迟校验。名称精确匹配，校验由 Form.Validate 或应用执行。原生 Hint 可以在没有 Label 时独立显示。

### FieldGroup：连接多个控件

将输入框和按钮组成连续的一行时使用 FieldGroup：

```go
nui.FieldGroup(c, theme, nui.FieldGroupOptions{Size: nui.MD}, []nui.FieldGroupItem{
    {ID: "name", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Input(c, inner, &app.Name, nui.InputOptions{Label: "姓名"}).Grow(1)
    }},
    {ID: "check", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Button(c, inner, "检查", nui.ButtonOptions{
            Variant: nui.Outline,
        }).OnClick(app.Check)
    }},
}).FillWidth()
```

每项 ID 必须唯一且非空，Build 不能为 nil，否则 panic。Build 同步创建当前内容。

FieldGroup 会处理 Input、Textarea、InputMenu、InputTags、Select、SelectMultiple、SelectMenu、Button、Badge 的首尾圆角，默认相邻边框重叠 1 DIP，并拉伸到相同高度。外侧圆角随书写方向调整；Vertical 改为纵向。

Disabled 向内传递。Size 优先使用当前实例设置，其次是 ComponentFieldGroup 的默认尺寸，再继承外层 FormField；仍未设置时，由子组件使用自己的默认尺寸。子控件弹出的面板保留自己的圆角，例如 SelectMenu 的搜索框和 InputDate 的日历。

Root 部件可调整间距。组内继续使用各控件的 Tab 顺序，不增加整个组的焦点位置。MyGo 尚无公开 z-index，后绘制的相邻控件可能遮住共享边缘处的焦点环。

## Checkbox、CheckboxGroup、RadioGroup、Switch

### 选择绑定类型

| 组件 | 绑定值 | 适用场景 |
| --- | --- | --- |
| Checkbox / Switch | 非 nil `*bool` | 一个开关状态 |
| CheckboxGroup | 非 nil `*[]T`，T 必须 comparable | 一组可多选项目；nil 切片表示空选择 |
| RadioGroup | 非 nil `*T` | 一组互斥项目；需要空选择时用 Optional[T] 或明确的无选项 |

选项值必须唯一，显示名称可以重复。CheckboxGroup 选中时追加值，取消时移除值；重排不改变已选值，暂时不在选项中的值也会保留，由应用决定何时清理。RadioGroup 不会将缺失的值自动改为第一项，也不预留特殊零值。

三态复选框使用 [CheckboxTri](forms-data.md#checkboxtri)，支持 Unchecked、Checked、Indeterminate。

### 选项和键盘操作

Checkbox / Switch 共用 ToggleOptions，CheckboxOptions / SwitchOptions 是其别名。常用字段包括 Label、Description、ErrorMessage、Disabled、Invalid、Color、Size、OnChange 和 Decoration。

共享类型不表示所有字段对两种控件都生效。Checkbox 不使用 Loading、LoadingIcon 和 UncheckedIcon；Switch 使用这些字段。RadioGroup 的 Max 不生效，它始终单选；CheckboxGroup 的 Max 限制选中数量。网站保留这些公开字段，并按实际组件显示不适用提示。

单个 Checkbox 的 ChoiceTable，以及 Switch 的 Variant / Indicator，是 Go 原生扩展；CheckboxGroup 和 RadioGroup 的 ChoiceTable 则对应 Nuxt UI 已有的组样式。这些原生扩展继续可用。

两类选项组使用 ChoiceGroupOptions，另有 Horizontal、Loop 和无参数的 OnChange。默认纵向；Horizontal 允许换行，ChoiceTable 除外。选项的 Label / Description 用作文字；Option.Icon 在多选组中替换勾选图标，在两种组的隐藏模式中显示于标签上方。Option.Build 在选项组中不生效，仍只用于 Select / InputMenu。

Checkbox / Switch 使用点击或 Space 切换。两类选项组都作为一个 Tab 停靠点；横向使用左右键，纵向使用上下键，Home / End 到首尾，禁用项会跳过。CheckboxGroup 的方向键只移动焦点，Space 切换选择；RadioGroup 移动时同步选择。Loop 使用 Optional[bool]：多选组默认 false，单选组默认 true，显式 Set(false) 可停止循环。状态即时更新，当前没有切换动画。

组的可见标题可由外层 FormField.Label 提供。需要统一多选图标时，为未指定图标的 Option.Icon 设置同一 SVG；已指定的项目图标继续保留。

### 图标、样式和间距

Checkbox 的 CheckedIcon 可替换默认对勾；CheckboxTri 的 IndeterminateIcon 可替换部分选中时的减号。IndicatorHidden 会将显式 CheckedIcon 放在标签上方，图标和文字居中；未指定图标时只显示文字。部分选中专用图标只用于可见指示器，不替换隐藏模式的标签图标。UncheckedIcon 仅用于 Switch。

Checkbox 的 Leading 部件设置勾选图标或隐藏模式标签图标的样式。ChoiceCard / ChoiceTable 的键盘焦点圈显示在外框，ChoiceList 的可见指示器使用自己的焦点圈。Switch 轨道默认不描边，Highlight 或错误状态才显示强调色或错误色外圈。

可见指示器的卡片只在指示器上显示 Highlight；隐藏指示器后，卡片外框承担强调状态。卡片外观遵循组件默认样式、实例语义配置、实例 UI 和 Style 的覆盖顺序。

必填标记通过 FormField.Required 设置；它显示星号，并将必填提示与字段说明、控件描述一起提供给辅助技术。它不自动校验选中值，提交约束需要配置 FormState 校验规则。

组的横向间距为 8 DIP，纵向间距按 XS 到 XL 为 2/2/4/4/6 DIP。ChoiceTable 相邻项共用 1 DIP 边框，外侧圆角为 8 DIP，并按书写方向调整。选中项之后的相邻边框会省去，以显示完整的选中边框。

Checkbox / Radio 的指示器容器与文字行同高，按尺寸为 16/16/20/20/24 DIP；Radio 内圆点为 4/4/6/6/8 DIP。

普通标签与描述按各自的行高相接，不另加间距；隐藏指示器的居中布局使用 4 DIP 间距。CheckboxTri 的部分选中状态使用强调色指示器，卡片背景仍与未选中状态一致，只有 Checked 应用已选背景。

| 组件 | 可用部件 |
| --- | --- |
| Checkbox | Root、Indicator、Leading、Label、Description |
| Switch | Root、Indicator、Thumb、Label、Description |
| CheckboxGroup | Root、Item 作用于组和行；其他部件及 Style 传给每个 Checkbox，也可用 ComponentCheckbox 定制 |
| RadioGroup | Root、Item、Indicator、Leading（隐藏模式的图标）、Label、Description |

## Slider

Slider 绑定非 nil `*float64`，默认范围 0–100、水平排列、MD 尺寸。Min / Max 使用 Optional[float64]；边界必须有限且 Max>Min，Step 必须非负且有限。

Label 的显式名称优先，其次采用 FormField 标签；都为空时使用本地化的 Slider / 滑块。标签不绘制为可见文字，Description 提供补充说明。

外部值会被限制到范围内，NaN / Inf 先按 0 处理再限制。Slider 没有空值模型。Step=0 使用默认步长 1。

用户操作按相对 Min 的步长取值，并根据 Min、Step 的十进制精度去掉运算产生的浮点尾数；例如从 0.2 增加 0.1 后显示 0.3。禁用组件或其父级会取消正在进行的拖拽，不再触发 OnCommit。

| 操作 | 行为 |
| --- | --- |
| 方向键 | 按组件方向、Inverted 和 RTL 移动一个 Step |
| PageUp/PageDown、Shift+方向键 | 移动 10 个 Step |
| Home/End | 到达边界 |
| OnChange | 实时报告值变化 |
| OnCommit | 有变化的按键或拖拽结束时通知 |

默认水平高度 24 DIP，垂直高度 160 DIP，可通过返回元素调整。轨道按 XS 到 XL 为 6/7/8/9/10 DIP，滑块使用背景色圆点和 2 DIP 主题色边框；键盘焦点外圈宽 3 DIP，距滑块 2 DIP。RangeSlider 使用相同外观。

Slider / RangeSlider 的 Root 控制整体样式，Indicator.Background 控制选中轨道和圆点边框，Thumb.Background 控制圆点填色。Neutral 使用 InvertedBackground；FormField 错误优先使用 Error，显式实例部件颜色仍可覆盖。普通轨道使用 Theme.Accented，轨道和圆点由绘图回调绘制；其他 Thumb / Indicator 样式字段不生效，Style 只接收 Root。

Tooltip 为布尔开关，悬停圆点或聚焦时立即显示当前数值，Escape 可以关闭。它没有完整 TooltipOptions 或 ComponentTooltip 部件配置。单滑块的 UI.Root.PaddingX / PaddingY 改变轨道有效区时，绘制、指针计算和提示定位同步；RangeSlider 保持半个圆点大小的固定轨道内缩。

双滑块、反向、值提示和纵向辅助技术限制见[参数与默认值](options.md#滑块和内容切换)。

## Select 与 InputMenu

### 先选择交互方式

| 需要什么 | 使用方式 |
| --- | --- |
| 点击后从已有项目中选一项 | `Select`，绑定 `*T` |
| 点击后选择多项 | `SelectMultiple`，绑定 `*[]T` |
| 输入文字时显示候选项 | `InputMenu`，绑定 `*SearchState` |
| 触发器与弹层搜索框分开 | [SelectMenu](forms-data.md#selectmenu搜索弹层) |

Select 和 InputMenu 共用 `Option[T comparable]`，包含 Value、Label、Description、Disabled、Icon、Build。Value 是稳定标识，必须唯一且可反身比较，例如不能用 NaN；Label 可重复。自定义 Build 时也要填写 Label，供搜索和辅助技术使用。

禁用项仍显示，但不能作为键盘候选。列表使用虚拟化，只为可见及少量预留行调用 Build；不要在行构建中执行需要覆盖所有数据的业务操作。输入数据比较仍随数据量增加，远程分页和异步请求由应用处理。

### Select：单选和空值

Select 不预留特殊零值。需要明确表示未选择时，可以将业务值包装为 Optional：

```go
// app.Region 的类型是 nui.Optional[string]。
nui.Select(c, theme, &app.Region, []nui.Option[nui.Optional[string]]{
    {Value: nui.Optional[string]{}, Label: "无地区"},
    {Value: nui.Set("ap"), Label: "亚太", Description: "新加坡"},
    {Value: nui.Set("us"), Label: "美国", Disabled: true},
}, nui.SelectOptions{
    Label: "地区",
    Placeholder: "选择地区",
}).FillWidth()
```

本例把 Optional 零值也列为一个选项，因此初始显示无地区。没有匹配项的值会被保留，并显示 Placeholder；Placeholder 默认空字符串。没有选项时显示内置的没有数据文案。

SelectOptions 包括 Label、Placeholder、ErrorMessage、Disabled、Invalid、Color、Variant、Size、Popup、OnChange 和 Decoration。Leading / Trailing 配置触发器图标，Trailing 默认 ChevronDown；Avatar 在没有前置图标时显示。SelectedIcon 替换选中项的 Check，不替换选项自己的 Icon。

Loading 只显示加载图标，不禁用选择。默认在前侧显示 LoaderCircle；LoadingTrailing=true 可移到尾部，显式 Leading 仍让加载位于前侧。加载图标可由 LoadingIcon 替换，并采用对应部件的前景色。未指定 Avatar.Size 时，头像直径随选择器使用 16 / 16 / 20 / 20 / 24 DIP；显式头像尺寸优先。

Highlight 在失焦时保持颜色边框，None 样式也生效；错误色优先。Neutral 的边框和键盘焦点外圈使用 InvertedBackground。

Open 是可选的持久 `*bool`。不传时使用内部状态；传入后，应用赋值、触发器操作及关闭都会更新同一状态，关闭后返回触发器焦点。打开或外部修改选中值不会触发 OnChange；用户操作改变选中值才通知一次。Popup.Dismissible 默认 true，显式 false 阻止 Escape 和外部点击关闭，显式 Open=false 仍可关闭。

打开后用方向键、Home、End 或文字前缀定位，Enter / Space 选择，Escape 关闭；重复字母会循环查找，并跳过禁用项。单选关闭时输入前缀可直接切换选中值，这是原生便捷行为。多选的前缀导航在展开列表内使用。重排按 ID 保留高亮；高亮项被删除后，回到有效选中项或第一个可用项。多选显示已选名称，用逗号连接并忽略空标签；暂时缺失的已选项保留原名称。

可用部件是 Root、Label、Leading、Trailing、Panel、Item、EmptySlot；SelectMultiple 将 Root 样式映射到触发器。Option.Build 可组合头像、角标等行内容，不保留构建期元素。现有模型不提供 label / separator 项目种类或嵌套分组；Popup 不提供箭头、DOM portal 和上游 item-aligned 定位。单选不使用 Popup.Mode / OpenDelay / CloseDelay，多选沿用 SelectMenu 的原生悬停扩展。

### InputMenu：区分草稿和选中值

SearchState.Query 保存编辑文字，Selected 保存单选 ID。Selected 为空表示未选择，因此 InputMenu 的选项 ID 不能是空字符串。多选以 SelectedValues 为准。删除数据源选项不会自动清理应用的选择。普通单选初始草稿为空时，显示已选名称；外部修改 Selected 在未编辑时同步显示，正在编辑的草稿保留到关闭或失焦。

默认只按 Label 做忽略大小写的包含匹配。FilterFields 可改搜索字段；旧的 SearchSelect 便捷入口保留 Label + Description。数据、查询或过滤字段变化时才重算本地过滤结果。

InputMenuOptions 嵌入 InputOptions，并增加 EmptyText、Popup、AllowCustom、OnSelect 等选项。默认点击只聚焦，输入、方向键或尾部箭头打开面板；OpenOnClick=true 可恢复点击即打开。输入框保持焦点，Trailing 可换图标；Loading 不禁用编辑或选择，Avatar、LoadingTrailing 和 Highlight 同样适用。

Clearable 在有草稿或选中值时显示清空按钮，因此多选草稿为空时也可清空全部标签。ResetModelValueOnClear 默认 true；显式 false 只清空草稿并保留选择。普通文字删除不会清除已选 ID。SelectedIcon、DeleteIcon、ClearIcon 分别替换 Check、标签删除的 X 和清空的 X；Autocomplete 没有选中标记或多选标签。键盘删除标签后返回输入框。

Open 接收持久 bool 指针，nil 时使用内部状态；外部打开会聚焦输入，禁用或只读会关闭。Loop 默认 false，HighlightOnHover 默认 true，均跳过禁用项。Popup.Dismissible=false 阻止 Escape 和外部焦点关闭，Open=false 仍可显式关闭；Popup 的悬停模式和延时不用于 InputMenu。

| 动作 | 回调与状态 |
| --- | --- |
| 修改草稿 | OnChange 报告文字变化 |
| 选择候选 | 调用一次 OnSelect，不同时提交自由文本 |
| AllowCustom=true 且 Enter 未选择候选 | 调用 InputOptions.OnSubmit(Query)，不伪造 Selected，也不触发 OnSelect |
| 清空按钮 | 清空 Query；按 ResetModelValueOnClear 决定是否清除选择，调用 OnClear。多选值实际改变时调用 OnValuesChange |
| 多选删除 | 点击标签 × 或在空草稿时按 Backspace；选择变化见 OnValuesChange |

默认只允许选择候选，不提交自由文本。多选值显示为小标签，只有 × 按钮可点击删除，× 仍参与 Tab 顺序；目前没有标签间的方向键导航。

没有匹配或候选均禁用时显示 EmptyText，默认文案区分没有数据和没有匹配的数据。上下键可导航到屏幕外候选，选择后焦点保留在输入框；Escape 或失焦关闭面板。空状态不参与选择，也不占用焦点或拦截外部点击。

选项的 ID、Label、Description 或 Disabled 变化时，当前面板会关闭，重新打开后使用新数据，避免接受旧候选。组件本身不负责取消网络请求。多选、创建、自动补全和重置策略见 [InputMenu 参数](options.md#inputmenu-值与搜索策略)。

InputMenu 支持 Input 的四个部件，以及 Panel、Item、EmptySlot。描述行中的图标与首行对齐，只有没有匹配候选时才显示空提示，全部禁用的候选仍正常显示。继承的 Style 只接收输入框和图标；面板样式用 Popup.Decoration.Style。新代码建议直接使用 InputMenu 及完整 Options；SearchSelect 保留名称加描述过滤及点击打开的旧行为。

### 选项行的外观

Select、InputMenu、SelectMenu、Listbox、菜单和 CommandPalette 使用相同的选项行样式：

| 尺寸 | XS | SM | MD | LG | XL |
| --- | --- | --- | --- | --- | --- |
| 内边距和间距（DIP） | 4 | 6 | 6 | 8 | 8 |
| 字号 / 行高（DIP） | 12/16 | 12/16 | 14/20 | 14/20 | 16/24 |
| 图标大小（DIP） | 16 | 16 | 20 | 20 | 24 |

行圆角为 6 DIP。键盘或指针高亮的行使用 Elevated 的 50% 背景，文字为 Highlighted，前侧图标由 Dimmed 变为 Text。已选项通过行尾对勾表示，不额外铺背景。

描述在标签下一行，与标签同字号，使用 Muted 色。组标题为半粗体，比条目小一档；空状态居中并使用 Muted 色。

## Popover 与 ScrollArea

### Popover：在控件旁显示内容

```go
anchor := nui.Button(c, theme, "帮助", nui.ButtonOptions{Variant: nui.Outline}).
    OnClick(func() { app.Help = !app.Help })
nui.Popover(c, theme, anchor, &app.Help, nui.PopoverOptions{
    Label: "帮助面板",
    Width: nui.Set(float32(300)),
    Placement: nui.BottomStart,
}, func(inner nui.Theme) {
    ui.Text(c, "说明文字").TextColor(inner.Muted)
    nui.Button(c, inner, "关闭", nui.ButtonOptions{}).
        OnClick(func() { app.Help = false })
})
```

anchor 必须是当前构建中的元素，open 必须是长期保存的非 nil 指针。锚点无效或禁用时会关闭。返回值是本次构建的面板，关闭时为零值 Element，不可跨构建保存。

Popover 是非模态弹层：外部点击会关闭并继续传给目标；Escape 关闭并恢复打开位置的焦点，Tab 不会被限制在面板中。内容使用回调传入的 Theme，以保持局部主题。

| 选项 | 默认行为 |
| --- | --- |
| Placement | BottomStart；支持上、下、左、右各 start/center/end，共 12 个位置 |
| Width / MatchAnchor | 默认宽度由内容决定，可指定宽度或匹配锚点 |
| MaxHeight | 320 DIP，只限制高度，不自动增加滚动 |
| Offset | 与锚点间距 8 DIP |
| Decoration | 使用 Panel 部件设置外观 |

上下位置的 start/end 随文字方向变化，左右位置使用屏幕的物理方向。MyGo 的 AttachTo 负责窗口边缘避让。

通用 Popover 默认使用主题背景、1 DIP 边框、6 DIP 圆角、8 DIP 内容间距、阴影和 16 DIP 内边距。需要内容贴边时，将 UI[Panel] 的内边距设为 0。Select、InputMenu 和菜单沿用相同基础样式，但通常采用 240 DIP 最大高度、4 DIP 内边距、0 间距，并至少与触发器同宽；菜单自身的高度选项见[菜单说明](overlays-navigation.md#dropdownmenu-与-contextmenu)。

Popover 的 Panel 字体、文字颜色、行高和 Gap 直接作用于普通内容；原生子控件继续使用传入的局部主题，显式配置自己的文字样式时仍优先。修改锚点和面板尺寸会保留内容身份、草稿与选区。焦点位于弹层内容时关闭，会返回当前锚点；点击外部控件关闭时，该外部控件继续获得焦点。

入场动画遵循减少动态效果；显示期间切换该偏好不会重新播放入场。关闭立即卸载，不提供上游的离场动画。窗口碰撞使用 MyGo 的 4 DIP 边缘留白，上游默认是 8 px；两侧都没有空间时，侧向弹层仍可能超出窗口，应约束内容宽高或改用上下定位。

打开时在 100ms 内从 95% 大小淡入，缩放从靠近锚点的一侧开始。减少动态效果时直接显示。关闭立即卸载面板并恢复焦点，没有退出动画。选择列表自己的 Panel 覆盖最后应用，Popup.UI 的显式内边距也会生效。

### ScrollArea：给内容设置滚动区域

ScrollArea 提供 ScrollVertical（默认）、ScrollHorizontal、ScrollBoth 三种方向，可选传入 `*ui.ScrollState` 读取滚动位置。

每个实例使用独立的持久状态。X / Y 是以 DIP 表示的偏移，可直接赋值或使用 ScrollTo；MaxX / MaxY 由布局计算。超界值被限制到有效范围，切换方向时不可用轴归零。RTL 下 X 从右侧起点计算；Scrolling() 只表示程序化平滑滚动，不包括所有滚轮操作。

**把 Height、Fill 或 Grow 设置在返回的外层元素上**，才能限制滚动区域的尺寸；内容的内边距和间距放在 Body 部件。可用部件为 Root、Body。

横向使用行布局，纵向和双向使用列布局。Root 的圆角会裁剪内容；改变滚动条样式不会重置偏移。根容器默认绘制主题背景，需要透明背景时显式设置 Root.Background。

ScrollbarColor 默认 BorderAccented，ScrollbarWidth 默认沿用 MyGo 的 6 DIP；小于等于 0 也使用默认宽度，不会隐藏滚动条。显式透明颜色只隐藏绘制，命中区域仍存在。滚轮、拖动和滚动边界由原生 Scroll 处理；内层到达边界后，未消费的滚轮可到达外层。

ScrollArea 支持嵌套和双向滚动，不提供上游 items / virtualize、边缘渐变阴影、无限加载或自定义滚动条交互。全部内容每次都会构建；长列表使用 [Listbox 或 Table](forms-data.md) 的虚拟列表。Popover 中较长的内容也需自行放入 ScrollArea。

## 相关组件

复杂表单、三态复选框、范围滑块、日期和数据展示见[表单与数据](forms-data.md)。对话框、菜单与通知见[弹层与导航](overlays-navigation.md)，主题设置见[主题与基础组件](foundations.md)。
