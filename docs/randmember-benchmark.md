# Randmember Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/randmember
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkPicker_Pick/1_picks-32                 32535195                36.82 ns/op            0 B/op          0 allocs/op
BenchmarkPicker_Pick/2_picks-32                 19861132                60.13 ns/op            0 B/op          0 allocs/op
BenchmarkPicker_Pick/4_picks-32                 12402576                96.26 ns/op            0 B/op          0 allocs/op
BenchmarkPicker_Pick/8_picks-32                  6953916               173.0 ns/op             0 B/op          0 allocs/op
BenchmarkPicker_Pick/16_picks-32                 3554724               338.3 ns/op             0 B/op          0 allocs/op
BenchmarkPicker_Pick/32_picks-32                 1635343               733.7 ns/op             0 B/op          0 allocs/op
BenchmarkPicker_PickWithout/1_picks-32          18074688                66.35 ns/op            0 B/op          0 allocs/op
BenchmarkPicker_PickWithout/2_picks-32          13538220                88.35 ns/op            0 B/op          0 allocs/op
BenchmarkPicker_PickWithout/4_picks-32           8784054               136.3 ns/op             0 B/op          0 allocs/op
BenchmarkPicker_PickWithout/8_picks-32           5137761               227.1 ns/op             0 B/op          0 allocs/op
BenchmarkPicker_PickWithout/16_picks-32          2854844               418.6 ns/op             0 B/op          0 allocs/op
BenchmarkPicker_PickWithout/32_picks-32          1395978               860.8 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/membership/internal/randmember    14.359s
```
