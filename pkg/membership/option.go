package membership

import (
	"time"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/go-logr/logr"
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

// WithDirectPingTimeout sets the given direct ping timeout for the list.
func WithDirectPingTimeout(directPingTimeout time.Duration) Option {
	return func(config *Config) {
		config.DirectPingTimeout = directPingTimeout
	}
}

func WithBootstrapMember(address encoding.Address) Option {
	return func(config *Config) {
		config.BootstrapMembers = append(config.BootstrapMembers, address)
	}
}

func WithBootstrapMembers(addresses []encoding.Address) Option {
	return func(config *Config) {
		for _, address := range addresses {
			config.BootstrapMembers = append(config.BootstrapMembers, address)
		}
	}
}

func WithAdvertisedAddress(address encoding.Address) Option {
	return func(config *Config) {
		config.AdvertisedAddress = address
	}
}

func WithMaxDatagramLength(maxDatagramLength int) Option {
	return func(config *Config) {
		config.MaxDatagramLength = maxDatagramLength
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
