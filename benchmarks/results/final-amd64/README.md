# Native AMD Ryzen 9 5950X comparison

The optimized implementation substantially improves most measured throughput
workloads on native AMD64. The semaphore read and full-capacity probe regressions
also carry over, and the adaptive mutex generally trades longer acquisition tails
for higher throughput. Optional spinning gives mixed results; handwritten assembly
is slower than compiler intrinsics in the isolated CAS-pair benchmark.

[All benchmark medians](all-results.md) include all 180 throughput/allocation
cases, every latency diagnostic, named-once leaders/op, spin comparisons, and
assembly results. [Benchstat output](comparison.txt) includes confidence intervals
and per-row significance tests. [JSON medians](medians.json) support further analysis.

## Host and method

- AMD Ryzen 9 5950X: 16 physical cores, 32 hardware threads, SMT enabled.
- Native Linux x86_64, Ubuntu kernel 6.8.0-138-generic, Go 1.27.1, GOAMD64=v1.
- Conservative CPU governor, acpi-cpufreq driver, boost enabled. Host settings were preserved.
- The host had background services. The initial load average was 0.71; later
  readings include our benchmarks. See [environment](environment.txt) and [load](load.txt).
- The transferred source matches the implementation and benchmark bodies used
  in the final ARM64 run. All 49 transferred files and the seven frozen baseline
  files were hash-verified. [Source manifest](source-manifest.json) records the
  snapshot; its README hash predates the documentation added with these results.
- Source snapshot retained on the remote host at
  `/home/vlad/workspace/nsync-bench-20260906.DG6nGy`.

The full suite uses 45 workloads × GOMAXPROCS 1, 4, 16, 32 × six samples per
implementation, 150 ms per sample. Baseline and optimized runs execute
sequentially. Acquisition diagnostics use 2,000,000 iterations, six samples,
and GOMAXPROCS 4, 16, 32. No benchmark processes overlapped one another or race tests.

Run from the repository root on native Linux AMD64:

```sh
./benchmarks/compare-linux-amd64.sh benchmarks/results/local-amd64
benchstat benchmarks/results/local-amd64/baseline.txt benchmarks/results/local-amd64/optimized.txt
```

Results below are medians in ns/op. Parallel ns/op measures aggregate throughput,
not one operation's latency. The ARM column compares ARM's baseline with ARM's
optimized result; it is not a direct AMD-versus-ARM timing ratio. CPU, operating
system, timer behavior, and host conditions differ, so these runs do not isolate
an instruction-set effect.

## Representative throughput results

Ratios above 1 mean higher measured throughput. These are microbenchmark results,
not application-wide speedup estimates.

| Workload | Ps | AMD baseline ns/op | AMD optimized ns/op | AMD ratio | ARM ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| TryMutex/Uncontended | 1 | 21.48 | 3.66 | 5.87× | 6.35× |
| TryMutex/Contended0 | 16 | 155.85 | 9.88 | 15.77× | 16.14× |
| TryMutex/Contended100 | 16 | 235.40 | 115.65 | 2.04× | 1.50× |
| Semaphore/1/Uncontended | 1 | 22.45 | 3.84 | 5.85× | 6.39× |
| Semaphore/8/Uncontended | 1 | 23.73 | 9.37 | 2.53× | 2.72× |
| Semaphore/8/Parallel | 16 | 493.65 | 29.19 | 16.91× | 4.13× |
| Semaphore/64/Parallel | 16 | 87.67 | 23.34 | 3.76× | 1.34× |
| Semaphore/8/Parallel | 32 | 677.85 | 40.37 | 16.79× | — |
| Semaphore/64/Parallel | 32 | 93.43 | 25.59 | 3.65× | — |
| NamedMutex/HotKey | 1 | 42.86 | 9.23 | 4.64× | 4.60× |
| NamedMutex/LongKey1024 | 1 | 42.14 | 11.69 | 3.60× | 4.68× |
| NamedMutex/IndependentKeys | 16 | 264.35 | 2.43 | 108.99× | 157.56× |
| NamedOnceMutex/Uncontended | 1 | 88.51 | 28.23 | 3.13× | 3.33× |
| NamedOnceMutex/IndependentKeys | 16 | 398.80 | 2.41 | 165.55× | 170.86× |
| ControlWaitGroup/16 | 16 | 807.70 | 425.85 | 1.90× | 1.23× |

All selected AMD throughput gains above have p=0.002 with six samples in each
build. Some confidence intervals remain wide: capacity-8 semaphore parallel at
16 Ps is 493.65 ns ±21% versus 29.19 ns ±41%. The gain is clear in these samples;
the precise ratio is less stable.

## Regressions, allocations, and controls

- `Semaphore.Value`, one P: **1.337 → 3.783 ns/op**, 2.83× the elapsed time.
- Full capacity-8 semaphore `TryAcquire`, one P: **2.897 → 4.298 ns/op**, 48% more time.
  Both regressions occur at every tested GOMAXPROCS setting (p=0.002).
- `OnceMutex/First` at 16 Ps initially measured **21.56 → 26.22 ns/op**, 22% more
  time. It includes construction and first completion; allocations remain 16 B,
  one event. The longer alternating confirmation below checks this difference.
- `ControlWaitGroup/256` at 16 Ps initially measured **539.35 → 601.05 ns/op**;
  that difference was not significant (p=0.065). At 32 Ps it was also inconclusive.
- The unchanged `StandardMutex/Contended-4` control initially measured
  **11.61 → 17.43 ns/op**, a 50% difference. This is evidence of variation between
  runs, not a library regression. Uncontended controls were statistically unchanged.
  Small changes should be interpreted cautiously on this host.
- Creating 1,024 named mutexes used **223,720 → 166,616 B** at one P, while
  allocation events increased **1,047 → 1,463**. CPU-specific padding makes this
  allocation result differ from ARM64.
- TryMutex construction: **120 B / 2 allocations → 96 B / 1**. Semaphore
  construction: **120 B / 2 → 112 B / 1**. NamedMutex first key:
  **384 B / 4 → 208 B / 2**. NamedOnce first key: **288 B / 4 → 216 B / 2**.
- A parked 10 µs timeout measured approximately **0.67–1.05 ms** on this host
  in both implementations, with no significant timing change. Its allocation
  cost fell from **248 B / 3 allocations to 112 B / 1**. The requested timeout
  is not a promise about how quickly the scheduler resumes a goroutine.
- `NamedOnceMutex/SameKeyParallel-16` changed **0.8603 → 0.9642 leaders/op**.
  One iteration may lead or join an operation, so ns/op alone does not compare
  identical amounts of completed useful work. All leaders/op rows are saved.

## Longer alternating confirmation

The uncertain first-use and large-pool results were repeated alongside unchanged
standard-mutex controls: eight samples of 500 ms each at 4, 16, and 32 Ps.
Each round ran both builds, reversing their order on alternate rounds. These
samples are kept separate from the full suite.

| Workload | Ps | Baseline ns/op | Optimized ns/op | Assessment |
| --- | ---: | ---: | ---: | --- |
| OnceMutex/First | 4 | 21.24 | 23.80 | 12% more time, p=0.001 |
| OnceMutex/First | 16 | 23.27 | 24.85 | Inconclusive, p=0.123 |
| OnceMutex/First | 32 | 22.95 | 23.45 | Inconclusive, p=0.083 |
| ControlWaitGroup/256 | 4 | 561.1 | 525.2 | 6% less time, p=0.010 |
| ControlWaitGroup/256 | 16 | 554.4 | 603.9 | 9% more time, p=0.010; optimized CI ±57% |
| ControlWaitGroup/256 | 32 | 560.0 | 568.8 | Inconclusive, p=0.382 |

Note added in round 3: this confirmation took one sample per process with
`-cpu=4,16,32`, and Go 1.27 measures a `b.Loop` benchmark's first `-cpu` value
at the GOMAXPROCS left by the previous measurement. The 4-P rows of
`OnceMutex/First` and `ControlWaitGroup/256` above therefore ran at 32 Ps.

The initial 22% first-use slowdown at 16 Ps did not reproduce at that magnitude.
The 256-worker case at 16 Ps continued to measure worse, although its optimized
samples were variable. The unchanged contended control at 4 Ps now differed by
5%, compared with 50% in the initial suite. That residual variation limits how
precisely small changes can be attributed to implementation differences.

All eight rounds passed: [comparison](recheck-comparison.txt),
[baseline](recheck-baseline.txt), [optimized](recheck-optimized.txt),
[commands](recheck.sh), [run log](recheck.log), [load](recheck-load.txt),
[exit status 0](recheck-exit-status.txt). All twelve confirmation cases and their
allocations also appear in [the complete table](all-results.md).

## Acquisition latency

Four goroutines per P, one in 64 acquisitions sampled into bounded per-worker
rings. Values are medians of six per-run statistics, in microseconds. Sample
maxima are not worst-case guarantees. Clock sampling and interface dispatch
make these diagnostics different from the uninstrumented throughput benchmarks.

| Work per critical section | Ps | Baseline p99 µs | Optimized p99 µs | Baseline sample max µs | Optimized sample max µs |
| --- | ---: | ---: | ---: | ---: | ---: |
| 0 | 4 | 5.74 | 0.07 | 37.67 | 281.82 |
| 0 | 16 | 20.76 | 111.79 | 186.73 | 429.10 |
| 0 | 32 | 41.61 | 174.06 | 171.05 | 449.81 |
| 100 | 4 | 10.58 | 125.56 | 58.25 | 316.33 |
| 100 | 16 | 29.78 | 145.90 | 69.56 | 378.27 |
| 100 | 32 | 65.38 | 193.74 | 153.74 | 460.52 |

The optimized median wait is only 40–50 ns in these cases, but its sampled
maximum is higher in every TryMutex workload. The four-P empty critical section
has a much lower p99; other cases have higher p99. Faster average throughput does
not imply tighter tail latency. [Full latency comparison](latency-comparison.txt).

## Assembly and optional spin

The native AMD64 CAS acquire/release pair measured **3.590 ns with intrinsics**
versus **3.9315 ns with handwritten assembly**, about 9.5% more time for assembly.
Both report zero allocations. This experiment is separate from the library's
backend; the primitive tests are sequential. [Raw assembly results](assembly.txt).

The optional `nsync_spin` tag remained disabled by default. At 16 Ps, the
100-step throughput workload changed **115.65 → 152.65 ns/op** with spin, 32%
more time (p=0.009); at 32 Ps it was 19% more time (p=0.026). Other main-throughput
cases showed no significant benefit. In the oversubscribed latency diagnostic,
spin improved empty-work p99 at 16 Ps from **111.79 µs to 0.536 µs**, but worsened
100-step p99 there from **145.90 µs to 186.09 µs**. It is workload-dependent.
See [throughput](spin-comparison.txt) and [latency](latency-spin-comparison.txt).

## Native validation

- `go vet ./...` and `go vet -tags=nsync_spin ./...`: passed; [log](vet.txt).
- Ten shuffled native race runs, default: passed; [log](race-default.txt).
- Ten shuffled native race runs, `nsync_spin`: passed; [log](race-spin.txt).
- Every benchmark command completed with PASS; [run log](run.log),
  [exit status 0](exit-status.txt).

No production behavior or public interface changed for this AMD comparison.
A later comment in `gate.go` documents that `limit` is immutable after construction;
the source manifest records the measured snapshot before that comment.
Benchstat version used for analysis: `golang.org/x/perf` at
`v0.0.0-20260825160852-19be9d8e6c70`.
