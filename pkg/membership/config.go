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
	// This callback executes under the lock of the membership list. If you call any method on the membership list
	// during that callback, you create a deadlock. If you need to call the membership list during your callback,
	// create a go routine which executes what you want to do.
	// This design decision was done, because most callback situations will not require calling the membership list.
	// That way the overhead of starting and managing a go routine is left to the user, and he only needs to pay that
	// price when necessary.
	MemberAddedCallback func(address encoding.Address)

	// MemberRemovedCallback is the callback which is triggered when a member is removed from the list.
	// This callback executes under the lock of the membership list. If you call any method on the membership list
	// during that callback, you create a deadlock. If you need to call the membership list during your callback,
	// create a go routine which executes what you want to do.
	// This design decision was done, because most callback situations will not require calling the membership list.
	// That way the overhead of starting and managing a go routine is left to the user, and he only needs to pay that
	// price when necessary.
	MemberRemovedCallback func(address encoding.Address)

	// SafetyFactor is a multiplier which describes the safety margin for disseminating gossip and declaring a suspect
	// as faulty. A factor of 1.0 wil return the minimal number of periods required in a perfect world. A factor of 2.0
	// will double the number of periods. Small values between 2.0 and 4.0 should usually be a safe value.
	SafetyFactor float64

	// ShutdownMemberCount is the number of members which are informed about this member shutting down. This helps in
	// disseminating the missing member quicker to all other members, as we do not have to rely on direct and indirect
	// pings failing against this member.
	ShutdownMemberCount int

	// IndirectPingMemberCount is the number of members to request a ping of some other member which did not respond
	// in time.
	IndirectPingMemberCount int
}

var DefaultConfig = Config{
	ProtocolPeriod:           scheduler.DefaultConfig.ProtocolPeriod,
	DirectPingTimeout:        scheduler.DefaultConfig.DirectPingTimeout,
	MaxDatagramLengthSend:    intmembership.DefaultConfig.MaxDatagramLengthSend,
	MaxDatagramLengthReceive: intmembership.DefaultConfig.MaxDatagramLengthSend,
	BindAddress:              ":3000",
	MaxSleepDuration:         scheduler.DefaultConfig.MaxSleepDuration,
	ListRequestInterval:      scheduler.DefaultConfig.ListRequestInterval,
	SafetyFactor:             intmembership.DefaultConfig.SafetyFactor,
	ShutdownMemberCount:      intmembership.DefaultConfig.ShutdownMemberCount,
	IndirectPingMemberCount:  intmembership.DefaultConfig.IndirectPingMemberCount,
}
