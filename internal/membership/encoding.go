package membership

import (
	"encoding/binary"
	"errors"
	"math"
)

// Endian is the endianness membership uses for serializing/deserializing integers to network messages.
var Endian = binary.LittleEndian

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

// AppendMemberCountToBuffer appends the number of members to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendMemberCountToBuffer(buffer []byte, memberCount int) ([]byte, int, error) {
	if memberCount < 0 || math.MaxUint32 < memberCount {
		return buffer, 0, errors.New("member count out of bounds")
	}
	return Endian.AppendUint32(buffer, uint32(memberCount)), 4, nil
}

// MemberCountFromBuffer reads the number of members from the provided buffer.
// Returns the number of members, the number of bytes read and any error which occurred.
func MemberCountFromBuffer(buffer []byte) (int, int, error) {
	if len(buffer) < 4 {
		return 0, 0, errors.New("member count buffer too small")
	}
	return int(Endian.Uint16(buffer)), 4, nil
}
