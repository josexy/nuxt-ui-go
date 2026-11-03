import { createServer } from 'node:http'
import { readFile, stat } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = fileURLToPath(new URL('../.output/public/', import.meta.url))
const base = `/${(process.env.NUXT_APP_BASE_URL || '/').replace(/^\/+|\/+$/g, '')}`.replace(
  /\/?$/,
  '/'
)
const port = Number(process.env.PORT || 4173)
const types = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript',
  '.mjs': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.wasm': 'application/wasm',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.ico': 'image/x-icon'
}
const server = createServer(async (request, response) => {
  try {
    const url = new URL(request.url, 'http://localhost')
    if (!url.pathname.startsWith(base)) {
      response.writeHead(404)
      response.end('Not found')
      return
    }
    let file = path.resolve(root, decodeURIComponent(url.pathname.slice(base.length)))
    const relative = path.relative(root, file)
    if (relative.startsWith('..') || path.isAbsolute(relative)) {
      response.writeHead(403)
      response.end()
      return
    }
    if ((await stat(file)).isDirectory()) file = path.join(file, 'index.html')
    const content = await readFile(file)
    response.writeHead(200, {
      'Content-Type': types[path.extname(file)] || 'application/octet-stream',
      'Cache-Control': /app\.[a-f0-9]+\.wasm$/.test(file)
        ? 'public,max-age=31536000,immutable'
        : 'no-cache'
    })
    response.end(content)
  } catch {
    response.writeHead(404, { 'Content-Type': 'text/plain' })
    response.end('Not found')
  }
})
server.listen(port, '127.0.0.1', () => console.log(`Local: http://127.0.0.1:${port}${base}`))
process.on('SIGTERM', () => server.close())
process.on('SIGINT', () => server.close())
