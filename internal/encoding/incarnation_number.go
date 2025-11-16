package encoding

import (
	"errors"
)

// AppendIncarnationNumberToBuffer appends the incarnation number to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendIncarnationNumberToBuffer(buffer []byte, incarnationNumber uint16) ([]byte, int, error) {
	return Endian.AppendUint16(buffer, incarnationNumber), 2, nil
}

// IncarnationNumberFromBuffer reads the incarnation number from the provided buffer.
// Returns the sequence number, the number of bytes read and any error which occurred.
func IncarnationNumberFromBuffer(buffer []byte) (uint16, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("incarnation buffer too small")
	}
	return Endian.Uint16(buffer), 2, nil
}
