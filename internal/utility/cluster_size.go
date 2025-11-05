package utility

import (
	"iter"
	"math"
)

// ClusterSize returns a range over function type which returns a linear sequence of integers from minMemberCount to
// linearCutoff and afterward doubles until maxMemberCount is reached. This allows for easy iteration over several cluster sizes
func ClusterSize(minMemberCount int, linearCutoff int, maxMemberCount int) iter.Seq[int] {
	if minMemberCount < 1 {
		panic("the min member count needs to be a positive integer")
	}
	if linearCutoff < minMemberCount {
		panic("the linear cutoff needs to be equal or bigger than the min member count")
	}
	if maxMemberCount < linearCutoff {
		panic("the max member count needs to be equal or bigger than the linear cutoff")
	}
	if maxMemberCount > math.MaxInt>>1 {
		panic("the max member count exceeds the maximum limit")
	}

	return func(yield func(int) bool) {
		var lastYieldedMemberCount int
		for memberCount := minMemberCount; memberCount <= maxMemberCount; {
			if !yield(memberCount) {
				return
			}
			lastYieldedMemberCount = memberCount

			if memberCount < linearCutoff {
				memberCount++
			} else {
				memberCount *= 2
			}
		}

		// There is the chance to overshoot and never yield for max member count. We tackle that special case here.
		if lastYieldedMemberCount < maxMemberCount {
			yield(maxMemberCount)
		}
	}
}
