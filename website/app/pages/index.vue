<script setup lang="ts">
const { t } = useI18n()
const localePath = useLocalePath()
const { components, local } = useCatalog()
const { groups } = useNavigation()
const count = components.length
const install = 'go get github.com/josexy/nuxt-ui-go/ui'
// Real catalog output: the Button example's default snippet.
const snippet = components.find(c => c.id === 'button')?.snippet ?? ''
const features = computed(() => [
  { icon: 'i-lucide-app-window', title: t('featureNative'), description: t('featureNativeText') },
  { icon: 'i-lucide-play', title: t('featurePreview'), description: t('featurePreviewText') },
  {
    icon: 'i-lucide-sliders-horizontal',
    title: t('featureFields'),
    description: t('featureFieldsText')
  },
  { icon: 'i-lucide-code', title: t('featureCode'), description: t('featureCodeText') }
])
useSeoMeta({
  title: () => `${t('overview')} · NuxtUI Go`,
  description: () => t('heroDescription', { count })
})
</script>

<template>
  <UContainer>
    <UPage>
      <template #left>
        <DocsAside />
      </template>

      <div
        id="main-content"
        tabindex="-1"
        class="pb-24 focus:outline-none"
      >
        <UPageHero
          orientation="horizontal"
          :title="t('heroTitle')"
          :description="t('heroDescription', { count })"
          :ui="{
            container: 'px-0 sm:px-0 lg:px-0 py-12 sm:py-16 lg:py-20 gap-10 sm:gap-y-12',
            title: 'text-4xl sm:text-5xl',
            description: 'text-base sm:text-lg',
            body: 'mt-8',
            footer: 'mt-8'
          }"
        >
          <template #links>
            <UButton
              :to="localePath('/components/button')"
              trailing-icon="i-lucide-arrow-right"
              size="lg"
              >{{ t('start') }}</UButton
            >
            <UButton
              to="https://github.com/josexy/nuxt-ui-go"
              target="_blank"
              icon="i-lucide-github"
              color="neutral"
              variant="outline"
              size="lg"
              >GitHub<span class="sr-only"> {{ t('newTab') }}</span></UButton
            >
          </template>
          <template #headline>
            <UBadge
              color="neutral"
              variant="subtle"
              icon="i-lucide-braces"
              :label="t('count', { count })"
            />
          </template>
          <div class="min-w-0 space-y-4">
            <div class="space-y-2">
              <CodePanel
                :code="install"
                :title="t('quickStart')"
                language="shell"
                copyable
              />
              <p class="text-sm text-muted">{{ t('installNote') }}</p>
            </div>
            <CodePanel
              :code="snippet"
              :title="`Button · ${t('fragment')}`"
              badge="Go"
              copyable
            />
          </div>
        </UPageHero>

        <UPageSection
          :title="t('featuresTitle')"
          :description="t('featuresDescription')"
          :features="features"
          :ui="{
            container: 'px-0 sm:px-0 lg:px-0 py-12 sm:py-14 lg:py-16 gap-8 sm:gap-10',
            title: 'text-2xl sm:text-3xl lg:text-3xl',
            features: 'sm:grid-cols-2 lg:grid-cols-2 gap-6'
          }"
        />

        <section
          v-for="group in groups"
          :key="group.category"
          class="mt-12"
          :aria-labelledby="group.category"
        >
          <h2
            :id="group.category"
            class="mb-5 flex items-center gap-2.5 text-xl font-bold text-highlighted"
          >
            {{ group.title }}
            <UBadge
              :label="String(group.components.length)"
              color="neutral"
              variant="subtle"
            />
          </h2>
          <UPageGrid class="gap-4 lg:grid-cols-2 xl:grid-cols-3">
            <UPageCard
              v-for="component in group.components"
              :key="component.id"
              :title="component.name"
              :description="local(component.description)"
              :to="localePath(`/components/${component.id}`)"
              variant="outline"
              :ui="{ container: 'p-4 sm:p-5', description: 'text-sm' }"
            />
          </UPageGrid>
        </section>
      </div>
    </UPage>
  </UContainer>
</template>
