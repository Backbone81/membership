package gossip_test

import (
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/gossip"
)

var testMessageSuspect = gossip.MessageSuspect{
	Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
	Destination:       encoding.NewAddress(net.IPv4(11, 12, 13, 14), 1024),
	IncarnationNumber: 7,
}

var _ = Describe("MessageSuspect", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := testMessageSuspect.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := testMessageSuspect.AppendToBuffer(localBuffer[:0])
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		buffer, appendN, err := testMessageSuspect.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		var readMessage gossip.MessageSuspect
		readN, err := readMessage.FromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(testMessageSuspect).To(Equal(readMessage))
	})

	It("should fail to read from nil buffer", func() {
		var readMessage gossip.MessageSuspect
		Expect(readMessage.FromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := testMessageSuspect.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(testMessageSuspect.FromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkMessageSuspect_AppendToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := testMessageSuspect.AppendToBuffer(buffer[:0]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageSuspect_FromBuffer(b *testing.B) {
	buffer, _, err := testMessageSuspect.AppendToBuffer(nil)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, err := testMessageSuspect.FromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
