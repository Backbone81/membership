# Gossip Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/gossip
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkQueue_Add/1024_gossip_in_8_buckets-32                                   6501876               174.8 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/1024_gossip_in_16_buckets-32                                  6580299               174.2 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/1024_gossip_in_32_buckets-32                                  6999284               160.2 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/2048_gossip_in_8_buckets-32                                   7227819               161.9 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/2048_gossip_in_16_buckets-32                                  7178751               165.4 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/2048_gossip_in_32_buckets-32                                  7246663               162.9 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/4096_gossip_in_8_buckets-32                                   6969194               161.3 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/4096_gossip_in_16_buckets-32                                  6984096               164.5 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/4096_gossip_in_32_buckets-32                                  7294213               161.8 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/8192_gossip_in_8_buckets-32                                   7217012               165.9 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/8192_gossip_in_16_buckets-32                                  7164850               161.6 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/8192_gossip_in_32_buckets-32                                  7107538               161.3 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/16384_gossip_in_8_buckets-32                                  7180933               162.5 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/16384_gossip_in_16_buckets-32                                 7217600               162.5 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/16384_gossip_in_32_buckets-32                                 7056691               161.9 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/1024_gossip-32                                        81292862                14.75 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/2048_gossip-32                                        81670976                14.69 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/4096_gossip-32                                        79767609                14.96 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/8192_gossip-32                                        81770872                14.70 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/16384_gossip-32                                       80708756                14.68 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_All/1024_gossip-32                                               319854372                3.773 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/2048_gossip-32                                               320446406                3.782 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/4096_gossip-32                                               317746539                3.748 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/8192_gossip-32                                               320883574                3.740 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/16384_gossip-32                                              322705090                3.742 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_1_transmissions-32              566925696                2.114 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_2_transmissions-32              569903257                2.107 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_4_transmissions-32              572014539                2.115 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_8_transmissions-32              569836707                2.099 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_16_transmissions-32             568460600                2.106 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_32_transmissions-32             569782796                2.100 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_64_transmissions-32             560902519                2.108 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_128_transmissions-32            567143391                2.110 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_1_transmissions-32              565440645                2.104 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_2_transmissions-32              565340667                2.111 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_4_transmissions-32              561863371                2.115 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_8_transmissions-32              565931145                2.116 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_16_transmissions-32             563024658                2.116 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_32_transmissions-32             569353472                2.110 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_64_transmissions-32             579261272                2.094 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_128_transmissions-32            569386639                2.105 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_1_transmissions-32              568301664                2.108 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_2_transmissions-32              565245088                2.114 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_4_transmissions-32              550183300                2.131 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_8_transmissions-32              560949558                2.118 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_16_transmissions-32             569029444                2.115 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_32_transmissions-32             569379957                2.108 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_64_transmissions-32             561569356                2.113 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_128_transmissions-32            562759040                2.111 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_1_transmissions-32              562182457                2.100 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_2_transmissions-32              568844710                2.096 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_4_transmissions-32              566588854                2.109 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_8_transmissions-32              566439648                2.098 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_16_transmissions-32             522854254                2.141 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_32_transmissions-32             567728722                2.094 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_64_transmissions-32             569188863                2.102 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_128_transmissions-32            565979196                2.102 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_1_transmissions-32             561077253                2.114 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_2_transmissions-32             565095321                2.109 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_4_transmissions-32             551565818                2.122 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_8_transmissions-32             567641235                2.113 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_16_transmissions-32            566353126                2.107 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_32_transmissions-32            532797979                2.115 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_64_transmissions-32            567934384                2.113 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_128_transmissions-32           573518916                2.094 ns/op           0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/membership/internal/gossip        89.773s
```
