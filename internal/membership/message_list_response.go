package membership

import "errors"

// MessageListResponse provides a list of all known members. This message can become quite big and should always be
// transmitted over TCP and not UDP.
type MessageListResponse struct {
	Source  Address
	Members []Member
}

// AppendToBuffer appends the message to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func (m *MessageListResponse) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeListResponse)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	countBuffer := Endian.AppendUint32(sourceBuffer, uint32(len(m.Members)))

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

	return memberBuffer, messageTypeN + sourceN + 4 + memberN, nil
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

	count := int(Endian.Uint32(buffer[messageTypeN+sourceN:]))

	var memberN int
	for i := 0; i < count; i++ {
		member, n, err := MemberFromBuffer(buffer[messageTypeN+sourceN+4+memberN:])
		if err != nil {
			return 0, err
		}
		memberN += n
		m.Members = append(m.Members, member)
	}

	return messageTypeN + sourceN + 4 + memberN, nil
}
