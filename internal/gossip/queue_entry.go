package gossip

// QueueEntry is a helper struct making up each entry in the queue.
type QueueEntry struct {
	// Message is the message to gossip about.
	Message Message

	// TransmissionCount is the number of times the message has been gossiped.
	TransmissionCount int
}
