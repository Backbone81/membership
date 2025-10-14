package membership

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageDirectPing is a ping message directly sent to the recipient.
// This is the `ping` message of SWIM chapter 3.1. SWIM Failure Detector.
type MessageDirectPing struct {
	// Source is the member sending the ping.
	Source encoding.Address

	// SequenceNumber is the sequence we expect to get back in the direct ack. The sequence number should be different
	// for every direct ping we send out.
	SequenceNumber int
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageDirectPing) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := encoding.AppendMessageTypeToBuffer(buffer, encoding.MessageTypeDirectPing)
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
func (m *MessageDirectPing) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := encoding.MessageTypeFromBuffer(buffer)
	if messageType != encoding.MessageTypeDirectPing {
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
