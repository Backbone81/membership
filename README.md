# Membership

This Go library provides a membership list for distributed processes based on the
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914)
with improvements from
[Lifeguard: Local Health Awareness for More Accurate Failure Detection](https://doi.org/10.48550/arXiv.1707.00788).

## Features

- Peer-to-peer without a single point of failure.
- Joining the cluster requires knowledge of at least one other cluster member.
- Eventual consistency through gossip based propagation of changes.
- Constant low CPU and network load for each member independent of the cluster size (except for the periodic full list
  sync).
- Pings and their responses with gossip piggybacked are sent as UDP messages.
- Each member will ping some other member every n-1 protocol periods on average and every 2n-1 protocol periods in the
  worst case, where n is the size of the cluster.
- The timeout for a direct ping which triggers an indirect ping is dynamically adjusted according to the round trip
  times of past network messages.
- A periodic full synchronization between two members is done over TCP to address situations where bad luck resulted in
  an inconsistent state which would not self-heal anymore.
- Failure detection at any member happens in constant time.
- Dissemination of changes to the whole cluster happens in logarithmic time on the cluster size.
- Changes which were gossiped the least are prioritized over changes which were gossipped more when deciding on what
  to gossip next.
- Gossip about a member is always gossipped to that member with priority to allow quicker corrections of false suspects.
- A graceful shutdown will propagate the failure of the node shutting down, reducing the detection time.

## Mechanic

The library works in cycles called protocol periods. One protocol period is usually one second.

Each protocol period starts out with picking one other known member at random and sending a ping message. Once the
destination receives that ping message, it responds with an acknowledgment message.

If the destination does not respond within some timeout, some other member is randomly chosen and sent a message which
requests that member to ping the not responding member. If that indirect ping is answered until the end of the protocol
period, the member stays alive. If no answer is received, the member is marked as being suspected to have failed.

If the member stays suspect for a few protocol periods, it is declared failed and removed from the membership list.

All changes about members being alive, suspect or faulty are piggybacked on the ping and acknowledge messages which
disseminates the changes throughout the cluster. After some maximum number of times of having been disseminated, the
gossip is dropped.

Situations can arise where some member might not receive a specific change and the change is not gossiped any more by
other members. To tackle this issue, there is a full exchange of membership list at a low frequency. A random member
is picked and the full membership list is requested.

## TODOs

### Basic Requirements

- Improve test coverage
- Cleanup the faulty member list after some time to avoid endless growth in situations where members are very dynamic.
- Serialize the current state on shutdown and allow that state to be re-used during startup.
- We might want to separate gossip count from suspect timeout
- We should find a mechanic which tells a member the last known incarnation number to allow joining members without
  having to remember the incarnation number. This could be done with the full list sync. We would need to update our
  own incarnation number when we learn about ourselves.
- Extend metrics for better insights.

### More Advanced Topics

- Can we drop the interface for Message and collapse all message types into a single message which has all fields any
  message could have? That way we could be able to drop the interface and prevent memory allocations due to interface
  conversion.
- Add encryption and support multiple encryption keys for key rollover. The first key is always used for encryption, all
  keys are used for decryption.
- Investigate how we can increase the suspicion timeout when we are under high CPU load. High CPU load can be detected
  by the scheduler as the times between direct pings, indirect pings and end of protocol are either significant shorter
  than expected or even overshot immediately.
- What should we do when a large amount of gossip is piling up in the gossip queue? We should have a way to speed up
  dissemination to quicker reach a stable state again. We could increase the number of direct pings dynamically when the
  number of gossip in the gossip queue is larger than what can usually be piggybacked in one ping.
- How can a member re-join when it was disconnected through a network partition from everybody else for a long time?
  We probably need to deal with the bootstrap members in a way where we try to contact them periodically when they
  dropped out of our member list. Depending on a configuration, bootstrap members could be re-added regularly again
  under the assumption that the bootstrap members are always there. If bootstrap members are ephemeral, this should be
  disabled.
- Replace the roundtriptime.Tracker sort implementation with a quick select implementation for faster results.

### Nice to Have

- Should the FromBuffer functions return the remaining buffer to make it easier and less error-prone to work with?
- Make sure we provide enough context for all error returns.
- Check if we really need to use panic anywhere.
- The tests for messages are very repeating. Look into re-using code.
- The message setup in the message tests are quite repetitive. We should consolidate them.
- Introduce jitter into the scheduler to avoid spikes in network traffic.
- Regularly log the statistical information about ping chance and time until all know about gossip
- Do a TCP ping when the UDP ping times out for networks which do not correctly route UDP.
