package scheduler

import "time"

// Target is the interface which the implementation of the membership algorithm must implement to be driven
// by the scheduler.
type Target interface {
	// DirectPing is the start of the protocol period which is executing the direct ping.
	DirectPing() error

	// IndirectPing is after some time elapsed and indirect pings need to be executed.
	IndirectPing() error

	// EndOfProtocolPeriod is the end of the protocol period where suspects and faulty members need to be declared.
	EndOfProtocolPeriod() error

	// RequestList fetches the full member list from a randomly chosen member.
	RequestList() error

	// ExpectedRoundTripTime returns the round trip time the target is expecting.
	ExpectedRoundTripTime() time.Duration
}
