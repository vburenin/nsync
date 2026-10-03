| Call | Ps | Without context ns/op | `Background` ns/op | Cancelable ns/op | Allocated per op |
| --- | ---: | ---: | ---: | ---: | --- |
| TryMutex.LockContext | 1 | 3.63 | 3.77 (+0.14) | 4.02 (+0.39) | none |
| TryMutex.LockContext | 4 | 3.62 | 3.74 (+0.12) | 4.05 (+0.43) | none |
| TryMutex.LockContext | 16 | 3.62 | 3.76 (+0.14) | 4.07 (+0.45) | none |
| TryMutex.LockContext | 32 | 3.61 | 3.76 (+0.16) | 4.07 (+0.46) | none |
| Semaphore(8).AcquireContext | 1 | 3.74 | 3.78 (+0.04) | 4.28 (+0.53) | none |
| Semaphore(8).AcquireContext | 4 | 3.74 | 3.79 (+0.05) | 4.28 (+0.54) | none |
| Semaphore(8).AcquireContext | 16 | 3.76 | 3.81 (+0.04) | 4.30 (+0.54) | none |
| Semaphore(8).AcquireContext | 32 | 3.76 | 3.83 (+0.07) | 4.33 (+0.56) | none |
| NamedMutex.LockContext | 1 | 6.24 | 7.77 (+1.53) | 8.42 (+2.18) | none |
| NamedMutex.LockContext | 4 | 6.28 | 7.83 (+1.55) | 8.45 (+2.17) | none |
| NamedMutex.LockContext | 16 | 6.27 | 7.84 (+1.57) | 8.41 (+2.14) | none |
| NamedMutex.LockContext | 32 | 6.27 | 7.82 (+1.55) | 8.43 (+2.16) | none |
| TryMutex.LockContext, contended | 1 | 3.64 | 3.93 (+0.29) | 3.95 (+0.31) | none |
| TryMutex.LockContext, contended | 4 | 3.69 | 4.08 (+0.39) | 4.36 (+0.67) | none |
| TryMutex.LockContext, contended | 16 | 3.85 | 4.27 (+0.42) | 4.61 (+0.76) | none |
| TryMutex.LockContext, contended | 32 | 3.99 | 4.33 (+0.34) | 5.06 (+1.07) | none |
