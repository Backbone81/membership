# Membership

This Go library provides a peer-to-peer gossip based membership implementation. It implements
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914).

## TODOs

### Basic Requirements

- Provide a function which returns the full list of members
- Provide callbacks when members change
- Address should be serialized to buffer with variable length instead of ipv6 length all the time.
- Support different settings for max datagram length for receive and send
- Have the members and faulty members always sorted by address and use binary search to find members in them.
- Simplify the usage of the membership library by consolidating the individual components for users.
- Add metrics to expose what is happening.
- Add encryption and support multiple encryption keys for key rollover. The first key is always used for encryption, all
  keys are used for decryption.
- Shutdown of the membership should send a faulty message to propagate the not existing member.
- The number a gossip is gossiped needs to be dynamically adjusted to the size of the member cluster.

### More Advanced Topics

- How should we deal with sequence number wrap-around?
- How should we deal with incarnation number wrap-around?
- Provide an auto round-trip timeout which is derived from the 99th percentile of past network messages and use +10%
- Support more than one direct probes during the protocol period
- Member.LastStateChange is only needed to mark a suspect as faulty. We should get rid of it. We also don't want to
  have timeouts in the core logic of the implementation to have the possibility to drive tests without having to wait
  for timeouts. We might need to replace the LastStateChange with a period counter.
- Can we drop the interface for Message and collapse all message types into a single message which has all fields any
  message could have? That way we could be able to drop the interface and prevent memory allocations due to interface
  conversion.
- We should find a mechanic which tells a member the last known incarnation number to allow joining members without
  having to remember the incarnation number.

### Nice to Have

- Should the FromBuffer functions return the remaining buffer to make it easier and less error-prone to work with?
- Make sure we provide enough context for all error returns.
- Check if we really need to use panic anywhere.
- The tests for messages are very repeating. Look into re-using code.
- The message setup in the message tests are quite repetitive. We should consolidate them.
- Introduce jitter into the scheduler to avoid spikes in network traffic.
