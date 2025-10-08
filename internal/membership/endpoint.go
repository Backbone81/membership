package membership

import (
	"net"
	"slices"
	"strconv"
)

type Endpoint struct {
	IP   net.IP
	Port int
}

func (e Endpoint) Equal(endpoint Endpoint) bool {
	return slices.Equal(e.IP, endpoint.IP) && e.Port == endpoint.Port
}

func (e Endpoint) String() string {
	return net.JoinHostPort(e.IP.String(), strconv.Itoa(e.Port))
}

func (e Endpoint) IsEmpty() bool {
	return e.Port == 0
}

func AppendEndpointToBuffer(buffer []byte, endpoint Endpoint) ([]byte, int, error) {
	ipBuffer, ipN, err := AppendIPToBuffer(buffer, endpoint.IP)
	if err != nil {
		return buffer, 0, err
	}

	portBuffer, portN, err := AppendPortToBuffer(ipBuffer, endpoint.Port)
	if err != nil {
		return buffer, 0, err
	}
	return portBuffer, ipN + portN, nil
}

func EndpointFromBuffer(buffer []byte) (Endpoint, int, error) {
	ip, ipN, err := IPFromBuffer(buffer)
	if err != nil {
		return Endpoint{}, 0, err
	}

	port, portN, err := PortFromBuffer(buffer[ipN:])
	if err != nil {
		return Endpoint{}, 0, err
	}
	return Endpoint{
		IP:   ip,
		Port: port,
	}, ipN + portN, nil
}
