<script setup lang="ts">
import type { ButtonProps } from '@nuxt/ui'

const { t, locale, setLocale } = useI18n()
const localePath = useLocalePath()
const { mygo } = useCatalog()
const { navigation, searchGroups } = useNavigation()
const menuOpen = ref(false)
// UHeader binds this object onto its menu button, so the button states whether
// the navigation panel is open. aria-expanded is an attribute, not a prop.
const toggle = computed(() => ({ 'aria-expanded': menuOpen.value }) as ButtonProps)
const repository = 'https://github.com/josexy/nuxt-ui-go'
// Search matches the name and both languages' descriptions as substrings, as
// the sidebar filter it replaces did. Each category can list all its entries.
const fuse = {
  fuseOptions: { keys: ['label', 'keywords'], threshold: 0, ignoreLocation: true },
  resultLimit: 100
}
async function changeLanguage() {
  await setLocale(locale.value === 'zh' ? 'en' : 'zh')
}
// Move focus, not only the scroll position, past the header and sidebar.
function skipToContent(event: MouseEvent) {
  const main = document.getElementById('main-content')
  if (!main) return
  event.preventDefault()
  main.focus()
}
</script>

<template>
  <a
    href="#main-content"
    class="sr-only focus:not-sr-only focus:fixed focus:start-4 focus:top-3 focus:z-[60] focus:rounded-md focus:bg-default focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-highlighted focus:ring focus:ring-accented"
    @click="skipToContent"
    >{{ t('skipToContent') }}</a
  >
  <UHeader
    v-model:open="menuOpen"
    :to="localePath('/')"
    :toggle="toggle"
    :ui="{ right: 'gap-0.5 sm:gap-1.5' }"
  >
    <template #left>
      <NuxtLink
        :to="localePath('/')"
        class="flex shrink-0 items-center gap-1.5 text-lg font-bold text-highlighted sm:text-xl"
      >
        <UIcon
          name="i-lucide-braces"
          class="size-6 text-primary"
        />
        <span>NuxtUI <span class="text-primary">Go</span></span>
      </NuxtLink>
      <span class="ms-3 hidden border-s border-default ps-3 text-sm text-muted xl:inline">{{
        t('brand')
      }}</span>
    </template>

    <UContentSearchButton
      :label="t('search')"
      :collapsed="false"
      :kbds="['meta', 'k']"
      class="w-72"
    />

    <template #right>
      <UContentSearchButton
        :label="t('search')"
        collapsed
        class="lg:hidden"
      />
      <UButton
        color="neutral"
        variant="ghost"
        icon="i-lucide-languages"
        class="whitespace-nowrap"
        :ui="{ leadingIcon: 'max-sm:hidden' }"
        @click="changeLanguage"
        ><span class="sr-only">{{ t('language') }}: </span
        ><span :lang="locale === 'zh' ? 'en' : 'zh-Hans'">{{
          locale === 'zh' ? 'EN' : '中文'
        }}</span></UButton
      >
      <UColorModeSelect
        :aria-label="t('appearance')"
        color="neutral"
        variant="ghost"
        :ui="{ base: 'max-sm:pe-1.5', value: 'max-sm:sr-only', trailing: 'max-sm:hidden' }"
      />
      <UButton
        :to="repository"
        target="_blank"
        icon="i-lucide-github"
        color="neutral"
        variant="ghost"
        :aria-label="`${t('github')} ${t('newTab')}`"
      />
    </template>

    <template #body>
      <UContentNavigation
        :navigation="navigation"
        highlight
      />
    </template>
  </UHeader>

  <UMain>
    <slot />
  </UMain>

  <UFooter>
    <template #left>
      <p class="text-sm text-muted">{{ t('footer') }}</p>
    </template>
    <p class="text-sm text-muted">MyGo {{ mygo }}</p>
    <template #right>
      <UButton
        :to="repository"
        target="_blank"
        icon="i-lucide-github"
        color="neutral"
        variant="ghost"
        :aria-label="`${t('github')} ${t('newTab')}`"
      />
    </template>
  </UFooter>

  <ClientOnly>
    <UContentSearch
      :groups="searchGroups"
      :fuse="fuse"
      :color-mode="false"
      :title="t('searchTitle')"
      :placeholder="t('searchLabel')"
      preserve-group-order
    />
  </ClientOnly>
</template>
