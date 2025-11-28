package membership

import (
	"github.com/backbone81/membership/internal/roundtriptime"
	"github.com/go-logr/logr"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/transport"
)

// Config provides the configuration for the membership list.
type Config struct {
	// Logger is the Logger to use for outputting status information.
	Logger logr.Logger

	// BootstrapMembers is a list of members which are contacted to join the members. This list does not have to be
	// complete. One or two known members are enough.
	BootstrapMembers []encoding.Address

	// AdvertisedAddress is the address for contacting this member.
	AdvertisedAddress encoding.Address

	// UDPClient is the transport for sending unreliable UDP network messages.
	UDPClient transport.Transport

	// TCPClient is the transport for sending reliable TCP network messages.
	TCPClient transport.Transport

	// MaxDatagramLengthSend is the maximum length in bytes we should not exceed for sending UDP network messages.
	MaxDatagramLengthSend int

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

	// DirectPingMemberCount is the number of members to ping directly.
	DirectPingMemberCount int

	// IndirectPingMemberCount is the number of members to request a ping of some other member which did not respond
	// in time.
	IndirectPingMemberCount int

	// RoundTripTimeTracker is the roundtrip time tracker which the membership list records the measured network round trips to.
	RoundTripTimeTracker *roundtriptime.Tracker

	// PendingPingPreAllocation is the number of pending pings are pre-allocated to reduce allocations later. This
	// option is primarily used for benchmarks to avoid memory allocations. There should be no real need to ever
	// set this for real use cases.
	PendingPingPreAllocation int
}

// DefaultConfig provides a default configuration which should work for most use-cases.
var DefaultConfig = Config{
	MaxDatagramLengthSend:    512,
	SafetyFactor:             3,
	ShutdownMemberCount:      3,
	DirectPingMemberCount:    1,
	IndirectPingMemberCount:  3,
	PendingPingPreAllocation: 16,
}
