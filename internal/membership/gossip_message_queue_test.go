package membership_test

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GossipMessageQueue", func() {
	var queue *membership.GossipMessageQueue

	BeforeEach(func() {
		queue = membership.NewGossipMessageQueue(math.MaxInt)
		Expect(queue.Len()).To(Equal(0))
	})

	It("should add gossip", func() {
		gossipMessage := membership.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage)

		Expect(queue.Len()).To(Equal(1))
		Expect(queue.Get(0)).To(Equal(&gossipMessage))
	})

	It("should add gossip with different address", func() {
		gossipMessage1 := membership.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage1)

		gossipMessage2 := membership.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage2)

		Expect(queue.Len()).To(Equal(2))
		Expect(queue.Get(0)).To(Equal(&gossipMessage1))
		Expect(queue.Get(1)).To(Equal(&gossipMessage2))
	})

	It("should not add duplicate gossip", func() {
		gossipMessage1 := membership.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage1)

		gossipMessage2 := membership.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage2)

		Expect(queue.Len()).To(Equal(1))
		Expect(queue.Get(0)).To(Equal(&gossipMessage1))
	})

	DescribeTable("Messages should overwrite in the correct priority",
		func(message1 membership.GossipMessage, message2 membership.GossipMessage, overwrite bool) {
			gossipQueue := membership.NewGossipMessageQueue(math.MaxInt)
			gossipQueue.Add(message1)
			gossipQueue.Add(message2)
			if overwrite {
				Expect(gossipQueue.Get(0)).To(Equal(message2))
			} else {
				Expect(gossipQueue.Get(0)).To(Equal(message1))
			}
		},
		Entry("Alive with lower incarnation number should NOT overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Alive with same incarnation number should NOT overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Alive with bigger incarnation number should overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Suspect with lower incarnation number should NOT overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Suspect with same incarnation number should overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			true,
		),
		Entry("Suspect with bigger incarnation number should overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Faulty with lower incarnation number should NOT overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with same incarnation number should overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			true,
		),
		Entry("Faulty with bigger incarnation number should overwrite alive",
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),

		Entry("Alive with lower incarnation number should NOT overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Alive with same incarnation number should NOT overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Alive with bigger incarnation number should overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Suspect with lower incarnation number should NOT overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Suspect with same incarnation number should NOT overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Suspect with bigger incarnation number should overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Faulty with lower incarnation number should NOT overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with same incarnation number should overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			true,
		),
		Entry("Faulty with bigger incarnation number should overwrite suspect",
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),

		Entry("Alive with lower incarnation number should NOT overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Alive with same incarnation number should NOT overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Alive with bigger incarnation number should overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Suspect with lower incarnation number should NOT overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Suspect with same incarnation number should NOT overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Suspect with bigger incarnation number should overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Faulty with lower incarnation number should NOT overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with same incarnation number should NOT overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with bigger incarnation number should overwrite faulty",
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&membership.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
	)

	It("should correctly sort gossip by gossip count when preparing", func() {
		message1 := &membership.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &membership.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &membership.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkTransmitted(2)

		queue.PrioritizeForAddress(membership.Address{})
		Expect(queue.Get(0)).To(Equal(message3))
		Expect(queue.Get(1)).To(Equal(message1))
		Expect(queue.Get(2)).To(Equal(message2))
	})

	It("should correctly prioritize suspect gossip when preparing", func() {
		message1 := &membership.MessageSuspect{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &membership.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &membership.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkTransmitted(2)

		queue.PrioritizeForAddress(TestAddress)
		Expect(queue.Get(0)).To(Equal(message1))
		Expect(queue.Get(1)).To(Equal(message3))
		Expect(queue.Get(2)).To(Equal(message2))
	})

	It("should correctly prioritize faulty gossip when preparing", func() {
		message1 := &membership.MessageFaulty{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &membership.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &membership.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkTransmitted(2)

		queue.PrioritizeForAddress(TestAddress)
		Expect(queue.Get(0)).To(Equal(message1))
		Expect(queue.Get(1)).To(Equal(message3))
		Expect(queue.Get(2)).To(Equal(message2))
	})

	It("should correctly prioritize alive gossip when preparing", func() {
		message1 := &membership.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &membership.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &membership.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkTransmitted(2)

		queue.PrioritizeForAddress(TestAddress)
		Expect(queue.Get(0)).To(Equal(message3))
		Expect(queue.Get(1)).To(Equal(message2))
		Expect(queue.Get(2)).To(Equal(message1))
	})

	It("should correctly remove messages which were transmitted enough", func() {
		gossipQueue := membership.NewGossipMessageQueue(3)
		gossipQueue.Add(&membership.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		})
		gossipQueue.PrioritizeForAddress(membership.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkTransmitted(0)
		gossipQueue.PrioritizeForAddress(membership.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkTransmitted(0)
		gossipQueue.PrioritizeForAddress(membership.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkTransmitted(0)
		gossipQueue.PrioritizeForAddress(membership.Address{})
		Expect(gossipQueue.Len()).To(Equal(0))
	})
})

func BenchmarkGossipQueue_Add(b *testing.B) {
	for gossipCount := 1; gossipCount <= 16*1024; gossipCount *= 2 {
		gossipQueue := membership.NewGossipMessageQueue(math.MaxInt)
		for i := range gossipCount {
			gossipQueue.Add(&membership.MessageAlive{
				Source:            membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			for b.Loop() {
				gossipQueue.Add(&membership.MessageAlive{
					Source:            membership.NewAddress(net.IPv4(11, 12, 13, 14), 1024),
					IncarnationNumber: 0,
				})
				//gossipQueue.UndoAdd()
			}
		})
	}
}

func BenchmarkGossipQueue_PrepareFor(b *testing.B) {
	for gossipCount := 1; gossipCount <= 16*1024; gossipCount *= 2 {
		gossipQueue := membership.NewGossipMessageQueue(math.MaxInt)
		for i := range gossipCount {
			gossipQueue.Add(&membership.MessageAlive{
				Source:            membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			for b.Loop() {
				gossipQueue.PrioritizeForAddress(membership.Address{})
			}
		})
	}
}
