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
function setQuick(control: Control, raw: string | number | boolean | null) {
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
        : (raw as string | boolean)
}
// Choices for a select quick parameter. "__custom__" stands for a value set in
// All fields that the list does not offer; it shows but cannot be chosen.
function choiceItems(control: Control) {
  const items = (control.choices || []).map(c => ({
    label: c.value === 'default' ? t('inherit') : c.value,
    value: c.value,
    disabled: false
  }))
  return quickValue(control) === '__custom__'
    ? [{ label: t('customValue'), value: '__custom__', disabled: true }, ...items]
    : items
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
const synchronized = ref(false)
const codeTabs = computed(() => [
  { value: 'fragment', label: t('fragment'), code: snippet.value },
  { value: 'complete', label: t('complete'), code: source.value }
])
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
</script>

<template>
  <section>
    <div class="mb-3 flex items-center justify-between gap-3">
      <h2
        id="preview"
        class="text-2xl font-bold text-highlighted"
      >
        {{ t('preview') }}
      </h2>
      <UButton
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="ghost"
        size="sm"
        @click="reset"
        >{{ t('reset') }}</UButton
      >
    </div>
    <div class="overflow-hidden rounded-lg border border-default">
      <div
        class="flex flex-wrap items-center gap-x-5 gap-y-2 border-b border-default bg-elevated/50 px-4 py-2.5"
      >
        <PalettePicker
          v-model="primaryColor"
          :label="t('primary')"
          :names="primary"
        />
        <PalettePicker
          v-model="neutralColor"
          :label="t('neutral')"
          :names="neutral"
        />
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
        <template #fallback>
          <div
            class="grid place-items-center text-sm text-muted"
            :style="{ height: `${demo.height}px` }"
          >
            <span class="flex items-center gap-2"
              ><UIcon
                name="i-lucide-loader-circle"
                class="size-5 animate-spin"
              />{{ t('loading') }}</span
            >
          </div>
        </template>
      </ClientOnly>
      <div class="border-t border-default p-4 sm:p-5">
        <h3
          id="parameters"
          class="text-base font-semibold text-highlighted"
        >
          {{ t('parameters') }}
        </h3>
        <p class="mt-1 mb-4 text-sm text-muted">{{ t('parameterHint') }}</p>
        <div class="grid grid-cols-1 gap-x-5 gap-y-4 sm:grid-cols-2 2xl:grid-cols-3">
          <UFormField
            v-for="control in demo.controls"
            :key="control.key"
            :label="local(control.label)"
            size="sm"
            :ui="{ hint: 'min-w-0' }"
          >
            <template #hint>
              <code class="truncate font-mono text-xs text-muted">{{ control.key }}</code>
            </template>
            <USelect
              v-if="control.kind === 'select'"
              :model-value="String(quickValue(control))"
              :items="choiceItems(control)"
              :data-param="control.key"
              class="w-full"
              @update:model-value="setQuick(control, $event)"
            />
            <USelect
              v-else-if="control.kind === 'optionalBoolean'"
              :model-value="quickValue(control) == null ? 'unset' : String(quickValue(control))"
              :items="[
                { label: t('inherit'), value: 'unset' },
                { label: t('yes'), value: 'true' },
                { label: t('no'), value: 'false' }
              ]"
              :data-param="control.key"
              class="w-full"
              @update:model-value="setQuick(control, $event)"
            />
            <USwitch
              v-else-if="control.kind === 'boolean'"
              :model-value="Boolean(quickValue(control))"
              :data-param="control.key"
              @update:model-value="setQuick(control, $event)"
            />
            <UInputNumber
              v-else-if="control.kind === 'integer' || control.kind === 'number'"
              :model-value="Number(quickValue(control))"
              :min="control.min ?? 0"
              :max="control.max"
              :step="1"
              :step-snapping="control.kind === 'integer'"
              :format-options="{ useGrouping: false, maximumFractionDigits: 6 }"
              :data-param="control.key"
              class="w-full"
              @update:model-value="setQuick(control, $event)"
            />
            <UInput
              v-else
              :model-value="String(quickValue(control) ?? '')"
              maxlength="500"
              :data-param="control.key"
              class="w-full"
              @update:model-value="setQuick(control, String($event))"
            />
          </UFormField>
        </div>
        <UCollapsible
          v-if="fieldCount"
          class="mt-5 border-t border-default pt-3"
        >
          <UButton
            color="neutral"
            variant="ghost"
            trailing-icon="i-lucide-chevron-down"
            class="group -ms-2.5"
            :ui="{
              trailingIcon: 'transition-transform duration-200 group-data-[state=open]:rotate-180'
            }"
          >
            {{ t('allFields') }}
            <UBadge
              :label="String(fieldCount)"
              color="neutral"
              variant="subtle"
              size="sm"
            />
          </UButton>
          <template #content>
            <UInput
              v-model="fieldFilter"
              type="search"
              icon="i-lucide-search"
              size="sm"
              :placeholder="t('searchFields')"
              :aria-label="t('searchFields')"
              class="my-3 w-full"
            />
            <div
              v-for="binding in visibleBindings"
              :key="binding.id"
              class="mt-5 first:mt-2"
            >
              <h4 class="mb-3 font-mono text-xs font-medium text-muted">{{ binding.id }}</h4>
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
          </template>
        </UCollapsible>
      </div>
      <div
        aria-live="polite"
        :class="events.length ? 'border-t border-default px-4 py-3 sm:px-5' : ''"
      >
        <template v-if="events.length">
          <div class="mb-1.5 flex items-center justify-between gap-2">
            <h4 class="text-sm font-semibold text-highlighted">{{ t('events') }}</h4>
            <UButton
              color="neutral"
              variant="link"
              size="xs"
              icon="i-lucide-trash-2"
              @click="events = []"
              >{{ t('clear') }}</UButton
            >
          </div>
          <ol class="space-y-0.5 font-mono text-xs text-muted">
            <li
              v-for="(event, index) in events"
              :key="index"
              class="wrap-anywhere"
            >
              {{ event.name }}{{ event.args.length ? `(${event.args.join(', ')})` : '()' }}
            </li>
          </ol>
        </template>
      </div>
    </div>
  </section>
  <UAlert
    color="neutral"
    variant="subtle"
    icon="i-lucide-info"
    :title="t('platform')"
    :description="t('platformNote')"
  />
  <section>
    <h2
      id="code"
      class="mb-3 text-2xl font-bold text-highlighted"
    >
      {{ t('code') }}
    </h2>
    <CodePanel
      v-model:tab="codeMode"
      :tabs="codeTabs"
      badge="Go"
      :pending="!synchronized && status === 'ready'"
      :copyable="synchronized"
    />
    <p class="mt-2 text-sm text-muted">{{ t(synchronized ? 'codeHint' : 'defaultCode') }}</p>
  </section>
</template>
