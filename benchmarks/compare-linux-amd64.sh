#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Native comparison of the exact same benchmark bodies used on ARM64, plus
# GOMAXPROCS=32 for the 5950X's hardware threads. Run on an otherwise idle host.
test "$(uname -s)" = Linux
test "$(uname -m)" = x86_64
export GOTOOLCHAIN=go1.27.1
export GOAMD64=v1
task_output=${1:-benchmarks/results/linux-amd64}
mkdir -p "$task_output"
trap 'task_status=$?; printf "%s\n" "$task_status" > "$task_output/exit-status.txt"' EXIT

{
  date -u
  uname -a
  go version
  go env GOOS GOARCH GOAMD64 GOTOOLCHAIN CGO_ENABLED
  lscpu
  for task_file in /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor \
    /sys/devices/system/cpu/cpu0/cpufreq/scaling_driver /sys/devices/system/cpu/smt/active; do
    if [ -f "$task_file" ]; then
      printf '%s: ' "$task_file"
      cat "$task_file"
    fi
  done
  uptime
} > "$task_output/environment.txt"

printf 'Validating native default and spin implementations\n'
go vet ./... > "$task_output/vet.txt" 2>&1
go vet -tags=nsync_spin ./... >> "$task_output/vet.txt" 2>&1
go test -race -shuffle=on -count=10 -timeout=3m ./... > "$task_output/race-default.txt" 2>&1
go test -tags=nsync_spin -race -shuffle=on -count=10 -timeout=3m ./... > "$task_output/race-spin.txt" 2>&1

task_bench='^Benchmark(TryMutex|Semaphore|NamedMutex|OnceMutex|NamedOnceMutex|TryFailureParallel|Construction|ControlWaitGroup|SyncFlag|StandardMutex)$'
printf 'Running baseline throughput and allocation benchmarks\n'
uptime >> "$task_output/load.txt"
go test -tags=nsync_baseline -run '^$' -bench "$task_bench" \
  -benchmem -benchtime=150ms -count=6 -cpu=1,4,16,32 > "$task_output/baseline.txt"
printf 'Running optimized throughput and allocation benchmarks\n'
uptime >> "$task_output/load.txt"
go test -run '^$' -bench "$task_bench" \
  -benchmem -benchtime=150ms -count=6 -cpu=1,4,16,32 > "$task_output/optimized.txt"

printf 'Running baseline and optimized acquisition-latency diagnostics\n'
go test -tags=nsync_baseline -run '^$' -bench '^BenchmarkWaitLatency$' \
  -benchtime=2000000x -count=6 -cpu=4,16,32 > "$task_output/latency-baseline.txt"
go test -run '^$' -bench '^BenchmarkWaitLatency$' \
  -benchtime=2000000x -count=6 -cpu=4,16,32 > "$task_output/latency-optimized.txt"

printf 'Running handwritten assembly and bounded-spin comparisons\n'
go test ./benchmarks/asm -run . -bench . -benchtime=200ms -count=6 -cpu=1 > "$task_output/assembly.txt"
go test -tags=nsync_spin -run '^$' -bench '^BenchmarkTryMutex$/(Uncontended|Contended)' \
  -benchmem -benchtime=150ms -count=6 -cpu=1,4,16,32 > "$task_output/spin.txt"
go test -tags=nsync_spin -run '^$' -bench '^BenchmarkWaitLatency$/TryMutex' \
  -benchtime=2000000x -count=6 -cpu=4,16,32 > "$task_output/latency-spin.txt"
uptime >> "$task_output/load.txt"
printf 'All native AMD64 comparisons completed\n'
