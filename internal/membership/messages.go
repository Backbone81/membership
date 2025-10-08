package membership

import "errors"

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
)

func AppendMessageTypeToBuffer(buffer []byte, messageType MessageType) ([]byte, int, error) {
	return append(buffer, byte(messageType)), 1, nil
}

func MessageTypeFromBuffer(buffer []byte) (MessageType, int, error) {
	if len(buffer) < 1 {
		return 0, 0, errors.New("message type buffer too small")
	}
	return MessageType(buffer[0]), 1, nil
}

// MessageDirectPing is a ping message directly sent to the recipient.
// This is the `ping` message of SWIM chapter 3.1. SWIM Failure Detector.
type MessageDirectPing struct {
	Source         Endpoint
	SequenceNumber int
}

func (m *MessageDirectPing) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeDirectPing)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendEndpointToBuffer(messageTypeBuffer, m.Source)
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
	m.Source, sourceN, err = EndpointFromBuffer(buffer[messageTypeN:])
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
	Source         Endpoint
	SequenceNumber int
}

func (m *MessageDirectAck) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeDirectAck)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendEndpointToBuffer(messageTypeBuffer, m.Source)
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
	m.Source, sourceN, err = EndpointFromBuffer(buffer[messageTypeN:])
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
	Source         Endpoint
	Destination    Endpoint
	SequenceNumber int
}

func (m *MessageIndirectPing) IsEmpty() bool {
	return m.Source.IsEmpty()
}

func (m *MessageIndirectPing) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeIndirectPing)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendEndpointToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := AppendEndpointToBuffer(sourceBuffer, m.Destination)
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
	m.Source, sourceN, err = EndpointFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = EndpointFromBuffer(buffer[messageTypeN+sourceN:])
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
	Source         Endpoint
	SequenceNumber int
}

func (m *MessageIndirectAck) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeIndirectAck)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendEndpointToBuffer(messageTypeBuffer, m.Source)
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
	m.Source, sourceN, err = EndpointFromBuffer(buffer[messageTypeN:])
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
	Source            Endpoint
	Destination       Endpoint
	IncarnationNumber int
}

func (m *MessageSuspect) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeSuspect)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendEndpointToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := AppendEndpointToBuffer(sourceBuffer, m.Destination)
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
	m.Source, sourceN, err = EndpointFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = EndpointFromBuffer(buffer[messageTypeN+sourceN:])
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
	Source            Endpoint
	IncarnationNumber int
}

func (m *MessageAlive) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeAlive)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendEndpointToBuffer(messageTypeBuffer, m.Source)
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
	m.Source, sourceN, err = EndpointFromBuffer(buffer[messageTypeN:])
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
	Source            Endpoint
	Destination       Endpoint
	IncarnationNumber int
}

func (m *MessageFaulty) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	messageTypeBuffer, messageTypeN, err := AppendMessageTypeToBuffer(buffer, MessageTypeFaulty)
	if err != nil {
		return buffer, 0, err
	}

	sourceBuffer, sourceN, err := AppendEndpointToBuffer(messageTypeBuffer, m.Source)
	if err != nil {
		return buffer, 0, err
	}

	destinationBuffer, destinationN, err := AppendEndpointToBuffer(sourceBuffer, m.Destination)
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
	m.Source, sourceN, err = EndpointFromBuffer(buffer[messageTypeN:])
	if err != nil {
		return 0, err
	}

	m.Destination, destinationN, err = EndpointFromBuffer(buffer[messageTypeN+sourceN:])
	if err != nil {
		return 0, err
	}

	m.IncarnationNumber, incarnationNumberN, err = IncarnationNumberFromBuffer(buffer[messageTypeN+sourceN+destinationN:])
	if err != nil {
		return 0, err
	}

	return messageTypeN + sourceN + destinationN + incarnationNumberN, nil
}
