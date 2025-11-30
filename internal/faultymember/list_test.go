package faultymember_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/backbone81/membership/internal/faultymember"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
)

var _ = Describe("List", func() {
	Context("NewList", func() {
		It("should create faulty member list with default configuration", func() {
			list := faultymember.NewList()

			Expect(list).NotTo(BeNil())
			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount))
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(list.ValidateInternalState()).To(Succeed())

			config := list.Config()
			Expect(config.MaxListRequestCount).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(config.PreAllocationCount).To(Equal(faultymember.DefaultConfig.PreAllocationCount))
		})

		It("should create faulty member list with custom max list request count", func() {
			list := faultymember.NewList(faultymember.WithMaxListRequestCount(faultymember.DefaultConfig.MaxListRequestCount + 10))

			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount + 10))
			Expect(list.ValidateInternalState()).To(Succeed())

			config := list.Config()
			Expect(config.MaxListRequestCount).To(Equal(faultymember.DefaultConfig.MaxListRequestCount + 10))
			Expect(config.PreAllocationCount).To(Equal(faultymember.DefaultConfig.PreAllocationCount))
		})

		It("should create faulty member list with custom pre-allocation count", func() {
			list := faultymember.NewList(faultymember.WithPreAllocationCount(faultymember.DefaultConfig.PreAllocationCount + 1024))

			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount + 1024))
			Expect(list.ValidateInternalState()).To(Succeed())

			config := list.Config()
			Expect(config.MaxListRequestCount).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(config.PreAllocationCount).To(Equal(faultymember.DefaultConfig.PreAllocationCount + 1024))
		})

		It("should create faulty member list with multiple options", func() {
			list := faultymember.NewList(
				faultymember.WithMaxListRequestCount(faultymember.DefaultConfig.MaxListRequestCount+10),
				faultymember.WithPreAllocationCount(faultymember.DefaultConfig.PreAllocationCount+1024),
			)

			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount + 1024))
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount + 10))
			Expect(list.ValidateInternalState()).To(Succeed())

			config := list.Config()
			Expect(config.MaxListRequestCount).To(Equal(faultymember.DefaultConfig.MaxListRequestCount + 10))
			Expect(config.PreAllocationCount).To(Equal(faultymember.DefaultConfig.PreAllocationCount + 1024))
		})

		It("should enforce minimum max list request count", func() {
			list := faultymember.NewList(faultymember.WithMaxListRequestCount(0))

			Expect(list.Buckets()).To(Equal(1))
			Expect(list.ValidateInternalState()).To(Succeed())

			config := list.Config()
			Expect(config.MaxListRequestCount).To(Equal(1))
		})

		It("should enforce minimum pre-allocation count", func() {
			list := faultymember.NewList(faultymember.WithPreAllocationCount(0))

			Expect(list.Cap()).To(Equal(1))
			Expect(list.ValidateInternalState()).To(Succeed())

			config := list.Config()
			Expect(config.PreAllocationCount).To(Equal(1))
		})
	})

	Context("Clear", func() {
		It("should clear faulty member list with single member", func() {
			list := faultymember.NewList()
			list.Add(encoding.Member{
				Address: TestAddress,
			})

			Expect(list.Len()).To(Equal(1))
			Expect(list.IsEmpty()).To(BeFalse())
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount))

			list.Clear()

			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount))
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should clear faulty member list with multiple members", func() {
			list := faultymember.NewList()

			for i := 0; i < 10; i++ {
				list.Add(encoding.Member{
					Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
				})
			}

			Expect(list.Len()).To(Equal(10))
			Expect(list.IsEmpty()).To(BeFalse())
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount))

			list.Clear()

			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount))
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should clear members distributed across buckets", func() {
			list := faultymember.NewList()

			for i := 0; i < 10; i++ {
				list.Add(encoding.Member{
					Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
				})
			}

			list.ListRequestObserved()
			list.ListRequestObserved()
			list.ListRequestObserved()

			Expect(list.Len()).To(Equal(10))
			Expect(list.IsEmpty()).To(BeFalse())
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount))

			list.Clear()

			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			Expect(list.Buckets()).To(Equal(faultymember.DefaultConfig.MaxListRequestCount))
			Expect(list.Cap()).To(Equal(faultymember.DefaultConfig.PreAllocationCount))
			Expect(list.ValidateInternalState()).To(Succeed())
		})
	})

	Context("Add", func() {
		It("should add member to empty faulty member list", func() {
			list := faultymember.NewList()
			member := encoding.Member{
				Address: TestAddress,
			}
			list.Add(member)

			Expect(list.Len()).To(Equal(1))
			Expect(list.IsEmpty()).To(BeFalse())
			Expect(GetFromListByIndex(list, 0)).To(Equal(member))
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should add multiple members with different addresses", func() {
			list := faultymember.NewList()
			member1 := encoding.Member{
				Address: TestAddress,
			}
			member2 := encoding.Member{
				Address: TestAddress2,
			}
			list.Add(member1)
			list.Add(member2)

			Expect(list.Len()).To(Equal(2))
			Expect(list.IsEmpty()).To(BeFalse())
			Expect(GetFromListByIndex(list, 0)).To(Equal(member1))
			Expect(GetFromListByIndex(list, 1)).To(Equal(member2))
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should not add duplicate members for same address", func() {
			list := faultymember.NewList()
			member1 := encoding.Member{
				Address: TestAddress,
			}
			member2 := encoding.Member{
				Address: TestAddress,
			}
			list.Add(member1)
			list.Add(member2)

			Expect(list.Len()).To(Equal(1))
			Expect(list.IsEmpty()).To(BeFalse())
			Expect(GetFromListByIndex(list, 0)).To(Equal(member1))
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should grow ring buffer when capacity is reached", func() {
			list := faultymember.NewList(faultymember.WithPreAllocationCount(4))
			for i := 0; i < 3; i++ {
				list.Add(encoding.Member{
					Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
				})
			}
			Expect(list.Cap()).To(Equal(4))
			Expect(list.Len()).To(Equal(3))

			list.Add(encoding.Member{
				Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+3),
			})
			Expect(list.Cap()).To(Equal(8))
			Expect(list.Len()).To(Equal(4))

			for i := 0; i < 4; i++ {
				Expect(GetFromListByIndex(list, i).Address.Port()).To(Equal(1024 + i))
			}
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should grow ring buffer when wrapped around", func() {
			list := faultymember.NewList(
				faultymember.WithPreAllocationCount(4),
				faultymember.WithMaxListRequestCount(3),
			)

			for i := 0; i < 3; i++ {
				list.Add(encoding.Member{
					Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
				})
			}
			// Remove them to create wraparound
			list.ListRequestObserved()
			list.ListRequestObserved()
			list.ListRequestObserved()
			Expect(list.Len()).To(Equal(0))

			// Now add members that will cause wraparound + growth
			for i := 3; i < 7; i++ {
				list.Add(encoding.Member{
					Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
				})
			}

			Expect(list.Cap()).To(Equal(8))
			Expect(list.Len()).To(Equal(4))

			// Verify all members accessible
			for i := 0; i < 4; i++ {
				Expect(GetFromListByIndex(list, i).Address.Port()).To(Equal(1024 + 3 + i))
			}
			Expect(list.ValidateInternalState()).To(Succeed())
		})
	})

	Context("Get", func() {
		It("should get existing member", func() {
			list := faultymember.NewList()
			member := encoding.Member{
				Address:           TestAddress,
				IncarnationNumber: 42,
			}
			list.Add(member)

			retrieved, found := list.Get(TestAddress)
			Expect(found).To(BeTrue())
			Expect(retrieved).To(Equal(member))
		})

		It("should not get non-existent member", func() {
			list := faultymember.NewList()

			_, found := list.Get(TestAddress)
			Expect(found).To(BeFalse())
		})

		It("should get updated member after overwrite", func() {
			list := faultymember.NewList()
			member1 := encoding.Member{
				Address:           TestAddress,
				IncarnationNumber: 1,
			}
			member2 := encoding.Member{
				Address:           TestAddress,
				IncarnationNumber: 2,
			}

			list.Add(member1)
			list.Add(member2)

			retrieved, found := list.Get(TestAddress)
			Expect(found).To(BeTrue())
			Expect(retrieved.IncarnationNumber).To(Equal(uint16(2)))
		})

		It("should not get member after removal", func() {
			list := faultymember.NewList()
			member := encoding.Member{Address: TestAddress}

			list.Add(member)
			list.Remove(TestAddress)

			_, found := list.Get(TestAddress)
			Expect(found).To(BeFalse())
		})
	})

	Context("Remove", func() {
		It("should remove member", func() {
			list := faultymember.NewList()
			member := encoding.Member{Address: TestAddress}
			list.Add(member)

			list.Remove(TestAddress)

			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			_, found := list.Get(TestAddress)
			Expect(found).To(BeFalse())
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should remove correct member when multiple members exist", func() {
			list := faultymember.NewList()
			member1 := encoding.Member{Address: TestAddress}
			member2 := encoding.Member{Address: TestAddress2}
			member3 := encoding.Member{
				Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3),
			}

			list.Add(member1)
			list.Add(member2)
			list.Add(member3)

			list.Remove(TestAddress2)

			Expect(list.Len()).To(Equal(2))
			_, found := list.Get(TestAddress)
			Expect(found).To(BeTrue())
			_, found = list.Get(TestAddress2)
			Expect(found).To(BeFalse())
			_, found = list.Get(member3.Address)
			Expect(found).To(BeTrue())
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should do nothing when removing non-existent member", func() {
			list := faultymember.NewList()
			member := encoding.Member{Address: TestAddress}
			list.Add(member)

			list.Remove(TestAddress2) // Remove different address

			Expect(list.Len()).To(Equal(1))
			_, found := list.Get(TestAddress)
			Expect(found).To(BeTrue())
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should handle remove on empty list", func() {
			list := faultymember.NewList()

			list.Remove(TestAddress) // Should not panic

			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			Expect(list.ValidateInternalState()).To(Succeed())
		})
	})

	Context("ForEach", func() {
		It("should only iterate first 50% of buckets", func() {
			list := faultymember.NewList(faultymember.WithMaxListRequestCount(4))

			// Add members and distribute across buckets
			member1 := encoding.Member{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)}
			member2 := encoding.Member{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)}
			member3 := encoding.Member{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3)}
			member4 := encoding.Member{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 4)}

			list.Add(member1)
			list.ListRequestObserved() // member1 -> bucket 1
			list.Add(member2)
			list.ListRequestObserved() // member1 -> bucket 2, member2 -> bucket 1
			list.Add(member3)
			list.ListRequestObserved() // member1 -> bucket 3, member2 -> bucket 2, member3 -> bucket 1
			list.Add(member4)          // member1 in bucket 3, member2 in bucket 2, member3 in bucket 1, member4 in bucket 0

			// ForEach should only return members in buckets 0 and 1 (first 50% of 4 buckets)
			var returned []encoding.Member
			list.ForEach(func(member encoding.Member) bool {
				returned = append(returned, member)
				return true
			})

			Expect(returned).To(HaveLen(2))
			// Should contain member3 and member4 (buckets 0 and 1)
			Expect(returned).To(ContainElement(member3))
			Expect(returned).To(ContainElement(member4))
			// Should NOT contain member1 and member2 (buckets 2 and 3, older than 50%)
			Expect(returned).NotTo(ContainElement(member1))
			Expect(returned).NotTo(ContainElement(member2))
		})

		It("should iterate from oldest to newest within propagation window", func() {
			list := faultymember.NewList(faultymember.WithMaxListRequestCount(4))

			member1 := encoding.Member{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)}
			member2 := encoding.Member{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)}

			list.Add(member1)
			list.ListRequestObserved() // member1 -> bucket 1
			list.Add(member2)          // member2 -> bucket 0

			var returned []encoding.Member
			list.ForEach(func(member encoding.Member) bool {
				returned = append(returned, member)
				return true
			})

			// Should iterate from bucket 1 to bucket 0 (oldest to newest within window)
			Expect(returned).To(HaveLen(2))
			Expect(returned[0]).To(Equal(member1)) // Older (bucket 1)
			Expect(returned[1]).To(Equal(member2)) // Newer (bucket 0)
		})

		It("should handle early abort", func() {
			list := faultymember.NewList()
			for i := 0; i < 10; i++ {
				list.Add(encoding.Member{
					Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
				})
			}

			count := 0
			list.ForEach(func(member encoding.Member) bool {
				count++
				return count < 3 // Abort after 3 iterations
			})

			Expect(count).To(Equal(3))
		})

		It("should handle empty list", func() {
			list := faultymember.NewList()

			count := 0
			list.ForEach(func(member encoding.Member) bool {
				count++
				return true
			})

			Expect(count).To(Equal(0))
		})
	})

	Context("ListRequestObserved", func() {
		It("should remove members that exceed max list request count", func() {
			list := faultymember.NewList(faultymember.WithMaxListRequestCount(3))

			member := encoding.Member{
				Address: TestAddress,
			}
			list.Add(member)

			// Move through all buckets: 0 -> 1 -> 2 -> removed
			list.ListRequestObserved()
			Expect(list.Len()).To(Equal(1))

			list.ListRequestObserved()
			Expect(list.Len()).To(Equal(1))

			list.ListRequestObserved()
			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			Expect(list.ValidateInternalState()).To(Succeed())
		})

		It("should handle empty faulty member list", func() {
			list := faultymember.NewList()

			list.ListRequestObserved()

			Expect(list.Len()).To(Equal(0))
			Expect(list.IsEmpty()).To(BeTrue())
			Expect(list.ValidateInternalState()).To(Succeed())
		})
	})

	It("should maintain valid internal state under random operations", func() {
		// This test is a kind of monte carlo test. Creating random inputs and validating the internal state to be
		// correct.
		list := faultymember.NewList()
		var addresses []encoding.Address
		for i := range 5 {
			addresses = append(addresses, encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i))
		}
		for range 100_000 {
			switch selection := rand.Intn(100); {
			case selection < 25: // 25% of the time we add a member
				list.Add(encoding.Member{
					Address: addresses[rand.Intn(len(addresses))],
				})
			case selection < 100: // 75% of the time we observe a list request
				list.ListRequestObserved()
			}
			Expect(list.ValidateInternalState()).To(Succeed())
		}
	})
})

// BenchmarkList_Add is measuring the time an addition of a new member needs depending on the number of
// members already there and the number of buckets the members are distributed over.
func BenchmarkList_Add(b *testing.B) {
	// We want to test for members up to 16k.
	for memberCount := 1024; memberCount <= 16*1024; memberCount *= 2 {
		// We want to test for bucket counts of up to 32.
		for bucketCount := 8; bucketCount <= 32; bucketCount *= 2 {
			// We fill a new faulty member list with members until member count is reached.
			list := faultymember.NewList(
				faultymember.WithMaxListRequestCount(bucketCount),
				faultymember.WithPreAllocationCount(10_000_000),
			)
			for i := range memberCount {
				list.Add(encoding.Member{
					// We differentiate every source by a different port number.
					Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
				})

				// move members to the next bucket.
				if i%memberCount/bucketCount == 0 {
					list.ListRequestObserved()
				}
			}
			b.Run(fmt.Sprintf("%d members in %d buckets", memberCount, bucketCount), func(b *testing.B) {
				// We need to prepare enough unique IP addresses to have real additions and get not cut short by
				// members which are already in the list. We differentiate the ip addresses by counting
				// the ip address up and keep a port which is different from what we put in before.
				addresses := make([]encoding.Address, b.N)
				for i := range b.N {
					ipBytes := encoding.Endian.AppendUint32(nil, uint32(i+1))
					addresses[i] = encoding.NewAddress(net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]), 512)
				}
				b.ResetTimer()
				for i := range b.N {
					list.Add(encoding.Member{
						Address: addresses[i],
					})
				}
			})
		}
	}
}

func BenchmarkList_ForEach(b *testing.B) {
	for memberCount := 1024; memberCount <= 16*1024; memberCount *= 2 {
		list := faultymember.NewList()
		for i := range memberCount {
			list.Add(encoding.Member{
				Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
			})
		}
		b.Run(fmt.Sprintf("%d members", memberCount), func(b *testing.B) {
			var count int
			for {
				list.ForEach(func(member encoding.Member) bool {
					if count == b.N {
						return false
					}
					count++
					return true
				})
				if count == b.N {
					return
				}
			}
		})
	}
}

func BenchmarkList_ListRequestObserved(b *testing.B) {
	for memberCount := 1024; memberCount <= 16*1024; memberCount *= 2 {
		list := faultymember.NewList()
		for i := range memberCount {
			list.Add(encoding.Member{
				Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1024+i),
			})
		}
		b.Run(fmt.Sprintf("%d members", memberCount), func(b *testing.B) {
			for b.Loop() {
				list.ListRequestObserved()
			}
		})
	}
}
