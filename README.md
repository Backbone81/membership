# Membership

This Go library provides a high-performance, zero-allocation membership list for distributed processes. It is designed
for systems that need to know all other nodes in the cluster and need timely information about a node not being
available. The implementation is based on the
[SWIM: Scalable Weakly-consistent Infection-style Process Group Membership Protocol](https://doi.org/10.1109/DSN.2002.1028914)
with improvements from
[Lifeguard: Local Health Awareness for More Accurate Failure Detection](https://doi.org/10.48550/arXiv.1707.00788).

## What is a Distributed Membership List?

A distributed membership list provides a list of all members alive in the cluster. The list is maintained and updated
in a peer-to-peer process without a single point of failure. Changes to the membership list are propagated by gossip
mechanics in an infection style way.

## Use Cases

- **Distributed Databases**: Maintain cluster topology and route requests to live nodes.
- **Service Meshes**: Track healthy service instances for load balancing.
- **Cache Clusters**: Coordinate cache invalidation across nodes.
- **Consensus Systems**: Track participant availability before elections.
- **Monitoring Systems**: Detect and report node failures automatically.

## Why Use This Library?

- **No Single Point of Failure**: Pure peer-to-peer design.
- **Constant Network Load**: O(1) network traffic regardless of cluster size.
- **Adaptive Timing**: Dynamic timeout adjustment based on network RTT.
- **Encrypted by Default**: AES-256-GCM encryption on all network traffic.
- **Custom Metrics**: Integrate with your monitoring stack for operational insights.
- **Zero Allocations**: Engineered for minimal GC pressure and maximum throughput.

## Quick Start

Install:

```shell
go get github.com/backbone81/membership
```

A minimal example to run the membership list looks like this:

```go
package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/stdr"

	"github.com/backbone81/membership/pkg/membership"
)

func main() {
	if err := execute(); err != nil {
		log.Fatalln(err)
	}
}

func execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))

	bindAddress := membership.NewAddress(net.IPv4(127, 0, 0, 1), 3000)
	bootstrapMemberAddress := membership.NewAddress(net.IPv4(127, 0, 0, 1), 3001)
	membershipList, err := membership.NewList(
		membership.WithLogger(logger),
		membership.WithBootstrapMembers([]membership.Address{bootstrapMemberAddress}),
		membership.WithAdvertisedAddress(bindAddress),
		membership.WithBindAddress(bindAddress.String()),
		membership.WithEncryptionKey(membership.NewRandomKey()),
	)
	if err != nil {
		return err
	}

	if err := membershipList.Startup(); err != nil {
		return err
	}
	<-ctx.Done()
	if err := membershipList.Shutdown(); err != nil {
		return err
	}
	return nil
}
```

When the membership list is up and running, you can get notified with callbacks registered by
membership.WithMemberAddedCallback() and membership.WithMemberRemovedCallback() of members being added or removed, or
you can iterate over all members with list.ForEach().

Build your own application on top of that membership list then.

See the [examples](examples) folder for more examples.

## Configuration

All available [configuration options](pkg/membership/option.go) for the membership list can be seen in the source code.

## CLI

You can also use the CLI for running simulations or for generating random encryption keys. To install:

```shell
go install github.com/backbone81/membership/cmd/membership
```

Use the `--help` flag to get an overview of all options:

```
Command line tool supporting the membership Go library.

Usage:
  membership [command]

Available Commands:
  completion          Generate the autocompletion script for the specified shell
  failure-detection   How long a cluster needs to detect a failed member.
  failure-propagation How long a cluster needs to propagate a failed member.
  help                Help about any command
  infection           Simulates information dissemination through infection.
  join-propagation    How long a cluster needs to propagate a joined member.
  keygen              Creates a random encryption key.
  lossy-join          Joins a set of new members through a lossy network.
  statistics          Displays some analytical statistics about clusters.

Flags:
  -h, --help   help for membership

Use "membership [command] --help" for more information about a command.
```

## How It Works

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

## Joining a Cluster

For a member to join a cluster, it only needs to know one other member of the cluster. That member is then used as a
bootstrap member from which the full member list is retrieved.

The bootstrap members can also be used to automatically heal a network partition. If a network partition separates
parts of a cluster from each other, every part wil declare the other parts as faulty. Those faulty nodes cannot
refute their faulty state, because they cannot be reached. After the network partition is healed, the cluster parts
would stay separated, if no action was taken.

By default, bootstrap members are regularly re-added back to the membership list in an attempt to re-connect cluster
parts. It is therefore beneficial to use bootstrap members which span multiple availability zones or data centers.
Note, that this assumes that the bootstrap members are quite static. In situations where bootstrap members themselves
are ephemeral, this approach might lead to necessary re-joins followed by suspect and faulty declarations. The re-add
functionality can be disabled in such situations.

## Eventual Consistency

The membership list is provided with eventual consistency. As changes in membership are propagated by gossip through
the cluster, it might take 9 protocol periods for small clusters (16 members) to 30 protocol periods for big clusters
(16,000 members) for changes to propagate.

## Failure Detection and Propagation

Detecting a faulty member happens in a constant time which is around 1.6 protocol periods. This is independent of the
cluster size. As a member is first declared suspect and the suspicion timeout is dependent on the cluster size, the
actual declaration of a member as faulty increases slightly with cluster size.

As the propagation of gossip through the cluster happens in the logarithm of the cluster size, propagation is
dependent on the cluster size. As the suspect timeout should provide the suspect the chance to refute itself being
suspect, the suspicion timeout is dependent on the propagation time.

## Picking Members

When picking members for direct pings, we want to make sure that every member is picked as a target at some point in
time. Picking members completely at random could result in a situation where a specific member is by bad luck never
picked. To address this issue, direct ping targets are picked by having a random list of all members and that list is
worked through one by one. At the end of the list, the list is shuffled again, and we start from the beginning of that
list. This ensures that in a worst case scenario where a new member is added right before our current position and the
next shuffle places that member farthest away, we have 2n protocol periods as an upper bound.

When picking members to request indirect pings, we pick those members completely at random. We do not have the same
requirements for an upper bound here.

When picking a member to execute a full membership list sync, we pick those members completely at random as well.

## Anti Entropy

Because members are selected at random, there might be situations where because of bad luck, gossip does not reach every
single member. To heal those situations, a periodic full membership list sync is used to restore the membership list.

## Network Messages

The membership list communicates primarily with UDP messages. Care should be taken to choose the maximum message size
in a way which prevents splitting of those UDP messages. Otherwise, the reliability of the membership list deteriorates
quickly. The default value is a safe value for most networks, including the internet.

TCP connections are used for the full membership list sync, as this transfers more data, which needs to be exchanged
reliably.

## CPU and Network Load

The CPU and network load for each member is low and basically independent of the cluster size. Only the periodic full
membership list sync scales with the cluster size. But that full sync is not happening often and is still low overhead.

## Dynamic Direct Ping Timeout Adjustments

The membership list is keeping track of the round trip time for direct and indirect pings. By default, it keeps track of
the last 100 round trip times. The 99th percentile of those measured round trip times is then used to dynamically adjust
the timeout of direct pings. This allows to initiate an indirect ping earlier, when we know that the other member should
have responded already. In situations where the network is slow, because of high network load, that 99th percentile
causes the timeout to be extended up to a maximum which still fits into the protocol period.

Note, that the 99th percentile reacts very quickly when the network gets slower, but needs some time to catch up, when
the network is faster again. This is a desired property, as we want to adapt quickly to slowness, but be careful with
speeding up again.

## Gossip

With each network message, gossip is piggybacked. Gossip is chosen in a way where the gossip which was gossiped the
least amount of time is prioritized. This helps with distributing new information quicker in the cluster.

In addition, if the member has some gossip which is about the other member it is communicating with, that gossip is
always sent first to help that member with refuting suspect or faulty declarations as quickly as possible.

When gracefully shutting down, members will gossip their own failure immediately to some random members. This allows
the member in shutdown to be removed from the membership lists without waiting for pings to fail and suspects to
time out.

In situations where a lot of gossip needs to be propagated and that gossip is significantly more than what can be
piggybacked on a single network message, we dynamically increase the number of members which are targeted by a direct
ping in a single protocol period. This helps with distributing gossip quicker at the expense of increased CPU and
network load.

## Encryption And Key Rotation

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

## Logging

This library is using log levels to provide different details about its operation. The higher log levels always include
all logs of the lower log levels. Log level 0 is intended for production use while the other log levels are intended
for debugging purposes.

- Log level 0: General status information like members added and removed
- Log level 1: Network messages sent
- Log level 2: Network messages received
- Log level 3: Gossip messages received

## Metrics

Several metrics are provided to gain insights into the operation of the membership list. You can register those metrics
with your prometheus registerer with `membership.RegisterMetrics()`.

## Benchmarks

All parts of this library are covered with extensive benchmarks. See [docs](docs) for details.

## TODOs

- We might want to separate gossip count from suspect timeout
- Introduce jitter into the scheduler to avoid spikes in network traffic.
- Do a TCP ping when the UDP ping times out for networks which do not correctly route UDP.
- Replace the roundtriptime.Tracker sort implementation with a quick select implementation for faster results.
