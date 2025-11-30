package randmember

import (
	"math/rand"

	"github.com/backbone81/membership/internal/encoding"
)

// Picker provides functionality to pick a given number of unique random members from a member slice.
//
// The implementation is using a partial Fisher-Yates shuffle
// (https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle). This avoids copying and shuffling the entire members
// slice and is basically the same what rand.Shuffle is using for shuffling full slices.
// As we try to avoid copying the full slice, we use a map to remember those swaps we already did. This allows us to
// have minimal memory consumption even for large member slices.
//
// Picker is not safe for concurrent use by multiple goroutines. Callers must serialize access to all methods.
type Picker struct {
	pickRandomMembersSwap map[int]int
}

// NewPicker creates a new picker.
func NewPicker() *Picker {
	return &Picker{
		pickRandomMembersSwap: make(map[int]int, 16),
	}
}

// Pick triggers the given callback count times with unique random members. If the members slice does not contain enough
// members to fulfill count unique members, the length of the members slice is used as an upper bound instead. Unique
// members means that it is guaranteed that no member is returned twice over all callback calls of a single Pick call.
func (p *Picker) Pick(count int, members []encoding.Member, fn func(member encoding.Member)) {
	count = min(count, len(members))
	if count == 0 {
		return
	}

	clear(p.pickRandomMembersSwap)

	// We iterate over the number of elements we want to retrieve.
	for i := 0; i < count; i++ {
		// For every element, we pick a random other element which is identical to the current element or bigger.
		j := i + rand.Intn(len(members)-i)

		// We look up the real indexes according to what swaps we already did in the past.
		iReal := p.pickRandomMemberIndex(i)
		jReal := p.pickRandomMemberIndex(j)

		// Let's remember that the j element is now replaced by the real i element. Note that we do not remember the
		// i element, because we will never look at it again, we are only swapping with elements to the right of i, not
		// left of i.
		p.pickRandomMembersSwap[j] = iReal

		// Append the swapped member to the result.
		fn(members[jReal])
	}
}

// pickRandomMemberIndex is a helper method which resolves a given index through the swap map to get the real index.
func (p *Picker) pickRandomMemberIndex(index int) int {
	if value, found := p.pickRandomMembersSwap[index]; found {
		return value
	}
	return index
}

// PickWithout triggers the given callback count times with unique random members which do not include exclude. If the
// members slice does not contain enough members to fulfill count unique members, the length of the members slice is
// used as an upper bound instead. Unique members means that it is guaranteed that no member is returned twice over
// all callback calls of a single PickWithout call.
func (p *Picker) PickWithout(count int, members []encoding.Member, exclude encoding.Address, fn func(member encoding.Member)) {
	// We retrieve one member more that requested to allow for an additional member if we find the excluded address.
	var counter int
	p.Pick(count+1, members, func(member encoding.Member) {
		if member.Address.Equal(exclude) || counter == count {
			return
		}
		fn(member)
		counter++
	})
}
