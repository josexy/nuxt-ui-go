<script setup lang="ts">
// A palette selector for the preview: each name shows the swatch genwebsite
// took from the Go palette the preview paints with.
defineProps<{ label: string; names: string[] }>()
const model = defineModel<string>({ required: true })
const { swatches } = useCatalog()
</script>

<template>
  <UFormField
    :label="label"
    orientation="horizontal"
    size="sm"
    :ui="{ root: 'items-center gap-2.5', label: 'font-normal text-muted' }"
  >
    <USelectMenu
      v-model="model"
      :items="names"
      :search-input="false"
      size="sm"
      color="neutral"
      class="w-36"
    >
      <template #leading>
        <span
          class="block size-3 rounded-full ring ring-default"
          :style="{ background: swatches[model] }"
        />
      </template>
      <template #item-leading="{ item }">
        <span
          class="size-3 shrink-0 self-center rounded-full ring ring-default"
          :style="{ background: swatches[item] }"
        />
      </template>
    </USelectMenu>
  </UFormField>
</template>
