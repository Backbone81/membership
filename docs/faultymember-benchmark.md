# Faultymember Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/faultymember
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkList_Add/1024_members_in_8_buckets-32           9002352               136.3 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/1024_members_in_16_buckets-32          8885200               136.5 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/1024_members_in_32_buckets-32          9465667               133.4 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/2048_members_in_8_buckets-32           9377200               130.8 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/2048_members_in_16_buckets-32          9446826               131.0 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/2048_members_in_32_buckets-32          9442651               131.0 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/4096_members_in_8_buckets-32           9350275               131.8 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/4096_members_in_16_buckets-32          9365809               131.3 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/4096_members_in_32_buckets-32          9300632               130.5 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/8192_members_in_8_buckets-32           9402187               131.0 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/8192_members_in_16_buckets-32          9430659               130.5 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/8192_members_in_32_buckets-32          8579780               130.0 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/16384_members_in_8_buckets-32          9406464               132.2 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/16384_members_in_16_buckets-32         9331982               136.4 ns/op             0 B/op          0 allocs/op
BenchmarkList_Add/16384_members_in_32_buckets-32         9297165               131.0 ns/op             0 B/op          0 allocs/op
BenchmarkList_ForEach/1024_members-32                   323862840                3.661 ns/op           0 B/op          0 allocs/op
BenchmarkList_ForEach/2048_members-32                   324448695                3.685 ns/op           0 B/op          0 allocs/op
BenchmarkList_ForEach/4096_members-32                   324422773                3.704 ns/op           0 B/op          0 allocs/op
BenchmarkList_ForEach/8192_members-32                   324019576                3.694 ns/op           0 B/op          0 allocs/op
BenchmarkList_ForEach/16384_members-32                  321798825                3.707 ns/op           0 B/op          0 allocs/op
BenchmarkList_ListRequestObserved/1024_members-32       355923963                3.392 ns/op           0 B/op          0 allocs/op
BenchmarkList_ListRequestObserved/2048_members-32       349599511                3.380 ns/op           0 B/op          0 allocs/op
BenchmarkList_ListRequestObserved/4096_members-32       355463512                3.364 ns/op           0 B/op          0 allocs/op
BenchmarkList_ListRequestObserved/8192_members-32       346550115                3.408 ns/op           0 B/op          0 allocs/op
BenchmarkList_ListRequestObserved/16384_members-32      358661935                3.359 ns/op           0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/membership/internal/faultymember  41.548s
```
