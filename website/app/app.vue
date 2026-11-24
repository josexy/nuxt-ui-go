<script setup lang="ts">
import { en, zh_cn } from '@nuxt/ui/locale'

const { locale } = useI18n()
const colorMode = useColorMode()
useHead({
  htmlAttrs: { lang: computed(() => (locale.value === 'zh' ? 'zh-Hans' : 'en')) },
  // The browser chrome matches the page surface: white, or slate-900 in dark.
  meta: [
    {
      name: 'theme-color',
      content: computed(() => (colorMode.value === 'dark' ? '#0f172b' : '#ffffff'))
    }
  ]
})
</script>

<template>
  <UApp :locale="locale === 'zh' ? zh_cn : en">
    <NuxtLayout>
      <NuxtPage :page-key="route => String(route.params.id || 'home')" />
    </NuxtLayout>
  </UApp>
</template>
