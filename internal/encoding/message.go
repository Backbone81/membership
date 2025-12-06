package encoding

import (
	"fmt"
)

// Message is a struct which holds all potential fields of all network messages we are sending and receiving. This
// allows us to prevent memory allocations caused by interface conversions.
type Message struct {
	// Type is the message type. The other fields in this struct are filled according to this field.
	Type MessageType

	// Source is the member which providing gossip about Destination.
	Source Address

	// Destination is the member which is targeted by gossip of Source.
	Destination Address

	// IncarnationNumber is the incarnation to distinguish an outdated message from a new one. Only the member itself
	// can increase the incarnation number.
	IncarnationNumber uint16

	// SequenceNumber is the sequence we expect to get back in the direct ack. The sequence number should be different
	// for every direct ping we send out.
	SequenceNumber uint16

	// Members is the full member list returned by the member.
	Members []Member
}

//nolint:cyclop
func (m Message) String() string {
	switch m.Type {
	case MessageTypeDirectPing:
		return m.ToDirectPing().String()
	case MessageTypeDirectAck:
		return m.ToDirectAck().String()
	case MessageTypeIndirectPing:
		return m.ToIndirectPing().String()
	case MessageTypeIndirectAck:
		return m.ToIndirectAck().String()
	case MessageTypeSuspect:
		return m.ToSuspect().String()
	case MessageTypeAlive:
		return m.ToAlive().String()
	case MessageTypeFaulty:
		return m.ToFaulty().String()
	case MessageTypeListRequest:
		return m.ToListRequest().String()
	case MessageTypeListResponse:
		return m.ToListResponse().String()
	default:
		return "<unknown message type>"
	}
}

//nolint:cyclop
func (m Message) AppendToBuffer(buffer []byte) ([]byte, int, error) {
	switch m.Type {
	case MessageTypeDirectPing:
		return m.ToDirectPing().AppendToBuffer(buffer)
	case MessageTypeDirectAck:
		return m.ToDirectAck().AppendToBuffer(buffer)
	case MessageTypeIndirectPing:
		return m.ToIndirectPing().AppendToBuffer(buffer)
	case MessageTypeIndirectAck:
		return m.ToIndirectAck().AppendToBuffer(buffer)
	case MessageTypeSuspect:
		return m.ToSuspect().AppendToBuffer(buffer)
	case MessageTypeAlive:
		return m.ToAlive().AppendToBuffer(buffer)
	case MessageTypeFaulty:
		return m.ToFaulty().AppendToBuffer(buffer)
	case MessageTypeListRequest:
		return m.ToListRequest().AppendToBuffer(buffer)
	case MessageTypeListResponse:
		return m.ToListResponse().AppendToBuffer(buffer)
	default:
		return buffer, 0, fmt.Errorf("unknown message type %d", m.Type)
	}
}

func (m Message) ToAlive() MessageAlive {
	return MessageAlive{
		Destination:       m.Destination,
		IncarnationNumber: m.IncarnationNumber,
	}
}

func (m Message) ToSuspect() MessageSuspect {
	return MessageSuspect{
		Source:            m.Source,
		Destination:       m.Destination,
		IncarnationNumber: m.IncarnationNumber,
	}
}

func (m Message) ToFaulty() MessageFaulty {
	return MessageFaulty{
		Source:            m.Source,
		Destination:       m.Destination,
		IncarnationNumber: m.IncarnationNumber,
	}
}

func (m Message) ToDirectPing() MessageDirectPing {
	return MessageDirectPing{
		Source:         m.Source,
		SequenceNumber: m.SequenceNumber,
	}
}

func (m Message) ToDirectAck() MessageDirectAck {
	return MessageDirectAck{
		Source:         m.Source,
		SequenceNumber: m.SequenceNumber,
	}
}

func (m Message) ToIndirectPing() MessageIndirectPing {
	return MessageIndirectPing{
		Source:         m.Source,
		Destination:    m.Destination,
		SequenceNumber: m.SequenceNumber,
	}
}

func (m Message) ToIndirectAck() MessageIndirectAck {
	return MessageIndirectAck{
		Source:         m.Source,
		SequenceNumber: m.SequenceNumber,
	}
}

func (m Message) ToListRequest() MessageListRequest {
	return MessageListRequest{
		Source: m.Source,
	}
}

func (m Message) ToListResponse() MessageListResponse {
	return MessageListResponse{
		Source:  m.Source,
		Members: m.Members,
	}
}
