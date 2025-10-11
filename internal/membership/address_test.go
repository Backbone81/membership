package membership_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Address", func() {
	It("should correctly store ip and port", func() {
		ip := net.IPv4(1, 2, 3, 4)
		port := 1024
		address := membership.NewAddress(ip, port)
		Expect(address.IP()).To(Equal(ip))
		Expect(address.Port()).To(Equal(port))
	})

	It("should correctly report identical addresses", func() {
		address1 := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		address2 := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		address3 := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1025)
		address4 := membership.NewAddress(net.IPv4(1, 2, 3, 5), 1024)
		Expect(address1.Equal(address2)).To(BeTrue())
		Expect(address1.Equal(address3)).To(BeFalse())
		Expect(address1.Equal(address4)).To(BeFalse())
	})

	It("should correctly report zero values", func() {
		Expect(membership.Address{}.IsZero()).To(BeTrue())
		Expect(membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024).IsZero()).ToNot(BeTrue())
	})

	It("should correctly return a string", func() {
		address := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		Expect(address.String()).To(Equal("1.2.3.4:1024"))
	})

	It("should append to nil buffer", func() {
		address := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		buffer, _, err := membership.AppendAddressToBuffer(nil, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should append to buffer", func() {
		var localBuffer [10]byte
		address := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		buffer, _, err := membership.AppendAddressToBuffer(localBuffer[:0], address)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})

	It("should read from buffer", func() {
		appendAddress := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		buffer, appendN, err := membership.AppendAddressToBuffer(nil, appendAddress)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		readAddress, readN, err := membership.AddressFromBuffer(buffer)
		Expect(err).ToNot(HaveOccurred())

		Expect(appendN).To(Equal(readN))
		Expect(appendAddress).To(Equal(readAddress))
	})

	It("should fail to read from nil buffer", func() {
		Expect(membership.AddressFromBuffer(nil)).Error().To(HaveOccurred())
	})

	It("should fail to read from buffer which is too small", func() {
		address := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
		buffer, _, err := membership.AppendAddressToBuffer(nil, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())

		for i := len(buffer) - 1; i >= 0; i-- {
			Expect(membership.AddressFromBuffer(buffer[:i])).Error().To(HaveOccurred())
		}
	})
})

func BenchmarkAppendAddressToBuffer(b *testing.B) {
	address := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
	var buffer [1024]byte
	for b.Loop() {
		if _, _, err := membership.AppendAddressToBuffer(buffer[:0], address); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddressFromBuffer(b *testing.B) {
	address := membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
	buffer, _, err := membership.AppendAddressToBuffer(nil, address)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, _, err := membership.AddressFromBuffer(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
