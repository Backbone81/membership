package encoding

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
