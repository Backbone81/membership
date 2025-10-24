package membership

import (
	"time"

	"github.com/backbone81/membership/internal/encoding"
	intmembership "github.com/backbone81/membership/internal/membership"
	"github.com/backbone81/membership/internal/scheduler"
	"github.com/go-logr/logr"
)

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

	// BootstrapMembers is a list of members which are contacted to join the members. This list does not have to be
	// complete. One or two known members are enough.
	BootstrapMembers []encoding.Address

	// AdvertisedAddress is the address for contacting this member.
	AdvertisedAddress encoding.Address

	// MaxDatagramLengthSend is the maximum length in bytes we should not exceed for sending UDP network messages.
	MaxDatagramLengthSend int

	// MaxDatagramLengthReceive is the maximum length in bytes we should not exceed for receiving UDP network messages.
	MaxDatagramLengthReceive int

	BindAddress string

	MaxSleepDuration time.Duration

	ListRequestInterval time.Duration

	// MemberAddedCallback is the callback which is triggered when a new member is added to the list.
	MemberAddedCallback func(address encoding.Address)

	// MemberRemovedCallback is the callback which is triggered when a member is removed from the list.
	MemberRemovedCallback func(address encoding.Address)
}

var DefaultConfig = Config{
	ProtocolPeriod:           scheduler.DefaultConfig.ProtocolPeriod,
	DirectPingTimeout:        scheduler.DefaultConfig.DirectPingTimeout,
	MaxDatagramLengthSend:    intmembership.DefaultConfig.MaxDatagramLengthSend,
	MaxDatagramLengthReceive: intmembership.DefaultConfig.MaxDatagramLengthSend,
	BindAddress:              ":3000",
	MaxSleepDuration:         scheduler.DefaultConfig.MaxSleepDuration,
	ListRequestInterval:      scheduler.DefaultConfig.ListRequestInterval,
}
