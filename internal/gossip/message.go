package gossip

import (
	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/utility"
)

// ShouldReplaceExistingWithNew reports if the new message has higher precedence as the existing message. A new message
// with higher precedence should replace the existing one, while same or lower precedence should be dropped.
func ShouldReplaceExistingWithNew(existing encoding.Message, new encoding.Message) bool {
	if utility.IncarnationLessThan(new.IncarnationNumber, existing.IncarnationNumber) {
		// No need to overwrite when the incarnation number is lower.
		return false
	}

	if new.IncarnationNumber == existing.IncarnationNumber &&
		(new.Type == encoding.MessageTypeAlive ||
			new.Type == encoding.MessageTypeSuspect && existing.Type != encoding.MessageTypeAlive ||
			new.Type == encoding.MessageTypeFaulty && existing.Type != encoding.MessageTypeAlive && existing.Type != encoding.MessageTypeSuspect) {
		// No need to overwrite with the same incarnation number and the wrong priorities.
		return false
	}
	return true
}
