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
    './internal/showcase/...',
    '-run',
    'TestAllFieldEditorsUpdateRealViews|TestClearedFieldSources|TestNestedFieldSources|TestBlackPrimaryConfiguration',
    '-count=1'
  ],
  {
    cwd: root,
    env: { ...env, SHOWCASE_SNIPPETS_DIR: configured },
    stdio: 'inherit'
  }
)
// There are over 3,000 configured programs. go vet type-checks each one against
// the compiled library without compiling it, which is several times faster than
// go test. These analyzers are the subset go test runs by default.
const vetAnalyzers = [
  'atomic',
  'bools',
  'buildtag',
  'directive',
  'errorsas',
  'ifaceassert',
  'nilfunc',
  'printf',
  'stdversion',
  'stringintconv',
  'tests'
]
execFileSync(
  'go',
  [
    'vet',
    ...vetAnalyzers.map(name => `-${name}`),
    `./${path.relative(root, configured).replaceAll('\\', '/')}/...`
  ],
  {
    cwd: root,
    env,
    stdio: ['ignore', 'pipe', 'inherit']
  }
)
console.log(
  'Every public field configuration and data example type-checks against the public library'
)
