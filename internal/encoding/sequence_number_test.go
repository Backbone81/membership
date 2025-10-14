package encoding_test

import (
	"math"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SequenceNumber", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := encoding.AppendSequenceNumberToBuffer(nil, 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := encoding.AppendSequenceNumberToBuffer(localBuffer[:0], 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	DescribeTable("should append to buffer with valid sequence numbers",
		func(sequenceNumber int) {
			buffer, _, err := encoding.AppendSequenceNumberToBuffer(nil, sequenceNumber)
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		},
		Entry("zero", 0),
		Entry("small positive", 80),
		Entry("big positive", 3000),
	)

	DescribeTable("should fail to append to buffer with invalid sequence numbers",
		func(sequenceNumber int) {
			Expect(encoding.AppendSequenceNumberToBuffer(nil, sequenceNumber)).Error().To(HaveOccurred())
		},
		Entry("negative", -10),
		Entry("too big of a sequence number", math.MaxUint16+1),
	)

	It("should read from buffer", func() {
		appendSequenceNumber := 1024
		buffer, appendN, err := encoding.AppendSequenceNumberToBuffer(nil, appendSequenceNumber)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readSequenceNumber, readN, err := encoding.SequenceNumberFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendSequenceNumber).To(Equal(readSequenceNumber))
	})

	It("should fail to read from nil buffer", func() {
		Expect(encoding.SequenceNumberFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := encoding.AppendSequenceNumberToBuffer(nil, 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(encoding.SequenceNumberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendSequenceNumberToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := encoding.AppendSequenceNumberToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSequenceNumberFromBuffer(b *testing.B) {
	buffer, _, err := encoding.AppendSequenceNumberToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := encoding.SequenceNumberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
