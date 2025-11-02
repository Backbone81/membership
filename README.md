# Membership

This Go library provides a peer-to-peer gossip based membership implementation. It implements
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914).

## TODOs

### Basic Requirements

- Add metrics to expose what is happening.
- Add encryption and support multiple encryption keys for key rollover. The first key is always used for encryption, all
  keys are used for decryption.
- Shutdown of the membership should send a faulty message to propagate the not existing member.
- Improve test coverage
- Check where we can replace a slices.Delete with a switch-with-last and reduce size by one.
- The performance characteristics of the gossip queue is suboptimal. When creating a cluster with 16k members, the
  creating of the cluster alone without any gossip takes excessive amount of time. Look into a ringbuffer
  implementation.

### More Advanced Topics

- How should we deal with sequence number wrap-around?
- How should we deal with incarnation number wrap-around?
- Provide an auto round-trip timeout which is derived from the 99th percentile of past network messages and use +10%
- Support more than one direct probes during the protocol period
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
- Regularly log the statistical information about ping chance and time until all know about gossip
