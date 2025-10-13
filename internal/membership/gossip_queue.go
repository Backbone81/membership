package membership

import (
	"cmp"
	"slices"
)

// GossipQueue is responsible for managing the messages we need to gossip. It will sort the gossip messages in a way
// which helps distribute new gossip quickly.
type GossipQueue struct {
	queue          []GossipQueueEntry
	indexByAddress map[Address]int
	maxGossipCount int
}

func NewGossipQueue(maxGossipCount int) *GossipQueue {
	return &GossipQueue{
		indexByAddress: make(map[Address]int),
		maxGossipCount: maxGossipCount,
	}
}

type GossipQueueEntry struct {
	Message GossipMessage
	Count   int
}

// GossipMessage is the interface all gossip network messages need to implement.
type GossipMessage interface {
	AppendToBuffer(buffer []byte) ([]byte, int, error)
	FromBuffer(buffer []byte) (int, error)
	GetAddress() Address
	GetType() MessageType
	GetIncarnationNumber() int
}

func (q *GossipQueue) PrepareFor(address Address) {
	q.queue = slices.DeleteFunc(q.queue, func(entry GossipQueueEntry) bool {
		result := entry.Count >= q.maxGossipCount
		if result {
			delete(q.indexByAddress, entry.Message.GetAddress())
		}
		return result
	})

	localQueue := q.queue
	index, ok := q.indexByAddress[address]
	if ok {
		if localQueue[index].Message.GetType() == MessageTypeSuspect || localQueue[index].Message.GetType() == MessageTypeFaulty {
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
	slices.SortFunc(localQueue, func(a, b GossipQueueEntry) int {
		return cmp.Compare(a.Count, b.Count)
	})

	for i, entry := range q.queue {
		q.indexByAddress[entry.Message.GetAddress()] = i
	}
}

func (q *GossipQueue) Len() int {
	return len(q.queue)
}

func (q *GossipQueue) Get(index int) GossipMessage {
	return q.queue[index].Message
}

func (q *GossipQueue) MarkGossiped(index int) {
	q.queue[index].Count++
}

func (q *GossipQueue) Add(message GossipMessage) {
	index, ok := q.indexByAddress[message.GetAddress()]
	if ok {
		// The queue already contains a message for that address. Let's check if we need to overwrite it.
		if message.GetIncarnationNumber() < q.queue[index].Message.GetIncarnationNumber() {
			// No need to overwrite when the incarnation number is lower.
			return
		}
		if message.GetIncarnationNumber() == q.queue[index].Message.GetIncarnationNumber() &&
			(message.GetType() == MessageTypeAlive ||
				message.GetType() == MessageTypeSuspect && q.queue[index].Message.GetType() != MessageTypeAlive ||
				message.GetType() == MessageTypeFaulty && q.queue[index].Message.GetType() != MessageTypeAlive && q.queue[index].Message.GetType() != MessageTypeSuspect) {
			// No need to overwrite with the same incarnation number and the wrong priorities.
			return
		}

		// Either we have the same incarnation number with the right priorities, or the incarnation number is bigger.
		q.queue[index].Message = message
		q.queue[index].Count = 0
		return
	}

	q.queue = append(q.queue, GossipQueueEntry{
		Message: message,
	})
	q.indexByAddress[message.GetAddress()] = len(q.queue) - 1
}
