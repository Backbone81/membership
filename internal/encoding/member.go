package encoding

// Member is a single member which we know of.
type Member struct {
	// Address is the address the member can be reached.
	Address Address

	// State is the state the member is currently in.
	State MemberState

	// IncarnationNumber is the incarnation the member gave about itself. It is monotonically increasing each time
	// somebody suspects the member. Only the member itself is allowed to increase the incarnation.
	IncarnationNumber int

	// SuspicionPeriodCounter is the number of protocol periods the member is in a suspicion state. This is useful when
	// deciding about declaring a member as faulty when it was under suspicion long enough.
	SuspicionPeriodCounter int
}

// CompareMember orders members by address.
func CompareMember(lhs Member, rhs Member) int {
	return CompareAddress(lhs.Address, rhs.Address)
}

// AppendMemberToBuffer appends the member to the provided buffer encoded for network transfer.
// Note that the LastStateChange is not encoded to the buffer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendMemberToBuffer(buffer []byte, member Member) ([]byte, int, error) {
	addressBuffer, addressN, err := AppendAddressToBuffer(buffer, member.Address)
	if err != nil {
		return buffer, 0, err
	}

	stateBuffer, stateN, err := AppendMemberStateToBuffer(addressBuffer, member.State)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := AppendIncarnationNumberToBuffer(stateBuffer, member.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, addressN + stateN + incarnationNumberN, nil
}

// MemberFromBuffer reads the member from the provided buffer.
// Note that the LastStateChange is always the zero value.
// Returns the member, the number of bytes read and any error which occurred.
func MemberFromBuffer(buffer []byte) (Member, int, error) {
	address, addressN, err := AddressFromBuffer(buffer)
	if err != nil {
		return Member{}, 0, err
	}

	state, stateN, err := MemberStateFromBuffer(buffer[addressN:])
	if err != nil {
		return Member{}, 0, err
	}

	incarnationNumber, incarnationNumberN, err := IncarnationNumberFromBuffer(buffer[addressN+stateN:])
	if err != nil {
		return Member{}, 0, err
	}

	return Member{
		Address:           address,
		State:             state,
		IncarnationNumber: incarnationNumber,
	}, addressN + stateN + incarnationNumberN, nil
}
