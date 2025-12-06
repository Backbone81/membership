package transport_test

import (
	"io"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/encryption"
	"github.com/backbone81/membership/internal/transport"
)

var _ = Describe("TCPClient", func() {
	var key1 encryption.Key

	BeforeEach(func() {
		key1 = encryption.NewRandomKey()
	})

	It("should not send plaintext", func() {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		Expect(err).ToNot(HaveOccurred())

		listener, err := net.ListenTCP("tcp", addr)
		Expect(err).ToNot(HaveOccurred())
		defer listener.Close() //nolint:errcheck
		listenerAddr, ok := listener.Addr().(*net.TCPAddr)
		Expect(ok).To(BeTrue())

		client, err := transport.NewTCPClient(key1)
		Expect(err).ToNot(HaveOccurred())

		payload := []byte("foo bar")
		Expect(client.Send(encoding.NewAddress(listenerAddr.IP, listenerAddr.Port), payload)).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		serverConnection, err := listener.Accept()
		Expect(err).ToNot(HaveOccurred())
		defer serverConnection.Close() //nolint:errcheck
		buffer, err := io.ReadAll(serverConnection)
		Expect(err).ToNot(HaveOccurred())

		Expect(buffer).ToNot(Equal(payload))
		Expect(buffer).To(HaveLen(4 + encryption.Overhead + len(payload) + encryption.Overhead))
	})

	It("should encrypt the message in a different way when sent multiple times", func() {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		Expect(err).ToNot(HaveOccurred())

		listener, err := net.ListenTCP("tcp", addr)
		Expect(err).ToNot(HaveOccurred())
		defer listener.Close() //nolint:errcheck
		listenerAddr, ok := listener.Addr().(*net.TCPAddr)
		Expect(ok).To(BeTrue())

		client, err := transport.NewTCPClient(key1)
		Expect(err).ToNot(HaveOccurred())

		payload := []byte("foo bar")
		Expect(client.Send(encoding.NewAddress(listenerAddr.IP, listenerAddr.Port), payload)).To(Succeed())
		Expect(client.Send(encoding.NewAddress(listenerAddr.IP, listenerAddr.Port), payload)).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		serverConnection1, err := listener.Accept()
		Expect(err).ToNot(HaveOccurred())
		defer serverConnection1.Close() //nolint:errcheck
		buffer1, err := io.ReadAll(serverConnection1)
		Expect(err).ToNot(HaveOccurred())

		serverConnection2, err := listener.Accept()
		Expect(err).ToNot(HaveOccurred())
		defer serverConnection2.Close() //nolint:errcheck
		buffer2, err := io.ReadAll(serverConnection2)
		Expect(err).ToNot(HaveOccurred())

		Expect(buffer1).ToNot(Equal(buffer2))
	})
})
