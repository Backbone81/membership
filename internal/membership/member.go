package membership

type MemberState int

const (
	MemberStateAlive MemberState = iota
	MemberStateSuspect
	MemberStateFailed
)

type Member struct {
	Endpoint Endpoint
	State    MemberState
}
