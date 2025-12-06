//nolint:dupl
package encoding

import (
	"errors"
	"fmt"
)

// MessageIndirectAck is a response message sent back in response to receiving a MessageIndirectPing and receiving
// a MessageDirectAck from the destination. The message is identical to MessageDirectAck but allows us to differentiate
// those messages when calculating round trip times later.
type MessageIndirectAck struct {
	// Source is the member sending the indirect ack.
	Source Address

	// SequenceNumber is the same sequence which was initially sent with the indirect ping. This enables us to ignore
	// indirect acks which arrive late.
	SequenceNumber uint16
}

func (m MessageIndirectAck) String() string {
	return fmt.Sprintf("IndirectAck (by %s, sequence %d)", m.Source, m.SequenceNumber)
}

// ToMessage converts the specific message into the general purpose message.
func (m MessageIndirectAck) ToMessage() Message {
	return Message{
		Type:           MessageTypeIndirectAck,
		Source:         m.Source,
		SequenceNumber: m.SequenceNumber,
	}
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m MessageIndirectAck) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeIndirectAck)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	sequenceNumberBuffer, sequenceNumberN, err := AppendSequenceNumberToBuffer(sourceBuffer, m.SequenceNumber)
	if err != nil {
		return buffer, 0, err
	}

	return sequenceNumberBuffer, messageTypeN + sourceN + sequenceNumberN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageIndirectAck) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if err != nil {
		return 0, err
	}
	if messageType != MessageTypeIndirectAck {
		return 0, errors.New("invalid message type")
	}

	var sourceN, sequenceNumberN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.SequenceNumber, sequenceNumberN, err = SequenceNumberFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + sequenceNumberN, nil
}
