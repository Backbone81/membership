package membership

import (
	"time"

	"github.com/backbone81/membership/internal/encoding"
)

// PendingDirectPing provides bookkeeping for a direct ping which is still active.
type PendingDirectPing struct {
	// Timestamp is the point in time the direct ping was initiated.
	Timestamp time.Time

	// Destination is the address which the direct ping was sent to.
	Destination encoding.Address

	// MessageDirectPing is a copy of the message which was sent for the direct ping.
	MessageDirectPing MessageDirectPing

	// MessageIndirectPing is a copy of a received indirect ping request. It is the zero value in case the direct
	// ping was not initiated in response to an indirect ping request.
	MessageIndirectPing MessageIndirectPing
}
