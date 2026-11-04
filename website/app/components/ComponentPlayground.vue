<script setup lang="ts">
import type { Demo, PreviewConfig, Binding, JSONValue, Control } from '~/types/catalog'
const props = defineProps<{ demo: Demo }>()
const { t, locale } = useI18n()
const { primary, neutral, local, matchesField } = useCatalog()
const colorMode = useColorMode()
const values = useState<PreviewConfig['values']>(`parameters-${props.demo.id}`, () =>
  Object.fromEntries(props.demo.controls.map(c => [c.key, c.default]))
)
const options = useState<PreviewConfig['options']>(`fields-${props.demo.id}`, () => ({}))
const fieldFilter = ref('')
const events = ref<{ name: string; args: string[] }[]>([])
const visibleBindings = computed(() =>
  (props.demo.bindings || []).filter(
    b =>
      !fieldFilter.value ||
      !b.fields?.length ||
      b.fields.some(f => matchesField(f, fieldFilter.value))
  )
)
const fieldCount = computed(() =>
  (props.demo.bindings || []).reduce((n, b) => n + (b.fields?.length || 1), 0)
)
function configure(id: string, value: JSONValue | undefined) {
  if (value === undefined) delete options.value[id]
  else options.value[id] = value
}
function quickValue(control: Control) {
  for (const binding of props.demo.bindings || []) {
    const field = binding.fields?.find(f => f.control === control.key)
    const patch = options.value[binding.id]
    if (
      !field ||
      !patch ||
      typeof patch !== 'object' ||
      Array.isArray(patch) ||
      !Object.hasOwn(patch, field.name)
    )
      continue
    const raw = patch[field.name] ?? null
    if (control.kind === 'select') {
      const choice = control.choices?.find(
        c =>
          c.go === raw ||
          (raw === null && c.go === 'nil') ||
          (typeof raw === 'string' &&
            c.go ===
              `lucide.${raw
                .split('-')
                .map(p => p.charAt(0).toUpperCase() + p.slice(1))
                .join('')}()`)
      )
      return choice?.value || '__custom__'
    }
    return raw as string | number | boolean | null
  }
  return values.value[control.key]
}
function setQuick(control: Control, raw: string | boolean) {
  for (const binding of props.demo.bindings || []) {
    const patch = options.value[binding.id]
    if (!patch || typeof patch !== 'object' || Array.isArray(patch)) continue
    for (const field of binding.fields || [])
      if (field.control === control.key) delete patch[field.name]
  }
  values.value[control.key] =
    control.kind === 'optionalBoolean'
      ? raw === 'true'
        ? true
        : raw === 'false'
          ? false
          : null
      : control.kind === 'number' || control.kind === 'integer'
        ? Number(raw)
        : raw
}
function inputValue(event: Event) {
  return (event.target as HTMLInputElement).value
}
function defaults(binding: Binding): JSONValue {
  if (
    !binding.fields?.length ||
    !binding.defaults ||
    typeof binding.defaults !== 'object' ||
    Array.isArray(binding.defaults)
  )
    return binding.defaults
  const result = { ...binding.defaults }
  for (const field of binding.fields) {
    const control = props.demo.controls.find(c => c.key === field.control)
    if (!control) continue
    let value: JSONValue = values.value[control.key] ?? null
    if (control.kind === 'select') {
      const go = control.choices?.find(c => c.value === value)?.go || ''
      if (go.startsWith('nui.')) value = go
      else if (go.startsWith('lucide.'))
        value = go.slice(7, -2).replace(/[A-Z]/g, (s, i) => `${i ? '-' : ''}${s.toLowerCase()}`)
      else if (go === 'nil') value = null
    }
    result[field.name] = value
  }
  return result
}
function record(name: string, args: string[]) {
  events.value = [{ name, args }, ...events.value].slice(0, 5)
}
const primaryColor = useState('preview-primary', () => 'green')
const neutralColor = useState('preview-neutral', () => 'slate')
const config = computed<PreviewConfig>(() => ({
  values: values.value,
  options: options.value,
  primary: primaryColor.value,
  neutral: neutralColor.value,
  mode:
    colorMode.preference === 'system' ? 'system' : colorMode.value === 'dark' ? 'dark' : 'light',
  locale: locale.value === 'zh' ? 'zh-Hans' : 'en'
}))
const resetKey = ref(0)
const status = ref('loading')
const codeMode = ref('fragment')
const snippet = ref(props.demo.snippet)
const source = ref(props.demo.source)
const copyState = ref('copy')
const synchronized = ref(false)
const currentCode = computed(() => (codeMode.value === 'complete' ? source.value : snippet.value))
watch(
  config,
  () => {
    synchronized.value = false
  },
  { deep: true, flush: 'sync' }
)
function previewStatus(value: string) {
  status.value = value
  if (value !== 'ready') synchronized.value = false
}
function applied(nextSnippet: string, nextSource: string) {
  snippet.value = nextSnippet
  source.value = nextSource
  synchronized.value = true
}
function reset() {
  values.value = Object.fromEntries(props.demo.controls.map(c => [c.key, c.default]))
  options.value = {}
  events.value = []
  resetKey.value++
}
let copyTimer: ReturnType<typeof setTimeout> | undefined
async function copy() {
  if (!synchronized.value) return
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
onBeforeUnmount(() => clearTimeout(copyTimer))
</script>

<template>
  <section
    id="preview"
    class="playground-section"
  >
    <div class="section-title">
      <h2>{{ t('preview') }}</h2>
      <UButton
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="ghost"
        size="sm"
        @click="reset"
        >{{ t('reset') }}</UButton
      >
    </div>
    <div class="playground">
      <div class="theme-toolbar">
        <label
          >{{ t('primary')
          }}<select
            v-model="primaryColor"
            :aria-label="t('primary')"
          >
            <option
              v-for="name in primary"
              :key="name"
              :value="name"
            >
              {{ name }}
            </option>
          </select></label
        >
        <label
          >{{ t('neutral')
          }}<select
            v-model="neutralColor"
            :aria-label="t('neutral')"
          >
            <option
              v-for="name in neutral"
              :key="name"
              :value="name"
            >
              {{ name }}
            </option>
          </select></label
        >
      </div>
      <ClientOnly>
        <PreviewFrame
          :demo="demo"
          :config="config"
          :reset-key="resetKey"
          @applied="applied"
          @status="previewStatus"
          @event="record"
        />
        <template #fallback
          ><div
            class="preview-placeholder"
            :style="{ height: `${demo.height + 40}px` }"
          >
            {{ t('loading') }}
          </div></template
        >
      </ClientOnly>
      <div
        id="parameters"
        class="parameter-panel"
      >
        <h3>{{ t('parameters') }}</h3>
        <p class="parameter-hint">{{ t('parameterHint') }}</p>
        <div class="parameter-grid">
          <label
            v-for="control in demo.controls"
            :key="control.key"
            class="parameter-field"
            :for="`param-${control.key}`"
          >
            <span
              >{{ local(control.label) }} <code>{{ control.key }}</code></span
            >
            <select
              v-if="control.kind === 'select'"
              :id="`param-${control.key}`"
              :value="quickValue(control)"
              :data-param="control.key"
              @change="setQuick(control, inputValue($event))"
            >
              <option
                v-if="control.kind === 'select' && quickValue(control) === '__custom__'"
                value="__custom__"
                disabled
              >
                {{ t('customValue') }}
              </option>
              <option
                v-for="choice in control.choices"
                :key="choice.value"
                :value="choice.value"
              >
                {{ choice.value === 'default' ? t('inherit') : choice.value }}
              </option>
            </select>
            <select
              v-else-if="control.kind === 'optionalBoolean'"
              :id="`param-${control.key}`"
              :value="quickValue(control) == null ? 'unset' : String(quickValue(control))"
              :data-param="control.key"
              @change="setQuick(control, inputValue($event))"
            >
              <option value="unset">{{ t('inherit') }}</option>
              <option value="true">{{ t('yes') }}</option>
              <option value="false">{{ t('no') }}</option>
            </select>
            <span
              v-else-if="control.kind === 'boolean'"
              class="boolean-field"
              ><input
                :id="`param-${control.key}`"
                :checked="Boolean(quickValue(control))"
                type="checkbox"
                :data-param="control.key"
                @change="setQuick(control, ($event.target as HTMLInputElement).checked)"
              /><span>{{ quickValue(control) ? t('yes') : t('no') }}</span></span
            >
            <input
              v-else-if="control.kind === 'integer' || control.kind === 'number'"
              :id="`param-${control.key}`"
              :value="quickValue(control)"
              @input="setQuick(control, inputValue($event))"
              type="number"
              :min="control.min ?? 0"
              :max="control.max"
              :step="control.kind === 'integer' ? 1 : 'any'"
              :data-param="control.key"
            />
            <input
              v-else
              :id="`param-${control.key}`"
              :value="quickValue(control)"
              @input="setQuick(control, inputValue($event))"
              type="text"
              maxlength="500"
              :data-param="control.key"
            />
          </label>
        </div>
        <details
          v-if="fieldCount"
          class="all-fields"
        >
          <summary>
            {{ t('allFields') }} <span>{{ fieldCount }}</span>
          </summary>
          <input
            v-model="fieldFilter"
            type="search"
            :placeholder="t('searchFields')"
            :aria-label="t('searchFields')"
            class="field-search"
          />
          <div
            v-for="binding in visibleBindings"
            :key="binding.id"
            class="binding-fields"
          >
            <h4>{{ binding.id }}</h4>
            <ValueEditor
              :id="binding.id"
              :type="binding.type"
              :fields="binding.fields"
              :model-value="options[binding.id]"
              :fallback="defaults(binding)"
              :filter="fieldFilter"
              @update:model-value="configure(binding.id, $event)"
            />
          </div>
        </details>
      </div>
      <div
        v-if="events.length"
        class="event-log"
        aria-live="polite"
      >
        <div class="config-field-heading">
          <span>{{ t('events') }}</span
          ><button
            type="button"
            @click="events = []"
          >
            {{ t('clear') }}
          </button>
        </div>
        <code
          v-for="(event, index) in events"
          :key="index"
          >{{ event.name }}{{ event.args.length ? `(${event.args.join(', ')})` : '()' }}</code
        >
      </div>
    </div>
  </section>
  <section
    id="code"
    class="code-section"
  >
    <div class="section-title">
      <h2>
        {{ t('code')
        }}<UIcon
          v-if="!synchronized && status === 'ready'"
          name="i-lucide-loader-circle"
          class="animate-spin code-pending"
          :aria-label="t('loading')"
        />
      </h2>
      <UButton
        :disabled="!synchronized"
        :icon="copyState === 'copied' ? 'i-lucide-check' : 'i-lucide-copy'"
        color="neutral"
        variant="ghost"
        size="sm"
        @click="copy"
        >{{ t(copyState) }}</UButton
      >
    </div>
    <div class="code-box">
      <div class="code-tabs">
        <button
          :class="{ selected: codeMode === 'fragment' }"
          :aria-pressed="codeMode === 'fragment'"
          @click="codeMode = 'fragment'"
        >
          {{ t('fragment') }}</button
        ><button
          :class="{ selected: codeMode === 'complete' }"
          :aria-pressed="codeMode === 'complete'"
          @click="codeMode = 'complete'"
        >
          {{ t('complete') }}</button
        ><span>Go</span>
      </div>
      <pre tabindex="0"><code>{{ currentCode }}</code></pre>
    </div>
    <p class="code-hint">{{ t(synchronized ? 'codeHint' : 'defaultCode') }}</p>
  </section>
</template>
