package membership

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageIndirectAck is a response message sent back in response to receiving a MessageIndirectPing and receiving
// a MessageDirectAck from the destination. The message is identical to MessageDirectAck but allows us to differentiate
// those messages when calculating round trip times later.
type MessageIndirectAck struct {
	// Source is the member sending the indirect ack.
	Source encoding.Address

	// SequenceNumber is the same sequence which was initially sent with the indirect ping. This enables us to ignore
	// indirect acks which arrive late.
	SequenceNumber int
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageIndirectAck) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := encoding.AppendMessageTypeToBuffer(buffer, encoding.MessageTypeIndirectAck)
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
func (m *MessageIndirectAck) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := encoding.MessageTypeFromBuffer(buffer)
	if messageType != encoding.MessageTypeIndirectAck {
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
