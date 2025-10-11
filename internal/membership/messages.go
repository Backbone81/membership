package membership

import "errors"

// MessageDirectPing is a ping message directly sent to the recipient.
// This is the `ping` message of SWIM chapter 3.1. SWIM Failure Detector.
type MessageDirectPing struct {
	Source         Address
	SequenceNumber int
}

func (m *MessageDirectPing) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeDirectPing)
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

func (m *MessageDirectPing) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeDirectPing {
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

// MessageDirectAck is a response message sent back in answer to receiving a MessageDirectPing.
// This is the `ack` message of SWIM chapter 3.1. SWIM Failure Detector.
type MessageDirectAck struct {
	Source         Address
	SequenceNumber int
}

func (m *MessageDirectAck) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeDirectAck)
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

func (m *MessageDirectAck) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeDirectAck {
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

// MessageIndirectPing is a request of the recipient to send a MessageDirectPing to the destination.
// This is the `ping-req` message of SWIM chapter 3.1. SWIM Failure Detector.
type MessageIndirectPing struct {
	Source         Address
	Destination    Address
	SequenceNumber int
}

func (m *MessageIndirectPing) IsEmpty() bool {
	return m.Source.IsZero()
}

func (m *MessageIndirectPing) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeIndirectPing)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := AppendAddressToBuffer(sourceBuffer, m.Destination)
	if err != nil {
		return buffer, 0, err
	}

	sequenceNumberBuffer, sequenceNumberN, err := AppendSequenceNumberToBuffer(destinationBuffer, m.SequenceNumber)
	if err != nil {
		return buffer, 0, err
	}

	return sequenceNumberBuffer, messageTypeN + sourceN + destinationN + sequenceNumberN, nil
}

func (m *MessageIndirectPing) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeIndirectPing {
		return 0, errors.New("invalid message type")
	}

	var sourceN, destinationN, sequenceNumberN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = AddressFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	m.SequenceNumber, sequenceNumberN, err = SequenceNumberFromBuffer(buffer[messageTypeN+sourceN+destinationN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + destinationN + sequenceNumberN, nil
}

// MessageIndirectAck is a response message sent back in response to receiving a MessageIndirectPing and receiving
// a MessageDirectAck from the destination. The message is identical to MessageDirectAck but allows us to differentiate
// those messages when calculating round trip times later.
type MessageIndirectAck struct {
	Source         Address
	SequenceNumber int
}

func (m *MessageIndirectAck) AppendToBuffer(buffer []byte) ([]byte, int, error) {
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

func (m *MessageIndirectAck) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
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

// MessageSuspect declares the destination as being suspected for failure by the source.
// This is the `Suspect` message of SWIM chapter 4.2. Suspicion Mechanism: Reducing the Frequency of False Positives.
type MessageSuspect struct {
	Source            Address
	Destination       Address
	IncarnationNumber int
}

func (m *MessageSuspect) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeSuspect)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := AppendAddressToBuffer(sourceBuffer, m.Destination)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := AppendIncarnationNumberToBuffer(destinationBuffer, m.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

func (m *MessageSuspect) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeSuspect {
		return 0, errors.New("invalid message type")
	}

	var sourceN, destinationN, incarnationNumberN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = AddressFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	m.IncarnationNumber, incarnationNumberN, err = IncarnationNumberFromBuffer(buffer[messageTypeN+sourceN+destinationN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

// MessageAlive declares the destination as being alive by the source.
// This is the `Alive` message of SWIM chapter 4.2. Suspicion Mechanism: Reducing the Frequency of False Positives.
type MessageAlive struct {
	Source            Address
	IncarnationNumber int
}

func (m *MessageAlive) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeAlive)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := AppendIncarnationNumberToBuffer(sourceBuffer, m.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, messageTypeN + sourceN + incarnationNumberN, nil
}

func (m *MessageAlive) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeAlive {
		return 0, errors.New("invalid message type")
	}

	var sourceN, incarnationNumberN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.IncarnationNumber, incarnationNumberN, err = IncarnationNumberFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + incarnationNumberN, nil
}

// MessageFaulty declares the destination as being faulty by the source.
// This is the `Confirm` message of SWIM chapter 4.2. Suspicion Mechanism: Reducing the Frequency of False Positives.
type MessageFaulty struct {
	Source            Address
	Destination       Address
	IncarnationNumber int
}

func (m *MessageFaulty) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeFaulty)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendAddressToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := AppendAddressToBuffer(sourceBuffer, m.Destination)
	if err != nil {
		return buffer, 0, err
	}

	incarnationNumberBuffer, incarnationNumberN, err := AppendIncarnationNumberToBuffer(destinationBuffer, m.IncarnationNumber)
	if err != nil {
		return buffer, 0, err
	}

	return incarnationNumberBuffer, messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

func (m *MessageFaulty) FromBuffer(buffer []byte) (int, error) {
	messageType, messageTypeN, err := MessageTypeFromBuffer(buffer)
	if messageType != MessageTypeFaulty {
		return 0, errors.New("invalid message type")
	}

	var sourceN, destinationN, incarnationNumberN int
	m.Source, sourceN, err = AddressFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = AddressFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	m.IncarnationNumber, incarnationNumberN, err = IncarnationNumberFromBuffer(buffer[messageTypeN+sourceN+destinationN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}

// MessageListRequest asks the recipient to send its current full member list to the source. This helps in making sure
// new-joiners quickly get an overview over all members.
type MessageListRequest struct {
	Source Address
}

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

// MessageListResponse provides a list of all known members. This message can become quite big and should always be
// transmitted over TCP and not UDP.
type MessageListResponse struct {
	Source  Address
	Members []Member
}

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
