package encoding_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Address", func() {
	It("should correctly store ip and port", func() {
		ip := net.IPv4(1, 2, 3, 4)
		port := 1024
		address := encoding.NewAddress(ip, port)
		Expect(address.IP()).To(Equal(ip))
		Expect(address.Port()).To(Equal(port))
	})

	It("should correctly report identical addresses", func() {
		address1 := encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		address2 := encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		address3 := encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1025)
		address4 := encoding.NewAddress(net.IPv4(1, 2, 3, 5), 1024)
		Expect(address1.Equal(address2)).To(BeTrue())
		Expect(address1.Equal(address3)).To(BeFalse())
		Expect(address1.Equal(address4)).To(BeFalse())
	})

	It("should correctly report zero values", func() {
		Expect(encoding.Address{}.IsZero()).To(BeTrue())
		Expect(TestAddress.IsZero()).ToNot(BeTrue())
	})

	It("should correctly return a string", func() {
		Expect(TestAddress.String()).To(Equal("1.2.3.4:1024"))
	})

	It("should append to nil buffer", func() {
		buffer, _, err := encoding.AppendAddressToBuffer(nil, TestAddress)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		buffer, _, err := encoding.AppendAddressToBuffer(localBuffer[:0], TestAddress)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		buffer, appendN, err := encoding.AppendAddressToBuffer(nil, TestAddress)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readAddress, readN, err := encoding.AddressFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(TestAddress).To(Equal(readAddress))
	})

	It("should fail to read from nil buffer", func() {
		Expect(encoding.AddressFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		buffer, _, err := encoding.AppendAddressToBuffer(nil, TestAddress)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(encoding.AddressFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendAddressToBuffer(b *testing.B) {
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := encoding.AppendAddressToBuffer(buffer[:0], TestAddress); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddressFromBuffer(b *testing.B) {
	buffer, _, err := encoding.AppendAddressToBuffer(nil, TestAddress)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := encoding.AddressFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
