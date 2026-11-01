#!/usr/bin/env bash
set -euo pipefail

hidden=0
case "${1:-}" in
    '') ;;
    --hidden-smoke) hidden=1 ;;
    --help|-h)
        printf 'Usage: bash scripts/test-native.sh [--hidden-smoke]\nRequires a Linux desktop session; hidden smoke still needs a display.\n'
        exit 0 ;;
    *) printf 'Unknown argument: %s\n' "$1" >&2; exit 2 ;;
esac
if (( $# > 1 )); then
    printf 'Expected at most one argument.\n' >&2
    exit 2
fi
if [[ -z "${DISPLAY:-}" && -z "${WAYLAND_DISPLAY:-}" ]]; then
    printf 'Native tests require a Linux desktop session (DISPLAY or WAYLAND_DISPLAY).\n' >&2
    exit 1
fi

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd -- "$project_root"
export CGO_ENABLED=0 MYGO_LOCALE=en NUI_NATIVE_HIDDEN="$hidden"
unset GOOS GOARCH
export NUI_NATIVE_ARTIFACTS="$project_root/.work/native"
mkdir -p -- "$NUI_NATIVE_ARTIFACTS"
go build -tags e2e -o "$NUI_NATIVE_ARTIFACTS/gallery-e2e" ./examples/gallery

# Keep the Unix socket path short even when the checkout has a long path.
socket_dir="$(mktemp -d /tmp/nui-native.XXXXXXXX)"
gallery_pid=''
cleanup() {
    local status=$?
    if [[ -n "$gallery_pid" ]]; then
        kill "$gallery_pid" 2>/dev/null || true
        wait "$gallery_pid" 2>/dev/null || true
    fi
    rm -f -- "$socket_dir/gallery.sock"
    rmdir -- "$socket_dir"
    return "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
export MYGO_AUTOMATION="$socket_dir/gallery.sock"
export NUI_AUTOMATION_SOCKET="$MYGO_AUTOMATION"
args=(-static)
if (( hidden )); then args+=(-hidden); fi
"$NUI_NATIVE_ARTIFACTS/gallery-e2e" "${args[@]}" \
    > "$NUI_NATIVE_ARTIFACTS/stdout.log" 2> "$NUI_NATIVE_ARTIFACTS/stderr.log" &
gallery_pid=$!
if ! go test ./internal/nativeprobe -run '^TestNativeGallery$' -v -count=1; then
    printf 'Native validation failed; see %s\n' "$NUI_NATIVE_ARTIFACTS" >&2
    exit 1
fi
