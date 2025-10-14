package gossip_test

import (
	"fmt"
	"math"
	"net"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageQueue", func() {
	var queue *gossip.MessageQueue

	BeforeEach(func() {
		queue = gossip.NewMessageQueue(math.MaxInt)
		Expect(queue.Len()).To(Equal(0))
	})

	It("should add gossip", func() {
		gossipMessage := gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage)

		Expect(queue.Len()).To(Equal(1))
		Expect(queue.Get(0)).To(Equal(&gossipMessage))
	})

	It("should add gossip with different address", func() {
		gossipMessage1 := gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage1)

		gossipMessage2 := gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage2)

		Expect(queue.Len()).To(Equal(2))
		Expect(queue.Get(0)).To(Equal(&gossipMessage1))
		Expect(queue.Get(1)).To(Equal(&gossipMessage2))
	})

	It("should not add duplicate gossip", func() {
		gossipMessage1 := gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage1)

		gossipMessage2 := gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(&gossipMessage2)

		Expect(queue.Len()).To(Equal(1))
		Expect(queue.Get(0)).To(Equal(&gossipMessage1))
	})

	DescribeTable("Messages should overwrite in the correct priority",
		func(message1 gossip.Message, message2 gossip.Message, overwrite bool) {
			gossipQueue := gossip.NewMessageQueue(math.MaxInt)
			gossipQueue.Add(message1)
			gossipQueue.Add(message2)
			if overwrite {
				Expect(gossipQueue.Get(0)).To(Equal(message2))
			} else {
				Expect(gossipQueue.Get(0)).To(Equal(message1))
			}
		},
		Entry("Alive with lower incarnation number should NOT overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Alive with same incarnation number should NOT overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Alive with bigger incarnation number should overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Suspect with lower incarnation number should NOT overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Suspect with same incarnation number should overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			true,
		),
		Entry("Suspect with bigger incarnation number should overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Faulty with lower incarnation number should NOT overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with same incarnation number should overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			true,
		),
		Entry("Faulty with bigger incarnation number should overwrite alive",
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),

		Entry("Alive with lower incarnation number should NOT overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Alive with same incarnation number should NOT overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Alive with bigger incarnation number should overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Suspect with lower incarnation number should NOT overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Suspect with same incarnation number should NOT overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Suspect with bigger incarnation number should overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Faulty with lower incarnation number should NOT overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with same incarnation number should overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			true,
		),
		Entry("Faulty with bigger incarnation number should overwrite suspect",
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),

		Entry("Alive with lower incarnation number should NOT overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Alive with same incarnation number should NOT overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Alive with bigger incarnation number should overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageAlive{
				Source:            TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Suspect with lower incarnation number should NOT overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Suspect with same incarnation number should NOT overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			false,
		),
		Entry("Suspect with bigger incarnation number should overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
		Entry("Faulty with lower incarnation number should NOT overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with same incarnation number should NOT overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1,
			},
			false,
		),
		Entry("Faulty with bigger incarnation number should overwrite faulty",
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 2,
			},
			&gossip.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 3,
			},
			true,
		),
	)

	It("should correctly sort gossip by gossip count when preparing", func() {
		message1 := &gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &gossip.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkTransmitted(2)

		queue.PrioritizeForAddress(encoding.Address{})
		Expect(queue.Get(0)).To(Equal(message3))
		Expect(queue.Get(1)).To(Equal(message1))
		Expect(queue.Get(2)).To(Equal(message2))
	})

	It("should correctly prioritize suspect gossip when preparing", func() {
		message1 := &gossip.MessageSuspect{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &gossip.MessageAlive{
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
		message1 := &gossip.MessageFaulty{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &gossip.MessageAlive{
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
		message1 := &gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkTransmitted(0)
		queue.MarkTransmitted(0)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)
		queue.MarkTransmitted(1)

		message3 := &gossip.MessageAlive{
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
		gossipQueue := gossip.NewMessageQueue(3)
		gossipQueue.Add(&gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		})
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkTransmitted(0)
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkTransmitted(0)
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkTransmitted(0)
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(0))
	})
})

func BenchmarkGossipQueue_Add(b *testing.B) {
	for gossipCount := 1; gossipCount <= 16*1024; gossipCount *= 2 {
		gossipQueue := gossip.NewMessageQueue(math.MaxInt)
		for i := range gossipCount {
			gossipQueue.Add(&gossip.MessageAlive{
				Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			for b.Loop() {
				gossipQueue.Add(&gossip.MessageAlive{
					Source:            encoding.NewAddress(net.IPv4(11, 12, 13, 14), 1024),
					IncarnationNumber: 0,
				})
				//gossipQueue.UndoAdd()
			}
		})
	}
}

func BenchmarkGossipQueue_PrepareFor(b *testing.B) {
	for gossipCount := 1; gossipCount <= 16*1024; gossipCount *= 2 {
		gossipQueue := gossip.NewMessageQueue(math.MaxInt)
		for i := range gossipCount {
			gossipQueue.Add(&gossip.MessageAlive{
				Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			for b.Loop() {
				gossipQueue.PrioritizeForAddress(encoding.Address{})
			}
		})
	}
}
