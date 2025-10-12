package transport

import (
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/backbone81/membership/internal/membership"
)

// TCPClient provides reliable transport for sending data to a member.
type TCPClient struct{}

// NewTCPClient creates a new TCPClient transport.
func NewTCPClient() *TCPClient {
	return &TCPClient{}
}

// Send transmits the given buffer to the member with the given address.
func (t *TCPClient) Send(address membership.Address, buffer []byte) error {
	if err := t.send(address, buffer); err != nil {
		return fmt.Errorf("TCP client transport send: %w", err)
	}
	return nil
}

func (t *TCPClient) send(address membership.Address, buffer []byte) error {
	// Make sure we are not exceeding the maximum datagram length with the given buffer.
	if len(buffer) > math.MaxUint32 {
		return errors.New("buffer length exceeds maximum datagram length")
	}

	connection, err := net.Dial("tcp", address.String())
	if err != nil {
		return fmt.Errorf("connecting to remote host at %q: %w", address, err)
	}
	defer connection.Close()

	var lengthBuffer [4]byte
	membership.Endian.PutUint32(lengthBuffer[:], uint32(len(buffer)))
	if _, err := connection.Write(lengthBuffer[:]); err != nil {
		return fmt.Errorf("sending the datagram length: %w", err)
	}
	if _, err := connection.Write(buffer); err != nil {
		return fmt.Errorf("sending the datagram payload: %w", err)
	}
	return nil
}
