package encoding

import "errors"

// MessageType is the kind of network message which is following.
type MessageType int

const (
	MessageTypeNone MessageType = iota // We start with a placeholder message type to detect missing types.
	MessageTypeDirectPing
	MessageTypeDirectAck
	MessageTypeIndirectPing
	MessageTypeIndirectAck
	MessageTypeSuspect
	MessageTypeAlive
	MessageTypeFaulty
	MessageTypeListRequest
	MessageTypeListResponse
)

// AppendMessageTypeToBuffer appends the message type to the provided buffer encoded for network transfer.
// Returns the buffer with the data appended, the number of bytes appended and any error which occurred.
func AppendMessageTypeToBuffer(buffer []byte, messageType MessageType) ([]byte, int, error) {
	return append(buffer, byte(messageType)), 1, nil
}

// MessageTypeFromBuffer reads the message type from the provided buffer.
// Returns the message type, the number of bytes read and any error which occurred.
func MessageTypeFromBuffer(buffer []byte) (MessageType, int, error) {
	if len(buffer) < 1 {
		return 0, 0, errors.New("message type buffer too small")
	}
	return MessageType(buffer[0]), 1, nil
}

func (t MessageType) String() string {
	switch t {
	case MessageTypeDirectPing:
		return "DirectPing"
	case MessageTypeDirectAck:
		return "DirectAck"
	case MessageTypeIndirectPing:
		return "IndirectPing"
	case MessageTypeIndirectAck:
		return "IndirectAck"
	case MessageTypeSuspect:
		return "Suspect"
	case MessageTypeAlive:
		return "Alive"
	case MessageTypeFaulty:
		return "Faulty"
	case MessageTypeListRequest:
		return "ListRequest"
	case MessageTypeListResponse:
		return "ListResponse"
	default:
		return "<unknown>"
	}
}
