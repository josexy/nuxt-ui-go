<script setup lang="ts">
import type { Demo, PreviewConfig } from '~/types/catalog'
const props = defineProps<{ demo: Demo; config: PreviewConfig; resetKey: number }>()
const emit = defineEmits<{
  applied: [snippet: string, source: string]
  status: [value: string]
  event: [name: string, args: string[]]
}>()
const { t } = useI18n()
const runtime = useRuntimeConfig()
const frame = ref<HTMLIFrameElement>()
const state = ref<'loading' | 'ready' | 'error'>('loading')
const attempt = ref(0)
const configurationError = ref('')
let ready = false
let sequence = 0
let requested = 0
let timer: ReturnType<typeof setTimeout> | undefined
let updateTimer: ReturnType<typeof setTimeout> | undefined
let previousReset = props.resetKey
let pendingReset = false
const initialMode = props.config.mode
const frameURL = computed(
  () =>
    `${runtime.app.baseURL}preview/index.html?component=${encodeURIComponent(props.demo.id)}&mode=${initialMode}&locale=${props.config.locale}&attempt=${attempt.value}`
)
// Keep the frame URL stable on locale/theme changes to preserve Go widget state.
const stableURL = ref(frameURL.value)
function status(value: 'loading' | 'ready' | 'error') {
  state.value = value
  emit('status', value)
}
function send() {
  if (!ready || !frame.value?.contentWindow) return
  requested = ++sequence
  frame.value.contentWindow.postMessage(
    {
      channel: 'nuxt-ui-go',
      type: 'configure',
      component: props.demo.id,
      id: requested,
      reset: pendingReset,
      config: JSON.parse(JSON.stringify(props.config))
    },
    window.location.origin
  )
  pendingReset = false
}
function onMessage(event: MessageEvent) {
  if (
    event.origin !== window.location.origin ||
    event.source !== frame.value?.contentWindow ||
    event.data?.channel !== 'nuxt-ui-go'
  )
    return
  const message = event.data
  if (message.type === 'ready' && message.component === props.demo.id) {
    ready = true
    send()
  }
  if (message.type === 'applied' && message.id === requested) {
    clearTimeout(timer)
    configurationError.value = ''
    status('ready')
    emit('applied', message.snippet, message.source)
  }
  if (message.type === 'fatal') {
    ready = false
    clearTimeout(timer)
    status('error')
  } else if (message.type === 'error' && message.id === requested) {
    clearTimeout(timer)
    configurationError.value = String(message.message)
    status('ready')
    emit('status', 'invalid')
  }
  if (message.type === 'event') emit('event', String(message.name), message.args || [])
}
function startTimer() {
  clearTimeout(timer)
  timer = setTimeout(() => {
    if (state.value !== 'ready') status('error')
  }, 45000)
}
function retry() {
  ready = false
  requested = ++sequence
  configurationError.value = ''
  status('loading')
  attempt.value++
  stableURL.value = frameURL.value
  startTimer()
}
watch(
  () => props.config,
  () => {
    // Invalidate in-flight replies immediately, including during the debounce.
    // They describe the previous settings and must not re-enable copying.
    requested = ++sequence
    clearTimeout(updateTimer)
    updateTimer = setTimeout(send, 50)
  },
  { deep: true, flush: 'sync' }
)
watch(
  () => props.resetKey,
  value => {
    if (value !== previousReset) {
      pendingReset = true
      previousReset = value
      send()
    }
  }
)
onMounted(() => {
  window.addEventListener('message', onMessage)
  startTimer()
})
onBeforeUnmount(() => {
  window.removeEventListener('message', onMessage)
  clearTimeout(timer)
  clearTimeout(updateTimer)
})
</script>

<template>
  <div
    class="live-frame"
    :data-preview-state="state"
  >
    <div
      v-if="state === 'loading'"
      class="preview-loading"
      role="status"
    >
      {{ t('loading') }}
    </div>
    <iframe
      ref="frame"
      :key="attempt"
      :src="stableURL"
      :title="`${demo.name} — ${t('preview')}`"
      :style="{ height: `${demo.height}px` }"
      @error="status('error')"
    />
    <div
      v-if="state === 'error'"
      class="preview-error"
    >
      <p>{{ t('previewError') }}</p>
      <UButton
        icon="i-lucide-refresh-cw"
        variant="outline"
        @click="retry"
        >{{ t('retry') }}</UButton
      >
    </div>
    <p
      v-if="configurationError"
      class="configuration-error"
      role="alert"
    >
      {{ configurationError }}
    </p>
  </div>
</template>
