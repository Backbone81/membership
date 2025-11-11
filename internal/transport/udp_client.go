package transport

import (
	"errors"
	"fmt"
	"net"

	"github.com/backbone81/membership/internal/encoding"
)

// UDPClient provides unreliable transport for sending data to a member.
type UDPClient struct {
	maxDatagramLength int
}

// UDPClient implements Transport.
var _ Transport = (*UDPClient)(nil)

// NewUDPClient creates a new UDPClient transport.
func NewUDPClient(maxDatagramLength int) *UDPClient {
	return &UDPClient{
		maxDatagramLength: maxDatagramLength,
	}
}

// Send transmits the given buffer to the member with the given address. The length of the buffer is validated against
// the max datagram length provided during construction. If the length exceeds the maximum length, no data is sent
// and an error is returned.
func (c *UDPClient) Send(address encoding.Address, buffer []byte) error {
	if err := c.send(address, buffer); err != nil {
		return fmt.Errorf("UDP client transport send: %w", err)
	}
	return nil
}

func (c *UDPClient) send(address encoding.Address, buffer []byte) error {
	if len(buffer) > c.maxDatagramLength {
		return errors.New("buffer length exceeds maximum datagram length")
	}

	connection, err := net.Dial("udp", address.String())
	if err != nil {
		return fmt.Errorf("connecting to remote host at %q: %w", address, err)
	}
	defer connection.Close()

	if _, err := connection.Write(buffer); err != nil {
		return fmt.Errorf("sending the datagram payload: %w", err)
	}
	return nil
}
