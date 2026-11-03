<script setup lang="ts">
const { locale } = useI18n()
useHead({ htmlAttrs: { lang: computed(() => (locale.value === 'zh' ? 'zh-Hans' : 'en')) } })
</script>

<template>
  <UApp>
    <NuxtLayout>
      <NuxtPage :page-key="route => String(route.params.id || 'home')" />
    </NuxtLayout>
  </UApp>
</template>
