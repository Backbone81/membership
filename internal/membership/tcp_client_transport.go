package membership

import (
	"errors"
	"fmt"
	"math"
	"net"
)

type TCPClientTransport struct{}

func NewTCPClientTransport() *TCPClientTransport {
	return &TCPClientTransport{}
}

func (t *TCPClientTransport) Send(address Address, buffer []byte) error {
	if err := t.send(address, buffer); err != nil {
		return fmt.Errorf("TCP client transport send: %w", err)
	}
	return nil
}

func (t *TCPClientTransport) send(address Address, buffer []byte) error {
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
	Endian.PutUint32(lengthBuffer[:], uint32(len(buffer)))
	if _, err := connection.Write(lengthBuffer[:]); err != nil {
		return fmt.Errorf("sending the datagram length: %w", err)
	}
	if _, err := connection.Write(buffer); err != nil {
		return fmt.Errorf("sending the datagram payload: %w", err)
	}
	return nil
}
