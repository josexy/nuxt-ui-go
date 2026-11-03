import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { copyFile, mkdir, readFile, readdir, unlink, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

const website = fileURLToPath(new URL('..', import.meta.url))
const root = path.resolve(website, '..')
const env = { ...process.env, CGO_ENABLED: '0' }
delete env.GOOS
delete env.GOARCH
execFileSync('go', ['run', './internal/cmd/genshowcase'], { cwd: root, env, stdio: 'inherit' })
execFileSync('go', ['run', './internal/cmd/genwebsite'], { cwd: root, env, stdio: 'inherit' })
const directory = path.join(website, 'public', 'preview')
await mkdir(directory, { recursive: true })
const temporary = path.join(root, '.work', 'website', 'app.wasm')
await mkdir(path.dirname(temporary), { recursive: true })
execFileSync(
  'go',
  [
    'build',
    '-trimpath',
    '-tags=mygo_noinspector',
    '-ldflags=-s -w',
    '-o',
    temporary,
    './internal/cmd/sitepreview'
  ],
  {
    cwd: root,
    env: { ...env, GOOS: 'js', GOARCH: 'wasm' },
    stdio: 'inherit'
  }
)
const wasm = await readFile(temporary)
const hash = createHash('sha256').update(wasm).digest('hex').slice(0, 16)
const file = `app.${hash}.wasm`
await writeFile(path.join(directory, file), wasm)
const goroot = execFileSync('go', ['env', 'GOROOT'], { cwd: root, env, encoding: 'utf8' }).trim()
await copyFile(
  path.join(goroot, 'lib', 'wasm', 'wasm_exec.js'),
  path.join(directory, 'wasm_exec.js')
)
const template = await readFile(path.join(website, 'preview', 'index.html'), 'utf8')
await writeFile(path.join(directory, 'index.html'), template.replaceAll('__WASM_FILE__', file))
// Keep each static release small. Preserve obsolete generated binaries outside
// the repository before removing their exact, validated asset paths.
const backup = path.join(tmpdir(), 'nuxt-ui-go-website-wasm-backups')
for (const entry of await readdir(directory, { withFileTypes: true })) {
  if (!entry.isFile() || !/^app\.[a-f0-9]{16}\.wasm$/.test(entry.name) || entry.name === file)
    continue
  const obsolete = path.resolve(directory, entry.name)
  if (path.dirname(obsolete) !== path.resolve(directory))
    throw new Error('Invalid generated asset path')
  await mkdir(backup, { recursive: true })
  await copyFile(obsolete, path.join(backup, entry.name))
  await unlink(obsolete)
}
console.log(
  `Built real Go previews: ${(wasm.length / 1024 / 1024).toFixed(1)} MiB before compression`
)
