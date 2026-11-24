<script setup lang="ts">
// A focusable, self-scrolling code block. Go is highlighted once Shiki loads
// in the browser; before that, or if it fails, the same code shows as plain
// text, so the block always shows exactly `code`.
const props = defineProps<{ code: string; label: string; language?: 'go' | 'shell' }>()
const { codeToHtml } = useGoHighlighter()
const highlighted = shallowRef<{ code: string; html: string }>()
const html = computed(() => (highlighted.value?.code === props.code ? highlighted.value.html : ''))
async function highlight() {
  if (props.language === 'shell') return
  const code = props.code
  try {
    const result = await codeToHtml(code, props.label)
    if (code === props.code) highlighted.value = { code, html: result }
  } catch {
    // The plain block stays; it is readable and copyable.
  }
}
onMounted(() => watch(() => [props.code, props.label], highlight, { immediate: true }))
</script>

<template>
  <div
    class="text-sm/6 [&_pre]:overflow-auto [&_pre]:px-4 [&_pre]:py-3 [&_pre]:font-mono [&_pre]:[tab-size:4]"
  >
    <div
      v-if="html"
      class="contents"
      v-html="html"
    />
    <pre
      v-else
      tabindex="0"
      :aria-label="label"
      class="text-highlighted"
    ><code>{{ code }}</code></pre>
  </div>
</template>
