# Round 3 on native AMD64

Round 3 against `6e6bc42`, the round 2 implementation, on an AMD Ryzen 9 5950X
(16 cores, 32 threads), Linux 7.0, Go 1.27.1, `GOAMD64=v1`. For the run, every
core used the `performance` governor and energy-performance preference of the
`amd-pstate-epp` driver, with boost on; the host's usual `powersave` and
`balance_power` settings were restored afterwards ([run.log](run.log)). No
other benchmark or build ran; the host's background services kept running.
[load.txt](load.txt) samples the load average every 30 s: 0.7 when the run
began and at most 6.1 during it, in the throughput phase.

These are microbenchmarks, not predictions of application speedups. Ps is
GOMAXPROCS, and parallel rows report aggregate throughput.

## Method

```sh
./benchmarks/compare-revisions.sh 6e6bc42 benchmarks/results/round3-amd64 10
./benchmarks/compare-revisions.sh 6e6bc42 benchmarks/results/round3-amd64/latency 10 \
  '^BenchmarkWaitLatency$' 4,16,32 2000000x
```

- Both builds run the working tree's `benchmark_test.go`, so the benchmark
  bodies are identical. The measured round 3 build differs from the final code
  only in a later fix to misused-`Release` slow paths and in tests; see the
  final-code check below.
- Ten rounds. In each round both builds run every benchmark at each GOMAXPROCS
  value, one process per value, and the build that runs first alternates
  between rounds. Each sample lasts 150 ms, so every row has ten samples per
  build.
- Medians, 95% confidence intervals, and Mann–Whitney U p-values come from
  `benchstat`. A difference is called significant at p < 0.05. With 184 rows
  some significant differences are expected by chance, and code layout alone
  moves identical code: unchanged `SyncFlag` and uncontended `sync.Mutex` rows
  differ by up to 0.8%, some of them significantly, and sub-nanosecond loops by
  much more (below). Contended `sync.Mutex` medians at 4–32 Ps differ by up to
  9%, none significantly.
- A separate process per GOMAXPROCS value is necessary. Given a `-test.cpu`
  list, Go 1.27 measures a `b.Loop` benchmark's first value at the GOMAXPROCS
  that the previous measurement left behind. An earlier version of this report
  passed `-test.cpu 1,4,16,32` and took one sample per process, so every 1-P
  row of a `b.Loop` benchmark had actually run at 32 Ps; for example, it showed
  `ControlWaitGroup` with 256 slots as 1.41× faster at 1 P, where it is
  0.91×. `RunParallel` benchmarks were not affected.
- Rows under a few nanoseconds depend on code alignment. In a check of two
  builds whose `ControlWaitGroup.Do` after `Abort` loops have identical
  instructions, the loop that starts 32 bytes earlier fits in one 64-byte cache
  line and runs in 0.46 ns instead of 0.62–0.63 ns ([alignment
  check](alignment-check/comparison.txt)). A ten-round A/B of the measured
  build against the final code ([final-code
  check](final-code-check/comparison.txt)) shows the same effect: completed
  `OnceMutex`, a failed `TryAcquire` on a full `Semaphore(8)`, `Do` after
  `Abort`, and a failed `TryLock` in parallel take 27–36% longer in the final
  build. Other single-goroutine rows move by at most 6.5%, and parallel rows by
  −8% to +18% (both extremes are key-per-goroutine rows; contended nsync rows
  move by at most 3.1%, and the unchanged contended `sync.Mutex` control by
  7%). In the final build, the four rows are about 2.7×, 4.9×, and 2.3× faster
  than round 2 and, for the parallel failed `TryLock`, 0.75–0.82× as fast,
  against 3.6×, 6.5×, and 3.2× in the table below and 1.00–1.04× in
  [table-full.md](table-full.md). Treat ratios of rows under a nanosecond as
  approximate.
- Hardware counters come from `perf stat` at GOMAXPROCS 1, counting user-space
  events only: each benchmark runs with two fixed iteration counts, and the
  difference in counts divided by the difference in iterations gives counts
  per operation (median of three).

## Summary

Of 184 rows (46 workloads at 4 GOMAXPROCS values), round 3 is significantly
faster in 130, slower in 22, and unchanged in 32. The geometric mean of ns/op
fell 36%. Over the 44 rows that allocate, the geometric mean of allocated bytes
fell 48% and of allocations 22%; most of that drop comes from the two
`NamedMutex` workloads that create names, and without the non-steady-state
`UniqueNames` rows it is 39% and 18%.

Ratios of 6e6bc42 ns/op to round 3 ns/op, so values above 1× mean round 3 is
faster. Ratios in parentheses are not significant; † marks rows that move by
about a quarter with code alignment alone; in the final build, a failed
`TryAcquire` is about 4.9×, completed `OnceMutex` 2.7×, and `Do` after `Abort`
2.3× (see Method).

| Workload | 1 P | 4 Ps | 16 Ps | 32 Ps |
| --- | ---: | ---: | ---: | ---: |
| TryMutex Lock/Unlock, uncontended | (0.99×) | 1.00× | (1.00×) | (1.00×) |
| TryMutex, empty critical section | (1.00×) | 1.09× | 1.16× | 1.15× |
| TryMutex, 100 work steps | (1.00×) | 1.02× | 1.04× | (1.03×) |
| TryLockTimeout on a free lock | 0.97× | 0.98× | 0.98× | 0.99× |
| TryLockTimeout, already expired | 1.09× | 1.10× | 1.10× | 1.10× |
| Semaphore(1) uncontended | 0.97× | 0.97× | 0.97× | 0.97× |
| Semaphore(8) uncontended | 2.39× | 2.38× | 2.37× | 2.38× |
| Semaphore(8) parallel | 2.38× | 2.03× | 3.30× | 3.95× |
| Semaphore(64) parallel | 2.39× | 1.99× | 2.85× | 2.86× |
| Semaphore.Value | 7.74× | 7.78× | 7.73× | 7.75× |
| TryAcquire on a full Semaphore(8) † | 6.49× | 6.50× | 6.51× | 6.48× |
| NamedMutex, one hot name | 1.36× | 1.40× | 1.40× | 1.39× |
| NamedMutex, 1,024 names in turn | 1.28× | 1.32× | 1.32× | 1.31× |
| NamedMutex, a name per goroutine | 1.36× | 1.19× | (1.06×) | (1.00×) |
| NamedMutex, a new name each time | 5.49× | 5.14× | 5.14× | 4.96× |
| NamedMutex, create 1,024 names | 2.79× | 3.09× | 3.28× | 3.24× |
| OnceMutex, completed † | 3.64× | 3.63× | 3.61× | 3.59× |
| NamedOnceMutex, one key | 1.55× | 1.55× | 1.55× | 1.55× |
| NamedOnceMutex, same key | 1.52× | 1.60× | 2.06× | 2.05× |
| NamedOnceMutex, a key per goroutine | 1.86× | 1.23× | (1.07×) | (1.02×) |
| NewTryMutex | 1.73× | 1.97× | 2.18× | 2.22× |
| ControlWaitGroup(1) | 0.92× | 0.90× | 0.90× | 0.91× |
| ControlWaitGroup(16) | 0.91× | 1.07× | 1.09× | 1.12× |
| ControlWaitGroup(256) | 0.91× | 1.17× | 2.66× | 1.40× |
| ControlWaitGroup.Do after Abort † | 3.17× | 3.19× | 3.18× | 3.16× |

`ControlWaitGroup(256)` at 16 Ps rests on a bimodal 6e6bc42 sample (384 ns
±43%), and at 32 Ps on a skewed one (201 ns ±68%); its 1-P and 4-P rows are
steadier. The `ControlWaitGroup(16)` gains at 4–32 Ps are marginal (p =
0.015–0.023) and within the spread seen between two builds of the same code.
The rows marked † are the alignment-sensitive ones described above;
`Semaphore.Value`, also under a nanosecond, did not move between builds.

Memory: `NewTryMutex` allocates 24 B instead of 96 B, and `NewSemaphore` 48 B
instead of 112 B. A `NamedMutex` that locks a new name each time used to
allocate about 165 B in two objects per name and keep all of them. It now
keeps at most 129 idle locks and allocates about 17 B per name: the 16-byte
block for the name string that the benchmark builds, plus a new 64-byte lock
about once every 64 releases in a shard, when the shard replaces its cached
lock. Because the old cost grows with the number of names retained, that row
is not a steady-state figure for `6e6bc42`. Creating 1,024 names allocated
about 168,000 B in 1,470 objects; it now takes 21,136 B in 195.

Every row: [medians](table-full.md), [benchstat](comparison.txt),
[6e6bc42 samples](base.txt), [round 3 samples](new.txt),
[environment](environment.txt).

## Regressions

- **`ControlWaitGroup` when every completion must wake the submitter**: with
  one slot at any GOMAXPROCS, 0.90–0.92× (255–266 → 284–291 ns), and with 16 or
  256 slots at GOMAXPROCS 1, 0.91× (262–263 → 287–288 ns). At one P the
  submitter parks whenever the pool is full and the next completion wakes it,
  so every task costs a park, a wakeup, and a goroutine start, whatever the
  pool size. The hardware counters at one P show 176 more instructions, 119
  more cycles, and one more branch miss per task (3,045 → 3,221 instructions,
  1,247 → 1,366 cycles). A per-function `perf` profile taken during
  development (not published) put the extra instructions in the admission and
  completion slow paths, which maintain the FIFO queue and context support on
  every park and wakeup; this was not isolated further. Round 2 let woken
  submitters compete with newcomers through one shared condition variable.
  With 4 or more Ps, round 3 is faster: marginally with 16 slots (1.07–1.12×)
  and 1.17–2.66× with 256, where the larger ratios rest on unstable round 2
  samples.
- **`Semaphore(1)`**: 0.97× uncontended (3.60–3.64 → 3.71–3.75 ns) and 0.98×
  in parallel at 1 P; in parallel at 16 and 32 Ps it is 1.10× faster. Round 2
  special-cased one permit as a bare compare-and-swap in both directions.
  Round 3 loads the word and compares it with the capacity before its
  compare-and-swap, which avoids failing atomic instructions when other
  permits are held, and releases with an atomic add. Either change may be the
  cause; neither was isolated.
- **`TryLock` and `TryLockTimeout` on a free lock**: `TryLockTimeout` is
  0.97–0.99× (3.50–3.56 → 3.59–3.61 ns) and `TryLock` 0.99× at 1, 16, and 32
  Ps (+0.9–1.5%). Both builds take the lock with a single compare-and-swap,
  and the cause was not isolated. `TryLock`'s difference is close to the shifts
  of unchanged code; `TryLockTimeout`'s, up to 2.8% at 1 P, is larger.
- **Tail latency** of contended `TryMutex` (next section): with an empty
  critical section, p99.9 nearly doubles (1.9× and 2.0×) at 16 and 32 Ps and
  the longest sampled waits are 2–2.5× as long at every GOMAXPROCS; with 100
  work steps, p99.9 rises 63% at 32 Ps and p99 rises 1–9%.
- Allocation: a `TryLockTimeout` that waits for its full timeout averages 1 B
  more at 1 P (116 → 117 B). Both builds allocate a timer for each such wait.

## Acquisition latency

Four goroutines per P and 2,000,000 acquisitions per run. Each goroutine times
every 64th of its acquisitions into a ring that keeps its last 8,192 samples.
Values are medians of ten per-run statistics, 6e6bc42 → round 3. The
percentiles are over sampled acquisitions, so goroutines that acquire more
often weigh more; they do not show how long each goroutine waits. Sample maxima
are not worst-case bounds, and values below 200 ns move in steps of about
10 ns, the resolution of the clock. Only `TryMutex` was measured; `Semaphore`
and `NamedMutex` use the same lock core and probably behave alike.

| Workload | Ps | ns/op | p50 ns | p99 µs | p99.9 µs | Sample max µs |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| TryMutex/Work0 | 4 | 8.1 → 8.1 | 21 → 20 | 0.04 → 0.04 | 114.9 → 72.7 | 265.2 → 669.6 |
| TryMutex/Work0 | 16 | 9.3 → 8.7 | 20 → 20 | 0.04 → 0.09 | 135.0 → 250.6 | 453.6 → 954.6 |
| TryMutex/Work0 | 32 | 10.8 → 10.0 | 20 → 20 | 114.24 → 0.10 | 197.3 → 396.6 | 549.5 → 1076.5 |
| TryMutex/Work100 | 4 | 91.4 → 90.9 | 20 → 20 | 113.36 → 114.70 | 129.8 → 126.7 | 254.7 → 210.7 |
| TryMutex/Work100 | 16 | 104.0 → 101.0 | 30 → 20 | 126.46 → 131.57 | 162.4 → 199.0 | 359.3 → 283.3 |
| TryMutex/Work100 | 32 | 121.4 → 120.2 | 30 → 20 | 148.94 → 161.62 | 206.5 → 336.3 | 415.5 → 415.5 |

With an empty critical section:

- At 16 and 32 Ps, round 3 takes 6% and 8% less time and p99 at 32 Ps falls
  from 114 µs to 0.1 µs, but p99.9 nearly doubles (1.9× and 2.0×: 135 → 251 µs,
  197 → 397 µs; every round 3 run is above every round 2 run) and the sample
  maximum is 2.1× and 2.0× as high.
- At 4 Ps, the time is unchanged and the sample maximum is 2.5× as high: every
  round 3 run reached 0.60–0.94 ms, nine of ten round 2 runs 0.19–0.36 ms.
  p99.9 is bimodal in round 3 (five runs at 0.6–35 µs and five at 110–128 µs,
  against 94–116 µs in nine of ten round 2 runs), so its median is not
  representative.

With 100 work steps:

- The time changes by at most 3%; only the 3% drop at 16 Ps is significant.
  p99 rises 1%, 4%, and 9% at 4, 16, and 32 Ps.
- p99.9 rises 63% at 32 Ps (every round 3 run is above every round 2 run). At
  16 Ps it is bimodal in round 3: six runs at 174–204 µs, near round 2's
  153–203 µs, and four at 0.47–1.0 ms, with sample maxima of 0.9–2.2 ms, so the
  median maximum's fall from 359 to 283 µs does not mean shorter waits. At 4
  Ps, p99.9 is unchanged and the sample maximum falls 17%.

The benchmark is noisy in its tails: the unchanged `sync.Mutex` control moved
significantly by up to 30% (with 100 work steps at 4 Ps, its p99.9 fell 24% and
its maximum 30%), so smaller tail changes are within noise. Work was also
spread very unevenly in some runs of both builds. A run keeps fewer than about
31,250 samples only when a goroutine overflowed its ring, and with 100 work
steps at 4 and 16 Ps, two to six of the ten runs per build did; in the worst
run of each build (17,727 and 22,300 samples kept), the one or two goroutines
that overflowed made at least 69% and 54% of all acquisitions.

In the design, running goroutines may take a just-released lock first, and
only a woken waiter that loses a race can switch the lock to direct handoff,
after it has waited 100 µs. A waiter whose wakeup is delayed therefore holds up
the queue behind it until it runs again, which can take longer than 100 µs.
This explanation was not isolated by measurement. Exploratory runs of other
handoff policies on the shared host had six samples per build, too few to
settle their effect on these tails.

For comparison, the `sync.Mutex` controls in the same runs have longer tails
with an empty critical section: p99.9 of 156, 460, and 1,165 µs and sample
maxima of 0.47, 1.21, and 2.40 ms at 4, 16, and 32 Ps, above round 3's except
the 4-P maximum. With 100 work steps, `sync.Mutex` has a much lower p99 at 4 Ps
(median 21 µs against 115 µs) and a higher p99.9 (0.79–2.8 ms). The original
channel implementation, measured on this host in round 2, had much tighter
tails (sample maxima of 38–187 µs with an empty critical section); round 3 was
not latency-tested against it. [Comparison](latency/comparison.txt),
[6e6bc42 samples](latency/base.txt), [round 3 samples](latency/new.txt).

## Context-aware acquisition

Uncontended acquisition with a context against the plain call, measured in the
same ten rounds (median ns/op for a lock/unlock pair, difference in
parentheses). The contended rows use `RunParallel` with an empty critical
section.

| Call | Ps | Without context | `Background` | Cancelable | Allocated per op |
| --- | ---: | ---: | ---: | ---: | --- |
| TryMutex.LockContext | 1 | 3.63 | 3.77 (+0.14) | 4.02 (+0.39) | none |
| TryMutex.LockContext | 32 | 3.61 | 3.76 (+0.16) | 4.07 (+0.46) | none |
| Semaphore(8).AcquireContext | 1 | 3.74 | 3.78 (+0.04) | 4.28 (+0.53) | none |
| Semaphore(8).AcquireContext | 32 | 3.76 | 3.83 (+0.07) | 4.33 (+0.56) | none |
| NamedMutex.LockContext | 1 | 6.24 | 7.77 (+1.53) | 8.42 (+2.18) | none |
| NamedMutex.LockContext | 32 | 6.27 | 7.82 (+1.55) | 8.43 (+2.16) | none |
| TryMutex.LockContext, contended | 16 | 3.85 | 4.27 (+0.42) | 4.61 (+0.76) | none |
| TryMutex.LockContext, contended | 32 | 3.99 | 4.33 (+0.34) | 5.06 (+1.07) | none |

The context is checked before the first attempt, so an already-canceled
context fails even when the lock is free; after that it is consulted only
while waiting. These benchmarks measure the fast path: none parks with a
cancelable context, so the cost of registering it with `context.AfterFunc` on
each wait was not measured. [All GOMAXPROCS values](table-context.md),
[context samples](context.txt), [plain samples](context-plain.txt).

## Hardware counters

User-space counts per operation at GOMAXPROCS 1, 6e6bc42 → round 3
([counters.tsv](counters.tsv)).

| Benchmark | Instructions/op | Cycles/op | Branch misses/op |
| --- | ---: | ---: | ---: |
| TryMutex/Uncontended | 26.00 → 26.00 | 17.58 → 17.59 | 0.00 → 0.00 |
| TryMutex/TimeoutExpired | 576.0 → 444.0 | 329.9 → 302.6 | 0.00 → 0.00 |
| Semaphore/8/Uncontended | 98.00 → 36.00 | 43.19 → 18.12 | 0.00 → 0.00 |
| Semaphore/Value | 37.00 → 11.00 | 17.46 → 2.24 | 0.00 → 0.00 |
| Semaphore/Full8TryFailure | 49.00 → 15.00 | 19.55 → 3.00 | 0.00 → 0.00 |
| NamedMutex/HotKey | 155.0 → 122.0 | 42.01 → 30.01 | 0.00 → 0.00 |
| NamedMutex/ManyKeys | 616.5 → 611.0 | 210.5 → 162.3 | 0.28 → 0.09 |
| NamedMutex/UniqueNames | 2544.7 → 1160.0 | 2056.3 → 298.0 | 1.86 → 0.09 |
| NamedMutex/Create1024 | 1,810,694 → 872,403 | 618,734 → 230,956 | 2223.20 → 220.21 |
| OnceMutex/Completed | 24.00 → 10.00 | 8.00 → 2.23 | 0.00 → 0.00 |
| NamedOnceMutex/Uncontended | 455.0 → 301.0 | 130.2 → 84.21 | 0.00 → 0.00 |
| Construction/TryMutex | 459.4 → 280.7 | 126.4 → 74.06 | 0.22 → 0.06 |
| ControlWaitGroup/1 | 3044.7 → 3220.7 | 1247.1 → 1365.8 | 5.62 → 6.57 |
| ControlWaitGroup/256 | 3059.2 → 3237.3 | 1250.8 → 1373.0 | 5.62 → 6.60 |
| SyncFlag/Write | 62.00 → 62.00 | 48.04 → 48.03 | 0.00 → 0.00 |

Most single-goroutine gains come with fewer instructions: `Semaphore(8)` drops
from 98 to 36 instructions and `Value` from 37 to 11, and a `TryLockTimeout`
that expires executes 23% fewer. Cycles often fall further than instructions
(`Value` 7.8× fewer cycles for 3.4× fewer instructions), and
`NamedMutex/ManyKeys` runs about the same number of instructions in 23% fewer
cycles, with fewer branch misses. Uncontended `TryMutex` and `SyncFlag` are
unchanged, as their code paths are. Counters at one P say nothing about the
contended gains.

## Against the channel implementation

The original channel-based implementation (`nsync_baseline` build tag)
against round 3, in the same ten-round layout. Round 3 is significantly faster
in 164 of 184 rows and slower in 4: the first acquisition of a `OnceMutex` at
1 and 16 Ps (+2.6% and +5.0%) and two unchanged-code controls (+0.2–0.3%). The
geometric mean of ns/op fell 78%. Ratios in parentheses are not significant.

| Workload | 1 P | 4 Ps | 16 Ps | 32 Ps |
| --- | ---: | ---: | ---: | ---: |
| TryMutex Lock/Unlock, uncontended | 6.07× | 6.09× | 6.08× | 6.07× |
| TryMutex, empty critical section | 6.08× | 29.90× | 28.89× | 27.71× |
| TryMutex, 100 work steps | 1.16× | 2.26× | 2.17× | 2.19× |
| Semaphore(8) parallel | 5.99× | 7.99× | 59.16× | 59.66× |
| NamedMutex, one hot name | 6.39× | 6.55× | 6.53× | 6.54× |
| NamedMutex, a name per goroutine | 6.57× | 19.09× | 69.26× | 78.05× |
| NamedOnceMutex, same key | 4.71× | 5.03× | 5.30× | 6.73× |
| ControlWaitGroup(1) | 1.24× | 1.27× | 1.27× | 1.24× |
| ControlWaitGroup(256) | 1.25× | 1.83× | 3.03× | 2.61× |

[All medians](table-cumulative.md), [benchstat](cumulative/comparison.txt).

## Optional spinning

The `nsync_spin` build polls briefly before parking. Against the default build
in the same ten-round layout, it does not help:

| Workload | Ps | Default ns/op | `nsync_spin` ns/op | Default/spin |
| --- | ---: | ---: | ---: | ---: |
| TryMutex/Contended0 | 4 | 3.75 | 5.40 | 0.69× |
| TryMutex/Contended0 | 16 | 3.86 | 5.20 | 0.74× |
| TryMutex/Contended0 | 32 | 3.96 | 5.20 | 0.76× |
| TryMutex/Contended100 | 4 | 84.84 | 87.00 | 0.98× |
| TryMutex/Contended100 | 16 | 87.76 | 92.62 | 0.95× |
| TryMutex/Contended100 | 32 | 90.19 | 102.1 | 0.88× |

Every workload at 1 P, where it never spins, is within 0.4%, and uncontended
operations, which never reach the spin loop, within 1.3% (significant only at
32 Ps, 0.6%). Without spinning, contended `TryMutex` with an empty critical
section already runs close to its uncontended speed (3.75–3.96 ns against
3.59–3.65 ns), so polling has little to gain and probably only adds
cache-line traffic.
[Every row](table-spin.md), [benchstat](spin-comparison.txt).

## 32-bit x86

`GOARCH=386` builds of both revisions, run on the same host (32-bit code on
the 64-bit kernel), in six interleaved rounds ([medians](386/table-full.md),
[benchstat](386/comparison.txt)). 32-bit ARM and MIPS were not measured. Of 184
rows, round 3 is significantly faster in 114, slower in 43, and unchanged in
27; the geometric mean of ns/op fell 22%. Ratios in parentheses are not
significant.

| Workload | 1 P | 4 Ps | 16 Ps | 32 Ps |
| --- | ---: | ---: | ---: | ---: |
| TryMutex Lock/Unlock, uncontended | 0.95× | 0.95× | 0.95× | 0.96× |
| TryMutex, empty critical section | (0.99×) | 1.04× | 1.12× | 1.10× |
| TryLockTimeout on a free lock | 0.97× | 0.95× | 0.95× | 0.94× |
| Semaphore(1) uncontended | 1.09× | 1.08× | 1.08× | 1.11× |
| Semaphore(8) parallel | 2.02× | 1.95× | 3.29× | 4.39× |
| NamedMutex, one hot name | 1.29× | 1.27× | 1.27× | 1.27× |
| NamedMutex, a new name each time | 3.70× | 2.89× | 3.06× | 3.06× |
| OnceMutex, completed | 1.04× | (1.05×) | 1.04× | 1.04× |
| NamedOnceMutex, same key | 1.95× | 2.15× | 2.63× | 2.39× |
| ControlWaitGroup(1) | 0.95× | 0.93× | 0.92× | 0.95× |
| ControlWaitGroup(256) | 0.96× | 1.12× | 1.40× | 1.16× |
| ControlWaitGroup.Do after Abort | 0.87× | 0.87× | 0.87× | 0.88× |

On this platform the unchanged controls move by several percent in both
directions: `SyncFlag` reads are about 10% faster, uncontended `sync.Mutex` is
2–3% slower, and contended `sync.Mutex` at 1 P is 5% slower, with identical
code. Differences of a few percent here therefore do not all come from the
code. Beyond that range, a failed `TryLock` in parallel (+10–16%), completed
`OnceMutex` in parallel at 1 and 4 Ps (+16%), and `Do` after `Abort` (+14–15%)
are slower. Uncontended `TryMutex`, `TryLock`, and `TryLockTimeout` (+4–6%), a
failed `TryLock` (+7–8%), `Semaphore.TryAcquireTimeout` on a free semaphore
(+4–9%), `ControlWaitGroup` with every completion waking the submitter (+5–9%),
a first `OnceMutex` acquisition at 4 Ps, and completed `OnceMutex` in parallel
at 32 Ps (+2%) are slower by amounts within or near that range.

The first version of round 3 kept a 64-bit lock word on every platform. On
32-bit x86, Go calls a function for every atomic operation, but the 64-bit ones
also check alignment and use `CMPXCHG8B`, MMX moves for loads, and a
compare-and-swap loop for addition. That version made uncontended `TryMutex`
0.77× and `Semaphore(1)` 0.74× as fast as round 2
([benchstat](386-64bit-word/comparison.txt)). On 32-bit MIPS, Go implements
64-bit atomic operations with a single lock shared by the whole process. The
lock word is now pointer-sized, so 32-bit platforms use 32-bit atomic
operations for it, as round 2 did.

## Files

- `base.txt`, `new.txt`, `comparison.txt`, `table-full.md`: throughput and
  allocations, 6e6bc42 against round 3.
- `latency/`, `table-latency.md`: acquisition latency.
- `context.txt`, `context-plain.txt`, `table-context.md`: context-aware
  acquisition.
- `counters.tsv`, `table-counters.md`: hardware counters.
- `cumulative/`, `table-cumulative.md`: the channel implementation against
  round 3.
- `spin-default.txt`, `spin.txt`, `spin-comparison.txt`, `table-spin.md`:
  `nsync_spin` against the default build.
- `386/`: 32-bit x86, with `table-full.md`; `386-64bit-word/`: the same
  comparison with the earlier 64-bit lock word.
- `alignment-check/`: two builds of `ControlWaitGroup` whose `Do` after
  `Abort` loops differ only in placement.
- `final-code-check/`: the measured round 3 build against the final code, ten
  interleaved rounds.
- `validation.txt`: tests, vet, cross-builds, and soak runs of the final code.
- `environment.txt`, `run.log`, `load.txt`: host, settings, and load. The run
  wrote to `round3-amd64-final`, renamed afterwards.
