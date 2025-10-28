package transport

import (
	"math/rand"

	"github.com/backbone81/membership/internal/encoding"
)

// Unreliable provides a client transport which forwards the call to some other transport only some time. The other
// time the transmission is silently dropped. This helps in simulating a lossy network.
type Unreliable struct {
	// Transport is the transport to forward calls to.
	Transport Transport

	// Reliability is how often a send request is forwarded to the transport. Value is 0.0 to 1.0 with 0.0 meaning that
	// sends are never forwarded, 0.5 means that sends are forwarded 50% of the time and 1.0 means that sends are always
	// forwarded.
	Reliability float64
}

// Unreliable implements Transport
var _ Transport = (*Unreliable)(nil)

func (u *Unreliable) Send(address encoding.Address, buffer []byte) error {
	if rand.Float64() > u.Reliability {
		// We exceeded the reliability, so we drop the send and exit early.
		return nil
	}
	return u.Transport.Send(address, buffer)
}
