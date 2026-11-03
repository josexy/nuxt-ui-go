import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { mkdirSync, mkdtempSync } from 'node:fs'
import path from 'node:path'
const root = fileURLToPath(new URL('../../', import.meta.url))
const env = { ...process.env, CGO_ENABLED: '0' }
delete env.GOOS
delete env.GOARCH
execFileSync('go', ['run', './internal/cmd/genwebsite', '-snippets', '.work/website/snippets'], {
  cwd: root,
  env,
  stdio: 'inherit'
})
execFileSync('go', ['build', '-p', '2', './.work/website/snippets/...'], {
  cwd: root,
  env,
  stdio: 'inherit'
})
console.log('All 73 standalone examples compile against the public library')
const work = path.join(root, '.work', 'website')
mkdirSync(work, { recursive: true })
const configured = mkdtempSync(path.join(work, 'configured-'))
execFileSync(
  'go',
  [
    'test',
    './internal/showcase/preview',
    '-run',
    'TestAllFieldEditorsUpdateRealViews|TestClearedFieldSources|TestNestedFieldSources',
    '-count=1'
  ],
  {
    cwd: root,
    env: { ...env, SHOWCASE_SNIPPETS_DIR: configured },
    stdio: 'inherit'
  }
)
execFileSync(
  'go',
  ['test', '-p', '2', `./${path.relative(root, configured).replaceAll('\\', '/')}/...`],
  {
    cwd: root,
    env,
    stdio: ['ignore', 'pipe', 'inherit']
  }
)
console.log('Every public field configuration and data example compiles against the public library')
