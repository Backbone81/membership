# Design

This document describes some of the core design decisions for the library.

## Separation of Algorithm from Temporal Aspects and Scheduling

The main algorithm which is executing the SWIM protocol is designed in a way which does not require any knowledge about
timeouts or delays. This allows us to run tests and simulations in a synchronous lockstep fashion without having to
wait for actual timeouts to happen. Hundreds of protocol periods can be executed in under one second.

The only time relevant work the main algorithm is doing is measuring the round trip time for pings. This information
is then provided to the scheduler which is responsible for driving the main algorithm with all its timeouts, which
allows for a dynamic modification of timeouts by the scheduler according to the observed round trip time.

Having all timing related logic consolidated with the scheduler also allows for an easier time detecting a local
overload situation. When the scheduler expects for example 300ms if time between a direct ping and an indirect ping
but notices that in fact there are only 20ms left or that the indirect ping is happening way after the expected timeout,
it can directly infer the overload situation and immediately adjust the timing of the scheduler. This allows for a
significant simpler implementation compared to the local health mechanic described in the Lifeguard paper.

## Support Huge Clusters

The algorithms and data structures chosen for the implementation should support huge clusters without any issues. A
huge cluster can be seen as a cluster with 16,000 members. In such a situation our goal is that each high level
protocol step which is driven by the scheduler (executing direct pings, executing indirect pings and handling timeouts
at the end of the protocol period) must complete in under 1 millisecond, preferably in under 500 microseconds. This
ensures that one step of the algorithm is guaranteed to complete within a single timeslice allocated by the operating
scheduler and is less likely to be preempted.

While the described goal is achievable in a lot of situations, some functionality scales linear with the size of the
cluster. This is for example the case for a full membership list sync done with some other member or the listing of
all known members in the membership list. Those activities are then explicitly excluded from the time goal, as they
are inherently difficult or even impossible to optimize for those runtime restrictions.

## No Memory Allocations

Memory allocations and garbage collection slow down a process. By aiming for an implementation without any memory
allocations, the algorithm runs faster, does not put any pressure on the garbage collection and is more likely to
complete under high load scenarios.

While memory allocations cannot be avoided in all situations, we try to amortize memory allocations wherever they happen
by re-using the allocated memory over as many protocol periods as possible.

In practice, this design goal means that we do not start new go routines during runtime. Starting some go routines
during library startup is fine, but starting new go routines later would incur memory allocations. The same is with
channels. We do not allocate channels during runtime, as that would incur memory allocations. Make calls should only
happen once during initialization and always allocate a capacity which can hold enough data for several data points.
Append calls should always happen on slices which have a pre-allocated capacity. Maps should be pre-allocated for
a specific capacity and wiped by a clear call instead of re-allocating it. Timers should only be created during library
startup but not during normal runtime, as timers create go routines under the hood.

## Keeping Things Private

All the implementation is seen as internal code and therefore must live under the internal folder. This prevents
users from accidentally referencing internal implementation details.

The public interface is a wrapper around those internal implementation details and may export some of the internal data
types. This makes the interface to the library a deliberate decision and prevents conflicts of interest when testing
code which requires additional methods exposed, which should only be used during testing.

## Deviations from SWIM and Lifeguard

While the SWIM and Lifeguard papers describe the overall algorithm and some optimizations, there are some aspects
which we deliberately deviate from.

The SWIM paper describes the network messages ping, ack, ping-req. To allow for more clarity and also to measure
round trip times correctly for dynamic timeout adjustments, those network messages are replaced by direct ping, direct
ack, indirect ping, indirect ack.

The SWIM paper describes the gossip messages alive, suspect, confirm. To allow for more clarity, those gossip
messages are replaced by alive, suspect, faulty.

The Lifeguard paper describes the local health with bookkeeping about which and how many messages come in. To simplify
the local health detection, the scheduler is observing the time it is able to actually schedule steps of the algorithm
and derive overload information from that timing.

The SWIM and Lifeguard papers are talking about "membership groups". While this is the correct academic term, this
library is using the more technical term "clusters" to refer to the same concept.
