package membership

import (
	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/transport"
	"github.com/go-logr/logr"
)

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
}

var DefaultConfig = Config{
	MaxDatagramLengthSend: 512,
}
