import data from '~/data/catalog.json'
import type { Demo, LocalText, ValueSchema, ValueField } from '~/types/catalog'

export function useCatalog() {
  const { locale } = useI18n()
  const components = data.components as unknown as Demo[]
  const local = (text: LocalText) => (locale.value === 'zh' ? text.zh : text.en)
  const categories = ['foundations', 'inputs', 'overlays', 'data', 'advanced']
  const schemas = data.schemas as unknown as Record<string, ValueSchema>
  function matchesField(field: ValueField, query: string, visited = new Set<string>()): boolean {
    const needle = query.toLowerCase().trim()
    if (
      !needle ||
      `${field.name} ${field.type} ${field.description.zh} ${field.description.en}`
        .toLowerCase()
        .includes(needle)
    )
      return true
    if (visited.has(field.type)) return false
    const seen = new Set(visited).add(field.type)
    const schema = schemas[field.type]
    if (schema?.fields?.some(f => matchesField(f, query, seen))) return true
    const item =
      schema?.item ||
      (schema?.kind === 'function' && schema.choices?.includes('style') ? 'nui.Styles' : undefined)
    return Boolean(
      item && matchesField({ name: '', type: item, description: { zh: '', en: '' } }, query, seen)
    )
  }
  return {
    components,
    local,
    categories,
    primary: data.primary,
    neutral: data.neutral,
    schemas,
    matchesField
  }
}
