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

受影响的入口包括 Input、Textarea、InputMenu、InputTags、PinInput 文本模式、SelectMenu、ColorPicker 内部 HEX 输入框和 CommandPalette。应用需要编辑重要文本时，应在目标系统的输入法上验证这个过程。

`PinInputOptions{Numeric: true}`、InputNumber、InputDate、InputTime 使用数字或日期时间格式，不启用 IME。其他输入方式仍需按目标平台检查。

## 功能边界

### 外观、输入和弹层

| 范围 | 支持内容 | 使用时注意 |
| --- | --- | --- |
| 主题与尺寸 | DIP、原生字体、XS/SM/MD/LG/XL 尺寸 | 字体和回退字体可能增加行高；无 Tailwind/CSS 运行时；原生禁用透明度为 50%，部分尺寸、状态和动画与 Nuxt 不同 |
| 语言与输入 | 英文、简体中文、公历日期、区域分段顺序、12/24 小时制 | 数字使用 ASCII，小数点为 `.`；无完整地区数字输入、多历法、时间范围和分段撤销/重做 |
| 弹层 | Tooltip、Popover、Modal、Slideover 等 | Tooltip 悬停延迟为 600ms；Popover、Select、菜单、Tooltip、Modal 关闭时立即消失；Slideover 有退出动画 |
| 通知 | Toast 队列、动作按钮、超时和暂停 | 同时最多显示 3 条，较早的排队项仍会计时 |

### 数据和文件

| 范围 | 支持内容 | 尚未提供或需由应用实现 |
| --- | --- | --- |
| Table / Tree | 原生虚拟列表、稳定 ID；Tree 支持已加载节点的父子选择联动 | 完整 TanStack 配置、冻结列、多层表头、行分组聚合、树拖拽；Tree 迁移说明见[组件文档](forms-data.md#tree) |
| Chart | 6 类图、分类/数值/时间轴、堆叠、图例交互 | 完整 Unovis 功能、双 Y 轴、对数轴、刷选缩放 |
| FileUpload | 原生文件选择、规则检查、进度和取消状态 | 网络上传、文件内容验证、文件流和持久化 |

### 布局和导航

| 范围 | 当前行为与限制 |
| --- | --- |
| Drawer | 通过底部独立手柄拖动，不支持四边拖拽或正文滚动联动手势 |
| Carousel / Splitter | Carousel 可从末项回到首项，拖拽仍有边界；Splitter 只分成两个面板 |
| NavigationMenu / Sidebar | 支持图标导航、Tooltip、子级 Popover、调宽和按窗口 DIP 宽度切换弹层；路由和持久化由应用实现 |
| 布局切换 | Sidebar 跨断点会卸载并重建控件；应用模型可以保留，控件局部滚动状态不会自动迁移 |
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
