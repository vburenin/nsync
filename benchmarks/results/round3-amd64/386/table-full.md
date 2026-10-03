| Workload | Ps | 6e6bc42 ns/op | New ns/op | Ratio | Allocated per op (6e6bc42 → new) |
| --- | ---: | ---: | ---: | ---: | --- |
| TryMutex/Uncontended | 1 | 6.099 | 6.441 | 0.95× | none |
| TryMutex/Uncontended | 4 | 6.141 | 6.471 | 0.95× | none |
| TryMutex/Uncontended | 16 | 6.173 | 6.485 | 0.95× | none |
| TryMutex/Uncontended | 32 | 6.205 | 6.492 | 0.96× | none |
| TryMutex/TrySuccess | 1 | 6.962 | 7.321 | 0.95× | none |
| TryMutex/TrySuccess | 4 | 6.889 | 7.275 | 0.95× | none |
| TryMutex/TrySuccess | 16 | 6.941 | 7.367 | 0.94× | none |
| TryMutex/TrySuccess | 32 | 6.970 | 7.293 | 0.96× | none |
| TryMutex/TryFailure | 1 | 2.891 | 3.099 | 0.93× | none |
| TryMutex/TryFailure | 4 | 2.905 | 3.109 | 0.93× | none |
| TryMutex/TryFailure | 16 | 2.901 | 3.128 | 0.93× | none |
| TryMutex/TryFailure | 32 | 2.917 | 3.120 | 0.93× | none |
| TryMutex/TimeoutSuccess | 1 | 6.684 | 6.921 | 0.97× | none |
| TryMutex/TimeoutSuccess | 4 | 6.409 | 6.780 | 0.95× | none |
| TryMutex/TimeoutSuccess | 16 | 6.440 | 6.793 | 0.95× | none |
| TryMutex/TimeoutSuccess | 32 | 6.430 | 6.804 | 0.94× | none |
| TryMutex/TimeoutExpired | 1 | 116.7 | 97.06 | 1.20× | none |
| TryMutex/TimeoutExpired | 4 | 116.8 | 96.97 | 1.20× | none |
| TryMutex/TimeoutExpired | 16 | 117.0 | 97.70 | 1.20× | none |
| TryMutex/TimeoutExpired | 32 | 117.5 | 98.33 | 1.19× | none |
| TryMutex/TimeoutPark | 1 | 1,040,368 | 1,039,882 | 1.00× | 83 B in 1 → 83 B in 1 |
| TryMutex/TimeoutPark | 4 | 1,012,559 | 1,032,956 | 0.98× | 87 B in 1 → 87.5 B in 1 |
| TryMutex/TimeoutPark | 16 | 1,009,754 | 1,019,640 | 0.99× | 95.5 B in 1 → 96 B in 1 |
| TryMutex/TimeoutPark | 32 | 1,007,519 | 1,002,116 | 1.01× | 109.5 B in 1 → 111 B in 1 |
| TryMutex/Contended0 | 1 | 6.804 | 6.890 | 0.99× | none |
| TryMutex/Contended0 | 4 | 7.793 | 7.511 | 1.04× | none |
| TryMutex/Contended0 | 16 | 8.735 | 7.833 | 1.12× | none |
| TryMutex/Contended0 | 32 | 8.812 | 7.996 | 1.10× | none |
| TryMutex/Contended100 | 1 | 106.0 | 105.7 | 1.00× | none |
| TryMutex/Contended100 | 4 | 113.2 | 113.6 | 1.00× | none |
| TryMutex/Contended100 | 16 | 120.5 | 122.0 | 0.99× | none |
| TryMutex/Contended100 | 32 | 125.8 | 127.7 | 0.99× | none |
| Semaphore/1/Uncontended | 1 | 7.737 | 7.117 | 1.09× | none |
| Semaphore/1/Uncontended | 4 | 7.788 | 7.219 | 1.08× | none |
| Semaphore/1/Uncontended | 16 | 7.794 | 7.191 | 1.08× | none |
| Semaphore/1/Uncontended | 32 | 7.967 | 7.184 | 1.11× | none |
| Semaphore/1/Parallel | 1 | 8.603 | 8.256 | 1.04× | none |
| Semaphore/1/Parallel | 4 | 9.549 | 9.435 | 1.01× | none |
| Semaphore/1/Parallel | 16 | 10.44 | 9.250 | 1.13× | none |
| Semaphore/1/Parallel | 32 | 10.50 | 9.695 | 1.08× | none |
| Semaphore/8/Uncontended | 1 | 15.96 | 7.152 | 2.23× | none |
| Semaphore/8/Uncontended | 4 | 16.13 | 7.331 | 2.20× | none |
| Semaphore/8/Uncontended | 16 | 16.16 | 7.218 | 2.24× | none |
| Semaphore/8/Uncontended | 32 | 16.17 | 7.204 | 2.25× | none |
| Semaphore/8/Parallel | 1 | 16.64 | 8.255 | 2.02× | none |
| Semaphore/8/Parallel | 4 | 18.35 | 9.431 | 1.95× | none |
| Semaphore/8/Parallel | 16 | 27.82 | 8.451 | 3.29× | none |
| Semaphore/8/Parallel | 32 | 36.77 | 8.368 | 4.39× | none |
| Semaphore/64/Uncontended | 1 | 15.98 | 7.147 | 2.24× | none |
| Semaphore/64/Uncontended | 4 | 16.09 | 7.219 | 2.23× | none |
| Semaphore/64/Uncontended | 16 | 16.22 | 7.219 | 2.25× | none |
| Semaphore/64/Uncontended | 32 | 16.27 | 7.216 | 2.25× | none |
| Semaphore/64/Parallel | 1 | 16.63 | 8.229 | 2.02× | none |
| Semaphore/64/Parallel | 4 | 18.49 | 9.478 | 1.95× | none |
| Semaphore/64/Parallel | 16 | 20.30 | 8.263 | 2.46× | none |
| Semaphore/64/Parallel | 32 | 20.20 | 8.340 | 2.42× | none |
| Semaphore/TryFailure | 1 | 4.138 | 3.503 | 1.18× | none |
| Semaphore/TryFailure | 4 | 4.167 | 3.543 | 1.18× | none |
| Semaphore/TryFailure | 16 | 4.180 | 3.542 | 1.18× | none |
| Semaphore/TryFailure | 32 | 4.169 | 3.540 | 1.18× | none |
| Semaphore/TimeoutSuccess | 1 | 8.117 | 8.877 | 0.91× | none |
| Semaphore/TimeoutSuccess | 4 | 8.127 | 8.818 | 0.92× | none |
| Semaphore/TimeoutSuccess | 16 | 8.154 | 8.588 | 0.95× | none |
| Semaphore/TimeoutSuccess | 32 | 8.170 | 8.522 | 0.96× | none |
| Semaphore/Value | 1 | 7.268 | 1.855 | 3.92× | none |
| Semaphore/Value | 4 | 7.440 | 1.863 | 3.99× | none |
| Semaphore/Value | 16 | 7.514 | 1.871 | 4.02× | none |
| Semaphore/Value | 32 | 7.456 | 1.867 | 3.99× | none |
| Semaphore/Full8TryFailure | 1 | 9.069 | 3.497 | 2.59× | none |
| Semaphore/Full8TryFailure | 4 | 9.114 | 3.525 | 2.59× | none |
| Semaphore/Full8TryFailure | 16 | 9.117 | 3.545 | 2.57× | none |
| Semaphore/Full8TryFailure | 32 | 9.138 | 3.510 | 2.60× | none |
| NamedMutex/LongKey32 | 1 | 15.71 | 12.44 | 1.26× | none |
| NamedMutex/LongKey32 | 4 | 15.75 | 12.50 | 1.26× | none |
| NamedMutex/LongKey32 | 16 | 15.75 | 12.53 | 1.26× | none |
| NamedMutex/LongKey32 | 32 | 15.66 | 12.47 | 1.26× | none |
| NamedMutex/LongKey128 | 1 | 19.02 | 15.23 | 1.25× | none |
| NamedMutex/LongKey128 | 4 | 18.99 | 15.34 | 1.24× | none |
| NamedMutex/LongKey128 | 16 | 18.96 | 15.26 | 1.24× | none |
| NamedMutex/LongKey128 | 32 | 19.23 | 15.45 | 1.25× | none |
| NamedMutex/LongKey1024 | 1 | 18.96 | 15.29 | 1.24× | none |
| NamedMutex/LongKey1024 | 4 | 19.01 | 15.50 | 1.23× | none |
| NamedMutex/LongKey1024 | 16 | 19.19 | 15.38 | 1.25× | none |
| NamedMutex/LongKey1024 | 32 | 19.07 | 15.38 | 1.24× | none |
| NamedMutex/HotKey | 1 | 16.06 | 12.46 | 1.29× | none |
| NamedMutex/HotKey | 4 | 15.98 | 12.53 | 1.27× | none |
| NamedMutex/HotKey | 16 | 15.91 | 12.54 | 1.27× | none |
| NamedMutex/HotKey | 32 | 15.90 | 12.52 | 1.27× | none |
| NamedMutex/SameKeyParallel | 1 | 16.50 | 13.55 | 1.22× | none |
| NamedMutex/SameKeyParallel | 4 | 17.91 | 14.19 | 1.26× | none |
| NamedMutex/SameKeyParallel | 16 | 18.94 | 14.50 | 1.31× | none |
| NamedMutex/SameKeyParallel | 32 | 19.41 | 15.19 | 1.28× | none |
| NamedMutex/IndependentKeys | 1 | 16.69 | 13.75 | 1.21× | none |
| NamedMutex/IndependentKeys | 4 | 8.424 | 8.154 | 1.03× | none |
| NamedMutex/IndependentKeys | 16 | 3.933 | 4.167 | 0.94× | none |
| NamedMutex/IndependentKeys | 32 | 3.945 | 4.031 | 0.98× | none |
| NamedMutex/ManyKeys | 1 | 82.05 | 75.70 | 1.08× | none |
| NamedMutex/ManyKeys | 4 | 83.77 | 76.81 | 1.09× | none |
| NamedMutex/ManyKeys | 16 | 83.72 | 77.36 | 1.08× | none |
| NamedMutex/ManyKeys | 32 | 83.60 | 77.48 | 1.08× | none |
| NamedMutex/UniqueNames | 1 | 419.1 | 113.2 | 3.70× | 119.5 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 4 | 323.6 | 112.0 | 2.89× | 121.5 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 16 | 352.9 | 115.3 | 3.06× | 123 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 32 | 352.6 | 115.2 | 3.06× | 124.5 B in 2 → 17 B in 1 |
| NamedMutex/Create1024 | 1 | 188,794 | 89652.0 | 2.11× | 111588 B in 1471.5 → 21928 B in 195 |
| NamedMutex/Create1024 | 4 | 170,951 | 85884.0 | 1.99× | 110988 B in 1469.5 → 21928 B in 195 |
| NamedMutex/Create1024 | 16 | 181,294 | 88458.5 | 2.05× | 110496 B in 1466.5 → 21928 B in 195 |
| NamedMutex/Create1024 | 32 | 186,668 | 90708.5 | 2.06× | 110636 B in 1468 → 21928 B in 195 |
| OnceMutex/First | 1 | 23.23 | 23.38 | 0.99× | 8 B in 1 → 8 B in 1 |
| OnceMutex/First | 4 | 20.96 | 21.38 | 0.98× | 8 B in 1 → 8 B in 1 |
| OnceMutex/First | 16 | 21.58 | 22.01 | 0.98× | 8 B in 1 → 8 B in 1 |
| OnceMutex/First | 32 | 22.10 | 22.39 | 0.99× | 8 B in 1 → 8 B in 1 |
| OnceMutex/Completed | 1 | 3.114 | 3.008 | 1.04× | none |
| OnceMutex/Completed | 4 | 3.136 | 3.000 | 1.05× | none |
| OnceMutex/Completed | 16 | 3.128 | 3.020 | 1.04× | none |
| OnceMutex/Completed | 32 | 3.127 | 3.013 | 1.04× | none |
| OnceMutex/CompletedParallel | 1 | 3.516 | 4.077 | 0.86× | none |
| OnceMutex/CompletedParallel | 4 | 0.919 | 1.067 | 0.86× | none |
| OnceMutex/CompletedParallel | 16 | 0.295 | 0.333 | 0.89× | none |
| OnceMutex/CompletedParallel | 32 | 0.282 | 0.287 | 0.98× | none |
| NamedOnceMutex/Uncontended | 1 | 57.80 | 28.83 | 2.00× | none |
| NamedOnceMutex/Uncontended | 4 | 59.11 | 29.45 | 2.01× | none |
| NamedOnceMutex/Uncontended | 16 | 59.09 | 29.55 | 2.00× | none |
| NamedOnceMutex/Uncontended | 32 | 59.20 | 29.47 | 2.01× | none |
| NamedOnceMutex/SameKeyParallel | 1 | 58.95 | 30.24 | 1.95× | none |
| NamedOnceMutex/SameKeyParallel | 4 | 70.94 | 33.03 | 2.15× | none |
| NamedOnceMutex/SameKeyParallel | 16 | 110.0 | 41.88 | 2.63× | 1 B in 0 → 0 B in 0 |
| NamedOnceMutex/SameKeyParallel | 32 | 120.0 | 50.21 | 2.39× | 1 B in 0 → 0 B in 0 |
| NamedOnceMutex/IndependentKeys | 1 | 67.40 | 26.85 | 2.51× | none |
| NamedOnceMutex/IndependentKeys | 4 | 17.91 | 12.48 | 1.43× | none |
| NamedOnceMutex/IndependentKeys | 16 | 6.581 | 5.652 | 1.16× | none |
| NamedOnceMutex/IndependentKeys | 32 | 5.987 | 5.214 | 1.15× | none |
| TryFailureParallel | 1 | 3.709 | 4.297 | 0.86× | none |
| TryFailureParallel | 4 | 0.974 | 1.124 | 0.87× | none |
| TryFailureParallel | 16 | 0.308 | 0.345 | 0.89× | none |
| TryFailureParallel | 32 | 0.284 | 0.313 | 0.91× | none |
| Construction/TryMutex | 1 | 26.95 | 21.74 | 1.24× | 64 B in 1 → 16 B in 1 |
| Construction/TryMutex | 4 | 22.34 | 17.17 | 1.30× | 64 B in 1 → 16 B in 1 |
| Construction/TryMutex | 16 | 27.18 | 18.60 | 1.46× | 64 B in 1 → 16 B in 1 |
| Construction/TryMutex | 32 | 29.45 | 18.95 | 1.55× | 64 B in 1 → 16 B in 1 |
| Construction/Semaphore | 1 | 27.59 | 21.46 | 1.29× | 80 B in 1 → 24 B in 1 |
| Construction/Semaphore | 4 | 25.61 | 19.22 | 1.33× | 80 B in 1 → 24 B in 1 |
| Construction/Semaphore | 16 | 27.46 | 20.98 | 1.31× | 80 B in 1 → 24 B in 1 |
| Construction/Semaphore | 32 | 32.42 | 20.48 | 1.58× | 80 B in 1 → 24 B in 1 |
| Construction/NamedMutexFirstKey | 1 | 72.19 | 68.81 | 1.05× | 136 B in 2 → 136 B in 2 |
| Construction/NamedMutexFirstKey | 4 | 66.48 | 63.70 | 1.04× | 136 B in 2 → 136 B in 2 |
| Construction/NamedMutexFirstKey | 16 | 70.58 | 73.45 | 0.96× | 136 B in 2 → 136 B in 2 |
| Construction/NamedMutexFirstKey | 32 | 78.03 | 73.73 | 1.06× | 136 B in 2 → 136 B in 2 |
| Construction/NamedOnceFirstKey | 1 | 121.8 | 105.8 | 1.15× | 224 B in 2 → 144 B in 2 |
| Construction/NamedOnceFirstKey | 4 | 116.2 | 97.19 | 1.20× | 224 B in 2 → 144 B in 2 |
| Construction/NamedOnceFirstKey | 16 | 124.8 | 107.6 | 1.16× | 224 B in 2 → 144 B in 2 |
| Construction/NamedOnceFirstKey | 32 | 133.2 | 106.0 | 1.26× | 224 B in 2 → 144 B in 2 |
| ControlWaitGroup/1 | 1 | 473.4 | 498.1 | 0.95× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/1 | 4 | 439.5 | 471.8 | 0.93× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/1 | 16 | 447.8 | 487.2 | 0.92× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/1 | 32 | 454.4 | 476.7 | 0.95× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/16 | 1 | 470.1 | 503.4 | 0.93× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/16 | 4 | 248.9 | 250.0 | 1.00× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/16 | 16 | 243.1 | 240.9 | 1.01× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/16 | 32 | 265.1 | 233.4 | 1.14× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/256 | 1 | 484.7 | 507.1 | 0.96× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/256 | 4 | 217.1 | 193.0 | 1.12× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/256 | 16 | 297.9 | 213.3 | 1.40× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/256 | 32 | 243.1 | 210.2 | 1.16× | 16 B in 1 → 16 B in 1 |
| ControlWaitGroup/Aborted | 1 | 2.886 | 3.316 | 0.87× | none |
| ControlWaitGroup/Aborted | 4 | 2.921 | 3.349 | 0.87× | none |
| ControlWaitGroup/Aborted | 16 | 2.928 | 3.351 | 0.87× | none |
| ControlWaitGroup/Aborted | 32 | 2.938 | 3.339 | 0.88× | none |
| SyncFlag/Read | 1 | 2.048 | 1.851 | 1.11× | none |
| SyncFlag/Read | 4 | 2.058 | 1.857 | 1.11× | none |
| SyncFlag/Read | 16 | 2.083 | 1.860 | 1.12× | none |
| SyncFlag/Read | 32 | 2.067 | 1.868 | 1.11× | none |
| SyncFlag/Write | 1 | 13.57 | 13.52 | 1.00× | none |
| SyncFlag/Write | 4 | 13.68 | 13.50 | 1.01× | none |
| SyncFlag/Write | 16 | 13.69 | 13.56 | 1.01× | none |
| SyncFlag/Write | 32 | 13.75 | 13.63 | 1.01× | none |
| StandardMutex/Uncontended | 1 | 5.310 | 5.450 | 0.97× | none |
| StandardMutex/Uncontended | 4 | 5.326 | 5.465 | 0.97× | none |
| StandardMutex/Uncontended | 16 | 5.344 | 5.472 | 0.98× | none |
| StandardMutex/Uncontended | 32 | 5.328 | 5.457 | 0.98× | none |
| StandardMutex/Contended | 1 | 6.028 | 6.332 | 0.95× | none |
| StandardMutex/Contended | 4 | 24.41 | 24.37 | 1.00× | none |
| StandardMutex/Contended | 16 | 64.16 | 64.61 | 0.99× | none |
| StandardMutex/Contended | 32 | 66.03 | 67.03 | 0.99× | none |
