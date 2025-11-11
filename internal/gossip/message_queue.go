package gossip

import (
	"fmt"

	"github.com/backbone81/membership/internal/encoding"
)

// MessageQueue is responsible for managing the messages we need to gossip. It will sort the gossip messages in a way
// which helps distribute new gossip quickly.
//
// The implementation works with buckets. One bucket for all gossip which was transmitted the same amount of times.
// The buckets are flattened out into a single slice.
type MessageQueue struct {
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

// NewMessageQueue creates a new gossip message queue.
func NewMessageQueue(maxTransmissionCount int) *MessageQueue {
	return &MessageQueue{
		queue:                make([]MessageQueueEntry, 0, 1024),
		endOfBucketIndices:   make([]int, 0, 64),
		indexByAddress:       make(map[encoding.Address]int, 1024),
		maxTransmissionCount: maxTransmissionCount, // TODO: there should be an option to configure it at the start.
		priorityQueueIndex:   -1,
	}
}

// SetMaxTransmissionCount updates the max transmission count to the new value. The new value will be used on the next
// call to MarkFirstNMessagesTransmitted.
func (q *MessageQueue) SetMaxTransmissionCount(maxTransmissionCount int) {
	q.maxTransmissionCount = max(3, maxTransmissionCount)
}

// Len returns the number of entries currently stored in the queue.
func (q *MessageQueue) Len() int {
	return len(q.queue)
}

// Clear removes all gossip from the queue.
func (q *MessageQueue) Clear() {
	q.queue = q.queue[:0]
	q.endOfBucketIndices = q.endOfBucketIndices[:0]
	clear(q.indexByAddress)
}

// Add adds a new message to the gossip message queue.
func (q *MessageQueue) Add(message Message) {
	queueIndex, ok := q.indexByAddress[message.GetAddress()]
	if ok {
		queueEntry := &q.queue[queueIndex]

		// The queue already contains a message for that address. Let's check if we need to overwrite it.
		newIncarnationNumber := message.GetIncarnationNumber()
		existingIncarnationNumber := queueEntry.Message.GetIncarnationNumber()
		if newIncarnationNumber < existingIncarnationNumber {
			// No need to overwrite when the incarnation number is lower.
			return
		}

		newMessageType := message.GetType()
		existingMessageType := queueEntry.Message.GetType()
		if newIncarnationNumber == existingIncarnationNumber &&
			(newMessageType == encoding.MessageTypeAlive ||
				newMessageType == encoding.MessageTypeSuspect && existingMessageType != encoding.MessageTypeAlive ||
				newMessageType == encoding.MessageTypeFaulty && existingMessageType != encoding.MessageTypeAlive && existingMessageType != encoding.MessageTypeSuspect) {
			// No need to overwrite with the same incarnation number and the wrong priorities.
			return
		}

		// Either we have the same incarnation number with the right priorities, or the incarnation number is bigger.
		queueEntry.Message = message
		if queueEntry.TransmissionCount != 0 {
			queueIndex = q.moveBackwardToBucket(queueEntry.TransmissionCount, 0, queueIndex)
			queueEntry = &q.queue[queueIndex]
		}
		queueEntry.TransmissionCount = 0

		// As we might have moved the last element from the last bucket to the front, we need to get rid of empty
		// buckets at the end. Note that we do not count this towards the add message metric. Otherwise, the
		// number of gossip messages could not be calculated by subtracting remove gossip message metric from add gossip
		// message metric.
		q.trimEmptyBuckets()
		return
	}
	q.insertIntoBucket(0, message)
	AddMessageTotal.Inc()
}

// insertIntoBucket inserts the given message into the bucket.
func (q *MessageQueue) insertIntoBucket(destinationBucketIndex int, message Message) {
	// Make sure we have enough end of bucket indices in place to deal with the target bucket index.
	q.ensureBucketAvailable(destinationBucketIndex)

	// Append the new element at the end of the queue and extend the last bucket to hold the new element.
	q.queue = append(q.queue, MessageQueueEntry{
		Message:           message,
		TransmissionCount: destinationBucketIndex,
	})
	// Note: We deliberately do not set indexByAddress here, because the index is set at the end of
	// moveBackwardToBucket anyway.
	q.endOfBucketIndices[len(q.endOfBucketIndices)-1]++

	// Move the new element backwards through all buckets until it has reached the desired bucket.
	q.moveBackwardToBucket(len(q.endOfBucketIndices)-1, destinationBucketIndex, len(q.queue)-1)
}

// moveBackwardToBucket moves the element at queueIndex from sourceBucketIndex to destinationBucketIndex.
// destinationBucketIndex needs to be smaller or equal to sourceBucketIndex.
// Returns the new queueIndex the element was moved to.
func (q *MessageQueue) moveBackwardToBucket(sourceBucketIndex int, destinationBucketIndex int, queueIndex int) int {
	currentBucketIndex := sourceBucketIndex
	for currentBucketIndex > destinationBucketIndex {
		// As long as we have not reached our destination bucket, swap the element with the first of the bucket and
		// extend the bucket before by one.
		// The first of the bucket is the index for the end of bucket of the bucket before.
		newQueueIndex := q.endOfBucketIndices[currentBucketIndex-1]
		q.endOfBucketIndices[currentBucketIndex-1]++
		q.queue[queueIndex], q.queue[newQueueIndex] = q.queue[newQueueIndex], q.queue[queueIndex]

		// We need to update the indices for the swapped elements. Otherwise, they would point to the wrong element.
		// As our queueIndex element potentially moves through multiple buckets, we only update the index for the
		// element we swapped in. The other element is updated at the end of the function.
		q.indexByAddress[q.queue[queueIndex].Message.GetAddress()] = queueIndex

		currentBucketIndex--
		queueIndex = newQueueIndex
	}

	// When we are done moving our element through all buckets, we update indexByAddress with the final index.
	q.indexByAddress[q.queue[queueIndex].Message.GetAddress()] = queueIndex
	return queueIndex
}

// PrioritizeForAddress marks a gossip for the given address for priority. If such a message exists, it will always be
// returns with Get(0). Otherwise, this method has no effect.
func (q *MessageQueue) PrioritizeForAddress(address encoding.Address) {
	queueIndex, found := q.indexByAddress[address]
	if found {
		messageType := q.queue[queueIndex].Message.GetType()
		if messageType == encoding.MessageTypeSuspect ||
			messageType == encoding.MessageTypeFaulty {
			q.priorityQueueIndex = queueIndex
			return
		}
	}
	q.priorityQueueIndex = -1
}

// Get returns the message at the given index within the queue. If PrioritizeForAddress was called before, the
// message for that address is always returned by Get(0).
// index needs to be between 0 and Len() exclusive.
func (q *MessageQueue) Get(index int) Message {
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

// MarkFirstNMessagesTransmitted increases the transmission counter for the first count messages in the queue by
// one, moving them to buckets of higher tiers. When the maximum transmission count is reached, it discards them.
func (q *MessageQueue) MarkFirstNMessagesTransmitted(count int) {
	count = min(count, len(q.queue))

	// We move backwards through the first count elements to avoid swapping with an element which is part of the first
	// count elements.
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

// trimMessages removes all messages from the end of the queue which have reached the maximum transmission count.
func (q *MessageQueue) trimMessages() {
	// Drop all buckets which contain messages which have reached the maximum transmission count.
	// We are iterating from the back to the front to not remove buckets which need to stay.
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
		RemoveMessageTotal.Add(float64(endOfBucketIndex - startOfBucketIndex))
	}
}

// trimEmptyBuckets removes all buckets from the end which are empty. The first bucket always stays.
func (q *MessageQueue) trimEmptyBuckets() {
	for bucketIndex := len(q.endOfBucketIndices) - 1; bucketIndex > 0; bucketIndex-- {
		if q.endOfBucketIndices[bucketIndex] != q.endOfBucketIndices[bucketIndex-1] {
			return
		}
		q.endOfBucketIndices = q.endOfBucketIndices[:bucketIndex]
	}
}

// moveForwardToBucket moves the element at queueIndex from sourceBucketIndex to destinationBucketIndex.
// destinationBucketIndex needs to be bigger or equal to sourceBucketIndex.
// Returns the new queueIndex the element was moved to.
func (q *MessageQueue) moveForwardToBucket(sourceBucketIndex int, destinationBucketIndex int, queueIndex int) int {
	currentBucketIndex := sourceBucketIndex
	for currentBucketIndex < destinationBucketIndex {
		// As long as we have not reached our destination bucket, swap the element with the last of the bucket and
		// shorten the bucket by one.
		newQueueIndex := q.endOfBucketIndices[currentBucketIndex] - 1
		q.endOfBucketIndices[currentBucketIndex]--
		q.queue[queueIndex], q.queue[newQueueIndex] = q.queue[newQueueIndex], q.queue[queueIndex]

		// We need to update the indices for the swapped elements. Otherwise, they would point to the wrong element.
		// As our queueIndex element potentially moves through multiple buckets, we only update the index for the
		// element we swapped in. The other element is updated at the end of the function.
		q.indexByAddress[q.queue[queueIndex].Message.GetAddress()] = queueIndex

		currentBucketIndex++
		queueIndex = newQueueIndex
	}

	// When we are done moving our element through all buckets, we update indexByAddress with the final index.
	q.indexByAddress[q.queue[queueIndex].Message.GetAddress()] = queueIndex
	return queueIndex
}

// ensureBucketAvailable creates buckets if needed. Allowing other code to safely access the bucket without worrying
// if the index is valid or not.
func (q *MessageQueue) ensureBucketAvailable(bucketIndex int) {
	for len(q.endOfBucketIndices) <= bucketIndex {
		q.endOfBucketIndices = append(q.endOfBucketIndices, len(q.queue))
	}
}

// ValidateInternalState reports if the internal state is valid.
// This function is expensive and should not be called outside of tests.
func (q *MessageQueue) ValidateInternalState() error {
	// Make sure that end of bucket indices are always equal or bigger than the one before.
	for i := range len(q.endOfBucketIndices) - 1 {
		if q.endOfBucketIndices[i] > q.endOfBucketIndices[i+1] {
			return fmt.Errorf("end of bucket index %d is bigger than %d", i, i+1)
		}
	}

	// Make sure that every entry in the queue has an associated bucket
	for queueIndex, entry := range q.queue {
		bucketIndex := entry.TransmissionCount
		if bucketIndex >= len(q.endOfBucketIndices) {
			return fmt.Errorf("invalid bucket index for entry %d", queueIndex)
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
