package membership_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageListRequest", func() {
	It("should append to nil buffer", func() {
		message := membership.MessageListRequest{
			Source: membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		message := membership.MessageListRequest{
			Source: membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
		}
		buffer, _, err := message.AppendToBuffer(localBuffer[:0])
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		appendMessage := membership.MessageListRequest{
			Source: membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
		}
		buffer, appendN, err := appendMessage.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		var readMessage membership.MessageListRequest
		readN, err := readMessage.FromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendMessage).To(Equal(readMessage))
	})

	It("should fail to read from nil buffer", func() {
		var readMessage membership.MessageListRequest
		Expect(readMessage.FromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		message := membership.MessageListRequest{
			Source: membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(message.FromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkMessageListRequest_AppendToBuffer(b *testing.B) {
	message := membership.MessageListRequest{
		Source: membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
	}
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := message.AppendToBuffer(buffer[:0]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageListRequest_FromBuffer(b *testing.B) {
	message := membership.MessageListRequest{
		Source: membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
	}
	buffer, _, err := message.AppendToBuffer(nil)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, err := message.FromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
