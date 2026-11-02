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
- Select、InputMenu、SelectMenu、Listbox、菜单、CommandPalette 共用 Nuxt 的选项行配方：内边距、gap、字号和图标随 Size 变化（MD 为 6/6/14/20），圆角 6；高亮行改为 Elevated 50% 底色加 Highlighted 文字；选中对勾放在行尾、跟随文字色；描述与标签同字号。Listbox 行高由 40 降为约 32，空状态居中，键盘聚焦时显示组件颜色的焦点环；SelectMenu 的 Size/Color 传给弹层里的搜索框和列表。
- 菜单中带 Color 的条目高亮时铺该颜色的 10% 底色；子菜单箭头放大到 20。
- 圆角对齐 Nuxt：Badge XS/SM、Kbd、Checkbox 改为 4；Listbox 改为 8。
- 标题字号对齐 Nuxt 继承的 16px：Card 由 14、Modal/Slideover/Drawer 由 18、Sidebar 由 14 改为 16。Drawer 的默认标题和描述现在使用 Label/Description 部件样式。
- 修复 SelectMultiple 选项带 Icon 时图标画了两次。
- 弹出层和 Tooltip 打开时有 100ms scale-in，Modal 有 200ms 缩放淡入并淡入遮罩；减少动态效果时直接显示。关闭仍立即完成。

### 组件外观

- Slider / RangeSlider 轨道加粗为 6–10；滑块改为背景色圆点加 2 DIP 外圈，焦点环向外偏移 2 DIP。
- Checkbox、RadioGroup、Switch 的文字行高与指示器容器一致（16/16/20/20/24），SM/LG 的 Checkbox 和 Radio 矮 2 DIP；Radio 内圆点固定为 4/4/6/6/8。
- CheckboxGroup / RadioGroup 项间距改为横向 8、纵向 2/4/6；ChoiceTable 相邻项共用边框，只有外侧圆角，横向时不换行。
- FormField 的必填星号改为单独的红色 `*`；描述紧贴标签，内容与标签间距 4，Error/Help 与内容间距 8，横向两列间距 8。

- Modal / Slideover 标题区最低高 64、上下内边距 16，与正文和 Footer 之间加分隔线；正文内边距 24，没有 content 时不创建正文区；Footer gap 6；关闭按钮改为 MD。上下方向的 Slideover 默认按内容高度显示（此前固定 448）。
- Toast 宽度默认 384、内边距 16；倒计时条改为 SM 并贴在底边；关闭按钮图标 20，悬停变为 Text 色。
- Tooltip 在文字和快捷键之间显示"·"。
- DropdownMenu / ContextMenu 默认按内容宽度（至少 128，此前固定 240）；选中项的对勾移到行尾、跟随文字色；Shortcut 按 `+` 拆成多个 Kbd 显示。

- Pagination 的首页/上一页/下一页/末页改为 chevron 图标方形按钮并随 RTL 镜像；页码按钮为方形；省略号改为与按钮同样外观的方块。英文可访问名称由 "Page First" 等改为 "First page"、"Previous page"、"Next page"、"Last page"（行为变化：按名称查找控件的测试或自动化脚本需要更新）。
- Breadcrumb 祖先链接改为 14 DIP 无内边距；当前项使用 Primary 色加半粗体；分隔符 20 DIP、Muted；间距 6。
- Tabs underline 变体增加贯穿列表的基线，活动指示线压在基线上；触发器间距随尺寸变化。
- Accordion / Collapsible 触发器去掉悬停底色和水平内边距，上下内边距 14，图标 20；Accordion 正文改为默认文字色，只保留 14 的下边距。

- SelectMenu 弹层改为平铺：搜索框为 None 变体加底部分隔线，列表无边框，默认最大高 240（此前 400）；占位文字改为 Dimmed；Subtle 变体悬停加深；清空按钮移到触发器尾部；创建项放进列表，方向键可以选到。
- InputMenu 输入框尾部增加 chevron 开关按钮；多选的已选值和 InputTags 的标签改为小芯片（水平内边距 6、垂直 2，字号随字段尺寸），只有 × 能删除。新增文案键 `showOptions`。

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
