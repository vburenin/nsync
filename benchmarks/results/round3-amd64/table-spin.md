| Workload | Ps | Default ns/op | nsync_spin ns/op | Default/spin |
| --- | ---: | ---: | ---: | ---: |
| TryMutex/Uncontended | 1 | 3.59 | 3.58 | 1.00× |
| TryMutex/Uncontended | 4 | 3.65 | 3.60 | 1.01× |
| TryMutex/Uncontended | 16 | 3.65 | 3.62 | 1.01× |
| TryMutex/Uncontended | 32 | 3.61 | 3.59 | 1.01× |
| TryMutex/Contended0 | 1 | 3.64 | 3.65 | 1.00× |
| TryMutex/Contended0 | 4 | 3.75 | 5.40 | 0.69× |
| TryMutex/Contended0 | 16 | 3.86 | 5.20 | 0.74× |
| TryMutex/Contended0 | 32 | 3.96 | 5.20 | 0.76× |
| TryMutex/Contended100 | 1 | 81.07 | 80.77 | 1.00× |
| TryMutex/Contended100 | 4 | 84.84 | 87.00 | 0.98× |
| TryMutex/Contended100 | 16 | 87.76 | 92.62 | 0.95× |
| TryMutex/Contended100 | 32 | 90.19 | 102.1 | 0.88× |
