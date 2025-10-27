package gossip_test

import (
	"fmt"
	"math"
	"math/rand"
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
		Expect(queue.ValidateInternalState()).To(Succeed())
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
		Expect(queue.ValidateInternalState()).To(Succeed())
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
		Expect(queue.ValidateInternalState()).To(Succeed())
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
			Expect(queue.ValidateInternalState()).To(Succeed())
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
		queue.MarkFirstNMessagesTransmitted(1)
		queue.MarkFirstNMessagesTransmitted(1)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)

		message3 := &gossip.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkFirstNMessagesTransmitted(3)

		queue.PrioritizeForAddress(encoding.Address{})
		Expect(queue.Get(0)).To(Equal(message3))
		Expect(queue.Get(1)).To(Equal(message2))
		Expect(queue.Get(2)).To(Equal(message1))
		Expect(queue.ValidateInternalState()).To(Succeed())
	})

	It("should correctly prioritize suspect gossip when preparing", func() {
		message1 := &gossip.MessageSuspect{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkFirstNMessagesTransmitted(1)
		queue.MarkFirstNMessagesTransmitted(1)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)

		message3 := &gossip.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkFirstNMessagesTransmitted(3)

		queue.PrioritizeForAddress(TestAddress)
		Expect(queue.Get(0)).To(Equal(message1))
		Expect(queue.Get(1)).To(Equal(message3))
		Expect(queue.Get(2)).To(Equal(message2))
		Expect(queue.ValidateInternalState()).To(Succeed())
	})

	It("should correctly prioritize faulty gossip when preparing", func() {
		message1 := &gossip.MessageFaulty{
			Source:            TestAddress2,
			Destination:       TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkFirstNMessagesTransmitted(1)
		queue.MarkFirstNMessagesTransmitted(1)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)

		message3 := &gossip.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkFirstNMessagesTransmitted(3)

		queue.PrioritizeForAddress(TestAddress)
		Expect(queue.Get(0)).To(Equal(message1))
		Expect(queue.Get(1)).To(Equal(message3))
		Expect(queue.Get(2)).To(Equal(message2))
		Expect(queue.ValidateInternalState()).To(Succeed())
	})

	It("should correctly prioritize alive gossip when preparing", func() {
		message1 := &gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		}
		queue.Add(message1)
		queue.MarkFirstNMessagesTransmitted(1)
		queue.MarkFirstNMessagesTransmitted(1)

		message2 := &gossip.MessageAlive{
			Source:            TestAddress2,
			IncarnationNumber: 0,
		}
		queue.Add(message2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)
		queue.MarkFirstNMessagesTransmitted(2)

		message3 := &gossip.MessageAlive{
			Source:            TestAddress3,
			IncarnationNumber: 0,
		}
		queue.Add(message3)
		queue.MarkFirstNMessagesTransmitted(3)

		queue.PrioritizeForAddress(TestAddress)
		Expect(queue.Get(0)).To(Equal(message3))
		Expect(queue.Get(1)).To(Equal(message2))
		Expect(queue.Get(2)).To(Equal(message1))
		Expect(queue.ValidateInternalState()).To(Succeed())
	})

	It("should correctly remove messages which were transmitted enough", func() {
		gossipQueue := gossip.NewMessageQueue(3)
		gossipQueue.Add(&gossip.MessageAlive{
			Source:            TestAddress,
			IncarnationNumber: 0,
		})
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkFirstNMessagesTransmitted(1)
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkFirstNMessagesTransmitted(1)
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(1))

		gossipQueue.MarkFirstNMessagesTransmitted(1)
		gossipQueue.PrioritizeForAddress(encoding.Address{})
		Expect(gossipQueue.Len()).To(Equal(0))
		Expect(queue.ValidateInternalState()).To(Succeed())
	})

	It("internal state should always be valid", func() {
		// This test is a kind of monte carlo test. Creating random inputs and validating the internal state to be
		// correct.
		queue = gossip.NewMessageQueue(10)
		var addresses []encoding.Address
		for i := range 5 {
			addresses = append(addresses, encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i))
		}
		for range 1_000_000 {
			switch selection := rand.Intn(100); {
			case selection < 25: // 25% of the time we add a message
				queue.Add(&gossip.MessageAlive{
					Source:            addresses[rand.Intn(len(addresses))],
					IncarnationNumber: rand.Intn(5),
				})
			case selection < 100: // 75% of the time we mark 3 messages as transmitted
				queue.MarkFirstNMessagesTransmitted(rand.Intn(3))
			}
			Expect(queue.ValidateInternalState()).To(Succeed())
		}
	})
})

// BenchmarkMessageQueue2_Add is measuring the time an addition of a new gossip message needs depending on the number of
// gossip already there and the number of buckets the gossip is distributed over.
func BenchmarkMessageQueue_Add(b *testing.B) {
	// We want to test for gossip up to 16k. This could in theory happen with a cluster of 16k members and there is one
	// gossip for every member.
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		// We want to test for bucket counts of up to 32. With a cluster of 16k members and a security factor of 3, it
		// would require 29 transmissions of every gossip message before it could be dropped as safely gossiped. A limit
		// of 32 is adding some additional buffer to stay in powers of two.
		for bucketCount := 8; bucketCount <= 32; bucketCount *= 2 {
			// We fill a new gossip queue with gossip messages until gossip count is reached.
			gossipQueue := gossip.NewMessageQueue(math.MaxInt)
			for i := range gossipCount {
				gossipQueue.Add(&gossip.MessageAlive{
					// We differentiate every source by a different port number.
					Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})

				// Mark all messages as transmitted once to move all messages to the next bucket.
				if i%gossipCount/bucketCount == 0 {
					gossipQueue.MarkFirstNMessagesTransmitted(gossipQueue.Len())
				}
			}
			b.Run(fmt.Sprintf("%d gossip in %d buckets", gossipCount, bucketCount), func(b *testing.B) {
				// We need to prepare enough unique IP addresses to have real additions and get not cut short by
				// messages for members which are already in the queue. We differentiate the ip addresses by counting
				// the ip address up and keep a port which is different from what we put in before.
				addresses := make([]encoding.Address, b.N)
				for i := range b.N {
					ipBytes := encoding.Endian.AppendUint32(nil, uint32(i+1))
					addresses[i] = encoding.NewAddress(net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]), 512)
				}
				b.ResetTimer()
				for i := range b.N {
					gossipQueue.Add(&gossip.MessageAlive{
						Source:            addresses[i],
						IncarnationNumber: 0,
					})
				}
			})
		}
	}
}

func BenchmarkMessageQueue_PrioritizeForAddress(b *testing.B) {
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		gossipQueue := gossip.NewMessageQueue(math.MaxInt)
		for i := range gossipCount {
			gossipQueue.Add(&gossip.MessageAlive{
				Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			// Make sure we are using an address which actually exists in the queue. That way the code is taking the
			// slower path and is not exiting early.
			address := encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+1)
			for b.Loop() {
				gossipQueue.PrioritizeForAddress(address)
			}
		})
	}
}

func BenchmarkMessageQueue_Get(b *testing.B) {
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		gossipQueue := gossip.NewMessageQueue(math.MaxInt)
		for i := range gossipCount {
			gossipQueue.Add(&gossip.MessageAlive{
				Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		gossipQueue.PrioritizeForAddress(encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+gossipCount/2))
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			for i := range b.N {
				gossipQueue.Get(i % gossipCount)
			}
		})
	}
}

func BenchmarkMessageQueue_MarkFirstNMessagesTransmitted(b *testing.B) {
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		gossipQueue := gossip.NewMessageQueue(math.MaxInt)
		for i := range gossipCount {
			gossipQueue.Add(&gossip.MessageAlive{
				Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		for messagesTransmitted := 1; messagesTransmitted <= 128; messagesTransmitted *= 2 {
			b.Run(fmt.Sprintf("%d gossip with %d transmissions", gossipCount, messagesTransmitted), func(b *testing.B) {
				for b.Loop() {
					gossipQueue.MarkFirstNMessagesTransmitted(messagesTransmitted)
				}
			})
		}
	}
}
