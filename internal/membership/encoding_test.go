package membership_test

import (
	"math"
	"testing"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encoding", func() {
	Context("SequenceNumber", func() {
		It("should append to nil buffer", func() {
			buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should append to buffer", func() {
			var localBuffer [10]byte
			buffer, _, err := membership.AppendSequenceNumberToBuffer(localBuffer[:0], 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		DescribeTable("should append to buffer with valid sequence numbers",
			func(sequenceNumber int) {
				buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, sequenceNumber)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer).ToNot(BeNil())
			},
			Entry("zero", 0),
			Entry("small positive", 80),
			Entry("big positive", 3000),
		)

		DescribeTable("should fail to append to buffer with invalid sequence numbers",
			func(sequenceNumber int) {
				Expect(membership.AppendSequenceNumberToBuffer(nil, sequenceNumber)).Error().To(HaveOccurred())
			},
			Entry("negative", -10),
			Entry("too big of a sequence number", math.MaxUint16+1),
		)

		It("should read from buffer", func() {
			appendSequenceNumber := 1024
			buffer, appendN, err := membership.AppendSequenceNumberToBuffer(nil, appendSequenceNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			readSequenceNumber, readN, err := membership.SequenceNumberFromBuffer(buffer)
			Expect(err).ToNot(HaveOccurred())

			Expect(appendN).To(Equal(readN))
			Expect(appendSequenceNumber).To(Equal(readSequenceNumber))
		})

		It("should fail to read from nil buffer", func() {
			Expect(membership.SequenceNumberFromBuffer(nil)).Error().To(HaveOccurred())
		})

		It("should fail to read from buffer which is too small", func() {
			buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			for i := len(buffer) - 1; i >= 0; i-- {
				Expect(membership.SequenceNumberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
			}
		})
	})

	Context("IncarnationNumber", func() {
		It("should append to nil buffer", func() {
			buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should append to buffer", func() {
			var localBuffer [10]byte
			buffer, _, err := membership.AppendIncarnationNumberToBuffer(localBuffer[:0], 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		DescribeTable("should append to buffer with valid incarnation numbers",
			func(incarnationNumber int) {
				buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, incarnationNumber)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer).ToNot(BeNil())
			},
			Entry("zero", 0),
			Entry("small positive", 80),
			Entry("big positive", 3000),
		)

		DescribeTable("should fail to append to buffer with invalid incarnation numbers",
			func(incarnationNumber int) {
				Expect(membership.AppendIncarnationNumberToBuffer(nil, incarnationNumber)).Error().To(HaveOccurred())
			},
			Entry("negative", -10),
			Entry("too big of an incarnation number", math.MaxUint16+1),
		)

		It("should read from buffer", func() {
			appendIncarnationNumber := 1024
			buffer, appendN, err := membership.AppendIncarnationNumberToBuffer(nil, appendIncarnationNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			readIncarnationNumber, readN, err := membership.IncarnationNumberFromBuffer(buffer)
			Expect(err).ToNot(HaveOccurred())

			Expect(appendN).To(Equal(readN))
			Expect(appendIncarnationNumber).To(Equal(readIncarnationNumber))
		})

		It("should fail to read from nil buffer", func() {
			Expect(membership.IncarnationNumberFromBuffer(nil)).Error().To(HaveOccurred())
		})

		It("should fail to read from buffer which is too small", func() {
			buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			for i := len(buffer) - 1; i >= 0; i-- {
				Expect(membership.IncarnationNumberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
			}
		})
	})

	Context("MemberCount", func() {
		It("should append to nil buffer", func() {
			buffer, _, err := membership.AppendMemberCountToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		It("should append to buffer", func() {
			var localBuffer [10]byte
			buffer, _, err := membership.AppendMemberCountToBuffer(localBuffer[:0], 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		})

		DescribeTable("should append to buffer with valid member counts",
			func(memberCount int) {
				buffer, _, err := membership.AppendMemberCountToBuffer(nil, memberCount)
				Expect(err).ToNot(HaveOccurred())
				Expect(buffer).ToNot(BeNil())
			},
			Entry("zero", 0),
			Entry("small positive", 80),
			Entry("big positive", 3000),
		)

		DescribeTable("should fail to append to buffer with invalid member count",
			func(memberCount int) {
				Expect(membership.AppendMemberCountToBuffer(nil, memberCount)).Error().To(HaveOccurred())
			},
			Entry("negative", -10),
			Entry("too big of an incarnation number", math.MaxUint32+1),
		)

		It("should read from buffer", func() {
			appendMemberCount := 1024
			buffer, appendN, err := membership.AppendMemberCountToBuffer(nil, appendMemberCount)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			readMemberCount, readN, err := membership.MemberCountFromBuffer(buffer)
			Expect(err).ToNot(HaveOccurred())

			Expect(appendN).To(Equal(readN))
			Expect(appendMemberCount).To(Equal(readMemberCount))
		})

		It("should fail to read from nil buffer", func() {
			Expect(membership.MemberCountFromBuffer(nil)).Error().To(HaveOccurred())
		})

		It("should fail to read from buffer which is too small", func() {
			buffer, _, err := membership.AppendMemberCountToBuffer(nil, 1024)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())

			for i := len(buffer) - 1; i >= 0; i-- {
				Expect(membership.MemberCountFromBuffer(buffer[:i])).Error().To(HaveOccurred())
			}
		})
	})
})

func BenchmarkAppendSequenceNumberToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendSequenceNumberToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSequenceNumberFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendSequenceNumberToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.SequenceNumberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAppendIncarnationNumberToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendIncarnationNumberToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIncarnationNumberFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendIncarnationNumberToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.IncarnationNumberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAppendMemberCountToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendMemberCountToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemberCountFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendMemberCountToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.MemberCountFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
