# All native AMD64 benchmark medians

AMD Ryzen 9 5950X, Linux, Go 1.27.1, GOAMD64=v1. Six samples per case.
See [the report](README.md) for methodology, limitations, and interpretation,
and [benchstat](comparison.txt) for confidence intervals and significance.

The number after the final dash is GOMAXPROCS; no suffix means 1.
Parallel ns/op is aggregate throughput. Ratios are baseline / optimized;
values above 1 mean higher measured throughput. These ratios alone do not establish significance.

## Throughput and allocations

All 180 main cases. Times are ns/op. Bytes and allocation events are per operation.
Each allocation cell is baseline → optimized.

| Workload | Baseline ns/op | Optimized ns/op | Ratio | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| TryMutex/Uncontended | 21.485 | 3.662 | 5.87× | 0 → 0 | 0 → 0 |
| TryMutex/Uncontended-4 | 22.67 | 3.958 | 5.73× | 0 → 0 | 0 → 0 |
| TryMutex/Uncontended-16 | 22.67 | 3.909 | 5.80× | 0 → 0 | 0 → 0 |
| TryMutex/Uncontended-32 | 23.21 | 3.829 | 6.06× | 0 → 0 | 0 → 0 |
| TryMutex/TrySuccess | 23.94 | 4.038 | 5.93× | 0 → 0 | 0 → 0 |
| TryMutex/TrySuccess-4 | 24.44 | 4.072 | 6.00× | 0 → 0 | 0 → 0 |
| TryMutex/TrySuccess-16 | 25.98 | 4.261 | 6.10× | 0 → 0 | 0 → 0 |
| TryMutex/TrySuccess-32 | 24.935 | 4.274 | 5.83× | 0 → 0 | 0 → 0 |
| TryMutex/TryFailure | 3.091 | 0.649 | 4.76× | 0 → 0 | 0 → 0 |
| TryMutex/TryFailure-4 | 3.164 | 0.669 | 4.73× | 0 → 0 | 0 → 0 |
| TryMutex/TryFailure-16 | 3.164 | 0.691 | 4.58× | 0 → 0 | 0 → 0 |
| TryMutex/TryFailure-32 | 3.224 | 0.693 | 4.65× | 0 → 0 | 0 → 0 |
| TryMutex/TimeoutSuccess | 25.1 | 3.608 | 6.96× | 0 → 0 | 0 → 0 |
| TryMutex/TimeoutSuccess-4 | 26.1 | 3.819 | 6.83× | 0 → 0 | 0 → 0 |
| TryMutex/TimeoutSuccess-16 | 25.98 | 3.852 | 6.75× | 0 → 0 | 0 → 0 |
| TryMutex/TimeoutSuccess-32 | 25.85 | 3.912 | 6.61× | 0 → 0 | 0 → 0 |
| TryMutex/TimeoutExpired | 324.15 | 69.315 | 4.68× | 248 → 0 | 3 → 0 |
| TryMutex/TimeoutExpired-4 | 404.1 | 75.01 | 5.39× | 248 → 0 | 3 → 0 |
| TryMutex/TimeoutExpired-16 | 500.1 | 75.85 | 6.59× | 248 → 0 | 3 → 0 |
| TryMutex/TimeoutExpired-32 | 518.25 | 75.375 | 6.88× | 248 → 0 | 3 → 0 |
| TryMutex/TimeoutPark | 1,052,727.5 | 1,052,089.5 | 1.00× | 248 → 112 | 3 → 1 |
| TryMutex/TimeoutPark-4 | 697,258.5 | 714,870.5 | 0.98× | 248 → 113.5 | 3 → 1 |
| TryMutex/TimeoutPark-16 | 689,974 | 699,832 | 0.99× | 248 → 121.5 | 3 → 1 |
| TryMutex/TimeoutPark-32 | 676,327.5 | 669,876.5 | 1.01× | 248 → 130 | 3 → 1 |
| TryMutex/Contended0 | 23.45 | 3.719 | 6.31× | 0 → 0 | 0 → 0 |
| TryMutex/Contended0-4 | 216.1 | 8.838 | 24.45× | 0 → 0 | 0 → 0 |
| TryMutex/Contended0-16 | 155.85 | 9.88 | 15.77× | 0 → 0 | 0 → 0 |
| TryMutex/Contended0-32 | 138.4 | 10.345 | 13.38× | 0 → 0 | 0 → 0 |
| TryMutex/Contended100 | 99.02 | 85.42 | 1.16× | 0 → 0 | 0 → 0 |
| TryMutex/Contended100-4 | 341.2 | 135.9 | 2.51× | 0 → 0 | 0 → 0 |
| TryMutex/Contended100-16 | 235.4 | 115.65 | 2.04× | 0 → 0 | 0 → 0 |
| TryMutex/Contended100-32 | 256 | 133.45 | 1.92× | 0 → 0 | 0 → 0 |
| Semaphore/1/Uncontended | 22.45 | 3.841 | 5.85× | 0 → 0 | 0 → 0 |
| Semaphore/1/Uncontended-4 | 23.4 | 3.839 | 6.10× | 0 → 0 | 0 → 0 |
| Semaphore/1/Uncontended-16 | 23.535 | 3.996 | 5.89× | 0 → 0 | 0 → 0 |
| Semaphore/1/Uncontended-32 | 23.62 | 3.944 | 5.99× | 0 → 0 | 0 → 0 |
| Semaphore/1/Parallel | 23.545 | 3.768 | 6.25× | 0 → 0 | 0 → 0 |
| Semaphore/1/Parallel-4 | 176.2 | 8.185 | 21.53× | 0 → 0 | 0 → 0 |
| Semaphore/1/Parallel-16 | 166.6 | 9.763 | 17.06× | 0 → 0 | 0 → 0 |
| Semaphore/1/Parallel-32 | 135.25 | 10.705 | 12.63× | 0 → 0 | 0 → 0 |
| Semaphore/8/Uncontended | 23.735 | 9.369 | 2.53× | 0 → 0 | 0 → 0 |
| Semaphore/8/Uncontended-4 | 24.93 | 9.525 | 2.62× | 0 → 0 | 0 → 0 |
| Semaphore/8/Uncontended-16 | 24.68 | 9.902 | 2.49× | 0 → 0 | 0 → 0 |
| Semaphore/8/Uncontended-32 | 24.175 | 9.881 | 2.45× | 0 → 0 | 0 → 0 |
| Semaphore/8/Parallel | 24.11 | 9.014 | 2.67× | 0 → 0 | 0 → 0 |
| Semaphore/8/Parallel-4 | 45.555 | 21.665 | 2.10× | 0 → 0 | 0 → 0 |
| Semaphore/8/Parallel-16 | 493.65 | 29.185 | 16.91× | 0 → 0 | 0 → 0 |
| Semaphore/8/Parallel-32 | 677.85 | 40.365 | 16.79× | 0 → 0 | 0 → 0 |
| Semaphore/64/Uncontended | 24.03 | 9.364 | 2.57× | 0 → 0 | 0 → 0 |
| Semaphore/64/Uncontended-4 | 24.415 | 9.751 | 2.50× | 0 → 0 | 0 → 0 |
| Semaphore/64/Uncontended-16 | 24.53 | 9.773 | 2.51× | 0 → 0 | 0 → 0 |
| Semaphore/64/Uncontended-32 | 24.215 | 9.882 | 2.45× | 0 → 0 | 0 → 0 |
| Semaphore/64/Parallel | 23.66 | 9.431 | 2.51× | 0 → 0 | 0 → 0 |
| Semaphore/64/Parallel-4 | 44.955 | 21.095 | 2.13× | 0 → 0 | 0 → 0 |
| Semaphore/64/Parallel-16 | 87.67 | 23.34 | 3.76× | 0 → 0 | 0 → 0 |
| Semaphore/64/Parallel-32 | 93.425 | 25.595 | 3.65× | 0 → 0 | 0 → 0 |
| Semaphore/TryFailure | 3.345 | 1.546 | 2.16× | 0 → 0 | 0 → 0 |
| Semaphore/TryFailure-4 | 3.311 | 1.594 | 2.08× | 0 → 0 | 0 → 0 |
| Semaphore/TryFailure-16 | 3.474 | 1.582 | 2.20× | 0 → 0 | 0 → 0 |
| Semaphore/TryFailure-32 | 3.458 | 1.603 | 2.16× | 0 → 0 | 0 → 0 |
| Semaphore/TimeoutSuccess | 25.565 | 4.285 | 5.97× | 0 → 0 | 0 → 0 |
| Semaphore/TimeoutSuccess-4 | 25.025 | 4.653 | 5.38× | 0 → 0 | 0 → 0 |
| Semaphore/TimeoutSuccess-16 | 26.835 | 4.515 | 5.94× | 0 → 0 | 0 → 0 |
| Semaphore/TimeoutSuccess-32 | 26.29 | 4.611 | 5.70× | 0 → 0 | 0 → 0 |
| Semaphore/Value | 1.337 | 3.783 | 0.35× | 0 → 0 | 0 → 0 |
| Semaphore/Value-4 | 1.387 | 3.822 | 0.36× | 0 → 0 | 0 → 0 |
| Semaphore/Value-16 | 1.373 | 3.95 | 0.35× | 0 → 0 | 0 → 0 |
| Semaphore/Value-32 | 1.393 | 3.957 | 0.35× | 0 → 0 | 0 → 0 |
| Semaphore/Full8TryFailure | 2.897 | 4.299 | 0.67× | 0 → 0 | 0 → 0 |
| Semaphore/Full8TryFailure-4 | 2.896 | 4.454 | 0.65× | 0 → 0 | 0 → 0 |
| Semaphore/Full8TryFailure-16 | 2.962 | 4.443 | 0.67× | 0 → 0 | 0 → 0 |
| Semaphore/Full8TryFailure-32 | 2.909 | 4.468 | 0.65× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey32 | 43.08 | 9.096 | 4.74× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey32-4 | 45.095 | 9.585 | 4.70× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey32-16 | 45.54 | 9.38 | 4.86× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey32-32 | 44.54 | 9.23 | 4.83× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey128 | 41.87 | 11.64 | 3.60× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey128-4 | 45.535 | 11.82 | 3.85× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey128-16 | 44.45 | 11.615 | 3.83× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey128-32 | 44.955 | 12.05 | 3.73× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey1024 | 42.135 | 11.69 | 3.60× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey1024-4 | 45.09 | 11.8 | 3.82× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey1024-16 | 45.175 | 11.855 | 3.81× | 0 → 0 | 0 → 0 |
| NamedMutex/LongKey1024-32 | 45.96 | 12.15 | 3.78× | 0 → 0 | 0 → 0 |
| NamedMutex/HotKey | 42.86 | 9.229 | 4.64× | 0 → 0 | 0 → 0 |
| NamedMutex/HotKey-4 | 43.995 | 9.495 | 4.63× | 0 → 0 | 0 → 0 |
| NamedMutex/HotKey-16 | 44.775 | 9.562 | 4.68× | 0 → 0 | 0 → 0 |
| NamedMutex/HotKey-32 | 45.635 | 9.527 | 4.79× | 0 → 0 | 0 → 0 |
| NamedMutex/SameKeyParallel | 42.34 | 9.048 | 4.68× | 0 → 0 | 0 → 0 |
| NamedMutex/SameKeyParallel-4 | 249.3 | 19.37 | 12.87× | 0 → 0 | 0 → 0 |
| NamedMutex/SameKeyParallel-16 | 175.6 | 21.98 | 7.99× | 0 → 0 | 0 → 0 |
| NamedMutex/SameKeyParallel-32 | 166.8 | 23.815 | 7.00× | 0 → 0 | 0 → 0 |
| NamedMutex/IndependentKeys | 43.04 | 8.85 | 4.86× | 0 → 0 | 0 → 0 |
| NamedMutex/IndependentKeys-4 | 120.45 | 4.617 | 26.09× | 0 → 0 | 0 → 0 |
| NamedMutex/IndependentKeys-16 | 264.35 | 2.425 | 108.99× | 0 → 0 | 0 → 0 |
| NamedMutex/IndependentKeys-32 | 309.15 | 2.385 | 129.65× | 0 → 0 | 0 → 0 |
| NamedMutex/ManyKeys | 49.47 | 46.535 | 1.06× | 0 → 0 | 0 → 0 |
| NamedMutex/ManyKeys-4 | 52.485 | 46.685 | 1.12× | 0 → 0 | 0 → 0 |
| NamedMutex/ManyKeys-16 | 55.12 | 48.2 | 1.14× | 0 → 0 | 0 → 0 |
| NamedMutex/ManyKeys-32 | 54.465 | 50.485 | 1.08× | 0 → 0 | 0 → 0 |
| NamedMutex/Create1024 | 169,164.5 | 146,703.5 | 1.15× | 223,720 → 166,616 | 1,047 → 1,463 |
| NamedMutex/Create1024-4 | 199,946 | 188,960 | 1.06× | 223,720 → 166,616 | 1,047 → 1,463 |
| NamedMutex/Create1024-16 | 275,104.5 | 235,578.5 | 1.17× | 223,720 → 166,616 | 1,047 → 1,463 |
| NamedMutex/Create1024-32 | 275,413.5 | 249,743 | 1.10× | 223,720 → 166,616 | 1,047 → 1,463 |
| OnceMutex/First | 15.455 | 16.15 | 0.96× | 16 → 16 | 1 → 1 |
| OnceMutex/First-4 | 16.59 | 18.67 | 0.89× | 16 → 16 | 1 → 1 |
| OnceMutex/First-16 | 21.56 | 26.22 | 0.82× | 16 → 16 | 1 → 1 |
| OnceMutex/First-32 | 24.255 | 27.37 | 0.89× | 16 → 16 | 1 → 1 |
| OnceMutex/Completed | 3.661 | 1.758 | 2.08× | 0 → 0 | 0 → 0 |
| OnceMutex/Completed-4 | 3.873 | 1.837 | 2.11× | 0 → 0 | 0 → 0 |
| OnceMutex/Completed-16 | 3.79 | 1.815 | 2.09× | 0 → 0 | 0 → 0 |
| OnceMutex/Completed-32 | 3.841 | 1.821 | 2.11× | 0 → 0 | 0 → 0 |
| OnceMutex/CompletedParallel | 3.745 | 1.458 | 2.57× | 0 → 0 | 0 → 0 |
| OnceMutex/CompletedParallel-4 | 23.565 | 0.393 | 59.92× | 0 → 0 | 0 → 0 |
| OnceMutex/CompletedParallel-16 | 58.545 | 0.119 | 490.53× | 0 → 0 | 0 → 0 |
| OnceMutex/CompletedParallel-32 | 110.95 | 0.106 | 1047.69× | 0 → 0 | 0 → 0 |
| NamedOnceMutex/Uncontended | 88.515 | 28.235 | 3.13× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/Uncontended-4 | 101.965 | 28.565 | 3.57× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/Uncontended-16 | 111.45 | 28.41 | 3.92× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/Uncontended-32 | 107.95 | 29.96 | 3.60× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/SameKeyParallel | 91.775 | 29.19 | 3.14× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/SameKeyParallel-4 | 202.15 | 72.965 | 2.77× | 15 → 0 | 0 → 0 |
| NamedOnceMutex/SameKeyParallel-16 | 373.85 | 117.1 | 3.19× | 13 → 0 | 0 → 0 |
| NamedOnceMutex/SameKeyParallel-32 | 557.3 | 141.55 | 3.94× | 10 → 0 | 0 → 0 |
| NamedOnceMutex/IndependentKeys | 88.215 | 26.26 | 3.36× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/IndependentKeys-4 | 196.3 | 7.094 | 27.67× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/IndependentKeys-16 | 398.8 | 2.409 | 165.55× | 16 → 0 | 1 → 0 |
| NamedOnceMutex/IndependentKeys-32 | 434.45 | 2.237 | 194.25× | 16 → 0 | 1 → 0 |
| TryFailureParallel | 2.828 | 0.643 | 4.40× | 0 → 0 | 0 → 0 |
| TryFailureParallel-4 | 0.804 | 0.168 | 4.79× | 0 → 0 | 0 → 0 |
| TryFailureParallel-16 | 0.236 | 0.053 | 4.42× | 0 → 0 | 0 → 0 |
| TryFailureParallel-32 | 0.215 | 0.05 | 4.32× | 0 → 0 | 0 → 0 |
| Construction/TryMutex | 45.025 | 34.23 | 1.32× | 120 → 96 | 2 → 1 |
| Construction/TryMutex-4 | 73.075 | 53.98 | 1.35× | 120 → 96 | 2 → 1 |
| Construction/TryMutex-16 | 81.54 | 63.065 | 1.29× | 120 → 96 | 2 → 1 |
| Construction/TryMutex-32 | 85.62 | 62.295 | 1.37× | 120 → 96 | 2 → 1 |
| Construction/Semaphore | 46.23 | 34.605 | 1.34× | 120 → 112 | 2 → 1 |
| Construction/Semaphore-4 | 73.88 | 67.18 | 1.10× | 120 → 112 | 2 → 1 |
| Construction/Semaphore-16 | 82.65 | 67.925 | 1.22× | 120 → 112 | 2 → 1 |
| Construction/Semaphore-32 | 83.93 | 68.495 | 1.23× | 120 → 112 | 2 → 1 |
| Construction/NamedMutexFirstKey | 181.35 | 74.145 | 2.45× | 384 → 208 | 4 → 2 |
| Construction/NamedMutexFirstKey-4 | 260.25 | 103.7 | 2.51× | 384 → 208 | 4 → 2 |
| Construction/NamedMutexFirstKey-16 | 332.05 | 136.2 | 2.44× | 384 → 208 | 4 → 2 |
| Construction/NamedMutexFirstKey-32 | 331.5 | 138.85 | 2.39× | 384 → 208 | 4 → 2 |
| Construction/NamedOnceFirstKey | 198.25 | 103.8 | 1.91× | 288 → 216 | 4 → 2 |
| Construction/NamedOnceFirstKey-4 | 243.9 | 142.2 | 1.72× | 288 → 216 | 4 → 2 |
| Construction/NamedOnceFirstKey-16 | 299.5 | 168.7 | 1.78× | 288 → 216 | 4 → 2 |
| Construction/NamedOnceFirstKey-32 | 330.15 | 156.65 | 2.11× | 288 → 216 | 4 → 2 |
| ControlWaitGroup/1 | 392.85 | 288.1 | 1.36× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/1-4 | 660.45 | 480.65 | 1.37× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/1-16 | 786.65 | 557 | 1.41× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/1-32 | 850.2 | 551.85 | 1.54× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/16 | 397.1 | 274.45 | 1.45× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/16-4 | 753.5 | 357.3 | 2.11× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/16-16 | 807.7 | 425.85 | 1.90× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/16-32 | 797.75 | 354.3 | 2.25× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/256 | 374.75 | 272.6 | 1.37× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/256-4 | 514.55 | 381.95 | 1.35× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/256-16 | 539.35 | 601.05 | 0.90× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/256-32 | 569.35 | 583.6 | 0.98× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/Aborted | 3.751 | 1.535 | 2.44× | 0 → 0 | 0 → 0 |
| ControlWaitGroup/Aborted-4 | 3.92 | 1.598 | 2.45× | 0 → 0 | 0 → 0 |
| ControlWaitGroup/Aborted-16 | 3.862 | 1.583 | 2.44× | 0 → 0 | 0 → 0 |
| ControlWaitGroup/Aborted-32 | 3.93 | 1.609 | 2.44× | 0 → 0 | 0 → 0 |
| SyncFlag/Read | 1.071 | 1.043 | 1.03× | 0 → 0 | 0 → 0 |
| SyncFlag/Read-4 | 1.075 | 1.032 | 1.04× | 0 → 0 | 0 → 0 |
| SyncFlag/Read-16 | 1.117 | 1.138 | 0.98× | 0 → 0 | 0 → 0 |
| SyncFlag/Read-32 | 1.13 | 1.118 | 1.01× | 0 → 0 | 0 → 0 |
| SyncFlag/Write | 10.055 | 10.16 | 0.99× | 0 → 0 | 0 → 0 |
| SyncFlag/Write-4 | 10.555 | 10.39 | 1.02× | 0 → 0 | 0 → 0 |
| SyncFlag/Write-16 | 10.695 | 10.64 | 1.01× | 0 → 0 | 0 → 0 |
| SyncFlag/Write-32 | 10.905 | 10.8 | 1.01× | 0 → 0 | 0 → 0 |
| StandardMutex/Uncontended | 3.679 | 3.674 | 1.00× | 0 → 0 | 0 → 0 |
| StandardMutex/Uncontended-4 | 3.7 | 3.81 | 0.97× | 0 → 0 | 0 → 0 |
| StandardMutex/Uncontended-16 | 3.758 | 3.821 | 0.98× | 0 → 0 | 0 → 0 |
| StandardMutex/Uncontended-32 | 3.822 | 3.873 | 0.99× | 0 → 0 | 0 → 0 |
| StandardMutex/Contended | 3.505 | 3.684 | 0.95× | 0 → 0 | 0 → 0 |
| StandardMutex/Contended-4 | 11.61 | 17.425 | 0.67× | 0 → 0 | 0 → 0 |
| StandardMutex/Contended-16 | 62.56 | 96.94 | 0.65× | 0 → 0 | 0 → 0 |
| StandardMutex/Contended-32 | 119.9 | 115.7 | 1.04× | 0 → 0 | 0 → 0 |

## Named-once completed work

One iteration can lead a new operation or join an existing operation.
A change in leaders/op changes the amount of completed useful work per iteration.

| Workload | Baseline leaders/op | Optimized leaders/op |
| --- | ---: | ---: |
| NamedOnceMutex/SameKeyParallel | 1 | 1 |
| NamedOnceMutex/SameKeyParallel-4 | 0.979 | 0.994 |
| NamedOnceMutex/SameKeyParallel-16 | 0.86 | 0.964 |
| NamedOnceMutex/SameKeyParallel-32 | 0.673 | 0.941 |

## Acquisition latency

Four goroutines per P, 2,000,000 iterations per run; one in 64 acquisitions sampled.
All times below are nanoseconds. Each cell is baseline → optimized.
Percentiles and sample maxima are medians of six per-run measurements, not worst-case bounds.
Clock sampling and interface dispatch make ns/op distinct from the main throughput suite.

| Workload | ns/op | p50 wait ns | p99 wait ns | Sample max ns | Samples/run |
| --- | ---: | ---: | ---: | ---: | ---: |
| WaitLatency/TryMutex/Work0-4 | 153.2 → 16.73 | 1,949 → 50 | 5,741 → 70 | 37,671 → 281,820 | 31,257.5 → 31,259.5 |
| WaitLatency/TryMutex/Work0-16 | 144.8 → 21.78 | 8,096 → 50 | 20,758.5 → 111,790 | 186,731 → 429,102 | 31,295.5 → 31,280.5 |
| WaitLatency/TryMutex/Work0-32 | 144.35 → 34.09 | 17,899 → 50 | 41,613 → 174,057 | 171,046.5 → 449,810.5 | 31,360.5 → 31,314 |
| WaitLatency/StandardMutex/Work0-4 | 67.69 → 81.565 | 130 → 120 | 50,119.5 → 38,102 | 426,531.5 → 409,630 | 31,258 → 31,257.5 |
| WaitLatency/StandardMutex/Work0-16 | 130.75 → 109.84 | 185 → 150 | 351,881.5 → 269,086 | 1,419,920 → 1,453,268.5 | 31,281 → 31,282.5 |
| WaitLatency/StandardMutex/Work0-32 | 136.5 → 137.45 | 190 → 190 | 757,269.5 → 763,225.5 | 3,939,859.5 → 3,775,921 | 31,312.5 → 31,312 |
| WaitLatency/TryMutex/Work100-4 | 254.25 → 160.75 | 3,376.5 → 40 | 10,579.5 → 125,555.5 | 58,254.5 → 316,334.5 | 31,257 → 31,258.5 |
| WaitLatency/TryMutex/Work100-16 | 231.35 → 190.8 | 12,979.5 → 40 | 29,775.5 → 145,904.5 | 69,561 → 378,266.5 | 31,296.5 → 31,282 |
| WaitLatency/TryMutex/Work100-32 | 244.4 → 198.85 | 28,548.5 → 40 | 65,383 → 193,739.5 | 153,739 → 460,516 | 31,360.5 → 31,314.5 |
| WaitLatency/StandardMutex/Work100-4 | 248.75 → 211.85 | 55 → 50 | 131,763 → 107,056 | 979,707.5 → 1,060,179 | 31,258.5 → 31,257.5 |
| WaitLatency/StandardMutex/Work100-16 | 225.25 → 218.45 | 60 → 60 | 947,702.5 → 1,124,801 | 6,628,230.5 → 7,682,508.5 | 31,283 → 31,284 |
| WaitLatency/StandardMutex/Work100-32 | 263.85 → 249.3 | 60 → 60 | 1,127,705.5 → 1,557,800 | 4,092,141 → 7,193,909.5 | 31,318.5 → 31,316.5 |

## Optional bounded spin

Default optimized implementation versus the same code with nsync_spin.
All rows report zero bytes and allocation events per operation.

| Workload | Default ns/op | Spin ns/op | Default / spin |
| --- | ---: | ---: | ---: |
| TryMutex/Uncontended | 3.662 | 3.716 | 0.99× |
| TryMutex/Uncontended-4 | 3.958 | 3.841 | 1.03× |
| TryMutex/Uncontended-16 | 3.909 | 3.846 | 1.02× |
| TryMutex/Uncontended-32 | 3.829 | 3.87 | 0.99× |
| TryMutex/Contended0 | 3.719 | 3.863 | 0.96× |
| TryMutex/Contended0-4 | 8.838 | 10.143 | 0.87× |
| TryMutex/Contended0-16 | 9.88 | 10.465 | 0.94× |
| TryMutex/Contended0-32 | 10.345 | 11.7 | 0.88× |
| TryMutex/Contended100 | 85.42 | 82.205 | 1.04× |
| TryMutex/Contended100-4 | 135.9 | 121.6 | 1.12× |
| TryMutex/Contended100-16 | 115.65 | 152.65 | 0.76× |
| TryMutex/Contended100-32 | 133.45 | 159 | 0.84× |

Spin acquisition latency, with the same diagnostics as above. Each cell is default → spin.

| Workload | ns/op | p50 wait ns | p99 wait ns | Sample max ns | Samples/run |
| --- | ---: | ---: | ---: | ---: | ---: |
| WaitLatency/TryMutex/Work0-4 | 16.73 → 12.215 | 50 → 35.5 | 70 → 120 | 281,820 → 230,077.5 | 31,259.5 → 31,171.5 |
| WaitLatency/TryMutex/Work0-16 | 21.78 → 18.805 | 50 → 40 | 111,790 → 536 | 429,102 → 329,515 | 31,280.5 → 31,285.5 |
| WaitLatency/TryMutex/Work0-32 | 34.09 → 67.075 | 50 → 40 | 174,057 → 180,434.5 | 449,810.5 → 374,209 | 31,314 → 31,320.5 |
| WaitLatency/TryMutex/Work100-4 | 160.75 → 190.75 | 40 → 50 | 125,555.5 → 126,047 | 316,334.5 → 318,559 | 31,258.5 → 31,258.5 |
| WaitLatency/TryMutex/Work100-16 | 190.8 → 246.75 | 40 → 40.5 | 145,904.5 → 186,085 | 378,266.5 → 441,731 | 31,282 → 31,282.5 |
| WaitLatency/TryMutex/Work100-32 | 198.85 → 235.85 | 40 → 30 | 193,739.5 → 192,041 | 460,516 → 394,742.5 | 31,314.5 → 31,312.5 |

## Isolated atomic assembly experiment

Matching CAS acquire/release pairs, one P. This is not the library lock backend.

| Workload | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| AtomicPair/Intrinsics | 3.59 | 0 | 0 |
| AtomicPair/Assembly | 3.931 | 0 | 0 |

## Longer alternating confirmation

Eight samples per case, 500 ms each, alternating build order between rounds.
These samples are separate from the main suite; see [benchstat](recheck-comparison.txt).

| Workload | Baseline ns/op | Optimized ns/op | Ratio | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| OnceMutex/First-4 | 21.24 | 23.8 | 0.89× | 16 → 16 | 1 → 1 |
| OnceMutex/First-16 | 23.27 | 24.85 | 0.94× | 16 → 16 | 1 → 1 |
| OnceMutex/First-32 | 22.945 | 23.45 | 0.98× | 16 → 16 | 1 → 1 |
| ControlWaitGroup/256-4 | 561.1 | 525.2 | 1.07× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/256-16 | 554.4 | 603.95 | 0.92× | 24 → 24 | 1 → 1 |
| ControlWaitGroup/256-32 | 560.05 | 568.8 | 0.98× | 24 → 24 | 1 → 1 |
| StandardMutex/Uncontended-4 | 3.648 | 3.698 | 0.99× | 0 → 0 | 0 → 0 |
| StandardMutex/Uncontended-16 | 3.656 | 3.592 | 1.02× | 0 → 0 | 0 → 0 |
| StandardMutex/Uncontended-32 | 3.632 | 3.74 | 0.97× | 0 → 0 | 0 → 0 |
| StandardMutex/Contended-4 | 17.01 | 17.89 | 0.95× | 0 → 0 | 0 → 0 |
| StandardMutex/Contended-16 | 71.14 | 63.21 | 1.13× | 0 → 0 | 0 → 0 |
| StandardMutex/Contended-32 | 113.45 | 113.25 | 1.00× | 0 → 0 | 0 → 0 |
