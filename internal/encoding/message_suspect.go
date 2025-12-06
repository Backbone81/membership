//nolint:dupl
package encoding

import (
	"errors"
	"fmt"
)

// MessageSuspect declares the destination as being suspected for failure by the source.
// This is the `Suspect` message of SWIM chapter 4.2. Suspicion Mechanism: Reducing the Frequency of False Positives.
type MessageSuspect struct {
	// Source is the member which declared the destination as suspect.
	Source Address

	// Destination is the member which was declared as suspect by source.
	Destination Address

	// IncarnationNumber is the incarnation which source saw and based its decision on. This helps in identifying
	// outdated messages.
	IncarnationNumber uint16
}

// ToMessage converts the specific message into the general purpose message.
func (m MessageSuspect) ToMessage() Message {
	return Message{
		Type:              MessageTypeSuspect,
		Source:            m.Source,
		Destination:       m.Destination,
		IncarnationNumber: m.IncarnationNumber,
	}
}

func (m MessageSuspect) String() string {
	return fmt.Sprintf("Suspect %s (by %s, incarnation %d)", m.Destination, m.Source, m.IncarnationNumber)
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m MessageSuspect) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeSuspect)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := AppendAddressToBuffer(sourceBuffer, m.Destination)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := AppendIncarnationNumberToBuffer(destinationBuffer, m.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageSuspect) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if err != nil {
		return 0, err
	}
	if messageType != MessageTypeSuspect {
		return 0, errors.New("invalid message type")
	}

	var sourceN, destinationN, incarnationNumberN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = AddressFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	m.IncarnationNumber, incarnationNumberN, err = IncarnationNumberFromBuffer(buffer[messageTypeN+sourceN+destinationN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

// GetAddress returns the address which is relevant for this message. Needed for the gossip queue to check for equality.
func (m *MessageSuspect) GetAddress() Address {
	return m.Destination
}

func (m *MessageSuspect) GetType() MessageType {
	return MessageTypeSuspect
}

func (m *MessageSuspect) GetIncarnationNumber() uint16 {
	return m.IncarnationNumber
}
