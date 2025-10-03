package membership

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
)

// Endian is the endianness membership uses for serializing/deserializing integers to network.
var Endian = binary.LittleEndian

func AppendIPToBuffer(buffer []byte, ip net.IP) ([]byte, int, error) {
	ip = ip.To16()
	if ip == nil {
		return buffer, 0, errors.New("invalid ip")
	}
	buffer = append(buffer, []byte(ip)...)
	return buffer, net.IPv6len, nil
}

func IPFromBuffer(buffer []byte) (net.IP, int, error) {
	if len(buffer) < net.IPv6len {
		return nil, 0, errors.New("ip buffer too small")
	}
	return buffer[:net.IPv6len], net.IPv6len, nil
}

func AppendPortToBuffer(buffer []byte, port int) ([]byte, int, error) {
	if port < 1 || math.MaxUint16 < port {
		return buffer, 0, errors.New("port out of bounds")
	}
	return Endian.AppendUint16(buffer, uint16(port)), 2, nil
}

func PortFromBuffer(buffer []byte) (int, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("port buffer too small")
	}
	return int(Endian.Uint16(buffer)), 2, nil
}

func AppendSequenceNumberToBuffer(buffer []byte, sequenceNumber int) ([]byte, int, error) {
	if sequenceNumber < 0 || math.MaxUint16 < sequenceNumber {
		return buffer, 0, errors.New("sequence number out of bounds")
	}
	return Endian.AppendUint16(buffer, uint16(sequenceNumber)), 2, nil
}

func SequenceNumberFromBuffer(buffer []byte) (int, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("sequence buffer too small")
	}
	return int(Endian.Uint16(buffer)), 2, nil
}

func AppendIncarnationNumberToBuffer(buffer []byte, incarnationNumber int) ([]byte, int, error) {
	if incarnationNumber < 0 || math.MaxUint16 < incarnationNumber {
		return buffer, 0, errors.New("incarnation number out of bounds")
	}
	return Endian.AppendUint16(buffer, uint16(incarnationNumber)), 2, nil
}

func IncarnationNumberFromBuffer(buffer []byte) (int, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("incarnation buffer too small")
	}
	return int(Endian.Uint16(buffer)), 2, nil
}
