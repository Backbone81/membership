package scheduler

import (
	"time"

	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/go-logr/logr"
)

// Config is the configuration the scheduler is using.
type Config struct {
	// Logger is the Logger to use for outputting status information.
	Logger logr.Logger

	// ProtocolPeriod is the time for a full cycle of direct ping followed by indirect pings. If there is no response
	// from the target member within that time, we have to assume the member to have failed.
	// Note that the protocol period must be at least three times the round-trip time.
	ProtocolPeriod time.Duration

	// MaxSleepDuration is the maximum duration which the scheduler is allowed to sleep. Making this value bigger will
	// result in delays during shutdown. Making this value smaller will wake the scheduler more often to check for
	// a shutdown in progress which will cause more load during runtime.
	MaxSleepDuration time.Duration

	// ListRequestInterval is the time interval in which a full member list is requested from a randomly selected member.
	ListRequestInterval time.Duration

	// RoundTripTimeTracker is the roundtrip time tracker which the membership list records the measured network round
	// trips to.
	RoundTripTimeTracker *roundtriptime.Tracker
}

// DefaultConfig provides a scheduler configuration with sane defaults for most situations.
var DefaultConfig = Config{
	ProtocolPeriod:      1 * time.Second,
	MaxSleepDuration:    100 * time.Millisecond,
	ListRequestInterval: 1 * time.Minute,
}
