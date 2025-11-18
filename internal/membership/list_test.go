package membership_test

import (
	"fmt"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/backbone81/membership/internal/utility"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
	"github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/transport"
)

var _ = Describe("List", func() {
	Context("NewList", func() {
		It("should create a list with default configuration", func() {
			list := newTestList()

			Expect(list).NotTo(BeNil())
			Expect(list.Len()).To(Equal(0))
			Expect(list.Get()).To(BeEmpty())
			Expect(membership.DebugList(list).GetMembers()).To(BeEmpty())
			Expect(membership.DebugList(list).GetFaultyMembers()).To(BeEmpty())
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

			Expect(list.Len()).To(Equal(3))
			members := membership.DebugList(list).GetMembers()
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

			Expect(list.Len()).To(Equal(2), "Self should be excluded from member list")
			members := membership.DebugList(list).GetMembers()

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
			var addedMutex sync.Mutex
			var addedAddresses []encoding.Address

			_ = newTestList(
				membership.WithMemberAddedCallback(func(address encoding.Address) {
					addedMutex.Lock()
					defer addedMutex.Unlock()
					addedCount++
					addedAddresses = append(addedAddresses, address)
				}),
				membership.WithBootstrapMembers([]encoding.Address{
					encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1),
					encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2),
				}),
			)

			// Callbacks should have been invoked for bootstrap members
			Eventually(func() int {
				addedMutex.Lock()
				defer addedMutex.Unlock()
				return addedCount
			}).Should(Equal(2))

			addedMutex.Lock()
			defer addedMutex.Unlock()
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

			addresses := list.Get()
			Expect(addresses).To(HaveLen(3))

			// Verify sorted ascending
			for i := 0; i < len(addresses)-1; i++ {
				Expect(encoding.CompareAddress(addresses[i], addresses[i+1])).To(Equal(-1),
					"Addresses should be sorted ascending")
			}
		})

		It("should initialize with gossip about self", func() {
			list := newTestList()

			gossipQueue := membership.DebugList(list).GetGossip()
			Expect(gossipQueue.Len()).To(Equal(1))

			msg := membership.DebugList(list).GetGossip().Get(0)
			aliveMsg, ok := msg.(*gossip.MessageAlive)
			Expect(ok).To(BeTrue())
			Expect(aliveMsg.Destination).To(Equal(TestAddress))
			Expect(aliveMsg.IncarnationNumber).To(Equal(uint16(0)))
		})
	})

	// ========== OLD TESTS ==========

	var list *membership.List

	BeforeEach(func() {
		list = membership.NewList(
			membership.WithLogger(GinkgoLogr),
			membership.WithUDPClient(&transport.Discard{}),
			membership.WithTCPClient(&transport.Discard{}),
			membership.WithAdvertisedAddress(TestAddress),
			membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
		)
		membership.DebugList(list).GetGossip().Clear()
	})

	It("should return the member list", func() {
		list := createListWithMembers(8)
		Expect(list.Get()).To(HaveLen(8))
		Expect(list.Len()).To(Equal(8))
	})

	It("should trigger the member callbacks", func() {
		var membersAdded, membersRemoved atomic.Int64
		var callbacks sync.WaitGroup

		list := membership.NewList(
			membership.WithLogger(GinkgoLogr),
			membership.WithUDPClient(&transport.Discard{}),
			membership.WithTCPClient(&transport.Discard{}),
			membership.WithMemberAddedCallback(func(address encoding.Address) {
				membersAdded.Add(1)
				callbacks.Done()
			}),
			membership.WithMemberRemovedCallback(func(address encoding.Address) {
				membersRemoved.Add(1)
				callbacks.Done()
			}),
			membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
		)
		membership.DebugList(list).GetGossip().Clear()

		for i := range 10 {
			messageAlive := gossip.MessageAlive{
				Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			}
			buffer, _, err := messageAlive.AppendToBuffer(nil)
			Expect(err).ToNot(HaveOccurred())
			callbacks.Add(1)
			Expect(list.DispatchDatagram(buffer)).To(Succeed())
		}
		callbacks.Wait()
		Expect(int(membersAdded.Load())).To(Equal(10))
		Expect(int(membersRemoved.Load())).To(Equal(0))

		for i := range 10 {
			messageAlive := gossip.MessageFaulty{
				Source:            TestAddress,
				Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			}
			buffer, _, err := messageAlive.AppendToBuffer(nil)
			Expect(err).ToNot(HaveOccurred())
			callbacks.Add(1)
			Expect(list.DispatchDatagram(buffer)).To(Succeed())
		}
		callbacks.Wait()
		Expect(int(membersAdded.Load())).To(Equal(10))
		Expect(int(membersRemoved.Load())).To(Equal(10))
	})

	It("should not do a ping without members", func() {
		var storeClient transport.Store
		list := membership.NewList(
			membership.WithLogger(GinkgoLogr),
			membership.WithUDPClient(&storeClient),
			membership.WithTCPClient(&transport.Discard{}),
			membership.WithAdvertisedAddress(TestAddress),
			membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
		)
		membership.DebugList(list).GetGossip().Clear()

		Expect(list.DirectPing()).To(Succeed())
		Expect(storeClient.Addresses).To(BeEmpty())
	})

	It("should do round robin direct pings", func() {
		var storeClient transport.Store
		list := membership.NewList(
			membership.WithLogger(GinkgoLogr),
			membership.WithUDPClient(&storeClient),
			membership.WithTCPClient(&transport.Discard{}),
			membership.WithAdvertisedAddress(TestAddress),
			membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
		)
		membership.DebugList(list).GetGossip().Clear()

		// Add a few members
		for i := range 10 {
			message := gossip.MessageAlive{
				Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+1+i),
				IncarnationNumber: 0,
			}
			buffer, _, err := message.AppendToBuffer(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.DispatchDatagram(buffer)).To(Succeed())
		}

		// Catch the direct ping messages and remember which address was pinged how often
		directPingAddresses := make(map[encoding.Address]int)
		for range 100 {
			Expect(list.DirectPing()).To(Succeed())

			Expect(storeClient.Addresses).To(HaveLen(1))
			var message membership.MessageDirectPing
			Expect(message.FromBuffer(storeClient.Buffers[0])).Error().ToNot(HaveOccurred())
			Expect(message.Source).To(Equal(TestAddress))
			directPingAddresses[storeClient.Addresses[0]]++

			storeClient.Addresses = nil
			storeClient.Buffers = nil
		}

		// We expect every address to be pinged the same number of times.
		Expect(directPingAddresses).To(HaveLen(10))
		for _, value := range directPingAddresses {
			Expect(value).To(Equal(10))
		}
	})

	It("should ignore alive about self", func() {
		Expect(membership.DebugList(list).GetGossip().Len()).To(Equal(0))
		message := gossip.MessageAlive{
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(list.DispatchDatagram(buffer)).To(Succeed())

		Expect(membership.DebugList(list).GetMembers()).To(BeEmpty())
		Expect(membership.DebugList(list).GetFaultyMembers()).To(BeEmpty())
		Expect(membership.DebugList(list).GetGossip().Len()).To(Equal(0))
	})

	It("should refute suspect about self", func() {
		Expect(membership.DebugList(list).GetGossip().Len()).To(Equal(0))
		message := gossip.MessageSuspect{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(list.DispatchDatagram(buffer)).To(Succeed())

		Expect(membership.DebugList(list).GetMembers()).To(BeEmpty())
		Expect(membership.DebugList(list).GetFaultyMembers()).To(BeEmpty())
		Expect(membership.DebugList(list).GetGossip().Len()).To(Equal(1))
		Expect(membership.DebugList(list).GetGossip().Get(0)).To(Equal(&gossip.MessageAlive{
			Destination:       TestAddress,
			IncarnationNumber: 1,
		}))
	})

	It("should refute faulty about self", func() {
		Expect(membership.DebugList(list).GetGossip().Len()).To(Equal(0))
		message := gossip.MessageFaulty{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(list.DispatchDatagram(buffer)).To(Succeed())

		Expect(membership.DebugList(list).GetMembers()).To(BeEmpty())
		Expect(membership.DebugList(list).GetFaultyMembers()).To(BeEmpty())
		Expect(membership.DebugList(list).GetGossip().Len()).To(Equal(1))
		Expect(membership.DebugList(list).GetGossip().Get(0)).To(Equal(&gossip.MessageAlive{
			Destination:       TestAddress,
			IncarnationNumber: 1,
		}))
	})

	DescribeTable("Gossip should update the memberlist correctly",
		func(beforeMembers []encoding.Member, beforeFaultyMembers []encoding.Member, message gossip.Message, afterMembers []encoding.Member, afterFaultyMembers []encoding.Member) {
			list := membership.NewList(
				membership.WithUDPClient(&transport.Discard{}),
				membership.WithTCPClient(&transport.Discard{}),
				membership.WithRoundTripTimeTracker(roundtriptime.NewTracker()),
			)
			membership.DebugList(list).GetGossip().Clear()
			membership.DebugList(list).SetMembers(beforeMembers)
			membership.DebugList(list).SetFaultyMembers(beforeFaultyMembers)

			buffer, _, err := message.AppendToBuffer(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.DispatchDatagram(buffer)).To(Succeed())

			if afterMembers == nil {
				Expect(membership.DebugList(list).GetMembers()).To(BeEmpty())
			} else {
				Expect(membership.DebugList(list).GetMembers()).To(Equal(afterMembers))
			}
			if afterFaultyMembers == nil {
				Expect(membership.DebugList(list).GetFaultyMembers()).To(BeEmpty())
			} else {
				Expect(membership.DebugList(list).GetFaultyMembers()).To(Equal(afterFaultyMembers))
			}
		},
		Entry("Alive should add a member",
			nil,
			nil,
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateAlive,
					IncarnationNumber: 3,
				},
			},
			nil,
		),
		Entry("Suspect should add member",
			nil,
			nil,
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateSuspect,
					IncarnationNumber: 3,
				},
			},
			nil,
		),
		Entry("Faulty should add faulty member",
			nil,
			nil,
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			nil,
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateFaulty,
					IncarnationNumber: 3,
				},
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateAlive,
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateSuspect,
					IncarnationNumber: 3,
				},
			},
			nil,
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			nil,
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateFaulty,
					IncarnationNumber: 3,
				},
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateAlive,
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			[]encoding.Member{
				{
					Address:           TestAddress,
					State:             encoding.MemberStateSuspect,
					IncarnationNumber: 3,
				},
			},
			nil,
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
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
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
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

	It("should remove a member after some time when no response", func() {
		// TODO: implementation
	})

	// This test is flaky and therefore disabled.
	PIt("newly joined member should propagate after a limited number of protocol periods", func() {
		for memberCount := 1; memberCount <= 256; memberCount *= 2 {
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
				membership.DebugList(newList).ClearGossip()
				memoryTransport.AddTarget(address, newList)
				lists = append(lists, newList)
			}
			for _, list := range lists {
				Expect(list.Get()).To(HaveLen(memberCount - 1))
			}

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

			// TODO: This is a bad test. It is so unreliable, think about something better.
			periodCount := int(math.Ceil(utility.DisseminationPeriods(membership.DefaultConfig.SafetyFactor*16, len(lists))))
			for i := range periodCount {
				GinkgoLogr.Info("> Start of protocol period", "period", i)
				GinkgoLogr.Info("> Executing direct pings")
				for _, list := range lists {
					Expect(list.DirectPing()).To(Succeed())
				}
				Expect(memoryTransport.FlushAllPendingSends()).To(Succeed())

				GinkgoLogr.Info("> Executing indirect pings")
				for _, list := range lists {
					Expect(list.IndirectPing()).To(Succeed())
				}
				Expect(memoryTransport.FlushAllPendingSends()).To(Succeed())

				GinkgoLogr.Info("> Executing end of protocol period")
				for _, list := range lists {
					Expect(list.EndOfProtocolPeriod()).To(Succeed())
				}
			}

			for _, list := range lists[:len(lists)-1] {
				Expect(list.Get()).To(HaveLen(memberCount))
			}
		}
	})
})

func BenchmarkList_Get(b *testing.B) {
	executeFunctionWithMembers(b, func(list *membership.List) {
		_ = list.Get()
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

func BenchmarkList_handleDirectPing(b *testing.B) {
	message := membership.MessageDirectPing{
		Source:         BenchmarkAddress,
		SequenceNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleDirectAck(b *testing.B) {
	message := membership.MessageDirectAck{
		Source:         BenchmarkAddress,
		SequenceNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleIndirectPing(b *testing.B) {
	message := membership.MessageIndirectPing{
		Source:         TestAddress,
		Destination:    TestAddress2,
		SequenceNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleIndirectAck(b *testing.B) {
	message := membership.MessageIndirectAck{
		Source:         BenchmarkAddress,
		SequenceNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleSuspect(b *testing.B) {
	message := gossip.MessageSuspect{
		Source:            TestAddress,
		Destination:       TestAddress2,
		IncarnationNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleAlive(b *testing.B) {
	message := gossip.MessageAlive{
		Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+32000),
		IncarnationNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleFaulty(b *testing.B) {
	message := gossip.MessageFaulty{
		Source:            BenchmarkAddress,
		IncarnationNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleListRequest(b *testing.B) {
	message := membership.MessageListRequest{
		Source: TestAddress,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleListResponse(b *testing.B) {
	message := membership.MessageListResponse{
		Source:  TestAddress,
		Members: nil, // TODO: We should setup the test in a way which allows us to send different member counts
	}
	dispatchDatagramWithMembers(b, &message)
}

func dispatchDatagramWithMembers(b *testing.B, message membership.Message) {
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
	for memberCount := 1; memberCount <= 16*1024; memberCount *= 2 {
		list := createListWithMembers(memberCount)
		b.Run(fmt.Sprintf("%d members", memberCount), func(b *testing.B) {
			for b.Loop() {
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

	for i := range memberCount {
		messageAlive := gossip.MessageAlive{
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
	membership.DebugList(list).GetGossip().Clear()
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
