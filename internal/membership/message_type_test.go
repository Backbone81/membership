package membership_test

import (
	"testing"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageType", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := membership.AppendMessageTypeToBuffer(nil, membership.MessageTypeDirectPing)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := membership.AppendMessageTypeToBuffer(localBuffer[:0], membership.MessageTypeDirectPing)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		appendMessageType := membership.MessageTypeDirectPing
		buffer, appendN, err := membership.AppendMessageTypeToBuffer(nil, appendMessageType)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readMessageType, readN, err := membership.MessageTypeFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendMessageType).To(Equal(readMessageType))
	})

	It("should fail to read from nil buffer", func() {
		Expect(membership.MessageTypeFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := membership.AppendMessageTypeToBuffer(nil, membership.MessageTypeDirectPing)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(membership.MessageTypeFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendMessageTypeToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendMessageTypeToBuffer(buffer[:0], membership.MessageTypeDirectPing); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageTypeFromBuffer(b *testing.B) {
	buffer, _, err := membership.AppendMessageTypeToBuffer(nil, membership.MessageTypeDirectPing)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.MessageTypeFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
