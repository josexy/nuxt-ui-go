# 输入、选择与字段

输入值和选择状态由应用模型持有。可运行 `go run ./examples/settings` 查看组合用法；主题配置见 [主题与基础组件](foundations.md)。文本输入受 [输入法取消问题](compatibility.md#输入法取消) 影响。

## 状态、主题与事件

方向、加载态、选择上限、多选、过滤及悬停策略见 [参数与默认值](options.md)。

所有入口接收当前 `*ui.Context` 与 `nui.Theme`。持久模型保存输入值、选择 ID、`SearchState`、`ui.ScrollState`、弹层 bool 或 `ui.Handle`；不保存 Element、Context、Parts。每个动态控件先用 `c.Key(stableID)` 定义身份；同一个业务模型可以被多个视图显示，但每个窗口的独立编辑状态由调用方创建。

输入控件支持 `XS/SM/MD/LG/XL`，默认 MD；单行外框在默认字号、密度下的基准高度是 24/28/32/36/40 DIP。原生字体或回退字体需要更高的行高时，控件会随之增高以容纳文字；例如 Linux 使用 Noto CJK 回退字体时，含中文的 MD Input 可高于 32 DIP。Input、Textarea、Select、InputMenu 支持字段变体 `Outline`（默认）、`Soft`、`Subtle`、`Ghost`、`None`，不接受 Button 的 Solid/LinkVariant。错误态仍有错误边框；None 取消普通背景与焦点外环。Color 主要决定焦点/选中强调色。浅深色、字号、圆角、密度来自 Theme。

文本输入框在鼠标点击和 Tab 聚焦后均显示强调边框/外圈，失焦后恢复；只读字段仍可聚焦、选择与复制，错误边框优先保持错误色。Select 等非文本触发器保留键盘焦点提示；Select 的 Outline/Subtle 变体具有悬停背景，禁用时不响应悬停。已展开的 Select 弹层会继承触发器的有效禁用状态（包括祖先禁用），不能在禁用后继续修改选中值；当前 MyGo Base 可保留禁用弹层直到关闭。

组件默认使用 `theme.WithComponent(nui.ComponentInput, nui.ComponentTheme{...})` 等 Component 标识。尺寸顺序为实例 Size → FormField/FieldGroup 传入的 Size → 对应组件默认值；字段禁用向内传播，实例 Disabled=false 不能反向启用禁用父容器。字段 Required/Error 是内部 Theme 的纯数据，不持有 UI 引用。

固定 MyGo 构造器基于上一构建的配置消费排队输入；`Changed/Submitted` 是当前构建的通知，`OnClick/OnChange/OnSubmit` 在当前构建后运行。包装组件保持这个约定。同步 Style、内容 Build 只用于当次构建；业务副作用放进动作回调，不能无条件放进 builder。

## Input 与 Textarea

```go
// app.Name、app.Bio、app.NameError 是窗口模型字段；以下在 view 内执行。
nui.FormField(c, theme, nui.FormFieldOptions{
    Label: "姓名", Required: true, Error: app.NameError,
}, func(field nui.Theme) {
    nui.Input(c, field, &app.Name, nui.InputOptions{
        Placeholder: "输入姓名", Clearable: true, ClearLabel: "清空姓名",
        OnChange: func(value string) { app.NameError = "" },
    }).FillWidth()
})
nui.Textarea(c, theme, &app.Bio, nui.TextareaOptions{
    InputOptions: nui.InputOptions{Label: "简介", Placeholder: "介绍一下自己"},
    Rows: 3, MaxRows: 6, AutoResize: true,
}).FillWidth()
```

值为非 nil `*string`；空字符串有效。`InputOptions` 提供 Label、Placeholder、ErrorMessage、Disabled、ReadOnly、Invalid、Password、Clearable、ClearLabel、Color、Variant、Size、Leading/Trailing Lucide SVG、OnChange/OnSubmit/OnClear，以及 UI 和 Style。ErrorMessage/Invalid 提供控件错误语义；要显示独立错误文字，使用 FormField.Error。

输入编辑、选区、复制粘贴、撤销和 IME 由 TextInputBase/TextAreaBase 处理。密码模式只适用于单行 Input；Textarea 不隐藏文字。ReadOnly 允许选择/复制，阻止修改并关闭编辑 IME；Disabled 阻止交互。原生密码可访问树只有 password 状态，不暴露明文值。

清空按钮只在值非空且非只读时出现，继承禁用状态；一次操作清空绑定值，调用一次 OnChange 和一次 OnClear，并恢复编辑器焦点。清空是程序设置绑定值，**不承诺是编辑器可撤销的原子操作**。若业务要求撤销清空，需要另行设计编辑命令接口。

InputOptions 支持 Loading/LoadingIcon（保留可编辑状态）、Highlight 和 MaxLength。MaxLength 按 UTF-16 单元限制用户插入，0 不限，先限长再调用 OnChange；支持替换选区并保留未改动后缀。FormField 的 Label/Hint/Description/Help/ErrorSlot 均接受部件样式，三态 Checkbox 的标签激活遵循混合态→Checked。

Textarea 默认 Rows=3；关闭 AutoResize 时固定显示对应行数。打开后按实际换行行数增长，MaxRows≤0 表示不限制行数；显式正值小于 Rows 时按 Rows 处理；达到上限后由原生编辑器内部滚动。它不是任意文档编辑器，也没有字数计数器、语法高亮或富文本。

`UI` 支持 Root、Editor、Leading、Trailing。`Style(root, editor, leading, trailing)` 最后同步执行；返回值是外壳，绑定编辑焦点/编辑命令请在 Style 的 editor 上操作或 Bind Handle，不能把外壳当成文本编辑器。缺席图标是零 Element。清空按钮通过 ClearLabel 命名，暂不提供独立样式 slot。

已知限制：选中原文后开始组合输入，再取消时可能丢失原选中文字，Undo 可恢复。受影响范围与平台状态见 [兼容性说明](compatibility.md#输入法取消)。

## FormField 与 FieldGroup

`FormField(c,t,FormFieldOptions,func(Theme))` 通过本库字段绑定关联首个控件的标签、描述、错误和持久 Handle。Options 包含 Label、Description、Help、Hint、Error、Required、Disabled、Size、Orientation、Decoration。Required 显示星号，并在支持的子控件上追加必填描述；它不执行校验，底层控件未暴露独立的原生 required 标志。错误与必填描述的文案规则见 [国际化](i18n.md)。

Label/Hint/Description 在标签列，输入和 Help/Error 在内容列。默认纵向，Orientation=Horizontal 时两列左右排列。标签激活首个绑定控件，Checkbox/Switch/CheckboxGroup 保留切换与通知；多输入建议分别包 FormField。直接使用自定义 MyGo 控件需自行提供标签关联。可覆盖 Root、Label、Hint、Description、Help、ErrorSlot 部件。

`FieldGroup` 接收带唯一非空 ID 的同步 builder 列表，重复 ID、nil builder 会 panic：

```go
nui.FieldGroup(c, theme, nui.FieldGroupOptions{Size: nui.MD}, []nui.FieldGroupItem{
    {ID: "name", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Input(c, inner, &app.Name, nui.InputOptions{Label: "姓名"}).Grow(1)
    }},
    {ID: "check", Build: func(c *ui.Context, inner nui.Theme) ui.Element {
        return nui.Button(c, inner, "检查", nui.ButtonOptions{Variant: nui.Outline}).OnClick(app.Check)
    }},
}).FillWidth()
```

支持 Input、Textarea、InputMenu、Select、Button 的首尾圆角，默认相邻边框重叠 1 DIP；Vertical 改为纵向，Size/Disabled 向内传递。Root slot 可调整间距。保留子控件本来的 Tab 顺序，没有新增整个组的焦点站点。固定 MyGo 没有公开 z-index，焦点环仍按原生兄弟绘制顺序；共享边缘可能被后一个兄弟遮住，不承诺网页 focus-within 的置顶效果。

## Checkbox、CheckboxGroup、RadioGroup、Switch

Checkbox/Switch 绑定非 nil `*bool`，共享 ToggleOptions（CheckboxOptions/SwitchOptions 为别名）：Label、Description、ErrorMessage、Disabled、Invalid、Color、Size、OnChange(func(bool))、Decoration。使用公开 Base 的点击、Space、禁用和角色语义；更新即时呈现，暂未增加切换动画。支持 Root、Indicator、Label、Description，Switch 另有 Thumb。三态使用 [CheckboxTri](forms-data.md#数值标签pin评分)。

CheckboxGroup 绑定 `*[]T`，nil slice 代表空选择，外层指针不可 nil；选项值类型 T 必须 comparable，选项值唯一。选中时追加 ID，取消时移除该 ID；选项重排不改变已选业务值，暂时不在选项集中的值保留，由应用决定何时清理。每个可用复选框独立参与 Tab。

RadioGroup 绑定 `*T`，不保留特殊零值；空选择用 Optional[T] 或明确的“无”选项表达。复用原生 RadioGroup/RadioBase 的分组焦点，方向键、Home/End 跳过禁用项。不存在或被删除的选中值不会自动改为第一项。

两类组共享 ChoiceGroupOptions：Label、Description、ErrorMessage、Disabled、Invalid、Horizontal、Color、Size、OnChange(func())、Decoration。默认纵向，Horizontal 允许换行；禁用项可见但不可修改。选项 Label/Description 作为文本展示；这里不使用 Option.Build/Icon（它们用于 Select/InputMenu）。CheckboxGroup 支持 Root/Item；内部 Checkbox 可通过 ComponentCheckbox 定制。RadioGroup 支持 Root/Item/Indicator/Label/Description。

## Slider

绑定非 nil `*float64`；SliderOptions 包含 Label、Description、Min/Max（Optional[float64]）、Step、Color、Size、Disabled、Vertical、OnChange、Decoration。默认 0..100、水平、MD；边界必须有限且 Max>Min，Step 非负且有限。输入值钳制到范围，NaN/Inf 按 0 再钳制；空值不是 Slider 的模型。

Step=0 表示默认步长 1；PageUp/PageDown 与 Shift+方向键走 10 步，Home/End 到边界。公开 Range/Step 始终表示实际业务值。OnChange 是实时变化，OnCommit 是有变化的按键或拖拽结束通知。默认水平高 24 DIP，垂直高 160 DIP，可后置改宽高。方向、反向、范围滑块及纵向无障碍元数据限制见 [参数与默认值](options.md)。

Root slot 和 Style(Root) 控制整体元素；Indicator.Background 控制已选轨道色、Thumb.Background 控制滑块内色。轨道与滑块由同步 Draw 使用数值/颜色快照绘制，没有独立可交互的 Thumb 元素；其他 Thumb/Indicator 样式字段当前无效。

## Select 与 InputMenu

公共选项为 `Option[T comparable]{Value, Label, Description, Disabled, Icon, Build}`。Value 与 Label 分开，允许重复显示名称；ID 必须唯一、可反身比较（拒绝 NaN）。自定义 Build 同步构建选项内容，仍应填写 Label 供辅助技术与搜索使用。禁用项只显示，不进入 Base 的键盘候选列表。列表尚未虚拟化，不承诺大数据异步搜索性能。

Select 绑定 `*T`，默认 Placeholder 为空；没有内置零值哨兵。以下字段无选择时是 Optional 的零值，真实选项仍可使用字符串零值：

```go
// app.Region 为 nui.Optional[string]
nui.Select(c, theme, &app.Region, []nui.Option[nui.Optional[string]]{
    {Value: nui.Optional[string]{}, Label: "无地区"},
    {Value: nui.Set("ap"), Label: "亚太", Description: "Singapore"},
    {Value: nui.Set("us"), Label: "美国", Disabled: true},
}, nui.SelectOptions{Label: "地区", Placeholder: "选择地区"}).FillWidth()
```

SelectOptions 有 Label、Placeholder、ErrorMessage、Disabled、Invalid、Color、Variant、Size、Popup、OnChange(func())、Decoration。无选项时显示 No options；未匹配值保留并显示占位。打开时使用方向键/Home/End、Enter 选择、Escape 关闭；动态重排按业务 ID 恢复高亮，原高亮被删除时退到当前有效选中项或首个可用项。UI slots 为 Root、Label、Trailing、Panel、Item、EmptySlot。

InputMenu 绑定 `*SearchState`，其中 Query 是编辑草稿，Selected 是单选的稳定 string ID；Selected="" 表示未选，所以选项 ID 不允许空字符串。默认仅按 Label 忽略大小写包含匹配，可显式设置 FilterFields；SearchSelect 兼容入口保留 Label+Description。多选以 SelectedValues 为唯一选中来源。应用删除选项时不自动清理模型。

InputMenuOptions 嵌入 InputOptions，增加 EmptyText、Popup、AllowCustom、OnSelect(func(string))。默认只允许选候选，不执行自由文本 OnSubmit；AllowCustom=true 时，无候选选择的 Enter 执行 InputOptions.OnSubmit(Query)，不会伪造 Selected 或触发 OnSelect。选择候选只触发一次 OnSelect，不同时提交自由文本。Clearable 清空 Query 和 Selected，并调用 OnClear；一般输入的 OnChange 用于草稿变化，业务选择监听 OnSelect。

没有匹配/只有禁用候选时，显示 EmptyText（默认 No results found）状态。通过公开 Overlay+AttachTo 补充 ComboboxBase 零候选隐藏面板的行为；空状态无可选哨兵、不抢焦点、不吃外部点击。Escape/失焦跟随原生 Combobox。选项的 ID、Label、Description、Disabled 集合变化时关闭正在显示的面板，重新打开后使用新集合，避免基于索引的高亮错选；没有实现异步请求取消/加载状态。

InputMenu 的 UI 包含 Input 的四个 slot 及 Panel、Item、EmptySlot；继承的 Style 回调仍只接收输入外壳/编辑器/两侧图标，面板可用 Popup.Decoration.Style。SearchSelect 是便捷入口，内部转到 InputMenu；新代码使用完整 Options。

## Popover 与 ScrollArea

```go
anchor := nui.Button(c, theme, "帮助", nui.ButtonOptions{Variant: nui.Outline}).
    OnClick(func() { app.Help = !app.Help })
nui.Popover(c, theme, anchor, &app.Help, nui.PopoverOptions{
    Label: "帮助面板", Width: nui.Set(float32(300)), Placement: nui.BottomStart,
}, func(inner nui.Theme) {
    ui.Text(c, "说明文字").TextColor(inner.Muted)
    nui.Button(c, inner, "关闭", nui.ButtonOptions{}).OnClick(func() { app.Help = false })
})
```

Popover 使用 PopoverBase；anchor 必须来自当前构建，open 指针持久且非 nil。锚点无效/禁用时关闭。非模态：外部点击关闭并传递给目标；Escape 关闭、恢复打开位置的焦点；没有焦点陷阱。内容显式使用传入 Theme 和局部原生主题，避免脱离原局部主题的 Overlay 变色。

PopoverOptions 为 Label、Placement、Width/MaxHeight/Offset（Optional[float32]）、MatchAnchor、Decoration。默认 BottomStart、间距 8 DIP、最大高度 320 DIP、内容决定宽度；支持上/下/左/右各 start/center/end 共 12 个方位。上下 start/end 跟随文字方向，左右为物理方向；间距使用 AttachTo 的单侧 margin，避免对称 margin 相互抵消；窗口边缘由 AttachTo 避让。返回面板，关闭时为空元素，不能跨构建保存。样式 slot 为 Panel。

面板配方为主题背景、边框、圆角 8、padding 12、gap 8 与阴影。Select/InputMenu 共用该配方，采用默认最大高度 240、padding 4、gap 0，并至少匹配触发器宽度。选择列表自己的 Panel 覆盖最后应用；Popup.UI 中的显式 padding 也有效。通用 Popover 的大段内容需要调用方加入 ScrollArea；MaxHeight 不等同于自动滚动。

ScrollAreaOptions 有 Direction（ScrollVertical 默认、ScrollHorizontal、ScrollBoth）、ScrollbarColor/ScrollbarWidth、State（*ui.ScrollState）、Decoration。内部公开 Scroll 负责滚轮、拖动和滚动边界；滚动条颜色默认 BorderAccented，宽度默认使用 MyGo（当前 6 DIP；设置 0 也采用其默认值，不代表隐藏）。Theme wrapper 创建实际布局元素，必须给**返回的 wrapper**设置 Height/Fill/Grow，内容的 padding/gap 放在 Body slot。Root/Body 是可覆盖部件，State 可选用于读取滚动位置。

ScrollArea 支持嵌套与水平滚动，不提供虚拟列表、无限加载或自定义滚动条交互。大量条目可使用 [Listbox 或 Table](forms-data.md) 的虚拟化能力。

## 相关组件

复杂表单、三态复选框、范围滑块、可搜索多选、日期和数据展示见 [表单与数据](forms-data.md)。对话框、菜单与通知见 [弹层与导航](overlays-navigation.md)。
