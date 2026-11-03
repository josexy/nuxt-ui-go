<script setup lang="ts">
import type { JSONValue, ValueField } from '~/types/catalog'
const props = defineProps<{
  id: string
  type: string
  modelValue?: JSONValue
  fallback?: JSONValue
  choices?: string[]
  slots?: string[]
  fields?: ValueField[]
  filter?: string
}>()
const emit = defineEmits<{ 'update:modelValue': [value: JSONValue | undefined] }>()
const { schemas, matchesField } = useCatalog()
const { t } = useI18n()
const schema = computed(() => schemas[props.type]!)
const current = computed<JSONValue>(() =>
  props.modelValue !== undefined
    ? props.modelValue
    : props.fallback !== undefined
      ? props.fallback
      : zero(props.type)
)
const object = computed(() => asObject(current.value))
const entries = computed(() => (Array.isArray(current.value) ? current.value : []))
const available = computed(() =>
  props.choices?.length
    ? props.choices
    : schema.value.kind === 'map' && props.slots?.length
      ? props.slots
      : schema.value.choices || []
)
const fieldList = computed(() =>
  (props.fields || schema.value.fields || []).filter(
    f => !props.filter || matchesField(f, props.filter)
  )
)
const mode = computed(() => String(object.value.mode || 'none'))
const newKey = ref('')
const fileError = ref('')
function asObject(v: JSONValue | undefined): Record<string, JSONValue> {
  return v && typeof v === 'object' && !Array.isArray(v) ? v : {}
}
function zero(type: string): JSONValue {
  const s = schemas[type]!
  if (
    ['optional', 'pointer', 'icon', 'image', 'handle', 'function', 'scroll', 'window'].includes(
      s.kind
    )
  )
    return null
  if (['number', 'integer', 'duration'].includes(s.kind)) return 0
  if (s.kind === 'color') return '#00000000'
  if (s.kind === 'date') return ''
  if (s.kind === 'array' && s.length) return Array.from({ length: s.length }, () => zero(s.item!))
  return s.sample
}
function sample(type: string): JSONValue {
  const s = schemas[type]!
  if (s.kind === 'array' && s.length) return Array.from({ length: s.length }, () => sample(s.item!))
  if (s.kind === 'object') {
    const value: Record<string, JSONValue> = {}
    for (const field of s.fields || []) {
      if (['Label', 'Title', 'Value'].includes(field.name) && schemas[field.type]?.kind === 'text')
        value[field.name] = 'Example'
      if (field.name === 'Build') value.Build = { mode: 'content', text: 'Example' }
    }
    return value
  }
  return JSON.parse(JSON.stringify(s.sample))
}
function set(value: JSONValue | undefined) {
  emit('update:modelValue', value)
}
function text(event: Event) {
  return (event.target as HTMLInputElement).value
}
function active(field: ValueField): boolean {
  if (!field.when) return true
  const name = field.when.field
  const value = asObject(props.modelValue)[name] ?? asObject(props.fallback)[name]
  return value === field.when.value
}
function patch(key: string, value: JSONValue | undefined) {
  const next = { ...asObject(props.modelValue) }
  if (value === undefined) delete next[key]
  else next[key] = value
  set(next)
}
function replaceMember(key: string, value: JSONValue | undefined) {
  const next = { ...object.value }
  if (value === undefined) delete next[key]
  else next[key] = value
  set(next)
}
function item(index: number, value: JSONValue | undefined) {
  const next = [...entries.value]
  if (value === undefined) next.splice(index, 1)
  else next[index] = value
  set(next)
}
function addEntry() {
  const next =
    available.value.length && schema.value.item === 'string'
      ? available.value[0]!
      : sample(schema.value.item!)
  if (
    next &&
    typeof next === 'object' &&
    !Array.isArray(next) &&
    schemas[schema.value.item!]?.fields?.some(f => f.name === 'ID')
  )
    next.ID = `item-${Date.now().toString(36)}`
  set([...entries.value, next])
}
function addKey() {
  const key = newKey.value.trim()
  if (key && !Object.hasOwn(object.value, key)) replaceMember(key, sample(schema.value.item!))
  newKey.value = ''
}
function callback(nextMode: string) {
  if (nextMode === 'example') {
    set(undefined)
    return
  }
  if (nextMode === 'none') {
    set(null)
    return
  }
  const next: Record<string, JSONValue> = { mode: nextMode }
  if (nextMode === 'content') next.text = 'Example'
  if (nextMode === 'style') next.parts = { [props.slots?.[0] || 'root']: {} }
  if (nextMode === 'format')
    Object.assign(next, { text: '', suffix: '', precision: 2, layout: '2006-01-02' })
  if (nextMode === 'result') next.results = (schema.value.returns || []).map(sample)
  if (nextMode === 'search' || nextMode === 'load') next.items = []
  set(next)
}
function result(index: number, value: JSONValue | undefined) {
  const values = Array.isArray(object.value.results) ? [...object.value.results] : []
  values[index] = value ?? null
  replaceMember('results', values)
}
async function upload(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (file.size > 2_000_000) {
    fileError.value = t('imageSize')
    return
  }
  fileError.value = ''
  const reader = new FileReader()
  reader.onload = () => set(String(reader.result))
  reader.onerror = () => {
    fileError.value = t('imageError')
  }
  reader.readAsDataURL(file)
}
</script>

<template>
  <div
    class="value-editor"
    :data-kind="schema.kind"
  >
    <template v-if="schema.kind === 'object'">
      <component
        :is="fields ? 'div' : 'details'"
        class="nested-fields"
        :open="Boolean(filter)"
      >
        <summary v-if="!fields">{{ type.replace('nui.', '') }}</summary>
        <div class="config-grid">
          <ConfigurationField
            v-for="field in fieldList"
            :key="field.name"
            :field="field"
            :active="active(field)"
            :path="`${id}.${field.name}`"
            :model-value="asObject(modelValue)[field.name]"
            :fallback="asObject(fallback)[field.name]"
            @update:model-value="patch(field.name, $event)"
          />
        </div>
      </component>
    </template>
    <template v-else-if="schema.kind === 'optional' || schema.kind === 'pointer'">
      <div class="optional-value">
        <select
          :id="id"
          :value="current === null ? 'unset' : 'set'"
          :aria-label="id"
          @change="set(text($event) === 'unset' ? null : sample(schema.item!))"
        >
          <option value="unset">{{ t(schema.kind === 'optional' ? 'inherit' : 'none') }}</option>
          <option value="set">{{ t('setValue') }}</option>
        </select>
        <ValueEditor
          v-if="current !== null"
          :id="`${id}.value`"
          :type="schema.item!"
          :model-value="current"
          @update:model-value="set"
        />
      </div>
    </template>
    <select
      v-else-if="schema.kind === 'enum' || (schema.kind === 'text' && available.length)"
      :id="id"
      :value="current"
      @change="set(text($event))"
    >
      <option
        v-for="choice in available"
        :key="choice"
        :value="choice"
      >
        {{ choice.replace(/^(nui|ui|time)\./, '') }}
      </option>
    </select>
    <select
      v-else-if="schema.kind === 'boolean'"
      :id="id"
      :value="String(current)"
      @change="set(text($event) === 'true')"
    >
      <option value="true">true</option>
      <option value="false">false</option>
    </select>
    <input
      v-else-if="['integer', 'number', 'duration'].includes(schema.kind)"
      :id="id"
      :value="current"
      type="number"
      :step="schema.kind === 'integer' ? '1' : 'any'"
      :aria-label="schema.kind === 'duration' ? `${id} (ms)` : id"
      @change="set(Number(text($event)))"
    />
    <input
      v-else-if="schema.kind === 'date'"
      :id="id"
      :value="current"
      type="date"
      @change="set(text($event))"
    />
    <div
      v-else-if="schema.kind === 'color'"
      class="color-value"
    >
      <input
        type="color"
        :aria-label="`${id} color`"
        :value="String(current).slice(0, 7)"
        @input="set(text($event))"
      /><input
        :id="id"
        :value="current"
        pattern="#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?"
        @change="set(text($event))"
      />
    </div>
    <template v-else-if="schema.kind === 'icon'">
      <select
        :id="id"
        :value="current === null ? '' : String(current).includes('<svg') ? 'custom' : current"
        @change="
          set(
            text($event) === 'custom'
              ? '<svg viewBox=&quot;0 0 24 24&quot;><circle cx=&quot;12&quot; cy=&quot;12&quot; r=&quot;8&quot; /></svg>'
              : text($event) || null
          )
        "
      >
        <option value="">{{ t('none') }}</option>
        <option
          v-for="icon in available"
          :key="icon"
        >
          {{ icon }}
        </option>
        <option value="custom">SVG</option>
      </select>
      <textarea
        v-if="String(current).includes('<svg')"
        :aria-label="`${id} SVG`"
        :value="current"
        rows="3"
        @change="set(text($event))"
      />
    </template>
    <div
      v-else-if="schema.kind === 'image'"
      class="image-value"
    >
      <input
        :id="id"
        type="file"
        accept="image/png,image/jpeg,image/gif"
        @change="upload"
      /><img
        v-if="current"
        :src="String(current)"
        alt=""
      /><button
        v-if="current"
        type="button"
        @click="set(null)"
      >
        {{ t('remove') }}
      </button>
      <p
        v-if="fileError"
        role="alert"
      >
        {{ fileError }}
      </p>
    </div>
    <select
      v-else-if="schema.kind === 'handle'"
      :id="id"
      :value="current || ''"
      @change="set(text($event) || null)"
    >
      <option value="">{{ t('none') }}</option>
      <option value="target">{{ t('focusTarget') }}</option>
    </select>
    <select
      v-else-if="schema.kind === 'window'"
      :id="id"
      :value="current || ''"
      @change="set(text($event) || null)"
    >
      <option value="">{{ t('none') }}</option>
      <option value="current">{{ t('currentWindow') }}</option>
    </select>
    <div
      v-else-if="schema.kind === 'scroll'"
      class="config-grid"
    >
      <label
        v-for="axis in ['X', 'Y']"
        :key="axis"
        >{{ axis
        }}<input
          type="number"
          min="0"
          :aria-label="`${id}.${axis}`"
          :value="object[axis] ?? 0"
          @change="replaceMember(axis, Number(text($event)))"
      /></label>
    </div>
    <div
      v-else-if="schema.kind === 'flags'"
      class="flag-values"
    >
      <label
        v-for="flag in available"
        :key="flag"
        ><input
          type="checkbox"
          :checked="entries.includes(flag)"
          @change="
            set(entries.includes(flag) ? entries.filter(v => v !== flag) : [...entries, flag])
          "
        />{{ flag.replace('ui.', '') }}</label
      >
    </div>
    <details
      v-else-if="schema.kind === 'array'"
      class="nested-fields"
    >
      <summary>{{ t('items') }} · {{ entries.length }}</summary>
      <div
        v-for="(entry, index) in entries"
        :key="index"
        class="array-entry"
      >
        <ValueEditor
          :id="`${id}[${index}]`"
          :type="schema.item!"
          :choices="choices"
          :model-value="entry"
          @update:model-value="item(index, $event)"
        /><button
          v-if="!schema.length"
          type="button"
          :aria-label="`${t('remove')} ${index + 1}`"
          @click="item(index, undefined)"
        >
          <UIcon name="i-lucide-x" />
        </button>
      </div>
      <button
        v-if="!schema.length"
        type="button"
        class="add-value"
        @click="addEntry"
      >
        + {{ t('add') }}</button
      ><button
        v-if="schema.length && entries.length !== schema.length"
        type="button"
        @click="set(sample(type))"
      >
        {{ t('setValue') }}
      </button>
    </details>
    <details
      v-else-if="schema.kind === 'map'"
      class="nested-fields"
    >
      <summary>{{ type.replace('nui.', '') }} · {{ Object.keys(object).length }}</summary>
      <div
        v-for="(entry, key) in object"
        :key="key"
        class="map-entry"
      >
        <div class="config-field-heading">
          <code>{{ key }}</code
          ><button
            type="button"
            :aria-label="`${t('remove')} ${key}`"
            @click="replaceMember(key, undefined)"
          >
            <UIcon name="i-lucide-x" />
          </button>
        </div>
        <ValueEditor
          :id="`${id}.${key}`"
          :type="schema.item!"
          :model-value="entry"
          @update:model-value="replaceMember(key, $event)"
        />
      </div>
      <form
        class="add-map-entry"
        @submit.prevent="addKey"
      >
        <select
          v-if="available.length"
          v-model="newKey"
          :aria-label="`${id} ${t('field')}`"
        >
          <option
            value=""
            disabled
          >
            {{ t('field') }}
          </option>
          <option
            v-for="key in available.filter(k => !Object.hasOwn(object, k))"
            :key="key"
          >
            {{ key }}
          </option></select
        ><input
          v-else
          v-model="newKey"
          :aria-label="`${id} ${t('field')}`"
          :placeholder="t('field')"
        /><button
          type="submit"
          :disabled="!newKey"
        >
          {{ t('add') }}
        </button>
      </form>
    </details>
    <template v-else-if="schema.kind === 'function'">
      <select
        :id="id"
        :value="mode"
        @change="callback(text($event))"
      >
        <option value="example">{{ t('exampleBehavior') }}</option>
        <option value="none">{{ t('none') }}</option>
        <option
          v-for="choice in available"
          :key="choice"
          :value="choice"
        >
          {{ t(`callback_${choice}`) }}
        </option>
      </select>
      <textarea
        v-if="mode === 'content'"
        :aria-label="`${id} ${t('text')}`"
        :value="object.text"
        rows="2"
        @input="replaceMember('text', text($event))"
      />
      <ValueEditor
        v-if="mode === 'style'"
        :id="`${id}.parts`"
        type="nui.Styles"
        :choices="slots"
        :model-value="object.parts"
        @update:model-value="replaceMember('parts', $event)"
      />
      <ValueEditor
        v-if="mode === 'search' || mode === 'load'"
        :id="`${id}.items`"
        :type="mode === 'search' ? '[]nui.CommandAction' : '[]nui.TreeNode'"
        :model-value="object.items"
        @update:model-value="replaceMember('items', $event)"
      />
      <div
        v-if="mode === 'format'"
        class="config-grid"
      >
        <label
          >{{ t('prefix')
          }}<input
            :value="object.text"
            @input="replaceMember('text', text($event))" /></label
        ><label
          >{{ t('suffix')
          }}<input
            :value="object.suffix"
            @input="replaceMember('suffix', text($event))" /></label
        ><label v-if="schema.arguments?.some(a => a === 'float64' || a === 'float32')"
          >{{ t('precision')
          }}<input
            type="number"
            min="0"
            max="12"
            :value="object.precision"
            @change="replaceMember('precision', Number(text($event)))" /></label
        ><label v-if="schema.arguments?.includes('time.Time')"
          >{{ t('dateFormat')
          }}<input
            :value="object.layout"
            @change="replaceMember('layout', text($event))"
        /></label>
      </div>
      <div
        v-if="mode === 'result'"
        class="callback-results"
      >
        <ValueEditor
          v-for="(resultType, index) in schema.returns"
          :id="`${id}.results[${index}]`"
          :key="index"
          :type="resultType"
          :model-value="Array.isArray(object.results) ? object.results[index] : undefined"
          @update:model-value="result(index, $event)"
        />
      </div>
    </template>
    <input
      v-else
      :id="id"
      type="text"
      :value="current ?? ''"
      maxlength="2000"
      @input="set(text($event))"
    />
  </div>
</template>
