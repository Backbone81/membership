package scheduler_test

import (
	"net"

	"github.com/backbone81/membership/internal/membership"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scheduler", func() {
	It("should append to nil buffer", func() {
		message := membership.MessageSuspect{
			Source:            membership.NewAddress(net.IPv4(1, 2, 3, 4), 1024),
			Destination:       membership.NewAddress(net.IPv4(11, 12, 13, 14), 1024),
			IncarnationNumber: 7,
		}
		buffer, _, err := message.AppendToBuffer(nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(buffer).ToNot(BeNil())
	})
})
