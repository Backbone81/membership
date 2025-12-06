package membership

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/backbone81/membership/internal/encoding"
)

func ResolveAdvertiseAddress(advertiseAddress string, bindAddress string) (encoding.Address, error) {
	if advertiseAddress != "" {
		return resolveAdvertiseAddress(advertiseAddress)
	}
	return resolveAdvertiseAddressFromBindAddress(bindAddress)
}

func resolveAdvertiseAddress(advertiseAddress string) (encoding.Address, error) {
	addr, err := net.ResolveUDPAddr("udp", advertiseAddress)
	if err != nil {
		return encoding.Address{}, fmt.Errorf("resolving advertise address: %w", err)
	}
	return encoding.NewAddress(
		addr.IP,
		addr.Port,
	), nil
}

func resolveAdvertiseAddressFromBindAddress(bindAddress string) (encoding.Address, error) {
	_, port, err := net.SplitHostPort(bindAddress)
	if err != nil {
		return encoding.Address{}, err
	}
	typedPort, err := strconv.Atoi(port)
	if err != nil {
		return encoding.Address{}, err
	}

	localIp, err := getLocalIPAddress()
	if err != nil {
		return encoding.Address{}, err
	}
	return encoding.NewAddress(
		localIp,
		typedPort,
	), nil
}

func getLocalIPAddress() (net.IP, error) {
	connection, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer connection.Close() //nolint:errcheck

	localAddr, ok := connection.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, errors.New("expected net.UDPAddr to be returned by local address")
	}
	return localAddr.IP, nil
}

func ResolveBootstrapMembers(bootstrapMembers []string) ([]encoding.Address, error) {
	var joinedErr error
	result := make([]encoding.Address, len(bootstrapMembers))
	for i, bootstrapMember := range bootstrapMembers {
		addr, err := net.ResolveUDPAddr("udp", bootstrapMember)
		if err != nil {
			joinedErr = errors.Join(joinedErr, fmt.Errorf("resolving member %q: %w", bootstrapMember, err))
			continue
		}
		result[i] = encoding.NewAddress(
			addr.IP,
			addr.Port,
		)
	}
	return result, joinedErr
}
