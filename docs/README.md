# 组件指南

[双语组件网站](../website/README.md) 提供逐组件参数配置、真实 Go/WASM 预览和完整代码；以下指南继续说明原生 Go API 和使用约定。

NuxtUI Go 提供原生 Go 组件、主题与布局入口。先按 [快速开始](../README.md#安装与最小应用) 创建窗口，再根据下表选择组件；公共配置见 [参数与默认值](options.md)，平台支持及已知问题见 [兼容性说明](compatibility.md)。

## 组件目录

| 分类 | 组件与入口 | 文档 |
| --- | --- | --- |
| 主题与基础组件 | App、Theme、Container、Main、Icon、Button、Link、Badge、Card、Alert、Avatar、AvatarGroup、Chip、Separator、Kbd、Skeleton、Progress、ProgressGroup、Empty、User | [主题、样式与基础组件](foundations.md) |
| 输入与选择 | Input、Textarea、FormField、FieldGroup、Checkbox、CheckboxGroup、RadioGroup、Switch、Slider、Select、InputMenu、Popover、ScrollArea | [输入、选择与字段](inputs.md) |
| 弹层与导航 | Tooltip、Modal、Slideover、Toast、DropdownMenu、ContextMenu、Tabs、Collapsible、Accordion、Breadcrumb、Pagination、Stepper、Timeline、Banner、Error | [弹层、菜单与导航](overlays-navigation.md) |
| 表单与数据 | Form、InputNumber、InputTags、PinInput、InputRating、Listbox、SelectMenu、Calendar、InputDate、InputTime、ColorPicker、Table、Tree | [表单、日期与数据](forms-data.md) |
| 高级交互与布局 | CommandPalette、NavigationMenu、Drawer、FileUpload、Chart、Carousel、Marquee、Splitter、Sidebar、Header、Footer、FooterColumns | [高级交互、图表与布局](advanced.md) |

上述目录包含 73 项组件和布局概念。部分入口按 Go 命名约束调整：Error 使用 `ErrorView`，Header/Footer 使用 `HeaderLayout`/`FooterLayout`。另外提供 `CheckboxTri`、`RangeSlider`、`SelectMultiple`、`InputDateRange` 等相关入口，分别在对应文档说明。

## 常见任务

| 任务 | 入口 |
| --- | --- |
| 设置主色、明暗模式、局部主题与组件样式 | [主题与样式](foundations.md#应用与主题) |
| 设置方向、默认值、选择策略和浮层行为 | [参数与默认值](options.md) |
| 添加字段标签、校验、提交和异步结果处理 | [输入字段](inputs.md#formfield-与-fieldgroup)、[Form](forms-data.md#form) |
| 添加图标 | [Lucide 使用说明](../icons/lucide/README.md) |
| 切换语言或覆盖组件文案 | [国际化](i18n.md) |
| 查看平台支持、输入法问题与 Nuxt 的功能差异 | [兼容性与已知限制](compatibility.md) |
| 运行示例与验证 | [示例](../README.md#运行示例)、[Task 任务](../Taskfile.yml) |
| 查看版本变化 | [更新日志](../CHANGELOG.md) |

文档中的代码片段默认位于视图构建函数内，`c` 为当前 `*ui.Context`、`theme` 或 `t` 为 `nui.Theme`，状态字段由应用持有。完整导入与启动方式见 [最小应用](../examples/minimal/main.go)。可用字段、类型及函数签名也可通过 `go doc github.com/josexy/nuxt-ui-go` 查询。
