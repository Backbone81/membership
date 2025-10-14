package membership

import (
	"errors"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageListResponse provides a list of all known members. This message can become quite big and should always be
// transmitted over TCP and not UDP.
type MessageListResponse struct {
	Source  encoding.Address
	Members []encoding.Member
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageListResponse) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := encoding.AppendMessageTypeToBuffer(buffer, encoding.MessageTypeListResponse)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := encoding.AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	countBuffer, countN, err := encoding.AppendMemberCountToBuffer(sourceBuffer, len(m.Members))
	if err != nil {
		return buffer, 0, err
	}

	memberBuffer := countBuffer
	var memberN int
	for _, member := range m.Members {
		appendedBuffer, n, err := encoding.AppendMemberToBuffer(memberBuffer, member)
		if err != nil {
			return buffer, 0, err
		}
		memberBuffer = appendedBuffer
		memberN += n
	}

	return memberBuffer, messageTypeN + sourceN + countN + memberN, nil
}

// FromBuffer reads the message from the provided buffer.
// Returns the number of bytes read and any error which occurred.
func (m *MessageListResponse) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := encoding.MessageTypeFromBuffer(buffer)
	if messageType != encoding.MessageTypeListResponse {
		return 0, errors.New("invalid message type")
	}

	var sourceN int
	m.Source, sourceN, err = encoding.AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	count, countN, err := encoding.MemberCountFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	if cap(m.Members) < count {
		m.Members = make([]encoding.Member, 0, count)
	}
	if len(m.Members) != 0 {
		// We reset the members slice after the check for capacity. In case the capacity is not enough, we save the
		// reset because the length after the capacity increase already is 0.
		m.Members = m.Members[:0]
	}

	var memberN int
	for i := 0; i < count; i++ {
		member, n, err := encoding.MemberFromBuffer(buffer[messageTypeN+sourceN+countN+memberN:])
		if err != nil {
			return 0, err
		}
		memberN += n
		m.Members = append(m.Members, member)
	}

	return messageTypeN + sourceN + countN + memberN, nil
}
