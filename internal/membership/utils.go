package membership

import (
	"math"
	"time"
)

// PingTargetProbability returns the probability of a member being chosen as ping target in a protocol period.
// The probability is returned as a value between 0 and 1.
// This formula is taken from SWIM chapter 3.1. SWIM Failure Detector.
// memberCount is the number of members in total.
// memberReliability is the probability a member is available as a value between 0 and 1.
func PingTargetProbability(memberCount int, memberReliability float64) float64 {
	return 1.0 - math.Pow(1.0-1.0/float64(memberCount)*memberReliability, float64(memberCount)-1.0)
}

// PingTargetProbabilityLimes returns the probability of a member being chosen as ping target in a protocol period
// when the number of members approaches infinity.
// The probability is returned as a value between 0 and 1.
// This formula is taken from SWIM chapter 3.1. SWIM Failure Detector.
// memberReliability is the probability a member is available as a value between 0 and 1.
func PingTargetProbabilityLimes(memberReliability float64) float64 {
	return 1.0 - math.Exp(-memberReliability)
}

// FailureDetectionDuration is the expected time between failure of an arbitrary member and its detection by some other
// member.
// This formula is taken from SWIM chapter 3.1. SWIM Failure Detector.
// protocolPeriod is the time for a full protocol cycle.
// memberCount is the number of members in total.
// memberReliability is the probability a member is available as a value between 0 and 1.
func FailureDetectionDuration(protocolPeriod time.Duration, memberCount int, memberReliability float64) time.Duration {
	return time.Duration(float64(protocolPeriod) * 1.0 / PingTargetProbability(memberCount, memberReliability))
}

// FailureDetectionDurationLimes is the expected time between failure of an arbitrary member and its detection by some
// other member when the number of members approaches infinity.
// This formula is taken from SWIM chapter 3.1. SWIM Failure Detector.
// protocolPeriod is the time for a full protocol cycle.
// memberReliability is the probability a member is available as a value between 0 and 1.
func FailureDetectionDurationLimes(protocolPeriod time.Duration, memberReliability float64) time.Duration {
	return time.Duration(float64(protocolPeriod) * 1.0 / PingTargetProbabilityLimes(memberReliability))
}

// FailureDetectionFalsePositiveProbability returns the probability of a false positive failure detection.
// The probability is returned as a value between 0 and 1.
// This formula is taken from SWIM chapter 3.1. SWIM Failure Detector.
// networkReliability is the probability a network message is received by its recipient as a value between 0 and 1.
// memberReliability is the probability a member is available as a value between 0 and 1.
func FailureDetectionFalsePositiveProbability(networkReliability float64, memberReliability float64) float64 {
	return memberReliability *
		(1.0 - math.Pow(networkReliability, 2.0)) *
		(1.0 - memberReliability*math.Pow(networkReliability, 4.0)) *
		math.Exp(memberReliability) / (math.Exp(memberReliability) - 1.0)
}
