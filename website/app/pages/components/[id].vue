<script setup lang="ts">
const route = useRoute()
const { locale, t } = useI18n()
const localePath = useLocalePath()
const { components, local } = useCatalog()
const { surround } = useNavigation()
const demo = computed(() => components.find(c => c.id === route.params.id))
if (!demo.value) throw createError({ statusCode: 404, statusMessage: 'Component not found' })
const usageMarkdown = computed(() => {
  if (!demo.value) return ''
  const note = local(demo.value.note)
  return [local(demo.value.guide), note ? `### ${t('usageNotes')}\n\n${note}` : '']
    .filter(Boolean)
    .join('\n\n')
})
const { data: page } = await useAsyncData(
  () => `content-${locale.value}-${route.params.id}`,
  () =>
    queryCollection(locale.value === 'zh' ? 'zh' : 'en')
      .where('component', '=', String(route.params.id))
      .first()
)
// The section headings carry these ids, which existing links already use.
const toc = computed(() => [
  {
    id: 'preview',
    text: t('preview'),
    depth: 2,
    children: [{ id: 'parameters', text: t('parameters'), depth: 3 }]
  },
  { id: 'code', text: t('code'), depth: 2 },
  { id: 'usage', text: t('usage'), depth: 2 },
  { id: 'api', text: t('api'), depth: 2 }
])
const sourceURL = (file: string, line: number) =>
  `https://github.com/josexy/nuxt-ui-go/blob/main/${file}#L${line}`
useSeoMeta({
  title: () => `${demo.value?.name} · NuxtUI Go`,
  description: () => (demo.value ? local(demo.value.description) : '')
})
</script>

<template>
  <UContainer v-if="demo">
    <UPage>
      <template #left>
        <DocsAside />
      </template>

      <div
        id="main-content"
        tabindex="-1"
        class="focus:outline-none"
      >
        <UBreadcrumb
          class="pt-8"
          :items="[
            { label: t('components'), to: localePath('/') },
            { label: t(demo.category), active: false }
          ]"
        />
        <UPageHeader
          :title="demo.name"
          :description="local(demo.description)"
          :headline="t(demo.category)"
          :ui="{ root: 'pt-4' }"
        />
        <UPageBody>
          <ComponentPlayground
            :key="demo.id"
            :demo="demo"
          />

          <section>
            <h2
              id="usage"
              class="mb-2 text-2xl font-bold text-highlighted"
            >
              {{ t('usage') }}
            </h2>
            <div
              class="max-w-[75ch] [&_:not(pre)>code]:max-w-full [&_:not(pre)>code]:wrap-anywhere [&>*:first-child]:mt-0"
            >
              <ContentRenderer
                v-if="page"
                :value="page"
              />
              <MDC
                v-else
                :value="usageMarkdown"
              />
            </div>
          </section>

          <section class="space-y-8">
            <div>
              <h2
                id="api"
                class="mb-3 text-2xl font-bold text-highlighted"
              >
                {{ t('api') }}
              </h2>
              <p class="max-w-[75ch] text-muted">{{ t('apiHint') }}</p>
            </div>
            <div
              v-for="entry in demo.api"
              :key="entry.name"
              class="space-y-3"
            >
              <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
                <h3 class="font-mono text-lg font-semibold text-highlighted">{{ entry.name }}</h3>
                <ULink
                  :to="sourceURL(entry.file, entry.line)"
                  target="_blank"
                  raw
                  class="inline-flex items-center gap-1 rounded-sm text-sm font-medium text-primary hover:underline"
                  >{{ t('source') }}<span class="sr-only"> {{ t('newTab') }}</span
                  ><UIcon
                    name="i-lucide-arrow-up-right"
                    class="size-4"
                /></ULink>
              </div>
              <CodeBlock
                :code="entry.signature"
                :label="t('signature', { name: entry.name })"
                class="overflow-hidden rounded-lg border border-default bg-muted [&_pre]:max-h-96"
              />
              <div
                v-if="entry.fields?.length"
                class="overflow-x-auto rounded-lg border border-default"
                tabindex="0"
                role="region"
                :aria-label="t('fields', { name: entry.name })"
              >
                <table class="w-full min-w-[36rem] border-collapse text-left text-sm">
                  <thead class="bg-elevated/50 text-highlighted">
                    <tr>
                      <th
                        v-for="heading in ['field', 'type', 'description']"
                        :key="heading"
                        scope="col"
                        class="border-b border-default px-4 py-2.5 font-semibold"
                      >
                        {{ t(heading) }}
                      </th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-default">
                    <tr
                      v-for="field in entry.fields"
                      :key="field.name"
                      class="align-top"
                    >
                      <td class="px-4 py-3">
                        <code class="font-mono font-medium text-highlighted">{{ field.name }}</code>
                      </td>
                      <td class="px-4 py-3">
                        <code class="font-mono text-default wrap-anywhere">{{ field.type }}</code>
                        <p class="mt-1.5 text-xs text-muted">
                          {{ t('zeroValue') }}:
                          <code class="font-mono">{{
                            field.zero === 'unset'
                              ? t('unsetValue')
                              : field.zero === 'zero value'
                                ? t('typeZeroValue')
                                : field.zero
                          }}</code>
                        </p>
                      </td>
                      <td class="px-4 py-3 text-default">{{ local(field.description) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          <USeparator />
          <UContentSurround :surround="surround(demo.id)" />
        </UPageBody>
      </div>

      <template #right>
        <UContentToc
          :title="t('onThisPage')"
          :links="toc"
          highlight
        />
      </template>
    </UPage>
  </UContainer>
</template>
