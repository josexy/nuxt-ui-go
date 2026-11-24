<script setup lang="ts">
// The wide-viewport sidebar: the overview link and every component by category.
const { t } = useI18n()
const { navigation } = useNavigation()
const aside = useTemplateRef<{ $el: HTMLElement }>('aside')
// Keep the current component's link visible when the page opens, even for a
// component near the end of the catalog. Only the sidebar scrolls.
onMounted(() => {
  const root = aside.value?.$el
  const link = root?.querySelector<HTMLElement>('[aria-current="page"]')
  if (!root || !link) return
  const top = link.getBoundingClientRect().top - root.getBoundingClientRect().top
  if (top < 0 || top > root.clientHeight - link.offsetHeight)
    root.scrollTop += top - root.clientHeight / 2
})
</script>

<template>
  <UPageAside
    ref="aside"
    :aria-label="t('components')"
  >
    <UContentNavigation
      :navigation="navigation"
      highlight
    />
  </UPageAside>
</template>
