package membership

import (
	"fmt"
	"net"
)

type ClientTransport struct{}

func (t *ClientTransport) Send(endpoint Endpoint, buffer []byte) error {
	var connection net.UDPConn
	_, err := connection.WriteToUDP(buffer, &net.UDPAddr{
		IP:   endpoint.IP,
		Port: endpoint.Port,
	})
	if err != nil {
		return fmt.Errorf("sending UDP: %w", err)
	}
	return nil
}
