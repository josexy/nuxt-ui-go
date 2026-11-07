# 兼容性与已知限制

选用组件前，先确认目标平台的验证范围，以及应用是否需要本文列出的未支持功能。

NuxtUI Go 使用 MyGo 的原生布局、输入和绘制，外观参考 Nuxt UI v4。开发时使用本库的 Go API；Vue 组件导入、CSS 类、Nuxt 路由和浏览器插件不能直接用于原生应用。API 和默认值可能随版本调整，变化记录在[更新日志](../CHANGELOG.md)。

## 环境与平台

需要 Go **1.27.3 或更高版本**，构建时使用 `CGO_ENABLED=0`。本仓库将 MyGo 运行库和工具固定为 `v0.7.4`，具体见 [go.mod](../go.mod)。原生应用运行时不需要 Node.js、Vue 或 WebView，但需要桌面图形会话。

下表记录项目已有的验证范围，不表示每个平台都经过相同程度的测试。

| 平台 | 已验证 | 尚未验证 |
| --- | --- | --- |
| Windows amd64 | 构建、组件自动化测试、真实 Gallery 窗口 | 完整系统输入法流程、跨屏 DPI、屏幕阅读器体验 |
| Linux amd64 | WSL Ubuntu 22.04 下的构建、组件自动化测试、原生隐藏窗口首帧 | 完整桌面交互、系统输入法、其他发行版 |
| macOS arm64 | 交叉构建 | 桌面运行和交互 |
| 浏览器 WASM 预览 | 项目提供[组件网站预览与检查](../website/README.md) | 原生文件对话框及完整辅助技术体验不在浏览器预览的验证范围内 |
| 其他架构 | 暂无验证记录 | 需在目标环境检查 |

原生文件选择的取消、多选、过滤器、跨应用拖放，以及窗口关闭后返回的结果，都需要在目标系统检查。macOS 沙盒权限和文件访问 bookmark 由应用管理。

## 输入法取消

当前 MyGo 依赖有一个已知问题：**选中文字后开始输入法（IME）组合输入，再取消组合，原选中文字可能丢失。** 可通过撤销恢复。

受影响的入口包括 Input、Textarea、InputMenu、InputTags、PinInput 文本模式、Listbox 筛选框、SelectMenu、ColorPicker 内部 HEX 输入框和 CommandPalette。应用需要编辑重要文本时，应在目标系统的输入法上验证这个过程。

`PinInputOptions{Numeric: true}`、InputNumber、InputDate、InputTime 使用数字或日期时间格式，不启用 IME。其他输入方式仍需按目标平台检查。

## 功能边界

### 外观、输入和弹层

| 范围 | 支持内容 | 使用时注意 |
| --- | --- | --- |
| 主题与尺寸 | DIP、原生字体、XS/SM/MD/LG/XL 尺寸 | 字体和回退字体可能增加行高；无 Tailwind/CSS 运行时；原生禁用透明度为 50%，部分尺寸、状态和动画与 Nuxt 不同 |
| 语言与输入 | 英文、简体中文、公历日期、区域分段顺序、12/24 小时制 | 数字使用 ASCII，小数点为 `.`；无完整地区数字输入、多历法、时间范围和分段撤销/重做 |
| 分段日期时间 | 日期单值与范围、地区顺序、步长、日期/时间边界及不可用谓词、时间精度、原生可选日历 | Date 不包含时刻；ClockTime 不包含日期、毫秒或时区；时间范围用两个 InputTime 组合，完整读屏体验未验证 |
| ColorPicker | sRGB 字节通道、HEX/HEXA、原生键盘、透明度、色板、尺寸与通知节流 | 采用 MyGo 横向轨道；内部 SV/轨道/HEX 部件未公开，无 HSL/CMYK/Lab 文本编辑或广色域编辑，完整读屏体验未验证 |
| 弹层 | Tooltip、Popover、Modal、Slideover 等 | Tooltip 悬停默认延迟为 600ms，可用 DelayDuration 覆盖；Popover、Select、菜单、Tooltip、Modal 关闭时立即消失；Slideover 有退出动画 |
| 通知 | Toast 队列、动作按钮、超时和暂停 | 同时最多显示 3 条，较早的排队项仍会计时 |

### 数据和文件

| 范围 | 支持内容 | 尚未提供或需由应用实现 |
| --- | --- | --- |
| Calendar | 公历单日/范围、日/月/年导航、多月换行、范围限制 | 不提供离散多日期或独立年月值模型；WeekStart 默认周日，周序号按显式周起点计算，不承诺完整 Intl locale 周规则 |
| Table / Tree | 原生虚拟列表、稳定 ID；Table 支持多列/手动排序、独立模型通知与加载内容；Tree 支持已加载节点的父子选择联动 | Table 表头无独立键盘焦点；完整 TanStack 配置、冻结列、多层表头、行分组聚合、树拖拽未提供；见 [Table 文档](forms-data.md#table)及 [Tree 文档](forms-data.md#tree) |
| Listbox / SelectMenu | 稳定字符串 ID、结构行、头像与圆点、过滤排序、单项取消动作和原生虚拟列表 | 不接收任意 Vue 对象或动态属性路径；搜索使用 Unicode 大小写折叠及附加符号归一化，不等同于完整 Intl 地区排序；虚拟化无 overscan / estimateSize 公开配置 |
| Chart | 6 类原生图形、分类/数值/时间轴、堆叠、交互图例、坐标与提示配置、RTL 和行/扇区选择通知 | 参考 TanStack Charts；不提供任意 definition、独立 Web point/group 对象、双 Y 轴、对数轴或刷选缩放，见 [Chart 文档](advanced.md#chart) |
| FileUpload | 原生选择与队列、类型/大小/数量规则、拒绝通知、应用提供的预览配件、进度和取消状态 | 网络上传、内容验证、图片加载、文件流和持久化；Dialog.Multiple 只控制单次选择，网站使用模拟 Choose，详见 [FileUpload 文档](advanced.md#fileupload-与资源所有权) |

### 布局和导航

| 范围 | 当前行为与限制 |
| --- | --- |
| Drawer | 通过底部独立手柄拖动，内容整体滚动；不支持四边拖拽、Vaul 背景缩放或正文滚动联动手势。SnapPoints 只接受高度比例，原生关闭阈值与动画另行说明 |
| Carousel / Splitter | Carousel 的按钮/键盘/自动播放可循环，拖拽仍有边界；Active 只门控自动播放，无完整 Embla 插件或 ARIA tablist；Splitter 只分成两个面板 |
| NavigationMenu / Sidebar | 支持图标导航、Tooltip、子级 Popover、调宽和按窗口 DIP 宽度切换弹层；路由和持久化由应用实现 |
| 布局切换 | Sidebar 跨断点会重建控件，保留应用模型；移除当前编辑内容后恢复可用入口焦点，控件局部滚动及选择状态不自动迁移 |
| Header / Footer / FooterColumns | 原生标题区与页脚使用构建回调，不包含网页全部插槽、路由和粘性定位。Header 菜单由 Sidebar 等组件组合；FooterColumns 的 Active、单项可取消导航及样式已接入，其他 NavigationItem 字段另有适用说明 |
| 其他 Nuxt 功能 | 没有 Web mega menu、CSS 断点、DOM 插槽或 NuxtLink 兼容层；也未提供完整 Page、Dashboard、AI Chat、Editor、Content、颜色模式专用组件和 LocaleSelect，可直接调用主题与语言 API |

各组件的可用选项、默认值、键盘操作和样式部件以[组件文档](README.md)为准。

## 无障碍与动态效果

组件使用 MyGo 的焦点、键盘和原生辅助技术语义，支持系统主题偏好及减少动态效果。完整屏幕阅读器体验尚未验证，尤其需要注意以下范围：

| 范围 | 当前限制 |
| --- | --- |
| 多选列表 | 底层未开放完整的多选元数据 |
| 纵向 Slider / RangeSlider | 缺少方向元数据，包括未反向的纵向模式 |
| InputRating | 使用单焦点 slider 角色，与网页的 radio 语义有差异 |
| Marquee | 重复内容对屏幕阅读器的呈现仍需验证 |
| Tree | 提供 Tree/TreeItem、层级、选择和活动项语义；当前虚拟列表没有 `ui.Outline` 内置的辅助技术展开/折叠动作 |

自定义内容时，应保留控件名称、焦点和交互语义。自定义颜色需要单独检查对比度。鼠标、键盘或 Canvas WASM 预览的测试，只覆盖相应交互，不能替代目标系统上的屏幕阅读器检查。
