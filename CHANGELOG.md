# 更新日志

## 未发布

按 Nuxt UI v4.12.0 对齐。以下为可见的行为变化，未列出的公开 API 保持不变。

### 破坏性变更与迁移

- `StepperOptions.Linear` 由 `bool` 改为 `Optional[bool]`，未设置时为 true（Nuxt 的 `linear: true`）。迁移：`Linear: true` 可直接删除或改为 `Linear: nui.Set(true)`；原来依赖默认非线性的 Stepper 需要显式写 `Linear: nui.Set(false)`。
- `InputTimeOptions.TwelveHour` 由 `bool` 改为 `Optional[bool]`，未设置时随语言和用户区域格式（en 为 12 小时，zh-Hans 为 24 小时）。迁移：`TwelveHour: true` 改为 `TwelveHour: nui.Set(true)`；原来依赖默认 24 小时的英文界面需要显式写 `TwelveHour: nui.Set(false)`。

### 默认值变更（D2–D15）

- D2：Toast 的 Actions 默认改为 XS、Solid、通知自身的颜色（此前 Neutral/Outline/SM）。
- D3：Popover 默认内边距改为 16（见"共享配方"）。D4：Modal/Slideover 的 Footer 保持右对齐，作为文档化差异。
- D6：User 的 AvatarOptions 各字段都为空时不再显示首字母头像。
- D7：FileUpload 的可交互 UploadArea 整个区域成为可聚焦按钮，Enter/Space 打开选择器，不再显示 Choose 按钮；Choose 按钮只在 UploadButton 或 `Interactive: nui.Set(false)` 时显示。
- D8：ErrorView 只在传入 Icon 时显示图标并改用 Primary 色；状态码、标题、说明改为 16 半粗体 Primary / 48 粗体 / 18；默认重试按钮改为 LG Primary Solid。
- D9：Kbd 修饰键保留桌面习惯（Win、Shift），只补齐按键符号并转大写。
- D10：空状态文案改为 Nuxt 的 `noData`（No data）和 `noMatch`（No matching data），新增两个文案键；旧键保留，应用对旧键的覆盖继续生效。
- D11：Marquee 默认间距 64，两端渐隐（新增 `Overlay Optional[bool]`，默认 true），内容重复 4 份并始终滚动；builder 每次构建调用 4 次。
- D12：Sidebar 默认宽 256（此前 224）。
- D13：CheckboxGroup 达到 Max 后未选项仍可聚焦，只是不能勾选。
- D14：InputTags 默认分隔符改为 `,`（此前 `,;\n`）。
- D15：Carousel 去掉可见的"1 / N"计数，当前位置写进根元素的可访问值（新增文案键 `slidePosition`）。

### 缺陷修复

- SelectMenu 的创建行不再进入全选或范围选择的结果，也不再传给业务条目的自定义 `Build`；仍支持鼠标、Enter 和 Space 创建。
- InputMenu 的尾部切换图标恢复主题、实例 `UI[Trailing]` 和 `Style` 回调的样式覆盖。
- Popover、菜单和 Tooltip 的打开动画保留自定义面板透明度；Modal 同样按目标透明度淡入，`Style` 回调仍为最终覆盖。
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
- 修复 Select、InputMenu 弹层中已选中但未高亮的行被填成强调色（MyGo 的 Selected 默认外观）；现在与 Nuxt 一样只显示行尾对勾。
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

- Button 的 Link 变体悬停不再加下划线；Block 按钮的尾部图标推到最右端。Link 键盘聚焦时只显示焦点环，不改文字色。
- Skeleton 底色改为 Elevated，并以 RoleStatus 和"加载中…"名称告知读屏。
- Kbd 转大写，补齐 tab、backspace、delete、capslock、方向键、pageup/pagedown、home/end 的符号。
- Empty 的图标放进 Elevated 圆形底（32–48），标题到描述间距 8。Card 标题到描述间距 4；Alert 纵向操作区上边距改为 6（与标题到描述的 4 合计为 Nuxt 的 10）；Avatar 图标为直径的 0.5 倍。
- Carousel 前后箭头改为叠在视口两侧的 Outline 圆形图标按钮（RTL 时交换），圆点为 12 DIP、间距 12，当前页为 inverted 色。
- Drawer 手柄改为 48×6、Accented 色。
- FileUpload 拖放区改为虚线边框，拖入时底色为 Elevated 25%；图标改为圆形 Avatar 底上的 upload 图标；文件大小按 Nuxt 格式显示为 B/KB/MB 等（新增文案键 `fileSize`，参数 `size` 为已格式化的字符串；原 `fileSummary` 键保留但不再使用，覆盖过它的应用需要改为覆盖 `fileSize`）。
- Header 最低高 64，标题 20 DIP 粗体；Footer 去掉顶部边框；FooterColumns 标题到列表间距 24、条目间距 16，链接去掉内边距。
- Banner 内容居中、单行、高 48；Actions 默认使用 Neutral XS 按钮。
- Calendar 星期表头改用 Primary 色；Tree 的文件夹图标随展开状态在 folder / folder-open 之间切换；Table 表头使用 Highlighted 色、单元格使用 Muted 色。
- Progress 数值变化有 200ms 过渡，减少动态效果时立即到位。

- InputRating 对齐 Nuxt：方向键到首尾停止，不再循环（行为变化：此前 5 再按 → 回到 1，空评分按 ← 跳到 5）；空格选中当前星，Clearable 时清零，键盘也能清除评分；键盘焦点环画在当前星上，组件不再横向拉伸到容器宽度；可交互时悬停的星放大到 1.1 倍；Hoverable 的预览在点击清零后立即消失，划过星间空隙不再闪回；在出错的 FormField 中星星使用 Error 色。

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
