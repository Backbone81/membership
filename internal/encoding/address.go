package encoding

import (
	"bytes"
	"errors"
	"math"
	"net"
	"strconv"
)

// Address represents the combination of ip and port.
// It is important that this type can be used as a key into maps and is orderable.
// To make this possible and also to reduce the amount of memory allocations, we use an array of bytes in which we pack
// the ip and the port.
type Address [net.IPv6len + 2]byte

// ZeroAddress provides a zero valued address. This can be used to check for zero values.
var ZeroAddress Address

// NewAddress creates a new address with the given ip and port.
// Panics if the ip or port are out of range.
func NewAddress(ip net.IP, port int) Address {
	if ip.IsUnspecified() || port < 0 || port > math.MaxUint16 {

		// TODO: Do we really want to panic here?

		panic("invalid address")
	}
	var result Address
	copy(result[:net.IPv6len], ip.To16())
	Endian.PutUint16(result[net.IPv6len:], uint16(port))
	return result
}

// IP returns the ip of the address.
func (a Address) IP() net.IP {
	return a[:net.IPv6len]
}

// Port returns the port of the address.
func (a Address) Port() int {
	return int(Endian.Uint16(a[net.IPv6len:]))
}

// Equal reports if two addresses are the same.
func (a Address) Equal(address Address) bool {
	return bytes.Equal(a[:], address[:])
}

// IsZero reports if the address is its zero value.
func (a Address) IsZero() bool {
	return a.Equal(ZeroAddress)
}

// String returns a string representing the address with ip and port joined together.
func (a Address) String() string {
	return net.JoinHostPort(a.IP().String(), strconv.Itoa(a.Port()))
}

// AppendAddressToBuffer appends the address to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendAddressToBuffer(buffer []byte, address Address) ([]byte, int, error) {
	buffer = append(buffer, address[:]...)
	return buffer, len(address), nil
}

// AddressFromBuffer reads the address from the provided buffer.
// Returns the address, the number of bytes read and any error which occurred.
func AddressFromBuffer(buffer []byte) (Address, int, error) {
	var result Address
	if len(buffer) < len(result) {
		return Address{}, 0, errors.New("address buffer too small")
	}
	copy(result[:], buffer[:len(result)])
	return result, len(result), nil
}
