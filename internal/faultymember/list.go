package faultymember

import (
	"fmt"

	"github.com/backbone81/membership/internal/encoding"
)

// List is responsible for managing the faulty members. It will manage the members in a way which helps with removing
// them quickly after their retention time is up.
//
// This implementation works like this:
//
// Index:        [0 1 2 3 4 5 6]
// Members:      [A B C D E F G]
// BucketStarts: [3   2   1 0  ]
//
// New faulty members are always appended at the end of the list and are (nearly) never moved from their initial
// location. Member G was added last and placed after member F.
//
// Within the list, we maintain buckets of members. Each bucket holds all members which reside in the list for the same
// number of list requests. The bucket itself is identified by the number of list requests its members are there.
// Bucket 0 is the bucket which contains all members which are in the list for 0 list requests, while bucket 3
// contains all members which are in the list for 3 list requests. As new members are always appended at the end of
// the list and therefore must be part of bucket 0, the bucket number grows towards the start of the list. Buckets are
// located through their BucketStart index which points to the location where the first element of that bucket can be
// found. The end of the bucket is the start of the bucket which comes after it. The end of bucket 3 is the start of
// bucket 2.
//
// When we iterate over the members, we do not require a specific order. We therefore iterate from the member which is
// in the list longest to the member which is in the list shortest.
//
// At the end of each list request, all members are moved to the next bucket, which means moving the start of all
// buckets to the start of the bucket before. The member itself stays at its place.
//
// Members are deduplicated based on their address. New members with an address already present in the list are
// silently dropped.
//
// The list is implemented as a ring buffer to reduce the number of memory allocations and to allow elements to stay
// at their location as long as possible. This means that we need to consider wrap-arounds at the end of the ring buffer
// throughout our code. When the ring buffer is full, it can grow by copying the content over into a new buffer.
//
// This implementation results in the most common operations being O(1) and some less common situations to be
// O(buckets) in a worst case scenario.
//
// List is not safe for concurrent use by multiple goroutines. Callers must serialize access to all methods.
type List struct {
	// config holds the current configuration of the list.
	config Config

	// ring holds the storage for the ring buffer. We can assume that "len(ring) > 0" always holds.
	ring []encoding.Member

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
}

// NewList creates a new faulty member list.
func NewList(options ...Option) *List {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}
	return &List{
		config:         config,
		ring:           make([]encoding.Member, config.PreAllocationCount),
		bucketStarts:   make([]int, config.MaxListRequestCount),
		indexByAddress: make(map[encoding.Address]int, config.PreAllocationCount),
	}
}

// Config returns the current configuration of the list.
func (l *List) Config() Config {
	return l.config
}

// IsEmpty reports if the list is empty.
func (l *List) IsEmpty() bool {
	// Note that this works, because Add grows the ring buffer before the ring buffer is full. Otherwise, this check
	// could fail with a full ring buffer.
	return l.head == l.tail
}

// Len returns the number of elements currently stored inside the list.
func (l *List) Len() int {
	return (l.head - l.tail + len(l.ring)) % len(l.ring)
}

// Cap returns the capacity the list can take without growing the ring buffer.
func (l *List) Cap() int {
	return len(l.ring)
}

// Buckets returns the number of buckets available for elements.
func (l *List) Buckets() int {
	return len(l.bucketStarts)
}

// Clear removes all elements from the list. It retains the memory which was already allocated to be used with upcoming
// members.
func (l *List) Clear() {
	clear(l.ring)
	clear(l.bucketStarts)
	clear(l.indexByAddress)
	l.head = 0
	l.tail = 0
}

// Add puts the given member at the end of the list into bucket 0. If the list already contains a member with the same
// address, it is overwritten, but the bucket is never changed.
func (l *List) Add(member encoding.Member) {
	index, found := l.indexByAddress[member.Address]
	if found {
		// We overwrite the member, but we keep it at the same location in the same bucket, because the member is
		// still faulty.
		entry := &l.ring[index]
		*entry = member
		return
	}

	// Make sure we have space left in our ring buffer to add the new element.
	if (l.head+1)%len(l.ring) == l.tail {
		l.grow()
	}

	// Add element and update bookkeeping.
	l.ring[l.head] = member
	l.indexByAddress[member.Address] = l.head
	l.head = (l.head + 1) % len(l.ring)
}

// Get returns the member with the given address. Returns true when the member was found, returns false when the member
// was not found.
func (l *List) Get(address encoding.Address) (encoding.Member, bool) {
	index, found := l.indexByAddress[address]
	if !found {
		return encoding.Member{}, false
	}
	return l.ring[index], true
}

// Remove deletes the member with the given address from the list. Does nothing if the address is not found.
func (l *List) Remove(address encoding.Address) {
	index, ok := l.indexByAddress[address]
	if !ok {
		return
	}

	// We move the member through all buckets by swapping it with the first element in the bucket and moving the start
	// of the bucket by one. When the member drops out of the last bucket we can remove it by cleaning up the tail.
	found := false
	bucketEnd := l.head
	for i := 0; i < len(l.bucketStarts); i++ {
		bucketStart := l.bucketStarts[i]

		// Skip buckets until we found the first bucket which contains the element.
		if !found {
			if !l.inBucket(index, bucketStart, bucketEnd) {
				bucketEnd = bucketStart
				continue
			}
			found = true
		}

		// Move the element backward by one bucket. We do this by swapping it with the first element in the current
		// bucket and adjusting the bucket start of the current bucket.
		l.swapElements(index, bucketStart)
		index = bucketStart
		bucketStart = (bucketStart + 1) % len(l.ring)
		l.bucketStarts[i] = bucketStart
		bucketEnd = bucketStart
	}
	l.cleanupTail()
}

// ForEach executes the given function for all elements stored in the list. The members which are in the list for the
// longest time are returned first, the members which were added recently are returned last. Return false to abort the
// iteration.
//
// This function only returns the members of the first 50% of buckets. This is necessary to avoid situations where one
// memberlist drops a faulty member because it outlived the maximum count, but is then added back again after a full
// list sync with some other member which still has the faulty member in his list. By applying a shorter timeout to
// what is returned by a full list sync, we make sure that we do not re-add failed members.
//
// Note that we are explicitly not providing a range over function type for iterating over all members, because
// that would cause memory allocations for the range over for loop, as it needs to introduce state which is allocated
// on the heap. The solution with ForEach is less nice, but it allows for zero allocations.
func (l *List) ForEach(fn func(encoding.Member) bool) {
	// The start index is the index of the bucket in the middle.
	halfwayBucket := len(l.bucketStarts) / 2
	if halfwayBucket == 0 {
		// Nothing to return
		return
	}

	startIndex := l.bucketStarts[halfwayBucket-1]
	for index := startIndex; index != l.head; index = (index + 1) % len(l.ring) {
		if !fn(l.ring[index]) {
			return
		}
	}
}

// ListRequestObserved moves all members into the next bucket. If they leave the last bucket, they are deleted from the
// list.
func (l *List) ListRequestObserved() {
	bucketEnd := l.head
	for i := range l.bucketStarts {
		bucketStart := l.bucketStarts[i]
		l.bucketStarts[i] = bucketEnd
		bucketEnd = bucketStart
	}
	l.cleanupTail()
}

// cleanupTail deletes all elements between tail and the last bucket. They have exceeded their maximum list requests
// count and can be dropped.
func (l *List) cleanupTail() {
	lastBucketStart := l.bucketStarts[len(l.bucketStarts)-1]
	for l.tail != lastBucketStart {
		delete(l.indexByAddress, l.ring[l.tail].Address)
		l.tail = (l.tail + 1) % len(l.ring)
	}
}

// grow increases the capacity of the ring buffer. Call this method when you need more space. Be aware that this is an
// expensive operation and should not happen often. Once grown, the ring buffer will never shrink back to a smaller
// size.
func (l *List) grow() {
	// Copy all elements into a new ring buffer. Moving everything to the start of the buffer.
	newRing := make([]encoding.Member, len(l.ring)*2)
	var n int
	if l.tail <= l.head {
		copy(newRing, l.ring[l.tail:l.head])
	} else {
		n = copy(newRing, l.ring[l.tail:])
		copy(newRing[n:], l.ring[:l.head])
	}

	// As the indices changed, we have to update our bookkeeping with new indices. Indices are updated by applying
	// the offset they were moved.
	for i := range l.bucketStarts {
		l.bucketStarts[i] = l.adjustIndexAfterGrow(l.bucketStarts[i], l.tail, l.head, n)
	}
	for address, ringIndex := range l.indexByAddress {
		l.indexByAddress[address] = l.adjustIndexAfterGrow(ringIndex, l.tail, l.head, n)
	}

	// At ethe end update our ring buffer bookkeeping. Make sure to overwrite l.ring last, otherwise the length will be
	// off.
	l.head = l.Len()
	l.tail = 0
	l.ring = newRing
}

// adjustIndexAfterGrow returns the new index after the ring buffer was grown to a bigger size.
func (l *List) adjustIndexAfterGrow(oldIndex int, oldTail int, oldHead int, n int) int {
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

// inBucket reports if the given index is inside the bucket.
func (l *List) inBucket(index int, bucketStart int, bucketEnd int) bool {
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
func (l *List) swapElements(index1 int, index2 int) {
	if index1 == index2 {
		// Shortcut for when the element is already at the desired place.
		return
	}

	// Swap the elements
	l.ring[index1], l.ring[index2] = l.ring[index2], l.ring[index1]

	// Update the index by address bookkeeping
	l.indexByAddress[l.ring[index1].Address] = index1
	l.indexByAddress[l.ring[index2].Address] = index2
}

// ValidateInternalState reports if the internal state is valid.
// This function is expensive and should not be called outside of tests.
func (l *List) ValidateInternalState() error {
	bucketEnd := l.head
	for i := range l.bucketStarts {
		bucketStart := l.bucketStarts[i]
		bucketSize := (bucketEnd - bucketStart + len(l.ring)) % len(l.ring)
		for j := 0; j < bucketSize; j++ {
			index := (bucketStart + j) % len(l.ring)

			// Make sure that all entries are correctly stored in the index by address
			index2, found := l.indexByAddress[l.ring[index].Address]
			if !found {
				return fmt.Errorf("message %d could not be found in index map", index)
			}
			if index2 != index {
				return fmt.Errorf("message %d has wrong index in index map", index)
			}
		}
		bucketEnd = bucketStart
	}

	// Make sure that every entry in the index map can be found in the list
	for address, index := range l.indexByAddress {
		if index >= len(l.ring) {
			return fmt.Errorf("index map for address %s points out of bounds", address)
		}
		if !l.ring[index].Address.Equal(address) {
			return fmt.Errorf("index map index mismatch for list element %d", index)
		}
	}
	return nil
}
