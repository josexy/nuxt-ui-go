import type { HighlighterCore } from 'shiki/core'

// One highlighter for the whole page, loaded on first use in the browser only:
// Shiki's core, its JavaScript regex engine (no Oniguruma WASM), the Go grammar
// and the github-light/github-dark themes that Nuxt Content uses for prose.
let highlighter: Promise<HighlighterCore> | undefined

function load() {
  highlighter ??= Promise.all([
    import('shiki/core'),
    import('shiki/engine/javascript'),
    import('shiki/langs/go.mjs'),
    import('shiki/themes/github-light.mjs'),
    import('shiki/themes/github-dark.mjs')
  ])
    .then(([{ createHighlighterCore }, { createJavaScriptRegexEngine }, go, light, dark]) =>
      createHighlighterCore({
        langs: [go.default],
        themes: [light.default, dark.default],
        engine: createJavaScriptRegexEngine()
      })
    )
    .catch(error => {
      // Let a later call try again, for example after a network failure.
      highlighter = undefined
      throw error
    })
  return highlighter
}

// Recent results, so switching code tabs or remounting a page does not
// highlight the same source again.
const cache = new Map<string, string>()

export function useGoHighlighter() {
  // Returns the code as HTML with both themes' colors as CSS variables, so
  // main.css can switch them with the appearance. Rejects when the
  // highlighter cannot load; callers then keep showing plain text.
  async function codeToHtml(code: string, label = ''): Promise<string> {
    if (!import.meta.client) throw new Error('Go highlighting runs in the browser only')
    const key = `${label}\n${code}`
    const cached = cache.get(key)
    if (cached !== undefined) return cached
    const shiki = await load()
    const html = shiki.codeToHtml(code, {
      lang: 'go',
      themes: { light: 'github-light', dark: 'github-dark' },
      defaultColor: false,
      transformers: [
        {
          pre(node) {
            // The scroll region is focusable, so it needs a name. The panel
            // supplies the surface; main.css ignores the themes' backgrounds.
            if (label) node.properties['aria-label'] = label
          }
        }
      ]
    })
    if (cache.size >= 64) cache.delete(cache.keys().next().value!)
    cache.set(key, html)
    return html
  }
  return { codeToHtml }
}
