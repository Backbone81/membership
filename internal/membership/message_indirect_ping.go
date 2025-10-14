package membership

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageIndirectPing is a request of the recipient to send a MessageDirectPing to the destination.
// This is the `ping-req` message of SWIM chapter 3.1. SWIM Failure Detector.
type MessageIndirectPing struct {
	// Source is the member requesting the indirect ping.
	Source encoding.Address

	// Destination is the member which should be directly pinged by the member receiving this message.
	Destination encoding.Address

	// SequenceNumber is the sequence which should be unique for every direct ping. This sequence number should
	// match the sequence number of the previous direct ping to allow correlation of direct and indirect pings.
	SequenceNumber int
}

// IsZero reports if this message is the zero value.
func (m *MessageIndirectPing) IsZero() bool {
	return m.SequenceNumber == 0 && m.Source.IsZero() && m.Destination.IsZero()
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageIndirectPing) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := encoding.AppendMessageTypeToBuffer(buffer, encoding.MessageTypeIndirectPing)
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

	sequenceNumberBuffer, sequenceNumberN, err := encoding.AppendSequenceNumberToBuffer(destinationBuffer, m.SequenceNumber)
	if err != nil {
		return buffer, 0, err
	}

	return sequenceNumberBuffer, messageTypeN + sourceN + destinationN + sequenceNumberN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageIndirectPing) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := encoding.MessageTypeFromBuffer(buffer)
	if messageType != encoding.MessageTypeIndirectPing {
		return 0, errors.New("invalid message type")
	}

	var sourceN, destinationN, sequenceNumberN int
	m.Source, sourceN, err = encoding.AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = encoding.AddressFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	m.SequenceNumber, sequenceNumberN, err = encoding.SequenceNumberFromBuffer(buffer[messageTypeN+sourceN+destinationN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + destinationN + sequenceNumberN, nil
}
