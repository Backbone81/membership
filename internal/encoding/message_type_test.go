package encoding_test

import (
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageType", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := encoding.AppendMessageTypeToBuffer(nil, encoding.MessageTypeDirectPing)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := encoding.AppendMessageTypeToBuffer(localBuffer[:0], encoding.MessageTypeDirectPing)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		buffer, appendN, err := encoding.AppendMessageTypeToBuffer(nil, encoding.MessageTypeDirectPing)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readMessageType, readN, err := encoding.MessageTypeFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(encoding.MessageTypeDirectPing).To(Equal(readMessageType))
	})

	It("should fail to read from nil buffer", func() {
		Expect(encoding.MessageTypeFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := encoding.AppendMessageTypeToBuffer(nil, encoding.MessageTypeDirectPing)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(encoding.MessageTypeFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendMessageTypeToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := encoding.AppendMessageTypeToBuffer(buffer[:0], encoding.MessageTypeDirectPing); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageTypeFromBuffer(b *testing.B) {
	buffer, _, err := encoding.AppendMessageTypeToBuffer(nil, encoding.MessageTypeDirectPing)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := encoding.MessageTypeFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
