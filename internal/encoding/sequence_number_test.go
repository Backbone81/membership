package encoding_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
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
			buffer, _, err := encoding.AppendSequenceNumberToBuffer(nil, uint16(sequenceNumber))
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		},
		Entry("zero", 0),
		Entry("small positive", 80),
		Entry("big positive", 3000),
	)

	It("should read from buffer", func() {
		buffer, appendN, err := encoding.AppendSequenceNumberToBuffer(nil, 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readSequenceNumber, readN, err := encoding.SequenceNumberFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(uint16(1024)).To(Equal(readSequenceNumber))
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
