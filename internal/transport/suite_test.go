package transport_test

import (
	"testing"

	"github.com/backbone81/membership/internal/transport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transport Suite")
}

// TestTarget provides a target implementation for testing the scheduler without actually running the full membership
// algorithm.
type TestTarget struct {
	DataReceived []byte
}

// TestTarget implements scheduler.Target
var _ transport.Target = (*TestTarget)(nil)

func (t *TestTarget) DispatchDatagram(buffer []byte) error {
	t.DataReceived = buffer
	return nil
}
