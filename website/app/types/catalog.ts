export interface LocalText {
  zh: string
  en: string
}
export interface Control {
  key: string
  label: LocalText
  kind: string
  default: string | number | boolean | null
  choices?: { value: string; go: string }[]
  min?: number
  max?: number
  goType?: string
}
export interface API {
  name: string
  signature: string
  fields:
    { name: string; type: string; comment: string; description: LocalText; zero: string }[] | null
  file: string
  line: number
}
export interface Demo {
  id: string
  name: string
  category: string
  description: LocalText
  guide: LocalText
  note: LocalText
  controls: Control[]
  api: API[]
  snippet: string
  source: string
  height: number
  bindings: Binding[] | null
}
export type JSONValue =
  null | boolean | number | string | JSONValue[] | { [key: string]: JSONValue }
export interface ValueField {
  name: string
  type: string
  description: LocalText
  control?: string
  choices?: string[]
  hint?: LocalText
  when?: { field: string; value: string; label: LocalText }
  slots?: string[]
}
export interface ValueSchema {
  type: string
  kind: string
  item?: string
  fields?: ValueField[]
  choices?: string[]
  returns?: string[]
  arguments?: string[]
  length?: number
  sample: JSONValue
}
export interface Binding {
  id: string
  name: string
  type: string
  fields?: ValueField[]
  defaults: JSONValue
}
export interface PreviewConfig {
  values: Record<string, string | number | boolean | null>
  options: Record<string, JSONValue>
  primary: string
  neutral: string
  mode: 'light' | 'dark' | 'system'
  locale: 'en' | 'zh-Hans'
}
