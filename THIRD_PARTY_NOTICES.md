# 第三方来源

## Nuxt UI

组件外观配方参考 [nuxt/ui](https://github.com/nuxt/ui/tree/200b0bfb7a5666c870742f3f43c83cf09234aed9)，版本 4.12.0，提交 `200b0bfb7a5666c870742f3f43c83cf09234aed9`。

Copyright (c) 2023 Nuxt，MIT License。完整许可见 [nuxt-ui.txt](third_party/licenses/nuxt-ui.txt)。运行时不包含 Vue 或 Tailwind。

## Lucide

完整收录 [lucide-icons/lucide](https://github.com/lucide-icons/lucide/tree/a630cec77eac6568b989e87e4947a8bfe45caacd/icons) 固定提交 `a630cec77eac6568b989e87e4947a8bfe45caacd` 的 `icons/` 目录中全部 **2,015 个 SVG**；文件名与逐文件 SHA-256 见 [manifest.json](icons/lucide/manifest.json)。

完整 ISC 许可及派生 Feather 的 MIT 声明见 [icons/lucide/LICENSE](icons/lucide/LICENSE)。SVG 原样保留，Go 包增加按图标惰性解析缓存。

## Tailwind CSS 色板

色阶来自 [Tailwind CSS v4.3.3 theme.css](https://github.com/tailwindlabs/tailwindcss/blob/v4.3.3/packages/tailwindcss/theme.css)。生成器使用的原始数据保存在 [theme.css](third_party/tailwind/theme.css)，SHA-256 为 `443d7af364200f8ec8352dc78e39485a118a1840d498c97967bf0ae402b39167`；由 Go 生成器提取 17 套彩色色板和 9 套中性色板，共 26 套 OKLCH 色阶。Black 是本库随明暗模式切换的覆盖配置，不是额外提取的色板。

MIT License，Copyright (c) Tailwind Labs, Inc.，完整许可见 [tailwindcss.txt](third_party/licenses/tailwindcss.txt)。运行时仅包含生成的 Go 色值，没有 CSS 引擎。

## MyGo

原生布局、输入、文本、绘制、焦点、选择与弹层能力通过 Go 模块依赖 [egoist/mygo](https://github.com/egoist/mygo/tree/402a3cf8eb35590078695ac402df40bf5c255506) 使用。具体版本及传递依赖见 [go.mod](go.mod) 与 [go.sum](go.sum)，各依赖保留自己的许可。

本项目自身代码采用 [MIT License](LICENSE)，版权声明为 NuxtUI Go contributors。上述第三方材料保留各自的版权与许可声明；本项目的许可不替代这些声明。
