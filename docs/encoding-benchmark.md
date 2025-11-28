# Encoding Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/membership/internal/encoding
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkAppendAddressToBuffer-32                                       246639382                4.801 ns/op           0 B/op          0 allocs/op
BenchmarkAddressFromBuffer-32                                           181789999                6.622 ns/op           0 B/op          0 allocs/op
BenchmarkAppendIncarnationNumberToBuffer-32                             1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkIncarnationNumberFromBuffer-32                                 1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberCountToBuffer-32                                   1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMemberCountFromBuffer-32                                       1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberStateToBuffer-32                                   1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMemberStateFromBuffer-32                                       1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendMemberToBuffer-32                                        135976038                8.822 ns/op           0 B/op          0 allocs/op
BenchmarkMemberFromBuffer-32                                            74119423                16.23 ns/op            0 B/op          0 allocs/op
BenchmarkMessageAlive_AppendToBuffer-32                                 141756174                8.466 ns/op           0 B/op          0 allocs/op
BenchmarkMessageAlive_FromBuffer-32                                     86889590                13.92 ns/op            0 B/op          0 allocs/op
BenchmarkMessageDirectAck_AppendToBuffer-32                             141898882                8.455 ns/op           0 B/op          0 allocs/op
BenchmarkMessageDirectAck_FromBuffer-32                                 86085639                13.84 ns/op            0 B/op          0 allocs/op
BenchmarkMessageDirectPing_AppendToBuffer-32                            141888640                8.455 ns/op           0 B/op          0 allocs/op
BenchmarkMessageDirectPing_FromBuffer-32                                84911056                13.83 ns/op            0 B/op          0 allocs/op
BenchmarkMessageFaulty_AppendToBuffer-32                                83889678                14.28 ns/op            0 B/op          0 allocs/op
BenchmarkMessageFaulty_FromBuffer-32                                    43975212                27.33 ns/op            0 B/op          0 allocs/op
BenchmarkMessageIndirectAck_AppendToBuffer-32                           141830646                8.464 ns/op           0 B/op          0 allocs/op
BenchmarkMessageIndirectAck_FromBuffer-32                               85659016                13.83 ns/op            0 B/op          0 allocs/op
BenchmarkMessageIndirectPing_AppendToBuffer-32                          92083814                13.03 ns/op            0 B/op          0 allocs/op
BenchmarkMessageIndirectPing_FromBuffer-32                              44916490                26.51 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListRequest_AppendToBuffer-32                           147755739                8.118 ns/op           0 B/op          0 allocs/op
BenchmarkMessageListRequest_FromBuffer-32                               87612951                13.86 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/2_members-32                40785277                28.88 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/3_members-32                31509572                38.05 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/4_members-32                24936814                47.48 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/5_members-32                21128031                56.83 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/6_members-32                18140056                66.11 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/7_members-32                15831951                75.62 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/8_members-32                14154776                84.95 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/16_members-32                7246317               161.6 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/32_members-32                3856576               310.5 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/64_members-32                1967973               609.2 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_AppendToBuffer/128_members-32                980091              1208 ns/op               0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/2_members-32                    19916733                59.69 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/3_members-32                    14458402                82.43 ns/op            0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/4_members-32                    11489391               104.3 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/5_members-32                     9467956               126.9 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/6_members-32                     8018354               149.4 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/7_members-32                     6965781               172.4 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/8_members-32                     6119695               195.5 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/16_members-32                    3184021               376.7 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/32_members-32                    1625958               737.8 ns/op             0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/64_members-32                     818203              1461 ns/op               0 B/op          0 allocs/op
BenchmarkMessageListResponse_FromBuffer/128_members-32                    413665              2907 ns/op               0 B/op          0 allocs/op
BenchmarkMessageSuspect_AppendToBuffer-32                               84231344                14.25 ns/op            0 B/op          0 allocs/op
BenchmarkMessageSuspect_FromBuffer-32                                   43984370                27.35 ns/op            0 B/op          0 allocs/op
BenchmarkAppendMessageTypeToBuffer-32                                   1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkMessageTypeFromBuffer-32                                       1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkAppendSequenceNumberToBuffer-32                                1000000000               1.000 ns/op           0 B/op          0 allocs/op
BenchmarkSequenceNumberFromBuffer-32                                    1000000000               1.000 ns/op           0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/membership/internal/encoding      60.272s
```
