package membership

import (
	"time"

	"github.com/backbone81/membership/internal/encoding"
)

// Member is a single member which we know of.
type Member struct {
	// Address is the address the member can be reached.
	Address encoding.Address

	// State is the state the member is currently in.
	State MemberState

	// LastStateChange is the point in time the state of the member last changed.
	LastStateChange time.Time

	// IncarnationNumber is the incarnation the member gave about itself. It is monotonically increasing each time
	// somebody suspects the member. Only the member itself is allowed to increase the incarnation.
	IncarnationNumber int
}

// AppendMemberToBuffer appends the member to the provided buffer encoded for network transfer.
// Note that the LastStateChange is not encoded to the buffer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendMemberToBuffer(buffer []byte, member Member) ([]byte, int, error) {
	addressBuffer, addressN, err := encoding.AppendAddressToBuffer(buffer, member.Address)
	if err != nil {
		return buffer, 0, err
	}

	stateBuffer, stateN, err := AppendMemberStateToBuffer(addressBuffer, member.State)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := encoding.AppendIncarnationNumberToBuffer(stateBuffer, member.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, addressN + stateN + incarnationNumberN, nil
}

// MemberFromBuffer reads the member from the provided buffer.
// Note that the LastStateChange is always the zero value.
// Returns the member, the number of bytes read and any error which occurred.
func MemberFromBuffer(buffer []byte) (Member, int, error) {
	address, addressN, err := encoding.AddressFromBuffer(buffer)
	if err != nil {
		return Member{}, 0, err
	}

	state, stateN, err := MemberStateFromBuffer(buffer[addressN:])
	if err != nil {
		return Member{}, 0, err
	}

	incarnationNumber, incarnationNumberN, err := encoding.IncarnationNumberFromBuffer(buffer[addressN+stateN:])
	if err != nil {
		return Member{}, 0, err
	}

	return Member{
		Address:           address,
		State:             state,
		IncarnationNumber: incarnationNumber,
	}, addressN + stateN + incarnationNumberN, nil
}
