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
- Zero memory allocations during normal operation.
- The number of members targeted by a direct ping during each protocol period is dynamically adjusted according to the
  gossip messages waiting to be disseminated. With more messages in the queue a higher direct ping member count can
  help with disseminating those messages faster.
- Faulty members are dropped after they have been propagated often enough through the full memberlist sync.
- Network messages are always encrypted with AES-256 in GCM mode.

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

## Logging

This library is using log levels to provide different details about its operation. The higher log levels always include
all logs of the lower log levels. Log level 0 is intended for production use while the other log levels are intended
for debugging purposes.

- Log level 0: General status information like members added and removed
- Log level 1: Network messages sent
- Log level 2: Network messages received
- Log level 3: Gossip messages received

## Encryption

All network messages exchanged between members are encrypted with AES-256 with GCM. This allows members to operate
over untrusted networks without leaking confidential data. All members need to have the same encryption key configured
to be able to communicate with each other.

To allow encryption key rotation without shutting down all members in the cluster, multiple encryption keys can be
specified. The first encryption key is always used for encrypting sent messages, while all encryption keys are tried
in order to decrypt received messages. If a new encryption key is first added as the last key to all members, it
allows all members to decrypt messages with that key. When all members know about the new encryption key, the new key
can be moved from the last position to the first one. This makes the key active for encryption. When all members are
using the new encryption key, the old key can be removed.

To make sure that the encryption cannot be broken, you need to rotate the encryption key after some number of encryption
operations. Recommendations range from
2^24.5 = 23,726,566 (https://www.rfc-editor.org/rfc/rfc8446.html#section-5.5) to
2^32 = 4,294,967,296 (https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf) encryption
operations to stay safe. This is a property of the 96 bits of nonce used. Use the metric
`membership_list_transport_encryptions_total` to keep track of encryption operation count.

A cluster of 256 members with a protocol period of 1 second will have at least 256 (members) * 2 (ping + ack) *
60 (seconds) * 60 (minutes) * 24 (hours) = 44,236,800 encryption operations per day. Make sure to rotate the key
accordingly, if your cluster spans untrusted networks.

To create a new random encryption key, you can use the `keygen` subcommand of the `membership` cli.

```shell
go run ./cmd/membership keygen
```

## TODOs

### Important Topics

- Investigate how we can increase the suspicion timeout when we are under high CPU load. High CPU load can be detected
  by the scheduler as the times between direct pings, indirect pings and end of protocol are either significant shorter
  than expected or even overshot immediately.
- How can a member re-join when it was disconnected through a network partition from everybody else for a long time?
  We probably need to deal with the bootstrap members in a way where we try to contact them periodically when they
  dropped out of our member list. Depending on a configuration, bootstrap members could be re-added regularly again
  under the assumption that the bootstrap members are always there. If bootstrap members are ephemeral, this should be
  disabled.
- Use timeouts for the tcp transports.
- Align code with linter
- There is a bug, which causes every list request to trigger a refute about being alive with an increase in incarnation
  number.

### Nice to Have

- We might want to separate gossip count from suspect timeout
- Introduce jitter into the scheduler to avoid spikes in network traffic.
- Do a TCP ping when the UDP ping times out for networks which do not correctly route UDP.
- Replace the roundtriptime.Tracker sort implementation with a quick select implementation for faster results.
