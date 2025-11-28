package encoding

import (
	"errors"
	"fmt"
)

// MessageListResponse provides a list of all known members. This message can become quite big and should always be
// transmitted over TCP and not UDP.
type MessageListResponse struct {
	Source  Address
	Members []Member
}

func (m MessageListResponse) String() string {
	return fmt.Sprintf("ListResponse (by %s)", m.Source)
}

// ToMessage converts the specific message into the general purpose message.
func (m MessageListResponse) ToMessage() Message {
	return Message{
		Type:    MessageTypeListResponse,
		Source:  m.Source,
		Members: m.Members,
	}
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m MessageListResponse) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeListResponse)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	countBuffer, countN, err := AppendMemberCountToBuffer(sourceBuffer, len(m.Members))
	if err != nil {
		return buffer, 0, err
	}

	memberBuffer := countBuffer
	var memberN int
	for _, member := range m.Members {
		appendedBuffer, n, err := AppendMemberToBuffer(memberBuffer, member)
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
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeListResponse {
		return 0, errors.New("invalid message type")
	}

	var sourceN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	count, countN, err := MemberCountFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	if cap(m.Members) < count {
		m.Members = make([]Member, 0, count)
	}
	if len(m.Members) != 0 {
		// We reset the members slice after the check for capacity. In case the capacity is not enough, we save the
		// reset because the length after the capacity increase already is 0.
		m.Members = m.Members[:0]
	}

	var memberN int
	for range count {
		member, n, err := MemberFromBuffer(buffer[messageTypeN+sourceN+countN+memberN:])
		if err != nil {
			return 0, err
		}
		memberN += n
		m.Members = append(m.Members, member)
	}

	return messageTypeN + sourceN + countN + memberN, nil
}
