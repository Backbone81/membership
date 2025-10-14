package encoding

import (
	"errors"
	"math"
)

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
