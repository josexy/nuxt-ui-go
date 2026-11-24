<script setup lang="ts">
// A code block with a copy button. With `tabs` it switches between several
// sources (the component snippet and the complete program); with `code` it
// shows a single block, as on the overview page. Copying places the source
// string itself on the clipboard, never the rendered markup.
const props = defineProps<{
  tabs?: { value: string; label: string; code: string }[]
  code?: string
  // Names the single block and its scroll region.
  title?: string
  badge?: string
  language?: 'go' | 'shell'
  // Newer settings await confirmation: show a spinner.
  pending?: boolean
  // Copying is allowed only for confirmed code.
  copyable?: boolean
}>()
const tab = defineModel<string>('tab')
const { t } = useI18n()
const current = computed(
  () => props.tabs?.find(item => item.value === tab.value) ?? props.tabs?.[0]
)
const currentCode = computed(() => (props.tabs ? current.value?.code : props.code) ?? '')
const label = computed(() => (props.tabs ? current.value?.label : props.title) ?? '')

// UTabs renders only the tab list here, so the code block below serves as the
// selected tab's panel: it is labelled by that tab, and the tab points at it.
// (Reka sets aria-controls only for panels it renders itself.)
const tabsRef = useTemplateRef<{ $el: HTMLElement }>('tabsRef')
const panelId = useId()
const panel = ref<{ id?: string; 'aria-labelledby'?: string }>({})
function linkPanel() {
  const tabs = tabsRef.value?.$el.querySelectorAll('[role="tab"]') ?? []
  const selected = [...tabs].find(tab => tab.getAttribute('aria-selected') === 'true')
  for (const tab of tabs)
    if (tab === selected) tab.setAttribute('aria-controls', panelId)
    else tab.removeAttribute('aria-controls')
  panel.value = selected ? { id: panelId, 'aria-labelledby': selected.id } : {}
}

const copyState = ref('copy')
let copyTimer: ReturnType<typeof setTimeout> | undefined
async function copy() {
  if (!props.copyable) return
  try {
    await navigator.clipboard.writeText(currentCode.value)
    copyState.value = 'copied'
  } catch {
    copyState.value = 'copyFailed'
  }
  clearTimeout(copyTimer)
  copyTimer = setTimeout(() => {
    copyState.value = 'copy'
  }, 2500)
}
onMounted(() => {
  linkPanel()
  watch(tab, () => nextTick(linkPanel))
})
onBeforeUnmount(() => clearTimeout(copyTimer))
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-default">
    <div
      class="flex min-h-11 flex-wrap items-center gap-x-3 border-b border-default bg-elevated/50 ps-3 pe-2"
    >
      <UTabs
        v-if="tabs"
        ref="tabsRef"
        v-model="tab"
        :items="tabs"
        :content="false"
        variant="link"
        size="sm"
        class="min-w-0 self-stretch"
        :ui="{
          root: 'gap-0',
          list: 'h-full w-auto border-b-0 mb-0 p-0 gap-3',
          trigger: 'px-0.5'
        }"
      />
      <span
        v-else
        class="text-sm font-medium text-highlighted"
        >{{ title }}</span
      >
      <div class="ms-auto flex items-center gap-2 py-1.5">
        <UIcon
          v-if="pending"
          name="i-lucide-loader-circle"
          class="size-4 animate-spin text-muted"
          role="img"
          :aria-label="t('codePending')"
        />
        <UBadge
          v-if="badge"
          :label="badge"
          color="neutral"
          variant="outline"
          size="sm"
        />
        <UButton
          :disabled="!copyable"
          :icon="
            copyState === 'copied'
              ? 'i-lucide-check'
              : copyState === 'copyFailed'
                ? 'i-lucide-x'
                : 'i-lucide-copy'
          "
          color="neutral"
          variant="ghost"
          :ui="{ leadingIcon: copyState === 'copied' ? 'text-primary' : '' }"
          size="sm"
          @click="copy"
          >{{ t(copyState === 'copied' ? 'copied' : 'copy') }}</UButton
        >
      </div>
    </div>
    <CodeBlock
      v-bind="tabs ? { role: 'tabpanel', ...panel } : {}"
      :code="currentCode"
      :label="label"
      :language="language"
      class="bg-default [&_pre]:max-h-[36rem]"
    />
    <p
      v-if="copyState === 'copyFailed'"
      role="alert"
      class="flex items-center gap-2 border-t border-default px-4 py-2 text-sm text-default"
    >
      <UIcon
        name="i-lucide-circle-alert"
        class="size-4 shrink-0 text-error"
      />{{ t('copyFailed') }}
    </p>
    <span
      class="sr-only"
      aria-live="polite"
      >{{ copyState === 'copied' ? t('copied') : '' }}</span
    >
  </div>
</template>
