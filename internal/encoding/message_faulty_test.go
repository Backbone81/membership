package encoding_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testMessageFaulty = encoding.MessageFaulty{
	Source:            encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
	Destination:       encoding.NewAddress(net.IPv4(11, 12, 13, 14), 1024),
	IncarnationNumber: 7,
}

var _ = Describe("MessageFaulty", func() {
	It("should append to nil buffer", func() {
		buffer, _, err := testMessageFaulty.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := testMessageFaulty.AppendToBuffer(localBuffer[:0])
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		buffer, appendN, err := testMessageFaulty.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		var readMessage encoding.MessageFaulty
		readN, err := readMessage.FromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(testMessageFaulty).To(Equal(readMessage))
	})

	It("should fail to read from nil buffer", func() {
		var readMessage encoding.MessageFaulty
		Expect(readMessage.FromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := testMessageFaulty.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(testMessageFaulty.FromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkMessageFaulty_AppendToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := testMessageFaulty.AppendToBuffer(buffer[:0]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageFaulty_FromBuffer(b *testing.B) {
	buffer, _, err := testMessageFaulty.AppendToBuffer(nil)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, err := testMessageFaulty.FromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
