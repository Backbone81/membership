package encoding

import (
	"errors"
	"fmt"
)

// MessageAlive declares the destination as being alive by the source.
// This is the `Alive` message of SWIM chapter 4.2. Suspicion Mechanism: Reducing the Frequency of False Positives.
type MessageAlive struct {
	// Destination is the member declaring itself as alive.
	Destination Address

	// IncarnationNumber is the incarnation to distinguish an outdated alive message from a new one.
	IncarnationNumber uint16
}

// ToMessage converts the specific message into the general purpose message.
func (m *MessageAlive) ToMessage() Message {
	return Message{
		Type:              MessageTypeAlive,
		Destination:       m.Destination,
		IncarnationNumber: m.IncarnationNumber,
	}
}

func (m *MessageAlive) String() string {
	return fmt.Sprintf("Alive %s (incarnation %d)", m.Destination, m.IncarnationNumber)
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageAlive) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeAlive)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Destination)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := AppendIncarnationNumberToBuffer(sourceBuffer, m.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, messageTypeN + sourceN + incarnationNumberN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageAlive) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if err != nil {
		return 0, err
	}
	if messageType != MessageTypeAlive {
		return 0, errors.New("invalid message type")
	}

	var sourceN, incarnationNumberN int
	m.Destination, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.IncarnationNumber, incarnationNumberN, err = IncarnationNumberFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + incarnationNumberN, nil
}

// GetAddress returns the address which is relevant for this message. Needed for the gossip queue to check for equality.
func (m *MessageAlive) GetAddress() Address {
	return m.Destination
}

func (m *MessageAlive) GetType() MessageType {
	return MessageTypeAlive
}

func (m *MessageAlive) GetIncarnationNumber() uint16 {
	return m.IncarnationNumber
}
