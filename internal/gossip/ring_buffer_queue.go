package gossip

import (
	"iter"

	"github.com/backbone81/membership/internal/encoding"
)

type RingBufferQueue struct {
	ring []MessageQueueEntry

	// head provides the index into the ring of the next write
	head int

	// tail provides the index into the ring of the next read
	tail int

	// bucketStarts provides the index into the ring where the bucket starts.
	// Bucket 0: [bucketStarts[0], head)
	// Bucket 1: [bucketStarts[1], bucketStarts[0])
	// Bucket 2: [bucketStarts[2], bucketStarts[1])
	// ...
	// Bucket n: [tail, bucketStarts[n-1])
	bucketStarts []int

	// indexByAddress provides the index into the ring for a given address. This helps in making checks for
	// existing gossip faster.
	indexByAddress map[encoding.Address]int

	// priorityIndex is the ring index which should be returned with priority and always with Get(0). This allows
	// us to prioritize suspect and faulty gossip when we are talking to that node right now.
	priorityIndex int
}

// NewRingBufferQueue creates a new gossip message queue.
func NewRingBufferQueue(maxTransmissionCount int) *RingBufferQueue {
	maxTransmissionCount = max(1, maxTransmissionCount)
	return &RingBufferQueue{
		ring:           make([]MessageQueueEntry, 1024),
		bucketStarts:   make([]int, maxTransmissionCount),
		indexByAddress: make(map[encoding.Address]int, 1024),
		priorityIndex:  -1,
	}
}

// SetMaxTransmissionCount updates the max transmission count to the new value. The new value will be used on the next
// call to MarkFirstNMessagesTransmitted.
func (q *RingBufferQueue) SetMaxTransmissionCount(maxTransmissionCount int) {
	maxTransmissionCount = max(1, maxTransmissionCount)
	if maxTransmissionCount < len(q.bucketStarts) {
		q.bucketStarts = q.bucketStarts[:maxTransmissionCount]
		q.cleanupTail()
	} else {
		for len(q.bucketStarts) < maxTransmissionCount {
			// New buckets start always at the same index as the last bucket to create a zero length bucket.
			q.bucketStarts = append(q.bucketStarts, q.bucketStarts[len(q.bucketStarts)-1])
		}
	}
}

func (q *RingBufferQueue) Len() int {
	return (q.head - q.tail + len(q.ring)) % len(q.ring)
}

func (q *RingBufferQueue) Clear() {
	clear(q.ring)
	clear(q.bucketStarts)
	clear(q.indexByAddress)
	q.head = 0
	q.tail = 0
	q.priorityIndex = -1
}

func (q *RingBufferQueue) Add(message Message) {
	ringIndex, found := q.indexByAddress[message.GetAddress()]
	if found {
		entry := &q.ring[ringIndex]

		// The queue already contains a message for that address. Let's check if we need to overwrite it.
		newIncarnationNumber := message.GetIncarnationNumber()
		existingIncarnationNumber := entry.Message.GetIncarnationNumber()
		if newIncarnationNumber < existingIncarnationNumber {
			// No need to overwrite when the incarnation number is lower.
			return
		}

		newMessageType := message.GetType()
		existingMessageType := entry.Message.GetType()
		if newIncarnationNumber == existingIncarnationNumber &&
			(newMessageType == encoding.MessageTypeAlive ||
				newMessageType == encoding.MessageTypeSuspect && existingMessageType != encoding.MessageTypeAlive ||
				newMessageType == encoding.MessageTypeFaulty && existingMessageType != encoding.MessageTypeAlive && existingMessageType != encoding.MessageTypeSuspect) {
			// No need to overwrite with the same incarnation number and the wrong priorities.
			return
		}

		// Either we have the same incarnation number with the right priorities, or the incarnation number is bigger.
		entry.Message = message
		if entry.TransmissionCount != 0 {
			ringIndex = q.moveToFirstBucket(ringIndex)
		}
		return
	}

	// Make sure we have space left in our ring buffer to add the new element.
	if (q.head+1)%len(q.ring) == q.tail {
		q.grow()
	}

	// Add element and update bookkeeping.
	q.ring[q.head] = MessageQueueEntry{
		Message:           message,
		TransmissionCount: 0,
	}
	q.indexByAddress[message.GetAddress()] = q.head
	q.head = (q.head + 1) % len(q.ring)
	AddMessageTotal.Inc()
}

func (q *RingBufferQueue) PrioritizeForAddress(address encoding.Address) {
	ringIndex, found := q.indexByAddress[address]
	if found {
		messageType := q.ring[ringIndex].Message.GetType()
		if messageType == encoding.MessageTypeSuspect ||
			messageType == encoding.MessageTypeFaulty {
			q.priorityIndex = ringIndex
			return
		}
	}
	q.priorityIndex = -1
}

// Get returns the element at the logical index.
// idx  [5 6 7 2 3 4 0 1  ]
// Ring [A B C I J K L M _]
//
//	     ^tail       head^
//	start of bucket 0^
//	           ^start of bucket 1
//	     ^start of bucket 2
func (q *RingBufferQueue) Get(logicalIndex int) Message {
	// TODO: This function should be dropped, as we do not want to use the queue this way. If we need it for tests,
	// introduce a test function which does what we do here.
	for i, msg := range q.All() {
		if i == logicalIndex {
			return msg
		}
	}
	panic("ring buffer queue index out of bounds")
}

func (q *RingBufferQueue) All() iter.Seq2[int, Message] {
	return q.all
}

func (q *RingBufferQueue) all(yield func(int, Message) bool) {
	logicalIndex := 0

	// We need to handle the priority element if set.
	if q.priorityIndex >= 0 {
		if !yield(logicalIndex, q.ring[q.priorityIndex].Message) {
			return
		}
		logicalIndex++
	}

	for i := range q.bucketStarts {
		// We get the bounds of the current bucket.
		bucketStart := q.bucketStarts[i]
		var bucketEnd int
		if i == 0 {
			// The first bucket always ends with the head.
			bucketEnd = q.head
		} else {
			// Other buckets end with the start of the previous bucket.
			bucketEnd = q.bucketStarts[i-1]
		}

		for ringIndex := bucketStart; ringIndex != bucketEnd; ringIndex = (ringIndex + 1) % len(q.ring) {
			if ringIndex == q.priorityIndex {
				// Skip the element if this is the priority index. We already returned that one as the first element.
				continue
			}
			if !yield(logicalIndex, q.ring[ringIndex].Message) {
				return
			}
			logicalIndex++
		}
	}
}

func (q *RingBufferQueue) MarkFirstNMessagesTransmitted(count int) {
	count = min(count, q.Len())

	bucketEnd := q.head
	for i := range q.bucketStarts {
		if count == 0 {
			// We already moved everything. We can stop here.
			break
		}

		// See where the bucket starts and how many elements it contains.
		bucketStart := q.bucketStarts[i]
		bucketSize := (bucketEnd - bucketStart + len(q.ring)) % len(q.ring)
		if bucketSize == 0 {
			// The bucket is empty. Look at the next bucket.
			// Because the bucket is empty, we do not need to update bucketEnd as this will never change.
			continue
		}

		// Update the transmission count for the elements we intend to move to the next bucket.
		moveCount := min(count, bucketSize)
		for j := bucketStart; j < bucketStart+moveCount; j++ {
			q.ring[j%len(q.ring)].TransmissionCount++
		}

		// Move the start of the bucket to the right. This effectively makes the moved elements part of the next bucket.
		q.bucketStarts[i] = (bucketStart + moveCount) % len(q.ring)
		count -= moveCount
		bucketEnd = bucketStart
	}
	q.cleanupTail()
}

func (q *RingBufferQueue) cleanupTail() {
	// All elements between q.tail and the last bucket start have exceeded their maximum transmission count and are
	// now dropped.
	for q.tail != q.bucketStarts[len(q.bucketStarts)-1] {
		delete(q.indexByAddress, q.ring[q.tail].Message.GetAddress())
		if q.priorityIndex == q.tail {
			q.priorityIndex = -1
		}
		q.tail = (q.tail + 1) % len(q.ring)
		RemoveMessageTotal.Inc()
	}
}

// grow increases the capacity of the ring buffer. Call this method when you need more space. Be aware that this is an
// expensive operation and should not happen often. Once grown, the ring buffer will never shrink back to a smaller
// size.
func (q *RingBufferQueue) grow() {
	// Copy all elements into a new ring buffer. Moving everything to the start of the buffer.
	newRing := make([]MessageQueueEntry, len(q.ring)*2)
	var n int
	if q.tail <= q.head {
		copy(newRing, q.ring[q.tail:q.head])
	} else {
		n = copy(newRing, q.ring[q.tail:])
		copy(newRing[n:], q.ring[:q.head])
	}

	// As the indices changed, we have to update our bookkeeping with new indices. Indices are updated by applying
	// the offset they were moved.
	for i := range q.bucketStarts {
		q.bucketStarts[i] = q.adjustIndexAfterGrow(q.bucketStarts[i], q.tail, q.head, n)
	}
	for address, ringIndex := range q.indexByAddress {
		q.indexByAddress[address] = q.adjustIndexAfterGrow(ringIndex, q.tail, q.head, n)
	}
	if q.priorityIndex >= 0 {
		q.priorityIndex = q.adjustIndexAfterGrow(q.priorityIndex, q.tail, q.head, n)
	}

	// At ethe end update our ring buffer bookkeeping. Make sure to overwrite q.ring last, otherwise the length will be
	// off.
	q.head = q.Len()
	q.tail = 0
	q.ring = newRing
}

// adjustIndexAfterGrow returns the new index after the ring buffer was grown to a bigger size.
func (q *RingBufferQueue) adjustIndexAfterGrow(oldRingIndex int, oldTail int, oldHead int, n int) int {
	if oldTail <= oldHead {
		// There was no wraparound over ring buffer boundaries. The content was simply copied to the first position
		// in the new ring buffer. All indices are adjusted by the oldTail offset to match their new location.
		return oldRingIndex - oldTail
	}

	if oldTail <= oldRingIndex {
		// The content did wrap around over ring buffer boundaries. The content until the end of the ring buffer was
		// simply copied to the first position in the new ring buffer. All indices are adjusted by the oldTail offset
		// to match their new location.
		return oldRingIndex - oldTail
	}

	// The content did wrap around over ring buffer boundaries. The content from the start of the ring buffer was copied
	// after the rest of the content in the new ring buffer. All indices are adjusted by the n offset to match their new
	// location.
	return oldRingIndex + n
}

func (q *RingBufferQueue) moveToFirstBucket(ringIndex int) int {
	found := false
	// Note that we do not loop into bucket 0, because when we swapped the element from bucket 1 into bucket 0, we don't
	// have to do anything in bucket 0 anymore.
	for i := len(q.bucketStarts) - 1; i > 0; i-- {
		bucketStart := q.bucketStarts[i]
		bucketEnd := q.bucketStarts[i-1]

		// Skip buckets until we found the first bucket which contains the element.
		if !found {
			if !q.inBucket(ringIndex, bucketStart, bucketEnd) {
				continue
			}
			found = true
		}

		// Move the element forward by one bucket. We do this by swapping it with the last element in the current bucket
		// and adjusting the bucket start of the next bucket.
		newRingIndex := bucketEnd - 1
		if newRingIndex < 0 {
			newRingIndex += len(q.ring)
		}
		q.swapElements(ringIndex, newRingIndex)
		q.bucketStarts[i-1] = newRingIndex
		ringIndex = newRingIndex
	}
	q.ring[ringIndex].TransmissionCount = 0
	return ringIndex
}

func (q *RingBufferQueue) inBucket(ringIndex int, bucketStart int, bucketEnd int) bool {
	if bucketStart == bucketEnd {
		// The bucket is empty, it cannot hold any elements.
		return false
	}

	if bucketStart < bucketEnd {
		// The bucket does not wrap around
		return bucketStart <= ringIndex && ringIndex < bucketEnd
	}

	// The bucket does wrap around
	return bucketStart <= ringIndex || ringIndex < bucketEnd
}

func (q *RingBufferQueue) swapElements(ringIndex1 int, ringIndex2 int) {
	if ringIndex1 == ringIndex2 {
		// Shortcut for when the element is already at the desired place.
		return
	}

	// Swap the elements
	q.ring[ringIndex1], q.ring[ringIndex2] = q.ring[ringIndex2], q.ring[ringIndex1]

	// Update the index by address bookkeeping
	q.indexByAddress[q.ring[ringIndex1].Message.GetAddress()] = ringIndex1
	q.indexByAddress[q.ring[ringIndex2].Message.GetAddress()] = ringIndex2

	// Fix priority queue index
	if q.priorityIndex == ringIndex1 {
		q.priorityIndex = ringIndex2
	} else if q.priorityIndex == ringIndex2 {
		q.priorityIndex = ringIndex1
	}
}

// ValidateInternalState reports if the internal state is valid.
// This function is expensive and should not be called outside of tests.
func (q *RingBufferQueue) ValidateInternalState() error {
	// TODO: How to validate internal state?
	return nil
}
