package encoding

import (
	"errors"
	"math"
)

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
		return 0, 0, errors.New("sequence number buffer too small")
	}
	return int(Endian.Uint16(buffer)), 2, nil
}
