# 更新日志

## 未发布

视觉与交互对齐 Nuxt UI v4.12.0。

### 迁移说明

- Sidebar 新增 Description 和 Actions，补齐桌面与弹层的实际部件及区域间距；空标题区不再占位，零宽或跨断点移除编辑内容后恢复可用入口焦点。正文、底部垂直内边距改为 16 DIP。
- HeaderLayout 的前导内容和标题文字容器增加样式入口，稳定区域身份保留动作编辑器，空标题不生成标题节点；Header/ Footer 动作默认内部间距为 6 DIP，Footer 的 Actions 配置现在作用于独立容器。
- FooterColumns 接入 Active、单项 OnSelect/PreventDefault 与单项样式，链接采用原生 Link 角色；列间距改为 32 DIP，首选最小列宽参与换行并受可用空间约束，各列顶部对齐。Gallery、网站和双语示例同步。

- Splitter 的 Disabled 改为只锁定分隔条；需要禁用面板内容时请禁用父容器。可见分隔线占 1 DIP，8 DIP 拖拽区域覆盖面板边缘，修正嵌套交点的空隙；新增 Orientation、Label、KeyboardResizeBy 与分隔条样式部件，修正边界重复通知，支持 Home / End，辅助范围使用绑定值单位。
- Marquee 新增 Repeat 和 Body 样式部件；未设置仍为 4 份，显式 0 不构建内容。父级禁用时停止运动，偏移随内容周期归一。两者的 Gallery、网站参数和双语示例已同步。

- Carousel 修正 Wrap=true 时拖拽意外循环、删除当前页后的焦点丢失与固定尺寸箭头居中问题。增加前后按钮/图标配置、WatchDrag、单项样式及部件；StopOnInteraction 默认 true，启用自动播放后手动导航会暂停，可显式设 false 保留自动恢复。计时和拖拽通知改在构建后发送，禁用与控件焦点会暂停自动播放；Gallery 与网站同步。

- Chart 默认高度改为 300 DIP，图例未设置时仅在多个系列或扇区时显示；圆环默认使用 24 DIP 厚度，显式 DonutInnerRadius 保留优先级。补齐坐标轴、数据点、提示、中心内容、配色和选择通知，修正 RTL 坐标与颜色循环；明确使用 MyGo 绘制及 TanStack Charts 参考边界。

- FileUpload 支持通配符、逗号分隔和复合扩展名；关闭 Dropzone 后不再注册拖放目标，父容器禁用也会丢弃迟到的选择结果。OnChange 仅报告实际队列变化，新增 OnReject；删除和取消后返回选择入口焦点。补齐图标、预览配件、删除按钮配置、Highlight、列表位置及样式部件，保留原生多文件队列。Gallery、网站参数、双语说明与复制程序同步。

- Drawer 默认不显示关闭按钮，Close 配置可隐式启用，CloseButton=true 可保留旧行为。OnSnap 只在索引实际变化时通知，禁止关闭的拖拽现在发送 OnClosePrevent。新增 Handle / Inset、可见指示条及整体滚动容器，修正最低高度下页脚裁切，Sidebar 组合保留固定页脚和关闭入口；关闭阈值须有限非负，0 使用默认值。Gallery、网站及双语说明同步。

- NavigationMenu 默认改为 Primary / Soft（pill），可用 Color=Neutral 保留中性文字；新增 LinkVariant、活动指示线和单项配件/样式。Compact 仅对纵向生效，横向方向键到边界停止。新增可取消导航、公开 Expanded、初始/同级展开控制及通知；修正焦点恢复与 RTL 缩进，Gallery、网站和双语说明同步。紧凑提示与弹层保留原生默认开启。

- CommandPalette 默认搜索新增描述与后缀，同组结果连续显示；修改上限、关闭重开、返回子级与禁用时正确取消旧请求。补齐延时、输入及图标配置、命令配件、可取消执行与保持打开，修正返回焦点和 Root 样式作用范围。默认上限仍为 200；替换自定义 Search 闭包后调用 Refresh。Gallery、网站和双语说明同步。

- Tree 的部分选中状态现在遵循 Multiple、BubbleSelect 和 PropagateSelect；选择回调替换模型后，旧点击不再展开新节点。新增 OnExpandedChange、节点样式及首尾内容入口，修正字符查找、Neutral 焦点和名称继承。已加载的空分支不再自动补文件夹/箭头，需要固定外观时显式设置 Icon；Gallery、网站和双语说明同步。

- Table 清除排序后恢复最近一次 SetRows 的输入顺序，比较值相等时也保持该顺序。List.SortBy 现在实际执行多列排序，移除或关闭排序的列会清理对应排序项。比较器依赖的外部条件改变后调用 InvalidateSort；需要自行排序数据时使用 ManualSorting。新增独立排序/选择通知、行激活、标题与加载内容，以及按列样式；Gallery 已改为使用本项目组件，网站与双语说明同步。

- ColorPicker 默认宽度从固定 280 DIP 改为 MD 的 168 DIP，增加五档尺寸。OnChange 默认按 50 毫秒合并，绑定颜色仍即时更新；需要逐次通知时设置 `Throttle: nui.Set(time.Duration(0))`。重新填入原通道值不再通知，RGBA 表单草稿按一个字段跟踪，实际通道编辑重建 sRGB 颜色。RTL 下原生颜色面板固定 LTR，修正滑条输入与绘制方向不一致；RGBA 字段仍继承页面方向。

- InputDateRange 清空两端只发送一次范围通知，新增 OnRangeChange；Calendar.OnChange 保留兼容。范围输入现在执行 Calendar 的完整范围限制。InputTime 在 12 小时制下增减小时会正确跨过中午和午夜。日期时间分段输入补齐头像、尾部加载、Neutral 高亮、步长与参考值；时间增加上下界、不可用谓词、精度和可选步长对齐。

- Calendar 默认再次选择同一天会取消，可设置 PreventDeselect=true 保留原行为；Home / End 改为周边界，默认显示年导航和标题视图切换。原有 WeekStart 默认周日保持不变。

- SelectMenu 隐藏搜索框后仍使用保留的查询词筛选；需要直接显示应用结果时设置 ExternalFilter 或 IgnoreFilter。清空按钮始终清除查询，ResetModelValueOnClear 只控制选中值。Root / Leading / Label / Trailing 样式不再作用于弹层选项，请使用对应 Item 部件；默认搜索框不再附加搜索图标。
- InputNumber 空值首次增减改为起始值，StepSnapping / FocusOnChange 默认 true；需要原有自由步进或按钮焦点时显式关闭对应选项。InputTags 的 Backspace 改为先选中再删除，Style.root 现在指整个字段。
- Form 的 LoadingAuto 默认 true，BeginAsync 后自动禁用内部控件；需要等待期间继续编辑时设置 `nui.Set(false)`。成功提交清理当前字段 Dirty，Reset 以下次构建的当前模型作为新基线。
- Breadcrumb 现在执行末项的 OnSelect；仅展示当前位置时将该回调留空。Banner 的容器内边距改由 Body 配置；Stepper 的步骤列表由 List 配置，Root 包含列表及新增正文。
- DropdownMenu / ContextMenu 的方向键默认在首尾停止；需要原有循环行为时设置 MenuOptions.Loop=true。UI.Trailing 现在作用于行尾容器，勾选图标可通过 Indicator 单独配置。
- ToastOptions.UI.Actions 现在配置操作容器；单个按钮样式放入 Actions 项的 ButtonOptions.UI。使用旧 Action 字符串并单独定制按钮样式时，可改为单项 Actions 列表。
- InputMenu 的 OpenOnClick 默认改为 false，与上游参数一致；需要点击即打开时显式设置 `nui.Set(true)`。SearchSelect 保留原有点击打开行为。
- `StepperOptions.Linear` 改为 `Optional[bool]`，默认 true。原布尔值需用 `nui.Set(...)` 包装；需要非线性导航时显式设置 `nui.Set(false)`。
- `InputTimeOptions.TwelveHour` 改为 `Optional[bool]`，默认随语言和区域格式。原布尔值需用 `nui.Set(...)` 包装；需要固定 24 小时制时设置 `nui.Set(false)`。
- FileUpload 文件大小文案改用 `fileSize`，参数 `size` 改为已格式化的字符串。覆盖过旧键 `fileSummary` 的应用需同步迁移。
- Pagination 英文可访问名称改为 `First page`、`Previous page`、`Next page`、`Last page`；按名称定位的测试或自动化脚本需更新。

### 新增与增强

- Calendar 补齐只读、颜色/样式/尺寸、月年控制、日/月/年导航、多月布局、周序号、星期格式、可变周数和内容构建器；修正键盘边界、禁用跳过、范围选中外观、实际改值通知和受控月份。增加不可用日期、范围上限、固定端点及预览规则，Gallery、网站和双语说明同步更新。

- Listbox 支持实例尺寸与颜色、Highlight、筛选及加载状态、选中图标、结构行、头像、圆点、单项样式和可取消选择；初始化与外部改值优先高亮已选项。Listbox / SelectMenu 搜索补齐忽略大小写及附加符号的匹配排序。
- SelectMenu 补齐触发器配件、清空按钮配置、搜索框配置和焦点选择、查询及开关通知；修复搜索框回车不能直接选择、触发器样式影响选项的问题。Gallery、网站、API 与双语说明同步新增筛选、取消选择、清空及模拟远程请求示例。
- InputRating 补齐 Indicator、IconSlot 和 EmptyIconSlot 样式部件，修正 Neutral 焦点颜色及只读光标；Gallery 与网站增加悬停、图标、样式和回调计数展示。
- InputNumber 支持起始值、步长对齐、增减按钮配置、滚轮与长按重复，修正十进制精度和方向边界。InputTags 支持行内换行、键盘标签选择、删除图标、值转换/显示与单项事件；PinInput 补齐 Highlight、分隔符样式及间距计算，修正 Neutral 焦点色和长占位文字。Gallery、网站参数和双语说明同步更新。
- Form 支持 LoadingAuto、OnError 和错误读取/替换/清理方法，修复异步重复提交、禁用请求延迟执行、全局错误遗漏及 Reset 使用旧值作为基线的问题；同步 Gallery、网站和复制程序的提交及取消示例。
- Breadcrumb 补齐颜色、分隔图标、头像、活动状态和项目样式，修复末项操作、静态项目焦点及截断分隔符。Pagination 支持控制图标和更细的部件样式。
- Stepper 支持单项 UI 和当前步骤正文，并恢复删除或切换后的焦点。Timeline 补齐日期与容器部件，修正日期行高、正文间距和横向等宽布局；两者修正 Neutral 背景。
- Banner 支持关闭配置、组件内关闭状态及整条通知动作，修正响应式对齐、限宽容器和部件身份。ErrorView 支持默认操作开关、按钮配置和状态码部件，修正窄窗口标题、图标、空内容及间距。相关 Gallery、网站和双语说明同步更新。
- Collapsible / Accordion 采用 Theme.Motion.Reduce，关闭中的正文立即停止接收输入，切换 KeepMounted 保留编辑器状态。Accordion 全局 Disabled 只禁用标题；补齐全局及单项 TrailingIcon、单项 UI 和 Header 部件，移除空正文留白，并在外部关闭或删除聚焦项后恢复焦点。Gallery、网站参数和双语说明同步更新。
- Tabs 支持头像、徽标、单项 UI、Content 开关和 Indicator 部件；修正自动模式响应错误方向键、纯文字行高、Neutral 指示颜色及纵向对齐。切换或删除聚焦面板后恢复标签焦点，正文可通过 Tab 进入；重排、布局变化和 KeepMounted 保留现有编辑状态。Gallery、网站参数与双语说明同步更新。
- ContextMenu 默认在定位点右侧 2 DIP 打开；修复连续右键误关闭、旧指针位置影响受控重开、目标布局移动带动菜单，以及禁用菜单连带禁用目标自身操作的问题。键盘定位使用持久 Handle 的已提交边界，Gallery 和网站同步增加定位及目标动作示例。
- DropdownMenu / ContextMenu 支持 XS–XL、头像、加载与勾选图标、项目 UI、保留打开的选择动作、勾选与根开关通知，以及单层位置和子菜单配置。补齐根菜单和子菜单筛选，修正描述行对齐、快捷键尺寸与间距、外部关闭和空菜单焦点。Gallery、网站参数、双语说明和复制程序同步更新；ContextMenu 筛选为原生扩展。
- Toast 支持 Avatar、CloseOptions、CloseIcon、ProgressOptions 和稳定动作 ID；修复顶部顺序、RTL 位置、空标题、横向对齐、操作间距、视口宽度约束、减少动态效果时的倒计时和来源字段禁用状态。网站及复制程序补齐 ToastViewport，并同步 Gallery、嵌套参数编辑与双语说明。
- Slideover 支持 Inset，四边尺寸扣除间距后限制到窗口范围；修复自定义遮罩颜色不参与渐变的问题。空正文继续保留默认内边距，顶部和底部的自动高度与上游一致；Gallery、网站及无效继承字段说明同步更新。
- Modal 支持 Close / CloseIcon、Transition 和 OnClosePrevent；修复遮罩动画末尾跳变、空标题区、面板文字与间距继承、圆角内容裁剪，并在标题区和全屏切换后保留正文编辑状态。共享区域与关闭配置同步作用于 Slideover / Drawer。
- Tooltip 支持可选 DelayDuration，显式 0 立即显示；修正快捷键间距、Panel.Gap 和外部按下关闭行为，快捷键同时追加到锚点辅助说明。Gallery 与网站提供延迟、宽度、受控开关及仅快捷键示例。
- Popover 支持面板最终文字样式和内容间距直接继承，修复更换锚点或受控打开后关闭焦点不正确，以及恢复动态效果时重新播放入场的问题；局部原生主题继续隔离。
- InputMenu 支持受控 Open、Loop、HighlightOnHover、ResetModelValueOnClear 和选择、标签删除、清空图标，补齐空草稿清空多选、自动补全文字回填及外部选中值显示。键盘删除标签后恢复输入焦点，全部禁用的候选不再显示无结果提示。
- Select / SelectMultiple 支持触发器图标、头像、加载位置与图标、Highlight、SelectedIcon 和持久 Open 状态，并补充前缀导航；Gallery、网站参数和双语用法同步。
- Button 支持可选前导头像，按按钮尺寸设置默认头像大小，并保持图标与加载状态的优先级。
- Badge 支持前导头像；空文字的标签自动使用四边相同的内边距，MD 纯图标标签宽度由 32 改为 24 DIP。
- Alert 支持优先于图标的头像、Close 按钮配置和 CloseIcon，并提供 Body / Actions 部件样式。
- AvatarGroup 支持 Item 部件，统一配置普通与溢出头像，并允许调整或移除外侧分隔环。
- ProgressGroup 支持项目图标、局部 UI 和自定义图例 Build，重排和切换方向时保留图例控件状态。
- Empty 支持头像以及 Header / Leading / Body 内容构建入口，补齐页头、正文、操作区与页脚部件样式。
- User 支持头像角标、Orientation、Leading / Body 内容与部件样式，保留自定义内容在头像和方向变化后的状态。
- CheckboxTri 支持 IndeterminateIcon；Checkbox 隐藏指示器时将显式图标显示在标签上方，并支持 Leading 部件样式。
- CheckboxGroup / RadioGroup 支持项目图标与 Loop 配置，重排时保留选择和项目身份。
- Sidebar 保留名称和调用签名，支持拖动与键盘调宽、双击复位、宽度记忆、左右布局及 RTL；默认宽度由 224 改为 256。
- Sidebar 新增可选状态、感知收起状态的内容插槽，以及窄窗口 Slideover / Modal / Drawer 模式和导航切换自动关闭；默认关闭调宽与响应式断点。
- 紧凑 NavigationMenu 支持 Tooltip 和侧向子菜单，选中子项时高亮祖先，并保留展开模式的分组状态。

### 默认行为变化

- ScrollArea 的内容按 Root 圆角裁剪，修复内容覆盖圆角的问题；Gallery 和网站补充方向、滚动条及位置控制。
- Select 选项图标与描述首行对齐，多选忽略空标签的分隔符；Neutral 焦点颜色、显式名称、关闭后触发器焦点和单选 Dismissible=false 行为已修正。Popover 内部主题容器不再遮盖面板背景、边框和圆角。
- Container 的默认水平内边距随窗口宽度使用 16 / 24 / 32 DIP，断点为 640 和 1024 DIP；显式 Padding 保持原行为。
- InputTags 默认仅以逗号分隔；需要恢复原分隔规则时设置 `Separators: ",;\n"`。
- FileUpload 的交互式 UploadArea 整体可聚焦，Enter / Space 打开选择器；Choose 按钮仅在 UploadButton 或 `Interactive: nui.Set(false)` 时显示。
- Marquee 默认启用两端渐隐，内容重复 4 份，短内容也会滚动；内容 builder 默认每次构建调用 4 次，可通过 Repeat 调整。
- InputRating 方向键到首尾停止，启用 Clearable 时支持键盘清除评分；CheckboxGroup 达到 Max 后，未选项仍可聚焦。
- CheckboxGroup 改为一个 Tab 停靠点，方向键只移动焦点，默认不循环；RadioGroup 默认循环。两类组的方向键跟随横纵布局，Loop 可显式改变循环规则。
- User 未配置头像时不再生成首字母头像；Carousel 页码计数改为可访问值，不再直接显示。
- 空状态文案统一为 `noData` / `noMatch`，兼容旧文案键的应用覆盖。

### 修复与样式调整

- 修复 Slider / RangeSlider 的 Neutral 与字段错误配色、双滑块部件颜色、小数步长尾数和父级禁用后的拖拽提交；修正极小最小间隔的判断。
- Slider 补齐本地化默认名称并保留显式名称优先级；自定义轨道内边距后，数值提示跟随实际圆点位置。
- 修复 FormField 无标签内容时的多余间距和独立星号、横向文字基线对齐，补充 Header / Body 区域样式；切换标签、方向和错误提示时保留编辑状态。
- Input / Textarea 补充前导头像和尾部加载入口，修复加载图标颜色、Neutral 焦点颜色及 None 样式的 Highlight；同步 Gallery、网站参数和双语说明。
- 修复 MaxLength 完全拒绝输入后仍触发 OnChange 的问题；明确 Textarea 中 Password、OnSubmit 和 MaxRows 的适用范围。
- 修复选择卡片同时显示外框与内部指示器焦点圈的问题；更正 Checkbox / Switch、单选组 / 多选组共享字段的参数说明与示例。
- 修复选择卡片状态覆盖自定义根样式和 Highlight 强调部件错误的问题。
- 修复选择卡片的悬停/焦点边框、部分选中卡片背景、指示器圆角覆盖和标签描述间距。
- 修复网站纵向选项卡片裁掉变更计数的问题；更正主题对比度说明的适用范围。
- 修复 Card 窄窗口内边距，以及只有描述时缺少顶部间距的问题。
- 修复 Card 移除页头后正文输入丢失焦点和选区的问题。
- 修复 Alert 切换前导内容或方向后操作控件失焦、空正文操作区间距错误，以及 Actions 样式不生效的问题。
- Avatar 的位图随根容器尺寸填充，并开放 Body 图片部件样式。
- Chip 的文字宽度按合并后的 Indicator 字体样式和水平内边距计算，避免自定义文字被裁切。
- Chip 的 Neutral 背景使用 InvertedBackground，修复自定义主题中错误沿用 Highlighted 的问题。
- 修复 Separator 默认头像尺寸、头像 Leading 样式、切换内容位置时的焦点保留，以及纵向长标签换行。
- Kbd 补齐 command / win 的符号映射；修复 Button、Badge、Alert、Kbd 的 Neutral / Solid 背景忽略 InvertedBackground 的问题。
- 修复 Skeleton 自绘背景忽略最终圆角的问题，Style 与返回元素设置的圆角现在均能生效。
- Progress 的 Neutral 填充使用 InvertedBackground；未知进度指示器长度按上游默认轮播修正为轨道的一半。
- 修复 ProgressGroup 的 Neutral 分段与色点配色，以及图例字号、图标和色点未随 Size 变化的问题；分段改为原生元素并开放对应部件样式。
- 修复 Empty 的响应式内边距、区域宽度与对齐、内容变化后的焦点保留，以及加载图标颜色覆盖；默认标题提供二级标题语义。
- 修复 User 空姓名产生的间距，补齐可激活状态的文字强调与头像缩放。
- 修复 Button、Progress、InputRating 和 User 在恢复动画时短暂回到旧颜色、进度或缩放值的问题。
- 修复 PinInput 连续输入、粘贴和光标显示，完善焦点与错误态；修复 FieldGroup 尺寸继承、圆角和 RangeSlider 的 FormField 关联。
- 修复 SelectMenu 创建行混入多选结果、SelectMultiple 图标重复，以及 Select / InputMenu 已选项高亮异常。
- 修复部件样式、自定义透明度和主题覆盖问题，完善 RTL 下的图标、圆角和边框方向。
- 统一输入、菜单、表单、弹层与导航的尺寸、间距、配色和圆角；Tooltip 长文本改为单行截断；完善弹层与进度动画，并遵循减少动态效果设置。

### 依赖

- MyGo 升级至 `v0.7.4` 正式版。

安装与示例见 [README](README.md)，API 与默认值见 [组件指南](docs/README.md)，平台差异见 [兼容性与已知限制](docs/compatibility.md)。
