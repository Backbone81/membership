package membership

type MemberState int

const (
	MemberStateAlive MemberState = iota
	MemberStateSuspect
	MemberStateFaulty
)

type Member struct {
	Endpoint          Endpoint
	State             MemberState
	IncarnationNumber int
}
