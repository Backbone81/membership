package gossip_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
)

var _ = Describe("Queue", func() {
	Context("NewQueue", func() {
		It("should create queue with default configuration", func() {
			queue := gossip.NewQueue()

			Expect(queue).NotTo(BeNil())
			Expect(queue.Len()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount))
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(queue.ValidateInternalState()).To(Succeed())

			config := queue.Config()
			Expect(config.MaxTransmissionCount).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(config.PreAllocationCount).To(Equal(gossip.DefaultConfig.PreAllocationCount))
		})

		It("should create queue with custom max transmission count", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(gossip.DefaultConfig.MaxTransmissionCount + 10))

			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount + 10))
			Expect(queue.ValidateInternalState()).To(Succeed())

			config := queue.Config()
			Expect(config.MaxTransmissionCount).To(Equal(gossip.DefaultConfig.MaxTransmissionCount + 10))
			Expect(config.PreAllocationCount).To(Equal(gossip.DefaultConfig.PreAllocationCount))
		})

		It("should create queue with custom pre-allocation count", func() {
			queue := gossip.NewQueue(gossip.WithPreAllocationCount(gossip.DefaultConfig.PreAllocationCount + 1024))

			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount + 1024))
			Expect(queue.ValidateInternalState()).To(Succeed())

			config := queue.Config()
			Expect(config.MaxTransmissionCount).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(config.PreAllocationCount).To(Equal(gossip.DefaultConfig.PreAllocationCount + 1024))
		})

		It("should create queue with multiple options", func() {
			queue := gossip.NewQueue(
				gossip.WithMaxTransmissionCount(gossip.DefaultConfig.MaxTransmissionCount+10),
				gossip.WithPreAllocationCount(gossip.DefaultConfig.PreAllocationCount+1024),
			)

			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount + 1024))
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount + 10))
			Expect(queue.ValidateInternalState()).To(Succeed())

			config := queue.Config()
			Expect(config.MaxTransmissionCount).To(Equal(gossip.DefaultConfig.MaxTransmissionCount + 10))
			Expect(config.PreAllocationCount).To(Equal(gossip.DefaultConfig.PreAllocationCount + 1024))
		})

		It("should enforce minimum max transmission count", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(0))

			Expect(queue.Buckets()).To(Equal(1))
			Expect(queue.ValidateInternalState()).To(Succeed())

			config := queue.Config()
			Expect(config.MaxTransmissionCount).To(Equal(1))
		})

		It("should enforce minimum pre-allocation count", func() {
			queue := gossip.NewQueue(gossip.WithPreAllocationCount(0))

			Expect(queue.Cap()).To(Equal(1))
			Expect(queue.ValidateInternalState()).To(Succeed())

			config := queue.Config()
			Expect(config.PreAllocationCount).To(Equal(1))
		})
	})

	Context("Clear", func() {
		It("should clear queue with single message", func() {
			queue := gossip.NewQueue()
			queue.Add(&encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			})

			Expect(queue.Len()).To(Equal(1))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount))

			queue.Clear()

			Expect(queue.Len()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should clear queue with multiple messages", func() {
			queue := gossip.NewQueue()

			for i := 0; i < 10; i++ {
				queue.Add(&encoding.MessageAlive{
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})
			}

			Expect(queue.Len()).To(Equal(10))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount))

			queue.Clear()

			Expect(queue.Len()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should clear messages distributed across buckets", func() {
			queue := gossip.NewQueue()

			for i := 0; i < 10; i++ {
				queue.Add(&encoding.MessageAlive{
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})
			}

			queue.MarkTransmitted(3)
			queue.MarkTransmitted(3)
			queue.MarkTransmitted(2)

			Expect(queue.Len()).To(Equal(10))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount))

			queue.Clear()

			Expect(queue.Len()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
			Expect(queue.Buckets()).To(Equal(gossip.DefaultConfig.MaxTransmissionCount))
			Expect(queue.Cap()).To(Equal(gossip.DefaultConfig.PreAllocationCount))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})
	})

	Context("SetMaxTransmissionCount", func() {
		It("should increase bucket count", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(3))

			Expect(queue.Buckets()).To(Equal(3))

			queue.SetMaxTransmissionCount(10)

			Expect(queue.Buckets()).To(Equal(10))
			Expect(queue.Config().MaxTransmissionCount).To(Equal(10))
			Expect(queue.Len()).To(Equal(0))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should decrease bucket count", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(10))

			Expect(queue.Buckets()).To(Equal(10))

			queue.SetMaxTransmissionCount(3)

			Expect(queue.Buckets()).To(Equal(3))
			Expect(queue.Config().MaxTransmissionCount).To(Equal(3))
			Expect(queue.Len()).To(Equal(0))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should remove messages which are falling out of the active buckets", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(4))

			for i := 0; i < 10; i++ {
				queue.Add(&encoding.MessageAlive{
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})
			}

			queue.MarkTransmitted(10) // move all to bucket 1
			queue.MarkTransmitted(10) // move all to bucket 2
			Expect(queue.Len()).To(Equal(10))

			queue.SetMaxTransmissionCount(3)

			Expect(queue.Buckets()).To(Equal(3))
			Expect(queue.Config().MaxTransmissionCount).To(Equal(3))
			Expect(queue.Len()).To(Equal(10))
			Expect(queue.ValidateInternalState()).To(Succeed())

			queue.SetMaxTransmissionCount(2)

			Expect(queue.Buckets()).To(Equal(2))
			Expect(queue.Config().MaxTransmissionCount).To(Equal(2))
			Expect(queue.Len()).To(Equal(0))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should enforce minimum value of 1", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(5))
			Expect(queue.Buckets()).To(Equal(5))
			Expect(queue.Config().MaxTransmissionCount).To(Equal(5))

			queue.SetMaxTransmissionCount(0)

			Expect(queue.Buckets()).To(Equal(1))
			Expect(queue.Config().MaxTransmissionCount).To(Equal(1))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})
	})

	Context("Add", func() {
		It("should add message to empty queue", func() {
			queue := gossip.NewQueue()
			message := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(message)

			Expect(queue.Len()).To(Equal(1))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.Get(0)).To(Equal(message))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should add multiple messages with different addresses", func() {
			queue := gossip.NewQueue()
			message1 := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			message2 := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(message1)
			queue.Add(message2)

			Expect(queue.Len()).To(Equal(2))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.Get(0)).To(Equal(message1))
			Expect(queue.Get(1)).To(Equal(message2))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should not add duplicate message for same address", func() {
			queue := gossip.NewQueue()
			message1 := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			message2 := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(message1)
			queue.Add(message2)

			Expect(queue.Len()).To(Equal(1))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.Get(0)).To(Equal(message1))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should grow ring buffer when capacity is reached", func() {
			queue := gossip.NewQueue(gossip.WithPreAllocationCount(4))
			for i := 0; i < 3; i++ {
				queue.Add(&encoding.MessageAlive{
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})
			}
			Expect(queue.Cap()).To(Equal(4))
			Expect(queue.Len()).To(Equal(3))

			queue.Add(&encoding.MessageAlive{
				Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+3),
				IncarnationNumber: 0,
			})
			Expect(queue.Cap()).To(Equal(8))
			Expect(queue.Len()).To(Equal(4))

			for i := 0; i < 4; i++ {
				Expect(queue.Get(i).GetAddress().Port()).To(Equal(1024 + i))
			}
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should grow ring buffer when wrapped around", func() {
			queue := gossip.NewQueue(
				gossip.WithPreAllocationCount(4),
				gossip.WithMaxTransmissionCount(3),
			)

			for i := 0; i < 3; i++ {
				queue.Add(&encoding.MessageAlive{
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})
			}
			// Remove them to create wraparound
			queue.MarkTransmitted(3)
			queue.MarkTransmitted(3)
			queue.MarkTransmitted(3)
			Expect(queue.Len()).To(Equal(0))

			// Now add messages that will cause wraparound + growth
			for i := 3; i < 7; i++ {
				queue.Add(&encoding.MessageAlive{
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})
			}

			Expect(queue.Cap()).To(Equal(8))
			Expect(queue.Len()).To(Equal(4))

			// Verify all messages accessible
			for i := 0; i < 4; i++ {
				Expect(queue.Get(i).GetAddress().Port()).To(Equal(1024 + 3 + i))
			}
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should move overwritten message back to bucket 0", func() {
			queue := gossip.NewQueue()
			message1 := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 1,
			}
			queue.Add(message1)
			queue.MarkTransmitted(1)
			queue.MarkTransmitted(1)

			message2 := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(message2)
			queue.MarkTransmitted(1)

			Expect(queue.Get(0)).To(Equal(message2))
			Expect(queue.Get(1)).To(Equal(message1))

			message1Updated := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 2,
			}
			queue.Add(message1Updated)

			Expect(queue.Len()).To(Equal(2))
			Expect(queue.Get(0)).To(Equal(message1Updated))
			Expect(queue.Get(1)).To(Equal(message2))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		DescribeTable("Messages should overwrite in the correct priority",
			func(message1 gossip.Message, message2 gossip.Message, overwrite bool) {
				queue := gossip.NewQueue()
				queue.Add(message1)
				queue.Add(message2)
				if overwrite {
					Expect(queue.Get(0)).To(Equal(message2))
				} else {
					Expect(queue.Get(0)).To(Equal(message1))
				}
				Expect(queue.ValidateInternalState()).To(Succeed())
			},
			Entry("Alive with lower incarnation number should NOT overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Alive with same incarnation number should NOT overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				false,
			),
			Entry("Alive with bigger incarnation number should overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),
			Entry("Suspect with lower incarnation number should NOT overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Suspect with same incarnation number should overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				true,
			),
			Entry("Suspect with bigger incarnation number should overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),
			Entry("Faulty with lower incarnation number should NOT overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Faulty with same incarnation number should overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				true,
			),
			Entry("Faulty with bigger incarnation number should overwrite alive",
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),

			Entry("Alive with lower incarnation number should NOT overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Alive with same incarnation number should NOT overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				false,
			),
			Entry("Alive with bigger incarnation number should overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),
			Entry("Suspect with lower incarnation number should NOT overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Suspect with same incarnation number should NOT overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				false,
			),
			Entry("Suspect with bigger incarnation number should overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),
			Entry("Faulty with lower incarnation number should NOT overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Faulty with same incarnation number should overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				true,
			),
			Entry("Faulty with bigger incarnation number should overwrite suspect",
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),

			Entry("Alive with lower incarnation number should NOT overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Alive with same incarnation number should NOT overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				false,
			),
			Entry("Alive with bigger incarnation number should overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageAlive{
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),
			Entry("Suspect with lower incarnation number should NOT overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Suspect with same incarnation number should NOT overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				false,
			),
			Entry("Suspect with bigger incarnation number should overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageSuspect{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),
			Entry("Faulty with lower incarnation number should NOT overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Faulty with same incarnation number should NOT overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 1,
				},
				false,
			),
			Entry("Faulty with bigger incarnation number should overwrite faulty",
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 2,
				},
				&encoding.MessageFaulty{
					Source:            TestAddress2,
					Destination:       TestAddress,
					IncarnationNumber: 3,
				},
				true,
			),
		)
	})

	Context("Prioritize", func() {
		It("should not prioritize when address does not exist", func() {
			queue := gossip.NewQueue()
			message1 := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			message2 := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(message1)
			queue.Add(message2)

			queue.Prioritize(TestAddress3)

			Expect(queue.Get(0)).To(Equal(message1))
			Expect(queue.Get(1)).To(Equal(message2))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should prioritize suspect message", func() {
			queue := gossip.NewQueue()

			alive1 := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(alive1)

			suspect := &encoding.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(suspect)

			alive2 := &encoding.MessageAlive{
				Destination:       TestAddress3,
				IncarnationNumber: 0,
			}
			queue.Add(alive2)

			queue.Prioritize(encoding.Address{})
			Expect(queue.Get(0)).To(Equal(alive1))
			Expect(queue.Get(1)).To(Equal(suspect))
			Expect(queue.Get(2)).To(Equal(alive2))

			queue.Prioritize(TestAddress)
			Expect(queue.Get(0)).To(Equal(suspect))
			Expect(queue.Get(1)).To(Equal(alive1))
			Expect(queue.Get(2)).To(Equal(alive2))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should prioritize faulty message", func() {
			queue := gossip.NewQueue()

			alive1 := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(alive1)

			faulty := &encoding.MessageFaulty{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(faulty)

			alive2 := &encoding.MessageAlive{
				Destination:       TestAddress3,
				IncarnationNumber: 0,
			}
			queue.Add(alive2)

			queue.Prioritize(encoding.Address{})
			Expect(queue.Get(0)).To(Equal(alive1))
			Expect(queue.Get(1)).To(Equal(faulty))
			Expect(queue.Get(2)).To(Equal(alive2))

			queue.Prioritize(TestAddress)
			Expect(queue.Get(0)).To(Equal(faulty))
			Expect(queue.Get(1)).To(Equal(alive1))
			Expect(queue.Get(2)).To(Equal(alive2))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should not prioritize alive message", func() {
			queue := gossip.NewQueue()

			alive1 := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(alive1)

			alive2 := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(alive2)

			alive3 := &encoding.MessageAlive{
				Destination:       TestAddress3,
				IncarnationNumber: 0,
			}
			queue.Add(alive3)

			queue.Prioritize(encoding.Address{})
			Expect(queue.Get(0)).To(Equal(alive1))
			Expect(queue.Get(1)).To(Equal(alive2))
			Expect(queue.Get(2)).To(Equal(alive3))

			queue.Prioritize(TestAddress)
			Expect(queue.Get(0)).To(Equal(alive1))
			Expect(queue.Get(1)).To(Equal(alive2))
			Expect(queue.Get(2)).To(Equal(alive3))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should clear priority when prioritized message is removed", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(3))

			suspect := &encoding.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(suspect)
			queue.MarkTransmitted(1)

			alive := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(alive)

			Expect(queue.Get(0)).To(Equal(alive))
			Expect(queue.Get(1)).To(Equal(suspect))

			// Prioritize the suspect message
			queue.Prioritize(TestAddress)
			Expect(queue.Get(0)).To(Equal(suspect))
			Expect(queue.Get(1)).To(Equal(alive))

			queue.SetMaxTransmissionCount(1)

			Expect(queue.Len()).To(Equal(1))
			Expect(queue.Get(0)).To(Equal(alive))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should preserve priority index when ring buffer grows", func() {
			queue := gossip.NewQueue(gossip.WithPreAllocationCount(4))

			suspect := &encoding.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(suspect)
			queue.MarkTransmitted(1)

			queue.Add(&encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			})
			queue.Add(&encoding.MessageAlive{
				Destination:       TestAddress3,
				IncarnationNumber: 0,
			})

			queue.Prioritize(TestAddress)
			Expect(queue.Get(0)).To(Equal(suspect))

			Expect(queue.Cap()).To(Equal(4))
			Expect(queue.Len()).To(Equal(3))
			alive := &encoding.MessageAlive{
				Destination:       encoding.NewAddress(net.IPv4(5, 6, 7, 8), 9999),
				IncarnationNumber: 0,
			}
			queue.Add(alive)

			Expect(queue.Cap()).To(Equal(8))
			Expect(queue.Len()).To(Equal(4))

			Expect(queue.Get(0)).To(Equal(suspect))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should update priority index when prioritized message is swapped", func() {
			queue := gossip.NewQueue(gossip.WithPreAllocationCount(4))

			suspect := &encoding.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(suspect)
			queue.MarkTransmitted(1)

			alive := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(alive)
			queue.MarkTransmitted(2)

			queue.Prioritize(TestAddress)
			Expect(queue.Get(0)).To(Equal(suspect))
			Expect(queue.Get(1)).To(Equal(alive))

			suspectUpdated := &encoding.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 1, // higher incarnation number
			}
			queue.Add(suspectUpdated)

			Expect(queue.Get(0)).To(Equal(suspectUpdated))
			Expect(queue.Len()).To(Equal(2))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should update priority index when message is swapped with prioritized", func() {
			queue := gossip.NewQueue(gossip.WithPreAllocationCount(4))

			alive := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			queue.Add(alive)
			queue.MarkTransmitted(1)

			suspect := &encoding.MessageSuspect{
				Source:            TestAddress2,
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(suspect)
			queue.MarkTransmitted(2)

			queue.Prioritize(TestAddress)
			Expect(queue.Get(0)).To(Equal(suspect))
			Expect(queue.Get(1)).To(Equal(alive))

			aliveUpdated := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 1, // higher incarnation number
			}
			queue.Add(aliveUpdated)

			Expect(queue.Get(0)).To(Equal(suspect))
			Expect(queue.Len()).To(Equal(2))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})
	})

	Context("MarkTransmitted", func() {
		It("should move requested number of messages", func() {
			queue := gossip.NewQueue()
			message1 := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			message2 := &encoding.MessageAlive{
				Destination:       TestAddress2,
				IncarnationNumber: 0,
			}
			message3 := &encoding.MessageAlive{
				Destination:       TestAddress3,
				IncarnationNumber: 0,
			}
			queue.Add(message1)
			queue.Add(message2)
			queue.Add(message3)

			Expect(queue.Get(0)).To(Equal(message1))
			Expect(queue.Get(1)).To(Equal(message2))
			Expect(queue.Get(2)).To(Equal(message3))
			Expect(queue.Len()).To(Equal(3))

			queue.MarkTransmitted(2)

			Expect(queue.Get(0)).To(Equal(message3))
			Expect(queue.Get(1)).To(Equal(message1))
			Expect(queue.Get(2)).To(Equal(message2))
			Expect(queue.Len()).To(Equal(3))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should remove messages that exceed max transmission count", func() {
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(3))

			message := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(message)

			// Move through all buckets: 0 -> 1 -> 2 -> removed
			queue.MarkTransmitted(1)
			Expect(queue.Len()).To(Equal(1))

			queue.MarkTransmitted(1)
			Expect(queue.Len()).To(Equal(1))

			queue.MarkTransmitted(1)
			Expect(queue.Len()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should handle marking more messages than exist", func() {
			queue := gossip.NewQueue()
			message := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(message)

			queue.MarkTransmitted(10)

			Expect(queue.Len()).To(Equal(1))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should handle marking zero messages", func() {
			queue := gossip.NewQueue()
			message := &encoding.MessageAlive{
				Destination:       TestAddress,
				IncarnationNumber: 0,
			}
			queue.Add(message)

			queue.MarkTransmitted(0)

			Expect(queue.Len()).To(Equal(1))
			Expect(queue.IsEmpty()).To(BeFalse())
			Expect(queue.Get(0)).To(Equal(message))
			Expect(queue.ValidateInternalState()).To(Succeed())
		})

		It("should handle empty queue", func() {
			queue := gossip.NewQueue()

			queue.MarkTransmitted(10)

			Expect(queue.Len()).To(Equal(0))
			Expect(queue.IsEmpty()).To(BeTrue())
			Expect(queue.ValidateInternalState()).To(Succeed())
		})
	})

	Context("Get", func() {
		It("should panic when accessed out of range", func() {
			queue := gossip.NewQueue()
			Expect(func() {
				queue.Get(1024)
			}).To(Panic())
		})
	})

	It("should maintain valid internal state under random operations", func() {
		// This test is a kind of monte carlo test. Creating random inputs and validating the internal state to be
		// correct.
		queue := gossip.NewQueue()
		var addresses []encoding.Address
		for i := range 5 {
			addresses = append(addresses, encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i))
		}
		for range 100_000 {
			switch selection := rand.Intn(100); {
			case selection < 25: // 25% of the time we add a message
				queue.Add(&encoding.MessageAlive{
					Destination:       addresses[rand.Intn(len(addresses))],
					IncarnationNumber: uint16(rand.Intn(5)),
				})
			case selection < 100: // 75% of the time we mark 3 messages as transmitted
				queue.MarkTransmitted(rand.Intn(3))
			}
			Expect(queue.ValidateInternalState()).To(Succeed())
		}
	})
})

// BenchmarkQueue_Add is measuring the time an addition of a new gossip message needs depending on the number of
// gossip already there and the number of buckets the gossip is distributed over.
func BenchmarkQueue_Add(b *testing.B) {
	// We want to test for gossip up to 16k. This could in theory happen with a cluster of 16k members and there is one
	// gossip for every member.
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		// We want to test for bucket counts of up to 32. With a cluster of 16k members and a security factor of 3, it
		// would require 29 transmissions of every gossip message before it could be dropped as safely gossiped. A limit
		// of 32 is adding some additional buffer to stay in powers of two.
		for bucketCount := 8; bucketCount <= 32; bucketCount *= 2 {
			// We fill a new gossip queue with gossip messages until gossip count is reached.
			queue := gossip.NewQueue(gossip.WithMaxTransmissionCount(bucketCount))
			for i := range gossipCount {
				queue.Add(&encoding.MessageAlive{
					// We differentiate every source by a different port number.
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})

				// Mark all messages as transmitted once to move all messages to the next bucket.
				if i%gossipCount/bucketCount == 0 {
					queue.MarkTransmitted(queue.Len())
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
					queue.Add(&encoding.MessageAlive{
						Destination:       addresses[i],
						IncarnationNumber: 0,
					})
				}
			})
		}
	}
}

func BenchmarkQueue_Prioritize(b *testing.B) {
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		queue := gossip.NewQueue()
		for i := range gossipCount {
			queue.Add(&encoding.MessageAlive{
				Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			// Make sure we are using an address which actually exists in the queue. That way the code is taking the
			// slower path and is not exiting early.
			address := encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+1)
			for b.Loop() {
				queue.Prioritize(address)
			}
		})
	}
}

func BenchmarkQueue_All(b *testing.B) {
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		queue := gossip.NewQueue()
		for i := range gossipCount {
			queue.Add(&encoding.MessageAlive{
				Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
				IncarnationNumber: 0,
			})
		}
		queue.Prioritize(encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+gossipCount/2))
		b.Run(fmt.Sprintf("%d gossip", gossipCount), func(b *testing.B) {
			var count int
			for {
				for range queue.All() {
					count++
					if count == b.N {
						return
					}
				}
			}
		})
	}
}

func BenchmarkQueue_MarkTransmitted(b *testing.B) {
	for gossipCount := 1024; gossipCount <= 16*1024; gossipCount *= 2 {
		for messagesTransmitted := 1; messagesTransmitted <= 128; messagesTransmitted *= 2 {
			queue := gossip.NewQueue()
			for i := range gossipCount {
				queue.Add(&encoding.MessageAlive{
					Destination:       encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024+i),
					IncarnationNumber: 0,
				})
			}
			b.Run(fmt.Sprintf("%d gossip with %d transmissions", gossipCount, messagesTransmitted), func(b *testing.B) {
				for b.Loop() {
					queue.MarkTransmitted(messagesTransmitted)
				}
			})
		}
	}
}
