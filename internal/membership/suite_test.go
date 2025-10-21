package membership_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	TestAddress  = encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
	TestAddress2 = encoding.NewAddress(net.IPv4(11, 12, 13, 14), 1024)
	TestAddress3 = encoding.NewAddress(net.IPv4(21, 22, 23, 24), 1024)
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Membership Suite")
}

// DiscardClient provides a client transport which discards all data and always reports success. This is useful for
// tests and benchmarks, when we do not want to send network messages for real.
type DiscardClient struct{}

func (c *DiscardClient) Send(address encoding.Address, buffer []byte) error {
	return nil
}

// StoreClient provides a client transport which stores all data and always reports success. This is useful for
// tests, when we need to check if some specific data was transmitted.
type StoreClient struct {
	Addresses []encoding.Address
	Buffers   [][]byte
}

func (c *StoreClient) Send(address encoding.Address, buffer []byte) error {
	c.Addresses = append(c.Addresses, address)
	c.Buffers = append(c.Buffers, buffer)
	return nil
}
