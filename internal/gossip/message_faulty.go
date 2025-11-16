package gossip

import (
	"errors"
	"fmt"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageFaulty declares the destination as being faulty by the source.
// This is the `Confirm` message of SWIM chapter 4.2. Suspicion Mechanism: Reducing the Frequency of False Positives.
type MessageFaulty struct {
	// Source is the member which declared the destination as faulty.
	Source encoding.Address

	// Destination is the member which was declared as faulty by source.
	Destination encoding.Address

	// IncarnationNumber is the incarnation which source saw and based its decision on. This helps in identifying
	// outdated messages.
	IncarnationNumber uint16
}

func (m *MessageFaulty) String() string {
	return fmt.Sprintf("Faulty %s (by %s, incarnation %d)", m.Destination, m.Source, m.IncarnationNumber)
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageFaulty) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := encoding.AppendMessageTypeToBuffer(buffer, encoding.MessageTypeFaulty)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := encoding.AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := encoding.AppendAddressToBuffer(sourceBuffer, m.Destination)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := encoding.AppendIncarnationNumberToBuffer(destinationBuffer, m.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageFaulty) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := encoding.MessageTypeFromBuffer(buffer)
	if err != nil {
		return 0, err
	}
	if messageType != encoding.MessageTypeFaulty {
		return 0, errors.New("invalid message type")
	}

	var sourceN, destinationN, incarnationNumberN int
	m.Source, sourceN, err = encoding.AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = encoding.AddressFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	m.IncarnationNumber, incarnationNumberN, err = encoding.IncarnationNumberFromBuffer(buffer[messageTypeN+sourceN+destinationN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

// GetAddress returns the address which is relevant for this message. Needed for the gossip queue to check for equality.
func (m *MessageFaulty) GetAddress() encoding.Address {
	return m.Destination
}

func (m *MessageFaulty) GetType() encoding.MessageType {
	return encoding.MessageTypeFaulty
}

func (m *MessageFaulty) GetIncarnationNumber() uint16 {
	return m.IncarnationNumber
}
