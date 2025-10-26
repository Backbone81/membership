package gossip

import (
	"cmp"
	"slices"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageQueue is responsible for managing the messages we need to gossip. It will sort the gossip messages in a way
// which helps distribute new gossip quickly.
type MessageQueue struct {
	// queue is the list of gossip entries.
	queue []MessageQueueEntry

	// indexByAddress is correlating the position within queue with the address. This helps in making checks for
	// existing gossip faster.
	indexByAddress map[encoding.Address]int

	// maxTransmissionCount is the maximum times each gossip is transmitted before it is dropped from the queue.
	maxTransmissionCount int
}

// NewMessageQueue creates a new gossip message queue.
func NewMessageQueue(maxTransmissionCount int) *MessageQueue {
	return &MessageQueue{
		indexByAddress:       make(map[encoding.Address]int),
		maxTransmissionCount: maxTransmissionCount,
	}
}

// PrioritizeForAddress brings the messages in the gossip message queue into the desired order.
//
// Messages which have been transmitted the least amount are placed first, those which have been transmitted most are
// placed last.
// If the queue contains gossip about the address given as parameter, that gossip is placed in the first place if it is
// a message about that member being suspect or faulty to enable the member an early action against that gossip. If the
// gossip is an alive message, it is placed last, as that is not important for the member.
//
// This method will delete gossip which has exceeded the maximum transmission count.
func (q *MessageQueue) PrioritizeForAddress(address encoding.Address) {
	q.queue = slices.DeleteFunc(q.queue, func(entry MessageQueueEntry) bool {
		return entry.TransmissionCount >= q.maxTransmissionCount
	})
	clear(q.indexByAddress)
	for i, entry := range q.queue {
		q.indexByAddress[entry.Message.GetAddress()] = i
	}

	localQueue := q.queue
	index, ok := q.indexByAddress[address]
	if ok {
		if localQueue[index].Message.GetType() == encoding.MessageTypeSuspect || localQueue[index].Message.GetType() == encoding.MessageTypeFaulty {
			// We already have a suspect or faulty gossip message for the address we are preparing. Move that gossip to
			// the start of the queue to gossip it with high priority.
			localQueue[0], localQueue[index] = localQueue[index], localQueue[0]
			localQueue = localQueue[1:]
		} else {
			// We already have an alive gossip message for the address we are preparing. Move that gossip to
			// the end of the queue to gossip it with low priority.
			localQueue[len(localQueue)-1], localQueue[index] = localQueue[index], localQueue[len(localQueue)-1]
			localQueue = localQueue[:len(localQueue)-1]
		}
	}

	// Let's make sure that our gossip is ordered with the least gossiped first. We are sorting what we did not move
	// to the start or end of the queue if the address was already there.
	slices.SortFunc(localQueue, func(a, b MessageQueueEntry) int {
		return cmp.Compare(a.TransmissionCount, b.TransmissionCount)
	})

	clear(q.indexByAddress)
	for i, entry := range q.queue {
		q.indexByAddress[entry.Message.GetAddress()] = i
	}
}

// Len returns the number of entries currently stored in the queue.
func (q *MessageQueue) Len() int {
	return len(q.queue)
}

// Clear removes all gossip from the queue.
func (q *MessageQueue) Clear() {
	q.queue = q.queue[:0]
	clear(q.indexByAddress)
}

// Get returns the message at the given index within the queue.
// Call PrioritizeForAddress before you iterate over the content of the queue. Otherwise, the order of messages is undefined.
func (q *MessageQueue) Get(index int) Message {
	return q.queue[index].Message
}

// MarkTransmitted increases the transmission counter for the given message by one.
func (q *MessageQueue) MarkTransmitted(index int) {
	q.queue[index].TransmissionCount++
}

// Add adds a new message to the gossip message queue.
//
// In case the gossip message queue already contains a message for the same member, the existing message is overwritten
// by the priority of messages. Whenever a message is overwritten, the transmission counter is reset to zero.
//
// Note that calls to this method need to be fast, as we might need to add multiple gossip messages when receiving
// pings and acks.
func (q *MessageQueue) Add(message Message) {
	index, ok := q.indexByAddress[message.GetAddress()]
	if ok {
		// The queue already contains a message for that address. Let's check if we need to overwrite it.
		if message.GetIncarnationNumber() < q.queue[index].Message.GetIncarnationNumber() {
			// No need to overwrite when the incarnation number is lower.
			return
		}
		if message.GetIncarnationNumber() == q.queue[index].Message.GetIncarnationNumber() &&
			(message.GetType() == encoding.MessageTypeAlive ||
				message.GetType() == encoding.MessageTypeSuspect && q.queue[index].Message.GetType() != encoding.MessageTypeAlive ||
				message.GetType() == encoding.MessageTypeFaulty && q.queue[index].Message.GetType() != encoding.MessageTypeAlive && q.queue[index].Message.GetType() != encoding.MessageTypeSuspect) {
			// No need to overwrite with the same incarnation number and the wrong priorities.
			return
		}

		// Either we have the same incarnation number with the right priorities, or the incarnation number is bigger.
		q.queue[index].Message = message
		q.queue[index].TransmissionCount = 0
		return
	}

	q.queue = append(q.queue, MessageQueueEntry{
		Message: message,
	})
	q.indexByAddress[message.GetAddress()] = len(q.queue) - 1
}
