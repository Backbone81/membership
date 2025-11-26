package encoding

import (
	"errors"
)

// MessageListRequest asks the recipient to send its current full member list to the source. This helps in making sure
// new-joiners quickly get an overview over all members.
type MessageListRequest struct {
	// Source is the member sending this message
	Source Address
}

// ToMessage converts the specific message into the general purpose message.
func (m *MessageListRequest) ToMessage() Message {
	return Message{
		Type:   MessageTypeListRequest,
		Source: m.Source,
	}
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageListRequest) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeListRequest)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	return sourceBuffer, messageTypeN + sourceN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageListRequest) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeListRequest {
		return 0, errors.New("invalid message type")
	}

	var sourceN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN, nil
}
