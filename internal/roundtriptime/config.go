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
	Count:      100,
	Percentile: 0.99,
	Alpha:      0.3,
	// The minimum needs to account for scheduling inconsistencies. 5 ms is still safe without triggering missed
	// timeouts in the scheduler.
	Minimum: 5 * time.Millisecond,
	Default: 100 * time.Millisecond,
	// As the default protocol period is 1 second, the RTT needs to be below 1/3, because we need 3 round trips in
	// a full ping, indirect ping protocol period. Therefore, the default is set to 300 ms which still provides some
	// leeway.
	Maximum: 300 * time.Millisecond,
}
