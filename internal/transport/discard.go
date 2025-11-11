package transport

import "github.com/backbone81/membership/internal/encoding"

// Discard provides a client transport which discards all data and always reports success. This is useful for
// tests and benchmarks, when we do not want to send network messages for real.
type Discard struct{}

// Discard implements Transport.
var _ Transport = (*Discard)(nil)

func (d *Discard) Send(address encoding.Address, buffer []byte) error {
	return nil
}
