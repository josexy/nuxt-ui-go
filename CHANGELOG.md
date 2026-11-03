# 更新日志

## 未发布

视觉与交互对齐 Nuxt UI v4.12.0。

### 迁移说明

- `StepperOptions.Linear` 改为 `Optional[bool]`，默认 true。原布尔值需用 `nui.Set(...)` 包装；需要非线性导航时显式设置 `nui.Set(false)`。
- `InputTimeOptions.TwelveHour` 改为 `Optional[bool]`，默认随语言和区域格式。原布尔值需用 `nui.Set(...)` 包装；需要固定 24 小时制时设置 `nui.Set(false)`。
- FileUpload 文件大小文案改用 `fileSize`，参数 `size` 改为已格式化的字符串。覆盖过旧键 `fileSummary` 的应用需同步迁移。
- Pagination 英文可访问名称改为 `First page`、`Previous page`、`Next page`、`Last page`；按名称定位的测试或自动化脚本需更新。

### 新增与增强

- Sidebar 保留名称和调用签名，支持拖动与键盘调宽、双击复位、宽度记忆、左右布局及 RTL；默认宽度由 224 改为 256。
- Sidebar 新增可选状态、感知收起状态的内容插槽，以及窄窗口 Slideover / Modal / Drawer 模式和导航切换自动关闭；默认关闭调宽与响应式断点。
- 紧凑 NavigationMenu 支持 Tooltip 和侧向子菜单，选中子项时高亮祖先，并保留展开模式的分组状态。

### 默认行为变化

- InputTags 默认仅以逗号分隔；需要恢复原分隔规则时设置 `Separators: ",;\n"`。
- FileUpload 的交互式 UploadArea 整体可聚焦，Enter / Space 打开选择器；Choose 按钮仅在 UploadButton 或 `Interactive: nui.Set(false)` 时显示。
- Marquee 默认启用两端渐隐，内容重复 4 份，短内容也会滚动；内容 builder 每次构建调用 4 次。
- InputRating 方向键到首尾停止，启用 Clearable 时支持键盘清除评分；CheckboxGroup 达到 Max 后，未选项仍可聚焦。
- User 未配置头像时不再生成首字母头像；Carousel 页码计数改为可访问值，不再直接显示。
- 空状态文案统一为 `noData` / `noMatch`，兼容旧文案键的应用覆盖。

### 修复与样式调整

- 修复 PinInput 连续输入、粘贴和光标显示，完善焦点与错误态；修复 FieldGroup 尺寸继承、圆角和 RangeSlider 的 FormField 关联。
- 修复 SelectMenu 创建行混入多选结果、SelectMultiple 图标重复，以及 Select / InputMenu 已选项高亮异常。
- 修复部件样式、自定义透明度和主题覆盖问题，完善 RTL 下的图标、圆角和边框方向。
- 统一输入、菜单、表单、弹层与导航的尺寸、间距、配色和圆角；Tooltip 长文本改为单行截断；完善弹层与进度动画，并遵循减少动态效果设置。

### 依赖

- MyGo 升级至 `v0.7.4` 正式版。

安装与示例见 [README](README.md)，API 与默认值见 [组件指南](docs/README.md)，平台差异见 [兼容性与已知限制](docs/compatibility.md)。
