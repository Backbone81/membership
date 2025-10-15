package membership_test

import (
	"fmt"
	"net"
	"testing"

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
		)
	})

	It("should return the member list", func() {
		Expect(list).ToNot(BeNil())
	})

	It("should trigger the member callbacks", func() {

	})

	It("should correctly add a member when gossiped alive", func() {

	})

	It("should keep a member when gossiped suspect", func() {

	})

	It("should remove a member when gossiped faulty", func() {

	})

	It("should remove a member after some time when no response", func() {

	})
})

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
		Source:         TestAddress,
		SequenceNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleDirectAck(b *testing.B) {
	message := membership.MessageDirectAck{
		Source:         TestAddress,
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
		Source:         TestAddress,
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
		Source:            TestAddress,
		IncarnationNumber: 0,
	}
	dispatchDatagramWithMembers(b, &message)
}

func BenchmarkList_handleFaulty(b *testing.B) {
	message := gossip.MessageFaulty{
		Source:            TestAddress,
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
		list := createListWithMembers(b, memberCount)
		b.Run(fmt.Sprintf("%d members", memberCount), func(b *testing.B) {
			for b.Loop() {
				f(list)
			}
		})
	}
}

func createListWithMembers(b *testing.B, memberCount int) *membership.List {
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
			b.Fatal(err)
		}
		if err := list.DispatchDatagram(buffer); err != nil {
			b.Fatal(err)
		}
	}

	// TODO: Check if we now have the correct member count

	return list
}
