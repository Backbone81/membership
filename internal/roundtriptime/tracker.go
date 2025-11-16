package roundtriptime

import (
	"math"
	"slices"
	"sync"
	"time"
)

// Tracker provides a mechanic to track network round trip times and calculate a calculatedRTT on a specific percentile of
// observed observedRTTs. This can be used to dynamically adjust timeouts to a reasonable value.
//
// Tracker is safe for concurrent use by multiple goroutines. Access is synchronized internally.
type Tracker struct {
	mutex              sync.Mutex
	config             Config
	nextIndex          int
	observedRTTs       []time.Duration
	observedRTTsSorted []time.Duration
	calculatedRTT      time.Duration
}

// NewTracker creates a new Tracker.
func NewTracker(options ...Option) *Tracker {
	config := DefaultConfig
	for _, option := range options {
		option(&config)
	}

	if config.Maximum < config.Minimum {
		// The maximum is smaller than the minimum. Adjust the minimum to match the maximum.
		config.Minimum = config.Maximum
	}
	if config.Default < config.Minimum {
		// The default is smaller than the minimum. Adjust the default to match the minimum.
		config.Default = config.Minimum
	}
	if config.Maximum < config.Default {
		// The default is bigger than the maximum. Adjust the default to match the maximum.
		config.Default = config.Maximum
	}

	result := Tracker{
		config:             config,
		observedRTTs:       make([]time.Duration, config.Count),
		observedRTTsSorted: make([]time.Duration, config.Count),
		calculatedRTT:      config.Default,
	}
	// We initialize all observed RTTs with the default value. That way, we do not have to deal with the edge case of
	// the slice not being completely filled with observed values.
	for i := range result.observedRTTs {
		result.observedRTTs[i] = config.Default
	}
	return &result
}

// Config returns the config the tracker was created with.
func (t *Tracker) Config() Config {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.config
}

// Reset overwrites all observed RTTs with the default and also resets the calculated RTT to the default. This restores
// the same state as if the tracker was newly created.
func (t *Tracker) Reset() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for i := range t.observedRTTs {
		t.observedRTTs[i] = t.config.Default
	}
	t.calculatedRTT = t.config.Default
	t.nextIndex = 0
}

// AddObserved will add the given RTT to the observed RTTs. It will overwrite the oldest observed RTT in the local
// buffer.
func (t *Tracker) AddObserved(roundTripTime time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.observedRTTs[t.nextIndex] = roundTripTime
	t.nextIndex = (t.nextIndex + 1) % len(t.observedRTTs)
}

// UpdateCalculated updates the calculated RTT according to the given percentile and the smoothing factor.
func (t *Tracker) UpdateCalculated() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// We need to copy everything over into a temporary slice and sort it increasingly. Note that we might look into
	// using quickselect to avoid sorting the full slice - but it might not have any measurable impact as we are dealing
	// with slices of 60 to 120 entries in normal situations.
	copy(t.observedRTTsSorted, t.observedRTTs)
	slices.Sort(t.observedRTTsSorted)

	// The desired value is fetched according to the percentile given.
	percentileIndex := int(math.Floor(float64(len(t.observedRTTsSorted)-1) * t.config.Percentile))
	newRtt := t.observedRTTsSorted[percentileIndex]

	// The new value is mixed with the previously calculated RTT to smoothen out rapid changes.
	smoothedRtt := time.Duration(t.config.Alpha*float64(newRtt) + (1-t.config.Alpha)*float64(t.calculatedRTT))
	t.calculatedRTT = max(t.config.Minimum, min(smoothedRtt, t.config.Maximum))
}

// GetCalculated returns the calculated RTT.
func (t *Tracker) GetCalculated() time.Duration {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.calculatedRTT
}
