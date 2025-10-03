# Membership

This Go library provides a peer-to-peer gossip based membership implementation. It implements
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914).

## TODOs

- Provide a fixed round-trip timeout
- Provide an auto round-trip timeout which is derived from the 99th percentile of past network messages and use +10%