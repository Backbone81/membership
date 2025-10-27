package gossip

// MessageQueueEntry is a helper struct making up each entry in the queue.
type MessageQueueEntry struct {
	// Message is the message to gossip about.
	Message Message

	// TransmissionCount is the number of times the message has been gossiped.
	TransmissionCount int
}
