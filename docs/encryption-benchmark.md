# Encryption Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/encryption
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkNewRandomKey-32                 5762476               208.2 ns/op             0 B/op          0 allocs/op
BenchmarkParseKeyFromHexString-32       56426367                21.51 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/membership/internal/encryption    2.419s
```
