package transport

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
)

// Error provides a client transport which always returns an error. This is useful for tests and benchmarks, when we do
// not want to send network messages for real.
type Error struct{}

// Discard implements Transport.
var _ Transport = (*Error)(nil)

func (d *Error) Send(address encoding.Address, buffer []byte) error {
	return errors.New("some transport error occurred")
}
