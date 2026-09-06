#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../.."
export GOTOOLCHAIN=go1.27.1 GOAMD64=v1
test "$(uname -s)" = Linux
test "$(uname -m)" = x86_64
task_output=benchmarks/results/final-amd64
trap 'task_status=$?; printf "%s\n" "$task_status" > "$task_output/recheck-exit-status.txt"' EXIT
task_bench='^Benchmark(OnceMutex|ControlWaitGroup|StandardMutex)$/(First|256|Uncontended|Contended)$'
: > "$task_output/recheck-baseline.txt"
: > "$task_output/recheck-optimized.txt"
for task_round in {1..8}; do
  printf 'Round %s of 8\n' "$task_round"
  uptime >> "$task_output/recheck-load.txt"
  if (( task_round % 2 )); then
    task_order=(baseline optimized)
  else
    task_order=(optimized baseline)
  fi
  for task_mode in "${task_order[@]}"; do
    task_tags=()
    if [ "$task_mode" = baseline ]; then task_tags=(-tags=nsync_baseline); fi
    go test "${task_tags[@]}" -run '^$' -bench "$task_bench" \
      -benchmem -benchtime=500ms -count=1 -cpu=4,16,32 >> "$task_output/recheck-$task_mode.txt"
  done
done
printf 'Alternating confirmation completed\n'
