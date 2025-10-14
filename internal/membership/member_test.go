package membership_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Member", func() {
	It("should append to nil buffer", func() {
		member := membership.Member{
			Address:           encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
			State:             membership.MemberStateAlive,
			IncarnationNumber: 1,
		}
		buffer, _, err := membership.AppendMemberToBuffer(nil, member)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		member := membership.Member{
			Address:           encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
			State:             membership.MemberStateAlive,
			IncarnationNumber: 1,
		}
		var localBuffer [10]byte
		buffer, _, err := membership.AppendMemberToBuffer(localBuffer[:0], member)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		appendMember := membership.Member{
			Address:           encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
			State:             membership.MemberStateAlive,
			IncarnationNumber: 1,
		}
		buffer, appendN, err := membership.AppendMemberToBuffer(nil, appendMember)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readMember, readN, err := membership.MemberFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendMember).To(Equal(readMember))
	})

	It("should fail to read from nil buffer", func() {
		Expect(membership.MemberFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		member := membership.Member{
			Address:           encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
			State:             membership.MemberStateAlive,
			IncarnationNumber: 1,
		}
		buffer, _, err := membership.AppendMemberToBuffer(nil, member)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(membership.MemberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendMemberToBuffer(b *testing.B) {
	member := membership.Member{
		Address:           encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
		State:             membership.MemberStateAlive,
		IncarnationNumber: 1,
	}
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendMemberToBuffer(buffer[:0], member); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemberFromBuffer(b *testing.B) {
	member := membership.Member{
		Address:           encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
		State:             membership.MemberStateAlive,
		IncarnationNumber: 1,
	}
	buffer, _, err := membership.AppendMemberToBuffer(nil, member)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.MemberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
