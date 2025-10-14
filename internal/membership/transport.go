package membership

import "github.com/backbone81/membership/internal/encoding"

// Transport is the interface the transport needs to implement for transmitting data between members.
type Transport interface {
	Send(address encoding.Address, buffer []byte) error
}
