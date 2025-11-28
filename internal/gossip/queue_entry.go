package gossip

import "github.com/backbone81/membership/internal/encoding"

// QueueEntry is a helper struct making up each entry in the queue.
type QueueEntry struct {
	// Message is the message to gossip about.
	Message encoding.Message

	// TransmissionCount is the number of times the message has been gossiped.
	TransmissionCount int
}
