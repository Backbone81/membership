package membership

import (
	"time"

	"github.com/backbone81/membership/internal/encoding"
	"github.com/go-logr/logr"
)

type Config struct {
	// Logger is the Logger to use for outputting status information.
	Logger logr.Logger

	// ProtocolPeriod is the time for a full cycle of direct ping followed by indirect pings. If there is no response
	// from the target member within that time, we have to assume the member to have failed.
	// Note that the protocol period must be at least three times the round-trip time.
	// TODO: This setting should not be necessary, because we want to be timing-independent.
	ProtocolPeriod time.Duration

	// DirectPingTimeout is the time to wait for a direct ping response. If there is no response within this duration,
	// we need to start indirect pings.
	// TODO: This setting should not be necessary, because we want to be timing-independent.
	DirectPingTimeout time.Duration

	// BootstrapMembers is a list of members which are contacted to join the members. This list does not have to be
	// complete. One or two known members are enough.
	BootstrapMembers []encoding.Address

	// AdvertisedAddress is the address for contacting this member.
	AdvertisedAddress encoding.Address

	// UDPClient is the transport for sending unreliable UDP network messages.
	UDPClient Transport

	// TCPClient is the transport for sending reliable TCP network messages.
	TCPClient Transport

	// MaxDatagramLengthSend is the maximum length in bytes we should not exceed for sending UDP network messages.
	MaxDatagramLengthSend int

	// MemberAddedCallback is the callback which is triggered when a new member is added to the list.
	MemberAddedCallback func(address encoding.Address)

	// MemberRemovedCallback is the callback which is triggered when a member is removed from the list.
	MemberRemovedCallback func(address encoding.Address)
}

var DefaultConfig = Config{
	DirectPingTimeout:     100 * time.Millisecond,
	ProtocolPeriod:        1 * time.Second,
	MaxDatagramLengthSend: 512,
}
