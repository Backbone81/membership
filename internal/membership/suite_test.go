package membership_test

import (
	"math"
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/membership/internal/encoding"
)

var (
	TestAddress  = encoding.NewAddress(net.IPv4(1, 2, 3, 4), 1024)
	TestAddress2 = encoding.NewAddress(net.IPv4(11, 12, 13, 14), 1024)
	TestAddress3 = encoding.NewAddress(net.IPv4(21, 22, 23, 24), 1024)

	// BenchmarkAddress is an address which is coming last in a sorted list of members. Using this address in benchmarks
	// surfaces issues with slices scans, as the scan will take longer the more members are in a slice.
	BenchmarkAddress = encoding.NewAddress(net.IPv4(math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8), math.MaxUint16)
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Membership Suite")
}
