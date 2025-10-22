# Encryption Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/encryption
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkEncrypt/8_bytes-32              1000000              2326 ns/op            1280 B/op          2 allocs/op
BenchmarkEncrypt/16_bytes-32              706916              1858 ns/op            1280 B/op          2 allocs/op
BenchmarkEncrypt/32_bytes-32              581776              1746 ns/op            1280 B/op          2 allocs/op
BenchmarkEncrypt/64_bytes-32             1000000              2237 ns/op            1280 B/op          2 allocs/op
BenchmarkEncrypt/128_bytes-32             699327              2325 ns/op            1280 B/op          2 allocs/op
BenchmarkEncrypt/256_bytes-32             485745              2474 ns/op            1280 B/op          2 allocs/op
BenchmarkEncrypt/512_bytes-32             593668              2518 ns/op            1280 B/op          2 allocs/op
BenchmarkEncrypt/1024_bytes-32            643186              2060 ns/op            1280 B/op          2 allocs/op
BenchmarkDecrypt/8_bytes-32              2176633               929.7 ns/op          1280 B/op          2 allocs/op
BenchmarkDecrypt/16_bytes-32             4131478               688.9 ns/op          1280 B/op          2 allocs/op
BenchmarkDecrypt/32_bytes-32              895046              1345 ns/op            1280 B/op          2 allocs/op
BenchmarkDecrypt/64_bytes-32              926905              1235 ns/op            1280 B/op          2 allocs/op
BenchmarkDecrypt/128_bytes-32             856802              1169 ns/op            1280 B/op          2 allocs/op
BenchmarkDecrypt/256_bytes-32             730240              1506 ns/op            1280 B/op          2 allocs/op
BenchmarkDecrypt/512_bytes-32             919022              1331 ns/op            1280 B/op          2 allocs/op
BenchmarkDecrypt/1024_bytes-32            571034              1860 ns/op            1280 B/op          2 allocs/op
BenchmarkNewRandomKey-32                 5064688               213.5 ns/op             0 B/op          0 allocs/op
BenchmarkParseKeyFromHexString-32       62111334                18.89 ns/op            0 B/op          0 allocs/op
```
