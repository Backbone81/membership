package scheduler

import (
	"time"

	"github.com/go-logr/logr"
)

// Option is the function signature for all scheduler options to implement.
type Option func(config *Config)

// WithLogger sets the given logger for the scheduler.
func WithLogger(logger logr.Logger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

// WithProtocolPeriod sets the given protocol period for the scheduler.
func WithProtocolPeriod(protocolPeriod time.Duration) Option {
	return func(config *Config) {
		config.ProtocolPeriod = protocolPeriod
	}
}

// WithDirectPingTimeout sets the given direct ping timeout for the scheduler.
func WithDirectPingTimeout(directPingTimeout time.Duration) Option {
	return func(config *Config) {
		config.DirectPingTimeout = directPingTimeout
	}
}

// WithMaxSleepDuration sets the given max sleep duration for the scheduler.
func WithMaxSleepDuration(maxSleepDuration time.Duration) Option {
	return func(config *Config) {
		config.MaxSleepDuration = maxSleepDuration
	}
}

// WithListRequestInterval sets the given list request interval for the scheduler.
func WithListRequestInterval(listRequestInterval time.Duration) Option {
	return func(config *Config) {
		config.ListRequestInterval = listRequestInterval
	}
}
