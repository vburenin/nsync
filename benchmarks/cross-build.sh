#!/usr/bin/env bash
set -euo pipefail

# Compile the library and its tests, including portable spin fallbacks.
# These binaries are not executed: successful compilation is not a runtime test.
task_output=$(mktemp -d "${TMPDIR:-/tmp}/nsync-cross.XXXXXX")
trap 'rm -rf "$task_output"' EXIT
for task_target in \
  linux/amd64 linux/arm64 linux/arm linux/386 linux/riscv64 \
  linux/ppc64le linux/s390x linux/loong64 linux/mips64le \
  darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 \
  freebsd/amd64 freebsd/arm64 js/wasm; do
  task_os=${task_target%/*}
  task_arch=${task_target#*/}
  for task_tags in default nsync_spin; do
    GOOS=$task_os GOARCH=$task_arch GOARM=7 CGO_ENABLED=0 \
      go test -p=2 -tags="$task_tags" -c -o "$task_output/$task_os-$task_arch-$task_tags.test" .
  done
  printf '%s compiled (default and nsync_spin)\n' "$task_target"
done
