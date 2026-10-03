#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Interleaved comparison of a git revision with the working tree. Both builds
# run the working tree's benchmark_test.go, so the benchmark bodies are
# identical; benchmarks of APIs the revision lacks belong in other files.
# Rounds alternate the build order, so drifting host load affects both alike.
#
# Usage: benchmarks/compare-revisions.sh <revision> <output-dir> [rounds] [bench] [cpus] [benchtime]
task_rev=${1:?revision to compare against}
task_output=${2:?output directory}
task_rounds=${3:-6}
task_bench=${4:-'^Benchmark(TryMutex|Semaphore|NamedMutex|OnceMutex|NamedOnceMutex|TryFailureParallel|Construction|ControlWaitGroup|SyncFlag|StandardMutex)$'}
task_cpus=${5:-1,4,16,32}
task_benchtime=${6:-150ms}

task_work=$(mktemp -d "${TMPDIR:-/tmp}/nsync-compare.XXXXXX")
trap 'rm -rf "$task_work"' EXIT
mkdir -p "$task_output" "$task_work/base"
git archive "$task_rev" | tar -x -C "$task_work/base"
cp benchmark_test.go "$task_work/base/"
(cd "$task_work/base" && go test -c -o "$task_work/base.test" .)
go test -c -o "$task_work/new.test" .

{
  date -u
  uname -a
  go version
  go env GOOS GOARCH GOAMD64 GOARM64
  grep -m1 'model name' /proc/cpuinfo 2>/dev/null || sysctl -n machdep.cpu.brand_string 2>/dev/null || true
  for task_file in /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor /sys/devices/system/cpu/cpu0/cpufreq/scaling_driver; do
    if [ -f "$task_file" ]; then
      printf '%s: %s\n' "$task_file" "$(cat "$task_file")"
    fi
  done
  printf 'base: %s\n' "$(git rev-parse "$task_rev")"
  printf 'new: %s plus working tree changes:\n' "$(git rev-parse HEAD)"
  git status --short
  uptime
} > "$task_output/environment.txt"

: > "$task_output/base.txt"
: > "$task_output/new.txt"
IFS=, read -r -a task_cpu_list <<< "$task_cpus"
for ((task_i = 0; task_i < task_rounds; task_i++)); do
  if ((task_i % 2 == 0)); then task_order="base new"; else task_order="new base"; fi
  # One process per GOMAXPROCS value: given a -test.cpu list, Go 1.27 runs
  # a b.Loop benchmark's first value at the GOMAXPROCS of the previous run.
  for task_cpu in "${task_cpu_list[@]}"; do
    for task_build in $task_order; do
      "$task_work/$task_build.test" -test.run '^$' -test.bench "$task_bench" -test.benchmem \
        -test.benchtime "$task_benchtime" -test.count 1 -test.cpu "$task_cpu" >> "$task_output/$task_build.txt"
    done
  done
done
uptime >> "$task_output/environment.txt"
printf 'Results: benchstat %s/base.txt %s/new.txt\n' "$task_output" "$task_output"
