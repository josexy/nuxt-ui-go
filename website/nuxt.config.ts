import catalog from './app/data/catalog.json'

const baseURL = process.env.NUXT_APP_BASE_URL || '/'
export default defineNuxtConfig({
  compatibilityDate: '2026-11-02',
  devtools: { enabled: false },
  modules: ['@nuxt/ui', '@nuxtjs/i18n', '@nuxt/content'],
  css: ['~/assets/main.css'],
  app: {
    baseURL,
    head: { title: 'NuxtUI Go', meta: [{ name: 'theme-color', content: '#16a34a' }] }
  },
  ui: { fonts: false },
  content: { experimental: { sqliteConnector: 'native' } },
  icon: { clientBundle: { scan: true }, serverBundle: { collections: ['lucide'] } },
  i18n: {
    defaultLocale: 'zh',
    strategy: 'prefix_except_default',
    detectBrowserLanguage: false,
    locales: [
      { code: 'zh', language: 'zh-Hans', name: '简体中文', file: 'zh.json' },
      { code: 'en', language: 'en', name: 'English', file: 'en.json' }
    ]
  },
  nitro: {
    prerender: {
      routes: [
        '/',
        '/en',
        ...catalog.components.flatMap(c => [`/components/${c.id}`, `/en/components/${c.id}`])
      ],
      crawlLinks: true
    },
    output: { publicDir: '.output/public' }
  },
  typescript: { strict: true }
})
