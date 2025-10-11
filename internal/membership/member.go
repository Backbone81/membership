package membership

import "time"

type MemberState int

const (
	MemberStateAlive MemberState = iota
	MemberStateSuspect
	MemberStateFaulty
)

type Member struct {
	Endpoint          Endpoint
	State             MemberState
	LastStateChange   time.Time
	IncarnationNumber int
}

func AppendMemberToBuffer(buffer []byte, member Member) ([]byte, int, error) {
	endpointBuffer, endpointN, err := AppendEndpointToBuffer(buffer, member.Endpoint)
	if err != nil {
		return buffer, 0, err
	}

	stateBuffer := append(endpointBuffer, byte(member.State))

	incarnationNumberBuffer := Endian.AppendUint16(stateBuffer, uint16(member.IncarnationNumber))

	return incarnationNumberBuffer, endpointN + 1 + 2, nil
}

func MemberFromBuffer(buffer []byte) (Member, int, error) {
	endpoint, endpointN, err := EndpointFromBuffer(buffer)
	if err != nil {
		return Member{}, 0, err
	}

	memberState := MemberState(buffer[endpointN])

	incarnationNumber := int(Endian.Uint16(buffer[endpointN+1:]))
	return Member{
		Endpoint:          endpoint,
		State:             memberState,
		IncarnationNumber: incarnationNumber,
	}, 0, nil
}
