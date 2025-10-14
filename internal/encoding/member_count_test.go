package encoding_test

import (
	"math"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemberCount", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := encoding.AppendMemberCountToBuffer(nil, 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := encoding.AppendMemberCountToBuffer(localBuffer[:0], 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	DescribeTable("should append to buffer with valid member counts",
		func(memberCount int) {
			buffer, _, err := encoding.AppendMemberCountToBuffer(nil, memberCount)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		},
		Entry("zero", 0),
		Entry("small positive", 80),
		Entry("big positive", 3000),
	)

	DescribeTable("should fail to append to buffer with invalid member count",
		func(memberCount int) {
			Expect(encoding.AppendMemberCountToBuffer(nil, memberCount)).Error().To(HaveOccurred())
		},
		Entry("negative", -10),
		Entry("too big of an incarnation number", math.MaxUint32+1),
	)

	It("should read from buffer", func() {
		appendMemberCount := 1024
		buffer, appendN, err := encoding.AppendMemberCountToBuffer(nil, appendMemberCount)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readMemberCount, readN, err := encoding.MemberCountFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendMemberCount).To(Equal(readMemberCount))
	})

	It("should fail to read from nil buffer", func() {
		Expect(encoding.MemberCountFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := encoding.AppendMemberCountToBuffer(nil, 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(encoding.MemberCountFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendMemberCountToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := encoding.AppendMemberCountToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemberCountFromBuffer(b *testing.B) {
	buffer, _, err := encoding.AppendMemberCountToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := encoding.MemberCountFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
