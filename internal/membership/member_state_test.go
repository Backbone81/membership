package membership_test

import (
	"testing"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemberState", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := membership.AppendMemberStateToBuffer(nil, membership.MemberStateAlive)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := membership.AppendMemberStateToBuffer(localBuffer[:0], membership.MemberStateAlive)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		appendMemberState := membership.MemberStateAlive
		buffer, appendN, err := membership.AppendMemberStateToBuffer(nil, appendMemberState)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readMemberState, readN, err := membership.MemberStateFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendMemberState).To(Equal(readMemberState))
	})

	It("should fail to read from nil buffer", func() {
		Expect(membership.MemberStateFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := membership.AppendMemberStateToBuffer(nil, membership.MemberStateAlive)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(membership.MemberStateFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendMemberStateToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendMemberStateToBuffer(buffer[:0], membership.MemberStateAlive); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemberStateFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendMemberStateToBuffer(nil, membership.MemberStateAlive)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.MemberStateFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
