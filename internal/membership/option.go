package membership

import (
	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/go-logr/logr"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/transport"
)

// Option is the function signature for all list options to implement.
type Option func(config *Config)

// WithLogger sets the given logger for the list.
func WithLogger(logger logr.Logger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithBootstrapMember(address encoding.Address) Option {
	return func(config *Config) {
		config.BootstrapMembers = append(config.BootstrapMembers, address)
	}
}

func WithBootstrapMembers(addresses []encoding.Address) Option {
	return func(config *Config) {
		config.BootstrapMembers = append(config.BootstrapMembers, addresses...)
	}
}

func WithAdvertisedAddress(address encoding.Address) Option {
	return func(config *Config) {
		config.AdvertisedAddress = address
	}
}

func WithUDPClient(transport transport.Transport) Option {
	return func(config *Config) {
		config.UDPClient = transport
	}
}

func WithTCPClient(transport transport.Transport) Option {
	return func(config *Config) {
		config.TCPClient = transport
	}
}

func WithMaxDatagramLengthSend(maxDatagramLength int) Option {
	return func(config *Config) {
		config.MaxDatagramLengthSend = maxDatagramLength
	}
}

func WithMemberAddedCallback(memberAddedCallback func(address encoding.Address)) Option {
	return func(config *Config) {
		config.MemberAddedCallback = memberAddedCallback
	}
}

func WithMemberRemovedCallback(memberRemovedCallback func(address encoding.Address)) Option {
	return func(config *Config) {
		config.MemberRemovedCallback = memberRemovedCallback
	}
}

func WithSafetyFactor(safetyFactor float64) Option {
	return func(config *Config) {
		config.SafetyFactor = safetyFactor
	}
}

func WithShutdownMemberCount(memberCount int) Option {
	return func(config *Config) {
		config.ShutdownMemberCount = max(1, min(memberCount, 64))
	}
}

func WithDirectPingMemberCount(memberCount int) Option {
	return func(config *Config) {
		config.DirectPingMemberCount = max(1, min(memberCount, 64))
	}
}

func WithIndirectPingMemberCount(memberCount int) Option {
	return func(config *Config) {
		config.IndirectPingMemberCount = max(1, min(memberCount, 64))
	}
}

func WithRoundTripTimeTracker(rttTracker *roundtriptime.Tracker) Option {
	return func(config *Config) {
		config.RoundTripTimeTracker = rttTracker
	}
}
