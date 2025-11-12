package gossip

import "github.com/backbone81/membership/internal/encoding"

type RingBufferQueue struct {
	ring []MessageQueueEntry

	// head provides the position of the next write
	head int

	// tail provides the position of the next read
	tail int

	// bucketStarts provides the index into the ring where the bucket starts.
	// Bucket 0: [bucketStarts[0], head)
	// Bucket 1: [bucketStarts[1], bucketStarts[0])
	// Bucket 2: [bucketStarts[2], bucketStarts[1])
	// ...
	// Bucket n: [tail, bucketStarts[n-1])
	bucketStarts []int

	// indexByAddress is correlating the position within queue with the address. This helps in making checks for
	// existing gossip faster.
	indexByAddress map[encoding.Address]int

	// priorityQueueIndex is the queue index which should be returned with priority and always with Get(0). This allows
	// us to prioritize suspect and faulty gossip when we are talking to that node right now.
	priorityQueueIndex int
}

// NewRingBufferQueue creates a new gossip message queue.
func NewRingBufferQueue(maxTransmissionCount int) *RingBufferQueue {
	maxTransmissionCount = max(1, maxTransmissionCount)
	return &RingBufferQueue{
		ring:               make([]MessageQueueEntry, 0, 1024),
		bucketStarts:       make([]int, maxTransmissionCount),
		indexByAddress:     make(map[encoding.Address]int, 1024),
		priorityQueueIndex: -1,
	}
}

// SetMaxTransmissionCount updates the max transmission count to the new value. The new value will be used on the next
// call to MarkFirstNMessagesTransmitted.
func (q *RingBufferQueue) SetMaxTransmissionCount(maxTransmissionCount int) {
	maxTransmissionCount = max(1, maxTransmissionCount)
	if maxTransmissionCount < len(q.bucketStarts) {
		// Note that we drop the bucket here, but the elements stay around between tail and the new last bucket until
		// MarkFirstNMessagesTransmitted is called the next time which will then clean up those elements.
		q.bucketStarts = q.bucketStarts[:maxTransmissionCount]
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
	q.priorityQueueIndex = -1
}

func (q *RingBufferQueue) Add(message Message) {
	queueIndex, found := q.indexByAddress[message.GetAddress()]
	if found {
		queueEntry := &q.ring[queueIndex]

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
			queueIndex = q.moveToFirstBucket(queueIndex)
			queueEntry = &q.ring[queueIndex]
			queueEntry.TransmissionCount = 0
		}
		// TODO: We need to update the existing message
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
	queueIndex, found := q.indexByAddress[address]
	if found {
		messageType := q.ring[queueIndex].Message.GetType()
		if messageType == encoding.MessageTypeSuspect ||
			messageType == encoding.MessageTypeFaulty {
			q.priorityQueueIndex = queueIndex
			return
		}
	}
	q.priorityQueueIndex = -1
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
	// We need to handle a priority element if set.
	if q.priorityQueueIndex >= 0 {
		if logicalIndex == 0 {
			return q.ring[q.priorityQueueIndex].Message
		}
		// As we are not accessing with index 0 anymore, we need to decrement once to not skip the real first element
		// in the queue.
		logicalIndex--
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

		// We need the size of the current bucket.
		bucketLength := (bucketEnd - bucketStart + len(q.ring)) % len(q.ring)
		if bucketLength == 0 {
			// This bucket is empty, look into the next one.
			continue
		}

		// We need to make sure that we skip the priority element in case we move over it. Otherwise, we would return
		// it twice during iteration.
		if q.priorityQueueIndex >= 0 {
			for j := 0; j < min(logicalIndex, bucketLength); j++ {
				ringIndex := (bucketStart + j) % len(q.ring)
				if ringIndex == q.priorityQueueIndex {
					// Adjust the offset back again, as we now crossed the priority element. No need to continue
					// looping as the priority element only occurs once.
					logicalIndex++
					break
				}
			}
		}

		if logicalIndex < bucketLength {
			// The logical index is located in the current bucket. Return the message.
			ringIndex := (bucketStart + logicalIndex) % len(q.ring)
			return q.ring[ringIndex].Message
		}

		// The logical index was not located in the current bucket, so we subtract the bucket length from the logical
		// index and have a look at the next bucket.
		logicalIndex -= bucketLength
	}
	panic("ring buffer queue index out of bounds")
}

func (q *RingBufferQueue) MarkFirstNMessagesTransmitted(count int) {
	count = min(count, q.Len())

	previousBucketStart := q.head
	for i := range q.bucketStarts {
		if count == 0 {
			// We already moved everything. We can stop here.
			break
		}

		// See where the bucket starts and how many elements it contains.
		bucketStart := q.bucketStarts[i]
		bucketSize := (previousBucketStart - bucketStart + len(q.ring)) % len(q.ring)
		if bucketSize == 0 {
			// The bucket is empty. Look at the next bucket.
			previousBucketStart = bucketStart
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
		previousBucketStart = bucketStart
	}

	// All elements between q.tail and the last bucket start have exceeded their maximum transmission count and are
	// now dropped.
	for q.tail != q.bucketStarts[len(q.bucketStarts)-1] {
		delete(q.indexByAddress, q.ring[q.tail].Message.GetAddress())
		if q.priorityQueueIndex == q.tail {
			q.priorityQueueIndex = -1
		}
		q.tail = (q.tail + 1) % len(q.ring)
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
	for address, index := range q.indexByAddress {
		q.indexByAddress[address] = q.adjustIndexAfterGrow(index, q.tail, q.head, n)
	}
	if q.priorityQueueIndex >= 0 {
		q.priorityQueueIndex = q.adjustIndexAfterGrow(q.priorityQueueIndex, q.tail, q.head, n)
	}

	// At ethe end update our ring buffer bookkeeping. Make sure to overwrite q.ring last, otherwise the length will be
	// off.
	q.head = q.Len()
	q.tail = 0
	q.ring = newRing
}

// adjustIndexAfterGrow returns the new index after the ring buffer was grown to a bigger size.
func (q *RingBufferQueue) adjustIndexAfterGrow(oldIndex int, oldTail int, oldHead int, n int) int {
	if oldTail <= oldHead {
		// There was no wraparound over ring buffer boundaries. The content was simply copied to the first position
		// in the new ring buffer. All indices are adjusted by the oldTail offset to match their new location.
		return oldIndex - oldTail
	}

	if oldTail <= oldIndex {
		// The content did wrap around over ring buffer boundaries. The content until the end of the ring buffer was
		// simply copied to the first position in the new ring buffer. All indices are adjusted by the oldTail offset
		// to match their new location.
		return oldIndex - oldTail
	}

	// The content did wrap around over ring buffer boundaries. The content from the start of the ring buffer was copied
	// after the rest of the content in the new ring buffer. All indices are adjusted by the n offset to match their new
	// location.
	return oldIndex + n
}
