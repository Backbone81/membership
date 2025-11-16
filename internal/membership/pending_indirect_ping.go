package membership

import "time"

// PendingIndirectPing provides bookkeeping for an indirect ping which is still active.
type PendingIndirectPing struct {
	// Timestamp is the point in time the indirect ping was initiated.
	Timestamp time.Time

	// MessageIndirectPing is a copy of the message which was sent for an indirect ping.
	MessageIndirectPing MessageIndirectPing
}
