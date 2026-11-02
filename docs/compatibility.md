# 兼容性与已知限制

NuxtUI Go 是采用 Nuxt UI v4 视觉风格的原生 Go 组件库，使用 MyGo 布局、输入和绘制。它提供文档中列出的 Go API，不兼容 Vue 组件导入、CSS 类、Nuxt 路由或浏览器插件。

API 和默认行为可能随更新调整，后续变化会记录在 [更新日志](../CHANGELOG.md)。

## 环境与平台

需要 Go 1.27.3 或更高版本，应用以 `CGO_ENABLED=0` 构建。MyGo 运行库及工具固定到 `v0.7.4`；具体依赖见 [go.mod](../go.mod)。运行时不依赖 Node、Vue 或 WebView。

| 平台 | 当前状态 |
| --- | --- |
| Windows amd64 | 构建、组件自动化与真实 Gallery 窗口验证通过；完整系统输入法、跨屏 DPI 和读屏体验尚未验证 |
| Linux amd64 | WSL Ubuntu 22.04 下的构建、组件自动化与原生隐藏窗口首帧检查通过；完整桌面交互、系统输入法及其他发行版尚未验证 |
| macOS arm64 | 交叉构建通过，桌面运行尚未验证 |
| 其他架构 / WASM | 尚未验证 |

原生应用需要桌面图形会话。原生文件选择取消、多选、过滤器、跨应用拖放与关闭期间返回结果，仍需在目标平台检查；macOS sandbox 权限和 bookmark 由应用管理。

## 输入法取消

当前 MyGo 依赖存在已知问题：**选中文字后开始 IME 组合输入，再取消组合，原选中文字可能丢失。** 可通过撤销恢复原选中文字。

受影响的文本入口包括 Input、Textarea、InputMenu、InputTags、PinInput 文本模式、SelectMenu、ColorPicker 内部 HEX 输入和 CommandPalette。涉及重要文本的应用应在目标系统输入法上验证该流程。

`PinInputOptions{Numeric: true}` 的数字模式关闭 IME。InputNumber、InputDate、InputTime 也采用明确的数字或日期时间格式，不启用 IME；这不代表其他输入方式和平台已完成全部验证。

## 功能边界

| 范围 | 当前行为与限制 |
| --- | --- |
| 主题与尺寸 | 使用 DIP 和原生字体，通用 Size 为 XS/SM/MD/LG/XL；字体及回退字体的度量可能增加行高和控件高度。不提供 Tailwind/CSS 运行时。原生禁用透明度为 50%，部分尺寸、状态和动画与 Nuxt 不同 |
| 输入与国际化 | 内置 en/zh-Hans；日期为公历 YYYY-MM-DD，数字使用小数点，时间按 Options 选择 12/24 小时。没有完整地区格式或多历法 |
| 弹层与通知 | Tooltip 原生悬停延迟为 600ms；Toast 同时最多显示 3 条；Modal/Slideover 使用文档中的固定过渡与焦点策略。Popover、Select、菜单、Tooltip 和 Modal 只有打开动画，关闭立即完成 |
| Table / Tree | 使用原生数据控件与稳定 ID；不覆盖全部 TanStack 配置、冻结列、多层表头、行分组聚合或树拖拽级联选择 |
| Chart | 支持 6 类图、分类/数值/时间轴、堆叠和图例；使用 Go 数据模型，不覆盖全部 Unovis 功能、双 Y 轴、对数轴或刷选缩放 |
| Drawer / Carousel / Splitter | Drawer 仅底部手柄拖拽；Carousel Wrap 是有界翻页，不是无缝克隆循环；Splitter 为双面板，不提供多面板布局协议 |
| FileUpload | 提供原生文件选择、规则检查、进度和取消状态；网络上传、内容验证、文件流和持久化由应用实现 |
| NavigationMenu / 应用布局 | 导航路由由应用管理；没有 Web mega menu、移动断点、DOM 插槽或 NuxtLink 属性兼容层 |
| 高级界面 | 未提供完整 Page、Dashboard、AI Chat、Editor、Content、颜色模式专用组件或 LocaleSelect；可直接使用主题和语言 API |

每个组件的 Options、默认值、键盘行为和部件范围以 [组件文档](README.md) 为准。

## 无障碍与动态效果

组件沿用 MyGo 的焦点、键盘和原生语义，支持减少动态效果及主题偏好；完整屏幕阅读器体验尚未验证。多选列表的完整多选元数据、纵向 Slider/RangeSlider 的方向元数据、InputRating 的 slider 角色与网页 radio 语义差异，以及 Marquee 重复内容的读屏呈现仍有限制。

显示文本、图标和自定义 builder 时，应用应保留名称、焦点和交互语义。自定义颜色和业务组合需要各自验证对比度与可访问性。
