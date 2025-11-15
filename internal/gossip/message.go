package gossip

import "github.com/backbone81/membership/internal/encoding"

// Message is the interface all gossip network messages need to implement.
type Message interface {
	AppendToBuffer(buffer []byte) ([]byte, int, error)
	FromBuffer(buffer []byte) (int, error)
	GetAddress() encoding.Address
	GetType() encoding.MessageType
	GetIncarnationNumber() int
	String() string
}

// ShouldReplaceExistingWithNew reports if the new message has higher precedence as the existing message. A new message
// with higher precedence should replace the existing one, while same or lower precedence should be dropped.
func ShouldReplaceExistingWithNew(existing Message, new Message) bool {
	newIncarnationNumber := new.GetIncarnationNumber()
	existingIncarnationNumber := existing.GetIncarnationNumber()
	if newIncarnationNumber < existingIncarnationNumber {
		// No need to overwrite when the incarnation number is lower.
		return false
	}

	newMessageType := new.GetType()
	existingMessageType := existing.GetType()
	if newIncarnationNumber == existingIncarnationNumber &&
		(newMessageType == encoding.MessageTypeAlive ||
			newMessageType == encoding.MessageTypeSuspect && existingMessageType != encoding.MessageTypeAlive ||
			newMessageType == encoding.MessageTypeFaulty && existingMessageType != encoding.MessageTypeAlive && existingMessageType != encoding.MessageTypeSuspect) {
		// No need to overwrite with the same incarnation number and the wrong priorities.
		return false
	}
	return true
}
