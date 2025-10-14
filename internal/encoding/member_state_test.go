package encoding_test

import (
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemberState", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := encoding.AppendMemberStateToBuffer(nil, encoding.MemberStateAlive)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := encoding.AppendMemberStateToBuffer(localBuffer[:0], encoding.MemberStateAlive)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		appendMemberState := encoding.MemberStateAlive
		buffer, appendN, err := encoding.AppendMemberStateToBuffer(nil, appendMemberState)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readMemberState, readN, err := encoding.MemberStateFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendMemberState).To(Equal(readMemberState))
	})

	It("should fail to read from nil buffer", func() {
		Expect(encoding.MemberStateFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := encoding.AppendMemberStateToBuffer(nil, encoding.MemberStateAlive)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(encoding.MemberStateFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendMemberStateToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := encoding.AppendMemberStateToBuffer(buffer[:0], encoding.MemberStateAlive); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemberStateFromBuffer(b *testing.B) {
	buffer, _, err := encoding.AppendMemberStateToBuffer(nil, encoding.MemberStateAlive)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := encoding.MemberStateFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
