package membership

import (
	"fmt"
	"net"
)

type ClientTransport struct{
	connection *net.UDPConn
}

func NewClientTransport(connection *net.UDPConn) *ClientTransport {
	return &ClientTransport{
		connection: connection,
	}
}

func (t *ClientTransport) UpdateConnection(connection *net.UDPConn) {
	t.connection = connection
}

func (t *ClientTransport) Send(endpoint Endpoint, buffer []byte) error {
	_, err := t.connection.WriteToUDP(buffer, &net.UDPAddr{
		IP:   endpoint.IP,
		Port: endpoint.Port,
	})
	if err != nil {
		return fmt.Errorf("sending UDP: %w", err)
	}
	return nil
}
