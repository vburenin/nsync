| Workload | Ps | 6e6bc42 ns/op | New ns/op | Ratio | Allocated per op (6e6bc42 → new) |
| --- | ---: | ---: | ---: | ---: | --- |
| TryMutex/Uncontended | 1 | 3.598 | 3.635 | 0.99× | none |
| TryMutex/Uncontended | 4 | 3.622 | 3.629 | 1.00× | none |
| TryMutex/Uncontended | 16 | 3.620 | 3.629 | 1.00× | none |
| TryMutex/Uncontended | 32 | 3.629 | 3.641 | 1.00× | none |
| TryMutex/TrySuccess | 1 | 3.842 | 3.888 | 0.99× | none |
| TryMutex/TrySuccess | 4 | 3.870 | 3.889 | 1.00× | none |
| TryMutex/TrySuccess | 16 | 3.870 | 3.927 | 0.99× | none |
| TryMutex/TrySuccess | 32 | 3.869 | 3.905 | 0.99× | none |
| TryMutex/TryFailure | 1 | 0.619 | 0.619 | 1.00× | none |
| TryMutex/TryFailure | 4 | 0.622 | 0.621 | 1.00× | none |
| TryMutex/TryFailure | 16 | 0.626 | 0.627 | 1.00× | none |
| TryMutex/TryFailure | 32 | 0.627 | 0.627 | 1.00× | none |
| TryMutex/TimeoutSuccess | 1 | 3.498 | 3.597 | 0.97× | none |
| TryMutex/TimeoutSuccess | 4 | 3.514 | 3.587 | 0.98× | none |
| TryMutex/TimeoutSuccess | 16 | 3.541 | 3.600 | 0.98× | none |
| TryMutex/TimeoutSuccess | 32 | 3.556 | 3.609 | 0.99× | none |
| TryMutex/TimeoutExpired | 1 | 66.92 | 61.16 | 1.09× | none |
| TryMutex/TimeoutExpired | 4 | 67.11 | 61.17 | 1.10× | none |
| TryMutex/TimeoutExpired | 16 | 67.66 | 61.71 | 1.10× | none |
| TryMutex/TimeoutExpired | 32 | 67.48 | 61.54 | 1.10× | none |
| TryMutex/TimeoutPark | 1 | 1,044,866 | 1,044,983 | 1.00× | 116 B in 1 → 117 B in 1 |
| TryMutex/TimeoutPark | 4 | 1,014,511 | 1,016,932 | 1.00× | 120.5 B in 1 → 122 B in 1 |
| TryMutex/TimeoutPark | 16 | 1,013,310 | 995,122 | 1.02× | 130 B in 1 → 131 B in 1 |
| TryMutex/TimeoutPark | 32 | 1,023,148 | 1,023,434 | 1.00× | 145.5 B in 1 → 146.5 B in 1 |
| TryMutex/Contended0 | 1 | 3.621 | 3.625 | 1.00× | none |
| TryMutex/Contended0 | 4 | 4.004 | 3.677 | 1.09× | none |
| TryMutex/Contended0 | 16 | 4.454 | 3.845 | 1.16× | none |
| TryMutex/Contended0 | 32 | 4.484 | 3.907 | 1.15× | none |
| TryMutex/Contended100 | 1 | 80.79 | 80.81 | 1.00× | none |
| TryMutex/Contended100 | 4 | 87.18 | 85.15 | 1.02× | none |
| TryMutex/Contended100 | 16 | 90.97 | 87.06 | 1.04× | none |
| TryMutex/Contended100 | 32 | 92.69 | 89.79 | 1.03× | none |
| Semaphore/1/Uncontended | 1 | 3.602 | 3.706 | 0.97× | none |
| Semaphore/1/Uncontended | 4 | 3.634 | 3.740 | 0.97× | none |
| Semaphore/1/Uncontended | 16 | 3.638 | 3.741 | 0.97× | none |
| Semaphore/1/Uncontended | 32 | 3.642 | 3.749 | 0.97× | none |
| Semaphore/1/Parallel | 1 | 3.664 | 3.727 | 0.98× | none |
| Semaphore/1/Parallel | 4 | 4.027 | 3.972 | 1.01× | none |
| Semaphore/1/Parallel | 16 | 4.447 | 4.044 | 1.10× | none |
| Semaphore/1/Parallel | 32 | 4.521 | 4.123 | 1.10× | none |
| Semaphore/8/Uncontended | 1 | 8.848 | 3.708 | 2.39× | none |
| Semaphore/8/Uncontended | 4 | 8.904 | 3.740 | 2.38× | none |
| Semaphore/8/Uncontended | 16 | 8.933 | 3.762 | 2.37× | none |
| Semaphore/8/Uncontended | 32 | 8.938 | 3.756 | 2.38× | none |
| Semaphore/8/Parallel | 1 | 8.867 | 3.725 | 2.38× | none |
| Semaphore/8/Parallel | 4 | 9.785 | 4.819 | 2.03× | none |
| Semaphore/8/Parallel | 16 | 12.68 | 3.841 | 3.30× | none |
| Semaphore/8/Parallel | 32 | 15.14 | 3.835 | 3.95× | none |
| Semaphore/64/Uncontended | 1 | 8.838 | 3.710 | 2.38× | none |
| Semaphore/64/Uncontended | 4 | 8.927 | 3.745 | 2.38× | none |
| Semaphore/64/Uncontended | 16 | 8.950 | 3.755 | 2.38× | none |
| Semaphore/64/Uncontended | 32 | 8.991 | 3.750 | 2.40× | none |
| Semaphore/64/Parallel | 1 | 8.890 | 3.726 | 2.39× | none |
| Semaphore/64/Parallel | 4 | 9.684 | 4.867 | 1.99× | none |
| Semaphore/64/Parallel | 16 | 10.91 | 3.835 | 2.85× | none |
| Semaphore/64/Parallel | 32 | 11.00 | 3.846 | 2.86× | none |
| Semaphore/TryFailure | 1 | 1.646 | 0.620 | 2.66× | none |
| Semaphore/TryFailure | 4 | 1.668 | 0.627 | 2.66× | none |
| Semaphore/TryFailure | 16 | 1.666 | 0.627 | 2.66× | none |
| Semaphore/TryFailure | 32 | 1.669 | 0.626 | 2.67× | none |
| Semaphore/TimeoutSuccess | 1 | 4.262 | 3.734 | 1.14× | none |
| Semaphore/TimeoutSuccess | 4 | 4.273 | 3.767 | 1.13× | none |
| Semaphore/TimeoutSuccess | 16 | 4.285 | 3.748 | 1.14× | none |
| Semaphore/TimeoutSuccess | 32 | 4.294 | 3.737 | 1.15× | none |
| Semaphore/Value | 1 | 3.571 | 0.461 | 7.74× | none |
| Semaphore/Value | 4 | 3.603 | 0.463 | 7.78× | none |
| Semaphore/Value | 16 | 3.589 | 0.464 | 7.73× | none |
| Semaphore/Value | 32 | 3.590 | 0.463 | 7.75× | none |
| Semaphore/Full8TryFailure | 1 | 4.016 | 0.619 | 6.49× | none |
| Semaphore/Full8TryFailure | 4 | 4.042 | 0.622 | 6.50× | none |
| Semaphore/Full8TryFailure | 16 | 4.042 | 0.621 | 6.51× | none |
| Semaphore/Full8TryFailure | 32 | 4.033 | 0.622 | 6.48× | none |
| NamedMutex/LongKey32 | 1 | 8.436 | 5.986 | 1.41× | none |
| NamedMutex/LongKey32 | 4 | 8.473 | 6.014 | 1.41× | none |
| NamedMutex/LongKey32 | 16 | 8.497 | 6.007 | 1.41× | none |
| NamedMutex/LongKey32 | 32 | 8.514 | 6.005 | 1.42× | none |
| NamedMutex/LongKey128 | 1 | 10.91 | 8.844 | 1.23× | none |
| NamedMutex/LongKey128 | 4 | 10.97 | 8.901 | 1.23× | none |
| NamedMutex/LongKey128 | 16 | 11.02 | 8.905 | 1.24× | none |
| NamedMutex/LongKey128 | 32 | 10.94 | 8.869 | 1.23× | none |
| NamedMutex/LongKey1024 | 1 | 10.89 | 8.870 | 1.23× | none |
| NamedMutex/LongKey1024 | 4 | 10.98 | 8.871 | 1.24× | none |
| NamedMutex/LongKey1024 | 16 | 11.03 | 8.864 | 1.24× | none |
| NamedMutex/LongKey1024 | 32 | 10.93 | 8.882 | 1.23× | none |
| NamedMutex/HotKey | 1 | 8.652 | 6.383 | 1.36× | none |
| NamedMutex/HotKey | 4 | 8.687 | 6.212 | 1.40× | none |
| NamedMutex/HotKey | 16 | 8.692 | 6.226 | 1.40× | none |
| NamedMutex/HotKey | 32 | 8.654 | 6.210 | 1.39× | none |
| NamedMutex/SameKeyParallel | 1 | 8.433 | 6.167 | 1.37× | none |
| NamedMutex/SameKeyParallel | 4 | 9.230 | 6.245 | 1.48× | none |
| NamedMutex/SameKeyParallel | 16 | 10.09 | 6.415 | 1.57× | none |
| NamedMutex/SameKeyParallel | 32 | 10.18 | 6.560 | 1.55× | none |
| NamedMutex/IndependentKeys | 1 | 8.436 | 6.189 | 1.36× | none |
| NamedMutex/IndependentKeys | 4 | 4.380 | 3.692 | 1.19× | none |
| NamedMutex/IndependentKeys | 16 | 1.852 | 1.748 | 1.06× | none |
| NamedMutex/IndependentKeys | 32 | 1.813 | 1.806 | 1.00× | none |
| NamedMutex/ManyKeys | 1 | 43.39 | 33.81 | 1.28× | none |
| NamedMutex/ManyKeys | 4 | 44.38 | 33.58 | 1.32× | none |
| NamedMutex/ManyKeys | 16 | 44.75 | 33.98 | 1.32× | none |
| NamedMutex/ManyKeys | 32 | 44.86 | 34.29 | 1.31× | none |
| NamedMutex/UniqueNames | 1 | 333.7 | 60.76 | 5.49× | 165.5 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 4 | 279.5 | 54.37 | 5.14× | 166 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 16 | 286.6 | 55.77 | 5.14× | 165.5 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 32 | 283.3 | 57.09 | 4.96× | 165 B in 2 → 17 B in 1 |
| NamedMutex/Create1024 | 1 | 133,765 | 47906.5 | 2.79× | 168448 B in 1470 → 21136 B in 195 |
| NamedMutex/Create1024 | 4 | 124,901 | 40471.0 | 3.09× | 168676 B in 1472.5 → 21136 B in 195 |
| NamedMutex/Create1024 | 16 | 137,700 | 41980.5 | 3.28× | 168220 B in 1471.5 → 21136 B in 195 |
| NamedMutex/Create1024 | 32 | 138,456 | 42679.5 | 3.24× | 168236 B in 1472 → 21136 B in 195 |
| OnceMutex/First | 1 | 14.36 | 13.65 | 1.05× | 16 B in 1 → 16 B in 1 |
| OnceMutex/First | 4 | 12.66 | 11.97 | 1.06× | 16 B in 1 → 16 B in 1 |
| OnceMutex/First | 16 | 13.65 | 12.84 | 1.06× | 16 B in 1 → 16 B in 1 |
| OnceMutex/First | 32 | 14.51 | 13.32 | 1.09× | 16 B in 1 → 16 B in 1 |
| OnceMutex/Completed | 1 | 1.645 | 0.452 | 3.64× | none |
| OnceMutex/Completed | 4 | 1.665 | 0.458 | 3.63× | none |
| OnceMutex/Completed | 16 | 1.657 | 0.459 | 3.61× | none |
| OnceMutex/Completed | 32 | 1.662 | 0.462 | 3.59× | none |
| OnceMutex/CompletedParallel | 1 | 1.441 | 0.413 | 3.49× | none |
| OnceMutex/CompletedParallel | 4 | 0.375 | 0.110 | 3.40× | none |
| OnceMutex/CompletedParallel | 16 | 0.120 | 0.038 | 3.20× | none |
| OnceMutex/CompletedParallel | 32 | 0.105 | 0.034 | 3.08× | none |
| NamedOnceMutex/Uncontended | 1 | 26.80 | 17.31 | 1.55× | none |
| NamedOnceMutex/Uncontended | 4 | 27.38 | 17.66 | 1.55× | none |
| NamedOnceMutex/Uncontended | 16 | 27.44 | 17.67 | 1.55× | none |
| NamedOnceMutex/Uncontended | 32 | 27.40 | 17.66 | 1.55× | none |
| NamedOnceMutex/SameKeyParallel | 1 | 27.24 | 17.91 | 1.52× | none |
| NamedOnceMutex/SameKeyParallel | 4 | 31.25 | 19.58 | 1.60× | none |
| NamedOnceMutex/SameKeyParallel | 16 | 45.75 | 22.21 | 2.06× | none |
| NamedOnceMutex/SameKeyParallel | 32 | 53.37 | 25.98 | 2.05× | none |
| NamedOnceMutex/IndependentKeys | 1 | 24.55 | 13.23 | 1.86× | none |
| NamedOnceMutex/IndependentKeys | 4 | 6.520 | 5.285 | 1.23× | none |
| NamedOnceMutex/IndependentKeys | 16 | 2.399 | 2.244 | 1.07× | none |
| NamedOnceMutex/IndependentKeys | 32 | 2.266 | 2.232 | 1.02× | none |
| TryFailureParallel | 1 | 0.619 | 0.618 | 1.00× | none |
| TryFailureParallel | 4 | 0.163 | 0.161 | 1.01× | none |
| TryFailureParallel | 16 | 0.053 | 0.051 | 1.04× | none |
| TryFailureParallel | 32 | 0.050 | 0.050 | 1.00× | none |
| Construction/TryMutex | 1 | 26.86 | 15.48 | 1.73× | 96 B in 1 → 24 B in 1 |
| Construction/TryMutex | 4 | 25.45 | 12.91 | 1.97× | 96 B in 1 → 24 B in 1 |
| Construction/TryMutex | 16 | 31.11 | 14.25 | 2.18× | 96 B in 1 → 24 B in 1 |
| Construction/TryMutex | 32 | 34.15 | 15.37 | 2.22× | 96 B in 1 → 24 B in 1 |
| Construction/Semaphore | 1 | 26.64 | 15.75 | 1.69× | 112 B in 1 → 48 B in 1 |
| Construction/Semaphore | 4 | 26.20 | 14.41 | 1.82× | 112 B in 1 → 48 B in 1 |
| Construction/Semaphore | 16 | 30.30 | 15.86 | 1.91× | 112 B in 1 → 48 B in 1 |
| Construction/Semaphore | 32 | 33.76 | 17.56 | 1.92× | 112 B in 1 → 48 B in 1 |
| Construction/NamedMutexFirstKey | 1 | 58.77 | 48.51 | 1.21× | 208 B in 2 → 144 B in 2 |
| Construction/NamedMutexFirstKey | 4 | 55.39 | 44.67 | 1.24× | 208 B in 2 → 144 B in 2 |
| Construction/NamedMutexFirstKey | 16 | 61.64 | 48.48 | 1.27× | 208 B in 2 → 144 B in 2 |
| Construction/NamedMutexFirstKey | 32 | 67.07 | 51.65 | 1.30× | 208 B in 2 → 144 B in 2 |
| Construction/NamedOnceFirstKey | 1 | 80.20 | 70.19 | 1.14× | 216 B in 2 → 152 B in 2 |
| Construction/NamedOnceFirstKey | 4 | 74.06 | 62.50 | 1.18× | 216 B in 2 → 152 B in 2 |
| Construction/NamedOnceFirstKey | 16 | 81.16 | 67.03 | 1.21× | 216 B in 2 → 152 B in 2 |
| Construction/NamedOnceFirstKey | 32 | 91.49 | 72.06 | 1.27× | 216 B in 2 → 152 B in 2 |
| ControlWaitGroup/1 | 1 | 266.5 | 290.7 | 0.92× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/1 | 4 | 255.3 | 284.5 | 0.90× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/1 | 16 | 257.1 | 284.7 | 0.90× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/1 | 32 | 260.8 | 287.3 | 0.91× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 1 | 262.4 | 287.0 | 0.91× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 4 | 181.4 | 170.2 | 1.07× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 16 | 177.4 | 163.5 | 1.09× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 32 | 186.2 | 166.7 | 1.12× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 1 | 262.9 | 287.7 | 0.91× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 4 | 163.9 | 139.9 | 1.17× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 16 | 384.1 | 144.4 | 2.66× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 32 | 200.6 | 143.3 | 1.40× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/Aborted | 1 | 1.441 | 0.455 | 3.17× | none |
| ControlWaitGroup/Aborted | 4 | 1.468 | 0.461 | 3.19× | none |
| ControlWaitGroup/Aborted | 16 | 1.473 | 0.463 | 3.18× | none |
| ControlWaitGroup/Aborted | 32 | 1.469 | 0.464 | 3.16× | none |
| SyncFlag/Read | 1 | 1.011 | 1.015 | 1.00× | none |
| SyncFlag/Read | 4 | 1.018 | 1.021 | 1.00× | none |
| SyncFlag/Read | 16 | 1.018 | 1.022 | 1.00× | none |
| SyncFlag/Read | 32 | 1.019 | 1.026 | 0.99× | none |
| SyncFlag/Write | 1 | 9.727 | 9.713 | 1.00× | none |
| SyncFlag/Write | 4 | 9.761 | 9.781 | 1.00× | none |
| SyncFlag/Write | 16 | 9.834 | 9.777 | 1.01× | none |
| SyncFlag/Write | 32 | 9.758 | 9.803 | 1.00× | none |
| StandardMutex/Uncontended | 1 | 3.555 | 3.534 | 1.01× | none |
| StandardMutex/Uncontended | 4 | 3.558 | 3.545 | 1.00× | none |
| StandardMutex/Uncontended | 16 | 3.593 | 3.564 | 1.01× | none |
| StandardMutex/Uncontended | 32 | 3.567 | 3.561 | 1.00× | none |
| StandardMutex/Contended | 1 | 3.542 | 3.532 | 1.00× | none |
| StandardMutex/Contended | 4 | 10.55 | 10.36 | 1.02× | none |
| StandardMutex/Contended | 16 | 47.05 | 51.49 | 0.91× | none |
| StandardMutex/Contended | 32 | 54.94 | 55.08 | 1.00× | none |
