package membership

import (
	"errors"
	"fmt"
	"net"
)

type UDPClientTransport struct {
	maxDatagramLength int
}

func NewUDPClientTransport(maxDatagramLength int) *UDPClientTransport {
	return &UDPClientTransport{
		maxDatagramLength: maxDatagramLength,
	}
}

func (t *UDPClientTransport) Send(address Address, buffer []byte) error {
	if err := t.send(address, buffer); err != nil {
		return fmt.Errorf("UDP client transport send: %w", err)
	}
	return nil
}

func (t *UDPClientTransport) send(address Address, buffer []byte) error {
	if len(buffer) > t.maxDatagramLength {
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
