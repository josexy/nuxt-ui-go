# 参数与默认值

本文汇总方向、选择、加载和弹层等常用配置。各组件的值模型与完整用法见 [组件文档](README.md)。

## 选择与分页

| 项目 | 当前行为 |
| --- | --- |
| SelectMenu 查询重置 | 默认选择后及关闭后清空 Query；同值重选也关闭、重置并触发 OnSelect，数据 OnChange 仅在值变化时触发。覆盖 Escape、外部点击与受控关闭，可分别设置 ResetSearchTermOnSelect / ResetSearchTermOnBlur |
| SelectMenu 键盘打开 | 触发器上的 ↑、↓ 均可打开；默认聚焦搜索，搜索关闭时聚焦列表；禁用触发器不打开 |
| Pagination 零个相邻页 | 使用 `SiblingCount: nui.Set(0)`。`Siblings int` 的零值默认 2；显式设置的 SiblingCount 优先，限制范围 0–10 |

Pagination 同时按 [Reka UI 2.11.0 分页规则](https://github.com/unovue/reka-ui/blob/v2.11.0/packages/core/src/Pagination/utils.ts) 处理边缘窗口：页 5/总 10/相邻数 0，ShowEdges=true 显示 `1 … 5 … 10`，false 显示 `5`；接近首页、末页时扩大边缘段，小页数不显示省略号。

## Orientation

复用 `nui.Horizontal` / `nui.Vertical`。默认纵向的组件使用 `Optional[Orientation]`，区分未设置与显式横向。

| Options | 字段类型 / 默认值 | 切换内容 |
| --- | --- | --- |
| FormFieldOptions | `Optional[Orientation]` / Vertical | Label / Hint / Description 属于标签列，输入 / Help / Error 属于内容列；保持本库控件的标签激活和编辑器身份 |
| AlertOptions / ToastOptions | `Optional[Orientation]` / Vertical | 操作放在正文下方或旁边 |
| ListboxOptions | `Optional[Orientation]` / Vertical | 纵向虚拟列表或按 ItemWidth 裁剪的横向虚拟列表；对应 ↑↓ / ←→，水平键随 RTL 镜像 |
| InputRatingOptions | `Orientation` / Horizontal | 星级横向或纵向排列 |
| RangeSliderOptions | `Orientation` / Horizontal | 共用轨道、两拇指几何与拖拽坐标同步变化；保留两个独立焦点与范围值 |
| ProgressGroupOptions | `Orientation` / Horizontal | 水平分段轨道或纵向分段轨道；纵向 Height 默认 160 DIP |
| CarouselOptions | `Orientation` / Horizontal | 视口排布、拖拽轴、速度判定与方向键同步切换；纵向使用 ↑↓；切换时取消未完成拖拽 |

InputNumber 使用 `Orientation`，默认 Horizontal。部分组件也提供 Vertical / Horizontal 布尔选项，具体以各组件 Options 为准。

```go
nui.FormField(c, t, nui.FormFieldOptions{
    Label: "Workspace", Orientation: nui.Set(nui.Horizontal),
}, func(inner nui.Theme) {
    nui.Input(c, inner, &state.Name, nui.InputOptions{})
})

nui.RangeSlider(c, t, &state.Low, &state.High, nui.RangeSliderOptions{
    Label: "Budget", Orientation: nui.Vertical,
    Min: 0, Max: 100, Step: 5, MinStepsBetweenThumbs: 2,
})
```

## 常用参数入口

| 组件 | 入口与约定 |
| --- | --- |
| Switch | Loading 阻止鼠标和键盘切换；LoadingIcon / CheckedIcon / UncheckedIcon 自定义图标；减少动态效果时停止转动。Checkbox 只使用 CheckedIcon |
| Checkbox / CheckboxGroup / RadioGroup | Variant 使用 ChoiceList / ChoiceCard / ChoiceTable；Indicator 使用 IndicatorStart / End / Hidden。卡片选择后自动着色，隐藏指示器仍可辨认；Highlight 强制强调边框/指示器，未选中也生效。CheckboxGroup.Max 达上限后，未选项与 Nuxt 一样仍可聚焦，但拒绝勾选（可访问状态为只读） |
| SelectMenu | SearchInput、Color、Variant、Size、两个 ResetSearchTerm 选项；SearchInput=false 不构建搜索，也不按保留的 Query 过滤。Invalid / ErrorMessage 与字段错误语义和边框对齐 |
| Select | `SelectMultiple[T](c,t,*[]T,[]Option[T],SelectOptions)` 保持原 `Select(*T)` 类型安全；单选/多选均继承 ComponentSelect 的 Color / Variant / Size / UI，实例覆盖优先；保留稳定 ID、自定义 Option.Build 和已选标签 |
| InputMenu | Multiple、ExternalFilter、FilterFields、Mode、OpenOnClick / OpenOnFocus、两个 ResetSearchTerm 选项、OnCreate / CreateWhenAlways、OnValuesChange；见下一节 |
| Pagination | ShowControls / ShowEdges、Size、Color / Variant、ActiveColor / ActiveVariant；所选语义不覆盖活动页的外观配置 |
| InputTags | AddOnTab / AddOnBlur、AddOnPaste、Duplicate、MaxLength。MaxLength 在编辑层按 UTF-16 单元限长，0 不限，保留插入点后的原文本；不是提交时的标签长度校验。AddOnPaste=true 消费粘贴事件，包括重复/超上限拒绝；false 作为普通文本粘贴。Tab / Shift+Tab 在非空草稿时提交并留焦，空草稿正常移动焦点 |
| Tabs | ActivationMode=ActivationAutomatic（默认）或 ActivationManual；manual 方向键只移动焦点，Enter / Space 激活并通知一次 |
| Accordion | Collapsible 为 Optional[bool]，默认 true；单开模式设置 false 后，激活已开项不收起，仍可切换到另一项；Multiple 模式不受此选项限制 |
| Popover | Mode=PopoverClick（默认）/ PopoverHover；OpenDelay / CloseDelay 仅作用于 Hover，使用 time.Duration，默认 0；网站面板以毫秒输入。Dismissible 默认 true，控制 Escape / 外部点击关闭，不阻止应用主动修改 open。Hover 跨锚点间隙进入面板保持打开，Escape 关闭后不因鼠标仍停留原地而立即重开 |
| Tooltip | Placement / Offset、受控 Open *bool、Kbds；受控状态同样响应原生悬停延迟、键盘聚焦、失焦、触发器点击和 Escape；外部 false 在仍聚焦/悬停时不立即重开，新的提示关闭旧受控提示 |
| Carousel | Arrows 为 Optional[bool]；Dots 开启圆点导航，圆点有本地化的页码名称 |
| Toast | Close / Progress、Actions []ToastActionOptions；每项含 Label、OnAction、ButtonOptions 和 Close（默认 true）；Close=false 的操作执行后保留通知；持续时间条复用原生暂停时钟，常驻通知不显示倒计时 |
| PinInput | Color / Variant / Size；Separator []int 在指定的第几个槽位后插入间隔，例如 `{3}` 表示三位后分组；保留单个编辑器和 Tab 停靠点 |
| InputRating | Clearable 时点击当前评分或在当前星上按空格归零；Hoverable 默认 false，开启后悬停只预览，清零后预览随即消失；EmptyIcon 未指定时继承 Icon。横向 ←→（RTL 镜像）、纵向 ↑↓ 按项导航，到首尾停止、不循环；忽略另一轴，Home/PageUp 到 1，End/PageDown 到 Max |
| Slider / RangeSlider | PageUp/PageDown 与 Shift+方向键走 10 个 Step，方向按横纵、反向、RTL 映射；OnChange 报告实时变化，OnCommit 在有变化的按键/拖拽结束时通知。范围拇指可越过彼此，自动排序并跟随焦点；不满足 Step × MinStepsBetweenThumbs 的移动被拒绝。Tooltip 挂在每个拇指上显示各自值 |
| Listbox | Loop、HighlightOnHover、SelectionBehavior；SelectionToggle（默认）在单选/多选下都能取消当前项，SelectionReplace 激活时替换；SelectMenu 单选固定使用替换语义。SelectionReplace 多选模式支持 Shift+方向键/Home/End 扩展连续范围并跳过禁用项；Ctrl/Cmd+A 选择可用项 |
| FileUpload | Variant=UploadArea / UploadButton；Layout=UploadList / UploadGrid；Dropzone、Interactive、FileDelete、Preview 及 Color / Size。与 Nuxt 一样，可交互的 UploadArea 整个区域就是可聚焦的按钮（名称取 Label，Enter/Space 或点击打开选择器），不再显示 Choose 按钮；Choose 按钮只在 UploadButton 或 `Interactive: nui.Set(false)` 时显示。关闭 Preview 只隐藏预览，不删除数据 |

`Optional` 的未设置和显式 false/0 有区别：`nui.Set(false)` 关闭默认开启项。当前默认值：Pagination PageSize=10、Siblings=2、ShowEdges=false、ShowControls=true、Size=MD；Carousel.Arrows=false；InputTags.AddOnPaste=false；InputRating.Hoverable=false；FileUpload.Layout=UploadGrid；Slider.Step=1。UploadLayout 的零值为 UploadLayoutDefault；使用枚举名，不依赖底层数字序号。

SelectMenu.SearchInput、重置策略、Listbox.HighlightOnHover、Accordion.Collapsible、Popover.Dismissible、Toast.Close / Progress、FileUpload 的四个布尔选项默认 true。

## InputMenu 值与搜索策略

- 单选使用 `SearchState.Selected` 保存 ID，`Query` 保存编辑文本。多选使用 `SelectedValues []string`，移除按钮和空草稿 Backspace 均可移除，变更通知 OnValuesChange；OnSelect 报告被激活的 ID。
- `ExternalFilter=true` 表示直接展示调用方给出的结果；FilterFields 支持 `label` / `description` / `value`，nil 默认仅 label。SearchSelect 兼容入口显式保留 label + description；请求和取消仍由 QueryState/应用服务处理。
- `InputMenuAutocomplete` 使用 Query 作为可自由输入的文本，Selected 保存最后选中的选项 ID；强制单选并允许自定义提交。
- OnCreate 提供可键盘激活的创建行，默认在无匹配时显示；CreateWhenAlways 在非空查询且无相同 value 时提供。创建请求保留 Query、筛选和打开状态，不触发普通选择，也不替应用插入选项。
- OpenOnClick 默认 true、OpenOnFocus 默认 false。两个 reset 选项默认 true；选择以及关闭（含 Escape、仍保持焦点）分别遵从 reset 策略；单选恢复所选标签、多选清空草稿；autocomplete 保留自由文本。
- 选项身份/标签/禁用集合改变时关闭旧弹层，避免底层索引高亮错误地选择另一条记录。ExternalFilter 不意味着更新结果时自动保持弹层打开。

## 示例

Gallery 的 **Props & behavior** 展示加载、选择上限、标签策略、手动 Tabs、查询重置、分页、浮层与文件预览；**Orientation** 展示方向切换。

## 使用限制

Listbox 两个方向均限制实际构建数量；横向项目与组标题采用固定 ItemWidth，纵向由原生 List 管理行高。Slider / RangeSlider 统一通过公开 Range / Step / 焦点和快捷键实现 Nuxt 的按键规则，因为 SliderBase 会先消耗按键；实际值、上下限和步长可访问，但 MyGo 通用控件缺少纵向元数据 setter，**此限制目前适用于所有纵向 Slider / RangeSlider，包括非反向模式**。InputRating 仍使用单焦点 slider 角色，未宣称完整 radio 读屏等价。

Go Options 与 Nuxt Vue Props 使用不同的数据模型，支持字段以各 API 文档为准。滑块值提示按拇指定位，未提供 Tooltip 延迟配置对象。FormField 通过本库控件的字段绑定处理标签激活；直接传入自定义 MyGo 控件时，需由应用处理其标签关联。表单校验、日期/数字格式与 Table/Tree 的支持范围见 [表单与数据](forms-data.md)，读屏、输入法与 DPI 的验证范围见 [兼容性说明](compatibility.md)。

Toast 的动作回调和 ButtonOptions.Style 会随通知保存，不能捕获临时 Element / Parts / Context；Style 的参数仅在当次构建有效。平台支持和输入法问题见 [兼容性说明](compatibility.md)。
