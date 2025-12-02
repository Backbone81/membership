package membership

import (
	"time"

	"github.com/backbone81/membership/internal/encryption"
	"github.com/go-logr/logr"

	"github.com/backbone81/membership/internal/encoding"
)

type Option func(config *Config)

// WithLogger sets the given logger for the list.
func WithLogger(logger logr.Logger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

// WithProtocolPeriod sets the given protocol period for the list.
func WithProtocolPeriod(protocolPeriod time.Duration) Option {
	return func(config *Config) {
		config.ProtocolPeriod = protocolPeriod
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

func WithMaxDatagramLengthSend(maxDatagramLength int) Option {
	return func(config *Config) {
		config.MaxDatagramLengthSend = maxDatagramLength
	}
}

func WithMaxDatagramLengthReceive(maxDatagramLength int) Option {
	return func(config *Config) {
		config.MaxDatagramLengthReceive = maxDatagramLength
	}
}

func WithBindAddress(bindAddress string) Option {
	return func(config *Config) {
		config.BindAddress = bindAddress
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
		config.ShutdownMemberCount = memberCount
	}
}

func WithDirectPingMemberCount(memberCount int) Option {
	return func(config *Config) {
		config.DirectPingMemberCount = memberCount
	}
}

func WithMinDirectPingMemberCount(memberCount int) Option {
	return func(config *Config) {
		config.MinDirectPingMemberCount = max(1, memberCount)
	}
}

func WithMaxDirectPingMemberCount(memberCount int) Option {
	return func(config *Config) {
		config.MaxDirectPingMemberCount = max(1, memberCount)
	}
}

func WithIndirectPingMemberCount(memberCount int) Option {
	return func(config *Config) {
		config.IndirectPingMemberCount = memberCount
	}
}

func WithEncryptionKey(key encryption.Key) Option {
	return func(config *Config) {
		config.EncryptionKeys = append(config.EncryptionKeys, key)
	}
}
