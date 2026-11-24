import type { ContentNavigationItem } from '@nuxt/content'
import type { Demo } from '~/types/catalog'

// Navigation built from catalog.json: the sidebar and mobile menu, the search
// groups, and previous/next links all follow the same category order.
export function useNavigation() {
  const { t } = useI18n()
  const localePath = useLocalePath()
  const { components, categories, local } = useCatalog()
  const pathOf = (demo: Demo) => localePath(`/components/${demo.id}`)
  // Every component in sidebar order: by category, then catalog order.
  const flat = categories.flatMap(category => components.filter(c => c.category === category))
  // exact: only the page's own link is active and gets aria-current="page".
  const item = (demo: Demo): ContentNavigationItem => ({
    title: demo.name,
    path: pathOf(demo),
    description: local(demo.description),
    exact: true
  })
  const groups = computed(() =>
    categories.map(category => ({
      category,
      title: t(category),
      components: flat.filter(c => c.category === category)
    }))
  )
  // Sidebar: the overview link, then one group per category.
  const navigation = computed<ContentNavigationItem[]>(() => [
    { title: t('overview'), path: localePath('/'), icon: 'i-lucide-layout-grid', exact: true },
    ...groups.value.map(group => ({
      title: group.title,
      path: `#${group.category}`,
      children: group.components.map(item)
    }))
  ])
  // Search: one group per category, matched on the name and both descriptions.
  const searchGroups = computed(() =>
    groups.value.map(group => ({
      id: group.category,
      label: group.title,
      items: group.components.map(demo => ({
        label: demo.name,
        suffix: local(demo.description),
        to: pathOf(demo),
        keywords: `${demo.name} ${demo.description.zh} ${demo.description.en}`
      }))
    }))
  )
  // The previous and next components in sidebar order, crossing categories.
  // A missing neighbor is null, which UContentSurround leaves empty, as with
  // Nuxt Content's own surroundings query.
  function surround(id: string) {
    const index = flat.findIndex(c => c.id === id)
    const prev = index > 0 ? flat[index - 1] : undefined
    const next = index >= 0 ? flat[index + 1] : undefined
    return [prev ? item(prev) : null, next ? item(next) : null] as ContentNavigationItem[]
  }
  return { flat, groups, navigation, searchGroups, surround }
}
