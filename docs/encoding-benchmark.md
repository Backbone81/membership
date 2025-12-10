# Encoding Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/encoding
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkAppendAddressToBuffer-32                                       248878304                4.782 ns/op           0 B/op          0 allocs/op
BenchmarkAddressFromBuffer-32                                           179157482                6.694 ns/op           0 B/op          0 allocs/op
BenchmarkAppendIncarnationNumberToBuffer-32                             1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkIncarnationNumberFromBuffer-32                                 1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberCountToBuffer-32                                   1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMemberCountFromBuffer-32                                       1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberStateToBuffer-32                                   1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMemberStateFromBuffer-32                                       1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberToBuffer-32                                        133822360                8.971 ns/op           0 B/op          0 allocs/op
BenchmarkMemberFromBuffer-32                                            74581856                15.89 ns/op            0 B/op          0 allocs/op
BenchmarkMessageAlive_AppendToBuffer-32                                 141921948                8.450 ns/op           0 B/op          0 allocs/op
BenchmarkMessageAlive_FromBuffer-32                                     86402730                13.89 ns/op            0 B/op          0 allocs/op
BenchmarkMessageDirectAck_AppendToBuffer-32                             141927057                8.454 ns/op           0 B/op          0 allocs/op
BenchmarkMessageDirectAck_FromBuffer-32                                 86445340                13.89 ns/op            0 B/op          0 allocs/op
BenchmarkMessageDirectPing_AppendToBuffer-32                            139884799                8.569 ns/op           0 B/op          0 allocs/op
BenchmarkMessageDirectPing_FromBuffer-32                                85722416                13.93 ns/op            0 B/op          0 allocs/op
BenchmarkMessageFaulty_AppendToBuffer-32                                84010657                14.28 ns/op            0 B/op          0 allocs/op
BenchmarkMessageFaulty_FromBuffer-32                                    44066404                27.30 ns/op            0 B/op          0 allocs/op
BenchmarkMessageIndirectAck_AppendToBuffer-32                           141480243                8.477 ns/op           0 B/op          0 allocs/op
BenchmarkMessageIndirectAck_FromBuffer-32                               86704474                13.87 ns/op            0 B/op          0 allocs/op
BenchmarkMessageIndirectPing_AppendToBuffer-32                          92195348                13.03 ns/op            0 B/op          0 allocs/op
BenchmarkMessageIndirectPing_FromBuffer-32                              44985169                26.75 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListRequest_AppendToBuffer-32                           148179265                8.099 ns/op           0 B/op          0 allocs/op
BenchmarkMessageListRequest_FromBuffer-32                               86616576                13.70 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/2_members-32                42762835                27.79 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/3_members-32                32119588                37.15 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/4_members-32                25850833                46.47 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/5_members-32                21515163                55.76 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/6_members-32                18412543                65.16 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/7_members-32                16064851                74.45 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/8_members-32                14104060                83.90 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/16_members-32                7556377               158.8 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/32_members-32                3854968               309.1 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/64_members-32                1975185               608.3 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/128_members-32                992503              1206 ns/op               0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/2_members-32                    20838411                57.52 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/3_members-32                    14730141                79.80 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/4_members-32                    11792217               101.5 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/5_members-32                     9707992               123.3 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/6_members-32                     8259940               145.4 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/7_members-32                     7085450               168.7 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/8_members-32                     6307464               190.3 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/16_members-32                    3270093               366.2 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/32_members-32                    1676335               715.6 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/64_members-32                     839869              1414 ns/op               0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/128_members-32                    425488              2816 ns/op               0 B/op          0 allocs/op
BenchmarkMessageSuspect_AppendToBuffer-32                               84130569                14.26 ns/op            0 B/op          0 allocs/op
BenchmarkMessageSuspect_FromBuffer-32                                   43506484                27.61 ns/op            0 B/op          0 allocs/op
BenchmarkAppendMessageTypeToBuffer-32                                   1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMessageTypeFromBuffer-32                                       1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendSequenceNumberToBuffer-32                                1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkSequenceNumberFromBuffer-32                                    1000000000               1.000 ns/op           0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/membership/internal/encoding      60.298s
```
