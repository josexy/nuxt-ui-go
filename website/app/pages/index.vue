<script setup lang="ts">
const { t } = useI18n()
const localePath = useLocalePath()
const { components, categories, local } = useCatalog()
useSeoMeta({ title: () => `${t('overview')} · NuxtUI Go`, description: () => t('heroDescription') })
</script>

<template>
  <div class="overview-page">
    <div class="hero">
      <div class="hero-mark"><UIcon name="i-lucide-braces" /><span>NuxtUI Go</span></div>
      <h1>{{ t('heroTitle') }}</h1>
      <p>{{ t('heroDescription') }}</p>
      <div class="hero-actions">
        <UButton
          :to="localePath('/components/button')"
          size="lg"
          icon="i-lucide-mouse-pointer-2"
          >{{ t('start') }}</UButton
        ><span>{{ t('count') }}</span>
      </div>
    </div>
    <section class="install-panel">
      <h2>{{ t('quickStart') }}</h2>
      <pre><code>go get github.com/josexy/nuxt-ui-go/ui</code></pre>
      <p>{{ t('installNote') }}</p>
    </section>
    <section
      v-for="category in categories"
      :key="category"
      class="catalog-section"
    >
      <h2>
        {{ t(category) }} <span>{{ components.filter(c => c.category === category).length }}</span>
      </h2>
      <div class="catalog-grid">
        <NuxtLink
          v-for="component in components.filter(c => c.category === category)"
          :key="component.id"
          :to="localePath(`/components/${component.id}`)"
          class="catalog-entry"
          ><strong>{{ component.name }}</strong>
          <p>{{ local(component.description) }}</p></NuxtLink
        >
      </div>
    </section>
  </div>
</template>
