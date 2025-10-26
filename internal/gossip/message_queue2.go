package gossip

import (
	"fmt"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageQueue2 is responsible for managing the messages we need to gossip. It will sort the gossip messages in a way
// which helps distribute new gossip quickly.
//
// The implementation works with buckets. One bucket for all gossip which was transmitted the same amount of times.
// The buckets are flattened out into a single slice.
type MessageQueue2 struct {
	// queue is the list of gossip entries. The messages are sorted by transmission count increasing. All consecutive
	// messages with the same transmission count form a bucket within the queue. The order of elements within the same
	// bucket is not stable and subject to changes of position when messages are added, removed or moved between
	// buckets.
	queue []MessageQueueEntry

	// endOfBucketIndices provides the indices where the end of the bucket can be found. It is always one bigger than
	// the last element and provides the index where the next element for that bucket can be inserted. The index is
	// the first element of the bucket following.
	endOfBucketIndices []int

	// indexByAddress is correlating the position within queue with the address. This helps in making checks for
	// existing gossip faster.
	indexByAddress map[encoding.Address]int

	// maxTransmissionCount is the maximum times each gossip is transmitted before it is dropped from the queue.
	maxTransmissionCount int

	// priorityQueueIndex is the queue index which should be returned with priority and always with Get(0). This allows
	// us to prioritize suspect and faulty gossip when we are talking to that node right now.
	priorityQueueIndex int
}

// NewMessageQueue2 creates a new gossip message queue.
func NewMessageQueue2(maxTransmissionCount int) *MessageQueue2 {
	return &MessageQueue2{
		queue:                make([]MessageQueueEntry, 0, 1024),
		endOfBucketIndices:   make([]int, 0, 64),
		indexByAddress:       make(map[encoding.Address]int, 1024),
		maxTransmissionCount: maxTransmissionCount,
	}
}

// MessageQueueEntry is a helper struct making up each entry in the queue.
type MessageQueueEntry struct {
	// Message is the message to gossip about.
	Message Message

	// TransmissionCount is the number of times the message has been gossiped.
	TransmissionCount int
}

// Message is the interface all gossip network messages need to implement.
type Message interface {
	AppendToBuffer(buffer []byte) ([]byte, int, error)
	FromBuffer(buffer []byte) (int, error)
	GetAddress() encoding.Address
	GetType() encoding.MessageType
	GetIncarnationNumber() int
}

// Len returns the number of entries currently stored in the queue.
func (q *MessageQueue2) Len() int {
	return len(q.queue)
}

// Clear removes all gossip from the queue.
func (q *MessageQueue2) Clear() {
	q.queue = q.queue[:0]
	q.endOfBucketIndices = q.endOfBucketIndices[:0]
	clear(q.indexByAddress)
}

// Get returns the message at the given index within the queue.
func (q *MessageQueue2) Get(index int) Message {
	if q.priorityQueueIndex == -1 {
		return q.queue[index].Message
	}
	if index == 0 {
		return q.queue[q.priorityQueueIndex].Message
	}
	if index <= q.priorityQueueIndex {
		return q.queue[index-1].Message
	}
	return q.queue[index].Message
}

func (q *MessageQueue2) PrioritizeForAddress(address encoding.Address) {
	queueIndex, found := q.indexByAddress[address]
	if found &&
		(q.queue[queueIndex].Message.GetType() == encoding.MessageTypeSuspect ||
			q.queue[queueIndex].Message.GetType() == encoding.MessageTypeFaulty) {
		q.priorityQueueIndex = queueIndex
		return
	}
	q.priorityQueueIndex = -1
}

// MarkFirstNMessagesTransmitted increases the transmission counter for the first count messages in the queue by
// one, moving them to buckets of higher tiers. When the maximum transmission count is reached, it discards them.
func (q *MessageQueue2) MarkFirstNMessagesTransmitted(count int) {
	count = min(count, len(q.queue))
	for queueIndex := count - 1; 0 <= queueIndex; queueIndex-- {
		// Make sure we have at least one more bucket behind the bucket of our element. This is where we move the
		// element to.
		bucketIndex := q.queue[queueIndex].TransmissionCount
		q.ensureBucketAvailable(bucketIndex + 1)

		// Move the element by one bucket
		q.queue[queueIndex].TransmissionCount++
		q.moveForwardToBucket(bucketIndex, bucketIndex+1, queueIndex)
	}

	q.trimMessages()
	q.trimEmptyBuckets()
}

func (q *MessageQueue2) trimMessages() {
	// Drop all buckets which contain messages which have reached the maximum transmission count.
	for bucketIndex := len(q.endOfBucketIndices) - 1; q.maxTransmissionCount <= bucketIndex; bucketIndex-- {
		startOfBucketIndex := q.endOfBucketIndices[bucketIndex-1]
		endOfBucketIndex := q.endOfBucketIndices[bucketIndex]

		// Remove all elements we are about to drop from the indexByAddress map
		for queueIndex := startOfBucketIndex; queueIndex < endOfBucketIndex; queueIndex++ {
			delete(q.indexByAddress, q.queue[queueIndex].Message.GetAddress())
		}

		// Shorten queue and buckets
		q.queue = q.queue[:startOfBucketIndex]
		q.endOfBucketIndices = q.endOfBucketIndices[:len(q.endOfBucketIndices)-1]
	}
}

func (q *MessageQueue2) trimEmptyBuckets() {
	for bucketIndex := len(q.endOfBucketIndices) - 1; bucketIndex > 0; bucketIndex-- {
		if q.endOfBucketIndices[bucketIndex] != q.endOfBucketIndices[bucketIndex-1] {
			return
		}
		q.endOfBucketIndices = q.endOfBucketIndices[:bucketIndex]
	}
}

// Add adds a new message to the gossip message queue. Always adds to the first bucket, as this message was never
// transmitted.
func (q *MessageQueue2) Add(message Message) {
	queueIndex, ok := q.indexByAddress[message.GetAddress()]
	if ok {
		// The queue already contains a message for that address. Let's check if we need to overwrite it.
		if message.GetIncarnationNumber() < q.queue[queueIndex].Message.GetIncarnationNumber() {
			// No need to overwrite when the incarnation number is lower.
			return
		}
		if message.GetIncarnationNumber() == q.queue[queueIndex].Message.GetIncarnationNumber() &&
			(message.GetType() == encoding.MessageTypeAlive ||
				message.GetType() == encoding.MessageTypeSuspect && q.queue[queueIndex].Message.GetType() != encoding.MessageTypeAlive ||
				message.GetType() == encoding.MessageTypeFaulty && q.queue[queueIndex].Message.GetType() != encoding.MessageTypeAlive && q.queue[queueIndex].Message.GetType() != encoding.MessageTypeSuspect) {
			// No need to overwrite with the same incarnation number and the wrong priorities.
			return
		}

		// Either we have the same incarnation number with the right priorities, or the incarnation number is bigger.
		q.queue[queueIndex].Message = message
		if q.queue[queueIndex].TransmissionCount != 0 {
			queueIndex = q.moveBackwardToBucket(q.queue[queueIndex].TransmissionCount, 0, queueIndex)
		}
		q.queue[queueIndex].TransmissionCount = 0
		q.trimEmptyBuckets()
		return
	}
	q.insertIntoBucket(0, message)
}

func (q *MessageQueue2) insertIntoBucket(targetBucketIndex int, message Message) {
	// Make sure we have enough end of bucket indices in place to deal with the target bucket index.
	q.ensureBucketAvailable(targetBucketIndex)

	// Append the new element at the end of the queue and extend the last bucket to hold the new element.
	q.queue = append(q.queue, MessageQueueEntry{
		Message:           message,
		TransmissionCount: targetBucketIndex,
	})
	queueIndex := len(q.queue) - 1
	q.indexByAddress[message.GetAddress()] = queueIndex
	q.endOfBucketIndices[len(q.endOfBucketIndices)-1]++

	// Move the new element backwards through all buckets until it has reached the desired bucket.
	q.moveBackwardToBucket(len(q.endOfBucketIndices)-1, targetBucketIndex, queueIndex)
}

func (q *MessageQueue2) moveBackwardToBucket(sourceBucketIndex int, destinationBucketIndex int, queueIndex int) int {
	currentBucketIndex := sourceBucketIndex
	for currentBucketIndex > destinationBucketIndex {
		// As long as we have not reached our destination bucket, swap the element with the first of the bucket and
		// extend the bucket before by one.
		// The first of the bucket is the index for the end of bucket of the bucket before.
		newQueueIndex := q.endOfBucketIndices[currentBucketIndex-1]
		q.endOfBucketIndices[currentBucketIndex-1]++
		q.queue[queueIndex], q.queue[newQueueIndex] = q.queue[newQueueIndex], q.queue[queueIndex]

		// We need to update the indices for the swapped elements. Otherwise, they would point to the wrong element.
		q.indexByAddress[q.queue[queueIndex].Message.GetAddress()] = queueIndex
		q.indexByAddress[q.queue[newQueueIndex].Message.GetAddress()] = newQueueIndex

		currentBucketIndex--
		queueIndex = newQueueIndex
	}
	return queueIndex
}

func (q *MessageQueue2) removeFromBucket(queueIndex int) {
	// We temporarily add a new bucket at the end. We then move the element through all buckets into the last bucket.
	// Afterward, we drop that bucket again.

	// Create the new bucket.
	q.ensureBucketAvailable(len(q.endOfBucketIndices))

	// Move the element to the new bucket at the end.
	sourceBucketIndex := q.queue[queueIndex].TransmissionCount
	queueIndex = q.moveForwardToBucket(sourceBucketIndex, len(q.endOfBucketIndices)-1, queueIndex)

	// Drop the bucket and shorten the queue
	delete(q.indexByAddress, q.queue[queueIndex].Message.GetAddress())
	q.endOfBucketIndices = q.endOfBucketIndices[:len(q.endOfBucketIndices)-1]
	q.queue = q.queue[:len(q.queue)-1]
}

func (q *MessageQueue2) moveForwardToBucket(sourceBucketIndex int, destinationBucketIndex int, queueIndex int) int {
	currentBucketIndex := sourceBucketIndex
	for currentBucketIndex < destinationBucketIndex {
		// As long as we have not reached our destination bucket, swap the element with the last of the bucket and
		// shorten the bucket by one.
		newQueueIndex := q.endOfBucketIndices[currentBucketIndex] - 1
		q.endOfBucketIndices[currentBucketIndex]--
		q.queue[queueIndex], q.queue[newQueueIndex] = q.queue[newQueueIndex], q.queue[queueIndex]

		// We need to update the indices for the swapped elements. Otherwise, they would point to the wrong element.
		q.indexByAddress[q.queue[queueIndex].Message.GetAddress()] = queueIndex
		q.indexByAddress[q.queue[newQueueIndex].Message.GetAddress()] = newQueueIndex

		currentBucketIndex++
		queueIndex = newQueueIndex
	}
	return queueIndex
}

func (q *MessageQueue2) ensureBucketAvailable(bucketIndex int) {
	for len(q.endOfBucketIndices) <= bucketIndex {
		q.endOfBucketIndices = append(q.endOfBucketIndices, len(q.queue))
	}
}

// Validate reports if the internal state is valid. This is helpful for automated tests.
func (q *MessageQueue2) Validate() error {
	// Make sure that end of bucket indices are always equal or bigger than the one before.
	for i := 0; i < len(q.endOfBucketIndices)-2; i++ {
		if q.endOfBucketIndices[i] > q.endOfBucketIndices[i+1] {
			return fmt.Errorf("end of bucket index %d is bigger than %d", i, i+1)
		}
	}

	// Make sure that every entry in the queue has an associated bucket
	for queueIndex, entry := range q.queue {
		bucketIndex := entry.TransmissionCount
		if bucketIndex > len(q.endOfBucketIndices)+1 {
			return fmt.Errorf("invalid bucket index for entry %q", queueIndex)
		}
		if queueIndex >= q.endOfBucketIndices[bucketIndex] {
			return fmt.Errorf("queue index %d does not fit into bucket %d", queueIndex, bucketIndex)
		}
		if bucketIndex > 0 && queueIndex < q.endOfBucketIndices[bucketIndex-1] {
			return fmt.Errorf("queue index %d does not fit into bucket %d", queueIndex, bucketIndex)
		}
	}

	// Make sure that no empty buckets are at the end - one bucket always stays
	if len(q.endOfBucketIndices) > 1 &&
		q.endOfBucketIndices[len(q.endOfBucketIndices)-2] == q.endOfBucketIndices[len(q.endOfBucketIndices)-1] {
		return fmt.Errorf("bucket %d is empty", len(q.endOfBucketIndices)-1)
	}

	// Make sure that buckets always point to content in the queue
	for bucketIndex, endOfBucketIndex := range q.endOfBucketIndices {
		if len(q.queue) < endOfBucketIndex {
			return fmt.Errorf("bucket %d points to members out of bounds", bucketIndex)
		}
	}

	// Make sure that every message can be found in the index map
	for queueIndex, message := range q.queue {
		index, found := q.indexByAddress[message.Message.GetAddress()]
		if !found {
			return fmt.Errorf("message %d could not be found in index map", queueIndex)
		}
		if index != queueIndex {
			return fmt.Errorf("message %d has wrong index in index map", queueIndex)
		}
	}

	// Make sure that every entry in the index map can be found in the queue
	for address, queueIndex := range q.indexByAddress {
		if queueIndex >= len(q.queue) {
			return fmt.Errorf("index map for address %s points out of bounds", address)
		}
		if !q.queue[queueIndex].Message.GetAddress().Equal(address) {
			return fmt.Errorf("index map index mismatch for queue element %d", queueIndex)
		}
	}
	return nil
}
