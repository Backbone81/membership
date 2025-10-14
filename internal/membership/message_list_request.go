package membership

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageListRequest asks the recipient to send its current full member list to the source. This helps in making sure
// new-joiners quickly get an overview over all members.
type MessageListRequest struct {
	// Source is the member sending this message
	Source encoding.Address
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageListRequest) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := encoding.AppendMessageTypeToBuffer(buffer, encoding.MessageTypeListRequest)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := encoding.AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	return sourceBuffer, messageTypeN + sourceN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageListRequest) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := encoding.MessageTypeFromBuffer(buffer)
	if messageType != encoding.MessageTypeListRequest {
		return 0, errors.New("invalid message type")
	}

	var sourceN int
	m.Source, sourceN, err = encoding.AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN, nil
}
