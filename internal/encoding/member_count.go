package encoding

import (
	"errors"
	"math"
)

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
