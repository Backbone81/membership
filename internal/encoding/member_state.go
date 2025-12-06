package encoding

import "errors"

// MemberState describes the state the member is in.
type MemberState int

const (
	MemberStateAlive MemberState = iota + 1 // We start with a placeholder member state to detect missing states.
	MemberStateSuspect
	MemberStateFaulty
)

// AppendMemberStateToBuffer appends the member state to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendMemberStateToBuffer(buffer []byte, memberState MemberState) ([]byte, int, error) {
	return append(buffer, byte(memberState)), 1, nil
}

// MemberStateFromBuffer reads the member state from the provided buffer.
// Returns the member state, the number of bytes read and any error which occurred.
func MemberStateFromBuffer(buffer []byte) (MemberState, int, error) {
	if len(buffer) < 1 {
		return 0, 0, errors.New("member state buffer too small")
	}
	return MemberState(buffer[0]), 1, nil
}
