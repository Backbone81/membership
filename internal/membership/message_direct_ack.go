package membership

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageDirectAck is a response message sent back in answer to receiving a MessageDirectPing.
// This is the `ack` message of SWIM chapter 3.1. SWIM Failure Detector.
type MessageDirectAck struct {
	// Source is the member sending the direct ack. This is the same member which received the direct ping before.
	Source encoding.Address

	// SequenceNumber is the same sequence which the member received with the direct ping. This makes sure that direct
	// acks which arrive too late are ignored.
	SequenceNumber uint16
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageDirectAck) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := encoding.AppendMessageTypeToBuffer(buffer, encoding.MessageTypeDirectAck)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := encoding.AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	sequenceNumberBuffer, sequenceNumberN, err := encoding.AppendSequenceNumberToBuffer(sourceBuffer, m.SequenceNumber)
	if err != nil {
		return buffer, 0, err
	}

	return sequenceNumberBuffer, messageTypeN + sourceN + sequenceNumberN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageDirectAck) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := encoding.MessageTypeFromBuffer(buffer)
	if messageType != encoding.MessageTypeDirectAck {
		return 0, errors.New("invalid message type")
	}

	var sourceN, sequenceNumberN int
	m.Source, sourceN, err = encoding.AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.SequenceNumber, sequenceNumberN, err = encoding.SequenceNumberFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + sequenceNumberN, nil
}
