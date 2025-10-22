# Encoding Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/encoding
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkAppendAddressToBuffer-32               253972956                4.712 ns/op           0 B/op          0 allocs/op
BenchmarkAddressFromBuffer-32                   177514814                6.762 ns/op           0 B/op          0 allocs/op
BenchmarkAppendIncarnationNumberToBuffer-32     1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkIncarnationNumberFromBuffer-32         1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberCountToBuffer-32           1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMemberCountFromBuffer-32               1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberStateToBuffer-32           1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMemberStateFromBuffer-32               1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberToBuffer-32                136648677                8.778 ns/op           0 B/op          0 allocs/op
BenchmarkMemberFromBuffer-32                    70252678                16.41 ns/op            0 B/op          0 allocs/op
BenchmarkAppendMessageTypeToBuffer-32           1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMessageTypeFromBuffer-32               1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendSequenceNumberToBuffer-32        1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkSequenceNumberFromBuffer-32            1000000000               1.000 ns/op           0 B/op          0 allocs/op
```
