package membership

import (
	"github.com/backbone81/membership/internal/encoding"
	"github.com/go-logr/logr"
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

func WithUDPClient(transport Transport) Option {
	return func(config *Config) {
		config.UDPClient = transport
	}
}

func WithTCPClient(transport Transport) Option {
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
