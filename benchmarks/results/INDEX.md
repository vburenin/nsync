# Benchmark result files

## Round 3 (current code) on native AMD64

Round 3 against `6e6bc42`, the round 2 implementation, measured with
[`compare-revisions.sh`](../compare-revisions.sh) in ten interleaved rounds,
with one process per GOMAXPROCS value and the `performance` governor;
[load.txt](round3-amd64/load.txt) records the load average every 30 s.

- [Report](round3-amd64/README.md)
- Throughput and allocations: [medians](round3-amd64/table-full.md), [comparison.txt](round3-amd64/comparison.txt), [base.txt](round3-amd64/base.txt), [new.txt](round3-amd64/new.txt), [environment.txt](round3-amd64/environment.txt)
- Acquisition latency: [medians](round3-amd64/table-latency.md), [comparison.txt](round3-amd64/latency/comparison.txt), [base.txt](round3-amd64/latency/base.txt), [new.txt](round3-amd64/latency/new.txt), [environment.txt](round3-amd64/latency/environment.txt)
- Context-aware acquisition: [medians](round3-amd64/table-context.md), [context.txt](round3-amd64/context.txt), [context-plain.txt](round3-amd64/context-plain.txt)
- Hardware counters: [table](round3-amd64/table-counters.md), [counters.tsv](round3-amd64/counters.tsv)
- Against the channel implementation: [medians](round3-amd64/table-cumulative.md), [comparison.txt](round3-amd64/cumulative/comparison.txt), [channels.txt](round3-amd64/cumulative/channels.txt), [new.txt](round3-amd64/cumulative/new.txt)
- `nsync_spin` against the default build: [medians](round3-amd64/table-spin.md), [spin-comparison.txt](round3-amd64/spin-comparison.txt), [spin-default.txt](round3-amd64/spin-default.txt), [spin.txt](round3-amd64/spin.txt)
- 32-bit x86 (`GOARCH=386`) on the same host, in six rounds: [medians](round3-amd64/386/table-full.md), [comparison.txt](round3-amd64/386/comparison.txt), [base.txt](round3-amd64/386/base.txt), [new.txt](round3-amd64/386/new.txt), [environment.txt](round3-amd64/386/environment.txt); the [run log](round3-amd64/run.log) covers it
- 32-bit x86 with the earlier 64-bit lock word: [comparison.txt](round3-amd64/386-64bit-word/comparison.txt), [environment.txt](round3-amd64/386-64bit-word/environment.txt), [run.log](round3-amd64/386-64bit-word/run.log)
- Code alignment check (`ControlWaitGroup`, identical instructions placed 32 bytes apart): [comparison.txt](round3-amd64/alignment-check/comparison.txt), [before.txt](round3-amd64/alignment-check/before.txt), [after.txt](round3-amd64/alignment-check/after.txt)
- Final-code check (the measured build against the final code, ten rounds): [comparison.txt](round3-amd64/final-code-check/comparison.txt), [run.log](round3-amd64/final-code-check/run.log)
- [Validation of the final code](round3-amd64/validation.txt): tests, vet, cross-builds, and soak runs
- [Run log](round3-amd64/run.log), [load.txt](round3-amd64/load.txt)

## Round 2

Saved results from Go 1.27.1 on an Apple M3 Max (native ARM64) and an AMD Ryzen 9 5950X (native Linux AMD64).
The final throughput suite has 45 workloads × 6 samples per implementation, at 3 GOMAXPROCS settings on ARM64 and 4 on AMD64.
Parallel ns/op measures aggregate throughput. Earlier experiments may use different subsets, sample counts, or candidate implementations.
Go 1.27 measures a `b.Loop` benchmark's first `-cpu` value at the GOMAXPROCS left by the previous measurement. In the final throughput suites, which ran six samples per process with a `-cpu` list, the first of the six 1-P samples of each `b.Loop` benchmark therefore ran at the list's last value; it is often an outlier, which moves a median of six only slightly. The recheck files took one sample per process with `-cpu=4,16,32`, so all of their 4-P samples of `b.Loop` benchmarks (`OnceMutex/First`, `ControlWaitGroup/256`, `StandardMutex/Uncontended`) ran at 32 Ps.

### Native AMD64 comparisons

- [Comparison report](final-amd64/README.md)
- [All benchmark medians](final-amd64/all-results.md)
- [baseline.txt](final-amd64/baseline.txt)
- [optimized.txt](final-amd64/optimized.txt)
- [comparison.txt](final-amd64/comparison.txt)
- [medians.json](final-amd64/medians.json)
- [latency-baseline.txt](final-amd64/latency-baseline.txt)
- [latency-optimized.txt](final-amd64/latency-optimized.txt)
- [latency-comparison.txt](final-amd64/latency-comparison.txt)
- [assembly.txt](final-amd64/assembly.txt)
- [spin.txt](final-amd64/spin.txt)
- [spin-comparison.txt](final-amd64/spin-comparison.txt)
- [latency-spin.txt](final-amd64/latency-spin.txt)
- [latency-spin-comparison.txt](final-amd64/latency-spin-comparison.txt)
- [environment.txt](final-amd64/environment.txt)
- [load.txt](final-amd64/load.txt)
- [source-manifest.json](final-amd64/source-manifest.json)
- [race-default.txt](final-amd64/race-default.txt)
- [race-spin.txt](final-amd64/race-spin.txt)
- [vet.txt](final-amd64/vet.txt)
- [run.log](final-amd64/run.log)
- [exit-status.txt](final-amd64/exit-status.txt)
- [process.pid](final-amd64/process.pid)
- [recheck-baseline.txt](final-amd64/recheck-baseline.txt)
- [recheck-optimized.txt](final-amd64/recheck-optimized.txt)
- [recheck-comparison.txt](final-amd64/recheck-comparison.txt)
- [recheck.sh](final-amd64/recheck.sh)
- [recheck.log](final-amd64/recheck.log)
- [recheck-load.txt](final-amd64/recheck-load.txt)
- [recheck-exit-status.txt](final-amd64/recheck-exit-status.txt)

### Final ARM64 comparisons and samples

- [assembly-arm64-lse.txt](assembly-arm64-lse.txt)
- [final-arm64/baseline.txt](final-arm64/baseline.txt)
- [final-arm64/comparison.txt](final-arm64/comparison.txt)
- [final-arm64/environment.txt](final-arm64/environment.txt)
- [final-arm64/medians.json](final-arm64/medians.json)
- [final-arm64/optimized.txt](final-arm64/optimized.txt)
- [final-arm64/source-sha256.json](final-arm64/source-sha256.json)
- [latency-baseline-arm64.txt](latency-baseline-arm64.txt)
- [latency-comparison.txt](latency-comparison.txt)
- [latency-optimized-arm64.txt](latency-optimized-arm64.txt)
- [spin-final-arm64.txt](spin-final-arm64.txt)
- [spin-final-comparison.txt](spin-final-comparison.txt)

### Earlier optimization experiments

- [assembly-cas-swap-arm64-lse.txt](assembly-cas-swap-arm64-lse.txt)
- [atomic-cond-arm64.txt](atomic-cond-arm64.txt)
- [baseline-arm64.txt](baseline-arm64.txt)
- [combined-arm64.txt](combined-arm64.txt)
- [compact-arm64.txt](compact-arm64.txt)
- [fetch-add-arm64.txt](fetch-add-arm64.txt)
- [fetch-add-comparison.txt](fetch-add-comparison.txt)
- [first-key-cache-arm64.txt](first-key-cache-arm64.txt)
- [handoff-100us-arm64.txt](handoff-100us-arm64.txt)
- [handoff-250us-arm64.txt](handoff-250us-arm64.txt)
- [handoff-25us-arm64.txt](handoff-25us-arm64.txt)
- [handoff-arm64.txt](handoff-arm64.txt)
- [inline-once-arm64.txt](inline-once-arm64.txt)
- [latency-handoff-1ms-arm64.txt](latency-handoff-1ms-arm64.txt)
- [latency-handoff-1ms-comparison.txt](latency-handoff-1ms-comparison.txt)
- [latency-unfair-arm64.txt](latency-unfair-arm64.txt)
- [lazy-tables-arm64.txt](lazy-tables-arm64.txt)
- [long-cache-standard-cwg-arm64.txt](long-cache-standard-cwg-arm64.txt)
- [plain-counter-arm64.txt](plain-counter-arm64.txt)
- [pre-long-cache-cwg-arm64.txt](pre-long-cache-cwg-arm64.txt)
- [serialized-counter-arm64.txt](serialized-counter-arm64.txt)
- [serialized-counter-comparison.txt](serialized-counter-comparison.txt)
- [sharded-named-arm64.txt](sharded-named-arm64.txt)
- [sharded-named-comparison.txt](sharded-named-comparison.txt)
- [sharded-once-arm64.txt](sharded-once-arm64.txt)
- [sharded-once-fastlock-arm64.txt](sharded-once-fastlock-arm64.txt)
- [specialized-arm64.txt](specialized-arm64.txt)
- [spin16-arm64.txt](spin16-arm64.txt)
- [spin16-comparison.txt](spin16-comparison.txt)
- [three-state-arm64.txt](three-state-arm64.txt)
- [unshared-reuse-arm64.txt](unshared-reuse-arm64.txt)

### Validation and disassembly

- [coverage.txt](coverage.txt)
- [cross-build.txt](cross-build.txt)
- [trymutex-lock-arm64.asm.txt](trymutex-lock-arm64.asm.txt)
- [trymutex-unlock-arm64.asm.txt](trymutex-unlock-arm64.asm.txt)
