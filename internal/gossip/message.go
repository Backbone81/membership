package gossip

import (
	"github.com/backbone81/membership/internal/encoding"
	"github.com/backbone81/membership/internal/utility"
)

// ShouldReplaceExistingWithNew reports if the new message has higher precedence as the existing message. A new message
// with higher precedence should replace the existing one, while same or lower precedence should be dropped.
func ShouldReplaceExistingWithNew(existingMsg encoding.Message, newMsg encoding.Message) bool {
	if utility.IncarnationLessThan(newMsg.IncarnationNumber, existingMsg.IncarnationNumber) {
		// No need to overwrite when the incarnation number is lower.
		return false
	}

	if newMsg.IncarnationNumber == existingMsg.IncarnationNumber &&
		(newMsg.Type == encoding.MessageTypeAlive ||
			newMsg.Type == encoding.MessageTypeSuspect && existingMsg.Type != encoding.MessageTypeAlive ||
			newMsg.Type == encoding.MessageTypeFaulty && existingMsg.Type != encoding.MessageTypeAlive && existingMsg.Type != encoding.MessageTypeSuspect) {
		// No need to overwrite with the same incarnation number and the wrong priorities.
		return false
	}
	return true
}
