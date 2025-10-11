# Membership

This Go library provides a peer-to-peer gossip based membership implementation. It implements
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914).

## TODOs

- Should the FromBuffer functions return the remaining buffer to make it easier and less error-prone to work with?
- How should we deal with sequence number wrap-around?
- How should we deal with incarnation number wrap-around?
- Address should be serialized to buffer with variable length instead of ipv6 length all the time.
- Check if we really need to use panic anywhere.
- Support different settings for max datagram length for receive and send

- Provide a fixed round-trip timeout
- Provide an auto round-trip timeout which is derived from the 99th percentile of past network messages and use +10%
