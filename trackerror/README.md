# If you need to use concurrent map access, just use sync.Map

```bash
(go test -bench=. -benchmem -cpu=1,2,4)
  ⎿  goos: darwin                                                                                                 
     goarch: amd64
     pkg: github.com/anngdinh/operator-helper/trackerror
     cpu: VirtualApple @ 2.50GHz
     BenchmarkTrackError_Store                           2608353               424.0 ns/op            93 B/op          2 allocs/op
     BenchmarkTrackError_Store-2                         3137593               382.1 ns/op            84 B/op          2 allocs/op
     BenchmarkTrackError_Store-4                         3048278               378.1 ns/op            85 B/op          2 allocs/op
     BenchmarkSyncMapTrackError_Store                    1000000              1027 ns/op             187 B/op          5 allocs/op
     BenchmarkSyncMapTrackError_Store-2                  1866268               862.2 ns/op           196 B/op          5 allocs/op
     BenchmarkSyncMapTrackError_Store-4                  1810501               730.9 ns/op           200 B/op          5 allocs/op
     BenchmarkTrackError_Load                           10729756               120.7 ns/op            13 B/op          1 allocs/op
     BenchmarkTrackError_Load-2                         11124535               111.9 ns/op            13 B/op          1 allocs/op
     BenchmarkTrackError_Load-4                         11133829               112.8 ns/op            13 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_Load                     8705527               152.9 ns/op            13 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_Load-2                   8843690               190.5 ns/op            13 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_Load-4                   7205382               188.5 ns/op            13 B/op          1 allocs/op
     BenchmarkTrackError_Delete                          4034198               614.1 ns/op            23 B/op          1 allocs/op
     BenchmarkTrackError_Delete-2                        4778845               288.9 ns/op            23 B/op          1 allocs/op
     BenchmarkTrackError_Delete-4                        4717842               347.8 ns/op            23 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_Delete                   3573387               668.7 ns/op            23 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_Delete-2                 3850700               312.9 ns/op            23 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_Delete-4                 3987849               464.6 ns/op            23 B/op          1 allocs/op
     BenchmarkTrackError_Count                          77518753                14.62 ns/op            0 B/op          0 allocs/op
     BenchmarkTrackError_Count-2                        84576776                14.53 ns/op            0 B/op          0 allocs/op
     BenchmarkTrackError_Count-4                        85202818                15.11 ns/op            0 B/op          0 allocs/op
     BenchmarkSyncMapTrackError_Count                   1000000000               0.4471 ns/op          0 B/op          0 allocs/op
     BenchmarkSyncMapTrackError_Count-2                 1000000000               0.3215 ns/op          0 B/op          0 allocs/op
     BenchmarkSyncMapTrackError_Count-4                 1000000000               0.3265 ns/op          0 B/op          0 allocs/op
     BenchmarkTrackError_ConcurrentStore                 2334339               481.5 ns/op           101 B/op          2 allocs/op
     BenchmarkTrackError_ConcurrentStore-2               3652642               416.9 ns/op            73 B/op          2 allocs/op
     BenchmarkTrackError_ConcurrentStore-4               2575492               406.9 ns/op            41 B/op          2 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentStore          1000000              1034 ns/op             187 B/op          5 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentStore-2        1877722               711.6 ns/op           121 B/op          4 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentStore-4        2755119               539.9 ns/op            93 B/op          3 allocs/op
     BenchmarkTrackError_ConcurrentLoad                  8840262               139.3 ns/op            15 B/op          1 allocs/op
     BenchmarkTrackError_ConcurrentLoad-2               14844824                80.27 ns/op           15 B/op          1 allocs/op
     BenchmarkTrackError_ConcurrentLoad-4               16093286                77.00 ns/op           15 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentLoad           7482520               200.7 ns/op            15 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentLoad-2        13655973                93.42 ns/op           15 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentLoad-4        25354812                49.06 ns/op           15 B/op          1 allocs/op
     BenchmarkTrackError_ConcurrentMixed                10620259               149.6 ns/op            13 B/op          1 allocs/op
     BenchmarkTrackError_ConcurrentMixed-2               9805290               121.2 ns/op            13 B/op          1 allocs/op
     BenchmarkTrackError_ConcurrentMixed-4               8725561               166.9 ns/op            13 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentMixed          7761699               158.2 ns/op            17 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentMixed-2       15952099               118.1 ns/op            17 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_ConcurrentMixed-4       22178269                51.02 ns/op           17 B/op          1 allocs/op
     BenchmarkTrackError_HighContention                  6752445               214.2 ns/op            16 B/op          1 allocs/op
     BenchmarkTrackError_HighContention-2                4391556               275.4 ns/op            16 B/op          1 allocs/op
     BenchmarkTrackError_HighContention-4                4108689               298.6 ns/op            16 B/op          1 allocs/op
     BenchmarkSyncMapTrackError_HighContention           3520459               334.1 ns/op            54 B/op          2 allocs/op
     BenchmarkSyncMapTrackError_HighContention-2         8722540               132.5 ns/op            35 B/op          2 allocs/op
     BenchmarkSyncMapTrackError_HighContention-4        16110147                75.87 ns/op           35 B/op          2 allocs/op
     PASS
     ok         github.com/anngdinh/operator-helper/trackerror  110.812s
```

## Performance Comparison Results

  Key Findings:

  🏆 Winner by Operation:

- Store: Original (2-3x faster, less memory)
- Load: Original (1.3x faster)
- Delete: Similar performance
- Count: sync.Map (30x faster!)
- Concurrent Load: sync.Map (better scaling)
- High Contention: sync.Map (4x faster with 4 CPUs)

  📊 Detailed Analysis:

  Single-threaded operations:

- Original implementation excels at Store/Load due to simpler map operations
- sync.Map has overhead from internal complexity but shines at Count (atomic vs mutex)

  Concurrent operations:

- sync.Map scales much better with CPU count
- At 4 CPUs: sync.Map shows 2-4x performance improvement
- Original implementation suffers from mutex contention

  Memory usage:

- Original: 84-93 B/op for Store, 0 B for Count
- sync.Map: 187-200 B/op for Store, 0 B for Count

  💡 Recommendation:

  Use sync.Map implementation for high-concurrency scenarios, original implementation for single-threaded or low-contention use cases.
