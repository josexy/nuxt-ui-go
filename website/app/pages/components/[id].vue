<script setup lang="ts">
const route = useRoute()
const { locale, t } = useI18n()
const localePath = useLocalePath()
const { components, local } = useCatalog()
const demo = computed(() => components.find(c => c.id === route.params.id))
if (!demo.value) throw createError({ statusCode: 404, statusMessage: 'Component not found' })
const { data: page } = await useAsyncData(
  () => `content-${locale.value}-${route.params.id}`,
  () =>
    queryCollection(locale.value === 'zh' ? 'zh' : 'en')
      .where('component', '=', String(route.params.id))
      .first()
)
useSeoMeta({
  title: () => `${demo.value?.name} · NuxtUI Go`,
  description: () => (demo.value ? local(demo.value.description) : '')
})
</script>

<template>
  <div
    v-if="demo"
    class="component-page"
  >
    <article class="component-article">
      <nav class="breadcrumbs">
        <NuxtLink :to="localePath('/')">{{ t('components') }}</NuxtLink
        ><UIcon name="i-lucide-chevron-right" /><span>{{ t(demo.category) }}</span>
      </nav>
      <div class="page-heading">
        <h1>{{ demo.name }}</h1>
        <p>{{ local(demo.description) }}</p>
      </div>
      <ComponentPlayground
        :key="demo.id"
        :demo="demo"
      />
      <section
        id="usage"
        class="usage-section"
      >
        <h2>{{ t('usage') }}</h2>
        <ContentRenderer
          v-if="page"
          :value="page"
        />
        <p v-else>{{ local(demo.guide) }}</p>
      </section>
      <section
        id="api"
        class="api-section"
      >
        <h2>{{ t('api') }}</h2>
        <div
          v-for="entry in demo.api"
          :key="entry.name"
          class="api-entry"
        >
          <div class="api-title">
            <h3>{{ entry.name }}</h3>
            <a
              :href="`https://github.com/josexy/nuxt-ui-go/blob/main/${entry.file}#L${entry.line}`"
              target="_blank"
              rel="noopener noreferrer"
              >{{ t('source') }} <UIcon name="i-lucide-arrow-up-right"
            /></a>
          </div>
          <pre
            tabindex="0"
            class="signature"
          ><code>{{ entry.signature }}</code></pre>
          <div
            v-if="entry.fields?.length"
            class="table-scroll"
          >
            <table>
              <thead>
                <tr>
                  <th>{{ t('field') }}</th>
                  <th>{{ t('type') }}</th>
                  <th>{{ t('description') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="field in entry.fields"
                  :key="field.name"
                >
                  <td>
                    <code>{{ field.name }}</code>
                  </td>
                  <td>
                    <code>{{ field.type }}</code>
                    <p class="field-zero">
                      {{ t('zeroValue') }}: <code>{{ field.zero }}</code>
                    </p>
                  </td>
                  <td>
                    <p>{{ local(field.description) }}</p>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </article>
    <aside class="page-toc">
      <p>{{ t('onThisPage') }}</p>
      <a
        v-for="anchor in ['preview', 'parameters', 'code', 'usage', 'api']"
        :key="anchor"
        :href="`#${anchor}`"
        >{{ t(anchor) }}</a
      >
    </aside>
  </div>
</template>
