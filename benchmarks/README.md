# Performance and implementation notes

The library uses atomic fast paths and condition variables for parking. Its
production source contains no channels, runtime `linkname` calls, or unsafe
pointer conversions. `unsafe.Sizeof` is used only for cache-line padding. Public
method signatures are preserved, including value copies of `TryMutex` sharing
one underlying lock.

Two optimization rounds are recorded here. Round 2 replaced the original
channel-based implementation. Round 3, the current code, added context-aware
acquisition, made `NamedMutex` discard unlocked names, and replaced the lock
core. Each round is measured against the code it replaced.

## Round 3 results

Native AMD Ryzen 9 5950X, Linux, Go 1.27.1, against `6e6bc42` (round 2), with
the `performance` governor and no other benchmark or build running. Ten
interleaved rounds per build on the same benchmark bodies, with one process
per GOMAXPROCS value. The table shows ratios of round 2 to round 3 median
ns/op, so values above 1× mean round 3 is faster; ratios in parentheses are not
significant. Parallel rows measure aggregate throughput.

| Workload | 1 P | 4 Ps | 16 Ps | 32 Ps |
| --- | ---: | ---: | ---: | ---: |
| TryMutex Lock/Unlock, uncontended | (0.99×) | 1.00× | (1.00×) | (1.00×) |
| TryMutex, empty critical section | (1.00×) | 1.09× | 1.16× | 1.15× |
| TryMutex, 100 work steps | (1.00×) | 1.02× | 1.04× | (1.03×) |
| Semaphore(1) uncontended | 0.97× | 0.97× | 0.97× | 0.97× |
| Semaphore(8) uncontended | 2.39× | 2.38× | 2.37× | 2.38× |
| Semaphore(8) parallel | 2.38× | 2.03× | 3.30× | 3.95× |
| NamedMutex, one hot name | 1.36× | 1.40× | 1.40× | 1.39× |
| NamedMutex, 1,024 names in turn | 1.28× | 1.32× | 1.32× | 1.31× |
| NamedMutex, a new name each time | 5.49× | 5.14× | 5.14× | 4.96× |
| NamedOnceMutex, one key | 1.55× | 1.55× | 1.55× | 1.55× |
| NamedOnceMutex, same key | 1.52× | 1.60× | 2.06× | 2.05× |
| ControlWaitGroup(1) | 0.92× | 0.90× | 0.90× | 0.91× |
| ControlWaitGroup(256) | 0.91× | 1.17× | 2.66× | 1.40× |

The `ControlWaitGroup(256)` ratios at 16 and 32 Ps rest on unstable round 2
samples, and a new name each time is not a steady state for round 2, whose cost
grows with the names it retains. Some rows that take under a nanosecond depend
on code alignment: between the measured build and the final code, which differ
only in how they handle a misused `Release` and in tests, four of them moved by
27–36% and the others by under 4%.

Across all 184 rows, 130 are significantly faster and 22 slower; the geometric
mean of ns/op fell 36%, and over the 44 rows that allocate, the geometric mean
of allocated bytes fell 48%. `NewTryMutex` allocates 24 B instead of 96 B. A
`NamedMutex` that locked a new name each time used to retain every name; it now
retains at most 129 idle locks. Uncontended `TryMutex` `Lock` and `Unlock` are
unchanged at about 3.6 ns a pair: one compare-and-swap each way, as before.

The slower cases are `ControlWaitGroup` whenever each completion must wake the
submitter (one slot, or any pool size at GOMAXPROCS 1: 0.90–0.92×), uncontended
`Semaphore(1)` and its parallel benchmark at 1 P (0.97–0.98×), `TryLock` and
`TryLockTimeout` on a free lock (0.97–0.99×), and the tail latency of contended
`TryMutex`: with an empty critical section, p99.9 nearly doubles (1.9× and
2.0×) at 16 and 32 Ps and the longest sampled waits are 2–2.5× as long at every
GOMAXPROCS (while p99 at 32 Ps falls from 114 µs to 0.1 µs); with 100 work
steps, p99.9 rises 63% at 32 Ps and p99 rises 1–9%. The other four slower rows,
unchanged `SyncFlag` reads at 1, 16, and 32 Ps and uncontended `TryMutex` at 4
Ps, differ by under 0.7%.

Against the original channel-based implementation, round 3 lowers the geometric
mean of ns/op by 78%. Built as 32-bit x86 on the same host, round 3 lowers it
by 22%; there, some operations that take under 10 ns are 4–16% slower and
`ControlWaitGroup` with one slot, or any pool size at 1 P, is 5–9% slower,
while unchanged code moves by up to 11%. The [round 3
report](results/round3-amd64/README.md) has every benchmark, the regressions,
the acquisition-latency tradeoff, the cost of contexts, hardware counters, and
the 32-bit results.

To compare the working tree with any revision on the same host:

```sh
./benchmarks/compare-revisions.sh 6e6bc42 benchmarks/results/local 10
benchstat benchmarks/results/local/base.txt benchmarks/results/local/new.txt
```

The script starts one process per GOMAXPROCS value. Given a `-test.cpu` list,
Go 1.27 measures a `b.Loop` benchmark's first value at the GOMAXPROCS left by
the previous measurement, so a list would mislabel the 1-P rows.

## Round 3 design

Round 3 starts from commit `6e6bc42`, the round 2 implementation described
further below. It keeps every public method and its semantics, except that on
32-bit platforms a `Semaphore` counts at most 67,108,863 permits (the 64-bit
limit, 2^58-1, and the `ControlWaitGroup` pool limit cannot be reached in
practice). It adds
context-aware acquisition, makes `NamedMutex` discard unlocked names, and
replaces the lock core.

- **One lock core.** `TryMutex`, `NamedMutex` locks, and `Semaphore` share a
  two-word `lockState` (16 bytes on 64-bit platforms): one atomic word holding
  the number of held permits and the queue flags, plus a lazily allocated wait
  queue. A mutex is a limit of one. Round 2 used 80 bytes per lock. Starving
  and retired flags sit above the permit count, so `state < limit<<3` alone
  means a permit may be taken. The word is pointer-sized, as is the
  `ControlWaitGroup` state: on 32-bit platforms Go implements 64-bit atomic
  operations with costlier calls than 32-bit ones (`LOCK CMPXCHG8B` on x86),
  and on 32-bit MIPS with a lock shared by the whole process.
- **Fast paths.** Uncontended `Lock` and `Unlock` are one compare-and-swap
  each. `Semaphore.Acquire` is one load and compare-and-swap; `Release` is one
  wait-free fetch-add. `Value` and a failed `TryAcquire` are a single load.
- **Queue.** Waiters park in FIFO order, each on its own `sync.Cond`, so a
  release, timeout, or cancellation wakes exactly the waiter concerned, and
  `testing/synctest` sees parked goroutines as durably blocked. A release that
  finds the queued flag wakes one waiter and clears the flag; the woken waiter
  sets it again if waiters still need a wakeup. Running goroutines therefore
  keep using the fast paths and may take a released lock first. A woken waiter
  that loses keeps its place at the head; after 100 µs, releases hand permits
  to the head directly, as in round 2.
- **Cancellation.** Timeouts use a timer and contexts use `context.AfterFunc`
  (no channels). The callback removes only its own waiter; if it already
  started when the waiter leaves, the waiter waits for it before recycling.
  Each queue keeps one finished waiter for reuse, so parking normally avoids
  `sync.Pool`.
- **Semaphore gate.** When a semaphore with more than one permit is busy or
  contended, acquirers take turns through a second lock, whose holder may wait
  for a permit; a one-permit semaphore queues on its lock directly, like a
  mutex. Releases never take the gate. Without it, goroutines retrying one
  cache line collapsed throughput (see the rejected experiments).
- **NamedMutex.** A name's lock exists while it is held or awaited: entries
  count their holders and waiters under the shard lock and are removed by the
  last one. Each shard keeps in-use entries in a four-entry table keyed by the
  name hash that also selects the shard, with a map only for overflow, which is
  dropped when it empties after growing. Lock-free lookups use the lock of the
  first name and one cached lock per shard; the cached pointer and its name
  hash sit on their own cache line, so other names reject it by hash without
  reading the entry its holders write. A cached lock is replaced only when idle
  and its queue is empty, because a woken waiter may be about to retry while
  the word reads idle. At most 129 idle locks are retained per instance.
- **NamedOnceMutex.** String, `int`, `int64`, and `uint64` keys use
  type-specific hashing; if an instance's first key has one of these types, it
  is kept in typed fields and compared before hashing. Entries no other caller
  ever saw are reused without any atomic state change.
- **ControlWaitGroup.** Running tasks and flags share one word: with no
  submitter or `Wait` call queued, a completion and an admission into a free
  slot are a compare-and-swap; an admission that loses that race takes the
  queue lock. Blocked submissions are admitted in arrival order, and a
  finishing task hands its slot to the first one directly, so no newcomer can
  take it in between. The queued flag stays set after such a handoff until a
  completion finds the queue empty, so when the submitter queues again at
  once, two atomic updates are saved: the completion no longer clears the flag
  and the submitter no longer sets it.

### Correctness

New stress tests mix blocking, nonblocking, timed (including timeouts that
expire while queued), deadline-context, and concurrently canceled acquisitions
of `TryMutex`, `Semaphore`, and `NamedMutex` across 48 goroutines; the
`NamedOnceMutex` and `ControlWaitGroup` tests (48 and 32 goroutines) mix
blocking calls with deadline contexts. All run with the default handoff
threshold and with handoff after any lost race. They check exclusion or
capacity, that named-once waiters return only after the operation they joined
completed, that `ControlWaitGroup` never exceeds its limit, and that every lock
word returns exactly to its idle state (for `NamedOnceMutex`, that no completed
operation is retained). Deterministic tests, most using `synctest`, cover each
context method, cancellation storms, the bounded retention of `NamedMutex`,
the retirement race below, misused `Semaphore.Release` calls racing with each
other and with acquirers, and named locks whose hashes share their low 32
bits.

During development seven bugs were found. The first four were fixed before the
final measurements; the last three, which affect only a misused `Release`,
were fixed afterwards in slow paths that the benchmarks barely touch (an A/B
check is in the [round 3 report](results/round3-amd64/README.md)):

1. A semaphore release whose fetch-add ran before the queue started starving
   took the permit back for the queue head after a running goroutine had
   already taken it, exceeding the capacity.
2. Retiring a cached named lock whose word read idle while a woken waiter was
   about to retry stranded that waiter.
3. `ControlWaitGroup` released a slot and then took it back for a queued
   submitter; if the queue emptied and refilled in between, it exceeded the
   limit.
4. Found in review: a `Semaphore.Release` without an acquisition makes the
   word negative until it is undone, and acquirers took the negative word for
   a retired lock, so `Acquire` could return without a permit and
   `AcquireContext` could return nil. They now wait, and undoing the release
   wakes them.
5. Found in review of that fix: when misused releases overlapped, or an
   acquirer gave up while the word was negative, clearing the queue flags also
   cleared a bit of the wrapped count, so the word did not return to idle once
   the releases were undone. A negative word now keeps that bit, and only the
   last undo wakes waiters.
6. Found in review: a misused `Release` that landed while a semaphore was
   handing a permit to a starving waiter, after its holders' decrements had
   brought the count to zero, borrowed from the starving flag instead of making
   the word negative. It went undetected, and the waiter deadlocked. The
   negative flag now sits directly above the count, so any decrement below zero
   sets it, and a starving waiter no longer sets the starving flag in a
   negative word.
7. Found in review: undoing a misused `Release` read the word and then added to
   it, so a valid `Release` on its slow path could undo the same decrement. One
   misuse then caused two panics and left a permit that nobody held, and a
   waiter could stay parked. Each unit below zero is now undone once, with a
   compare-and-swap.

Final checks, on the final code: the shuffled suite passed 30 soak runs
(default, race detector, and `nsync_spin` builds at GOMAXPROCS 1, 2, 4, 8, and
32, twice each) and 10 more natively as 32-bit `GOARCH=386`, and it compiles
for all 18 cross-build targets in both modes
([validation log](results/round3-amd64/validation.txt)).

### Rejected round 3 experiments

Exploratory interleaved runs with four to six samples per build, on the same
host while it was shared and on the `powersave` governor, so absolute values
are higher than in the results above. They passed `-test.cpu` lists, so their
1-P rows of `b.Loop` benchmarks ran at the list's last GOMAXPROCS value. Ratios
are against `6e6bc42` unless stated otherwise; these files are not published.

| Candidate | Outcome |
| --- | --- |
| A queued flag that stays set while waiters are queued | Every release by a running goroutine took the slow path: contended `TryMutex` throughput halved (5.4 → 11.4 ns/op at 16 Ps). Replaced by the flag that the woken waiter re-arms. |
| A separate "woken" bit in the lock word | Same slow-path problem; the per-waiter flag under the queue lock is enough. |
| A lock-free semaphore count without the gate | Retries on one cache line: capacity 64 with 32 Ps went from 13.6 to 212 ns/op. |
| The gate around every busy semaphore operation | Helped the saturated case but cost the free-permit case (capacity 64, 16 Ps: 9.9 → 17.3 ns/op against the candidate without it). Only colliding or blocked acquirers use it now. |
| An inlined `CAS(0, 1 permit)` semaphore fast path | Fits the inlining budget but wastes a locked instruction whenever other permits are held: 0.72–0.86× at 4 Ps. A load followed by a compare-and-swap, not inlined, costs about 0.15 ns more uncontended and about 10% more at 16 and 32 Ps, but is more than twice as fast at 4 Ps. |
| Go maps for in-use named locks | Insertion and deletion on every cycle, with the name hashed twice: `ManyKeys` 76 ns/op (0.58×). The four-entry hashed table made it 33–36 ns/op (1.2–1.35×). |
| A last-byte check before comparing the first name | Cost 0.5 ns on the hot-name path. |
| A generic first-key hint in `NamedOnceMutex` (`atomic.Value` and an interface comparison) | As expensive as hashing an integer key: independent `int64` keys 0.72× at 16 Ps. Typed fields fixed it. |
| Handoff thresholds of 50 and 250 µs | Inconclusive for the deep tail: with six samples per build, these runs could not resolve tail changes of the size seen (50 µs measured a lower p99.9 with 100 work steps at 4 and 16 Ps, 124 vs 154 µs and 171 vs 208 µs, but cost about 16% throughput at 32 Ps). 250 µs doubled p99 with 100 work steps at 16 and 32 Ps, as p99 follows the threshold there. |
| Moving a woken waiter that lost to the back of the queue | Within noise of keeping its place. |
| Waking another waiter on every slow release | Cut the sample maximum of the empty critical section from about 1.05 ms to 0.6–0.75 ms at 16 and 32 Ps, but was 10–20% slower there than without the extra wakeups and brought back the 115 µs p99 at 32 Ps. |
| Yielding (`runtime.Gosched`) after a direct handoff | Up to 1.9× slower than without yielding, with no better tails. |
| A 64-bit lock word on every platform | On 32-bit platforms Go implements 64-bit atomic operations with costlier calls than 32-bit ones (`LOCK CMPXCHG8B` on x86), and on 32-bit MIPS with a lock shared by the whole process. On 32-bit x86, uncontended `TryMutex` ran 0.77× and `Semaphore(1)` 0.74× as fast as round 2 ([results](results/round3-amd64/386-64bit-word/comparison.txt)). The word is now pointer-sized. |
| The `nsync_spin` build, polling with `PAUSE` or `YIELD` before parking | Remeasured with the final code: 0.69–0.76× of the default build with an empty critical section and 0.88–0.98× with 100 work steps at 4–32 Ps, and within 0.6% at 1 P, where it does not spin. Contended `TryMutex` already runs close to its uncontended speed without it. It remains an opt-in experiment ([results](results/round3-amd64/spin-comparison.txt)). |

## Round 2 notes

The rest of this document describes round 2. Its baseline is the original
channel-based implementation, frozen in `internal/baseline`, and its design
section describes code that round 3 replaced.

### Reproduce the round 2 measurements

The baseline is the reviewed, Go 1.27.1-compatible implementation immediately
before optimization. It is frozen in
[`internal/baseline`](../internal/baseline).
[`baseline-sha256.json`](baseline-sha256.json) records the original source hashes:
remove the added first comment line and change `package baseline` back to
`package nsync` before comparing them. All seven hashes were verified.

The same concrete-type benchmark bodies select the baseline with the
`nsync_baseline` build tag. Setup and allocations are excluded where `b.Loop`
permits; constructor and cold-key benchmarks intentionally measure them.

```sh
./benchmarks/run.sh benchmarks/results/local
benchstat benchmarks/results/local/baseline.txt
benchmarks/results/local/optimized.txt
```

The script runs six samples of each workload at GOMAXPROCS 1, 4, and 16, with
150 ms per sample. Run it on an otherwise idle host. Do not overlap benchmark
processes, cross-compilation, or race tests. `RunParallel` ns/op is aggregate
throughput, not individual operation latency; a result below 1 ns does not mean
one acquisition completes in less than a nanosecond.

Development host: Apple M3 Max, 16 cores, macOS, Go 1.27.1, default
GOARM64=v8.0. ARM64 measurements are native. The default ARM64 binary selects
LSE atomics on supported CPUs and retains the compiler's portable fallback.
Darwin AMD64 tests on this host run under Rosetta and establish no native x86
performance claim.

The same source and frozen baseline were also measured natively on an AMD Ryzen
9 5950X, Linux, Go 1.27.1, GOAMD64=v1, at GOMAXPROCS 1, 4, 16, and 32. See the
[AMD64 comparison report](results/final-amd64/README.md) and
[all AMD64 benchmark medians](results/final-amd64/all-results.md), including
allocations, acquisition latency, handwritten assembly, and optional spinning.
To repeat that suite on native Linux AMD64:

```sh
./benchmarks/compare-linux-amd64.sh benchmarks/results/local-amd64
```

Final raw throughput/allocation results are in
[`results/final-arm64`](results/final-arm64). The other result files record
exploratory candidates, sometimes with shorter runs or different benchmark
subsets. They are not interchangeable with the final comparison. The unchanged
`sync.Mutex` controls also help expose host noise between runs. The checked-in
environment records the original repository HEAD; the optimized source is the
accompanying working-tree change.

### Measured round 2 ARM64 results

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
`NamedOnceMutex/SameKeyParallel` additionally reports `leaders/op`: one
iteration may lead an operation or join an existing one. Changes in that ratio
mean its ns/op is not a direct comparison of completed useful work.

Allocation and performance tradeoffs:

- Reading `Semaphore.Value()` increased from 1.87 to 2.60 ns/op. General counters are now protected by a mutex.
- Trying a full capacity-8 semaphore increased from 2.13 to 3.34 ns/op. General counters are now protected by a mutex.
- Creating 1,024 names used about 166.8 KiB versus 218.5 KiB, but increased allocation events from 1047 to 1478. Sharding trades more small allocations for less contention.
- Padding and embedded shard storage make empty named-once instances larger. The first-key construction benchmark measures the populated case; it is not a universal memory-saving claim.
- A parked 10 µs timeout has similar elapsed time, with allocations reduced from 3 to 1 and bytes from 248 to 112 per call. Tiny already-expired deadlines usually avoid a timer entirely.
- Repeated uncontended named-once operations went from 1 allocation per operation to zero after warm-up. TryMutex construction went from two allocations to one, and first-key construction of both named types went from four to two.
- `SyncFlag` was already close to the atomic instruction cost and remains essentially unchanged. Serial iteration over 1,024 named locks also changed little.

### Round 2 acquisition latency

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

Thresholds of 25, 100, 250, and 1,000 microseconds were measured. The 25
microsecond candidate substantially reduced long waits but cost throughput with
64 busy contenders. The 100 microsecond threshold balances these objectives. It
does not guarantee FIFO ordering or completion within 100 microseconds.

Median of six runs, 64 goroutines on 16 Ps; times are microseconds:

| Critical section | Baseline p99 | Optimized p99 | Baseline sample max | Optimized sample max |
| --- | ---: | ---: | ---: | ---: |
| Empty | 15.2 | 61.8 | 78.9 | 457.6 |
| 100 arithmetic steps | 36.9 | 198.4 | 74.4 | 417.5 |

See [`latency-comparison.txt`](results/latency-comparison.txt) for all latency
metrics. The baseline's more even handoff still has tighter tails in these
workloads. Choose based on the application's latency and throughput
requirements.

### Round 2 design

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

### Rejected or optional round 2 experiments

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

### Round 2 correctness and portability

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
recycling. `testing/synctest` makes deadline and queue tests deterministic. An
AST check rejects channels in production source, and layout tests check
cache-line padding invariants. Tests and the frozen historical baseline may use
channels.

[`cross-build.sh`](cross-build.sh) compiles default and `nsync_spin` library/test
binaries for 18 targets: Linux AMD64, ARM64, ARMv7, 386, RISC-V64, PPC64LE,
s390x, Loong64, MIPS64LE, and, since round 3, MIPS and MIPSLE; macOS
AMD64/ARM64; Windows AMD64/ARM64; FreeBSD AMD64/ARM64; and JS/Wasm. Compilation
is not execution. Cache-line sizes follow Go's architecture-specific CPU
definitions, including 128 bytes on ARM64 and 256 bytes on s390x.

The GitHub Actions matrix is configured for native Linux x86_64/ARM64, macOS
ARM64, and Windows x86_64 race tests, plus cross-compilation and, since round
3, tests built as 32-bit x86 on Linux. Those remote jobs have not been run by
these local changes.
