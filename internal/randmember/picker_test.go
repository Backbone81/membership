package randmember_test

import (
	"fmt"
	"math"
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/randmember"
)

var _ = Describe("Picker", func() {
	var picker *randmember.Picker

	BeforeEach(func() {
		picker = randmember.NewPicker()
	})

	Context("Pick", func() {
		It("should return nothing when members slice is empty", func() {
			var members []encoding.Member

			var called int
			picker.Pick(5, members, func(member encoding.Member) {
				called++
			})
			Expect(called).To(Equal(0))
		})

		It("should return nothing when count is zero", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
			}

			var called int
			picker.Pick(0, members, func(member encoding.Member) {
				called++
			})
			Expect(called).To(Equal(0))
		})

		It("should return single member when count is one", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3)},
			}

			var picked []encoding.Member
			picker.Pick(1, members, func(member encoding.Member) {
				picked = append(picked, member)
			})
			Expect(picked).To(HaveLen(1))
			Expect(members).To(ContainElement(picked[0]))
		})

		It("should return all members when count equals member count", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3)},
			}

			var picked []encoding.Member
			picker.Pick(3, members, func(member encoding.Member) {
				picked = append(picked, member)
			})
			Expect(picked).To(HaveLen(3))
			Expect(picked).To(ConsistOf(members))
		})

		It("should clamp count to member slice length", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
			}

			var picked []encoding.Member
			picker.Pick(10, members, func(member encoding.Member) {
				picked = append(picked, member)
			})

			Expect(picked).To(HaveLen(2))
			Expect(picked).To(ConsistOf(members))
		})
	})

	Context("PickWithout", func() {
		It("should return nothing when members slice is empty", func() {
			var members []encoding.Member
			exclude := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)

			var called int
			picker.PickWithout(5, members, exclude, func(member encoding.Member) {
				called++
			})
			Expect(called).To(Equal(0))
		})

		It("should return nothing when count is zero", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
			}
			exclude := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)

			var called int
			picker.PickWithout(0, members, exclude, func(member encoding.Member) {
				called++
			})
			Expect(called).To(Equal(0))
		})

		It("should return nothing when the only member is excluded", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
			}
			exclude := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)

			var called int
			picker.PickWithout(5, members, exclude, func(member encoding.Member) {
				called++
			})
			Expect(called).To(Equal(0))
		})

		It("should exclude the specified member", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3)},
			}
			exclude := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)

			var picked []encoding.Member
			picker.PickWithout(2, members, exclude, func(member encoding.Member) {
				picked = append(picked, member)
			})

			expected := []encoding.Member{members[0], members[2]}
			Expect(picked).To(ConsistOf(expected))
		})

		It("should clamp count to available non-excluded members", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
			}
			exclude := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)

			var picked []encoding.Member
			picker.PickWithout(10, members, exclude, func(member encoding.Member) {
				picked = append(picked, member)
			})

			Expect(picked).To(HaveLen(1))
			Expect(picked[0]).To(Equal(members[0]))
		})

		It("should work when excluded member is not in the slice", func() {
			members := []encoding.Member{
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 1)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 2)},
				{Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), 3)},
			}
			exclude := encoding.NewAddress(net.IPv4(255, 255, 255, 255), 99)

			var picked []encoding.Member
			picker.PickWithout(2, members, exclude, func(member encoding.Member) {
				picked = append(picked, member)
			})

			Expect(picked).To(HaveLen(2))
			Expect(members).To(ContainElement(picked[0]))
			Expect(members).To(ContainElement(picked[1]))
		})
	})
})

func BenchmarkPicker_Pick(b *testing.B) {
	members := make([]encoding.Member, 0, 100)
	for i := range 100 {
		members = append(members, encoding.Member{
			Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), i),
		})
	}

	picker := randmember.NewPicker()
	for pickCount := 1; pickCount <= 32; pickCount *= 2 {
		b.Run(fmt.Sprintf("%d picks", pickCount), func(b *testing.B) {
			for b.Loop() {
				picker.Pick(pickCount, members, func(member encoding.Member) {
					_ = member.Address
				})
			}
		})
	}
}

func BenchmarkPicker_PickWithout(b *testing.B) {
	members := make([]encoding.Member, 0, 100)
	for i := range 100 {
		members = append(members, encoding.Member{
			Address: encoding.NewAddress(net.IPv4(255, 255, 255, 255), i),
		})
	}
	exclude := encoding.NewAddress(net.IPv4(255, 255, 255, 255), math.MaxUint16)

	picker := randmember.NewPicker()
	for pickCount := 1; pickCount <= 32; pickCount *= 2 {
		b.Run(fmt.Sprintf("%d picks", pickCount), func(b *testing.B) {
			for b.Loop() {
				picker.PickWithout(pickCount, members, exclude, func(member encoding.Member) {
					_ = member.Address
				})
			}
		})
	}
}
