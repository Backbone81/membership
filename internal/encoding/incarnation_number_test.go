package encoding_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
)

var _ = Describe("IncarnationNumber", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := encoding.AppendIncarnationNumberToBuffer(nil, 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := encoding.AppendIncarnationNumberToBuffer(localBuffer[:0], 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	DescribeTable("should append to buffer with valid incarnation numbers",
		func(incarnationNumber int) {
			buffer, _, err := encoding.AppendIncarnationNumberToBuffer(nil, uint16(incarnationNumber))
			Expect(err).ToNot(HaveOccurred())
			Expect(buffer).ToNot(BeNil())
		},
		Entry("zero", 0),
		Entry("small positive", 80),
		Entry("big positive", 3000),
	)

	It("should read from buffer", func() {
		appendIncarnationNumber := uint16(1024)
		buffer, appendN, err := encoding.AppendIncarnationNumberToBuffer(nil, appendIncarnationNumber)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readIncarnationNumber, readN, err := encoding.IncarnationNumberFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendIncarnationNumber).To(Equal(readIncarnationNumber))
	})

	It("should fail to read from nil buffer", func() {
		Expect(encoding.IncarnationNumberFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := encoding.AppendIncarnationNumberToBuffer(nil, 1024)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(encoding.IncarnationNumberFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendIncarnationNumberToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := encoding.AppendIncarnationNumberToBuffer(buffer[:0], 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIncarnationNumberFromBuffer(b *testing.B) {
	buffer, _, err := encoding.AppendIncarnationNumberToBuffer(nil, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := encoding.IncarnationNumberFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
