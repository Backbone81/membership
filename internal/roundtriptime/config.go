package roundtriptime

import "time"

// Config provides the configuration for Tracker.
type Config struct {
	// Count is the number of observed RTTs to keep in memory and calculate the percentile from.
	Count int

	// Percentile is the percentage where we find the calculated RTT in a sorted list of observed RTTs. In the range of
	// 0.0 to 1.0.
	Percentile float64

	// Alpha is the smoothing factor to apply to a newly calculated RTT. In the range of 0.0 to 1.0.
	Alpha float64

	// Default is the RTT to use after startup.
	Default time.Duration

	// Minimum is the minimum RTT the tracker will return.
	Minimum time.Duration

	// Maximum is the maximum RTT the tracker will return.
	Maximum time.Duration
}

// DefaultConfig is the default configuration for Tracker which should work fine in most situations.
var DefaultConfig = Config{
	Count:      60,
	Percentile: 0.95,
	Alpha:      0.3,
	Default:    100 * time.Millisecond,
	Minimum:    5 * time.Millisecond,
	Maximum:    300 * time.Millisecond,
}
