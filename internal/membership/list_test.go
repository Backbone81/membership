package membership_test

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/backbone81/membership/internal/utility"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/transport"
)

var _ = Describe("List", func() {
	Context("NewList", func() {
		It("should create a list with default configuration", func() {
			list := newTestList()
			debugList := membership.DebugList(list)

			Expect(list).NotTo(BeNil())
			Expect(list.Len()).To(Equal(0))
			Expect(Collect(list)).To(BeEmpty())
			Expect(debugList.GetMembers()).To(BeEmpty())
			Expect(debugList.GetFaultyMembers()).To(BeEmpty())
		})

		It("should apply WithAdvertisedAddress option", func() {
			address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			list := newTestList(
				membership.WithAdvertisedAddress(address),
			)

			config := list.Config()
			Expect(config.AdvertisedAddress).To(Equal(address))
		})

		It("should apply WithSafetyFactor option", func() {
			list := newTestList(
				membership.WithSafetyFactor(5.0),
			)

			config := list.Config()
			Expect(config.SafetyFactor).To(Equal(5.0))
		})

		It("should set negative safety factor to 0", func() {
			list := newTestList(
				membership.WithSafetyFactor(-5.0),
			)

			config := list.Config()
			Expect(config.SafetyFactor).To(Equal(0.0))
		})

		It("should apply WithDirectPingMemberCount option", func() {
			list := newTestList(
				membership.WithDirectPingMemberCount(2),
			)

			config := list.Config()
			Expect(config.DirectPingMemberCount).To(Equal(2))
		})

		It("should set negative direct ping member count to 1", func() {
			list := newTestList(
				membership.WithDirectPingMemberCount(-5),
			)

			config := list.Config()
			Expect(config.DirectPingMemberCount).To(Equal(1))
		})

		It("should apply WithIndirectPingMemberCount option", func() {
			list := newTestList(
				membership.WithIndirectPingMemberCount(5),
			)

			config := list.Config()
			Expect(config.IndirectPingMemberCount).To(Equal(5))
		})

		It("should set negative indirect ping member count to 1", func() {
			list := newTestList(
				membership.WithIndirectPingMemberCount(-5),
			)

			config := list.Config()
			Expect(config.IndirectPingMemberCount).To(Equal(1))
		})

		It("should apply WithShutdownMemberCount option", func() {
			list := newTestList(
				membership.WithShutdownMemberCount(10),
			)

			config := list.Config()
			Expect(config.ShutdownMemberCount).To(Equal(10))
		})

		It("should set negative shutdown member count to 1", func() {
			list := newTestList(
				membership.WithShutdownMemberCount(-10),
			)

			config := list.Config()
			Expect(config.ShutdownMemberCount).To(Equal(1))
		})

		It("should apply WithMaxDatagramLengthSend option", func() {
			list := newTestList(
				membership.WithMaxDatagramLengthSend(2000),
			)

			config := list.Config()
			Expect(config.MaxDatagramLengthSend).To(Equal(2000))
		})

		It("should set negative max datagram length send to 1", func() {
			list := newTestList(
				membership.WithMaxDatagramLengthSend(-2000),
			)

			config := list.Config()
			Expect(config.MaxDatagramLengthSend).To(Equal(1))
		})

		It("should process bootstrap members as alive", func() {
			bootstrap := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
			}

			list := newTestList(
				membership.WithBootstrapMembers(bootstrap),
			)
			debugList := membership.DebugList(list)

			Expect(list.Len()).To(Equal(3))
			members := debugList.GetMembers()
			Expect(members).To(HaveLen(3))

			for i, member := range members {
				Expect(member.Address).To(Equal(bootstrap[i]))
				Expect(member.State).To(Equal(encoding.MemberStateAlive))
				Expect(member.IncarnationNumber).To(Equal(uint16(0)))
			}
		})

		It("should exclude self from bootstrap members", func() {
			self := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			bootstrap := []encoding.Address{
				self, // Should be excluded
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
			}

			list := newTestList(
				membership.WithAdvertisedAddress(self),
				membership.WithBootstrapMembers(bootstrap),
			)
			debugList := membership.DebugList(list)

			Expect(list.Len()).To(Equal(2), "Self should be excluded from member list")
			members := debugList.GetMembers()

			for _, member := range members {
				Expect(member.Address).NotTo(Equal(self))
			}
		})

		It("should deduplicate bootstrap members", func() {
			duplicate := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			bootstrap := []encoding.Address{
				duplicate,
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				duplicate,
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
			}

			list := newTestList(
				membership.WithBootstrapMembers(bootstrap),
			)

			Expect(list.Len()).To(Equal(3), "Duplicates should be removed")
		})

		It("should register member added callback", func() {
			var addedCount int
			var addedAddresses []encoding.Address

			_ = newTestList(
				membership.WithMemberAddedCallback(func(address encoding.Address) {
					addedCount++
					addedAddresses = append(addedAddresses, address)
				}),
				membership.WithBootstrapMembers([]encoding.Address{
					encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
					encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				}),
			)

			// Callbacks should have been invoked for bootstrap members
			Expect(addedCount).To(Equal(2))
			Expect(addedAddresses).To(HaveLen(2))
		})

		It("should allow options to override each other", func() {
			list := newTestList(
				membership.WithSafetyFactor(3.0),
				membership.WithSafetyFactor(5.0), // Should override previous
			)

			config := list.Config()
			Expect(config.SafetyFactor).To(Equal(5.0), "Later option should override")
		})

		It("should maintain sorted member list from bootstrap", func() {
			bootstrap := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}

			list := newTestList(
				membership.WithBootstrapMembers(bootstrap),
			)

			addresses := Collect(list)
			Expect(addresses).To(HaveLen(3))

			// Verify sorted ascending
			for i := 0; i < len(addresses)-1; i++ {
				Expect(encoding.CompareAddress(addresses[i], addresses[i+1])).To(Equal(-1),
					"Addresses should be sorted ascending")
			}
		})

		It("should initialize with gossip about self", func() {
			list := newTestList()
			debugList := membership.DebugList(list)

			gossipQueue := debugList.GetGossip()
			Expect(gossipQueue.Len()).To(Equal(1))

			msg := debugList.GetGossip().Get(0)
			Expect(msg.Type).To(Equal(encoding.MessageTypeAlive))
			Expect(msg.Destination).To(Equal(TestAddress))
			Expect(msg.IncarnationNumber).To(Equal(uint16(0)))
		})
	})

	Context("All", func() {
		It("should return empty iterator for empty list", func() {
			list := newTestList()

			var count int
			list.ForEach(func(address encoding.Address) bool {
				count++
				return true
			})

			Expect(count).To(Equal(0))
		})

		It("should iterate over multiple members", func() {
			bootstrap := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrap),
			)

			var addresses []encoding.Address
			list.ForEach(func(address encoding.Address) bool {
				addresses = append(addresses, address)
				return true
			})

			Expect(addresses).To(HaveLen(3))
			Expect(addresses).To(ConsistOf(bootstrap))
		})

		It("should yield members in sorted ascending order", func() {
			bootstrap := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrap),
			)

			var addresses []encoding.Address
			list.ForEach(func(address encoding.Address) bool {
				addresses = append(addresses, address)
				return true
			})

			for i := 0; i < len(addresses)-1; i++ {
				Expect(encoding.CompareAddress(addresses[i], addresses[i+1])).To(Equal(-1),
					"Address at index %d (%v) should be less than address at index %d (%v)",
					i, addresses[i], i+1, addresses[i+1])
			}
		})

		It("should support early exit from iteration", func() {
			bootstrap := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrap),
			)

			var count int
			target := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3)
			var found encoding.Address

			list.ForEach(func(address encoding.Address) bool {
				count++
				if address == target {
					found = address
					return false
				}
				return true
			})

			Expect(found).To(Equal(target))
			Expect(count).To(Equal(3))
		})

		It("should not include faulty members", func() {
			aliveAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			faultyAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)

			list := newTestList(
				membership.WithBootstrapMembers([]encoding.Address{aliveAddr, faultyAddr}),
			)
			Expect(list.Len()).To(Equal(2))

			// Mark as faulty
			Expect(DispatchDatagram(list, encoding.MessageFaulty{
				Source:            TestAddress,
				Destination:       faultyAddr,
				IncarnationNumber: 0,
			}.ToMessage())).To(Succeed())
			Expect(list.Len()).To(Equal(1))

			// Should only iterate over alive member
			addresses := Collect(list)
			Expect(addresses).To(HaveLen(1))
			Expect(addresses[0]).To(Equal(aliveAddr))
			Expect(addresses).NotTo(ContainElement(faultyAddr))
		})

		It("should include suspect members", func() {
			aliveAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			suspectAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)

			list := newTestList(
				membership.WithBootstrapMembers([]encoding.Address{aliveAddr, suspectAddr}),
			)

			// Mark one as suspect
			Expect(DispatchDatagram(list, encoding.MessageSuspect{
				Source:            TestAddress,
				Destination:       suspectAddr,
				IncarnationNumber: 0,
			}.ToMessage())).To(Succeed())
			Expect(list.Len()).To(Equal(2))

			// Should iterate over both alive and suspect
			addresses := Collect(list)
			Expect(addresses).To(HaveLen(2))
			Expect(addresses).To(ContainElement(aliveAddr))
			Expect(addresses).To(ContainElement(suspectAddr))
		})
	})

	Context("DirectPing", func() {
		It("should not send ping when member list is empty", func() {
			var store transport.Store
			list := newTestList(
				membership.WithUDPClient(&store),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			Expect(store.Addresses).To(BeEmpty())
			pendingPings := debugList.GetPendingDirectPings()
			Expect(pendingPings).To(HaveLen(0))
		})

		It("should send ping to single member", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).To(Equal(bootstrapMembers[0]))
			Expect(store.Buffers).To(HaveLen(1))
			pendingPings := debugList.GetPendingDirectPings()
			Expect(pendingPings).To(HaveLen(1))
		})

		It("should send pings to multiple members based on DirectPingMemberCount", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithDirectPingMemberCount(3),
			)
			debugList := membership.DebugList(list)

			By("Executing 3 direct pings")
			Expect(list.DirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(3))
			Expect(store.Buffers).To(HaveLen(3))
			pendingPings := debugList.GetPendingDirectPings()
			Expect(pendingPings).To(HaveLen(3))
		})

		It("should use round-robin selection across multiple calls", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing 10 direct ping (2 full rounds through 5 members)")
			pingCounts := make(map[encoding.Address]int)
			for range 10 {
				store.Clear()
				Expect(list.DirectPing()).To(Succeed())
				Expect(store.Addresses).To(HaveLen(1))
				Expect(store.Buffers).To(HaveLen(1))
				pingCounts[store.Addresses[0]]++
			}

			Expect(pingCounts).To(HaveLen(5))
			for addr, count := range pingCounts {
				Expect(count).To(Equal(2), "Address %v should be pinged twice", addr)
			}
		})

		It("should distribute pings evenly with larger DirectPingMemberCount", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 6),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 7),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 8),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 9),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 10),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithDirectPingMemberCount(2),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing 20 direct pings (10 members, 2 pings each)")
			pingCounts := make(map[encoding.Address]int)
			for range 10 {
				store.Clear()
				Expect(list.DirectPing()).To(Succeed())
				Expect(store.Addresses).To(HaveLen(2))
				Expect(store.Buffers).To(HaveLen(2))
				for _, addr := range store.Addresses {
					pingCounts[addr]++
				}
			}

			Expect(pingCounts).To(HaveLen(10))
			for addr, count := range pingCounts {
				Expect(count).To(Equal(2), "Address %v should be pinged twice", addr)
			}
		})

		It("should handle DirectPingMemberCount larger than member list", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithDirectPingMemberCount(5), // More than 2 members
			)
			debugList := membership.DebugList(list)

			By("Executing 2 direct pings")
			Expect(list.DirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(2))
			Expect(store.Buffers).To(HaveLen(2))
			pendingPings := debugList.GetPendingDirectPings()
			Expect(pendingPings).To(HaveLen(2))
		})

		It("should track pending pings for timeout detection", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingPings := debugList.GetPendingDirectPings()
			Expect(pendingPings).To(HaveLen(1))
			Expect(pendingPings[0].Destination).To(Equal(bootstrapMembers[0]))
		})

		It("should increment sequence numbers for each ping", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			var msg1 encoding.MessageDirectPing
			Expect(msg1.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			seq1 := msg1.SequenceNumber

			By("Executing 1 direct ping")
			store.Clear()
			Expect(list.DirectPing()).To(Succeed())
			var msg2 encoding.MessageDirectPing
			Expect(msg2.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			seq2 := msg2.SequenceNumber

			// Sequence numbers should increment
			Expect(seq2).To(BeNumerically(">", seq1))
		})

		It("should ping suspect members", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithDirectPingMemberCount(2),
			)

			By("Marking one member as suspect")
			Expect(DispatchDatagram(list, encoding.MessageSuspect{
				Source:            TestAddress,
				Destination:       bootstrapMembers[1],
				IncarnationNumber: 0,
			}.ToMessage())).To(Succeed())

			By("Executing 1 direct ping")
			store.Clear()
			Expect(list.DirectPing()).To(Succeed())

			// Both members should be pinged (suspect members are still active)
			Expect(store.Addresses).To(HaveLen(2))
			Expect(store.Addresses).To(ContainElement(bootstrapMembers[0]))
			Expect(store.Addresses).To(ContainElement(bootstrapMembers[1]))
		})

		It("should handle UDP send errors gracefully", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&transport.Error{}),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).ToNot(Succeed())
		})
	})

	Context("IndirectPing", func() {
		It("should not send indirect ping when no pending direct pings", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing 1 indirect ping")
			Expect(list.IndirectPing()).To(Succeed())
			Expect(store.Addresses).To(BeEmpty())
		})

		It("should send indirect ping requests for pending direct pings", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingDirectPing := debugList.GetPendingDirectPings()[0]

			By("Executing 1 indirect ping")
			store.Clear()
			Expect(list.IndirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).ToNot(Equal(pendingDirectPing.Destination))

			By("Verifying the network message")
			var msg encoding.MessageIndirectPing
			Expect(msg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(msg.Source).To(Equal(TestAddress))
			Expect(msg.Destination).To(Equal(pendingDirectPing.Destination))
		})

		It("should use IndirectPingMemberCount for number of proxies", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithIndirectPingMemberCount(3),
			)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())

			By("Executing 3 indirect pings")
			store.Clear()
			Expect(list.IndirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(3))
			Expect(store.Buffers).To(HaveLen(3))
		})

		It("should not use target member as proxy", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithIndirectPingMemberCount(2),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingDirectPings := debugList.GetPendingDirectPings()
			Expect(pendingDirectPings).To(HaveLen(1))

			By("Executing 1 indirect ping when 2 were requested")
			store.Clear()
			Expect(list.IndirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).NotTo(Equal(pendingDirectPings[0].Destination))
			pendingIndirectPings := debugList.GetPendingIndirectPings()
			Expect(pendingIndirectPings).To(HaveLen(1))
		})

		It("should handle multiple pending direct pings", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithDirectPingMemberCount(2),
			)
			debugList := membership.DebugList(list)

			By("Executing 2 direct pings")
			Expect(list.DirectPing()).To(Succeed())
			pendingDirectPings := debugList.GetPendingDirectPings()
			Expect(pendingDirectPings).To(HaveLen(2))

			By("Executing 6 indirect pings")
			store.Clear()
			Expect(list.IndirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(6))
			pendingIndirectPings := debugList.GetPendingIndirectPings()
			Expect(pendingIndirectPings).To(HaveLen(2))
		})

		It("should handle insufficient members for proxies", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithIndirectPingMemberCount(5),
			)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())

			By("Executing 0 indirect ping")
			store.Clear()
			Expect(list.IndirectPing()).To(Succeed())
			Expect(store.Addresses).To(BeEmpty())
		})

		It("should track pending indirect pings", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingDirectPing := debugList.GetPendingDirectPings()
			Expect(pendingDirectPing).To(HaveLen(1))

			By("Executing 1 indirect ping")
			store.Clear()
			Expect(list.IndirectPing()).To(Succeed())
			Expect(store.Addresses).To(HaveLen(1))
			pendingIndirectPings := debugList.GetPendingIndirectPings()
			Expect(pendingIndirectPings).To(HaveLen(1))
			Expect(pendingIndirectPings[0].MessageIndirectPing.Destination).To(Equal(pendingDirectPing[0].Destination))
		})

		It("should use same sequence number as direct ping", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			var directMsg encoding.MessageDirectPing
			Expect(directMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			directSeq := directMsg.SequenceNumber

			By("Executing 1 indirect ping")
			store.Clear()
			Expect(list.IndirectPing()).To(Succeed())
			var indirectMsg encoding.MessageIndirectPing
			Expect(indirectMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(indirectMsg.SequenceNumber).To(Equal(directSeq))
		})

		It("should handle UDP send errors gracefully", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&transport.Discard{}),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())

			By("Executing 1 indirect ping with broken network")
			config := list.Config()
			config.UDPClient = &transport.Error{}
			debugList.SetConfig(config)
			err := list.IndirectPing()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("EndOfProtocolPeriod", func() {
		It("should do nothing when no pending pings", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing end of protocol period")
			Expect(list.EndOfProtocolPeriod()).To(Succeed())

			By("Verifying no state changes")
			Expect(list.Len()).To(Equal(1))
			members := debugList.GetMembers()
			Expect(members[0].State).To(Equal(encoding.MemberStateAlive))
		})

		It("should mark member as suspect after failed direct ping", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing 1 direct ping")
			Expect(list.DirectPing()).To(Succeed())
			Expect(debugList.GetPendingDirectPings()).To(HaveLen(1))

			By("Executing end of protocol period")
			debugList.ClearGossip()
			Expect(list.EndOfProtocolPeriod()).To(Succeed())

			By("Verifying member is now suspect")
			members := debugList.GetMembers()
			Expect(members).To(HaveLen(1))
			Expect(members[0].State).To(Equal(encoding.MemberStateSuspect))
			Expect(debugList.GetPendingDirectPings()).To(BeEmpty())
			gossipQueue := debugList.GetGossip()
			Expect(gossipQueue.Len()).To(Equal(1))

			msg := gossipQueue.Get(0)
			Expect(msg.Type).To(Equal(encoding.MessageTypeSuspect))
			Expect(msg.Destination).To(Equal(bootstrapMembers[0]))
		})

		It("should increment suspect counter for existing suspects", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithSafetyFactor(1000),
			)
			debugList := membership.DebugList(list)

			By("Marking member as suspect")
			Expect(list.DirectPing()).To(Succeed())
			Expect(list.EndOfProtocolPeriod()).To(Succeed())
			members := debugList.GetMembers()
			Expect(members[0].State).To(Equal(encoding.MemberStateSuspect))
			Expect(members[0].SuspicionPeriodCounter).To(Equal(1))

			By("Executing end of protocol period")
			Expect(list.EndOfProtocolPeriod()).To(Succeed())
			members = debugList.GetMembers()
			Expect(members[0].SuspicionPeriodCounter).To(Equal(2))

			By("Executing another end of protocol period")
			Expect(list.EndOfProtocolPeriod()).To(Succeed())
			members = debugList.GetMembers()
			Expect(members[0].SuspicionPeriodCounter).To(Equal(3))
		})

		It("should mark suspect as faulty after exceeding threshold", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithSafetyFactor(0), // Threshold = 0, immediate faulty
			)
			debugList := membership.DebugList(list)

			By("Marking member as suspect")
			Expect(list.DirectPing()).To(Succeed())
			debugList.ClearGossip()
			Expect(list.EndOfProtocolPeriod()).To(Succeed())
			Expect(debugList.GetMembers()).To(HaveLen(0))
			members := debugList.GetFaultyMembers()
			Expect(members).To(HaveLen(1))
			Expect(members[0].State).To(Equal(encoding.MemberStateFaulty))

			By("Verifying faulty gossip message added")
			gossipQueue := debugList.GetGossip()
			msg := gossipQueue.Get(0)
			Expect(msg.Type).To(Equal(encoding.MessageTypeFaulty))
			Expect(msg.Destination).To(Equal(bootstrapMembers[0]))
		})

		It("should calculate correct suspicion threshold based on member count and safety factor", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithDirectPingMemberCount(5), // ping all members
				membership.WithSafetyFactor(3.0),
			)

			expectedThreshold := int(math.Ceil(utility.DisseminationPeriods(3, len(bootstrapMembers))))

			By("Marking members as suspect")
			for range expectedThreshold {
				Expect(list.DirectPing()).To(Succeed())
				Expect(list.EndOfProtocolPeriod()).To(Succeed())
			}
			Expect(list.Len()).To(Equal(5))

			By("Removing failed members")
			Expect(list.DirectPing()).To(Succeed())
			Expect(list.EndOfProtocolPeriod()).To(Succeed())
			Expect(list.Len()).To(Equal(0))
		})

		It("should invoke member removed callback when marking as faulty", func() {
			var removedCounter int
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithSafetyFactor(0),
				membership.WithMemberRemovedCallback(func(address encoding.Address) {
					removedCounter++
				}),
			)

			By("Marking member as faulty")
			Expect(list.DirectPing()).To(Succeed())
			Expect(list.EndOfProtocolPeriod()).To(Succeed())

			By("Verifying callback was invoked")
			Expect(removedCounter).To(Equal(1))
			Expect(list.Len()).To(Equal(0))
		})

		It("should clear pending indirect pings", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing direct and indirect pings")
			Expect(list.DirectPing()).To(Succeed())
			Expect(list.IndirectPing()).To(Succeed())
			Expect(debugList.GetPendingIndirectPings()).To(HaveLen(1))

			By("Executing end of protocol period")
			Expect(list.EndOfProtocolPeriod()).To(Succeed())
			Expect(debugList.GetPendingIndirectPings()).To(BeEmpty())
		})
	})

	Context("RequestList", func() {
		It("should not send request when member list is empty", func() {
			var store transport.Store
			list := newTestList(
				membership.WithUDPClient(&store),
			)

			By("Executing list request")
			Expect(list.RequestList()).To(Succeed())

			By("Verifying no network messages sent")
			Expect(store.Addresses).To(BeEmpty())
		})

		It("should send request to one of multiple members", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing list request")
			Expect(list.RequestList()).To(Succeed())

			By("Verifying network message sent")
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Buffers).To(HaveLen(1))

			By("Verifying message is ListRequest")
			var msg encoding.MessageListRequest
			Expect(msg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
		})
	})

	Context("BroadcastShutdown", func() {
		It("should not send shutdown when member list is empty", func() {
			var store transport.Store
			list := newTestList(
				membership.WithUDPClient(&store),
			)

			By("Executing broadcast shutdown")
			Expect(list.BroadcastShutdown()).To(Succeed())

			By("Verifying no network messages sent")
			Expect(store.Addresses).To(BeEmpty())
		})

		It("should send shutdown to single member", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Executing broadcast shutdown")
			Expect(list.BroadcastShutdown()).To(Succeed())

			By("Verifying network message sent")
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).To(Equal(bootstrapMembers[0]))
			Expect(store.Buffers).To(HaveLen(1))
		})

		It("should send shutdown to multiple members based on ShutdownMemberCount", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 5),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithShutdownMemberCount(3),
			)

			By("Executing broadcast shutdown")
			Expect(list.BroadcastShutdown()).To(Succeed())

			By("Verifying 3 shutdown messages sent")
			Expect(store.Addresses).To(HaveLen(3))
			Expect(store.Buffers).To(HaveLen(3))

			By("Verifying all targets are different")
			uniqueAddresses := make(map[encoding.Address]bool)
			for _, addr := range store.Addresses {
				uniqueAddresses[addr] = true
			}
			Expect(uniqueAddresses).To(HaveLen(3))
		})
	})

	Context("handleDirectPing", func() {
		It("should send direct ack when receiving direct ping", func() {
			var store transport.Store
			list := newTestList(
				membership.WithUDPClient(&store),
			)

			By("Sending direct ping")
			sourceAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			Expect(DispatchDatagram(list, encoding.MessageDirectPing{
				Source:         sourceAddr,
				SequenceNumber: 42,
			}.ToMessage())).To(Succeed())

			By("Verifying direct ack sent back to source")
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).To(Equal(sourceAddr))
			Expect(store.Buffers).To(HaveLen(1))

			By("Verifying ack message format")
			var ackMsg encoding.MessageDirectAck
			Expect(ackMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(ackMsg.SequenceNumber).To(Equal(uint16(42)))
		})

		It("should handle UDP send errors gracefully", func() {
			list := newTestList(
				membership.WithUDPClient(&transport.Error{}),
			)

			By("Sending direct ping")
			sourceAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			Expect(DispatchDatagram(list, encoding.MessageDirectPing{
				Source:         sourceAddr,
				SequenceNumber: 42,
			}.ToMessage())).ToNot(Succeed())
		})
	})

	Context("handleDirectAck", func() {
		It("should remove pending direct ping when receiving matching ack", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingPings := debugList.GetPendingDirectPings()
			Expect(pendingPings).To(HaveLen(1))

			targetAddr := pendingPings[0].Destination
			seqNum := pendingPings[0].MessageDirectPing.SequenceNumber

			By("Sending direct ack")
			Expect(DispatchDatagram(list, encoding.MessageDirectAck{
				Source:         targetAddr,
				SequenceNumber: seqNum,
			}.ToMessage())).To(Succeed())

			By("Verifying pending ping removed")
			Expect(debugList.GetPendingDirectPings()).To(BeEmpty())
		})

		It("should ignore ack with non-matching sequence number", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingPings := debugList.GetPendingDirectPings()
			Expect(pendingPings).To(HaveLen(1))

			By("Sending ack with wrong sequence number")
			Expect(DispatchDatagram(list, encoding.MessageDirectAck{
				Source:         pendingPings[0].Destination,
				SequenceNumber: 9999,
			}.ToMessage())).To(Succeed())

			By("Verifying pending ping still present")
			Expect(debugList.GetPendingDirectPings()).To(HaveLen(1))
		})

		It("should handle ack when no pending pings exist", func() {
			list := newTestList()

			By("Sending ack without any pending pings")
			Expect(DispatchDatagram(list, encoding.MessageDirectAck{
				Source:         encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				SequenceNumber: 42,
			}.ToMessage())).To(Succeed())
		})
	})

	Context("handleIndirectPing", func() {
		It("should send direct ping to target when receiving indirect ping request", func() {
			var store transport.Store
			list := newTestList(
				membership.WithUDPClient(&store),
			)

			By("Sending indirect ping request")
			sourceAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			targetAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)
			Expect(DispatchDatagram(list, encoding.MessageIndirectPing{
				Source:         sourceAddr,
				Destination:    targetAddr,
				SequenceNumber: 42,
			}.ToMessage())).To(Succeed())

			By("Verifying direct ping sent to target")
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).To(Equal(targetAddr))
			Expect(store.Buffers).To(HaveLen(1))

			By("Verifying direct ping message format")
			var directPingMsg encoding.MessageDirectPing
			Expect(directPingMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
		})

		It("should handle UDP send errors gracefully", func() {
			list := newTestList(
				membership.WithUDPClient(&transport.Error{}),
			)

			By("Sending indirect ping request")
			sourceAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			targetAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)
			Expect(DispatchDatagram(list, encoding.MessageIndirectPing{
				Source:         sourceAddr,
				Destination:    targetAddr,
				SequenceNumber: 42,
			}.ToMessage())).ToNot(Succeed())
		})
	})

	Context("handleIndirectAck", func() {
		It("should remove pending direct ping and indirect ping when receiving matching ack", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingDirectPings := debugList.GetPendingDirectPings()
			Expect(pendingDirectPings).To(HaveLen(1))

			By("Executing indirect ping")
			Expect(list.IndirectPing()).To(Succeed())
			pendingIndirectPings := debugList.GetPendingIndirectPings()
			Expect(pendingIndirectPings).To(HaveLen(1))

			targetAddr := pendingDirectPings[0].Destination
			seqNum := pendingDirectPings[0].MessageDirectPing.SequenceNumber

			By("Sending indirect ack")
			Expect(DispatchDatagram(list, encoding.MessageIndirectAck{
				Source:         targetAddr,
				SequenceNumber: seqNum,
			}.ToMessage())).To(Succeed())

			By("Verifying both pending pings removed")
			Expect(debugList.GetPendingDirectPings()).To(BeEmpty())
			Expect(debugList.GetPendingIndirectPings()).To(BeEmpty())
		})

		It("should ignore ack with non-matching sequence number", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithUDPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			debugList := membership.DebugList(list)

			By("Executing direct ping")
			Expect(list.DirectPing()).To(Succeed())
			pendingDirectPings := debugList.GetPendingDirectPings()
			Expect(pendingDirectPings).To(HaveLen(1))

			By("Executing indirect ping")
			Expect(list.IndirectPing()).To(Succeed())
			pendingIndirectPings := debugList.GetPendingIndirectPings()
			Expect(pendingIndirectPings).To(HaveLen(1))

			By("Sending ack with wrong sequence number")
			Expect(DispatchDatagram(list, encoding.MessageIndirectAck{
				Source:         pendingDirectPings[0].Destination,
				SequenceNumber: 9999,
			}.ToMessage())).To(Succeed())

			By("Verifying pending pings still present")
			Expect(debugList.GetPendingDirectPings()).To(HaveLen(1))
			Expect(debugList.GetPendingIndirectPings()).To(HaveLen(1))
		})

		It("should handle ack when no pending pings exist", func() {
			list := newTestList()

			By("Sending ack without any pending pings")
			Expect(DispatchDatagram(list, encoding.MessageIndirectAck{
				Source:         encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				SequenceNumber: 42,
			}.ToMessage())).To(Succeed())
		})
	})

	Context("handleAlive", func() {
		It("should refute alive about self", func() {
			list := newTestList()
			debugList := membership.DebugList(list)
			debugList.GetGossip().Clear()

			Expect(DispatchDatagram(list, encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 55,
			}.ToMessage())).To(Succeed())

			Expect(debugList.GetMembers()).To(BeEmpty())
			Expect(debugList.GetFaultyMembers()).To(BeEmpty())
			Expect(debugList.GetGossip().Len()).To(Equal(1))
			Expect(debugList.GetGossip().Get(0)).To(Equal(encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 56,
			}.ToMessage()))
		})

		It("should invoke member added callback when adding new member", func() {
			var addedCount int
			list := newTestList(
				membership.WithMemberAddedCallback(func(address encoding.Address) {
					addedCount++
				}),
			)

			By("Sending alive message for new member")
			Expect(DispatchDatagram(list, encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 5,
			}.ToMessage())).To(Succeed())

			By("Verifying callback invoked")
			Expect(addedCount).To(Equal(1))
		})

		DescribeTable("Gossip should update the memberlist correctly",
			func(beforeMembers []encoding.Member, beforeFaultyMembers []encoding.Member, message encoding.Message, afterMembers []encoding.Member, afterFaultyMembers []encoding.Member) {
				list := membership.NewList(
					membership.WithUDPClient(&transport.Discard{}),
					membership.WithTCPClient(&transport.Discard{}),
					membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
				)
				debugList := membership.DebugList(list)
				debugList.GetGossip().Clear()
				debugList.SetMembers(beforeMembers)
				debugList.SetFaultyMembers(beforeFaultyMembers)

				Expect(DispatchDatagram(list, message)).To(Succeed())

				if afterMembers == nil {
					Expect(debugList.GetMembers()).To(BeEmpty())
				} else {
					Expect(debugList.GetMembers()).To(Equal(afterMembers))
				}
				if afterFaultyMembers == nil {
					Expect(debugList.GetFaultyMembers()).To(BeEmpty())
				} else {
					Expect(debugList.GetFaultyMembers()).To(Equal(afterFaultyMembers))
				}
			},
			Entry("Alive should add a member",
				nil,
				nil,
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Alive with lower incarnation number should NOT overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Alive with same incarnation number should NOT overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Alive with bigger incarnation number should overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 3,
					},
				},
				nil,
			),

			Entry("Alive with lower incarnation number should NOT overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Alive with same incarnation number should NOT overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Alive with bigger incarnation number should overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 3,
					},
				},
				nil,
			),

			Entry("Alive with lower incarnation number should NOT overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Alive with same incarnation number should NOT overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Alive with bigger incarnation number should overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 3,
					},
				},
				nil,
			),
		)
	})

	Context("handleSuspect", func() {
		It("should refute suspect about self", func() {
			list := newTestList()
			debugList := membership.DebugList(list)
			debugList.GetGossip().Clear()

			By("Receiving a suspect message")
			Expect(DispatchDatagram(list, encoding.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 55,
			}.ToMessage())).To(Succeed())

			Expect(debugList.GetMembers()).To(BeEmpty())
			Expect(debugList.GetFaultyMembers()).To(BeEmpty())
			Expect(debugList.GetGossip().Len()).To(Equal(1))
			Expect(debugList.GetGossip().Get(0)).To(Equal(encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 56,
			}.ToMessage()))
		})

		It("should invoke member added callback when suspect", func() {
			var addedCount int
			list := newTestList(
				membership.WithMemberAddedCallback(func(address encoding.Address) {
					addedCount++
				}),
			)

			By("Sending suspect message")
			Expect(DispatchDatagram(list, encoding.MessageSuspect{
				Destination:       TestAddress2,
				IncarnationNumber: 5,
			}.ToMessage())).To(Succeed())

			By("Verifying callback invoked")
			Expect(addedCount).To(Equal(1))
		})

		DescribeTable("Gossip should update the memberlist correctly",
			func(beforeMembers []encoding.Member, beforeFaultyMembers []encoding.Member, message encoding.Message, afterMembers []encoding.Member, afterFaultyMembers []encoding.Member) {
				list := membership.NewList(
					membership.WithUDPClient(&transport.Discard{}),
					membership.WithTCPClient(&transport.Discard{}),
					membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
				)
				debugList := membership.DebugList(list)
				debugList.GetGossip().Clear()
				debugList.SetMembers(beforeMembers)
				debugList.SetFaultyMembers(beforeFaultyMembers)

				Expect(DispatchDatagram(list, message)).To(Succeed())

				if afterMembers == nil {
					Expect(debugList.GetMembers()).To(BeEmpty())
				} else {
					Expect(debugList.GetMembers()).To(Equal(afterMembers))
				}
				if afterFaultyMembers == nil {
					Expect(debugList.GetFaultyMembers()).To(BeEmpty())
				} else {
					Expect(debugList.GetFaultyMembers()).To(Equal(afterFaultyMembers))
				}
			},
			Entry("Suspect should add member",
				nil,
				nil,
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Suspect with lower incarnation number should NOT overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Suspect with same incarnation number should overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Suspect with bigger incarnation number should overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 3,
					},
				},
				nil,
			),

			Entry("Suspect with lower incarnation number should NOT overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Suspect with same incarnation number should NOT overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Suspect with bigger incarnation number should NOT overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 3,
					},
				},
				nil,
			),

			Entry("Suspect with lower incarnation number should NOT overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Suspect with same incarnation number should NOT overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Suspect with bigger incarnation number should overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 3,
					},
				},
				nil,
			),
		)
	})

	Context("handleFaulty", func() {
		It("should refute faulty about self", func() {
			list := newTestList()
			debugList := membership.DebugList(list)
			debugList.GetGossip().Clear()

			Expect(DispatchDatagram(list, encoding.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 55,
			}.ToMessage())).To(Succeed())

			Expect(debugList.GetMembers()).To(BeEmpty())
			Expect(debugList.GetFaultyMembers()).To(BeEmpty())
			Expect(debugList.GetGossip().Len()).To(Equal(1))
			Expect(debugList.GetGossip().Get(0)).To(Equal(encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 56,
			}.ToMessage()))
		})

		It("should invoke member removed callback when faulty", func() {
			var removedCount int
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithMemberRemovedCallback(func(address encoding.Address) {
					removedCount++
				}),
			)

			By("Sending faulty message for existing")
			Expect(DispatchDatagram(list, encoding.MessageFaulty{
				Destination:       bootstrapMembers[0],
				IncarnationNumber: 5,
			}.ToMessage())).To(Succeed())

			By("Verifying callback invoked")
			Expect(removedCount).To(Equal(1))
		})

		DescribeTable("Gossip should update the memberlist correctly",
			func(beforeMembers []encoding.Member, beforeFaultyMembers []encoding.Member, message encoding.Message, afterMembers []encoding.Member, afterFaultyMembers []encoding.Member) {
				list := membership.NewList(
					membership.WithUDPClient(&transport.Discard{}),
					membership.WithTCPClient(&transport.Discard{}),
					membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
				)
				debugList := membership.DebugList(list)
				debugList.GetGossip().Clear()
				debugList.SetMembers(beforeMembers)
				debugList.SetFaultyMembers(beforeFaultyMembers)

				Expect(DispatchDatagram(list, message)).To(Succeed())

				if afterMembers == nil {
					Expect(debugList.GetMembers()).To(BeEmpty())
				} else {
					Expect(debugList.GetMembers()).To(Equal(afterMembers))
				}
				if afterFaultyMembers == nil {
					Expect(debugList.GetFaultyMembers()).To(BeEmpty())
				} else {
					Expect(debugList.GetFaultyMembers()).To(Equal(afterFaultyMembers))
				}
			},
			Entry("Faulty should add faulty member",
				nil,
				nil,
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Faulty with lower incarnation number should NOT overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Faulty with same incarnation number should overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Faulty with bigger incarnation number should overwrite alive",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateAlive,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 3,
					},
				},
			),

			Entry("Faulty with lower incarnation number should NOT overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
			),
			Entry("Faulty with same incarnation number should overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Faulty with bigger incarnation number should overwrite suspect",
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateSuspect,
						IncarnationNumber: 2,
					},
				},
				nil,
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 3,
					},
				},
			),

			Entry("Faulty with lower incarnation number should NOT overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Faulty with same incarnation number should NOT overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
			),
			Entry("Faulty with bigger incarnation number should NOT overwrite faulty",
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 2,
					},
				},
				encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				}.ToMessage(),
				nil,
				[]encoding.Member{
					{
						Address:           TestAddress,
						State:             encoding.MemberStateFaulty,
						IncarnationNumber: 3,
					},
				},
			),
		)
	})

	Context("handleListRequest", func() {
		It("should send list response with empty member list", func() {
			var store transport.Store
			list := newTestList(
				membership.WithTCPClient(&store),
			)

			By("Sending list request")
			sourceAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)
			Expect(DispatchDatagram(list, encoding.MessageListRequest{
				Source: sourceAddr,
			}.ToMessage())).To(Succeed())

			By("Verifying list response sent back to source")
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).To(Equal(sourceAddr))
			Expect(store.Buffers).To(HaveLen(1))

			By("Verifying response message format")
			var responseMsg encoding.MessageListResponse
			Expect(responseMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(responseMsg.Members).To(BeEmpty())
		})

		It("should send list response with alive members", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
			}
			list := newTestList(
				membership.WithTCPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Sending list request")
			sourceAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4)
			Expect(DispatchDatagram(list, encoding.MessageListRequest{
				Source: sourceAddr,
			}.ToMessage())).To(Succeed())

			By("Verifying list response sent")
			Expect(store.Addresses).To(HaveLen(1))
			Expect(store.Addresses[0]).To(Equal(sourceAddr))

			By("Verifying response contains all members")
			var responseMsg encoding.MessageListResponse
			Expect(responseMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(responseMsg.Members).To(HaveLen(3))
			var responseAddresses []encoding.Address
			for _, member := range responseMsg.Members {
				responseAddresses = append(responseAddresses, member.Address)
			}
			Expect(responseAddresses).To(ConsistOf(bootstrapMembers))
		})

		It("should include suspect members in response", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithTCPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
			)

			By("Marking one member as suspect")
			Expect(list.DirectPing()).To(Succeed())
			Expect(list.EndOfProtocolPeriod()).To(Succeed())

			By("Sending list request")
			store.Clear()
			sourceAddr := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4)
			Expect(DispatchDatagram(list, encoding.MessageListRequest{
				Source: sourceAddr,
			}.ToMessage())).To(Succeed())

			By("Verifying response contains both alive and suspect members")
			var responseMsg encoding.MessageListResponse
			Expect(responseMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(responseMsg.Members).To(HaveLen(2))
			var responseAddresses []encoding.Address
			for _, member := range responseMsg.Members {
				responseAddresses = append(responseAddresses, member.Address)
			}
			Expect(responseAddresses).To(ConsistOf(bootstrapMembers))
		})

		It("should include faulty members in response", func() {
			var store transport.Store
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithTCPClient(&store),
				membership.WithBootstrapMembers(bootstrapMembers),
				membership.WithSafetyFactor(0),
			)

			By("Marking one member as faulty")
			Expect(list.DirectPing()).To(Succeed())
			Expect(list.EndOfProtocolPeriod()).To(Succeed())
			Expect(list.Len()).To(Equal(1))

			By("Sending list request")
			store.Clear()
			Expect(DispatchDatagram(list, encoding.MessageListRequest{
				Source: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
			}.ToMessage())).To(Succeed())

			By("Verifying response contains falty member")
			var responseMsg encoding.MessageListResponse
			Expect(responseMsg.FromBuffer(store.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(responseMsg.Members).To(HaveLen(2))
		})

		It("should handle TCP send errors gracefully", func() {
			list := newTestList(
				membership.WithTCPClient(&transport.Error{}),
			)

			By("Sending list request")
			Expect(DispatchDatagram(list, encoding.MessageListRequest{
				Source: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
			}.ToMessage())).ToNot(Succeed())
		})
	})

	Context("handleListResponse", func() {
		It("should add new members from response", func() {
			list := newTestList()
			debugList := membership.DebugList(list)

			By("Receiving list response with members")
			responseMembers := []encoding.Member{
				{
					Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 5,
				},
				{
					Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
					State:             encoding.MemberStateSuspect,
					IncarnationNumber: 3,
				},
				{
					Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 0,
				},
			}
			Expect(DispatchDatagram(list, encoding.MessageListResponse{
				Source:  encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				Members: responseMembers,
			}.ToMessage())).To(Succeed())

			By("Verifying members added with correct state and incarnation")
			Expect(list.Len()).To(Equal(3))
			members := debugList.GetMembers()
			Expect(members).To(HaveLen(3))
			Expect(members).To(ContainElement(responseMembers[0]))
			Expect(members).To(ContainElement(responseMembers[1]))
			Expect(members).To(ContainElement(responseMembers[2]))
		})

		It("should handle empty member list", func() {
			list := newTestList()

			By("Receiving list response with no members")
			Expect(DispatchDatagram(list, encoding.MessageListResponse{
				Source:  encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				Members: []encoding.Member{},
			}.ToMessage())).To(Succeed())

			By("Verifying no members added")
			Expect(list.Len()).To(Equal(0))
		})

		It("should not add self to member list", func() {
			list := newTestList()
			debugList := membership.DebugList(list)

			By("Receiving list response containing self")
			responseMembers := []encoding.Member{
				{
					Address:           TestAddress, // Self should be excluded
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 2,
				},
				{
					Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 0,
				},
				{
					Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 0,
				},
			}
			Expect(DispatchDatagram(list, encoding.MessageListResponse{
				Source:  encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				Members: responseMembers,
			}.ToMessage())).To(Succeed())

			By("Verifying self excluded")
			Expect(list.Len()).To(Equal(2))
			members := debugList.GetMembers()
			for _, member := range members {
				Expect(member.Address).NotTo(Equal(TestAddress))
			}
		})

		It("should not duplicate existing members", func() {
			bootstrapMembers := []encoding.Address{
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
				encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
			}
			list := newTestList(
				membership.WithBootstrapMembers(bootstrapMembers),
			)
			Expect(list.Len()).To(Equal(2))

			By("Receiving list response with overlapping members")
			responseMembers := []encoding.Member{
				{
					Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2), // Already exists
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 0,
				},
				{
					Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3), // New
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 0,
				},
			}
			Expect(DispatchDatagram(list, encoding.MessageListResponse{
				Source:  encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4),
				Members: responseMembers,
			}.ToMessage())).To(Succeed())

			By("Verifying total member count")
			Expect(list.Len()).To(Equal(3))
		})
	})

	It("newly joined member should propagate after a limited number of protocol periods", func() {
		for memberCount := range utility.ClusterSize(2, 8, 128) {
			By("Setting up the initial cluster")
			memoryTransport := transport.NewMemory()
			var lists []*membership.List
			for i := range memberCount {
				address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), i+1)
				options := []membership.Option{
					membership.WithLogger(GinkgoLogr.WithValues("list", address)),
					membership.WithAdvertisedAddress(address),
					membership.WithUDPClient(memoryTransport.Client()),
					membership.WithTCPClient(memoryTransport.Client()),
					membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
				}
				for j := range memberCount {
					options = append(options,
						membership.WithBootstrapMember(encoding.NewAddress(net.IPv4(255, 255, 255, 255), j+1)),
					)
				}
				newList := membership.NewList(options...)
				debugList := membership.DebugList(newList)
				debugList.ClearGossip()
				memoryTransport.AddTarget(address, newList)
				lists = append(lists, newList)
			}
			for _, list := range lists {
				Expect(list.Len()).To(Equal(memberCount - 1))
			}

			By("joining a new member")
			address := encoding.NewAddress(net.IPv4(255, 255, 255, 255), math.MaxUint16)
			options := []membership.Option{
				membership.WithLogger(GinkgoLogr.WithValues("list", address)),
				membership.WithAdvertisedAddress(address),
				membership.WithUDPClient(memoryTransport.Client()),
				membership.WithTCPClient(memoryTransport.Client()),
				membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
			}
			options = append(options,
				membership.WithBootstrapMember(encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)),
			)
			newList := membership.NewList(options...)
			memoryTransport.AddTarget(address, newList)
			lists = append(lists, newList)

			periodCount := int(math.Ceil(utility.DisseminationPeriods(membership.DefaultConfig.SafetyFactor, len(lists))))
			for i := range max(periodCount, 128*len(lists)) {
				GinkgoLogr.Info("> Start of protocol period", "period", i)
				By("Executing direct ping")
				for _, list := range lists {
					Expect(list.DirectPing()).To(Succeed())
				}
				Expect(memoryTransport.FlushAllPendingSends()).To(Succeed())

				By("Executing indirect ping")
				for _, list := range lists {
					Expect(list.IndirectPing()).To(Succeed())
				}
				Expect(memoryTransport.FlushAllPendingSends()).To(Succeed())

				if i > 0 && i%(periodCount*2) == 0 {
					By("Executing request list")
					for _, list := range lists {
						Expect(list.RequestList()).To(Succeed())
					}
				}

				By("Executing end of protocol period")
				for _, list := range lists {
					Expect(list.EndOfProtocolPeriod()).To(Succeed())
				}

				propagationDone := true
				for _, list := range lists[:len(lists)-1] {
					if list.Len() != memberCount {
						propagationDone = false
					}
				}
				if propagationDone {
					// Exit early when we already have propagated to all.
					break
				}
			}
			for _, list := range lists[:len(lists)-1] {
				Expect(list.Len()).To(Equal(memberCount))
			}
		}
	})
})

func BenchmarkList_All(b *testing.B) {
	executeFunctionWithMembers(b, func(list *membership.List) {
		list.ForEach(func(address encoding.Address) bool {
			return true
		})
	})
}

func BenchmarkList_DirectPing(b *testing.B) {
	executeFunctionWithMembers(b, func(list *membership.List) {
		if err := list.DirectPing(); err != nil {
			b.Fatal(err)
		}
	})
}

func BenchmarkList_IndirectPing(b *testing.B) {
	executeFunctionWithMembers(b, func(list *membership.List) {
		if err := list.IndirectPing(); err != nil {
			b.Fatal(err)
		}
	})
}

func BenchmarkList_EndOfProtocolPeriod(b *testing.B) {
	executeFunctionWithMembers(b, func(list *membership.List) {
		if err := list.EndOfProtocolPeriod(); err != nil {
			b.Fatal(err)
		}
	})
}

func BenchmarkList_RequestList(b *testing.B) {
	executeFunctionWithMembers(b, func(list *membership.List) {
		if err := list.RequestList(); err != nil {
			b.Fatal(err)
		}
	})
}

func BenchmarkList_BroadcastShutdown(b *testing.B) {
	executeFunctionWithMembers(b, func(list *membership.List) {
		if err := list.BroadcastShutdown(); err != nil {
			b.Fatal(err)
		}
	})
}

func BenchmarkList_handleDirectPing(b *testing.B) {
	message := encoding.MessageDirectPing{
		Source:         BenchmarkAddress,
		SequenceNumber: 0,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleDirectAck(b *testing.B) {
	message := encoding.MessageDirectAck{
		Source:         BenchmarkAddress,
		SequenceNumber: 0,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleIndirectPing(b *testing.B) {
	message := encoding.MessageIndirectPing{
		Source:         TestAddress,
		Destination:    TestAddress2,
		SequenceNumber: 0,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleIndirectAck(b *testing.B) {
	message := encoding.MessageIndirectAck{
		Source:         BenchmarkAddress,
		SequenceNumber: 0,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleSuspect(b *testing.B) {
	message := encoding.MessageSuspect{
		Source:            TestAddress,
		Destination:       TestAddress2,
		IncarnationNumber: 0,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleAlive(b *testing.B) {
	message := encoding.MessageAlive{
		Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+32000),
		IncarnationNumber: 0,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleFaulty(b *testing.B) {
	message := encoding.MessageFaulty{
		Source:            BenchmarkAddress,
		Destination:       TestAddress,
		IncarnationNumber: 0,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleListRequest(b *testing.B) {
	message := encoding.MessageListRequest{
		Source: TestAddress,
	}.ToMessage()
	dispatchDatagramWithMembers(b, message)
}

func BenchmarkList_handleListResponse(b *testing.B) {
	for memberCount := range utility.ClusterSize(2, 8, 128) {
		responseMembers := make([]encoding.Member, memberCount)
		for i := range memberCount {
			responseMembers[i] = encoding.Member{
				Address:           encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1+i),
				State:             encoding.MemberStateAlive,
				IncarnationNumber: 0,
			}
		}

		message := encoding.MessageListResponse{
			Source:  TestAddress,
			Members: responseMembers,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("%d response members", memberCount), func(b *testing.B) {
			for range b.N {
				list := createListWithMembers(0)
				if err := list.DispatchDatagram(buffer); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func dispatchDatagramWithMembers(b *testing.B, message encoding.Message) {
	buffer, _, err := message.AppendToBuffer(nil)
	if err != nil {
		b.Fatal(err)
	}
	executeFunctionWithMembers(b, func(list *membership.List) {
		if err := list.DispatchDatagram(buffer); err != nil {
			b.Fatal(err)
		}
	})
}

func executeFunctionWithMembers(b *testing.B, f func(list *membership.List)) {
	for memberCount := range utility.ClusterSize(2, 8, 128) {
		list := createListWithMembers(memberCount)
		b.Run(fmt.Sprintf("%d members", memberCount), func(b *testing.B) {
			for range b.N {
				f(list)
			}
		})
	}
}

func createListWithMembers(memberCount int) *membership.List {
	list := membership.NewList(
		membership.WithUDPClient(&transport.Discard{}),
		membership.WithTCPClient(&transport.Discard{}),
		membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
	)
	debugList := membership.DebugList(list)

	for i := range memberCount {
		messageAlive := encoding.MessageAlive{
			Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
			IncarnationNumber: 0,
		}
		buffer, _, err := messageAlive.AppendToBuffer(nil)
		if err != nil {
			panic(err)
		}
		if err := list.DispatchDatagram(buffer); err != nil {
			panic(err)
		}
	}
	if list.Len() != memberCount {
		panic("member count does not match expected value")
	}
	debugList.GetGossip().Clear()
	return list
}

func newTestList(options ...membership.Option) *membership.List {
	defaultOptions := []membership.Option{
		membership.WithLogger(GinkgoLogr),
		membership.WithAdvertisedAddress(TestAddress),
		membership.WithUDPClient(&transport.Discard{}),
		membership.WithTCPClient(&transport.Discard{}),
		membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
	}
	allOptions := append(defaultOptions, options...)
	return membership.NewList(allOptions...)
}
