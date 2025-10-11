# Membership

This Go library provides a peer-to-peer gossip based membership implementation. It implements
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914).

## TODOs

- Should the FromBuffer functions return the remaining buffer to make it easier and less error-prone to work with?
- Address should be serialized to buffer with variable length instead of ipv6 length all the time.
- Support different settings for max datagram length for receive and send
- Member.LastStateChange is only needed to mark a suspect as faulty. We should get rid of it. We also don't want to
  have timeouts in the core logic of the implementation to have the possibility to drive tests without having to wait
  for timeouts. We might need to replace the LastStateChange with a period counter.
- Make sure we provide enough context for all error returns.
- Have the members and faulty members always sorted by address and use binary search to find members in them.
 
- Provide a fixed round-trip timeout
- Provide an auto round-trip timeout which is derived from the 99th percentile of past network messages and use +10%

- Provide a function which returns the full list of members
- Provide callbacks when members change
- Simplify the usage of the membership library by consolidating the individual components for users.

- How should we deal with sequence number wrap-around?
- How should we deal with incarnation number wrap-around?
- Check if we really need to use panic anywhere.
