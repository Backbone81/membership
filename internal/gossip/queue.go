package gossip

import (
	"fmt"
	"iter"

	"github.com/backbone81/membership/internal/encoding"
)

// Queue is responsible for managing the messages we need to gossip. It will manage the messages in a way which helps
// distribute new messages quickly.
//
// This implementation works like this:
//
// Index:        [0 1 2 3 4 5 6]
// Messages:     [A B C D E F G]
// BucketStarts: [3   2   1 0  ]
// Iteration:    [5 6 3 4 2 0 1]
//
// New messages are always appended at the end of the queue and are (nearly) never moved from their initial location.
// Message G was added last and placed after message F.
//
// Within the queue, we maintain buckets of messages. Each bucket holds all messages which were gossiped the same number
// of times. The bucket itself is identified by the number of times its messages were gossiped. Bucket 0 is the bucket
// which contains all messages which were gossiped 0 times, while bucket 3 contains all messages which were gossiped 3
// times. As new messages are always appended at the end of the queue and therefore must be part of bucket 0, the bucket
// number grows towards the start of the queue. Buckets are located through their BucketStart index which points to the
// location where the first element of that bucket can be found. The end of the bucket is the start of the bucket which
// comes after it. The end of bucket 3 is the start of bucket 2.
//
// When we iterate over the messages, we want to start with what was gossiped the least amount of time and sits in the
// queue the longest. This means that iteration through the queue is not linear but always starts at the start of the
// bucket and moves to the end of the bucket before jumping over to the start of the next bucket.
//
// Marking a message as having been gossiped once means moving the start of the bucket by one position. The message
// itself stays at its place. Marking the first message F as having been transmitted will move the bucket start for
// bucket 0 from index 5 to index 6, which results in message F now being part of bucket 1.
//
// The queue is implemented as a ring buffer to reduce the number of memory allocations and to allow elements to stay
// at their location as long as possible. This means that we need to consider wrap-arounds at the end of the ring buffer
// throughout our code. When the ring buffer is full, it can grow by copying the content over into a new buffer.
//
// This implementation results in the most common operations being O(1) and some less common situations to be
// O(buckets) in a worst case scenario.
//
// Queue is not safe for concurrent use by multiple goroutines. Callers must serialize access to all methods.
type Queue struct {
	// config holds the current configuration of the queue.
	config Config

	// ring holds the storage for the ring buffer. We can assume that "len(ring) > 0" always holds.
	ring []MessageQueueEntry

	// head provides the index into the ring for the next write.
	head int

	// tail provides the index into the ring for the next read.
	tail int

	// bucketStarts provides the index into the ring where the bucket starts.
	// Bucket 0: [bucketStarts[0], head)
	// Bucket 1: [bucketStarts[1], bucketStarts[0])
	// ...
	// Bucket n: [tail, bucketStarts[n-1])
	bucketStarts []int

	// indexByAddress provides the index into the ring for a given address. This helps in making checks for
	// existing messages faster.
	indexByAddress map[encoding.Address]int

	// priorityIndex is the ring index which should be returned as the first element when iterating of all messages.
	// This allows us to prioritize suspect and faulty messages when we are talking to that node right now.
	priorityIndex int
}

// NewQueue creates a new gossip message queue.
func NewQueue(options ...Option) *Queue {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}
	return &Queue{
		config:         config,
		ring:           make([]MessageQueueEntry, config.PreAllocationCount),
		bucketStarts:   make([]int, config.MaxTransmissionCount),
		indexByAddress: make(map[encoding.Address]int, config.PreAllocationCount),
		priorityIndex:  -1,
	}
}

// Config returns the current configuration of the queue.
func (q *Queue) Config() Config {
	return q.config
}

// IsEmpty reports if the queue is empty.
func (q *Queue) IsEmpty() bool {
	// Note that this works, because Add grows the ring buffer before the ring buffer is full. Otherwise, this check
	// could fail with a full ring buffer.
	return q.head == q.tail
}

// Len returns the number of elements currently stored inside the queue.
func (q *Queue) Len() int {
	return (q.head - q.tail + len(q.ring)) % len(q.ring)
}

// Cap returns the capacity the queue can take without growing the ring buffer.
func (q *Queue) Cap() int {
	return len(q.ring)
}

// Clear removes all elements from the queue. It retains the memory which was already allocated to be used with upcoming
// messages.
func (q *Queue) Clear() {
	clear(q.ring)
	clear(q.bucketStarts)
	clear(q.indexByAddress)
	q.head = 0
	q.tail = 0
	q.priorityIndex = -1
}

// SetMaxTransmissionCount updates the max transmission count to the new value. Messages are held in the queue until
// the reach that count.
func (q *Queue) SetMaxTransmissionCount(count int) {
	// We are re-using the option here, to not duplicate the validation and possible clamping of values.
	WithMaxTransmissionCount(count)(&q.config)

	if q.config.MaxTransmissionCount < len(q.bucketStarts) {
		// The new max transmission count is smaller than what we have right now. We therefore shrink the number of
		// buckets down and remove those elements from the queue.
		q.bucketStarts = q.bucketStarts[:q.config.MaxTransmissionCount]
		q.cleanupTail()
	} else {
		// The new max transmission count is bigger than what we have right now. We need to grow the number of buckets
		// to match the new count. New buckets start always at the same index as the last bucket to create a zero length
		// bucket.
		for len(q.bucketStarts) < q.config.MaxTransmissionCount {
			q.bucketStarts = append(q.bucketStarts, q.bucketStarts[len(q.bucketStarts)-1])
		}
	}
}

// Add puts the given message at the end of the queue into bucket 0. If the queue already contains a message about the
// given address, the existing message is overwritten with the correct message precedence and moved to the first bucket
// again. Note that overwriting an existing message is O(buckets/2) on average, whereas adding a new message is O(1).
func (q *Queue) Add(message Message) {
	index, found := q.indexByAddress[message.GetAddress()]
	if found {
		// The queue already contains a message for that address. Let's see if we need to overwrite it.
		entry := &q.ring[index]
		if !ShouldReplaceExistingWithNew(entry.Message, message) {
			return
		}

		entry.Message = message
		if entry.TransmissionCount != 0 {
			index = q.moveToFirstBucket(index)
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

// Prioritize marks a message for the given address as priority. If such a message exists, it will always be
// returned first when iterating over all. Otherwise, this method has no effect.
func (q *Queue) Prioritize(address encoding.Address) {
	index, found := q.indexByAddress[address]
	if found {
		messageType := q.ring[index].Message.GetType()
		if messageType == encoding.MessageTypeSuspect ||
			messageType == encoding.MessageTypeFaulty {
			// We are only prioritizing suspect or faulty messages. We do not have to tell the member that we know that
			// it is alive, for example.
			q.priorityIndex = index
			return
		}
	}
	q.priorityIndex = -1
}

// Get returns the element with the given logical index. This is a quality of life function which is implemented as a
// wrapper around All. Prefer using All for better performance when ranging over multiple elements.
func (q *Queue) Get(logicalIndex int) Message {
	for index, message := range q.All() {
		if index == logicalIndex {
			return message
		}
	}
	panic("queue index out of bounds")
}

// All returns a range over function which iterates over all elements stored in the queue. The messages transmitted
// the least amount of time are returned first, the most transmitted messages last.
func (q *Queue) All() iter.Seq2[int, Message] {
	return q.all
}

// all is a helper function which is the range over function returned by All. This avoids a memory allocation
// which an anonymous function with a closure would cause.
func (q *Queue) all(yield func(int, Message) bool) {
	logicalIndex := 0

	// We need to handle the priority element if set.
	if q.priorityIndex >= 0 {
		if !yield(logicalIndex, q.ring[q.priorityIndex].Message) {
			return
		}
		logicalIndex++
	}

	bucketEnd := q.head
	for i := range q.bucketStarts {
		bucketStart := q.bucketStarts[i]
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
		bucketEnd = bucketStart
	}
}

// MarkTransmitted moves the first count messages to the next bucket. If they leave the last bucket, they
// are deleted from the queue.
func (q *Queue) MarkTransmitted(count int) {
	remaining := min(count, q.Len())

	bucketEnd := q.head
	for i := range q.bucketStarts {
		if remaining == 0 {
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
		moveCount := min(remaining, bucketSize)
		for j := bucketStart; j < bucketStart+moveCount; j++ {
			q.ring[j%len(q.ring)].TransmissionCount++
		}

		// Move the start of the bucket to the right. This effectively makes the moved elements part of the next bucket.
		q.bucketStarts[i] = (bucketStart + moveCount) % len(q.ring)
		remaining -= moveCount
		bucketEnd = bucketStart
	}
	q.cleanupTail()
}

// cleanupTail deletes all elements between tail and the last bucket. They have exceeded their maximum transmission
// count and can be dropped.
func (q *Queue) cleanupTail() {
	lastBucketStart := q.bucketStarts[len(q.bucketStarts)-1]
	for q.tail != lastBucketStart {
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
func (q *Queue) grow() {
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
func (q *Queue) adjustIndexAfterGrow(oldIndex int, oldTail int, oldHead int, n int) int {
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

// moveToFirstBucket moves the element at the given index to the first bucket. It returns the new index the element
// arrived at the end.
func (q *Queue) moveToFirstBucket(index int) int {
	found := false
	// Note that we do not loop into bucket 0, because when we swapped the element from bucket 1 into bucket 0, we don't
	// have to do anything in bucket 0 anymore.
	for i := len(q.bucketStarts) - 1; i > 0; i-- {
		bucketStart := q.bucketStarts[i]
		bucketEnd := q.bucketStarts[i-1]

		// Skip buckets until we found the first bucket which contains the element.
		if !found {
			if !q.inBucket(index, bucketStart, bucketEnd) {
				continue
			}
			found = true
		}

		// Move the element forward by one bucket. We do this by swapping it with the last element in the current bucket
		// and adjusting the bucket start of the next bucket.
		newIndex := bucketEnd - 1
		if newIndex < 0 {
			newIndex += len(q.ring)
		}
		q.swapElements(index, newIndex)
		q.bucketStarts[i-1] = newIndex
		index = newIndex
	}
	q.ring[index].TransmissionCount = 0
	return index
}

// inBucket reports if the given index is inside the bucket.
func (q *Queue) inBucket(index int, bucketStart int, bucketEnd int) bool {
	if bucketStart == bucketEnd {
		// The bucket is empty, it cannot hold any elements.
		return false
	}

	if bucketStart < bucketEnd {
		// The bucket does not wrap around
		return bucketStart <= index && index < bucketEnd
	}

	// The bucket does wrap around
	return bucketStart <= index || index < bucketEnd
}

// swapElements replaces two elements with each other. Updates other relevant bookkeeping to keep internal state sane.
func (q *Queue) swapElements(index1 int, index2 int) {
	if index1 == index2 {
		// Shortcut for when the element is already at the desired place.
		return
	}

	// Swap the elements
	q.ring[index1], q.ring[index2] = q.ring[index2], q.ring[index1]

	// Update the index by address bookkeeping
	q.indexByAddress[q.ring[index1].Message.GetAddress()] = index1
	q.indexByAddress[q.ring[index2].Message.GetAddress()] = index2

	// Fix priority queue index
	if q.priorityIndex == index1 {
		q.priorityIndex = index2
	} else if q.priorityIndex == index2 {
		q.priorityIndex = index1
	}
}

// ValidateInternalState reports if the internal state is valid.
// This function is expensive and should not be called outside of tests.
func (q *Queue) ValidateInternalState() error {
	bucketEnd := q.head
	for i := range q.bucketStarts {
		bucketStart := q.bucketStarts[i]
		bucketSize := (bucketEnd - bucketStart + len(q.ring)) % len(q.ring)
		for j := 0; j < bucketSize; j++ {
			index := (bucketStart + j) % len(q.ring)

			// Make sure that all entries are in the correct bucket.
			if q.ring[index].TransmissionCount != i {
				return fmt.Errorf("invalid transmission count at index %d", index)
			}

			// Make sure that all entries are correctly stored in the index by address
			index2, found := q.indexByAddress[q.ring[index].Message.GetAddress()]
			if !found {
				return fmt.Errorf("message %d could not be found in index map", index)
			}
			if index2 != index {
				return fmt.Errorf("message %d has wrong index in index map", index)
			}
		}
		bucketEnd = bucketStart
	}

	// Make sure that every entry in the index map can be found in the queue
	for address, index := range q.indexByAddress {
		if index >= len(q.ring) {
			return fmt.Errorf("index map for address %s points out of bounds", address)
		}
		if !q.ring[index].Message.GetAddress().Equal(address) {
			return fmt.Errorf("index map index mismatch for queue element %d", index)
		}
	}
	return nil
}
