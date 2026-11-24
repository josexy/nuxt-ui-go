<script setup lang="ts">
import type { JSONValue, ValueField } from '~/types/catalog'
const props = defineProps<{
  field: ValueField
  modelValue?: JSONValue
  fallback?: JSONValue
  path: string
  active?: boolean
}>()
const emit = defineEmits<{ 'update:modelValue': [value: JSONValue | undefined] }>()
const { local, schemas } = useCatalog()
const { t } = useI18n()
const configured = computed(() => props.modelValue !== undefined)
</script>

<template>
  <fieldset
    class="config-field"
    :data-field="path"
    :disabled="active === false"
  >
    <div class="config-field-heading">
      <label
        :for="path"
        :title="local(field.description)"
        ><code class="font-mono">{{ field.name }}</code
        ><span
          v-if="schemas[field.type]?.kind === 'duration'"
          class="field-annotation"
          >ms</span
        ><span
          v-if="field.hint"
          class="field-annotation"
          >{{ local(field.hint) }}</span
        ><span
          v-if="active === false && field.when"
          class="field-annotation"
          >{{ local(field.when.label) }}</span
        ></label
      >
      <UButton
        v-if="configured"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="ghost"
        size="xs"
        :aria-label="`${t('reset')} ${field.name}`"
        @click="emit('update:modelValue', undefined)"
      />
    </div>
    <ValueEditor
      :id="path"
      :type="field.type"
      :choices="field.choices"
      :slots="field.slots"
      :model-value="modelValue"
      :fallback="fallback"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </fieldset>
</template>
