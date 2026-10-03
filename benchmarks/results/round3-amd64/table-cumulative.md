| Workload | Ps | channels ns/op | New ns/op | Ratio | Allocated per op (channels → new) |
| --- | ---: | ---: | ---: | ---: | --- |
| TryMutex/Uncontended | 1 | 22.23 | 3.661 | 6.07× | none |
| TryMutex/Uncontended | 4 | 22.09 | 3.625 | 6.09× | none |
| TryMutex/Uncontended | 16 | 22.09 | 3.633 | 6.08× | none |
| TryMutex/Uncontended | 32 | 22.09 | 3.637 | 6.07× | none |
| TryMutex/TrySuccess | 1 | 22.27 | 3.890 | 5.73× | none |
| TryMutex/TrySuccess | 4 | 22.08 | 3.889 | 5.68× | none |
| TryMutex/TrySuccess | 16 | 22.20 | 3.910 | 5.68× | none |
| TryMutex/TrySuccess | 32 | 22.13 | 3.939 | 5.62× | none |
| TryMutex/TryFailure | 1 | 3.087 | 0.619 | 4.98× | none |
| TryMutex/TryFailure | 4 | 3.095 | 0.623 | 4.97× | none |
| TryMutex/TryFailure | 16 | 3.117 | 0.627 | 4.97× | none |
| TryMutex/TryFailure | 32 | 3.138 | 0.625 | 5.02× | none |
| TryMutex/TimeoutSuccess | 1 | 23.53 | 3.586 | 6.56× | none |
| TryMutex/TimeoutSuccess | 4 | 23.50 | 3.607 | 6.51× | none |
| TryMutex/TimeoutSuccess | 16 | 23.44 | 3.593 | 6.52× | none |
| TryMutex/TimeoutSuccess | 32 | 23.64 | 3.611 | 6.55× | none |
| TryMutex/TimeoutExpired | 1 | 282.2 | 61.12 | 4.62× | 248 B in 3 → 0 B in 0 |
| TryMutex/TimeoutExpired | 4 | 271.5 | 61.30 | 4.43× | 248 B in 3 → 0 B in 0 |
| TryMutex/TimeoutExpired | 16 | 286.4 | 61.62 | 4.65× | 248 B in 3 → 0 B in 0 |
| TryMutex/TimeoutExpired | 32 | 288.2 | 61.35 | 4.70× | 248 B in 3 → 0 B in 0 |
| TryMutex/TimeoutPark | 1 | 1,044,701 | 1,044,788 | 1.00× | 249 B in 3 → 117 B in 1 |
| TryMutex/TimeoutPark | 4 | 1,043,617 | 1,021,739 | 1.02× | 249 B in 3 → 122 B in 1 |
| TryMutex/TimeoutPark | 16 | 1,043,764 | 996,589 | 1.05× | 248 B in 3 → 131 B in 1 |
| TryMutex/TimeoutPark | 32 | 1,038,172 | 1,023,474 | 1.01× | 249 B in 3 → 146 B in 1 |
| TryMutex/Contended0 | 1 | 22.05 | 3.623 | 6.08× | none |
| TryMutex/Contended0 | 4 | 110.4 | 3.692 | 29.90× | none |
| TryMutex/Contended0 | 16 | 111.7 | 3.867 | 28.89× | none |
| TryMutex/Contended0 | 32 | 109.7 | 3.958 | 27.71× | none |
| TryMutex/Contended100 | 1 | 93.55 | 80.79 | 1.16× | none |
| TryMutex/Contended100 | 4 | 191.4 | 84.81 | 2.26× | none |
| TryMutex/Contended100 | 16 | 189.8 | 87.45 | 2.17× | none |
| TryMutex/Contended100 | 32 | 196.4 | 89.53 | 2.19× | none |
| Semaphore/1/Uncontended | 1 | 21.61 | 3.707 | 5.83× | none |
| Semaphore/1/Uncontended | 4 | 22.12 | 3.739 | 5.92× | none |
| Semaphore/1/Uncontended | 16 | 22.12 | 3.732 | 5.93× | none |
| Semaphore/1/Uncontended | 32 | 22.05 | 3.762 | 5.86× | none |
| Semaphore/1/Parallel | 1 | 21.26 | 3.734 | 5.69× | none |
| Semaphore/1/Parallel | 4 | 110.3 | 4.005 | 27.56× | none |
| Semaphore/1/Parallel | 16 | 109.8 | 4.095 | 26.83× | none |
| Semaphore/1/Parallel | 32 | 110.2 | 4.117 | 26.77× | none |
| Semaphore/8/Uncontended | 1 | 22.71 | 3.708 | 6.13× | none |
| Semaphore/8/Uncontended | 4 | 23.20 | 3.743 | 6.20× | none |
| Semaphore/8/Uncontended | 16 | 23.24 | 3.750 | 6.20× | none |
| Semaphore/8/Uncontended | 32 | 23.23 | 3.760 | 6.18× | none |
| Semaphore/8/Parallel | 1 | 22.33 | 3.726 | 5.99× | none |
| Semaphore/8/Parallel | 4 | 39.20 | 4.909 | 7.99× | none |
| Semaphore/8/Parallel | 16 | 227.6 | 3.848 | 59.16× | none |
| Semaphore/8/Parallel | 32 | 231.4 | 3.878 | 59.66× | none |
| Semaphore/64/Uncontended | 1 | 22.13 | 3.708 | 5.97× | none |
| Semaphore/64/Uncontended | 4 | 22.60 | 3.747 | 6.03× | none |
| Semaphore/64/Uncontended | 16 | 23.02 | 3.744 | 6.15× | none |
| Semaphore/64/Uncontended | 32 | 22.88 | 3.752 | 6.10× | none |
| Semaphore/64/Parallel | 1 | 21.74 | 3.732 | 5.83× | none |
| Semaphore/64/Parallel | 4 | 38.34 | 4.791 | 8.00× | none |
| Semaphore/64/Parallel | 16 | 40.30 | 3.845 | 10.48× | none |
| Semaphore/64/Parallel | 32 | 44.23 | 3.849 | 11.49× | none |
| Semaphore/TryFailure | 1 | 2.885 | 0.621 | 4.65× | none |
| Semaphore/TryFailure | 4 | 2.930 | 0.627 | 4.67× | none |
| Semaphore/TryFailure | 16 | 2.936 | 0.627 | 4.68× | none |
| Semaphore/TryFailure | 32 | 2.953 | 0.627 | 4.71× | none |
| Semaphore/TimeoutSuccess | 1 | 23.34 | 3.732 | 6.25× | none |
| Semaphore/TimeoutSuccess | 4 | 23.42 | 3.742 | 6.26× | none |
| Semaphore/TimeoutSuccess | 16 | 23.52 | 3.760 | 6.25× | none |
| Semaphore/TimeoutSuccess | 32 | 23.73 | 3.755 | 6.32× | none |
| Semaphore/Value | 1 | 1.234 | 0.462 | 2.67× | none |
| Semaphore/Value | 4 | 1.236 | 0.463 | 2.67× | none |
| Semaphore/Value | 16 | 1.240 | 0.464 | 2.67× | none |
| Semaphore/Value | 32 | 1.240 | 0.463 | 2.68× | none |
| Semaphore/Full8TryFailure | 1 | 2.883 | 0.619 | 4.66× | none |
| Semaphore/Full8TryFailure | 4 | 2.889 | 0.622 | 4.64× | none |
| Semaphore/Full8TryFailure | 16 | 2.893 | 0.621 | 4.66× | none |
| Semaphore/Full8TryFailure | 32 | 2.907 | 0.623 | 4.67× | none |
| NamedMutex/LongKey32 | 1 | 40.64 | 5.996 | 6.78× | none |
| NamedMutex/LongKey32 | 4 | 40.66 | 6.005 | 6.77× | none |
| NamedMutex/LongKey32 | 16 | 40.78 | 6.011 | 6.78× | none |
| NamedMutex/LongKey32 | 32 | 40.89 | 6.022 | 6.79× | none |
| NamedMutex/LongKey128 | 1 | 41.06 | 8.849 | 4.64× | none |
| NamedMutex/LongKey128 | 4 | 40.80 | 8.881 | 4.59× | none |
| NamedMutex/LongKey128 | 16 | 41.09 | 8.885 | 4.62× | none |
| NamedMutex/LongKey128 | 32 | 41.20 | 8.864 | 4.65× | none |
| NamedMutex/LongKey1024 | 1 | 41.09 | 8.849 | 4.64× | none |
| NamedMutex/LongKey1024 | 4 | 40.88 | 8.861 | 4.61× | none |
| NamedMutex/LongKey1024 | 16 | 40.97 | 8.912 | 4.60× | none |
| NamedMutex/LongKey1024 | 32 | 41.09 | 8.872 | 4.63× | none |
| NamedMutex/HotKey | 1 | 40.81 | 6.388 | 6.39× | none |
| NamedMutex/HotKey | 4 | 40.53 | 6.189 | 6.55× | none |
| NamedMutex/HotKey | 16 | 40.72 | 6.239 | 6.53× | none |
| NamedMutex/HotKey | 32 | 40.64 | 6.215 | 6.54× | none |
| NamedMutex/SameKeyParallel | 1 | 40.69 | 6.171 | 6.59× | none |
| NamedMutex/SameKeyParallel | 4 | 138.4 | 6.213 | 22.29× | none |
| NamedMutex/SameKeyParallel | 16 | 138.4 | 6.413 | 21.58× | none |
| NamedMutex/SameKeyParallel | 32 | 136.1 | 6.576 | 20.70× | none |
| NamedMutex/IndependentKeys | 1 | 40.67 | 6.191 | 6.57× | none |
| NamedMutex/IndependentKeys | 4 | 70.61 | 3.698 | 19.09× | none |
| NamedMutex/IndependentKeys | 16 | 118.4 | 1.710 | 69.26× | none |
| NamedMutex/IndependentKeys | 32 | 140.9 | 1.804 | 78.05× | none |
| NamedMutex/ManyKeys | 1 | 48.59 | 33.88 | 1.43× | none |
| NamedMutex/ManyKeys | 4 | 49.15 | 33.62 | 1.46× | none |
| NamedMutex/ManyKeys | 16 | 49.61 | 33.64 | 1.47× | none |
| NamedMutex/ManyKeys | 32 | 49.75 | 33.89 | 1.47× | none |
| NamedMutex/UniqueNames | 1 | 328.8 | 61.11 | 5.38× | 198 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 4 | 285.5 | 54.84 | 5.21× | 199 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 16 | 302.1 | 56.73 | 5.33× | 203 B in 2 → 17 B in 1 |
| NamedMutex/UniqueNames | 32 | 299.4 | 57.39 | 5.22× | 201.5 B in 2 → 17 B in 1 |
| NamedMutex/Create1024 | 1 | 155,665 | 47881.0 | 3.25× | 223720 B in 1047 → 21136 B in 195 |
| NamedMutex/Create1024 | 4 | 139,982 | 40647.5 | 3.44× | 223720 B in 1047 → 21136 B in 195 |
| NamedMutex/Create1024 | 16 | 151,830 | 42027.5 | 3.61× | 223720 B in 1047 → 21136 B in 195 |
| NamedMutex/Create1024 | 32 | 155,142 | 43015.5 | 3.61× | 223720 B in 1047 → 21136 B in 195 |
| OnceMutex/First | 1 | 13.30 | 13.65 | 0.97× | 16 B in 1 → 16 B in 1 |
| OnceMutex/First | 4 | 12.07 | 12.06 | 1.00× | 16 B in 1 → 16 B in 1 |
| OnceMutex/First | 16 | 12.69 | 13.32 | 0.95× | 16 B in 1 → 16 B in 1 |
| OnceMutex/First | 32 | 13.21 | 13.77 | 0.96× | 16 B in 1 → 16 B in 1 |
| OnceMutex/Completed | 1 | 3.502 | 0.456 | 7.69× | none |
| OnceMutex/Completed | 4 | 3.510 | 0.461 | 7.61× | none |
| OnceMutex/Completed | 16 | 3.529 | 0.460 | 7.66× | none |
| OnceMutex/Completed | 32 | 3.524 | 0.460 | 7.66× | none |
| OnceMutex/CompletedParallel | 1 | 3.543 | 0.414 | 8.56× | none |
| OnceMutex/CompletedParallel | 4 | 16.49 | 0.112 | 146.84× | none |
| OnceMutex/CompletedParallel | 16 | 44.49 | 0.038 | 1183.24× | none |
| OnceMutex/CompletedParallel | 32 | 36.78 | 0.033 | 1098.58× | none |
| NamedOnceMutex/Uncontended | 1 | 85.80 | 17.32 | 4.95× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/Uncontended | 4 | 75.02 | 17.67 | 4.25× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/Uncontended | 16 | 75.79 | 17.74 | 4.27× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/Uncontended | 32 | 77.45 | 17.63 | 4.39× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/SameKeyParallel | 1 | 84.35 | 17.90 | 4.71× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/SameKeyParallel | 4 | 97.67 | 19.43 | 5.03× | 15 B in 0 → 0 B in 0 |
| NamedOnceMutex/SameKeyParallel | 16 | 127.9 | 24.12 | 5.30× | 14 B in 0 → 0 B in 0 |
| NamedOnceMutex/SameKeyParallel | 32 | 188.5 | 28.02 | 6.73× | 12 B in 0 → 0 B in 0 |
| NamedOnceMutex/IndependentKeys | 1 | 79.33 | 13.22 | 6.00× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/IndependentKeys | 4 | 92.91 | 5.323 | 17.45× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/IndependentKeys | 16 | 148.7 | 2.177 | 68.28× | 16 B in 1 → 0 B in 0 |
| NamedOnceMutex/IndependentKeys | 32 | 163.6 | 2.293 | 71.33× | 16 B in 1 → 0 B in 0 |
| TryFailureParallel | 1 | 3.087 | 0.618 | 5.00× | none |
| TryFailureParallel | 4 | 0.803 | 0.161 | 4.98× | none |
| TryFailureParallel | 16 | 0.270 | 0.053 | 5.09× | none |
| TryFailureParallel | 32 | 0.227 | 0.050 | 4.57× | none |
| Construction/TryMutex | 1 | 37.72 | 15.48 | 2.44× | 120 B in 2 → 24 B in 1 |
| Construction/TryMutex | 4 | 32.80 | 12.86 | 2.55× | 120 B in 2 → 24 B in 1 |
| Construction/TryMutex | 16 | 41.40 | 14.50 | 2.85× | 120 B in 2 → 24 B in 1 |
| Construction/TryMutex | 32 | 45.10 | 15.61 | 2.89× | 120 B in 2 → 24 B in 1 |
| Construction/Semaphore | 1 | 37.47 | 15.82 | 2.37× | 120 B in 2 → 48 B in 1 |
| Construction/Semaphore | 4 | 32.52 | 14.52 | 2.24× | 120 B in 2 → 48 B in 1 |
| Construction/Semaphore | 16 | 38.91 | 16.25 | 2.39× | 120 B in 2 → 48 B in 1 |
| Construction/Semaphore | 32 | 44.74 | 19.10 | 2.34× | 120 B in 2 → 48 B in 1 |
| Construction/NamedMutexFirstKey | 1 | 158.5 | 48.51 | 3.27× | 384 B in 4 → 144 B in 2 |
| Construction/NamedMutexFirstKey | 4 | 151.4 | 45.34 | 3.34× | 384 B in 4 → 144 B in 2 |
| Construction/NamedMutexFirstKey | 16 | 176.8 | 48.55 | 3.64× | 384 B in 4 → 144 B in 2 |
| Construction/NamedMutexFirstKey | 32 | 176.3 | 59.25 | 2.98× | 384 B in 4 → 144 B in 2 |
| Construction/NamedOnceFirstKey | 1 | 168.8 | 70.30 | 2.40× | 288 B in 4 → 152 B in 2 |
| Construction/NamedOnceFirstKey | 4 | 154.7 | 63.17 | 2.45× | 288 B in 4 → 152 B in 2 |
| Construction/NamedOnceFirstKey | 16 | 167.9 | 68.00 | 2.47× | 288 B in 4 → 152 B in 2 |
| Construction/NamedOnceFirstKey | 32 | 170.0 | 77.69 | 2.19× | 288 B in 4 → 152 B in 2 |
| ControlWaitGroup/1 | 1 | 362.0 | 291.9 | 1.24× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/1 | 4 | 357.2 | 281.7 | 1.27× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/1 | 16 | 360.4 | 284.0 | 1.27× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/1 | 32 | 361.1 | 291.4 | 1.24× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 1 | 357.2 | 285.9 | 1.25× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 4 | 278.6 | 171.1 | 1.63× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 16 | 268.5 | 160.0 | 1.68× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/16 | 32 | 278.4 | 167.4 | 1.66× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 1 | 358.6 | 286.3 | 1.25× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 4 | 255.8 | 140.0 | 1.83× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 16 | 434.6 | 143.2 | 3.03× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/256 | 32 | 436.0 | 166.8 | 2.61× | 24 B in 1 → 24 B in 1 |
| ControlWaitGroup/Aborted | 1 | 3.538 | 0.455 | 7.78× | none |
| ControlWaitGroup/Aborted | 4 | 3.590 | 0.460 | 7.80× | none |
| ControlWaitGroup/Aborted | 16 | 3.599 | 0.467 | 7.71× | none |
| ControlWaitGroup/Aborted | 32 | 3.603 | 0.462 | 7.79× | none |
| SyncFlag/Read | 1 | 1.013 | 1.016 | 1.00× | none |
| SyncFlag/Read | 4 | 1.019 | 1.020 | 1.00× | none |
| SyncFlag/Read | 16 | 1.029 | 1.021 | 1.01× | none |
| SyncFlag/Read | 32 | 1.027 | 1.024 | 1.00× | none |
| SyncFlag/Write | 1 | 9.723 | 9.717 | 1.00× | none |
| SyncFlag/Write | 4 | 9.749 | 9.730 | 1.00× | none |
| SyncFlag/Write | 16 | 9.762 | 9.780 | 1.00× | none |
| SyncFlag/Write | 32 | 9.768 | 9.787 | 1.00× | none |
| StandardMutex/Uncontended | 1 | 3.535 | 3.543 | 1.00× | none |
| StandardMutex/Uncontended | 4 | 3.546 | 3.543 | 1.00× | none |
| StandardMutex/Uncontended | 16 | 3.559 | 3.553 | 1.00× | none |
| StandardMutex/Uncontended | 32 | 3.577 | 3.579 | 1.00× | none |
| StandardMutex/Contended | 1 | 3.532 | 3.535 | 1.00× | none |
| StandardMutex/Contended | 4 | 10.43 | 10.28 | 1.02× | none |
| StandardMutex/Contended | 16 | 49.22 | 47.03 | 1.05× | none |
| StandardMutex/Contended | 32 | 42.00 | 44.10 | 0.95× | none |
