import { defineContentConfig, defineCollection, z } from '@nuxt/content'

export default defineContentConfig({
  collections: {
    zh: defineCollection({
      type: 'page',
      source: 'zh/**/*.md',
      schema: z.object({ component: z.string() })
    }),
    en: defineCollection({
      type: 'page',
      source: 'en/**/*.md',
      schema: z.object({ component: z.string() })
    })
  }
})
