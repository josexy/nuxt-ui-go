<script setup lang="ts">
const { t, locale, setLocale } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const colorMode = useColorMode()
const { components, local, categories } = useCatalog()
const query = useState('component-search', () => '')
const menuOpen = ref(false)
const searchInput = ref<HTMLInputElement>()
const filtered = computed(() =>
  components.filter(c =>
    `${c.name} ${c.description.zh} ${c.description.en}`
      .toLowerCase()
      .includes(query.value.trim().toLowerCase())
  )
)
watch(
  () => route.path,
  () => {
    menuOpen.value = false
  }
)
defineShortcuts({
  meta_k: () => {
    menuOpen.value = true
    nextTick(() => searchInput.value?.focus())
  },
  escape: () => {
    menuOpen.value = false
  }
})
async function changeLanguage() {
  await setLocale(locale.value === 'zh' ? 'en' : 'zh')
}
</script>

<template>
  <a
    class="skip-link"
    href="#main-content"
    >{{ t('components') }}</a
  >
  <header class="site-header">
    <div class="header-inner">
      <UButton
        class="mobile-menu"
        icon="i-lucide-menu"
        color="neutral"
        variant="ghost"
        :aria-label="t('menu')"
        :aria-expanded="menuOpen"
        @click="menuOpen = !menuOpen"
      />
      <NuxtLink
        class="brand"
        :to="localePath('/')"
      >
        <span class="brand-symbol"><UIcon name="i-lucide-braces" /></span>
        <strong>NuxtUI <span>Go</span></strong>
      </NuxtLink>
      <span class="header-description">{{ t('brand') }}</span>
      <div class="header-actions">
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-lucide-languages"
          :aria-label="t('language')"
          @click="changeLanguage"
          >{{ locale === 'zh' ? 'EN' : '中文' }}</UButton
        >
        <ClientOnly>
          <select
            v-model="colorMode.preference"
            class="appearance-select"
            :aria-label="t('appearance')"
          >
            <option value="system">{{ t('system') }}</option>
            <option value="light">{{ t('light') }}</option>
            <option value="dark">{{ t('dark') }}</option>
          </select>
          <template #fallback><span class="appearance-placeholder" /></template>
        </ClientOnly>
        <UButton
          as="a"
          href="https://github.com/josexy/nuxt-ui-go"
          target="_blank"
          rel="noopener noreferrer"
          icon="i-lucide-github"
          color="neutral"
          variant="ghost"
          :aria-label="t('github')"
        />
      </div>
    </div>
  </header>
  <div class="docs-shell">
    <button
      v-if="menuOpen"
      class="mobile-scrim"
      :aria-label="t('menu')"
      @click="menuOpen = false"
    />
    <aside
      class="component-sidebar"
      :class="{ open: menuOpen }"
      :aria-label="t('components')"
    >
      <div class="search-field">
        <UIcon name="i-lucide-search" /><input
          ref="searchInput"
          v-model="query"
          type="search"
          :placeholder="t('search')"
          :aria-label="t('searchLabel')"
        /><kbd>⌘ K</kbd>
      </div>
      <nav>
        <NuxtLink
          class="overview-link"
          :to="localePath('/')"
          ><UIcon name="i-lucide-grid-2x2" />{{ t('overview') }}</NuxtLink
        >
        <p
          v-if="!filtered.length"
          class="search-empty"
        >
          {{ t('noResults') }}
        </p>
        <section
          v-for="category in categories"
          :key="category"
          class="nav-group"
        >
          <template v-if="filtered.some(c => c.category === category)">
            <h2>{{ t(category) }}</h2>
            <NuxtLink
              v-for="component in filtered.filter(c => c.category === category)"
              :key="component.id"
              :to="localePath(`/components/${component.id}`)"
              :title="local(component.description)"
              :aria-current="route.params.id === component.id ? 'page' : undefined"
              >{{ component.name }}</NuxtLink
            >
          </template>
        </section>
      </nav>
    </aside>
    <main
      id="main-content"
      class="main-content"
    >
      <slot />
    </main>
  </div>
  <footer class="site-footer">
    <span>{{ t('footer') }}</span
    ><span>MyGo 0.7.4</span>
  </footer>
</template>
