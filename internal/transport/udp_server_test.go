package transport_test

import (
	"time"

	"github.com/backbone81/membership/internal/transport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDPServer", func() {
	It("should correctly receive data", func() {
		var target TestTarget
		server := transport.NewUDPServer(GinkgoLogr, &target, "localhost:0", 512)
		Expect(server.Startup()).To(Succeed())
		serverAddress, err := server.Addr()
		Expect(err).ToNot(HaveOccurred())

		client := transport.NewUDPClient(512)
		Expect(client.Send(serverAddress, []byte("foo bar"))).To(Succeed())
		time.Sleep(100 * time.Millisecond)

		Expect(server.Shutdown()).To(Succeed())
		Expect(target.DataReceived).To(Equal([]byte("foo bar")))
	})
})
