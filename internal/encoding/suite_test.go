package encoding_test

import (
	"net"
	"testing"

	"github.com/backbone81/membership/internal/encoding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	TestAddress = encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encoding Suite")
}
