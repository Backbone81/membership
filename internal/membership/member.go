package membership

import "time"

type MemberState int

const (
	MemberStateAlive MemberState = iota
	MemberStateSuspect
	MemberStateFaulty
)

type Member struct {
	Address           Address
	State             MemberState
	LastStateChange   time.Time
	IncarnationNumber int
}

func AppendMemberToBuffer(buffer []byte, member Member) ([]byte, int, error) {
	addressBuffer, addressN, err := AppendAddressToBuffer(buffer, member.Address)
	if err != nil {
		return buffer, 0, err
	}

	stateBuffer := append(addressBuffer, byte(member.State))

	incarnationNumberBuffer := Endian.AppendUint16(stateBuffer, uint16(member.IncarnationNumber))

	return incarnationNumberBuffer, addressN + 1 + 2, nil
}

func MemberFromBuffer(buffer []byte) (Member, int, error) {
	address, addressN, err := AddressFromBuffer(buffer)
	if err != nil {
		return Member{}, 0, err
	}

	memberState := MemberState(buffer[addressN])

	incarnationNumber := int(Endian.Uint16(buffer[addressN+1:]))
	return Member{
		Address:           address,
		State:             memberState,
		IncarnationNumber: incarnationNumber,
	}, 0, nil
}
