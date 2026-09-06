#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Run sequentially on an otherwise idle host. These are throughput/allocation
# benchmarks; run BenchmarkWaitLatency separately for sampled waiting times.
task_output=${1:-benchmarks/results/local}
mkdir -p "$task_output"
{
  date -u
  go version
  go env GOOS GOARCH GOAMD64 GOARM64 GOTOOLCHAIN
  git rev-parse HEAD
} > "$task_output/environment.txt"
task_bench='^Benchmark(TryMutex|Semaphore|NamedMutex|OnceMutex|NamedOnceMutex|TryFailureParallel|Construction|ControlWaitGroup|SyncFlag|StandardMutex)$'
go test -tags=nsync_baseline -run '^$' -bench "$task_bench" \
  -benchmem -benchtime=150ms -count=6 -cpu=1,4,16 > "$task_output/baseline.txt"
go test -run '^$' -bench "$task_bench" \
  -benchmem -benchtime=150ms -count=6 -cpu=1,4,16 > "$task_output/optimized.txt"
printf 'Results: %s/{baseline,optimized}.txt\n' "$task_output"
