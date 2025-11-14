package gossip_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/gossip"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
)

var (
	TestAddress  = encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
	TestAddress2 = encoding.NewAddress(net.IPv4(11, 12, 13, 14), 1024)
	TestAddress3 = encoding.NewAddress(net.IPv4(21, 22, 23, 24), 1024)
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gossip Suite")
}

func GetMessageAt(queue *gossip.Queue, index int) gossip.Message {
	for i, msg := range queue.All() {
		if i == index {
			return msg
		}
	}
	panic("index out of bounds")
}
