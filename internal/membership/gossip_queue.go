package membership

import (
	"cmp"
	"slices"
)

type GossipQueue struct {
	queue          []GossipQueueEntry
	maxGossipCount int
}

func NewGossipQueue(maxGossipCount int) *GossipQueue {
	return &GossipQueue{
		maxGossipCount: maxGossipCount,
	}
}

type GossipQueueEntry struct {
	Message Message
	Count   int
}

func (q *GossipQueue) Prepare() {
	// Let's first make sure that our gossip is ordered with the least gossiped first.
	slices.SortFunc(q.queue, func(a, b GossipQueueEntry) int {
		return cmp.Compare(a.Count, b.Count)
	})

	// As we now have a correctly ordered list, let's drop the gossip which we already have gossiped enough. That is
	// always at the end of our list, so we look from the end until we find an entry which is below the threshold.
	for i := len(q.queue) - 1; i >= 0; i-- {
		if q.queue[i].Count < q.maxGossipCount {
			break
		}
		q.queue = q.queue[:i]
	}
}

func (q *GossipQueue) Len() int {
	return len(q.queue)
}

func (q *GossipQueue) Get(index int) Message {
	return q.queue[index].Message
}

func (q *GossipQueue) MarkGossiped(index int) {
	q.queue[index].Count++
}

func (q *GossipQueue) Add(message Message) {

	// TODO: We need to replace messages targeting the same member with a message priority guided by the incarnation
	// number.

	q.queue = append(q.queue, GossipQueueEntry{
		Message: message,
	})
}
