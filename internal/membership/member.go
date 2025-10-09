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
