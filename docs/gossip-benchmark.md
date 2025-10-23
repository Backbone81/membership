# Gossip Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/gossip
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkMessageAlive_AppendToBuffer-32                 195944264                6.004 ns/op           0 B/op          0 allocs/op
BenchmarkMessageAlive_FromBuffer-32                     87454849                13.48 ns/op            0 B/op          0 allocs/op
BenchmarkMessageFaulty_AppendToBuffer-32                100000000               11.01 ns/op            0 B/op          0 allocs/op
BenchmarkMessageFaulty_FromBuffer-32                    44865614                26.62 ns/op            0 B/op          0 allocs/op
BenchmarkGossipQueue_Add/1_gossip-32                    21028182               125.3 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/2_gossip-32                     9815571               122.8 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/4_gossip-32                    10432384               122.6 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/8_gossip-32                    15803245               122.3 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/16_gossip-32                    8373177               120.5 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/32_gossip-32                   13992217               123.8 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/64_gossip-32                    8918554               132.2 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/128_gossip-32                  18152010               118.6 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/256_gossip-32                  15119942               121.7 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/512_gossip-32                   9916072               119.4 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/1024_gossip-32                  7692504               150.5 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/2048_gossip-32                  7830708               148.9 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/4096_gossip-32                  8301060               145.8 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/8192_gossip-32                  9628671               143.2 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_Add/16384_gossip-32                 8002975               131.0 ns/op            48 B/op          2 allocs/op
BenchmarkGossipQueue_PrepareFor/1_gossip-32             23859262                47.23 ns/op            0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/2_gossip-32             15158020                78.24 ns/op            0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/4_gossip-32              8817129               136.1 ns/op             0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/8_gossip-32              4575602               258.5 ns/op             0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/16_gossip-32             1928830               619.6 ns/op             0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/32_gossip-32              955029              1246 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/64_gossip-32              480867              2469 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/128_gossip-32             234405              5088 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/256_gossip-32             118687             10003 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/512_gossip-32              58227             20236 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/1024_gossip-32             24628             48538 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/2048_gossip-32             10000            101877 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/4096_gossip-32              5601            208958 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/8192_gossip-32              2768            431843 ns/op               0 B/op          0 allocs/op
BenchmarkGossipQueue_PrepareFor/16384_gossip-32             1254            908224 ns/op               0 B/op          0 allocs/op
BenchmarkMessageSuspect_AppendToBuffer-32               100000000               11.01 ns/op            0 B/op          0 allocs/op
BenchmarkMessageSuspect_FromBuffer-32                   39081330                30.60 ns/op            0 B/op          0 allocs/op
```
