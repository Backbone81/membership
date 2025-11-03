# Membership

This Go library provides a peer-to-peer gossip based membership implementation. It implements
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914).

## TODOs

### Basic Requirements

- Improve test coverage
- The performance characteristics of the gossip queue is suboptimal. When creating a cluster with 16k members, the
  creating of the cluster alone without any gossip takes excessive amount of time. Look into a ringbuffer
  implementation.
- Replicate the benchmarks about reliability of the protocol from the SWIM paper and papers improving it to have a
  comparison between implementations.
  - Parameters for SWIM benchmarks: indirect pings K=1, protocol period 2 sec, dissemination periods 3 * log(N+1) for
    gossip and suspect timeout
  - Message Load: Messages send and received for any group member, measured over 40 protocol periods. up to a group size
    of 55 members. Expectation: 2.0. Note that send and receive are not handled as separate messages.
  - Average first detection: group size to protocol periods until anyone detects an outage.
  - Latency of spread: Group size to protocol periods until infection happens. calculate median infection time
  - Suspicion time-out: Group size to protocol periods
  - Failure Detection False Positives: 10% packet drop, add 17 processes one after the other. Protocol periods to group
    size (87 protocol periods).
- Provide an auto round-trip timeout which is derived from the 99th percentile of past network messages and use +10%
- Support more than one direct probes during the protocol period

### More Advanced Topics

- How should we deal with sequence number wrap-around?
- How should we deal with incarnation number wrap-around?
- We should find a mechanic which tells a member the last known incarnation number to allow joining members without
  having to remember the incarnation number.
- Add encryption and support multiple encryption keys for key rollover. The first key is always used for encryption, all
  keys are used for decryption.
- Extend metrics for better insights.
- Can we drop the interface for Message and collapse all message types into a single message which has all fields any
  message could have? That way we could be able to drop the interface and prevent memory allocations due to interface
  conversion.

### Nice to Have

- Should the FromBuffer functions return the remaining buffer to make it easier and less error-prone to work with?
- Make sure we provide enough context for all error returns.
- Check if we really need to use panic anywhere.
- The tests for messages are very repeating. Look into re-using code.
- The message setup in the message tests are quite repetitive. We should consolidate them.
- Introduce jitter into the scheduler to avoid spikes in network traffic.
- Regularly log the statistical information about ping chance and time until all know about gossip
