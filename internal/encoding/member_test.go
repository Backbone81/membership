package encoding_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testMember = encoding.Member{
	Address:           encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
	State:             encoding.MemberStateAlive,
	IncarnationNumber: 1,
}

var _ = Describe("Member", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := encoding.AppendMemberToBuffer(nil, testMember)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := encoding.AppendMemberToBuffer(localBuffer[:0], testMember)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		buffer, appendN, err := encoding.AppendMemberToBuffer(nil, testMember)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readMember, readN, err := encoding.MemberFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(testMember).To(Equal(readMember))
	})

	It("should fail to read from nil buffer", func() {
		Expect(encoding.MemberFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := encoding.AppendMemberToBuffer(nil, testMember)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(encoding.MemberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendMemberToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := encoding.AppendMemberToBuffer(buffer[:0], testMember); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemberFromBuffer(b *testing.B) {
	buffer, _, err := encoding.AppendMemberToBuffer(nil, testMember)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := encoding.MemberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
