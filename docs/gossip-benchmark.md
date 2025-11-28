# Gossip Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/gossip
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkQueue_Add/1024_gossip_in_8_buckets-32                                   6685018               172.9 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/1024_gossip_in_16_buckets-32                                  6489988               173.7 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/1024_gossip_in_32_buckets-32                                  7182207               162.0 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/2048_gossip_in_8_buckets-32                                   7255778               161.0 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/2048_gossip_in_16_buckets-32                                  7148382               161.6 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/2048_gossip_in_32_buckets-32                                  7209590               161.6 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/4096_gossip_in_8_buckets-32                                   7260692               163.2 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/4096_gossip_in_16_buckets-32                                  7264878               162.2 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/4096_gossip_in_32_buckets-32                                  6968998               161.3 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/8192_gossip_in_8_buckets-32                                   7229366               161.5 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/8192_gossip_in_16_buckets-32                                  7290273               162.5 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/8192_gossip_in_32_buckets-32                                  6984379               161.9 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/16384_gossip_in_8_buckets-32                                  6962452               163.5 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/16384_gossip_in_16_buckets-32                                 7197718               161.8 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Add/16384_gossip_in_32_buckets-32                                 7276788               162.1 ns/op             0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/1024_gossip-32                                        81267314                14.70 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/2048_gossip-32                                        80323209                14.96 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/4096_gossip-32                                        81679543                14.68 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/8192_gossip-32                                        81409551                14.71 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_Prioritize/16384_gossip-32                                       80152482                14.98 ns/op            0 B/op          0 allocs/op
BenchmarkQueue_All/1024_gossip-32                                               319066978                3.765 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/2048_gossip-32                                               318831739                3.763 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/4096_gossip-32                                               320997165                3.744 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/8192_gossip-32                                               318942589                3.747 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_All/16384_gossip-32                                              320238762                3.786 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_1_transmissions-32              567073588                2.123 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_2_transmissions-32              522351500                2.151 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_4_transmissions-32              564890926                2.122 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_8_transmissions-32              567249428                2.113 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_16_transmissions-32             569485064                2.114 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_32_transmissions-32             568852888                2.114 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_64_transmissions-32             568205282                2.120 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/1024_gossip_with_128_transmissions-32            566416254                2.116 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_1_transmissions-32              566838638                2.137 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_2_transmissions-32              564214765                2.119 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_4_transmissions-32              563004093                2.125 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_8_transmissions-32              560831502                2.130 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_16_transmissions-32             554454888                2.129 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_32_transmissions-32             540690927                2.136 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_64_transmissions-32             566088296                2.120 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/2048_gossip_with_128_transmissions-32            565479004                2.119 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_1_transmissions-32              566165696                2.114 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_2_transmissions-32              567591742                2.124 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_4_transmissions-32              565563016                2.123 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_8_transmissions-32              562217380                2.127 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_16_transmissions-32             552817426                2.126 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_32_transmissions-32             568134696                2.115 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_64_transmissions-32             567930546                2.117 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/4096_gossip_with_128_transmissions-32            562762462                2.121 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_1_transmissions-32              563898556                2.124 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_2_transmissions-32              557120583                2.135 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_4_transmissions-32              566585190                2.121 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_8_transmissions-32              566441110                2.125 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_16_transmissions-32             564345312                2.120 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_32_transmissions-32             565348971                2.116 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_64_transmissions-32             566531116                2.113 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/8192_gossip_with_128_transmissions-32            568774756                2.117 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_1_transmissions-32             564842871                2.114 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_2_transmissions-32             565149639                2.133 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_4_transmissions-32             557045181                2.130 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_8_transmissions-32             567777402                2.115 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_16_transmissions-32            565653154                2.122 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_32_transmissions-32            567277590                2.119 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_64_transmissions-32            562478145                2.118 ns/op           0 B/op          0 allocs/op
BenchmarkQueue_MarkTransmitted/16384_gossip_with_128_transmissions-32           565663503                2.126 ns/op           0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/membership/internal/gossip        90.039s
```
