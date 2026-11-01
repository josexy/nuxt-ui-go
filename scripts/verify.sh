#!/usr/bin/env bash
set -euo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd -- "$project_root"
export CGO_ENABLED=0
unset GOOS GOARCH

mapfile -d '' -t go_files < <(
    find . -maxdepth 1 -type f -name '*.go' -print0
    find icons internal examples -type f -name '*.go' -print0
)
for ((offset = 0; offset < ${#go_files[@]}; offset += 100)); do
    unformatted="$(gofmt -l "${go_files[@]:offset:100}")"
    if [[ -n "$unformatted" ]]; then
        printf 'gofmt required:\n%s\n' "$unformatted" >&2
        exit 1
    fi
done

go test ./... -count=1
go vet ./...
go tool mygo vet .

# Include file names as well as contents, so additions and removals also fail.
generated_fingerprint() {
    {
        find icons/lucide -maxdepth 1 -type f -name '*.go' -print0
        printf '%s\0' palette_gen.go
    } | LC_ALL=C sort -z | xargs -0 sha256sum
}
before="$(generated_fingerprint)"
go generate ./...
after="$(generated_fingerprint)"
if [[ "$before" != "$after" ]]; then
    printf 'Generated sources changed; regenerate and include the updated files.\n' >&2
    exit 1
fi

for target in windows/amd64 linux/amd64 darwin/arm64; do
    GOOS="${target%/*}" GOARCH="${target#*/}" go build ./...
    printf 'Built %s with CGO_ENABLED=0\n' "$target"
done
