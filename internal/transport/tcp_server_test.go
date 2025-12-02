package transport_test

import (
	"io"
	"net"
	"time"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/encryption"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/transport"
)

var _ = Describe("TCPServer", func() {
	var (
		key1 encryption.Key
		key2 encryption.Key
		key3 encryption.Key
	)

	BeforeEach(func() {
		key1 = encryption.NewRandomKey()
		key2 = encryption.NewRandomKey()
		key3 = encryption.NewRandomKey()
	})

	It("should correctly receive data with the same key", func() {
		var target TestTarget
		server, err := transport.NewTCPServer(GinkgoLogr, &target, "localhost:0", []encryption.Key{key1})
		Expect(err).ToNot(HaveOccurred())
		Expect(server.Startup()).To(Succeed())
		serverAddress, err := server.Addr()
		Expect(err).ToNot(HaveOccurred())

		client, err := transport.NewTCPClient(key1)
		Expect(err).ToNot(HaveOccurred())
		Expect(client.Send(serverAddress, []byte("foo bar"))).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		Expect(server.Shutdown()).To(Succeed())
		Expect(target.DataReceived).To(Equal([]byte("foo bar")))
	})

	It("should support additional keys", func() {
		var target TestTarget
		server, err := transport.NewTCPServer(GinkgoLogr, &target, "localhost:0", []encryption.Key{key1, key2, key3})
		Expect(err).ToNot(HaveOccurred())
		Expect(server.Startup()).To(Succeed())
		serverAddress, err := server.Addr()
		Expect(err).ToNot(HaveOccurred())

		client, err := transport.NewTCPClient(key1)
		Expect(err).ToNot(HaveOccurred())
		Expect(client.Send(serverAddress, []byte("foo bar"))).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		Expect(server.Shutdown()).To(Succeed())
		Expect(target.DataReceived).To(Equal([]byte("foo bar")))
	})

	It("should try all keys to decrypt", func() {
		var target TestTarget
		server, err := transport.NewTCPServer(GinkgoLogr, &target, "localhost:0", []encryption.Key{key1, key2, key3})
		Expect(err).ToNot(HaveOccurred())
		Expect(server.Startup()).To(Succeed())
		serverAddress, err := server.Addr()
		Expect(err).ToNot(HaveOccurred())

		client, err := transport.NewTCPClient(key3)
		Expect(err).ToNot(HaveOccurred())
		Expect(client.Send(serverAddress, []byte("foo bar"))).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		Expect(server.Shutdown()).To(Succeed())
		Expect(target.DataReceived).To(Equal([]byte("foo bar")))
	})

	It("should fail to decrypt with the wrong key", func() {
		var target TestTarget
		server, err := transport.NewTCPServer(GinkgoLogr, &target, "localhost:0", []encryption.Key{key1})
		Expect(err).ToNot(HaveOccurred())
		Expect(server.Startup()).To(Succeed())
		serverAddress, err := server.Addr()
		Expect(err).ToNot(HaveOccurred())

		client, err := transport.NewTCPClient(key2)
		Expect(err).ToNot(HaveOccurred())
		Expect(client.Send(serverAddress, []byte("foo bar"))).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		Expect(server.Shutdown()).To(Succeed())
		Expect(target.DataReceived).To(BeEmpty())
	})

	It("should fail to decrypt when the message has been tampered with", func() {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		Expect(err).ToNot(HaveOccurred())

		listener, err := net.ListenTCP("tcp", addr)
		Expect(err).ToNot(HaveOccurred())
		defer listener.Close()
		listenerAddr := listener.Addr().(*net.TCPAddr)

		client, err := transport.NewTCPClient(key1)
		Expect(err).ToNot(HaveOccurred())
		Expect(client.Send(encoding.NewAddress(listenerAddr.IP, listenerAddr.Port), []byte("foo bar"))).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		serverConnection, err := listener.Accept()
		Expect(err).ToNot(HaveOccurred())
		defer serverConnection.Close()
		buffer, err := io.ReadAll(serverConnection)
		Expect(err).ToNot(HaveOccurred())

		var target TestTarget
		server, err := transport.NewTCPServer(GinkgoLogr, &target, "localhost:0", []encryption.Key{key1})
		Expect(err).ToNot(HaveOccurred())
		Expect(server.Startup()).To(Succeed())
		serverAddress, err := server.Addr()
		Expect(err).ToNot(HaveOccurred())

		clientConnection, err := net.Dial("tcp", serverAddress.String())
		Expect(err).ToNot(HaveOccurred())
		defer clientConnection.Close()

		// We send one byte less than the full message.
		Expect(clientConnection.Write(buffer[:len(buffer)-1])).Error().ToNot(HaveOccurred())
		time.Sleep(100 * time.Millisecond)
		clientConnection.Close()

		Expect(server.Shutdown()).To(Succeed())
		Expect(target.DataReceived).To(BeEmpty())
	})
})
