package membership

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
)

// Endian is the endianness membership uses for serializing/deserializing integers to network messages.
var Endian = binary.LittleEndian

// AppendIPToBuffer appends the ip to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendIPToBuffer(buffer []byte, ip net.IP) ([]byte, int, error) {
	ip = ip.To16()
	if ip == nil {
		return buffer, 0, errors.New("invalid ip")
	}
	buffer = append(buffer, []byte(ip)...)
	return buffer, net.IPv6len, nil
}

// IPFromBuffer reads the ip from the provided buffer.
// Returns the ip, the number of bytes read and any error which occurred.
func IPFromBuffer(buffer []byte) (net.IP, int, error) {
	if len(buffer) < net.IPv6len {
		return nil, 0, errors.New("ip buffer too small")
	}
	return buffer[:net.IPv6len], net.IPv6len, nil
}

// AppendPortToBuffer appends the port to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendPortToBuffer(buffer []byte, port int) ([]byte, int, error) {
	if port < 1 || math.MaxUint16 < port {
		return buffer, 0, errors.New("port out of bounds")
	}
	return Endian.AppendUint16(buffer, uint16(port)), 2, nil
}

// PortFromBuffer reads the port from the provided buffer.
// Returns the port, the number of bytes read and any error which occurred.
func PortFromBuffer(buffer []byte) (int, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("port buffer too small")
	}
	return int(Endian.Uint16(buffer)), 2, nil
}

// AppendSequenceNumberToBuffer appends the sequence number to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendSequenceNumberToBuffer(buffer []byte, sequenceNumber int) ([]byte, int, error) {
	if sequenceNumber < 0 || math.MaxUint16 < sequenceNumber {
		return buffer, 0, errors.New("sequence number out of bounds")
	}
	return Endian.AppendUint16(buffer, uint16(sequenceNumber)), 2, nil
}

// SequenceNumberFromBuffer reads the sequence number from the provided buffer.
// Returns the sequence number, the number of bytes read and any error which occurred.
func SequenceNumberFromBuffer(buffer []byte) (int, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("sequence buffer too small")
	}
	return int(Endian.Uint16(buffer)), 2, nil
}

// AppendIncarnationNumberToBuffer appends the incarnation number to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendIncarnationNumberToBuffer(buffer []byte, incarnationNumber int) ([]byte, int, error) {
	if incarnationNumber < 0 || math.MaxUint16 < incarnationNumber {
		return buffer, 0, errors.New("incarnation number out of bounds")
	}
	return Endian.AppendUint16(buffer, uint16(incarnationNumber)), 2, nil
}

// IncarnationNumberFromBuffer reads the incarnation number from the provided buffer.
// Returns the sequence number, the number of bytes read and any error which occurred.
func IncarnationNumberFromBuffer(buffer []byte) (int, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("incarnation buffer too small")
	}
	return int(Endian.Uint16(buffer)), 2, nil
}
