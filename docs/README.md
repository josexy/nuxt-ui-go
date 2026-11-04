# 组件指南

第一次使用 NuxtUI Go，先按[安装与最小应用](../README.md#安装与最小应用)运行一个窗口。之后按要完成的任务查阅本文，无需依次读完所有组件。

## 从当前任务开始

| 你要做什么 | 建议阅读 |
| --- | --- |
| 设置窗口主题，添加按钮、卡片或状态提示 | [主题与基础组件](foundations.md) |
| 输入文字，添加标签、复选框或下拉选择 | [输入、选择与字段](inputs.md) |
| 打开对话框、菜单，切换页面或显示通知 | [弹层、菜单与导航](overlays-navigation.md) |
| 校验表单，处理日期，展示表格或树 | [表单、日期与数据组件](forms-data.md) |
| 添加侧栏、命令面板、文件选择或图表 | [高级交互、图表与布局](advanced.md) |

已经找到组件时，可用[参数与默认值](options.md)查询常用选项。[双语组件网站](../website/README.md)还提供逐组件参数编辑、Go/WASM 预览和可复制的完整程序。

## 理解组件的基本写法

多数组件的调用形式相似：传入当前界面上下文、主题、需要编辑的数据和选项。下面的代码放在视图函数中，`app.Name` 是应用长期保存的字符串字段：

```go
nui.Input(c.Key("name"), theme, &app.Name, nui.InputOptions{
    Label: "姓名",
    Placeholder: "输入姓名",
    Clearable: true,
}).FillWidth()
```

| 代码 | 含义 |
| --- | --- |
| `c *ui.Context` | 当前窗口正在构建界面时使用的上下文；由 MyGo 传入 |
| `theme nui.Theme` | 当前主题，通常由 `nui.App(c, config)` 返回 |
| `c.Key("name")` | 给控件一个稳定标识，帮助 MyGo 在重新构建时找到同一个控件 |
| `&app.Name` | 绑定应用保存的数据；用户输入后，组件会更新这个字段 |
| `nui.InputOptions{...}` | 组件配置；返回值上的 `FillWidth()` 让控件填满可用宽度 |

文档中的 `c`、`theme` 或 `t` 都沿用这些含义。`app`、`state` 是示意性的应用模型名，需要替换为你自己的变量。未写出 `package main` 的代码块是局部片段；完整导入和启动代码见[最小示例](../examples/minimal/main.go)。图标示例使用 `lucide`，导入方式见[图标文档](../icons/lucide/README.md)。

### 状态放在哪里

MyGo 会在输入、数据变化和动画期间反复调用视图函数。**输入值和组件状态要放在视图函数外保存**，否则每次构建都会重置。

应用可以保存业务数据、`AppConfig`、`SearchState`、`FormState` 等状态对象，以及用于焦点和位置查询的 `ui.Handle`。各窗口应分别保存自己的编辑和交互状态；已使用的组件状态不要随意复制。

组件返回的 `ui.Element` 和部件只在本次构建中有效。可以立即调用 `FillWidth()` 等方法，也可以传给本次构建的 Tooltip 或 Popover，但不能存入模型供下次使用。`*ui.Context` 的父级作用域会在构建中变化，也不要保存到模型或后台任务；需要长期使用的窗口服务，通过 `c.Services()` 获取。

动态列表应在创建控件前使用 `c.Key(业务ID)`。ID 要稳定、唯一，避免使用会随排序变化的下标或可修改的显示名称。

### 修改数据和执行任务

按钮点击、输入变化和表单提交分别通过 `OnClick`、`OnChange`、`OnSubmit` 等回调处理。它们适合修改数据和启动业务操作。

`Build`、`Content`、`Style` 等构建回调可能反复执行，只用于创建或配置当前界面。不要在其中无条件提交表单、修改正在遍历的集合或发起请求。

耗时计算和 I/O 放到后台任务中，结果通过 `Window.Update` 回到 UI 线程后再写入模型。离开页面或关闭窗口时取消任务；使用异步票据的组件还会检查结果是否已经过期，具体见 [Form 和 QueryState](forms-data.md)。

### 未设置与显式设置

有些选项的类型是 `Optional[T]`，用于区分默认行为和显式值。例如关闭默认启用的弹层关闭行为：

```go
options := nui.ModalOptions{
    Dismissible: nui.Set(false),
    CloseButton: nui.Set(false),
}
```

`Optional[T]{}` 表示未设置，`nui.Set(false)` 和 `nui.Set(0)` 表示明确使用 `false` 或 `0`。普通 `bool`、数值和枚举字段直接赋值，不需要包一层 `Set`。详见[参数与默认值](options.md#未设置零值和默认值)。

## 文档中常见的术语

| 术语 | 在本库中的含义 |
| --- | --- |
| 部件（slot） | 组件中可单独设置样式的区域，如 `Root` 外层、`Label` 文字、`Panel` 弹出面板；可用部件因组件而异 |
| 受控状态 | 应用持有并可直接修改的值，如 `open *bool`、当前页码或选中 ID |
| 构建回调（builder） | 在构建界面时调用的函数，用于填入正文、单元格或其他自定义内容 |
| DIP | 逻辑尺寸单位，由系统映射到屏幕像素；通常不需要手动乘屏幕缩放比例 |
| LTR / RTL | 从左到右 / 从右到左的书写方向；支持方向适配的组件会调整排列或按键行为 |

## 组件目录

下面列出 73 项组件与布局概念。名称用于查找；完整函数签名可在仓库中通过 `go doc` 查询，例如 `go doc . Input`。

### 主题与基础组件

[主题与基础组件](foundations.md)：App、Theme、Container、Main、Icon、Button、Link、Badge、Card、Alert、Avatar、AvatarGroup、Chip、Separator、Kbd、Skeleton、Progress、ProgressGroup、Empty、User。

### 输入与选择

[输入、选择与字段](inputs.md)：Input、Textarea、FormField、FieldGroup、Checkbox、CheckboxGroup、RadioGroup、Switch、Slider、Select、InputMenu、Popover、ScrollArea。

### 弹层与导航

[弹层、菜单与导航](overlays-navigation.md)：Tooltip、Modal、Slideover、Toast、DropdownMenu、ContextMenu、Tabs、Collapsible、Accordion、Breadcrumb、Pagination、Stepper、Timeline、Banner、ErrorView。

### 表单与数据

[表单、日期与数据组件](forms-data.md)：Form、InputNumber、InputTags、PinInput、InputRating、Listbox、SelectMenu、Calendar、InputDate、InputTime、ColorPicker、Table、Tree。

### 高级交互与布局

[高级交互、图表与布局](advanced.md)：CommandPalette、NavigationMenu、Drawer、FileUpload、Chart、Carousel、Marquee、Splitter、Sidebar、HeaderLayout、FooterLayout、FooterColumns。

Go 中的 `Error` 已用于语义颜色，因此错误展示组件命名为 `ErrorView`；`Header`、`Footer` 是部件名，对应布局函数命名为 `HeaderLayout`、`FooterLayout`。此外还有 `CheckboxTri`、`RangeSlider`、`SelectMultiple`、`InputDateRange` 等相关函数，用法在对应组件章节中说明。

## 其他资料

| 需要查什么 | 文档 |
| --- | --- |
| 中英文切换和文案覆盖 | [国际化](i18n.md) |
| Lucide 图标名称和用法 | [图标文档](../icons/lucide/README.md) |
| 平台支持、输入法和无障碍限制 | [兼容性与已知限制](compatibility.md) |
| 可运行示例和开发检查 | [运行示例](../README.md#运行示例)、[Taskfile.yml](../Taskfile.yml) |
| 版本变化 | [更新日志](../CHANGELOG.md) |
