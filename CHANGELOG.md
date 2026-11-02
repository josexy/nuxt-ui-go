# 更新日志

## 未发布

按 Nuxt UI v4.12.0 对齐。以下为可见的行为变化，未列出的公开 API 保持不变。

### 缺陷修复

- CheckboxGroup 的 Label、Description、Indicator 等部件样式和 Style 回调现在作用于每个 Checkbox；Root/Item 仍只作用于组和行。
- Checkbox 使用 `CheckedIcon` 作为勾选图标（默认 Check）；`UncheckedIcon` 只用于 Switch。
- Switch 轨道去掉 1 DIP 描边，只在 Highlight 或出错时画一圈强调色/错误色。
- FieldGroup 未设置 Size 时不再强制 MD，而是沿用外层 FormField 的尺寸；首尾圆角按书写方向镜像；项目拉伸到同一高度。Badge、SelectMenu、SelectMultiple 加入组圆角，Badge 继承字段/组尺寸。组内控件弹出的面板不再套用组圆角。
- RangeSlider 接入 FormField：点击字段标签聚焦第一个拇指，并带上字段描述和错误。
- PinInput 在出错的 FormField 中显示错误色边框。
- Breadcrumb 当前项、FooterColumns 条目显示 `Icon`。
- Card 没有正文时不再渲染空正文区和多余分隔线；Badge 标签为空时不占间距。
- Chart 默认系列色改为 Nuxt 的六种不重复颜色：Primary、blue、violet、amber、rose、cyan（浅色 500 / 深色 400）。此前第 2 条和第 5 条系列同为 blue。
- Banner 关闭按钮悬停改为 10% 背景色；Alert 关闭按钮改为无内边距的 MD link 按钮。两者悬停时不再出现遮住反色图标的不透明底色。
- CommandPalette 在标签下显示条目描述；Listbox 行可随内容增高。
- 右到左布局下，Breadcrumb 分隔符、Calendar 翻页按钮、Sidebar 收起按钮的图标随方向镜像；Sidebar 收起按钮改用面板图标，分隔边框放在结束侧。

### 共享配方

- 弹出层面板（Popover、Select、InputMenu、SelectMenu、DropdownMenu、ContextMenu）圆角由 8 改为 6；Modal、Slideover、Toast 保持 8。
- Popover 默认内边距由 12 改为 16（D3：贴近 Nuxt 文档示例的 `p-4`）。
- Tooltip 改为高 24 的单行，超长截断；圆角 4、水平内边距 10、小阴影。此前长文本会换行。
- 弹出层和 Tooltip 打开时有 100ms scale-in，Modal 有 200ms 缩放淡入并淡入遮罩；减少动态效果时直接显示。关闭仍立即完成。

### 依赖

- MyGo 从 `v0.7.4-0.20261031084912-402a3cf8eb35` 升级到 `v0.7.4` 正式版。新增的 API 只有增量，没有删除导出 API；滚动条统一绘制后仍读取 `Theme.Scrollbar` 和 `ScrollbarWidth`。
- 所有文本输入的占位文字改为 Dimmed 色（此前为 MyGo 默认的 TextMuted，即 Muted 色），与 Nuxt 一致。

## 当前内容

采用 MIT 许可证，提供 Nuxt UI v4 风格的 MyGo 原生组件。

- 73 项组件与布局入口，覆盖主题、基础展示、输入、选择、表单、弹层、导航、日期、表格、树及高级交互。
- 浅色、深色、系统模式和局部主题；26 套色板、语义色、部件样式及减少动态效果。
- 2,015 个按需使用的 Lucide SVG 图标。
- 英文和简体中文组件文案、全局语言切换与应用文案覆盖。
- 命令搜索、原生文件选择、6 类图表、轮播、滚动展示和桌面布局。
- Gallery、最小应用、设置页和数据浏览器示例。

平台支持与已知输入法问题见 [兼容性与已知限制](docs/compatibility.md)。

安装与完整示例见 [README](README.md)，API 见 [组件指南](docs/README.md)。
