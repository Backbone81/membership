package encoding

import (
	"errors"
)

// AppendSequenceNumberToBuffer appends the sequence number to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendSequenceNumberToBuffer(buffer []byte, sequenceNumber uint16) ([]byte, int, error) {
	return Endian.AppendUint16(buffer, sequenceNumber), 2, nil
}

// SequenceNumberFromBuffer reads the sequence number from the provided buffer.
// Returns the sequence number, the number of bytes read and any error which occurred.
func SequenceNumberFromBuffer(buffer []byte) (uint16, int, error) {
	if len(buffer) < 2 {
		return 0, 0, errors.New("sequence number buffer too small")
	}
	return Endian.Uint16(buffer), 2, nil
}
