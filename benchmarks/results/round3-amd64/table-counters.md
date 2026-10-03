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
