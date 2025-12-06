package transport_test

import (
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/encryption"
	"github.com/backbone81/membership/internal/transport"
)

var _ = Describe("UDPClient", func() {
	var key1 encryption.Key

	BeforeEach(func() {
		key1 = encryption.NewRandomKey()
	})

	It("should not send plaintext", func() {
		addr, err := net.ResolveUDPAddr("udp", "localhost:0")
		Expect(err).ToNot(HaveOccurred())

		listener, err := net.ListenUDP("udp", addr)
		Expect(err).ToNot(HaveOccurred())
		defer listener.Close() //nolint:errcheck
		listenerAddr, ok := listener.LocalAddr().(*net.UDPAddr)
		Expect(ok).To(BeTrue())

		client, err := transport.NewUDPClient(512, key1)
		Expect(err).ToNot(HaveOccurred())

		payload := []byte("foo bar")
		Expect(client.Send(encoding.NewAddress(listenerAddr.IP, listenerAddr.Port), payload)).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		var buffer [512]byte
		n, _, err := listener.ReadFromUDP(buffer[:])
		Expect(err).ToNot(HaveOccurred())

		Expect(buffer[:n]).ToNot(Equal(payload))
		Expect(buffer[:n]).To(HaveLen(len(payload) + encryption.Overhead))
	})

	It("should encrypt the same message in a different way when sent multiple times", func() {
		addr, err := net.ResolveUDPAddr("udp", "localhost:0")
		Expect(err).ToNot(HaveOccurred())

		listener, err := net.ListenUDP("udp", addr)
		Expect(err).ToNot(HaveOccurred())
		defer listener.Close() //nolint:errcheck
		listenerAddr, ok := listener.LocalAddr().(*net.UDPAddr)
		Expect(ok).To(BeTrue())

		client, err := transport.NewUDPClient(512, key1)
		Expect(err).ToNot(HaveOccurred())

		payload := []byte("foo bar")
		Expect(client.Send(encoding.NewAddress(listenerAddr.IP, listenerAddr.Port), payload)).To(Succeed())
		Expect(client.Send(encoding.NewAddress(listenerAddr.IP, listenerAddr.Port), payload)).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		var buffer1 [512]byte
		n1, _, err := listener.ReadFromUDP(buffer1[:])
		Expect(err).ToNot(HaveOccurred())

		var buffer2 [512]byte
		n2, _, err := listener.ReadFromUDP(buffer2[:])
		Expect(err).ToNot(HaveOccurred())

		Expect(buffer1[:n1]).ToNot(Equal(buffer2[:n2]))
	})
})
