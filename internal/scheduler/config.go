package scheduler

import (
	"time"

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

	// DirectPingTimeout is the time to wait for a direct ping response. If there is no response within this duration,
	// we need to start indirect pings.
	DirectPingTimeout time.Duration

	// MaxSleepDuration is the maximum duration which the scheduler is allowed to sleep. Making this value bigger will
	// result in delays during shutdown. Making this value smaller will wake the scheduler more often to check for
	// a shutdown in progress which will cause more load during runtime.
	MaxSleepDuration time.Duration

	// ListRequestInterval is the time interval in which a full member list is requested from a randomly selected member.
	ListRequestInterval time.Duration
}

// DefaultConfig provides a scheduler configuration with sane defaults for most situations.
var DefaultConfig = Config{
	ProtocolPeriod:      1 * time.Second,
	DirectPingTimeout:   100 * time.Millisecond,
	MaxSleepDuration:    100 * time.Millisecond,
	ListRequestInterval: 1 * time.Minute,
}
