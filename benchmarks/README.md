# Performance and implementation notes

The library uses atomic fast paths and condition variables for parking. Its
production source contains no channels, runtime `linkname` calls, or unsafe
pointer conversions. `unsafe.Sizeof` is used only for cache-line padding. Public
method signatures are preserved, including value copies of `TryMutex` sharing
one underlying lock.

## Reproduce the measurements

The baseline is the reviewed, Go 1.27.1-compatible implementation immediately
before optimization. It is frozen in [`internal/baseline`](../internal/baseline).
[`baseline-sha256.json`](baseline-sha256.json) records the original source hashes:
remove the added first comment line and change `package baseline` back to
`package nsync` before comparing them. All seven hashes were verified.

The same concrete-type benchmark bodies select the baseline with the
`nsync_baseline` build tag. Setup and allocations are excluded where `b.Loop`
permits; constructor and cold-key benchmarks intentionally measure them.

```sh
./benchmarks/run.sh benchmarks/results/local
benchstat benchmarks/results/local/baseline.txt benchmarks/results/local/optimized.txt
```

The script runs six samples of each workload at GOMAXPROCS 1, 4, and 16, with
150 ms per sample. Run it on an otherwise idle host. Do not overlap benchmark
processes, cross-compilation, or race tests. `RunParallel` ns/op is aggregate
throughput, not individual operation latency; a result below 1 ns does not mean
one acquisition completes in less than a nanosecond.

Development host: Apple M3 Max, 16 cores, macOS, Go 1.27.1, default GOARM64=v8.0.
ARM64 measurements are native. The default ARM64 binary selects LSE atomics on
supported CPUs and retains the compiler's portable fallback. Darwin AMD64 tests
on this host run under Rosetta and establish no native x86 performance claim.

The same source and frozen baseline were also measured natively on an AMD Ryzen
9 5950X, Linux, Go 1.27.1, GOAMD64=v1, at GOMAXPROCS 1, 4, 16, and 32. See the
[AMD64 comparison report](results/final-amd64/README.md) and
[all AMD64 benchmark medians](results/final-amd64/all-results.md), including
allocations, acquisition latency, handwritten assembly, and optional spinning.
To repeat that suite on native Linux AMD64:

```sh
./benchmarks/compare-linux-amd64.sh benchmarks/results/local-amd64
```

Final raw throughput/allocation results are in [`results/final-arm64`](results/final-arm64).
The other result files record exploratory candidates, sometimes with shorter
runs or different benchmark subsets. They are not interchangeable with the
final comparison. The unchanged `sync.Mutex` controls also help expose host
noise between runs. The checked-in environment records the original repository
HEAD; the optimized source is the accompanying working-tree change.

## Measured ARM64 results

Median ns/op over six samples; lower is better. “Ps” means GOMAXPROCS.
Parallel rows report aggregate throughput.

| Workload | Ps | Baseline | Optimized | Throughput ratio |
| --- | ---: | ---: | ---: | ---: |
| TryMutex Lock/Unlock | 1 | 18.73 | 2.95 | 6.35× |
| TryMutex, empty critical section (parallel) | 16 | 107.90 | 6.69 | 16.14× |
| TryMutex, 100 arithmetic steps (parallel) | 16 | 418.75 | 279.55 | 1.50× |
| Semaphore, capacity 8 | 1 | 19.96 | 7.35 | 2.72× |
| Semaphore, capacity 64 (parallel) | 16 | 36.62 | 27.41 | 1.34× |
| NamedMutex, one hot key | 1 | 33.98 | 7.39 | 4.60× |
| NamedMutex, repeated 1 KB key | 1 | 36.55 | 7.80 | 4.68× |
| NamedMutex, independent keys (parallel) | 16 | 232.80 | 1.48 | 157.56× |
| NamedOnceMutex, repeated operation | 1 | 73.09 | 21.94 | 3.33× |
| NamedOnceMutex, independent keys (parallel) | 16 | 277.05 | 1.62 | 170.86× |
| ControlWaitGroup, pool 256 | 16 | 364.85 | 297.60 | 1.23× |

Full confidence intervals, allocation counts, and all benchmark rows are in
[`comparison.txt`](results/final-arm64/comparison.txt).

These are local microbenchmarks, not application-wide speedup predictions.
`NamedOnceMutex/SameKeyParallel` additionally reports `leaders/op`: one iteration
may lead an operation or join an existing one. Changes in that ratio mean its
ns/op is not a direct comparison of completed useful work.

Allocation and performance tradeoffs:

- Reading `Semaphore.Value()` increased from 1.87 to 2.60 ns/op. General counters are now protected by a mutex.
- Trying a full capacity-8 semaphore increased from 2.13 to 3.34 ns/op. General counters are now protected by a mutex.
- Creating 1,024 names used about 166.8 KiB versus 218.5 KiB, but increased allocation events from 1047 to 1478. Sharding trades more small allocations for less contention.
- Padding and embedded shard storage make empty named-once instances larger. The first-key construction benchmark measures the populated case; it is not a universal memory-saving claim.
- A parked 10 µs timeout has similar elapsed time, with allocations reduced from 3 to 1 and bytes from 248 to 112 per call. Tiny already-expired deadlines usually avoid a timer entirely.
- Repeated uncontended named-once operations went from 1 allocation per operation to zero after warm-up. TryMutex construction went from two allocations to one, and first-key construction of both named types went from four to two.
- `SyncFlag` was already close to the atomic instruction cost and remains essentially unchanged. Serial iteration over 1,024 named locks also changed little.

## Acquisition latency

A separate diagnostic uses four goroutines per P and samples one acquisition
in 64 into bounded per-worker rings. It reports sample p50, p99, and maximum.
It includes interface dispatch and clock sampling, so its ns/op should not be
compared with uninstrumented throughput benchmarks. These samples characterize
the tested workload; they do not establish a worst-case scheduling bound.

```sh
go test -tags=nsync_baseline -run '^$' -bench '^BenchmarkWaitLatency$' \
  -benchtime=2000000x -count=6 -cpu=4,16 > latency-baseline.txt
go test -run '^$' -bench '^BenchmarkWaitLatency$' \
  -benchtime=2000000x -count=6 -cpu=4,16 > latency-optimized.txt
benchstat latency-baseline.txt latency-optimized.txt
```

The first three-state mutex candidate permitted sustained reacquisition by a
running goroutine. With 64 contenders and a short arithmetic critical section,
the median of six sampled maxima was about 539 ms. That candidate was rejected.
The retained implementation adaptively reserves a handoff for a waiting
acquisition. New arrivals cannot consume a reserved handoff. A canceling timed
waiter passes the reservation onward, or clears it if it is the last waiter.

Thresholds of 25, 100, 250, and 1,000 microseconds were measured. The 25 microsecond
candidate substantially reduced long waits but cost throughput with 64 busy
contenders. The 100 microsecond threshold balances these objectives. It does
not guarantee FIFO ordering or completion within 100 microseconds.

Median of six runs, 64 goroutines on 16 Ps; times are microseconds:

| Critical section | Baseline p99 | Optimized p99 | Baseline sample max | Optimized sample max |
| --- | ---: | ---: | ---: | ---: |
| Empty | 15.2 | 61.8 | 78.9 | 457.6 |
| 100 arithmetic steps | 36.9 | 198.4 | 74.4 | 417.5 |

See [`latency-comparison.txt`](results/latency-comparison.txt) for all latency
metrics. The baseline's more even handoff still has tighter tails in these
workloads. Choose based on the application's latency and throughput requirements.

## Retained design

- **Mutex:** one compare-and-swap on uncontended `Lock` and `Unlock`. Failed
  `TryLock` first reads the state, avoiding a cache-line write when many callers
  probe a held lock. Slow paths register under a queue mutex and park through
  `sync.Cond`. Four states distinguish free, held, held with waiters, and reserved
  handoff. The queue mutex protects waiter counts, lazy condition initialization,
  and the handoff mode. The last canceling waiter can race with release, so the
  release slow path accepts both held states.
- **Semaphore:** capacity one uses the mutex directly. Higher capacities keep
  a plain count under that mutex. This measured better than racing CAS/fetch-add
  updates under heavy contention. Slot condition variables are allocated lazily.
  `Value` reads the general counter under the same mutex.
- **NamedMutex:** 64 lazily allocated shards, a direct immutable cache for the
  first key, and another immutable first-key slot per shard. Other names
  use a typed map under the shard mutex. Long names use a suffix check before
  full comparison, avoiding unnecessary scans of differing names with long
  common prefixes. Cache-line padding separates frequently read metadata from
  lock writes. As before, created names are retained for the instance's lifetime.
- **OnceMutex:** completed calls need an atomic load. A condition variable is
  allocated only when an incomplete operation has concurrent waiters.
- **NamedOnceMutex:** 64 lazy shards, with the first shard embedded and a direct
  slot for its first active key. Only entries never observed by another caller
  may be reused. Shared entries remain alive through their waiters' references;
  a later operation cannot reset an old waiter's completion state. Completed key
  references are cleared; one unused entry per shard may be retained.
- **Timeouts:** reuse callback state through `sync.Pool`, but create a new timer
  when parking is necessary. Already-expired tiny deadlines skip timer creation.
  When a timer callback has started, cancellation waits for its protected work
  before recycling state. Timer objects are not shared across synctest bubbles.
- **ControlWaitGroup:** separate condition variables for worker slots and group
  completion, guarded by one standard `sync.Mutex`. This avoids waking a
  completion waiter when a queued submitter needs a slot. Running and queued counts make a separate
  `sync.WaitGroup` redundant. Abort is permanent; already admitted functions can
  finish. Each accepted task still starts its own goroutine, preserving that API
  behavior.

## Rejected or optional experiments

| Candidate | Outcome |
| --- | --- |
| CAS-only semaphore counters | Cache-line contention and retry loops hurt multi-core throughput. |
| Fetch-add reservations with rollback | Better than CAS retries in some cases; still lost to serialized short counter updates. An early rollback path deadlocked and was discarded. |
| `sync.Map` for retained named locks | Better than the original global map lock, but typed shards and immutable first-key caches improved hot paths and cold allocation costs further. |
| Per-shard condition broadcasts and reference-counted named-once entries | Removed allocations but created wakeup contention; per-entry completion and reuse of unshared entries performed better. |
| Using the adaptive mutex for worker counters | The standard mutex measured better for larger goroutine pools and was retained there. |
| Hashing every long name | Regressed repeated 1 KB keys; the first-key cache now handles long names with a suffix precheck. |
| Mutex without adaptive handoff | Excellent average throughput, unacceptable sustained barging in the latency diagnostic. |
| Twenty-five-microsecond handoff threshold | Tighter tails, but a substantial throughput cost in the oversubscribed work benchmark. |
| Assembly polling before parking | The 16-round ARM64 experiment produced mixed throughput results across workloads and runs. Disabled by default. |
| Handwritten atomic acquire/release assembly | Compared with the same operations expressed through compiler intrinsics; the function boundary costs more than the inlined native instructions. |

The optional `nsync_spin` build tag retains bounded ARM64 `YIELD` and AMD64
`PAUSE` experiments, with a portable fallback. It never spins indefinitely and
is not a recommended default. Measure your own workload before using it.

The isolated [`asm`](asm) package compares handwritten ARM64/AMD64 CAS pairs
with Go intrinsics; it is not the library's lock backend. Its primitive tests
are sequential and do not make it a race-instrumented synchronization API.
On LSE-capable ARM64 hardware:

```sh
GOARM64=v8.1 go test ./benchmarks/asm -run . -bench . \
  -benchtime=200ms -count=6 -cpu=1
```

The final CAS-pair medians were **2.05 ns with intrinsics** and **3.03 ns with
handwritten assembly** (six samples, one P). Both use native LSE instructions
in that comparison. Results are in
[`assembly-arm64-lse.txt`](results/assembly-arm64-lse.txt). The additional call
boundaries do not pay for themselves, so the library retains inlined intrinsics.

These experiments cover the main available algorithm, contention, allocation,
layout, inlining, and assembly choices for the current API. Further tuning would
need a particular application workload or native measurements on additional
hardware; the results do not prove a universal fastest implementation.

## Correctness and portability

Final checks passed:

- ARM64: 50 shuffled race runs of the default backend and 20 of `nsync_spin`.
- AMD64 through Rosetta: 20 shuffled race runs in each build mode.
- Native Linux AMD64, Ryzen 9 5950X: 10 shuffled race runs and `go vet` in each build mode; [logs](results/final-amd64/README.md).
- `go vet ./...` with default and `nsync_spin` builds.
- All 16 cross-compilation targets, in both modes.
- Ten coverage runs: more than 97% statement coverage in the root package; see
  [`coverage.txt`](results/coverage.txt). Historical baseline code is excluded
  from that percentage.

Tests exercise mutual exclusion and publication, concurrent lazy creation,
forced hash collisions, once-generation isolation, semaphore limits, abort and
submission races, multiple completion waiters, goroutine `Goexit`, invalid
operations, timeout/release races, handoff cancellation, and timer callback
recycling. `testing/synctest` makes deadline and queue tests deterministic. An AST
check rejects channels in production source, and layout tests check cache-line
padding invariants. Tests and the frozen historical baseline may use channels.

[`cross-build.sh`](cross-build.sh) compiles default and `nsync_spin` library/test
binaries for 16 targets: Linux AMD64, ARM64, ARMv7, 386, RISC-V64, PPC64LE, s390x,
Loong64, MIPS64LE; macOS AMD64/ARM64; Windows AMD64/ARM64; FreeBSD AMD64/ARM64;
and JS/Wasm. Compilation is not execution. Cache-line sizes follow Go's
architecture-specific CPU definitions, including 128 bytes on ARM64 and
256 bytes on s390x.

The GitHub Actions matrix is configured for native Linux x86_64/ARM64, macOS
ARM64, and Windows x86_64 race tests, plus cross-compilation. Those remote jobs
have not been run by this local change.
