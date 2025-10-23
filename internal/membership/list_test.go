package membership_test

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List", func() {
	var list *membership.List

	BeforeEach(func() {
		list = membership.NewList(
			membership.WithLogger(GinkgoLogr),
			membership.WithUDPClient(&DiscardClient{}),
			membership.WithTCPClient(&DiscardClient{}),
			membership.WithAdvertisedAddress(TestAddress),
		)
		list.GetGossip().Clear()
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
			membership.WithUDPClient(&DiscardClient{}),
			membership.WithTCPClient(&DiscardClient{}),
			membership.WithMemberAddedCallback(func(address encoding.Address) {
				membersAdded.Add(1)
				callbacks.Done()
			}),
			membership.WithMemberRemovedCallback(func(address encoding.Address) {
				membersRemoved.Add(1)
				callbacks.Done()
			}),
		)
		list.GetGossip().Clear()

		for i := range 10 {
			messageAlive := gossip.MessageAlive{
				Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
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
		var storeClient StoreClient
		list := membership.NewList(
			membership.WithLogger(GinkgoLogr),
			membership.WithUDPClient(&storeClient),
			membership.WithTCPClient(&DiscardClient{}),
			membership.WithAdvertisedAddress(TestAddress),
		)
		list.GetGossip().Clear()

		Expect(list.DirectPing()).To(Succeed())
		Expect(storeClient.Addresses).To(HaveLen(0))
	})

	It("should do round robin direct pings", func() {
		var storeClient StoreClient
		list := membership.NewList(
			membership.WithLogger(GinkgoLogr),
			membership.WithUDPClient(&storeClient),
			membership.WithTCPClient(&DiscardClient{}),
			membership.WithAdvertisedAddress(TestAddress),
		)
		list.GetGossip().Clear()

		// Add a few members
		for i := range 10 {
			message := gossip.MessageAlive{
				Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+1+i),
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
		Expect(list.GetGossip().Len()).To(Equal(0))
		message := gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(list.DispatchDatagram(buffer)).To(Succeed())

		Expect(list.GetMembers()).To(BeEmpty())
		Expect(list.GetFaultyMembers()).To(BeEmpty())
		Expect(list.GetGossip().Len()).To(Equal(0))
	})

	It("should refute suspect about self", func() {
		Expect(list.GetGossip().Len()).To(Equal(0))
		message := gossip.MessageSuspect{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(list.DispatchDatagram(buffer)).To(Succeed())

		Expect(list.GetMembers()).To(BeEmpty())
		Expect(list.GetFaultyMembers()).To(BeEmpty())
		Expect(list.GetGossip().Len()).To(Equal(1))
		Expect(list.GetGossip().Get(0)).To(Equal(&gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 1,
		}))
	})

	It("should refute faulty about self", func() {
		Expect(list.GetGossip().Len()).To(Equal(0))
		message := gossip.MessageFaulty{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(list.DispatchDatagram(buffer)).To(Succeed())

		Expect(list.GetMembers()).To(BeEmpty())
		Expect(list.GetFaultyMembers()).To(BeEmpty())
		Expect(list.GetGossip().Len()).To(Equal(1))
		Expect(list.GetGossip().Get(0)).To(Equal(&gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 1,
		}))
	})

	DescribeTable("Gossip should update the memberlist correctly",
		func(beforeMembers []encoding.Member, beforeFaultyMembers []encoding.Member, message gossip.Message, afterMembers []encoding.Member, afterFaultyMembers []encoding.Member) {
			list := membership.NewList(
				membership.WithUDPClient(&DiscardClient{}),
				membership.WithTCPClient(&DiscardClient{}),
			)
			list.GetGossip().Clear()
			list.SetMembers(beforeMembers)
			list.SetFaultyMembers(beforeFaultyMembers)

			buffer, _, err := message.AppendToBuffer(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.DispatchDatagram(buffer)).To(Succeed())

			if afterMembers == nil {
				Expect(cleanLastStateChange(list.GetMembers())).To(BeEmpty())
			} else {
				Expect(cleanLastStateChange(list.GetMembers())).To(Equal(afterMembers))
			}
			if afterFaultyMembers == nil {
				Expect(cleanLastStateChange(list.GetFaultyMembers())).To(BeEmpty())
			} else {
				Expect(cleanLastStateChange(list.GetFaultyMembers())).To(Equal(afterFaultyMembers))
			}
		},
		Entry("Alive should add a member",
			nil,
			nil,
			&gossip.MessageAlive{
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
				Source:            TestAddress,
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
		Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+32000),
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
		membership.WithUDPClient(&DiscardClient{}),
		membership.WithTCPClient(&DiscardClient{}),
	)

	for i := range memberCount {
		messageAlive := gossip.MessageAlive{
			Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
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
	list.GetGossip().Clear()
	return list
}

func cleanLastStateChange(members []encoding.Member) []encoding.Member {
	for i := range members {
		members[i].LastStateChange = time.Time{}
	}
	return members
}
